package member

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-16 · Modificar los datos o el rol de un integrante (Issue #30)
//
// Trazabilidad: CA-16.1 a CA-16.5. La persistencia (CA-16.6) está en
// repository_update_contract_test.go y la API (CA-16.7) en internal/httpapi.

// updateStubRepository envuelve un repositorio en memoria y puede fallar al
// buscar o al actualizar, para probar las condiciones de error (CA-16.5).
type updateStubRepository struct {
	*MemoryRepository
	failGet    bool
	failUpdate bool
}

func (r *updateStubRepository) GetByID(id int) (Member, error) {
	if r.failGet {
		return Member{}, errStorage
	}
	return r.MemoryRepository.GetByID(id)
}

func (r *updateStubRepository) Update(m Member) (Member, error) {
	if r.failUpdate {
		return Member{}, errStorage
	}
	return r.MemoryRepository.Update(m)
}

// update modifica un integrante que el test espera que se modifique bien.
func (e *memberEnv) update(t *testing.T, id int, name, role string) Member {
	t.Helper()
	m, err := e.members.Update(id, name, role)
	if err != nil {
		t.Fatalf("Update(%d, %q, %q) devolvió %v", id, name, role, err)
	}
	return m
}

// expectStored verifica los integrantes guardados en el proyecto.
func (e *memberEnv) expectStored(t *testing.T, projectID int, want ...Member) {
	t.Helper()
	if got := e.stored(t, projectID); !reflect.DeepEqual(got, want) {
		t.Errorf("guardados = %+v, se esperaba %+v", got, want)
	}
}

// CA-16.1 · Se cambian el nombre y el rol en una sola operación.
func TestUpdate_CambiaElNombreYElRol(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got := e.update(t, 1, "Ana Pérez", "AGILE_ENABLER")

	want := Member{ID: 1, ProjectID: 1, Name: "Ana Pérez", Role: RoleAgileEnabler}
	if got != want {
		t.Errorf("Update = %+v, se esperaba %+v", got, want)
	}
	e.expectStored(t, 1, want)
}

// CA-16.1 · Se cambia solo el nombre reenviando el rol actual.
func TestUpdate_CambiaSoloElNombre(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got := e.update(t, 1, "Ana Pérez", "PRODUCT_BUILDER")

	if got.Name != "Ana Pérez" || got.Role != RoleProductBuilder {
		t.Errorf("Update = %+v, se esperaba el nombre nuevo y el mismo rol", got)
	}
}

// CA-16.1 · Se cambia solo el rol reenviando el nombre actual.
func TestUpdate_CambiaSoloElRol(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got := e.update(t, 1, "Ana", "PRODUCT_ARCHITECT")

	if got.Name != "Ana" || got.Role != RoleProductArchitect {
		t.Errorf("Update = %+v, se esperaba el mismo nombre y el rol nuevo", got)
	}
}

// RN-16.6 · Reenviar el mismo nombre y el mismo rol es válido.
func TestUpdate_SinCambiosEsValido(t *testing.T) {
	e := newMemberEnv(t)
	saved := e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got := e.update(t, 1, "Ana", "PRODUCT_BUILDER")

	if got != saved {
		t.Errorf("Update = %+v, se esperaba %+v", got, saved)
	}
	e.expectStored(t, 1, saved)
}

// CA-16.4 y RN-16.4 · Cambiar solo las mayúsculas del propio nombre es válido:
// el integrante no se compara consigo mismo.
func TestUpdate_CambiarLasMayusculasDelPropioNombreEsValido(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got := e.update(t, 1, "ANA", "PRODUCT_BUILDER")

	if got.Name != "ANA" {
		t.Errorf("Name = %q, se esperaba %q", got.Name, "ANA")
	}
}

// CA-16.1 · El nombre se guarda sin espacios en los extremos y el rol
// normalizado.
func TestUpdate_NormalizaElNombreYElRol(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got := e.update(t, 1, "  Ana Pérez  ", "  agile_enabler ")

	want := Member{ID: 1, ProjectID: 1, Name: "Ana Pérez", Role: RoleAgileEnabler}
	if got != want {
		t.Errorf("Update = %+v, se esperaba %+v", got, want)
	}
	e.expectStored(t, 1, want)
}

// CA-16.2 · Un nombre vacío, un rol vacío o un rol inválido se rechazan y el
// integrante queda como estaba.
func TestUpdate_RechazaNombreORolInvalidos(t *testing.T) {
	casos := []struct {
		nombre, name, role string
		want               error
	}{
		{"nombre vacío", "", "PRODUCT_BUILDER", ErrEmptyName},
		{"nombre con solo espacios", "   ", "PRODUCT_BUILDER", ErrEmptyName},
		{"nombre y rol vacíos", "", "", ErrEmptyName},
		{"rol vacío", "Ana", "", ErrMissingRole},
		{"rol con solo espacios", "Ana", "   ", ErrMissingRole},
		{"rol con espacio intermedio", "Ana", "Product Builder", ErrInvalidRole},
		{"rol que no existe", "Ana", "SCRUM_MASTER", ErrInvalidRole},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := newMemberEnv(t)
			saved := e.register(t, 1, "Ana", "PRODUCT_BUILDER")

			_, err := e.members.Update(1, c.name, c.role)

			if !errors.Is(err, c.want) {
				t.Errorf("err = %v, se esperaba %v", err, c.want)
			}
			e.expectStored(t, 1, saved)
		})
	}
}

// RN-16.7 · Los datos se validan antes de buscar al integrante.
func TestUpdate_NombreVacioEnIntegranteInexistenteDevuelveErrEmptyName(t *testing.T) {
	e := newMemberEnv(t)

	_, err := e.members.Update(99, "", "PRODUCT_BUILDER")

	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("err = %v, se esperaba %v", err, ErrEmptyName)
	}
}

// CA-16.4 · El nombre de otro integrante del proyecto se rechaza, sin
// distinguir mayúsculas y con cualquier rol.
func TestUpdate_RechazaElNombreDeOtroIntegranteDelProyecto(t *testing.T) {
	casos := []struct {
		nombre, name, role string
	}{
		{"mismo nombre", "Bruno", "PRODUCT_BUILDER"},
		{"otras mayúsculas", "BRUNO", "PRODUCT_BUILDER"},
		{"espacios en los extremos", "  Bruno ", "PRODUCT_BUILDER"},
		{"otro rol", "Bruno", "AGILE_ENABLER"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := newMemberEnv(t)
			ana := e.register(t, 1, "Ana", "PRODUCT_BUILDER")
			bruno := e.register(t, 1, "Bruno", "AGILE_ENABLER")

			_, err := e.members.Update(ana.ID, c.name, c.role)

			if !errors.Is(err, ErrDuplicateMember) {
				t.Errorf("err = %v, se esperaba %v", err, ErrDuplicateMember)
			}
			e.expectStored(t, 1, ana, bruno)
		})
	}
}

// CA-16.4 · El nombre de un integrante de otro proyecto se acepta.
func TestUpdate_AceptaElNombreDeUnIntegranteDeOtroProyecto(t *testing.T) {
	e := newMemberEnv(t)
	ana := e.register(t, 1, "Ana", "PRODUCT_BUILDER")
	e.register(t, 2, "Bruno", "AGILE_ENABLER")

	got := e.update(t, ana.ID, "Bruno", "PRODUCT_BUILDER")

	if got.Name != "Bruno" || got.ProjectID != 1 {
		t.Errorf("Update = %+v, se esperaba Bruno en el proyecto 1", got)
	}
}

// CA-16.3 · Un ID inexistente, 0 o negativo devuelve ErrMemberNotFound.
func TestUpdate_IntegranteInexistenteDevuelveErrMemberNotFound(t *testing.T) {
	for _, id := range []int{99, 0, -1} {
		t.Run(fmt.Sprintf("ID %d", id), func(t *testing.T) {
			e := newMemberEnv(t)
			saved := e.register(t, 1, "Ana", "PRODUCT_BUILDER")

			_, err := e.members.Update(id, "Bruno", "AGILE_ENABLER")

			if !errors.Is(err, ErrMemberNotFound) {
				t.Errorf("err = %v, se esperaba %v", err, ErrMemberNotFound)
			}
			e.expectStored(t, 1, saved)
		})
	}
}

// CA-16.3 · Sin integrantes registrados, también es ErrMemberNotFound.
func TestUpdate_SinIntegrantesDevuelveErrMemberNotFound(t *testing.T) {
	e := newMemberEnv(t)

	_, err := e.members.Update(1, "Ana", "PRODUCT_BUILDER")

	if !errors.Is(err, ErrMemberNotFound) {
		t.Errorf("err = %v, se esperaba %v", err, ErrMemberNotFound)
	}
}

// RN-16.9 · Se modifica un integrante de un proyecto FINALIZADO.
func TestUpdate_AceptaUnProyectoFinalizado(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")
	if _, err := e.projects.Update(1, "Sistema de ventas", "", string(project.StatusFinished)); err != nil {
		t.Fatalf("finalizar el proyecto: %v", err)
	}

	got := e.update(t, 1, "Ana Pérez", "PRODUCT_BUILDER")

	if got.Name != "Ana Pérez" {
		t.Errorf("Name = %q, se esperaba %q", got.Name, "Ana Pérez")
	}
}

// RN-16.5 · El integrante conserva su ID y su proyecto.
func TestUpdate_ConservaElIDYElProyecto(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")
	bruno := e.register(t, 2, "Bruno", "AGILE_ENABLER")

	got := e.update(t, bruno.ID, "Bruno Díaz", "AGILE_ENABLER")

	if got.ID != bruno.ID || got.ProjectID != 2 {
		t.Errorf("Update = %+v, se esperaba el ID %d en el proyecto 2", got, bruno.ID)
	}
	e.expectStored(t, 2, got)
}

// CA-16.5 · Update devuelve una copia.
func TestUpdate_DevuelveUnaCopia(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got := e.update(t, 1, "Ana Pérez", "AGILE_ENABLER")
	want := got
	got.Name = "Otro"
	got.Role = RoleProductArchitect

	e.expectStored(t, 1, want)
}

// CA-16.5 · Una falla del repositorio al buscar se devuelve tal cual.
func TestUpdate_ErrorAlBuscarSePropaga(t *testing.T) {
	e := newMemberEnv(t)
	saved := e.register(t, 1, "Ana", "PRODUCT_BUILDER")
	repo := &updateStubRepository{MemoryRepository: e.repo.MemoryRepository, failGet: true}

	_, err := NewManager(repo, e.projects).Update(1, "Bruno", "AGILE_ENABLER")

	if !errors.Is(err, errStorage) {
		t.Errorf("err = %v, se esperaba %v", err, errStorage)
	}
	e.expectStored(t, 1, saved)
}

// CA-16.5 · Una falla del repositorio al listar se devuelve tal cual.
func TestUpdate_ErrorAlLeerLosIntegrantesSePropaga(t *testing.T) {
	e := newMemberEnv(t)
	saved := e.register(t, 1, "Ana", "PRODUCT_BUILDER")
	e.repo.failList = true

	_, err := e.members.Update(1, "Bruno", "AGILE_ENABLER")

	if !errors.Is(err, errStorage) {
		t.Errorf("err = %v, se esperaba %v", err, errStorage)
	}
	e.expectStored(t, 1, saved)
}

// CA-16.5 · Una falla del repositorio al actualizar se devuelve tal cual.
func TestUpdate_ErrorAlActualizarSePropaga(t *testing.T) {
	e := newMemberEnv(t)
	saved := e.register(t, 1, "Ana", "PRODUCT_BUILDER")
	repo := &updateStubRepository{MemoryRepository: e.repo.MemoryRepository, failUpdate: true}

	_, err := NewManager(repo, e.projects).Update(1, "Bruno", "AGILE_ENABLER")

	if !errors.Is(err, errStorage) {
		t.Errorf("err = %v, se esperaba %v", err, errStorage)
	}
	e.expectStored(t, 1, saved)
}

// CA-16.5 y RN-16.10 · Varias modificaciones simultáneas al mismo nombre no
// dejan un duplicado. Corre con go test -race en el CI.
func TestUpdate_ModificacionesSimultaneasNoDejanDuplicados(t *testing.T) {
	e := newMemberEnv(t)
	const n = 20
	for i := range n {
		e.register(t, 1, fmt.Sprintf("Integrante %d", i), "PRODUCT_BUILDER")
	}
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = e.members.Update(i+1, "Ana", "PRODUCT_BUILDER")
		}()
	}
	wg.Wait()

	ok := 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case !errors.Is(err, ErrDuplicateMember):
			t.Errorf("err = %v, se esperaba nil o %v", err, ErrDuplicateMember)
		}
	}
	if ok != 1 {
		t.Errorf("modificaciones correctas = %d, se esperaba 1", ok)
	}
}

// RN-16.10 · Una modificación y un registro simultáneos del mismo nombre no
// dejan un duplicado: los dos usan el mismo mutex.
func TestUpdate_ModificacionYRegistroSimultaneosNoDejanDuplicados(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")
	var updateErr, registerErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, updateErr = e.members.Update(1, "Bruno", "PRODUCT_BUILDER")
	}()
	go func() {
		defer wg.Done()
		_, registerErr = e.members.Register(1, "Bruno", "AGILE_ENABLER")
	}()
	wg.Wait()

	if (updateErr == nil) == (registerErr == nil) {
		t.Fatalf("Update = %v y Register = %v, se esperaba que falle solo uno", updateErr, registerErr)
	}
	failed := updateErr
	if failed == nil {
		failed = registerErr
	}
	if !errors.Is(failed, ErrDuplicateMember) {
		t.Errorf("err = %v, se esperaba %v", failed, ErrDuplicateMember)
	}
}
