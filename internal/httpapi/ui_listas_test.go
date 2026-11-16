package httpapi_test

import "testing"

// HU-53 · Navegar la interfaz por secciones y elegir de listas (Issue #211):
// proyecto y sprint actuales (CA-53.2) y desplegables de los formularios
// (CA-53.3), en un navegador real.

// HU-53 · CA-53.2 · Verifica que «Proyecto actual» ofrezca los proyectos por nombre, que elegir uno llene «Sprint actual» con nombre y estado, y que lo elegido se vea en todas las secciones.
func TestProyectoActual_ElegirProyectoYSprint(t *testing.T) {
	s := newHU53Test(t)
	s.existingProjects("Sistema de ventas", "Inventario")
	s.closedAndOpenSprints("Sistema de ventas", "Sprint 1", "Sprint 2")
	s.openApp()

	s.offers("«Proyecto actual»", "Inventario, Sistema de ventas")
	s.chooseIn("Sistema de ventas", "«Proyecto actual»")
	s.offers("«Sprint actual»", "Sprint 1 (CERRADO), Sprint 2 (ABIERTO)")
	s.chooseIn("Sprint 2 (ABIERTO)", "«Sprint actual»")
	s.goTo("Backlog")

	s.shows("«Proyecto actual»", "Sistema de ventas")
	s.shows("«Sprint actual»", "Sprint 2 (ABIERTO)")
}

// HU-53 · CA-53.2 · Verifica que cambiar a un proyecto sin sprints deje «Sprint actual» vacío, con su aviso, y sin sprint actual.
func TestProyectoActual_CambiarAUnProyectoSinSprints(t *testing.T) {
	s := newHU53Test(t)
	s.existingProjects("Sistema de ventas", "Inventario")
	s.closedAndOpenSprints("Sistema de ventas", "Sprint 1", "Sprint 2")
	s.openApp()
	s.chooseIn("Sistema de ventas", "«Proyecto actual»")
	s.chooseIn("Sprint 2 (ABIERTO)", "«Sprint actual»")

	s.chooseIn("Inventario", "«Proyecto actual»")

	s.shows("«Sprint actual»", "Sin sprints: creá uno en Sprints")
	s.noCurrent("«Sprint actual»")
}

// HU-53 · CA-53.2 · Verifica que crear un proyecto desde la sección Proyectos lo agregue al desplegable en su lugar por nombre y lo deje elegido.
func TestProyectoActual_CrearUnProyectoLoDejaElegido(t *testing.T) {
	s := newHU53Test(t)
	s.existingProjects("Sistema de ventas", "Inventario")
	s.open("#proyectos")

	s.createProjectFromSection("Agenda")

	s.offers("«Proyecto actual»", "Agenda, Inventario, Sistema de ventas")
	s.shows("«Proyecto actual»", "Agenda")
}

// HU-53 · CA-53.2 · Verifica que sin proyectos «Proyecto actual» muestre su aviso y no haya proyecto actual.
func TestProyectoActual_SinProyectos(t *testing.T) {
	s := newHU53Test(t)

	s.openApp()

	s.shows("«Proyecto actual»", "Sin proyectos: creá uno en Proyectos")
	s.noCurrent("«Proyecto actual»")
}

// HU-53 · CA-53.2 · Verifica que si la lista de proyectos falla, el encabezado muestre el error de la API y el desplegable quede sin proyectos.
func TestProyectoActual_FallaLaLecturaDeLosProyectos(t *testing.T) {
	s := newHU53Test(t)
	s.projectsFail("no se pudieron leer los proyectos")

	s.openApp()

	s.headerMessageIs("no se pudieron leer los proyectos")
	s.offersNothing("«Proyecto actual»")
}

// HU-53 · CA-53.3 · Verifica que cada desplegable de los formularios ofrezca, por su nombre, los ítems del sprint o del proyecto, los sprints, los integrantes, los defectos y los esfuerzos.
func TestListas_QueOfreceCadaDesplegable(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.itemWithoutSprint("Sistema de ventas", "Consultar stock")
	s.member("Sistema de ventas", "Ana", "PRODUCT_BUILDER")
	s.effort("Ana", 3, "Registrar venta")
	s.defect("Registrar venta", "El total no suma")

	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	for _, c := range []struct{ field, options string }{
		{"«Ítem» de Registrar esfuerzo", "Registrar venta"},
		{"«Ítem» de Registrar defecto", "Consultar stock, Registrar venta"},
		{"«Sprint de detección» de Registrar defecto", "Sprint 2 (ABIERTO)"},
		{"«Sprint de resolución» de Cambiar estado", "Sprint 2 (ABIERTO)"},
		{"«Integrante» de Modificar integrante", "Ana (Product Builder)"},
		{"«Defecto» de Cambiar estado", "El total no suma (ABIERTO)"},
		{"«Esfuerzo» de Corregir esfuerzo", "2026-10-02 · Ana · Desarrollo · 3 h"},
		{"«Ítem» de Consultar esfuerzos", "Consultar stock, Registrar venta"},
	} {
		s.offers(c.field, c.options)
	}
}

// HU-53 · CA-53.3 · Verifica que se registre un esfuerzo eligiendo el ítem del desplegable y que el esfuerzo nuevo aparezca en «Esfuerzo» de Corregir esfuerzo.
func TestListas_RegistrarEsfuerzoEligiendoElItem(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.registerEffort("Ana", 3, "Registrar venta")

	s.messageIs("esfuerzo", "Esfuerzo registrado para «Registrar venta».")
	s.offers("«Esfuerzo» de Corregir esfuerzo", "2026-10-02 · Ana · Desarrollo · 3 h")
}

// HU-53 · CA-53.3 · Verifica que se registre un defecto eligiendo el ítem y el sprint de detección, y que aparezca en «Defecto» de Cambiar estado.
func TestListas_RegistrarDefectoEligiendoItemYSprint(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.registerDefect("El total no suma", "Registrar venta", "Sprint 2 (ABIERTO)")

	s.messageIs("defectos", "Defecto registrado para «Registrar venta».")
	s.offers("«Defecto» de Cambiar estado", "El total no suma (ABIERTO)")
}

// HU-53 · CA-53.3 · Verifica que sin ítems en el sprint actual el desplegable muestre su aviso y registrar el esfuerzo avise sin llamar a la API.
func TestListas_SinItemsEnElSprintAvisaSinLlamarALaAPI(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithoutItems("Sistema de ventas", "Sprint 2")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")
	s.shows("«Ítem» de Registrar esfuerzo", "Sin ítems en el sprint actual")

	s.tryRegisterEffort()

	s.messageIs("esfuerzo", "Sin ítems en el sprint actual")
	s.noAPICalls()
}

// HU-53 · CA-53.3 · Verifica que si la API rechaza el defecto por un sprint cerrado, se muestre su error sin limpiar el formulario.
func TestListas_LaAPIRechazaElDefectoEnUnSprintCerrado(t *testing.T) {
	s := newHU53Test(t)
	s.closedSprint("Sistema de ventas", "Sprint 0")
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.registerDefect("El total no suma", "Registrar venta", "Sprint 0 (CERRADO)")

	s.errorIs("defectos", "el sprint está cerrado")
	s.defectDescriptionIs("El total no suma")
}
