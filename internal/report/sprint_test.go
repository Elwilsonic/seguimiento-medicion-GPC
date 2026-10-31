package report

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/metrics"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-39 · Generar el reporte de un sprint (Issue #57)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-39-reporte-sprint.md). Se reutilizan el entorno,
// los repositorios espía y los datos de los tests de HU-38 (project_test.go).
// Solo se agrega un espía de la lectura de un sprint por ID, que es la
// primera lectura del reporte del sprint (RN-39.3).

var errSprintRead = errors.New("falla la lectura del sprint")

// spySprintGetRepository cuenta las lecturas de un sprint por ID, además de
// las de la lista de sprints que ya cuenta spySprintRepository.
type spySprintGetRepository struct {
	*spySprintRepository
	gets, getFailAt int
}

func (r *spySprintGetRepository) GetByID(id int) (sprint.Sprint, error) {
	r.gets++
	if failing(r.gets, r.getFailAt) {
		return sprint.Sprint{}, errSprintRead
	}
	return r.spySprintRepository.GetByID(id)
}

// sprintReportEnv es el entorno de HU-38 con el espía de la lectura de un
// sprint por ID.
type sprintReportEnv struct {
	*reportEnv
	sprintGetSpy *spySprintGetRepository
}

func newSprintReportEnv() *sprintReportEnv {
	e := &reportEnv{
		projectSpy: &spyProjectRepository{Repository: project.NewMemoryRepository()},
		itemSpy:    &spyItemRepository{Repository: backlog.NewMemoryRepository()},
		sprintSpy:  &spySprintRepository{Repository: sprint.NewMemoryRepository()},
		effortSpy:  &spyEffortRepository{Repository: effort.NewMemoryRepository()},
		defectSpy:  &spyDefectRepository{Repository: defect.NewMemoryRepository()},
	}
	get := &spySprintGetRepository{spySprintRepository: e.sprintSpy}
	e.projects = project.NewManager(e.projectSpy)
	e.items = backlog.NewManager(e.itemSpy, e.projects)
	e.sprints = sprint.NewManager(get, e.projects, e.items)
	e.efforts = effort.NewManager(e.effortSpy, e.sprints, e.items)
	e.defects = defect.NewManager(e.defectSpy, e.items, e.sprints)
	e.metrics = metrics.NewManager(e.items, e.sprints, e.efforts, e.defects)
	e.reports = NewManager(e.projects, e.items, e.sprints, e.metrics, e.efforts, e.defects)
	return &sprintReportEnv{reportEnv: e, sprintGetSpy: get}
}

// resetSprintReads pone en cero los contadores, incluido el de la lectura de
// un sprint por ID.
func (e *sprintReportEnv) resetSprintReads() {
	e.resetReads()
	e.sprintGetSpy.gets = 0
}

// tec03Sprint arma el sprint del ejemplo 1 de la SDD: el ejemplo de TEC-03,
// del 2026-10-05 al 2026-10-16, con 12 h de Ana en A, 5 h de Luis en B, 3 h de
// Ana en C, un defecto de A detectado y resuelto en el sprint, y cerrado.
func (e *sprintReportEnv) tec03Sprint(t *testing.T) sprint.Sprint {
	t.Helper()
	projectID := e.project(t, "Software Metrics", "Estimación y seguimiento de proyectos")
	s, err := e.sprints.Create(projectID, "Sprint 1", time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), "Registrar el esfuerzo del equipo")
	if err != nil {
		t.Fatalf("crear el sprint: %v", err)
	}
	a := e.item(t, projectID, "A", 5, 10, backlog.StatusFinished, s.ID)
	b := e.item(t, projectID, "B", 3, 6, backlog.StatusFinished, s.ID)
	c := e.item(t, projectID, "C", 2, 0, backlog.StatusInProgress, s.ID)
	e.item(t, projectID, "D", 0, 4, backlog.StatusPending, s.ID)
	e.effort(t, s.ID, a.ID, "Ana", 12)
	e.effort(t, s.ID, b.ID, "Luis", 5)
	e.effort(t, s.ID, c.ID, "Ana", 3)
	e.saveDefect(t, a.ID, s.ID, defect.StatusResolved, s.ID)
	e.closeSprint(t, s.ID)
	closed, err := e.sprints.Get(s.ID)
	if err != nil {
		t.Fatalf("leer el sprint: %v", err)
	}
	return closed
}

func generateSprint(t *testing.T, e *sprintReportEnv, sprintID int) SprintReport {
	t.Helper()
	got, err := e.reports.Sprint(sprintID)
	if err != nil {
		t.Fatalf("Sprint devolvió un error inesperado: %v", err)
	}
	return got
}

// sprintDefectCounts cuenta, en la lista del reporte, los defectos detectados
// en el sprint y los RESUELTO con ese sprint de resolución (RN-39.5).
func sprintDefectCounts(defects []defect.Defect, sprintID int) metrics.Defects {
	var d metrics.Defects
	for _, df := range defects {
		if df.DetectionSprintID == sprintID {
			d.Detected++
		}
		if df.Status == defect.StatusResolved && df.ResolutionSprintID == sprintID {
			d.Resolved++
		}
	}
	return d
}

// CA-39.1 · El reporte del sprint del ejemplo de TEC-03 tiene los datos del
// sprint, M1, M2 y M4 a M10, los tres esfuerzos con su total, el defecto y
// los dos integrantes.
func TestSprint_EjemploDeTEC03(t *testing.T) {
	e := newSprintReportEnv()
	s := e.tec03Sprint(t)

	got := generateSprint(t, e, s.ID)

	wantSprint := sprint.Sprint{ID: s.ID, ProjectID: s.ProjectID, Name: "Sprint 1", StartDate: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		EndDate: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), Goal: "Registrar el esfuerzo del equipo", Status: sprint.StatusClosed}
	if got.Sprint != wantSprint {
		t.Errorf("Sprint = %+v, se esperaba %+v", got.Sprint, wantSprint)
	}
	wantMetrics := metrics.SprintMetrics{
		SprintID:   s.ID,
		Points:     metrics.SprintPoints{SprintID: s.ID, Planned: 10, Completed: 8, Unestimated: 1},
		Completion: metrics.Completion{Items: 4, Finished: 2, Percent: 50, PercentAvailable: true},
		Hours:      metrics.Hours{Estimated: 20, Actual: 20, Unestimated: 1, DeviationAvailable: true},
		Defects:    metrics.Defects{Detected: 1, Resolved: 1},
	}
	if got.Metrics != wantMetrics {
		t.Errorf("Metrics = %+v, se esperaba %+v", got.Metrics, wantMetrics)
	}
	wantEfforts := []effortKey{{1, s.ID, 1, "Ana", 12}, {2, s.ID, 2, "Luis", 5}, {3, s.ID, 3, "Ana", 3}}
	if keys := keysOf(got.Efforts.Efforts); !reflect.DeepEqual(keys, wantEfforts) || got.Efforts.TotalHours != 20 {
		t.Errorf("Efforts = %+v con total %v, se esperaba %+v con total 20", keys, got.Efforts.TotalHours, wantEfforts)
	}
	if len(got.Defects.Defects) != 1 || got.Defects.Defects[0].Status != defect.StatusResolved || got.Defects.Defects[0].ItemID != 1 {
		t.Errorf("Defects = %+v, se esperaba el defecto resuelto del ítem 1", got.Defects.Defects)
	}
	if want := []string{"Ana", "Luis"}; !reflect.DeepEqual(got.Members.Names, want) {
		t.Errorf("Members = %q, se esperaba %q", got.Members.Names, want)
	}
}

// CA-39.2 y RN-39.7 · El reporte se genera para un sprint abierto y, con los
// mismos datos, es el mismo después de cerrarlo, salvo el estado del sprint.
func TestSprint_SprintAbiertoYCerrado(t *testing.T) {
	e := newSprintReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.item(t, projectID, "A", 5, 10, backlog.StatusFinished, s.ID)
	e.effort(t, s.ID, a.ID, "Ana", 8)
	e.saveDefect(t, a.ID, s.ID, defect.StatusOpen, 0)

	open := generateSprint(t, e, s.ID)
	e.closeSprint(t, s.ID)
	closed := generateSprint(t, e, s.ID)

	if open.Sprint.Status != sprint.StatusOpen || closed.Sprint.Status != sprint.StatusClosed {
		t.Errorf("estados = %s y %s, se esperaba %s y %s", open.Sprint.Status, closed.Sprint.Status, sprint.StatusOpen, sprint.StatusClosed)
	}
	if open.Metrics.Points.Planned != 5 || len(open.Efforts.Efforts) != 1 || len(open.Defects.Defects) != 1 {
		t.Errorf("reporte del sprint abierto = %+v, se esperaban 5 SP, un esfuerzo y un defecto", open)
	}
	closed.Sprint.Status = open.Sprint.Status
	if !reflect.DeepEqual(closed, open) {
		t.Errorf("reporte del sprint cerrado = %+v, se esperaba el mismo que abierto %+v", closed, open)
	}
}

// CA-39.5 y RN-39.9 · Con los mismos datos, el sprint, las métricas y los
// esfuerzos son iguales a sus operaciones de origen, el total de horas es M5
// y los defectos detectados en el sprint son los de ListBySprint (HU-31).
func TestSprint_CadaParteEsIgualASuOperacionDeOrigen(t *testing.T) {
	e := newSprintReportEnv()
	_, sprints := e.multiSprintProject(t)
	for _, s := range sprints {
		wantSprint, _ := e.sprints.Get(s.ID)
		wantMetrics, _ := e.metrics.Sprint(s.ID)
		wantEfforts, _ := e.efforts.ListBySprint(s.ID)
		wantDetected, _ := e.defects.ListBySprint(s.ID)

		got := generateSprint(t, e, s.ID)

		if got.Sprint != wantSprint {
			t.Errorf("sprint %d: Sprint = %+v, se esperaba %+v", s.ID, got.Sprint, wantSprint)
		}
		if got.Metrics != wantMetrics {
			t.Errorf("sprint %d: Metrics = %+v, se esperaba %+v", s.ID, got.Metrics, wantMetrics)
		}
		if !reflect.DeepEqual(got.Efforts.Efforts, wantEfforts.Efforts) {
			t.Errorf("sprint %d: Efforts = %+v, se esperaba %+v", s.ID, keysOf(got.Efforts.Efforts), keysOf(wantEfforts.Efforts))
		}
		if got.Efforts.TotalHours != wantMetrics.Hours.Actual {
			t.Errorf("sprint %d: TotalHours = %v, se esperaba %v (M5)", s.ID, got.Efforts.TotalHours, wantMetrics.Hours.Actual)
		}
		var detected []defect.Defect
		for _, d := range got.Defects.Defects {
			if d.DetectionSprintID == s.ID {
				detected = append(detected, d)
			}
		}
		if len(detected) != len(wantDetected) || (len(detected) > 0 && !reflect.DeepEqual(detected, wantDetected)) {
			t.Errorf("sprint %d: detectados = %v, se esperaba %v", s.ID, defectIDs(detected), defectIDs(wantDetected))
		}
	}
}

// CA-39.5 y RN-39.5 · Los defectos del sprint son los detectados en él o
// resueltos en él, ordenados por ID, y sus conteos son M9 y M10. En el
// proyecto de varios sprints, el defecto 2 se detecta en el sprint 1 y se
// resuelve en el 2, el 1 queda ABIERTO en el 1, el 3 EN_CORRECCION en el 2 y
// el 4 es de otro proyecto.
func TestSprint_DefectosDetectadosOResueltosEnElSprint(t *testing.T) {
	e := newSprintReportEnv()
	_, sprints := e.multiSprintProject(t)
	want := map[int][]int{sprints[0].ID: {1, 2}, sprints[1].ID: {2, 3}, sprints[2].ID: {}}
	for _, s := range sprints {
		got := generateSprint(t, e, s.ID)

		if ids := defectIDs(got.Defects.Defects); !reflect.DeepEqual(ids, want[s.ID]) {
			t.Errorf("sprint %d: Defects = %v, se esperaba %v", s.ID, ids, want[s.ID])
		}
		if counts := sprintDefectCounts(got.Defects.Defects, s.ID); counts != got.Metrics.Defects {
			t.Errorf("sprint %d: conteos de la lista = %+v, se esperaba M9 y M10 = %+v", s.ID, counts, got.Metrics.Defects)
		}
	}
}

// RN-39.5 · Un defecto detectado y resuelto en el mismo sprint aparece una
// sola vez y cuenta en M9 y en M10.
func TestSprint_DefectoDetectadoYResueltoEnElSprintApareceUnaVez(t *testing.T) {
	e := newSprintReportEnv()
	s := e.tec03Sprint(t)

	got := generateSprint(t, e, s.ID)

	if ids := defectIDs(got.Defects.Defects); !reflect.DeepEqual(ids, []int{1}) {
		t.Errorf("Defects = %v, se esperaba [1]", ids)
	}
	if want := (metrics.Defects{Detected: 1, Resolved: 1}); got.Metrics.Defects != want || sprintDefectCounts(got.Defects.Defects, s.ID) != want {
		t.Errorf("M9 y M10 = %+v, se esperaba %+v en las métricas y en la lista", got.Metrics.Defects, want)
	}
}

// CA-39.5 y RN-39.4 · Con horas con decimales, el total es la suma exacta,
// sin redondear, y es el mismo float64 que M5.
func TestSprint_TotalDeHorasExactoEIgualAHorasReales(t *testing.T) {
	e := newSprintReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.item(t, projectID, "A", 0, 1, backlog.StatusPending, s.ID)
	e.effort(t, s.ID, a.ID, "Ana", 0.1)
	e.effort(t, s.ID, a.ID, "Ana", 0.2)
	// Suma en float64 en el orden de los esfuerzos (una suma de constantes de
	// Go sería exacta y daría 0.3).
	h1, h2 := 0.1, 0.2
	wantTotal := 0.0 + h1 + h2

	got := generateSprint(t, e, s.ID)

	if got.Efforts.TotalHours != wantTotal {
		t.Errorf("TotalHours = %v, se esperaba la suma exacta %v", got.Efforts.TotalHours, wantTotal)
	}
	if got.Efforts.TotalHours != got.Metrics.Hours.Actual {
		t.Errorf("TotalHours = %v, se esperaba %v (M5)", got.Efforts.TotalHours, got.Metrics.Hours.Actual)
	}
}

// CA-39.6 y RN-39.6 · Solo aparecen los esfuerzos y los integrantes del
// sprint, cada integrante una sola vez, en el orden de su primer esfuerzo;
// no aparecen los de otros sprints ni los de otro proyecto.
func TestSprint_IntegrantesYEsfuerzosSoloDelSprint(t *testing.T) {
	e := newSprintReportEnv()
	_, sprints := e.multiSprintProject(t)
	cases := []struct {
		sprintID int
		members  []string
		efforts  []effortKey
	}{
		{sprints[0].ID, []string{"Ana", "Luis"}, []effortKey{{1, sprints[0].ID, 1, "Ana", 5}, {2, sprints[0].ID, 2, "Luis", 1.5}}},
		{sprints[1].ID, []string{"Luis", "Ana"}, []effortKey{{3, sprints[1].ID, 3, "Luis", 4}, {4, sprints[1].ID, 3, "Ana", 0.5}}},
		{sprints[2].ID, []string{"ana"}, []effortKey{{5, sprints[2].ID, 4, "ana", 2}}},
	}
	for _, c := range cases {
		got := generateSprint(t, e, c.sprintID)

		if !reflect.DeepEqual(got.Members.Names, c.members) {
			t.Errorf("sprint %d: Members = %q, se esperaba %q", c.sprintID, got.Members.Names, c.members)
		}
		if keys := keysOf(got.Efforts.Efforts); !reflect.DeepEqual(keys, c.efforts) {
			t.Errorf("sprint %d: Efforts = %+v, se esperaba %+v", c.sprintID, keys, c.efforts)
		}
	}
}

// expectEmptySprintParts verifica que las listas del reporte estén vacías
// pero no sean nil, con total 0 (RN-39.8).
func expectEmptySprintParts(t *testing.T, got SprintReport) {
	t.Helper()
	if got.Efforts.Efforts == nil || len(got.Efforts.Efforts) != 0 || got.Efforts.TotalHours != 0 {
		t.Errorf("Efforts = %#v con total %v, se esperaba una lista vacía no nil y total 0", got.Efforts.Efforts, got.Efforts.TotalHours)
	}
	if got.Defects.Defects == nil || len(got.Defects.Defects) != 0 {
		t.Errorf("Defects = %#v, se esperaba una lista vacía no nil", got.Defects.Defects)
	}
	if got.Members.Names == nil || len(got.Members.Names) != 0 {
		t.Errorf("Members = %#v, se esperaba una lista vacía no nil", got.Members.Names)
	}
}

// CA-39.4 y RN-39.8 · Un sprint recién creado genera el reporte con sus
// datos, las listas vacías y las métricas no disponibles.
func TestSprint_SprintSinDatos(t *testing.T) {
	e := newSprintReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	s := e.openSprint(t, projectID, "Sprint vacío")

	got := generateSprint(t, e, s.ID)

	if got.Sprint != s {
		t.Errorf("Sprint = %+v, se esperaba %+v", got.Sprint, s)
	}
	if got.Metrics.SprintID != s.ID || got.Metrics.Completion.PercentAvailable || got.Metrics.Hours.DeviationAvailable {
		t.Errorf("Metrics = %+v, se esperaban las métricas del sprint %d con las no disponibles en false", got.Metrics, s.ID)
	}
	expectEmptySprintParts(t, got)
}

// CA-39.4 · Un sprint con ítems y sin esfuerzos tiene las métricas de sus
// ítems y las listas vacías.
func TestSprint_SprintConItemsYSinEsfuerzos(t *testing.T) {
	e := newSprintReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.item(t, projectID, "A", 3, 4, backlog.StatusPending, s.ID)

	got := generateSprint(t, e, s.ID)

	if got.Metrics.Points.Planned != 3 || got.Metrics.Hours.Estimated != 4 || got.Metrics.Hours.DeviationAvailable {
		t.Errorf("Metrics = %+v, se esperaban 3 SP, 4 h estimadas y la desviación no disponible", got.Metrics)
	}
	expectEmptySprintParts(t, got)
}

// CA-39.3 y RN-39.3 · Un sprint inexistente (incluidos 0 y negativos)
// devuelve ErrSprintNotFound sin leer nada más que el sprint.
func TestSprint_SprintInexistenteNoLeeNadaMas(t *testing.T) {
	for _, id := range []int{99, 0, -1} {
		e := newSprintReportEnv()
		e.multiSprintProject(t)
		e.resetSprintReads()

		got, err := e.reports.Sprint(id)

		if !errors.Is(err, sprint.ErrSprintNotFound) {
			t.Errorf("id %d: err = %v, se esperaba %v", id, err, sprint.ErrSprintNotFound)
		}
		if !reflect.DeepEqual(got, SprintReport{}) {
			t.Errorf("id %d: Sprint = %+v, se esperaba el valor cero", id, got)
		}
		if e.sprintGetSpy.gets != 1 || e.itemSpy.calls != 0 || e.effortSpy.calls != 0 || e.defectSpy.calls != 0 {
			t.Errorf("id %d: lecturas del sprint, ítems, esfuerzos y defectos = %d, %d, %d y %d; se esperaba 1, 0, 0 y 0",
				id, e.sprintGetSpy.gets, e.itemSpy.calls, e.effortSpy.calls, e.defectSpy.calls)
		}
	}
}

// CA-39.3 y RN-39.3 · Se lee sprint → métricas → esfuerzos → defectos: se
// devuelve el error de la lectura que falla, sin reporte parcial y sin las
// lecturas siguientes. Las lecturas de esfuerzos y defectos del reporte son
// las que siguen a las de metrics.Manager.Sprint.
func TestSprint_ErroresEnElOrdenDeLectura(t *testing.T) {
	cases := []struct {
		name             string
		fail             func(e *sprintReportEnv)
		want             error
		efforts, defects int
	}{
		{name: "sprint", fail: func(e *sprintReportEnv) { e.sprintGetSpy.getFailAt = 1 }, want: errSprintRead, efforts: 0, defects: 0},
		{name: "métricas", fail: func(e *sprintReportEnv) { e.itemSpy.failAt = 1 }, want: errItemsRead, efforts: 0, defects: 0},
		{name: "esfuerzos", fail: func(e *sprintReportEnv) { e.effortSpy.failAt = 2 }, want: errEffortsRead, efforts: 2, defects: 1},
		{name: "defectos", fail: func(e *sprintReportEnv) { e.defectSpy.failAt = 2 }, want: errDefectsRead, efforts: 2, defects: 2},
	}
	for _, c := range cases {
		e := newSprintReportEnv()
		_, sprints := e.multiSprintProject(t)
		e.resetSprintReads()
		c.fail(e)

		got, err := e.reports.Sprint(sprints[1].ID)

		if !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, se esperaba %v", c.name, err, c.want)
		}
		if !reflect.DeepEqual(got, SprintReport{}) {
			t.Errorf("%s: Sprint = %+v, se esperaba el valor cero", c.name, got)
		}
		if e.effortSpy.calls != c.efforts || e.defectSpy.calls != c.defects {
			t.Errorf("%s: lecturas de esfuerzos y defectos = %d y %d; se esperaba %d y %d",
				c.name, e.effortSpy.calls, e.defectSpy.calls, c.efforts, c.defects)
		}
	}
}

// CA-39.7, RN-39.11 y RN-39.12 · Generar el reporte dos veces sin cambios da
// lo mismo, modificar el reporte devuelto no cambia el siguiente y no se
// modifica nada de lo guardado.
func TestSprint_NoModificaNadaYDevuelveCopias(t *testing.T) {
	e := newSprintReportEnv()
	projectID, sprints := e.multiSprintProject(t)
	s := sprints[1]
	itemsBefore, _ := e.itemSpy.Repository.ListByProject(projectID)
	sprintsBefore, _ := e.sprintSpy.Repository.ListByProject(projectID)
	effortsBefore, _ := e.effortSpy.Repository.ListBySprint(s.ID)
	defectsBefore, _ := e.defectSpy.Repository.List()

	first := generateSprint(t, e, s.ID)
	if len(first.Efforts.Efforts) == 0 || len(first.Defects.Defects) == 0 || len(first.Members.Names) == 0 {
		t.Fatalf("el reporte = %+v, se esperaban esfuerzos, defectos e integrantes", first)
	}
	changed := generateSprint(t, e, s.ID)
	if !reflect.DeepEqual(changed, first) {
		t.Errorf("la segunda consulta sin cambios = %+v, se esperaba %+v", changed, first)
	}
	changed.Efforts.Efforts[0].Hours = 999
	changed.Defects.Defects[0].Status = "OTRO"
	changed.Members.Names[0] = "Otro"

	again := generateSprint(t, e, s.ID)

	if !reflect.DeepEqual(again, first) {
		t.Errorf("modificar el reporte cambió el siguiente: %+v, se esperaba %+v", again, first)
	}
	itemsAfter, _ := e.itemSpy.Repository.ListByProject(projectID)
	sprintsAfter, _ := e.sprintSpy.Repository.ListByProject(projectID)
	effortsAfter, _ := e.effortSpy.Repository.ListBySprint(s.ID)
	defectsAfter, _ := e.defectSpy.Repository.List()
	if !reflect.DeepEqual(itemsAfter, itemsBefore) || !reflect.DeepEqual(sprintsAfter, sprintsBefore) ||
		!reflect.DeepEqual(effortsAfter, effortsBefore) || !reflect.DeepEqual(defectsAfter, defectsBefore) {
		t.Error("el reporte modificó los ítems, los sprints, los esfuerzos o los defectos")
	}
}
