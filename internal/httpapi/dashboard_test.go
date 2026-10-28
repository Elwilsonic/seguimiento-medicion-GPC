package httpapi_test

import (
	"fmt"
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

// HU-36 · Ver el resumen del estado del proyecto (Issue #54)
//
// CA-36.6: el resumen se consulta en GET /api/proyectos/{id}/resumen, con
// sprint_actual en el formato de los sprints (sprintResponse) o null
// (RN-36.3). El 400 de {id} no numérico está en
// TestRutasConID_IDNoNumericoDevuelve400.

type summaryJSON struct {
	ProyectoID      int         `json:"proyecto_id"`
	Nombre          string      `json:"nombre"`
	Estado          string      `json:"estado"`
	ItemsPendientes int         `json:"items_pendientes"`
	ItemsEnCurso    int         `json:"items_en_curso"`
	ItemsTerminados int         `json:"items_terminados"`
	SprintActual    *sprintJSON `json:"sprint_actual"`
}

// summaryItem crea por la API un ítem del proyecto 1, lo asigna al sprint si
// sprintID no es 0 y lo lleva al estado pedido.
func summaryItem(t *testing.T, h http.Handler, title, status string, sprintID int) {
	t.Helper()
	var it itemJSON
	rec := do(t, h, http.MethodPost, "/api/proyectos/1/backlog",
		fmt.Sprintf(`{"titulo":%q,"prioridad":"MEDIA","criterios":["Criterio"]}`, title))
	expectStatus(t, rec, http.StatusCreated)
	decode(t, rec, &it)
	if sprintID != 0 {
		expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d", sprintID, it.ID), ""), http.StatusOK)
	}
	if status != "PENDIENTE" {
		expectStatus(t, do(t, h, http.MethodPatch, fmt.Sprintf("/api/items/%d/estado", it.ID), `{"estado":"EN_CURSO"}`), http.StatusOK)
	}
	if status == "TERMINADO" {
		expectStatus(t, do(t, h, http.MethodPatch, fmt.Sprintf("/api/items/%d/estado", it.ID), `{"estado":"TERMINADO"}`), http.StatusOK)
	}
}

// CA-36.1 y CA-36.6 · GET /api/proyectos/{id}/resumen devuelve el proyecto,
// los ítems por estado y el sprint abierto con los campos de sprintResponse.
func TestResumenProyecto_Devuelve200ConElResumen(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Sistema de ventas","descripcion":"","estado":"EN_CURSO"}`), http.StatusOK)
	rec := do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint)
	expectStatus(t, rec, http.StatusCreated)
	var sp sprintJSON
	decode(t, rec, &sp)
	summaryItem(t, h, "A", "PENDIENTE", sp.ID)
	summaryItem(t, h, "B", "PENDIENTE", 0)
	summaryItem(t, h, "C", "EN_CURSO", sp.ID)
	summaryItem(t, h, "D", "TERMINADO", sp.ID)
	summaryItem(t, h, "E", "TERMINADO", sp.ID)
	summaryItem(t, h, "F", "TERMINADO", 0)

	rec = do(t, h, http.MethodGet, "/api/proyectos/1/resumen", "")

	expectStatus(t, rec, http.StatusOK)
	var got summaryJSON
	decode(t, rec, &got)
	if got.ProyectoID != 1 || got.Nombre != "Sistema de ventas" || got.Estado != "EN_CURSO" {
		t.Errorf("proyecto = %d %q %s, se esperaba 1 \"Sistema de ventas\" EN_CURSO", got.ProyectoID, got.Nombre, got.Estado)
	}
	if got.ItemsPendientes != 2 || got.ItemsEnCurso != 1 || got.ItemsTerminados != 3 {
		t.Errorf("ítems = %d, %d y %d, se esperaba 2, 1 y 3", got.ItemsPendientes, got.ItemsEnCurso, got.ItemsTerminados)
	}
	want := sprintJSON{ID: sp.ID, ProyectoID: 1, Nombre: "Sprint 1", FechaInicio: "2026-10-01", FechaFin: "2026-10-14", Objetivo: "Gestionar el backlog", Estado: "ABIERTO"}
	if got.SprintActual == nil || *got.SprintActual != want {
		t.Errorf("sprint_actual = %+v, se esperaba %+v", got.SprintActual, want)
	}
}

// CA-36.3 y CA-36.6 · Con solo sprints cerrados, sprint_actual es null.
func TestResumenProyecto_SinSprintAbiertoSprintActualEsNull(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/resumen", "")

	expectStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), `"sprint_actual":null`) {
		t.Errorf("la respuesta no tiene sprint_actual en null: %s", rec.Body.String())
	}
}

// CA-36.2 y CA-36.6 · Un proyecto sin ítems ni sprints devuelve las
// cantidades en 0 y sprint_actual en null.
func TestResumenProyecto_ProyectoSinDatosDevuelveCerosYNull(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/resumen", "")

	expectStatus(t, rec, http.StatusOK)
	var got summaryJSON
	decode(t, rec, &got)
	want := summaryJSON{ProyectoID: 1, Nombre: "Sistema de ventas", Estado: "PLANIFICADO"}
	if got != want {
		t.Errorf("resumen = %+v, se esperaba %+v", got, want)
	}
	for _, field := range []string{`"items_pendientes":0`, `"items_en_curso":0`, `"items_terminados":0`, `"sprint_actual":null`} {
		if !strings.Contains(rec.Body.String(), field) {
			t.Errorf("la respuesta no contiene %s: %s", field, rec.Body.String())
		}
	}
}

// CA-36.5 y CA-36.6 · Un proyecto inexistente devuelve 404.
func TestResumenProyecto_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos/99/resumen", "")

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

// CA-36.6 · Una falla del repositorio devuelve 500 sin exponer el detalle.
func TestResumenProyecto_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(brokenItemRepository{}, projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/resumen", "")

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
