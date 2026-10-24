package poker

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-21-abrir-sesion-planning-poker.feature
// con godog.
func TestFeatures_HU21AbrirSesionPlanningPoker(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializePokerScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-21-abrir-sesion-planning-poker.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-21")
	}
}

// pokerScenario guarda el estado de un escenario de HU-21.
type pokerScenario struct {
	projects  *project.Manager
	items     *backlog.Manager
	repo      *MemoryRepository
	manager   *Manager
	projectID int

	itemsByTitle map[string]backlog.Item
	sessions     []Session
	errs         []error
	cards        []int
}

func initializePokerScenario(ctx *godog.ScenarioContext) {
	s := &pokerScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.items = backlog.NewManager(backlog.NewMemoryRepository(), s.projects)
		s.repo = NewMemoryRepository()
		s.manager = NewManager(s.repo, s.items)
		s.projectID = 0
		s.itemsByTitle = map[string]backlog.Item{}
		s.sessions = nil
		s.errs = nil
		s.cards = nil
		return ctx, nil
	})

	ctx.Step(`^un proyecto de prueba$`, s.createProject)
	ctx.Step(`^un ítem "([^"]*)" en estado "([^"]*)" con (\d+) story points$`, s.createItem)
	ctx.Step(`^que ya hay una sesión abierta para el ítem "([^"]*)"$`, s.openExistingSession)
	ctx.Step(`^abro una sesión de Planning Poker para el ítem "([^"]*)"$`, s.openByTitle)
	ctx.Step(`^abro una sesión de Planning Poker para el ítem con el ID (\d+)$`, s.openByID)
	ctx.Step(`^la apertura es exitosa$`, s.openOK)
	ctx.Step(`^la sesión está asociada al ítem "([^"]*)"$`, s.sessionBelongsTo)
	ctx.Step(`^la sesión está abierta, en la ronda 1 y sin votos$`, s.sessionIsFresh)
	ctx.Step(`^las dos aperturas son exitosas$`, s.bothOpenOK)
	ctx.Step(`^las dos sesiones están abiertas y tienen IDs distintos$`, s.sessionsAreOpenAndDistinct)
	ctx.Step(`^la apertura falla porque ya hay una sesión abierta$`, s.failsAlreadyOpen)
	ctx.Step(`^queda una sola sesión abierta para el ítem "([^"]*)"$`, s.onlyOneSession)
	ctx.Step(`^la apertura falla porque el ítem no existe$`, s.failsItemNotFound)
	ctx.Step(`^la apertura falla porque el ítem está terminado$`, s.failsItemFinished)
	ctx.Step(`^consulto las cartas de Planning Poker$`, s.readCards)
	ctx.Step(`^las cartas son la escala de HU-20$`, s.cardsAreScale)
	ctx.Step(`^modifico la lista de cartas devuelta$`, s.mutateCards)
	ctx.Step(`^las cartas siguen siendo la escala de HU-20$`, s.cardsStillAreScale)
}

func (s *pokerScenario) createProject() error {
	p, err := s.projects.Create("Proyecto de prueba", "Proyecto para los escenarios de HU-21")
	if err != nil {
		return fmt.Errorf("crear el proyecto de prueba: %w", err)
	}
	s.projectID = p.ID
	return nil
}

func (s *pokerScenario) createItem(title, status string, points int) error {
	it, err := s.items.Create(s.projectID, title, "Descripción de "+title, "MEDIA", []string{"Criterio de " + title})
	if err != nil {
		return fmt.Errorf("crear el ítem %q: %w", title, err)
	}
	if points > 0 {
		it, err = s.items.AssignStoryPoints(it.ID, points)
		if err != nil {
			return fmt.Errorf("asignar story points al ítem %q: %w", title, err)
		}
	}
	switch backlog.Status(status) {
	case backlog.StatusInProgress:
		it, err = s.items.ChangeStatus(it.ID, string(backlog.StatusInProgress))
	case backlog.StatusFinished:
		if _, err = s.items.ChangeStatus(it.ID, string(backlog.StatusInProgress)); err == nil {
			it, err = s.items.ChangeStatus(it.ID, string(backlog.StatusFinished))
		}
	}
	if err != nil {
		return fmt.Errorf("pasar el ítem %q a %s: %w", title, status, err)
	}
	s.itemsByTitle[title] = it
	return nil
}

func (s *pokerScenario) openExistingSession(title string) error {
	it, ok := s.itemsByTitle[title]
	if !ok {
		return fmt.Errorf("no hay ningún ítem llamado %q", title)
	}
	if _, err := s.manager.OpenSession(it.ID); err != nil {
		return fmt.Errorf("abrir la sesión previa del ítem %q: %w", title, err)
	}
	return nil
}

func (s *pokerScenario) openByTitle(title string) error {
	it, ok := s.itemsByTitle[title]
	if !ok {
		return fmt.Errorf("no hay ningún ítem llamado %q", title)
	}
	return s.openByID(it.ID)
}

func (s *pokerScenario) openByID(id int) error {
	session, err := s.manager.OpenSession(id)
	s.sessions = append(s.sessions, session)
	s.errs = append(s.errs, err)
	return nil
}

func (s *pokerScenario) lastSession() (Session, error) {
	if len(s.sessions) == 0 || len(s.errs) == 0 {
		return Session{}, errors.New("no se abrió ninguna sesión")
	}
	return s.sessions[len(s.sessions)-1], s.errs[len(s.errs)-1]
}

func (s *pokerScenario) openOK() error {
	_, err := s.lastSession()
	if err != nil {
		return fmt.Errorf("la apertura falló: %w", err)
	}
	return nil
}

func (s *pokerScenario) sessionBelongsTo(title string) error {
	session, err := s.lastSession()
	if err != nil {
		return err
	}
	it, ok := s.itemsByTitle[title]
	if !ok {
		return fmt.Errorf("no hay ningún ítem llamado %q", title)
	}
	if session.ItemID != it.ID {
		return fmt.Errorf("la sesión está asociada al ítem %d, se esperaba el ítem %d (%q)", session.ItemID, it.ID, title)
	}
	return nil
}

func (s *pokerScenario) sessionIsFresh() error {
	session, err := s.lastSession()
	if err != nil {
		return err
	}
	if session.Status != StatusOpen {
		return fmt.Errorf("estado = %q, se esperaba %q", session.Status, StatusOpen)
	}
	if session.Round != 1 {
		return fmt.Errorf("ronda = %d, se esperaba 1", session.Round)
	}
	if session.Votes == nil {
		return errors.New("los votos son nil, se esperaba una lista vacía")
	}
	if len(session.Votes) != 0 {
		return fmt.Errorf("votos = %+v, se esperaba una sesión sin votos", session.Votes)
	}
	return nil
}

func (s *pokerScenario) bothOpenOK() error {
	if len(s.errs) != 2 {
		return fmt.Errorf("aperturas = %d, se esperaban 2", len(s.errs))
	}
	for i, err := range s.errs {
		if err != nil {
			return fmt.Errorf("la apertura %d falló: %w", i+1, err)
		}
	}
	return nil
}

func (s *pokerScenario) sessionsAreOpenAndDistinct() error {
	if len(s.sessions) != 2 {
		return fmt.Errorf("sesiones = %d, se esperaban 2", len(s.sessions))
	}
	for i, session := range s.sessions {
		if session.Status != StatusOpen {
			return fmt.Errorf("la sesión %d está en estado %q, se esperaba %q", i+1, session.Status, StatusOpen)
		}
	}
	if s.sessions[0].ID == s.sessions[1].ID {
		return fmt.Errorf("las dos sesiones tienen el mismo ID %d", s.sessions[0].ID)
	}
	return nil
}

func (s *pokerScenario) failsAlreadyOpen() error {
	_, err := s.lastSession()
	if !errors.Is(err, ErrSessionAlreadyOpen) {
		return fmt.Errorf("error = %v, se esperaba poker.ErrSessionAlreadyOpen", err)
	}
	return nil
}

func (s *pokerScenario) onlyOneSession(title string) error {
	it, ok := s.itemsByTitle[title]
	if !ok {
		return fmt.Errorf("no hay ningún ítem llamado %q", title)
	}
	if got := countOpenSessions(s.repo, it.ID); got != 1 {
		return fmt.Errorf("sesiones abiertas = %d, se esperaba 1", got)
	}
	return nil
}

func (s *pokerScenario) failsItemNotFound() error {
	_, err := s.lastSession()
	if !errors.Is(err, backlog.ErrItemNotFound) {
		return fmt.Errorf("error = %v, se esperaba backlog.ErrItemNotFound", err)
	}
	return nil
}

func (s *pokerScenario) failsItemFinished() error {
	_, err := s.lastSession()
	if !errors.Is(err, ErrItemFinished) {
		return fmt.Errorf("error = %v, se esperaba poker.ErrItemFinished", err)
	}
	return nil
}

func (s *pokerScenario) readCards() error {
	s.cards = Cards()
	return nil
}

func (s *pokerScenario) cardsAreScale() error {
	want := backlog.StoryPointScale()
	if len(s.cards) != len(want) {
		return fmt.Errorf("cartas = %v, se esperaba %v", s.cards, want)
	}
	for i := range want {
		if s.cards[i] != want[i] {
			return fmt.Errorf("cartas = %v, se esperaba %v", s.cards, want)
		}
	}
	return nil
}

func (s *pokerScenario) mutateCards() error {
	if len(s.cards) == 0 {
		return errors.New("la lista de cartas está vacía")
	}
	s.cards[0] = 99
	return nil
}

func (s *pokerScenario) cardsStillAreScale() error {
	// La comprobación vuelve a llamar a Cards(): así se verifica que la lista
	// devuelta antes no comparte el estado con la escala.
	s.cards = Cards()
	return s.cardsAreScale()
}
