package httpapi_test

import (
	"regexp"
	"strings"
	"testing"
)

// HU-52 · Identificadores UUID del resto de las entidades, parte de defectos y
// esfuerzos (Issue #197): interfaz web.
//
// Los tests no tienen un motor de JavaScript, así que verifican el contenido de
// index.html y app.js; el uso en el navegador se registra en la evidencia TDD.

// formInput devuelve la etiqueta del campo name dentro del formulario formID
// de la página: un <input> o, desde HU-53, un <select> (D9 de su diseño).
func formInput(t *testing.T, page, formID, name string) string {
	t.Helper()
	inicio := strings.Index(page, `<form id="`+formID+`"`)
	if inicio == -1 {
		t.Fatalf("la página no tiene el formulario %q", formID)
	}
	fin := strings.Index(page[inicio:], "</form>")
	if fin == -1 {
		t.Fatalf("el formulario %q no se cierra", formID)
	}
	input := regexp.MustCompile(`<(input|select) name="` + name + `"[^>]*>`).FindString(page[inicio : inicio+fin])
	if input == "" {
		t.Fatalf("el formulario %q no tiene el campo %q", formID, name)
	}
	return input
}

// expectUUIDTextInput verifica que el campo acepte un UUID como texto. Un
// <select> lo acepta siempre: el valor de cada opción es el UUID (HU-53). Un
// <input> tiene type="text", un placeholder que nombra el UUID, sin
// autocompletar ni corrector, y sin type="number" ni min.
func expectUUIDTextInput(t *testing.T, label, input string) {
	t.Helper()
	if strings.HasPrefix(input, "<select") {
		return
	}
	for _, want := range []string{`type="text"`, `autocomplete="off"`, `spellcheck="false"`} {
		if !strings.Contains(input, want) {
			t.Errorf("%s = %s, no tiene %s", label, input, want)
		}
	}
	if !regexp.MustCompile(`placeholder="[^"]*UUID[^"]*"`).MatchString(input) {
		t.Errorf("%s = %s, se esperaba un placeholder que nombre el UUID", label, input)
	}
	for _, unwanted := range []string{`type="number"`, `min=`} {
		if strings.Contains(input, unwanted) {
			t.Errorf("%s = %s, no debería tener %s", label, input, unwanted)
		}
	}
}

// HU-52 · CA-52.6 · Verifica que los 3 campos de ID de la sección de esfuerzos («Ítem (ID)», «Esfuerzo (ID)» e «ID» de la consulta) acepten el UUID como texto.
func TestInterfaz_CamposDeIDDeEsfuerzosAceptanElUUIDComoTexto(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	for _, f := range []struct{ form, name, label string }{
		{"form-registrar-esfuerzo", "item_id", "«Ítem (ID)» de Registrar esfuerzo"},
		{"form-corregir-esfuerzo", "esfuerzo_id", "«Esfuerzo (ID)» de Corregir esfuerzo"},
		{"form-consultar-esfuerzos", "referencia", "«ID» de Consultar esfuerzos"},
	} {
		expectUUIDTextInput(t, f.label, formInput(t, page, f.form, f.name))
	}
}

// HU-52 · CA-52.6 · Verifica que los 4 campos de ID de la sección de defectos («Ítem (ID)», «Sprint de detección (ID)», «Defecto (ID)» y «Sprint de resolución (ID)») acepten el UUID como texto.
func TestInterfaz_CamposDeIDDeDefectosAceptanElUUIDComoTexto(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	for _, f := range []struct{ form, name, label string }{
		{"form-registrar-defecto", "item_id", "«Ítem (ID)» de Registrar defecto"},
		{"form-registrar-defecto", "sprint_deteccion_id", "«Sprint de detección (ID)»"},
		{"form-cambiar-estado-defecto", "defecto_id", "«Defecto (ID)»"},
		{"form-cambiar-estado-defecto", "sprint_resolucion_id", "«Sprint de resolución (ID)»"},
	} {
		expectUUIDTextInput(t, f.label, formInput(t, page, f.form, f.name))
	}
}

// HU-52 · CA-52.6 · Verifica que los campos obligatorios de defectos y de la consulta de esfuerzos sigan siendo obligatorios después de pasar a texto.
func TestInterfaz_CamposDeIDObligatoriosSiguenSiendoObligatorios(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	for _, f := range []struct{ form, name string }{
		{"form-registrar-defecto", "sprint_deteccion_id"},
		{"form-cambiar-estado-defecto", "defecto_id"},
		{"form-consultar-esfuerzos", "referencia"},
	} {
		if input := formInput(t, page, f.form, f.name); !strings.Contains(input, "required") {
			t.Errorf("%s de %s = %s, debería ser obligatorio", f.name, f.form, input)
		}
	}
}

// HU-52 · CA-52.6 · Verifica que app.js envíe los sprints de detección y de resolución como texto, sin convertirlos a número, y que no convierta a número ninguno de los 7 IDs de esfuerzos y defectos.
func TestInterfaz_NoConvierteLosIDsDeDefectosYEsfuerzosANumero(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`sprint_deteccion_id: form.sprint_deteccion_id.value.trim(),`,
		`cuerpo.sprint_resolucion_id = sprint;`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
	conversion := regexp.MustCompile(`(Number|parseInt|parseFloat)\((form\.(item_id|esfuerzo_id|referencia|sprint_deteccion_id|defecto_id|sprint_resolucion_id)[^)]*|sprint|item|id|referencia|defecto)\)`)
	for _, found := range conversion.FindAllString(js, -1) {
		t.Errorf("app.js convierte un ID de esfuerzos o defectos a número: %s", found)
	}
}
