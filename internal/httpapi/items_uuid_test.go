package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-52 · Identificadores UUID del resto de las entidades, parte de ítems
// (Issue #197). Prueba solo las rutas de ítems cuyo handler está en items.go
// (tabla "Rutas por responsable" de D3): GET y PUT /api/items/{id},
// PATCH /api/items/{id}/estado y /story-points, más la creación y el listado
// del backlog. El id se decodifica con itemUUIDJSON y no con idRef, porque
// estos tests verifican que sea un UUID.

// missingItemID es un UUID versión 4 que nunca se asigna a un ítem, y
// nilItemID es el UUID nulo (CA-52.4).
const (
	missingItemID = "6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f"
	nilItemID     = "00000000-0000-0000-0000-000000000000"
)

// itemUUIDJSON es la respuesta de un ítem con los campos de HU-52: el id y
// el sprint como texto.
type itemUUIDJSON struct {
	ID       string `json:"id"`
	SprintID string `json:"sprint_id"`
	Titulo   string `json:"titulo"`
	Estado   string `json:"estado"`
}

// itemRoute es una de las 4 rutas de items.go que reciben el ítem, con un
// cuerpo válido. Todas responden 200 con el ítem.
type itemRoute struct {
	method, tail, body string
}

// itemRoutes están en un orden en el que cada una funciona sobre el ítem
// recién creado.
var itemRoutes = []itemRoute{
	{http.MethodGet, "", ""},
	{http.MethodPut, "", `{"titulo":"Login con Google","descripcion":"OAuth","prioridad":"MEDIA","criterios":["Redirige a Google"]}`},
	{http.MethodPatch, "/estado", `{"estado":"EN_CURSO"}`},
	{http.MethodPatch, "/story-points", `{"story_points":5}`},
}

// jsonItemID es el "id" del n-ésimo ítem creado por la API en h, como texto
// JSON, para armar las respuestas esperadas (HU-52, CA-52.3).
func jsonItemID(t *testing.T, h http.Handler, n int) string {
	t.Helper()
	return `"` + string(itemRef(t, h, n)) + `"`
}

// jsonItemSprintID es el "sprint_id" del n-ésimo ítem creado por la API en
// h, como valor JSON, leído con GET /api/items/{id}. Los contratos de los
// reportes lo usan porque, mientras los sprints no expongan su UUID en su
// propio "id" (parte de Buttini Cristobal), el UUID del sprint solo se ve en
// el ítem. Que sea el UUID correcto lo verifica
// TestConsultarItem_SprintIDEsElUUIDDelSprint.
func jsonItemSprintID(t *testing.T, h http.Handler, n int) string {
	t.Helper()
	rec := do(t, h, http.MethodGet, "/api/items/"+string(itemRef(t, h, n)), "")
	expectStatus(t, rec, http.StatusOK)
	var it map[string]json.RawMessage
	decode(t, rec, &it)
	return string(it["sprint_id"])
}

// itemNumber devuelve el orden de creación por la API, contando desde 1, del
// ítem con ese id, que es su ID interno. Lo usan los tests que comparan el
// orden de los ítems (HU-52: el id ya no es el número).
func itemNumber(t *testing.T, h http.Handler, id idRef) int {
	t.Helper()
	entityIDsMu.Lock()
	ids := entityIDs[h][kindItem]
	entityIDsMu.Unlock()
	for n, got := range ids {
		if got == string(id) {
			return n + 1
		}
	}
	t.Fatalf("itemNumber: el ítem %q no fue creado por la API en este handler", string(id))
	return 0
}

// createItemUUID crea un ítem en el primer proyecto de h y devuelve su id,
// que tiene que ser un UUID versión 4.
func createItemUUID(t *testing.T, h http.Handler, title string) string {
	t.Helper()
	rec := do(t, h, http.MethodPost, projectPath(t, h, 1)+"/backlog", `{"titulo":"`+title+`","prioridad":"ALTA","criterios":["Ok"]}`)
	expectStatus(t, rec, http.StatusCreated)
	id := decodeItemUUID(t, rec).ID
	if !uuidV4Canonical.MatchString(id) {
		t.Fatalf("id del ítem = %q, se esperaba un UUID versión 4", id)
	}
	return id
}

// decodeItemUUID lee la respuesta de un ítem, con el id y el sprint como
// texto: un número es un error, porque HU-52 no expone el ID entero.
func decodeItemUUID(t *testing.T, rec *httptest.ResponseRecorder) itemUUIDJSON {
	t.Helper()
	var it itemUUIDJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &it); err != nil {
		t.Fatalf("el ítem no tiene el id y el sprint_id como texto (%v): %s", err, rec.Body.String())
	}
	return it
}

// newHandlerWithItemUUID arma la API con el proyecto "Sistema de ventas" y
// un ítem, y devuelve la ruta del ítem armada con su id.
func newHandlerWithItemUUID(t *testing.T) (http.Handler, string) {
	t.Helper()
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	return h, createItemUUID(t, h, "Login de usuarios")
}

// HU-52 · CA-52.1 · Verifica que POST /api/proyectos/{id}/backlog responda 201 con el id del ítem como texto (un UUID versión 4 en minúsculas) y el sprint_id "" (sin asignar).
func TestCrearItem_IDEsUUIDVersion4(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPost, projectPath(t, h, 1)+"/backlog", validItem)

	expectStatus(t, rec, http.StatusCreated)
	it := decodeItemUUID(t, rec)
	if !uuidV4Canonical.MatchString(it.ID) {
		t.Errorf("id = %q, se esperaba un UUID versión 4 en minúsculas", it.ID)
	}
	if it.SprintID != "" {
		t.Errorf("sprint_id = %q, se esperaba \"\" (sin asignar)", it.SprintID)
	}
}

// HU-52 · CA-52.2 · Verifica que dos ítems creados por la API tengan ids distintos y que ninguno sea un número.
func TestCrearItem_DosItemsTienenUUIDDistintos(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	first := createItemUUID(t, h, "Login de usuarios")
	second := createItemUUID(t, h, "Reportes")

	if first == second {
		t.Errorf("los dos ítems tienen el id %q", first)
	}
}

// HU-52 · CA-52.3 · Verifica que las 4 rutas de ítem de items.go acepten el UUID del ítem y respondan 200 con ese mismo UUID en id.
func TestRutasDeItem_AceptanElUUID(t *testing.T) {
	h, id := newHandlerWithItemUUID(t)

	for _, r := range itemRoutes {
		path := "/api/items/" + id + r.tail
		rec := serveRaw(h, r.method, path, r.body)
		if rec.Code != http.StatusOK {
			t.Errorf("%s %s: código = %d, se esperaba 200; cuerpo: %s", r.method, path, rec.Code, rec.Body.String())
			continue
		}
		if got := decodeItemUUID(t, rec).ID; got != id {
			t.Errorf("%s %s: id = %q, se esperaba %q", r.method, path, got, id)
		}
	}
}

// HU-52 · CA-52.3 · Verifica que las rutas de ítem acepten el UUID en mayúsculas y respondan con el id en minúsculas.
func TestRutasDeItem_AceptanElUUIDEnMayusculas(t *testing.T) {
	h, id := newHandlerWithItemUUID(t)

	for _, r := range itemRoutes {
		path := "/api/items/" + strings.ToUpper(id) + r.tail
		rec := serveRaw(h, r.method, path, r.body)
		if rec.Code != http.StatusOK {
			t.Errorf("%s %s: código = %d, se esperaba 200; cuerpo: %s", r.method, path, rec.Code, rec.Body.String())
			continue
		}
		if got := decodeItemUUID(t, rec).ID; got != id {
			t.Errorf("%s %s: id = %q, se esperaba %q en minúsculas", r.method, path, got, id)
		}
	}
}

// HU-52 · CA-52.4 · Verifica que en las 4 rutas de ítem un id numérico (también 0 y negativo) responda 400 con "el id del ítem debe ser un UUID", aunque exista el ítem 1.
func TestRutasDeItem_IDNumericoDevuelve400(t *testing.T) {
	h := newHandlerWithItem(t)

	for _, id := range []string{"1", "0", "-1"} {
		for _, r := range itemRoutes {
			path := "/api/items/" + id + r.tail
			expectErrorBody(t, r.method+" "+path, serveRaw(h, r.method, path, r.body), http.StatusBadRequest, "el id del ítem debe ser un UUID")
		}
	}
}

// HU-52 · CA-52.4 · Verifica que el UUID de un ítem existente escrito entre llaves, con urn:uuid: o sin guiones responda 400 en las rutas de ítem.
func TestRutasDeItem_OtrasFormasDelUUIDDevuelven400(t *testing.T) {
	h, id := newHandlerWithItemUUID(t)

	for _, other := range []string{"{" + id + "}", "urn:uuid:" + id, strings.ReplaceAll(id, "-", "")} {
		for _, r := range itemRoutes {
			path := "/api/items/" + other + r.tail
			expectErrorBody(t, r.method+" "+path, serveRaw(h, r.method, path, r.body), http.StatusBadRequest, "el id del ítem debe ser un UUID")
		}
	}
}

// HU-52 · CA-52.4 · Verifica que en las 4 rutas de ítem un UUID que nunca se asignó, o el UUID nulo, responda 404 con ErrItemNotFound.
func TestRutasDeItem_UUIDInexistenteDevuelve404(t *testing.T) {
	h := newHandlerWithItem(t)

	for _, id := range []string{missingItemID, nilItemID} {
		for _, r := range itemRoutes {
			path := "/api/items/" + id + r.tail
			expectErrorBody(t, r.method+" "+path, serveRaw(h, r.method, path, r.body), http.StatusNotFound, backlog.ErrItemNotFound.Error())
		}
	}
}

// HU-52 · CA-52.4 · Verifica que el id del ítem se valide antes que el cuerpo: un id inválido con un cuerpo que no es JSON responde el error del id.
func TestOrdenDeErroresItem_IDInvalidoAntesQueElCuerpo(t *testing.T) {
	h := newHandlerWithItem(t)

	for _, r := range itemRoutes {
		if r.body == "" {
			continue
		}
		path := "/api/items/abc" + r.tail
		expectErrorBody(t, r.method+" "+path, serveRaw(h, r.method, path, `no es json`), http.StatusBadRequest, "el id del ítem debe ser un UUID")
	}
}

// HU-52 · CA-52.4 · Verifica que, con un UUID de ítem inexistente, el título vacío se informe antes que la inexistencia, como con un ID entero (orden de errores de HU-07).
func TestOrdenDeErroresItem_UUIDInexistenteConservaElOrden(t *testing.T) {
	h := newHandler()
	path := "/api/items/" + missingItemID

	rec := serveRaw(h, http.MethodPut, path, `{"titulo":"  ","prioridad":"ALTA","criterios":["Ok"]}`)

	expectErrorBody(t, "PUT "+path, rec, http.StatusBadRequest, backlog.ErrEmptyTitle.Error())
}

// expectItemIDsAreUUID verifica que cada ítem de la respuesta (un objeto o
// una lista) tenga su id como texto y sea uno de want, y que el sprint_id
// sea "" (ítems sin sprint asignado, D7).
func expectItemIDsAreUUID(t *testing.T, what string, rec *httptest.ResponseRecorder, want map[string]bool) {
	t.Helper()
	body := rec.Body.Bytes()
	var list []map[string]json.RawMessage
	if err := json.Unmarshal(body, &list); err != nil {
		var one map[string]json.RawMessage
		if err := json.Unmarshal(body, &one); err != nil {
			t.Errorf("%s: cuerpo no es JSON válido (%v): %s", what, err, body)
			return
		}
		list = []map[string]json.RawMessage{one}
	}
	if len(list) == 0 {
		t.Errorf("%s: la respuesta no tiene ítems: %s", what, body)
	}
	for _, it := range list {
		var id string
		if err := json.Unmarshal(it["id"], &id); err != nil || !want[id] {
			t.Errorf("%s: id = %s, se esperaba el UUID de un ítem creado, como texto", what, it["id"])
		}
		if got := string(it["sprint_id"]); got != `""` {
			t.Errorf("%s: sprint_id = %s, se esperaba \"\" (sin asignar)", what, got)
		}
	}
}

// HU-52 · CA-52.3 · Verifica que ninguna respuesta de ítem exponga el ID entero: crear, listar el backlog (también por prioridad), consultar, modificar, cambiar el estado y estimar responden el id como UUID y el sprint_id como "".
func TestRespuestasDeItem_NoExponenElIDEntero(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	rec := do(t, h, http.MethodPost, projectPath(t, h, 1)+"/backlog", validItem)
	expectStatus(t, rec, http.StatusCreated)
	want := map[string]bool{}
	var created map[string]json.RawMessage
	decode(t, rec, &created)
	var id string
	if err := json.Unmarshal(created["id"], &id); err != nil || !uuidV4Canonical.MatchString(id) {
		t.Fatalf("ítem creado: id = %s, se esperaba un UUID versión 4 como texto", created["id"])
	}
	want[id] = true
	expectItemIDsAreUUID(t, "ítem creado", rec, want)
	want[createItemUUID(t, h, "Reportes")] = true

	base := "/api/items/" + id
	steps := []struct{ what, method, path, body string }{
		{"backlog", http.MethodGet, projectPath(t, h, 1) + "/backlog", ""},
		{"backlog por prioridad", http.MethodGet, projectPath(t, h, 1) + "/backlog?orden=prioridad", ""},
		{"ítem por su UUID", http.MethodGet, base, ""},
		{"ítem modificado", http.MethodPut, base, validItem},
		{"estado del ítem", http.MethodPatch, base + "/estado", `{"estado":"EN_CURSO"}`},
		{"story points", http.MethodPatch, base + "/story-points", `{"story_points":5}`},
	}
	for _, s := range steps {
		rec := serveRaw(h, s.method, s.path, s.body)
		if rec.Code != http.StatusOK {
			t.Errorf("%s (%s %s): código = %d; cuerpo: %s", s.what, s.method, s.path, rec.Code, rec.Body.String())
			continue
		}
		expectItemIDsAreUUID(t, s.what, rec, want)
	}
}

// HU-52 · CA-52.3 · Verifica que un ítem asignado a un sprint responda en sprint_id el UUID del sprint, y no su ID entero.
func TestConsultarItem_SprintIDEsElUUIDDelSprint(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items, anyMember{})
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items, anyMember{})
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))
	createProject(t, h, "Sistema de ventas")
	id := createItemUUID(t, h, "Login de usuarios")
	expectStatus(t, do(t, h, http.MethodPost, projectPath(t, h, 1)+"/sprints", validSprint), http.StatusCreated)
	// La asignación se hace en el dominio, para no depender de la ruta de
	// sprints (de Buttini Cristobal, D3).
	it, err := items.GetByPublicID(uuid.MustParse(id))
	if err != nil {
		t.Fatalf("buscar el ítem por su UUID: %v", err)
	}
	sp, err := sprints.Get(1)
	if err != nil {
		t.Fatalf("leer el sprint: %v", err)
	}
	if _, err := sprints.AssignItem(sp.ID, it.ID); err != nil {
		t.Fatalf("asignar el ítem al sprint: %v", err)
	}

	rec := serveRaw(h, http.MethodGet, "/api/items/"+id, "")

	expectStatus(t, rec, http.StatusOK)
	if got := decodeItemUUID(t, rec).SprintID; got != sp.PublicID.String() {
		t.Errorf("sprint_id = %q, se esperaba el UUID del sprint %q", got, sp.PublicID)
	}
}
