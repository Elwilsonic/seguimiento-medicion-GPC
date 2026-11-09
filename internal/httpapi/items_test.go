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
// Trazabilidad: CA-TEC04.9, la API expone las historias de ítems ya
// integradas en main:
//   - GET /api/items/{id} expone HU-05 (consultar ítem).
//   - PUT /api/items/{id} expone HU-07 (modificar ítem).
//   - PATCH /api/items/{id}/estado expone HU-08 (cambiar estado).
//   - GET /api/proyectos/{id}/backlog?orden=prioridad expone HU-09 (ordenar
//     por prioridad).
// Por endpoint hay un caso exitoso y un test por cada código de error. Los
// errores que comparten código van como subtests: verifican la traducción a
// HTTP, no vuelven a probar las reglas de negocio de cada historia.

// newHandlerWithItem arma la API con el proyecto 1 "Sistema de ventas" y su
// ítem 1 creado con validItem.
func newHandlerWithItem(t *testing.T) http.Handler {
	t.Helper()
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem), http.StatusCreated)
	return h
}

// TEC-04 · CA-TEC04.9 · Verifica que GET /api/items/{id} consulte el ítem de HU-05 y responda 200.
// CA-TEC04.9 · GET /api/items/{id} expone HU-05 (consultar ítem).
func TestConsultarItem_Devuelve200ConElItem(t *testing.T) {
	h := newHandlerWithItem(t)

	rec := do(t, h, http.MethodGet, "/api/items/1", "")

	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.ID != 1 || it.ProyectoID != 1 || it.Titulo != "Login de usuarios" || it.Descripcion != "Acceso" ||
		it.Prioridad != "ALTA" || it.Estado != "PENDIENTE" || it.HorasEstimadas != 0 || it.SprintID != 0 ||
		len(it.Criterios) != 1 || it.Criterios[0] != "Credenciales válidas" {
		t.Errorf("ítem = %+v", it)
	}
}

// TEC-04 · CA-TEC04.9 · Verifica que consultar un ítem inexistente responda 404.
func TestConsultarItem_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/items/999", "")

	expectError(t, rec, http.StatusNotFound, backlog.ErrItemNotFound.Error())
}

// TEC-04 · CA-TEC04.9 · Verifica que PUT /api/items/{id} modifique el ítem de HU-07 y responda 200.
// CA-TEC04.9 · PUT /api/items/{id} expone HU-07 (modificar ítem).
func TestModificarItem_Devuelve200ConElItemModificado(t *testing.T) {
	h := newHandlerWithItem(t)

	rec := do(t, h, http.MethodPut, "/api/items/1",
		`{"titulo":"Login con Google","descripcion":"OAuth","prioridad":"media","criterios":["Redirige a Google"],"horas_estimadas":6.5}`)

	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.ID != 1 || it.Titulo != "Login con Google" || it.Descripcion != "OAuth" || it.Prioridad != "MEDIA" ||
		it.HorasEstimadas != 6.5 || len(it.Criterios) != 1 || it.Criterios[0] != "Redirige a Google" {
		t.Errorf("ítem = %+v", it)
	}

	// La consulta posterior devuelve los datos modificados.
	rec = do(t, h, http.MethodGet, "/api/items/1", "")
	var got itemJSON
	decode(t, rec, &got)
	if got.Titulo != "Login con Google" || got.HorasEstimadas != 6.5 {
		t.Errorf("ítem consultado = %+v", got)
	}
}

// TEC-04 · CA-TEC04.9 · Verifica que modificar un ítem con datos inválidos responda 400 con el error de cada caso.
func TestModificarItem_DatosInvalidosDevuelven400(t *testing.T) {
	cases := map[string]struct {
		body string
		want error
	}{
		"título vacío":       {`{"titulo":"  ","prioridad":"ALTA","criterios":["Ok"]}`, backlog.ErrEmptyTitle},
		"prioridad inválida": {`{"titulo":"Login","prioridad":"URGENTE","criterios":["Ok"]}`, backlog.ErrInvalidPriority},
		"sin criterios":      {`{"titulo":"Login","prioridad":"ALTA","criterios":[]}`, backlog.ErrNoAcceptanceCriteria},
		"horas negativas":    {`{"titulo":"Login","prioridad":"ALTA","criterios":["Ok"],"horas_estimadas":-1}`, backlog.ErrInvalidEstimatedHours},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithItem(t)

			rec := do(t, h, http.MethodPut, "/api/items/1", tc.body)

			expectError(t, rec, http.StatusBadRequest, tc.want.Error())
		})
	}
}

// TEC-04 · CA-TEC04.9 · Verifica que modificar un ítem inexistente responda 404.
func TestModificarItem_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPut, "/api/items/999", validItem)

	expectError(t, rec, http.StatusNotFound, backlog.ErrItemNotFound.Error())
}

// TEC-04 · CA-TEC04.9 · Verifica que PATCH /api/items/{id}/estado cambie el estado de HU-08 y responda 200.
// CA-TEC04.9 · PATCH /api/items/{id}/estado expone HU-08 (cambiar estado).
func TestCambiarEstadoItem_Devuelve200ConElEstadoNuevo(t *testing.T) {
	h := newHandlerWithItem(t)

	rec := do(t, h, http.MethodPatch, "/api/items/1/estado", `{"estado":"en_curso"}`)

	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.ID != 1 || it.Estado != "EN_CURSO" {
		t.Errorf("ítem = %+v, se esperaba el ítem 1 EN_CURSO", it)
	}
}

// TEC-04 · CA-TEC04.9 · Verifica que un estado inválido responda 400.
func TestCambiarEstadoItem_EstadoInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithItem(t)

	rec := do(t, h, http.MethodPatch, "/api/items/1/estado", `{"estado":"BLOQUEADO"}`)

	expectError(t, rec, http.StatusBadRequest, backlog.ErrInvalidStatus.Error())
}

// TEC-04 · CA-TEC04.9 · Verifica que cambiar el estado de un ítem inexistente responda 404.
func TestCambiarEstadoItem_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPatch, "/api/items/999/estado", `{"estado":"EN_CURSO"}`)

	expectError(t, rec, http.StatusNotFound, backlog.ErrItemNotFound.Error())
}

// TEC-04 · CA-TEC04.9 · Verifica que una transición de estado no permitida responda 409.
func TestCambiarEstadoItem_TransicionNoPermitidaDevuelve409(t *testing.T) {
	h := newHandlerWithItem(t)

	rec := do(t, h, http.MethodPatch, "/api/items/1/estado", `{"estado":"TERMINADO"}`)

	expectError(t, rec, http.StatusConflict, backlog.ErrInvalidTransition.Error())
}

// TEC-04 · CA-TEC04.9 · Verifica que GET /api/proyectos/{id}/backlog?orden=prioridad devuelva el backlog ordenado por prioridad (HU-09).
// CA-TEC04.9 · GET /api/proyectos/{id}/backlog?orden=prioridad expone HU-09
// (ordenar por prioridad). Sin el parámetro, el listado sigue ordenado por ID
// (TestListarBacklog_Devuelve200ConLosItemsDelProyecto).
func TestListarBacklogPorPrioridad_Devuelve200OrdenadoPorPrioridad(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	do(t, h, http.MethodPost, "/api/proyectos/1/backlog", `{"titulo":"Reportes","prioridad":"BAJA","criterios":["Ok"]}`)
	do(t, h, http.MethodPost, "/api/proyectos/1/backlog", `{"titulo":"Cierre de sesión","prioridad":"MEDIA","criterios":["Ok"]}`)
	do(t, h, http.MethodPost, "/api/proyectos/1/backlog", `{"titulo":"Login","prioridad":"ALTA","criterios":["Ok"]}`)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/backlog?orden=prioridad", "")

	expectStatus(t, rec, http.StatusOK)
	var items []itemJSON
	decode(t, rec, &items)
	if len(items) != 3 || items[0].ID != 3 || items[1].ID != 2 || items[2].ID != 1 {
		t.Errorf("ítems = %+v, se esperaban los ítems 3 (ALTA), 2 (MEDIA) y 1 (BAJA)", items)
	}
}

// TEC-04 · CA-TEC04.9 · Verifica que listar por prioridad el backlog de un proyecto inexistente responda 404.
func TestListarBacklogPorPrioridad_ProyectoInexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos/999/backlog?orden=prioridad", "")

	expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
}

// HU-20 · Asignar story points a un ítem del backlog (Issue #38)
//
// Trazabilidad: CA-20.8, PATCH /api/items/{id}/story-points expone HU-20.
// Hay un caso exitoso y un test por código de error (400, 404, 409, 413 y
// 500); el 413 y el texto después del JSON están en bodyEndpoints
// (server_test.go).

// HU-20 · CA-20.8 · Verifica que PATCH /api/items/{id}/story-points asigne los Story Points y responda 200 con el ítem.
func TestAsignarStoryPoints_Devuelve200ConLosStoryPointsAsignados(t *testing.T) {
	h := newHandlerWithItem(t)

	rec := do(t, h, http.MethodPatch, "/api/items/1/story-points", `{"story_points":5}`)

	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.ID != 1 || it.StoryPoints != 5 || it.Titulo != "Login de usuarios" || it.Estado != "PENDIENTE" {
		t.Errorf("ítem = %+v, se esperaba el ítem 1 con 5 story points", it)
	}

	// La consulta posterior devuelve los story points asignados.
	rec = do(t, h, http.MethodGet, "/api/items/1", "")
	var got itemJSON
	decode(t, rec, &got)
	if got.StoryPoints != 5 {
		t.Errorf("story points consultados = %d, se esperaban 5", got.StoryPoints)
	}
}

// HU-20 · CA-20.8 · Verifica que el texto "8" se acepte como el número 8 en la API.
// SDD de HU-20, caso límite: con json.Number, el texto "5" se acepta como 5.
func TestAsignarStoryPoints_TextoNumericoSeAceptaComoNumero(t *testing.T) {
	h := newHandlerWithItem(t)

	rec := do(t, h, http.MethodPatch, "/api/items/1/story-points", `{"story_points":"8"}`)

	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.StoryPoints != 8 {
		t.Errorf("story points = %d, se esperaban 8", it.StoryPoints)
	}
}

// HU-20 · CA-20.3 · Verifica que un valor fuera de la escala Fibonacci responda 400 y no cambie el ítem.
func TestAsignarStoryPoints_ValorFueraDeLaEscalaDevuelve400(t *testing.T) {
	cases := map[string]string{
		"cero":          `{"story_points":0}`,
		"negativo":      `{"story_points":-1}`,
		"intermedio":    `{"story_points":4}`,
		"mayor que 21":  `{"story_points":22}`,
		"decimal":       `{"story_points":2.5}`,
		"número enorme": `{"story_points":1e400}`,
		"null":          `{"story_points":null}`,
		"campo ausente": `{}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithItem(t)

			rec := do(t, h, http.MethodPatch, "/api/items/1/story-points", body)

			expectError(t, rec, http.StatusBadRequest, backlog.ErrInvalidStoryPoints.Error())
			expectItemWithoutStoryPoints(t, h)
		})
	}
}

// HU-20 · CA-20.8 · Verifica que un cuerpo que no es JSON o un valor no numérico responda 400.
func TestAsignarStoryPoints_JSONInvalidoDevuelve400(t *testing.T) {
	cases := map[string]string{
		"no es JSON":        `story_points=5`,
		"valor no numérico": `{"story_points":true}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithItem(t)

			rec := do(t, h, http.MethodPatch, "/api/items/1/story-points", body)

			expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
			expectItemWithoutStoryPoints(t, h)
		})
	}
}

// HU-20 · CA-20.8 · Verifica que un {id} no numérico responda 400.
func TestAsignarStoryPoints_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandlerWithItem(t), http.MethodPatch, "/api/items/abc/story-points", `{"story_points":5}`)

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// HU-20 · CA-20.4 · Verifica que asignar Story Points a un ítem inexistente responda 404.
func TestAsignarStoryPoints_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPatch, "/api/items/999/story-points", `{"story_points":5}`)

	expectError(t, rec, http.StatusNotFound, backlog.ErrItemNotFound.Error())
}

// HU-20 · CA-20.5 · Verifica que asignar Story Points a un ítem TERMINADO responda 409.
func TestAsignarStoryPoints_ItemTerminadoDevuelve409(t *testing.T) {
	h := newHandlerWithItem(t)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/items/1/estado", `{"estado":"EN_CURSO"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/items/1/estado", `{"estado":"TERMINADO"}`), http.StatusOK)

	rec := do(t, h, http.MethodPatch, "/api/items/1/story-points", `{"story_points":8}`)

	expectError(t, rec, http.StatusConflict, backlog.ErrItemFinished.Error())
	expectItemWithoutStoryPoints(t, h)
}

// brokenItemRepository es un repositorio de ítems que siempre falla.
type brokenItemRepository struct{}

func (brokenItemRepository) Save(backlog.Item) (backlog.Item, error) {
	return backlog.Item{}, errStorage
}
func (brokenItemRepository) ListByProject(int) ([]backlog.Item, error) { return nil, errStorage }
func (brokenItemRepository) GetByID(int) (backlog.Item, error) {
	return backlog.Item{}, errStorage
}
func (brokenItemRepository) Update(backlog.Item) (backlog.Item, error) {
	return backlog.Item{}, errStorage
}

// HU-20 · CA-20.8 · Verifica que una falla del repositorio responda 500 sin detalle.
func TestAsignarStoryPoints_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(brokenItemRepository{}, projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))

	rec := do(t, h, http.MethodPatch, "/api/items/1/story-points", `{"story_points":5}`)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// expectItemWithoutStoryPoints verifica que el ítem 1 siga sin estimar.
func expectItemWithoutStoryPoints(t *testing.T, h http.Handler) {
	t.Helper()
	rec := do(t, h, http.MethodGet, "/api/items/1", "")
	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.StoryPoints != 0 {
		t.Errorf("story points = %d, se esperaba el ítem sin estimar", it.StoryPoints)
	}
}
