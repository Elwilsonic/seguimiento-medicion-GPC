package effort

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository guarda los esfuerzos en PostgreSQL (RN-26.10). Cumple el
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

// effortColumns son las columnas de un esfuerzo, en el orden de scanEffort.
const effortColumns = "id, member, effort_date, activity, item_id, sprint_id, hours"

// Save guarda el esfuerzo; la base asigna el siguiente ID correlativo global
// (RN-26.8). Devuelve el esfuerzo tal como quedó guardado, con la fecha
// normalizada a UTC como en los sprints.
func (r *PostgresRepository) Save(effort Effort) (Effort, error) {
	row := r.pool.QueryRow(context.Background(),
		`INSERT INTO efforts (member, effort_date, activity, item_id, sprint_id, hours)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+effortColumns,
		effort.Member, effort.Date, effort.Activity, effort.ItemID, effort.SprintID, effort.Hours)
	saved, err := scanEffort(row)
	if err != nil {
		return Effort{}, fmt.Errorf("guardar el esfuerzo: %w", err)
	}
	return saved, nil
}

// scanEffort lee un esfuerzo con las columnas de effortColumns y normaliza la
// fecha a UTC: pgx devuelve timestamptz en la zona local de la máquina.
func scanEffort(row pgx.Row) (Effort, error) {
	var effort Effort
	err := row.Scan(&effort.ID, &effort.Member, &effort.Date, &effort.Activity,
		&effort.ItemID, &effort.SprintID, &effort.Hours)
	effort.Date = effort.Date.UTC()
	return effort, err
}
