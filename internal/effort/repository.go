package effort

// Repository guarda los esfuerzos.
type Repository interface {
	Save(Effort) (Effort, error)
}
