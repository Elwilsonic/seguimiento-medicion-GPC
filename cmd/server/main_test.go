package main

import (
	"errors"
	"strings"
	"testing"
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
