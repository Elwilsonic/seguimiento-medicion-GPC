package report

import (
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/metrics"
	"seguimiento-medicion-gpc/internal/project"
)

// ProjectReport es el reporte de un proyecto (HU-38): sus datos, sus
// métricas M1 a M10, sus esfuerzos, sus defectos y los integrantes que
// registraron esfuerzo, cada parte en su propio tipo (RN-38.1).
type ProjectReport struct {
	Project project.Project
	Metrics metrics.ProjectMetrics
	Efforts Efforts
	Defects Defects
	Members Members
}

// Efforts son los esfuerzos de los sprints del proyecto, ordenados por sprint
// y por ID, con la suma exacta de sus horas (RN-38.4).
type Efforts struct {
	Efforts    []effort.Effort
	TotalHours float64
}

// Defects son los defectos del proyecto, en cualquier estado, ordenados por
// ID (RN-38.5).
type Defects struct {
	Defects []defect.Defect
}

// Members son los integrantes que registraron esfuerzo en el proyecto, cada
// uno una sola vez, en orden de primera aparición (RN-38.6).
type Members struct {
	Names []string
}

// Project devuelve el reporte de un proyecto existente (HU-38) en una sola
// consulta (RN-38.1). Lee, en este orden, el proyecto (que valida que exista,
// project.ErrProjectNotFound), las métricas, los sprints del proyecto, los
// esfuerzos de cada sprint en orden de ID y los defectos del proyecto, y
// corta en el primer error, sin reporte parcial (RN-38.3).
func (m *Manager) Project(projectID int) (ProjectReport, error) {
	p, err := m.projects.Get(projectID)
	if err != nil {
		return ProjectReport{}, err
	}
	projectMetrics, err := m.metrics.Project(projectID)
	if err != nil {
		return ProjectReport{}, err
	}
	sprints, err := m.sprints.ListByProject(projectID)
	if err != nil {
		return ProjectReport{}, err
	}
	// RN-38.4: los esfuerzos de cada sprint, en orden de ID de sprint, y su
	// total sumado sprint por sprint, igual que M5 (RN-35.5).
	efforts := Efforts{Efforts: []effort.Effort{}}
	for _, sp := range sprints {
		summary, err := m.efforts.ListBySprint(sp.ID)
		if err != nil {
			return ProjectReport{}, err
		}
		efforts.Efforts = append(efforts.Efforts, summary.Efforts...)
		efforts.TotalHours += summary.TotalHours
	}
	defects, err := m.defects.ListByProject(projectID)
	if err != nil {
		return ProjectReport{}, err
	}

	return ProjectReport{
		Project: p,
		Metrics: projectMetrics,
		Efforts: efforts,
		Defects: Defects{Defects: defects},
		Members: membersOf(efforts.Efforts),
	}, nil
}

// membersOf devuelve los integrantes de los esfuerzos, cada uno una sola vez,
// en el orden de su primer esfuerzo y tal como están guardados (RN-38.6).
func membersOf(efforts []effort.Effort) Members {
	names := []string{}
	seen := map[string]bool{}
	for _, ef := range efforts {
		if !seen[ef.Member] {
			seen[ef.Member] = true
			names = append(names, ef.Member)
		}
	}
	return Members{Names: names}
}
