package postgres

import (
	"context"
	"strings"
	"testing"
)

// TEC-05 · Persistencia con PostgreSQL en Docker (Issue #35)

// testPassword es una contraseña que los errores nunca deben mostrar.
const testPassword = "s3cr3t-no-mostrar"

// CA-TEC05.4 · Si la base no responde, Open devuelve un error (no necesita
// base: el puerto 1 de 127.0.0.1 rechaza la conexión).
func TestOpen_BaseNoDisponibleDaError(t *testing.T) {
	url := "postgres://gpc:" + testPassword + "@127.0.0.1:1/x_test?sslmode=disable&connect_timeout=2"

	pool, err := Open(context.Background(), url)
	if err == nil {
		pool.Close()
		t.Fatal("Open no devolvió error con una base que no responde")
	}
	if strings.Contains(err.Error(), testPassword) {
		t.Errorf("el error muestra la contraseña: %v", err)
	}
}

// CA-TEC05.8 · Una URL mal formada da error y el error no muestra la contraseña.
func TestOpen_URLInvalidaDaErrorSinMostrarLaContrasena(t *testing.T) {
	url := "postgres://gpc:" + testPassword + "@localhost:puerto/x_test"

	pool, err := Open(context.Background(), url)
	if err == nil {
		pool.Close()
		t.Fatal("Open no devolvió error con una URL inválida")
	}
	if strings.Contains(err.Error(), testPassword) {
		t.Errorf("el error muestra la contraseña: %v", err)
	}
}
