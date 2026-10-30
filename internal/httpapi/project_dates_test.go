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

// HU-19 · Modificar las fechas del proyecto (Issue #37)
//
// Trazabilidad: CA-19.1 a CA-19.5 y CA-19.8. PATCH /api/proyectos/{id}/fechas,
// con un caso exitoso por cada dirección (postergar y adelantar) y un test por
// código de error. El 413, el texto después del JSON y el ID no numérico de
// la ruta se prueban en server_test.go.

// fechas arma el cuerpo de PATCH /api/proyectos/{id}/fechas.
func fechas(inicio, fin string) string {
	return `{"fecha_inicio":"` + inicio + `","fecha_fin":"` + fin + `"}`
}

// newHandlerWithDates arma la API con el proyecto 1 "Sistema de ventas" con
// la fecha de inicio 2026-10-23 y la de fin 2026-12-18.
func newHandlerWithDates(t *testing.T) http.Handler {
	t.Helper()
	h := newHandlerWithStartDate(t)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/proyectos/1/fecha-fin", fechaFin("2026-12-18")), http.StatusOK)
	return h
}

// expectProjectDates verifica las dos fechas que devuelve la consulta del
// proyecto.
func expectProjectDates(t *testing.T, h http.Handler, id, inicio, fin string) {
	t.Helper()
	rec := do(t, h, http.MethodGet, "/api/proyectos/"+id, "")
	expectStatus(t, rec, http.StatusOK)
	var p projectJSON
	decode(t, rec, &p)
	if p.FechaInicio != inicio || p.FechaFin != fin {
		t.Errorf("fechas consultadas: inicio %q y fin %q, se esperaban %q y %q", p.FechaInicio, p.FechaFin, inicio, fin)
	}
}

// CA-19.1 y CA-19.8: postergar y adelantar las dos fechas devuelve 200 con el
// proyecto, sin cambiar sus demás datos.
func TestModificarFechasProyecto_Devuelve200ConElProyecto(t *testing.T) {
	casos := map[string][2]string{
		"postergar": {"2027-01-11", "2027-03-19"},
		"adelantar": {"2026-06-01", "2026-08-31"},
	}
	for caseName, dates := range casos {
		t.Run(caseName, func(t *testing.T) {
			h := newHandlerWithDates(t)
			expectStatus(t, do(t, h, http.MethodPut, "/api/proyectos/1", `{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas","estado":"EN_CURSO"}`), http.StatusOK)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fechas", fechas(dates[0], dates[1]))

			expectStatus(t, rec, http.StatusOK)
			var p projectJSON
			decode(t, rec, &p)
			want := projectJSON{ID: 1, Nombre: "Sistema de ventas", Descripcion: "Gestión de ventas", Estado: "EN_CURSO", FechaInicio: dates[0], FechaFin: dates[1]}
			if p != want {
				t.Errorf("proyecto = %+v, se esperaba %+v", p, want)
			}
			expectProjectDates(t, h, "1", dates[0], dates[1])
		})
	}
}

// CA-19.5: sin fecha de inicio (vacía, null o ausente), con o sin fecha de
// fin, da 400.
func TestModificarFechasProyecto_FechaDeInicioAusenteDevuelve400(t *testing.T) {
	bodies := map[string]string{
		"vacía":          `{"fecha_inicio":"","fecha_fin":"2027-03-19"}`,
		"null":           `{"fecha_inicio":null,"fecha_fin":"2027-03-19"}`,
		"ausente":        `{"fecha_fin":"2027-03-19"}`,
		"las dos vacías": `{"fecha_inicio":"","fecha_fin":""}`,
		"las dos null":   `{"fecha_inicio":null,"fecha_fin":null}`,
		"cuerpo vacío":   `{}`,
	}
	for caseName, body := range bodies {
		t.Run(caseName, func(t *testing.T) {
			h := newHandlerWithDates(t)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fechas", body)

			expectError(t, rec, http.StatusBadRequest, project.ErrMissingStartDate.Error())
			expectProjectDates(t, h, "1", "2026-10-23", "2026-12-18")
		})
	}
}

// CA-19.5: sin fecha de fin (vacía, null o ausente), con fecha de inicio
// válida, da 400.
func TestModificarFechasProyecto_FechaDeFinAusenteDevuelve400(t *testing.T) {
	bodies := map[string]string{
		"vacía":   `{"fecha_inicio":"2027-01-11","fecha_fin":""}`,
		"null":    `{"fecha_inicio":"2027-01-11","fecha_fin":null}`,
		"ausente": `{"fecha_inicio":"2027-01-11"}`,
	}
	for caseName, body := range bodies {
		t.Run(caseName, func(t *testing.T) {
			h := newHandlerWithDates(t)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fechas", body)

			expectError(t, rec, http.StatusBadRequest, project.ErrMissingEndDate.Error())
			expectProjectDates(t, h, "1", "2026-10-23", "2026-12-18")
		})
	}
}

// CA-19.5 (RN-19.2): una fecha con otro formato o inexistente, en cualquiera
// de los dos campos, da 400.
func TestModificarFechasProyecto_FormatoOFechaInvalidaDevuelve400(t *testing.T) {
	for _, date := range []string{"2027-02-30", "2027-13-01", "11/01/2027", "2027-1-11", "2027-01-11T00:00:00Z", " 2027-01-11 "} {
		for campo, body := range map[string]string{
			"fecha_inicio": fechas(date, "2027-03-19"),
			"fecha_fin":    fechas("2027-01-11", date),
		} {
			t.Run(campo+"_"+date, func(t *testing.T) {
				h := newHandlerWithDates(t)

				rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fechas", body)

				expectError(t, rec, http.StatusBadRequest, "las fechas deben tener el formato AAAA-MM-DD")
				expectProjectDates(t, h, "1", "2026-10-23", "2026-12-18")
			})
		}
	}
}

// CA-19.8: un cuerpo que no es JSON o una fecha que no es texto es JSON
// inválido.
func TestModificarFechasProyecto_JSONInvalidoDevuelve400(t *testing.T) {
	for _, body := range []string{
		`no es json`,
		`{"fecha_inicio":20270111,"fecha_fin":"2027-03-19"}`,
		`{"fecha_inicio":"2027-01-11","fecha_fin":20270319}`,
	} {
		t.Run(body, func(t *testing.T) {
			h := newHandlerWithDates(t)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fechas", body)

			expectError(t, rec, http.StatusBadRequest, "el cuerpo del pedido no es un JSON válido")
			expectProjectDates(t, h, "1", "2026-10-23", "2026-12-18")
		})
	}
}

// CA-19.8: un {id} que no es un número entero da 400.
func TestModificarFechasProyecto_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandlerWithDates(t), http.MethodPatch, "/api/proyectos/abc/fechas", fechas("2027-01-11", "2027-03-19"))

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-19.3
func TestModificarFechasProyecto_InexistenteDevuelve404(t *testing.T) {
	for _, id := range []string{"999", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			rec := do(t, newHandlerWithDates(t), http.MethodPatch, "/api/proyectos/"+id+"/fechas", fechas("2027-01-11", "2027-03-19"))

			expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
		})
	}
}

// CA-19.5 (orden de validación): las fechas se validan antes que el proyecto.
func TestModificarFechasProyecto_FechasVaciasEnProyectoInexistenteDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodPatch, "/api/proyectos/999/fechas", fechas("", ""))

	expectError(t, rec, http.StatusBadRequest, project.ErrMissingStartDate.Error())
}

// CA-19.4: un proyecto sin fechas o solo con fecha de inicio da 409.
func TestModificarFechasProyecto_SinLasDosFechasDevuelve409(t *testing.T) {
	t.Run("sin fechas", func(t *testing.T) {
		h := newHandler()
		createProject(t, h, "Sistema de ventas")

		rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fechas", fechas("2027-01-11", "2027-03-19"))

		expectError(t, rec, http.StatusConflict, project.ErrDatesNotRegistered.Error())
		expectProjectDates(t, h, "1", "", "")
	})

	t.Run("solo con inicio", func(t *testing.T) {
		h := newHandlerWithStartDate(t)

		rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fechas", fechas("2027-01-11", "2027-03-19"))

		expectError(t, rec, http.StatusConflict, project.ErrDatesNotRegistered.Error())
		expectProjectDates(t, h, "1", "2026-10-23", "")
	})
}

// CA-19.2 (D4): un fin nuevo igual o anterior al inicio nuevo da 409 y no
// cambia ninguna de las dos fechas.
func TestModificarFechasProyecto_FinNoPosteriorAlInicioDevuelve409(t *testing.T) {
	for _, fin := range []string{"2027-01-11", "2027-01-10"} {
		t.Run(fin, func(t *testing.T) {
			h := newHandlerWithDates(t)

			rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fechas", fechas("2027-01-11", fin))

			expectError(t, rec, http.StatusConflict, project.ErrEndNotAfterStart.Error())
			expectProjectDates(t, h, "1", "2026-10-23", "2026-12-18")
		})
	}
}

// CA-19.6 y CA-19.8: una falla del repositorio devuelve 500 sin detalle.
func TestModificarFechasProyecto_ErrorInternoDevuelve500(t *testing.T) {
	projects := project.NewManager(brokenRepository{})
	items := backlog.NewManager(backlog.NewMemoryRepository(), projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))

	rec := do(t, h, http.MethodPatch, "/api/proyectos/1/fechas", fechas("2027-01-11", "2027-03-19"))

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
}
