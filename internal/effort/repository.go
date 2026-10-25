package effort

// Repository guarda los esfuerzos.
type Repository interface {
	Save(Effort) (Effort, error)

	// ListByItem devuelve una lista nueva con copias de los esfuerzos de una
	// historia, ordenados por ID ascendente. Sin registros devuelve una lista
	// vacía, no nil (HU-27).
	ListByItem(itemID int) ([]Effort, error)

	// ListBySprint devuelve una lista nueva con copias de los esfuerzos de un
	// sprint, ordenados por ID ascendente. Sin registros devuelve una lista
	// vacía, no nil (HU-27).
	ListBySprint(sprintID int) ([]Effort, error)
}
