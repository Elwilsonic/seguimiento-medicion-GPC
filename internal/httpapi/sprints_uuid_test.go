package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197), parte
// de sprints (Buttini Cristobal).

// createdID devuelve el "id" que devolvió la API al crear la n-ésima entidad
// de ese tipo en h, contando desde 1, con el registro de do() de la base.
func createdID(t *testing.T, h http.Handler, kind entityKind, n int) string {
	t.Helper()
	entityIDsMu.Lock()
	ids := entityIDs[h][kind]
	entityIDsMu.Unlock()
	if n < 1 || n > len(ids) {
		t.Fatalf("createdID: el handler tiene %d entidades de tipo %s creadas por la API, no existe la %d", len(ids), kind, n)
	}
	return ids[n-1]
}

// sprintRef es el "id" del n-ésimo sprint creado por la API en h.
func sprintRef(t *testing.T, h http.Handler, n int) idRef {
	t.Helper()
	return idRef(createdID(t, h, kindSprint, n))
}

// sessionRef es el "id" de la n-ésima sesión de Planning Poker creada por la
// API en h.
func sessionRef(t *testing.T, h http.Handler, n int) idRef {
	t.Helper()
	return idRef(createdID(t, h, kindSession, n))
}

// itemRef es el "id" del n-ésimo ítem creado por la API en h.
func itemRef(t *testing.T, h http.Handler, n int) idRef {
	t.Helper()
	return idRef(createdID(t, h, kindItem, n))
}

// jsonRef es ref como valor JSON, tal como lo devolvió la API (un número o un
// texto entre comillas), para armar cuerpos de pedido.
func jsonRef(ref idRef) string {
	if regexp.MustCompile(`^[0-9]+$`).MatchString(string(ref)) {
		return string(ref)
	}
	return `"` + string(ref) + `"`
}

// missingSprintID, missingSessionID y missingPokerItemID son UUID versión 4
// que nunca se asignan a un sprint, a una sesión ni a un ítem (HU-52,
// CA-52.4). missingPokerItemID es solo para las rutas de esta parte que
// reciben un ítem, así no choca con las constantes de la parte de ítems.
const (
	missingSprintID    = "3b9e7c2d-5f41-4a8e-b6d0-9c1e2f3a4b5c"
	missingSessionID   = "7d2a1f6e-3c84-4b9a-a5e1-0f2b3c4d5e6a"
	missingPokerItemID = "c5e8a1b2-9d3f-4e6a-8b7c-2d1e0f9a8b6c"
)

// sprintsPokerInvalidIDMessages es el mensaje de 400 de cada ruta de esta parte
// en TestRutasConID_IDNoNumericoDevuelve400, por nombre de caso (HU-52,
// CA-52.4). En las rutas con {id} y {itemId}, el caso dice cuál de los dos es
// inválido.
var sprintsPokerInvalidIDMessages = map[string]string{
	"votar Planning Poker":                "el id de la sesión debe ser un UUID",
	"revelar Planning Poker":              "el id de la sesión debe ser un UUID",
	"registrar estimación Planning Poker": "el id de la sesión debe ser un UUID",
	"repetir ronda Planning Poker":        "el id de la sesión debe ser un UUID",
	"asignar ítem: sprint":                "el id del sprint debe ser un UUID",
	"asignar ítem: ítem":                  "el id del ítem debe ser un UUID",
	"completar ítem: sprint":              "el id del sprint debe ser un UUID",
	"completar ítem: ítem":                "el id del ítem debe ser un UUID",
	"listar completados":                  "el id del sprint debe ser un UUID",
	"cerrar sprint":                       "el id del sprint debe ser un UUID",
	"métricas del sprint":                 "el id del sprint debe ser un UUID",
	"reporte del sprint":                  "el id del sprint debe ser un UUID",
}

// uuidStack es un handler con sus Manager a mano, para crear los sprints, los
// ítems y las sesiones en el dominio y conocer su UUID sin depender de cómo
// responde la API de los ítems (la parte de Cardozo Leandro, D3).
type uuidStack struct {
	h        http.Handler
	projects *project.Manager
	items    *backlog.Manager
	sprints  *sprint.Manager
	poker    *poker.Manager
	project  project.Project
}

func newUUIDStack(t *testing.T) *uuidStack {
	t.Helper()
	s := &uuidStack{projects: project.NewManager(project.NewMemoryRepository())}
	s.items = backlog.NewManager(backlog.NewMemoryRepository(), s.projects)
	s.sprints = sprint.NewManager(sprint.NewMemoryRepository(), s.projects, s.items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), s.sprints, s.items)
	s.poker = poker.NewManager(poker.NewMemoryRepository(), s.items)
	defects := defect.NewManager(defect.NewMemoryRepository(), s.items, s.sprints)
	s.h = httpapi.NewHandler(s.projects, s.items, s.sprints, efforts, s.poker, defects, newMembers(s.projects))
	p, err := s.projects.Create("Inventario", "")
	if err != nil {
		t.Fatal(err)
	}
	s.project = p
	return s
}

// sprint crea un sprint abierto en el dominio.
func (s *uuidStack) sprint(t *testing.T, name string) sprint.Sprint {
	t.Helper()
	sp, err := s.sprints.Create(s.project.ID, name, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), "Gestionar el backlog")
	if err != nil {
		t.Fatal(err)
	}
	return sp
}

// item crea un ítem en el dominio.
func (s *uuidStack) item(t *testing.T, title string) backlog.Item {
	t.Helper()
	it, err := s.items.Create(s.project.ID, title, "", "ALTA", []string{"Criterio"})
	if err != nil {
		t.Fatal(err)
	}
	return it
}

// serve atiende el pedido sin traducir la ruta: los tests de esta parte ya
// la arman con los UUID.
func (s *uuidStack) serve(method, path, body string) *httptest.ResponseRecorder {
	return serveRaw(s.h, method, path, body)
}

// jsonString devuelve el campo field del objeto JSON de rec, que tiene que ser
// un texto (CA-52.3: ninguna respuesta expone el ID entero).
func jsonString(t *testing.T, what string, rec *httptest.ResponseRecorder, field string) string {
	t.Helper()
	raw, err := jsonField(rec, field)
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		t.Fatalf("%s: %s = %s, se esperaba un texto (el UUID)", what, field, raw)
	}
	return text
}

// quoted es s como texto JSON, para comparar con un json.RawMessage.
func quoted(s string) string {
	return `"` + s + `"`
}

// sprintRoutes son las 6 rutas de sprints de esta parte que reciben el sprint
// por su {id} (D3): sprints.go, metrics.go y sprint_reports.go. "{item}" es el
// UUID del ítem.
var sprintRoutes = []struct{ name, method, path string }{
	{"asignar ítem", http.MethodPost, "/api/sprints/{sprint}/items/{item}"},
	{"completar ítem", http.MethodPost, "/api/sprints/{sprint}/items/{item}/completar"},
	{"listar completados", http.MethodGet, "/api/sprints/{sprint}/completados"},
	{"métricas del sprint", http.MethodGet, "/api/sprints/{sprint}/metricas"},
	{"reporte del sprint", http.MethodGet, "/api/sprints/{sprint}/reporte"},
	{"cerrar sprint", http.MethodPost, "/api/sprints/{sprint}/cerrar"},
}

func routeWith(path, sprintID, itemID string) string {
	return strings.NewReplacer("{sprint}", sprintID, "{item}", itemID).Replace(path)
}

// HU-52 · CA-52.1 · Verifica que crear un sprint por la API responda 201 con un id que es un UUID versión 4 en minúsculas, como texto.
func TestCrearSprint_IDEsUnUUIDVersion4(t *testing.T) {
	s := newUUIDStack(t)

	rec := s.serve(http.MethodPost, "/api/proyectos/"+s.project.PublicID.String()+"/sprints", validSprint)

	expectStatus(t, rec, http.StatusCreated)
	if id := jsonString(t, "sprint creado", rec, "id"); !uuidV4Canonical.MatchString(id) {
		t.Errorf("id = %q, se esperaba un UUID versión 4 en minúsculas", id)
	}
}

// HU-52 · CA-52.3 · Verifica que las 6 rutas de sprints de esta parte acepten el UUID del sprint y del ítem, también en mayúsculas, y respondan 200.
func TestRutasDeSprint_AceptanElUUID(t *testing.T) {
	for _, upper := range []bool{false, true} {
		s := newUUIDStack(t)
		sp := s.sprint(t, "Sprint 1")
		it := s.item(t, "Login")
		if _, err := s.items.ChangeStatus(it.ID, "EN_CURSO"); err != nil {
			t.Fatal(err)
		}
		sprintID, itemID := sp.PublicID.String(), it.PublicID.String()
		if upper {
			sprintID, itemID = strings.ToUpper(sprintID), strings.ToUpper(itemID)
		}
		for _, r := range sprintRoutes {
			rec := s.serve(r.method, routeWith(r.path, sprintID, itemID), "")
			if rec.Code != http.StatusOK {
				t.Errorf("%s (mayúsculas: %v): código = %d, se esperaba 200; cuerpo: %s", r.name, upper, rec.Code, rec.Body.String())
			}
		}
	}
}

// HU-52 · CA-52.4 · Verifica que un {id} de sprint que no es un UUID canónico (un número, entre llaves, con urn:uuid: o sin guiones) responda 400 con el mensaje del sprint en las 6 rutas.
func TestRutasDeSprint_IDInvalidoDevuelve400(t *testing.T) {
	s := newUUIDStack(t)
	valid := s.sprint(t, "Sprint 1").PublicID.String()
	itemID := s.item(t, "Login").PublicID.String()
	for _, id := range []string{"1", "{" + valid + "}", "urn:uuid:" + valid, strings.ReplaceAll(valid, "-", "")} {
		for _, r := range sprintRoutes {
			rec := s.serve(r.method, routeWith(r.path, id, itemID), "")
			expectErrorBody(t, r.name+" con "+id, rec, http.StatusBadRequest, "el id del sprint debe ser un UUID")
		}
	}
}

// HU-52 · CA-52.4 · Verifica que un UUID que no corresponde a ningún sprint (uno que nunca se asignó o el nulo) responda 404 con el mensaje del sprint en las 6 rutas.
func TestRutasDeSprint_UUIDInexistenteDevuelve404(t *testing.T) {
	s := newUUIDStack(t)
	itemID := s.item(t, "Login").PublicID.String()
	for _, id := range []string{missingSprintID, nilProjectID} {
		for _, r := range sprintRoutes {
			rec := s.serve(r.method, routeWith(r.path, id, itemID), "")
			expectErrorBody(t, r.name+" con "+id, rec, http.StatusNotFound, sprint.ErrSprintNotFound.Error())
		}
	}
}

// HU-52 · CA-52.4 · Verifica que en asignar y completar se lea el {id} del sprint antes que el {itemId}, y que un {itemId} inválido o inexistente dé el error del ítem.
func TestAsignarYCompletar_ElSprintSeValidaAntesQueElItem(t *testing.T) {
	s := newUUIDStack(t)
	sprintID := s.sprint(t, "Sprint 1").PublicID.String()
	for _, suffix := range []string{"", "/completar"} {
		rec := s.serve(http.MethodPost, "/api/sprints/abc/items/xyz"+suffix, "")
		expectErrorBody(t, "sprint e ítem inválidos"+suffix, rec, http.StatusBadRequest, "el id del sprint debe ser un UUID")

		rec = s.serve(http.MethodPost, "/api/sprints/"+sprintID+"/items/xyz"+suffix, "")
		expectErrorBody(t, "ítem inválido"+suffix, rec, http.StatusBadRequest, "el id del ítem debe ser un UUID")

		rec = s.serve(http.MethodPost, "/api/sprints/"+sprintID+"/items/"+missingPokerItemID+suffix, "")
		expectErrorBody(t, "ítem inexistente"+suffix, rec, http.StatusNotFound, backlog.ErrItemNotFound.Error())
	}
}

// HU-52 · CA-52.3 · Verifica que las respuestas identifiquen al sprint con su UUID: el id al listar y al cerrar, el sprint_id de sus métricas, el sprint_actual del resumen, las métricas por sprint y el reporte del sprint.
func TestRespuestasDeSprint_IdentificanElSprintConSuUUID(t *testing.T) {
	s := newUUIDStack(t)
	project := "/api/proyectos/" + s.project.PublicID.String()
	want := s.sprint(t, "Sprint 1").PublicID.String()

	var list []struct {
		ID json.RawMessage `json:"id"`
	}
	rec := s.serve(http.MethodGet, project+"/sprints", "")
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &list)
	if len(list) != 1 || string(list[0].ID) != quoted(want) {
		t.Errorf("sprints del proyecto: %s, se esperaba un sprint con id %q", rec.Body.String(), want)
	}

	rec = s.serve(http.MethodGet, "/api/sprints/"+want+"/metricas", "")
	expectStatus(t, rec, http.StatusOK)
	if got := jsonString(t, "métricas del sprint", rec, "sprint_id"); got != want {
		t.Errorf("métricas del sprint: sprint_id = %q, se esperaba %q", got, want)
	}

	var summary struct {
		SprintActual struct {
			ID json.RawMessage `json:"id"`
		} `json:"sprint_actual"`
	}
	rec = s.serve(http.MethodGet, project+"/resumen", "")
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &summary)
	if string(summary.SprintActual.ID) != quoted(want) {
		t.Errorf("resumen: sprint_actual.id = %s, se esperaba %q", summary.SprintActual.ID, want)
	}

	type sprintWithMetrics struct {
		Sprint struct {
			ID json.RawMessage `json:"id"`
		} `json:"sprint"`
		Metricas struct {
			SprintID json.RawMessage `json:"sprint_id"`
		} `json:"metricas"`
	}
	var perSprint []sprintWithMetrics
	rec = s.serve(http.MethodGet, project+"/metricas/sprints", "")
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &perSprint)
	if len(perSprint) != 1 || string(perSprint[0].Sprint.ID) != quoted(want) || string(perSprint[0].Metricas.SprintID) != quoted(want) {
		t.Errorf("métricas por sprint: %s, se esperaba el sprint %q en sprint.id y en metricas.sprint_id", rec.Body.String(), want)
	}

	var report sprintWithMetrics
	rec = s.serve(http.MethodGet, "/api/sprints/"+want+"/reporte", "")
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &report)
	if string(report.Sprint.ID) != quoted(want) || string(report.Metricas.SprintID) != quoted(want) {
		t.Errorf("reporte del sprint: sprint.id = %s y metricas.sprint_id = %s, se esperaba %q", report.Sprint.ID, report.Metricas.SprintID, want)
	}

	rec = s.serve(http.MethodPost, "/api/sprints/"+want+"/cerrar", "")
	expectStatus(t, rec, http.StatusOK)
	if got := jsonString(t, "sprint cerrado", rec, "id"); got != want {
		t.Errorf("sprint cerrado: id = %q, se esperaba %q", got, want)
	}
}
