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

// HU-19 · Modificar las fechas del proyecto (Issue #37)
//
// Trazabilidad: CA-19.1 a CA-19.7 y los casos límite del Manager. El formato
// AAAA-MM-DD y CA-19.8 se prueban en internal/httpapi.

// newProjectWithDates crea el proyecto 1 "Proyecto A" con las fechas de
// inicio start y de fin end, registradas con las operaciones de HU-18 y
// HU-41, y devuelve el Manager y su repositorio.
func newProjectWithDates(t *testing.T, start, end string) (*project.Manager, *updateFailingRepository) {
	t.Helper()
	manager, repo := newProjectWithStartDate(t, start)
	mustSetEndDate(t, manager, 1, day(t, end))
	return manager, repo
}

// mustSetDates modifica las dos fechas y falla el test si hay un error o si
// no quedaron guardadas.
func mustSetDates(t *testing.T, manager *project.Manager, id int, start, end time.Time) {
	t.Helper()
	if _, err := manager.SetDates(id, start, end); err != nil {
		t.Fatalf("SetDates(%d, %v, %v) devolvió un error inesperado: %v", id, start, end, err)
	}
	expectStoredDates(t, manager, id, start, end)
}

// expectStoredDates verifica las dos fechas del proyecto guardado.
func expectStoredDates(t *testing.T, manager *project.Manager, id int, start, end time.Time) {
	t.Helper()
	p := mustGet(t, manager, id)
	if !p.StartDate.Equal(start) || !p.EndDate.Equal(end) {
		t.Fatalf("fechas guardadas: inicio %v y fin %v, se esperaban %v y %v", p.StartDate, p.EndDate, start, end)
	}
}

// CA-19.1: postergar las dos fechas, con el inicio nuevo posterior al fin
// viejo, conserva el nombre, la descripción y el estado.
func TestSetDates_PostergaLasDosFechasYConservaLosDemasDatos(t *testing.T) {
	manager, _ := newManager()
	if _, err := manager.Create("Sistema de ventas", "Gestión de ventas"); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if _, err := manager.Update(1, "Sistema de ventas", "Gestión de ventas", "EN_CURSO"); err != nil {
		t.Fatalf("Update() devolvió un error inesperado: %v", err)
	}
	mustSetStartDate(t, manager, 1, day(t, "2026-10-23"))
	mustSetEndDate(t, manager, 1, day(t, "2026-12-18"))
	start, end := day(t, "2027-01-11"), day(t, "2027-03-19")

	p, err := manager.SetDates(1, start, end)

	if err != nil {
		t.Fatalf("SetDates() devolvió un error inesperado: %v", err)
	}
	want := project.Project{ID: 1, Name: "Sistema de ventas", Description: "Gestión de ventas", Status: project.StatusInProgress, StartDate: start, EndDate: end}
	if p != want {
		t.Errorf("proyecto devuelto = %+v, se esperaba %+v", p, want)
	}
	if stored := mustGet(t, manager, 1); stored != want {
		t.Errorf("proyecto guardado = %+v, se esperaba %+v", stored, want)
	}
}

// CA-19.1: adelantar las dos fechas, con el fin nuevo anterior al inicio
// viejo.
func TestSetDates_AdelantaLasDosFechas(t *testing.T) {
	manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")
	start, end := day(t, "2026-06-01"), day(t, "2026-08-31")

	p, err := manager.SetDates(1, start, end)

	if err != nil {
		t.Fatalf("SetDates() devolvió un error inesperado: %v", err)
	}
	if !p.StartDate.Equal(start) || !p.EndDate.Equal(end) {
		t.Errorf("proyecto devuelto = %+v, se esperaba inicio %v y fin %v", p, start, end)
	}
	expectStoredDates(t, manager, 1, start, end)
}

// CA-19.1 (RN-19.8): vale en los tres estados del proyecto.
func TestSetDates_SeModificaEnCualquierEstado(t *testing.T) {
	for _, status := range []project.Status{project.StatusPlanned, project.StatusInProgress, project.StatusFinished} {
		t.Run(string(status), func(t *testing.T) {
			manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")
			if _, err := manager.Update(1, "Proyecto A", "", string(status)); err != nil {
				t.Fatalf("Update() devolvió un error inesperado: %v", err)
			}
			start, end := day(t, "2027-01-11"), day(t, "2027-03-19")

			p, err := manager.SetDates(1, start, end)

			if err != nil {
				t.Fatalf("SetDates() devolvió un error inesperado: %v", err)
			}
			if p.Status != status {
				t.Errorf("estado = %q, se esperaba %q", p.Status, status)
			}
			expectStoredDates(t, manager, 1, start, end)
		})
	}
}

// CA-19.1: se puede repetir una de las dos fechas y cambiar solo la otra.
func TestSetDates_CambiaSoloUnaDeLasDosFechas(t *testing.T) {
	casos := map[string][2]string{
		"solo el fin":    {"2026-10-23", "2027-01-29"},
		"solo el inicio": {"2026-11-02", "2026-12-18"},
	}
	for caseName, dates := range casos {
		t.Run(caseName, func(t *testing.T) {
			manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")

			mustSetDates(t, manager, 1, day(t, dates[0]), day(t, dates[1]))
		})
	}
}

// CA-19.1 (RN-19.9): las mismas fechas son válidas y dejan el proyecto igual.
func TestSetDates_MismasFechasEsValido(t *testing.T) {
	manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")
	before := mustGet(t, manager, 1)

	p, err := manager.SetDates(1, before.StartDate, before.EndDate)

	if err != nil {
		t.Fatalf("SetDates() devolvió un error inesperado: %v", err)
	}
	if p != before {
		t.Errorf("proyecto devuelto = %+v, se esperaba %+v", p, before)
	}
	expectUnchanged(t, manager, before)
}

// CA-19.1 y CA-19.2: el fin al día siguiente del inicio y un 29 de febrero se
// aceptan.
func TestSetDates_AceptaElDiaSiguienteYElAnioBisiesto(t *testing.T) {
	for _, dates := range [][2]string{
		{"2027-01-11", "2027-01-12"},
		{"2028-02-29", "2028-03-01"},
		{"2028-01-01", "2028-02-29"},
	} {
		t.Run(dates[0]+"_"+dates[1], func(t *testing.T) {
			manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")

			mustSetDates(t, manager, 1, day(t, dates[0]), day(t, dates[1]))
		})
	}
}

// CA-19.2 (RN-19.5): un fin nuevo igual o anterior al inicio nuevo se rechaza
// y no cambia ninguna de las dos fechas.
func TestSetDates_FinIgualOAnteriorAlInicioDevuelveErrEndNotAfterStart(t *testing.T) {
	for _, end := range []string{"2027-01-11", "2027-01-10", "2026-12-31"} {
		t.Run(end, func(t *testing.T) {
			manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")
			before := mustGet(t, manager, 1)

			p, err := manager.SetDates(1, day(t, "2027-01-11"), day(t, end))

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

// CA-19.2 (RN-19.5): el par nuevo no se compara con las fechas guardadas; un
// inicio nuevo posterior al fin viejo es válido si el fin nuevo es posterior.
func TestSetDates_NoComparaConLasFechasGuardadas(t *testing.T) {
	manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")

	mustSetDates(t, manager, 1, day(t, "2026-12-19"), day(t, "2026-12-20"))
}

// CA-19.5 (RN-19.1): sin fecha de inicio, con o sin fecha de fin.
func TestSetDates_FechaDeInicioAusenteDevuelveErrMissingStartDate(t *testing.T) {
	casos := map[string]time.Time{
		"con fin": day(t, "2027-03-19"),
		"sin fin": {},
	}
	for caseName, end := range casos {
		t.Run(caseName, func(t *testing.T) {
			manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")
			before := mustGet(t, manager, 1)

			_, err := manager.SetDates(1, time.Time{}, end)

			if !errors.Is(err, project.ErrMissingStartDate) {
				t.Fatalf("err = %v, se esperaba %v", err, project.ErrMissingStartDate)
			}
			expectUnchanged(t, manager, before)
		})
	}
}

// CA-19.5 (RN-19.1)
func TestSetDates_FechaDeFinAusenteDevuelveErrMissingEndDate(t *testing.T) {
	manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")
	before := mustGet(t, manager, 1)

	_, err := manager.SetDates(1, day(t, "2027-01-11"), time.Time{})

	if !errors.Is(err, project.ErrMissingEndDate) {
		t.Fatalf("err = %v, se esperaba %v", err, project.ErrMissingEndDate)
	}
	expectUnchanged(t, manager, before)
}

// CA-19.5 (orden de validación): las fechas ausentes se informan sin
// consultar el repositorio, aunque el proyecto no exista.
func TestSetDates_FechasAusentesSeInformanSinConsultarElRepositorio(t *testing.T) {
	manager := project.NewManager(failingRepository{getErr: errRepository})

	if _, err := manager.SetDates(999, time.Time{}, time.Time{}); !errors.Is(err, project.ErrMissingStartDate) {
		t.Errorf("sin las dos fechas: err = %v, se esperaba %v", err, project.ErrMissingStartDate)
	}
	if _, err := manager.SetDates(999, day(t, "2027-01-11"), time.Time{}); !errors.Is(err, project.ErrMissingEndDate) {
		t.Errorf("sin fecha de fin: err = %v, se esperaba %v", err, project.ErrMissingEndDate)
	}
}

// CA-19.3 (RN-19.3)
func TestSetDates_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	for _, id := range []int{999, 0, -1} {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")

			_, err := manager.SetDates(id, day(t, "2027-01-11"), day(t, "2027-03-19"))

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Errorf("err = %v, se esperaba %v", err, project.ErrProjectNotFound)
			}
		})
	}
}

// CA-19.3 (D5): la existencia se valida antes que la relación del par.
func TestSetDates_ParInvalidoEnProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")

	_, err := manager.SetDates(999, day(t, "2027-03-19"), day(t, "2027-01-11"))

	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Errorf("err = %v, se esperaba %v", err, project.ErrProjectNotFound)
	}
}

// CA-19.4 (RN-19.4): un proyecto sin alguna de las dos fechas no se modifica,
// aunque el par sea inválido (D5).
func TestSetDates_ProyectoSinLasDosFechasDevuelveErrDatesNotRegistered(t *testing.T) {
	casos := map[string]func(t *testing.T, manager *project.Manager, repo *updateFailingRepository){
		"sin fechas": func(t *testing.T, manager *project.Manager, repo *updateFailingRepository) {},
		"solo con inicio": func(t *testing.T, manager *project.Manager, repo *updateFailingRepository) {
			mustSetStartDate(t, manager, 1, day(t, "2026-10-23"))
		},
		"solo con fin": func(t *testing.T, manager *project.Manager, repo *updateFailingRepository) {
			storeEndDate(t, repo, 1, day(t, "2026-12-18"))
		},
	}
	pairs := map[string][2]string{
		"par válido":   {"2027-01-11", "2027-03-19"},
		"par inválido": {"2027-03-19", "2027-01-11"},
	}
	for caseName, setup := range casos {
		for pairName, pair := range pairs {
			t.Run(caseName+"_"+pairName, func(t *testing.T) {
				repo := &updateFailingRepository{MemoryRepository: project.NewMemoryRepository()}
				manager := project.NewManager(repo)
				if _, err := manager.Create("Proyecto A", ""); err != nil {
					t.Fatalf("Create() devolvió un error inesperado: %v", err)
				}
				setup(t, manager, repo)
				before := mustGet(t, manager, 1)

				_, err := manager.SetDates(1, day(t, pair[0]), day(t, pair[1]))

				if !errors.Is(err, project.ErrDatesNotRegistered) {
					t.Fatalf("err = %v, se esperaba %v", err, project.ErrDatesNotRegistered)
				}
				expectUnchanged(t, manager, before)
			})
		}
	}
}

// CA-19.6 (RN-19.10)
func TestSetDates_DevuelveUnaCopia(t *testing.T) {
	manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")
	start, end := day(t, "2027-01-11"), day(t, "2027-03-19")

	p, err := manager.SetDates(1, start, end)
	if err != nil {
		t.Fatalf("SetDates() devolvió un error inesperado: %v", err)
	}
	p.StartDate = day(t, "2030-01-01")
	p.EndDate = day(t, "2030-12-31")
	p.Name = "Modificado"

	stored := mustGet(t, manager, 1)
	if !stored.StartDate.Equal(start) || !stored.EndDate.Equal(end) || stored.Name != "Proyecto A" {
		t.Errorf("proyecto guardado = %+v, se esperaba inicio %v, fin %v y nombre \"Proyecto A\"", stored, start, end)
	}
}

// CA-19.6 (RN-19.6): si el repositorio falla al actualizar, se devuelve ese
// error y las dos fechas quedan como estaban.
func TestSetDates_ErrorDelRepositorioAlActualizar(t *testing.T) {
	manager, repo := newProjectWithDates(t, "2026-10-23", "2026-12-18")
	before := mustGet(t, manager, 1)
	repo.failUpdate = true

	_, err := manager.SetDates(1, day(t, "2027-01-11"), day(t, "2027-03-19"))

	if !errors.Is(err, errRepository) {
		t.Fatalf("err = %v, se esperaba %v", err, errRepository)
	}
	repo.failUpdate = false
	expectUnchanged(t, manager, before)
}

// CA-19.6: si el repositorio falla al buscar, se devuelve ese error.
func TestSetDates_ErrorDelRepositorioAlBuscar(t *testing.T) {
	manager := project.NewManager(failingRepository{getErr: errRepository})

	_, err := manager.SetDates(1, day(t, "2027-01-11"), day(t, "2027-03-19"))

	if !errors.Is(err, errRepository) {
		t.Errorf("err = %v, se esperaba %v", err, errRepository)
	}
}

// Caso límite: después de SetDates, SetStartDate y SetEndDate siguen
// validando contra la fecha guardada, como en HU-18 y HU-41.
func TestSetDates_LasOperacionesDeHU18YHU41ValidanContraLasFechasNuevas(t *testing.T) {
	manager, _ := newProjectWithDates(t, "2026-10-23", "2026-12-18")
	mustSetDates(t, manager, 1, day(t, "2027-01-11"), day(t, "2027-03-19"))

	if _, err := manager.SetStartDate(1, day(t, "2027-03-19")); !errors.Is(err, project.ErrStartNotBeforeEnd) {
		t.Errorf("SetStartDate(2027-03-19): err = %v, se esperaba %v", err, project.ErrStartNotBeforeEnd)
	}
	if _, err := manager.SetEndDate(1, day(t, "2027-01-11")); !errors.Is(err, project.ErrEndNotAfterStart) {
		t.Errorf("SetEndDate(2027-01-11): err = %v, se esperaba %v", err, project.ErrEndNotAfterStart)
	}
	expectStoredDates(t, manager, 1, day(t, "2027-01-11"), day(t, "2027-03-19"))
}

// CA-19.7 (RN-19.11): SetDates, SetStartDate, SetEndDate, Update y Get
// simultáneos no producen carreras (go test -race) y nunca dejan el fin igual
// o anterior al inicio. Las fechas de partida (inicio 2026-10-01 y fin
// 2026-12-31) y las elegidas hacen que el resultado de cada operación no
// dependa del orden: con el inicio en 2026-10-01, 2026-10-05 o 2026-10-10 y el
// fin en 2026-12-20, 2026-12-25 o 2026-12-31, el par 2026-10-10/2026-12-20, el
// inicio 2026-10-05 y el fin 2026-12-25 siempre son válidos, y el par
// 2027-02-01/2027-01-01, el inicio 2027-01-15 y el fin 2026-09-01 nunca lo
// son.
func TestSetDates_ModificacionesSimultaneasMantienenElOrdenDeLasFechas(t *testing.T) {
	manager, _ := newProjectWithDates(t, "2026-10-01", "2026-12-31")
	pairStart, pairEnd := day(t, "2026-10-10"), day(t, "2026-12-20")
	badPairStart, badPairEnd := day(t, "2027-02-01"), day(t, "2027-01-01")
	validStart, lateStart := day(t, "2026-10-05"), day(t, "2027-01-15")
	validEnd, earlyEnd := day(t, "2026-12-25"), day(t, "2026-09-01")
	const workers = 20

	var wg sync.WaitGroup
	errs := make(chan string, 8*workers)
	report := func(op string, err, want error) {
		if (want == nil && err != nil) || (want != nil && !errors.Is(err, want)) {
			errs <- fmt.Sprintf("%s: err = %v, se esperaba %v", op, err, want)
		}
	}
	for i := 0; i < workers; i++ {
		wg.Add(8)
		go func() {
			defer wg.Done()
			_, err := manager.SetDates(1, pairStart, pairEnd)
			report("SetDates(2026-10-10, 2026-12-20)", err, nil)
		}()
		go func() {
			defer wg.Done()
			_, err := manager.SetDates(1, badPairStart, badPairEnd)
			report("SetDates(2027-02-01, 2027-01-01)", err, project.ErrEndNotAfterStart)
		}()
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
			report("SetEndDate(2026-12-25)", err, nil)
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
	if !final.StartDate.Equal(pairStart) && !final.StartDate.Equal(validStart) {
		t.Errorf("fecha de inicio final = %v, se esperaba %v o %v", final.StartDate, pairStart, validStart)
	}
	if !final.EndDate.Equal(pairEnd) && !final.EndDate.Equal(validEnd) {
		t.Errorf("fecha de fin final = %v, se esperaba %v o %v", final.EndDate, pairEnd, validEnd)
	}
	if !final.EndDate.After(final.StartDate) {
		t.Errorf("la fecha de fin %v no es posterior a la de inicio %v", final.EndDate, final.StartDate)
	}
	if final.Status != project.StatusInProgress {
		t.Errorf("estado = %q, se esperaba EN_CURSO", final.Status)
	}
}
