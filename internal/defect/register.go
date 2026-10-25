package defect

import "errors"

// errNotImplemented es el error temporal de la fase RED.
var errNotImplemented = errors.New("no implementado")

// Register registra un defecto asociado a una historia del backlog.
//
// Firma de HU-29: el estado inicial y el sprint de resolución no se reciben.
// En la fase RED devuelve siempre errNotImplemented.
func (m *Manager) Register(itemID int, description, severity string, detectionSprintID int) (Defect, error) {
	return Defect{}, errNotImplemented
}
