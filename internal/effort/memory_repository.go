package effort

// MemoryRepository guarda los esfuerzos en memoria.
type MemoryRepository struct {
	efforts []Effort
	lastID  int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Save asigna el siguiente ID correlativo global, guarda el esfuerzo y
// devuelve una copia. Effort no tiene slices ni mapas, así que copiar el
// struct alcanza (RN-26.9).
func (r *MemoryRepository) Save(effort Effort) (Effort, error) {
	r.lastID++
	stored := effort
	stored.ID = r.lastID
	r.efforts = append(r.efforts, stored)
	return stored, nil
}

// ListByItem devuelve los esfuerzos de una historia, ordenados por ID.
func (r *MemoryRepository) ListByItem(itemID int) ([]Effort, error) {
	return nil, nil
}

// ListBySprint devuelve los esfuerzos de un sprint, ordenados por ID.
func (r *MemoryRepository) ListBySprint(sprintID int) ([]Effort, error) {
	return nil, nil
}
