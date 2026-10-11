package project_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-01 · Crear proyecto (Issue #14)

func newManager() (*project.Manager, *project.MemoryRepository) {
	repo := project.NewMemoryRepository()
	return project.NewManager(repo), repo
}

func savedProjects(t *testing.T, repo *project.MemoryRepository) []project.Project {
	t.Helper()
	projects, err := repo.List()
	if err != nil {
		t.Fatalf("List() devolvió un error inesperado: %v", err)
	}
	return projects
}

// CA-01.1, CA-01.4
func TestCreate_ProyectoValidoQuedaConIDNombreDescripcionYEstadoInicial(t *testing.T) {
	manager, _ := newManager()

	p, err := manager.Create("Sistema de ventas", "Gestión de ventas de la empresa")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if p.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", p.ID)
	}
	if p.Name != "Sistema de ventas" {
		t.Errorf("Name = %q, se esperaba %q", p.Name, "Sistema de ventas")
	}
	if p.Description != "Gestión de ventas de la empresa" {
		t.Errorf("Description = %q, se esperaba %q", p.Description, "Gestión de ventas de la empresa")
	}
	if p.Status != project.StatusPlanned {
		t.Errorf("Status = %q, se esperaba %q", p.Status, project.StatusPlanned)
	}
}

// CA-01.1
func TestCreate_IDsCorrelativosEmpiezanEnUno(t *testing.T) {
	manager, _ := newManager()

	for want := 1; want <= 3; want++ {
		p, err := manager.Create(fmt.Sprintf("Proyecto %d", want), "")
		if err != nil {
			t.Fatalf("Create() devolvió un error inesperado: %v", err)
		}
		if p.ID != want {
			t.Errorf("ID = %d, se esperaba %d", p.ID, want)
		}
	}
}

// CA-01.1
func TestCreate_ProyectoQuedaGuardado(t *testing.T) {
	manager, repo := newManager()

	created, err := manager.Create("Proyecto A", "Descripción")
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	projects := savedProjects(t, repo)
	if len(projects) != 1 {
		t.Fatalf("hay %d proyectos guardados, se esperaba 1", len(projects))
	}
	if projects[0] != created {
		t.Errorf("proyecto guardado = %+v, se esperaba %+v", projects[0], created)
	}
}

// CA-01.2
func TestCreate_NombreVacioDevuelveErrEmptyName(t *testing.T) {
	cases := map[string]string{
		"vacío":              "",
		"solo espacios":      "   ",
		"solo tabs y saltos": "\t\n ",
	}
	for caseName, name := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo := newManager()

			_, err := manager.Create(name, "Sin nombre")

			if !errors.Is(err, project.ErrEmptyName) {
				t.Errorf("error = %v, se esperaba ErrEmptyName", err)
			}
			if n := len(savedProjects(t, repo)); n != 0 {
				t.Errorf("hay %d proyectos guardados, se esperaba 0", n)
			}
		})
	}
}

// CA-01.3
func TestCreate_NombreRepetidoDevuelveErrDuplicateName(t *testing.T) {
	cases := map[string]string{
		"mismo nombre":             "Proyecto A",
		"otras mayúsculas":         "PROYECTO A",
		"minúsculas":               "proyecto a",
		"espacios en los extremos": "  Proyecto A  ",
		"mayúsculas y espacios":    " pRoYeCtO a ",
	}
	for caseName, name := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo := newManager()
			if _, err := manager.Create("Proyecto A", "Original"); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			_, err := manager.Create(name, "Repetido")

			if !errors.Is(err, project.ErrDuplicateName) {
				t.Errorf("error = %v, se esperaba ErrDuplicateName", err)
			}
			if n := len(savedProjects(t, repo)); n != 1 {
				t.Errorf("hay %d proyectos guardados, se esperaba 1", n)
			}
		})
	}
}

// CA-01.3
func TestCreate_NombreRepetidoSinDistinguirMayusculasConLetrasAcentuadas(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Gestión Ñandú", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	_, err := manager.Create("GESTIÓN ÑANDÚ", "")

	if !errors.Is(err, project.ErrDuplicateName) {
		t.Errorf("error = %v, se esperaba ErrDuplicateName", err)
	}
}

// CA-01.3 (caso límite)
func TestCreate_NombresQueDifierenEnEspaciosInternosSonDistintos(t *testing.T) {
	manager, repo := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Create("Proyecto  A", "")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if p.Name != "Proyecto  A" {
		t.Errorf("Name = %q, se esperaba %q", p.Name, "Proyecto  A")
	}
	if n := len(savedProjects(t, repo)); n != 2 {
		t.Errorf("hay %d proyectos guardados, se esperaba 2", n)
	}
}

// CA-01.2 (prioridad de errores): si el nombre está vacío no se controlan repetidos.
func TestCreate_NombreVacioSeInformaAntesQueRepetido(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	_, err := manager.Create("  ", "")

	if !errors.Is(err, project.ErrEmptyName) {
		t.Errorf("error = %v, se esperaba ErrEmptyName", err)
	}
}

// CA-01.5
func TestCreate_NombreSeGuardaSinEspaciosEnLosExtremos(t *testing.T) {
	manager, repo := newManager()

	p, err := manager.Create("   Proyecto A \t", "")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if p.Name != "Proyecto A" {
		t.Errorf("Name devuelto = %q, se esperaba %q", p.Name, "Proyecto A")
	}
	if saved := savedProjects(t, repo)[0]; saved.Name != "Proyecto A" {
		t.Errorf("Name guardado = %q, se esperaba %q", saved.Name, "Proyecto A")
	}
}

// CA-01.5
func TestCreate_DescripcionSeGuardaTalComoLlega(t *testing.T) {
	cases := map[string]string{
		"vacía":                    "",
		"con espacios en extremos": "  Descripción con espacios  ",
	}
	for caseName, description := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo := newManager()

			p, err := manager.Create("Proyecto A", description)

			if err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}
			if p.Description != description {
				t.Errorf("Description devuelta = %q, se esperaba %q", p.Description, description)
			}
			if saved := savedProjects(t, repo)[0]; saved.Description != description {
				t.Errorf("Description guardada = %q, se esperaba %q", saved.Description, description)
			}
		})
	}
}

// CA-01.6
func TestCreate_ModificarElProyectoDevueltoNoAlteraElGuardado(t *testing.T) {
	manager, repo := newManager()
	p, err := manager.Create("Proyecto A", "Original")
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p.Name = "Otro nombre"
	p.Description = "Otra descripción"
	p.Status = project.StatusFinished

	saved := savedProjects(t, repo)[0]
	if saved.Name != "Proyecto A" || saved.Description != "Original" || saved.Status != project.StatusPlanned {
		t.Errorf("el proyecto guardado cambió: %+v", saved)
	}
}

// CA-01.6
func TestMemoryRepository_ListDevuelveCopias(t *testing.T) {
	manager, repo := newManager()
	if _, err := manager.Create("Proyecto A", "Original"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	projects := savedProjects(t, repo)
	projects[0].Name = "Otro nombre"

	if saved := savedProjects(t, repo)[0]; saved.Name != "Proyecto A" {
		t.Errorf("Name guardado = %q, se esperaba %q", saved.Name, "Proyecto A")
	}
}

// CA-01.7
func TestCreate_CreacionesSimultaneasConElMismoNombreCreanUnSoloProyecto(t *testing.T) {
	const attempts = 50
	manager, repo := newManager()

	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = manager.Create("Proyecto concurrente", "")
		}()
	}
	wg.Wait()

	duplicates := 0
	for _, err := range errs {
		if errors.Is(err, project.ErrDuplicateName) {
			duplicates++
		}
	}
	if n := len(savedProjects(t, repo)); n != 1 {
		t.Errorf("hay %d proyectos guardados, se esperaba 1", n)
	}
	if duplicates != attempts-1 {
		t.Errorf("%d creaciones fallaron por nombre repetido, se esperaban %d", duplicates, attempts-1)
	}
}

// CA-01.7
func TestCreate_CreacionesSimultaneasConNombresDistintosTienenIDsDistintos(t *testing.T) {
	const attempts = 50
	manager, repo := newManager()

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
	wg.Wait()

	projects := savedProjects(t, repo)
	if len(projects) != attempts {
		t.Fatalf("hay %d proyectos guardados, se esperaban %d", len(projects), attempts)
	}
	ids := make(map[int]bool)
	for _, p := range projects {
		if ids[p.ID] {
			t.Errorf("el ID %d está repetido", p.ID)
		}
		ids[p.ID] = true
	}
}

// RN-01.6
func TestStatus_EstadosDelProyecto(t *testing.T) {
	cases := map[project.Status]string{
		project.StatusPlanned:    "PLANIFICADO",
		project.StatusInProgress: "EN_CURSO",
		project.StatusFinished:   "FINALIZADO",
	}
	for status, want := range cases {
		if string(status) != want {
			t.Errorf("estado = %q, se esperaba %q", status, want)
		}
	}
}
