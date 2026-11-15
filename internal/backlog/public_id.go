package backlog

import "uuid"

// GetByPublicID devuelve una copia del ítem con ese identificador público
// (HU-52, D4), o ErrItemNotFound si no existe.
func (m *Manager) GetByPublicID(id uuid.UUID) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.GetByPublicID(id)
}
