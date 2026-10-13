package project

// Repository es el almacenamiento de proyectos. La lógica de negocio accede
// a los datos solo a través de esta interfaz.
type Repository interface {
	// Save guarda un proyecto nuevo, le asigna el ID y devuelve una copia.
	Save(p Project) (Project, error)
	// List devuelve copias de todos los proyectos guardados.
	List() ([]Project, error)
	// GetByID devuelve una copia del proyecto con ese ID, o ErrProjectNotFound
	// si no existe (HU-03).
	GetByID(id int) (Project, error)
	// Update reemplaza el proyecto guardado con ese ID y devuelve una copia, o
	// ErrProjectNotFound si no existe (HU-02).
	Update(p Project) (Project, error)
}
