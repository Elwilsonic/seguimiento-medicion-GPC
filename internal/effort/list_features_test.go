package effort

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-27-consultar-esfuerzos.feature
// con godog. Los pasos viven en el paquete effort porque la historia es de E7.
func TestFeatures_HU27ConsultarEsfuerzos(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeListScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-27-consultar-esfuerzos.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-27")
	}
}

// listEffortScenario guarda el estado de un escenario de HU-27.
type listEffortScenario struct {
	projects   *project.Manager
	projectID  int
	itemRepo   *backlog.MemoryRepository
	items      *backlog.Manager
	sprintRepo *sprint.MemoryRepository
	sprints    *sprint.Manager
	memory     *MemoryRepository
	manager    *Manager

	summary Summary
	err     error
	errs    []error
}

func initializeListScenario(ctx *godog.ScenarioContext) {
	s := &listEffortScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.itemRepo = backlog.NewMemoryRepository()
		s.items = backlog.NewManager(s.itemRepo, s.projects)
		s.sprintRepo = sprint.NewMemoryRepository()
		s.sprints = sprint.NewManager(s.sprintRepo, s.projects, s.items)
		s.memory = NewMemoryRepository()
		s.manager = NewManager(s.memory, s.sprints, s.items)
		s.projectID = 0
		s.summary = Summary{}
		s.err = nil
		s.errs = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un ítem "([^"]*)" en el proyecto "([^"]*)"$`, s.existingItem)
	ctx.Step(`^que existe un sprint (abierto|cerrado) "([^"]*)" en el proyecto "([^"]*)"$`, s.existingSprint)
	ctx.Step(`^que el ítem con el ID (-?\d+) está asignado al sprint con el ID (-?\d+)$`, s.assignedItem)
	ctx.Step(`^que existe un esfuerzo de ([0-9.]+) horas del integrante "([^"]*)" con actividad "([^"]*)" para el ítem con el ID (-?\d+) del sprint con el ID (-?\d+)$`, s.seedEffort)
	ctx.Step(`^que el repositorio de esfuerzos falla al listar$`, s.repoFailsOnList)
	ctx.Step(`^consulto los esfuerzos del sprint con el ID (-?\d+)$`, s.querySprint)
	ctx.Step(`^consulto los esfuerzos del ítem con el ID (-?\d+)$`, s.queryItem)
	ctx.Step(`^vuelvo a consultar los esfuerzos del sprint con el ID (-?\d+)$`, s.querySprint)
	ctx.Step(`^la consulta es exitosa$`, s.queryOK)
	ctx.Step(`^se listan (\d+) esfuerzos$`, s.listedCount)
	ctx.Step(`^los esfuerzos listados tienen IDs ([\d, ]+)$`, s.listedIDs)
	ctx.Step(`^el total de horas es ([0-9.]+)$`, s.totalHours)
	ctx.Step(`^el primer esfuerzo listado tiene integrante "([^"]*)" y actividad "([^"]*)"$`, s.firstEffort)
	ctx.Step(`^la lista de esfuerzos no es nil$`, s.listNotNil)
	ctx.Step(`^no se devuelve ningún esfuerzo$`, s.noEfforts)
	ctx.Step(`^la consulta falla porque no existe el sprint$`, s.failsSprintNotFound)
	ctx.Step(`^la consulta falla porque no existe el ítem$`, s.failsItemNotFound)
	ctx.Step(`^la consulta falla con el error del repositorio$`, s.failsRepository)
	ctx.Step(`^no se devuelve ningún resumen$`, s.noSummary)
	ctx.Step(`^modifico el primer esfuerzo devuelto$`, s.modifyFirst)
	ctx.Step(`^los esfuerzos guardados no cambian$`, s.savedUnchanged)
	ctx.Step(`^registro (\d+) esfuerzos y consulto en simultáneo para el ítem con el ID (-?\d+) y el sprint con el ID (-?\d+)$`, s.registerAndListConcurrently)
	ctx.Step(`^todos los registros y listados son exitosos$`, s.allConcurrentOK)
	ctx.Step(`^quedan (\d+) esfuerzos guardados$`, s.remainingEfforts)
	ctx.Step(`^el total final es ([0-9.]+)$`, s.finalTotal)
}

func (s *listEffortScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectID = p.ID
	return nil
}

func (s *listEffortScenario) existingItem(title, projectName string) error {
	if s.projectID == 0 {
		return fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	_, err := s.itemRepo.Save(backlog.Item{
		ProjectID:          s.projectID,
		Title:              title,
		Description:        "Formulario de acceso",
		Priority:           backlog.PriorityHigh,
		AcceptanceCriteria: []string{"Login OK"},
		Status:             backlog.StatusInProgress,
		StoryPoints:        5,
		EstimatedHours:     2.5,
	})
	return err
}

func (s *listEffortScenario) existingSprint(kind, name, projectName string) error {
	if s.projectID == 0 {
		return fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	status := sprint.StatusOpen
	if kind == "cerrado" {
		status = sprint.StatusClosed
	}
	_, err := s.sprintRepo.Save(sprint.Sprint{
		ProjectID: s.projectID,
		Name:      name,
		StartDate: contractDate,
		EndDate:   contractDate.AddDate(0, 0, 6),
		Goal:      "Entregar el MVP",
		Status:    status,
	})
	return err
}

func (s *listEffortScenario) assignedItem(itemID, sprintID int) error {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return err
	}
	item.SprintID = sprintID
	_, err = s.itemRepo.Update(item)
	return err
}

func (s *listEffortScenario) seedEffort(hoursRaw, member, activity string, itemID, sprintID int) error {
	hours, err := strconv.ParseFloat(hoursRaw, 64)
	if err != nil {
		return fmt.Errorf("horas de prueba inválidas %q: %w", hoursRaw, err)
	}
	_, err = s.memory.Save(Effort{
		Member:   member,
		Date:     contractDate,
		Activity: activity,
		ItemID:   itemID,
		SprintID: sprintID,
		Hours:    hours,
	})
	return err
}

func (s *listEffortScenario) repoFailsOnList() error {
	s.manager = NewManager(failingEffortRepository{err: errEffortRepository}, s.sprints, s.items)
	return nil
}

func (s *listEffortScenario) querySprint(id int) error {
	s.summary, s.err = s.manager.ListBySprint(id)
	return nil
}

func (s *listEffortScenario) queryItem(id int) error {
	s.summary, s.err = s.manager.ListByItem(id)
	return nil
}

func (s *listEffortScenario) queryOK() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	return nil
}

func (s *listEffortScenario) listedCount(count int) error {
	if len(s.summary.Efforts) != count {
		return fmt.Errorf("esfuerzos listados = %d, se esperaban %d", len(s.summary.Efforts), count)
	}
	return nil
}

func (s *listEffortScenario) listedIDs(idsRaw string) error {
	parts := strings.Split(idsRaw, ",")
	want := make([]int, len(parts))
	for i, part := range parts {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return fmt.Errorf("ID de prueba inválido %q: %w", part, err)
		}
		want[i] = id
	}
	if len(s.summary.Efforts) != len(want) {
		return fmt.Errorf("esfuerzos listados = %d, se esperaban %d", len(s.summary.Efforts), len(want))
	}
	for i, id := range want {
		if s.summary.Efforts[i].ID != id {
			return fmt.Errorf("ID en la posición %d = %d, se esperaba %d", i, s.summary.Efforts[i].ID, id)
		}
	}
	return nil
}

func (s *listEffortScenario) totalHours(totalRaw string) error {
	want, err := strconv.ParseFloat(totalRaw, 64)
	if err != nil {
		return fmt.Errorf("total de prueba inválido %q: %w", totalRaw, err)
	}
	if math.Abs(s.summary.TotalHours-want) > 1e-9 {
		return fmt.Errorf("total de horas = %v, se esperaba %v", s.summary.TotalHours, want)
	}
	return nil
}

func (s *listEffortScenario) firstEffort(member, activity string) error {
	if len(s.summary.Efforts) == 0 {
		return fmt.Errorf("no se listó ningún esfuerzo")
	}
	got := s.summary.Efforts[0]
	if got.Member != member || got.Activity != activity {
		return fmt.Errorf("primer esfuerzo = %q/%q, se esperaba %q/%q", got.Member, got.Activity, member, activity)
	}
	return nil
}

func (s *listEffortScenario) listNotNil() error {
	if s.summary.Efforts == nil {
		return fmt.Errorf("la lista de esfuerzos es nil")
	}
	return nil
}

func (s *listEffortScenario) noEfforts() error {
	if len(s.summary.Efforts) != 0 {
		return fmt.Errorf("se listaron %d esfuerzos, se esperaba ninguno", len(s.summary.Efforts))
	}
	return nil
}

func (s *listEffortScenario) failsSprintNotFound() error {
	if s.err == nil || !errors.Is(s.err, sprint.ErrSprintNotFound) {
		return fmt.Errorf("error = %v, se esperaba %v", s.err, sprint.ErrSprintNotFound)
	}
	return nil
}

func (s *listEffortScenario) failsItemNotFound() error {
	if s.err == nil || !errors.Is(s.err, backlog.ErrItemNotFound) {
		return fmt.Errorf("error = %v, se esperaba %v", s.err, backlog.ErrItemNotFound)
	}
	return nil
}

func (s *listEffortScenario) failsRepository() error {
	if s.err == nil || !errors.Is(s.err, errEffortRepository) {
		return fmt.Errorf("error = %v, se esperaba %v", s.err, errEffortRepository)
	}
	return nil
}

func (s *listEffortScenario) noSummary() error {
	if s.summary.Efforts != nil || s.summary.TotalHours != 0 {
		return fmt.Errorf("se devolvió un resumen inesperado: %+v", s.summary)
	}
	return nil
}

func (s *listEffortScenario) modifyFirst() error {
	if len(s.summary.Efforts) == 0 {
		return fmt.Errorf("no hay esfuerzos para modificar")
	}
	s.summary.Efforts[0].Member = "Otra persona"
	s.summary.Efforts[0].Activity = "Otra actividad"
	s.summary.Efforts[0].Hours = 99
	return nil
}

func (s *listEffortScenario) savedUnchanged() error {
	if len(s.summary.Efforts) == 0 {
		return fmt.Errorf("no se listó ningún esfuerzo")
	}
	got := s.summary.Efforts[0]
	if got.Member != "García Santiago" || got.Activity != "Implementar el registro" || got.Hours != 2.5 {
		return fmt.Errorf("el esfuerzo guardado cambió: %+v", got)
	}
	return nil
}

func (s *listEffortScenario) registerAndListConcurrently(count, itemID, sprintID int) error {
	s.errs = make([]error, count*3)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(count * 3)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			_, s.errs[i] = s.manager.Register(sprintID, itemID, "García Santiago", contractDate, "Implementar el registro", 1)
		}(i)
		go func(i int) {
			defer wg.Done()
			<-start
			_, s.errs[count+i] = s.manager.ListByItem(itemID)
		}(i)
		go func(i int) {
			defer wg.Done()
			<-start
			_, s.errs[2*count+i] = s.manager.ListBySprint(sprintID)
		}(i)
	}
	close(start)
	wg.Wait()
	s.summary, s.err = s.manager.ListByItem(itemID)
	return nil
}

func (s *listEffortScenario) allConcurrentOK() error {
	for i, err := range s.errs {
		if err != nil {
			return fmt.Errorf("operación simultánea %d falló: %v", i, err)
		}
	}
	if s.err != nil {
		return fmt.Errorf("la consulta final falló: %v", s.err)
	}
	return nil
}

func (s *listEffortScenario) remainingEfforts(count int) error {
	if len(s.summary.Efforts) != count {
		return fmt.Errorf("esfuerzos guardados = %d, se esperaban %d", len(s.summary.Efforts), count)
	}
	return nil
}

func (s *listEffortScenario) finalTotal(totalRaw string) error {
	return s.totalHours(totalRaw)
}
