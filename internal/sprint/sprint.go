package sprint

import (
	"errors"
	"time"
)

// Status es el estado de un sprint (RN-10.7). El cambio a CERRADO lo define
// HU-13.
type Status string

const (
	StatusOpen   Status = "ABIERTO"
	StatusClosed Status = "CERRADO"
)

// Sprint es un sprint de un proyecto. No tiene slices ni mapas, así que una
// copia no comparte estado con el sprint guardado (RN-10.8).
type Sprint struct {
	ID        int
	ProjectID int
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Goal      string
	Status    Status
}

var (
	// ErrEmptyName indica que el nombre está vacío o tiene solo espacios (RN-10.1).
	ErrEmptyName = errors.New("el nombre del sprint es obligatorio")
	// ErrEmptyGoal indica que el sprint goal está vacío o tiene solo espacios (RN-10.2).
	ErrEmptyGoal = errors.New("el sprint goal es obligatorio")
	// ErrMissingStartDate indica que la fecha de inicio no se recibió (RN-10.3).
	ErrMissingStartDate = errors.New("la fecha de inicio del sprint es obligatoria")
	// ErrMissingEndDate indica que la fecha de fin no se recibió (RN-10.3).
	ErrMissingEndDate = errors.New("la fecha de fin del sprint es obligatoria")
	// ErrEndNotAfterStart indica que la fecha de fin no es posterior a la de inicio (RN-10.3).
	ErrEndNotAfterStart = errors.New("la fecha de fin debe ser posterior a la de inicio")
	// ErrOpenSprintExists indica que ya existe un sprint ABIERTO en el proyecto (RN-10.5).
	ErrOpenSprintExists = errors.New("ya existe un sprint abierto en el proyecto")
	// ErrSprintNotFound indica que no existe un sprint con el ID recibido,
	// incluidos los IDs 0 y negativos (RN-11.1).
	ErrSprintNotFound = errors.New("no existe un sprint con ese id")
	// ErrSprintClosed indica que el sprint está CERRADO y no acepta
	// asignaciones (RN-11.1).
	ErrSprintClosed = errors.New("el sprint está cerrado")
	// ErrItemFromAnotherProject indica que el ítem pertenece a otro proyecto
	// distinto del sprint (RN-11.2).
	ErrItemFromAnotherProject = errors.New("el ítem pertenece a otro proyecto")
)
