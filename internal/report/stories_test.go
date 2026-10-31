package report

import (
	"errors"
	"reflect"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
)

// HU-49 · Ver las historias y sus estimaciones en el reporte (Issue #170)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-49-historias-en-reporte.md). Se reutilizan el
// entorno, los repositorios espía y los datos de los tests de HU-38 y HU-39.
// Las historias se leen al final (RN-49.7), así que su lectura es la que
// sigue a las lecturas de ítems que ya hacen las métricas y los defectos;
// los dos helpers de abajo las cuentan haciendo esas mismas consultas.

// projectItemReadsBeforeStories devuelve cuántas lecturas de ítems hacen las
// partes del reporte del proyecto que se leen antes de las historias
// (proyecto → métricas → sprints → esfuerzos → defectos), y deja los
// contadores en cero.
func (e *reportEnv) projectItemReadsBeforeStories(t *testing.T, projectID int) int {
	t.Helper()
	e.resetReads()
	_, _ = e.projects.Get(projectID)
	_, _ = e.metrics.Project(projectID)
	sprints, _ := e.sprints.ListByProject(projectID)
	for _, sp := range sprints {
		_, _ = e.efforts.ListBySprint(sp.ID)
	}
	_, _ = e.defects.ListByProject(projectID)
	n := e.itemSpy.calls
	e.resetReads()
	return n
}

// sprintItemReadsBeforeStories es lo mismo para el reporte del sprint
// (sprint → métricas → esfuerzos → defectos).
func (e *sprintReportEnv) sprintItemReadsBeforeStories(t *testing.T, sprintID int) int {
	t.Helper()
	e.resetSprintReads()
	sp, _ := e.sprints.Get(sprintID)
	_, _ = e.metrics.Sprint(sprintID)
	_, _ = e.efforts.ListBySprint(sprintID)
	_, _ = e.defects.ListByProject(sp.ProjectID)
	n := e.itemSpy.calls
	e.resetSprintReads()
	return n
}

// storedItems devuelve las historias guardadas del proyecto, sin pasar por
// los contadores.
func (e *reportEnv) storedItems(t *testing.T, projectID int) []backlog.Item {
	t.Helper()
	items, err := e.itemSpy.Repository.ListByProject(projectID)
	if err != nil {
		t.Fatalf("leer las historias guardadas: %v", err)
	}
	return items
}

// storiesOfSprint devuelve, de las historias guardadas, las del sprint.
func storiesOfSprint(items []backlog.Item, sprintID int) []backlog.Item {
	selected := []backlog.Item{}
	for _, it := range items {
		if it.SprintID == sprintID {
			selected = append(selected, it)
		}
	}
	return selected
}

func titlesOf(items []backlog.Item) []string {
	titles := []string{}
	for _, it := range items {
		titles = append(titles, it.Title)
	}
	return titles
}

// CA-49.1, RN-49.1, RN-49.3, RN-49.4 y RN-49.6 · El reporte del proyecto
// incluye todas las historias de su backlog, en cualquier estado y con o sin
// sprint, ordenadas por ID, con sus datos y las cantidades; no las de otro
// proyecto.
func TestProject_IncluyeLasHistoriasDelBacklog(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.story(t, projectID, "Registrar venta", backlog.PriorityHigh, 3, 8, backlog.StatusFinished, s.ID)
	e.story(t, projectID, "Emitir factura", backlog.PriorityLow, 5, 6, backlog.StatusInProgress, s.ID)
	e.story(t, projectID, "Consultar stock", backlog.PriorityMedium, 0, 0, backlog.StatusPending, 0)
	otherID := e.project(t, "Sistema de compras", "")
	e.story(t, otherID, "Registrar compra", backlog.PriorityHigh, 8, 12, backlog.StatusPending, 0)

	got := generate(t, e, projectID)

	want := e.storedItems(t, projectID)
	if !reflect.DeepEqual(got.Stories.Stories, want) {
		t.Errorf("Stories = %+v, se esperaba %+v", got.Stories.Stories, want)
	}
	if titles := titlesOf(got.Stories.Stories); !reflect.DeepEqual(titles, []string{"Registrar venta", "Emitir factura", "Consultar stock"}) {
		t.Errorf("títulos = %v, se esperaba el orden por ID", titles)
	}
	if got.Stories.Completed != 1 {
		t.Errorf("Completed = %d, se esperaba 1", got.Stories.Completed)
	}
}

// CA-49.4 y RN-49.3 · Una historia sin estimar, sin horas y sin sprint tiene
// 0 Story Points, 0 horas y SprintID 0.
func TestProject_HistoriaSinEstimarSinHorasNiSprint(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	e.story(t, projectID, "Consultar stock", backlog.PriorityMedium, 0, 0, backlog.StatusPending, 0)

	got := generate(t, e, projectID)

	if len(got.Stories.Stories) != 1 {
		t.Fatalf("Stories = %+v, se esperaba una historia", got.Stories.Stories)
	}
	it := got.Stories.Stories[0]
	if it.StoryPoints != 0 || it.EstimatedHours != 0 || it.SprintID != 0 || it.Status != backlog.StatusPending {
		t.Errorf("historia = %+v, se esperaba sin estimar, sin horas, sin sprint y PENDIENTE", it)
	}
	if got.Stories.Completed != 0 {
		t.Errorf("Completed = %d, se esperaba 0", got.Stories.Completed)
	}
}

// CA-49.3 y RN-49.5 · Las cantidades del reporte del proyecto coinciden con
// los ítems totales y terminados de sus métricas.
func TestProject_CantidadesDeHistoriasCoincidenConLasMetricas(t *testing.T) {
	e := newReportEnv()
	projectID, _ := e.multiSprintProject(t)

	got := generate(t, e, projectID)

	if len(got.Stories.Stories) != 5 || got.Stories.Completed != 2 {
		t.Errorf("cantidad = %d y completadas = %d, se esperaba 5 y 2", len(got.Stories.Stories), got.Stories.Completed)
	}
	if len(got.Stories.Stories) != got.Metrics.Completion.Items || got.Stories.Completed != got.Metrics.Completion.Finished {
		t.Errorf("cantidad = %d y completadas = %d, las métricas dicen %d y %d",
			len(got.Stories.Stories), got.Stories.Completed, got.Metrics.Completion.Items, got.Metrics.Completion.Finished)
	}
}

// CA-49.2, CA-49.3, RN-49.2 y RN-49.5 · El reporte de cada sprint incluye
// solo sus historias, con sus cantidades iguales a las de sus métricas; no
// las de otros sprints ni las sin asignar. Un sprint cerrado muestra sus
// historias sin completar con su estado.
func TestSprint_IncluyeSoloLasHistoriasDelSprint(t *testing.T) {
	e := newSprintReportEnv()
	projectID, sprints := e.multiSprintProject(t)
	stored := e.storedItems(t, projectID)
	wantTitles := [][]string{{"A", "B"}, {"C"}, {"D"}}
	wantCompleted := []int{1, 1, 0}

	for i, sp := range sprints {
		got := generateSprint(t, e, sp.ID)

		if !reflect.DeepEqual(got.Stories.Stories, storiesOfSprint(stored, sp.ID)) {
			t.Errorf("%s: Stories = %+v, se esperaba %+v", sp.Name, got.Stories.Stories, storiesOfSprint(stored, sp.ID))
		}
		if titles := titlesOf(got.Stories.Stories); !reflect.DeepEqual(titles, wantTitles[i]) {
			t.Errorf("%s: títulos = %v, se esperaba %v", sp.Name, titles, wantTitles[i])
		}
		if got.Stories.Completed != wantCompleted[i] {
			t.Errorf("%s: Completed = %d, se esperaba %d", sp.Name, got.Stories.Completed, wantCompleted[i])
		}
		if len(got.Stories.Stories) != got.Metrics.Completion.Items || got.Stories.Completed != got.Metrics.Completion.Finished {
			t.Errorf("%s: cantidad = %d y completadas = %d, las métricas dicen %d y %d", sp.Name,
				len(got.Stories.Stories), got.Stories.Completed, got.Metrics.Completion.Items, got.Metrics.Completion.Finished)
		}
	}
}

// CA-49.4 · Un proyecto sin historias tiene la lista vacía (no nil) y
// cantidades 0.
func TestProject_SinHistorias(t *testing.T) {
	e := newReportEnv()

	got := generate(t, e, e.project(t, "Proyecto vacío", ""))

	if got.Stories.Stories == nil || len(got.Stories.Stories) != 0 || got.Stories.Completed != 0 {
		t.Errorf("Stories = %#v, se esperaba una lista vacía y 0 completadas", got.Stories)
	}
}

// CA-49.4 · Un sprint sin historias asignadas tiene la lista vacía (no nil),
// aunque el proyecto tenga historias sin sprint.
func TestSprint_SinHistorias(t *testing.T) {
	e := newSprintReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.story(t, projectID, "Consultar stock", backlog.PriorityMedium, 0, 0, backlog.StatusPending, 0)

	got := generateSprint(t, e, s.ID)

	if got.Stories.Stories == nil || len(got.Stories.Stories) != 0 || got.Stories.Completed != 0 {
		t.Errorf("Stories = %#v, se esperaba una lista vacía y 0 completadas", got.Stories)
	}
}

// CA-49.5 y RN-49.7 · Las historias se leen una sola vez y al final, después
// de los defectos: si falla esa lectura, se devuelve su error y el valor cero
// del reporte, sin reporte parcial.
func TestProject_FallaLaLecturaDeLasHistorias(t *testing.T) {
	e := newReportEnv()
	projectID, sprints := e.multiSprintProject(t)
	e.itemSpy.failAt = e.projectItemReadsBeforeStories(t, projectID) + 1

	got, err := e.reports.Project(projectID)

	if !errors.Is(err, errItemsRead) {
		t.Errorf("err = %v, se esperaba %v", err, errItemsRead)
	}
	if !reflect.DeepEqual(got, ProjectReport{}) {
		t.Errorf("Project = %+v, se esperaba el valor cero", got)
	}
	if e.effortSpy.calls != len(sprints) || e.defectSpy.calls != 1 {
		t.Errorf("lecturas de esfuerzos y defectos = %d y %d; se esperaba %d y 1, antes de las historias",
			e.effortSpy.calls, e.defectSpy.calls, len(sprints))
	}
}

// CA-49.5 y RN-49.7 · Lo mismo en el reporte del sprint.
func TestSprint_FallaLaLecturaDeLasHistorias(t *testing.T) {
	e := newSprintReportEnv()
	s := e.tec03Sprint(t)
	e.itemSpy.failAt = e.sprintItemReadsBeforeStories(t, s.ID) + 1

	got, err := e.reports.Sprint(s.ID)

	if !errors.Is(err, errItemsRead) {
		t.Errorf("err = %v, se esperaba %v", err, errItemsRead)
	}
	if !reflect.DeepEqual(got, SprintReport{}) {
		t.Errorf("Sprint = %+v, se esperaba el valor cero", got)
	}
	if e.effortSpy.calls != 1 || e.defectSpy.calls != 1 {
		t.Errorf("lecturas de esfuerzos y defectos = %d y %d; se esperaba 1 y 1, antes de las historias",
			e.effortSpy.calls, e.defectSpy.calls)
	}
}

// CA-49.5 y RN-49.7 · Un reporte sin fallas lee las historias una sola vez,
// después de las demás partes.
func TestReports_LeenLasHistoriasUnaSolaVez(t *testing.T) {
	e := newSprintReportEnv()
	projectID, sprints := e.multiSprintProject(t)

	before := e.projectItemReadsBeforeStories(t, projectID)
	generate(t, e.reportEnv, projectID)
	if e.itemSpy.calls != before+1 {
		t.Errorf("proyecto: lecturas de ítems = %d, se esperaba %d", e.itemSpy.calls, before+1)
	}

	before = e.sprintItemReadsBeforeStories(t, sprints[0].ID)
	generateSprint(t, e, sprints[0].ID)
	if e.itemSpy.calls != before+1 {
		t.Errorf("sprint: lecturas de ítems = %d, se esperaba %d", e.itemSpy.calls, before+1)
	}
}

// CA-49.5, RN-49.8 y RN-49.9 · Modificar las historias del reporte (o sus
// criterios) no cambia lo guardado ni el reporte siguiente, en el proyecto y
// en el sprint.
func TestReports_HistoriasSonCopias(t *testing.T) {
	e := newSprintReportEnv()
	projectID, sprints := e.multiSprintProject(t)
	stored := e.storedItems(t, projectID)

	projectReport := generate(t, e.reportEnv, projectID)
	sprintReport := generateSprint(t, e, sprints[0].ID)
	if len(projectReport.Stories.Stories) == 0 || len(sprintReport.Stories.Stories) == 0 {
		t.Fatalf("Stories del proyecto = %+v y del sprint = %+v, se esperaban historias",
			projectReport.Stories.Stories, sprintReport.Stories.Stories)
	}
	for _, stories := range [][]backlog.Item{projectReport.Stories.Stories, sprintReport.Stories.Stories} {
		stories[0].Title = "modificado"
		stories[0].AcceptanceCriteria[0] = "modificado"
	}

	if again := generate(t, e.reportEnv, projectID); !reflect.DeepEqual(again.Stories.Stories, stored) {
		t.Errorf("proyecto: el reporte siguiente = %+v, se esperaba %+v", again.Stories.Stories, stored)
	}
	if again := generateSprint(t, e, sprints[0].ID); !reflect.DeepEqual(again.Stories.Stories, storiesOfSprint(stored, sprints[0].ID)) {
		t.Errorf("sprint: el reporte siguiente = %+v, se esperaba %+v", again.Stories.Stories, storiesOfSprint(stored, sprints[0].ID))
	}
	if after := e.storedItems(t, projectID); !reflect.DeepEqual(after, stored) {
		t.Errorf("historias guardadas = %+v, antes eran %+v", after, stored)
	}
}
