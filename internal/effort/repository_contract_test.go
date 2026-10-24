package effort

import (
	"testing"
	"time"
)

// HU-26 · Registrar esfuerzo (Issue #44)
//
// Contrato de Repository.Save: toda implementación, en memoria o en
// PostgreSQL, asigna IDs correlativos desde 1, conserva los campos y no
// depende del ID recibido (RN-26.8 y RN-26.9).

const (
	contractItemID   = 1
	contractSprintID = 1
)

var contractDate = time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)

// runRepositoryContract verifica el comportamiento que comparte toda
// implementación de Repository. newRepo devuelve un repositorio vacío en cada
// llamada.
func runRepositoryContract(t *testing.T, newRepo func(t *testing.T) Repository) {
	t.Run("Save asigna IDs correlativos desde 1 y conserva los campos", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, contractEffort("García Santiago", "Implementar el registro", 2.5))
		second := contractSave(t, repo, contractEffort("Buttini", "Revisar el registro", 1))

		if first.ID != 1 || second.ID != 2 {
			t.Fatalf("IDs = %d y %d, se esperaban 1 y 2", first.ID, second.ID)
		}
		contractAssertEffort(t, "primer Save", first, contractEffortWithID(1, "García Santiago", "Implementar el registro", 2.5))
		contractAssertEffort(t, "segundo Save", second, contractEffortWithID(2, "Buttini", "Revisar el registro", 1))
	})

	t.Run("Save ignora el ID recibido y asigna el siguiente", func(t *testing.T) {
		repo := newRepo(t)
		received := contractEffort("García Santiago", "Implementar", 2)
		received.ID = 99

		saved := contractSave(t, repo, received)

		if saved.ID != 1 {
			t.Errorf("ID = %d, se esperaba 1", saved.ID)
		}
	})

	t.Run("Save acepta horas fraccionadas", func(t *testing.T) {
		repo := newRepo(t)

		saved := contractSave(t, repo, contractEffort("García Santiago", "Ajustar", 0.001))

		if saved.Hours != 0.001 {
			t.Errorf("horas = %v, se esperaba 0.001", saved.Hours)
		}
	})
}

// contractEffort devuelve un esfuerzo válido con la fecha fija del contrato.
func contractEffort(member, activity string, hours float64) Effort {
	return Effort{
		Member:   member,
		Date:     contractDate,
		Activity: activity,
		ItemID:   contractItemID,
		SprintID: contractSprintID,
		Hours:    hours,
	}
}

func contractEffortWithID(id int, member, activity string, hours float64) Effort {
	effort := contractEffort(member, activity, hours)
	effort.ID = id
	return effort
}

// contractSave guarda effort en repo y falla el test si hay un error.
func contractSave(t *testing.T, repo Repository, effort Effort) Effort {
	t.Helper()
	saved, err := repo.Save(effort)
	if err != nil {
		t.Fatalf("Save devolvió un error inesperado: %v", err)
	}
	return saved
}

// contractAssertEffort compara got con want; la fecha se compara con Equal
// (el mismo instante), no con ==.
func contractAssertEffort(t *testing.T, op string, got, want Effort) {
	t.Helper()
	if !got.Date.Equal(want.Date) {
		t.Errorf("%s: fecha = %v, se esperaba %v", op, got.Date, want.Date)
	}
	got.Date, want.Date = time.Time{}, time.Time{}
	if got != want {
		t.Errorf("%s = %+v, se esperaba %+v", op, got, want)
	}
}

// Los tests de contrato de memoria caracterizan el comportamiento actual:
// pasan sin cambios en MemoryRepository.
func TestMemoryRepository_Contrato(t *testing.T) {
	runRepositoryContract(t, func(*testing.T) Repository { return NewMemoryRepository() })
}
