package httpapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// HU-52 · Identificadores UUID del resto de las entidades, parte de defectos y
// esfuerzos (Issue #197): runners de godog y pasos compartidos por
// features/HU-52-defectos.feature y features/HU-52-esfuerzos.feature.

// HU-52 · BDD · Verifica que se cumplan los escenarios de features/HU-52-defectos.feature.
func TestFeatures_HU52Defectos(t *testing.T) {
	runDefectEffortFeature(t, "../../features/HU-52-defectos.feature", "defectos")
}

// HU-52 · BDD · Verifica que se cumplan los escenarios de features/HU-52-esfuerzos.feature.
func TestFeatures_HU52Esfuerzos(t *testing.T) {
	runDefectEffortFeature(t, "../../features/HU-52-esfuerzos.feature", "esfuerzos")
}

func runDefectEffortFeature(t *testing.T, path, name string) {
	t.Helper()
	suite := godog.TestSuite{
		ScenarioInitializer: initializeDefectEffortScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{path},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatalf("fallaron escenarios BDD de HU-52 (%s)", name)
	}
}

// defectEffortScenario guarda el estado de un escenario.
type defectEffortScenario struct {
	w *defectEffortWorld
	// ids guarda los UUID que reemplazan a {nombre} en las rutas y cuerpos.
	ids map[string]string
	rec *httptest.ResponseRecorder
}

func initializeDefectEffortScenario(ctx *godog.ScenarioContext) {
	s := &defectEffortScenario{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.w, s.ids, s.rec = nil, nil, nil
		return ctx, nil
	})

	ctx.Step(`^que existe un ítem asignado a un sprint abierto$`, s.itemInOpenSprint)
	ctx.Step(`^que no hay (defectos|esfuerzos) registrados$`, s.noneRegistered)
	ctx.Step(`^que el ítem tiene un defecto registrado$`, s.itemHasDefect)
	ctx.Step(`^que el defecto está en EN_CORRECCION$`, s.defectInCorrection)
	ctx.Step(`^que el ítem tiene (\d+) esfuerzos registrados$`, s.itemHasEfforts)
	ctx.Step(`^pido (GET|POST|PUT|PATCH) "([^"]*)"$`, s.request)
	ctx.Step(`^pido (GET|POST|PUT|PATCH) "([^"]*)" con el cuerpo:$`, s.requestWithBody)
	ctx.Step(`^la respuesta tiene el código (\d+)$`, s.statusIs)
	ctx.Step(`^el error es "([^"]*)"$`, s.errorIs)
	ctx.Step(`^el "([^"]*)" de la respuesta es "([^"]*)"$`, s.fieldIs)
	ctx.Step(`^el "([^"]*)" de la respuesta es el número ([0-9.]+)$`, s.fieldIsNumber)
	ctx.Step(`^el "([^"]*)" de la respuesta es un texto vacío$`, s.fieldIsEmptyText)
	ctx.Step(`^el "([^"]*)" de la respuesta es un UUID versión 4 en minúsculas, como texto$`, s.fieldIsUUIDv4)
	ctx.Step(`^el "([^"]*)" de la respuesta no es un número ni el UUID nulo$`, s.fieldIsNotNumberNorNil)
	ctx.Step(`^el "([^"]*)" de la respuesta es el UUID del (ítem|sprint|defecto|esfuerzo)$`, s.fieldIsUUIDOf)
	ctx.Step(`^la lista "([^"]*)" tiene (\d+) elementos$`, s.listHasLength)
	ctx.Step(`^cada elemento de "([^"]*)" tiene en "([^"]*)" un UUID versión 4$`, s.eachIsUUIDv4)
	ctx.Step(`^cada elemento de "([^"]*)" tiene en "([^"]*)" el UUID del (ítem|sprint|defecto|esfuerzo)$`, s.eachIsUUIDOf)
	ctx.Step(`^cada elemento de "([^"]*)" tiene en "([^"]*)" un texto vacío$`, s.eachIsEmptyText)
}

func (s *defectEffortScenario) itemInOpenSprint() error {
	w, err := buildDefectEffortWorld()
	if err != nil {
		return err
	}
	s.w = w
	s.ids = map[string]string{
		"proyecto":    w.projectUUID(),
		"ítem":        w.itemUUID(),
		"sprint":      w.sprintUUID(),
		"inexistente": neverAssignedUUID,
	}
	return nil
}

func (s *defectEffortScenario) noneRegistered(kind string) error {
	key := map[string]string{"defectos": "defecto", "esfuerzos": "esfuerzo"}[kind]
	if _, ok := s.ids[key]; ok {
		return fmt.Errorf("ya hay %s registrados en el escenario", kind)
	}
	return nil
}

func (s *defectEffortScenario) itemHasDefect() error {
	id, err := s.w.registerDefect()
	if err != nil {
		return err
	}
	s.setEntity("defecto", id)
	return nil
}

func (s *defectEffortScenario) defectInCorrection() error {
	return s.w.startCorrection()
}

func (s *defectEffortScenario) itemHasEfforts(n int) error {
	for i := 0; i < n; i++ {
		id, err := s.w.registerEffort(float64(i) + 1)
		if err != nil {
			return err
		}
		s.setEntity("esfuerzo", id)
	}
	return nil
}

// setEntity guarda el UUID de la entidad y sus formas escritas de otra
// manera, para {defecto en mayúsculas}, {esfuerzo entre llaves}, etc.
func (s *defectEffortScenario) setEntity(name, id string) {
	s.ids[name] = id
	s.ids[name+" en mayúsculas"] = strings.ToUpper(id)
	s.ids[name+" entre llaves"] = "{" + id + "}"
	s.ids[name+" con urn:uuid: adelante"] = "urn:uuid:" + id
	s.ids[name+" sin guiones"] = strings.ReplaceAll(id, "-", "")
}

func (s *defectEffortScenario) request(method, path string) error {
	return s.requestBody(method, path, "")
}

func (s *defectEffortScenario) requestWithBody(method, path string, body *godog.DocString) error {
	return s.requestBody(method, path, body.Content)
}

func (s *defectEffortScenario) requestBody(method, path, body string) error {
	path, err := s.expand(path)
	if err != nil {
		return err
	}
	if body, err = s.expand(body); err != nil {
		return err
	}
	s.rec = s.w.serve(method, path, body)
	return nil
}

// placeholder reconoce {nombre} sin comillas ni llaves adentro, para no
// confundirlo con un objeto JSON del cuerpo.
var placeholder = regexp.MustCompile(`\{([^{}"]+)\}`)

// expand reemplaza cada {nombre} por el UUID guardado con ese nombre.
func (s *defectEffortScenario) expand(text string) (string, error) {
	var missing string
	out := placeholder.ReplaceAllStringFunc(text, func(m string) string {
		id, ok := s.ids[m[1:len(m)-1]]
		if !ok {
			missing = m
		}
		return id
	})
	if missing != "" {
		return "", fmt.Errorf("el escenario no tiene un UUID para %s", missing)
	}
	return out, nil
}

func (s *defectEffortScenario) statusIs(want int) error {
	if s.rec.Code != want {
		return fmt.Errorf("código = %d, se esperaba %d; cuerpo: %s", s.rec.Code, want, s.rec.Body.String())
	}
	return nil
}

func (s *defectEffortScenario) errorIs(want string) error {
	var body errorJSON
	if err := json.Unmarshal(s.rec.Body.Bytes(), &body); err != nil {
		return fmt.Errorf("cuerpo no es JSON válido (%v): %s", err, s.rec.Body.String())
	}
	if body.Error != want {
		return fmt.Errorf("error = %q, se esperaba %q", body.Error, want)
	}
	return nil
}

func (s *defectEffortScenario) fieldIs(field, want string) error {
	return s.expectText(field, want)
}

func (s *defectEffortScenario) fieldIsEmptyText(field string) error {
	return s.expectText(field, "")
}

func (s *defectEffortScenario) fieldIsNumber(field string, want float64) error {
	raw, err := jsonField(s.rec, field)
	if err != nil {
		return err
	}
	var got float64
	if err := json.Unmarshal(raw, &got); err != nil || got != want {
		return fmt.Errorf("%s = %s, se esperaba el número %v", field, raw, want)
	}
	return nil
}

func (s *defectEffortScenario) fieldIsUUIDv4(field string) error {
	got, err := s.text(field)
	if err != nil {
		return err
	}
	return checkUUIDv4(field, got)
}

func (s *defectEffortScenario) fieldIsNotNumberNorNil(field string) error {
	got, err := s.text(field)
	if err != nil {
		return err
	}
	if got == nilUUID || numericText.MatchString(got) {
		return fmt.Errorf("%s = %q, no debería ser un número ni el UUID nulo", field, got)
	}
	return nil
}

func (s *defectEffortScenario) fieldIsUUIDOf(field, name string) error {
	return s.expectText(field, s.ids[name])
}

func (s *defectEffortScenario) listHasLength(key string, want int) error {
	elements, err := s.list(key)
	if err != nil {
		return err
	}
	if len(elements) != want {
		return fmt.Errorf("%s tiene %d elementos, se esperaban %d", key, len(elements), want)
	}
	return nil
}

func (s *defectEffortScenario) eachIsUUIDv4(key, field string) error {
	return s.each(key, field, func(got string) error { return checkUUIDv4(field, got) })
}

func (s *defectEffortScenario) eachIsUUIDOf(key, field, name string) error {
	return s.each(key, field, func(got string) error { return checkEqual(field, got, s.ids[name]) })
}

func (s *defectEffortScenario) eachIsEmptyText(key, field string) error {
	return s.each(key, field, func(got string) error { return checkEqual(field, got, "") })
}

// each aplica check al campo field, como texto, de cada elemento de la lista.
func (s *defectEffortScenario) each(key, field string, check func(string) error) error {
	elements, err := s.list(key)
	if err != nil {
		return err
	}
	if len(elements) == 0 {
		return fmt.Errorf("%s está vacía", key)
	}
	for i, e := range elements {
		got, err := asText(fmt.Sprintf("%s[%d].%s", key, i, field), e[field])
		if err != nil {
			return err
		}
		if err := check(got); err != nil {
			return fmt.Errorf("%s[%d]: %w", key, i, err)
		}
	}
	return nil
}

func (s *defectEffortScenario) list(key string) ([]map[string]json.RawMessage, error) {
	raw, err := jsonField(s.rec, key)
	if err != nil {
		return nil, err
	}
	var elements []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &elements); err != nil {
		return nil, fmt.Errorf("%s = %s, se esperaba una lista de objetos", key, raw)
	}
	return elements, nil
}

func (s *defectEffortScenario) expectText(field, want string) error {
	got, err := s.text(field)
	if err != nil {
		return err
	}
	return checkEqual(field, got, want)
}

// text devuelve el campo de la respuesta, que tiene que ser texto: un número
// es un error, porque HU-52 no expone el ID entero (CA-52.3).
func (s *defectEffortScenario) text(field string) (string, error) {
	raw, err := jsonField(s.rec, field)
	if err != nil {
		return "", err
	}
	return asText(field, raw)
}

func asText(what string, raw json.RawMessage) (string, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return "", fmt.Errorf("%s = %s, se esperaba un texto", what, raw)
	}
	return text, nil
}

func checkEqual(field, got, want string) error {
	if got != want {
		return fmt.Errorf("%s = %q, se esperaba %q", field, got, want)
	}
	return nil
}

func checkUUIDv4(field, got string) error {
	if !uuidV4Canonical.MatchString(got) {
		return fmt.Errorf("%s = %q, se esperaba un UUID versión 4 en forma canónica y en minúsculas", field, got)
	}
	return nil
}
