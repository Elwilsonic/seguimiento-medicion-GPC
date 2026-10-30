package member

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-15 · Registrar integrante del proyecto (Issue #29)

var errStorage = errors.New("falla del almacenamiento")

// stubRepository envuelve un repositorio en memoria y puede fallar al listar
// o al guardar, para probar las condiciones de error (CA-15.5).
type stubRepository struct {
	*MemoryRepository
	failList bool
	failSave bool
	saves    int
}

func (r *stubRepository) Save(m Member) (Member, error) {
	r.saves++
	if r.failSave {
		return Member{}, errStorage
	}
	return r.MemoryRepository.Save(m)
}

func (r *stubRepository) ListByProject(projectID int) ([]Member, error) {
	if r.failList {
		return nil, errStorage
	}
	return r.MemoryRepository.ListByProject(projectID)
}

// memberEnv tiene un Manager de proyectos con dos proyectos (1 y 2) y el
// Manager de integrantes sobre un repositorio que puede fallar.
type memberEnv struct {
	projects *project.Manager
	repo     *stubRepository
	members  *Manager
}

func newMemberEnv(t *testing.T) *memberEnv {
	t.Helper()
	projects := project.NewManager(project.NewMemoryRepository())
	for _, name := range []string{"Sistema de ventas", "Inventario"} {
		if _, err := projects.Create(name, ""); err != nil {
			t.Fatalf("crear el proyecto %q: %v", name, err)
		}
	}
	repo := &stubRepository{MemoryRepository: NewMemoryRepository()}
	return &memberEnv{projects: projects, repo: repo, members: NewManager(repo, projects)}
}

func (e *memberEnv) register(t *testing.T, projectID int, name, role string) Member {
	t.Helper()
	m, err := e.members.Register(projectID, name, role)
	if err != nil {
		t.Fatalf("Register(%d, %q, %q) devolvió %v", projectID, name, role, err)
	}
	return m
}

func (e *memberEnv) stored(t *testing.T, projectID int) []Member {
	t.Helper()
	list, err := e.repo.MemoryRepository.ListByProject(projectID)
	if err != nil {
		t.Fatalf("leer los integrantes del proyecto %d: %v", projectID, err)
	}
	return list
}

// CA-15.1 · Se registra con nombre y rol válidos.
func TestRegister_GuardaElIntegranteConNombreYRol(t *testing.T) {
	e := newMemberEnv(t)

	got := e.register(t, 1, "Ana Pérez", "PRODUCT_BUILDER")

	want := Member{ID: 1, ProjectID: 1, Name: "Ana Pérez", Role: RoleProductBuilder}
	if got != want {
		t.Errorf("Register = %+v, se esperaba %+v", got, want)
	}
	if stored := e.stored(t, 1); !reflect.DeepEqual(stored, []Member{want}) {
		t.Errorf("guardados = %+v, se esperaba %+v", stored, []Member{want})
	}
}

// CA-15.1 · Los tres roles son válidos y los IDs son correlativos.
func TestRegister_AceptaLosTresRolesConIDsCorrelativos(t *testing.T) {
	e := newMemberEnv(t)

	a := e.register(t, 1, "Ana", "PRODUCT_ARCHITECT")
	b := e.register(t, 1, "Bruno", "AGILE_ENABLER")
	c := e.register(t, 2, "Carla", "PRODUCT_BUILDER")

	got := []Member{a, b, c}
	want := []Member{
		{ID: 1, ProjectID: 1, Name: "Ana", Role: RoleProductArchitect},
		{ID: 2, ProjectID: 1, Name: "Bruno", Role: RoleAgileEnabler},
		{ID: 3, ProjectID: 2, Name: "Carla", Role: RoleProductBuilder},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("integrantes = %+v, se esperaba %+v", got, want)
	}
}

// RN-15.1 y RN-15.2 · El nombre se guarda sin espacios en los extremos y el
// rol se normaliza.
func TestRegister_NormalizaElNombreYElRol(t *testing.T) {
	cases := []struct {
		role string
		want Role
	}{
		{"product_builder", RoleProductBuilder},
		{"  Agile_Enabler  ", RoleAgileEnabler},
		{"Product_Architect", RoleProductArchitect},
	}
	for _, c := range cases {
		t.Run(c.role, func(t *testing.T) {
			e := newMemberEnv(t)

			got := e.register(t, 1, "  Ana Pérez \t", c.role)

			if got.Name != "Ana Pérez" || got.Role != c.want {
				t.Errorf("Register = %+v, se esperaba nombre %q y rol %q", got, "Ana Pérez", c.want)
			}
		})
	}
}

// CA-15.2 · Nombre y rol son obligatorios, y el rol tiene que ser válido.
func TestRegister_RechazaNombreORolInvalidos(t *testing.T) {
	cases := []struct {
		name, role string
		want       error
	}{
		{"", "PRODUCT_BUILDER", ErrEmptyName},
		{"   ", "PRODUCT_BUILDER", ErrEmptyName},
		{"Ana", "", ErrMissingRole},
		{"Ana", "   ", ErrMissingRole},
		{"Ana", "Product Builder", ErrInvalidRole},
		{"Ana", "SCRUM_MASTER", ErrInvalidRole},
		{"Ana", "DEV", ErrInvalidRole},
		// RN-15.8: el nombre se valida antes que el rol.
		{"", "", ErrEmptyName},
	}
	for _, c := range cases {
		t.Run(c.name+"/"+c.role, func(t *testing.T) {
			e := newMemberEnv(t)

			_, err := e.members.Register(1, c.name, c.role)

			if !errors.Is(err, c.want) {
				t.Errorf("err = %v, se esperaba %v", err, c.want)
			}
			if stored := e.stored(t, 1); len(stored) != 0 {
				t.Errorf("se guardó %+v, se esperaba nada", stored)
			}
		})
	}
}

// CA-15.3 · El mismo nombre en el proyecto, sin distinguir mayúsculas y con
// cualquier rol, es un duplicado.
func TestRegister_RechazaElMismoNombreEnElProyecto(t *testing.T) {
	cases := []struct{ name, role string }{
		{"Ana", "PRODUCT_BUILDER"},
		{"ana", "PRODUCT_BUILDER"},
		{" ANA ", "PRODUCT_BUILDER"},
		// D1: con otro rol también es la misma persona.
		{"Ana", "AGILE_ENABLER"},
	}
	for _, c := range cases {
		t.Run(c.name+"/"+c.role, func(t *testing.T) {
			e := newMemberEnv(t)
			first := e.register(t, 1, "Ana", "PRODUCT_BUILDER")

			_, err := e.members.Register(1, c.name, c.role)

			if !errors.Is(err, ErrDuplicateMember) {
				t.Errorf("err = %v, se esperaba %v", err, ErrDuplicateMember)
			}
			if stored := e.stored(t, 1); !reflect.DeepEqual(stored, []Member{first}) {
				t.Errorf("guardados = %+v, se esperaba solo %+v", stored, first)
			}
		})
	}
}

// CA-15.3 · El mismo nombre en otro proyecto se acepta.
func TestRegister_AceptaElMismoNombreEnOtroProyecto(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got := e.register(t, 2, "Ana", "PRODUCT_BUILDER")

	if got.ProjectID != 2 || got.ID != 2 {
		t.Errorf("Register = %+v, se esperaba el integrante 2 del proyecto 2", got)
	}
}

// RN-15.5 · No hay límite por rol: dos Agile Enabler en el mismo proyecto.
func TestRegister_AceptaDosIntegrantesConElMismoRol(t *testing.T) {
	e := newMemberEnv(t)
	e.register(t, 1, "Ana", "AGILE_ENABLER")

	e.register(t, 1, "Bruno", "AGILE_ENABLER")

	if stored := e.stored(t, 1); len(stored) != 2 {
		t.Errorf("guardados = %+v, se esperaban 2", stored)
	}
}

// RN-15.6 · Se registra en un proyecto FINALIZADO.
func TestRegister_AceptaUnProyectoFinalizado(t *testing.T) {
	e := newMemberEnv(t)
	if _, err := e.projects.Update(1, "Sistema de ventas", "", string(project.StatusFinished)); err != nil {
		t.Fatalf("finalizar el proyecto: %v", err)
	}

	got := e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	if got.ProjectID != 1 {
		t.Errorf("Register = %+v, se esperaba en el proyecto 1", got)
	}
}

// CA-15.4 · Un proyecto inexistente, 0 o negativo devuelve
// project.ErrProjectNotFound y no se guarda nada.
func TestRegister_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	for _, id := range []int{99, 0, -1} {
		e := newMemberEnv(t)

		_, err := e.members.Register(id, "Ana", "PRODUCT_BUILDER")

		if !errors.Is(err, project.ErrProjectNotFound) {
			t.Errorf("Register(%d) err = %v, se esperaba %v", id, err, project.ErrProjectNotFound)
		}
		if e.repo.saves != 0 {
			t.Errorf("Register(%d) guardó %d vez(veces), se esperaba ninguna", id, e.repo.saves)
		}
	}
}

// RN-15.8 · Los datos se validan antes que el proyecto.
func TestRegister_NombreVacioEnProyectoInexistenteDevuelveErrEmptyName(t *testing.T) {
	e := newMemberEnv(t)

	_, err := e.members.Register(99, "", "PRODUCT_BUILDER")

	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("err = %v, se esperaba %v", err, ErrEmptyName)
	}
}

// CA-15.5 · Register devuelve una copia: modificarla no cambia lo guardado.
func TestRegister_DevuelveUnaCopia(t *testing.T) {
	e := newMemberEnv(t)
	got := e.register(t, 1, "Ana", "PRODUCT_BUILDER")

	got.Name = "Otro"
	got.Role = RoleAgileEnabler

	want := []Member{{ID: 1, ProjectID: 1, Name: "Ana", Role: RoleProductBuilder}}
	if stored := e.stored(t, 1); !reflect.DeepEqual(stored, want) {
		t.Errorf("guardados = %+v, se esperaba %+v", stored, want)
	}
}

// CA-15.5 · Si falla la lectura de los integrantes, se devuelve el error y no
// se guarda nada.
func TestRegister_ErrorAlLeerLosIntegrantesSePropaga(t *testing.T) {
	e := newMemberEnv(t)
	e.repo.failList = true

	_, err := e.members.Register(1, "Ana", "PRODUCT_BUILDER")

	if !errors.Is(err, errStorage) {
		t.Errorf("err = %v, se esperaba %v", err, errStorage)
	}
	if e.repo.saves != 0 {
		t.Errorf("se guardó %d vez(veces), se esperaba ninguna", e.repo.saves)
	}
}

// CA-15.5 · Si falla el guardado, se devuelve el error.
func TestRegister_ErrorAlGuardarSePropaga(t *testing.T) {
	e := newMemberEnv(t)
	e.repo.failSave = true

	_, err := e.members.Register(1, "Ana", "PRODUCT_BUILDER")

	if !errors.Is(err, errStorage) {
		t.Errorf("err = %v, se esperaba %v", err, errStorage)
	}
	if stored := e.stored(t, 1); len(stored) != 0 {
		t.Errorf("guardados = %+v, se esperaba nada", stored)
	}
}

// CA-15.5 y RN-15.9 · Dos registros simultáneos del mismo nombre no dejan un
// duplicado. Corre con go test -race en el CI.
func TestRegister_RegistrosSimultaneosNoDejanDuplicados(t *testing.T) {
	e := newMemberEnv(t)
	const n = 20
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = e.members.Register(1, "Ana", "PRODUCT_BUILDER")
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
		t.Errorf("registros correctos = %d, se esperaba 1", ok)
	}
	if stored := e.stored(t, 1); len(stored) != 1 {
		t.Errorf("guardados = %+v, se esperaba 1", stored)
	}
}
