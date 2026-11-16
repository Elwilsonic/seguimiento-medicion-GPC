package httpapi_test

import (
	"regexp"
	"strings"
	"testing"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197), parte
// de sprints y Planning Poker (Buttini Cristobal): interfaz web.
//
// Como en HU-51, los tests no tienen un motor de JavaScript, así que verifican
// el contenido de index.html y app.js; el uso en el navegador se registra en
// la evidencia TDD.

// currentSprintInput devuelve la etiqueta del campo «Sprint actual»: un <input> o, desde
// HU-53, un <select> (D3 de su diseño).
func currentSprintInput(t *testing.T) string {
	t.Helper()
	input := regexp.MustCompile(`<(input|select) id="sprint-actual"[^>]*>`).FindString(contextSection(t))
	if input == "" {
		t.Fatal("la sección de contexto no tiene el campo «Sprint actual»")
	}
	return input
}

// HU-52 · CA-52.6 · Verifica que el campo «Sprint actual» acepte el UUID como texto: type="text", sin min, con placeholder de UUID, sin autocompletar ni corrector.
func TestInterfaz_SprintActualAceptaElUUIDComoTexto(t *testing.T) {
	input := currentSprintInput(t)
	// Un <select> acepta el UUID siempre: es el valor de cada opción (HU-53).
	if strings.HasPrefix(input, "<select") {
		return
	}

	for _, want := range []string{`type="text"`, `placeholder="UUID del sprint"`, `autocomplete="off"`, `spellcheck="false"`} {
		if !strings.Contains(input, want) {
			t.Errorf("«Sprint actual» = %s, no tiene %s", input, want)
		}
	}
	for _, unwanted := range []string{`type="number"`, `min=`} {
		if strings.Contains(input, unwanted) {
			t.Errorf("«Sprint actual» = %s, no debería tener %s", input, unwanted)
		}
	}
}

// HU-52 · CA-52.6 · Verifica que la interfaz no convierta a número el ID del sprint actual ni el de la sesión: crear un sprint y «Elegir» dejan su id en «Sprint actual», y Planning Poker usa el id de la sesión como texto en las rutas y en los mensajes.
func TestInterfaz_NoConvierteElIDDelSprintNiDeLaSesionANumero(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`$("sprint-actual").value = s.id;`,
		`form.dataset.sessionId = s.id;`,
		`"/poker-sessions/" + form.dataset.sessionId + "/votes"`,
		`"Sesión " + sesion.id + " abierta para el ítem " + sesion.item_id + "."`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
	for _, unwanted := range regexp.MustCompile(`(Number|parseInt|parseFloat)\([^)]*(sprintActual|sprint-actual|sessionId|sesion\.id|s\.id)[^)]*\)`).FindAllString(js, -1) {
		t.Errorf("app.js convierte el ID del sprint o de la sesión a número: %s", unwanted)
	}
}
