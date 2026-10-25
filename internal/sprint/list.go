package sprint

import "errors"

// errNotImplemented es el error de los stubs del RED de HU-32.
var errNotImplemented = errors.New("no implementado")

// Get devuelve una copia del sprint con ese ID o ErrSprintNotFound (HU-32).
func (m *Manager) Get(id int) (Sprint, error) {
	return Sprint{}, errNotImplemented
}

// ListByProject devuelve una lista nueva con copias de los sprints de un
// proyecto, ordenados por ID ascendente (HU-32).
func (m *Manager) ListByProject(projectID int) ([]Sprint, error) {
	return nil, errNotImplemented
}
