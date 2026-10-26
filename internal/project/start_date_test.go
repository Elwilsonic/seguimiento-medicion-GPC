package project_test

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-18 · Registrar la fecha de inicio del proyecto (Issue #36)

// updateFailingRepository es un MemoryRepository que, con failUpdate, falla
// al actualizar. Permite verificar que el proyecto guardado no cambia cuando
// la escritura falla (CA-18.8).
type updateFailingRepository struct {
	*project.MemoryRepository
	failUpdate bool
}

func (r *updateFailingRepository) Update(p project.Project) (project.Project, error) {
	if r.failUpdate {
		return project.Project{}, errRepository
	}
	return r.MemoryRepository.Update(p)
}

// day devuelve la fecha AAAA-MM-DD a las 00:00 UTC, como la lee la API.
func day(t *testing.T, date string) time.Time {
	t.Helper()
	d, err := parseDay(date)
	if err != nil {
		t.Fatalf("fecha de prueba inválida %q: %v", date, err)
	}
	return d
}

// mustSetStartDate registra la fecha y falla el test si hay un error.
func mustSetStartDate(t *testing.T, manager *project.Manager, id int, date time.Time) {
	t.Helper()
	if _, err := manager.SetStartDate(id, date); err != nil {
		t.Fatalf("SetStartDate(%d, %v) devolvió un error inesperado: %v", id, date, err)
	}
}

// mustGet consulta el proyecto y falla el test si hay un error.
func mustGet(t *testing.T, manager *project.Manager, id int) project.Project {
	t.Helper()
	p, err := manager.Get(id)
	if err != nil {
		t.Fatalf("Get(%d) devolvió un error inesperado: %v", id, err)
	}
	return p
}

// CA-18.1
func TestSetStartDate_RegistraLaFechaYConservaLosDemasDatos(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Sistema de ventas", "Gestión de ventas"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if _, err := manager.Update(1, "Sistema de ventas", "Gestión de ventas", "EN_CURSO"); err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	date := day(t, "2026-10-23")

	p, err := manager.SetStartDate(1, date)

	if err != nil {
		t.Fatalf("SetStartDate() devolvió un error inesperado: %v", err)
	}
	want := project.Project{ID: 1, Name: "Sistema de ventas", Description: "Gestión de ventas", Status: project.StatusInProgress, StartDate: date}
	if p != want {
		t.Errorf("proyecto devuelto = %+v, se esperaba %+v", p, want)
	}
	if stored := mustGet(t, manager, 1); stored != want {
		t.Errorf("proyecto guardado = %+v, se esperaba %+v", stored, want)
	}
}

// CA-18.1 (RN-18.5): vale en los tres estados del proyecto.
func TestSetStartDate_SeRegistraEnCualquierEstado(t *testing.T) {
	for _, status := range []project.Status{project.StatusPlanned, project.StatusInProgress, project.StatusFinished} {
		t.Run(string(status), func(t *testing.T) {
			manager, _ := newManager()
			if _, err := manager.Create("Proyecto A", ""); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}
			if _, err := manager.Update(1, "Proyecto A", "", string(status)); err != nil {
				t.Fatalf("Update() devolvió un error inesperado: %v", err)
			}
			date := day(t, "2026-10-23")

			p, err := manager.SetStartDate(1, date)

			if err != nil {
				t.Fatalf("SetStartDate() devolvió un error inesperado: %v", err)
			}
			if !p.StartDate.Equal(date) || p.Status != status {
				t.Errorf("proyecto = %+v, se esperaba la fecha %v y el estado %q", p, date, status)
			}
		})
	}
}

// CA-18.1 (RN-18.5): no se limita a fechas pasadas ni futuras.
func TestSetStartDate_AceptaFechasPasadasFuturasYDeAnioBisiesto(t *testing.T) {
	for _, date := range []string{"2020-01-01", "2030-12-31", "2028-02-29"} {
		t.Run(date, func(t *testing.T) {
			manager, _ := newManager()
			if _, err := manager.Create("Proyecto A", ""); err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}

			_, err := manager.SetStartDate(1, day(t, date))

			if err != nil {
				t.Fatalf("SetStartDate() devolvió un error inesperado: %v", err)
			}
			if got := mustGet(t, manager, 1).StartDate; !got.Equal(day(t, date)) {
				t.Errorf("fecha guardada = %v, se esperaba %s", got, date)
			}
		})
	}
}

// CA-18.5
func TestSetStartDate_VolverARegistrarReemplazaLaFecha(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	mustSetStartDate(t, manager, 1, day(t, "2026-10-23"))

	p, err := manager.SetStartDate(1, day(t, "2026-11-02"))

	if err != nil {
		t.Fatalf("SetStartDate() devolvió un error inesperado: %v", err)
	}
	if !p.StartDate.Equal(day(t, "2026-11-02")) {
		t.Errorf("fecha devuelta = %v, se esperaba 2026-11-02", p.StartDate)
	}
	if got := mustGet(t, manager, 1).StartDate; !got.Equal(day(t, "2026-11-02")) {
		t.Errorf("fecha guardada = %v, se esperaba 2026-11-02", got)
	}
}

// CA-18.5 (RN-18.4): registrar la misma fecha es válido.
func TestSetStartDate_MismaFechaEsValida(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	mustSetStartDate(t, manager, 1, day(t, "2026-10-23"))

	p, err := manager.SetStartDate(1, day(t, "2026-10-23"))

	if err != nil {
		t.Fatalf("SetStartDate() devolvió un error inesperado: %v", err)
	}
	if !p.StartDate.Equal(day(t, "2026-10-23")) {
		t.Errorf("fecha devuelta = %v, se esperaba 2026-10-23", p.StartDate)
	}
}

// CA-18.3
func TestSetStartDate_FechaAusenteDevuelveErrMissingStartDateYNoCambia(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	mustSetStartDate(t, manager, 1, day(t, "2026-10-23"))

	p, err := manager.SetStartDate(1, time.Time{})

	if !errors.Is(err, project.ErrMissingStartDate) {
		t.Errorf("error = %v, se esperaba ErrMissingStartDate", err)
	}
	if p != (project.Project{}) {
		t.Errorf("proyecto devuelto = %+v, se esperaba un proyecto vacío", p)
	}
	if got := mustGet(t, manager, 1).StartDate; !got.Equal(day(t, "2026-10-23")) {
		t.Errorf("fecha guardada = %v, se esperaba 2026-10-23", got)
	}
}

// CA-18.3 (orden de validación): la fecha ausente se informa sin consultar el
// repositorio, aunque el proyecto no exista o el repositorio falle.
func TestSetStartDate_FechaAusenteSeInformaSinConsultarElRepositorio(t *testing.T) {
	cases := map[string]project.Repository{
		"proyecto inexistente":  project.NewMemoryRepository(),
		"repositorio que falla": failingRepository{getErr: errRepository, updateErr: errRepository},
	}
	for caseName, repo := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager := project.NewManager(repo)

			_, err := manager.SetStartDate(1, time.Time{})

			if !errors.Is(err, project.ErrMissingStartDate) {
				t.Errorf("error = %v, se esperaba ErrMissingStartDate", err)
			}
		})
	}
}

// CA-18.4
func TestSetStartDate_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	for _, id := range []int{999, 0, -1} {
		t.Run(fmt.Sprint(id), func(t *testing.T) {
			p, err := manager.SetStartDate(id, day(t, "2026-10-23"))

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
			}
			if p != (project.Project{}) {
				t.Errorf("proyecto devuelto = %+v, se esperaba un proyecto vacío", p)
			}
		})
	}
	if got := mustGet(t, manager, 1).StartDate; !got.IsZero() {
		t.Errorf("fecha del proyecto 1 = %v, se esperaba sin registrar", got)
	}
}

// CA-18.4
func TestSetStartDate_SinProyectosCreadosDevuelveErrProjectNotFound(t *testing.T) {
	manager, _ := newManager()

	_, err := manager.SetStartDate(1, day(t, "2026-10-23"))

	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
	}
}

// CA-18.1 (RN-18.6): registrar la fecha de un proyecto no cambia los demás.
func TestSetStartDate_NoModificaOtrosProyectos(t *testing.T) {
	manager, _ := newManager()
	for _, name := range []string{"Proyecto A", "Proyecto B"} {
		if _, err := manager.Create(name, ""); err != nil {
			t.Fatalf("Create() devolvió un error inesperado: %v", err)
		}
	}

	mustSetStartDate(t, manager, 2, day(t, "2026-10-23"))

	if got := mustGet(t, manager, 1); got != (project.Project{ID: 1, Name: "Proyecto A", Status: project.StatusPlanned}) {
		t.Errorf("proyecto 1 = %+v, se esperaba sin cambios", got)
	}
	if got := mustGet(t, manager, 2).StartDate; !got.Equal(day(t, "2026-10-23")) {
		t.Errorf("fecha del proyecto 2 = %v, se esperaba 2026-10-23", got)
	}
}

// CA-18.6 (RN-18.7)
func TestCreate_ProyectoNuevoNoTieneFechaDeInicio(t *testing.T) {
	manager, _ := newManager()

	p, err := manager.Create("Proyecto A", "")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if !p.StartDate.IsZero() {
		t.Errorf("fecha devuelta = %v, se esperaba sin registrar", p.StartDate)
	}
	if got := mustGet(t, manager, 1).StartDate; !got.IsZero() {
		t.Errorf("fecha guardada = %v, se esperaba sin registrar", got)
	}
}

// CA-18.6: la consulta (HU-03) devuelve la fecha registrada.
func TestGet_DevuelveLaFechaDeInicioRegistrada(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	mustSetStartDate(t, manager, 1, day(t, "2026-10-23"))

	p := mustGet(t, manager, 1)

	if !p.StartDate.Equal(day(t, "2026-10-23")) {
		t.Errorf("fecha consultada = %v, se esperaba 2026-10-23", p.StartDate)
	}
}

// CA-18.7
func TestSetStartDate_ModificarElResultadoDevueltoNoAlteraElGuardado(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	p, err := manager.SetStartDate(1, day(t, "2026-10-23"))
	if err != nil {
		t.Fatalf("SetStartDate() devolvió un error inesperado: %v", err)
	}
	p.StartDate = day(t, "2020-01-01")
	p.Name = "Otro nombre"

	stored := mustGet(t, manager, 1)
	if !stored.StartDate.Equal(day(t, "2026-10-23")) || stored.Name != "Proyecto A" {
		t.Errorf("proyecto guardado = %+v, se esperaba la fecha 2026-10-23 y el nombre Proyecto A", stored)
	}
}

// CA-18.8: si el repositorio falla al buscar o al actualizar, se devuelve su
// error y el proyecto no cambia.
func TestSetStartDate_ErrorDelRepositorioSeDevuelveYElProyectoNoCambia(t *testing.T) {
	t.Run("al buscar", func(t *testing.T) {
		manager := project.NewManager(failingRepository{getErr: errRepository})

		p, err := manager.SetStartDate(1, day(t, "2026-10-23"))

		if !errors.Is(err, errRepository) {
			t.Errorf("error = %v, se esperaba el error del repositorio", err)
		}
		if p != (project.Project{}) {
			t.Errorf("proyecto devuelto = %+v, se esperaba un proyecto vacío", p)
		}
	})

	t.Run("al actualizar", func(t *testing.T) {
		repo := &updateFailingRepository{MemoryRepository: project.NewMemoryRepository()}
		manager := project.NewManager(repo)
		if _, err := manager.Create("Proyecto A", ""); err != nil {
			t.Fatalf("Create() devolvió un error inesperado: %v", err)
		}
		repo.failUpdate = true

		p, err := manager.SetStartDate(1, day(t, "2026-10-23"))

		if !errors.Is(err, errRepository) {
			t.Errorf("error = %v, se esperaba el error del repositorio", err)
		}
		if p != (project.Project{}) {
			t.Errorf("proyecto devuelto = %+v, se esperaba un proyecto vacío", p)
		}
		if got := mustGet(t, manager, 1).StartDate; !got.IsZero() {
			t.Errorf("fecha guardada = %v, se esperaba sin registrar", got)
		}
	})
}

// CA-18.8 (RN-18.10): registros de fecha, modificaciones y consultas
// simultáneos sobre el mismo proyecto no se pisan. Este test se verifica
// corriendo con `go test -race` (lo corre el CI).
func TestSetStartDate_RegistrosModificacionesYConsultasSimultaneosSinAccesosSinProteger(t *testing.T) {
	const registros = 20
	const modificaciones = 20
	const consultas = 20
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	dates := make([]time.Time, registros)
	for i := range registros {
		dates[i] = day(t, "2026-10-01").AddDate(0, 0, i)
	}
	names := make([]string, modificaciones)
	for i := range modificaciones {
		names[i] = fmt.Sprintf("Proyecto A v%d", i)
	}

	opErrs := make([]error, registros+modificaciones+consultas)
	var wg sync.WaitGroup
	for i := range registros {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, opErrs[i] = manager.SetStartDate(1, dates[i])
		}()
	}
	for i := range modificaciones {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, opErrs[registros+i] = manager.Update(1, names[i], "", "EN_CURSO")
		}()
	}
	for i := range consultas {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, opErrs[registros+modificaciones+i] = manager.Get(1)
		}()
	}
	wg.Wait()

	for _, err := range opErrs {
		if err != nil {
			t.Errorf("operación simultánea con error inesperado: %v", err)
		}
	}
	p := mustGet(t, manager, 1)
	if !slices.ContainsFunc(dates, p.StartDate.Equal) {
		t.Errorf("fecha final = %v, se esperaba la de uno de los registros", p.StartDate)
	}
	if !slices.Contains(names, p.Name) || p.Status != project.StatusInProgress {
		t.Errorf("proyecto final = %+v, se esperaban los datos de una de las modificaciones", p)
	}
}
