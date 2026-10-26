package poker

import (
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-25 · Registrar la estimación acordada (Issue #43)

// revealedVotes son los votos de las sesiones reveladas de estos tests.
var revealedVotes = []Vote{{Voter: "Ana", Points: 5}, {Voter: "Bruno", Points: 8}}

// revealTestSession abre una sesión para un ítem nuevo con los story points
// indicados, registra revealedVotes y la revela.
func revealTestSession(t *testing.T, env *testEnv, points int) (backlog.Item, Session) {
	t.Helper()
	item := env.createItem(t, "Historia de ejemplo", backlog.StatusPending, points)
	session, err := env.manager.OpenSession(item.ID)
	if err != nil {
		t.Fatalf("abrir sesión: %v", err)
	}
	voteAll(t, env.manager, session.ID, revealedVotes...)
	if _, err := env.manager.Reveal(session.ID); err != nil {
		t.Fatalf("revelar: %v", err)
	}
	return item, session
}

// expectStoredStatus verifica el estado guardado de la sesión.
func expectStoredStatus(t *testing.T, repo Repository, sessionID int, want Status) {
	t.Helper()
	stored, err := repo.GetByID(sessionID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Status != want {
		t.Errorf("estado guardado = %q, se esperaba %q", stored.Status, want)
	}
}

// expectItemPoints verifica los story points guardados del ítem.
func expectItemPoints(t *testing.T, items *backlog.Manager, itemID, want int) {
	t.Helper()
	it, err := items.GetItem(itemID)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if it.StoryPoints != want {
		t.Errorf("story points del ítem = %d, se esperaban %d", it.StoryPoints, want)
	}
}

// CA-25.1 y CA-25.3 · Registrar asigna los puntos al ítem y cierra la sesión.
func TestEstimate_SesionReveladaAsignaLosPuntosYCierraLaSesion(t *testing.T) {
	env := newTestEnv(t)
	item, session := revealTestSession(t, env, 0)

	got, err := env.manager.Estimate(session.ID, 5)

	if err != nil {
		t.Fatalf("Estimate devolvió error: %v", err)
	}
	if got.ID != session.ID || got.Status != StatusClosed {
		t.Errorf("sesión devuelta = %+v, se esperaba la sesión %d CERRADA", got, session.ID)
	}
	if !slices.Equal(got.Votes, revealedVotes) {
		t.Errorf("votos devueltos = %+v, se esperaba %+v", got.Votes, revealedVotes)
	}
	expectStoredStatus(t, env.repo, session.ID, StatusClosed)
	expectItemPoints(t, env.items, item.ID, 5)
}

// CA-25.2 · El valor acordado no tiene que coincidir con los votos.
func TestEstimate_ValorQueNadieVotoSeAcepta(t *testing.T) {
	env := newTestEnv(t)
	item, session := revealTestSession(t, env, 0)

	if _, err := env.manager.Estimate(session.ID, 13); err != nil {
		t.Fatalf("Estimate devolvió error: %v", err)
	}

	expectItemPoints(t, env.items, item.ID, 13)
	expectStoredStatus(t, env.repo, session.ID, StatusClosed)
}

// CA-25.2 · Cada valor de la escala se puede registrar.
func TestEstimate_AceptaCadaValorDeLaEscala(t *testing.T) {
	for _, points := range Cards() {
		t.Run(strconv.Itoa(points), func(t *testing.T) {
			env := newTestEnv(t)
			item, session := revealTestSession(t, env, 0)

			if _, err := env.manager.Estimate(session.ID, points); err != nil {
				t.Fatalf("Estimate(%d) devolvió error: %v", points, err)
			}

			expectItemPoints(t, env.items, item.ID, points)
		})
	}
}

// CA-25.2 · Una carta fuera de la escala da ErrInvalidCard y no cambia nada.
func TestEstimate_CartaFueraDeLaEscalaDevuelveErrInvalidCard(t *testing.T) {
	env := newTestEnv(t)
	item, session := revealTestSession(t, env, 0)

	for _, points := range []int{0, -1, 4, 22} {
		if _, err := env.manager.Estimate(session.ID, points); !errors.Is(err, ErrInvalidCard) {
			t.Errorf("Estimate(%d): error = %v, se esperaba ErrInvalidCard", points, err)
		}
	}

	expectStoredStatus(t, env.repo, session.ID, StatusRevealed)
	expectItemPoints(t, env.items, item.ID, 0)
}

// CA-25.2 · Los puntos se validan antes que la existencia de la sesión.
func TestEstimate_CartaInvalidaSeInformaAntesQueLaSesionInexistente(t *testing.T) {
	env := newTestEnv(t)

	_, err := env.manager.Estimate(999, 4)

	if !errors.Is(err, ErrInvalidCard) {
		t.Fatalf("error = %v, se esperaba ErrInvalidCard", err)
	}
}

// CA-25.1 · El valor acordado reemplaza los story points anteriores.
func TestEstimate_ReemplazaLosStoryPointsAnterioresDelItem(t *testing.T) {
	env := newTestEnv(t)
	item, session := revealTestSession(t, env, 3)

	if _, err := env.manager.Estimate(session.ID, 8); err != nil {
		t.Fatalf("Estimate devolvió error: %v", err)
	}

	expectItemPoints(t, env.items, item.ID, 8)
}

// CA-25.4 · Una sesión inexistente da ErrSessionNotFound.
func TestEstimate_SesionInexistenteDevuelveErrSessionNotFound(t *testing.T) {
	env := newTestEnv(t)
	item, session := revealTestSession(t, env, 0)

	for _, id := range []int{999, 0, -1} {
		if _, err := env.manager.Estimate(id, 5); !errors.Is(err, ErrSessionNotFound) {
			t.Errorf("Estimate(%d): error = %v, se esperaba ErrSessionNotFound", id, err)
		}
	}

	expectStoredStatus(t, env.repo, session.ID, StatusRevealed)
	expectItemPoints(t, env.items, item.ID, 0)
}

// CA-25.4 · Una sesión ABIERTA no se puede estimar y el ítem no cambia.
func TestEstimate_SesionAbiertaDevuelveErrSessionNotRevealed(t *testing.T) {
	env, session := openTestSession(t)
	voteAll(t, env.manager, session.ID, revealedVotes...)

	_, err := env.manager.Estimate(session.ID, 5)

	if !errors.Is(err, ErrSessionNotRevealed) {
		t.Fatalf("error = %v, se esperaba ErrSessionNotRevealed", err)
	}
	expectStoredStatus(t, env.repo, session.ID, StatusOpen)
	expectItemPoints(t, env.items, session.ItemID, 0)
}

// CA-25.4 · Una sesión CERRADA no se vuelve a estimar: no es idempotente.
func TestEstimate_SesionCerradaDevuelveErrSessionNotRevealed(t *testing.T) {
	env := newTestEnv(t)
	item, session := revealTestSession(t, env, 0)
	if _, err := env.manager.Estimate(session.ID, 5); err != nil {
		t.Fatalf("primer registro: %v", err)
	}

	_, err := env.manager.Estimate(session.ID, 8)

	if !errors.Is(err, ErrSessionNotRevealed) {
		t.Fatalf("error = %v, se esperaba ErrSessionNotRevealed", err)
	}
	expectStoredStatus(t, env.repo, session.ID, StatusClosed)
	expectItemPoints(t, env.items, item.ID, 5)
}

// CA-25.5 · Un ítem inexistente propaga backlog.ErrItemNotFound y la sesión
// sigue REVELADA.
func TestEstimate_ItemInexistenteDevuelveErrItemNotFoundYSesionSigueRevelada(t *testing.T) {
	env := newTestEnv(t)
	session, err := env.repo.Save(Session{ItemID: 999, Status: StatusRevealed, Round: 1, Votes: revealedVotes})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	_, err = env.manager.Estimate(session.ID, 5)

	if !errors.Is(err, backlog.ErrItemNotFound) {
		t.Fatalf("error = %v, se esperaba backlog.ErrItemNotFound", err)
	}
	expectStoredStatus(t, env.repo, session.ID, StatusRevealed)
}

// CA-25.5 · Un ítem TERMINADO propaga backlog.ErrItemFinished y la sesión
// sigue REVELADA. Si el ítem se reabre (HU-08), la sesión se puede cerrar.
func TestEstimate_ItemTerminadoDevuelveErrItemFinishedYSesionSigueRevelada(t *testing.T) {
	env := newTestEnv(t)
	item, session := revealTestSession(t, env, 0)
	for _, next := range []backlog.Status{backlog.StatusInProgress, backlog.StatusFinished} {
		if _, err := env.items.ChangeStatus(item.ID, string(next)); err != nil {
			t.Fatalf("pasar el ítem a %s: %v", next, err)
		}
	}

	_, err := env.manager.Estimate(session.ID, 5)

	if !errors.Is(err, backlog.ErrItemFinished) {
		t.Fatalf("error = %v, se esperaba backlog.ErrItemFinished", err)
	}
	expectStoredStatus(t, env.repo, session.ID, StatusRevealed)
	expectItemPoints(t, env.items, item.ID, 0)

	// RN-20.5: un ítem reabierto vuelve a ser estimable.
	if _, err := env.items.ChangeStatus(item.ID, string(backlog.StatusInProgress)); err != nil {
		t.Fatalf("reabrir el ítem: %v", err)
	}
	if _, err := env.manager.Estimate(session.ID, 5); err != nil {
		t.Fatalf("Estimate después de reabrir: %v", err)
	}
	expectStoredStatus(t, env.repo, session.ID, StatusClosed)
	expectItemPoints(t, env.items, item.ID, 5)
}

// revealFailingSession revela una sesión sobre un repositorio de poker que se
// puede hacer fallar. openFailingSession ya registra el voto de Ana (5); acá
// se agrega el de Bruno (8), así los votos son los de revealedVotes.
func revealFailingSession(t *testing.T) (*failingRevealRepository, *Manager, *backlog.Manager, Session) {
	t.Helper()
	repo, manager, session := openFailingSession(t)
	voteAll(t, manager, session.ID, Vote{Voter: "Bruno", Points: 8})
	if _, err := manager.Reveal(session.ID); err != nil {
		t.Fatalf("revelar: %v", err)
	}
	return repo, manager, manager.items, session
}

// CA-25.4 · Si falla la lectura de la sesión, se propaga el error y el ítem
// no cambia.
func TestEstimate_ErrorAlLeerPropagaError(t *testing.T) {
	repo, manager, items, session := revealFailingSession(t)
	repo.failGet = true

	_, err := manager.Estimate(session.ID, 5)

	if !errors.Is(err, errRepository) {
		t.Fatalf("error = %v, se esperaba errRepository", err)
	}
	expectItemPoints(t, items, session.ItemID, 0)
}

// failingItemRepository permite forzar un fallo del backlog al actualizar.
type failingItemRepository struct {
	*backlog.MemoryRepository
	failUpdate bool
}

func (r *failingItemRepository) Update(item backlog.Item) (backlog.Item, error) {
	if r.failUpdate {
		return backlog.Item{}, errRepository
	}
	return r.MemoryRepository.Update(item)
}

// CA-25.5 · Si falla la asignación en el backlog, se propaga el error y la
// sesión sigue REVELADA.
func TestEstimate_ErrorDelBacklogAlAsignarPropagaErrorYSesionSigueRevelada(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	p, err := projects.Create("Proyecto de prueba", "Proyecto para HU-25")
	if err != nil {
		t.Fatalf("crear proyecto: %v", err)
	}
	itemRepo := &failingItemRepository{MemoryRepository: backlog.NewMemoryRepository()}
	items := backlog.NewManager(itemRepo, projects)
	item, err := items.Create(p.ID, "Historia de ejemplo", "Descripción", "MEDIA", []string{"Criterio"})
	if err != nil {
		t.Fatalf("crear ítem: %v", err)
	}
	repo := NewMemoryRepository()
	manager := NewManager(repo, items)
	session, err := manager.OpenSession(item.ID)
	if err != nil {
		t.Fatalf("abrir sesión: %v", err)
	}
	voteAll(t, manager, session.ID, revealedVotes...)
	if _, err := manager.Reveal(session.ID); err != nil {
		t.Fatalf("revelar: %v", err)
	}
	itemRepo.failUpdate = true

	_, err = manager.Estimate(session.ID, 5)

	if !errors.Is(err, errRepository) {
		t.Fatalf("error = %v, se esperaba errRepository", err)
	}
	expectStoredStatus(t, repo, session.ID, StatusRevealed)
	expectItemPoints(t, items, item.ID, 0)
}

// CA-25.6 · Si falla el guardado de la sesión, el ítem queda estimado y la
// sesión sigue REVELADA; repetir el registro la cierra (RN-25.7).
func TestEstimate_ErrorAlGuardarLaSesionDejaElItemEstimadoYLaSesionRevelada(t *testing.T) {
	repo, manager, items, session := revealFailingSession(t)
	repo.failUpdate = true

	_, err := manager.Estimate(session.ID, 5)

	if !errors.Is(err, errRepository) {
		t.Fatalf("error = %v, se esperaba errRepository", err)
	}
	expectItemPoints(t, items, session.ItemID, 5)
	expectStoredStatus(t, repo.MemoryRepository, session.ID, StatusRevealed)

	repo.failUpdate = false
	if _, err := manager.Estimate(session.ID, 5); err != nil {
		t.Fatalf("repetir el registro: %v", err)
	}
	expectStoredStatus(t, repo.MemoryRepository, session.ID, StatusClosed)
	expectItemPoints(t, items, session.ItemID, 5)
}

// CA-25.3 · Registrar conserva la ronda, el ítem y el orden de los votos.
func TestEstimate_ConservaRondaYOrdenDeLosVotos(t *testing.T) {
	env, session := openTestSession(t)
	want := []Vote{{Voter: "Carla", Points: 13}, {Voter: "Ana", Points: 2}, {Voter: "Bruno", Points: 21}}
	voteAll(t, env.manager, session.ID, want...)
	if _, err := env.manager.Reveal(session.ID); err != nil {
		t.Fatalf("revelar: %v", err)
	}

	got, err := env.manager.Estimate(session.ID, 8)

	if err != nil {
		t.Fatalf("Estimate devolvió error: %v", err)
	}
	if got.ItemID != session.ItemID || got.Round != session.Round || !slices.Equal(got.Votes, want) {
		t.Errorf("sesión devuelta = %+v, se esperaba el ítem %d en la ronda %d con %+v", got, session.ItemID, session.Round, want)
	}
	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Round != session.Round || !slices.Equal(stored.Votes, want) {
		t.Errorf("sesión guardada = %+v, se esperaba la ronda %d con %+v", stored, session.Round, want)
	}
}

// CA-25.7 · Modificar la sesión devuelta no altera la guardada.
func TestEstimate_DevuelveCopiaDeLosVotos(t *testing.T) {
	env := newTestEnv(t)
	_, session := revealTestSession(t, env, 0)

	got, err := env.manager.Estimate(session.ID, 5)
	if err != nil {
		t.Fatalf("Estimate devolvió error: %v", err)
	}
	if len(got.Votes) != len(revealedVotes) {
		t.Fatalf("votos devueltos = %+v, se esperaba %+v", got.Votes, revealedVotes)
	}
	got.Votes[0].Points = 21
	got.Votes = append(got.Votes, Vote{Voter: "Intruso", Points: 1})

	stored, err := env.repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !slices.Equal(stored.Votes, revealedVotes) {
		t.Errorf("votos guardados = %+v, la mutación externa no debía afectarlos", stored.Votes)
	}
}

// CA-25.3 · Una sesión CERRADA no acepta votos ni se vuelve a revelar.
func TestVoteYReveal_SesionCerradaDevuelvenErrSessionNotOpen(t *testing.T) {
	env := newTestEnv(t)
	_, session := revealTestSession(t, env, 0)
	if _, err := env.manager.Estimate(session.ID, 5); err != nil {
		t.Fatalf("Estimate: %v", err)
	}

	if _, err := env.manager.Vote(session.ID, "Carla", 3); !errors.Is(err, ErrSessionNotOpen) {
		t.Errorf("Vote: error = %v, se esperaba ErrSessionNotOpen", err)
	}
	if _, err := env.manager.Reveal(session.ID); !errors.Is(err, ErrSessionNotOpen) {
		t.Errorf("Reveal: error = %v, se esperaba ErrSessionNotOpen", err)
	}
	expectStoredStatus(t, env.repo, session.ID, StatusClosed)
}

// CA-25.8 · Dos registros simultáneos sobre la misma sesión: solo uno se
// aplica y el ítem queda con sus puntos. Se corre con -race.
func TestEstimate_SimultaneosSobreLaMismaSesionSoloUnoSeAplica(t *testing.T) {
	const attempts = 20
	for i := 0; i < attempts; i++ {
		env := newTestEnv(t)
		item, session := revealTestSession(t, env, 0)

		var (
			wg   sync.WaitGroup
			mu   sync.Mutex
			errs = map[int]error{}
		)
		start := make(chan struct{})
		for _, points := range []int{5, 8} {
			wg.Add(1)
			go func(points int) {
				defer wg.Done()
				<-start
				_, err := env.manager.Estimate(session.ID, points)
				mu.Lock()
				defer mu.Unlock()
				errs[points] = err
			}(points)
		}
		close(start)
		wg.Wait()

		winner := 0
		for points, err := range errs {
			switch {
			case err == nil:
				if winner != 0 {
					t.Fatalf("intento %d: se aplicaron los dos registros", i)
				}
				winner = points
			case !errors.Is(err, ErrSessionNotRevealed):
				t.Fatalf("intento %d: registro de %d puntos falló con %v, se esperaba ErrSessionNotRevealed", i, points, err)
			}
		}
		if winner == 0 {
			t.Fatalf("intento %d: ningún registro se aplicó (%v)", i, errs)
		}
		expectStoredStatus(t, env.repo, session.ID, StatusClosed)
		expectItemPoints(t, env.items, item.ID, winner)
	}
}
