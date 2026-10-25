package effort

import (
	"errors"
	"math"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-27 — Consultar los registros de esfuerzo (Issue #45)

func saveListEffort(t *testing.T, env *effortEnv, itemID, sprintID int, member, activity string, hours float64) Effort {
	t.Helper()
	saved, err := env.memory.Save(Effort{
		Member:   member,
		Date:     testEffortDate,
		Activity: activity,
		ItemID:   itemID,
		SprintID: sprintID,
		Hours:    hours,
	})
	if err != nil {
		t.Fatalf("guardar el esfuerzo de prueba: %v", err)
	}
	return saved
}

func TestListBySprint_DevuelveLosEsfuerzosConElTotal(t *testing.T) {
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, sp.ID)
	saveListEffort(t, env, item.ID, sp.ID, "García Santiago", "Implementar el registro", 2.5)
	saveListEffort(t, env, item.ID, sp.ID, "Buttini", "Revisar el registro", 1)

	got, err := env.manager.ListBySprint(sp.ID)

	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}
	if len(got.Efforts) != 2 {
		t.Fatalf("esfuerzos = %d, se esperaban 2", len(got.Efforts))
	}
	if got.Efforts[0].ID != 1 || got.Efforts[1].ID != 2 {
		t.Errorf("IDs = %d y %d, se esperaban 1 y 2", got.Efforts[0].ID, got.Efforts[1].ID)
	}
	if got.TotalHours != 3.5 {
		t.Errorf("total = %v, se esperaba 3.5", got.TotalHours)
	}
	if got.Efforts[0].Member != "García Santiago" || got.Efforts[0].Activity != "Implementar el registro" {
		t.Errorf("primer esfuerzo = %+v", got.Efforts[0])
	}
}

func TestListByItem_DevuelveLosEsfuerzosConElTotal(t *testing.T) {
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, sp.ID)
	saveListEffort(t, env, item.ID, sp.ID, "García Santiago", "Implementar el registro", 2.5)
	saveListEffort(t, env, item.ID, sp.ID, "Buttini", "Revisar el registro", 1)

	got, err := env.manager.ListByItem(item.ID)

	if err != nil {
		t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
	}
	if len(got.Efforts) != 2 {
		t.Fatalf("esfuerzos = %d, se esperaban 2", len(got.Efforts))
	}
	if got.Efforts[0].ID != 1 || got.Efforts[1].ID != 2 {
		t.Errorf("IDs = %d y %d, se esperaban 1 y 2", got.Efforts[0].ID, got.Efforts[1].ID)
	}
	if got.TotalHours != 3.5 {
		t.Errorf("total = %v, se esperaba 3.5", got.TotalHours)
	}
}

func TestListBySprint_SinEsfuerzosDevuelveListaVacia(t *testing.T) {
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)

	got, err := env.manager.ListBySprint(sp.ID)

	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}
	if got.Efforts == nil {
		t.Error("la lista es nil, se esperaba una lista vacía")
	}
	if len(got.Efforts) != 0 {
		t.Errorf("esfuerzos = %d, se esperaban 0", len(got.Efforts))
	}
	if got.TotalHours != 0 {
		t.Errorf("total = %v, se esperaba 0", got.TotalHours)
	}
}

func TestListByItem_SinEsfuerzosDevuelveListaVacia(t *testing.T) {
	env := newEffortEnv(t)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, 0)

	got, err := env.manager.ListByItem(item.ID)

	if err != nil {
		t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
	}
	if got.Efforts == nil {
		t.Error("la lista es nil, se esperaba una lista vacía")
	}
	if len(got.Efforts) != 0 {
		t.Errorf("esfuerzos = %d, se esperaban 0", len(got.Efforts))
	}
	if got.TotalHours != 0 {
		t.Errorf("total = %v, se esperaba 0", got.TotalHours)
	}
}

func TestListBySprint_AislaLosEsfuerzosDeOtroSprint(t *testing.T) {
	env := newEffortEnv(t)
	sp1 := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	sp2 := env.addSprint(t, "Sprint 2", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, sp1.ID)
	saveListEffort(t, env, item.ID, sp1.ID, "García Santiago", "Implementar el registro", 2.5)
	saveListEffort(t, env, item.ID, sp2.ID, "Buttini", "Preparar el despliegue", 4)

	got, err := env.manager.ListBySprint(sp1.ID)

	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}
	if len(got.Efforts) != 1 || got.Efforts[0].ID != 1 {
		t.Fatalf("esfuerzos = %+v, se esperaba solo el esfuerzo 1", got.Efforts)
	}
	if got.TotalHours != 2.5 {
		t.Errorf("total = %v, se esperaba 2.5", got.TotalHours)
	}
}

func TestListByItem_AislaLosEsfuerzosDeOtraHistoria(t *testing.T) {
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item1 := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, sp.ID)
	item2 := env.addItem(t, "Cierre de sesión", backlog.StatusInProgress, sp.ID)
	saveListEffort(t, env, item1.ID, sp.ID, "García Santiago", "Implementar el registro", 2.5)
	saveListEffort(t, env, item2.ID, sp.ID, "Buttini", "Implementar el cierre", 4)

	got, err := env.manager.ListByItem(item1.ID)

	if err != nil {
		t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
	}
	if len(got.Efforts) != 1 || got.Efforts[0].ID != 1 {
		t.Fatalf("esfuerzos = %+v, se esperaba solo el esfuerzo 1", got.Efforts)
	}
	if got.TotalHours != 2.5 {
		t.Errorf("total = %v, se esperaba 2.5", got.TotalHours)
	}
}

func TestListBySprint_SprintCerradoSePuedeConsultar(t *testing.T) {
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusClosed)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, sp.ID)
	saveListEffort(t, env, item.ID, sp.ID, "García Santiago", "Implementar el registro", 2)

	got, err := env.manager.ListBySprint(sp.ID)

	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}
	if len(got.Efforts) != 1 || got.TotalHours != 2 {
		t.Errorf("resumen = %+v, se esperaba un esfuerzo de 2 horas", got)
	}
}

func TestListBySprint_SprintInexistente(t *testing.T) {
	for _, id := range []int{999, 0, -1} {
		t.Run("sprint "+itoa(id), func(t *testing.T) {
			env := newEffortEnv(t)

			got, err := env.manager.ListBySprint(id)

			if !errors.Is(err, sprint.ErrSprintNotFound) {
				t.Fatalf("err = %v, se esperaba %v", err, sprint.ErrSprintNotFound)
			}
			if got.Efforts != nil || got.TotalHours != 0 {
				t.Errorf("resumen = %+v, se esperaba el resumen cero", got)
			}
		})
	}
}

func TestListByItem_ItemInexistente(t *testing.T) {
	for _, id := range []int{999, 0, -1} {
		t.Run("ítem "+itoa(id), func(t *testing.T) {
			env := newEffortEnv(t)

			got, err := env.manager.ListByItem(id)

			if !errors.Is(err, backlog.ErrItemNotFound) {
				t.Fatalf("err = %v, se esperaba %v", err, backlog.ErrItemNotFound)
			}
			if got.Efforts != nil || got.TotalHours != 0 {
				t.Errorf("resumen = %+v, se esperaba el resumen cero", got)
			}
		})
	}
}

func TestListBySprint_ErrorDelRepositorio(t *testing.T) {
	env := newEffortEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	env.manager = NewManager(failingEffortRepository{err: errEffortRepository}, env.sprints, env.items)

	got, err := env.manager.ListBySprint(sp.ID)

	if !errors.Is(err, errEffortRepository) {
		t.Fatalf("err = %v, se esperaba %v", err, errEffortRepository)
	}
	if got.Efforts != nil || got.TotalHours != 0 {
		t.Errorf("resumen = %+v, se esperaba el resumen cero", got)
	}
}

func TestListByItem_ErrorDelRepositorio(t *testing.T) {
	env := newEffortEnv(t)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, 0)
	env.manager = NewManager(failingEffortRepository{err: errEffortRepository}, env.sprints, env.items)

	got, err := env.manager.ListByItem(item.ID)

	if !errors.Is(err, errEffortRepository) {
		t.Fatalf("err = %v, se esperaba %v", err, errEffortRepository)
	}
	if got.Efforts != nil || got.TotalHours != 0 {
		t.Errorf("resumen = %+v, se esperaba el resumen cero", got)
	}
}

func TestListByItem_TotalExacto(t *testing.T) {
	env := newEffortEnv(t)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, 0)
	saveListEffort(t, env, item.ID, 0, "García Santiago", "Ajustar", 0.001)
	saveListEffort(t, env, item.ID, 0, "García Santiago", "Ajustar", 0.002)

	got, err := env.manager.ListByItem(item.ID)

	if err != nil {
		t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
	}
	if math.Abs(got.TotalHours-0.003) > 1e-12 {
		t.Errorf("total = %v, se esperaba 0.003", got.TotalHours)
	}
}

func TestListByItem_CopiasIndependientes(t *testing.T) {
	env := newEffortEnv(t)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, 0)
	saveListEffort(t, env, item.ID, 0, "García Santiago", "Implementar el registro", 2.5)

	got, err := env.manager.ListByItem(item.ID)
	if err != nil {
		t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
	}
	got.Efforts[0].Member = "Otra persona"
	got.Efforts[0].Activity = "Otra actividad"
	got.Efforts[0].Hours = 99

	again, err := env.manager.ListByItem(item.ID)
	if err != nil {
		t.Fatalf("ListByItem devolvió un error inesperado: %v", err)
	}
	if again.Efforts[0].Member != "García Santiago" || again.Efforts[0].Activity != "Implementar el registro" || again.Efforts[0].Hours != 2.5 {
		t.Errorf("el esfuerzo guardado cambió: %+v", again.Efforts[0])
	}
}

func TestListByItem_ConcurrenteConRegister(t *testing.T) {
	env, item, sp := envForRegister(t)

	const n = 30
	errs := make([]error, n*3)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n * 3)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			_, errs[i] = env.manager.Register(sp.ID, item.ID, "García Santiago", testEffortDate, "Implementar", 1)
		}(i)
		go func(i int) {
			defer wg.Done()
			<-start
			_, errs[n+i] = env.manager.ListByItem(item.ID)
		}(i)
		go func(i int) {
			defer wg.Done()
			<-start
			_, errs[2*n+i] = env.manager.ListBySprint(sp.ID)
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("operación concurrente %d falló: %v", i, err)
		}
	}
	got, err := env.manager.ListByItem(item.ID)
	if err != nil {
		t.Fatalf("ListByItem final falló: %v", err)
	}
	if len(got.Efforts) != n {
		t.Fatalf("esfuerzos = %d, se esperaban %d", len(got.Efforts), n)
	}
	if got.TotalHours != float64(n) {
		t.Errorf("total = %v, se esperaba %d", got.TotalHours, n)
	}
}
