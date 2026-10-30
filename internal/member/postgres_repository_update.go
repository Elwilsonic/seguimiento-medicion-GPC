package member

// GetByID devuelve el integrante con ese ID (HU-16).
func (r *PostgresRepository) GetByID(id int) (Member, error) {
	return Member{}, errNotImplemented
}

// Update reemplaza el nombre y el rol del integrante (HU-16).
func (r *PostgresRepository) Update(m Member) (Member, error) {
	return Member{}, errNotImplemented
}
