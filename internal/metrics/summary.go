package metrics

// SprintMetrics son todas las métricas de un sprint (HU-35): M1, M2 y M4 a
// M10, formadas por los tipos de HU-32, HU-34, HU-33 y HU-47.
type SprintMetrics struct {
	SprintID   int
	Points     SprintPoints
	Completion Completion
	Hours      Hours
	Defects    Defects
}

// ProjectMetrics son todas las métricas de un proyecto (HU-35): M1 a M10,
// formadas por los tipos de HU-32, HU-34, HU-33 y HU-47.
type ProjectMetrics struct {
	ProjectID  int
	Points     ProjectPoints
	Completion Completion
	Hours      Hours
	Defects    Defects
}

// Sprint devuelve todas las métricas de un sprint en una sola consulta.
func (m *Manager) Sprint(sprintID int) (SprintMetrics, error) {
	return SprintMetrics{}, nil
}

// Project devuelve todas las métricas de un proyecto en una sola consulta.
func (m *Manager) Project(projectID int) (ProjectMetrics, error) {
	return ProjectMetrics{}, nil
}
