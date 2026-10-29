package project

import (
	"errors"
	"time"
)

var (
	// ErrMissingEndDate indica que la fecha de fin está ausente (valor cero
	// de time.Time) (RN-41.1, HU-41).
	ErrMissingEndDate = errors.New("la fecha de fin del proyecto es obligatoria")
	// ErrStartDateNotRegistered indica que el proyecto no tiene fecha de
	// inicio con la que comparar la de fin (RN-41.4).
	ErrStartDateNotRegistered = errors.New("el proyecto no tiene fecha de inicio registrada")
	// ErrEndNotAfterStart indica que la fecha de fin es igual o anterior a la
	// de inicio (RN-41.5).
	ErrEndNotAfterStart = errors.New("la fecha de fin debe ser posterior a la de inicio")
	// ErrStartNotBeforeEnd indica que la fecha de inicio es igual o posterior
	// a la fecha de fin registrada (RN-41.7).
	ErrStartNotBeforeEnd = errors.New("la fecha de inicio debe ser anterior a la de fin")
)

// SetEndDate registra la fecha de fin de un proyecto existente (HU-41).
func (m *Manager) SetEndDate(id int, date time.Time) (Project, error) {
	return Project{}, nil
}
