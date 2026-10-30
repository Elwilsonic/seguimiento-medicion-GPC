package member

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"seguimiento-medicion-gpc/internal/postgrestest"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-15 · Registrar integrante del proyecto (Issue #29). Estos tests
// necesitan TEST_DATABASE_URL; sin ella (como en el CI) se omiten
// (CA-TEC05.7).

// newPostgresTestRepository vacía la base de pruebas, crea los proyectos 1 y 2
// que necesita la clave foránea y devuelve un repositorio PostgreSQL.
func newPostgresTestRepository(t *testing.T, pool *pgxpool.Pool) Repository {
	t.Helper()
	postgrestest.Reset(t, pool)
	projects := project.NewPostgresRepository(pool)
	for _, name := range []string{"Proyecto A", "Proyecto B"} {
		if _, err := projects.Save(project.Project{Name: name, Status: project.StatusPlanned}); err != nil {
			t.Fatalf("crear el proyecto de prueba %q: %v", name, err)
		}
	}
	return NewPostgresRepository(pool)
}

// El repositorio PostgreSQL cumple el mismo contrato que el de memoria
// (CA-15.6).
func TestPostgresRepository_Contrato(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryContract(t, func(t *testing.T) Repository {
		return newPostgresTestRepository(t, pool)
	})
}

// Guarda el integrante en la tabla members con todos sus campos.
func TestPostgresRepository_GuardaElIntegrante(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)

	saved := contractSave(t, repo, Member{ProjectID: 1, Name: "Ana Pérez", Role: RoleAgileEnabler})

	var got Member
	err := pool.QueryRow(context.Background(),
		"SELECT id, project_id, name, role FROM members WHERE id = $1", saved.ID).
		Scan(&got.ID, &got.ProjectID, &got.Name, &got.Role)
	if err != nil {
		t.Fatalf("leer el integrante guardado: %v", err)
	}
	if got != saved {
		t.Errorf("integrante guardado = %+v, se esperaba %+v", got, saved)
	}
}

// La clave foránea impide guardar un integrante de un proyecto inexistente.
func TestPostgresRepository_RechazaUnProyectoInexistente(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)

	if _, err := repo.Save(Member{ProjectID: 99, Name: "Ana", Role: RoleProductBuilder}); err == nil {
		t.Error("Save en el proyecto 99 no devolvió error")
	}
}
