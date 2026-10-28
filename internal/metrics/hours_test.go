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

// HU-33 · Calcular horas estimadas, horas reales y desviación (Issue #51)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-33-horas-desviacion.md). Los datos se arman con
// metricsEnv de HU-32, que usa los Manager reales en memoria.

// CA-33.1, CA-33.2 y CA-33.3 · Ejemplo de TEC-03 en un sprint.
func TestSprintHours_EjemploDeTEC03(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.itemWithHours(t, projectID, "A", 10, s.ID)
	b := e.itemWithHours(t, projectID, "B", 6, s.ID)
	c := e.itemWithHours(t, projectID, "C", 0, s.ID)
	e.itemWithHours(t, projectID, "D", 4, s.ID)
	e.registerEffort(t, s.ID, a.ID, 12)
	e.registerEffort(t, s.ID, b.ID, 5)
	e.registerEffort(t, s.ID, c.ID, 3)

	got, err := e.metrics.SprintHours(s.ID)

	if err != nil {
		t.Fatalf("SprintHours devolvió un error inesperado: %v", err)
	}
	want := Hours{Estimated: 20, Actual: 20, Deviation: 0, DeviationPercent: 0, Unestimated: 1, DeviationAvailable: true}
	if got != want {
		t.Errorf("horas = %+v, se esperaba %+v", got, want)
	}
}

// CA-33.7 · Un sprint sin ítems no tiene desviación.
func TestSprintHours_SinItems(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")

	got, err := e.metrics.SprintHours(s.ID)

	if err != nil {
		t.Fatalf("SprintHours devolvió un error inesperado: %v", err)
	}
	if got != (Hours{}) {
		t.Errorf("horas = %+v, se esperaba vacío", got)
	}
}

// CA-33.7 · Un proyecto sin ítems no tiene desviación.
func TestProjectHours_SinItems(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")

	got, err := e.metrics.ProjectHours(projectID)

	if err != nil {
		t.Fatalf("ProjectHours devolvió un error inesperado: %v", err)
	}
	if got != (Hours{}) {
		t.Errorf("horas = %+v, se esperaba vacío", got)
	}
}

// CA-33.6 · El proyecto suma todo el backlog y los esfuerzos de sus sprints.
func TestProjectHours_SumaTodoElBacklogYLosEsfuerzos(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.itemWithHours(t, projectID, "A", 10, s.ID)
	b := e.itemWithHours(t, projectID, "B", 6, s.ID)
	e.registerEffort(t, s.ID, a.ID, 12)
	e.registerEffort(t, s.ID, b.ID, 5)
	e.itemWithHours(t, projectID, "E", 4, 0)
	e.itemWithHours(t, projectID, "F", 0, 0)

	got, err := e.metrics.ProjectHours(projectID)

	if err != nil {
		t.Fatalf("ProjectHours devolvió un error inesperado: %v", err)
	}
	want := Hours{Estimated: 20, Actual: 17, Deviation: -3, DeviationPercent: -15, Unestimated: 1, DeviationAvailable: true}
	if got != want {
		t.Errorf("horas = %+v, se esperaba %+v", got, want)
	}
}

// CA-33.4 · Sin horas estimadas, la desviación no está disponible.
func TestSprintHours_SinHorasEstimadasNoEstaDisponible(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.itemWithHours(t, projectID, "A", 0, s.ID)
	e.registerEffort(t, s.ID, a.ID, 3)

	got, err := e.metrics.SprintHours(s.ID)

	if err != nil {
		t.Fatalf("SprintHours devolvió un error inesperado: %v", err)
	}
	want := Hours{Estimated: 0, Actual: 3, Deviation: 0, DeviationPercent: 0, Unestimated: 1, DeviationAvailable: false}
	if got != want {
		t.Errorf("horas = %+v, se esperaba %+v", got, want)
	}
}

// CA-33.4 · Sin esfuerzos, la desviación no está disponible.
func TestSprintHours_SinEsfuerzosNoEstaDisponible(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.itemWithHours(t, projectID, "A", 10, s.ID)

	got, err := e.metrics.SprintHours(s.ID)

	if err != nil {
		t.Fatalf("SprintHours devolvió un error inesperado: %v", err)
	}
	want := Hours{Estimated: 10, Actual: 0, Deviation: 0, DeviationPercent: 0, Unestimated: 0, DeviationAvailable: false}
	if got != want {
		t.Errorf("horas = %+v, se esperaba %+v", got, want)
	}
}

// CA-33.2 · Un sprint solo suma los esfuerzos registrados en ese sprint.
func TestSprintHours_SoloSumaLosEsfuerzosDelSprint(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s1 := e.openSprint(t, projectID, "Sprint 1")
	a := e.itemWithHours(t, projectID, "A", 10, s1.ID)
	e.registerEffort(t, s1.ID, a.ID, 2)
	if _, err := e.sprints.Close(s1.ID); err != nil {
		t.Fatalf("cerrar el sprint %q: %v", s1.Name, err)
	}
	s2 := e.openSprint(t, projectID, "Sprint 2")
	b := e.itemWithHours(t, projectID, "B", 6, s2.ID)
	e.registerEffort(t, s2.ID, b.ID, 4)

	got, err := e.metrics.SprintHours(s1.ID)

	if err != nil {
		t.Fatalf("SprintHours devolvió un error inesperado: %v", err)
	}
	want := Hours{Estimated: 10, Actual: 2, Deviation: -8, DeviationPercent: -80, Unestimated: 0, DeviationAvailable: true}
	if got != want {
		t.Errorf("horas = %+v, se esperaba %+v", got, want)
	}
}

// CA-33.8 · Un sprint inexistente devuelve ErrSprintNotFound.
func TestSprintHours_SprintInexistenteDevuelveErrSprintNotFound(t *testing.T) {
	e := newMetricsEnv()

	for _, id := range []int{99, 0, -1} {
		got, err := e.metrics.SprintHours(id)

		if !errors.Is(err, sprint.ErrSprintNotFound) {
			t.Errorf("SprintHours(%d): error = %v, se esperaba %v", id, err, sprint.ErrSprintNotFound)
		}
		if got != (Hours{}) {
			t.Errorf("SprintHours(%d) = %+v, se esperaba vacío", id, got)
		}
	}
}

// CA-33.8 · Un proyecto inexistente devuelve ErrProjectNotFound.
func TestProjectHours_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	e := newMetricsEnv()

	for _, id := range []int{99, 0, -1} {
		got, err := e.metrics.ProjectHours(id)

		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("ProjectHours(%d): error = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if got != (Hours{}) {
			t.Errorf("ProjectHours(%d) = %+v, se esperaba vacío", id, got)
		}
	}
}

// failingEffortRepository falla al leer los esfuerzos de un sprint.
type failingEffortRepository struct {
	effort.Repository
	err error
}

func (r failingEffortRepository) ListBySprint(int) ([]effort.Effort, error) {
	return nil, r.err
}

// CA-33.8 · Una falla del repositorio de esfuerzos se propaga en los dos niveles.
func TestHours_ErrorDelRepositorioDeEsfuerzosSePropaga(t *testing.T) {
	errRepo := errors.New("base caída")
	e := newMetricsEnvWith(backlog.NewMemoryRepository(), sprint.NewMemoryRepository(), failingEffortRepository{Repository: effort.NewMemoryRepository(), err: errRepo})
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.itemWithHours(t, projectID, "A", 10, s.ID)

	if got, err := e.metrics.SprintHours(s.ID); !errors.Is(err, errRepo) || got != (Hours{}) {
		t.Errorf("sprint: horas = %+v, error = %v, se esperaba vacío y %v", got, err, errRepo)
	}
	if got, err := e.metrics.ProjectHours(projectID); !errors.Is(err, errRepo) || got != (Hours{}) {
		t.Errorf("proyecto: horas = %+v, error = %v, se esperaba vacío y %v", got, err, errRepo)
	}
}

// CA-33.8 · Consultar no modifica los ítems, los sprints ni los esfuerzos.
func TestHours_NoModificaNada(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.itemWithHours(t, projectID, "A", 10, s.ID)
	e.itemWithHours(t, projectID, "B", 4, 0)
	e.registerEffort(t, s.ID, a.ID, 12)
	itemsBefore, err := e.itemRepo.ListByProject(projectID)
	if err != nil {
		t.Fatal(err)
	}
	sprintsBefore := e.sprintsOf(projectID)
	effortsBefore, err := e.effortRepo.ListBySprint(s.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := e.metrics.SprintHours(s.ID); err != nil {
		t.Fatalf("SprintHours: %v", err)
	}
	if _, err := e.metrics.ProjectHours(projectID); err != nil {
		t.Fatalf("ProjectHours: %v", err)
	}

	itemsAfter, err := e.itemRepo.ListByProject(projectID)
	if err != nil {
		t.Fatal(err)
	}
	effortsAfter, err := e.effortRepo.ListBySprint(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(itemsAfter, itemsBefore) || !reflect.DeepEqual(e.sprintsOf(projectID), sprintsBefore) || !reflect.DeepEqual(effortsAfter, effortsBefore) {
		t.Error("consultar las métricas modificó los datos")
	}
}

// CA-33.10 · El resultado es una copia por valor: modificarlo no altera el
// estado interno ni la próxima consulta.
func TestSprintHours_DevuelveUnaCopiaPorValor(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.itemWithHours(t, projectID, "A", 10, s.ID)
	e.registerEffort(t, s.ID, a.ID, 12)

	got, err := e.metrics.SprintHours(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	got.Estimated = 999

	again, err := e.metrics.SprintHours(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again.Estimated != 10 {
		t.Errorf("horas estimadas = %v, se esperaba 10", again.Estimated)
	}
}

// CA-33.9 · El dominio no redondea: 0.1 + 0.2 no es 0.3 exacto.
func TestSprintHours_ElDominioNoRedondea(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.itemWithHours(t, projectID, "A", 0.1, s.ID)
	e.itemWithHours(t, projectID, "B", 0.2, s.ID)
	e.registerEffort(t, s.ID, a.ID, 0.3)

	got, err := e.metrics.SprintHours(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := 0.1
	want += 0.2
	if got.Estimated != want {
		t.Errorf("horas estimadas = %v, se esperaba el valor exacto %v", got.Estimated, want)
	}
	if !almostEqual(got.Actual, 0.3) {
		t.Errorf("horas reales = %v, se esperaba 0.3", got.Actual)
	}
}
