package metrics

import (
	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Manager calcula las métricas leyendo los ítems, los sprints, los esfuerzos
// y los defectos.
// No tiene estado propio, así que no necesita un mutex: cada lectura queda
// protegida por el mutex del Manager que la hace.
type Manager struct {
	items   *backlog.Manager
	sprints *sprint.Manager
	efforts *effort.Manager
	defects defectLister
}

// NewManager crea un Manager que lee los ítems con items, los sprints con
// sprints, los esfuerzos con efforts y los defectos con defects (HU-47).
func NewManager(items *backlog.Manager, sprints *sprint.Manager, efforts *effort.Manager, defects defectLister) *Manager {
	return &Manager{items: items, sprints: sprints, efforts: efforts, defects: defects}
}
