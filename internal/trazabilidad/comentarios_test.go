package trazabilidad

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"unicode"
	"unicode/utf8"
)

// Formato acordado en TEC-11 (docs/evidencias/TEC-11-tests-por-historia.md):
//
//	// HU-XX · CA-XX.N · Verifica que ...   (criterio de aceptación)
//	// HU-XX · RN-XX.N · Verifica que ...   (regla de negocio)
//	// HU-XX · Error · Verifica que ...     (condición de error sin número)
//	// HU-XX · BDD · Verifica que ...       (runner de godog)
//	// TEC-XX · CA-TECXX.N · Verifica que ... (tarea técnica)
var commentFormat = regexp.MustCompile(
	`^// ((HU|TEC)-(\d+)) · (CA-(?:TEC)?(\d+)\.\d+|RN-(\d+)\.\d+|Error|BDD) · Verifica que\b`)

// testFunc es una función de test del repositorio con su comentario.
type testFunc struct {
	pos   string   // archivo:línea, relativo a la raíz del repositorio
	name  string   // nombre de la función
	lines []string // líneas del comentario de arriba de la función
}

// reference es la línea con el formato de TEC-11 de un test.
type reference struct {
	id     string // HU-XX o TEC-XX
	kind   string // HU o TEC
	number string // XX
	ref    string // CA-XX.N, CA-TECXX.N, RN-XX.N, Error o BDD
	refNum string // XX del CA o de la RN; vacío para Error y BDD
}

var (
	scanOnce  sync.Once
	scanRoot  string
	scanTests []testFunc
	scanErr   error
)

// repoTests recorre una sola vez todos los archivos _test.go del repositorio.
func repoTests(t *testing.T) (string, []testFunc) {
	t.Helper()
	scanOnce.Do(func() {
		scanRoot, scanErr = findRoot()
		if scanErr == nil {
			scanTests, scanErr = collectTests(scanRoot)
		}
	})
	if scanErr != nil {
		t.Fatalf("no se pudieron leer los tests del repositorio: %v", scanErr)
	}
	if len(scanTests) == 0 {
		t.Fatal("no se encontró ningún test en el repositorio")
	}
	return scanRoot, scanTests
}

// findRoot sube desde el directorio actual hasta el que tiene go.mod.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no se encontró go.mod")
		}
		dir = parent
	}
}

func collectTests(root string) ([]testFunc, error) {
	var tests []testFunc
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !isTestName(fn.Name.Name) {
				continue
			}
			var lines []string
			if fn.Doc != nil {
				for _, c := range fn.Doc.List {
					lines = append(lines, c.Text)
				}
			}
			p := fset.Position(fn.Pos())
			rel, _ := filepath.Rel(root, p.Filename)
			tests = append(tests, testFunc{
				pos:   fmt.Sprintf("%s:%d", filepath.ToSlash(rel), p.Line),
				name:  fn.Name.Name,
				lines: lines,
			})
		}
		return nil
	})
	return tests, err
}

// isTestName sigue la regla de go test: "Test" seguido de nada o de algo que
// no empieza con minúscula. TestMain no es un test.
func isTestName(name string) bool {
	if !strings.HasPrefix(name, "Test") || name == "TestMain" {
		return false
	}
	rest := name[len("Test"):]
	if rest == "" {
		return true
	}
	r, _ := utf8.DecodeRuneInString(rest)
	return !unicode.IsLower(r)
}

// referenceOf devuelve la primera línea del comentario que tiene el formato.
func referenceOf(tf testFunc) (reference, bool) {
	for _, line := range tf.lines {
		m := commentFormat.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		r := reference{id: m[1], kind: m[2], number: m[3], ref: m[4], refNum: m[5]}
		if m[6] != "" {
			r.refNum = m[6]
		}
		return r, true
	}
	return reference{}, false
}

var (
	specOnce  sync.Once
	specCache map[string]string
	specErr   error
)

// specification devuelve el texto de las especificaciones de una historia o
// tarea, o "" si no tiene ninguna:
//   - HU-XX: su SDD (docs/especificaciones/HU-XX-*.md) y, para los cambios con
//     OpenSpec (TEC-10), las specs de openspec/ que nombran la historia;
//   - TEC-XX: su evidencia (docs/evidencias/TEC-XX*.md), donde están sus criterios.
func specification(t *testing.T, root, id string) string {
	t.Helper()
	specOnce.Do(func() { specCache, specErr = loadSpecs(root) })
	if specErr != nil {
		t.Fatalf("no se pudieron leer las especificaciones: %v", specErr)
	}
	return specCache[id]
}

func loadSpecs(root string) (map[string]string, error) {
	specs := map[string]string{}
	add := func(id, path string) error {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		specs[id] += string(b) + "\n"
		return nil
	}
	idInName := regexp.MustCompile(`^((?:HU|TEC)-\d+)`)
	for _, dir := range []string{"docs/especificaciones", "docs/evidencias"} {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			m := idInName.FindStringSubmatch(e.Name())
			if e.IsDir() || m == nil || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			// De docs/evidencias solo cuentan las de las tareas técnicas.
			if dir == "docs/evidencias" && !strings.HasPrefix(m[1], "TEC-") {
				continue
			}
			if err := add(m[1], filepath.Join(root, dir, e.Name())); err != nil {
				return nil, err
			}
		}
	}
	huInText := regexp.MustCompile(`HU-\d+`)
	err := filepath.WalkDir(filepath.Join(root, "openspec"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return filepath.SkipDir
			}
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, id := range huInText.FindAllString(string(b), -1) {
			if !seen[id] {
				seen[id] = true
				specs[id] += string(b) + "\n"
			}
		}
		return nil
	})
	return specs, err
}

// containsRef dice si el texto nombra la referencia exacta: CA-03.1 no cuenta
// como CA-03.10.
func containsRef(text, ref string) bool {
	return regexp.MustCompile(regexp.QuoteMeta(ref) + `([^0-9]|$)`).MatchString(text)
}

// TEC-11 · CA-TEC11.4 · Verifica que cada función de test del repositorio tenga arriba la línea con su historia (o tarea), lo que prueba y "Verifica que…".
func TestComentarios_CadaTestTieneLaLineaConElFormato(t *testing.T) {
	_, tests := repoTests(t)
	for _, tf := range tests {
		if _, ok := referenceOf(tf); !ok {
			t.Errorf("%s: %s no tiene arriba la línea \"// HU-XX · CA-XX.N · Verifica que …\" (ver docs/evidencias/TEC-11-tests-por-historia.md)", tf.pos, tf.name)
		}
	}
}

// TEC-11 · CA-TEC11.4 · Verifica que la historia o tarea citada por cada test tenga su especificación en el repositorio.
func TestComentarios_LaHistoriaTieneSuEspecificacion(t *testing.T) {
	root, tests := repoTests(t)
	for _, tf := range tests {
		r, ok := referenceOf(tf)
		if !ok {
			continue // lo informa TestComentarios_CadaTestTieneLaLineaConElFormato
		}
		if specification(t, root, r.id) == "" {
			where := "docs/especificaciones/" + r.id + "-*.md ni en openspec/"
			if r.kind == "TEC" {
				where = "docs/evidencias/" + r.id + "*.md"
			}
			t.Errorf("%s: %s cita %s, que no tiene especificación en %s", tf.pos, tf.name, r.id, where)
		}
	}
}

// TEC-11 · CA-TEC11.4 · Verifica que el criterio o la regla citados existan en la especificación de esa historia y lleven su número.
func TestComentarios_ElCriterioExisteYEsDeSuHistoria(t *testing.T) {
	root, tests := repoTests(t)
	for _, tf := range tests {
		r, ok := referenceOf(tf)
		if !ok || r.refNum == "" {
			continue // sin línea, Error o BDD
		}
		if strings.TrimLeft(r.refNum, "0") != strings.TrimLeft(r.number, "0") {
			t.Errorf("%s: %s cita %s, que no es de %s", tf.pos, tf.name, r.ref, r.id)
			continue
		}
		spec := specification(t, root, r.id)
		if spec != "" && !containsRef(spec, r.ref) {
			t.Errorf("%s: %s cita %s, que no aparece en la especificación de %s", tf.pos, tf.name, r.ref, r.id)
		}
	}
}

// TEC-11 · CA-TEC11.4 · Verifica que un test con la etiqueta BDD sea de una historia que tiene su archivo .feature.
func TestComentarios_BDDTieneSuFeature(t *testing.T) {
	root, tests := repoTests(t)
	for _, tf := range tests {
		r, ok := referenceOf(tf)
		if !ok || r.ref != "BDD" {
			continue
		}
		if r.kind != "HU" {
			t.Errorf("%s: %s usa BDD con %s; BDD es solo para historias", tf.pos, tf.name, r.id)
			continue
		}
		matches, err := filepath.Glob(filepath.Join(root, "features", r.id+"-*.feature"))
		if err != nil || len(matches) == 0 {
			t.Errorf("%s: %s usa BDD, pero no existe features/%s-*.feature", tf.pos, tf.name, r.id)
		}
	}
}

// TEC-11 · CA-TEC11.4 · Verifica que el formato acepte las cinco formas acordadas y rechace las que no lo cumplen.
func TestComentarios_FormatoAceptaSoloLasFormasAcordadas(t *testing.T) {
	valid := []string{
		"// HU-03 · CA-03.1 · Verifica que se obtenga el proyecto.",
		"// HU-03 · RN-03.4 · Verifica que consultar no cambie nada.",
		"// HU-03 · Error · Verifica que, si el repositorio falla, se devuelva ese error.",
		"// HU-22 · BDD · Verifica que se cumplan los escenarios de features/HU-22-votar-oculto.feature.",
		"// TEC-04 · CA-TEC04.9 · Verifica que la ruta responda 200.",
	}
	invalid := []string{
		"// CA-03.1",
		"// HU-03 · CA-03.1",
		"// HU-03 · CA-03.1 · Verifica el orden de los errores.",
		"// HU-03 · Cambiar el estado (Issue #16).",
		"// HU-03 - CA-03.1 - Verifica que se obtenga el proyecto.",
	}
	for _, line := range valid {
		if !commentFormat.MatchString(line) {
			t.Errorf("se esperaba válido: %q", line)
		}
	}
	for _, line := range invalid {
		if commentFormat.MatchString(line) {
			t.Errorf("se esperaba inválido: %q", line)
		}
	}
}

// TEC-11 · CA-TEC11.4 · Verifica que un criterio no se confunda con otro que empieza igual (CA-03.1 no es CA-03.10).
func TestComentarios_ReferenciaExactaDelCriterio(t *testing.T) {
	if !containsRef("- **CA-03.1** Se obtiene el proyecto.", "CA-03.1") {
		t.Error("CA-03.1 debería encontrarse")
	}
	if containsRef("- **CA-03.10** Otro criterio.", "CA-03.1") {
		t.Error("CA-03.1 no debería encontrarse dentro de CA-03.10")
	}
}
