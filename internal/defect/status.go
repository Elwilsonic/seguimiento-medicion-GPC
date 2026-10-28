package defect

// ChangeStatus cambia el estado de un defecto y registra el sprint de
// resolución (HU-30). Firma vacía del RED: se implementa en el GREEN.
func (m *Manager) ChangeStatus(id int, status string, resolutionSprintID int) (Defect, error) {
	return Defect{}, nil
}
