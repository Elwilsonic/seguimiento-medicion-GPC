package effort

import (
	"math"
	"strings"
	"time"

	"seguimiento-medicion-gpc/internal/sprint"
)

// Register registra un esfuerzo sobre una historia asignada a un sprint
// ABIERTO (HU-26). Valida los datos en el orden de la SDD y delega en
// WithOpenSprint la existencia, el estado y el mutex del sprint, para que el
// guardado sea atómico con el cierre (RN-26.11).
func (m *Manager) Register(sprintID, itemID int, member string, date time.Time, activity string, hours float64) (Effort, error) {
	member = strings.TrimSpace(member)
	activity = strings.TrimSpace(activity)

	if member == "" {
		return Effort{}, ErrEmptyMember
	}
	if activity == "" {
		return Effort{}, ErrEmptyActivity
	}
	if date.IsZero() {
		return Effort{}, ErrMissingDate
	}
	if hours <= 0 || math.IsNaN(hours) || math.IsInf(hours, 0) {
		return Effort{}, ErrInvalidHours
	}

	var saved Effort
	err := m.sprints.WithOpenSprint(sprintID, func(sprint.Sprint) error {
		item, err := m.items.GetItem(itemID)
		if err != nil {
			return err
		}
		if item.SprintID != sprintID {
			return sprint.ErrItemNotAssignedToSprint
		}
		saved, err = m.repo.Save(Effort{
			Member:   member,
			Date:     date,
			Activity: activity,
			ItemID:   itemID,
			SprintID: sprintID,
			Hours:    hours,
		})
		return err
	})
	if err != nil {
		return Effort{}, err
	}
	return saved, nil
}
