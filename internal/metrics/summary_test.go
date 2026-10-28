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

// HU-35 · Consultar las métricas del proyecto (Issue #53)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-35-consulta-metricas.md). Los repositorios en
// memoria se envuelven en repositorios espía que cuentan las lecturas y que
// pueden fallar a pedido, después de cargar los datos, para verificar la
// lectura única del backlog (RN-35.3) y el orden de lectura (RN-35.4).

// spyItemRepository cuenta las lecturas de los ítems de un proyecto y falla
// con err si está cargado.
type spyItemRepository struct {
	backlog.Repository
	err       error
	listCalls int
}

func (r *spyItemRepository) ListByProject(projectID int) ([]backlog.Item, error) {
	r.listCalls++
	if r.err != nil {
		return nil, r.err
	}
	return r.Repository.ListByProject(projectID)
}

// spySprintRepository falla al leer un sprint (getErr) o la lista de sprints
// de un proyecto (listErr) si están cargados.
type spySprintRepository struct {
	sprint.Repository
	getErr  error
	listErr error
}

func (r *spySprintRepository) GetByID(id int) (sprint.Sprint, error) {
	if r.getErr != nil {
		return sprint.Sprint{}, r.getErr
	}
	return r.Repository.GetByID(id)
}

func (r *spySprintRepository) ListByProject(projectID int) ([]sprint.Sprint, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.Repository.ListByProject(projectID)
}

// spyEffortRepository cuenta las lecturas de los esfuerzos de un sprint y
// falla con err si está cargado.
type spyEffortRepository struct {
	effort.Repository
	err       error
	listCalls int
}

func (r *spyEffortRepository) ListBySprint(sprintID int) ([]effort.Effort, error) {
	r.listCalls++
	if r.err != nil {
		return nil, r.err
	}
	return r.Repository.ListBySprint(sprintID)
}

// spyDefectRepository cuenta las lecturas de los defectos y falla con err si
// está cargado.
type spyDefectRepository struct {
	defect.Repository
	err       error
	listCalls int
}

func (r *spyDefectRepository) List() ([]defect.Defect, error) {
	r.listCalls++
	if r.err != nil {
		return nil, r.err
	}
	return r.Repository.List()
}

// summaryEnv es el entorno de métricas con los cuatro repositorios espía.
type summaryEnv struct {
	*metricsEnv
	itemSpy   *spyItemRepository
	sprintSpy *spySprintRepository
	effortSpy *spyEffortRepository
	defectSpy *spyDefectRepository
}

func newSummaryEnv() *summaryEnv {
	s := &summaryEnv{
		itemSpy:   &spyItemRepository{Repository: backlog.NewMemoryRepository()},
		sprintSpy: &spySprintRepository{Repository: sprint.NewMemoryRepository()},
		effortSpy: &spyEffortRepository{Repository: effort.NewMemoryRepository()},
		defectSpy: &spyDefectRepository{Repository: defect.NewMemoryRepository()},
	}
	s.metricsEnv = newMetricsEnvWithDefects(s.itemSpy, s.sprintSpy, s.effortSpy, s.defectSpy)
	return s
}

// resetReads pone en cero los contadores de lectura.
func (e *summaryEnv) resetReads() {
	e.itemSpy.listCalls = 0
	e.effortSpy.listCalls = 0
	e.defectSpy.listCalls = 0
}

// fullItem crea un ítem con story points (0 = sin estimar) y horas estimadas
// (0 = sin horas), lo asigna al sprint si sprintID no es 0, registra el
// esfuerzo indicado (0 = ninguno) y lo lleva al estado pedido.
func (e *metricsEnv) fullItem(t *testing.T, projectID int, title string, points int, hours float64, status backlog.Status, sprintID int, effortHours float64) backlog.Item {
	t.Helper()
	it := e.itemWithHours(t, projectID, title, hours, sprintID)
	var err error
	if points > 0 {
		if it, err = e.items.AssignStoryPoints(it.ID, points); err != nil {
			t.Fatalf("asignar %d story points a %q: %v", points, title, err)
		}
	}
	if sprintID != 0 {
		e.registerEffort(t, sprintID, it.ID, effortHours)
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

// tec03Full arma el ejemplo de docs/metricas.md en un sprint abierto, con
// Story Points, horas, esfuerzos y un defecto detectado y resuelto en el
// sprint: A (5 SP, 10 h, TERMINADO, 12 h), B (3, 6, TERMINADO, 5), C (2, 0,
// EN_CURSO, 3) y D (0, 4, PENDIENTE, sin esfuerzo).
func (e *metricsEnv) tec03Full(t *testing.T, projectID int) sprint.Sprint {
	t.Helper()
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.fullItem(t, projectID, "A", 5, 10, backlog.StatusFinished, s.ID, 12)
	e.fullItem(t, projectID, "B", 3, 6, backlog.StatusFinished, s.ID, 5)
	e.fullItem(t, projectID, "C", 2, 0, backlog.StatusInProgress, s.ID, 3)
	e.fullItem(t, projectID, "D", 0, 4, backlog.StatusPending, s.ID, 0)
	e.saveDefect(t, a.ID, s.ID, defect.StatusResolved, s.ID)
	return s
}

// multiSprintProject arma un proyecto con dos sprints cerrados, uno abierto,
// ítems sin sprint, esfuerzos en los tres sprints y defectos detectados y
// resueltos en sprints distintos. Devuelve el proyecto y sus sprints.
func (e *metricsEnv) multiSprintProject(t *testing.T) (int, []sprint.Sprint) {
	t.Helper()
	projectID := e.project(t, "Sistema de ventas")
	s1 := e.openSprint(t, projectID, "Sprint 1")
	a := e.fullItem(t, projectID, "A", 3, 4, backlog.StatusFinished, s1.ID, 5)
	e.fullItem(t, projectID, "B", 2, 0, backlog.StatusInProgress, s1.ID, 1.5)
	e.saveDefect(t, a.ID, s1.ID, defect.StatusOpen, 0)
	if _, err := e.sprints.Close(s1.ID); err != nil {
		t.Fatalf("cerrar el sprint 1: %v", err)
	}
	s2 := e.openSprint(t, projectID, "Sprint 2")
	c := e.fullItem(t, projectID, "C", 5, 6, backlog.StatusFinished, s2.ID, 4)
	e.saveDefect(t, c.ID, s1.ID, defect.StatusResolved, s2.ID)
	if _, err := e.sprints.Close(s2.ID); err != nil {
		t.Fatalf("cerrar el sprint 2: %v", err)
	}
	s3 := e.openSprint(t, projectID, "Sprint 3")
	e.fullItem(t, projectID, "D", 8, 10, backlog.StatusPending, s3.ID, 2)
	e.fullItem(t, projectID, "E", 0, 3, backlog.StatusPending, 0, 0)
	e.fullItem(t, projectID, "F", 1, 0, backlog.StatusPending, 0, 0)
	return projectID, []sprint.Sprint{s1, s2, s3}
}

func expectSprintMetrics(t *testing.T, got SprintMetrics, err error, want SprintMetrics) {
	t.Helper()
	if err != nil {
		t.Fatalf("Sprint devolvió un error inesperado: %v", err)
	}
	if got != want {
		t.Errorf("Sprint = %+v, se esperaba %+v", got, want)
	}
}

func expectProjectMetrics(t *testing.T, got ProjectMetrics, err error, want ProjectMetrics) {
	t.Helper()
	if err != nil {
		t.Fatalf("Project devolvió un error inesperado: %v", err)
	}
	if got != want {
		t.Errorf("Project = %+v, se esperaba %+v", got, want)
	}
}

// separateSprint arma el resultado esperado de un sprint con las cuatro
// operaciones separadas de HU-32, HU-34, HU-33 y HU-47.
func separateSprint(t *testing.T, e *metricsEnv, sprintID int) SprintMetrics {
	t.Helper()
	points, err := e.metrics.SprintStoryPoints(sprintID)
	if err != nil {
		t.Fatalf("SprintStoryPoints: %v", err)
	}
	completion, err := e.metrics.SprintCompletion(sprintID)
	if err != nil {
		t.Fatalf("SprintCompletion: %v", err)
	}
	hours, err := e.metrics.SprintHours(sprintID)
	if err != nil {
		t.Fatalf("SprintHours: %v", err)
	}
	defects, err := e.metrics.SprintDefects(sprintID)
	if err != nil {
		t.Fatalf("SprintDefects: %v", err)
	}
	return SprintMetrics{SprintID: sprintID, Points: points, Completion: completion, Hours: hours, Defects: defects}
}

// separateProject arma el resultado esperado de un proyecto con las cuatro
// operaciones separadas.
func separateProject(t *testing.T, e *metricsEnv, projectID int) ProjectMetrics {
	t.Helper()
	points, err := e.metrics.ProjectStoryPoints(projectID)
	if err != nil {
		t.Fatalf("ProjectStoryPoints: %v", err)
	}
	completion, err := e.metrics.ProjectCompletion(projectID)
	if err != nil {
		t.Fatalf("ProjectCompletion: %v", err)
	}
	hours, err := e.metrics.ProjectHours(projectID)
	if err != nil {
		t.Fatalf("ProjectHours: %v", err)
	}
	defects, err := e.metrics.ProjectDefects(projectID)
	if err != nil {
		t.Fatalf("ProjectDefects: %v", err)
	}
	return ProjectMetrics{ProjectID: projectID, Points: points, Completion: completion, Hours: hours, Defects: defects}
}

// CA-35.1 · Ejemplo de TEC-03 en un sprint: M1, M2 y M4 a M10 en una sola
// consulta.
func TestSprint_EjemploDeTEC03(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.tec03Full(t, projectID)

	got, err := e.metrics.Sprint(s.ID)

	expectSprintMetrics(t, got, err, SprintMetrics{
		SprintID:   s.ID,
		Points:     SprintPoints{SprintID: s.ID, Planned: 10, Completed: 8, Unestimated: 1},
		Completion: Completion{Items: 4, Finished: 2, Percent: 50, PercentAvailable: true},
		Hours:      Hours{Estimated: 20, Actual: 20, Unestimated: 1, DeviationAvailable: true},
		Defects:    Defects{Detected: 1, Resolved: 1},
	})
}

// CA-35.1 · El mismo proyecto con ese sprint cerrado: M1 a M10, con la
// velocidad.
func TestProject_EjemploDeTEC03ConElSprintCerrado(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.tec03Full(t, projectID)
	if _, err := e.sprints.Close(s.ID); err != nil {
		t.Fatalf("cerrar el sprint: %v", err)
	}

	got, err := e.metrics.Project(projectID)

	expectProjectMetrics(t, got, err, ProjectMetrics{
		ProjectID:  projectID,
		Points:     ProjectPoints{ProjectID: projectID, Planned: 10, Completed: 8, Unestimated: 1, Velocity: 8, VelocityAvailable: true, ClosedSprints: 1},
		Completion: Completion{Items: 4, Finished: 2, Percent: 50, PercentAvailable: true},
		Hours:      Hours{Estimated: 20, Actual: 20, Unestimated: 1, DeviationAvailable: true},
		Defects:    Defects{Detected: 1, Resolved: 1},
	})
}

// CA-35.5 · Cada parte de Sprint es igual a su operación separada.
func TestSprint_EsIgualALasOperacionesSeparadas(t *testing.T) {
	e := newSummaryEnv()
	_, sprints := e.multiSprintProject(t)

	for _, s := range sprints {
		want := separateSprint(t, e.metricsEnv, s.ID)

		got, err := e.metrics.Sprint(s.ID)

		expectSprintMetrics(t, got, err, want)
	}
}

// CA-35.5 · Cada parte de Project es igual a su operación separada.
func TestProject_EsIgualALasOperacionesSeparadas(t *testing.T) {
	e := newSummaryEnv()
	projectID, _ := e.multiSprintProject(t)
	want := separateProject(t, e.metricsEnv, projectID)

	got, err := e.metrics.Project(projectID)

	expectProjectMetrics(t, got, err, want)
}

// CA-35.2 · Un proyecto recién creado tiene las cuatro métricas no
// disponibles y las demás en 0, sin error.
func TestProject_RecienCreadoLasCuatroNoDisponibles(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")

	got, err := e.metrics.Project(projectID)

	expectProjectMetrics(t, got, err, ProjectMetrics{
		ProjectID: projectID,
		Points:    ProjectPoints{ProjectID: projectID},
	})
}

// CA-35.2 · Un sprint sin ítems, esfuerzos ni defectos tiene el porcentaje y
// la desviación no disponibles.
func TestSprint_SinItemsPorcentajeYDesviacionNoDisponibles(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")

	got, err := e.metrics.Sprint(s.ID)

	expectSprintMetrics(t, got, err, SprintMetrics{
		SprintID: s.ID,
		Points:   SprintPoints{SprintID: s.ID},
	})
}

// CA-35.2 · Un proyecto con ítems y sin sprints cerrados solo tiene la
// velocidad no disponible.
func TestProject_SinSprintsCerradosSoloLaVelocidadNoDisponible(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.fullItem(t, projectID, "A", 5, 10, backlog.StatusFinished, s.ID, 12)

	got, err := e.metrics.Project(projectID)

	deviation := 12.0 - 10.0
	expectProjectMetrics(t, got, err, ProjectMetrics{
		ProjectID:  projectID,
		Points:     ProjectPoints{ProjectID: projectID, Planned: 5, Completed: 5},
		Completion: Completion{Items: 1, Finished: 1, Percent: 100, PercentAvailable: true},
		Hours:      Hours{Estimated: 10, Actual: 12, Deviation: deviation, DeviationPercent: deviation / 10 * 100, DeviationAvailable: true},
	})
}

// CA-35.2 · Ítems estimados sin esfuerzos: solo la desviación no está
// disponible (M5 = 0).
func TestSprint_SinEsfuerzosSoloLaDesviacionNoDisponible(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.fullItem(t, projectID, "A", 5, 10, backlog.StatusPending, s.ID, 0)

	got, err := e.metrics.Sprint(s.ID)

	expectSprintMetrics(t, got, err, SprintMetrics{
		SprintID:   s.ID,
		Points:     SprintPoints{SprintID: s.ID, Planned: 5},
		Completion: Completion{Items: 1, PercentAvailable: true},
		Hours:      Hours{Estimated: 10},
	})
}

// CA-35.2 · Esfuerzos sin horas estimadas: solo la desviación no está
// disponible (M4 = 0).
func TestSprint_SinHorasEstimadasSoloLaDesviacionNoDisponible(t *testing.T) {
	e := newSummaryEnv()
	projectID := e.project(t, "Sistema de ventas")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.fullItem(t, projectID, "A", 5, 0, backlog.StatusInProgress, s.ID, 3)

	got, err := e.metrics.Sprint(s.ID)

	expectSprintMetrics(t, got, err, SprintMetrics{
		SprintID:   s.ID,
		Points:     SprintPoints{SprintID: s.ID, Planned: 5},
		Completion: Completion{Items: 1, PercentAvailable: true},
		Hours:      Hours{Actual: 3, Unestimated: 1},
	})
}

// CA-35.3 · Un sprint inexistente, 0 o negativo devuelve ErrSprintNotFound.
func TestSprint_InexistenteDevuelveErrSprintNotFound(t *testing.T) {
	e := newSummaryEnv()
	for _, id := range []int{99, 0, -1} {
		got, err := e.metrics.Sprint(id)

		if !errors.Is(err, sprint.ErrSprintNotFound) {
			t.Errorf("Sprint(%d): err = %v, se esperaba %v", id, err, sprint.ErrSprintNotFound)
		}
		if got != (SprintMetrics{}) {
			t.Errorf("Sprint(%d) = %+v, se esperaba el valor cero", id, got)
		}
	}
}

// CA-35.3 · Un proyecto inexistente, 0 o negativo devuelve
// ErrProjectNotFound.
func TestProject_InexistenteDevuelveErrProjectNotFound(t *testing.T) {
	e := newSummaryEnv()
	for _, id := range []int{99, 0, -1} {
		got, err := e.metrics.Project(id)

		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("Project(%d): err = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if got != (ProjectMetrics{}) {
			t.Errorf("Project(%d) = %+v, se esperaba el valor cero", id, got)
		}
	}
}

var (
	errSprintRead  = errors.New("falla al leer el sprint")
	errSprintsRead = errors.New("falla al leer los sprints")
	errItemsRead   = errors.New("falla al leer los ítems")
	errEffortsRead = errors.New("falla al leer los esfuerzos")
	errDefectsRead = errors.New("falla al leer los defectos")
)

// readOrderCase hace fallar dos lecturas consecutivas (o la última sola) y
// espera el error de la primera, sin leer los esfuerzos ni los defectos si
// la lectura que falla es anterior (RN-35.4).
type readOrderCase struct {
	name        string
	fail        func(e *summaryEnv)
	want        error
	effortsRead bool
	defectsRead bool
}

func checkReads(t *testing.T, e *summaryEnv, c readOrderCase) {
	t.Helper()
	if !c.effortsRead && e.effortSpy.listCalls != 0 {
		t.Errorf("%s: se leyeron los esfuerzos %d veces, se esperaba ninguna", c.name, e.effortSpy.listCalls)
	}
	if !c.defectsRead && e.defectSpy.listCalls != 0 {
		t.Errorf("%s: se leyeron los defectos %d veces, se esperaba ninguna", c.name, e.defectSpy.listCalls)
	}
}

// CA-35.3 y RN-35.4 · En el sprint se lee sprint → ítems → esfuerzos →
// defectos: se devuelve el error de la primera lectura que falla y no se
// hacen las siguientes.
func TestSprint_ErroresEnElOrdenDeLectura(t *testing.T) {
	cases := []readOrderCase{
		{name: "sprint e ítems", fail: func(e *summaryEnv) { e.sprintSpy.getErr, e.itemSpy.err = errSprintRead, errItemsRead }, want: errSprintRead},
		{name: "ítems y esfuerzos", fail: func(e *summaryEnv) { e.itemSpy.err, e.effortSpy.err = errItemsRead, errEffortsRead }, want: errItemsRead},
		{name: "esfuerzos y defectos", fail: func(e *summaryEnv) { e.effortSpy.err, e.defectSpy.err = errEffortsRead, errDefectsRead }, want: errEffortsRead, effortsRead: true},
		{name: "defectos", fail: func(e *summaryEnv) { e.defectSpy.err = errDefectsRead }, want: errDefectsRead, effortsRead: true, defectsRead: true},
	}
	for _, c := range cases {
		e := newSummaryEnv()
		projectID := e.project(t, "Sistema de ventas")
		s := e.tec03Full(t, projectID)
		c.fail(e)
		e.resetReads()

		got, err := e.metrics.Sprint(s.ID)

		if !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, se esperaba %v", c.name, err, c.want)
		}
		if got != (SprintMetrics{}) {
			t.Errorf("%s: Sprint = %+v, se esperaba el valor cero", c.name, got)
		}
		checkReads(t, e, c)
	}
}

// CA-35.3 y RN-35.4 · En el proyecto se lee ítems → sprints → esfuerzos →
// defectos: se devuelve el error de la primera lectura que falla y no se
// hacen las siguientes.
func TestProject_ErroresEnElOrdenDeLectura(t *testing.T) {
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

		got, err := e.metrics.Project(projectID)

		if !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, se esperaba %v", c.name, err, c.want)
		}
		if got != (ProjectMetrics{}) {
			t.Errorf("%s: Project = %+v, se esperaba el valor cero", c.name, got)
		}
		checkReads(t, e, c)
	}
}

// CA-35.4 · Sprint lee los ítems una sola vez: el repositorio de ítems
// recibe 2 lecturas, 1 de metrics y 1 interna de defect.Manager.ListByProject
// (límite conocido de la SDD).
func TestSprint_LeeLosItemsUnaSolaVez(t *testing.T) {
	e := newSummaryEnv()
	_, sprints := e.multiSprintProject(t)
	e.resetReads()

	if _, err := e.metrics.Sprint(sprints[1].ID); err != nil {
		t.Fatalf("Sprint devolvió un error inesperado: %v", err)
	}

	if e.itemSpy.listCalls != 2 {
		t.Errorf("lecturas de ítems = %d, se esperaba 2 (1 de metrics y 1 interna de defect)", e.itemSpy.listCalls)
	}
}

// CA-35.4 · Project lee los ítems una sola vez, con un proyecto de tres
// sprints: 2 lecturas del repositorio (1 de metrics y 1 interna de defect).
func TestProject_LeeLosItemsUnaSolaVez(t *testing.T) {
	e := newSummaryEnv()
	projectID, _ := e.multiSprintProject(t)
	e.resetReads()

	if _, err := e.metrics.Project(projectID); err != nil {
		t.Fatalf("Project devolvió un error inesperado: %v", err)
	}

	if e.itemSpy.listCalls != 2 {
		t.Errorf("lecturas de ítems = %d, se esperaba 2 (1 de metrics y 1 interna de defect)", e.itemSpy.listCalls)
	}
}

// CA-35.7 · Consultar no modifica nada, y modificar el resultado no cambia la
// próxima consulta.
func TestSummary_NoModificaNadaYDevuelveCopias(t *testing.T) {
	e := newSummaryEnv()
	projectID, sprints := e.multiSprintProject(t)
	itemsBefore, _ := e.itemRepo.ListByProject(projectID)
	sprintsBefore := e.sprintsOf(projectID)
	defectsBefore, _ := e.defectRepo.List()
	var effortsBefore []effort.Effort
	for _, s := range sprintsBefore {
		found, _ := e.effortRepo.ListBySprint(s.ID)
		effortsBefore = append(effortsBefore, found...)
	}

	first, err := e.metrics.Project(projectID)
	if err != nil {
		t.Fatalf("Project devolvió un error inesperado: %v", err)
	}
	firstSprint, err := e.metrics.Sprint(sprints[0].ID)
	if err != nil {
		t.Fatalf("Sprint devolvió un error inesperado: %v", err)
	}
	if first.ProjectID != projectID || firstSprint.SprintID != sprints[0].ID {
		t.Fatalf("IDs = %d y %d, se esperaba %d y %d", first.ProjectID, firstSprint.SprintID, projectID, sprints[0].ID)
	}
	changed := first
	changed.Points.Planned, changed.Hours.Estimated, changed.Defects.Detected = 999, -1, 999
	changedSprint := firstSprint
	changedSprint.Completion.Items = 999

	again, _ := e.metrics.Project(projectID)
	againSprint, _ := e.metrics.Sprint(sprints[0].ID)

	if again != first || againSprint != firstSprint {
		t.Errorf("la segunda consulta cambió: %+v y %+v, se esperaba %+v y %+v", again, againSprint, first, firstSprint)
	}
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
