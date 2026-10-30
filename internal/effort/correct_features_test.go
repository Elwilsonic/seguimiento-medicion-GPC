package effort

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de features/HU-28-corregir-esfuerzo.feature con
// godog.
func TestFeatures_HU28CorregirEsfuerzo(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeCorrectScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-28-corregir-esfuerzo.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-28")
	}
}

// correctScenario guarda el estado de un escenario de HU-28.
type correctScenario struct {
	t       *testing.T
	env     *effortEnv
	sprints map[string]sprint.Sprint
	items   map[string]backlog.Item

	// registered son los esfuerzos registrados, por ID, tal como se
	// guardaron, para verificar que la corrección no cambie los demás campos.
	registered map[int]Effort
	before     []Effort
	result     Effort
	err        error
}

func initializeCorrectScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &correctScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*s = correctScenario{
			t:          t,
			env:        newEffortEnv(t),
			sprints:    map[string]sprint.Sprint{},
			items:      map[string]backlog.Item{},
			registered: map[int]Effort{},
		}
		return ctx, nil
	})

	ctx.Step(`^que existe un sprint abierto "([^"]*)"$`, s.openSprint)
	ctx.Step(`^que la historia "([^"]*)" está en el sprint "([^"]*)"$`, s.itemInSprint)
	ctx.Step(`^que "([^"]*)" registró ([\d.]+) horas en la historia "([^"]*)" del sprint "([^"]*)"$`, s.registeredEffort)
	ctx.Step(`^que el sprint "([^"]*)" está cerrado$`, s.closeSprint)
	ctx.Step(`^corrijo las horas del esfuerzo (-?\d+) a (\S+)$`, s.correct)
	ctx.Step(`^la corrección devuelve el esfuerzo (\d+) con ([\d.]+) horas$`, s.resultIs)
	ctx.Step(`^las horas guardadas del esfuerzo (\d+) son ([\d.]+)$`, s.storedHoursAre)
	ctx.Step(`^el integrante, la fecha, la actividad, la historia y el sprint del esfuerzo (\d+) no cambiaron$`, s.otherFieldsKept)
	ctx.Step(`^la corrección falla porque las horas son inválidas$`, s.failsWith(ErrInvalidHours))
	ctx.Step(`^la corrección falla porque el sprint está cerrado$`, s.failsWith(sprint.ErrSprintClosed))
	ctx.Step(`^la corrección falla porque el esfuerzo no existe$`, s.failsWith(ErrEffortNotFound))
	ctx.Step(`^los esfuerzos guardados no cambiaron$`, s.nothingChanged)
	ctx.Step(`^el total de horas del sprint "([^"]*)" es ([\d.]+)$`, s.sprintTotalIs)
}

func (s *correctScenario) openSprint(name string) error {
	s.sprints[name] = s.env.addSprint(s.t, name, sprint.StatusOpen)
	return nil
}

func (s *correctScenario) itemInSprint(title, sprintName string) error {
	s.items[title] = s.env.addItem(s.t, title, backlog.StatusInProgress, s.sprints[sprintName].ID)
	return nil
}

func (s *correctScenario) registeredEffort(member, hoursText, title, sprintName string) error {
	hours, err := strconv.ParseFloat(hoursText, 64)
	if err != nil {
		return err
	}
	saved, err := s.env.manager.Register(s.sprints[sprintName].ID, s.items[title].ID, member, testEffortDate, "Desarrollo", hours)
	if err != nil {
		return err
	}
	s.registered[saved.ID] = saved
	return nil
}

func (s *correctScenario) closeSprint(name string) error {
	sp := s.sprints[name]
	sp.Status = sprint.StatusClosed
	_, err := s.env.sprintRepo.Update(sp)
	return err
}

// stored lee los esfuerzos guardados directo del repositorio, sin pasar por
// GetByID, que es parte de lo que se prueba.
func (s *correctScenario) stored() []Effort {
	return s.env.memory.efforts
}

func (s *correctScenario) correct(idText, hoursText string) error {
	id, err := strconv.Atoi(idText)
	if err != nil {
		return err
	}
	hours, err := strconv.ParseFloat(hoursText, 64)
	if err != nil {
		return err
	}
	s.before = append([]Effort(nil), s.stored()...)
	s.result, s.err = s.env.manager.CorrectHours(id, hours)
	return nil
}

func (s *correctScenario) resultIs(id int, hours float64) error {
	if s.err != nil {
		return fmt.Errorf("la corrección falló: %v", s.err)
	}
	if s.result.ID != id || s.result.Hours != hours {
		return fmt.Errorf("resultado = %+v, se esperaba el esfuerzo %d con %v horas", s.result, id, hours)
	}
	return nil
}

func (s *correctScenario) storedOf(id int) (Effort, error) {
	for _, e := range s.stored() {
		if e.ID == id {
			return e, nil
		}
	}
	return Effort{}, fmt.Errorf("no hay un esfuerzo guardado con ID %d", id)
}

func (s *correctScenario) storedHoursAre(id int, hours float64) error {
	e, err := s.storedOf(id)
	if err != nil {
		return err
	}
	if e.Hours != hours {
		return fmt.Errorf("horas guardadas = %v, se esperaba %v", e.Hours, hours)
	}
	return nil
}

func (s *correctScenario) otherFieldsKept(id int) error {
	e, err := s.storedOf(id)
	if err != nil {
		return err
	}
	want := s.registered[id]
	want.Hours = e.Hours
	if e != want {
		return fmt.Errorf("esfuerzo guardado = %+v, se esperaba %+v salvo las horas", e, want)
	}
	if s.result.Member != want.Member || !s.result.Date.Equal(want.Date) || s.result.Activity != want.Activity ||
		s.result.ItemID != want.ItemID || s.result.SprintID != want.SprintID {
		return fmt.Errorf("resultado = %+v, se esperaban los campos de %+v", s.result, want)
	}
	return nil
}

func (s *correctScenario) failsWith(want error) func() error {
	return func() error {
		if !errors.Is(s.err, want) {
			return fmt.Errorf("err = %v, se esperaba %v", s.err, want)
		}
		if s.result != (Effort{}) {
			return fmt.Errorf("resultado = %+v, se esperaba el valor cero", s.result)
		}
		return nil
	}
}

func (s *correctScenario) nothingChanged() error {
	if after := s.stored(); !reflect.DeepEqual(after, s.before) {
		return fmt.Errorf("esfuerzos = %+v, se esperaba %+v", after, s.before)
	}
	return nil
}

func (s *correctScenario) sprintTotalIs(name string, total float64) error {
	summary, err := s.env.manager.ListBySprint(s.sprints[name].ID)
	if err != nil {
		return err
	}
	if summary.TotalHours != total {
		return fmt.Errorf("total = %v, se esperaba %v", summary.TotalHours, total)
	}
	return nil
}
