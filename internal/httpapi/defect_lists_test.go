package httpapi_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-31 · Consultar los defectos de un proyecto o de un sprint por HTTP
// (Issue #49)
//
// Trazabilidad: CA-31.9. GET /api/proyectos/{id}/defectos y
// GET /api/sprints/{id}/defectos exponen la consulta, con ?estado= opcional.
// Hay un caso exitoso por ruta, el filtro (incluidos ?estado= y un filtro con
// solo espacios, que no filtran, D3), lista vacía y un test por cada código de
// error de la tabla HTTP de la SDD (400, 404 y 500).

type defectListJSON struct {
	Defectos []defectJSON `json:"defectos"`
	Cantidad int          `json:"cantidad"`
}

// List completa brokenDefectRepository (HU-29 y HU-30) para que siga
// cumpliendo defect.Repository. Se declara acá para no modificar
// defects_test.go.
func (brokenDefectRepository) List() ([]defect.Defect, error) {
	return nil, errStorage
}

// newHandlerWithDefectList arma la API con:
//   - el proyecto 1, su ítem 1 y su sprint 1 abierto, con los defectos 1
//     (ABIERTO), 2 (EN_CORRECCION) y 3 (RESUELTO en el sprint 1);
//   - el proyecto 2, su ítem 2 y su sprint 2 abierto, con el defecto 4
//     (ABIERTO).
func newHandlerWithDefectList(t *testing.T) http.Handler {
	t.Helper()
	h := newHandlerWithDefects(t)
	for range 3 {
		expectStatus(t, do(t, h, http.MethodPost, defectsPath, validDefect), http.StatusCreated)
	}
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/2/estado", `{"estado":"EN_CORRECCION"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/3/estado", `{"estado":"EN_CORRECCION"}`), http.StatusOK)
	expectStatus(t, do(t, h, http.MethodPatch, "/api/defectos/3/estado", `{"estado":"RESUELTO","sprint_resolucion_id":1}`), http.StatusOK)

	createProject(t, h, "Sistema de stock")
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/2/backlog", validItem), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/proyectos/2/sprints", validSprint), http.StatusCreated)
	other := `{"descripcion":"El stock queda negativo","severidad":"MEDIA","sprint_deteccion_id":2}`
	expectStatus(t, do(t, h, http.MethodPost, "/api/items/2/defectos", other), http.StatusCreated)
	return h
}

// listDefects hace el GET, exige 200 y devuelve la respuesta.
func listDefects(t *testing.T, h http.Handler, path string) defectListJSON {
	t.Helper()
	rec := do(t, h, http.MethodGet, path, "")
	expectStatus(t, rec, http.StatusOK)
	var got defectListJSON
	decode(t, rec, &got)
	return got
}

func defectIDs(list defectListJSON) []int {
	ids := make([]int, 0, len(list.Defectos))
	for _, d := range list.Defectos {
		ids = append(ids, d.ID)
	}
	return ids
}

func assertDefectIDs(t *testing.T, path string, got defectListJSON, want []int) {
	t.Helper()
	if ids := defectIDs(got); !slices.Equal(ids, want) {
		t.Errorf("%s: IDs = %v, se esperaban %v", path, ids, want)
	}
	if got.Cantidad != len(want) {
		t.Errorf("%s: cantidad = %d, se esperaba %d", path, got.Cantidad, len(want))
	}
}

func TestListarDefectosDelProyecto_Devuelve200ConLosDefectosYLaCantidad(t *testing.T) {
	h := newHandlerWithDefectList(t)

	got := listDefects(t, h, "/api/proyectos/1/defectos")

	assertDefectIDs(t, "proyecto 1", got, []int{1, 2, 3})
	if len(got.Defectos) != 3 {
		t.Fatalf("defectos = %+v, se esperaban 3", got.Defectos)
	}
	wantStates := []defect.Status{defect.StatusOpen, defect.StatusInCorrection, defect.StatusResolved}
	for i, want := range wantStates {
		if got.Defectos[i].Estado != want {
			t.Errorf("estado del defecto %d = %q, se esperaba %q", got.Defectos[i].ID, got.Defectos[i].Estado, want)
		}
	}
	wantResolved := defectJSON{
		ID:                 3,
		Descripcion:        "El total no descuenta el descuento",
		Severidad:          defect.SeverityHigh,
		Estado:             defect.StatusResolved,
		ItemID:             1,
		SprintDeteccionID:  1,
		SprintResolucionID: 1,
	}
	if got.Defectos[2] != wantResolved {
		t.Errorf("defecto 3 = %+v, se esperaba %+v", got.Defectos[2], wantResolved)
	}

	assertDefectIDs(t, "proyecto 2", listDefects(t, h, "/api/proyectos/2/defectos"), []int{4})
}

func TestListarDefectosDelSprint_Devuelve200ConLosDefectosYLaCantidad(t *testing.T) {
	h := newHandlerWithDefectList(t)

	assertDefectIDs(t, "sprint 1", listDefects(t, h, "/api/sprints/1/defectos"), []int{1, 2, 3})
	assertDefectIDs(t, "sprint 2", listDefects(t, h, "/api/sprints/2/defectos"), []int{4})
}

func TestListarDefectos_FiltraPorEstado(t *testing.T) {
	h := newHandlerWithDefectList(t)
	cases := []struct {
		path string
		want []int
	}{
		{"/api/proyectos/1/defectos?estado=RESUELTO", []int{3}},
		{"/api/proyectos/1/defectos?estado=en_correccion", []int{2}},
		{"/api/proyectos/1/defectos?estado=%20Abierto%20", []int{1}},
		{"/api/sprints/1/defectos?estado=ABIERTO", []int{1}},
		{"/api/sprints/1/defectos?estado=resuelto", []int{3}},
		{"/api/sprints/2/defectos?estado=ABIERTO", []int{4}},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			assertDefectIDs(t, c.path, listDefects(t, h, c.path), c.want)
		})
	}
}

// D3 (ajuste de Hernán): el handler recorta ?estado= y, si queda vacío, no
// filtra; un filtro con solo espacios no da 400.
func TestListarDefectos_FiltroVacioOConSoloEspaciosNoFiltra(t *testing.T) {
	h := newHandlerWithDefectList(t)

	for _, base := range []string{"/api/proyectos/1/defectos", "/api/sprints/1/defectos"} {
		for _, query := range []string{"", "?estado=", "?estado=%20%20%20", "?estado=%09"} {
			path := base + query
			t.Run(path, func(t *testing.T) {
				assertDefectIDs(t, path, listDefects(t, h, path), []int{1, 2, 3})
			})
		}
	}
}

func TestListarDefectos_SinDefectosDevuelveListaVaciaYCantidadCero(t *testing.T) {
	empty := newHandlerWithDefects(t)
	populated := newHandlerWithDefectList(t)
	cases := []struct {
		name string
		h    http.Handler
		path string
	}{
		{"proyecto sin defectos", empty, "/api/proyectos/1/defectos"},
		{"sprint sin defectos", empty, "/api/sprints/1/defectos"},
		{"filtro sin coincidencias en el proyecto", populated, "/api/proyectos/2/defectos?estado=RESUELTO"},
		{"filtro sin coincidencias en el sprint", populated, "/api/sprints/2/defectos?estado=EN_CORRECCION"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := do(t, c.h, http.MethodGet, c.path, "")

			expectStatus(t, rec, http.StatusOK)
			body := rec.Body.String()
			if !strings.Contains(body, `"defectos":[]`) {
				t.Errorf("cuerpo = %s, se esperaba una lista vacía", body)
			}
			if !strings.Contains(body, `"cantidad":0`) {
				t.Errorf("cuerpo = %s, se esperaba la cantidad en 0", body)
			}
		})
	}
}

func TestListarDefectos_EstadoInvalidoDevuelve400(t *testing.T) {
	h := newHandlerWithDefectList(t)

	for _, path := range []string{
		"/api/proyectos/1/defectos?estado=CERRADO",
		"/api/sprints/1/defectos?estado=PENDIENTE",
		"/api/sprints/1/defectos?estado=EN%20CORRECCION",
		// El filtro se valida antes de consultar: 400 y no 404.
		"/api/proyectos/999/defectos?estado=CERRADO",
		"/api/sprints/999/defectos?estado=CERRADO",
	} {
		t.Run(path, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, path, "")

			expectError(t, rec, http.StatusBadRequest, defect.ErrInvalidStatus.Error())
		})
	}
}

func TestListarDefectos_InexistenteDevuelve404(t *testing.T) {
	h := newHandlerWithDefectList(t)
	cases := []struct {
		path string
		want error
	}{
		{"/api/proyectos/999/defectos", project.ErrProjectNotFound},
		{"/api/proyectos/0/defectos", project.ErrProjectNotFound},
		{"/api/proyectos/-1/defectos?estado=ABIERTO", project.ErrProjectNotFound},
		{"/api/sprints/999/defectos", sprint.ErrSprintNotFound},
		{"/api/sprints/0/defectos", sprint.ErrSprintNotFound},
		{"/api/sprints/-1/defectos?estado=ABIERTO", sprint.ErrSprintNotFound},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, c.path, "")

			expectError(t, rec, http.StatusNotFound, c.want.Error())
		})
	}
}

func TestListarDefectos_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	h := newHandlerWithDefectRepository(t, brokenDefectRepository{})

	for _, path := range []string{"/api/proyectos/1/defectos", "/api/sprints/1/defectos?estado=ABIERTO"} {
		t.Run(path, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, path, "")

			expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
		})
	}
}
