package dashboard

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-36 · Ver el resumen del estado del proyecto (Issue #54)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-36-resumen-proyecto.md). Los datos se arman con
// los Manager reales de proyectos, backlog y sprints en memoria, así que los
// ítems y los sprints pasan por las mismas reglas que en la aplicación.

var (
	testStart = time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	testEnd   = time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
)

// dashboardEnv arma los Manager en memoria y el Manager del dashboard.
type dashboardEnv struct {
	projectRepo project.Repository
	itemRepo    backlog.Repository
	sprintRepo  sprint.Repository
	projects    *project.Manager
	items       *backlog.Manager
	sprints     *sprint.Manager
	dashboard   *Manager
}

func newDashboardEnv() *dashboardEnv {
	return newDashboardEnvWith(project.NewMemoryRepository(), backlog.NewMemoryRepository(), sprint.NewMemoryRepository())
}

// newDashboardEnvWith permite usar repositorios que fallan ("Condiciones de
// error" de la SDD).
func newDashboardEnvWith(projectRepo project.Repository, itemRepo backlog.Repository, sprintRepo sprint.Repository) *dashboardEnv {
	projects := project.NewManager(projectRepo)
	items := backlog.NewManager(itemRepo, projects)
	sprints := sprint.NewManager(sprintRepo, projects, items)
	return &dashboardEnv{
		projectRepo: projectRepo, itemRepo: itemRepo, sprintRepo: sprintRepo,
		projects: projects, items: items, sprints: sprints,
		dashboard: NewManager(projects, items, sprints),
	}
}

func (e *dashboardEnv) project(t *testing.T, name string) int {
	t.Helper()
	p, err := e.projects.Create(name, "")
	if err != nil {
		t.Fatalf("crear el proyecto %q: %v", name, err)
	}
	return p.ID
}

func (e *dashboardEnv) setProjectStatus(t *testing.T, projectID int, status project.Status) {
	t.Helper()
	p, err := e.projects.Get(projectID)
	if err != nil {
		t.Fatalf("leer el proyecto %d: %v", projectID, err)
	}
	if _, err := e.projects.Update(projectID, p.Name, p.Description, string(status)); err != nil {
		t.Fatalf("pasar el proyecto %d a %s: %v", projectID, status, err)
	}
}

func (e *dashboardEnv) openSprint(t *testing.T, projectID int, name string) sprint.Sprint {
	t.Helper()
	s, err := e.sprints.Create(projectID, name, testStart, testEnd, "Goal de "+name)
	if err != nil {
		t.Fatalf("crear el sprint %q: %v", name, err)
	}
	return s
}

func (e *dashboardEnv) closeSprint(t *testing.T, sprintID int) sprint.Sprint {
	t.Helper()
	s, err := e.sprints.Close(sprintID)
	if err != nil {
		t.Fatalf("cerrar el sprint %d: %v", sprintID, err)
	}
	return s
}

func (e *dashboardEnv) closedSprint(t *testing.T, projectID int, name string) sprint.Sprint {
	t.Helper()
	return e.closeSprint(t, e.openSprint(t, projectID, name).ID)
}

// item crea un ítem, lo asigna al sprint si sprintID no es 0 y lo lleva al
// estado pedido con HU-08.
func (e *dashboardEnv) item(t *testing.T, projectID int, title string, status backlog.Status, sprintID int) backlog.Item {
	t.Helper()
	it, err := e.items.Create(projectID, title, "", "MEDIA", []string{"Criterio de " + title})
	if err != nil {
		t.Fatalf("crear el ítem %q: %v", title, err)
	}
	if sprintID != 0 {
		if it, err = e.sprints.AssignItem(sprintID, it.ID); err != nil {
			t.Fatalf("asignar %q al sprint %d: %v", title, sprintID, err)
		}
	}
	return e.changeStatus(t, it, status)
}

// changeStatus lleva el ítem al estado pedido, un paso por vez (RN-08.2).
func (e *dashboardEnv) changeStatus(t *testing.T, it backlog.Item, status backlog.Status) backlog.Item {
	t.Helper()
	steps := map[backlog.Status]map[backlog.Status][]backlog.Status{
		backlog.StatusPending: {
			backlog.StatusInProgress: {backlog.StatusInProgress},
			backlog.StatusFinished:   {backlog.StatusInProgress, backlog.StatusFinished},
		},
		backlog.StatusInProgress: {
			backlog.StatusPending:  {backlog.StatusPending},
			backlog.StatusFinished: {backlog.StatusFinished},
		},
	}
	var err error
	for _, next := range steps[it.Status][status] {
		if it, err = e.items.ChangeStatus(it.ID, string(next)); err != nil {
			t.Fatalf("pasar %q a %s: %v", it.Title, next, err)
		}
	}
	return it
}

// snapshot lee el proyecto, los ítems y los sprints directamente de los
// repositorios en memoria, sin usar lo que se prueba (CA-36.5).
func (e *dashboardEnv) snapshot(t *testing.T, projectID int) (project.Project, []backlog.Item, []sprint.Sprint) {
	t.Helper()
	p, err := e.projectRepo.GetByID(projectID)
	if err != nil {
		t.Fatalf("leer el proyecto %d: %v", projectID, err)
	}
	items, err := e.itemRepo.ListByProject(projectID)
	if err != nil {
		t.Fatalf("leer los ítems del proyecto %d: %v", projectID, err)
	}
	sprints, err := e.sprintRepo.ListByProject(projectID)
	if err != nil {
		t.Fatalf("leer los sprints del proyecto %d: %v", projectID, err)
	}
	return p, items, sprints
}

func (e *dashboardEnv) summary(t *testing.T, projectID int) Summary {
	t.Helper()
	got, err := e.dashboard.ProjectSummary(projectID)
	if err != nil {
		t.Fatalf("ProjectSummary(%d) devolvió un error inesperado: %v", projectID, err)
	}
	return got
}

func expectCounts(t *testing.T, got Summary, pending, inProgress, finished int) {
	t.Helper()
	if got.Pending != pending || got.InProgress != inProgress || got.Finished != finished {
		t.Errorf("ítems = %d pendientes, %d en curso y %d terminados; se esperaba %d, %d y %d",
			got.Pending, got.InProgress, got.Finished, pending, inProgress, finished)
	}
}

func expectCurrentSprint(t *testing.T, got Summary, want sprint.Sprint) {
	t.Helper()
	if !got.HasCurrentSprint || !reflect.DeepEqual(got.CurrentSprint, want) {
		t.Errorf("sprint actual = %+v (hay: %v), se esperaba %+v", got.CurrentSprint, got.HasCurrentSprint, want)
	}
}

func expectNoCurrentSprint(t *testing.T, got Summary) {
	t.Helper()
	if got.HasCurrentSprint || got.CurrentSprint != (sprint.Sprint{}) {
		t.Errorf("sprint actual = %+v (hay: %v), se esperaba ninguno", got.CurrentSprint, got.HasCurrentSprint)
	}
}

// CA-36.1, RN-36.1 a RN-36.3 · El resumen muestra el nombre y el estado del
// proyecto, los ítems por estado y el sprint abierto.
func TestProjectSummary_MuestraProyectoItemsPorEstadoYSprintActual(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	e.setProjectStatus(t, p, project.StatusInProgress)
	s := e.openSprint(t, p, "Sprint 1")
	e.item(t, p, "A", backlog.StatusPending, s.ID)
	e.item(t, p, "B", backlog.StatusPending, s.ID)
	e.item(t, p, "C", backlog.StatusInProgress, s.ID)
	e.item(t, p, "D", backlog.StatusFinished, s.ID)
	e.item(t, p, "E", backlog.StatusFinished, s.ID)
	e.item(t, p, "F", backlog.StatusFinished, s.ID)

	got := e.summary(t, p)

	if got.ProjectID != p || got.ProjectName != "Sistema de ventas" || got.ProjectStatus != project.StatusInProgress {
		t.Errorf("proyecto = %d %q %s, se esperaba %d \"Sistema de ventas\" EN_CURSO", got.ProjectID, got.ProjectName, got.ProjectStatus, p)
	}
	expectCounts(t, got, 2, 1, 3)
	expectCurrentSprint(t, got, s)
}

// CA-36.1 y RN-36.2 · Se cuentan los ítems con y sin sprint asignado.
func TestProjectSummary_CuentaLosItemsConYSinSprint(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, p, "Sprint 1")
	e.item(t, p, "En el sprint", backlog.StatusFinished, s.ID)
	e.item(t, p, "Sin sprint pendiente", backlog.StatusPending, 0)
	e.item(t, p, "Sin sprint terminado", backlog.StatusFinished, 0)

	expectCounts(t, e.summary(t, p), 1, 0, 2)
}

// CA-36.1 y RN-36.2 · Los estados sin ítems se informan en 0.
func TestProjectSummary_ItemsEnUnSoloEstadoLosDemasEnCero(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	e.item(t, p, "A", backlog.StatusInProgress, 0)
	e.item(t, p, "B", backlog.StatusInProgress, 0)

	expectCounts(t, e.summary(t, p), 0, 2, 0)
}

// CA-36.2 y RN-36.4 · Un proyecto sin ítems ni sprints devuelve el resumen
// vacío, sin error.
func TestProjectSummary_ProyectoSinDatosDevuelveResumenVacio(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")

	got := e.summary(t, p)

	want := Summary{ProjectID: p, ProjectName: "Sistema de ventas", ProjectStatus: project.StatusPlanned}
	if got != want {
		t.Errorf("resumen = %+v, se esperaba %+v", got, want)
	}
}

// CA-36.3 y RN-36.3 · Con solo sprints cerrados no hay sprint actual.
func TestProjectSummary_SoloSprintsCerradosNoHaySprintActual(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	e.closedSprint(t, p, "Sprint 1")
	e.closedSprint(t, p, "Sprint 2")

	got := e.summary(t, p)

	if got.ProjectID != p || got.ProjectName != "Sistema de ventas" {
		t.Errorf("proyecto = %d %q, se esperaba %d \"Sistema de ventas\"", got.ProjectID, got.ProjectName, p)
	}
	expectNoCurrentSprint(t, got)
}

// CA-36.1 y RN-36.3 · El sprint actual es el abierto, aunque haya cerrados.
func TestProjectSummary_ElSprintAbiertoEsElActualAunqueHayaCerrados(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	e.closedSprint(t, p, "Sprint 1")
	s := e.openSprint(t, p, "Sprint 2")

	expectCurrentSprint(t, e.summary(t, p), s)
}

// CA-36.1 · No se cuentan ítems ni sprints de otro proyecto.
func TestProjectSummary_NoCuentaItemsNiSprintsDeOtroProyecto(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	other := e.project(t, "Inventario")
	s := e.openSprint(t, other, "Sprint de inventario")
	e.item(t, other, "X", backlog.StatusFinished, s.ID)
	e.item(t, other, "Y", backlog.StatusInProgress, 0)
	e.item(t, p, "A", backlog.StatusPending, 0)

	got := e.summary(t, p)

	expectCounts(t, got, 1, 0, 0)
	expectNoCurrentSprint(t, got)
}

// CA-36.4 y RN-36.5 · Después de cambiar el estado de un ítem, el resumen lo
// refleja.
func TestProjectSummary_ReflejaElCambioDeEstadoDeUnItem(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	a := e.item(t, p, "A", backlog.StatusInProgress, 0)
	e.item(t, p, "B", backlog.StatusPending, 0)
	expectCounts(t, e.summary(t, p), 1, 1, 0)

	e.changeStatus(t, a, backlog.StatusFinished)

	expectCounts(t, e.summary(t, p), 1, 0, 1)
}

// CA-36.4 y RN-36.5 · Después de cerrar el sprint actual, el resumen ya no
// tiene sprint actual.
func TestProjectSummary_ReflejaElCierreDelSprint(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, p, "Sprint 1")
	expectCurrentSprint(t, e.summary(t, p), s)

	e.closeSprint(t, s.ID)

	expectNoCurrentSprint(t, e.summary(t, p))
}

// CA-36.5 · Un proyecto inexistente, 0 o negativo devuelve
// project.ErrProjectNotFound y un resumen vacío.
func TestProjectSummary_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	e := newDashboardEnv()
	e.project(t, "Sistema de ventas")

	for _, id := range []int{99, 0, -1} {
		got, err := e.dashboard.ProjectSummary(id)

		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("ProjectSummary(%d): error = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if got != (Summary{}) {
			t.Errorf("ProjectSummary(%d) = %+v, se esperaba un resumen vacío", id, got)
		}
	}
}

// CA-36.5 y RN-36.6 · Consultar dos veces da el mismo resultado y no
// modifica el proyecto, los ítems ni los sprints.
func TestProjectSummary_ConsultarNoModificaNada(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, p, "Sprint 1")
	e.item(t, p, "A", backlog.StatusInProgress, s.ID)
	e.item(t, p, "B", backlog.StatusFinished, 0)
	projectBefore, itemsBefore, sprintsBefore := e.snapshot(t, p)

	first := e.summary(t, p)
	second := e.summary(t, p)

	if first != second {
		t.Errorf("dos consultas seguidas dieron %+v y %+v", first, second)
	}
	expectCounts(t, first, 0, 1, 1)
	projectAfter, itemsAfter, sprintsAfter := e.snapshot(t, p)
	if projectAfter != projectBefore || !reflect.DeepEqual(itemsAfter, itemsBefore) || !reflect.DeepEqual(sprintsAfter, sprintsBefore) {
		t.Error("consultar el resumen modificó el proyecto, los ítems o los sprints")
	}
}

// RN-36.7 · Modificar el resumen devuelto no altera el sprint guardado.
func TestProjectSummary_ModificarElResumenNoAlteraLoGuardado(t *testing.T) {
	e := newDashboardEnv()
	p := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, p, "Sprint 1")

	got := e.summary(t, p)
	got.CurrentSprint.Name = "Cambiado"
	got.CurrentSprint.Status = sprint.StatusClosed
	got.ProjectName = "Cambiado"

	again := e.summary(t, p)
	expectCurrentSprint(t, again, s)
	if again.ProjectName != "Sistema de ventas" {
		t.Errorf("nombre del proyecto = %q, se esperaba \"Sistema de ventas\"", again.ProjectName)
	}
}

// failingProjectRepository falla al leer un proyecto por ID.
type failingProjectRepository struct {
	project.Repository
	err error
}

func (r failingProjectRepository) GetByID(int) (project.Project, error) {
	return project.Project{}, r.err
}

// failingItemRepository falla al listar los ítems de un proyecto.
type failingItemRepository struct {
	backlog.Repository
	err error
}

func (r failingItemRepository) ListByProject(int) ([]backlog.Item, error) { return nil, r.err }

// failingSprintRepository falla al listar los sprints de un proyecto.
type failingSprintRepository struct {
	sprint.Repository
	err error
}

func (r failingSprintRepository) ListByProject(int) ([]sprint.Sprint, error) { return nil, r.err }

// Condiciones de error · Una falla al leer el proyecto, los ítems o los
// sprints se propaga y no se devuelve un resumen parcial.
func TestProjectSummary_ErrorDeLosRepositoriosSePropaga(t *testing.T) {
	errRepo := errors.New("base caída")
	cases := map[string]func() *dashboardEnv{
		"proyecto": func() *dashboardEnv {
			return newDashboardEnvWith(failingProjectRepository{Repository: project.NewMemoryRepository(), err: errRepo}, backlog.NewMemoryRepository(), sprint.NewMemoryRepository())
		},
		"ítems": func() *dashboardEnv {
			return newDashboardEnvWith(project.NewMemoryRepository(), failingItemRepository{Repository: backlog.NewMemoryRepository(), err: errRepo}, sprint.NewMemoryRepository())
		},
		"sprints": func() *dashboardEnv {
			return newDashboardEnvWith(project.NewMemoryRepository(), backlog.NewMemoryRepository(), failingSprintRepository{Repository: sprint.NewMemoryRepository(), err: errRepo})
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			e := build()
			p := e.project(t, "Sistema de ventas")

			got, err := e.dashboard.ProjectSummary(p)

			if !errors.Is(err, errRepo) {
				t.Errorf("error = %v, se esperaba %v", err, errRepo)
			}
			if got != (Summary{}) {
				t.Errorf("resumen = %+v, se esperaba vacío", got)
			}
		})
	}
}
