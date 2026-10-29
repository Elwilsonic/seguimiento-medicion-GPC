package metrics

import (
	"errors"
	"reflect"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-37 · Ver las métricas en el dashboard con gráficos (Issue #55)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-37-graficos-dashboard.md). Usa el entorno de
// HU-35 (summaryEnv), con los repositorios espía que cuentan las lecturas y
// que pueden fallar a pedido.

func expectProjectSprints(t *testing.T, got []SprintWithMetrics, err error, want []SprintWithMetrics) {
	t.Helper()
	if err != nil {
		t.Fatalf("ProjectSprints devolvió un error inesperado: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ProjectSprints =\n%+v\nse esperaba\n%+v", got, want)
	}
}

// CA-37.1 y CA-37.4 · El ejemplo de TEC-03 en un sprint cerrado y un sprint
// abierto después: cada sprint con su copia y sus métricas, en orden de ID.
func TestProjectSprints_EjemploDeTEC03ConUnSprintCerradoYOtroAbierto(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	s1 := e.tec03Full(t, projectID)
	if _, err := e.sprints.Close(s1.ID); err != nil {
		t.Fatalf("cerrar el sprint 1: %v", err)
	}
	s2 := e.openSprint(t, projectID, "Sprint 2")
	e.fullItem(t, projectID, "E", 3, 4, backlog.StatusFinished, s2.ID, 10)
	stored := e.sprintsOf(projectID)

	got, err := e.metrics.ProjectSprints(projectID)

	expectProjectSprints(t, got, err, []SprintWithMetrics{
		{
			Sprint: stored[0],
			Metrics: SprintMetrics{
				SprintID:   s1.ID,
				Points:     SprintPoints{SprintID: s1.ID, Planned: 10, Completed: 8, Unestimated: 1},
				Completion: Completion{Items: 4, Finished: 2, Percent: 50, PercentAvailable: true},
				Hours:      Hours{Estimated: 20, Actual: 20, Unestimated: 1, DeviationAvailable: true},
				Defects:    Defects{Detected: 1, Resolved: 1},
			},
		},
		{
			Sprint: stored[1],
			Metrics: SprintMetrics{
				SprintID:   s2.ID,
				Points:     SprintPoints{SprintID: s2.ID, Planned: 3, Completed: 3},
				Completion: Completion{Items: 1, Finished: 1, Percent: 100, PercentAvailable: true},
				Hours:      Hours{Estimated: 4, Actual: 10, Deviation: 6, DeviationPercent: 150, DeviationAvailable: true},
			},
		},
	})
	if len(stored) != 2 || stored[0].Status != sprint.StatusClosed || stored[1].Status != sprint.StatusOpen {
		t.Errorf("sprints guardados = %+v, se esperaba el 1 CERRADO y el 2 ABIERTO", stored)
	}
}

// CA-37.3 y RN-37.1 · Las métricas de cada sprint de la lista son iguales a
// las de Sprint(id) de HU-35, y cada sprint es igual al guardado.
func TestProjectSprints_EsIgualASprintParaCadaSprint(t *testing.T) {
	e := newSummaryEnv()
	projectID, sprints := e.multiSprintProject(t)
	stored := e.sprintsOf(projectID)

	got, err := e.metrics.ProjectSprints(projectID)
	if err != nil {
		t.Fatalf("ProjectSprints devolvió un error inesperado: %v", err)
	}

	if len(got) != len(sprints) {
		t.Fatalf("ProjectSprints devolvió %d sprints, se esperaba %d", len(got), len(sprints))
	}
	for i, entry := range got {
		want, err := e.metrics.Sprint(sprints[i].ID)
		if err != nil {
			t.Fatalf("Sprint(%d): %v", sprints[i].ID, err)
		}
		if entry.Metrics != want {
			t.Errorf("sprint %d: métricas = %+v, Sprint da %+v", sprints[i].ID, entry.Metrics, want)
		}
		if entry.Sprint != stored[i] {
			t.Errorf("sprint %d: sprint = %+v, el guardado es %+v", sprints[i].ID, entry.Sprint, stored[i])
		}
	}
}

// CA-37.4 y RN-37.4 · La lista tiene todos los sprints, abiertos y cerrados,
// ordenados por ID ascendente.
func TestProjectSprints_OrdenadosPorIDConAbiertosYCerrados(t *testing.T) {
	e := newSummaryEnv()
	projectID, sprints := e.multiSprintProject(t)

	got, err := e.metrics.ProjectSprints(projectID)
	if err != nil {
		t.Fatalf("ProjectSprints devolvió un error inesperado: %v", err)
	}

	wantStatus := []sprint.Status{sprint.StatusClosed, sprint.StatusClosed, sprint.StatusOpen}
	if len(got) != len(sprints) {
		t.Fatalf("ProjectSprints devolvió %d sprints, se esperaba %d", len(got), len(sprints))
	}
	for i, entry := range got {
		if entry.Sprint.ID != sprints[i].ID || entry.Metrics.SprintID != sprints[i].ID {
			t.Errorf("posición %d: sprint %d con métricas del %d, se esperaba el %d", i, entry.Sprint.ID, entry.Metrics.SprintID, sprints[i].ID)
		}
		if entry.Sprint.Status != wantStatus[i] {
			t.Errorf("posición %d: estado = %s, se esperaba %s", i, entry.Sprint.Status, wantStatus[i])
		}
	}
}

// CA-37.3 · Un ítem sin sprint no cuenta en ningún sprint de la lista: en
// multiSprintProject, E y F no tienen sprint.
func TestProjectSprints_ItemSinSprintNoCuentaEnNingunSprint(t *testing.T) {
	e := newSummaryEnv()
	projectID, _ := e.multiSprintProject(t)

	got, err := e.metrics.ProjectSprints(projectID)
	if err != nil {
		t.Fatalf("ProjectSprints devolvió un error inesperado: %v", err)
	}

	wantPlanned := []int{5, 5, 8}
	wantItems := []int{2, 1, 1}
	if len(got) != len(wantPlanned) {
		t.Fatalf("ProjectSprints devolvió %d sprints, se esperaba %d", len(got), len(wantPlanned))
	}
	for i, entry := range got {
		if entry.Metrics.Points.Planned != wantPlanned[i] || entry.Metrics.Completion.Items != wantItems[i] {
			t.Errorf("posición %d: %d SP planificados en %d ítems, se esperaba %d en %d", i, entry.Metrics.Points.Planned, entry.Metrics.Completion.Items, wantPlanned[i], wantItems[i])
		}
	}
}

// CA-37.2 y CA-37.4 · Un proyecto sin sprints devuelve una lista vacía, no
// nil, aunque tenga ítems.
func TestProjectSprints_SinSprintsDevuelveListaVaciaNoNil(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	e.fullItem(t, projectID, "A", 3, 4, backlog.StatusPending, 0, 0)

	got, err := e.metrics.ProjectSprints(projectID)

	if err != nil {
		t.Fatalf("ProjectSprints devolvió un error inesperado: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ProjectSprints = %#v, se esperaba una lista vacía, no nil", got)
	}
}

// CA-37.2 · Un sprint sin ítems tiene todo en 0, y el porcentaje y la
// desviación no disponibles.
func TestProjectSprints_SprintSinItemsTodoEnCeroYNoDisponibles(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")

	got, err := e.metrics.ProjectSprints(projectID)

	expectProjectSprints(t, got, err, []SprintWithMetrics{
		{Sprint: e.sprintsOf(projectID)[0], Metrics: SprintMetrics{SprintID: s.ID, Points: SprintPoints{SprintID: s.ID}}},
	})
}

// CA-37.2 · Un sprint con horas estimadas y sin esfuerzos tiene la
// desviación no disponible (M5 = 0).
func TestProjectSprints_SinEsfuerzosLaDesviacionNoDisponible(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.fullItem(t, projectID, "A", 5, 10, backlog.StatusPending, s.ID, 0)

	got, err := e.metrics.ProjectSprints(projectID)

	expectProjectSprints(t, got, err, []SprintWithMetrics{
		{
			Sprint: e.sprintsOf(projectID)[0],
			Metrics: SprintMetrics{
				SprintID:   s.ID,
				Points:     SprintPoints{SprintID: s.ID, Planned: 5},
				Completion: Completion{Items: 1, PercentAvailable: true},
				Hours:      Hours{Estimated: 10},
			},
		},
	})
}

// CA-37.4 · Un proyecto inexistente, 0 o negativo devuelve
// ErrProjectNotFound y ninguna lista.
func TestProjectSprints_InexistenteDevuelveErrProjectNotFound(t *testing.T) {
	e := newSummaryEnv()
	for _, id := range []int{99, 0, -1} {
		got, err := e.metrics.ProjectSprints(id)

		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("ProjectSprints(%d): err = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if got != nil {
			t.Errorf("ProjectSprints(%d) = %+v, se esperaba nil", id, got)
		}
	}
}

// CA-37.4 y RN-37.2 · Se lee ítems → sprints → esfuerzos → defectos: se
// devuelve el error de la primera lectura que falla, sin lista parcial, y no
// se hacen las siguientes.
func TestProjectSprints_ErroresEnElOrdenDeLectura(t *testing.T) {
	cases := []readOrderCase{
		{name: "ítems y sprints", fail: func(e *summaryEnv) { e.itemSpy.err, e.sprintSpy.listErr = errItemsRead, errSprintsRead }, want: errItemsRead},
		{name: "sprints y esfuerzos", fail: func(e *summaryEnv) { e.sprintSpy.listErr, e.effortSpy.err = errSprintsRead, errEffortsRead }, want: errSprintsRead},
		{name: "esfuerzos y defectos", fail: func(e *summaryEnv) { e.effortSpy.err, e.defectSpy.err = errEffortsRead, errDefectsRead }, want: errEffortsRead, effortsRead: true},
		{name: "defectos", fail: func(e *summaryEnv) { e.defectSpy.err = errDefectsRead }, want: errDefectsRead, effortsRead: true, defectsRead: true},
	}
	for _, c := range cases {
		e := newSummaryEnv()
		projectID, _ := e.multiSprintProject(t)
		c.fail(e)
		e.resetReads()

		got, err := e.metrics.ProjectSprints(projectID)

		if !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, se esperaba %v", c.name, err, c.want)
		}
		if got != nil {
			t.Errorf("%s: ProjectSprints = %+v, se esperaba nil", c.name, got)
		}
		checkReads(t, e, c)
	}
}

// RN-37.2 y "Restricciones" · Con tres sprints, el repositorio de ítems
// recibe 2 lecturas (1 de metrics y 1 interna de defect.Manager.ListByProject),
// el de esfuerzos 1 por sprint y el de defectos 1: no se llama a Sprint(id)
// por cada sprint.
func TestProjectSprints_LeeLosItemsYLosDefectosUnaSolaVez(t *testing.T) {
	e := newSummaryEnv()
	projectID, sprints := e.multiSprintProject(t)
	e.resetReads()

	if _, err := e.metrics.ProjectSprints(projectID); err != nil {
		t.Fatalf("ProjectSprints devolvió un error inesperado: %v", err)
	}

	if e.itemSpy.listCalls != 2 {
		t.Errorf("lecturas de ítems = %d, se esperaba 2 (1 de metrics y 1 interna de defect)", e.itemSpy.listCalls)
	}
	if e.effortSpy.listCalls != len(sprints) {
		t.Errorf("lecturas de esfuerzos = %d, se esperaba %d (una por sprint)", e.effortSpy.listCalls, len(sprints))
	}
	if e.defectSpy.listCalls != 1 {
		t.Errorf("lecturas de defectos = %d, se esperaba 1", e.defectSpy.listCalls)
	}
}

// CA-37.5 · Consultar no modifica nada, y modificar la lista devuelta no
// cambia la próxima consulta.
func TestProjectSprints_NoModificaNadaYDevuelveCopias(t *testing.T) {
	e := newSummaryEnv()
	projectID, _ := e.multiSprintProject(t)
	itemsBefore, _ := e.itemRepo.ListByProject(projectID)
	sprintsBefore := e.sprintsOf(projectID)
	defectsBefore, _ := e.defectRepo.List()
	var effortsBefore []effort.Effort
	for _, s := range sprintsBefore {
		found, _ := e.effortRepo.ListBySprint(s.ID)
		effortsBefore = append(effortsBefore, found...)
	}

	first, err := e.metrics.ProjectSprints(projectID)
	if err != nil {
		t.Fatalf("ProjectSprints devolvió un error inesperado: %v", err)
	}
	if len(first) != len(sprintsBefore) {
		t.Fatalf("ProjectSprints devolvió %d sprints, se esperaba %d", len(first), len(sprintsBefore))
	}
	want := make([]SprintWithMetrics, len(first))
	copy(want, first)
	first[0].Sprint.Name = "Modificado"
	first[0].Metrics.Points.Planned = 999
	first[1].Metrics.Hours.Estimated = -1
	first[2] = SprintWithMetrics{}

	again, err := e.metrics.ProjectSprints(projectID)

	expectProjectSprints(t, again, err, want)
	itemsAfter, _ := e.itemRepo.ListByProject(projectID)
	defectsAfter, _ := e.defectRepo.List()
	var effortsAfter []effort.Effort
	for _, s := range e.sprintsOf(projectID) {
		found, _ := e.effortRepo.ListBySprint(s.ID)
		effortsAfter = append(effortsAfter, found...)
	}
	if !reflect.DeepEqual(itemsAfter, itemsBefore) || !reflect.DeepEqual(e.sprintsOf(projectID), sprintsBefore) ||
		!reflect.DeepEqual(effortsAfter, effortsBefore) || !reflect.DeepEqual(defectsAfter, defectsBefore) {
		t.Error("la consulta modificó los ítems, los sprints, los esfuerzos o los defectos")
	}
}
