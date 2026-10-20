package sprint

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-12-registrar-completadas.feature
// con godog. Los pasos viven en el paquete sprint porque la historia es de E3
// y orquesta los paquetes sprint y backlog (ver SDD).
func TestFeatures_HU12RegistrarCompletadas(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeCompleteScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-12-registrar-completadas.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-12")
	}
}

// completeScenario guarda el estado de un escenario de HU-12.
type completeScenario struct {
	projects *project.Manager
	items    *backlog.Manager
	itemRepo *backlog.MemoryRepository
	repo     *MemoryRepository
	manager  *Manager

	projectIDs     map[string]int
	completed      backlog.Item
	completedItems []backlog.Item
	err            error
	errs           []error
}

func initializeCompleteScenario(ctx *godog.ScenarioContext) {
	s := &completeScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		projectRepo := project.NewMemoryRepository()
		s.projects = project.NewManager(projectRepo)
		s.itemRepo = backlog.NewMemoryRepository()
		s.items = backlog.NewManager(s.itemRepo, s.projects)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects, s.items)
		s.projectIDs = map[string]int{}
		s.completed = backlog.Item{}
		s.completedItems = nil
		s.err = nil
		s.errs = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un ítem "([^"]*)" en estado "([^"]*)" en el proyecto "([^"]*)"$`, s.existingItem)
	ctx.Step(`^que existe un sprint (abierto|cerrado) "([^"]*)" en el proyecto "([^"]*)"$`, s.existingSprint)
	ctx.Step(`^que el ítem con el ID (-?\d+) está asignado al sprint con el ID (-?\d+)$`, s.assignedItem)
	ctx.Step(`^completo el ítem con el ID (-?\d+) del sprint con el ID (-?\d+)$`, s.completeItem)
	ctx.Step(`^consulto los completados del sprint con el ID (-?\d+)$`, s.listCompletedItems)
	ctx.Step(`^completo en simultáneo (\d+) veces el ítem con el ID (-?\d+) del sprint con el ID (-?\d+)$`, s.completeConcurrently)
	ctx.Step(`^el completado es exitoso$`, s.completeOK)
	ctx.Step(`^el completado falla porque el ítem no está asignado al sprint$`, s.failsWithItemNotAssigned)
	ctx.Step(`^el completado falla porque la transición de estado no está permitida$`, s.failsWithInvalidTransition)
	ctx.Step(`^el completado falla porque el sprint está cerrado$`, s.failsWithSprintClosed)
	ctx.Step(`^el completado falla porque no existe un sprint con ese id$`, s.failsWithSprintNotFound)
	ctx.Step(`^el completado falla porque no existe un ítem del backlog con ese id$`, s.failsWithItemNotFound)
	ctx.Step(`^el ítem con el ID (-?\d+) queda en estado "([^"]*)"$`, s.itemHasStatus)
	ctx.Step(`^el ítem guardado con el ID (-?\d+) conserva el título "([^"]*)"$`, s.itemKeepsTitle)
	ctx.Step(`^el ítem guardado con el ID (-?\d+) conserva la descripción "([^"]*)" y los criterios "([^"]*)"$`, s.itemKeepsDescriptionAndCriteria)
	ctx.Step(`^la consulta es exitosa$`, s.listOK)
	ctx.Step(`^la lista devuelta tiene (\d+) ítems?$`, s.listHasCount)
	ctx.Step(`^la lista devuelta está vacía$`, s.listIsEmpty)
	ctx.Step(`^la lista contiene el ítem con el ID (-?\d+)$`, s.listContains)
	ctx.Step(`^la lista no contiene el ítem con el ID (-?\d+)$`, s.listNotContains)
	ctx.Step(`^la lista está ordenada por ID ascendente$`, s.listOrderedByID)
	ctx.Step(`^cambio el título y los criterios del ítem completado devuelto$`, s.mutateCompletedItem)
	ctx.Step(`^al reconsultar el ítem con el ID (-?\d+) obtengo el título "([^"]*)" y el estado "([^"]*)"$`, s.reconsultTitleAndStatus)
	ctx.Step(`^todas las marcaciones tienen éxito$`, s.allCompleteOK)
}

func (s *completeScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *completeScenario) projectID(projectName string) (int, error) {
	id, ok := s.projectIDs[projectName]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	return id, nil
}

func (s *completeScenario) existingItem(title, status, projectName string) error {
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

func (s *completeScenario) existingSprint(kind, name, projectName string) error {
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

func (s *completeScenario) assignedItem(itemID, sprintID int) error {
	_, err := s.items.AssignToSprint(itemID, sprintID)
	return err
}

func (s *completeScenario) completeItem(itemID, sprintID int) error {
	s.completed, s.err = s.manager.CompleteItem(sprintID, itemID)
	return nil
}

func (s *completeScenario) listCompletedItems(sprintID int) error {
	s.completedItems, s.err = s.manager.ListCompletedItems(sprintID)
	return nil
}

func (s *completeScenario) completeConcurrently(attempts, itemID, sprintID int) error {
	s.errs = make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.errs[i] = s.manager.CompleteItem(sprintID, itemID)
		}()
	}
	wg.Wait()
	return nil
}

func (s *completeScenario) completeOK() error {
	if s.err != nil {
		return fmt.Errorf("el completado falló: %v", s.err)
	}
	return nil
}

func (s *completeScenario) failsWithItemNotAssigned() error {
	if !errors.Is(s.err, ErrItemNotAssignedToSprint) {
		return fmt.Errorf("error = %v, se esperaba ErrItemNotAssignedToSprint", s.err)
	}
	return nil
}

func (s *completeScenario) failsWithInvalidTransition() error {
	if !errors.Is(s.err, backlog.ErrInvalidTransition) {
		return fmt.Errorf("error = %v, se esperaba ErrInvalidTransition", s.err)
	}
	return nil
}

func (s *completeScenario) failsWithSprintClosed() error {
	if !errors.Is(s.err, ErrSprintClosed) {
		return fmt.Errorf("error = %v, se esperaba ErrSprintClosed", s.err)
	}
	return nil
}

func (s *completeScenario) failsWithSprintNotFound() error {
	if !errors.Is(s.err, ErrSprintNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrSprintNotFound", s.err)
	}
	return nil
}

func (s *completeScenario) failsWithItemNotFound() error {
	if !errors.Is(s.err, backlog.ErrItemNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrItemNotFound", s.err)
	}
	return nil
}

func (s *completeScenario) itemHasStatus(id int, status string) error {
	item, err := s.items.GetItem(id)
	if err != nil {
		return err
	}
	if string(item.Status) != status {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", item.Status, status)
	}
	return nil
}

func (s *completeScenario) itemKeepsTitle(id int, title string) error {
	item, err := s.items.GetItem(id)
	if err != nil {
		return err
	}
	if item.Title != title {
		return fmt.Errorf("título guardado = %q, se esperaba %q", item.Title, title)
	}
	return nil
}

func (s *completeScenario) itemKeepsDescriptionAndCriteria(id int, description, criterion string) error {
	item, err := s.items.GetItem(id)
	if err != nil {
		return err
	}
	if item.Description != description {
		return fmt.Errorf("descripción guardada = %q, se esperaba %q", item.Description, description)
	}
	if len(item.AcceptanceCriteria) != 1 || item.AcceptanceCriteria[0] != criterion {
		return fmt.Errorf("criterios guardados = %q, se esperaban [%q]", item.AcceptanceCriteria, criterion)
	}
	return nil
}

func (s *completeScenario) listOK() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	return nil
}

func (s *completeScenario) listHasCount(count int) error {
	if err := s.listOK(); err != nil {
		return err
	}
	if len(s.completedItems) != count {
		return fmt.Errorf("cantidad devuelta = %d, se esperaban %d", len(s.completedItems), count)
	}
	return nil
}

func (s *completeScenario) listIsEmpty() error {
	if err := s.listOK(); err != nil {
		return err
	}
	if s.completedItems == nil {
		return fmt.Errorf("la lista devuelta fue nil, se esperaba una lista vacía")
	}
	if len(s.completedItems) != 0 {
		return fmt.Errorf("cantidad devuelta = %d, se esperaba 0", len(s.completedItems))
	}
	return nil
}

func (s *completeScenario) listContains(id int) error {
	if err := s.listOK(); err != nil {
		return err
	}
	for _, item := range s.completedItems {
		if item.ID == id {
			return nil
		}
	}
	return fmt.Errorf("la lista no contiene el ítem %d: %+v", id, s.completedItems)
}

func (s *completeScenario) listNotContains(id int) error {
	if err := s.listOK(); err != nil {
		return err
	}
	for _, item := range s.completedItems {
		if item.ID == id {
			return fmt.Errorf("la lista contiene el ítem %d y no debería: %+v", id, s.completedItems)
		}
	}
	return nil
}

func (s *completeScenario) listOrderedByID() error {
	if err := s.listOK(); err != nil {
		return err
	}
	for i := 1; i < len(s.completedItems); i++ {
		if s.completedItems[i-1].ID >= s.completedItems[i].ID {
			return fmt.Errorf("la lista no está ordenada por ID: %+v", s.completedItems)
		}
	}
	return nil
}

func (s *completeScenario) mutateCompletedItem() error {
	if err := s.completeOK(); err != nil {
		return err
	}
	if len(s.completed.AcceptanceCriteria) == 0 {
		return fmt.Errorf("el ítem completado no tiene criterios para modificar")
	}
	s.completed.Title = "Otro título"
	s.completed.AcceptanceCriteria[0] = "Otro criterio"
	s.completed.SprintID = 999
	return nil
}

func (s *completeScenario) reconsultTitleAndStatus(id int, title, status string) error {
	item, err := s.items.GetItem(id)
	if err != nil {
		return err
	}
	if item.Title != title {
		return fmt.Errorf("título guardado = %q, se esperaba %q", item.Title, title)
	}
	if string(item.Status) != status {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", item.Status, status)
	}
	return nil
}

func (s *completeScenario) allCompleteOK() error {
	for i, err := range s.errs {
		if err != nil {
			return fmt.Errorf("la marcación %d falló: %v", i+1, err)
		}
	}
	return nil
}
