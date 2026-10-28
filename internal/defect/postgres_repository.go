package defect

import (
	"context"
	"errors"
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

// GetByID devuelve el defecto con ese ID o ErrDefectNotFound si no existe,
// incluidos los IDs 0 y negativos (RN-30.3).
func (r *PostgresRepository) GetByID(id int) (Defect, error) {
	row := r.pool.QueryRow(context.Background(),
		"SELECT "+defectColumns+" FROM defects WHERE id = $1", id)
	defect, err := scanDefect(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Defect{}, ErrDefectNotFound
	}
	if err != nil {
		return Defect{}, fmt.Errorf("buscar el defecto: %w", err)
	}
	return defect, nil
}

// Update reemplaza el defecto con ese ID y lo devuelve tal como quedó
// guardado (RN-30.9). Si no existe, devuelve ErrDefectNotFound sin crear ni
// modificar nada.
func (r *PostgresRepository) Update(defect Defect) (Defect, error) {
	row := r.pool.QueryRow(context.Background(),
		`UPDATE defects
		SET description = $2, severity = $3, status = $4, item_id = $5, detection_sprint_id = $6, resolution_sprint_id = $7
		WHERE id = $1 RETURNING `+defectColumns,
		defect.ID, defect.Description, string(defect.Severity), string(defect.Status),
		defect.ItemID, defect.DetectionSprintID, defect.ResolutionSprintID)
	updated, err := scanDefect(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Defect{}, ErrDefectNotFound
	}
	if err != nil {
		return Defect{}, fmt.Errorf("actualizar el defecto: %w", err)
	}
	return updated, nil
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
