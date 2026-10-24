package backlog

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-20-asignar-story-points.feature con godog.
func TestFeatures_HU20AsignarStoryPoints(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeStoryPointsScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-20-asignar-story-points.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-20")
	}
}

// storyPointsScenario guarda el estado de un escenario de HU-20.
type storyPointsScenario struct {
	projects *project.Manager
	repo     *MemoryRepository
	manager  *Manager

	created  Item
	assigned Item
	err      error

	scale []int

	concurrentErrs []error
	sentPoints     []int
}

func initializeStoryPointsScenario(ctx *godog.ScenarioContext) {
	s := &storyPointsScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects)
		s.created = Item{}
		s.assigned = Item{}
		s.err = nil
		s.scale = nil
		s.concurrentErrs = nil
		s.sentPoints = nil
		return ctx, nil
	})

	ctx.Step(`^que existe un ítem sin estimar en estado "([^"]*)"$`, s.existingUnestimatedItem)
	ctx.Step(`^que existe un ítem en estado "([^"]*)" con (\d+) story points$`, s.existingItem)
	ctx.Step(`^que hay (\d+) ítems sin estimar creados$`, s.existingItems)
	ctx.Step(`^que el ítem con el ID (\d+) está asignado al sprint con el ID (\d+)$`, s.itemInSprint)
	ctx.Step(`^que el repositorio falla (al buscar|al actualizar)$`, s.repoFails)
	ctx.Step(`^que asigné (-?\d+) story points al ítem con el ID (-?\d+)$`, s.assignedBefore)
	ctx.Step(`^asigno (-?\d+) story points al ítem con el ID (-?\d+)$`, s.assign)
	ctx.Step(`^cambio el estado del ítem con el ID (\d+) a "([^"]*)"$`, s.changeStatus)
	ctx.Step(`^modifico los datos del ítem con el ID (\d+) con el título "([^"]*)"$`, s.updateItem)
	ctx.Step(`^modifico los story points y los criterios del ítem devuelto$`, s.mutateAssigned)
	ctx.Step(`^consulto la escala de story points$`, s.readScale)
	ctx.Step(`^modifico el primer valor de la escala devuelta$`, s.mutateScale)
	ctx.Step(`^se asignan en simultáneo (\d+) veces story points de la escala al ítem con el ID (\d+) mientras se cambia su estado a "([^"]*)"$`, s.concurrentAssignments)
	ctx.Step(`^la asignación es exitosa$`, s.assignOK)
	ctx.Step(`^la modificación es exitosa$`, s.assignOK)
	ctx.Step(`^la asignación falla porque el valor no está en la escala$`, s.failsInvalidStoryPoints)
	ctx.Step(`^la asignación falla porque el ítem no existe$`, s.failsItemNotFound)
	ctx.Step(`^la asignación falla porque el ítem está terminado$`, s.failsItemFinished)
	ctx.Step(`^la asignación falla por un error del repositorio$`, s.failsRepository)
	ctx.Step(`^el ítem devuelto tiene (-?\d+) story points$`, s.assignedWithPoints)
	ctx.Step(`^el ítem guardado con el ID (\d+) tiene (-?\d+) story points$`, s.storedWithPoints)
	ctx.Step(`^el ítem guardado con el ID (\d+) conserva los demás campos$`, s.storedKeepsOtherFields)
	ctx.Step(`^el ítem guardado con el ID (\d+) sigue asignado al sprint con el ID (\d+)$`, s.storedInSprint)
	ctx.Step(`^el ítem guardado con el ID (\d+) está en el estado "([^"]*)"$`, s.storedWithStatus)
	ctx.Step(`^no se modificó ningún ítem guardado$`, s.noItemModified)
	ctx.Step(`^la escala de story points es "([^"]*)"$`, s.scaleIs)
	ctx.Step(`^cada asignación es exitosa o falla porque el ítem está terminado$`, s.eachAssignmentOKOrFinished)
	ctx.Step(`^el ítem guardado con el ID (\d+) tiene los story points de una asignación exitosa, o 0 si ninguna lo fue$`, s.storedFromSucceeded)
}

// seed guarda un ítem con ese estado y esos story points, con el resto de los
// campos completos, para verificar que la asignación no los toca (RN-20.6).
func (s *storyPointsScenario) seed(status Status, points int) error {
	created, err := s.repo.Save(Item{
		ProjectID:          1,
		Title:              validTitle,
		Description:        "Formulario de acceso",
		Priority:           PriorityHigh,
		AcceptanceCriteria: []string{"Login OK", "Carga correcta"},
		Status:             status,
		StoryPoints:        points,
		EstimatedHours:     6,
	})
	if err != nil {
		return fmt.Errorf("guardar el ítem de prueba: %w", err)
	}
	s.created = created
	return nil
}

func (s *storyPointsScenario) existingUnestimatedItem(status string) error {
	return s.seed(Status(status), 0)
}

func (s *storyPointsScenario) existingItem(status string, points int) error {
	return s.seed(Status(status), points)
}

func (s *storyPointsScenario) existingItems(count int) error {
	for range count {
		if err := s.seed(StatusPending, 0); err != nil {
			return err
		}
	}
	return nil
}

// itemInSprint escribe el SprintID directamente en el repositorio: HU-20 no
// consulta el sprint, así que no hace falta crearlo.
func (s *storyPointsScenario) itemInSprint(id, sprintID int) error {
	item, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	item.SprintID = sprintID
	updated, err := s.repo.Update(item)
	if err != nil {
		return fmt.Errorf("asignar el ítem al sprint: %w", err)
	}
	s.created = updated
	return nil
}

// stored busca el ítem directamente en el repositorio, sin pasar por el
// Manager: así los escenarios de error del repositorio también pueden verificar
// lo guardado.
func (s *storyPointsScenario) stored(id int) (Item, bool) {
	for _, item := range s.repo.items {
		if item.ID == id {
			return item, true
		}
	}
	return Item{}, false
}

// repoFails reemplaza el Manager por uno cuyo repositorio falla al buscar o al
// actualizar (CA-20.7). Reutiliza los repositorios falsos de HU-08.
func (s *storyPointsScenario) repoFails(operation string) error {
	if operation == "al buscar" {
		s.manager = NewManager(failingGetStatusRepository{}, s.projects)
		return nil
	}
	s.manager = NewManager(failingUpdateStatusRepository{s.repo}, s.projects)
	return nil
}

func (s *storyPointsScenario) assignedBefore(points, id int) error {
	if _, err := s.manager.AssignStoryPoints(id, points); err != nil {
		return fmt.Errorf("la asignación previa falló: %v", err)
	}
	return nil
}

func (s *storyPointsScenario) assign(points, id int) error {
	s.assigned, s.err = s.manager.AssignStoryPoints(id, points)
	return nil
}

func (s *storyPointsScenario) changeStatus(id int, status string) error {
	if _, err := s.manager.ChangeStatus(id, status); err != nil {
		return fmt.Errorf("el cambio de estado falló: %v", err)
	}
	return nil
}

// updateItem modifica los datos con HU-07, reenviando los demás campos
// actuales (CA-20.6).
func (s *storyPointsScenario) updateItem(id int, title string) error {
	current, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	_, s.err = s.manager.Update(id, title, current.Description, string(current.Priority),
		current.AcceptanceCriteria, current.EstimatedHours)
	return nil
}

func (s *storyPointsScenario) mutateAssigned() error {
	if err := s.assignOK(); err != nil {
		return err
	}
	s.assigned.StoryPoints = 21
	if len(s.assigned.AcceptanceCriteria) > 0 {
		s.assigned.AcceptanceCriteria[0] = "Otro criterio"
	}
	return nil
}

func (s *storyPointsScenario) readScale() error {
	s.scale = StoryPointScale()
	return nil
}

func (s *storyPointsScenario) mutateScale() error {
	if len(s.scale) == 0 {
		return fmt.Errorf("la escala devuelta está vacía")
	}
	s.scale[0] = 99
	return nil
}

// concurrentAssignments lanza las asignaciones y un cambio de estado en
// simultáneo. Si el cambio a TERMINADO llega antes, las asignaciones
// siguientes devuelven ErrItemFinished (CA-20.7).
func (s *storyPointsScenario) concurrentAssignments(attempts, id int, status string) error {
	scale := StoryPointScale()
	if len(scale) == 0 {
		return fmt.Errorf("la escala de story points está vacía")
	}
	s.concurrentErrs = make([]error, attempts)
	s.sentPoints = make([]int, attempts)
	var statusErr error
	var wg sync.WaitGroup
	for i := range attempts {
		s.sentPoints[i] = scale[i%len(scale)]
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, s.concurrentErrs[i] = s.manager.AssignStoryPoints(id, s.sentPoints[i])
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, statusErr = s.manager.ChangeStatus(id, status)
	}()
	wg.Wait()
	if statusErr != nil {
		return fmt.Errorf("el cambio de estado simultáneo falló: %v", statusErr)
	}
	return nil
}

func (s *storyPointsScenario) assignOK() error {
	if s.err != nil {
		return fmt.Errorf("la operación falló: %v", s.err)
	}
	return nil
}

func (s *storyPointsScenario) failsInvalidStoryPoints() error {
	if !errors.Is(s.err, ErrInvalidStoryPoints) {
		return fmt.Errorf("error = %v, se esperaba ErrInvalidStoryPoints", s.err)
	}
	return nil
}

func (s *storyPointsScenario) failsItemNotFound() error {
	if !errors.Is(s.err, ErrItemNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrItemNotFound", s.err)
	}
	return nil
}

func (s *storyPointsScenario) failsItemFinished() error {
	if !errors.Is(s.err, ErrItemFinished) {
		return fmt.Errorf("error = %v, se esperaba ErrItemFinished", s.err)
	}
	return nil
}

func (s *storyPointsScenario) failsRepository() error {
	if !errors.Is(s.err, errRepository) {
		return fmt.Errorf("error = %v, se esperaba errRepository", s.err)
	}
	return nil
}

func (s *storyPointsScenario) assignedWithPoints(points int) error {
	if s.assigned.StoryPoints != points {
		return fmt.Errorf("story points devueltos = %d, se esperaban %d", s.assigned.StoryPoints, points)
	}
	return nil
}

func (s *storyPointsScenario) storedWithPoints(id, points int) error {
	item, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	if item.StoryPoints != points {
		return fmt.Errorf("story points guardados = %d, se esperaban %d", item.StoryPoints, points)
	}
	return nil
}

// storedKeepsOtherFields compara todo menos los story points, que son lo que
// cambia la asignación (RN-20.6).
func (s *storyPointsScenario) storedKeepsOtherFields(id int) error {
	item, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	if item.ID != s.created.ID || item.ProjectID != s.created.ProjectID ||
		item.Title != s.created.Title || item.Description != s.created.Description ||
		item.Priority != s.created.Priority || item.Status != s.created.Status ||
		item.EstimatedHours != s.created.EstimatedHours || item.SprintID != s.created.SprintID {
		return fmt.Errorf("el ítem guardado cambió: %+v, se esperaba %+v", item, s.created)
	}
	if !slices.Equal(item.AcceptanceCriteria, s.created.AcceptanceCriteria) {
		return fmt.Errorf("criterios guardados = %q, se esperaban %q", item.AcceptanceCriteria, s.created.AcceptanceCriteria)
	}
	return nil
}

func (s *storyPointsScenario) storedInSprint(id, sprintID int) error {
	item, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	if item.SprintID != sprintID {
		return fmt.Errorf("sprint guardado = %d, se esperaba %d", item.SprintID, sprintID)
	}
	return nil
}

func (s *storyPointsScenario) storedWithStatus(id int, status string) error {
	item, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	if string(item.Status) != status {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", item.Status, status)
	}
	return nil
}

// noItemModified verifica que todos los ítems siguen sin estimar.
func (s *storyPointsScenario) noItemModified() error {
	for _, item := range s.repo.items {
		if item.StoryPoints != 0 {
			return fmt.Errorf("el ítem %d quedó con %d story points", item.ID, item.StoryPoints)
		}
	}
	return nil
}

func (s *storyPointsScenario) scaleIs(want string) error {
	got := StoryPointScale()
	parts := make([]string, len(got))
	for i, v := range got {
		parts[i] = strconv.Itoa(v)
	}
	if strings.Join(parts, ", ") != want {
		return fmt.Errorf("escala = %v, se esperaba %s", got, want)
	}
	return nil
}

func (s *storyPointsScenario) eachAssignmentOKOrFinished() error {
	for i, err := range s.concurrentErrs {
		if err != nil && !errors.Is(err, ErrItemFinished) {
			return fmt.Errorf("la asignación simultánea %d falló con %v, se esperaba nil o ErrItemFinished", i+1, err)
		}
	}
	return nil
}

// storedFromSucceeded no depende del orden de las goroutines: después de
// TERMINADO no se escribe nada y la escala no tiene el 0, así que lo guardado
// es el valor de la última asignación exitosa, o 0 si no hubo ninguna.
func (s *storyPointsScenario) storedFromSucceeded(id int) error {
	item, ok := s.stored(id)
	if !ok {
		return fmt.Errorf("no hay ningún ítem guardado con el ID %d", id)
	}
	var succeeded []int
	for i, err := range s.concurrentErrs {
		if err == nil {
			succeeded = append(succeeded, s.sentPoints[i])
		}
	}
	if len(succeeded) == 0 && item.StoryPoints != 0 {
		return fmt.Errorf("ninguna asignación fue exitosa, pero quedaron %d story points", item.StoryPoints)
	}
	if len(succeeded) > 0 && !slices.Contains(succeeded, item.StoryPoints) {
		return fmt.Errorf("story points guardados = %d, no es ninguno de los asignados con éxito %v", item.StoryPoints, succeeded)
	}
	return nil
}
