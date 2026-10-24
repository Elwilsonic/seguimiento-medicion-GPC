// Command server arranca la aplicación (TEC-01): un servidor HTTP con los
// datos en memoria o en PostgreSQL según STORAGE (TEC-05). El puerto se
// configura con la variable de entorno PORT.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
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
	sprints := sprint.NewManager(repos.sprints, projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)

	addr := address(os.Getenv)
	log.Printf("servidor escuchando en %s (almacenamiento: %s)", addr, cfg.kind)
	log.Fatal(newServer(addr, httpapi.NewHandler(projects, items, sprints, efforts)).ListenAndServe())
}

// newServer arma el servidor HTTP con timeouts, para que una conexión lenta o
// colgada no lo bloquee (CA-TEC04.7).
func newServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
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
