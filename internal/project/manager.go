package project

import "sync"

// Manager aplica las reglas de negocio de los proyectos sobre un Repository.
// Cada historia agrega sus operaciones en su propio archivo.
type Manager struct {
	// mu protege las operaciones que leen y escriben el repositorio en
	// varios pasos, para que dos pedidos simultáneos no se pisen (CA-01.7).
	mu   sync.Mutex
	repo Repository
}

// NewManager crea un Manager que guarda los proyectos en repo.
func NewManager(repo Repository) *Manager {
	return &Manager{repo: repo}
}
