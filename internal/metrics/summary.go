package metrics

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
)

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

// Sprint devuelve todas las métricas de un sprint existente en una sola
// consulta (RN-35.1). Lee una sola vez el sprint y los ítems de su proyecto,
// y después los esfuerzos del sprint y los defectos del proyecto, en ese
// orden; corta en el primer error (RN-35.3 y RN-35.4). Un sprint inexistente
// devuelve sprint.ErrSprintNotFound.
func (m *Manager) Sprint(sprintID int) (SprintMetrics, error) {
	sp, err := m.sprints.Get(sprintID)
	if err != nil {
		return SprintMetrics{}, err
	}
	items, err := m.items.ListItems(sp.ProjectID)
	if err != nil {
		return SprintMetrics{}, err
	}
	summary, err := m.efforts.ListBySprint(sprintID)
	if err != nil {
		return SprintMetrics{}, err
	}
	defects, err := m.defects.ListByProject(sp.ProjectID)
	if err != nil {
		return SprintMetrics{}, err
	}

	return sprintMetricsOf(sprintID, items, summary.TotalHours, defects), nil
}

// sprintMetricsOf calcula M1, M2 y M4 a M10 de un sprint a partir de los
// ítems y los defectos de su proyecto y de sus horas reales, sin leer nada.
// La usan Sprint (HU-35) y ProjectSprints (HU-37).
func sprintMetricsOf(sprintID int, items []backlog.Item, actual float64, defects []defect.Defect) SprintMetrics {
	of := itemsOfSprint(items, sprintID)
	planned, completed, unestimated := storyPoints(of)
	return SprintMetrics{
		SprintID:   sprintID,
		Points:     SprintPoints{SprintID: sprintID, Planned: planned, Completed: completed, Unestimated: unestimated},
		Completion: completion(of),
		Hours:      hoursOf(of, actual),
		Defects:    sprintDefects(defects, sprintID),
	}
}

// Project devuelve todas las métricas de un proyecto existente en una sola
// consulta (RN-35.1). Lee una sola vez los ítems (que valida el proyecto) y
// la lista de sprints, y después los esfuerzos de cada sprint en orden de ID
// y los defectos del proyecto; corta en el primer error (RN-35.3 y RN-35.4).
// M5 se suma sprint por sprint, como HU-33 (RN-35.5). Un proyecto inexistente
// devuelve project.ErrProjectNotFound.
func (m *Manager) Project(projectID int) (ProjectMetrics, error) {
	items, err := m.items.ListItems(projectID)
	if err != nil {
		return ProjectMetrics{}, err
	}
	sprints, err := m.sprints.ListByProject(projectID)
	if err != nil {
		return ProjectMetrics{}, err
	}
	actual := 0.0
	for _, sp := range sprints {
		summary, err := m.efforts.ListBySprint(sp.ID)
		if err != nil {
			return ProjectMetrics{}, err
		}
		actual += summary.TotalHours
	}
	defects, err := m.defects.ListByProject(projectID)
	if err != nil {
		return ProjectMetrics{}, err
	}

	return ProjectMetrics{
		ProjectID:  projectID,
		Points:     projectPoints(projectID, items, sprints),
		Completion: completion(items),
		Hours:      hoursOf(items, actual),
		Defects:    projectDefects(defects),
	}, nil
}
