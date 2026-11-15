package poker

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sync"
	"testing"
	"uuid"

	"seguimiento-medicion-gpc/internal/backlog"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197)

// canonicalV4 es la forma canónica de un UUID versión 4, variante RFC 9562,
// en minúsculas (CA-52.1).
var canonicalV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// expectCanonicalV4 verifica que id sea un UUID versión 4 en forma canónica.
func expectCanonicalV4(t *testing.T, what string, id uuid.UUID) {
	t.Helper()
	if s := id.String(); !canonicalV4.MatchString(s) {
		t.Errorf("%s: PublicID = %q, se esperaba un UUID versión 4 en forma canónica", what, s)
	}
}

// pidEnv arma el entorno de HU-21 con el proyecto de prueba.
type pidEnv struct {
	*testEnv
}

func newPIDEnv(t *testing.T) pidEnv {
	t.Helper()
	return pidEnv{testEnv: newTestEnv(t)}
}

// items crea n ítems PENDIENTE en el proyecto de prueba, uno por sesión.
func (e pidEnv) itemsFor(t *testing.T, n int) []backlog.Item {
	t.Helper()
	items := make([]backlog.Item, n)
	for i := range n {
		items[i] = e.createItem(t, fmt.Sprintf("Ítem %04d", i), backlog.StatusPending, 0)
	}
	return items
}

func (e pidEnv) all() []Session { return slices.Clone(e.repo.sessions) }

func (e pidEnv) saved(t *testing.T, id int) Session {
	t.Helper()
	for _, x := range e.repo.sessions {
		if x.ID == id {
			return x
		}
	}
	t.Fatalf("no hay ninguna sesión guardada con el ID %d", id)
	return Session{}
}

// HU-52 · CA-52.1 · Verifica que OpenSession le asigne a la sesión un identificador público UUID versión 4 en forma canónica, que no sea el nulo, y que el repositorio guarde ese mismo valor.
func TestOpenSession_AsignaUnUUIDVersion4ALaSesion(t *testing.T) {
	env := newPIDEnv(t)
	x, err := env.manager.OpenSession(env.createItem(t, "Login de usuarios", backlog.StatusPending, 0).ID)

	if err != nil {
		t.Fatalf("OpenSession() devolvió un error inesperado: %v", err)
	}
	expectCanonicalV4(t, "OpenSession", x.PublicID)
	if x.PublicID == uuid.Nil() {
		t.Error("PublicID es el UUID nulo")
	}
	if saved := env.saved(t, x.ID); saved.PublicID != x.PublicID {
		t.Errorf("PublicID guardado = %s, se esperaba el devuelto por OpenSession (%s)", saved.PublicID, x.PublicID)
	}
}

// HU-52 · CA-52.2 · Verifica que 1000 sesiones abiertas seguidas tengan identificadores públicos distintos, todos versión 4 y no correlativos.
func TestOpenSession_MilSesionesConUUIDDistintosYNoCorrelativos(t *testing.T) {
	const total = 1000
	env := newPIDEnv(t)
	items := env.itemsFor(t, total)
	ids := make([]uuid.UUID, 0, total)
	seen := make(map[uuid.UUID]bool, total)

	for i := range total {
		x, err := env.manager.OpenSession(items[i].ID)
		if err != nil {
			t.Fatalf("OpenSession(%d) devolvió un error inesperado: %v", i, err)
		}
		if seen[x.PublicID] {
			t.Fatalf("OpenSession(%d): PublicID %s repetido", i, x.PublicID)
		}
		seen[x.PublicID] = true
		expectCanonicalV4(t, fmt.Sprintf("OpenSession(%d)", i), x.PublicID)
		ids = append(ids, x.PublicID)
	}

	sorted := slices.Clone(ids)
	slices.SortFunc(sorted, uuid.UUID.Compare)
	if slices.Equal(sorted, ids) {
		t.Error("ordenar los PublicID da el orden de creación: son correlativos")
	}
}

// HU-52 · CA-52.7 · Verifica que 50 aperturas simultáneas generen 50 identificadores públicos distintos, sin condiciones de carrera con go test -race.
func TestOpenSession_AperturasSimultaneasConUUIDDistintos(t *testing.T) {
	const total = 50
	env := newPIDEnv(t)
	items := env.itemsFor(t, total)

	var wg sync.WaitGroup
	for i := range total {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := env.manager.OpenSession(items[i].ID); err != nil {
				t.Errorf("OpenSession(%d) devolvió un error inesperado: %v", i, err)
			}
		}()
	}
	wg.Wait()

	all := env.all()
	if len(all) != total {
		t.Fatalf("se guardaron %d sesiones, se esperaban %d", len(all), total)
	}
	seen := map[uuid.UUID]bool{}
	for _, x := range all {
		expectCanonicalV4(t, fmt.Sprintf("ID %d", x.ID), x.PublicID)
		seen[x.PublicID] = true
	}
	if len(seen) != total {
		t.Errorf("hay %d PublicID distintos, se esperaban %d", len(seen), total)
	}
}

// HU-52 · CA-52.3 · Verifica que GetByPublicID devuelva la sesión con ese identificador público entre varias, como una copia.
func TestGetByPublicID_DevuelveLaSesion(t *testing.T) {
	env := newPIDEnv(t)
	items := env.itemsFor(t, 2)
	if _, err := env.manager.OpenSession(items[0].ID); err != nil {
		t.Fatalf("OpenSession() devolvió un error inesperado: %v", err)
	}
	want, err := env.manager.OpenSession(items[1].ID)
	if err != nil {
		t.Fatalf("OpenSession() devolvió un error inesperado: %v", err)
	}

	got, err := env.manager.GetByPublicID(want.PublicID)

	if err != nil {
		t.Fatalf("GetByPublicID() devolvió un error inesperado: %v", err)
	}
	if got.ID != want.ID || got.PublicID != want.PublicID || got.ItemID != want.ItemID || got.Status != want.Status {
		t.Errorf("GetByPublicID() = %+v, se esperaba %+v", got, want)
	}

}

// HU-52 · CA-52.4 · Verifica que GetByPublicID de un UUID que no existe, o del UUID nulo, devuelva ErrSessionNotFound.
func TestGetByPublicID_LaSesionInexistenteDevuelveErrSessionNotFound(t *testing.T) {
	env := newPIDEnv(t)
	if _, err := env.manager.OpenSession(env.createItem(t, "Login de usuarios", backlog.StatusPending, 0).ID); err != nil {
		t.Fatalf("OpenSession() devolvió un error inesperado: %v", err)
	}

	for _, id := range []uuid.UUID{uuid.NewV4(), uuid.Nil()} {
		got, err := env.manager.GetByPublicID(id)
		if !errors.Is(err, ErrSessionNotFound) {
			t.Errorf("GetByPublicID(%s): error = %v, se esperaba ErrSessionNotFound", id, err)
		}
		if got.ID != 0 {
			t.Errorf("GetByPublicID(%s) = %+v, se esperaba un valor vacío", id, got)
		}
	}
}

// HU-52 · CA-52.5 · Verifica que votar, revelar, repetir la ronda y registrar la estimación conserven el identificador público de la sesión.
func TestOperaciones_ConservanElUUIDDeLaSesion(t *testing.T) {
	env := newPIDEnv(t)
	x, err := env.manager.OpenSession(env.createItem(t, "Login de usuarios", backlog.StatusPending, 0).ID)
	if err != nil {
		t.Fatalf("OpenSession() devolvió un error inesperado: %v", err)
	}
	expectCanonicalV4(t, "OpenSession", x.PublicID)

	ops := []struct {
		name string
		run  func() (Session, error)
	}{
		{"Vote Ana", func() (Session, error) { return env.manager.Vote(x.ID, "Ana", 3) }},
		{"Vote Bruno", func() (Session, error) { return env.manager.Vote(x.ID, "Bruno", 8) }},
		{"Reveal", func() (Session, error) { return env.manager.Reveal(x.ID) }},
		{"RepeatRound", func() (Session, error) { return env.manager.RepeatRound(x.ID) }},
		{"Vote ronda 2", func() (Session, error) { return env.manager.Vote(x.ID, "Ana", 5) }},
		{"Reveal ronda 2", func() (Session, error) { return env.manager.Reveal(x.ID) }},
		{"Estimate", func() (Session, error) { return env.manager.Estimate(x.ID, 5) }},
	}
	for _, op := range ops {
		got, err := op.run()
		if err != nil {
			t.Fatalf("%s devolvió un error inesperado: %v", op.name, err)
		}
		if got.PublicID != x.PublicID {
			t.Errorf("%s: PublicID = %s, se esperaba %s", op.name, got.PublicID, x.PublicID)
		}
		if saved := env.saved(t, x.ID); saved.PublicID != x.PublicID {
			t.Errorf("%s: PublicID guardado = %s, se esperaba %s", op.name, saved.PublicID, x.PublicID)
		}
	}
}
