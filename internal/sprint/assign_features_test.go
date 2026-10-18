package sprint

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-11-asignar-item-sprint.feature
// con godog. Los pasos viven en el paquete sprint porque la historia es de E3
// y orquesta los paquetes sprint y backlog (ver SDD).
func TestFeatures_HU11AsignarItemSprint(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeAssignItemScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-11-asignar-item-sprint.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-11")
	}
}

// assignItemScenario guarda el estado de un escenario de HU-11.
type assignItemScenario struct {
	projectRepo *project.MemoryRepository
	projects    *project.Manager
	itemRepo    *backlog.MemoryRepository
	items       *backlog.Manager
	repo        *MemoryRepository
	manager     *Manager

	projectIDs map[string]int
	assigned   backlog.Item
	err        error
	errs       []error
	updateErr  error
}

func initializeAssignItemScenario(ctx *godog.ScenarioContext) {
	s := &assignItemScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projectRepo = project.NewMemoryRepository()
		s.projects = project.NewManager(s.projectRepo)
		s.itemRepo = backlog.NewMemoryRepository()
		s.items = backlog.NewManager(s.itemRepo, s.projects)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects, s.items)
		s.projectIDs = map[string]int{}
		s.assigned = backlog.Item{}
		s.err = nil
		s.errs = nil
		s.updateErr = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un ítem con título "([^"]*)", descripción "([^"]*)", prioridad "([^"]*)" y criterios de aceptación "([^"]*)" en el proyecto "([^"]*)"$`, s.existingItem)
	ctx.Step(`^que existe un ítem con título "([^"]*)", descripción "([^"]*)", prioridad "([^"]*)" y criterios de aceptación "([^"]*)" y "([^"]*)" en el proyecto "([^"]*)"$`, s.existingItemTwoCriteria)
	ctx.Step(`^que existe un ítem con título "([^"]*)", descripción "([^"]*)", prioridad "([^"]*)" y criterios de aceptación "([^"]*)" en el estado "([^"]*)" en el proyecto "([^"]*)"$`, s.existingItemWithStatus)
	ctx.Step(`^que existen (\d+) ítems en el proyecto "([^"]*)"$`, s.existingItems)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que existe un sprint cerrado "([^"]*)" en el proyecto "([^"]*)"$`, s.existingClosedSprint)
	ctx.Step(`^que existen dos sprints abiertos "([^"]*)" y "([^"]*)" en el proyecto "([^"]*)"$`, s.existingTwoOpenSprints)
	ctx.Step(`^que no hay sprints creados$`, s.noSprintsCreated)
	ctx.Step(`^que no hay ítems creados$`, s.noItemsCreated)
	ctx.Step(`^asigno el ítem con el ID (-?\d+) al sprint con el ID (-?\d+)$`, s.assignItem)
	ctx.Step(`^se asignan en simultáneo (\d+) veces el ítem con el ID (\d+) a los sprints con los IDs (\d+) y (\d+)$`, s.concurrentAssignmentsSameItem)
	ctx.Step(`^se asignan en simultáneo los (\d+) ítems al sprint con el ID (\d+)$`, s.concurrentAssignmentsDistinctItems)
	ctx.Step(`^modifico el ítem con el ID (\d+) con título "([^"]*)", descripción "([^"]*)", prioridad "([^"]*)", criterios de aceptación "([^"]*)" y horas estimadas (\d+(?:\.\d+)?) con el Update de HU-07$`, s.updateItem)
	ctx.Step(`^cambio el título, los criterios y el sprint del ítem devuelto$`, s.mutateAssignedItem)
	ctx.Step(`^la asignación es exitosa$`, s.assignOK)
	ctx.Step(`^la modificación es exitosa$`, s.updateOK)
	ctx.Step(`^el ítem devuelto queda asignado al sprint con el ID (\d+)$`, s.assignedWithSprintID)
	ctx.Step(`^el ítem guardado queda asignado al sprint con el ID (\d+)$`, s.savedWithSprintID)
	ctx.Step(`^el ítem guardado sigue asignado al sprint con el ID (\d+)$`, s.savedWithSprintID)
	ctx.Step(`^el ítem guardado sigue sin sprint asignado$`, s.savedWithoutSprint)
	ctx.Step(`^los dos ítems quedan asignados al sprint con el ID (\d+)$`, s.twoItemsWithSprintID)
	ctx.Step(`^los (\d+) ítems quedan asignados al sprint con el ID (\d+)$`, s.itemsWithSprintID)
	ctx.Step(`^el ítem guardado conserva el título "([^"]*)", la descripción "([^"]*)", la prioridad "([^"]*)", los criterios de aceptación "([^"]*)" y "([^"]*)", el estado "([^"]*)", los story points (\d+) y las horas estimadas (\d+(?:\.\d+)?)$`, s.savedKeepsFields)
	ctx.Step(`^al reconsultar el ítem con el ID (\d+) obtengo el título "([^"]*)" y sigue asignado al sprint con el ID (\d+)$`, s.reconsultKeepsTitleAndSprint)
	ctx.Step(`^exactamente una asignación tiene éxito$`, s.exactlyOneAssignmentOK)
	ctx.Step(`^las demás fallan porque el ítem ya está asignado a un sprint$`, s.othersFailWithAlreadyAssigned)
	ctx.Step(`^el ítem guardado queda asignado a uno de los sprints$`, s.savedWithEitherSprint)
	ctx.Step(`^las (\d+) asignaciones tienen éxito$`, s.allAssignmentsOK)
	ctx.Step(`^la asignación falla porque el ítem ya está asignado a un sprint$`, s.failsWithAlreadyAssigned)
	ctx.Step(`^la asignación falla porque el sprint está cerrado$`, s.failsWithSprintClosed)
	ctx.Step(`^la asignación falla porque el ítem pertenece a otro proyecto$`, s.failsWithItemFromAnotherProject)
	ctx.Step(`^la asignación falla porque no existe un sprint con ese id$`, s.failsWithSprintNotFound)
	ctx.Step(`^la asignación falla porque no existe un ítem del backlog con ese id$`, s.failsWithItemNotFound)
}

func (s *assignItemScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *assignItemScenario) projectID(projectName string) (int, error) {
	id, ok := s.projectIDs[projectName]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	return id, nil
}

func (s *assignItemScenario) existingItem(title, description, priority, criterion, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	_, err = s.items.Create(projectID, title, description, priority, []string{criterion})
	return err
}

func (s *assignItemScenario) existingItemTwoCriteria(title, description, priority, first, second, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	_, err = s.items.Create(projectID, title, description, priority, []string{first, second})
	return err
}

// existingItemWithStatus guarda un ítem directo en el repositorio con el
// estado recibido: el Manager siempre crea PENDIENTE y el estado de un ítem
// no impide asignarlo (RN-11.6).
func (s *assignItemScenario) existingItemWithStatus(title, description, priority, criterion, status, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	_, err = s.itemRepo.Save(backlog.Item{
		ProjectID:          projectID,
		Title:              title,
		Description:        description,
		Priority:           backlog.Priority(priority),
		AcceptanceCriteria: []string{criterion},
		Status:             backlog.Status(status),
	})
	return err
}

func (s *assignItemScenario) existingItems(count int, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	for i := 1; i <= count; i++ {
		if _, err := s.items.Create(projectID, fmt.Sprintf("Ítem %d", i), "", "ALTA", []string{"Criterio OK"}); err != nil {
			return err
		}
	}
	return nil
}

// saveSprint guarda un sprint directo en el repositorio con el estado
// recibido (los pasos son internos al paquete, ver SDD).
func (s *assignItemScenario) saveSprint(name, projectName string, status Status) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
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

func (s *assignItemScenario) existingOpenSprint(name, projectName string) error {
	return s.saveSprint(name, projectName, StatusOpen)
}

func (s *assignItemScenario) existingClosedSprint(name, projectName string) error {
	return s.saveSprint(name, projectName, StatusClosed)
}

// existingTwoOpenSprints guarda dos sprints ABIERTOS directo en el
// repositorio: la regla de un solo sprint abierto por proyecto la aplica el
// Manager (RN-10.5), no el repositorio, y HU-11 necesita dos destinos válidos
// para el mismo ítem (RN-11.3 y CA-11.6).
func (s *assignItemScenario) existingTwoOpenSprints(firstName, secondName, projectName string) error {
	if err := s.saveSprint(firstName, projectName, StatusOpen); err != nil {
		return err
	}
	return s.saveSprint(secondName, projectName, StatusOpen)
}

func (s *assignItemScenario) noSprintsCreated() error {
	return nil
}

func (s *assignItemScenario) noItemsCreated() error {
	return nil
}

func (s *assignItemScenario) assignItem(itemID, sprintID int) error {
	s.assigned, s.err = s.manager.AssignItem(sprintID, itemID)
	return nil
}

func (s *assignItemScenario) concurrentAssignmentsSameItem(attempts, itemID, firstSprintID, secondSprintID int) error {
	s.errs = make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sprintID := firstSprintID
			if i%2 == 1 {
				sprintID = secondSprintID
			}
			_, s.errs[i] = s.manager.AssignItem(sprintID, itemID)
		}()
	}
	wg.Wait()
	return nil
}

func (s *assignItemScenario) concurrentAssignmentsDistinctItems(count, sprintID int) error {
	s.errs = make([]error, count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.errs[i] = s.manager.AssignItem(sprintID, i+1)
		}()
	}
	wg.Wait()
	return nil
}

func (s *assignItemScenario) updateItem(id int, title, description, priority, criterion, hoursRaw string) error {
	hours, err := strconv.ParseFloat(hoursRaw, 64)
	if err != nil {
		return err
	}
	_, s.updateErr = s.items.Update(id, title, description, priority, []string{criterion}, hours)
	return nil
}

func (s *assignItemScenario) mutateAssignedItem() error {
	if err := s.assignOK(); err != nil {
		return err
	}
	s.assigned.Title = "Otro título"
	s.assigned.AcceptanceCriteria = []string{"Otro criterio"}
	s.assigned.SprintID = 999
	return nil
}

func (s *assignItemScenario) assignOK() error {
	if s.err != nil {
		return fmt.Errorf("la asignación falló: %v", s.err)
	}
	return nil
}

func (s *assignItemScenario) updateOK() error {
	if s.updateErr != nil {
		return fmt.Errorf("la modificación falló: %v", s.updateErr)
	}
	return nil
}

func (s *assignItemScenario) assignedWithSprintID(sprintID int) error {
	if err := s.assignOK(); err != nil {
		return err
	}
	if s.assigned.SprintID != sprintID {
		return fmt.Errorf("SprintID devuelto = %d, se esperaba %d", s.assigned.SprintID, sprintID)
	}
	return nil
}

// savedItem consulta el ítem guardado con el Manager de backlog (HU-05).
func (s *assignItemScenario) savedItem(id int) (backlog.Item, error) {
	return s.items.GetItem(id)
}

func (s *assignItemScenario) savedWithSprintID(sprintID int) error {
	item, err := s.savedItem(1)
	if err != nil {
		return err
	}
	if item.SprintID != sprintID {
		return fmt.Errorf("SprintID guardado = %d, se esperaba %d", item.SprintID, sprintID)
	}
	return nil
}

func (s *assignItemScenario) savedWithoutSprint() error {
	item, err := s.savedItem(1)
	if err != nil {
		return err
	}
	if item.SprintID != 0 {
		return fmt.Errorf("SprintID guardado = %d, se esperaba 0", item.SprintID)
	}
	return nil
}

func (s *assignItemScenario) twoItemsWithSprintID(sprintID int) error {
	for id := 1; id <= 2; id++ {
		item, err := s.savedItem(id)
		if err != nil {
			return err
		}
		if item.SprintID != sprintID {
			return fmt.Errorf("SprintID guardado del ítem %d = %d, se esperaba %d", id, item.SprintID, sprintID)
		}
	}
	return nil
}

func (s *assignItemScenario) itemsWithSprintID(count, sprintID int) error {
	for id := 1; id <= count; id++ {
		item, err := s.savedItem(id)
		if err != nil {
			return err
		}
		if item.SprintID != sprintID {
			return fmt.Errorf("SprintID guardado del ítem %d = %d, se esperaba %d", id, item.SprintID, sprintID)
		}
	}
	return nil
}

func (s *assignItemScenario) savedKeepsFields(title, description, priority, firstCriterion, secondCriterion, status string, storyPoints int, hoursRaw string) error {
	hours, err := strconv.ParseFloat(hoursRaw, 64)
	if err != nil {
		return err
	}
	item, err := s.savedItem(1)
	if err != nil {
		return err
	}
	if item.Title != title {
		return fmt.Errorf("título guardado = %q, se esperaba %q", item.Title, title)
	}
	if item.Description != description {
		return fmt.Errorf("descripción guardada = %q, se esperaba %q", item.Description, description)
	}
	if string(item.Priority) != priority {
		return fmt.Errorf("prioridad guardada = %q, se esperaba %q", item.Priority, priority)
	}
	wantCriteria := []string{firstCriterion, secondCriterion}
	if len(item.AcceptanceCriteria) != len(wantCriteria) {
		return fmt.Errorf("criterios guardados = %q, se esperaban %q", item.AcceptanceCriteria, wantCriteria)
	}
	for i, criterion := range wantCriteria {
		if item.AcceptanceCriteria[i] != criterion {
			return fmt.Errorf("criterio guardado %d = %q, se esperaba %q", i+1, item.AcceptanceCriteria[i], criterion)
		}
	}
	if string(item.Status) != status {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", item.Status, status)
	}
	if item.StoryPoints != storyPoints {
		return fmt.Errorf("story points guardados = %d, se esperaban %d", item.StoryPoints, storyPoints)
	}
	if item.EstimatedHours != hours {
		return fmt.Errorf("horas guardadas = %v, se esperaban %v", item.EstimatedHours, hours)
	}
	return nil
}

func (s *assignItemScenario) reconsultKeepsTitleAndSprint(id int, title string, sprintID int) error {
	item, err := s.savedItem(id)
	if err != nil {
		return err
	}
	if item.Title != title {
		return fmt.Errorf("título guardado = %q, se esperaba %q", item.Title, title)
	}
	if item.SprintID != sprintID {
		return fmt.Errorf("SprintID guardado = %d, se esperaba %d", item.SprintID, sprintID)
	}
	return nil
}

func (s *assignItemScenario) exactlyOneAssignmentOK() error {
	successes := 0
	for _, err := range s.errs {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		return fmt.Errorf("tuvieron éxito %d asignaciones, se esperaba exactamente 1", successes)
	}
	return nil
}

func (s *assignItemScenario) othersFailWithAlreadyAssigned() error {
	for i, err := range s.errs {
		if err != nil && !errors.Is(err, backlog.ErrItemAlreadyAssigned) {
			return fmt.Errorf("la asignación %d falló con un error inesperado: %v", i+1, err)
		}
	}
	return nil
}

func (s *assignItemScenario) savedWithEitherSprint() error {
	item, err := s.savedItem(1)
	if err != nil {
		return err
	}
	if item.SprintID != 1 && item.SprintID != 2 {
		return fmt.Errorf("SprintID guardado = %d, se esperaba 1 o 2", item.SprintID)
	}
	return nil
}

func (s *assignItemScenario) allAssignmentsOK(count int) error {
	successes := 0
	for _, err := range s.errs {
		if err == nil {
			successes++
		}
	}
	if successes != count {
		return fmt.Errorf("tuvieron éxito %d asignaciones, se esperaban %d", successes, count)
	}
	return nil
}

func (s *assignItemScenario) failsWithAlreadyAssigned() error {
	if !errors.Is(s.err, backlog.ErrItemAlreadyAssigned) {
		return fmt.Errorf("error = %v, se esperaba ErrItemAlreadyAssigned", s.err)
	}
	return nil
}

func (s *assignItemScenario) failsWithSprintClosed() error {
	if !errors.Is(s.err, ErrSprintClosed) {
		return fmt.Errorf("error = %v, se esperaba ErrSprintClosed", s.err)
	}
	return nil
}

func (s *assignItemScenario) failsWithItemFromAnotherProject() error {
	if !errors.Is(s.err, ErrItemFromAnotherProject) {
		return fmt.Errorf("error = %v, se esperaba ErrItemFromAnotherProject", s.err)
	}
	return nil
}

func (s *assignItemScenario) failsWithSprintNotFound() error {
	if !errors.Is(s.err, ErrSprintNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrSprintNotFound", s.err)
	}
	return nil
}

func (s *assignItemScenario) failsWithItemNotFound() error {
	if !errors.Is(s.err, backlog.ErrItemNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrItemNotFound", s.err)
	}
	return nil
}
