package httpapi

import (
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"uuid"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/member"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197). Test
// interno del paquete, porque los helpers de la base no se exportan.

// entityIDErrors son los errores de formato de las 6 entidades, con su
// mensaje (CA-52.4).
var entityIDErrors = []struct {
	name string
	err  error
	msg  string
}{
	{"ítem", errInvalidItemID, "el id del ítem debe ser un UUID"},
	{"sprint", errInvalidSprintID, "el id del sprint debe ser un UUID"},
	{"integrante", errInvalidMemberID, "el id del integrante debe ser un UUID"},
	{"defecto", errInvalidDefectID, "el id del defecto debe ser un UUID"},
	{"esfuerzo", errInvalidEffortID, "el id del esfuerzo debe ser un UUID"},
	{"sesión", errInvalidSessionID, "el id de la sesión debe ser un UUID"},
}

// HU-52 · CA-52.4 · Verifica que los 6 errores de formato tengan el mensaje de su entidad y respondan 400.
func TestStatusFor_IDsDeEntidadInvalidosSon400(t *testing.T) {
	for _, e := range entityIDErrors {
		if got := e.err.Error(); got != e.msg {
			t.Errorf("error de %s = %q, se esperaba %q", e.name, got, e.msg)
		}
		if got := statusFor(e.err); got != http.StatusBadRequest {
			t.Errorf("statusFor(error de %s) = %d, se esperaba 400", e.name, got)
		}
	}
}

// HU-52 · CA-52.3 · Verifica que parsePublicID acepte un UUID en forma canónica, también en mayúsculas, y lo devuelva en minúsculas.
func TestParsePublicID_AceptaLaFormaCanonicaYMayusculas(t *testing.T) {
	for _, s := range []string{canonicalProjectID, strings.ToUpper(canonicalProjectID)} {
		got, err := parsePublicID(s, errInvalidItemID)

		if err != nil {
			t.Fatalf("parsePublicID(%q) devolvió un error inesperado: %v", s, err)
		}
		if got.String() != canonicalProjectID {
			t.Errorf("parsePublicID(%q) = %s, se esperaba %s", s, got, canonicalProjectID)
		}
	}
}

// HU-52 · CA-52.4 · Verifica que parsePublicID rechace con el error de la entidad todo lo que no es un UUID en forma canónica: números, llaves, urn:uuid:, sin guiones, vacío y dígitos no hexadecimales.
func TestParsePublicID_RechazaConElErrorDeLaEntidad(t *testing.T) {
	cases := []string{
		"1", "0", "-1", "abc", "",
		"{" + canonicalProjectID + "}",
		"urn:uuid:" + canonicalProjectID,
		strings.ReplaceAll(canonicalProjectID, "-", ""),
		"3f2b8c1d-5e6a-4b7c-8d9e-0f1a2b3c4d5g",
		canonicalProjectID[:35],
		" " + canonicalProjectID + " ",
	}
	for _, e := range entityIDErrors {
		for _, s := range cases {
			if _, err := parsePublicID(s, e.err); !errors.Is(err, e.err) {
				t.Errorf("parsePublicID(%q) de %s: err = %v, se esperaba %v", s, e.name, err, e.err)
			}
		}
	}
}

// HU-52 · CA-52.4 · Verifica que missingInternalID sea el máximo de INTEGER en PostgreSQL, que ningún repositorio asigna y que no es el "sprint faltante" (0).
func TestMissingInternalID_EsElMaximoDeInteger(t *testing.T) {
	if missingInternalID != math.MaxInt32 {
		t.Errorf("missingInternalID = %d, se esperaba math.MaxInt32 (%d)", missingInternalID, math.MaxInt32)
	}
}

// publicIDServer es un server interno con los Manager en memoria y un ítem,
// un sprint, un integrante, un defecto, un esfuerzo y una sesión creados.
type publicIDServer struct {
	*server
	item    backlog.Item
	sprint  sprint.Sprint
	member  member.Member
	defect  defect.Defect
	effort  effort.Effort
	session poker.Session
}

// itemCounter cuenta las lecturas por ID del repositorio de ítems, para
// verificar la cache del resolvedor.
type itemCounter struct {
	*backlog.MemoryRepository
	gets int
	err  error
}

func (r *itemCounter) GetByID(id int) (backlog.Item, error) {
	r.gets++
	if r.err != nil {
		return backlog.Item{}, r.err
	}
	return r.MemoryRepository.GetByID(id)
}

func newPublicIDServer(t *testing.T) (*publicIDServer, *itemCounter) {
	t.Helper()
	projects := project.NewManager(project.NewMemoryRepository())
	p, err := projects.Create("Sistema de ventas", "")
	if err != nil {
		t.Fatalf("crear el proyecto: %v", err)
	}
	itemRepo := &itemCounter{MemoryRepository: backlog.NewMemoryRepository()}
	items := backlog.NewManager(itemRepo, projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	members := member.NewManager(member.NewMemoryRepository(), projects)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokers := poker.NewManager(poker.NewMemoryRepository(), items)
	env := &publicIDServer{server: &server{projects: projects, items: items, sprints: sprints, members: members, defects: defects, efforts: efforts, poker: pokers}}

	must := func(what string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", what, err)
		}
	}
	env.item, err = items.Create(p.ID, "Login de usuarios", "", "ALTA", []string{"Login OK"})
	must("crear el ítem", err)
	env.sprint, err = sprints.Create(p.ID, "Sprint 1", time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), "Entregar el MVP")
	must("crear el sprint", err)
	_, err = sprints.AssignItem(env.sprint.ID, env.item.ID)
	must("asignar el ítem", err)
	_, err = items.ChangeStatus(env.item.ID, "EN_CURSO")
	must("pasar el ítem a EN_CURSO", err)
	env.member, err = members.Register(p.ID, "Ana Pérez", "PRODUCT_BUILDER")
	must("registrar el integrante", err)
	env.defect, err = defects.Register(env.item.ID, "El total no descuenta", "ALTA", env.sprint.ID)
	must("registrar el defecto", err)
	env.effort, err = efforts.Register(env.sprint.ID, env.item.ID, "Ana Pérez", time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC), "Implementar", 2.5)
	must("registrar el esfuerzo", err)
	env.session, err = pokers.OpenSession(env.item.ID)
	must("abrir la sesión", err)
	return env, itemRepo
}

// pathRequest arma un pedido con el parámetro de ruta name = value.
func pathRequest(name, value string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetPathValue(name, value)
	return r
}

// fromPathCase es un xIDFromPath con el UUID y el ID interno de su entidad.
type fromPathCase struct {
	name     string
	param    string
	read     func(*http.Request) (int, error)
	publicID uuid.UUID
	id       int
	err      error
}

func fromPathCases(env *publicIDServer) []fromPathCase {
	return []fromPathCase{
		{"itemIDFromPath {id}", "id", func(r *http.Request) (int, error) { return env.itemIDFromPath(r, "id") }, env.item.PublicID, env.item.ID, errInvalidItemID},
		{"itemIDFromPath {itemId}", "itemId", func(r *http.Request) (int, error) { return env.itemIDFromPath(r, "itemId") }, env.item.PublicID, env.item.ID, errInvalidItemID},
		{"sprintIDFromPath", "id", env.sprintIDFromPath, env.sprint.PublicID, env.sprint.ID, errInvalidSprintID},
		{"memberIDFromPath", "id", env.memberIDFromPath, env.member.PublicID, env.member.ID, errInvalidMemberID},
		{"defectIDFromPath", "id", env.defectIDFromPath, env.defect.PublicID, env.defect.ID, errInvalidDefectID},
		{"effortIDFromPath", "id", env.effortIDFromPath, env.effort.PublicID, env.effort.ID, errInvalidEffortID},
		{"sessionIDFromPath", "id", env.sessionIDFromPath, env.session.PublicID, env.session.ID, errInvalidSessionID},
	}
}

// HU-52 · CA-52.3 · Verifica que cada xIDFromPath traduzca el UUID de su entidad, también en mayúsculas, al ID interno.
func TestIDFromPath_UUIDExistenteDaElIDInterno(t *testing.T) {
	env, _ := newPublicIDServer(t)
	for _, c := range fromPathCases(env) {
		for _, s := range []string{c.publicID.String(), strings.ToUpper(c.publicID.String())} {
			got, err := c.read(pathRequest(c.param, s))

			if err != nil {
				t.Errorf("%s(%q) devolvió un error inesperado: %v", c.name, s, err)
				continue
			}
			if got != c.id {
				t.Errorf("%s(%q) = %d, se esperaba %d", c.name, s, got, c.id)
			}
		}
	}
}

// HU-52 · CA-52.4 · Verifica que cada xIDFromPath traduzca un UUID que no existe a missingInternalID, sin error, para que el Manager informe el 404 en su orden de validaciones.
func TestIDFromPath_UUIDInexistenteDaMissingInternalID(t *testing.T) {
	env, _ := newPublicIDServer(t)
	for _, c := range fromPathCases(env) {
		got, err := c.read(pathRequest(c.param, uuid.NewV4().String()))

		if err != nil {
			t.Errorf("%s devolvió un error inesperado: %v", c.name, err)
		}
		if got != missingInternalID {
			t.Errorf("%s = %d, se esperaba missingInternalID (%d)", c.name, got, missingInternalID)
		}
	}
}

// HU-52 · CA-52.4 · Verifica que cada xIDFromPath rechace un ID que no es un UUID con el error de su entidad, también un número entero.
func TestIDFromPath_FormatoInvalidoDaElErrorDeLaEntidad(t *testing.T) {
	env, _ := newPublicIDServer(t)
	for _, c := range fromPathCases(env) {
		for _, s := range []string{"1", "abc", ""} {
			got, err := c.read(pathRequest(c.param, s))

			if !errors.Is(err, c.err) {
				t.Errorf("%s(%q): err = %v, se esperaba %v", c.name, s, err, c.err)
			}
			if got != 0 {
				t.Errorf("%s(%q) = %d, se esperaba 0", c.name, s, got)
			}
		}
	}
}

// HU-52 · CA-52.3 · Verifica que sprintIDFromBody traduzca el UUID de un sprint del cuerpo a su ID interno, el texto vacío a 0 ("sprint faltante") y un UUID inexistente a missingInternalID.
func TestSprintIDFromBody_TraduceElUUIDDelCuerpo(t *testing.T) {
	env, _ := newPublicIDServer(t)
	cases := []struct {
		name string
		text string
		want int
	}{
		{"UUID existente", env.sprint.PublicID.String(), env.sprint.ID},
		{"en mayúsculas", strings.ToUpper(env.sprint.PublicID.String()), env.sprint.ID},
		{"vacío", "", 0},
		{"inexistente", uuid.NewV4().String(), missingInternalID},
	}
	for _, c := range cases {
		got, err := env.sprintIDFromBody(c.text)

		if err != nil {
			t.Errorf("%s: sprintIDFromBody(%q) devolvió un error inesperado: %v", c.name, c.text, err)
		}
		if got != c.want {
			t.Errorf("%s: sprintIDFromBody(%q) = %d, se esperaba %d", c.name, c.text, got, c.want)
		}
	}
}

// HU-52 · CA-52.4 · Verifica que sprintIDFromBody rechace un texto que no es un UUID, también un número, con el error del sprint.
func TestSprintIDFromBody_FormatoInvalidoDaErrInvalidSprintID(t *testing.T) {
	env, _ := newPublicIDServer(t)
	for _, s := range []string{"1", "abc", " "} {
		got, err := env.sprintIDFromBody(s)

		if !errors.Is(err, errInvalidSprintID) {
			t.Errorf("sprintIDFromBody(%q): err = %v, se esperaba %v", s, err, errInvalidSprintID)
		}
		if got != 0 {
			t.Errorf("sprintIDFromBody(%q) = %d, se esperaba 0", s, got)
		}
	}
}

// HU-52 · CA-52.3 · Verifica que el resolvedor devuelva el UUID del ítem y del sprint por su ID interno, y "" para el 0 (D7).
func TestPublicIDs_DevuelveLosUUIDYVacioParaElCero(t *testing.T) {
	env, _ := newPublicIDServer(t)
	ids := env.newPublicIDs()

	if got := ids.item(env.item.ID); got != env.item.PublicID.String() {
		t.Errorf("item(%d) = %q, se esperaba %q", env.item.ID, got, env.item.PublicID)
	}
	if got := ids.sprint(env.sprint.ID); got != env.sprint.PublicID.String() {
		t.Errorf("sprint(%d) = %q, se esperaba %q", env.sprint.ID, got, env.sprint.PublicID)
	}
	if got := ids.item(0); got != "" {
		t.Errorf("item(0) = %q, se esperaba \"\"", got)
	}
	if got := ids.sprint(0); got != "" {
		t.Errorf("sprint(0) = %q, se esperaba \"\"", got)
	}
	if ids.err != nil {
		t.Errorf("el resolvedor guardó un error inesperado: %v", ids.err)
	}
}

// HU-52 · CA-52.3 · Verifica que el resolvedor busque cada ID una sola vez, aunque se lo pida varias veces.
func TestPublicIDs_BuscaCadaIDUnaSolaVez(t *testing.T) {
	env, counter := newPublicIDServer(t)
	ids := env.newPublicIDs()
	counter.gets = 0

	for range 3 {
		if got := ids.item(env.item.ID); got != env.item.PublicID.String() {
			t.Fatalf("item(%d) = %q, se esperaba %q", env.item.ID, got, env.item.PublicID)
		}
	}

	if counter.gets != 1 {
		t.Errorf("se leyó el ítem %d veces, se esperaba 1", counter.gets)
	}
}

// HU-52 · CA-52.3 · Verifica que, si el resolvedor no puede buscar un UUID, writeResolved responda 500 con el mensaje genérico, sin el detalle ni la respuesta.
func TestWriteResolved_ErrorDelResolvedorEs500(t *testing.T) {
	env, counter := newPublicIDServer(t)
	counter.err = errors.New("detalle interno del almacenamiento")
	ids := env.newPublicIDs()
	_ = ids.item(env.item.ID)
	rec := httptest.NewRecorder()

	env.writeResolved(rec, http.StatusOK, map[string]string{"id": "respuesta"}, ids)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, se esperaba 500", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, errInternal.Error()) || strings.Contains(body, "detalle interno") || strings.Contains(body, "respuesta") {
		t.Errorf("cuerpo = %q, se esperaba solo el error genérico", body)
	}
}

// HU-52 · CA-52.3 · Verifica que, sin errores en el resolvedor, writeResolved responda con el código y la respuesta en JSON.
func TestWriteResolved_SinErrorRespondeLaRespuesta(t *testing.T) {
	env, _ := newPublicIDServer(t)
	ids := env.newPublicIDs()
	_ = ids.item(env.item.ID)
	rec := httptest.NewRecorder()

	env.writeResolved(rec, http.StatusCreated, map[string]string{"id": ids.item(env.item.ID)}, ids)

	if rec.Code != http.StatusCreated {
		t.Errorf("código = %d, se esperaba 201", rec.Code)
	}
	if want := `{"id":"` + env.item.PublicID.String() + `"}`; strings.TrimSpace(rec.Body.String()) != want {
		t.Errorf("cuerpo = %q, se esperaba %q", rec.Body.String(), want)
	}
}
