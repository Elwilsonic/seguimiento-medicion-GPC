package backlog

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-11 · Asignar historias del backlog a un sprint (Issue #25). AssignToSprint
// escribe la asignación en el ítem; la validación del sprint (existencia,
// estado y proyecto) la hace sprint.Manager.AssignItem, su único llamador
// (RN-11.7).

// CA-11.1 y RN-11.4
func TestAssignToSprint_ItemSinSprintQuedaAsignado(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)

	assigned, err := manager.AssignToSprint(itemID, 3)

	if err != nil {
		t.Fatalf("AssignToSprint() devolvió un error inesperado: %v", err)
	}
	if assigned.SprintID != 3 {
		t.Errorf("SprintID devuelto = %d, se esperaba 3", assigned.SprintID)
	}
	if assigned.ID != itemID || assigned.Title != validTitle {
		t.Errorf("ítem devuelto = %+v, se esperaba el ítem %d", assigned, itemID)
	}
	saved := savedItem(t, repo, itemID)
	if saved.SprintID != 3 {
		t.Errorf("SprintID guardado = %d, se esperaba 3", saved.SprintID)
	}

	// RN-11.7: la lista de criterios devuelta no comparte memoria con la
	// guardada.
	assigned.AcceptanceCriteria[0] = "Otro criterio"
	if !slices.Equal(savedItem(t, repo, itemID).AcceptanceCriteria, validCriteria()) {
		t.Errorf("la lista devuelta comparte memoria con la guardada")
	}
}

// CA-11.2 y RN-11.3
func TestAssignToSprint_ItemYaAsignadoDaErrItemAlreadyAssigned(t *testing.T) {
	cases := map[string]int{"al mismo sprint": 1, "a otro sprint": 2}
	for name, secondSprint := range cases {
		t.Run(name, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")
			itemID := createTestItem(t, manager, projectID, validTitle)

			if _, err := manager.AssignToSprint(itemID, 1); err != nil {
				t.Fatalf("la primera asignación falló: %v", err)
			}

			assigned, err := manager.AssignToSprint(itemID, secondSprint)

			if !errors.Is(err, ErrItemAlreadyAssigned) {
				t.Errorf("error = %v, se esperaba ErrItemAlreadyAssigned", err)
			}
			if assigned.ID != 0 || assigned.SprintID != 0 {
				t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
			}
			if saved := savedItem(t, repo, itemID); saved.SprintID != 1 {
				t.Errorf("SprintID guardado = %d, se esperaba 1", saved.SprintID)
			}
		})
	}
}

// CA-11.4
func TestAssignToSprint_ItemInexistenteDaErrItemNotFound(t *testing.T) {
	for name, itemID := range map[string]int{"ID grande": 999, "ID cero": 0, "ID negativo": -1} {
		t.Run(name, func(t *testing.T) {
			manager, _, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")
			createTestItem(t, manager, projectID, validTitle)

			assigned, err := manager.AssignToSprint(itemID, 1)

			if !errors.Is(err, ErrItemNotFound) {
				t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
			}
			if assigned.ID != 0 {
				t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
			}
		})
	}
}

// CA-11.5
func TestAssignToSprint_NoCambiaNingunOtroCampo(t *testing.T) {
	repo := NewMemoryRepository()
	projects := project.NewManager(project.NewMemoryRepository())
	manager := NewManager(repo, projects)
	created, err := repo.Save(Item{
		ProjectID:          1,
		Title:              validTitle,
		Description:        "Formulario de acceso",
		Priority:           PriorityHigh,
		AcceptanceCriteria: []string{"Login OK", "Carga correcta"},
		Status:             StatusInProgress,
		StoryPoints:        5,
		EstimatedHours:     2.5,
	})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}

	if _, err := manager.AssignToSprint(created.ID, 3); err != nil {
		t.Fatalf("AssignToSprint() devolvió un error inesperado: %v", err)
	}

	assigned := savedItem(t, repo, created.ID)
	assigned.SprintID = created.SprintID
	// Item contiene una lista, así que se compara como en el contrato de
	// TEC-05: los criterios con slices.Equal y el resto con DeepEqual.
	assignedCriteria, createdCriteria := assigned.AcceptanceCriteria, created.AcceptanceCriteria
	assigned.AcceptanceCriteria, created.AcceptanceCriteria = nil, nil
	if !reflect.DeepEqual(assigned, created) {
		t.Errorf("ítem guardado cambió: %+v, se esperaba %+v", assigned, created)
	}
	if !slices.Equal(assignedCriteria, createdCriteria) {
		t.Errorf("criterios guardados = %q, se esperaban %q", assignedCriteria, createdCriteria)
	}
}

// CA-11.7
func TestAssignToSprint_ModificarElItemDevueltoNoAlteraElGuardado(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)

	assigned, err := manager.AssignToSprint(itemID, 3)
	if err != nil {
		t.Fatalf("AssignToSprint() devolvió un error inesperado: %v", err)
	}

	assigned.Title = "Otro título"
	assigned.AcceptanceCriteria[0] = "Otro criterio"
	assigned.SprintID = 999

	saved := savedItem(t, repo, itemID)
	if saved.Title != validTitle {
		t.Errorf("título guardado = %q, se esperaba %q", saved.Title, validTitle)
	}
	if want := validCriteria(); !slices.Equal(saved.AcceptanceCriteria, want) {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, want)
	}
	if saved.SprintID != 3 {
		t.Errorf("SprintID guardado = %d, se esperaba 3", saved.SprintID)
	}
}

// CA-11.5 y RN-11.5 (la única línea de HU-07 que HU-11 toca)
func TestAssignToSprint_UpdateConservaElSprintID(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	itemID := createTestItem(t, manager, projectID, validTitle)

	if _, err := manager.AssignToSprint(itemID, 1); err != nil {
		t.Fatalf("asignar el ítem de prueba: %v", err)
	}

	if _, err := manager.Update(itemID, "Login con email", "", validPriority, validCriteria(), 4); err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}

	saved := savedItem(t, repo, itemID)
	if saved.SprintID != 1 {
		t.Errorf("SprintID guardado = %d, se esperaba 1", saved.SprintID)
	}
	if saved.Title != "Login con email" {
		t.Errorf("título guardado = %q, se esperaba %q", saved.Title, "Login con email")
	}
}

// RN-11.5 (la única línea de HU-04 que HU-11 toca)
func TestAssignToSprint_CreateDejaElSprintIDEnCero(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	if _, err := manager.Create(projectID, validTitle, "", validPriority, validCriteria()); err != nil {
		t.Fatalf("crear el ítem de prueba: %v", err)
	}

	if saved := savedItem(t, repo, 1); saved.SprintID != 0 {
		t.Errorf("SprintID guardado = %d, se esperaba 0", saved.SprintID)
	}
}

// CA-11.8
func TestAssignToSprint_ErrorDelRepositorioAlConsultarSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	manager := NewManager(failingRepository{getErr: errRepository}, projects)

	assigned, err := manager.AssignToSprint(1, 3)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
	if assigned.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
	}
}

// CA-11.8
func TestAssignToSprint_ErrorDelRepositorioAlEscribirSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	manager := NewManager(failingRepository{updateErr: errRepository}, projects)
	// failingRepository.GetByID devuelve un ítem vacío sin error, por eso el
	// flujo de AssignToSprint llega hasta la escritura y devuelve el error
	// del repositorio.

	assigned, err := manager.AssignToSprint(1, 3)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
	if assigned.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
	}
}
