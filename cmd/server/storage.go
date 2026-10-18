package main

import (
	"errors"
	"fmt"
)

// Valores de la variable de entorno STORAGE (CA-TEC05.4).
const (
	storageMemory   = "memory"
	storagePostgres = "postgres"
)

var (
	// errMissingDatabaseURL indica que STORAGE=postgres pero falta DATABASE_URL.
	errMissingDatabaseURL = errors.New("STORAGE=postgres necesita la variable DATABASE_URL")
	// errUnknownStorage indica que STORAGE no es memory ni postgres.
	errUnknownStorage = errors.New("STORAGE debe ser memory o postgres")
)

// storage es la configuración del almacenamiento de la aplicación.
type storage struct {
	kind        string
	databaseURL string
}

// storageConfig lee la configuración del almacenamiento de las variables de
// entorno (CA-TEC05.4): STORAGE vale memory (el valor por defecto) o
// postgres, y con postgres DATABASE_URL es obligatoria. Los valores se
// comparan tal cual. Los errores nunca incluyen DATABASE_URL, que tiene la
// contraseña (CA-TEC05.8).
func storageConfig(getenv func(string) string) (storage, error) {
	switch kind := getenv("STORAGE"); kind {
	case "", storageMemory:
		return storage{kind: storageMemory}, nil
	case storagePostgres:
		url := getenv("DATABASE_URL")
		if url == "" {
			return storage{}, errMissingDatabaseURL
		}
		return storage{kind: storagePostgres, databaseURL: url}, nil
	default:
		return storage{}, fmt.Errorf("%w (se recibió %q)", errUnknownStorage, kind)
	}
}
