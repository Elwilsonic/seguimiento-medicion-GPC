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

// Ejecuta los escenarios BDD de features/HU-03-consultar-proyecto.feature con godog.
func TestFeatures_HU03ConsultarProyecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeGetScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-03-consultar-proyecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-03")
	}
}

// getScenario guarda el estado de un escenario de HU-03.
type getScenario struct {
	manager    *project.Manager
	repo       *project.MemoryRepository
	got        project.Project
	err        error
	errs       []error
	createErrs []error
}

func initializeGetScenario(ctx *godog.ScenarioContext) {
	s := &getScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.repo = project.NewMemoryRepository()
		s.manager = project.NewManager(s.repo)
		s.got = project.Project{}
		s.err = nil
		s.errs = nil
		s.createErrs = nil
		return ctx, nil
	})

	ctx.Step(`^que no hay proyectos creados$`, s.noProjects)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" y descripción "([^"]*)"$`, s.existingProjectWithDescription)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" en el estado "([^"]*)"$`, s.existingProjectWithStatus)
	ctx.Step(`^consulto el proyecto con el ID (-?\d+)$`, s.getProject)
	ctx.Step(`^obtengo el proyecto con el nombre "([^"]*)"$`, s.gotWithName)
	ctx.Step(`^el proyecto tiene la descripción "([^"]*)"$`, s.gotWithDescription)
	ctx.Step(`^el proyecto está en el estado "([^"]*)"$`, s.gotWithStatus)
	ctx.Step(`^la consulta falla porque el proyecto no existe$`, s.failsWithProjectNotFound)
	ctx.Step(`^cambio el nombre del proyecto obtenido a "([^"]*)"$`, s.changeGotName)
	ctx.Step(`^el proyecto guardado con el ID (\d+) tiene el nombre "([^"]*)"$`, s.savedWithIDAndName)
	ctx.Step(`^al reconsultar el proyecto con el ID (\d+) obtengo el nombre "([^"]*)"$`, s.regetHasName)
	ctx.Step(`^se crean (\d+) proyectos en simultáneo con nombres distintos mientras se consultan en simultáneo los IDs del 1 al (\d+)$`, s.concurrentCreatesAndGets)
	ctx.Step(`^ninguna consulta simultánea devuelve un error inesperado$`, s.noUnexpectedQueryErrors)
	ctx.Step(`^hay (\d+) proyectos guardados$`, s.savedCount)
	ctx.Step(`^al reconsultar los IDs del 1 al (\d+) se obtienen los (\d+) proyectos$`, s.regetAllIDs)
}

func (s *getScenario) noProjects() error {
	projects, err := s.repo.List()
	if err != nil {
		return err
	}
	if len(projects) != 0 {
		return fmt.Errorf("hay %d proyectos, se esperaba 0", len(projects))
	}
	return nil
}

func (s *getScenario) existingProject(name string) error {
	_, err := s.manager.Create(name, "")
	return err
}

func (s *getScenario) existingProjectWithDescription(name, description string) error {
	_, err := s.manager.Create(name, description)
	return err
}

func (s *getScenario) existingProjectWithStatus(name, status string) error {
	// HU-02 aún no cambia estados, así que se usa un repositorio falso de
	// prueba que guarda los proyectos con el estado indicado.
	s.manager = project.NewManager(&fixedStatusRepository{status: project.Status(status)})
	_, err := s.manager.Create(name, "")
	return err
}

func (s *getScenario) getProject(id int) error {
	s.got, s.err = s.manager.Get(id)
	return nil
}

func (s *getScenario) gotOK() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	return nil
}

func (s *getScenario) gotWithName(name string) error {
	if err := s.gotOK(); err != nil {
		return err
	}
	if s.got.Name != name {
		return fmt.Errorf("nombre = %q, se esperaba %q", s.got.Name, name)
	}
	return nil
}

func (s *getScenario) gotWithDescription(description string) error {
	if err := s.gotOK(); err != nil {
		return err
	}
	if s.got.Description != description {
		return fmt.Errorf("descripción = %q, se esperaba %q", s.got.Description, description)
	}
	return nil
}

func (s *getScenario) gotWithStatus(status string) error {
	if err := s.gotOK(); err != nil {
		return err
	}
	if string(s.got.Status) != status {
		return fmt.Errorf("estado = %q, se esperaba %q", s.got.Status, status)
	}
	return nil
}

func (s *getScenario) failsWithProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrProjectNotFound", s.err)
	}
	return nil
}

func (s *getScenario) changeGotName(name string) error {
	if err := s.gotOK(); err != nil {
		return err
	}
	s.got.Name = name
	return nil
}

func (s *getScenario) savedWithIDAndName(id int, name string) error {
	projects, err := s.repo.List()
	if err != nil {
		return err
	}
	for _, p := range projects {
		if p.ID == id {
			if p.Name != name {
				return fmt.Errorf("nombre guardado = %q, se esperaba %q", p.Name, name)
			}
			return nil
		}
	}
	return fmt.Errorf("no hay ningún proyecto guardado con el ID %d", id)
}

func (s *getScenario) regetHasName(id int, name string) error {
	p, err := s.manager.Get(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	if p.Name != name {
		return fmt.Errorf("nombre reconsultado = %q, se esperaba %q", p.Name, name)
	}
	return nil
}

func (s *getScenario) concurrentCreatesAndGets(creates, maxID int) error {
	s.errs = make([]error, maxID)
	var wg sync.WaitGroup
	s.createErrs = make([]error, creates)
	for i := range creates {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.createErrs[i] = s.manager.Create(fmt.Sprintf("Proyecto %d", i), "")
		}()
	}
	for id := 1; id <= maxID; id++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, s.errs[id-1] = s.manager.Get(id)
		}()
	}
	wg.Wait()
	return nil
}

func (s *getScenario) noUnexpectedQueryErrors() error {
	for i, err := range s.errs {
		if err != nil && !errors.Is(err, project.ErrProjectNotFound) {
			return fmt.Errorf("la consulta %d devolvió un error inesperado: %v", i+1, err)
		}
	}
	for i, err := range s.createErrs {
		if err != nil {
			return fmt.Errorf("la creación %d devolvió un error inesperado: %v", i+1, err)
		}
	}
	return nil
}

func (s *getScenario) savedCount(want int) error {
	projects, err := s.repo.List()
	if err != nil {
		return err
	}
	if len(projects) != want {
		return fmt.Errorf("hay %d proyectos guardados, se esperaban %d", len(projects), want)
	}
	return nil
}

func (s *getScenario) regetAllIDs(maxID, want int) error {
	if maxID != want {
		return fmt.Errorf("se esperaban %d proyectos para reconsultar %d IDs", want, maxID)
	}
	// Los IDs se asignan en cualquier orden con creaciones simultáneas, así
	// que se verifica que cada ID devuelva un proyecto con ese ID y que el
	// conjunto de nombres sea exactamente "Proyecto 0" a "Proyecto N-1".
	names := make(map[string]bool, maxID)
	for id := 1; id <= maxID; id++ {
		p, err := s.manager.Get(id)
		if err != nil {
			return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
		}
		if p.ID != id {
			return fmt.Errorf("Get(%d) devolvió el proyecto con ID %d", id, p.ID)
		}
		names[p.Name] = true
	}
	for i := range maxID {
		if !names[fmt.Sprintf("Proyecto %d", i)] {
			return fmt.Errorf("falta el proyecto con nombre %q", fmt.Sprintf("Proyecto %d", i))
		}
	}
	return nil
}
