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
