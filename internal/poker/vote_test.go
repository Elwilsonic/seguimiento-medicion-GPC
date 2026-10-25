package poker

import (
	"errors"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
)

func openTestSession(t *testing.T) (*testEnv, Session) {
	t.Helper()
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	session, err := env.manager.OpenSession(item.ID)
	if err != nil {
		t.Fatalf("abrir sesión: %v", err)
	}
	return env, session
}

func TestVote_VotoValidoPermaneceOculto(t *testing.T) {
	env, session := openTestSession(t)

	got, err := env.manager.Vote(session.ID, "Ana", 5)

	if err != nil {
		t.Fatalf("Vote devolvió error: %v", err)
	}
	if got.Votes == nil || len(got.Votes) != 0 {
		t.Errorf("votos devueltos = %+v, se esperaba una lista vacía no nula", got.Votes)
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(stored.Votes) != 1 || stored.Votes[0].Points != 5 {
		t.Errorf("votos guardados = %+v, se esperaba el voto de Ana", stored.Votes)
	}
}

func TestVote_ReemplazaVotoYGuardaNombreRecortado(t *testing.T) {
	env, session := openTestSession(t)
	if _, err := env.manager.Vote(session.ID, "Ana", 5); err != nil {
		t.Fatalf("primer voto: %v", err)
	}

	got, err := env.manager.Vote(session.ID, " ANA ", 8)

	if err != nil {
		t.Fatalf("reemplazar voto: %v", err)
	}
	if len(got.Votes) != 0 {
		t.Errorf("votos devueltos = %+v, se esperaba ocultamiento", got.Votes)
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(stored.Votes) != 1 || stored.Votes[0].Points != 8 || stored.Votes[0].Voter != "ANA" {
		t.Errorf("votos guardados = %+v, se esperaba un único voto de ANA con 8 puntos", stored.Votes)
	}
}

func TestVote_VotanteVacioDevuelveError(t *testing.T) {
	env, session := openTestSession(t)

	_, err := env.manager.Vote(session.ID, "   ", 5)

	if !errors.Is(err, ErrEmptyVoter) {
		t.Fatalf("error = %v, se esperaba ErrEmptyVoter", err)
	}
}

func TestVote_CartaInvalidaDevuelveError(t *testing.T) {
	env, session := openTestSession(t)

	_, err := env.manager.Vote(session.ID, "Ana", 4)

	if !errors.Is(err, ErrInvalidCard) {
		t.Fatalf("error = %v, se esperaba ErrInvalidCard", err)
	}
}

func TestVote_SesionInexistenteDevuelveError(t *testing.T) {
	env := newTestEnv(t)

	_, err := env.manager.Vote(999, "Ana", 5)

	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("error = %v, se esperaba ErrSessionNotFound", err)
	}
}

func TestVote_SesionNoAbiertaDevuelveError(t *testing.T) {
	env, session := openTestSession(t)
	session.Status = Status("REVELADA")
	if _, err := env.repo.Update(session); err != nil {
		t.Fatalf("preparar sesión no abierta: %v", err)
	}

	_, err := env.manager.Vote(session.ID, "Ana", 5)

	if !errors.Is(err, ErrSessionNotOpen) {
		t.Fatalf("error = %v, se esperaba ErrSessionNotOpen", err)
	}
}

type failingVoteRepository struct {
	*MemoryRepository
	failUpdate bool
}

func (r *failingVoteRepository) Update(session Session) (Session, error) {
	if r.failUpdate {
		return Session{}, errRepository
	}
	return r.MemoryRepository.Update(session)
}

func TestVote_ErrorAlGuardarPropagaErrorYSesionNoSeModifica(t *testing.T) {
	env := newTestEnv(t)
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, 0)
	repo := &failingVoteRepository{MemoryRepository: NewMemoryRepository()}
	manager := NewManager(repo, env.items)
	session, err := manager.OpenSession(item.ID)
	if err != nil {
		t.Fatalf("abrir sesión: %v", err)
	}
	repo.failUpdate = true

	_, err = manager.Vote(session.ID, "Ana", 5)

	if !errors.Is(err, errRepository) {
		t.Fatalf("error = %v, se esperaba errRepository", err)
	}
	repo.failUpdate = false
	stored, err := repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(stored.Votes) != 0 {
		t.Errorf("votos guardados = %+v, se esperaba que la sesión no se modificara", stored.Votes)
	}
}

func TestVote_ConcurrenteConservaVotosDeIntegrantesDistintos(t *testing.T) {
	env, session := openTestSession(t)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, voter := range []string{"Ana", "Bruno"} {
		wg.Add(1)
		go func(voter string) {
			defer wg.Done()
			_, err := env.manager.Vote(session.ID, voter, 5)
			errs <- err
		}(voter)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("voto concurrente: %v", err)
		}
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(stored.Votes) != 2 {
		t.Errorf("votos guardados = %+v, se esperaban dos votos", stored.Votes)
	}
}
