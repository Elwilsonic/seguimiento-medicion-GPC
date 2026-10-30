package member

// GetByID devuelve una copia del integrante con ese ID (HU-16).
func (r *MemoryRepository) GetByID(id int) (Member, error) {
	return Member{}, errNotImplemented
}

// Update reemplaza el nombre y el rol del integrante (HU-16).
func (r *MemoryRepository) Update(m Member) (Member, error) {
	return Member{}, errNotImplemented
}
