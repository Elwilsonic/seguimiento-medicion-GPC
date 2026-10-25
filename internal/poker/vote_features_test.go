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

func TestFeatures_HU22VotarOculto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeVoteScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-22-votar-oculto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-22")
	}
}

type voteScenario struct {
	projects  *project.Manager
	items     *backlog.Manager
	repo      *MemoryRepository
	manager   *Manager
	projectID int
	itemID    int
	session   Session
	err       error
}

func initializeVoteScenario(ctx *godog.ScenarioContext) {
	s := &voteScenario{}
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.items = backlog.NewManager(backlog.NewMemoryRepository(), s.projects)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.items)
		s.session = Session{}
		s.err = nil
		return ctx, nil
	})
	ctx.Step(`^un proyecto de prueba$`, s.createProject)
	ctx.Step(`^un ítem "([^"]*)" en estado "([^"]*)" con (\d+) story points$`, s.createItem)
	ctx.Step(`^una sesión abierta para el ítem "([^"]*)"$`, s.openSession)
	ctx.Step(`^"([^"]*)" vota (\d+) puntos$`, s.vote)
	ctx.Step(`^que "([^"]*)" votó (\d+) puntos$`, s.previousVote)
	ctx.Step(`^"([^"]*)" vota (\d+) puntos en la sesión (\d+)$`, s.voteInSession)
	ctx.Step(`^"([^"]*)" vota simultáneamente (\d+) puntos$`, s.voteConcurrent)
	ctx.Step(`^"([^"]*)" y "([^"]*)" votan simultáneamente (\d+) puntos$`, s.voteTwoConcurrent)
	ctx.Step(`^"([^"]*)" vota (\d+) puntos mediante la API$`, s.vote)
	ctx.Step(`^la sesión tiene un estado distinto de "ABIERTA"$`, s.closeSession)
	ctx.Step(`^el voto es exitoso$`, s.voteSucceeded)
	ctx.Step(`^la sesión sigue sin mostrar votos$`, s.votesHidden)
	ctx.Step(`^la sesión sigue teniendo un solo voto guardado$`, s.oneStoredVote)
	ctx.Step(`^el nombre guardado es "([^"]*)"$`, s.storedName)
	ctx.Step(`^el voto falla porque la carta no es válida$`, s.failsInvalidCard)
	ctx.Step(`^el voto falla porque el votante está vacío$`, s.failsEmptyVoter)
	ctx.Step(`^el voto falla porque la sesión no existe$`, s.failsNotFound)
	ctx.Step(`^el voto falla porque la sesión no está abierta$`, s.failsNotOpen)
	ctx.Step(`^la sesión no tiene votos guardados$`, s.noStoredVotes)
	ctx.Step(`^ambos votos quedan guardados$`, s.twoStoredVotes)
	ctx.Step(`^la respuesta es 200 y no muestra votos$`, s.voteSucceeded)
	ctx.Step(`^que la sesión tiene un estado distinto de "ABIERTA"$`, s.closeSession)
}

func (s *voteScenario) createProject() error {
	p, err := s.projects.Create("Proyecto de prueba", "Proyecto para HU-22")
	if err != nil {
		return err
	}
	s.projectID = p.ID
	return nil
}

func (s *voteScenario) createItem(title, status string, points int) error {
	it, err := s.items.Create(s.projectID, title, "Descripción", "MEDIA", []string{"Criterio"})
	if err != nil {
		return err
	}
	if points > 0 {
		if _, err = s.items.AssignStoryPoints(it.ID, points); err != nil {
			return err
		}
	}
	if status == string(backlog.StatusInProgress) {
		it, err = s.items.ChangeStatus(it.ID, status)
	} else if status == string(backlog.StatusFinished) {
		if _, err = s.items.ChangeStatus(it.ID, string(backlog.StatusInProgress)); err == nil {
			it, err = s.items.ChangeStatus(it.ID, status)
		}
	}
	if err != nil {
		return err
	}
	s.itemID = it.ID
	return nil
}

func (s *voteScenario) openSession(_ string) error {
	s.session, s.err = s.manager.OpenSession(s.itemID)
	return s.err
}

func (s *voteScenario) vote(voter string, points int) error {
	sessionID := s.session.ID
	updated, err := s.manager.Vote(sessionID, voter, points)
	s.err = err
	if err == nil {
		s.session = updated
	}
	return nil
}

func (s *voteScenario) previousVote(voter string, points int) error {
	return s.vote(voter, points)
}

func (s *voteScenario) voteInSession(voter string, points, sessionID int) error {
	updated, err := s.manager.Vote(sessionID, voter, points)
	s.err = err
	if err == nil {
		s.session = updated
	}
	return nil
}

func (s *voteScenario) voteConcurrent(voter string, points int) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, s.err = s.manager.Vote(s.session.ID, voter, points)
	}()
	wg.Wait()
	return nil
}

func (s *voteScenario) voteTwoConcurrent(first, second string, points int) error {
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, voter := range []string{first, second} {
		wg.Add(1)
		go func(voter string) {
			defer wg.Done()
			_, err := s.manager.Vote(s.session.ID, voter, points)
			errs <- err
		}(voter)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			s.err = err
			return nil
		}
	}
	return nil
}

func (s *voteScenario) closeSession() error {
	s.session.Status = Status("REVELADA")
	_, s.err = s.repo.Update(s.session)
	return s.err
}

func (s *voteScenario) voteSucceeded() error {
	if s.err != nil {
		return fmt.Errorf("voto falló: %w", s.err)
	}
	return nil
}

func (s *voteScenario) votesHidden() error {
	if s.session.Votes == nil || len(s.session.Votes) != 0 {
		return fmt.Errorf("votos = %+v, se esperaba ocultamiento", s.session.Votes)
	}
	return nil
}

func (s *voteScenario) oneStoredVote() error {
	stored, err := s.repo.GetByID(s.session.ID)
	if err != nil {
		return err
	}
	if len(stored.Votes) != 1 {
		return fmt.Errorf("votos guardados = %+v, se esperaba uno", stored.Votes)
	}
	return nil
}

func (s *voteScenario) storedName(name string) error {
	stored, err := s.repo.GetByID(s.session.ID)
	if err != nil {
		return err
	}
	if stored.Votes[0].Voter != name {
		return fmt.Errorf("votante guardado = %q, se esperaba %q", stored.Votes[0].Voter, name)
	}
	return nil
}

func (s *voteScenario) failsInvalidCard() error {
	return expectVoteError(s.err, ErrInvalidCard)
}

func (s *voteScenario) failsEmptyVoter() error {
	return expectVoteError(s.err, ErrEmptyVoter)
}

func (s *voteScenario) failsNotFound() error {
	return expectVoteError(s.err, ErrSessionNotFound)
}

func (s *voteScenario) failsNotOpen() error {
	return expectVoteError(s.err, ErrSessionNotOpen)
}

func expectVoteError(got, want error) error {
	if !errors.Is(got, want) {
		return fmt.Errorf("error = %v, se esperaba %v", got, want)
	}
	return nil
}

func (s *voteScenario) noStoredVotes() error {
	stored, err := s.repo.GetByID(s.session.ID)
	if err != nil {
		return err
	}
	if len(stored.Votes) != 0 {
		return fmt.Errorf("votos guardados = %+v, se esperaban cero", stored.Votes)
	}
	return nil
}

func (s *voteScenario) twoStoredVotes() error {
	stored, err := s.repo.GetByID(s.session.ID)
	if err != nil {
		return err
	}
	if len(stored.Votes) != 2 {
		return fmt.Errorf("votos guardados = %+v, se esperaban dos", stored.Votes)
	}
	return nil
}
