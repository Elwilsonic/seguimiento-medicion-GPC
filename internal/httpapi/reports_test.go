package httpapi_test

import (
	"bytes"
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

// HU-38 · Generar el reporte del proyecto (Issue #56)
//
// CA-38.7: el reporte se genera en GET /api/proyectos/{id}/reporte. Los dos
// ejemplos del contrato de la SDD se verifican comparando el cuerpo crudo con
// el JSON esperado, ambos con json.Compact, así se verifican también el orden
// de las partes y de los campos, los null y las listas vacías (como HU-35).
// RN-38.8: cada parte tiene el formato del endpoint que ya existe, y con los
// mismos datos es igual a su respuesta.

// reportEffortSpec es un esfuerzo del ejemplo 1 de la SDD.
type reportEffortSpec struct {
	member, date, activity string
	hours                  float64
}

// reportTEC03 crea por la API el proyecto del ejemplo 1 de la SDD: el
// ejemplo de TEC-03 en el sprint 1, con 12 h de Ana en A, 5 h de Luis en B y
// 3 h de Ana en C, un defecto de A detectado y resuelto en el sprint y el
// sprint cerrado.
func reportTEC03(t *testing.T, h http.Handler) {
	t.Helper()
	createProject(t, h, "Software Metrics")
	expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", `{"fecha_inicio":"2026-10-05"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPut, "/api/proyectos/1",
		`{"nombre":"Software Metrics","descripcion":"Estimación y seguimiento de proyectos","estado":"EN_CURSO"}`), http.StatusOK)
	var sp sprintJSON
	rec := do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint)
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

// CA-38.1 y CA-38.7 · Contrato del reporte con el ejemplo 1 de la SDD.
func TestReporteProyecto_ContratoDelEjemploDeTEC03(t *testing.T) {
	h := newHandler()
	reportTEC03(t, h)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/reporte", "")

	expectRawJSON(t, rec, `{
  "proyecto": {
    "id": 1,
    "nombre": "Software Metrics",
    "descripcion": "Estimación y seguimiento de proyectos",
    "estado": "EN_CURSO",
    "fecha_inicio": "2026-10-05"
  },
  "metricas": {
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
  },
  "esfuerzos": {
    "esfuerzos": [
      {
        "id": 1,
        "sprint_id": 1,
        "item_id": 1,
        "integrante": "Ana",
        "fecha": "2026-10-06",
        "actividad": "Desarrollo",
        "horas": 12
      },
      {
        "id": 2,
        "sprint_id": 1,
        "item_id": 2,
        "integrante": "Luis",
        "fecha": "2026-10-07",
        "actividad": "Desarrollo",
        "horas": 5
      },
      {
        "id": 3,
        "sprint_id": 1,
        "item_id": 3,
        "integrante": "Ana",
        "fecha": "2026-10-08",
        "actividad": "Pruebas",
        "horas": 3
      }
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

// CA-38.2 y CA-38.7 · Contrato del reporte con el ejemplo 2 de la SDD: un
// proyecto recién creado, con las listas vacías ([], no null) y las cuatro
// métricas no disponibles en null.
func TestReporteProyecto_ContratoDelProyectoVacio(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Software Metrics")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"Proyecto sin datos","descripcion":"Recién creado"}`), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/proyectos/2/reporte", "")

	expectRawJSON(t, rec, `{
  "proyecto": {
    "id": 2,
    "nombre": "Proyecto sin datos",
    "descripcion": "Recién creado",
    "estado": "PLANIFICADO",
    "fecha_inicio": ""
  },
  "metricas": {
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

// compactJSON normaliza un JSON con json.Compact para comparar cuerpos.
func compactJSON(t *testing.T, raw []byte) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		t.Fatalf("JSON inválido (%v): %s", err, raw)
	}
	return buf.String()
}

// CA-38.4 y RN-38.8 · Con los mismos datos, "proyecto", "metricas",
// "esfuerzos" y "defectos" son iguales a las respuestas de
// GET /api/proyectos/{id}, GET /api/proyectos/{id}/metricas,
// GET /api/sprints/{id}/esfuerzos (el proyecto tiene un solo sprint) y
// GET /api/proyectos/{id}/defectos.
func TestReporteProyecto_CadaParteEsIgualASuEndpoint(t *testing.T) {
	h := newHandler()
	reportTEC03(t, h)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/reporte", "")

	expectStatus(t, rec, http.StatusOK)
	var parts map[string]json.RawMessage
	decode(t, rec, &parts)
	for part, path := range map[string]string{
		"proyecto":  "/api/proyectos/1",
		"metricas":  "/api/proyectos/1/metricas",
		"esfuerzos": "/api/sprints/1/esfuerzos",
		"defectos":  "/api/proyectos/1/defectos",
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
}

// RN-38.9 · La API redondea el total de horas a un decimal, como HU-27: 0.1
// y 0.2 horas suman 0.3.
func TestReporteProyecto_RedondeaElTotalDeHorasAUnDecimal(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	it := createHoursItem(t, h, "A", 1)
	expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/1/items/%d", it.ID), ""), http.StatusOK)
	for _, hours := range []string{"0.1", "0.2"} {
		body := `{"integrante":"Ana","fecha":"2026-10-06","actividad":"Desarrollo","horas":` + hours + `}`
		expectStatus(t, do(t, h, http.MethodPost, fmt.Sprintf("/api/sprints/1/items/%d/esfuerzos", it.ID), body), http.StatusCreated)
	}

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/reporte", "")

	expectStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), `"total_horas":0.3`) {
		t.Errorf("la respuesta no tiene total_horas 0.3: %s", rec.Body.String())
	}
}

// CA-38.3 y CA-38.7 · Un proyecto inexistente, 0 o negativo devuelve 404.
func TestReporteProyecto_InexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodGet, "/api/proyectos/"+id+"/reporte", "")

		expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
	}
}

// CA-38.7 · Un {id} no numérico devuelve 400.
func TestReporteProyecto_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos/abc/reporte", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-38.7 · Una falla del repositorio devuelve 500 sin exponer el detalle.
func TestReporteProyecto_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(brokenItemRepository{}, projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/reporte", "")

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
