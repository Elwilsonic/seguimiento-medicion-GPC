package effort

import (
	"uuid"

	"errors"
	"time"
)

// Effort es un registro de horas trabajadas en una historia de un sprint.
type Effort struct {
	ID int
	// PublicID es el identificador público del esfuerzo (HU-52, D1): el único que
	// muestran la API y la interfaz. ID queda como clave interna.
	PublicID uuid.UUID
	Member   string
	Date     time.Time
	Activity string
	ItemID   int
	SprintID int
	Hours    float64
}

var (
	ErrEmptyMember   = errors.New("el integrante es obligatorio")
	ErrEmptyActivity = errors.New("la actividad es obligatoria")
	ErrMissingDate   = errors.New("la fecha del esfuerzo es obligatoria")
	ErrInvalidHours  = errors.New("las horas trabajadas deben ser un número positivo")
	// ErrEffortNotFound indica que no existe un esfuerzo con ese ID, incluidos
	// los IDs 0 y negativos (HU-28).
	ErrEffortNotFound = errors.New("no existe un esfuerzo con ese id")
)
