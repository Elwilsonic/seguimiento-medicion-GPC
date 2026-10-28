package metrics

import "seguimiento-medicion-gpc/internal/defect"

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

// SprintDefects devuelve M9 y M10 de un sprint existente: los defectos
// detectados en el sprint y los RESUELTO con ese sprint de resolución
// (RN-47.1 y RN-47.2). Un sprint inexistente devuelve sprint.ErrSprintNotFound.
func (m *Manager) SprintDefects(sprintID int) (Defects, error) {
	sp, err := m.sprints.Get(sprintID)
	if err != nil {
		return Defects{}, err
	}
	defects, err := m.defects.ListByProject(sp.ProjectID)
	if err != nil {
		return Defects{}, err
	}
	var d Defects
	for _, df := range defects {
		if df.DetectionSprintID == sprintID {
			d.Detected++
		}
		if df.Status == defect.StatusResolved && df.ResolutionSprintID == sprintID {
			d.Resolved++
		}
	}
	return d, nil
}

// ProjectDefects devuelve M9 y M10 de un proyecto existente: todos sus
// defectos y los que están RESUELTO (RN-47.3 y RN-47.4). Un proyecto
// inexistente devuelve project.ErrProjectNotFound, que valida ListByProject.
func (m *Manager) ProjectDefects(projectID int) (Defects, error) {
	defects, err := m.defects.ListByProject(projectID)
	if err != nil {
		return Defects{}, err
	}
	d := Defects{Detected: len(defects)}
	for _, df := range defects {
		if df.Status == defect.StatusResolved {
			d.Resolved++
		}
	}
	return d, nil
}
