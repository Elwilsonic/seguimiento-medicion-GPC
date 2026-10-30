package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

// HU-16 · Modificar los datos o el rol de un integrante (Issue #30)
//
// CA-16.7: el integrante se modifica en PUT /api/integrantes/{id} y con el
// formulario "Modificar integrante". El 400 de ID y los de cuerpo (413 y
// texto después del JSON) están en las listas comunes de server_test.go.

// GetByID y Update completan member.Repository en brokenMemberRepository
// (members_test.go), que falla siempre, para probar el 500.
func (brokenMemberRepository) GetByID(int) (member.Member, error) {
	return member.Member{}, errStorage
}

func (brokenMemberRepository) Update(member.Member) (member.Member, error) {
	return member.Member{}, errStorage
}

// expectNoMemberSaved es el untouched de la fila de bodyEndpoints: el entorno
// no tiene integrantes y el primer registro válido recibe el ID 1, así que el
// pedido rechazado no guardó nada.
func expectNoMemberSaved(t *testing.T, h http.Handler) {
	t.Helper()
	rec := do(t, h, http.MethodPost, "/api/proyectos/1/integrantes", validMember)
	expectStatus(t, rec, http.StatusCreated)
	expectMemberJSON(t, rec, `{"id":1,"proyecto_id":1,"nombre":"Ana Pérez","rol":"PRODUCT_BUILDER"}`)
}

// newHandlerWithMembers arma la API con el proyecto 1 y los integrantes 1
// "Ana Pérez" (PRODUCT_BUILDER) y 2 "Bruno" (AGILE_ENABLER).
func newHandlerWithMembers(t *testing.T) http.Handler {
	t.Helper()
	h := newHandlerWithProject(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/integrantes", validMember), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/integrantes", `{"nombre":"Bruno","rol":"AGILE_ENABLER"}`), http.StatusCreated)
	return h
}

// CA-16.1 y CA-16.7 · Responde 200 con el integrante modificado.
func TestModificarIntegrante_Devuelve200ConElIntegrante(t *testing.T) {
	h := newHandlerWithMembers(t)

	rec := do(t, h, http.MethodPut, "/api/integrantes/1", `{"nombre":"  Ana María ","rol":"agile_enabler"}`)

	expectStatus(t, rec, http.StatusOK)
	expectMemberJSON(t, rec, `{"id":1,"proyecto_id":1,"nombre":"Ana María","rol":"AGILE_ENABLER"}`)
}

// CA-16.2 y CA-16.7 · Los datos inválidos devuelven 400 con el mensaje del
// error.
func TestModificarIntegrante_DatosInvalidosDevuelven400(t *testing.T) {
	casos := []struct {
		nombre, body string
		want         error
	}{
		{"nombre vacío", `{"nombre":"  ","rol":"PRODUCT_BUILDER"}`, member.ErrEmptyName},
		{"nombre ausente", `{"rol":"PRODUCT_BUILDER"}`, member.ErrEmptyName},
		{"cuerpo sin campos", `{}`, member.ErrEmptyName},
		{"rol vacío", `{"nombre":"Ana","rol":""}`, member.ErrMissingRole},
		{"rol ausente", `{"nombre":"Ana"}`, member.ErrMissingRole},
		{"rol inválido", `{"nombre":"Ana","rol":"SCRUM_MASTER"}`, member.ErrInvalidRole},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			h := newHandlerWithMembers(t)

			rec := do(t, h, http.MethodPut, "/api/integrantes/1", c.body)

			expectError(t, rec, http.StatusBadRequest, c.want.Error())
		})
	}
}

// CA-16.7 · Un cuerpo que no es JSON o un campo que no es texto devuelven 400.
func TestModificarIntegrante_CuerpoInvalidoDevuelve400(t *testing.T) {
	for _, body := range []string{`no es JSON`, `{"nombre":5,"rol":"PRODUCT_BUILDER"}`} {
		t.Run(body, func(t *testing.T) {
			h := newHandlerWithMembers(t)

			rec := do(t, h, http.MethodPut, "/api/integrantes/1", body)

			expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
		})
	}
}

// CA-16.3 y CA-16.7 · Un integrante inexistente, 0 o negativo devuelve 404.
func TestModificarIntegrante_InexistenteDevuelve404(t *testing.T) {
	for _, path := range []string{"/api/integrantes/999", "/api/integrantes/0", "/api/integrantes/-1"} {
		t.Run(path, func(t *testing.T) {
			h := newHandlerWithMembers(t)

			rec := do(t, h, http.MethodPut, path, validMember)

			expectError(t, rec, http.StatusNotFound, member.ErrMemberNotFound.Error())
		})
	}
}

// CA-16.4 y CA-16.7 · El nombre de otro integrante del proyecto devuelve 409.
func TestModificarIntegrante_NombreDeOtroIntegranteDevuelve409(t *testing.T) {
	h := newHandlerWithMembers(t)

	rec := do(t, h, http.MethodPut, "/api/integrantes/1", `{"nombre":"BRUNO","rol":"PRODUCT_BUILDER"}`)

	expectError(t, rec, http.StatusConflict, member.ErrDuplicateMember.Error())
}

// CA-16.4 · Cambiar solo las mayúsculas del propio nombre responde 200.
func TestModificarIntegrante_PropioNombreConOtrasMayusculasDevuelve200(t *testing.T) {
	h := newHandlerWithMembers(t)

	rec := do(t, h, http.MethodPut, "/api/integrantes/1", `{"nombre":"ANA PÉREZ","rol":"PRODUCT_BUILDER"}`)

	expectStatus(t, rec, http.StatusOK)
	expectMemberJSON(t, rec, `{"id":1,"proyecto_id":1,"nombre":"ANA PÉREZ","rol":"PRODUCT_BUILDER"}`)
}

// CA-16.7 · Los demás métodos devuelven 405. GET necesita su ruta porque, sin
// ella, lo atendería la interfaz web con un 404.
func TestModificarIntegrante_MetodoNoPermitidoDevuelve405(t *testing.T) {
	h := newHandlerWithMembers(t)

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(method, "/api/integrantes/1", nil))

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

// CA-16.7 · El 405 de GET anuncia PUT.
func TestModificarIntegrante_GETAnunciaPUTEnAllow(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandlerWithMembers(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/integrantes/1", nil))

	if got := rec.Header().Get("Allow"); got != http.MethodPut {
		t.Errorf("Allow = %q, se esperaba %q", got, http.MethodPut)
	}
}

// CA-16.5 y CA-16.7 · Una falla del almacenamiento devuelve 500 sin detalle.
func TestModificarIntegrante_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	members := member.NewManager(brokenMemberRepository{}, projects)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, members)

	rec := do(t, h, http.MethodPut, "/api/integrantes/1", validMember)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// RN-16.8 · Cambiar el nombre de un integrante no cambia sus esfuerzos: el
// integrante de un esfuerzo es un texto libre.
func TestModificarIntegrante_NoCambiaLosEsfuerzosRegistrados(t *testing.T) {
	h := newHandlerWithEffort(t)
	expectStatus(t, do(t, h, http.MethodPost, effortsPath, validEffort), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/integrantes", `{"nombre":"García Santiago","rol":"PRODUCT_BUILDER"}`), http.StatusCreated)

	expectStatus(t, do(t, h, http.MethodPut, "/api/integrantes/1", `{"nombre":"Santiago García","rol":"PRODUCT_BUILDER"}`), http.StatusOK)

	rec := do(t, h, http.MethodGet, "/api/items/1/esfuerzos", "")
	expectStatus(t, rec, http.StatusOK)
	var got effortListJSON
	decode(t, rec, &got)
	if len(got.Esfuerzos) != 1 || got.Esfuerzos[0].Integrante != "García Santiago" {
		t.Errorf("esfuerzos = %+v, se esperaba el nombre original", got.Esfuerzos)
	}
}

// CA-16.7 · La sección "Integrantes" tiene el formulario "Modificar
// integrante" con los tres roles.
func TestInterfaz_SeccionIntegrantesConElFormularioDeModificar(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<form id="form-modificar-integrante">`)
	if inicio == -1 {
		t.Fatal("la página no contiene el formulario de modificar integrante")
	}
	fin := strings.Index(page[inicio:], "</form>")
	if fin == -1 {
		t.Fatal("el formulario de modificar integrante no se cierra")
	}
	form := page[inicio : inicio+fin]
	for _, want := range []string{
		"<h3>Modificar integrante</h3>",
		`name="id"`,
		`name="nombre"`,
		`name="rol"`,
		`<option value="PRODUCT_ARCHITECT">Product Architect</option>`,
		`<option value="AGILE_ENABLER">Agile Enabler</option>`,
		`<option value="PRODUCT_BUILDER">Product Builder</option>`,
		"Guardar cambios",
	} {
		if !strings.Contains(form, want) {
			t.Errorf("el formulario no contiene %q", want)
		}
	}
	seccion := page[strings.Index(page, `<section id="integrantes">`):]
	if strings.Index(seccion, "</section>") < strings.Index(seccion, `id="form-modificar-integrante"`) {
		t.Error("el formulario no está dentro de la sección de integrantes")
	}
}

// CA-16.7 · El formulario llama a PUT /integrantes/{id} y muestra el resultado
// en msg-integrantes.
func TestInterfaz_ModificarIntegranteLlamaALaAPI(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`$("form-modificar-integrante")`,
		`api("PUT", "/integrantes/" + `,
		`" modificado: "`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}
