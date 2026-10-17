package sprint

import (
	"strings"
	"time"
)

// Create crea un sprint en un proyecto existente (HU-10) y devuelve una
// copia. El sprint queda en estado ABIERTO (RN-10.7).
//
// El orden de validación es: nombre, sprint goal, fecha de inicio, fecha de
// fin, relación entre fechas, proyecto, sprint abierto y guardado; se
// informa el primer error encontrado.
func (m *Manager) Create(projectID int, name string, startDate, endDate time.Time, goal string) (Sprint, error) {
	// RN-10.1: el nombre se guarda sin espacios en blanco Unicode en los
	// extremos (strings.TrimSpace) y no puede quedar vacío.
	name = strings.TrimSpace(name)
	if name == "" {
		return Sprint{}, ErrEmptyName
	}

	// RN-10.2: el sprint goal sigue la misma regla que el nombre.
	goal = strings.TrimSpace(goal)
	if goal == "" {
		return Sprint{}, ErrEmptyGoal
	}

	// RN-10.3: las dos fechas son obligatorias (el valor zero es una fecha
	// ausente) y la de fin es estrictamente posterior a la de inicio.
	if startDate.IsZero() {
		return Sprint{}, ErrMissingStartDate
	}
	if endDate.IsZero() {
		return Sprint{}, ErrMissingEndDate
	}
	if !endDate.After(startDate) {
		return Sprint{}, ErrEndNotAfterStart
	}

	// RN-10.4: el proyecto tiene que existir. Se reutiliza la búsqueda por ID
	// de HU-03 y su error se propaga tal cual. La validación se hace antes de
	// tomar el mutex, para no tomar dos mutex a la vez.
	if _, err := m.projects.Get(projectID); err != nil {
		return Sprint{}, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// RN-10.5: no puede haber más de un sprint ABIERTO por proyecto. La
	// consulta y el guardado van juntos bajo el mutex, para que dos
	// creaciones simultáneas no dejen dos sprints abiertos (CA-10.6).
	_, exists, err := m.repo.OpenByProject(projectID)
	if err != nil {
		return Sprint{}, err
	}
	if exists {
		return Sprint{}, ErrOpenSprintExists
	}

	sprint := Sprint{
		ProjectID: projectID,
		Name:      name,
		StartDate: startDate,
		EndDate:   endDate,
		Goal:      goal,
		Status:    StatusOpen,
	}
	return m.repo.Save(sprint)
}
