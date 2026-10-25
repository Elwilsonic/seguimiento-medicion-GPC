package poker

import (
	"errors"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

var errRepository = errors.New("falla de almacenamiento")

// testEnv arma los managers de proyectos, backlog y Planning Poker con un
// proyecto "Proyecto de prueba" ya creado.
type testEnv struct {
	projects  *project.Manager
	items     *backlog.Manager
	repo      *MemoryRepository
	manager   *Manager
	projectID int
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	projects := project.NewManager(project.NewMemoryRepository())
	p, err := projects.Create("Proyecto de prueba", "Proyecto para los tests de HU-21")
	if err != nil {
		t.Fatalf("crear proyecto: %v", err)
	}
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	repo := NewMemoryRepository()
	return &testEnv{
		projects:  projects,
		items:     items,
		repo:      repo,
		manager:   NewManager(repo, items),
		projectID: p.ID,
	}
}

// createItem crea un ítem en el proyecto de prueba con el estado y los story
// points indicados.
func (e *testEnv) createItem(t *testing.T, title string, status backlog.Status, points int) backlog.Item {
	t.Helper()
	it, err := e.items.Create(e.projectID, title, "Descripción de "+title, "MEDIA", []string{"Criterio de " + title})
	if err != nil {
		t.Fatalf("crear ítem %q: %v", title, err)
	}
	if points > 0 {
		it, err = e.items.AssignStoryPoints(it.ID, points)
		if err != nil {
			t.Fatalf("asignar story points al ítem %q: %v", title, err)
		}
	}
	switch status {
	case backlog.StatusInProgress:
		it, err = e.items.ChangeStatus(it.ID, string(backlog.StatusInProgress))
		if err != nil {
			t.Fatalf("pasar el ítem %q a EN_CURSO: %v", title, err)
		}
	case backlog.StatusFinished:
		for _, next := range []backlog.Status{backlog.StatusInProgress, backlog.StatusFinished} {
			it, err = e.items.ChangeStatus(it.ID, string(next))
			if err != nil {
				t.Fatalf("pasar el ítem %q a %s: %v", title, next, err)
			}
		}
	}
	return it
}

// countOpenSessions cuenta las sesiones ABIERTA guardadas para un ítem.
func countOpenSessions(repo *MemoryRepository, itemID int) int {
	count := 0
	for _, s := range repo.sessions {
		if s.ItemID == itemID && s.Status == StatusOpen {
			count++
		}
	}
	return count
}

// CA-21.1 · Abrir una sesión deja una sesión ABIERTA, en la ronda 1 y sin
// votos, asociada al ítem.
func TestOpenSession_CreaSesionAbiertaSinVotos(t *testing.T) {
	env := newTestEnv(t)
	it := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)

	s, err := env.manager.OpenSession(it.ID)

	if err != nil {
		t.Fatalf("OpenSession(%d) devolvió error: %v", it.ID, err)
	}
	if s.ID != 1 || s.ItemID != it.ID || s.Status != StatusOpen || s.Round != 1 {
		t.Errorf("sesión = %+v, se esperaba ID 1, ítem %d, ABIERTA y ronda 1", s, it.ID)
	}
	if s.Votes == nil {
		t.Error("Votes = nil, se esperaba una lista vacía no nula")
	}
	if len(s.Votes) != 0 {
		t.Errorf("votos = %+v, se esperaba una sesión sin votos", s.Votes)
	}
	if got := countOpenSessions(env.repo, it.ID); got != 1 {
		t.Errorf("sesiones abiertas guardadas = %d, se esperaba 1", got)
	}
}

// CA-21.1 · Los IDs de las sesiones son correlativos globales.
func TestOpenSession_AsignaIDsCorrelativosGlobales(t *testing.T) {
	env := newTestEnv(t)
	first := env.createItem(t, "Primera historia", backlog.StatusPending, 0)
	second := env.createItem(t, "Segunda historia", backlog.StatusPending, 0)

	s1, err := env.manager.OpenSession(first.ID)
	if err != nil {
		t.Fatalf("abrir la primera sesión: %v", err)
	}
	s2, err := env.manager.OpenSession(second.ID)
	if err != nil {
		t.Fatalf("abrir la segunda sesión: %v", err)
	}

	if s1.ID != 1 || s2.ID != 2 {
		t.Errorf("IDs = %d y %d, se esperaban 1 y 2", s1.ID, s2.ID)
	}
	if s1.ItemID != first.ID || s2.ItemID != second.ID {
		t.Errorf("ítems = %d y %d, se esperaban %d y %d", s1.ItemID, s2.ItemID, first.ID, second.ID)
	}
}

// CA-21.1 · Un ítem EN_CURSO con story points también se puede volver a
// estimar.
func TestOpenSession_ItemEnCursoConStoryPoints(t *testing.T) {
	env := newTestEnv(t)
	it := env.createItem(t, "Historia estimada", backlog.StatusInProgress, 5)

	s, err := env.manager.OpenSession(it.ID)

	if err != nil {
		t.Fatalf("OpenSession(%d) devolvió error: %v", it.ID, err)
	}
	if s.ItemID != it.ID || s.Status != StatusOpen || len(s.Votes) != 0 {
		t.Errorf("sesión = %+v, se esperaba una sesión abierta y sin votos para el ítem %d", s, it.ID)
	}
}

// CA-21.4 · Un ítem inexistente se informa con el error de HU-04.
func TestOpenSession_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
	env := newTestEnv(t)

	_, err := env.manager.OpenSession(999)

	if !errors.Is(err, backlog.ErrItemNotFound) {
		t.Fatalf("error = %v, se esperaba backlog.ErrItemNotFound", err)
	}
}

// CA-21.4 · Un ítem TERMINADO no admite una sesión nueva.
func TestOpenSession_ItemTerminadoDevuelveErrItemFinished(t *testing.T) {
	env := newTestEnv(t)
	it := env.createItem(t, "Historia terminada", backlog.StatusFinished, 0)

	_, err := env.manager.OpenSession(it.ID)

	if !errors.Is(err, ErrItemFinished) {
		t.Fatalf("error = %v, se esperaba poker.ErrItemFinished", err)
	}
	if got := countOpenSessions(env.repo, it.ID); got != 0 {
		t.Errorf("sesiones abiertas guardadas = %d, se esperaba 0", got)
	}
}

// CA-21.2 · Un ítem no admite dos sesiones abiertas al mismo tiempo.
func TestOpenSession_SegundaSesionDevuelveErrSessionAlreadyOpen(t *testing.T) {
	env := newTestEnv(t)
	it := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)

	if _, err := env.manager.OpenSession(it.ID); err != nil {
		t.Fatalf("abrir la primera sesión: %v", err)
	}
	_, err := env.manager.OpenSession(it.ID)

	if !errors.Is(err, ErrSessionAlreadyOpen) {
		t.Fatalf("error = %v, se esperaba poker.ErrSessionAlreadyOpen", err)
	}
	if got := countOpenSessions(env.repo, it.ID); got != 1 {
		t.Errorf("sesiones abiertas guardadas = %d, se esperaba 1", got)
	}
}

// RN-21.8 · La sesión devuelta no comparte la lista de votos con la guardada.
func TestOpenSession_DevuelveCopiaDeLosVotos(t *testing.T) {
	env := newTestEnv(t)
	it := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)

	s, err := env.manager.OpenSession(it.ID)
	if err != nil {
		t.Fatalf("OpenSession(%d) devolvió error: %v", it.ID, err)
	}
	s.Votes = append(s.Votes, Vote{Voter: "Cristobal", Points: 5})

	stored, exists, err := env.repo.OpenByItem(it.ID)
	if err != nil || !exists {
		t.Fatalf("OpenByItem(%d) = (%+v, %v, %v), se esperaba la sesión guardada", it.ID, stored, exists, err)
	}
	if len(stored.Votes) != 0 {
		t.Errorf("votos guardados = %+v, la mutación externa no debía afectarlos", stored.Votes)
	}
}

// CA-21.6 · Dos aperturas simultáneas dejan una sola sesión abierta.
func TestOpenSession_ConcurrenteSoloUnaSesionAbierta(t *testing.T) {
	env := newTestEnv(t)
	it := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make([]error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = env.manager.OpenSession(it.ID)
		}(i)
	}
	wg.Wait()

	opened := 0
	for _, err := range errs {
		switch {
		case err == nil:
			opened++
		case !errors.Is(err, ErrSessionAlreadyOpen):
			t.Errorf("error inesperado: %v", err)
		}
	}
	if opened != 1 {
		t.Errorf("aperturas exitosas = %d, se esperaba exactamente 1", opened)
	}
	if got := countOpenSessions(env.repo, it.ID); got != 1 {
		t.Errorf("sesiones abiertas guardadas = %d, se esperaba 1", got)
	}
}

// CA-21.5 · Las cartas son la escala de HU-20 y cada llamada devuelve una
// copia.
func TestCards_DevuelveLaEscalaYUnaCopia(t *testing.T) {
	want := backlog.StoryPointScale()

	got := Cards()
	if len(got) != len(want) {
		t.Fatalf("cartas = %v, se esperaba %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("cartas = %v, se esperaba %v", got, want)
		}
	}

	got[0] = 99
	if again := Cards(); again[0] != want[0] {
		t.Errorf("cartas = %v, la mutación externa no debía afectar la escala", again)
	}
}

// failingRepository permite forzar un fallo de apertura o de guardado.
type failingRepository struct {
	base     Repository
	failOpen bool
	failSave bool
}

func (r failingRepository) Save(session Session) (Session, error) {
	if r.failSave {
		return Session{}, errRepository
	}
	return r.base.Save(session)
}

func (r failingRepository) OpenByItem(itemID int) (Session, bool, error) {
	if r.failOpen {
		return Session{}, false, errRepository
	}
	return r.base.OpenByItem(itemID)
}

func (r failingRepository) GetByID(id int) (Session, error) {
	return r.base.GetByID(id)
}

func (r failingRepository) Update(session Session) (Session, error) {
	return r.base.Update(session)
}

// CA-21.4 · Los errores del repositorio se propagan sin convertir.
func TestOpenSession_ErrorDelRepositorioSEPropaga(t *testing.T) {
	env := newTestEnv(t)
	it := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)

	t.Run("al buscar la sesión abierta", func(t *testing.T) {
		manager := NewManager(failingRepository{base: env.repo, failOpen: true}, env.items)

		if _, err := manager.OpenSession(it.ID); !errors.Is(err, errRepository) {
			t.Fatalf("error = %v, se esperaba errRepository", err)
		}
	})

	t.Run("al guardar la sesión", func(t *testing.T) {
		manager := NewManager(failingRepository{base: env.repo, failSave: true}, env.items)

		if _, err := manager.OpenSession(it.ID); !errors.Is(err, errRepository) {
			t.Fatalf("error = %v, se esperaba errRepository", err)
		}
	})
}
