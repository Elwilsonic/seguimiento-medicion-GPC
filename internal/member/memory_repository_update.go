package member

// GetByID devuelve una copia del integrante con ese ID, o ErrMemberNotFound
// si no existe (HU-16).
func (r *MemoryRepository) GetByID(id int) (Member, error) {
	for _, m := range r.members {
		if m.ID == id {
			return m, nil
		}
	}
	return Member{}, ErrMemberNotFound
}

// Update reemplaza el nombre y el rol del integrante con el ID de m, sin
// cambiar su proyecto, y devuelve una copia de lo guardado, o
// ErrMemberNotFound si no existe (HU-16).
func (r *MemoryRepository) Update(m Member) (Member, error) {
	for i := range r.members {
		if r.members[i].ID == m.ID {
			r.members[i].Name = m.Name
			r.members[i].Role = m.Role
			return r.members[i], nil
		}
	}
	return Member{}, ErrMemberNotFound
}
