package postgrestest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"seguimiento-medicion-gpc/internal/postgres"
)

// TEC-05 · Persistencia con PostgreSQL en Docker (Issue #35). Ninguno de
// estos tests necesita una base de datos.

// testPassword es una contraseña que los errores y los motivos nunca deben
// mostrar.
const testPassword = "s3cr3t-no-mostrar"

// fakeOpen cuenta las veces que se intenta conectar y devuelve pool y err.
type fakeOpen struct {
	calls int
	pool  *pgxpool.Pool
	err   error
}

func (f *fakeOpen) open(context.Context, string) (*pgxpool.Pool, error) {
	f.calls++
	return f.pool, f.err
}

// CA-TEC05.7 · Se acepta una base cuyo nombre termina en _test, tanto en
// formato URL como key=value.
func TestCheckDatabaseName_AceptaSufijoTest(t *testing.T) {
	for _, url := range []string{
		"postgres://gpc:gpc_dev@localhost:5433/gpc_test?sslmode=disable",
		"host=localhost port=5433 user=gpc dbname=gpc_test",
	} {
		if err := checkDatabaseName(url); err != nil {
			t.Errorf("checkDatabaseName(%q) = %v, se esperaba nil", url, err)
		}
	}
}

// CA-TEC05.7 · Una base que no termina en _test (por ejemplo, la de
// desarrollo) se rechaza.
func TestCheckDatabaseName_RechazaBaseSinSufijo(t *testing.T) {
	err := checkDatabaseName("postgres://gpc:gpc_dev@localhost:5433/gpc?sslmode=disable")
	if !errors.Is(err, errNotTestDatabase) {
		t.Errorf("err = %v, se esperaba %v", err, errNotTestDatabase)
	}
}

// CA-TEC05.7 · El sufijo tiene que estar al final del nombre.
func TestCheckDatabaseName_RechazaSufijoQueNoEstaAlFinal(t *testing.T) {
	err := checkDatabaseName("postgres://gpc:gpc_dev@localhost:5433/gpc_test_viejo?sslmode=disable")
	if !errors.Is(err, errNotTestDatabase) {
		t.Errorf("err = %v, se esperaba %v", err, errNotTestDatabase)
	}
}

// CA-TEC05.7 · Una URL sin base (y sin PGDATABASE) se rechaza: PostgreSQL
// usaría una base con el nombre del usuario.
func TestCheckDatabaseName_RechazaURLSinBase(t *testing.T) {
	t.Setenv("PGDATABASE", "")

	err := checkDatabaseName("postgres://gpc:gpc_dev@localhost:5433")
	if !errors.Is(err, errNotTestDatabase) {
		t.Errorf("err = %v, se esperaba %v", err, errNotTestDatabase)
	}
}

// CA-TEC05.7 · Una URL mal formada se rechaza.
func TestCheckDatabaseName_RechazaURLInvalida(t *testing.T) {
	err := checkDatabaseName("postgres://gpc:gpc_dev@localhost:puerto/gpc_test")
	if !errors.Is(err, errInvalidURL) {
		t.Errorf("err = %v, se esperaba %v", err, errInvalidURL)
	}
}

// CA-TEC05.8 · Los errores de checkDatabaseName no muestran la contraseña.
func TestCheckDatabaseName_ErroresNoMuestranLaContrasena(t *testing.T) {
	for _, url := range []string{
		"postgres://gpc:" + testPassword + "@localhost:5433/gpc",
		"postgres://gpc:" + testPassword + "@localhost:puerto/gpc_test",
	} {
		err := checkDatabaseName(url)
		if err == nil {
			t.Fatalf("checkDatabaseName(%q) no devolvió error", url)
		}
		if strings.Contains(err.Error(), testPassword) {
			t.Errorf("el error muestra la contraseña: %v", err)
		}
	}
}

// CA-TEC05.7 · Sin TEST_DATABASE_URL (como en el CI) se saltea, sin intentar
// conectarse.
func TestDecide_SinURLSaltea(t *testing.T) {
	open := &fakeOpen{}

	d := decide(context.Background(), "", open.open)
	if d.action != actionSkip {
		t.Errorf("action = %v, se esperaba actionSkip", d.action)
	}
	if open.calls != 0 {
		t.Errorf("se intentó conectar %d veces, se esperaba 0", open.calls)
	}
}

// CA-TEC05.7 · Una base sin _test se rechaza antes de intentar conectarse,
// así nunca se vacía una base que no es de pruebas.
func TestDecide_BaseSinSufijoTestRechazaSinConectarse(t *testing.T) {
	open := &fakeOpen{}

	d := decide(context.Background(), "postgres://gpc:gpc_dev@localhost:5433/gpc", open.open)
	if d.action != actionReject {
		t.Errorf("action = %v, se esperaba actionReject", d.action)
	}
	if open.calls != 0 {
		t.Errorf("se intentó conectar %d veces, se esperaba 0", open.calls)
	}
}

// CA-TEC05.7 · Una URL mal formada se rechaza antes de intentar conectarse.
func TestDecide_URLInvalidaRechazaSinConectarse(t *testing.T) {
	open := &fakeOpen{}

	d := decide(context.Background(), "postgres://gpc:gpc_dev@localhost:puerto/gpc_test", open.open)
	if d.action != actionReject {
		t.Errorf("action = %v, se esperaba actionReject", d.action)
	}
	if open.calls != 0 {
		t.Errorf("se intentó conectar %d veces, se esperaba 0", open.calls)
	}
}

// CA-TEC05.7 · Si la base de pruebas no responde, se saltea.
func TestDecide_BaseNoDisponibleSaltea(t *testing.T) {
	open := &fakeOpen{err: errors.New("la base no responde")}

	d := decide(context.Background(), "postgres://gpc:gpc_dev@localhost:5433/gpc_test", open.open)
	if d.action != actionSkip {
		t.Errorf("action = %v, se esperaba actionSkip", d.action)
	}
	if open.calls != 1 {
		t.Errorf("se intentó conectar %d veces, se esperaba 1", open.calls)
	}
}

// CA-TEC05.7 · Con una base de pruebas disponible, se usa el pool abierto.
func TestDecide_BaseDisponibleDevuelveElPool(t *testing.T) {
	pool := &pgxpool.Pool{}
	open := &fakeOpen{pool: pool}

	d := decide(context.Background(), "postgres://gpc:gpc_dev@localhost:5433/gpc_test", open.open)
	if d.action != actionUse {
		t.Errorf("action = %v, se esperaba actionUse", d.action)
	}
	if d.pool != pool {
		t.Error("decide no devolvió el pool abierto")
	}
}

// CA-TEC05.8 · Los motivos de saltear o rechazar no muestran la contraseña.
// La base que no responde usa la conexión real (127.0.0.1:1 la rechaza).
func TestDecide_MotivosNoMuestranLaContrasena(t *testing.T) {
	cases := map[string]string{
		"base sin _test":       "postgres://gpc:" + testPassword + "@localhost:5433/gpc",
		"URL inválida":         "postgres://gpc:" + testPassword + "@localhost:puerto/gpc_test",
		"base que no responde": "postgres://gpc:" + testPassword + "@127.0.0.1:1/x_test?sslmode=disable&connect_timeout=2",
	}
	for name, url := range cases {
		t.Run(name, func(t *testing.T) {
			d := decide(context.Background(), url, postgres.Open)
			if d.action == actionUse {
				d.pool.Close()
				t.Fatal("decide devolvió actionUse, se esperaba saltear o rechazar")
			}
			if strings.Contains(d.reason, testPassword) {
				t.Errorf("el motivo muestra la contraseña: %s", d.reason)
			}
		})
	}
}
