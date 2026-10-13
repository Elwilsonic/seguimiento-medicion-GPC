package backlog

// ListItems devuelve una lista nueva con copias de todos los ítems del
// backlog de un proyecto existente, ordenados por ID ascendente (HU-06).
//
// La validación del proyecto se hace antes de tomar el mutex del Manager,
// igual que en Create (HU-04), para no tomar dos mutex a la vez.
func (m *Manager) ListItems(projectID int) ([]Item, error) {
	// RN-06.4: el proyecto tiene que existir. Se reutiliza la búsqueda por ID
	// de HU-03 y su error se propaga tal cual.
	if _, err := m.projects.Get(projectID); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.ListByProject(projectID)
}
