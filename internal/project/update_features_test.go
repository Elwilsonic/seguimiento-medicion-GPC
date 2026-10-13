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

// Ejecuta los escenarios BDD de features/HU-02-modificar-proyecto.feature con godog.
func TestFeatures_HU02ModificarProyecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeUpdateScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-02-modificar-proyecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-02")
	}
}

// updateScenario guarda el estado de un escenario de HU-02.
type updateScenario struct {
	manager *project.Manager
	repo    *project.MemoryRepository
	updated project.Project
	err     error
	opErrs  []error
}

func initializeUpdateScenario(ctx *godog.ScenarioContext) {
	s := &updateScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.repo = project.NewMemoryRepository()
		s.manager = project.NewManager(s.repo)
		s.updated = project.Project{}
		s.err = nil
		s.opErrs = nil
		return ctx, nil
	})

	ctx.Step(`^que no hay proyectos creados$`, s.noProjects)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" y descripción "([^"]*)"$`, s.existingProjectWithDescription)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" en el estado "([^"]*)"$`, s.existingProjectWithStatus)
	ctx.Step(`^que existen (\d+) proyectos numerados$`, s.existingNumberedProjects)
	ctx.Step(`^modifico el proyecto con el ID (-?\d+) con nombre "([^"]*)", descripción "([^"]*)" y estado "([^"]*)"$`, s.updateProject)
	ctx.Step(`^la modificación es exitosa$`, s.updatedOK)
	ctx.Step(`^el proyecto modificado tiene el nombre "([^"]*)"$`, s.updatedWithName)
	ctx.Step(`^el proyecto modificado tiene la descripción "([^"]*)"$`, s.updatedWithDescription)
	ctx.Step(`^el proyecto modificado está en el estado "([^"]*)"$`, s.updatedWithStatus)
	ctx.Step(`^la modificación falla porque el nombre es obligatorio$`, s.failsWithEmptyName)
	ctx.Step(`^la modificación falla porque el estado es inválido$`, s.failsWithInvalidStatus)
	ctx.Step(`^la modificación falla porque el nombre está repetido$`, s.failsWithDuplicateName)
	ctx.Step(`^la modificación falla porque el proyecto no existe$`, s.failsWithProjectNotFound)
	ctx.Step(`^el proyecto guardado con el ID (\d+) sigue teniendo el nombre "([^"]*)" y la descripción "([^"]*)"$`, s.projectHasNameAndDescription)
	ctx.Step(`^al reconsultar el proyecto con el ID (\d+) obtengo el nombre "([^"]*)" y la descripción "([^"]*)"$`, s.projectHasNameAndDescription)
	ctx.Step(`^el proyecto guardado con el ID (\d+) está en el estado "([^"]*)"$`, s.projectHasStatus)
	ctx.Step(`^cambio el nombre del proyecto modificado a "([^"]*)"$`, s.changeUpdatedName)
	ctx.Step(`^se modifican en simultáneo los proyectos del 1 al (\d+) mientras se crean (\d+) proyectos en simultáneo con nombres distintos$`, s.concurrentUpdatesAndCreates)
	ctx.Step(`^ninguna operación simultánea devuelve un error inesperado$`, s.noUnexpectedOperationErrors)
	ctx.Step(`^hay (\d+) proyectos guardados$`, s.savedCount)
	ctx.Step(`^los proyectos del 1 al (\d+) quedaron con su nombre actualizado y el estado "([^"]*)"$`, s.updatedProjectsHaveNameAndStatus)
}

func (s *updateScenario) noProjects() error {
	projects, err := s.repo.List()
	if err != nil {
		return err
	}
	if len(projects) != 0 {
		return fmt.Errorf("hay %d proyectos, se esperaba 0", len(projects))
	}
	return nil
}

func (s *updateScenario) existingProject(name string) error {
	_, err := s.manager.Create(name, "")
	return err
}

func (s *updateScenario) existingProjectWithDescription(name, description string) error {
	_, err := s.manager.Create(name, description)
	return err
}

func (s *updateScenario) existingProjectWithStatus(name, status string) error {
	// Con HU-02 disponible, el estado inicial distinto de PLANIFICADO se deja
	// con la propia operación de modificación.
	if _, err := s.manager.Create(name, ""); err != nil {
		return err
	}
	_, err := s.manager.Update(1, name, "", status)
	return err
}

func (s *updateScenario) existingNumberedProjects(count int) error {
	for i := range count {
		if _, err := s.manager.Create(fmt.Sprintf("Proyecto %d", i), ""); err != nil {
			return err
		}
	}
	return nil
}

func (s *updateScenario) updateProject(id int, name, description, status string) error {
	s.updated, s.err = s.manager.Update(id, name, description, status)
	return nil
}

func (s *updateScenario) updatedOK() error {
	if s.err != nil {
		return fmt.Errorf("la modificación falló: %v", s.err)
	}
	return nil
}

func (s *updateScenario) updatedWithName(name string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if s.updated.Name != name {
		return fmt.Errorf("nombre = %q, se esperaba %q", s.updated.Name, name)
	}
	return nil
}

func (s *updateScenario) updatedWithDescription(description string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if s.updated.Description != description {
		return fmt.Errorf("descripción = %q, se esperaba %q", s.updated.Description, description)
	}
	return nil
}

func (s *updateScenario) updatedWithStatus(status string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if string(s.updated.Status) != status {
		return fmt.Errorf("estado = %q, se esperaba %q", s.updated.Status, status)
	}
	return nil
}

func (s *updateScenario) failsWithEmptyName() error {
	if !errors.Is(s.err, project.ErrEmptyName) {
		return fmt.Errorf("error = %v, se esperaba ErrEmptyName", s.err)
	}
	return nil
}

func (s *updateScenario) failsWithInvalidStatus() error {
	if !errors.Is(s.err, project.ErrInvalidStatus) {
		return fmt.Errorf("error = %v, se esperaba ErrInvalidStatus", s.err)
	}
	return nil
}

func (s *updateScenario) failsWithDuplicateName() error {
	if !errors.Is(s.err, project.ErrDuplicateName) {
		return fmt.Errorf("error = %v, se esperaba ErrDuplicateName", s.err)
	}
	return nil
}

func (s *updateScenario) failsWithProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrProjectNotFound", s.err)
	}
	return nil
}

func (s *updateScenario) projectHasNameAndDescription(id int, name, description string) error {
	p, err := s.manager.Get(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	if p.Name != name {
		return fmt.Errorf("nombre = %q, se esperaba %q", p.Name, name)
	}
	if p.Description != description {
		return fmt.Errorf("descripción = %q, se esperaba %q", p.Description, description)
	}
	return nil
}

func (s *updateScenario) projectHasStatus(id int, status string) error {
	p, err := s.manager.Get(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	if string(p.Status) != status {
		return fmt.Errorf("estado = %q, se esperaba %q", p.Status, status)
	}
	return nil
}

func (s *updateScenario) changeUpdatedName(name string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	s.updated.Name = name
	return nil
}

func (s *updateScenario) concurrentUpdatesAndCreates(maxID, creates int) error {
	s.opErrs = make([]error, maxID+creates)
	var wg sync.WaitGroup
	for i := range maxID {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.opErrs[i] = s.manager.Update(i+1, fmt.Sprintf("Proyecto %d actualizado", i), "", "EN_CURSO")
		}()
	}
	for i := range creates {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.opErrs[maxID+i] = s.manager.Create(fmt.Sprintf("Proyecto %d", i+maxID), "")
		}()
	}
	wg.Wait()
	return nil
}

func (s *updateScenario) noUnexpectedOperationErrors() error {
	for i, err := range s.opErrs {
		if err != nil {
			return fmt.Errorf("la operación simultánea %d devolvió un error inesperado: %v", i+1, err)
		}
	}
	return nil
}

func (s *updateScenario) savedCount(want int) error {
	projects, err := s.repo.List()
	if err != nil {
		return err
	}
	if len(projects) != want {
		return fmt.Errorf("hay %d proyectos guardados, se esperaban %d", len(projects), want)
	}
	return nil
}

func (s *updateScenario) updatedProjectsHaveNameAndStatus(maxID int, status string) error {
	for i := range maxID {
		p, err := s.manager.Get(i + 1)
		if err != nil {
			return fmt.Errorf("la reconsulta del ID %d falló: %v", i+1, err)
		}
		if want := fmt.Sprintf("Proyecto %d actualizado", i); p.Name != want {
			return fmt.Errorf("nombre del ID %d = %q, se esperaba %q", i+1, p.Name, want)
		}
		if string(p.Status) != status {
			return fmt.Errorf("estado del ID %d = %q, se esperaba %q", i+1, p.Status, status)
		}
	}
	return nil
}
