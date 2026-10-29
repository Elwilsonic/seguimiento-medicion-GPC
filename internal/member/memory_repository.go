package member

// MemoryRepository guarda los integrantes en memoria. No es seguro para uso
// concurrente: se accede a través del Manager, que lo protege.
type MemoryRepository struct{}

// NewMemoryRepository crea un repositorio en memoria vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Save guarda el integrante.
func (r *MemoryRepository) Save(m Member) (Member, error) {
	return Member{}, errNotImplemented
}

// ListByProject devuelve los integrantes de un proyecto.
func (r *MemoryRepository) ListByProject(projectID int) ([]Member, error) {
	return nil, errNotImplemented
}
