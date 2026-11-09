package httpapi_test

import (
	"strings"
	"testing"
)

// HU-50 · Listar los proyectos (Issue #183): interfaz web.
//
// Los tests no tienen un motor de JavaScript, así que verifican el contenido de
// index.html y app.js; el uso en el navegador se registra en la evidencia TDD.

// contextSection devuelve la sección «Proyecto y sprint actuales» de la página.
func contextSection(t *testing.T) string {
	t.Helper()
	page := get(t, newHandler(), "/").Body.String()
	inicio := strings.Index(page, `<section class="contexto" aria-label="Proyecto y sprint actuales">`)
	if inicio == -1 {
		t.Fatal("la página no contiene la sección de proyecto y sprint actuales")
	}
	fin := strings.Index(page[inicio:], "</section>")
	if fin == -1 {
		t.Fatal("la sección de proyecto y sprint actuales no se cierra")
	}
	return page[inicio : inicio+fin]
}

// HU-50 · CA-50.3 · Verifica que la sección de proyecto y sprint actuales
// tiene el botón «Ver proyectos» y la tabla de proyectos con sus columnas, y
// conserva el campo «Proyecto actual (ID)».
func TestInterfaz_ContextoTieneElBotonVerProyectosYLaTabla(t *testing.T) {
	seccion := contextSection(t)

	for _, want := range []string{
		`<button type="button" id="ver-proyectos" class="secundario">Ver proyectos</button>`,
		"<th>ID</th><th>Nombre</th><th>Descripción</th><th>Estado</th><th>Inicio</th><th>Fin</th><th>Acciones</th>",
		`<tbody id="lista-proyectos"></tbody>`,
		`id="proyecto-actual"`,
		`id="msg-contexto"`,
	} {
		if !strings.Contains(seccion, want) {
			t.Errorf("la sección de contexto no contiene %q", want)
		}
	}
}

// HU-50 · CA-50.3 · Verifica que la página ya no dice que no hay listado de
// proyectos.
func TestInterfaz_ComentarioDelContextoCitaHU50(t *testing.T) {
	page := get(t, newHandler(), "/").Body.String()

	if strings.Contains(page, "No hay listado de proyectos") {
		t.Error("la página todavía dice que no hay listado de proyectos")
	}
	if !strings.Contains(page, "HU-50") {
		t.Error("la página no cita HU-50 en el comentario de la sección de contexto")
	}
}

// HU-50 · CA-50.3 y CA-50.5 · Verifica que «Ver proyectos» consulta GET
// /api/proyectos, llena la tabla con los datos de cada proyecto y un botón
// «Elegir» que lo deja como proyecto actual con cargarProyecto, informa la
// cantidad y muestra los errores en el mensaje del contexto, sin innerHTML.
func TestInterfaz_VerProyectosLlenaLaTablaYPermiteElegir(t *testing.T) {
	js := get(t, newHandler(), "/app.js").Body.String()
	body := handlerBody(t, js, "ver-proyectos")

	for _, want := range []string{
		`accion("msg-contexto"`,
		`api("GET", "/proyectos")`,
		`$("lista-proyectos").replaceChildren(`,
		"p.nombre",
		"p.descripcion",
		"p.estado",
		"p.fecha_inicio",
		"p.fecha_fin",
		`"Elegir"`,
		`$("proyecto-actual").value = p.id`,
		"await cargarProyecto()",
		`"msg-contexto"`,
		`" proyecto(s)."`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("el handler de ver-proyectos no contiene %q", want)
		}
	}
	if strings.Contains(body, "innerHTML") {
		t.Error("el handler de ver-proyectos usa innerHTML")
	}
}
