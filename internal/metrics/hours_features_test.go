package metrics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-33-horas-desviacion.feature con
// godog.
func TestFeatures_HU33CalcularHorasDesviacion(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeHoursScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-33-horas-desviacion.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-33")
	}
}

// hoursScenario guarda el estado de un escenario de HU-33.
type hoursScenario struct {
	t        *testing.T
	env      *metricsEnv
	projects map[string]int
	sprints  map[string]sprint.Sprint

	hours Hours
	err   error

	itemsBefore   []backlog.Item
	sprintsBefore []sprint.Sprint
	effortsBefore []effort.Effort
	projectID     int
}

func initializeHoursScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &hoursScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.env = newMetricsEnv()
		s.projects = map[string]int{}
		s.sprints = map[string]sprint.Sprint{}
		s.hours = Hours{}
		s.err = nil
		s.itemsBefore, s.sprintsBefore, s.effortsBefore = nil, nil, nil
		s.projectID = 0
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que cierro el sprint "([^"]*)"$`, s.closeSprint)
	ctx.Step(`^que el sprint "([^"]*)" tiene los ítems con horas y esfuerzos:$`, s.sprintItems)
	ctx.Step(`^que el proyecto "([^"]*)" tiene los ítems sin sprint con horas:$`, s.projectItems)
	ctx.Step(`^consulto las horas del sprint "([^"]*)"$`, s.querySprintByName)
	ctx.Step(`^consulto las horas del sprint con el ID (-?\d+)$`, s.querySprint)
	ctx.Step(`^consulto las horas del proyecto "([^"]*)"$`, s.queryProjectByName)
	ctx.Step(`^consulto las horas del proyecto con el ID (-?\d+)$`, s.queryProject)
	ctx.Step(`^las horas estimadas son ([\d.]+)$`, s.estimatedAre)
	ctx.Step(`^las horas reales son ([\d.]+)$`, s.actualAre)
	ctx.Step(`^la desviación en horas es (-?[\d.]+)$`, s.deviationAre)
	ctx.Step(`^la desviación porcentual es (-?[\d.]+)$`, s.deviationPercentAre)
	ctx.Step(`^hay (\d+) ítems? sin horas estimadas$`, s.unestimatedAre)
	ctx.Step(`^la desviación no está disponible$`, s.deviationNotAvailable)
	ctx.Step(`^la consulta falla porque el sprint no existe$`, s.failsSprintNotFound)
	ctx.Step(`^la consulta falla porque el proyecto no existe$`, s.failsProjectNotFound)
	ctx.Step(`^los ítems, los sprints y los esfuerzos del proyecto no cambiaron$`, s.nothingChanged)
}

func (s *hoursScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name)
	s.projectID = s.projects[name]
	return nil
}

func (s *hoursScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	s.projectID = s.projects[projectName]
	return nil
}

func (s *hoursScenario) closeSprint(name string) error {
	closed, err := s.env.sprints.Close(s.sprints[name].ID)
	if err != nil {
		return fmt.Errorf("cerrar el sprint %q: %w", name, err)
	}
	s.sprints[name] = closed
	return nil
}

// sprintItems crea los ítems de la tabla (titulo, horas, esfuerzo) en el
// sprint y registra el esfuerzo cuando es mayor que cero.
func (s *hoursScenario) sprintItems(name string, table *godog.Table) error {
	sp := s.sprints[name]
	s.projectID = sp.ProjectID
	for _, row := range table.Rows[1:] {
		hours, err := strconv.ParseFloat(row.Cells[1].Value, 64)
		if err != nil {
			return fmt.Errorf("horas inválidas %q: %w", row.Cells[1].Value, err)
		}
		effortHours, err := strconv.ParseFloat(row.Cells[2].Value, 64)
		if err != nil {
			return fmt.Errorf("esfuerzo inválido %q: %w", row.Cells[2].Value, err)
		}
		it := s.env.itemWithHours(s.t, sp.ProjectID, row.Cells[0].Value, hours, sp.ID)
		s.env.registerEffort(s.t, sp.ID, it.ID, effortHours)
	}
	return nil
}

// projectItems crea los ítems de la tabla (titulo, horas) sin sprint.
func (s *hoursScenario) projectItems(projectName string, table *godog.Table) error {
	projectID := s.projects[projectName]
	s.projectID = projectID
	for _, row := range table.Rows[1:] {
		hours, err := strconv.ParseFloat(row.Cells[1].Value, 64)
		if err != nil {
			return fmt.Errorf("horas inválidas %q: %w", row.Cells[1].Value, err)
		}
		s.env.itemWithHours(s.t, projectID, row.Cells[0].Value, hours, 0)
	}
	return nil
}

// snapshot guarda los ítems, sprints y esfuerzos antes de la primera consulta
// para comprobar después que no cambiaron (CA-33.8).
func (s *hoursScenario) snapshot() {
	if s.itemsBefore != nil || s.projectID == 0 {
		return
	}
	s.itemsBefore, _ = s.env.itemRepo.ListByProject(s.projectID)
	s.sprintsBefore = s.env.sprintsOf(s.projectID)
	for _, sp := range s.sprintsBefore {
		efforts, _ := s.env.effortRepo.ListBySprint(sp.ID)
		s.effortsBefore = append(s.effortsBefore, efforts...)
	}
}

func (s *hoursScenario) querySprintByName(name string) error {
	return s.querySprint(s.sprints[name].ID)
}

func (s *hoursScenario) querySprint(id int) error {
	s.snapshot()
	s.hours, s.err = s.env.metrics.SprintHours(id)
	return nil
}

func (s *hoursScenario) queryProjectByName(name string) error {
	return s.queryProject(s.projects[name])
}

func (s *hoursScenario) queryProject(id int) error {
	s.snapshot()
	s.hours, s.err = s.env.metrics.ProjectHours(id)
	return nil
}

func (s *hoursScenario) result() (Hours, error) {
	if s.err != nil {
		return Hours{}, fmt.Errorf("la consulta falló: %v", s.err)
	}
	return s.hours, nil
}

func (s *hoursScenario) estimatedAre(want float64) error {
	got, err := s.result()
	if err != nil {
		return err
	}
	if !almostEqual(got.Estimated, want) {
		return fmt.Errorf("horas estimadas = %v, se esperaba %v", got.Estimated, want)
	}
	return nil
}

func (s *hoursScenario) actualAre(want float64) error {
	got, err := s.result()
	if err != nil {
		return err
	}
	if !almostEqual(got.Actual, want) {
		return fmt.Errorf("horas reales = %v, se esperaba %v", got.Actual, want)
	}
	return nil
}

func (s *hoursScenario) deviationAre(want float64) error {
	got, err := s.result()
	if err != nil {
		return err
	}
	if !got.DeviationAvailable || !almostEqual(got.Deviation, want) {
		return fmt.Errorf("desviación = %v (disponible %v), se esperaba %v", got.Deviation, got.DeviationAvailable, want)
	}
	return nil
}

func (s *hoursScenario) deviationPercentAre(want float64) error {
	got, err := s.result()
	if err != nil {
		return err
	}
	if !got.DeviationAvailable || !almostEqual(got.DeviationPercent, want) {
		return fmt.Errorf("desviación porcentual = %v (disponible %v), se esperaba %v", got.DeviationPercent, got.DeviationAvailable, want)
	}
	return nil
}

func (s *hoursScenario) unestimatedAre(want int) error {
	got, err := s.result()
	if err != nil {
		return err
	}
	if got.Unestimated != want {
		return fmt.Errorf("ítems sin horas estimadas = %d, se esperaba %d", got.Unestimated, want)
	}
	return nil
}

func (s *hoursScenario) deviationNotAvailable() error {
	got, err := s.result()
	if err != nil {
		return err
	}
	if got.DeviationAvailable || got.Deviation != 0 || got.DeviationPercent != 0 {
		return fmt.Errorf("desviación = %v (%v %%), disponible %v, se esperaba no disponible", got.Deviation, got.DeviationPercent, got.DeviationAvailable)
	}
	return nil
}

func (s *hoursScenario) failsSprintNotFound() error {
	if !errors.Is(s.err, sprint.ErrSprintNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, sprint.ErrSprintNotFound)
	}
	return nil
}

func (s *hoursScenario) failsProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	return nil
}

func (s *hoursScenario) nothingChanged() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	items, err := s.env.itemRepo.ListByProject(s.projectID)
	if err != nil {
		return err
	}
	sprints := s.env.sprintsOf(s.projectID)
	var efforts []effort.Effort
	for _, sp := range sprints {
		found, err := s.env.effortRepo.ListBySprint(sp.ID)
		if err != nil {
			return err
		}
		efforts = append(efforts, found...)
	}
	if s.itemsBefore == nil || !reflect.DeepEqual(items, s.itemsBefore) ||
		!reflect.DeepEqual(sprints, s.sprintsBefore) ||
		!reflect.DeepEqual(efforts, s.effortsBefore) {
		return fmt.Errorf("la consulta modificó los datos: ítems %+v, sprints %+v, esfuerzos %+v", items, sprints, efforts)
	}
	return nil
}

func almostEqual(got, want float64) bool {
	return math.Abs(got-want) < 1e-9
}
