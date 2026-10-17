package backlog

import (
	"errors"
	"math"
	"strings"
)

// ErrInvalidEstimatedHours indica que las horas estimadas son negativas, NaN
// o infinitas (RN-07.5).
var ErrInvalidEstimatedHours = errors.New("las horas estimadas deben ser un número mayor o igual a cero")

// Update modifica un ítem existente identificado por su ID (HU-07) y
// devuelve una copia. Reemplaza el título, la descripción, la prioridad, los
// criterios de aceptación y las horas estimadas; conserva el ID, el
// ProjectID, el estado y los story points (los cambian HU-08 y HU-20).
//
// El orden de validación es: título, prioridad, criterios y horas, sin
// consultar el repositorio; después, con el mutex, la existencia del ítem
// (con GetByID, RN-07.6) y la escritura. Si falla cualquiera, no se escribe
// nada y el ítem queda como estaba.
func (m *Manager) Update(id int, title, description, priority string, criteria []string, estimatedHours float64) (Item, error) {
	// RN-07.1: mismo criterio que RN-04.1, el título se guarda recortado y no
	// puede quedar vacío.
	title = strings.TrimSpace(title)
	if title == "" {
		return Item{}, ErrEmptyTitle
	}

	// RN-07.3: mismo criterio que RN-04.3, se normaliza y tiene que ser
	// ALTA, MEDIA o BAJA.
	normalized, err := normalizePriority(priority)
	if err != nil {
		return Item{}, err
	}

	// RN-07.4: mismo criterio que RN-04.4, al menos un criterio no vacío.
	criteria = cleanCriteria(criteria)
	if len(criteria) == 0 {
		return Item{}, ErrNoAcceptanceCriteria
	}

	// RN-07.5: las horas tienen que ser un número >= 0, sin NaN ni infinito;
	// 0 significa sin horas estimadas.
	if err := validateEstimatedHours(estimatedHours); err != nil {
		return Item{}, err
	}

	// RN-07.6: la existencia se verifica adentro del mutex, porque Update es
	// una operación leer-modificar-escribir (CA-07.6).
	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := m.repo.GetByID(id)
	if err != nil {
		return Item{}, err
	}

	item := Item{
		ID:                 current.ID,
		ProjectID:          current.ProjectID,
		Title:              title,
		Description:        description,
		Priority:           normalized,
		AcceptanceCriteria: criteria,
		Status:             current.Status,
		StoryPoints:        current.StoryPoints,
		EstimatedHours:     estimatedHours,
	}
	return m.repo.Update(item)
}

// validateEstimatedHours rechaza los valores negativos, NaN e infinitos
// (RN-07.5).
func validateEstimatedHours(hours float64) error {
	if hours < 0 || math.IsNaN(hours) || math.IsInf(hours, 0) {
		return ErrInvalidEstimatedHours
	}
	return nil
}
