package member

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository guarda los integrantes en PostgreSQL (RN-15.11). Cumple
// el mismo contrato que MemoryRepository. Repository no recibe un contexto,
// así que cada consulta usa context.Background().
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository crea un repositorio sobre pool. El pool lo cierra
// quien lo abrió.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// memberColumns son las columnas de un integrante, en el orden de scanMember.
const memberColumns = "id, project_id, name, role"

// Save guarda el integrante; la base asigna el siguiente ID correlativo
// global (RN-15.7). Devuelve el integrante tal como quedó guardado.
func (r *PostgresRepository) Save(m Member) (Member, error) {
	row := r.pool.QueryRow(context.Background(),
		"INSERT INTO members (project_id, name, role) VALUES ($1, $2, $3) RETURNING "+memberColumns,
		m.ProjectID, m.Name, m.Role)
	saved, err := scanMember(row)
	if err != nil {
		return Member{}, fmt.Errorf("guardar el integrante: %w", err)
	}
	return saved, nil
}

// ListByProject devuelve los integrantes del proyecto, ordenados por ID. Sin
// integrantes devuelve una lista vacía, no nil.
func (r *PostgresRepository) ListByProject(projectID int) ([]Member, error) {
	rows, err := r.pool.Query(context.Background(),
		"SELECT "+memberColumns+" FROM members WHERE project_id = $1 ORDER BY id", projectID)
	if err != nil {
		return nil, fmt.Errorf("listar los integrantes: %w", err)
	}
	defer rows.Close()

	list := []Member{}
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, fmt.Errorf("leer un integrante: %w", err)
		}
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listar los integrantes: %w", err)
	}
	return list, nil
}

// scanMember lee un integrante con las columnas de memberColumns.
func scanMember(row pgx.Row) (Member, error) {
	var m Member
	err := row.Scan(&m.ID, &m.ProjectID, &m.Name, &m.Role)
	return m, err
}
