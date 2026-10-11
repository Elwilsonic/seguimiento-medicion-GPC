package project

import "errors"

// ErrProjectNotFound indica que no existe un proyecto con el ID consultado
// (RN-03.2, HU-03).
var ErrProjectNotFound = errors.New("no existe un proyecto con ese id")

// Get devuelve una copia del proyecto con ese ID (HU-03).
func (m *Manager) Get(id int) (Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.GetByID(id)
}
