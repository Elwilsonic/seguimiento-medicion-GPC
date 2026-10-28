package dashboard

import (
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

// ProjectSummary devuelve el resumen de un proyecto existente (HU-36).
func (m *Manager) ProjectSummary(projectID int) (Summary, error) {
	return Summary{}, nil
}
