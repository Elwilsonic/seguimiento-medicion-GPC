package defect

import "errors"

// Status es el estado de un defecto (RN-29.3). HU-29 solo usa StatusOpen;
// StatusInCorrection y StatusResolved quedan definidos para HU-30.
type Status string

const (
	StatusOpen         Status = "ABIERTO"
	StatusInCorrection Status = "EN_CORRECCION"
	StatusResolved     Status = "RESUELTO"
)

// Severity es la severidad de un defecto (RN-29.2).
type Severity string

const (
	SeverityCritical Severity = "CRITICA"
	SeverityHigh     Severity = "ALTA"
	SeverityMedium   Severity = "MEDIA"
	SeverityLow      Severity = "BAJA"
)

var severityValues = [...]string{"CRITICA", "ALTA", "MEDIA", "BAJA"}

// SeverityValues devuelve una lista nueva con las severidades válidas. La
// lista se puede modificar sin alterar la validación del paquete.
func SeverityValues() []string {
	values := make([]string, len(severityValues))
	copy(values, severityValues[:])
	return values
}

// Defect es un defecto asociado a una historia del backlog. No tiene slices
// ni mapas, así que una copia no comparte estado con el defecto guardado.
type Defect struct {
	ID                 int
	Description        string
	Severity           Severity
	Status             Status
	ItemID             int
	DetectionSprintID  int
	ResolutionSprintID int
}

var (
	// ErrEmptyDescription indica que la descripción está vacía o tiene solo
	// espacios (RN-29.1).
	ErrEmptyDescription = errors.New("la descripción del defecto es obligatoria")
	// ErrInvalidSeverity indica que la severidad no es CRITICA, ALTA, MEDIA
	// ni BAJA (RN-29.2).
	ErrInvalidSeverity = errors.New("la severidad debe ser CRITICA, ALTA, MEDIA o BAJA")
	// ErrMissingDetectionSprint indica que el sprint de detección es
	// obligatorio y llegó ausente, 0 o negativo (RN-29.6).
	ErrMissingDetectionSprint = errors.New("el sprint de detección es obligatorio")
	// ErrDetectionSprintFromAnotherProject indica que el sprint de detección
	// pertenece a otro proyecto distinto del de la historia (RN-29.6).
	ErrDetectionSprintFromAnotherProject = errors.New("el sprint de detección pertenece a otro proyecto")
)
