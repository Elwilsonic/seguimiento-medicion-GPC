package project

import (
	"errors"
	"time"
)

// ErrMissingStartDate indica que la fecha de inicio está ausente (valor cero
// de time.Time) (RN-18.1, HU-18).
var ErrMissingStartDate = errors.New("la fecha de inicio del proyecto es obligatoria")

// SetStartDate registra la fecha de inicio de un proyecto existente (HU-18) y
// devuelve una copia. Si ya tenía una, la reemplaza (RN-18.4); no cambia
// ningún otro dato del proyecto (RN-18.6).
//
// El orden de validación es: fecha ausente, existencia y guardado; se informa
// el primer error encontrado y, si algo falla, el proyecto queda como estaba.
func (m *Manager) SetStartDate(id int, date time.Time) (Project, error) {
	// RN-18.1: la fecha ausente se informa sin consultar el repositorio.
	if date.IsZero() {
		return Project{}, ErrMissingStartDate
	}

	// RN-18.10: leer, modificar y escribir con el mutex tomado, como Update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// RN-18.3: se reutiliza la búsqueda por ID de HU-03 y su error.
	current, err := m.repo.GetByID(id)
	if err != nil {
		return Project{}, err
	}

	current.StartDate = date
	return m.repo.Update(current)
}
