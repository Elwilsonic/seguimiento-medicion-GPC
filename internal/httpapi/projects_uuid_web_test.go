package httpapi_test

import (
	"regexp"
	"strings"
	"testing"
)

// HU-51 · Identificadores UUID, parte de proyectos (Issue #193): interfaz web.
//
// Los tests no tienen un motor de JavaScript, así que verifican el contenido de
// index.html y app.js; el uso en el navegador se registra en la evidencia TDD.

// currentProjectInput devuelve la etiqueta <input> del campo «Proyecto actual».
func currentProjectInput(t *testing.T) string {
	t.Helper()
	input := regexp.MustCompile(`<input id="proyecto-actual"[^>]*>`).FindString(contextSection(t))
	if input == "" {
		t.Fatal("la sección de contexto no tiene el campo «Proyecto actual»")
	}
	return input
}

// HU-51 · CA-51.6 · Verifica que el campo «Proyecto actual» acepte el UUID como texto: type="text", sin min, con placeholder de UUID, sin autocompletar ni corrector.
func TestInterfaz_ProyectoActualAceptaElUUIDComoTexto(t *testing.T) {
	input := currentProjectInput(t)

	for _, want := range []string{`type="text"`, `placeholder="UUID del proyecto"`, `autocomplete="off"`, `spellcheck="false"`} {
		if !strings.Contains(input, want) {
			t.Errorf("«Proyecto actual» = %s, no tiene %s", input, want)
		}
	}
	for _, unwanted := range []string{`type="number"`, `min=`} {
		if strings.Contains(input, unwanted) {
			t.Errorf("«Proyecto actual» = %s, no debería tener %s", input, unwanted)
		}
	}
}

// HU-51 · CA-51.6 · Verifica que la interfaz no convierta el ID del proyecto a número: usa el texto del campo en las rutas, en «Elegir» y en el nombre del PDF.
func TestInterfaz_NoConvierteElIDDelProyectoANumero(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`const valor = $(inputId).value.trim();`,
		`$("proyecto-actual").value = p.id`,
		`api("GET", "/proyectos/" + proyectoActual())`,
		`enlace.download = "reporte-proyecto-" + id + ".pdf";`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
	for _, unwanted := range regexp.MustCompile(`(Number|parseInt|parseFloat)\([^)]*(proyecto|p\.id)[^)]*\)`).FindAllString(js, -1) {
		t.Errorf("app.js convierte el ID del proyecto a número: %s", unwanted)
	}
}
