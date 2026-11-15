package member

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sync"
	"testing"
	"uuid"
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

// pidEnv arma el entorno de HU-15 (dos proyectos, 1 y 2) con acceso al
// repositorio en memoria.
type pidEnv struct {
	*memberEnv
	manager *Manager
}

func newPIDEnv(t *testing.T) pidEnv {
	t.Helper()
	e := newMemberEnv(t)
	return pidEnv{memberEnv: e, manager: e.members}
}

func (e pidEnv) all() []Member { return slices.Clone(e.repo.MemoryRepository.members) }

func (e pidEnv) saved(t *testing.T, id int) Member {
	t.Helper()
	for _, m := range e.repo.MemoryRepository.members {
		if m.ID == id {
			return m
		}
	}
	t.Fatalf("no hay ningún integrante guardado con el ID %d", id)
	return Member{}
}

// HU-52 · CA-52.1 · Verifica que Register le asigne al integrante un identificador público UUID versión 4 en forma canónica, que no sea el nulo, y que el repositorio guarde ese mismo valor.
func TestRegister_AsignaUnUUIDVersion4AlIntegrante(t *testing.T) {
	env := newPIDEnv(t)
	x, err := env.manager.Register(1, "Ana Pérez", "PRODUCT_BUILDER")

	if err != nil {
		t.Fatalf("Register() devolvió un error inesperado: %v", err)
	}
	expectCanonicalV4(t, "Register", x.PublicID)
	if x.PublicID == uuid.Nil() {
		t.Error("PublicID es el UUID nulo")
	}
	if saved := env.saved(t, x.ID); saved.PublicID != x.PublicID {
		t.Errorf("PublicID guardado = %s, se esperaba el devuelto por Register (%s)", saved.PublicID, x.PublicID)
	}
}

// HU-52 · CA-52.2 · Verifica que 1000 integrantes registrados seguidos tengan identificadores públicos distintos, todos versión 4 y no correlativos.
func TestRegister_MilIntegrantesConUUIDDistintosYNoCorrelativos(t *testing.T) {
	const total = 1000
	env := newPIDEnv(t)

	ids := make([]uuid.UUID, 0, total)
	seen := make(map[uuid.UUID]bool, total)

	for i := range total {
		x, err := env.manager.Register(1, fmt.Sprintf("Integrante %04d", i), "PRODUCT_BUILDER")
		if err != nil {
			t.Fatalf("Register(%d) devolvió un error inesperado: %v", i, err)
		}
		if seen[x.PublicID] {
			t.Fatalf("Register(%d): PublicID %s repetido", i, x.PublicID)
		}
		seen[x.PublicID] = true
		expectCanonicalV4(t, fmt.Sprintf("Register(%d)", i), x.PublicID)
		ids = append(ids, x.PublicID)
	}

	sorted := slices.Clone(ids)
	slices.SortFunc(sorted, uuid.UUID.Compare)
	if slices.Equal(sorted, ids) {
		t.Error("ordenar los PublicID da el orden de creación: son correlativos")
	}
}

// HU-52 · CA-52.7 · Verifica que 50 registros simultáneos generen 50 identificadores públicos distintos, sin condiciones de carrera con go test -race.
func TestRegister_RegistrosSimultaneosConUUIDDistintos(t *testing.T) {
	const total = 50
	env := newPIDEnv(t)

	var wg sync.WaitGroup
	for i := range total {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := env.manager.Register(1, fmt.Sprintf("Integrante %04d", i), "PRODUCT_BUILDER"); err != nil {
				t.Errorf("Register(%d) devolvió un error inesperado: %v", i, err)
			}
		}()
	}
	wg.Wait()

	all := env.all()
	if len(all) != total {
		t.Fatalf("se guardaron %d integrantes, se esperaban %d", len(all), total)
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

// HU-52 · CA-52.3 · Verifica que GetByPublicID devuelva el integrante con ese identificador público entre varios, como una copia.
func TestGetByPublicID_DevuelveElIntegrante(t *testing.T) {
	env := newPIDEnv(t)

	if _, err := env.manager.Register(1, "Ana Pérez", "PRODUCT_BUILDER"); err != nil {
		t.Fatalf("Register() devolvió un error inesperado: %v", err)
	}
	want, err := env.manager.Register(2, "Bruno Díaz", "AGILE_ENABLER")
	if err != nil {
		t.Fatalf("Register() devolvió un error inesperado: %v", err)
	}

	got, err := env.manager.GetByPublicID(want.PublicID)

	if err != nil {
		t.Fatalf("GetByPublicID() devolvió un error inesperado: %v", err)
	}
	if got != want {
		t.Errorf("GetByPublicID() = %+v, se esperaba %+v", got, want)
	}

}

// HU-52 · CA-52.4 · Verifica que GetByPublicID de un UUID que no existe, o del UUID nulo, devuelva ErrMemberNotFound.
func TestGetByPublicID_ElIntegranteInexistenteDevuelveErrMemberNotFound(t *testing.T) {
	env := newPIDEnv(t)
	if _, err := env.manager.Register(1, "Ana Pérez", "PRODUCT_BUILDER"); err != nil {
		t.Fatalf("Register() devolvió un error inesperado: %v", err)
	}

	for _, id := range []uuid.UUID{uuid.NewV4(), uuid.Nil()} {
		got, err := env.manager.GetByPublicID(id)
		if !errors.Is(err, ErrMemberNotFound) {
			t.Errorf("GetByPublicID(%s): error = %v, se esperaba ErrMemberNotFound", id, err)
		}
		if got.ID != 0 {
			t.Errorf("GetByPublicID(%s) = %+v, se esperaba un valor vacío", id, got)
		}
	}
}

// HU-52 · CA-52.5 · Verifica que modificar un integrante conserve el identificador público del integrante.
func TestUpdate_ConservaElUUIDDelIntegrante(t *testing.T) {
	env := newPIDEnv(t)
	x, err := env.manager.Register(1, "Ana Pérez", "PRODUCT_BUILDER")
	if err != nil {
		t.Fatalf("Register() devolvió un error inesperado: %v", err)
	}
	expectCanonicalV4(t, "Register", x.PublicID)

	ops := []struct {
		name string
		run  func() (Member, error)
	}{
		{"Update", func() (Member, error) { return env.manager.Update(x.ID, "Ana María Pérez", "PRODUCT_ARCHITECT") }},
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
