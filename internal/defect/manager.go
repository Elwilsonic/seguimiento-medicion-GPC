package defect

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Manager aplica las reglas de negocio de los defectos sobre un Repository.
type Manager struct {
	repo    Repository
	items   *backlog.Manager
	sprints *sprint.Manager
}

// NewManager crea un Manager que guarda los defectos en repo y valida la
// historia y el sprint de detección con items y sprints.
func NewManager(repo Repository, items *backlog.Manager, sprints *sprint.Manager) *Manager {
	return &Manager{repo: repo, items: items, sprints: sprints}
}
