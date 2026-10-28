package metrics

import (
	"errors"

	"seguimiento-medicion-gpc/internal/defect"
)

// errNotImplemented es el error de las firmas vacías del RED de HU-47: así
// fallan también los casos que esperan cero.
var errNotImplemented = errors.New("HU-47: sin implementar")

// Defects son los defectos detectados (M9) y resueltos (M10) de un sprint o
// de un proyecto (HU-47, TEC-03). Son cantidades enteras, sin "no
// disponible" (RN-47.7 y RN-47.8).
type Defects struct {
	Detected int
	Resolved int
}

// defectLister es la lectura de defectos que usa HU-47: la de HU-31
// (defect.Manager.ListByProject).
type defectLister interface {
	ListByProject(projectID int) ([]defect.Defect, error)
}

// SprintDefects devuelve M9 y M10 de un sprint existente (RN-47.1 y RN-47.2).
func (m *Manager) SprintDefects(sprintID int) (Defects, error) {
	return Defects{}, errNotImplemented
}

// ProjectDefects devuelve M9 y M10 de un proyecto existente (RN-47.3 y
// RN-47.4).
func (m *Manager) ProjectDefects(projectID int) (Defects, error) {
	return Defects{}, errNotImplemented
}
