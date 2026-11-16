package effort

import (
	"errors"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/member"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-54 · Registrar esfuerzo solo con integrantes del proyecto (Issue #219).

// anyMember acepta cualquier nombre tal como llega: lo usan los tests de las
// historias anteriores a HU-54, que no registran integrantes (D5).
type anyMember struct{}

func (anyMember) CanonicalName(_ int, name string) (string, error) {
	return strings.TrimSpace(name), nil
}

// projectMembers es un falso con los integrantes de cada proyecto.
type projectMembers struct {
	byProject map[int][]string
	asked     []int
}

func (p *projectMembers) CanonicalName(projectID int, name string) (string, error) {
	p.asked = append(p.asked, projectID)
	for _, m := range p.byProject[projectID] {
		if strings.EqualFold(m, strings.TrimSpace(name)) {
			return m, nil
		}
	}
	return "", member.ErrNotProjectMember
}

// envWithMembers arma la pila de HU-26 con un Manager que valida contra los
// integrantes indicados, un sprint ABIERTO y un ítem asignado a él.
func envWithMembers(t *testing.T, names ...string) (*effortEnv, *projectMembers, backlog.Item, sprint.Sprint) {
	t.Helper()
	env := newEffortEnv(t)
	members := &projectMembers{byProject: map[int][]string{env.projectID: names}}
	env.manager = NewManager(env.memory, env.sprints, env.items, members)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item := env.addItem(t, "Registrar venta", backlog.StatusInProgress, sp.ID)
	return env, members, item, sp
}

// HU-54 · CA-54.3 · Verifica que un integrante del proyecto registre su esfuerzo y que se consulte el proyecto del ítem.
func TestRegister_IntegranteDelProyecto(t *testing.T) {
	env, members, item, sp := envWithMembers(t, "García Santiago")

	got, err := env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Desarrollo", 4)

	if err != nil {
		t.Fatalf("Register devolvió %v", err)
	}
	if got.Member != "García Santiago" || len(env.memory.efforts) != 1 {
		t.Errorf("esfuerzo = %+v, guardados = %d", got, len(env.memory.efforts))
	}
	if len(members.asked) != 1 || members.asked[0] != env.projectID {
		t.Errorf("proyectos consultados = %v, se esperaba [%d]", members.asked, env.projectID)
	}
}

// HU-54 · CA-54.4 · Verifica que el esfuerzo se guarde con el nombre registrado aunque se escriba con otras mayúsculas.
func TestRegister_GuardaElNombreRegistrado(t *testing.T) {
	env, _, item, sp := envWithMembers(t, "García Santiago")

	got, err := env.manager.Register(sp.ID, item.ID, "GARCÍA SANTIAGO", testEffortDate, "Desarrollo", 4)

	if err != nil {
		t.Fatalf("Register devolvió %v", err)
	}
	if got.Member != "García Santiago" || env.memory.efforts[0].Member != "García Santiago" {
		t.Errorf("integrante = %q, guardado = %q; se esperaba \"García Santiago\"", got.Member, env.memory.efforts[0].Member)
	}
}

// HU-54 · CA-54.3 · Verifica que se rechace con ErrNotProjectMember a quien no es integrante, sin guardar el esfuerzo.
func TestRegister_RechazaNoIntegranteSinGuardar(t *testing.T) {
	env, _, item, sp := envWithMembers(t, "García Santiago")

	_, err := env.manager.Register(sp.ID, item.ID, "hernan", testEffortDate, "Desarrollo", 4)

	if !errors.Is(err, member.ErrNotProjectMember) {
		t.Fatalf("Register devolvió %v, se esperaba ErrNotProjectMember", err)
	}
	if len(env.memory.efforts) != 0 {
		t.Errorf("esfuerzos guardados = %d, se esperaba ninguno", len(env.memory.efforts))
	}
}

// HU-54 · CA-54.3 · Verifica que un ítem no asignado al sprint se informe antes que el integrante (D3).
func TestRegister_ItemNoAsignadoAntesQueIntegrante(t *testing.T) {
	env, members, _, sp := envWithMembers(t, "García Santiago")
	other := env.addItem(t, "Emitir factura", backlog.StatusPending, 0)

	_, err := env.manager.Register(sp.ID, other.ID, "hernan", testEffortDate, "Desarrollo", 2)

	if !errors.Is(err, sprint.ErrItemNotAssignedToSprint) {
		t.Fatalf("Register devolvió %v, se esperaba ErrItemNotAssignedToSprint", err)
	}
	if len(members.asked) != 0 {
		t.Errorf("se consultaron los integrantes %v antes de validar el ítem", members.asked)
	}
}
