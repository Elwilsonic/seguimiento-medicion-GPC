package httpapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

// TEC-01 · Punto de entrada ejecutable mínimo (Issue #23)
//
// Trazabilidad: cada endpoint expone una historia ya integrada en main y cita
// su criterio de TEC-01:
//   - CA-TEC01.2: los endpoints de proyectos (HU-01, HU-02, HU-03) y de
//     backlog (HU-04, HU-06).
//   - CA-TEC01.3: estos tests son las pruebas automatizadas de los endpoints.
//   - CA-TEC01.5: los datos se guardan en los repositorios en memoria.
//
// TEC-04 (Issue #34), CA-TEC04.8: la API responde bajo /api/…, así que todos
// los pedidos de estos tests usan rutas con ese prefijo.

// projectJSON y itemJSON son las respuestas que devuelve la API.
type projectJSON struct {
	ID          int    `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
	// HU-18, CA-18.6: fecha de inicio en AAAA-MM-DD, o "" si no está
	// registrada.
	FechaInicio string `json:"fecha_inicio"`
	// HU-41, CA-41.9: fecha de fin en AAAA-MM-DD, o "" si no está
	// registrada.
	FechaFin string `json:"fecha_fin"`
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
	// TEC-04, CA-TEC04.9: la respuesta suma las horas estimadas (HU-07) y el
	// sprint asignado (HU-11).
	HorasEstimadas float64 `json:"horas_estimadas"`
	SprintID       int     `json:"sprint_id"`
}

type errorJSON struct {
	Error string `json:"error"`
}

// newHandler arma la API con los repositorios en memoria.
func newHandler() http.Handler {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	return httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)
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
	rec := do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"`+name+`","descripcion":""}`)
	expectStatus(t, rec, http.StatusCreated)
	var p projectJSON
	decode(t, rec, &p)
	return p.ID
}

const validItem = `{"titulo":"Login de usuarios","descripcion":"Acceso","prioridad":"alta","criterios":["Credenciales válidas"]}`

// CA-TEC01.2 · POST /proyectos expone HU-01 (crear proyecto).
func TestCrearProyecto_Devuelve201ConElProyectoCreado(t *testing.T) {
	h := newHandler()

	rec := do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"  Sistema de ventas  ","descripcion":"Gestión de ventas"}`)

	expectStatus(t, rec, http.StatusCreated)
	var p projectJSON
	decode(t, rec, &p)
	want := projectJSON{ID: 1, Nombre: "Sistema de ventas", Descripcion: "Gestión de ventas", Estado: "PLANIFICADO"}
	if p != want {
		t.Errorf("proyecto = %+v, se esperaba %+v", p, want)
	}
}

func TestCrearProyecto_NombreVacioDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/api/proyectos", `{"nombre":"   ","descripcion":""}`)

	expectError(t, rec, http.StatusBadRequest, project.ErrEmptyName.Error())
}

func TestCrearProyecto_NombreRepetidoDevuelve409(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"SISTEMA DE VENTAS","descripcion":""}`)

	expectError(t, rec, http.StatusConflict, project.ErrDuplicateName.Error())
}

func TestCrearProyecto_JSONInvalidoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/api/proyectos", `{"nombre":`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

// CA-TEC01.2 · GET /proyectos/{id} expone HU-03 (consultar proyecto).
func TestConsultarProyecto_Devuelve200ConElProyecto(t *testing.T) {
	h := newHandler()
	id := createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/api/proyectos/1", "")

	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	if p.ID != id || p.Nombre != "Sistema de ventas" || p.Estado != "PLANIFICADO" {
		t.Errorf("proyecto = %+v", p)
	}
}

func TestConsultarProyecto_InexistenteDevuelve404(t *testing.T) {
	for _, path := range []string{"/api/proyectos/999", "/api/proyectos/0", "/api/proyectos/-1"} {
		t.Run(path, func(t *testing.T) {
			rec := do(t, newHandler(), http.MethodGet, path, "")

			expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
		})
	}
}

func TestConsultarProyecto_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos/abc", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-TEC01.2 · Las demás rutas con {id} también rechazan un id no numérico.
func TestRutasConID_IDNoNumericoDevuelve400(t *testing.T) {
	casos := []struct {
		nombre, method, path, body string
	}{
		{"modificar proyecto", http.MethodPut, "/api/proyectos/abc", `{"nombre":"Sistema de ventas","estado":"PLANIFICADO"}`},
		// HU-18, CA-18.10: registrar la fecha de inicio del proyecto.
		{"registrar fecha de inicio", http.MethodPatch, "/api/proyectos/abc/fecha-inicio", `{"fecha_inicio":"2026-10-23"}`},
		// HU-41, CA-41.13: registrar la fecha de fin del proyecto.
		{"registrar fecha de fin", http.MethodPatch, "/api/proyectos/abc/fecha-fin", `{"fecha_fin":"2026-12-18"}`},
		{"crear ítem", http.MethodPost, "/api/proyectos/abc/backlog", validItem},
		{"listar backlog", http.MethodGet, "/api/proyectos/abc/backlog", ""},
		// TEC-04, CA-TEC04.9: endpoints de ítems.
		{"consultar ítem", http.MethodGet, "/api/items/abc", ""},
		{"modificar ítem", http.MethodPut, "/api/items/abc", validItem},
		{"cambiar estado", http.MethodPatch, "/api/items/abc/estado", `{"estado":"EN_CURSO"}`},
		{"votar Planning Poker", http.MethodPost, "/api/poker-sessions/abc/votes", `{"integrante":"Ana","puntos":5}`},
		// HU-23, CA-23.8: revelar los votos.
		{"revelar Planning Poker", http.MethodPost, "/api/poker-sessions/abc/reveal", ""},
		// HU-25, CA-25.9: registrar la estimación acordada.
		{"registrar estimación Planning Poker", http.MethodPost, "/api/poker-sessions/abc/estimate", `{"puntos":5}`},
		// HU-24, CA-24.12: repetir la ronda.
		{"repetir ronda Planning Poker", http.MethodPost, "/api/poker-sessions/abc/rounds", ""},
		// TEC-04, CA-TEC04.9: endpoints de sprints.
		{"crear sprint", http.MethodPost, "/api/proyectos/abc/sprints", validSprint},
		// HU-14, CA-14.4: listar los sprints del proyecto.
		{"listar sprints del proyecto", http.MethodGet, "/api/proyectos/abc/sprints", ""},
		{"asignar ítem: sprint", http.MethodPost, "/api/sprints/abc/items/1", ""},
		{"asignar ítem: ítem", http.MethodPost, "/api/sprints/1/items/abc", ""},
		{"completar ítem: sprint", http.MethodPost, "/api/sprints/abc/items/1/completar", ""},
		{"completar ítem: ítem", http.MethodPost, "/api/sprints/1/items/abc/completar", ""},
		{"listar completados", http.MethodGet, "/api/sprints/abc/completados", ""},
		{"cerrar sprint", http.MethodPost, "/api/sprints/abc/cerrar", ""},
		// HU-32, CA-32.9: métricas de un sprint y de un proyecto.
		{"métricas del sprint", http.MethodGet, "/api/sprints/abc/metricas", ""},
		{"métricas del proyecto", http.MethodGet, "/api/proyectos/abc/metricas", ""},
		// HU-36, CA-36.6: resumen del estado del proyecto.
		{"resumen del proyecto", http.MethodGet, "/api/proyectos/abc/resumen", ""},
		// HU-37, CA-37.4: métricas de cada sprint del proyecto.
		{"métricas por sprint del proyecto", http.MethodGet, "/api/proyectos/abc/metricas/sprints", ""},
		// HU-38, CA-38.7: reporte del proyecto.
		{"reporte del proyecto", http.MethodGet, "/api/proyectos/abc/reporte", ""},
		// HU-39, CA-39.8: reporte de un sprint.
		{"reporte del sprint", http.MethodGet, "/api/sprints/abc/reporte", ""},
		// HU-40, CA-40.1: PDF del reporte del proyecto.
		{"PDF del reporte del proyecto", http.MethodGet, "/api/proyectos/abc/reporte/pdf", ""},
		// HU-27, CA-27.10: listado de esfuerzos por historia y por sprint.
		{"listar esfuerzos del ítem", http.MethodGet, "/api/items/abc/esfuerzos", ""},
		{"listar esfuerzos del sprint", http.MethodGet, "/api/sprints/abc/esfuerzos", ""},
		// HU-30, CA-30.9: cambiar el estado de un defecto.
		{"cambiar estado de defecto", http.MethodPatch, "/api/defectos/abc/estado", `{"estado":"EN_CORRECCION"}`},
		// HU-31, CA-31.9: consultar los defectos de un proyecto y de un sprint.
		{"listar defectos del proyecto", http.MethodGet, "/api/proyectos/abc/defectos", ""},
		{"listar defectos del sprint", http.MethodGet, "/api/sprints/abc/defectos?estado=ABIERTO", ""},
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

	rec := do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Ventas 2.0","descripcion":"Nueva versión","estado":"en_curso"}`)

	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	want := projectJSON{ID: 1, Nombre: "Ventas 2.0", Descripcion: "Nueva versión", Estado: "EN_CURSO"}
	if p != want {
		t.Errorf("proyecto = %+v, se esperaba %+v", p, want)
	}

	// La consulta posterior devuelve los datos modificados.
	rec = do(t, h, http.MethodGet, "/api/proyectos/1", "")
	decode(t, rec, &p)
	if p != want {
		t.Errorf("proyecto consultado = %+v, se esperaba %+v", p, want)
	}
}

func TestModificarProyecto_EstadoInvalidoDevuelve400(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Sistema de ventas","descripcion":"","estado":"PAUSADO"}`)

	expectError(t, rec, http.StatusBadRequest, project.ErrInvalidStatus.Error())
}

func TestModificarProyecto_NombreRepetidoDevuelve409(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	createProject(t, h, "Portal de clientes")

	rec := do(t, h, http.MethodPut, "/api/proyectos/2", `{"nombre":"Sistema de ventas","descripcion":"","estado":"PLANIFICADO"}`)

	expectError(t, rec, http.StatusConflict, project.ErrDuplicateName.Error())
}

func TestModificarProyecto_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPut, "/api/proyectos/999", `{"nombre":"Otro","descripcion":"","estado":"PLANIFICADO"}`)

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

func TestModificarProyecto_JSONInvalidoDevuelve400(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPut, "/api/proyectos/1", `no es json`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

// HU-18 · Registrar la fecha de inicio del proyecto (Issue #36)
//
// Trazabilidad: CA-18.2, CA-18.3, CA-18.4, CA-18.6 y CA-18.10. PATCH
// /api/proyectos/{id}/fecha-inicio, con un caso exitoso y un test por código
// de error; fecha_inicio en las respuestas de POST, GET y PUT.

// fechaInicio arma el cuerpo de PATCH /api/proyectos/{id}/fecha-inicio.
func fechaInicio(date string) string {
	return `{"fecha_inicio":"` + date + `"}`
}

// expectProjectStartDate verifica la fecha de inicio que devuelve la consulta
// del proyecto.
func expectProjectStartDate(t *testing.T, h http.Handler, id, want string) {
	t.Helper()
	rec := do(t, h, http.MethodGet, "/api/proyectos/"+id, "")
	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	if p.FechaInicio != want {
		t.Errorf("fecha_inicio consultada = %q, se esperaba %q", p.FechaInicio, want)
	}
}

// CA-18.1 y CA-18.10
func TestRegistrarFechaInicioProyecto_Devuelve200ConElProyecto(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas","estado":"EN_CURSO"}`)

	rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-10-23"))

	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	want := projectJSON{ID: 1, Nombre: "Sistema de ventas", Descripcion: "Gestión de ventas", Estado: "EN_CURSO", FechaInicio: "2026-10-23"}
	if p != want {
		t.Errorf("proyecto = %+v, se esperaba %+v", p, want)
	}
	expectProjectStartDate(t, h, "1", "2026-10-23")
}

// CA-18.5
func TestRegistrarFechaInicioProyecto_VolverARegistrarReemplazaLaFecha(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-10-23"))

	rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-11-02"))

	expectStatus(t, rec, http.StatusOK)
	expectProjectStartDate(t, h, "1", "2026-11-02")
}

// CA-18.3
func TestRegistrarFechaInicioProyecto_FechaAusenteDevuelve400(t *testing.T) {
	bodies := map[string]string{
		"vacía":   `{"fecha_inicio":""}`,
		"null":    `{"fecha_inicio":null}`,
		"ausente": `{}`,
	}
	for caseName, body := range bodies {
		t.Run(caseName, func(t *testing.T) {
			h := newHandler()
			createProject(t, h, "Sistema de ventas")

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", body)

			expectError(t, rec, http.StatusBadRequest, project.ErrMissingStartDate.Error())
			expectProjectStartDate(t, h, "1", "")
		})
	}
}

// CA-18.2
func TestRegistrarFechaInicioProyecto_FormatoOFechaInvalidaDevuelve400(t *testing.T) {
	for _, date := range []string{"2026-02-30", "2026-13-01", "23/10/2026", "2026-1-5", "2026-10-23T00:00:00Z", " 2026-10-23 "} {
		t.Run(date, func(t *testing.T) {
			h := newHandler()
			createProject(t, h, "Sistema de ventas")

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio(date))

			expectError(t, rec, http.StatusBadRequest, "las fechas deben tener el formato AAAA-MM-DD")
			expectProjectStartDate(t, h, "1", "")
		})
	}
}

// CA-18.10: un cuerpo que no es JSON o una fecha que no es texto es JSON
// inválido.
func TestRegistrarFechaInicioProyecto_JSONInvalidoDevuelve400(t *testing.T) {
	for _, body := range []string{`no es json`, `{"fecha_inicio":20261023}`} {
		t.Run(body, func(t *testing.T) {
			h := newHandler()
			createProject(t, h, "Sistema de ventas")

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", body)

			expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
		})
	}
}

// CA-18.4
func TestRegistrarFechaInicioProyecto_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPatch, "/api/proyectos/999/fecha-inicio", fechaInicio("2026-10-23"))

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

// CA-18.3 (orden de validación): la fecha se valida antes que el proyecto.
func TestRegistrarFechaInicioProyecto_FechaVaciaEnProyectoInexistenteDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPatch, "/api/proyectos/999/fecha-inicio", fechaInicio(""))

	expectError(t, rec, http.StatusBadRequest, project.ErrMissingStartDate.Error())
}

// CA-18.8 y CA-18.10: una falla del repositorio devuelve 500 sin detalle.
func TestRegistrarFechaInicioProyecto_ErrorInternoDevuelve500(t *testing.T) {
	projects := project.NewManager(brokenRepository{})
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)

	rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-10-23"))

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// CA-18.6: POST, GET y PUT devuelven fecha_inicio; "" si no está registrada.
func TestProyecto_RespuestasIncluyenLaFechaDeInicio(t *testing.T) {
	h := newHandler()

	rec := do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"Sistema de ventas","descripcion":""}`)
	expectStatus(t, rec, http.StatusCreated)
	var created map[string]any
	decode(t, rec, &created)
	if got, ok := created["fecha_inicio"]; !ok || got != "" {
		t.Errorf("POST: fecha_inicio = %v (presente: %t), se esperaba \"\"", got, ok)
	}

	do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-10-23"))
	expectProjectStartDate(t, h, "1", "2026-10-23")

	rec = do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Ventas 2.0","descripcion":"","estado":"EN_CURSO"}`)
	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	if p.FechaInicio != "2026-10-23" {
		t.Errorf("PUT: fecha_inicio = %q, se esperaba 2026-10-23", p.FechaInicio)
	}
}

// CA-18.6 (RN-18.6): PUT no recibe la fecha; si un cliente la envía, se
// ignora y la registrada se conserva.
func TestModificarProyecto_ConservaLaFechaDeInicio(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-10-23"))

	rec := do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Ventas 2.0","descripcion":"","estado":"EN_CURSO","fecha_inicio":"2020-01-01"}`)

	expectStatus(t, rec, http.StatusOK)
	expectProjectStartDate(t, h, "1", "2026-10-23")
}

// CA-TEC01.2 · POST /proyectos/{id}/backlog expone HU-04 (crear ítem).
func TestCrearItem_Devuelve201ConElItemCreado(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem)

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

			rec := do(t, h, http.MethodPost, "/api/proyectos/1/backlog", tc.body)

			expectError(t, rec, http.StatusBadRequest, tc.want.Error())
		})
	}
}

func TestCrearItem_ProyectoInexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/api/proyectos/999/backlog", validItem)

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

func TestCrearItem_JSONInvalidoDevuelve400(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPost, "/api/proyectos/1/backlog", `{"titulo":`)

	expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
}

// CA-TEC01.2 · GET /proyectos/{id}/backlog expone HU-06 (listar backlog).
func TestListarBacklog_Devuelve200ConLosItemsDelProyecto(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	createProject(t, h, "Portal de clientes")
	do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem)
	do(t, h, http.MethodPost, "/api/proyectos/2/backlog", `{"titulo":"Otro proyecto","prioridad":"BAJA","criterios":["Ok"]}`)
	do(t, h, http.MethodPost, "/api/proyectos/1/backlog", `{"titulo":"Cierre de sesión","prioridad":"MEDIA","criterios":["Ok"]}`)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/backlog", "")

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

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/backlog", "")

	expectStatus(t, rec, http.StatusOK)
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("cuerpo = %s, se esperaba []", got)
	}
}

func TestListarBacklog_ProyectoInexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos/999/backlog", "")

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
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)

	rec := do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"Sistema de ventas"}`)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// TEC-04 · Interfaz usable (Issue #34)
//
// Trazabilidad: las mejoras del servidor citan su criterio de TEC-04:
//   - CA-TEC04.5: un cuerpo de más de 1 MiB se rechaza con 413 y no se procesa.
//   - CA-TEC04.6: un cuerpo con texto después del JSON se rechaza con 400 y no
//     se procesa.
// Cada criterio se prueba en todos los endpoints que reciben un cuerpo.

// bodyEndpoint es un endpoint que recibe un cuerpo JSON. untouched verifica
// que el pedido rechazado no creó ni modificó nada.
type bodyEndpoint struct {
	name      string
	method    string
	path      string
	untouched func(t *testing.T, h http.Handler)
}

// bodyEndpoints parte de un proyecto 1 "Sistema de ventas" con su ítem 1
// creado con validItem (newHandlerWithItem).
var bodyEndpoints = []bodyEndpoint{
	{
		name:   "crear proyecto",
		method: http.MethodPost,
		path:   "/api/proyectos",
		untouched: func(t *testing.T, h http.Handler) {
			rec := do(t, h, http.MethodGet, "/api/proyectos/2", "")
			expectStatus(t, rec, http.StatusNotFound)
		},
	},
	{
		name:   "modificar proyecto",
		method: http.MethodPut,
		path:   "/api/proyectos/1",
		untouched: func(t *testing.T, h http.Handler) {
			rec := do(t, h, http.MethodGet, "/api/proyectos/1", "")
			expectStatus(t, rec, http.StatusOK)
			var p projectJSON
			decode(t, rec, &p)
			if p.Nombre != "Sistema de ventas" || p.Descripcion != "" {
				t.Errorf("proyecto = %+v, se esperaba sin cambios", p)
			}
		},
	},
	// HU-18, CA-18.10: el cuerpo se rechaza antes de llegar al Manager, así
	// que el proyecto 1 sigue sin fecha de inicio.
	{
		name:   "registrar fecha de inicio",
		method: http.MethodPatch,
		path:   "/api/proyectos/1/fecha-inicio",
		untouched: func(t *testing.T, h http.Handler) {
			rec := do(t, h, http.MethodGet, "/api/proyectos/1", "")
			expectStatus(t, rec, http.StatusOK)
			var p projectJSON
			decode(t, rec, &p)
			if p.FechaInicio != "" {
				t.Errorf("proyecto = %+v, se esperaba sin fecha de inicio", p)
			}
		},
	},
	// HU-41, CA-41.13: el cuerpo se rechaza antes de llegar al Manager, así
	// que el proyecto 1 sigue sin fecha de fin.
	{
		name:   "registrar fecha de fin",
		method: http.MethodPatch,
		path:   "/api/proyectos/1/fecha-fin",
		untouched: func(t *testing.T, h http.Handler) {
			rec := do(t, h, http.MethodGet, "/api/proyectos/1", "")
			expectStatus(t, rec, http.StatusOK)
			var p projectJSON
			decode(t, rec, &p)
			if p.FechaFin != "" {
				t.Errorf("proyecto = %+v, se esperaba sin fecha de fin", p)
			}
		},
	},
	{
		name:   "crear ítem",
		method: http.MethodPost,
		path:   "/api/proyectos/1/backlog",
		untouched: func(t *testing.T, h http.Handler) {
			rec := do(t, h, http.MethodGet, "/api/proyectos/1/backlog", "")
			expectStatus(t, rec, http.StatusOK)
			var items []itemJSON
			decode(t, rec, &items)
			if len(items) != 1 {
				t.Errorf("backlog = %+v, se esperaba solo el ítem 1", items)
			}
		},
	},
	// TEC-04, CA-TEC04.9: endpoints de ítems que reciben un cuerpo.
	{
		name:      "modificar ítem",
		method:    http.MethodPut,
		path:      "/api/items/1",
		untouched: expectItemUntouched,
	},
	{
		name:      "cambiar estado",
		method:    http.MethodPatch,
		path:      "/api/items/1/estado",
		untouched: expectItemUntouched,
	},
	// HU-20, CA-20.8: asignar story points.
	{
		name:      "asignar story points",
		method:    http.MethodPatch,
		path:      "/api/items/1/story-points",
		untouched: expectItemWithoutStoryPoints,
	},
	{
		name:   "votar Planning Poker",
		method: http.MethodPost,
		path:   "/api/poker-sessions/1/votes",
		untouched: func(t *testing.T, h http.Handler) {
			t.Helper()
		},
	},
	// HU-25, CA-25.9: el cuerpo se rechaza antes de llegar al Manager, así
	// que el ítem 1 sigue sin estimar.
	{
		name:      "registrar estimación Planning Poker",
		method:    http.MethodPost,
		path:      "/api/poker-sessions/1/estimate",
		untouched: expectItemWithoutStoryPoints,
	},
	// TEC-04, CA-TEC04.9: endpoint de sprints que recibe un cuerpo.
	{
		name:   "crear sprint",
		method: http.MethodPost,
		path:   "/api/proyectos/1/sprints",
		untouched: func(t *testing.T, h http.Handler) {
			rec := do(t, h, http.MethodPost, "/api/sprints/1/cerrar", "")
			expectStatus(t, rec, http.StatusNotFound)
		},
	},
	// HU-30, CA-30.9: el cuerpo se rechaza antes de llegar al Manager. No
	// hay defectos ni una consulta de defectos (HU-31) para comprobar más.
	{
		name:   "cambiar estado de defecto",
		method: http.MethodPatch,
		path:   "/api/defectos/1/estado",
		untouched: func(t *testing.T, h http.Handler) {
			t.Helper()
		},
	},
}

// expectItemUntouched verifica que el ítem 1 siga como lo creó validItem.
func expectItemUntouched(t *testing.T, h http.Handler) {
	t.Helper()
	rec := do(t, h, http.MethodGet, "/api/items/1", "")
	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.Descripcion != "Acceso" || it.Estado != "PENDIENTE" {
		t.Errorf("ítem = %+v, se esperaba sin cambios", it)
	}
}

// validBody es un cuerpo válido para cualquiera de los endpoints: los campos
// que un endpoint no usa se ignoran. EN_CURSO es un estado válido tanto para
// el proyecto como para el ítem.
func validBody(description string) string {
	return `{"nombre":"Portal de clientes","descripcion":"` + description +
		`","estado":"EN_CURSO","titulo":"Login de usuarios","prioridad":"ALTA","criterios":["Ok"]` +
		`,"fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Ok"}`
}

// CA-TEC04.5 · Un cuerpo de más de 1 MiB se rechaza con 413 y no se procesa.
// El relleno son espacios dentro del JSON: el cuerpo sigue siendo válido y la
// respuesta no repite el MiB de datos si el test falla.
func TestHandler_CuerpoMayorA1MiBDa413(t *testing.T) {
	body := "{" + strings.Repeat(" ", 1<<20) + validBody("Otro proyecto")[1:]

	for _, e := range bodyEndpoints {
		t.Run(e.name, func(t *testing.T) {
			h := newHandlerWithItem(t)

			rec := do(t, h, e.method, e.path, body)

			expectError(t, rec, http.StatusRequestEntityTooLarge, "el cuerpo del pedido supera el tamaño máximo de 1 MiB")
			e.untouched(t, h)
		})
	}
}

// CA-TEC04.6 · Un cuerpo con texto después del JSON se rechaza como JSON
// inválido y no se procesa.
func TestHandler_TextoDespuesDelJSONDa400(t *testing.T) {
	body := validBody("Otro proyecto") + ` basura`

	for _, e := range bodyEndpoints {
		t.Run(e.name, func(t *testing.T) {
			h := newHandlerWithItem(t)

			rec := do(t, h, e.method, e.path, body)

			expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
			e.untouched(t, h)
		})
	}
}
