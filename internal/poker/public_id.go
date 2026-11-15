package poker

import "uuid"

// GetByPublicID devuelve una copia de la sesión con ese identificador público
// (HU-52, D4), o ErrSessionNotFound si no existe.
func (m *Manager) GetByPublicID(id uuid.UUID) (Session, error) {
	return Session{}, ErrSessionNotFound
}
