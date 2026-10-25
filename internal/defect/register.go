package defect

import (
	"strings"

	"seguimiento-medicion-gpc/internal/sprint"
)

// Register registra un defecto asociado a una historia del backlog. Valida los
// datos en el orden de la SDD y guarda dentro de WithOpenSprint, para que el
// registro sea atómico con el cierre del sprint (RN-29.11).
func (m *Manager) Register(itemID int, description, severity string, detectionSprintID int) (Defect, error) {
	description = strings.TrimSpace(description)
	severity = strings.ToUpper(strings.TrimSpace(severity))

	if description == "" {
		return Defect{}, ErrEmptyDescription
	}
	if !isValidSeverity(severity) {
		return Defect{}, ErrInvalidSeverity
	}
	if detectionSprintID <= 0 {
		return Defect{}, ErrMissingDetectionSprint
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var saved Defect
	err := m.sprints.WithOpenSprint(detectionSprintID, func(sp sprint.Sprint) error {
		item, err := m.items.GetItem(itemID)
		if err != nil {
			return err
		}
		if item.ProjectID != sp.ProjectID {
			return ErrDetectionSprintFromAnotherProject
		}

		saved, err = m.repo.Save(Defect{
			Description:        description,
			Severity:           Severity(severity),
			Status:             StatusOpen,
			ItemID:             itemID,
			DetectionSprintID:  detectionSprintID,
			ResolutionSprintID: 0,
		})
		return err
	})
	if err != nil {
		return Defect{}, err
	}
	return saved, nil
}

func isValidSeverity(severity string) bool {
	for _, value := range severityValues {
		if severity == value {
			return true
		}
	}
	return false
}
