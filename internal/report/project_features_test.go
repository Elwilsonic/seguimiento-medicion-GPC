package report

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-38-reporte-proyecto.feature con
// godog.
func TestFeatures_HU38ReporteProyecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeReportScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-38-reporte-proyecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-38")
	}
}

// storedData son los datos guardados del proyecto, leídos directo de los
// repositorios, para verificar que el reporte no modifica nada.
type storedData struct {
	items   []backlog.Item
	sprints []sprint.Sprint
	efforts []effort.Effort
	defects []defect.Defect
}

// reportScenario guarda el estado de un escenario de HU-38.
type reportScenario struct {
	t        *testing.T
	env      *reportEnv
	projects map[string]int
	sprints  map[string]sprint.Sprint
	items    map[string]backlog.Item

	// failSource es la lectura que tiene que fallar ("" = ninguna).
	failSource string
	sprintsN   int

	projectID int
	result    ProjectReport
	err       error
	before    storedData
}

func initializeReportScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &reportScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*s = reportScenario{
			t:        t,
			env:      newReportEnv(),
			projects: map[string]int{},
			sprints:  map[string]sprint.Sprint{},
			items:    map[string]backlog.Item{},
		}
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el proyecto "([^"]*)" está EN_CURSO con fecha de inicio "([^"]*)"$`, s.startedProject)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que cierro el sprint "([^"]*)"$`, s.closeSprint)
	ctx.Step(`^que el sprint "([^"]*)" tiene los ítems:$`, s.sprintItemsTable)
	ctx.Step(`^que el proyecto "([^"]*)" tiene los ítems sin sprint:$`, s.projectItemsTable)
	ctx.Step(`^que se registraron los esfuerzos:$`, s.effortsTable)
	ctx.Step(`^que hay un defecto (ABIERTO|EN_CORRECCION) del ítem "([^"]*)" detectado en el sprint "([^"]*)"$`, s.unresolvedDefect)
	ctx.Step(`^que hay un defecto RESUELTO del ítem "([^"]*)" detectado y resuelto en el sprint "([^"]*)"$`, s.resolvedDefect)
	ctx.Step(`^que falla la lectura de (el proyecto|las métricas|los sprints|los esfuerzos|los defectos) del reporte$`, s.failRead)
	ctx.Step(`^genero el reporte del proyecto "([^"]*)"$`, s.generateByName)
	ctx.Step(`^genero el reporte del proyecto con el ID (-?\d+)$`, s.generate)
	ctx.Step(`^el proyecto del reporte es "([^"]*)" en estado "([^"]*)" con fecha de inicio "([^"]*)"$`, s.projectWithDateIs)
	ctx.Step(`^el proyecto del reporte es "([^"]*)" en estado "([^"]*)" sin fecha de inicio$`, s.projectWithoutDateIs)
	ctx.Step(`^los story points planificados son (\d+) y los completados son (\d+)$`, s.storyPointsAre)
	ctx.Step(`^la velocidad es ([\d.]+)$`, s.velocityIs)
	ctx.Step(`^los ítems totales son (\d+) y los terminados son (\d+)$`, s.itemCountsAre)
	ctx.Step(`^las horas estimadas son ([\d.]+) y las reales son ([\d.]+)$`, s.hoursAre)
	ctx.Step(`^los defectos detectados son (\d+) y los resueltos son (\d+)$`, s.defectsAre)
	ctx.Step(`^las métricas no disponibles son: (.+)$`, s.notAvailableAre)
	ctx.Step(`^los esfuerzos del reporte son:$`, s.effortsAre)
	ctx.Step(`^el total de horas es ([\d.]+)$`, s.totalIs)
	ctx.Step(`^el total de horas es ([\d.]+) al redondear a un decimal$`, s.roundedTotalIs)
	ctx.Step(`^el total de horas es igual a las horas reales de las métricas$`, s.totalEqualsActual)
	ctx.Step(`^los defectos del reporte son:$`, s.defectsListIs)
	ctx.Step(`^los integrantes son "([^"]*)"$`, s.membersAre)
	ctx.Step(`^el reporte no tiene esfuerzos, defectos ni integrantes$`, s.emptyParts)
	ctx.Step(`^cada parte del reporte es igual a su consulta de origen$`, s.equalToSources)
	ctx.Step(`^el reporte falla porque el proyecto no existe$`, s.failsNotFound)
	ctx.Step(`^solo se leyó el proyecto$`, s.onlyProjectRead)
	ctx.Step(`^el reporte falla con el error de la lectura de (el proyecto|las métricas|los sprints|los esfuerzos|los defectos)$`, s.failsWithReadError)
	ctx.Step(`^no se hicieron las lecturas siguientes a (el proyecto|las métricas|los sprints|los esfuerzos|los defectos)$`, s.laterReadsSkipped)
	ctx.Step(`^generar el reporte de nuevo da el mismo resultado$`, s.sameAgain)
	ctx.Step(`^modificar el reporte no cambia el próximo reporte$`, s.resultIsACopy)
	ctx.Step(`^los datos guardados no cambiaron$`, s.nothingChanged)
}

func (s *reportScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name, "")
	if s.projectID == 0 {
		s.projectID = s.projects[name]
	}
	return nil
}

func (s *reportScenario) startedProject(name, date string) error {
	start, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return err
	}
	s.env.startProject(s.t, s.projects[name], start)
	return nil
}

func (s *reportScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	return nil
}

func (s *reportScenario) closeSprint(name string) error {
	s.env.closeSprint(s.t, s.sprints[name].ID)
	return nil
}

// itemsTable crea los ítems de una tabla con titulo, sp, horas y estado,
// asignados a sprintID si no es 0.
func (s *reportScenario) itemsTable(projectID, sprintID int, table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		points, err := strconv.Atoi(row.Cells[1].Value)
		if err != nil {
			return err
		}
		hours, err := strconv.ParseFloat(row.Cells[2].Value, 64)
		if err != nil {
			return err
		}
		title := row.Cells[0].Value
		s.items[title] = s.env.item(s.t, projectID, title, points, hours, backlog.Status(row.Cells[3].Value), sprintID)
	}
	return nil
}

func (s *reportScenario) sprintItemsTable(name string, table *godog.Table) error {
	sp := s.sprints[name]
	return s.itemsTable(sp.ProjectID, sp.ID, table)
}

func (s *reportScenario) projectItemsTable(name string, table *godog.Table) error {
	return s.itemsTable(s.projects[name], 0, table)
}

// effortsTable registra los esfuerzos de una tabla con sprint, item,
// integrante y horas.
func (s *reportScenario) effortsTable(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		hours, err := strconv.ParseFloat(row.Cells[3].Value, 64)
		if err != nil {
			return err
		}
		s.env.effort(s.t, s.sprints[row.Cells[0].Value].ID, s.items[row.Cells[1].Value].ID, row.Cells[2].Value, hours)
	}
	return nil
}

func (s *reportScenario) unresolvedDefect(status, item, sprintName string) error {
	s.env.saveDefect(s.t, s.items[item].ID, s.sprints[sprintName].ID, defect.Status(status), 0)
	return nil
}

func (s *reportScenario) resolvedDefect(item, sprintName string) error {
	id := s.sprints[sprintName].ID
	s.env.saveDefect(s.t, s.items[item].ID, id, defect.StatusResolved, id)
	return nil
}

func (s *reportScenario) failRead(source string) error {
	s.failSource = source
	return nil
}

// applyFailure hace fallar la lectura del reporte de esa fuente (RN-38.3):
// las de sprints, esfuerzos y defectos fallan después de las que ya hizo
// metrics.Manager.Project.
func (s *reportScenario) applyFailure() {
	switch s.failSource {
	case "el proyecto":
		s.env.projectSpy.failAt = 1
	case "las métricas":
		s.env.itemSpy.failAt = 1
	case "los sprints":
		s.env.sprintSpy.failAt = 2
	case "los esfuerzos":
		s.env.effortSpy.failAt = s.sprintsN + 1
	case "los defectos":
		s.env.defectSpy.failAt = 2
	}
}

// stored lee los datos del proyecto directo de los repositorios, sin pasar
// por los espías.
func (s *reportScenario) stored() storedData {
	var data storedData
	data.items, _ = s.env.itemSpy.Repository.ListByProject(s.projectID)
	data.sprints, _ = s.env.sprintSpy.Repository.ListByProject(s.projectID)
	for _, sp := range data.sprints {
		found, _ := s.env.effortSpy.Repository.ListBySprint(sp.ID)
		data.efforts = append(data.efforts, found...)
	}
	data.defects, _ = s.env.defectSpy.Repository.List()
	return data
}

func (s *reportScenario) generateByName(name string) error {
	return s.generate(strconv.Itoa(s.projects[name]))
}

func (s *reportScenario) generate(idText string) error {
	id, err := strconv.Atoi(idText)
	if err != nil {
		return err
	}
	s.before = s.stored()
	s.sprintsN = len(s.before.sprints)
	s.env.resetReads()
	s.applyFailure()
	s.result, s.err = s.env.reports.Project(id)
	return nil
}

func (s *reportScenario) succeeded() error {
	if s.err != nil {
		return fmt.Errorf("el reporte falló: %v", s.err)
	}
	return nil
}

func (s *reportScenario) projectIs(name, status string, start time.Time) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	p := s.result.Project
	if p.ID != s.projectID || p.Name != name || string(p.Status) != status || !p.StartDate.Equal(start) {
		return fmt.Errorf("proyecto = %+v, se esperaba %q en %s con fecha de inicio %v", p, name, status, start)
	}
	return nil
}

func (s *reportScenario) projectWithDateIs(name, status, date string) error {
	start, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return err
	}
	return s.projectIs(name, status, start)
}

func (s *reportScenario) projectWithoutDateIs(name, status string) error {
	return s.projectIs(name, status, time.Time{})
}

func (s *reportScenario) storyPointsAre(planned, completed int) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if p := s.result.Metrics.Points; p.Planned != planned || p.Completed != completed {
		return fmt.Errorf("story points = %d y %d, se esperaba %d y %d", p.Planned, p.Completed, planned, completed)
	}
	return nil
}

func (s *reportScenario) velocityIs(velocity float64) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if p := s.result.Metrics.Points; !p.VelocityAvailable || p.Velocity != velocity {
		return fmt.Errorf("velocidad = %v (disponible: %v), se esperaba %v", p.Velocity, p.VelocityAvailable, velocity)
	}
	return nil
}

func (s *reportScenario) itemCountsAre(total, finished int) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if c := s.result.Metrics.Completion; c.Items != total || c.Finished != finished {
		return fmt.Errorf("ítems = %d y %d terminados, se esperaba %d y %d", c.Items, c.Finished, total, finished)
	}
	return nil
}

func (s *reportScenario) hoursAre(estimated, actual float64) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if h := s.result.Metrics.Hours; h.Estimated != estimated || h.Actual != actual {
		return fmt.Errorf("horas = %v estimadas y %v reales, se esperaba %v y %v", h.Estimated, h.Actual, estimated, actual)
	}
	return nil
}

func (s *reportScenario) defectsAre(detected, resolved int) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if d := s.result.Metrics.Defects; d.Detected != detected || d.Resolved != resolved {
		return fmt.Errorf("defectos = %d detectados y %d resueltos, se esperaba %d y %d", d.Detected, d.Resolved, detected, resolved)
	}
	return nil
}

// notAvailableAre verifica exactamente qué métricas no están disponibles:
// velocidad, porcentaje y desviación (RN-35.6).
func (s *reportScenario) notAvailableAre(list string) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	m := s.result.Metrics
	var got []string
	if !m.Points.VelocityAvailable {
		got = append(got, "velocidad")
	}
	if !m.Completion.PercentAvailable {
		got = append(got, "porcentaje")
	}
	if !m.Hours.DeviationAvailable {
		got = append(got, "desviación")
	}
	if want := strings.Split(list, ", "); !slices.Equal(got, want) {
		return fmt.Errorf("métricas no disponibles = %v, se esperaba %v", got, want)
	}
	return nil
}

// effortsAre compara los esfuerzos del reporte, en orden, con la tabla de
// sprint, item, integrante y horas.
func (s *reportScenario) effortsAre(table *godog.Table) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	var want, got []string
	for _, row := range table.Rows[1:] {
		hours, err := strconv.ParseFloat(row.Cells[3].Value, 64)
		if err != nil {
			return err
		}
		want = append(want, fmt.Sprintf("%d/%d/%s/%v", s.sprints[row.Cells[0].Value].ID, s.items[row.Cells[1].Value].ID, row.Cells[2].Value, hours))
	}
	for _, ef := range s.result.Efforts.Efforts {
		got = append(got, fmt.Sprintf("%d/%d/%s/%v", ef.SprintID, ef.ItemID, ef.Member, ef.Hours))
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("esfuerzos (sprint/ítem/integrante/horas) = %v, se esperaba %v", got, want)
	}
	return nil
}

func (s *reportScenario) totalIs(total float64) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if s.result.Efforts.TotalHours != total {
		return fmt.Errorf("total de horas = %v, se esperaba %v", s.result.Efforts.TotalHours, total)
	}
	return nil
}

func (s *reportScenario) roundedTotalIs(total float64) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if rounded := math.Round(s.result.Efforts.TotalHours*10) / 10; rounded != total {
		return fmt.Errorf("total de horas redondeado = %v (exacto %v), se esperaba %v", rounded, s.result.Efforts.TotalHours, total)
	}
	return nil
}

func (s *reportScenario) totalEqualsActual() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if s.result.Efforts.TotalHours != s.result.Metrics.Hours.Actual {
		return fmt.Errorf("total de horas = %v, se esperaba %v (horas reales)", s.result.Efforts.TotalHours, s.result.Metrics.Hours.Actual)
	}
	return nil
}

// defectsListIs compara los defectos del reporte, en orden, con la tabla de
// item y estado.
func (s *reportScenario) defectsListIs(table *godog.Table) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	var want, got []string
	for _, row := range table.Rows[1:] {
		want = append(want, fmt.Sprintf("%d/%s", s.items[row.Cells[0].Value].ID, row.Cells[1].Value))
	}
	for _, d := range s.result.Defects.Defects {
		got = append(got, fmt.Sprintf("%d/%s", d.ItemID, d.Status))
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("defectos (ítem/estado) = %v, se esperaba %v", got, want)
	}
	return nil
}

func (s *reportScenario) membersAre(list string) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if want := strings.Split(list, ", "); !slices.Equal(s.result.Members.Names, want) {
		return fmt.Errorf("integrantes = %q, se esperaba %q", s.result.Members.Names, want)
	}
	return nil
}

// emptyParts verifica las listas vacías, no nil, y el total 0 (RN-38.7).
func (s *reportScenario) emptyParts() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	r := s.result
	if r.Efforts.Efforts == nil || len(r.Efforts.Efforts) != 0 || r.Efforts.TotalHours != 0 {
		return fmt.Errorf("esfuerzos = %#v con total %v, se esperaba una lista vacía y total 0", r.Efforts.Efforts, r.Efforts.TotalHours)
	}
	if r.Defects.Defects == nil || len(r.Defects.Defects) != 0 {
		return fmt.Errorf("defectos = %#v, se esperaba una lista vacía", r.Defects.Defects)
	}
	if r.Members.Names == nil || len(r.Members.Names) != 0 {
		return fmt.Errorf("integrantes = %#v, se esperaba una lista vacía", r.Members.Names)
	}
	return nil
}

// equalToSources compara cada parte con su operación de origen (RN-38.8).
func (s *reportScenario) equalToSources() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	e := s.env
	wantProject, _ := e.projects.Get(s.projectID)
	wantMetrics, _ := e.metrics.Project(s.projectID)
	sprints, _ := e.sprints.ListByProject(s.projectID)
	var wantEfforts []effort.Effort
	for _, sp := range sprints {
		summary, _ := e.efforts.ListBySprint(sp.ID)
		wantEfforts = append(wantEfforts, summary.Efforts...)
	}
	wantDefects, _ := e.defects.ListByProject(s.projectID)
	r := s.result
	if r.Project != wantProject || r.Metrics != wantMetrics || !reflect.DeepEqual(r.Efforts.Efforts, wantEfforts) || !reflect.DeepEqual(r.Defects.Defects, wantDefects) {
		return errors.New("alguna parte del reporte no es igual a su consulta de origen")
	}
	return nil
}

func (s *reportScenario) failsNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	return nil
}

func (s *reportScenario) onlyProjectRead() error {
	e := s.env
	if e.projectSpy.calls != 1 || e.itemSpy.calls != 0 || e.sprintSpy.calls != 0 || e.effortSpy.calls != 0 || e.defectSpy.calls != 0 {
		return fmt.Errorf("lecturas de proyecto, ítems, sprints, esfuerzos y defectos = %d, %d, %d, %d y %d; se esperaba 1, 0, 0, 0 y 0",
			e.projectSpy.calls, e.itemSpy.calls, e.sprintSpy.calls, e.effortSpy.calls, e.defectSpy.calls)
	}
	return nil
}

func (s *reportScenario) failsWithReadError(source string) error {
	want := map[string]error{
		"el proyecto":   errProjectRead,
		"las métricas":  errItemsRead,
		"los sprints":   errSprintsRead,
		"los esfuerzos": errEffortsRead,
		"los defectos":  errDefectsRead,
	}[source]
	if !errors.Is(s.err, want) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, want)
	}
	if !reflect.DeepEqual(s.result, ProjectReport{}) {
		return fmt.Errorf("reporte = %+v, se esperaba el valor cero", s.result)
	}
	return nil
}

// laterReadsSkipped verifica las lecturas de sprints, esfuerzos y defectos
// hechas hasta la que falla, con N sprints (RN-38.3).
func (s *reportScenario) laterReadsSkipped(source string) error {
	n := s.sprintsN
	want := map[string][3]int{
		"el proyecto":   {0, 0, 0},
		"las métricas":  {0, 0, 0},
		"los sprints":   {2, n, 1},
		"los esfuerzos": {2, n + 1, 1},
		"los defectos":  {2, 2 * n, 2},
	}[source]
	e := s.env
	if got := [3]int{e.sprintSpy.calls, e.effortSpy.calls, e.defectSpy.calls}; got != want {
		return fmt.Errorf("lecturas de sprints, esfuerzos y defectos = %v, se esperaba %v", got, want)
	}
	return nil
}

func (s *reportScenario) sameAgain() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if s.result.Project.ID != s.projectID || len(s.result.Efforts.Efforts) == 0 {
		return fmt.Errorf("reporte = %+v, se esperaban los datos del proyecto %d y sus esfuerzos", s.result, s.projectID)
	}
	again, err := s.env.reports.Project(s.projectID)
	if err != nil || !reflect.DeepEqual(again, s.result) {
		return fmt.Errorf("la segunda consulta = %+v (%v), se esperaba %+v", again, err, s.result)
	}
	return nil
}

func (s *reportScenario) resultIsACopy() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	r := s.result
	if len(r.Efforts.Efforts) == 0 || len(r.Defects.Defects) == 0 || len(r.Members.Names) == 0 {
		return fmt.Errorf("reporte = %+v, se esperaban esfuerzos, defectos e integrantes", r)
	}
	changed, err := s.env.reports.Project(s.projectID)
	if err != nil {
		return err
	}
	changed.Efforts.Efforts[0].Hours = 999
	changed.Defects.Defects[0].Status = "OTRO"
	changed.Members.Names[0] = "Otro"
	again, err := s.env.reports.Project(s.projectID)
	if err != nil || !reflect.DeepEqual(again, r) {
		return fmt.Errorf("después de modificar el reporte = %+v (%v), se esperaba %+v", again, err, r)
	}
	return nil
}

func (s *reportScenario) nothingChanged() error {
	if after := s.stored(); !reflect.DeepEqual(after, s.before) {
		return errors.New("el reporte modificó los ítems, los sprints, los esfuerzos o los defectos")
	}
	return nil
}
