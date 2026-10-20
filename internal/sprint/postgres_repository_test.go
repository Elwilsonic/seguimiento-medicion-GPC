package sprint

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"seguimiento-medicion-gpc/internal/postgrestest"
	"seguimiento-medicion-gpc/internal/project"
)

// TEC-05 · Persistencia con PostgreSQL en Docker (Issue #35). Estos tests
// necesitan TEST_DATABASE_URL; sin ella (como en el CI) se omiten
// (CA-TEC05.7).

// newPostgresTestRepository vacía la base de pruebas, crea los proyectos
// contractProjectA y contractProjectB (la tabla de sprints tiene clave
// foránea a proyectos) y devuelve un repositorio PostgreSQL sobre pool.
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

// Solo PostgreSQL: las fechas se devuelven en UTC y con precisión de
// microsegundos, que es lo que guarda PostgreSQL, tanto al guardar como al
// leer.
//
// pgx devuelve timestamptz en time.Local. El test fija time.Local en -03:00
// (y lo restaura al terminar), así un repositorio que no normalice a UTC
// falla también en una máquina o un CI configurados en UTC.
func TestPostgresRepository_FechasEnUTCConMicrosegundos(t *testing.T) {
	zone := time.FixedZone("UTC-3", -3*60*60)
	previousLocal := time.Local
	time.Local = zone
	t.Cleanup(func() { time.Local = previousLocal })

	repo := newPostgresTestRepository(t, postgrestest.Pool(t))
	sprint := contractSprint(contractProjectA, "Sprint 1", StatusOpen)
	sprint.StartDate = time.Date(2026, 10, 16, 9, 30, 0, 123456789, zone)
	sprint.EndDate = time.Date(2026, 10, 22, 18, 0, 0, 987654321, zone)
	wantStart := sprint.StartDate.Truncate(time.Microsecond).UTC()
	wantEnd := sprint.EndDate.Truncate(time.Microsecond).UTC()

	saved := contractSave(t, repo, sprint)
	read, _, err := repo.OpenByProject(contractProjectA)
	if err != nil {
		t.Fatalf("OpenByProject devolvió un error inesperado: %v", err)
	}

	for op, got := range map[string]Sprint{"Save": saved, "OpenByProject": read} {
		if got.StartDate != wantStart || got.EndDate != wantEnd {
			t.Errorf("%s: fechas = %v a %v, se esperaba %v a %v", op, got.StartDate, got.EndDate, wantStart, wantEnd)
		}
	}
}

// Solo PostgreSQL: guardar un sprint de un proyecto inexistente falla por la
// clave foránea (23503, foreign_key_violation). La memoria no lo valida; lo
// valida el Manager (HU-10).
func TestPostgresRepository_SaveConProyectoInexistenteFalla(t *testing.T) {
	repo := newPostgresTestRepository(t, postgrestest.Pool(t))

	_, err := repo.Save(contractSprint(99, "Huérfano", StatusOpen))

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("err = %v, se esperaba una violación de clave foránea (23503)", err)
	}
}

// HU-13: Update persiste el estado, conserva las fechas en UTC con precisión
// de microsegundos y devuelve una copia (CA-13.1 y RN-13.4). Necesita la base
// levantada; sin TEST_DATABASE_URL se omite.
func TestPostgresRepository_UpdatePersisteElEstadoYDevuelveFechasUTC(t *testing.T) {
	repo := newPostgresTestRepository(t, postgrestest.Pool(t))
	saved := contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusOpen))

	updated := saved
	updated.Name = "Sprint cerrado"
	updated.Goal = "Cerrar la iteración"
	updated.Status = StatusClosed

	got, err := repo.Update(updated)
	if err != nil {
		t.Fatalf("Update devolvió un error inesperado: %v", err)
	}
	if got.ID != saved.ID {
		t.Errorf("ID devuelto = %d, se esperaba %d", got.ID, saved.ID)
	}
	if got.Name != updated.Name || got.Goal != updated.Goal || got.Status != StatusClosed {
		t.Errorf("sprint devuelto = %+v, se esperaba %+v", got, updated)
	}
	if got.StartDate != contractStart.UTC() || got.EndDate != contractEnd.UTC() {
		t.Errorf("fechas devueltas = %v a %v, se esperaban %v a %v", got.StartDate, got.EndDate, contractStart.UTC(), contractEnd.UTC())
	}

	stored, err := repo.GetByID(saved.ID)
	if err != nil {
		t.Fatalf("GetByID después de Update devolvió un error inesperado: %v", err)
	}
	if stored.Status != StatusClosed || stored.Name != updated.Name {
		t.Errorf("sprint guardado = %+v, se esperaba %+v", stored, updated)
	}
}
