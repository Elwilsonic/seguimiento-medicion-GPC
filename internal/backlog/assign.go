package backlog

import "errors"

// ErrItemAlreadyAssigned indica que el ítem ya está asignado a un sprint,
// aunque sea el mismo al que se quiere asignar (RN-11.3).
var ErrItemAlreadyAssigned = errors.New("el ítem ya está asignado a un sprint")

// AssignToSprint escribe la asignación del ítem al sprint en el propio ítem
// (HU-11, RN-11.4) y devuelve una copia con su propia lista de criterios
// (RN-11.7). No valida el sprint: existencia, estado y proyecto los valida
// sprint.Manager.AssignItem, su único llamador (RN-11.7).
//
// La lectura, la verificación de RN-11.3 y la escritura van juntas bajo el
// mutex, para que de dos asignaciones simultáneas del mismo ítem exactamente
// una gane (CA-11.6). Si falla, no se escribe nada y el ítem queda como
// estaba.
func (m *Manager) AssignToSprint(itemID, sprintID int) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := m.repo.GetByID(itemID)
	if err != nil {
		return Item{}, err
	}

	// RN-11.3: un ítem solo se asigna una vez; no hay reasignación.
	if current.SprintID != 0 {
		return Item{}, ErrItemAlreadyAssigned
	}

	item := current
	item.SprintID = sprintID
	return m.repo.Update(item)
}
