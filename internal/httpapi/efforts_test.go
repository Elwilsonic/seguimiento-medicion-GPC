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

// HU-26 · Registrar esfuerzo (Issue #44)
//
// Trazabilidad: CA-26.11. POST /api/sprints/{id}/items/{itemId}/esfuerzos
// expone el registro por HTTP. Hay un caso exitoso y un test por cada código de
// error de la tabla HTTP de la SDD (400, 404, 409, 413 y 500).

type effortJSON struct {
	ID         int     `json:"id"`
	SprintID   int     `json:"sprint_id"`
	ItemID     int     `json:"item_id"`
	Integrante string  `json:"integrante"`
	Fecha      string  `json:"fecha"`
	Actividad  string  `json:"actividad"`
	Horas      float64 `json:"horas"`
}

const validEffort = `{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar el registro","horas":2.5}`

const effortsPath = "/api/sprints/1/items/1/esfuerzos"

// newHandlerWithEffort arma la API con el proyecto 1, el ítem 1 y el sprint 1
// ABIERTO, con el ítem asignado al sprint.
func newHandlerWithEffort(t *testing.T) http.Handler {
	t.Helper()
	h := newHandlerWithSprint(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/items/1", ""), http.StatusOK)
	return h
}

func TestRegistrarEsfuerzo_Devuelve201ConElEsfuerzo(t *testing.T) {
	h := newHandlerWithEffort(t)

	rec := do(t, h, http.MethodPost, effortsPath, validEffort)

	expectStatus(t, rec, http.StatusCreated)
	var got effortJSON
	decode(t, rec, &got)
	want := effortJSON{ID: 1, SprintID: 1, ItemID: 1, Integrante: "García Santiago",
		Fecha: "2026-10-24", Actividad: "Implementar el registro", Horas: 2.5}
	if got != want {
		t.Errorf("esfuerzo = %+v, se esperaba %+v", got, want)
	}
}

func TestRegistrarEsfuerzo_AceptaHorasComoTextoNumerico(t *testing.T) {
	h := newHandlerWithEffort(t)
	body := `{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar","horas":"2.5"}`

	rec := do(t, h, http.MethodPost, effortsPath, body)

	expectStatus(t, rec, http.StatusCreated)
	var got effortJSON
	decode(t, rec, &got)
	if got.Horas != 2.5 {
		t.Errorf("horas = %v, se esperaba 2.5", got.Horas)
	}
}

func TestRegistrarEsfuerzo_DatosInvalidosDevuelven400(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"integrante vacío": {
			`{"integrante":" ","fecha":"2026-10-24","actividad":"Implementar","horas":2}`,
			effort.ErrEmptyMember.Error(),
		},
		"actividad vacía": {
			`{"integrante":"García Santiago","fecha":"2026-10-24","actividad":" ","horas":2}`,
			effort.ErrEmptyActivity.Error(),
		},
		"fecha ausente": {
			`{"integrante":"García Santiago","actividad":"Implementar","horas":2}`,
			effort.ErrMissingDate.Error(),
		},
		"fecha con otro formato": {
			`{"integrante":"García Santiago","fecha":"24/10/2026","actividad":"Implementar","horas":2}`,
			"las fechas deben tener el formato AAAA-MM-DD",
		},
		"horas cero": {
			`{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar","horas":0}`,
			effort.ErrInvalidHours.Error(),
		},
		"horas negativas": {
			`{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar","horas":-1}`,
			effort.ErrInvalidHours.Error(),
		},
		"horas null": {
			`{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar","horas":null}`,
			effort.ErrInvalidHours.Error(),
		},
		"horas ausentes": {
			`{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar"}`,
			effort.ErrInvalidHours.Error(),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithEffort(t)

			rec := do(t, h, http.MethodPost, effortsPath, tc.body)

			expectError(t, rec, http.StatusBadRequest, tc.want)
		})
	}
}

func TestRegistrarEsfuerzo_CuerpoInvalidoDevuelve400(t *testing.T) {
	cases := map[string]string{
		"JSON inválido":          `{"integrante":`,
		"texto después del JSON": validEffort + ` basura`,
		"horas booleanas":        `{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar","horas":true}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithEffort(t)

			rec := do(t, h, http.MethodPost, effortsPath, body)

			expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
		})
	}
}

func TestRegistrarEsfuerzo_HorasNoNumericasDevuelven400(t *testing.T) {
	h := newHandlerWithEffort(t)
	body := `{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar","horas":"muchas"}`

	rec := do(t, h, http.MethodPost, effortsPath, body)

	expectStatus(t, rec, http.StatusBadRequest)
}

func TestRegistrarEsfuerzo_IDNoNumericoDevuelve400(t *testing.T) {
	cases := map[string]string{
		"sprint": "/api/sprints/abc/items/1/esfuerzos",
		"ítem":   "/api/sprints/1/items/abc/esfuerzos",
	}
	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithEffort(t)

			rec := do(t, h, http.MethodPost, path, validEffort)

			expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
		})
	}
}

func TestRegistrarEsfuerzo_InexistenteDevuelve404(t *testing.T) {
	cases := map[string]struct {
		path string
		want error
	}{
		"sprint inexistente": {"/api/sprints/999/items/1/esfuerzos", sprint.ErrSprintNotFound},
		"ítem inexistente":   {"/api/sprints/1/items/999/esfuerzos", backlog.ErrItemNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithEffort(t)

			rec := do(t, h, http.MethodPost, tc.path, validEffort)

			expectError(t, rec, http.StatusNotFound, tc.want.Error())
		})
	}
}

func TestRegistrarEsfuerzo_SprintCerradoDevuelve409(t *testing.T) {
	h := newHandlerWithEffort(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)

	rec := do(t, h, http.MethodPost, effortsPath, validEffort)

	expectError(t, rec, http.StatusConflict, sprint.ErrSprintClosed.Error())
}

func TestRegistrarEsfuerzo_ItemSinSprintDevuelve409(t *testing.T) {
	h := newHandlerWithSprint(t)

	rec := do(t, h, http.MethodPost, effortsPath, validEffort)

	expectError(t, rec, http.StatusConflict, sprint.ErrItemNotAssignedToSprint.Error())
}

func TestRegistrarEsfuerzo_CuerpoMayorA1MiBDa413(t *testing.T) {
	h := newHandlerWithEffort(t)
	body := "{" + strings.Repeat(" ", 1<<20) + validEffort[1:]

	rec := do(t, h, http.MethodPost, effortsPath, body)

	expectError(t, rec, http.StatusRequestEntityTooLarge, "el cuerpo del pedido supera el tamaño máximo de 1 MiB")

	// El cuerpo rechazado no se procesó: el siguiente registro recibe el ID 1.
	rec = do(t, h, http.MethodPost, effortsPath, validEffort)
	expectStatus(t, rec, http.StatusCreated)
	var got effortJSON
	decode(t, rec, &got)
	if got.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", got.ID)
	}
}

func TestRegistrarEsfuerzo_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
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
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/items/1", ""), http.StatusOK)

	rec := do(t, h, http.MethodPost, effortsPath, validEffort)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// brokenEffortRepository siempre falla al guardar.
type brokenEffortRepository struct{}

func (brokenEffortRepository) Save(effort.Effort) (effort.Effort, error) {
	return effort.Effort{}, errStorage
}
