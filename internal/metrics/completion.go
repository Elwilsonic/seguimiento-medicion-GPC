package metrics

// Completion es el porcentaje de historias completadas de un sprint o de un
// proyecto (HU-34, M8 de TEC-03).
type Completion struct {
	Items            int
	Finished         int
	Percent          float64
	PercentAvailable bool
}

// SprintCompletion devuelve M8 de un sprint. Stub del RED.
func (m *Manager) SprintCompletion(sprintID int) (Completion, error) {
	return Completion{}, nil
}

// ProjectCompletion devuelve M8 de un proyecto. Stub del RED.
func (m *Manager) ProjectCompletion(projectID int) (Completion, error) {
	return Completion{}, nil
}
