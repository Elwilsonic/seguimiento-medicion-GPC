package member

import (
	"errors"
	"testing"

	"seguimiento-medicion-gpc/internal/postgrestest"
)

// HU-16 · Modificar los datos o el rol de un integrante (Issue #30)
//
// Contrato de GetByID y Update: toda implementación, en memoria o en
// PostgreSQL, busca por ID, reemplaza el nombre y el rol sin cambiar el
// proyecto, devuelve copias y usa ErrMemberNotFound (CA-16.6).

// runRepositoryUpdateContract verifica GetByID y Update. newRepo devuelve un
// repositorio vacío en cada llamada, con los proyectos 1 y 2 creados.
func runRepositoryUpdateContract(t *testing.T, newRepo func(t *testing.T) Repository) {
	t.Run("GetByID devuelve el integrante guardado", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, Member{ProjectID: 1, Name: "Ana", Role: RoleProductBuilder})
		saved := contractSave(t, repo, Member{ProjectID: 2, Name: "Bruno", Role: RoleAgileEnabler})

		got, err := repo.GetByID(saved.ID)

		if err != nil {
			t.Fatalf("GetByID(%d) devolvió %v", saved.ID, err)
		}
		if got != saved {
			t.Errorf("GetByID = %+v, se esperaba %+v", got, saved)
		}
	})

	t.Run("GetByID devuelve ErrMemberNotFound si no existe", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, Member{ProjectID: 1, Name: "Ana", Role: RoleProductBuilder})

		for _, id := range []int{99, 0, -1} {
			if _, err := repo.GetByID(id); !errors.Is(err, ErrMemberNotFound) {
				t.Errorf("GetByID(%d): err = %v, se esperaba %v", id, err, ErrMemberNotFound)
			}
		}
	})

	t.Run("GetByID devuelve una copia", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, Member{ProjectID: 1, Name: "Ana", Role: RoleProductBuilder})

		first := contractGet(t, repo, saved.ID)
		first.Name = "Otro"

		if got := contractGet(t, repo, saved.ID); got != saved {
			t.Errorf("GetByID = %+v, se esperaba %+v", got, saved)
		}
	})

	t.Run("Update reemplaza el nombre y el rol y conserva el proyecto", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, Member{ProjectID: 1, Name: "Ana", Role: RoleProductBuilder})

		got, err := repo.Update(Member{ID: saved.ID, ProjectID: 2, Name: "Ana Pérez", Role: RoleAgileEnabler})

		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		want := Member{ID: saved.ID, ProjectID: 1, Name: "Ana Pérez", Role: RoleAgileEnabler}
		if got != want {
			t.Errorf("Update = %+v, se esperaba %+v", got, want)
		}
		if stored := contractGet(t, repo, saved.ID); stored != want {
			t.Errorf("guardado = %+v, se esperaba %+v", stored, want)
		}
	})

	t.Run("Update no cambia a los demás integrantes", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, Member{ProjectID: 1, Name: "Ana", Role: RoleProductBuilder})
		second := contractSave(t, repo, Member{ProjectID: 1, Name: "Bruno", Role: RoleAgileEnabler})

		if _, err := repo.Update(Member{ID: first.ID, Name: "Ana Pérez", Role: RoleProductArchitect}); err != nil {
			t.Fatalf("Update devolvió %v", err)
		}

		if got := contractGet(t, repo, second.ID); got != second {
			t.Errorf("GetByID = %+v, se esperaba %+v", got, second)
		}
	})

	t.Run("Update devuelve ErrMemberNotFound si no existe", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, Member{ProjectID: 1, Name: "Ana", Role: RoleProductBuilder})

		_, err := repo.Update(Member{ID: 99, Name: "Bruno", Role: RoleAgileEnabler})

		if !errors.Is(err, ErrMemberNotFound) {
			t.Errorf("err = %v, se esperaba %v", err, ErrMemberNotFound)
		}
		if got := contractGet(t, repo, saved.ID); got != saved {
			t.Errorf("GetByID = %+v, se esperaba %+v", got, saved)
		}
	})

	t.Run("Update devuelve una copia", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, Member{ProjectID: 1, Name: "Ana", Role: RoleProductBuilder})

		got, err := repo.Update(Member{ID: saved.ID, Name: "Ana Pérez", Role: RoleAgileEnabler})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		want := got
		got.Name = "Otro"

		if stored := contractGet(t, repo, saved.ID); stored != want {
			t.Errorf("guardado = %+v, se esperaba %+v", stored, want)
		}
	})
}

func contractGet(t *testing.T, repo Repository, id int) Member {
	t.Helper()
	m, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID(%d) devolvió %v", id, err)
	}
	return m
}

// El repositorio en memoria cumple el contrato de GetByID y Update (CA-16.6).
func TestMemoryRepository_ContratoUpdate(t *testing.T) {
	runRepositoryUpdateContract(t, func(t *testing.T) Repository {
		return NewMemoryRepository()
	})
}

// El repositorio PostgreSQL cumple el mismo contrato que el de memoria
// (CA-16.6). Necesita TEST_DATABASE_URL; sin ella se omite.
func TestPostgresRepository_ContratoUpdate(t *testing.T) {
	pool := postgrestest.Pool(t)
	runRepositoryUpdateContract(t, func(t *testing.T) Repository {
		return newPostgresTestRepository(t, pool)
	})
}
