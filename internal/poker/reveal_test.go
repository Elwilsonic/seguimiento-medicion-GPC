package poker

import (
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
)

// HU-23 · Revelar las estimaciones (Issue #41)

// voteAll registra los votos en orden y falla el test si alguno da error.
func voteAll(t *testing.T, manager *Manager, sessionID int, votes ...Vote) {
	t.Helper()
	for _, v := range votes {
		if _, err := manager.Vote(sessionID, v.Voter, v.Points); err != nil {
			t.Fatalf("votar %+v: %v", v, err)
		}
	}
}

// CA-23.1 · Revelar devuelve todos los votos y deja la sesión REVELADA.
func TestReveal_SesionConVotosDevuelveTodosLosVotosYEstadoRevelada(t *testing.T) {
	env, session := openTestSession(t)
	want := []Vote{{Voter: "Ana", Points: 5}, {Voter: "Bruno", Points: 8}}
	voteAll(t, env.manager, session.ID, want...)

	got, err := env.manager.Reveal(session.ID)

	if err != nil {
		t.Fatalf("Reveal devolvió error: %v", err)
	}
	if got.Status != StatusRevealed {
		t.Errorf("estado devuelto = %q, se esperaba %q", got.Status, StatusRevealed)
	}
	if !slices.Equal(got.Votes, want) {
		t.Errorf("votos devueltos = %+v, se esperaba %+v", got.Votes, want)
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Status != StatusRevealed {
		t.Errorf("estado guardado = %q, se esperaba %q", stored.Status, StatusRevealed)
	}
}

// CA-23.1 · Con un solo voto ya se puede revelar.
func TestReveal_SesionConUnSoloVoto(t *testing.T) {
	env, session := openTestSession(t)
	voteAll(t, env.manager, session.ID, Vote{Voter: "Ana", Points: 3})

	got, err := env.manager.Reveal(session.ID)

	if err != nil {
		t.Fatalf("Reveal devolvió error: %v", err)
	}
	if want := []Vote{{Voter: "Ana", Points: 3}}; !slices.Equal(got.Votes, want) {
		t.Errorf("votos devueltos = %+v, se esperaba %+v", got.Votes, want)
	}
}

// CA-23.2 · Una sesión sin votos no se puede revelar y sigue ABIERTA.
func TestReveal_SesionSinVotosDevuelveErrNoVotes(t *testing.T) {
	env, session := openTestSession(t)

	_, err := env.manager.Reveal(session.ID)

	if !errors.Is(err, ErrNoVotes) {
		t.Fatalf("error = %v, se esperaba ErrNoVotes", err)
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Status != StatusOpen {
		t.Errorf("estado guardado = %q, se esperaba %q", stored.Status, StatusOpen)
	}
}

// CA-23.3 · Después de revelar no se puede votar y los votos no cambian.
func TestVote_DespuesDeRevelarDevuelveErrSessionNotOpen(t *testing.T) {
	env, session := openTestSession(t)
	voteAll(t, env.manager, session.ID, Vote{Voter: "Ana", Points: 5})
	if _, err := env.manager.Reveal(session.ID); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	for _, voter := range []string{"Ana", "Bruno"} {
		if _, err := env.manager.Vote(session.ID, voter, 8); !errors.Is(err, ErrSessionNotOpen) {
			t.Errorf("voto de %s: error = %v, se esperaba ErrSessionNotOpen", voter, err)
		}
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if want := []Vote{{Voter: "Ana", Points: 5}}; !slices.Equal(stored.Votes, want) {
		t.Errorf("votos guardados = %+v, se esperaba %+v", stored.Votes, want)
	}
}

// CA-23.4 · Revelar una sesión inexistente devuelve ErrSessionNotFound.
func TestReveal_SesionInexistenteDevuelveErrSessionNotFound(t *testing.T) {
	env, session := openTestSession(t)
	voteAll(t, env.manager, session.ID, Vote{Voter: "Ana", Points: 5})

	for _, id := range []int{999, 0, -1} {
		if _, err := env.manager.Reveal(id); !errors.Is(err, ErrSessionNotFound) {
			t.Errorf("Reveal(%d): error = %v, se esperaba ErrSessionNotFound", id, err)
		}
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Status != StatusOpen {
		t.Errorf("estado guardado = %q, la sesión existente no debía cambiar", stored.Status)
	}
}

// CA-23.4 · La segunda revelación no es idempotente: ErrSessionNotOpen.
func TestReveal_SesionYaReveladaDevuelveErrSessionNotOpen(t *testing.T) {
	env, session := openTestSession(t)
	voteAll(t, env.manager, session.ID, Vote{Voter: "Ana", Points: 5})
	if _, err := env.manager.Reveal(session.ID); err != nil {
		t.Fatalf("primera revelación: %v", err)
	}

	_, err := env.manager.Reveal(session.ID)

	if !errors.Is(err, ErrSessionNotOpen) {
		t.Fatalf("error = %v, se esperaba ErrSessionNotOpen", err)
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Status != StatusRevealed || len(stored.Votes) != 1 {
		t.Errorf("sesión guardada = %+v, se esperaba REVELADA con el voto de Ana", stored)
	}
}

// failingRevealRepository permite forzar un fallo al leer o al guardar.
type failingRevealRepository struct {
	*MemoryRepository
	failGet    bool
	failUpdate bool
}

func (r *failingRevealRepository) GetByID(id int) (Session, error) {
	if r.failGet {
		return Session{}, errRepository
	}
	return r.MemoryRepository.GetByID(id)
}

func (r *failingRevealRepository) Update(session Session) (Session, error) {
	if r.failUpdate {
		return Session{}, errRepository
	}
	return r.MemoryRepository.Update(session)
}

// openFailingSession abre una sesión con un voto sobre un repositorio que se
// puede hacer fallar.
func openFailingSession(t *testing.T) (*failingRevealRepository, *Manager, Session) {
	t.Helper()
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	repo := &failingRevealRepository{MemoryRepository: NewMemoryRepository()}
	manager := NewManager(repo, env.items)
	session, err := manager.OpenSession(item.ID)
	if err != nil {
		t.Fatalf("abrir sesión: %v", err)
	}
	voteAll(t, manager, session.ID, Vote{Voter: "Ana", Points: 5})
	return repo, manager, session
}

// CA-23.5 · Si falla la lectura, se propaga el error.
func TestReveal_ErrorAlLeerPropagaError(t *testing.T) {
	repo, manager, session := openFailingSession(t)
	repo.failGet = true

	_, err := manager.Reveal(session.ID)

	if !errors.Is(err, errRepository) {
		t.Fatalf("error = %v, se esperaba errRepository", err)
	}
}

// CA-23.5 · Si falla el guardado, se propaga el error y la sesión sigue
// ABIERTA con sus votos.
func TestReveal_ErrorAlGuardarPropagaErrorYSesionNoCambia(t *testing.T) {
	repo, manager, session := openFailingSession(t)
	repo.failUpdate = true

	_, err := manager.Reveal(session.ID)

	if !errors.Is(err, errRepository) {
		t.Fatalf("error = %v, se esperaba errRepository", err)
	}
	stored, err := repo.MemoryRepository.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Status != StatusOpen {
		t.Errorf("estado guardado = %q, se esperaba %q", stored.Status, StatusOpen)
	}
	if want := []Vote{{Voter: "Ana", Points: 5}}; !slices.Equal(stored.Votes, want) {
		t.Errorf("votos guardados = %+v, se esperaba %+v", stored.Votes, want)
	}
}

// CA-23.6 · Revelar conserva la ronda, el ítem y el orden de guardado.
func TestReveal_ConservaRondaItemYOrdenDeLosVotos(t *testing.T) {
	env, session := openTestSession(t)
	want := []Vote{{Voter: "Carla", Points: 13}, {Voter: "Ana", Points: 2}, {Voter: "Bruno", Points: 21}}
	voteAll(t, env.manager, session.ID, want...)

	got, err := env.manager.Reveal(session.ID)

	if err != nil {
		t.Fatalf("Reveal devolvió error: %v", err)
	}
	if got.ID != session.ID || got.ItemID != session.ItemID || got.Round != session.Round {
		t.Errorf("sesión devuelta = %+v, se esperaba ID %d, ítem %d y ronda %d", got, session.ID, session.ItemID, session.Round)
	}
	if !slices.Equal(got.Votes, want) {
		t.Errorf("votos devueltos = %+v, se esperaba %+v en ese orden", got.Votes, want)
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Round != session.Round || !slices.Equal(stored.Votes, want) {
		t.Errorf("sesión guardada = %+v, se esperaba la ronda %d con %+v", stored, session.Round, want)
	}
}

// CA-23.6 · Revelar no modifica el ítem ni sus Story Points.
func TestReveal_NoModificaElItemNiSusStoryPoints(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia estimada", backlog.StatusInProgress, 3)
	session, err := env.manager.OpenSession(item.ID)
	if err != nil {
		t.Fatalf("abrir sesión: %v", err)
	}
	voteAll(t, env.manager, session.ID, Vote{Voter: "Ana", Points: 8})

	if _, err := env.manager.Reveal(session.ID); err != nil {
		t.Fatalf("Reveal devolvió error: %v", err)
	}

	got, err := env.items.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if got.StoryPoints != 3 || got.Status != backlog.StatusInProgress {
		t.Errorf("ítem = %+v, se esperaba EN_CURSO con 3 story points", got)
	}
}

// CA-23.6 · Modificar la sesión devuelta no altera la guardada.
func TestReveal_DevuelveCopiaDeLosVotos(t *testing.T) {
	env, session := openTestSession(t)
	voteAll(t, env.manager, session.ID, Vote{Voter: "Ana", Points: 5})

	got, err := env.manager.Reveal(session.ID)
	if err != nil {
		t.Fatalf("Reveal devolvió error: %v", err)
	}
	if len(got.Votes) != 1 {
		t.Fatalf("votos devueltos = %+v, se esperaba el voto de Ana", got.Votes)
	}
	got.Votes[0].Points = 21
	got.Votes = append(got.Votes, Vote{Voter: "Intruso", Points: 1})

	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if want := []Vote{{Voter: "Ana", Points: 5}}; !slices.Equal(stored.Votes, want) {
		t.Errorf("votos guardados = %+v, la mutación externa no debía afectarlos", stored.Votes)
	}
}

// CA-23.7 · Votos y revelación simultáneos: cada voto queda guardado antes
// de revelar o se rechaza con ErrSessionNotOpen. Se corre con -race.
func TestReveal_ConcurrenteConVotosNoPierdeVotosAceptados(t *testing.T) {
	env, session := openTestSession(t)
	// El voto previo evita que la revelación falle por ErrNoVotes.
	voteAll(t, env.manager, session.ID, Vote{Voter: "Ana", Points: 5})

	const voters = 20
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		accepted  []string
		rejected  []error
		revealed  Session
		revealErr error
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
		s, err := env.manager.Reveal(session.ID)
		mu.Lock()
		defer mu.Unlock()
		revealed, revealErr = s, err
	}()
	close(start)
	wg.Wait()

	if revealErr != nil {
		t.Fatalf("Reveal devolvió error: %v", revealErr)
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
	if stored.Status != StatusRevealed {
		t.Errorf("estado guardado = %q, se esperaba %q", stored.Status, StatusRevealed)
	}
	for _, voter := range accepted {
		if !slices.ContainsFunc(stored.Votes, func(v Vote) bool { return v.Voter == voter }) {
			t.Errorf("el voto aceptado de %q no está entre los guardados", voter)
		}
	}
	if len(stored.Votes) != 1+len(accepted) {
		t.Errorf("votos guardados = %d, se esperaban %d (Ana más los aceptados)", len(stored.Votes), 1+len(accepted))
	}
	if !slices.Equal(revealed.Votes, stored.Votes) {
		t.Errorf("votos revelados = %+v, se esperaba lo guardado %+v", revealed.Votes, stored.Votes)
	}
}
