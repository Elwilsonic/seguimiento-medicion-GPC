package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// CA-21.7 · POST /api/items/{id}/poker-sessions expone la apertura de una
// sesión de Planning Poker. Por endpoint hay un caso exitoso y un test por
// cada código de error.

type pokerSessionJSON struct {
	ID     int             `json:"id"`
	ItemID int             `json:"item_id"`
	Estado string          `json:"estado"`
	Ronda  int             `json:"ronda"`
	Votos  []pokerVoteJSON `json:"votos"`
	Cartas []int           `json:"cartas"`
}

type pokerVoteJSON struct {
	Integrante string `json:"integrante"`
	Puntos     int    `json:"puntos"`
}

func TestAbrirSesionPoker_Devuelve201ConLaSesionAbierta(t *testing.T) {
	h := newHandlerWithItem(t)

	rec := do(t, h, http.MethodPost, "/api/items/1/poker-sessions", "")

	expectStatus(t, rec, http.StatusCreated)
	var got pokerSessionJSON
	decode(t, rec, &got)
	if got.ID != 1 || got.ItemID != 1 || got.Estado != "ABIERTA" || got.Ronda != 1 {
		t.Errorf("sesión = %+v, se esperaba ID 1, ítem 1, ABIERTA y ronda 1", got)
	}
	if got.Votos == nil {
		t.Error("votos = null, se esperaba una lista vacía")
	}
	if len(got.Votos) != 0 {
		t.Errorf("votos = %+v, se esperaba una sesión sin votos", got.Votos)
	}
	wantCards := []int{1, 2, 3, 5, 8, 13, 21}
	if len(got.Cartas) != len(wantCards) {
		t.Fatalf("cartas = %v, se esperaba %v", got.Cartas, wantCards)
	}
	for i := range wantCards {
		if got.Cartas[i] != wantCards[i] {
			t.Fatalf("cartas = %v, se esperaba %v", got.Cartas, wantCards)
		}
	}
}

func TestAbrirSesionPoker_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandlerWithItem(t), http.MethodPost, "/api/items/abc/poker-sessions", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

func TestAbrirSesionPoker_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPost, "/api/items/999/poker-sessions", "")

	expectError(t, rec, http.StatusNotFound, backlog.ErrItemNotFound.Error())
}

func TestAbrirSesionPoker_ItemTerminadoDevuelve409(t *testing.T) {
	h := newHandlerWithItem(t)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/items/1/estado", `{"estado":"EN_CURSO"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/items/1/estado", `{"estado":"TERMINADO"}`), http.StatusOK)

	rec := do(t, h, http.MethodPost, "/api/items/1/poker-sessions", "")

	expectError(t, rec, http.StatusConflict, poker.ErrItemFinished.Error())
}

func TestAbrirSesionPoker_SesionAbiertaDevuelve409(t *testing.T) {
	h := newHandlerWithItem(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/items/1/poker-sessions", ""), http.StatusCreated)

	rec := do(t, h, http.MethodPost, "/api/items/1/poker-sessions", "")

	expectError(t, rec, http.StatusConflict, poker.ErrSessionAlreadyOpen.Error())
}

func TestAbrirSesionPoker_MetodoNoPermitidoDevuelve405(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandlerWithItem(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/items/1/poker-sessions", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// brokenPokerRepository es un repositorio de sesiones que siempre falla.
type brokenPokerRepository struct{}

func (brokenPokerRepository) Save(poker.Session) (poker.Session, error) {
	return poker.Session{}, errStorage
}
func (brokenPokerRepository) OpenByItem(int) (poker.Session, bool, error) {
	return poker.Session{}, false, errStorage
}

func TestAbrirSesionPoker_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	h := httpapi.NewHandler(projects, items, sprints, efforts, poker.NewManager(brokenPokerRepository{}, items))
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem), http.StatusCreated)

	rec := do(t, h, http.MethodPost, "/api/items/1/poker-sessions", "")

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
