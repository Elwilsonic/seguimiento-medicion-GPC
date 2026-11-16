package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197).
//
// Los tests de la API escriben rutas con IDs fijos, como "/api/items/1", que
// dejan de funcionar cuando cada entidad pase a recibir su UUID. Para no
// reescribir esas rutas en los 15 archivos (D9 de design.md), do() guarda, por
// handler y por entidad, el "id" que devolvió la API al crear cada ítem,
// sprint, integrante, defecto, esfuerzo y sesión, en el orden de creación, y
// traduce en la ruta el número N al "id" de la N-ésima entidad creada por la
// API. Mientras la API devuelva enteros correlativos, la N-ésima entidad
// creada tiene el ID N, así que la traducción es la identidad. Un número que
// no corresponde a una entidad creada por la API (por ejemplo, el 999 de un
// caso de inexistente, o una entidad guardada directo en el repositorio) se
// deja como está: esos casos los adapta el responsable de cada entidad.

// entityKind es el tipo de entidad de una ruta.
type entityKind string

const (
	kindItem    entityKind = "ítem"
	kindSprint  entityKind = "sprint"
	kindMember  entityKind = "integrante"
	kindDefect  entityKind = "defecto"
	kindEffort  entityKind = "esfuerzo"
	kindSession entityKind = "sesión"
)

var (
	entityIDsMu sync.Mutex
	entityIDs   = map[http.Handler]map[entityKind][]string{}
)

// creationRoutes son las rutas POST que crean cada entidad.
var creationRoutes = []struct {
	pattern *regexp.Regexp
	kind    entityKind
}{
	{regexp.MustCompile(`^/api/proyectos/[^/]+/backlog$`), kindItem},
	{regexp.MustCompile(`^/api/proyectos/[^/]+/sprints$`), kindSprint},
	{regexp.MustCompile(`^/api/proyectos/[^/]+/integrantes$`), kindMember},
	{regexp.MustCompile(`^/api/items/[^/]+/defectos$`), kindDefect},
	{regexp.MustCompile(`^/api/sprints/[^/]+/items/[^/]+/esfuerzos$`), kindEffort},
	{regexp.MustCompile(`^/api/items/[^/]+/poker-sessions$`), kindSession},
}

// pathSegments dice, para el primer segmento de una ruta /api/..., qué
// entidad nombra el segmento siguiente.
var pathSegments = map[string]entityKind{
	"items":          kindItem,
	"sprints":        kindSprint,
	"integrantes":    kindMember,
	"defectos":       kindDefect,
	"esfuerzos":      kindEffort,
	"poker-sessions": kindSession,
}

// serveRecording atiende el pedido y, si creó una entidad, guarda su "id".
func serveRecording(t *testing.T, h http.Handler, rec *httptest.ResponseRecorder, req *http.Request) {
	t.Helper()
	h.ServeHTTP(rec, req)
	if req.Method != http.MethodPost || rec.Code != http.StatusCreated {
		return
	}
	for _, route := range creationRoutes {
		if !route.pattern.MatchString(req.URL.Path) {
			continue
		}
		var body struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.ID == nil {
			return
		}
		entityIDsMu.Lock()
		defer entityIDsMu.Unlock()
		if entityIDs[h] == nil {
			entityIDs[h] = map[entityKind][]string{}
		}
		entityIDs[h][route.kind] = append(entityIDs[h][route.kind], strings.Trim(string(body.ID), `"`))
		return
	}
}

// uuidRoutes son las rutas cuyo handler ya recibe el UUID de sus entidades.
// do() traduce los números de una ruta solo si la ruta está en esta tabla:
// así, cuando un responsable pasa sus handlers a UUID, sus tests siguen
// funcionando con las rutas fijas, y en las rutas de los demás, que todavía
// leen un entero, el id se traduce al revés (withEntityIDs; desviaciones de
// la base, D9 de design.md).
//
// Cada responsable agrega, en el GREEN de su entidad, las rutas cuyos
// handlers están en sus archivos (tabla "Rutas por responsable" de D3), con
// el método y el patrón de http.ServeMux, por ejemplo "GET /api/items/{id}".
// En una ruta de la tabla se traducen todos los números que nombran una
// entidad (también el {itemId} de /api/sprints/{id}/items/{itemId}), porque
// el handler de esa ruta es del mismo responsable. Cada uno escribe solo en
// su bloque, para que los merges no choquen.
var uuidRoutes = []string{
	// Cardozo Leandro (5): GET y PUT /api/items/{id},
	// PATCH /api/items/{id}/estado y /api/items/{id}/story-points (items.go),
	// y PUT /api/integrantes/{id} (members_update.go).
	"GET /api/items/{id}",
	"PUT /api/items/{id}",
	"PATCH /api/items/{id}/estado",
	"PATCH /api/items/{id}/story-points",
	"PUT /api/integrantes/{id}",

	// Buttini Cristobal (11): POST /api/items/{id}/poker-sessions y
	// POST /api/poker-sessions/{id}/votes, /reveal, /estimate y /rounds
	// (poker.go); POST /api/sprints/{id}/items/{itemId} y
	// /api/sprints/{id}/items/{itemId}/completar, GET /api/sprints/{id}/completados
	// y POST /api/sprints/{id}/cerrar (sprints.go); GET /api/sprints/{id}/metricas
	// (metrics.go); GET /api/sprints/{id}/reporte (sprint_reports.go).
	"POST /api/items/{id}/poker-sessions",
	"POST /api/poker-sessions/{id}/votes",
	"POST /api/poker-sessions/{id}/reveal",
	"POST /api/poker-sessions/{id}/estimate",
	"POST /api/poker-sessions/{id}/rounds",
	"POST /api/sprints/{id}/items/{itemId}",
	"POST /api/sprints/{id}/items/{itemId}/completar",
	"GET /api/sprints/{id}/completados",
	"POST /api/sprints/{id}/cerrar",
	"GET /api/sprints/{id}/metricas",
	"GET /api/sprints/{id}/reporte",

	// García Santiago (7): POST /api/items/{id}/defectos,
	// PATCH /api/defectos/{id}/estado y GET /api/sprints/{id}/defectos
	// (defects.go); GET /api/items/{id}/esfuerzos,
	// POST /api/sprints/{id}/items/{itemId}/esfuerzos, GET /api/sprints/{id}/esfuerzos
	// y PATCH /api/esfuerzos/{id}/horas (efforts.go).
}

// inUUIDRoutes dice si el pedido method path (sin la query) corresponde a
// una ruta de uuidRoutes. Un segmento {…} del patrón acepta cualquier valor.
func inUUIDRoutes(method, path string) bool {
	parts := strings.Split(path, "/")
	for _, route := range uuidRoutes {
		routeMethod, pattern, _ := strings.Cut(route, " ")
		want := strings.Split(pattern, "/")
		if routeMethod != method || len(want) != len(parts) {
			continue
		}
		match := true
		for i := range want {
			if !strings.HasPrefix(want[i], "{") && want[i] != parts[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// withEntityIDs traduce en path los ids de las entidades. Entiende
// /api/{entidad}/{N}/... y el {itemId} de /api/sprints/{N}/items/{M}/...
//
//   - Si la ruta del pedido está en uuidRoutes, cada número N pasa al "id" de
//     la N-ésima entidad de ese tipo creada por la API en h.
//   - Si no está, es al revés (segunda desviación de la base, D9 de
//     design.md): un "id" que la API devolvió al crear una entidad pasa al
//     número N de su orden de creación, que es su ID interno, porque el
//     handler de esa ruta todavía lee un entero. Así, un test que arma la
//     ruta de otro responsable con el id que devolvió la API (por ejemplo,
//     el UUID de un ítem en POST /api/sprints/{id}/items/{itemId}) sigue
//     funcionando hasta que ese responsable pase su ruta a UUID.
//
// Las dos traducciones suponen lo mismo: que las entidades de ese tipo se
// crean por la API, en orden, así que la N-ésima tiene el ID interno N.
func withEntityIDs(h http.Handler, method, path string) string {
	if !strings.HasPrefix(path, "/api/") {
		return path
	}
	rest, query, _ := strings.Cut(path, "?")
	toUUID := inUUIDRoutes(method, rest)
	parts := strings.Split(rest, "/")
	entityIDsMu.Lock()
	defer entityIDsMu.Unlock()
	ids := entityIDs[h]
	for i := 2; i+1 < len(parts); i++ {
		kind, ok := pathSegments[parts[i]]
		if !ok || (i > 2 && !(parts[2] == "sprints" && parts[i] == "items")) {
			continue
		}
		if !toUUID {
			for n, id := range ids[kind] {
				if id == parts[i+1] {
					parts[i+1] = strconv.Itoa(n + 1)
					break
				}
			}
			continue
		}
		n, err := strconv.Atoi(parts[i+1])
		if err != nil || n < 1 || n > len(ids[kind]) {
			continue
		}
		parts[i+1] = ids[kind][n-1]
	}
	out := strings.Join(parts, "/")
	if query != "" {
		out += "?" + query
	}
	return out
}
