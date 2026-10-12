package backlog

import "errors"

// Priority es la prioridad de un ítem del backlog (RN-04.3).
type Priority string

const (
	PriorityHigh   Priority = "ALTA"
	PriorityMedium Priority = "MEDIA"
	PriorityLow    Priority = "BAJA"
)

// Status es el estado de un ítem del backlog (RN-04.7). Las transiciones
// entre estados las define HU-08.
type Status string

const (
	StatusPending    Status = "PENDIENTE"
	StatusInProgress Status = "EN_CURSO"
	StatusFinished   Status = "TERMINADO"
)

// Item es un ítem del Product Backlog. Copiar un Item no copia la lista de
// criterios de aceptación: el repositorio y el Manager siempre devuelven
// listas nuevas (RN-04.9).
type Item struct {
	ID                 int
	ProjectID          int
	Title              string
	Description        string
	Priority           Priority
	AcceptanceCriteria []string
	Status             Status
	StoryPoints        int
}

var (
	// ErrEmptyTitle indica que el título está vacío o tiene solo espacios (RN-04.1).
	ErrEmptyTitle = errors.New("el título del ítem es obligatorio")
	// ErrInvalidPriority indica que la prioridad no es ALTA, MEDIA ni BAJA (RN-04.3).
	ErrInvalidPriority = errors.New("la prioridad debe ser ALTA, MEDIA o BAJA")
	// ErrNoAcceptanceCriteria indica que no quedó ningún criterio de aceptación
	// no vacío después de descartar los vacíos (RN-04.4).
	ErrNoAcceptanceCriteria = errors.New("el ítem necesita al menos un criterio de aceptación")
)
