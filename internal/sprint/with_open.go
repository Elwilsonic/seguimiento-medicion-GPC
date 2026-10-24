package sprint

// WithOpenSprint ejecuta fn con el sprint ABIERTO indicado y el mutex del
// Manager tomado (HU-26, RN-26.11). Si el sprint no existe devuelve
// ErrSprintNotFound y, si está CERRADO, ErrSprintClosed. Mantener el mutex
// durante fn serializa la operación con Close y con las demás escrituras del
// sprint, con el orden sprint → backlog/effort.
func (m *Manager) WithOpenSprint(sprintID int, fn func(Sprint) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sprint, err := m.repo.GetByID(sprintID)
	if err != nil {
		return err
	}
	if sprint.Status != StatusOpen {
		return ErrSprintClosed
	}
	return fn(sprint)
}
