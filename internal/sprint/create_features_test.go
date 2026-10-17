package sprint

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-10-crear-sprint.feature con godog.
func TestFeatures_HU10CrearSprint(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeCreateSprintScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-10-crear-sprint.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-10")
	}
}

// createSprintScenario guarda el estado de un escenario de HU-10.
type createSprintScenario struct {
	projectRepo *project.MemoryRepository
	projects    *project.Manager
	repo        *MemoryRepository
	manager     *Manager

	projectIDs     map[string]int
	created        Sprint
	createdSprints []Sprint
	createErrs     []error
	err            error
	errs           []error
}

func initializeCreateSprintScenario(ctx *godog.ScenarioContext) {
	s := &createSprintScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projectRepo = project.NewMemoryRepository()
		s.projects = project.NewManager(s.projectRepo)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects)
		s.projectIDs = map[string]int{}
		s.created = Sprint{}
		s.createdSprints = nil
		s.createErrs = nil
		s.err = nil
		s.errs = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un proyecto finalizado con nombre "([^"]*)"$`, s.existingFinishedProject)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que existe un sprint cerrado "([^"]*)" en el proyecto "([^"]*)"$`, s.existingClosedSprint)
	ctx.Step(`^que no hay proyectos creados$`, s.noProjectsCreated)
	ctx.Step(`^creo el sprint "([^"]*)" con inicio "([^"]*)", fin "([^"]*)" y sprint goal "([^"]*)" en el proyecto "([^"]*)"$`, s.createSprintInProject)
	ctx.Step(`^creo el sprint "([^"]*)" con inicio "([^"]*)", fin "([^"]*)" y sprint goal "([^"]*)" en el proyecto con ID (-?\d+)$`, s.createSprintInProjectByID)
	ctx.Step(`^el sprint se crea con el ID (\d+)$`, s.createdWithID)
	ctx.Step(`^los sprints se crean con los IDs (\d+) y (\d+)$`, s.createdSprintsWithIDs)
	ctx.Step(`^el sprint queda en el estado "([^"]*)"$`, s.createdWithStatus)
	ctx.Step(`^el sprint tiene el nombre "([^"]*)"$`, s.createdWithName)
	ctx.Step(`^el sprint tiene el sprint goal "([^"]*)"$`, s.createdWithGoal)
	ctx.Step(`^el sprint tiene inicio "([^"]*)" y fin "([^"]*)"$`, s.createdWithDates)
	ctx.Step(`^modifico el nombre, el sprint goal y las fechas del sprint devuelto$`, s.mutateCreatedSprint)
	ctx.Step(`^el sprint guardado conserva el nombre "([^"]*)"$`, s.savedWithName)
	ctx.Step(`^el sprint guardado conserva el sprint goal "([^"]*)"$`, s.savedWithGoal)
	ctx.Step(`^el sprint guardado conserva el estado "([^"]*)"$`, s.savedWithStatus)
	ctx.Step(`^se crean (\d+) sprints en simultáneo en el proyecto "([^"]*)"$`, s.concurrentCreates)
	ctx.Step(`^exactamente una creación tiene éxito$`, s.exactlyOneSuccess)
	ctx.Step(`^las demás fallan porque ya existe un sprint abierto en el proyecto$`, s.othersFailWithOpenSprint)
	ctx.Step(`^queda un solo sprint ABIERTO guardado$`, s.singleOpenSprintSaved)
	ctx.Step(`^la creación falla porque el nombre del sprint es obligatorio$`, s.failsWithEmptyName)
	ctx.Step(`^la creación falla porque el sprint goal es obligatorio$`, s.failsWithEmptyGoal)
	ctx.Step(`^la creación falla porque la fecha de inicio del sprint es obligatoria$`, s.failsWithMissingStartDate)
	ctx.Step(`^la creación falla porque la fecha de fin del sprint es obligatoria$`, s.failsWithMissingEndDate)
	ctx.Step(`^la creación falla porque la fecha de fin no es posterior a la de inicio$`, s.failsWithEndNotAfterStart)
	ctx.Step(`^la creación falla porque ya existe un sprint abierto en el proyecto$`, s.failsWithOpenSprintExists)
	ctx.Step(`^la creación falla porque el proyecto no existe$`, s.failsWithProjectNotFound)
}

// parseDate convierte el texto de fecha de un paso BDD en time.Time. El texto
// "ausente" es la fecha zero, que Create trata como fecha ausente (RN-10.3).
func parseDate(raw string) (time.Time, error) {
	if raw == "ausente" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", raw)
}

func (s *createSprintScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *createSprintScenario) existingFinishedProject(name string) error {
	p, err := s.projectRepo.Save(project.Project{Name: name, Status: project.StatusFinished})
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

// saveSprint guarda un sprint directamente en el repositorio con el estado
// recibido (los pasos son internos al paquete, ver SDD).
func (s *createSprintScenario) saveSprint(name, projectName string, status Status) error {
	projectID, ok := s.projectIDs[projectName]
	if !ok {
		return fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	_, err := s.repo.Save(Sprint{
		ProjectID: projectID,
		Name:      name,
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    status,
	})
	return err
}

func (s *createSprintScenario) existingOpenSprint(name, projectName string) error {
	return s.saveSprint(name, projectName, StatusOpen)
}

func (s *createSprintScenario) existingClosedSprint(name, projectName string) error {
	return s.saveSprint(name, projectName, StatusClosed)
}

func (s *createSprintScenario) noProjectsCreated() error {
	return nil
}

func (s *createSprintScenario) createSprintInProject(name, startRaw, endRaw, goal, projectName string) error {
	projectID, ok := s.projectIDs[projectName]
	if !ok {
		return fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	return s.create(projectID, name, startRaw, endRaw, goal)
}

func (s *createSprintScenario) createSprintInProjectByID(name, startRaw, endRaw, goal string, projectID int) error {
	return s.create(projectID, name, startRaw, endRaw, goal)
}

func (s *createSprintScenario) create(projectID int, name, startRaw, endRaw, goal string) error {
	start, err := parseDate(startRaw)
	if err != nil {
		return err
	}
	end, err := parseDate(endRaw)
	if err != nil {
		return err
	}
	s.created, s.err = s.manager.Create(projectID, name, start, end, goal)
	s.createdSprints = append(s.createdSprints, s.created)
	s.createErrs = append(s.createErrs, s.err)
	return nil
}

func (s *createSprintScenario) createdOK() error {
	if s.err != nil {
		return fmt.Errorf("la creación falló: %v", s.err)
	}
	return nil
}

func (s *createSprintScenario) createdWithID(id int) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.ID != id {
		return fmt.Errorf("ID = %d, se esperaba %d", s.created.ID, id)
	}
	return nil
}

func (s *createSprintScenario) createdSprintsWithIDs(first, second int) error {
	if len(s.createdSprints) != 2 {
		return fmt.Errorf("se crearon %d sprints, se esperaban 2", len(s.createdSprints))
	}
	for i, err := range s.createErrs {
		if err != nil {
			return fmt.Errorf("la creación %d falló: %v", i+1, err)
		}
	}
	if s.createdSprints[0].ID != first || s.createdSprints[1].ID != second {
		return fmt.Errorf("IDs = %d y %d, se esperaban %d y %d",
			s.createdSprints[0].ID, s.createdSprints[1].ID, first, second)
	}
	return nil
}

func (s *createSprintScenario) createdWithStatus(status string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if string(s.created.Status) != status {
		return fmt.Errorf("estado = %q, se esperaba %q", s.created.Status, status)
	}
	return nil
}

func (s *createSprintScenario) createdWithName(name string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.Name != name {
		return fmt.Errorf("nombre = %q, se esperaba %q", s.created.Name, name)
	}
	return nil
}

func (s *createSprintScenario) createdWithGoal(goal string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.Goal != goal {
		return fmt.Errorf("sprint goal = %q, se esperaba %q", s.created.Goal, goal)
	}
	return nil
}

func (s *createSprintScenario) createdWithDates(startRaw, endRaw string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	wantStart, err := parseDate(startRaw)
	if err != nil {
		return err
	}
	wantEnd, err := parseDate(endRaw)
	if err != nil {
		return err
	}
	if !s.created.StartDate.Equal(wantStart) {
		return fmt.Errorf("inicio = %s, se esperaba %s", s.created.StartDate, wantStart)
	}
	if !s.created.EndDate.Equal(wantEnd) {
		return fmt.Errorf("fin = %s, se esperaba %s", s.created.EndDate, wantEnd)
	}
	return nil
}

func (s *createSprintScenario) mutateCreatedSprint() error {
	if err := s.createdOK(); err != nil {
		return err
	}
	s.created.Name = "Otro nombre"
	s.created.Goal = "Otro objetivo"
	s.created.Status = StatusClosed
	s.created.StartDate = validStart.AddDate(0, 0, -7)
	s.created.EndDate = validEnd.AddDate(0, 0, -7)
	return nil
}

// savedSprint devuelve el sprint guardado con el ID del último creado.
func (s *createSprintScenario) savedSprint() (Sprint, error) {
	for _, sprint := range s.repo.sprints {
		if sprint.ID == s.created.ID {
			return sprint, nil
		}
	}
	return Sprint{}, fmt.Errorf("no hay ningún sprint guardado con el ID %d", s.created.ID)
}

func (s *createSprintScenario) savedWithName(name string) error {
	saved, err := s.savedSprint()
	if err != nil {
		return err
	}
	if saved.Name != name {
		return fmt.Errorf("nombre guardado = %q, se esperaba %q", saved.Name, name)
	}
	return nil
}

func (s *createSprintScenario) savedWithGoal(goal string) error {
	saved, err := s.savedSprint()
	if err != nil {
		return err
	}
	if saved.Goal != goal {
		return fmt.Errorf("sprint goal guardado = %q, se esperaba %q", saved.Goal, goal)
	}
	return nil
}

func (s *createSprintScenario) savedWithStatus(status string) error {
	saved, err := s.savedSprint()
	if err != nil {
		return err
	}
	if string(saved.Status) != status {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", saved.Status, status)
	}
	return nil
}

func (s *createSprintScenario) concurrentCreates(attempts int, projectName string) error {
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
			_, s.errs[i] = s.manager.Create(projectID, fmt.Sprintf("Sprint %d", i), validStart, validEnd, validGoal)
		}()
	}
	wg.Wait()
	return nil
}

func (s *createSprintScenario) exactlyOneSuccess() error {
	successes := 0
	for _, err := range s.errs {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		return fmt.Errorf("tuvieron éxito %d creaciones, se esperaba exactamente 1", successes)
	}
	return nil
}

func (s *createSprintScenario) othersFailWithOpenSprint() error {
	for i, err := range s.errs {
		if err != nil && !errors.Is(err, ErrOpenSprintExists) {
			return fmt.Errorf("la creación %d falló con un error inesperado: %v", i+1, err)
		}
	}
	return nil
}

func (s *createSprintScenario) singleOpenSprintSaved() error {
	if len(s.repo.sprints) != 1 {
		return fmt.Errorf("hay %d sprints guardados, se esperaba 1", len(s.repo.sprints))
	}
	if s.repo.sprints[0].Status != StatusOpen {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", s.repo.sprints[0].Status, StatusOpen)
	}
	return nil
}

func (s *createSprintScenario) failsWithEmptyName() error {
	if !errors.Is(s.err, ErrEmptyName) {
		return fmt.Errorf("error = %v, se esperaba ErrEmptyName", s.err)
	}
	return nil
}

func (s *createSprintScenario) failsWithEmptyGoal() error {
	if !errors.Is(s.err, ErrEmptyGoal) {
		return fmt.Errorf("error = %v, se esperaba ErrEmptyGoal", s.err)
	}
	return nil
}

func (s *createSprintScenario) failsWithMissingStartDate() error {
	if !errors.Is(s.err, ErrMissingStartDate) {
		return fmt.Errorf("error = %v, se esperaba ErrMissingStartDate", s.err)
	}
	return nil
}

func (s *createSprintScenario) failsWithMissingEndDate() error {
	if !errors.Is(s.err, ErrMissingEndDate) {
		return fmt.Errorf("error = %v, se esperaba ErrMissingEndDate", s.err)
	}
	return nil
}

func (s *createSprintScenario) failsWithEndNotAfterStart() error {
	if !errors.Is(s.err, ErrEndNotAfterStart) {
		return fmt.Errorf("error = %v, se esperaba ErrEndNotAfterStart", s.err)
	}
	return nil
}

func (s *createSprintScenario) failsWithOpenSprintExists() error {
	if !errors.Is(s.err, ErrOpenSprintExists) {
		return fmt.Errorf("error = %v, se esperaba ErrOpenSprintExists", s.err)
	}
	return nil
}

func (s *createSprintScenario) failsWithProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrProjectNotFound", s.err)
	}
	return nil
}
