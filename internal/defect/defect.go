package defect

import "errors"

// Status es el estado de un defecto (RN-29.3). Un defecto se registra en
// StatusOpen (HU-29); las transiciones entre estados están en status.go
// (HU-30).
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

var statusValues = [...]string{"ABIERTO", "EN_CORRECCION", "RESUELTO"}

// StatusValues devuelve una lista nueva con los estados válidos, en orden
// (RN-30.1). La lista se puede modificar sin alterar la validación del paquete.
func StatusValues() []string {
	values := make([]string, len(statusValues))
	copy(values, statusValues[:])
	return values
}

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
	// ErrDefectNotFound indica que no existe un defecto con ese ID, incluidos
	// los IDs 0 y negativos (RN-30.3).
	ErrDefectNotFound = errors.New("no existe un defecto con ese id")
	// ErrInvalidStatus indica que el estado no es ABIERTO, EN_CORRECCION ni
	// RESUELTO (RN-30.1).
	ErrInvalidStatus = errors.New("el estado debe ser ABIERTO, EN_CORRECCION o RESUELTO")
	// ErrInvalidTransition indica que la transición de estado no es de un
	// paso (RN-30.2).
	ErrInvalidTransition = errors.New("la transición de estado del defecto no está permitida")
	// ErrMissingResolutionSprint indica que el destino es RESUELTO y el
	// sprint de resolución llegó ausente, 0 o negativo (RN-30.5).
	ErrMissingResolutionSprint = errors.New("el sprint de resolución es obligatorio para resolver un defecto")
	// ErrResolutionSprintFromAnotherProject indica que el sprint de
	// resolución pertenece a otro proyecto distinto del de la historia
	// (RN-30.6).
	ErrResolutionSprintFromAnotherProject = errors.New("el sprint de resolución pertenece a otro proyecto")
	// ErrResolutionBeforeDetection indica que el sprint de resolución es
	// anterior al de detección (RN-30.6). Es una validación defensiva: desde
	// la API no se puede producir.
	ErrResolutionBeforeDetection = errors.New("el sprint de resolución no puede ser anterior al de detección")
)
