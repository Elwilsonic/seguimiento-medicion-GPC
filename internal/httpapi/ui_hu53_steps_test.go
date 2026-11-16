package httpapi_test

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/chromedp/chromedp"
)

// HU-53 · Navegar la interfaz por secciones y elegir de listas (Issue #211).
//
// hu53 es un escenario de HU-53: carga los datos por la API, abre la interfaz
// en el navegador y verifica lo que se ve. Sus métodos son los pasos del
// .feature (ui_features_test.go) y los usan también los tests de Go
// (ui_navegacion_test.go, ui_listas_test.go y ui_acciones_test.go), así el
// mismo recorrido se prueba de las dos formas.

// menuSections son las secciones del menú, en orden, con su texto (D2).
var menuSections = []struct{ id, label string }{
	{"proyectos", "Proyectos"},
	{"integrantes", "Integrantes"},
	{"backlog", "Backlog"},
	{"sprints", "Sprints"},
	{"poker", "Planning Poker"},
	{"esfuerzo", "Esfuerzo"},
	{"defectos", "Defectos"},
	{"metricas", "Métricas y dashboard"},
	{"reportes", "Reportes"},
}

// fields son los desplegables por su nombre en el .feature.
var fields = map[string]string{
	"«Proyecto actual»":                          "#proyecto-actual",
	"«Sprint actual»":                            "#sprint-actual",
	"«Ítem» de Registrar esfuerzo":               "#form-registrar-esfuerzo [name=item_id]",
	"«Ítem» de Registrar defecto":                "#form-registrar-defecto [name=item_id]",
	"«Sprint de detección» de Registrar defecto": "#form-registrar-defecto [name=sprint_deteccion_id]",
	"«Sprint de resolución» de Cambiar estado":   "#form-cambiar-estado-defecto [name=sprint_resolucion_id]",
	"«Integrante» de Modificar integrante":       "#form-modificar-integrante [name=id]",
	"«Defecto» de Cambiar estado":                "#form-cambiar-estado-defecto [name=defecto_id]",
	"«Esfuerzo» de Corregir esfuerzo":            "#form-corregir-esfuerzo [name=esfuerzo_id]",
	"«Ítem» de Consultar esfuerzos":              "#form-consultar-esfuerzos [name=referencia]",
}

// messages son los mensajes de cada sección.
var messages = map[string]string{
	"backlog":     "#msg-backlog",
	"esfuerzo":    "#msg-esfuerzos",
	"defectos":    "#msg-defectos",
	"integrantes": "#msg-integrantes",
}

type hu53 struct {
	f       failer
	browser string // ejecutable del navegador
	width   int
	h       http.Handler
	data    uiData
	ids     map[string]string // nombre → UUID de proyectos, sprints, ítems y defectos
	project string            // último proyecto cargado
	sprint  string            // último sprint cargado
	u       *ui
	stops   []func()
	rec     int      // código de la última respuesta de la API
	visited []string // fragmentos recorridos con el menú
}

// newHU53 prepara un escenario nuevo con una API vacía.
func newHU53(f failer, browser string) *hu53 {
	s := &hu53{f: f, browser: browser, width: 1280, h: newHandler(), ids: map[string]string{}}
	s.data = uiData{t: f, h: s.h}
	return s
}

// close cierra el navegador y el servidor del escenario.
func (s *hu53) close() {
	for _, stop := range s.stops {
		stop()
	}
	s.stops = nil
}

func (s *hu53) field(name string) string {
	s.f.Helper()
	sel, ok := fields[name]
	if !ok {
		s.f.Fatalf("no hay un desplegable %s en el test", name)
	}
	return sel
}

// choices devuelve las opciones del desplegable con valor (sin las que solo
// avisan, como "Elegí un proyecto"), o nil si no es un desplegable.
func (s *hu53) choices(sel string) []string {
	s.f.Helper()
	return eval[[]string](s.u, `(s => s && s.options ? [...s.options].filter(o => o.value !== "").map(o => o.textContent.trim()) : null)(`+q(sel)+`)`)
}

// --- Datos (Dado) ---

func (s *hu53) existingProjects(a, b string) error {
	s.ids[a] = s.data.project(a)
	s.ids[b] = s.data.project(b)
	s.project = a
	return nil
}

func (s *hu53) ensureProject(name string) string {
	if id, ok := s.ids[name]; ok {
		return id
	}
	s.ids[name] = s.data.project(name)
	s.project = name
	return s.ids[name]
}

func (s *hu53) closedAndOpenSprints(project, closed, open string) error {
	p := s.ensureProject(project)
	s.ids[closed] = s.data.sprint(p, closed)
	s.data.closeSprint(s.ids[closed])
	s.ids[open] = s.data.sprint(p, open)
	s.sprint = open
	return nil
}

// closedSprint agrega un sprint cerrado. Como un proyecto tiene un solo
// sprint abierto, el .feature lo carga antes del abierto.
func (s *hu53) closedSprint(project, name string) error {
	p := s.ensureProject(project)
	s.ids[name] = s.data.sprint(p, name)
	s.data.closeSprint(s.ids[name])
	return nil
}

func (s *hu53) openSprintWithItem(project, sprint, item string) error {
	p := s.ensureProject(project)
	s.ids[sprint] = s.data.sprint(p, sprint)
	s.sprint = sprint
	s.ids[item] = s.data.item(p, item)
	s.data.assign(s.ids[sprint], s.ids[item])
	return nil
}

func (s *hu53) openSprintWithoutItems(project, sprint string) error {
	p := s.ensureProject(project)
	s.ids[sprint] = s.data.sprint(p, sprint)
	s.sprint = sprint
	return nil
}

func (s *hu53) itemWithoutSprint(project, item string) error {
	s.ids[item] = s.data.item(s.ensureProject(project), item)
	return nil
}

func (s *hu53) manyItems(project string, n int) error {
	p := s.ensureProject(project)
	for i := 1; i <= n; i++ {
		s.data.item(p, "Historia con un título largo para ver cómo se acomoda la tabla número "+strings.Repeat("x", i))
	}
	return nil
}

func (s *hu53) member(project, name, role string) error {
	s.ids[name] = s.data.member(s.ensureProject(project), name, role)
	return nil
}

func (s *hu53) effort(member string, hours int, item string) error {
	s.data.effort(s.ids[s.sprint], s.ids[item], member, float64(hours))
	return nil
}

func (s *hu53) defect(item, description string) error {
	s.ids[description] = s.data.defect(s.ids[item], s.ids[s.sprint], description)
	return nil
}

func (s *hu53) projectsFail(msg string) error {
	inner := s.h
	s.h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/proyectos" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
			return
		}
		inner.ServeHTTP(w, r)
	})
	return nil
}

func (s *hu53) windowWidth(width int) error {
	s.width = width
	return nil
}

// --- Abrir la interfaz ---

func (s *hu53) open(fragment string) error {
	s.f.Helper()
	ctx, url, stop := startBrowser(s.f, s.browser, s.h, s.width)
	s.stops = append(s.stops, stop)
	s.u = &ui{t: s.f, ctx: ctx, url: url}
	s.u.navigate("/" + fragment)
	s.markPage()
	return nil
}

func (s *hu53) openApp() error { return s.open("") }

func (s *hu53) openWithProjectAndSprint(project, sprint string) error {
	if err := s.open(""); err != nil {
		return err
	}
	if err := s.chooseIn(project, "«Proyecto actual»"); err != nil {
		return err
	}
	return s.chooseIn(sprint+" (ABIERTO)", "«Sprint actual»")
}

func (s *hu53) openWithProjectIn(project, fragment string) error {
	if err := s.open(fragment); err != nil {
		return err
	}
	return s.chooseIn(project, "«Proyecto actual»")
}

// --- Acciones (Cuando) ---

func (s *hu53) goTo(label string) error {
	s.f.Helper()
	s.u.clickText("#menu", label)
	return nil
}

func (s *hu53) eachMenuEntry() error {
	s.f.Helper()
	var got []string
	for _, sec := range menuSections {
		s.u.clickText("#menu", sec.label)
		s.u.waitTrue(`location.hash === "#`+sec.id+`"`, "la dirección no cambió a #"+sec.id)
		got = append(got, s.u.hash())
	}
	s.visited = got
	return nil
}

func (s *hu53) back() error {
	browse(s.f, s.u.ctx, chromedp.NavigateBack())
	return nil
}

func (s *hu53) refreshBacklogWithoutProject() error {
	s.u.countAPICalls()
	s.u.click("#actualizar-backlog")
	return nil
}

func (s *hu53) chooseIn(option, field string) error {
	s.f.Helper()
	s.u.choose(s.field(field), option)
	return nil
}

func (s *hu53) createProjectFromSection(name string) error {
	s.goTo("Proyectos")
	s.u.fill("#form-crear-proyecto [name=nombre]", name)
	s.u.click("#form-crear-proyecto button[type=submit]")
	return nil
}

func (s *hu53) createItemFromSection(title string) error {
	s.goTo("Backlog")
	s.u.fill("#form-crear-item [name=titulo]", title)
	s.u.fill("#form-crear-item [name=criterios]", "Criterio")
	s.u.click("#form-crear-item button[type=submit]")
	return nil
}

func (s *hu53) registerEffort(member string, hours int, item string) error {
	// HU-54 (D5): «Integrante» es un desplegable con los integrantes del
	// proyecto. Si el integrante del escenario no está registrado, se
	// registra por la API y se vuelven a elegir el proyecto y el sprint para
	// recargar las listas, sin cambiar lo que verifica el escenario.
	if !slices.Contains(s.u.options(effortMemberField), member) {
		s.data.member(s.ids[s.project], member, "PRODUCT_BUILDER")
		s.chooseIn(s.project, "«Proyecto actual»")
		s.chooseIn(s.sprint+" (ABIERTO)", "«Sprint actual»")
	}
	s.goTo("Esfuerzo")
	s.u.choose(fields["«Ítem» de Registrar esfuerzo"], item)
	s.u.choose(effortMemberField, member)
	s.u.fill("#form-registrar-esfuerzo [name=fecha]", "2026-10-02")
	s.u.fill("#form-registrar-esfuerzo [name=actividad]", "Desarrollo")
	s.u.fill("#form-registrar-esfuerzo [name=horas]", itoa(hours))
	s.u.click("#form-registrar-esfuerzo button[type=submit]")
	return nil
}

func (s *hu53) tryRegisterEffort() error {
	s.goTo("Esfuerzo")
	s.u.countAPICalls()
	s.u.click("#form-registrar-esfuerzo button[type=submit]")
	return nil
}

func (s *hu53) registerDefect(description, item, sprint string) error {
	s.goTo("Defectos")
	s.u.choose(fields["«Ítem» de Registrar defecto"], item)
	s.u.fill("#form-registrar-defecto [name=descripcion]", description)
	s.u.fill("#form-registrar-defecto [name=severidad]", "ALTA")
	s.u.choose(fields["«Sprint de detección» de Registrar defecto"], sprint)
	s.u.click("#form-registrar-defecto button[type=submit]")
	return nil
}

// rowAction hace clic en el botón action de la fila que contiene text dentro
// de la tabla tbody.
func (s *hu53) rowAction(tbody, text, action string) {
	s.f.Helper()
	row := `[...document.querySelectorAll(` + quote(tbody+" tr") + `)].find(r => r.textContent.includes(` + quote(text) + `))`
	s.u.waitTrue(row+" !== undefined", "no hay una fila con "+text+" en "+tbody)
	btn := `[...` + row + `.querySelectorAll("button")].find(b => b.textContent.trim() === ` + quote(action) + `)`
	s.u.waitTrue(btn+" !== undefined", "la fila "+text+" no tiene el botón "+action)
	eval[bool](s.u, btn+".click(), true")
}

// showTable llena la tabla de la sección con su botón de consulta.
func (s *hu53) showTable(section string) string {
	s.f.Helper()
	switch section {
	case "backlog":
		s.goTo("Backlog")
		s.u.click("#actualizar-backlog")
		return "#items"
	case "esfuerzo":
		s.goTo("Esfuerzo")
		s.u.fill("#form-consultar-esfuerzos [name=tipo]", "sprint")
		s.u.click("#form-consultar-esfuerzos button[type=submit]")
		return "#esfuerzos"
	case "defectos":
		s.goTo("Defectos")
		s.u.click("#form-consultar-defectos button[type=submit]")
		return "#lista-defectos"
	case "integrantes":
		s.goTo("Integrantes")
		s.u.click("#ver-integrantes")
		return "#lista-integrantes"
	}
	s.f.Fatalf("no hay una tabla de %s en el test", section)
	return ""
}

func (s *hu53) backlogRowAction(action, item string) error {
	s.rowAction(s.showTable("backlog"), item, action)
	return nil
}

func (s *hu53) memberRowAction(action, member string) error {
	s.rowAction(s.showTable("integrantes"), member, action)
	return nil
}

func (s *hu53) correctEffortTo(member string, hours int) error {
	s.rowAction(s.showTable("esfuerzo"), member, "Corregir horas")
	s.u.fill("#form-corregir-esfuerzo [name=horas]", itoa(hours))
	s.u.click("#form-corregir-esfuerzo button[type=submit]")
	return nil
}

func (s *hu53) walkAllSections() error {
	for _, sec := range menuSections {
		s.goTo(sec.label)
	}
	return nil
}

func (s *hu53) apiGet(path string) error {
	s.rec = serveRaw(s.h, http.MethodGet, path, "").Code
	return nil
}

// --- Verificaciones (Entonces) ---

func (s *hu53) hashesInOrder(list string) error {
	want := splitList(list)
	if !slices.Equal(s.visited, want) {
		s.f.Fatalf("fragmentos = %q, se esperaba %q", s.visited, want)
	}
	return nil
}

func (s *hu53) onlySection(id string) error {
	s.f.Helper()
	s.u.waitTrue(`(() => { const v = [...document.querySelectorAll("main > section[id]")].filter(s => !s.hidden && s.offsetParent !== null).map(s => s.id); return v.length === 1 && v[0] === `+quote(id)+` })()`,
		"no se ve solo la sección "+id+"; se ven "+joined(s.u.visibleSections()))
	return nil
}

func (s *hu53) hashIs(want string) error {
	if got := s.u.hash(); got != want {
		s.f.Fatalf("fragmento = %q, se esperaba %q", got, want)
	}
	return nil
}

func (s *hu53) menuMarked(label string) error {
	got := eval[string](s.u, `[...document.querySelectorAll('#menu [aria-current="page"]')].map(a => a.textContent.trim()).join(",")`)
	if got != label {
		s.f.Fatalf("entrada marcada del menú = %q, se esperaba %q", got, label)
	}
	return nil
}

func (s *hu53) markPage() {
	eval[bool](s.u, "window.__sinRecargar = true, true")
}

func (s *hu53) notReloaded() error {
	if !eval[bool](s.u, "window.__sinRecargar === true") {
		s.f.Fatalf("la página se recargó")
	}
	return nil
}

func (s *hu53) messageIs(section, want string) error {
	s.f.Helper()
	sel := messages[section]
	s.u.waitTrue(q(sel)+` && !`+q(sel)+`.hidden && `+q(sel)+`.textContent.trim() === `+quote(want),
		"el mensaje de "+section+" es "+quote(s.u.text(sel))+", se esperaba "+quote(want))
	return nil
}

func (s *hu53) errorIs(section, want string) error {
	s.f.Helper()
	if err := s.messageIs(section, want); err != nil {
		return err
	}
	if !eval[bool](s.u, q(messages[section])+`.classList.contains("error")`) {
		s.f.Fatalf("el mensaje de %s no se muestra como error", section)
	}
	return nil
}

func (s *hu53) headerMessageIs(want string) error {
	s.u.waitTrue(`!`+q("#msg-contexto")+`.hidden && `+q("#msg-contexto")+`.textContent.trim() === `+quote(want),
		"el mensaje del encabezado es "+quote(s.u.text("#msg-contexto"))+", se esperaba "+quote(want))
	return nil
}

func (s *hu53) noAPICalls() error {
	if n := s.u.apiCalls(); n != 0 {
		s.f.Fatalf("se hicieron %d pedidos a la API, se esperaba ninguno", n)
	}
	return nil
}

func (s *hu53) offers(field, list string) error {
	s.f.Helper()
	sel := s.field(field)
	want := splitList(list)
	s.u.waitTrue(`(() => { const s = `+q(sel)+`; return s && s.options && JSON.stringify([...s.options].filter(o => o.value !== "").map(o => o.textContent.trim())) === `+quote(toJSON(want))+` })()`,
		field+" ofrece "+quote(joined(s.choices(sel)))+", se esperaba "+quote(joined(want)))
	return nil
}

func (s *hu53) offersNothing(field string) error {
	if got := s.choices(s.field(field)); got == nil || len(got) != 0 {
		s.f.Fatalf("%s ofrece %q, se esperaba que no ofrezca nada", field, got)
	}
	return nil
}

func (s *hu53) shows(field, want string) error {
	s.f.Helper()
	sel := s.field(field)
	s.u.waitTrue(`(s => s && s.selectedIndex >= 0 && s.options[s.selectedIndex].textContent.trim() === `+quote(want)+`)(`+q(sel)+`)`,
		field+" muestra "+quote(s.u.selectedText(sel))+", se esperaba "+quote(want))
	return nil
}

func (s *hu53) noCurrent(field string) error {
	if v := s.u.valueOf(s.field(field)); v != "" {
		s.f.Fatalf("%s vale %q, se esperaba que no haya ninguno elegido", field, v)
	}
	return nil
}

func (s *hu53) defectDescriptionIs(want string) error {
	if got := s.u.valueOf("#form-registrar-defecto [name=descripcion]"); got != want {
		s.f.Fatalf("descripción del defecto = %q, se esperaba %q", got, want)
	}
	return nil
}

func (s *hu53) rowsHave(section, list string) error {
	s.f.Helper()
	tbody := s.showTable(section)
	for _, action := range splitList(list) {
		s.u.waitTrue(`(() => { const r = [...document.querySelectorAll(`+quote(tbody+" tr")+`)]; return r.length > 0 && r.every(f => [...f.querySelectorAll("button")].some(b => b.textContent.trim() === `+quote(action)+`)) })()`,
			"no todas las filas de "+section+" tienen el botón "+action)
	}
	return nil
}

func (s *hu53) memberFormHas(name, role string) error {
	if got := s.u.valueOf("#form-modificar-integrante [name=nombre]"); got != name {
		s.f.Fatalf("nombre = %q, se esperaba %q", got, name)
	}
	if got := s.u.valueOf("#form-modificar-integrante [name=rol]"); got != role {
		s.f.Fatalf("rol = %q, se esperaba %q", got, role)
	}
	return nil
}

// menuBeside verifica que el menú esté al costado del contenido.
func (s *hu53) menuBeside() error {
	if !eval[bool](s.u, `(() => { const m = document.querySelector("#menu"), c = document.querySelector("main"); if (!m || !c) return false; const a = m.getBoundingClientRect(), b = c.getBoundingClientRect(); return a.right <= b.left + 1 && a.height > 0 })()`) {
		s.f.Fatalf("el menú no está al costado del contenido")
	}
	return nil
}

// menuAbove verifica que el menú esté arriba del contenido.
func (s *hu53) menuAbove() error {
	if !eval[bool](s.u, `(() => { const m = document.querySelector("#menu"), c = document.querySelector("main"); if (!m || !c) return false; const a = m.getBoundingClientRect(), b = c.getBoundingClientRect(); return a.bottom <= b.top + 1 && a.height > 0 })()`) {
		s.f.Fatalf("el menú no está arriba del contenido")
	}
	return nil
}

func (s *hu53) noHorizontalScroll() error {
	s.showTable("backlog")
	if !eval[bool](s.u, "document.documentElement.scrollWidth <= window.innerWidth") {
		s.f.Fatalf("la página mide %d px de ancho en una ventana de %d px",
			eval[int](s.u, "document.documentElement.scrollWidth"), eval[int](s.u, "window.innerWidth"))
	}
	return nil
}

func (s *hu53) backlogTableScrolls() error {
	if !eval[bool](s.u, `(() => { const t = document.querySelector("#items").closest(".tabla"); const o = getComputedStyle(t).overflowX; return (o === "auto" || o === "scroll") && t.scrollWidth > t.clientWidth })()`) {
		s.f.Fatalf("la tabla del backlog no se desplaza dentro de su recuadro")
	}
	return nil
}

func (s *hu53) sameOrigin() error {
	others := eval[[]string](s.u, `performance.getEntriesByType("resource").map(e => e.name).filter(n => new URL(n).origin !== location.origin)`)
	if len(others) != 0 {
		s.f.Fatalf("pedidos a otros dominios: %q", others)
	}
	return nil
}

func (s *hu53) apiAnswered(want int) error {
	if s.rec != want {
		s.f.Fatalf("la API respondió %d, se esperaba %d", s.rec, want)
	}
	return nil
}

// noItemChosen verifica que «Ítem» de Registrar esfuerzo no tenga un ítem
// elegido.
func (s *hu53) noItemChosen() error {
	return s.noCurrent("«Ítem» de Registrar esfuerzo")
}

// --- Utilidades ---

func splitList(list string) []string {
	var out []string
	for _, p := range strings.Split(list, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// toJSON devuelve v en JSON, para armar expresiones de JavaScript.
func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// quote devuelve el texto como literal de JavaScript.
func quote(s string) string { return toJSON(s) }

func itoa(n int) string { return strconv.Itoa(n) }
