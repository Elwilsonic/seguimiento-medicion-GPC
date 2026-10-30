package httpapi_test

import (
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

// HU-28 · Corregir un registro de esfuerzo (Issue #46)
//
// Trazabilidad: CA-28.7. PATCH /api/esfuerzos/{id}/horas expone la corrección
// por HTTP, con un caso exitoso y un test por cada código de la tabla HTTP de
// la SDD (400, 404, 409 y 500). CA-28.5: las métricas y la consulta de
// esfuerzos del sprint reflejan la corrección.

const correctEffortPath = "/api/esfuerzos/1/horas"

// newHandlerWithRegisteredEffort arma la API con el esfuerzo 1 registrado en
// el ítem 1 del sprint 1 ABIERTO, con 2.5 horas (validEffort).
func newHandlerWithRegisteredEffort(t *testing.T) http.Handler {
	t.Helper()
	h := newHandlerWithEffort(t)
	expectStatus(t, do(t, h, http.MethodPost, effortsPath, validEffort), http.StatusCreated)
	return h
}

func TestCorregirEsfuerzo_Devuelve200ConElEsfuerzoCorregido(t *testing.T) {
	h := newHandlerWithRegisteredEffort(t)

	rec := do(t, h, http.MethodPatch, correctEffortPath, `{"horas":10}`)

	expectStatus(t, rec, http.StatusOK)
	var got effortJSON
	decode(t, rec, &got)
	want := effortJSON{ID: 1, SprintID: 1, ItemID: 1, Integrante: "García Santiago",
		Fecha: "2026-10-24", Actividad: "Implementar el registro", Horas: 10}
	if got != want {
		t.Errorf("esfuerzo = %+v, se esperaba %+v", got, want)
	}
}

// CA-28.2 · Horas 0, negativas, ausentes o null devuelven 400 con el error de
// HU-26.
func TestCorregirEsfuerzo_HorasInvalidasDevuelve400(t *testing.T) {
	for _, body := range []string{`{"horas":0}`, `{"horas":-1}`, `{}`, `{"horas":null}`} {
		h := newHandlerWithRegisteredEffort(t)

		rec := do(t, h, http.MethodPatch, correctEffortPath, body)

		expectError(t, rec, http.StatusBadRequest, effort.ErrInvalidHours.Error())
	}
}

func TestCorregirEsfuerzo_JSONInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithRegisteredEffort(t)

	rec := do(t, h, http.MethodPatch, correctEffortPath, `{"horas":`)

	expectStatus(t, rec, http.StatusBadRequest)
}

func TestCorregirEsfuerzo_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPatch, "/api/esfuerzos/abc/horas", `{"horas":10}`)

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-28.4 · Un esfuerzo inexistente, 0 o negativo devuelve 404.
func TestCorregirEsfuerzo_InexistenteDevuelve404(t *testing.T) {
	h := newHandlerWithRegisteredEffort(t)
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodPatch, "/api/esfuerzos/"+id+"/horas", `{"horas":10}`)

		expectError(t, rec, http.StatusNotFound, effort.ErrEffortNotFound.Error())
	}
}

// CA-28.3 · Un esfuerzo de un sprint cerrado devuelve 409 y no cambia.
func TestCorregirEsfuerzo_SprintCerradoDevuelve409(t *testing.T) {
	h := newHandlerWithRegisteredEffort(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)

	rec := do(t, h, http.MethodPatch, correctEffortPath, `{"horas":10}`)

	expectError(t, rec, http.StatusConflict, sprint.ErrSprintClosed.Error())
	list := do(t, h, http.MethodGet, "/api/sprints/1/esfuerzos", "")
	expectStatus(t, list, http.StatusOK)
	if !strings.Contains(list.Body.String(), `"total_horas":2.5`) {
		t.Errorf("los esfuerzos cambiaron: %s", list.Body.String())
	}
}

// CA-28.5 · Después de corregir, la consulta de esfuerzos (HU-27) y las
// métricas del sprint (HU-35) muestran las horas corregidas.
func TestCorregirEsfuerzo_LasMetricasYLaConsultaReflejanLaCorreccion(t *testing.T) {
	h := newHandlerWithRegisteredEffort(t)

	expectStatus(t, do(t, h, http.MethodPatch, correctEffortPath, `{"horas":10}`), http.StatusOK)

	list := do(t, h, http.MethodGet, "/api/sprints/1/esfuerzos", "")
	expectStatus(t, list, http.StatusOK)
	if !strings.Contains(list.Body.String(), `"total_horas":10`) {
		t.Errorf("la consulta de esfuerzos no refleja la corrección: %s", list.Body.String())
	}
	metrics := do(t, h, http.MethodGet, "/api/sprints/1/metricas", "")
	expectStatus(t, metrics, http.StatusOK)
	if !strings.Contains(metrics.Body.String(), `"horas_reales":10`) {
		t.Errorf("las métricas no reflejan la corrección: %s", metrics.Body.String())
	}
}

// CA-28.7 · Una falla del repositorio devuelve 500 sin exponer el detalle.
func TestCorregirEsfuerzo_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(brokenEffortRepository{}, sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))

	rec := do(t, h, http.MethodPatch, correctEffortPath, `{"horas":10}`)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
