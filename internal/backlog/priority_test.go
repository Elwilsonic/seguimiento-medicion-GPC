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

// HU-09 · Priorizar los ítems del backlog (Issue #22)

// createItemsWithPriorities crea un ítem por prioridad, en ese orden, y
// devuelve los ítems creados.
func createItemsWithPriorities(t *testing.T, manager *Manager, projectID int, priorities ...string) []Item {
	t.Helper()
	created := make([]Item, 0, len(priorities))
	for i, priority := range priorities {
		item, err := manager.Create(projectID, fmt.Sprintf("Ítem %d", i+1), "", priority, []string{"Criterio"})
		if err != nil {
			t.Fatalf("crear el ítem de prueba %d: %v", i+1, err)
		}
		created = append(created, item)
	}
	return created
}

// assertIDs verifica los IDs de la lista, en orden.
func assertIDs(t *testing.T, got []Item, want []int) {
	t.Helper()
	if ids := itemIDs(got); !slices.Equal(ids, want) {
		t.Errorf("IDs = %v, se esperaban %v", ids, want)
	}
}

// CA-09.1 y RN-09.1
func TestListItemsByPriority_OrdenaAltaMediaBajaConTodosSusDatos(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	if _, err := manager.Create(projectID, "Exportar reportes", "Reportes en PDF", "BAJA", []string{"PDF generado"}); err != nil {
		t.Fatalf("crear el ítem BAJA: %v", err)
	}
	if _, err := manager.Create(projectID, "Recuperar contraseña", "", "MEDIA", []string{"Mail enviado"}); err != nil {
		t.Fatalf("crear el ítem MEDIA: %v", err)
	}
	if _, err := manager.Update(2, "Recuperar contraseña", "Por mail", "MEDIA", []string{"Mail enviado", "Link vence"}, 4.5); err != nil {
		t.Fatalf("modificar el ítem MEDIA: %v", err)
	}
	if _, err := manager.Create(projectID, validTitle, "Acceso con usuario y contraseña", "ALTA", []string{"Login OK"}); err != nil {
		t.Fatalf("crear el ítem ALTA: %v", err)
	}

	got, err := manager.ListItemsByPriority(projectID)

	if err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}
	want := []Item{
		{ID: 3, ProjectID: projectID, Title: validTitle, Description: "Acceso con usuario y contraseña", Priority: PriorityHigh, AcceptanceCriteria: []string{"Login OK"}, Status: StatusPending},
		{ID: 2, ProjectID: projectID, Title: "Recuperar contraseña", Description: "Por mail", Priority: PriorityMedium, AcceptanceCriteria: []string{"Mail enviado", "Link vence"}, Status: StatusPending, EstimatedHours: 4.5},
		{ID: 1, ProjectID: projectID, Title: "Exportar reportes", Description: "Reportes en PDF", Priority: PriorityLow, AcceptanceCriteria: []string{"PDF generado"}, Status: StatusPending},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListItemsByPriority() = %+v, se esperaba %+v", got, want)
	}
}

// CA-09.2 y RN-09.2. Son 24 ítems (más de 12) para que un ordenamiento no
// estable se note: hasta 12 elementos slices.SortFunc usa inserción, que no
// cambia el orden de los iguales.
func TestListItemsByPriority_MismaPrioridadConservaElOrdenPorID(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	priorities := make([]string, 24)
	for i := range priorities {
		priorities[i] = []string{"BAJA", "ALTA", "MEDIA"}[i%3]
	}
	createItemsWithPriorities(t, manager, projectID, priorities...)

	got, err := manager.ListItemsByPriority(projectID)

	if err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}
	assertIDs(t, got, []int{
		2, 5, 8, 11, 14, 17, 20, 23, // ALTA
		3, 6, 9, 12, 15, 18, 21, 24, // MEDIA
		1, 4, 7, 10, 13, 16, 19, 22, // BAJA
	})
}

// CA-09.2: si todos tienen la misma prioridad, el orden es el de ListItems.
func TestListItemsByPriority_TodosConLaMismaPrioridadQuedanPorID(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	createItemsWithPriorities(t, manager, projectID, "MEDIA", "MEDIA", "MEDIA", "MEDIA")

	got, err := manager.ListItemsByPriority(projectID)

	if err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}
	assertIDs(t, got, []int{1, 2, 3, 4})
}

// Caso límite de la SDD: un solo ítem.
func TestListItemsByPriority_UnSoloItem(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	createItemsWithPriorities(t, manager, projectID, "BAJA")

	got, err := manager.ListItemsByPriority(projectID)

	if err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}
	assertIDs(t, got, []int{1})
}

// CA-09.3 y RN-09.7
func TestListItemsByPriority_ProyectoSinItemsDevuelveListaVaciaNoNil(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	got, err := manager.ListItemsByPriority(projectID)

	if err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}
	if got == nil {
		t.Fatal("ListItemsByPriority() devolvió nil, se esperaba una lista vacía")
	}
	if len(got) != 0 {
		t.Errorf("ListItemsByPriority() devolvió %d ítems, se esperaban 0", len(got))
	}
}

// CA-09.4 y RN-09.4: ítems en distintos estados, sin los de otro proyecto.
func TestListItemsByPriority_DevuelveLosMismosItemsQueListItems(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	first := createTestProject(t, projects, "Sistema de ventas")
	second := createTestProject(t, projects, "Portal de clientes")

	createItemsWithPriorities(t, manager, first, "MEDIA")  // ID 1, PENDIENTE
	createItemsWithPriorities(t, manager, second, "ALTA")  // ID 2, otro proyecto
	createItemsWithPriorities(t, manager, first, "BAJA")   // ID 3, EN_CURSO
	createItemsWithPriorities(t, manager, first, "ALTA")   // ID 4, TERMINADO
	createItemsWithPriorities(t, manager, second, "MEDIA") // ID 5, otro proyecto
	for _, change := range []struct {
		id     int
		status string
	}{{3, "EN_CURSO"}, {4, "EN_CURSO"}, {4, "TERMINADO"}} {
		if _, err := manager.ChangeStatus(change.id, change.status); err != nil {
			t.Fatalf("cambiar el estado del ítem %d a %s: %v", change.id, change.status, err)
		}
	}

	got, err := manager.ListItemsByPriority(first)
	if err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}
	assertIDs(t, got, []int{4, 1, 3})

	listed, err := manager.ListItems(first)
	if err != nil {
		t.Fatalf("ListItems() devolvió un error inesperado: %v", err)
	}
	byID := slices.Clone(got)
	slices.SortFunc(byID, func(a, b Item) int { return a.ID - b.ID })
	if !reflect.DeepEqual(byID, listed) {
		t.Errorf("ítems por prioridad ordenados por ID = %+v, ListItems() = %+v", byID, listed)
	}
}

// CA-09.4 y RN-09.4 (RN-06.1)
func TestListItemsByPriority_ProyectoFinalizadoListaSusItems(t *testing.T) {
	manager, _, _, projectRepo := newTestManager()

	p, err := projectRepo.Save(project.Project{Name: "Proyecto cerrado", Status: project.StatusFinished})
	if err != nil {
		t.Fatalf("guardar el proyecto finalizado de prueba: %v", err)
	}
	createItemsWithPriorities(t, manager, p.ID, "BAJA", "ALTA")

	got, err := manager.ListItemsByPriority(p.ID)

	if err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}
	assertIDs(t, got, []int{2, 1})
}

// CA-09.5 y RN-09.5
func TestListItemsByPriority_NoCambiaElOrdenDeListItems(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	createItemsWithPriorities(t, manager, projectID, "BAJA", "MEDIA", "ALTA")

	if _, err := manager.ListItemsByPriority(projectID); err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}
	listed, err := manager.ListItems(projectID)

	if err != nil {
		t.Fatalf("ListItems() devolvió un error inesperado: %v", err)
	}
	assertIDs(t, listed, []int{1, 2, 3})
}

// CA-09.6 y RN-09.3: los ítems con prioridad desconocida se guardan directo en
// el repositorio, porque Create y Update no la aceptan.
func TestListItemsByPriority_PrioridadDesconocidaVaAlFinalPorID(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	saveUnknown := func(priority string) {
		t.Helper()
		if _, err := repo.Save(Item{ProjectID: projectID, Title: "Desconocida " + priority, Priority: Priority(priority), AcceptanceCriteria: []string{"Criterio"}, Status: StatusPending}); err != nil {
			t.Fatalf("guardar el ítem con prioridad %q: %v", priority, err)
		}
	}
	saveUnknown("URGENTE")                                            // ID 1
	createItemsWithPriorities(t, manager, projectID, "BAJA")          // ID 2
	saveUnknown("alta")                                               // ID 3, sin normalizar
	createItemsWithPriorities(t, manager, projectID, "ALTA", "MEDIA") // IDs 4 y 5
	saveUnknown("")                                                   // ID 6

	got, err := manager.ListItemsByPriority(projectID)

	if err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}
	assertIDs(t, got, []int{4, 5, 2, 1, 3, 6})
}

// CA-09.7 y RN-09.6
func TestListItemsByPriority_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	cases := map[string]int{
		"ID inexistente": 999,
		"ID cero":        0,
		"ID negativo":    -1,
	}
	for caseName, projectID := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, _, projects, _ := newTestManager()
			createTestProject(t, projects, "Sistema de ventas")

			got, err := manager.ListItemsByPriority(projectID)

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
			}
			if got != nil {
				t.Errorf("lista = %v, se esperaba nil", got)
			}
		})
	}
}

// CA-09.7
func TestListItemsByPriority_SinProyectosCreadosDevuelveErrProjectNotFound(t *testing.T) {
	manager, _, _, _ := newTestManager()

	got, err := manager.ListItemsByPriority(1)

	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
	}
	if got != nil {
		t.Errorf("lista = %v, se esperaba nil", got)
	}
}

// CA-09.8
func TestListItemsByPriority_ErrorDelRepositorioAlListarSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	if _, err := projects.Create("Sistema de ventas", ""); err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	manager := NewManager(failingRepository{listErr: errRepository}, projects)

	got, err := manager.ListItemsByPriority(1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if got != nil {
		t.Errorf("lista = %v, se esperaba nil", got)
	}
}

// CA-09.9: corre con go test -race en el CI.
func TestListItemsByPriority_ListadosSimultaneosConCreaciones(t *testing.T) {
	const attempts = 50
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	errs := make([]error, attempts*2)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(2)
		go func() {
			defer wg.Done()
			priority := []string{"BAJA", "ALTA", "MEDIA"}[i%3]
			_, errs[i] = manager.Create(projectID, fmt.Sprintf("Ítem %d", i), "", priority, []string{"Criterio"})
		}()
		go func() {
			defer wg.Done()
			_, errs[attempts+i] = manager.ListItemsByPriority(projectID)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("la operación %d falló: %v", i+1, err)
		}
	}
	got, err := manager.ListItemsByPriority(projectID)
	if err != nil {
		t.Fatalf("ListItemsByPriority() final devolvió un error inesperado: %v", err)
	}
	if len(got) != attempts {
		t.Fatalf("el listado final tiene %d ítems, se esperaban %d", len(got), attempts)
	}

	// El listado final tiene que estar ordenado por prioridad y, a igual
	// prioridad, por ID. El rango se define acá a propósito, sin usar el del
	// código de producción.
	rank := map[Priority]int{PriorityHigh: 0, PriorityMedium: 1, PriorityLow: 2}
	for i := 1; i < len(got); i++ {
		prev, curr := got[i-1], got[i]
		if rank[curr.Priority] < rank[prev.Priority] {
			t.Fatalf("posición %d: prioridad %s después de %s", i+1, curr.Priority, prev.Priority)
		}
		if rank[curr.Priority] == rank[prev.Priority] && curr.ID < prev.ID {
			t.Fatalf("posición %d: con prioridad %s, el ID %d está después del %d", i+1, curr.Priority, curr.ID, prev.ID)
		}
	}
}

// CA-09.10 y RN-09.8
func TestListItemsByPriority_ModificarLaListaDevueltaNoAlteraLosItemsGuardados(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	createItemsWithPriorities(t, manager, projectID, "BAJA", "ALTA")

	got, err := manager.ListItemsByPriority(projectID)
	if err != nil {
		t.Fatalf("ListItemsByPriority() devolvió un error inesperado: %v", err)
	}

	got[0].Title = "Otro título"
	got[0].Priority = PriorityLow
	got[0].Status = StatusFinished
	got[0].AcceptanceCriteria[0] = "Otro criterio"
	got[0], got[1] = got[1], got[0]
	got = append(got, Item{Title: "Intruso"})

	fresh, err := manager.ListItemsByPriority(projectID)
	if err != nil {
		t.Fatalf("volver a listar devolvió un error inesperado: %v", err)
	}
	want := []Item{
		{ID: 2, ProjectID: projectID, Title: "Ítem 2", Priority: PriorityHigh, AcceptanceCriteria: []string{"Criterio"}, Status: StatusPending},
		{ID: 1, ProjectID: projectID, Title: "Ítem 1", Priority: PriorityLow, AcceptanceCriteria: []string{"Criterio"}, Status: StatusPending},
	}
	if !reflect.DeepEqual(fresh, want) {
		t.Errorf("listado nuevo = %+v, se esperaba %+v", fresh, want)
	}
}
