package sprint

// MemoryRepository guarda los sprints en memoria (Sprint 1). No es seguro
// para uso concurrente: se accede a través del Manager, que lo protege.
type MemoryRepository struct {
	sprints []Sprint
	lastID  int
}

// NewMemoryRepository crea un repositorio en memoria vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Save asigna el siguiente ID correlativo global (RN-10.6), guarda el sprint
// y devuelve una copia. Sprint no tiene slices ni mapas, así que copiar el
// struct alcanza (RN-10.8).
func (r *MemoryRepository) Save(sprint Sprint) (Sprint, error) {
	r.lastID++
	stored := sprint
	stored.ID = r.lastID
	r.sprints = append(r.sprints, stored)
	return stored, nil
}

// OpenByProject devuelve el sprint ABIERTO del proyecto y si existe (RN-10.5).
// La regla es por proyecto: un sprint ABIERTO en otro proyecto no aparece.
func (r *MemoryRepository) OpenByProject(projectID int) (Sprint, bool, error) {
	for _, sprint := range r.sprints {
		if sprint.ProjectID == projectID && sprint.Status == StatusOpen {
			return sprint, true, nil
		}
	}
	return Sprint{}, false, nil
}

// GetByID devuelve el sprint con ese ID (RN-11.1) o ErrSprintNotFound si no
// existe, incluidos los IDs 0 y negativos. Sprint no tiene slices ni mapas,
// así que devolver la copia del rango alcanza (RN-10.8).
func (r *MemoryRepository) GetByID(id int) (Sprint, error) {
	for _, sprint := range r.sprints {
		if sprint.ID == id {
			return sprint, nil
		}
	}
	return Sprint{}, ErrSprintNotFound
}

// Update reemplaza el sprint con ese ID y devuelve una copia (RN-13.4). Si no
// existe, devuelve ErrSprintNotFound y no modifica nada.
func (r *MemoryRepository) Update(sprint Sprint) (Sprint, error) {
	for i := range r.sprints {
		if r.sprints[i].ID == sprint.ID {
			r.sprints[i] = sprint
			return sprint, nil
		}
	}
	return Sprint{}, ErrSprintNotFound
}

// ListByProject devuelve una lista nueva con los sprints del proyecto, en el
// orden en que se guardaron, que es el de ID ascendente (HU-32). Sprint no
// tiene slices ni mapas, así que copiar cada elemento alcanza.
func (r *MemoryRepository) ListByProject(projectID int) ([]Sprint, error) {
	sprints := make([]Sprint, 0)
	for _, sprint := range r.sprints {
		if sprint.ProjectID == projectID {
			sprints = append(sprints, sprint)
		}
	}
	return sprints, nil
}
