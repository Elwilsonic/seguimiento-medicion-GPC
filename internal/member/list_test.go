package member

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-17 · Consultar los integrantes de un proyecto (Issue #31)
//
// Trazabilidad: CA-17.1 a CA-17.5. La API y la interfaz (CA-17.4 por HTTP y
// CA-17.6) están en internal/httpapi.

// listStubRepository envuelve un repositorio en memoria, cuenta las lecturas
// de ListByProject y falla en ellas si se configuró listErr (RN-17.8).
type listStubRepository struct {
	*MemoryRepository
	listErr error
	lists   int
}

func (r *listStubRepository) ListByProject(projectID int) ([]Member, error) {
	r.lists++
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.MemoryRepository.ListByProject(projectID)
}

// failingProjectRepository es un repositorio de proyectos que falla al
// buscar por ID (RN-17.8), como el de los tests de History de HU-14.
type failingProjectRepository struct {
	project.Repository
	getErr error
}

func (r failingProjectRepository) GetByID(int) (project.Project, error) {
	return project.Project{}, r.getErr
}

// list consulta los integrantes de un proyecto que el test espera que se
// consulten bien.
func (e *memberEnv) list(t *testing.T, projectID int) []Member {
	t.Helper()
	got, err := e.members.List(projectID)
	if err != nil {
		t.Fatalf("List(%d) devolvió %v", projectID, err)
	}
	return got
}

// CA-17.1, RN-17.1 y RN-17.3 · Se listan los integrantes con su nombre y su
// rol, ordenados por ID.
func TestList_DevuelveLosIntegrantesOrdenadosPorIDConSuRol(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana Pérez", "AGILE_ENABLER")
	e.register(t, 1, "Bruno Díaz", "PRODUCT_BUILDER")
	e.register(t, 1, "Carla Ruiz", "PRODUCT_ARCHITECT")

	got := e.list(t, 1)

	want := []Member{
		{ID: 1, ProjectID: 1, Name: "Ana Pérez", Role: RoleAgileEnabler},
		{ID: 2, ProjectID: 1, Name: "Bruno Díaz", Role: RoleProductBuilder},
		{ID: 3, ProjectID: 1, Name: "Carla Ruiz", Role: RoleProductArchitect},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List = %+v, se esperaba %+v", got, want)
	}
}

// CA-17.3 y RN-17.5 · Solo se devuelven los integrantes del proyecto, aunque
// otro proyecto tenga uno con el mismo nombre.
func TestList_SoloDevuelveLosIntegrantesDelProyecto(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 2, "Ana", "PRODUCT_BUILDER")
	mine := e.register(t, 1, "Ana", "AGILE_ENABLER")
	e.register(t, 2, "Bruno", "PRODUCT_BUILDER")

	got := e.list(t, 1)

	if !reflect.DeepEqual(got, []Member{mine}) {
		t.Errorf("List = %+v, se esperaba %+v", got, []Member{mine})
	}
}

// CA-17.2 y RN-17.4 · Un proyecto sin integrantes devuelve una lista vacía,
// no nil.
func TestList_ProyectoSinIntegrantesDevuelveListaVacia(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 2, "Ana", "PRODUCT_BUILDER")

	got := e.list(t, 1)

	if got == nil || len(got) != 0 {
		t.Errorf("List = %#v, se esperaba una lista vacía, no nil", got)
	}
}

// RN-17.9 · Un integrante modificado con HU-16 aparece con sus datos nuevos,
// en la misma posición.
func TestList_MuestraLosDatosModificados(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")
	bruno := e.register(t, 1, "Bruno", "PRODUCT_BUILDER")
	e.update(t, 1, "Ana Pérez", "AGILE_ENABLER")

	got := e.list(t, 1)

	want := []Member{{ID: 1, ProjectID: 1, Name: "Ana Pérez", Role: RoleAgileEnabler}, bruno}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List = %+v, se esperaba %+v", got, want)
	}
}

// RN-17.2 · Se listan los integrantes de un proyecto FINALIZADO.
func TestList_AceptaUnProyectoFinalizado(t *testing.T) {
	e := newMemberEnv(t)
	ana := e.register(t, 1, "Ana", "PRODUCT_BUILDER")
	if _, err := e.projects.Update(1, "Sistema de ventas", "", string(project.StatusFinished)); err != nil {
		t.Fatalf("finalizar el proyecto: %v", err)
	}

	got := e.list(t, 1)

	if !reflect.DeepEqual(got, []Member{ana}) {
		t.Errorf("List = %+v, se esperaba %+v", got, []Member{ana})
	}
}

// CA-17.4, RN-17.2 y RN-17.8 · Un proyecto inexistente, incluidos los IDs 0
// y negativos, da ErrProjectNotFound sin leer los integrantes.
func TestList_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	repo := &listStubRepository{MemoryRepository: NewMemoryRepository()}
	members := NewManager(repo, projects)

	for _, id := range []int{99, 0, -1} {
		got, err := members.List(id)
		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("List(%d): err = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if got != nil {
			t.Errorf("List(%d) = %#v, se esperaba nil", id, got)
		}
	}
	if repo.lists != 0 {
		t.Errorf("se leyeron los integrantes %d veces, se esperaba 0", repo.lists)
	}
}

// CA-17.5 y RN-17.6 · Modificar la lista devuelta no cambia los integrantes
// guardados.
func TestList_ModificarLaListaNoCambiaLosIntegrantesGuardados(t *testing.T) {
	e := newMemberEnv(t)
	ana := e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got := e.list(t, 1)
	got[0].Name = "Cambiado"
	got[0].Role = RoleAgileEnabler

	e.expectStored(t, 1, ana)
	if again := e.list(t, 1); !reflect.DeepEqual(again, []Member{ana}) {
		t.Errorf("List = %+v, se esperaba %+v", again, []Member{ana})
	}
}

// RN-17.8 · Una falla del repositorio de proyectos se devuelve sin leer los
// integrantes.
func TestList_ErrorAlLeerElProyectoSeDevuelve(t *testing.T) {
	projects := project.NewManager(failingProjectRepository{getErr: errStorage})
	repo := &listStubRepository{MemoryRepository: NewMemoryRepository()}
	members := NewManager(repo, projects)

	got, err := members.List(1)

	if !errors.Is(err, errStorage) {
		t.Errorf("err = %v, se esperaba %v", err, errStorage)
	}
	if got != nil {
		t.Errorf("List = %#v, se esperaba nil", got)
	}
	if repo.lists != 0 {
		t.Errorf("se leyeron los integrantes %d veces, se esperaba 0", repo.lists)
	}
}

// RN-17.8 · Una falla del repositorio de integrantes se devuelve sin lista
// parcial.
func TestList_ErrorAlLeerLosIntegrantesSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	if _, err := projects.Create("Sistema de ventas", ""); err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	repo := &listStubRepository{MemoryRepository: NewMemoryRepository(), listErr: errStorage}
	members := NewManager(repo, projects)

	got, err := members.List(1)

	if !errors.Is(err, errStorage) {
		t.Errorf("err = %v, se esperaba %v", err, errStorage)
	}
	if got != nil {
		t.Errorf("List = %#v, se esperaba nil", got)
	}
	if repo.lists != 1 {
		t.Errorf("se leyeron los integrantes %d veces, se esperaba 1", repo.lists)
	}
}

// RN-17.7 · Consultar mientras se registran integrantes no compite con el
// registro: la lectura va bajo el mutex del Manager (lo detecta -race).
func TestList_ConsultasSimultaneasConRegistros(t *testing.T) {
	e := newMemberEnv(t)
	names := []string{"Ana", "Bruno", "Carla", "Diego", "Elena", "Federico", "Gabriela", "Hugo"}

	var wg sync.WaitGroup
	for _, name := range names {
		wg.Add(2)
		go func(name string) {
			defer wg.Done()
			if _, err := e.members.Register(1, name, "PRODUCT_BUILDER"); err != nil {
				t.Errorf("Register(%q) devolvió %v", name, err)
			}
		}(name)
		go func() {
			defer wg.Done()
			if _, err := e.members.List(1); err != nil {
				t.Errorf("List devolvió %v", err)
			}
		}()
	}
	wg.Wait()

	if got := e.list(t, 1); len(got) != len(names) {
		t.Errorf("List devolvió %d integrantes, se esperaban %d", len(got), len(names))
	}
}
