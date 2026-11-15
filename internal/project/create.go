package project

import (
	"strings"
	"uuid"
)

// Create crea un proyecto en estado PLANIFICADO (HU-01) y devuelve una copia.
func (m *Manager) Create(name, description string) (Project, error) {
	// RN-01.1 y RN-01.2: el nombre se guarda sin espacios en blanco Unicode
	// en los extremos (strings.TrimSpace), y no puede quedar vacío.
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

	// HU-51 (D3): el identificador público es un UUID versión 4 aleatorio, que
	// se genera con el mutex tomado (CA-51.7).
	return m.repo.Save(Project{PublicID: uuid.NewV4(), Name: name, Description: description, Status: StatusPlanned})
}
