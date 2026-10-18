package project

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"seguimiento-medicion-gpc/internal/postgrestest"
)

// TEC-05 · Persistencia con PostgreSQL en Docker (Issue #35). Estos tests
// necesitan TEST_DATABASE_URL; sin ella (como en el CI) se omiten
// (CA-TEC05.7).

// newPostgresTestRepository vacía la base de pruebas y devuelve un
// repositorio PostgreSQL sobre pool.
func newPostgresTestRepository(t *testing.T, pool *pgxpool.Pool) Repository {
	t.Helper()
	postgrestest.Reset(t, pool)
	return NewPostgresRepository(pool)
}

// CA-TEC05.3 · El repositorio PostgreSQL cumple el mismo contrato que el de
// memoria.
func TestPostgresRepository_Contrato(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryContract(t, func(t *testing.T) Repository {
		return newPostgresTestRepository(t, pool)
	})
}
