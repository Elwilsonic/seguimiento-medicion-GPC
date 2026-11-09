package backlog

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-07 · Modificar datos de un ítem (Issue #19)

// createTestItem crea un ítem válido en ese proyecto y devuelve su ID.
func createTestItem(t *testing.T, manager *Manager, projectID int, title string) int {
	t.Helper()
	item, err := manager.Create(projectID, title, "", validPriority, validCriteria())
	if err != nil {
		t.Fatalf("crear el ítem de prueba: %v", err)
	}
	return item.ID
}

// HU-07 · CA-07.1 · Verifica que Update cambie todos los campos editables y
// conserve el ID, el proyecto, el estado y los story points.
func TestManagerUpdate_ModificaTodosLosCamposConservaIDProyectoEstadoYStoryPoints(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, "Login de usuarios")

	updated, err := manager.Update(itemID, "  Login con email  ", "Acceso por email", " media ", []string{"Login por email OK"}, 6)

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if updated.ID != itemID {
		t.Errorf("ID = %d, se esperaba %d", updated.ID, itemID)
	}
	if updated.ProjectID != projectID {
		t.Errorf("ProjectID = %d, se esperaba %d", updated.ProjectID, projectID)
	}
	if updated.Title != "Login con email" {
		t.Errorf("Title = %q, se esperaba %q", updated.Title, "Login con email")
	}
	if updated.Description != "Acceso por email" {
		t.Errorf("Description = %q, se esperaba %q", updated.Description, "Acceso por email")
	}
	if updated.Priority != PriorityMedium {
		t.Errorf("Priority = %q, se esperaba %q", updated.Priority, PriorityMedium)
	}
	if want := []string{"Login por email OK"}; !slices.Equal(updated.AcceptanceCriteria, want) {
		t.Errorf("AcceptanceCriteria = %q, se esperaba %q", updated.AcceptanceCriteria, want)
	}
	if updated.Status != StatusPending {
		t.Errorf("Status = %q, se esperaba %q", updated.Status, StatusPending)
	}
	if updated.StoryPoints != 0 {
		t.Errorf("StoryPoints = %d, se esperaban 0", updated.StoryPoints)
	}
	if updated.EstimatedHours != 6 {
		t.Errorf("EstimatedHours = %v, se esperaban 6", updated.EstimatedHours)
	}
}

// HU-07 · CA-07.1 · Verifica que Update no toque los story points ni el
// estado, aunque no sean los iniciales.
func TestManagerUpdate_ConservaStoryPointsYEstadoAsignados(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	// El ítem se guarda directo en el repositorio con story points y estado
	// que no son los iniciales de HU-04.
	saved, err := repo.Save(Item{ProjectID: projectID, Title: validTitle,
		Priority: PriorityHigh, AcceptanceCriteria: validCriteria(),
		Status: StatusInProgress, StoryPoints: 5})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}

	updated, err := manager.Update(saved.ID, "Login con email", "", "MEDIA", []string{"Login OK"}, 4)

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if updated.StoryPoints != 5 {
		t.Errorf("StoryPoints = %d, se esperaban 5", updated.StoryPoints)
	}
	if updated.Status != StatusInProgress {
		t.Errorf("Status = %q, se esperaba %q", updated.Status, StatusInProgress)
	}
	if s := savedItem(t, repo, saved.ID); s.StoryPoints != 5 || s.Status != StatusInProgress {
		t.Errorf("lo guardado cambió: StoryPoints %d y Status %q", s.StoryPoints, s.Status)
	}
}

// HU-07 · CA-07.1 · Verifica que se pueda cambiar un solo campo, las horas,
// reenviando el resto con su valor actual.
func TestManagerUpdate_CambiaSoloLasHorasReenviandoElResto(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, "Login de usuarios")

	updated, err := manager.Update(itemID, "Login de usuarios", "", validPriority, validCriteria(), 4)

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if updated.Title != "Login de usuarios" {
		t.Errorf("Title = %q, se esperaba %q", updated.Title, "Login de usuarios")
	}
	if updated.Priority != PriorityHigh {
		t.Errorf("Priority = %q, se esperaba %q", updated.Priority, PriorityHigh)
	}
	if updated.EstimatedHours != 4 {
		t.Errorf("EstimatedHours = %v, se esperaban 4", updated.EstimatedHours)
	}
}

// HU-07 · CA-07.4 · Verifica que el ítem creado por HU-04 quede sin horas
// estimadas.
func TestCreate_ItemNuevoQuedaSinHorasEstimadas(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	item, err := manager.Create(projectID, validTitle, "", validPriority, validCriteria())
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if item.EstimatedHours != 0 {
		t.Errorf("EstimatedHours = %v, se esperaban 0 (sin horas)", item.EstimatedHours)
	}
}

// HU-07 · CA-07.4 · Verifica que se acepten horas estimadas fraccionadas.
func TestManagerUpdate_AceptaHorasFraccionadas(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)

	updated, err := manager.Update(itemID, validTitle, "", validPriority, validCriteria(), 2.5)

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if updated.EstimatedHours != 2.5 {
		t.Errorf("EstimatedHours = %v, se esperaban 2.5", updated.EstimatedHours)
	}
}

// HU-07 · CA-07.4 · Verifica que 0 horas signifique sin horas y quite el valor
// anterior.
func TestManagerUpdate_HorasEnCeroDejanElItemSinHoras(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)
	if _, err := manager.Update(itemID, validTitle, "", validPriority, validCriteria(), 8.5); err != nil {
		t.Fatalf("asignar las horas de prueba: %v", err)
	}

	updated, err := manager.Update(itemID, validTitle, "", validPriority, validCriteria(), 0)

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if updated.EstimatedHours != 0 {
		t.Errorf("EstimatedHours = %v, se esperaban 0 (sin horas)", updated.EstimatedHours)
	}
	if saved := savedItem(t, repo, itemID); saved.EstimatedHours != 0 {
		t.Errorf("horas guardadas = %v, se esperaban 0", saved.EstimatedHours)
	}
}

// HU-07 · CA-07.4 · Verifica que horas negativas, NaN o infinitas sean
// inválidas y el ítem no cambie.
func TestManagerUpdate_RechazaHorasNegativasNaN_e_Infinitas(t *testing.T) {
	invalid := []struct {
		name  string
		hours float64
	}{
		{"negativas", -1},
		{"NaN", math.NaN()},
		{"infinitas", math.Inf(1)},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")
			itemID := createTestItem(t, manager, projectID, validTitle)

			_, err := manager.Update(itemID, validTitle, "", validPriority, validCriteria(), tc.hours)

			if !errors.Is(err, ErrInvalidEstimatedHours) {
				t.Errorf("error = %v, se esperaba ErrInvalidEstimatedHours", err)
			}
			if saved := savedItem(t, repo, itemID); saved.EstimatedHours != 0 {
				t.Errorf("horas guardadas = %v, se esperaban 0: el ítem no debía cambiar", saved.EstimatedHours)
			}
		})
	}
}

// HU-07 · CA-07.2 · Verifica que un título vacío devuelva ErrEmptyTitle: los
// campos obligatorios siguen validándose.
func TestManagerUpdate_RechazaTituloVacio(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)

	_, err := manager.Update(itemID, "   ", "", validPriority, validCriteria(), 4)

	if !errors.Is(err, ErrEmptyTitle) {
		t.Errorf("error = %v, se esperaba ErrEmptyTitle", err)
	}
	if saved := savedItem(t, repo, itemID); saved.Title != validTitle || saved.EstimatedHours != 0 {
		t.Errorf("el ítem guardado cambió: título %q y horas %v", saved.Title, saved.EstimatedHours)
	}
}

// HU-07 · CA-07.2 · Verifica que una prioridad inválida devuelva
// ErrInvalidPriority.
func TestManagerUpdate_RechazaPrioridadInvalida(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)

	_, err := manager.Update(itemID, validTitle, "", "URGENTE", validCriteria(), 4)

	if !errors.Is(err, ErrInvalidPriority) {
		t.Errorf("error = %v, se esperaba ErrInvalidPriority", err)
	}
	if saved := savedItem(t, repo, itemID); saved.Priority != PriorityHigh {
		t.Errorf("prioridad guardada = %q, se esperaba %q: el ítem no debía cambiar", saved.Priority, PriorityHigh)
	}
}

// HU-07 · CA-07.2 · Verifica que una modificación sin ningún criterio de
// aceptación se rechace.
func TestManagerUpdate_RechazaModificacionSinCriterios(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)

	for _, criteria := range [][]string{nil, {}, {""}, {"  "}} {
		_, err := manager.Update(itemID, validTitle, "", validPriority, criteria, 4)

		if !errors.Is(err, ErrNoAcceptanceCriteria) {
			t.Errorf("criterios %q: error = %v, se esperaba ErrNoAcceptanceCriteria", criteria, err)
		}
	}
	if saved := savedItem(t, repo, itemID); !slices.Equal(saved.AcceptanceCriteria, validCriteria()) {
		t.Errorf("criterios guardados = %q, se esperaban %q: el ítem no debía cambiar", saved.AcceptanceCriteria, validCriteria())
	}
}

// HU-07 · Error · Verifica que el contenido se valide antes de consultar el
// repositorio y se informe el primer error (restricción "Orden de
// validaciones").
func TestManagerUpdate_PrimerErrorEsElDelOrdenDeValidacion(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)

	_, err := manager.Update(itemID, "", "", validPriority, validCriteria(), -1)

	if !errors.Is(err, ErrEmptyTitle) {
		t.Errorf("error = %v, se esperaba ErrEmptyTitle (el primer error del orden de validación)", err)
	}
}

// HU-07 · CA-07.3 · Verifica que modificar un ítem inexistente devuelva
// ErrItemNotFound.
func TestManagerUpdate_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	createTestItem(t, manager, projectID, validTitle)

	for _, id := range []int{999, 0, -1} {
		if _, err := manager.Update(id, validTitle, "", validPriority, validCriteria(), 4); !errors.Is(err, ErrItemNotFound) {
			t.Errorf("ID %d: error = %v, se esperaba ErrItemNotFound", id, err)
		}
	}
}

// HU-07 · CA-07.3 · Verifica que, sin ningún ítem creado, Update devuelva
// ErrItemNotFound.
func TestManagerUpdate_SinItemsCreadosDevuelveErrItemNotFound(t *testing.T) {
	manager, _, _, _ := newTestManager()

	if _, err := manager.Update(1, validTitle, "", validPriority, validCriteria(), 4); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
	}
}

// HU-07 · CA-07.5 · Verifica que Update devuelva una copia por valor, incluida
// la lista de criterios.
func TestManagerUpdate_CopiaPorValorDelResultado(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)
	inputCriteria := []string{"Login OK"}

	updated, err := manager.Update(itemID, "Login con email", "", validPriority, inputCriteria, 4)
	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}

	updated.Title = "Otro título"
	updated.AcceptanceCriteria[0] = "Otro criterio"
	inputCriteria[0] = "Otro criterio recibido"

	saved := savedItem(t, repo, itemID)
	if saved.Title != "Login con email" {
		t.Errorf("título guardado = %q, se esperaba %q", saved.Title, "Login con email")
	}
	if want := []string{"Login OK"}; !slices.Equal(saved.AcceptanceCriteria, want) {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, want)
	}
}

// HU-07 · CA-07.7 · Verifica que, si el repositorio falla al buscar, se
// devuelva ese error.
func TestManagerUpdate_ErrorDelRepositorioAlBuscarSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	repo := failingRepository{getErr: errRepository}
	manager := NewManager(repo, projects)

	if _, err := manager.Update(1, validTitle, "", validPriority, validCriteria(), 4); !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
}

// HU-07 · CA-07.7 · Verifica que, si el repositorio falla al actualizar, se
// devuelva ese error.
func TestManagerUpdate_ErrorDelRepositorioAlActualizarSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	repo := failingRepository{updateErr: errRepository}
	manager := NewManager(repo, projects)
	// failingRepository.GetByID devuelve un ítem vacío sin error, por eso el
	// flujo de Update llega hasta la escritura y devuelve el error del repositorio.

	if _, err := manager.Update(1, validTitle, "", validPriority, validCriteria(), 4); !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
}

// HU-07 · CA-07.5 · Verifica que Update del repositorio reemplace el ítem
// guardado, devuelva una copia con su propia lista y no toque a los demás.
// También cubre CA-07.7 en la operación nueva de la interfaz.
func TestMemoryRepositoryUpdate_ReemplazaElItemYDevuelveCopia(t *testing.T) {
	repo := NewMemoryRepository()
	saved, err := repo.Save(Item{ProjectID: 1, Title: validTitle, Priority: PriorityHigh, AcceptanceCriteria: validCriteria()})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}

	updated, err := repo.Update(Item{ID: saved.ID, ProjectID: 1, Title: "Login con email", Priority: PriorityMedium, AcceptanceCriteria: []string{"Login por email OK"}})

	if err != nil {
		t.Fatalf("Update() del repositorio devolvió un error inesperado: %v", err)
	}
	if updated.Title != "Login con email" || updated.Priority != PriorityMedium {
		t.Errorf("ítem actualizado = %+v, se esperaba el ítem modificado", updated)
	}
	if !slices.Equal(repo.items[0].AcceptanceCriteria, []string{"Login por email OK"}) {
		t.Errorf("ítem guardado = %q, se esperaba el reemplazado", repo.items[0].AcceptanceCriteria)
	}
	updated.AcceptanceCriteria[0] = "Otro criterio"
	if !slices.Equal(repo.items[0].AcceptanceCriteria, []string{"Login por email OK"}) {
		t.Errorf("la lista devuelta comparte memoria con la guardada")
	}
}

// HU-07 · CA-07.3 · Verifica que Update del repositorio, sobre un ítem
// inexistente, devuelva ErrItemNotFound (operación nueva de la interfaz).
func TestMemoryRepositoryUpdate_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
	repo := NewMemoryRepository()

	if _, err := repo.Update(Item{ID: 1, Title: validTitle, Priority: PriorityHigh, AcceptanceCriteria: validCriteria()}); !errors.Is(err, ErrItemNotFound) {
		t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
	}
}

// HU-07 · CA-07.6 · Verifica que, después de varias modificaciones simultáneas
// del mismo ítem, el ítem guardado coincida completo con una de las versiones
// enviadas.
func TestManagerUpdate_ModificacionesSimultaneasDelMismoItem(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)
	createTestItem(t, manager, projectID, "Exportar a PDF")
	createTestItem(t, manager, projectID, "Dashboard")

	const attempts = 50
	versions := make([]Item, attempts)
	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		versions[i] = Item{
			ID:                 itemID,
			ProjectID:          projectID,
			Title:              fmt.Sprintf("Título %d", i),
			Priority:           PriorityHigh,
			AcceptanceCriteria: []string{fmt.Sprintf("Criterio %d", i)},
			Status:             StatusPending,
			EstimatedHours:     float64(i),
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = manager.Update(itemID, versions[i].Title, "", "ALTA", versions[i].AcceptanceCriteria, versions[i].EstimatedHours)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("la modificación %d falló: %v", i+1, err)
		}
	}
	if len(repo.items) != 3 {
		t.Fatalf("hay %d ítems guardados, se esperaban 3", len(repo.items))
	}
	saved := savedItem(t, repo, itemID)
	matched := false
	for _, v := range versions {
		if saved.Title == v.Title && saved.Priority == v.Priority &&
			saved.Status == v.Status && saved.EstimatedHours == v.EstimatedHours &&
			slices.Equal(saved.AcceptanceCriteria, v.AcceptanceCriteria) {
			matched = true
			break
		}
	}
	if !matched {
		t.Errorf("el ítem guardado %+v no coincide con ninguna de las %d versiones enviadas", saved, attempts)
	}
}
