package member

// MemoryRepository guarda los integrantes en memoria. No es seguro para uso
// concurrente: se accede a través del Manager, que lo protege (RN-15.9).
type MemoryRepository struct {
	members []Member
	lastID  int
}

// NewMemoryRepository crea un repositorio en memoria vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Save asigna el siguiente ID correlativo global (RN-15.7), guarda el
// integrante y devuelve una copia. Member no tiene slices ni mapas, así que
// copiar el struct alcanza (RN-15.10).
func (r *MemoryRepository) Save(m Member) (Member, error) {
	r.lastID++
	stored := m
	stored.ID = r.lastID
	r.members = append(r.members, stored)
	return stored, nil
}

// ListByProject devuelve una lista nueva con copias de los integrantes del
// proyecto, en orden de ID. Sin integrantes devuelve una lista vacía, no nil.
func (r *MemoryRepository) ListByProject(projectID int) ([]Member, error) {
	list := []Member{}
	for _, m := range r.members {
		if m.ProjectID == projectID {
			list = append(list, m)
		}
	}
	return list, nil
}
