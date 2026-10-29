package project_test

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-41 · Registrar la fecha de fin del proyecto (Issue #115)
//
// Trazabilidad: CA-41.1, CA-41.2 y CA-41.4 a CA-41.11. Los casos de
// SetStartDate con fecha de fin (CA-41.8) van acá para no modificar los tests
// de HU-18.

// newProjectWithStartDate crea el proyecto 1 "Proyecto A" con la fecha de
// inicio start y devuelve el Manager y su repositorio.
func newProjectWithStartDate(t *testing.T, start string) (*project.Manager, *updateFailingRepository) {
	t.Helper()
	repo := &updateFailingRepository{MemoryRepository: project.NewMemoryRepository()}
	manager := project.NewManager(repo)
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	mustSetStartDate(t, manager, 1, day(t, start))
	return manager, repo
}

// mustSetEndDate registra la fecha de fin y falla el test si hay un error o
// si no quedó guardada.
func mustSetEndDate(t *testing.T, manager *project.Manager, id int, date time.Time) {
	t.Helper()
	if _, err := manager.SetEndDate(id, date); err != nil {
		t.Fatalf("SetEndDate(%d, %v) devolvió un error inesperado: %v", id, date, err)
	}
	if got := mustGet(t, manager, id).EndDate; !got.Equal(date) {
		t.Fatalf("SetEndDate(%d, %v): fecha de fin guardada = %v", id, date, got)
	}
}

// storeEndDate guarda la fecha de fin directamente en el repositorio, sin
// SetEndDate, para probar SetStartDate de forma aislada (CA-41.8).
func storeEndDate(t *testing.T, repo *updateFailingRepository, id int, date time.Time) {
	t.Helper()
	p, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID(%d) devolvió un error inesperado: %v", id, err)
	}
	p.EndDate = date
	if _, err := repo.MemoryRepository.Update(p); err != nil {
		t.Fatalf("guardar la fecha de fin de prueba: %v", err)
	}
}

// expectUnchanged verifica que el proyecto guardado siga igual a before.
func expectUnchanged(t *testing.T, manager *project.Manager, before project.Project) {
	t.Helper()
	if after := mustGet(t, manager, before.ID); after != before {
		t.Errorf("el proyecto guardado cambió: antes %+v, después %+v", before, after)
	}
}

// CA-41.1
func TestSetEndDate_RegistraLaFechaYConservaLosDemasDatos(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Sistema de ventas", "Gestión de ventas"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if _, err := manager.Update(1, "Sistema de ventas", "Gestión de ventas", "EN_CURSO"); err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	start := day(t, "2026-10-23")
	mustSetStartDate(t, manager, 1, start)
	end := day(t, "2026-12-18")

	p, err := manager.SetEndDate(1, end)

	if err != nil {
		t.Fatalf("SetEndDate() devolvió un error inesperado: %v", err)
	}
	want := project.Project{ID: 1, Name: "Sistema de ventas", Description: "Gestión de ventas", Status: project.StatusInProgress, StartDate: start, EndDate: end}
	if p != want {
		t.Errorf("proyecto devuelto = %+v, se esperaba %+v", p, want)
	}
	if stored := mustGet(t, manager, 1); stored != want {
		t.Errorf("proyecto guardado = %+v, se esperaba %+v", stored, want)
	}
}

// CA-41.1 (RN-41.9): vale en los tres estados del proyecto.
func TestSetEndDate_SeRegistraEnCualquierEstado(t *testing.T) {
	for _, status := range []project.Status{project.StatusPlanned, project.StatusInProgress, project.StatusFinished} {
		t.Run(string(status), func(t *testing.T) {
			manager, _ := newProjectWithStartDate(t, "2026-10-23")
			if _, err := manager.Update(1, "Proyecto A", "", string(status)); err != nil {
				t.Fatalf("Update() devolvió un error inesperado: %v", err)
			}
			end := day(t, "2026-12-18")

			p, err := manager.SetEndDate(1, end)

			if err != nil {
				t.Fatalf("SetEndDate() devolvió un error inesperado: %v", err)
			}
			if !p.EndDate.Equal(end) || p.Status != status {
				t.Errorf("proyecto = %+v, se esperaba la fecha de fin %v y el estado %q", p, end, status)
			}
		})
	}
}

// CA-41.1 y CA-41.2 (RN-41.9): el día siguiente al inicio, una fecha futura y
// un 29 de febrero se aceptan.
func TestSetEndDate_AceptaElDiaSiguienteFechasFuturasYDeAnioBisiesto(t *testing.T) {
	for _, date := range []string{"2026-10-24", "2030-12-31", "2028-02-29"} {
		t.Run(date, func(t *testing.T) {
			manager, _ := newProjectWithStartDate(t, "2026-10-23")

			_, err := manager.SetEndDate(1, day(t, date))

			if err != nil {
				t.Fatalf("SetEndDate() devolvió un error inesperado: %v", err)
			}
			if got := mustGet(t, manager, 1).EndDate; !got.Equal(day(t, date)) {
				t.Errorf("fecha de fin guardada = %v, se esperaba %s", got, date)
			}
		})
	}
}

// CA-41.2 (RN-41.5): el mismo día y los días anteriores se rechazan.
func TestSetEndDate_FechaIgualOAnteriorAlInicioDevuelveErrEndNotAfterStart(t *testing.T) {
	for _, date := range []string{"2026-10-23", "2026-10-22", "2020-01-01"} {
		t.Run(date, func(t *testing.T) {
			manager, _ := newProjectWithStartDate(t, "2026-10-23")
			before := mustGet(t, manager, 1)

			p, err := manager.SetEndDate(1, day(t, date))

			if !errors.Is(err, project.ErrEndNotAfterStart) {
				t.Fatalf("err = %v, se esperaba %v", err, project.ErrEndNotAfterStart)
			}
			if p != (project.Project{}) {
				t.Errorf("proyecto devuelto = %+v, se esperaba el valor cero", p)
			}
			expectUnchanged(t, manager, before)
		})
	}
}

// CA-41.4 (RN-41.1)
func TestSetEndDate_FechaAusenteDevuelveErrMissingEndDate(t *testing.T) {
	manager, _ := newProjectWithStartDate(t, "2026-10-23")
	before := mustGet(t, manager, 1)

	_, err := manager.SetEndDate(1, time.Time{})

	if !errors.Is(err, project.ErrMissingEndDate) {
		t.Fatalf("err = %v, se esperaba %v", err, project.ErrMissingEndDate)
	}
	expectUnchanged(t, manager, before)
}

// CA-41.4 (orden de validación): la fecha ausente se informa sin consultar el
// repositorio, aunque el proyecto no exista.
func TestSetEndDate_FechaAusenteEnProyectoInexistenteDevuelveErrMissingEndDate(t *testing.T) {
	manager := project.NewManager(failingRepository{getErr: errRepository})

	_, err := manager.SetEndDate(999, time.Time{})

	if !errors.Is(err, project.ErrMissingEndDate) {
		t.Errorf("err = %v, se esperaba %v", err, project.ErrMissingEndDate)
	}
}

// CA-41.5 (RN-41.3)
func TestSetEndDate_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	for _, id := range []int{999, 0, -1} {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			manager, _ := newProjectWithStartDate(t, "2026-10-23")

			_, err := manager.SetEndDate(id, day(t, "2026-12-18"))

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Errorf("err = %v, se esperaba %v", err, project.ErrProjectNotFound)
			}
		})
	}
}

// CA-41.6 (RN-41.4)
func TestSetEndDate_SinFechaDeInicioDevuelveErrStartDateNotRegistered(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Proyecto A", ""); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	before := mustGet(t, manager, 1)

	_, err := manager.SetEndDate(1, day(t, "2026-12-18"))

	if !errors.Is(err, project.ErrStartDateNotRegistered) {
		t.Fatalf("err = %v, se esperaba %v", err, project.ErrStartDateNotRegistered)
	}
	expectUnchanged(t, manager, before)
}

// CA-41.7 (RN-41.6)
func TestSetEndDate_VolverARegistrarReemplazaLaFecha(t *testing.T) {
	manager, _ := newProjectWithStartDate(t, "2026-10-23")
	mustSetEndDate(t, manager, 1, day(t, "2026-12-18"))

	p, err := manager.SetEndDate(1, day(t, "2027-01-29"))

	if err != nil {
		t.Fatalf("SetEndDate() devolvió un error inesperado: %v", err)
	}
	if !p.EndDate.Equal(day(t, "2027-01-29")) {
		t.Errorf("fecha de fin devuelta = %v, se esperaba 2027-01-29", p.EndDate)
	}
	if got := mustGet(t, manager, 1).EndDate; !got.Equal(day(t, "2027-01-29")) {
		t.Errorf("fecha de fin guardada = %v, se esperaba 2027-01-29", got)
	}
}

// CA-41.7: volver a registrar con una fecha inválida conserva la anterior.
func TestSetEndDate_VolverARegistrarUnaFechaInvalidaConservaLaAnterior(t *testing.T) {
	manager, _ := newProjectWithStartDate(t, "2026-10-23")
	mustSetEndDate(t, manager, 1, day(t, "2026-12-18"))
	before := mustGet(t, manager, 1)

	_, err := manager.SetEndDate(1, day(t, "2026-10-01"))

	if !errors.Is(err, project.ErrEndNotAfterStart) {
		t.Fatalf("err = %v, se esperaba %v", err, project.ErrEndNotAfterStart)
	}
	expectUnchanged(t, manager, before)
}

// CA-41.7 (RN-41.6): registrar la misma fecha es válido y deja el proyecto
// igual.
func TestSetEndDate_MismaFechaEsValida(t *testing.T) {
	manager, _ := newProjectWithStartDate(t, "2026-10-23")
	mustSetEndDate(t, manager, 1, day(t, "2026-12-18"))
	before := mustGet(t, manager, 1)

	p, err := manager.SetEndDate(1, day(t, "2026-12-18"))

	if err != nil {
		t.Fatalf("SetEndDate() devolvió un error inesperado: %v", err)
	}
	if p != before {
		t.Errorf("proyecto devuelto = %+v, se esperaba %+v", p, before)
	}
	expectUnchanged(t, manager, before)
}

// CA-41.8 (RN-41.7): con fecha de fin, un inicio el mismo día o después se
// rechaza. La fecha de fin se guarda con el repositorio para probar
// SetStartDate de forma aislada.
func TestSetStartDate_ConFechaDeFinRechazaUnInicioIgualOPosterior(t *testing.T) {
	for _, date := range []string{"2026-12-18", "2026-12-19", "2030-01-01"} {
		t.Run(date, func(t *testing.T) {
			manager, repo := newProjectWithStartDate(t, "2026-10-23")
			storeEndDate(t, repo, 1, day(t, "2026-12-18"))
			before := mustGet(t, manager, 1)

			_, err := manager.SetStartDate(1, day(t, date))

			if !errors.Is(err, project.ErrStartNotBeforeEnd) {
				t.Fatalf("err = %v, se esperaba %v", err, project.ErrStartNotBeforeEnd)
			}
			expectUnchanged(t, manager, before)
		})
	}
}

// CA-41.8: con fecha de fin, un inicio anterior se acepta y la fecha de fin se
// conserva.
func TestSetStartDate_ConFechaDeFinAceptaUnInicioAnterior(t *testing.T) {
	manager, repo := newProjectWithStartDate(t, "2026-10-23")
	end := day(t, "2026-12-18")
	storeEndDate(t, repo, 1, end)

	p, err := manager.SetStartDate(1, day(t, "2026-12-17"))

	if err != nil {
		t.Fatalf("SetStartDate() devolvió un error inesperado: %v", err)
	}
	if !p.StartDate.Equal(day(t, "2026-12-17")) || !p.EndDate.Equal(end) {
		t.Errorf("proyecto = %+v, se esperaba inicio 2026-12-17 y fin 2026-12-18", p)
	}
	if stored := mustGet(t, manager, 1); stored != p {
		t.Errorf("proyecto guardado = %+v, se esperaba %+v", stored, p)
	}
}

// CA-41.9 (RN-41.10)
func TestSetEndDate_ProyectoNuevoNoTieneFechaDeFin(t *testing.T) {
	manager, _ := newManager()

	p, err := manager.Create("Proyecto A", "")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if !p.EndDate.IsZero() || !mustGet(t, manager, 1).EndDate.IsZero() {
		t.Errorf("proyecto = %+v, se esperaba sin fecha de fin", p)
	}
}

// CA-41.9 (RN-41.8): Update (HU-02) y SetStartDate (HU-18) conservan la fecha
// de fin.
func TestSetEndDate_UpdateYSetStartDateConservanLaFechaDeFin(t *testing.T) {
	manager, _ := newProjectWithStartDate(t, "2026-10-23")
	end := day(t, "2026-12-18")
	mustSetEndDate(t, manager, 1, end)

	if _, err := manager.Update(1, "Proyecto B", "Nueva", "EN_CURSO"); err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	if got := mustGet(t, manager, 1).EndDate; !got.Equal(end) {
		t.Errorf("después de Update: fecha de fin = %v, se esperaba %v", got, end)
	}
	mustSetStartDate(t, manager, 1, day(t, "2026-11-02"))
	if got := mustGet(t, manager, 1).EndDate; !got.Equal(end) {
		t.Errorf("después de SetStartDate: fecha de fin = %v, se esperaba %v", got, end)
	}
}

// CA-41.10 (RN-41.11)
func TestSetEndDate_DevuelveUnaCopia(t *testing.T) {
	manager, _ := newProjectWithStartDate(t, "2026-10-23")
	end := day(t, "2026-12-18")

	p, err := manager.SetEndDate(1, end)
	if err != nil {
		t.Fatalf("SetEndDate() devolvió un error inesperado: %v", err)
	}
	p.EndDate = day(t, "2030-01-01")
	p.Name = "Modificado"

	stored := mustGet(t, manager, 1)
	if !stored.EndDate.Equal(end) || stored.Name != "Proyecto A" {
		t.Errorf("proyecto guardado = %+v, se esperaba fin %v y nombre \"Proyecto A\"", stored, end)
	}
}

// CA-41.11: si el repositorio falla al actualizar, se devuelve ese error y el
// proyecto no cambia.
func TestSetEndDate_ErrorDelRepositorioAlActualizar(t *testing.T) {
	manager, repo := newProjectWithStartDate(t, "2026-10-23")
	before := mustGet(t, manager, 1)
	repo.failUpdate = true

	_, err := manager.SetEndDate(1, day(t, "2026-12-18"))

	if !errors.Is(err, errRepository) {
		t.Fatalf("err = %v, se esperaba %v", err, errRepository)
	}
	repo.failUpdate = false
	expectUnchanged(t, manager, before)
}

// CA-41.11: si el repositorio falla al buscar, se devuelve ese error.
func TestSetEndDate_ErrorDelRepositorioAlBuscar(t *testing.T) {
	manager := project.NewManager(failingRepository{getErr: errRepository})

	_, err := manager.SetEndDate(1, day(t, "2026-12-18"))

	if !errors.Is(err, errRepository) {
		t.Errorf("err = %v, se esperaba %v", err, errRepository)
	}
}

// CA-41.11 (RN-41.13): registros simultáneos de las dos fechas, con
// modificaciones y consultas, no producen carreras (go test -race) y nunca
// dejan el fin igual o anterior al inicio. Las fechas de partida (inicio
// 2026-10-01 y fin 2026-12-31) hacen que el resultado de cada registro no
// dependa del orden: el inicio 2026-10-05 y el fin 2026-12-20 siempre son
// válidos, y el inicio 2027-01-15 y el fin 2026-09-01 nunca lo son.
func TestSetEndDate_RegistrosSimultaneosMantienenElOrdenDeLasFechas(t *testing.T) {
	manager, repo := newProjectWithStartDate(t, "2026-10-01")
	storeEndDate(t, repo, 1, day(t, "2026-12-31"))
	validStart, lateStart := day(t, "2026-10-05"), day(t, "2027-01-15")
	validEnd, earlyEnd := day(t, "2026-12-20"), day(t, "2026-09-01")
	const workers = 20

	var wg sync.WaitGroup
	errs := make(chan string, 6*workers)
	report := func(op string, err, want error) {
		if (want == nil && err != nil) || (want != nil && !errors.Is(err, want)) {
			errs <- fmt.Sprintf("%s: err = %v, se esperaba %v", op, err, want)
		}
	}
	for i := 0; i < workers; i++ {
		wg.Add(6)
		go func() {
			defer wg.Done()
			_, err := manager.SetStartDate(1, validStart)
			report("SetStartDate(2026-10-05)", err, nil)
		}()
		go func() {
			defer wg.Done()
			_, err := manager.SetStartDate(1, lateStart)
			report("SetStartDate(2027-01-15)", err, project.ErrStartNotBeforeEnd)
		}()
		go func() {
			defer wg.Done()
			_, err := manager.SetEndDate(1, validEnd)
			report("SetEndDate(2026-12-20)", err, nil)
		}()
		go func() {
			defer wg.Done()
			_, err := manager.SetEndDate(1, earlyEnd)
			report("SetEndDate(2026-09-01)", err, project.ErrEndNotAfterStart)
		}()
		go func() {
			defer wg.Done()
			_, err := manager.Update(1, "Proyecto A", "", "EN_CURSO")
			report("Update", err, nil)
		}()
		go func() {
			defer wg.Done()
			_, err := manager.Get(1)
			report("Get", err, nil)
		}()
	}
	wg.Wait()
	close(errs)
	for msg := range errs {
		t.Error(msg)
	}

	final := mustGet(t, manager, 1)
	if !final.StartDate.Equal(validStart) || !final.EndDate.Equal(validEnd) {
		t.Errorf("fechas finales: inicio %v y fin %v, se esperaban %v y %v", final.StartDate, final.EndDate, validStart, validEnd)
	}
	if !final.EndDate.After(final.StartDate) {
		t.Errorf("la fecha de fin %v no es posterior a la de inicio %v", final.EndDate, final.StartDate)
	}
	if final.Status != project.StatusInProgress {
		t.Errorf("estado = %q, se esperaba EN_CURSO", final.Status)
	}
}
