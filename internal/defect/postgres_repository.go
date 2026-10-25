package defect

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository guarda los defectos en PostgreSQL (RN-29.10). Cumple el
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

// defectColumns son las columnas de un defecto, en el orden de scanDefect.
const defectColumns = "id, description, severity, status, item_id, detection_sprint_id, resolution_sprint_id"

// Save guarda el defecto; la base asigna el siguiente ID correlativo global
// (RN-29.8). Devuelve el defecto tal como quedó guardado.
func (r *PostgresRepository) Save(defect Defect) (Defect, error) {
	row := r.pool.QueryRow(context.Background(),
		`INSERT INTO defects (description, severity, status, item_id, detection_sprint_id, resolution_sprint_id)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+defectColumns,
		defect.Description, string(defect.Severity), string(defect.Status),
		defect.ItemID, defect.DetectionSprintID, defect.ResolutionSprintID)
	saved, err := scanDefect(row)
	if err != nil {
		return Defect{}, fmt.Errorf("guardar el defecto: %w", err)
	}
	return saved, nil
}

// scanDefect lee un defecto con las columnas de defectColumns.
func scanDefect(row pgx.Row) (Defect, error) {
	var defect Defect
	var severity, status string
	err := row.Scan(&defect.ID, &defect.Description, &severity, &status,
		&defect.ItemID, &defect.DetectionSprintID, &defect.ResolutionSprintID)
	defect.Severity = Severity(severity)
	defect.Status = Status(status)
	return defect, err
}
