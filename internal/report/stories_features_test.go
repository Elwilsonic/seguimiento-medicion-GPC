package report

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-49-historias-en-reporte.feature
// con godog.
func TestFeatures_HU49HistoriasEnReporte(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeStoriesScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-49-historias-en-reporte.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-49")
	}
}

// storiesScenario guarda el estado de un escenario de HU-49. Usa el entorno
// del reporte del sprint, que también arma el del proyecto.
type storiesScenario struct {
	t        *testing.T
	env      *sprintReportEnv
	projects map[string]int
	sprints  map[string]sprint.Sprint

	// failStories indica que tiene que fallar la lectura de las historias.
	failStories bool
	// readsBefore son las lecturas que el reporte hace antes de las
	// historias, contadas cuando failStories es true.
	readsBefore readsBeforeStories

	projectID int
	// kind es "proyecto" o "sprint": el reporte que se generó.
	kind          string
	projectReport ProjectReport
	sprintReport  SprintReport
	err           error
	before        []backlog.Item
}

func initializeStoriesScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &storiesScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*s = storiesScenario{
			t:        t,
			env:      newSprintReportEnv(),
			projects: map[string]int{},
			sprints:  map[string]sprint.Sprint{},
		}
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que cierro el sprint "([^"]*)"$`, s.closeSprint)
	ctx.Step(`^que el proyecto "([^"]*)" tiene las historias:$`, s.storiesTable)
	ctx.Step(`^que falla la lectura de las historias del reporte$`, s.failRead)
	ctx.Step(`^genero el reporte del proyecto "([^"]*)"$`, s.generateProjectByName)
	ctx.Step(`^genero el reporte del proyecto con el ID (-?\d+)$`, s.generateProject)
	ctx.Step(`^genero el reporte del sprint "([^"]*)"$`, s.generateSprintByName)
	ctx.Step(`^las historias del reporte son:$`, s.storiesAre)
	ctx.Step(`^el reporte no tiene historias$`, s.noStories)
	ctx.Step(`^la cantidad de historias es (\d+) y las completadas son (\d+)$`, s.countsAre)
	ctx.Step(`^las cantidades de historias coinciden con las métricas del reporte$`, s.countsMatchMetrics)
	ctx.Step(`^cada historia del reporte es igual a su consulta en el backlog$`, s.equalToBacklog)
	ctx.Step(`^la sección "Historias" del PDF del reporte, entre "Métricas" y "Esfuerzos", es:$`, s.pdfSectionIs)
	ctx.Step(`^el reporte falla con el error de la lectura de las historias$`, s.failsWithReadError)
	ctx.Step(`^las historias se leyeron después de los defectos$`, s.storiesReadAfterDefects)
	ctx.Step(`^el reporte falla porque el proyecto no existe$`, s.failsNotFound)
	ctx.Step(`^no se leyeron las historias$`, s.storiesNotRead)
	ctx.Step(`^modificar las historias del reporte no cambia el próximo reporte$`, s.resultIsACopy)
	ctx.Step(`^las historias guardadas no cambiaron$`, s.nothingChanged)
}

func (s *storiesScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name, "")
	if s.projectID == 0 {
		s.projectID = s.projects[name]
	}
	return nil
}

func (s *storiesScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	return nil
}

func (s *storiesScenario) closeSprint(name string) error {
	s.env.closeSprint(s.t, s.sprints[name].ID)
	return nil
}

// storiesTable crea las historias de una tabla con titulo, prioridad, sp,
// horas, estado y sprint (vacío = sin sprint).
func (s *storiesScenario) storiesTable(projectName string, table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		points, err := strconv.Atoi(row.Cells[2].Value)
		if err != nil {
			return err
		}
		hours, err := strconv.ParseFloat(row.Cells[3].Value, 64)
		if err != nil {
			return err
		}
		sprintID := 0
		if name := row.Cells[5].Value; name != "" {
			sp, ok := s.sprints[name]
			if !ok {
				return fmt.Errorf("no existe el sprint %q", name)
			}
			sprintID = sp.ID
		}
		s.env.story(s.t, s.projects[projectName], row.Cells[0].Value, backlog.Priority(row.Cells[1].Value),
			points, hours, backlog.Status(row.Cells[4].Value), sprintID)
	}
	return nil
}

func (s *storiesScenario) failRead() error {
	s.failStories = true
	return nil
}

// stored devuelve las historias guardadas del proyecto, sin pasar por los
// contadores.
func (s *storiesScenario) stored() []backlog.Item {
	items, err := s.env.itemSpy.Repository.ListByProject(s.projectID)
	if err != nil {
		s.t.Fatalf("leer las historias guardadas: %v", err)
	}
	return items
}

func (s *storiesScenario) generateProjectByName(name string) error {
	return s.generateProject(s.projects[name])
}

func (s *storiesScenario) generateProject(projectID int) error {
	s.kind = "proyecto"
	s.before = s.stored()
	if s.failStories {
		s.readsBefore = s.env.projectReadsBeforeStories(s.t, projectID)
		s.env.itemSpy.failAt = s.readsBefore.items + 1
	}
	s.env.resetSprintReads()
	s.projectReport, s.err = s.env.reports.Project(projectID)
	return nil
}

func (s *storiesScenario) generateSprintByName(name string) error {
	s.kind = "sprint"
	s.before = s.stored()
	sp := s.sprints[name]
	if s.failStories {
		s.readsBefore = s.env.sprintReadsBeforeStories(s.t, sp.ID)
		s.env.itemSpy.failAt = s.readsBefore.items + 1
	}
	s.env.resetSprintReads()
	s.sprintReport, s.err = s.env.reports.Sprint(sp.ID)
	return nil
}

// stories devuelve la parte de las historias del reporte generado.
func (s *storiesScenario) stories() (Stories, error) {
	if s.err != nil {
		return Stories{}, fmt.Errorf("el reporte devolvió un error inesperado: %w", s.err)
	}
	if s.kind == "sprint" {
		return s.sprintReport.Stories, nil
	}
	return s.projectReport.Stories, nil
}

// sprintName devuelve el nombre del sprint con ese ID, o "" si es 0.
func (s *storiesScenario) sprintName(id int) string {
	for name, sp := range s.sprints {
		if sp.ID == id {
			return name
		}
	}
	return ""
}

func (s *storiesScenario) storiesAre(table *godog.Table) error {
	got, err := s.stories()
	if err != nil {
		return err
	}
	var want, have [][]string
	for _, row := range table.Rows[1:] {
		var cells []string
		for _, c := range row.Cells {
			cells = append(cells, c.Value)
		}
		want = append(want, cells)
	}
	for _, it := range got.Stories {
		have = append(have, []string{it.Title, string(it.Priority), string(it.Status), strconv.Itoa(it.StoryPoints),
			strconv.FormatFloat(it.EstimatedHours, 'f', -1, 64), s.sprintName(it.SprintID)})
	}
	if !reflect.DeepEqual(have, want) {
		return fmt.Errorf("historias = %v, se esperaba %v", have, want)
	}
	// RN-49.6: ordenadas por ID.
	for i := 1; i < len(got.Stories); i++ {
		if got.Stories[i-1].ID >= got.Stories[i].ID {
			return fmt.Errorf("las historias no están ordenadas por ID: %v", got.Stories)
		}
	}
	return nil
}

func (s *storiesScenario) noStories() error {
	got, err := s.stories()
	if err != nil {
		return err
	}
	if got.Stories == nil || len(got.Stories) != 0 {
		return fmt.Errorf("historias = %#v, se esperaba una lista vacía, no nil", got.Stories)
	}
	return nil
}

func (s *storiesScenario) countsAre(count, completed int) error {
	got, err := s.stories()
	if err != nil {
		return err
	}
	if len(got.Stories) != count || got.Completed != completed {
		return fmt.Errorf("cantidad = %d y completadas = %d, se esperaba %d y %d", len(got.Stories), got.Completed, count, completed)
	}
	return nil
}

// countsMatchMetrics compara con M8 del mismo reporte (RN-49.5).
func (s *storiesScenario) countsMatchMetrics() error {
	got, err := s.stories()
	if err != nil {
		return err
	}
	completion := s.projectReport.Metrics.Completion
	if s.kind == "sprint" {
		completion = s.sprintReport.Metrics.Completion
	}
	if len(got.Stories) != completion.Items || got.Completed != completion.Finished {
		return fmt.Errorf("cantidad = %d y completadas = %d, las métricas dicen %d y %d",
			len(got.Stories), got.Completed, completion.Items, completion.Finished)
	}
	return nil
}

func (s *storiesScenario) equalToBacklog() error {
	got, err := s.stories()
	if err != nil {
		return err
	}
	if len(got.Stories) == 0 {
		return errors.New("el reporte no tiene historias para comparar")
	}
	for _, it := range got.Stories {
		want, err := s.env.items.GetItem(it.ID)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(it, want) {
			return fmt.Errorf("historia %d = %+v, el backlog devuelve %+v", it.ID, it, want)
		}
	}
	return nil
}

// pdfSectionIs compara las líneas entre el título "Historias" y el siguiente
// título, y verifica que ese título vaya después de "Métricas" y antes de
// "Esfuerzos".
func (s *storiesScenario) pdfSectionIs(table *godog.Table) error {
	if s.err != nil {
		return fmt.Errorf("el reporte devolvió un error inesperado: %w", s.err)
	}
	lines := projectReportLines(s.projectReport)
	var headings []string
	start := -1
	for i, l := range lines {
		if l.Heading {
			headings = append(headings, l.Text)
			if l.Text == "Historias" {
				start = i
			}
		}
	}
	if start < 0 {
		return fmt.Errorf("el PDF no tiene la sección \"Historias\": títulos %v", headings)
	}
	var have []string
	end := start + 1
	for ; end < len(lines) && !lines[end].Heading; end++ {
		have = append(have, lines[end].Text)
	}
	if end == len(lines) || lines[end].Text != "Esfuerzos" || !followsHeading(lines, start, "Métricas") {
		return fmt.Errorf("la sección \"Historias\" no está entre \"Métricas\" y \"Esfuerzos\": títulos %v", headings)
	}
	var want []string
	for _, row := range table.Rows[1:] {
		want = append(want, row.Cells[0].Value)
	}
	if !reflect.DeepEqual(have, want) {
		return fmt.Errorf("sección \"Historias\" = %q, se esperaba %q", have, want)
	}
	return nil
}

// followsHeading indica si el último título antes de la línea i es heading.
func followsHeading(lines []pdfLine, i int, heading string) bool {
	for j := i - 1; j >= 0; j-- {
		if lines[j].Heading {
			return lines[j].Text == heading
		}
	}
	return false
}

// failsWithReadError verifica el error y que no haya reporte parcial
// (RN-49.7).
func (s *storiesScenario) failsWithReadError() error {
	if !errors.Is(s.err, errItemsRead) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, errItemsRead)
	}
	if !reflect.DeepEqual(s.projectReport, ProjectReport{}) || !reflect.DeepEqual(s.sprintReport, SprintReport{}) {
		return errors.New("se devolvió un reporte parcial")
	}
	return nil
}

// storiesReadAfterDefects verifica que, al fallar las historias, ya se
// hicieron todas las lecturas de esfuerzos y defectos del reporte (RN-49.7).
func (s *storiesScenario) storiesReadAfterDefects() error {
	if s.env.effortSpy.calls != s.readsBefore.efforts || s.env.defectSpy.calls != s.readsBefore.defects {
		return fmt.Errorf("lecturas de esfuerzos y defectos = %d y %d, se esperaba %d y %d antes de las historias",
			s.env.effortSpy.calls, s.env.defectSpy.calls, s.readsBefore.efforts, s.readsBefore.defects)
	}
	return nil
}

func (s *storiesScenario) failsNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	return nil
}

func (s *storiesScenario) storiesNotRead() error {
	if s.env.itemSpy.calls != 0 {
		return fmt.Errorf("lecturas de ítems = %d, se esperaba 0", s.env.itemSpy.calls)
	}
	return nil
}

// resultIsACopy modifica el título y los criterios de cada historia del
// reporte y verifica que el próximo reporte no cambie (RN-49.8).
func (s *storiesScenario) resultIsACopy() error {
	got, err := s.stories()
	if err != nil {
		return err
	}
	if len(got.Stories) == 0 {
		return errors.New("el reporte no tiene historias para modificar")
	}
	for i := range got.Stories {
		got.Stories[i].Title = "modificado"
		got.Stories[i].AcceptanceCriteria[0] = "modificado"
	}
	got.Stories[0] = backlog.Item{}
	next, err := s.env.reports.Project(s.projectID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(next.Stories.Stories, s.before) {
		return fmt.Errorf("el próximo reporte tiene %+v, se esperaba %+v", next.Stories.Stories, s.before)
	}
	return nil
}

func (s *storiesScenario) nothingChanged() error {
	if after := s.stored(); !reflect.DeepEqual(after, s.before) {
		return fmt.Errorf("historias guardadas = %+v, antes eran %+v", after, s.before)
	}
	return nil
}
