package project_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-18-fecha-inicio-proyecto.feature con godog.
func TestFeatures_HU18FechaInicioProyecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeStartDateScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-18-fecha-inicio-proyecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-18")
	}
}

// startDateScenario guarda el estado de un escenario de HU-18.
type startDateScenario struct {
	manager  *project.Manager
	repo     *updateFailingRepository
	returned project.Project
	err      error
}

func initializeStartDateScenario(ctx *godog.ScenarioContext) {
	s := &startDateScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.repo = &updateFailingRepository{MemoryRepository: project.NewMemoryRepository()}
		s.manager = project.NewManager(s.repo)
		s.returned = project.Project{}
		s.err = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" y descripción "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" en el estado "([^"]*)"$`, s.existingProjectWithStatus)
	ctx.Step(`^el proyecto con el ID (\d+) tiene registrada la fecha de inicio "([^"]*)"$`, s.projectWithStartDate)
	ctx.Step(`^el repositorio falla al actualizar$`, s.repositoryFailsOnUpdate)
	ctx.Step(`^registro la fecha de inicio "([^"]*)" en el proyecto con el ID (-?\d+)$`, s.setStartDate)
	ctx.Step(`^registro una fecha de inicio ausente en el proyecto con el ID (-?\d+)$`, s.setMissingStartDate)
	ctx.Step(`^modifico el proyecto con el ID (\d+) con nombre "([^"]*)", descripción "([^"]*)" y estado "([^"]*)"$`, s.updateProject)
	ctx.Step(`^cambio la fecha de inicio del proyecto devuelto a "([^"]*)"$`, s.changeReturnedStartDate)
	ctx.Step(`^el registro es exitoso$`, s.registeredOK)
	ctx.Step(`^el proyecto devuelto tiene la fecha de inicio "([^"]*)"$`, s.returnedHasStartDate)
	ctx.Step(`^el proyecto guardado con el ID (\d+) tiene la fecha de inicio "([^"]*)"$`, s.storedHasStartDate)
	ctx.Step(`^el proyecto guardado con el ID (\d+) no tiene fecha de inicio$`, s.storedWithoutStartDate)
	ctx.Step(`^el proyecto guardado con el ID (\d+) sigue teniendo el nombre "([^"]*)", la descripción "([^"]*)" y el estado "([^"]*)"$`, s.storedHasData)
	ctx.Step(`^el registro falla porque la fecha de inicio es obligatoria$`, s.failsWithMissingStartDate)
	ctx.Step(`^el registro falla porque el proyecto no existe$`, s.failsWithProjectNotFound)
	ctx.Step(`^el registro falla con el error del repositorio$`, s.failsWithRepositoryError)
}

func (s *startDateScenario) existingProject(name, description string) error {
	_, err := s.manager.Create(name, description)
	return err
}

func (s *startDateScenario) existingProjectWithStatus(name, status string) error {
	p, err := s.manager.Create(name, "")
	if err != nil {
		return err
	}
	_, err = s.manager.Update(p.ID, name, "", status)
	return err
}

// projectWithStartDate deja la fecha registrada con la propia operación de
// HU-18 y falla el paso si el registro no la guardó.
func (s *startDateScenario) projectWithStartDate(id int, date string) error {
	day, err := parseDay(date)
	if err != nil {
		return err
	}
	if _, err := s.manager.SetStartDate(id, day); err != nil {
		return fmt.Errorf("SetStartDate devolvió un error inesperado: %v", err)
	}
	return s.storedHasStartDate(id, date)
}

func (s *startDateScenario) repositoryFailsOnUpdate() error {
	s.repo.failUpdate = true
	return nil
}

func (s *startDateScenario) setStartDate(date string, id int) error {
	day, err := parseDay(date)
	if err != nil {
		return err
	}
	s.returned, s.err = s.manager.SetStartDate(id, day)
	return nil
}

func (s *startDateScenario) setMissingStartDate(id int) error {
	s.returned, s.err = s.manager.SetStartDate(id, time.Time{})
	return nil
}

func (s *startDateScenario) updateProject(id int, name, description, status string) error {
	_, err := s.manager.Update(id, name, description, status)
	return err
}

func (s *startDateScenario) changeReturnedStartDate(date string) error {
	if err := s.registeredOK(); err != nil {
		return err
	}
	day, err := parseDay(date)
	if err != nil {
		return err
	}
	s.returned.StartDate = day
	return nil
}

func (s *startDateScenario) registeredOK() error {
	if s.err != nil {
		return fmt.Errorf("el registro falló: %v", s.err)
	}
	return nil
}

func (s *startDateScenario) returnedHasStartDate(date string) error {
	if err := s.registeredOK(); err != nil {
		return err
	}
	return expectStartDate(s.returned, date)
}

func (s *startDateScenario) storedHasStartDate(id int, date string) error {
	p, err := s.manager.Get(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	return expectStartDate(p, date)
}

func (s *startDateScenario) storedWithoutStartDate(id int) error {
	p, err := s.manager.Get(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	if !p.StartDate.IsZero() {
		return fmt.Errorf("fecha de inicio = %v, se esperaba sin registrar", p.StartDate)
	}
	return nil
}

func (s *startDateScenario) storedHasData(id int, name, description, status string) error {
	p, err := s.manager.Get(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	if p.Name != name || p.Description != description || string(p.Status) != status {
		return fmt.Errorf("proyecto = %q, %q, %q; se esperaba %q, %q, %q",
			p.Name, p.Description, p.Status, name, description, status)
	}
	return nil
}

func (s *startDateScenario) failsWithMissingStartDate() error {
	if !errors.Is(s.err, project.ErrMissingStartDate) {
		return fmt.Errorf("error = %v, se esperaba ErrMissingStartDate", s.err)
	}
	return nil
}

func (s *startDateScenario) failsWithProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrProjectNotFound", s.err)
	}
	return nil
}

func (s *startDateScenario) failsWithRepositoryError() error {
	if !errors.Is(s.err, errRepository) {
		return fmt.Errorf("error = %v, se esperaba el error del repositorio", s.err)
	}
	return nil
}

// parseDay lee una fecha AAAA-MM-DD, como la recibe la API (00:00 UTC).
func parseDay(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}

// expectStartDate verifica que p tenga la fecha de inicio date (AAAA-MM-DD).
func expectStartDate(p project.Project, date string) error {
	want, err := parseDay(date)
	if err != nil {
		return err
	}
	if !p.StartDate.Equal(want) {
		return fmt.Errorf("fecha de inicio = %v, se esperaba %s", p.StartDate, date)
	}
	return nil
}
