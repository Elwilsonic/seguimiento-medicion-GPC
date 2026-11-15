package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/poker"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197), parte
// de Planning Poker (Buttini Cristobal). Las sesiones y sus ítems se crean en
// el dominio con uuidStack (sprints_uuid_test.go).

// revealedSession abre una sesión sobre un ítem nuevo, con los votos de Ana y
// Bruno, y la revela.
func (s *uuidStack) revealedSession(t *testing.T, ana, bruno int) poker.Session {
	t.Helper()
	session := s.votedSession(t, ana, bruno)
	if _, err := s.poker.Reveal(session.ID); err != nil {
		t.Fatal(err)
	}
	return session
}

// votedSession abre una sesión sobre un ítem nuevo, con los votos de Ana y
// Bruno.
func (s *uuidStack) votedSession(t *testing.T, ana, bruno int) poker.Session {
	t.Helper()
	session, err := s.poker.OpenSession(s.item(t, "Login").ID)
	if err != nil {
		t.Fatal(err)
	}
	for voter, points := range map[string]int{"Ana": ana, "Bruno": bruno} {
		if _, err := s.poker.Vote(session.ID, voter, points); err != nil {
			t.Fatal(err)
		}
	}
	return session
}

// sessionRoutes son las 4 rutas de una sesión (D3, poker.go), con un cuerpo
// válido para cada una.
var sessionRoutes = []struct{ name, path, body string }{
	{"votar", "/votes", `{"integrante":"Ana","puntos":5}`},
	{"revelar", "/reveal", ""},
	{"registrar estimación", "/estimate", `{"puntos":5}`},
	{"repetir ronda", "/rounds", ""},
}

// HU-52 · CA-52.1 · Verifica que abrir una sesión responda 201 con un id que es un UUID versión 4 en minúsculas y con el UUID del ítem en item_id.
func TestAbrirSesion_IDEsUnUUIDVersion4YItemIDEsElDelItem(t *testing.T) {
	s := newUUIDStack(t)
	it := s.item(t, "Login")

	rec := s.serve(http.MethodPost, "/api/items/"+it.PublicID.String()+"/poker-sessions", "")

	expectStatus(t, rec, http.StatusCreated)
	if id := jsonString(t, "sesión abierta", rec, "id"); !uuidV4Canonical.MatchString(id) {
		t.Errorf("id = %q, se esperaba un UUID versión 4 en minúsculas", id)
	}
	if got := jsonString(t, "sesión abierta", rec, "item_id"); got != it.PublicID.String() {
		t.Errorf("item_id = %q, se esperaba el UUID del ítem %q", got, it.PublicID)
	}
}

// HU-52 · CA-52.1 · Verifica que repetir la ronda de una sesión revelada responda la ronda 2 con el mismo UUID de la sesión.
func TestRepetirRonda_ConservaElUUIDDeLaSesion(t *testing.T) {
	s := newUUIDStack(t)
	session := s.revealedSession(t, 3, 8)

	rec := s.serve(http.MethodPost, "/api/poker-sessions/"+session.PublicID.String()+"/rounds", "")

	expectStatus(t, rec, http.StatusOK)
	if got := jsonString(t, "ronda repetida", rec, "id"); got != session.PublicID.String() {
		t.Errorf("id = %q, se esperaba el UUID de la sesión %q", got, session.PublicID)
	}
}

// HU-52 · CA-52.3 · Verifica que las 4 rutas de una sesión acepten su UUID, también en mayúsculas, y respondan 200 con el UUID de la sesión en minúsculas y el del ítem en item_id.
func TestRutasDeSesion_AceptanElUUID(t *testing.T) {
	for _, upper := range []bool{false, true} {
		for _, r := range sessionRoutes {
			s := newUUIDStack(t)
			session := s.votedSession(t, 5, 5)
			if r.name != "votar" && r.name != "revelar" {
				if _, err := s.poker.Reveal(session.ID); err != nil {
					t.Fatal(err)
				}
			}
			id := session.PublicID.String()
			if upper {
				id = strings.ToUpper(id)
			}

			rec := s.serve(http.MethodPost, "/api/poker-sessions/"+id+r.path, r.body)

			if rec.Code != http.StatusOK {
				t.Errorf("%s (mayúsculas: %v): código = %d, se esperaba 200; cuerpo: %s", r.name, upper, rec.Code, rec.Body.String())
				continue
			}
			if got := jsonString(t, r.name, rec, "id"); got != session.PublicID.String() {
				t.Errorf("%s (mayúsculas: %v): id = %q, se esperaba %q", r.name, upper, got, session.PublicID)
			}
			it, err := s.items.GetItem(session.ItemID)
			if err != nil {
				t.Fatal(err)
			}
			if got := jsonString(t, r.name, rec, "item_id"); got != it.PublicID.String() {
				t.Errorf("%s (mayúsculas: %v): item_id = %q, se esperaba %q", r.name, upper, got, it.PublicID)
			}
		}
	}
}

// HU-52 · CA-52.4 · Verifica que un {id} de sesión que no es un UUID canónico (un número, entre llaves, con urn:uuid: o sin guiones) responda 400 con el mensaje de la sesión en las 4 rutas, antes de leer el cuerpo.
func TestRutasDeSesion_IDInvalidoDevuelve400(t *testing.T) {
	s := newUUIDStack(t)
	valid := s.votedSession(t, 5, 5).PublicID.String()
	for _, id := range []string{"1", "{" + valid + "}", "urn:uuid:" + valid, strings.ReplaceAll(valid, "-", "")} {
		for _, r := range sessionRoutes {
			rec := s.serve(http.MethodPost, "/api/poker-sessions/"+id+r.path, "esto no es JSON")
			expectErrorBody(t, r.name+" con "+id, rec, http.StatusBadRequest, "el id de la sesión debe ser un UUID")
		}
	}
}

// HU-52 · CA-52.4 · Verifica que un UUID que no corresponde a ninguna sesión (uno que nunca se asignó o el nulo) responda 404 con ErrSessionNotFound en las 4 rutas.
func TestRutasDeSesion_UUIDInexistenteDevuelve404(t *testing.T) {
	s := newUUIDStack(t)
	for _, id := range []string{missingSessionID, nilProjectID} {
		for _, r := range sessionRoutes {
			rec := s.serve(http.MethodPost, "/api/poker-sessions/"+id+r.path, r.body)
			expectErrorBody(t, r.name+" con "+id, rec, http.StatusNotFound, poker.ErrSessionNotFound.Error())
		}
	}
}

// HU-52 · CA-52.4 · Verifica que votar en una sesión inexistente con el votante vacío responda 400 con el error del votante, como con un ID entero inexistente (HU-22).
func TestVotar_SeConservaElOrdenDeErrores(t *testing.T) {
	s := newUUIDStack(t)

	rec := s.serve(http.MethodPost, "/api/poker-sessions/"+missingSessionID+"/votes", `{"integrante":"","puntos":5}`)

	expectErrorBody(t, "votante vacío en una sesión inexistente", rec, http.StatusBadRequest, poker.ErrEmptyVoter.Error())
}

// HU-52 · CA-52.4 · Verifica que abrir una sesión con un {id} de ítem que no es un UUID responda 400 con el mensaje del ítem, y con un UUID inexistente, 404 con ErrItemNotFound.
func TestAbrirSesion_ItemInvalidoOInexistente(t *testing.T) {
	s := newUUIDStack(t)

	rec := s.serve(http.MethodPost, "/api/items/1/poker-sessions", "")
	expectErrorBody(t, "ítem numérico", rec, http.StatusBadRequest, "el id del ítem debe ser un UUID")

	rec = s.serve(http.MethodPost, "/api/items/"+missingPokerItemID+"/poker-sessions", "")
	expectErrorBody(t, "ítem inexistente", rec, http.StatusNotFound, backlog.ErrItemNotFound.Error())
}
