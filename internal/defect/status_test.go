package defect

import (
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-30 · Cambiar el estado de un defecto (Issue #48)
//
// Trazabilidad: CA-30.1 a CA-30.8. Los defectos de partida se guardan con
// Save del repositorio en memoria, porque registrarlos es HU-29 y algunos
// casos (el sprint de resolución anterior al de detección) no se pueden armar
// con Register.

// GetByID y Update completan los dobles de HU-29 para que sigan cumpliendo
// Repository. Los tests de HU-29 no los usan.
func (r *testDefectRepository) GetByID(id int) (Defect, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, defect := range r.defects {
		if defect.ID == id {
			return defect, nil
		}
	}
	return Defect{}, ErrDefectNotFound
}

func (r *testDefectRepository) Update(defect Defect) (Defect, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.defects {
		if r.defects[i].ID == defect.ID {
			r.defects[i] = defect
			return defect, nil
		}
	}
	return Defect{}, ErrDefectNotFound
}

func (r failingDefectRepository) GetByID(int) (Defect, error) {
	return Defect{}, r.err
}

func (r failingDefectRepository) Update(Defect) (Defect, error) {
	return Defect{}, r.err
}

// countingDefectRepository cuenta las llamadas a Update para comprobar el
// no-op de RN-30.4 y puede fallar al buscar o al actualizar (CA-30.8).
type countingDefectRepository struct {
	Repository
	mu        sync.Mutex
	updates   int
	getErr    error
	updateErr error
	log       *defectEventLog
}

func (r *countingDefectRepository) GetByID(id int) (Defect, error) {
	if r.getErr != nil {
		return Defect{}, r.getErr
	}
	return r.Repository.GetByID(id)
}

func (r *countingDefectRepository) Update(defect Defect) (Defect, error) {
	r.mu.Lock()
	r.updates++
	r.mu.Unlock()
	if r.updateErr != nil {
		return Defect{}, r.updateErr
	}
	updated, err := r.Repository.Update(defect)
	if err == nil && r.log != nil && defect.Status == StatusResolved {
		r.log.add("resolve")
	}
	return updated, err
}

func (r *countingDefectRepository) updateCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.updates
}

// memorySnapshot devuelve una copia de lo guardado en el repositorio en
// memoria, sin pasar por GetByID.
func memorySnapshot(r *MemoryRepository) []Defect {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Defect(nil), r.defects...)
}

// statusEnv es la pila de HU-30: proyecto, ítem 1, sprint 1 abierto y un
// repositorio en memoria envuelto para contar los Update.
type statusEnv struct {
	*defectEnv
	memory  *MemoryRepository
	counter *countingDefectRepository
	status  *Manager
	item    backlog.Item
	sprint1 sprint.Sprint
}

func newStatusEnv(t *testing.T) *statusEnv {
	t.Helper()
	env := newDefectEnv(t)
	sp := env.addSprint(t, "Sprint 1", sprint.StatusOpen)
	item := env.addItem(t, "Login de usuarios", backlog.StatusInProgress, 0)
	memory := NewMemoryRepository()
	counter := &countingDefectRepository{Repository: memory}
	return &statusEnv{
		defectEnv: env,
		memory:    memory,
		counter:   counter,
		status:    NewManager(counter, env.items, env.sprints),
		item:      item,
		sprint1:   sp,
	}
}

// addDefect guarda un defecto del ítem del entorno con Save. Un defecto
// RESUELTO queda resuelto en resolutionSprintID; los demás, en 0.
func (e *statusEnv) addDefect(t *testing.T, status Status, detectionSprintID, resolutionSprintID int) Defect {
	t.Helper()
	saved, err := e.memory.Save(Defect{
		Description:        "El total no descuenta el descuento",
		Severity:           SeverityHigh,
		Status:             status,
		ItemID:             e.item.ID,
		DetectionSprintID:  detectionSprintID,
		ResolutionSprintID: resolutionSprintID,
	})
	if err != nil {
		t.Fatalf("guardar el defecto de prueba: %v", err)
	}
	return saved
}

// storedDefect devuelve el defecto guardado con ese ID, leído sin GetByID.
func (e *statusEnv) storedDefect(t *testing.T, id int) Defect {
	t.Helper()
	for _, defect := range memorySnapshot(e.memory) {
		if defect.ID == id {
			return defect
		}
	}
	t.Fatalf("no hay un defecto guardado con el ID %d", id)
	return Defect{}
}

func assertStoredUnchanged(t *testing.T, env *statusEnv, before Defect) {
	t.Helper()
	if after := env.storedDefect(t, before.ID); after != before {
		t.Errorf("el defecto guardado cambió: antes %+v, después %+v", before, after)
	}
}

func TestChangeStatus_AbiertoAEnCorreccion(t *testing.T) {
	env := newStatusEnv(t)
	before := env.addDefect(t, StatusOpen, env.sprint1.ID, 0)

	got, err := env.status.ChangeStatus(before.ID, "EN_CORRECCION", 0)
	if err != nil {
		t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
	}

	want := before
	want.Status = StatusInCorrection
	if got != want {
		t.Errorf("defecto devuelto = %+v, se esperaba %+v", got, want)
	}
	if stored := env.storedDefect(t, before.ID); stored != want {
		t.Errorf("defecto guardado = %+v, se esperaba %+v", stored, want)
	}
}

func TestChangeStatus_EnCorreccionAResuelto(t *testing.T) {
	env := newStatusEnv(t)
	before := env.addDefect(t, StatusInCorrection, env.sprint1.ID, 0)

	got, err := env.status.ChangeStatus(before.ID, "RESUELTO", env.sprint1.ID)
	if err != nil {
		t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
	}

	want := before
	want.Status = StatusResolved
	want.ResolutionSprintID = env.sprint1.ID
	if got != want {
		t.Errorf("defecto devuelto = %+v, se esperaba %+v", got, want)
	}
	if stored := env.storedDefect(t, before.ID); stored != want {
		t.Errorf("defecto guardado = %+v, se esperaba %+v", stored, want)
	}
}

func TestChangeStatus_ResolverEnSprintPosteriorAlDeDeteccion(t *testing.T) {
	env := newStatusEnv(t)
	detection := env.addSprint(t, "Sprint 0", sprint.StatusClosed)
	later := env.addSprint(t, "Sprint 2", sprint.StatusOpen)
	before := env.addDefect(t, StatusInCorrection, detection.ID, 0)

	got, err := env.status.ChangeStatus(before.ID, "RESUELTO", later.ID)
	if err != nil {
		t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
	}
	if got.Status != StatusResolved || got.ResolutionSprintID != later.ID {
		t.Errorf("defecto devuelto = %+v, se esperaba RESUELTO en el sprint %d", got, later.ID)
	}
}

func TestChangeStatus_EnCorreccionAAbierto(t *testing.T) {
	env := newStatusEnv(t)
	before := env.addDefect(t, StatusInCorrection, env.sprint1.ID, 0)

	got, err := env.status.ChangeStatus(before.ID, "ABIERTO", 0)
	if err != nil {
		t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
	}
	if got.Status != StatusOpen || got.ResolutionSprintID != 0 {
		t.Errorf("defecto devuelto = %+v, se esperaba ABIERTO sin sprint de resolución", got)
	}
	if stored := env.storedDefect(t, before.ID); stored.Status != StatusOpen {
		t.Errorf("estado guardado = %q, se esperaba %q", stored.Status, StatusOpen)
	}
}

func TestChangeStatus_ReabrirLimpiaElSprintDeResolucion(t *testing.T) {
	env := newStatusEnv(t)
	before := env.addDefect(t, StatusResolved, env.sprint1.ID, env.sprint1.ID)

	got, err := env.status.ChangeStatus(before.ID, "EN_CORRECCION", 0)
	if err != nil {
		t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
	}

	want := before
	want.Status = StatusInCorrection
	want.ResolutionSprintID = 0
	if got != want {
		t.Errorf("defecto devuelto = %+v, se esperaba %+v", got, want)
	}
	if stored := env.storedDefect(t, before.ID); stored != want {
		t.Errorf("defecto guardado = %+v, se esperaba %+v", stored, want)
	}
}

func TestChangeStatus_TransicionesNoPermitidas(t *testing.T) {
	cases := map[string]struct {
		from       Status
		resolution int
		to         string
	}{
		"ABIERTO a RESUELTO": {StatusOpen, 0, "RESUELTO"},
		"RESUELTO a ABIERTO": {StatusResolved, 1, "ABIERTO"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := newStatusEnv(t)
			before := env.addDefect(t, tc.from, env.sprint1.ID, tc.resolution)

			_, err := env.status.ChangeStatus(before.ID, tc.to, env.sprint1.ID)

			assertDefectError(t, err, ErrInvalidTransition)
			assertStoredUnchanged(t, env, before)
		})
	}
}

func TestChangeStatus_MismoEstadoEsNoOp(t *testing.T) {
	cases := map[string]struct {
		status     Status
		resolution int
		sprintID   int
	}{
		"ABIERTO":                            {StatusOpen, 0, 0},
		"EN_CORRECCION":                      {StatusInCorrection, 0, 0},
		"RESUELTO con el mismo sprint":       {StatusResolved, 1, 1},
		"RESUELTO con otro sprint válido":    {StatusResolved, 1, 2},
		"RESUELTO con un sprint inexistente": {StatusResolved, 1, 999},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := newStatusEnv(t)
			env.addSprint(t, "Sprint 2", sprint.StatusOpen)
			before := env.addDefect(t, tc.status, env.sprint1.ID, tc.resolution)

			got, err := env.status.ChangeStatus(before.ID, string(tc.status), tc.sprintID)
			if err != nil {
				t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
			}
			if got != before {
				t.Errorf("defecto devuelto = %+v, se esperaba %+v", got, before)
			}
			assertStoredUnchanged(t, env, before)
			if n := env.counter.updateCount(); n != 0 {
				t.Errorf("Update se llamó %d veces, se esperaba ninguna", n)
			}
		})
	}
}

// RN-30.4: el sprint <= 0 se valida antes de leer el defecto, así que
// RESUELTO → RESUELTO sin sprint de resolución no es un no-op.
func TestChangeStatus_ResueltoAResueltoSinSprintDeResolucion(t *testing.T) {
	for _, sprintID := range []int{0, -1} {
		t.Run("sprint "+strconv.Itoa(sprintID), func(t *testing.T) {
			env := newStatusEnv(t)
			before := env.addDefect(t, StatusResolved, env.sprint1.ID, env.sprint1.ID)

			_, err := env.status.ChangeStatus(before.ID, "RESUELTO", sprintID)

			assertDefectError(t, err, ErrMissingResolutionSprint)
			assertStoredUnchanged(t, env, before)
		})
	}
}

func TestChangeStatus_NormalizaElEstado(t *testing.T) {
	cases := map[string]struct {
		from   Status
		status string
		want   Status
	}{
		"minúsculas y espacios a EN_CORRECCION": {StatusOpen, "  en_correccion ", StatusInCorrection},
		"mayúsculas mezcladas a RESUELTO":       {StatusInCorrection, " Resuelto", StatusResolved},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := newStatusEnv(t)
			before := env.addDefect(t, tc.from, env.sprint1.ID, 0)

			got, err := env.status.ChangeStatus(before.ID, tc.status, env.sprint1.ID)
			if err != nil {
				t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
			}
			if got.Status != tc.want {
				t.Errorf("estado = %q, se esperaba %q", got.Status, tc.want)
			}
		})
	}
}

func TestChangeStatus_EstadoInvalido(t *testing.T) {
	for _, status := range []string{"", "   ", "CERRADO", "BLOQUEADO"} {
		t.Run("estado "+strconv.Quote(status), func(t *testing.T) {
			env := newStatusEnv(t)
			before := env.addDefect(t, StatusOpen, env.sprint1.ID, 0)

			_, err := env.status.ChangeStatus(before.ID, status, env.sprint1.ID)

			assertDefectError(t, err, ErrInvalidStatus)
			assertStoredUnchanged(t, env, before)
		})
	}
}

func TestChangeStatus_EstadoInvalidoSobreDefectoInexistente(t *testing.T) {
	env := newStatusEnv(t)

	_, err := env.status.ChangeStatus(999, "CERRADO", 0)

	assertDefectError(t, err, ErrInvalidStatus)
}

func TestChangeStatus_DefectoInexistente(t *testing.T) {
	for _, id := range []int{999, 0, -1} {
		t.Run("ID "+strconv.Itoa(id), func(t *testing.T) {
			env := newStatusEnv(t)
			before := env.addDefect(t, StatusOpen, env.sprint1.ID, 0)

			_, err := env.status.ChangeStatus(id, "EN_CORRECCION", 0)

			assertDefectError(t, err, ErrDefectNotFound)
			assertStoredUnchanged(t, env, before)
		})
	}
}

func TestChangeStatus_SprintDeResolucionObligatorio(t *testing.T) {
	for _, sprintID := range []int{0, -1} {
		t.Run("sprint "+strconv.Itoa(sprintID), func(t *testing.T) {
			env := newStatusEnv(t)
			before := env.addDefect(t, StatusInCorrection, env.sprint1.ID, 0)

			_, err := env.status.ChangeStatus(before.ID, "RESUELTO", sprintID)

			assertDefectError(t, err, ErrMissingResolutionSprint)
			assertStoredUnchanged(t, env, before)
		})
	}
}

// El sprint de resolución se valida antes de consultar el repositorio: un
// defecto inexistente sin sprint da ErrMissingResolutionSprint.
func TestChangeStatus_SprintAusenteSobreDefectoInexistente(t *testing.T) {
	env := newStatusEnv(t)

	_, err := env.status.ChangeStatus(999, "RESUELTO", 0)

	assertDefectError(t, err, ErrMissingResolutionSprint)
}

func TestChangeStatus_IgnoraElSprintDeResolucionConOtroDestino(t *testing.T) {
	cases := map[string]struct {
		from Status
		to   string
		want Status
	}{
		"ABIERTO a EN_CORRECCION": {StatusOpen, "EN_CORRECCION", StatusInCorrection},
		"EN_CORRECCION a ABIERTO": {StatusInCorrection, "ABIERTO", StatusOpen},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := newStatusEnv(t)
			before := env.addDefect(t, tc.from, env.sprint1.ID, 0)

			got, err := env.status.ChangeStatus(before.ID, tc.to, 999)
			if err != nil {
				t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
			}
			if got.Status != tc.want || got.ResolutionSprintID != 0 {
				t.Errorf("defecto devuelto = %+v, se esperaba %q sin sprint de resolución", got, tc.want)
			}
			if stored := env.storedDefect(t, before.ID); stored.Status != tc.want || stored.ResolutionSprintID != 0 {
				t.Errorf("defecto guardado = %+v, se esperaba %q sin sprint de resolución", stored, tc.want)
			}
		})
	}
}

func TestChangeStatus_SprintDeResolucionInexistente(t *testing.T) {
	env := newStatusEnv(t)
	before := env.addDefect(t, StatusInCorrection, env.sprint1.ID, 0)

	_, err := env.status.ChangeStatus(before.ID, "RESUELTO", 999)

	assertDefectError(t, err, sprint.ErrSprintNotFound)
	assertStoredUnchanged(t, env, before)
}

func TestChangeStatus_SprintDeResolucionCerrado(t *testing.T) {
	env := newStatusEnv(t)
	closed := env.addSprint(t, "Sprint 2", sprint.StatusClosed)
	before := env.addDefect(t, StatusInCorrection, env.sprint1.ID, 0)

	_, err := env.status.ChangeStatus(before.ID, "RESUELTO", closed.ID)

	assertDefectError(t, err, sprint.ErrSprintClosed)
	assertStoredUnchanged(t, env, before)
}

func TestChangeStatus_SprintDeResolucionDeOtroProyecto(t *testing.T) {
	env := newStatusEnv(t)
	other, err := env.projects.Create("Sistema de facturación", "")
	if err != nil {
		t.Fatalf("crear el segundo proyecto: %v", err)
	}
	otherSprint := env.addSprintForProject(t, other.ID, "Sprint 1", sprint.StatusOpen)
	before := env.addDefect(t, StatusInCorrection, env.sprint1.ID, 0)

	_, err = env.status.ChangeStatus(before.ID, "RESUELTO", otherSprint.ID)

	assertDefectError(t, err, ErrResolutionSprintFromAnotherProject)
	assertStoredUnchanged(t, env, before)
}

// RN-30.6: validación defensiva. El defecto se guarda con Save, detectado en
// un sprint posterior al de resolución, porque con Register no se puede armar.
func TestChangeStatus_SprintDeResolucionAnteriorAlDeDeteccion(t *testing.T) {
	env := newStatusEnv(t)
	later := env.addSprint(t, "Sprint 2", sprint.StatusOpen)
	before := env.addDefect(t, StatusInCorrection, later.ID, 0)

	_, err := env.status.ChangeStatus(before.ID, "RESUELTO", env.sprint1.ID)

	assertDefectError(t, err, ErrResolutionBeforeDetection)
	assertStoredUnchanged(t, env, before)
}

func TestChangeStatus_ResolverEnElMismoSprintDeDeteccion(t *testing.T) {
	env := newStatusEnv(t)
	before := env.addDefect(t, StatusInCorrection, env.sprint1.ID, 0)

	got, err := env.status.ChangeStatus(before.ID, "RESUELTO", before.DetectionSprintID)
	if err != nil {
		t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
	}
	if got.ResolutionSprintID != before.DetectionSprintID {
		t.Errorf("sprint de resolución = %d, se esperaba %d", got.ResolutionSprintID, before.DetectionSprintID)
	}
}

// La transición se valida antes que el sprint de resolución.
func TestChangeStatus_TransicionAntesQueSprint(t *testing.T) {
	env := newStatusEnv(t)
	closed := env.addSprint(t, "Sprint 2", sprint.StatusClosed)
	before := env.addDefect(t, StatusOpen, env.sprint1.ID, 0)

	_, err := env.status.ChangeStatus(before.ID, "RESUELTO", closed.ID)

	assertDefectError(t, err, ErrInvalidTransition)
	assertStoredUnchanged(t, env, before)
}

func TestChangeStatus_NoModificaHistoriaSprintsNiOtrosCampos(t *testing.T) {
	env := newStatusEnv(t)
	before := env.addDefect(t, StatusInCorrection, env.sprint1.ID, 0)
	itemBefore, err := env.items.GetItem(env.item.ID)
	if err != nil {
		t.Fatalf("GetItem antes del cambio: %v", err)
	}
	sprintBefore, err := env.sprintRepo.GetByID(env.sprint1.ID)
	if err != nil {
		t.Fatalf("GetByID del sprint antes del cambio: %v", err)
	}

	got, err := env.status.ChangeStatus(before.ID, "RESUELTO", env.sprint1.ID)
	if err != nil {
		t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
	}

	stored := env.storedDefect(t, before.ID)
	if stored.Status != StatusResolved || stored.ResolutionSprintID != env.sprint1.ID {
		t.Errorf("defecto guardado = %+v, se esperaba RESUELTO en el sprint %d", stored, env.sprint1.ID)
	}
	for _, d := range []Defect{got, stored} {
		if d.ID != before.ID || d.Description != before.Description || d.Severity != before.Severity ||
			d.ItemID != before.ItemID || d.DetectionSprintID != before.DetectionSprintID {
			t.Errorf("cambiaron campos que no son el estado ni el sprint de resolución: antes %+v, después %+v", before, d)
		}
	}
	itemAfter, err := env.items.GetItem(env.item.ID)
	if err != nil {
		t.Fatalf("GetItem después del cambio: %v", err)
	}
	if !sameItem(itemAfter, itemBefore) {
		t.Errorf("la historia cambió: antes %+v, después %+v", itemBefore, itemAfter)
	}
	sprintAfter, err := env.sprintRepo.GetByID(env.sprint1.ID)
	if err != nil {
		t.Fatalf("GetByID del sprint después del cambio: %v", err)
	}
	if sprintAfter != sprintBefore {
		t.Errorf("el sprint cambió: antes %+v, después %+v", sprintBefore, sprintAfter)
	}
}

func TestChangeStatus_CopiaIndependiente(t *testing.T) {
	env := newStatusEnv(t)
	before := env.addDefect(t, StatusOpen, env.sprint1.ID, 0)

	got, err := env.status.ChangeStatus(before.ID, "EN_CORRECCION", 0)
	if err != nil {
		t.Fatalf("ChangeStatus devolvió un error inesperado: %v", err)
	}
	if got.Status != StatusInCorrection {
		t.Fatalf("estado devuelto = %q, se esperaba %q", got.Status, StatusInCorrection)
	}
	got.Description = "Otra descripción"
	got.Severity = SeverityLow
	got.Status = StatusResolved
	got.ItemID = 999
	got.DetectionSprintID = 999
	got.ResolutionSprintID = 999

	want := before
	want.Status = StatusInCorrection
	if stored := env.storedDefect(t, before.ID); stored != want {
		t.Errorf("el defecto guardado cambió al modificar el devuelto: %+v, se esperaba %+v", stored, want)
	}
}

func TestChangeStatus_ErrorDelRepositorio(t *testing.T) {
	cases := map[string]func(r *countingDefectRepository){
		"al buscar":     func(r *countingDefectRepository) { r.getErr = errDefectRepository },
		"al actualizar": func(r *countingDefectRepository) { r.updateErr = errDefectRepository },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			env := newStatusEnv(t)
			before := env.addDefect(t, StatusOpen, env.sprint1.ID, 0)
			setup(env.counter)

			_, err := env.status.ChangeStatus(before.ID, "EN_CORRECCION", 0)

			assertDefectError(t, err, errDefectRepository)
			assertStoredUnchanged(t, env, before)
		})
	}
}

func TestChangeStatus_Concurrente(t *testing.T) {
	env := newStatusEnv(t)
	before := env.addDefect(t, StatusOpen, env.sprint1.ID, 0)

	const n = 50
	errs := make([]error, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			_, errs[i] = env.status.ChangeStatus(before.ID, "EN_CORRECCION", 0)
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("cambio %d falló: %v", i, err)
		}
	}
	if stored := env.storedDefect(t, before.ID); stored.Status != StatusInCorrection {
		t.Errorf("estado guardado = %q, se esperaba %q", stored.Status, StatusInCorrection)
	}
	if n := env.counter.updateCount(); n != 1 {
		t.Errorf("Update se llamó %d veces, se esperaba 1 (los demás cambios son no-op)", n)
	}
}

func TestChangeStatus_ConcurrenteConCierre(t *testing.T) {
	for i := 0; i < 20; i++ {
		t.Run("intento "+strconv.Itoa(i), func(t *testing.T) {
			env := newStatusEnv(t)
			before := env.addDefect(t, StatusInCorrection, env.sprint1.ID, 0)
			log := &defectEventLog{}
			sprintRepo := &recordingSprintRepository{Repository: env.sprintRepo, log: log}
			sprintManager := sprint.NewManager(sprintRepo, env.projects, env.items)
			env.counter.log = log
			manager := NewManager(env.counter, env.items, sprintManager)

			start := make(chan struct{})
			var wg sync.WaitGroup
			var changeErr, closeErr error
			wg.Add(2)
			go func() {
				defer wg.Done()
				<-start
				_, changeErr = manager.ChangeStatus(before.ID, "RESUELTO", env.sprint1.ID)
			}()
			go func() {
				defer wg.Done()
				<-start
				_, closeErr = sprintManager.Close(env.sprint1.ID)
			}()
			close(start)
			wg.Wait()

			if closeErr != nil {
				t.Fatalf("Close falló: %v", closeErr)
			}
			if changeErr != nil && !errors.Is(changeErr, sprint.ErrSprintClosed) {
				t.Fatalf("ChangeStatus falló con un error inesperado: %v", changeErr)
			}
			stored := env.storedDefect(t, before.ID)
			if changeErr == nil && stored.Status != StatusResolved {
				t.Fatalf("ChangeStatus terminó bien pero el defecto quedó %+v", stored)
			}
			if changeErr != nil && stored != before {
				t.Fatalf("ChangeStatus falló pero el defecto cambió: %+v", stored)
			}
			events := log.snapshot()
			closeIndex := slices.Index(events, "close")
			if closeIndex == -1 {
				t.Fatalf("no se registró el cierre: %v", events)
			}
			if slices.Contains(events[closeIndex+1:], "resolve") {
				t.Fatalf("se resolvió el defecto después del cierre: %v", events)
			}
		})
	}
}

func TestStatusValues_DevuelveLosEstadosEnOrden(t *testing.T) {
	want := []string{string(StatusOpen), string(StatusInCorrection), string(StatusResolved)}

	if got := StatusValues(); !slices.Equal(got, want) {
		t.Errorf("StatusValues() = %v, se esperaba %v", got, want)
	}
}

func TestStatusValues_DevuelveUnaListaNueva(t *testing.T) {
	values := StatusValues()
	values[0] = "MODIFICADO"

	if got := StatusValues(); got[0] != string(StatusOpen) {
		t.Errorf("modificar la lista devuelta cambió los estados: %v", got)
	}
}
