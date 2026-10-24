package poker

import (
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// runRepositoryContract prueba el contrato que deben cumplir
// MemoryRepository y PostgresRepository. newRepo devuelve un repositorio
// vacío y el ID de un ítem existente.
func runRepositoryContract(t *testing.T, newRepo func(t *testing.T) (Repository, int)) {
	t.Helper()

	t.Run("Save asigna IDs correlativos y conserva los campos", func(t *testing.T) {
		repo, itemID := newRepo(t)

		first, err := repo.Save(Session{ItemID: itemID, Status: StatusOpen, Round: 1, Votes: []Vote{}})
		if err != nil {
			t.Fatalf("Save de la primera sesión: %v", err)
		}
		second, err := repo.Save(Session{ItemID: itemID, Status: StatusOpen, Round: 2, Votes: []Vote{}})
		if err != nil {
			t.Fatalf("Save de la segunda sesión: %v", err)
		}

		if first.ID != 1 || first.ItemID != itemID || first.Status != StatusOpen || first.Round != 1 {
			t.Errorf("primera sesión = %+v, se esperaba ID 1, ítem %d, ABIERTA y ronda 1", first, itemID)
		}
		if second.ID != 2 || second.Status != StatusOpen || second.Round != 2 {
			t.Errorf("segunda sesión = %+v, se esperaba ID 2, ABIERTA y ronda 2", second)
		}
	})

	t.Run("Save normaliza una lista de votos nil", func(t *testing.T) {
		repo, itemID := newRepo(t)

		saved, err := repo.Save(Session{ItemID: itemID, Status: StatusOpen, Round: 1})
		if err != nil {
			t.Fatalf("Save: %v", err)
		}
		if saved.Votes == nil {
			t.Error("Votes = nil, se esperaba una lista vacía no nula")
		}
		if len(saved.Votes) != 0 {
			t.Errorf("votos = %+v, se esperaba una sesión sin votos", saved.Votes)
		}
	})

	t.Run("Save devuelve una copia de los votos", func(t *testing.T) {
		repo, itemID := newRepo(t)

		saved, err := repo.Save(Session{
			ItemID: itemID,
			Status: StatusOpen,
			Round:  1,
			Votes:  []Vote{{Voter: "Cristobal", Points: 5}},
		})
		if err != nil {
			t.Fatalf("Save: %v", err)
		}
		if len(saved.Votes) != 1 {
			t.Fatalf("votos guardados = %+v, se esperaba 1 voto", saved.Votes)
		}
		saved.Votes[0].Points = 21

		stored, exists, err := repo.OpenByItem(itemID)
		if err != nil || !exists {
			t.Fatalf("OpenByItem(%d) = (%+v, %v, %v), se esperaba la sesión guardada", itemID, stored, exists, err)
		}
		if stored.Votes[0].Points != 5 {
			t.Errorf("puntos guardados = %d, la mutación externa no debía afectarlos", stored.Votes[0].Points)
		}
	})

	t.Run("OpenByItem devuelve la sesión abierta", func(t *testing.T) {
		repo, itemID := newRepo(t)

		if _, err := repo.Save(Session{ItemID: itemID, Status: StatusOpen, Round: 1, Votes: []Vote{}}); err != nil {
			t.Fatalf("Save: %v", err)
		}
		got, exists, err := repo.OpenByItem(itemID)
		if err != nil {
			t.Fatalf("OpenByItem(%d): %v", itemID, err)
		}
		if !exists {
			t.Fatalf("OpenByItem(%d) no encontró la sesión guardada", itemID)
		}
		if got.ItemID != itemID || got.Status != StatusOpen || got.Round != 1 {
			t.Errorf("sesión = %+v, se esperaba la sesión ABIERTA del ítem %d", got, itemID)
		}
	})

	t.Run("OpenByItem sin sesión devuelve false", func(t *testing.T) {
		repo, itemID := newRepo(t)

		got, exists, err := repo.OpenByItem(itemID)
		if err != nil {
			t.Fatalf("OpenByItem(%d): %v", itemID, err)
		}
		if exists {
			t.Errorf("OpenByItem(%d) = (%+v, true, nil), se esperaba exists = false", itemID, got)
		}
	})

	t.Run("OpenByItem ignora las sesiones de otro ítem", func(t *testing.T) {
		repo, itemID := newRepo(t)

		if _, err := repo.Save(Session{ItemID: itemID, Status: StatusOpen, Round: 1, Votes: []Vote{}}); err != nil {
			t.Fatalf("Save: %v", err)
		}
		_, exists, err := repo.OpenByItem(itemID + 999)
		if err != nil {
			t.Fatalf("OpenByItem: %v", err)
		}
		if exists {
			t.Error("OpenByItem devolvió una sesión de otro ítem")
		}
	})

	t.Run("OpenByItem devuelve una copia de los votos", func(t *testing.T) {
		repo, itemID := newRepo(t)

		if _, err := repo.Save(Session{
			ItemID: itemID,
			Status: StatusOpen,
			Round:  1,
			Votes:  []Vote{{Voter: "Cristobal", Points: 5}},
		}); err != nil {
			t.Fatalf("Save: %v", err)
		}
		got, _, err := repo.OpenByItem(itemID)
		if err != nil {
			t.Fatalf("OpenByItem: %v", err)
		}
		if len(got.Votes) != 1 {
			t.Fatalf("votos = %+v, se esperaba 1 voto", got.Votes)
		}
		got.Votes[0].Points = 13

		again, _, err := repo.OpenByItem(itemID)
		if err != nil {
			t.Fatalf("segundo OpenByItem: %v", err)
		}
		if len(again.Votes) != 1 {
			t.Fatalf("votos = %+v, se esperaba 1 voto", again.Votes)
		}
		if again.Votes[0].Points != 5 {
			t.Errorf("puntos = %d, la mutación externa no debía afectar la sesión guardada", again.Votes[0].Points)
		}
	})
}

// newMemoryRepo arma un repositorio en memoria con un proyecto y un ítem
// existentes.
func newMemoryRepo(t *testing.T) (Repository, int) {
	t.Helper()
	projects := project.NewManager(project.NewMemoryRepository())
	p, err := projects.Create("Proyecto de prueba", "Proyecto para el contrato")
	if err != nil {
		t.Fatalf("crear proyecto: %v", err)
	}
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	it, err := items.Create(p.ID, "Historia de ejemplo", "Descripción", "MEDIA", []string{"Criterio"})
	if err != nil {
		t.Fatalf("crear ítem: %v", err)
	}
	return NewMemoryRepository(), it.ID
}

func TestMemoryRepository_Contrato(t *testing.T) {
	runRepositoryContract(t, newMemoryRepo)
}
