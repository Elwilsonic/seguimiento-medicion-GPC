package httpapi_test

import (
	"net/http"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// TEC-04 · Interfaz usable (Issue #34)
//
// Trazabilidad: CA-TEC04.9, la API expone las historias de sprints ya
// integradas en main:
//   - POST /api/proyectos/{id}/sprints expone HU-10 (crear sprint).
//   - POST /api/sprints/{id}/items/{itemId} expone HU-11 (asignar ítem).
//   - POST /api/sprints/{id}/items/{itemId}/completar y
//     GET /api/sprints/{id}/completados exponen HU-12 (completar ítems).
//   - POST /api/sprints/{id}/cerrar expone HU-13 (cerrar sprint).
// Por endpoint hay un caso exitoso y un test por cada código de error. Los
// errores que comparten código van como subtests: verifican la traducción a
// HTTP, no vuelven a probar las reglas de negocio de cada historia.

// sprintJSON es la respuesta de un sprint.
type sprintJSON struct {
	ID          int    `json:"id"`
	ProyectoID  int    `json:"proyecto_id"`
	Nombre      string `json:"nombre"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
	Objetivo    string `json:"objetivo"`
	Estado      string `json:"estado"`
}

const validSprint = `{"nombre":"Sprint 1","fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Gestionar el backlog"}`

// newHandlerWithSprint arma la API con el proyecto 1, su ítem 1 y el sprint 1
// abierto creado con validSprint.
func newHandlerWithSprint(t *testing.T) http.Handler {
	t.Helper()
	h := newHandlerWithItem(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint), http.StatusCreated)
	return h
}

// CA-TEC04.9 · POST /api/proyectos/{id}/sprints expone HU-10 (crear sprint).
func TestCrearSprint_Devuelve201ConElSprintCreado(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint)

	expectStatus(t, rec, http.StatusCreated)
	var s sprintJSON
	decode(t, rec, &s)
	want := sprintJSON{ID: 1, ProyectoID: 1, Nombre: "Sprint 1", FechaInicio: "2026-10-01",
		FechaFin: "2026-10-14", Objetivo: "Gestionar el backlog", Estado: "ABIERTO"}
	if s != want {
		t.Errorf("sprint = %+v, se esperaba %+v", s, want)
	}
}

func TestCrearSprint_DatosInvalidosDevuelven400(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"nombre vacío":        {`{"nombre":" ","fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Ok"}`, sprint.ErrEmptyName.Error()},
		"objetivo vacío":      {`{"nombre":"Sprint 1","fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":""}`, sprint.ErrEmptyGoal.Error()},
		"sin fecha de inicio": {`{"nombre":"Sprint 1","fecha_fin":"2026-10-14","objetivo":"Ok"}`, sprint.ErrMissingStartDate.Error()},
		"sin fecha de fin":    {`{"nombre":"Sprint 1","fecha_inicio":"2026-10-01","objetivo":"Ok"}`, sprint.ErrMissingEndDate.Error()},
		"fin antes de inicio": {`{"nombre":"Sprint 1","fecha_inicio":"2026-10-14","fecha_fin":"2026-10-01","objetivo":"Ok"}`, sprint.ErrEndNotAfterStart.Error()},
		"fecha de inicio con otro formato": {`{"nombre":"Sprint 1","fecha_inicio":"01/10/2026","fecha_fin":"2026-10-14","objetivo":"Ok"}`,
			"las fechas deben tener el formato AAAA-MM-DD"},
		"fecha de fin con otro formato": {`{"nombre":"Sprint 1","fecha_inicio":"2026-10-01","fecha_fin":"14/10/2026","objetivo":"Ok"}`,
			"las fechas deben tener el formato AAAA-MM-DD"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandler()
			createProject(t, h, "Sistema de ventas")

			rec := do(t, h, http.MethodPost, "/api/proyectos/1/sprints", tc.body)

			expectError(t, rec, http.StatusBadRequest, tc.want)
		})
	}
}

func TestCrearSprint_ProyectoInexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/api/proyectos/999/sprints", validSprint)

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

func TestCrearSprint_SprintAbiertoExistenteDevuelve409(t *testing.T) {
	h := newHandlerWithSprint(t)

	rec := do(t, h, http.MethodPost, "/api/proyectos/1/sprints", validSprint)

	expectError(t, rec, http.StatusConflict, sprint.ErrOpenSprintExists.Error())
}

// CA-TEC04.9 · POST /api/sprints/{id}/items/{itemId} expone HU-11 (asignar
// ítem).
func TestAsignarItem_Devuelve200ConElItemAsignado(t *testing.T) {
	h := newHandlerWithSprint(t)

	rec := do(t, h, http.MethodPost, "/api/sprints/1/items/1", "")

	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.ID != 1 || it.SprintID != 1 {
		t.Errorf("ítem = %+v, se esperaba el ítem 1 en el sprint 1", it)
	}
}

func TestAsignarItem_InexistenteDevuelve404(t *testing.T) {
	cases := map[string]struct {
		path string
		want error
	}{
		"sprint inexistente": {"/api/sprints/999/items/1", sprint.ErrSprintNotFound},
		"ítem inexistente":   {"/api/sprints/1/items/999", backlog.ErrItemNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithSprint(t)

			rec := do(t, h, http.MethodPost, tc.path, "")

			expectError(t, rec, http.StatusNotFound, tc.want.Error())
		})
	}
}

func TestAsignarItem_ConflictosDevuelven409(t *testing.T) {
	cases := map[string]struct {
		setup func(t *testing.T, h http.Handler)
		path  string
		want  error
	}{
		"sprint cerrado": {
			setup: func(t *testing.T, h http.Handler) {
				expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)
			},
			path: "/api/sprints/1/items/1",
			want: sprint.ErrSprintClosed,
		},
		"ítem de otro proyecto": {
			setup: func(t *testing.T, h http.Handler) {
				createProject(t, h, "Portal de clientes")
				expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/2/backlog", validItem), http.StatusCreated)
			},
			path: "/api/sprints/1/items/2",
			want: sprint.ErrItemFromAnotherProject,
		},
		"ítem ya asignado": {
			setup: func(t *testing.T, h http.Handler) {
				expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/items/1", ""), http.StatusOK)
			},
			path: "/api/sprints/1/items/1",
			want: backlog.ErrItemAlreadyAssigned,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithSprint(t)
			tc.setup(t, h)

			rec := do(t, h, http.MethodPost, tc.path, "")

			expectError(t, rec, http.StatusConflict, tc.want.Error())
		})
	}
}

// CA-TEC04.9 · POST /api/sprints/{id}/items/{itemId}/completar expone HU-12
// (completar ítem).
func TestCompletarItem_Devuelve200ConElItemTerminado(t *testing.T) {
	h := newHandlerWithSprint(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/items/1", ""), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/items/1/estado", `{"estado":"EN_CURSO"}`), http.StatusOK)

	rec := do(t, h, http.MethodPost, "/api/sprints/1/items/1/completar", "")

	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.ID != 1 || it.Estado != "TERMINADO" {
		t.Errorf("ítem = %+v, se esperaba el ítem 1 TERMINADO", it)
	}
}

func TestCompletarItem_InexistenteDevuelve404(t *testing.T) {
	cases := map[string]struct {
		path string
		want error
	}{
		"sprint inexistente": {"/api/sprints/999/items/1/completar", sprint.ErrSprintNotFound},
		"ítem inexistente":   {"/api/sprints/1/items/999/completar", backlog.ErrItemNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithSprint(t)

			rec := do(t, h, http.MethodPost, tc.path, "")

			expectError(t, rec, http.StatusNotFound, tc.want.Error())
		})
	}
}

func TestCompletarItem_ConflictosDevuelven409(t *testing.T) {
	cases := map[string]struct {
		setup func(t *testing.T, h http.Handler)
		want  error
	}{
		"sprint cerrado": {
			setup: func(t *testing.T, h http.Handler) {
				expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/items/1", ""), http.StatusOK)
				expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)
			},
			want: sprint.ErrSprintClosed,
		},
		"ítem no asignado al sprint": {
			setup: func(t *testing.T, h http.Handler) {},
			want:  sprint.ErrItemNotAssignedToSprint,
		},
		"transición no permitida": {
			setup: func(t *testing.T, h http.Handler) {
				expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/items/1", ""), http.StatusOK)
			},
			want: backlog.ErrInvalidTransition,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithSprint(t)
			tc.setup(t, h)

			rec := do(t, h, http.MethodPost, "/api/sprints/1/items/1/completar", "")

			expectError(t, rec, http.StatusConflict, tc.want.Error())
		})
	}
}

// CA-TEC04.9 · GET /api/sprints/{id}/completados expone HU-12 (listar los
// ítems completados).
func TestListarCompletados_Devuelve200ConLosItemsTerminados(t *testing.T) {
	h := newHandlerWithSprint(t)
	do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem)
	do(t, h, http.MethodPost, "/api/sprints/1/items/1", "")
	do(t, h, http.MethodPost, "/api/sprints/1/items/2", "")
	do(t, h, http.MethodPatch, "/api/items/2/estado", `{"estado":"EN_CURSO"}`)
	do(t, h, http.MethodPost, "/api/sprints/1/items/2/completar", "")

	rec := do(t, h, http.MethodGet, "/api/sprints/1/completados", "")

	expectStatus(t, rec, http.StatusOK)
	var items []itemJSON
	decode(t, rec, &items)
	if len(items) != 1 || items[0].ID != 2 || items[0].Estado != "TERMINADO" {
		t.Errorf("ítems = %+v, se esperaba solo el ítem 2 TERMINADO", items)
	}
}

func TestListarCompletados_SprintInexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/sprints/999/completados", "")

	expectError(t, rec, http.StatusNotFound, sprint.ErrSprintNotFound.Error())
}

// CA-TEC04.9 · POST /api/sprints/{id}/cerrar expone HU-13 (cerrar sprint).
func TestCerrarSprint_Devuelve200ConElSprintCerrado(t *testing.T) {
	h := newHandlerWithSprint(t)

	rec := do(t, h, http.MethodPost, "/api/sprints/1/cerrar", "")

	expectStatus(t, rec, http.StatusOK)
	var s sprintJSON
	decode(t, rec, &s)
	if s.ID != 1 || s.Nombre != "Sprint 1" || s.Estado != "CERRADO" {
		t.Errorf("sprint = %+v, se esperaba el sprint 1 CERRADO", s)
	}
}

func TestCerrarSprint_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/api/sprints/999/cerrar", "")

	expectError(t, rec, http.StatusNotFound, sprint.ErrSprintNotFound.Error())
}

func TestCerrarSprint_YaCerradoDevuelve409(t *testing.T) {
	h := newHandlerWithSprint(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)

	rec := do(t, h, http.MethodPost, "/api/sprints/1/cerrar", "")

	expectError(t, rec, http.StatusConflict, sprint.ErrSprintClosed.Error())
}

// HU-14 · Consultar sprints anteriores (Issue #28)
//
// GET /api/proyectos/{id}/sprints lista los sprints del proyecto, abiertos y
// cerrados, ordenados por ID, con el formato de toSprintResponse (CA-14.1 a
// CA-14.4).

// failingSprintListRepository falla al listar los sprints de un proyecto.
type failingSprintListRepository struct {
	sprint.Repository
	err error
}

func (r failingSprintListRepository) ListByProject(int) ([]sprint.Sprint, error) {
	return nil, r.err
}

// CA-14.1 · Devuelve el sprint cerrado y el abierto, en orden de ID y con el
// formato exacto de la creación.
func TestListarSprintsDelProyecto_Devuelve200ConLosSprintsOrdenados(t *testing.T) {
	h := newHandlerWithSprint(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/sprints/1/cerrar", ""), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/sprints",
		`{"nombre":"Sprint 2","fecha_inicio":"2026-10-15","fecha_fin":"2026-10-28","objetivo":"Métricas"}`), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/sprints", "")

	expectRawJSON(t, rec, `[
		{"id":1,"proyecto_id":1,"nombre":"Sprint 1","fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Gestionar el backlog","estado":"CERRADO"},
		{"id":2,"proyecto_id":1,"nombre":"Sprint 2","fecha_inicio":"2026-10-15","fecha_fin":"2026-10-28","objetivo":"Métricas","estado":"ABIERTO"}
	]`)
}

// CA-14.2 · Un proyecto sin sprints devuelve 200 y [], no null.
func TestListarSprintsDelProyecto_SinSprintsDevuelveListaVacia(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/sprints", "")

	expectRawJSON(t, rec, `[]`)
}

// CA-14.4 · Un {id} no numérico devuelve 400.
func TestListarSprintsDelProyecto_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos/abc/sprints", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-14.4 · Un proyecto inexistente, 0 o negativo devuelve 404.
func TestListarSprintsDelProyecto_ProyectoInexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodGet, "/api/proyectos/"+id+"/sprints", "")

		expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
	}
}

// RN-14.7 · Una falla al leer el proyecto o los sprints devuelve 500 sin
// detalle. Usa brokenRepository de TEC-01 para los proyectos.
func TestListarSprintsDelProyecto_FallaDeUnRepositorioDevuelve500(t *testing.T) {
	newHandlerWith := func(projectRepo project.Repository, sprintRepo sprint.Repository) http.Handler {
		projects := project.NewManager(projectRepo)
		items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
		sprints := sprint.NewManager(sprintRepo, projects, items)
		efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
		pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
		defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
		return httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)
	}
	projectsBroken := newHandlerWith(brokenRepository{}, sprint.NewMemoryRepository())
	sprintsBroken := newHandlerWith(project.NewMemoryRepository(),
		failingSprintListRepository{Repository: sprint.NewMemoryRepository(), err: errStorage})
	createProject(t, sprintsBroken, "Sistema de ventas")

	for _, h := range []http.Handler{projectsBroken, sprintsBroken} {
		rec := do(t, h, http.MethodGet, "/api/proyectos/1/sprints", "")

		expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
	}
}
