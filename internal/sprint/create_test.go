package sprint

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-10 · Crear sprint con nombre, fechas y Sprint Goal (Issue #24)

const (
	validName = "Sprint 1"
	validGoal = "Entregar el MVP"
)

// validStart y validEnd son fechas válidas de un sprint.
var (
	validStart = time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	validEnd   = time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
)

// newTestManager arma un Manager de sprints con un repositorio y un Manager
// de proyectos en memoria vacíos. Desde HU-11 el Manager también recibe el
// Manager de ítems del backlog, aunque HU-10 no lo use.
func newTestManager() (*Manager, *MemoryRepository, *project.Manager, *project.MemoryRepository) {
	projectRepo := project.NewMemoryRepository()
	projects := project.NewManager(projectRepo)
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	repo := NewMemoryRepository()
	return NewManager(repo, projects, items), repo, projects, projectRepo
}

// createTestProject crea un proyecto con ese nombre y devuelve su ID.
func createTestProject(t *testing.T, projects *project.Manager, name string) int {
	t.Helper()
	p, err := projects.Create(name, "")
	if err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	return p.ID
}

// savedSprint devuelve el sprint guardado con ese ID, mirando el repositorio
// en memoria directamente (los tests son internos al paquete, ver SDD).
func savedSprint(t *testing.T, repo *MemoryRepository, id int) Sprint {
	t.Helper()
	for _, sprint := range repo.sprints {
		if sprint.ID == id {
			return sprint
		}
	}
	t.Fatalf("no hay ningún sprint guardado con el ID %d", id)
	return Sprint{}
}

// CA-10.1, RN-10.1, RN-10.2 y RN-10.7
func TestCreate_SprintValidoQuedaConID1EstadoAbiertoYCamposSinEspacios(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	created, err := manager.Create(projectID, "  Sprint 1  ", validStart, validEnd, "  Entregar el MVP  ")

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", created.ID)
	}
	if created.ProjectID != projectID {
		t.Errorf("ProjectID = %d, se esperaba %d", created.ProjectID, projectID)
	}
	if created.Name != validName {
		t.Errorf("Name = %q, se esperaba %q", created.Name, validName)
	}
	if created.Goal != validGoal {
		t.Errorf("Goal = %q, se esperaba %q", created.Goal, validGoal)
	}
	if created.Status != StatusOpen {
		t.Errorf("Status = %q, se esperaba %q", created.Status, StatusOpen)
	}
	if !created.StartDate.Equal(validStart) {
		t.Errorf("StartDate = %s, se esperaba %s", created.StartDate, validStart)
	}
	if !created.EndDate.Equal(validEnd) {
		t.Errorf("EndDate = %s, se esperaba %s", created.EndDate, validEnd)
	}

	if saved := savedSprint(t, repo, created.ID); !reflect.DeepEqual(saved, created) {
		t.Errorf("sprint guardado = %+v, se esperaba %+v", saved, created)
	}
}

// RN-10.6
func TestCreate_IDsCorrelativosYGlobalesEntreProyectos(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	firstProject := createTestProject(t, projects, "Sistema de ventas")
	secondProject := createTestProject(t, projects, "Portal de clientes")

	// RN-10.1: el mismo nombre en otro proyecto se permite.
	first, err := manager.Create(firstProject, validName, validStart, validEnd, validGoal)
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	second, err := manager.Create(secondProject, validName, validStart, validEnd, validGoal)
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	if first.ID != 1 || second.ID != 2 {
		t.Errorf("IDs = %d y %d, se esperaban 1 y 2", first.ID, second.ID)
	}
	if len(repo.sprints) != 2 {
		t.Errorf("hay %d sprints guardados, se esperaban 2", len(repo.sprints))
	}
}

// CA-10.4, RN-10.1
func TestCreate_NombreVacioDevuelveErrEmptyName(t *testing.T) {
	cases := map[string]string{
		"vacío":              "",
		"solo espacios":      "   ",
		"solo tabs y saltos": "\t\n ",
	}
	for caseName, name := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")

			_, err := manager.Create(projectID, name, validStart, validEnd, validGoal)

			if !errors.Is(err, ErrEmptyName) {
				t.Errorf("error = %v, se esperaba ErrEmptyName", err)
			}
			if len(repo.sprints) != 0 {
				t.Errorf("hay %d sprints guardados, se esperaban 0", len(repo.sprints))
			}
		})
	}
}

// CA-10.4, RN-10.2
func TestCreate_SprintGoalVacioDevuelveErrEmptyGoal(t *testing.T) {
	cases := map[string]string{
		"vacío":              "",
		"solo espacios":      "   ",
		"solo tabs y saltos": "\t\n ",
	}
	for caseName, goal := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")

			_, err := manager.Create(projectID, validName, validStart, validEnd, goal)

			if !errors.Is(err, ErrEmptyGoal) {
				t.Errorf("error = %v, se esperaba ErrEmptyGoal", err)
			}
			if len(repo.sprints) != 0 {
				t.Errorf("hay %d sprints guardados, se esperaban 0", len(repo.sprints))
			}
		})
	}
}

// CA-10.2, RN-10.3
func TestCreate_FechaInicioAusenteDevuelveErrMissingStartDate(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	_, err := manager.Create(projectID, validName, time.Time{}, validEnd, validGoal)

	if !errors.Is(err, ErrMissingStartDate) {
		t.Errorf("error = %v, se esperaba ErrMissingStartDate", err)
	}
	if len(repo.sprints) != 0 {
		t.Errorf("hay %d sprints guardados, se esperaban 0", len(repo.sprints))
	}
}

// CA-10.2, RN-10.3
func TestCreate_FechaFinAusenteDevuelveErrMissingEndDate(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	_, err := manager.Create(projectID, validName, validStart, time.Time{}, validGoal)

	if !errors.Is(err, ErrMissingEndDate) {
		t.Errorf("error = %v, se esperaba ErrMissingEndDate", err)
	}
	if len(repo.sprints) != 0 {
		t.Errorf("hay %d sprints guardados, se esperaban 0", len(repo.sprints))
	}
}

// CA-10.2, RN-10.3
func TestCreate_FinNoPosteriorAlInicioDevuelveErrEndNotAfterStart(t *testing.T) {
	cases := map[string]time.Time{
		"fin igual al inicio":    validStart,
		"fin anterior al inicio": validStart.AddDate(0, 0, -1),
	}
	for caseName, end := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")

			_, err := manager.Create(projectID, validName, validStart, end, validGoal)

			if !errors.Is(err, ErrEndNotAfterStart) {
				t.Errorf("error = %v, se esperaba ErrEndNotAfterStart", err)
			}
			if len(repo.sprints) != 0 {
				t.Errorf("hay %d sprints guardados, se esperaban 0", len(repo.sprints))
			}
		})
	}
}

// CA-10.5
func TestCreate_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
	cases := map[string]int{
		"ID inexistente": 999,
		"ID cero":        0,
		"ID negativo":    -1,
	}
	for caseName, projectID := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			createTestProject(t, projects, "Sistema de ventas")

			_, err := manager.Create(projectID, validName, validStart, validEnd, validGoal)

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
			}
			if len(repo.sprints) != 0 {
				t.Errorf("hay %d sprints guardados, se esperaban 0", len(repo.sprints))
			}
		})
	}
}

// CA-10.5
func TestCreate_SinProyectosCreadosDevuelveErrProjectNotFound(t *testing.T) {
	manager, repo, _, _ := newTestManager()

	_, err := manager.Create(1, validName, validStart, validEnd, validGoal)

	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
	}
	if len(repo.sprints) != 0 {
		t.Errorf("hay %d sprints guardados, se esperaban 0", len(repo.sprints))
	}
}

// RN-10.4
func TestCreate_ProyectoFinalizadoAceptaSprints(t *testing.T) {
	manager, repo, _, projectRepo := newTestManager()

	p, err := projectRepo.Save(project.Project{Name: "Proyecto cerrado", Status: project.StatusFinished})
	if err != nil {
		t.Fatalf("guardar el proyecto finalizado de prueba: %v", err)
	}

	created, err := manager.Create(p.ID, validName, validStart, validEnd, validGoal)

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if created.Status != StatusOpen {
		t.Errorf("Status = %q, se esperaba %q", created.Status, StatusOpen)
	}
	if len(repo.sprints) != 1 {
		t.Errorf("hay %d sprints guardados, se esperaba 1", len(repo.sprints))
	}
}

// CA-10.3, RN-10.5
func TestCreate_SegundoSprintAbiertoEnElMismoProyectoDevuelveErrOpenSprintExists(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	first, err := manager.Create(projectID, validName, validStart, validEnd, validGoal)
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	_, err = manager.Create(projectID, "Sprint 2", validStart, validEnd, "Otro objetivo")

	if !errors.Is(err, ErrOpenSprintExists) {
		t.Errorf("error = %v, se esperaba ErrOpenSprintExists", err)
	}
	if len(repo.sprints) != 1 {
		t.Errorf("hay %d sprints guardados, se esperaba 1", len(repo.sprints))
	}
	if saved := savedSprint(t, repo, first.ID); saved.Status != StatusOpen {
		t.Errorf("el primer sprint cambió de estado: %q", saved.Status)
	}
}

// RN-10.5
func TestCreate_SprintAbiertoEnOtroProyectoNoImpide(t *testing.T) {
	manager, _, projects, _ := newTestManager()
	firstProject := createTestProject(t, projects, "Sistema de ventas")
	secondProject := createTestProject(t, projects, "Portal de clientes")

	if _, err := manager.Create(firstProject, validName, validStart, validEnd, validGoal); err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	second, err := manager.Create(secondProject, validName, validStart, validEnd, validGoal)

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if second.Status != StatusOpen {
		t.Errorf("Status = %q, se esperaba %q", second.Status, StatusOpen)
	}
}

// RN-10.1
func TestCreate_NombreRepetidoEnElMismoProyectoConElAnteriorCerradoSePermite(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	if _, err := repo.Save(Sprint{
		ProjectID: projectID,
		Name:      validName,
		StartDate: validStart,
		EndDate:   validEnd,
		Goal:      validGoal,
		Status:    StatusClosed,
	}); err != nil {
		t.Fatalf("guardar el sprint cerrado de prueba: %v", err)
	}

	created, err := manager.Create(projectID, validName, validStart, validEnd, validGoal)

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if created.ID != 2 {
		t.Errorf("ID = %d, se esperaba 2", created.ID)
	}
	if created.Status != StatusOpen {
		t.Errorf("Status = %q, se esperaba %q", created.Status, StatusOpen)
	}
}

// Orden de validación de la SDD
func TestCreate_InformaElPrimerErrorSegunElOrdenDeValidacion(t *testing.T) {
	cases := []struct {
		name   string
		sprint string
		goal   string
		start  time.Time
		end    time.Time
		want   error
	}{
		{"el nombre antes que el sprint goal, las fechas y el proyecto", "   ", "   ", time.Time{}, time.Time{}, ErrEmptyName},
		{"el sprint goal antes que las fechas y el proyecto", validName, "   ", time.Time{}, time.Time{}, ErrEmptyGoal},
		{"la fecha de inicio antes que la de fin y el proyecto", validName, validGoal, time.Time{}, time.Time{}, ErrMissingStartDate},
		{"la fecha de fin antes que la relación entre fechas y el proyecto", validName, validGoal, validStart, time.Time{}, ErrMissingEndDate},
		{"la relación entre fechas antes que el proyecto", validName, validGoal, validStart, validStart.AddDate(0, 0, -1), ErrEndNotAfterStart},
		{"el proyecto antes que el sprint abierto", validName, validGoal, validStart, validEnd, project.ErrProjectNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			createTestProject(t, projects, "Sistema de ventas")

			_, err := manager.Create(999, tc.sprint, tc.start, tc.end, tc.goal)

			if !errors.Is(err, tc.want) {
				t.Errorf("error = %v, se esperaba %v", err, tc.want)
			}
			if len(repo.sprints) != 0 {
				t.Errorf("hay %d sprints guardados, se esperaban 0", len(repo.sprints))
			}
		})
	}
}

// CA-10.7, RN-10.8
func TestCreate_ModificarElSprintDevueltoNoAlteraElGuardado(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	created, err := manager.Create(projectID, validName, validStart, validEnd, validGoal)
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	created.ID = 999
	created.ProjectID = 999
	created.Name = "Otro nombre"
	created.Goal = "Otro objetivo"
	created.Status = StatusClosed
	created.StartDate = validStart.AddDate(0, 0, -7)
	created.EndDate = validEnd.AddDate(0, 0, -7)

	saved := savedSprint(t, repo, 1)
	if saved.Name != validName || saved.Goal != validGoal || saved.Status != StatusOpen {
		t.Errorf("el sprint guardado cambió: %+v", saved)
	}
	if !saved.StartDate.Equal(validStart) || !saved.EndDate.Equal(validEnd) {
		t.Errorf("las fechas guardadas cambiaron: %s y %s", saved.StartDate, saved.EndDate)
	}
	if saved.ProjectID != projectID {
		t.Errorf("ProjectID guardado = %d, se esperaba %d", saved.ProjectID, projectID)
	}
}

// CA-10.6
func TestCreate_CreacionesSimultaneasEnElMismoProyectoDejanUnSoloSprintAbierto(t *testing.T) {
	const attempts = 50
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = manager.Create(projectID, fmt.Sprintf("Sprint %d", i), validStart, validEnd, validGoal)
		}()
	}
	wg.Wait()

	successes := 0
	openErrs := 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrOpenSprintExists):
			openErrs++
		}
	}
	if successes != 1 {
		t.Errorf("tuvieron éxito %d creaciones, se esperaba exactamente 1", successes)
	}
	if openErrs != attempts-1 {
		t.Errorf("fallaron %d creaciones con ErrOpenSprintExists, se esperaban %d", openErrs, attempts-1)
	}
	if len(repo.sprints) != 1 {
		t.Fatalf("hay %d sprints guardados, se esperaba 1", len(repo.sprints))
	}
	if repo.sprints[0].Status != StatusOpen {
		t.Errorf("el sprint guardado quedó en estado %q, se esperaba %q", repo.sprints[0].Status, StatusOpen)
	}
}

// CA-10.6
func TestCreate_CreacionesSimultaneasEnProyectosDistintosSeCreanTodas(t *testing.T) {
	const attempts = 50
	manager, repo, projects, _ := newTestManager()
	projectIDs := make([]int, attempts)
	for i := range attempts {
		projectIDs[i] = createTestProject(t, projects, fmt.Sprintf("Proyecto %d", i))
	}

	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = manager.Create(projectIDs[i], validName, validStart, validEnd, validGoal)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("la creación %d falló: %v", i+1, err)
		}
	}
	if len(repo.sprints) != attempts {
		t.Fatalf("hay %d sprints guardados, se esperaban %d", len(repo.sprints), attempts)
	}
	ids := make(map[int]bool)
	for _, sprint := range repo.sprints {
		if ids[sprint.ID] {
			t.Errorf("el ID %d está repetido", sprint.ID)
		}
		ids[sprint.ID] = true
	}
}

// failingRepository es un Repository que falla al consultar el sprint
// abierto, al guardar o al buscar por ID (HU-11).
type failingRepository struct {
	openErr    error
	saveErr    error
	getByIDErr error
	updateErr  error
}

func (r failingRepository) OpenByProject(projectID int) (Sprint, bool, error) {
	if r.openErr != nil {
		return Sprint{}, false, r.openErr
	}
	return Sprint{}, false, nil
}

func (r failingRepository) Save(sprint Sprint) (Sprint, error) {
	return Sprint{}, r.saveErr
}

// GetByID falla si se configuró un error de búsqueda (HU-11).
func (r failingRepository) GetByID(id int) (Sprint, error) {
	return Sprint{}, r.getByIDErr
}

// Update falla si se configuró un error de escritura (HU-13).
func (r failingRepository) Update(sprint Sprint) (Sprint, error) {
	return Sprint{}, r.updateErr
}

var errRepository = errors.New("falla del repositorio")

// CA-10.8
func TestCreate_ErrorDelRepositorioAlGuardarSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	if _, err := projects.Create("Sistema de ventas", ""); err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	manager := NewManager(failingRepository{saveErr: errRepository}, projects, items)

	created, err := manager.Create(1, validName, validStart, validEnd, validGoal)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if !reflect.DeepEqual(created, Sprint{}) {
		t.Errorf("sprint devuelto = %+v, se esperaba un sprint vacío", created)
	}
}

// CA-10.8
func TestCreate_ErrorDelRepositorioAlConsultarElSprintAbiertoSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	if _, err := projects.Create("Sistema de ventas", ""); err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	manager := NewManager(failingRepository{openErr: errRepository}, projects, items)

	created, err := manager.Create(1, validName, validStart, validEnd, validGoal)

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if !reflect.DeepEqual(created, Sprint{}) {
		t.Errorf("sprint devuelto = %+v, se esperaba un sprint vacío", created)
	}
}
