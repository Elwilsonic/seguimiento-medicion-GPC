package report

import (
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/metrics"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Manager arma los reportes leyendo los proyectos, los sprints, las
// métricas, los esfuerzos y los defectos.
// No tiene estado propio, así que no necesita un mutex: cada lectura queda
// protegida por el mutex del Manager que la hace.
type Manager struct {
	projects *project.Manager
	sprints  *sprint.Manager
	metrics  *metrics.Manager
	efforts  *effort.Manager
	defects  *defect.Manager
}

// NewManager crea un Manager que lee los proyectos con projects, los sprints
// con sprints, las métricas con metrics, los esfuerzos con efforts y los
// defectos con defects.
func NewManager(projects *project.Manager, sprints *sprint.Manager, metrics *metrics.Manager, efforts *effort.Manager, defects *defect.Manager) *Manager {
	return &Manager{projects: projects, sprints: sprints, metrics: metrics, efforts: efforts, defects: defects}
}
