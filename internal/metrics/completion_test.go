package metrics

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-34 · Calcular el porcentaje de historias completadas (Issue #52)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-34-porcentaje-completadas.md). Los datos se arman
// con metricsEnv de HU-32: los Manager reales en memoria.

// expectCompletion compara M8 con tolerancia, porque el dominio no redondea
// (RN-34.5).
func expectCompletion(t *testing.T, got Completion, items, finished int, percent float64) {
	t.Helper()
	if got.Items != items || got.Finished != finished || !got.PercentAvailable || math.Abs(got.Percent-percent) > 1e-9 {
		t.Errorf("completion = %+v, se esperaba %d de %d ítems terminados, %v %% disponible", got, finished, items, percent)
	}
}

func expectCompletionNotAvailable(t *testing.T, got Completion) {
	t.Helper()
	if got != (Completion{}) {
		t.Errorf("completion = %+v, se esperaba 0 ítems y el porcentaje no disponible", got)
	}
}

// CA-34.1 · Ejemplo de TEC-03: 2 de 4 ítems terminados = 50 %.
func TestSprintCompletion_EjemploDeTEC03(t *testing.T) {
	e := newMetricsEnv()
	s := e.tec03Sprint(t, e.project(t, "Sistema de ventas"))

	got, err := e.metrics.SprintCompletion(s.ID)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expectCompletion(t, got, 4, 2, 50)
}

// CA-34.1 y RN-34.1 · El sprint solo cuenta sus ítems: no los de otro sprint
// ni los que no tienen sprint.
func TestSprintCompletion_SoloCuentaLosItemsDelSprint(t *testing.T) {
	e := newMetricsEnv()
	p := e.project(t, "Sistema de ventas")
	e.closedSprint(t, p, "Sprint 1", 5, 3)
	s := e.openSprint(t, p, "Sprint 2")
	e.item(t, p, "A", 2, backlog.StatusFinished, s.ID)
	e.item(t, p, "B", 1, backlog.StatusPending, s.ID)
	e.item(t, p, "Sin sprint", 8, backlog.StatusFinished, 0)

	got, err := e.metrics.SprintCompletion(s.ID)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expectCompletion(t, got, 2, 1, 50)
}

// CA-34.1 y RN-34.8 · Un sprint CERRADO se calcula igual.
func TestSprintCompletion_SprintCerradoSeCalculaIgual(t *testing.T) {
	e := newMetricsEnv()
	s := e.closedSprint(t, e.project(t, "Sistema de ventas"), "Sprint 1", 5, 3)

	got, err := e.metrics.SprintCompletion(s.ID)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expectCompletion(t, got, 2, 2, 100)
}

// CA-34.1 · Ningún ítem terminado da 0 %, disponible.
func TestSprintCompletion_NingunoTerminadoEsCeroDisponible(t *testing.T) {
	e := newMetricsEnv()
	p := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, p, "Sprint 1")
	e.item(t, p, "A", 5, backlog.StatusPending, s.ID)
	e.item(t, p, "B", 3, backlog.StatusInProgress, s.ID)

	got, err := e.metrics.SprintCompletion(s.ID)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expectCompletion(t, got, 2, 0, 0)
}

// CA-34.3 y RN-34.6 · Un sprint sin ítems no tiene porcentaje disponible, sin
// error.
func TestSprintCompletion_SprintSinItemsNoDisponible(t *testing.T) {
	e := newMetricsEnv()
	s := e.openSprint(t, e.project(t, "Sistema de ventas"), "Sprint 1")

	got, err := e.metrics.SprintCompletion(s.ID)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expectCompletionNotAvailable(t, got)
}

// CA-34.5 · Un sprint inexistente devuelve sprint.ErrSprintNotFound.
func TestSprintCompletion_SprintInexistenteDevuelveErrSprintNotFound(t *testing.T) {
	e := newMetricsEnv()
	e.project(t, "Sistema de ventas")

	for _, id := range []int{99, 0, -1} {
		got, err := e.metrics.SprintCompletion(id)

		if !errors.Is(err, sprint.ErrSprintNotFound) {
			t.Errorf("sprint %d: error = %v, se esperaba %v", id, err, sprint.ErrSprintNotFound)
		}
		if got != (Completion{}) {
			t.Errorf("sprint %d: completion = %+v, se esperaba vacío", id, got)
		}
	}
}

// CA-34.2 y RN-34.2 · El proyecto usa todo su backlog, asignado o no a un
// sprint.
func TestProjectCompletion_UsaTodoElBacklog(t *testing.T) {
	e := newMetricsEnv()
	p := e.project(t, "Sistema de ventas")
	e.closedSprint(t, p, "Sprint 1", 5)
	s := e.openSprint(t, p, "Sprint 2")
	e.item(t, p, "A", 3, backlog.StatusInProgress, s.ID)
	e.item(t, p, "E", 8, backlog.StatusPending, 0)
	e.item(t, p, "F", 2, backlog.StatusFinished, 0)

	got, err := e.metrics.ProjectCompletion(p)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expectCompletion(t, got, 4, 2, 50)
}

// CA-34.2 · Solo cuentan los ítems del proyecto consultado.
func TestProjectCompletion_NoCuentaOtrosProyectos(t *testing.T) {
	e := newMetricsEnv()
	p := e.project(t, "Sistema de ventas")
	otro := e.project(t, "Otro")
	e.item(t, p, "A", 3, backlog.StatusFinished, 0)
	e.item(t, otro, "X", 3, backlog.StatusPending, 0)

	got, err := e.metrics.ProjectCompletion(p)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expectCompletion(t, got, 1, 1, 100)
}

// CA-34.3 y RN-34.6 · Un proyecto sin ítems no tiene porcentaje disponible,
// sin error.
func TestProjectCompletion_ProyectoSinItemsNoDisponible(t *testing.T) {
	e := newMetricsEnv()
	p := e.project(t, "Sistema de ventas")

	got, err := e.metrics.ProjectCompletion(p)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expectCompletionNotAvailable(t, got)
}

// CA-34.5 · Un proyecto inexistente devuelve project.ErrProjectNotFound.
func TestProjectCompletion_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	e := newMetricsEnv()

	for _, id := range []int{99, 0, -1} {
		got, err := e.metrics.ProjectCompletion(id)

		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("proyecto %d: error = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if got != (Completion{}) {
			t.Errorf("proyecto %d: completion = %+v, se esperaba vacío", id, got)
		}
	}
}

// CA-34.4 y RN-34.4 · Los ítems sin Story Points cuentan igual.
func TestCompletion_CuentaItemsSinStoryPoints(t *testing.T) {
	e := newMetricsEnv()
	p := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, p, "Sprint 1")
	e.item(t, p, "A", 0, backlog.StatusFinished, s.ID)
	e.item(t, p, "B", 0, backlog.StatusPending, s.ID)

	sprintGot, err := e.metrics.SprintCompletion(s.ID)
	if err != nil {
		t.Fatalf("sprint: error inesperado: %v", err)
	}
	projectGot, err := e.metrics.ProjectCompletion(p)
	if err != nil {
		t.Fatalf("proyecto: error inesperado: %v", err)
	}

	expectCompletion(t, sprintGot, 2, 1, 50)
	expectCompletion(t, projectGot, 2, 1, 50)
}

// RN-34.5 · El dominio no redondea: 1 de 3 es 33,333…
func TestCompletion_PorcentajeSinRedondear(t *testing.T) {
	e := newMetricsEnv()
	p := e.project(t, "Sistema de ventas")
	e.item(t, p, "A", 1, backlog.StatusFinished, 0)
	e.item(t, p, "B", 1, backlog.StatusPending, 0)
	e.item(t, p, "C", 1, backlog.StatusPending, 0)

	got, err := e.metrics.ProjectCompletion(p)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expectCompletion(t, got, 3, 1, 100.0/3)
}

// CA-34.5 y RN-34.9 · Calcular no modifica los ítems ni los sprints.
func TestCompletion_NoModificaNada(t *testing.T) {
	e := newMetricsEnv()
	p := e.project(t, "Sistema de ventas")
	s := e.tec03Sprint(t, p)
	e.item(t, p, "Sin sprint", 8, backlog.StatusPending, 0)
	itemsBefore, _ := e.itemRepo.ListByProject(p)
	sprintsBefore := e.sprintsOf(p)

	if _, err := e.metrics.SprintCompletion(s.ID); err != nil {
		t.Fatalf("sprint: error inesperado: %v", err)
	}
	if _, err := e.metrics.ProjectCompletion(p); err != nil {
		t.Fatalf("proyecto: error inesperado: %v", err)
	}

	itemsAfter, _ := e.itemRepo.ListByProject(p)
	if !reflect.DeepEqual(itemsAfter, itemsBefore) {
		t.Errorf("los ítems cambiaron: %+v, antes %+v", itemsAfter, itemsBefore)
	}
	if sprintsAfter := e.sprintsOf(p); !reflect.DeepEqual(sprintsAfter, sprintsBefore) {
		t.Errorf("los sprints cambiaron: %+v, antes %+v", sprintsAfter, sprintsBefore)
	}
}

// failingSprintGetRepository falla al leer un sprint por ID.
type failingSprintGetRepository struct {
	sprint.Repository
	err error
}

func (r failingSprintGetRepository) GetByID(int) (sprint.Sprint, error) {
	return sprint.Sprint{}, r.err
}

// CA-34.5 · Una falla al leer el sprint se propaga.
func TestSprintCompletion_ErrorDelRepositorioDeSprintsSePropaga(t *testing.T) {
	errRepo := errors.New("base caída")
	e := newMetricsEnvWith(backlog.NewMemoryRepository(), failingSprintGetRepository{Repository: sprint.NewMemoryRepository(), err: errRepo}, effort.NewMemoryRepository())

	got, err := e.metrics.SprintCompletion(1)

	if !errors.Is(err, errRepo) {
		t.Errorf("error = %v, se esperaba %v", err, errRepo)
	}
	if got != (Completion{}) {
		t.Errorf("completion = %+v, se esperaba vacío", got)
	}
}

// CA-34.5 · Una falla al leer los ítems se propaga en los dos niveles.
func TestCompletion_ErrorDelRepositorioDeItemsSePropaga(t *testing.T) {
	errRepo := errors.New("base caída")
	e := newMetricsEnvWith(failingItemRepository{Repository: backlog.NewMemoryRepository(), err: errRepo}, sprint.NewMemoryRepository(), effort.NewMemoryRepository())
	p := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, p, "Sprint 1")

	if got, err := e.metrics.SprintCompletion(s.ID); !errors.Is(err, errRepo) || got != (Completion{}) {
		t.Errorf("sprint: completion = %+v, error = %v, se esperaba vacío y %v", got, err, errRepo)
	}
	if got, err := e.metrics.ProjectCompletion(p); !errors.Is(err, errRepo) || got != (Completion{}) {
		t.Errorf("proyecto: completion = %+v, error = %v, se esperaba vacío y %v", got, err, errRepo)
	}
}
