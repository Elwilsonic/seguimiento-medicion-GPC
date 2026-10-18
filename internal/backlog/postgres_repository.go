package backlog

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository guarda los ítems en PostgreSQL (TEC-05). Cumple el mismo
// contrato que MemoryRepository. Repository no recibe un contexto, así que
// cada consulta usa context.Background().
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository crea un repositorio sobre pool. El pool lo cierra
// quien lo abrió.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// itemColumns son las columnas de un ítem, en el orden de scanItem.
const itemColumns = "id, project_id, title, description, priority, acceptance_criteria, status, story_points, estimated_hours"

// Save guarda el ítem; la base asigna el siguiente ID correlativo global
// (RN-04.6). Devuelve el ítem tal como quedó guardado, con una lista de
// criterios propia (RN-04.9).
func (r *PostgresRepository) Save(item Item) (Item, error) {
	row := r.pool.QueryRow(context.Background(),
		`INSERT INTO backlog_items (project_id, title, description, priority, acceptance_criteria, status, story_points, estimated_hours)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING `+itemColumns,
		item.ProjectID, item.Title, item.Description, item.Priority, criteriaParam(item.AcceptanceCriteria),
		item.Status, item.StoryPoints, item.EstimatedHours)
	saved, err := scanItem(row)
	if err != nil {
		return Item{}, fmt.Errorf("guardar el ítem: %w", err)
	}
	return saved, nil
}

// GetByID devuelve el ítem con ese ID (RN-05.1), o ErrItemNotFound si no
// existe (RN-05.2).
func (r *PostgresRepository) GetByID(id int) (Item, error) {
	row := r.pool.QueryRow(context.Background(), "SELECT "+itemColumns+" FROM backlog_items WHERE id = $1", id)
	item, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrItemNotFound
	}
	if err != nil {
		return Item{}, fmt.Errorf("buscar el ítem: %w", err)
	}
	return item, nil
}

// Update reemplaza el ítem con ese ID y lo devuelve tal como quedó guardado
// (RN-07.7), o ErrItemNotFound si no existe (RN-07.6).
func (r *PostgresRepository) Update(item Item) (Item, error) {
	row := r.pool.QueryRow(context.Background(),
		`UPDATE backlog_items SET project_id = $2, title = $3, description = $4, priority = $5,
		acceptance_criteria = $6, status = $7, story_points = $8, estimated_hours = $9
		WHERE id = $1 RETURNING `+itemColumns,
		item.ID, item.ProjectID, item.Title, item.Description, item.Priority, criteriaParam(item.AcceptanceCriteria),
		item.Status, item.StoryPoints, item.EstimatedHours)
	updated, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrItemNotFound
	}
	if err != nil {
		return Item{}, fmt.Errorf("modificar el ítem: %w", err)
	}
	return updated, nil
}

// ListByProject devuelve una lista nueva con los ítems del proyecto, en orden
// de ID ascendente (RN-06.6). Solo aparecen los ítems de ese proyecto
// (RN-06.3) y un proyecto sin ítems devuelve una lista vacía, no nil (RN-06.2).
func (r *PostgresRepository) ListByProject(projectID int) ([]Item, error) {
	rows, err := r.pool.Query(context.Background(),
		"SELECT "+itemColumns+" FROM backlog_items WHERE project_id = $1 ORDER BY id", projectID)
	if err != nil {
		return nil, fmt.Errorf("listar los ítems: %w", err)
	}
	items := make([]Item, 0)
	items, err = pgx.AppendRows(items, rows, func(row pgx.CollectableRow) (Item, error) {
		return scanItem(row)
	})
	if err != nil {
		return nil, fmt.Errorf("listar los ítems: %w", err)
	}
	return items, nil
}

// criteriaParam devuelve los criterios para guardar: la columna es NOT NULL,
// así que nil se guarda como una lista vacía.
func criteriaParam(criteria []string) []string {
	if criteria == nil {
		return []string{}
	}
	return criteria
}

// scanItem lee un ítem con las columnas de itemColumns. Cada lectura arma una
// lista de criterios nueva, que no comparte nada con otra (RN-04.9).
func scanItem(row pgx.Row) (Item, error) {
	var item Item
	err := row.Scan(&item.ID, &item.ProjectID, &item.Title, &item.Description, &item.Priority,
		&item.AcceptanceCriteria, &item.Status, &item.StoryPoints, &item.EstimatedHours)
	return item, err
}
