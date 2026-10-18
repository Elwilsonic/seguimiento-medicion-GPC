package sprint

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository guarda los sprints en PostgreSQL (TEC-05). Cumple el
// mismo contrato que MemoryRepository. Repository no recibe un contexto, así
// que cada consulta usa context.Background().
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository crea un repositorio sobre pool. El pool lo cierra
// quien lo abrió.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// sprintColumns son las columnas de un sprint, en el orden de scanSprint.
const sprintColumns = "id, project_id, name, start_date, end_date, goal, status"

// Save guarda el sprint; la base asigna el siguiente ID correlativo global
// (RN-10.6). Devuelve el sprint tal como quedó guardado: las fechas en UTC y
// con precisión de microsegundos, la de PostgreSQL.
func (r *PostgresRepository) Save(sprint Sprint) (Sprint, error) {
	row := r.pool.QueryRow(context.Background(),
		`INSERT INTO sprints (project_id, name, start_date, end_date, goal, status)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+sprintColumns,
		sprint.ProjectID, sprint.Name, sprint.StartDate, sprint.EndDate, sprint.Goal, sprint.Status)
	saved, err := scanSprint(row)
	if err != nil {
		return Sprint{}, fmt.Errorf("guardar el sprint: %w", err)
	}
	return saved, nil
}

// OpenByProject devuelve el sprint ABIERTO del proyecto y si existe
// (RN-10.5). Un sprint ABIERTO de otro proyecto no aparece.
func (r *PostgresRepository) OpenByProject(projectID int) (Sprint, bool, error) {
	row := r.pool.QueryRow(context.Background(),
		"SELECT "+sprintColumns+" FROM sprints WHERE project_id = $1 AND status = $2 ORDER BY id LIMIT 1",
		projectID, StatusOpen)
	sprint, err := scanSprint(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Sprint{}, false, nil
	}
	if err != nil {
		return Sprint{}, false, fmt.Errorf("buscar el sprint abierto: %w", err)
	}
	return sprint, true, nil
}

// scanSprint lee un sprint con las columnas de sprintColumns y normaliza las
// fechas a UTC: pgx devuelve timestamptz en la zona local de la máquina.
func scanSprint(row pgx.Row) (Sprint, error) {
	var sprint Sprint
	err := row.Scan(&sprint.ID, &sprint.ProjectID, &sprint.Name, &sprint.StartDate, &sprint.EndDate, &sprint.Goal, &sprint.Status)
	sprint.StartDate = sprint.StartDate.UTC()
	sprint.EndDate = sprint.EndDate.UTC()
	return sprint, err
}
