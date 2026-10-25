package poker

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-23 · Revelar las estimaciones (Issue #41). Los pasos llaman al Manager,
// no al handler HTTP; la API se prueba en internal/httpapi.
func TestFeatures_HU23RevelarEstimaciones(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeRevealScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-23-revelar-estimaciones.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-23")
	}
}

type revealScenario struct {
	projects  *project.Manager
	items     *backlog.Manager
	repo      *MemoryRepository
	manager   *Manager
	projectID int
	itemID    int
	session   Session
	revealed  Session
	revealErr error
	voteErr   error
	accepted  []string
	rejected  []error
}

func initializeRevealScenario(ctx *godog.ScenarioContext) {
	s := &revealScenario{}
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		*s = revealScenario{}
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
	ctx.Step(`^se revelan los votos de la sesión$`, s.reveal)
	ctx.Step(`^se revelan los votos de la sesión (\d+)$`, s.revealSession)
	ctx.Step(`^"([^"]*)" vota (\d+) puntos$`, s.vote)
	ctx.Step(`^(\d+) integrantes votan mientras se revelan los votos de la sesión$`, s.voteWhileRevealing)
	ctx.Step(`^la revelación es exitosa$`, s.revealSucceeded)
	ctx.Step(`^la sesión revelada está en estado "([^"]*)"$`, s.revealedStatus)
	ctx.Step(`^se muestran los votos:$`, s.revealedVotes)
	ctx.Step(`^la sesión revelada sigue en la ronda (\d+) del ítem "([^"]*)"$`, s.revealedRoundAndItem)
	ctx.Step(`^el ítem "([^"]*)" sigue en estado "([^"]*)" con (\d+) story points$`, s.itemUnchanged)
	ctx.Step(`^la revelación falla porque la sesión no tiene votos$`, s.revealFailsNoVotes)
	ctx.Step(`^la revelación falla porque la sesión no está abierta$`, s.revealFailsNotOpen)
	ctx.Step(`^la revelación falla porque la sesión no existe$`, s.revealFailsNotFound)
	ctx.Step(`^la sesión guardada está en estado "([^"]*)"$`, s.storedStatus)
	ctx.Step(`^el voto falla porque la sesión no está abierta$`, s.voteFailsNotOpen)
	ctx.Step(`^los votos guardados son:$`, s.storedVotes)
	ctx.Step(`^cada voto aceptado está entre los votos guardados$`, s.acceptedVotesStored)
	ctx.Step(`^cada voto rechazado falló porque la sesión no está abierta$`, s.rejectedVotesNotOpen)
}

func (s *revealScenario) createProject() error {
	p, err := s.projects.Create("Proyecto de prueba", "Proyecto para HU-23")
	if err != nil {
		return err
	}
	s.projectID = p.ID
	return nil
}

func (s *revealScenario) createItem(title, status string, points int) error {
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

func (s *revealScenario) openSession(_ string) error {
	session, err := s.manager.OpenSession(s.itemID)
	if err != nil {
		return err
	}
	s.session = session
	return nil
}

func (s *revealScenario) previousVote(voter string, points int) error {
	_, err := s.manager.Vote(s.session.ID, voter, points)
	return err
}

func (s *revealScenario) previousReveal() error {
	_, err := s.manager.Reveal(s.session.ID)
	return err
}

func (s *revealScenario) reveal() error {
	return s.revealSession(s.session.ID)
}

func (s *revealScenario) revealSession(sessionID int) error {
	s.revealed, s.revealErr = s.manager.Reveal(sessionID)
	return nil
}

func (s *revealScenario) vote(voter string, points int) error {
	_, s.voteErr = s.manager.Vote(s.session.ID, voter, points)
	return nil
}

// voteWhileRevealing lanza n votos de integrantes distintos y una revelación
// al mismo tiempo, y separa los votos aceptados de los rechazados (CA-23.7).
func (s *revealScenario) voteWhileRevealing(n int) error {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	start := make(chan struct{})
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(voter string) {
			defer wg.Done()
			<-start
			_, err := s.manager.Vote(s.session.ID, voter, 8)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				s.accepted = append(s.accepted, voter)
			} else {
				s.rejected = append(s.rejected, err)
			}
		}("Integrante " + strconv.Itoa(i))
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		revealed, err := s.manager.Reveal(s.session.ID)
		mu.Lock()
		defer mu.Unlock()
		s.revealed, s.revealErr = revealed, err
	}()
	close(start)
	wg.Wait()
	return nil
}

func (s *revealScenario) revealSucceeded() error {
	if s.revealErr != nil {
		return fmt.Errorf("la revelación falló: %w", s.revealErr)
	}
	return nil
}

func (s *revealScenario) revealedStatus(status string) error {
	if s.revealed.Status != Status(status) {
		return fmt.Errorf("estado devuelto = %q, se esperaba %q", s.revealed.Status, status)
	}
	return nil
}

func (s *revealScenario) revealedVotes(table *godog.Table) error {
	return compareVotes("votos revelados", s.revealed.Votes, table)
}

func (s *revealScenario) revealedRoundAndItem(round int, _ string) error {
	if s.revealed.Round != round || s.revealed.ItemID != s.itemID {
		return fmt.Errorf("sesión revelada = %+v, se esperaba la ronda %d del ítem %d", s.revealed, round, s.itemID)
	}
	return nil
}

func (s *revealScenario) itemUnchanged(_ string, status string, points int) error {
	it, err := s.items.GetItem(s.itemID)
	if err != nil {
		return err
	}
	if it.Status != backlog.Status(status) || it.StoryPoints != points {
		return fmt.Errorf("ítem = %+v, se esperaba estado %q con %d story points", it, status, points)
	}
	return nil
}

func (s *revealScenario) revealFailsNoVotes() error {
	return expectRevealError(s.revealErr, ErrNoVotes)
}

func (s *revealScenario) revealFailsNotOpen() error {
	return expectRevealError(s.revealErr, ErrSessionNotOpen)
}

func (s *revealScenario) revealFailsNotFound() error {
	return expectRevealError(s.revealErr, ErrSessionNotFound)
}

func (s *revealScenario) voteFailsNotOpen() error {
	return expectRevealError(s.voteErr, ErrSessionNotOpen)
}

func expectRevealError(got, want error) error {
	if !errors.Is(got, want) {
		return fmt.Errorf("error = %v, se esperaba %v", got, want)
	}
	return nil
}

func (s *revealScenario) storedStatus(status string) error {
	stored, err := s.repo.GetByID(s.session.ID)
	if err != nil {
		return err
	}
	if stored.Status != Status(status) {
		return fmt.Errorf("estado guardado = %q, se esperaba %q", stored.Status, status)
	}
	return nil
}

func (s *revealScenario) storedVotes(table *godog.Table) error {
	stored, err := s.repo.GetByID(s.session.ID)
	if err != nil {
		return err
	}
	return compareVotes("votos guardados", stored.Votes, table)
}

func (s *revealScenario) acceptedVotesStored() error {
	stored, err := s.repo.GetByID(s.session.ID)
	if err != nil {
		return err
	}
	for _, voter := range s.accepted {
		found := false
		for _, v := range stored.Votes {
			if v.Voter == voter {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("el voto aceptado de %q no está entre los guardados %+v", voter, stored.Votes)
		}
	}
	// El voto previo de Ana más los votos aceptados.
	if len(stored.Votes) != 1+len(s.accepted) {
		return fmt.Errorf("votos guardados = %d, se esperaban %d", len(stored.Votes), 1+len(s.accepted))
	}
	return nil
}

func (s *revealScenario) rejectedVotesNotOpen() error {
	for _, err := range s.rejected {
		if !errors.Is(err, ErrSessionNotOpen) {
			return fmt.Errorf("voto rechazado con %v, se esperaba ErrSessionNotOpen", err)
		}
	}
	return nil
}

// compareVotes compara los votos con la tabla (integrante, puntos), en orden.
func compareVotes(what string, votes []Vote, table *godog.Table) error {
	rows := table.Rows[1:]
	if len(votes) != len(rows) {
		return fmt.Errorf("%s = %+v, se esperaban %d", what, votes, len(rows))
	}
	for i, row := range rows {
		points, err := strconv.Atoi(row.Cells[1].Value)
		if err != nil {
			return fmt.Errorf("puntos inválidos %q: %w", row.Cells[1].Value, err)
		}
		want := Vote{Voter: row.Cells[0].Value, Points: points}
		if votes[i] != want {
			return fmt.Errorf("%s[%d] = %+v, se esperaba %+v", what, i, votes[i], want)
		}
	}
	return nil
}
