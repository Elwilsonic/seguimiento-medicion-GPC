package project

import (
	"context"
	"errors"
	"fmt"

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
const projectColumns = "id, name, description, status"

// Save guarda el proyecto; la base asigna el siguiente ID correlativo
// (RN-01.5). Devuelve el proyecto tal como quedó guardado.
func (r *PostgresRepository) Save(p Project) (Project, error) {
	row := r.pool.QueryRow(context.Background(),
		"INSERT INTO projects (name, description, status) VALUES ($1, $2, $3) RETURNING "+projectColumns,
		p.Name, p.Description, p.Status)
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
		"UPDATE projects SET name = $2, description = $3, status = $4 WHERE id = $1 RETURNING "+projectColumns,
		p.ID, p.Name, p.Description, p.Status)
	updated, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, fmt.Errorf("modificar el proyecto: %w", err)
	}
	return updated, nil
}

// scanProject lee un proyecto con las columnas de projectColumns.
func scanProject(row pgx.Row) (Project, error) {
	var p Project
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Status)
	return p, err
}
