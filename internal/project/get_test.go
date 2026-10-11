package project_test

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-03 · Consultar estado del proyecto (Issue #16)

// fixedStatusRepository es un Repository de prueba que guarda los proyectos
// con un estado fijo, para consultar estados distintos del inicial sin HU-02.
type fixedStatusRepository struct {
	status   project.Status
	projects []project.Project
}

func (r *fixedStatusRepository) Save(p project.Project) (project.Project, error) {
	p.ID = len(r.projects) + 1
	p.Status = r.status
	r.projects = append(r.projects, p)
	return p, nil
}

func (r *fixedStatusRepository) List() ([]project.Project, error) {
	return slices.Clone(r.projects), nil
}

func (r *fixedStatusRepository) GetByID(id int) (project.Project, error) {
	for _, p := range r.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return project.Project{}, project.ErrProjectNotFound
}

// CA-03.1
func TestGet_ProyectoExistenteDevuelveNombreDescripcionYEstado(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Sistema de ventas", "Gestión de ventas"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Get(1)

	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	if p.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", p.ID)
	}
	if p.Name != "Sistema de ventas" {
		t.Errorf("Name = %q, se esperaba %q", p.Name, "Sistema de ventas")
	}
	if p.Description != "Gestión de ventas" {
		t.Errorf("Description = %q, se esperaba %q", p.Description, "Gestión de ventas")
	}
	if p.Status != project.StatusPlanned {
		t.Errorf("Status = %q, se esperaba %q", p.Status, project.StatusPlanned)
	}
}

// CA-03.1
func TestGet_ProyectoQueNoEsElPrimero(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if _, err := manager.Create("Proyecto B", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Get(2)

	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	if p.Name != "Proyecto B" {
		t.Errorf("Name = %q, se esperaba %q", p.Name, "Proyecto B")
	}
}

// CA-03.1
func TestGet_ProyectoConDescripcionVacia(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Get(1)

	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	if p.Description != "" {
		t.Errorf("Description = %q, se esperaba vacía", p.Description)
	}
}

// CA-03.1 (con un repositorio falso de prueba, porque HU-02 aún no cambia estados)
func TestGet_ProyectoEnEstadoDistintoDelInicial(t *testing.T) {
	manager := project.NewManager(&fixedStatusRepository{status: project.StatusInProgress})
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Get(1)

	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	if p.Name != "Proyecto A" {
		t.Errorf("Name = %q, se esperaba %q", p.Name, "Proyecto A")
	}
	if p.Status != project.StatusInProgress {
		t.Errorf("Status = %q, se esperaba %q", p.Status, project.StatusInProgress)
	}
}

// CA-03.2
func TestGet_IDMayorAlUltimoAsignadoDevuelveErrProjectNotFound(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	_, err := manager.Get(999)

	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
	}
}

// CA-03.2
func TestGet_SinProyectosCreadosDevuelveErrProjectNotFound(t *testing.T) {
	manager, _ := newManager()

	_, err := manager.Get(1)

	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
	}
}

// CA-03.2
func TestGet_IDCeroYNegativoDevuelvenErrProjectNotFound(t *testing.T) {
	cases := map[string]int{
		"cero":     0,
		"negativo": -1,
	}
	for caseName, id := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, _ := newManager()
			if _, err := manager.Create("Proyecto A", ""); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			_, err := manager.Get(id)

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
			}
		})
	}
}

// CA-03.3
func TestGet_ModificarElResultadoDevueltoNoAlteraElGuardado(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", "Original"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Get(1)
	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	p.Name = "Otro nombre"
	p.Description = "Otra descripción"
	p.Status = project.StatusFinished

	again, err := manager.Get(1)
	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	if again.Name != "Proyecto A" || again.Description != "Original" || again.Status != project.StatusPlanned {
		t.Errorf("el proyecto guardado cambió: %+v", again)
	}
}

// Condición de error: si el repositorio falla al buscar, se devuelve su error.
func TestGet_ErrorDelRepositorioSeDevuelve(t *testing.T) {
	manager := project.NewManager(failingRepository{getErr: errRepository})

	p, err := manager.Get(1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if p != (project.Project{}) {
		t.Errorf("proyecto devuelto = %+v, se esperaba un proyecto vacío", p)
	}
}

// CA-03.4: creaciones y consultas simultáneas sin accesos concurrentes sin
// protección. Este test se verifica corriendo con `go test -race` (lo corre el CI).
func TestGet_CreacionesYConsultasSimultaneasSinAccesosSinProteger(t *testing.T) {
	const attempts = 50
	manager, repo := newManager()

	queryErrs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := manager.Create(fmt.Sprintf("Proyecto %d", i), ""); err != nil {
				t.Errorf("Create() devolvió un error inesperado: %v", err)
			}
		}()
	}
	for id := 1; id <= attempts; id++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, queryErrs[id-1] = manager.Get(id)
		}()
	}
	wg.Wait()

	for _, err := range queryErrs {
		if err != nil && !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("consulta con error inesperado: %v", err)
		}
	}

	projects := savedProjects(t, repo)
	if len(projects) != attempts {
		t.Fatalf("hay %d proyectos guardados, se esperaban %d", len(projects), attempts)
	}
	// Los IDs se asignan en cualquier orden con creaciones simultáneas, así
	// que se verifica que cada ID devuelva un proyecto con ese ID y que el
	// conjunto de nombres sea exactamente "Proyecto 0" a "Proyecto 49".
	names := make(map[string]bool, attempts)
	for id := 1; id <= attempts; id++ {
		p, err := manager.Get(id)
		if err != nil {
			t.Fatalf("Get(%d) devolvió un error inesperado: %v", id, err)
		}
		if p.ID != id {
			t.Errorf("Get(%d) devolvió el proyecto con ID %d", id, p.ID)
		}
		names[p.Name] = true
	}
	for i := range attempts {
		if !names[fmt.Sprintf("Proyecto %d", i)] {
			t.Errorf("falta el proyecto con nombre %q", fmt.Sprintf("Proyecto %d", i))
		}
	}
}
