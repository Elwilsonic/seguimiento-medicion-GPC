package member

import "errors"

// ErrMemberNotFound indica que no existe un integrante con ese ID (RN-16.3).
var ErrMemberNotFound = errors.New("no existe un integrante con ese id")

// errNotImplemented es el error de las firmas vacías del RED de HU-16.
var errNotImplemented = errors.New("sin implementar")

// Update modifica el nombre y el rol de un integrante (HU-16).
func (m *Manager) Update(id int, name, role string) (Member, error) {
	return Member{}, errNotImplemented
}
