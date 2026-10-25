package metrics

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-32-sp-velocidad.feature con
// godog.
func TestFeatures_HU32CalcularStoryPoints(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeStoryPointsScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-32-sp-velocidad.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-32")
	}
}

// storyPointsScenario guarda el estado de un escenario de HU-32.
type storyPointsScenario struct {
	t        *testing.T
	env      *metricsEnv
	projects map[string]int
	sprints  map[string]sprint.Sprint

	sprintPoints  SprintPoints
	projectPoints ProjectPoints
	err           error

	itemsBefore   []backlog.Item
	sprintsBefore []sprint.Sprint
	projectID     int
}

func initializeStoryPointsScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &storyPointsScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.env = newMetricsEnv()
		s.projects = map[string]int{}
		s.sprints = map[string]sprint.Sprint{}
		s.sprintPoints = SprintPoints{}
		s.projectPoints = ProjectPoints{}
		s.err = nil
		s.itemsBefore, s.sprintsBefore = nil, nil
		s.projectID = 0
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que el sprint "([^"]*)" tiene los ítems:$`, s.sprintItems)
	ctx.Step(`^que el proyecto "([^"]*)" tiene los ítems sin sprint:$`, s.projectItems)
	ctx.Step(`^que el proyecto "([^"]*)" tiene sprints cerrados con story points completados "([^"]*)"$`, s.closedSprints)
	ctx.Step(`^consulto los story points del sprint "([^"]*)"$`, s.querySprintByName)
	ctx.Step(`^consulto los story points del sprint con el ID (-?\d+)$`, s.querySprint)
	ctx.Step(`^consulto los story points del proyecto "([^"]*)"$`, s.queryProjectByName)
	ctx.Step(`^consulto los story points del proyecto con el ID (-?\d+)$`, s.queryProject)
	ctx.Step(`^los story points planificados son (\d+)$`, s.plannedAre)
	ctx.Step(`^los story points completados son (\d+)$`, s.completedAre)
	ctx.Step(`^hay (\d+) ítems sin estimar$`, s.unestimatedAre)
	ctx.Step(`^la velocidad es ([\d.]+) con (\d+) sprints cerrados$`, s.velocityIs)
	ctx.Step(`^la velocidad no está disponible$`, s.velocityNotAvailable)
	ctx.Step(`^la consulta falla porque el sprint no existe$`, s.failsSprintNotFound)
	ctx.Step(`^la consulta falla porque el proyecto no existe$`, s.failsProjectNotFound)
	ctx.Step(`^los ítems y los sprints del proyecto no cambiaron$`, s.nothingChanged)
}

func (s *storyPointsScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name)
	return nil
}

func (s *storyPointsScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	return nil
}

// addItems crea los ítems de la tabla (titulo, story_points, estado).
func (s *storyPointsScenario) addItems(projectID, sprintID int, table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		points, err := strconv.Atoi(row.Cells[1].Value)
		if err != nil {
			return fmt.Errorf("story points inválidos %q: %w", row.Cells[1].Value, err)
		}
		s.env.item(s.t, projectID, row.Cells[0].Value, points, backlog.Status(row.Cells[2].Value), sprintID)
	}
	return nil
}

func (s *storyPointsScenario) sprintItems(name string, table *godog.Table) error {
	sp := s.sprints[name]
	s.projectID = sp.ProjectID
	return s.addItems(sp.ProjectID, sp.ID, table)
}

func (s *storyPointsScenario) projectItems(projectName string, table *godog.Table) error {
	return s.addItems(s.projects[projectName], 0, table)
}

// closedSprints crea un sprint cerrado por cada parte separada por ";". Cada
// parte lista los story points de sus ítems terminados separados por "+", o
// "-" para un sprint sin ítems.
func (s *storyPointsScenario) closedSprints(projectName, spec string) error {
	projectID := s.projects[projectName]
	for i, part := range strings.Split(spec, ";") {
		part = strings.TrimSpace(part)
		var points []int
		if part != "-" {
			for _, p := range strings.Split(part, "+") {
				n, err := strconv.Atoi(strings.TrimSpace(p))
				if err != nil {
					return fmt.Errorf("story points inválidos %q: %w", p, err)
				}
				points = append(points, n)
			}
		}
		s.env.closedSprint(s.t, projectID, "Sprint "+strconv.Itoa(i+1), points...)
	}
	return nil
}

// snapshot guarda, antes de la primera consulta, los ítems y sprints del
// proyecto del escenario para comparar después (CA-32.7). Lee los
// repositorios en memoria directamente, así no depende de lo que se prueba.
func (s *storyPointsScenario) snapshot() {
	if s.itemsBefore != nil || s.projectID == 0 {
		return
	}
	s.itemsBefore, _ = s.env.itemRepo.ListByProject(s.projectID)
	s.sprintsBefore = s.env.sprintsOf(s.projectID)
}

func (s *storyPointsScenario) querySprintByName(name string) error {
	return s.querySprint(s.sprints[name].ID)
}

func (s *storyPointsScenario) querySprint(id int) error {
	s.snapshot()
	s.sprintPoints, s.err = s.env.metrics.SprintStoryPoints(id)
	return nil
}

func (s *storyPointsScenario) queryProjectByName(name string) error {
	return s.queryProject(s.projects[name])
}

func (s *storyPointsScenario) queryProject(id int) error {
	s.snapshot()
	s.projectPoints, s.err = s.env.metrics.ProjectStoryPoints(id)
	return nil
}

// result devuelve planificados, completados y sin estimar de la última
// consulta (de sprint o de proyecto).
func (s *storyPointsScenario) result() (int, int, int, error) {
	if s.err != nil {
		return 0, 0, 0, fmt.Errorf("la consulta falló: %v", s.err)
	}
	if s.projectPoints != (ProjectPoints{}) {
		return s.projectPoints.Planned, s.projectPoints.Completed, s.projectPoints.Unestimated, nil
	}
	return s.sprintPoints.Planned, s.sprintPoints.Completed, s.sprintPoints.Unestimated, nil
}

func (s *storyPointsScenario) plannedAre(want int) error {
	got, _, _, err := s.result()
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("planificados = %d, se esperaba %d", got, want)
	}
	return nil
}

func (s *storyPointsScenario) completedAre(want int) error {
	_, got, _, err := s.result()
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("completados = %d, se esperaba %d", got, want)
	}
	return nil
}

func (s *storyPointsScenario) unestimatedAre(want int) error {
	_, _, got, err := s.result()
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("ítems sin estimar = %d, se esperaba %d", got, want)
	}
	return nil
}

func (s *storyPointsScenario) velocityIs(want float64, used int) error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	p := s.projectPoints
	if !p.VelocityAvailable || p.Velocity != want || p.ClosedSprints != used {
		return fmt.Errorf("velocidad = %v (disponible %v) con %d sprints, se esperaba %v con %d", p.Velocity, p.VelocityAvailable, p.ClosedSprints, want, used)
	}
	return nil
}

func (s *storyPointsScenario) velocityNotAvailable() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	p := s.projectPoints
	if p.ProjectID == 0 || p.VelocityAvailable || p.Velocity != 0 || p.ClosedSprints != 0 {
		return fmt.Errorf("velocidad = %v (disponible %v) con %d sprints, se esperaba no disponible", p.Velocity, p.VelocityAvailable, p.ClosedSprints)
	}
	return nil
}

func (s *storyPointsScenario) failsSprintNotFound() error {
	if !errors.Is(s.err, sprint.ErrSprintNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, sprint.ErrSprintNotFound)
	}
	return nil
}

func (s *storyPointsScenario) failsProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	return nil
}

func (s *storyPointsScenario) nothingChanged() error {
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
