package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
)

// TEC-04 · Interfaz usable (Issue #34)
//
// Trazabilidad: la interfaz web se sirve desde el mismo handler que la API:
//   - CA-TEC04.8: la interfaz web responde en /.
//   - CA-TEC04.4: sus archivos van dentro del binario (embed), sin librerías
//     ni recursos externos.
// CA-TEC04.1 y CA-TEC04.2 (uso desde el navegador y errores entendibles) se
// verifican con el recorrido en el navegador de la evidencia TDD.

// get hace un GET al handler sin exigir una respuesta JSON.
func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// TEC-04 · CA-TEC04.8 · Verifica que la interfaz web responda en /.
func TestInterfaz_RaizDevuelveLaPaginaHTML(t *testing.T) {
	rec := get(t, newHandler(), "/")

	expectStatus(t, rec, http.StatusOK)
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, se esperaba text/html", ct)
	}
	if body := rec.Body.String(); !strings.Contains(body, `<script src="app.js"></script>`) {
		t.Errorf("la página no carga app.js; cuerpo: %s", body)
	}
}

// TEC-04 · CA-TEC04.4 · Verifica que el JavaScript y los estilos de la interfaz
// se sirvan desde el binario, con su tipo de contenido.
func TestInterfaz_SirveLosArchivosEstaticos(t *testing.T) {
	cases := map[string]string{
		"/app.js":     "text/javascript",
		"/styles.css": "text/css",
	}
	for path, wantType := range cases {
		t.Run(path, func(t *testing.T) {
			rec := get(t, newHandler(), path)

			expectStatus(t, rec, http.StatusOK)
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, wantType) {
				t.Errorf("Content-Type = %q, se esperaba %s", ct, wantType)
			}
			if rec.Body.Len() == 0 {
				t.Error("el archivo está vacío")
			}
		})
	}
}

// HU-20 · Asignar story points a un ítem del backlog (Issue #38)
//
// Trazabilidad: CA-20.8, la tabla del backlog muestra los story points de cada
// ítem y permite asignarlos. El recorrido en el navegador queda en la
// evidencia TDD.

// HU-20 · CA-20.8 · Verifica que la tabla del backlog tenga la columna SP y que
// app.js muestre y asigne los Story Points con PATCH /story-points.
func TestInterfaz_TablaDelBacklogMuestraYAsignaStoryPoints(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	if !strings.Contains(page, "<th>SP</th>") {
		t.Error("la tabla del backlog no tiene la columna SP")
	}
	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`"/story-points"`, `"Asignar SP"`, "it.story_points"} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %s", want)
		}
	}
}

// CA-20.8 y SDD de HU-20 ("Escala de Story Points"): el select de app.js usa
// los mismos valores que backlog.StoryPointScale(), para que la interfaz no se
// desincronice de la escala del paquete.
// HU-26 · Registrar esfuerzo (Issue #44)
//
// Trazabilidad: CA-26.11. La interfaz tiene un formulario para registrar el
// esfuerzo y el campo Horas acepta cualquier decimal positivo (step="any"),
// igual que la API.

// HU-26 · CA-26.11 · Verifica que el formulario "Registrar esfuerzo" acepte
// horas con decimales (step="any") y que app.js lo envíe a /esfuerzos.
func TestInterfaz_FormularioDeEsfuerzoUsaStepAny(t *testing.T) {
	h := newHandler()
	page := get(t, h, "/").Body.String()
	for _, want := range []string{`id="form-registrar-esfuerzo"`, `name="horas"`, `type="number"`, `step="any"`} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %s", want)
		}
	}

	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`"/esfuerzos"`, `"Registrar esfuerzo"`, `"Esfuerzo "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %s", want)
		}
	}
}

// HU-20 · CA-20.8 · Verifica que el select de Story Points de app.js use los
// mismos valores que backlog.StoryPointScale().
func TestInterfaz_SelectDeStoryPointsUsaLaEscalaDelBacklog(t *testing.T) {
	scale := backlog.StoryPointScale()
	if len(scale) == 0 {
		t.Fatal("backlog.StoryPointScale() devolvió una lista vacía")
	}
	values := make([]string, len(scale))
	for i, v := range scale {
		values[i] = strconv.Itoa(v)
	}
	want := "const escalaStoryPoints = [" + strings.Join(values, ", ") + "];"

	js := get(t, newHandler(), "/app.js").Body.String()

	if !strings.Contains(js, want) {
		t.Errorf("app.js no define %s", want)
	}
}

// HU-21 · Abrir una sesión de Planning Poker (Issue #39)
//
// Trazabilidad: CA-21.7, desde el backlog se abre una sesión y la interfaz
// muestra sus datos y las cartas. El recorrido en el navegador queda en la
// evidencia TDD.

// HU-21 · CA-21.7 · Verifica que el backlog tenga el botón "Abrir Poker" y que
// la interfaz muestre la sesión abierta y sus cartas.
func TestInterfaz_BacklogPermiteAbrirPlanningPoker(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	for _, want := range []string{"Planning Poker", `id="datos-poker"`, `id="cartas-poker"`} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %q", want)
		}
	}

	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`"/poker-sessions"`, `"Abrir Poker"`, "s.cartas"} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-22 · CA-22.7 · Verifica que la interfaz tenga el formulario de voto, con el
// integrante y los botones de cartas, y que app.js envíe el voto a /votes.
// Votar de forma oculta (Issue #40).
func TestInterfaz_FormularioDeVoto(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	for _, want := range []string{`id="form-votar-poker"`, `name="integrante"`, `id="botones-cartas-poker"`, `id="msg-poker"`} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %q", want)
		}
	}
	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`"/poker-sessions/"`, `"/votes"`, `"Votar "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-23 · CA-23.8 · Verifica que el botón "Revelar votos" revele la sesión y la
// interfaz muestre la lista de votos.
func TestInterfaz_BotonRevelarVotos(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	for _, want := range []string{`id="btn-revelar-poker"`, "Revelar votos", `id="votos-poker"`} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %q", want)
		}
	}
	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`"/reveal"`, `"Votos revelados en la sesión "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-25 · CA-25.9 · Verifica que el formulario "Registrar estimación acordada"
// registre la carta elegida, muestre el mensaje y recargue el backlog.
func TestInterfaz_FormularioRegistrarEstimacion(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	for _, want := range []string{`id="form-estimar-poker"`, "Registrar estimación acordada", `id="botones-estimar-poker"`} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %q", want)
		}
	}
	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`"/estimate"`, `"Estimación registrada: "`, `" story points en la sesión "`, "cargarBacklog"} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-24 · CA-24.12 · Verifica que después de revelar se muestren la comparación
// y el botón "Repetir ronda", dentro del bloque de la estimación. El botón es
// type="button" para que el clic no envíe el formulario.
func TestInterfaz_DiferenciasYBotonRepetirRonda(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	start := strings.Index(page, `<form id="form-estimar-poker"`)
	if start < 0 {
		t.Fatal(`la página no contiene <form id="form-estimar-poker"`)
	}
	end := strings.Index(page[start:], "</form>")
	if end < 0 {
		t.Fatal("el formulario form-estimar-poker no se cierra")
	}
	form := page[start : start+end]
	for _, want := range []string{`id="diferencias-poker"`, `id="btn-repetir-ronda-poker"`, "Repetir ronda"} {
		if !strings.Contains(form, want) {
			t.Errorf("form-estimar-poker no contiene %q", want)
		}
	}
	if id := strings.Index(form, `id="btn-repetir-ronda-poker"`); id >= 0 {
		open := strings.LastIndex(form[:id], "<button")
		closeTag := strings.Index(form[id:], ">")
		if open < 0 || closeTag < 0 {
			t.Fatal(`btn-repetir-ronda-poker no es un <button>`)
		}
		if tag := form[open : id+closeTag]; !strings.Contains(tag, `type="button"`) {
			t.Errorf(`botón %s, se esperaba type="button"`, tag)
		}
	}
	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`"/rounds"`, "s.comparacion", "mostrarDiferenciasPoker", `"Hay diferencias: menor "`, `"Todos votaron "`, `" iniciada en la sesión "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-32 · CA-32.9 · Verifica que la sección "Métricas" muestre los Story Points
// del proyecto y del sprint actual, y la velocidad.
func TestInterfaz_SeccionDeMetricas(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	for _, want := range []string{"Métricas", "Ver métricas", `id="datos-metricas"`} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %q", want)
		}
	}
	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`"/metricas"`, "sp_planificados", "sp_completados", "items_sin_estimar", "velocidad"} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-34 · CA-34.6 · Verifica que la sección "Métricas" muestre el porcentaje de
// historias completadas con un decimal y las cantidades, o "sin datos" si no
// hay ítems.
func TestInterfaz_MetricasMuestraElPorcentajeDeHistoriasCompletadas(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{"porcentaje_completadas", "items_terminados", "items_totales", "Historias completadas", "sin datos"} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-33 · CA-33.5 · Verifica que la sección "Métricas" muestre M4 a M7 y la
// cantidad de ítems sin horas estimadas, con un decimal y "sin datos" cuando la
// desviación no está disponible. También cubre CA-33.9.
func TestInterfaz_MetricasMuestraHorasYDesviacion(t *testing.T) {
	h := newHandler()
	page := get(t, h, "/").Body.String()
	for _, want := range []string{"Métricas", "Ver métricas", "las horas"} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %q", want)
		}
	}

	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{
		"horas_estimadas",
		"horas_reales",
		"desviacion_horas",
		"desviacion_porcentual",
		"items_sin_horas_estimadas",
		`"Horas: "`,
		`" estimadas, "`,
		`" reales, "`,
		`" ítem(s) sin horas estimadas. Desviación: "`,
		"sin datos",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-33 · CA-33.9 · Verifica que numero1 convierta el texto "-0.0" en "0.0".
func TestInterfaz_Numero1NormalizaElCeroNegativo(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{"function numero1", `"-0.0"`, `"0.0"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-29 · Registrar defecto (Issue #47)
//
// Trazabilidad: CA-29.11. La interfaz tiene una sección "Defectos" con el
// formulario "Registrar defecto", que llama a POST /items/{id}/defectos con el
// sprint de detección informado. El mensaje de éxito y el de error usan el
// mismo mecanismo que los demás formularios.

// HU-29 · CA-29.11 · Verifica que la sección "Defectos" tenga el formulario
// "Registrar defecto" con sus campos y que app.js lo envíe a /defectos.
func TestInterfaz_FormularioDeDefecto(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	indice := strings.Index(page, `<section id="defectos">`)
	if indice == -1 {
		t.Fatal("la página no contiene la sección de defectos")
	}
	seccionDefectos := page[indice:]

	for _, want := range []string{
		`<section id="defectos">`,
		"<h2>Defectos</h2>",
		`id="form-registrar-defecto"`,
		"Registrar defecto",
		`name="sprint_deteccion_id" type="number" min="1" required`,
		`id="msg-defectos"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %q", want)
		}
	}
	for _, want := range []string{`name="item_id"`, `name="descripcion"`, `name="severidad"`} {
		if !strings.Contains(seccionDefectos, want) {
			t.Errorf("la sección de defectos no contiene %q", want)
		}
	}

	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`$("form-registrar-defecto")`, `"/defectos"`, `"Registrar defecto"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-29 · CA-29.11 · Verifica que, al registrar un defecto, la interfaz muestre
// el mensaje de éxito con el defecto y el ítem.
func TestInterfaz_RegistrarDefectoMuestraElMensajeDeExito(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`mostrar("msg-defectos", "Defecto "`, `" registrado para el ítem "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-29 · CA-29.11 · Verifica que, si registrar un defecto falla, la interfaz
// muestre el mensaje de error en msg-defectos.
func TestInterfaz_RegistrarDefectoMuestraElMensajeDeError(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	if !strings.Contains(page, `<p class="mensaje" id="msg-defectos" role="alert" hidden></p>`) {
		t.Error("la página no tiene el mensaje de error de defectos")
	}

	js := get(t, h, "/app.js").Body.String()
	if !strings.Contains(js, `accion("msg-defectos"`) {
		t.Error("app.js no muestra los errores de defectos con accion")
	}
}

// HU-29 · CA-29.11 · Verifica que la lista de severidades de app.js use los
// mismos valores que defect.SeverityValues(), para que la interfaz no se
// desincronice de la validación del paquete (SDD de HU-29, "Interfaz web").
func TestInterfaz_ListaDeSeveridadesDeDefectosCoincideConElPaquete(t *testing.T) {
	values := defect.SeverityValues()
	if len(values) == 0 {
		t.Fatal("defect.SeverityValues() devolvió una lista vacía")
	}
	want := `const severidadesDefecto = ["` + strings.Join(values, `", "`) + `"];`

	js := get(t, newHandler(), "/app.js").Body.String()

	if !strings.Contains(js, want) {
		t.Errorf("app.js no define %s", want)
	}
}

// HU-30 · Cambiar el estado de un defecto (Issue #48)
//
// Trazabilidad: CA-30.9. La sección "Defectos" tiene, debajo de "Registrar
// defecto", el formulario "Cambiar estado de defecto", que llama a
// PATCH /defectos/{id}/estado y envía el sprint de resolución solo si tiene
// valor. El mensaje de éxito y el de error usan el mismo mecanismo que los
// demás formularios.

// HU-30 · CA-30.9 · Verifica que el formulario "Cambiar estado de defecto" esté
// debajo de "Registrar defecto", con sus campos, y que app.js envíe el PATCH a
// /defectos/{id}/estado.
func TestInterfaz_FormularioCambiarEstadoDeDefecto(t *testing.T) {
	h := newHandler()

	page := get(t, h, "/").Body.String()
	registrar := strings.Index(page, `id="form-registrar-defecto"`)
	cambiar := strings.Index(page, `id="form-cambiar-estado-defecto"`)
	if cambiar == -1 {
		t.Fatal("la página no contiene el formulario para cambiar el estado de un defecto")
	}
	if registrar == -1 || cambiar < registrar {
		t.Error("el formulario para cambiar el estado no está debajo de \"Registrar defecto\"")
	}
	formulario := page[cambiar:]
	if fin := strings.Index(formulario, "</form>"); fin != -1 {
		formulario = formulario[:fin]
	}
	for _, want := range []string{
		"Cambiar estado de defecto",
		`name="defecto_id" type="number" min="1" required`,
		`<select name="estado" required></select>`,
		`name="sprint_resolucion_id" type="number" min="1" placeholder`,
	} {
		if !strings.Contains(formulario, want) {
			t.Errorf("el formulario para cambiar el estado no contiene %q", want)
		}
	}

	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{
		`$("form-cambiar-estado-defecto")`,
		`"Cambiar estado de defecto"`,
		`api("PATCH", "/defectos/" + defecto + "/estado"`,
		`sprint_resolucion_id`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-30 · CA-30.9 · Verifica que, al cambiar el estado de un defecto, la
// interfaz muestre el estado nuevo y el sprint de resolución.
func TestInterfaz_CambiarEstadoDeDefectoMuestraElMensajeDeExito(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`" ahora está en "`, `" resuelto en el sprint "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-30 · CA-30.9 · Verifica que, si cambiar el estado falla, la interfaz
// muestre el mensaje de error, igual que al registrar un defecto.
func TestInterfaz_CambiarEstadoDeDefectoMuestraElMensajeDeError(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	// Registrar y cambiar el estado muestran sus errores en msg-defectos.
	if n := strings.Count(js, `accion("msg-defectos"`); n < 2 {
		t.Errorf("app.js muestra con accion los errores de %d formulario(s) de defectos, se esperaban 2", n)
	}
}

// HU-30 · CA-30.9 · Verifica que la lista de estados de app.js use los mismos
// valores que defect.StatusValues() (SDD de HU-30, "Interfaz web").
func TestInterfaz_ListaDeEstadosDeDefectosCoincideConElPaquete(t *testing.T) {
	values := defect.StatusValues()
	if len(values) == 0 {
		t.Fatal("defect.StatusValues() devolvió una lista vacía")
	}
	want := `const estadosDefecto = ["` + strings.Join(values, `", "`) + `"];`

	js := get(t, newHandler(), "/app.js").Body.String()

	if !strings.Contains(js, want) {
		t.Errorf("app.js no define %s", want)
	}
}

// HU-31 · Consultar los defectos de un proyecto o de un sprint (Issue #49)
//
// Trazabilidad: CA-31.9. La sección "Defectos" tiene, debajo de "Cambiar
// estado de defecto", el formulario "Consultar defectos" para el proyecto o el
// sprint actual, con un filtro de estado opcional, una tabla y la cantidad. El
// recorrido en el navegador queda en la evidencia TDD.

// formularioConsultarDefectos devuelve el HTML del formulario de consulta.
func formularioConsultarDefectos(t *testing.T, page string) string {
	t.Helper()
	inicio := strings.Index(page, `<form id="form-consultar-defectos">`)
	if inicio == -1 {
		t.Fatal("la página no contiene el formulario para consultar defectos")
	}
	formulario := page[inicio:]
	if fin := strings.Index(formulario, "</form>"); fin != -1 {
		formulario = formulario[:fin]
	}
	return formulario
}

// HU-31 · CA-31.9 · Verifica que el formulario "Consultar defectos" esté debajo
// de "Cambiar estado de defecto", con el tipo, el filtro de estado, la tabla,
// la cantidad y el mensaje.
func TestInterfaz_FormularioConsultarDefectos(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	formulario := formularioConsultarDefectos(t, page)
	cambiar := strings.Index(page, `id="form-cambiar-estado-defecto"`)
	if cambiar == -1 || strings.Index(page, `id="form-consultar-defectos"`) < cambiar {
		t.Error("el formulario para consultar defectos no está debajo de \"Cambiar estado de defecto\"")
	}
	for _, want := range []string{
		"Consultar defectos",
		`<select name="tipo">`,
		`<option value="proyecto">Proyecto actual</option>`,
		`<option value="sprint">Sprint actual</option>`,
		`<select name="estado">`,
		`<option value="">Todos</option>`,
		`<button type="submit">Consultar</button>`,
	} {
		if !strings.Contains(formulario, want) {
			t.Errorf("el formulario para consultar defectos no contiene %q", want)
		}
	}
	for _, want := range []string{
		`<tbody id="lista-defectos"></tbody>`,
		`id="cantidad-defectos"`,
		`<p class="mensaje" id="msg-consultar-defectos" role="alert" hidden></p>`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %q", want)
		}
	}
}

// HU-31 · CA-31.9 · Verifica que la tabla de defectos tenga sus columnas, debajo
// del formulario de consulta.
func TestInterfaz_TablaDeDefectosTieneLasColumnas(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()
	inicio := strings.Index(page, `<form id="form-consultar-defectos">`)
	fin := strings.Index(page, `id="lista-defectos"`)
	if inicio == -1 || fin == -1 || fin < inicio {
		t.Fatal("la página no contiene la tabla de defectos debajo del formulario de consulta")
	}
	tabla := page[inicio:fin]

	for _, want := range []string{"<th>ID</th>", "<th>Descripción</th>", "<th>Severidad</th>", "<th>Estado</th>", "<th>Ítem</th>", "<th>Sprint de detección</th>", "<th>Sprint de resolución</th>"} {
		if !strings.Contains(tabla, want) {
			t.Errorf("la tabla de defectos no contiene %q", want)
		}
	}
}

// HU-31 · CA-31.9 · Verifica que "Consultar defectos" llame a la ruta del
// proyecto o del sprint actual, con el filtro ?estado=, y muestre la lista, la
// cantidad o "Sin defectos.".
func TestInterfaz_ConsultarDefectosLlamaALaAPI(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`$("form-consultar-defectos")`,
		`"Consultar defectos"`,
		`"/proyectos/" + proyectoActual() + "/defectos"`,
		`"/sprints/" + sprintActual() + "/defectos"`,
		`"?estado=" + encodeURIComponent(`,
		`$("lista-defectos")`,
		`"Sin defectos."`,
		`"Cantidad: "`,
		`" defecto(s)"`,
		`"—"`,
		`accion("msg-consultar-defectos"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-31 · CA-31.9 · Verifica que el filtro de estado use la lista de HU-30, que
// ya se compara con defect.StatusValues(), y no repita los estados (SDD de
// HU-31, "Interfaz web").
func TestInterfaz_FiltroDeEstadoDeDefectosUsaLaListaDelPaquete(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	if n := strings.Count(js, "const estadosDefecto = "); n != 1 {
		t.Errorf("app.js define la lista de estados de defectos %d veces, se esperaba 1", n)
	}
	if !strings.Contains(js, "for (const estado of estadosDefecto) {\n  formConsultaDefectos.estado.append(") {
		t.Error("el filtro de estado de \"Consultar defectos\" no se arma con estadosDefecto")
	}
}

// HU-27 — Consultar los registros de esfuerzo (Issue #45)
//
// Trazabilidad: CA-27.10. La interfaz tiene un formulario "Consultar
// esfuerzos" con selector Sprint/Historia, una tabla y el total con un
// decimal. El recorrido en el navegador queda en la evidencia TDD.

// HU-27 · CA-27.10 · Verifica que el formulario "Consultar esfuerzos" permita
// elegir sprint o historia y que app.js muestre la lista y el total con un
// decimal. También cubre CA-27.2.
func TestInterfaz_FormularioDeConsultaDeEsfuerzos(t *testing.T) {
	h := newHandler()
	page := get(t, h, "/").Body.String()

	for _, want := range []string{
		`id="form-consultar-esfuerzos"`,
		"Consultar esfuerzos",
		`<option value="sprint">Sprint</option>`,
		`<option value="item">Historia</option>`,
		`name="referencia" type="number" min="1"`,
		`id="esfuerzos"`,
		`id="total-esfuerzos"`,
		`id="msg-esfuerzos"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("la página no contiene %q", want)
		}
	}

	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{
		`$("form-consultar-esfuerzos")`,
		`"/esfuerzos"`,
		`"Consultar esfuerzos"`,
		`"total-esfuerzos"`,
		".toFixed(1)",
		`"No hay esfuerzos registrados."`,
		`accion("msg-esfuerzos"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-27 · CA-27.10 · Verifica que la tabla de esfuerzos tenga sus columnas.
func TestInterfaz_TablaDeEsfuerzosTieneLasColumnas(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()
	indice := strings.Index(page, `<form id="form-consultar-esfuerzos">`)
	if indice == -1 {
		t.Fatal("la página no contiene el formulario de consulta de esfuerzos")
	}
	seccion := page[indice:]

	for _, want := range []string{"<th>ID</th>", "<th>Fecha</th>", "<th>Integrante</th>", "<th>Actividad</th>", "<th>Ítem</th>", "<th>Sprint</th>", "<th>Horas</th>"} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la tabla de esfuerzos no contiene %q", want)
		}
	}
}

// HU-27 · CA-27.10 · Verifica que, si la consulta de esfuerzos falla, la
// interfaz muestre el mensaje de error.
func TestInterfaz_ConsultaDeEsfuerzosMuestraElError(t *testing.T) {
	h := newHandler()
	page := get(t, h, "/").Body.String()
	if !strings.Contains(page, `<p class="mensaje" id="msg-esfuerzos" role="alert" hidden></p>`) {
		t.Error("la página no tiene el mensaje de consulta de esfuerzos")
	}

	js := get(t, h, "/app.js").Body.String()
	if !strings.Contains(js, `accion("msg-esfuerzos"`) {
		t.Error("app.js no muestra los errores de la consulta con accion")
	}
}

// HU-18 · Registrar la fecha de inicio del proyecto (Issue #36)
//
// Trazabilidad: CA-18.10. La sección Proyectos tiene un campo de fecha y el
// botón "Registrar fecha de inicio" para el proyecto actual, y los datos del
// proyecto muestran la fecha. El recorrido en el navegador queda en la
// evidencia TDD.

// seccionProyectos devuelve el HTML de la sección Proyectos.
func seccionProyectos(t *testing.T, page string) string {
	t.Helper()
	inicio := strings.Index(page, `<section id="proyectos">`)
	fin := strings.Index(page, `<section id="backlog">`)
	if inicio == -1 || fin < inicio {
		t.Fatal("la página no contiene la sección de proyectos")
	}
	return page[inicio:fin]
}

// HU-18 · CA-18.10 · Verifica que la sección Proyectos tenga el formulario
// "Registrar fecha de inicio" y que app.js lo envíe a /fecha-inicio.
func TestInterfaz_FormularioDeFechaDeInicioDelProyecto(t *testing.T) {
	h := newHandler()

	seccion := seccionProyectos(t, get(t, h, "/").Body.String())
	for _, want := range []string{
		`id="form-fecha-inicio-proyecto"`,
		`name="fecha_inicio" type="date" required`,
		"Registrar fecha de inicio",
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección de proyectos no contiene %q", want)
		}
	}

	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`$("form-fecha-inicio-proyecto")`, `"/fecha-inicio"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-18 · CA-18.10 · Verifica que, al registrar la fecha de inicio, la interfaz
// muestre el mensaje de éxito con la fecha.
func TestInterfaz_RegistrarFechaDeInicioMuestraElMensajeDeExito(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`mostrar("msg-proyectos", "Fecha de inicio del proyecto "`, `" registrada: "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-18 · CA-18.10 · Verifica que los datos del proyecto actual muestren la
// fecha de inicio o "sin fecha de inicio".
func TestInterfaz_DatosDelProyectoMuestranLaFechaDeInicio(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{"p.fecha_inicio", `"sin fecha de inicio"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-41 · Registrar la fecha de fin del proyecto (Issue #115)
//
// Trazabilidad: CA-41.13. La sección Proyectos tiene, debajo de "Fecha de
// inicio", un campo de fecha y el botón "Registrar fecha de fin" para el
// proyecto actual, y los datos del proyecto muestran la fecha de fin. El
// recorrido en el navegador queda en la evidencia TDD.

// HU-41 · CA-41.13 · Verifica que la sección Proyectos tenga el formulario
// "Registrar fecha de fin", debajo del de fecha de inicio, y que app.js lo
// envíe a /fecha-fin.
func TestInterfaz_FormularioDeFechaDeFinDelProyecto(t *testing.T) {
	h := newHandler()

	seccion := seccionProyectos(t, get(t, h, "/").Body.String())
	for _, want := range []string{
		`id="form-fecha-fin-proyecto"`,
		`name="fecha_fin" type="date" required`,
		"Registrar fecha de fin",
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección de proyectos no contiene %q", want)
		}
	}
	inicio := strings.Index(seccion, `id="form-fecha-inicio-proyecto"`)
	fin := strings.Index(seccion, `id="form-fecha-fin-proyecto"`)
	if inicio == -1 || fin < inicio {
		t.Error("el formulario de fecha de fin no está debajo de \"Fecha de inicio\"")
	}

	js := get(t, h, "/app.js").Body.String()
	for _, want := range []string{`$("form-fecha-fin-proyecto")`, `"/fecha-fin"`, "fecha_fin: form.fecha_fin.value"} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-41 · CA-41.13 · Verifica que, al registrar la fecha de fin, la interfaz
// muestre el mensaje de éxito con la fecha.
func TestInterfaz_RegistrarFechaDeFinMuestraElMensajeDeExito(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`mostrar("msg-proyectos", "Fecha de fin del proyecto "`, "p.fecha_fin + \".\""} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-41 · CA-41.13 · Verifica que los datos del proyecto actual muestren la
// fecha de fin o "sin fecha de fin".
func TestInterfaz_DatosDelProyectoMuestranLaFechaDeFin(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`$("form-fecha-fin-proyecto").fecha_fin.value = p.fecha_fin`, `"sin fecha de fin"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-19 · Modificar las fechas del proyecto (Issue #37)
//
// Trazabilidad: CA-19.8. La sección Proyectos tiene, debajo de "Fecha de
// fin", el formulario "Modificar fechas" con los dos campos de fecha para el
// proyecto actual; app.js llama a PATCH /proyectos/{id}/fechas, muestra el
// mensaje de éxito y precarga los dos campos en mostrarProyecto. El recorrido
// en el navegador queda en la evidencia TDD.

// formularioModificarFechas devuelve el formulario "Modificar fechas" de la
// sección Proyectos.
func formularioModificarFechas(t *testing.T, seccion string) string {
	t.Helper()
	inicio := strings.Index(seccion, `<form id="form-modificar-fechas-proyecto">`)
	if inicio == -1 {
		t.Fatal("la sección de proyectos no contiene el formulario \"Modificar fechas\"")
	}
	fin := strings.Index(seccion[inicio:], "</form>")
	if fin == -1 {
		t.Fatal("el formulario \"Modificar fechas\" no está cerrado")
	}
	return seccion[inicio : inicio+fin]
}

// manejadorModificarFechas devuelve el manejador del formulario "Modificar
// fechas" en app.js.
func manejadorModificarFechas(t *testing.T, js string) string {
	t.Helper()
	inicio := strings.Index(js, `$("form-modificar-fechas-proyecto").addEventListener("submit"`)
	if inicio == -1 {
		t.Fatal("app.js no maneja el envío del formulario \"Modificar fechas\"")
	}
	fin := strings.Index(js[inicio:], "\n});")
	if fin == -1 {
		t.Fatal("el manejador del formulario \"Modificar fechas\" no está cerrado")
	}
	return js[inicio : inicio+fin]
}

// HU-19 · CA-19.8 · Verifica que la sección Proyectos tenga el formulario
// "Modificar fechas", con las dos fechas, debajo del de fecha de fin.
func TestInterfaz_FormularioModificarFechasDelProyecto(t *testing.T) {
	seccion := seccionProyectos(t, get(t, newHandler(), "/").Body.String())

	form := formularioModificarFechas(t, seccion)
	for _, want := range []string{
		"<h3>Modificar fechas</h3>",
		`name="fecha_inicio" type="date" required`,
		`name="fecha_fin" type="date" required`,
		`<button type="submit">Modificar fechas</button>`,
	} {
		if !strings.Contains(form, want) {
			t.Errorf("el formulario \"Modificar fechas\" no contiene %q", want)
		}
	}
	fechaFin := strings.Index(seccion, `id="form-fecha-fin-proyecto"`)
	modificar := strings.Index(seccion, `id="form-modificar-fechas-proyecto"`)
	if fechaFin == -1 || modificar < fechaFin {
		t.Error("el formulario \"Modificar fechas\" no está debajo de \"Fecha de fin\"")
	}
}

// HU-19 · CA-19.8 · Verifica que "Modificar fechas" envíe el PATCH a
// /proyectos/{id}/fechas con el proyecto actual y muestre el proyecto
// actualizado.
func TestInterfaz_ModificarFechasLlamaALaAPIConElProyectoActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	manejador := manejadorModificarFechas(t, js)
	for _, want := range []string{
		`accion("msg-proyectos"`,
		`api("PATCH", "/proyectos/" + proyectoActual() + "/fechas"`,
		"fecha_inicio: form.fecha_inicio.value",
		"fecha_fin: form.fecha_fin.value",
		"mostrarProyecto(p)",
	} {
		if !strings.Contains(manejador, want) {
			t.Errorf("el manejador de \"Modificar fechas\" no contiene %q", want)
		}
	}
}

// HU-19 · CA-19.8 · Verifica que, al modificar las fechas, la interfaz muestre
// el mensaje de éxito con el inicio y el fin.
func TestInterfaz_ModificarFechasMuestraElMensajeDeExito(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	manejador := manejadorModificarFechas(t, js)
	want := `mostrar("msg-proyectos", "Fechas del proyecto " + p.id + " modificadas: inicio " + p.fecha_inicio + ", fin " + p.fecha_fin + ".", false)`
	if !strings.Contains(manejador, want) {
		t.Errorf("el manejador de \"Modificar fechas\" no contiene %q", want)
	}
}

// HU-19 · CA-19.8 · Verifica que mostrarProyecto precargue el formulario
// "Modificar fechas" con las fechas del proyecto.
func TestInterfaz_DatosDelProyectoPrecarganElFormularioModificarFechas(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	inicio := strings.Index(js, "function mostrarProyecto(p)")
	if inicio == -1 {
		t.Fatal("app.js no contiene mostrarProyecto")
	}
	fin := strings.Index(js[inicio:], "\n}")
	if fin == -1 {
		t.Fatal("mostrarProyecto no está cerrada")
	}
	funcion := js[inicio : inicio+fin]
	for _, want := range []string{
		`$("form-modificar-fechas-proyecto").fecha_inicio.value = p.fecha_inicio`,
		`$("form-modificar-fechas-proyecto").fecha_fin.value = p.fecha_fin`,
	} {
		if !strings.Contains(funcion, want) {
			t.Errorf("mostrarProyecto no contiene %q", want)
		}
	}
}

// HU-36 · Ver el resumen del estado del proyecto (Issue #54)
//
// Trazabilidad: CA-36.6. La interfaz tiene una sección "Dashboard" con el
// botón "Ver resumen", que llama a GET /proyectos/{id}/resumen con el
// proyecto actual y muestra el proyecto, los ítems por estado y el sprint
// actual, o "ninguno abierto".

// HU-36 · CA-36.6 · Verifica que la sección "Dashboard" tenga el botón "Ver
// resumen", el lugar de los datos y el mensaje.
func TestInterfaz_SeccionDashboardConElBotonVerResumen(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<section id="dashboard">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección del dashboard")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección del dashboard no se cierra")
	}
	seccion := page[inicio : inicio+fin]
	for _, want := range []string{
		"<h2>Dashboard</h2>",
		`id="ver-resumen"`,
		"Ver resumen",
		`id="datos-resumen"`,
		`id="msg-resumen"`,
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección del dashboard no contiene %q", want)
		}
	}
}

// HU-36 · CA-36.6 · Verifica que "Ver resumen" llame a /resumen con el
// proyecto actual.
func TestInterfaz_VerResumenLlamaALaAPIConElProyectoActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`$("ver-resumen")`, `accion("msg-resumen"`, `proyectoActual() + "/resumen"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-36 · CA-36.6 · Verifica que "Ver resumen" muestre los ítems por estado y
// el sprint actual, o "ninguno abierto".
func TestInterfaz_VerResumenMuestraLosItemsPorEstadoYElSprintActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		"items_pendientes",
		"items_en_curso",
		"items_terminados",
		"sprint_actual",
		"Sprint actual: ",
		"ninguno abierto",
		`$("datos-resumen")`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-47 · CA-47.6 · Verifica que la sección "Métricas" muestre los defectos
// detectados y resueltos del proyecto y del sprint actual, como enteros.
func TestInterfaz_MetricasMuestraLosDefectosDetectadosYResueltos(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		"defectos_detectados",
		"defectos_resueltos",
		`"Defectos: "`,
		`" detectados, "`,
		`" resuelto(s)."`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-37 · Ver las métricas en el dashboard con gráficos (Issue #55)
//
// CA-37.1, CA-37.2 y CA-37.6: la sección "Dashboard" tiene el botón "Ver
// gráficos", que muestra M1 a M10 del proyecto y tres gráficos SVG dibujados
// con JavaScript propio. Los tests no tienen un motor de JavaScript, así que
// verifican el contenido de index.html, app.js y styles.css; el dibujo se
// verifica en el navegador (evidencia TDD).

// handlerBody devuelve el código del handler del botón id en app.js, desde
// $("id").addEventListener hasta el cierre "\n);\n".
func handlerBody(t *testing.T, js, id string) string {
	t.Helper()
	start := strings.Index(js, `$("`+id+`").addEventListener`)
	if start == -1 {
		t.Fatalf("app.js no tiene el handler de %q", id)
	}
	end := strings.Index(js[start:], "\n);\n")
	if end == -1 {
		t.Fatalf("el handler de %q no se cierra", id)
	}
	return js[start : start+end]
}

// HU-37 · CA-37.1 · Verifica que la sección "Dashboard" tenga el botón "Ver
// gráficos", el texto de las métricas, el contenedor de los gráficos y el
// mensaje, y conserve el botón "Ver resumen" de HU-36.
func TestInterfaz_DashboardTieneElBotonVerGraficos(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<section id="dashboard">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección del dashboard")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección del dashboard no se cierra")
	}
	seccion := page[inicio : inicio+fin]
	for _, want := range []string{
		`id="ver-resumen"`,
		`id="ver-graficos"`,
		"Ver gráficos",
		`id="datos-graficos"`,
		`id="graficos"`,
		`id="msg-graficos"`,
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección del dashboard no contiene %q", want)
		}
	}
}

// HU-37 · CA-37.1 · Verifica que "Ver gráficos" pida las métricas del proyecto
// (HU-35) y las de cada sprint con el proyecto actual. También cubre CA-37.3.
func TestInterfaz_VerGraficosLlamaALasDosRutas(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()
	body := handlerBody(t, js, "ver-graficos")

	for _, want := range []string{
		`accion("msg-graficos"`,
		`proyectoActual() + "/metricas"`,
		`proyectoActual() + "/metricas/sprints"`,
		`$("datos-graficos")`,
		`$("graficos")`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("el handler de ver-graficos no contiene %q", want)
		}
	}
}

// HU-37 · CA-37.1 · Verifica que el texto de M1 a M3 esté en una función
// compartida, que usan "Ver métricas" y "Ver gráficos", junto con las de M4 a
// M10.
func TestInterfaz_TextoDeStoryPointsCompartidoPorLosDosBotones(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	if !strings.Contains(js, "function storyPointsYVelocidad(") {
		t.Fatal("app.js no tiene la función storyPointsYVelocidad")
	}
	for _, id := range []string{"ver-metricas", "ver-graficos"} {
		body := handlerBody(t, js, id)
		for _, want := range []string{"storyPointsYVelocidad(p)", "historiasCompletadas(p)", "horasYDesviacion(p)", "defectos(p)"} {
			if !strings.Contains(body, want) {
				t.Errorf("el handler de %s no usa %q", id, want)
			}
		}
	}
}

// HU-37 · CA-37.1 · Verifica que los tres gráficos tengan su título, la
// velocidad como línea de referencia y solo los sprints CERRADO en el gráfico
// de velocidad.
func TestInterfaz_GraficosTienenLosTresTitulos(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`"Story Points completados por sprint cerrado"`,
		`"Story Points planificados y completados por sprint"`,
		`"Horas estimadas y reales por sprint"`,
		`"Velocidad: "`,
		`"CERRADO"`,
		"sp_planificados",
		"sp_completados",
		"horas_estimadas",
		"horas_reales",
		"desviacion_horas",
		"desviacion_porcentual",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-37 · CA-37.2 · Verifica que una métrica no disponible se muestre como "sin
// datos", y que un proyecto sin sprints muestre el aviso en lugar de los
// gráficos.
func TestInterfaz_GraficosMuestranSinDatos(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`"Velocidad: sin datos (no hay sprints cerrados)."`,
		`"Sin sprints: no hay datos para los gráficos."`,
		`"Desviación: sin datos"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-37 · CA-37.6 · Verifica que los gráficos sean SVG creados con
// createElementNS, que los textos se asignen con textContent y que no se use
// innerHTML ni recursos externos.
func TestInterfaz_GraficosSeDibujanConSVGPropioSinLibrerias(t *testing.T) {
	h := newHandler()
	js := get(t, h, "/app.js").Body.String()
	page := get(t, h, "/").Body.String()

	for _, want := range []string{
		"document.createElementNS",
		`"http://www.w3.org/2000/svg"`,
		`"title"`,
		"textContent",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
	if strings.Contains(js, "innerHTML") {
		t.Error("app.js usa innerHTML")
	}
	for _, external := range []string{`src="http`, `src="//`, `href="http`, `href="//`} {
		if strings.Contains(page, external) {
			t.Errorf("index.html carga un recurso externo (%s)", external)
		}
	}
}

// HU-37 · CA-37.1 · Verifica que los colores de las series salgan de variables
// CSS de styles.css.
func TestInterfaz_GraficosUsanVariablesDeColor(t *testing.T) {
	css := get(t, newHandler(), "/styles.css").Body.String()

	for _, want := range []string{"--serie-a:", "--serie-b:", "--referencia:", ".grafico"} {
		if !strings.Contains(css, want) {
			t.Errorf("styles.css no contiene %q", want)
		}
	}
}

// HU-37 · CA-37.1 · Verifica la corrección de un defecto: el ancho de cada grupo
// de barras tiene en cuenta el nombre del sprint y la nota de desviación, así
// un texto más largo que el grupo, como "Desviación: -2.5 h (-19.2 %)", no queda
// cortado en el borde del gráfico ni se pisa con el del sprint siguiente. El
// dibujo se verifica en el navegador (evidencia TDD de HU-37).
func TestInterfaz_GraficoDeBarrasEnsanchaLosGruposSegunSusTextos(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	if !strings.Contains(js, "function anchoDeTexto(") {
		t.Fatal("app.js no tiene la función anchoDeTexto")
	}
	start := strings.Index(js, "function graficoDeBarras(")
	if start == -1 {
		t.Fatal("app.js no tiene la función graficoDeBarras")
	}
	end := strings.Index(js[start:], "\n}\n")
	if end == -1 {
		t.Fatal("la función graficoDeBarras no se cierra")
	}
	body := js[start : start+end]
	for _, want := range []string{"anchoDeTexto(g.nombre)", "anchoDeTexto(g.nota)"} {
		if !strings.Contains(body, want) {
			t.Errorf("graficoDeBarras no calcula el ancho del grupo con %q", want)
		}
	}
}

// HU-38 · Generar el reporte del proyecto (Issue #56)
//
// Trazabilidad: CA-38.7. La interfaz tiene una sección "Reportes" con el
// botón "Generar reporte", que llama a GET /proyectos/{id}/reporte con el
// proyecto actual y muestra cada parte del reporte en su bloque.

// HU-38 · CA-38.7 · Verifica que la sección "Reportes" tenga el botón "Generar
// reporte" y un lugar para cada parte del reporte del proyecto.
func TestInterfaz_SeccionReportesConElBotonGenerarReporte(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<section id="reportes">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección de reportes")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección de reportes no se cierra")
	}
	seccion := page[inicio : inicio+fin]
	for _, want := range []string{
		"<h2>Reportes</h2>",
		`id="generar-reporte"`,
		"Generar reporte",
		`id="datos-reporte"`,
		`id="msg-reporte"`,
		`id="reporte-proyecto"`,
		`id="reporte-metricas"`,
		`id="reporte-esfuerzos"`,
		`id="reporte-defectos"`,
		`id="reporte-integrantes"`,
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección de reportes no contiene %q", want)
		}
	}
}

// HU-38 · CA-38.7 · Verifica que "Generar reporte" llame a /reporte con el
// proyecto actual.
func TestInterfaz_GenerarReporteLlamaALaAPIConElProyectoActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`$("generar-reporte")`, `accion("msg-reporte"`, `proyectoActual() + "/reporte"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-38 · CA-38.7 · Verifica que el reporte del proyecto muestre cada parte, con
// los textos de las secciones vacías. También cubre CA-38.2.
func TestInterfaz_GenerarReporteMuestraCadaParte(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`$("reporte-proyecto")`,
		`$("reporte-metricas")`,
		`$("reporte-esfuerzos")`,
		`$("reporte-defectos")`,
		`$("reporte-integrantes")`,
		"Sin esfuerzos registrados.",
		"Sin defectos registrados.",
		"Integrantes que registraron esfuerzo: ",
		"Inicio: sin registrar",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-39 · Generar el reporte de un sprint (Issue #57)
//
// Trazabilidad: CA-39.8. La sección "Reportes" tiene el botón "Generar
// reporte del sprint", que llama a GET /sprints/{id}/reporte con el sprint
// actual y muestra cada parte del reporte en su propio contenedor, sin tocar
// los ids del reporte del proyecto (HU-38).

// HU-39 · CA-39.8 · Verifica que la sección "Reportes" tenga el botón "Generar
// reporte del sprint" y un lugar para cada parte del reporte del sprint.
func TestInterfaz_SeccionReportesConElBotonGenerarReporteDelSprint(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<section id="reportes">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección de reportes")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección de reportes no se cierra")
	}
	seccion := page[inicio : inicio+fin]
	for _, want := range []string{
		`id="generar-reporte-sprint"`,
		"Generar reporte del sprint",
		`id="datos-reporte-sprint"`,
		`id="msg-reporte-sprint"`,
		`id="reporte-sprint"`,
		`id="reporte-sprint-metricas"`,
		`id="reporte-sprint-esfuerzos"`,
		`id="reporte-sprint-total-horas"`,
		`id="reporte-sprint-defectos"`,
		`id="reporte-sprint-cantidad-defectos"`,
		`id="reporte-sprint-integrantes"`,
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección de reportes no contiene %q", want)
		}
	}
}

// HU-39 · CA-39.8 · Verifica que "Generar reporte del sprint" llame a /reporte
// con el sprint actual.
func TestInterfaz_GenerarReporteDelSprintLlamaALaAPIConElSprintActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`$("generar-reporte-sprint")`, `accion("msg-reporte-sprint"`, `sprintActual() + "/reporte"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-39 · CA-39.8 · Verifica que el reporte del sprint muestre cada parte,
// incluido el objetivo del sprint.
func TestInterfaz_GenerarReporteDelSprintMuestraCadaParte(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`$("reporte-sprint")`,
		`$("reporte-sprint-metricas")`,
		`$("reporte-sprint-esfuerzos")`,
		`$("reporte-sprint-total-horas")`,
		`$("reporte-sprint-defectos")`,
		`$("reporte-sprint-cantidad-defectos")`,
		`$("reporte-sprint-integrantes")`,
		"Objetivo: ",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-49 · Ver las historias y sus estimaciones en el reporte (Issue #170)
//
// Trazabilidad: CA-49.7. En los dos bloques de la sección "Reportes", después
// del texto de las métricas y antes de la tabla de esfuerzos, hay una tabla de
// historias con las columnas ID, Título, Prioridad, Estado, SP, Horas
// estimadas y Sprint, y debajo la cantidad de historias y de completadas.

// seccionReportes devuelve el HTML de la sección "Reportes".
func seccionReportes(t *testing.T) string {
	t.Helper()
	page := get(t, newHandler(), "/").Body.String()
	inicio := strings.Index(page, `<section id="reportes">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección de reportes")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección de reportes no se cierra")
	}
	return page[inicio : inicio+fin]
}

// HU-49 · CA-49.7 · Verifica que los reportes del proyecto y del sprint tengan la
// tabla de historias, con sus columnas y su cantidad, entre las métricas y los
// esfuerzos.
func TestInterfaz_ReportesTienenLaTablaDeHistorias(t *testing.T) {
	seccion := seccionReportes(t)

	const encabezado = "<th>ID</th><th>Título</th><th>Prioridad</th><th>Estado</th><th>SP</th><th>Horas estimadas</th><th>Sprint</th>"
	for _, bloque := range []struct{ metricas, historias, cantidad, esfuerzos string }{
		{`id="reporte-metricas"`, `id="reporte-historias"`, `id="reporte-cantidad-historias"`, `id="reporte-esfuerzos"`},
		{`id="reporte-sprint-metricas"`, `id="reporte-sprint-historias"`, `id="reporte-sprint-cantidad-historias"`, `id="reporte-sprint-esfuerzos"`},
	} {
		metricas := strings.Index(seccion, bloque.metricas)
		historias := strings.Index(seccion, bloque.historias)
		cantidad := strings.Index(seccion, bloque.cantidad)
		esfuerzos := strings.Index(seccion, bloque.esfuerzos)
		if metricas == -1 || historias == -1 || cantidad == -1 || esfuerzos == -1 {
			t.Errorf("la sección de reportes no contiene %s, %s y %s", bloque.historias, bloque.cantidad, bloque.esfuerzos)
			continue
		}
		if !(metricas < historias && historias < cantidad && cantidad < esfuerzos) {
			t.Errorf("%s y %s no están entre %s y %s", bloque.historias, bloque.cantidad, bloque.metricas, bloque.esfuerzos)
		}
		if !strings.Contains(seccion[metricas:historias], encabezado) {
			t.Errorf("la tabla de %s no tiene las columnas %q", bloque.historias, encabezado)
		}
	}
}

// HU-49 · CA-49.7 · Verifica que los dos reportes muestren las historias y las
// cantidades, o el aviso de que no hay historias.
func TestInterfaz_GenerarLosReportesMuestraLasHistorias(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`$("reporte-historias")`,
		`$("reporte-cantidad-historias")`,
		`$("reporte-sprint-historias")`,
		`$("reporte-sprint-cantidad-historias")`,
		"r.historias.historias",
		"r.historias.cantidad",
		"r.historias.completadas",
		" historia(s), ",
		" completada(s)",
		"Sin historias en el backlog.",
		"Sin historias en el sprint.",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-49 · CA-49.7 · Verifica que la fila de una historia del reporte muestre "—"
// en SP si no está estimada y en Sprint si no está asignada, como la tabla del
// backlog, y las horas estimadas. También cubre CA-49.4.
func TestInterfaz_FilaDeHistoriaDelReporteIndicaSinEstimarYSinSprint(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	inicio := strings.Index(js, "function filaHistoriaReporte(")
	if inicio == -1 {
		t.Fatal("app.js no tiene la función filaHistoriaReporte")
	}
	fin := strings.Index(js[inicio:], "\n}\n")
	if fin == -1 {
		t.Fatal("la función filaHistoriaReporte no se cierra")
	}
	cuerpo := js[inicio : inicio+fin]
	for _, want := range []string{
		`h.story_points === 0 ? "—" : h.story_points`,
		"h.horas_estimadas",
		`h.sprint_id === 0 ? "—" : h.sprint_id`,
	} {
		if !strings.Contains(cuerpo, want) {
			t.Errorf("filaHistoriaReporte no contiene %q", want)
		}
	}
	for _, want := range []string{
		`filasReporte($("reporte-historias"), r.historias.historias, filaHistoriaReporte`,
		`filasReporte($("reporte-sprint-historias"), r.historias.historias, filaHistoriaReporte`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-40 · Exportar un reporte a un archivo PDF (Issue #58)
//
// Trazabilidad: CA-40.1. La sección "Reportes" tiene el botón "Descargar PDF",
// que pide GET /proyectos/{id}/reporte/pdf con el proyecto actual y descarga
// el archivo, sin tocar los ids del reporte del proyecto (HU-38).

// HU-40 · CA-40.1 · Verifica que la sección "Reportes" tenga el botón
// "Descargar PDF", junto a "Generar reporte".
func TestInterfaz_SeccionReportesConElBotonDescargarPDF(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<section id="reportes">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección de reportes")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección de reportes no se cierra")
	}
	seccion := page[inicio : inicio+fin]
	for _, want := range []string{
		`<button type="button" id="descargar-reporte-pdf">Descargar PDF</button>`,
		`id="generar-reporte"`,
		`id="msg-reporte"`,
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección de reportes no contiene %q", want)
		}
	}
}

// HU-40 · CA-40.1 · Verifica que "Descargar PDF" llame a /reporte/pdf con el
// proyecto actual y descargue el archivo reporte-proyecto-N.pdf.
func TestInterfaz_DescargarPDFLlamaALaAPIConElProyectoActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`$("descargar-reporte-pdf")`,
		`accion("msg-reporte"`,
		`proyectoActual()`,
		`"/reporte/pdf"`,
		`"reporte-proyecto-" + `,
		`.pdf"`,
		"PDF del proyecto ",
		// El enlace temporal se libera después del clic, no en el mismo
		// paso: liberarlo antes puede cancelar la descarga.
		"setTimeout(() => URL.revokeObjectURL(url), 0)",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-14 · Consultar sprints anteriores (Issue #28)
//
// CA-14.6: la sección "Sprints" tiene el botón "Ver sprints del proyecto", que
// llena una tabla con los sprints del proyecto actual; cada fila tiene el
// botón "Ver ítems completados", que deja ese sprint como sprint actual y
// avisa el cambio en el mensaje (ajuste del Agile Enabler en la SDD).

// HU-14 · CA-14.6 · Verifica que la sección "Sprints" tenga el botón y la tabla
// de sprints, con sus columnas, y conserve el botón "Ver ítems completados" de
// HU-12.
func TestInterfaz_SeccionSprintsTieneElBotonVerSprintsDelProyecto(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<section id="sprints">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección de sprints")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección de sprints no se cierra")
	}
	seccion := page[inicio : inicio+fin]
	for _, want := range []string{
		`<button type="button" id="ver-sprints" class="secundario">Ver sprints del proyecto</button>`,
		`<tbody id="lista-sprints"></tbody>`,
		"<th>ID</th><th>Nombre</th><th>Inicio</th><th>Fin</th><th>Sprint goal</th><th>Estado</th><th>Acciones</th>",
		`id="ver-completados"`,
		`id="msg-sprints"`,
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección de sprints no contiene %q", want)
		}
	}
}

// HU-14 · CA-14.6 · Verifica que "Ver sprints del proyecto" llame a la ruta
// nueva con el proyecto actual, llene la tabla e informe la cantidad. También
// cubre CA-14.1 y CA-14.2.
func TestInterfaz_VerSprintsLlamaALaAPIConElProyectoActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()
	body := handlerBody(t, js, "ver-sprints")

	for _, want := range []string{
		`accion("msg-sprints"`,
		`api("GET", "/proyectos/" + proyecto + "/sprints")`,
		`$("lista-sprints")`,
		"s.fecha_inicio",
		"s.fecha_fin",
		"s.objetivo",
		"s.estado",
		`" sprint(s) en el proyecto "`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("el handler de ver-sprints no contiene %q", want)
		}
	}
	if strings.Contains(body, "innerHTML") {
		t.Error("el handler de ver-sprints usa innerHTML")
	}
}

// HU-14 · CA-14.3 · Verifica que el botón de cada fila deje el sprint como sprint
// actual (mostrarSprint), muestre sus ítems completados con la consulta de
// HU-12 y avise "Sprint N seleccionado: M ítem(s) completado(s).". También
// cubre CA-14.6.
func TestInterfaz_FilaDeSprintSeleccionaElSprintYMuestraSusCompletados(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()
	body := handlerBody(t, js, "ver-sprints")

	for _, want := range []string{
		`"Ver ítems completados"`,
		"mostrarSprint(s)",
		"mostrarCompletados(s.id)",
		`"Sprint " + s.id + " seleccionado: "`,
		`" ítem(s) completado(s)."`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("el handler de ver-sprints no contiene %q", want)
		}
	}
}

// HU-14 · CA-14.3 · Verifica que la consulta de los ítems completados sea una
// sola función, que usan el botón de HU-12 y el de cada fila.
func TestInterfaz_CompletadosCompartidosPorLosDosBotones(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	if !strings.Contains(js, "async function mostrarCompletados(") {
		t.Fatal("app.js no tiene la función mostrarCompletados")
	}
	if !strings.Contains(js, `api("GET", "/sprints/" + sprint + "/completados")`) {
		t.Error("app.js no consulta los ítems completados de HU-12")
	}
	if !strings.Contains(handlerBody(t, js, "ver-completados"), "mostrarCompletados(sprint)") {
		t.Error("el handler de ver-completados no usa mostrarCompletados")
	}
}

// HU-15 · Registrar integrante del proyecto (Issue #29)
//
// Trazabilidad: CA-15.7. La sección "Integrantes" tiene el formulario
// "Registrar integrante", que llama a POST /proyectos/{id}/integrantes con el
// proyecto actual.

// HU-15 · CA-15.7 · Verifica que la sección "Integrantes" tenga el formulario
// "Registrar integrante", con el nombre y los tres roles.
func TestInterfaz_SeccionIntegrantesConElFormularioDeRegistro(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<section id="integrantes">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección de integrantes")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección de integrantes no se cierra")
	}
	seccion := page[inicio : inicio+fin]
	for _, want := range []string{
		"<h2>Integrantes</h2>",
		`id="form-registrar-integrante"`,
		`name="nombre"`,
		`name="rol"`,
		`<option value="PRODUCT_ARCHITECT">Product Architect</option>`,
		`<option value="AGILE_ENABLER">Agile Enabler</option>`,
		`<option value="PRODUCT_BUILDER">Product Builder</option>`,
		"Registrar integrante",
		`id="msg-integrantes"`,
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección de integrantes no contiene %q", want)
		}
	}
}

// HU-15 · CA-15.7 · Verifica que "Registrar integrante" llame a /integrantes
// con el proyecto actual y muestre el integrante registrado.
func TestInterfaz_RegistrarIntegranteLlamaALaAPIConElProyectoActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{
		`$("form-registrar-integrante")`,
		`accion("msg-integrantes"`,
		`proyectoActual() + "/integrantes"`,
		`"Integrante " + `,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-28 · Corregir un registro de esfuerzo (Issue #46)
//
// Trazabilidad: CA-28.7. La sección "Sprints" tiene el formulario "Corregir
// esfuerzo", con el ID del esfuerzo y las horas, que llama a
// PATCH /esfuerzos/{id}/horas.

// HU-28 · CA-28.7 · Verifica que el formulario "Corregir esfuerzo" esté en la
// sección Sprints, entre "Registrar esfuerzo" y "Consultar esfuerzos".
func TestInterfaz_FormularioCorregirEsfuerzoEnSprints(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<form id="form-corregir-esfuerzo">`)
	if inicio == -1 {
		t.Fatal("la página no contiene el formulario para corregir esfuerzos")
	}
	fin := strings.Index(page[inicio:], "</form>")
	if fin == -1 {
		t.Fatal("el formulario para corregir esfuerzos no se cierra")
	}
	form := page[inicio : inicio+fin]
	for _, want := range []string{"<h3>Corregir esfuerzo</h3>", `name="esfuerzo_id"`, `name="horas"`, "Corregir esfuerzo</button>"} {
		if !strings.Contains(form, want) {
			t.Errorf("el formulario para corregir esfuerzos no contiene %q", want)
		}
	}
	sprints := strings.Index(page, `<section id="sprints">`)
	registrar := strings.Index(page, `<form id="form-registrar-esfuerzo">`)
	consultar := strings.Index(page, `<form id="form-consultar-esfuerzos">`)
	if !(sprints < registrar && registrar < inicio && inicio < consultar) {
		t.Error("el formulario para corregir esfuerzos tiene que estar en Sprints, entre Registrar esfuerzo y Consultar esfuerzos")
	}
}

// HU-28 · CA-28.7 · Verifica que "Corregir esfuerzo" envíe el PATCH a
// /esfuerzos/{id}/horas y muestre el esfuerzo corregido.
func TestInterfaz_CorregirEsfuerzoLlamaALaAPI(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`$("form-corregir-esfuerzo")`, `"PATCH"`, `"/esfuerzos/"`, `"/horas"`, " corregido: "} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-17 · Consultar los integrantes de un proyecto (Issue #31)
//
// Trazabilidad: CA-17.6. La sección "Integrantes" tiene el botón "Ver
// integrantes del proyecto", que llena una tabla con el ID, el nombre y el
// rol legible de los integrantes del proyecto actual.

// HU-17 · CA-17.6 · Verifica que el botón y la tabla estén en la sección
// "Integrantes", después del formulario "Modificar integrante" y antes del
// mensaje.
func TestInterfaz_SeccionIntegrantesTieneElBotonVerIntegrantes(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	inicio := strings.Index(page, `<section id="integrantes">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección de integrantes")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección de integrantes no se cierra")
	}
	seccion := page[inicio : inicio+fin]
	for _, want := range []string{
		`<button type="button" id="ver-integrantes" class="secundario">Ver integrantes del proyecto</button>`,
		`<tbody id="lista-integrantes"></tbody>`,
		"<th>ID</th><th>Nombre</th><th>Rol</th>",
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección de integrantes no contiene %q", want)
		}
	}
	modificar := strings.Index(seccion, `id="form-modificar-integrante"`)
	boton := strings.Index(seccion, `id="ver-integrantes"`)
	mensaje := strings.Index(seccion, `id="msg-integrantes"`)
	if !(modificar < boton && boton < mensaje) {
		t.Error("el botón Ver integrantes tiene que estar entre Modificar integrante y el mensaje")
	}
}

// HU-17 · CA-17.6 · Verifica que "Ver integrantes del proyecto" llame a la ruta
// con el proyecto actual, llene la tabla con el rol legible e informe la
// cantidad. También cubre CA-17.1 y CA-17.2.
func TestInterfaz_VerIntegrantesLlamaALaAPIConElProyectoActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()
	body := handlerBody(t, js, "ver-integrantes")

	for _, want := range []string{
		`accion("msg-integrantes"`,
		`api("GET", "/proyectos/" + proyecto + "/integrantes")`,
		`$("lista-integrantes")`,
		"m.id",
		"m.nombre",
		"nombresDeRoles[m.rol]",
		`" integrante(s) en el proyecto "`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("el handler de ver-integrantes no contiene %q", want)
		}
	}
	if strings.Contains(body, "innerHTML") {
		t.Error("el handler de ver-integrantes usa innerHTML")
	}
}
