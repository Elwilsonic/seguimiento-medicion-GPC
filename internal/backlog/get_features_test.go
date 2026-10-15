package backlog

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-05-consultar-item.feature con godog.
func TestFeatures_HU05ConsultarItem(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeGetItemScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-05-consultar-item.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-05")
	}
}

// getItemScenario guarda el estado de un escenario de HU-05.
type getItemScenario struct {
	projectRepo *project.MemoryRepository
	projects    *project.Manager
	repo        *MemoryRepository
	manager     *Manager

	projectIDs map[string]int
	fetched    Item
	err        error

	concurrentErrs   []error
	concurrentIDs    []int
	concurrentTitles []string
}

func initializeGetItemScenario(ctx *godog.ScenarioContext) {
	s := &getItemScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projectRepo = project.NewMemoryRepository()
		s.projects = project.NewManager(s.projectRepo)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects)
		s.projectIDs = map[string]int{}
		s.fetched = Item{}
		s.err = nil
		s.concurrentErrs = nil
		s.concurrentIDs = nil
		s.concurrentTitles = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^creo el ítem "([^"]*)" en el backlog de "([^"]*)" con prioridad "([^"]*)" y criterio "([^"]*)"$`, s.createItem)
	ctx.Step(`^consulto el ítem con ID (-?\d+)$`, s.consultItem)
	ctx.Step(`^vuelvo a consultar el ítem con ID (-?\d+)$`, s.consultItem)
	ctx.Step(`^obtengo un ítem con el ID (\d+) del proyecto "([^"]*)"$`, s.fetchedWithIDOfProject)
	ctx.Step(`^el ítem consultado tiene el título "([^"]*)"$`, s.fetchedHasTitle)
	ctx.Step(`^el ítem consultado tiene la descripción "([^"]*)"$`, s.fetchedHasDescription)
	ctx.Step(`^el ítem consultado tiene la prioridad "([^"]*)"$`, s.fetchedHasPriority)
	ctx.Step(`^el ítem consultado tiene el criterio "([^"]*)"$`, s.fetchedHasCriterion)
	ctx.Step(`^el ítem consultado queda en el estado "([^"]*)"$`, s.fetchedHasStatus)
	ctx.Step(`^el ítem consultado tiene los story points en (\d+)$`, s.fetchedHasStoryPoints)
	ctx.Step(`^modifico el título, la prioridad y los criterios del ítem consultado$`, s.mutateFetchedItem)
	ctx.Step(`^se crean (\d+) ítems en simultáneo en el proyecto "([^"]*)" mientras se consulta (\d+) veces el ítem con ID (\d+)$`, s.concurrentCreatesAndGets)
	ctx.Step(`^hay (\d+) ítems guardados$`, s.savedCount)
	ctx.Step(`^ninguna consulta simultánea devuelve un error$`, s.noConcurrentGetErrors)
	ctx.Step(`^todas las consultas simultáneas obtienen el ítem con ID (\d+) y el título "([^"]*)"$`, s.allConcurrentGetsHaveIDAndTitle)
	ctx.Step(`^la consulta falla porque el ítem no existe$`, s.getFailsItemNotFound)
	ctx.Step(`^no obtengo ningún ítem$`, s.noItemFetched)
}

func (s *getItemScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *getItemScenario) projectIDOf(name string) (int, error) {
	projectID, ok := s.projectIDs[name]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", name)
	}
	return projectID, nil
}

func (s *getItemScenario) createItem(title, projectName, priority, criterion string) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	_, err = s.manager.Create(projectID, title, "", priority, []string{criterion})
	return err
}

func (s *getItemScenario) consultItem(id int) error {
	s.fetched, s.err = s.manager.GetItem(id)
	return nil
}

func (s *getItemScenario) fetchedOK() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	return nil
}

func (s *getItemScenario) fetchedWithIDOfProject(id int, projectName string) error {
	if err := s.fetchedOK(); err != nil {
		return err
	}
	if s.fetched.ID != id {
		return fmt.Errorf("ID = %d, se esperaba %d", s.fetched.ID, id)
	}
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	if s.fetched.ProjectID != projectID {
		return fmt.Errorf("ProjectID = %d, se esperaba %d (proyecto %q)", s.fetched.ProjectID, projectID, projectName)
	}
	return nil
}

func (s *getItemScenario) fetchedHasTitle(title string) error {
	if err := s.fetchedOK(); err != nil {
		return err
	}
	if s.fetched.Title != title {
		return fmt.Errorf("título = %q, se esperaba %q", s.fetched.Title, title)
	}
	return nil
}

func (s *getItemScenario) fetchedHasDescription(description string) error {
	if err := s.fetchedOK(); err != nil {
		return err
	}
	if s.fetched.Description != description {
		return fmt.Errorf("descripción = %q, se esperaba %q", s.fetched.Description, description)
	}
	return nil
}

func (s *getItemScenario) fetchedHasPriority(priority string) error {
	if err := s.fetchedOK(); err != nil {
		return err
	}
	if string(s.fetched.Priority) != priority {
		return fmt.Errorf("prioridad = %q, se esperaba %q", s.fetched.Priority, priority)
	}
	return nil
}

func (s *getItemScenario) fetchedHasCriterion(criterion string) error {
	if err := s.fetchedOK(); err != nil {
		return err
	}
	if len(s.fetched.AcceptanceCriteria) != 1 || s.fetched.AcceptanceCriteria[0] != criterion {
		return fmt.Errorf("criterios = %q, se esperaban %q", s.fetched.AcceptanceCriteria, []string{criterion})
	}
	return nil
}

func (s *getItemScenario) fetchedHasStatus(status string) error {
	if err := s.fetchedOK(); err != nil {
		return err
	}
	if string(s.fetched.Status) != status {
		return fmt.Errorf("estado = %q, se esperaba %q", s.fetched.Status, status)
	}
	return nil
}

func (s *getItemScenario) fetchedHasStoryPoints(want int) error {
	if err := s.fetchedOK(); err != nil {
		return err
	}
	if s.fetched.StoryPoints != want {
		return fmt.Errorf("story points = %d, se esperaban %d", s.fetched.StoryPoints, want)
	}
	return nil
}

func (s *getItemScenario) mutateFetchedItem() error {
	if err := s.fetchedOK(); err != nil {
		return err
	}
	if len(s.fetched.AcceptanceCriteria) == 0 {
		return errors.New("el ítem consultado no tiene criterios para modificar")
	}
	s.fetched.Title = "Otro título"
	s.fetched.Priority = "BAJA"
	s.fetched.AcceptanceCriteria[0] = "Otro criterio"
	return nil
}

func (s *getItemScenario) concurrentCreatesAndGets(attempts int, projectName string, consults int, consultID int) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	s.concurrentErrs = make([]error, attempts+consults)
	s.concurrentIDs = make([]int, consults)
	s.concurrentTitles = make([]string, consults)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.concurrentErrs[i] = s.manager.Create(projectID, fmt.Sprintf("Login de usuarios %d", i), "", "ALTA", []string{"Login OK"})
		}()
	}
	for i := range consults {
		wg.Add(1)
		go func() {
			defer wg.Done()
			item, err := s.manager.GetItem(consultID)
			s.concurrentErrs[attempts+i] = err
			s.concurrentIDs[i] = item.ID
			s.concurrentTitles[i] = item.Title
		}()
	}
	wg.Wait()
	return nil
}

func (s *getItemScenario) savedCount(want int) error {
	if len(s.repo.items) != want {
		return fmt.Errorf("hay %d ítems guardados, se esperaban %d", len(s.repo.items), want)
	}
	return nil
}

func (s *getItemScenario) noConcurrentGetErrors() error {
	for i, err := range s.concurrentErrs {
		if err != nil {
			return fmt.Errorf("la operación simultánea %d falló: %v", i+1, err)
		}
	}
	return nil
}

func (s *getItemScenario) allConcurrentGetsHaveIDAndTitle(id int, title string) error {
	for i := range s.concurrentIDs {
		if s.concurrentIDs[i] != id {
			return fmt.Errorf("la consulta simultánea %d obtuvo el ID %d, se esperaba %d", i+1, s.concurrentIDs[i], id)
		}
		if s.concurrentTitles[i] != title {
			return fmt.Errorf("la consulta simultánea %d obtuvo el título %q, se esperaba %q", i+1, s.concurrentTitles[i], title)
		}
	}
	return nil
}

func (s *getItemScenario) getFailsItemNotFound() error {
	if !errors.Is(s.err, ErrItemNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrItemNotFound", s.err)
	}
	return nil
}

func (s *getItemScenario) noItemFetched() error {
	if !reflect.DeepEqual(s.fetched, Item{}) {
		return fmt.Errorf("se devolvió un ítem junto con el error: %+v", s.fetched)
	}
	return nil
}
