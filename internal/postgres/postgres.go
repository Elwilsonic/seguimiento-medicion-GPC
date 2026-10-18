// Package postgres abre la conexión a PostgreSQL (TEC-05). Los repositorios
// de cada paquete reciben el pool que devuelve Open.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// pingTimeout es el tiempo máximo para verificar que la base responde.
const pingTimeout = 5 * time.Second

// errInvalidURL indica que la URL no es una conexión de PostgreSQL válida.
var errInvalidURL = errors.New("la URL de PostgreSQL no es válida")

// Open abre un pool de conexiones y verifica que la base responda. Los
// errores nunca incluyen la URL, que tiene la contraseña (CA-TEC05.8): el
// error de ParseConfig no se envuelve porque puede repetir la URL.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errInvalidURL
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("crear el pool de PostgreSQL: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectar a PostgreSQL: %w", err)
	}
	return pool, nil
}
