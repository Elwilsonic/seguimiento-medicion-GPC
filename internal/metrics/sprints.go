package metrics

import "seguimiento-medicion-gpc/internal/sprint"

// SprintWithMetrics es un sprint de un proyecto con sus métricas (HU-37): la
// copia del sprint y las mismas métricas que devuelve Sprint (HU-35).
type SprintWithMetrics struct {
	Sprint  sprint.Sprint
	Metrics SprintMetrics
}

// ProjectSprints devuelve cada sprint de un proyecto existente con sus
// métricas, en una lista nueva ordenada por ID (RN-37.1 y RN-37.4). Lee una
// sola vez los ítems (que valida el proyecto) y la lista de sprints, después
// los esfuerzos de cada sprint en orden de ID y los defectos del proyecto,
// como Project (RN-37.2); corta en el primer error, sin lista parcial. Un
// proyecto inexistente devuelve project.ErrProjectNotFound.
func (m *Manager) ProjectSprints(projectID int) ([]SprintWithMetrics, error) {
	items, err := m.items.ListItems(projectID)
	if err != nil {
		return nil, err
	}
	sprints, err := m.sprints.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	actual := make([]float64, len(sprints))
	for i, sp := range sprints {
		summary, err := m.efforts.ListBySprint(sp.ID)
		if err != nil {
			return nil, err
		}
		actual[i] = summary.TotalHours
	}
	defects, err := m.defects.ListByProject(projectID)
	if err != nil {
		return nil, err
	}

	result := make([]SprintWithMetrics, 0, len(sprints))
	for i, sp := range sprints {
		result = append(result, SprintWithMetrics{Sprint: sp, Metrics: sprintMetricsOf(sp.ID, items, actual[i], defects)})
	}
	return result, nil
}
