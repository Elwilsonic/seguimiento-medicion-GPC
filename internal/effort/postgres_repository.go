package effort

import "github.com/jackc/pgx/v5/pgxpool"

// PostgresRepository guarda los esfuerzos en PostgreSQL (RN-26.10).
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository crea un repositorio sobre pool. El pool lo cierra
// quien lo abrió.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Save(effort Effort) (Effort, error) {
	return Effort{}, errNotImplemented
}
