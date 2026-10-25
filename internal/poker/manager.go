package poker

import (
	"errors"
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

// errNotImplemented es el stub del RED de HU-23; se elimina en el GREEN.
var errNotImplemented = errors.New("no implementado")

// Reveal revela todos los votos de una sesión abierta (HU-23). Stub del RED.
func (m *Manager) Reveal(sessionID int) (Session, error) {
	return Session{}, errNotImplemented
}

func validCard(points int) bool {
	for _, card := range Cards() {
		if card == points {
			return true
		}
	}
	return false
}
