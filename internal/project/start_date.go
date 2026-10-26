package project

import (
	"errors"
	"time"
)

// ErrMissingStartDate indica que la fecha de inicio está ausente (valor cero
// de time.Time) (RN-18.1, HU-18).
var ErrMissingStartDate = errors.New("la fecha de inicio del proyecto es obligatoria")

// SetStartDate registra la fecha de inicio de un proyecto existente (HU-18) y
// devuelve una copia.
func (m *Manager) SetStartDate(id int, date time.Time) (Project, error) {
	return Project{}, nil
}
