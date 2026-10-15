package backlog

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-04 · Crear ítem del backlog (Issue #17)

const (
	validTitle    = "Login de usuarios"
	validPriority = "ALTA"
)

// validCriteria devuelve criterios válidos nuevos, para que ningún test
// comparta la misma lista.
func validCriteria() []string {
	return []string{"Credenciales válidas"}
}

// newTestManager arma un Manager de backlog con un repositorio y un Manager
// de proyectos en memoria vacíos.
func newTestManager() (*Manager, *MemoryRepository, *project.Manager, *project.MemoryRepository) {
	projectRepo := project.NewMemoryRepository()
	projects := project.NewManager(projectRepo)
	repo := NewMemoryRepository()
	return NewManager(repo, projects), repo, projects, projectRepo
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

// savedItem devuelve el ítem guardado con ese ID, mirando el repositorio en
// memoria directamente (los tests son internos al paquete, ver SDD).
func savedItem(t *testing.T, repo *MemoryRepository, id int) Item {
	t.Helper()
	for _, item := range repo.items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("no hay ningún ítem guardado con el ID %d", id)
	return Item{}
}

// CA-04.1, CA-04.4
func TestCreate_ItemValidoQuedaConIDEstadoPendienteYStoryPointsEnCero(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	item, err := manager.Create(projectID, "  Login de usuarios  ", "Acceso con usuario y contraseña", " alta ", []string{"  Credenciales válidas  ", "", "   "})

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if item.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", item.ID)
	}
	if item.ProjectID != projectID {
		t.Errorf("ProjectID = %d, se esperaba %d", item.ProjectID, projectID)
	}
	if item.Title != validTitle {
		t.Errorf("Title = %q, se esperaba %q", item.Title, validTitle)
	}
	if item.Description != "Acceso con usuario y contraseña" {
		t.Errorf("Description = %q, se esperaba %q", item.Description, "Acceso con usuario y contraseña")
	}
	if item.Priority != PriorityHigh {
		t.Errorf("Priority = %q, se esperaba %q", item.Priority, PriorityHigh)
	}
	if !slices.Equal(item.AcceptanceCriteria, validCriteria()) {
		t.Errorf("AcceptanceCriteria = %q, se esperaba %q", item.AcceptanceCriteria, validCriteria())
	}
	if item.Status != StatusPending {
		t.Errorf("Status = %q, se esperaba %q", item.Status, StatusPending)
	}
	if item.StoryPoints != 0 {
		t.Errorf("StoryPoints = %d, se esperaba 0", item.StoryPoints)
	}

	if saved := savedItem(t, repo, item.ID); !reflect.DeepEqual(saved, item) {
		t.Errorf("ítem guardado = %+v, se esperaba %+v", saved, item)
	}
}

// CA-04.4
func TestCreate_PrimerItemRecibeID1EstadoPendienteYStoryPointsEnCero(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	item, err := manager.Create(projectID, validTitle, "", validPriority, validCriteria())

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if item.ID != 1 {
		t.Errorf("ID = %d, se esperaba 1", item.ID)
	}
	if item.Status != StatusPending {
		t.Errorf("Status = %q, se esperaba %q", item.Status, StatusPending)
	}
	if item.StoryPoints != 0 {
		t.Errorf("StoryPoints = %d, se esperaba 0", item.StoryPoints)
	}
	if len(repo.items) != 1 {
		t.Errorf("hay %d ítems guardados, se esperaba 1", len(repo.items))
	}
}

// RN-04.2
func TestCreate_DescripcionOpcionalSeGuardaTalComoLlega(t *testing.T) {
	cases := map[string]string{
		"vacía":                    "",
		"con espacios en extremos": "  Detalle con espacios  ",
	}
	for caseName, description := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")

			item, err := manager.Create(projectID, validTitle, description, validPriority, validCriteria())

			if err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}
			if item.Description != description {
				t.Errorf("Description devuelta = %q, se esperaba %q", item.Description, description)
			}
			if saved := savedItem(t, repo, item.ID); saved.Description != description {
				t.Errorf("Description guardada = %q, se esperaba %q", saved.Description, description)
			}
		})
	}
}

// RN-04.1
func TestCreate_TituloRepetidoEnElMismoProyectoSePermite(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	first, err := manager.Create(projectID, validTitle, "Primero", validPriority, validCriteria())
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	second, err := manager.Create(projectID, validTitle, "Segundo", "MEDIA", validCriteria())
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	if first.ID != 1 || second.ID != 2 {
		t.Errorf("IDs = %d y %d, se esperaban 1 y 2", first.ID, second.ID)
	}
	if len(repo.items) != 2 {
		t.Errorf("hay %d ítems guardados, se esperaban 2", len(repo.items))
	}
}

// CA-04.2
func TestCreate_TituloVacioDevuelveErrEmptyTitle(t *testing.T) {
	cases := map[string]string{
		"vacío":              "",
		"solo espacios":      "   ",
		"solo tabs y saltos": "\t\n ",
	}
	for caseName, title := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")

			_, err := manager.Create(projectID, title, "", validPriority, validCriteria())

			if !errors.Is(err, ErrEmptyTitle) {
				t.Errorf("error = %v, se esperaba ErrEmptyTitle", err)
			}
			if len(repo.items) != 0 {
				t.Errorf("hay %d ítems guardados, se esperaba 0", len(repo.items))
			}
		})
	}
}

// CA-04.3
func TestCreate_PrioridadSeNormalizaAMayusculas(t *testing.T) {
	cases := map[string]Priority{
		" alta ": PriorityHigh,
		"Media":  PriorityMedium,
		"baja":   PriorityLow,
	}
	for priority, want := range cases {
		t.Run(priority, func(t *testing.T) {
			manager, _, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")

			item, err := manager.Create(projectID, validTitle, "", priority, validCriteria())

			if err != nil {
				t.Fatalf("Create() devolvió un error inesperado: %v", err)
			}
			if item.Priority != want {
				t.Errorf("Priority = %q, se esperaba %q", item.Priority, want)
			}
		})
	}
}

// CA-04.3
func TestCreate_PrioridadInvalidaDevuelveErrInvalidPriority(t *testing.T) {
	cases := map[string]string{
		"vacía":         "",
		"solo espacios": "   ",
		"desconocida":   "URGENTE",
	}
	for caseName, priority := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")

			_, err := manager.Create(projectID, validTitle, "", priority, validCriteria())

			if !errors.Is(err, ErrInvalidPriority) {
				t.Errorf("error = %v, se esperaba ErrInvalidPriority", err)
			}
			if len(repo.items) != 0 {
				t.Errorf("hay %d ítems guardados, se esperaba 0", len(repo.items))
			}
		})
	}
}

// CA-04.2
func TestCreate_SinCriteriosValidosDevuelveErrNoAcceptanceCriteria(t *testing.T) {
	cases := map[string][]string{
		"nula":        nil,
		"vacía":       {},
		"solo vacíos": {"", "   ", "\t"},
	}
	for caseName, criteria := range cases {
		t.Run(caseName, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			projectID := createTestProject(t, projects, "Sistema de ventas")

			_, err := manager.Create(projectID, validTitle, "", validPriority, criteria)

			if !errors.Is(err, ErrNoAcceptanceCriteria) {
				t.Errorf("error = %v, se esperaba ErrNoAcceptanceCriteria", err)
			}
			if len(repo.items) != 0 {
				t.Errorf("hay %d ítems guardados, se esperaba 0", len(repo.items))
			}
		})
	}
}

// RN-04.4
func TestCreate_DescartaCriteriosVaciosYConservaElOrden(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	item, err := manager.Create(projectID, validTitle, "", validPriority, []string{" Primero ", "   ", "", "\t", "Segundo"})

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	want := []string{"Primero", "Segundo"}
	if !slices.Equal(item.AcceptanceCriteria, want) {
		t.Errorf("AcceptanceCriteria = %q, se esperaba %q", item.AcceptanceCriteria, want)
	}
	if saved := savedItem(t, repo, item.ID); !slices.Equal(saved.AcceptanceCriteria, want) {
		t.Errorf("criterios guardados = %q, se esperaba %q", saved.AcceptanceCriteria, want)
	}
}

// CA-04.5
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

			_, err := manager.Create(projectID, validTitle, "", validPriority, validCriteria())

			if !errors.Is(err, project.ErrProjectNotFound) {
				t.Errorf("error = %v, se esperaba ErrProjectNotFound", err)
			}
			if len(repo.items) != 0 {
				t.Errorf("hay %d ítems guardados, se esperaba 0", len(repo.items))
			}
		})
	}
}

// RN-04.5
func TestCreate_ProyectoFinalizadoAceptaItems(t *testing.T) {
	manager, repo, _, projectRepo := newTestManager()

	p, err := projectRepo.Save(project.Project{Name: "Proyecto cerrado", Status: project.StatusFinished})
	if err != nil {
		t.Fatalf("guardar el proyecto finalizado de prueba: %v", err)
	}

	item, err := manager.Create(p.ID, validTitle, "", validPriority, validCriteria())

	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	if item.Status != StatusPending {
		t.Errorf("Status = %q, se esperaba %q", item.Status, StatusPending)
	}
	if len(repo.items) != 1 {
		t.Errorf("hay %d ítems guardados, se esperaba 1", len(repo.items))
	}
}

// CA-04.4, RN-04.6
func TestCreate_IDsCorrelativosYGlobalesEntreProyectos(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	firstProject := createTestProject(t, projects, "Sistema de ventas")
	secondProject := createTestProject(t, projects, "Portal de clientes")

	first, err := manager.Create(firstProject, validTitle, "", validPriority, validCriteria())
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	second, err := manager.Create(secondProject, "Historial de pedidos", "", "BAJA", validCriteria())
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}
	third, err := manager.Create(firstProject, "Cierre de sesión", "", "MEDIA", validCriteria())
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	if first.ID != 1 || second.ID != 2 || third.ID != 3 {
		t.Errorf("IDs = %d, %d y %d, se esperaban 1, 2 y 3", first.ID, second.ID, third.ID)
	}
	if len(repo.items) != 3 {
		t.Errorf("hay %d ítems guardados, se esperaban 3", len(repo.items))
	}
}

// CA-04.2
func TestCreate_InformaElPrimerErrorSegunElOrdenDeValidacion(t *testing.T) {
	cases := []struct {
		name     string
		title    string
		priority string
		criteria []string
		want     error
	}{
		{"título antes que la prioridad, los criterios y el proyecto", "   ", "URGENTE", nil, ErrEmptyTitle},
		{"prioridad antes que los criterios y el proyecto", validTitle, "URGENTE", nil, ErrInvalidPriority},
		{"criterios antes que el proyecto", validTitle, validPriority, nil, ErrNoAcceptanceCriteria},
		{"proyecto inexistente al final", validTitle, validPriority, validCriteria(), project.ErrProjectNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manager, repo, projects, _ := newTestManager()
			createTestProject(t, projects, "Sistema de ventas")

			_, err := manager.Create(999, tc.title, "", tc.priority, tc.criteria)

			if !errors.Is(err, tc.want) {
				t.Errorf("error = %v, se esperaba %v", err, tc.want)
			}
			if len(repo.items) != 0 {
				t.Errorf("hay %d ítems guardados, se esperaba 0", len(repo.items))
			}
		})
	}
}

// CA-04.6
func TestCreate_ModificarElItemDevueltoNoAlteraElGuardado(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	item, err := manager.Create(projectID, validTitle, "Original", validPriority, []string{"Primero"})
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	item.Title = "Otro título"
	item.Description = "Otra descripción"
	item.Priority = PriorityLow
	item.Status = StatusFinished
	item.StoryPoints = 5
	item.AcceptanceCriteria[0] = "Otro criterio"

	saved := savedItem(t, repo, 1)
	if saved.Title != validTitle || saved.Description != "Original" || saved.Priority != PriorityHigh ||
		saved.Status != StatusPending || saved.StoryPoints != 0 {
		t.Errorf("el ítem guardado cambió: %+v", saved)
	}
	if !slices.Equal(saved.AcceptanceCriteria, []string{"Primero"}) {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, []string{"Primero"})
	}
}

// CA-04.6
func TestCreate_ModificarLaListaDeCriteriosRecibidaNoAlteraElGuardado(t *testing.T) {
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	criteria := []string{"Primero"}
	item, err := manager.Create(projectID, validTitle, "", validPriority, criteria)
	if err != nil {
		t.Fatalf("Create() devolvió un error inesperado: %v", err)
	}

	criteria[0] = "Otro criterio"

	if !slices.Equal(item.AcceptanceCriteria, []string{"Primero"}) {
		t.Errorf("criterios devueltos = %q, se esperaban %q", item.AcceptanceCriteria, []string{"Primero"})
	}
	if saved := savedItem(t, repo, 1); !slices.Equal(saved.AcceptanceCriteria, []string{"Primero"}) {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, []string{"Primero"})
	}
}

// CA-04.7
func TestCreate_CreacionesSimultaneasRecibenIDsDistintos(t *testing.T) {
	const attempts = 50
	manager, repo, projects, _ := newTestManager()
	projectID := createTestProject(t, projects, "Sistema de ventas")

	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = manager.Create(projectID, fmt.Sprintf("Login de usuarios %d", i), "", validPriority, []string{"Login OK"})
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("la creación %d falló: %v", i+1, err)
		}
	}
	if len(repo.items) != attempts {
		t.Fatalf("hay %d ítems guardados, se esperaban %d", len(repo.items), attempts)
	}
	ids := make(map[int]bool)
	for _, item := range repo.items {
		if ids[item.ID] {
			t.Errorf("el ID %d está repetido", item.ID)
		}
		ids[item.ID] = true
	}
}

// failingRepository es un Repository que falla al guardar, al buscar o al listar.
type failingRepository struct {
	saveErr error
	getErr  error
	listErr error
}

func (r failingRepository) Save(item Item) (Item, error) {
	return Item{}, r.saveErr
}

// GetByID falla si se configuró un error de búsqueda (HU-05).
func (r failingRepository) GetByID(id int) (Item, error) {
	return Item{}, r.getErr
}

func (r failingRepository) ListByProject(projectID int) ([]Item, error) {
	return nil, r.listErr
}

var errRepository = errors.New("falla del repositorio")

// CA-04.8
func TestCreate_ErrorDelRepositorioAlGuardarSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	if _, err := projects.Create("Sistema de ventas", ""); err != nil {
		t.Fatalf("crear el proyecto de prueba: %v", err)
	}
	manager := NewManager(failingRepository{saveErr: errRepository}, projects)

	item, err := manager.Create(1, validTitle, "", validPriority, validCriteria())

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba el error del repositorio", err)
	}
	if !reflect.DeepEqual(item, Item{}) {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", item)
	}
}
