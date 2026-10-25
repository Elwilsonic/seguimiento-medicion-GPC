package sprint

// Repository es el almacenamiento de los sprints. La lógica de negocio accede
// a los datos solo a través de esta interfaz.
type Repository interface {
	// Save guarda un sprint nuevo, le asigna el ID y devuelve una copia.
	Save(sprint Sprint) (Sprint, error)

	// OpenByProject devuelve una copia del sprint ABIERTO del proyecto y si
	// existe (RN-10.5). HU-11 a HU-14 agregan las operaciones que necesiten.
	OpenByProject(projectID int) (Sprint, bool, error)

	// GetByID devuelve el sprint con ese ID o ErrSprintNotFound si no existe
	// (RN-11.1).
	GetByID(id int) (Sprint, error)

	// Update reemplaza el sprint con ese ID, lo devuelve como copia o da
	// ErrSprintNotFound si no existe (RN-13.4).
	Update(sprint Sprint) (Sprint, error)

	// ListByProject devuelve una lista nueva, ordenada por ID ascendente, con
	// copias de los sprints de un proyecto. Un proyecto sin sprints devuelve
	// una lista vacía, no nil (HU-32).
	ListByProject(projectID int) ([]Sprint, error)
}
