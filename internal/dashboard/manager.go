package dashboard

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Manager arma el dashboard leyendo los proyectos, los ítems y los sprints.
// No tiene estado propio, así que no necesita un mutex: cada lectura queda
// protegida por el mutex del Manager que la hace.
type Manager struct {
	projects *project.Manager
	items    *backlog.Manager
	sprints  *sprint.Manager
}

// NewManager crea un Manager que lee los proyectos con projects, los ítems
// con items y los sprints con sprints.
func NewManager(projects *project.Manager, items *backlog.Manager, sprints *sprint.Manager) *Manager {
	return &Manager{projects: projects, items: items, sprints: sprints}
}
