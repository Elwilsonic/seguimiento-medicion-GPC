package poker

import (
	"errors"
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

// Vote registra un voto de forma oculta. La implementación se completará en
// la fase GREEN de HU-22.
func (m *Manager) Vote(_ int, _ string, _ int) (Session, error) {
	return Session{}, errors.New("Vote no implementado")
}
