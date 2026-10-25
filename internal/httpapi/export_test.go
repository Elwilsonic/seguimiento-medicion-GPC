package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// NewHandlerWithDefects permite a los tests de HU-29 inyectar el Manager de
// defectos sin cablear todavía la dependencia en NewHandler ni en cmd/server.
func NewHandlerWithDefects(projects *project.Manager, items *backlog.Manager, sprints *sprint.Manager, efforts *effort.Manager, pokerManager *poker.Manager, defects *defect.Manager) http.Handler {
	return newHandler(projects, items, sprints, efforts, pokerManager, defects)
}
