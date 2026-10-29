package project_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-41-fecha-fin-proyecto.feature con godog.
func TestFeatures_HU41FechaFinProyecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeEndDateScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-41-fecha-fin-proyecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-41")
	}
}

// endDateScenario guarda el estado de un escenario de HU-41.
type endDateScenario struct {
	manager  *project.Manager
	repo     *updateFailingRepository
	returned project.Project
	err      error
	// unexpected junta los resultados de los registros simultáneos que no
	// coinciden con el esperado.
	unexpected []string
}

func initializeEndDateScenario(ctx *godog.ScenarioContext) {
	s := &endDateScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.repo = &updateFailingRepository{MemoryRepository: project.NewMemoryRepository()}
		s.manager = project.NewManager(s.repo)
		s.returned = project.Project{}
		s.err = nil
		s.unexpected = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" y descripción "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" en el estado "([^"]*)"$`, s.existingProjectWithStatus)
	ctx.Step(`^el proyecto con el ID (\d+) tiene registrada la fecha de inicio "([^"]*)"$`, s.projectWithStartDate)
	ctx.Step(`^el proyecto con el ID (\d+) tiene registrada la fecha de fin "([^"]*)"$`, s.projectWithEndDate)
	ctx.Step(`^el repositorio falla al actualizar$`, s.repositoryFailsOnUpdate)
	ctx.Step(`^registro la fecha de fin "([^"]*)" en el proyecto con el ID (-?\d+)$`, s.setEndDate)
	ctx.Step(`^registro una fecha de fin ausente en el proyecto con el ID (-?\d+)$`, s.setMissingEndDate)
	ctx.Step(`^registro la fecha de inicio "([^"]*)" en el proyecto con el ID (\d+)$`, s.setStartDate)
	ctx.Step(`^modifico el proyecto con el ID (\d+) con nombre "([^"]*)", descripción "([^"]*)" y estado "([^"]*)"$`, s.updateProject)
	ctx.Step(`^cambio la fecha de fin del proyecto devuelto a "([^"]*)"$`, s.changeReturnedEndDate)
	ctx.Step(`^registro en simultáneo (\d+) veces las fechas de inicio "([^"]*)" y "([^"]*)" y las de fin "([^"]*)" y "([^"]*)" en el proyecto con el ID (\d+)$`, s.setDatesConcurrently)
	ctx.Step(`^el registro es exitoso$`, s.registeredOK)
	ctx.Step(`^el proyecto devuelto tiene la fecha de fin "([^"]*)"$`, s.returnedHasEndDate)
	ctx.Step(`^el proyecto guardado con el ID (\d+) tiene la fecha de fin "([^"]*)"$`, s.storedHasEndDate)
	ctx.Step(`^el proyecto guardado con el ID (\d+) no tiene fecha de fin$`, s.storedWithoutEndDate)
	ctx.Step(`^el proyecto guardado con el ID (\d+) tiene la fecha de inicio "([^"]*)"$`, s.storedHasStartDate)
	ctx.Step(`^el proyecto guardado con el ID (\d+) sigue teniendo el nombre "([^"]*)", la descripción "([^"]*)" y el estado "([^"]*)"$`, s.storedHasData)
	ctx.Step(`^cada registro simultáneo termina con el resultado esperado$`, s.concurrentResultsOK)
	ctx.Step(`^el registro falla porque la fecha de fin es obligatoria$`, s.failsWith(project.ErrMissingEndDate))
	ctx.Step(`^el registro falla porque el proyecto no existe$`, s.failsWith(project.ErrProjectNotFound))
	ctx.Step(`^el registro falla porque el proyecto no tiene fecha de inicio$`, s.failsWith(project.ErrStartDateNotRegistered))
	ctx.Step(`^el registro falla porque la fecha de fin debe ser posterior a la de inicio$`, s.failsWith(project.ErrEndNotAfterStart))
	ctx.Step(`^el registro falla porque la fecha de inicio debe ser anterior a la de fin$`, s.failsWith(project.ErrStartNotBeforeEnd))
	ctx.Step(`^el registro falla con el error del repositorio$`, s.failsWith(errRepository))
}

func (s *endDateScenario) existingProject(name, description string) error {
	_, err := s.manager.Create(name, description)
	return err
}

func (s *endDateScenario) existingProjectWithStatus(name, status string) error {
	p, err := s.manager.Create(name, "")
	if err != nil {
		return err
	}
	_, err = s.manager.Update(p.ID, name, "", status)
	return err
}

// projectWithStartDate deja la fecha de inicio registrada con la operación de
// HU-18.
func (s *endDateScenario) projectWithStartDate(id int, date string) error {
	day, err := parseDay(date)
	if err != nil {
		return err
	}
	if _, err := s.manager.SetStartDate(id, day); err != nil {
		return fmt.Errorf("SetStartDate devolvió un error inesperado: %v", err)
	}
	return s.storedHasStartDate(id, date)
}

// projectWithEndDate deja la fecha de fin registrada con la operación de
// HU-41 y falla el paso si el registro no la guardó.
func (s *endDateScenario) projectWithEndDate(id int, date string) error {
	day, err := parseDay(date)
	if err != nil {
		return err
	}
	if _, err := s.manager.SetEndDate(id, day); err != nil {
		return fmt.Errorf("SetEndDate devolvió un error inesperado: %v", err)
	}
	return s.storedHasEndDate(id, date)
}

func (s *endDateScenario) repositoryFailsOnUpdate() error {
	s.repo.failUpdate = true
	return nil
}

func (s *endDateScenario) setEndDate(date string, id int) error {
	day, err := parseDay(date)
	if err != nil {
		return err
	}
	s.returned, s.err = s.manager.SetEndDate(id, day)
	return nil
}

func (s *endDateScenario) setMissingEndDate(id int) error {
	s.returned, s.err = s.manager.SetEndDate(id, time.Time{})
	return nil
}

func (s *endDateScenario) setStartDate(date string, id int) error {
	day, err := parseDay(date)
	if err != nil {
		return err
	}
	s.returned, s.err = s.manager.SetStartDate(id, day)
	return nil
}

func (s *endDateScenario) updateProject(id int, name, description, status string) error {
	_, err := s.manager.Update(id, name, description, status)
	return err
}

func (s *endDateScenario) changeReturnedEndDate(date string) error {
	if err := s.registeredOK(); err != nil {
		return err
	}
	day, err := parseDay(date)
	if err != nil {
		return err
	}
	s.returned.EndDate = day
	return nil
}

// setDatesConcurrently registra en simultáneo, times veces cada una, una
// fecha de inicio válida (validStart), una posterior al fin (lateStart), una
// fecha de fin válida (validEnd) y una anterior al inicio (earlyEnd). Con
// fechas de partida anteriores a validStart y posteriores a validEnd, el
// resultado de cada registro no depende del orden: las válidas terminan bien
// y las otras con su error de fechas.
func (s *endDateScenario) setDatesConcurrently(times int, validStart, lateStart, validEnd, earlyEnd string, id int) error {
	dates := map[string]time.Time{}
	for _, date := range []string{validStart, lateStart, validEnd, earlyEnd} {
		day, err := parseDay(date)
		if err != nil {
			return err
		}
		dates[date] = day
	}
	var mu sync.Mutex
	check := func(op, date string, err, want error) {
		if errors.Is(err, want) && (want != nil || err == nil) {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		s.unexpected = append(s.unexpected, fmt.Sprintf("%s(%s): err = %v, se esperaba %v", op, date, err, want))
	}

	var wg sync.WaitGroup
	for i := 0; i < times; i++ {
		wg.Add(5)
		go func() {
			defer wg.Done()
			_, err := s.manager.SetStartDate(id, dates[validStart])
			check("SetStartDate", validStart, err, nil)
		}()
		go func() {
			defer wg.Done()
			_, err := s.manager.SetStartDate(id, dates[lateStart])
			check("SetStartDate", lateStart, err, project.ErrStartNotBeforeEnd)
		}()
		go func() {
			defer wg.Done()
			_, err := s.manager.SetEndDate(id, dates[validEnd])
			check("SetEndDate", validEnd, err, nil)
		}()
		go func() {
			defer wg.Done()
			_, err := s.manager.SetEndDate(id, dates[earlyEnd])
			check("SetEndDate", earlyEnd, err, project.ErrEndNotAfterStart)
		}()
		go func() {
			defer wg.Done()
			_, err := s.manager.Get(id)
			check("Get", "", err, nil)
		}()
	}
	wg.Wait()
	return nil
}

func (s *endDateScenario) registeredOK() error {
	if s.err != nil {
		return fmt.Errorf("el registro falló: %v", s.err)
	}
	return nil
}

func (s *endDateScenario) returnedHasEndDate(date string) error {
	if err := s.registeredOK(); err != nil {
		return err
	}
	return expectEndDate(s.returned, date)
}

func (s *endDateScenario) stored(id int) (project.Project, error) {
	p, err := s.manager.Get(id)
	if err != nil {
		return project.Project{}, fmt.Errorf("Get(%d) devolvió un error inesperado: %v", id, err)
	}
	return p, nil
}

func (s *endDateScenario) storedHasEndDate(id int, date string) error {
	p, err := s.stored(id)
	if err != nil {
		return err
	}
	return expectEndDate(p, date)
}

func (s *endDateScenario) storedWithoutEndDate(id int) error {
	p, err := s.stored(id)
	if err != nil {
		return err
	}
	if !p.EndDate.IsZero() {
		return fmt.Errorf("fecha de fin = %v, se esperaba sin registrar", p.EndDate)
	}
	return nil
}

func (s *endDateScenario) storedHasStartDate(id int, date string) error {
	p, err := s.stored(id)
	if err != nil {
		return err
	}
	return expectStartDate(p, date)
}

func (s *endDateScenario) storedHasData(id int, name, description, status string) error {
	p, err := s.stored(id)
	if err != nil {
		return err
	}
	if p.Name != name || p.Description != description || string(p.Status) != status {
		return fmt.Errorf("proyecto = %+v, se esperaba nombre %q, descripción %q y estado %q", p, name, description, status)
	}
	return nil
}

func (s *endDateScenario) concurrentResultsOK() error {
	if len(s.unexpected) > 0 {
		return fmt.Errorf("%d registros simultáneos con un resultado inesperado, por ejemplo: %s", len(s.unexpected), s.unexpected[0])
	}
	return nil
}

func (s *endDateScenario) failsWith(want error) func() error {
	return func() error {
		if !errors.Is(s.err, want) {
			return fmt.Errorf("err = %v, se esperaba %v", s.err, want)
		}
		return nil
	}
}

// expectEndDate verifica que p tenga la fecha de fin date (AAAA-MM-DD).
func expectEndDate(p project.Project, date string) error {
	want, err := parseDay(date)
	if err != nil {
		return err
	}
	if !p.EndDate.Equal(want) {
		return fmt.Errorf("fecha de fin = %v, se esperaba %s", p.EndDate, date)
	}
	return nil
}
