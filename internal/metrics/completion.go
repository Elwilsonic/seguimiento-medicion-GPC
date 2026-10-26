package metrics

import "seguimiento-medicion-gpc/internal/backlog"

// Completion es el porcentaje de historias completadas de un sprint o de un
// proyecto (HU-34, M8 de TEC-03): la cantidad de ítems, la de ítems
// TERMINADO y el porcentaje sin redondear (RN-34.5). Sin ítems,
// PercentAvailable es false (RN-34.6).
type Completion struct {
	Items            int
	Finished         int
	Percent          float64
	PercentAvailable bool
}

// SprintCompletion devuelve M8 de un sprint existente: los ítems con SprintID
// igual al sprint (RN-34.1). Un sprint inexistente devuelve
// sprint.ErrSprintNotFound.
func (m *Manager) SprintCompletion(sprintID int) (Completion, error) {
	sp, err := m.sprints.Get(sprintID)
	if err != nil {
		return Completion{}, err
	}
	items, err := m.items.ListItems(sp.ProjectID)
	if err != nil {
		return Completion{}, err
	}
	return completion(itemsOfSprint(items, sprintID)), nil
}

// ProjectCompletion devuelve M8 de todo el backlog de un proyecto existente,
// asignado o no a un sprint (RN-34.2). Un proyecto inexistente devuelve
// project.ErrProjectNotFound, que valida ListItems.
func (m *Manager) ProjectCompletion(projectID int) (Completion, error) {
	items, err := m.items.ListItems(projectID)
	if err != nil {
		return Completion{}, err
	}
	return completion(items), nil
}

// completion cuenta los ítems y los TERMINADO, tengan o no Story Points
// (RN-34.3 y RN-34.4), y calcula el porcentaje sin redondear.
func completion(items []backlog.Item) Completion {
	c := Completion{Items: len(items)}
	for _, item := range items {
		if item.Status == backlog.StatusFinished {
			c.Finished++
		}
	}
	if c.Items > 0 {
		c.Percent = float64(c.Finished) * 100 / float64(c.Items)
		c.PercentAvailable = true
	}
	return c
}
