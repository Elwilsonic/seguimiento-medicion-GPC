package httpapi_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"
	"uuid"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/member"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197). El
// runner de la base está en este paquete porque es el único que ya usa los
// Manager de las 6 entidades; los pasos llaman a los Manager, no a la API.

// uuidV4Lower es la forma canónica de un UUID versión 4 en minúsculas
// (CA-52.1).
var uuidV4Lower = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// HU-52 · BDD · Verifica que se cumplan los escenarios de features/HU-52-base.feature.
func TestFeatures_HU52Base(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeUUIDBaseScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-52-base.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-52 (base)")
	}
}

// baseEntity es lo que el escenario necesita de una entidad creada: su ID
// interno y su identificador público.
type baseEntity struct {
	id       int
	publicID uuid.UUID
}

// uuidBaseScenario guarda el estado de un escenario de la base de HU-52.
type uuidBaseScenario struct {
	projects  *project.Manager
	items     *backlog.Manager
	sprints   *sprint.Manager
	members   *member.Manager
	defects   *defect.Manager
	efforts   *effort.Manager
	poker     *poker.Manager
	projectID int
	item      backlog.Item
	sprint    sprint.Sprint
	n         int
	created   []baseEntity
	modified  baseEntity
	found     baseEntity
	err       error
}

func initializeUUIDBaseScenario(ctx *godog.ScenarioContext) {
	s := &uuidBaseScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*s = uuidBaseScenario{}
		s.projects = project.NewManager(project.NewMemoryRepository())
		s.items = backlog.NewManager(backlog.NewMemoryRepository(), s.projects)
		s.sprints = sprint.NewManager(sprint.NewMemoryRepository(), s.projects, s.items)
		s.members = member.NewManager(member.NewMemoryRepository(), s.projects)
		s.defects = defect.NewManager(defect.NewMemoryRepository(), s.items, s.sprints)
		s.efforts = effort.NewManager(effort.NewMemoryRepository(), s.sprints, s.items)
		s.poker = poker.NewManager(poker.NewMemoryRepository(), s.items)
		return ctx, nil
	})

	ctx.Step(`^que existe el proyecto "([^"]*)" con un ítem en curso asignado a un sprint abierto$`, s.baseProject)
	ctx.Step(`^(?:que )?cre(?:o|é) (un ítem|un sprint|un integrante|un defecto|un esfuerzo|una sesión)$`, s.createOne)
	ctx.Step(`^creo (\d+) veces (un ítem|un sprint|un integrante|un defecto|un esfuerzo|una sesión)$`, s.createMany)
	ctx.Step(`^(?:un ítem|un sprint|un integrante|un defecto|un esfuerzo|una sesión) creado tiene un identificador público UUID versión 4 en minúsculas$`, s.isCanonicalV4)
	ctx.Step(`^no es el UUID nulo$`, s.isNotNil)
	ctx.Step(`^los (\d+) identificadores públicos son distintos$`, s.allDistinct)
	ctx.Step(`^busco (un ítem|un sprint|un integrante|un defecto|un esfuerzo|una sesión) por su identificador público$`, s.findCreated)
	ctx.Step(`^busco (un ítem|un sprint|un integrante|un defecto|un esfuerzo|una sesión) por el UUID nulo$`, s.findNil)
	ctx.Step(`^busco (un ítem|un sprint|un integrante|un defecto|un esfuerzo|una sesión) por un UUID versión 4 que nunca se asignó$`, s.findMissing)
	ctx.Step(`^obtengo (?:un ítem|un sprint|un integrante|un defecto|un esfuerzo|una sesión) creado$`, s.foundCreated)
	ctx.Step(`^modifico (un ítem|un sprint|un integrante|un defecto|un esfuerzo|una sesión)$`, s.modify)
	ctx.Step(`^(?:un ítem|un sprint|un integrante|un defecto|un esfuerzo|una sesión) modificado conserva su identificador público$`, s.modifiedKeepsPublicID)
	ctx.Step(`^la búsqueda falla con "([^"]*)"$`, s.findFailsWith)
}

func (s *uuidBaseScenario) baseProject(name string) error {
	p, err := s.projects.Create(name, "")
	if err != nil {
		return err
	}
	s.projectID = p.ID
	if s.item, err = s.items.Create(p.ID, "Login de usuarios", "", "ALTA", []string{"Login OK"}); err != nil {
		return err
	}
	if s.sprint, err = s.sprints.Create(p.ID, "Sprint 1", time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), "Entregar el MVP"); err != nil {
		return err
	}
	if _, err = s.sprints.AssignItem(s.sprint.ID, s.item.ID); err != nil {
		return err
	}
	_, err = s.items.ChangeStatus(s.item.ID, "EN_CURSO")
	return err
}

// create crea una entidad del tipo kind con el Manager de su paquete. Cada
// llamada usa datos nuevos (s.n), para que se puedan crear muchas.
func (s *uuidBaseScenario) create(kind string) (baseEntity, error) {
	s.n++
	switch kind {
	case "un ítem":
		it, err := s.items.Create(s.projectID, fmt.Sprintf("Ítem %d", s.n), "", "MEDIA", []string{"Criterio"})
		return baseEntity{it.ID, it.PublicID}, err
	case "un sprint":
		// Un proyecto nuevo por sprint, porque solo puede haber uno ABIERTO.
		p, err := s.projects.Create(fmt.Sprintf("Proyecto %d", s.n), "")
		if err != nil {
			return baseEntity{}, err
		}
		sp, err := s.sprints.Create(p.ID, "Sprint 1", time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), "Objetivo")
		return baseEntity{sp.ID, sp.PublicID}, err
	case "un integrante":
		m, err := s.members.Register(s.projectID, fmt.Sprintf("Integrante %d", s.n), "PRODUCT_BUILDER")
		return baseEntity{m.ID, m.PublicID}, err
	case "un defecto":
		d, err := s.defects.Register(s.item.ID, fmt.Sprintf("Defecto %d", s.n), "ALTA", s.sprint.ID)
		return baseEntity{d.ID, d.PublicID}, err
	case "un esfuerzo":
		e, err := s.efforts.Register(s.sprint.ID, s.item.ID, "Ana Pérez", time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC), fmt.Sprintf("Tarea %d", s.n), 1)
		return baseEntity{e.ID, e.PublicID}, err
	case "una sesión":
		// Un ítem nuevo por sesión, porque solo puede haber una ABIERTA por ítem.
		it, err := s.items.Create(s.projectID, fmt.Sprintf("Historia %d", s.n), "", "MEDIA", []string{"Criterio"})
		if err != nil {
			return baseEntity{}, err
		}
		ses, err := s.poker.OpenSession(it.ID)
		return baseEntity{ses.ID, ses.PublicID}, err
	}
	return baseEntity{}, fmt.Errorf("entidad desconocida %q", kind)
}

// find busca una entidad del tipo kind por su identificador público.
func (s *uuidBaseScenario) find(kind string, id uuid.UUID) (baseEntity, error) {
	switch kind {
	case "un ítem":
		v, err := s.items.GetByPublicID(id)
		return baseEntity{v.ID, v.PublicID}, err
	case "un sprint":
		v, err := s.sprints.GetByPublicID(id)
		return baseEntity{v.ID, v.PublicID}, err
	case "un integrante":
		v, err := s.members.GetByPublicID(id)
		return baseEntity{v.ID, v.PublicID}, err
	case "un defecto":
		v, err := s.defects.GetByPublicID(id)
		return baseEntity{v.ID, v.PublicID}, err
	case "un esfuerzo":
		v, err := s.efforts.GetByPublicID(id)
		return baseEntity{v.ID, v.PublicID}, err
	case "una sesión":
		v, err := s.poker.GetByPublicID(id)
		return baseEntity{v.ID, v.PublicID}, err
	}
	return baseEntity{}, fmt.Errorf("entidad desconocida %q", kind)
}

func (s *uuidBaseScenario) createOne(kind string) error {
	e, err := s.create(kind)
	if err != nil {
		return fmt.Errorf("crear %s: %w", kind, err)
	}
	s.created = append(s.created, e)
	return nil
}

func (s *uuidBaseScenario) createMany(n int, kind string) error {
	for range n {
		if err := s.createOne(kind); err != nil {
			return err
		}
	}
	return nil
}

func (s *uuidBaseScenario) last() baseEntity {
	if len(s.created) == 0 {
		return baseEntity{}
	}
	return s.created[len(s.created)-1]
}

func (s *uuidBaseScenario) isCanonicalV4() error {
	if got := s.last().publicID.String(); !uuidV4Lower.MatchString(got) {
		return fmt.Errorf("identificador público = %q, se esperaba un UUID versión 4 en minúsculas", got)
	}
	return nil
}

func (s *uuidBaseScenario) isNotNil() error {
	if s.last().publicID == uuid.Nil() {
		return fmt.Errorf("el identificador público es el UUID nulo")
	}
	return nil
}

func (s *uuidBaseScenario) allDistinct(n int) error {
	seen := map[uuid.UUID]bool{}
	for _, e := range s.created {
		seen[e.publicID] = true
	}
	if len(s.created) != n || len(seen) != n {
		return fmt.Errorf("hay %d identificadores distintos entre %d creados, se esperaban %d", len(seen), len(s.created), n)
	}
	return nil
}

func (s *uuidBaseScenario) findCreated(kind string) error {
	s.found, s.err = s.find(kind, s.last().publicID)
	return nil
}

func (s *uuidBaseScenario) findNil(kind string) error {
	s.found, s.err = s.find(kind, uuid.Nil())
	return nil
}

func (s *uuidBaseScenario) findMissing(kind string) error {
	s.found, s.err = s.find(kind, uuid.NewV4())
	return nil
}

func (s *uuidBaseScenario) foundCreated() error {
	if s.err != nil {
		return fmt.Errorf("la búsqueda falló: %v", s.err)
	}
	if want := s.last(); s.found != want {
		return fmt.Errorf("se encontró %+v, se esperaba %+v", s.found, want)
	}
	return nil
}

func (s *uuidBaseScenario) modify(kind string) error {
	id := s.last().id
	var got baseEntity
	var err error
	switch kind {
	case "un ítem":
		var v backlog.Item
		v, err = s.items.Update(id, "Título nuevo", "", "ALTA", []string{"Criterio nuevo"}, 2)
		got = baseEntity{v.ID, v.PublicID}
	case "un sprint":
		var v sprint.Sprint
		v, err = s.sprints.Close(id)
		got = baseEntity{v.ID, v.PublicID}
	case "un integrante":
		var v member.Member
		v, err = s.members.Update(id, "Nombre nuevo", "AGILE_ENABLER")
		got = baseEntity{v.ID, v.PublicID}
	case "un defecto":
		var v defect.Defect
		v, err = s.defects.ChangeStatus(id, "EN_CORRECCION", 0)
		got = baseEntity{v.ID, v.PublicID}
	case "un esfuerzo":
		var v effort.Effort
		v, err = s.efforts.CorrectHours(id, 3)
		got = baseEntity{v.ID, v.PublicID}
	case "una sesión":
		var v poker.Session
		v, err = s.poker.Vote(id, "Ana Pérez", 5)
		got = baseEntity{v.ID, v.PublicID}
	default:
		return fmt.Errorf("entidad desconocida %q", kind)
	}
	if err != nil {
		return fmt.Errorf("modificar %s: %w", kind, err)
	}
	s.modified = got
	return nil
}

func (s *uuidBaseScenario) modifiedKeepsPublicID() error {
	want := s.last()
	if s.modified.publicID != want.publicID || s.modified.publicID == uuid.Nil() {
		return fmt.Errorf("identificador público después de modificar = %s, se esperaba %s", s.modified.publicID, want.publicID)
	}
	return nil
}

func (s *uuidBaseScenario) findFailsWith(msg string) error {
	if s.err == nil {
		return fmt.Errorf("la búsqueda no falló: encontró %+v", s.found)
	}
	if s.err.Error() != msg {
		return fmt.Errorf("error = %q, se esperaba %q", s.err, msg)
	}
	if s.found != (baseEntity{}) {
		return fmt.Errorf("se encontró %+v, se esperaba nada", s.found)
	}
	return nil
}
