package effort

// Summary es el resultado de consultar los esfuerzos de una historia o de un
// sprint. TotalHours conserva la suma exacta de las horas (HU-27).
type Summary struct {
	Efforts    []Effort
	TotalHours float64
}

// ListByItem devuelve los esfuerzos de una historia con su total exacto.
func (m *Manager) ListByItem(itemID int) (Summary, error) {
	return Summary{}, errNotImplemented
}

// ListBySprint devuelve los esfuerzos de un sprint con su total exacto.
func (m *Manager) ListBySprint(sprintID int) (Summary, error) {
	return Summary{}, errNotImplemented
}
