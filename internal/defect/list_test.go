package defect

import (
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-31 · Consultar los defectos de un proyecto o de un sprint (Issue #49)
//
// Trazabilidad: CA-31.1 a CA-31.8. Los defectos de partida se guardan con
// Save del repositorio en memoria, porque registrarlos es HU-29 y cambiar su
// estado es HU-30.

// List completa los dobles de HU-29 y HU-30 para que sigan cumpliendo
// Repository. Se declara acá para no modificar los tests de esas historias.
func (r *testDefectRepository) List() ([]Defect, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append(make([]Defect, 0, len(r.defects)), r.defects...), nil
}

func (r failingDefectRepository) List() ([]Defect, error) {
	return nil, r.err
}

// countingListRepository cuenta las llamadas a List, para comprobar que una
// entidad inexistente no consulta el repositorio de defectos.
type countingListRepository struct {
	Repository
	mu    sync.Mutex
	lists int
}

func (r *countingListRepository) List() ([]Defect, error) {
	r.mu.Lock()
	r.lists++
	r.mu.Unlock()
	return r.Repository.List()
}

func (r *countingListRepository) listCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lists
}

var errItemList = errors.New("fallo del repositorio del backlog")

// failingItemListRepository falla al listar los ítems de un proyecto.
type failingItemListRepository struct {
	backlog.Repository
}

func (failingItemListRepository) ListByProject(int) ([]backlog.Item, error) {
	return nil, errItemList
}

// listEnv es la pila de HU-31: el proyecto del entorno con dos ítems y el
// sprint 1 abierto, y un repositorio en memoria envuelto para contar los List.
type listEnv struct {
	*defectEnv
	memory  *MemoryRepository
	counter *countingListRepository
	list    *Manager
	item1   backlog.Item
	item2   backlog.Item
	sprint1 sprint.Sprint
}

func newListEnv(t *testing.T) *listEnv {
	t.Helper()
	env := newDefectEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item1 := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, 0)
	item2 := env.addItem(t, "Carrito de compras", backlog.StatusPending, 0)
	memory := NewMemoryRepository()
	counter := &countingListRepository{Repository: memory}
	return &listEnv{
		defectEnv: env,
		memory:    memory,
		counter:   counter,
		list:      NewManager(counter, env.items, env.sprints),
		item1:     item1,
		item2:     item2,
		sprint1:   sp,
	}
}

// addOtherProject crea otro proyecto con un ítem y un sprint abierto.
func (e *listEnv) addOtherProject(t *testing.T) (project.Project, backlog.Item, sprint.Sprint) {
	t.Helper()
	p, err := e.projects.Create("Sistema de stock", "")
	if err != nil {
		t.Fatalf("crear el otro proyecto: %v", err)
	}
	sp := e.addSprintForProject(t, p.ID, "Sprint A", sprint.StatusOpen)
	item, err := e.itemRepo.Save(backlog.Item{
		ProjectID:          p.ID,
		Title:              "Alta de productos",
		Description:        "Formulario de alta",
		Priority:           backlog.PriorityMedium,
		AcceptanceCriteria: []string{"Producto guardado"},
		Status:             backlog.StatusPending,
	})
	if err != nil {
		t.Fatalf("guardar el ítem del otro proyecto: %v", err)
	}
	return p, item, sp
}

// saveDefect guarda un defecto con Save. Un defecto RESUELTO queda resuelto
// en resolutionSprintID; los demás, en 0.
func (e *listEnv) saveDefect(t *testing.T, itemID int, status Status, detectionSprintID, resolutionSprintID int) Defect {
	t.Helper()
	saved, err := e.memory.Save(Defect{
		Description:        "Defecto " + strconv.Itoa(itemID) + " " + string(status),
		Severity:           SeverityMedium,
		Status:             status,
		ItemID:             itemID,
		DetectionSprintID:  detectionSprintID,
		ResolutionSprintID: resolutionSprintID,
	})
	if err != nil {
		t.Fatalf("guardar el defecto de prueba: %v", err)
	}
	return saved
}

func assertDefectList(t *testing.T, op string, got, want []Defect) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s devolvió una lista nil, se esperaba %+v", op, want)
	}
	if !slices.Equal(got, want) {
		t.Errorf("%s = %+v, se esperaba %+v", op, got, want)
	}
}

func TestListByProject_DevuelveTodosLosDefectosEnCualquierEstadoOrdenadosPorID(t *testing.T) {
	env := newListEnv(t)
	open := env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)
	inCorrection := env.saveDefect(t, env.item2.ID, StatusInCorrection, env.sprint1.ID, 0)
	resolved := env.saveDefect(t, env.item1.ID, StatusResolved, env.sprint1.ID, env.sprint1.ID)

	got, err := env.list.ListByProject(env.projectID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}

	assertDefectList(t, "ListByProject", got, []Defect{open, inCorrection, resolved})
}

func TestListByProject_NoIncluyeDefectosDeOtroProyecto(t *testing.T) {
	env := newListEnv(t)
	own := env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)
	other, otherItem, otherSprint := env.addOtherProject(t)
	foreign := env.saveDefect(t, otherItem.ID, StatusOpen, otherSprint.ID, 0)

	got, err := env.list.ListByProject(env.projectID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}
	assertDefectList(t, "ListByProject del proyecto del entorno", got, []Defect{own})

	got, err = env.list.ListByProject(other.ID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}
	assertDefectList(t, "ListByProject del otro proyecto", got, []Defect{foreign})
}

func TestListByProject_ProyectoSinHistoriasDevuelveListaVacia(t *testing.T) {
	env := newListEnv(t)
	env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)
	empty, err := env.projects.Create("Proyecto vacío", "")
	if err != nil {
		t.Fatalf("crear el proyecto vacío: %v", err)
	}

	got, err := env.list.ListByProject(empty.ID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}

	assertDefectList(t, "ListByProject", got, []Defect{})
}

func TestListByProject_ProyectoSinDefectosDevuelveListaVacia(t *testing.T) {
	env := newListEnv(t)

	got, err := env.list.ListByProject(env.projectID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}

	assertDefectList(t, "ListByProject", got, []Defect{})
}

func TestListByProject_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	env := newListEnv(t)
	env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)

	for _, id := range []int{999, 0, -1} {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			got, err := env.list.ListByProject(id)

			assertDefectError(t, err, project.ErrProjectNotFound)
			if got != nil {
				t.Errorf("lista = %+v, se esperaba nil", got)
			}
		})
	}
	if n := env.counter.listCount(); n != 0 {
		t.Errorf("List se llamó %d veces, se esperaba ninguna", n)
	}
}

func TestListBySprint_DevuelveLosDefectosDetectadosEnElSprintEnCualquierEstado(t *testing.T) {
	env := newListEnv(t)
	open := env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)
	inCorrection := env.saveDefect(t, env.item2.ID, StatusInCorrection, env.sprint1.ID, 0)
	resolved := env.saveDefect(t, env.item1.ID, StatusResolved, env.sprint1.ID, env.sprint1.ID)

	got, err := env.list.ListBySprint(env.sprint1.ID)
	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}

	assertDefectList(t, "ListBySprint", got, []Defect{open, inCorrection, resolved})
}

func TestListBySprint_NoIncluyeDefectosDetectadosEnOtroSprintAunqueSeResolvieranEnEl(t *testing.T) {
	env := newListEnv(t)
	sprint2 := env.addSprint(t, "Sprint 2", sprint.StatusOpen)
	resolvedLater := env.saveDefect(t, env.item1.ID, StatusResolved, env.sprint1.ID, sprint2.ID)
	detectedLater := env.saveDefect(t, env.item2.ID, StatusOpen, sprint2.ID, 0)

	got, err := env.list.ListBySprint(env.sprint1.ID)
	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}
	assertDefectList(t, "ListBySprint del sprint 1", got, []Defect{resolvedLater})

	got, err = env.list.ListBySprint(sprint2.ID)
	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}
	assertDefectList(t, "ListBySprint del sprint 2", got, []Defect{detectedLater})

	got, err = env.list.ListByProject(env.projectID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}
	assertDefectList(t, "ListByProject", got, []Defect{resolvedLater, detectedLater})
}

func TestListBySprint_NoIncluyeDefectosDeSprintsDeOtroProyecto(t *testing.T) {
	env := newListEnv(t)
	own := env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)
	_, otherItem, otherSprint := env.addOtherProject(t)
	env.saveDefect(t, otherItem.ID, StatusOpen, otherSprint.ID, 0)

	got, err := env.list.ListBySprint(env.sprint1.ID)
	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}

	assertDefectList(t, "ListBySprint", got, []Defect{own})
}

func TestListBySprint_SprintCerradoSeConsulta(t *testing.T) {
	env := newListEnv(t)
	closed := env.addSprint(t, "Sprint 0", sprint.StatusClosed)
	defect := env.saveDefect(t, env.item1.ID, StatusInCorrection, closed.ID, 0)

	got, err := env.list.ListBySprint(closed.ID)
	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}

	assertDefectList(t, "ListBySprint", got, []Defect{defect})
}

func TestListBySprint_SprintSinDefectosDevuelveListaVacia(t *testing.T) {
	env := newListEnv(t)
	sprint2 := env.addSprint(t, "Sprint 2", sprint.StatusOpen)
	env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)

	got, err := env.list.ListBySprint(sprint2.ID)
	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}

	assertDefectList(t, "ListBySprint", got, []Defect{})
}

func TestListBySprint_SprintInexistenteDevuelveErrSprintNotFound(t *testing.T) {
	env := newListEnv(t)
	env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)

	for _, id := range []int{999, 0, -1} {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			got, err := env.list.ListBySprint(id)

			assertDefectError(t, err, sprint.ErrSprintNotFound)
			if got != nil {
				t.Errorf("lista = %+v, se esperaba nil", got)
			}
		})
	}
	if n := env.counter.listCount(); n != 0 {
		t.Errorf("List se llamó %d veces, se esperaba ninguna", n)
	}
}

func TestList_ModificarLoDevueltoNoAlteraLoGuardado(t *testing.T) {
	env := newListEnv(t)
	first := env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)
	second := env.saveDefect(t, env.item2.ID, StatusResolved, env.sprint1.ID, env.sprint1.ID)
	want := []Defect{first, second}

	lists := map[string]func() ([]Defect, error){
		"ListByProject": func() ([]Defect, error) { return env.list.ListByProject(env.projectID) },
		"ListBySprint":  func() ([]Defect, error) { return env.list.ListBySprint(env.sprint1.ID) },
	}
	for name, list := range lists {
		t.Run(name, func(t *testing.T) {
			got, err := list()
			if err != nil {
				t.Fatalf("%s devolvió un error inesperado: %v", name, err)
			}
			assertDefectList(t, name, got, want)

			got[0].Description = "Descripción modificada"
			got[0].Status = StatusResolved
			got[1].ItemID = 999
			got[1].ResolutionSprintID = 0
			got = append(got[:1], Defect{ID: 99})

			if stored := memorySnapshot(env.memory); !slices.Equal(stored, want) {
				t.Errorf("lo guardado = %+v, se esperaba %+v", stored, want)
			}
			again, err := list()
			if err != nil {
				t.Fatalf("%s devolvió un error inesperado: %v", name, err)
			}
			assertDefectList(t, name+" otra vez", again, want)
		})
	}
}

func TestList_ConsultarNoModificaNada(t *testing.T) {
	env := newListEnv(t)
	env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)
	env.saveDefect(t, env.item2.ID, StatusResolved, env.sprint1.ID, env.sprint1.ID)
	before := memorySnapshot(env.memory)
	beforeItem, err := env.items.GetItem(env.item1.ID)
	if err != nil {
		t.Fatalf("leer el ítem: %v", err)
	}
	beforeSprint, err := env.sprints.Get(env.sprint1.ID)
	if err != nil {
		t.Fatalf("leer el sprint: %v", err)
	}

	first, err := env.list.ListByProject(env.projectID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}
	if _, err := env.list.ListBySprint(env.sprint1.ID); err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}
	second, err := env.list.ListByProject(env.projectID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}

	assertDefectList(t, "ListByProject", first, before)
	assertDefectList(t, "ListByProject otra vez", second, first)
	if after := memorySnapshot(env.memory); !slices.Equal(after, before) {
		t.Errorf("los defectos guardados cambiaron: antes %+v, después %+v", before, after)
	}
	if item, _ := env.items.GetItem(env.item1.ID); !sameItem(item, beforeItem) {
		t.Errorf("el ítem cambió: antes %+v, después %+v", beforeItem, item)
	}
	if sp, _ := env.sprints.Get(env.sprint1.ID); sp != beforeSprint {
		t.Errorf("el sprint cambió: antes %+v, después %+v", beforeSprint, sp)
	}
}

func TestListByProject_ErrorDelRepositorioDeDefectos(t *testing.T) {
	env := newListEnv(t)
	manager := NewManager(failingDefectRepository{err: errDefectRepository}, env.items, env.sprints)

	got, err := manager.ListByProject(env.projectID)

	assertDefectError(t, err, errDefectRepository)
	if got != nil {
		t.Errorf("lista = %+v, se esperaba nil", got)
	}
}

func TestListBySprint_ErrorDelRepositorioDeDefectos(t *testing.T) {
	env := newListEnv(t)
	manager := NewManager(failingDefectRepository{err: errDefectRepository}, env.items, env.sprints)

	got, err := manager.ListBySprint(env.sprint1.ID)

	assertDefectError(t, err, errDefectRepository)
	if got != nil {
		t.Errorf("lista = %+v, se esperaba nil", got)
	}
}

func TestListByProject_ErrorDelBacklog(t *testing.T) {
	env := newListEnv(t)
	env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)
	items := backlog.NewManager(failingItemListRepository{Repository: env.itemRepo}, env.projects)
	manager := NewManager(env.memory, items, env.sprints)

	got, err := manager.ListByProject(env.projectID)

	assertDefectError(t, err, errItemList)
	if got != nil {
		t.Errorf("lista = %+v, se esperaba nil", got)
	}
}

func TestParseStatus_NormalizaLosEstadosValidos(t *testing.T) {
	cases := map[string]Status{
		"ABIERTO":           StatusOpen,
		"en_correccion":     StatusInCorrection,
		" Resuelto ":        StatusResolved,
		"\tabierto\n":       StatusOpen,
		"  EN_CORRECCION  ": StatusInCorrection,
	}
	for input, want := range cases {
		t.Run(strconv.Quote(input), func(t *testing.T) {
			got, err := ParseStatus(input)
			if err != nil {
				t.Fatalf("ParseStatus(%q) devolvió un error inesperado: %v", input, err)
			}
			if got != want {
				t.Errorf("ParseStatus(%q) = %q, se esperaba %q", input, got, want)
			}
		})
	}
}

func TestParseStatus_RechazaEstadosInvalidos(t *testing.T) {
	for _, input := range []string{"", "   ", "CERRADO", "PENDIENTE", "ABIERTO X", "EN CORRECCION"} {
		t.Run(strconv.Quote(input), func(t *testing.T) {
			got, err := ParseStatus(input)

			assertDefectError(t, err, ErrInvalidStatus)
			if got != "" {
				t.Errorf("ParseStatus(%q) = %q, se esperaba un estado vacío", input, got)
			}
		})
	}
}

func TestFilterByStatus_DevuelveSoloLosDelEstadoEnElMismoOrden(t *testing.T) {
	defects := []Defect{
		{ID: 1, Status: StatusOpen},
		{ID: 2, Status: StatusInCorrection},
		{ID: 3, Status: StatusResolved},
		{ID: 4, Status: StatusOpen},
	}

	cases := map[Status][]Defect{
		StatusOpen:         {defects[0], defects[3]},
		StatusInCorrection: {defects[1]},
		StatusResolved:     {defects[2]},
	}
	for status, want := range cases {
		t.Run(string(status), func(t *testing.T) {
			assertDefectList(t, "FilterByStatus", FilterByStatus(defects, status), want)
		})
	}
}

func TestFilterByStatus_SinCoincidenciasDevuelveListaVaciaNoNil(t *testing.T) {
	cases := map[string][]Defect{
		"sin coincidencias": {{ID: 1, Status: StatusOpen}},
		"lista vacía":       {},
		"lista nil":         nil,
	}
	for name, defects := range cases {
		t.Run(name, func(t *testing.T) {
			assertDefectList(t, "FilterByStatus", FilterByStatus(defects, StatusResolved), []Defect{})
		})
	}
}

func TestFilterByStatus_NoModificaLaListaRecibida(t *testing.T) {
	defects := []Defect{
		{ID: 1, Status: StatusOpen},
		{ID: 2, Status: StatusResolved},
		{ID: 3, Status: StatusOpen},
	}
	before := slices.Clone(defects)

	got := FilterByStatus(defects, StatusOpen)
	assertDefectList(t, "FilterByStatus", got, []Defect{before[0], before[2]})
	got[0].Status = StatusResolved
	got[1].ID = 99

	if !slices.Equal(defects, before) {
		t.Errorf("la lista recibida cambió: antes %+v, después %+v", before, defects)
	}
}

func TestList_ConsultasSimultaneasConRegistrosYCambiosDeEstado(t *testing.T) {
	env := newListEnv(t)
	changed := env.saveDefect(t, env.item1.ID, StatusOpen, env.sprint1.ID, 0)
	const workers = 20

	var wg sync.WaitGroup
	errs := make(chan error, 4*workers)
	for i := 0; i < workers; i++ {
		wg.Add(4)
		go func() {
			defer wg.Done()
			_, err := env.list.Register(env.item1.ID, "Defecto simultáneo "+strconv.Itoa(i), "ALTA", env.sprint1.ID)
			errs <- err
		}()
		go func() {
			defer wg.Done()
			_, err := env.list.ChangeStatus(changed.ID, "EN_CORRECCION", 0)
			errs <- err
		}()
		go func() {
			defer wg.Done()
			_, err := env.list.ListByProject(env.projectID)
			errs <- err
		}()
		go func() {
			defer wg.Done()
			_, err := env.list.ListBySprint(env.sprint1.ID)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Errorf("una operación simultánea falló: %v", err)
		}
	}

	byProject, err := env.list.ListByProject(env.projectID)
	if err != nil {
		t.Fatalf("ListByProject devolvió un error inesperado: %v", err)
	}
	bySprint, err := env.list.ListBySprint(env.sprint1.ID)
	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}
	stored := memorySnapshot(env.memory)
	if len(stored) != workers+1 {
		t.Fatalf("defectos guardados = %d, se esperaban %d", len(stored), workers+1)
	}
	assertDefectList(t, "ListByProject final", byProject, stored)
	assertDefectList(t, "ListBySprint final", bySprint, stored)
}
