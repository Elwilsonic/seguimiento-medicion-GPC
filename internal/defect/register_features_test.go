package defect

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-29-registrar-defecto.feature con
// godog. Los pasos viven en el paquete defect porque la historia orquesta
// defect, sprint y backlog.
func TestFeatures_HU29RegistrarDefecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeDefectScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-29-registrar-defecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-29")
	}
}

// defectScenario guarda el estado de un escenario de HU-29.
type defectScenario struct {
	projects   *project.Manager
	projectIDs map[string]int
	itemRepo   *backlog.MemoryRepository
	items      *backlog.Manager
	sprintRepo *sprint.MemoryRepository
	sprints    *sprint.Manager
	memory     *testDefectRepository
	repo       Repository
	manager    *Manager

	description, severity string
	itemID, sprintID      int
	beforeItem            backlog.Item
	hasBeforeItem         bool
	beforeSprint          sprint.Sprint
	hasBeforeSprint       bool

	saved    Defect
	err      error
	closeErr error
	errs     []error
	ids      []int
	count    int
	log      *defectEventLog
}

func initializeDefectScenario(ctx *godog.ScenarioContext) {
	s := &defectScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.projectIDs = map[string]int{}
		s.itemRepo = backlog.NewMemoryRepository()
		s.items = backlog.NewManager(s.itemRepo, s.projects)
		s.sprintRepo = sprint.NewMemoryRepository()
		s.sprints = sprint.NewManager(s.sprintRepo, s.projects, s.items)
		s.memory = &testDefectRepository{}
		s.repo = s.memory
		s.manager = NewManager(s.repo, s.items, s.sprints)
		s.description, s.severity = "", ""
		s.itemID, s.sprintID = 0, 0
		s.beforeItem = backlog.Item{}
		s.hasBeforeItem = false
		s.beforeSprint = sprint.Sprint{}
		s.hasBeforeSprint = false
		s.saved = Defect{}
		s.err = nil
		s.closeErr = nil
		s.errs = nil
		s.ids = nil
		s.count = 0
		s.log = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un ítem "([^"]*)" en estado "([^"]*)" en el proyecto "([^"]*)"$`, s.existingItem)
	ctx.Step(`^que existe un sprint (abierto|cerrado) "([^"]*)" en el proyecto "([^"]*)"$`, s.existingSprint)
	ctx.Step(`^que el ítem con el ID (-?\d+) está asignado al sprint con el ID (-?\d+)$`, s.assignedItem)
	ctx.Step(`^que el repositorio de defectos falla al guardar$`, s.repoFailsOnSave)
	ctx.Step(`^registro un defecto con descripción "([^"]*)", severidad "([^"]*)" para el ítem con el ID (-?\d+) en el sprint de detección con el ID (-?\d+)$`, s.register)
	ctx.Step(`^registro un defecto con descripción "([^"]*)", severidad "([^"]*)" para el ítem con el ID (-?\d+) sin sprint de detección$`, s.registerWithoutSprint)
	ctx.Step(`^registro en simultáneo (\d+) defectos con descripción "([^"]*)", severidad "([^"]*)" para el ítem con el ID (-?\d+) en el sprint de detección con el ID (-?\d+)$`, s.registerConcurrently)
	ctx.Step(`^registro un defecto con descripción "([^"]*)", severidad "([^"]*)" para el ítem con el ID (-?\d+) en el sprint de detección con el ID (-?\d+) y en simultáneo cierro el sprint con el ID (-?\d+)$`, s.registerAndClose)
	ctx.Step(`^el registro del defecto es exitoso$`, s.registerOK)
	ctx.Step(`^el defecto guardado tiene el ID (\d+)$`, s.savedHasID)
	ctx.Step(`^el defecto guardado tiene la descripción "([^"]*)"$`, s.savedHasDescription)
	ctx.Step(`^el defecto guardado tiene la severidad "([^"]*)"$`, s.savedHasSeverity)
	ctx.Step(`^el defecto guardado está en estado "([^"]*)"$`, s.savedHasStatus)
	ctx.Step(`^el defecto guardado tiene el ítem con el ID (\d+) y el sprint de detección con el ID (\d+)$`, s.savedHasItemAndSprint)
	ctx.Step(`^el defecto guardado no tiene sprint de resolución$`, s.savedHasNoResolutionSprint)
	ctx.Step(`^el ítem con el ID (-?\d+) sigue en estado "([^"]*)" y con los mismos datos$`, s.itemUnchanged)
	ctx.Step(`^el ítem con el ID (-?\d+) sigue asignado al sprint con el ID (-?\d+)$`, s.itemStillAssigned)
	ctx.Step(`^el sprint con el ID (-?\d+) sigue en estado "([^"]*)" y con los mismos datos$`, s.sprintUnchanged)
	ctx.Step(`^el registro del defecto falla porque el sprint de detección es obligatorio$`, s.failsMissingDetectionSprint)
	ctx.Step(`^el registro del defecto falla porque no existe un sprint con ese id$`, s.failsSprintNotFound)
	ctx.Step(`^el registro del defecto falla porque el sprint está cerrado$`, s.failsSprintClosed)
	ctx.Step(`^el registro del defecto falla porque la descripción es obligatoria$`, s.failsEmptyDescription)
	ctx.Step(`^el registro del defecto falla porque la severidad no es válida$`, s.failsInvalidSeverity)
	ctx.Step(`^el registro del defecto falla porque la historia no existe$`, s.failsItemNotFound)
	ctx.Step(`^el registro del defecto falla porque el sprint de detección es de otro proyecto$`, s.failsSprintFromAnotherProject)
	ctx.Step(`^el registro del defecto falla con el error del repositorio$`, s.failsRepository)
	ctx.Step(`^no se guardó ningún defecto$`, s.noDefectSaved)
	ctx.Step(`^modifico todos los datos del defecto devuelto$`, s.mutateSaved)
	ctx.Step(`^el defecto guardado conserva todos los datos del registro$`, s.savedKeepsInput)
	ctx.Step(`^todos los registros de defectos son exitosos$`, s.allRegistersOK)
	ctx.Step(`^los defectos guardados tienen IDs únicos$`, s.uniqueIDs)
	ctx.Step(`^el sprint con el ID (-?\d+) queda en estado "([^"]*)"$`, s.sprintHasStatus)
	ctx.Step(`^cada registro simultáneo de defecto termina bien o porque el sprint está cerrado$`, s.registerOKOrClosed)
	ctx.Step(`^ningún defecto se guarda después del cierre$`, s.noDefectAfterClose)
}

func (s *defectScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *defectScenario) projectID(name string) (int, error) {
	id, ok := s.projectIDs[name]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", name)
	}
	return id, nil
}

func (s *defectScenario) existingItem(title, status, projectName string) error {
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

func (s *defectScenario) existingSprint(kind, name, projectName string) error {
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

func (s *defectScenario) assignedItem(itemID, sprintID int) error {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return err
	}
	item.SprintID = sprintID
	_, err = s.itemRepo.Update(item)
	return err
}

func (s *defectScenario) repoFailsOnSave() error {
	s.repo = failingDefectRepository{err: errDefectRepository}
	s.manager = NewManager(s.repo, s.items, s.sprints)
	return nil
}

func (s *defectScenario) captureBefore(itemID, sprintID int) {
	if item, err := s.itemRepo.GetByID(itemID); err == nil {
		s.beforeItem = item
		s.hasBeforeItem = true
	}
	if sp, err := s.sprintRepo.GetByID(sprintID); err == nil {
		s.beforeSprint = sp
		s.hasBeforeSprint = true
	}
}

func (s *defectScenario) register(description, severity string, itemID, sprintID int) error {
	s.description, s.severity = description, severity
	s.itemID, s.sprintID = itemID, sprintID
	s.captureBefore(itemID, sprintID)
	s.saved, s.err = s.manager.Register(itemID, description, severity, sprintID)
	return nil
}

func (s *defectScenario) registerWithoutSprint(description, severity string, itemID int) error {
	return s.register(description, severity, itemID, 0)
}

func (s *defectScenario) registerConcurrently(count int, description, severity string, itemID, sprintID int) error {
	s.description, s.severity = description, severity
	s.itemID, s.sprintID = itemID, sprintID
	s.captureBefore(itemID, sprintID)
	s.count = count
	s.errs = make([]error, count)
	s.ids = make([]int, count)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			got, err := s.manager.Register(itemID, description, severity, sprintID)
			s.errs[i] = err
			s.ids[i] = got.ID
		}(i)
	}
	close(start)
	wg.Wait()
	return nil
}

func (s *defectScenario) registerAndClose(description, severity string, itemID, sprintID, closeSprintID int) error {
	s.description, s.severity = description, severity
	s.itemID, s.sprintID = itemID, sprintID
	s.captureBefore(itemID, sprintID)
	log := &defectEventLog{}
	sprintRepo := &recordingSprintRepository{Repository: s.sprintRepo, log: log}
	defectRepo := &recordingDefectRepository{testDefectRepository: s.memory, log: log}
	sprintManager := sprint.NewManager(sprintRepo, s.projects, s.items)
	manager := NewManager(defectRepo, s.items, sprintManager)

	start := make(chan struct{})
	var wg sync.WaitGroup
	var registerErr, closeErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, registerErr = manager.Register(itemID, description, severity, sprintID)
	}()
	go func() {
		defer wg.Done()
		<-start
		_, closeErr = sprintManager.Close(closeSprintID)
	}()
	close(start)
	wg.Wait()

	s.err = registerErr
	s.closeErr = closeErr
	s.log = log
	return nil
}

func (s *defectScenario) registerOK() error {
	if s.err != nil {
		return fmt.Errorf("el registro del defecto falló: %v", s.err)
	}
	return nil
}

func (s *defectScenario) savedDefect() (Defect, error) {
	saved := s.memory.snapshot()
	if len(saved) != 1 {
		return Defect{}, fmt.Errorf("defectos guardados = %d, se esperaba 1", len(saved))
	}
	return saved[0], nil
}

func (s *defectScenario) savedHasID(id int) error {
	defect, err := s.savedDefect()
	if err != nil {
		return err
	}
	if defect.ID != id {
		return fmt.Errorf("ID guardado = %d, se esperaba %d", defect.ID, id)
	}
	return nil
}

func (s *defectScenario) savedHasDescription(description string) error {
	defect, err := s.savedDefect()
	if err != nil {
		return err
	}
	if defect.Description != description {
		return fmt.Errorf("descripción guardada = %q, se esperaba %q", defect.Description, description)
	}
	return nil
}

func (s *defectScenario) savedHasSeverity(severity string) error {
	defect, err := s.savedDefect()
	if err != nil {
		return err
	}
	if string(defect.Severity) != severity {
		return fmt.Errorf("severidad guardada = %q, se esperaba %q", defect.Severity, severity)
	}
	return nil
}

func (s *defectScenario) savedHasStatus(status string) error {
	defect, err := s.savedDefect()
	if err != nil {
		return err
	}
	if string(defect.Status) != status {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", defect.Status, status)
	}
	return nil
}

func (s *defectScenario) savedHasItemAndSprint(itemID, sprintID int) error {
	defect, err := s.savedDefect()
	if err != nil {
		return err
	}
	if defect.ItemID != itemID || defect.DetectionSprintID != sprintID {
		return fmt.Errorf("ítem/sprint guardados = %d/%d, se esperaban %d/%d",
			defect.ItemID, defect.DetectionSprintID, itemID, sprintID)
	}
	return nil
}

func (s *defectScenario) savedHasNoResolutionSprint() error {
	defect, err := s.savedDefect()
	if err != nil {
		return err
	}
	if defect.ResolutionSprintID != 0 {
		return fmt.Errorf("sprint de resolución = %d, se esperaba 0", defect.ResolutionSprintID)
	}
	return nil
}

func (s *defectScenario) itemUnchanged(itemID int, status string) error {
	if !s.hasBeforeItem {
		return fmt.Errorf("no se guardó el ítem %d antes del registro", itemID)
	}
	item, err := s.items.GetItem(itemID)
	if err != nil {
		return err
	}
	if string(item.Status) != status {
		return fmt.Errorf("estado del ítem %d = %q, se esperaba %q", itemID, item.Status, status)
	}
	if !sameItem(item, s.beforeItem) {
		return fmt.Errorf("ítem modificado: antes %+v, después %+v", s.beforeItem, item)
	}
	return nil
}

func (s *defectScenario) itemStillAssigned(itemID, sprintID int) error {
	item, err := s.items.GetItem(itemID)
	if err != nil {
		return err
	}
	if item.SprintID != sprintID {
		return fmt.Errorf("sprint del ítem %d = %d, se esperaba %d", itemID, item.SprintID, sprintID)
	}
	return nil
}

func (s *defectScenario) sprintUnchanged(sprintID int, status string) error {
	if !s.hasBeforeSprint {
		return fmt.Errorf("no se guardó el sprint %d antes del registro", sprintID)
	}
	sp, err := s.sprintRepo.GetByID(sprintID)
	if err != nil {
		return err
	}
	if string(sp.Status) != status {
		return fmt.Errorf("estado del sprint %d = %q, se esperaba %q", sprintID, sp.Status, status)
	}
	if sp != s.beforeSprint {
		return fmt.Errorf("sprint modificado: antes %+v, después %+v", s.beforeSprint, sp)
	}
	return nil
}

func (s *defectScenario) failsWith(target error) error {
	if s.err == nil {
		return fmt.Errorf("se esperaba %v, no hubo error", target)
	}
	if !errors.Is(s.err, target) {
		return fmt.Errorf("error = %v, se esperaba %v", s.err, target)
	}
	return nil
}

func (s *defectScenario) failsMissingDetectionSprint() error {
	return s.failsWith(ErrMissingDetectionSprint)
}
func (s *defectScenario) failsSprintNotFound() error { return s.failsWith(sprint.ErrSprintNotFound) }
func (s *defectScenario) failsSprintClosed() error   { return s.failsWith(sprint.ErrSprintClosed) }
func (s *defectScenario) failsEmptyDescription() error {
	return s.failsWith(ErrEmptyDescription)
}
func (s *defectScenario) failsInvalidSeverity() error { return s.failsWith(ErrInvalidSeverity) }
func (s *defectScenario) failsItemNotFound() error    { return s.failsWith(backlog.ErrItemNotFound) }
func (s *defectScenario) failsSprintFromAnotherProject() error {
	return s.failsWith(ErrDetectionSprintFromAnotherProject)
}
func (s *defectScenario) failsRepository() error { return s.failsWith(errDefectRepository) }

func (s *defectScenario) noDefectSaved() error {
	if got := s.memory.count(); got != 0 {
		return fmt.Errorf("defectos guardados = %d, se esperaba ninguno", got)
	}
	return nil
}

func (s *defectScenario) mutateSaved() error {
	s.saved.Description = "Otra descripción"
	s.saved.Severity = SeverityLow
	s.saved.Status = StatusResolved
	s.saved.ItemID = 999
	s.saved.DetectionSprintID = 999
	s.saved.ResolutionSprintID = 999
	return nil
}

func (s *defectScenario) savedKeepsInput() error {
	defect, err := s.savedDefect()
	if err != nil {
		return err
	}
	want := Defect{
		ID:                 defect.ID,
		Description:        strings.TrimSpace(s.description),
		Severity:           Severity(strings.ToUpper(strings.TrimSpace(s.severity))),
		Status:             StatusOpen,
		ItemID:             s.itemID,
		DetectionSprintID:  s.sprintID,
		ResolutionSprintID: 0,
	}
	if defect.ID <= 0 {
		return fmt.Errorf("el defecto guardado no tiene un ID válido: %+v", defect)
	}
	if defect != want {
		return fmt.Errorf("defecto guardado = %+v, se esperaba %+v", defect, want)
	}
	return nil
}

func (s *defectScenario) allRegistersOK() error {
	for i, err := range s.errs {
		if err != nil {
			return fmt.Errorf("registro %d falló: %v", i, err)
		}
	}
	return nil
}

func (s *defectScenario) uniqueIDs() error {
	saved := s.memory.snapshot()
	if len(saved) != s.count {
		return fmt.Errorf("defectos guardados = %d, se esperaban %d", len(saved), s.count)
	}
	seen := map[int]bool{}
	for _, defect := range saved {
		if defect.ID <= 0 || seen[defect.ID] {
			return fmt.Errorf("ID inválido o repetido: %d en %+v", defect.ID, saved)
		}
		seen[defect.ID] = true
	}
	gotIDs := idsFromDefects(saved)
	wantIDs := append([]int(nil), s.ids...)
	slices.Sort(gotIDs)
	slices.Sort(wantIDs)
	if !slices.Equal(wantIDs, gotIDs) {
		return fmt.Errorf("IDs devueltos = %v, guardados = %v", s.ids, gotIDs)
	}
	return nil
}

func idsFromDefects(defects []Defect) []int {
	ids := make([]int, len(defects))
	for i, defect := range defects {
		ids[i] = defect.ID
	}
	return ids
}

func (s *defectScenario) sprintHasStatus(sprintID int, status string) error {
	sp, err := s.sprintRepo.GetByID(sprintID)
	if err != nil {
		return err
	}
	if string(sp.Status) != status {
		return fmt.Errorf("estado del sprint %d = %q, se esperaba %q", sprintID, sp.Status, status)
	}
	return nil
}

func (s *defectScenario) registerOKOrClosed() error {
	if s.closeErr != nil {
		return fmt.Errorf("el cierre falló: %v", s.closeErr)
	}
	if s.err != nil && !errors.Is(s.err, sprint.ErrSprintClosed) {
		return fmt.Errorf("el registro falló con un error inesperado: %v", s.err)
	}
	if s.err == nil && s.memory.count() != 1 {
		return fmt.Errorf("el registro terminó bien pero hay %d defectos", s.memory.count())
	}
	if s.err != nil && s.memory.count() != 0 {
		return fmt.Errorf("el registro falló pero hay %d defectos", s.memory.count())
	}
	return nil
}

func (s *defectScenario) noDefectAfterClose() error {
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
		return fmt.Errorf("no se registró el cierre: %v", events)
	}
	for _, event := range events[closeIndex+1:] {
		if event == "defect" {
			return fmt.Errorf("se guardó un defecto después del cierre: %v", events)
		}
	}
	return nil
}
