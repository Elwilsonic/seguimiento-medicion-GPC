package sprint

import "seguimiento-medicion-gpc/internal/backlog"

// AssignItem asigna un ítem del backlog a un sprint (HU-11) y devuelve el
// ítem actualizado como copia, con su propia lista de criterios (RN-11.7).
// La asignación se guarda en el ítem, en su SprintID (RN-11.4), delegando la
// escritura en backlog.Manager.AssignToSprint, el dueño del ítem.
//
// El orden de validación es: sprint (existencia y estado), ítem, mismo
// proyecto, ya asignado y escritura; se informa el primer error encontrado.
// Si falla cualquiera, no se modifica ni el ítem ni el sprint.
func (m *Manager) AssignItem(sprintID, itemID int) (backlog.Item, error) {
	// RN-11.1 y RN-13.9: el sprint tiene que existir y estar ABIERTO. El
	// mutex del sprint se mantiene durante toda la operación, incluida la
	// escritura en backlog, con orden sprint → backlog, para que el cierre no
	// ocurra entre la validación y la asignación.
	m.mu.Lock()
	defer m.mu.Unlock()

	sprint, err := m.repo.GetByID(sprintID)
	if err != nil {
		return backlog.Item{}, err
	}
	if sprint.Status != StatusOpen {
		return backlog.Item{}, ErrSprintClosed
	}

	// RN-11.2: el ítem tiene que pertenecer al proyecto del sprint. El
	// ProjectID de un ítem nunca cambia, pero la lectura también va bajo el
	// mutex del sprint para serializar con Close.
	item, err := m.items.GetItem(itemID)
	if err != nil {
		return backlog.Item{}, err
	}
	if item.ProjectID != sprint.ProjectID {
		return backlog.Item{}, ErrItemFromAnotherProject
	}

	// RN-11.3 y RN-11.4: la verificación de "ya asignado" y la escritura van
	// juntas bajo el mutex de backlog (CA-11.6). El mutex del sprint sigue
	// tomado hasta terminar la escritura.
	return m.items.AssignToSprint(itemID, sprintID)
}
