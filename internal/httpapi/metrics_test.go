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

// HU-32 · Calcular Story Points planificados, completados y velocidad (Issue #50)
//
// CA-32.9: las métricas se consultan en GET /api/sprints/{id}/metricas y
// GET /api/proyectos/{id}/metricas, con la velocidad redondeada a un decimal o
// null (RN-32.6 y RN-32.7). Un caso exitoso y uno por código de error
// (criterio de TEC-04). El 400 de {id} no numérico está en
// TestRutasConID_IDNoNumericoDevuelve400.

type sprintMetricsJSON struct {
	SprintID        int `json:"sprint_id"`
	SPPlanificados  int `json:"sp_planificados"`
	SPCompletados   int `json:"sp_completados"`
	ItemsSinEstimar int `json:"items_sin_estimar"`
}

type projectMetricsJSON struct {
	ProyectoID      int      `json:"proyecto_id"`
	SPPlanificados  int      `json:"sp_planificados"`
	SPCompletados   int      `json:"sp_completados"`
	ItemsSinEstimar int      `json:"items_sin_estimar"`
	Velocidad       *float64 `json:"velocidad"`
	SprintsCerrados int      `json:"sprints_cerrados"`
}

// metricsSprint crea por la API un sprint del proyecto 1 con un ítem por cada
// valor de points (0 = sin estimar), en el estado de states, y lo cierra si
// closeIt es true. Devuelve el ID del sprint.
func metricsSprint(t *testing.T, h http.Handler, name string, points []int, states []string, closeIt bool) int {
	t.Helper()
	var sp sprintJSON
	rec := do(t, h, http.MethodPost, "/api/proyectos/1/sprints",
		fmt.Sprintf(`{"nombre":%q,"fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Goal"}`, name))
	expectStatus(t, rec, http.StatusCreated)
	decode(t, rec, &sp)
	for i, p := range points {
		var it itemJSON
		rec := do(t, h, http.MethodPost, "/api/proyectos/1/backlog",
			fmt.Sprintf(`{"titulo":"%s ítem %d","prioridad":"MEDIA","criterios":["Criterio"]}`, name, i+1))
		expectStatus(t, rec, http.StatusCreated)
		decode(t, rec, &it)
		if p > 0 {
			expectStatus(t, do(t, h, http.MethodPatch, fmt.Sprintf("/api/items/%d/story-points", it.ID), fmt.Sprintf(`{"story_points":%d}`, p)), http.StatusOK)
		}
		expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d", sp.ID, it.ID), ""), http.StatusOK)
		if states[i] != "PENDIENTE" {
			expectStatus(t, do(t, h, http.MethodPatch, fmt.Sprintf("/api/items/%d/estado", it.ID), `{"estado":"EN_CURSO"}`), http.StatusOK)
		}
		if states[i] == "TERMINADO" {
			expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d/completar", sp.ID, it.ID), ""), http.StatusOK)
		}
	}
	if closeIt {
		expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/cerrar", sp.ID), ""), http.StatusOK)
	}
	return sp.ID
}

func terminados(n int) []string {
	states := make([]string, n)
	for i := range states {
		states[i] = "TERMINADO"
	}
	return states
}

// CA-32.9 · GET /api/sprints/{id}/metricas con el ejemplo de TEC-03.
func TestMetricasSprint_Devuelve200ConLosStoryPoints(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	id := metricsSprint(t, h, "Sprint 1", []int{5, 3, 2, 0}, []string{"TERMINADO", "TERMINADO", "EN_CURSO", "PENDIENTE"}, false)

	rec := do(t, h, http.MethodGet, fmt.Sprintf("/api/sprints/%d/metricas", id), "")

	expectStatus(t, rec, http.StatusOK)
	var got sprintMetricsJSON
	decode(t, rec, &got)
	want := sprintMetricsJSON{SprintID: id, SPPlanificados: 10, SPCompletados: 8, ItemsSinEstimar: 1}
	if got != want {
		t.Errorf("métricas = %+v, se esperaba %+v", got, want)
	}
}

// CA-32.9 · GET /api/proyectos/{id}/metricas redondea la velocidad a un
// decimal: (10 + 15 + 12) / 3 = 12,333… → 12.3.
func TestMetricasProyecto_Devuelve200ConLaVelocidadRedondeada(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	metricsSprint(t, h, "Sprint 1", []int{5, 5}, terminados(2), true)
	metricsSprint(t, h, "Sprint 2", []int{13, 2}, terminados(2), true)
	metricsSprint(t, h, "Sprint 3", []int{8, 3, 1}, terminados(3), true)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas", "")

	expectStatus(t, rec, http.StatusOK)
	var got projectMetricsJSON
	decode(t, rec, &got)
	if got.ProyectoID != 1 || got.SPPlanificados != 37 || got.SPCompletados != 37 || got.ItemsSinEstimar != 0 || got.SprintsCerrados != 3 {
		t.Errorf("métricas = %+v, se esperaba proyecto 1, 37 planificados, 37 completados, 0 sin estimar y 3 sprints", got)
	}
	if got.Velocidad == nil || *got.Velocidad != 12.3 {
		t.Errorf("velocidad = %v, se esperaba 12.3", got.Velocidad)
	}
}

// CA-32.9 y CA-32.6 · Sin sprints cerrados, la velocidad es null.
func TestMetricasProyecto_SinSprintsCerradosLaVelocidadEsNull(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	metricsSprint(t, h, "Sprint 1", []int{8}, terminados(1), false)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas", "")

	expectStatus(t, rec, http.StatusOK)
	var got projectMetricsJSON
	decode(t, rec, &got)
	if got.Velocidad != nil || got.SprintsCerrados != 0 {
		t.Errorf("velocidad = %v con %d sprints, se esperaba null con 0", got.Velocidad, got.SprintsCerrados)
	}
}

// CA-32.9 · Un sprint inexistente devuelve 404.
func TestMetricasSprint_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/sprints/99/metricas", "")

	expectError(t, rec, http.StatusNotFound, sprint.ErrSprintNotFound.Error())
}

// CA-32.9 · Un proyecto inexistente devuelve 404.
func TestMetricasProyecto_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos/99/metricas", "")

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

// CA-32.9 · Una falla del repositorio devuelve 500 sin exponer el detalle.
func TestMetricas_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(brokenItemRepository{}, projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)

	for _, path := range []string{"/api/sprints/1/metricas", "/api/proyectos/1/metricas"} {
		rec := do(t, h, http.MethodGet, path, "")

		expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
	}
}

// HU-34 · Calcular el porcentaje de historias completadas (Issue #52)
//
// CA-34.6: los endpoints de métricas de HU-32 suman items_totales,
// items_terminados y porcentaje_completadas, con un decimal o null (RN-34.6 y
// RN-34.7). Los errores 400, 404 y 500 de esas rutas ya los cubren los tests
// de HU-32 y TestRutasConID_IDNoNumericoDevuelve400.

type completionJSON struct {
	ItemsTotales          int      `json:"items_totales"`
	ItemsTerminados       int      `json:"items_terminados"`
	PorcentajeCompletadas *float64 `json:"porcentaje_completadas"`
}

func expectCompletionJSON(t *testing.T, got completionJSON, items, finished int, percent float64) {
	t.Helper()
	if got.ItemsTotales != items || got.ItemsTerminados != finished || got.PorcentajeCompletadas == nil || *got.PorcentajeCompletadas != percent {
		t.Errorf("porcentaje = %v con %d de %d, se esperaba %v con %d de %d", got.PorcentajeCompletadas, got.ItemsTerminados, got.ItemsTotales, percent, finished, items)
	}
}

// CA-34.6 · GET /api/sprints/{id}/metricas incluye M8 con el ejemplo de
// TEC-03: 2 de 4 ítems terminados.
func TestMetricasSprint_IncluyeElPorcentajeDeHistoriasCompletadas(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	id := metricsSprint(t, h, "Sprint 1", []int{5, 3, 2, 0}, []string{"TERMINADO", "TERMINADO", "EN_CURSO", "PENDIENTE"}, false)

	rec := do(t, h, http.MethodGet, fmt.Sprintf("/api/sprints/%d/metricas", id), "")

	expectStatus(t, rec, http.StatusOK)
	var got completionJSON
	decode(t, rec, &got)
	expectCompletionJSON(t, got, 4, 2, 50)
}

// CA-34.6 y RN-34.7 · GET /api/proyectos/{id}/metricas redondea M8 a un
// decimal: 1 de 3 = 33,333… → 33.3.
func TestMetricasProyecto_PorcentajeRedondeadoAUnDecimal(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	metricsSprint(t, h, "Sprint 1", []int{1, 1, 1}, []string{"TERMINADO", "PENDIENTE", "EN_CURSO"}, false)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas", "")

	expectStatus(t, rec, http.StatusOK)
	var got completionJSON
	decode(t, rec, &got)
	expectCompletionJSON(t, got, 3, 1, 33.3)
}

// CA-34.6 y RN-34.6 · Sin ítems, el porcentaje es null en los dos niveles.
func TestMetricas_SinItemsElPorcentajeEsNull(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	id := metricsSprint(t, h, "Sprint 1", nil, nil, false)

	for _, path := range []string{fmt.Sprintf("/api/sprints/%d/metricas", id), "/api/proyectos/1/metricas"} {
		rec := do(t, h, http.MethodGet, path, "")

		expectStatus(t, rec, http.StatusOK)
		var got completionJSON
		decode(t, rec, &got)
		if got.PorcentajeCompletadas != nil || got.ItemsTotales != 0 || got.ItemsTerminados != 0 {
			t.Errorf("%s: porcentaje = %v con %d de %d, se esperaba null con 0 de 0", path, got.PorcentajeCompletadas, got.ItemsTerminados, got.ItemsTotales)
		}
		if !strings.Contains(rec.Body.String(), `"porcentaje_completadas":null`) {
			t.Errorf("%s: la respuesta no tiene porcentaje_completadas en null: %s", path, rec.Body.String())
		}
	}
}
