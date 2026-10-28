package defect

import (
	"strings"

	"seguimiento-medicion-gpc/internal/sprint"
)

// allowedTransitions define los cambios permitidos: un paso por vez, hacia
// adelante o hacia atrás, como HU-08 (RN-30.2).
var allowedTransitions = map[Status][]Status{
	StatusOpen:         {StatusInCorrection},
	StatusInCorrection: {StatusResolved, StatusOpen},
	StatusResolved:     {StatusInCorrection},
}

// ChangeStatus cambia el estado de un defecto existente y registra el sprint
// de resolución (HU-30). Valida en el orden de la SDD: sin tocar repositorios,
// el estado (RN-30.1) y el sprint de resolución <= 0 al resolver (RN-30.5);
// con el mutex, la existencia (RN-30.3), el mismo estado (RN-30.4) y la
// transición (RN-30.2). Al resolver guarda dentro de WithOpenSprint (RN-30.6
// y RN-30.11); con otro destino el sprint de resolución queda en 0 (RN-30.7).
func (m *Manager) ChangeStatus(id int, status string, resolutionSprintID int) (Defect, error) {
	normalized, err := normalizeStatus(status)
	if err != nil {
		return Defect{}, err
	}
	if normalized == StatusResolved && resolutionSprintID <= 0 {
		return Defect{}, ErrMissingResolutionSprint
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := m.repo.GetByID(id)
	if err != nil {
		return Defect{}, err
	}

	// RN-30.4: el mismo estado es un no-op válido, sin llamar a Update.
	if current.Status == normalized {
		return current, nil
	}

	if !isTransitionAllowed(current.Status, normalized) {
		return Defect{}, ErrInvalidTransition
	}

	if normalized != StatusResolved {
		current.Status = normalized
		current.ResolutionSprintID = 0
		return m.repo.Update(current)
	}

	var updated Defect
	err = m.sprints.WithOpenSprint(resolutionSprintID, func(sp sprint.Sprint) error {
		item, err := m.items.GetItem(current.ItemID)
		if err != nil {
			return err
		}
		if item.ProjectID != sp.ProjectID {
			return ErrResolutionSprintFromAnotherProject
		}
		// Validación defensiva: dentro de un proyecto los IDs siguen el orden
		// de creación de los sprints (RN-30.6).
		if resolutionSprintID < current.DetectionSprintID {
			return ErrResolutionBeforeDetection
		}

		current.Status = StatusResolved
		current.ResolutionSprintID = resolutionSprintID
		updated, err = m.repo.Update(current)
		return err
	})
	if err != nil {
		return Defect{}, err
	}
	return updated, nil
}

// normalizeStatus valida el estado destino contra ABIERTO, EN_CORRECCION y
// RESUELTO, sin distinguir mayúsculas ni contar los espacios de los extremos
// (RN-30.1).
func normalizeStatus(status string) (Status, error) {
	normalized := strings.ToUpper(strings.TrimSpace(status))
	for _, value := range statusValues {
		if normalized == value {
			return Status(normalized), nil
		}
	}
	return "", ErrInvalidStatus
}

// isTransitionAllowed indica si se puede pasar de from a to (RN-30.2).
func isTransitionAllowed(from, to Status) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
