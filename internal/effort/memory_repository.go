package effort

// MemoryRepository guarda los esfuerzos en memoria.
type MemoryRepository struct {
	efforts []Effort
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

func (r *MemoryRepository) Save(Effort) (Effort, error) {
	return Effort{}, errNotImplemented
}
