package effort

import "uuid"

// GetByPublicID devuelve una copia del esfuerzo con ese identificador público
// (HU-52, D4), o ErrEffortNotFound si no existe.
func (m *Manager) GetByPublicID(id uuid.UUID) (Effort, error) {
	return m.repo.GetByPublicID(id)
}
