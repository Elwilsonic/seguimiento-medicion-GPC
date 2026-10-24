package effort

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

// Ejecuta los escenarios BDD de features/HU-26-registrar-esfuerzo.feature con
// godog. Los pasos viven en el paquete effort porque la historia es de E7 y
// orquesta los paquetes effort, sprint y backlog (ver SDD).
func TestFeatures_HU26RegistrarEsfuerzo(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeEffortScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-26-registrar-esfuerzo.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-26")
	}
}

// effortScenario guarda el estado de un escenario de HU-26.
type effortScenario struct {
	projects   *project.Manager
	projectIDs map[string]int
	itemRepo   *backlog.MemoryRepository
	items      *backlog.Manager
	sprintRepo sprint.Repository
	sprints    *sprint.Manager
	memory     *MemoryRepository
	repo       Repository
	manager    *Manager

	member, activity, dateRaw string
	date                      time.Time
	hours                     float64
	itemID, sprintID          int
	itemBefore                backlog.Item
	hasItemBefore             bool

	saved    Effort
	err      error
	closeErr error
	errs     []error
	ids      []int
	count    int
	log      *eventLog
}

func initializeEffortScenario(ctx *godog.ScenarioContext) {
	s := &effortScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.projectIDs = map[string]int{}
		s.itemRepo = backlog.NewMemoryRepository()
		s.items = backlog.NewManager(s.itemRepo, s.projects)
		s.sprintRepo = sprint.NewMemoryRepository()
		s.sprints = sprint.NewManager(s.sprintRepo, s.projects, s.items)
		s.memory = NewMemoryRepository()
		s.repo = s.memory
		s.manager = NewManager(s.repo, s.sprints, s.items)
		s.member, s.activity, s.dateRaw = "", "", ""
		s.date = time.Time{}
		s.hours = 0
		s.itemID, s.sprintID = 0, 0
		s.itemBefore = backlog.Item{}
		s.hasItemBefore = false
		s.saved = Effort{}
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
	ctx.Step(`^que el repositorio de esfuerzos falla al guardar$`, s.repoFailsOnSave)
	ctx.Step(`^registro un esfuerzo del integrante "([^"]*)" con fecha "([^"]*)", actividad "([^"]*)" y horas ([^ ]+) para el ítem con el ID (-?\d+) del sprint con el ID (-?\d+)$`, s.register)
	ctx.Step(`^registro en simultáneo (\d+) esfuerzos del integrante "([^"]*)" con fecha "([^"]*)", actividad "([^"]*)" y horas ([^ ]+) para el ítem con el ID (-?\d+) del sprint con el ID (-?\d+)$`, s.registerConcurrently)
	ctx.Step(`^registro un esfuerzo del integrante "([^"]*)" con fecha "([^"]*)", actividad "([^"]*)" y horas ([^ ]+) para el ítem con el ID (-?\d+) del sprint con el ID (-?\d+) y en simultáneo cierro el sprint con el ID (-?\d+)$`, s.registerAndClose)
	ctx.Step(`^el registro es exitoso$`, s.registerOK)
	ctx.Step(`^el esfuerzo guardado tiene el ID (\d+)$`, s.savedHasID)
	ctx.Step(`^el esfuerzo guardado conserva todos los datos del registro$`, s.savedKeepsInput)
	ctx.Step(`^el registro falla porque las horas no son válidas$`, s.failsInvalidHours)
	ctx.Step(`^el registro falla porque el ítem no está asignado al sprint$`, s.failsItemNotAssigned)
	ctx.Step(`^el registro falla porque no existe un sprint con ese id$`, s.failsSprintNotFound)
	ctx.Step(`^el registro falla porque no existe un ítem del backlog con ese id$`, s.failsItemNotFound)
	ctx.Step(`^el registro falla porque el sprint está cerrado$`, s.failsSprintClosed)
	ctx.Step(`^el registro falla porque el integrante es obligatorio$`, s.failsEmptyMember)
	ctx.Step(`^el registro falla porque la actividad es obligatoria$`, s.failsEmptyActivity)
	ctx.Step(`^el registro falla porque la fecha del esfuerzo es obligatoria$`, s.failsMissingDate)
	ctx.Step(`^el registro falla con el error del repositorio$`, s.failsRepository)
	ctx.Step(`^no se guardó ningún esfuerzo$`, s.noEffortSaved)
	ctx.Step(`^el ítem con el ID (-?\d+) sigue en estado "([^"]*)" y con los mismos datos$`, s.itemUnchanged)
	ctx.Step(`^cambio todos los datos del esfuerzo devuelto$`, s.mutateSaved)
	ctx.Step(`^todos los registros son exitosos$`, s.allRegistersOK)
	ctx.Step(`^los esfuerzos guardados tienen IDs únicos$`, s.uniqueIDs)
	ctx.Step(`^el sprint con el ID (-?\d+) queda en estado "([^"]*)"$`, s.sprintHasStatus)
	ctx.Step(`^cada registro simultáneo termina bien o porque el sprint está cerrado$`, s.registerOKOrClosed)
	ctx.Step(`^ningún esfuerzo se guarda después del cierre$`, s.noEffortAfterClose)
}

func (s *effortScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *effortScenario) projectID(name string) (int, error) {
	id, ok := s.projectIDs[name]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", name)
	}
	return id, nil
}

func (s *effortScenario) existingItem(title, status, projectName string) error {
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

func (s *effortScenario) existingSprint(kind, name, projectName string) error {
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

func (s *effortScenario) assignedItem(itemID, sprintID int) error {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return err
	}
	item.SprintID = sprintID
	_, err = s.itemRepo.Update(item)
	return err
}

func (s *effortScenario) repoFailsOnSave() error {
	s.repo = failingEffortRepository{err: errEffortRepository}
	s.manager = NewManager(s.repo, s.sprints, s.items)
	return nil
}

func (s *effortScenario) parseRegister(member, dateRaw, activity, hoursRaw string, itemID, sprintID int) (time.Time, float64, error) {
	hours, err := strconv.ParseFloat(hoursRaw, 64)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("horas de prueba inválidas %q: %w", hoursRaw, err)
	}
	var date time.Time
	if dateRaw != "" {
		date, err = time.Parse("2006-01-02", dateRaw)
		if err != nil {
			return time.Time{}, 0, fmt.Errorf("fecha de prueba inválida %q: %w", dateRaw, err)
		}
	}
	if item, err := s.itemRepo.GetByID(itemID); err == nil {
		s.itemBefore = item
		s.hasItemBefore = true
	}
	s.member, s.activity, s.dateRaw = member, activity, dateRaw
	s.date, s.hours = date, hours
	s.itemID, s.sprintID = itemID, sprintID
	return date, hours, nil
}

func (s *effortScenario) register(member, dateRaw, activity, hoursRaw string, itemID, sprintID int) error {
	date, hours, err := s.parseRegister(member, dateRaw, activity, hoursRaw, itemID, sprintID)
	if err != nil {
		return err
	}
	s.saved, s.err = s.manager.Register(sprintID, itemID, member, date, activity, hours)
	return nil
}

func (s *effortScenario) registerConcurrently(count int, member, dateRaw, activity, hoursRaw string, itemID, sprintID int) error {
	date, hours, err := s.parseRegister(member, dateRaw, activity, hoursRaw, itemID, sprintID)
	if err != nil {
		return err
	}
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
			got, err := s.manager.Register(sprintID, itemID, member, date, activity, hours)
			s.errs[i] = err
			s.ids[i] = got.ID
		}(i)
	}
	close(start)
	wg.Wait()
	return nil
}

func (s *effortScenario) registerAndClose(member, dateRaw, activity, hoursRaw string, itemID, sprintID, closeSprintID int) error {
	date, hours, err := s.parseRegister(member, dateRaw, activity, hoursRaw, itemID, sprintID)
	if err != nil {
		return err
	}
	log := &eventLog{}
	sprintRepo := &recordingSprintRepository{Repository: s.sprintRepo, log: log}
	effortRepo := &recordingEffortRepository{MemoryRepository: s.memory, log: log}
	sprintManager := sprint.NewManager(sprintRepo, s.projects, s.items)
	manager := NewManager(effortRepo, sprintManager, s.items)

	start := make(chan struct{})
	var wg sync.WaitGroup
	var registerErr, closeErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, registerErr = manager.Register(sprintID, itemID, member, date, activity, hours)
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

func (s *effortScenario) registerOK() error {
	if s.err != nil {
		return fmt.Errorf("el registro falló: %v", s.err)
	}
	return nil
}

func (s *effortScenario) savedHasID(id int) error {
	if len(s.memory.efforts) != 1 {
		return fmt.Errorf("esfuerzos guardados = %d, se esperaba 1", len(s.memory.efforts))
	}
	if s.memory.efforts[0].ID != id {
		return fmt.Errorf("ID guardado = %d, se esperaba %d", s.memory.efforts[0].ID, id)
	}
	return nil
}

func (s *effortScenario) savedKeepsInput() error {
	if len(s.memory.efforts) != 1 {
		return fmt.Errorf("esfuerzos guardados = %d, se esperaba 1", len(s.memory.efforts))
	}
	got := s.memory.efforts[0]
	want := Effort{
		Member:   strings.TrimSpace(s.member),
		Date:     s.date,
		Activity: strings.TrimSpace(s.activity),
		ItemID:   s.itemID,
		SprintID: s.sprintID,
		Hours:    s.hours,
	}
	if got.ID <= 0 {
		return fmt.Errorf("el esfuerzo guardado no tiene un ID válido: %+v", got)
	}
	got.ID = 0
	if !got.Date.Equal(want.Date) {
		return fmt.Errorf("fecha guardada = %v, se esperaba %v", got.Date, want.Date)
	}
	got.Date = want.Date
	if got != want {
		return fmt.Errorf("esfuerzo guardado = %+v, se esperaba %+v", got, want)
	}
	return nil
}

func (s *effortScenario) failsWith(target error) error {
	if s.err == nil {
		return fmt.Errorf("se esperaba %v, no hubo error", target)
	}
	if !errors.Is(s.err, target) {
		return fmt.Errorf("error = %v, se esperaba %v", s.err, target)
	}
	return nil
}

func (s *effortScenario) failsInvalidHours() error { return s.failsWith(ErrInvalidHours) }
func (s *effortScenario) failsItemNotAssigned() error {
	return s.failsWith(sprint.ErrItemNotAssignedToSprint)
}
func (s *effortScenario) failsSprintNotFound() error { return s.failsWith(sprint.ErrSprintNotFound) }
func (s *effortScenario) failsItemNotFound() error   { return s.failsWith(backlog.ErrItemNotFound) }
func (s *effortScenario) failsSprintClosed() error   { return s.failsWith(sprint.ErrSprintClosed) }
func (s *effortScenario) failsEmptyMember() error    { return s.failsWith(ErrEmptyMember) }
func (s *effortScenario) failsEmptyActivity() error  { return s.failsWith(ErrEmptyActivity) }
func (s *effortScenario) failsMissingDate() error    { return s.failsWith(ErrMissingDate) }
func (s *effortScenario) failsRepository() error     { return s.failsWith(errEffortRepository) }

func (s *effortScenario) noEffortSaved() error {
	if len(s.memory.efforts) != 0 {
		return fmt.Errorf("se guardaron %d esfuerzos, se esperaba ninguno", len(s.memory.efforts))
	}
	return nil
}

func (s *effortScenario) itemUnchanged(itemID int, status string) error {
	if !s.hasItemBefore {
		return fmt.Errorf("no se guardó el ítem %d antes del registro", itemID)
	}
	item, err := s.items.GetItem(itemID)
	if err != nil {
		return err
	}
	if string(item.Status) != status {
		return fmt.Errorf("estado del ítem %d = %q, se esperaba %q", itemID, item.Status, status)
	}
	if item.ID != s.itemBefore.ID || item.ProjectID != s.itemBefore.ProjectID ||
		item.Title != s.itemBefore.Title || item.Description != s.itemBefore.Description ||
		item.Priority != s.itemBefore.Priority || item.Status != s.itemBefore.Status ||
		item.StoryPoints != s.itemBefore.StoryPoints || item.EstimatedHours != s.itemBefore.EstimatedHours ||
		item.SprintID != s.itemBefore.SprintID ||
		!slices.Equal(item.AcceptanceCriteria, s.itemBefore.AcceptanceCriteria) {
		return fmt.Errorf("ítem modificado: antes %+v, después %+v", s.itemBefore, item)
	}
	return nil
}

func (s *effortScenario) mutateSaved() error {
	s.saved.Member = "Otra persona"
	s.saved.Activity = "Otra actividad"
	s.saved.Date = s.saved.Date.AddDate(0, 0, 1)
	s.saved.Hours = 99
	s.saved.ItemID = 999
	s.saved.SprintID = 999
	return nil
}

func (s *effortScenario) allRegistersOK() error {
	for i, err := range s.errs {
		if err != nil {
			return fmt.Errorf("registro %d falló: %v", i, err)
		}
	}
	return nil
}

func (s *effortScenario) uniqueIDs() error {
	if len(s.memory.efforts) != s.count {
		return fmt.Errorf("esfuerzos guardados = %d, se esperaban %d", len(s.memory.efforts), s.count)
	}
	seen := map[int]bool{}
	for _, e := range s.memory.efforts {
		if e.ID <= 0 || seen[e.ID] {
			return fmt.Errorf("ID inválido o repetido: %d en %+v", e.ID, s.memory.efforts)
		}
		seen[e.ID] = true
	}
	return nil
}

func (s *effortScenario) sprintHasStatus(sprintID int, status string) error {
	sp, err := s.sprintRepo.GetByID(sprintID)
	if err != nil {
		return err
	}
	if string(sp.Status) != status {
		return fmt.Errorf("estado del sprint %d = %q, se esperaba %q", sprintID, sp.Status, status)
	}
	return nil
}

func (s *effortScenario) registerOKOrClosed() error {
	if s.closeErr != nil {
		return fmt.Errorf("el cierre falló: %v", s.closeErr)
	}
	if s.err != nil && !errors.Is(s.err, sprint.ErrSprintClosed) {
		return fmt.Errorf("el registro falló con un error inesperado: %v", s.err)
	}
	if s.err == nil && len(s.memory.efforts) != 1 {
		return fmt.Errorf("el registro terminó bien pero hay %d esfuerzos", len(s.memory.efforts))
	}
	if s.err != nil && len(s.memory.efforts) != 0 {
		return fmt.Errorf("el registro falló pero hay %d esfuerzos", len(s.memory.efforts))
	}
	return nil
}

func (s *effortScenario) noEffortAfterClose() error {
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
		if event == "effort" {
			return fmt.Errorf("se guardó un esfuerzo después del cierre: %v", events)
		}
	}
	return nil
}
