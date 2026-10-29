package metrics

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/sprint"
)

// SprintStoryPoints devuelve M1 y M2 de un sprint existente (HU-32): los
// ítems con SprintID igual al sprint (RN-32.1 y RN-32.2). Un sprint
// inexistente devuelve sprint.ErrSprintNotFound.
func (m *Manager) SprintStoryPoints(sprintID int) (SprintPoints, error) {
	sp, err := m.sprints.Get(sprintID)
	if err != nil {
		return SprintPoints{}, err
	}
	items, err := m.items.ListItems(sp.ProjectID)
	if err != nil {
		return SprintPoints{}, err
	}
	planned, completed, unestimated := storyPoints(itemsOfSprint(items, sprintID))
	return SprintPoints{SprintID: sprintID, Planned: planned, Completed: completed, Unestimated: unestimated}, nil
}

// ProjectStoryPoints devuelve M1 y M2 de todo el backlog de un proyecto
// existente, asignado o no a un sprint (RN-32.1), y la velocidad con sus
// sprints cerrados (RN-32.5 y RN-32.6). Un proyecto inexistente devuelve
// project.ErrProjectNotFound, que valida ListItems.
func (m *Manager) ProjectStoryPoints(projectID int) (ProjectPoints, error) {
	items, err := m.items.ListItems(projectID)
	if err != nil {
		return ProjectPoints{}, err
	}
	sprints, err := m.sprints.ListByProject(projectID)
	if err != nil {
		return ProjectPoints{}, err
	}
	return projectPoints(projectID, items, sprints), nil
}

// projectPoints calcula M1, M2 y la velocidad de un proyecto a partir de
// sus ítems y sus sprints, sin leer nada (RN-32.1 a RN-32.6). La usan
// ProjectStoryPoints y Project (HU-35).
func projectPoints(projectID int, items []backlog.Item, sprints []sprint.Sprint) ProjectPoints {
	var closed []closedSprint
	for _, sp := range sprints {
		if sp.Status != sprint.StatusClosed {
			continue
		}
		_, completed, _ := storyPoints(itemsOfSprint(items, sp.ID))
		closed = append(closed, closedSprint{ID: sp.ID, Completed: completed})
	}

	planned, completed, unestimated := storyPoints(items)
	v, used := velocity(closed)
	return ProjectPoints{
		ProjectID:         projectID,
		Planned:           planned,
		Completed:         completed,
		Unestimated:       unestimated,
		Velocity:          v,
		VelocityAvailable: used > 0,
		ClosedSprints:     used,
	}
}

// itemsOfSprint devuelve los ítems asignados al sprint.
func itemsOfSprint(items []backlog.Item, sprintID int) []backlog.Item {
	var of []backlog.Item
	for _, item := range items {
		if item.SprintID == sprintID {
			of = append(of, item)
		}
	}
	return of
}
