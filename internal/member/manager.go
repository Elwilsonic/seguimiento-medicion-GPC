package member

import (
	"errors"
	"sync"

	"seguimiento-medicion-gpc/internal/project"
)

// errNotImplemented es el resultado de las firmas vacías del RED de HU-15.
var errNotImplemented = errors.New("HU-15 todavía no está implementada")

// Manager aplica las reglas de los integrantes. Su mutex protege la lectura
// de los integrantes del proyecto y el guardado (RN-15.9).
type Manager struct {
	mu       sync.Mutex
	repo     Repository
	projects *project.Manager
}

// NewManager crea un Manager que guarda los integrantes en repo y valida los
// proyectos con projects.
func NewManager(repo Repository, projects *project.Manager) *Manager {
	return &Manager{repo: repo, projects: projects}
}
