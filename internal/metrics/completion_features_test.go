package metrics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-34-porcentaje-completadas.feature
// con godog.
func TestFeatures_HU34PorcentajeCompletadas(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeCompletionScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-34-porcentaje-completadas.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-34")
	}
}

// completionScenario guarda el estado de un escenario de HU-34.
type completionScenario struct {
	t        *testing.T
	env      *metricsEnv
	projects map[string]int
	sprints  map[string]sprint.Sprint

	completion Completion
	err        error

	itemsBefore   []backlog.Item
	sprintsBefore []sprint.Sprint
	projectID     int
}

func initializeCompletionScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &completionScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.env = newMetricsEnv()
		s.projects = map[string]int{}
		s.sprints = map[string]sprint.Sprint{}
		s.completion = Completion{}
		s.err = nil
		s.itemsBefore, s.sprintsBefore = nil, nil
		s.projectID = 0
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que el sprint "([^"]*)" tiene los ítems:$`, s.sprintItems)
	ctx.Step(`^que el proyecto "([^"]*)" tiene los ítems sin sprint:$`, s.projectItems)
	ctx.Step(`^consulto el porcentaje de historias completadas del sprint "([^"]*)"$`, s.querySprintByName)
	ctx.Step(`^consulto el porcentaje de historias completadas del sprint con el ID (-?\d+)$`, s.querySprint)
	ctx.Step(`^consulto el porcentaje de historias completadas del proyecto "([^"]*)"$`, s.queryProjectByName)
	ctx.Step(`^consulto el porcentaje de historias completadas del proyecto con el ID (-?\d+)$`, s.queryProject)
	ctx.Step(`^el porcentaje de historias completadas es ([\d.]+) con (\d+) de (\d+) ítems terminados$`, s.percentIs)
	ctx.Step(`^el porcentaje de historias completadas no está disponible$`, s.percentNotAvailable)
	ctx.Step(`^la consulta falla porque el sprint no existe$`, s.failsSprintNotFound)
	ctx.Step(`^la consulta falla porque el proyecto no existe$`, s.failsProjectNotFound)
	ctx.Step(`^los ítems y los sprints del proyecto no cambiaron$`, s.nothingChanged)
}

func (s *completionScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name)
	s.projectID = s.projects[name]
	return nil
}

func (s *completionScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	return nil
}

// addItems crea los ítems de la tabla (titulo, story_points, estado).
func (s *completionScenario) addItems(projectID, sprintID int, table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		points, err := strconv.Atoi(row.Cells[1].Value)
		if err != nil {
			return fmt.Errorf("story points inválidos %q: %w", row.Cells[1].Value, err)
		}
		s.env.item(s.t, projectID, row.Cells[0].Value, points, backlog.Status(row.Cells[2].Value), sprintID)
	}
	return nil
}

func (s *completionScenario) sprintItems(name string, table *godog.Table) error {
	sp := s.sprints[name]
	return s.addItems(sp.ProjectID, sp.ID, table)
}

func (s *completionScenario) projectItems(projectName string, table *godog.Table) error {
	return s.addItems(s.projects[projectName], 0, table)
}

// snapshot guarda, antes de la primera consulta, los ítems y sprints del
// proyecto del escenario para comparar después (CA-34.5). Lee los
// repositorios en memoria directamente, así no depende de lo que se prueba.
func (s *completionScenario) snapshot() {
	if s.itemsBefore != nil || s.projectID == 0 {
		return
	}
	s.itemsBefore, _ = s.env.itemRepo.ListByProject(s.projectID)
	s.sprintsBefore = s.env.sprintsOf(s.projectID)
}

func (s *completionScenario) querySprintByName(name string) error {
	return s.querySprint(s.sprints[name].ID)
}

func (s *completionScenario) querySprint(id int) error {
	s.snapshot()
	s.completion, s.err = s.env.metrics.SprintCompletion(id)
	return nil
}

func (s *completionScenario) queryProjectByName(name string) error {
	return s.queryProject(s.projects[name])
}

func (s *completionScenario) queryProject(id int) error {
	s.snapshot()
	s.completion, s.err = s.env.metrics.ProjectCompletion(id)
	return nil
}

func (s *completionScenario) percentIs(want float64, finished, items int) error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	c := s.completion
	if c.Items != items || c.Finished != finished || !c.PercentAvailable || math.Abs(c.Percent-want) > 1e-9 {
		return fmt.Errorf("completion = %+v, se esperaba %v %% con %d de %d ítems terminados", c, want, finished, items)
	}
	return nil
}

func (s *completionScenario) percentNotAvailable() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	if s.completion != (Completion{}) {
		return fmt.Errorf("completion = %+v, se esperaba 0 ítems y el porcentaje no disponible", s.completion)
	}
	return nil
}

func (s *completionScenario) failsSprintNotFound() error {
	if !errors.Is(s.err, sprint.ErrSprintNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, sprint.ErrSprintNotFound)
	}
	return nil
}

func (s *completionScenario) failsProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	return nil
}

func (s *completionScenario) nothingChanged() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	items, err := s.env.itemRepo.ListByProject(s.projectID)
	if err != nil {
		return err
	}
	sprints := s.env.sprintsOf(s.projectID)
	if s.itemsBefore == nil || !reflect.DeepEqual(items, s.itemsBefore) || !reflect.DeepEqual(sprints, s.sprintsBefore) {
		return fmt.Errorf("la consulta modificó los datos: ítems %+v, sprints %+v", items, sprints)
	}
	return nil
}
