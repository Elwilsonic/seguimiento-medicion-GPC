package member

import "strings"

// CanonicalName devuelve el nombre registrado del integrante del proyecto
// cuyo nombre coincide con name, sin espacios en los extremos y sin distinguir
// mayúsculas, la misma regla de los duplicados (RN-15.4). Si no hay ninguno,
// devuelve ErrNotProjectMember (HU-54, D1). La usan Planning Poker y Esfuerzo
// para validar al votante y al integrante del esfuerzo.
func (m *Manager) CanonicalName(projectID int, name string) (string, error) {
	name = strings.TrimSpace(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	members, err := m.repo.ListByProject(projectID)
	if err != nil {
		return "", err
	}
	for _, other := range members {
		if strings.EqualFold(other.Name, name) {
			return other.Name, nil
		}
	}
	return "", ErrNotProjectMember
}
