package backlog

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-07-modificar-item.feature con godog.
func TestFeatures_HU07ModificarItem(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeUpdateItemScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-07-modificar-item.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-07")
	}
}

// updateItemScenario guarda el estado de un escenario de HU-07.
type updateItemScenario struct {
	projectRepo *project.MemoryRepository
	projects    *project.Manager
	repo        *MemoryRepository
	manager     *Manager

	updated   Item
	err       error
	versions  []Item
	hoursSent string
}

func initializeUpdateItemScenario(ctx *godog.ScenarioContext) {
	s := &updateItemScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projectRepo = project.NewMemoryRepository()
		s.projects = project.NewManager(s.projectRepo)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.projects)
		s.updated = Item{}
		s.err = nil
		s.versions = nil
		s.hoursSent = ""
		return ctx, nil
	})

	ctx.Step(`^que no hay ítems creados$`, s.noItems)
	ctx.Step(`^que existe un ítem con título "([^"]*)", descripción "([^"]*)", prioridad "([^"]*)" y criterios de aceptación (.*)$`, s.existingItem)
	ctx.Step(`^que existe un ítem con título "([^"]*)", descripción "([^"]*)", prioridad "([^"]*)", criterios de aceptación (.*) y horas estimadas (.*)$`, s.existingItemWithHours)
	ctx.Step(`^que existen (\d+) ítems en el proyecto "([^"]*)"$`, s.existingItemsInProject)
	ctx.Step(`^modifico el ítem con el ID (-?\d+) con título "([^"]*)", descripción "([^"]*)", prioridad "([^"]*)", criterios de aceptación (.*) y horas estimadas (.*)$`, s.updateItem)
	ctx.Step(`^la modificación es exitosa$`, s.updatedOK)
	ctx.Step(`^el ítem modificado tiene el título "([^"]*)"$`, s.updatedWithTitle)
	ctx.Step(`^el ítem modificado tiene la descripción "([^"]*)"$`, s.updatedWithDescription)
	ctx.Step(`^el ítem modificado tiene la prioridad "([^"]*)"$`, s.updatedWithPriority)
	ctx.Step(`^el ítem modificado tiene los criterios de aceptación (.*)$`, s.updatedWithCriteria)
	ctx.Step(`^el ítem modificado tiene las horas estimadas (.*)$`, s.updatedWithHours)
	ctx.Step(`^el ítem modificado queda sin horas estimadas$`, s.updatedWithoutHours)
	ctx.Step(`^el ítem modificado pertenece al proyecto con ID (\d+)$`, s.updatedWithProjectID)
	ctx.Step(`^el ítem modificado está en el estado "([^"]*)"$`, s.updatedWithStatus)
	ctx.Step(`^la modificación falla porque el título es obligatorio$`, s.failsWithEmptyTitle)
	ctx.Step(`^la modificación falla porque la prioridad debe ser ALTA, MEDIA o BAJA$`, s.failsWithInvalidPriority)
	ctx.Step(`^la modificación falla porque el ítem necesita al menos un criterio de aceptación$`, s.failsWithNoAcceptanceCriteria)
	ctx.Step(`^la modificación falla porque las horas estimadas son inválidas$`, s.failsWithInvalidEstimatedHours)
	ctx.Step(`^la modificación falla porque el ítem no existe$`, s.failsWithItemNotFound)
	ctx.Step(`^el ítem guardado con el ID (\d+) sigue teniendo el título "([^"]*)" y la prioridad "([^"]*)"$`, s.savedItemHasTitleAndPriority)
	ctx.Step(`^el ítem guardado con el ID (\d+) sigue teniendo las horas estimadas (.*)$`, s.savedItemHasHours)
	ctx.Step(`^al reconsultar el ítem con el ID (\d+) obtengo el título "([^"]*)" y los criterios de aceptación (.*)$`, s.requeriedItemHasTitleAndCriteria)
	ctx.Step(`^cambio el título y los criterios del ítem modificado$`, s.mutateUpdatedItem)
	ctx.Step(`^se modifican en simultáneo (\d+) veces el ítem con el ID (\d+) con títulos distintos$`, s.concurrentUpdates)
	ctx.Step(`^hay (\d+) ítems guardados$`, s.savedCount)
	ctx.Step(`^el ítem guardado con el ID (\d+) coincide con una de las versiones enviadas$`, s.savedItemMatchesOneVersion)
}

// parseHours interpreta el valor de horas de un paso BDD: "sin horas" es 0,
// "NaN" y "Inf" son los valores especiales y el resto se convierte con
// strconv.ParseFloat.
func parseHours(raw string) (float64, error) {
	switch strings.TrimSpace(raw) {
	case "sin horas", "0":
		return 0, nil
	case "NaN":
		return math.NaN(), nil
	case "Inf":
		return math.Inf(1), nil
	default:
		hours, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return 0, fmt.Errorf("no se pudo interpretar las horas estimadas %q: %w", raw, err)
		}
		return hours, nil
	}
}

// equalHours compara horas; NaN es igual a NaN para los pasos BDD.
func equalHours(got, want float64) bool {
	if math.IsNaN(want) {
		return math.IsNaN(got)
	}
	return got == want
}

func (s *updateItemScenario) noItems() error {
	if len(s.repo.items) != 0 {
		return fmt.Errorf("hay %d ítems, se esperaban 0", len(s.repo.items))
	}
	return nil
}

func (s *updateItemScenario) existingItem(title, description, priority, criteriaRaw string) error {
	return s.createExistingItem(title, description, priority, criteriaRaw, 0)
}

func (s *updateItemScenario) existingItemWithHours(title, description, priority, criteriaRaw, hoursRaw string) error {
	hours, err := parseHours(hoursRaw)
	if err != nil {
		return err
	}
	return s.createExistingItem(title, description, priority, criteriaRaw, hours)
}

// createExistingItem crea un proyecto de prueba y un ítem en él, y le asigna
// las horas estimadas con la propia modificación de HU-07.
func (s *updateItemScenario) createExistingItem(title, description, priority, criteriaRaw string, hours float64) error {
	if _, err := s.projects.Create("Sistema de ventas", ""); err != nil {
		return err
	}
	criteria := parseCriteria(criteriaRaw)
	created, err := s.manager.Create(1, title, description, priority, criteria)
	if err != nil {
		return err
	}
	_, err = s.manager.Update(created.ID, title, description, priority, criteria, hours)
	return err
}

func (s *updateItemScenario) existingItemsInProject(count int, projectName string) error {
	if _, err := s.projects.Create(projectName, ""); err != nil {
		return err
	}
	for i := range count {
		if _, err := s.manager.Create(1, fmt.Sprintf("Ítem %d", i+1), "", "ALTA", []string{"Criterio OK"}); err != nil {
			return err
		}
	}
	return nil
}

func (s *updateItemScenario) updateItem(id int, title, description, priority, criteriaRaw, hoursRaw string) error {
	hours, err := parseHours(hoursRaw)
	if err != nil {
		return err
	}
	s.hoursSent = hoursRaw
	criteria := parseCriteria(criteriaRaw)
	s.updated, s.err = s.manager.Update(id, title, description, priority, criteria, hours)
	return nil
}

func (s *updateItemScenario) updatedOK() error {
	if s.err != nil {
		return fmt.Errorf("la modificación falló: %v", s.err)
	}
	return nil
}

func (s *updateItemScenario) updatedWithTitle(title string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if s.updated.Title != title {
		return fmt.Errorf("título = %q, se esperaba %q", s.updated.Title, title)
	}
	return nil
}

func (s *updateItemScenario) updatedWithDescription(description string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if s.updated.Description != description {
		return fmt.Errorf("descripción = %q, se esperaba %q", s.updated.Description, description)
	}
	return nil
}

func (s *updateItemScenario) updatedWithPriority(priority string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if string(s.updated.Priority) != priority {
		return fmt.Errorf("prioridad = %q, se esperaba %q", s.updated.Priority, priority)
	}
	return nil
}

func (s *updateItemScenario) updatedWithCriteria(raw string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	return equalCriteria(s.updated.AcceptanceCriteria, parseCriteria(raw))
}

func (s *updateItemScenario) updatedWithHours(raw string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	want, err := parseHours(raw)
	if err != nil {
		return err
	}
	if !equalHours(s.updated.EstimatedHours, want) {
		return fmt.Errorf("horas estimadas = %v, se esperaban %v", s.updated.EstimatedHours, raw)
	}
	return nil
}

func (s *updateItemScenario) updatedWithoutHours() error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if s.updated.EstimatedHours != 0 {
		return fmt.Errorf("horas estimadas = %v, se esperaban 0 (sin horas)", s.updated.EstimatedHours)
	}
	return nil
}

func (s *updateItemScenario) updatedWithProjectID(projectID int) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if s.updated.ProjectID != projectID {
		return fmt.Errorf("ProjectID = %d, se esperaba %d", s.updated.ProjectID, projectID)
	}
	return nil
}

func (s *updateItemScenario) updatedWithStatus(status string) error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if string(s.updated.Status) != status {
		return fmt.Errorf("estado = %q, se esperaba %q", s.updated.Status, status)
	}
	return nil
}

func (s *updateItemScenario) failsWithEmptyTitle() error {
	if !errors.Is(s.err, ErrEmptyTitle) {
		return fmt.Errorf("error = %v, se esperaba ErrEmptyTitle", s.err)
	}
	return nil
}

func (s *updateItemScenario) failsWithInvalidPriority() error {
	if !errors.Is(s.err, ErrInvalidPriority) {
		return fmt.Errorf("error = %v, se esperaba ErrInvalidPriority", s.err)
	}
	return nil
}

func (s *updateItemScenario) failsWithNoAcceptanceCriteria() error {
	if !errors.Is(s.err, ErrNoAcceptanceCriteria) {
		return fmt.Errorf("error = %v, se esperaba ErrNoAcceptanceCriteria", s.err)
	}
	return nil
}

func (s *updateItemScenario) failsWithInvalidEstimatedHours() error {
	if !errors.Is(s.err, ErrInvalidEstimatedHours) {
		return fmt.Errorf("error = %v, se esperaba ErrInvalidEstimatedHours", s.err)
	}
	return nil
}

func (s *updateItemScenario) failsWithItemNotFound() error {
	if !errors.Is(s.err, ErrItemNotFound) {
		return fmt.Errorf("error = %v, se esperaba ErrItemNotFound", s.err)
	}
	return nil
}

func (s *updateItemScenario) savedItemHasTitleAndPriority(id int, title, priority string) error {
	item, err := s.manager.GetItem(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	if item.Title != title {
		return fmt.Errorf("título guardado = %q, se esperaba %q", item.Title, title)
	}
	if string(item.Priority) != priority {
		return fmt.Errorf("prioridad guardada = %q, se esperaba %q", item.Priority, priority)
	}
	return nil
}

func (s *updateItemScenario) savedItemHasHours(id int, raw string) error {
	item, err := s.manager.GetItem(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	want, err := parseHours(raw)
	if err != nil {
		return err
	}
	if !equalHours(item.EstimatedHours, want) {
		return fmt.Errorf("horas estimadas guardadas = %v, se esperaban %q", item.EstimatedHours, raw)
	}
	return nil
}

func (s *updateItemScenario) requeriedItemHasTitleAndCriteria(id int, title, criteriaRaw string) error {
	item, err := s.manager.GetItem(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	if item.Title != title {
		return fmt.Errorf("título reconsultado = %q, se esperaba %q", item.Title, title)
	}
	return equalCriteria(item.AcceptanceCriteria, parseCriteria(criteriaRaw))
}

func (s *updateItemScenario) mutateUpdatedItem() error {
	if err := s.updatedOK(); err != nil {
		return err
	}
	if len(s.updated.AcceptanceCriteria) == 0 {
		return errors.New("el ítem modificado no tiene criterios para modificar")
	}
	s.updated.Title = "Otro título"
	s.updated.AcceptanceCriteria[0] = "Otro criterio"
	return nil
}

// concurrentUpdates modifica el mismo ítem en simultáneo; cada modificación
// envía una versión distinta (título, criterio y horas correlativos) y las
// versiones quedan guardadas para verificar el resultado final.
func (s *updateItemScenario) concurrentUpdates(attempts, itemID int) error {
	s.versions = make([]Item, attempts)
	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		criteria := []string{fmt.Sprintf("Criterio %d", i)}
		s.versions[i] = Item{
			ID:                 itemID,
			Title:              fmt.Sprintf("Título %d", i),
			Priority:           PriorityHigh,
			AcceptanceCriteria: criteria,
			EstimatedHours:     float64(i),
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = s.manager.Update(itemID, s.versions[i].Title, "", "ALTA", criteria, s.versions[i].EstimatedHours)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			return fmt.Errorf("la modificación %d falló: %v", i+1, err)
		}
	}
	return nil
}

func (s *updateItemScenario) savedCount(want int) error {
	if len(s.repo.items) != want {
		return fmt.Errorf("hay %d ítems guardados, se esperaban %d", len(s.repo.items), want)
	}
	return nil
}

func (s *updateItemScenario) savedItemMatchesOneVersion(id int) error {
	saved, err := s.manager.GetItem(id)
	if err != nil {
		return fmt.Errorf("la reconsulta del ID %d falló: %v", id, err)
	}
	for _, v := range s.versions {
		if saved.Title == v.Title && string(saved.Priority) == string(v.Priority) &&
			saved.EstimatedHours == v.EstimatedHours &&
			slices.Equal(saved.AcceptanceCriteria, v.AcceptanceCriteria) {
			return nil
		}
	}
	return fmt.Errorf("el ítem guardado %+v no coincide con ninguna de las %d versiones enviadas", saved, len(s.versions))
}
