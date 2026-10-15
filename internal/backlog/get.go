package backlog

import "errors"

// ErrItemNotFound indica que no existe un ítem del backlog con el ID
// recibido, incluidos los IDs 0 y negativos (RN-05.2).
var ErrItemNotFound = errors.New("no existe un ítem del backlog con ese id")

// GetItem devuelve una copia del ítem del backlog con ese ID (RN-05.1), con
// su propia copia de la lista de criterios de aceptación (RN-05.3). Si ningún
// ítem guardado tiene ese ID, devuelve ErrItemNotFound (RN-05.2). La consulta
// es de solo lectura (RN-05.4).
func (m *Manager) GetItem(id int) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.GetByID(id)
}
