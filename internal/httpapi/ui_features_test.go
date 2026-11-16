package httpapi_test

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
)

// HU-53 · BDD · Verifica que los escenarios de features/HU-53-navegacion-por-secciones.feature se cumplan en un navegador real.
func TestFeatures_HU53Navegacion(t *testing.T) {
	browser := requireBrowser(t)
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeHU53Scenario(ctx, browser) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-53-navegacion-por-secciones.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-53")
	}
}

func initializeHU53Scenario(ctx *godog.ScenarioContext, browser string) {
	var s *hu53
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		s = newHU53(stepFailer{}, browser)
		return c, nil
	})
	ctx.After(func(c context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		s.close()
		return c, nil
	})

	// Datos
	ctx.Step(`^que existen los proyectos "([^"]*)" e "([^"]*)"$`, func(a, b string) error { return s.existingProjects(a, b) })
	ctx.Step(`^que "([^"]*)" tiene el sprint cerrado "([^"]*)" y el sprint abierto "([^"]*)"$`, func(p, c, o string) error { return s.closedAndOpenSprints(p, c, o) })
	ctx.Step(`^que "([^"]*)" tiene el sprint cerrado "([^"]*)"$`, func(p, c string) error { return s.closedSprint(p, c) })
	ctx.Step(`^que el proyecto "([^"]*)" tiene el sprint abierto "([^"]*)" con el ítem "([^"]*)"$`, func(p, sp, it string) error { return s.openSprintWithItem(p, sp, it) })
	ctx.Step(`^que el proyecto "([^"]*)" tiene el sprint abierto "([^"]*)" sin ítems$`, func(p, sp string) error { return s.openSprintWithoutItems(p, sp) })
	ctx.Step(`^que el proyecto "([^"]*)" tiene el ítem "([^"]*)" sin sprint$`, func(p, it string) error { return s.itemWithoutSprint(p, it) })
	ctx.Step(`^que el proyecto "([^"]*)" tiene (\d+) ítems$`, func(p string, n int) error { return s.manyItems(p, n) })
	ctx.Step(`^que "([^"]*)" tiene la integrante "([^"]*)" con rol "([^"]*)"$`, func(p, m, r string) error { return s.member(p, m, r) })
	ctx.Step(`^que "([^"]*)" registró (\d+) horas en "([^"]*)"$`, func(m string, h int, it string) error { return s.effort(m, h, it) })
	ctx.Step(`^que "([^"]*)" tiene el defecto "([^"]*)"$`, func(it, d string) error { return s.defect(it, d) })
	ctx.Step(`^que la lista de proyectos responde 500 con el error "([^"]*)"$`, func(msg string) error { return s.projectsFail(msg) })
	ctx.Step(`^que la ventana mide (\d+) px de ancho$`, func(w int) error { return s.windowWidth(w) })

	// Abrir la interfaz
	ctx.Step(`^que abro la aplicación$`, func() error { return s.openApp() })
	ctx.Step(`^que abro la aplicación en "([^"]*)"$`, func(f string) error { return s.open(f) })
	ctx.Step(`^que abro la aplicación con "([^"]*)" y "([^"]*)" elegidos$`, func(p, sp string) error { return s.openWithProjectAndSprint(p, sp) })
	ctx.Step(`^que abro la aplicación con "([^"]*)" elegido en "([^"]*)"$`, func(p, f string) error { return s.openWithProjectIn(p, f) })

	// Acciones
	ctx.Step(`^elijo cada entrada del menú$`, func() error { return s.eachMenuEntry() })
	ctx.Step(`^elijo "([^"]*)" en el menú$`, func(l string) error { return s.goTo(l) })
	ctx.Step(`^vuelvo atrás en el navegador$`, func() error { return s.back() })
	ctx.Step(`^pido actualizar el backlog sin proyecto actual$`, func() error { return s.refreshBacklogWithoutProject() })
	ctx.Step(`^elijo "([^"]*)" en («[^»]*»)$`, func(o, f string) error { return s.chooseIn(o, f) })
	ctx.Step(`^creo el proyecto "([^"]*)" desde la sección Proyectos$`, func(n string) error { return s.createProjectFromSection(n) })
	ctx.Step(`^creo el ítem "([^"]*)" desde la sección Backlog$`, func(n string) error { return s.createItemFromSection(n) })
	ctx.Step(`^registro un esfuerzo de "([^"]*)" de (\d+) horas en "([^"]*)" desde la sección Esfuerzo$`, func(m string, h int, it string) error { return s.registerEffort(m, h, it) })
	ctx.Step(`^intento registrar un esfuerzo desde la sección Esfuerzo$`, func() error { return s.tryRegisterEffort() })
	ctx.Step(`^registro el defecto "([^"]*)" de "([^"]*)" detectado en "([^"]*)"$`, func(d, it, sp string) error { return s.registerDefect(d, it, sp) })
	ctx.Step(`^elijo "([^"]*)" en la fila "([^"]*)" del backlog$`, func(a, it string) error { return s.backlogRowAction(a, it) })
	ctx.Step(`^elijo "([^"]*)" en la fila "([^"]*)" de los integrantes$`, func(a, m string) error { return s.memberRowAction(a, m) })
	ctx.Step(`^elijo "Corregir horas" en la fila del esfuerzo de "([^"]*)" y corrijo a (\d+) horas$`, func(m string, h int) error { return s.correctEffortTo(m, h) })
	ctx.Step(`^recorro todas las secciones del menú$`, func() error { return s.walkAllSections() })
	ctx.Step(`^pido GET "([^"]*)" a la API$`, func(p string) error { return s.apiGet(p) })

	// Verificaciones
	ctx.Step(`^la dirección termina, en orden, en "([^"]*)"$`, func(l string) error { return s.hashesInOrder(l) })
	ctx.Step(`^se ve solo la sección "([^"]*)"$`, func(id string) error { return s.onlySection(id) })
	ctx.Step(`^la dirección termina en "([^"]*)"$`, func(h string) error { return s.hashIs(h) })
	ctx.Step(`^la entrada "([^"]*)" del menú está marcada$`, func(l string) error { return s.menuMarked(l) })
	ctx.Step(`^la página no se recargó$`, func() error { return s.notReloaded() })
	ctx.Step(`^el mensaje de "([^"]*)" es "([^"]*)"$`, func(sec, m string) error { return s.messageIs(sec, m) })
	ctx.Step(`^el mensaje de "([^"]*)" es el error "([^"]*)"$`, func(sec, m string) error { return s.errorIs(sec, m) })
	ctx.Step(`^el mensaje del encabezado es "([^"]*)"$`, func(m string) error { return s.headerMessageIs(m) })
	ctx.Step(`^no se llamó a la API$`, func() error { return s.noAPICalls() })
	ctx.Step(`^(«[^»]*»(?: de [A-Za-zÁÉÍÓÚáéíóúñ ]+?)?) ofrece "([^"]*)"$`, func(f, l string) error { return s.offers(f, l) })
	ctx.Step(`^(«[^»]*») no ofrece proyectos$`, func(f string) error { return s.offersNothing(f) })
	ctx.Step(`^(«[^»]*»(?: de [A-Za-zÁÉÍÓÚáéíóúñ ]+?)?) muestra "([^"]*)"$`, func(f, w string) error { return s.shows(f, w) })
	ctx.Step(`^no hay proyecto actual$`, func() error { return s.noCurrent("«Proyecto actual»") })
	ctx.Step(`^no hay sprint actual$`, func() error { return s.noCurrent("«Sprint actual»") })
	ctx.Step(`^los desplegables ofrecen:$`, func(table *godog.Table) error {
		for _, row := range table.Rows[1:] {
			if err := s.offers(row.Cells[0].Value, row.Cells[1].Value); err != nil {
				return err
			}
		}
		return nil
	})
	ctx.Step(`^la descripción del defecto sigue siendo "([^"]*)"$`, func(d string) error { return s.defectDescriptionIs(d) })
	ctx.Step(`^cada fila de la tabla de "([^"]*)" tiene "([^"]*)"$`, func(sec, l string) error { return s.rowsHave(sec, l) })
	ctx.Step(`^el formulario de modificar integrante tiene el nombre "([^"]*)" y el rol "([^"]*)"$`, func(n, r string) error { return s.memberFormHas(n, r) })
	ctx.Step(`^«Ítem» de Registrar esfuerzo no tiene un ítem elegido$`, func() error { return s.noItemChosen() })
	ctx.Step(`^el menú se ve al costado del contenido$`, func() error { return s.menuBeside() })
	ctx.Step(`^el menú se ve arriba del contenido$`, func() error { return s.menuAbove() })
	ctx.Step(`^la página no se desplaza en horizontal$`, func() error { return s.noHorizontalScroll() })
	ctx.Step(`^la tabla del backlog se desplaza dentro de su recuadro$`, func() error { return s.backlogTableScrolls() })
	ctx.Step(`^todos los pedidos del navegador fueron al mismo servidor$`, func() error { return s.sameOrigin() })
	ctx.Step(`^la API responde (\d+)$`, func(c int) error { return s.apiAnswered(c) })
}
