package poker

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/postgrestest"
	"seguimiento-medicion-gpc/internal/project"
)

// CA-21.8 · Persistencia en PostgreSQL. Estos tests necesitan
// TEST_DATABASE_URL; sin ella (como en el CI) se omiten.

// newPostgresPokerRepo vacía la base de pruebas, crea un proyecto y un ítem
// (la tabla de sesiones tiene clave foránea a backlog_items) y devuelve un
// repositorio PostgreSQL y el ID del ítem.
func newPostgresPokerRepo(t *testing.T, pool *pgxpool.Pool) (Repository, int) {
	t.Helper()
	postgrestest.Reset(t, pool)
	projects := project.NewManager(project.NewPostgresRepository(pool))
	p, err := projects.Create("Proyecto de prueba", "Proyecto para el contrato de sesiones")
	if err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	items := backlog.NewManager(backlog.NewPostgresRepository(pool), projects)
	it, err := items.Create(p.ID, "Historia de ejemplo", "Descripción", "MEDIA", []string{"Criterio"})
	if err != nil {
		t.Fatalf("crear el ítem de prueba: %v", err)
	}
	return NewPostgresRepository(pool), it.ID
}

// CA-21.8 · El repositorio PostgreSQL cumple el mismo contrato que el de
// memoria.
func TestPostgresRepository_Contrato(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryContract(t, func(t *testing.T) (Repository, int) {
		return newPostgresPokerRepo(t, pool)
	})
}

// CA-21.8 · La clave foránea impide guardar una sesión de un ítem inexistente.
func TestPostgresRepository_SaveConItemInexistenteFalla(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := NewPostgresRepository(pool)

	_, err := repo.Save(Session{ItemID: 999999, Status: StatusOpen, Round: 1, Votes: []Vote{}})

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Fatalf("error = %v, se esperaba una violación de clave foránea (23503)", err)
	}
}
