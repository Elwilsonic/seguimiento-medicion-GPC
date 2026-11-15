package effort

import "uuid"

// Repository guarda los esfuerzos.
type Repository interface {
	// GetByPublicID devuelve una copia del esfuerzo con ese identificador
	// público, o ErrEffortNotFound si no existe (HU-52).
	GetByPublicID(id uuid.UUID) (Effort, error)
	Save(Effort) (Effort, error)

	// ListByItem devuelve una lista nueva con copias de los esfuerzos de una
	// historia, ordenados por ID ascendente. Sin registros devuelve una lista
	// vacía, no nil (HU-27).
	ListByItem(itemID int) ([]Effort, error)

	// ListBySprint devuelve una lista nueva con copias de los esfuerzos de un
	// sprint, ordenados por ID ascendente. Sin registros devuelve una lista
	// vacía, no nil (HU-27).
	ListBySprint(sprintID int) ([]Effort, error)

	// GetByID devuelve una copia del esfuerzo con ese ID o ErrEffortNotFound
	// si no existe, incluidos los IDs 0 y negativos (HU-28).
	GetByID(id int) (Effort, error)

	// Update reemplaza el esfuerzo guardado cuyo ID coincida y devuelve una
	// copia, o ErrEffortNotFound si no existe, sin crear ni modificar nada
	// (HU-28).
	Update(Effort) (Effort, error)
}
