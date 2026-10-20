package sprint

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-13-cerrar-sprint.feature con
// godog. Los pasos viven en el paquete sprint porque la historia es de E3 y
// orquesta los paquetes sprint y backlog (ver SDD).
func TestFeatures_HU13CerrarSprint(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeCloseScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-13-cerrar-sprint.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-13")
	}
}

// closeScenario guarda el estado de un escenario de HU-13.
type closeScenario struct {
	projects *project.Manager
	items    *backlog.Manager
	itemRepo *backlog.MemoryRepository
	repo     Repository
	manager  *Manager

	projectIDs  map[string]int
	closed      Sprint
	created     Sprint
	closeErr    error
	createErr   error
	completeErr error
	assignErr   error
	log         *closeEventLog
}

func initializeCloseScenario(ctx *godog.ScenarioContext) {
	s := &closeScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.itemRepo = backlog.NewMemoryRepository()
		s.items = backlog.NewManager(s.itemRepo, s.projects)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects, s.items)
		s.projectIDs = map[string]int{}
		s.closed = Sprint{}
		s.created = Sprint{}
		s.closeErr = nil
		s.createErr = nil
		s.completeErr = nil
		s.assignErr = nil
		s.log = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un ítem "([^"]*)" en estado "([^"]*)" en el proyecto "([^"]*)"$`, s.existingItem)
	ctx.Step(`^que existe un sprint (abierto|cerrado) "([^"]*)" en el proyecto "([^"]*)"$`, s.existingSprint)
	ctx.Step(`^que el ítem con el ID (-?\d+) está asignado al sprint con el ID (-?\d+)$`, s.assignedItem)
	ctx.Step(`^que el repositorio de sprints falla al actualizar$`, s.repoFailsOnUpdate)
	ctx.Step(`^cierro el sprint con el ID (-?\d+)$`, s.closeSprint)
	ctx.Step(`^creo el sprint "([^"]*)" con inicio "([^"]*)", fin "([^"]*)" y sprint goal "([^"]*)" en el proyecto "([^"]*)"$`, s.createSprint)
	ctx.Step(`^completo el ítem con el ID (-?\d+) del sprint con el ID (-?\d+)$`, s.completeItem)
	ctx.Step(`^asigno el ítem con el ID (-?\d+) al sprint con el ID (-?\d+)$`, s.assignItem)
	ctx.Step(`^cierro el sprint y en simultáneo completo el ítem con el ID (-?\d+) y asigno el ítem con el ID (-?\d+) al sprint con el ID (-?\d+)$`, s.concurrentCloseAssignComplete)
	ctx.Step(`^cambio el nombre, el sprint goal y las fechas del sprint devuelto$`, s.mutateClosedSprint)
	ctx.Step(`^el cierre es exitoso$`, s.closeOK)
	ctx.Step(`^el cierre falla porque el sprint está cerrado$`, s.closeFailsSprintClosed)
	ctx.Step(`^el cierre falla porque no existe un sprint con ese id$`, s.closeFailsSprintNotFound)
	ctx.Step(`^el cierre falla con el error del repositorio$`, s.closeFailsRepository)
	ctx.Step(`^la creación es exitosa$`, s.createOK)
	ctx.Step(`^el sprint con el ID (-?\d+) queda en estado "([^"]*)"$`, s.sprintHasStatus)
	ctx.Step(`^el sprint con el ID (-?\d+) sigue en estado "([^"]*)"$`, s.sprintHasStatus)
	ctx.Step(`^el sprint guardado con el ID (-?\d+) conserva el nombre "([^"]*)", el sprint goal "([^"]*)" y el proyecto "([^"]*)"$`, s.savedSprintKeepsFields)
	ctx.Step(`^el completado falla porque el sprint está cerrado$`, s.completeFailsSprintClosed)
	ctx.Step(`^la asignación falla porque el sprint está cerrado$`, s.assignFailsSprintClosed)
	ctx.Step(`^el ítem con el ID (-?\d+) queda en estado "([^"]*)"$`, s.itemHasStatus)
	ctx.Step(`^el ítem con el ID (-?\d+) sigue asignado al sprint con el ID (-?\d+)$`, s.itemHasSprintID)
	ctx.Step(`^el ítem con el ID (-?\d+) sigue sin sprint asignado$`, s.itemWithoutSprint)
	ctx.Step(`^al reconsultar el sprint con el ID (-?\d+) obtengo el nombre "([^"]*)", el sprint goal "([^"]*)" y el estado "([^"]*)"$`, s.reconsultSprint)
	ctx.Step(`^cada operación simultánea termina bien o porque el sprint está cerrado$`, s.concurrentOpsOKOrClosed)
	ctx.Step(`^ninguna escritura de ítems ocurre después del cierre$`, s.noItemWriteAfterClose)
}

func (s *closeScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *closeScenario) projectID(projectName string) (int, error) {
	id, ok := s.projectIDs[projectName]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	return id, nil
}

func (s *closeScenario) existingItem(title, status, projectName string) error {
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
		StoryPoints:        5,
		EstimatedHours:     2.5,
	})
	return err
}

func (s *closeScenario) existingSprint(kind, name, projectName string) error {
	projectID, err := s.projectID(projectName)
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
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    status,
	})
	return err
}

func (s *closeScenario) assignedItem(itemID, sprintID int) error {
	_, err := s.items.AssignToSprint(itemID, sprintID)
	return err
}

func (s *closeScenario) repoFailsOnUpdate() error {
	s.repo = &failingUpdateSprintRepository{Repository: s.repo, updateErr: errRepository}
	s.manager = NewManager(s.repo, s.projects, s.items)
	return nil
}

func (s *closeScenario) closeSprint(sprintID int) error {
	s.closed, s.closeErr = s.manager.Close(sprintID)
	return nil
}

func (s *closeScenario) createSprint(name, startRaw, endRaw, goal, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	start, err := time.Parse("2006-01-02", startRaw)
	if err != nil {
		return err
	}
	end, err := time.Parse("2006-01-02", endRaw)
	if err != nil {
		return err
	}
	s.created, s.createErr = s.manager.Create(projectID, name, start, end, goal)
	return nil
}

func (s *closeScenario) completeItem(itemID, sprintID int) error {
	_, s.completeErr = s.manager.CompleteItem(sprintID, itemID)
	return nil
}

func (s *closeScenario) assignItem(itemID, sprintID int) error {
	_, s.assignErr = s.manager.AssignItem(sprintID, itemID)
	return nil
}

func (s *closeScenario) concurrentCloseAssignComplete(completeItemID, assignItemID, sprintID int) error {
	log := &closeEventLog{}
	sprintRepo := &recordingSprintRepository{Repository: s.repo, log: log}
	itemRepo := &recordingItemRepository{
		MemoryRepository: s.itemRepo,
		log:              log,
		delay:            time.Millisecond,
	}
	manager := NewManager(sprintRepo, s.projects, backlog.NewManager(itemRepo, s.projects))

	start := make(chan struct{})
	var wg sync.WaitGroup
	var closeErr, completeErr, assignErr error

	wg.Add(3)
	go func() {
		defer wg.Done()
		<-start
		_, closeErr = manager.Close(sprintID)
	}()
	go func() {
		defer wg.Done()
		<-start
		_, completeErr = manager.CompleteItem(sprintID, completeItemID)
	}()
	go func() {
		defer wg.Done()
		<-start
		_, assignErr = manager.AssignItem(sprintID, assignItemID)
	}()
	close(start)
	wg.Wait()

	s.closeErr = closeErr
	s.completeErr = completeErr
	s.assignErr = assignErr
	s.log = log
	return nil
}

func (s *closeScenario) mutateClosedSprint() error {
	if err := s.closeOK(); err != nil {
		return err
	}
	s.closed.Name = "Otro nombre"
	s.closed.Goal = "Otro objetivo"
	s.closed.StartDate = validStart.AddDate(0, 0, -7)
	s.closed.EndDate = validEnd.AddDate(0, 0, -7)
	return nil
}

func (s *closeScenario) closeOK() error {
	if s.closeErr != nil {
		return fmt.Errorf("el cierre falló: %v", s.closeErr)
	}
	return nil
}

func (s *closeScenario) closeFailsSprintClosed() error {
	if !errors.Is(s.closeErr, ErrSprintClosed) {
		return fmt.Errorf("error = %v, se esperaba ErrSprintClosed", s.closeErr)
	}
	return nil
}

func (s *closeScenario) closeFailsSprintNotFound() error {
	if !errors.Is(s.closeErr, ErrSprintNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrSprintNotFound", s.closeErr)
	}
	return nil
}

func (s *closeScenario) closeFailsRepository() error {
	if !errors.Is(s.closeErr, errRepository) {
		return fmt.Errorf("error = %v, se esperaba errRepository", s.closeErr)
	}
	return nil
}

func (s *closeScenario) createOK() error {
	if s.createErr != nil {
		return fmt.Errorf("la creación falló: %v", s.createErr)
	}
	return nil
}

func (s *closeScenario) sprintHasStatus(sprintID int, status string) error {
	saved, err := s.repo.GetByID(sprintID)
	if err != nil {
		return err
	}
	if string(saved.Status) != status {
		return fmt.Errorf("estado del sprint %d = %q, se esperaba %q", sprintID, saved.Status, status)
	}
	return nil
}

func (s *closeScenario) savedSprintKeepsFields(sprintID int, name, goal, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	saved, err := s.repo.GetByID(sprintID)
	if err != nil {
		return err
	}
	if saved.Name != name || saved.Goal != goal {
		return fmt.Errorf("sprint guardado = %q/%q, se esperaba %q/%q", saved.Name, saved.Goal, name, goal)
	}
	if saved.ProjectID != projectID {
		return fmt.Errorf("ProjectID guardado = %d, se esperaba %d", saved.ProjectID, projectID)
	}
	return nil
}

func (s *closeScenario) completeFailsSprintClosed() error {
	if !errors.Is(s.completeErr, ErrSprintClosed) {
		return fmt.Errorf("error = %v, se esperaba ErrSprintClosed", s.completeErr)
	}
	return nil
}

func (s *closeScenario) assignFailsSprintClosed() error {
	if !errors.Is(s.assignErr, ErrSprintClosed) {
		return fmt.Errorf("error = %v, se esperaba ErrSprintClosed", s.assignErr)
	}
	return nil
}

func (s *closeScenario) itemHasStatus(itemID int, status string) error {
	item, err := s.items.GetItem(itemID)
	if err != nil {
		return err
	}
	if string(item.Status) != status {
		return fmt.Errorf("estado del ítem %d = %q, se esperaba %q", itemID, item.Status, status)
	}
	return nil
}

func (s *closeScenario) itemHasSprintID(itemID, sprintID int) error {
	item, err := s.items.GetItem(itemID)
	if err != nil {
		return err
	}
	if item.SprintID != sprintID {
		return fmt.Errorf("SprintID del ítem %d = %d, se esperaba %d", itemID, item.SprintID, sprintID)
	}
	return nil
}

func (s *closeScenario) itemWithoutSprint(itemID int) error {
	item, err := s.items.GetItem(itemID)
	if err != nil {
		return err
	}
	if item.SprintID != 0 {
		return fmt.Errorf("SprintID del ítem %d = %d, se esperaba 0", itemID, item.SprintID)
	}
	return nil
}

func (s *closeScenario) reconsultSprint(sprintID int, name, goal, status string) error {
	saved, err := s.repo.GetByID(sprintID)
	if err != nil {
		return err
	}
	if saved.Name != name || saved.Goal != goal {
		return fmt.Errorf("sprint guardado = %q/%q, se esperaba %q/%q", saved.Name, saved.Goal, name, goal)
	}
	if string(saved.Status) != status {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", saved.Status, status)
	}
	return nil
}

func (s *closeScenario) concurrentOpsOKOrClosed() error {
	if s.closeErr != nil {
		return fmt.Errorf("el cierre falló: %v", s.closeErr)
	}
	for name, err := range map[string]error{"completado": s.completeErr, "asignación": s.assignErr} {
		if err != nil && !errors.Is(err, ErrSprintClosed) {
			return fmt.Errorf("la operación simultánea de %s falló con un error inesperado: %v", name, err)
		}
	}
	return nil
}

func (s *closeScenario) noItemWriteAfterClose() error {
	if s.log == nil {
		return fmt.Errorf("no se registraron las escrituras del escenario")
	}
	events := s.log.snapshot()
	closeIndex := -1
	for i, event := range events {
		if event == "close" {
			closeIndex = i
			break
		}
	}
	if closeIndex == -1 {
		return fmt.Errorf("no se registró la escritura de cierre")
	}
	for _, event := range events[closeIndex+1:] {
		if event == "item" {
			return fmt.Errorf("se escribió un ítem después del cierre: %v", events)
		}
	}
	return nil
}
