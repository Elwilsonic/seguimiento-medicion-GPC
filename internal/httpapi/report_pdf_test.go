package httpapi_test

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/httpapi"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// HU-40 · Exportar un reporte a un archivo PDF (Issue #58)
//
// CA-40.1: el PDF se descarga en GET /api/proyectos/{id}/reporte/pdf. CA-40.3:
// tiene los mismos valores que GET /api/proyectos/{id}/reporte. Los errores
// responden en JSON, como el reporte de HU-38 (CA-40.2 y CA-40.6).

// pdfBytes devuelve un texto tal como queda dentro del PDF: en Windows-1252,
// con los paréntesis y la barra invertida escapados. Alcanza para los textos
// de estos tests (ASCII y Latin-1).
func pdfBytes(s string) []byte {
	var b []byte
	for _, r := range s {
		switch {
		case r == '(' || r == ')' || r == '\\':
			b = append(b, '\\', byte(r))
		case r < 0x80 || (r >= 0xA0 && r <= 0xFF):
			b = append(b, byte(r))
		default:
			b = append(b, '.')
		}
	}
	return b
}

func expectPDFHas(t *testing.T, pdf []byte, texts ...string) {
	t.Helper()
	for _, text := range texts {
		if !bytes.Contains(pdf, pdfBytes(text)) {
			t.Errorf("el PDF no contiene %q", text)
		}
	}
}

// CA-40.1 · Responde 200 con el archivo PDF y los encabezados de descarga.
func TestReportePDF_Devuelve200ConElArchivo(t *testing.T) {
	h := newHandler()
	reportTEC03(t, h)

	// get y no do, porque do exige una respuesta JSON.
	rec := get(t, h, "/api/proyectos/1/reporte/pdf")

	expectStatus(t, rec, http.StatusOK)
	if got := rec.Header().Get("Content-Type"); got != "application/pdf" {
		t.Errorf("Content-Type = %q, se esperaba application/pdf", got)
	}
	if got, want := rec.Header().Get("Content-Disposition"), `attachment; filename="reporte-proyecto-1.pdf"`; got != want {
		t.Errorf("Content-Disposition = %q, se esperaba %q", got, want)
	}
	pdf := rec.Body.Bytes()
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) || !bytes.HasSuffix(bytes.TrimSpace(pdf), []byte("%%EOF")) {
		t.Fatalf("el cuerpo no es un PDF: %q", rec.Body.String()[:min(rec.Body.Len(), 40)])
	}
	expectPDFHas(t, pdf,
		"Reporte del proyecto «Software Metrics»",
		"Inicio: 2026-10-05",
		"#3 · 2026-10-08 · Ana · Pruebas · ítem 3 · sprint 1 · 3 h",
		// La línea del defecto sigue en la línea siguiente porque es más larga
		// que el ancho de la página: se verifica su comienzo.
		"#1 · ALTA · RESUELTO · ítem 1 · detectado en sprint 1",
		"Integrantes que registraron esfuerzo: Ana, Luis",
	)
}

// CA-40.3 · Los valores del PDF son los mismos que los del reporte en JSON
// (HU-38), con el mismo redondeo.
func TestReportePDF_TieneLosMismosValoresQueElReporteJSON(t *testing.T) {
	h := newHandler()
	reportTEC03(t, h)
	var rep struct {
		Metricas struct {
			SPPlanificados int     `json:"sp_planificados"`
			SPCompletados  int     `json:"sp_completados"`
			Velocidad      float64 `json:"velocidad"`
			Porcentaje     float64 `json:"porcentaje_completadas"`
			HorasEst       float64 `json:"horas_estimadas"`
			HorasReales    float64 `json:"horas_reales"`
			Detectados     int     `json:"defectos_detectados"`
			Resueltos      int     `json:"defectos_resueltos"`
		} `json:"metricas"`
		Esfuerzos struct {
			TotalHoras float64 `json:"total_horas"`
		} `json:"esfuerzos"`
		Defectos struct {
			Cantidad int `json:"cantidad"`
		} `json:"defectos"`
		Integrantes struct {
			Integrantes []string `json:"integrantes"`
		} `json:"integrantes"`
	}
	jsonRec := do(t, h, http.MethodGet, "/api/proyectos/1/reporte", "")
	expectStatus(t, jsonRec, http.StatusOK)
	decode(t, jsonRec, &rep)

	// get y no do, porque do exige una respuesta JSON.
	rec := get(t, h, "/api/proyectos/1/reporte/pdf")

	expectStatus(t, rec, http.StatusOK)
	m := rep.Metricas
	expectPDFHas(t, rec.Body.Bytes(),
		fmt.Sprintf("Story points: %d planificados, %d completados", m.SPPlanificados, m.SPCompletados),
		fmt.Sprintf("Velocidad: %.1f SP por sprint", m.Velocidad),
		fmt.Sprintf("Historias completadas: %.1f %%", m.Porcentaje),
		fmt.Sprintf("Horas: %.1f estimadas, %.1f reales", m.HorasEst, m.HorasReales),
		fmt.Sprintf("Defectos: %d detectados, %d resuelto(s).", m.Detectados, m.Resueltos),
		fmt.Sprintf("Total: %.1f horas", rep.Esfuerzos.TotalHoras),
		fmt.Sprintf("Cantidad: %d defecto(s)", rep.Defectos.Cantidad),
		"Integrantes que registraron esfuerzo: "+strings.Join(rep.Integrantes.Integrantes, ", "),
	)
}

// CA-40.6 · Un proyecto inexistente, 0 o negativo devuelve 404 en JSON.
func TestReportePDF_InexistenteDevuelve404(t *testing.T) {
	h := newHandler()
	for _, id := range []string{"99", "0", "-1"} {
		rec := do(t, h, http.MethodGet, "/api/proyectos/"+id+"/reporte/pdf", "")

		expectError(t, rec, http.StatusNotFound, project.ErrProjectNotFound.Error())
	}
}

// CA-40.1 · Un {id} no numérico devuelve 400 en JSON.
func TestReportePDF_IDNoNumericoDevuelve400(t *testing.T) {
	rec := do(t, newHandler(), http.MethodGet, "/api/proyectos/abc/reporte/pdf", "")

	expectError(t, rec, http.StatusBadRequest, "el id debe ser un número entero")
}

// CA-40.2 · Una falla al armar el reporte devuelve 500 en JSON, sin detalle
// y sin un PDF a medias.
func TestReportePDF_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	items := backlog.NewManager(brokenItemRepository{}, projects)
	sprints := sprint.NewManager(sprint.NewMemoryRepository(), projects, items)
	efforts := effort.NewManager(effort.NewMemoryRepository(), sprints, items)
	pokerManager := poker.NewManager(poker.NewMemoryRepository(), items)
	defects := defect.NewManager(defect.NewMemoryRepository(), items, sprints)
	h := httpapi.NewHandler(projects, items, sprints, efforts, pokerManager, defects, newMembers(projects))
	createProject(t, h, "Sistema de ventas")

	rec := do(t, h, http.MethodGet, "/api/proyectos/1/reporte/pdf", "")

	expectError(t, rec, http.StatusInternalServerError, "error interno del servidor")
	if bytes.Contains(rec.Body.Bytes(), []byte("%PDF-")) {
		t.Error("la respuesta de error contiene un PDF")
	}
}
