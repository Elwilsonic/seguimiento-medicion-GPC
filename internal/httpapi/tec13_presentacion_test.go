package httpapi_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TEC-13 · Pulir la presentación de cada sección de la interfaz (Issue #217).

// mainSections devuelve el HTML de cada <section id="…"> de <main>, por id.
func mainSections(t *testing.T, page string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, m := range regexp.MustCompile(`(?s)<section id="([a-z]+)">(.*?)</section>`).FindAllStringSubmatch(page, -1) {
		out[m[1]] = m[2]
	}
	if len(out) == 0 {
		t.Fatal("la página no tiene secciones")
	}
	return out
}

// TEC-13 · CA-TEC13.1 · Verifica que ningún texto de la interfaz pida escribir o cargar un ID.
func TestPresentacion_NingunTextoPideUnID(t *testing.T) {
	h := newHandler()
	pedido := regexp.MustCompile(`(?i)(cargá|escribí|ingresá|pegá)[^."<]*\bsu ID\b`)
	for _, path := range []string{"/", "/app.js"} {
		if m := pedido.FindAllString(get(t, h, path).Body.String(), -1); len(m) != 0 {
			t.Errorf("%s pide escribir un ID: %q", path, m)
		}
	}
}

// TEC-13 · CA-TEC13.1 · Verifica que, sin proyecto elegido, «Sprint actual» diga "Elegí un proyecto primero".
func TestPresentacion_SprintActualSinProyecto(t *testing.T) {
	s := newHU53Test(t)
	s.existingProjects("Sistema de ventas", "Inventario")

	s.openApp()

	s.shows("«Sprint actual»", "Elegí un proyecto primero")
}

// TEC-13 · CA-TEC13.2 · Verifica que cada formulario de las secciones tenga su título y que el CSS los muestre como tarjetas.
func TestPresentacion_FormulariosComoTarjetas(t *testing.T) {
	for id, html := range mainSections(t, get(t, newHandler(), "/").Body.String()) {
		for _, form := range regexp.MustCompile(`(?s)<form id="([a-z-]+)"[^>]*>(.*?)</form>`).FindAllStringSubmatch(html, -1) {
			if !strings.Contains(form[2], "<h3>") {
				t.Errorf("el formulario %s de la sección %s no tiene título (<h3>)", form[1], id)
			}
		}
	}
	css := stylesCSS(t)
	inicio := strings.Index(css, "section form {")
	if inicio == -1 {
		t.Fatal("styles.css no tiene la regla de las tarjetas de formulario (section form)")
	}
	bloque := css[inicio : inicio+strings.Index(css[inicio:], "}")]
	for _, want := range []string{"border", "var(--radio)", "background"} {
		if !strings.Contains(bloque, want) {
			t.Errorf("la regla de los formularios no tiene %q: %s", want, bloque)
		}
	}
}

// TEC-13 · CA-TEC13.3 · Verifica que cada sección tenga su título seguido de una línea que dice para qué sirve.
func TestPresentacion_EncabezadoDeCadaSeccion(t *testing.T) {
	encabezado := regexp.MustCompile(`(?s)^\s*<h2>[^<]+</h2>\s*<p class="descripcion">([^<]{20,})</p>`)
	for id, html := range mainSections(t, get(t, newHandler(), "/").Body.String()) {
		if !encabezado.MatchString(html) {
			t.Errorf("la sección %s no empieza con <h2> y una <p class=\"descripcion\"> de al menos 20 caracteres", id)
		}
	}
}

// TEC-13 · CA-TEC13.4 · Verifica que los estados se muestren como etiquetas en las tablas de proyectos, ítems, sprints y defectos.
func TestPresentacion_EstadosComoEtiquetas(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.defect("Registrar venta", "El total no suma")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.u.click("#ver-proyectos")
	s.u.waitTrue(`document.querySelector("#lista-proyectos .etiqueta.estado-PLANIFICADO") !== null`, "el estado del proyecto no es una etiqueta")
	for _, c := range []struct{ section, tbody, estado string }{
		{"backlog", "#items", "PENDIENTE"},
		{"defectos", "#lista-defectos", "ABIERTO"},
	} {
		s.showTable(c.section)
		s.u.waitTrue(`document.querySelector("`+c.tbody+` .etiqueta.estado-`+c.estado+`") !== null`, "el estado "+c.estado+" no es una etiqueta en "+c.section)
	}
	s.goTo("Sprints")
	s.u.click("#ver-sprints")
	s.u.waitTrue(`document.querySelector("#lista-sprints .etiqueta.estado-ABIERTO") !== null`, "el estado del sprint no es una etiqueta")
}

// TEC-13 · CA-TEC13.4 · Verifica que el texto de cada etiqueta de estado tenga un contraste de al menos 4.5:1 con su fondo.
func TestPresentacion_ContrasteDeLasEtiquetas(t *testing.T) {
	vars := paletteVars(t)
	for _, tono := range []string{"neutro", "info", "ok", "error", "aviso"} {
		texto, ok1 := vars["--etiqueta-"+tono+"-texto"]
		fondo, ok2 := vars["--etiqueta-"+tono+"-fondo"]
		if !ok1 || !ok2 {
			t.Errorf("la paleta no tiene --etiqueta-%s-texto o --etiqueta-%s-fondo", tono, tono)
			continue
		}
		if c := contrast(texto, fondo); c < 4.5 {
			t.Errorf("etiqueta %s: contraste %.2f:1, se esperaba al menos 4.5:1", tono, c)
		}
	}
}

// TEC-13 · CA-TEC13.5 · Verifica que cada entrada del menú tenga su ícono SVG dentro de la página, sin imágenes ni recursos externos.
func TestPresentacion_IconosDelMenu(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()
	inicio := strings.Index(page, `<nav id="menu"`)
	if inicio == -1 {
		t.Fatal("la página no tiene el menú")
	}
	menu := page[inicio : inicio+strings.Index(page[inicio:], "</nav>")]
	enlaces := regexp.MustCompile(`(?s)<a href="#[a-z]+">(.*?)</a>`).FindAllStringSubmatch(menu, -1)
	if len(enlaces) != 9 {
		t.Fatalf("el menú tiene %d entradas, se esperaban 9", len(enlaces))
	}
	for _, e := range enlaces {
		if !strings.Contains(e[1], "<svg") || !strings.Contains(e[1], `aria-hidden="true"`) {
			t.Errorf("la entrada %q no tiene un ícono SVG oculto para lectores de pantalla", e[0])
		}
	}
	for _, prohibido := range []string{"<img", "src=", "href=\"http"} {
		if strings.Contains(menu, prohibido) {
			t.Errorf("el menú contiene %q", prohibido)
		}
	}
}

// TEC-13 · CA-TEC13.6 · Verifica que el botón que envía cada formulario quede al pie, alineado a la derecha.
func TestPresentacion_BotonAlPieDerecha(t *testing.T) {
	css := stylesCSS(t)
	inicio := strings.Index(css, `section form > button[type="submit"]`)
	if inicio == -1 {
		t.Fatal(`styles.css no tiene la regla "section form > button[type="submit"]"`)
	}
	bloque := css[inicio : inicio+strings.Index(css[inicio:], "}")]
	if !strings.Contains(bloque, "align-self: flex-end") {
		t.Errorf("el botón de los formularios no se alinea a la derecha: %s", bloque)
	}
}

// TEC-13 · CA-TEC13.7 · Verifica que el manual de usuario tenga una captura de TEC-13 y que exista.
func TestPresentacion_ManualConCapturaNueva(t *testing.T) {
	manual, err := os.ReadFile("../../docs/manual-de-usuario.md")
	if err != nil {
		t.Fatal(err)
	}
	capturas := regexp.MustCompile(`img/manual-de-usuario/(tec13-[a-z0-9-]+\.png)`).FindAllStringSubmatch(string(manual), -1)
	if len(capturas) == 0 {
		t.Fatal("el manual no tiene capturas de TEC-13")
	}
	for _, c := range capturas {
		if _, err := os.Stat("../../docs/img/manual-de-usuario/" + c[1]); err != nil {
			t.Errorf("no existe la captura %s", c[1])
		}
	}
}
