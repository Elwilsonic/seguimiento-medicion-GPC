package sprint

import "uuid"

// GetByPublicID devuelve una copia del sprint con ese identificador público
// (HU-52, D4), o ErrSprintNotFound si no existe.
func (m *Manager) GetByPublicID(id uuid.UUID) (Sprint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.GetByPublicID(id)
}
