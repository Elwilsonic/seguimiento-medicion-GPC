package backlog

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-08-cambiar-estado-item.feature con godog.
func TestFeatures_HU08CambiarEstado(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeChangeStatusScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-08-cambiar-estado-item.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-08")
	}
}

// changeStatusScenario guarda el estado de un escenario de HU-08.
type changeStatusScenario struct {
	projects *project.Manager
	repo     *MemoryRepository
	manager  *Manager

	created        Item
	changed        Item
	err            error
	concurrentErrs []error
}

func initializeChangeStatusScenario(ctx *godog.ScenarioContext) {
	s := &changeStatusScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects)
		s.created = Item{}
		s.changed = Item{}
		s.err = nil
		s.concurrentErrs = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un ítem con estado "([^"]*)"$`, s.existingItem)
	ctx.Step(`^que hay (\d+) ítems creados en estado "([^"]*)"$`, s.existingItems)
	ctx.Step(`^que el repositorio falla (al buscar|al actualizar)$`, s.repoFails)
	ctx.Step(`^cambio el estado del ítem con el ID (-?\d+) a "([^"]*)"$`, s.changeStatus)
	ctx.Step(`^el cambio de estado es exitoso$`, s.changeOK)
	ctx.Step(`^el cambio de estado falla porque la transición no está permitida$`, s.failsInvalidTransition)
	ctx.Step(`^el cambio de estado falla porque el estado no es válido$`, s.failsInvalidStatus)
	ctx.Step(`^el cambio de estado falla porque el ítem no existe$`, s.failsItemNotFound)
	ctx.Step(`^el cambio de estado falla por un error del repositorio$`, s.failsRepository)
	ctx.Step(`^el ítem devuelto está en el estado "([^"]*)"$`, s.changedWithStatus)
	ctx.Step(`^el ítem guardado con el ID (-?\d+) está en el estado "([^"]*)"$`, s.storedWithStatus)
	ctx.Step(`^el ítem guardado con el ID (-?\d+) sigue en el estado "([^"]*)"$`, s.storedWithStatus)
	ctx.Step(`^el ítem guardado con el ID (-?\d+) conserva los demás campos$`, s.storedKeepsOtherFields)
	ctx.Step(`^no se modificó ningún ítem guardado$`, s.noItemModified)
	ctx.Step(`^modifico el estado y los criterios del ítem devuelto$`, s.mutateChanged)
	ctx.Step(`^al reconsultar el ítem con el ID (-?\d+) obtengo el estado "([^"]*)"$`, s.requeriedStatus)
	ctx.Step(`^los criterios guardados del ítem con el ID (-?\d+) no cambiaron$`, s.storedCriteriaUnchanged)
	ctx.Step(`^se cambian en simultáneo (\d+) veces el estado del ítem con el ID (-?\d+) a "([^"]*)"$`, s.concurrentChanges)
	ctx.Step(`^ninguna operación falla$`, s.noOperationFailed)
}

// seed guarda un ítem con ese estado y con el resto de los campos completos,
// para verificar que el cambio de estado no los toca (RN-08.5).
func (s *changeStatusScenario) seed(status Status) error {
	created, err := s.repo.Save(Item{
		ProjectID:          1,
		Title:              validTitle,
		Description:        "Formulario de acceso",
		Priority:           PriorityHigh,
		AcceptanceCriteria: []string{"Login OK", "Carga correcta"},
		Status:             status,
		StoryPoints:        5,
		EstimatedHours:     6,
		SprintID:           3,
	})
	if err != nil {
		return fmt.Errorf("guardar el ítem de prueba: %w", err)
	}
	s.created = created
	return nil
}

func (s *changeStatusScenario) existingItem(status string) error {
	return s.seed(Status(status))
}

func (s *changeStatusScenario) existingItems(count int, status string) error {
	if count == 0 {
		return nil
	}
	return s.seed(Status(status))
}

// stored busca el ítem directamente en el repositorio, sin pasar por el
// Manager: así los escenarios de error del repositorio también pueden verificar
// el estado guardado.
func (s *changeStatusScenario) stored(id int) (Item, bool) {
	for _, item := range s.repo.items {
		if item.ID == id {
			return item, true
		}
	}
	return Item{}, false
}

// repoFails reemplaza el Manager por uno cuyo repositorio falla al buscar o al
// actualizar (CA-08.7 y CA-08.10).
func (s *changeStatusScenario) repoFails(operation string) error {
	if operation == "al buscar" {
		s.manager = NewManager(failingGetStatusRepository{}, s.projects)
		return nil
	}
	s.manager = NewManager(failingUpdateStatusRepository{s.repo}, s.projects)
	return nil
}

func (s *changeStatusScenario) changeStatus(id int, status string) error {
	s.changed, s.err = s.manager.ChangeStatus(id, status)
	return nil
}

func (s *changeStatusScenario) changeOK() error {
	if s.err != nil {
		return fmt.Errorf("el cambio de estado falló: %v", s.err)
	}
	return nil
}

func (s *changeStatusScenario) failsInvalidTransition() error {
	if !errors.Is(s.err, ErrInvalidTransition) {
		return fmt.Errorf("error = %v, se esperaba ErrInvalidTransition", s.err)
	}
	return nil
}

func (s *changeStatusScenario) failsInvalidStatus() error {
	if !errors.Is(s.err, ErrInvalidStatus) {
		return fmt.Errorf("error = %v, se esperaba ErrInvalidStatus", s.err)
	}
	return nil
}

func (s *changeStatusScenario) failsItemNotFound() error {
	if !errors.Is(s.err, ErrItemNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrItemNotFound", s.err)
	}
	return nil
}

func (s *changeStatusScenario) failsRepository() error {
	if !errors.Is(s.err, errRepository) {
		return fmt.Errorf("error = %v, se esperaba errRepository", s.err)
	}
	return nil
}

func (s *changeStatusScenario) changedWithStatus(status string) error {
	if string(s.changed.Status) != status {
		return fmt.Errorf("estado devuelto = %q, se esperaba %q", s.changed.Status, status)
	}
	return nil
}

func (s *changeStatusScenario) storedWithStatus(id int, status string) error {
	item, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	if string(item.Status) != status {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", item.Status, status)
	}
	return nil
}

func (s *changeStatusScenario) storedKeepsOtherFields(id int) error {
	item, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	if item.ID != s.created.ID || item.ProjectID != s.created.ProjectID ||
		item.Title != s.created.Title || item.Description != s.created.Description ||
		item.Priority != s.created.Priority || item.StoryPoints != s.created.StoryPoints ||
		item.EstimatedHours != s.created.EstimatedHours || item.SprintID != s.created.SprintID {
		return fmt.Errorf("el ítem guardado cambió: %+v, se esperaba %+v", item, s.created)
	}
	if !slices.Equal(item.AcceptanceCriteria, s.created.AcceptanceCriteria) {
		return fmt.Errorf("criterios guardados = %q, se esperaban %q", item.AcceptanceCriteria, s.created.AcceptanceCriteria)
	}
	return nil
}

func (s *changeStatusScenario) noItemModified() error {
	for _, item := range s.repo.items {
		if item.Status != StatusPending {
			return fmt.Errorf("el ítem %d cambió de estado a %q", item.ID, item.Status)
		}
	}
	return nil
}

func (s *changeStatusScenario) mutateChanged() error {
	if err := s.changeOK(); err != nil {
		return err
	}
	s.changed.Status = StatusFinished
	s.changed.Title = "Otro título"
	if len(s.changed.AcceptanceCriteria) > 0 {
		s.changed.AcceptanceCriteria[0] = "Otro criterio"
	}
	return nil
}

func (s *changeStatusScenario) requeriedStatus(id int, status string) error {
	item, err := s.manager.GetItem(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	if string(item.Status) != status {
		return fmt.Errorf("estado reconsultado = %q, se esperaba %q", item.Status, status)
	}
	return nil
}

func (s *changeStatusScenario) storedCriteriaUnchanged(id int) error {
	item, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	if !slices.Equal(item.AcceptanceCriteria, s.created.AcceptanceCriteria) {
		return fmt.Errorf("criterios guardados = %q, se esperaban %q", item.AcceptanceCriteria, s.created.AcceptanceCriteria)
	}
	return nil
}

// concurrentChanges cambia el mismo ítem al mismo estado en simultáneo; el
// primero cambia el estado y los demás son no-op (CA-08.9).
func (s *changeStatusScenario) concurrentChanges(attempts, id int, status string) error {
	s.concurrentErrs = make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, s.concurrentErrs[i] = s.manager.ChangeStatus(id, status)
		}(i)
	}
	wg.Wait()
	return nil
}

func (s *changeStatusScenario) noOperationFailed() error {
	for i, err := range s.concurrentErrs {
		if err != nil {
			return fmt.Errorf("la operación simultánea %d falló: %v", i+1, err)
		}
	}
	return nil
}

// failingUpdateStatusRepository delega en un MemoryRepository pero falla al
// actualizar, para verificar que el no-op de CA-08.7 no llama a Update.
type failingUpdateStatusRepository struct {
	*MemoryRepository
}

func (r failingUpdateStatusRepository) Update(item Item) (Item, error) {
	return Item{}, errRepository
}

// failingGetStatusRepository falla al buscar, para CA-08.10.
type failingGetStatusRepository struct{}

func (r failingGetStatusRepository) Save(item Item) (Item, error) {
	return Item{}, nil
}

func (r failingGetStatusRepository) GetByID(id int) (Item, error) {
	return Item{}, errRepository
}

func (r failingGetStatusRepository) ListByProject(projectID int) ([]Item, error) {
	return nil, nil
}

func (r failingGetStatusRepository) Update(item Item) (Item, error) {
	return Item{}, nil
}
