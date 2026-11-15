package sprint

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

// HU-52 · CA-52.1 · Verifica que Create le asigne al sprint un identificador público UUID versión 4 en forma canónica, que no sea el nulo, y que el repositorio guarde ese mismo valor.
func TestCreate_AsignaUnUUIDVersion4AlSprint(t *testing.T) {
	m, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	sp, err := m.Create(projectID, "Sprint 1", validStart, validEnd, "Entregar el MVP")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	expectCanonicalV4(t, "Create", sp.PublicID)
	if sp.PublicID == uuid.Nil() {
		t.Error("PublicID es el UUID nulo")
	}
	if saved := savedSprint(t, repo, sp.ID); saved.PublicID != sp.PublicID {
		t.Errorf("PublicID guardado = %s, se esperaba el devuelto por Create (%s)", saved.PublicID, sp.PublicID)
	}
}

// HU-52 · CA-52.2 · Verifica que 1000 sprints creados seguidos (cada uno cerrado antes del siguiente) tengan identificadores públicos distintos, todos versión 4 y no correlativos.
func TestCreate_MilSprintsConUUIDDistintosYNoCorrelativos(t *testing.T) {
	const total = 1000
	m, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	ids := make([]uuid.UUID, 0, total)
	seen := make(map[uuid.UUID]bool, total)

	for i := range total {
		sp, err := m.Create(projectID, fmt.Sprintf("Sprint %04d", i), validStart, validEnd, "Objetivo")
		if err != nil {
			t.Fatalf("Create(%d) devolvió un error inesperado: %v", i, err)
		}
		if _, err := m.Close(sp.ID); err != nil {
			t.Fatalf("Close(%d) devolvió un error inesperado: %v", i, err)
		}
		if seen[sp.PublicID] {
			t.Fatalf("Create(%d): PublicID %s repetido", i, sp.PublicID)
		}
		seen[sp.PublicID] = true
		expectCanonicalV4(t, fmt.Sprintf("Create(%d)", i), sp.PublicID)
		ids = append(ids, sp.PublicID)
	}

	sorted := slices.Clone(ids)
	slices.SortFunc(sorted, uuid.UUID.Compare)
	if slices.Equal(sorted, ids) {
		t.Error("ordenar los PublicID da el orden de creación: son correlativos")
	}
}

// HU-52 · CA-52.7 · Verifica que 50 creaciones simultáneas de sprints, en proyectos distintos, generen 50 identificadores públicos distintos, sin condiciones de carrera con go test -race.
func TestCreate_CreacionesSimultaneasDeSprintsConUUIDDistintos(t *testing.T) {
	const attempts = 50
	m, repo, projects, _ := newTestManager()
	projectIDs := make([]int, attempts)
	for i := range attempts {
		projectIDs[i] = createTestProject(t, projects, fmt.Sprintf("Proyecto %d", i))
	}

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := m.Create(projectIDs[i], "Sprint 1", validStart, validEnd, "Objetivo"); err != nil {
				t.Errorf("Create(%d) devolvió un error inesperado: %v", i, err)
			}
		}()
	}
	wg.Wait()

	if len(repo.sprints) != attempts {
		t.Fatalf("se guardaron %d sprints, se esperaban %d", len(repo.sprints), attempts)
	}
	seen := map[uuid.UUID]bool{}
	for _, sp := range repo.sprints {
		expectCanonicalV4(t, fmt.Sprintf("sprint %d", sp.ID), sp.PublicID)
		seen[sp.PublicID] = true
	}
	if len(seen) != attempts {
		t.Errorf("hay %d PublicID distintos, se esperaban %d", len(seen), attempts)
	}
}

// HU-52 · CA-52.3 · Verifica que GetByPublicID devuelva el sprint con ese identificador público entre varios.
func TestGetByPublicID_DevuelveElSprint(t *testing.T) {
	m, _, projects, _ := newTestManager()
	first := createTestProject(t, projects, "Sistema de ventas")
	second := createTestProject(t, projects, "Inventario")
	if _, err := m.Create(first, "Sprint 1", validStart, validEnd, "Objetivo"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	want, err := m.Create(second, "Sprint A", validStart, validEnd, "Otro objetivo")
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	got, err := m.GetByPublicID(want.PublicID)

	if err != nil {
		t.Fatalf("GetByPublicID() devolvió un error inesperado: %v", err)
	}
	if got != want {
		t.Errorf("GetByPublicID() = %+v, se esperaba %+v", got, want)
	}
}

// HU-52 · CA-52.4 · Verifica que GetByPublicID de un UUID que no existe, o del UUID nulo, devuelva ErrSprintNotFound.
func TestGetByPublicID_SprintInexistenteDevuelveErrSprintNotFound(t *testing.T) {
	m, _, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	if _, err := m.Create(projectID, "Sprint 1", validStart, validEnd, "Objetivo"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	for _, id := range []uuid.UUID{uuid.NewV4(), uuid.Nil()} {
		got, err := m.GetByPublicID(id)
		if !errors.Is(err, ErrSprintNotFound) {
			t.Errorf("GetByPublicID(%s): error = %v, se esperaba ErrSprintNotFound", id, err)
		}
		if got != (Sprint{}) {
			t.Errorf("GetByPublicID(%s) = %+v, se esperaba un sprint vacío", id, got)
		}
	}
}

// HU-52 · CA-52.5 · Verifica que cerrar un sprint conserve su identificador público.
func TestClose_ConservaElUUIDDelSprint(t *testing.T) {
	m, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")
	sp, err := m.Create(projectID, "Sprint 1", validStart, validEnd, "Objetivo")
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	expectCanonicalV4(t, "Create", sp.PublicID)

	closed, err := m.Close(sp.ID)

	if err != nil {
		t.Fatalf("Close() devolvió un error inesperado: %v", err)
	}
	if closed.PublicID != sp.PublicID {
		t.Errorf("Close: PublicID = %s, se esperaba %s", closed.PublicID, sp.PublicID)
	}
	if saved := savedSprint(t, repo, sp.ID); saved.PublicID != sp.PublicID {
		t.Errorf("PublicID guardado = %s, se esperaba %s", saved.PublicID, sp.PublicID)
	}
}
