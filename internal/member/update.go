package member

import (
	"errors"
	"strings"
)

// ErrMemberNotFound indica que no existe un integrante con ese ID (RN-16.3).
var ErrMemberNotFound = errors.New("no existe un integrante con ese id")

// Update modifica el nombre y el rol de un integrante (HU-16) y devuelve una
// copia. El ID y el proyecto no cambian (RN-16.5). El orden de validación es:
// nombre, rol, existencia del integrante, duplicado y guardado; se informa el
// primer error encontrado y, si algo falla, no se guarda nada (RN-16.7).
func (m *Manager) Update(id int, name, role string) (Member, error) {
	// RN-16.1: la misma regla del nombre que en el registro (RN-15.1).
	name = strings.TrimSpace(name)
	if name == "" {
		return Member{}, ErrEmptyName
	}

	normalized, err := normalizeRole(role)
	if err != nil {
		return Member{}, err
	}

	// RN-16.10: la búsqueda, la comparación y el guardado van juntos bajo el
	// mismo mutex de Register, así no quedan dos integrantes con el mismo
	// nombre en el proyecto.
	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := m.repo.GetByID(id)
	if err != nil {
		return Member{}, err
	}

	members, err := m.repo.ListByProject(current.ProjectID)
	if err != nil {
		return Member{}, err
	}
	// RN-16.4: el integrante no se compara consigo mismo.
	if nameTaken(members, name, id) {
		return Member{}, ErrDuplicateMember
	}

	current.Name = name
	current.Role = normalized
	return m.repo.Update(current)
}

// nameTaken indica si otro integrante de la lista, distinto de exceptID, ya
// tiene ese nombre, sin distinguir mayúsculas (RN-15.4, D1 de HU-15).
// Register pasa 0, que no es el ID de ningún integrante (RN-15.7).
func nameTaken(members []Member, name string, exceptID int) bool {
	for _, other := range members {
		if other.ID != exceptID && strings.EqualFold(other.Name, name) {
			return true
		}
	}
	return false
}
