package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-52 · Identificadores UUID del resto de las entidades, parte de esfuerzos
// (Issue #197). Prueba las 4 rutas de efforts.go (D3 de design.md):
// registrar, listar por historia, listar por sprint y corregir las horas. El
// mundo de los tests (defectEffortWorld) está en defects_uuid_test.go.

const uuidEffortBody = `{"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar el registro","horas":2.5}`

// effortsPathFor es la ruta del registro de esfuerzos con esos IDs.
func effortsPathFor(sprintID, itemID string) string {
	return "/api/sprints/" + sprintID + "/items/" + itemID + "/esfuerzos"
}

// mustRegisterEffort registra un esfuerzo con el dominio o termina el test.
func (w *defectEffortWorld) mustRegisterEffort(t *testing.T, hours float64) string {
	t.Helper()
	id, err := w.registerEffort(hours)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// HU-52 · CA-52.1 · Verifica que POST /api/sprints/{id}/items/{itemId}/esfuerzos con los UUID del sprint y del ítem responda 201 con el id del esfuerzo como UUID versión 4, y con los UUID del ítem y del sprint en item_id y sprint_id.
func TestRegistrarEsfuerzo_IDEsUUIDVersion4YReferenciasSonUUID(t *testing.T) {
	w := newDefectEffortWorld(t)

	rec := w.serve(http.MethodPost, effortsPathFor(w.sprintUUID(), w.itemUUID()), uuidEffortBody)

	expectStatus(t, rec, http.StatusCreated)
	expectUUIDv4(t, "id", responseText(t, rec, "id"))
	if got := responseText(t, rec, "item_id"); got != w.itemUUID() {
		t.Errorf("item_id = %q, se esperaba el UUID del ítem %q", got, w.itemUUID())
	}
	if got := responseText(t, rec, "sprint_id"); got != w.sprintUUID() {
		t.Errorf("sprint_id = %q, se esperaba el UUID del sprint %q", got, w.sprintUUID())
	}
}

// HU-52 · CA-52.1 · Verifica que el primer esfuerzo registrado no tenga el id 1 ni ningún número, ni el UUID nulo.
func TestRegistrarEsfuerzo_PrimerEsfuerzoNoEsUnNumero(t *testing.T) {
	w := newDefectEffortWorld(t)

	rec := w.serve(http.MethodPost, effortsPathFor(w.sprintUUID(), w.itemUUID()), uuidEffortBody)

	expectStatus(t, rec, http.StatusCreated)
	expectUUIDv4(t, "id del primer esfuerzo", responseText(t, rec, "id"))
}

// HU-52 · CA-52.2 · Verifica que dos esfuerzos registrados por la API tengan UUID distintos.
func TestRegistrarEsfuerzo_DosEsfuerzosTienenUUIDDistintos(t *testing.T) {
	w := newDefectEffortWorld(t)
	path := effortsPathFor(w.sprintUUID(), w.itemUUID())

	first := w.serve(http.MethodPost, path, uuidEffortBody)
	second := w.serve(http.MethodPost, path, uuidEffortBody)

	expectStatus(t, first, http.StatusCreated)
	expectStatus(t, second, http.StatusCreated)
	if a, b := responseText(t, first, "id"), responseText(t, second, "id"); a == b {
		t.Errorf("los dos esfuerzos tienen el id %q", a)
	}
}

// HU-52 · CA-52.3 · Verifica que PATCH /api/esfuerzos/{id}/horas con el UUID del esfuerzo corrija las horas y responda con el mismo UUID en id.
func TestCorregirEsfuerzo_CorrigePorUUID(t *testing.T) {
	w := newDefectEffortWorld(t)
	id := w.mustRegisterEffort(t, 2.5)

	rec := w.serve(http.MethodPatch, "/api/esfuerzos/"+id+"/horas", `{"horas":3}`)

	expectStatus(t, rec, http.StatusOK)
	if got := responseText(t, rec, "id"); got != id {
		t.Errorf("id = %q, se esperaba el UUID del esfuerzo %q", got, id)
	}
	var body struct {
		Horas float64 `json:"horas"`
	}
	decode(t, rec, &body)
	if body.Horas != 3 {
		t.Errorf("horas = %v, se esperaba 3", body.Horas)
	}
}

// HU-52 · CA-52.3 · Verifica que PATCH /api/esfuerzos/{id}/horas acepte el UUID del esfuerzo en mayúsculas y responda con el id en minúsculas.
func TestCorregirEsfuerzo_AceptaElUUIDEnMayusculas(t *testing.T) {
	w := newDefectEffortWorld(t)
	id := w.mustRegisterEffort(t, 2.5)

	rec := w.serve(http.MethodPatch, "/api/esfuerzos/"+strings.ToUpper(id)+"/horas", `{"horas":3}`)

	expectStatus(t, rec, http.StatusOK)
	if got := responseText(t, rec, "id"); got != id {
		t.Errorf("id = %q, se esperaba %q en minúsculas", got, id)
	}
}

// HU-52 · CA-52.3 · Verifica que las listas de esfuerzos por el UUID del sprint y por el UUID de la historia muestren cada esfuerzo con su UUID y con los UUID de su ítem y su sprint, sin ningún número.
func TestListarEsfuerzos_PorSprintYPorHistoriaMuestranLosUUID(t *testing.T) {
	w := newDefectEffortWorld(t)
	ids := map[string]bool{w.mustRegisterEffort(t, 2.5): true, w.mustRegisterEffort(t, 1): true}

	for _, path := range []string{
		"/api/sprints/" + w.sprintUUID() + "/esfuerzos",
		"/api/items/" + w.itemUUID() + "/esfuerzos",
	} {
		rec := w.serve(http.MethodGet, path, "")
		expectStatus(t, rec, http.StatusOK)
		elements := listElements(t, rec, "esfuerzos")
		if len(elements) != 2 {
			t.Errorf("GET %s: %d esfuerzos, se esperaban 2", path, len(elements))
			continue
		}
		for _, e := range elements {
			if id := textOf(t, path+" id", e["id"]); !ids[id] {
				t.Errorf("GET %s: id = %q, no es el UUID de ninguno de los esfuerzos registrados", path, id)
			}
			if got := textOf(t, path+" item_id", e["item_id"]); got != w.itemUUID() {
				t.Errorf("GET %s: item_id = %q, se esperaba %q", path, got, w.itemUUID())
			}
			if got := textOf(t, path+" sprint_id", e["sprint_id"]); got != w.sprintUUID() {
				t.Errorf("GET %s: sprint_id = %q, se esperaba %q", path, got, w.sprintUUID())
			}
		}
	}
}

// effortRoute es una de las rutas de efforts.go, con el mensaje de error de
// formato y el de inexistencia de la entidad que tiene el ID mal escrito.
type effortRoute struct {
	method, path, body     string
	formatMsg, notFoundMsg string
}

// effortRoutesWith devuelve las 4 rutas de efforts.go con bad en el lugar de
// cada ID que reciben: el registro, una vez en {id} y otra en {itemId}.
func effortRoutesWith(w *defectEffortWorld, bad string) []effortRoute {
	const (
		itemMsg   = "el id del ítem debe ser un UUID"
		sprintMsg = "el id del sprint debe ser un UUID"
	)
	return []effortRoute{
		{http.MethodPost, effortsPathFor(bad, w.itemUUID()), uuidEffortBody, sprintMsg, sprint.ErrSprintNotFound.Error()},
		{http.MethodPost, effortsPathFor(w.sprintUUID(), bad), uuidEffortBody, itemMsg, backlog.ErrItemNotFound.Error()},
		{http.MethodGet, "/api/items/" + bad + "/esfuerzos", "", itemMsg, backlog.ErrItemNotFound.Error()},
		{http.MethodGet, "/api/sprints/" + bad + "/esfuerzos", "", sprintMsg, sprint.ErrSprintNotFound.Error()},
		{http.MethodPatch, "/api/esfuerzos/" + bad + "/horas", `{"horas":3}`, "el id del esfuerzo debe ser un UUID", effort.ErrEffortNotFound.Error()},
	}
}

// HU-52 · CA-52.4 · Verifica que las 4 rutas de esfuerzos respondan 400 con el mensaje de su entidad cuando el ID es un número, está entre llaves, tiene urn:uuid: adelante o no tiene guiones.
func TestRutasDeEsfuerzos_IDQueNoEsUUIDDevuelve400(t *testing.T) {
	w := newDefectEffortWorld(t)
	id := w.mustRegisterEffort(t, 2.5)

	for _, bad := range []string{"1", "{" + id + "}", "urn:uuid:" + id, strings.ReplaceAll(id, "-", "")} {
		for _, r := range effortRoutesWith(w, bad) {
			expectErrorBody(t, r.method+" "+r.path, w.serve(r.method, r.path, r.body), http.StatusBadRequest, r.formatMsg)
		}
	}
}

// HU-52 · CA-52.4 · Verifica que las 4 rutas de esfuerzos respondan 404 con el mensaje de su entidad cuando el UUID no existe o es el nulo.
func TestRutasDeEsfuerzos_UUIDInexistenteDevuelve404(t *testing.T) {
	w := newDefectEffortWorld(t)
	w.mustRegisterEffort(t, 2.5)

	for _, missing := range []string{neverAssignedUUID, nilUUID} {
		for _, r := range effortRoutesWith(w, missing) {
			expectErrorBody(t, r.method+" "+r.path, w.serve(r.method, r.path, r.body), http.StatusNotFound, r.notFoundMsg)
		}
	}
}

// HU-52 · CA-52.4 · Verifica que en el registro el {id} del sprint se valide antes que el {itemId}: con los dos mal escritos responde el 400 del sprint.
func TestRegistrarEsfuerzo_IDDelSprintSeValidaAntesQueElDelItem(t *testing.T) {
	w := newDefectEffortWorld(t)

	rec := w.serve(http.MethodPost, effortsPathFor("abc", "abc"), uuidEffortBody)

	expectErrorBody(t, "POST con los dos IDs mal escritos", rec, http.StatusBadRequest, "el id del sprint debe ser un UUID")
}

// HU-52 · CA-52.4 · Verifica que el {id} del esfuerzo se valide antes que el cuerpo: un ID que no es UUID con un cuerpo que no es JSON da el 400 del ID.
func TestCorregirEsfuerzo_IDSeValidaAntesQueElCuerpo(t *testing.T) {
	w := newDefectEffortWorld(t)

	rec := w.serve(http.MethodPatch, "/api/esfuerzos/abc/horas", "esto no es JSON")

	expectErrorBody(t, "PATCH /api/esfuerzos/abc/horas", rec, http.StatusBadRequest, "el id del esfuerzo debe ser un UUID")
}

// HU-52 · CA-52.4 · Verifica que con un UUID de esfuerzo inexistente se conserve el orden de errores de HU-28: horas inválidas dan 400, y no 404.
func TestCorregirEsfuerzo_UUIDInexistenteConservaElOrdenDeErrores(t *testing.T) {
	w := newDefectEffortWorld(t)

	rec := w.serve(http.MethodPatch, "/api/esfuerzos/"+neverAssignedUUID+"/horas", `{"horas":0}`)

	expectErrorBody(t, "PATCH con UUID inexistente", rec, http.StatusBadRequest, effort.ErrInvalidHours.Error())
}
