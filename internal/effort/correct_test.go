package effort

import (
	"errors"
	"math"
	"testing"

	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-28 · Corregir un registro de esfuerzo (Issue #46)
//
// Trazabilidad: cada test cita su criterio de la SDD
// (docs/especificaciones/HU-28-corregir-esfuerzo.md). Se reutiliza el entorno
// de HU-26 (register_test.go).

// envForCorrect arma un entorno con un esfuerzo de 12 h de "Ana" en el sprint
// 1 ABIERTO.
func envForCorrect(t *testing.T) (*effortEnv, Effort, sprint.Sprint) {
	t.Helper()
	env, item, sp := envForRegister(t)
	saved, err := env.manager.Register(sp.ID, item.ID, "Ana", testEffortDate, "Desarrollo", 12)
	if err != nil {
		t.Fatalf("registrar el esfuerzo de prueba: %v", err)
	}
	return env, saved, sp
}

// storedEffort lee el esfuerzo guardado directo del repositorio en memoria,
// sin pasar por GetByID.
func storedEffort(t *testing.T, env *effortEnv, id int) Effort {
	t.Helper()
	for _, e := range env.memory.efforts {
		if e.ID == id {
			return e
		}
	}
	t.Fatalf("no hay un esfuerzo guardado con ID %d", id)
	return Effort{}
}

// closeSprint cierra el sprint directo en el repositorio.
func closeSprint(t *testing.T, env *effortEnv, sp sprint.Sprint) {
	t.Helper()
	sp.Status = sprint.StatusClosed
	if _, err := env.sprintRepo.Update(sp); err != nil {
		t.Fatalf("cerrar el sprint de prueba: %v", err)
	}
}

// CA-28.1 y RN-28.1 · Se corrigen las horas y se conservan los demás campos.
func TestCorrectHours_Exitoso(t *testing.T) {
	env, saved, _ := envForCorrect(t)
	want := saved
	want.Hours = 10

	got, err := env.manager.CorrectHours(saved.ID, 10)

	if err != nil {
		t.Fatalf("CorrectHours devolvió un error inesperado: %v", err)
	}
	if got != want {
		t.Errorf("CorrectHours = %+v, se esperaba %+v", got, want)
	}
	if stored := storedEffort(t, env, saved.ID); stored != want {
		t.Errorf("esfuerzo guardado = %+v, se esperaba %+v", stored, want)
	}
}

// CA-28.1 y RN-28.2 · Las horas con decimales se guardan sin redondear.
func TestCorrectHours_HorasFraccionadas(t *testing.T) {
	env, saved, _ := envForCorrect(t)

	if _, err := env.manager.CorrectHours(saved.ID, 2.75); err != nil {
		t.Fatalf("CorrectHours devolvió un error inesperado: %v", err)
	}
	if stored := storedEffort(t, env, saved.ID); stored.Hours != 2.75 {
		t.Errorf("horas guardadas = %v, se esperaba 2.75", stored.Hours)
	}
}

// RN-28.7 · Corregir con las mismas horas es válido y devuelve el esfuerzo
// igual.
func TestCorrectHours_MismasHoras(t *testing.T) {
	env, saved, _ := envForCorrect(t)

	got, err := env.manager.CorrectHours(saved.ID, 12)

	if err != nil {
		t.Fatalf("CorrectHours devolvió un error inesperado: %v", err)
	}
	if got != saved {
		t.Errorf("CorrectHours = %+v, se esperaba %+v", got, saved)
	}
}

// CA-28.2 y RN-28.2 · Las horas inválidas devuelven ErrInvalidHours, como al
// registrar, y no cambian nada.
func TestCorrectHours_HorasInvalidas(t *testing.T) {
	for _, hours := range []float64{0, -3, math.NaN(), math.Inf(1), math.Inf(-1)} {
		env, saved, _ := envForCorrect(t)

		got, err := env.manager.CorrectHours(saved.ID, hours)

		if !errors.Is(err, ErrInvalidHours) {
			t.Errorf("horas %v: err = %v, se esperaba %v", hours, err, ErrInvalidHours)
		}
		if got != (Effort{}) {
			t.Errorf("horas %v: CorrectHours = %+v, se esperaba el valor cero", hours, got)
		}
		if stored := storedEffort(t, env, saved.ID); stored != saved {
			t.Errorf("horas %v: el esfuerzo guardado cambió: %+v", hours, stored)
		}
	}
}

// CA-28.3 y RN-28.4 · Un esfuerzo de un sprint cerrado no se corrige.
func TestCorrectHours_SprintCerrado(t *testing.T) {
	env, saved, sp := envForCorrect(t)
	closeSprint(t, env, sp)

	got, err := env.manager.CorrectHours(saved.ID, 10)

	if !errors.Is(err, sprint.ErrSprintClosed) {
		t.Errorf("err = %v, se esperaba %v", err, sprint.ErrSprintClosed)
	}
	if got != (Effort{}) {
		t.Errorf("CorrectHours = %+v, se esperaba el valor cero", got)
	}
	if stored := storedEffort(t, env, saved.ID); stored != saved {
		t.Errorf("el esfuerzo guardado cambió: %+v", stored)
	}
}

// CA-28.4 y RN-28.3 · Un esfuerzo inexistente (incluidos 0 y negativos)
// devuelve ErrEffortNotFound.
func TestCorrectHours_EsfuerzoInexistente(t *testing.T) {
	env, saved, _ := envForCorrect(t)

	for _, id := range []int{99, 0, -1} {
		got, err := env.manager.CorrectHours(id, 10)

		if !errors.Is(err, ErrEffortNotFound) {
			t.Errorf("id %d: err = %v, se esperaba %v", id, err, ErrEffortNotFound)
		}
		if got != (Effort{}) {
			t.Errorf("id %d: CorrectHours = %+v, se esperaba el valor cero", id, got)
		}
	}
	if stored := storedEffort(t, env, saved.ID); stored != saved {
		t.Errorf("el esfuerzo guardado cambió: %+v", stored)
	}
}

// RN-28.6 · Primero las horas, después la existencia y por último el sprint:
// horas inválidas con un esfuerzo inexistente o de un sprint cerrado
// devuelven ErrInvalidHours, y un esfuerzo inexistente no llega al sprint.
func TestCorrectHours_OrdenDeLosErrores(t *testing.T) {
	env, saved, sp := envForCorrect(t)
	closeSprint(t, env, sp)

	if _, err := env.manager.CorrectHours(99, 0); !errors.Is(err, ErrInvalidHours) {
		t.Errorf("inexistente con horas inválidas: err = %v, se esperaba %v", err, ErrInvalidHours)
	}
	if _, err := env.manager.CorrectHours(saved.ID, -1); !errors.Is(err, ErrInvalidHours) {
		t.Errorf("sprint cerrado con horas inválidas: err = %v, se esperaba %v", err, ErrInvalidHours)
	}
	if _, err := env.manager.CorrectHours(99, 10); !errors.Is(err, ErrEffortNotFound) {
		t.Errorf("inexistente con el sprint cerrado: err = %v, se esperaba %v", err, ErrEffortNotFound)
	}
}

// RN-28.8 · Si la historia se reasignó a otro sprint, el esfuerzo se corrige
// igual y conserva su sprint.
func TestCorrectHours_HistoriaReasignada(t *testing.T) {
	env, saved, _ := envForCorrect(t)
	other := env.addSprint(t, "Sprint 2", sprint.StatusOpen)
	item, err := env.itemRepo.GetByID(saved.ItemID)
	if err != nil {
		t.Fatalf("leer el ítem: %v", err)
	}
	item.SprintID = other.ID
	if _, err := env.itemRepo.Update(item); err != nil {
		t.Fatalf("reasignar el ítem: %v", err)
	}

	got, err := env.manager.CorrectHours(saved.ID, 10)

	if err != nil {
		t.Fatalf("CorrectHours devolvió un error inesperado: %v", err)
	}
	if got.Hours != 10 || got.SprintID != saved.SprintID {
		t.Errorf("CorrectHours = %+v, se esperaban 10 horas en el sprint %d", got, saved.SprintID)
	}
}

// RN-28.9 · Modificar el esfuerzo devuelto no cambia lo guardado.
func TestCorrectHours_CopiaIndependiente(t *testing.T) {
	env, saved, _ := envForCorrect(t)

	got, err := env.manager.CorrectHours(saved.ID, 10)
	if err != nil {
		t.Fatalf("CorrectHours devolvió un error inesperado: %v", err)
	}
	got.Hours = 99
	got.Member = "Otra persona"

	if stored := storedEffort(t, env, saved.ID); stored.Hours != 10 || stored.Member != "Ana" {
		t.Errorf("el esfuerzo guardado cambió: %+v", stored)
	}
}

// CA-28.5 y RN-28.10 · Las horas reales del sprint (M5, HU-27) reflejan la
// corrección.
func TestCorrectHours_ReflejaLasHorasRealesDelSprint(t *testing.T) {
	env, saved, sp := envForCorrect(t)
	if _, err := env.manager.Register(sp.ID, saved.ItemID, "Luis", testEffortDate, "Pruebas", 5); err != nil {
		t.Fatalf("registrar el segundo esfuerzo: %v", err)
	}

	if _, err := env.manager.CorrectHours(saved.ID, 10); err != nil {
		t.Fatalf("CorrectHours devolvió un error inesperado: %v", err)
	}

	summary, err := env.manager.ListBySprint(sp.ID)
	if err != nil {
		t.Fatalf("ListBySprint devolvió un error inesperado: %v", err)
	}
	if summary.TotalHours != 15 {
		t.Errorf("total = %v, se esperaba 15", summary.TotalHours)
	}
}

// RN-28.3 · Una falla del repositorio al buscar el esfuerzo se propaga.
func TestCorrectHours_ErrorDelRepositorio(t *testing.T) {
	env := newEffortEnv(t)
	env.manager = NewManager(failingEffortRepository{err: errEffortRepository}, env.sprints, env.items)

	_, err := env.manager.CorrectHours(1, 10)

	if !errors.Is(err, errEffortRepository) {
		t.Errorf("err = %v, se esperaba %v", err, errEffortRepository)
	}
}
