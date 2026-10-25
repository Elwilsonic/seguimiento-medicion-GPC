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
	r.lastID++
	session.ID = r.lastID
	session = cloneSession(session)
	r.sessions = append(r.sessions, session)
	return cloneSession(session), nil
}

// OpenByItem devuelve una copia de la sesión ABIERTA del ítem y si existe
// (RN-21.2).
func (r *MemoryRepository) OpenByItem(itemID int) (Session, bool, error) {
	for _, session := range r.sessions {
		if session.ItemID == itemID && session.Status == StatusOpen {
			return cloneSession(session), true, nil
		}
	}
	return Session{}, false, nil
}

// GetByID devuelve una sesión por ID. La implementación se completará en
// GREEN.
func (r *MemoryRepository) GetByID(id int) (Session, error) {
	for _, session := range r.sessions {
		if session.ID == id {
			return cloneSession(session), nil
		}
	}
	return Session{}, ErrSessionNotFound
}

// Update reemplaza una sesión existente. La implementación se completará en
// GREEN.
func (r *MemoryRepository) Update(session Session) (Session, error) {
	for i, stored := range r.sessions {
		if stored.ID == session.ID {
			r.sessions[i] = cloneSession(session)
			return cloneSession(session), nil
		}
	}
	return Session{}, ErrSessionNotFound
}
