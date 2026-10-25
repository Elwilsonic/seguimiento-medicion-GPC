package metrics

// SprintStoryPoints devuelve M1 y M2 de un sprint existente (HU-32).
func (m *Manager) SprintStoryPoints(sprintID int) (SprintPoints, error) {
	return SprintPoints{}, errNotImplemented
}

// ProjectStoryPoints devuelve M1 y M2 de todo el backlog de un proyecto
// existente y la velocidad del equipo (HU-32).
func (m *Manager) ProjectStoryPoints(projectID int) (ProjectPoints, error) {
	return ProjectPoints{}, errNotImplemented
}
