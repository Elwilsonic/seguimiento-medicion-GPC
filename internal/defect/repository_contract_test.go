package defect

import (
	"errors"
	"testing"
)

// HU-29 · Registrar defecto (Issue #47)
//
// Contrato de Repository.Save: toda implementación, en memoria o en
// PostgreSQL, asigna IDs correlativos desde 1, conserva todos los campos, no
// depende del ID recibido y devuelve una copia (RN-29.8, RN-29.9 y RN-29.10).

const (
	contractItemID            = 1
	contractDetectionSprintID = 1
)

// storedReader permite leer lo que quedó guardado para comprobar que Save
// devolvió una copia. Lo implementan los repositorios en sus tests.
type storedReader interface {
	stored(t *testing.T) []Defect
}

// runRepositoryContract verifica el comportamiento que comparte toda
// implementación de Repository. newRepo devuelve un repositorio vacío en cada
// llamada.
func runRepositoryContract(t *testing.T, newRepo func(t *testing.T) Repository) {
	t.Run("Save asigna IDs correlativos desde 1 y conserva todos los campos", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, contractDefect("Descripción crítica", SeverityCritical))
		second := contractSave(t, repo, contractDefect("Descripción alta", SeverityHigh))

		if first.ID != 1 || second.ID != 2 {
			t.Fatalf("IDs = %d y %d, se esperaban 1 y 2", first.ID, second.ID)
		}
		contractAssertDefect(t, "primer Save", first, contractDefectWithID(1, "Descripción crítica", SeverityCritical))
		contractAssertDefect(t, "segundo Save", second, contractDefectWithID(2, "Descripción alta", SeverityHigh))
	})

	t.Run("Save ignora el ID recibido y asigna el siguiente", func(t *testing.T) {
		repo := newRepo(t)
		received := contractDefect("Descripción alta", SeverityHigh)
		received.ID = 99

		saved := contractSave(t, repo, received)

		if saved.ID != 1 {
			t.Errorf("ID = %d, se esperaba 1", saved.ID)
		}
	})

	t.Run("Save devuelve una copia y no comparte estado con lo guardado", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractDefect("Descripción original", SeverityHigh))

		reader, ok := repo.(storedReader)
		if !ok {
			t.Fatalf("el repositorio %T no permite comprobar lo guardado", repo)
		}
		before := reader.stored(t)
		if len(before) != 1 || before[0] != saved {
			t.Fatalf("lo guardado = %+v, se esperaba %+v", before, saved)
		}

		saved.Description = "Descripción modificada"
		saved.Status = StatusResolved
		after := reader.stored(t)
		if len(after) != 1 || after[0] != before[0] {
			t.Errorf("modificar el valor devuelto cambió lo guardado: %+v", after)
		}
	})
}

// contractDefect devuelve un defecto válido con todos los campos del contrato.
func contractDefect(description string, severity Severity) Defect {
	return Defect{
		Description:        description,
		Severity:           severity,
		Status:             StatusOpen,
		ItemID:             contractItemID,
		DetectionSprintID:  contractDetectionSprintID,
		ResolutionSprintID: 0,
	}
}

func contractDefectWithID(id int, description string, severity Severity) Defect {
	defect := contractDefect(description, severity)
	defect.ID = id
	return defect
}

// contractSave guarda defect en repo y falla el test si hay un error.
func contractSave(t *testing.T, repo Repository, defect Defect) Defect {
	t.Helper()
	saved, err := repo.Save(defect)
	if err != nil {
		t.Fatalf("Save devolvió un error inesperado: %v", err)
	}
	return saved
}

// contractAssertDefect compara got con want.
func contractAssertDefect(t *testing.T, op string, got, want Defect) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %+v, se esperaba %+v", op, got, want)
	}
}

// stored devuelve una copia de los defectos guardados en memoria.
func (r *MemoryRepository) stored(t *testing.T) []Defect {
	t.Helper()
	return append([]Defect(nil), r.defects...)
}

func TestMemoryRepository_Contrato(t *testing.T) {
	runRepositoryContract(t, func(*testing.T) Repository { return NewMemoryRepository() })
}

// HU-30 · Cambiar el estado de un defecto (Issue #48)
//
// Contrato de Repository.GetByID y Repository.Update: toda implementación
// busca por ID, reemplaza el defecto completo, devuelve ErrDefectNotFound si
// no existe y devuelve copias (RN-30.3, RN-30.9 y RN-30.10).
func runRepositoryStatusContract(t *testing.T, newRepo func(t *testing.T) Repository) {
	t.Run("GetByID devuelve el defecto guardado", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractDefect("Descripción crítica", SeverityCritical))
		saved := contractSave(t, repo, contractDefect("Descripción alta", SeverityHigh))

		got, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertDefect(t, "GetByID", got, saved)
	})

	t.Run("GetByID de un ID inexistente devuelve ErrDefectNotFound", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractDefect("Descripción alta", SeverityHigh))

		for _, id := range []int{999, 0, -1} {
			if _, err := repo.GetByID(id); !errors.Is(err, ErrDefectNotFound) {
				t.Errorf("GetByID(%d): err = %v, se esperaba %v", id, err, ErrDefectNotFound)
			}
		}
	})

	t.Run("Update reemplaza el defecto y devuelve una copia", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractDefect("Descripción alta", SeverityHigh))
		changed := saved
		changed.Status = StatusResolved
		changed.ResolutionSprintID = contractDetectionSprintID

		updated, err := repo.Update(changed)
		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		contractAssertDefect(t, "Update", updated, changed)
		got, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertDefect(t, "GetByID después de Update", got, changed)
	})

	t.Run("Update vuelve el sprint de resolución a 0", func(t *testing.T) {
		repo := newRepo(t)
		resolved := contractDefect("Descripción alta", SeverityHigh)
		resolved.Status = StatusResolved
		resolved.ResolutionSprintID = contractDetectionSprintID
		saved := contractSave(t, repo, resolved)
		reopened := saved
		reopened.Status = StatusInCorrection
		reopened.ResolutionSprintID = 0

		if _, err := repo.Update(reopened); err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		got, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertDefect(t, "GetByID después de reabrir", got, reopened)
	})

	t.Run("Update de un ID inexistente devuelve ErrDefectNotFound y no guarda nada", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractDefect("Descripción alta", SeverityHigh))
		missing := saved
		missing.ID = 999
		missing.Status = StatusInCorrection

		if _, err := repo.Update(missing); !errors.Is(err, ErrDefectNotFound) {
			t.Fatalf("Update: err = %v, se esperaba %v", err, ErrDefectNotFound)
		}
		reader, ok := repo.(storedReader)
		if !ok {
			t.Fatalf("el repositorio %T no permite comprobar lo guardado", repo)
		}
		if stored := reader.stored(t); len(stored) != 1 || stored[0] != saved {
			t.Errorf("lo guardado = %+v, se esperaba solo %+v", stored, saved)
		}
	})

	t.Run("GetByID y Update devuelven copias", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractDefect("Descripción alta", SeverityHigh))

		got, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		got.Status = StatusResolved
		changed := saved
		changed.Status = StatusInCorrection
		updated, err := repo.Update(changed)
		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		updated.Description = "Descripción modificada"

		again, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertDefect(t, "GetByID después de modificar las copias", again, changed)
	})
}

func TestMemoryRepository_ContratoGetByIDYUpdate(t *testing.T) {
	runRepositoryStatusContract(t, func(*testing.T) Repository { return NewMemoryRepository() })
}
