package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/member"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-15 · Registrar integrante del proyecto (Issue #29)
//
// CA-15.7: el integrante se registra en POST /api/proyectos/{id}/integrantes.

const validMember = `{"nombre":"Ana Pérez","rol":"PRODUCT_BUILDER"}`

// expectMemberJSON compara el cuerpo con el JSON esperado, los dos
// compactados. No usa expectRawJSON, que además exige el código 200: el
// registro responde 201.
func expectMemberJSON(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	var got, expected bytes.Buffer
	if err := json.Compact(&got, rec.Body.Bytes()); err != nil {
		t.Fatalf("cuerpo no es JSON válido (%v): %s", err, rec.Body.String())
	}
	if err := json.Compact(&expected, []byte(want)); err != nil {
		t.Fatalf("JSON esperado inválido: %v", err)
	}
	if !bytes.Equal(got.Bytes(), expected.Bytes()) {
		t.Errorf("cuerpo =\n%s\nse esperaba\n%s", got.String(), expected.String())
	}
}

// newHandlerWithProject arma la API con el proyecto 1 "Sistema de ventas".
func newHandlerWithProject(t *testing.T) http.Handler {
	t.Helper()
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	return h
}

// CA-15.1 y CA-15.7 · Responde 201 con el integrante guardado.
func TestRegistrarIntegrante_Devuelve201ConElIntegrante(t *testing.T) {
	h := newHandlerWithProject(t)

	rec := do(t, h, http.MethodPost, "/api/proyectos/1/integrantes", `{"nombre":"  Ana Pérez ","rol":"product_builder"}`)

	expectStatus(t, rec, http.StatusCreated)
	expectMemberJSON(t, rec, `{"id":1,"proyecto_id":1,"nombre":"Ana Pérez","rol":"PRODUCT_BUILDER"}`)
}

// CA-15.2 · Nombre vacío, rol vacío o rol inválido devuelven 400.
func TestRegistrarIntegrante_DatosInvalidosDevuelven400(t *testing.T) {
	cases := []struct{ name, body, msg string }{
		{"nombre vacío", `{"nombre":"  ","rol":"PRODUCT_BUILDER"}`, member.ErrEmptyName.Error()},
		{"rol vacío", `{"nombre":"Ana","rol":""}`, member.ErrMissingRole.Error()},
		{"rol ausente", `{"nombre":"Ana"}`, member.ErrMissingRole.Error()},
		{"rol inválido", `{"nombre":"Ana","rol":"SCRUM_MASTER"}`, member.ErrInvalidRole.Error()},
		{"JSON inválido", `{"nombre":5,"rol":"PRODUCT_BUILDER"}`, "el cuerpo del pedido no es un JSON válido"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHandlerWithProject(t)

			rec := do(t, h, http.MethodPost, "/api/proyectos/1/integrantes", c.body)

			expectError(t, rec, http.StatusBadRequest, c.msg)
		})
	}
}

// CA-15.4 · Un proyecto inexistente, 0 o negativo devuelve 404.
func TestRegistrarIntegrante_ProyectoInexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodPost, "/api/proyectos/"+id+"/integrantes", validMember)

		expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
	}
}

// CA-15.7 · Un {id} no numérico devuelve 400.
func TestRegistrarIntegrante_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/api/proyectos/abc/integrantes", validMember)

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-15.3 · El mismo nombre en el proyecto devuelve 409.
func TestRegistrarIntegrante_DuplicadoDevuelve409(t *testing.T) {
	h := newHandlerWithProject(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/integrantes", validMember), http.StatusCreated)

	rec := do(t, h, http.MethodPost, "/api/proyectos/1/integrantes", `{"nombre":"ana pérez","rol":"AGILE_ENABLER"}`)

	expectError(t, rec, http.StatusConflict, member.ErrDuplicateMember.Error())
}

// brokenMemberRepository falla siempre, para probar el 500.
type brokenMemberRepository struct{}

func (brokenMemberRepository) Save(member.Member) (member.Member, error) {
	return member.Member{}, errStorage
}
func (brokenMemberRepository) ListByProject(int) ([]member.Member, error) { return nil, errStorage }

// CA-15.5 y CA-15.7 · Una falla del almacenamiento devuelve 500 sin detalle.
func TestRegistrarIntegrante_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	members := member.NewManager(brokenMemberRepository{}, projects)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, members)
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPost, "/api/proyectos/1/integrantes", validMember)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
