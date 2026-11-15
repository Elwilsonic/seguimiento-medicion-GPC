package httpapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// HU-52 · Identificadores UUID del resto de las entidades, parte de ítems e
// integrantes (Issue #197). Pasos de features/HU-52-items.feature y
// features/HU-52-integrantes.feature, a nivel de la API.

// HU-52 · BDD · Verifica que se cumplan los escenarios de features/HU-52-items.feature.
func TestFeatures_HU52Items(t *testing.T) {
	runHU52Feature(t, "../../features/HU-52-items.feature", "ítems")
}

// HU-52 · BDD · Verifica que se cumplan los escenarios de features/HU-52-integrantes.feature.
func TestFeatures_HU52Integrantes(t *testing.T) {
	runHU52Feature(t, "../../features/HU-52-integrantes.feature", "integrantes")
}

func runHU52Feature(t *testing.T, path, what string) {
	t.Helper()
	suite := godog.TestSuite{
		ScenarioInitializer: initializeHU52EntityScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{path},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatalf("fallaron escenarios BDD de HU-52 (%s)", what)
	}
}

// hu52EntityScenario guarda el estado de un escenario de ítems o integrantes.
type hu52EntityScenario struct {
	h http.Handler
	// ids guarda el "id" que devolvió la API al crear cada proyecto, ítem o
	// integrante, por nombre, como texto y sin comillas.
	ids map[string]string
	// created cuenta los ítems y los integrantes creados en el escenario.
	created map[string]int
	rec     *httptest.ResponseRecorder
}

func initializeHU52EntityScenario(ctx *godog.ScenarioContext) {
	s := &hu52EntityScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.h = newHandler()
		s.ids = map[string]string{}
		s.created = map[string]int{}
		s.rec = nil
		return ctx, nil
	})

	ctx.Step(`^que existe el proyecto "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el proyecto "([^"]*)" no tiene (ítems|integrantes)$`, s.projectHasNone)
	ctx.Step(`^que el proyecto "([^"]*)" tiene el ítem "([^"]*)"$`, s.projectWithItem)
	ctx.Step(`^que el proyecto "([^"]*)" tiene el integrante "([^"]*)"$`, s.projectWithMember)
	ctx.Step(`^pido (GET|POST|PUT|PATCH) "([^"]*)"$`, s.request)
	ctx.Step(`^pido (GET|POST|PUT|PATCH) "([^"]*)" con el cuerpo:$`, s.requestWithBody)
	ctx.Step(`^pido GET a la ruta del ítem "([^"]*)" escrito (en mayúsculas|entre llaves|con urn:uuid: adelante|sin guiones)$`, s.getItemWrittenAs)
	ctx.Step(`^pido PUT a la ruta del integrante "([^"]*)" escrito (en mayúsculas|entre llaves|con urn:uuid: adelante|sin guiones) con el cuerpo:$`, s.putMemberWrittenAs)
	ctx.Step(`^la respuesta tiene el código (\d+)$`, s.statusIs)
	ctx.Step(`^el error es "([^"]*)"$`, s.errorIs)
	ctx.Step(`^el "([^"]*)" de la respuesta es "([^"]*)"$`, s.fieldIs)
	ctx.Step(`^el "([^"]*)" de la respuesta es un UUID versión 4 en minúsculas, como texto$`, s.fieldIsUUIDv4)
	ctx.Step(`^el "([^"]*)" de la respuesta no es un número ni el UUID nulo$`, s.fieldIsNotNumberNorNil)
	ctx.Step(`^el "([^"]*)" de la respuesta es el UUID de "([^"]*)"$`, s.fieldIsUUIDOf)
	ctx.Step(`^cada (ítem|integrante) de la lista tiene su UUID en "([^"]*)", distinto del de los demás$`, s.listHasDistinctUUIDs)
	ctx.Step(`^cada ítem de la lista tiene "" en "([^"]*)"$`, s.listHasEmptyText)
}

// existingProject crea el proyecto por la API y guarda su "id".
func (s *hu52EntityScenario) existingProject(name string) error {
	return s.create(name, "", "/api/proyectos", `{"nombre":"`+name+`","descripcion":""}`)
}

// projectHasNone verifica que el proyecto todavía no tenga ítems o
// integrantes, para los casos de "el primero".
func (s *hu52EntityScenario) projectHasNone(name, what string) error {
	tail := "/backlog"
	if what == "integrantes" {
		tail = "/integrantes"
	}
	path, err := s.expand("/api/proyectos/{" + name + "}" + tail)
	if err != nil {
		return err
	}
	rec := s.serve(http.MethodGet, path, "")
	var list []json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || rec.Code != http.StatusOK {
		return fmt.Errorf("GET %s: código %d (%v); cuerpo: %s", path, rec.Code, err, rec.Body.String())
	}
	if len(list) != 0 {
		return fmt.Errorf("el proyecto %q tiene %d %s, se esperaba ninguno", name, len(list), what)
	}
	return nil
}

func (s *hu52EntityScenario) projectWithItem(projectName, title string) error {
	path, err := s.expand("/api/proyectos/{" + projectName + "}/backlog")
	if err != nil {
		return err
	}
	return s.create(title, "ítem", path, `{"titulo":"`+title+`","descripcion":"","prioridad":"ALTA","criterios":["Ok"]}`)
}

func (s *hu52EntityScenario) projectWithMember(projectName, memberName string) error {
	path, err := s.expand("/api/proyectos/{" + projectName + "}/integrantes")
	if err != nil {
		return err
	}
	return s.create(memberName, "integrante", path, `{"nombre":"`+memberName+`","rol":"PRODUCT_BUILDER"}`)
}

// create hace el POST y guarda el "id" que devolvió la API con el nombre
// name, sea texto o número: el tipo lo verifican los pasos Entonces.
func (s *hu52EntityScenario) create(name, kind, path, body string) error {
	rec := s.serve(http.MethodPost, path, body)
	if rec.Code != http.StatusCreated {
		return fmt.Errorf("crear %q: código %d; cuerpo: %s", name, rec.Code, rec.Body.String())
	}
	raw, err := jsonField(rec, "id")
	if err != nil {
		return err
	}
	s.ids[name] = strings.Trim(string(raw), `"`)
	if kind != "" {
		s.created[kind]++
	}
	return nil
}

func (s *hu52EntityScenario) request(method, path string) error {
	return s.requestBody(method, path, "")
}

func (s *hu52EntityScenario) requestWithBody(method, path string, body *godog.DocString) error {
	return s.requestBody(method, path, body.Content)
}

func (s *hu52EntityScenario) requestBody(method, path, body string) error {
	path, err := s.expand(path)
	if err != nil {
		return err
	}
	s.rec = s.serve(method, path, body)
	return nil
}

func (s *hu52EntityScenario) getItemWrittenAs(name, form string) error {
	id, err := s.writtenAs(name, form)
	if err != nil {
		return err
	}
	s.rec = s.serve(http.MethodGet, "/api/items/"+id, "")
	return nil
}

func (s *hu52EntityScenario) putMemberWrittenAs(name, form string, body *godog.DocString) error {
	id, err := s.writtenAs(name, form)
	if err != nil {
		return err
	}
	s.rec = s.serve(http.MethodPut, "/api/integrantes/"+id, body.Content)
	return nil
}

// writtenAs devuelve el "id" de name escrito de otra forma.
func (s *hu52EntityScenario) writtenAs(name, form string) (string, error) {
	id, ok := s.ids[name]
	if !ok {
		return "", fmt.Errorf("no existe %q en el escenario", name)
	}
	switch form {
	case "en mayúsculas":
		id = strings.ToUpper(id)
	case "entre llaves":
		id = "{" + id + "}"
	case "con urn:uuid: adelante":
		id = "urn:uuid:" + id
	case "sin guiones":
		id = strings.ReplaceAll(id, "-", "")
	}
	return id, nil
}

func (s *hu52EntityScenario) statusIs(want int) error {
	if s.rec.Code != want {
		return fmt.Errorf("código = %d, se esperaba %d; cuerpo: %s", s.rec.Code, want, s.rec.Body.String())
	}
	return nil
}

func (s *hu52EntityScenario) errorIs(want string) error {
	var body errorJSON
	if err := json.Unmarshal(s.rec.Body.Bytes(), &body); err != nil {
		return fmt.Errorf("cuerpo no es JSON válido (%v): %s", err, s.rec.Body.String())
	}
	if body.Error != want {
		return fmt.Errorf("error = %q, se esperaba %q", body.Error, want)
	}
	return nil
}

func (s *hu52EntityScenario) fieldIs(field, want string) error {
	got, err := hu52StringField(s.rec, field)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("%s = %q, se esperaba %q", field, got, want)
	}
	return nil
}

func (s *hu52EntityScenario) fieldIsUUIDv4(field string) error {
	got, err := hu52StringField(s.rec, field)
	if err != nil {
		return err
	}
	if !uuidV4Canonical.MatchString(got) {
		return fmt.Errorf("%s = %q, se esperaba un UUID versión 4 en forma canónica y en minúsculas", field, got)
	}
	return nil
}

func (s *hu52EntityScenario) fieldIsNotNumberNorNil(field string) error {
	got, err := hu52StringField(s.rec, field)
	if err != nil {
		return err
	}
	if got == nilProjectID {
		return fmt.Errorf("%s es el UUID nulo", field)
	}
	if regexp.MustCompile(`^-?[0-9]+$`).MatchString(got) {
		return fmt.Errorf("%s = %q es un número", field, got)
	}
	return nil
}

func (s *hu52EntityScenario) fieldIsUUIDOf(field, name string) error {
	got, err := hu52StringField(s.rec, field)
	if err != nil {
		return err
	}
	if want := s.ids[name]; got != want {
		return fmt.Errorf("%s = %q, se esperaba el UUID de %q (%q)", field, got, name, want)
	}
	return nil
}

// listHasDistinctUUIDs verifica que la lista tenga una entrada por cada
// ítem o integrante creado, con un UUID versión 4 distinto en field.
func (s *hu52EntityScenario) listHasDistinctUUIDs(kind, field string) error {
	list, err := s.list()
	if err != nil {
		return err
	}
	if len(list) != s.created[kind] {
		return fmt.Errorf("la lista tiene %d elementos, se esperaban %d: %s", len(list), s.created[kind], s.rec.Body.String())
	}
	seen := map[string]bool{}
	for i, obj := range list {
		var text string
		if err := json.Unmarshal(obj[field], &text); err != nil {
			return fmt.Errorf("elemento %d: %s = %s, se esperaba un texto", i, field, obj[field])
		}
		if !uuidV4Canonical.MatchString(text) {
			return fmt.Errorf("elemento %d: %s = %q, se esperaba un UUID versión 4", i, field, text)
		}
		if seen[text] {
			return fmt.Errorf("elemento %d: %s = %q se repite", i, field, text)
		}
		seen[text] = true
	}
	return nil
}

// listHasEmptyText verifica que cada elemento de la lista tenga "" en field
// (un ítem sin sprint asignado, D7 de design.md).
func (s *hu52EntityScenario) listHasEmptyText(field string) error {
	list, err := s.list()
	if err != nil {
		return err
	}
	for i, obj := range list {
		if got := string(obj[field]); got != `""` {
			return fmt.Errorf("elemento %d: %s = %s, se esperaba \"\"", i, field, got)
		}
	}
	return nil
}

func (s *hu52EntityScenario) list() ([]map[string]json.RawMessage, error) {
	var list []map[string]json.RawMessage
	if err := json.Unmarshal(s.rec.Body.Bytes(), &list); err != nil {
		return nil, fmt.Errorf("cuerpo no es una lista JSON (%v): %s", err, s.rec.Body.String())
	}
	return list, nil
}

// expand reemplaza {Nombre} en la ruta por el "id" del proyecto, ítem o
// integrante Nombre.
func (s *hu52EntityScenario) expand(path string) (string, error) {
	var missing string
	out := regexp.MustCompile(`\{([^}]+)\}`).ReplaceAllStringFunc(path, func(m string) string {
		name := m[1 : len(m)-1]
		id, ok := s.ids[name]
		if !ok {
			missing = name
		}
		return id
	})
	if missing != "" {
		return "", fmt.Errorf("no existe %q en el escenario", missing)
	}
	return out, nil
}

func (s *hu52EntityScenario) serve(method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	s.h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

// hu52StringField devuelve el campo de la respuesta, que tiene que ser
// texto: un número es un error, porque HU-52 no expone el ID entero
// (CA-52.3).
func hu52StringField(rec *httptest.ResponseRecorder, field string) (string, error) {
	raw, err := jsonField(rec, field)
	if err != nil {
		return "", err
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return "", fmt.Errorf("%s = %s, se esperaba un texto", field, raw)
	}
	return text, nil
}
