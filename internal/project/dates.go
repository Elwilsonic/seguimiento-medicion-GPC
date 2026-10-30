package project

import (
	"errors"
	"time"
)

// ErrDatesNotRegistered indica que el proyecto no tiene registradas la fecha
// de inicio y la de fin, así que no hay fechas que modificar (RN-19.4, HU-19).
var ErrDatesNotRegistered = errors.New("el proyecto no tiene registradas las fechas de inicio y de fin")

// SetDates modifica juntas las fechas de inicio y de fin de un proyecto que ya
// las tiene registradas (HU-19). Firma vacía del RED: se implementa en el
// GREEN.
func (m *Manager) SetDates(id int, start, end time.Time) (Project, error) {
	return Project{}, nil
}
