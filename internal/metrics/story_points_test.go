package metrics

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-32 · Calcular Story Points planificados, completados y velocidad (Issue #50)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-32-sp-velocidad.md). Los datos se arman con los
// Manager reales de proyectos, backlog y sprints en memoria, así que los
// ítems y los sprints pasan por las mismas reglas que en la aplicación.

var (
	testStart = time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	testEnd   = time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
)

// metricsEnv arma los Manager en memoria y el Manager de métricas.
type metricsEnv struct {
	itemRepo   backlog.Repository
	sprintRepo sprint.Repository
	effortRepo effort.Repository
	defectRepo defect.Repository
	projects   *project.Manager
	items      *backlog.Manager
	sprints    *sprint.Manager
	efforts    *effort.Manager
	defects    *defect.Manager
	metrics    *Manager
}

func newMetricsEnv() *metricsEnv {
	return newMetricsEnvWith(backlog.NewMemoryRepository(), sprint.NewMemoryRepository(), effort.NewMemoryRepository())
}

// newMetricsEnvWith permite usar repositorios que fallan (CA-32.7).
func newMetricsEnvWith(itemRepo backlog.Repository, sprintRepo sprint.Repository, effortRepo effort.Repository) *metricsEnv {
	return newMetricsEnvWithDefects(itemRepo, sprintRepo, effortRepo, defect.NewMemoryRepository())
}

// newMetricsEnvWithDefects permite además usar un repositorio de defectos que
// falla (HU-47).
func newMetricsEnvWithDefects(itemRepo backlog.Repository, sprintRepo sprint.Repository, effortRepo effort.Repository, defectRepo defect.Repository) *metricsEnv {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(itemRepo, projects)
	sprints := sprint.NewManager(sprintRepo, projects, items)
	efforts := effort.NewManager(effortRepo, sprints, items)
	defects := defect.NewManager(defectRepo, items, sprints)
	return &metricsEnv{itemRepo: itemRepo, sprintRepo: sprintRepo, effortRepo: effortRepo, defectRepo: defectRepo, projects: projects, items: items, sprints: sprints, efforts: efforts, defects: defects, metrics: NewManager(items, sprints, efforts, defects)}
}

// sprintsOf lee los sprints del proyecto con GetByID del repositorio (IDs
// correlativos desde 1), sin usar ListByProject, que es parte de lo que se
// prueba.
func (e *metricsEnv) sprintsOf(projectID int) []sprint.Sprint {
	var found []sprint.Sprint
	for id := 1; ; id++ {
		sp, err := e.sprintRepo.GetByID(id)
		if err != nil {
			return found
		}
		if sp.ProjectID == projectID {
			found = append(found, sp)
		}
	}
}

func (e *metricsEnv) project(t *testing.T, name string) int {
	t.Helper()
	p, err := e.projects.Create(name, "")
	if err != nil {
		t.Fatalf("crear el proyecto %q: %v", name, err)
	}
	return p.ID
}

func (e *metricsEnv) openSprint(t *testing.T, projectID int, name string) sprint.Sprint {
	t.Helper()
	s, err := e.sprints.Create(projectID, name, testStart, testEnd, "Goal de "+name)
	if err != nil {
		t.Fatalf("crear el sprint %q: %v", name, err)
	}
	return s
}

// item crea un ítem con esos story points (0 = sin estimar), lo asigna al
// sprint si sprintID no es 0 y lo lleva al estado pedido con HU-08.
func (e *metricsEnv) item(t *testing.T, projectID int, title string, points int, status backlog.Status, sprintID int) backlog.Item {
	t.Helper()
	it, err := e.items.Create(projectID, title, "", "MEDIA", []string{"Criterio de " + title})
	if err != nil {
		t.Fatalf("crear el ítem %q: %v", title, err)
	}
	if points > 0 {
		if it, err = e.items.AssignStoryPoints(it.ID, points); err != nil {
			t.Fatalf("asignar %d story points a %q: %v", points, title, err)
		}
	}
	if sprintID != 0 {
		if it, err = e.sprints.AssignItem(sprintID, it.ID); err != nil {
			t.Fatalf("asignar %q al sprint %d: %v", title, sprintID, err)
		}
	}
	steps := map[backlog.Status][]backlog.Status{
		backlog.StatusPending:    nil,
		backlog.StatusInProgress: {backlog.StatusInProgress},
		backlog.StatusFinished:   {backlog.StatusInProgress, backlog.StatusFinished},
	}
	for _, next := range steps[status] {
		if it, err = e.items.ChangeStatus(it.ID, string(next)); err != nil {
			t.Fatalf("pasar %q a %s: %v", title, next, err)
		}
	}
	return it
}

// itemWithHours crea un ítem con las horas estimadas indicadas (0 = sin
// horas), lo asigna al sprint si sprintID no es 0 y lo devuelve.
func (e *metricsEnv) itemWithHours(t *testing.T, projectID int, title string, hours float64, sprintID int) backlog.Item {
	t.Helper()
	it, err := e.items.Create(projectID, title, "", "MEDIA", []string{"Criterio de " + title})
	if err != nil {
		t.Fatalf("crear el ítem %q: %v", title, err)
	}
	if it, err = e.items.Update(it.ID, title, "", "MEDIA", []string{"Criterio de " + title}, hours); err != nil {
		t.Fatalf("asignar %v horas a %q: %v", hours, title, err)
	}
	if sprintID != 0 {
		if it, err = e.sprints.AssignItem(sprintID, it.ID); err != nil {
			t.Fatalf("asignar %q al sprint %d: %v", title, sprintID, err)
		}
	}
	return it
}

// registerEffort registra horas del integrante de prueba en un sprint abierto.
// Con 0 no registra nada, para reutilizar las tablas de BDD.
func (e *metricsEnv) registerEffort(t *testing.T, sprintID, itemID int, hours float64) effort.Effort {
	t.Helper()
	if hours == 0 {
		return effort.Effort{}
	}
	saved, err := e.efforts.Register(sprintID, itemID, "García Santiago", testStart, "Actividad de prueba", hours)
	if err != nil {
		t.Fatalf("registrar %v horas en el ítem %d: %v", hours, itemID, err)
	}
	return saved
}

// closedSprint crea un sprint con un ítem TERMINADO por cada valor de
// points y lo cierra (HU-13). Sin points, el sprint queda sin ítems.
func (e *metricsEnv) closedSprint(t *testing.T, projectID int, name string, points ...int) sprint.Sprint {
	t.Helper()
	s := e.openSprint(t, projectID, name)
	for i, p := range points {
		e.item(t, projectID, name+" ítem "+strconv.Itoa(i+1), p, backlog.StatusFinished, s.ID)
	}
	closed, err := e.sprints.Close(s.ID)
	if err != nil {
		t.Fatalf("cerrar el sprint %q: %v", name, err)
	}
	return closed
}

// tec03Sprint arma el ejemplo de docs/metricas.md en un sprint abierto: A (5,
// TERMINADO), B (3, TERMINADO), C (2, EN_CURSO) y D (sin estimar, PENDIENTE).
func (e *metricsEnv) tec03Sprint(t *testing.T, projectID int) sprint.Sprint {
	t.Helper()
	s := e.openSprint(t, projectID, "Sprint 1")
	e.item(t, projectID, "A", 5, backlog.StatusFinished, s.ID)
	e.item(t, projectID, "B", 3, backlog.StatusFinished, s.ID)
	e.item(t, projectID, "C", 2, backlog.StatusInProgress, s.ID)
	e.item(t, projectID, "D", 0, backlog.StatusPending, s.ID)
	return s
}

// CA-32.1, CA-32.2 y CA-32.3 · Ejemplo de TEC-03 en un sprint.
func TestSprintStoryPoints_EjemploDeTEC03(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.tec03Sprint(t, projectID)

	got, err := e.metrics.SprintStoryPoints(s.ID)
	if err != nil {
		t.Fatalf("SprintStoryPoints devolvió un error inesperado: %v", err)
	}
	want := SprintPoints{SprintID: s.ID, Planned: 10, Completed: 8, Unestimated: 1}
	if got != want {
		t.Errorf("SprintStoryPoints = %+v, se esperaba %+v", got, want)
	}
}

// CA-32.1 · Un sprint solo cuenta sus ítems: los de otro sprint y los sin
// asignar no suman.
func TestSprintStoryPoints_SoloCuentaLosItemsDelSprint(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	first := e.closedSprint(t, projectID, "Sprint 1", 13)
	second := e.openSprint(t, projectID, "Sprint 2")
	e.item(t, projectID, "En el sprint 2", 3, backlog.StatusPending, second.ID)
	e.item(t, projectID, "Sin sprint", 8, backlog.StatusPending, 0)

	got, err := e.metrics.SprintStoryPoints(second.ID)
	if err != nil {
		t.Fatalf("SprintStoryPoints devolvió un error inesperado: %v", err)
	}
	want := SprintPoints{SprintID: second.ID, Planned: 3, Completed: 0, Unestimated: 0}
	if got != want {
		t.Errorf("SprintStoryPoints = %+v, se esperaba %+v (el sprint %d no tiene que sumar)", got, want, first.ID)
	}
}

// CA-32.4 · Un sprint sin ítems devuelve cero, sin error.
func TestSprintStoryPoints_SprintSinItemsDevuelveCero(t *testing.T) {
	e := newMetricsEnv()
	s := e.openSprint(t, e.project(t, "Sistema de ventas"), "Sprint 1")

	got, err := e.metrics.SprintStoryPoints(s.ID)
	if err != nil {
		t.Fatalf("SprintStoryPoints devolvió un error inesperado: %v", err)
	}
	if want := (SprintPoints{SprintID: s.ID}); got != want {
		t.Errorf("SprintStoryPoints = %+v, se esperaba %+v", got, want)
	}
}

// CA-32.7 · Un sprint inexistente, incluidos 0 y negativos, da
// ErrSprintNotFound.
func TestSprintStoryPoints_SprintInexistenteDevuelveErrSprintNotFound(t *testing.T) {
	e := newMetricsEnv()

	for _, id := range []int{99, 0, -1} {
		if _, err := e.metrics.SprintStoryPoints(id); !errors.Is(err, sprint.ErrSprintNotFound) {
			t.Errorf("SprintStoryPoints(%d): err = %v, se esperaba %v", id, err, sprint.ErrSprintNotFound)
		}
	}
}

// CA-32.1, CA-32.2 y CA-32.3 · El nivel proyecto usa todo el backlog,
// asignado o no a un sprint (decisión de TEC-03).
func TestProjectStoryPoints_UsaTodoElBacklog(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	e.tec03Sprint(t, projectID)
	e.item(t, projectID, "E", 8, backlog.StatusPending, 0)
	e.item(t, projectID, "F", 2, backlog.StatusFinished, 0)
	e.item(t, projectID, "G", 0, backlog.StatusFinished, 0)

	got, err := e.metrics.ProjectStoryPoints(projectID)
	if err != nil {
		t.Fatalf("ProjectStoryPoints devolvió un error inesperado: %v", err)
	}
	if got.ProjectID != projectID || got.Planned != 20 || got.Completed != 10 || got.Unestimated != 2 {
		t.Errorf("ProjectStoryPoints = %+v, se esperaba planificados 20, completados 10 y 2 sin estimar", got)
	}
}

// CA-32.4 y CA-32.6 · Un proyecto sin ítems devuelve cero y la velocidad no
// está disponible, sin error.
func TestProjectStoryPoints_ProyectoSinItemsDevuelveCero(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")

	got, err := e.metrics.ProjectStoryPoints(projectID)
	if err != nil {
		t.Fatalf("ProjectStoryPoints devolvió un error inesperado: %v", err)
	}
	if want := (ProjectPoints{ProjectID: projectID}); got != want {
		t.Errorf("ProjectStoryPoints = %+v, se esperaba %+v", got, want)
	}
}

// CA-32.5 · La velocidad usa los últimos 3 sprints cerrados del proyecto,
// incluidos los cerrados sin ítems, y no ve los de otros proyectos.
func TestProjectStoryPoints_VelocidadConLosUltimosSprintsCerrados(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	otro := e.project(t, "Otro proyecto")
	e.closedSprint(t, otro, "Sprint de otro proyecto", 21)
	e.closedSprint(t, projectID, "Sprint 1", 2)
	e.closedSprint(t, projectID, "Sprint 2")
	e.closedSprint(t, projectID, "Sprint 3", 5, 5)
	e.closedSprint(t, projectID, "Sprint 4", 8, 3, 1)

	got, err := e.metrics.ProjectStoryPoints(projectID)
	if err != nil {
		t.Fatalf("ProjectStoryPoints devolvió un error inesperado: %v", err)
	}
	// Sprints 4, 3 y 2: (12 + 10 + 0) / 3.
	if !got.VelocityAvailable || got.Velocity != 22.0/3 || got.ClosedSprints != 3 {
		t.Errorf("velocidad = %v (disponible %v) con %d sprints, se esperaba %v con 3", got.Velocity, got.VelocityAvailable, got.ClosedSprints, 22.0/3)
	}
}

// CA-32.5 y CA-32.6 · Un sprint abierto no cuenta: sin sprints cerrados la
// velocidad no está disponible.
func TestProjectStoryPoints_SprintAbiertoNoCuentaParaLaVelocidad(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.item(t, projectID, "A", 13, backlog.StatusFinished, s.ID)

	got, err := e.metrics.ProjectStoryPoints(projectID)
	if err != nil {
		t.Fatalf("ProjectStoryPoints devolvió un error inesperado: %v", err)
	}
	if got.VelocityAvailable || got.Velocity != 0 || got.ClosedSprints != 0 {
		t.Errorf("velocidad = %v (disponible %v) con %d sprints, se esperaba no disponible", got.Velocity, got.VelocityAvailable, got.ClosedSprints)
	}
}

// CA-32.7 · Un proyecto inexistente, incluidos 0 y negativos, da
// ErrProjectNotFound.
func TestProjectStoryPoints_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	e := newMetricsEnv()

	for _, id := range []int{99, 0, -1} {
		if _, err := e.metrics.ProjectStoryPoints(id); !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("ProjectStoryPoints(%d): err = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
	}
}

// CA-32.7 · Consultar no modifica los ítems ni los sprints.
func TestStoryPoints_NoModificaNada(t *testing.T) {
	e := newMetricsEnv()
	projectID := e.project(t, "Sistema de ventas")
	e.closedSprint(t, projectID, "Sprint 1", 5)
	s := e.tec03Sprint(t, projectID)
	itemsBefore, _ := e.itemRepo.ListByProject(projectID)
	sprintsBefore := e.sprintsOf(projectID)

	if _, err := e.metrics.SprintStoryPoints(s.ID); err != nil {
		t.Fatalf("SprintStoryPoints devolvió un error inesperado: %v", err)
	}
	if _, err := e.metrics.ProjectStoryPoints(projectID); err != nil {
		t.Fatalf("ProjectStoryPoints devolvió un error inesperado: %v", err)
	}

	itemsAfter, _ := e.itemRepo.ListByProject(projectID)
	sprintsAfter := e.sprintsOf(projectID)
	if !reflect.DeepEqual(itemsAfter, itemsBefore) || !reflect.DeepEqual(sprintsAfter, sprintsBefore) {
		t.Errorf("la consulta modificó los datos:\nítems %+v\nantes %+v\nsprints %+v\nantes %+v", itemsAfter, itemsBefore, sprintsAfter, sprintsBefore)
	}
}

// failingSprintRepository falla al listar los sprints de un proyecto.
type failingSprintRepository struct {
	sprint.Repository
	err error
}

func (r failingSprintRepository) ListByProject(int) ([]sprint.Sprint, error) {
	return nil, r.err
}

// failingItemRepository falla al listar los ítems de un proyecto.
type failingItemRepository struct {
	backlog.Repository
	err error
}

func (r failingItemRepository) ListByProject(int) ([]backlog.Item, error) {
	return nil, r.err
}

// CA-32.7 · Si falla el repositorio de sprints, ProjectStoryPoints devuelve
// ese error.
func TestProjectStoryPoints_ErrorDelRepositorioDeSprintsSePropaga(t *testing.T) {
	errRepo := errors.New("falla de sprints")
	e := newMetricsEnvWith(backlog.NewMemoryRepository(), failingSprintRepository{Repository: sprint.NewMemoryRepository(), err: errRepo}, effort.NewMemoryRepository())
	projectID := e.project(t, "Sistema de ventas")

	if _, err := e.metrics.ProjectStoryPoints(projectID); !errors.Is(err, errRepo) {
		t.Errorf("err = %v, se esperaba %v", err, errRepo)
	}
}

// CA-32.7 · Si falla el repositorio de ítems, las dos consultas devuelven
// ese error.
func TestStoryPoints_ErrorDelRepositorioDeItemsSePropaga(t *testing.T) {
	errRepo := errors.New("falla de ítems")
	e := newMetricsEnvWith(failingItemRepository{Repository: backlog.NewMemoryRepository(), err: errRepo}, sprint.NewMemoryRepository(), effort.NewMemoryRepository())
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")

	if _, err := e.metrics.SprintStoryPoints(s.ID); !errors.Is(err, errRepo) {
		t.Errorf("SprintStoryPoints: err = %v, se esperaba %v", err, errRepo)
	}
	if _, err := e.metrics.ProjectStoryPoints(projectID); !errors.Is(err, errRepo) {
		t.Errorf("ProjectStoryPoints: err = %v, se esperaba %v", err, errRepo)
	}
}
