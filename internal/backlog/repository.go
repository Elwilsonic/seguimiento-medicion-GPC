package backlog

// Repository es el almacenamiento de los ítems del Product Backlog. La
// lógica de negocio accede a los ítems solo a través de esta interfaz.
// HU-05 agrega GetByID y HU-06 agrega ListByProject.
type Repository interface {
	// Save guarda un ítem nuevo, le asigna el ID y devuelve una copia.
	Save(item Item) (Item, error)

	// ListByProject devuelve una lista nueva, ordenada por ID ascendente,
	// con copias de los ítems de un proyecto (HU-06). Un proyecto sin ítems
	// devuelve una lista vacía, no nil.
	ListByProject(projectID int) ([]Item, error)
}
