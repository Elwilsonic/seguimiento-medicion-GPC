package backlog

import "strings"

// Create crea un ítem en el backlog de un proyecto existente (HU-04) y
// devuelve una copia. El ítem queda en estado PENDIENTE y con los story
// points en 0 (sin asignar, RN-04.7 y RN-04.8).
//
// El orden de validación es: título, prioridad, criterios, proyecto y
// guardado; se informa el primer error encontrado (CA-04.2).
func (m *Manager) Create(projectID int, title, description, priority string, criteria []string) (Item, error) {
	// RN-04.1: el título se guarda sin espacios en blanco Unicode en los
	// extremos (strings.TrimSpace) y no puede quedar vacío.
	title = strings.TrimSpace(title)
	if title == "" {
		return Item{}, ErrEmptyTitle
	}

	// RN-04.3: se acepta ALTA, MEDIA o BAJA sin distinguir mayúsculas ni
	// contar los espacios de los extremos; se guarda en mayúsculas.
	normalized, err := normalizePriority(priority)
	if err != nil {
		return Item{}, err
	}

	// RN-04.4: los criterios vacíos se descartan; tiene que quedar al menos uno.
	criteria = cleanCriteria(criteria)
	if len(criteria) == 0 {
		return Item{}, ErrNoAcceptanceCriteria
	}

	// RN-04.5: el proyecto tiene que existir. Se reutiliza la búsqueda por ID
	// de HU-03 y su error se propaga tal cual. La validación se hace antes de
	// tomar el mutex, para no tomar dos mutex a la vez.
	if _, err := m.projects.Get(projectID); err != nil {
		return Item{}, err
	}

	item := Item{
		ProjectID:          projectID,
		Title:              title,
		Description:        description,
		Priority:           normalized,
		AcceptanceCriteria: criteria,
		Status:             StatusPending,
		StoryPoints:        0,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.Save(item)
}

// normalizePriority valida la prioridad contra ALTA, MEDIA y BAJA (RN-04.3).
func normalizePriority(priority string) (Priority, error) {
	normalized := Priority(strings.ToUpper(strings.TrimSpace(priority)))
	switch normalized {
	case PriorityHigh, PriorityMedium, PriorityLow:
		return normalized, nil
	default:
		return "", ErrInvalidPriority
	}
}

// cleanCriteria descarta los criterios vacíos, recorta los espacios de los
// extremos y conserva el orden. Siempre devuelve una lista nueva, para no
// compartir memoria con la recibida (RN-04.4 y RN-04.9).
func cleanCriteria(criteria []string) []string {
	cleaned := make([]string, 0, len(criteria))
	for _, criterion := range criteria {
		criterion = strings.TrimSpace(criterion)
		if criterion != "" {
			cleaned = append(cleaned, criterion)
		}
	}
	return cleaned
}
