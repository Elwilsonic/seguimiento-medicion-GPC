package backlog

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"
)

// TEC-05 · Persistencia con PostgreSQL en Docker (Issue #35)

// Proyectos que usan los tests de contrato. En PostgreSQL existen porque
// la función que arma el repositorio los crea antes (clave foránea).
const (
	contractProjectA = 1
	contractProjectB = 2
)

// runRepositoryContract verifica el comportamiento que comparte toda
// implementación de Repository, en memoria o en PostgreSQL (CA-TEC05.3).
// newRepo devuelve un repositorio vacío en cada llamada.
func runRepositoryContract(t *testing.T, newRepo func(t *testing.T) Repository) {
	t.Run("Save asigna IDs correlativos y globales entre proyectos", func(t *testing.T) {
		repo := newRepo(t)

		first := contractSave(t, repo, contractItem(contractProjectA, "Login"))
		second := contractSave(t, repo, contractItem(contractProjectB, "Reporte"))
		third := contractSave(t, repo, contractItem(contractProjectA, "Logout"))

		if got := []int{first.ID, second.ID, third.ID}; !slices.Equal(got, []int{1, 2, 3}) {
			t.Errorf("IDs = %v, se esperaba [1 2 3]", got)
		}
	})

	t.Run("Save guarda todos los campos y GetByID los devuelve", func(t *testing.T) {
		repo := newRepo(t)
		item := Item{
			ProjectID:          contractProjectA,
			Title:              "Login",
			Description:        "Ingresar con usuario y contraseña",
			Priority:           PriorityHigh,
			AcceptanceCriteria: []string{"Valida el usuario", "Muestra un error", "Recuerda la sesión"},
			Status:             StatusInProgress,
			StoryPoints:        5,
			EstimatedHours:     7.5,
		}

		saved := contractSave(t, repo, item)
		item.ID = saved.ID
		contractAssertItem(t, "Save", saved, item)

		got, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertItem(t, "GetByID", got, item)
	})

	t.Run("Save no comparte la lista de criterios con el llamador", func(t *testing.T) {
		repo := newRepo(t)
		criteria := []string{"Primero", "Segundo"}
		item := contractItem(contractProjectA, "Login")
		item.AcceptanceCriteria = criteria

		saved := contractSave(t, repo, item)
		criteria[0] = "Cambiado en la lista enviada"
		saved.AcceptanceCriteria[1] = "Cambiado en la lista devuelta"

		got, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		if want := []string{"Primero", "Segundo"}; !slices.Equal(got.AcceptanceCriteria, want) {
			t.Errorf("criterios guardados = %q, se esperaba %q", got.AcceptanceCriteria, want)
		}
	})

	for name, criteria := range map[string][]string{"nil": nil, "vacía": {}} {
		t.Run("Criterios "+name+" se guardan y se leen como lista vacía", func(t *testing.T) {
			repo := newRepo(t)
			item := contractItem(contractProjectA, "Sin criterios")
			item.AcceptanceCriteria = criteria

			saved := contractSave(t, repo, item)
			got, err := repo.GetByID(saved.ID)
			if err != nil {
				t.Fatalf("GetByID devolvió un error inesperado: %v", err)
			}
			listed, err := repo.ListByProject(contractProjectA)
			if err != nil {
				t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
			}
			if len(listed) != 1 {
				t.Fatalf("ListByProject devolvió %d ítems, se esperaba 1", len(listed))
			}
			for op, got := range map[string][]string{"Save": saved.AcceptanceCriteria, "GetByID": got.AcceptanceCriteria, "ListByProject": listed[0].AcceptanceCriteria} {
				if len(got) != 0 {
					t.Errorf("%s: criterios = %q, se esperaba una lista vacía", op, got)
				}
			}
		})
	}

	t.Run("GetByID de un ID inexistente da ErrItemNotFound", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractItem(contractProjectA, "Login"))

		_, err := repo.GetByID(99)
		if !errors.Is(err, ErrItemNotFound) {
			t.Errorf("err = %v, se esperaba %v", err, ErrItemNotFound)
		}
	})

	t.Run("ListByProject devuelve solo los ítems del proyecto en orden de ID", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, contractItem(contractProjectA, "Login"))
		contractSave(t, repo, contractItem(contractProjectB, "Reporte"))
		third := contractSave(t, repo, contractItem(contractProjectA, "Logout"))

		got, err := repo.ListByProject(contractProjectA)
		if err != nil {
			t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("ListByProject devolvió %d ítems, se esperaban 2", len(got))
		}
		contractAssertItem(t, "ListByProject[0]", got[0], first)
		contractAssertItem(t, "ListByProject[1]", got[1], third)
	})

	// En PostgreSQL un UPDATE deja la fila al final de la tabla: sin ORDER BY,
	// el ítem modificado saldría último. El ítem pasa a otro proyecto y vuelve,
	// porque si solo cambia el título PostgreSQL puede hacer un HOT update que
	// conserva la posición en el índice de project_id y el orden no cambia.
	t.Run("ListByProject sigue en orden de ID después de modificar el primero", func(t *testing.T) {
		repo := newRepo(t)
		first := contractSave(t, repo, contractItem(contractProjectA, "Login"))
		second := contractSave(t, repo, contractItem(contractProjectA, "Logout"))
		third := contractSave(t, repo, contractItem(contractProjectA, "Perfil"))
		first.Title = "Login modificado"
		for _, projectID := range []int{contractProjectB, contractProjectA} {
			first.ProjectID = projectID
			if _, err := repo.Update(first); err != nil {
				t.Fatalf("Update devolvió un error inesperado: %v", err)
			}
		}

		got, err := repo.ListByProject(contractProjectA)
		if err != nil {
			t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("ListByProject devolvió %d ítems, se esperaban 3", len(got))
		}
		for i, want := range []Item{first, second, third} {
			contractAssertItem(t, fmt.Sprintf("ListByProject[%d]", i), got[i], want)
		}
	})

	t.Run("ListByProject de un proyecto sin ítems devuelve una lista vacía, no nil", func(t *testing.T) {
		repo := newRepo(t)
		contractSave(t, repo, contractItem(contractProjectA, "Login"))

		got, err := repo.ListByProject(contractProjectB)
		if err != nil {
			t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("ListByProject = %#v, se esperaba una lista vacía no nil", got)
		}
	})

	t.Run("Update reemplaza todos los campos del ítem", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractItem(contractProjectA, "Login"))
		changed := Item{
			ID:                 saved.ID,
			ProjectID:          contractProjectA,
			Title:              "Login con dos pasos",
			Description:        "Nueva descripción",
			Priority:           PriorityLow,
			AcceptanceCriteria: []string{"Pide un código", "Vence a los 5 minutos"},
			Status:             StatusFinished,
			StoryPoints:        8,
			EstimatedHours:     12.25,
		}

		got, err := repo.Update(changed)
		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		contractAssertItem(t, "Update", got, changed)
		stored, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		contractAssertItem(t, "GetByID después de Update", stored, changed)
	})

	t.Run("Update con criterios nil se lee como lista vacía", func(t *testing.T) {
		repo := newRepo(t)
		saved := contractSave(t, repo, contractItem(contractProjectA, "Login"))
		saved.AcceptanceCriteria = nil

		if _, err := repo.Update(saved); err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		got, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		if len(got.AcceptanceCriteria) != 0 {
			t.Errorf("criterios = %q, se esperaba una lista vacía", got.AcceptanceCriteria)
		}
	})

	t.Run("Update de un ID inexistente da ErrItemNotFound", func(t *testing.T) {
		repo := newRepo(t)
		item := contractItem(contractProjectA, "Nadie")
		item.ID = 99

		_, err := repo.Update(item)
		if !errors.Is(err, ErrItemNotFound) {
			t.Errorf("err = %v, se esperaba %v", err, ErrItemNotFound)
		}
	})

	// HU-11: el SprintID asignado se guarda, se actualiza y se lee igual en
	// memoria y en PostgreSQL.
	t.Run("Save y Update conservan el SprintID del ítem", func(t *testing.T) {
		repo := newRepo(t)
		item := contractItem(contractProjectA, "Login")
		item.SprintID = 7

		saved := contractSave(t, repo, item)
		if saved.SprintID != 7 {
			t.Errorf("SprintID guardado = %d, se esperaba 7", saved.SprintID)
		}
		saved.Title = "Login modificado"
		updated, err := repo.Update(saved)
		if err != nil {
			t.Fatalf("Update devolvió un error inesperado: %v", err)
		}
		if updated.SprintID != 7 {
			t.Errorf("SprintID actualizado = %d, se esperaba 7", updated.SprintID)
		}
		got, err := repo.GetByID(saved.ID)
		if err != nil {
			t.Fatalf("GetByID devolvió un error inesperado: %v", err)
		}
		if got.SprintID != 7 {
			t.Errorf("SprintID leído = %d, se esperaba 7", got.SprintID)
		}
	})
}

// contractItem devuelve un ítem válido del proyecto con ese título.
func contractItem(projectID int, title string) Item {
	return Item{
		ProjectID:          projectID,
		Title:              title,
		Priority:           PriorityMedium,
		AcceptanceCriteria: []string{"Criterio de " + title},
		Status:             StatusPending,
	}
}

// contractSave guarda item en repo y falla el test si hay un error.
func contractSave(t *testing.T, repo Repository, item Item) Item {
	t.Helper()
	saved, err := repo.Save(item)
	if err != nil {
		t.Fatalf("Save devolvió un error inesperado: %v", err)
	}
	return saved
}

// contractAssertItem compara got con want: los criterios con slices.Equal
// (así nil y una lista vacía son iguales) y el resto de los campos con
// reflect.DeepEqual, con los criterios en nil de los dos lados.
func contractAssertItem(t *testing.T, op string, got, want Item) {
	t.Helper()
	gotCriteria, wantCriteria := got.AcceptanceCriteria, want.AcceptanceCriteria
	got.AcceptanceCriteria, want.AcceptanceCriteria = nil, nil
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %+v, se esperaba %+v", op, got, want)
	}
	if !slices.Equal(gotCriteria, wantCriteria) {
		t.Errorf("%s: criterios = %q, se esperaba %q", op, gotCriteria, wantCriteria)
	}
}

// Los tests de contrato de memoria caracterizan el comportamiento actual:
// pasan sin cambios en MemoryRepository.
func TestMemoryRepository_Contrato(t *testing.T) {
	runRepositoryContract(t, func(*testing.T) Repository { return NewMemoryRepository() })
}
