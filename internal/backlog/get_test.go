package backlog

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-05 · Consultar un ítem del backlog (Issue #18)

// CA-05.1
func TestGetItem_ItemExistenteDevuelveCopiaConTodosSusDatos(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	created, err := manager.Create(projectID, validTitle, "Acceso con usuario y contraseña", validPriority, []string{"Credenciales válidas", "Credenciales inválidas"})
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	fetched, err := manager.GetItem(created.ID)

	if err != nil {
		t.Fatalf("GetItem() devolvió un error inesperado: %v", err)
	}
	if !reflect.DeepEqual(fetched, created) {
		t.Errorf("ítem consultado = %+v, se esperaba %+v", fetched, created)
	}
	if fetched.ID != created.ID {
		t.Errorf("ID = %d, se esperaba %d", fetched.ID, created.ID)
	}
	if fetched.ProjectID != projectID {
		t.Errorf("ProjectID = %d, se esperaba %d", fetched.ProjectID, projectID)
	}
	if fetched.Title != validTitle {
		t.Errorf("Title = %q, se esperaba %q", fetched.Title, validTitle)
	}
	if fetched.Description != "Acceso con usuario y contraseña" {
		t.Errorf("Description = %q, se esperaba %q", fetched.Description, "Acceso con usuario y contraseña")
	}
	if fetched.Priority != PriorityHigh {
		t.Errorf("Priority = %q, se esperaba %q", fetched.Priority, PriorityHigh)
	}
	if !slices.Equal(fetched.AcceptanceCriteria, []string{"Credenciales válidas", "Credenciales inválidas"}) {
		t.Errorf("AcceptanceCriteria = %q, se esperaban %q", fetched.AcceptanceCriteria, []string{"Credenciales válidas", "Credenciales inválidas"})
	}
	if fetched.Status != StatusPending {
		t.Errorf("Status = %q, se esperaba %q", fetched.Status, StatusPending)
	}
	if fetched.StoryPoints != 0 {
		t.Errorf("StoryPoints = %d, se esperaba 0", fetched.StoryPoints)
	}
}

// CA-05.2
func TestGetItem_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
	cases := map[string]int{
		"ID inexistente": 999,
		"ID cero":        0,
		"ID negativo":    -1,
	}
	for caseName, id := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, _, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")
			if _, err := manager.Create(projectID, validTitle, "", validPriority, validCriteria()); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			item, err := manager.GetItem(id)

			if !errors.Is(err, ErrItemNotFound) {
				t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
			}
			if !reflect.DeepEqual(item, Item{}) {
				t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", item)
			}
		})
	}
}

// CA-05.2
func TestGetItem_ConsultarSinItemsCreadosDevuelveErrItemNotFound(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	createTestProject(t, projects, "Sistema de ventas")

	item, err := manager.GetItem(1)

	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
	}
	if !reflect.DeepEqual(item, Item{}) {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", item)
	}
}

// CA-05.3
func TestGetItem_ModificarElItemConsultadoYSusCriteriosNoAlteraElGuardado(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	if _, err := manager.Create(projectID, validTitle, "", validPriority, []string{"Login OK"}); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	fetched, err := manager.GetItem(1)
	if err != nil {
		t.Fatalf("GetItem() devolvió un error inesperado: %v", err)
	}

	fetched.Title = "Otro título"
	fetched.Description = "Otra descripción"
	fetched.Priority = PriorityLow
	fetched.Status = StatusFinished
	fetched.StoryPoints = 5
	fetched.AcceptanceCriteria[0] = "Otro criterio"
	fetched.AcceptanceCriteria = append(fetched.AcceptanceCriteria, "Criterio extra")

	again, err := manager.GetItem(1)
	if err != nil {
		t.Fatalf("GetItem() devolvió un error inesperado: %v", err)
	}
	if again.Title != validTitle || again.Description != "" || again.Priority != PriorityHigh ||
		again.Status != StatusPending || again.StoryPoints != 0 {
		t.Errorf("al volver a consultar el ítem cambió: %+v", again)
	}
	if !slices.Equal(again.AcceptanceCriteria, []string{"Login OK"}) {
		t.Errorf("criterios al volver a consultar = %q, se esperaban %q", again.AcceptanceCriteria, []string{"Login OK"})
	}
	saved := savedItem(t, repo, 1)
	if !slices.Equal(saved.AcceptanceCriteria, []string{"Login OK"}) {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, []string{"Login OK"})
	}
}

// CA-05.4
func TestGetItem_ConsultasSimultaneasMientrasSeCreanItems(t *testing.T) {
	const attempts = 50
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	first, err := manager.Create(projectID, validTitle, "", validPriority, validCriteria())
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	errs := make([]error, attempts*2)
	results := make([]Item, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, errs[i] = manager.Create(projectID, fmt.Sprintf("Login de usuarios %d", i), "", validPriority, validCriteria())
		}()
		go func() {
			defer wg.Done()
			results[i], errs[attempts+i] = manager.GetItem(first.ID)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("la operación simultánea %d falló: %v", i+1, err)
		}
	}
	for i, got := range results {
		if got.ID != first.ID {
			t.Errorf("la consulta %d obtuvo el ID %d, se esperaba %d", i+1, got.ID, first.ID)
		}
		if got.Title != validTitle {
			t.Errorf("la consulta %d obtuvo el título %q, se esperaba %q", i+1, got.Title, validTitle)
		}
	}
	if len(repo.items) != attempts+1 {
		t.Errorf("hay %d ítems guardados, se esperaban %d", len(repo.items), attempts+1)
	}
}

// CA-05.5
func TestGetItem_ErrorDelRepositorioSeDevuelveConItemVacio(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	if _, err := projects.Create("Sistema de ventas", ""); err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	manager := NewManager(failingRepository{getErr: errRepository}, projects)

	item, err := manager.GetItem(1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if !reflect.DeepEqual(item, Item{}) {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", item)
	}
}

// CA-05.1, RN-05.3
func TestGetByID_ItemExistenteDevuelveCopiaConCriteriosPropios(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	created, err := manager.Create(projectID, validTitle, "", validPriority, []string{"Login OK"})
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	fetched, err := repo.GetByID(created.ID)

	if err != nil {
		t.Fatalf("GetByID() devolvió un error inesperado: %v", err)
	}
	if !reflect.DeepEqual(fetched, created) {
		t.Errorf("ítem consultado = %+v, se esperaba %+v", fetched, created)
	}

	fetched.AcceptanceCriteria[0] = "Otro criterio"

	if saved := savedItem(t, repo, created.ID); !slices.Equal(saved.AcceptanceCriteria, []string{"Login OK"}) {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, []string{"Login OK"})
	}
	if !slices.Equal(created.AcceptanceCriteria, []string{"Login OK"}) {
		t.Errorf("criterios del ítem creado = %q, se esperaban %q", created.AcceptanceCriteria, []string{"Login OK"})
	}
}

// CA-05.2
func TestGetByID_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
	cases := map[string]int{
		"ID inexistente": 999,
		"ID cero":        0,
		"ID negativo":    -1,
	}
	for caseName, id := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")
			if _, err := manager.Create(projectID, validTitle, "", validPriority, validCriteria()); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			item, err := repo.GetByID(id)

			if !errors.Is(err, ErrItemNotFound) {
				t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
			}
			if !reflect.DeepEqual(item, Item{}) {
				t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", item)
			}
		})
	}
}

// CA-05.2
func TestGetByID_ConsultarSinItemsCreadosDevuelveErrItemNotFound(t *testing.T) {
	_, repo, _, _ := newTestManager()

	item, err := repo.GetByID(1)

	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
	}
	if !reflect.DeepEqual(item, Item{}) {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", item)
	}
}
