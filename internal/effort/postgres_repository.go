package effort

import (
	"context"
	"errors"
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

// ListByItem devuelve los esfuerzos de una historia, ordenados por ID.
func (r *PostgresRepository) ListByItem(itemID int) ([]Effort, error) {
	return r.listEfforts("item_id", itemID)
}

// ListBySprint devuelve los esfuerzos de un sprint, ordenados por ID.
func (r *PostgresRepository) ListBySprint(sprintID int) ([]Effort, error) {
	return r.listEfforts("sprint_id", sprintID)
}

// listEfforts lista los esfuerzos que cumplen el filtro indicado. El filtro es
// una columna interna del repositorio, nunca un valor recibido del llamador.
func (r *PostgresRepository) listEfforts(filter string, id int) ([]Effort, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT `+effortColumns+` FROM efforts WHERE `+filter+` = $1 ORDER BY id`, id)
	if err != nil {
		return nil, fmt.Errorf("listar los esfuerzos: %w", err)
	}
	defer rows.Close()

	efforts := make([]Effort, 0)
	for rows.Next() {
		saved, err := scanEffort(rows)
		if err != nil {
			return nil, fmt.Errorf("leer el esfuerzo: %w", err)
		}
		efforts = append(efforts, saved)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listar los esfuerzos: %w", err)
	}
	return efforts, nil
}

// GetByID devuelve el esfuerzo con ese ID o ErrEffortNotFound si no existe,
// incluidos los IDs 0 y negativos (HU-28).
func (r *PostgresRepository) GetByID(id int) (Effort, error) {
	row := r.pool.QueryRow(context.Background(),
		"SELECT "+effortColumns+" FROM efforts WHERE id = $1", id)
	effort, err := scanEffort(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Effort{}, ErrEffortNotFound
	}
	if err != nil {
		return Effort{}, fmt.Errorf("buscar el esfuerzo: %w", err)
	}
	return effort, nil
}

// Update reemplaza el esfuerzo con ese ID y devuelve el registro tal como
// quedó guardado (HU-28). Si no existe, devuelve ErrEffortNotFound sin crear
// ni modificar filas.
func (r *PostgresRepository) Update(effort Effort) (Effort, error) {
	row := r.pool.QueryRow(context.Background(),
		`UPDATE efforts
		SET member = $2, effort_date = $3, activity = $4, item_id = $5, sprint_id = $6, hours = $7
		WHERE id = $1 RETURNING `+effortColumns,
		effort.ID, effort.Member, effort.Date, effort.Activity, effort.ItemID, effort.SprintID, effort.Hours)
	updated, err := scanEffort(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Effort{}, ErrEffortNotFound
	}
	if err != nil {
		return Effort{}, fmt.Errorf("actualizar el esfuerzo: %w", err)
	}
	return updated, nil
}
