package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/member"
)

// HU-52 · Identificadores UUID del resto de las entidades, parte de
// integrantes (Issue #197). Prueba el registro (HU-15), la lista (HU-17) y
// PUT /api/integrantes/{id} (HU-16), la única ruta que recibe un integrante.
// El id se decodifica con memberUUIDJSON y no con idRef, porque estos tests
// verifican que sea un UUID.

// missingMemberID es un UUID versión 4 que nunca se asigna a un integrante,
// y nilMemberID es el UUID nulo (CA-52.4).
const (
	missingMemberID = "6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f"
	nilMemberID     = "00000000-0000-0000-0000-000000000000"
)

// memberUUIDJSON es la respuesta de un integrante con el id como texto.
type memberUUIDJSON struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Rol    string `json:"rol"`
}

// jsonMemberID es el "id" del n-ésimo integrante registrado por la API en h,
// como texto JSON, para armar las respuestas esperadas (HU-52, CA-52.3).
func jsonMemberID(t *testing.T, h http.Handler, n int) string {
	t.Helper()
	entityIDsMu.Lock()
	ids := entityIDs[h][kindMember]
	entityIDsMu.Unlock()
	if n < 1 || n > len(ids) {
		t.Fatalf("jsonMemberID: el handler tiene %d integrantes registrados por la API, no existe el %d", len(ids), n)
	}
	return `"` + ids[n-1] + `"`
}

// expectOnlyOneMember verifica que el primer proyecto de h tenga un solo
// integrante. Reemplaza lo que antes probaba el "id":1 del primer registro
// válido (que no se había guardado nada antes), porque con UUID el id ya no
// cuenta los registros (HU-52).
func expectOnlyOneMember(t *testing.T, h http.Handler) {
	t.Helper()
	rec := do(t, h, http.MethodGet, projectPath(t, h, 1)+"/integrantes", "")
	expectStatus(t, rec, http.StatusOK)
	var list []json.RawMessage
	decode(t, rec, &list)
	if len(list) != 1 {
		t.Errorf("el proyecto tiene %d integrantes, se esperaba solo el del último registro: %s", len(list), rec.Body.String())
	}
}

// decodeMemberUUID lee la respuesta de un integrante, con el id como texto:
// un número es un error, porque HU-52 no expone el ID entero.
func decodeMemberUUID(t *testing.T, rec *httptest.ResponseRecorder) memberUUIDJSON {
	t.Helper()
	var m memberUUIDJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("el integrante no tiene el id como texto (%v): %s", err, rec.Body.String())
	}
	return m
}

// registerMemberUUID registra un integrante en el primer proyecto de h y
// devuelve su id, que tiene que ser un UUID versión 4.
func registerMemberUUID(t *testing.T, h http.Handler, name string) string {
	t.Helper()
	rec := do(t, h, http.MethodPost, projectPath(t, h, 1)+"/integrantes", `{"nombre":"`+name+`","rol":"PRODUCT_BUILDER"}`)
	expectStatus(t, rec, http.StatusCreated)
	id := decodeMemberUUID(t, rec).ID
	if !uuidV4Canonical.MatchString(id) {
		t.Fatalf("id del integrante = %q, se esperaba un UUID versión 4", id)
	}
	return id
}

// HU-52 · CA-52.1 · Verifica que POST /api/proyectos/{id}/integrantes responda 201 con el id del integrante como texto: un UUID versión 4 en minúsculas.
func TestRegistrarIntegrante_IDEsUUIDVersion4(t *testing.T) {
	h := newHandlerWithProject(t)

	rec := do(t, h, http.MethodPost, projectPath(t, h, 1)+"/integrantes", validMember)

	expectStatus(t, rec, http.StatusCreated)
	if id := decodeMemberUUID(t, rec).ID; !uuidV4Canonical.MatchString(id) {
		t.Errorf("id = %q, se esperaba un UUID versión 4 en minúsculas", id)
	}
}

// HU-52 · CA-52.2 · Verifica que dos integrantes registrados por la API tengan ids distintos y que ninguno sea un número.
func TestRegistrarIntegrante_DosIntegrantesTienenUUIDDistintos(t *testing.T) {
	h := newHandlerWithProject(t)

	first := registerMemberUUID(t, h, "Ana Pérez")
	second := registerMemberUUID(t, h, "Bruno")

	if first == second {
		t.Errorf("los dos integrantes tienen el id %q", first)
	}
}

// HU-52 · CA-52.3 · Verifica que PUT /api/integrantes/{id} acepte el UUID del integrante y responda 200 con los datos nuevos y el mismo UUID en id.
func TestModificarIntegrante_AceptaElUUID(t *testing.T) {
	h := newHandlerWithProject(t)
	id := registerMemberUUID(t, h, "Ana Pérez")

	rec := serveRaw(h, http.MethodPut, "/api/integrantes/"+id, `{"nombre":"Ana María","rol":"AGILE_ENABLER"}`)

	expectStatus(t, rec, http.StatusOK)
	if got := decodeMemberUUID(t, rec); got != (memberUUIDJSON{ID: id, Nombre: "Ana María", Rol: "AGILE_ENABLER"}) {
		t.Errorf("integrante = %+v, se esperaba Ana María, AGILE_ENABLER, con el id %q", got, id)
	}
}

// HU-52 · CA-52.3 · Verifica que PUT /api/integrantes/{id} acepte el UUID en mayúsculas y responda con el id en minúsculas.
func TestModificarIntegrante_AceptaElUUIDEnMayusculas(t *testing.T) {
	h := newHandlerWithProject(t)
	id := registerMemberUUID(t, h, "Ana Pérez")

	rec := serveRaw(h, http.MethodPut, "/api/integrantes/"+strings.ToUpper(id), `{"nombre":"Ana Pérez","rol":"PRODUCT_ARCHITECT"}`)

	expectStatus(t, rec, http.StatusOK)
	if got := decodeMemberUUID(t, rec).ID; got != id {
		t.Errorf("id = %q, se esperaba %q en minúsculas", got, id)
	}
}

// HU-52 · CA-52.4 · Verifica que en PUT /api/integrantes/{id} un id numérico (también 0 y negativo) responda 400 con "el id del integrante debe ser un UUID", aunque exista el integrante 1.
func TestModificarIntegrante_IDNumericoDevuelve400(t *testing.T) {
	h := newHandlerWithMembers(t)

	for _, id := range []string{"1", "0", "-1"} {
		path := "/api/integrantes/" + id
		expectErrorBody(t, "PUT "+path, serveRaw(h, http.MethodPut, path, validMember), http.StatusBadRequest, "el id del integrante debe ser un UUID")
	}
}

// HU-52 · CA-52.4 · Verifica que el UUID de un integrante existente escrito entre llaves, con urn:uuid: o sin guiones responda 400.
func TestModificarIntegrante_OtrasFormasDelUUIDDevuelven400(t *testing.T) {
	h := newHandlerWithProject(t)
	id := registerMemberUUID(t, h, "Ana Pérez")

	for _, other := range []string{"{" + id + "}", "urn:uuid:" + id, strings.ReplaceAll(id, "-", "")} {
		path := "/api/integrantes/" + other
		expectErrorBody(t, "PUT "+path, serveRaw(h, http.MethodPut, path, validMember), http.StatusBadRequest, "el id del integrante debe ser un UUID")
	}
}

// HU-52 · CA-52.4 · Verifica que un UUID que nunca se asignó a un integrante, o el UUID nulo, responda 404 con ErrMemberNotFound.
func TestModificarIntegrante_UUIDInexistenteDevuelve404(t *testing.T) {
	h := newHandlerWithMembers(t)

	for _, id := range []string{missingMemberID, nilMemberID} {
		path := "/api/integrantes/" + id
		expectErrorBody(t, "PUT "+path, serveRaw(h, http.MethodPut, path, validMember), http.StatusNotFound, member.ErrMemberNotFound.Error())
	}
}

// HU-52 · CA-52.4 · Verifica que el id del integrante se valide antes que el cuerpo: un id inválido con un cuerpo que no es JSON responde el error del id.
func TestOrdenDeErroresIntegrante_IDInvalidoAntesQueElCuerpo(t *testing.T) {
	path := "/api/integrantes/abc"

	rec := serveRaw(newHandlerWithMembers(t), http.MethodPut, path, `no es json`)

	expectErrorBody(t, "PUT "+path, rec, http.StatusBadRequest, "el id del integrante debe ser un UUID")
}

// HU-52 · CA-52.4 · Verifica que, con un UUID de integrante inexistente, el nombre vacío se informe antes que la inexistencia, como con un ID entero (orden de errores de HU-16).
func TestOrdenDeErroresIntegrante_UUIDInexistenteConservaElOrden(t *testing.T) {
	path := "/api/integrantes/" + missingMemberID

	rec := serveRaw(newHandlerWithMembers(t), http.MethodPut, path, `{"nombre":"  ","rol":"AGILE_ENABLER"}`)

	expectErrorBody(t, "PUT "+path, rec, http.StatusBadRequest, member.ErrEmptyName.Error())
}

// HU-52 · CA-52.3 · Verifica que GET /api/proyectos/{id}/integrantes identifique cada integrante con su UUID, como texto, y en el mismo orden en que se registraron.
func TestListarIntegrantes_IdentificaCadaIntegranteConSuUUID(t *testing.T) {
	h := newHandlerWithProject(t)
	first := registerMemberUUID(t, h, "Ana Pérez")
	second := registerMemberUUID(t, h, "Bruno")

	rec := do(t, h, http.MethodGet, projectPath(t, h, 1)+"/integrantes", "")

	expectStatus(t, rec, http.StatusOK)
	var list []memberUUIDJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("la lista no tiene los id como texto (%v): %s", err, rec.Body.String())
	}
	if len(list) != 2 || list[0].ID != first || list[1].ID != second {
		t.Errorf("integrantes = %+v, se esperaban %q y %q", list, first, second)
	}
}
