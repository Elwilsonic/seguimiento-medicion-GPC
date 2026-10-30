package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/member"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/postgrestest"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// CA-TEC01.1 · El puerto se configura con la variable de entorno PORT.
func TestAddress_UsaLaVariablePORT(t *testing.T) {
	getenv := func(key string) string {
		if key == "PORT" {
			return "9090"
		}
		return ""
	}

	if got := address(getenv); got != ":9090" {
		t.Errorf("address = %q, se esperaba %q", got, ":9090")
	}
}

func TestAddress_SinPORTUsa8080(t *testing.T) {
	getenv := func(string) string { return "" }

	if got := address(getenv); got != ":8080" {
		t.Errorf("address = %q, se esperaba %q", got, ":8080")
	}
}

// TEC-05 · Persistencia con PostgreSQL en Docker (Issue #35)

// fakeEnv devuelve una función getenv que lee las variables de vars; las que
// no están valen "", como con os.Getenv.
func fakeEnv(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// testDatabaseURL es una URL de ejemplo con una contraseña que los errores
// nunca deben mostrar.
const (
	testPassword    = "s3cr3t-no-mostrar"
	testDatabaseURL = "postgres://gpc:" + testPassword + "@localhost:5433/gpc?sslmode=disable"
)

// CA-TEC05.4 · Sin STORAGE, la aplicación usa memoria (el valor por defecto).
func TestStorageConfig_SinSTORAGEUsaMemoria(t *testing.T) {
	cfg, err := storageConfig(fakeEnv(nil))
	if err != nil {
		t.Fatalf("storageConfig devolvió un error inesperado: %v", err)
	}
	if cfg.kind != storageMemory {
		t.Errorf("kind = %q, se esperaba %q", cfg.kind, storageMemory)
	}
}

// CA-TEC05.4 · STORAGE=memory usa memoria, aunque haya DATABASE_URL.
func TestStorageConfig_MemoryUsaMemoria(t *testing.T) {
	cfg, err := storageConfig(fakeEnv(map[string]string{
		"STORAGE":      "memory",
		"DATABASE_URL": testDatabaseURL,
	}))
	if err != nil {
		t.Fatalf("storageConfig devolvió un error inesperado: %v", err)
	}
	if cfg.kind != storageMemory {
		t.Errorf("kind = %q, se esperaba %q", cfg.kind, storageMemory)
	}
}

// CA-TEC05.4 · STORAGE=postgres usa PostgreSQL con la URL de DATABASE_URL.
func TestStorageConfig_PostgresConDATABASE_URL(t *testing.T) {
	cfg, err := storageConfig(fakeEnv(map[string]string{
		"STORAGE":      "postgres",
		"DATABASE_URL": testDatabaseURL,
	}))
	if err != nil {
		t.Fatalf("storageConfig devolvió un error inesperado: %v", err)
	}
	if cfg.kind != storagePostgres {
		t.Errorf("kind = %q, se esperaba %q", cfg.kind, storagePostgres)
	}
	if cfg.databaseURL != testDatabaseURL {
		t.Errorf("databaseURL = %q, se esperaba %q", cfg.databaseURL, testDatabaseURL)
	}
}

// CA-TEC05.4 · STORAGE=postgres sin DATABASE_URL es un error de configuración.
func TestStorageConfig_PostgresSinDATABASE_URLDaError(t *testing.T) {
	_, err := storageConfig(fakeEnv(map[string]string{"STORAGE": "postgres"}))
	if !errors.Is(err, errMissingDatabaseURL) {
		t.Errorf("err = %v, se esperaba %v", err, errMissingDatabaseURL)
	}
}

// CA-TEC05.4 · Un valor de STORAGE que no es memory ni postgres es un error.
// Los valores se comparan tal cual, sin pasar a minúsculas ni recortar.
func TestStorageConfig_ValorDesconocidoDaError(t *testing.T) {
	// " memory" (con un espacio adelante) es el único caso que detecta un
	// strings.TrimSpace sobre STORAGE: no lo quites sin agregar otro igual.
	for _, value := range []string{"mysql", "Postgres", " memory"} {
		t.Run(value, func(t *testing.T) {
			_, err := storageConfig(fakeEnv(map[string]string{"STORAGE": value}))
			if !errors.Is(err, errUnknownStorage) {
				t.Errorf("err = %v, se esperaba %v", err, errUnknownStorage)
			}
		})
	}
}

// CA-TEC05.8 · Los errores de configuración no muestran la contraseña de
// DATABASE_URL.
func TestStorageConfig_ErroresNoMuestranLaContrasena(t *testing.T) {
	cases := map[string]map[string]string{
		"valor desconocido": {"STORAGE": "mysql", "DATABASE_URL": testDatabaseURL},
		"mayúsculas":        {"STORAGE": "POSTGRES", "DATABASE_URL": testDatabaseURL},
	}
	for name, vars := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := storageConfig(fakeEnv(vars))
			if err == nil {
				t.Fatal("storageConfig no devolvió error")
			}
			if strings.Contains(err.Error(), testPassword) {
				t.Errorf("el error muestra la contraseña: %v", err)
			}
		})
	}
}

// CA-TEC05.4 y CA-TEC05.6 · Con memoria no hace falta ninguna base.
func TestNewRepositories_MemoriaNoNecesitaBase(t *testing.T) {
	repos, closeRepos, err := newRepositories(context.Background(), storage{kind: storageMemory})
	if err != nil {
		t.Fatalf("newRepositories devolvió un error inesperado: %v", err)
	}
	defer closeRepos()

	if _, ok := repos.projects.(*project.MemoryRepository); !ok {
		t.Errorf("projects es %T, se esperaba *project.MemoryRepository", repos.projects)
	}
	if _, ok := repos.items.(*backlog.MemoryRepository); !ok {
		t.Errorf("items es %T, se esperaba *backlog.MemoryRepository", repos.items)
	}
	if _, ok := repos.sprints.(*sprint.MemoryRepository); !ok {
		t.Errorf("sprints es %T, se esperaba *sprint.MemoryRepository", repos.sprints)
	}
	if _, ok := repos.efforts.(*effort.MemoryRepository); !ok {
		t.Errorf("efforts es %T, se esperaba *effort.MemoryRepository", repos.efforts)
	}
	if _, ok := repos.poker.(*poker.MemoryRepository); !ok {
		t.Errorf("poker es %T, se esperaba *poker.MemoryRepository", repos.poker)
	}
	if _, ok := repos.defects.(*defect.MemoryRepository); !ok {
		t.Errorf("defects es %T, se esperaba *defect.MemoryRepository", repos.defects)
	}
	if _, ok := repos.members.(*member.MemoryRepository); !ok {
		t.Errorf("members es %T, se esperaba *member.MemoryRepository", repos.members)
	}
}

// HU-21 · Abrir una sesión de Planning Poker (Issue #39): con
// STORAGE=postgres, las sesiones se guardan en el repositorio PostgreSQL, no
// en memoria. Necesita TEST_DATABASE_URL; sin ella se omite (CA-TEC05.7).
func TestNewRepositories_PostgresUsaElRepositorioDePoker(t *testing.T) {
	postgrestest.Pool(t)
	cfg := storage{kind: storagePostgres, databaseURL: os.Getenv(postgrestest.EnvVar)}

	repos, closeRepos, err := newRepositories(context.Background(), cfg)
	if err != nil {
		t.Fatalf("newRepositories devolvió un error inesperado: %v", err)
	}
	defer closeRepos()

	if _, ok := repos.poker.(*poker.PostgresRepository); !ok {
		t.Errorf("poker es %T, se esperaba *poker.PostgresRepository", repos.poker)
	}
}

// HU-26 · Registrar esfuerzo (Issue #44): con STORAGE=postgres, los esfuerzos
// se guardan en el repositorio PostgreSQL, no en memoria. Necesita
// TEST_DATABASE_URL; sin ella se omite (CA-TEC05.7).
func TestNewRepositories_PostgresUsaElRepositorioDeEsfuerzos(t *testing.T) {
	postgrestest.Pool(t)
	cfg := storage{kind: storagePostgres, databaseURL: os.Getenv(postgrestest.EnvVar)}

	repos, closeRepos, err := newRepositories(context.Background(), cfg)
	if err != nil {
		t.Fatalf("newRepositories devolvió un error inesperado: %v", err)
	}
	defer closeRepos()

	if _, ok := repos.efforts.(*effort.PostgresRepository); !ok {
		t.Errorf("efforts es %T, se esperaba *effort.PostgresRepository", repos.efforts)
	}
}

// HU-29 · Registrar defecto (Issue #47): con STORAGE=postgres, los defectos
// se guardan en el repositorio PostgreSQL, no en memoria. Necesita
// TEST_DATABASE_URL; sin ella se omite (CA-TEC05.7).
func TestNewRepositories_PostgresUsaElRepositorioDeDefectos(t *testing.T) {
	postgrestest.Pool(t)
	cfg := storage{kind: storagePostgres, databaseURL: os.Getenv(postgrestest.EnvVar)}

	repos, closeRepos, err := newRepositories(context.Background(), cfg)
	if err != nil {
		t.Fatalf("newRepositories devolvió un error inesperado: %v", err)
	}
	defer closeRepos()

	if _, ok := repos.defects.(*defect.PostgresRepository); !ok {
		t.Errorf("defects es %T, se esperaba *defect.PostgresRepository", repos.defects)
	}
	if _, ok := repos.members.(*member.PostgresRepository); !ok {
		t.Errorf("members es %T, se esperaba *member.PostgresRepository", repos.members)
	}
}

// CA-TEC05.4 y CA-TEC05.8 · Con STORAGE=postgres y una base que no responde,
// newRepositories devuelve un error que no muestra la contraseña (no
// necesita base: el puerto 1 de 127.0.0.1 rechaza la conexión).
func TestNewRepositories_PostgresBaseNoDisponibleDaError(t *testing.T) {
	cfg := storage{
		kind:        storagePostgres,
		databaseURL: "postgres://gpc:" + testPassword + "@127.0.0.1:1/x_test?sslmode=disable&connect_timeout=2",
	}

	_, _, err := newRepositories(context.Background(), cfg)
	if err == nil {
		t.Fatal("newRepositories no devolvió error con una base que no responde")
	}
	if strings.Contains(err.Error(), testPassword) {
		t.Errorf("el error muestra la contraseña: %v", err)
	}
}

// CA-TEC05.8 · Con una DATABASE_URL mal formada, el error no muestra la
// contraseña.
func TestNewRepositories_PostgresURLInvalidaNoMuestraLaContrasena(t *testing.T) {
	cfg := storage{
		kind:        storagePostgres,
		databaseURL: "postgres://gpc:" + testPassword + "@localhost:puerto/x_test",
	}

	_, _, err := newRepositories(context.Background(), cfg)
	if err == nil {
		t.Fatal("newRepositories no devolvió error con una URL inválida")
	}
	if strings.Contains(err.Error(), testPassword) {
		t.Errorf("el error muestra la contraseña: %v", err)
	}
}

// CA-TEC05.5 · Con STORAGE=postgres, lo que se guarda sigue ahí después de
// cerrar y volver a abrir la aplicación. Necesita TEST_DATABASE_URL; sin ella
// se omite (CA-TEC05.7).
func TestNewRepositories_PostgresConservaLosDatosAlReabrir(t *testing.T) {
	postgrestest.Pool(t) // saltea o rechaza según TEST_DATABASE_URL y vacía la base
	ctx := context.Background()
	cfg := storage{kind: storagePostgres, databaseURL: os.Getenv(postgrestest.EnvVar)}

	repos, closeRepos, err := newRepositories(ctx, cfg)
	if err != nil {
		t.Fatalf("newRepositories devolvió un error inesperado: %v", err)
	}
	projects := project.NewManager(repos.projects)
	p, err := projects.Create("Persistente", "Sobrevive al reinicio")
	if err != nil {
		t.Fatalf("crear el proyecto: %v", err)
	}
	item, err := backlog.NewManager(repos.items, projects).Create(p.ID, "Login", "", "ALTA", []string{"Valida el usuario"})
	if err != nil {
		t.Fatalf("crear el ítem: %v", err)
	}
	start := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	s, err := sprint.NewManager(repos.sprints, projects, backlog.NewManager(repos.items, projects)).Create(p.ID, "Sprint 1", start, start.AddDate(0, 0, 6), "Persistir los datos")
	if err != nil {
		t.Fatalf("crear el sprint: %v", err)
	}
	closeRepos()

	reopened, closeReopened, err := newRepositories(ctx, cfg)
	if err != nil {
		t.Fatalf("newRepositories al reabrir devolvió un error inesperado: %v", err)
	}
	defer closeReopened()

	if got, err := reopened.projects.GetByID(p.ID); err != nil || got != p {
		t.Errorf("proyecto al reabrir = %+v, %v; se esperaba %+v", got, err, p)
	}
	if got, err := reopened.items.GetByID(item.ID); err != nil || got.Title != item.Title || got.ProjectID != p.ID {
		t.Errorf("ítem al reabrir = %+v, %v; se esperaba %+v", got, err, item)
	}
	if got, exists, err := reopened.sprints.OpenByProject(p.ID); err != nil || !exists || got.ID != s.ID {
		t.Errorf("sprint al reabrir = %+v, %v, %v; se esperaba %+v", got, exists, err, s)
	}
}

// TEC-04 · Interfaz usable (Issue #34)

// CA-TEC04.7 · El servidor tiene timeouts de lectura de encabezados (5 s),
// lectura (10 s), escritura (10 s) e inactividad (60 s).
func TestNewServer_Timeouts(t *testing.T) {
	handler := http.NewServeMux()

	srv := newServer(":9090", handler)

	if srv.Addr != ":9090" {
		t.Errorf("Addr = %q, se esperaba %q", srv.Addr, ":9090")
	}
	if srv.Handler != handler {
		t.Error("Handler no es el handler recibido")
	}
	timeouts := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"ReadHeaderTimeout", srv.ReadHeaderTimeout, 5 * time.Second},
		{"ReadTimeout", srv.ReadTimeout, 10 * time.Second},
		{"WriteTimeout", srv.WriteTimeout, 10 * time.Second},
		{"IdleTimeout", srv.IdleTimeout, 60 * time.Second},
	}
	for _, tt := range timeouts {
		if tt.got != tt.want {
			t.Errorf("%s = %v, se esperaba %v", tt.name, tt.got, tt.want)
		}
	}
}
