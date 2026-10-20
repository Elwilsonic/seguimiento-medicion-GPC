package sprint

import (
	"errors"
	"slices"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-12 · Registrar historias completadas (Issue #26)

// saveCompleteItem guarda un ítem directo en el repositorio con el estado
// recibido, para preparar los casos de HU-12 sin depender de las transiciones
// de HU-08.
func saveCompleteItem(t *testing.T, s assignStack, projectID int, title string, status backlog.Status) backlog.Item {
	t.Helper()
	item, err := s.itemRepo.Save(backlog.Item{
		ProjectID:          projectID,
		Title:              title,
		Description:        "Formulario de acceso",
		Priority:           backlog.PriorityHigh,
		AcceptanceCriteria: []string{"Login OK"},
		Status:             status,
		StoryPoints:        5,
		EstimatedHours:     2.5,
	})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}
	return item
}

// assignCompleteItem asigna el ítem al sprint con la operación de backlog de
// HU-11, sin validar el estado del sprint: los tests de HU-12 necesitan
// preparar ítems asignados a sprints cerrados.
func assignCompleteItem(t *testing.T, s assignStack, itemID, sprintID int) {
	t.Helper()
	if _, err := s.items.AssignToSprint(itemID, sprintID); err != nil {
		t.Fatalf("asignar el ítem %d al sprint %d: %v", itemID, sprintID, err)
	}
}

// countingItemRepository cuenta las escrituras para comprobar que un no-op no
// llama a Update.
type countingItemRepository struct {
	*backlog.MemoryRepository
	updates int
}

func (r *countingItemRepository) Update(item backlog.Item) (backlog.Item, error) {
	r.updates++
	return r.MemoryRepository.Update(item)
}

// failingUpdateItemRepository permite que la lectura funcione y que la
// escritura devuelva el error configurado.
type failingUpdateItemRepository struct {
	*backlog.MemoryRepository
	updateErr error
}

func (r *failingUpdateItemRepository) Update(item backlog.Item) (backlog.Item, error) {
	return backlog.Item{}, r.updateErr
}

// failingListItemRepository falla al listar los ítems del proyecto.
type failingListItemRepository struct {
	*backlog.MemoryRepository
	listErr error
}

func (r *failingListItemRepository) ListByProject(projectID int) ([]backlog.Item, error) {
	return nil, r.listErr
}

// CA-12.1 y RN-12.1
func TestCompleteItem_ItemEnCursoAsignadoQuedaTerminado(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusInProgress)
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	assignCompleteItem(t, s, item.ID, sprintID)

	completed, err := s.manager.CompleteItem(sprintID, item.ID)

	if err != nil {
		t.Fatalf("CompleteItem() devolvió un error inesperado: %v", err)
	}
	if completed.Status != backlog.StatusFinished {
		t.Errorf("estado devuelto = %q, se esperaba %q", completed.Status, backlog.StatusFinished)
	}
	if completed.ID != item.ID || completed.ProjectID != item.ProjectID {
		t.Errorf("identidad devuelta = %+v, se esperaba el ítem %d del proyecto %d", completed, item.ID, item.ProjectID)
	}
	if completed.Title != item.Title || completed.Description != item.Description {
		t.Errorf("texto devuelto = %q/%q, se esperaba %q/%q", completed.Title, completed.Description, item.Title, item.Description)
	}
	if completed.Priority != item.Priority || completed.StoryPoints != item.StoryPoints || completed.EstimatedHours != item.EstimatedHours {
		t.Errorf("campos devueltos = %+v, se esperaban los del ítem %+v", completed, item)
	}
	if completed.SprintID != sprintID {
		t.Errorf("SprintID devuelto = %d, se esperaba %d", completed.SprintID, sprintID)
	}
	if !slices.Equal(completed.AcceptanceCriteria, item.AcceptanceCriteria) {
		t.Errorf("criterios devueltos = %q, se esperaban %q", completed.AcceptanceCriteria, item.AcceptanceCriteria)
	}

	saved := s.savedAssignItem(t, item.ID)
	if saved.Status != backlog.StatusFinished {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, backlog.StatusFinished)
	}
	if saved.Title != item.Title || saved.Description != item.Description {
		t.Errorf("el ítem guardado cambió sus textos: %+v", saved)
	}
	if saved.SprintID != sprintID || saved.StoryPoints != item.StoryPoints || saved.EstimatedHours != item.EstimatedHours {
		t.Errorf("el ítem guardado cambió otros campos: %+v", saved)
	}
}

// CA-12.4 y RN-12.2
func TestCompleteItem_ItemYaTerminadoEsNoOpSinUpdate(t *testing.T) {
	projectRepo := project.NewMemoryRepository()
	projects := project.NewManager(projectRepo)
	projectCreated, err := projects.Create("Sistema de ventas", "")
	if err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	itemRepo := &countingItemRepository{MemoryRepository: backlog.NewMemoryRepository()}
	items := backlog.NewManager(itemRepo, projects)
	repo := NewMemoryRepository()
	manager := NewManager(repo, projects, items)
	sprint, err := repo.Save(Sprint{
		ProjectID: projectCreated.ID,
		Name:      "Sprint 1",
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    StatusOpen,
	})
	if err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}
	item, err := itemRepo.Save(backlog.Item{
		ProjectID:          projectCreated.ID,
		Title:              "Login de usuarios",
		AcceptanceCriteria: []string{"Login OK"},
		Status:             backlog.StatusFinished,
		SprintID:           sprint.ID,
	})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}

	completed, err := manager.CompleteItem(sprint.ID, item.ID)

	if err != nil {
		t.Fatalf("CompleteItem() devolvió un error inesperado: %v", err)
	}
	if completed.Status != backlog.StatusFinished {
		t.Errorf("estado devuelto = %q, se esperaba %q", completed.Status, backlog.StatusFinished)
	}
	if itemRepo.updates != 0 {
		t.Errorf("Update se llamó %d veces, se esperaba 0", itemRepo.updates)
	}
}

// CA-12.6 y RN-12.2
func TestCompleteItem_ItemPendienteDaErrInvalidTransition(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusPending)
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	assignCompleteItem(t, s, item.ID, sprintID)

	completed, err := s.manager.CompleteItem(sprintID, item.ID)

	if !errors.Is(err, backlog.ErrInvalidTransition) {
		t.Errorf("error = %v, se esperaba ErrInvalidTransition", err)
	}
	if completed.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", completed)
	}
	if saved := s.savedAssignItem(t, item.ID); saved.Status != backlog.StatusPending {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, backlog.StatusPending)
	}
}

// CA-12.2 y RN-12.3
func TestCompleteItem_ItemSinAsignarDaErrItemNotAssignedToSprint(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusInProgress)
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

	completed, err := s.manager.CompleteItem(sprintID, item.ID)

	if !errors.Is(err, ErrItemNotAssignedToSprint) {
		t.Errorf("error = %v, se esperaba ErrItemNotAssignedToSprint", err)
	}
	if completed.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", completed)
	}
	if saved := s.savedAssignItem(t, item.ID); saved.Status != backlog.StatusInProgress || saved.SprintID != 0 {
		t.Errorf("ítem guardado = %+v, se esperaba EN_CURSO y sin sprint", saved)
	}
}

// CA-12.2 y RN-12.3
func TestCompleteItem_ItemAsignadoAOtroSprintDaErrItemNotAssignedToSprint(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusInProgress)
	first := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	second := s.saveAssignSprint(t, projectID, "Sprint 2", StatusOpen)
	assignCompleteItem(t, s, item.ID, second)

	completed, err := s.manager.CompleteItem(first, item.ID)

	if !errors.Is(err, ErrItemNotAssignedToSprint) {
		t.Errorf("error = %v, se esperaba ErrItemNotAssignedToSprint", err)
	}
	if completed.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", completed)
	}
	saved := s.savedAssignItem(t, item.ID)
	if saved.Status != backlog.StatusInProgress || saved.SprintID != second {
		t.Errorf("ítem guardado = %+v, se esperaba EN_CURSO en el sprint %d", saved, second)
	}
}

// CA-12.5 y RN-12.4
func TestCompleteItem_SprintCerradoDaErrSprintClosed(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusInProgress)
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusClosed)
	assignCompleteItem(t, s, item.ID, sprintID)

	completed, err := s.manager.CompleteItem(sprintID, item.ID)

	if !errors.Is(err, ErrSprintClosed) {
		t.Errorf("error = %v, se esperaba ErrSprintClosed", err)
	}
	if completed.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", completed)
	}
	if saved := s.savedAssignItem(t, item.ID); saved.Status != backlog.StatusInProgress {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, backlog.StatusInProgress)
	}
}

// CA-12.5
func TestCompleteItem_SprintInexistenteDaErrSprintNotFound(t *testing.T) {
	cases := map[string]int{"ID grande": 999, "ID cero": 0, "ID negativo": -1}
	for name, sprintID := range cases {
		t.Run(name, func(t *testing.T) {
			s := newAssignStack()
			projectID := s.createAssignProject(t, "Sistema de ventas")
			item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusInProgress)
			s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

			completed, err := s.manager.CompleteItem(sprintID, item.ID)

			if !errors.Is(err, ErrSprintNotFound) {
				t.Errorf("error = %v, se esperaba ErrSprintNotFound", err)
			}
			if completed.ID != 0 {
				t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", completed)
			}
			if saved := s.savedAssignItem(t, item.ID); saved.Status != backlog.StatusInProgress {
				t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, backlog.StatusInProgress)
			}
		})
	}
}

// CA-12.5
func TestCompleteItem_ItemInexistenteDaErrItemNotFound(t *testing.T) {
	cases := map[string]int{"ID grande": 999, "ID cero": 0, "ID negativo": -1}
	for name, itemID := range cases {
		t.Run(name, func(t *testing.T) {
			s := newAssignStack()
			projectID := s.createAssignProject(t, "Sistema de ventas")
			sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)

			completed, err := s.manager.CompleteItem(sprintID, itemID)

			if !errors.Is(err, backlog.ErrItemNotFound) {
				t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
			}
			if completed.ID != 0 {
				t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", completed)
			}
		})
	}
}

// CA-12.8
func TestCompleteItem_ErrorDelRepositorioDeSprintsSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	manager := NewManager(failingRepository{getByIDErr: errRepository}, projects, items)

	completed, err := manager.CompleteItem(1, 1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if completed.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", completed)
	}
}

// CA-12.8
func TestCompleteItem_ErrorDelRepositorioDeItemsSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(failingItemRepository{getErr: errRepository}, projects)
	repo := NewMemoryRepository()
	manager := NewManager(repo, projects, items)
	if _, err := repo.Save(Sprint{ProjectID: 1, Name: "Sprint 1", StartDate: validStart, EndDate: validEnd, Goal: validGoal, Status: StatusOpen}); err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}

	completed, err := manager.CompleteItem(1, 1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if completed.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", completed)
	}
}

// CA-12.8
func TestCompleteItem_ErrorDelRepositorioAlActualizarSeDevuelve(t *testing.T) {
	projectRepo := project.NewMemoryRepository()
	projects := project.NewManager(projectRepo)
	projectCreated, err := projects.Create("Sistema de ventas", "")
	if err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	itemRepo := &failingUpdateItemRepository{
		MemoryRepository: backlog.NewMemoryRepository(),
		updateErr:        errRepository,
	}
	items := backlog.NewManager(itemRepo, projects)
	repo := NewMemoryRepository()
	manager := NewManager(repo, projects, items)
	sprint, err := repo.Save(Sprint{
		ProjectID: projectCreated.ID,
		Name:      "Sprint 1",
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    StatusOpen,
	})
	if err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}
	item, err := itemRepo.Save(backlog.Item{
		ProjectID:          projectCreated.ID,
		Title:              "Login de usuarios",
		AcceptanceCriteria: []string{"Login OK"},
		Status:             backlog.StatusInProgress,
		SprintID:           sprint.ID,
	})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}

	completed, err := manager.CompleteItem(sprint.ID, item.ID)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if completed.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", completed)
	}
}

// CA-12.7 y RN-12.8
func TestCompleteItem_ModificarElItemDevueltoNoAlteraElGuardado(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusInProgress)
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	assignCompleteItem(t, s, item.ID, sprintID)

	completed, err := s.manager.CompleteItem(sprintID, item.ID)
	if err != nil {
		t.Fatalf("CompleteItem() devolvió un error inesperado: %v", err)
	}

	completed.Title = "Otro título"
	completed.AcceptanceCriteria[0] = "Otro criterio"
	completed.SprintID = 999
	completed.Status = backlog.StatusPending

	saved := s.savedAssignItem(t, item.ID)
	if saved.Title != item.Title {
		t.Errorf("título guardado = %q, se esperaba %q", saved.Title, item.Title)
	}
	if !slices.Equal(saved.AcceptanceCriteria, item.AcceptanceCriteria) {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, item.AcceptanceCriteria)
	}
	if saved.SprintID != sprintID {
		t.Errorf("SprintID guardado = %d, se esperaba %d", saved.SprintID, sprintID)
	}
	if saved.Status != backlog.StatusFinished {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, backlog.StatusFinished)
	}
}

// CA-12.9
func TestCompleteItem_MarcacionesSimultaneasDejanUnSoloUpdate(t *testing.T) {
	const attempts = 50
	projectRepo := project.NewMemoryRepository()
	projects := project.NewManager(projectRepo)
	projectCreated, err := projects.Create("Sistema de ventas", "")
	if err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	itemRepo := &countingItemRepository{MemoryRepository: backlog.NewMemoryRepository()}
	items := backlog.NewManager(itemRepo, projects)
	repo := NewMemoryRepository()
	manager := NewManager(repo, projects, items)
	sprint, err := repo.Save(Sprint{
		ProjectID: projectCreated.ID,
		Name:      "Sprint 1",
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    StatusOpen,
	})
	if err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}
	item, err := itemRepo.Save(backlog.Item{
		ProjectID:          projectCreated.ID,
		Title:              "Login de usuarios",
		AcceptanceCriteria: []string{"Login OK"},
		Status:             backlog.StatusInProgress,
		SprintID:           sprint.ID,
	})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}

	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = manager.CompleteItem(sprint.ID, item.ID)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("la marcación %d falló: %v", i+1, err)
		}
	}
	if itemRepo.updates != 1 {
		t.Errorf("Update se llamó %d veces, se esperaba 1", itemRepo.updates)
	}
	saved, err := items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("consultar el ítem guardado: %v", err)
	}
	if saved.Status != backlog.StatusFinished {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, backlog.StatusFinished)
	}
}

// CA-12.3 y RN-12.6
func TestListCompletedItems_SoloDevuelveTerminadosDelSprintOrdenados(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	first := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	second := s.saveAssignSprint(t, projectID, "Sprint 2", StatusOpen)
	inProgress := saveCompleteItem(t, s, projectID, "En curso", backlog.StatusInProgress)
	finishedFirst := saveCompleteItem(t, s, projectID, "Terminado 1", backlog.StatusFinished)
	pending := saveCompleteItem(t, s, projectID, "Pendiente", backlog.StatusPending)
	finishedSecond := saveCompleteItem(t, s, projectID, "Terminado 2", backlog.StatusFinished)
	otherSprint := saveCompleteItem(t, s, projectID, "Otro sprint", backlog.StatusFinished)
	unassigned := saveCompleteItem(t, s, projectID, "Sin asignar", backlog.StatusFinished)
	assignCompleteItem(t, s, inProgress.ID, first)
	assignCompleteItem(t, s, finishedFirst.ID, first)
	assignCompleteItem(t, s, pending.ID, first)
	assignCompleteItem(t, s, finishedSecond.ID, first)
	assignCompleteItem(t, s, otherSprint.ID, second)

	completed, err := s.manager.ListCompletedItems(first)

	if err != nil {
		t.Fatalf("ListCompletedItems() devolvió un error inesperado: %v", err)
	}
	if len(completed) != 2 {
		t.Fatalf("cantidad devuelta = %d, se esperaban 2", len(completed))
	}
	wantIDs := []int{finishedFirst.ID, finishedSecond.ID}
	for i, item := range completed {
		if item.ID != wantIDs[i] {
			t.Errorf("ítem %d = ID %d, se esperaba %d", i+1, item.ID, wantIDs[i])
		}
	}
	for _, item := range completed {
		if item.ID == inProgress.ID || item.ID == pending.ID || item.ID == otherSprint.ID || item.ID == unassigned.ID {
			t.Errorf("la lista incluyó un ítem que no corresponde: %+v", item)
		}
	}
}

// CA-12.3
func TestListCompletedItems_SprintSinCompletadosDevuelveListaVaciaNoNil(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusInProgress)
	assignCompleteItem(t, s, item.ID, sprintID)

	completed, err := s.manager.ListCompletedItems(sprintID)

	if err != nil {
		t.Fatalf("ListCompletedItems() devolvió un error inesperado: %v", err)
	}
	if completed == nil {
		t.Fatal("la lista devuelta fue nil, se esperaba una lista vacía")
	}
	if len(completed) != 0 {
		t.Errorf("cantidad devuelta = %d, se esperaba 0", len(completed))
	}
}

// CA-12.3 y RN-12.4
func TestListCompletedItems_SprintCerradoSePuedeConsultar(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusClosed)
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusFinished)
	assignCompleteItem(t, s, item.ID, sprintID)

	completed, err := s.manager.ListCompletedItems(sprintID)

	if err != nil {
		t.Fatalf("ListCompletedItems() devolvió un error inesperado: %v", err)
	}
	if len(completed) != 1 || completed[0].ID != item.ID {
		t.Errorf("lista devuelta = %+v, se esperaba el ítem %d", completed, item.ID)
	}
}

// CA-12.5
func TestListCompletedItems_SprintInexistenteDaErrSprintNotFound(t *testing.T) {
	cases := map[string]int{"ID grande": 999, "ID cero": 0, "ID negativo": -1}
	for name, sprintID := range cases {
		t.Run(name, func(t *testing.T) {
			s := newAssignStack()

			completed, err := s.manager.ListCompletedItems(sprintID)

			if !errors.Is(err, ErrSprintNotFound) {
				t.Errorf("error = %v, se esperaba ErrSprintNotFound", err)
			}
			if completed != nil {
				t.Errorf("lista devuelta = %+v, se esperaba nil", completed)
			}
		})
	}
}

// CA-12.8
func TestListCompletedItems_ErrorDelRepositorioSeDevuelve(t *testing.T) {
	projectRepo := project.NewMemoryRepository()
	projects := project.NewManager(projectRepo)
	if _, err := projects.Create("Sistema de ventas", ""); err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	itemRepo := &failingListItemRepository{
		MemoryRepository: backlog.NewMemoryRepository(),
		listErr:          errRepository,
	}
	items := backlog.NewManager(itemRepo, projects)
	repo := NewMemoryRepository()
	manager := NewManager(repo, projects, items)
	if _, err := repo.Save(Sprint{ProjectID: 1, Name: "Sprint 1", StartDate: validStart, EndDate: validEnd, Goal: validGoal, Status: StatusOpen}); err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}

	completed, err := manager.ListCompletedItems(1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if completed != nil {
		t.Errorf("lista devuelta = %+v, se esperaba nil", completed)
	}
}

// CA-12.7 y RN-12.8
func TestListCompletedItems_ModificarLaListaDevueltaNoAlteraElGuardado(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusFinished)
	assignCompleteItem(t, s, item.ID, sprintID)

	completed, err := s.manager.ListCompletedItems(sprintID)
	if err != nil {
		t.Fatalf("ListCompletedItems() devolvió un error inesperado: %v", err)
	}
	if len(completed) != 1 {
		t.Fatalf("cantidad devuelta = %d, se esperaba 1", len(completed))
	}

	completed[0].Title = "Otro título"
	completed[0].AcceptanceCriteria[0] = "Otro criterio"

	saved := s.savedAssignItem(t, item.ID)
	if saved.Title != item.Title {
		t.Errorf("título guardado = %q, se esperaba %q", saved.Title, item.Title)
	}
	if !slices.Equal(saved.AcceptanceCriteria, item.AcceptanceCriteria) {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, item.AcceptanceCriteria)
	}
}

// RN-12.6
func TestListCompletedItems_ItemReabiertoNoAparece(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, "Sprint 1", StatusOpen)
	item := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusFinished)
	assignCompleteItem(t, s, item.ID, sprintID)

	completed, err := s.manager.ListCompletedItems(sprintID)
	if err != nil {
		t.Fatalf("ListCompletedItems() devolvió un error inesperado: %v", err)
	}
	if len(completed) != 1 {
		t.Fatalf("cantidad devuelta antes de reabrir = %d, se esperaba 1", len(completed))
	}

	if _, err := s.items.ChangeStatus(item.ID, string(backlog.StatusInProgress)); err != nil {
		t.Fatalf("reabrir el ítem: %v", err)
	}

	completed, err = s.manager.ListCompletedItems(sprintID)
	if err != nil {
		t.Fatalf("ListCompletedItems() después de reabrir devolvió un error: %v", err)
	}
	if len(completed) != 0 {
		t.Errorf("cantidad devuelta después de reabrir = %d, se esperaba 0", len(completed))
	}
}
