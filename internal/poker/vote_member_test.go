package poker

import (
	"errors"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/member"
)

// HU-54 · Votar solo con integrantes del proyecto (Issue #219).

// anyMember acepta cualquier nombre tal como llega: lo usan los tests de las
// historias anteriores a HU-54, que no registran integrantes (D5).
type anyMember struct{}

func (anyMember) CanonicalName(_ int, name string) (string, error) {
	return strings.TrimSpace(name), nil
}

// projectMembers es un falso con los integrantes de cada proyecto. Guarda los
// proyectos consultados para verificar que se usa el proyecto del ítem.
type projectMembers struct {
	byProject map[int][]string
	asked     []int
	err       error
}

func (p *projectMembers) CanonicalName(projectID int, name string) (string, error) {
	p.asked = append(p.asked, projectID)
	if p.err != nil {
		return "", p.err
	}
	for _, m := range p.byProject[projectID] {
		if strings.EqualFold(m, strings.TrimSpace(name)) {
			return m, nil
		}
	}
	return "", member.ErrNotProjectMember
}

// openSessionWithMembers abre una sesión sobre un ítem del proyecto de prueba,
// con un Manager que valida contra los integrantes indicados.
func openSessionWithMembers(t *testing.T, names ...string) (*testEnv, *projectMembers, Session) {
	t.Helper()
	env := newTestEnv(t)
	members := &projectMembers{byProject: map[int][]string{env.projectID: names}}
	env.manager = NewManager(env.repo, env.items, members)
	item := env.createItem(t, "Registrar venta", backlog.StatusPending, 0)
	session, err := env.manager.OpenSession(item.ID)
	if err != nil {
		t.Fatalf("abrir sesión: %v", err)
	}
	return env, members, session
}

func storedVotes(t *testing.T, env *testEnv, sessionID int) []Vote {
	t.Helper()
	stored, err := env.repo.GetByID(sessionID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	return stored.Votes
}

// HU-54 · CA-54.2 · Verifica que un integrante del proyecto pueda votar y que se consulte el proyecto del ítem de la sesión.
func TestVote_IntegranteDelProyecto(t *testing.T) {
	env, members, session := openSessionWithMembers(t, "Sosa Ricardo")

	if _, err := env.manager.Vote(session.ID, "Sosa Ricardo", 5); err != nil {
		t.Fatalf("Vote devolvió %v", err)
	}
	if votes := storedVotes(t, env, session.ID); len(votes) != 1 || votes[0] != (Vote{Voter: "Sosa Ricardo", Points: 5}) {
		t.Errorf("votos guardados = %+v", votes)
	}
	if len(members.asked) != 1 || members.asked[0] != env.projectID {
		t.Errorf("proyectos consultados = %v, se esperaba [%d]", members.asked, env.projectID)
	}
}

// HU-54 · CA-54.4 · Verifica que el voto se guarde con el nombre registrado aunque se escriba con otras mayúsculas y espacios.
func TestVote_GuardaElNombreRegistrado(t *testing.T) {
	env, _, session := openSessionWithMembers(t, "Sosa Ricardo")

	if _, err := env.manager.Vote(session.ID, "  sosa ricardo ", 8); err != nil {
		t.Fatalf("Vote devolvió %v", err)
	}
	if votes := storedVotes(t, env, session.ID); len(votes) != 1 || votes[0].Voter != "Sosa Ricardo" {
		t.Errorf("votos guardados = %+v, se esperaba el nombre \"Sosa Ricardo\"", votes)
	}
}

// HU-54 · CA-54.4 · Verifica que el mismo integrante escrito con otras mayúsculas reemplace su voto anterior (HU-22).
func TestVote_ReemplazaElVotoConOtrasMayusculas(t *testing.T) {
	env, _, session := openSessionWithMembers(t, "Sosa Ricardo")

	if _, err := env.manager.Vote(session.ID, "Sosa Ricardo", 5); err != nil {
		t.Fatalf("primer voto: %v", err)
	}
	if _, err := env.manager.Vote(session.ID, "SOSA RICARDO", 8); err != nil {
		t.Fatalf("segundo voto: %v", err)
	}
	if votes := storedVotes(t, env, session.ID); len(votes) != 1 || votes[0] != (Vote{Voter: "Sosa Ricardo", Points: 8}) {
		t.Errorf("votos guardados = %+v, se esperaba un solo voto de \"Sosa Ricardo\" con 8", votes)
	}
}

// HU-54 · CA-54.2 · Verifica que se rechace con ErrNotProjectMember a quien no es integrante, sin guardar el voto.
func TestVote_RechazaNoIntegranteSinGuardar(t *testing.T) {
	env, _, session := openSessionWithMembers(t, "Sosa Ricardo")

	_, err := env.manager.Vote(session.ID, "hernan", 5)

	if !errors.Is(err, member.ErrNotProjectMember) {
		t.Fatalf("Vote devolvió %v, se esperaba ErrNotProjectMember", err)
	}
	if votes := storedVotes(t, env, session.ID); len(votes) != 0 {
		t.Errorf("votos guardados = %+v, se esperaba ninguno", votes)
	}
}

// HU-54 · CA-54.2 · Verifica que con la sesión no abierta se informe ErrSessionNotOpen antes que el integrante (D3).
func TestVote_SesionNoAbiertaAntesQueIntegrante(t *testing.T) {
	env, members, session := openSessionWithMembers(t, "Sosa Ricardo")
	if _, err := env.manager.Vote(session.ID, "Sosa Ricardo", 5); err != nil {
		t.Fatalf("voto previo: %v", err)
	}
	if _, err := env.manager.Reveal(session.ID); err != nil {
		t.Fatalf("revelar: %v", err)
	}
	members.asked = nil

	_, err := env.manager.Vote(session.ID, "hernan", 5)

	if !errors.Is(err, ErrSessionNotOpen) {
		t.Fatalf("Vote devolvió %v, se esperaba ErrSessionNotOpen", err)
	}
	if len(members.asked) != 0 {
		t.Errorf("se consultaron los integrantes %v antes de rechazar la sesión", members.asked)
	}
}

// HU-54 · Error · Verifica que un error al consultar los integrantes se devuelva tal cual y no guarde el voto.
func TestVote_ErrorAlConsultarIntegrantes(t *testing.T) {
	env, members, session := openSessionWithMembers(t, "Sosa Ricardo")
	want := errors.New("falla de la base")
	members.err = want

	if _, err := env.manager.Vote(session.ID, "Sosa Ricardo", 5); !errors.Is(err, want) {
		t.Fatalf("Vote devolvió %v, se esperaba el error de los integrantes", err)
	}
	if votes := storedVotes(t, env, session.ID); len(votes) != 0 {
		t.Errorf("votos guardados = %+v, se esperaba ninguno", votes)
	}
}
