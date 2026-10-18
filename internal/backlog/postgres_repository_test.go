package backlog

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"seguimiento-medicion-gpc/internal/postgres"
	"seguimiento-medicion-gpc/internal/postgrestest"
	"seguimiento-medicion-gpc/internal/project"
)

// TEC-05 · Persistencia con PostgreSQL en Docker (Issue #35). Estos tests
// necesitan TEST_DATABASE_URL; sin ella (como en el CI) se omiten
// (CA-TEC05.7).

// newPostgresTestRepository vacía la base de pruebas, crea los proyectos
// contractProjectA y contractProjectB (la tabla de ítems tiene clave foránea
// a proyectos) y devuelve un repositorio PostgreSQL sobre pool.
func newPostgresTestRepository(t *testing.T, pool *pgxpool.Pool) Repository {
	t.Helper()
	postgrestest.Reset(t, pool)
	projects := project.NewPostgresRepository(pool)
	for _, name := range []string{"Proyecto A", "Proyecto B"} {
		if _, err := projects.Save(project.Project{Name: name, Status: project.StatusPlanned}); err != nil {
			t.Fatalf("crear el proyecto de prueba: %v", err)
		}
	}
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

// Solo PostgreSQL: guardar un ítem de un proyecto inexistente falla por la
// clave foránea (23503, foreign_key_violation). La memoria no lo valida; lo
// valida el Manager (HU-04).
func TestPostgresRepository_SaveConProyectoInexistenteFalla(t *testing.T) {
	repo := newPostgresTestRepository(t, postgrestest.Pool(t))

	_, err := repo.Save(contractItem(99, "Huérfano"))

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("err = %v, se esperaba una violación de clave foránea (23503)", err)
	}
}

// Un error de la base (acá, un pool cerrado) se devuelve como error y no se
// confunde con ErrItemNotFound: el ítem puede existir aunque la base no
// responda. Se usa un pool propio, no el de Pool(t), que tiene tomado el lock
// hasta que termina el test.
func TestPostgresRepository_ErrorDeLaBaseNoEsNotFound(t *testing.T) {
	postgrestest.Pool(t)
	closed, err := postgres.Open(context.Background(), os.Getenv(postgrestest.EnvVar))
	if err != nil {
		t.Fatalf("abrir el pool: %v", err)
	}
	closed.Close()
	repo := NewPostgresRepository(closed)
	item := contractItem(contractProjectA, "Login")
	item.ID = 1

	_, getErr := repo.GetByID(1)
	_, updateErr := repo.Update(item)

	for op, err := range map[string]error{"GetByID": getErr, "Update": updateErr} {
		if err == nil || errors.Is(err, ErrItemNotFound) {
			t.Errorf("%s: err = %v, se esperaba un error de la base distinto de %v", op, err, ErrItemNotFound)
		}
	}
}
