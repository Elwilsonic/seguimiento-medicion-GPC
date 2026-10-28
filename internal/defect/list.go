package defect

// ListByProject devuelve los defectos de un proyecto, en cualquier estado,
// ordenados por ID (HU-31). Es la lectura que usa HU-47.
func (m *Manager) ListByProject(projectID int) ([]Defect, error) {
	return nil, nil
}

// ListBySprint devuelve los defectos detectados en un sprint, en cualquier
// estado, ordenados por ID (HU-31).
func (m *Manager) ListBySprint(sprintID int) ([]Defect, error) {
	return nil, nil
}

// ParseStatus valida y normaliza un estado de defecto (HU-31).
func ParseStatus(status string) (Status, error) {
	return "", nil
}

// FilterByStatus devuelve una lista nueva con los defectos en ese estado
// (HU-31).
func FilterByStatus(defects []Defect, status Status) []Defect {
	return nil
}
