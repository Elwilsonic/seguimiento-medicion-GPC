package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
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
func (brokenPokerRepository) GetByID(int) (poker.Session, error) {
	return poker.Session{}, errStorage
}
func (brokenPokerRepository) Update(poker.Session) (poker.Session, error) {
	return poker.Session{}, errStorage
}

func TestAbrirSesionPoker_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, poker.NewManager(brokenPokerRepository{}, items), defects)
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem), http.StatusCreated)

	rec := do(t, h, http.MethodPost, "/api/items/1/poker-sessions", "")

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

func TestVotarPoker_VotoNuevoYReemplazoDevuelven200SinVotos(t *testing.T) {
	h := newHandlerWithItem(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/items/1/poker-sessions", ""), http.StatusCreated)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", `{"integrante":"Ana","puntos":5}`)
	expectStatus(t, rec, http.StatusOK)
	var first pokerSessionJSON
	decode(t, rec, &first)
	if len(first.Votos) != 0 {
		t.Errorf("votos del primer voto = %+v, se esperaba ocultamiento", first.Votos)
	}

	rec = do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", `{"integrante":" ana ","puntos":8}`)
	expectStatus(t, rec, http.StatusOK)
	var replacement pokerSessionJSON
	decode(t, rec, &replacement)
	if len(replacement.Votos) != 0 {
		t.Errorf("votos del reemplazo = %+v, se esperaba ocultamiento", replacement.Votos)
	}
}

func TestVotarPoker_ErroresDevuelvenCodigoCorrespondiente(t *testing.T) {
	h := newHandlerWithItem(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/items/1/poker-sessions", ""), http.StatusCreated)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", `{"integrante":" ","puntos":5}`)
	expectError(t, rec, http.StatusBadRequest, poker.ErrEmptyVoter.Error())

	rec = do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", `{"integrante":"Ana","puntos":4}`)
	expectError(t, rec, http.StatusBadRequest, poker.ErrInvalidCard.Error())

	rec = do(t, h, http.MethodPost, "/api/poker-sessions/999/votes", `{"integrante":"Ana","puntos":5}`)
	expectError(t, rec, http.StatusNotFound, poker.ErrSessionNotFound.Error())
}

func TestVotarPoker_SesionNoAbiertaDevuelve409(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	pokerRepo := poker.NewMemoryRepository()
	pokerManager := poker.NewManager(pokerRepo, items)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	p, err := projects.Create("Sistema de ventas", "")
	if err != nil {
		t.Fatal(err)
	}
	item, err := items.Create(p.ID, "Historia", "Descripción", "MEDIA", []string{"Criterio"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := pokerManager.OpenSession(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	session.Status = poker.Status("REVELADA")
	if _, err := pokerRepo.Update(session); err != nil {
		t.Fatal(err)
	}
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", `{"integrante":"Ana","puntos":5}`)

	expectError(t, rec, http.StatusConflict, poker.ErrSessionNotOpen.Error())
}

type failingVoteHTTPRepository struct {
	*poker.MemoryRepository
	failUpdate bool
}

func (r *failingVoteHTTPRepository) Update(session poker.Session) (poker.Session, error) {
	if r.failUpdate {
		return poker.Session{}, errStorage
	}
	return r.MemoryRepository.Update(session)
}

func TestVotarPoker_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	pokerRepo := &failingVoteHTTPRepository{MemoryRepository: poker.NewMemoryRepository()}
	pokerManager := poker.NewManager(pokerRepo, items)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	p, err := projects.Create("Sistema de ventas", "")
	if err != nil {
		t.Fatal(err)
	}
	item, err := items.Create(p.ID, "Historia", "Descripción", "MEDIA", []string{"Criterio"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pokerManager.OpenSession(item.ID); err != nil {
		t.Fatal(err)
	}
	pokerRepo.failUpdate = true
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", `{"integrante":"Ana","puntos":5}`)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// HU-23 · Revelar las estimaciones (Issue #41)
//
// CA-23.8 · POST /api/poker-sessions/{id}/reveal expone la revelación. Un
// caso exitoso y un test por cada código de error.

// openSessionWithVotes abre la sesión 1 del ítem 1 y registra los votos por
// la API.
func openSessionWithVotes(t *testing.T, h http.Handler, votes ...string) {
	t.Helper()
	expectStatus(t, do(t, h, http.MethodPost, "/api/items/1/poker-sessions", ""), http.StatusCreated)
	for _, body := range votes {
		expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", body), http.StatusOK)
	}
}

func TestRevelarPoker_Devuelve200ConTodosLosVotos(t *testing.T) {
	h := newHandlerWithItem(t)
	openSessionWithVotes(t, h, `{"integrante":"Ana","puntos":5}`, `{"integrante":"Bruno","puntos":8}`)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", "")

	expectStatus(t, rec, http.StatusOK)
	var got pokerSessionJSON
	decode(t, rec, &got)
	if got.ID != 1 || got.ItemID != 1 || got.Estado != "REVELADA" || got.Ronda != 1 {
		t.Errorf("sesión = %+v, se esperaba la sesión 1 del ítem 1, REVELADA, en la ronda 1", got)
	}
	want := []pokerVoteJSON{{Integrante: "Ana", Puntos: 5}, {Integrante: "Bruno", Puntos: 8}}
	if !slices.Equal(got.Votos, want) {
		t.Errorf("votos = %+v, se esperaba %+v", got.Votos, want)
	}
	if !slices.Equal(got.Cartas, poker.Cards()) {
		t.Errorf("cartas = %v, se esperaba %v", got.Cartas, poker.Cards())
	}
}

func TestRevelarPoker_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandlerWithItem(t), http.MethodPost, "/api/poker-sessions/abc/reveal", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

func TestRevelarPoker_InexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandlerWithItem(t), http.MethodPost, "/api/poker-sessions/999/reveal", "")

	expectError(t, rec, http.StatusNotFound, poker.ErrSessionNotFound.Error())
}

func TestRevelarPoker_SesionNoAbiertaDevuelve409(t *testing.T) {
	h := newHandlerWithItem(t)
	openSessionWithVotes(t, h, `{"integrante":"Ana","puntos":5}`)
	expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", ""), http.StatusOK)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", "")

	expectError(t, rec, http.StatusConflict, poker.ErrSessionNotOpen.Error())
}

func TestRevelarPoker_SinVotosDevuelve409(t *testing.T) {
	h := newHandlerWithItem(t)
	openSessionWithVotes(t, h)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", "")

	expectError(t, rec, http.StatusConflict, poker.ErrNoVotes.Error())
}

func TestRevelarPoker_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	pokerRepo := &failingVoteHTTPRepository{MemoryRepository: poker.NewMemoryRepository()}
	pokerManager := poker.NewManager(pokerRepo, items)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	p, err := projects.Create("Sistema de ventas", "")
	if err != nil {
		t.Fatal(err)
	}
	item, err := items.Create(p.ID, "Historia", "Descripción", "MEDIA", []string{"Criterio"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := pokerManager.OpenSession(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pokerManager.Vote(session.ID, "Ana", 5); err != nil {
		t.Fatal(err)
	}
	pokerRepo.failUpdate = true
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", "")

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
	stored, err := pokerRepo.GetByID(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != poker.StatusOpen || len(stored.Votes) != 1 {
		t.Errorf("sesión guardada = %+v, se esperaba ABIERTA con el voto de Ana", stored)
	}
}

func TestRevelarPoker_MetodoNoPermitidoDevuelve405(t *testing.T) {
	h := newHandlerWithItem(t)
	openSessionWithVotes(t, h, `{"integrante":"Ana","puntos":5}`)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, "/api/poker-sessions/1/reveal", nil))

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: código = %d, se esperaba %d", method, rec.Code, http.StatusMethodNotAllowed)
		}
	}
}

func TestRevelarPoker_IgnoraElCuerpo(t *testing.T) {
	h := newHandlerWithItem(t)
	openSessionWithVotes(t, h, `{"integrante":"Ana","puntos":5}`)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", `no es JSON {"estado":"ABIERTA"}`)

	expectStatus(t, rec, http.StatusOK)
	var got pokerSessionJSON
	decode(t, rec, &got)
	if got.Estado != "REVELADA" || len(got.Votos) != 1 {
		t.Errorf("sesión = %+v, se esperaba REVELADA con el voto de Ana", got)
	}
}

// HU-25 · Registrar la estimación acordada (Issue #43)
//
// CA-25.9 · POST /api/poker-sessions/{id}/estimate expone el registro. Un
// caso exitoso y un test por cada código de error; el 413 y el texto después
// del JSON están en bodyEndpoints (server_test.go).

// revealSessionWithVotes abre la sesión 1 del ítem 1, registra los votos de
// Ana (5) y Bruno (8) y la revela por la API.
func revealSessionWithVotes(t *testing.T, h http.Handler) {
	t.Helper()
	openSessionWithVotes(t, h, `{"integrante":"Ana","puntos":5}`, `{"integrante":"Bruno","puntos":8}`)
	expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", ""), http.StatusOK)
}

// expectItemStoryPoints verifica los story points del ítem 1 por la API.
func expectItemStoryPoints(t *testing.T, h http.Handler, want int) {
	t.Helper()
	rec := do(t, h, http.MethodGet, "/api/items/1", "")
	expectStatus(t, rec, http.StatusOK)
	var it itemJSON
	decode(t, rec, &it)
	if it.StoryPoints != want {
		t.Errorf("story points del ítem = %d, se esperaban %d", it.StoryPoints, want)
	}
}

func TestRegistrarEstimacionPoker_Devuelve200ConSesionCerradaYStoryPointsEnElItem(t *testing.T) {
	h := newHandlerWithItem(t)
	revealSessionWithVotes(t, h)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", `{"puntos":13}`)

	expectStatus(t, rec, http.StatusOK)
	var got pokerSessionJSON
	decode(t, rec, &got)
	if got.ID != 1 || got.ItemID != 1 || got.Estado != "CERRADA" || got.Ronda != 1 {
		t.Errorf("sesión = %+v, se esperaba la sesión 1 del ítem 1, CERRADA, en la ronda 1", got)
	}
	want := []pokerVoteJSON{{Integrante: "Ana", Puntos: 5}, {Integrante: "Bruno", Puntos: 8}}
	if !slices.Equal(got.Votos, want) {
		t.Errorf("votos = %+v, se esperaba %+v", got.Votos, want)
	}
	if !slices.Equal(got.Cartas, poker.Cards()) {
		t.Errorf("cartas = %v, se esperaba %v", got.Cartas, poker.Cards())
	}
	expectItemStoryPoints(t, h, 13)
}

func TestRegistrarEstimacionPoker_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandlerWithItem(t), http.MethodPost, "/api/poker-sessions/abc/estimate", `{"puntos":5}`)

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

func TestRegistrarEstimacionPoker_JSONInvalidoDevuelve400(t *testing.T) {
	for _, body := range []string{`no es JSON`, `{"puntos":"5"}`, `{"puntos":5.5}`} {
		t.Run(body, func(t *testing.T) {
			h := newHandlerWithItem(t)
			revealSessionWithVotes(t, h)

			rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", body)

			expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
			expectItemWithoutStoryPoints(t, h)
		})
	}
}

func TestRegistrarEstimacionPoker_CartaFueraDeLaEscalaDevuelve400(t *testing.T) {
	for _, body := range []string{`{"puntos":4}`, `{"puntos":0}`, `{}`} {
		t.Run(body, func(t *testing.T) {
			h := newHandlerWithItem(t)
			revealSessionWithVotes(t, h)

			rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", body)

			expectError(t, rec, http.StatusBadRequest, poker.ErrInvalidCard.Error())
			expectItemWithoutStoryPoints(t, h)
		})
	}
}

func TestRegistrarEstimacionPoker_SesionInexistenteDevuelve404(t *testing.T) {
	rec := do(t, newHandlerWithItem(t), http.MethodPost, "/api/poker-sessions/999/estimate", `{"puntos":5}`)

	expectError(t, rec, http.StatusNotFound, poker.ErrSessionNotFound.Error())
}

// newEstimateHandler arma el handler con el repositorio de poker indicado.
func newEstimateHandler(pokerRepo poker.Repository) http.Handler {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	pokerManager := poker.NewManager(pokerRepo, items)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	return httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)
}

func TestRegistrarEstimacionPoker_ItemInexistenteDevuelve404(t *testing.T) {
	pokerRepo := poker.NewMemoryRepository()
	// La sesión se guarda directamente en el repositorio, apuntando a un ítem
	// que no existe.
	session, err := pokerRepo.Save(poker.Session{ItemID: 99, Status: poker.StatusRevealed, Round: 1, Votes: []poker.Vote{{Voter: "Ana", Points: 5}}})
	if err != nil {
		t.Fatal(err)
	}
	h := newEstimateHandler(pokerRepo)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", `{"puntos":5}`)

	expectError(t, rec, http.StatusNotFound, backlog.ErrItemNotFound.Error())
	stored, err := pokerRepo.GetByID(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != poker.StatusRevealed {
		t.Errorf("estado guardado = %q, se esperaba REVELADA", stored.Status)
	}
}

func TestRegistrarEstimacionPoker_SesionNoReveladaDevuelve409(t *testing.T) {
	t.Run("ABIERTA", func(t *testing.T) {
		h := newHandlerWithItem(t)
		openSessionWithVotes(t, h, `{"integrante":"Ana","puntos":5}`)

		rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", `{"puntos":5}`)

		expectError(t, rec, http.StatusConflict, poker.ErrSessionNotRevealed.Error())
		expectItemWithoutStoryPoints(t, h)
	})
	t.Run("CERRADA", func(t *testing.T) {
		h := newHandlerWithItem(t)
		revealSessionWithVotes(t, h)
		expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", `{"puntos":5}`), http.StatusOK)

		rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", `{"puntos":8}`)

		expectError(t, rec, http.StatusConflict, poker.ErrSessionNotRevealed.Error())
		expectItemStoryPoints(t, h, 5)
	})
}

func TestRegistrarEstimacionPoker_ItemTerminadoDevuelve409(t *testing.T) {
	h := newHandlerWithItem(t)
	revealSessionWithVotes(t, h)
	for _, estado := range []string{"EN_CURSO", "TERMINADO"} {
		expectStatus(t, do(t, h, http.MethodPatch, "/api/items/1/estado", `{"estado":"`+estado+`"}`), http.StatusOK)
	}

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", `{"puntos":5}`)

	expectError(t, rec, http.StatusConflict, backlog.ErrItemFinished.Error())
	expectItemWithoutStoryPoints(t, h)
}

// CA-25.6 · Si falla el guardado de la sesión, la API responde 500 sin
// detalle, el ítem queda estimado y la sesión sigue REVELADA (RN-25.7).
func TestRegistrarEstimacionPoker_ErrorAlGuardarLaSesionDevuelve500SinDetalle(t *testing.T) {
	pokerRepo := &failingVoteHTTPRepository{MemoryRepository: poker.NewMemoryRepository()}
	h := newEstimateHandler(pokerRepo)
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem), http.StatusCreated)
	revealSessionWithVotes(t, h)
	pokerRepo.failUpdate = true

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", `{"puntos":5}`)

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
	expectItemStoryPoints(t, h, 5)
	stored, err := pokerRepo.GetByID(1)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != poker.StatusRevealed {
		t.Errorf("estado guardado = %q, se esperaba REVELADA", stored.Status)
	}
}

func TestRegistrarEstimacionPoker_MetodoNoPermitidoDevuelve405(t *testing.T) {
	h := newHandlerWithItem(t)
	revealSessionWithVotes(t, h)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, "/api/poker-sessions/1/estimate", nil))

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: código = %d, se esperaba %d", method, rec.Code, http.StatusMethodNotAllowed)
		}
	}
}

// HU-24 · Detectar diferencias y repetir ronda (Issue #42)
//
// CA-24.12 · POST /api/poker-sessions/{id}/reveal suma el campo comparacion y
// POST /api/poker-sessions/{id}/rounds repite la ronda. Un caso exitoso y un
// test por cada código de error.

type pokerVoteGroupJSON struct {
	Puntos      int      `json:"puntos"`
	Integrantes []string `json:"integrantes"`
}

type pokerComparisonJSON struct {
	HayDiferencias bool               `json:"hay_diferencias"`
	Menor          pokerVoteGroupJSON `json:"menor"`
	Mayor          pokerVoteGroupJSON `json:"mayor"`
}

type pokerRevealJSON struct {
	pokerSessionJSON
	Comparacion *pokerComparisonJSON `json:"comparacion"`
}

// expectVoteGroupJSON verifica un extremo de la comparación.
func expectVoteGroupJSON(t *testing.T, what string, got pokerVoteGroupJSON, points int, voters ...string) {
	t.Helper()
	if got.Puntos != points || !slices.Equal(got.Integrantes, voters) {
		t.Errorf("%s = %+v, se esperaba %d de %v", what, got, points, voters)
	}
}

func TestRevelarPoker_IncluyeComparacionConDiferencias(t *testing.T) {
	h := newHandlerWithItem(t)
	openSessionWithVotes(t, h, `{"integrante":"Ana","puntos":5}`, `{"integrante":"Bruno","puntos":8}`)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", "")

	expectStatus(t, rec, http.StatusOK)
	var got pokerRevealJSON
	decode(t, rec, &got)
	if got.Estado != "REVELADA" || len(got.Votos) != 2 {
		t.Errorf("sesión = %+v, se esperaba REVELADA con dos votos", got.pokerSessionJSON)
	}
	if got.Comparacion == nil {
		t.Fatal("comparacion ausente, se esperaba en la respuesta de revelar")
	}
	if !got.Comparacion.HayDiferencias {
		t.Error("hay_diferencias = false, se esperaba true")
	}
	expectVoteGroupJSON(t, "menor", got.Comparacion.Menor, 5, "Ana")
	expectVoteGroupJSON(t, "mayor", got.Comparacion.Mayor, 8, "Bruno")
}

func TestRevelarPoker_IncluyeComparacionSinDiferencias(t *testing.T) {
	h := newHandlerWithItem(t)
	openSessionWithVotes(t, h, `{"integrante":"Ana","puntos":5}`, `{"integrante":"Bruno","puntos":5}`)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", "")

	expectStatus(t, rec, http.StatusOK)
	var got pokerRevealJSON
	decode(t, rec, &got)
	if got.Comparacion == nil {
		t.Fatal("comparacion ausente, se esperaba en la respuesta de revelar")
	}
	if got.Comparacion.HayDiferencias {
		t.Error("hay_diferencias = true, se esperaba false")
	}
	expectVoteGroupJSON(t, "menor", got.Comparacion.Menor, 5, "Ana", "Bruno")
	expectVoteGroupJSON(t, "mayor", got.Comparacion.Mayor, 5, "Ana", "Bruno")
}

// expectNoComparison verifica que la respuesta no tenga el campo comparacion.
func expectNoComparison(t *testing.T, what string, rec *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]any
	decode(t, rec, &body)
	if _, ok := body["comparacion"]; ok {
		t.Errorf("%s: la respuesta incluye comparacion, solo debe incluirla revelar", what)
	}
}

// Abrir, votar, repetir la ronda y registrar la estimación no incluyen
// comparacion: su JSON no cambia.
func TestVotarPoker_RespuestaSinComparacion(t *testing.T) {
	h := newHandlerWithItem(t)

	rec := do(t, h, http.MethodPost, "/api/items/1/poker-sessions", "")
	expectStatus(t, rec, http.StatusCreated)
	expectNoComparison(t, "abrir", rec)

	rec = do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", `{"integrante":"Ana","puntos":5}`)
	expectStatus(t, rec, http.StatusOK)
	expectNoComparison(t, "votar", rec)

	expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", ""), http.StatusOK)
	rec = do(t, h, http.MethodPost, "/api/poker-sessions/1/rounds", "")
	expectStatus(t, rec, http.StatusOK)
	expectNoComparison(t, "repetir ronda", rec)

	expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", `{"integrante":"Ana","puntos":5}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", ""), http.StatusOK)
	rec = do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", `{"puntos":5}`)
	expectStatus(t, rec, http.StatusOK)
	expectNoComparison(t, "registrar estimación", rec)
}

func TestRepetirRondaPoker_Devuelve200ConRondaSiguienteYSinVotos(t *testing.T) {
	h := newHandlerWithItem(t)
	revealSessionWithVotes(t, h)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/rounds", "")

	expectStatus(t, rec, http.StatusOK)
	var got pokerSessionJSON
	decode(t, rec, &got)
	if got.ID != 1 || got.ItemID != 1 || got.Estado != "ABIERTA" || got.Ronda != 2 {
		t.Errorf("sesión = %+v, se esperaba la sesión 1 del ítem 1, ABIERTA, en la ronda 2", got)
	}
	if got.Votos == nil || len(got.Votos) != 0 {
		t.Errorf("votos = %#v, se esperaba una lista vacía", got.Votos)
	}
	if !slices.Equal(got.Cartas, poker.Cards()) {
		t.Errorf("cartas = %v, se esperaba %v", got.Cartas, poker.Cards())
	}
	// La sesión vuelve a aceptar votos.
	expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/votes", `{"integrante":"Ana","puntos":13}`), http.StatusOK)
}

func TestRepetirRondaPoker_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandlerWithItem(t), http.MethodPost, "/api/poker-sessions/abc/rounds", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

func TestRepetirRondaPoker_InexistenteDevuelve404(t *testing.T) {
	h := newHandlerWithItem(t)
	revealSessionWithVotes(t, h)

	for _, id := range []string{"999", "0", "-1"} {
		rec := do(t, h, http.MethodPost, "/api/poker-sessions/"+id+"/rounds", "")

		expectError(t, rec, http.StatusNotFound, poker.ErrSessionNotFound.Error())
	}
}

func TestRepetirRondaPoker_SesionNoReveladaDevuelve409(t *testing.T) {
	h := newHandlerWithItem(t)
	openSessionWithVotes(t, h, `{"integrante":"Ana","puntos":5}`)

	// ABIERTA.
	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/rounds", "")
	expectError(t, rec, http.StatusConflict, poker.ErrSessionNotRevealed.Error())

	// CERRADA.
	expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/reveal", ""), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPost, "/api/poker-sessions/1/estimate", `{"puntos":5}`), http.StatusOK)
	rec = do(t, h, http.MethodPost, "/api/poker-sessions/1/rounds", "")
	expectError(t, rec, http.StatusConflict, poker.ErrSessionNotRevealed.Error())
}

func TestRepetirRondaPoker_OtraSesionAbiertaDevuelve409(t *testing.T) {
	h := newHandlerWithItem(t)
	revealSessionWithVotes(t, h)
	expectStatus(t, do(t, h, http.MethodPost, "/api/items/1/poker-sessions", ""), http.StatusCreated)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/rounds", "")

	expectError(t, rec, http.StatusConflict, poker.ErrSessionAlreadyOpen.Error())
}

func TestRepetirRondaPoker_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	pokerRepo := &failingVoteHTTPRepository{MemoryRepository: poker.NewMemoryRepository()}
	h := newEstimateHandler(pokerRepo)
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/1/backlog", validItem), http.StatusCreated)
	revealSessionWithVotes(t, h)
	pokerRepo.failUpdate = true

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/rounds", "")

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
	stored, err := pokerRepo.GetByID(1)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != poker.StatusRevealed || stored.Round != 1 || len(stored.Votes) != 2 {
		t.Errorf("sesión guardada = %+v, se esperaba REVELADA en la ronda 1 con dos votos", stored)
	}
}

func TestRepetirRondaPoker_MetodoNoPermitidoDevuelve405(t *testing.T) {
	h := newHandlerWithItem(t)
	revealSessionWithVotes(t, h)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, "/api/poker-sessions/1/rounds", nil))

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: código = %d, se esperaba %d", method, rec.Code, http.StatusMethodNotAllowed)
		}
	}
}

func TestRepetirRondaPoker_IgnoraElCuerpo(t *testing.T) {
	h := newHandlerWithItem(t)
	revealSessionWithVotes(t, h)

	rec := do(t, h, http.MethodPost, "/api/poker-sessions/1/rounds", `no es JSON {"ronda":99}`)

	expectStatus(t, rec, http.StatusOK)
	var got pokerSessionJSON
	decode(t, rec, &got)
	if got.Estado != "ABIERTA" || got.Ronda != 2 {
		t.Errorf("sesión = %+v, se esperaba ABIERTA en la ronda 2", got)
	}
}
