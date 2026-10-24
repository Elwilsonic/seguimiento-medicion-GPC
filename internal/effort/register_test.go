package effort

import (
	"errors"
	"math"
	"slices"
	"sort"
	"sync"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-26 · Registrar esfuerzo (Issue #44)
//
// Los tests usan los repositorios en memoria de proyectos, backlog, sprints y
// esfuerzos. El ítem y el sprint se guardan directamente para preparar cada
// caso sin depender de las validaciones de HU-10 a HU-13.

var testEffortDate = time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)

// effortEnv es la pila completa de HU-26.
type effortEnv struct {
	projectID  int
	projects   *project.Manager
	itemRepo   *backlog.MemoryRepository
	items      *backlog.Manager
	sprintRepo *sprint.MemoryRepository
	sprints    *sprint.Manager
	memory     *MemoryRepository
	manager    *Manager
}

// newEffortEnv arma un proyecto con un repositorio de esfuerzos en memoria.
func newEffortEnv(t *testing.T) *effortEnv {
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
	memory := NewMemoryRepository()

	return &effortEnv{
		projectID:  p.ID,
		projects:   projects,
		itemRepo:   itemRepo,
		items:      items,
		sprintRepo: sprintRepo,
		sprints:    sprints,
		memory:     memory,
		manager:    NewManager(memory, sprints, items),
	}
}

// addSprint guarda un sprint sin pasar por las validaciones de HU-10.
func (e *effortEnv) addSprint(t *testing.T, name string, status sprint.Status) sprint.Sprint {
	t.Helper()
	saved, err := e.sprintRepo.Save(sprint.Sprint{
		ProjectID: e.projectID,
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

// addItem guarda un ítem sin pasar por las validaciones de HU-04.
func (e *effortEnv) addItem(t *testing.T, title string, status backlog.Status, sprintID int) backlog.Item {
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

// envForRegister arma un entorno con un ítem EN_CURSO en el sprint 1 ABIERTO.
func envForRegister(t *testing.T) (*effortEnv, backlog.Item, sprint.Sprint) {
	t.Helper()
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, sp.ID)
	return env, item, sp
}

func TestRegister_Exitoso(t *testing.T) {
	env, item, sp := envForRegister(t)

	got, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar el registro", 2.5)
	if err != nil {
		t.Fatalf("Register devolvió un error inesperado: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", got.ID)
	}
	if got.Member != "García Santiago" || got.Activity != "Implementar el registro" {
		t.Errorf("textos = %q/%q", got.Member, got.Activity)
	}
	if !got.Date.Equal(testEffortDate) || got.ItemID != item.ID || got.SprintID != sp.ID || got.Hours != 2.5 {
		t.Errorf("esfuerzo = %+v", got)
	}
	if len(env.memory.efforts) != 1 {
		t.Fatalf("esfuerzos guardados = %d, se esperaba 1", len(env.memory.efforts))
	}
	if env.memory.efforts[0] != got {
		t.Errorf("guardado = %+v, devuelto = %+v", env.memory.efforts[0], got)
	}
}

func TestRegister_RecortaIntegranteYActividad(t *testing.T) {
	env, item, sp := envForRegister(t)

	got, err := env.manager.Register(sp.ID, item.ID, "  García Santiago  ", testEffortDate, "  Implementar el registro  ", 2)
	if err != nil {
		t.Fatalf("Register devolvió un error inesperado: %v", err)
	}
	if got.Member != "García Santiago" {
		t.Errorf("integrante = %q, se esperaba sin espacios", got.Member)
	}
	if got.Activity != "Implementar el registro" {
		t.Errorf("actividad = %q, se esperaba sin espacios", got.Activity)
	}
}

func TestRegister_HorasFraccionadas(t *testing.T) {
	env, item, sp := envForRegister(t)

	got, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Ajustar el formulario", 0.001)
	if err != nil {
		t.Fatalf("Register devolvió un error inesperado: %v", err)
	}
	if got.Hours != 0.001 {
		t.Errorf("horas = %v, se esperaba 0.001", got.Hours)
	}
}

func TestRegister_ItemTerminado(t *testing.T) {
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusFinished, sp.ID)

	if _, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Verificar el login", 1); err != nil {
		t.Fatalf("Register sobre un ítem TERMINADO devolvió un error: %v", err)
	}
}

func TestRegister_HorasInvalidas(t *testing.T) {
	cases := map[string]float64{
		"cero":         0,
		"negativa":     -1,
		"NaN":          math.NaN(),
		"infinito":     math.Inf(1),
		"infinito neg": math.Inf(-1),
	}
	for name, hours := range cases {
		t.Run(name, func(t *testing.T) {
			env, item, sp := envForRegister(t)

			_, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar", hours)

			if !errors.Is(err, ErrInvalidHours) {
				t.Fatalf("err = %v, se esperaba %v", err, ErrInvalidHours)
			}
			if len(env.memory.efforts) != 0 {
				t.Errorf("se guardaron %d esfuerzos, se esperaba ninguno", len(env.memory.efforts))
			}
		})
	}
}

func TestRegister_ItemNoAsignado(t *testing.T) {
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, 0)

	_, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar", 2)

	if !errors.Is(err, sprint.ErrItemNotAssignedToSprint) {
		t.Fatalf("err = %v, se esperaba %v", err, sprint.ErrItemNotAssignedToSprint)
	}
	if len(env.memory.efforts) != 0 {
		t.Errorf("se guardaron %d esfuerzos, se esperaba ninguno", len(env.memory.efforts))
	}
}

func TestRegister_ItemDeOtroSprint(t *testing.T) {
	env := newEffortEnv(t)
	sp1 := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	sp2 := env.addSprint(t, "Sprint 2", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, sp2.ID)

	_, err := env.manager.Register(sp1.ID, item.ID, "García Santiago", testEffortDate, "Implementar", 2)

	if !errors.Is(err, sprint.ErrItemNotAssignedToSprint) {
		t.Fatalf("err = %v, se esperaba %v", err, sprint.ErrItemNotAssignedToSprint)
	}
}

func TestRegister_SprintInexistente(t *testing.T) {
	for _, sprintID := range []int{999, 0, -1} {
		t.Run("sprint "+itoa(sprintID), func(t *testing.T) {
			env, item, _ := envForRegister(t)

			_, err := env.manager.Register(sprintID, item.ID, "García Santiago", testEffortDate, "Implementar", 2)

			if !errors.Is(err, sprint.ErrSprintNotFound) {
				t.Fatalf("err = %v, se esperaba %v", err, sprint.ErrSprintNotFound)
			}
			if len(env.memory.efforts) != 0 {
				t.Errorf("se guardaron %d esfuerzos, se esperaba ninguno", len(env.memory.efforts))
			}
		})
	}
}

func TestRegister_ItemInexistente(t *testing.T) {
	for _, itemID := range []int{999, 0, -1} {
		t.Run("ítem "+itoa(itemID), func(t *testing.T) {
			env, _, sp := envForRegister(t)

			_, err := env.manager.Register(sp.ID, itemID, "García Santiago", testEffortDate, "Implementar", 2)

			if !errors.Is(err, backlog.ErrItemNotFound) {
				t.Fatalf("err = %v, se esperaba %v", err, backlog.ErrItemNotFound)
			}
			if len(env.memory.efforts) != 0 {
				t.Errorf("se guardaron %d esfuerzos, se esperaba ninguno", len(env.memory.efforts))
			}
		})
	}
}

func TestRegister_SprintCerrado(t *testing.T) {
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusClosed)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, sp.ID)

	_, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar", 2)

	if !errors.Is(err, sprint.ErrSprintClosed) {
		t.Fatalf("err = %v, se esperaba %v", err, sprint.ErrSprintClosed)
	}
	if len(env.memory.efforts) != 0 {
		t.Errorf("se guardaron %d esfuerzos, se esperaba ninguno", len(env.memory.efforts))
	}
}

func TestRegister_DatosObligatorios(t *testing.T) {
	cases := map[string]struct {
		member, activity string
		date             time.Time
		want             error
	}{
		"integrante vacío":  {"", "Implementar", testEffortDate, ErrEmptyMember},
		"integrante spaces": {"   ", "Implementar", testEffortDate, ErrEmptyMember},
		"actividad vacía":   {"García Santiago", "", testEffortDate, ErrEmptyActivity},
		"actividad spaces":  {"García Santiago", "   ", testEffortDate, ErrEmptyActivity},
		"fecha zero":        {"García Santiago", "Implementar", time.Time{}, ErrMissingDate},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env, item, sp := envForRegister(t)

			_, err := env.manager.Register(sp.ID, item.ID, tc.member, tc.date, tc.activity, 2)

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, se esperaba %v", err, tc.want)
			}
			if len(env.memory.efforts) != 0 {
				t.Errorf("se guardaron %d esfuerzos, se esperaba ninguno", len(env.memory.efforts))
			}
		})
	}
}

func TestRegister_OrdenDeValidacion(t *testing.T) {
	env, item, sp := envForRegister(t)

	_, err := env.manager.Register(sp.ID, item.ID, "   ", testEffortDate, "   ", 0)

	if !errors.Is(err, ErrEmptyMember) {
		t.Fatalf("err = %v, se esperaba %v como primer error", err, ErrEmptyMember)
	}
}

func TestRegister_NoModificaElItem(t *testing.T) {
	env, item, sp := envForRegister(t)
	before, err := env.items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem antes de registrar: %v", err)
	}

	if _, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar", 2); err != nil {
		t.Fatalf("Register devolvió un error inesperado: %v", err)
	}

	after, err := env.items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem después de registrar: %v", err)
	}
	if after.ID != before.ID || after.ProjectID != before.ProjectID || after.Title != before.Title ||
		after.Description != before.Description || after.Priority != before.Priority ||
		after.Status != before.Status || after.StoryPoints != before.StoryPoints ||
		after.EstimatedHours != before.EstimatedHours || after.SprintID != before.SprintID ||
		!slices.Equal(after.AcceptanceCriteria, before.AcceptanceCriteria) {
		t.Errorf("ítem modificado por el registro: antes %+v, después %+v", before, after)
	}
}

func TestRegister_CopiaIndependiente(t *testing.T) {
	env, item, sp := envForRegister(t)

	got, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar", 2)
	if err != nil {
		t.Fatalf("Register devolvió un error inesperado: %v", err)
	}
	got.Member = "Otra persona"
	got.Activity = "Otra actividad"
	got.Date = got.Date.AddDate(0, 0, 1)
	got.Hours = 99
	got.ItemID = 999
	got.SprintID = 999

	saved := env.memory.efforts[0]
	if saved.Member != "García Santiago" || saved.Activity != "Implementar" ||
		!saved.Date.Equal(testEffortDate) || saved.Hours != 2 || saved.ItemID != item.ID || saved.SprintID != sp.ID {
		t.Errorf("el esfuerzo guardado cambió al modificar el devuelto: %+v", saved)
	}
}

func TestRegister_ErrorDelRepositorio(t *testing.T) {
	env, item, sp := envForRegister(t)
	env.manager = NewManager(failingEffortRepository{err: errEffortRepository}, env.sprints, env.items)

	_, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar", 2)

	if !errors.Is(err, errEffortRepository) {
		t.Fatalf("err = %v, se esperaba %v", err, errEffortRepository)
	}
	if len(env.memory.efforts) != 0 {
		t.Errorf("se guardaron %d esfuerzos, se esperaba ninguno", len(env.memory.efforts))
	}
}

func TestRegister_Concurrente(t *testing.T) {
	env, item, sp := envForRegister(t)

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
			got, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar", 1)
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
	if len(env.memory.efforts) != n {
		t.Fatalf("esfuerzos guardados = %d, se esperaban %d", len(env.memory.efforts), n)
	}
	sort.Ints(ids)
	for i, id := range ids {
		if id != i+1 {
			t.Fatalf("IDs = %v, se esperaban del 1 al %d", ids, n)
		}
	}
}

func TestRegister_ConcurrenteConCierre(t *testing.T) {
	for i := 0; i < 20; i++ {
		t.Run("intento "+itoa(i), func(t *testing.T) {
			env, item, sp := envForRegister(t)
			log := &eventLog{}
			sprintRepo := &recordingSprintRepository{Repository: env.sprintRepo, log: log}
			effortRepo := &recordingEffortRepository{MemoryRepository: env.memory, log: log}
			sprintManager := sprint.NewManager(sprintRepo, env.projects, env.items)
			manager := NewManager(effortRepo, sprintManager, env.items)

			start := make(chan struct{})
			var wg sync.WaitGroup
			var registerErr, closeErr error
			wg.Add(2)
			go func() {
				defer wg.Done()
				<-start
				_, registerErr = manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar", 2)
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
			if registerErr == nil && len(env.memory.efforts) != 1 {
				t.Fatalf("Register terminó bien pero hay %d esfuerzos", len(env.memory.efforts))
			}
			if registerErr != nil && len(env.memory.efforts) != 0 {
				t.Fatalf("Register falló pero hay %d esfuerzos guardados", len(env.memory.efforts))
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
				if event == "effort" {
					t.Fatalf("se guardó un esfuerzo después del cierre: %v", events)
				}
			}
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

// errEffortRepository es el error que devuelve el repositorio que falla.
var errEffortRepository = errors.New("fallo del repositorio de esfuerzos")

// failingEffortRepository siempre falla al guardar.
type failingEffortRepository struct {
	err error
}

func (r failingEffortRepository) Save(Effort) (Effort, error) {
	return Effort{}, r.err
}

// eventLog registra el orden de las escrituras de un escenario concurrente.
type eventLog struct {
	mu     sync.Mutex
	events []string
}

func (l *eventLog) add(event string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, event)
}

func (l *eventLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.events...)
}

// recordingSprintRepository registra el cierre efectivo del sprint.
type recordingSprintRepository struct {
	sprint.Repository
	log *eventLog
}

func (r *recordingSprintRepository) Update(sp sprint.Sprint) (sprint.Sprint, error) {
	updated, err := r.Repository.Update(sp)
	if err == nil && sp.Status == sprint.StatusClosed {
		r.log.add("close")
	}
	return updated, err
}

// recordingEffortRepository registra cada esfuerzo guardado.
type recordingEffortRepository struct {
	*MemoryRepository
	log *eventLog
}

func (r *recordingEffortRepository) Save(e Effort) (Effort, error) {
	saved, err := r.MemoryRepository.Save(e)
	if err == nil {
		r.log.add("effort")
	}
	return saved, err
}
