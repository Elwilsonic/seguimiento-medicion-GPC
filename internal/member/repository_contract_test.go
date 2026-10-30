package member

import (
	"reflect"
	"testing"
)

// HU-15 · Registrar integrante del proyecto (Issue #29)
//
// Contrato de Repository: toda implementación, en memoria o en PostgreSQL,
// asigna IDs correlativos desde 1, conserva los campos, devuelve copias y
// lista los integrantes de un proyecto ordenados por ID (CA-15.6).

// runRepositoryContract verifica el comportamiento que comparte toda
// implementación de Repository. newRepo devuelve un repositorio vacío en cada
// llamada, con los proyectos 1 y 2 creados.
func runRepositoryContract(t *testing.T, newRepo func(t *testing.T) Repository) {
	t.Run("Save asigna IDs correlativos desde 1 y conserva los campos", func(t *testing.T) {
		repo := newRepo(t)

		first := contractSave(t, repo, Member{ProjectID: 1, Name: "Ana Pérez", Role: RoleProductBuilder})
		second := contractSave(t, repo, Member{ProjectID: 2, Name: "Bruno", Role: RoleAgileEnabler})

		if want := (Member{ID: 1, ProjectID: 1, Name: "Ana Pérez", Role: RoleProductBuilder}); first != want {
			t.Errorf("primer Save = %+v, se esperaba %+v", first, want)
		}
		if want := (Member{ID: 2, ProjectID: 2, Name: "Bruno", Role: RoleAgileEnabler}); second != want {
			t.Errorf("segundo Save = %+v, se esperaba %+v", second, want)
		}
	})

	t.Run("Save ignora el ID recibido y asigna el siguiente", func(t *testing.T) {
		repo := newRepo(t)

		saved := contractSave(t, repo, Member{ID: 99, ProjectID: 1, Name: "Ana", Role: RoleProductArchitect})

		if saved.ID != 1 {
			t.Errorf("ID = %d, se esperaba 1", saved.ID)
		}
	})

	t.Run("ListByProject devuelve los integrantes del proyecto ordenados por ID", func(t *testing.T) {
		repo := newRepo(t)
		a := contractSave(t, repo, Member{ProjectID: 1, Name: "Ana", Role: RoleProductBuilder})
		contractSave(t, repo, Member{ProjectID: 2, Name: "Bruno", Role: RoleProductBuilder})
		c := contractSave(t, repo, Member{ProjectID: 1, Name: "Carla", Role: RoleAgileEnabler})

		got, err := repo.ListByProject(1)

		if err != nil {
			t.Fatalf("ListByProject devolvió %v", err)
		}
		if want := []Member{a, c}; !reflect.DeepEqual(got, want) {
			t.Errorf("ListByProject = %+v, se esperaba %+v", got, want)
		}
	})

	t.Run("ListByProject sin integrantes devuelve una lista vacía, no nil", func(t *testing.T) {
		repo := newRepo(t)

		got, err := repo.ListByProject(1)

		if err != nil {
			t.Fatalf("ListByProject devolvió %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("ListByProject = %#v, se esperaba []Member{}", got)
		}
	})

	t.Run("ListByProject devuelve una lista nueva", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, Member{ProjectID: 1, Name: "Ana", Role: RoleProductBuilder})
		first, err := repo.ListByProject(1)
		if err != nil {
			t.Fatalf("ListByProject devolvió %v", err)
		}

		first[0].Name = "Otro"

		again, err := repo.ListByProject(1)
		if err != nil {
			t.Fatalf("ListByProject devolvió %v", err)
		}
		if !reflect.DeepEqual(again, []Member{saved}) {
			t.Errorf("ListByProject = %+v, se esperaba %+v", again, []Member{saved})
		}
	})
}

func contractSave(t *testing.T, repo Repository, m Member) Member {
	t.Helper()
	saved, err := repo.Save(m)
	if err != nil {
		t.Fatalf("Save(%+v) devolvió %v", m, err)
	}
	return saved
}
