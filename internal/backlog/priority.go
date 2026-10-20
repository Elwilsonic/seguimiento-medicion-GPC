package backlog

import "slices"

// ListItemsByPriority devuelve los ítems del backlog de un proyecto existente
// ordenados por prioridad: ALTA, MEDIA, BAJA y al final las desconocidas
// (HU-09). Es una opción del listado de HU-06: ListItems no cambia (RN-09.5).
//
// Reutiliza ListItems, que valida el proyecto (RN-09.6), toma el mutex del
// Manager y devuelve una lista propia ordenada por ID. Esa lista se ordena
// fuera del mutex con un ordenamiento estable, que conserva el orden por ID a
// igual prioridad (RN-09.2).
func (m *Manager) ListItemsByPriority(projectID int) ([]Item, error) {
	items, err := m.ListItems(projectID)
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(items, func(a, b Item) int {
		return priorityRank(a.Priority) - priorityRank(b.Priority)
	})
	return items, nil
}

// priorityRank da el lugar de cada prioridad en el orden (RN-09.1). Una
// prioridad desconocida va después de BAJA y no se normaliza (RN-09.3).
func priorityRank(priority Priority) int {
	switch priority {
	case PriorityHigh:
		return 0
	case PriorityMedium:
		return 1
	case PriorityLow:
		return 2
	default:
		return 3
	}
}
