package httpapi_test

import (
	"net/http"
	"testing"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197).
// Tests del helper de rutas de los tests (entity_ids_test.go): do() traduce
// los números de una ruta solo si la ruta está en uuidRoutes.

// fakeItemUUID es el "id" que el test guarda como el del primer ítem, para
// que la traducción se note (la API todavía devuelve el 1).
const fakeItemUUID = "3f2b8c1e-5a4d-4e6f-9b7a-0c1d2e3f4a5b"

// withFakeItemID registra fakeItemUUID como el id del primer ítem de un
// handler nuevo y lo borra al terminar el test.
func withFakeItemID(t *testing.T) http.Handler {
	t.Helper()
	h := newHandler()
	entityIDsMu.Lock()
	entityIDs[h] = map[entityKind][]string{kindItem: {fakeItemUUID}}
	entityIDsMu.Unlock()
	t.Cleanup(func() {
		entityIDsMu.Lock()
		delete(entityIDs, h)
		entityIDsMu.Unlock()
	})
	return h
}

// HU-52 · CA-52.3 · Verifica que, con la tabla uuidRoutes vacía, do() no traduzca ninguna ruta: el número queda igual aunque haya un id guardado.
func TestWithEntityIDs_TablaVaciaNoTraduce(t *testing.T) {
	h := withFakeItemID(t)
	saved := uuidRoutes
	uuidRoutes = nil
	t.Cleanup(func() { uuidRoutes = saved })

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/api/items/1"},
		{http.MethodPatch, "/api/items/1/estado"},
		{http.MethodPost, "/api/sprints/1/items/1"},
	} {
		if got := withEntityIDs(h, c.method, c.path); got != c.path {
			t.Errorf("%s %s: ruta = %q, se esperaba sin cambios", c.method, c.path, got)
		}
	}
}

// HU-52 · CA-52.3 · Verifica que, con una ruta cargada en uuidRoutes, do() traduzca solo esa ruta (mismo método y patrón) y deje igual las demás.
func TestWithEntityIDs_TraduceSoloLasRutasDeLaTabla(t *testing.T) {
	h := withFakeItemID(t)
	saved := uuidRoutes
	uuidRoutes = []string{"GET /api/items/{id}"}
	t.Cleanup(func() { uuidRoutes = saved })

	cases := []struct{ method, path, want string }{
		{http.MethodGet, "/api/items/1", "/api/items/" + fakeItemUUID},
		{http.MethodGet, "/api/items/1?orden=prioridad", "/api/items/" + fakeItemUUID + "?orden=prioridad"},
		{http.MethodPut, "/api/items/1", "/api/items/1"},
		{http.MethodPatch, "/api/items/1/estado", "/api/items/1/estado"},
		{http.MethodGet, "/api/items/1/esfuerzos", "/api/items/1/esfuerzos"},
		{http.MethodPost, "/api/sprints/1/items/1", "/api/sprints/1/items/1"},
	}
	for _, c := range cases {
		if got := withEntityIDs(h, c.method, c.path); got != c.want {
			t.Errorf("%s %s: ruta = %q, se esperaba %q", c.method, c.path, got, c.want)
		}
	}
}
