package metrics

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Manager calcula las métricas leyendo los ítems y los sprints. No tiene
// estado propio, así que no necesita un mutex: cada lectura queda protegida
// por el mutex del Manager que la hace.
type Manager struct {
	items   *backlog.Manager
	sprints *sprint.Manager
}

// NewManager crea un Manager que lee los ítems con items y los sprints con
// sprints.
func NewManager(items *backlog.Manager, sprints *sprint.Manager) *Manager {
	return &Manager{items: items, sprints: sprints}
}
