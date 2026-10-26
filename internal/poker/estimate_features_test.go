package poker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-25 · Registrar la estimación acordada (Issue #43). Los pasos llaman al
// Manager de poker (y al de backlog para preparar y consultar el ítem), no al
// handler HTTP; la API se prueba en internal/httpapi.
func TestFeatures_HU25RegistrarEstimacionAcordada(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeEstimateScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-25-registrar-estimacion-acordada.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-25")
	}
}

// estimateResult es el resultado de un registro: los puntos pedidos y el
// error devuelto.
type estimateResult struct {
	points int
	err    error
}

type estimateScenario struct {
	projects    *project.Manager
	items       *backlog.Manager
	repo        *MemoryRepository
	manager     *Manager
	projectID   int
	itemIDs     map[string]int
	session     Session
	estimated   Session
	estimateErr error
	results     []estimateResult
}

func initializeEstimateScenario(ctx *godog.ScenarioContext) {
	s := &estimateScenario{}
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		*s = estimateScenario{itemIDs: map[string]int{}}
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.items = backlog.NewManager(backlog.NewMemoryRepository(), s.projects)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.items)
		return ctx, nil
	})
	ctx.Step(`^un proyecto de prueba$`, s.createProject)
	ctx.Step(`^un ítem "([^"]*)" en estado "([^"]*)" con (\d+) story points$`, s.createItem)
	ctx.Step(`^una sesión abierta para el ítem "([^"]*)"$`, s.openSession)
	ctx.Step(`^que "([^"]*)" votó (\d+) puntos$`, s.previousVote)
	ctx.Step(`^que los votos de la sesión fueron revelados$`, s.previousReveal)
	ctx.Step(`^que el ítem "([^"]*)" tiene (\d+) story points asignados$`, s.assignPoints)
	ctx.Step(`^que el ítem "([^"]*)" pasó a estado "TERMINADO"$`, s.finishItem)
	ctx.Step(`^que se registró la estimación acordada de (\d+) puntos$`, s.previousEstimate)
	ctx.Step(`^se registra la estimación acordada de (-?\d+) puntos$`, s.estimate)
	ctx.Step(`^se registra la estimación acordada de (\d+) puntos en la sesión (\d+)$`, s.estimateSession)
	ctx.Step(`^se registran al mismo tiempo las estimaciones de (\d+) y (\d+) puntos$`, s.estimateConcurrently)
	ctx.Step(`^el registro es exitoso$`, s.estimateSucceeded)
	ctx.Step(`^la sesión devuelta está en estado "([^"]*)"$`, s.returnedStatus)
	ctx.Step(`^el ítem "([^"]*)" tiene (\d+) story points$`, s.itemPoints)
	ctx.Step(`^la sesión guardada está en estado "([^"]*)"$`, s.storedStatus)
	ctx.Step(`^los votos guardados son:$`, s.storedVotes)
	ctx.Step(`^el registro falla porque la carta no pertenece a la escala$`, s.failsInvalidCard)
	ctx.Step(`^el registro falla porque la sesión no está revelada$`, s.failsNotRevealed)
	ctx.Step(`^el registro falla porque la sesión no existe$`, s.failsNotFound)
	ctx.Step(`^el registro falla porque el ítem está terminado$`, s.failsItemFinished)
	ctx.Step(`^solo uno de los registros se aplica$`, s.onlyOneApplied)
	ctx.Step(`^el otro registro falla porque la sesión no está revelada$`, s.otherNotRevealed)
	ctx.Step(`^el ítem "([^"]*)" tiene los story points del registro aplicado$`, s.itemHasWinnerPoints)
}

func (s *estimateScenario) createProject() error {
	p, err := s.projects.Create("Proyecto de prueba", "Proyecto para HU-25")
	if err != nil {
		return err
	}
	s.projectID = p.ID
	return nil
}

func (s *estimateScenario) createItem(title, status string, points int) error {
	it, err := s.items.Create(s.projectID, title, "Descripción", "MEDIA", []string{"Criterio"})
	if err != nil {
		return err
	}
	if points > 0 {
		if _, err = s.items.AssignStoryPoints(it.ID, points); err != nil {
			return err
		}
	}
	if status != string(backlog.StatusPending) {
		return fmt.Errorf("estado %q no soportado por este paso", status)
	}
	s.itemIDs[title] = it.ID
	return nil
}

// openSession abre una sesión para el ítem; la última sesión abierta es la
// que usan los pasos siguientes.
func (s *estimateScenario) openSession(title string) error {
	session, err := s.manager.OpenSession(s.itemIDs[title])
	if err != nil {
		return err
	}
	s.session = session
	return nil
}

func (s *estimateScenario) previousVote(voter string, points int) error {
	_, err := s.manager.Vote(s.session.ID, voter, points)
	return err
}

func (s *estimateScenario) previousReveal() error {
	_, err := s.manager.Reveal(s.session.ID)
	return err
}

func (s *estimateScenario) assignPoints(title string, points int) error {
	_, err := s.items.AssignStoryPoints(s.itemIDs[title], points)
	return err
}

func (s *estimateScenario) finishItem(title string) error {
	for _, next := range []backlog.Status{backlog.StatusInProgress, backlog.StatusFinished} {
		if _, err := s.items.ChangeStatus(s.itemIDs[title], string(next)); err != nil {
			return err
		}
	}
	return nil
}

func (s *estimateScenario) previousEstimate(points int) error {
	_, err := s.manager.Estimate(s.session.ID, points)
	return err
}

func (s *estimateScenario) estimate(points int) error {
	return s.estimateSession(points, s.session.ID)
}

func (s *estimateScenario) estimateSession(points, sessionID int) error {
	s.estimated, s.estimateErr = s.manager.Estimate(sessionID, points)
	return nil
}

// estimateConcurrently lanza los dos registros al mismo tiempo sobre la misma
// sesión y guarda el resultado de cada uno (CA-25.8).
func (s *estimateScenario) estimateConcurrently(first, second int) error {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	start := make(chan struct{})
	for _, points := range []int{first, second} {
		wg.Add(1)
		go func(points int) {
			defer wg.Done()
			<-start
			_, err := s.manager.Estimate(s.session.ID, points)
			mu.Lock()
			defer mu.Unlock()
			s.results = append(s.results, estimateResult{points: points, err: err})
		}(points)
	}
	close(start)
	wg.Wait()
	return nil
}

func (s *estimateScenario) estimateSucceeded() error {
	if s.estimateErr != nil {
		return fmt.Errorf("el registro falló: %w", s.estimateErr)
	}
	return nil
}

func (s *estimateScenario) returnedStatus(status string) error {
	if s.estimated.Status != Status(status) {
		return fmt.Errorf("estado devuelto = %q, se esperaba %q", s.estimated.Status, status)
	}
	return nil
}

func (s *estimateScenario) itemPoints(title string, points int) error {
	it, err := s.items.GetItem(s.itemIDs[title])
	if err != nil {
		return err
	}
	if it.StoryPoints != points {
		return fmt.Errorf("story points del ítem %q = %d, se esperaban %d", title, it.StoryPoints, points)
	}
	return nil
}

func (s *estimateScenario) storedStatus(status string) error {
	stored, err := s.repo.GetByID(s.session.ID)
	if err != nil {
		return err
	}
	if stored.Status != Status(status) {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", stored.Status, status)
	}
	return nil
}

func (s *estimateScenario) storedVotes(table *godog.Table) error {
	stored, err := s.repo.GetByID(s.session.ID)
	if err != nil {
		return err
	}
	return compareVotes("votos guardados", stored.Votes, table)
}

func (s *estimateScenario) failsInvalidCard() error {
	return expectRevealError(s.estimateErr, ErrInvalidCard)
}

func (s *estimateScenario) failsNotRevealed() error {
	return expectRevealError(s.estimateErr, ErrSessionNotRevealed)
}

func (s *estimateScenario) failsNotFound() error {
	return expectRevealError(s.estimateErr, ErrSessionNotFound)
}

func (s *estimateScenario) failsItemFinished() error {
	return expectRevealError(s.estimateErr, backlog.ErrItemFinished)
}

// winner devuelve el único registro sin error, o un error si no hay
// exactamente uno.
func (s *estimateScenario) winner() (estimateResult, error) {
	var applied []estimateResult
	for _, r := range s.results {
		if r.err == nil {
			applied = append(applied, r)
		}
	}
	if len(applied) != 1 {
		return estimateResult{}, fmt.Errorf("registros aplicados = %d (%+v), se esperaba exactamente 1", len(applied), s.results)
	}
	return applied[0], nil
}

func (s *estimateScenario) onlyOneApplied() error {
	_, err := s.winner()
	return err
}

func (s *estimateScenario) otherNotRevealed() error {
	for _, r := range s.results {
		if r.err != nil && !errors.Is(r.err, ErrSessionNotRevealed) {
			return fmt.Errorf("registro de %d puntos falló con %v, se esperaba ErrSessionNotRevealed", r.points, r.err)
		}
	}
	return nil
}

func (s *estimateScenario) itemHasWinnerPoints(title string) error {
	w, err := s.winner()
	if err != nil {
		return err
	}
	return s.itemPoints(title, w.points)
}
