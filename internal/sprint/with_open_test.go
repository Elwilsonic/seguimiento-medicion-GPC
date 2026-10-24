package sprint

import (
	"errors"
	"sync"
	"testing"
)

// HU-26 · Registrar esfuerzo (Issue #44)
//
// WithOpenSprint es el punto de coordinación que usa effort.Manager.Register:
// valida que el sprint exista y esté ABIERTO y ejecuta la función con el mutex
// del sprint tomado, con el mismo orden sprint → backlog/effort que Close.

func newWithOpenEnv(t *testing.T) (*Manager, *MemoryRepository, int) {
	t.Helper()
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	return manager, repo, projectID
}

func TestWithOpenSprint_SprintAbiertoEjecutaLaFuncion(t *testing.T) {
	manager, repo, projectID := newWithOpenEnv(t)
	saved, err := repo.Save(Sprint{
		ProjectID: projectID,
		Name:      validName,
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    StatusOpen,
	})
	if err != nil {
		t.Fatalf("guardar el sprint: %v", err)
	}

	called := false
	err = manager.WithOpenSprint(saved.ID, func(got Sprint) error {
		called = true
		if got != saved {
			t.Errorf("sprint recibido = %+v, se esperaba %+v", got, saved)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("WithOpenSprint devolvió un error inesperado: %v", err)
	}
	if !called {
		t.Error("la función no se ejecutó para un sprint ABIERTO")
	}
}

func TestWithOpenSprint_SprintCerradoDevuelveErrSprintClosed(t *testing.T) {
	manager, repo, projectID := newWithOpenEnv(t)
	saved, err := repo.Save(Sprint{
		ProjectID: projectID,
		Name:      validName,
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    StatusClosed,
	})
	if err != nil {
		t.Fatalf("guardar el sprint: %v", err)
	}

	called := false
	err = manager.WithOpenSprint(saved.ID, func(Sprint) error {
		called = true
		return nil
	})

	if !errors.Is(err, ErrSprintClosed) {
		t.Fatalf("err = %v, se esperaba %v", err, ErrSprintClosed)
	}
	if called {
		t.Error("la función se ejecutó para un sprint CERRADO")
	}
}

func TestWithOpenSprint_SprintInexistenteDevuelveErrSprintNotFound(t *testing.T) {
	manager, _, _ := newWithOpenEnv(t)

	called := false
	err := manager.WithOpenSprint(999, func(Sprint) error {
		called = true
		return nil
	})

	if !errors.Is(err, ErrSprintNotFound) {
		t.Fatalf("err = %v, se esperaba %v", err, ErrSprintNotFound)
	}
	if called {
		t.Error("la función se ejecutó para un sprint inexistente")
	}
}

func TestWithOpenSprint_PropagaElErrorDeLaFuncion(t *testing.T) {
	manager, repo, projectID := newWithOpenEnv(t)
	saved, err := repo.Save(Sprint{
		ProjectID: projectID,
		Name:      validName,
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    StatusOpen,
	})
	if err != nil {
		t.Fatalf("guardar el sprint: %v", err)
	}
	want := errors.New("error de la función")

	err = manager.WithOpenSprint(saved.ID, func(Sprint) error { return want })

	if !errors.Is(err, want) {
		t.Fatalf("err = %v, se esperaba %v", err, want)
	}
}

func TestWithOpenSprint_SerializaConClose(t *testing.T) {
	for i := 0; i < 20; i++ {
		manager, repo, projectID := newWithOpenEnv(t)
		saved, err := repo.Save(Sprint{
			ProjectID: projectID,
			Name:      validName,
			StartDate: validStart,
			EndDate:   validEnd,
			Goal:      validGoal,
			Status:    StatusOpen,
		})
		if err != nil {
			t.Fatalf("guardar el sprint: %v", err)
		}

		start := make(chan struct{})
		var wg sync.WaitGroup
		var withErr, closeErr error
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			withErr = manager.WithOpenSprint(saved.ID, func(Sprint) error {
				// Si la función corre, el cierre todavía no ocurrió.
				current, err := repo.GetByID(saved.ID)
				if err != nil {
					return err
				}
				if current.Status != StatusOpen {
					return errors.New("la función corrió después del cierre")
				}
				return nil
			})
		}()
		go func() {
			defer wg.Done()
			<-start
			_, closeErr = manager.Close(saved.ID)
		}()
		close(start)
		wg.Wait()

		if closeErr != nil {
			t.Fatalf("Close falló: %v", closeErr)
		}
		if withErr != nil && !errors.Is(withErr, ErrSprintClosed) {
			t.Fatalf("WithOpenSprint falló con un error inesperado: %v", withErr)
		}
		current, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if current.Status != StatusClosed {
			t.Fatalf("estado = %q, se esperaba CERRADO", current.Status)
		}
	}
}
