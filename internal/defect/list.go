package defect

// ListByProject devuelve una lista nueva con copias de los defectos de un
// proyecto, en cualquier estado, ordenados por ID (RN-31.2). Es la lectura
// que usa HU-47. Un defecto pertenece al proyecto de su historia: las
// historias se obtienen con ListItems, que además valida el proyecto
// (project.ErrProjectNotFound, RN-31.5). El orden de locks es defect →
// project → backlog (RN-31.12).
func (m *Manager) ListByProject(projectID int) ([]Defect, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	items, err := m.items.ListItems(projectID)
	if err != nil {
		return nil, err
	}
	itemIDs := make(map[int]bool, len(items))
	for _, item := range items {
		itemIDs[item.ID] = true
	}
	return m.selectDefects(func(d Defect) bool { return itemIDs[d.ItemID] })
}

// ListBySprint devuelve una lista nueva con copias de los defectos detectados
// en un sprint, en cualquier estado, ordenados por ID (RN-31.3). Valida el
// sprint con Get (sprint.ErrSprintNotFound, RN-31.5); se puede consultar un
// sprint abierto o cerrado (RN-31.9). El orden de locks es defect → sprint.
func (m *Manager) ListBySprint(sprintID int) ([]Defect, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := m.sprints.Get(sprintID); err != nil {
		return nil, err
	}
	return m.selectDefects(func(d Defect) bool { return d.DetectionSprintID == sprintID })
}

// selectDefects lee todos los defectos y se queda con los que cumplen keep,
// en el orden del repositorio. Sin coincidencias devuelve una lista vacía, no
// nil (RN-31.6).
func (m *Manager) selectDefects(keep func(Defect) bool) ([]Defect, error) {
	defects, err := m.repo.List()
	if err != nil {
		return nil, err
	}
	selected := make([]Defect, 0, len(defects))
	for _, defect := range defects {
		if keep(defect) {
			selected = append(selected, defect)
		}
	}
	return selected, nil
}

// ParseStatus valida y normaliza un estado de defecto, sin distinguir
// mayúsculas ni contar los espacios de los extremos (RN-31.4). Un valor vacío
// también es ErrInvalidStatus: decidir que un filtro vacío no filtra es del
// handler (D3).
func ParseStatus(status string) (Status, error) {
	return normalizeStatus(status)
}

// FilterByStatus devuelve una lista nueva con los defectos en ese estado, en
// el mismo orden; nunca nil y sin modificar la lista recibida (RN-31.4).
func FilterByStatus(defects []Defect, status Status) []Defect {
	filtered := make([]Defect, 0, len(defects))
	for _, defect := range defects {
		if defect.Status == status {
			filtered = append(filtered, defect)
		}
	}
	return filtered
}
