package sprint

import (
	"errors"
	"testing"
	"uuid"

	"seguimiento-medicion-gpc/internal/postgrestest"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197)

// contractPublicIDA y contractPublicIDB son identificadores públicos fijos
// para el contrato de GetByPublicID.
var (
	contractPublicIDA = uuid.MustParse("7c1e4b2a-9d3f-4e8a-b6c5-1a2b3c4d5e6f")
	contractPublicIDB = uuid.MustParse("e2d4f6a8-0b1c-4d3e-9f5a-6b7c8d9e0f1a")
)

// runRepositoryPublicIDContract verifica que toda implementación de
// Repository guarde el identificador público que recibe, lo busque con
// GetByPublicID y lo conserve al actualizar el sprint (CA-52.5).
func runRepositoryPublicIDContract(t *testing.T, newRepo func(t *testing.T) Repository) {
	t.Run("Save conserva el PublicID y GetByPublicID devuelve el sprint", func(t *testing.T) {
		repo := newRepo(t)
		a := contractSprint(contractProjectA, "Sprint 1", StatusOpen)
		a.PublicID = contractPublicIDA
		b := contractSprint(contractProjectB, "Sprint 2", StatusOpen)
		b.PublicID = contractPublicIDB
		savedA, err := repo.Save(a)
		if err != nil {
			t.Fatalf("Save devolvió un error inesperado: %v", err)
		}
		savedB, err := repo.Save(b)
		if err != nil {
			t.Fatalf("Save devolvió un error inesperado: %v", err)
		}
		if savedA.PublicID != contractPublicIDA || savedB.PublicID != contractPublicIDB {
			t.Errorf("Save: PublicID = %s y %s, se esperaba %s y %s", savedA.PublicID, savedB.PublicID, contractPublicIDA, contractPublicIDB)
		}

		for _, want := range []Sprint{savedA, savedB} {
			got, err := repo.GetByPublicID(want.PublicID)
			if err != nil {
				t.Fatalf("GetByPublicID(%s) devolvió un error inesperado: %v", want.PublicID, err)
			}
			if got.ID != want.ID || got.PublicID != want.PublicID {
				t.Errorf("GetByPublicID(%s) = %+v, se esperaba %+v", want.PublicID, got, want)
			}
			byID, err := repo.GetByID(want.ID)
			if err != nil {
				t.Fatalf("GetByID(%d) devolvió un error inesperado: %v", want.ID, err)
			}
			if byID.PublicID != want.PublicID {
				t.Errorf("GetByID(%d): PublicID = %s, se esperaba %s", want.ID, byID.PublicID, want.PublicID)
			}
		}
	})

	t.Run("GetByPublicID de un UUID inexistente da ErrSprintNotFound", func(t *testing.T) {
		repo := newRepo(t)
		a := contractSprint(contractProjectA, "Sprint 1", StatusOpen)
		a.PublicID = contractPublicIDA
		if _, err := repo.Save(a); err != nil {
			t.Fatalf("Save devolvió un error inesperado: %v", err)
		}

		for _, id := range []uuid.UUID{contractPublicIDB, uuid.Nil()} {
			got, err := repo.GetByPublicID(id)
			if !errors.Is(err, ErrSprintNotFound) {
				t.Errorf("GetByPublicID(%s): error = %v, se esperaba ErrSprintNotFound", id, err)
			}
			if got.ID != 0 {
				t.Errorf("GetByPublicID(%s) = %+v, se esperaba un valor vacío", id, got)
			}
		}
	})

	t.Run("Update conserva el PublicID", func(t *testing.T) {
		repo := newRepo(t)
		a := contractSprint(contractProjectA, "Sprint 1", StatusOpen)
		a.PublicID = contractPublicIDA
		saved, err := repo.Save(a)
		if err != nil {
			t.Fatalf("Save devolvió un error inesperado: %v", err)
		}
		saved.Status = StatusClosed

		updated, err := repo.Update(saved)

		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		if updated.PublicID != contractPublicIDA {
			t.Errorf("Update: PublicID = %s, se esperaba %s", updated.PublicID, contractPublicIDA)
		}
		got, err := repo.GetByPublicID(contractPublicIDA)
		if err != nil {
			t.Fatalf("GetByPublicID devolvió un error inesperado: %v", err)
		}
		if got.ID != saved.ID || got.Status != StatusClosed {
			t.Errorf("GetByPublicID después de Update = %+v, se esperaba %+v", got, updated)
		}
	})
}

// HU-52 · CA-52.5 · Verifica que el repositorio en memoria guarde el identificador público del sprint, lo busque con GetByPublicID y lo conserve al actualizar.
func TestMemoryRepository_ContratoPublicID(t *testing.T) {
	runRepositoryPublicIDContract(t, func(*testing.T) Repository { return NewMemoryRepository() })
}

// HU-52 · CA-52.5 · Verifica que el repositorio PostgreSQL guarde el identificador público del sprint en su columna, lo busque con GetByPublicID y lo conserve al actualizar.
func TestPostgresRepository_ContratoPublicID(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryPublicIDContract(t, func(t *testing.T) Repository {
		return newPostgresTestRepository(t, pool)
	})
}

// HU-52 · CA-52.5 · Verifica que PostgreSQL rechace guardar un sprint con un identificador público que ya usa otro, sin crearlo.
func TestPostgresRepository_PublicIDRepetidoFalla(t *testing.T) {
	pool := postgrestest.Pool(t)
	repo := newPostgresTestRepository(t, pool)
	a := contractSprint(contractProjectA, "Sprint 1", StatusOpen)
	a.PublicID = contractPublicIDA
	savedA, err := repo.Save(a)
	if err != nil {
		t.Fatalf("Save devolvió un error inesperado: %v", err)
	}
	b := contractSprint(contractProjectB, "Sprint 2", StatusOpen)
	b.PublicID = contractPublicIDA

	savedB, err := repo.Save(b)

	if err == nil {
		t.Fatalf("Save con un PublicID repetido devolvió %+v sin error", savedB)
	}
	got, err := repo.GetByPublicID(contractPublicIDA)
	if err != nil {
		t.Fatalf("GetByPublicID devolvió un error inesperado: %v", err)
	}
	if got.ID != savedA.ID {
		t.Errorf("GetByPublicID = %+v, se esperaba el primero (%+v)", got, savedA)
	}
	if _, err := repo.GetByID(savedA.ID + 1); !errors.Is(err, ErrSprintNotFound) {
		t.Errorf("GetByID(%d): error = %v, se esperaba ErrSprintNotFound (no se creó el segundo)", savedA.ID+1, err)
	}
}
