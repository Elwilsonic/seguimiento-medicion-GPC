package backlog

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-06-listar-backlog.feature con godog.
func TestFeatures_HU06ListarBacklog(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeListBacklogScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-06-listar-backlog.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-06")
	}
}

// listItemScenario guarda el estado de un escenario de HU-06.
type listItemScenario struct {
	projectRepo *project.MemoryRepository
	projects    *project.Manager
	repo        *MemoryRepository
	manager     *Manager

	projectIDs     map[string]int
	listed         []Item
	err            error
	concurrentErrs []error
}

func initializeListBacklogScenario(ctx *godog.ScenarioContext) {
	s := &listItemScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projectRepo = project.NewMemoryRepository()
		s.projects = project.NewManager(s.projectRepo)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects)
		s.projectIDs = map[string]int{}
		s.listed = nil
		s.err = nil
		s.concurrentErrs = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un proyecto finalizado con nombre "([^"]*)"$`, s.existingFinishedProject)
	ctx.Step(`^que no hay proyectos creados$`, s.noProjects)
	ctx.Step(`^creo el ítem "([^"]*)" en el backlog de "([^"]*)" con prioridad "([^"]*)" y criterio "([^"]*)"$`, s.createItem)
	ctx.Step(`^listo el backlog del proyecto "([^"]*)"$`, s.listByProjectName)
	ctx.Step(`^listo el backlog del proyecto con ID (-?\d+)$`, s.listByProjectID)
	ctx.Step(`^obtengo (\d+) ítems?$`, s.listedCount)
	ctx.Step(`^obtengo una lista vacía$`, s.listedEmptyList)
	ctx.Step(`^obtengo los ítems con los IDs (.*)$`, s.listedWithIDs)
	ctx.Step(`^el ítem en la posición (\d+) tiene el título "([^"]*)" y la prioridad "([^"]*)"$`, s.listedPositionHasTitleAndPriority)
	ctx.Step(`^el ítem en la posición (\d+) tiene el criterio "([^"]*)"$`, s.listedPositionHasCriterion)
	ctx.Step(`^modifico el primer ítem de la lista devuelta y sus criterios$`, s.mutateFirstListedItem)
	ctx.Step(`^al volver a listar el backlog de "([^"]*)" obtengo (\d+) ítems ordenados por ID ascendente$`, s.relistOrderedByID)
	ctx.Step(`^se crean (\d+) ítems en simultáneo en el proyecto "([^"]*)" mientras se listan en simultáneo su backlog$`, s.concurrentCreatesAndLists)
	ctx.Step(`^hay (\d+) ítems guardados$`, s.savedCount)
	ctx.Step(`^ningún listado simultáneo devuelve un error$`, s.noConcurrentListErrors)
	ctx.Step(`^el listado falla porque el proyecto no existe$`, s.listFailsProjectNotFound)
}

var idNumbers = regexp.MustCompile(`-?\d+`)

// parseIDs extrae los IDs numéricos del paso BDD, por ejemplo "1 y 3".
func parseIDs(raw string) []int {
	matches := idNumbers.FindAllString(raw, -1)
	ids := make([]int, 0, len(matches))
	for _, match := range matches {
		id, err := strconv.Atoi(match)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func (s *listItemScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *listItemScenario) existingFinishedProject(name string) error {
	p, err := s.projectRepo.Save(project.Project{Name: name, Status: project.StatusFinished})
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *listItemScenario) noProjects() error {
	return nil
}

func (s *listItemScenario) projectIDOf(name string) (int, error) {
	projectID, ok := s.projectIDs[name]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", name)
	}
	return projectID, nil
}

func (s *listItemScenario) createItem(title, projectName, priority, criterion string) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	_, err = s.manager.Create(projectID, title, "", priority, []string{criterion})
	return err
}

func (s *listItemScenario) listByProjectName(projectName string) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	s.listed, s.err = s.manager.ListItems(projectID)
	return nil
}

func (s *listItemScenario) listByProjectID(projectID int) error {
	s.listed, s.err = s.manager.ListItems(projectID)
	return nil
}

func (s *listItemScenario) listedOK() error {
	if s.err != nil {
		return fmt.Errorf("el listado falló: %v", s.err)
	}
	return nil
}

func (s *listItemScenario) listedCount(want int) error {
	if err := s.listedOK(); err != nil {
		return err
	}
	if len(s.listed) != want {
		return fmt.Errorf("el listado devolvió %d ítems, se esperaban %d", len(s.listed), want)
	}
	return nil
}

func (s *listItemScenario) listedEmptyList() error {
	if err := s.listedOK(); err != nil {
		return err
	}
	if s.listed == nil {
		return errors.New("la lista devuelta es nil, se esperaba una lista vacía")
	}
	if len(s.listed) != 0 {
		return fmt.Errorf("la lista devuelta tiene %d ítems, se esperaban 0", len(s.listed))
	}
	return nil
}

func (s *listItemScenario) listedWithIDs(raw string) error {
	if err := s.listedOK(); err != nil {
		return err
	}
	want := parseIDs(raw)
	got := make([]int, len(s.listed))
	for i, item := range s.listed {
		got[i] = item.ID
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("IDs = %v, se esperaban %v", got, want)
	}
	return nil
}

func (s *listItemScenario) listedPositionHasTitleAndPriority(position int, title, priority string) error {
	if err := s.listedOK(); err != nil {
		return err
	}
	if position < 1 || position > len(s.listed) {
		return fmt.Errorf("no hay ítem en la posición %d", position)
	}
	item := s.listed[position-1]
	if item.Title != title {
		return fmt.Errorf("título = %q, se esperaba %q", item.Title, title)
	}
	if string(item.Priority) != priority {
		return fmt.Errorf("prioridad = %q, se esperaba %q", item.Priority, priority)
	}
	return nil
}

func (s *listItemScenario) listedPositionHasCriterion(position int, criterion string) error {
	if err := s.listedOK(); err != nil {
		return err
	}
	if position < 1 || position > len(s.listed) {
		return fmt.Errorf("no hay ítem en la posición %d", position)
	}
	item := s.listed[position-1]
	if len(item.AcceptanceCriteria) != 1 || item.AcceptanceCriteria[0] != criterion {
		return fmt.Errorf("criterios = %q, se esperaban %q", item.AcceptanceCriteria, []string{criterion})
	}
	return nil
}

func (s *listItemScenario) mutateFirstListedItem() error {
	if err := s.listedOK(); err != nil {
		return err
	}
	if len(s.listed) == 0 {
		return errors.New("no hay ítems en la lista devuelta para modificar")
	}
	if len(s.listed[0].AcceptanceCriteria) == 0 {
		return errors.New("el primer ítem no tiene criterios para modificar")
	}
	s.listed[0].Title = "Otro título"
	s.listed[0].Priority = "BAJA"
	s.listed[0].AcceptanceCriteria[0] = "Otro criterio"
	return nil
}

func (s *listItemScenario) relistOrderedByID(projectName string, want int) error {
	projectID, err := s.projectIDOf(projectName)
	if err != nil {
		return err
	}
	s.listed, s.err = s.manager.ListItems(projectID)
	if err := s.listedOK(); err != nil {
		return err
	}
	if len(s.listed) != want {
		return fmt.Errorf("el listado devolvió %d ítems, se esperaban %d", len(s.listed), want)
	}
	for i := 1; i < len(s.listed); i++ {
		if s.listed[i-1].ID >= s.listed[i].ID {
			return fmt.Errorf("el listado no está ordenado por ID ascendente: %d y %d", s.listed[i-1].ID, s.listed[i].ID)
		}
	}
	return nil
}

func (s *listItemScenario) concurrentCreatesAndLists(attempts int, projectName string) error {
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
			_, s.concurrentErrs[i] = s.manager.Create(projectID, fmt.Sprintf("Login de usuarios %d", i), "", "ALTA", []string{"Login OK"})
		}()
		go func() {
			defer wg.Done()
			_, s.concurrentErrs[attempts+i] = s.manager.ListItems(projectID)
		}()
	}
	wg.Wait()
	return nil
}

func (s *listItemScenario) savedCount(want int) error {
	if len(s.repo.items) != want {
		return fmt.Errorf("hay %d ítems guardados, se esperaban %d", len(s.repo.items), want)
	}
	return nil
}

func (s *listItemScenario) noConcurrentListErrors() error {
	for i, err := range s.concurrentErrs {
		if err != nil {
			return fmt.Errorf("la operación simultánea %d falló: %v", i+1, err)
		}
	}
	return nil
}

func (s *listItemScenario) listFailsProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrProjectNotFound", s.err)
	}
	if s.listed != nil {
		return fmt.Errorf("se devolvió una lista junto con el error: %v", s.listed)
	}
	return nil
}
