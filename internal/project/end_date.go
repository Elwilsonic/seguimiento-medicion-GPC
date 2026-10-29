package project

import (
	"errors"
	"time"
)

var (
	// ErrMissingEndDate indica que la fecha de fin está ausente (valor cero
	// de time.Time) (RN-41.1, HU-41).
	ErrMissingEndDate = errors.New("la fecha de fin del proyecto es obligatoria")
	// ErrStartDateNotRegistered indica que el proyecto no tiene fecha de
	// inicio con la que comparar la de fin (RN-41.4).
	ErrStartDateNotRegistered = errors.New("el proyecto no tiene fecha de inicio registrada")
	// ErrEndNotAfterStart indica que la fecha de fin es igual o anterior a la
	// de inicio (RN-41.5).
	ErrEndNotAfterStart = errors.New("la fecha de fin debe ser posterior a la de inicio")
	// ErrStartNotBeforeEnd indica que la fecha de inicio es igual o posterior
	// a la fecha de fin registrada (RN-41.7).
	ErrStartNotBeforeEnd = errors.New("la fecha de inicio debe ser anterior a la de fin")
)

// SetEndDate registra la fecha de fin de un proyecto existente (HU-41) y
// devuelve una copia. Si ya tenía una, la reemplaza (RN-41.6); no cambia
// ningún otro dato del proyecto (RN-41.8).
//
// El orden de validación es: fecha ausente, existencia, fecha de inicio
// registrada, fin posterior al inicio y guardado; se informa el primer error
// encontrado y, si algo falla, el proyecto queda como estaba.
func (m *Manager) SetEndDate(id int, date time.Time) (Project, error) {
	// RN-41.1: la fecha ausente se informa sin consultar el repositorio.
	if date.IsZero() {
		return Project{}, ErrMissingEndDate
	}

	// RN-41.13: leer, validar y escribir con el mutex tomado, como
	// SetStartDate y Update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// RN-41.3: se reutiliza la búsqueda por ID de HU-03 y su error.
	current, err := m.repo.GetByID(id)
	if err != nil {
		return Project{}, err
	}

	// RN-41.4: sin fecha de inicio no hay con qué comparar (D2).
	if current.StartDate.IsZero() {
		return Project{}, ErrStartDateNotRegistered
	}
	// RN-41.5: estrictamente posterior; el mismo día se rechaza (D3).
	if !date.After(current.StartDate) {
		return Project{}, ErrEndNotAfterStart
	}

	current.EndDate = date
	return m.repo.Update(current)
}
