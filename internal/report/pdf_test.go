package report

import (
	"bytes"
	"errors"
	"reflect"
	"regexp"
	"strconv"
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
// la sección "Historias" de HU-49. Los IDs van como marcadores ({historia1},
// {sprint1}, {esfuerzo1}, {defecto1}…) que withPublicIDs reemplaza por los UUID
// del reporte (CA-49.6 y HU-52, CA-52.3).
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
	{Text: "#{historia1} · A · MEDIA · TERMINADO · 5 SP · 10.0 h estimadas · sprint {sprint1}"},
	{Text: "#{historia2} · B · MEDIA · TERMINADO · 3 SP · 6.0 h estimadas · sprint {sprint1}"},
	{Text: "#{historia3} · C · MEDIA · EN_CURSO · 2 SP · sin horas estimadas · sprint {sprint1}"},
	{Text: "#{historia4} · D · MEDIA · PENDIENTE · sin estimar · 4.0 h estimadas · sprint {sprint1}"},
	{Text: "Cantidad: 4 historia(s), 2 completada(s)"},
	{Text: "Esfuerzos", Heading: true},
	{Text: "#{esfuerzo1} · 2026-10-06 · Ana · Desarrollo · ítem {historia1} · sprint {sprint1} · 12 h"},
	{Text: "#{esfuerzo2} · 2026-10-06 · Luis · Desarrollo · ítem {historia2} · sprint {sprint1} · 5 h"},
	{Text: "#{esfuerzo3} · 2026-10-06 · Ana · Desarrollo · ítem {historia3} · sprint {sprint1} · 3 h"},
	{Text: "Total: 20.0 horas"},
	{Text: "Defectos", Heading: true},
	{Text: "#{defecto1} · ALTA · RESUELTO · ítem {historia1} · detectado en sprint {sprint1} · resuelto en sprint {sprint1} · Defecto"},
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

// pdfTextObject es el texto de una línea escrita en el PDF: BT x y Td (…)Tj ET.
var pdfTextObject = regexp.MustCompile(`Td \(((?:\\.|[^\\)])*)\)Tj`)

// pdfContent devuelve el texto de las líneas del PDF unidas por un espacio.
// Una línea más larga que el ancho de la página sigue en la siguiente, cortada
// en un espacio, así que unirlas así devuelve la línea original (HU-52: los
// UUID alargan las líneas).
func pdfContent(pdf []byte) []byte {
	var parts [][]byte
	for _, m := range pdfTextObject.FindAllSubmatch(pdf, -1) {
		parts = append(parts, m[1])
	}
	return bytes.Join(parts, []byte(" "))
}

// expectPDFContains falla si el PDF no contiene alguno de los textos.
func expectPDFContains(t *testing.T, pdf []byte, texts ...string) {
	t.Helper()
	content := pdfContent(pdf)
	for _, text := range texts {
		if !bytes.Contains(content, pdfText(text)) {
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

// HU-40 · CA-40.3 · Verifica que las líneas del reporte en PDF coincidan con el ejemplo de TEC-03 de la SDD.
// CA-40.3 · Las líneas del ejemplo de TEC-03 de la SDD.
func TestProjectReportLines_EjemploDeTEC03(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.tec03Project(t))

	got := projectReportLines(rep)

	want := withPublicIDs(rep, tec03Lines)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("líneas =\n%v\nse esperaba\n%v", got, want)
	}
}

// placeholder es un marcador de las líneas esperadas: {historia1}, {sprint2}…
var placeholder = regexp.MustCompile(`\{(historia|sprint|esfuerzo|defecto)(\d+)\}`)

// withPublicIDs devuelve una copia de lines con cada marcador reemplazado por
// el UUID de esa entidad en el reporte: {historiaN} es la historia N,
// {sprintN} el sprint N, {esfuerzoN} el esfuerzo N y {defectoN} el defecto N,
// en el orden del reporte (HU-52, CA-52.3).
func withPublicIDs(rep ProjectReport, lines []pdfLine) []pdfLine {
	out := make([]pdfLine, len(lines))
	for i, l := range lines {
		l.Text = placeholder.ReplaceAllStringFunc(l.Text, func(m string) string {
			g := placeholder.FindStringSubmatch(m)
			n, _ := strconv.Atoi(g[2])
			switch g[1] {
			case "historia":
				return rep.Stories.Stories[n-1].PublicID.String()
			case "sprint":
				return rep.Sprints[n-1].PublicID.String()
			case "esfuerzo":
				return rep.Efforts.Efforts[n-1].PublicID.String()
			default:
				return rep.Defects.Defects[n-1].PublicID.String()
			}
		})
		out[i] = l
	}
	return out
}

// HU-52 · CA-52.3 · Verifica que el texto del PDF no muestre el ID entero de ninguna historia, esfuerzo, defecto ni sprint.
func TestProjectReportLines_SinIDsEnteros(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.tec03Project(t))
	integerID := regexp.MustCompile(`#\d+ ·|ítem \d+ ·|sprint \d+( ·|$)`)

	for _, l := range projectReportLines(rep) {
		if integerID.MatchString(l.Text) {
			t.Errorf("la línea %q muestra un ID entero", l.Text)
		}
	}
}

// HU-40 · CA-40.5 · Verifica que un proyecto sin datos tenga las secciones vacías y las métricas como "sin datos".
// CA-40.5 · Un proyecto sin datos tiene las secciones vacías y "sin datos".
func TestProjectReportLines_ProyectoSinDatos(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.project(t, "Proyecto vacío", ""))

	got := projectReportLines(rep)

	if !reflect.DeepEqual(got, emptyProjectLines) {
		t.Errorf("líneas =\n%v\nse esperaba\n%v", got, emptyProjectLines)
	}
}

// HU-49 · CA-49.6 · Verifica que la línea de cada historia muestre prioridad, estado, Story Points, horas y sprint, o "sin estimar", "sin horas estimadas" y "sin sprint".
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

	rep := generate(t, e, projectID)
	got := projectReportLines(rep)

	expectLines(t, got, texts(withPublicIDs(rep, []pdfLine{
		{Text: "#{historia1} · Registrar venta · ALTA · TERMINADO · 3 SP · 8.0 h estimadas · sprint {sprint1}"},
		{Text: "#{historia2} · Consultar stock · MEDIA · PENDIENTE · sin estimar · sin horas estimadas · sin sprint"},
		{Text: "#{historia3} · Emitir factura · BAJA · EN_CURSO · sin estimar · 2.3 h estimadas · sin sprint"},
		{Text: "Cantidad: 3 historia(s), 1 completada(s)"},
	}))...)
}

// HU-40 · CA-40.3 · Verifica que el PDF muestre la fecha de fin del proyecto cuando está registrada.
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

// texts devuelve el texto de cada línea.
func texts(lines []pdfLine) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.Text
	}
	return out
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

// HU-40 · RN-40.3 · Verifica que las métricas y el total se redondeen a un decimal como la API, sin -0.0.
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

// HU-40 · RN-40.3 · Verifica que las horas de cada esfuerzo se muestren tal como están guardadas.
// RN-40.3 · Las horas de cada esfuerzo van tal como están guardadas.
func TestProjectReportLines_HorasDeCadaEsfuerzoTalCual(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Horas", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, 0, backlog.StatusPending, s.ID)
	e.effort(t, s.ID, it.ID, "Ana", 1.5)
	e.effort(t, s.ID, it.ID, "Luis", 0.25)

	rep := generate(t, e, projectID)
	got := projectReportLines(rep)

	expectLines(t, got, texts(withPublicIDs(rep, []pdfLine{
		{Text: "#{esfuerzo1} · 2026-10-06 · Ana · Desarrollo · ítem {historia1} · sprint {sprint1} · 1.5 h"},
		{Text: "#{esfuerzo2} · 2026-10-06 · Luis · Desarrollo · ítem {historia1} · sprint {sprint1} · 0.25 h"},
		{Text: "Total: 1.8 horas"},
	}))...)
}

// HU-40 · CA-40.3 · Verifica que un defecto sin sprint de resolución lo indique en su línea.
// CA-40.3 · Un defecto sin sprint de resolución lo indica.
func TestProjectReportLines_DefectoSinSprintDeResolucion(t *testing.T) {
	e := newReportEnv()
	projectID := e.project(t, "Defectos", "")
	s := e.openSprint(t, projectID, "Sprint 1")
	it := e.item(t, projectID, "A", 0, 0, backlog.StatusPending, s.ID)
	e.saveDefect(t, it.ID, s.ID, defect.StatusOpen, 0)

	rep := generate(t, e, projectID)
	got := projectReportLines(rep)

	expectLines(t, got, texts(withPublicIDs(rep, []pdfLine{
		{Text: "#{defecto1} · ALTA · ABIERTO · ítem {historia1} · detectado en sprint {sprint1} · sin sprint de resolución · Defecto"},
		{Text: "Cantidad: 1 defecto(s)"},
	}))...)
}

// HU-40 · CA-40.1 · Verifica que se genere un PDF válido con cada línea del reporte y los títulos en negrita.
// CA-40.1 y CA-40.3 · El archivo es un PDF y contiene cada línea del
// reporte, con los títulos en negrita.
func TestWriteProjectPDF_GeneraUnPDFConCadaLineaDelReporte(t *testing.T) {
	e := newReportEnv()
	rep := generate(t, e, e.tec03Project(t))

	pdf := writePDF(t, rep)

	expectValidPDF(t, pdf)
	for _, l := range withPublicIDs(rep, tec03Lines) {
		expectPDFContains(t, pdf, l.Text)
	}
	for _, font := range []string{"/BaseFont /Helvetica-Bold", "/BaseFont /Helvetica"} {
		if !bytes.Contains(pdf, []byte(font)) {
			t.Errorf("el PDF no usa la fuente %q", font)
		}
	}
}

// HU-40 · CA-40.5 · Verifica que un proyecto sin datos genere un PDF, no un error.
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

// HU-40 · CA-40.7 · Verifica que las tildes, la ñ y los signos se vean, y que un carácter fuera de Windows-1252 se reemplace por "." sin impedir el PDF.
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

// HU-40 · RN-40.1 · Verifica que un texto largo siga en la línea siguiente y que muchos esfuerzos sumen páginas, sin perder contenido.
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

	rep := generate(t, e, projectID)
	pdf := writePDF(t, rep)

	expectValidPDF(t, pdf)
	expectPDFContains(t, pdf, "Total: 120.0 horas", "#"+rep.Efforts.Efforts[119].PublicID.String()+" · 2026-10-06 · Ana")
	if pages := bytes.Count(pdf, []byte("<</Type /Page\n")); pages < 2 {
		t.Errorf("el PDF tiene %d página(s), se esperaban al menos 2", pages)
	}
}

// HU-40 · CA-40.2 · Verifica que, si el destino no acepta la escritura, se devuelva ese error.
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

// HU-40 · CA-40.8 · Verifica que exportar a PDF no modifique el reporte recibido.
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
