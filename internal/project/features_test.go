package project_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-01-crear-proyecto.feature con godog.
func TestFeatures_HU01CrearProyecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeCreateScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-01-crear-proyecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-01")
	}
}

// createScenario guarda el estado de un escenario de HU-01.
type createScenario struct {
	manager *project.Manager
	repo    *project.MemoryRepository
	created project.Project
	err     error
	errs    []error
}

func initializeCreateScenario(ctx *godog.ScenarioContext) {
	s := &createScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.repo = project.NewMemoryRepository()
		s.manager = project.NewManager(s.repo)
		return ctx, nil
	})

	ctx.Step(`^que no hay proyectos creados$`, s.noProjects)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^creo un proyecto con nombre "([^"]*)" y descripción "([^"]*)"$`, s.createProject)
	ctx.Step(`^cambio el nombre del proyecto devuelto a "([^"]*)"$`, s.changeReturnedName)
	ctx.Step(`^se crean (\d+) proyectos en simultáneo con el nombre "([^"]*)"$`, s.createConcurrentSameName)
	ctx.Step(`^se crean (\d+) proyectos en simultáneo con nombres distintos$`, s.createConcurrentDistinctNames)
	ctx.Step(`^el proyecto se crea con el ID (\d+)$`, s.createdWithID)
	ctx.Step(`^el proyecto tiene el nombre "([^"]*)"$`, s.createdWithName)
	ctx.Step(`^el proyecto tiene la descripción "([^"]*)"$`, s.createdWithDescription)
	ctx.Step(`^el proyecto queda en el estado "([^"]*)"$`, s.createdWithStatus)
	ctx.Step(`^el proyecto guardado tiene el nombre "([^"]*)"$`, s.savedWithName)
	ctx.Step(`^la creación falla porque el nombre es obligatorio$`, s.failsWithEmptyName)
	ctx.Step(`^la creación falla porque el nombre está repetido$`, s.failsWithDuplicateName)
	ctx.Step(`^hay (\d+) proyectos guardados$`, s.savedCount)
	ctx.Step(`^(\d+) creaciones fallan porque el nombre está repetido$`, s.concurrentDuplicates)
	ctx.Step(`^los IDs de los proyectos guardados son todos distintos$`, s.distinctIDs)
}

func (s *createScenario) noProjects() error {
	projects, err := s.repo.List()
	if err != nil {
		return err
	}
	if len(projects) != 0 {
		return fmt.Errorf("hay %d proyectos, se esperaba 0", len(projects))
	}
	return nil
}

func (s *createScenario) existingProject(name string) error {
	_, err := s.manager.Create(name, "")
	return err
}

func (s *createScenario) createProject(name, description string) error {
	s.created, s.err = s.manager.Create(name, description)
	return nil
}

func (s *createScenario) changeReturnedName(name string) error {
	if s.err != nil {
		return fmt.Errorf("la creación falló: %v", s.err)
	}
	s.created.Name = name
	return nil
}

func (s *createScenario) createConcurrentSameName(attempts int, name string) error {
	s.errs = make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.errs[i] = s.manager.Create(name, "")
		}()
	}
	wg.Wait()
	return nil
}

func (s *createScenario) createConcurrentDistinctNames(attempts int) error {
	s.errs = make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.errs[i] = s.manager.Create(fmt.Sprintf("Proyecto %d", i), "")
		}()
	}
	wg.Wait()
	for _, err := range s.errs {
		if err != nil {
			return fmt.Errorf("una creación falló: %v", err)
		}
	}
	return nil
}

func (s *createScenario) createdOK() error {
	if s.err != nil {
		return fmt.Errorf("la creación falló: %v", s.err)
	}
	return nil
}

func (s *createScenario) createdWithID(id int) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.ID != id {
		return fmt.Errorf("ID = %d, se esperaba %d", s.created.ID, id)
	}
	return nil
}

func (s *createScenario) createdWithName(name string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.Name != name {
		return fmt.Errorf("nombre = %q, se esperaba %q", s.created.Name, name)
	}
	return nil
}

func (s *createScenario) createdWithDescription(description string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if s.created.Description != description {
		return fmt.Errorf("descripción = %q, se esperaba %q", s.created.Description, description)
	}
	return nil
}

func (s *createScenario) createdWithStatus(status string) error {
	if err := s.createdOK(); err != nil {
		return err
	}
	if string(s.created.Status) != status {
		return fmt.Errorf("estado = %q, se esperaba %q", s.created.Status, status)
	}
	return nil
}

func (s *createScenario) savedWithName(name string) error {
	projects, err := s.repo.List()
	if err != nil {
		return err
	}
	if len(projects) != 1 {
		return fmt.Errorf("hay %d proyectos guardados, se esperaba 1", len(projects))
	}
	if projects[0].Name != name {
		return fmt.Errorf("nombre guardado = %q, se esperaba %q", projects[0].Name, name)
	}
	return nil
}

func (s *createScenario) failsWithEmptyName() error {
	if !errors.Is(s.err, project.ErrEmptyName) {
		return fmt.Errorf("error = %v, se esperaba ErrEmptyName", s.err)
	}
	return nil
}

func (s *createScenario) failsWithDuplicateName() error {
	if !errors.Is(s.err, project.ErrDuplicateName) {
		return fmt.Errorf("error = %v, se esperaba ErrDuplicateName", s.err)
	}
	return nil
}

func (s *createScenario) savedCount(want int) error {
	projects, err := s.repo.List()
	if err != nil {
		return err
	}
	if len(projects) != want {
		return fmt.Errorf("hay %d proyectos guardados, se esperaban %d", len(projects), want)
	}
	return nil
}

func (s *createScenario) concurrentDuplicates(want int) error {
	duplicates := 0
	for _, err := range s.errs {
		if errors.Is(err, project.ErrDuplicateName) {
			duplicates++
		}
	}
	if duplicates != want {
		return fmt.Errorf("%d creaciones fallaron por nombre repetido, se esperaban %d", duplicates, want)
	}
	return nil
}

func (s *createScenario) distinctIDs() error {
	projects, err := s.repo.List()
	if err != nil {
		return err
	}
	ids := make(map[int]bool)
	for _, p := range projects {
		if ids[p.ID] {
			return fmt.Errorf("el ID %d está repetido", p.ID)
		}
		ids[p.ID] = true
	}
	return nil
}
