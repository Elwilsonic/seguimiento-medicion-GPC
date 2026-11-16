package httpapi_test

import (
	"encoding/json"
	"fmt"
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

// HU-52 · Identificadores UUID del resto de las entidades, parte de defectos
// (Issue #197). Prueba las 3 rutas de defects.go (D3 de design.md): registrar,
// cambiar el estado y listar por sprint, más la lista por proyecto, que
// muestra los defectos.
//
// El proyecto, el ítem, el sprint y su asignación se crean con el dominio
// (defectEffortWorld), así que estos tests no dependen de que las rutas de
// ítems y sprints, que son de otros responsables, ya usen el UUID.

// neverAssignedUUID es un UUID versión 4 que nunca se asigna a un ítem, un
// sprint, un defecto ni un esfuerzo (HU-52, CA-52.4).
const neverAssignedUUID = "9d3f6b2a-5c1e-4f8a-b7d4-2e6c0a9f1b35"

// nilUUID es el UUID nulo: tiene la forma de un UUID, pero nunca se asigna.
const nilUUID = "00000000-0000-0000-0000-000000000000"

// numericText reconoce un texto que es solo un número entero.
var numericText = regexp.MustCompile(`^[0-9]+$`)

// defectEffortWorld es la API con un proyecto, un ítem y un sprint abierto,
// con el ítem asignado al sprint. Guarda los Manager de defectos y esfuerzos
// para registrar con el dominio lo que los escenarios dan por existente.
type defectEffortWorld struct {
	h       http.Handler
	defects *defect.Manager
	efforts *effort.Manager
	project project.Project
	item    backlog.Item
	sprint  sprint.Sprint
}

// buildDefectEffortWorld arma el mundo sin *testing.T, para que lo usen
// también los pasos de godog.
func buildDefectEffortWorld() (*defectEffortWorld, error) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items, anyMember{})
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items, anyMember{})
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	w := &defectEffortWorld{
		h:       httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects)),
		defects: defects,
		efforts: efforts,
	}

	var err error
	if w.project, err = projects.Create("Sistema de ventas", ""); err != nil {
		return nil, fmt.Errorf("crear el proyecto: %w", err)
	}
	if w.item, err = items.Create(w.project.ID, "Login de usuarios", "Acceso", "ALTA", []string{"Credenciales válidas"}); err != nil {
		return nil, fmt.Errorf("crear el ítem: %w", err)
	}
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	if w.sprint, err = sprints.Create(w.project.ID, "Sprint 1", start, end, "Gestionar el backlog"); err != nil {
		return nil, fmt.Errorf("crear el sprint: %w", err)
	}
	if w.item, err = sprints.AssignItem(w.sprint.ID, w.item.ID); err != nil {
		return nil, fmt.Errorf("asignar el ítem al sprint: %w", err)
	}
	return w, nil
}

func newDefectEffortWorld(t *testing.T) *defectEffortWorld {
	t.Helper()
	w, err := buildDefectEffortWorld()
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func (w *defectEffortWorld) projectUUID() string { return w.project.PublicID.String() }
func (w *defectEffortWorld) itemUUID() string    { return w.item.PublicID.String() }
func (w *defectEffortWorld) sprintUUID() string  { return w.sprint.PublicID.String() }

// registerDefect registra con el dominio un defecto ALTA del ítem, detectado
// en el sprint, y devuelve su UUID.
func (w *defectEffortWorld) registerDefect() (string, error) {
	d, err := w.defects.Register(w.item.ID, "El total no descuenta el descuento", "ALTA", w.sprint.ID)
	if err != nil {
		return "", fmt.Errorf("registrar el defecto: %w", err)
	}
	return d.PublicID.String(), nil
}

// startCorrection pasa el defecto a EN_CORRECCION con el dominio.
func (w *defectEffortWorld) startCorrection() error {
	defects, err := w.defects.ListBySprint(w.sprint.ID)
	if err != nil {
		return err
	}
	for _, d := range defects {
		if _, err := w.defects.ChangeStatus(d.ID, "EN_CORRECCION", 0); err != nil {
			return fmt.Errorf("pasar el defecto a EN_CORRECCION: %w", err)
		}
	}
	return nil
}

// registerEffort registra con el dominio un esfuerzo del ítem en el sprint y
// devuelve su UUID.
func (w *defectEffortWorld) registerEffort(hours float64) (string, error) {
	date := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
	e, err := w.efforts.Register(w.sprint.ID, w.item.ID, "García Santiago", date, "Implementar el registro", hours)
	if err != nil {
		return "", fmt.Errorf("registrar el esfuerzo: %w", err)
	}
	return e.PublicID.String(), nil
}

func (w *defectEffortWorld) serve(method, path, body string) *httptest.ResponseRecorder {
	return serveRaw(w.h, method, path, body)
}

// mustRegisterDefect registra el defecto con el dominio o termina el test.
func (w *defectEffortWorld) mustRegisterDefect(t *testing.T) string {
	t.Helper()
	id, err := w.registerDefect()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// defectBody es el cuerpo de un registro válido con ese sprint de detección,
// como texto (HU-52: el campo recibe el UUID del sprint).
func defectBody(sprintID string) string {
	return fmt.Sprintf(`{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":%q}`, sprintID)
}

// textOf devuelve el valor JSON como texto, y falla si es un número u otro
// tipo: HU-52 no expone el ID entero (CA-52.3).
func textOf(t *testing.T, what string, raw json.RawMessage) string {
	t.Helper()
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		t.Fatalf("%s = %s, se esperaba un texto", what, raw)
	}
	return text
}

// responseText devuelve el campo de la respuesta, que tiene que ser texto.
func responseText(t *testing.T, rec *httptest.ResponseRecorder, field string) string {
	t.Helper()
	raw, err := jsonField(rec, field)
	if err != nil {
		t.Fatal(err)
	}
	return textOf(t, field, raw)
}

// expectUUIDv4 verifica que el texto sea un UUID versión 4 canónico en
// minúsculas, que no sea un número ni el UUID nulo (CA-52.1).
func expectUUIDv4(t *testing.T, what, got string) {
	t.Helper()
	if numericText.MatchString(got) || got == nilUUID {
		t.Errorf("%s = %q, no debería ser un número ni el UUID nulo", what, got)
		return
	}
	if !uuidV4Canonical.MatchString(got) {
		t.Errorf("%s = %q, se esperaba un UUID versión 4 en minúsculas", what, got)
	}
}

// listElements devuelve los elementos de la lista key de la respuesta.
func listElements(t *testing.T, rec *httptest.ResponseRecorder, key string) []map[string]json.RawMessage {
	t.Helper()
	raw, err := jsonField(rec, key)
	if err != nil {
		t.Fatal(err)
	}
	var elements []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &elements); err != nil {
		t.Fatalf("%s = %s, se esperaba una lista de objetos", key, raw)
	}
	return elements
}

// HU-52 · CA-52.1 · Verifica que POST /api/items/{id}/defectos con el UUID del ítem responda 201 con el id del defecto como UUID versión 4, item_id y sprint_deteccion_id con los UUID del ítem y del sprint, y sprint_resolucion_id vacío.
func TestRegistrarDefecto_IDEsUUIDVersion4YReferenciasSonUUID(t *testing.T) {
	w := newDefectEffortWorld(t)

	rec := w.serve(http.MethodPost, "/api/items/"+w.itemUUID()+"/defectos", defectBody(w.sprintUUID()))

	expectStatus(t, rec, http.StatusCreated)
	expectUUIDv4(t, "id", responseText(t, rec, "id"))
	if got := responseText(t, rec, "item_id"); got != w.itemUUID() {
		t.Errorf("item_id = %q, se esperaba el UUID del ítem %q", got, w.itemUUID())
	}
	if got := responseText(t, rec, "sprint_deteccion_id"); got != w.sprintUUID() {
		t.Errorf("sprint_deteccion_id = %q, se esperaba el UUID del sprint %q", got, w.sprintUUID())
	}
	if got := responseText(t, rec, "sprint_resolucion_id"); got != "" {
		t.Errorf(`sprint_resolucion_id = %q, se esperaba "" (sin resolver)`, got)
	}
}

// HU-52 · CA-52.1 · Verifica que el primer defecto registrado no tenga el id 1 ni ningún número, ni el UUID nulo.
func TestRegistrarDefecto_PrimerDefectoNoEsUnNumero(t *testing.T) {
	w := newDefectEffortWorld(t)

	rec := w.serve(http.MethodPost, "/api/items/"+w.itemUUID()+"/defectos", defectBody(w.sprintUUID()))

	expectStatus(t, rec, http.StatusCreated)
	expectUUIDv4(t, "id del primer defecto", responseText(t, rec, "id"))
}

// HU-52 · CA-52.2 · Verifica que dos defectos registrados por la API tengan UUID distintos.
func TestRegistrarDefecto_DosDefectosTienenUUIDDistintos(t *testing.T) {
	w := newDefectEffortWorld(t)
	path := "/api/items/" + w.itemUUID() + "/defectos"

	first := w.serve(http.MethodPost, path, defectBody(w.sprintUUID()))
	second := w.serve(http.MethodPost, path, defectBody(w.sprintUUID()))

	expectStatus(t, first, http.StatusCreated)
	expectStatus(t, second, http.StatusCreated)
	a, b := responseText(t, first, "id"), responseText(t, second, "id")
	if a == b {
		t.Errorf("los dos defectos tienen el id %q", a)
	}
}

// HU-52 · CA-52.3 · Verifica que PATCH /api/defectos/{id}/estado con el UUID del defecto lo resuelva con el UUID de un sprint en sprint_resolucion_id, y responda con los mismos UUID.
func TestCambiarEstadoDefecto_ResuelvePorUUID(t *testing.T) {
	w := newDefectEffortWorld(t)
	id := w.mustRegisterDefect(t)
	path := "/api/defectos/" + id + "/estado"
	expectStatus(t, w.serve(http.MethodPatch, path, `{"estado":"EN_CORRECCION"}`), http.StatusOK)

	rec := w.serve(http.MethodPatch, path, fmt.Sprintf(`{"estado":"RESUELTO","sprint_resolucion_id":%q}`, w.sprintUUID()))

	expectStatus(t, rec, http.StatusOK)
	if got := responseText(t, rec, "estado"); got != string(defect.StatusResolved) {
		t.Errorf("estado = %q, se esperaba RESUELTO", got)
	}
	if got := responseText(t, rec, "id"); got != id {
		t.Errorf("id = %q, se esperaba el UUID del defecto %q", got, id)
	}
	if got := responseText(t, rec, "sprint_resolucion_id"); got != w.sprintUUID() {
		t.Errorf("sprint_resolucion_id = %q, se esperaba el UUID del sprint %q", got, w.sprintUUID())
	}
}

// HU-52 · CA-52.3 · Verifica que PATCH /api/defectos/{id}/estado acepte el UUID del defecto en mayúsculas y responda con el id en minúsculas.
func TestCambiarEstadoDefecto_AceptaElUUIDEnMayusculas(t *testing.T) {
	w := newDefectEffortWorld(t)
	id := w.mustRegisterDefect(t)

	rec := w.serve(http.MethodPatch, "/api/defectos/"+strings.ToUpper(id)+"/estado", `{"estado":"EN_CORRECCION"}`)

	expectStatus(t, rec, http.StatusOK)
	if got := responseText(t, rec, "id"); got != id {
		t.Errorf("id = %q, se esperaba %q en minúsculas", got, id)
	}
}

// HU-52 · CA-52.3 · Verifica que las listas de defectos por proyecto y por el UUID del sprint muestren cada defecto con su UUID, los UUID de su ítem y su sprint de detección, y "" en sprint_resolucion_id.
func TestListarDefectos_MuestranLosUUID(t *testing.T) {
	w := newDefectEffortWorld(t)
	id := w.mustRegisterDefect(t)

	for _, path := range []string{
		"/api/proyectos/" + w.projectUUID() + "/defectos",
		"/api/sprints/" + w.sprintUUID() + "/defectos",
	} {
		rec := w.serve(http.MethodGet, path, "")
		expectStatus(t, rec, http.StatusOK)
		elements := listElements(t, rec, "defectos")
		if len(elements) != 1 {
			t.Errorf("GET %s: %d defectos, se esperaba 1", path, len(elements))
			continue
		}
		d := elements[0]
		want := map[string]string{"id": id, "item_id": w.itemUUID(), "sprint_deteccion_id": w.sprintUUID(), "sprint_resolucion_id": ""}
		for field, value := range want {
			if got := textOf(t, path+" "+field, d[field]); got != value {
				t.Errorf("GET %s: %s = %q, se esperaba %q", path, field, got, value)
			}
		}
	}
}

// defectRoutesWith devuelve las 3 rutas de defects.go con un ID mal escrito
// en el lugar de cada entidad, y el mensaje de error de esa entidad (CA-52.4).
func defectRoutesWith(w *defectEffortWorld, bad string) []struct{ method, path, body, msg string } {
	return []struct{ method, path, body, msg string }{
		{http.MethodPost, "/api/items/" + bad + "/defectos", defectBody(w.sprintUUID()), "el id del ítem debe ser un UUID"},
		{http.MethodPatch, "/api/defectos/" + bad + "/estado", `{"estado":"EN_CORRECCION"}`, "el id del defecto debe ser un UUID"},
		{http.MethodGet, "/api/sprints/" + bad + "/defectos", "", "el id del sprint debe ser un UUID"},
	}
}

// HU-52 · CA-52.4 · Verifica que las 3 rutas de defectos respondan 400 con el mensaje de su entidad cuando el ID es un número, está entre llaves, tiene urn:uuid: adelante o no tiene guiones.
func TestRutasDeDefectos_IDQueNoEsUUIDDevuelve400(t *testing.T) {
	w := newDefectEffortWorld(t)
	id := w.mustRegisterDefect(t)

	for _, bad := range []string{"1", "{" + id + "}", "urn:uuid:" + id, strings.ReplaceAll(id, "-", "")} {
		for _, r := range defectRoutesWith(w, bad) {
			expectErrorBody(t, r.method+" "+r.path, w.serve(r.method, r.path, r.body), http.StatusBadRequest, r.msg)
		}
	}
}

// HU-52 · CA-52.4 · Verifica que las 3 rutas de defectos respondan 404 con el mensaje de su entidad cuando el UUID no existe o es el nulo.
func TestRutasDeDefectos_UUIDInexistenteDevuelve404(t *testing.T) {
	w := newDefectEffortWorld(t)
	w.mustRegisterDefect(t)
	notFound := map[string]string{
		http.MethodPost:  "no existe un ítem del backlog con ese id",
		http.MethodPatch: "no existe un defecto con ese id",
		http.MethodGet:   "no existe un sprint con ese id",
	}

	for _, missing := range []string{neverAssignedUUID, nilUUID} {
		for _, r := range defectRoutesWith(w, missing) {
			expectErrorBody(t, r.method+" "+r.path, w.serve(r.method, r.path, r.body), http.StatusNotFound, notFound[r.method])
		}
	}
}

// HU-52 · CA-52.4 · Verifica que el {id} del defecto se valide antes que el cuerpo: un ID que no es UUID con un cuerpo que no es JSON da el 400 del ID.
func TestCambiarEstadoDefecto_IDSeValidaAntesQueElCuerpo(t *testing.T) {
	w := newDefectEffortWorld(t)

	rec := w.serve(http.MethodPatch, "/api/defectos/abc/estado", "esto no es JSON")

	expectErrorBody(t, "PATCH /api/defectos/abc/estado", rec, http.StatusBadRequest, "el id del defecto debe ser un UUID")
}

// HU-52 · CA-52.4 · Verifica que un sprint_deteccion_id que es un texto sin la forma de un UUID responda 400 con el mensaje del sprint.
func TestRegistrarDefecto_SprintDeteccionQueNoEsUUIDDevuelve400(t *testing.T) {
	w := newDefectEffortWorld(t)
	path := "/api/items/" + w.itemUUID() + "/defectos"

	for _, bad := range []string{"3", "abc", "{" + w.sprintUUID() + "}", strings.ReplaceAll(w.sprintUUID(), "-", "")} {
		rec := w.serve(http.MethodPost, path, defectBody(bad))
		expectErrorBody(t, "sprint_deteccion_id "+bad, rec, http.StatusBadRequest, "el id del sprint debe ser un UUID")
	}
}

// HU-52 · CA-52.4 · Verifica que un sprint_deteccion_id ausente o vacío siga siendo un sprint faltante: 400 de HU-29.
func TestRegistrarDefecto_SprintDeteccionAusenteOVacioDevuelve400(t *testing.T) {
	w := newDefectEffortWorld(t)
	path := "/api/items/" + w.itemUUID() + "/defectos"

	for _, body := range []string{
		`{"descripcion":"Sin sprint","severidad":"ALTA"}`,
		`{"descripcion":"Sin sprint","severidad":"ALTA","sprint_deteccion_id":""}`,
	} {
		expectErrorBody(t, body, w.serve(http.MethodPost, path, body), http.StatusBadRequest, defect.ErrMissingDetectionSprint.Error())
	}
}

// HU-52 · CA-52.4 · Verifica que un sprint_deteccion_id con forma de UUID que no existe (o el nulo) responda 404 del sprint, y no el 400 de sprint faltante.
func TestRegistrarDefecto_SprintDeteccionInexistenteDevuelve404(t *testing.T) {
	w := newDefectEffortWorld(t)
	path := "/api/items/" + w.itemUUID() + "/defectos"

	for _, missing := range []string{neverAssignedUUID, nilUUID} {
		rec := w.serve(http.MethodPost, path, defectBody(missing))
		expectErrorBody(t, "sprint_deteccion_id "+missing, rec, http.StatusNotFound, sprint.ErrSprintNotFound.Error())
	}
}

// HU-52 · CA-52.4 · Verifica que al resolver, un sprint_resolucion_id que no es UUID dé 400 del sprint, uno inexistente dé 404 del sprint, y uno vacío dé el 400 de sprint de resolución faltante de HU-30.
func TestCambiarEstadoDefecto_SprintResolucionPorUUID(t *testing.T) {
	w := newDefectEffortWorld(t)
	id := w.mustRegisterDefect(t)
	path := "/api/defectos/" + id + "/estado"
	expectStatus(t, w.serve(http.MethodPatch, path, `{"estado":"EN_CORRECCION"}`), http.StatusOK)

	cases := []struct {
		sprintID string
		status   int
		msg      string
	}{
		{"3", http.StatusBadRequest, "el id del sprint debe ser un UUID"},
		{"abc", http.StatusBadRequest, "el id del sprint debe ser un UUID"},
		{neverAssignedUUID, http.StatusNotFound, sprint.ErrSprintNotFound.Error()},
		{"", http.StatusBadRequest, defect.ErrMissingResolutionSprint.Error()},
	}
	for _, c := range cases {
		body := fmt.Sprintf(`{"estado":"RESUELTO","sprint_resolucion_id":%q}`, c.sprintID)
		expectErrorBody(t, "sprint_resolucion_id "+c.sprintID, w.serve(http.MethodPatch, path, body), c.status, c.msg)
	}
}

// HU-52 · CA-52.4 · Verifica que con un UUID de defecto inexistente se conserve el orden de errores de HU-30: RESUELTO sin sprint de resolución da 400, y no 404.
func TestCambiarEstadoDefecto_UUIDInexistenteConservaElOrdenDeErrores(t *testing.T) {
	w := newDefectEffortWorld(t)

	rec := w.serve(http.MethodPatch, "/api/defectos/"+neverAssignedUUID+"/estado", `{"estado":"RESUELTO"}`)

	expectErrorBody(t, "PATCH con UUID inexistente", rec, http.StatusBadRequest, defect.ErrMissingResolutionSprint.Error())
}
