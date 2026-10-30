package member

import "strings"

// Register registra un integrante en un proyecto existente (HU-15) y
// devuelve una copia. El orden de validación es: nombre, rol, existencia del
// proyecto, duplicado y guardado; se informa el primer error encontrado y, si
// algo falla, no se guarda nada (RN-15.8).
func (m *Manager) Register(projectID int, name, role string) (Member, error) {
	// RN-15.1: el nombre se guarda sin espacios en blanco en los extremos y
	// no puede quedar vacío.
	name = strings.TrimSpace(name)
	if name == "" {
		return Member{}, ErrEmptyName
	}

	normalized, err := normalizeRole(role)
	if err != nil {
		return Member{}, err
	}

	// RN-15.3 y RN-15.9: el proyecto tiene que existir. Se valida con la
	// búsqueda de HU-03 antes de tomar el mutex, como sprint.Manager.Create,
	// para no tomar dos mutex a la vez.
	if _, err := m.projects.Get(projectID); err != nil {
		return Member{}, err
	}

	// RN-15.9: la lectura, la comparación y el guardado van juntos bajo el
	// mutex, así dos registros simultáneos no dejan un duplicado.
	m.mu.Lock()
	defer m.mu.Unlock()

	members, err := m.repo.ListByProject(projectID)
	if err != nil {
		return Member{}, err
	}
	// RN-15.4: el mismo nombre en el proyecto, sin distinguir mayúsculas y
	// con cualquier rol, es el mismo integrante (D1).
	if nameTaken(members, name, 0) {
		return Member{}, ErrDuplicateMember
	}

	return m.repo.Save(Member{ProjectID: projectID, Name: name, Role: normalized})
}

// normalizeRole valida el rol contra los tres roles válidos, sin distinguir
// mayúsculas ni contar los espacios de los extremos (RN-15.2).
func normalizeRole(role string) (Role, error) {
	normalized := Role(strings.ToUpper(strings.TrimSpace(role)))
	switch normalized {
	case "":
		return "", ErrMissingRole
	case RoleProductArchitect, RoleAgileEnabler, RoleProductBuilder:
		return normalized, nil
	default:
		return "", ErrInvalidRole
	}
}
