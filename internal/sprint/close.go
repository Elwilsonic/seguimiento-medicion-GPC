package sprint

// Close cierra un sprint ABIERTO (HU-13) y devuelve una copia actualizada.
// Solo cambia el estado a CERRADO; no modifica el resto de los datos.
//
// El mutex del Manager se mantiene durante toda la operación, incluida la
// escritura en el repositorio, para que no se pueda asignar ni completar un
// ítem después del cierre (RN-13.9). El orden es siempre sprint → backlog.
func (m *Manager) Close(sprintID int) (Sprint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// RN-13.1: el sprint tiene que existir.
	sprint, err := m.repo.GetByID(sprintID)
	if err != nil {
		return Sprint{}, err
	}
	// RN-13.2 y RN-13.3: solo se cierra un sprint ABIERTO; un sprint ya
	// CERRADO devuelve ErrSprintClosed.
	if sprint.Status != StatusOpen {
		return Sprint{}, ErrSprintClosed
	}

	// RN-13.4: solo cambia el estado. RN-13.8: si Update falla, se propaga
	// el error y el sprint no queda cerrado.
	sprint.Status = StatusClosed
	return m.repo.Update(sprint)
}
