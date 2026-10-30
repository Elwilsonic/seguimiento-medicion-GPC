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

// Ejecuta los escenarios BDD de features/HU-17-consultar-integrantes.feature
// con godog.
func TestFeatures_HU17ConsultarIntegrantes(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeListScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-17-consultar-integrantes.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-17")
	}
}

// listScenario guarda el estado de un escenario de HU-17.
type listScenario struct {
	projects   *project.Manager
	repo       *listStubRepository
	members    *Manager
	projectIDs map[string]int
	memberIDs  map[string]int

	result []Member
	err    error
}

func initializeListScenario(ctx *godog.ScenarioContext) {
	s := &listScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		projects := project.NewManager(project.NewMemoryRepository())
		repo := &listStubRepository{MemoryRepository: NewMemoryRepository()}
		*s = listScenario{projects: projects, repo: repo, members: NewManager(repo, projects), projectIDs: map[string]int{}, memberIDs: map[string]int{}}
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el integrante "([^"]*)" está registrado con el rol "([^"]*)" en el proyecto "([^"]*)"$`, s.registeredMember)
	ctx.Step(`^que el integrante "([^"]*)" se modificó con el nombre "([^"]*)" y el rol "([^"]*)"$`, s.updatedMember)
	ctx.Step(`^que el proyecto "([^"]*)" está finalizado$`, s.finishedProject)
	ctx.Step(`^que el repositorio de integrantes falla al listar$`, s.repoFailsOnList)
	ctx.Step(`^consulto los integrantes del proyecto "([^"]*)"$`, s.listByName)
	ctx.Step(`^consulto los integrantes del proyecto con el ID (-?\d+)$`, s.list)
	ctx.Step(`^cambio el nombre y el rol del primer integrante devuelto$`, s.mutateFirstMember)
	ctx.Step(`^la consulta es exitosa$`, s.listOK)
	ctx.Step(`^obtengo estos integrantes:$`, s.membersAre)
	ctx.Step(`^obtengo una lista vacía de integrantes$`, s.membersEmpty)
	ctx.Step(`^la consulta falla porque no existe un proyecto con ese id$`, s.failsProjectNotFound)
	ctx.Step(`^la consulta falla con el error del repositorio$`, s.failsRepository)
}

func (s *listScenario) existingProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectIDs[name] = p.ID
	return nil
}

func (s *listScenario) projectID(projectName string) (int, error) {
	id, ok := s.projectIDs[projectName]
	if !ok {
		return 0, fmt.Errorf("no existe el proyecto de prueba %q", projectName)
	}
	return id, nil
}

func (s *listScenario) registeredMember(name, role, projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	m, err := s.members.Register(projectID, name, role)
	if err != nil {
		return err
	}
	s.memberIDs[name] = m.ID
	return nil
}

func (s *listScenario) updatedMember(oldName, name, role string) error {
	id, ok := s.memberIDs[oldName]
	if !ok {
		return fmt.Errorf("no existe el integrante de prueba %q", oldName)
	}
	_, err := s.members.Update(id, name, role)
	return err
}

func (s *listScenario) finishedProject(projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	_, err = s.projects.Update(projectID, projectName, "", string(project.StatusFinished))
	return err
}

func (s *listScenario) repoFailsOnList() error {
	s.repo.listErr = errStorage
	return nil
}

func (s *listScenario) listByName(projectName string) error {
	projectID, err := s.projectID(projectName)
	if err != nil {
		return err
	}
	return s.list(projectID)
}

func (s *listScenario) list(projectID int) error {
	s.result, s.err = s.members.List(projectID)
	return nil
}

func (s *listScenario) mutateFirstMember() error {
	if len(s.result) == 0 {
		return errors.New("la lista de integrantes está vacía")
	}
	s.result[0].Name = "Cambiado"
	s.result[0].Role = RoleAgileEnabler
	return nil
}

func (s *listScenario) listOK() error {
	if s.err != nil {
		return fmt.Errorf("se esperaba una consulta exitosa, falló con: %v", s.err)
	}
	return nil
}

func (s *listScenario) membersAre(table *godog.Table) error {
	rows := table.Rows[1:]
	if len(s.result) != len(rows) {
		return fmt.Errorf("se obtuvieron %d integrantes, se esperaban %d: %+v", len(s.result), len(rows), s.result)
	}
	for i, row := range rows {
		id, err := strconv.Atoi(row.Cells[0].Value)
		if err != nil {
			return err
		}
		got := s.result[i]
		if got.ID != id || got.Name != row.Cells[1].Value || string(got.Role) != row.Cells[2].Value {
			return fmt.Errorf("integrante %d = %d %q %q, se esperaba %d %q %q",
				i+1, got.ID, got.Name, got.Role, id, row.Cells[1].Value, row.Cells[2].Value)
		}
	}
	return nil
}

func (s *listScenario) membersEmpty() error {
	if s.result == nil || len(s.result) != 0 {
		return fmt.Errorf("se esperaba una lista vacía, no nil, y se obtuvo %#v", s.result)
	}
	return nil
}

func (s *listScenario) failsProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	if s.result != nil {
		return fmt.Errorf("se esperaba nil y se obtuvo %#v", s.result)
	}
	return nil
}

func (s *listScenario) failsRepository() error {
	if !errors.Is(s.err, errStorage) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, errStorage)
	}
	if s.result != nil {
		return fmt.Errorf("se esperaba nil y se obtuvo %#v", s.result)
	}
	return nil
}
