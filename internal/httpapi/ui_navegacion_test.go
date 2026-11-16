package httpapi_test

import "testing"

// HU-53 · Navegar la interfaz por secciones y elegir de listas (Issue #211):
// navegación por secciones (CA-53.1). Cada test abre la interfaz en un
// navegador real con los pasos de hu53 (ui_hu53_steps_test.go).

// newHU53Test prepara un escenario de HU-53 para un test de Go; el navegador
// se cierra al terminar el test.
func newHU53Test(t *testing.T) *hu53 {
	t.Helper()
	s := newHU53(t, requireBrowser(t))
	t.Cleanup(s.close)
	return s
}

// HU-53 · CA-53.1 · Verifica que el menú tenga las nueve secciones y que cada una deje su fragmento en la dirección, en orden.
func TestNavegacion_CadaSeccionTieneSuFragmento(t *testing.T) {
	s := newHU53Test(t)
	s.openApp()

	s.eachMenuEntry()

	s.hashesInOrder("#proyectos, #integrantes, #backlog, #sprints, #poker, #esfuerzo, #defectos, #metricas, #reportes")
}

// HU-53 · CA-53.1 · Verifica que elegir una sección en el menú muestre solo esa sección, la marque en el menú y no recargue la página.
func TestNavegacion_IrAUnaSeccionMuestraSoloEsa(t *testing.T) {
	s := newHU53Test(t)
	s.open("#proyectos")

	s.goTo("Backlog")

	s.onlySection("backlog")
	s.hashIs("#backlog")
	s.menuMarked("Backlog")
	s.notReloaded()
}

// HU-53 · CA-53.1 · Verifica que un link directo abra su sección y que el botón "Atrás" vuelva a la anterior.
func TestNavegacion_LinkDirectoYVolverAtras(t *testing.T) {
	s := newHU53Test(t)
	s.open("#sprints")
	s.onlySection("sprints")

	s.goTo("Defectos")
	s.onlySection("defectos")
	s.back()

	s.onlySection("sprints")
}

// HU-53 · CA-53.1 · Verifica que una dirección sin fragmento o con una sección inexistente muestre Proyectos.
func TestNavegacion_SinSeccionMuestraProyectos(t *testing.T) {
	for _, fragment := range []string{"", "#inexistente"} {
		t.Run(fragment, func(t *testing.T) {
			s := newHU53Test(t)

			s.open(fragment)

			s.onlySection("proyectos")
			s.menuMarked("Proyectos")
		})
	}
}

// HU-53 · CA-53.1 · Verifica que una sección que necesita un proyecto avise sin llamar a la API cuando no hay proyecto actual.
func TestNavegacion_SinProyectoAvisaSinLlamarALaAPI(t *testing.T) {
	s := newHU53Test(t)
	s.open("#backlog")

	s.refreshBacklogWithoutProject()

	s.messageIs("backlog", "Elegí un proyecto: creá uno o elegilo en «Proyecto actual».")
	s.noAPICalls()
}
