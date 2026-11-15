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

// HU-51 · Identificadores UUID, parte de proyectos (Issue #193)

// uuidV4Canonical es la forma canónica de un UUID versión 4, variante RFC
// 9562, en minúsculas (CA-51.1).
var uuidV4Canonical = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// HU-51 · BDD · Verifica que se cumplan los escenarios de features/HU-51-identificadores-uuid.feature.
func TestFeatures_HU51IdentificadoresUUID(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeUUIDScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-51-identificadores-uuid.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-51")
	}
}

// uuidScenario guarda el estado de un escenario de HU-51.
type uuidScenario struct {
	h http.Handler
	// ids guarda el "id" que devolvió la API al crear cada proyecto, por
	// nombre, como texto.
	ids map[string]string
	rec *httptest.ResponseRecorder
}

func initializeUUIDScenario(ctx *godog.ScenarioContext) {
	s := &uuidScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.h = newHandler()
		s.ids = map[string]string{}
		s.rec = nil
		return ctx, nil
	})

	ctx.Step(`^que no hay proyectos creados$`, s.noProjects)
	ctx.Step(`^que existe el proyecto "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el proyecto "([^"]*)" tiene un ítem en su backlog$`, s.projectWithItem)
	ctx.Step(`^pido (GET|POST|PUT|PATCH) "([^"]*)"$`, s.request)
	ctx.Step(`^pido (GET|POST|PUT|PATCH) "([^"]*)" con el cuerpo:$`, s.requestWithBody)
	ctx.Step(`^pido GET a la ruta del proyecto "([^"]*)" escrito (en mayúsculas|entre llaves|con urn:uuid: adelante|sin guiones)$`, s.requestWrittenAs)
	ctx.Step(`^la respuesta tiene el código (\d+)$`, s.statusIs)
	ctx.Step(`^el error es "([^"]*)"$`, s.errorIs)
	ctx.Step(`^el "([^"]*)" de la respuesta es "([^"]*)"$`, s.fieldIs)
	ctx.Step(`^el "([^"]*)" de la respuesta es un UUID versión 4 en minúsculas, como texto$`, s.fieldIsUUIDv4)
	ctx.Step(`^el "([^"]*)" de la respuesta no es un número ni el UUID nulo$`, s.fieldIsNotNumberNorNil)
	ctx.Step(`^el "([^"]*)" de la respuesta es el UUID de "([^"]*)"$`, s.fieldIsUUIDOf)
	ctx.Step(`^estas rutas del proyecto "([^"]*)" responden 200:$`, s.routesRespond200)
}

func (s *uuidScenario) noProjects() error {
	if len(s.ids) != 0 {
		return fmt.Errorf("hay %d proyectos creados, se esperaba 0", len(s.ids))
	}
	return nil
}

// existingProject crea el proyecto por la API y guarda el "id" que devolvió,
// sea texto o número: la verificación del tipo la hacen los pasos Entonces.
func (s *uuidScenario) existingProject(name string) error {
	rec := s.serve(http.MethodPost, "/api/proyectos", `{"nombre":"`+name+`","descripcion":""}`)
	if rec.Code != http.StatusCreated {
		return fmt.Errorf("crear %q: código %d; cuerpo: %s", name, rec.Code, rec.Body.String())
	}
	raw, err := jsonField(rec, "id")
	if err != nil {
		return err
	}
	s.ids[name] = strings.Trim(string(raw), `"`)
	return nil
}

func (s *uuidScenario) projectWithItem(name string) error {
	path, err := s.expand("/api/proyectos/{" + name + "}/backlog")
	if err != nil {
		return err
	}
	if rec := s.serve(http.MethodPost, path, validItem); rec.Code != http.StatusCreated {
		return fmt.Errorf("crear el ítem: código %d; cuerpo: %s", rec.Code, rec.Body.String())
	}
	return nil
}

func (s *uuidScenario) request(method, path string) error {
	return s.requestBody(method, path, "")
}

func (s *uuidScenario) requestWithBody(method, path string, body *godog.DocString) error {
	return s.requestBody(method, path, body.Content)
}

func (s *uuidScenario) requestBody(method, path, body string) error {
	path, err := s.expand(path)
	if err != nil {
		return err
	}
	s.rec = s.serve(method, path, body)
	return nil
}

// requestWrittenAs pide el proyecto con su UUID escrito de otra forma.
func (s *uuidScenario) requestWrittenAs(name, form string) error {
	id, ok := s.ids[name]
	if !ok {
		return fmt.Errorf("no existe el proyecto %q en el escenario", name)
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
	s.rec = s.serve(http.MethodGet, "/api/proyectos/"+id, "")
	return nil
}

func (s *uuidScenario) statusIs(want int) error {
	if s.rec.Code != want {
		return fmt.Errorf("código = %d, se esperaba %d; cuerpo: %s", s.rec.Code, want, s.rec.Body.String())
	}
	return nil
}

func (s *uuidScenario) errorIs(want string) error {
	var body errorJSON
	if err := json.Unmarshal(s.rec.Body.Bytes(), &body); err != nil {
		return fmt.Errorf("cuerpo no es JSON válido (%v): %s", err, s.rec.Body.String())
	}
	if body.Error != want {
		return fmt.Errorf("error = %q, se esperaba %q", body.Error, want)
	}
	return nil
}

func (s *uuidScenario) fieldIs(field, want string) error {
	got, err := s.stringField(field)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("%s = %q, se esperaba %q", field, got, want)
	}
	return nil
}

func (s *uuidScenario) fieldIsUUIDv4(field string) error {
	got, err := s.stringField(field)
	if err != nil {
		return err
	}
	if !uuidV4Canonical.MatchString(got) {
		return fmt.Errorf("%s = %q, se esperaba un UUID versión 4 en forma canónica y en minúsculas", field, got)
	}
	return nil
}

func (s *uuidScenario) fieldIsNotNumberNorNil(field string) error {
	got, err := s.stringField(field)
	if err != nil {
		return err
	}
	if got == "00000000-0000-0000-0000-000000000000" {
		return fmt.Errorf("%s es el UUID nulo", field)
	}
	if regexp.MustCompile(`^[0-9]+$`).MatchString(got) {
		return fmt.Errorf("%s = %q es un número", field, got)
	}
	return nil
}

func (s *uuidScenario) fieldIsUUIDOf(field, name string) error {
	got, err := s.stringField(field)
	if err != nil {
		return err
	}
	if want := s.ids[name]; got != want {
		return fmt.Errorf("%s = %q, se esperaba el UUID de %q (%q)", field, got, name, want)
	}
	return nil
}

func (s *uuidScenario) routesRespond200(name string, table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		path, err := s.expand("/api/proyectos/{" + name + "}" + row.Cells[0].Value)
		if err != nil {
			return err
		}
		if rec := s.serve(http.MethodGet, path, ""); rec.Code != http.StatusOK {
			return fmt.Errorf("GET %s: código %d, se esperaba 200; cuerpo: %s", path, rec.Code, rec.Body.String())
		}
	}
	return nil
}

// stringField devuelve el campo de la respuesta, que tiene que ser texto: un
// número es un error, porque HU-51 no expone el ID entero (CA-51.3).
func (s *uuidScenario) stringField(field string) (string, error) {
	raw, err := jsonField(s.rec, field)
	if err != nil {
		return "", err
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return "", fmt.Errorf("%s = %s, se esperaba un texto", field, raw)
	}
	return text, nil
}

// expand reemplaza {Nombre} en la ruta por el "id" del proyecto Nombre.
func (s *uuidScenario) expand(path string) (string, error) {
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
		return "", fmt.Errorf("no existe el proyecto %q en el escenario", missing)
	}
	return out, nil
}

func (s *uuidScenario) serve(method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	s.h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

// jsonField devuelve el valor crudo de un campo del objeto JSON de la respuesta.
func jsonField(rec *httptest.ResponseRecorder, field string) (json.RawMessage, error) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		return nil, fmt.Errorf("cuerpo no es un objeto JSON (%v): %s", err, rec.Body.String())
	}
	raw, ok := body[field]
	if !ok {
		return nil, fmt.Errorf("la respuesta no tiene el campo %q: %s", field, rec.Body.String())
	}
	return raw, nil
}
