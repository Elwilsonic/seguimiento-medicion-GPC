package defect

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

// HU-29 · Registrar defecto (Issue #47). Estos tests necesitan
// TEST_DATABASE_URL; sin ella (como en el CI) se omiten (CA-TEC05.7).

var contractStart = time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)

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
		StartDate: contractStart,
		EndDate:   contractStart.AddDate(0, 0, 6),
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

// stored devuelve una copia de los defectos guardados en PostgreSQL.
func (r *PostgresRepository) stored(t *testing.T) []Defect {
	t.Helper()
	rows, err := r.pool.Query(context.Background(),
		"SELECT "+defectColumns+" FROM defects ORDER BY id")
	if err != nil {
		t.Fatalf("leer los defectos guardados: %v", err)
	}
	defer rows.Close()

	var defects []Defect
	for rows.Next() {
		defect, err := scanDefect(rows)
		if err != nil {
			t.Fatalf("leer un defecto guardado: %v", err)
		}
		defects = append(defects, defect)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("leer los defectos guardados: %v", err)
	}
	return defects
}

// El repositorio PostgreSQL cumple el mismo contrato que el de memoria.
func TestPostgresRepository_Contrato(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryContract(t, func(t *testing.T) Repository {
		return newPostgresTestRepository(t, pool)
	})
}

// Guarda el defecto en la tabla defects con todos sus campos.
func TestPostgresRepository_GuardaElDefecto(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)

	saved := contractSave(t, repo, contractDefect("Descripción alta", SeverityHigh))

	got, err := scanDefect(pool.QueryRow(context.Background(),
		"SELECT "+defectColumns+" FROM defects WHERE id = $1", saved.ID))
	if err != nil {
		t.Fatalf("leer el defecto guardado: %v", err)
	}
	contractAssertDefect(t, "defecto guardado", got, saved)
}

// Un ítem inexistente falla por la clave foránea (23503).
func TestPostgresRepository_ItemInexistenteFalla(t *testing.T) {
	repo := newPostgresTestRepository(t, postgrestest.Pool(t))
	defect := contractDefect("Descripción alta", SeverityHigh)
	defect.ItemID = 999

	_, err := repo.Save(defect)

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("err = %v, se esperaba una violación de clave foránea (23503)", err)
	}
}

// Un sprint de detección inexistente falla por la clave foránea (23503).
func TestPostgresRepository_SprintDeDeteccionInexistenteFalla(t *testing.T) {
	repo := newPostgresTestRepository(t, postgrestest.Pool(t))
	defect := contractDefect("Descripción alta", SeverityHigh)
	defect.DetectionSprintID = 999

	_, err := repo.Save(defect)

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("err = %v, se esperaba una violación de clave foránea (23503)", err)
	}
}

// El sprint de resolución 0 se guarda sin clave foránea (RN-29.4).
func TestPostgresRepository_SprintDeResolucionCeroSeGuarda(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)

	saved := contractSave(t, repo, contractDefect("Descripción media", SeverityMedium))
	if saved.ResolutionSprintID != 0 {
		t.Errorf("sprint de resolución = %d, se esperaba 0", saved.ResolutionSprintID)
	}

	var resolutionSprintID int
	if err := pool.QueryRow(context.Background(),
		"SELECT resolution_sprint_id FROM defects WHERE id = $1", saved.ID).
		Scan(&resolutionSprintID); err != nil {
		t.Fatalf("leer el sprint de resolución: %v", err)
	}
	if resolutionSprintID != 0 {
		t.Errorf("sprint de resolución guardado = %d, se esperaba 0", resolutionSprintID)
	}
}

// HU-30 · Cambiar el estado de un defecto (Issue #48). GetByID y Update
// cumplen el mismo contrato que en memoria.
func TestPostgresRepository_ContratoGetByIDYUpdate(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryStatusContract(t, func(t *testing.T) Repository {
		return newPostgresTestRepository(t, pool)
	})
}

// Update guarda el estado y el sprint de resolución en la tabla defects, sin
// cambios de esquema (RN-30.9).
func TestPostgresRepository_UpdateGuardaElCambioEnLaTabla(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)
	saved := contractSave(t, repo, contractDefect("Descripción alta", SeverityHigh))
	changed := saved
	changed.Status = StatusResolved
	changed.ResolutionSprintID = contractDetectionSprintID

	if _, err := repo.Update(changed); err != nil {
		t.Fatalf("Update devolvió un error inesperado: %v", err)
	}

	got, err := scanDefect(pool.QueryRow(context.Background(),
		"SELECT "+defectColumns+" FROM defects WHERE id = $1", saved.ID))
	if err != nil {
		t.Fatalf("leer el defecto actualizado: %v", err)
	}
	contractAssertDefect(t, "defecto actualizado", got, changed)
}

// Update de un ID inexistente no crea filas.
func TestPostgresRepository_UpdateInexistenteNoCreaFilas(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)
	missing := contractDefectWithID(999, "Descripción alta", SeverityHigh)

	if _, err := repo.Update(missing); !errors.Is(err, ErrDefectNotFound) {
		t.Fatalf("Update: err = %v, se esperaba %v", err, ErrDefectNotFound)
	}
	var count int
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM defects").Scan(&count); err != nil {
		t.Fatalf("contar los defectos: %v", err)
	}
	if count != 0 {
		t.Errorf("defectos en la tabla = %d, se esperaba ninguno", count)
	}
}

// HU-31 · Consultar los defectos de un proyecto o de un sprint (Issue #49).
// List cumple el mismo contrato que en memoria.
func TestPostgresRepository_ContratoList(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryListContract(t, func(t *testing.T) Repository {
		return newPostgresTestRepository(t, pool)
	})
}

// List lee la tabla defects completa, ordenada por ID, sin cambios de esquema
// (RN-31.8 y D5).
func TestPostgresRepository_ListLeeLaTablaOrdenadaPorID(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)
	for _, description := range []string{"Primero", "Segundo", "Tercero"} {
		contractSave(t, repo, contractDefect(description, SeverityHigh))
	}
	want := repo.(*PostgresRepository).stored(t)

	got, err := repo.List()
	if err != nil {
		t.Fatalf("List devolvió un error inesperado: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("List = %+v, se esperaba %+v", got, want)
	}
	for i := range want {
		contractAssertDefect(t, "List", got[i], want[i])
	}
}
