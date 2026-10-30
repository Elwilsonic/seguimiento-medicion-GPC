package effort

import (
	"errors"
	"testing"
	"time"
)

// HU-26 · Registrar esfuerzo (Issue #44)
//
// Contrato de Repository.Save: toda implementación, en memoria o en
// PostgreSQL, asigna IDs correlativos desde 1, conserva los campos y no
// depende del ID recibido (RN-26.8 y RN-26.9).

const (
	contractItemID        = 1
	contractSprintID      = 1
	contractOtherItemID   = 2
	contractOtherSprintID = 2
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

	t.Run("ListByItem devuelve una lista vacía, no nil", func(t *testing.T) {
		repo := newRepo(t)

		got, err := repo.ListByItem(contractItemID)

		if err != nil {
			t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
		}
		if got == nil {
			t.Fatal("la lista es nil, se esperaba una lista vacía")
		}
		if len(got) != 0 {
			t.Errorf("esfuerzos = %d, se esperaban 0", len(got))
		}
	})

	t.Run("ListBySprint devuelve una lista vacía, no nil", func(t *testing.T) {
		repo := newRepo(t)

		got, err := repo.ListBySprint(contractSprintID)

		if err != nil {
			t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
		}
		if got == nil {
			t.Fatal("la lista es nil, se esperaba una lista vacía")
		}
		if len(got) != 0 {
			t.Errorf("esfuerzos = %d, se esperaban 0", len(got))
		}
	})

	t.Run("ListByItem devuelve los esfuerzos del ítem ordenados por ID", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, contractEffort("García Santiago", "Implementar", 2.5))
		contractSave(t, repo, contractEffortFor(contractOtherItemID, contractSprintID, "Buttini", "Otra historia", 4))
		third := contractSave(t, repo, contractEffort("García Santiago", "Revisar", 1))

		got, err := repo.ListByItem(contractItemID)

		if err != nil {
			t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("esfuerzos = %d, se esperaban 2", len(got))
		}
		if got[0].ID != first.ID || got[1].ID != third.ID {
			t.Errorf("IDs = %d y %d, se esperaban %d y %d", got[0].ID, got[1].ID, first.ID, third.ID)
		}
	})

	t.Run("ListBySprint devuelve los esfuerzos del sprint ordenados por ID", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, contractEffort("García Santiago", "Implementar", 2.5))
		contractSave(t, repo, contractEffortFor(contractItemID, contractOtherSprintID, "Buttini", "Otro sprint", 4))
		third := contractSave(t, repo, contractEffort("García Santiago", "Revisar", 1))

		got, err := repo.ListBySprint(contractSprintID)

		if err != nil {
			t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("esfuerzos = %d, se esperaban 2", len(got))
		}
		if got[0].ID != first.ID || got[1].ID != third.ID {
			t.Errorf("IDs = %d y %d, se esperaban %d y %d", got[0].ID, got[1].ID, first.ID, third.ID)
		}
	})

	t.Run("ListByItem aísla los esfuerzos de otra historia", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractEffort("García Santiago", "Implementar", 2.5))
		other := contractSave(t, repo, contractEffortFor(contractOtherItemID, contractSprintID, "Buttini", "Otra historia", 4))

		got, err := repo.ListByItem(contractOtherItemID)

		if err != nil {
			t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
		}
		if len(got) != 1 || got[0].ID != other.ID {
			t.Errorf("esfuerzos = %+v, se esperaba solo el esfuerzo %d", got, other.ID)
		}
	})

	t.Run("ListBySprint aísla los esfuerzos de otro sprint", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractEffort("García Santiago", "Implementar", 2.5))
		other := contractSave(t, repo, contractEffortFor(contractItemID, contractOtherSprintID, "Buttini", "Otro sprint", 4))

		got, err := repo.ListBySprint(contractOtherSprintID)

		if err != nil {
			t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
		}
		if len(got) != 1 || got[0].ID != other.ID {
			t.Errorf("esfuerzos = %+v, se esperaba solo el esfuerzo %d", got, other.ID)
		}
	})

	t.Run("Las listas devuelven copias", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractEffort("García Santiago", "Implementar", 2.5))

		got, err := repo.ListByItem(contractItemID)
		if err != nil {
			t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esfuerzos = %d, se esperaba 1", len(got))
		}
		got[0].Member = "Otra persona"
		got[0].Hours = 99

		again, err := repo.ListByItem(contractItemID)
		if err != nil {
			t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
		}
		if len(again) != 1 {
			t.Fatalf("esfuerzos = %d, se esperaba 1", len(again))
		}
		if again[0].Member != "García Santiago" || again[0].Hours != 2.5 {
			t.Errorf("el esfuerzo guardado cambió: %+v", again[0])
		}
	})

	// HU-28 · Corregir un registro de esfuerzo (Issue #46): GetByID y Update
	// cumplen el mismo contrato en memoria y en PostgreSQL (RN-28.11).
	t.Run("GetByID devuelve el esfuerzo guardado", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractEffort("García Santiago", "Implementar", 2.5))
		second := contractSave(t, repo, contractEffort("Buttini", "Revisar", 1))

		got, err := repo.GetByID(second.ID)

		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertEffort(t, "GetByID", got, second)
	})

	t.Run("GetByID de un esfuerzo inexistente devuelve ErrEffortNotFound", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractEffort("García Santiago", "Implementar", 2.5))

		for _, id := range []int{99, 0, -1} {
			got, err := repo.GetByID(id)

			if !errors.Is(err, ErrEffortNotFound) {
				t.Errorf("GetByID(%d): err = %v, se esperaba %v", id, err, ErrEffortNotFound)
			}
			if got != (Effort{}) {
				t.Errorf("GetByID(%d) = %+v, se esperaba el valor cero", id, got)
			}
		}
	})

	t.Run("Update reemplaza el esfuerzo y lo devuelve", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractEffort("García Santiago", "Implementar", 12))
		other := contractSave(t, repo, contractEffort("Buttini", "Revisar", 1))
		changed := saved
		changed.Hours = 10

		got, err := repo.Update(changed)

		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		contractAssertEffort(t, "Update", got, changed)
		stored, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertEffort(t, "esfuerzo guardado", stored, changed)
		untouched, err := repo.GetByID(other.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertEffort(t, "otro esfuerzo", untouched, other)
	})

	t.Run("Update de un esfuerzo inexistente devuelve ErrEffortNotFound sin crear nada", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractEffort("García Santiago", "Implementar", 2.5))
		missing := saved
		missing.ID = 99

		got, err := repo.Update(missing)

		if !errors.Is(err, ErrEffortNotFound) {
			t.Errorf("err = %v, se esperaba %v", err, ErrEffortNotFound)
		}
		if got != (Effort{}) {
			t.Errorf("Update = %+v, se esperaba el valor cero", got)
		}
		list, err := repo.ListBySprint(contractSprintID)
		if err != nil {
			t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
		}
		if len(list) != 1 || list[0].ID != saved.ID {
			t.Errorf("esfuerzos = %+v, se esperaba solo el esfuerzo %d", list, saved.ID)
		}
	})

	t.Run("GetByID y Update devuelven copias", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractEffort("García Santiago", "Implementar", 2.5))

		got, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		got.Hours = 99
		updated, err := repo.Update(saved)
		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		updated.Hours = 77

		again, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		if again.Hours != 2.5 {
			t.Errorf("el esfuerzo guardado cambió: %+v", again)
		}
	})
}

// contractEffort devuelve un esfuerzo válido con la fecha fija del contrato.
func contractEffort(member, activity string, hours float64) Effort {
	return contractEffortFor(contractItemID, contractSprintID, member, activity, hours)
}

func contractEffortFor(itemID, sprintID int, member, activity string, hours float64) Effort {
	return Effort{
		Member:   member,
		Date:     contractDate,
		Activity: activity,
		ItemID:   itemID,
		SprintID: sprintID,
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
