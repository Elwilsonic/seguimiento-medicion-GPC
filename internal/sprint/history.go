package sprint

// History devuelve una lista nueva con copias de los sprints de un proyecto
// existente, abiertos y cerrados, ordenados por ID ascendente (HU-14,
// RN-14.1 y RN-14.3). Un proyecto sin sprints devuelve una lista vacía, no
// nil (RN-14.4); ante cualquier error devuelve nil y ese error (RN-14.7).
func (m *Manager) History(projectID int) ([]Sprint, error) {
	// RN-14.2 y RN-14.6: el proyecto tiene que existir. Se valida con la
	// búsqueda de HU-03 antes de tomar el mutex, como en Create (RN-10.4).
	if _, err := m.projects.Get(projectID); err != nil {
		return nil, err
	}

	sprints, err := m.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	return sprints, nil
}
