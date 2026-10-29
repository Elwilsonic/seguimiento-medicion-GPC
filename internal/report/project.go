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

// Project devuelve el reporte de un proyecto existente (HU-38).
func (m *Manager) Project(projectID int) (ProjectReport, error) {
	return ProjectReport{}, nil
}
