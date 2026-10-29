package poker

import (
	"strings"
	"sync"

	"seguimiento-medicion-gpc/internal/backlog"
)

// Manager aplica las reglas de negocio de las sesiones de Planning Poker sobre
// un Repository. Cada historia agrega sus operaciones en su propio archivo.
type Manager struct {
	// mu protege las operaciones que consultan y escriben el repositorio, para
	// que dos aperturas simultáneas no dejen dos sesiones abiertas para el
	// mismo ítem (CA-21.6).
	mu    sync.Mutex
	repo  Repository
	items *backlog.Manager
}

// NewManager crea un Manager que guarda las sesiones en repo y valida los
// ítems con items (HU-21).
func NewManager(repo Repository, items *backlog.Manager) *Manager {
	return &Manager{repo: repo, items: items}
}

// Vote registra o reemplaza un voto y devuelve la sesión sin exponer sus
// votos.
func (m *Manager) Vote(sessionID int, voter string, points int) (Session, error) {
	voter = strings.TrimSpace(voter)
	if voter == "" {
		return Session{}, ErrEmptyVoter
	}
	if !validCard(points) {
		return Session{}, ErrInvalidCard
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	session, err := m.repo.GetByID(sessionID)
	if err != nil {
		return Session{}, err
	}
	if session.Status != StatusOpen {
		return Session{}, ErrSessionNotOpen
	}

	replaced := false
	for i, vote := range session.Votes {
		if strings.EqualFold(vote.Voter, voter) {
			session.Votes[i] = Vote{Voter: voter, Points: points}
			replaced = true
			break
		}
	}
	if !replaced {
		session.Votes = append(session.Votes, Vote{Voter: voter, Points: points})
	}
	if _, err := m.repo.Update(session); err != nil {
		return Session{}, err
	}
	session.Votes = []Vote{}
	return session, nil
}

// Reveal pasa una sesión ABIERTA con votos a REVELADA y devuelve todos sus
// votos (HU-23). Leer, validar y guardar van bajo el mismo mutex que Vote,
// así un voto simultáneo queda guardado antes de revelar o se rechaza
// (RN-23.8).
func (m *Manager) Reveal(sessionID int) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, err := m.repo.GetByID(sessionID)
	if err != nil {
		return Session{}, err
	}
	if session.Status != StatusOpen {
		return Session{}, ErrSessionNotOpen
	}
	if len(session.Votes) == 0 {
		return Session{}, ErrNoVotes
	}

	session.Status = StatusRevealed
	if _, err := m.repo.Update(session); err != nil {
		return Session{}, err
	}
	return cloneSession(session), nil
}

// Estimate registra la estimación acordada de una sesión REVELADA (HU-25):
// asigna los puntos al ítem con AssignStoryPoints (HU-20) y cierra la sesión.
// Todo va bajo el mutex de poker, así dos registros simultáneos no asignan
// los dos (RN-25.9). El orden de locks es siempre poker → backlog, y backlog
// nunca llama a poker, así que no hay deadlock.
func (m *Manager) Estimate(sessionID int, points int) (Session, error) {
	if !validCard(points) {
		return Session{}, ErrInvalidCard
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	session, err := m.repo.GetByID(sessionID)
	if err != nil {
		return Session{}, err
	}
	if session.Status != StatusRevealed {
		return Session{}, ErrSessionNotRevealed
	}

	// RN-25.7: primero el ítem y después la sesión. Si falla el guardado de la
	// sesión, el ítem ya quedó estimado y la sesión sigue REVELADA, así que
	// repetir el registro es seguro (RN-20.3).
	if _, err := m.items.AssignStoryPoints(session.ItemID, points); err != nil {
		return Session{}, err
	}
	session.Status = StatusClosed
	if _, err := m.repo.Update(session); err != nil {
		return Session{}, err
	}
	return cloneSession(session), nil
}

// RepeatRound repite la ronda de una sesión REVELADA en la misma sesión
// (HU-24): vuelve a ABIERTA, pasa a la ronda siguiente y descarta los votos.
// Todo va bajo el mutex de poker, igual que OpenSession, así una apertura
// simultánea para el mismo ítem no deja dos sesiones ABIERTA (RN-24.11).
func (m *Manager) RepeatRound(sessionID int) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, err := m.repo.GetByID(sessionID)
	if err != nil {
		return Session{}, err
	}
	if session.Status != StatusRevealed {
		return Session{}, ErrSessionNotRevealed
	}
	// RN-24.7: con D10 de HU-23 y HU-25 pudo abrirse otra sesión para el ítem
	// mientras esta estaba REVELADA; reabrirla rompería RN-21.2.
	if _, exists, err := m.repo.OpenByItem(session.ItemID); err != nil {
		return Session{}, err
	} else if exists {
		return Session{}, ErrSessionAlreadyOpen
	}

	session.Status = StatusOpen
	session.Round++
	session.Votes = []Vote{}
	if _, err := m.repo.Update(session); err != nil {
		return Session{}, err
	}
	return cloneSession(session), nil
}

func validCard(points int) bool {
	for _, card := range Cards() {
		if card == points {
			return true
		}
	}
	return false
}
