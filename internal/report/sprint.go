package report

import (
	"seguimiento-medicion-gpc/internal/metrics"
	"seguimiento-medicion-gpc/internal/sprint"
)

// SprintReport es el reporte de un sprint (HU-39): sus datos, sus métricas
// M1, M2 y M4 a M10, sus esfuerzos, sus defectos y los integrantes que
// registraron esfuerzo, con las mismas partes que el reporte del proyecto
// (RN-39.1).
type SprintReport struct {
	Sprint  sprint.Sprint
	Metrics metrics.SprintMetrics
	Efforts Efforts
	Defects Defects
	Members Members
}

// Sprint devuelve el reporte de un sprint existente, abierto o cerrado
// (HU-39). Firma vacía del RED: se implementa en el GREEN.
func (m *Manager) Sprint(sprintID int) (SprintReport, error) {
	return SprintReport{}, nil
}
