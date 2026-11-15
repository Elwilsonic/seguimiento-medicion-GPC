package defect

import "uuid"

// GetByPublicID devuelve una copia del defecto con ese identificador público
// (HU-52, D4), o ErrDefectNotFound si no existe.
func (m *Manager) GetByPublicID(id uuid.UUID) (Defect, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.GetByPublicID(id)
}
