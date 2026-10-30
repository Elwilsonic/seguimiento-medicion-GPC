package member

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// GetByID devuelve el integrante con ese ID, o ErrMemberNotFound si no existe
// (HU-16).
func (r *PostgresRepository) GetByID(id int) (Member, error) {
	row := r.pool.QueryRow(context.Background(),
		"SELECT "+memberColumns+" FROM members WHERE id = $1", id)
	m, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrMemberNotFound
	}
	if err != nil {
		return Member{}, fmt.Errorf("leer el integrante: %w", err)
	}
	return m, nil
}

// Update reemplaza el nombre y el rol del integrante con el ID de m, sin
// cambiar su proyecto, y devuelve el integrante tal como quedó guardado, o
// ErrMemberNotFound si no existe (HU-16).
func (r *PostgresRepository) Update(m Member) (Member, error) {
	row := r.pool.QueryRow(context.Background(),
		"UPDATE members SET name = $2, role = $3 WHERE id = $1 RETURNING "+memberColumns,
		m.ID, m.Name, m.Role)
	updated, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrMemberNotFound
	}
	if err != nil {
		return Member{}, fmt.Errorf("modificar el integrante: %w", err)
	}
	return updated, nil
}
