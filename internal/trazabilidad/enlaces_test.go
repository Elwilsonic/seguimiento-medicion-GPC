package trazabilidad

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// tableRow es una fila de una tabla de la sección "Trazabilidad" de una
// evidencia TDD.
type tableRow struct {
	pos  string // archivo:línea de la evidencia, relativo a la raíz
	path string // ruta de la evidencia en el disco
	text string
}

// traceabilityRows devuelve las filas de tabla de la sección "## Trazabilidad"
// de cada evidencia TDD de historia (docs/evidencias/HU-XX-tdd.md).
func traceabilityRows(t *testing.T, root string) []tableRow {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, "docs", "evidencias", "HU-*-tdd.md"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no se encontraron evidencias TDD: %v", err)
	}
	var rows []tableRow
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("leer %s: %v", f, err)
		}
		rel, _ := filepath.Rel(root, f)
		inSection := false
		for i, line := range strings.Split(string(b), "\n") {
			switch {
			case strings.HasPrefix(line, "## "):
				inSection = strings.HasPrefix(line, "## Trazabilidad")
			case inSection && strings.HasPrefix(line, "|"):
				rows = append(rows, tableRow{
					pos:  fmt.Sprintf("%s:%d", filepath.ToSlash(rel), i+1),
					path: f,
					text: line,
				})
			}
		}
	}
	return rows
}

var (
	// Un link de las tablas: [`Nombre`](ruta) o [`Nombre`](ruta#L12).
	linkPattern = regexp.MustCompile("\\[`([^`]+)`\\]\\(([^)#]+)(?:#L(\\d+))?\\)")
	// Un nombre de test entre comillas invertidas que no es el texto de un link.
	bareTestPattern = regexp.MustCompile("(^|[^\\[])`(Test[A-Za-z0-9_]+)`")
)

// declares dice si la línea declara el nombre: una función o un método de Go,
// un tipo, una variable o una constante (también dentro de un bloque, con su
// "="), o una función de JavaScript. Un campo de un struct no cuenta.
func declares(line, name string) bool {
	n := regexp.QuoteMeta(name)
	return regexp.MustCompile(
		`^func (\([^)]*\) )?` + n + `[\[(]` +
			`|^(type|var|const) ` + n + `\b` +
			`|^\t` + n + `\b[^:]*=` +
			`|^(async )?function ` + n + `\(`).MatchString(line)
}

// TEC-11 · CA-TEC11.7 · Verifica que cada test nombrado en una tabla de trazabilidad de las evidencias TDD tenga su link.
func TestEnlaces_CadaTestDeLasTablasTieneSuLink(t *testing.T) {
	root, _ := repoTests(t)
	for _, row := range traceabilityRows(t, root) {
		for _, m := range bareTestPattern.FindAllStringSubmatch(row.text, -1) {
			t.Errorf("%s: %s no tiene link a su test (%s; si el nombre se repite en varios paquetes, el link se agrega a mano)", row.pos, m[2], updateHint)
		}
	}
}

// TEC-11 · CA-TEC11.7 · Verifica que cada link de las tablas de trazabilidad abra un archivo que existe y, si lleva línea, que esa línea declare lo que el link nombra.
func TestEnlaces_CadaLinkApuntaALaLineaDeLoQueNombra(t *testing.T) {
	root, _ := repoTests(t)
	rows := traceabilityRows(t, root)
	links := 0
	for _, row := range rows {
		for _, m := range linkPattern.FindAllStringSubmatch(row.text, -1) {
			links++
			text, target, lineStr := m[1], m[2], m[3]
			file := filepath.Clean(filepath.Join(filepath.Dir(row.path), filepath.FromSlash(target)))
			b, err := os.ReadFile(file)
			if err != nil {
				t.Errorf("%s: el link de %s apunta a %s, que no existe", row.pos, text, target)
				continue
			}
			if lineStr == "" {
				continue
			}
			n, _ := strconv.Atoi(lineStr)
			lines := strings.Split(string(b), "\n")
			name := declaredName(text)
			if n < 1 || n > len(lines) || !declares(lines[n-1], name) {
				where := "no se encontró en el archivo"
				for i, l := range lines {
					if declares(l, name) {
						where = fmt.Sprintf("está en la línea %d", i+1)
						break
					}
				}
				t.Errorf("%s: el link de %s apunta a %s#L%d, pero ahí no se declara %s (%s; %s)", row.pos, text, target, n, name, where, updateHint)
			}
		}
	}
	if links == 0 {
		t.Error("las tablas de trazabilidad no tienen ningún link")
	}
}

// TEC-11 · CA-TEC11.7 · Verifica que el control de los links reconozca las declaraciones de Go y de JavaScript y no confunda un uso con una declaración.
func TestEnlaces_ReconoceLasDeclaraciones(t *testing.T) {
	cases := []struct {
		line, name string
		want       bool
	}{
		{"func TestGet_ProyectoQueNoEsElPrimero(t *testing.T) {", "TestGet_ProyectoQueNoEsElPrimero", true},
		{"func (m *Manager) Project(projectID int) (ProjectReport, error) {", "Project", true},
		{"func storiesOf(items []backlog.Item, keep func(backlog.Item) bool) Stories {", "storiesOf", true},
		{"type Stories struct {", "Stories", true},
		{"var ErrProjectNotFound = errors.New(\"no existe\")", "ErrProjectNotFound", true},
		{"\tErrEmptyDescription = errors.New(\"la descripción es obligatoria\")", "ErrEmptyDescription", true},
		{"\tStatusOpen Status = \"ABIERTO\"", "StatusOpen", true},
		{"function filaHistoriaReporte(h) {", "filaHistoriaReporte", true},
		{"async function cargarProyecto() {", "cargarProyecto", true},
		{"\tp, err := manager.Project(1)", "Project", false},
		{"\tProject project.Project", "Project", false},
		{"func TestGet_ProyectoQueNoEsElPrimeroBis(t *testing.T) {", "TestGet_ProyectoQueNoEsElPrimero", false},
	}
	for _, c := range cases {
		if got := declares(c.line, c.name); got != c.want {
			t.Errorf("declares(%q, %q) = %v, se esperaba %v", c.line, c.name, got, c.want)
		}
	}
}
