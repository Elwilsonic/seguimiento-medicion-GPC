package metrics

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-35-consulta-metricas.feature con
// godog.
func TestFeatures_HU35ConsultarMetricas(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeSummaryScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-35-consulta-metricas.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-35")
	}
}

// summaryScenario guarda el estado de un escenario de HU-35.
type summaryScenario struct {
	t           *testing.T
	env         *summaryEnv
	projects    map[string]int
	sprints     map[string]sprint.Sprint
	sprintItems map[string][]backlog.Item
	readErrors  map[string]error

	// Última consulta: nivel ("sprint" o "proyecto"), ID consultado y
	// resultado.
	level         string
	queriedID     int
	sprintResult  SprintMetrics
	projectResult ProjectMetrics
	err           error
	itemReads     int
	effortReads   int
	defectReads   int

	projectID     int
	itemsBefore   []backlog.Item
	sprintsBefore []sprint.Sprint
	effortsBefore []effort.Effort
	defectsBefore []defect.Defect
}

func initializeSummaryScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &summaryScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*s = summaryScenario{
			t:           t,
			env:         newSummaryEnv(),
			projects:    map[string]int{},
			sprints:     map[string]sprint.Sprint{},
			sprintItems: map[string][]backlog.Item{},
			readErrors: map[string]error{
				"el sprint":     errSprintRead,
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
	ctx.Step(`^que cierro el sprint "([^"]*)"$`, s.closeSprint)
	ctx.Step(`^que el sprint "([^"]*)" tiene los ítems:$`, s.sprintItemsTable)
	ctx.Step(`^que el proyecto "([^"]*)" tiene los ítems sin sprint:$`, s.projectItemsTable)
	ctx.Step(`^que hay un defecto detectado en el sprint "([^"]*)" y resuelto en el sprint "([^"]*)"$`, s.resolvedDefect)
	ctx.Step(`^que hay un defecto abierto detectado en el sprint "([^"]*)"$`, s.openDefect)
	ctx.Step(`^que falla la lectura de (el sprint|los sprints|los ítems|los esfuerzos|los defectos)$`, s.failRead)
	ctx.Step(`^consulto las métricas del (sprint|proyecto) "([^"]*)"$`, s.queryByName)
	ctx.Step(`^consulto las métricas del (sprint|proyecto) con el ID (-?\d+)$`, s.query)
	ctx.Step(`^los story points planificados son (\d+) y los completados son (\d+)$`, s.storyPointsAre)
	ctx.Step(`^hay (\d+) ítems? sin estimar$`, s.unestimatedAre)
	ctx.Step(`^la velocidad es ([\d.]+) con (\d+) sprints? cerrados?$`, s.velocityIs)
	ctx.Step(`^el porcentaje de historias completadas es ([\d.]+) con (\d+) de (\d+) ítems$`, s.percentIs)
	ctx.Step(`^los ítems totales son (\d+) y los terminados son (\d+)$`, s.itemCountsAre)
	ctx.Step(`^las horas estimadas son ([\d.]+) y las reales son ([\d.]+)$`, s.hoursAre)
	ctx.Step(`^la desviación es (-?[\d.]+) horas y (-?[\d.]+) %$`, s.deviationIs)
	ctx.Step(`^hay (\d+) ítems? sin horas estimadas$`, s.unestimatedHoursAre)
	ctx.Step(`^los defectos detectados son (\d+) y los resueltos son (\d+)$`, s.defectsAre)
	ctx.Step(`^todas las métricas están disponibles$`, s.allAvailable)
	ctx.Step(`^las métricas no disponibles son: (.+)$`, s.notAvailableAre)
	ctx.Step(`^cada métrica es igual a la de su operación separada$`, s.equalToSeparate)
	ctx.Step(`^los ítems del backlog se leyeron (\d+) veces en la consulta$`, s.itemReadsAre)
	ctx.Step(`^la consulta falla porque el (sprint|proyecto) no existe$`, s.failsNotFound)
	ctx.Step(`^la consulta falla con el error de la lectura de (el sprint|los sprints|los ítems|los esfuerzos|los defectos)$`, s.failsWithReadError)
	ctx.Step(`^no se leyeron los esfuerzos ni los defectos$`, s.effortsAndDefectsNotRead)
	ctx.Step(`^no se leyeron los defectos$`, s.defectsNotRead)
	ctx.Step(`^modificar el resultado no cambia la próxima consulta$`, s.resultIsACopy)
	ctx.Step(`^los datos del proyecto no cambiaron$`, s.nothingChanged)
}

func (s *summaryScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name)
	if s.projectID == 0 {
		s.projectID = s.projects[name]
	}
	return nil
}

func (s *summaryScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	return nil
}

func (s *summaryScenario) closeSprint(name string) error {
	closed, err := s.env.sprints.Close(s.sprints[name].ID)
	if err != nil {
		return fmt.Errorf("cerrar el sprint %q: %w", name, err)
	}
	s.sprints[name] = closed
	return nil
}

// itemRow es una fila de las tablas de ítems: titulo, sp, horas, estado y,
// en los ítems de un sprint, esfuerzo.
type itemRow struct {
	title  string
	points int
	hours  float64
	status backlog.Status
	effort float64
}

func parseItemRows(table *godog.Table) ([]itemRow, error) {
	header := map[string]int{}
	for i, cell := range table.Rows[0].Cells {
		header[cell.Value] = i
	}
	var rows []itemRow
	for _, row := range table.Rows[1:] {
		value := func(column string) string {
			if i, ok := header[column]; ok {
				return row.Cells[i].Value
			}
			return "0"
		}
		points, err := strconv.Atoi(value("sp"))
		if err != nil {
			return nil, fmt.Errorf("story points inválidos %q: %w", value("sp"), err)
		}
		hours, err := strconv.ParseFloat(value("horas"), 64)
		if err != nil {
			return nil, fmt.Errorf("horas inválidas %q: %w", value("horas"), err)
		}
		effortHours, err := strconv.ParseFloat(value("esfuerzo"), 64)
		if err != nil {
			return nil, fmt.Errorf("esfuerzo inválido %q: %w", value("esfuerzo"), err)
		}
		rows = append(rows, itemRow{title: value("titulo"), points: points, hours: hours, status: backlog.Status(value("estado")), effort: effortHours})
	}
	return rows, nil
}

func (s *summaryScenario) sprintItemsTable(name string, table *godog.Table) error {
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

func (s *summaryScenario) projectItemsTable(projectName string, table *godog.Table) error {
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
func (s *summaryScenario) defectItem(sprintName string) (backlog.Item, error) {
	items := s.sprintItems[sprintName]
	if len(items) == 0 {
		return backlog.Item{}, fmt.Errorf("el sprint %q no tiene ítems para asociar el defecto", sprintName)
	}
	return items[0], nil
}

func (s *summaryScenario) resolvedDefect(detection, resolution string) error {
	it, err := s.defectItem(detection)
	if err != nil {
		return err
	}
	s.env.saveDefect(s.t, it.ID, s.sprints[detection].ID, defect.StatusResolved, s.sprints[resolution].ID)
	return nil
}

func (s *summaryScenario) openDefect(detection string) error {
	it, err := s.defectItem(detection)
	if err != nil {
		return err
	}
	s.env.saveDefect(s.t, it.ID, s.sprints[detection].ID, defect.StatusOpen, 0)
	return nil
}

func (s *summaryScenario) failRead(source string) error {
	err := s.readErrors[source]
	switch source {
	case "el sprint":
		s.env.sprintSpy.getErr = err
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

func (s *summaryScenario) queryByName(level, name string) error {
	if level == "sprint" {
		return s.query(level, s.sprints[name].ID)
	}
	return s.query(level, s.projects[name])
}

// query consulta las métricas y guarda cuántas veces se leyeron los ítems,
// los esfuerzos y los defectos en esa consulta.
func (s *summaryScenario) query(level string, id int) error {
	s.snapshot()
	s.env.resetReads()
	s.level, s.queriedID = level, id
	if level == "sprint" {
		s.sprintResult, s.err = s.env.metrics.Sprint(id)
	} else {
		s.projectResult, s.err = s.env.metrics.Project(id)
	}
	s.itemReads, s.effortReads, s.defectReads = s.env.itemSpy.listCalls, s.env.effortSpy.listCalls, s.env.defectSpy.listCalls
	return nil
}

// snapshot guarda los datos del proyecto antes de la primera consulta para
// comprobar después que no cambiaron (CA-35.7).
func (s *summaryScenario) snapshot() {
	if s.itemsBefore != nil || s.projectID == 0 {
		return
	}
	s.itemsBefore, s.sprintsBefore, s.effortsBefore, s.defectsBefore = s.projectData()
}

func (s *summaryScenario) projectData() ([]backlog.Item, []sprint.Sprint, []effort.Effort, []defect.Defect) {
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

// summaryView es la parte común de las métricas de un sprint y de un
// proyecto, para que los pasos sirvan en los dos niveles.
type summaryView struct {
	id                int
	points            SprintPoints
	velocity          float64
	velocityAvailable bool
	closedSprints     int
	completion        Completion
	hours             Hours
	defects           Defects
}

// result devuelve la última consulta, o un error si falló o si devolvió las
// métricas de otro ID.
func (s *summaryScenario) result() (summaryView, error) {
	if s.err != nil {
		return summaryView{}, fmt.Errorf("la consulta falló: %v", s.err)
	}
	var v summaryView
	if s.level == "sprint" {
		r := s.sprintResult
		v = summaryView{id: r.SprintID, points: r.Points, completion: r.Completion, hours: r.Hours, defects: r.Defects}
	} else {
		r := s.projectResult
		v = summaryView{
			id:                r.ProjectID,
			points:            SprintPoints{Planned: r.Points.Planned, Completed: r.Points.Completed, Unestimated: r.Points.Unestimated},
			velocity:          r.Points.Velocity,
			velocityAvailable: r.Points.VelocityAvailable,
			closedSprints:     r.Points.ClosedSprints,
			completion:        r.Completion,
			hours:             r.Hours,
			defects:           r.Defects,
		}
	}
	if v.id != s.queriedID {
		return summaryView{}, fmt.Errorf("la consulta del %s devolvió el ID %d, se esperaba %d", s.level, v.id, s.queriedID)
	}
	return v, nil
}

func (s *summaryScenario) storyPointsAre(planned, completed int) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	if v.points.Planned != planned || v.points.Completed != completed {
		return fmt.Errorf("story points = %d planificados y %d completados, se esperaba %d y %d", v.points.Planned, v.points.Completed, planned, completed)
	}
	return nil
}

func (s *summaryScenario) unestimatedAre(want int) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	if v.points.Unestimated != want {
		return fmt.Errorf("ítems sin estimar = %d, se esperaba %d", v.points.Unestimated, want)
	}
	return nil
}

func (s *summaryScenario) velocityIs(want float64, closed int) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	if !v.velocityAvailable || !almostEqual(v.velocity, want) || v.closedSprints != closed {
		return fmt.Errorf("velocidad = %v (disponible %v) con %d sprints, se esperaba %v con %d", v.velocity, v.velocityAvailable, v.closedSprints, want, closed)
	}
	return nil
}

func (s *summaryScenario) percentIs(want float64, finished, items int) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	c := v.completion
	if !c.PercentAvailable || !almostEqual(c.Percent, want) || c.Finished != finished || c.Items != items {
		return fmt.Errorf("porcentaje = %+v, se esperaba %v con %d de %d", c, want, finished, items)
	}
	return nil
}

func (s *summaryScenario) itemCountsAre(items, finished int) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	if v.completion.Items != items || v.completion.Finished != finished {
		return fmt.Errorf("ítems = %d y terminados = %d, se esperaba %d y %d", v.completion.Items, v.completion.Finished, items, finished)
	}
	return nil
}

func (s *summaryScenario) hoursAre(estimated, actual float64) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	if !almostEqual(v.hours.Estimated, estimated) || !almostEqual(v.hours.Actual, actual) {
		return fmt.Errorf("horas = %v estimadas y %v reales, se esperaba %v y %v", v.hours.Estimated, v.hours.Actual, estimated, actual)
	}
	return nil
}

func (s *summaryScenario) deviationIs(hours, percent float64) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	h := v.hours
	if !h.DeviationAvailable || !almostEqual(h.Deviation, hours) || !almostEqual(h.DeviationPercent, percent) {
		return fmt.Errorf("desviación = %v h y %v %% (disponible %v), se esperaba %v h y %v %%", h.Deviation, h.DeviationPercent, h.DeviationAvailable, hours, percent)
	}
	return nil
}

func (s *summaryScenario) unestimatedHoursAre(want int) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	if v.hours.Unestimated != want {
		return fmt.Errorf("ítems sin horas estimadas = %d, se esperaba %d", v.hours.Unestimated, want)
	}
	return nil
}

func (s *summaryScenario) defectsAre(detected, resolved int) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	if want := (Defects{Detected: detected, Resolved: resolved}); v.defects != want {
		return fmt.Errorf("defectos = %+v, se esperaba %+v", v.defects, want)
	}
	return nil
}

// notAvailable devuelve los nombres de las métricas no disponibles de la
// consulta, en el orden de RN-35.6.
func (v summaryView) notAvailable(level string) []string {
	var names []string
	if level == "proyecto" && !v.velocityAvailable {
		names = append(names, "velocidad")
	}
	if !v.completion.PercentAvailable {
		names = append(names, "porcentaje")
	}
	if !v.hours.DeviationAvailable {
		names = append(names, "desviación")
	}
	return names
}

func (s *summaryScenario) allAvailable() error {
	v, err := s.result()
	if err != nil {
		return err
	}
	if got := v.notAvailable(s.level); len(got) != 0 {
		return fmt.Errorf("métricas no disponibles = %v, se esperaba ninguna", got)
	}
	return nil
}

func (s *summaryScenario) notAvailableAre(list string) error {
	v, err := s.result()
	if err != nil {
		return err
	}
	var want []string
	for _, name := range strings.Split(list, ",") {
		want = append(want, strings.TrimSpace(name))
	}
	if got := v.notAvailable(s.level); !slices.Equal(got, want) {
		return fmt.Errorf("métricas no disponibles = %v, se esperaba %v", got, want)
	}
	return nil
}

func (s *summaryScenario) equalToSeparate() error {
	if _, err := s.result(); err != nil {
		return err
	}
	if s.level == "sprint" {
		if want := separateSprint(s.t, s.env.metricsEnv, s.queriedID); s.sprintResult != want {
			return fmt.Errorf("Sprint = %+v, las operaciones separadas dan %+v", s.sprintResult, want)
		}
		return nil
	}
	if want := separateProject(s.t, s.env.metricsEnv, s.queriedID); s.projectResult != want {
		return fmt.Errorf("Project = %+v, las operaciones separadas dan %+v", s.projectResult, want)
	}
	return nil
}

func (s *summaryScenario) itemReadsAre(want int) error {
	if _, err := s.result(); err != nil {
		return err
	}
	if s.itemReads != want {
		return fmt.Errorf("lecturas de ítems = %d, se esperaba %d", s.itemReads, want)
	}
	return nil
}

func (s *summaryScenario) failsNotFound(level string) error {
	want := sprint.ErrSprintNotFound
	if level == "proyecto" {
		want = project.ErrProjectNotFound
	}
	if !errors.Is(s.err, want) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, want)
	}
	return nil
}

func (s *summaryScenario) failsWithReadError(source string) error {
	if want := s.readErrors[source]; !errors.Is(s.err, want) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, want)
	}
	return nil
}

func (s *summaryScenario) effortsAndDefectsNotRead() error {
	if s.effortReads != 0 || s.defectReads != 0 {
		return fmt.Errorf("se leyeron los esfuerzos %d veces y los defectos %d veces, se esperaba ninguna", s.effortReads, s.defectReads)
	}
	return nil
}

func (s *summaryScenario) defectsNotRead() error {
	if s.defectReads != 0 {
		return fmt.Errorf("se leyeron los defectos %d veces, se esperaba ninguna", s.defectReads)
	}
	return nil
}

// resultIsACopy modifica una copia del resultado y comprueba que la próxima
// consulta devuelve lo mismo que la anterior (CA-35.7).
func (s *summaryScenario) resultIsACopy() error {
	if _, err := s.result(); err != nil {
		return err
	}
	if s.level == "sprint" {
		first := s.sprintResult
		changed := first
		changed.Points.Planned, changed.Hours.Actual = 999, -1
		again, err := s.env.metrics.Sprint(s.queriedID)
		if err != nil || again != first {
			return fmt.Errorf("la segunda consulta dio %+v (err %v), se esperaba %+v", again, err, first)
		}
		return nil
	}
	first := s.projectResult
	changed := first
	changed.Points.Velocity, changed.Defects.Detected = 999, 999
	again, err := s.env.metrics.Project(s.queriedID)
	if err != nil || again != first {
		return fmt.Errorf("la segunda consulta dio %+v (err %v), se esperaba %+v", again, err, first)
	}
	return nil
}

func (s *summaryScenario) nothingChanged() error {
	if _, err := s.result(); err != nil {
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
