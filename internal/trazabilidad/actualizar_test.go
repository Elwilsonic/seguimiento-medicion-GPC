package trazabilidad

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// updateLinks activa la actualización de los links de las tablas de
// trazabilidad. Se usa así:
//
//	go test ./internal/trazabilidad/ -run TestEnlaces -args -update
var updateLinks = flag.Bool("update", false, "actualiza los links de las tablas de trazabilidad de las evidencias TDD")

// updateHint es lo que se le indica a quien rompe un link al mover un test o
// una función.
const updateHint = "para actualizarlo, corré: go test ./internal/trazabilidad/ -run TestEnlaces -args -update"

func TestMain(m *testing.M) {
	flag.Parse()
	if *updateLinks {
		if err := updateEvidences(); err != nil {
			fmt.Fprintln(os.Stderr, "no se pudieron actualizar las evidencias:", err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

// updateEvidences reescribe las tablas de trazabilidad de todas las
// evidencias TDD con fixRow y guarda las que cambian.
func updateEvidences() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	tests, err := collectTests(root)
	if err != nil {
		return err
	}
	byName := map[string][]string{} // nombre del test -> archivos, relativos a la raíz
	for _, tf := range tests {
		file := tf.pos[:strings.LastIndex(tf.pos, ":")]
		byName[tf.name] = append(byName[tf.name], file)
	}
	files, err := filepath.Glob(filepath.Join(root, "docs", "evidencias", "HU-*-tdd.md"))
	if err != nil {
		return err
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		lines := strings.Split(string(b), "\n")
		inSection, changed := false, false
		for i, line := range lines {
			switch {
			case strings.HasPrefix(line, "## "):
				inSection = strings.HasPrefix(line, "## Trazabilidad")
			case inSection && strings.HasPrefix(line, "|"):
				if fixed := fixRow(line, root, filepath.Dir(f), byName); fixed != line {
					lines[i], changed = fixed, true
				}
			}
		}
		if changed {
			if err := os.WriteFile(f, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, f)
			fmt.Println("actualizada:", filepath.ToSlash(rel))
		}
	}
	return nil
}

// fixRow corrige una fila de una tabla de trazabilidad:
//   - en cada link con línea, pone la línea donde hoy se declara lo que nombra;
//   - a cada test nombrado sin link le agrega el link, si su nombre es único en
//     el repositorio (si hay varios, lo deja para resolverlo a mano).
//
// root es la raíz del repositorio, dir la carpeta de la evidencia y byName,
// los archivos (relativos a la raíz) de cada test por nombre.
func fixRow(row, root, dir string, byName map[string][]string) string {
	row = linkPattern.ReplaceAllStringFunc(row, func(link string) string {
		m := linkPattern.FindStringSubmatch(link)
		text, target, lineStr := m[1], m[2], m[3]
		if lineStr == "" {
			return link
		}
		path, name := filepath.Join(dir, filepath.FromSlash(target)), declaredName(text)
		if current, _ := strconv.Atoi(lineStr); declaresAt(path, current, name) {
			return link // sigue bien: no se toca
		}
		n := declarationLine(path, name)
		if n == 0 {
			return link // no se encontró: lo informa el test
		}
		return fmt.Sprintf("[`%s`](%s#L%d)", text, target, n)
	})
	return bareTestPattern.ReplaceAllStringFunc(row, func(s string) string {
		m := bareTestPattern.FindStringSubmatch(s)
		prefix, name := m[1], m[2]
		files := byName[name]
		if len(files) != 1 {
			return s
		}
		path := filepath.Join(root, filepath.FromSlash(files[0]))
		n := declarationLine(path, name)
		rel, err := filepath.Rel(dir, path)
		if n == 0 || err != nil {
			return s
		}
		return fmt.Sprintf("%s[`%s`](%s#L%d)", prefix, name, filepath.ToSlash(rel), n)
	})
}

// declaredName es el nombre que declara el texto de un link: "Manager.Project"
// o "Project()" nombran a Project.
func declaredName(text string) string {
	name := strings.TrimSuffix(text, "()")
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// declaresAt dice si la línea n del archivo declara el nombre.
func declaresAt(path string, n int, name string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	lines := strings.Split(string(b), "\n")
	return n >= 1 && n <= len(lines) && declares(lines[n-1], name)
}

// declarationLine devuelve la primera línea del archivo que declara el nombre,
// o 0 si no hay ninguna o el archivo no se puede leer.
func declarationLine(path, name string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	for i, l := range strings.Split(string(b), "\n") {
		if declares(l, name) {
			return i + 1
		}
	}
	return 0
}

// TEC-11 · CA-TEC11.7 · Verifica que -update corrija la línea de un link corrido y agregue el link de un test nuevo de nombre único, sin tocar uno repetido.
func TestEnlaces_ActualizarCorrigeLineasYAgregaLinks(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "evidencias")
	write := func(rel, content string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("internal/x/x.go", "package x\n\n// Comentario agregado.\nfunc Hacer() {}\n")
	write("internal/x/x_test.go", "package x\n\nimport \"testing\"\n\n// Comentario agregado.\nfunc TestViejo(t *testing.T) {}\n\nfunc TestNuevo(t *testing.T) {}\n")
	byName := map[string][]string{
		"TestViejo":    {"internal/x/x_test.go"},
		"TestNuevo":    {"internal/x/x_test.go"},
		"TestRepetido": {"internal/x/x_test.go", "internal/y/y_test.go"},
	}
	row := "| CA-99.1 | [`TestViejo`](../../internal/x/x_test.go#L4), `TestNuevo`, `TestRepetido` | [`x.go`](../../internal/x/x.go) ([`Hacer`](../../internal/x/x.go#L2)) |"
	want := "| CA-99.1 | [`TestViejo`](../../internal/x/x_test.go#L6), [`TestNuevo`](../../internal/x/x_test.go#L8), `TestRepetido` | [`x.go`](../../internal/x/x.go) ([`Hacer`](../../internal/x/x.go#L4)) |"
	if got := fixRow(row, root, dir, byName); got != want {
		t.Errorf("fixRow:\n got  %s\n want %s", got, want)
	}
	if again := fixRow(want, root, dir, byName); again != want {
		t.Errorf("fixRow no es idempotente:\n got  %s\n want %s", again, want)
	}
}
