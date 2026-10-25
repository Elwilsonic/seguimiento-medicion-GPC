package httpapi_test

import (
	"net/http"
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

// HU-27 — Consultar los registros de esfuerzo (Issue #45)
//
// Trazabilidad: CA-27.10. Las dos rutas GET exponen la consulta; hay un caso
// exitoso por ruta, lista vacía, 404, 500 y el redondeo de total_horas.

type effortListJSON struct {
	Esfuerzos  []effortJSON `json:"esfuerzos"`
	TotalHoras float64      `json:"total_horas"`
}

func TestListarEsfuerzosPorSprint_Devuelve200ConLosEsfuerzosYElTotal(t *testing.T) {
	h := newHandlerWithEffort(t)
	expectStatus(t, do(t, h, http.MethodPost, effortsPath, validEffort), http.StatusCreated)
	second := `{"integrante":"Buttini","fecha":"2026-10-24","actividad":"Revisar el registro","horas":1}`
	expectStatus(t, do(t, h, http.MethodPost, effortsPath, second), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/sprints/1/esfuerzos", "")

	expectStatus(t, rec, http.StatusOK)
	var got effortListJSON
	decode(t, rec, &got)
	if len(got.Esfuerzos) != 2 {
		t.Fatalf("esfuerzos = %d, se esperaban 2", len(got.Esfuerzos))
	}
	if got.Esfuerzos[0].ID != 1 || got.Esfuerzos[1].ID != 2 {
		t.Errorf("IDs = %d y %d, se esperaban 1 y 2", got.Esfuerzos[0].ID, got.Esfuerzos[1].ID)
	}
	if got.Esfuerzos[0].Integrante != "García Santiago" || got.Esfuerzos[0].Actividad != "Implementar el registro" {
		t.Errorf("primer esfuerzo = %+v", got.Esfuerzos[0])
	}
	if got.TotalHoras != 3.5 {
		t.Errorf("total_horas = %v, se esperaba 3.5", got.TotalHoras)
	}
}

func TestListarEsfuerzosPorItem_Devuelve200ConLosEsfuerzosYElTotal(t *testing.T) {
	h := newHandlerWithEffort(t)
	expectStatus(t, do(t, h, http.MethodPost, effortsPath, validEffort), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/items/1/esfuerzos", "")

	expectStatus(t, rec, http.StatusOK)
	var got effortListJSON
	decode(t, rec, &got)
	if len(got.Esfuerzos) != 1 || got.Esfuerzos[0].ID != 1 {
		t.Fatalf("esfuerzos = %+v, se esperaba el esfuerzo 1", got.Esfuerzos)
	}
	if got.TotalHoras != 2.5 {
		t.Errorf("total_horas = %v, se esperaba 2.5", got.TotalHoras)
	}
}

func TestListarEsfuerzos_SinRegistrosDevuelveListaVaciaYTotalCero(t *testing.T) {
	h := newHandlerWithSprint(t)

	for _, path := range []string{"/api/sprints/1/esfuerzos", "/api/items/1/esfuerzos"} {
		t.Run(path, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, path, "")

			expectStatus(t, rec, http.StatusOK)
			if body := rec.Body.String(); !strings.Contains(body, `"esfuerzos":[]`) {
				t.Errorf("cuerpo = %s, se esperaba una lista vacía", body)
			}
			var got effortListJSON
			decode(t, rec, &got)
			if got.Esfuerzos == nil || len(got.Esfuerzos) != 0 {
				t.Errorf("esfuerzos = %+v, se esperaba una lista vacía no nil", got.Esfuerzos)
			}
			if got.TotalHoras != 0 {
				t.Errorf("total_horas = %v, se esperaba 0", got.TotalHoras)
			}
		})
	}
}

func TestListarEsfuerzos_InexistenteDevuelve404(t *testing.T) {
	h := newHandler()

	cases := map[string]struct {
		path string
		want error
	}{
		"ítem inexistente":   {"/api/items/999/esfuerzos", backlog.ErrItemNotFound},
		"sprint inexistente": {"/api/sprints/999/esfuerzos", sprint.ErrSprintNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, tc.path, "")

			expectError(t, rec, http.StatusNotFound, tc.want.Error())
		})
	}
}

func TestListarEsfuerzos_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	h := newHandlerWithBrokenEffortLists(t)

	for _, path := range []string{"/api/sprints/1/esfuerzos", "/api/items/1/esfuerzos"} {
		t.Run(path, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, path, "")

			expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
		})
	}
}

func TestListarEsfuerzos_RedondeaElTotalAUnDecimal(t *testing.T) {
	cases := []struct {
		name    string
		horas   string
		want    float64
		wantRaw string
	}{
		{"total muy chico", "0.001", 0, "0.001"},
		{"total con decimal", "0.06", 0.1, "0.06"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandlerWithEffort(t)
			body := `{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Ajustar","horas":` + tc.horas + `}`
			expectStatus(t, do(t, h, http.MethodPost, effortsPath, body), http.StatusCreated)

			rec := do(t, h, http.MethodGet, "/api/items/1/esfuerzos", "")

			expectStatus(t, rec, http.StatusOK)
			var got effortListJSON
			decode(t, rec, &got)
			if got.TotalHoras != tc.want {
				t.Errorf("total_horas = %v, se esperaba %v", got.TotalHoras, tc.want)
			}
			if len(got.Esfuerzos) != 1 || got.Esfuerzos[0].Horas != mustFloat(t, tc.wantRaw) {
				t.Errorf("el esfuerzo no conserva sus horas exactas: %+v", got.Esfuerzos)
			}
		})
	}
}

func newHandlerWithBrokenEffortLists(t *testing.T) http.Handler {
	t.Helper()

	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(brokenEffortRepository{}, sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)

	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	return h
}

func mustFloat(t *testing.T, raw string) float64 {
	t.Helper()
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		t.Fatalf("número de prueba inválido %q: %v", raw, err)
	}
	return value
}
