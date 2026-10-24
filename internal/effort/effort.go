package effort

import (
	"errors"
	"time"
)

// Effort es un registro de horas trabajadas en una historia de un sprint.
type Effort struct {
	ID       int
	Member   string
	Date     time.Time
	Activity string
	ItemID   int
	SprintID int
	Hours    float64
}

var (
	ErrEmptyMember    = errors.New("el integrante es obligatorio")
	ErrEmptyActivity  = errors.New("la actividad es obligatoria")
	ErrMissingDate    = errors.New("la fecha del esfuerzo es obligatoria")
	ErrInvalidHours   = errors.New("las horas trabajadas deben ser un número positivo")
	errNotImplemented = errors.New("no implementado")
)
