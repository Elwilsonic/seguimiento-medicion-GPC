package project_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-50 · Listar los proyectos (Issue #183)

// errListFails es el error del almacenamiento falso de los escenarios de error.
var errListFails = errors.New("falla del almacenamiento al listar")

// HU-50 · CA-50.1 · Verifica que se cumplan los escenarios de features/HU-50-listar-proyectos.feature con godog.
func TestFeatures_HU50ListarProyectos(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeListScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-50-listar-proyectos.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-50")
	}
}

// listScenario guarda el estado de un escenario de HU-50.
type listScenario struct {
	manager *project.Manager
	// ids guarda el ID que se le asignó a cada proyecto al crearlo, por nombre.
	ids  map[string]int
	got  []project.Project
	err  error
	done bool
}

func initializeListScenario(ctx *godog.ScenarioContext) {
	s := &listScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.manager = project.NewManager(project.NewMemoryRepository())
		s.ids = map[string]int{}
		s.got = nil
		s.err = nil
		s.done = false
		return ctx, nil
	})

	ctx.Step(`^que no hay proyectos creados$`, s.noProjects)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" y descripción "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el proyecto "([^"]*)" está en el estado "([^"]*)"$`, s.projectWithStatus)
	ctx.Step(`^que el proyecto "([^"]*)" tiene las fechas "([^"]*)" a "([^"]*)"$`, s.projectWithDates)
	ctx.Step(`^que el proyecto "([^"]*)" se modificó con el nombre "([^"]*)" y el estado "([^"]*)"$`, s.projectModified)
	ctx.Step(`^que el almacenamiento de proyectos falla al listar$`, s.failingStorage)
	ctx.Step(`^consulto la lista de proyectos$`, s.listProjects)
	ctx.Step(`^la consulta de la lista es exitosa$`, s.listOK)
	ctx.Step(`^obtengo estos proyectos, en este orden:$`, s.gotProjects)
	ctx.Step(`^cada proyecto de la lista tiene el ID que se le asignó al crearlo$`, s.gotAssignedIDs)
	ctx.Step(`^los nombres de la lista son, en este orden: (.+)$`, s.gotNames)
	ctx.Step(`^la lista está vacía y no es nula$`, s.gotEmptyNotNil)
	ctx.Step(`^cambio el nombre del primer proyecto de la lista a "([^"]*)" y su estado a "([^"]*)"$`, s.changeFirst)
	ctx.Step(`^al consultar el proyecto "([^"]*)" por su ID tiene el nombre "([^"]*)" y el estado "([^"]*)"$`, s.getHas)
	ctx.Step(`^al volver a consultar la lista los nombres son, en este orden: (.+)$`, s.relistNames)
	ctx.Step(`^la consulta de la lista falla con el error del almacenamiento$`, s.failsWithStorageError)
	ctx.Step(`^no se devuelve ninguna lista$`, s.noList)
}

func (s *listScenario) noProjects() error {
	if len(s.ids) != 0 {
		return fmt.Errorf("hay %d proyectos creados, se esperaba 0", len(s.ids))
	}
	return nil
}

func (s *listScenario) existingProject(name, description string) error {
	p, err := s.manager.Create(name, description)
	if err != nil {
		return err
	}
	s.ids[name] = p.ID
	return nil
}

func (s *listScenario) projectID(name string) (int, error) {
	id, ok := s.ids[name]
	if !ok {
		return 0, fmt.Errorf("no se creó el proyecto %q en el escenario", name)
	}
	return id, nil
}

func (s *listScenario) projectWithStatus(name, status string) error {
	id, err := s.projectID(name)
	if err != nil {
		return err
	}
	p, err := s.manager.Get(id)
	if err != nil {
		return err
	}
	_, err = s.manager.Update(id, p.Name, p.Description, status)
	return err
}

func (s *listScenario) projectWithDates(name, start, end string) error {
	id, err := s.projectID(name)
	if err != nil {
		return err
	}
	startDate, err := time.Parse(time.DateOnly, start)
	if err != nil {
		return err
	}
	endDate, err := time.Parse(time.DateOnly, end)
	if err != nil {
		return err
	}
	// SetDates (HU-19) exige que las fechas ya estén registradas, así que se
	// registran una por una con HU-18 y HU-41.
	if _, err := s.manager.SetStartDate(id, startDate); err != nil {
		return err
	}
	_, err = s.manager.SetEndDate(id, endDate)
	return err
}

func (s *listScenario) projectModified(name, newName, status string) error {
	id, err := s.projectID(name)
	if err != nil {
		return err
	}
	p, err := s.manager.Get(id)
	if err != nil {
		return err
	}
	if _, err := s.manager.Update(id, newName, p.Description, status); err != nil {
		return err
	}
	delete(s.ids, name)
	s.ids[newName] = id
	return nil
}

func (s *listScenario) failingStorage() error {
	s.manager = project.NewManager(failingRepository{listErr: errListFails})
	return nil
}

func (s *listScenario) listProjects() error {
	s.got, s.err = s.manager.List()
	s.done = true
	return nil
}

func (s *listScenario) listOK() error {
	if !s.done {
		return errors.New("no se consultó la lista")
	}
	if s.err != nil {
		return fmt.Errorf("la consulta de la lista falló: %v", s.err)
	}
	return nil
}

// formatDate muestra una fecha como la API: AAAA-MM-DD, o "" si no está
// registrada.
func formatDate(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.Format(time.DateOnly)
}

func (s *listScenario) gotProjects(table *godog.Table) error {
	if err := s.listOK(); err != nil {
		return err
	}
	rows := table.Rows[1:]
	if len(s.got) != len(rows) {
		return fmt.Errorf("la lista tiene %d proyectos, se esperaban %d", len(s.got), len(rows))
	}
	for i, row := range rows {
		want := []string{row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value, row.Cells[3].Value, row.Cells[4].Value}
		p := s.got[i]
		got := []string{p.Name, p.Description, string(p.Status), formatDate(p.StartDate), formatDate(p.EndDate)}
		for j := range want {
			if got[j] != want[j] {
				return fmt.Errorf("proyecto %d, columna %q = %q, se esperaba %q", i+1, table.Rows[0].Cells[j].Value, got[j], want[j])
			}
		}
	}
	return nil
}

func (s *listScenario) gotAssignedIDs() error {
	for _, p := range s.got {
		id, err := s.projectID(p.Name)
		if err != nil {
			return err
		}
		if p.ID != id {
			return fmt.Errorf("el proyecto %q tiene el ID %d, se le asignó %d", p.Name, p.ID, id)
		}
	}
	return nil
}

// parseNames lee una lista de nombres entre comillas separados por comas:
// "A", "B", "C".
func parseNames(list string) []string {
	var names []string
	for _, part := range strings.Split(list, ",") {
		names = append(names, strings.Trim(strings.TrimSpace(part), `"`))
	}
	return names
}

func namesOf(projects []project.Project) []string {
	names := make([]string, 0, len(projects))
	for _, p := range projects {
		names = append(names, p.Name)
	}
	return names
}

func compareNames(got []project.Project, list string) error {
	want := parseNames(list)
	names := namesOf(got)
	if strings.Join(names, "|") != strings.Join(want, "|") {
		return fmt.Errorf("nombres = %q, se esperaba %q", names, want)
	}
	return nil
}

func (s *listScenario) gotNames(list string) error {
	if err := s.listOK(); err != nil {
		return err
	}
	return compareNames(s.got, list)
}

func (s *listScenario) gotEmptyNotNil() error {
	if err := s.listOK(); err != nil {
		return err
	}
	if s.got == nil {
		return errors.New("la lista es nil, se esperaba una lista vacía")
	}
	if len(s.got) != 0 {
		return fmt.Errorf("la lista tiene %d proyectos, se esperaba vacía", len(s.got))
	}
	return nil
}

func (s *listScenario) changeFirst(name, status string) error {
	if err := s.listOK(); err != nil {
		return err
	}
	if len(s.got) == 0 {
		return errors.New("la lista está vacía")
	}
	s.got[0].Name = name
	s.got[0].Status = project.Status(status)
	return nil
}

func (s *listScenario) getHas(name, wantName, wantStatus string) error {
	id, err := s.projectID(name)
	if err != nil {
		return err
	}
	p, err := s.manager.Get(id)
	if err != nil {
		return err
	}
	if p.Name != wantName || string(p.Status) != wantStatus {
		return fmt.Errorf("proyecto guardado = %q (%s), se esperaba %q (%s)", p.Name, p.Status, wantName, wantStatus)
	}
	return nil
}

func (s *listScenario) relistNames(list string) error {
	got, err := s.manager.List()
	if err != nil {
		return fmt.Errorf("la segunda consulta de la lista falló: %v", err)
	}
	return compareNames(got, list)
}

func (s *listScenario) failsWithStorageError() error {
	if !s.done {
		return errors.New("no se consultó la lista")
	}
	if !errors.Is(s.err, errListFails) {
		return fmt.Errorf("error = %v, se esperaba el error del almacenamiento", s.err)
	}
	return nil
}

func (s *listScenario) noList() error {
	if s.got != nil {
		return fmt.Errorf("se devolvió una lista con %d proyectos, se esperaba nil", len(s.got))
	}
	return nil
}
