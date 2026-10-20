package backlog

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-09-priorizar-backlog.feature con godog.
func TestFeatures_HU09PriorizarBacklog(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializePriorityScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-09-priorizar-backlog.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-09")
	}
}

// priorityScenario guarda el estado de un escenario de HU-09.
type priorityScenario struct {
	projectRepo *project.MemoryRepository
	projects    *project.Manager
	repo        *MemoryRepository
	manager     *Manager

	projectIDs     map[string]int
	prioritized    []Item
	listed         []Item
	err            error
	concurrentErrs []error
}

func initializePriorityScenario(ctx *godog.ScenarioContext) {
	s := &priorityScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projectRepo = project.NewMemoryRepository()
		s.projects = project.NewManager(s.projectRepo)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects)
		s.projectIDs = map[string]int{}
		s.prioritized = nil
		s.listed = nil
		s.err = nil
		s.concurrentErrs = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un proyecto finalizado con nombre "([^"]*)"$`, s.existingFinishedProject)
	ctx.Step(`^que hay (\d+) proyectos creados$`, s.existingProjects)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" y el repositorio de ítems falla al listar$`, s.existingProjectWithFailingRepo)
	ctx.Step(`^creo el ítem "([^"]*)" en el backlog de "([^"]*)" con prioridad "([^"]*)" y criterio "([^"]*)"$`, s.createItem)
	ctx.Step(`^creo (\d+) ítems en el backlog de "([^"]*)" con las prioridades en ciclo "([^"]*)", "([^"]*)" y "([^"]*)"$`, s.createCyclicItems)
	ctx.Step(`^guardo directo en el repositorio el ítem "([^"]*)" del proyecto "([^"]*)" con prioridad "([^"]*)"$`, s.saveItemDirectly)
	ctx.Step(`^el ítem con el ID (\d+) pasa al estado "([^"]*)"$`, s.changeItemStatus)
	ctx.Step(`^listo por prioridad el backlog del proyecto "([^"]*)"$`, s.listByPriorityByName)
	ctx.Step(`^vuelvo a listar por prioridad el backlog del proyecto "([^"]*)"$`, s.listByPriorityByName)
	ctx.Step(`^listo por prioridad el backlog del proyecto con ID (-?\d+)$`, s.listByPriorityByID)
	ctx.Step(`^listo el backlog de HU-06 del proyecto "([^"]*)"$`, s.listByID)
	ctx.Step(`^obtengo (\d+) ítems ordenados por prioridad$`, s.prioritizedCount)
	ctx.Step(`^el listado por prioridad es una lista vacía$`, s.prioritizedEmptyList)
	ctx.Step(`^obtengo los ítems ordenados por prioridad con los IDs (.*)$`, s.prioritizedWithIDs)
	ctx.Step(`^el ítem en la posición (\d+) del listado por prioridad tiene el título "([^"]*)", la prioridad "([^"]*)" y el criterio "([^"]*)"$`, s.prioritizedPositionHas)
	ctx.Step(`^el listado por prioridad tiene los mismos ítems que el listado de HU-06 del proyecto "([^"]*)"$`, s.sameItemsAsListItems)
	ctx.Step(`^el listado de HU-06 tiene los ítems con los IDs (.*)$`, s.listedWithIDs)
	ctx.Step(`^el listado por prioridad falla porque el proyecto no existe$`, s.failsProjectNotFound)
	ctx.Step(`^el listado por prioridad falla por un error del repositorio$`, s.failsRepository)
	ctx.Step(`^se crean (\d+) ítems en simultáneo en el proyecto "([^"]*)" mientras se lista por prioridad su backlog en simultáneo$`, s.concurrentCreatesAndPriorityLists)
	ctx.Step(`^ningún listado por prioridad simultáneo devuelve un error$`, s.noConcurrentErrors)
	ctx.Step(`^al volver a listar por prioridad el backlog de "([^"]*)" obtengo (\d+) ítems$`, s.relistByPriorityCount)
	ctx.Step(`^modifico el primer ítem del listado por prioridad, sus criterios y el orden de la lista$`, s.mutatePrioritized)
}

func (s *priorityScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *priorityScenario) existingFinishedProject(name string) error {
	p, err := s.projectRepo.Save(project.Project{Name: name, Status: project.StatusFinished})
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *priorityScenario) existingProjects(count int) error {
	for i := range count {
		if err := s.existingProject(fmt.Sprintf("Proyecto %d", i+1)); err != nil {
			return err
		}
	}
	return nil
}

// existingProjectWithFailingRepo arma el Manager sobre un repositorio que
// falla al listar, con un proyecto real (CA-09.8).
func (s *priorityScenario) existingProjectWithFailingRepo(name string) error {
	if err := s.existingProject(name); err != nil {
		return err
	}
	s.manager = NewManager(failingRepository{listErr: errRepository}, s.projects)
	return nil
}

func (s *priorityScenario) projectIDOf(name string) (int, error) {
	projectID, ok := s.projectIDs[name]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", name)
	}
	return projectID, nil
}

func (s *priorityScenario) createItem(title, projectName, priority, criterion string) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	_, err = s.manager.Create(projectID, title, "", priority, []string{criterion})
	return err
}

func (s *priorityScenario) createCyclicItems(count int, projectName, first, second, third string) error {
	priorities := []string{first, second, third}
	for i := range count {
		title := fmt.Sprintf("Ítem %d", i+1)
		if err := s.createItem(title, projectName, priorities[i%len(priorities)], "Criterio"); err != nil {
			return err
		}
	}
	return nil
}

// saveItemDirectly guarda un ítem sin pasar por el Manager, para tener una
// prioridad que Create no acepta (RN-09.3).
func (s *priorityScenario) saveItemDirectly(title, projectName, priority string) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	_, err = s.repo.Save(Item{
		ProjectID:          projectID,
		Title:              title,
		Priority:           Priority(priority),
		AcceptanceCriteria: []string{"Criterio"},
		Status:             StatusPending,
	})
	return err
}

func (s *priorityScenario) changeItemStatus(id int, status string) error {
	_, err := s.manager.ChangeStatus(id, status)
	return err
}

func (s *priorityScenario) listByPriorityByName(projectName string) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	s.prioritized, s.err = s.manager.ListItemsByPriority(projectID)
	return nil
}

func (s *priorityScenario) listByPriorityByID(projectID int) error {
	s.prioritized, s.err = s.manager.ListItemsByPriority(projectID)
	return nil
}

func (s *priorityScenario) listByID(projectName string) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	s.listed, err = s.manager.ListItems(projectID)
	return err
}

func (s *priorityScenario) prioritizedOK() error {
	if s.err != nil {
		return fmt.Errorf("el listado por prioridad falló: %v", s.err)
	}
	return nil
}

func (s *priorityScenario) prioritizedCount(want int) error {
	if err := s.prioritizedOK(); err != nil {
		return err
	}
	if len(s.prioritized) != want {
		return fmt.Errorf("el listado por prioridad devolvió %d ítems, se esperaban %d", len(s.prioritized), want)
	}
	return nil
}

func (s *priorityScenario) prioritizedEmptyList() error {
	if err := s.prioritizedOK(); err != nil {
		return err
	}
	if s.prioritized == nil {
		return errors.New("la lista devuelta es nil, se esperaba una lista vacía")
	}
	if len(s.prioritized) != 0 {
		return fmt.Errorf("la lista devuelta tiene %d ítems, se esperaban 0", len(s.prioritized))
	}
	return nil
}

// itemIDs devuelve los IDs de los ítems, en el orden de la lista.
func itemIDs(items []Item) []int {
	ids := make([]int, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	return ids
}

func (s *priorityScenario) prioritizedWithIDs(raw string) error {
	if err := s.prioritizedOK(); err != nil {
		return err
	}
	want := parseIDs(raw)
	if got := itemIDs(s.prioritized); !slices.Equal(got, want) {
		return fmt.Errorf("IDs = %v, se esperaban %v", got, want)
	}
	return nil
}

func (s *priorityScenario) prioritizedPositionHas(position int, title, priority, criterion string) error {
	if err := s.prioritizedOK(); err != nil {
		return err
	}
	if position < 1 || position > len(s.prioritized) {
		return fmt.Errorf("no hay ítem en la posición %d", position)
	}
	item := s.prioritized[position-1]
	if item.Title != title {
		return fmt.Errorf("título = %q, se esperaba %q", item.Title, title)
	}
	if string(item.Priority) != priority {
		return fmt.Errorf("prioridad = %q, se esperaba %q", item.Priority, priority)
	}
	if !slices.Equal(item.AcceptanceCriteria, []string{criterion}) {
		return fmt.Errorf("criterios = %q, se esperaban %q", item.AcceptanceCriteria, []string{criterion})
	}
	return nil
}

// sameItemsAsListItems compara el listado por prioridad con ListItems,
// ordenando una copia por ID (RN-09.4).
func (s *priorityScenario) sameItemsAsListItems(projectName string) error {
	if err := s.prioritizedOK(); err != nil {
		return err
	}
	if err := s.listByID(projectName); err != nil {
		return err
	}
	byID := slices.Clone(s.prioritized)
	slices.SortFunc(byID, func(a, b Item) int { return a.ID - b.ID })
	if len(byID) != len(s.listed) {
		return fmt.Errorf("el listado por prioridad tiene %d ítems y ListItems %d", len(byID), len(s.listed))
	}
	for i := range byID {
		if !reflect.DeepEqual(byID[i], s.listed[i]) {
			return fmt.Errorf("ítem %d = %+v, ListItems devolvió %+v", i+1, byID[i], s.listed[i])
		}
	}
	return nil
}

func (s *priorityScenario) listedWithIDs(raw string) error {
	want := parseIDs(raw)
	if got := itemIDs(s.listed); !slices.Equal(got, want) {
		return fmt.Errorf("IDs de ListItems = %v, se esperaban %v", got, want)
	}
	return nil
}

func (s *priorityScenario) failsProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrProjectNotFound", s.err)
	}
	if s.prioritized != nil {
		return fmt.Errorf("se devolvió una lista junto con el error: %v", s.prioritized)
	}
	return nil
}

func (s *priorityScenario) failsRepository() error {
	if !errors.Is(s.err, errRepository) {
		return fmt.Errorf("error = %v, se esperaba el error del repositorio", s.err)
	}
	if s.prioritized != nil {
		return fmt.Errorf("se devolvió una lista junto con el error: %v", s.prioritized)
	}
	return nil
}

func (s *priorityScenario) concurrentCreatesAndPriorityLists(attempts int, projectName string) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	s.concurrentErrs = make([]error, attempts*2)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(2)
		go func() {
			defer wg.Done()
			priority := []string{"BAJA", "ALTA", "MEDIA"}[i%3]
			_, s.concurrentErrs[i] = s.manager.Create(projectID, fmt.Sprintf("Ítem %d", i), "", priority, []string{"Criterio"})
		}()
		go func() {
			defer wg.Done()
			_, s.concurrentErrs[attempts+i] = s.manager.ListItemsByPriority(projectID)
		}()
	}
	wg.Wait()
	return nil
}

func (s *priorityScenario) noConcurrentErrors() error {
	for i, err := range s.concurrentErrs {
		if err != nil {
			return fmt.Errorf("la operación simultánea %d falló: %v", i+1, err)
		}
	}
	return nil
}

func (s *priorityScenario) relistByPriorityCount(projectName string, want int) error {
	if err := s.listByPriorityByName(projectName); err != nil {
		return err
	}
	return s.prioritizedCount(want)
}

func (s *priorityScenario) mutatePrioritized() error {
	if err := s.prioritizedOK(); err != nil {
		return err
	}
	if len(s.prioritized) < 2 {
		return errors.New("se necesitan al menos 2 ítems en el listado para modificarlo")
	}
	s.prioritized[0].Title = "Otro título"
	s.prioritized[0].Priority = PriorityLow
	s.prioritized[0].AcceptanceCriteria[0] = "Otro criterio"
	s.prioritized[0], s.prioritized[1] = s.prioritized[1], s.prioritized[0]
	return nil
}
