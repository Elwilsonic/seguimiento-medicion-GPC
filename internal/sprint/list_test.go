package sprint

import (
	"errors"
	"testing"
)

// HU-32 · Calcular Story Points planificados, completados y velocidad (Issue #50)
//
// Lecturas de sprints que usa internal/metrics: Get valida el sprint de
// SprintStoryPoints (ver "Decisión sobre entidades referenciadas" de la SDD) y
// ListByProject encuentra los sprints cerrados de la velocidad (RN-32.5).

// CA-32.7 · Get devuelve una copia del sprint guardado.
func TestGet_DevuelveElSprint(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	created, err := manager.Create(projectID, validName, validStart, validEnd, validGoal)
	if err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}

	got, err := manager.Get(created.ID)
	if err != nil {
		t.Fatalf("Get devolvió un error inesperado: %v", err)
	}
	if got != created {
		t.Errorf("Get = %+v, se esperaba %+v", got, created)
	}
	got.Name = "Cambiado"
	if saved := savedSprint(t, repo, created.ID); saved != created {
		t.Errorf("modificar la copia cambió el sprint guardado: %+v", saved)
	}
}

// CA-32.7 · Un sprint inexistente, incluidos los IDs 0 y negativos, da
// ErrSprintNotFound.
func TestGet_SprintInexistenteDevuelveErrSprintNotFound(t *testing.T) {
	manager, _, _, _ := newTestManager()

	for _, id := range []int{99, 0, -1} {
		if _, err := manager.Get(id); !errors.Is(err, ErrSprintNotFound) {
			t.Errorf("Get(%d): err = %v, se esperaba %v", id, err, ErrSprintNotFound)
		}
	}
}

// CA-32.5 · ListByProject devuelve los sprints del proyecto, abiertos y
// cerrados, ordenados por ID, sin los de otros proyectos.
func TestListByProject_DevuelveLosSprintsDelProyecto(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	ventas := createTestProject(t, projects, "Sistema de ventas")
	otro := createTestProject(t, projects, "Otro proyecto")

	first, err := manager.Create(ventas, "Sprint 1", validStart, validEnd, validGoal)
	if err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}
	closed, err := manager.Close(first.ID)
	if err != nil {
		t.Fatalf("cerrar el sprint de prueba: %v", err)
	}
	if _, err := manager.Create(otro, "Sprint 1", validStart, validEnd, validGoal); err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}
	second, err := manager.Create(ventas, "Sprint 2", validStart, validEnd, validGoal)
	if err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}

	got, err := manager.ListByProject(ventas)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}
	if len(got) != 2 || got[0] != closed || got[1] != second {
		t.Errorf("ListByProject = %+v, se esperaba [%+v %+v]", got, closed, second)
	}
}

// CA-32.4 · Un proyecto sin sprints devuelve una lista vacía, no nil.
func TestListByProject_SinSprintsDevuelveListaVacia(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	got, err := manager.ListByProject(projectID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ListByProject = %#v, se esperaba una lista vacía, no nil", got)
	}
}
