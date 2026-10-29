package poker

import (
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
)

// HU-24 · Detectar diferencias y repetir ronda (Issue #42)

// revealWithVotes abre una sesión para el ítem, registra los votos y la
// revela.
func revealWithVotes(t *testing.T, env *testEnv, itemID int, votes ...Vote) Session {
	t.Helper()
	session, err := env.manager.OpenSession(itemID)
	if err != nil {
		t.Fatalf("abrir sesión: %v", err)
	}
	voteAll(t, env.manager, session.ID, votes...)
	if _, err := env.manager.Reveal(session.ID); err != nil {
		t.Fatalf("revelar: %v", err)
	}
	return session
}

// expectStoredRound verifica el estado, la ronda y los votos guardados.
func expectStoredRound(t *testing.T, repo Repository, sessionID int, status Status, round int, votes []Vote) {
	t.Helper()
	stored, err := repo.GetByID(sessionID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Status != status || stored.Round != round || !slices.Equal(stored.Votes, votes) {
		t.Errorf("sesión guardada = %+v, se esperaba %s en la ronda %d con %+v", stored, status, round, votes)
	}
}

// CA-24.2 y CA-24.3 · Repetir la ronda deja la sesión ABIERTA, en la ronda
// siguiente y sin votos.
func TestRepeatRound_SesionReveladaVuelveAAbiertaConRondaSiguienteYSinVotos(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 3}, Vote{Voter: "Bruno", Points: 8})

	got, err := env.manager.RepeatRound(session.ID)

	if err != nil {
		t.Fatalf("RepeatRound devolvió error: %v", err)
	}
	if got.Status != StatusOpen || got.Round != 2 {
		t.Errorf("sesión devuelta = %+v, se esperaba ABIERTA en la ronda 2", got)
	}
	if got.Votes == nil || len(got.Votes) != 0 {
		t.Errorf("votos devueltos = %#v, se esperaba una lista vacía no nil", got.Votes)
	}
	expectStoredRound(t, env.repo, session.ID, StatusOpen, 2, []Vote{})
}

// CA-24.3 y CA-24.6 · La sesión sigue siendo la misma y el ítem no cambia.
func TestRepeatRound_ConservaIDEItemYNoModificaStoryPoints(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia estimada", backlog.StatusInProgress, 3)
	session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 8})

	got, err := env.manager.RepeatRound(session.ID)

	if err != nil {
		t.Fatalf("RepeatRound devolvió error: %v", err)
	}
	if got.ID != session.ID || got.ItemID != item.ID {
		t.Errorf("sesión devuelta = %+v, se esperaba la sesión %d del ítem %d", got, session.ID, item.ID)
	}
	if len(env.repo.sessions) != 1 {
		t.Errorf("sesiones guardadas = %d, no se debía crear otra sesión", len(env.repo.sessions))
	}
	it, err := env.items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if it.StoryPoints != 3 || it.Status != backlog.StatusInProgress {
		t.Errorf("ítem = %+v, se esperaba EN_CURSO con 3 story points", it)
	}
}

// CA-24.4 · Se puede repetir la ronda aunque no haya diferencias.
func TestRepeatRound_SinDiferenciasTambienSePuedeRepetir(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 5}, Vote{Voter: "Bruno", Points: 5})

	got, err := env.manager.RepeatRound(session.ID)

	if err != nil {
		t.Fatalf("RepeatRound devolvió error: %v", err)
	}
	if got.Status != StatusOpen || got.Round != 2 {
		t.Errorf("sesión devuelta = %+v, se esperaba ABIERTA en la ronda 2", got)
	}
}

// CA-24.5 · En las rondas nuevas se vota, se revela y se registra la
// estimación como en la primera.
func TestRepeatRound_NuevaRondaPermiteVotarRevelarYEstimar(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 3}, Vote{Voter: "Bruno", Points: 8})

	if _, err := env.manager.RepeatRound(session.ID); err != nil {
		t.Fatalf("primera repetición: %v", err)
	}
	voteAll(t, env.manager, session.ID, Vote{Voter: "Bruno", Points: 5})
	second, err := env.manager.Reveal(session.ID)
	if err != nil {
		t.Fatalf("revelar la ronda 2: %v", err)
	}
	if want := []Vote{{Voter: "Bruno", Points: 5}}; second.Round != 2 || !slices.Equal(second.Votes, want) {
		t.Errorf("ronda revelada = %+v, se esperaba la ronda 2 con %+v", second, want)
	}

	if _, err := env.manager.RepeatRound(session.ID); err != nil {
		t.Fatalf("segunda repetición: %v", err)
	}
	voteAll(t, env.manager, session.ID, Vote{Voter: "Ana", Points: 8}, Vote{Voter: "Bruno", Points: 8})
	if _, err := env.manager.Reveal(session.ID); err != nil {
		t.Fatalf("revelar la ronda 3: %v", err)
	}
	closed, err := env.manager.Estimate(session.ID, 8)
	if err != nil {
		t.Fatalf("Estimate: %v", err)
	}
	if closed.Status != StatusClosed || closed.Round != 3 {
		t.Errorf("sesión estimada = %+v, se esperaba CERRADA en la ronda 3", closed)
	}
	expectItemPoints(t, env.items, item.ID, 8)
}

// CA-24.7 · Repetir una sesión inexistente devuelve ErrSessionNotFound.
func TestRepeatRound_SesionInexistenteDevuelveErrSessionNotFound(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 5})

	for _, id := range []int{999, 0, -1} {
		if _, err := env.manager.RepeatRound(id); !errors.Is(err, ErrSessionNotFound) {
			t.Errorf("RepeatRound(%d): error = %v, se esperaba ErrSessionNotFound", id, err)
		}
	}
	expectStoredRound(t, env.repo, session.ID, StatusRevealed, 1, []Vote{{Voter: "Ana", Points: 5}})
}

// CA-24.7 · Una sesión ABIERTA no se puede repetir y no cambia.
func TestRepeatRound_SesionAbiertaDevuelveErrSessionNotRevealed(t *testing.T) {
	env, session := openTestSession(t)
	voteAll(t, env.manager, session.ID, Vote{Voter: "Ana", Points: 5})

	_, err := env.manager.RepeatRound(session.ID)

	if !errors.Is(err, ErrSessionNotRevealed) {
		t.Fatalf("error = %v, se esperaba ErrSessionNotRevealed", err)
	}
	expectStoredRound(t, env.repo, session.ID, StatusOpen, 1, []Vote{{Voter: "Ana", Points: 5}})
}

// CA-24.7 · Una sesión CERRADA no se puede repetir y no cambia.
func TestRepeatRound_SesionCerradaDevuelveErrSessionNotRevealed(t *testing.T) {
	env := newTestEnv(t)
	item, session := revealTestSession(t, env, 0)
	if _, err := env.manager.Estimate(session.ID, 5); err != nil {
		t.Fatalf("Estimate: %v", err)
	}

	_, err := env.manager.RepeatRound(session.ID)

	if !errors.Is(err, ErrSessionNotRevealed) {
		t.Fatalf("error = %v, se esperaba ErrSessionNotRevealed", err)
	}
	expectStoredRound(t, env.repo, session.ID, StatusClosed, 1, revealedVotes)
	expectItemPoints(t, env.items, item.ID, 5)
}

// CA-24.8 · Si se abrió otra sesión para el ítem mientras esta estaba
// REVELADA, no se repite la ronda (RN-24.7).
func TestRepeatRound_OtraSesionAbiertaDelItemDevuelveErrSessionAlreadyOpen(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 5})
	if _, err := env.manager.OpenSession(item.ID); err != nil {
		t.Fatalf("abrir otra sesión (D10): %v", err)
	}

	_, err := env.manager.RepeatRound(session.ID)

	if !errors.Is(err, ErrSessionAlreadyOpen) {
		t.Fatalf("error = %v, se esperaba ErrSessionAlreadyOpen", err)
	}
	expectStoredRound(t, env.repo, session.ID, StatusRevealed, 1, []Vote{{Voter: "Ana", Points: 5}})
	if got := countOpenSessions(env.repo, item.ID); got != 1 {
		t.Errorf("sesiones abiertas del ítem = %d, se esperaba 1", got)
	}
}

// failingRoundRepository permite forzar un fallo al leer, al buscar la sesión
// abierta del ítem o al guardar.
type failingRoundRepository struct {
	*MemoryRepository
	failGet        bool
	failOpenByItem bool
	failUpdate     bool
}

func (r *failingRoundRepository) GetByID(id int) (Session, error) {
	if r.failGet {
		return Session{}, errRepository
	}
	return r.MemoryRepository.GetByID(id)
}

func (r *failingRoundRepository) OpenByItem(itemID int) (Session, bool, error) {
	if r.failOpenByItem {
		return Session{}, false, errRepository
	}
	return r.MemoryRepository.OpenByItem(itemID)
}

func (r *failingRoundRepository) Update(session Session) (Session, error) {
	if r.failUpdate {
		return Session{}, errRepository
	}
	return r.MemoryRepository.Update(session)
}

// revealFailingRoundSession revela una sesión con el voto de Ana sobre un
// repositorio que se puede hacer fallar.
func revealFailingRoundSession(t *testing.T) (*failingRoundRepository, *Manager, Session) {
	t.Helper()
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	repo := &failingRoundRepository{MemoryRepository: NewMemoryRepository()}
	manager := NewManager(repo, env.items)
	session, err := manager.OpenSession(item.ID)
	if err != nil {
		t.Fatalf("abrir sesión: %v", err)
	}
	voteAll(t, manager, session.ID, Vote{Voter: "Ana", Points: 5})
	if _, err := manager.Reveal(session.ID); err != nil {
		t.Fatalf("revelar: %v", err)
	}
	return repo, manager, session
}

// CA-24.9 · Si falla la lectura, se propaga el error.
func TestRepeatRound_ErrorAlLeerPropagaError(t *testing.T) {
	repo, manager, session := revealFailingRoundSession(t)
	repo.failGet = true

	_, err := manager.RepeatRound(session.ID)

	if !errors.Is(err, errRepository) {
		t.Fatalf("error = %v, se esperaba errRepository", err)
	}
}

// CA-24.9 · Si falla la búsqueda de otra sesión abierta, se propaga el error
// y la sesión sigue REVELADA.
func TestRepeatRound_ErrorAlBuscarSesionAbiertaPropagaError(t *testing.T) {
	repo, manager, session := revealFailingRoundSession(t)
	repo.failOpenByItem = true

	_, err := manager.RepeatRound(session.ID)

	if !errors.Is(err, errRepository) {
		t.Fatalf("error = %v, se esperaba errRepository", err)
	}
	expectStoredRound(t, repo.MemoryRepository, session.ID, StatusRevealed, 1, []Vote{{Voter: "Ana", Points: 5}})
}

// CA-24.9 · Si falla el guardado, se propaga el error y la sesión sigue
// REVELADA con su ronda y sus votos.
func TestRepeatRound_ErrorAlGuardarPropagaErrorYSesionNoCambia(t *testing.T) {
	repo, manager, session := revealFailingRoundSession(t)
	repo.failUpdate = true

	_, err := manager.RepeatRound(session.ID)

	if !errors.Is(err, errRepository) {
		t.Fatalf("error = %v, se esperaba errRepository", err)
	}
	expectStoredRound(t, repo.MemoryRepository, session.ID, StatusRevealed, 1, []Vote{{Voter: "Ana", Points: 5}})
}

// CA-24.6 · Modificar la sesión devuelta no altera la guardada.
func TestRepeatRound_DevuelveCopia(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 5})

	got, err := env.manager.RepeatRound(session.ID)
	if err != nil {
		t.Fatalf("RepeatRound devolvió error: %v", err)
	}
	got.Round = 99
	got.Status = StatusClosed
	got.Votes = append(got.Votes, Vote{Voter: "Intruso", Points: 1})

	expectStoredRound(t, env.repo, session.ID, StatusOpen, 2, []Vote{})
}

// CA-24.10 · Repeticiones simultáneas de la misma sesión aumentan la ronda
// una sola vez. Se corre con -race.
func TestRepeatRound_ConcurrenteAumentaLaRondaUnaSolaVez(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 5})

	const attempts = 10
	errs := make([]error, attempts)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, errs[i] = env.manager.RepeatRound(session.ID)
		}()
	}
	close(start)
	wg.Wait()

	succeeded := 0
	for _, err := range errs {
		switch {
		case err == nil:
			succeeded++
		case !errors.Is(err, ErrSessionNotRevealed):
			t.Errorf("repetición rechazada con %v, se esperaba ErrSessionNotRevealed", err)
		}
	}
	if succeeded != 1 {
		t.Errorf("repeticiones exitosas = %d, se esperaba 1", succeeded)
	}
	expectStoredRound(t, env.repo, session.ID, StatusOpen, 2, []Vote{})
}

// CA-24.10 · Votos simultáneos con la repetición: cada voto se rechaza con
// ErrSessionNotOpen o queda en la ronda nueva. Se corre con -race.
func TestRepeatRound_ConcurrenteConVotosNoMezclaRondas(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 5})

	const voters = 20
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		accepted  []string
		rejected  []error
		repeatErr error
	)
	start := make(chan struct{})
	for i := 1; i <= voters; i++ {
		wg.Add(1)
		go func(voter string) {
			defer wg.Done()
			<-start
			_, err := env.manager.Vote(session.ID, voter, 8)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				accepted = append(accepted, voter)
			} else {
				rejected = append(rejected, err)
			}
		}("Integrante " + strconv.Itoa(i))
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		_, err := env.manager.RepeatRound(session.ID)
		mu.Lock()
		defer mu.Unlock()
		repeatErr = err
	}()
	close(start)
	wg.Wait()

	if repeatErr != nil {
		t.Fatalf("RepeatRound devolvió error: %v", repeatErr)
	}
	for _, err := range rejected {
		if !errors.Is(err, ErrSessionNotOpen) {
			t.Errorf("voto rechazado con %v, se esperaba ErrSessionNotOpen", err)
		}
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Status != StatusOpen || stored.Round != 2 {
		t.Errorf("sesión guardada = %+v, se esperaba ABIERTA en la ronda 2", stored)
	}
	if slices.ContainsFunc(stored.Votes, func(v Vote) bool { return v.Voter == "Ana" }) {
		t.Errorf("votos guardados = %+v, el voto de la ronda 1 debía descartarse", stored.Votes)
	}
	for _, voter := range accepted {
		if !slices.ContainsFunc(stored.Votes, func(v Vote) bool { return v.Voter == voter }) {
			t.Errorf("el voto aceptado de %q no está en la ronda nueva", voter)
		}
	}
	if len(stored.Votes) != len(accepted) {
		t.Errorf("votos guardados = %d, se esperaban %d (los aceptados)", len(stored.Votes), len(accepted))
	}
}

// CA-24.10 · Repetir la ronda y abrir otra sesión para el mismo ítem al mismo
// tiempo nunca deja dos sesiones ABIERTA. Se corre con -race.
func TestRepeatRound_ConcurrenteConAperturaDejaUnaSolaSesionAbierta(t *testing.T) {
	for range 20 {
		env := newTestEnv(t)
		item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
		session := revealWithVotes(t, env, item.ID, Vote{Voter: "Ana", Points: 5})

		var (
			wg                 sync.WaitGroup
			repeatErr, openErr error
		)
		start := make(chan struct{})
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_, repeatErr = env.manager.RepeatRound(session.ID)
		}()
		go func() {
			defer wg.Done()
			<-start
			_, openErr = env.manager.OpenSession(item.ID)
		}()
		close(start)
		wg.Wait()

		if (repeatErr == nil) == (openErr == nil) {
			t.Fatalf("repetir = %v, abrir = %v; se esperaba que exactamente una tuviera éxito", repeatErr, openErr)
		}
		for _, err := range []error{repeatErr, openErr} {
			if err != nil && !errors.Is(err, ErrSessionAlreadyOpen) {
				t.Errorf("error = %v, se esperaba ErrSessionAlreadyOpen", err)
			}
		}
		if got := countOpenSessions(env.repo, item.ID); got != 1 {
			t.Fatalf("sesiones abiertas del ítem = %d, se esperaba 1", got)
		}
	}
}
