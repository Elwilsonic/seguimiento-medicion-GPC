package poker

// Repository es el almacenamiento de las sesiones de Planning Poker. La
// lógica de negocio accede a las sesiones solo a través de esta interfaz.
type Repository interface {
	// Save guarda una sesión nueva, le asigna el ID y devuelve una copia.
	Save(session Session) (Session, error)

	// OpenByItem devuelve una copia de la sesión ABIERTA del ítem y si existe
	// (RN-21.2). HU-22 agrega las operaciones que necesite.
	OpenByItem(itemID int) (Session, bool, error)

	// GetByID devuelve una copia de la sesión identificada por id.
	GetByID(id int) (Session, error)

	// Update reemplaza una sesión existente y devuelve una copia.
	Update(session Session) (Session, error)
}
