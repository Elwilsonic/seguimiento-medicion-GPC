package project

import "strings"

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
