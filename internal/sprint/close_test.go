package sprint

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-13 · Cerrar sprint (Issue #27)

// closeEventLog registra el orden de las escrituras de los repositorios
// instrumentados que usa el test de concurrencia de CA-13.8.
type closeEventLog struct {
	mu     sync.Mutex
	events []string
}

func (l *closeEventLog) add(event string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, event)
}

func (l *closeEventLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.events...)
}

// recordingSprintRepository registra la escritura de cierre de Close.
type recordingSprintRepository struct {
	Repository
	log *closeEventLog
}

func (r *recordingSprintRepository) Update(sprint Sprint) (Sprint, error) {
	saved, err := r.Repository.Update(sprint)
	if err == nil {
		r.log.add("close")
	}
	return saved, err
}

// failingUpdateSprintRepository permite que la lectura funcione y falla en la
// escritura de cierre.
type failingUpdateSprintRepository struct {
	Repository
	updateErr error
}

func (r *failingUpdateSprintRepository) Update(Sprint) (Sprint, error) {
	return Sprint{}, r.updateErr
}

// recordingItemRepository registra las escrituras de ítems. La demora antes
// de escribir agranda la ventana de la carrera y hace que el test de CA-13.8
// sea determinista si la implementación no sostiene el mutex del sprint.
type recordingItemRepository struct {
	*backlog.MemoryRepository
	log   *closeEventLog
	delay time.Duration
}

func (r *recordingItemRepository) Update(item backlog.Item) (backlog.Item, error) {
	if r.delay > 0 {
		time.Sleep(r.delay)
	}
	saved, err := r.MemoryRepository.Update(item)
	if err == nil {
		r.log.add("item")
	}
	return saved, err
}

// CA-13.1 y RN-13.4
func TestClose_SprintAbiertoQuedaCerrado(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, validName, StatusOpen)

	closed, err := s.manager.Close(sprintID)

	if err != nil {
		t.Fatalf("Close() devolvió un error inesperado: %v", err)
	}
	if closed.Status != StatusClosed {
		t.Errorf("estado devuelto = %q, se esperaba %q", closed.Status, StatusClosed)
	}
	if closed.ID != sprintID || closed.ProjectID != projectID {
		t.Errorf("identidad devuelta = %+v, se esperaba el sprint %d del proyecto %d", closed, sprintID, projectID)
	}
	if closed.Name != validName || closed.Goal != validGoal {
		t.Errorf("texto devuelto = %q/%q, se esperaba %q/%q", closed.Name, closed.Goal, validName, validGoal)
	}
	if !closed.StartDate.Equal(validStart) || !closed.EndDate.Equal(validEnd) {
		t.Errorf("fechas devueltas = %v a %v, se esperaban %v a %v", closed.StartDate, closed.EndDate, validStart, validEnd)
	}

	saved := savedSprint(t, s.repo, sprintID)
	if saved.Status != StatusClosed {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusClosed)
	}
	if saved.Name != validName || saved.Goal != validGoal || saved.ProjectID != projectID {
		t.Errorf("el sprint guardado cambió sus datos: %+v", saved)
	}
	if !saved.StartDate.Equal(validStart) || !saved.EndDate.Equal(validEnd) {
		t.Errorf("las fechas guardadas cambiaron: %v a %v", saved.StartDate, saved.EndDate)
	}
}

// CA-13.1
func TestClose_SprintSinItemsQuedaCerrado(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, validName, StatusOpen)

	if _, err := s.manager.Close(sprintID); err != nil {
		t.Fatalf("Close() devolvió un error inesperado: %v", err)
	}

	if saved := savedSprint(t, s.repo, sprintID); saved.Status != StatusClosed {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusClosed)
	}
}

// CA-13.1
func TestClose_DespuesDelCierreSePuedeCrearOtroSprint(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, validName, StatusOpen)

	if _, err := s.manager.Close(sprintID); err != nil {
		t.Fatalf("Close() devolvió un error inesperado: %v", err)
	}

	created, err := s.manager.Create(projectID, "Sprint 2", validEnd, validEnd.AddDate(0, 0, 7), validGoal)
	if err != nil {
		t.Fatalf("Create() después del cierre devolvió un error inesperado: %v", err)
	}
	if created.Status != StatusOpen {
		t.Errorf("estado del nuevo sprint = %q, se esperaba %q", created.Status, StatusOpen)
	}
	if created.ID == sprintID {
		t.Errorf("el nuevo sprint reutilizó el ID %d", sprintID)
	}
}

// CA-13.2 y RN-13.6
func TestClose_SprintCerradoNoAdmiteAsignarNiCompletar(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	inProgress := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusInProgress)
	pendingID := s.createAssignItem(t, projectID, "Listar proyectos")
	sprintID := s.saveAssignSprint(t, projectID, validName, StatusOpen)
	assignCompleteItem(t, s, inProgress.ID, sprintID)

	if _, err := s.manager.Close(sprintID); err != nil {
		t.Fatalf("Close() devolvió un error inesperado: %v", err)
	}

	if _, err := s.manager.AssignItem(sprintID, pendingID); !errors.Is(err, ErrSprintClosed) {
		t.Errorf("AssignItem() error = %v, se esperaba ErrSprintClosed", err)
	}
	if _, err := s.manager.CompleteItem(sprintID, inProgress.ID); !errors.Is(err, ErrSprintClosed) {
		t.Errorf("CompleteItem() error = %v, se esperaba ErrSprintClosed", err)
	}

	if saved := s.savedAssignItem(t, inProgress.ID); saved.Status != backlog.StatusInProgress {
		t.Errorf("estado del ítem completado = %q, se esperaba %q", saved.Status, backlog.StatusInProgress)
	}
	if saved := s.savedAssignItem(t, pendingID); saved.SprintID != 0 {
		t.Errorf("SprintID del ítem asignado = %d, se esperaba 0", saved.SprintID)
	}
}

// CA-13.3 y RN-13.3
func TestClose_SprintYaCerradoDaErrSprintClosed(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, validName, StatusClosed)

	closed, err := s.manager.Close(sprintID)

	if !errors.Is(err, ErrSprintClosed) {
		t.Errorf("error = %v, se esperaba ErrSprintClosed", err)
	}
	if closed != (Sprint{}) {
		t.Errorf("sprint devuelto = %+v, se esperaba un sprint vacío", closed)
	}
	if saved := savedSprint(t, s.repo, sprintID); saved.Status != StatusClosed {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusClosed)
	}
}

// CA-13.4 y RN-13.1
func TestClose_IDInexistenteDaErrSprintNotFound(t *testing.T) {
	cases := map[string]int{"ID grande": 999, "ID cero": 0, "ID negativo": -1}
	for name, sprintID := range cases {
		t.Run(name, func(t *testing.T) {
			s := newAssignStack()
			projectID := s.createAssignProject(t, "Sistema de ventas")
			s.saveAssignSprint(t, projectID, validName, StatusOpen)

			closed, err := s.manager.Close(sprintID)

			if !errors.Is(err, ErrSprintNotFound) {
				t.Errorf("error = %v, se esperaba ErrSprintNotFound", err)
			}
			if closed != (Sprint{}) {
				t.Errorf("sprint devuelto = %+v, se esperaba un sprint vacío", closed)
			}
		})
	}
}

// CA-13.5 y RN-13.8
func TestClose_ErrorDelRepositorioAlConsultarSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	manager := NewManager(failingRepository{getByIDErr: errRepository}, projects, items)

	closed, err := manager.Close(1)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
	if closed != (Sprint{}) {
		t.Errorf("sprint devuelto = %+v, se esperaba un sprint vacío", closed)
	}
}

// CA-13.5 y RN-13.8
func TestClose_ErrorDelRepositorioAlActualizarSeDevuelve(t *testing.T) {
	repo := &failingUpdateSprintRepository{Repository: NewMemoryRepository(), updateErr: errRepository}
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	manager := NewManager(repo, projects, items)
	saved, err := repo.Save(Sprint{
		ProjectID: 1,
		Name:      validName,
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    StatusOpen,
	})
	if err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}

	closed, err := manager.Close(saved.ID)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
	if closed != (Sprint{}) {
		t.Errorf("sprint devuelto = %+v, se esperaba un sprint vacío", closed)
	}
	if got, err := repo.GetByID(saved.ID); err != nil {
		t.Fatalf("consultar el sprint guardado: %v", err)
	} else if got.Status != StatusOpen {
		t.Errorf("estado guardado = %q, se esperaba %q", got.Status, StatusOpen)
	}
}

// CA-13.6 y RN-13.7
func TestClose_ModificarElSprintDevueltoNoAlteraElGuardado(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, validName, StatusOpen)

	closed, err := s.manager.Close(sprintID)
	if err != nil {
		t.Fatalf("Close() devolvió un error inesperado: %v", err)
	}

	closed.ID = 999
	closed.ProjectID = 999
	closed.Name = "Otro nombre"
	closed.Goal = "Otro objetivo"
	closed.Status = StatusOpen
	closed.StartDate = validStart.AddDate(0, 0, -7)
	closed.EndDate = validEnd.AddDate(0, 0, -7)

	saved := savedSprint(t, s.repo, sprintID)
	if saved.ID != sprintID || saved.ProjectID != projectID {
		t.Errorf("identidad guardada = %+v, se esperaba el sprint %d del proyecto %d", saved, sprintID, projectID)
	}
	if saved.Name != validName || saved.Goal != validGoal || saved.Status != StatusClosed {
		t.Errorf("el sprint guardado cambió: %+v", saved)
	}
	if !saved.StartDate.Equal(validStart) || !saved.EndDate.Equal(validEnd) {
		t.Errorf("las fechas guardadas cambiaron: %v a %v", saved.StartDate, saved.EndDate)
	}
}

// CA-13.7 y RN-13.10
func TestClose_SprintConItemsPendientesYEnCursoSeCierraSinModificarlos(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	pendingID := s.createAssignItem(t, projectID, "Listar proyectos")
	inProgress := saveCompleteItem(t, s, projectID, "Login de usuarios", backlog.StatusInProgress)
	sprintID := s.saveAssignSprint(t, projectID, validName, StatusOpen)
	assignCompleteItem(t, s, pendingID, sprintID)
	assignCompleteItem(t, s, inProgress.ID, sprintID)

	if _, err := s.manager.Close(sprintID); err != nil {
		t.Fatalf("Close() devolvió un error inesperado: %v", err)
	}

	for id, wantStatus := range map[int]backlog.Status{pendingID: backlog.StatusPending, inProgress.ID: backlog.StatusInProgress} {
		saved := s.savedAssignItem(t, id)
		if saved.Status != wantStatus {
			t.Errorf("estado del ítem %d = %q, se esperaba %q", id, saved.Status, wantStatus)
		}
		if saved.SprintID != sprintID {
			t.Errorf("SprintID del ítem %d = %d, se esperaba %d", id, saved.SprintID, sprintID)
		}
	}
}

// CA-13.7 y RN-13.10
func TestClose_ItemsNoTerminadosNoSePuedenReasignar(t *testing.T) {
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	pendingID := s.createAssignItem(t, projectID, "Listar proyectos")
	sprintID := s.saveAssignSprint(t, projectID, validName, StatusOpen)
	assignCompleteItem(t, s, pendingID, sprintID)

	if _, err := s.manager.Close(sprintID); err != nil {
		t.Fatalf("Close() devolvió un error inesperado: %v", err)
	}

	second, err := s.repo.Save(Sprint{
		ProjectID: projectID,
		Name:      "Sprint 2",
		StartDate: validEnd,
		EndDate:   validEnd.AddDate(0, 0, 7),
		Goal:      validGoal,
		Status:    StatusOpen,
	})
	if err != nil {
		t.Fatalf("guardar el segundo sprint: %v", err)
	}

	if _, err := s.manager.AssignItem(second.ID, pendingID); !errors.Is(err, backlog.ErrItemAlreadyAssigned) {
		t.Errorf("AssignItem() error = %v, se esperaba ErrItemAlreadyAssigned", err)
	}
	if saved := s.savedAssignItem(t, pendingID); saved.SprintID != sprintID {
		t.Errorf("SprintID del ítem = %d, se esperaba %d", saved.SprintID, sprintID)
	}
}

// CA-13.8
func TestClose_CierresSimultaneosDejanUnSoloCierreExitoso(t *testing.T) {
	const attempts = 20
	s := newAssignStack()
	projectID := s.createAssignProject(t, "Sistema de ventas")
	sprintID := s.saveAssignSprint(t, projectID, validName, StatusOpen)

	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = s.manager.Close(sprintID)
		}()
	}
	wg.Wait()

	successes := 0
	closedErrs := 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrSprintClosed):
			closedErrs++
		}
	}
	if successes != 1 {
		t.Errorf("tuvieron éxito %d cierres, se esperaba exactamente 1", successes)
	}
	if closedErrs != attempts-1 {
		t.Errorf("fallaron %d cierres con ErrSprintClosed, se esperaban %d", closedErrs, attempts-1)
	}
	if saved := savedSprint(t, s.repo, sprintID); saved.Status != StatusClosed {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusClosed)
	}
}

// CA-13.8 y RN-13.9
func TestClose_EnParaleloConAssignItemYCompleteItem(t *testing.T) {
	const itemCount = 10
	log := &closeEventLog{}
	baseSprintRepo := NewMemoryRepository()
	sprintRepo := &recordingSprintRepository{Repository: baseSprintRepo, log: log}
	itemRepo := &recordingItemRepository{
		MemoryRepository: backlog.NewMemoryRepository(),
		log:              log,
		delay:            time.Millisecond,
	}
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(itemRepo, projects)
	manager := NewManager(sprintRepo, projects, items)

	project, err := projects.Create("Sistema de ventas", "")
	if err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	sprint, err := sprintRepo.Save(Sprint{
		ProjectID: project.ID,
		Name:      validName,
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    StatusOpen,
	})
	if err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}

	assignIDs := make([]int, itemCount)
	completeIDs := make([]int, itemCount)
	for i := 0; i < itemCount; i++ {
		toAssign, err := itemRepo.Save(backlog.Item{
			ProjectID:          project.ID,
			Title:              fmt.Sprintf("Asignar %d", i+1),
			Priority:           backlog.PriorityHigh,
			AcceptanceCriteria: []string{"Criterio OK"},
			Status:             backlog.StatusPending,
		})
		if err != nil {
			t.Fatalf("guardar el ítem para asignar %d: %v", i+1, err)
		}
		assignIDs[i] = toAssign.ID

		toComplete, err := itemRepo.Save(backlog.Item{
			ProjectID:          project.ID,
			Title:              fmt.Sprintf("Completar %d", i+1),
			Priority:           backlog.PriorityHigh,
			AcceptanceCriteria: []string{"Criterio OK"},
			Status:             backlog.StatusInProgress,
			SprintID:           sprint.ID,
		})
		if err != nil {
			t.Fatalf("guardar el ítem para completar %d: %v", i+1, err)
		}
		completeIDs[i] = toComplete.ID
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	var closeErr error
	assignErrs := make([]error, itemCount)
	completeErrs := make([]error, itemCount)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		_, closeErr = manager.Close(sprint.ID)
	}()
	for i := 0; i < itemCount; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, assignErrs[i] = manager.AssignItem(sprint.ID, assignIDs[i])
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, completeErrs[i] = manager.CompleteItem(sprint.ID, completeIDs[i])
		}()
	}
	close(start)
	wg.Wait()

	if closeErr != nil {
		t.Fatalf("Close() en paralelo devolvió un error inesperado: %v", closeErr)
	}
	if saved := savedSprint(t, baseSprintRepo, sprint.ID); saved.Status != StatusClosed {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusClosed)
	}
	for i, err := range assignErrs {
		if err != nil && !errors.Is(err, ErrSprintClosed) {
			t.Errorf("AssignItem() %d error = %v, se esperaba nil o ErrSprintClosed", i+1, err)
		}
		item, err := items.GetItem(assignIDs[i])
		if err != nil {
			t.Fatalf("consultar el ítem asignado %d: %v", i+1, err)
		}
		if assignErrs[i] == nil {
			if item.SprintID != sprint.ID {
				t.Errorf("ítem asignado %d: SprintID = %d, se esperaba %d", i+1, item.SprintID, sprint.ID)
			}
		} else if item.SprintID != 0 {
			t.Errorf("ítem no asignado %d: SprintID = %d, se esperaba 0", i+1, item.SprintID)
		}
	}
	for i, err := range completeErrs {
		if err != nil && !errors.Is(err, ErrSprintClosed) {
			t.Errorf("CompleteItem() %d error = %v, se esperaba nil o ErrSprintClosed", i+1, err)
		}
		item, err := items.GetItem(completeIDs[i])
		if err != nil {
			t.Fatalf("consultar el ítem completado %d: %v", i+1, err)
		}
		if completeErrs[i] == nil {
			if item.Status != backlog.StatusFinished {
				t.Errorf("ítem completado %d: estado = %q, se esperaba %q", i+1, item.Status, backlog.StatusFinished)
			}
		} else if item.Status != backlog.StatusInProgress {
			t.Errorf("ítem no completado %d: estado = %q, se esperaba %q", i+1, item.Status, backlog.StatusInProgress)
		}
	}

	events := log.snapshot()
	closeIndex := -1
	for i, event := range events {
		if event == "close" {
			closeIndex = i
			break
		}
	}
	if closeIndex == -1 {
		t.Fatal("no se registró la escritura de cierre")
	}
	for _, event := range events[closeIndex+1:] {
		if event == "item" {
			t.Errorf("se escribió un ítem después del cierre: %v", events)
			break
		}
	}
}
