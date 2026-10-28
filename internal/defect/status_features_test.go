package defect

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-30-cambiar-estado-defecto.feature
// con godog. Los defectos de partida se guardan con Save del repositorio en
// memoria, porque registrarlos es HU-29.
func TestFeatures_HU30CambiarEstadoDefecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeStatusScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-30-cambiar-estado-defecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-30")
	}
}

// statusScenario guarda el estado de un escenario de HU-30.
type statusScenario struct {
	projects   *project.Manager
	projectIDs map[string]int
	itemRepo   *backlog.MemoryRepository
	items      *backlog.Manager
	sprintRepo *sprint.MemoryRepository
	sprints    *sprint.Manager
	memory     *MemoryRepository
	counter    *countingDefectRepository
	manager    *Manager

	before       map[int]Defect
	beforeItem   backlog.Item
	beforeSprint sprint.Sprint

	changed  Defect
	err      error
	closeErr error
	errs     []error
	log      *defectEventLog
}

func initializeStatusScenario(ctx *godog.ScenarioContext) {
	s := &statusScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.projectIDs = map[string]int{}
		s.itemRepo = backlog.NewMemoryRepository()
		s.items = backlog.NewManager(s.itemRepo, s.projects)
		s.sprintRepo = sprint.NewMemoryRepository()
		s.sprints = sprint.NewManager(s.sprintRepo, s.projects, s.items)
		s.memory = NewMemoryRepository()
		s.counter = &countingDefectRepository{Repository: s.memory}
		s.manager = NewManager(s.counter, s.items, s.sprints)
		s.before = map[int]Defect{}
		s.beforeItem = backlog.Item{}
		s.beforeSprint = sprint.Sprint{}
		s.changed = Defect{}
		s.err = nil
		s.closeErr = nil
		s.errs = nil
		s.log = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un ítem "([^"]*)" en el proyecto "([^"]*)"$`, s.existingItem)
	ctx.Step(`^que existe un sprint (abierto|cerrado) "([^"]*)" en el proyecto "([^"]*)"$`, s.existingSprint)
	ctx.Step(`^que el sprint con el ID (\d+) está cerrado$`, s.closedSprint)
	ctx.Step(`^que existe un defecto del ítem con el ID (\d+) detectado en el sprint con el ID (\d+) en estado "([^"]*)"$`, s.existingDefect)
	ctx.Step(`^que el repositorio de defectos falla al (buscar|actualizar)$`, s.repoFails)
	ctx.Step(`^cambio el estado del defecto con el ID (-?\d+) a "([^"]*)"$`, s.changeWithoutSprint)
	ctx.Step(`^cambio el estado del defecto con el ID (-?\d+) a "([^"]*)" sin sprint de resolución$`, s.changeWithoutSprint)
	ctx.Step(`^cambio el estado del defecto con el ID (-?\d+) a "([^"]*)" con el sprint de resolución con el ID (-?\d+)$`, s.change)
	ctx.Step(`^cambio en simultáneo (\d+) veces el estado del defecto con el ID (\d+) a "([^"]*)"$`, s.changeConcurrently)
	ctx.Step(`^cambio el estado del defecto con el ID (\d+) a "([^"]*)" con el sprint de resolución con el ID (\d+) y en simultáneo cierro el sprint con el ID (\d+)$`, s.changeAndClose)
	ctx.Step(`^el cambio de estado del defecto es exitoso$`, s.changeOK)
	ctx.Step(`^el defecto devuelto está en estado "([^"]*)"$`, s.returnedHasStatus)
	ctx.Step(`^el defecto guardado con el ID (\d+) está en estado "([^"]*)"$`, s.storedHasStatus)
	ctx.Step(`^el defecto guardado con el ID (\d+) tiene el sprint de resolución con el ID (\d+)$`, s.storedHasResolutionSprint)
	ctx.Step(`^el defecto guardado con el ID (\d+) no tiene sprint de resolución$`, s.storedHasNoResolutionSprint)
	ctx.Step(`^el defecto guardado con el ID (\d+) no cambió$`, s.storedUnchanged)
	ctx.Step(`^no se actualizó el repositorio de defectos$`, s.noUpdate)
	ctx.Step(`^el cambio de estado del defecto falla porque la transición no está permitida$`, s.failsWithErr(ErrInvalidTransition))
	ctx.Step(`^el cambio de estado del defecto falla porque el estado no es válido$`, s.failsWithErr(ErrInvalidStatus))
	ctx.Step(`^el cambio de estado del defecto falla porque no existe el defecto$`, s.failsWithErr(ErrDefectNotFound))
	ctx.Step(`^el cambio de estado del defecto falla porque el sprint de resolución es obligatorio$`, s.failsWithErr(ErrMissingResolutionSprint))
	ctx.Step(`^el cambio de estado del defecto falla porque no existe un sprint con ese id$`, s.failsWithErr(sprint.ErrSprintNotFound))
	ctx.Step(`^el cambio de estado del defecto falla porque el sprint está cerrado$`, s.failsWithErr(sprint.ErrSprintClosed))
	ctx.Step(`^el cambio de estado del defecto falla porque el sprint de resolución es de otro proyecto$`, s.failsWithErr(ErrResolutionSprintFromAnotherProject))
	ctx.Step(`^el cambio de estado del defecto falla porque el sprint de resolución es anterior al de detección$`, s.failsWithErr(ErrResolutionBeforeDetection))
	ctx.Step(`^el cambio de estado del defecto falla con el error del repositorio$`, s.failsWithErr(errDefectRepository))
	ctx.Step(`^el defecto guardado con el ID (\d+) conserva la descripción, la severidad, el ítem y el sprint de detección$`, s.storedKeepsOtherFields)
	ctx.Step(`^el ítem con el ID (\d+) y el sprint con el ID (\d+) no cambiaron$`, s.itemAndSprintUnchanged)
	ctx.Step(`^modifico todos los datos del defecto devuelto$`, s.mutateChanged)
	ctx.Step(`^todos los cambios de estado del defecto son exitosos$`, s.allChangesOK)
	ctx.Step(`^el sprint con el ID (\d+) queda en estado "([^"]*)"$`, s.sprintHasStatus)
	ctx.Step(`^el cambio simultáneo termina bien o porque el sprint está cerrado$`, s.changeOKOrClosed)
	ctx.Step(`^el defecto queda resuelto solo si el cambio terminó bien$`, s.resolvedOnlyIfOK)
}

func (s *statusScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *statusScenario) projectID(name string) (int, error) {
	id, ok := s.projectIDs[name]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", name)
	}
	return id, nil
}

func (s *statusScenario) existingItem(title, projectName string) error {
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
		StoryPoints:        5,
		EstimatedHours:     2.5,
	})
	return err
}

// existingSprint guarda el sprint con el repositorio, así un escenario puede
// tener más de un sprint abierto por proyecto.
func (s *statusScenario) existingSprint(kind, name, projectName string) error {
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

func (s *statusScenario) closedSprint(sprintID int) error {
	sp, err := s.sprintRepo.GetByID(sprintID)
	if err != nil {
		return err
	}
	sp.Status = sprint.StatusClosed
	_, err = s.sprintRepo.Update(sp)
	return err
}

// existingDefect guarda el defecto con Save. Un defecto RESUELTO queda
// resuelto en el mismo sprint de detección.
func (s *statusScenario) existingDefect(itemID, detectionSprintID int, status string) error {
	resolutionSprintID := 0
	if Status(status) == StatusResolved {
		resolutionSprintID = detectionSprintID
	}
	saved, err := s.memory.Save(Defect{
		Description:        "El total no descuenta el descuento",
		Severity:           SeverityHigh,
		Status:             Status(status),
		ItemID:             itemID,
		DetectionSprintID:  detectionSprintID,
		ResolutionSprintID: resolutionSprintID,
	})
	if err != nil {
		return err
	}
	s.before[saved.ID] = saved
	return nil
}

func (s *statusScenario) repoFails(operation string) error {
	if operation == "buscar" {
		s.counter.getErr = errDefectRepository
	} else {
		s.counter.updateErr = errDefectRepository
	}
	return nil
}

func (s *statusScenario) captureBefore() error {
	item, err := s.items.GetItem(1)
	if err != nil {
		return err
	}
	s.beforeItem = item
	sp, err := s.sprintRepo.GetByID(1)
	if err != nil {
		return err
	}
	s.beforeSprint = sp
	return nil
}

func (s *statusScenario) change(id int, status string, resolutionSprintID int) error {
	if err := s.captureBefore(); err != nil {
		return err
	}
	s.changed, s.err = s.manager.ChangeStatus(id, status, resolutionSprintID)
	return nil
}

func (s *statusScenario) changeWithoutSprint(id int, status string) error {
	return s.change(id, status, 0)
}

func (s *statusScenario) changeConcurrently(count, id int, status string) error {
	s.errs = make([]error, count)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			_, s.errs[i] = s.manager.ChangeStatus(id, status, 0)
		}(i)
	}
	close(start)
	wg.Wait()
	return nil
}

func (s *statusScenario) changeAndClose(id int, status string, resolutionSprintID, closeSprintID int) error {
	s.log = &defectEventLog{}
	sprintRepo := &recordingSprintRepository{Repository: s.sprintRepo, log: s.log}
	sprintManager := sprint.NewManager(sprintRepo, s.projects, s.items)
	s.counter.log = s.log
	manager := NewManager(s.counter, s.items, sprintManager)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		s.changed, s.err = manager.ChangeStatus(id, status, resolutionSprintID)
	}()
	go func() {
		defer wg.Done()
		<-start
		_, s.closeErr = sprintManager.Close(closeSprintID)
	}()
	close(start)
	wg.Wait()
	return nil
}

func (s *statusScenario) stored(id int) (Defect, error) {
	for _, defect := range memorySnapshot(s.memory) {
		if defect.ID == id {
			return defect, nil
		}
	}
	return Defect{}, fmt.Errorf("no hay un defecto guardado con el ID %d", id)
}

func (s *statusScenario) changeOK() error {
	if s.err != nil {
		return fmt.Errorf("el cambio de estado falló: %v", s.err)
	}
	return nil
}

func (s *statusScenario) returnedHasStatus(status string) error {
	if s.changed.Status != Status(status) {
		return fmt.Errorf("estado devuelto = %q, se esperaba %q", s.changed.Status, status)
	}
	return nil
}

func (s *statusScenario) storedHasStatus(id int, status string) error {
	defect, err := s.stored(id)
	if err != nil {
		return err
	}
	if defect.Status != Status(status) {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", defect.Status, status)
	}
	return nil
}

func (s *statusScenario) storedHasResolutionSprint(id, sprintID int) error {
	defect, err := s.stored(id)
	if err != nil {
		return err
	}
	if defect.ResolutionSprintID != sprintID {
		return fmt.Errorf("sprint de resolución guardado = %d, se esperaba %d", defect.ResolutionSprintID, sprintID)
	}
	return nil
}

func (s *statusScenario) storedHasNoResolutionSprint(id int) error {
	return s.storedHasResolutionSprint(id, 0)
}

func (s *statusScenario) storedUnchanged(id int) error {
	defect, err := s.stored(id)
	if err != nil {
		return err
	}
	if defect != s.before[id] {
		return fmt.Errorf("el defecto guardado cambió: antes %+v, después %+v", s.before[id], defect)
	}
	return nil
}

func (s *statusScenario) noUpdate() error {
	if n := s.counter.updateCount(); n != 0 {
		return fmt.Errorf("Update se llamó %d veces, se esperaba ninguna", n)
	}
	return nil
}

func (s *statusScenario) failsWithErr(target error) func() error {
	return func() error {
		if !errors.Is(s.err, target) {
			return fmt.Errorf("err = %v, se esperaba %v", s.err, target)
		}
		return nil
	}
}

func (s *statusScenario) storedKeepsOtherFields(id int) error {
	defect, err := s.stored(id)
	if err != nil {
		return err
	}
	before := s.before[id]
	if defect.ID != before.ID || defect.Description != before.Description || defect.Severity != before.Severity ||
		defect.ItemID != before.ItemID || defect.DetectionSprintID != before.DetectionSprintID {
		return fmt.Errorf("cambiaron otros campos: antes %+v, después %+v", before, defect)
	}
	return nil
}

func (s *statusScenario) itemAndSprintUnchanged(itemID, sprintID int) error {
	item, err := s.items.GetItem(itemID)
	if err != nil {
		return err
	}
	if !sameItem(item, s.beforeItem) {
		return fmt.Errorf("la historia cambió: antes %+v, después %+v", s.beforeItem, item)
	}
	sp, err := s.sprintRepo.GetByID(sprintID)
	if err != nil {
		return err
	}
	if sp != s.beforeSprint {
		return fmt.Errorf("el sprint cambió: antes %+v, después %+v", s.beforeSprint, sp)
	}
	return nil
}

func (s *statusScenario) mutateChanged() error {
	s.changed.Description = "Otra descripción"
	s.changed.Severity = SeverityLow
	s.changed.Status = StatusResolved
	s.changed.ItemID = 999
	s.changed.DetectionSprintID = 999
	s.changed.ResolutionSprintID = 999
	return nil
}

func (s *statusScenario) allChangesOK() error {
	for i, err := range s.errs {
		if err != nil {
			return fmt.Errorf("cambio %d falló: %v", i, err)
		}
	}
	return nil
}

func (s *statusScenario) sprintHasStatus(sprintID int, status string) error {
	if s.closeErr != nil {
		return fmt.Errorf("Close falló: %v", s.closeErr)
	}
	sp, err := s.sprintRepo.GetByID(sprintID)
	if err != nil {
		return err
	}
	if sp.Status != sprint.Status(status) {
		return fmt.Errorf("estado del sprint = %q, se esperaba %q", sp.Status, status)
	}
	return nil
}

func (s *statusScenario) changeOKOrClosed() error {
	if s.err != nil && !errors.Is(s.err, sprint.ErrSprintClosed) {
		return fmt.Errorf("el cambio falló con un error inesperado: %v", s.err)
	}
	return nil
}

func (s *statusScenario) resolvedOnlyIfOK() error {
	defect, err := s.stored(1)
	if err != nil {
		return err
	}
	if s.err == nil && defect.Status != StatusResolved {
		return fmt.Errorf("el cambio terminó bien pero el defecto quedó %+v", defect)
	}
	if s.err != nil && defect != s.before[1] {
		return fmt.Errorf("el cambio falló pero el defecto cambió: %+v", defect)
	}
	events := s.log.snapshot()
	closeIndex := slices.Index(events, "close")
	if closeIndex == -1 {
		return fmt.Errorf("no se registró el cierre: %v", events)
	}
	if slices.Contains(events[closeIndex+1:], "resolve") {
		return fmt.Errorf("se resolvió el defecto después del cierre: %v", events)
	}
	return nil
}
