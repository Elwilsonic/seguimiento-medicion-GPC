package main

import (
	"context"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/member"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/postgres"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// repositories son los repositorios de la aplicación, en memoria o en
// PostgreSQL según la configuración (CA-TEC05.4).
type repositories struct {
	projects project.Repository
	items    backlog.Repository
	sprints  sprint.Repository
	efforts  effort.Repository
	poker    poker.Repository
	defects  defect.Repository
	members  member.Repository
}

// newRepositories arma los repositorios según cfg y devuelve la función que
// libera la conexión. Con memoria no necesita ninguna base. Con PostgreSQL
// abre un solo pool, compartido por todos los repositorios; si la base no
// responde, devuelve el error de postgres.Open, que no incluye la URL.
func newRepositories(ctx context.Context, cfg storage) (repositories, func(), error) {
	if cfg.kind == storagePostgres {
		pool, err := postgres.Open(ctx, cfg.databaseURL)
		if err != nil {
			return repositories{}, nil, err
		}
		return repositories{
			projects: project.NewPostgresRepository(pool),
			items:    backlog.NewPostgresRepository(pool),
			sprints:  sprint.NewPostgresRepository(pool),
			efforts:  effort.NewPostgresRepository(pool),
			poker:    poker.NewPostgresRepository(pool),
			defects:  defect.NewPostgresRepository(pool),
			members:  member.NewPostgresRepository(pool),
		}, pool.Close, nil
	}
	return repositories{
		projects: project.NewMemoryRepository(),
		items:    backlog.NewMemoryRepository(),
		sprints:  sprint.NewMemoryRepository(),
		efforts:  effort.NewMemoryRepository(),
		poker:    poker.NewMemoryRepository(),
		defects:  defect.NewMemoryRepository(),
		members:  member.NewMemoryRepository(),
	}, func() {}, nil
}
