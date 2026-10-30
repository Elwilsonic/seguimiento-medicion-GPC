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

// Ejecuta los escenarios BDD de features/HU-19-modificar-fechas-proyecto.feature con godog.
func TestFeatures_HU19ModificarFechasProyecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeDatesScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-19-modificar-fechas-proyecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-19")
	}
}

// datesScenario guarda el estado de un escenario de HU-19.
type datesScenario struct {
	manager  *project.Manager
	repo     *updateFailingRepository
	returned project.Project
	err      error
	// before es el proyecto 1 guardado justo antes de la modificación.
	before project.Project
	// unexpected junta los resultados de las operaciones simultáneas que no
	// coinciden con el esperado.
	unexpected []string
}

func initializeDatesScenario(ctx *godog.ScenarioContext) {
	s := &datesScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.repo = &updateFailingRepository{MemoryRepository: project.NewMemoryRepository()}
		s.manager = project.NewManager(s.repo)
		s.returned = project.Project{}
		s.err = nil
		s.before = project.Project{}
		s.unexpected = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" y descripción "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)" en el estado "([^"]*)"$`, s.existingProjectWithStatus)
	ctx.Step(`^el proyecto con el ID (\d+) tiene registradas la fecha de inicio "([^"]*)" y la de fin "([^"]*)"$`, s.projectWithDates)
	ctx.Step(`^el proyecto con el ID (\d+) tiene registrada ninguna fecha$`, s.projectWithoutDates)
	ctx.Step(`^el proyecto con el ID (\d+) tiene registrada solo la fecha de inicio "([^"]*)"$`, s.projectWithStartDateOnly)
	ctx.Step(`^el repositorio falla al actualizar$`, s.repositoryFailsOnUpdate)
	ctx.Step(`^modifico las fechas del proyecto con el ID (-?\d+) a inicio "([^"]*)" y fin "([^"]*)"$`, s.setDates)
	ctx.Step(`^cambio las fechas del proyecto devuelto a inicio "([^"]*)" y fin "([^"]*)"$`, s.changeReturnedDates)
	ctx.Step(`^modifico en simultáneo (\d+) veces las fechas del proyecto con el ID (\d+) a inicio "([^"]*)" y fin "([^"]*)" y a inicio "([^"]*)" y fin "([^"]*)", y registro las fechas de inicio "([^"]*)" y "([^"]*)" y las de fin "([^"]*)" y "([^"]*)"$`, s.setDatesConcurrently)
	ctx.Step(`^la modificación es exitosa$`, s.modifiedOK)
	ctx.Step(`^el proyecto devuelto tiene la fecha de inicio "([^"]*)" y la de fin "([^"]*)"$`, s.returnedHasDates)
	ctx.Step(`^el proyecto guardado con el ID (\d+) tiene la fecha de inicio "([^"]*)" y la de fin "([^"]*)"$`, s.storedHasDates)
	ctx.Step(`^el proyecto guardado con el ID (\d+) sigue teniendo el nombre "([^"]*)", la descripción "([^"]*)" y el estado "([^"]*)"$`, s.storedHasData)
	ctx.Step(`^el proyecto guardado con el ID (\d+) no cambió$`, s.storedUnchanged)
	ctx.Step(`^cada operación simultánea termina con el resultado esperado$`, s.concurrentResultsOK)
	ctx.Step(`^la fecha de fin guardada del proyecto con el ID (\d+) es posterior a la de inicio$`, s.storedEndAfterStart)
	ctx.Step(`^la modificación falla porque la fecha de fin debe ser posterior a la de inicio$`, s.failsWith(project.ErrEndNotAfterStart))
	ctx.Step(`^la modificación falla porque el proyecto no existe$`, s.failsWith(project.ErrProjectNotFound))
	ctx.Step(`^la modificación falla porque el proyecto no tiene registradas las dos fechas$`, s.failsWith(project.ErrDatesNotRegistered))
	ctx.Step(`^la modificación falla porque la fecha de inicio es obligatoria$`, s.failsWith(project.ErrMissingStartDate))
	ctx.Step(`^la modificación falla porque la fecha de fin es obligatoria$`, s.failsWith(project.ErrMissingEndDate))
	ctx.Step(`^la modificación falla con el error del repositorio$`, s.failsWith(errRepository))
}

func (s *datesScenario) existingProject(name, description string) error {
	_, err := s.manager.Create(name, description)
	return err
}

func (s *datesScenario) existingProjectWithStatus(name, status string) error {
	p, err := s.manager.Create(name, "")
	if err != nil {
		return err
	}
	_, err = s.manager.Update(p.ID, name, "", status)
	return err
}

// projectWithDates deja las dos fechas registradas con las operaciones de
// HU-18 y HU-41.
func (s *datesScenario) projectWithDates(id int, start, end string) error {
	startDay, err := parseDay(start)
	if err != nil {
		return err
	}
	endDay, err := parseDay(end)
	if err != nil {
		return err
	}
	if _, err := s.manager.SetStartDate(id, startDay); err != nil {
		return fmt.Errorf("SetStartDate devolvió un error inesperado: %v", err)
	}
	if _, err := s.manager.SetEndDate(id, endDay); err != nil {
		return fmt.Errorf("SetEndDate devolvió un error inesperado: %v", err)
	}
	return s.storedHasDates(id, start, end)
}

// projectWithoutDates verifica que el proyecto recién creado no tenga fechas.
func (s *datesScenario) projectWithoutDates(id int) error {
	p, err := s.stored(id)
	if err != nil {
		return err
	}
	if !p.StartDate.IsZero() || !p.EndDate.IsZero() {
		return fmt.Errorf("proyecto = %+v, se esperaba sin fechas", p)
	}
	return nil
}

// projectWithStartDateOnly registra solo la fecha de inicio (HU-18).
func (s *datesScenario) projectWithStartDateOnly(id int, date string) error {
	day, err := parseDay(date)
	if err != nil {
		return err
	}
	if _, err := s.manager.SetStartDate(id, day); err != nil {
		return fmt.Errorf("SetStartDate devolvió un error inesperado: %v", err)
	}
	return nil
}

func (s *datesScenario) repositoryFailsOnUpdate() error {
	s.repo.failUpdate = true
	return nil
}

// optionalDay lee una fecha AAAA-MM-DD; la cadena vacía es la fecha ausente
// (valor cero), como la recibe el Manager desde la API.
func optionalDay(date string) (time.Time, error) {
	if date == "" {
		return time.Time{}, nil
	}
	return parseDay(date)
}

// setDates guarda el proyecto 1 antes de modificar, para comparar después.
func (s *datesScenario) setDates(id int, start, end string) error {
	startDay, err := optionalDay(start)
	if err != nil {
		return err
	}
	endDay, err := optionalDay(end)
	if err != nil {
		return err
	}
	if s.before, err = s.stored(1); err != nil {
		return err
	}
	s.returned, s.err = s.manager.SetDates(id, startDay, endDay)
	return nil
}

func (s *datesScenario) changeReturnedDates(start, end string) error {
	if err := s.modifiedOK(); err != nil {
		return err
	}
	startDay, err := parseDay(start)
	if err != nil {
		return err
	}
	endDay, err := parseDay(end)
	if err != nil {
		return err
	}
	s.returned.StartDate = startDay
	s.returned.EndDate = endDay
	return nil
}

// setDatesConcurrently modifica y registra las fechas en simultáneo, times
// veces cada operación: un par válido (pairStart, pairEnd), un par con el fin
// anterior al inicio (badStart, badEnd), una fecha de inicio válida
// (validStart), una posterior a cualquier fin (lateStart), una fecha de fin
// válida (validEnd) y una anterior a cualquier inicio (earlyEnd), además de
// consultas. Con las fechas del escenario, el resultado de cada operación no
// depende del orden: las válidas terminan bien y las otras con su error.
func (s *datesScenario) setDatesConcurrently(times, id int, pairStart, pairEnd, badStart, badEnd, validStart, lateStart, validEnd, earlyEnd string) error {
	dates := map[string]time.Time{}
	for _, date := range []string{pairStart, pairEnd, badStart, badEnd, validStart, lateStart, validEnd, earlyEnd} {
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
		wg.Add(7)
		go func() {
			defer wg.Done()
			_, err := s.manager.SetDates(id, dates[pairStart], dates[pairEnd])
			check("SetDates", pairStart+", "+pairEnd, err, nil)
		}()
		go func() {
			defer wg.Done()
			_, err := s.manager.SetDates(id, dates[badStart], dates[badEnd])
			check("SetDates", badStart+", "+badEnd, err, project.ErrEndNotAfterStart)
		}()
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

func (s *datesScenario) modifiedOK() error {
	if s.err != nil {
		return fmt.Errorf("la modificación falló: %v", s.err)
	}
	return nil
}

func (s *datesScenario) returnedHasDates(start, end string) error {
	if err := s.modifiedOK(); err != nil {
		return err
	}
	return expectDates(s.returned, start, end)
}

func (s *datesScenario) stored(id int) (project.Project, error) {
	p, err := s.manager.Get(id)
	if err != nil {
		return project.Project{}, fmt.Errorf("Get(%d) devolvió un error inesperado: %v", id, err)
	}
	return p, nil
}

func (s *datesScenario) storedHasDates(id int, start, end string) error {
	p, err := s.stored(id)
	if err != nil {
		return err
	}
	return expectDates(p, start, end)
}

func (s *datesScenario) storedHasData(id int, name, description, status string) error {
	p, err := s.stored(id)
	if err != nil {
		return err
	}
	if p.Name != name || p.Description != description || string(p.Status) != status {
		return fmt.Errorf("proyecto = %+v, se esperaba nombre %q, descripción %q y estado %q", p, name, description, status)
	}
	return nil
}

func (s *datesScenario) storedUnchanged(id int) error {
	p, err := s.stored(id)
	if err != nil {
		return err
	}
	if p != s.before {
		return fmt.Errorf("el proyecto guardado cambió: antes %+v, después %+v", s.before, p)
	}
	return nil
}

func (s *datesScenario) concurrentResultsOK() error {
	if len(s.unexpected) > 0 {
		return fmt.Errorf("%d operaciones simultáneas con un resultado inesperado, por ejemplo: %s", len(s.unexpected), s.unexpected[0])
	}
	return nil
}

func (s *datesScenario) storedEndAfterStart(id int) error {
	p, err := s.stored(id)
	if err != nil {
		return err
	}
	if !p.EndDate.After(p.StartDate) {
		return fmt.Errorf("la fecha de fin %v no es posterior a la de inicio %v", p.EndDate, p.StartDate)
	}
	return nil
}

func (s *datesScenario) failsWith(want error) func() error {
	return func() error {
		if !errors.Is(s.err, want) {
			return fmt.Errorf("err = %v, se esperaba %v", s.err, want)
		}
		return nil
	}
}

// expectDates verifica que p tenga las fechas de inicio start y de fin end
// (AAAA-MM-DD).
func expectDates(p project.Project, start, end string) error {
	if err := expectStartDate(p, start); err != nil {
		return err
	}
	return expectEndDate(p, end)
}
