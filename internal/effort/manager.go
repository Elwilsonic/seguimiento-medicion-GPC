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
	members Members
}

// Members resuelve el nombre registrado de un integrante del proyecto, o
// devuelve un error si no lo es (HU-54, D2). Lo cumple *member.Manager.
type Members interface {
	CanonicalName(projectID int, name string) (string, error)
}

func NewManager(repo Repository, sprints *sprint.Manager, items *backlog.Manager, members Members) *Manager {
	return &Manager{repo: repo, sprints: sprints, items: items, members: members}
}
