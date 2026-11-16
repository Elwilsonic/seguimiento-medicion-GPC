package httpapi_test

import (
	"net/http"
	"testing"
)

// HU-53 · Navegar la interfaz por secciones y elegir de listas (Issue #211):
// acciones de fila (CA-53.4), mensajes (CA-53.5), pantallas chicas (CA-53.6)
// y sin cambios en la API (CA-53.7), en un navegador real.

// HU-53 · CA-53.4 · Verifica que cada fila del backlog, de esfuerzos, de defectos y de integrantes tenga sus acciones.
func TestAcciones_CadaTablaTieneSusAcciones(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.member("Sistema de ventas", "Ana", "PRODUCT_BUILDER")
	s.effort("Ana", 3, "Registrar venta")
	s.defect("Registrar venta", "El total no suma")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.rowsHave("backlog", "Registrar esfuerzo, Registrar defecto")
	s.rowsHave("esfuerzo", "Corregir horas")
	s.rowsHave("defectos", "Cambiar estado")
	s.rowsHave("integrantes", "Modificar")
}

// HU-53 · CA-53.4 · Verifica que "Registrar defecto" en una fila del backlog lleve a Defectos con ese ítem elegido.
func TestAcciones_RegistrarDefectoDesdeElBacklog(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.itemWithoutSprint("Sistema de ventas", "Consultar stock")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.backlogRowAction("Registrar defecto", "Consultar stock")

	s.onlySection("defectos")
	s.shows("«Ítem» de Registrar defecto", "Consultar stock")
}

// HU-53 · CA-53.4 · Verifica que "Modificar" en la fila de un integrante lo deje elegido con su nombre y su rol cargados.
func TestAcciones_ModificarIntegranteDesdeSuFila(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.member("Sistema de ventas", "Ana", "PRODUCT_BUILDER")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.memberRowAction("Modificar", "Ana")

	s.shows("«Integrante» de Modificar integrante", "Ana (Product Builder)")
	s.memberFormHas("Ana", "PRODUCT_BUILDER")
}

// HU-53 · CA-53.4 · Verifica que "Registrar esfuerzo" desde un ítem que no es del sprint actual avise y no elija ningún ítem.
func TestAcciones_RegistrarEsfuerzoDeUnItemFueraDelSprintAvisa(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.itemWithoutSprint("Sistema de ventas", "Consultar stock")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.backlogRowAction("Registrar esfuerzo", "Consultar stock")

	s.onlySection("esfuerzo")
	s.messageIs("esfuerzo", "El ítem no está en el sprint actual: asignalo o elegí su sprint")
	s.noItemChosen()
}

// HU-53 · CA-53.4 · Verifica que corregir un esfuerzo desde su fila con 0 horas muestre el error de la API.
func TestAcciones_CorregirHorasInvalidasMuestraElError(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.effort("Ana", 3, "Registrar venta")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.correctEffortTo("Ana", 0)

	s.errorIs("esfuerzo", "las horas trabajadas deben ser un número positivo")
}

// HU-53 · CA-53.5 · Verifica que el mensaje de éxito nombre la entidad: "Ítem «Emitir factura» creado.".
func TestMensajes_ExitoConElNombre(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.createItemFromSection("Emitir factura")

	s.messageIs("backlog", "Ítem «Emitir factura» creado.")
}

// HU-53 · CA-53.5 · Verifica que el error de la API se muestre sin cambiar su texto y que el mensaje siguiente lo reemplace.
func TestMensajes_ElNuevoReemplazaAlErrorAnterior(t *testing.T) {
	s := newHU53Test(t)
	s.closedSprint("Sistema de ventas", "Sprint 0")
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")
	s.registerDefect("El total no suma", "Registrar venta", "Sprint 0 (CERRADO)")
	s.errorIs("defectos", "el sprint está cerrado")

	s.registerDefect("El total no suma", "Registrar venta", "Sprint 2 (ABIERTO)")

	s.messageIs("defectos", "Defecto registrado para «Registrar venta».")
}

// HU-53 · CA-53.6 · Verifica que en una ventana de 1280 px el menú se vea al costado del contenido.
func TestPantalla_EscritorioConElMenuAlCostado(t *testing.T) {
	s := newHU53Test(t)
	s.windowWidth(1280)

	s.openApp()

	s.menuBeside()
}

// HU-53 · CA-53.6 · Verifica que en una ventana de 360 px el menú quede arriba, la página no se desplace en horizontal y la tabla del backlog se desplace en su recuadro.
func TestPantalla_CelularSinDesplazamientoHorizontal(t *testing.T) {
	s := newHU53Test(t)
	s.manyItems("Sistema de ventas", 10)
	s.windowWidth(360)

	s.openWithProjectIn("Sistema de ventas", "#backlog")

	s.menuAbove()
	s.noHorizontalScroll()
	s.backlogTableScrolls()
}

// HU-53 · CA-53.7 · Verifica que al recorrer todas las secciones el navegador solo pida recursos al mismo servidor.
func TestSinCambios_LaInterfazNoPideRecursosAOtroDominio(t *testing.T) {
	s := newHU53Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.walkAllSections()

	s.sameOrigin()
}

// HU-53 · CA-53.7 · Verifica que una ruta de la API que no existe siga respondiendo 404.
func TestSinCambios_UnaRutaDeLaAPIInexistenteDa404(t *testing.T) {
	rec := serveRaw(newHandler(), http.MethodGet, "/api/inexistente", "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /api/inexistente = %d, se esperaba 404", rec.Code)
	}
}
