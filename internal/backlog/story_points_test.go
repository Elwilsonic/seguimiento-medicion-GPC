package backlog

import (
	"errors"
	"math"
	"slices"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-20 · Asignar story points a un ítem del backlog (Issue #38)

// seedPointsItem guarda un ítem completo con ese estado y esos story points,
// para verificar que la asignación no toca el resto de los campos (RN-20.6).
func seedPointsItem(t *testing.T, repo *MemoryRepository, status Status, points int) Item {
	t.Helper()
	item, err := repo.Save(Item{
		ProjectID:          1,
		Title:              validTitle,
		Description:        "Formulario de acceso",
		Priority:           PriorityHigh,
		AcceptanceCriteria: []string{"Login OK", "Carga correcta"},
		Status:             status,
		StoryPoints:        points,
		EstimatedHours:     6,
		SprintID:           3,
	})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}
	return item
}

// newPointsManager arma un Manager con un repositorio en memoria vacío.
func newPointsManager() (*Manager, *MemoryRepository) {
	repo := NewMemoryRepository()
	return NewManager(repo, project.NewManager(project.NewMemoryRepository())), repo
}

// assertAllButPointsUnchanged compara todos los campos menos los story points.
func assertAllButPointsUnchanged(t *testing.T, got, want Item) {
	t.Helper()
	if got.ID != want.ID || got.ProjectID != want.ProjectID || got.Title != want.Title ||
		got.Description != want.Description || got.Priority != want.Priority ||
		got.Status != want.Status || got.EstimatedHours != want.EstimatedHours ||
		got.SprintID != want.SprintID {
		t.Errorf("el ítem cambió en un campo que no son los story points: %+v, se esperaba %+v", got, want)
	}
	if !slices.Equal(got.AcceptanceCriteria, want.AcceptanceCriteria) {
		t.Errorf("criterios = %q, se esperaban %q", got.AcceptanceCriteria, want.AcceptanceCriteria)
	}
}

// CA-20.1
func TestAssignStoryPoints_ItemSinEstimarQuedaConElValorYConservaElResto(t *testing.T) {
	manager, repo := newPointsManager()
	created := seedPointsItem(t, repo, StatusPending, 0)

	assigned, err := manager.AssignStoryPoints(created.ID, 5)

	if err != nil {
		t.Fatalf("AssignStoryPoints() devolvió un error inesperado: %v", err)
	}
	if assigned.StoryPoints != 5 {
		t.Errorf("story points devueltos = %d, se esperaban 5", assigned.StoryPoints)
	}
	assertAllButPointsUnchanged(t, assigned, created)
	saved := savedItem(t, repo, created.ID)
	if saved.StoryPoints != 5 {
		t.Errorf("story points guardados = %d, se esperaban 5", saved.StoryPoints)
	}
	assertAllButPointsUnchanged(t, saved, created)
}

// CA-20.1 (reestimar reemplaza el valor anterior)
func TestAssignStoryPoints_ReestimarReemplazaElValorAnterior(t *testing.T) {
	manager, repo := newPointsManager()
	created := seedPointsItem(t, repo, StatusPending, 3)

	_, err := manager.AssignStoryPoints(created.ID, 8)

	if err != nil {
		t.Fatalf("AssignStoryPoints() devolvió un error inesperado: %v", err)
	}
	if saved := savedItem(t, repo, created.ID); saved.StoryPoints != 8 {
		t.Errorf("story points guardados = %d, se esperaban 8", saved.StoryPoints)
	}
}

// CA-20.1 (el mismo valor es válido)
func TestAssignStoryPoints_MismoValorEsValido(t *testing.T) {
	manager, repo := newPointsManager()
	created := seedPointsItem(t, repo, StatusPending, 5)

	assigned, err := manager.AssignStoryPoints(created.ID, 5)

	if err != nil {
		t.Fatalf("AssignStoryPoints() devolvió un error inesperado: %v", err)
	}
	if assigned.StoryPoints != 5 || assigned.ID != created.ID {
		t.Errorf("ítem devuelto = %+v, se esperaba el ítem %d con 5 story points", assigned, created.ID)
	}
	if saved := savedItem(t, repo, created.ID); saved.StoryPoints != 5 {
		t.Errorf("story points guardados = %d, se esperaban 5", saved.StoryPoints)
	}
}

// CA-20.2
func TestAssignStoryPoints_AceptaCadaValorDeLaEscala(t *testing.T) {
	for _, points := range []int{1, 2, 3, 5, 8, 13, 21} {
		manager, repo := newPointsManager()
		created := seedPointsItem(t, repo, StatusPending, 0)

		_, err := manager.AssignStoryPoints(created.ID, points)

		if err != nil {
			t.Errorf("AssignStoryPoints(%d) devolvió un error inesperado: %v", points, err)
			continue
		}
		if saved := savedItem(t, repo, created.ID); saved.StoryPoints != points {
			t.Errorf("story points guardados = %d, se esperaban %d", saved.StoryPoints, points)
		}
	}
}

// CA-20.2 (definición única de la escala)
func TestStoryPointScale_DevuelveLaEscalaFibonacciEnOrden(t *testing.T) {
	want := []int{1, 2, 3, 5, 8, 13, 21}

	got := StoryPointScale()

	if !slices.Equal(got, want) {
		t.Errorf("StoryPointScale() = %v, se esperaba %v", got, want)
	}
}

// CA-20.2 (la escala no se puede modificar desde afuera)
func TestStoryPointScale_DevuelveUnaListaNuevaEnCadaLlamada(t *testing.T) {
	first := StoryPointScale()
	if len(first) == 0 {
		t.Fatal("StoryPointScale() devolvió una lista vacía")
	}
	first[0] = 99

	second := StoryPointScale()

	if !slices.Equal(second, []int{1, 2, 3, 5, 8, 13, 21}) {
		t.Errorf("StoryPointScale() = %v después de modificar la lista anterior", second)
	}
	// La validación tampoco cambia: 1 sigue siendo válido y 99 no.
	manager, repo := newPointsManager()
	created := seedPointsItem(t, repo, StatusPending, 0)
	if _, err := manager.AssignStoryPoints(created.ID, 99); !errors.Is(err, ErrInvalidStoryPoints) {
		t.Errorf("AssignStoryPoints(99) error = %v, se esperaba ErrInvalidStoryPoints", err)
	}
	if _, err := manager.AssignStoryPoints(created.ID, 1); err != nil {
		t.Errorf("AssignStoryPoints(1) devolvió un error inesperado: %v", err)
	}
}

// CA-20.3
func TestAssignStoryPoints_RechazaValoresFueraDeLaEscala(t *testing.T) {
	for _, points := range []int{0, -1, -5, 4, 6, 7, 22, 34, math.MaxInt, math.MinInt} {
		manager, repo := newPointsManager()
		created := seedPointsItem(t, repo, StatusPending, 3)

		assigned, err := manager.AssignStoryPoints(created.ID, points)

		if !errors.Is(err, ErrInvalidStoryPoints) {
			t.Errorf("AssignStoryPoints(%d) error = %v, se esperaba ErrInvalidStoryPoints", points, err)
		}
		if assigned.ID != 0 {
			t.Errorf("AssignStoryPoints(%d) devolvió %+v, se esperaba un ítem vacío", points, assigned)
		}
		if saved := savedItem(t, repo, created.ID); saved.StoryPoints != 3 {
			t.Errorf("story points guardados = %d después de asignar %d, se esperaban 3", saved.StoryPoints, points)
		}
	}
}

// CA-20.3 (la escala se valida antes que la existencia y el estado del ítem)
func TestAssignStoryPoints_ValorFueraDeLaEscalaSeInformaPrimero(t *testing.T) {
	manager, repo := newPointsManager()
	finished := seedPointsItem(t, repo, StatusFinished, 3)

	if _, err := manager.AssignStoryPoints(999, 4); !errors.Is(err, ErrInvalidStoryPoints) {
		t.Errorf("ítem inexistente: error = %v, se esperaba ErrInvalidStoryPoints", err)
	}
	if _, err := manager.AssignStoryPoints(finished.ID, 4); !errors.Is(err, ErrInvalidStoryPoints) {
		t.Errorf("ítem terminado: error = %v, se esperaba ErrInvalidStoryPoints", err)
	}
}

// CA-20.4
func TestAssignStoryPoints_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
	for _, id := range []int{999, 0, -1} {
		manager, repo := newPointsManager()
		created := seedPointsItem(t, repo, StatusPending, 0)

		_, err := manager.AssignStoryPoints(id, 5)

		if !errors.Is(err, ErrItemNotFound) {
			t.Errorf("AssignStoryPoints(%d, 5) error = %v, se esperaba ErrItemNotFound", id, err)
		}
		if saved := savedItem(t, repo, created.ID); saved.StoryPoints != 0 {
			t.Errorf("se modificó el ítem %d: story points = %d", created.ID, saved.StoryPoints)
		}
	}
}

// CA-20.4 (sin ítems creados)
func TestAssignStoryPoints_SinItemsCreadosDevuelveErrItemNotFound(t *testing.T) {
	manager, _ := newPointsManager()

	_, err := manager.AssignStoryPoints(1, 5)

	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
	}
}

// CA-20.5
func TestAssignStoryPoints_ItemTerminadoDevuelveErrItemFinished(t *testing.T) {
	for _, before := range []int{0, 3} {
		manager, repo := newPointsManager()
		created := seedPointsItem(t, repo, StatusFinished, before)

		assigned, err := manager.AssignStoryPoints(created.ID, 8)

		if !errors.Is(err, ErrItemFinished) {
			t.Errorf("con %d story points: error = %v, se esperaba ErrItemFinished", before, err)
		}
		if assigned.ID != 0 {
			t.Errorf("con %d story points: devolvió %+v, se esperaba un ítem vacío", before, assigned)
		}
		if saved := savedItem(t, repo, created.ID); saved.StoryPoints != before {
			t.Errorf("story points guardados = %d, se esperaban %d", saved.StoryPoints, before)
		}
	}
}

// CA-20.5 (límite conocido: un ítem reabierto con HU-08 vuelve a ser estimable)
func TestAssignStoryPoints_ItemReabiertoVuelveASerEstimable(t *testing.T) {
	manager, repo := newPointsManager()
	created := seedPointsItem(t, repo, StatusFinished, 3)
	if _, err := manager.ChangeStatus(created.ID, "EN_CURSO"); err != nil {
		t.Fatalf("reabrir el ítem: %v", err)
	}

	_, err := manager.AssignStoryPoints(created.ID, 8)

	if err != nil {
		t.Fatalf("AssignStoryPoints() devolvió un error inesperado: %v", err)
	}
	if saved := savedItem(t, repo, created.ID); saved.StoryPoints != 8 {
		t.Errorf("story points guardados = %d, se esperaban 8", saved.StoryPoints)
	}
}

// CA-20.5 (los ítems no terminados se estiman aunque estén asignados a un sprint)
func TestAssignStoryPoints_ItemNoTerminadoAsignadoAUnSprint(t *testing.T) {
	for _, status := range []Status{StatusPending, StatusInProgress} {
		manager, repo := newPointsManager()
		created := seedPointsItem(t, repo, status, 0)

		_, err := manager.AssignStoryPoints(created.ID, 13)

		if err != nil {
			t.Errorf("estado %s: error inesperado: %v", status, err)
			continue
		}
		saved := savedItem(t, repo, created.ID)
		if saved.StoryPoints != 13 || saved.SprintID != created.SprintID || saved.Status != status {
			t.Errorf("estado %s: guardado = %+v, se esperaban 13 story points en el sprint %d", status, saved, created.SprintID)
		}
	}
}

// CA-20.6 (el Update de HU-07 conserva los story points asignados por HU-20)
func TestManagerUpdate_ConservaLosStoryPointsAsignadosConAssignStoryPoints(t *testing.T) {
	manager, repo := newPointsManager()
	created := seedPointsItem(t, repo, StatusPending, 0)
	if _, err := manager.AssignStoryPoints(created.ID, 8); err != nil {
		t.Fatalf("asignar los story points: %v", err)
	}

	updated, err := manager.Update(created.ID, "Login con OAuth", "", "MEDIA", []string{"Redirige"}, 4)

	if err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if updated.StoryPoints != 8 {
		t.Errorf("story points devueltos = %d, se esperaban 8", updated.StoryPoints)
	}
	if saved := savedItem(t, repo, created.ID); saved.StoryPoints != 8 || saved.Title != "Login con OAuth" {
		t.Errorf("guardado = %+v, se esperaban 8 story points y el título nuevo", saved)
	}
}

// CA-20.7 (copia por valor)
func TestAssignStoryPoints_ElResultadoEsUnaCopia(t *testing.T) {
	manager, repo := newPointsManager()
	created := seedPointsItem(t, repo, StatusPending, 0)

	assigned, err := manager.AssignStoryPoints(created.ID, 5)
	if err != nil {
		t.Fatalf("AssignStoryPoints() devolvió un error inesperado: %v", err)
	}
	if len(assigned.AcceptanceCriteria) == 0 {
		t.Fatalf("el ítem devuelto no tiene criterios: %+v", assigned)
	}
	assigned.StoryPoints = 21
	assigned.AcceptanceCriteria[0] = "Otro criterio"

	saved := savedItem(t, repo, created.ID)
	if saved.StoryPoints != 5 {
		t.Errorf("story points guardados = %d, se esperaban 5", saved.StoryPoints)
	}
	assertAllButPointsUnchanged(t, saved, created)
}

// CA-20.7 (error del repositorio al buscar)
func TestAssignStoryPoints_ErrorDelRepositorioAlBuscarSeDevuelve(t *testing.T) {
	manager := NewManager(failingRepository{getErr: errRepository}, project.NewManager(project.NewMemoryRepository()))

	assigned, err := manager.AssignStoryPoints(1, 5)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
	if assigned.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", assigned)
	}
}

// CA-20.7 (error del repositorio al actualizar)
func TestAssignStoryPoints_ErrorDelRepositorioAlActualizarSeDevuelveYNoCambiaElItem(t *testing.T) {
	repo := NewMemoryRepository()
	created := seedPointsItem(t, repo, StatusPending, 3)
	manager := NewManager(failingUpdateStatusRepository{repo}, project.NewManager(project.NewMemoryRepository()))

	_, err := manager.AssignStoryPoints(created.ID, 8)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
	if saved := savedItem(t, repo, created.ID); saved.StoryPoints != 3 {
		t.Errorf("story points guardados = %d, se esperaban 3", saved.StoryPoints)
	}
}

// CA-20.7 (concurrencia). Las asignaciones corren junto con modificaciones de
// HU-07 y un cambio a TERMINADO de HU-08. Las aserciones no dependen del orden
// en que corran las goroutines:
//   - cada asignación termina bien o con ErrItemFinished (si el cambio a
//     TERMINADO llegó antes);
//   - el ítem termina TERMINADO, con los datos de la modificación;
//   - como después de TERMINADO no se escribe ningún valor y la escala no
//     tiene el 0, los story points guardados son 0 si ninguna asignación
//     terminó bien y, si alguna terminó bien, uno de los valores que se
//     guardaron con éxito.
//
// En el CI corre con go test -race.
func TestAssignStoryPoints_SimultaneasConUpdateYChangeStatus(t *testing.T) {
	const attempts = 50
	manager, repo := newPointsManager()
	created := seedPointsItem(t, repo, StatusInProgress, 0)
	scale := []int{1, 2, 3, 5, 8, 13, 21}

	errs := make([]error, attempts)
	sent := make([]int, attempts)
	var updateErrs [attempts]error
	var statusErr error
	var wg sync.WaitGroup
	for i := range attempts {
		sent[i] = scale[i%len(scale)]
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = manager.AssignStoryPoints(created.ID, sent[i])
		}(i)
		go func(i int) {
			defer wg.Done()
			_, updateErrs[i] = manager.Update(created.ID, "Login con OAuth", "Formulario de acceso",
				"ALTA", []string{"Login OK", "Carga correcta"}, 6)
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, statusErr = manager.ChangeStatus(created.ID, "TERMINADO")
	}()
	wg.Wait()

	if statusErr != nil {
		t.Fatalf("el cambio a TERMINADO falló: %v", statusErr)
	}
	var succeeded []int
	for i, err := range errs {
		switch {
		case err == nil:
			succeeded = append(succeeded, sent[i])
		case !errors.Is(err, ErrItemFinished):
			t.Errorf("la asignación %d falló con %v, se esperaba nil o ErrItemFinished", i+1, err)
		}
	}
	for i, err := range updateErrs {
		if err != nil {
			t.Errorf("la modificación %d falló: %v", i+1, err)
		}
	}

	saved := savedItem(t, repo, created.ID)
	if saved.Status != StatusFinished || saved.Title != "Login con OAuth" {
		t.Errorf("guardado = %+v, se esperaba TERMINADO con el título modificado", saved)
	}
	switch {
	case len(succeeded) == 0 && saved.StoryPoints != 0:
		t.Errorf("ninguna asignación terminó bien, pero quedaron %d story points", saved.StoryPoints)
	case len(succeeded) > 0 && !slices.Contains(succeeded, saved.StoryPoints):
		t.Errorf("story points guardados = %d, no es ninguno de los asignados con éxito %v", saved.StoryPoints, succeeded)
	}
}
