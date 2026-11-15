package project

import (
	"errors"
	"uuid"
)

// ErrProjectNotFound indica que no existe un proyecto con el ID consultado
// (RN-03.2, HU-03).
var ErrProjectNotFound = errors.New("no existe un proyecto con ese id")

// Get devuelve una copia del proyecto con ese ID (HU-03).
func (m *Manager) Get(id int) (Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.GetByID(id)
}

// GetByPublicID devuelve una copia del proyecto con ese identificador público
// (HU-51, D4).
// HU-51 (RED): stub sin lógica.
func (m *Manager) GetByPublicID(id uuid.UUID) (Project, error) {
	return Project{}, ErrProjectNotFound
}
