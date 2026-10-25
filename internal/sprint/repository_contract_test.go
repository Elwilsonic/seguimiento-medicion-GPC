package sprint

import (
	"errors"
	"testing"
	"time"
)

// TEC-05 · Persistencia con PostgreSQL en Docker (Issue #35)

// Proyectos que usan los tests de contrato. En PostgreSQL existen porque
// la función que arma el repositorio los crea antes (clave foránea).
const (
	contractProjectA = 1
	contractProjectB = 2
)

// Fechas fijas en UTC y sin fracciones de segundo, iguales en memoria y en
// PostgreSQL.
var (
	contractStart = time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	contractEnd   = time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
)

// runRepositoryContract verifica el comportamiento que comparte toda
// implementación de Repository, en memoria o en PostgreSQL (CA-TEC05.3).
// newRepo devuelve un repositorio vacío en cada llamada.
func runRepositoryContract(t *testing.T, newRepo func(t *testing.T) Repository) {
	t.Run("Save asigna IDs correlativos y globales y conserva los campos", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSprint(contractProjectA, "Sprint 1", StatusOpen)
		second := contractSprint(contractProjectB, "Sprint 1", StatusClosed)

		savedFirst := contractSave(t, repo, first)
		savedSecond := contractSave(t, repo, second)

		first.ID, second.ID = 1, 2
		contractAssertSprint(t, "primer Save", savedFirst, first)
		contractAssertSprint(t, "segundo Save", savedSecond, second)
	})

	t.Run("OpenByProject sin sprints no encuentra ninguno", func(t *testing.T) {
		repo := newRepo(t)

		got, exists, err := repo.OpenByProject(contractProjectA)
		if err != nil {
			t.Fatalf("OpenByProject devolvió un error inesperado: %v", err)
		}
		if exists || got != (Sprint{}) {
			t.Errorf("OpenByProject = %+v, %v; se esperaba un sprint vacío y false", got, exists)
		}
	})

	// HU-11: GetByID es la operación de lectura que usa AssignItem (RN-11.1).
	t.Run("GetByID devuelve el sprint guardado", func(t *testing.T) {
		repo := newRepo(t)
		want := contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusOpen))

		got, err := repo.GetByID(want.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertSprint(t, "GetByID", got, want)
	})

	// HU-11: la validación de la existencia del sprint (RN-11.1).
	t.Run("GetByID de un ID inexistente da ErrSprintNotFound", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusOpen))

		_, err := repo.GetByID(99)
		if !errors.Is(err, ErrSprintNotFound) {
			t.Errorf("err = %v, se esperaba %v", err, ErrSprintNotFound)
		}
	})

	// HU-13: Update persiste el sprint y devuelve una copia (RN-13.4).
	t.Run("Update actualiza el sprint existente y devuelve una copia", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusOpen))

		updated := saved
		updated.Name = "Sprint cerrado"
		updated.Goal = "Cerrar la iteración"
		updated.Status = StatusClosed
		updated.StartDate = contractEnd
		updated.EndDate = contractEnd.AddDate(0, 0, 7)

		got, err := repo.Update(updated)
		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		contractAssertSprint(t, "Update", got, updated)

		got.Name = "Mutado"
		stored, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID después de Update devolvió un error inesperado: %v", err)
		}
		contractAssertSprint(t, "GetByID después de Update", stored, updated)
	})

	// HU-13: el ID identifica el sprint y no se altera; los demás no cambian.
	t.Run("Update conserva el ID y no altera otros sprints", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusOpen))
		second := contractSave(t, repo, contractSprint(contractProjectB, "Sprint 2", StatusOpen))

		updated := first
		updated.Name = "Sprint 1 actualizado"
		updated.Status = StatusClosed

		got, err := repo.Update(updated)
		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		if got.ID != first.ID {
			t.Errorf("ID devuelto = %d, se esperaba %d", got.ID, first.ID)
		}

		storedSecond, err := repo.GetByID(second.ID)
		if err != nil {
			t.Fatalf("GetByID del segundo sprint devolvió un error inesperado: %v", err)
		}
		contractAssertSprint(t, "segundo sprint después de Update", storedSecond, second)
	})

	// HU-13: Update de un ID inexistente no crea ni modifica nada.
	t.Run("Update de un ID inexistente da ErrSprintNotFound", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusOpen))

		missing := contractSprint(contractProjectA, "Inexistente", StatusOpen)
		missing.ID = 99

		_, err := repo.Update(missing)
		if !errors.Is(err, ErrSprintNotFound) {
			t.Errorf("err = %v, se esperaba %v", err, ErrSprintNotFound)
		}
	})

	t.Run("OpenByProject devuelve el sprint ABIERTO del proyecto", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusClosed))
		want := contractSave(t, repo, contractSprint(contractProjectA, "Sprint 2", StatusOpen))

		got, exists, err := repo.OpenByProject(contractProjectA)
		if err != nil {
			t.Fatalf("OpenByProject devolvió un error inesperado: %v", err)
		}
		if !exists {
			t.Fatal("OpenByProject no encontró el sprint ABIERTO")
		}
		contractAssertSprint(t, "OpenByProject", got, want)
	})

	t.Run("OpenByProject no ve el sprint ABIERTO de otro proyecto", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractSprint(contractProjectB, "Sprint 1", StatusOpen))

		_, exists, err := repo.OpenByProject(contractProjectA)
		if err != nil {
			t.Fatalf("OpenByProject devolvió un error inesperado: %v", err)
		}
		if exists {
			t.Error("OpenByProject encontró el sprint ABIERTO de otro proyecto")
		}
	})

	t.Run("OpenByProject no ve un sprint CERRADO", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusClosed))

		_, exists, err := repo.OpenByProject(contractProjectA)
		if err != nil {
			t.Fatalf("OpenByProject devolvió un error inesperado: %v", err)
		}
		if exists {
			t.Error("OpenByProject encontró un sprint CERRADO")
		}
	})

	// HU-32 (CA-32.8): ListByProject lee los sprints de un proyecto, para
	// encontrar los sprints cerrados de la velocidad.
	t.Run("ListByProject sin sprints devuelve una lista vacía", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractSprint(contractProjectB, "Sprint 1", StatusOpen))

		got, err := repo.ListByProject(contractProjectA)
		if err != nil {
			t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("ListByProject = %#v, se esperaba una lista vacía, no nil", got)
		}
	})

	t.Run("ListByProject devuelve los sprints del proyecto ordenados por ID", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusClosed))
		contractSave(t, repo, contractSprint(contractProjectB, "Sprint 1", StatusOpen))
		third := contractSave(t, repo, contractSprint(contractProjectA, "Sprint 2", StatusOpen))

		got, err := repo.ListByProject(contractProjectA)
		if err != nil {
			t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("ListByProject devolvió %d sprints, se esperaban 2: %+v", len(got), got)
		}
		contractAssertSprint(t, "primer sprint", got[0], first)
		contractAssertSprint(t, "segundo sprint", got[1], third)
	})

	t.Run("ListByProject devuelve una lista nueva", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractSprint(contractProjectA, "Sprint 1", StatusOpen))

		got, err := repo.ListByProject(contractProjectA)
		if err != nil || len(got) != 1 {
			t.Fatalf("ListByProject = %+v, %v; se esperaba un sprint", got, err)
		}
		got[0].Name = "Cambiado"

		again, err := repo.ListByProject(contractProjectA)
		if err != nil || len(again) != 1 {
			t.Fatalf("ListByProject = %+v, %v; se esperaba un sprint", again, err)
		}
		contractAssertSprint(t, "sprint guardado", again[0], saved)
	})
}

// contractSprint devuelve un sprint válido del proyecto, con las fechas fijas.
func contractSprint(projectID int, name string, status Status) Sprint {
	return Sprint{
		ProjectID: projectID,
		Name:      name,
		StartDate: contractStart,
		EndDate:   contractEnd,
		Goal:      "Goal de " + name,
		Status:    status,
	}
}

// contractSave guarda sprint en repo y falla el test si hay un error.
func contractSave(t *testing.T, repo Repository, sprint Sprint) Sprint {
	t.Helper()
	saved, err := repo.Save(sprint)
	if err != nil {
		t.Fatalf("Save devolvió un error inesperado: %v", err)
	}
	return saved
}

// contractAssertSprint compara got con want; las fechas se comparan con
// Equal (el mismo instante), no con ==.
func contractAssertSprint(t *testing.T, op string, got, want Sprint) {
	t.Helper()
	if !got.StartDate.Equal(want.StartDate) || !got.EndDate.Equal(want.EndDate) {
		t.Errorf("%s: fechas = %v a %v, se esperaba %v a %v", op, got.StartDate, got.EndDate, want.StartDate, want.EndDate)
	}
	got.StartDate, got.EndDate = want.StartDate, want.EndDate
	if got != want {
		t.Errorf("%s = %+v, se esperaba %+v", op, got, want)
	}
}

// Los tests de contrato de memoria caracterizan el comportamiento actual:
// pasan sin cambios en MemoryRepository.
func TestMemoryRepository_Contrato(t *testing.T) {
	runRepositoryContract(t, func(*testing.T) Repository { return NewMemoryRepository() })
}
