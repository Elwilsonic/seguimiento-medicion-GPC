package backlog

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-04-crear-item-backlog.feature con godog.
func TestFeatures_HU04CrearItem(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeCreateItemScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-04-crear-item-backlog.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-04")
	}
}

// createItemScenario guarda el estado de un escenario de HU-04.
type createItemScenario struct {
	projectRepo *project.MemoryRepository
	projects    *project.Manager
	repo        *MemoryRepository
	manager     *Manager

	projectIDs    map[string]int
	created       Item
	createdItems  []Item
	createErrs    []error
	inputCriteria []string
	err           error
	errs          []error
}

func initializeCreateItemScenario(ctx *godog.ScenarioContext) {
	s := &createItemScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projectRepo = project.NewMemoryRepository()
		s.projects = project.NewManager(s.projectRepo)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects)
		s.projectIDs = map[string]int{}
		s.created = Item{}
		s.createdItems = nil
		s.createErrs = nil
		s.inputCriteria = nil
		s.err = nil
		s.errs = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un proyecto finalizado con nombre "([^"]*)"$`, s.existingFinishedProject)
	ctx.Step(`^creo un ítem en el proyecto "([^"]*)" con título "([^"]*)", descripción "([^"]*)", prioridad "([^"]*)"(?: y (?:criterios de aceptación (.*)|sin criterios de aceptación))?$`, s.createItemInProject)
	ctx.Step(`^creo un ítem en el proyecto con ID (-?\d+), título "([^"]*)", descripción "([^"]*)", prioridad "([^"]*)"(?: y (?:criterios de aceptación (.*)|sin criterios de aceptación))?$`, s.createItemInProjectByID)
	ctx.Step(`^el ítem se crea con el ID (\d+)$`, s.createdWithID)
	ctx.Step(`^los ítems se crean con los IDs (\d+) y (\d+)$`, s.createdItemsWithIDs)
	ctx.Step(`^el ítem queda en el estado "([^"]*)"$`, s.createdWithStatus)
	ctx.Step(`^el ítem tiene los story points en (\d+)$`, s.createdWithStoryPoints)
	ctx.Step(`^el ítem tiene el título "([^"]*)"$`, s.createdWithTitle)
	ctx.Step(`^el ítem tiene la descripción "([^"]*)"$`, s.createdWithDescription)
	ctx.Step(`^el ítem tiene la prioridad "([^"]*)"$`, s.createdWithPriority)
	ctx.Step(`^el ítem tiene los criterios de aceptación (.*)$`, s.createdWithCriteria)
	ctx.Step(`^el ítem guardado conserva el título "([^"]*)"$`, s.savedWithTitle)
	ctx.Step(`^el ítem guardado conserva la prioridad "([^"]*)"$`, s.savedWithPriority)
	ctx.Step(`^el ítem guardado conserva los criterios de aceptación (.*)$`, s.savedWithCriteria)
	ctx.Step(`^modifico el título y los criterios del ítem devuelto$`, s.mutateCreatedItem)
	ctx.Step(`^modifico la lista de criterios que envié$`, s.mutateInputCriteria)
	ctx.Step(`^se crean (\d+) ítems en simultáneo en el proyecto "([^"]*)"$`, s.concurrentCreates)
	ctx.Step(`^hay (\d+) ítems guardados$`, s.savedCount)
	ctx.Step(`^los IDs de los ítems guardados son todos distintos$`, s.distinctIDs)
	ctx.Step(`^la creación falla porque el título es obligatorio$`, s.failsWithEmptyTitle)
	ctx.Step(`^la creación falla porque la prioridad debe ser ALTA, MEDIA o BAJA$`, s.failsWithInvalidPriority)
	ctx.Step(`^la creación falla porque el ítem necesita al menos un criterio de aceptación$`, s.failsWithNoAcceptanceCriteria)
	ctx.Step(`^la creación falla porque el proyecto no existe$`, s.failsWithProjectNotFound)
}

var quotedStrings = regexp.MustCompile(`"([^"]*)"`)

// parseCriteria extrae los textos entre comillas del paso BDD. Un paso "sin
// criterios de aceptación" llega con el texto vacío.
func parseCriteria(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	matches := quotedStrings.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return nil
	}
	criteria := make([]string, 0, len(matches))
	for _, m := range matches {
		criteria = append(criteria, m[1])
	}
	return criteria
}

// equalCriteria compara listas de criterios elemento por elemento, en orden.
func equalCriteria(got, want []string) error {
	if len(got) != len(want) {
		return fmt.Errorf("criterios = %q, se esperaban %q", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			return fmt.Errorf("criterios = %q, se esperaban %q", got, want)
		}
	}
	return nil
}

func (s *createItemScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *createItemScenario) existingFinishedProject(name string) error {
	p, err := s.projectRepo.Save(project.Project{Name: name, Status: project.StatusFinished})
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *createItemScenario) createItemInProject(name, title, description, priority, criteriaRaw string) error {
	projectID, ok := s.projectIDs[name]
	if !ok {
		return fmt.Errorf("no existe el proyecto de prueba %q", name)
	}
	return s.create(projectID, title, description, priority, criteriaRaw)
}

func (s *createItemScenario) createItemInProjectByID(projectID int, title, description, priority, criteriaRaw string) error {
	return s.create(projectID, title, description, priority, criteriaRaw)
}

func (s *createItemScenario) create(projectID int, title, description, priority, criteriaRaw string) error {
	criteria := parseCriteria(criteriaRaw)
	s.inputCriteria = criteria
	s.created, s.err = s.manager.Create(projectID, title, description, priority, criteria)
	s.createdItems = append(s.createdItems, s.created)
	s.createErrs = append(s.createErrs, s.err)
	return nil
}

func (s *createItemScenario) createdOK() error {
	if s.err != nil {
		return fmt.Errorf("la creación falló: %v", s.err)
	}
	return nil
}

func (s *createItemScenario) createdWithID(id int) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.ID != id {
		return fmt.Errorf("ID = %d, se esperaba %d", s.created.ID, id)
	}
	return nil
}

func (s *createItemScenario) createdItemsWithIDs(first, second int) error {
	if len(s.createdItems) != 2 {
		return fmt.Errorf("se crearon %d ítems, se esperaban 2", len(s.createdItems))
	}
	for i, err := range s.createErrs {
		if err != nil {
			return fmt.Errorf("la creación %d falló: %v", i+1, err)
		}
	}
	if s.createdItems[0].ID != first || s.createdItems[1].ID != second {
		return fmt.Errorf("IDs = %d y %d, se esperaban %d y %d",
			s.createdItems[0].ID, s.createdItems[1].ID, first, second)
	}
	return nil
}

func (s *createItemScenario) createdWithStatus(status string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if string(s.created.Status) != status {
		return fmt.Errorf("estado = %q, se esperaba %q", s.created.Status, status)
	}
	return nil
}

func (s *createItemScenario) createdWithStoryPoints(want int) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.StoryPoints != want {
		return fmt.Errorf("story points = %d, se esperaban %d", s.created.StoryPoints, want)
	}
	return nil
}

func (s *createItemScenario) createdWithTitle(title string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.Title != title {
		return fmt.Errorf("título = %q, se esperaba %q", s.created.Title, title)
	}
	return nil
}

func (s *createItemScenario) createdWithDescription(description string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.Description != description {
		return fmt.Errorf("descripción = %q, se esperaba %q", s.created.Description, description)
	}
	return nil
}

func (s *createItemScenario) createdWithPriority(priority string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if string(s.created.Priority) != priority {
		return fmt.Errorf("prioridad = %q, se esperaba %q", s.created.Priority, priority)
	}
	return nil
}

func (s *createItemScenario) createdWithCriteria(raw string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	return equalCriteria(s.created.AcceptanceCriteria, parseCriteria(raw))
}

// savedItem devuelve el ítem guardado con el ID del último ítem creado.
func (s *createItemScenario) savedItem() (Item, error) {
	for _, item := range s.repo.items {
		if item.ID == s.created.ID {
			return item, nil
		}
	}
	return Item{}, fmt.Errorf("no hay ningún ítem guardado con el ID %d", s.created.ID)
}

func (s *createItemScenario) savedWithTitle(title string) error {
	saved, err := s.savedItem()
	if err != nil {
		return err
	}
	if saved.Title != title {
		return fmt.Errorf("título guardado = %q, se esperaba %q", saved.Title, title)
	}
	return nil
}

func (s *createItemScenario) savedWithPriority(priority string) error {
	saved, err := s.savedItem()
	if err != nil {
		return err
	}
	if string(saved.Priority) != priority {
		return fmt.Errorf("prioridad guardada = %q, se esperaba %q", saved.Priority, priority)
	}
	return nil
}

func (s *createItemScenario) savedWithCriteria(raw string) error {
	saved, err := s.savedItem()
	if err != nil {
		return err
	}
	return equalCriteria(saved.AcceptanceCriteria, parseCriteria(raw))
}

func (s *createItemScenario) mutateCreatedItem() error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if len(s.created.AcceptanceCriteria) == 0 {
		return errors.New("el ítem creado no tiene criterios para modificar")
	}
	s.created.Title = "Otro título"
	s.created.AcceptanceCriteria[0] = "Otro criterio"
	return nil
}

func (s *createItemScenario) mutateInputCriteria() error {
	if len(s.inputCriteria) == 0 {
		return errors.New("no hay lista de criterios recibida para modificar")
	}
	s.inputCriteria[0] = "Otro criterio"
	return nil
}

func (s *createItemScenario) concurrentCreates(attempts int, projectName string) error {
	projectID, ok := s.projectIDs[projectName]
	if !ok {
		return fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	s.errs = make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.errs[i] = s.manager.Create(projectID, fmt.Sprintf("Login de usuarios %d", i), "", "ALTA", []string{"Login OK"})
		}()
	}
	wg.Wait()
	for i, err := range s.errs {
		if err != nil {
			return fmt.Errorf("la creación %d falló: %v", i+1, err)
		}
	}
	return nil
}

func (s *createItemScenario) savedCount(want int) error {
	if len(s.repo.items) != want {
		return fmt.Errorf("hay %d ítems guardados, se esperaban %d", len(s.repo.items), want)
	}
	return nil
}

func (s *createItemScenario) distinctIDs() error {
	ids := make(map[int]bool)
	for _, item := range s.repo.items {
		if ids[item.ID] {
			return fmt.Errorf("el ID %d está repetido", item.ID)
		}
		ids[item.ID] = true
	}
	return nil
}

func (s *createItemScenario) failsWithEmptyTitle() error {
	if !errors.Is(s.err, ErrEmptyTitle) {
		return fmt.Errorf("error = %v, se esperaba ErrEmptyTitle", s.err)
	}
	return nil
}

func (s *createItemScenario) failsWithInvalidPriority() error {
	if !errors.Is(s.err, ErrInvalidPriority) {
		return fmt.Errorf("error = %v, se esperaba ErrInvalidPriority", s.err)
	}
	return nil
}

func (s *createItemScenario) failsWithNoAcceptanceCriteria() error {
	if !errors.Is(s.err, ErrNoAcceptanceCriteria) {
		return fmt.Errorf("error = %v, se esperaba ErrNoAcceptanceCriteria", s.err)
	}
	return nil
}

func (s *createItemScenario) failsWithProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrProjectNotFound", s.err)
	}
	return nil
}
