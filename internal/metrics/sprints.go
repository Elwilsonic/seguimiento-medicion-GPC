package metrics

import (
	"errors"

	"seguimiento-medicion-gpc/internal/sprint"
)

// SprintWithMetrics es un sprint de un proyecto con sus métricas (HU-37): la
// copia del sprint y las mismas métricas que devuelve Sprint (HU-35).
type SprintWithMetrics struct {
	Sprint  sprint.Sprint
	Metrics SprintMetrics
}

// ProjectSprints devuelve cada sprint de un proyecto con sus métricas.
func (m *Manager) ProjectSprints(projectID int) ([]SprintWithMetrics, error) {
	return nil, errors.New("HU-37: sin implementar")
}
