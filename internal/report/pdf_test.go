package report

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/metrics"
	"seguimiento-medicion-gpc/internal/project"
)

// HU-40 · Exportar un reporte a un archivo PDF (Issue #58)
//
// El contenido del PDF (projectReportLines) se prueba comparando listas de
// texto (CA-40.3); el archivo (WriteProjectPDF) se prueba buscando cada línea
// en los bytes del PDF, que se escribe sin comprimir (decisión 6 de la SDD).

// tec03Lines son las 24 líneas del ejemplo de TEC-03 de la SDD, más las 6 de
// la sección "Historias" de HU-49 (CA-49.6).
var tec03Lines = []pdfLine{
	{Text: "Reporte del proyecto «Software Metrics»", Heading: true},
	{Text: "Proyecto", Heading: true},
	{Text: "Nombre: Software Metrics"},
	{Text: "Descripción: Estimación y seguimiento de proyectos"},
	{Text: "Estado: EN_CURSO"},
	{Text: "Inicio: 2026-10-05"},
	{Text: "Fin: sin registrar"},
	{Text: "Métricas", Heading: true},
	{Text: "Story points: 10 planificados, 8 completados, 1 ítem(s) sin estimar."},
	{Text: "Velocidad: 8.0 SP por sprint (1 sprint(s) cerrado(s))."},
	{Text: "Historias completadas: 50.0 % (2 de 4)."},
	{Text: "Horas: 20.0 estimadas, 20.0 reales, 1 ítem(s) sin horas estimadas."},
	{Text: "Desviación: 0.0 h (0.0 %)."},
	{Text: "Defectos: 1 detectados, 1 resuelto(s)."},
	{Text: "Historias", Heading: true},
	{Text: "#1 · A · MEDIA · TERMINADO · 5 SP · 10.0 h estimadas · sprint 1"},
	{Text: "#2 · B · MEDIA · TERMINADO · 3 SP · 6.0 h estimadas · sprint 1"},
	{Text: "#3 · C · MEDIA · EN_CURSO · 2 SP · sin horas estimadas · sprint 1"},
	{Text: "#4 · D · MEDIA · PENDIENTE · sin estimar · 4.0 h estimadas · sprint 1"},
	{Text: "Cantidad: 4 historia(s), 2 completada(s)"},
	{Text: "Esfuerzos", Heading: true},
	{Text: "#1 · 2026-10-06 · Ana · Desarrollo · ítem 1 · sprint 1 · 12 h"},
	{Text: "#2 · 2026-10-06 · Luis · Desarrollo · ítem 2 · sprint 1 · 5 h"},
	{Text: "#3 · 2026-10-06 · Ana · Desarrollo · ítem 3 · sprint 1 · 3 h"},
	{Text: "Total: 20.0 horas"},
	{Text: "Defectos", Heading: true},
	{Text: "#1 · ALTA · RESUELTO · ítem 1 · detectado en sprint 1 · resuelto en sprint 1 · Defecto"},
	{Text: "Cantidad: 1 defecto(s)"},
	{Text: "Integrantes", Heading: true},
	{Text: "Integrantes que registraron esfuerzo: Ana, Luis"},
}

// emptyProjectLines son las líneas de un proyecto recién creado, sin datos,
// con la sección "Historias" vacía de HU-49 (CA-49.4 y CA-49.6).
var emptyProjectLines = []pdfLine{
	{Text: "Reporte del proyecto «Proyecto vacío»", Heading: true},
	{Text: "Proyecto", Heading: true},
	{Text: "Nombre: Proyecto vacío"},
	{Text: "Descripción: "},
	{Text: "Estado: PLANIFICADO"},
	{Text: "Inicio: sin registrar"},
	{Text: "Fin: sin registrar"},
	{Text: "Métricas", Heading: true},
	{Text: "Story points: 0 planificados, 0 completados, 0 ítem(s) sin estimar."},
	{Text: "Velocidad: sin datos (0 sprint(s) cerrado(s))."},
	{Text: "Historias completadas: sin datos."},
	{Text: "Horas: 0.0 estimadas, 0.0 reales, 0 ítem(s) sin horas estimadas."},
	{Text: "Desviación: sin datos."},
	{Text: "Defectos: 0 detectados, 0 resuelto(s)."},
	{Text: "Historias", Heading: true},
	{Text: "Sin historias en el backlog."},
	{Text: "Cantidad: 0 historia(s), 0 completada(s)"},
	{Text: "Esfuerzos", Heading: true},
	{Text: "Sin esfuerzos registrados."},
	{Text: "Total: 0.0 horas"},
	{Text: "Defectos", Heading: true},
	{Text: "Sin defectos registrados."},
	{Text: "Cantidad: 0 defecto(s)"},
	{Text: "Integrantes", Heading: true},
	{Text: "Integrantes que registraron esfuerzo: ninguno"},
}

// pdfText devuelve un texto tal como queda dentro del PDF: en Windows-1252,
// con los paréntesis y la barra invertida escapados. Alcanza para los textos
// de estos tests (ASCII y Latin-1: tildes, ñ, «», ·); cualquier otro
// carácter queda como ".", como en el PDF (RN-40.6).
func pdfText(s string) []byte {
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

// expectValidPDF falla si pdf no empieza con %PDF- o no termina con %%EOF.
func expectValidPDF(t *testing.T, pdf []byte) {
	t.Helper()
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("el archivo no empieza con %%PDF-: %q", firstBytes(pdf))
	}
	if !bytes.HasSuffix(bytes.TrimSpace(pdf), []byte("%%EOF")) {
		t.Fatal("el archivo no termina con la marca de fin del PDF")
	}
}

func firstBytes(b []byte) []byte {
	if len(b) > 20 {
		return b[:20]
	}
	return b
}

// expectPDFContains falla si el PDF no contiene alguno de los textos.
func expectPDFContains(t *testing.T, pdf []byte, texts ...string) {
	t.Helper()
	for _, text := range texts {
		if !bytes.Contains(pdf, pdfText(text)) {
			t.Errorf("el PDF no contiene %q", text)
		}
	}
}

// writePDF escribe el PDF del reporte en memoria y falla si hay error.
func writePDF(t *testing.T, rep ProjectReport) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := WriteProjectPDF(&buf, rep); err != nil {
		t.Fatalf("WriteProjectPDF devolvió %v", err)
	}
	return buf.Bytes()
}

// failingWriter es un destino que no acepta la escritura (CA-40.2).
type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

// CA-40.3 · Las líneas del ejemplo de TEC-03 de la SDD.
func TestProjectReportLines_EjemploDeTEC03(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.tec03Project(t))

	got := projectReportLines(rep)

	if !reflect.DeepEqual(got, tec03Lines) {
		t.Errorf("líneas =\n%v\nse esperaba\n%v", got, tec03Lines)
	}
}

// CA-40.5 · Un proyecto sin datos tiene las secciones vacías y "sin datos".
func TestProjectReportLines_ProyectoSinDatos(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.project(t, "Proyecto vacío", ""))

	got := projectReportLines(rep)

	if !reflect.DeepEqual(got, emptyProjectLines) {
		t.Errorf("líneas =\n%v\nse esperaba\n%v", got, emptyProjectLines)
	}
}

// CA-49.4 y CA-49.6 · La línea de una historia indica la prioridad, el
// estado, los Story Points, las horas estimadas con un decimal y el sprint, o
// "sin estimar", "sin horas estimadas" y "sin sprint".
func TestProjectReportLines_HistoriasConYSinEstimacion(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Sistema de ventas", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	e.story(t, projectID, "Registrar venta", backlog.PriorityHigh, 3, 8, backlog.StatusFinished, s.ID)
	e.story(t, projectID, "Consultar stock", backlog.PriorityMedium, 0, 0, backlog.StatusPending, 0)
	e.story(t, projectID, "Emitir factura", backlog.PriorityLow, 0, 2.25, backlog.StatusInProgress, 0)

	got := projectReportLines(generate(t, e, projectID))

	expectLines(t, got,
		"#1 · Registrar venta · ALTA · TERMINADO · 3 SP · 8.0 h estimadas · sprint 1",
		"#2 · Consultar stock · MEDIA · PENDIENTE · sin estimar · sin horas estimadas · sin sprint",
		"#3 · Emitir factura · BAJA · EN_CURSO · sin estimar · 2.3 h estimadas · sin sprint",
		"Cantidad: 3 historia(s), 1 completada(s)",
	)
}

// CA-40.3 · Con la fecha de fin de HU-41 registrada, se muestra.
func TestProjectReportLines_ConFechaDeFin(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Con fechas", "")
	e.startProject(t, projectID, time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC))
	if _, err := e.projects.SetEndDate(projectID, time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("registrar la fecha de fin: %v", err)
	}

	got := projectReportLines(generate(t, e, projectID))

	expectLines(t, got, "Inicio: 2026-10-05", "Fin: 2026-12-18")
}

// expectLines falla si alguno de los textos no es una línea de lines.
func expectLines(t *testing.T, lines []pdfLine, texts ...string) {
	t.Helper()
	for _, text := range texts {
		found := false
		for _, l := range lines {
			if l.Text == text {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no hay una línea %q en %v", text, lines)
		}
	}
}

// RN-40.3 · Las métricas y el total se redondean a un decimal como la API,
// sin -0.0, y se escriben con un decimal.
func TestProjectReportLines_RedondeaAUnDecimal(t *testing.T) {
	rep := ProjectReport{
		Project: project.Project{ID: 1, Name: "Redondeo", Status: project.StatusInProgress},
		Metrics: metrics.ProjectMetrics{
			ProjectID:  1,
			Points:     metrics.ProjectPoints{ProjectID: 1, Planned: 3, Completed: 2, Velocity: 2.25, VelocityAvailable: true, ClosedSprints: 2},
			Completion: metrics.Completion{Items: 3, Finished: 2, Percent: 200.0 / 3, PercentAvailable: true},
			Hours:      metrics.Hours{Estimated: 10.04, Actual: 10, Deviation: -0.04, DeviationPercent: -0.398, DeviationAvailable: true},
		},
		Efforts: Efforts{Efforts: []effort.Effort{}, TotalHours: 0.1 + 0.2},
		Defects: Defects{Defects: []defect.Defect{}},
		Members: Members{Names: []string{}},
	}

	got := projectReportLines(rep)

	expectLines(t, got,
		"Velocidad: 2.3 SP por sprint (2 sprint(s) cerrado(s)).",
		"Historias completadas: 66.7 % (2 de 3).",
		"Horas: 10.0 estimadas, 10.0 reales, 0 ítem(s) sin horas estimadas.",
		"Desviación: 0.0 h (-0.4 %).",
		"Total: 0.3 horas",
	)
}

// RN-40.3 · Las horas de cada esfuerzo van tal como están guardadas.
func TestProjectReportLines_HorasDeCadaEsfuerzoTalCual(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Horas", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, 0, backlog.StatusPending, s.ID)
	e.effort(t, s.ID, it.ID, "Ana", 1.5)
	e.effort(t, s.ID, it.ID, "Luis", 0.25)

	got := projectReportLines(generate(t, e, projectID))

	expectLines(t, got,
		"#1 · 2026-10-06 · Ana · Desarrollo · ítem 1 · sprint 1 · 1.5 h",
		"#2 · 2026-10-06 · Luis · Desarrollo · ítem 1 · sprint 1 · 0.25 h",
		"Total: 1.8 horas",
	)
}

// CA-40.3 · Un defecto sin sprint de resolución lo indica.
func TestProjectReportLines_DefectoSinSprintDeResolucion(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Defectos", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, 0, backlog.StatusPending, s.ID)
	e.saveDefect(t, it.ID, s.ID, defect.StatusOpen, 0)

	got := projectReportLines(generate(t, e, projectID))

	expectLines(t, got,
		"#1 · ALTA · ABIERTO · ítem 1 · detectado en sprint 1 · sin sprint de resolución · Defecto",
		"Cantidad: 1 defecto(s)",
	)
}

// CA-40.1 y CA-40.3 · El archivo es un PDF y contiene cada línea del
// reporte, con los títulos en negrita.
func TestWriteProjectPDF_GeneraUnPDFConCadaLineaDelReporte(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.tec03Project(t))

	pdf := writePDF(t, rep)

	expectValidPDF(t, pdf)
	for _, l := range tec03Lines {
		expectPDFContains(t, pdf, l.Text)
	}
	expectPDFContains(t, pdf, "/BaseFont /Helvetica-Bold", "/BaseFont /Helvetica")
}

// CA-40.5 · Un proyecto sin datos genera un PDF, no un error.
func TestWriteProjectPDF_ProyectoSinDatos(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.project(t, "Proyecto vacío", ""))

	pdf := writePDF(t, rep)

	expectValidPDF(t, pdf)
	for _, l := range emptyProjectLines {
		expectPDFContains(t, pdf, l.Text)
	}
}

// CA-40.7 · Las tildes, la ñ y los signos se ven, y un carácter fuera de
// Windows-1252 se reemplaza por "." sin impedir el PDF.
func TestWriteProjectPDF_TildesYCaracteresFueraDeWindows1252(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.project(t, "Gestión de ñandúes «Sur» 中", "Año ° (beta) \\ final"))

	pdf := writePDF(t, rep)

	expectValidPDF(t, pdf)
	expectPDFContains(t, pdf,
		"Nombre: Gestión de ñandúes «Sur» .",
		"Descripción: Año ° (beta) \\ final",
	)
}

// Un texto más largo que el ancho de la página sigue en la línea siguiente, y
// muchos esfuerzos suman páginas.
func TestWriteProjectPDF_TextoLargoYVariasPaginas(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Largo", strings.Repeat("palabra ", 300))
	s := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, 0, backlog.StatusPending, s.ID)
	for range 120 {
		e.effort(t, s.ID, it.ID, "Ana", 1)
	}

	pdf := writePDF(t, generate(t, e, projectID))

	expectValidPDF(t, pdf)
	expectPDFContains(t, pdf, "Total: 120.0 horas", "#120 · 2026-10-06 · Ana")
	if pages := bytes.Count(pdf, []byte("<</Type /Page\n")); pages < 2 {
		t.Errorf("el PDF tiene %d página(s), se esperaban al menos 2", pages)
	}
}

// CA-40.2 · Si el destino no acepta la escritura, se devuelve ese error.
func TestWriteProjectPDF_ErrorAlEscribirDevuelveError(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.tec03Project(t))
	errDisk := errors.New("disco lleno")

	err := WriteProjectPDF(failingWriter{err: errDisk}, rep)

	if !errors.Is(err, errDisk) {
		t.Errorf("err = %v, se esperaba %v", err, errDisk)
	}
}

// CA-40.8 · Exportar no modifica el reporte recibido.
func TestWriteProjectPDF_NoModificaElReporte(t *testing.T) {
	e := newReportEnv()
	projectID := e.tec03Project(t)
	rep := generate(t, e, projectID)
	want := generate(t, e, projectID)

	writePDF(t, rep)

	if !reflect.DeepEqual(rep, want) {
		t.Errorf("el reporte cambió:\n%+v\nse esperaba\n%+v", rep, want)
	}
}
