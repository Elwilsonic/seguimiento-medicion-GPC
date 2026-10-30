package httpapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	// HU-33: M4 a M7 y la cantidad de ítems sin horas estimadas.
	HorasEstimadas         float64  `json:"horas_estimadas"`
	HorasReales            float64  `json:"horas_reales"`
	DesviacionHoras        *float64 `json:"desviacion_horas"`
	DesviacionPorcentual   *float64 `json:"desviacion_porcentual"`
	ItemsSinHorasEstimadas int      `json:"items_sin_horas_estimadas"`
}

type projectMetricsJSON struct {
	ProyectoID      int      `json:"proyecto_id"`
	SPPlanificados  int      `json:"sp_planificados"`
	SPCompletados   int      `json:"sp_completados"`
	ItemsSinEstimar int      `json:"items_sin_estimar"`
	Velocidad       *float64 `json:"velocidad"`
	SprintsCerrados int      `json:"sprints_cerrados"`
	// HU-33: M4 a M7 y la cantidad de ítems sin horas estimadas.
	HorasEstimadas         float64  `json:"horas_estimadas"`
	HorasReales            float64  `json:"horas_reales"`
	DesviacionHoras        *float64 `json:"desviacion_horas"`
	DesviacionPorcentual   *float64 `json:"desviacion_porcentual"`
	ItemsSinHorasEstimadas int      `json:"items_sin_horas_estimadas"`
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
	want := sprintMetricsJSON{SprintID: id, SPPlanificados: 10, SPCompletados: 8, ItemsSinEstimar: 1, ItemsSinHorasEstimadas: 4}
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
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))
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

// HU-33 · Calcular horas estimadas, horas reales y desviación (Issue #51)
//
// CA-33.5, CA-33.6, CA-33.9 y CA-33.10: los dos endpoints de métricas suman
// M4 a M7 y la cantidad de ítems sin horas estimadas, con un decimal y null
// cuando la desviación no está disponible. Los errores 400, 404 y 500 de las
// rutas ya los cubren los tests de HU-32 y HU-33.

type hoursItemSpec struct {
	title  string
	hours  float64
	effort float64
}

// createHoursItem crea un ítem en el proyecto 1 y le asigna las horas
// estimadas con PUT (el POST no las toma).
func createHoursItem(t *testing.T, h http.Handler, title string, hours float64) itemJSON {
	t.Helper()
	var it itemJSON
	rec := do(t, h, http.MethodPost, "/api/proyectos/1/backlog",
		fmt.Sprintf(`{"titulo":%q,"descripcion":"","prioridad":"MEDIA","criterios":["Criterio"]}`, title))
	expectStatus(t, rec, http.StatusCreated)
	decode(t, rec, &it)
	rec = do(t, h, http.MethodPut, fmt.Sprintf("/api/items/%d", it.ID),
		fmt.Sprintf(`{"titulo":%q,"descripcion":"","prioridad":"MEDIA","criterios":["Criterio"],"horas_estimadas":%v}`, title, hours))
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &it)
	return it
}

// hoursSprint crea un sprint del proyecto 1 con los ítems de specs, los asigna
// y registra el esfuerzo indicado (0 = sin esfuerzo). Devuelve el ID del sprint.
func hoursSprint(t *testing.T, h http.Handler, name string, specs []hoursItemSpec) int {
	t.Helper()
	var sp sprintJSON
	rec := do(t, h, http.MethodPost, "/api/proyectos/1/sprints",
		fmt.Sprintf(`{"nombre":%q,"fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Goal"}`, name))
	expectStatus(t, rec, http.StatusCreated)
	decode(t, rec, &sp)
	for _, spec := range specs {
		it := createHoursItem(t, h, spec.title, spec.hours)
		expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d", sp.ID, it.ID), ""), http.StatusOK)
		if spec.effort > 0 {
			body := fmt.Sprintf(`{"integrante":"García Santiago","fecha":"2026-10-10","actividad":"Implementación","horas":%v}`, spec.effort)
			expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d/esfuerzos", sp.ID, it.ID), body), http.StatusCreated)
		}
	}
	return sp.ID
}

func expectHour(t *testing.T, name string, got, want float64) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, se esperaba %v", name, got, want)
	}
}

func expectOptionalHour(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s = %v, se esperaba %v", name, got, want)
	}
}

// CA-33.5 · GET /api/sprints/{id}/metricas incluye M4 a M7 con el ejemplo de
// TEC-03: 20 h estimadas, 20 h reales, desviación 0 y 1 ítem sin horas.
func TestMetricasSprint_IncluyeHorasYDesviacion(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	id := hoursSprint(t, h, "Sprint 1", []hoursItemSpec{
		{"A", 10, 12},
		{"B", 6, 5},
		{"C", 0, 3},
		{"D", 4, 0},
	})

	rec := do(t, h, http.MethodGet, fmt.Sprintf("/api/sprints/%d/metricas", id), "")

	expectStatus(t, rec, http.StatusOK)
	var got sprintMetricsJSON
	decode(t, rec, &got)
	expectHour(t, "horas_estimadas", got.HorasEstimadas, 20)
	expectHour(t, "horas_reales", got.HorasReales, 20)
	expectOptionalHour(t, "desviacion_horas", got.DesviacionHoras, 0)
	expectOptionalHour(t, "desviacion_porcentual", got.DesviacionPorcentual, 0)
	if got.ItemsSinHorasEstimadas != 1 {
		t.Errorf("items_sin_horas_estimadas = %d, se esperaba 1", got.ItemsSinHorasEstimadas)
	}
}

// CA-33.6 · GET /api/proyectos/{id}/metricas suma todo el backlog y los
// esfuerzos de todos los sprints.
func TestMetricasProyecto_IncluyeTodoElBacklogYLosEsfuerzos(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	hoursSprint(t, h, "Sprint 1", []hoursItemSpec{
		{"A", 10, 12},
		{"B", 6, 5},
	})
	createHoursItem(t, h, "E", 4)
	createHoursItem(t, h, "F", 0)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas", "")

	expectStatus(t, rec, http.StatusOK)
	var got projectMetricsJSON
	decode(t, rec, &got)
	expectHour(t, "horas_estimadas", got.HorasEstimadas, 20)
	expectHour(t, "horas_reales", got.HorasReales, 17)
	expectOptionalHour(t, "desviacion_horas", got.DesviacionHoras, -3)
	expectOptionalHour(t, "desviacion_porcentual", got.DesviacionPorcentual, -15)
	if got.ItemsSinHorasEstimadas != 1 {
		t.Errorf("items_sin_horas_estimadas = %d, se esperaba 1", got.ItemsSinHorasEstimadas)
	}
}

// CA-33.4 y CA-33.5 · Sin datos, la desviación es null en los dos endpoints.
func TestMetricas_SinDatosLaDesviacionEsNull(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	id := hoursSprint(t, h, "Sprint 1", nil)

	for _, path := range []string{fmt.Sprintf("/api/sprints/%d/metricas", id), "/api/proyectos/1/metricas"} {
		rec := do(t, h, http.MethodGet, path, "")

		expectStatus(t, rec, http.StatusOK)
		body := rec.Body.String()
		for _, want := range []string{`"desviacion_horas":null`, `"desviacion_porcentual":null`} {
			if !strings.Contains(body, want) {
				t.Errorf("%s: el cuerpo no contiene %s: %s", path, want, body)
			}
		}
		var got sprintMetricsJSON
		decode(t, rec, &got)
		if got.DesviacionHoras != nil || got.DesviacionPorcentual != nil || got.HorasEstimadas != 0 || got.HorasReales != 0 || got.ItemsSinHorasEstimadas != 0 {
			t.Errorf("%s: métricas = %+v, se esperaba todo en cero y null", path, got)
		}
	}
}

// CA-33.9 · La API redondea M4, M5, M6 y M7 a un decimal: 21,25 h reales
// sobre 20 h estimadas dan 1,25 h y 6,25 %, que se informan como 1.3 y 6.3.
func TestMetricas_RedondeaHorasYDesviacionAUnDecimal(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	id := hoursSprint(t, h, "Sprint 1", []hoursItemSpec{{"A", 20, 21.25}})

	rec := do(t, h, http.MethodGet, fmt.Sprintf("/api/sprints/%d/metricas", id), "")

	expectStatus(t, rec, http.StatusOK)
	var got sprintMetricsJSON
	decode(t, rec, &got)
	expectHour(t, "horas_reales", got.HorasReales, 21.3)
	expectOptionalHour(t, "desviacion_horas", got.DesviacionHoras, 1.3)
	expectOptionalHour(t, "desviacion_porcentual", got.DesviacionPorcentual, 6.3)
}

// CA-33.9 · El cero negativo de M6 y M7 se normaliza a 0 en el JSON.
func TestMetricas_NormalizaElCeroNegativo(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	id := hoursSprint(t, h, "Sprint 1", []hoursItemSpec{
		{"A", 0.1, 0.3},
		{"B", 0.2, 0},
	})

	rec := do(t, h, http.MethodGet, fmt.Sprintf("/api/sprints/%d/metricas", id), "")

	expectStatus(t, rec, http.StatusOK)
	var got struct {
		HorasEstimadas       float64         `json:"horas_estimadas"`
		HorasReales          float64         `json:"horas_reales"`
		DesviacionHoras      json.RawMessage `json:"desviacion_horas"`
		DesviacionPorcentual json.RawMessage `json:"desviacion_porcentual"`
	}
	decode(t, rec, &got)
	expectHour(t, "horas_estimadas", got.HorasEstimadas, 0.3)
	expectHour(t, "horas_reales", got.HorasReales, 0.3)
	if string(got.DesviacionHoras) != "0" {
		t.Errorf("desviacion_horas = %s, se esperaba 0 sin signo", got.DesviacionHoras)
	}
	if string(got.DesviacionPorcentual) != "0" {
		t.Errorf("desviacion_porcentual = %s, se esperaba 0 sin signo", got.DesviacionPorcentual)
	}
}

// failingEffortListRepository falla al leer los esfuerzos de un sprint.
type failingEffortListRepository struct {
	effort.Repository
	err error
}

func (r failingEffortListRepository) ListBySprint(int) ([]effort.Effort, error) {
	return nil, r.err
}

// CA-33.8 · Una falla del repositorio de esfuerzos devuelve 500 sin detalle.
func TestMetricas_ErrorDelRepositorioDeEsfuerzosDevuelve500(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(failingEffortListRepository{Repository: effort.NewMemoryRepository(), err: errors.New("base caída")}, sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	it := createHoursItem(t, h, "A", 10)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/items/"+strconv.Itoa(it.ID), ""), http.StatusOK)

	for _, path := range []string{"/api/sprints/1/metricas", "/api/proyectos/1/metricas"} {
		rec := do(t, h, http.MethodGet, path, "")

		expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
	}
}

// HU-47 · Calcular los defectos detectados y resueltos (Issue #128)
//
// CA-47.6: los endpoints de métricas suman defectos_detectados y
// defectos_resueltos, como enteros (RN-47.8). Los errores 400 y 404 de esas
// rutas ya los cubren los tests de HU-32 y
// TestRutasConID_IDNoNumericoDevuelve400.

type defectsMetricsJSON struct {
	DefectosDetectados *int `json:"defectos_detectados"`
	DefectosResueltos  *int `json:"defectos_resueltos"`
}

func expectDefectsJSON(t *testing.T, rec *httptest.ResponseRecorder, detected, resolved int) {
	t.Helper()
	expectStatus(t, rec, http.StatusOK)
	var got defectsMetricsJSON
	decode(t, rec, &got)
	if got.DefectosDetectados == nil || *got.DefectosDetectados != detected {
		t.Errorf("defectos_detectados = %v, se esperaba %d; cuerpo: %s", got.DefectosDetectados, detected, rec.Body.String())
	}
	if got.DefectosResueltos == nil || *got.DefectosResueltos != resolved {
		t.Errorf("defectos_resueltos = %v, se esperaba %d; cuerpo: %s", got.DefectosResueltos, resolved, rec.Body.String())
	}
}

// CA-47.6 · Con un defecto resuelto en el sprint 1 y otro abierto, las
// métricas del sprint y del proyecto dan 2 detectados y 1 resuelto.
func TestMetricas_IncluyenLosDefectosDetectadosYResueltos(t *testing.T) {
	h := newHandlerWithDefects(t)
	expectStatus(t, do(t, h, http.MethodPost, defectsPath, validDefect), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, defectsPath, validDefect), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/1/estado", `{"estado":"EN_CORRECCION"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/1/estado", `{"estado":"RESUELTO","sprint_resolucion_id":1}`), http.StatusOK)

	expectDefectsJSON(t, do(t, h, http.MethodGet, "/api/sprints/1/metricas", ""), 2, 1)
	expectDefectsJSON(t, do(t, h, http.MethodGet, "/api/proyectos/1/metricas", ""), 2, 1)
}

// CA-47.3 y CA-47.6 · Sin defectos, los dos campos valen 0 (no null).
func TestMetricas_SinDefectosLosCamposValenCero(t *testing.T) {
	h := newHandlerWithDefects(t)

	expectDefectsJSON(t, do(t, h, http.MethodGet, "/api/sprints/1/metricas", ""), 0, 0)
	expectDefectsJSON(t, do(t, h, http.MethodGet, "/api/proyectos/1/metricas", ""), 0, 0)
}

// Condiciones de error de la SDD · Una falla al leer los defectos devuelve 500
// sin exponer el detalle.
func TestMetricas_ErrorDelRepositorioDeDefectosDevuelve500(t *testing.T) {
	h := newHandlerWithDefectRepository(t, brokenDefectRepository{})

	for _, path := range []string{"/api/sprints/1/metricas", "/api/proyectos/1/metricas"} {
		rec := do(t, h, http.MethodGet, path, "")

		expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
	}
}

// HU-35 · Consultar las métricas del proyecto (Issue #53)
//
// CA-35.6: las rutas, los campos, su orden y los códigos HTTP no cambian, y
// el contrato queda documentado con los tres ejemplos de la SDD. El orden de
// los campos se verifica comparando el cuerpo crudo con el JSON esperado,
// ambos normalizados con json.Compact, porque al decodificar a un struct o a
// un mapa se pierde el orden. CA-35.4: los ítems del backlog se leen una sola
// vez por consulta.

// contractItemSpec es un ítem del ejemplo de TEC-03 con Story Points, horas,
// estado y esfuerzo (0 = sin esfuerzo).
type contractItemSpec struct {
	title  string
	points int
	hours  float64
	status string
	effort float64
}

// contractTEC03 crea por la API el proyecto 1 con el sprint 1 abierto, los
// ítems del ejemplo de TEC-03 y un defecto detectado y resuelto en el sprint.
// Devuelve el ID del sprint.
func contractTEC03(t *testing.T, h http.Handler) int {
	t.Helper()
	createProject(t, h, "Sistema de ventas")
	var sp sprintJSON
	rec := do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint)
	expectStatus(t, rec, http.StatusCreated)
	decode(t, rec, &sp)
	specs := []contractItemSpec{
		{"A", 5, 10, "TERMINADO", 12},
		{"B", 3, 6, "TERMINADO", 5},
		{"C", 2, 0, "EN_CURSO", 3},
		{"D", 0, 4, "PENDIENTE", 0},
	}
	var first itemJSON
	for i, spec := range specs {
		it := createHoursItem(t, h, spec.title, spec.hours)
		if i == 0 {
			first = it
		}
		if spec.points > 0 {
			expectStatus(t, do(t, h, http.MethodPatch, fmt.Sprintf("/api/items/%d/story-points", it.ID), fmt.Sprintf(`{"story_points":%d}`, spec.points)), http.StatusOK)
		}
		expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d", sp.ID, it.ID), ""), http.StatusOK)
		if spec.effort > 0 {
			body := fmt.Sprintf(`{"integrante":"García Santiago","fecha":"2026-10-10","actividad":"Implementación","horas":%v}`, spec.effort)
			expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d/esfuerzos", sp.ID, it.ID), body), http.StatusCreated)
		}
		if spec.status != "PENDIENTE" {
			expectStatus(t, do(t, h, http.MethodPatch, fmt.Sprintf("/api/items/%d/estado", it.ID), `{"estado":"EN_CURSO"}`), http.StatusOK)
		}
		if spec.status == "TERMINADO" {
			expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d/completar", sp.ID, it.ID), ""), http.StatusOK)
		}
	}
	defectBody := fmt.Sprintf(`{"descripcion":"El total no descuenta el descuento","severidad":"alta","sprint_deteccion_id":%d}`, sp.ID)
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/items/%d/defectos", first.ID), defectBody), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/1/estado", `{"estado":"EN_CORRECCION"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/1/estado", fmt.Sprintf(`{"estado":"RESUELTO","sprint_resolucion_id":%d}`, sp.ID)), http.StatusOK)
	return sp.ID
}

// expectRawJSON compara el cuerpo crudo de la respuesta con el JSON
// esperado, ambos con json.Compact, así se verifican también el orden de los
// campos y los null.
func expectRawJSON(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	expectStatus(t, rec, http.StatusOK)
	var got, expected bytes.Buffer
	if err := json.Compact(&got, rec.Body.Bytes()); err != nil {
		t.Fatalf("cuerpo no es JSON válido (%v): %s", err, rec.Body.String())
	}
	if err := json.Compact(&expected, []byte(want)); err != nil {
		t.Fatalf("JSON esperado inválido: %v", err)
	}
	if !bytes.Equal(got.Bytes(), expected.Bytes()) {
		t.Errorf("cuerpo =\n%s\nse esperaba\n%s", got.String(), expected.String())
	}
}

// CA-35.1 y CA-35.6 · Contrato del sprint con el ejemplo de TEC-03.
func TestConsultaMetricas_ContratoDelSprint(t *testing.T) {
	h := newHandler()
	id := contractTEC03(t, h)

	rec := do(t, h, http.MethodGet, fmt.Sprintf("/api/sprints/%d/metricas", id), "")

	expectRawJSON(t, rec, `{
  "sprint_id": 1,
  "sp_planificados": 10,
  "sp_completados": 8,
  "items_sin_estimar": 1,
  "items_totales": 4,
  "items_terminados": 2,
  "porcentaje_completadas": 50,
  "horas_estimadas": 20,
  "horas_reales": 20,
  "desviacion_horas": 0,
  "desviacion_porcentual": 0,
  "items_sin_horas_estimadas": 1,
  "defectos_detectados": 1,
  "defectos_resueltos": 1
}`)
}

// CA-35.1 y CA-35.6 · Contrato del proyecto con ese sprint cerrado.
func TestConsultaMetricas_ContratoDelProyecto(t *testing.T) {
	h := newHandler()
	id := contractTEC03(t, h)
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/cerrar", id), ""), http.StatusOK)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas", "")

	expectRawJSON(t, rec, `{
  "proyecto_id": 1,
  "sp_planificados": 10,
  "sp_completados": 8,
  "items_sin_estimar": 1,
  "velocidad": 8,
  "sprints_cerrados": 1,
  "items_totales": 4,
  "items_terminados": 2,
  "porcentaje_completadas": 50,
  "horas_estimadas": 20,
  "horas_reales": 20,
  "desviacion_horas": 0,
  "desviacion_porcentual": 0,
  "items_sin_horas_estimadas": 1,
  "defectos_detectados": 1,
  "defectos_resueltos": 1
}`)
}

// CA-35.2 y CA-35.6 · Contrato de un proyecto recién creado: las cuatro
// métricas no disponibles en null y el resto en 0.
func TestConsultaMetricas_ContratoDelProyectoVacio(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	createProject(t, h, "Proyecto vacío")

	rec := do(t, h, http.MethodGet, "/api/proyectos/2/metricas", "")

	expectRawJSON(t, rec, `{
  "proyecto_id": 2,
  "sp_planificados": 0,
  "sp_completados": 0,
  "items_sin_estimar": 0,
  "velocidad": null,
  "sprints_cerrados": 0,
  "items_totales": 0,
  "items_terminados": 0,
  "porcentaje_completadas": null,
  "horas_estimadas": 0,
  "horas_reales": 0,
  "desviacion_horas": null,
  "desviacion_porcentual": null,
  "items_sin_horas_estimadas": 0,
  "defectos_detectados": 0,
  "defectos_resueltos": 0
}`)
}

// CA-35.3 · Un {id} no numérico devuelve 400 en las dos rutas.
func TestConsultaMetricas_IDNoNumericoDevuelve400(t *testing.T) {
	h := newHandler()
	for _, path := range []string{"/api/sprints/abc/metricas", "/api/proyectos/abc/metricas"} {
		rec := do(t, h, http.MethodGet, path, "")

		expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
	}
}

// CA-35.3 · Un sprint inexistente, 0 o negativo devuelve 404.
func TestConsultaMetricas_SprintInexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodGet, "/api/sprints/"+id+"/metricas", "")

		expectError(t, rec, http.StatusNotFound, sprint.ErrSprintNotFound.Error())
	}
}

// CA-35.3 · Un proyecto inexistente, 0 o negativo devuelve 404.
func TestConsultaMetricas_ProyectoInexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodGet, "/api/proyectos/"+id+"/metricas", "")

		expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
	}
}

// CA-35.3 · Una falla al leer los esfuerzos o los defectos devuelve 500 sin
// detalle, en las dos rutas. Usa los repositorios de prueba de HU-33
// (failingEffortListRepository) y de HU-31 (brokenDefectRepository).
func TestConsultaMetricas_FallaDeEsfuerzosODefectosDevuelve500(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(failingEffortListRepository{Repository: effort.NewMemoryRepository(), err: errors.New("base caída")}, sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	effortsBroken := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))
	createProject(t, effortsBroken, "Sistema de ventas")
	expectStatus(t, do(t, effortsBroken, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	defectsBroken := newHandlerWithDefectRepository(t, brokenDefectRepository{})

	for _, h := range []http.Handler{effortsBroken, defectsBroken} {
		for _, path := range []string{"/api/sprints/1/metricas", "/api/proyectos/1/metricas"} {
			rec := do(t, h, http.MethodGet, path, "")

			expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
		}
	}
}

// countingItemRepository cuenta las lecturas de los ítems de un proyecto.
type countingItemRepository struct {
	backlog.Repository
	listCalls int
}

func (r *countingItemRepository) ListByProject(projectID int) ([]backlog.Item, error) {
	r.listCalls++
	return r.Repository.ListByProject(projectID)
}

// CA-35.4 · A través del handler, cada consulta lee el repositorio de ítems 2
// veces: 1 de metrics y 1 interna de defect.Manager.ListByProject (límite
// conocido de la SDD). El proyecto tiene dos sprints.
func TestConsultaMetricas_LeeLosItemsUnaSolaVez(t *testing.T) {
	itemRepo := &countingItemRepository{Repository: backlog.NewMemoryRepository()}
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(itemRepo, projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))
	createProject(t, h, "Sistema de ventas")
	first := hoursSprint(t, h, "Sprint 1", []hoursItemSpec{{"A", 10, 12}, {"B", 6, 5}})
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/cerrar", first), ""), http.StatusOK)
	second := hoursSprint(t, h, "Sprint 2", []hoursItemSpec{{"C", 4, 3}})

	for _, path := range []string{fmt.Sprintf("/api/sprints/%d/metricas", second), "/api/proyectos/1/metricas"} {
		itemRepo.listCalls = 0

		expectStatus(t, do(t, h, http.MethodGet, path, ""), http.StatusOK)

		if itemRepo.listCalls != 2 {
			t.Errorf("%s: lecturas de ítems = %d, se esperaba 2 (1 de metrics y 1 interna de defect)", path, itemRepo.listCalls)
		}
	}
}

// HU-37 · Ver las métricas en el dashboard con gráficos (Issue #55)
//
// CA-37.4: GET /api/proyectos/{id}/metricas/sprints devuelve cada sprint del
// proyecto con sus métricas, en orden de ID, con los formatos de
// sprintResponse (HU-10) y sprintMetricsResponse (HU-32 a HU-35). El orden de
// los campos se verifica con expectRawJSON de HU-35. CA-37.3: las métricas de
// cada sprint son las de GET /api/sprints/{id}/metricas.

// secondSprint es un segundo sprint del proyecto 1, que se crea después de
// cerrar el primero (solo puede haber un sprint abierto).
const secondSprint = `{"nombre":"Sprint 2","fecha_inicio":"2026-10-15","fecha_fin":"2026-10-28","objetivo":"Mostrar el dashboard"}`

// CA-37.1 y CA-37.4 · Contrato con el ejemplo de TEC-03 en el sprint 1,
// cerrado, y el sprint 2 abierto y sin ítems: las métricas no disponibles
// del sprint 2 son null.
func TestMetricasPorSprint_ContratoConUnSprintCerradoYOtroAbierto(t *testing.T) {
	h := newHandler()
	id := contractTEC03(t, h)
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/cerrar", id), ""), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", secondSprint), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas/sprints", "")

	expectRawJSON(t, rec, `[
  {
    "sprint": {
      "id": 1,
      "proyecto_id": 1,
      "nombre": "Sprint 1",
      "fecha_inicio": "2026-10-01",
      "fecha_fin": "2026-10-14",
      "objetivo": "Gestionar el backlog",
      "estado": "CERRADO"
    },
    "metricas": {
      "sprint_id": 1,
      "sp_planificados": 10,
      "sp_completados": 8,
      "items_sin_estimar": 1,
      "items_totales": 4,
      "items_terminados": 2,
      "porcentaje_completadas": 50,
      "horas_estimadas": 20,
      "horas_reales": 20,
      "desviacion_horas": 0,
      "desviacion_porcentual": 0,
      "items_sin_horas_estimadas": 1,
      "defectos_detectados": 1,
      "defectos_resueltos": 1
    }
  },
  {
    "sprint": {
      "id": 2,
      "proyecto_id": 1,
      "nombre": "Sprint 2",
      "fecha_inicio": "2026-10-15",
      "fecha_fin": "2026-10-28",
      "objetivo": "Mostrar el dashboard",
      "estado": "ABIERTO"
    },
    "metricas": {
      "sprint_id": 2,
      "sp_planificados": 0,
      "sp_completados": 0,
      "items_sin_estimar": 0,
      "items_totales": 0,
      "items_terminados": 0,
      "porcentaje_completadas": null,
      "horas_estimadas": 0,
      "horas_reales": 0,
      "desviacion_horas": null,
      "desviacion_porcentual": null,
      "items_sin_horas_estimadas": 0,
      "defectos_detectados": 0,
      "defectos_resueltos": 0
    }
  }
]`)
}

// CA-37.3 · Las métricas de cada elemento son exactamente el cuerpo de
// GET /api/sprints/{id}/metricas, y el sprint el de la lista de sprints.
func TestMetricasPorSprint_ValoresIgualesALaConsultaDeCadaSprint(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	first := hoursSprint(t, h, "Sprint 1", []hoursItemSpec{{"A", 10, 12}, {"B", 6, 5}})
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/cerrar", first), ""), http.StatusOK)
	second := hoursSprint(t, h, "Sprint 2", []hoursItemSpec{{"C", 4, 3}})

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas/sprints", "")

	expectStatus(t, rec, http.StatusOK)
	var got []struct {
		Sprint   sprintJSON      `json:"sprint"`
		Metricas json.RawMessage `json:"metricas"`
	}
	decode(t, rec, &got)
	ids := []int{first, second}
	if len(got) != len(ids) {
		t.Fatalf("la lista tiene %d sprints, se esperaba %d", len(got), len(ids))
	}
	for i, entry := range got {
		if entry.Sprint.ID != ids[i] {
			t.Errorf("posición %d: sprint %d, se esperaba %d", i, entry.Sprint.ID, ids[i])
		}
		single := do(t, h, http.MethodGet, fmt.Sprintf("/api/sprints/%d/metricas", ids[i]), "")
		var want, compact bytes.Buffer
		if err := json.Compact(&want, single.Body.Bytes()); err != nil {
			t.Fatalf("métricas del sprint %d: %v", ids[i], err)
		}
		if err := json.Compact(&compact, entry.Metricas); err != nil {
			t.Fatalf("métricas de la lista: %v", err)
		}
		if !bytes.Equal(compact.Bytes(), want.Bytes()) {
			t.Errorf("sprint %d: métricas de la lista =\n%s\nGET /api/sprints/%d/metricas =\n%s", ids[i], compact.String(), ids[i], want.String())
		}
	}
}

// CA-37.4 · Un proyecto sin sprints devuelve 200 y [], no null.
func TestMetricasPorSprint_SinSprintsDevuelveListaVacia(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas/sprints", "")

	expectRawJSON(t, rec, `[]`)
}

// CA-37.4 · Un {id} no numérico devuelve 400.
func TestMetricasPorSprint_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos/abc/metricas/sprints", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-37.4 · Un proyecto inexistente, 0 o negativo devuelve 404.
func TestMetricasPorSprint_ProyectoInexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodGet, "/api/proyectos/"+id+"/metricas/sprints", "")

		expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
	}
}

// CA-37.4 · Una falla al leer los esfuerzos o los defectos devuelve 500 sin
// detalle. Usa los repositorios de prueba de HU-33
// (failingEffortListRepository) y de HU-31 (brokenDefectRepository).
func TestMetricasPorSprint_FallaDeUnRepositorioDevuelve500(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(failingEffortListRepository{Repository: effort.NewMemoryRepository(), err: errors.New("base caída")}, sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	effortsBroken := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))
	createProject(t, effortsBroken, "Sistema de ventas")
	expectStatus(t, do(t, effortsBroken, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	defectsBroken := newHandlerWithDefectRepository(t, brokenDefectRepository{})

	for _, h := range []http.Handler{effortsBroken, defectsBroken} {
		rec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas/sprints", "")

		expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
	}
}
