package member

import "github.com/jackc/pgx/v5/pgxpool"

// PostgresRepository guarda los integrantes en PostgreSQL (RN-15.11).
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository crea un repositorio sobre pool. El pool lo cierra
// quien lo abrió.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Save guarda el integrante.
func (r *PostgresRepository) Save(m Member) (Member, error) {
	return Member{}, errNotImplemented
}

// ListByProject devuelve los integrantes de un proyecto.
func (r *PostgresRepository) ListByProject(projectID int) ([]Member, error) {
	return nil, errNotImplemented
}
