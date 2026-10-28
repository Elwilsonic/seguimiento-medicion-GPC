package dashboard

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Summary es el resumen del estado de un proyecto (HU-36). No tiene slices,
// mapas ni punteros, así que una copia no comparte estado con lo guardado
// (RN-36.7).
type Summary struct {
	ProjectID     int
	ProjectName   string
	ProjectStatus project.Status
	Pending       int
	InProgress    int
	Finished      int
	// CurrentSprint es el sprint ABIERTO del proyecto; solo vale si
	// HasCurrentSprint es true (RN-36.3).
	CurrentSprint    sprint.Sprint
	HasCurrentSprint bool
}

// ProjectSummary devuelve el resumen de un proyecto existente (HU-36). Lee el
// proyecto con Get, que valida que exista (project.ErrProjectNotFound), los
// ítems con ListItems y los sprints con ListByProject. Si una lectura falla,
// devuelve el error y no un resumen parcial.
func (m *Manager) ProjectSummary(projectID int) (Summary, error) {
	p, err := m.projects.Get(projectID)
	if err != nil {
		return Summary{}, err
	}
	items, err := m.items.ListItems(projectID)
	if err != nil {
		return Summary{}, err
	}
	sprints, err := m.sprints.ListByProject(projectID)
	if err != nil {
		return Summary{}, err
	}

	summary := Summary{ProjectID: p.ID, ProjectName: p.Name, ProjectStatus: p.Status}
	// RN-36.2: todos los ítems del backlog, asignados o no a un sprint.
	for _, item := range items {
		switch item.Status {
		case backlog.StatusPending:
			summary.Pending++
		case backlog.StatusInProgress:
			summary.InProgress++
		case backlog.StatusFinished:
			summary.Finished++
		}
	}
	// RN-36.3: el sprint actual es el ABIERTO; por RN-10.5 hay como máximo uno.
	for _, sp := range sprints {
		if sp.Status == sprint.StatusOpen {
			summary.CurrentSprint = sp
			summary.HasCurrentSprint = true
		}
	}
	return summary, nil
}
