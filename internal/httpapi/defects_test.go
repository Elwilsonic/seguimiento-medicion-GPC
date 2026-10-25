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

// HU-29 · Registrar defecto por HTTP (Issue #47)
//
// Trazabilidad: CA-29.11. POST /api/items/{id}/defectos expone el registro de
// defectos. Hay un caso exitoso y un test por cada código de error de la tabla
// HTTP de la SDD (400, 404, 409, 413 y 500).

type defectJSON struct {
	ID                 int             `json:"id"`
	Descripcion        string          `json:"descripcion"`
	Severidad          defect.Severity `json:"severidad"`
	Estado             defect.Status   `json:"estado"`
	ItemID             int             `json:"item_id"`
	SprintDeteccionID  int             `json:"sprint_deteccion_id"`
	SprintResolucionID int             `json:"sprint_resolucion_id"`
}

const validDefect = `{"descripcion":"  El total no descuenta el descuento  ","severidad":" alta ","sprint_deteccion_id":1}`

const defectsPath = "/api/items/1/defectos"

// newHandlerWithDefects arma la API con el proyecto 1, su ítem 1 y el sprint 1
// abierto, y con un Manager de defectos inyectable.
func newHandlerWithDefects(t *testing.T) http.Handler {
	t.Helper()
	return newHandlerWithDefectRepository(t, defect.NewMemoryRepository())
}

func newHandlerWithDefectRepository(t *testing.T, repo defect.Repository) http.Handler {
	t.Helper()

	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(repo, items, sprints)
	h := httpapi.NewHandlerWithDefects(projects, items, sprints, efforts, pokerManager, defects)

	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	return h
}

func TestRegistrarDefecto_Devuelve201ConElDefecto(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, defectsPath, validDefect)

	expectStatus(t, rec, http.StatusCreated)
	var got defectJSON
	decode(t, rec, &got)
	want := defectJSON{
		ID:                 1,
		Descripcion:        "El total no descuenta el descuento",
		Severidad:          defect.SeverityHigh,
		Estado:             defect.StatusOpen,
		ItemID:             1,
		SprintDeteccionID:  1,
		SprintResolucionID: 0,
	}
	if got != want {
		t.Errorf("defecto = %+v, se esperaba %+v", got, want)
	}
	if !strings.Contains(rec.Body.String(), `"sprint_resolucion_id":0`) {
		t.Errorf("la respuesta no incluye el sprint de resolución en 0: %s", rec.Body.String())
	}
}

func TestRegistrarDefecto_IgnoraEstadoYSprintDeResolucion(t *testing.T) {
	h := newHandlerWithDefects(t)
	body := `{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":1,"estado":"RESUELTO","sprint_resolucion_id":9}`

	rec := do(t, h, http.MethodPost, defectsPath, body)

	expectStatus(t, rec, http.StatusCreated)
	var got defectJSON
	decode(t, rec, &got)
	if got.Estado != defect.StatusOpen {
		t.Errorf("estado = %q, se esperaba %q", got.Estado, defect.StatusOpen)
	}
	if got.SprintResolucionID != 0 {
		t.Errorf("sprint de resolución = %d, se esperaba 0", got.SprintResolucionID)
	}
}

func TestRegistrarDefecto_DatosInvalidosDevuelven400(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"descripción vacía": {
			`{"descripcion":"   ","severidad":"ALTA","sprint_deteccion_id":1}`,
			defect.ErrEmptyDescription.Error(),
		},
		"severidad inválida": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"URGENTE","sprint_deteccion_id":1}`,
			defect.ErrInvalidSeverity.Error(),
		},
		"sprint de detección ausente": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA"}`,
			defect.ErrMissingDetectionSprint.Error(),
		},
		"sprint de detección cero": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":0}`,
			defect.ErrMissingDetectionSprint.Error(),
		},
		"sprint de detección negativo": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":-1}`,
			defect.ErrMissingDetectionSprint.Error(),
		},
		"sprint de detección null": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":null}`,
			defect.ErrMissingDetectionSprint.Error(),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithDefects(t)

			rec := do(t, h, http.MethodPost, defectsPath, tc.body)

			expectError(t, rec, http.StatusBadRequest, tc.want)
		})
	}
}

func TestRegistrarDefecto_JSONInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, defectsPath, `{"descripcion":`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

func TestRegistrarDefecto_SprintDeDeteccionNoEnteroDevuelve400(t *testing.T) {
	cases := map[string]string{
		"decimal": `{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":1.5}`,
		"texto":   `{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":"abc"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithDefects(t)

			rec := do(t, h, http.MethodPost, defectsPath, body)

			expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
		})
	}
}

func TestRegistrarDefecto_TextoDespuesDelJSONDevuelve400(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, defectsPath, validDefect+` basura`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

func TestRegistrarDefecto_ItemIDInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, "/api/items/abc/defectos", validDefect)

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

func TestRegistrarDefecto_InexistenteDevuelve404(t *testing.T) {
	cases := map[string]struct {
		path string
		body string
		want string
	}{
		"historia inexistente": {
			"/api/items/999/defectos",
			validDefect,
			backlog.ErrItemNotFound.Error(),
		},
		"sprint de detección inexistente": {
			defectsPath,
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":999}`,
			sprint.ErrSprintNotFound.Error(),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithDefects(t)

			rec := do(t, h, http.MethodPost, tc.path, tc.body)

			expectError(t, rec, http.StatusNotFound, tc.want)
		})
	}
}

func TestRegistrarDefecto_SprintCerradoDevuelve409(t *testing.T) {
	h := newHandlerWithDefects(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)

	rec := do(t, h, http.MethodPost, defectsPath, validDefect)

	expectError(t, rec, http.StatusConflict, sprint.ErrSprintClosed.Error())
}

func TestRegistrarDefecto_SprintDeOtroProyectoDevuelve409(t *testing.T) {
	h := newHandlerWithDefects(t)
	createProject(t, h, "Sistema de facturación")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/2/sprints", validSprint), http.StatusCreated)
	body := `{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":2}`

	rec := do(t, h, http.MethodPost, defectsPath, body)

	expectError(t, rec, http.StatusConflict, defect.ErrDetectionSprintFromAnotherProject.Error())
}

func TestRegistrarDefecto_CuerpoMayorA1MiBDa413(t *testing.T) {
	h := newHandlerWithDefects(t)
	body := "{" + strings.Repeat(" ", 1<<20) + validDefect[1:]

	rec := do(t, h, http.MethodPost, defectsPath, body)

	expectError(t, rec, http.StatusRequestEntityTooLarge, "el cuerpo del pedido supera el tamaño máximo de 1 MiB")

	// El cuerpo rechazado no se procesó: el siguiente registro recibe el ID 1.
	rec = do(t, h, http.MethodPost, defectsPath, validDefect)
	expectStatus(t, rec, http.StatusCreated)
	var got defectJSON
	decode(t, rec, &got)
	if got.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", got.ID)
	}
}

func TestRegistrarDefecto_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	h := newHandlerWithDefectRepository(t, brokenDefectRepository{})

	rec := do(t, h, http.MethodPost, defectsPath, validDefect)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// brokenDefectRepository siempre falla al guardar.
type brokenDefectRepository struct{}

func (brokenDefectRepository) Save(defect.Defect) (defect.Defect, error) {
	return defect.Defect{}, errStorage
}
