package member

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-15-registrar-integrante.feature
// con godog.
func TestFeatures_HU15RegistrarIntegrante(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeRegisterScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-15-registrar-integrante.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-15")
	}
}

// registerScenario guarda el estado de un escenario de HU-15.
type registerScenario struct {
	projects   *project.Manager
	repo       *MemoryRepository
	members    *Manager
	projectIDs map[string]int

	result Member
	err    error
}

func initializeRegisterScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &registerScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		projects := project.NewManager(project.NewMemoryRepository())
		repo := NewMemoryRepository()
		*s = registerScenario{projects: projects, repo: repo, members: NewManager(repo, projects), projectIDs: map[string]int{}}
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el integrante "([^"]*)" está registrado con el rol "([^"]*)" en el proyecto "([^"]*)"$`, s.registeredMember)
	ctx.Step(`^registro al integrante "([^"]*)" con el rol "([^"]*)" en el proyecto "([^"]*)"$`, s.registerByName)
	ctx.Step(`^registro al integrante "([^"]*)" con el rol "([^"]*)" en el proyecto con el ID (-?\d+)$`, s.registerByID)
	ctx.Step(`^modifico el integrante devuelto$`, s.modifyResult)
	ctx.Step(`^el integrante queda registrado con el ID (\d+), el nombre "([^"]*)" y el rol "([^"]*)"$`, s.registeredAs)
	ctx.Step(`^el registro falla con el error "([^"]*)"$`, s.failsWith)
	ctx.Step(`^el registro falla porque el proyecto no existe$`, s.failsNotFound)
	ctx.Step(`^el proyecto "([^"]*)" no tiene integrantes$`, s.noMembers)
	ctx.Step(`^el proyecto "([^"]*)" tiene (\d+) integrantes?$`, s.memberCount)
	ctx.Step(`^el proyecto "([^"]*)" tiene al integrante "([^"]*)" con el rol "([^"]*)"$`, s.hasMember)
}

func (s *registerScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *registerScenario) registeredMember(name, role, projectName string) error {
	_, err := s.members.Register(s.projectIDs[projectName], name, role)
	return err
}

func (s *registerScenario) registerByName(name, role, projectName string) error {
	s.result, s.err = s.members.Register(s.projectIDs[projectName], name, role)
	return nil
}

func (s *registerScenario) registerByID(name, role, idText string) error {
	id, err := strconv.Atoi(idText)
	if err != nil {
		return err
	}
	s.result, s.err = s.members.Register(id, name, role)
	return nil
}

func (s *registerScenario) modifyResult() error {
	if s.err != nil {
		return fmt.Errorf("el registro falló: %v", s.err)
	}
	s.result.Name = "Otro"
	s.result.Role = RoleAgileEnabler
	return nil
}

func (s *registerScenario) registeredAs(id int, name, role string) error {
	if s.err != nil {
		return fmt.Errorf("el registro falló: %v", s.err)
	}
	if s.result.ID != id || s.result.Name != name || s.result.Role != Role(role) {
		return fmt.Errorf("integrante = %+v, se esperaba ID %d, nombre %q y rol %q", s.result, id, name, role)
	}
	return nil
}

func (s *registerScenario) failsWith(message string) error {
	if s.err == nil || s.err.Error() != message {
		return fmt.Errorf("err = %v, se esperaba %q", s.err, message)
	}
	return nil
}

func (s *registerScenario) failsNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	return nil
}

// stored lee los integrantes guardados del proyecto, directo del repositorio.
func (s *registerScenario) stored(projectName string) ([]Member, error) {
	return s.repo.ListByProject(s.projectIDs[projectName])
}

func (s *registerScenario) noMembers(projectName string) error {
	return s.memberCount(projectName, 0)
}

func (s *registerScenario) memberCount(projectName string, n int) error {
	list, err := s.stored(projectName)
	if err != nil {
		return err
	}
	if len(list) != n {
		return fmt.Errorf("integrantes = %+v, se esperaban %d", list, n)
	}
	return nil
}

func (s *registerScenario) hasMember(projectName, name, role string) error {
	list, err := s.stored(projectName)
	if err != nil {
		return err
	}
	if len(list) != 1 || list[0].Name != name || list[0].Role != Role(role) {
		return fmt.Errorf("integrantes = %+v, se esperaba solo %q con el rol %q", list, name, role)
	}
	return nil
}
