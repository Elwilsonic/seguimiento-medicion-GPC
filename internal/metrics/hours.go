package metrics

import "seguimiento-medicion-gpc/internal/backlog"

// Hours contiene M4 a M7 y la cantidad de ítems sin horas estimadas.
type Hours struct {
	Estimated          float64
	Actual             float64
	Deviation          float64
	DeviationPercent   float64
	Unestimated        int
	DeviationAvailable bool
}

// SprintHours devuelve M4 a M7 de un sprint existente (HU-33). Un sprint
// inexistente devuelve sprint.ErrSprintNotFound.
func (m *Manager) SprintHours(sprintID int) (Hours, error) {
	sp, err := m.sprints.Get(sprintID)
	if err != nil {
		return Hours{}, err
	}
	items, err := m.items.ListItems(sp.ProjectID)
	if err != nil {
		return Hours{}, err
	}
	summary, err := m.efforts.ListBySprint(sprintID)
	if err != nil {
		return Hours{}, err
	}
	return hoursOf(itemsOfSprint(items, sprintID), summary.TotalHours), nil
}

// ProjectHours devuelve M4 a M7 de todo el backlog de un proyecto existente.
// M5 suma los esfuerzos de todos sus sprints, abiertos y cerrados. Un proyecto
// inexistente devuelve project.ErrProjectNotFound, que valida ListItems.
func (m *Manager) ProjectHours(projectID int) (Hours, error) {
	items, err := m.items.ListItems(projectID)
	if err != nil {
		return Hours{}, err
	}
	sprints, err := m.sprints.ListByProject(projectID)
	if err != nil {
		return Hours{}, err
	}

	actual := 0.0
	for _, sp := range sprints {
		summary, err := m.efforts.ListBySprint(sp.ID)
		if err != nil {
			return Hours{}, err
		}
		actual += summary.TotalHours
	}
	return hoursOf(items, actual), nil
}

// hoursOf calcula M4 a M7 sin redondear. La desviación solo está disponible
// cuando hay horas estimadas y horas reales.
func hoursOf(items []backlog.Item, actual float64) Hours {
	h := Hours{Actual: actual}
	for _, item := range items {
		if item.EstimatedHours == 0 {
			h.Unestimated++
			continue
		}
		h.Estimated += item.EstimatedHours
	}
	if h.Estimated > 0 && h.Actual > 0 {
		h.Deviation = h.Actual - h.Estimated
		h.DeviationPercent = h.Deviation / h.Estimated * 100
		h.DeviationAvailable = true
	}
	return h
}
