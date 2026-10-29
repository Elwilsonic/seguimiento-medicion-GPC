package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository guarda los proyectos en PostgreSQL (TEC-05). Cumple el
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

// projectColumns son las columnas de un proyecto, en el orden de scanProject.
const projectColumns = "id, name, description, status, start_date, end_date"

// Save guarda el proyecto; la base asigna el siguiente ID correlativo
// (RN-01.5). Devuelve el proyecto tal como quedó guardado.
func (r *PostgresRepository) Save(p Project) (Project, error) {
	row := r.pool.QueryRow(context.Background(),
		"INSERT INTO projects (name, description, status, start_date, end_date) VALUES ($1, $2, $3, $4, $5) RETURNING "+projectColumns,
		p.Name, p.Description, p.Status, dateParam(p.StartDate), dateParam(p.EndDate))
	saved, err := scanProject(row)
	if err != nil {
		return Project{}, fmt.Errorf("guardar el proyecto: %w", err)
	}
	return saved, nil
}

// List devuelve los proyectos guardados en orden de ID.
func (r *PostgresRepository) List() ([]Project, error) {
	rows, err := r.pool.Query(context.Background(), "SELECT "+projectColumns+" FROM projects ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("listar los proyectos: %w", err)
	}
	projects, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Project, error) {
		return scanProject(row)
	})
	if err != nil {
		return nil, fmt.Errorf("listar los proyectos: %w", err)
	}
	return projects, nil
}

// GetByID devuelve el proyecto con ese ID, o ErrProjectNotFound si no
// existe (HU-03).
func (r *PostgresRepository) GetByID(id int) (Project, error) {
	row := r.pool.QueryRow(context.Background(), "SELECT "+projectColumns+" FROM projects WHERE id = $1", id)
	p, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, fmt.Errorf("buscar el proyecto: %w", err)
	}
	return p, nil
}

// Update reemplaza el proyecto con ese ID y lo devuelve tal como quedó
// guardado, o ErrProjectNotFound si no existe (HU-02).
func (r *PostgresRepository) Update(p Project) (Project, error) {
	row := r.pool.QueryRow(context.Background(),
		"UPDATE projects SET name = $2, description = $3, status = $4, start_date = $5, end_date = $6 WHERE id = $1 RETURNING "+projectColumns,
		p.ID, p.Name, p.Description, p.Status, dateParam(p.StartDate), dateParam(p.EndDate))
	updated, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, fmt.Errorf("modificar el proyecto: %w", err)
	}
	return updated, nil
}

// scanProject lee un proyecto con las columnas de projectColumns. Una
// start_date o una end_date NULL es una fecha sin registrar (valor cero,
// RN-18.7 y RN-41.10); una registrada se normaliza a UTC, porque pgx devuelve
// timestamptz en la zona local de la máquina (RN-18.9 y RN-41.12, como los
// sprints).
func scanProject(row pgx.Row) (Project, error) {
	var p Project
	var startDate, endDate *time.Time
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &startDate, &endDate)
	if startDate != nil {
		p.StartDate = startDate.UTC()
	}
	if endDate != nil {
		p.EndDate = endDate.UTC()
	}
	return p, err
}

// dateParam traduce una fecha del proyecto (inicio o fin) a su valor en la
// base: el cero (sin registrar) se guarda como NULL, para no guardar el año 1
// como una fecha real (HU-18 y HU-41).
func dateParam(date time.Time) any {
	if date.IsZero() {
		return nil
	}
	return date
}
