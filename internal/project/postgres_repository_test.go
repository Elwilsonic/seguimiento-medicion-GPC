package project

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

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

// HU-18, CA-18.9: un proyecto sin fecha de inicio se guarda como NULL y se lee
// sin fecha. Necesita la base levantada; sin TEST_DATABASE_URL se omite.
func TestPostgresRepository_SinFechaDeInicioSeGuardaComoNULL(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)

	saved := contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned})

	var isNull bool
	err := pool.QueryRow(context.Background(), "SELECT start_date IS NULL FROM projects WHERE id = $1", saved.ID).Scan(&isNull)
	if err != nil {
		t.Fatalf("leer start_date: %v", err)
	}
	if !isNull {
		t.Error("start_date no es NULL para un proyecto sin fecha de inicio")
	}
	read, err := repo.GetByID(saved.ID)
	if err != nil {
		t.Fatalf("GetByID devolvió un error inesperado: %v", err)
	}
	if !read.StartDate.IsZero() {
		t.Errorf("StartDate = %v, se esperaba sin registrar", read.StartDate)
	}
}

// HU-18, CA-18.9: la fecha de inicio se lee en UTC, el mismo instante y el
// mismo día, aunque la zona de la máquina sea UTC-3 (patrón de TEC-05).
// Necesita la base levantada; sin TEST_DATABASE_URL se omite.
func TestPostgresRepository_FechaDeInicioEnUTC(t *testing.T) {
	zone := time.FixedZone("UTC-3", -3*60*60)
	previousLocal := time.Local
	time.Local = zone
	t.Cleanup(func() { time.Local = previousLocal })

	repo := newPostgresTestRepository(t, postgrestest.Pool(t))
	saved := contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned})
	saved.StartDate = time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)

	updated, err := repo.Update(saved)
	if err != nil {
		t.Fatalf("Update devolvió un error inesperado: %v", err)
	}
	read, err := repo.GetByID(saved.ID)
	if err != nil {
		t.Fatalf("GetByID devolvió un error inesperado: %v", err)
	}

	for op, got := range map[string]Project{"Update": updated, "GetByID": read} {
		if got.StartDate != saved.StartDate || got.StartDate.Location() != time.UTC {
			t.Errorf("%s: StartDate = %v, se esperaba %v en UTC", op, got.StartDate, saved.StartDate)
		}
		if got.StartDate.Format("2006-01-02") != "2026-10-23" {
			t.Errorf("%s: día = %s, se esperaba 2026-10-23", op, got.StartDate.Format("2006-01-02"))
		}
	}
}

// HU-41 · Registrar la fecha de fin del proyecto (Issue #115). Estos tests
// necesitan TEST_DATABASE_URL y la base con la columna end_date; sin la
// variable se omiten.

// CA-41.12: el repositorio PostgreSQL conserva la fecha de fin con el mismo
// contrato que el de memoria.
func TestPostgresRepository_ContratoFechaDeFin(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryEndDateContract(t, func(t *testing.T) Repository {
		return newPostgresTestRepository(t, pool)
	})
}

// CA-41.12: un proyecto sin fecha de fin se guarda como NULL y se lee sin
// fecha.
func TestPostgresRepository_SinFechaDeFinSeGuardaComoNULL(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)

	saved := contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned, StartDate: contractStartDate})

	var isNull bool
	err := pool.QueryRow(context.Background(), "SELECT end_date IS NULL FROM projects WHERE id = $1", saved.ID).Scan(&isNull)
	if err != nil {
		t.Fatalf("leer end_date: %v", err)
	}
	if !isNull {
		t.Error("end_date no es NULL para un proyecto sin fecha de fin")
	}
	read, err := repo.GetByID(saved.ID)
	if err != nil {
		t.Fatalf("GetByID devolvió un error inesperado: %v", err)
	}
	if !read.EndDate.IsZero() {
		t.Errorf("EndDate = %v, se esperaba sin registrar", read.EndDate)
	}
}

// CA-41.12: la fecha de fin se lee en UTC, el mismo instante y el mismo día,
// aunque la zona de la máquina sea UTC-3 (patrón de TEC-05 y HU-18).
func TestPostgresRepository_FechaDeFinEnUTC(t *testing.T) {
	zone := time.FixedZone("UTC-3", -3*60*60)
	previousLocal := time.Local
	time.Local = zone
	t.Cleanup(func() { time.Local = previousLocal })

	repo := newPostgresTestRepository(t, postgrestest.Pool(t))
	saved := contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned, StartDate: contractStartDate})
	saved.EndDate = time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC)

	updated, err := repo.Update(saved)
	if err != nil {
		t.Fatalf("Update devolvió un error inesperado: %v", err)
	}
	read, err := repo.GetByID(saved.ID)
	if err != nil {
		t.Fatalf("GetByID devolvió un error inesperado: %v", err)
	}

	for op, got := range map[string]Project{"Update": updated, "GetByID": read} {
		if got.EndDate != saved.EndDate || got.EndDate.Location() != time.UTC {
			t.Errorf("%s: EndDate = %v, se esperaba %v en UTC", op, got.EndDate, saved.EndDate)
		}
		if got.EndDate.Format("2006-01-02") != "2026-12-18" {
			t.Errorf("%s: día = %s, se esperaba 2026-12-18", op, got.EndDate.Format("2006-01-02"))
		}
	}
}
