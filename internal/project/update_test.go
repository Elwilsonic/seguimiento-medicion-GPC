package project_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-02 · Modificar datos del proyecto (Issue #15)

// CA-02.1
func TestUpdate_NombreDescripcionYEstadoSeActualizan(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Sistema de ventas", "Gestión de ventas"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Update(1, "Sistema de facturación", "Facturación electrónica", "EN_CURSO")

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if p.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", p.ID)
	}
	if p.Name != "Sistema de facturación" {
		t.Errorf("Name = %q, se esperaba %q", p.Name, "Sistema de facturación")
	}
	if p.Description != "Facturación electrónica" {
		t.Errorf("Description = %q, se esperaba %q", p.Description, "Facturación electrónica")
	}
	if p.Status != project.StatusInProgress {
		t.Errorf("Status = %q, se esperaba %q", p.Status, project.StatusInProgress)
	}
}

// CA-02.1
func TestUpdate_CambiarSoloElEstadoMantieneNombreYDescripcion(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", "Original"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Update(1, "Proyecto A", "Original", "EN_CURSO")

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if p.Name != "Proyecto A" {
		t.Errorf("Name = %q, se esperaba %q", p.Name, "Proyecto A")
	}
	if p.Description != "Original" {
		t.Errorf("Description = %q, se esperaba %q", p.Description, "Original")
	}
	if p.Status != project.StatusInProgress {
		t.Errorf("Status = %q, se esperaba %q", p.Status, project.StatusInProgress)
	}
}

// CA-02.2 (no hay transiciones prohibidas)
func TestUpdate_TodosLosEstadosSonAlcanzables(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	transitions := []project.Status{project.StatusInProgress, project.StatusFinished, project.StatusPlanned}
	for _, want := range transitions {
		p, err := manager.Update(1, "Proyecto A", "", string(want))
		if err != nil {
			t.Fatalf("Update(%q) devolvió un error inesperado: %v", want, err)
		}
		if p.Status != want {
			t.Errorf("Update(%q): Status = %q", want, p.Status)
		}
	}
}

// CA-02.2
func TestUpdate_EstadoSeAceptaSinMayusculasNiEspaciosEnLosExtremos(t *testing.T) {
	cases := map[string]project.Status{
		"en_curso":       project.StatusInProgress,
		" En_Curso ":     project.StatusInProgress,
		"finalizado":     project.StatusFinished,
		" planificado  ": project.StatusPlanned,
	}
	for caseName, want := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, _ := newManager()
			if _, err := manager.Create("Proyecto A", ""); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			p, err := manager.Update(1, "Proyecto A", "", caseName)

			if err != nil {
				t.Fatalf("Update(%q) devolvió un error inesperado: %v", caseName, err)
			}
			if p.Status != want {
				t.Errorf("Status = %q, se esperaba %q", p.Status, want)
			}
		})
	}
}

// CA-02.1, RN-02.2
func TestUpdate_NombreNuevoSeGuardaSinEspaciosEnLosExtremos(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Update(1, "  Proyecto B  ", "", "PLANIFICADO")

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if p.Name != "Proyecto B" {
		t.Errorf("Name = %q, se esperaba %q", p.Name, "Proyecto B")
	}
	saved, err := manager.Get(1)
	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	if saved.Name != "Proyecto B" {
		t.Errorf("nombre guardado = %q, se esperaba %q", saved.Name, "Proyecto B")
	}
}

// CA-02.3
func TestUpdate_NombreVacioOSoloEspaciosDevuelveErrEmptyNameYNoCambia(t *testing.T) {
	cases := map[string]string{
		"vacío":              "",
		"solo espacios":      "   ",
		"solo tabs y saltos": "\t\n ",
	}
	for caseName, name := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, _ := newManager()
			if _, err := manager.Create("Proyecto A", "Original"); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			_, err := manager.Update(1, name, "Nueva", "EN_CURSO")

			if !errors.Is(err, project.ErrEmptyName) {
				t.Errorf("error = %v, se esperaba ErrEmptyName", err)
			}
			again, err := manager.Get(1)
			if err != nil {
				t.Fatalf("Get() devolvió un error inesperado: %v", err)
			}
			if again.Name != "Proyecto A" || again.Description != "Original" || again.Status != project.StatusPlanned {
				t.Errorf("el proyecto guardado cambió: %+v", again)
			}
		})
	}
}

// CA-02.2, CA-02.3
func TestUpdate_EstadoInvalidoDevuelveErrInvalidStatusYNoCambia(t *testing.T) {
	cases := map[string]string{
		"estado que no existe": "PAUSADO",
		"vacío":                "",
	}
	for caseName, status := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, _ := newManager()
			if _, err := manager.Create("Proyecto A", "Original"); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			_, err := manager.Update(1, "Proyecto A", "Nueva", status)

			if !errors.Is(err, project.ErrInvalidStatus) {
				t.Errorf("error = %v, se esperaba ErrInvalidStatus", err)
			}
			again, err := manager.Get(1)
			if err != nil {
				t.Fatalf("Get() devolvió un error inesperado: %v", err)
			}
			if again.Name != "Proyecto A" || again.Description != "Original" || again.Status != project.StatusPlanned {
				t.Errorf("el proyecto guardado cambió: %+v", again)
			}
		})
	}
}

// CA-02.5
func TestUpdate_NombreRepetidoConOtroProyectoDevuelveErrDuplicateNameYNoCambia(t *testing.T) {
	cases := map[string]string{
		"mismo nombre":             "Proyecto A",
		"otras mayúsculas":         "PROYECTO A",
		"espacios en los extremos": " Proyecto A ",
	}
	for caseName, name := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, _ := newManager()
			if _, err := manager.Create("Proyecto A", ""); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}
			if _, err := manager.Create("Proyecto B", ""); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			_, err := manager.Update(2, name, "", "PLANIFICADO")

			if !errors.Is(err, project.ErrDuplicateName) {
				t.Errorf("error = %v, se esperaba ErrDuplicateName", err)
			}
			a, err := manager.Get(1)
			if err != nil {
				t.Fatalf("Get() devolvió un error inesperado: %v", err)
			}
			b, err := manager.Get(2)
			if err != nil {
				t.Fatalf("Get() devolvió un error inesperado: %v", err)
			}
			if a.Name != "Proyecto A" || b.Name != "Proyecto B" {
				t.Errorf("algún proyecto cambió: %+v, %+v", a, b)
			}
		})
	}
}

// CA-02.5 (el propio proyecto se excluye del control de duplicados)
func TestUpdate_MismoNombreDelPropioProyectoEsValido(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Update(1, " PROYECTO a ", "Actualizado", "PLANIFICADO")

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if p.Name != "PROYECTO a" {
		t.Errorf("Name = %q, se esperaba %q", p.Name, "PROYECTO a")
	}
}

// CA-02.4
func TestUpdate_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	cases := map[string]int{
		"id mayor al último": 999,
		"cero":               0,
		"negativo":           -1,
	}
	for caseName, id := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, _ := newManager()
			if _, err := manager.Create("Proyecto A", ""); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			_, err := manager.Update(id, "Proyecto B", "", "PLANIFICADO")

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
			}
		})
	}
	t.Run("sin proyectos creados", func(t *testing.T) {
		manager, _ := newManager()

		_, err := manager.Update(1, "Proyecto A", "", "PLANIFICADO")

		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
		}
	})
}

// RN-02.5
func TestUpdate_DescripcionNuevaVaciaSeGuarda(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", "Original"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Update(1, "Proyecto A", "", "PLANIFICADO")

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if p.Description != "" {
		t.Errorf("Description = %q, se esperaba vacía", p.Description)
	}
	saved, err := manager.Get(1)
	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	if saved.Description != "" {
		t.Errorf("descripción guardada = %q, se esperaba vacía", saved.Description)
	}
}

// CA-02.6
func TestUpdate_ModificarElResultadoDevueltoNoAlteraElGuardado(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", "Original"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.Update(1, "Proyecto B", "Nueva", "EN_CURSO")
	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	p.Name = "Otro nombre"
	p.Description = "Otra descripción"
	p.Status = project.StatusFinished

	again, err := manager.Get(1)
	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	if again.Name != "Proyecto B" || again.Description != "Nueva" || again.Status != project.StatusInProgress {
		t.Errorf("el proyecto guardado cambió: %+v", again)
	}
}

// CA-02.8: si el repositorio falla al buscar, listar o actualizar, se
// devuelve su error. Con getErr en nil, el GetByID del repositorio falso
// devuelve un proyecto vacío sin error, para poder llegar a los puntos de
// falla posteriores (listar y actualizar).
func TestUpdate_ErrorDelRepositorioSeDevuelveYElProyectoNoCambia(t *testing.T) {
	cases := map[string]failingRepository{
		"al buscar":     {getErr: errRepository},
		"al listar":     {listErr: errRepository},
		"al actualizar": {updateErr: errRepository},
	}
	for caseName, repo := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager := project.NewManager(repo)

			p, err := manager.Update(1, "Proyecto A", "", "PLANIFICADO")

			if !errors.Is(err, errRepository) {
				t.Errorf("error = %v, se esperaba el error del repositorio", err)
			}
			if p != (project.Project{}) {
				t.Errorf("proyecto devuelto = %+v, se esperaba un proyecto vacío", p)
			}
		})
	}
}

// CA-02.7: modificaciones y creaciones simultáneas sin accesos concurrentes
// sin protección. Este test se verifica corriendo con `go test -race` (lo corre el CI).
func TestUpdate_CreacionesYModificacionesSimultaneasSinAccesosSinProteger(t *testing.T) {
	const existing = 10
	const creates = 50
	manager, repo := newManager()
	for i := 0; i < existing; i++ {
		if _, err := manager.Create(fmt.Sprintf("Proyecto %d", i), ""); err != nil {
			t.Fatalf("Create() devolvió un error inesperado: %v", err)
		}
	}

	opErrs := make([]error, existing+creates)
	var wg sync.WaitGroup
	for i := 0; i < existing; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, opErrs[i] = manager.Update(i+1, fmt.Sprintf("Proyecto %d actualizado", i), "", "EN_CURSO")
		}()
	}
	for i := 0; i < creates; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, opErrs[existing+i] = manager.Create(fmt.Sprintf("Proyecto %d", i+existing), "")
		}()
	}
	wg.Wait()

	for _, err := range opErrs {
		if err != nil {
			t.Errorf("operación simultánea con error inesperado: %v", err)
		}
	}

	if projects := savedProjects(t, repo); len(projects) != existing+creates {
		t.Fatalf("hay %d proyectos guardados, se esperaban %d", len(projects), existing+creates)
	}
	for id := 1; id <= existing; id++ {
		p, err := manager.Get(id)
		if err != nil {
			t.Fatalf("Get(%d) devolvió un error inesperado: %v", id, err)
		}
		if want := fmt.Sprintf("Proyecto %d actualizado", id-1); p.Name != want {
			t.Errorf("Get(%d): Name = %q, se esperaba %q", id, p.Name, want)
		}
		if p.Status != project.StatusInProgress {
			t.Errorf("Get(%d): Status = %q, se esperaba %q", id, p.Status, project.StatusInProgress)
		}
	}
}

// Contrato del repositorio (RN-02.6): Update sobre un repositorio donde no
// existe ese ID devuelve ErrProjectNotFound. La rama no es alcanzable a
// través del Manager (el GetByID previo ya encontró el proyecto), así que se
// prueba el repositorio directamente.
func TestMemoryRepository_UpdateProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	repo := project.NewMemoryRepository()

	_, err := repo.Update(project.Project{ID: 1, Name: "Proyecto A", Status: project.StatusPlanned})

	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
	}
}

// HU-18, CA-18.6 (RN-18.6): modificar el proyecto con HU-02 conserva la fecha
// de inicio registrada, porque Update solo reemplaza nombre, descripción y
// estado.
func TestUpdate_ConservaLaFechaDeInicioRegistrada(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	mustSetStartDate(t, manager, 1, day(t, "2026-10-23"))

	p, err := manager.Update(1, "Proyecto B", "Nueva", "EN_CURSO")

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if !p.StartDate.Equal(day(t, "2026-10-23")) {
		t.Errorf("fecha devuelta = %v, se esperaba 2026-10-23", p.StartDate)
	}
	if got := mustGet(t, manager, 1).StartDate; !got.Equal(day(t, "2026-10-23")) {
		t.Errorf("fecha guardada = %v, se esperaba 2026-10-23", got)
	}
}
