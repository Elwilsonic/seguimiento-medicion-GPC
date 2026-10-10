package project

// Repository es el almacenamiento de proyectos. La lógica de negocio accede
// a los datos solo a través de esta interfaz.
type Repository interface {
	// Save guarda un proyecto nuevo, le asigna el ID y devuelve una copia.
	Save(p Project) (Project, error)
	// List devuelve copias de todos los proyectos guardados.
	List() ([]Project, error)
}
