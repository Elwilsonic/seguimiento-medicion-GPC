package httpapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// HU-53 · Navegar la interfaz por secciones y elegir de listas (Issue #211).
//
// ui maneja la interfaz en el navegador para los tests de HU-53 y los pasos
// de su .feature, y uiData carga los datos por la API antes de abrir la
// página (R1 de design.md: datos listos y esperas por condición, no por
// tiempo).

// uiWait es el tiempo máximo que se espera una condición en la página.
const uiWait = 5 * time.Second

// ui es la interfaz abierta en el navegador.
type ui struct {
	t   failer
	ctx context.Context
	url string
}

// openUI abre la interfaz servida por h en la dirección path (por ejemplo,
// "/" o "/#sprints") y espera a que app.js termine de cargar los proyectos.
func openUI(t *testing.T, h http.Handler, path string) *ui {
	t.Helper()
	ctx, url := newBrowser(t, h)
	u := &ui{t: t, ctx: ctx, url: url}
	u.navigate(path)
	return u
}

// navigate abre path en la misma pestaña y espera a que la página cargue.
func (u *ui) navigate(path string) {
	u.t.Helper()
	browse(u.t, u.ctx, chromedp.Navigate(u.url+path))
	u.waitTrue(`document.readyState === "complete"`, "la página no terminó de cargar")
}

// eval evalúa una expresión de JavaScript en la página y devuelve su valor.
func eval[T any](u *ui, expr string) T {
	u.t.Helper()
	return value(u.t, u.ctx, chromedp.Evaluate[T](expr))
}

// waitTrue espera a que la expresión de JavaScript sea verdadera; si no lo es
// en uiWait, falla con el mensaje.
func (u *ui) waitTrue(expr, msg string) {
	u.t.Helper()
	deadline := time.Now().Add(uiWait)
	for {
		ok, err := chromedp.Run(u.ctx, chromedp.Evaluate[bool]("Boolean("+expr+")"))
		if err == nil && ok {
			return
		}
		if time.Now().After(deadline) {
			u.t.Fatalf("%s (esperé %s a que %s)", msg, uiWait, expr)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// q devuelve el texto JavaScript que busca el primer elemento con el selector
// CSS sel.
func q(sel string) string {
	return "document.querySelector(" + strconv.Quote(sel) + ")"
}

// exists dice si hay un elemento con el selector.
func (u *ui) exists(sel string) bool {
	u.t.Helper()
	return eval[bool](u, q(sel)+" !== null")
}

// click hace clic en el elemento con el selector, que tiene que existir.
func (u *ui) click(sel string) {
	u.t.Helper()
	u.waitTrue(q(sel)+" !== null", "no hay un elemento "+sel)
	eval[bool](u, q(sel)+".click(), true")
}

// clickText hace clic en el botón o link con ese texto dentro de within.
func (u *ui) clickText(within, text string) {
	u.t.Helper()
	expr := fmt.Sprintf(`[...document.querySelectorAll(%s)].find(e => e.textContent.trim() === %s)`,
		strconv.Quote(within+" button, "+within+" a"), strconv.Quote(text))
	u.waitTrue(expr+" !== undefined", fmt.Sprintf("no hay un botón %q en %s", text, within))
	eval[bool](u, expr+".click(), true")
}

// fill escribe value en el campo y dispara los eventos input y change.
func (u *ui) fill(sel, value string) {
	u.t.Helper()
	u.waitTrue(q(sel)+" !== null", "no hay un campo "+sel)
	eval[bool](u, fmt.Sprintf(`(e => { e.value = %s; e.dispatchEvent(new Event("input", {bubbles: true})); e.dispatchEvent(new Event("change", {bubbles: true})); return true })(%s)`,
		strconv.Quote(value), q(sel)))
}

// choose elige en el desplegable la opción cuyo texto es text, y dispara el
// evento change.
func (u *ui) choose(sel, text string) {
	u.t.Helper()
	opt := fmt.Sprintf(`[...(%s?.options ?? [])].find(o => o.textContent.trim() === %s)`, q(sel), strconv.Quote(text))
	u.waitTrue(opt+" !== undefined", fmt.Sprintf("el desplegable %s no tiene la opción %q (tiene %q)", sel, text, u.options(sel)))
	eval[bool](u, fmt.Sprintf(`(s => { s.value = %s.value; s.dispatchEvent(new Event("change", {bubbles: true})); return true })(%s)`, opt, q(sel)))
}

// options devuelve el texto de las opciones del desplegable, o nil si no
// existe o no es un desplegable.
func (u *ui) options(sel string) []string {
	u.t.Helper()
	return eval[[]string](u, fmt.Sprintf(`(s => s && s.options ? [...s.options].map(o => o.textContent.trim()) : null)(%s)`, q(sel)))
}

// selectedText devuelve el texto de la opción elegida en el desplegable.
func (u *ui) selectedText(sel string) string {
	u.t.Helper()
	return eval[string](u, fmt.Sprintf(`(s => s && s.selectedIndex >= 0 ? s.options[s.selectedIndex].textContent.trim() : "")(%s)`, q(sel)))
}

// valueOf devuelve el valor del campo o del desplegable.
func (u *ui) valueOf(sel string) string {
	u.t.Helper()
	return eval[string](u, q(sel)+"?.value ?? \"\"")
}

// text devuelve el texto visible del elemento, sin espacios en los extremos.
func (u *ui) text(sel string) string {
	u.t.Helper()
	return eval[string](u, q(sel)+"?.textContent.trim() ?? \"\"")
}

// visibleSections devuelve el id de las secciones de <main> que se ven.
func (u *ui) visibleSections() []string {
	u.t.Helper()
	return eval[[]string](u, `[...document.querySelectorAll("main > section[id]")].filter(s => !s.hidden && s.offsetParent !== null).map(s => s.id)`)
}

// hash devuelve el fragmento de la dirección (por ejemplo, "#backlog").
func (u *ui) hash() string {
	u.t.Helper()
	return eval[string](u, "location.hash")
}

// countAPICalls envuelve fetch para contar los pedidos a /api/ desde ahora; la
// cuenta queda en window.__llamadasAPI.
func (u *ui) countAPICalls() {
	u.t.Helper()
	eval[bool](u, `(() => { window.__llamadasAPI = 0; const f = window.fetch; window.fetch = (r, o) => { if (String(r).includes("/api/")) window.__llamadasAPI++; return f(r, o) }; return true })()`)
}

// apiCalls devuelve los pedidos a /api/ contados desde countAPICalls.
func (u *ui) apiCalls() int {
	u.t.Helper()
	return eval[int](u, "window.__llamadasAPI")
}

// uiData carga datos por la API del handler h, sin la traducción de rutas de
// do() (las rutas ya llevan los UUID que devolvió la API).
type uiData struct {
	t failer
	h http.Handler
}

// send hace el pedido, exige el código want y devuelve el "id" de la
// respuesta (vacío si no tiene).
func (d uiData) send(method, path, body string, want int) string {
	d.t.Helper()
	rec := serveRaw(d.h, method, path, body)
	if rec.Code != want {
		d.t.Fatalf("%s %s = %d, se esperaba %d; cuerpo: %s", method, path, rec.Code, want, rec.Body.String())
	}
	var resp struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	return resp.ID
}

// project crea un proyecto y devuelve su UUID.
func (d uiData) project(name string) string {
	d.t.Helper()
	return d.send(http.MethodPost, "/api/proyectos", fmt.Sprintf(`{"nombre":%q}`, name), http.StatusCreated)
}

// sprint crea un sprint abierto en el proyecto y devuelve su UUID.
func (d uiData) sprint(projectID, name string) string {
	d.t.Helper()
	return d.send(http.MethodPost, "/api/proyectos/"+projectID+"/sprints",
		fmt.Sprintf(`{"nombre":%q,"fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Objetivo"}`, name), http.StatusCreated)
}

// closeSprint cierra el sprint.
func (d uiData) closeSprint(sprintID string) {
	d.t.Helper()
	d.send(http.MethodPost, "/api/sprints/"+sprintID+"/cerrar", "", http.StatusOK)
}

// item crea un ítem en el backlog del proyecto y devuelve su UUID.
func (d uiData) item(projectID, title string) string {
	d.t.Helper()
	return d.send(http.MethodPost, "/api/proyectos/"+projectID+"/backlog",
		fmt.Sprintf(`{"titulo":%q,"descripcion":"","prioridad":"media","criterios":["Criterio"]}`, title), http.StatusCreated)
}

// assign asigna el ítem al sprint.
func (d uiData) assign(sprintID, itemID string) {
	d.t.Helper()
	d.send(http.MethodPost, "/api/sprints/"+sprintID+"/items/"+itemID, "", http.StatusOK)
}

// member registra un integrante en el proyecto y devuelve su UUID.
func (d uiData) member(projectID, name, role string) string {
	d.t.Helper()
	return d.send(http.MethodPost, "/api/proyectos/"+projectID+"/integrantes",
		fmt.Sprintf(`{"nombre":%q,"rol":%q}`, name, role), http.StatusCreated)
}

// effort registra un esfuerzo del integrante sobre el ítem en el sprint y
// devuelve su UUID.
func (d uiData) effort(sprintID, itemID, member string, hours float64) string {
	d.t.Helper()
	return d.send(http.MethodPost, "/api/sprints/"+sprintID+"/items/"+itemID+"/esfuerzos",
		fmt.Sprintf(`{"integrante":%q,"fecha":"2026-10-02","actividad":"Desarrollo","horas":%v}`, member, hours), http.StatusCreated)
}

// defect registra un defecto del ítem detectado en el sprint y devuelve su
// UUID.
func (d uiData) defect(itemID, sprintID, description string) string {
	d.t.Helper()
	return d.send(http.MethodPost, "/api/items/"+itemID+"/defectos",
		fmt.Sprintf(`{"descripcion":%q,"severidad":"ALTA","sprint_deteccion_id":%q}`, description, sprintID), http.StatusCreated)
}

// joined une textos para los mensajes de error.
func joined(s []string) string { return strings.Join(s, " | ") }
