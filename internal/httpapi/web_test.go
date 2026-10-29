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

// CA-TEC04.8 · La interfaz web responde en /.
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

// CA-TEC04.4 · El JavaScript y los estilos de la interfaz se sirven desde el
// binario, con su tipo de contenido.
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

// HU-22 · Votar de forma oculta (Issue #40)
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

// HU-23, CA-23.8 · El botón "Revelar votos" revela la sesión y la interfaz
// muestra la lista de votos.
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

// HU-25, CA-25.9 · El formulario "Registrar estimación acordada" registra la
// carta elegida, muestra el mensaje y recarga el backlog.
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

// HU-24, CA-24.12 · Después de revelar se muestra la comparación y el botón
// "Repetir ronda", dentro del bloque de la estimación. El botón es
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

// HU-32, CA-32.9 · La sección "Métricas" muestra los Story Points del
// proyecto y del sprint actual, y la velocidad.
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

// HU-34, CA-34.6 · La sección "Métricas" muestra el porcentaje de historias
// completadas con un decimal y las cantidades, o "sin datos" si no hay ítems.
func TestInterfaz_MetricasMuestraElPorcentajeDeHistoriasCompletadas(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{"porcentaje_completadas", "items_terminados", "items_totales", "Historias completadas", "sin datos"} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-33, CA-33.5 y CA-33.9 · La sección "Métricas" muestra M4 a M7 y la
// cantidad de ítems sin horas estimadas, con un decimal y "sin datos" cuando
// la desviación no está disponible.
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

// HU-33, CA-33.9 · numero1 convierte el texto "-0.0" en "0.0".
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

func TestInterfaz_RegistrarDefectoMuestraElMensajeDeExito(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`mostrar("msg-defectos", "Defecto "`, `" registrado para el ítem "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

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

// CA-29.11 y SDD de HU-29 ("Interfaz web"): la lista de severidades de app.js
// usa los mismos valores que defect.SeverityValues(), para que la interfaz no
// se desincronice de la validación del paquete.
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

func TestInterfaz_CambiarEstadoDeDefectoMuestraElMensajeDeExito(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`" ahora está en "`, `" resuelto en el sprint "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

func TestInterfaz_CambiarEstadoDeDefectoMuestraElMensajeDeError(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	// Registrar y cambiar el estado muestran sus errores en msg-defectos.
	if n := strings.Count(js, `accion("msg-defectos"`); n < 2 {
		t.Errorf("app.js muestra con accion los errores de %d formulario(s) de defectos, se esperaban 2", n)
	}
}

// CA-30.9 y SDD de HU-30 ("Interfaz web"): la lista de estados de app.js usa
// los mismos valores que defect.StatusValues().
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

// SDD de HU-31 ("Interfaz web"): el filtro de estado usa la lista de HU-30,
// que ya se compara con defect.StatusValues(), y no repite los estados.
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

func TestInterfaz_RegistrarFechaDeInicioMuestraElMensajeDeExito(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`mostrar("msg-proyectos", "Fecha de inicio del proyecto "`, `" registrada: "`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

func TestInterfaz_DatosDelProyectoMuestranLaFechaDeInicio(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{"p.fecha_inicio", `"sin fecha de inicio"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

// HU-36 · Ver el resumen del estado del proyecto (Issue #54)
//
// Trazabilidad: CA-36.6. La interfaz tiene una sección "Dashboard" con el
// botón "Ver resumen", que llama a GET /proyectos/{id}/resumen con el
// proyecto actual y muestra el proyecto, los ítems por estado y el sprint
// actual, o "ninguno abierto".

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

func TestInterfaz_VerResumenLlamaALaAPIConElProyectoActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`$("ver-resumen")`, `accion("msg-resumen"`, `proyectoActual() + "/resumen"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

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

// HU-47, CA-47.6 · La sección "Métricas" muestra los defectos detectados y
// resueltos del proyecto y del sprint actual, como enteros.
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

// CA-37.1 · La sección "Dashboard" tiene el botón "Ver gráficos", el texto de
// las métricas, el contenedor de los gráficos y el mensaje, y conserva el
// botón "Ver resumen" de HU-36.
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

// CA-37.1 y CA-37.3 · "Ver gráficos" pide las métricas del proyecto (HU-35) y
// las de cada sprint con el proyecto actual.
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

// CA-37.1 · El texto de M1 a M3 está en una función compartida, que usan
// "Ver métricas" y "Ver gráficos", junto con las de M4 a M10.
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

// CA-37.1 · Los tres gráficos tienen su título, la velocidad como línea de
// referencia y solo los sprints CERRADO en el gráfico de velocidad.
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

// CA-37.2 · Una métrica no disponible se muestra como "sin datos", y un
// proyecto sin sprints muestra el aviso en lugar de los gráficos.
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

// CA-37.6 · Los gráficos son SVG creados con createElementNS, los textos se
// asignan con textContent y no se usa innerHTML ni recursos externos.
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

// CA-37.1 · Los colores de las series salen de variables CSS de styles.css.
func TestInterfaz_GraficosUsanVariablesDeColor(t *testing.T) {
	css := get(t, newHandler(), "/styles.css").Body.String()

	for _, want := range []string{"--serie-a:", "--serie-b:", "--referencia:", ".grafico"} {
		if !strings.Contains(css, want) {
			t.Errorf("styles.css no contiene %q", want)
		}
	}
}

// HU-38 · Generar el reporte del proyecto (Issue #56)
//
// Trazabilidad: CA-38.7. La interfaz tiene una sección "Reportes" con el
// botón "Generar reporte", que llama a GET /proyectos/{id}/reporte con el
// proyecto actual y muestra cada parte del reporte en su bloque.

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

func TestInterfaz_GenerarReporteLlamaALaAPIConElProyectoActual(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()

	for _, want := range []string{`$("generar-reporte")`, `accion("msg-reporte"`, `proyectoActual() + "/reporte"`} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js no contiene %q", want)
		}
	}
}

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
