package sprint

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-11 · Asignar historias del backlog a un sprint (Issue #25)

// assignStack es la pila completa de HU-11: los Managers de proyectos, ítems
// y sprints con sus repositorios en memoria.
type assignStack struct {
	manager  *Manager
	repo     *MemoryRepository
	items    *backlog.Manager
	itemRepo *backlog.MemoryRepository
	projects *project.Manager
}

// newAssignStack arma la pila vacía de HU-11.
func newAssignStack() assignStack {
	projectRepo := project.NewMemoryRepository()
	projects := project.NewManager(projectRepo)
	itemRepo := backlog.NewMemoryRepository()
	items := backlog.NewManager(itemRepo, projects)
	repo := NewMemoryRepository()
	return assignStack{
		manager:  NewManager(repo, projects, items),
		repo:     repo,
		items:    items,
		itemRepo: itemRepo,
		projects: projects,
	}
}

// createAssignProject crea un proyecto y devuelve su ID.
func (s assignStack) createAssignProject(t *testing.T, name string) int {
	t.Helper()
	p, err := s.projects.Create(name, "")
	if err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	return p.ID
}

// createAssignItem crea un ítem PENDIENTE en el proyecto y devuelve su ID.
func (s assignStack) createAssignItem(t *testing.T, projectID int, title string) int {
	t.Helper()
	item, err := s.items.Create(projectID, title, "", "ALTA", []string{"Criterio OK"})
	if err != nil {
		t.Fatalf("crear el ítem de prueba: %v", err)
	}
	return item.ID
}

// saveAssignSprint guarda un sprint directo en el repositorio con el estado
// recibido y devuelve su ID. La regla de un solo sprint ABIERTO por proyecto
// la aplica el Manager (RN-10.5), no el repositorio.
func (s assignStack) saveAssignSprint(t *testing.T, projectID int, name string, status Status) int {
	t.Helper()
	saved, err := s.repo.Save(Sprint{
		ProjectID: projectID,
		Name:      name,
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}
	return saved.ID
}

// savedAssignItem consulta el ítem guardado con el Manager de backlog (HU-05).
func (s assignStack) savedAssignItem(t *testing.T, id int) backlog.Item {
	t.Helper()
	item, err := s.items.GetItem(id)
	if err != nil {
		t.Fatalf("consultar el ítem guardado: %v", err)
	}
	return item
}

// failingItemRepository es un backlog.Repository que falla al buscar; el
// resto de las operaciones no se usan en HU-11.
type failingItemRepository struct{ getErr error }

func (r failingItemRepository) Save(item backlog.Item) (backlog.Item, error) {
	return backlog.Item{}, nil
}

func (r failingItemRepository) GetByID(id int) (backlog.Item, error) {
	return backlog.Item{}, r.getErr
}

func (r failingItemRepository) ListByProject(projectID int) ([]backlog.Item, error) {
	return nil, nil
}

func (r failingItemRepository) Update(item backlog.Item) (backlog.Item, error) {
	return backlog.Item{}, nil
}

// CA-11.1, RN-11.1, RN-11.2 y RN-11.4
func TestAssignItem_ItemDelProyectoSeAsignaAlSprintAbierto(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	itemID := s.createAssignItem(t, projectID, "Login de usuarios")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

	assigned, err := s.manager.AssignItem(sprintID, itemID)

	if err != nil {
		t.Fatalf("AssignItem() devolvió un error inesperado: %v", err)
	}
	if assigned.SprintID != sprintID {
		t.Errorf("SprintID devuelto = %d, se esperaba %d", assigned.SprintID, sprintID)
	}
	if assigned.ID != itemID || assigned.Title != "Login de usuarios" {
		t.Errorf("ítem devuelto = %+v, se esperaba el ítem %d", assigned, itemID)
	}
	saved := s.savedAssignItem(t, itemID)
	if saved.SprintID != sprintID {
		t.Errorf("SprintID guardado = %d, se esperaba %d", saved.SprintID, sprintID)
	}
}

// CA-11.1
func TestAssignItem_DosItemsDistintosAlMismoSprint(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	first := s.createAssignItem(t, projectID, "Login de usuarios")
	second := s.createAssignItem(t, projectID, "Listar proyectos")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

	for _, itemID := range []int{first, second} {
		if _, err := s.manager.AssignItem(sprintID, itemID); err != nil {
			t.Fatalf("asignar el ítem %d devolvió un error inesperado: %v", itemID, err)
		}
	}

	for _, itemID := range []int{first, second} {
		saved := s.savedAssignItem(t, itemID)
		if saved.SprintID != sprintID {
			t.Errorf("SprintID guardado del ítem %d = %d, se esperaba %d", itemID, saved.SprintID, sprintID)
		}
	}
}

// CA-11.2 y RN-11.3
func TestAssignItem_ItemYaAsignadoAlMismoSprintDaErrItemAlreadyAssigned(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	itemID := s.createAssignItem(t, projectID, "Login de usuarios")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

	if _, err := s.manager.AssignItem(sprintID, itemID); err != nil {
		t.Fatalf("la primera asignación falló: %v", err)
	}

	assigned, err := s.manager.AssignItem(sprintID, itemID)

	if !errors.Is(err, backlog.ErrItemAlreadyAssigned) {
		t.Errorf("error = %v, se esperaba ErrItemAlreadyAssigned", err)
	}
	if assigned.ID != 0 || assigned.SprintID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
	}
	if saved := s.savedAssignItem(t, itemID); saved.SprintID != sprintID {
		t.Errorf("SprintID guardado = %d, se esperaba %d", saved.SprintID, sprintID)
	}
}

// CA-11.2 y RN-11.3
func TestAssignItem_ItemYaAsignadoAOtroSprintDaErrItemAlreadyAssigned(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	itemID := s.createAssignItem(t, projectID, "Login de usuarios")
	first := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	second := s.saveAssignSprint(t, projectID, "Sprint 2", StatusOpen)

	if _, err := s.manager.AssignItem(first, itemID); err != nil {
		t.Fatalf("la primera asignación falló: %v", err)
	}

	_, err := s.manager.AssignItem(second, itemID)

	if !errors.Is(err, backlog.ErrItemAlreadyAssigned) {
		t.Errorf("error = %v, se esperaba ErrItemAlreadyAssigned", err)
	}
	if saved := s.savedAssignItem(t, itemID); saved.SprintID != first {
		t.Errorf("SprintID guardado = %d, se esperaba %d", saved.SprintID, first)
	}
}

// CA-11.3 y RN-11.1
func TestAssignItem_SprintCerradoDaErrSprintClosed(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	itemID := s.createAssignItem(t, projectID, "Login de usuarios")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusClosed)

	assigned, err := s.manager.AssignItem(sprintID, itemID)

	if !errors.Is(err, ErrSprintClosed) {
		t.Errorf("error = %v, se esperaba ErrSprintClosed", err)
	}
	if assigned.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
	}
	if saved := s.savedAssignItem(t, itemID); saved.SprintID != 0 {
		t.Errorf("SprintID guardado = %d, se esperaba 0", saved.SprintID)
	}
}

// CA-11.3 y RN-11.2
func TestAssignItem_ItemDeOtroProyectoDaErrItemFromAnotherProject(t *testing.T) {
	s := newAssignStack()
	projectA := s.createAssignProject(t, "Sistema de ventas")
	projectB := s.createAssignProject(t, "Portal de clientes")
	itemID := s.createAssignItem(t, projectB, "Login de clientes")
	sprintID := s.saveAssignSprint(t, projectA, "Sprint 1", StatusOpen)

	assigned, err := s.manager.AssignItem(sprintID, itemID)

	if !errors.Is(err, ErrItemFromAnotherProject) {
		t.Errorf("error = %v, se esperaba ErrItemFromAnotherProject", err)
	}
	if assigned.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
	}
	if saved := s.savedAssignItem(t, itemID); saved.SprintID != 0 {
		t.Errorf("SprintID guardado = %d, se esperaba 0", saved.SprintID)
	}
}

// CA-11.4 y RN-11.1
func TestAssignItem_SprintInexistenteDaErrSprintNotFound(t *testing.T) {
	cases := map[string]int{"ID grande": 999, "ID cero": 0, "ID negativo": -1}
	for name, sprintID := range cases {
		t.Run(name, func(t *testing.T) {
			s := newAssignStack()
			projectID := s.createAssignProject(t, "Sistema de ventas")
			itemID := s.createAssignItem(t, projectID, "Login de usuarios")
			s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

			assigned, err := s.manager.AssignItem(sprintID, itemID)

			if !errors.Is(err, ErrSprintNotFound) {
				t.Errorf("error = %v, se esperaba ErrSprintNotFound", err)
			}
			if assigned.ID != 0 {
				t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
			}
			if saved := s.savedAssignItem(t, itemID); saved.SprintID != 0 {
				t.Errorf("SprintID guardado = %d, se esperaba 0", saved.SprintID)
			}
		})
	}
}

// CA-11.4
func TestAssignItem_SinSprintsCreadosDaErrSprintNotFound(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	itemID := s.createAssignItem(t, projectID, "Login de usuarios")

	if _, err := s.manager.AssignItem(1, itemID); !errors.Is(err, ErrSprintNotFound) {
		t.Errorf("error = %v, se esperaba ErrSprintNotFound", err)
	}
}

// CA-11.4
func TestAssignItem_ItemInexistenteDaErrItemNotFound(t *testing.T) {
	cases := map[string]int{"ID grande": 999, "ID cero": 0, "ID negativo": -1}
	for name, itemID := range cases {
		t.Run(name, func(t *testing.T) {
			s := newAssignStack()
			projectID := s.createAssignProject(t, "Sistema de ventas")
			s.createAssignItem(t, projectID, "Login de usuarios")
			sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

			assigned, err := s.manager.AssignItem(sprintID, itemID)

			if !errors.Is(err, backlog.ErrItemNotFound) {
				t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
			}
			if assigned.ID != 0 {
				t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
			}
		})
	}
}

// CA-11.4
func TestAssignItem_SinItemsCreadosDaErrItemNotFound(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

	if _, err := s.manager.AssignItem(sprintID, 1); !errors.Is(err, backlog.ErrItemNotFound) {
		t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
	}
}

// RN-11.6
func TestAssignItem_ItemEnCursoOTerminadoSeAsigna(t *testing.T) {
	for name, status := range map[string]backlog.Status{
		"EN_CURSO":  backlog.StatusInProgress,
		"TERMINADO": backlog.StatusFinished,
	} {
		t.Run(name, func(t *testing.T) {
			s := newAssignStack()
			projectID := s.createAssignProject(t, "Sistema de ventas")
			item, err := s.itemRepo.Save(backlog.Item{
				ProjectID:          projectID,
				Title:              "Login de usuarios",
				Priority:           backlog.PriorityHigh,
				AcceptanceCriteria: []string{"Criterio OK"},
				Status:             status,
			})
			if err != nil {
				t.Fatalf("guardar el ítem de prueba: %v", err)
			}
			sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

			assigned, err := s.manager.AssignItem(sprintID, item.ID)

			if err != nil {
				t.Fatalf("AssignItem() devolvió un error inesperado: %v", err)
			}
			if assigned.SprintID != sprintID {
				t.Errorf("SprintID devuelto = %d, se esperaba %d", assigned.SprintID, sprintID)
			}
			if saved := s.savedAssignItem(t, item.ID); saved.SprintID != sprintID {
				t.Errorf("SprintID guardado = %d, se esperaba %d", saved.SprintID, sprintID)
			}
		})
	}
}

// CA-11.5
func TestAssignItem_NoCambiaNingunOtroCampoNiElSprint(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	created, err := s.itemRepo.Save(backlog.Item{
		ProjectID:          projectID,
		Title:              "Login de usuarios",
		Description:        "Formulario de acceso",
		Priority:           backlog.PriorityHigh,
		AcceptanceCriteria: []string{"Login OK", "Carga correcta"},
		Status:             backlog.StatusInProgress,
		StoryPoints:        5,
		EstimatedHours:     2.5,
	})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

	if _, err := s.manager.AssignItem(sprintID, created.ID); err != nil {
		t.Fatalf("AssignItem() devolvió un error inesperado: %v", err)
	}

	assigned := s.savedAssignItem(t, created.ID)
	assigned.SprintID = created.SprintID
	if assigned != created {
		t.Errorf("ítem guardado cambió: %+v, se esperaba %+v", assigned, created)
	}
	saved := savedSprint(t, s.repo, sprintID)
	if saved.Status != StatusOpen || saved.Name != "Sprint 1" || saved.Goal != validGoal {
		t.Errorf("el sprint guardado cambió: %+v", saved)
	}
	if saved.ProjectID != projectID || !saved.StartDate.Equal(validStart) || !saved.EndDate.Equal(validEnd) {
		t.Errorf("las fechas o el proyecto del sprint cambiaron: %+v", saved)
	}
}

// CA-11.7
func TestAssignItem_ModificarElItemDevueltoNoAlteraElGuardado(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	itemID := s.createAssignItem(t, projectID, "Login de usuarios")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

	assigned, err := s.manager.AssignItem(sprintID, itemID)
	if err != nil {
		t.Fatalf("AssignItem() devolvió un error inesperado: %v", err)
	}

	assigned.Title = "Otro título"
	assigned.AcceptanceCriteria[0] = "Otro criterio"
	assigned.SprintID = 999

	saved := s.savedAssignItem(t, itemID)
	if saved.Title != "Login de usuarios" {
		t.Errorf("título guardado = %q, se esperaba %q", saved.Title, "Login de usuarios")
	}
	if want := []string{"Criterio OK"}; len(saved.AcceptanceCriteria) != 1 || saved.AcceptanceCriteria[0] != want[0] {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, want)
	}
	if saved.SprintID != sprintID {
		t.Errorf("SprintID guardado = %d, se esperaba %d", saved.SprintID, sprintID)
	}
}

// CA-11.6 y RN-11.3
func TestAssignItem_AsignacionesSimultaneasDelMismoItemDejanUnaSola(t *testing.T) {
	const attempts = 50
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	itemID := s.createAssignItem(t, projectID, "Login de usuarios")
	first := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	second := s.saveAssignSprint(t, projectID, "Sprint 2", StatusOpen)

	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sprintID := first
			if i%2 == 1 {
				sprintID = second
			}
			_, errs[i] = s.manager.AssignItem(sprintID, itemID)
		}()
	}
	wg.Wait()

	successes, alreadyAssigned := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, backlog.ErrItemAlreadyAssigned):
			alreadyAssigned++
		}
	}
	if successes != 1 {
		t.Errorf("tuvieron éxito %d asignaciones, se esperaba exactamente 1", successes)
	}
	if alreadyAssigned != attempts-1 {
		t.Errorf("fallaron %d asignaciones con ErrItemAlreadyAssigned, se esperaban %d", alreadyAssigned, attempts-1)
	}
	saved := s.savedAssignItem(t, itemID)
	if saved.SprintID != first && saved.SprintID != second {
		t.Errorf("SprintID guardado = %d, se esperaba %d o %d", saved.SprintID, first, second)
	}
}

// CA-11.6
func TestAssignItem_AsignacionesSimultaneasDeItemsDistintosTodasExitosas(t *testing.T) {
	const items = 10
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	itemIDs := make([]int, items)
	for i := range items {
		itemIDs[i] = s.createAssignItem(t, projectID, fmt.Sprintf("Ítem %d", i+1))
	}
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

	errs := make([]error, items)
	var wg sync.WaitGroup
	for i := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = s.manager.AssignItem(sprintID, itemIDs[i])
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("la asignación del ítem %d falló: %v", itemIDs[i], err)
		}
	}
	for _, itemID := range itemIDs {
		if saved := s.savedAssignItem(t, itemID); saved.SprintID != sprintID {
			t.Errorf("SprintID guardado del ítem %d = %d, se esperaba %d", itemID, saved.SprintID, sprintID)
		}
	}
}

// CA-11.8
func TestAssignItem_ErrorDelRepositorioDeSprintsSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	manager := NewManager(failingRepository{getByIDErr: errRepository}, projects, items)

	assigned, err := manager.AssignItem(1, 1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if assigned.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
	}
}

// CA-11.8
func TestAssignItem_ErrorDelRepositorioDeItemsSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(failingItemRepository{getErr: errRepository}, projects)
	repo := NewMemoryRepository()
	manager := NewManager(repo, projects, items)
	if _, err := repo.Save(Sprint{ProjectID: 1, Name: "Sprint 1", StartDate: validStart, EndDate: validEnd, Goal: validGoal, Status: StatusOpen}); err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}

	assigned, err := manager.AssignItem(1, 1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if assigned.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
	}
}
