// Package postgrestest prepara la base de las pruebas de integración contra
// PostgreSQL (CA-TEC05.7). Solo lo usan los tests.
package postgrestest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"seguimiento-medicion-gpc/internal/postgres"
)

// EnvVar es la variable de entorno con la URL de la base de pruebas.
const EnvVar = "TEST_DATABASE_URL"

// testSuffix es el sufijo obligatorio del nombre de la base de pruebas.
const testSuffix = "_test"

var (
	// errNotTestDatabase indica que el nombre de la base no termina en _test.
	errNotTestDatabase = errors.New("la base de " + EnvVar + " tiene que terminar en " + testSuffix)
	// errInvalidURL indica que la URL no es una conexión de PostgreSQL válida.
	errInvalidURL = errors.New(EnvVar + " no es una URL de PostgreSQL válida")
)

// action es lo que hace Pool con la base de pruebas.
type action int

const (
	actionUse    action = iota // usar el pool abierto
	actionSkip                 // saltear el test (sin URL o sin base)
	actionReject               // fallar el test (la URL no es de una base de pruebas)
)

// decision es el resultado de decide: la acción, el motivo (para saltear o
// rechazar) y el pool (para usar).
type decision struct {
	action action
	reason string
	pool   *pgxpool.Pool
}

// lockKey identifica el lock de la base de pruebas (pg_advisory_lock).
const lockKey = 3505

// Pool devuelve un pool a la base de pruebas vacía. Sin TEST_DATABASE_URL
// (como en el CI) o si la base no responde, saltea el test; si la base no
// termina en _test, falla sin conectarse. El pool se cierra al terminar el test.
//
// go test corre los paquetes en paralelo y todos comparten la base de
// pruebas: el test toma un lock de la base hasta terminar, así un paquete
// no vacía las tablas mientras otro las usa.
func Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	d := decide(context.Background(), os.Getenv(EnvVar), postgres.Open)
	switch d.action {
	case actionSkip:
		t.Skip(d.reason)
	case actionReject:
		t.Fatal(d.reason)
	}
	t.Cleanup(d.pool.Close)
	lock(t, d.pool)
	Reset(t, d.pool)
	return d.pool
}

// lock toma el lock de la base de pruebas en una conexión propia y lo suelta
// al terminar el test, antes de que se cierre el pool. Si el proceso de
// tests termina antes, PostgreSQL lo suelta al cerrarse la conexión.
func lock(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("reservar la conexión del lock: %v", err)
	}
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", lockKey); err != nil {
		conn.Release()
		t.Fatalf("tomar el lock de la base de pruebas: %v", err)
	}
	t.Cleanup(func() {
		_, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", lockKey)
		conn.Release()
	})
}

// Reset vacía las tablas de la base de pruebas y reinicia los IDs, así cada
// test arranca desde el ID 1.
//
// Usalo solo con el pool que devuelve Pool(t): Pool ya controló que la base
// termine en _test y tomó el lock. Reset no vuelve a controlar nada, así que
// con cualquier otro pool borraría los datos de esa base.
func Reset(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "TRUNCATE defects, poker_sessions, efforts, sprints, backlog_items, projects RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("vaciar la base de pruebas: %v", err)
	}
}

// decide resuelve qué hacer con url, en este orden: sin URL se saltea; si la
// base no termina en _test se rechaza sin intentar conectarse; si open falla
// se saltea; si no, se usa el pool.
func decide(ctx context.Context, url string, open func(context.Context, string) (*pgxpool.Pool, error)) decision {
	if url == "" {
		return decision{action: actionSkip, reason: EnvVar + " no está definida: se omiten las pruebas de PostgreSQL"}
	}
	if err := checkDatabaseName(url); err != nil {
		return decision{action: actionReject, reason: err.Error()}
	}
	pool, err := open(ctx, url)
	if err != nil {
		return decision{action: actionSkip, reason: "la base de pruebas no está disponible: " + err.Error()}
	}
	return decision{action: actionUse, pool: pool}
}

// checkDatabaseName controla que la base a la que se conectaría url termine
// en _test. pgconn.ParseConfig acepta URL y key=value y completa los datos
// que faltan con las variables PG* (por ejemplo, PGDATABASE). Los errores
// no incluyen la URL.
func checkDatabaseName(url string) error {
	cfg, err := pgconn.ParseConfig(url)
	if err != nil {
		return errInvalidURL
	}
	if !strings.HasSuffix(cfg.Database, testSuffix) {
		return fmt.Errorf("%w (base %q)", errNotTestDatabase, cfg.Database)
	}
	return nil
}
