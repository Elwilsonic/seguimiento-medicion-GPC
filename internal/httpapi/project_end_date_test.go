package httpapi_test

import (
	"net/http"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-41 · Registrar la fecha de fin del proyecto (Issue #115)
//
// Trazabilidad: CA-41.2 a CA-41.9 y CA-41.13. PATCH
// /api/proyectos/{id}/fecha-fin, con un caso exitoso y un test por código de
// error; el 409 nuevo de PATCH /api/proyectos/{id}/fecha-inicio; fecha_fin en
// las respuestas de proyectos y en el reporte.

// fechaFin arma el cuerpo de PATCH /api/proyectos/{id}/fecha-fin.
func fechaFin(date string) string {
	return `{"fecha_fin":"` + date + `"}`
}

// newHandlerWithStartDate arma la API con el proyecto 1 "Sistema de ventas"
// con la fecha de inicio 2026-10-23.
func newHandlerWithStartDate(t *testing.T) http.Handler {
	t.Helper()
	h := newHandler()
	createProject(t, h, "Sistema de ventas")
	expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-10-23")), http.StatusOK)
	return h
}

// expectProjectEndDate verifica la fecha de fin que devuelve la consulta del
// proyecto.
func expectProjectEndDate(t *testing.T, h http.Handler, id, want string) {
	t.Helper()
	rec := do(t, h, http.MethodGet, "/api/proyectos/"+id, "")
	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	if p.FechaFin != want {
		t.Errorf("fecha_fin consultada = %q, se esperaba %q", p.FechaFin, want)
	}
}

// expectHasFechaFin verifica que la respuesta tenga el campo fecha_fin con el
// valor want.
func expectHasFechaFin(t *testing.T, op string, body map[string]any, want string) {
	t.Helper()
	if got, ok := body["fecha_fin"]; !ok || got != want {
		t.Errorf("%s: fecha_fin = %v (presente: %t), se esperaba %q", op, got, ok, want)
	}
}

// CA-41.1 y CA-41.13
func TestRegistrarFechaFinProyecto_Devuelve200ConElProyecto(t *testing.T) {
	h := newHandlerWithStartDate(t)
	do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas","estado":"EN_CURSO"}`)

	rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18"))

	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	want := projectJSON{ID: 1, Nombre: "Sistema de ventas", Descripcion: "Gestión de ventas", Estado: "EN_CURSO", FechaInicio: "2026-10-23", FechaFin: "2026-12-18"}
	if p != want {
		t.Errorf("proyecto = %+v, se esperaba %+v", p, want)
	}
	expectProjectEndDate(t, h, "1", "2026-12-18")
}

// CA-41.7
func TestRegistrarFechaFinProyecto_VolverARegistrarReemplazaLaFecha(t *testing.T) {
	h := newHandlerWithStartDate(t)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18")), http.StatusOK)

	rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2027-01-29"))

	expectStatus(t, rec, http.StatusOK)
	expectProjectEndDate(t, h, "1", "2027-01-29")
}

// CA-41.4
func TestRegistrarFechaFinProyecto_FechaAusenteDevuelve400(t *testing.T) {
	bodies := map[string]string{
		"vacía":   `{"fecha_fin":""}`,
		"null":    `{"fecha_fin":null}`,
		"ausente": `{}`,
	}
	for caseName, body := range bodies {
		t.Run(caseName, func(t *testing.T) {
			h := newHandlerWithStartDate(t)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", body)

			expectError(t, rec, http.StatusBadRequest, project.ErrMissingEndDate.Error())
			expectProjectEndDate(t, h, "1", "")
		})
	}
}

// CA-41.3
func TestRegistrarFechaFinProyecto_FormatoOFechaInvalidaDevuelve400(t *testing.T) {
	for _, date := range []string{"2026-02-30", "2026-13-01", "18/12/2026", "2026-1-5", "2026-12-18T00:00:00Z", " 2026-12-18 "} {
		t.Run(date, func(t *testing.T) {
			h := newHandlerWithStartDate(t)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin(date))

			expectError(t, rec, http.StatusBadRequest, "las fechas deben tener el formato AAAA-MM-DD")
			expectProjectEndDate(t, h, "1", "")
		})
	}
}

// CA-41.13: un cuerpo que no es JSON o una fecha que no es texto es JSON
// inválido.
func TestRegistrarFechaFinProyecto_JSONInvalidoDevuelve400(t *testing.T) {
	for _, body := range []string{`no es json`, `{"fecha_fin":20261218}`} {
		t.Run(body, func(t *testing.T) {
			h := newHandlerWithStartDate(t)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", body)

			expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
		})
	}
}

// CA-41.5
func TestRegistrarFechaFinProyecto_InexistenteDevuelve404(t *testing.T) {
	for _, id := range []string{"999", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			rec := do(t, newHandlerWithStartDate(t), http.MethodPatch, "/api/proyectos/"+id+"/fecha-fin", fechaFin("2026-12-18"))

			expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
		})
	}
}

// CA-41.4 (orden de validación): la fecha se valida antes que el proyecto.
func TestRegistrarFechaFinProyecto_FechaVaciaEnProyectoInexistenteDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPatch, "/api/proyectos/999/fecha-fin", fechaFin(""))

	expectError(t, rec, http.StatusBadRequest, project.ErrMissingEndDate.Error())
}

// CA-41.6
func TestRegistrarFechaFinProyecto_SinFechaDeInicioDevuelve409(t *testing.T) {
	h := newHandler()
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18"))

	expectError(t, rec, http.StatusConflict, project.ErrStartDateNotRegistered.Error())
	expectProjectEndDate(t, h, "1", "")
}

// CA-41.2: el mismo día o un día anterior al inicio dan 409.
func TestRegistrarFechaFinProyecto_FechaNoPosteriorAlInicioDevuelve409(t *testing.T) {
	for _, date := range []string{"2026-10-23", "2026-10-22"} {
		t.Run(date, func(t *testing.T) {
			h := newHandlerWithStartDate(t)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin(date))

			expectError(t, rec, http.StatusConflict, project.ErrEndNotAfterStart.Error())
			expectProjectEndDate(t, h, "1", "")
		})
	}
}

// CA-41.11 y CA-41.13: una falla del repositorio devuelve 500 sin detalle.
func TestRegistrarFechaFinProyecto_ErrorInternoDevuelve500(t *testing.T) {
	projects := project.NewManager(brokenRepository{})
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects)

	rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18"))

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}

// CA-41.8: con fecha de fin, un inicio el mismo día o después da 409, y uno
// anterior da 200 y conserva la fecha de fin.
func TestRegistrarFechaInicioProyecto_ConFechaDeFinValidaElOrden(t *testing.T) {
	for _, date := range []string{"2026-12-18", "2026-12-19"} {
		t.Run(date, func(t *testing.T) {
			h := newHandlerWithStartDate(t)
			expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18")), http.StatusOK)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio(date))

			expectError(t, rec, http.StatusConflict, project.ErrStartNotBeforeEnd.Error())
			expectProjectStartDate(t, h, "1", "2026-10-23")
		})
	}

	t.Run("anterior", func(t *testing.T) {
		h := newHandlerWithStartDate(t)
		expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18")), http.StatusOK)

		rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-12-17"))

		expectStatus(t, rec, http.StatusOK)
		var p projectJSON
		decode(t, rec, &p)
		if p.FechaInicio != "2026-12-17" || p.FechaFin != "2026-12-18" {
			t.Errorf("proyecto = %+v, se esperaba inicio 2026-12-17 y fin 2026-12-18", p)
		}
	})
}

// CA-41.9: POST, GET, PUT y PATCH …/fecha-inicio devuelven fecha_fin; "" si
// no está registrada.
func TestProyecto_RespuestasIncluyenLaFechaDeFin(t *testing.T) {
	h := newHandler()

	rec := do(t, h, http.MethodPost, "/api/proyectos", `{"nombre":"Sistema de ventas","descripcion":""}`)
	expectStatus(t, rec, http.StatusCreated)
	var body map[string]any
	decode(t, rec, &body)
	expectHasFechaFin(t, "POST", body, "")

	rec = do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-10-23"))
	expectStatus(t, rec, http.StatusOK)
	body = nil
	decode(t, rec, &body)
	expectHasFechaFin(t, "PATCH fecha-inicio sin fecha de fin", body, "")

	expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18")), http.StatusOK)

	rec = do(t, h, http.MethodGet, "/api/proyectos/1", "")
	expectStatus(t, rec, http.StatusOK)
	body = nil
	decode(t, rec, &body)
	expectHasFechaFin(t, "GET", body, "2026-12-18")

	rec = do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Ventas 2.0","descripcion":"","estado":"EN_CURSO"}`)
	expectStatus(t, rec, http.StatusOK)
	body = nil
	decode(t, rec, &body)
	expectHasFechaFin(t, "PUT", body, "2026-12-18")

	rec = do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-inicio", fechaInicio("2026-11-02"))
	expectStatus(t, rec, http.StatusOK)
	body = nil
	decode(t, rec, &body)
	expectHasFechaFin(t, "PATCH fecha-inicio con fecha de fin", body, "2026-12-18")
}

// CA-41.9 (RN-41.8): PUT no recibe la fecha de fin; si un cliente la envía,
// se ignora y la registrada se conserva.
func TestModificarProyecto_ConservaLaFechaDeFin(t *testing.T) {
	h := newHandlerWithStartDate(t)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18")), http.StatusOK)

	rec := do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Ventas 2.0","descripcion":"","estado":"EN_CURSO","fecha_fin":"2030-01-01"}`)

	expectStatus(t, rec, http.StatusOK)
	expectProjectEndDate(t, h, "1", "2026-12-18")
}

// CA-41.9 (D9): el reporte del proyecto (HU-38) incluye fecha_fin en su campo
// proyecto.
func TestReporteProyecto_IncluyeLaFechaDeFin(t *testing.T) {
	h := newHandlerWithStartDate(t)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18")), http.StatusOK)

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/reporte", "")

	expectStatus(t, rec, http.StatusOK)
	var body struct {
		Proyecto map[string]any `json:"proyecto"`
	}
	decode(t, rec, &body)
	expectHasFechaFin(t, "reporte", body.Proyecto, "2026-12-18")
}
