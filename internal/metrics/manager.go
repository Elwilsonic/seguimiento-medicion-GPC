package metrics

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Manager calcula las métricas leyendo los ítems, los sprints y los esfuerzos.
// No tiene estado propio, así que no necesita un mutex: cada lectura queda
// protegida por el mutex del Manager que la hace.
type Manager struct {
	items   *backlog.Manager
	sprints *sprint.Manager
	efforts *effort.Manager
}

// NewManager crea un Manager que lee los ítems con items, los sprints con
// sprints y los esfuerzos con efforts.
func NewManager(items *backlog.Manager, sprints *sprint.Manager, efforts *effort.Manager) *Manager {
	return &Manager{items: items, sprints: sprints, efforts: efforts}
}
