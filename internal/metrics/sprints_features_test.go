package metrics

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-37-graficos-dashboard.feature con
// godog.
func TestFeatures_HU37GraficosDashboard(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeProjectSprintsScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-37-graficos-dashboard.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-37")
	}
}

// projectSprintsScenario guarda el estado de un escenario de HU-37.
type projectSprintsScenario struct {
	t           *testing.T
	env         *summaryEnv
	projects    map[string]int
	sprints     map[string]sprint.Sprint
	sprintItems map[string][]backlog.Item
	readErrors  map[string]error

	// Última consulta: proyecto consultado, resultado y lecturas de ítems.
	queriedID int
	result    []SprintWithMetrics
	err       error
	itemReads int

	projectID     int
	itemsBefore   []backlog.Item
	sprintsBefore []sprint.Sprint
	effortsBefore []effort.Effort
	defectsBefore []defect.Defect
}

func initializeProjectSprintsScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &projectSprintsScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*s = projectSprintsScenario{
			t:           t,
			env:         newSummaryEnv(),
			projects:    map[string]int{},
			sprints:     map[string]sprint.Sprint{},
			sprintItems: map[string][]backlog.Item{},
			readErrors: map[string]error{
				"los sprints":   errSprintsRead,
				"los ítems":     errItemsRead,
				"los esfuerzos": errEffortsRead,
				"los defectos":  errDefectsRead,
			},
		}
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que existe un sprint cerrado "([^"]*)" en el proyecto "([^"]*)" con un ítem terminado de (\d+) story points$`, s.existingClosedSprint)
	ctx.Step(`^que cierro el sprint "([^"]*)"$`, s.closeSprint)
	ctx.Step(`^que el sprint "([^"]*)" tiene los ítems:$`, s.sprintItemsTable)
	ctx.Step(`^que el proyecto "([^"]*)" tiene los ítems sin sprint:$`, s.projectItemsTable)
	ctx.Step(`^que hay un defecto detectado en el sprint "([^"]*)" y resuelto en el sprint "([^"]*)"$`, s.resolvedDefect)
	ctx.Step(`^que hay un defecto abierto detectado en el sprint "([^"]*)"$`, s.openDefect)
	ctx.Step(`^que falla la lectura de (los sprints|los ítems|los esfuerzos|los defectos)$`, s.failRead)
	ctx.Step(`^consulto las métricas por sprint del proyecto "([^"]*)"$`, s.queryByName)
	ctx.Step(`^consulto las métricas por sprint del proyecto con el ID (-?\d+)$`, s.query)
	ctx.Step(`^la lista tiene los sprints: (.+)$`, s.listHasSprints)
	ctx.Step(`^la lista de sprints está vacía$`, s.listIsEmpty)
	ctx.Step(`^el sprint "([^"]*)" de la lista está (ABIERTO|CERRADO)$`, s.sprintStatusIs)
	ctx.Step(`^en el sprint "([^"]*)" los story points planificados son (\d+) y los completados son (\d+)$`, s.storyPointsAre)
	ctx.Step(`^en el sprint "([^"]*)" el porcentaje de historias completadas es ([\d.]+) con (\d+) de (\d+) ítems$`, s.percentIs)
	ctx.Step(`^en el sprint "([^"]*)" las horas estimadas son ([\d.]+) y las reales son ([\d.]+)$`, s.hoursAre)
	ctx.Step(`^en el sprint "([^"]*)" la desviación es (-?[\d.]+) horas y (-?[\d.]+) %$`, s.deviationIs)
	ctx.Step(`^en el sprint "([^"]*)" los defectos detectados son (\d+) y los resueltos son (\d+)$`, s.defectsAre)
	ctx.Step(`^en el sprint "([^"]*)" todas las métricas están disponibles$`, s.allAvailable)
	ctx.Step(`^en el sprint "([^"]*)" las métricas no disponibles son: (.+)$`, s.notAvailableAre)
	ctx.Step(`^las métricas de cada sprint son iguales a las de su consulta individual$`, s.equalToSprint)
	ctx.Step(`^los ítems del backlog se leyeron (\d+) veces en la consulta$`, s.itemReadsAre)
	ctx.Step(`^la consulta falla porque el proyecto no existe$`, s.failsNotFound)
	ctx.Step(`^la consulta falla con el error de la lectura de (los sprints|los ítems|los esfuerzos|los defectos)$`, s.failsWithReadError)
	ctx.Step(`^no se devuelve ninguna lista$`, s.noList)
	ctx.Step(`^modificar el resultado no cambia la próxima consulta$`, s.resultIsACopy)
	ctx.Step(`^los datos del proyecto no cambiaron$`, s.nothingChanged)
}

func (s *projectSprintsScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name)
	if s.projectID == 0 {
		s.projectID = s.projects[name]
	}
	return nil
}

func (s *projectSprintsScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	return nil
}

func (s *projectSprintsScenario) existingClosedSprint(name, projectName string, points int) error {
	s.sprints[name] = s.env.closedSprint(s.t, s.projects[projectName], name, points)
	return nil
}

func (s *projectSprintsScenario) closeSprint(name string) error {
	closed, err := s.env.sprints.Close(s.sprints[name].ID)
	if err != nil {
		return fmt.Errorf("cerrar el sprint %q: %w", name, err)
	}
	s.sprints[name] = closed
	return nil
}

func (s *projectSprintsScenario) sprintItemsTable(name string, table *godog.Table) error {
	rows, err := parseItemRows(table)
	if err != nil {
		return err
	}
	sp := s.sprints[name]
	for _, r := range rows {
		it := s.env.fullItem(s.t, sp.ProjectID, r.title, r.points, r.hours, r.status, sp.ID, r.effort)
		s.sprintItems[name] = append(s.sprintItems[name], it)
	}
	return nil
}

func (s *projectSprintsScenario) projectItemsTable(projectName string, table *godog.Table) error {
	rows, err := parseItemRows(table)
	if err != nil {
		return err
	}
	for _, r := range rows {
		s.env.fullItem(s.t, s.projects[projectName], r.title, r.points, r.hours, r.status, 0, 0)
	}
	return nil
}

// defectItem devuelve el primer ítem del sprint, al que se asocian sus
// defectos.
func (s *projectSprintsScenario) defectItem(sprintName string) (backlog.Item, error) {
	items := s.sprintItems[sprintName]
	if len(items) == 0 {
		return backlog.Item{}, fmt.Errorf("el sprint %q no tiene ítems para asociar el defecto", sprintName)
	}
	return items[0], nil
}

func (s *projectSprintsScenario) resolvedDefect(detection, resolution string) error {
	it, err := s.defectItem(detection)
	if err != nil {
		return err
	}
	s.env.saveDefect(s.t, it.ID, s.sprints[detection].ID, defect.StatusResolved, s.sprints[resolution].ID)
	return nil
}

func (s *projectSprintsScenario) openDefect(detection string) error {
	it, err := s.defectItem(detection)
	if err != nil {
		return err
	}
	s.env.saveDefect(s.t, it.ID, s.sprints[detection].ID, defect.StatusOpen, 0)
	return nil
}

func (s *projectSprintsScenario) failRead(source string) error {
	err := s.readErrors[source]
	switch source {
	case "los sprints":
		s.env.sprintSpy.listErr = err
	case "los ítems":
		s.env.itemSpy.err = err
	case "los esfuerzos":
		s.env.effortSpy.err = err
	case "los defectos":
		s.env.defectSpy.err = err
	}
	return nil
}

func (s *projectSprintsScenario) queryByName(name string) error {
	return s.query(s.projects[name])
}

// query consulta las métricas por sprint y guarda cuántas veces se leyeron
// los ítems en esa consulta.
func (s *projectSprintsScenario) query(id int) error {
	s.snapshot()
	s.env.resetReads()
	s.queriedID = id
	s.result, s.err = s.env.metrics.ProjectSprints(id)
	s.itemReads = s.env.itemSpy.listCalls
	return nil
}

// snapshot guarda los datos del proyecto antes de la primera consulta para
// comprobar después que no cambiaron (CA-37.5).
func (s *projectSprintsScenario) snapshot() {
	if s.itemsBefore != nil || s.projectID == 0 {
		return
	}
	s.itemsBefore, s.sprintsBefore, s.effortsBefore, s.defectsBefore = s.projectData()
}

func (s *projectSprintsScenario) projectData() ([]backlog.Item, []sprint.Sprint, []effort.Effort, []defect.Defect) {
	items, _ := s.env.itemSpy.Repository.ListByProject(s.projectID)
	sprints := s.env.sprintsOf(s.projectID)
	var efforts []effort.Effort
	for _, sp := range sprints {
		found, _ := s.env.effortSpy.Repository.ListBySprint(sp.ID)
		efforts = append(efforts, found...)
	}
	defects, _ := s.env.defectSpy.Repository.List()
	return items, sprints, efforts, defects
}

// ok devuelve un error si la última consulta falló.
func (s *projectSprintsScenario) ok() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	return nil
}

// entry devuelve el elemento de la lista del sprint con ese nombre, y
// verifica que sus métricas sean de ese mismo sprint.
func (s *projectSprintsScenario) entry(name string) (SprintWithMetrics, error) {
	if err := s.ok(); err != nil {
		return SprintWithMetrics{}, err
	}
	id := s.sprints[name].ID
	for _, e := range s.result {
		if e.Sprint.ID == id {
			if e.Metrics.SprintID != id {
				return SprintWithMetrics{}, fmt.Errorf("el sprint %q tiene las métricas del sprint %d", name, e.Metrics.SprintID)
			}
			return e, nil
		}
	}
	return SprintWithMetrics{}, fmt.Errorf("el sprint %q no está en la lista", name)
}

func (s *projectSprintsScenario) listHasSprints(list string) error {
	if err := s.ok(); err != nil {
		return err
	}
	var want, got []string
	for _, name := range strings.Split(list, ",") {
		want = append(want, strings.TrimSpace(name))
	}
	for _, e := range s.result {
		got = append(got, e.Sprint.Name)
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("sprints de la lista = %v, se esperaba %v", got, want)
	}
	return nil
}

func (s *projectSprintsScenario) listIsEmpty() error {
	if err := s.ok(); err != nil {
		return err
	}
	if s.result == nil || len(s.result) != 0 {
		return fmt.Errorf("lista = %#v, se esperaba una lista vacía, no nil", s.result)
	}
	return nil
}

// sprintStatusIs verifica el estado del sprint en la lista y que el sprint
// sea igual al guardado.
func (s *projectSprintsScenario) sprintStatusIs(name, status string) error {
	e, err := s.entry(name)
	if err != nil {
		return err
	}
	if e.Sprint.Status != sprint.Status(status) {
		return fmt.Errorf("estado del sprint %q = %s, se esperaba %s", name, e.Sprint.Status, status)
	}
	if e.Sprint != s.sprints[name] {
		return fmt.Errorf("sprint %q = %+v, el guardado es %+v", name, e.Sprint, s.sprints[name])
	}
	return nil
}

func (s *projectSprintsScenario) storyPointsAre(name string, planned, completed int) error {
	e, err := s.entry(name)
	if err != nil {
		return err
	}
	p := e.Metrics.Points
	if p.Planned != planned || p.Completed != completed {
		return fmt.Errorf("sprint %q: %d SP planificados y %d completados, se esperaba %d y %d", name, p.Planned, p.Completed, planned, completed)
	}
	return nil
}

func (s *projectSprintsScenario) percentIs(name string, want float64, finished, items int) error {
	e, err := s.entry(name)
	if err != nil {
		return err
	}
	c := e.Metrics.Completion
	if !c.PercentAvailable || !almostEqual(c.Percent, want) || c.Finished != finished || c.Items != items {
		return fmt.Errorf("sprint %q: porcentaje = %+v, se esperaba %v con %d de %d", name, c, want, finished, items)
	}
	return nil
}

func (s *projectSprintsScenario) hoursAre(name string, estimated, actual float64) error {
	e, err := s.entry(name)
	if err != nil {
		return err
	}
	h := e.Metrics.Hours
	if !almostEqual(h.Estimated, estimated) || !almostEqual(h.Actual, actual) {
		return fmt.Errorf("sprint %q: %v horas estimadas y %v reales, se esperaba %v y %v", name, h.Estimated, h.Actual, estimated, actual)
	}
	return nil
}

func (s *projectSprintsScenario) deviationIs(name string, hours, percent float64) error {
	e, err := s.entry(name)
	if err != nil {
		return err
	}
	h := e.Metrics.Hours
	if !h.DeviationAvailable || !almostEqual(h.Deviation, hours) || !almostEqual(h.DeviationPercent, percent) {
		return fmt.Errorf("sprint %q: desviación = %v h y %v %% (disponible %v), se esperaba %v h y %v %%", name, h.Deviation, h.DeviationPercent, h.DeviationAvailable, hours, percent)
	}
	return nil
}

func (s *projectSprintsScenario) defectsAre(name string, detected, resolved int) error {
	e, err := s.entry(name)
	if err != nil {
		return err
	}
	if want := (Defects{Detected: detected, Resolved: resolved}); e.Metrics.Defects != want {
		return fmt.Errorf("sprint %q: defectos = %+v, se esperaba %+v", name, e.Metrics.Defects, want)
	}
	return nil
}

// notAvailable devuelve los nombres de las métricas no disponibles de un
// sprint, en el orden de RN-35.6.
func notAvailable(m SprintMetrics) []string {
	var names []string
	if !m.Completion.PercentAvailable {
		names = append(names, "porcentaje")
	}
	if !m.Hours.DeviationAvailable {
		names = append(names, "desviación")
	}
	return names
}

func (s *projectSprintsScenario) allAvailable(name string) error {
	e, err := s.entry(name)
	if err != nil {
		return err
	}
	if got := notAvailable(e.Metrics); len(got) != 0 {
		return fmt.Errorf("sprint %q: métricas no disponibles = %v, se esperaba ninguna", name, got)
	}
	return nil
}

func (s *projectSprintsScenario) notAvailableAre(name, list string) error {
	e, err := s.entry(name)
	if err != nil {
		return err
	}
	var want []string
	for _, n := range strings.Split(list, ",") {
		want = append(want, strings.TrimSpace(n))
	}
	if got := notAvailable(e.Metrics); !slices.Equal(got, want) {
		return fmt.Errorf("sprint %q: métricas no disponibles = %v, se esperaba %v", name, got, want)
	}
	return nil
}

// equalToSprint compara cada elemento con Sprint(id) de HU-35 (CA-37.3).
func (s *projectSprintsScenario) equalToSprint() error {
	if err := s.ok(); err != nil {
		return err
	}
	if len(s.result) == 0 {
		return fmt.Errorf("la lista está vacía")
	}
	for _, e := range s.result {
		want, err := s.env.metrics.Sprint(e.Sprint.ID)
		if err != nil {
			return fmt.Errorf("Sprint(%d): %v", e.Sprint.ID, err)
		}
		if e.Metrics != want {
			return fmt.Errorf("sprint %d: métricas = %+v, Sprint da %+v", e.Sprint.ID, e.Metrics, want)
		}
	}
	return nil
}

func (s *projectSprintsScenario) itemReadsAre(want int) error {
	if err := s.ok(); err != nil {
		return err
	}
	if s.itemReads != want {
		return fmt.Errorf("lecturas de ítems = %d, se esperaba %d", s.itemReads, want)
	}
	return nil
}

func (s *projectSprintsScenario) failsNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	return nil
}

func (s *projectSprintsScenario) failsWithReadError(source string) error {
	if want := s.readErrors[source]; !errors.Is(s.err, want) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, want)
	}
	return nil
}

func (s *projectSprintsScenario) noList() error {
	if s.result != nil {
		return fmt.Errorf("lista = %+v, se esperaba nil", s.result)
	}
	return nil
}

// resultIsACopy modifica la lista devuelta y comprueba que la próxima
// consulta devuelve lo mismo que la anterior (CA-37.5).
func (s *projectSprintsScenario) resultIsACopy() error {
	if err := s.ok(); err != nil {
		return err
	}
	if len(s.result) == 0 {
		return fmt.Errorf("la lista está vacía")
	}
	want := make([]SprintWithMetrics, len(s.result))
	copy(want, s.result)
	s.result[0].Sprint.Name = "Modificado"
	s.result[0].Metrics.Points.Planned = 999
	again, err := s.env.metrics.ProjectSprints(s.queriedID)
	if err != nil || !reflect.DeepEqual(again, want) {
		return fmt.Errorf("la segunda consulta dio %+v (err %v), se esperaba %+v", again, err, want)
	}
	s.result = again
	return nil
}

func (s *projectSprintsScenario) nothingChanged() error {
	if err := s.ok(); err != nil {
		return err
	}
	items, sprints, efforts, defects := s.projectData()
	if s.itemsBefore == nil || !reflect.DeepEqual(items, s.itemsBefore) ||
		!reflect.DeepEqual(sprints, s.sprintsBefore) ||
		!reflect.DeepEqual(efforts, s.effortsBefore) ||
		!reflect.DeepEqual(defects, s.defectsBefore) {
		return fmt.Errorf("la consulta modificó los datos del proyecto")
	}
	return nil
}
