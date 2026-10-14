package httpapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/project"
)

// TEC-01 · Punto de entrada ejecutable mínimo (Issue #23)
//
// Trazabilidad: cada endpoint expone una historia ya integrada en main y cita
// su criterio de TEC-01:
//   - CA-TEC01.2: los endpoints de proyectos (HU-01, HU-02, HU-03) y de
//     backlog (HU-04, HU-06).
//   - CA-TEC01.3: estos tests son las pruebas automatizadas de los endpoints.
//   - CA-TEC01.5: los datos se guardan en los repositorios en memoria.

// projectJSON y itemJSON son las respuestas que devuelve la API.
type projectJSON struct {
	ID          int    `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
}

type itemJSON struct {
	ID          int      `json:"id"`
	ProyectoID  int      `json:"proyecto_id"`
	Titulo      string   `json:"titulo"`
	Descripcion string   `json:"descripcion"`
	Prioridad   string   `json:"prioridad"`
	Criterios   []string `json:"criterios"`
	Estado      string   `json:"estado"`
	StoryPoints int      `json:"story_points"`
}

type errorJSON struct {
	Error string `json:"error"`
}

// newHandler arma la API con los repositorios en memoria.
func newHandler() http.Handler {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	return httpapi.NewHandler(projects, items)
}

// do hace un pedido a la API y verifica que la respuesta sea JSON.
func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("%s %s: Content-Type = %q, se esperaba application/json", method, path, ct)
	}
	return rec
}

// decode lee el cuerpo JSON de la respuesta en v.
func decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("cuerpo no es JSON válido (%v): %s", err, rec.Body.String())
	}
}

// expectStatus verifica el código HTTP de la respuesta.
func expectStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("código = %d, se esperaba %d; cuerpo: %s", rec.Code, want, rec.Body.String())
	}
}

// expectError verifica el código HTTP y que el cuerpo tenga el mensaje de error.
func expectError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantMsg string) {
	t.Helper()
	expectStatus(t, rec, wantStatus)
	var body errorJSON
	decode(t, rec, &body)
	if body.Error != wantMsg {
		t.Errorf("error = %q, se esperaba %q", body.Error, wantMsg)
	}
}

// createProject crea un proyecto por la API y devuelve su ID.
func createProject(t *testing.T, h http.Handler, name string) int {
	t.Helper()
	rec := do(t, h, http.MethodPost, "/proyectos", `{"nombre":"`+name+`","descripcion":""}`)
	expectStatus(t, rec, http.StatusCreated)
	var p projectJSON
	decode(t, rec, &p)
	return p.ID
}

const validItem = `{"titulo":"Login de usuarios","descripcion":"Acceso","prioridad":"alta","criterios":["Credenciales válidas"]}`

// CA-TEC01.2 · POST /proyectos expone HU-01 (crear proyecto).
func TestCrearProyecto_Devuelve201ConElProyectoCreado(t *testing.T) {
	h := newHandler()

	rec := do(t, h, http.MethodPost, "/proyectos", `{"nombre":"  Sistema de ventas  ","descripcion":"Gestión de ventas"}`)

	expectStatus(t, rec, http.StatusCreated)
	var p projectJSON
	decode(t, rec, &p)
	want := projectJSON{ID: 1, Nombre: "Sistema de ventas", Descripcion: "Gestión de ventas", Estado: "PLANIFICADO"}
	if p != want {
		t.Errorf("proyecto = %+v, se esperaba %+v", p, want)
	}
}

func TestCrearProyecto_NombreVacioDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/proyectos", `{"nombre":"   ","descripcion":""}`)

	expectError(t, rec, http.StatusBadRequest, project.ErrEmptyName.Error())
}

func TestCrearProyecto_NombreRepetidoDevuelve409(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPost, "/proyectos", `{"nombre":"SISTEMA DE VENTAS","descripcion":""}`)

	expectError(t, rec, http.StatusConflict, project.ErrDuplicateName.Error())
}

func TestCrearProyecto_JSONInvalidoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/proyectos", `{"nombre":`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

// CA-TEC01.2 · GET /proyectos/{id} expone HU-03 (consultar proyecto).
func TestConsultarProyecto_Devuelve200ConElProyecto(t *testing.T) {
	h := newHandler()
	id := createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/proyectos/1", "")

	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	if p.ID != id || p.Nombre != "Sistema de ventas" || p.Estado != "PLANIFICADO" {
		t.Errorf("proyecto = %+v", p)
	}
}

func TestConsultarProyecto_InexistenteDevuelve404(t *testing.T) {
	for _, path := range []string{"/proyectos/999", "/proyectos/0", "/proyectos/-1"} {
		t.Run(path, func(t *testing.T) {
			rec := do(t, newHandler(), http.MethodGet, path, "")

			expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
		})
	}
}

func TestConsultarProyecto_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/proyectos/abc", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-TEC01.2 · Las demás rutas con {id} también rechazan un id no numérico.
func TestRutasConID_IDNoNumericoDevuelve400(t *testing.T) {
	casos := []struct {
		nombre, method, path, body string
	}{
		{"modificar proyecto", http.MethodPut, "/proyectos/abc", `{"nombre":"Sistema de ventas","estado":"PLANIFICADO"}`},
		{"crear ítem", http.MethodPost, "/proyectos/abc/backlog", validItem},
		{"listar backlog", http.MethodGet, "/proyectos/abc/backlog", ""},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			rec := do(t, newHandler(), c.method, c.path, c.body)

			expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
		})
	}
}

// CA-TEC01.2 · PUT /proyectos/{id} expone HU-02 (modificar proyecto).
func TestModificarProyecto_Devuelve200ConElProyectoModificado(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPut, "/proyectos/1", `{"nombre":"Ventas 2.0","descripcion":"Nueva versión","estado":"en_curso"}`)

	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	want := projectJSON{ID: 1, Nombre: "Ventas 2.0", Descripcion: "Nueva versión", Estado: "EN_CURSO"}
	if p != want {
		t.Errorf("proyecto = %+v, se esperaba %+v", p, want)
	}

	// La consulta posterior devuelve los datos modificados.
	rec = do(t, h, http.MethodGet, "/proyectos/1", "")
	decode(t, rec, &p)
	if p != want {
		t.Errorf("proyecto consultado = %+v, se esperaba %+v", p, want)
	}
}

func TestModificarProyecto_EstadoInvalidoDevuelve400(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPut, "/proyectos/1", `{"nombre":"Sistema de ventas","descripcion":"","estado":"PAUSADO"}`)

	expectError(t, rec, http.StatusBadRequest, project.ErrInvalidStatus.Error())
}

func TestModificarProyecto_NombreRepetidoDevuelve409(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	createProject(t, h, "Portal de clientes")

	rec := do(t, h, http.MethodPut, "/proyectos/2", `{"nombre":"Sistema de ventas","descripcion":"","estado":"PLANIFICADO"}`)

	expectError(t, rec, http.StatusConflict, project.ErrDuplicateName.Error())
}

func TestModificarProyecto_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPut, "/proyectos/999", `{"nombre":"Otro","descripcion":"","estado":"PLANIFICADO"}`)

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

func TestModificarProyecto_JSONInvalidoDevuelve400(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPut, "/proyectos/1", `no es json`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

// CA-TEC01.2 · POST /proyectos/{id}/backlog expone HU-04 (crear ítem).
func TestCrearItem_Devuelve201ConElItemCreado(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPost, "/proyectos/1/backlog", validItem)

	expectStatus(t, rec, http.StatusCreated)
	var it itemJSON
	decode(t, rec, &it)
	if it.ID != 1 || it.ProyectoID != 1 || it.Titulo != "Login de usuarios" || it.Descripcion != "Acceso" ||
		it.Prioridad != "ALTA" || it.Estado != "PENDIENTE" || it.StoryPoints != 0 ||
		len(it.Criterios) != 1 || it.Criterios[0] != "Credenciales válidas" {
		t.Errorf("ítem = %+v", it)
	}
}

func TestCrearItem_DatosInvalidosDevuelven400(t *testing.T) {
	cases := map[string]struct {
		body string
		want error
	}{
		"título vacío":       {`{"titulo":"  ","prioridad":"ALTA","criterios":["Ok"]}`, backlog.ErrEmptyTitle},
		"prioridad inválida": {`{"titulo":"Login","prioridad":"URGENTE","criterios":["Ok"]}`, backlog.ErrInvalidPriority},
		"sin criterios":      {`{"titulo":"Login","prioridad":"ALTA","criterios":[]}`, backlog.ErrNoAcceptanceCriteria},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandler()
			createProject(t, h, "Sistema de ventas")

			rec := do(t, h, http.MethodPost, "/proyectos/1/backlog", tc.body)

			expectError(t, rec, http.StatusBadRequest, tc.want.Error())
		})
	}
}

func TestCrearItem_ProyectoInexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/proyectos/999/backlog", validItem)

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

func TestCrearItem_JSONInvalidoDevuelve400(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPost, "/proyectos/1/backlog", `{"titulo":`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

// CA-TEC01.2 · GET /proyectos/{id}/backlog expone HU-06 (listar backlog).
func TestListarBacklog_Devuelve200ConLosItemsDelProyecto(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	createProject(t, h, "Portal de clientes")
	do(t, h, http.MethodPost, "/proyectos/1/backlog", validItem)
	do(t, h, http.MethodPost, "/proyectos/2/backlog", `{"titulo":"Otro proyecto","prioridad":"BAJA","criterios":["Ok"]}`)
	do(t, h, http.MethodPost, "/proyectos/1/backlog", `{"titulo":"Cierre de sesión","prioridad":"MEDIA","criterios":["Ok"]}`)

	rec := do(t, h, http.MethodGet, "/proyectos/1/backlog", "")

	expectStatus(t, rec, http.StatusOK)
	var items []itemJSON
	decode(t, rec, &items)
	if len(items) != 2 || items[0].ID != 1 || items[1].ID != 3 {
		t.Errorf("ítems = %+v, se esperaban los ítems 1 y 3 del proyecto 1", items)
	}
}

func TestListarBacklog_SinItemsDevuelveListaVacia(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/proyectos/1/backlog", "")

	expectStatus(t, rec, http.StatusOK)
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("cuerpo = %s, se esperaba []", got)
	}
}

func TestListarBacklog_ProyectoInexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/proyectos/999/backlog", "")

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

// brokenRepository es un repositorio de proyectos que siempre falla.
type brokenRepository struct{}

var errStorage = errors.New("detalle interno del almacenamiento")

func (brokenRepository) Save(project.Project) (project.Project, error) {
	return project.Project{}, errStorage
}
func (brokenRepository) List() ([]project.Project, error) { return nil, errStorage }
func (brokenRepository) GetByID(int) (project.Project, error) {
	return project.Project{}, errStorage
}
func (brokenRepository) Update(project.Project) (project.Project, error) {
	return project.Project{}, errStorage
}

// CA-TEC01.2 · Un error inesperado devuelve 500 sin exponer el detalle interno.
func TestErrorInterno_Devuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(brokenRepository{})
	h := httpapi.NewHandler(projects, backlog.NewManager(backlog.NewMemoryRepository(), projects))

	rec := do(t, h, http.MethodPost, "/proyectos", `{"nombre":"Sistema de ventas"}`)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
