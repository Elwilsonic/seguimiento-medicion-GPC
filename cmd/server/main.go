// Command server arranca la aplicación (TEC-01): un servidor HTTP con los
// datos en memoria o en PostgreSQL según STORAGE (TEC-05). El puerto se
// configura con la variable de entorno PORT.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/project"
)

func main() {
	cfg, err := storageConfig(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	repos, closeRepos, err := newRepositories(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer closeRepos()

	projects := project.NewManager(repos.projects)
	items := backlog.NewManager(repos.items, projects)

	addr := address(os.Getenv)
	log.Printf("servidor escuchando en %s (almacenamiento: %s)", addr, cfg.kind)
	log.Fatal(http.ListenAndServe(addr, httpapi.NewHandler(projects, items)))
}

// address devuelve la dirección donde escucha el servidor: el puerto de la
// variable de entorno PORT, o 8080 si no está definida (CA-TEC01.1).
func address(getenv func(string) string) string {
	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
