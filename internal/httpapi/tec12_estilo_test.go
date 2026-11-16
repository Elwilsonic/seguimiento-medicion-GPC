package httpapi_test

import (
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TEC-12 · Renovar los colores y el estilo de la interfaz (Issue #212).
//
// Los colores se leen de las variables de :root de styles.css; el contraste se
// calcula con la fórmula de WCAG 2.x (luminancia relativa).

// cssColor es un color escrito en el CSS: #rgb, #rrggbb, rgb(…), rgba(…) o hsl(…).
var cssColor = regexp.MustCompile(`#[0-9a-fA-F]{3}\b|#[0-9a-fA-F]{6}\b|rgba?\(|hsla?\(`)

// stylesCSS devuelve styles.css tal como lo sirve la aplicación.
func stylesCSS(t *testing.T) string {
	t.Helper()
	return get(t, newHandler(), "/styles.css").Body.String()
}

// rootBlock separa el bloque :root { … } del resto del CSS.
func rootBlock(t *testing.T, css string) (root, rest string) {
	t.Helper()
	inicio := strings.Index(css, ":root {")
	if inicio == -1 {
		t.Fatal("styles.css no tiene el bloque :root")
	}
	fin := strings.Index(css[inicio:], "}")
	return css[inicio : inicio+fin], css[:inicio] + css[inicio+fin:]
}

// paletteVars devuelve las variables de color de :root (nombre → #rrggbb).
func paletteVars(t *testing.T) map[string]string {
	t.Helper()
	root, _ := rootBlock(t, stylesCSS(t))
	vars := map[string]string{}
	for _, m := range regexp.MustCompile(`(--[a-z0-9-]+):\s*(#[0-9a-fA-F]{6})\s*;`).FindAllStringSubmatch(root, -1) {
		vars[m[1]] = strings.ToLower(m[2])
	}
	return vars
}

// contrast devuelve el contraste de WCAG entre dos colores #rrggbb.
func contrast(a, b string) float64 {
	luminance := func(hex string) float64 {
		var c [3]float64
		for i := range c {
			v, _ := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
			s := float64(v) / 255
			if s <= 0.04045 {
				c[i] = s / 12.92
			} else {
				c[i] = math.Pow((s+0.055)/1.055, 2.4)
			}
		}
		return 0.2126*c[0] + 0.7152*c[1] + 0.0722*c[2]
	}
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// textPairs son los pares de texto y fondo de la interfaz.
var textPairs = []struct{ texto, fondo string }{
	{"--texto", "--fondo"},
	{"--texto", "--panel"},
	{"--suave", "--panel"},
	{"--suave", "--fondo"},
	{"--sobre-primario", "--primario"},
	{"--sobre-primario", "--primario-oscuro"},
	{"--primario", "--panel"},
	{"--texto", "--encabezado-tabla"},
	{"--error-texto", "--error-fondo"},
	{"--ok-texto", "--ok-fondo"},
	{"--aviso-texto", "--aviso-fondo"},
}

// TEC-12 · CA-TEC12.1 · Verifica que los colores estén escritos solo en las variables de :root y que el resto del CSS, el HTML y el JavaScript no tengan colores sueltos.
func TestEstilo_ColoresSoloEnLasVariablesDeRoot(t *testing.T) {
	_, rest := rootBlock(t, stylesCSS(t))
	if colores := cssColor.FindAllString(rest, -1); len(colores) != 0 {
		t.Errorf("styles.css tiene %d colores fuera de :root: %q", len(colores), colores)
	}
	h := newHandler()
	for _, path := range []string{"/", "/app.js"} {
		if colores := regexp.MustCompile(`#[0-9a-fA-F]{6}\b`).FindAllString(get(t, h, path).Body.String(), -1); len(colores) != 0 {
			t.Errorf("%s tiene colores escritos: %q", path, colores)
		}
	}
}

// TEC-12 · CA-TEC12.2 · Verifica que la documentación técnica tenga la paleta con cada variable de :root, su valor y su contraste.
func TestEstilo_PaletaDocumentada(t *testing.T) {
	doc, err := os.ReadFile("../../docs/documentacion-tecnica.md")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(doc))
	if !strings.Contains(text, "paleta de colores") {
		t.Fatal("la documentación técnica no tiene la sección de la paleta de colores")
	}
	for _, rol := range []string{"--primario", "--acento", "--fondo", "--texto", "--ok-texto", "--error-texto", "--aviso-texto"} {
		valor, ok := paletteVars(t)[rol]
		if !ok {
			t.Errorf("la paleta no tiene %s", rol)
			continue
		}
		if !strings.Contains(text, "`"+rol+"`") || !strings.Contains(text, valor) {
			t.Errorf("la documentación técnica no documenta %s (%s)", rol, valor)
		}
	}
	if !strings.Contains(text, ":1") {
		t.Error("la documentación técnica no indica el contraste de la paleta")
	}
}

// TEC-12 · CA-TEC12.3 · Verifica que cada par de texto y fondo de la paleta tenga un contraste de al menos 4.5:1 (WCAG AA).
func TestEstilo_ContrasteAA(t *testing.T) {
	vars := paletteVars(t)
	for _, p := range textPairs {
		texto, ok1 := vars[p.texto]
		fondo, ok2 := vars[p.fondo]
		if !ok1 || !ok2 {
			t.Errorf("la paleta no tiene %s o %s", p.texto, p.fondo)
			continue
		}
		if c := contrast(texto, fondo); c < 4.5 {
			t.Errorf("contraste de %s (%s) sobre %s (%s) = %.2f:1, se esperaba al menos 4.5:1", p.texto, texto, p.fondo, fondo, c)
		}
	}
}

// TEC-12 · CA-TEC12.4 · Verifica que botones, campos, tablas, mensajes y tarjetas tomen sus colores de las variables, con el mismo borde redondeado.
func TestEstilo_ComponentesUniformes(t *testing.T) {
	css := stylesCSS(t)
	for _, regla := range []string{"button {", "input,", "th {", ".mensaje {", ".tarjeta {", ".tarjetas {"} {
		inicio := strings.Index(css, regla)
		if inicio == -1 {
			t.Errorf("styles.css no tiene la regla %q", regla)
			continue
		}
		fin := strings.Index(css[inicio:], "}")
		bloque := css[inicio : inicio+fin]
		if regla != ".tarjetas {" && !strings.Contains(bloque, "var(--") {
			t.Errorf("la regla %q no usa las variables de la paleta: %s", regla, bloque)
		}
	}
	if !strings.Contains(css, "--radio:") {
		t.Error("styles.css no define el borde redondeado común (--radio)")
	}
}

// TEC-12 · CA-TEC12.4 · Verifica que "Ver resumen" y "Ver gráficos" muestren las métricas en tarjetas, en el navegador.
func TestEstilo_DashboardEnTarjetas(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")
	s.goTo("Métricas y dashboard")

	s.u.click("#ver-resumen")
	s.u.waitTrue(`document.querySelectorAll("#datos-resumen .tarjeta").length === 3`,
		"el resumen no se muestra en 3 tarjetas (estado, ítems y sprint actual)")
	s.u.click("#ver-graficos")
	s.u.waitTrue(`document.querySelectorAll("#datos-graficos .tarjeta").length === 4`,
		"las métricas de los gráficos no se muestran en 4 tarjetas")
}

// TEC-12 · CA-TEC12.5 · Verifica que las barras y la línea de referencia de los gráficos usen las variables de la paleta.
func TestEstilo_GraficosConLaPaleta(t *testing.T) {
	css := stylesCSS(t)
	vars := paletteVars(t)
	for regla, variable := range map[string]string{
		".grafico .barra-a":    "var(--serie-a)",
		".grafico .barra-b":    "var(--serie-b)",
		".grafico .referencia": "var(--referencia)",
	} {
		inicio := strings.Index(css, regla+" {")
		if inicio == -1 {
			t.Errorf("styles.css no tiene la regla %q", regla)
			continue
		}
		fin := strings.Index(css[inicio:], "}")
		if !strings.Contains(css[inicio:inicio+fin], variable) {
			t.Errorf("%s no usa %s", regla, variable)
		}
		if _, ok := vars[strings.TrimSuffix(strings.TrimPrefix(variable, "var("), ")")]; !ok {
			t.Errorf("la paleta no tiene %s", variable)
		}
	}
}

// TEC-12 · CA-TEC12.6 · Verifica que el CSS no cargue fuentes, imágenes ni hojas de estilo externas.
func TestEstilo_SinRecursosExternos(t *testing.T) {
	css := stylesCSS(t)
	for _, prohibido := range []string{"@import", "url(http", "url(//", "fonts.googleapis"} {
		if strings.Contains(css, prohibido) {
			t.Errorf("styles.css contiene %q", prohibido)
		}
	}
}

// TEC-12 · CA-TEC12.7 · Verifica que el manual de usuario muestre las capturas nuevas de TEC-12 y que existan.
func TestEstilo_ManualConLasCapturasNuevas(t *testing.T) {
	manual, err := os.ReadFile("../../docs/manual-de-usuario.md")
	if err != nil {
		t.Fatal(err)
	}
	capturas := regexp.MustCompile(`img/manual-de-usuario/(tec12-[a-z0-9-]+\.png)`).FindAllStringSubmatch(string(manual), -1)
	if len(capturas) < 2 {
		t.Fatalf("el manual tiene %d capturas de TEC-12, se esperaban al menos 2", len(capturas))
	}
	for _, c := range capturas {
		if _, err := os.Stat("../../docs/img/manual-de-usuario/" + c[1]); err != nil {
			t.Errorf("no existe la captura %s", c[1])
		}
	}
}
