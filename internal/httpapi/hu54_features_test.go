package httpapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// HU-54 · Votar y registrar esfuerzo solo con integrantes del proyecto (Issue #219)

// HU-54 · BDD · Verifica que se cumplan los escenarios de features/HU-54-integrantes-del-proyecto.feature.
func TestFeatures_HU54IntegrantesDelProyecto(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeHU54Scenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-54-integrantes-del-proyecto.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-54")
	}
}

// hu54Scenario guarda el estado de un escenario: los UUID que devolvió la API
// por nombre, la sesión de Planning Poker y la última respuesta.
type hu54Scenario struct {
	h        http.Handler
	projects map[string]string // nombre del proyecto → UUID
	sprints  map[string]string // nombre del proyecto → UUID de su sprint abierto
	items    map[string]string // título del ítem → UUID
	session  string
	rec      *httptest.ResponseRecorder
}

func initializeHU54Scenario(ctx *godog.ScenarioContext) {
	s := &hu54Scenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.h = newHandlerValidatingMembers()
		s.projects = map[string]string{}
		s.sprints = map[string]string{}
		s.items = map[string]string{}
		s.session = ""
		s.rec = nil
		return ctx, nil
	})

	ctx.Step(`^que existe el proyecto "([^"]*)" con los integrantes "([^"]*)" y "([^"]*)"$`, s.projectWithTwoMembers)
	ctx.Step(`^que existe el proyecto "([^"]*)" con el integrante "([^"]*)"$`, s.projectWithMember)
	ctx.Step(`^que existe el proyecto "([^"]*)" sin integrantes$`, s.projectWithoutMembers)
	ctx.Step(`^que el proyecto "([^"]*)" tiene el ítem "([^"]*)" en su sprint abierto$`, s.itemInOpenSprint)
	ctx.Step(`^que el proyecto "([^"]*)" tiene el ítem "([^"]*)" sin sprint$`, s.itemWithoutSprint)
	ctx.Step(`^que hay una sesión de Planning Poker abierta para "([^"]*)"$`, s.openSession)
	ctx.Step(`^que hay una sesión de Planning Poker revelada para "([^"]*)"$`, s.revealedSession)
	ctx.Step(`^"([^"]*)" vota (\d+)$`, s.vote)
	ctx.Step(`^"([^"]*)" registra (\d+) horas de "([^"]*)" en "([^"]*)"$`, s.registerEffort)
	ctx.Step(`^la respuesta tiene el código (\d+)$`, s.statusIs)
	ctx.Step(`^el error es "([^"]*)"$`, s.errorIs)
	ctx.Step(`^al revelar, los votos son:$`, s.revealedVotesAre)
	ctx.Step(`^al revelar después de que "([^"]*)" vota (\d+), los votos son:$`, s.voteThenRevealedVotesAre)
	ctx.Step(`^el esfuerzo registrado es de "([^"]*)"$`, s.effortMemberIs)
	ctx.Step(`^el sprint abierto de "([^"]*)" no tiene esfuerzos$`, s.sprintWithoutEfforts)
}

// call hace el pedido y devuelve la respuesta, sin guardarla como la última.
func (s *hu54Scenario) call(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	s.h.ServeHTTP(rec, req)
	return rec
}

// mustCall hace un pedido de preparación y exige el código esperado.
func (s *hu54Scenario) mustCall(method, path, body string, want int) (map[string]any, error) {
	rec := s.call(method, path, body)
	if rec.Code != want {
		return nil, fmt.Errorf("%s %s respondió %d, se esperaba %d: %s", method, path, rec.Code, want, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("%s %s: respuesta no es un objeto JSON: %v", method, path, err)
	}
	return out, nil
}

func (s *hu54Scenario) createProject(name string, members ...string) error {
	p, err := s.mustCall(http.MethodPost, "/api/proyectos", `{"nombre":"`+name+`","descripcion":"Proyecto de prueba"}`, http.StatusCreated)
	if err != nil {
		return err
	}
	id := p["id"].(string)
	s.projects[name] = id
	for _, m := range members {
		if _, err := s.mustCall(http.MethodPost, "/api/proyectos/"+id+"/integrantes", `{"nombre":"`+m+`","rol":"PRODUCT_BUILDER"}`, http.StatusCreated); err != nil {
			return err
		}
	}
	return nil
}

func (s *hu54Scenario) projectWithTwoMembers(name, a, b string) error {
	return s.createProject(name, a, b)
}

func (s *hu54Scenario) projectWithMember(name, member string) error {
	return s.createProject(name, member)
}

func (s *hu54Scenario) projectWithoutMembers(name string) error {
	return s.createProject(name)
}

// createItem crea un ítem en el backlog del proyecto y guarda su UUID.
func (s *hu54Scenario) createItem(projectName, title string) (string, error) {
	pid, ok := s.projects[projectName]
	if !ok {
		return "", fmt.Errorf("no existe el proyecto %q en el escenario", projectName)
	}
	it, err := s.mustCall(http.MethodPost, "/api/proyectos/"+pid+"/backlog",
		`{"titulo":"`+title+`","descripcion":"Ítem de prueba","prioridad":"ALTA","criterios":["Funciona"]}`, http.StatusCreated)
	if err != nil {
		return "", err
	}
	id := it["id"].(string)
	s.items[title] = id
	return id, nil
}

// openSprint devuelve el sprint abierto del proyecto, y lo crea la primera vez.
func (s *hu54Scenario) openSprint(projectName string) (string, error) {
	if id, ok := s.sprints[projectName]; ok {
		return id, nil
	}
	sp, err := s.mustCall(http.MethodPost, "/api/proyectos/"+s.projects[projectName]+"/sprints",
		`{"nombre":"Sprint 1","fecha_inicio":"2026-11-02","fecha_fin":"2026-11-15","objetivo":"Probar HU-54"}`, http.StatusCreated)
	if err != nil {
		return "", err
	}
	id := sp["id"].(string)
	s.sprints[projectName] = id
	return id, nil
}

func (s *hu54Scenario) itemInOpenSprint(projectName, title string) error {
	item, err := s.createItem(projectName, title)
	if err != nil {
		return err
	}
	sprintID, err := s.openSprint(projectName)
	if err != nil {
		return err
	}
	_, err = s.mustCall(http.MethodPost, "/api/sprints/"+sprintID+"/items/"+item, "", http.StatusOK)
	return err
}

func (s *hu54Scenario) itemWithoutSprint(projectName, title string) error {
	_, err := s.createItem(projectName, title)
	return err
}

func (s *hu54Scenario) openSession(title string) error {
	item, ok := s.items[title]
	if !ok {
		return fmt.Errorf("no existe el ítem %q en el escenario", title)
	}
	sess, err := s.mustCall(http.MethodPost, "/api/items/"+item+"/poker-sessions", "", http.StatusCreated)
	if err != nil {
		return err
	}
	s.session = sess["id"].(string)
	return nil
}

// revealedSession abre una sesión, vota con un integrante del proyecto y la
// revela.
func (s *hu54Scenario) revealedSession(title string) error {
	if err := s.openSession(title); err != nil {
		return err
	}
	if _, err := s.mustCall(http.MethodPost, "/api/poker-sessions/"+s.session+"/votes", `{"integrante":"Sosa Ricardo","puntos":3}`, http.StatusOK); err != nil {
		return err
	}
	_, err := s.mustCall(http.MethodPost, "/api/poker-sessions/"+s.session+"/reveal", "", http.StatusOK)
	return err
}

func (s *hu54Scenario) vote(member string, points int) error {
	body, _ := json.Marshal(map[string]any{"integrante": member, "puntos": points})
	s.rec = s.call(http.MethodPost, "/api/poker-sessions/"+s.session+"/votes", string(body))
	return nil
}

func (s *hu54Scenario) registerEffort(member string, hours int, activity, title string) error {
	item, ok := s.items[title]
	if !ok {
		return fmt.Errorf("no existe el ítem %q en el escenario", title)
	}
	sprintID, err := s.openSprint("Sistema de ventas")
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{"integrante": member, "fecha": "2026-11-05", "actividad": activity, "horas": hours})
	s.rec = s.call(http.MethodPost, "/api/sprints/"+sprintID+"/items/"+item+"/esfuerzos", string(body))
	return nil
}

func (s *hu54Scenario) statusIs(want int) error {
	if s.rec == nil {
		return fmt.Errorf("no se hizo ningún pedido")
	}
	if s.rec.Code != want {
		return fmt.Errorf("código %d, se esperaba %d: %s", s.rec.Code, want, s.rec.Body.String())
	}
	return nil
}

func (s *hu54Scenario) errorIs(want string) error {
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(s.rec.Body.Bytes(), &body); err != nil {
		return fmt.Errorf("la respuesta no es JSON: %v", err)
	}
	if body.Error != want {
		return fmt.Errorf("error %q, se esperaba %q", body.Error, want)
	}
	return nil
}

func (s *hu54Scenario) revealedVotesAre(table *godog.Table) error {
	rec := s.call(http.MethodPost, "/api/poker-sessions/"+s.session+"/reveal", "")
	if rec.Code != http.StatusOK {
		return fmt.Errorf("revelar respondió %d: %s", rec.Code, rec.Body.String())
	}
	var sess struct {
		Votos []struct {
			Integrante string `json:"integrante"`
			Puntos     int    `json:"puntos"`
		} `json:"votos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &sess); err != nil {
		return err
	}
	var got []string
	for _, v := range sess.Votos {
		got = append(got, fmt.Sprintf("%s=%d", v.Integrante, v.Puntos))
	}
	var want []string
	for _, row := range table.Rows[1:] {
		want = append(want, row.Cells[0].Value+"="+row.Cells[1].Value)
	}
	if strings.Join(got, ", ") != strings.Join(want, ", ") {
		return fmt.Errorf("votos revelados [%s], se esperaba [%s]", strings.Join(got, ", "), strings.Join(want, ", "))
	}
	return nil
}

func (s *hu54Scenario) voteThenRevealedVotesAre(member string, points int, table *godog.Table) error {
	body, _ := json.Marshal(map[string]any{"integrante": member, "puntos": points})
	if _, err := s.mustCall(http.MethodPost, "/api/poker-sessions/"+s.session+"/votes", string(body), http.StatusOK); err != nil {
		return err
	}
	return s.revealedVotesAre(table)
}

func (s *hu54Scenario) effortMemberIs(want string) error {
	var e struct {
		Integrante string `json:"integrante"`
	}
	if err := json.Unmarshal(s.rec.Body.Bytes(), &e); err != nil {
		return err
	}
	if e.Integrante != want {
		return fmt.Errorf("el esfuerzo es de %q, se esperaba %q", e.Integrante, want)
	}
	return nil
}

func (s *hu54Scenario) sprintWithoutEfforts(projectName string) error {
	rec := s.call(http.MethodGet, "/api/sprints/"+s.sprints[projectName]+"/esfuerzos", "")
	if rec.Code != http.StatusOK {
		return fmt.Errorf("listar esfuerzos respondió %d: %s", rec.Code, rec.Body.String())
	}
	var list struct {
		Esfuerzos []any `json:"esfuerzos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		return err
	}
	if len(list.Esfuerzos) != 0 {
		return fmt.Errorf("el sprint tiene %d esfuerzo(s), se esperaba ninguno", len(list.Esfuerzos))
	}
	return nil
}
