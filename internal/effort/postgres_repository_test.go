package effort

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/postgrestest"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-26 · Registrar esfuerzo (Issue #44). Estos tests necesitan
// TEST_DATABASE_URL; sin ella (como en el CI) se omiten (CA-TEC05.7).

// newPostgresTestRepository vacía la base de pruebas, crea el proyecto, el
// sprint y el ítem que necesitan las claves foráneas y devuelve un repositorio
// PostgreSQL sobre pool.
func newPostgresTestRepository(t *testing.T, pool *pgxpool.Pool) Repository {
	t.Helper()
	postgrestest.Reset(t, pool)

	projects := project.NewPostgresRepository(pool)
	if _, err := projects.Save(project.Project{Name: "Proyecto A", Status: project.StatusPlanned}); err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}

	sprints := sprint.NewPostgresRepository(pool)
	savedSprint, err := sprints.Save(sprint.Sprint{
		ProjectID: 1,
		Name:      "Sprint 1",
		StartDate: contractDate,
		EndDate:   contractDate.AddDate(0, 0, 6),
		Goal:      "Entregar el MVP",
		Status:    sprint.StatusOpen,
	})
	if err != nil {
		t.Fatalf("crear el sprint de prueba: %v", err)
	}

	items := backlog.NewPostgresRepository(pool)
	if _, err := items.Save(backlog.Item{
		ProjectID:          1,
		Title:              "Login de usuarios",
		Description:        "Formulario de acceso",
		Priority:           backlog.PriorityHigh,
		AcceptanceCriteria: []string{"Login OK"},
		Status:             backlog.StatusInProgress,
		SprintID:           savedSprint.ID,
	}); err != nil {
		t.Fatalf("crear el ítem de prueba: %v", err)
	}

	return NewPostgresRepository(pool)
}

// El repositorio PostgreSQL cumple el mismo contrato que el de memoria.
func TestPostgresRepository_Contrato(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryContract(t, func(t *testing.T) Repository {
		return newPostgresTestRepository(t, pool)
	})
}

// Guarda el esfuerzo en la tabla efforts con todos sus campos.
func TestPostgresRepository_GuardaElEsfuerzo(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)

	saved := contractSave(t, repo, contractEffort("García Santiago", "Implementar el registro", 2.5))

	var got Effort
	err := pool.QueryRow(context.Background(),
		"SELECT id, member, effort_date, activity, item_id, sprint_id, hours FROM efforts WHERE id = $1", saved.ID).
		Scan(&got.ID, &got.Member, &got.Date, &got.Activity, &got.ItemID, &got.SprintID, &got.Hours)
	if err != nil {
		t.Fatalf("leer el esfuerzo guardado: %v", err)
	}
	contractAssertEffort(t, "esfuerzo guardado", got, saved)
}

// La fecha se guarda y se devuelve normalizada a UTC, como en los sprints.
func TestPostgresRepository_NormalizaLaFechaAUTC(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)
	art := time.FixedZone("ART", -3*60*60)
	date := time.Date(2026, 10, 24, 21, 30, 0, 0, art)
	effort := contractEffort("García Santiago", "Implementar", 2)
	effort.Date = date

	saved := contractSave(t, repo, effort)

	if saved.Date.Location() != time.UTC {
		t.Errorf("zona = %v, se esperaba UTC", saved.Date.Location())
	}
	if !saved.Date.Equal(date) {
		t.Errorf("fecha = %v, se esperaba el mismo instante que %v", saved.Date, date)
	}
}

// Un ítem inexistente falla por la clave foránea (23503).
func TestPostgresRepository_SaveDeItemInexistenteFalla(t *testing.T) {
	repo := newPostgresTestRepository(t, postgrestest.Pool(t))
	effort := contractEffort("García Santiago", "Implementar", 2)
	effort.ItemID = 999

	_, err := repo.Save(effort)

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("err = %v, se esperaba una violación de clave foránea (23503)", err)
	}
}

// Un sprint inexistente falla por la clave foránea (23503).
func TestPostgresRepository_SaveDeSprintInexistenteFalla(t *testing.T) {
	repo := newPostgresTestRepository(t, postgrestest.Pool(t))
	effort := contractEffort("García Santiago", "Implementar", 2)
	effort.SprintID = 999

	_, err := repo.Save(effort)

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("err = %v, se esperaba una violación de clave foránea (23503)", err)
	}
}
