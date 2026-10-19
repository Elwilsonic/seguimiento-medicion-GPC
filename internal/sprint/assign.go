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
	// RN-11.1: el sprint tiene que existir y estar ABIERTO. La consulta va
	// bajo el mutex y se suelta antes de llamar al Manager de backlog: nunca
	// se toman dos mutex a la vez.
	m.mu.Lock()
	sprint, err := m.repo.GetByID(sprintID)
	m.mu.Unlock()
	if err != nil {
		return backlog.Item{}, err
	}
	if sprint.Status != StatusOpen {
		return backlog.Item{}, ErrSprintClosed
	}

	// RN-11.2: el ítem tiene que pertenecer al proyecto del sprint. Es seguro
	// comparar afuera del mutex de backlog porque el ProjectID de un ítem
	// nunca cambia.
	item, err := m.items.GetItem(itemID)
	if err != nil {
		return backlog.Item{}, err
	}
	if item.ProjectID != sprint.ProjectID {
		return backlog.Item{}, ErrItemFromAnotherProject
	}

	// RN-11.3 y RN-11.4: la verificación de "ya asignado" y la escritura van
	// juntas bajo el mutex de backlog (CA-11.6).
	return m.items.AssignToSprint(itemID, sprintID)
}
