package sprint

import (
	"errors"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-14 · Consultar sprints anteriores (Issue #28)

// historyRepository envuelve un Repository, cuenta las lecturas de
// ListByProject y falla en ellas si se configuró listErr (RN-14.7).
type historyRepository struct {
	Repository
	listErr error
	lists   int
}

func (r *historyRepository) ListByProject(projectID int) ([]Sprint, error) {
	r.lists++
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.Repository.ListByProject(projectID)
}

// failingProjectRepository es un repositorio de proyectos que falla al
// buscar por ID (RN-14.7).
type failingProjectRepository struct {
	project.Repository
	getErr error
}

func (r failingProjectRepository) GetByID(id int) (project.Project, error) {
	return project.Project{}, r.getErr
}

// CA-14.1, RN-14.1 y RN-14.3 · History devuelve los sprints cerrados y el
// abierto, ordenados por ID, con todos sus datos.
func TestHistory_DevuelveLosSprintsAbiertosYCerradosOrdenadosPorID(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	first, err := manager.Create(projectID, "Sprint 1", validStart, validEnd, "Backlog inicial")
	if err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}
	firstClosed, err := manager.Close(first.ID)
	if err != nil {
		t.Fatalf("cerrar el sprint de prueba: %v", err)
	}
	second, err := manager.Create(projectID, "Sprint 2", validStart, validEnd, "Métricas")
	if err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}

	got, err := manager.History(projectID)

	if err != nil {
		t.Fatalf("History devolvió un error inesperado: %v", err)
	}
	if len(got) != 2 || got[0] != firstClosed || got[1] != second {
		t.Errorf("History = %+v, se esperaba [%+v %+v]", got, firstClosed, second)
	}
	if got[0].Status != StatusClosed || got[1].Status != StatusOpen {
		t.Errorf("estados = %q y %q, se esperaba %q y %q", got[0].Status, got[1].Status, StatusClosed, StatusOpen)
	}
}

// CA-14.1 · Solo se devuelven los sprints del proyecto consultado.
func TestHistory_SoloDevuelveLosSprintsDelProyecto(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	ventas := createTestProject(t, projects, "Sistema de ventas")
	otro := createTestProject(t, projects, "Otro proyecto")
	if _, err := manager.Create(otro, "Sprint A", validStart, validEnd, validGoal); err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}
	mine, err := manager.Create(ventas, "Sprint 1", validStart, validEnd, validGoal)
	if err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}

	got, err := manager.History(ventas)

	if err != nil {
		t.Fatalf("History devolvió un error inesperado: %v", err)
	}
	if len(got) != 1 || got[0] != mine {
		t.Errorf("History = %+v, se esperaba [%+v]", got, mine)
	}
}

// CA-14.2 y RN-14.4 · Un proyecto sin sprints devuelve una lista vacía, no nil.
func TestHistory_ProyectoSinSprintsDevuelveListaVacia(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	got, err := manager.History(projectID)

	if err != nil {
		t.Fatalf("History devolvió un error inesperado: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("History = %#v, se esperaba una lista vacía, no nil", got)
	}
}

// CA-14.4, RN-14.2 y RN-14.7 · Un proyecto inexistente, incluidos los IDs 0
// y negativos, da ErrProjectNotFound sin leer los sprints.
func TestHistory_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	repo := &historyRepository{Repository: NewMemoryRepository()}
	manager := NewManager(repo, projects, items)

	for _, id := range []int{99, 0, -1} {
		got, err := manager.History(id)
		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("History(%d): err = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if got != nil {
			t.Errorf("History(%d) = %#v, se esperaba nil", id, got)
		}
	}
	if repo.lists != 0 {
		t.Errorf("se leyeron los sprints %d veces, se esperaba 0", repo.lists)
	}
}

// CA-14.5 y RN-14.5 · Modificar la lista devuelta no cambia los sprints
// guardados.
func TestHistory_ModificarLaListaNoCambiaLosSprintsGuardados(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	created, err := manager.Create(projectID, validName, validStart, validEnd, validGoal)
	if err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}

	got, err := manager.History(projectID)
	if err != nil {
		t.Fatalf("History devolvió un error inesperado: %v", err)
	}
	got[0].Name = "Cambiado"
	got[0].Status = StatusClosed

	if saved := savedSprint(t, repo, created.ID); saved != created {
		t.Errorf("modificar la lista cambió el sprint guardado: %+v", saved)
	}
	again, err := manager.History(projectID)
	if err != nil {
		t.Fatalf("History devolvió un error inesperado: %v", err)
	}
	if len(again) != 1 || again[0] != created {
		t.Errorf("History = %+v, se esperaba [%+v]", again, created)
	}
}

// RN-14.7 · Una falla del repositorio de proyectos se devuelve sin leer los
// sprints.
func TestHistory_ErrorAlLeerElProyectoSeDevuelve(t *testing.T) {
	projects := project.NewManager(failingProjectRepository{getErr: errRepository})
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	repo := &historyRepository{Repository: NewMemoryRepository()}
	manager := NewManager(repo, projects, items)

	got, err := manager.History(1)

	if !errors.Is(err, errRepository) {
		t.Errorf("err = %v, se esperaba %v", err, errRepository)
	}
	if got != nil {
		t.Errorf("History = %#v, se esperaba nil", got)
	}
	if repo.lists != 0 {
		t.Errorf("se leyeron los sprints %d veces, se esperaba 0", repo.lists)
	}
}

// RN-14.7 · Una falla del repositorio de sprints se devuelve sin lista
// parcial.
func TestHistory_ErrorAlLeerLosSprintsSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	projectID := createTestProject(t, projects, "Sistema de ventas")
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	repo := &historyRepository{Repository: NewMemoryRepository(), listErr: errRepository}
	manager := NewManager(repo, projects, items)

	got, err := manager.History(projectID)

	if !errors.Is(err, errRepository) {
		t.Errorf("err = %v, se esperaba %v", err, errRepository)
	}
	if got != nil {
		t.Errorf("History = %#v, se esperaba nil", got)
	}
	if repo.lists != 1 {
		t.Errorf("se leyeron los sprints %d veces, se esperaba 1", repo.lists)
	}
}
