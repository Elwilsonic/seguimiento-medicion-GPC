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

// Ejecuta los escenarios BDD de features/HU-16-modificar-integrante.feature
// con godog.
func TestFeatures_HU16ModificarIntegrante(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeUpdateScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-16-modificar-integrante.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-16")
	}
}

// updateScenario guarda el estado de un escenario de HU-16.
type updateScenario struct {
	projects   *project.Manager
	repo       *MemoryRepository
	members    *Manager
	projectIDs map[string]int
	memberIDs  map[string]int

	result Member
	err    error
}

func initializeUpdateScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &updateScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		projects := project.NewManager(project.NewMemoryRepository())
		repo := NewMemoryRepository()
		*s = updateScenario{projects: projects, repo: repo, members: NewManager(repo, projects), projectIDs: map[string]int{}, memberIDs: map[string]int{}}
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el integrante "([^"]*)" está registrado con el rol "([^"]*)" en el proyecto "([^"]*)"$`, s.registeredMember)
	ctx.Step(`^modifico al integrante "([^"]*)" con el nombre "([^"]*)" y el rol "([^"]*)"$`, s.updateByName)
	ctx.Step(`^modifico al integrante con el ID (-?\d+) con el nombre "([^"]*)" y el rol "([^"]*)"$`, s.updateByID)
	ctx.Step(`^cambio el integrante devuelto$`, s.changeResult)
	ctx.Step(`^el integrante queda con el ID (\d+), el nombre "([^"]*)" y el rol "([^"]*)" en el proyecto "([^"]*)"$`, s.updatedAs)
	ctx.Step(`^la modificación falla con el error "([^"]*)"$`, s.failsWith)
	ctx.Step(`^la modificación falla porque el integrante no existe$`, s.failsNotFound)
	ctx.Step(`^el proyecto "([^"]*)" tiene (\d+) integrantes?$`, s.memberCount)
	ctx.Step(`^el proyecto "([^"]*)" tiene al integrante "([^"]*)" con el rol "([^"]*)"$`, s.hasMember)
}

func (s *updateScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *updateScenario) registeredMember(name, role, projectName string) error {
	m, err := s.members.Register(s.projectIDs[projectName], name, role)
	if err != nil {
		return err
	}
	s.memberIDs[name] = m.ID
	return nil
}

func (s *updateScenario) updateByName(member, name, role string) error {
	s.result, s.err = s.members.Update(s.memberIDs[member], name, role)
	return nil
}

func (s *updateScenario) updateByID(idText, name, role string) error {
	id, err := strconv.Atoi(idText)
	if err != nil {
		return err
	}
	s.result, s.err = s.members.Update(id, name, role)
	return nil
}

func (s *updateScenario) changeResult() error {
	if s.err != nil {
		return fmt.Errorf("la modificación falló: %v", s.err)
	}
	s.result.Name = "Otro"
	s.result.Role = RoleProductArchitect
	return nil
}

func (s *updateScenario) updatedAs(id int, name, role, projectName string) error {
	if s.err != nil {
		return fmt.Errorf("la modificación falló: %v", s.err)
	}
	want := Member{ID: id, ProjectID: s.projectIDs[projectName], Name: name, Role: Role(role)}
	if s.result != want {
		return fmt.Errorf("integrante = %+v, se esperaba %+v", s.result, want)
	}
	return nil
}

func (s *updateScenario) failsWith(message string) error {
	if s.err == nil || s.err.Error() != message {
		return fmt.Errorf("err = %v, se esperaba %q", s.err, message)
	}
	return nil
}

func (s *updateScenario) failsNotFound() error {
	if !errors.Is(s.err, ErrMemberNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, ErrMemberNotFound)
	}
	return nil
}

func (s *updateScenario) memberCount(projectName string, n int) error {
	list, err := s.repo.ListByProject(s.projectIDs[projectName])
	if err != nil {
		return err
	}
	if len(list) != n {
		return fmt.Errorf("integrantes = %+v, se esperaban %d", list, n)
	}
	return nil
}

func (s *updateScenario) hasMember(projectName, name, role string) error {
	list, err := s.repo.ListByProject(s.projectIDs[projectName])
	if err != nil {
		return err
	}
	if len(list) != 1 || list[0].Name != name || list[0].Role != Role(role) {
		return fmt.Errorf("integrantes = %+v, se esperaba solo %q con el rol %q", list, name, role)
	}
	return nil
}
