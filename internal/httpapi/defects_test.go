package httpapi_test

import (
	"uuid"

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

// HU-29 · Registrar defecto por HTTP (Issue #47)
//
// Trazabilidad: CA-29.11. POST /api/items/{id}/defectos expone el registro de
// defectos. Hay un caso exitoso y un test por cada código de error de la tabla
// HTTP de la SDD (400, 404, 409, 413 y 500).

// HU-52, CA-52.3: el defecto, su ítem y sus sprints se identifican con su
// UUID, como texto, y un defecto sin resolver tiene "" en
// sprint_resolucion_id (D7).
type defectJSON struct {
	ID                 string          `json:"id"`
	Descripcion        string          `json:"descripcion"`
	Severidad          defect.Severity `json:"severidad"`
	Estado             defect.Status   `json:"estado"`
	ItemID             string          `json:"item_id"`
	SprintDeteccionID  string          `json:"sprint_deteccion_id"`
	SprintResolucionID string          `json:"sprint_resolucion_id"`
}

// validDefect tiene el sprint de detección como número, la forma anterior a
// HU-52. Los tests de defectos usan validDefectFor, que lo manda como texto.
const validDefect = `{"descripcion":"  El total no descuenta el descuento  ","severidad":" alta ","sprint_deteccion_id":1}`

const defectsPath = "/api/items/1/defectos"

// apiID devuelve el "id" que devolvió la API al crear la n-ésima entidad de
// ese tipo en h, el mismo que do() pone en las rutas (D9 de HU-52). Sirve
// para los IDs de los cuerpos y de las respuestas esperadas.
func apiID(t *testing.T, h http.Handler, kind entityKind, n int) string {
	t.Helper()
	entityIDsMu.Lock()
	defer entityIDsMu.Unlock()
	ids := entityIDs[h][kind]
	if n < 1 || n > len(ids) {
		t.Fatalf("la API no creó el %s %d (creó %d)", kind, n, len(ids))
	}
	return ids[n-1]
}

// apiIDJSON es apiID como texto JSON (entre comillas), para armar las
// respuestas esperadas: HU-52 identifica cada defecto y esfuerzo con su UUID,
// que es texto (CA-52.3).
func apiIDJSON(t *testing.T, h http.Handler, kind entityKind, n int) string {
	t.Helper()
	return strconv.Quote(apiID(t, h, kind, n))
}

// validDefectFor es validDefect con el sprint 1 de h como sprint de
// detección, como texto (HU-52, CA-52.3).
func validDefectFor(t *testing.T, h http.Handler) string {
	t.Helper()
	return validDefectIn(apiID(t, h, kindSprint, 1))
}

// validDefectIn es validDefect con sprintID como sprint de detección.
func validDefectIn(sprintID string) string {
	return `{"descripcion":"  El total no descuenta el descuento  ","severidad":" alta ","sprint_deteccion_id":"` + sprintID + `"}`
}

// resolveIn es el cuerpo que resuelve un defecto en el sprint sprintID.
func resolveIn(sprintID string) string {
	return `{"estado":"RESUELTO","sprint_resolucion_id":"` + sprintID + `"}`
}

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
	expectStatus(t, do(t, h, http.MethodPost, projectPath(t, h, 1)+"/backlog", validItem), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, projectPath(t, h, 1)+"/sprints", validSprint), http.StatusCreated)
	return h
}

// HU-52, CA-52.3: el id es un UUID, el ítem y el sprint de detección son sus
// UUID, y el sprint de resolución sin resolver es "" en lugar de 0 (D7).
// HU-29 · CA-29.11 · Verifica que POST /api/items/{id}/defectos responda 201 con el defecto ABIERTO, su UUID en id y el sprint de resolución vacío.
func TestRegistrarDefecto_Devuelve201ConElDefecto(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, defectsPath, validDefectFor(t, h))

	expectStatus(t, rec, http.StatusCreated)
	var got defectJSON
	decode(t, rec, &got)
	expectUUIDv4(t, "id", got.ID)
	want := defectJSON{
		ID:                 got.ID,
		Descripcion:        "El total no descuenta el descuento",
		Severidad:          defect.SeverityHigh,
		Estado:             defect.StatusOpen,
		ItemID:             apiID(t, h, kindItem, 1),
		SprintDeteccionID:  apiID(t, h, kindSprint, 1),
		SprintResolucionID: "",
	}
	if got != want {
		t.Errorf("defecto = %+v, se esperaba %+v", got, want)
	}
	if !strings.Contains(rec.Body.String(), `"sprint_resolucion_id":""`) {
		t.Errorf("la respuesta no incluye el sprint de resolución vacío: %s", rec.Body.String())
	}
}

// HU-29 · CA-29.3 · Verifica que la API ignore el estado y el sprint de resolución enviados y registre el defecto ABIERTO, sin sprint de resolución.
func TestRegistrarDefecto_IgnoraEstadoYSprintDeResolucion(t *testing.T) {
	h := newHandlerWithDefects(t)
	sprintID := apiID(t, h, kindSprint, 1)
	body := `{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":"` + sprintID + `","estado":"RESUELTO","sprint_resolucion_id":"` + sprintID + `"}`

	rec := do(t, h, http.MethodPost, defectsPath, body)

	expectStatus(t, rec, http.StatusCreated)
	var got defectJSON
	decode(t, rec, &got)
	if got.Estado != defect.StatusOpen {
		t.Errorf("estado = %q, se esperaba %q", got.Estado, defect.StatusOpen)
	}
	if got.SprintResolucionID != "" {
		t.Errorf("sprint de resolución = %q, se esperaba vacío", got.SprintResolucionID)
	}
}

// HU-52, CA-52.4: un sprint de detección vacío ("") sigue siendo un sprint
// faltante. El 0 y los negativos ya no son un sprint faltante: un número no
// es un texto (TestRegistrarDefecto_SprintDeDeteccionNoEnteroDevuelve400).
// HU-29 · CA-29.2 · Verifica que una descripción vacía, una severidad inválida o un sprint de detección ausente, vacío o null den 400.
func TestRegistrarDefecto_DatosInvalidosDevuelven400(t *testing.T) {
	cases := map[string]struct {
		body func(sprintID string) string
		want string
	}{
		"descripción vacía": {
			func(s string) string {
				return `{"descripcion":"   ","severidad":"ALTA","sprint_deteccion_id":"` + s + `"}`
			},
			defect.ErrEmptyDescription.Error(),
		},
		"severidad inválida": {
			func(s string) string {
				return `{"descripcion":"El total no descuenta el descuento","severidad":"URGENTE","sprint_deteccion_id":"` + s + `"}`
			},
			defect.ErrInvalidSeverity.Error(),
		},
		"sprint de detección ausente": {
			func(string) string {
				return `{"descripcion":"El total no descuenta el descuento","severidad":"ALTA"}`
			},
			defect.ErrMissingDetectionSprint.Error(),
		},
		"sprint de detección vacío": {
			func(string) string {
				return `{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":""}`
			},
			defect.ErrMissingDetectionSprint.Error(),
		},
		"sprint de detección null": {
			func(string) string {
				return `{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":null}`
			},
			defect.ErrMissingDetectionSprint.Error(),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithDefects(t)

			rec := do(t, h, http.MethodPost, defectsPath, tc.body(apiID(t, h, kindSprint, 1)))

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

// HU-52, CA-52.4: el sprint de detección es un texto. Un texto que no es un
// UUID da el 400 del sprint, y un número (entero, 0, negativo o decimal) no es
// un texto, así que el cuerpo no es un JSON válido.
// HU-29 · CA-29.11 · Verifica que un sprint de detección numérico dé 400 por JSON inválido, y uno de texto que no es UUID, el 400 del sprint.
func TestRegistrarDefecto_SprintDeDeteccionNoEnteroDevuelve400(t *testing.T) {
	cases := map[string]struct{ body, want string }{
		"entero": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":1}`,
			"el cuerpo del pedido no es un JSON válido",
		},
		"cero": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":0}`,
			"el cuerpo del pedido no es un JSON válido",
		},
		"negativo": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":-1}`,
			"el cuerpo del pedido no es un JSON válido",
		},
		"decimal": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":1.5}`,
			"el cuerpo del pedido no es un JSON válido",
		},
		"texto": {
			`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":"abc"}`,
			"el id del sprint debe ser un UUID",
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

// HU-29 · CA-29.11 · Verifica que texto después del JSON dé 400 por JSON inválido.
func TestRegistrarDefecto_TextoDespuesDelJSONDevuelve400(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, defectsPath, validDefectFor(t, h)+` basura`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

// HU-52, CA-52.4: el {id} de la historia es un UUID.
// HU-29 · CA-29.11 · Verifica que un {id} de historia que no es un UUID dé 400.
func TestRegistrarDefecto_ItemIDInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithDefects(t)

	rec := do(t, h, http.MethodPost, "/api/items/abc/defectos", validDefectFor(t, h))

	expectError(t, rec, http.StatusBadRequest, "el id del ítem debe ser un UUID")
}

// HU-52, CA-52.4: una historia o un sprint inexistentes son UUID que nunca se
// asignaron.
// HU-29 · CA-29.11 · Verifica que una historia o un sprint de detección inexistentes den 404.
func TestRegistrarDefecto_InexistenteDevuelve404(t *testing.T) {
	cases := map[string]struct {
		path string
		body func(h http.Handler) string
		want string
	}{
		"historia inexistente": {
			"/api/items/" + neverAssignedUUID + "/defectos",
			func(h http.Handler) string { return validDefectFor(t, h) },
			backlog.ErrItemNotFound.Error(),
		},
		"sprint de detección inexistente": {
			defectsPath,
			func(http.Handler) string { return validDefectIn(neverAssignedUUID) },
			sprint.ErrSprintNotFound.Error(),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithDefects(t)

			rec := do(t, h, http.MethodPost, tc.path, tc.body(h))

			expectError(t, rec, http.StatusNotFound, tc.want)
		})
	}
}

// HU-29 · CA-29.6 · Verifica que registrar un defecto en un sprint de detección CERRADO dé 409.
func TestRegistrarDefecto_SprintCerradoDevuelve409(t *testing.T) {
	h := newHandlerWithDefects(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)

	rec := do(t, h, http.MethodPost, defectsPath, validDefectFor(t, h))

	expectError(t, rec, http.StatusConflict, sprint.ErrSprintClosed.Error())
}

// HU-29 · CA-29.6 · Verifica que registrar un defecto con un sprint de detección de otro proyecto dé 409.
func TestRegistrarDefecto_SprintDeOtroProyectoDevuelve409(t *testing.T) {
	h := newHandlerWithDefects(t)
	createProject(t, h, "Sistema de facturación")
	expectStatus(t, do(t, h, http.MethodPost, projectPath(t, h, 2)+"/sprints", validSprint), http.StatusCreated)

	rec := do(t, h, http.MethodPost, defectsPath, validDefectIn(apiID(t, h, kindSprint, 2)))

	expectError(t, rec, http.StatusConflict, defect.ErrDetectionSprintFromAnotherProject.Error())
}

// HU-52, CA-52.1: el id ya no dice cuántos defectos hay, así que se cuenta
// con la lista del sprint.
// HU-29 · CA-29.11 · Verifica que un cuerpo mayor a 1 MiB dé 413 y no se procese.
func TestRegistrarDefecto_CuerpoMayorA1MiBDa413(t *testing.T) {
	h := newHandlerWithDefects(t)
	valid := validDefectFor(t, h)
	body := "{" + strings.Repeat(" ", 1<<20) + valid[1:]

	rec := do(t, h, http.MethodPost, defectsPath, body)

	expectError(t, rec, http.StatusRequestEntityTooLarge, "el cuerpo del pedido supera el tamaño máximo de 1 MiB")

	// El cuerpo rechazado no se procesó: después del siguiente registro, el
	// sprint tiene un solo defecto.
	expectStatus(t, do(t, h, http.MethodPost, defectsPath, valid), http.StatusCreated)
	list := do(t, h, http.MethodGet, "/api/sprints/1/defectos", "")
	expectStatus(t, list, http.StatusOK)
	if !strings.Contains(list.Body.String(), `"cantidad":1`) {
		t.Errorf("defectos del sprint = %s, se esperaba uno solo", list.Body.String())
	}
}

// HU-29 · CA-29.11 · Verifica que una falla del repositorio dé 500 sin detalle.
func TestRegistrarDefecto_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	h := newHandlerWithDefectRepository(t, brokenDefectRepository{})

	rec := do(t, h, http.MethodPost, defectsPath, validDefectFor(t, h))

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
	expectStatus(t, do(t, h, http.MethodPost, defectsPath, validDefectFor(t, h)), http.StatusCreated)
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

// HU-52, CA-52.3: el defecto, su ítem y sus sprints se identifican con su
// UUID, y el sprint de resolución se manda como texto.
// HU-30 · CA-30.9 · Verifica que PATCH /api/defectos/{id}/estado responda 200 con el defecto actualizado.
func TestCambiarEstadoDefecto_Devuelve200ConElDefecto(t *testing.T) {
	h := newHandlerWithDefect(t)
	sprintID := apiID(t, h, kindSprint, 1)

	got := changeDefectStatus(t, h, `{"estado":" en_correccion "}`)
	want := defectJSON{
		ID:                 apiID(t, h, kindDefect, 1),
		Descripcion:        "El total no descuenta el descuento",
		Severidad:          defect.SeverityHigh,
		Estado:             defect.StatusInCorrection,
		ItemID:             apiID(t, h, kindItem, 1),
		SprintDeteccionID:  sprintID,
		SprintResolucionID: "",
	}
	if got != want {
		t.Errorf("defecto = %+v, se esperaba %+v", got, want)
	}

	got = changeDefectStatus(t, h, resolveIn(sprintID))
	want.Estado = defect.StatusResolved
	want.SprintResolucionID = sprintID
	if got != want {
		t.Errorf("defecto = %+v, se esperaba %+v", got, want)
	}
}

// HU-52, CA-52.3: sin resolver, el sprint de resolución es "" en lugar de 0
// (D7).
// HU-30 · CA-30.2 · Verifica que reabrir un defecto por la API devuelva el sprint de resolución vacío.
func TestCambiarEstadoDefecto_ReabrirDevuelveElSprintDeResolucionEnCero(t *testing.T) {
	h := newHandlerWithDefect(t)
	sprintID := apiID(t, h, kindSprint, 1)
	changeDefectStatus(t, h, `{"estado":"EN_CORRECCION"}`)
	changeDefectStatus(t, h, resolveIn(sprintID))

	rec := do(t, h, http.MethodPatch, defectStatusPath, `{"estado":"EN_CORRECCION","sprint_resolucion_id":"`+sprintID+`"}`)

	expectStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), `"sprint_resolucion_id":""`) {
		t.Errorf("la respuesta no tiene el sprint de resolución vacío: %s", rec.Body.String())
	}
}

// HU-52, CA-52.4: el sprint de resolución es un texto. Vacío o null sigue
// siendo un sprint faltante; un número no es un texto (JSON inválido), y un
// texto que no es un UUID da el 400 del sprint.
// HU-30 · CA-30.9 · Verifica que un estado inválido, un sprint de resolución ausente o inválido o un JSON inválido den 400.
func TestCambiarEstadoDefecto_DatosInvalidosDevuelven400(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"estado vacío":                  {`{"estado":"   "}`, defect.ErrInvalidStatus.Error()},
		"estado inválido":               {`{"estado":"CERRADO"}`, defect.ErrInvalidStatus.Error()},
		"sprint de resolución ausente":  {`{"estado":"RESUELTO"}`, defect.ErrMissingResolutionSprint.Error()},
		"sprint de resolución vacío":    {`{"estado":"RESUELTO","sprint_resolucion_id":""}`, defect.ErrMissingResolutionSprint.Error()},
		"sprint de resolución null":     {`{"estado":"RESUELTO","sprint_resolucion_id":null}`, defect.ErrMissingResolutionSprint.Error()},
		"cuerpo que no es JSON":         {`{"estado":`, "el cuerpo del pedido no es un JSON válido"},
		"sprint de resolución entero":   {`{"estado":"RESUELTO","sprint_resolucion_id":1}`, "el cuerpo del pedido no es un JSON válido"},
		"sprint de resolución cero":     {`{"estado":"RESUELTO","sprint_resolucion_id":0}`, "el cuerpo del pedido no es un JSON válido"},
		"sprint de resolución negativo": {`{"estado":"RESUELTO","sprint_resolucion_id":-1}`, "el cuerpo del pedido no es un JSON válido"},
		"sprint de resolución decimal":  {`{"estado":"RESUELTO","sprint_resolucion_id":1.5}`, "el cuerpo del pedido no es un JSON válido"},
		"sprint de resolución texto":    {`{"estado":"RESUELTO","sprint_resolucion_id":"abc"}`, "el id del sprint debe ser un UUID"},
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

// HU-52, CA-52.4: el {id} del defecto es un UUID.
// HU-30 · CA-30.9 · Verifica que un {id} de defecto que no es un UUID dé 400.
func TestCambiarEstadoDefecto_IDInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithDefect(t)

	rec := do(t, h, http.MethodPatch, "/api/defectos/abc/estado", `{"estado":"EN_CORRECCION"}`)

	expectError(t, rec, http.StatusBadRequest, "el id del defecto debe ser un UUID")
}

// HU-52, CA-52.4: un defecto o un sprint inexistentes son UUID que nunca se
// asignaron.
// HU-30 · CA-30.9 · Verifica que un defecto o un sprint de resolución inexistentes den 404.
func TestCambiarEstadoDefecto_InexistenteDevuelve404(t *testing.T) {
	cases := map[string]struct {
		path string
		body string
		want string
	}{
		"defecto inexistente": {
			"/api/defectos/" + neverAssignedUUID + "/estado",
			`{"estado":"EN_CORRECCION"}`,
			defect.ErrDefectNotFound.Error(),
		},
		"sprint de resolución inexistente": {
			defectStatusPath,
			resolveIn(neverAssignedUUID),
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

	rec := do(t, h, http.MethodPatch, defectStatusPath, resolveIn(apiID(t, h, kindSprint, 1)))

	expectError(t, rec, http.StatusConflict, defect.ErrInvalidTransition.Error())
}

// HU-30 · CA-30.2 · Verifica que resolver en un sprint CERRADO dé 409.
func TestCambiarEstadoDefecto_SprintCerradoDevuelve409(t *testing.T) {
	h := newHandlerWithDefect(t)
	changeDefectStatus(t, h, `{"estado":"EN_CORRECCION"}`)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)

	rec := do(t, h, http.MethodPatch, defectStatusPath, resolveIn(apiID(t, h, kindSprint, 1)))

	expectError(t, rec, http.StatusConflict, sprint.ErrSprintClosed.Error())
}

// HU-30 · CA-30.2 · Verifica que resolver en un sprint de otro proyecto dé 409.
func TestCambiarEstadoDefecto_SprintDeOtroProyectoDevuelve409(t *testing.T) {
	h := newHandlerWithDefect(t)
	changeDefectStatus(t, h, `{"estado":"EN_CORRECCION"}`)
	createProject(t, h, "Sistema de facturación")
	expectStatus(t, do(t, h, http.MethodPost, projectPath(t, h, 2)+"/sprints", validSprint), http.StatusCreated)

	rec := do(t, h, http.MethodPatch, defectStatusPath, resolveIn(apiID(t, h, kindSprint, 2)))

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

// HU-52, CA-52.4: con el repositorio roto no se crea ningún defecto, así que
// la ruta lleva un UUID que no existe y la falla aparece al buscarlo.
// HU-30 · CA-30.9 · Verifica que una falla del repositorio dé 500 sin detalle.
func TestCambiarEstadoDefecto_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	h := newHandlerWithDefectRepository(t, brokenDefectRepository{})

	rec := do(t, h, http.MethodPatch, "/api/defectos/"+neverAssignedUUID+"/estado", `{"estado":"EN_CORRECCION"}`)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// GetByPublicID completa la interfaz de HU-52; este falso no lo usa.
func (brokenDefectRepository) GetByPublicID(uuid.UUID) (defect.Defect, error) {
	return defect.Defect{}, defect.ErrDefectNotFound
}
