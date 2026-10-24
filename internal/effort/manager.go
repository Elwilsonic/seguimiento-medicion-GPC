package effort

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Manager aplica las reglas de negocio de los esfuerzos.
type Manager struct {
	repo    Repository
	sprints *sprint.Manager
	items   *backlog.Manager
}

func NewManager(repo Repository, sprints *sprint.Manager, items *backlog.Manager) *Manager {
	return &Manager{repo: repo, sprints: sprints, items: items}
}
