package defect

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-31-consultar-defectos.feature con
// godog. Los defectos de partida se guardan con Save del repositorio en
// memoria, porque registrarlos es HU-29 y cambiar su estado es HU-30.
func TestFeatures_HU31ConsultarDefectos(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeListScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-31-consultar-defectos.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-31")
	}
}

// listScenario guarda el estado de un escenario de HU-31.
type listScenario struct {
	projects   *project.Manager
	projectIDs map[string]int
	itemRepo   *backlog.MemoryRepository
	items      *backlog.Manager
	sprintRepo *sprint.MemoryRepository
	sprints    *sprint.Manager
	memory     *MemoryRepository
	manager    *Manager

	listed []Defect
	err    error
	errs   []error
}

func initializeListScenario(ctx *godog.ScenarioContext) {
	s := &listScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.projectIDs = map[string]int{}
		s.itemRepo = backlog.NewMemoryRepository()
		s.items = backlog.NewManager(s.itemRepo, s.projects)
		s.sprintRepo = sprint.NewMemoryRepository()
		s.sprints = sprint.NewManager(s.sprintRepo, s.projects, s.items)
		s.memory = NewMemoryRepository()
		s.manager = NewManager(s.memory, s.items, s.sprints)
		s.listed = nil
		s.err = nil
		s.errs = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un ítem "([^"]*)" en el proyecto "([^"]*)"$`, s.existingItem)
	ctx.Step(`^que existe un sprint (abierto|cerrado) "([^"]*)" en el proyecto "([^"]*)"$`, s.existingSprint)
	ctx.Step(`^que el sprint con el ID (\d+) está cerrado$`, s.closedSprint)
	ctx.Step(`^que existe un defecto del ítem con el ID (\d+) detectado en el sprint con el ID (\d+) en estado "([^"]*)"$`, s.existingDefect)
	ctx.Step(`^que existe un defecto del ítem con el ID (\d+) detectado en el sprint con el ID (\d+) y resuelto en el sprint con el ID (\d+)$`, s.existingResolvedDefect)
	ctx.Step(`^que el repositorio de defectos falla al listar$`, s.repoFails)
	ctx.Step(`^consulto los defectos del proyecto "([^"]*)"$`, s.listByProjectName)
	ctx.Step(`^consulto los defectos del proyecto con el ID (-?\d+)$`, s.listByProjectID)
	ctx.Step(`^consulto los defectos del sprint con el ID (-?\d+)$`, s.listBySprint)
	ctx.Step(`^consulto los defectos del proyecto "([^"]*)" con el estado "([^"]*)"$`, s.listByProjectNameWithStatus)
	ctx.Step(`^consulto los defectos del proyecto con el ID (-?\d+) con el estado "([^"]*)"$`, s.listByProjectIDWithStatus)
	ctx.Step(`^consulto los defectos del sprint con el ID (-?\d+) con el estado "([^"]*)"$`, s.listBySprintWithStatus)
	ctx.Step(`^consulto en simultáneo (\d+) veces los defectos del proyecto "([^"]*)" y del sprint con el ID (\d+) mientras registro (\d+) defectos y cambio el estado del defecto con el ID (\d+)$`, s.listConcurrently)
	ctx.Step(`^la consulta de defectos es exitosa$`, s.listOK)
	ctx.Step(`^se listan los defectos con los IDs "([^"]*)"$`, s.listedIDs)
	ctx.Step(`^la lista de defectos está vacía$`, s.listedEmpty)
	ctx.Step(`^los defectos listados están en los estados "([^"]*)"$`, s.listedStatuses)
	ctx.Step(`^cada defecto listado tiene todos sus datos$`, s.listedHaveAllFields)
	ctx.Step(`^la consulta de defectos falla porque no existe el proyecto$`, s.failsWithErr(project.ErrProjectNotFound))
	ctx.Step(`^la consulta de defectos falla porque no existe el sprint$`, s.failsWithErr(sprint.ErrSprintNotFound))
	ctx.Step(`^la consulta de defectos falla porque el estado no es válido$`, s.failsWithErr(ErrInvalidStatus))
	ctx.Step(`^la consulta de defectos falla con el error del repositorio$`, s.failsWithErr(errDefectRepository))
	ctx.Step(`^modifico todos los datos de los defectos listados$`, s.mutateListed)
	ctx.Step(`^los defectos guardados no cambiaron$`, s.storedUnchanged)
	ctx.Step(`^al volver a consultar los defectos del proyecto "([^"]*)" se listan los mismos defectos guardados$`, s.projectListEqualsStored)
	ctx.Step(`^al volver a consultar los defectos del sprint con el ID (\d+) se listan los mismos defectos guardados$`, s.sprintListEqualsStored)
	ctx.Step(`^todas las consultas de defectos son exitosas$`, s.allOK)
	ctx.Step(`^al final se listan (\d+) defectos del proyecto "([^"]*)"$`, s.finalProjectCount)
	ctx.Step(`^al final se listan (\d+) defectos del sprint con el ID (\d+)$`, s.finalSprintCount)
}

func (s *listScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *listScenario) projectID(name string) (int, error) {
	id, ok := s.projectIDs[name]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", name)
	}
	return id, nil
}

func (s *listScenario) existingItem(title, projectName string) error {
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
		Status:             backlog.StatusInProgress,
	})
	return err
}

// existingSprint guarda el sprint con el repositorio, así un escenario puede
// tener más de un sprint abierto por proyecto.
func (s *listScenario) existingSprint(kind, name, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	status := sprint.StatusOpen
	if kind == "cerrado" {
		status = sprint.StatusClosed
	}
	_, err = s.sprintRepo.Save(sprint.Sprint{
		ProjectID: projectID,
		Name:      name,
		StartDate: time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC),
		Goal:      "Entregar el MVP",
		Status:    status,
	})
	return err
}

func (s *listScenario) closedSprint(sprintID int) error {
	sp, err := s.sprintRepo.GetByID(sprintID)
	if err != nil {
		return err
	}
	sp.Status = sprint.StatusClosed
	_, err = s.sprintRepo.Update(sp)
	return err
}

func (s *listScenario) saveDefect(itemID, detectionSprintID int, status Status, resolutionSprintID int) error {
	_, err := s.memory.Save(Defect{
		Description:        "El total no descuenta el descuento",
		Severity:           SeverityHigh,
		Status:             status,
		ItemID:             itemID,
		DetectionSprintID:  detectionSprintID,
		ResolutionSprintID: resolutionSprintID,
	})
	return err
}

// existingDefect guarda el defecto con Save. Un defecto RESUELTO queda
// resuelto en el mismo sprint de detección.
func (s *listScenario) existingDefect(itemID, detectionSprintID int, status string) error {
	resolutionSprintID := 0
	if Status(status) == StatusResolved {
		resolutionSprintID = detectionSprintID
	}
	return s.saveDefect(itemID, detectionSprintID, Status(status), resolutionSprintID)
}

func (s *listScenario) existingResolvedDefect(itemID, detectionSprintID, resolutionSprintID int) error {
	return s.saveDefect(itemID, detectionSprintID, StatusResolved, resolutionSprintID)
}

func (s *listScenario) repoFails() error {
	s.manager = NewManager(failingDefectRepository{err: errDefectRepository}, s.items, s.sprints)
	return nil
}

func (s *listScenario) listByProjectName(name string) error {
	id, err := s.projectID(name)
	if err != nil {
		return err
	}
	return s.listByProjectID(id)
}

func (s *listScenario) listByProjectID(id int) error {
	s.listed, s.err = s.manager.ListByProject(id)
	return nil
}

func (s *listScenario) listBySprint(id int) error {
	s.listed, s.err = s.manager.ListBySprint(id)
	return nil
}

// listWithStatus valida el estado antes de consultar, como el handler, y
// después filtra el resultado (D3).
func (s *listScenario) listWithStatus(status string, list func() ([]Defect, error)) error {
	s.listed = nil
	parsed, err := ParseStatus(status)
	if err != nil {
		s.err = err
		return nil
	}
	defects, err := list()
	if err != nil {
		s.err = err
		return nil
	}
	s.listed, s.err = FilterByStatus(defects, parsed), nil
	return nil
}

func (s *listScenario) listByProjectNameWithStatus(name, status string) error {
	id, err := s.projectID(name)
	if err != nil {
		return err
	}
	return s.listByProjectIDWithStatus(id, status)
}

func (s *listScenario) listByProjectIDWithStatus(id int, status string) error {
	return s.listWithStatus(status, func() ([]Defect, error) { return s.manager.ListByProject(id) })
}

func (s *listScenario) listBySprintWithStatus(id int, status string) error {
	return s.listWithStatus(status, func() ([]Defect, error) { return s.manager.ListBySprint(id) })
}

func (s *listScenario) listConcurrently(times int, projectName string, sprintID, registers, defectID int) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	record := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		s.errs = append(s.errs, err)
	}
	for i := 0; i < times; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := s.manager.ListByProject(projectID)
			record(err)
		}()
		go func() {
			defer wg.Done()
			_, err := s.manager.ListBySprint(sprintID)
			record(err)
		}()
	}
	for i := 0; i < registers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.manager.Register(1, "Defecto simultáneo "+strconv.Itoa(i), "ALTA", sprintID)
			record(err)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := s.manager.ChangeStatus(defectID, "EN_CORRECCION", 0)
		record(err)
	}()
	wg.Wait()
	return nil
}

func (s *listScenario) listOK() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	if s.listed == nil {
		return errors.New("la consulta devolvió una lista nil")
	}
	return nil
}

func (s *listScenario) listedIDs(ids string) error {
	if err := s.listOK(); err != nil {
		return err
	}
	var want []int
	for _, field := range strings.Split(ids, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil {
			return fmt.Errorf("ID de prueba inválido %q", field)
		}
		want = append(want, id)
	}
	got := make([]int, 0, len(s.listed))
	for _, defect := range s.listed {
		got = append(got, defect.ID)
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("IDs listados = %v, se esperaban %v", got, want)
	}
	return nil
}

func (s *listScenario) listedEmpty() error {
	if err := s.listOK(); err != nil {
		return err
	}
	if len(s.listed) != 0 {
		return fmt.Errorf("defectos listados = %+v, se esperaba una lista vacía", s.listed)
	}
	return nil
}

func (s *listScenario) listedStatuses(statuses string) error {
	var want []Status
	for _, field := range strings.Split(statuses, ",") {
		want = append(want, Status(strings.TrimSpace(field)))
	}
	got := make([]Status, 0, len(s.listed))
	for _, defect := range s.listed {
		got = append(got, defect.Status)
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("estados listados = %v, se esperaban %v", got, want)
	}
	return nil
}

// listedHaveAllFields compara cada defecto listado con el guardado.
func (s *listScenario) listedHaveAllFields() error {
	if err := s.listOK(); err != nil {
		return err
	}
	stored := memorySnapshot(s.memory)
	for _, defect := range s.listed {
		i := slices.IndexFunc(stored, func(d Defect) bool { return d.ID == defect.ID })
		if i == -1 {
			return fmt.Errorf("el defecto listado %d no está guardado", defect.ID)
		}
		if defect != stored[i] {
			return fmt.Errorf("defecto listado = %+v, se esperaba %+v", defect, stored[i])
		}
	}
	return nil
}

func (s *listScenario) failsWithErr(want error) func() error {
	return func() error {
		if !errors.Is(s.err, want) {
			return fmt.Errorf("err = %v, se esperaba %v", s.err, want)
		}
		if s.listed != nil {
			return fmt.Errorf("lista = %+v, se esperaba nil", s.listed)
		}
		return nil
	}
}

func (s *listScenario) mutateListed() error {
	if err := s.listOK(); err != nil {
		return err
	}
	if len(s.listed) == 0 {
		return errors.New("no hay defectos listados para modificar")
	}
	for i := range s.listed {
		s.listed[i].Description = "Descripción modificada"
		s.listed[i].Severity = SeverityLow
		s.listed[i].Status = StatusInCorrection
		s.listed[i].ItemID = 999
		s.listed[i].DetectionSprintID = 999
		s.listed[i].ResolutionSprintID = 999
	}
	return nil
}

func (s *listScenario) storedUnchanged() error {
	for _, defect := range memorySnapshot(s.memory) {
		if defect.Description == "Descripción modificada" || defect.ItemID == 999 {
			return fmt.Errorf("el defecto guardado cambió: %+v", defect)
		}
	}
	return nil
}

func (s *listScenario) equalsStored(got []Defect, err error) error {
	if err != nil {
		return fmt.Errorf("la consulta falló: %v", err)
	}
	if stored := memorySnapshot(s.memory); !slices.Equal(got, stored) {
		return fmt.Errorf("defectos listados = %+v, se esperaban %+v", got, stored)
	}
	return nil
}

func (s *listScenario) projectListEqualsStored(name string) error {
	id, err := s.projectID(name)
	if err != nil {
		return err
	}
	return s.equalsStored(s.manager.ListByProject(id))
}

func (s *listScenario) sprintListEqualsStored(id int) error {
	return s.equalsStored(s.manager.ListBySprint(id))
}

func (s *listScenario) allOK() error {
	for _, err := range s.errs {
		if err != nil {
			return fmt.Errorf("una operación simultánea falló: %v", err)
		}
	}
	return nil
}

func (s *listScenario) finalCount(want int, got []Defect, err error) error {
	if err != nil {
		return fmt.Errorf("la consulta falló: %v", err)
	}
	if len(got) != want {
		return fmt.Errorf("defectos listados = %d, se esperaban %d", len(got), want)
	}
	return nil
}

func (s *listScenario) finalProjectCount(want int, name string) error {
	id, err := s.projectID(name)
	if err != nil {
		return err
	}
	got, err := s.manager.ListByProject(id)
	return s.finalCount(want, got, err)
}

func (s *listScenario) finalSprintCount(want, sprintID int) error {
	got, err := s.manager.ListBySprint(sprintID)
	return s.finalCount(want, got, err)
}
