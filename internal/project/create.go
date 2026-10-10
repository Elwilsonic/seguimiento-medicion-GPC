package project

import (
	"strings"
	"sync"
)

// Manager aplica las reglas de negocio de los proyectos sobre un Repository.
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

// Create crea un proyecto en estado PLANIFICADO (HU-01) y devuelve una copia.
func (m *Manager) Create(name, description string) (Project, error) {
	// RN-01.1 y RN-01.2: el nombre se guarda sin espacios, tabs ni saltos de
	// línea en los extremos, y no puede quedar vacío.
	name = strings.TrimSpace(name)
	if name == "" {
		return Project{}, ErrEmptyName
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	projects, err := m.repo.List()
	if err != nil {
		return Project{}, err
	}
	// RN-01.3: los nombres se comparan sin distinguir mayúsculas.
	for _, p := range projects {
		if strings.EqualFold(p.Name, name) {
			return Project{}, ErrDuplicateName
		}
	}

	return m.repo.Save(Project{Name: name, Description: description, Status: StatusPlanned})
}
