package defect

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

var errDefectRepository = errors.New("fallo del repositorio de defectos")

// testDefectRepository es un repositorio en memoria para los tests. No es el
// repositorio de producción: solo permite observar lo que se intenta guardar.
type testDefectRepository struct {
	mu      sync.Mutex
	defects []Defect
	lastID  int
}

func (r *testDefectRepository) Save(defect Defect) (Defect, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastID++
	defect.ID = r.lastID
	r.defects = append(r.defects, defect)
	return defect, nil
}

func (r *testDefectRepository) snapshot() []Defect {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Defect(nil), r.defects...)
}

func (r *testDefectRepository) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.defects)
}

type failingDefectRepository struct {
	err error
}

func (r failingDefectRepository) Save(Defect) (Defect, error) {
	return Defect{}, r.err
}

type defectEventLog struct {
	mu     sync.Mutex
	events []string
}

func (l *defectEventLog) add(event string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, event)
}

func (l *defectEventLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.events...)
}

type recordingSprintRepository struct {
	sprint.Repository
	log *defectEventLog
}

func (r *recordingSprintRepository) Update(sp sprint.Sprint) (sprint.Sprint, error) {
	updated, err := r.Repository.Update(sp)
	if err == nil && sp.Status == sprint.StatusClosed {
		r.log.add("close")
	}
	return updated, err
}

type recordingDefectRepository struct {
	*testDefectRepository
	log *defectEventLog
}

func (r *recordingDefectRepository) Save(defect Defect) (Defect, error) {
	saved, err := r.testDefectRepository.Save(defect)
	if err == nil {
		r.log.add("defect")
	}
	return saved, err
}

// defectEnv es la pila de HU-29 sobre repositorios en memoria.
type defectEnv struct {
	projectID  int
	projects   *project.Manager
	itemRepo   *backlog.MemoryRepository
	items      *backlog.Manager
	sprintRepo *sprint.MemoryRepository
	sprints    *sprint.Manager
	repo       *testDefectRepository
	manager    *Manager
}

func newDefectEnv(t *testing.T) *defectEnv {
	t.Helper()

	projects := project.NewManager(project.NewMemoryRepository())
	p, err := projects.Create("Sistema de ventas", "")
	if err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	itemRepo := backlog.NewMemoryRepository()
	items := backlog.NewManager(itemRepo, projects)
	sprintRepo := sprint.NewMemoryRepository()
	sprints := sprint.NewManager(sprintRepo, projects, items)
	repo := &testDefectRepository{}

	return &defectEnv{
		projectID:  p.ID,
		projects:   projects,
		itemRepo:   itemRepo,
		items:      items,
		sprintRepo: sprintRepo,
		sprints:    sprints,
		repo:       repo,
		manager:    NewManager(repo, items, sprints),
	}
}

func (e *defectEnv) addSprint(t *testing.T, name string, status sprint.Status) sprint.Sprint {
	t.Helper()
	return e.addSprintForProject(t, e.projectID, name, status)
}

func (e *defectEnv) addSprintForProject(t *testing.T, projectID int, name string, status sprint.Status) sprint.Sprint {
	t.Helper()
	saved, err := e.sprintRepo.Save(sprint.Sprint{
		ProjectID: projectID,
		Name:      name,
		StartDate: time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC),
		Goal:      "Entregar el MVP",
		Status:    status,
	})
	if err != nil {
		t.Fatalf("guardar el sprint de prueba: %v", err)
	}
	return saved
}

func (e *defectEnv) addItem(t *testing.T, title string, status backlog.Status, sprintID int) backlog.Item {
	t.Helper()
	saved, err := e.itemRepo.Save(backlog.Item{
		ProjectID:          e.projectID,
		Title:              title,
		Description:        "Formulario de acceso",
		Priority:           backlog.PriorityHigh,
		AcceptanceCriteria: []string{"Login OK"},
		Status:             status,
		StoryPoints:        5,
		EstimatedHours:     2.5,
		SprintID:           sprintID,
	})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}
	return saved
}

func envForDefectRegister(t *testing.T) (*defectEnv, backlog.Item, sprint.Sprint) {
	t.Helper()
	env := newDefectEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, 0)
	return env, item, sp
}

func assertDefectError(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, se esperaba %v", err, want)
	}
}

func assertNoDefects(t *testing.T, env *defectEnv) {
	t.Helper()
	if got := env.repo.count(); got != 0 {
		t.Fatalf("defectos guardados = %d, se esperaba ninguno", got)
	}
}

func sameItem(a, b backlog.Item) bool {
	return a.ID == b.ID && a.ProjectID == b.ProjectID && a.Title == b.Title &&
		a.Description == b.Description && a.Priority == b.Priority && a.Status == b.Status &&
		a.StoryPoints == b.StoryPoints && a.EstimatedHours == b.EstimatedHours &&
		a.SprintID == b.SprintID && slices.Equal(a.AcceptanceCriteria, b.AcceptanceCriteria)
}

func TestRegister_Exitoso(t *testing.T) {
	env, item, sp := envForDefectRegister(t)

	got, err := env.manager.Register(item.ID, "  El total no descuenta el descuento  ", " alta ", sp.ID)
	if err != nil {
		t.Fatalf("Register devolvió un error inesperado: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", got.ID)
	}
	if got.Description != "El total no descuenta el descuento" {
		t.Errorf("descripción = %q, se esperaba recortada", got.Description)
	}
	if got.Severity != SeverityHigh {
		t.Errorf("severidad = %q, se esperaba %q", got.Severity, SeverityHigh)
	}
	if got.Status != StatusOpen {
		t.Errorf("estado = %q, se esperaba %q", got.Status, StatusOpen)
	}
	if got.ItemID != item.ID || got.DetectionSprintID != sp.ID || got.ResolutionSprintID != 0 {
		t.Errorf("defecto = %+v", got)
	}
	saved := env.repo.snapshot()
	if len(saved) != 1 {
		t.Fatalf("defectos guardados = %d, se esperaba 1", len(saved))
	}
	if saved[0] != got {
		t.Errorf("guardado = %+v, devuelto = %+v", saved[0], got)
	}
}

func TestRegister_HistoriaTerminada(t *testing.T) {
	env := newDefectEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusFinished, 0)
	before, err := env.items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem antes de registrar: %v", err)
	}

	if _, err := env.manager.Register(item.ID, "Falla al cerrar sesión", "MEDIA", sp.ID); err != nil {
		t.Fatalf("Register sobre una historia TERMINADO devolvió un error: %v", err)
	}

	after, err := env.items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem después de registrar: %v", err)
	}
	if !sameItem(after, before) {
		t.Errorf("la historia cambió: antes %+v, después %+v", before, after)
	}
}

func TestRegister_HistoriaDeOtroSprint(t *testing.T) {
	env := newDefectEnv(t)
	sp1 := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	sp2 := env.addSprint(t, "Sprint 2", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, sp2.ID)

	if _, err := env.manager.Register(item.ID, "Falla al iniciar sesión", "ALTA", sp1.ID); err != nil {
		t.Fatalf("Register con una historia de otro sprint devolvió un error: %v", err)
	}
	after, err := env.items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem después de registrar: %v", err)
	}
	if after.SprintID != sp2.ID {
		t.Errorf("sprint de la historia = %d, se esperaba %d", after.SprintID, sp2.ID)
	}
}

func TestRegister_DescripcionObligatoria(t *testing.T) {
	for _, description := range []string{"", "   "} {
		t.Run(fmt.Sprintf("descripción %q", description), func(t *testing.T) {
			env, item, sp := envForDefectRegister(t)

			_, err := env.manager.Register(item.ID, description, "ALTA", sp.ID)

			assertDefectError(t, err, ErrEmptyDescription)
			assertNoDefects(t, env)
		})
	}
}

func TestRegister_SeveridadInvalida(t *testing.T) {
	for _, severity := range []string{"", "   ", "URGENTE", "BLOQUEANTE"} {
		t.Run("severidad "+strconv.Quote(severity), func(t *testing.T) {
			env, item, sp := envForDefectRegister(t)

			_, err := env.manager.Register(item.ID, "El total no descuenta el descuento", severity, sp.ID)

			assertDefectError(t, err, ErrInvalidSeverity)
			assertNoDefects(t, env)
		})
	}
}

func TestRegister_SprintDeDeteccionObligatorio(t *testing.T) {
	for _, sprintID := range []int{0, -1} {
		t.Run("sprint "+strconv.Itoa(sprintID), func(t *testing.T) {
			env, item, _ := envForDefectRegister(t)

			_, err := env.manager.Register(item.ID, "El total no descuenta el descuento", "ALTA", sprintID)

			assertDefectError(t, err, ErrMissingDetectionSprint)
			assertNoDefects(t, env)
		})
	}
}

func TestRegister_SprintDeDeteccionInexistente(t *testing.T) {
	env, item, _ := envForDefectRegister(t)

	_, err := env.manager.Register(item.ID, "El total no descuenta el descuento", "ALTA", 999)

	assertDefectError(t, err, sprint.ErrSprintNotFound)
	assertNoDefects(t, env)
}

func TestRegister_SprintCerrado(t *testing.T) {
	env := newDefectEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusClosed)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, 0)

	_, err := env.manager.Register(item.ID, "El total no descuenta el descuento", "ALTA", sp.ID)

	assertDefectError(t, err, sprint.ErrSprintClosed)
	assertNoDefects(t, env)
}

func TestRegister_SprintDeOtroProyecto(t *testing.T) {
	env, item, _ := envForDefectRegister(t)
	other, err := env.projects.Create("Sistema de facturación", "")
	if err != nil {
		t.Fatalf("crear el segundo proyecto: %v", err)
	}
	sp := env.addSprintForProject(t, other.ID, "Sprint 1", sprint.StatusOpen)

	_, err = env.manager.Register(item.ID, "El total no descuenta el descuento", "ALTA", sp.ID)

	assertDefectError(t, err, ErrDetectionSprintFromAnotherProject)
	assertNoDefects(t, env)
}

func TestRegister_HistoriaInexistente(t *testing.T) {
	env, _, sp := envForDefectRegister(t)

	_, err := env.manager.Register(999, "El total no descuenta el descuento", "ALTA", sp.ID)

	assertDefectError(t, err, backlog.ErrItemNotFound)
	assertNoDefects(t, env)
}

func TestRegister_OrdenDeValidacion(t *testing.T) {
	cases := map[string]struct {
		itemID      int
		description string
		severity    string
		sprintID    int
		want        error
	}{
		"descripción antes que sprint": {1, "   ", "ALTA", 0, ErrEmptyDescription},
		"severidad antes que sprint":   {1, "El total no descuenta el descuento", "URGENTE", 0, ErrInvalidSeverity},
		"sprint antes que historia":    {999, "El total no descuenta el descuento", "ALTA", 0, ErrMissingDetectionSprint},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env, _, _ := envForDefectRegister(t)

			_, err := env.manager.Register(tc.itemID, tc.description, tc.severity, tc.sprintID)

			assertDefectError(t, err, tc.want)
			assertNoDefects(t, env)
		})
	}
}

func TestRegister_NoModificaHistoriaNiSprint(t *testing.T) {
	env, item, sp := envForDefectRegister(t)
	itemBefore, err := env.items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem antes de registrar: %v", err)
	}
	sprintBefore, err := env.sprintRepo.GetByID(sp.ID)
	if err != nil {
		t.Fatalf("GetByID del sprint antes de registrar: %v", err)
	}

	if _, err := env.manager.Register(item.ID, "El total no descuenta el descuento", "ALTA", sp.ID); err != nil {
		t.Fatalf("Register devolvió un error inesperado: %v", err)
	}

	itemAfter, err := env.items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem después de registrar: %v", err)
	}
	if !sameItem(itemAfter, itemBefore) {
		t.Errorf("la historia cambió: antes %+v, después %+v", itemBefore, itemAfter)
	}
	sprintAfter, err := env.sprintRepo.GetByID(sp.ID)
	if err != nil {
		t.Fatalf("GetByID del sprint después de registrar: %v", err)
	}
	if sprintAfter != sprintBefore {
		t.Errorf("el sprint cambió: antes %+v, después %+v", sprintBefore, sprintAfter)
	}
}

func TestRegister_CopiaIndependiente(t *testing.T) {
	env, item, sp := envForDefectRegister(t)

	got, err := env.manager.Register(item.ID, "El total no descuenta el descuento", "ALTA", sp.ID)
	if err != nil {
		t.Fatalf("Register devolvió un error inesperado: %v", err)
	}
	got.Description = "Otra descripción"
	got.Severity = SeverityLow
	got.Status = StatusResolved
	got.ItemID = 999
	got.DetectionSprintID = 999
	got.ResolutionSprintID = 999

	saved := env.repo.snapshot()[0]
	if saved.Description != "El total no descuenta el descuento" || saved.Severity != SeverityHigh ||
		saved.Status != StatusOpen || saved.ItemID != item.ID || saved.DetectionSprintID != sp.ID ||
		saved.ResolutionSprintID != 0 {
		t.Errorf("el defecto guardado cambió al modificar el devuelto: %+v", saved)
	}
}

func TestRegister_ErrorDelRepositorio(t *testing.T) {
	env, item, sp := envForDefectRegister(t)
	env.manager = NewManager(failingDefectRepository{err: errDefectRepository}, env.items, env.sprints)

	_, err := env.manager.Register(item.ID, "El total no descuenta el descuento", "ALTA", sp.ID)

	assertDefectError(t, err, errDefectRepository)
	assertNoDefects(t, env)
}

func TestRegister_Concurrente(t *testing.T) {
	env, item, sp := envForDefectRegister(t)

	const n = 50
	errs := make([]error, n)
	ids := make([]int, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			got, err := env.manager.Register(item.ID, "El total no descuenta el descuento", "ALTA", sp.ID)
			errs[i] = err
			ids[i] = got.ID
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("registro %d falló: %v", i, err)
		}
	}
	if got := env.repo.count(); got != n {
		t.Fatalf("defectos guardados = %d, se esperaban %d", got, n)
	}
	seen := map[int]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			t.Fatalf("ID inválido o repetido: %d en %v", id, ids)
		}
		seen[id] = true
	}
}

func TestRegister_ConcurrenteConCierre(t *testing.T) {
	for i := 0; i < 20; i++ {
		t.Run("intento "+strconv.Itoa(i), func(t *testing.T) {
			env, item, sp := envForDefectRegister(t)
			log := &defectEventLog{}
			sprintRepo := &recordingSprintRepository{Repository: env.sprintRepo, log: log}
			defectRepo := &recordingDefectRepository{testDefectRepository: env.repo, log: log}
			sprintManager := sprint.NewManager(sprintRepo, env.projects, env.items)
			manager := NewManager(defectRepo, env.items, sprintManager)

			start := make(chan struct{})
			var wg sync.WaitGroup
			var registerErr, closeErr error
			wg.Add(2)
			go func() {
				defer wg.Done()
				<-start
				_, registerErr = manager.Register(item.ID, "El total no descuenta el descuento", "ALTA", sp.ID)
			}()
			go func() {
				defer wg.Done()
				<-start
				_, closeErr = sprintManager.Close(sp.ID)
			}()
			close(start)
			wg.Wait()

			if closeErr != nil {
				t.Fatalf("Close falló: %v", closeErr)
			}
			closed, err := env.sprintRepo.GetByID(sp.ID)
			if err != nil {
				t.Fatalf("GetByID del sprint: %v", err)
			}
			if closed.Status != sprint.StatusClosed {
				t.Fatalf("estado del sprint = %q, se esperaba CERRADO", closed.Status)
			}
			if registerErr != nil && !errors.Is(registerErr, sprint.ErrSprintClosed) {
				t.Fatalf("Register falló con un error inesperado: %v", registerErr)
			}
			if registerErr == nil && env.repo.count() != 1 {
				t.Fatalf("Register terminó bien pero hay %d defectos", env.repo.count())
			}
			if registerErr != nil && env.repo.count() != 0 {
				t.Fatalf("Register falló pero hay %d defectos guardados", env.repo.count())
			}
			events := log.snapshot()
			closeIndex := -1
			for j, event := range events {
				if event == "close" {
					closeIndex = j
					break
				}
			}
			if closeIndex == -1 {
				t.Fatalf("no se registró el cierre: %v", events)
			}
			for _, event := range events[closeIndex+1:] {
				if event == "defect" {
					t.Fatalf("se guardó un defecto después del cierre: %v", events)
				}
			}
		})
	}
}
