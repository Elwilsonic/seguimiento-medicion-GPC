package project

import (
	"errors"
	"slices"
	"testing"
	"time"
)

// TEC-05 · Persistencia con PostgreSQL en Docker (Issue #35)

// runRepositoryContract verifica el comportamiento que comparte toda
// implementación de Repository, en memoria o en PostgreSQL (CA-TEC05.3).
// newRepo devuelve un repositorio vacío en cada llamada.
func runRepositoryContract(t *testing.T, newRepo func(t *testing.T) Repository) {
	t.Run("Save asigna IDs correlativos desde 1 y devuelve los campos", func(t *testing.T) {
		repo := newRepo(t)

		first := contractSave(t, repo, Project{Name: "Alfa", Description: "Primero", Status: StatusPlanned})
		second := contractSave(t, repo, Project{Name: "Beta", Status: StatusInProgress})

		if want := (Project{ID: 1, Name: "Alfa", Description: "Primero", Status: StatusPlanned}); first != want {
			t.Errorf("primer Save = %+v, se esperaba %+v", first, want)
		}
		if want := (Project{ID: 2, Name: "Beta", Status: StatusInProgress}); second != want {
			t.Errorf("segundo Save = %+v, se esperaba %+v", second, want)
		}
	})

	t.Run("List de un repositorio vacío no tiene proyectos", func(t *testing.T) {
		repo := newRepo(t)

		got, err := repo.List()
		if err != nil {
			t.Fatalf("List devolvió un error inesperado: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("List = %+v, se esperaba una lista vacía", got)
		}
	})

	t.Run("List devuelve los proyectos en orden de ID", func(t *testing.T) {
		repo := newRepo(t)
		want := []Project{
			contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned}),
			contractSave(t, repo, Project{Name: "Beta", Status: StatusInProgress}),
			contractSave(t, repo, Project{Name: "Gamma", Status: StatusFinished}),
		}

		got, err := repo.List()
		if err != nil {
			t.Fatalf("List devolvió un error inesperado: %v", err)
		}
		if !slices.Equal(got, want) {
			t.Errorf("List = %+v, se esperaba %+v", got, want)
		}
	})

	// En PostgreSQL un UPDATE deja la fila al final de la tabla: sin ORDER BY,
	// el proyecto modificado saldría último.
	t.Run("List sigue en orden de ID después de modificar el primero", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned})
		second := contractSave(t, repo, Project{Name: "Beta", Status: StatusPlanned})
		third := contractSave(t, repo, Project{Name: "Gamma", Status: StatusPlanned})
		first.Status = StatusInProgress
		if _, err := repo.Update(first); err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}

		got, err := repo.List()
		if err != nil {
			t.Fatalf("List devolvió un error inesperado: %v", err)
		}
		if want := []Project{first, second, third}; !slices.Equal(got, want) {
			t.Errorf("List = %+v, se esperaba %+v", got, want)
		}
	})

	t.Run("GetByID devuelve el proyecto guardado", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned})
		want := contractSave(t, repo, Project{Name: "Beta", Description: "Segundo", Status: StatusInProgress})

		got, err := repo.GetByID(want.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		if got != want {
			t.Errorf("GetByID = %+v, se esperaba %+v", got, want)
		}
	})

	t.Run("GetByID de un ID inexistente da ErrProjectNotFound", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned})

		got, err := repo.GetByID(99)
		if !errors.Is(err, ErrProjectNotFound) {
			t.Errorf("err = %v, se esperaba %v", err, ErrProjectNotFound)
		}
		if got != (Project{}) {
			t.Errorf("GetByID = %+v, se esperaba un proyecto vacío", got)
		}
	})

	t.Run("Update reemplaza el proyecto guardado", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, Project{Name: "Alfa", Description: "Antes", Status: StatusPlanned})
		changed := Project{ID: saved.ID, Name: "Alfa 2", Description: "Después", Status: StatusFinished}

		got, err := repo.Update(changed)
		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		if got != changed {
			t.Errorf("Update = %+v, se esperaba %+v", got, changed)
		}
		stored, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		if stored != changed {
			t.Errorf("GetByID después de Update = %+v, se esperaba %+v", stored, changed)
		}
	})

	t.Run("Update de un ID inexistente da ErrProjectNotFound", func(t *testing.T) {
		repo := newRepo(t)

		_, err := repo.Update(Project{ID: 99, Name: "Nadie", Status: StatusPlanned})
		if !errors.Is(err, ErrProjectNotFound) {
			t.Errorf("err = %v, se esperaba %v", err, ErrProjectNotFound)
		}
	})

	// HU-18, CA-18.9: la fecha de inicio se conserva al guardar, sin
	// registrar (cero) o registrada.
	t.Run("Save y GetByID conservan la fecha de inicio", func(t *testing.T) {
		repo := newRepo(t)
		withoutDate := contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned})
		withDate := contractSave(t, repo, Project{Name: "Beta", Status: StatusPlanned, StartDate: contractStartDate})

		if !withoutDate.StartDate.IsZero() {
			t.Errorf("Save sin fecha: StartDate = %v, se esperaba sin registrar", withoutDate.StartDate)
		}
		if !withDate.StartDate.Equal(contractStartDate) {
			t.Errorf("Save con fecha: StartDate = %v, se esperaba %v", withDate.StartDate, contractStartDate)
		}
		for _, want := range []Project{withoutDate, withDate} {
			got, err := repo.GetByID(want.ID)
			if err != nil {
				t.Fatalf("GetByID devolvió un error inesperado: %v", err)
			}
			if got != want {
				t.Errorf("GetByID = %+v, se esperaba %+v", got, want)
			}
		}
	})

	// HU-18, CA-18.5 y CA-18.9: Update escribe la fecha de inicio recibida y
	// la conserva en las lecturas.
	t.Run("Update escribe y conserva la fecha de inicio", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, Project{Name: "Alfa", Status: StatusPlanned})
		withDate := saved
		withDate.StartDate = contractStartDate

		got, err := repo.Update(withDate)
		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		if got != withDate {
			t.Errorf("Update = %+v, se esperaba %+v", got, withDate)
		}
		stored, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		if stored != withDate {
			t.Errorf("GetByID después de Update = %+v, se esperaba %+v", stored, withDate)
		}
		listed, err := repo.List()
		if err != nil {
			t.Fatalf("List devolvió un error inesperado: %v", err)
		}
		if want := []Project{withDate}; !slices.Equal(listed, want) {
			t.Errorf("List = %+v, se esperaba %+v", listed, want)
		}
	})
}

// contractStartDate es una fecha de inicio como la lee la API: 00:00 UTC.
var contractStartDate = time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)

// contractSave guarda p en repo y falla el test si hay un error.
func contractSave(t *testing.T, repo Repository, p Project) Project {
	t.Helper()
	saved, err := repo.Save(p)
	if err != nil {
		t.Fatalf("Save devolvió un error inesperado: %v", err)
	}
	return saved
}

// Los tests de contrato de memoria caracterizan el comportamiento actual:
// pasan sin cambios en MemoryRepository.
func TestMemoryRepository_Contrato(t *testing.T) {
	runRepositoryContract(t, func(*testing.T) Repository { return NewMemoryRepository() })
}
