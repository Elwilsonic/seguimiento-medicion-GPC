package project

import (
	"errors"
	"time"
	"uuid"
)

// Status es el estado de un proyecto (RN-01.6).
type Status string

const (
	StatusPlanned    Status = "PLANIFICADO"
	StatusInProgress Status = "EN_CURSO"
	StatusFinished   Status = "FINALIZADO"
)

// Project es un proyecto. Todos sus campos son valores, así que una copia
// no comparte estado con el proyecto guardado.
type Project struct {
	ID int
	// PublicID es el identificador público del proyecto (HU-51, D1): el único
	// que muestran la API y la interfaz. ID queda como clave interna.
	PublicID    uuid.UUID
	Name        string
	Description string
	Status      Status
	// StartDate es la fecha de inicio del proyecto (HU-18); el valor cero
	// significa que no está registrada (RN-18.7).
	StartDate time.Time
	// EndDate es la fecha de fin del proyecto (HU-41); el valor cero
	// significa que no está registrada (RN-41.10).
	EndDate time.Time
}

var (
	// ErrEmptyName indica que el nombre está vacío o tiene solo espacios (RN-01.1).
	ErrEmptyName = errors.New("el nombre del proyecto es obligatorio")
	// ErrDuplicateName indica que ya existe un proyecto con ese nombre (RN-01.3).
	ErrDuplicateName = errors.New("ya existe un proyecto con ese nombre")
)
