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

// HU-38 · Generar el reporte del proyecto (Issue #56)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-38-reporte-proyecto.md). Los datos se arman con
// los Manager reales en memoria, y sus repositorios se envuelven en
// repositorios espía que cuentan las lecturas y que fallan a pedido desde la
// lectura número failAt (contada desde resetReads). Así se puede hacer fallar
// la lectura del reporte después de la que ya hizo metrics.Manager.Project, y
// verificar el orden de lectura de RN-38.3.

var (
	testStart = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	testEnd   = time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

	errProjectRead = errors.New("falla la lectura del proyecto")
	errItemsRead   = errors.New("falla la lectura de los ítems")
	errSprintsRead = errors.New("falla la lectura de los sprints")
	errEffortsRead = errors.New("falla la lectura de los esfuerzos")
	errDefectsRead = errors.New("falla la lectura de los defectos")
)

// failing indica si la lectura número calls tiene que fallar: desde failAt
// en adelante, y nunca si failAt es 0.
func failing(calls, failAt int) bool {
	return failAt > 0 && calls >= failAt
}

// spyProjectRepository cuenta las lecturas de un proyecto por ID.
type spyProjectRepository struct {
	project.Repository
	calls, failAt int
}

func (r *spyProjectRepository) GetByID(id int) (project.Project, error) {
	r.calls++
	if failing(r.calls, r.failAt) {
		return project.Project{}, errProjectRead
	}
	return r.Repository.GetByID(id)
}

// spyItemRepository cuenta las lecturas de los ítems de un proyecto.
type spyItemRepository struct {
	backlog.Repository
	calls, failAt int
}

func (r *spyItemRepository) ListByProject(projectID int) ([]backlog.Item, error) {
	r.calls++
	if failing(r.calls, r.failAt) {
		return nil, errItemsRead
	}
	return r.Repository.ListByProject(projectID)
}

// spySprintRepository cuenta las lecturas de la lista de sprints de un
// proyecto.
type spySprintRepository struct {
	sprint.Repository
	calls, failAt int
}

func (r *spySprintRepository) ListByProject(projectID int) ([]sprint.Sprint, error) {
	r.calls++
	if failing(r.calls, r.failAt) {
		return nil, errSprintsRead
	}
	return r.Repository.ListByProject(projectID)
}

// spyEffortRepository cuenta las lecturas de los esfuerzos de un sprint.
type spyEffortRepository struct {
	effort.Repository
	calls, failAt int
}

func (r *spyEffortRepository) ListBySprint(sprintID int) ([]effort.Effort, error) {
	r.calls++
	if failing(r.calls, r.failAt) {
		return nil, errEffortsRead
	}
	return r.Repository.ListBySprint(sprintID)
}

// spyDefectRepository cuenta las lecturas de los defectos.
type spyDefectRepository struct {
	defect.Repository
	calls, failAt int
}

func (r *spyDefectRepository) List() ([]defect.Defect, error) {
	r.calls++
	if failing(r.calls, r.failAt) {
		return nil, errDefectsRead
	}
	return r.Repository.List()
}

// reportEnv arma los Manager en memoria, con los repositorios espía, y el
// Manager de reportes.
type reportEnv struct {
	projectSpy *spyProjectRepository
	itemSpy    *spyItemRepository
	sprintSpy  *spySprintRepository
	effortSpy  *spyEffortRepository
	defectSpy  *spyDefectRepository
	projects   *project.Manager
	items      *backlog.Manager
	sprints    *sprint.Manager
	efforts    *effort.Manager
	defects    *defect.Manager
	metrics    *metrics.Manager
	reports    *Manager
}

func newReportEnv() *reportEnv {
	e := &reportEnv{
		projectSpy: &spyProjectRepository{Repository: project.NewMemoryRepository()},
		itemSpy:    &spyItemRepository{Repository: backlog.NewMemoryRepository()},
		sprintSpy:  &spySprintRepository{Repository: sprint.NewMemoryRepository()},
		effortSpy:  &spyEffortRepository{Repository: effort.NewMemoryRepository()},
		defectSpy:  &spyDefectRepository{Repository: defect.NewMemoryRepository()},
	}
	e.projects = project.NewManager(e.projectSpy)
	e.items = backlog.NewManager(e.itemSpy, e.projects)
	e.sprints = sprint.NewManager(e.sprintSpy, e.projects, e.items)
	e.efforts = effort.NewManager(e.effortSpy, e.sprints, e.items)
	e.defects = defect.NewManager(e.defectSpy, e.items, e.sprints)
	e.metrics = metrics.NewManager(e.items, e.sprints, e.efforts, e.defects)
	e.reports = NewManager(e.projects, e.sprints, e.metrics, e.efforts, e.defects)
	return e
}

// resetReads pone en cero los contadores de lectura, para contar solo las del
// reporte.
func (e *reportEnv) resetReads() {
	e.projectSpy.calls = 0
	e.itemSpy.calls = 0
	e.sprintSpy.calls = 0
	e.effortSpy.calls = 0
	e.defectSpy.calls = 0
}

func (e *reportEnv) project(t *testing.T, name, description string) int {
	t.Helper()
	p, err := e.projects.Create(name, description)
	if err != nil {
		t.Fatalf("crear el proyecto %q: %v", name, err)
	}
	return p.ID
}

// startProject pasa el proyecto a EN_CURSO y le registra la fecha de inicio.
func (e *reportEnv) startProject(t *testing.T, projectID int, start time.Time) {
	t.Helper()
	p, err := e.projects.Get(projectID)
	if err != nil {
		t.Fatalf("leer el proyecto %d: %v", projectID, err)
	}
	if _, err := e.projects.Update(projectID, p.Name, p.Description, string(project.StatusInProgress)); err != nil {
		t.Fatalf("pasar el proyecto %d a EN_CURSO: %v", projectID, err)
	}
	if _, err := e.projects.SetStartDate(projectID, start); err != nil {
		t.Fatalf("registrar la fecha de inicio del proyecto %d: %v", projectID, err)
	}
}

func (e *reportEnv) openSprint(t *testing.T, projectID int, name string) sprint.Sprint {
	t.Helper()
	s, err := e.sprints.Create(projectID, name, testStart, testEnd, "Goal de "+name)
	if err != nil {
		t.Fatalf("crear el sprint %q: %v", name, err)
	}
	return s
}

func (e *reportEnv) closeSprint(t *testing.T, sprintID int) {
	t.Helper()
	if _, err := e.sprints.Close(sprintID); err != nil {
		t.Fatalf("cerrar el sprint %d: %v", sprintID, err)
	}
}

// item crea un ítem con story points (0 = sin estimar) y horas estimadas
// (0 = sin horas), lo asigna al sprint si sprintID no es 0 y lo lleva al
// estado pedido.
func (e *reportEnv) item(t *testing.T, projectID int, title string, points int, hours float64, status backlog.Status, sprintID int) backlog.Item {
	t.Helper()
	criteria := []string{"Criterio de " + title}
	it, err := e.items.Create(projectID, title, "", "MEDIA", criteria)
	if err != nil {
		t.Fatalf("crear el ítem %q: %v", title, err)
	}
	if it, err = e.items.Update(it.ID, title, "", "MEDIA", criteria, hours); err != nil {
		t.Fatalf("asignar %v horas a %q: %v", hours, title, err)
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

// effort registra horas de un integrante en un ítem de un sprint abierto
// (HU-26).
func (e *reportEnv) effort(t *testing.T, sprintID, itemID int, member string, hours float64) effort.Effort {
	t.Helper()
	saved, err := e.efforts.Register(sprintID, itemID, member, testStart.AddDate(0, 0, 5), "Desarrollo", hours)
	if err != nil {
		t.Fatalf("registrar %v horas de %q en el ítem %d: %v", hours, member, itemID, err)
	}
	return saved
}

// saveDefect guarda un defecto de la historia itemID, detectado en
// detectionSprintID, con ese estado y ese sprint de resolución (0 = ninguno),
// directo en el repositorio, como los tests de HU-47.
func (e *reportEnv) saveDefect(t *testing.T, itemID, detectionSprintID int, status defect.Status, resolutionSprintID int) defect.Defect {
	t.Helper()
	d, err := e.defectSpy.Repository.Save(defect.Defect{
		Description:        "Defecto",
		Severity:           defect.SeverityHigh,
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

// tec03Project arma el proyecto del ejemplo 1 de la SDD: el ejemplo de TEC-03
// en un sprint cerrado, con 12 h de Ana en A, 5 h de Luis en B, 3 h de Ana en
// C y un defecto de A detectado y resuelto en el sprint.
func (e *reportEnv) tec03Project(t *testing.T) int {
	t.Helper()
	projectID := e.project(t, "Software Metrics", "Estimación y seguimiento de proyectos")
	e.startProject(t, projectID, time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC))
	s := e.openSprint(t, projectID, "Sprint 1")
	a := e.item(t, projectID, "A", 5, 10, backlog.StatusFinished, s.ID)
	b := e.item(t, projectID, "B", 3, 6, backlog.StatusFinished, s.ID)
	c := e.item(t, projectID, "C", 2, 0, backlog.StatusInProgress, s.ID)
	e.item(t, projectID, "D", 0, 4, backlog.StatusPending, s.ID)
	e.effort(t, s.ID, a.ID, "Ana", 12)
	e.effort(t, s.ID, b.ID, "Luis", 5)
	e.effort(t, s.ID, c.ID, "Ana", 3)
	e.saveDefect(t, a.ID, s.ID, defect.StatusResolved, s.ID)
	e.closeSprint(t, s.ID)
	return projectID
}

// multiSprintProject arma un proyecto con dos sprints cerrados y uno abierto,
// esfuerzos de "Ana", "Luis" y "ana" en los tres, un ítem sin sprint, un
// defecto en cada estado y, después, otro proyecto con su propio sprint,
// esfuerzo y defecto. Devuelve el proyecto y sus tres sprints.
func (e *reportEnv) multiSprintProject(t *testing.T) (int, []sprint.Sprint) {
	t.Helper()
	projectID := e.project(t, "Sistema de ventas", "")
	s1 := e.openSprint(t, projectID, "Sprint 1")
	a := e.item(t, projectID, "A", 3, 4, backlog.StatusFinished, s1.ID)
	b := e.item(t, projectID, "B", 2, 0, backlog.StatusInProgress, s1.ID)
	e.effort(t, s1.ID, a.ID, "Ana", 5)
	e.effort(t, s1.ID, b.ID, "Luis", 1.5)
	e.saveDefect(t, a.ID, s1.ID, defect.StatusOpen, 0)
	e.closeSprint(t, s1.ID)
	s2 := e.openSprint(t, projectID, "Sprint 2")
	c := e.item(t, projectID, "C", 5, 6, backlog.StatusFinished, s2.ID)
	e.effort(t, s2.ID, c.ID, "Luis", 4)
	e.effort(t, s2.ID, c.ID, "Ana", 0.5)
	e.saveDefect(t, c.ID, s1.ID, defect.StatusResolved, s2.ID)
	e.saveDefect(t, c.ID, s2.ID, defect.StatusInCorrection, 0)
	e.closeSprint(t, s2.ID)
	s3 := e.openSprint(t, projectID, "Sprint 3")
	d := e.item(t, projectID, "D", 8, 10, backlog.StatusPending, s3.ID)
	e.effort(t, s3.ID, d.ID, "ana", 2)
	e.item(t, projectID, "E", 0, 3, backlog.StatusPending, 0)

	otherID := e.project(t, "Sistema de compras", "")
	x := e.openSprint(t, otherID, "Sprint X")
	z := e.item(t, otherID, "Z", 8, 20, backlog.StatusInProgress, x.ID)
	e.effort(t, x.ID, z.ID, "Pedro", 7)
	e.saveDefect(t, z.ID, x.ID, defect.StatusOpen, 0)
	return projectID, []sprint.Sprint{s1, s2, s3}
}

// effortKey son los campos de un esfuerzo que identifican su posición en el
// reporte.
type effortKey struct {
	ID, SprintID, ItemID int
	Member               string
	Hours                float64
}

func keysOf(efforts []effort.Effort) []effortKey {
	keys := make([]effortKey, 0, len(efforts))
	for _, ef := range efforts {
		keys = append(keys, effortKey{ef.ID, ef.SprintID, ef.ItemID, ef.Member, ef.Hours})
	}
	return keys
}

func defectIDs(defects []defect.Defect) []int {
	ids := make([]int, 0, len(defects))
	for _, d := range defects {
		ids = append(ids, d.ID)
	}
	return ids
}

func generate(t *testing.T, e *reportEnv, projectID int) ProjectReport {
	t.Helper()
	got, err := e.reports.Project(projectID)
	if err != nil {
		t.Fatalf("Project devolvió un error inesperado: %v", err)
	}
	return got
}

// CA-38.1 · El reporte del ejemplo de TEC-03 tiene los datos del proyecto, M1
// a M10, los tres esfuerzos con su total, el defecto y los dos integrantes.
func TestProject_EjemploDeTEC03(t *testing.T) {
	e := newReportEnv()
	projectID := e.tec03Project(t)

	got := generate(t, e, projectID)

	wantProject := project.Project{ID: projectID, Name: "Software Metrics", Description: "Estimación y seguimiento de proyectos",
		Status: project.StatusInProgress, StartDate: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)}
	if got.Project != wantProject {
		t.Errorf("Project = %+v, se esperaba %+v", got.Project, wantProject)
	}
	wantMetrics := metrics.ProjectMetrics{
		ProjectID:  projectID,
		Points:     metrics.ProjectPoints{ProjectID: projectID, Planned: 10, Completed: 8, Unestimated: 1, Velocity: 8, VelocityAvailable: true, ClosedSprints: 1},
		Completion: metrics.Completion{Items: 4, Finished: 2, Percent: 50, PercentAvailable: true},
		Hours:      metrics.Hours{Estimated: 20, Actual: 20, Unestimated: 1, DeviationAvailable: true},
		Defects:    metrics.Defects{Detected: 1, Resolved: 1},
	}
	if got.Metrics != wantMetrics {
		t.Errorf("Metrics = %+v, se esperaba %+v", got.Metrics, wantMetrics)
	}
	wantEfforts := []effortKey{{1, 1, 1, "Ana", 12}, {2, 1, 2, "Luis", 5}, {3, 1, 3, "Ana", 3}}
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

// CA-38.4 y RN-38.8 · Cada parte es igual a su operación de origen con los
// mismos datos, y el total de horas es igual a M5.
func TestProject_CadaParteEsIgualASuOperacionDeOrigen(t *testing.T) {
	e := newReportEnv()
	projectID, _ := e.multiSprintProject(t)
	wantProject, _ := e.projects.Get(projectID)
	wantMetrics, _ := e.metrics.Project(projectID)
	sprints, _ := e.sprints.ListByProject(projectID)
	var wantEfforts []effort.Effort
	for _, s := range sprints {
		summary, _ := e.efforts.ListBySprint(s.ID)
		wantEfforts = append(wantEfforts, summary.Efforts...)
	}
	wantDefects, _ := e.defects.ListByProject(projectID)

	got := generate(t, e, projectID)

	if got.Project != wantProject {
		t.Errorf("Project = %+v, se esperaba %+v", got.Project, wantProject)
	}
	if got.Metrics != wantMetrics {
		t.Errorf("Metrics = %+v, se esperaba %+v", got.Metrics, wantMetrics)
	}
	if !reflect.DeepEqual(got.Efforts.Efforts, wantEfforts) {
		t.Errorf("Efforts = %+v, se esperaba %+v", keysOf(got.Efforts.Efforts), keysOf(wantEfforts))
	}
	if !reflect.DeepEqual(got.Defects.Defects, wantDefects) {
		t.Errorf("Defects = %v, se esperaba %v", defectIDs(got.Defects.Defects), defectIDs(wantDefects))
	}
	if got.Efforts.TotalHours != wantMetrics.Hours.Actual {
		t.Errorf("TotalHours = %v, se esperaba %v (M5)", got.Efforts.TotalHours, wantMetrics.Hours.Actual)
	}
}

// CA-38.4 y RN-38.4 · Con horas con decimales, el total es la suma exacta,
// sin redondear, y es el mismo float64 que M5.
func TestProject_TotalDeHorasExactoEIgualAHorasReales(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	s1 := e.openSprint(t, projectID, "Sprint 1")
	a := e.item(t, projectID, "A", 0, 0.1, backlog.StatusPending, s1.ID)
	e.effort(t, s1.ID, a.ID, "Ana", 0.1)
	e.effort(t, s1.ID, a.ID, "Ana", 0.2)
	e.closeSprint(t, s1.ID)
	s2 := e.openSprint(t, projectID, "Sprint 2")
	b := e.item(t, projectID, "B", 0, 0.2, backlog.StatusPending, s2.ID)
	e.effort(t, s2.ID, b.ID, "Luis", 0.3)
	// Suma en float64 en el orden de RN-38.4: sprint por sprint, cada uno
	// desde 0 (una suma de constantes de Go sería exacta y daría 0.6).
	h1, h2, h3 := 0.1, 0.2, 0.3
	wantTotal := 0.0 + (0.0 + h1 + h2) + (0.0 + h3)

	got := generate(t, e, projectID)

	if got.Efforts.TotalHours != wantTotal {
		t.Errorf("TotalHours = %v, se esperaba la suma exacta %v", got.Efforts.TotalHours, wantTotal)
	}
	if got.Efforts.TotalHours != got.Metrics.Hours.Actual {
		t.Errorf("TotalHours = %v, se esperaba %v (M5)", got.Efforts.TotalHours, got.Metrics.Hours.Actual)
	}
}

// RN-38.4 · Los esfuerzos se ordenan por ID de sprint y, dentro de cada
// sprint, por ID de esfuerzo, aunque los IDs no sigan el orden de los sprints.
func TestProject_EsfuerzosOrdenadosPorSprintYPorID(t *testing.T) {
	e := newReportEnv()
	projectID, sprints := e.multiSprintProject(t)
	late, err := e.effortSpy.Repository.Save(effort.Effort{Member: "Luis", Date: testStart, Activity: "Corrección", ItemID: 1, SprintID: sprints[0].ID, Hours: 1})
	if err != nil {
		t.Fatalf("guardar el esfuerzo: %v", err)
	}

	got := generate(t, e, projectID)

	want := []effortKey{
		{1, sprints[0].ID, 1, "Ana", 5}, {2, sprints[0].ID, 2, "Luis", 1.5}, {late.ID, sprints[0].ID, 1, "Luis", 1},
		{3, sprints[1].ID, 3, "Luis", 4}, {4, sprints[1].ID, 3, "Ana", 0.5},
		{5, sprints[2].ID, 4, "ana", 2},
	}
	if keys := keysOf(got.Efforts.Efforts); !reflect.DeepEqual(keys, want) {
		t.Errorf("Efforts = %+v, se esperaba %+v", keys, want)
	}
}

// CA-38.5 y RN-38.6 · Cada integrante aparece una sola vez, en el orden de su
// primer esfuerzo; "Ana" y "ana" son dos integrantes (límite conocido).
func TestProject_IntegrantesSinRepetirEnOrdenDePrimeraAparicion(t *testing.T) {
	e := newReportEnv()
	projectID, _ := e.multiSprintProject(t)

	got := generate(t, e, projectID)

	if want := []string{"Ana", "Luis", "ana"}; !reflect.DeepEqual(got.Members.Names, want) {
		t.Errorf("Members = %q, se esperaba %q", got.Members.Names, want)
	}
}

// CA-38.1 y RN-38.5 · Los defectos del proyecto aparecen en los tres estados,
// ordenados por ID.
func TestProject_DefectosEnLosTresEstadosOrdenadosPorID(t *testing.T) {
	e := newReportEnv()
	projectID, _ := e.multiSprintProject(t)

	got := generate(t, e, projectID)

	wantStatuses := []defect.Status{defect.StatusOpen, defect.StatusResolved, defect.StatusInCorrection}
	var statuses []defect.Status
	for _, d := range got.Defects.Defects {
		statuses = append(statuses, d.Status)
	}
	if ids := defectIDs(got.Defects.Defects); !reflect.DeepEqual(ids, []int{1, 2, 3}) || !reflect.DeepEqual(statuses, wantStatuses) {
		t.Errorf("Defects = %v con estados %v, se esperaba [1 2 3] con %v", ids, statuses, wantStatuses)
	}
}

// CA-38.5 · Los esfuerzos, los defectos y los integrantes de otro proyecto no
// aparecen, en ninguno de los dos reportes.
func TestProject_NoIncluyeDatosDeOtroProyecto(t *testing.T) {
	e := newReportEnv()
	projectID, _ := e.multiSprintProject(t)
	otherID := projectID + 1

	got := generate(t, e, projectID)
	other := generate(t, e, otherID)

	for _, ef := range got.Efforts.Efforts {
		if ef.Member == "Pedro" {
			t.Errorf("el reporte incluye el esfuerzo %d de otro proyecto", ef.ID)
		}
	}
	for _, d := range got.Defects.Defects {
		if d.ID == 4 {
			t.Error("el reporte incluye el defecto 4 de otro proyecto")
		}
	}
	if want := []string{"Pedro"}; !reflect.DeepEqual(other.Members.Names, want) {
		t.Errorf("Members del otro proyecto = %q, se esperaba %q", other.Members.Names, want)
	}
	if ids := defectIDs(other.Defects.Defects); !reflect.DeepEqual(ids, []int{4}) || other.Efforts.TotalHours != 7 {
		t.Errorf("otro proyecto: defectos %v y total %v, se esperaba [4] y 7", ids, other.Efforts.TotalHours)
	}
}

// expectEmptyParts verifica que las listas del reporte estén vacías pero no
// sean nil, con total 0 (RN-38.7).
func expectEmptyParts(t *testing.T, got ProjectReport) {
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

// CA-38.2 y RN-38.7 · Un proyecto recién creado, sin fecha de inicio, genera
// el reporte con sus datos, las listas vacías y las cuatro métricas no
// disponibles.
func TestProject_ProyectoSinDatos(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Proyecto sin datos", "Recién creado")

	got := generate(t, e, projectID)

	want := project.Project{ID: projectID, Name: "Proyecto sin datos", Description: "Recién creado", Status: project.StatusPlanned}
	if got.Project != want {
		t.Errorf("Project = %+v, se esperaba %+v (sin fecha de inicio)", got.Project, want)
	}
	if got.Metrics.ProjectID != projectID || got.Metrics.Points.VelocityAvailable || got.Metrics.Completion.PercentAvailable || got.Metrics.Hours.DeviationAvailable {
		t.Errorf("Metrics = %+v, se esperaban las métricas del proyecto %d con las cuatro no disponibles", got.Metrics, projectID)
	}
	expectEmptyParts(t, got)
}

// CA-38.2 · Un proyecto con ítems y sin sprints tiene las métricas de sus
// ítems y las listas vacías.
func TestProject_ProyectoConItemsYSinSprints(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	e.item(t, projectID, "A", 3, 4, backlog.StatusPending, 0)
	e.item(t, projectID, "B", 2, 0, backlog.StatusFinished, 0)

	got := generate(t, e, projectID)

	if got.Metrics.Completion.Items != 2 || got.Metrics.Points.Planned != 5 || got.Metrics.Hours.Estimated != 4 {
		t.Errorf("Metrics = %+v, se esperaban 2 ítems, 5 SP y 4 h estimadas", got.Metrics)
	}
	expectEmptyParts(t, got)
}

// CA-38.2 · Con sprints sin esfuerzos, los esfuerzos y los integrantes están
// vacíos y el total es 0.
func TestProject_SprintsSinEsfuerzos(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	s1 := e.openSprint(t, projectID, "Sprint 1")
	e.item(t, projectID, "A", 3, 4, backlog.StatusFinished, s1.ID)
	e.closeSprint(t, s1.ID)
	e.openSprint(t, projectID, "Sprint 2")

	got := generate(t, e, projectID)

	if got.Metrics.Points.ClosedSprints != 1 || got.Metrics.Points.Completed != 3 {
		t.Errorf("Metrics = %+v, se esperaba 1 sprint cerrado con 3 SP completados", got.Metrics)
	}
	expectEmptyParts(t, got)
}

// CA-38.3 y RN-38.3 · Un proyecto inexistente (incluidos 0 y negativos)
// devuelve ErrProjectNotFound sin leer nada más que el proyecto.
func TestProject_ProyectoInexistenteNoLeeNadaMas(t *testing.T) {
	for _, id := range []int{99, 0, -1} {
		e := newReportEnv()
		e.multiSprintProject(t)
		e.resetReads()

		got, err := e.reports.Project(id)

		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("id %d: err = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if !reflect.DeepEqual(got, ProjectReport{}) {
			t.Errorf("id %d: Project = %+v, se esperaba el valor cero", id, got)
		}
		if e.projectSpy.calls != 1 || e.itemSpy.calls != 0 || e.sprintSpy.calls != 0 || e.effortSpy.calls != 0 || e.defectSpy.calls != 0 {
			t.Errorf("id %d: lecturas de proyecto, ítems, sprints, esfuerzos y defectos = %d, %d, %d, %d y %d; se esperaba 1, 0, 0, 0 y 0",
				id, e.projectSpy.calls, e.itemSpy.calls, e.sprintSpy.calls, e.effortSpy.calls, e.defectSpy.calls)
		}
	}
}

// reportReadCase hace fallar una lectura del reporte (RN-38.3) y espera su
// error y las lecturas que ya se hicieron, sin las siguientes. sprints es la
// cantidad de sprints del proyecto (N).
type reportReadCase struct {
	name                          string
	fail                          func(e *reportEnv, sprints int)
	want                          error
	sprintLists, efforts, defects func(sprints int) int
}

func reads(n int) func(int) int { return func(int) int { return n } }

// CA-38.3 y RN-38.3 · Se lee proyecto → métricas → sprints → esfuerzos →
// defectos: se devuelve el error de la lectura que falla, sin reporte parcial
// y sin las lecturas siguientes. Las lecturas de sprints, esfuerzos y
// defectos del reporte son las que siguen a las de metrics.Manager.Project.
func TestProject_ErroresEnElOrdenDeLectura(t *testing.T) {
	cases := []reportReadCase{
		{name: "proyecto", fail: func(e *reportEnv, _ int) { e.projectSpy.failAt = 1 }, want: errProjectRead,
			sprintLists: reads(0), efforts: reads(0), defects: reads(0)},
		{name: "métricas", fail: func(e *reportEnv, _ int) { e.itemSpy.failAt = 1 }, want: errItemsRead,
			sprintLists: reads(0), efforts: reads(0), defects: reads(0)},
		{name: "sprints", fail: func(e *reportEnv, _ int) { e.sprintSpy.failAt = 2 }, want: errSprintsRead,
			sprintLists: reads(2), efforts: func(n int) int { return n }, defects: reads(1)},
		{name: "esfuerzos", fail: func(e *reportEnv, n int) { e.effortSpy.failAt = n + 1 }, want: errEffortsRead,
			sprintLists: reads(2), efforts: func(n int) int { return n + 1 }, defects: reads(1)},
		{name: "defectos", fail: func(e *reportEnv, _ int) { e.defectSpy.failAt = 2 }, want: errDefectsRead,
			sprintLists: reads(2), efforts: func(n int) int { return 2 * n }, defects: reads(2)},
	}
	for _, c := range cases {
		e := newReportEnv()
		projectID, sprints := e.multiSprintProject(t)
		n := len(sprints)
		e.resetReads()
		c.fail(e, n)

		got, err := e.reports.Project(projectID)

		if !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, se esperaba %v", c.name, err, c.want)
		}
		if !reflect.DeepEqual(got, ProjectReport{}) {
			t.Errorf("%s: Project = %+v, se esperaba el valor cero", c.name, got)
		}
		if e.sprintSpy.calls != c.sprintLists(n) || e.effortSpy.calls != c.efforts(n) || e.defectSpy.calls != c.defects(n) {
			t.Errorf("%s: lecturas de sprints, esfuerzos y defectos = %d, %d y %d; se esperaba %d, %d y %d",
				c.name, e.sprintSpy.calls, e.effortSpy.calls, e.defectSpy.calls, c.sprintLists(n), c.efforts(n), c.defects(n))
		}
	}
}

// CA-38.6 y RN-38.11 · Generar el reporte dos veces sin cambios da lo mismo,
// modificar el reporte devuelto no cambia el siguiente y no se modifica nada
// de lo guardado.
func TestProject_NoModificaNadaYDevuelveCopias(t *testing.T) {
	e := newReportEnv()
	projectID, sprints := e.multiSprintProject(t)
	itemsBefore, _ := e.itemSpy.Repository.ListByProject(projectID)
	sprintsBefore, _ := e.sprintSpy.Repository.ListByProject(projectID)
	defectsBefore, _ := e.defectSpy.Repository.List()
	var effortsBefore []effort.Effort
	for _, s := range sprints {
		found, _ := e.effortSpy.Repository.ListBySprint(s.ID)
		effortsBefore = append(effortsBefore, found...)
	}

	first := generate(t, e, projectID)
	if len(first.Efforts.Efforts) == 0 || len(first.Defects.Defects) == 0 || len(first.Members.Names) == 0 {
		t.Fatalf("el reporte = %+v, se esperaban esfuerzos, defectos e integrantes", first)
	}
	changed := generate(t, e, projectID)
	if !reflect.DeepEqual(changed, first) {
		t.Errorf("la segunda consulta sin cambios = %+v, se esperaba %+v", changed, first)
	}
	changed.Efforts.Efforts[0].Hours = 999
	changed.Defects.Defects[0].Status = "OTRO"
	changed.Members.Names[0] = "Otro"

	again := generate(t, e, projectID)

	if !reflect.DeepEqual(again, first) {
		t.Errorf("modificar el reporte cambió el siguiente: %+v, se esperaba %+v", again, first)
	}
	itemsAfter, _ := e.itemSpy.Repository.ListByProject(projectID)
	sprintsAfter, _ := e.sprintSpy.Repository.ListByProject(projectID)
	defectsAfter, _ := e.defectSpy.Repository.List()
	var effortsAfter []effort.Effort
	for _, s := range sprints {
		found, _ := e.effortSpy.Repository.ListBySprint(s.ID)
		effortsAfter = append(effortsAfter, found...)
	}
	if !reflect.DeepEqual(itemsAfter, itemsBefore) || !reflect.DeepEqual(sprintsAfter, sprintsBefore) ||
		!reflect.DeepEqual(effortsAfter, effortsBefore) || !reflect.DeepEqual(defectsAfter, defectsBefore) {
		t.Error("el reporte modificó los ítems, los sprints, los esfuerzos o los defectos")
	}
}
