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

// HU-06 · Listar el backlog de un proyecto (Issue #21)

// CA-06.1, CA-06.3, RN-06.3 y RN-06.6
func TestListItems_DevuelveSoloLosItemsDelProyectoOrdenadosPorIDConTodosSusDatos(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	first := createTestProject(t, projects, "Sistema de ventas")
	second := createTestProject(t, projects, "Portal de clientes")

	if _, err := manager.Create(first, validTitle, "Acceso con usuario y contraseña", validPriority, []string{"Login OK"}); err != nil {
		t.Fatalf("crear el primer ítem de prueba: %v", err)
	}
	if _, err := manager.Create(second, validTitle, "Portal", validPriority, validCriteria()); err != nil {
		t.Fatalf("crear el ítem del otro proyecto: %v", err)
	}
	if _, err := manager.Create(first, "Cierre de sesión", "", "MEDIA", []string{"Sesión cerrada", ""}); err != nil {
		t.Fatalf("crear el segundo ítem de prueba: %v", err)
	}

	got, err := manager.ListItems(first)
	if err != nil {
		t.Fatalf("ListItems() devolvió un error inesperado: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListItems() devolvió %d ítems, se esperaban 2", len(got))
	}
	if got[0].ID != 1 || got[1].ID != 3 {
		t.Errorf("IDs = %d y %d, se esperaban 1 y 3", got[0].ID, got[1].ID)
	}

	firstItem := got[0]
	if firstItem.ProjectID != first {
		t.Errorf("ProjectID = %d, se esperaba %d", firstItem.ProjectID, first)
	}
	if firstItem.Title != validTitle {
		t.Errorf("Title = %q, se esperaba %q", firstItem.Title, validTitle)
	}
	if firstItem.Description != "Acceso con usuario y contraseña" {
		t.Errorf("Description = %q, se esperaba %q", firstItem.Description, "Acceso con usuario y contraseña")
	}
	if firstItem.Priority != PriorityHigh {
		t.Errorf("Priority = %q, se esperaba %q", firstItem.Priority, PriorityHigh)
	}
	if !slices.Equal(firstItem.AcceptanceCriteria, []string{"Login OK"}) {
		t.Errorf("AcceptanceCriteria = %q, se esperaban %q", firstItem.AcceptanceCriteria, []string{"Login OK"})
	}
	if firstItem.Status != StatusPending {
		t.Errorf("Status = %q, se esperaba %q", firstItem.Status, StatusPending)
	}
	if firstItem.StoryPoints != 0 {
		t.Errorf("StoryPoints = %d, se esperaba 0", firstItem.StoryPoints)
	}

	secondItem := got[1]
	if secondItem.Title != "Cierre de sesión" {
		t.Errorf("Title = %q, se esperaba %q", secondItem.Title, "Cierre de sesión")
	}
	if secondItem.Priority != PriorityMedium {
		t.Errorf("Priority = %q, se esperaba %q", secondItem.Priority, PriorityMedium)
	}
	if !slices.Equal(secondItem.AcceptanceCriteria, []string{"Sesión cerrada"}) {
		t.Errorf("AcceptanceCriteria = %q, se esperaban %q", secondItem.AcceptanceCriteria, []string{"Sesión cerrada"})
	}
}

// CA-06.2 y RN-06.2
func TestListItems_ProyectoSinItemsDevuelveListaVaciaNoNil(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	got, err := manager.ListItems(projectID)

	if err != nil {
		t.Fatalf("ListItems() devolvió un error inesperado: %v", err)
	}
	if got == nil {
		t.Fatal("ListItems() devolvió nil, se esperaba una lista vacía")
	}
	if len(got) != 0 {
		t.Errorf("ListItems() devolvió %d ítems, se esperaban 0", len(got))
	}
}

// CA-06.4 y RN-06.4
func TestListItems_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	cases := map[string]int{
		"ID inexistente": 999,
		"ID cero":        0,
		"ID negativo":    -1,
	}
	for caseName, projectID := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, _, projects, _ := newTestManager()
			createTestProject(t, projects, "Sistema de ventas")

			got, err := manager.ListItems(projectID)

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
			}
			if got != nil {
				t.Errorf("lista = %v, se esperaba nil", got)
			}
		})
	}
}

// CA-06.4
func TestListItems_SinProyectosCreadosDevuelveErrProjectNotFound(t *testing.T) {
	manager, _, _, _ := newTestManager()

	got, err := manager.ListItems(1)

	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
	}
	if got != nil {
		t.Errorf("lista = %v, se esperaba nil", got)
	}
}

// RN-06.1
func TestListItems_ProyectoFinalizadoListaSusItems(t *testing.T) {
	manager, _, _, projectRepo := newTestManager()

	p, err := projectRepo.Save(project.Project{Name: "Proyecto cerrado", Status: project.StatusFinished})
	if err != nil {
		t.Fatalf("guardar el proyecto finalizado de prueba: %v", err)
	}
	if _, err := manager.Create(p.ID, "Entrega final", "", validPriority, validCriteria()); err != nil {
		t.Fatalf("crear el ítem de prueba: %v", err)
	}

	got, err := manager.ListItems(p.ID)

	if err != nil {
		t.Fatalf("ListItems() devolvió un error inesperado: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListItems() devolvió %d ítems, se esperaba 1", len(got))
	}
	if got[0].Title != "Entrega final" {
		t.Errorf("Title = %q, se esperaba %q", got[0].Title, "Entrega final")
	}
}

// CA-06.6 y RN-06.5
func TestListItems_ModificarLaListaDevueltaNoAlteraLosItemsGuardados(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	if _, err := manager.Create(projectID, validTitle, "Original", validPriority, []string{"Primero"}); err != nil {
		t.Fatalf("crear el ítem de prueba: %v", err)
	}

	got, err := manager.ListItems(projectID)
	if err != nil {
		t.Fatalf("ListItems() devolvió un error inesperado: %v", err)
	}

	got[0].Title = "Otro título"
	got[0].Description = "Otra descripción"
	got[0].Priority = PriorityLow
	got[0].Status = StatusFinished
	got[0].StoryPoints = 5
	got[0].AcceptanceCriteria[0] = "Otro criterio"
	got = append(got, Item{Title: "Intruso"})

	fresh, err := manager.ListItems(projectID)
	if err != nil {
		t.Fatalf("volver a listar devolvió un error inesperado: %v", err)
	}
	want := Item{
		ID:                 1,
		ProjectID:          projectID,
		Title:              validTitle,
		Description:        "Original",
		Priority:           PriorityHigh,
		AcceptanceCriteria: []string{"Primero"},
		Status:             StatusPending,
		StoryPoints:        0,
	}
	if len(fresh) != 1 {
		t.Fatalf("la nueva lista tiene %d ítems, se esperaba 1", len(fresh))
	}
	if !reflect.DeepEqual(fresh[0], want) {
		t.Errorf("ítem re-listado = %+v, se esperaba %+v", fresh[0], want)
	}
}

// CA-06.5
func TestListItems_ListadosSimultaneosConCreaciones(t *testing.T) {
	const attempts = 50
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	errs := make([]error, attempts*2)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, errs[i] = manager.Create(projectID, fmt.Sprintf("Login de usuarios %d", i), "", validPriority, []string{"Login OK"})
		}()
		go func() {
			defer wg.Done()
			_, errs[attempts+i] = manager.ListItems(projectID)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("la operación %d falló: %v", i+1, err)
		}
	}
	if len(repo.items) != attempts {
		t.Fatalf("hay %d ítems guardados, se esperaban %d", len(repo.items), attempts)
	}

	got, err := manager.ListItems(projectID)
	if err != nil {
		t.Fatalf("ListItems() final devolvió un error inesperado: %v", err)
	}
	if len(got) != attempts {
		t.Fatalf("el listado final tiene %d ítems, se esperaban %d", len(got), attempts)
	}
	ids := make(map[int]bool)
	for _, item := range got {
		if ids[item.ID] {
			t.Errorf("el ID %d está repetido en el listado", item.ID)
		}
		ids[item.ID] = true
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].ID >= got[i].ID {
			t.Fatalf("el listado no está ordenado por ID ascendente: %v y %v", got[i-1].ID, got[i].ID)
		}
	}
}

// CA-06.7
func TestListItems_ErrorDelRepositorioAlListarSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	if _, err := projects.Create("Sistema de ventas", ""); err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	manager := NewManager(failingRepository{listErr: errRepository}, projects)

	got, err := manager.ListItems(1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if got != nil {
		t.Errorf("lista = %v, se esperaba nil", got)
	}
}
