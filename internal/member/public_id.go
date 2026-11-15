package member

import "uuid"

// GetByPublicID devuelve una copia del integrante con ese identificador público
// (HU-52, D4), o ErrMemberNotFound si no existe.
func (m *Manager) GetByPublicID(id uuid.UUID) (Member, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.GetByPublicID(id)
}
