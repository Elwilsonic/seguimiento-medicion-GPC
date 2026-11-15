package backlog

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

// publicIDEnv arma un Manager de backlog con un proyecto creado.
func publicIDEnv(t *testing.T) (*Manager, *MemoryRepository, int) {
	t.Helper()
	m, repo, projects, _ := newTestManager()
	return m, repo, createTestProject(t, projects, "Sistema de ventas")
}

// HU-52 · CA-52.1 · Verifica que Create le asigne al ítem un identificador público UUID versión 4 en forma canónica, que no sea el nulo, y que el repositorio guarde ese mismo valor.
func TestCreate_AsignaUnUUIDVersion4AlItem(t *testing.T) {
	m, repo, projectID := publicIDEnv(t)

	item, err := m.Create(projectID, "Login de usuarios", "", "MEDIA", validCriteria())

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	expectCanonicalV4(t, "Create", item.PublicID)
	if item.PublicID == uuid.Nil() {
		t.Error("PublicID es el UUID nulo")
	}
	if saved := savedItem(t, repo, item.ID); saved.PublicID != item.PublicID {
		t.Errorf("PublicID guardado = %s, se esperaba el devuelto por Create (%s)", saved.PublicID, item.PublicID)
	}
}

// HU-52 · CA-52.2 · Verifica que 1000 ítems creados seguidos tengan identificadores públicos distintos, todos versión 4 y no correlativos.
func TestCreate_MilItemsConUUIDDistintosYNoCorrelativos(t *testing.T) {
	const total = 1000
	m, _, projectID := publicIDEnv(t)
	ids := make([]uuid.UUID, 0, total)
	seen := make(map[uuid.UUID]bool, total)

	for i := range total {
		item, err := m.Create(projectID, fmt.Sprintf("Ítem %04d", i), "", "MEDIA", validCriteria())
		if err != nil {
			t.Fatalf("Create(%d) devolvió un error inesperado: %v", i, err)
		}
		if seen[item.PublicID] {
			t.Fatalf("Create(%d): PublicID %s repetido", i, item.PublicID)
		}
		seen[item.PublicID] = true
		expectCanonicalV4(t, fmt.Sprintf("Create(%d)", i), item.PublicID)
		ids = append(ids, item.PublicID)
	}

	sorted := slices.Clone(ids)
	slices.SortFunc(sorted, uuid.UUID.Compare)
	if slices.Equal(sorted, ids) {
		t.Error("ordenar los PublicID da el orden de creación: son correlativos")
	}
}

// HU-52 · CA-52.7 · Verifica que 50 creaciones simultáneas de ítems generen 50 identificadores públicos distintos, sin condiciones de carrera con go test -race.
func TestCreate_CreacionesSimultaneasDeItemsConUUIDDistintos(t *testing.T) {
	const attempts = 50
	m, repo, projectID := publicIDEnv(t)

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := m.Create(projectID, fmt.Sprintf("Ítem %d", i), "", "MEDIA", validCriteria()); err != nil {
				t.Errorf("Create(%d) devolvió un error inesperado: %v", i, err)
			}
		}()
	}
	wg.Wait()

	if len(repo.items) != attempts {
		t.Fatalf("se guardaron %d ítems, se esperaban %d", len(repo.items), attempts)
	}
	seen := map[uuid.UUID]bool{}
	for _, item := range repo.items {
		expectCanonicalV4(t, item.Title, item.PublicID)
		seen[item.PublicID] = true
	}
	if len(seen) != attempts {
		t.Errorf("hay %d PublicID distintos, se esperaban %d", len(seen), attempts)
	}
}

// HU-52 · CA-52.3 · Verifica que GetByPublicID devuelva el ítem con ese identificador público entre varios, y una copia que no comparte los criterios de aceptación con el guardado.
func TestGetByPublicID_DevuelveElItemYUnaCopia(t *testing.T) {
	m, _, projectID := publicIDEnv(t)
	if _, err := m.Create(projectID, "Login de usuarios", "", "MEDIA", validCriteria()); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	want, err := m.Create(projectID, "Reporte de ventas", "Mensual", "ALTA", []string{"Total por mes"})
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	got, err := m.GetByPublicID(want.PublicID)

	if err != nil {
		t.Fatalf("GetByPublicID() devolvió un error inesperado: %v", err)
	}
	if got.ID != want.ID || got.PublicID != want.PublicID || got.Title != want.Title || !slices.Equal(got.AcceptanceCriteria, want.AcceptanceCriteria) {
		t.Errorf("GetByPublicID() = %+v, se esperaba %+v", got, want)
	}
	got.AcceptanceCriteria[0] = "cambiado"
	again, err := m.GetByPublicID(want.PublicID)
	if err != nil {
		t.Fatalf("GetByPublicID() devolvió un error inesperado: %v", err)
	}
	if again.AcceptanceCriteria[0] != "Total por mes" {
		t.Errorf("modificar la copia cambió el ítem guardado: criterios = %v", again.AcceptanceCriteria)
	}
}

// HU-52 · CA-52.4 · Verifica que GetByPublicID de un UUID que no existe, o del UUID nulo, devuelva ErrItemNotFound.
func TestGetByPublicID_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
	m, _, projectID := publicIDEnv(t)
	if _, err := m.Create(projectID, "Login de usuarios", "", "MEDIA", validCriteria()); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	for _, id := range []uuid.UUID{uuid.NewV4(), uuid.Nil()} {
		got, err := m.GetByPublicID(id)
		if !errors.Is(err, ErrItemNotFound) {
			t.Errorf("GetByPublicID(%s): error = %v, se esperaba ErrItemNotFound", id, err)
		}
		if got.ID != 0 {
			t.Errorf("GetByPublicID(%s) = %+v, se esperaba un ítem vacío", id, got)
		}
	}
}

// HU-52 · CA-52.5 · Verifica que modificar, cambiar el estado y estimar un ítem conserven su identificador público.
func TestModificaciones_ConservanElUUIDDelItem(t *testing.T) {
	m, repo, projectID := publicIDEnv(t)
	item, err := m.Create(projectID, "Login de usuarios", "", "MEDIA", validCriteria())
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	expectCanonicalV4(t, "Create", item.PublicID)

	ops := []struct {
		name string
		run  func() (Item, error)
	}{
		{"Update", func() (Item, error) { return m.Update(item.ID, "Login", "Con correo", "ALTA", validCriteria(), 3) }},
		{"ChangeStatus", func() (Item, error) { return m.ChangeStatus(item.ID, "EN_CURSO") }},
		{"AssignStoryPoints", func() (Item, error) { return m.AssignStoryPoints(item.ID, 5) }},
	}
	for _, op := range ops {
		got, err := op.run()
		if err != nil {
			t.Fatalf("%s devolvió un error inesperado: %v", op.name, err)
		}
		if got.PublicID != item.PublicID {
			t.Errorf("%s: PublicID = %s, se esperaba %s", op.name, got.PublicID, item.PublicID)
		}
		if saved := savedItem(t, repo, item.ID); saved.PublicID != item.PublicID {
			t.Errorf("%s: PublicID guardado = %s, se esperaba %s", op.name, saved.PublicID, item.PublicID)
		}
	}
}
