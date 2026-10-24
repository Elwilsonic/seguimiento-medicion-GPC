package poker

import (
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
