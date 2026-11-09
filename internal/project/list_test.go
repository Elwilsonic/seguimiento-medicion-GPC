package project_test

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-50 · Listar los proyectos (Issue #183)

// mustCreate crea un proyecto y corta el test si falla.
func mustCreate(t *testing.T, m *project.Manager, name, description string) project.Project {
	t.Helper()
	p, err := m.Create(name, description)
	if err != nil {
		t.Fatalf("Create(%q) devolvió un error inesperado: %v", name, err)
	}
	return p
}

// mustList consulta la lista y corta el test si falla.
func mustList(t *testing.T, m *project.Manager) []project.Project {
	t.Helper()
	got, err := m.List()
	if err != nil {
		t.Fatalf("List() devolvió un error inesperado: %v", err)
	}
	return got
}

// HU-50 · CA-50.1 · Verifica que la lista trae todos los proyectos, cada uno
// con su ID, nombre, descripción, estado y fechas.
func TestList_DevuelveTodosLosProyectosConSusDatos(t *testing.T) {
	manager, _ := newManager()
	ventas := mustCreate(t, manager, "Sistema de ventas", "Ventas online")
	inventario := mustCreate(t, manager, "Inventario", "")
	if _, err := manager.Update(inventario.ID, "Inventario", "", "EN_CURSO"); err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	if _, err := manager.SetStartDate(inventario.ID, start); err != nil {
		t.Fatalf("SetStartDate() devolvió un error inesperado: %v", err)
	}
	if _, err := manager.SetEndDate(inventario.ID, end); err != nil {
		t.Fatalf("SetEndDate() devolvió un error inesperado: %v", err)
	}

	got := mustList(t, manager)

	want := []project.Project{
		{ID: inventario.ID, Name: "Inventario", Description: "", Status: project.StatusInProgress, StartDate: start, EndDate: end},
		{ID: ventas.ID, Name: "Sistema de ventas", Description: "Ventas online", Status: project.StatusPlanned},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %+v, se esperaba %+v", got, want)
	}
}

// HU-50 · CA-50.1 · Verifica que con un único proyecto la lista tiene
// exactamente ese proyecto.
func TestList_UnUnicoProyecto(t *testing.T) {
	manager, _ := newManager()
	agenda := mustCreate(t, manager, "Agenda", "Turnos")

	got := mustList(t, manager)

	if want := []project.Project{agenda}; !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %+v, se esperaba %+v", got, want)
	}
}

// HU-50 · CA-50.2 · Verifica que sin proyectos la lista está vacía y no es
// nil, aunque el repositorio en memoria devuelva nil.
func TestList_SinProyectosDevuelveListaVaciaNoNil(t *testing.T) {
	manager, _ := newManager()

	got, err := manager.List()

	if err != nil {
		t.Fatalf("List() devolvió un error inesperado: %v", err)
	}
	if got == nil {
		t.Fatal("List() = nil, se esperaba una lista vacía")
	}
	if len(got) != 0 {
		t.Errorf("List() tiene %d proyectos, se esperaba 0", len(got))
	}
}

// HU-50 · CA-50.4 · Verifica que modificar los proyectos de la lista devuelta
// no altera los proyectos guardados.
func TestList_ModificarLaListaDevueltaNoAlteraLosGuardados(t *testing.T) {
	manager, _ := newManager()
	agenda := mustCreate(t, manager, "Agenda", "Turnos")

	got := mustList(t, manager)
	if len(got) != 1 {
		t.Fatalf("List() tiene %d proyectos, se esperaba 1", len(got))
	}
	got[0].Name = "Otro"
	got[0].Description = "Otra"
	got[0].Status = project.StatusFinished

	saved, err := manager.Get(agenda.ID)
	if err != nil {
		t.Fatalf("Get() devolvió un error inesperado: %v", err)
	}
	if saved != agenda {
		t.Errorf("Get() = %+v, se esperaba %+v", saved, agenda)
	}
	if again := mustList(t, manager); !reflect.DeepEqual(again, []project.Project{agenda}) {
		t.Errorf("la segunda List() = %+v, se esperaba %+v", again, []project.Project{agenda})
	}
}

// HU-50 · CA-50.4 · Verifica que consultar la lista no modifica ningún
// proyecto: los datos guardados son los mismos antes y después.
func TestList_NoModificaLosProyectos(t *testing.T) {
	manager, repo := newManager()
	mustCreate(t, manager, "Sistema de ventas", "Ventas online")
	mustCreate(t, manager, "Agenda", "")
	before := savedProjects(t, repo)

	mustList(t, manager)
	mustList(t, manager)

	if after := savedProjects(t, repo); !reflect.DeepEqual(after, before) {
		t.Errorf("proyectos guardados después de listar = %+v, se esperaba %+v", after, before)
	}
}

// HU-50 · CA-50.5 · Verifica que si el repositorio falla al listar se
// devuelve ese error y ninguna lista.
func TestList_ErrorDelRepositorioDevuelveElErrorSinLista(t *testing.T) {
	errStorage := errors.New("falla del almacenamiento")
	manager := project.NewManager(failingRepository{listErr: errStorage})

	got, err := manager.List()

	if !errors.Is(err, errStorage) {
		t.Errorf("List() error = %v, se esperaba %v", err, errStorage)
	}
	if got != nil {
		t.Errorf("List() = %+v, se esperaba nil", got)
	}
}

// HU-50 · CA-50.6 · Verifica que la lista se ordena por nombre sin distinguir
// mayúsculas, y no por ID ni por orden de creación.
func TestList_OrdenaPorNombreSinDistinguirMayusculas(t *testing.T) {
	manager, _ := newManager()
	for _, name := range []string{"Sistema de ventas", "inventario", "Agenda", "zeta", "Beta"} {
		mustCreate(t, manager, name, "")
	}

	got := mustList(t, manager)

	want := []string{"Agenda", "Beta", "inventario", "Sistema de ventas", "zeta"}
	if names := namesOf(got); !reflect.DeepEqual(names, want) {
		t.Errorf("nombres = %q, se esperaba %q", names, want)
	}
}

// HU-50 · CA-50.7 · Verifica que listar mientras se crean proyectos en
// simultáneo no tiene condiciones de carrera (go test -race) y que al final
// la lista tiene todos los proyectos creados.
func TestList_ConcurrenteConCreateNoTieneCondicionesDeCarrera(t *testing.T) {
	manager, _ := newManager()
	const n = 20
	createErrs := make([]error, n)
	listErrs := make([]error, n)

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, createErrs[i] = manager.Create(fmt.Sprintf("Proyecto %02d", i), "")
		}()
		go func() {
			defer wg.Done()
			_, listErrs[i] = manager.List()
		}()
	}
	wg.Wait()

	for i := range n {
		if createErrs[i] != nil {
			t.Errorf("Create %d devolvió un error inesperado: %v", i, createErrs[i])
		}
		if listErrs[i] != nil {
			t.Errorf("List %d devolvió un error inesperado: %v", i, listErrs[i])
		}
	}
	got := mustList(t, manager)
	if len(got) != n {
		t.Fatalf("List() tiene %d proyectos, se esperaban %d", len(got), n)
	}
	for i, p := range got {
		if want := fmt.Sprintf("Proyecto %02d", i); p.Name != want {
			t.Errorf("proyecto %d = %q, se esperaba %q", i, p.Name, want)
		}
	}
}
