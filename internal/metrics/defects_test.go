package metrics

import (
	"errors"
	"reflect"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-47 · Calcular los defectos detectados y resueltos (Issue #128)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-47-defectos-detectados-resueltos.md). Los
// defectos se guardan con Save del repositorio en memoria, así un defecto
// puede quedar en RESUELTO o con un sprint de resolución cargado sin pasar
// por HU-30 (sección "Dependencias" de la SDD).

// saveDefect guarda un defecto de la historia itemID, detectado en
// detectionSprintID, con ese estado y ese sprint de resolución (0 = ninguno).
func (e *metricsEnv) saveDefect(t *testing.T, itemID, detectionSprintID int, status defect.Status, resolutionSprintID int) defect.Defect {
	t.Helper()
	d, err := e.defectRepo.Save(defect.Defect{
		Description:        "Defecto",
		Severity:           defect.SeverityMedium,
		Status:             status,
		ItemID:             itemID,
		DetectionSprintID:  detectionSprintID,
		ResolutionSprintID: resolutionSprintID,
	})
	if err != nil {
		t.Fatalf("guardar el defecto: %v", err)
	}
	return d
}

func expectDefects(t *testing.T, got Defects, err error, detected, resolved int) {
	t.Helper()
	if err != nil {
		t.Fatalf("err = %v, se esperaba nil", err)
	}
	if want := (Defects{Detected: detected, Resolved: resolved}); got != want {
		t.Errorf("defectos = %+v, se esperaba %+v", got, want)
	}
}

// CA-47.1 · Ejemplo de TEC-03: un defecto detectado y resuelto en el sprint y
// otro abierto dan 2 detectados y 1 resuelto (RN-47.1 y RN-47.2).
func TestSprintDefects_CuentaDetectadosYResueltosDelSprint(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	sp := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, backlog.StatusPending, sp.ID)
	e.saveDefect(t, it.ID, sp.ID, defect.StatusResolved, sp.ID)
	e.saveDefect(t, it.ID, sp.ID, defect.StatusOpen, 0)

	got, err := e.metrics.SprintDefects(sp.ID)

	expectDefects(t, got, err, 2, 1)
}

// CA-47.1 · Los defectos detectados en otro sprint del proyecto no cuentan
// (RN-47.1).
func TestSprintDefects_NoCuentaDefectosDeOtroSprint(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s1 := e.closedSprint(t, projectID, "Sprint 1")
	s2 := e.openSprint(t, projectID, "Sprint 2")
	it := e.item(t, projectID, "A", 0, backlog.StatusPending, 0)
	e.saveDefect(t, it.ID, s1.ID, defect.StatusOpen, 0)
	e.saveDefect(t, it.ID, s2.ID, defect.StatusOpen, 0)

	got, err := e.metrics.SprintDefects(s2.ID)

	expectDefects(t, got, err, 1, 0)
}

// CA-47.1 · Un sprint CERRADO se calcula igual, con el estado actual de los
// defectos (RN-47.9).
func TestSprintDefects_SprintCerradoSeCalculaIgual(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	sp := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, backlog.StatusPending, sp.ID)
	e.saveDefect(t, it.ID, sp.ID, defect.StatusResolved, sp.ID)
	if _, err := e.sprints.Close(sp.ID); err != nil {
		t.Fatalf("cerrar el sprint: %v", err)
	}

	got, err := e.metrics.SprintDefects(sp.ID)

	expectDefects(t, got, err, 1, 1)
}

// CA-47.2 · El proyecto cuenta todos sus defectos como detectados y los
// RESUELTO como resueltos (RN-47.3 y RN-47.4).
func TestProjectDefects_CuentaTodosLosDefectosDelProyecto(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s1 := e.closedSprint(t, projectID, "Sprint 1")
	s2 := e.openSprint(t, projectID, "Sprint 2")
	a := e.item(t, projectID, "A", 0, backlog.StatusPending, 0)
	b := e.item(t, projectID, "B", 0, backlog.StatusPending, s2.ID)
	e.saveDefect(t, a.ID, s1.ID, defect.StatusResolved, s2.ID)
	e.saveDefect(t, b.ID, s2.ID, defect.StatusOpen, 0)
	e.saveDefect(t, b.ID, s2.ID, defect.StatusInCorrection, 0)

	got, err := e.metrics.ProjectDefects(projectID)

	expectDefects(t, got, err, 3, 1)
}

// CA-47.2 · Los defectos de otro proyecto no cuentan (RN-47.3).
func TestProjectDefects_NoCuentaDefectosDeOtroProyecto(t *testing.T) {
	e := newMetricsEnv()
	sales := e.project(t, "Sistema de ventas")
	stock := e.project(t, "Inventario")
	salesSprint := e.openSprint(t, sales, "Sprint 1")
	stockSprint := e.openSprint(t, stock, "Sprint A")
	salesItem := e.item(t, sales, "A", 0, backlog.StatusPending, 0)
	stockItem := e.item(t, stock, "Z", 0, backlog.StatusPending, 0)
	e.saveDefect(t, salesItem.ID, salesSprint.ID, defect.StatusOpen, 0)
	e.saveDefect(t, stockItem.ID, stockSprint.ID, defect.StatusResolved, stockSprint.ID)

	got, err := e.metrics.ProjectDefects(sales)

	expectDefects(t, got, err, 1, 0)
}

// CA-47.1 y CA-47.2 · El defecto de una historia TERMINADO cuenta igual
// (casos límite de la SDD).
func TestDefects_DefectoDeHistoriaTerminadaCuenta(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	sp := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, backlog.StatusFinished, 0)
	e.saveDefect(t, it.ID, sp.ID, defect.StatusResolved, sp.ID)

	got, err := e.metrics.SprintDefects(sp.ID)
	expectDefects(t, got, err, 1, 1)
	got, err = e.metrics.ProjectDefects(projectID)
	expectDefects(t, got, err, 1, 1)
}

// CA-47.3 · Un sprint sin defectos da cero, sin error (RN-47.7).
func TestSprintDefects_SinDefectosDaCero(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	sp := e.openSprint(t, projectID, "Sprint 1")
	e.item(t, projectID, "A", 0, backlog.StatusPending, sp.ID)

	got, err := e.metrics.SprintDefects(sp.ID)

	expectDefects(t, got, err, 0, 0)
}

// CA-47.3 · Un proyecto sin defectos da cero, sin error (RN-47.7).
func TestProjectDefects_SinDefectosDaCero(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")

	got, err := e.metrics.ProjectDefects(projectID)

	expectDefects(t, got, err, 0, 0)
}

// CA-47.4 · Un defecto detectado en S1 y resuelto en S2 cuenta como detectado
// en S1, como resuelto en S2 y una vez en el proyecto (RN-47.6).
func TestDefects_DeteccionYResolucionEnSprintsDistintos(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s1 := e.closedSprint(t, projectID, "Sprint 1")
	s2 := e.openSprint(t, projectID, "Sprint 2")
	it := e.item(t, projectID, "A", 0, backlog.StatusPending, 0)
	e.saveDefect(t, it.ID, s1.ID, defect.StatusResolved, s2.ID)

	got, err := e.metrics.SprintDefects(s1.ID)
	expectDefects(t, got, err, 1, 0)
	got, err = e.metrics.SprintDefects(s2.ID)
	expectDefects(t, got, err, 0, 1)
	got, err = e.metrics.ProjectDefects(projectID)
	expectDefects(t, got, err, 1, 1)
}

// CA-47.4 · Solo cuenta como resuelto un defecto en RESUELTO: uno ABIERTO o
// EN_CORRECCION con sprint de resolución cargado no cuenta (RN-47.5).
func TestDefects_SoloCuentaComoResueltoElEstadoResuelto(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	sp := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, backlog.StatusPending, sp.ID)
	e.saveDefect(t, it.ID, sp.ID, defect.StatusOpen, sp.ID)
	e.saveDefect(t, it.ID, sp.ID, defect.StatusInCorrection, sp.ID)
	e.saveDefect(t, it.ID, sp.ID, defect.StatusResolved, sp.ID)

	got, err := e.metrics.SprintDefects(sp.ID)
	expectDefects(t, got, err, 3, 1)
	got, err = e.metrics.ProjectDefects(projectID)
	expectDefects(t, got, err, 3, 1)
}

// CA-47.5 · Un sprint inexistente, 0 o negativo devuelve
// sprint.ErrSprintNotFound.
func TestSprintDefects_SprintInexistenteDevuelveErrSprintNotFound(t *testing.T) {
	e := newMetricsEnv()
	e.project(t, "Sistema de ventas")

	for _, id := range []int{99, 0, -1} {
		got, err := e.metrics.SprintDefects(id)

		if !errors.Is(err, sprint.ErrSprintNotFound) {
			t.Errorf("SprintDefects(%d): err = %v, se esperaba %v", id, err, sprint.ErrSprintNotFound)
		}
		if got != (Defects{}) {
			t.Errorf("SprintDefects(%d) = %+v, se esperaba el valor cero", id, got)
		}
	}
}

// CA-47.5 · Un proyecto inexistente, 0 o negativo devuelve
// project.ErrProjectNotFound.
func TestProjectDefects_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	e := newMetricsEnv()

	for _, id := range []int{99, 0, -1} {
		got, err := e.metrics.ProjectDefects(id)

		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("ProjectDefects(%d): err = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if got != (Defects{}) {
			t.Errorf("ProjectDefects(%d) = %+v, se esperaba el valor cero", id, got)
		}
	}
}

// CA-47.5 · Consultar no modifica los defectos ni los sprints (RN-47.10).
func TestDefects_ConsultarNoModificaNada(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	sp := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, backlog.StatusPending, sp.ID)
	e.saveDefect(t, it.ID, sp.ID, defect.StatusResolved, sp.ID)
	e.saveDefect(t, it.ID, sp.ID, defect.StatusOpen, 0)
	defectsBefore, _ := e.defectRepo.List()
	sprintsBefore := e.sprintsOf(projectID)

	expectDefects(t, mustSprintDefects(t, e, sp.ID), nil, 2, 1)
	expectDefects(t, mustProjectDefects(t, e, projectID), nil, 2, 1)

	defectsAfter, _ := e.defectRepo.List()
	if !reflect.DeepEqual(defectsAfter, defectsBefore) {
		t.Errorf("defectos = %+v, se esperaba %+v", defectsAfter, defectsBefore)
	}
	if sprintsAfter := e.sprintsOf(projectID); !reflect.DeepEqual(sprintsAfter, sprintsBefore) {
		t.Errorf("sprints = %+v, se esperaba %+v", sprintsAfter, sprintsBefore)
	}
}

func mustSprintDefects(t *testing.T, e *metricsEnv, sprintID int) Defects {
	t.Helper()
	got, err := e.metrics.SprintDefects(sprintID)
	if err != nil {
		t.Fatalf("SprintDefects(%d): %v", sprintID, err)
	}
	return got
}

func mustProjectDefects(t *testing.T, e *metricsEnv, projectID int) Defects {
	t.Helper()
	got, err := e.metrics.ProjectDefects(projectID)
	if err != nil {
		t.Fatalf("ProjectDefects(%d): %v", projectID, err)
	}
	return got
}

// failingDefectListRepository falla al leer los defectos.
type failingDefectListRepository struct {
	defect.Repository
	err error
}

func (r failingDefectListRepository) List() ([]defect.Defect, error) {
	return nil, r.err
}

// Condiciones de error de la SDD · Una falla al leer los defectos se propaga
// y no se devuelve un resultado.
func TestDefects_ErrorAlLeerLosDefectosSePropaga(t *testing.T) {
	errRepo := errors.New("base caída")
	e := newMetricsEnvWithDefects(backlog.NewMemoryRepository(), sprint.NewMemoryRepository(), effort.NewMemoryRepository(),
		failingDefectListRepository{Repository: defect.NewMemoryRepository(), err: errRepo})
	projectID := e.project(t, "Sistema de ventas")
	sp := e.openSprint(t, projectID, "Sprint 1")

	if got, err := e.metrics.SprintDefects(sp.ID); !errors.Is(err, errRepo) || got != (Defects{}) {
		t.Errorf("SprintDefects = %+v, %v; se esperaba el valor cero y %v", got, err, errRepo)
	}
	if got, err := e.metrics.ProjectDefects(projectID); !errors.Is(err, errRepo) || got != (Defects{}) {
		t.Errorf("ProjectDefects = %+v, %v; se esperaba el valor cero y %v", got, err, errRepo)
	}
}

// Condiciones de error de la SDD · Una falla al leer el sprint se propaga.
func TestSprintDefects_ErrorAlLeerElSprintSePropaga(t *testing.T) {
	errRepo := errors.New("base caída")
	e := newMetricsEnvWith(backlog.NewMemoryRepository(), failingSprintGetRepository{Repository: sprint.NewMemoryRepository(), err: errRepo}, effort.NewMemoryRepository())

	got, err := e.metrics.SprintDefects(1)

	if !errors.Is(err, errRepo) || got != (Defects{}) {
		t.Errorf("SprintDefects = %+v, %v; se esperaba el valor cero y %v", got, err, errRepo)
	}
}
