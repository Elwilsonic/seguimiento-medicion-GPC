package httpapi_test

import (
	"net/http"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// HU-54 · Elegir el integrante de una lista al votar y al registrar esfuerzo (Issue #219).

// newHU54Test prepara la interfaz con el handler que valida integrantes,
// como cmd/server.
func newHU54Test(t *testing.T) *hu53 {
	t.Helper()
	s := newHU53Test(t)
	s.h = newHandlerValidatingMembers()
	s.data.h = s.h
	return s
}

const (
	voteMemberField   = "#form-votar-poker [name=integrante]"
	effortMemberField = "#form-registrar-esfuerzo [name=integrante]"
)

// HU-54 · CA-54.1 · Verifica que «Integrante» de "Votar" y de "Registrar esfuerzo" sea un desplegable y no un campo de texto.
func TestHU54_IntegranteEsUnDesplegable(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()
	for _, form := range []string{"form-votar-poker", "form-registrar-esfuerzo"} {
		m := regexp.MustCompile(`(?s)<form id="` + form + `"[^>]*>(.*?)</form>`).FindStringSubmatch(page)
		if m == nil {
			t.Fatalf("la página no tiene el formulario %s", form)
		}
		if !regexp.MustCompile(`<select name="integrante"`).MatchString(m[1]) {
			t.Errorf("%s: «Integrante» no es un <select name=\"integrante\">", form)
		}
		if regexp.MustCompile(`<input name="integrante"`).MatchString(m[1]) {
			t.Errorf("%s: «Integrante» sigue siendo un campo de texto", form)
		}
	}
}

// HU-54 · CA-54.1 · Verifica que los desplegables ofrezcan los integrantes del proyecto actual y cambien con él, en el navegador.
func TestHU54_DesplegablesConLosIntegrantesDelProyecto(t *testing.T) {
	s := newHU54Test(t)
	s.member("Sistema de ventas", "Sosa Ricardo", "PRODUCT_BUILDER")
	s.member("Sistema de ventas", "García Santiago", "PRODUCT_BUILDER")
	s.member("Inventario", "Cardozo Leandro", "PRODUCT_BUILDER")
	s.openWithProjectIn("Sistema de ventas", "#esfuerzo")

	want := []string{"Sosa Ricardo", "García Santiago"}
	for _, sel := range []string{voteMemberField, effortMemberField} {
		s.u.waitTrue(`(s => s && s.options.length === 2)(document.querySelector("`+sel+`"))`, sel+" no tiene los 2 integrantes del proyecto")
		if got := s.choices(sel); !slices.Equal(got, want) {
			t.Errorf("%s ofrece %q, se esperaba %q", sel, got, want)
		}
	}

	s.chooseIn("Inventario", "«Proyecto actual»")
	for _, sel := range []string{voteMemberField, effortMemberField} {
		s.u.waitTrue(`(s => s && s.options.length === 1 && s.options[0].textContent.trim() === "Cardozo Leandro")(document.querySelector("`+sel+`"))`,
			sel+" no cambió a los integrantes de Inventario")
	}
}

// HU-54 · CA-54.1 · Verifica que, sin integrantes, «Integrante» avise "Registrá integrantes primero" y el formulario no se pueda enviar.
func TestHU54_SinIntegrantesAvisa(t *testing.T) {
	s := newHU54Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	for _, sel := range []string{voteMemberField, effortMemberField} {
		s.u.waitTrue(`(s => s && s.options.length === 1 && s.options[0].textContent.trim() === "Registrá integrantes primero" && s.options[0].value === "")(document.querySelector("`+sel+`"))`,
			sel+" no avisa \"Registrá integrantes primero\"")
	}
	s.goTo("Esfuerzo")
	s.u.choose(fields["«Ítem» de Registrar esfuerzo"], "Registrar venta")
	s.u.countAPICalls()
	s.u.click("#form-registrar-esfuerzo button[type=submit]")
	s.messageIs("esfuerzo", "Registrá integrantes primero")
	if n := s.u.apiCalls(); n != 0 {
		t.Errorf("el formulario llamó %d vez/veces a la API sin integrante elegido", n)
	}
}

// HU-54 · CA-54.1 · Verifica que se registre un esfuerzo eligiendo al integrante de la lista, en el navegador.
func TestHU54_RegistrarEsfuerzoEligiendoDeLaLista(t *testing.T) {
	s := newHU54Test(t)
	s.openSprintWithItem("Sistema de ventas", "Sprint 2", "Registrar venta")
	s.member("Sistema de ventas", "Sosa Ricardo", "PRODUCT_BUILDER")
	s.member("Sistema de ventas", "García Santiago", "PRODUCT_BUILDER")
	s.openWithProjectAndSprint("Sistema de ventas", "Sprint 2")

	s.goTo("Esfuerzo")
	s.u.choose(fields["«Ítem» de Registrar esfuerzo"], "Registrar venta")
	s.u.choose(effortMemberField, "García Santiago")
	s.u.fill("#form-registrar-esfuerzo [name=fecha]", "2026-10-02")
	s.u.fill("#form-registrar-esfuerzo [name=actividad]", "Desarrollo")
	s.u.fill("#form-registrar-esfuerzo [name=horas]", "3")
	s.u.click("#form-registrar-esfuerzo button[type=submit]")

	s.u.waitTrue(`document.querySelector("#msg-esfuerzos") && document.querySelector("#msg-esfuerzos").textContent.includes("Esfuerzo registrado")`,
		"no se registró el esfuerzo")
	body := serveRaw(s.h, http.MethodGet, "/api/sprints/"+s.ids["Sprint 2"]+"/esfuerzos", "").Body.String()
	if !strings.Contains(body, `"integrante":"García Santiago"`) {
		t.Errorf("el esfuerzo no quedó a nombre de \"García Santiago\": %s", body)
	}
}
