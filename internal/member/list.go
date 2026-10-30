package member

// List devuelve una lista nueva con copias de los integrantes de un proyecto
// existente, ordenados por ID ascendente (HU-17, RN-17.1 y RN-17.3). Un
// proyecto sin integrantes devuelve una lista vacía, no nil (RN-17.4); ante
// cualquier error devuelve nil y ese error (RN-17.8).
func (m *Manager) List(projectID int) ([]Member, error) {
	// RN-17.2 y RN-17.7: el proyecto tiene que existir. Se valida con la
	// búsqueda de HU-03 antes de tomar el mutex, como en Register (RN-15.9).
	if _, err := m.projects.Get(projectID); err != nil {
		return nil, err
	}

	// RN-17.7: la lectura va bajo el mismo mutex con el que Register y
	// Update escriben, porque MemoryRepository no es seguro para uso
	// concurrente.
	m.mu.Lock()
	defer m.mu.Unlock()

	members, err := m.repo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	return members, nil
}
