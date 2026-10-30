package sprint

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-14-sprints-anteriores.feature con
// godog. Los pasos viven en el paquete sprint porque History es una operación
// del Manager de sprints (ver SDD).
func TestFeatures_HU14SprintsAnteriores(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeHistoryScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-14-sprints-anteriores.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-14")
	}
}

// historyScenario guarda el estado de un escenario de HU-14.
type historyScenario struct {
	projects *project.Manager
	items    *backlog.Manager
	itemRepo *backlog.MemoryRepository
	repo     *historyRepository
	manager  *Manager

	projectIDs   map[string]int
	sprints      []Sprint
	historyErr   error
	completed    []backlog.Item
	completedErr error
}

func initializeHistoryScenario(ctx *godog.ScenarioContext) {
	s := &historyScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.itemRepo = backlog.NewMemoryRepository()
		s.items = backlog.NewManager(s.itemRepo, s.projects)
		s.repo = &historyRepository{Repository: NewMemoryRepository()}
		s.manager = NewManager(s.repo, s.projects, s.items)
		s.projectIDs = map[string]int{}
		s.sprints = nil
		s.historyErr = nil
		s.completed = nil
		s.completedErr = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un sprint (abierto|cerrado) "([^"]*)" del "([^"]*)" al "([^"]*)" con sprint goal "([^"]*)" en el proyecto "([^"]*)"$`, s.existingSprint)
	ctx.Step(`^que existe un ítem "([^"]*)" en estado "([^"]*)" asignado al sprint con el ID (-?\d+) en el proyecto "([^"]*)"$`, s.existingItem)
	ctx.Step(`^que el repositorio de sprints falla al listar$`, s.repoFailsOnList)
	ctx.Step(`^consulto los sprints del proyecto "([^"]*)"$`, s.historyByName)
	ctx.Step(`^consulto los sprints del proyecto con el ID (-?\d+)$`, s.history)
	ctx.Step(`^consulto los ítems completados del primer sprint cerrado de la lista$`, s.completedOfFirstClosed)
	ctx.Step(`^cambio el nombre y el estado del primer sprint devuelto$`, s.mutateFirstSprint)
	ctx.Step(`^la consulta es exitosa$`, s.historyOK)
	ctx.Step(`^obtengo estos sprints:$`, s.sprintsAre)
	ctx.Step(`^obtengo una lista vacía de sprints$`, s.sprintsEmpty)
	ctx.Step(`^la consulta de ítems completados es exitosa$`, s.completedOK)
	ctx.Step(`^obtengo los ítems completados "([^"]*)" y "([^"]*)"$`, s.completedAre)
	ctx.Step(`^obtengo una lista vacía de ítems completados$`, s.completedEmpty)
	ctx.Step(`^la consulta falla porque no existe un proyecto con ese id$`, s.historyFailsProjectNotFound)
	ctx.Step(`^la consulta falla con el error del repositorio$`, s.historyFailsRepository)
}

func (s *historyScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *historyScenario) projectID(projectName string) (int, error) {
	id, ok := s.projectIDs[projectName]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	return id, nil
}

func (s *historyScenario) existingSprint(kind, name, startRaw, endRaw, goal, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	start, err := time.Parse(time.DateOnly, startRaw)
	if err != nil {
		return err
	}
	end, err := time.Parse(time.DateOnly, endRaw)
	if err != nil {
		return err
	}
	status := StatusOpen
	if kind == "cerrado" {
		status = StatusClosed
	}
	_, err = s.repo.Save(Sprint{
		ProjectID: projectID,
		Name:      name,
		StartDate: start,
		EndDate:   end,
		Goal:      goal,
		Status:    status,
	})
	return err
}

func (s *historyScenario) existingItem(title, status string, sprintID int, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	_, err = s.itemRepo.Save(backlog.Item{
		ProjectID:          projectID,
		Title:              title,
		Description:        "Formulario de acceso",
		Priority:           backlog.PriorityHigh,
		AcceptanceCriteria: []string{"Login OK"},
		Status:             backlog.Status(status),
		StoryPoints:        3,
		SprintID:           sprintID,
	})
	return err
}

func (s *historyScenario) repoFailsOnList() error {
	s.repo.listErr = errRepository
	return nil
}

func (s *historyScenario) historyByName(projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	return s.history(projectID)
}

func (s *historyScenario) history(projectID int) error {
	s.sprints, s.historyErr = s.manager.History(projectID)
	return nil
}

func (s *historyScenario) completedOfFirstClosed() error {
	if s.historyErr != nil {
		return fmt.Errorf("la consulta de sprints falló: %v", s.historyErr)
	}
	for _, sp := range s.sprints {
		if sp.Status == StatusClosed {
			s.completed, s.completedErr = s.manager.ListCompletedItems(sp.ID)
			return nil
		}
	}
	return errors.New("la lista no tiene ningún sprint cerrado")
}

func (s *historyScenario) mutateFirstSprint() error {
	if len(s.sprints) == 0 {
		return errors.New("la lista de sprints está vacía")
	}
	s.sprints[0].Name = "Cambiado"
	s.sprints[0].Status = StatusClosed
	return nil
}

func (s *historyScenario) historyOK() error {
	if s.historyErr != nil {
		return fmt.Errorf("se esperaba una consulta exitosa, falló con: %v", s.historyErr)
	}
	return nil
}

func (s *historyScenario) sprintsAre(table *godog.Table) error {
	rows := table.Rows[1:]
	if len(s.sprints) != len(rows) {
		return fmt.Errorf("se obtuvieron %d sprints, se esperaban %d: %+v", len(s.sprints), len(rows), s.sprints)
	}
	for i, row := range rows {
		id, err := strconv.Atoi(row.Cells[0].Value)
		if err != nil {
			return err
		}
		got := s.sprints[i]
		want := []string{row.Cells[1].Value, row.Cells[2].Value, row.Cells[3].Value, row.Cells[4].Value, row.Cells[5].Value}
		have := []string{got.Name, got.StartDate.Format(time.DateOnly), got.EndDate.Format(time.DateOnly), got.Goal, string(got.Status)}
		if got.ID != id || fmt.Sprint(have) != fmt.Sprint(want) {
			return fmt.Errorf("sprint %d = %d %v, se esperaba %d %v", i+1, got.ID, have, id, want)
		}
	}
	return nil
}

func (s *historyScenario) sprintsEmpty() error {
	if s.sprints == nil || len(s.sprints) != 0 {
		return fmt.Errorf("se esperaba una lista vacía, no nil, y se obtuvo %#v", s.sprints)
	}
	return nil
}

func (s *historyScenario) completedOK() error {
	if s.completedErr != nil {
		return fmt.Errorf("se esperaba una consulta exitosa, falló con: %v", s.completedErr)
	}
	return nil
}

func (s *historyScenario) completedAre(first, second string) error {
	if len(s.completed) != 2 || s.completed[0].Title != first || s.completed[1].Title != second {
		return fmt.Errorf("ítems completados = %+v, se esperaban %q y %q", s.completed, first, second)
	}
	return nil
}

func (s *historyScenario) completedEmpty() error {
	if s.completed == nil || len(s.completed) != 0 {
		return fmt.Errorf("se esperaba una lista vacía, no nil, y se obtuvo %#v", s.completed)
	}
	return nil
}

func (s *historyScenario) historyFailsProjectNotFound() error {
	if !errors.Is(s.historyErr, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.historyErr, project.ErrProjectNotFound)
	}
	if s.sprints != nil {
		return fmt.Errorf("se esperaba nil y se obtuvo %#v", s.sprints)
	}
	return nil
}

func (s *historyScenario) historyFailsRepository() error {
	if !errors.Is(s.historyErr, errRepository) {
		return fmt.Errorf("err = %v, se esperaba %v", s.historyErr, errRepository)
	}
	if s.sprints != nil {
		return fmt.Errorf("se esperaba nil y se obtuvo %#v", s.sprints)
	}
	return nil
}
