package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-51 · Identificadores UUID, parte de proyectos (Issue #193)

// projectRoute es una de las 17 rutas que reciben un proyecto (CA-51.3), con
// un cuerpo válido y el código que responde con un proyecto existente.
type projectRoute struct {
	method, tail, body string
	ok                 int
}

// projectRoutes son las rutas de proyecto, en un orden en el que cada una
// funciona con el proyecto que crea newHandlerForProjectRoutes.
var projectRoutes = []projectRoute{
	{http.MethodGet, "", "", http.StatusOK},
	{http.MethodPut, "", `{"nombre":"Inventario","descripcion":"Depósito","estado":"EN_CURSO"}`, http.StatusOK},
	{http.MethodPatch, "/fecha-inicio", `{"fecha_inicio":"2026-10-01"}`, http.StatusOK},
	{http.MethodPatch, "/fecha-fin", `{"fecha_fin":"2026-12-18"}`, http.StatusOK},
	{http.MethodPatch, "/fechas", `{"fecha_inicio":"2026-10-02","fecha_fin":"2026-12-19"}`, http.StatusOK},
	{http.MethodPost, "/integrantes", validMember, http.StatusCreated},
	{http.MethodGet, "/integrantes", "", http.StatusOK},
	{http.MethodPost, "/backlog", validItem, http.StatusCreated},
	{http.MethodGet, "/backlog", "", http.StatusOK},
	{http.MethodPost, "/sprints", validSprint, http.StatusCreated},
	{http.MethodGet, "/sprints", "", http.StatusOK},
	{http.MethodGet, "/defectos", "", http.StatusOK},
	{http.MethodGet, "/metricas", "", http.StatusOK},
	{http.MethodGet, "/metricas/sprints", "", http.StatusOK},
	{http.MethodGet, "/resumen", "", http.StatusOK},
	{http.MethodGet, "/reporte", "", http.StatusOK},
	{http.MethodGet, "/reporte/pdf", "", http.StatusOK},
}

// serveRaw hace un pedido sin exigir una respuesta JSON (el PDF no lo es).
func serveRaw(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

// expectErrorBody verifica el código y el mensaje de error de la respuesta,
// sin exigir el Content-Type (lo verifica do en los demás tests).
func expectErrorBody(t *testing.T, what string, rec *httptest.ResponseRecorder, status int, msg string) {
	t.Helper()
	if rec.Code != status {
		t.Errorf("%s: código = %d, se esperaba %d; cuerpo: %s", what, rec.Code, status, rec.Body.String())
		return
	}
	var body errorJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Errorf("%s: cuerpo no es JSON válido (%v): %s", what, err, rec.Body.String())
		return
	}
	if body.Error != msg {
		t.Errorf("%s: error = %q, se esperaba %q", what, body.Error, msg)
	}
}

// HU-51 · CA-51.1 · Verifica que POST /api/proyectos responda 201 con el id como texto: un UUID versión 4 en minúsculas, y no un número.
func TestCrearProyecto_IDEsUUIDVersion4(t *testing.T) {
	h := newHandler()

	rec := do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"Sistema de ventas","descripcion":""}`)

	expectStatus(t, rec, http.StatusCreated)
	var id string
	if err := json.Unmarshal(rawProjectJSON(t, rec), &id); err != nil {
		t.Fatalf("id = %s, se esperaba un texto", rawProjectJSON(t, rec))
	}
	if !uuidV4Canonical.MatchString(id) {
		t.Errorf("id = %q, se esperaba un UUID versión 4 en minúsculas", id)
	}
}

// HU-51 · CA-51.2 · Verifica que dos proyectos creados por la API tengan ids distintos y que ninguno sea un número.
func TestCrearProyecto_DosProyectosTienenUUIDDistintos(t *testing.T) {
	h := newHandler()
	first := createProject(t, h, "Sistema de ventas")
	second := createProject(t, h, "Inventario")

	if first == second {
		t.Errorf("los dos proyectos tienen el id %q", first)
	}
	for _, id := range []string{first, second} {
		if !uuidV4Canonical.MatchString(id) {
			t.Errorf("id = %q, se esperaba un UUID versión 4", id)
		}
	}
}

// HU-51 · CA-51.3 · Verifica que las 17 rutas de proyecto acepten el UUID del proyecto y respondan como antes con un proyecto existente.
func TestRutasDeProyecto_AceptanElUUID(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Inventario")
	base := projectPath(t, h, 1)
	if !uuidV4Canonical.MatchString(strings.TrimPrefix(base, "/api/proyectos/")) {
		t.Fatalf("ruta del proyecto = %q, se esperaba un UUID versión 4", base)
	}

	for _, r := range projectRoutes {
		rec := serveRaw(h, r.method, base+r.tail, r.body)
		if rec.Code != r.ok {
			t.Errorf("%s %s: código = %d, se esperaba %d; cuerpo: %s", r.method, base+r.tail, rec.Code, r.ok, rec.Body.String())
		}
	}
}

// HU-51 · CA-51.3 · Verifica que las rutas de proyecto acepten el UUID en mayúsculas y respondan con el id en minúsculas.
func TestRutasDeProyecto_AceptanElUUIDEnMayusculas(t *testing.T) {
	h := newHandler()
	id := createProject(t, h, "Inventario")
	if !uuidV4Canonical.MatchString(id) {
		t.Fatalf("id = %q, se esperaba un UUID versión 4", id)
	}
	upper := "/api/proyectos/" + strings.ToUpper(id)

	rec := do(t, h, http.MethodGet, upper, "")

	expectStatus(t, rec, http.StatusOK)
	if got := rawProjectID(t, rec); got != id {
		t.Errorf("id = %q, se esperaba %q en minúsculas", got, id)
	}
	for _, r := range projectRoutes[1:] {
		if rec := serveRaw(h, r.method, upper+r.tail, r.body); rec.Code != r.ok {
			t.Errorf("%s %s: código = %d, se esperaba %d; cuerpo: %s", r.method, upper+r.tail, rec.Code, r.ok, rec.Body.String())
		}
	}
}

// HU-51 · CA-51.4 · Verifica que en las 17 rutas de proyecto un id numérico (también 0 y negativo) responda 400 con "el id del proyecto debe ser un UUID", aunque exista un proyecto.
func TestRutasDeProyecto_IDNumericoDevuelve400(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Inventario")

	for _, id := range []string{"1", "0", "-1"} {
		for _, r := range projectRoutes {
			path := "/api/proyectos/" + id + r.tail
			expectErrorBody(t, r.method+" "+path, serveRaw(h, r.method, path, r.body), http.StatusBadRequest, "el id del proyecto debe ser un UUID")
		}
	}
}

// HU-51 · CA-51.4 · Verifica que el UUID de un proyecto existente escrito entre llaves, con urn:uuid: o sin guiones responda 400 en las rutas de proyecto.
func TestRutasDeProyecto_OtrasFormasDelUUIDDevuelven400(t *testing.T) {
	h := newHandler()
	id := createProject(t, h, "Inventario")

	for _, other := range []string{"{" + id + "}", "urn:uuid:" + id, strings.ReplaceAll(id, "-", "")} {
		for _, r := range projectRoutes {
			path := "/api/proyectos/" + other + r.tail
			expectErrorBody(t, r.method+" "+path, serveRaw(h, r.method, path, r.body), http.StatusBadRequest, "el id del proyecto debe ser un UUID")
		}
	}
}

// HU-51 · CA-51.4 · Verifica que en las 17 rutas de proyecto un UUID que nunca se asignó, o el UUID nulo, responda 404 con ErrProjectNotFound.
func TestRutasDeProyecto_UUIDInexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Inventario")

	for _, id := range []string{missingProjectID, nilProjectID} {
		for _, r := range projectRoutes {
			path := "/api/proyectos/" + id + r.tail
			expectErrorBody(t, r.method+" "+path, serveRaw(h, r.method, path, r.body), http.StatusNotFound, project.ErrProjectNotFound.Error())
		}
	}
}

// HU-51 · CA-51.4 · Verifica que el id de la ruta se valide antes que el cuerpo: un id inválido con un cuerpo que no es JSON responde el error del id (orden de HU-18).
func TestOrdenDeErrores_IDInvalidoAntesQueElCuerpo(t *testing.T) {
	h := newHandler()

	for _, r := range projectRoutes {
		if r.body == "" {
			continue
		}
		path := "/api/proyectos/abc" + r.tail
		expectErrorBody(t, r.method+" "+path, serveRaw(h, r.method, path, `no es json`), http.StatusBadRequest, "el id del proyecto debe ser un UUID")
	}
}

// HU-51 · CA-51.4 · Verifica que, con un UUID inexistente, las validaciones de contenido se informen antes que la inexistencia, como con un ID entero (HU-18, HU-41, HU-19 y HU-04).
func TestOrdenDeErrores_UUIDInexistenteConservaElOrden(t *testing.T) {
	h := newHandler()
	base := "/api/proyectos/" + missingProjectID
	cases := []struct {
		method, path, body string
		want               error
	}{
		{http.MethodPatch, base + "/fecha-inicio", `{"fecha_inicio":""}`, project.ErrMissingStartDate},
		{http.MethodPatch, base + "/fecha-fin", `{"fecha_fin":""}`, project.ErrMissingEndDate},
		{http.MethodPatch, base + "/fechas", `{"fecha_inicio":"","fecha_fin":""}`, project.ErrMissingStartDate},
		{http.MethodPost, base + "/backlog", `{"titulo":"  ","prioridad":"ALTA","criterios":["Ok"]}`, backlog.ErrEmptyTitle},
	}
	for _, c := range cases {
		expectErrorBody(t, c.method+" "+c.path, serveRaw(h, c.method, c.path, c.body), http.StatusBadRequest, c.want.Error())
	}
}

// HU-51 · CA-51.4 · Verifica que las rutas de las demás entidades conserven su ID entero y el mensaje "el id debe ser un número entero" (parte pendiente de HU-51), salvo las de ítems, integrantes, sprints y esfuerzos, que con HU-52 reciben su UUID.
func TestRutasDeOtrasEntidades_ConservanElIDEntero(t *testing.T) {
	h := newHandler()

	// HU-52, CA-52.4: las rutas de esfuerzos reciben el UUID del ítem (parte
	// de García Santiago, efforts.go).
	path := "/api/items/abc/esfuerzos"
	expectErrorBody(t, "GET "+path, serveRaw(h, http.MethodGet, path, ""), http.StatusBadRequest, "el id del ítem debe ser un UUID")
	// HU-52, CA-52.4: los sprints ya no tienen ID entero (parte de Buttini
	// Cristobal).
	for _, path := range []string{"/api/sprints/abc/metricas", "/api/sprints/abc/reporte"} {
		expectErrorBody(t, "GET "+path, serveRaw(h, http.MethodGet, path, ""), http.StatusBadRequest, "el id del sprint debe ser un UUID")
	}
	// HU-52, CA-52.4: los ítems y los integrantes ya no tienen ID entero
	// (cierra esa parte del pendiente R1 de HU-51).
	path = "/api/items/abc"
	expectErrorBody(t, "GET "+path, serveRaw(h, http.MethodGet, path, ""), http.StatusBadRequest, "el id del ítem debe ser un UUID")
	path = "/api/integrantes/abc"
	expectErrorBody(t, "PUT "+path, serveRaw(h, http.MethodPut, path, validMember), http.StatusBadRequest, "el id del integrante debe ser un UUID")
}

// projectIDsIn recorre un valor JSON y devuelve, como texto JSON crudo, cada
// identificador de proyecto que encuentra: los "proyecto_id" y el "id" de
// los objetos "proyecto".
func projectIDsIn(v any) []string {
	var found []string
	var walk func(v any, inProject bool)
	walk = func(v any, inProject bool) {
		switch x := v.(type) {
		case map[string]any:
			for k, child := range x {
				switch {
				case k == "proyecto_id", inProject && k == "id":
					b, _ := json.Marshal(child)
					found = append(found, string(b))
				default:
					walk(child, k == "proyecto")
				}
			}
		case []any:
			for _, child := range x {
				walk(child, inProject)
			}
		}
	}
	walk(v, false)
	return found
}

// expectProjectIDsAreUUID verifica que la respuesta tenga al menos min
// identificadores de proyecto y que todos sean el UUID want, como texto. Si
// isProject, la respuesta es un proyecto y su "id" también cuenta.
func expectProjectIDsAreUUID(t *testing.T, what string, rec *httptest.ResponseRecorder, want string, min int, isProject bool) {
	t.Helper()
	var body any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Errorf("%s: cuerpo no es JSON válido (%v): %s", what, err, rec.Body.String())
		return
	}
	if obj, ok := body.(map[string]any); ok && isProject {
		body = map[string]any{"proyecto": obj}
	}
	ids := projectIDsIn(body)
	if len(ids) < min {
		t.Errorf("%s: tiene %d identificadores de proyecto, se esperaban al menos %d: %s", what, len(ids), min, rec.Body.String())
	}
	for _, id := range ids {
		if id != `"`+want+`"` {
			t.Errorf("%s: identificador de proyecto = %s, se esperaba el UUID %q como texto", what, id, want)
		}
	}
}

// HU-51 · CA-51.3 · Verifica que ninguna respuesta de la API exponga el ID entero del proyecto: id y proyecto_id llevan su UUID en proyectos, ítems, sprints, integrantes, métricas, resumen y reportes.
func TestRespuestas_IdentificanElProyectoConSuUUID(t *testing.T) {
	h := newHandler()
	rec := do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"Inventario","descripcion":""}`)
	expectStatus(t, rec, http.StatusCreated)
	id := rawProjectID(t, rec)
	if !uuidV4Canonical.MatchString(id) {
		t.Errorf("POST /api/proyectos: id = %s, se esperaba un UUID versión 4 como texto", rawProjectJSON(t, rec))
	}
	base := "/api/proyectos/" + id

	steps := []struct {
		what, method, path, body string
		min                      int
	}{
		{"proyecto", http.MethodGet, base, "", 1},
		{"proyecto", http.MethodPut, base, `{"nombre":"Inventario","descripcion":"","estado":"EN_CURSO"}`, 1},
		{"proyecto", http.MethodPatch, base + "/fecha-inicio", `{"fecha_inicio":"2026-10-01"}`, 1},
		{"proyecto", http.MethodPatch, base + "/fecha-fin", `{"fecha_fin":"2026-12-18"}`, 1},
		{"proyecto", http.MethodPatch, base + "/fechas", `{"fecha_inicio":"2026-10-02","fecha_fin":"2026-12-19"}`, 1},
		{"lista de proyectos (HU-50)", http.MethodGet, "/api/proyectos", "", 0},
		{"ítem creado", http.MethodPost, base + "/backlog", validItem, 1},
		{"backlog", http.MethodGet, base + "/backlog", "", 1},
		{"backlog por prioridad", http.MethodGet, base + "/backlog?orden=prioridad", "", 1},
		{"ítem por su ID", http.MethodGet, "/api/items/1", "", 1},
		{"ítem modificado", http.MethodPut, "/api/items/1", validItem, 1},
		{"story points", http.MethodPatch, "/api/items/1/story-points", `{"story_points":5}`, 1},
		{"sprint creado", http.MethodPost, base + "/sprints", validSprint, 1},
		{"sprints del proyecto", http.MethodGet, base + "/sprints", "", 1},
		{"ítem asignado al sprint", http.MethodPost, "/api/sprints/1/items/1", "", 1},
		{"estado del ítem", http.MethodPatch, "/api/items/1/estado", `{"estado":"EN_CURSO"}`, 1},
		{"ítem completado", http.MethodPost, "/api/sprints/1/items/1/completar", "", 1},
		{"ítems completados del sprint", http.MethodGet, "/api/sprints/1/completados", "", 1},
		{"integrante registrado", http.MethodPost, base + "/integrantes", validMember, 1},
		{"integrantes", http.MethodGet, base + "/integrantes", "", 1},
		{"integrante modificado", http.MethodPut, "/api/integrantes/1", `{"nombre":"Ana Pérez","rol":"AGILE_ENABLER"}`, 1},
		{"métricas del proyecto", http.MethodGet, base + "/metricas", "", 1},
		{"métricas por sprint", http.MethodGet, base + "/metricas/sprints", "", 1},
		{"resumen", http.MethodGet, base + "/resumen", "", 2},
		{"reporte del proyecto", http.MethodGet, base + "/reporte", "", 3},
		{"reporte del sprint", http.MethodGet, "/api/sprints/1/reporte", "", 2},
		{"sprint cerrado", http.MethodPost, "/api/sprints/1/cerrar", "", 1},
	}
	for _, s := range steps {
		// HU-52: con do(), las rutas fijas de ítems e integrantes ("/api/items/1"
		// y "/api/integrantes/1") se traducen al id que devolvió la API, que
		// pasa a ser su UUID (uuidRoutes de entity_ids_test.go).
		rec := do(t, h, s.method, s.path, s.body)
		if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
			t.Errorf("%s (%s %s): código = %d; cuerpo: %s", s.what, s.method, s.path, rec.Code, rec.Body.String())
			continue
		}
		if s.what == "lista de proyectos (HU-50)" {
			var list []map[string]json.RawMessage
			if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list) != 1 {
				t.Errorf("%s: se esperaba un proyecto (%v): %s", s.what, err, rec.Body.String())
				continue
			}
			if got := string(list[0]["id"]); got != `"`+id+`"` {
				t.Errorf("%s: id = %s, se esperaba el UUID %q como texto", s.what, got, id)
			}
			continue
		}
		expectProjectIDsAreUUID(t, s.what+" ("+s.method+" "+s.path+")", rec, id, s.min, s.what == "proyecto")
	}
}

// HU-51 · CA-51.3 · Verifica que el PDF del reporte se descargue con el nombre reporte-proyecto-{UUID}.pdf.
func TestReportePDF_NombreDelArchivoConElUUID(t *testing.T) {
	h := newHandler()
	id := createProject(t, h, "Inventario")
	if !uuidV4Canonical.MatchString(id) {
		t.Errorf("id = %q, se esperaba un UUID versión 4", id)
	}

	rec := get(t, h, "/api/proyectos/"+id+"/reporte/pdf")

	expectStatus(t, rec, http.StatusOK)
	if got, want := rec.Header().Get("Content-Disposition"), `attachment; filename="reporte-proyecto-`+id+`.pdf"`; got != want {
		t.Errorf("Content-Disposition = %q, se esperaba %q", got, want)
	}
}
