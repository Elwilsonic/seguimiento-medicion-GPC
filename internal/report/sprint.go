package report

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/metrics"
	"seguimiento-medicion-gpc/internal/sprint"
)

// SprintReport es el reporte de un sprint (HU-39): sus datos, sus métricas
// M1, M2 y M4 a M10, sus esfuerzos, sus defectos, los integrantes que
// registraron esfuerzo y sus historias (HU-49), con las mismas partes que el
// reporte del proyecto (RN-39.1).
type SprintReport struct {
	Sprint  sprint.Sprint
	Metrics metrics.SprintMetrics
	Efforts Efforts
	Defects Defects
	Members Members
	Stories Stories
}

// Sprint devuelve el reporte de un sprint existente, abierto o cerrado
// (HU-39), en una sola consulta (RN-39.1). Lee, en este orden, el sprint (que
// valida que exista, sprint.ErrSprintNotFound), sus métricas, sus esfuerzos,
// los defectos de su proyecto y, al final, las historias de su proyecto
// (RN-49.7), y corta en el primer error, sin reporte parcial (RN-39.3).
func (m *Manager) Sprint(sprintID int) (SprintReport, error) {
	sp, err := m.sprints.Get(sprintID)
	if err != nil {
		return SprintReport{}, err
	}
	sprintMetrics, err := m.metrics.Sprint(sprintID)
	if err != nil {
		return SprintReport{}, err
	}
	// RN-39.4: la misma lectura que M5 del sprint, así el total es el mismo.
	summary, err := m.efforts.ListBySprint(sprintID)
	if err != nil {
		return SprintReport{}, err
	}
	defects, err := m.defects.ListByProject(sp.ProjectID)
	if err != nil {
		return SprintReport{}, err
	}
	// RN-49.2: las historias del proyecto asignadas al sprint, el mismo
	// conjunto que M1 y M8 del sprint (decisión D3 de la SDD).
	items, err := m.items.ListItems(sp.ProjectID)
	if err != nil {
		return SprintReport{}, err
	}

	return SprintReport{
		Sprint:  sp,
		Metrics: sprintMetrics,
		Efforts: Efforts{Efforts: summary.Efforts, TotalHours: summary.TotalHours},
		Defects: Defects{Defects: defectsOfSprint(defects, sprintID)},
		Members: membersOf(summary.Efforts),
		Stories: storiesOf(items, func(it backlog.Item) bool { return it.SprintID == sprintID }),
	}, nil
}

// defectsOfSprint devuelve una lista nueva con los defectos detectados en el
// sprint o RESUELTO con ese sprint de resolución, cada uno una sola vez y en
// el orden recibido (RN-39.5). Es la misma condición que cuentan M9 y M10.
func defectsOfSprint(defects []defect.Defect, sprintID int) []defect.Defect {
	selected := []defect.Defect{}
	for _, d := range defects {
		if d.DetectionSprintID == sprintID || (d.Status == defect.StatusResolved && d.ResolutionSprintID == sprintID) {
			selected = append(selected, d)
		}
	}
	return selected
}
