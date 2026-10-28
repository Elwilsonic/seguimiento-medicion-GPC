package dashboard

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-36-resumen-proyecto.feature con
// godog.
func TestFeatures_HU36ResumenProyecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeSummaryScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-36-resumen-proyecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-36")
	}
}

// summaryScenario guarda el estado de un escenario de HU-36.
type summaryScenario struct {
	t        *testing.T
	env      *dashboardEnv
	projects map[string]int
	sprints  map[string]sprint.Sprint
	items    map[string]backlog.Item

	summary Summary
	err     error

	projectBefore project.Project
	itemsBefore   []backlog.Item
	sprintsBefore []sprint.Sprint
	projectID     int
	snapshotTaken bool
}

func initializeSummaryScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &summaryScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.env = newDashboardEnv()
		s.projects = map[string]int{}
		s.sprints = map[string]sprint.Sprint{}
		s.items = map[string]backlog.Item{}
		s.summary = Summary{}
		s.err = nil
		s.projectBefore, s.itemsBefore, s.sprintsBefore = project.Project{}, nil, nil
		s.projectID = 0
		s.snapshotTaken = false
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el proyecto "([^"]*)" está en estado "([^"]*)"$`, s.projectInStatus)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que existe un sprint cerrado "([^"]*)" en el proyecto "([^"]*)"$`, s.existingClosedSprint)
	ctx.Step(`^que el proyecto "([^"]*)" tiene los ítems:$`, s.projectItems)
	ctx.Step(`^(?:que )?consulto el resumen del proyecto "([^"]*)"$`, s.queryProjectByName)
	ctx.Step(`^consulto el resumen del proyecto con el ID (-?\d+)$`, s.queryProject)
	ctx.Step(`^el ítem "([^"]*)" pasa al estado "([^"]*)"$`, s.itemChangesStatus)
	ctx.Step(`^se cierra el sprint "([^"]*)"$`, s.closeSprint)
	ctx.Step(`^el resumen muestra el proyecto "([^"]*)" en estado "([^"]*)"$`, s.showsProject)
	ctx.Step(`^(?:que )?el resumen muestra (\d+) ítems pendientes, (\d+) en curso y (\d+) terminados$`, s.showsCounts)
	ctx.Step(`^(?:que )?el sprint actual del resumen es "([^"]*)"$`, s.currentSprintIs)
	ctx.Step(`^el resumen no tiene sprint actual$`, s.noCurrentSprint)
	ctx.Step(`^la consulta falla porque el proyecto no existe$`, s.failsProjectNotFound)
	ctx.Step(`^el proyecto, los ítems y los sprints no cambiaron$`, s.nothingChanged)
}

func (s *summaryScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name)
	if s.projectID == 0 {
		s.projectID = s.projects[name]
	}
	return nil
}

func (s *summaryScenario) projectInStatus(name, status string) error {
	s.env.setProjectStatus(s.t, s.projects[name], project.Status(status))
	return nil
}

func (s *summaryScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	return nil
}

func (s *summaryScenario) existingClosedSprint(name, projectName string) error {
	s.sprints[name] = s.env.closedSprint(s.t, s.projects[projectName], name)
	return nil
}

// projectItems crea los ítems de la tabla (titulo, estado, sprint); un sprint
// vacío deja el ítem sin asignar.
func (s *summaryScenario) projectItems(projectName string, table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		title, status, sprintName := row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value
		sprintID := 0
		if sprintName != "" {
			sp, ok := s.sprints[sprintName]
			if !ok {
				return fmt.Errorf("el sprint %q no existe en el escenario", sprintName)
			}
			sprintID = sp.ID
		}
		s.items[title] = s.env.item(s.t, s.projects[projectName], title, backlog.Status(status), sprintID)
	}
	return nil
}

// snapshot guarda, antes de la primera consulta, el proyecto, los ítems y
// los sprints del proyecto del escenario (CA-36.5).
func (s *summaryScenario) snapshot() {
	if s.snapshotTaken || s.projectID == 0 {
		return
	}
	s.projectBefore, s.itemsBefore, s.sprintsBefore = s.env.snapshot(s.t, s.projectID)
	s.snapshotTaken = true
}

func (s *summaryScenario) queryProjectByName(name string) error {
	return s.queryProject(s.projects[name])
}

func (s *summaryScenario) queryProject(id int) error {
	s.snapshot()
	s.summary, s.err = s.env.dashboard.ProjectSummary(id)
	return nil
}

func (s *summaryScenario) itemChangesStatus(title, status string) error {
	it, err := s.env.items.GetItem(s.items[title].ID)
	if err != nil {
		return err
	}
	s.items[title] = s.env.changeStatus(s.t, it, backlog.Status(status))
	return nil
}

func (s *summaryScenario) closeSprint(name string) error {
	s.sprints[name] = s.env.closeSprint(s.t, s.sprints[name].ID)
	return nil
}

func (s *summaryScenario) noError() error {
	if s.err != nil {
		return fmt.Errorf("la consulta devolvió un error inesperado: %w", s.err)
	}
	return nil
}

func (s *summaryScenario) showsProject(name, status string) error {
	if err := s.noError(); err != nil {
		return err
	}
	if s.summary.ProjectID != s.projects[name] || s.summary.ProjectName != name || s.summary.ProjectStatus != project.Status(status) {
		return fmt.Errorf("proyecto = %d %q %s, se esperaba %d %q %s",
			s.summary.ProjectID, s.summary.ProjectName, s.summary.ProjectStatus, s.projects[name], name, status)
	}
	return nil
}

func (s *summaryScenario) showsCounts(pending, inProgress, finished int) error {
	if err := s.noError(); err != nil {
		return err
	}
	got := s.summary
	if got.Pending != pending || got.InProgress != inProgress || got.Finished != finished {
		return fmt.Errorf("ítems = %d pendientes, %d en curso y %d terminados; se esperaba %d, %d y %d",
			got.Pending, got.InProgress, got.Finished, pending, inProgress, finished)
	}
	return nil
}

func (s *summaryScenario) currentSprintIs(name string) error {
	if err := s.noError(); err != nil {
		return err
	}
	if !s.summary.HasCurrentSprint || !reflect.DeepEqual(s.summary.CurrentSprint, s.sprints[name]) {
		return fmt.Errorf("sprint actual = %+v (hay: %v), se esperaba %+v",
			s.summary.CurrentSprint, s.summary.HasCurrentSprint, s.sprints[name])
	}
	return nil
}

func (s *summaryScenario) noCurrentSprint() error {
	if err := s.noError(); err != nil {
		return err
	}
	if s.summary.HasCurrentSprint || s.summary.CurrentSprint != (sprint.Sprint{}) {
		return fmt.Errorf("sprint actual = %+v (hay: %v), se esperaba ninguno",
			s.summary.CurrentSprint, s.summary.HasCurrentSprint)
	}
	return nil
}

func (s *summaryScenario) failsProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("error = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	if s.summary != (Summary{}) {
		return fmt.Errorf("resumen = %+v, se esperaba vacío", s.summary)
	}
	return nil
}

func (s *summaryScenario) nothingChanged() error {
	if err := s.noError(); err != nil {
		return err
	}
	projectAfter, itemsAfter, sprintsAfter := s.env.snapshot(s.t, s.projectID)
	if projectAfter != s.projectBefore || !reflect.DeepEqual(itemsAfter, s.itemsBefore) || !reflect.DeepEqual(sprintsAfter, s.sprintsBefore) {
		return errors.New("consultar el resumen modificó el proyecto, los ítems o los sprints")
	}
	return nil
}
