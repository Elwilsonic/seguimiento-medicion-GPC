package sprint

import "seguimiento-medicion-gpc/internal/backlog"

// CompleteItem marca como TERMINADO un ítem asignado a un sprint ABIERTO
// (HU-12) y devuelve una copia actualizada. El estado se cambia con HU-08, así
// que se respetan sus transiciones: EN_CURSO → TERMINADO es válido, un ítem ya
// TERMINADO es un no-op y PENDIENTE → TERMINADO devuelve ErrInvalidTransition.
//
// El orden de validación es: sprint (existencia), estado del sprint, ítem
// (existencia), pertenencia al sprint y cambio de estado (RN-12.1 a RN-12.5).
func (m *Manager) CompleteItem(sprintID, itemID int) (backlog.Item, error) {
	// RN-12.4 y RN-12.5: el sprint tiene que existir y estar ABIERTO. La
	// consulta va bajo el mutex del sprint y se suelta antes de llamar al
	// Manager de backlog: nunca se toman dos mutex a la vez.
	m.mu.Lock()
	sprint, err := m.repo.GetByID(sprintID)
	m.mu.Unlock()
	if err != nil {
		return backlog.Item{}, err
	}
	if sprint.Status != StatusOpen {
		return backlog.Item{}, ErrSprintClosed
	}

	// RN-12.3 y RN-12.5: el ítem tiene que existir y pertenecer a este sprint.
	item, err := m.items.GetItem(itemID)
	if err != nil {
		return backlog.Item{}, err
	}
	if item.SprintID != sprintID {
		return backlog.Item{}, ErrItemNotAssignedToSprint
	}

	// RN-12.2: ChangeStatus conserva el no-op de TERMINADO sin escribir y
	// valida la transición. Además serializa las marcaciones simultáneas.
	return m.items.ChangeStatus(itemID, string(backlog.StatusFinished))
}

// ListCompletedItems devuelve una lista nueva con copias de los ítems del
// sprint cuyo estado es TERMINADO, ordenados por ID ascendente (RN-12.6). La
// consulta se permite tanto en un sprint ABIERTO como CERRADO.
func (m *Manager) ListCompletedItems(sprintID int) ([]backlog.Item, error) {
	// RN-12.5: el sprint tiene que existir. Se suelta el mutex del sprint antes
	// de llamar al Manager de backlog.
	m.mu.Lock()
	sprint, err := m.repo.GetByID(sprintID)
	m.mu.Unlock()
	if err != nil {
		return nil, err
	}

	// RN-12.6 y RN-12.9: se reutiliza HU-06, que ya valida el proyecto y
	// devuelve copias ordenadas por ID. El filtro no cambia el orden.
	items, err := m.items.ListItems(sprint.ProjectID)
	if err != nil {
		return nil, err
	}

	completed := make([]backlog.Item, 0, len(items))
	for _, item := range items {
		if item.SprintID == sprintID && item.Status == backlog.StatusFinished {
			completed = append(completed, item)
		}
	}
	return completed, nil
}
