package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-50 · Listar los proyectos (Issue #183): GET /api/proyectos.

// failingListProjectRepository es un repositorio de proyectos en memoria cuya
// lectura de la lista falla.
type failingListProjectRepository struct {
	*project.MemoryRepository
}

func (r failingListProjectRepository) List() ([]project.Project, error) {
	return nil, errStorage
}

// newHandlerWithFailingProjectList arma el handler con el repositorio de
// proyectos que falla al listar.
func newHandlerWithFailingProjectList() http.Handler {
	projects := project.NewManager(failingListProjectRepository{project.NewMemoryRepository()})
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	return httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))
}

// HU-50 · CA-50.1 · Verifica que GET /api/proyectos responda 200 con todos los proyectos, ordenados, con el formato de GET /api/proyectos/{id}.
func TestListarProyectos_DevuelveLosProyectosConElFormatoDeLaConsultaPorID(t *testing.T) {
	h := newHandler()
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"Sistema de ventas","descripcion":"Ventas online"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"Inventario","descripcion":""}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPut, "/api/proyectos/2", `{"nombre":"Inventario","descripcion":"","estado":"EN_CURSO"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/2/fecha-inicio", `{"fecha_inicio":"2026-03-01"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/2/fecha-fin", `{"fecha_fin":"2026-06-30"}`), http.StatusOK)

	rec := do(t, h, http.MethodGet, "/api/proyectos", "")

	expectStatus(t, rec, http.StatusOK)
	want := `[{"id":2,"nombre":"Inventario","descripcion":"","estado":"EN_CURSO","fecha_inicio":"2026-03-01","fecha_fin":"2026-06-30"},` +
		`{"id":1,"nombre":"Sistema de ventas","descripcion":"Ventas online","estado":"PLANIFICADO","fecha_inicio":"","fecha_fin":""}]`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("cuerpo = %s, se esperaba %s", got, want)
	}

	// Cada elemento es idéntico a la respuesta de GET /api/proyectos/{id}.
	var list []json.RawMessage
	decode(t, rec, &list)
	for i, path := range []string{"/api/proyectos/2", "/api/proyectos/1"} {
		one := do(t, h, http.MethodGet, path, "")
		expectStatus(t, one, http.StatusOK)
		if i < len(list) && string(list[i]) != strings.TrimSpace(one.Body.String()) {
			t.Errorf("elemento %d = %s, se esperaba el mismo cuerpo que GET %s: %s", i, list[i], path, one.Body.String())
		}
	}
}

// HU-50 · CA-50.2 · Verifica que sin proyectos GET /api/proyectos responda 200 con el arreglo vacío [], no null.
func TestListarProyectos_SinProyectosRespondeArregloVacio(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos", "")

	expectStatus(t, rec, http.StatusOK)
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("cuerpo = %s, se esperaba []", got)
	}
}

// HU-50 · CA-50.5 · Verifica que, si falla la lectura de los proyectos, GET /api/proyectos responda 500 sin el error interno.
func TestListarProyectos_ErrorDelRepositorioResponde500(t *testing.T) {
	rec := do(t, newHandlerWithFailingProjectList(), http.MethodGet, "/api/proyectos", "")

	expectStatus(t, rec, http.StatusInternalServerError)
	if got, want := strings.TrimSpace(rec.Body.String()), `{"error":"error interno del servidor"}`; got != want {
		t.Errorf("cuerpo = %s, se esperaba %s", got, want)
	}
	if strings.Contains(rec.Body.String(), errStorage.Error()) {
		t.Errorf("la respuesta muestra el error interno: %s", rec.Body.String())
	}
}
