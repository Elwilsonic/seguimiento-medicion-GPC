// Command server arranca la aplicación (TEC-01): un servidor HTTP con los
// datos en memoria. El puerto se configura con la variable de entorno PORT.
package main

import (
	"log"
	"net/http"
	"os"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/project"
)

func main() {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)

	addr := address(os.Getenv)
	log.Printf("servidor escuchando en %s", addr)
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
