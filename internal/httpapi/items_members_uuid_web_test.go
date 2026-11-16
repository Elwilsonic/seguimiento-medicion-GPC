package httpapi_test

import (
	"regexp"
	"strings"
	"testing"
)

// HU-52 · Identificadores UUID del resto de las entidades, parte de ítems e
// integrantes (Issue #197): interfaz web, secciones de Product Backlog e
// Integrantes (D3). Los tests no tienen un motor de JavaScript, así que
// verifican el contenido de index.html y app.js; el uso en el navegador se
// registra en la evidencia TDD.

// memberIDInput devuelve la etiqueta del campo «ID del integrante» del
// formulario "Modificar integrante": un <input> o, desde HU-53, un <select>.
func memberIDInput(t *testing.T) string {
	t.Helper()
	page := get(t, newHandler(), "/").Body.String()
	inicio := strings.Index(page, `<form id="form-modificar-integrante">`)
	if inicio == -1 {
		t.Fatal("la página no contiene el formulario de modificar integrante")
	}
	input := regexp.MustCompile(`<(input|select) name="id"[^>]*>`).FindString(page[inicio:])
	if input == "" {
		t.Fatal("el formulario de modificar integrante no tiene el campo «ID del integrante»")
	}
	return input
}

// HU-52 · CA-52.6 · Verifica que el campo «ID del integrante» acepte el UUID como texto: type="text", sin min, con placeholder de UUID, sin autocompletar ni corrector.
func TestInterfaz_IDDelIntegranteAceptaElUUIDComoTexto(t *testing.T) {
	input := memberIDInput(t)
	if !strings.Contains(input, "required") {
		t.Errorf("«ID del integrante» = %s, debería ser obligatorio", input)
	}
	// Un <select> acepta el UUID siempre: es el valor de cada opción (HU-53).
	if strings.HasPrefix(input, "<select") {
		return
	}

	for _, want := range []string{`type="text"`, `placeholder="UUID del integrante"`, `autocomplete="off"`, `spellcheck="false"`} {
		if !strings.Contains(input, want) {
			t.Errorf("«ID del integrante» = %s, no tiene %s", input, want)
		}
	}
	for _, unwanted := range []string{`type="number"`, `min=`} {
		if strings.Contains(input, unwanted) {
			t.Errorf("«ID del integrante» = %s, no debería tener %s", input, unwanted)
		}
	}
}

// HU-52 · CA-52.6 · Verifica que la fila del backlog muestre "—" para un ítem sin sprint, que ahora llega como "" y no como 0 (D7).
func TestInterfaz_BacklogMuestraSinSprintConElTextoVacio(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	if want := `celda(it.sprint_id === "" ? "—" : it.sprint_id)`; !strings.Contains(js, want) {
		t.Errorf("app.js no contiene %q", want)
	}
	if unwanted := `celda(it.sprint_id === 0 ? "—" : it.sprint_id)`; strings.Contains(js, unwanted) {
		t.Errorf("app.js todavía compara el sprint del ítem con 0: %q", unwanted)
	}
}

// HU-52 · CA-52.6 · Verifica que los botones del backlog y el formulario de integrantes usen el id como texto en las rutas, sin convertirlo a número.
func TestInterfaz_NoConvierteLosIDsDeItemsEIntegrantesANumero(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`boton("Editar", () => editarItem(it.id), "secundario")`,
		`api("GET", "/items/" + id)`,
		`api("PUT", "/items/" + form.dataset.id, {`,
		`api("PATCH", "/items/" + it.id + "/estado", { estado: estado.value })`,
		`api("PATCH", "/items/" + it.id + "/story-points", { story_points: valor })`,
		`api("PUT", "/integrantes/" + form.id.value, {`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
	for _, unwanted := range regexp.MustCompile(`(Number|parseInt|parseFloat)\([^)]*(it\.id|dataset\.id|form\.id)[^)]*\)`).FindAllString(js, -1) {
		t.Errorf("app.js convierte el ID del ítem o del integrante a número: %s", unwanted)
	}
}
