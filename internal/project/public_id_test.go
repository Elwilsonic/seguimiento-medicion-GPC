package project_test

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
	"uuid"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-51 · Identificadores UUID, parte de proyectos (Issue #193)

// canonicalV4 es la forma canónica de un UUID versión 4, variante RFC 9562,
// en minúsculas (CA-51.1).
var canonicalV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// expectCanonicalV4 verifica que id sea un UUID versión 4 en forma canónica.
func expectCanonicalV4(t *testing.T, what string, id uuid.UUID) {
	t.Helper()
	if s := id.String(); !canonicalV4.MatchString(s) {
		t.Errorf("%s: PublicID = %q, se esperaba un UUID versión 4 en forma canónica", what, s)
	}
}

// HU-51 · CA-51.1 · Verifica que Create le asigne al proyecto un identificador público UUID versión 4, en forma canónica de 36 caracteres en minúsculas.
func TestCreate_AsignaUnUUIDVersion4(t *testing.T) {
	manager, _ := newManager()

	p, err := manager.Create("Sistema de ventas", "")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	expectCanonicalV4(t, "Create", p.PublicID)
	if got := len(p.PublicID.String()); got != 36 {
		t.Errorf("largo del PublicID = %d, se esperaba 36", got)
	}
}

// HU-51 · CA-51.1 · Verifica que el primer proyecto creado no tenga como identificador público el UUID nulo ni un número.
func TestCreate_ElPrimerProyectoNoTieneElUUIDNulo(t *testing.T) {
	manager, _ := newManager()

	p, err := manager.Create("Primero", "")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if p.PublicID == uuid.Nil() {
		t.Errorf("PublicID = %s, no se esperaba el UUID nulo", p.PublicID)
	}
	if s := p.PublicID.String(); regexp.MustCompile(`^[0-9]+$`).MatchString(s) || s == "1" {
		t.Errorf("PublicID = %q, no se esperaba un número", s)
	}
}

// HU-51 · CA-51.1 · Verifica que el repositorio guarde el identificador público que asignó Create, y que la copia guardada sea la misma.
func TestCreate_GuardaElUUIDAsignado(t *testing.T) {
	manager, repo := newManager()

	p, err := manager.Create("Sistema de ventas", "")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	saved := savedProjects(t, repo)
	if len(saved) != 1 {
		t.Fatalf("se guardaron %d proyectos, se esperaba 1", len(saved))
	}
	if saved[0].PublicID != p.PublicID {
		t.Errorf("PublicID guardado = %s, se esperaba el devuelto por Create (%s)", saved[0].PublicID, p.PublicID)
	}
	expectCanonicalV4(t, "proyecto guardado", saved[0].PublicID)
}

// HU-51 · CA-51.2 · Verifica que 1000 proyectos creados seguidos tengan identificadores públicos distintos, todos versión 4 y no correlativos.
func TestCreate_MilProyectosConUUIDDistintosYNoCorrelativos(t *testing.T) {
	const total = 1000
	manager, _ := newManager()
	ids := make([]uuid.UUID, 0, total)
	seen := make(map[uuid.UUID]bool, total)

	for i := range total {
		p, err := manager.Create(fmt.Sprintf("Proyecto %04d", i), "")
		if err != nil {
			t.Fatalf("Create(%d) devolvió un error inesperado: %v", i, err)
		}
		if seen[p.PublicID] {
			t.Fatalf("Create(%d): PublicID %s repetido", i, p.PublicID)
		}
		seen[p.PublicID] = true
		expectCanonicalV4(t, fmt.Sprintf("Create(%d)", i), p.PublicID)
		ids = append(ids, p.PublicID)
	}

	// Si fueran correlativos, ordenarlos daría el orden de creación.
	sorted := slices.Clone(ids)
	slices.SortFunc(sorted, uuid.UUID.Compare)
	if slices.Equal(sorted, ids) {
		t.Error("ordenar los PublicID da el orden de creación: son correlativos")
	}
}

// HU-51 · CA-51.7 · Verifica que 50 creaciones simultáneas generen 50 identificadores públicos distintos, sin condiciones de carrera con go test -race.
func TestCreate_CreacionesSimultaneasConUUIDDistintos(t *testing.T) {
	const attempts = 50
	manager, repo := newManager()

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := manager.Create(fmt.Sprintf("Proyecto %d", i), ""); err != nil {
				t.Errorf("Create(%d) devolvió un error inesperado: %v", i, err)
			}
		}()
	}
	wg.Wait()

	saved := savedProjects(t, repo)
	if len(saved) != attempts {
		t.Fatalf("se guardaron %d proyectos, se esperaban %d", len(saved), attempts)
	}
	seen := map[uuid.UUID]bool{}
	for _, p := range saved {
		expectCanonicalV4(t, p.Name, p.PublicID)
		seen[p.PublicID] = true
	}
	if len(seen) != attempts {
		t.Errorf("hay %d PublicID distintos, se esperaban %d", len(seen), attempts)
	}
}

// HU-51 · CA-51.3 · Verifica que GetByPublicID devuelva el proyecto con ese identificador público, también entre varios proyectos.
func TestGetByPublicID_DevuelveElProyecto(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Sistema de ventas", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	want, err := manager.Create("Inventario", "Depósito")
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	got, err := manager.GetByPublicID(want.PublicID)

	if err != nil {
		t.Fatalf("GetByPublicID() devolvió un error inesperado: %v", err)
	}
	if got != want {
		t.Errorf("GetByPublicID() = %+v, se esperaba %+v", got, want)
	}
}

// HU-51 · CA-51.3 · Verifica que GetByPublicID devuelva una copia: cambiarla no cambia el proyecto guardado.
func TestGetByPublicID_DevuelveUnaCopia(t *testing.T) {
	manager, _ := newManager()
	created, err := manager.Create("Inventario", "Depósito")
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	first, err := manager.GetByPublicID(created.PublicID)
	if err != nil {
		t.Fatalf("GetByPublicID() devolvió un error inesperado: %v", err)
	}

	first.Name = "Cambiado afuera"
	first.PublicID = uuid.Nil()

	again, err := manager.GetByPublicID(created.PublicID)
	if err != nil {
		t.Fatalf("GetByPublicID() devolvió un error inesperado: %v", err)
	}
	if again != created {
		t.Errorf("GetByPublicID() después de cambiar la copia = %+v, se esperaba %+v", again, created)
	}
}

// HU-51 · CA-51.4 · Verifica que GetByPublicID devuelva ErrProjectNotFound para un UUID que nunca se asignó y para el UUID nulo.
func TestGetByPublicID_InexistenteDevuelveErrProjectNotFound(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Inventario", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	for _, id := range []uuid.UUID{uuid.MustParse("6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f"), uuid.Nil()} {
		got, err := manager.GetByPublicID(id)
		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("GetByPublicID(%s): err = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if got != (project.Project{}) {
			t.Errorf("GetByPublicID(%s) = %+v, se esperaba un proyecto vacío", id, got)
		}
	}
}

// HU-51 · CA-51.3 · Verifica que modificar el proyecto (datos, fecha de inicio, fecha de fin o las dos fechas) conserve su identificador público.
func TestModificar_ConservaElUUID(t *testing.T) {
	start := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		op   func(m *project.Manager, id int) (project.Project, error)
	}{
		{"Update", func(m *project.Manager, id int) (project.Project, error) {
			return m.Update(id, "Inventario 2", "", "EN_CURSO")
		}},
		{"SetStartDate", func(m *project.Manager, id int) (project.Project, error) {
			return m.SetStartDate(id, start)
		}},
		{"SetEndDate", func(m *project.Manager, id int) (project.Project, error) {
			if _, err := m.SetStartDate(id, start); err != nil {
				return project.Project{}, err
			}
			return m.SetEndDate(id, end)
		}},
		{"SetDates", func(m *project.Manager, id int) (project.Project, error) {
			if _, err := m.SetStartDate(id, start); err != nil {
				return project.Project{}, err
			}
			if _, err := m.SetEndDate(id, end); err != nil {
				return project.Project{}, err
			}
			return m.SetDates(id, start.AddDate(0, 0, 1), end.AddDate(0, 0, 1))
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			manager, _ := newManager()
			created, err := manager.Create("Inventario", "")
			if err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}
			expectCanonicalV4(t, "Create", created.PublicID)

			changed, err := c.op(manager, created.ID)
			if err != nil {
				t.Fatalf("%s devolvió un error inesperado: %v", c.name, err)
			}
			if changed.PublicID != created.PublicID {
				t.Errorf("%s: PublicID = %s, se esperaba %s", c.name, changed.PublicID, created.PublicID)
			}
			found, err := manager.GetByPublicID(created.PublicID)
			if err != nil {
				t.Fatalf("GetByPublicID() después de %s devolvió un error inesperado: %v", c.name, err)
			}
			if !strings.EqualFold(found.Name, changed.Name) || found.ID != created.ID {
				t.Errorf("GetByPublicID() después de %s = %+v, se esperaba %+v", c.name, found, changed)
			}
		})
	}
}
