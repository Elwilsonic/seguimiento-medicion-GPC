package member

// Repository guarda los integrantes (HU-15).
type Repository interface {
	// Save asigna el siguiente ID correlativo global, guarda el integrante y
	// devuelve una copia (RN-15.7).
	Save(Member) (Member, error)

	// ListByProject devuelve una lista nueva con copias de los integrantes de
	// un proyecto, ordenados por ID ascendente. Sin integrantes devuelve una
	// lista vacía, no nil (D5).
	ListByProject(projectID int) ([]Member, error)

	// GetByID devuelve una copia del integrante con ese ID, o
	// ErrMemberNotFound si no existe (HU-16).
	GetByID(id int) (Member, error)

	// Update reemplaza el nombre y el rol del integrante con el ID de m, sin
	// cambiar su proyecto, y devuelve una copia de lo guardado, o
	// ErrMemberNotFound si no existe (HU-16).
	Update(m Member) (Member, error)
}
