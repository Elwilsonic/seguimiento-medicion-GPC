package httpapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197), parte
// de sprints y Planning Poker (Buttini Cristobal).
//
// Los pasos "Dado" crean los sprints, los ítems y las sesiones con los Manager
// del dominio, como el runner de la base, y guardan su identificador público.
// Así los escenarios prueban solo las rutas de sprints.go, poker.go,
// metrics.go y sprint_reports.go (D3 de design.md), sin depender de cómo
// responden los ítems, que es la parte de Cardozo Leandro.

// HU-52 · BDD · Verifica que se cumplan los escenarios de features/HU-52-sprints.feature.
func TestFeatures_HU52Sprints(t *testing.T) {
	runSprintsPokerUUIDFeature(t, "../../features/HU-52-sprints.feature", "sprints")
}

// HU-52 · BDD · Verifica que se cumplan los escenarios de features/HU-52-planning-poker.feature.
func TestFeatures_HU52PlanningPoker(t *testing.T) {
	runSprintsPokerUUIDFeature(t, "../../features/HU-52-planning-poker.feature", "Planning Poker")
}

func runSprintsPokerUUIDFeature(t *testing.T, path, part string) {
	t.Helper()
	suite := godog.TestSuite{
		ScenarioInitializer: initializeSprintsPokerUUIDScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{path},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatalf("fallaron escenarios BDD de HU-52 (%s)", part)
	}
}

// sprintsPokerUUIDScenario reutiliza los pasos de HU-51 (código, error y
// campos de la respuesta) y suma los Manager para preparar los datos.
type sprintsPokerUUIDScenario struct {
	*uuidScenario
	projects *project.Manager
	items    *backlog.Manager
	sprints  *sprint.Manager
	poker    *poker.Manager
	// internal guarda el ID interno de cada entidad del escenario, por nombre.
	internal map[string]int
}

func initializeSprintsPokerUUIDScenario(ctx *godog.ScenarioContext) {
	s := &sprintsPokerUUIDScenario{uuidScenario: &uuidScenario{}}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.items = backlog.NewManager(backlog.NewMemoryRepository(), s.projects)
		s.sprints = sprint.NewManager(sprint.NewMemoryRepository(), s.projects, s.items)
		efforts := effort.NewManager(effort.NewMemoryRepository(), s.sprints, s.items)
		s.poker = poker.NewManager(poker.NewMemoryRepository(), s.items)
		defects := defect.NewManager(defect.NewMemoryRepository(), s.items, s.sprints)
		s.h = httpapi.NewHandler(s.projects, s.items, s.sprints, efforts, s.poker, defects, newMembers(s.projects))
		s.ids = map[string]string{}
		s.internal = map[string]int{}
		s.rec = nil
		return ctx, nil
	})

	ctx.Step(`^que existe el proyecto "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el proyecto "([^"]*)" tiene el sprint abierto "([^"]*)"$`, s.openSprint)
	ctx.Step(`^que el proyecto "([^"]*)" tiene el ítem "([^"]*)" sin sprint$`, s.itemWithoutSprint)
	ctx.Step(`^que el ítem "([^"]*)" tiene una sesión abierta$`, s.openSession)
	ctx.Step(`^que el ítem "([^"]*)" tiene una sesión abierta con los votos (\d+) y (\d+)$`, s.openSessionWithVotes)
	ctx.Step(`^que el ítem "([^"]*)" tiene una sesión revelada con los votos (\d+) y (\d+)$`, s.revealedSession)
	ctx.Step(`^pido (GET|POST|PUT|PATCH) "([^"]*)"$`, s.request)
	ctx.Step(`^pido (GET|POST|PUT|PATCH) "([^"]*)" con el cuerpo:$`, s.requestWithBody)
	ctx.Step(`^la respuesta tiene el código (\d+)$`, s.statusIs)
	ctx.Step(`^el error es "([^"]*)"$`, s.errorIs)
	ctx.Step(`^el "([^"]*)" de la respuesta es "([^"]*)"$`, s.fieldIs)
	ctx.Step(`^el "([^"]*)" de la respuesta es (\d+)$`, s.numberFieldIs)
	ctx.Step(`^el "([^"]*)" de la respuesta es un UUID versión 4 en minúsculas, como texto$`, s.fieldIsUUIDv4)
	ctx.Step(`^el "([^"]*)" de la respuesta no es un número ni el UUID nulo$`, s.fieldIsNotNumberNorNil)
	ctx.Step(`^el "([^"]*)" de la respuesta es el UUID de "([^"]*)"$`, s.fieldIsUUIDOf)
	ctx.Step(`^el ítem "([^"]*)" quedó asignado al sprint "([^"]*)"$`, s.itemAssignedTo)
}

// projectInternalID devuelve el ID interno del proyecto que el escenario creó
// por la API (HU-51), a partir de su UUID.
func (s *sprintsPokerUUIDScenario) projectInternalID(name string) (int, error) {
	id, err := uuid.Parse(s.ids[name])
	if err != nil {
		return 0, fmt.Errorf("el proyecto %q no tiene un UUID en el escenario: %v", name, err)
	}
	p, err := s.projects.GetByPublicID(id)
	if err != nil {
		return 0, err
	}
	return p.ID, nil
}

func (s *sprintsPokerUUIDScenario) openSprint(projectName, name string) error {
	projectID, err := s.projectInternalID(projectName)
	if err != nil {
		return err
	}
	sp, err := s.sprints.Create(projectID, name, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), "Gestionar el backlog")
	if err != nil {
		return err
	}
	s.ids[name], s.internal[name] = sp.PublicID.String(), sp.ID
	return nil
}

func (s *sprintsPokerUUIDScenario) itemWithoutSprint(projectName, title string) error {
	projectID, err := s.projectInternalID(projectName)
	if err != nil {
		return err
	}
	it, err := s.items.Create(projectID, title, "", "ALTA", []string{"Ingresa con su clave"})
	if err != nil {
		return err
	}
	s.ids[title], s.internal[title] = it.PublicID.String(), it.ID
	return nil
}

// sessionName es el nombre con el que los escenarios nombran la sesión de un
// ítem, por ejemplo "la sesión de Login".
func sessionName(item string) string {
	return "la sesión de " + item
}

func (s *sprintsPokerUUIDScenario) openSession(item string) error {
	itemID, ok := s.internal[item]
	if !ok {
		return fmt.Errorf("no existe el ítem %q en el escenario", item)
	}
	session, err := s.poker.OpenSession(itemID)
	if err != nil {
		return err
	}
	s.ids[sessionName(item)], s.internal[sessionName(item)] = session.PublicID.String(), session.ID
	return nil
}

func (s *sprintsPokerUUIDScenario) openSessionWithVotes(item string, first, second int) error {
	if err := s.openSession(item); err != nil {
		return err
	}
	sessionID := s.internal[sessionName(item)]
	if _, err := s.poker.Vote(sessionID, "Ana", first); err != nil {
		return err
	}
	_, err := s.poker.Vote(sessionID, "Bruno", second)
	return err
}

func (s *sprintsPokerUUIDScenario) revealedSession(item string, first, second int) error {
	if err := s.openSessionWithVotes(item, first, second); err != nil {
		return err
	}
	_, err := s.poker.Reveal(s.internal[sessionName(item)])
	return err
}

func (s *sprintsPokerUUIDScenario) request(method, path string) error {
	return s.requestBody(method, path, "")
}

func (s *sprintsPokerUUIDScenario) requestWithBody(method, path string, body *godog.DocString) error {
	return s.requestBody(method, path, body.Content)
}

func (s *sprintsPokerUUIDScenario) requestBody(method, path, body string) error {
	path, err := s.expand(path)
	if err != nil {
		return err
	}
	s.rec = s.serve(method, path, body)
	return nil
}

// expand reemplaza en la ruta {Nombre} por el UUID de la entidad Nombre,
// {Nombre en mayúsculas} por ese UUID en mayúsculas y {un UUID que nunca se
// asignó} por un UUID versión 4 nuevo.
func (s *sprintsPokerUUIDScenario) expand(path string) (string, error) {
	var missing string
	out := regexp.MustCompile(`\{([^}]+)\}`).ReplaceAllStringFunc(path, func(m string) string {
		name := m[1 : len(m)-1]
		if name == "un UUID que nunca se asignó" {
			return uuid.NewV4().String()
		}
		upper := strings.HasSuffix(name, " en mayúsculas")
		name = strings.TrimSuffix(name, " en mayúsculas")
		id, ok := s.ids[name]
		if !ok {
			missing = name
		}
		if upper {
			return strings.ToUpper(id)
		}
		return id
	})
	if missing != "" {
		return "", fmt.Errorf("no existe %q en el escenario", missing)
	}
	return out, nil
}

func (s *sprintsPokerUUIDScenario) numberFieldIs(field string, want int) error {
	raw, err := jsonField(s.rec, field)
	if err != nil {
		return err
	}
	var got int
	if err := json.Unmarshal(raw, &got); err != nil {
		return fmt.Errorf("%s = %s, se esperaba el número %d", field, raw, want)
	}
	if got != want {
		return fmt.Errorf("%s = %d, se esperaba %d", field, got, want)
	}
	return nil
}

// itemAssignedTo verifica en el dominio que la asignación quedó guardada.
func (s *sprintsPokerUUIDScenario) itemAssignedTo(item, sprintName string) error {
	it, err := s.items.GetItem(s.internal[item])
	if err != nil {
		return err
	}
	if it.SprintID != s.internal[sprintName] {
		return fmt.Errorf("el ítem %q quedó en el sprint %d, se esperaba %d (%q)", item, it.SprintID, s.internal[sprintName], sprintName)
	}
	return nil
}
