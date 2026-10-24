package poker

// MemoryRepository guarda las sesiones en memoria (Sprint 3). No es seguro
// para uso concurrente: se accede a través del Manager, que lo protege.
type MemoryRepository struct {
	sessions []Session
	lastID   int
}

// NewMemoryRepository crea un repositorio en memoria vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Save asigna el siguiente ID correlativo global, guarda la sesión y devuelve
// una copia.
func (r *MemoryRepository) Save(session Session) (Session, error) {
	return Session{}, errNotImplemented
}

// OpenByItem devuelve una copia de la sesión ABIERTA del ítem y si existe
// (RN-21.2).
func (r *MemoryRepository) OpenByItem(itemID int) (Session, bool, error) {
	return Session{}, false, errNotImplemented
}
