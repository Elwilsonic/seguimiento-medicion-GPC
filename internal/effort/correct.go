package effort

import (
	"math"

	"seguimiento-medicion-gpc/internal/sprint"
)

// CorrectHours corrige las horas de un esfuerzo de un sprint ABIERTO (HU-28)
// y conserva los demás campos (RN-28.1). Valida primero las horas (400),
// después busca el esfuerzo (404) y por último delega en WithOpenSprint el
// estado del sprint (409), para que la corrección sea atómica con el cierre
// (RN-28.5 y RN-28.6).
func (m *Manager) CorrectHours(id int, hours float64) (Effort, error) {
	if !validHours(hours) {
		return Effort{}, ErrInvalidHours
	}

	current, err := m.repo.GetByID(id)
	if err != nil {
		return Effort{}, err
	}

	var updated Effort
	err = m.sprints.WithOpenSprint(current.SprintID, func(sprint.Sprint) error {
		// Con el mutex del sprint tomado se vuelve a leer el esfuerzo, para no
		// pisar otra corrección hecha entre la primera lectura y esta.
		latest, err := m.repo.GetByID(id)
		if err != nil {
			return err
		}
		latest.Hours = hours
		updated, err = m.repo.Update(latest)
		return err
	})
	if err != nil {
		return Effort{}, err
	}
	return updated, nil
}

// validHours indica si las horas son positivas y finitas (RN-26.4).
func validHours(hours float64) bool {
	return hours > 0 && !math.IsNaN(hours) && !math.IsInf(hours, 0)
}
