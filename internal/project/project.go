package project

import "errors"

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
	ID          int
	Name        string
	Description string
	Status      Status
}

var (
	// ErrEmptyName indica que el nombre está vacío o tiene solo espacios (RN-01.1).
	ErrEmptyName = errors.New("el nombre del proyecto es obligatorio")
	// ErrDuplicateName indica que ya existe un proyecto con ese nombre (RN-01.3).
	ErrDuplicateName = errors.New("ya existe un proyecto con ese nombre")
)
