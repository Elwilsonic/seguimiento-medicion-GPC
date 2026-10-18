package project

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"seguimiento-medicion-gpc/internal/postgres"
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

// Un error de la base (acá, un pool cerrado) se devuelve como error y no se
// confunde con ErrProjectNotFound: el proyecto puede existir aunque la base
// no responda. Se usa un pool propio, no el de Pool(t), que tiene tomado el
// lock hasta que termina el test.
func TestPostgresRepository_ErrorDeLaBaseNoEsNotFound(t *testing.T) {
	postgrestest.Pool(t)
	closed, err := postgres.Open(context.Background(), os.Getenv(postgrestest.EnvVar))
	if err != nil {
		t.Fatalf("abrir el pool: %v", err)
	}
	closed.Close()
	repo := NewPostgresRepository(closed)

	_, getErr := repo.GetByID(1)
	_, updateErr := repo.Update(Project{ID: 1, Name: "Alfa", Status: StatusPlanned})

	for op, err := range map[string]error{"GetByID": getErr, "Update": updateErr} {
		if err == nil || errors.Is(err, ErrProjectNotFound) {
			t.Errorf("%s: err = %v, se esperaba un error de la base distinto de %v", op, err, ErrProjectNotFound)
		}
	}
}
