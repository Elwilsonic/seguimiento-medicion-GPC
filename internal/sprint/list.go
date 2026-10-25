package sprint

// Get devuelve una copia del sprint con ese ID o ErrSprintNotFound, incluidos
// los IDs 0 y negativos (HU-32). Lee bajo el mutex del Manager.
func (m *Manager) Get(id int) (Sprint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.GetByID(id)
}

// ListByProject devuelve una lista nueva con copias de los sprints de un
// proyecto, abiertos y cerrados, ordenados por ID ascendente (HU-32). Un
// proyecto sin sprints devuelve una lista vacía, no nil. No valida que el
// proyecto exista: internal/metrics lo valida antes con el backlog.
func (m *Manager) ListByProject(projectID int) ([]Sprint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.ListByProject(projectID)
}
