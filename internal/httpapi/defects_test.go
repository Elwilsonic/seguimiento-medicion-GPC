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
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))

	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	return h
}

// HU-29 · CA-29.11 · Verifica que POST /api/items/{id}/defectos responda 201 con el defecto ABIERTO y el sprint de resolución en 0.
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

// HU-29 · CA-29.3 · Verifica que la API ignore el estado y el sprint de resolución enviados y registre el defecto ABIERTO con 0.
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

// HU-29 · CA-29.2 · Verifica que una descripción vacía, una severidad inválida o un sprint de detección ausente, 0, negativo o null den 400.
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

// HU-29 · CA-29.11 · Verifica que un cuerpo que no es JSON válido dé 400.
func TestRegistrarDefecto_JSONInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, defectsPath, `{"descripcion":`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

// HU-29 · CA-29.11 · Verifica que un sprint de detección decimal o de texto dé 400 por JSON inválido.
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

// HU-29 · CA-29.11 · Verifica que texto después del JSON dé 400 por JSON inválido.
func TestRegistrarDefecto_TextoDespuesDelJSONDevuelve400(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, defectsPath, validDefect+` basura`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

// HU-29 · CA-29.11 · Verifica que un {id} de historia que no es un número entero dé 400.
func TestRegistrarDefecto_ItemIDInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, "/api/items/abc/defectos", validDefect)

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// HU-29 · CA-29.11 · Verifica que una historia o un sprint de detección inexistentes den 404.
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

// HU-29 · CA-29.6 · Verifica que registrar un defecto en un sprint de detección CERRADO dé 409.
func TestRegistrarDefecto_SprintCerradoDevuelve409(t *testing.T) {
	h := newHandlerWithDefects(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)

	rec := do(t, h, http.MethodPost, defectsPath, validDefect)

	expectError(t, rec, http.StatusConflict, sprint.ErrSprintClosed.Error())
}

// HU-29 · CA-29.6 · Verifica que registrar un defecto con un sprint de detección de otro proyecto dé 409.
func TestRegistrarDefecto_SprintDeOtroProyectoDevuelve409(t *testing.T) {
	h := newHandlerWithDefects(t)
	createProject(t, h, "Sistema de facturación")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/2/sprints", validSprint), http.StatusCreated)
	body := `{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":2}`

	rec := do(t, h, http.MethodPost, defectsPath, body)

	expectError(t, rec, http.StatusConflict, defect.ErrDetectionSprintFromAnotherProject.Error())
}

// HU-29 · CA-29.11 · Verifica que un cuerpo mayor a 1 MiB dé 413 y no se procese.
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

// HU-29 · CA-29.11 · Verifica que una falla del repositorio dé 500 sin detalle.
func TestRegistrarDefecto_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	h := newHandlerWithDefectRepository(t, brokenDefectRepository{})

	rec := do(t, h, http.MethodPost, defectsPath, validDefect)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// brokenDefectRepository siempre falla al guardar, al buscar y al actualizar.
type brokenDefectRepository struct{}

func (brokenDefectRepository) Save(defect.Defect) (defect.Defect, error) {
	return defect.Defect{}, errStorage
}

func (brokenDefectRepository) GetByID(int) (defect.Defect, error) {
	return defect.Defect{}, errStorage
}

func (brokenDefectRepository) Update(defect.Defect) (defect.Defect, error) {
	return defect.Defect{}, errStorage
}

// HU-30 · Cambiar el estado de un defecto por HTTP (Issue #48)
//
// Trazabilidad: CA-30.9. PATCH /api/defectos/{id}/estado expone el cambio de
// estado. Hay un caso exitoso y un test por cada código de error de la tabla
// HTTP de la SDD (400, 404, 409, 413 y 500). El 409 por un sprint de
// resolución anterior al de detección no se puede producir desde la API
// (validación defensiva, RN-30.6): se prueba en internal/defect.

const defectStatusPath = "/api/defectos/1/estado"

// newHandlerWithDefect arma la API de newHandlerWithDefects y registra el
// defecto 1 (ABIERTO, detectado en el sprint 1) con el POST de HU-29.
func newHandlerWithDefect(t *testing.T) http.Handler {
	t.Helper()
	h := newHandlerWithDefects(t)
	expectStatus(t, do(t, h, http.MethodPost, defectsPath, validDefect), http.StatusCreated)
	return h
}

// changeDefectStatus hace el PATCH y exige un 200.
func changeDefectStatus(t *testing.T, h http.Handler, body string) defectJSON {
	t.Helper()
	rec := do(t, h, http.MethodPatch, defectStatusPath, body)
	expectStatus(t, rec, http.StatusOK)
	var got defectJSON
	decode(t, rec, &got)
	return got
}

// HU-30 · CA-30.9 · Verifica que PATCH /api/defectos/{id}/estado responda 200 con el defecto actualizado.
func TestCambiarEstadoDefecto_Devuelve200ConElDefecto(t *testing.T) {
	h := newHandlerWithDefect(t)

	got := changeDefectStatus(t, h, `{"estado":" en_correccion "}`)
	want := defectJSON{
		ID:                 1,
		Descripcion:        "El total no descuenta el descuento",
		Severidad:          defect.SeverityHigh,
		Estado:             defect.StatusInCorrection,
		ItemID:             1,
		SprintDeteccionID:  1,
		SprintResolucionID: 0,
	}
	if got != want {
		t.Errorf("defecto = %+v, se esperaba %+v", got, want)
	}

	got = changeDefectStatus(t, h, `{"estado":"RESUELTO","sprint_resolucion_id":1}`)
	want.Estado = defect.StatusResolved
	want.SprintResolucionID = 1
	if got != want {
		t.Errorf("defecto = %+v, se esperaba %+v", got, want)
	}
}

// HU-30 · CA-30.2 · Verifica que reabrir un defecto por la API devuelva el sprint de resolución en 0.
func TestCambiarEstadoDefecto_ReabrirDevuelveElSprintDeResolucionEnCero(t *testing.T) {
	h := newHandlerWithDefect(t)
	changeDefectStatus(t, h, `{"estado":"EN_CORRECCION"}`)
	changeDefectStatus(t, h, `{"estado":"RESUELTO","sprint_resolucion_id":1}`)

	rec := do(t, h, http.MethodPatch, defectStatusPath, `{"estado":"EN_CORRECCION","sprint_resolucion_id":1}`)

	expectStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), `"sprint_resolucion_id":0`) {
		t.Errorf("la respuesta no tiene el sprint de resolución en 0: %s", rec.Body.String())
	}
}

// HU-30 · CA-30.9 · Verifica que un estado inválido, un sprint de resolución ausente o inválido o un JSON inválido den 400.
func TestCambiarEstadoDefecto_DatosInvalidosDevuelven400(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"estado vacío":                  {`{"estado":"   "}`, defect.ErrInvalidStatus.Error()},
		"estado inválido":               {`{"estado":"CERRADO"}`, defect.ErrInvalidStatus.Error()},
		"sprint de resolución ausente":  {`{"estado":"RESUELTO"}`, defect.ErrMissingResolutionSprint.Error()},
		"sprint de resolución cero":     {`{"estado":"RESUELTO","sprint_resolucion_id":0}`, defect.ErrMissingResolutionSprint.Error()},
		"sprint de resolución negativo": {`{"estado":"RESUELTO","sprint_resolucion_id":-1}`, defect.ErrMissingResolutionSprint.Error()},
		"sprint de resolución null":     {`{"estado":"RESUELTO","sprint_resolucion_id":null}`, defect.ErrMissingResolutionSprint.Error()},
		"cuerpo que no es JSON":         {`{"estado":`, "el cuerpo del pedido no es un JSON válido"},
		"sprint de resolución decimal":  {`{"estado":"RESUELTO","sprint_resolucion_id":1.5}`, "el cuerpo del pedido no es un JSON válido"},
		"sprint de resolución texto":    {`{"estado":"RESUELTO","sprint_resolucion_id":"abc"}`, "el cuerpo del pedido no es un JSON válido"},
		"texto después del JSON":        {`{"estado":"EN_CORRECCION"} basura`, "el cuerpo del pedido no es un JSON válido"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithDefect(t)

			rec := do(t, h, http.MethodPatch, defectStatusPath, tc.body)

			expectError(t, rec, http.StatusBadRequest, tc.want)
		})
	}
}

// HU-30 · CA-30.9 · Verifica que un {id} de defecto que no es un número entero dé 400.
func TestCambiarEstadoDefecto_IDInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithDefect(t)

	rec := do(t, h, http.MethodPatch, "/api/defectos/abc/estado", `{"estado":"EN_CORRECCION"}`)

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// HU-30 · CA-30.9 · Verifica que un defecto o un sprint de resolución inexistentes den 404.
func TestCambiarEstadoDefecto_InexistenteDevuelve404(t *testing.T) {
	cases := map[string]struct {
		path string
		body string
		want string
	}{
		"defecto inexistente": {
			"/api/defectos/999/estado",
			`{"estado":"EN_CORRECCION"}`,
			defect.ErrDefectNotFound.Error(),
		},
		"sprint de resolución inexistente": {
			defectStatusPath,
			`{"estado":"RESUELTO","sprint_resolucion_id":999}`,
			sprint.ErrSprintNotFound.Error(),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithDefect(t)
			changeDefectStatus(t, h, `{"estado":"EN_CORRECCION"}`)

			rec := do(t, h, http.MethodPatch, tc.path, tc.body)

			expectError(t, rec, http.StatusNotFound, tc.want)
		})
	}
}

// HU-30 · CA-30.1 · Verifica que una transición no permitida (ABIERTO → RESUELTO) dé 409.
func TestCambiarEstadoDefecto_TransicionNoPermitidaDevuelve409(t *testing.T) {
	h := newHandlerWithDefect(t)

	rec := do(t, h, http.MethodPatch, defectStatusPath, `{"estado":"RESUELTO","sprint_resolucion_id":1}`)

	expectError(t, rec, http.StatusConflict, defect.ErrInvalidTransition.Error())
}

// HU-30 · CA-30.2 · Verifica que resolver en un sprint CERRADO dé 409.
func TestCambiarEstadoDefecto_SprintCerradoDevuelve409(t *testing.T) {
	h := newHandlerWithDefect(t)
	changeDefectStatus(t, h, `{"estado":"EN_CORRECCION"}`)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)

	rec := do(t, h, http.MethodPatch, defectStatusPath, `{"estado":"RESUELTO","sprint_resolucion_id":1}`)

	expectError(t, rec, http.StatusConflict, sprint.ErrSprintClosed.Error())
}

// HU-30 · CA-30.2 · Verifica que resolver en un sprint de otro proyecto dé 409.
func TestCambiarEstadoDefecto_SprintDeOtroProyectoDevuelve409(t *testing.T) {
	h := newHandlerWithDefect(t)
	changeDefectStatus(t, h, `{"estado":"EN_CORRECCION"}`)
	createProject(t, h, "Sistema de facturación")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/2/sprints", validSprint), http.StatusCreated)

	rec := do(t, h, http.MethodPatch, defectStatusPath, `{"estado":"RESUELTO","sprint_resolucion_id":2}`)

	expectError(t, rec, http.StatusConflict, defect.ErrResolutionSprintFromAnotherProject.Error())
}

// HU-30 · CA-30.9 · Verifica que un cuerpo mayor a 1 MiB dé 413 y no cambie el estado.
func TestCambiarEstadoDefecto_CuerpoMayorA1MiBDa413(t *testing.T) {
	h := newHandlerWithDefect(t)
	body := `{"estado":"EN_CORRECCION"` + strings.Repeat(" ", 1<<20) + `}`

	rec := do(t, h, http.MethodPatch, defectStatusPath, body)

	expectError(t, rec, http.StatusRequestEntityTooLarge, "el cuerpo del pedido supera el tamaño máximo de 1 MiB")

	// El cuerpo rechazado no se procesó: el defecto sigue ABIERTO y puede
	// pasar a EN_CORRECCION.
	if got := changeDefectStatus(t, h, `{"estado":"EN_CORRECCION"}`); got.Estado != defect.StatusInCorrection {
		t.Errorf("estado = %q, se esperaba %q", got.Estado, defect.StatusInCorrection)
	}
}

// HU-30 · CA-30.9 · Verifica que una falla del repositorio dé 500 sin detalle.
func TestCambiarEstadoDefecto_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	h := newHandlerWithDefectRepository(t, brokenDefectRepository{})

	rec := do(t, h, http.MethodPatch, defectStatusPath, `{"estado":"EN_CORRECCION"}`)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
