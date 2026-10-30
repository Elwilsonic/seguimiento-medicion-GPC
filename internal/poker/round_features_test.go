package poker

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-24 · Detectar diferencias y repetir ronda (Issue #42). Los pasos llaman
// al Manager y a CompareVotes, no al handler HTTP; la API se prueba en
// internal/httpapi.
func TestFeatures_HU24RepetirRonda(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeRoundScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-24-repetir-ronda.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-24")
	}
}

type roundScenario struct {
	projects    *project.Manager
	items       *backlog.Manager
	repo        *failingRoundRepository
	manager     *Manager
	projectID   int
	itemID      int
	session     Session
	revealed    Session
	revealErr   error
	comparison  VoteComparison
	repeatErr   error
	repeatErrs  []error
	estimateErr error
}

func initializeRoundScenario(ctx *godog.ScenarioContext) {
	s := &roundScenario{}
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		*s = roundScenario{}
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.items = backlog.NewManager(backlog.NewMemoryRepository(), s.projects)
		s.repo = &failingRoundRepository{MemoryRepository: NewMemoryRepository()}
		s.manager = NewManager(s.repo, s.items)
		return ctx, nil
	})
	ctx.Step(`^un proyecto de prueba$`, s.createProject)
	ctx.Step(`^un ítem "([^"]*)" en estado "([^"]*)" con (\d+) story points$`, s.createItem)
	ctx.Step(`^una sesión abierta para el ítem "([^"]*)"$`, s.openSession)
	ctx.Step(`^que "([^"]*)" votó (\d+) puntos$`, s.previousVote)
	ctx.Step(`^que los votos de la sesión fueron revelados$`, s.previousReveal)
	ctx.Step(`^que se repitió la ronda de la sesión$`, s.previousRepeat)
	ctx.Step(`^que se registró la estimación acordada de (\d+) puntos$`, s.previousEstimate)
	ctx.Step(`^que se abrió otra sesión para el ítem "([^"]*)"$`, s.openAnotherSession)
	ctx.Step(`^que el almacenamiento de sesiones falla al guardar$`, s.failOnUpdate)
	ctx.Step(`^se revelan los votos de la sesión$`, s.reveal)
	ctx.Step(`^se repite la ronda de la sesión$`, s.repeat)
	ctx.Step(`^se repite la ronda de la sesión (-?\d+)$`, s.repeatSession)
	ctx.Step(`^se repite la ronda de la sesión (\d+) veces al mismo tiempo$`, s.repeatConcurrently)
	ctx.Step(`^se registra la estimación acordada de (\d+) puntos$`, s.estimate)
	ctx.Step(`^se informa que hay diferencias$`, s.hasDifferences)
	ctx.Step(`^se informa que no hay diferencias$`, s.hasNoDifferences)
	ctx.Step(`^el voto menor es (\d+), de "([^"]*)"$`, s.lowest)
	ctx.Step(`^el voto mayor es (\d+), de "([^"]*)"$`, s.highest)
	ctx.Step(`^se muestran los votos:$`, s.revealedVotes)
	ctx.Step(`^la repetición es exitosa$`, s.repeatSucceeded)
	ctx.Step(`^solo una repetición es exitosa$`, s.onlyOneRepeatSucceeded)
	ctx.Step(`^las demás repeticiones fallan porque la sesión no está revelada$`, s.otherRepeatsNotRevealed)
	ctx.Step(`^la repetición falla porque la sesión no está revelada$`, s.repeatFailsNotRevealed)
	ctx.Step(`^la repetición falla porque la sesión no existe$`, s.repeatFailsNotFound)
	ctx.Step(`^la repetición falla porque ya hay otra sesión abierta para el ítem$`, s.repeatFailsAlreadyOpen)
	ctx.Step(`^la repetición falla por el error del almacenamiento$`, s.repeatFailsStorage)
	ctx.Step(`^la sesión guardada está en estado "([^"]*)" en la ronda (\d+)$`, s.storedStatusAndRound)
	ctx.Step(`^la sesión guardada no tiene votos$`, s.storedWithoutVotes)
	ctx.Step(`^los votos guardados son:$`, s.storedVotes)
	ctx.Step(`^sigue siendo la única sesión del ítem "([^"]*)"$`, s.onlySessionOfItem)
	ctx.Step(`^el ítem "([^"]*)" sigue en estado "([^"]*)" con (\d+) story points$`, s.itemUnchanged)
	ctx.Step(`^el ítem "([^"]*)" tiene (\d+) story points$`, s.itemPoints)
}

func (s *roundScenario) createProject() error {
	p, err := s.projects.Create("Proyecto de prueba", "Proyecto para HU-24")
	if err != nil {
		return err
	}
	s.projectID = p.ID
	return nil
}

func (s *roundScenario) createItem(title, status string, points int) error {
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
	s.itemID = it.ID
	return nil
}

func (s *roundScenario) openSession(_ string) error {
	session, err := s.manager.OpenSession(s.itemID)
	if err != nil {
		return err
	}
	s.session = session
	return nil
}

func (s *roundScenario) previousVote(voter string, points int) error {
	_, err := s.manager.Vote(s.session.ID, voter, points)
	return err
}

func (s *roundScenario) previousReveal() error {
	_, err := s.manager.Reveal(s.session.ID)
	return err
}

func (s *roundScenario) previousRepeat() error {
	_, err := s.manager.RepeatRound(s.session.ID)
	return err
}

func (s *roundScenario) previousEstimate(points int) error {
	_, err := s.manager.Estimate(s.session.ID, points)
	return err
}

func (s *roundScenario) openAnotherSession(_ string) error {
	_, err := s.manager.OpenSession(s.itemID)
	return err
}

func (s *roundScenario) failOnUpdate() error {
	s.repo.failUpdate = true
	return nil
}

// reveal revela la sesión y compara sus votos, como hace la API (RN-24.3).
func (s *roundScenario) reveal() error {
	s.revealed, s.revealErr = s.manager.Reveal(s.session.ID)
	if s.revealErr != nil {
		return fmt.Errorf("la revelación falló: %w", s.revealErr)
	}
	s.comparison = CompareVotes(s.revealed.Votes)
	return nil
}

func (s *roundScenario) repeat() error {
	return s.repeatSession(s.session.ID)
}

func (s *roundScenario) repeatSession(sessionID int) error {
	_, s.repeatErr = s.manager.RepeatRound(sessionID)
	return nil
}

// repeatConcurrently lanza n repeticiones de la misma sesión al mismo tiempo
// (CA-24.10).
func (s *roundScenario) repeatConcurrently(n int) error {
	s.repeatErrs = make([]error, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, s.repeatErrs[i] = s.manager.RepeatRound(s.session.ID)
		}()
	}
	close(start)
	wg.Wait()
	return nil
}

func (s *roundScenario) estimate(points int) error {
	_, s.estimateErr = s.manager.Estimate(s.session.ID, points)
	if s.estimateErr != nil {
		return fmt.Errorf("el registro de la estimación falló: %w", s.estimateErr)
	}
	return nil
}

func (s *roundScenario) hasDifferences() error {
	if !s.comparison.HasDifferences {
		return fmt.Errorf("comparación = %+v, se esperaban diferencias", s.comparison)
	}
	return nil
}

func (s *roundScenario) hasNoDifferences() error {
	if s.comparison.HasDifferences {
		return fmt.Errorf("comparación = %+v, no se esperaban diferencias", s.comparison)
	}
	return nil
}

func (s *roundScenario) lowest(points int, voters string) error {
	return expectVoteGroup("voto menor", s.comparison.Lowest, points, voters)
}

func (s *roundScenario) highest(points int, voters string) error {
	return expectVoteGroup("voto mayor", s.comparison.Highest, points, voters)
}

// expectVoteGroup compara un extremo con los puntos y los integrantes
// separados por coma.
func expectVoteGroup(what string, got VoteGroup, points int, voters string) error {
	want := strings.Split(voters, ", ")
	if got.Points != points || !slices.Equal(got.Voters, want) {
		return fmt.Errorf("%s = %+v, se esperaba %d de %v", what, got, points, want)
	}
	return nil
}

func (s *roundScenario) revealedVotes(table *godog.Table) error {
	return compareVotes("votos revelados", s.revealed.Votes, table)
}

func (s *roundScenario) repeatSucceeded() error {
	if s.repeatErr != nil {
		return fmt.Errorf("la repetición falló: %w", s.repeatErr)
	}
	return nil
}

func (s *roundScenario) onlyOneRepeatSucceeded() error {
	succeeded := 0
	for _, err := range s.repeatErrs {
		if err == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		return fmt.Errorf("repeticiones exitosas = %d, se esperaba 1", succeeded)
	}
	return nil
}

func (s *roundScenario) otherRepeatsNotRevealed() error {
	for _, err := range s.repeatErrs {
		if err != nil && !errors.Is(err, ErrSessionNotRevealed) {
			return fmt.Errorf("repetición rechazada con %v, se esperaba ErrSessionNotRevealed", err)
		}
	}
	return nil
}

func (s *roundScenario) repeatFailsNotRevealed() error {
	return expectRevealError(s.repeatErr, ErrSessionNotRevealed)
}

func (s *roundScenario) repeatFailsNotFound() error {
	return expectRevealError(s.repeatErr, ErrSessionNotFound)
}

func (s *roundScenario) repeatFailsAlreadyOpen() error {
	return expectRevealError(s.repeatErr, ErrSessionAlreadyOpen)
}

func (s *roundScenario) repeatFailsStorage() error {
	return expectRevealError(s.repeatErr, errRepository)
}

// stored lee la sesión del repositorio en memoria, sin pasar por los fallos
// forzados.
func (s *roundScenario) stored() (Session, error) {
	return s.repo.MemoryRepository.GetByID(s.session.ID)
}

func (s *roundScenario) storedStatusAndRound(status string, round int) error {
	stored, err := s.stored()
	if err != nil {
		return err
	}
	if stored.Status != Status(status) || stored.Round != round {
		return fmt.Errorf("sesión guardada = %+v, se esperaba %s en la ronda %d", stored, status, round)
	}
	return nil
}

func (s *roundScenario) storedWithoutVotes() error {
	stored, err := s.stored()
	if err != nil {
		return err
	}
	if len(stored.Votes) != 0 {
		return fmt.Errorf("votos guardados = %+v, se esperaba ninguno", stored.Votes)
	}
	return nil
}

func (s *roundScenario) storedVotes(table *godog.Table) error {
	stored, err := s.stored()
	if err != nil {
		return err
	}
	return compareVotes("votos guardados", stored.Votes, table)
}

func (s *roundScenario) onlySessionOfItem(_ string) error {
	count := 0
	for _, session := range s.repo.sessions {
		if session.ItemID == s.itemID {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("sesiones del ítem = %d, se esperaba 1", count)
	}
	return nil
}

func (s *roundScenario) itemUnchanged(_ string, status string, points int) error {
	it, err := s.items.GetItem(s.itemID)
	if err != nil {
		return err
	}
	if it.Status != backlog.Status(status) || it.StoryPoints != points {
		return fmt.Errorf("ítem = %+v, se esperaba estado %q con %d story points", it, status, points)
	}
	return nil
}

func (s *roundScenario) itemPoints(_ string, points int) error {
	it, err := s.items.GetItem(s.itemID)
	if err != nil {
		return err
	}
	if it.StoryPoints != points {
		return fmt.Errorf("story points del ítem = %d, se esperaban %d", it.StoryPoints, points)
	}
	return nil
}
