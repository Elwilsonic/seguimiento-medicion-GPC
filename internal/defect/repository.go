package defect

// Repository es el almacenamiento de los defectos. La lógica de negocio
// accede a los datos solo a través de esta interfaz.
type Repository interface {
	// Save guarda un defecto nuevo, le asigna el ID y devuelve una copia.
	Save(Defect) (Defect, error)

	// GetByID devuelve una copia del defecto con ese ID o ErrDefectNotFound
	// si no existe, incluidos los IDs 0 y negativos (HU-30).
	GetByID(id int) (Defect, error)

	// Update reemplaza el defecto guardado cuyo ID coincida y devuelve una
	// copia, o ErrDefectNotFound si no existe (HU-30).
	Update(Defect) (Defect, error)

	// List devuelve una lista nueva con copias de todos los defectos,
	// ordenados por ID ascendente, o una lista vacía no nil si no hay
	// ninguno (HU-31).
	List() ([]Defect, error)
}
