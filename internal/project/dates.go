package project

import (
	"errors"
	"time"
)

// ErrDatesNotRegistered indica que el proyecto no tiene registradas la fecha
// de inicio y la de fin, así que no hay fechas que modificar (RN-19.4, HU-19).
var ErrDatesNotRegistered = errors.New("el proyecto no tiene registradas las fechas de inicio y de fin")

// SetDates modifica juntas las fechas de inicio y de fin de un proyecto que ya
// las tiene registradas (HU-19) y devuelve una copia. El par nuevo se valida
// entre sí, sin compararlo con las fechas guardadas (RN-19.5), así que se
// puede postergar o adelantar el proyecto en una sola operación; no cambia
// ningún otro dato del proyecto (RN-19.7).
//
// El orden de validación es: inicio ausente, fin ausente, existencia, fechas
// registradas, fin posterior al inicio y guardado; se informa el primer error
// encontrado y, si algo falla, el proyecto queda como estaba (RN-19.6).
func (m *Manager) SetDates(id int, start, end time.Time) (Project, error) {
	// RN-19.1: las fechas ausentes se informan sin consultar el repositorio.
	if start.IsZero() {
		return Project{}, ErrMissingStartDate
	}
	if end.IsZero() {
		return Project{}, ErrMissingEndDate
	}

	// RN-19.11: leer, validar y escribir con el mutex tomado, como
	// SetStartDate, SetEndDate y Update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// RN-19.3: se reutiliza la búsqueda por ID de HU-03 y su error.
	current, err := m.repo.GetByID(id)
	if err != nil {
		return Project{}, err
	}

	// RN-19.4: solo se modifican fechas ya registradas (D1).
	if current.StartDate.IsZero() || current.EndDate.IsZero() {
		return Project{}, ErrDatesNotRegistered
	}
	// RN-19.5: estrictamente posterior; el mismo día se rechaza.
	if !end.After(start) {
		return Project{}, ErrEndNotAfterStart
	}

	// RN-19.6: las dos fechas se guardan con un solo Update.
	current.StartDate = start
	current.EndDate = end
	return m.repo.Update(current)
}
