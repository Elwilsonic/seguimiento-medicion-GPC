package effort

// Summary es el resultado de consultar los esfuerzos de una historia o de un
// sprint. TotalHours conserva la suma exacta de las horas (HU-27).
type Summary struct {
	Efforts    []Effort
	TotalHours float64
}

// ListByItem devuelve los esfuerzos de una historia con su total exacto.
func (m *Manager) ListByItem(itemID int) (Summary, error) {
	if _, err := m.items.GetItem(itemID); err != nil {
		return Summary{}, err
	}
	efforts, err := m.repo.ListByItem(itemID)
	if err != nil {
		return Summary{}, err
	}
	return summarize(efforts), nil
}

// ListBySprint devuelve los esfuerzos de un sprint con su total exacto.
func (m *Manager) ListBySprint(sprintID int) (Summary, error) {
	if _, err := m.sprints.Get(sprintID); err != nil {
		return Summary{}, err
	}
	efforts, err := m.repo.ListBySprint(sprintID)
	if err != nil {
		return Summary{}, err
	}
	return summarize(efforts), nil
}

// summarize copia los esfuerzos y calcula su total exacto, sin redondear. La
// copia evita que modificar el resultado altere lo guardado (RN-27.9).
func summarize(efforts []Effort) Summary {
	copied := make([]Effort, len(efforts))
	copy(copied, efforts)

	total := 0.0
	for _, effort := range copied {
		total += effort.Hours
	}
	return Summary{Efforts: copied, TotalHours: total}
}
