package report

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-39-reporte-sprint.feature con
// godog.
func TestFeatures_HU39ReporteSprint(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeSprintReportScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-39-reporte-sprint.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-39")
	}
}

// sprintReportScenario guarda el estado de un escenario de HU-39.
type sprintReportScenario struct {
	t        *testing.T
	env      *sprintReportEnv
	projects map[string]int
	sprints  map[string]sprint.Sprint
	items    map[string]backlog.Item

	// failSource es la lectura que tiene que fallar ("" = ninguna).
	failSource string

	projectID int
	sprintID  int
	result    SprintReport
	err       error
	before    storedData
}

func initializeSprintReportScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &sprintReportScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*s = sprintReportScenario{
			t:        t,
			env:      newSprintReportEnv(),
			projects: map[string]int{},
			sprints:  map[string]sprint.Sprint{},
			items:    map[string]backlog.Item{},
		}
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" del "([^"]*)" al "([^"]*)" con el objetivo "([^"]*)" en el proyecto "([^"]*)"$`, s.existingSprintWithData)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que cierro el sprint "([^"]*)"$`, s.closeSprint)
	ctx.Step(`^que el sprint "([^"]*)" tiene los ítems:$`, s.sprintItemsTable)
	ctx.Step(`^que se registraron los esfuerzos:$`, s.effortsTable)
	ctx.Step(`^que hay un defecto (ABIERTO|EN_CORRECCION) del ítem "([^"]*)" detectado en el sprint "([^"]*)"$`, s.unresolvedDefect)
	ctx.Step(`^que hay un defecto RESUELTO del ítem "([^"]*)" detectado en el sprint "([^"]*)" y resuelto en el sprint "([^"]*)"$`, s.resolvedDefect)
	ctx.Step(`^que falla la lectura de (el sprint|las métricas|los esfuerzos|los defectos) del reporte del sprint$`, s.failRead)
	ctx.Step(`^genero el reporte del sprint "([^"]*)"$`, s.generateByName)
	ctx.Step(`^genero el reporte del sprint con el ID (-?\d+)$`, s.generate)
	ctx.Step(`^el sprint del reporte es "([^"]*)" en estado "([^"]*)" del "([^"]*)" al "([^"]*)" con el objetivo "([^"]*)"$`, s.sprintWithDataIs)
	ctx.Step(`^el sprint del reporte es "([^"]*)" en estado "([^"]*)"$`, s.sprintIs)
	ctx.Step(`^los story points planificados son (\d+) y los completados son (\d+)$`, s.storyPointsAre)
	ctx.Step(`^los ítems totales son (\d+) y los terminados son (\d+)$`, s.itemCountsAre)
	ctx.Step(`^las horas estimadas son ([\d.]+) y las reales son ([\d.]+)$`, s.hoursAre)
	ctx.Step(`^los defectos detectados son (\d+) y los resueltos son (\d+)$`, s.defectsAre)
	ctx.Step(`^las métricas no disponibles son: (.+)$`, s.notAvailableAre)
	ctx.Step(`^los esfuerzos del reporte son:$`, s.effortsAre)
	ctx.Step(`^el total de horas es ([\d.]+)$`, s.totalIs)
	ctx.Step(`^el total de horas es la suma exacta de los esfuerzos$`, s.totalIsExactSum)
	ctx.Step(`^el total de horas es igual a las horas reales de las métricas$`, s.totalEqualsActual)
	ctx.Step(`^los defectos del reporte son:$`, s.defectsListIs)
	ctx.Step(`^los defectos del reporte coinciden con los detectados y resueltos de las métricas$`, s.defectsMatchMetrics)
	ctx.Step(`^los integrantes son "([^"]*)"$`, s.membersAre)
	ctx.Step(`^el reporte no tiene esfuerzos, defectos ni integrantes$`, s.emptyParts)
	ctx.Step(`^el reporte no tiene defectos$`, s.noDefects)
	ctx.Step(`^cada parte del reporte es igual a su consulta de origen$`, s.equalToSources)
	ctx.Step(`^el reporte falla porque el sprint no existe$`, s.failsNotFound)
	ctx.Step(`^solo se leyó el sprint$`, s.onlySprintRead)
	ctx.Step(`^el reporte falla con el error de la lectura de (el sprint|las métricas|los esfuerzos|los defectos)$`, s.failsWithReadError)
	ctx.Step(`^no se hicieron las lecturas siguientes a (el sprint|las métricas|los esfuerzos|los defectos)$`, s.laterReadsSkipped)
	ctx.Step(`^generar el reporte de nuevo da el mismo resultado$`, s.sameAgain)
	ctx.Step(`^modificar el reporte no cambia el próximo reporte$`, s.resultIsACopy)
	ctx.Step(`^los datos guardados no cambiaron$`, s.nothingChanged)
}

func (s *sprintReportScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name, "")
	if s.projectID == 0 {
		s.projectID = s.projects[name]
	}
	return nil
}

func (s *sprintReportScenario) existingSprintWithData(name, start, end, goal, projectName string) error {
	from, err := time.Parse(time.DateOnly, start)
	if err != nil {
		return err
	}
	to, err := time.Parse(time.DateOnly, end)
	if err != nil {
		return err
	}
	sp, err := s.env.sprints.Create(s.projects[projectName], name, from, to, goal)
	if err != nil {
		return err
	}
	s.sprints[name] = sp
	return nil
}

func (s *sprintReportScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	return nil
}

func (s *sprintReportScenario) closeSprint(name string) error {
	s.env.closeSprint(s.t, s.sprints[name].ID)
	return nil
}

// sprintItemsTable crea los ítems de una tabla con titulo, sp, horas y
// estado, asignados al sprint.
func (s *sprintReportScenario) sprintItemsTable(name string, table *godog.Table) error {
	sp := s.sprints[name]
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
		s.items[title] = s.env.item(s.t, sp.ProjectID, title, points, hours, backlog.Status(row.Cells[3].Value), sp.ID)
	}
	return nil
}

// effortsTable registra los esfuerzos de una tabla con sprint, item,
// integrante y horas.
func (s *sprintReportScenario) effortsTable(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		hours, err := strconv.ParseFloat(row.Cells[3].Value, 64)
		if err != nil {
			return err
		}
		s.env.effort(s.t, s.sprints[row.Cells[0].Value].ID, s.items[row.Cells[1].Value].ID, row.Cells[2].Value, hours)
	}
	return nil
}

func (s *sprintReportScenario) unresolvedDefect(status, item, sprintName string) error {
	s.env.saveDefect(s.t, s.items[item].ID, s.sprints[sprintName].ID, defect.Status(status), 0)
	return nil
}

func (s *sprintReportScenario) resolvedDefect(item, detection, resolution string) error {
	s.env.saveDefect(s.t, s.items[item].ID, s.sprints[detection].ID, defect.StatusResolved, s.sprints[resolution].ID)
	return nil
}

func (s *sprintReportScenario) failRead(source string) error {
	s.failSource = source
	return nil
}

// applyFailure hace fallar la lectura del reporte de esa fuente (RN-39.3):
// las de esfuerzos y defectos fallan después de las que ya hizo
// metrics.Manager.Sprint.
func (s *sprintReportScenario) applyFailure() {
	switch s.failSource {
	case "el sprint":
		s.env.sprintGetSpy.getFailAt = 1
	case "las métricas":
		s.env.itemSpy.failAt = 1
	case "los esfuerzos":
		s.env.effortSpy.failAt = 2
	case "los defectos":
		s.env.defectSpy.failAt = 2
	}
}

// stored lee los datos del proyecto directo de los repositorios, sin pasar
// por los espías.
func (s *sprintReportScenario) stored() storedData {
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

func (s *sprintReportScenario) generateByName(name string) error {
	return s.generate(strconv.Itoa(s.sprints[name].ID))
}

func (s *sprintReportScenario) generate(idText string) error {
	id, err := strconv.Atoi(idText)
	if err != nil {
		return err
	}
	s.sprintID = id
	s.before = s.stored()
	s.env.resetSprintReads()
	s.applyFailure()
	s.result, s.err = s.env.reports.Sprint(id)
	return nil
}

func (s *sprintReportScenario) succeeded() error {
	if s.err != nil {
		return fmt.Errorf("el reporte falló: %v", s.err)
	}
	return nil
}

func (s *sprintReportScenario) sprintIs(name, status string) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	sp := s.result.Sprint
	if sp.ID != s.sprintID || sp.Name != name || string(sp.Status) != status {
		return fmt.Errorf("sprint = %+v, se esperaba el sprint %d %q en estado %s", sp, s.sprintID, name, status)
	}
	return nil
}

func (s *sprintReportScenario) sprintWithDataIs(name, status, start, end, goal string) error {
	if err := s.sprintIs(name, status); err != nil {
		return err
	}
	sp := s.result.Sprint
	if sp.StartDate.Format(time.DateOnly) != start || sp.EndDate.Format(time.DateOnly) != end || sp.Goal != goal {
		return fmt.Errorf("sprint del %s al %s con objetivo %q, se esperaba del %s al %s con %q",
			sp.StartDate.Format(time.DateOnly), sp.EndDate.Format(time.DateOnly), sp.Goal, start, end, goal)
	}
	return nil
}

// metricsOfSprint verifica que el reporte tenga las métricas del sprint
// consultado antes de comparar sus valores.
func (s *sprintReportScenario) metricsOfSprint() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if s.result.Metrics.SprintID != s.sprintID {
		return fmt.Errorf("métricas del sprint %d, se esperaban las del sprint %d", s.result.Metrics.SprintID, s.sprintID)
	}
	return nil
}

func (s *sprintReportScenario) storyPointsAre(planned, completed int) error {
	if err := s.metricsOfSprint(); err != nil {
		return err
	}
	if p := s.result.Metrics.Points; p.Planned != planned || p.Completed != completed {
		return fmt.Errorf("story points = %d y %d, se esperaba %d y %d", p.Planned, p.Completed, planned, completed)
	}
	return nil
}

func (s *sprintReportScenario) itemCountsAre(total, finished int) error {
	if err := s.metricsOfSprint(); err != nil {
		return err
	}
	if c := s.result.Metrics.Completion; c.Items != total || c.Finished != finished {
		return fmt.Errorf("ítems = %d y %d terminados, se esperaba %d y %d", c.Items, c.Finished, total, finished)
	}
	return nil
}

func (s *sprintReportScenario) hoursAre(estimated, actual float64) error {
	if err := s.metricsOfSprint(); err != nil {
		return err
	}
	if h := s.result.Metrics.Hours; h.Estimated != estimated || h.Actual != actual {
		return fmt.Errorf("horas = %v estimadas y %v reales, se esperaba %v y %v", h.Estimated, h.Actual, estimated, actual)
	}
	return nil
}

func (s *sprintReportScenario) defectsAre(detected, resolved int) error {
	if err := s.metricsOfSprint(); err != nil {
		return err
	}
	if d := s.result.Metrics.Defects; d.Detected != detected || d.Resolved != resolved {
		return fmt.Errorf("defectos = %d detectados y %d resueltos, se esperaba %d y %d", d.Detected, d.Resolved, detected, resolved)
	}
	return nil
}

// notAvailableAre verifica que exactamente esas métricas del sprint estén no
// disponibles (RN-35.6): "porcentaje de completadas" y "desviación".
func (s *sprintReportScenario) notAvailableAre(list string) error {
	if err := s.metricsOfSprint(); err != nil {
		return err
	}
	want := map[string]bool{}
	for _, name := range strings.Split(list, ", ") {
		want[name] = true
	}
	m := s.result.Metrics
	got := map[string]bool{
		"porcentaje de completadas": !m.Completion.PercentAvailable,
		"desviación":                !m.Hours.DeviationAvailable,
	}
	for name, unavailable := range got {
		if unavailable != want[name] {
			return fmt.Errorf("%s no disponible = %v, se esperaba %v", name, unavailable, want[name])
		}
	}
	return nil
}

func (s *sprintReportScenario) effortsAre(table *godog.Table) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	var want, got []string
	for _, row := range table.Rows[1:] {
		hours, err := strconv.ParseFloat(row.Cells[3].Value, 64)
		if err != nil {
			return err
		}
		want = append(want, fmt.Sprintf("%d %d %s %v", s.sprints[row.Cells[0].Value].ID, s.items[row.Cells[1].Value].ID, row.Cells[2].Value, hours))
	}
	for _, ef := range s.result.Efforts.Efforts {
		got = append(got, fmt.Sprintf("%d %d %s %v", ef.SprintID, ef.ItemID, ef.Member, ef.Hours))
	}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("esfuerzos = %q, se esperaba %q", got, want)
	}
	return nil
}

func (s *sprintReportScenario) totalIs(total float64) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if s.result.Efforts.TotalHours != total {
		return fmt.Errorf("total = %v, se esperaba %v", s.result.Efforts.TotalHours, total)
	}
	return nil
}

// totalIsExactSum suma en float64 las horas guardadas del sprint, en orden,
// sin redondear (RN-39.4).
func (s *sprintReportScenario) totalIsExactSum() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	stored, _ := s.env.effortSpy.Repository.ListBySprint(s.sprintID)
	want := 0.0
	for _, ef := range stored {
		want += ef.Hours
	}
	if len(stored) == 0 || len(s.result.Efforts.Efforts) != len(stored) || s.result.Efforts.TotalHours != want {
		return fmt.Errorf("total = %v con %d esfuerzos, se esperaba la suma exacta %v de %d esfuerzos",
			s.result.Efforts.TotalHours, len(s.result.Efforts.Efforts), want, len(stored))
	}
	return nil
}

func (s *sprintReportScenario) totalEqualsActual() error {
	if err := s.metricsOfSprint(); err != nil {
		return err
	}
	if s.result.Efforts.TotalHours != s.result.Metrics.Hours.Actual {
		return fmt.Errorf("total = %v, se esperaba %v (horas reales)", s.result.Efforts.TotalHours, s.result.Metrics.Hours.Actual)
	}
	return nil
}

// defectsListIs compara los defectos con una tabla de item, estado y sprints
// de detección y resolución (vacío = ninguno), en orden.
func (s *sprintReportScenario) defectsListIs(table *godog.Table) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	var want, got []string
	for _, row := range table.Rows[1:] {
		resolution := 0
		if name := row.Cells[3].Value; name != "" {
			resolution = s.sprints[name].ID
		}
		want = append(want, fmt.Sprintf("%d %s %d %d", s.items[row.Cells[0].Value].ID, row.Cells[1].Value, s.sprints[row.Cells[2].Value].ID, resolution))
	}
	for _, d := range s.result.Defects.Defects {
		got = append(got, fmt.Sprintf("%d %s %d %d", d.ItemID, d.Status, d.DetectionSprintID, d.ResolutionSprintID))
	}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("defectos = %q, se esperaba %q", got, want)
	}
	return nil
}

// defectsMatchMetrics verifica que los defectos de la lista detectados en el
// sprint sean M9 y los RESUELTO en el sprint sean M10 (RN-39.5).
func (s *sprintReportScenario) defectsMatchMetrics() error {
	if err := s.metricsOfSprint(); err != nil {
		return err
	}
	if counts := sprintDefectCounts(s.result.Defects.Defects, s.sprintID); counts != s.result.Metrics.Defects {
		return fmt.Errorf("conteos de la lista = %+v, se esperaba M9 y M10 = %+v", counts, s.result.Metrics.Defects)
	}
	return nil
}

func (s *sprintReportScenario) membersAre(list string) error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if want := strings.Split(list, ", "); !reflect.DeepEqual(s.result.Members.Names, want) {
		return fmt.Errorf("integrantes = %q, se esperaba %q", s.result.Members.Names, want)
	}
	return nil
}

func (s *sprintReportScenario) noDefects() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if d := s.result.Defects.Defects; d == nil || len(d) != 0 {
		return fmt.Errorf("defectos = %#v, se esperaba una lista vacía", d)
	}
	return nil
}

func (s *sprintReportScenario) emptyParts() error {
	if err := s.noDefects(); err != nil {
		return err
	}
	r := s.result
	if r.Efforts.Efforts == nil || len(r.Efforts.Efforts) != 0 || r.Efforts.TotalHours != 0 {
		return fmt.Errorf("esfuerzos = %#v con total %v, se esperaba una lista vacía y total 0", r.Efforts.Efforts, r.Efforts.TotalHours)
	}
	if r.Members.Names == nil || len(r.Members.Names) != 0 {
		return fmt.Errorf("integrantes = %#v, se esperaba una lista vacía", r.Members.Names)
	}
	return nil
}

// equalToSources compara cada parte con su operación de origen (RN-39.9).
func (s *sprintReportScenario) equalToSources() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	e := s.env
	wantSprint, _ := e.sprints.Get(s.sprintID)
	wantMetrics, _ := e.metrics.Sprint(s.sprintID)
	wantEfforts, _ := e.efforts.ListBySprint(s.sprintID)
	wantDetected, _ := e.defects.ListBySprint(s.sprintID)
	r := s.result
	detected := []defect.Defect{}
	for _, d := range r.Defects.Defects {
		if d.DetectionSprintID == s.sprintID {
			detected = append(detected, d)
		}
	}
	if r.Sprint != wantSprint || r.Metrics != wantMetrics || !reflect.DeepEqual(r.Efforts.Efforts, wantEfforts.Efforts) || !reflect.DeepEqual(detected, wantDetected) {
		return errors.New("alguna parte del reporte no es igual a su consulta de origen")
	}
	return nil
}

func (s *sprintReportScenario) failsNotFound() error {
	if !errors.Is(s.err, sprint.ErrSprintNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, sprint.ErrSprintNotFound)
	}
	return nil
}

func (s *sprintReportScenario) onlySprintRead() error {
	e := s.env
	if e.sprintGetSpy.gets != 1 || e.itemSpy.calls != 0 || e.effortSpy.calls != 0 || e.defectSpy.calls != 0 {
		return fmt.Errorf("lecturas del sprint, ítems, esfuerzos y defectos = %d, %d, %d y %d; se esperaba 1, 0, 0 y 0",
			e.sprintGetSpy.gets, e.itemSpy.calls, e.effortSpy.calls, e.defectSpy.calls)
	}
	return nil
}

func (s *sprintReportScenario) failsWithReadError(source string) error {
	want := map[string]error{
		"el sprint":     errSprintRead,
		"las métricas":  errItemsRead,
		"los esfuerzos": errEffortsRead,
		"los defectos":  errDefectsRead,
	}[source]
	if !errors.Is(s.err, want) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, want)
	}
	if !reflect.DeepEqual(s.result, SprintReport{}) {
		return fmt.Errorf("reporte = %+v, se esperaba el valor cero", s.result)
	}
	return nil
}

// laterReadsSkipped verifica las lecturas de esfuerzos y defectos hechas
// hasta la que falla (RN-39.3).
func (s *sprintReportScenario) laterReadsSkipped(source string) error {
	want := map[string][2]int{
		"el sprint":     {0, 0},
		"las métricas":  {0, 0},
		"los esfuerzos": {2, 1},
		"los defectos":  {2, 2},
	}[source]
	e := s.env
	if got := [2]int{e.effortSpy.calls, e.defectSpy.calls}; got != want {
		return fmt.Errorf("lecturas de esfuerzos y defectos = %v, se esperaba %v", got, want)
	}
	return nil
}

func (s *sprintReportScenario) sameAgain() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	if s.result.Sprint.ID != s.sprintID || len(s.result.Efforts.Efforts) == 0 {
		return fmt.Errorf("reporte = %+v, se esperaban los datos del sprint %d y sus esfuerzos", s.result, s.sprintID)
	}
	again, err := s.env.reports.Sprint(s.sprintID)
	if err != nil || !reflect.DeepEqual(again, s.result) {
		return fmt.Errorf("la segunda consulta = %+v (%v), se esperaba %+v", again, err, s.result)
	}
	return nil
}

func (s *sprintReportScenario) resultIsACopy() error {
	if err := s.succeeded(); err != nil {
		return err
	}
	r := s.result
	if len(r.Efforts.Efforts) == 0 || len(r.Defects.Defects) == 0 || len(r.Members.Names) == 0 {
		return fmt.Errorf("reporte = %+v, se esperaban esfuerzos, defectos e integrantes", r)
	}
	changed, err := s.env.reports.Sprint(s.sprintID)
	if err != nil {
		return err
	}
	changed.Efforts.Efforts[0].Hours = 999
	changed.Defects.Defects[0].Status = "OTRO"
	changed.Members.Names[0] = "Otro"
	again, err := s.env.reports.Sprint(s.sprintID)
	if err != nil || !reflect.DeepEqual(again, r) {
		return fmt.Errorf("después de modificar el reporte = %+v (%v), se esperaba %+v", again, err, r)
	}
	return nil
}

func (s *sprintReportScenario) nothingChanged() error {
	if after := s.stored(); !reflect.DeepEqual(after, s.before) {
		return errors.New("el reporte modificó los ítems, los sprints, los esfuerzos o los defectos")
	}
	return nil
}
