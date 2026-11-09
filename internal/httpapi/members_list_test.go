package httpapi_test

import (
	"net/http"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/member"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-17 · Consultar los integrantes de un proyecto (Issue #31)
//
// GET /api/proyectos/{id}/integrantes lista los integrantes del proyecto,
// ordenados por ID, con el formato de memberResponse (CA-17.1 a CA-17.4). El
// 400 de un ID no numérico está en TestRutasConID_IDNoNumericoDevuelve400.

// HU-17 · CA-17.1 · Verifica que GET devuelva los integrantes del proyecto en
// orden de ID y con el formato exacto del registro, sin los de otro proyecto.
// También cubre CA-17.3.
func TestListarIntegrantes_Devuelve200ConLosIntegrantesOrdenados(t *testing.T) {
	h := newHandlerWithProject(t)
	createProject(t, h, "Inventario")
	for _, req := range []struct{ path, body string }{
		{"/api/proyectos/1/integrantes", `{"nombre":"Ana Pérez","rol":"AGILE_ENABLER"}`},
		{"/api/proyectos/2/integrantes", `{"nombre":"Ana Pérez","rol":"PRODUCT_BUILDER"}`},
		{"/api/proyectos/1/integrantes", `{"nombre":"Juan Gómez","rol":"PRODUCT_BUILDER"}`},
	} {
		expectStatus(t, do(t, h, http.MethodPost, req.path, req.body), http.StatusCreated)
	}

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/integrantes", "")

	expectRawJSON(t, rec, `[
		{"id":1,"proyecto_id":1,"nombre":"Ana Pérez","rol":"AGILE_ENABLER"},
		{"id":3,"proyecto_id":1,"nombre":"Juan Gómez","rol":"PRODUCT_BUILDER"}
	]`)
}

// HU-17 · CA-17.2 · Verifica que un proyecto sin integrantes devuelva 200 y [],
// no null.
func TestListarIntegrantes_SinIntegrantesDevuelveListaVacia(t *testing.T) {
	rec := do(t, newHandlerWithProject(t), http.MethodGet, "/api/proyectos/1/integrantes", "")

	expectRawJSON(t, rec, `[]`)
}

// HU-17 · CA-17.4 · Verifica que un proyecto inexistente, 0 o negativo devuelva
// 404.
func TestListarIntegrantes_ProyectoInexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodGet, "/api/proyectos/"+id+"/integrantes", "")

		expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
	}
}

// HU-17 · RN-17.8 · Verifica que una falla del almacenamiento devuelva 500 sin
// detalle.
func TestListarIntegrantes_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	members := member.NewManager(brokenMemberRepository{}, projects)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, members)
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/integrantes", "")

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
