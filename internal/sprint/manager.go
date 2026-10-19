package sprint

import (
	"sync"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// Manager aplica las reglas de negocio de los sprints sobre un Repository.
// Cada historia agrega sus operaciones en su propio archivo.
type Manager struct {
	// mu protege las operaciones que consultan y escriben el repositorio,
	// para que dos pedidos simultáneos no dejen dos sprints abiertos en el
	// mismo proyecto (CA-10.6).
	mu       sync.Mutex
	repo     Repository
	projects *project.Manager
	items    *backlog.Manager
}

// NewManager crea un Manager que guarda los sprints en repo, valida los
// proyectos con projects (HU-10) y asigna ítems del backlog con items
// (HU-11).
func NewManager(repo Repository, projects *project.Manager, items *backlog.Manager) *Manager {
	return &Manager{repo: repo, projects: projects, items: items}
}
