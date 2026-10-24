package poker

import "github.com/jackc/pgx/v5/pgxpool"

// PostgresRepository guarda las sesiones en PostgreSQL (TEC-05). Cumple el
// mismo contrato que MemoryRepository.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository crea un repositorio sobre pool. El pool lo cierra
// quien lo abrió.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Save guarda la sesión; la base asigna el siguiente ID correlativo global
// (RN-21.7).
func (r *PostgresRepository) Save(session Session) (Session, error) {
	return Session{}, errNotImplemented
}

// OpenByItem devuelve una copia de la sesión ABIERTA del ítem y si existe
// (RN-21.2).
func (r *PostgresRepository) OpenByItem(itemID int) (Session, bool, error) {
	return Session{}, false, errNotImplemented
}
