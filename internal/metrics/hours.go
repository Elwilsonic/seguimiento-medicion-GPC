package metrics

// Hours contiene M4 a M7 y la cantidad de ítems sin horas estimadas.
type Hours struct {
	Estimated          float64
	Actual             float64
	Deviation          float64
	DeviationPercent   float64
	Unestimated        int
	DeviationAvailable bool
}

// SprintHours es el stub del cálculo de M4 a M7 de un sprint.
func (m *Manager) SprintHours(sprintID int) (Hours, error) {
	return Hours{}, nil
}

// ProjectHours es el stub del cálculo de M4 a M7 de un proyecto.
func (m *Manager) ProjectHours(projectID int) (Hours, error) {
	return Hours{}, nil
}
