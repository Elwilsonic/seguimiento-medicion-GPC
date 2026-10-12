package backlog

import (
	"sync"

	"seguimiento-medicion-gpc/internal/project"
)

// Manager aplica las reglas de negocio de los ítems del backlog sobre un
// Repository. Cada historia agrega sus operaciones en su propio archivo.
type Manager struct {
	// mu protege las operaciones que escriben el repositorio, para que dos
	// pedidos simultáneos no se pisen (CA-04.7).
	mu       sync.Mutex
	repo     Repository
	projects *project.Manager
}

// NewManager crea un Manager que guarda los ítems en repo y valida los
// proyectos con projects (HU-04).
func NewManager(repo Repository, projects *project.Manager) *Manager {
	return &Manager{repo: repo, projects: projects}
}
