package httpapi_test

import (
	"encoding/json"
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

// HU-39 · Generar el reporte de un sprint (Issue #57)
//
// CA-39.8: el reporte se genera en GET /api/sprints/{id}/reporte. Los
// ejemplos 1 y 2 del contrato de la SDD se verifican comparando el cuerpo
// crudo con el JSON esperado, ambos con json.Compact (como HU-35 y HU-38).
// RN-39.9: cada parte tiene el formato de una respuesta que ya existe, y con
// los mismos datos es igual a ella.

const tec03SprintBody = `{"nombre":"Sprint 1","fecha_inicio":"2026-10-05","fecha_fin":"2026-10-16","objetivo":"Registrar el esfuerzo del equipo"}`

// sprintReportTEC03 crea por la API el sprint del ejemplo 1 de la SDD: el
// ejemplo de TEC-03 en el sprint 1 del proyecto 1, con 12 h de Ana en A, 5 h
// de Luis en B y 3 h de Ana en C, un defecto de A detectado y resuelto en el
// sprint y el sprint cerrado.
func sprintReportTEC03(t *testing.T, h http.Handler) {
	t.Helper()
	createProject(t, h, "Software Metrics")
	var sp sprintJSON
	rec := do(t, h, http.MethodPost, "/api/proyectos/1/sprints", tec03SprintBody)
	expectStatus(t, rec, http.StatusCreated)
	decode(t, rec, &sp)
	specs := []struct {
		item   contractItemSpec
		effort *reportEffortSpec
	}{
		{contractItemSpec{"A", 5, 10, "TERMINADO", 0}, &reportEffortSpec{"Ana", "2026-10-06", "Desarrollo", 12}},
		{contractItemSpec{"B", 3, 6, "TERMINADO", 0}, &reportEffortSpec{"Luis", "2026-10-07", "Desarrollo", 5}},
		{contractItemSpec{"C", 2, 0, "EN_CURSO", 0}, &reportEffortSpec{"Ana", "2026-10-08", "Pruebas", 3}},
		{contractItemSpec{"D", 0, 4, "PENDIENTE", 0}, nil},
	}
	for _, spec := range specs {
		it := createHoursItem(t, h, spec.item.title, spec.item.hours)
		if spec.item.points > 0 {
			expectStatus(t, do(t, h, http.MethodPatch, fmt.Sprintf("/api/items/%d/story-points", it.ID), fmt.Sprintf(`{"story_points":%d}`, spec.item.points)), http.StatusOK)
		}
		expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d", sp.ID, it.ID), ""), http.StatusOK)
		if ef := spec.effort; ef != nil {
			body := fmt.Sprintf(`{"integrante":%q,"fecha":%q,"actividad":%q,"horas":%v}`, ef.member, ef.date, ef.activity, ef.hours)
			expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d/esfuerzos", sp.ID, it.ID), body), http.StatusCreated)
		}
		if spec.item.status != "PENDIENTE" {
			expectStatus(t, do(t, h, http.MethodPatch, fmt.Sprintf("/api/items/%d/estado", it.ID), `{"estado":"EN_CURSO"}`), http.StatusOK)
		}
		if spec.item.status == "TERMINADO" {
			expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/items/%d/completar", sp.ID, it.ID), ""), http.StatusOK)
		}
	}
	defectBody := fmt.Sprintf(`{"descripcion":"El total de horas no suma los decimales","severidad":"ALTA","sprint_deteccion_id":%d}`, sp.ID)
	expectStatus(t, do(t, h, http.MethodPost, "/api/items/1/defectos", defectBody), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/1/estado", `{"estado":"EN_CORRECCION"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/1/estado", fmt.Sprintf(`{"estado":"RESUELTO","sprint_resolucion_id":%d}`, sp.ID)), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/%d/cerrar", sp.ID), ""), http.StatusOK)
}

// CA-39.1 y CA-39.8 · Contrato del reporte con el ejemplo 1 de la SDD.
func TestReporteSprint_ContratoDelEjemploDeTEC03(t *testing.T) {
	h := newHandler()
	sprintReportTEC03(t, h)

	rec := do(t, h, http.MethodGet, "/api/sprints/1/reporte", "")

	expectRawJSON(t, rec, `{
  "sprint": {
    "id": 1,
    "proyecto_id": 1,
    "nombre": "Sprint 1",
    "fecha_inicio": "2026-10-05",
    "fecha_fin": "2026-10-16",
    "objetivo": "Registrar el esfuerzo del equipo",
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
  },
  "esfuerzos": {
    "esfuerzos": [
      {"id": 1, "sprint_id": 1, "item_id": 1, "integrante": "Ana", "fecha": "2026-10-06", "actividad": "Desarrollo", "horas": 12},
      {"id": 2, "sprint_id": 1, "item_id": 2, "integrante": "Luis", "fecha": "2026-10-07", "actividad": "Desarrollo", "horas": 5},
      {"id": 3, "sprint_id": 1, "item_id": 3, "integrante": "Ana", "fecha": "2026-10-08", "actividad": "Pruebas", "horas": 3}
    ],
    "total_horas": 20
  },
  "defectos": {
    "defectos": [
      {
        "id": 1,
        "descripcion": "El total de horas no suma los decimales",
        "severidad": "ALTA",
        "estado": "RESUELTO",
        "item_id": 1,
        "sprint_deteccion_id": 1,
        "sprint_resolucion_id": 1
      }
    ],
    "cantidad": 1
  },
  "integrantes": {
    "integrantes": ["Ana", "Luis"],
    "cantidad": 2
  }
}`)
}

// CA-39.2, CA-39.4 y CA-39.8 · Contrato del reporte con el ejemplo 2 de la
// SDD: un sprint recién creado, abierto, con las listas vacías ([], no null)
// y las tres métricas no disponibles en null.
func TestReporteSprint_ContratoDelSprintVacio(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Software Metrics")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	createProject(t, h, "Proyecto sin datos")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/2/sprints",
		`{"nombre":"Sprint vacío","fecha_inicio":"2026-10-19","fecha_fin":"2026-10-30","objetivo":"Sin trabajo todavía"}`), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/sprints/2/reporte", "")

	expectRawJSON(t, rec, `{
  "sprint": {
    "id": 2,
    "proyecto_id": 2,
    "nombre": "Sprint vacío",
    "fecha_inicio": "2026-10-19",
    "fecha_fin": "2026-10-30",
    "objetivo": "Sin trabajo todavía",
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
  },
  "esfuerzos": {
    "esfuerzos": [],
    "total_horas": 0
  },
  "defectos": {
    "defectos": [],
    "cantidad": 0
  },
  "integrantes": {
    "integrantes": [],
    "cantidad": 0
  }
}`)
}

// reportParts decodifica las partes del reporte sin interpretarlas.
func reportParts(t *testing.T, h http.Handler, path string) map[string]json.RawMessage {
	t.Helper()
	rec := do(t, h, http.MethodGet, path, "")
	expectStatus(t, rec, http.StatusOK)
	var parts map[string]json.RawMessage
	decode(t, rec, &parts)
	return parts
}

// CA-39.5 y RN-39.9 · Con los mismos datos, "metricas", "esfuerzos" y
// "defectos" son iguales a las respuestas de GET /api/sprints/{id}/metricas,
// GET /api/sprints/{id}/esfuerzos y GET /api/sprints/{id}/defectos (en el
// ejemplo 1 el único defecto se detectó en el sprint), y "sprint" es igual al
// campo sprint de GET /api/proyectos/{id}/metricas/sprints (HU-37).
func TestReporteSprint_CadaParteEsIgualASuEndpoint(t *testing.T) {
	h := newHandler()
	sprintReportTEC03(t, h)

	parts := reportParts(t, h, "/api/sprints/1/reporte")

	for part, path := range map[string]string{
		"metricas":  "/api/sprints/1/metricas",
		"esfuerzos": "/api/sprints/1/esfuerzos",
		"defectos":  "/api/sprints/1/defectos",
	} {
		source := do(t, h, http.MethodGet, path, "")
		expectStatus(t, source, http.StatusOK)
		raw, ok := parts[part]
		if !ok {
			t.Errorf("el reporte no tiene la parte %q", part)
			continue
		}
		if got, want := compactJSON(t, raw), compactJSON(t, source.Body.Bytes()); got != want {
			t.Errorf("%s =\n%s\nse esperaba la respuesta de %s\n%s", part, got, path, want)
		}
	}
	listRec := do(t, h, http.MethodGet, "/api/proyectos/1/metricas/sprints", "")
	expectStatus(t, listRec, http.StatusOK)
	var list []map[string]json.RawMessage
	decode(t, listRec, &list)
	if len(list) != 1 {
		t.Fatalf("GET /api/proyectos/1/metricas/sprints tiene %d sprints, se esperaba 1", len(list))
	}
	if got, want := compactJSON(t, parts["sprint"]), compactJSON(t, list[0]["sprint"]); got != want {
		t.Errorf("sprint =\n%s\nse esperaba\n%s", got, want)
	}
}

// CA-39.5 y RN-39.5 · Ejemplo 3 de la SDD: un defecto detectado en el sprint
// 1 y resuelto en el 2 aparece en el reporte del sprint 2, que lo cuenta en
// M10, aunque GET /api/sprints/2/defectos no lo lista.
func TestReporteSprint_DefectoResueltoEnElSprintYDetectadoEnOtro(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Software Metrics")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", tec03SprintBody), http.StatusCreated)
	it := createHoursItem(t, h, "A", 10)
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/1/items/%d", it.ID), ""), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/items/%d/defectos", it.ID),
		`{"descripcion":"El total de horas no suma los decimales","severidad":"ALTA","sprint_deteccion_id":1}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/1/estado", `{"estado":"EN_CORRECCION"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints",
		`{"nombre":"Sprint 2","fecha_inicio":"2026-10-19","fecha_fin":"2026-10-30","objetivo":"Corregir defectos"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/1/estado", `{"estado":"RESUELTO","sprint_resolucion_id":2}`), http.StatusOK)

	parts := reportParts(t, h, "/api/sprints/2/reporte")

	var m struct {
		Detectados int `json:"defectos_detectados"`
		Resueltos  int `json:"defectos_resueltos"`
	}
	if err := json.Unmarshal(parts["metricas"], &m); err != nil || m.Detectados != 0 || m.Resueltos != 1 {
		t.Errorf("metricas = %s (%v), se esperaban 0 detectados y 1 resuelto", parts["metricas"], err)
	}
	want := `{"defectos":[{"id":1,"descripcion":"El total de horas no suma los decimales","severidad":"ALTA","estado":"RESUELTO",` +
		`"item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":2}],"cantidad":1}`
	if raw, ok := parts["defectos"]; !ok || compactJSON(t, raw) != want {
		t.Errorf("defectos = %s, se esperaba %s", parts["defectos"], want)
	}
	listed := do(t, h, http.MethodGet, "/api/sprints/2/defectos", "")
	expectStatus(t, listed, http.StatusOK)
	if !strings.Contains(listed.Body.String(), `"cantidad":0`) {
		t.Errorf("GET /api/sprints/2/defectos = %s, se esperaba que no lo liste", listed.Body.String())
	}
}

// RN-39.10 · La API redondea el total de horas a un decimal, como HU-27: 0.1
// y 0.2 horas suman 0.3.
func TestReporteSprint_RedondeaElTotalDeHorasAUnDecimal(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	it := createHoursItem(t, h, "A", 1)
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/1/items/%d", it.ID), ""), http.StatusOK)
	for _, hours := range []string{"0.1", "0.2"} {
		body := `{"integrante":"Ana","fecha":"2026-10-06","actividad":"Desarrollo","horas":` + hours + `}`
		expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/1/items/%d/esfuerzos", it.ID), body), http.StatusCreated)
	}

	rec := do(t, h, http.MethodGet, "/api/sprints/1/reporte", "")

	expectStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), `"total_horas":0.3`) {
		t.Errorf("la respuesta no tiene total_horas 0.3: %s", rec.Body.String())
	}
}

// CA-39.3 y CA-39.8 · Un sprint inexistente, 0 o negativo devuelve 404.
func TestReporteSprint_InexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodGet, "/api/sprints/"+id+"/reporte", "")

		expectError(t, rec, http.StatusNotFound, sprint.ErrSprintNotFound.Error())
	}
}

// CA-39.8 · Un {id} no numérico devuelve 400.
func TestReporteSprint_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/sprints/abc/reporte", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-39.8 · Una falla del repositorio devuelve 500 sin exponer el detalle.
func TestReporteSprint_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(brokenEffortRepository{}, sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/sprints/1/reporte", "")

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
