package report

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
)

// pdfLine es una línea de texto del PDF del reporte (HU-40). Los títulos van
// en negrita.
type pdfLine struct {
	Text    string
	Heading bool
}

// projectReportLines arma las líneas del PDF del reporte de un proyecto, con
// los mismos textos que la interfaz de HU-38 (RN-40.1 y decisión 8). No
// calcula nada: solo da formato a los valores del reporte (RN-40.2).
func projectReportLines(rep ProjectReport) []pdfLine {
	p, m := rep.Project, rep.Metrics
	lines := []pdfLine{
		{Text: "Reporte del proyecto «" + p.Name + "»", Heading: true},
		{Text: "Proyecto", Heading: true},
		{Text: "Nombre: " + p.Name},
		{Text: "Descripción: " + p.Description},
		{Text: "Estado: " + string(p.Status)},
		{Text: "Inicio: " + dateOrNotRegistered(p.StartDate)},
		{Text: "Fin: " + dateOrNotRegistered(p.EndDate)},
		{Text: "Métricas", Heading: true},
		{Text: fmt.Sprintf("Story points: %d planificados, %d completados, %d ítem(s) sin estimar.",
			m.Points.Planned, m.Points.Completed, m.Points.Unestimated)},
	}

	// RN-40.4: una métrica no disponible es "sin datos", nunca 0.
	velocity := "sin datos"
	if m.Points.VelocityAvailable {
		velocity = oneDecimal(m.Points.Velocity) + " SP por sprint"
	}
	lines = append(lines, pdfLine{Text: fmt.Sprintf("Velocidad: %s (%d sprint(s) cerrado(s)).", velocity, m.Points.ClosedSprints)})
	completion := "Historias completadas: sin datos."
	if m.Completion.PercentAvailable {
		completion = fmt.Sprintf("Historias completadas: %s %% (%d de %d).",
			oneDecimal(m.Completion.Percent), m.Completion.Finished, m.Completion.Items)
	}
	lines = append(lines,
		pdfLine{Text: completion},
		pdfLine{Text: fmt.Sprintf("Horas: %s estimadas, %s reales, %d ítem(s) sin horas estimadas.",
			oneDecimal(m.Hours.Estimated), oneDecimal(m.Hours.Actual), m.Hours.Unestimated)},
	)
	deviation := "Desviación: sin datos."
	if m.Hours.DeviationAvailable {
		deviation = fmt.Sprintf("Desviación: %s h (%s %%).", oneDecimal(m.Hours.Deviation), oneDecimal(m.Hours.DeviationPercent))
	}
	lines = append(lines,
		pdfLine{Text: deviation},
		pdfLine{Text: fmt.Sprintf("Defectos: %d detectados, %d resuelto(s).", m.Defects.Detected, m.Defects.Resolved)},
		pdfLine{Text: "Historias", Heading: true},
	)

	// HU-49: entre las métricas y los esfuerzos, en el orden del reporte.
	for _, it := range rep.Stories.Stories {
		lines = append(lines, pdfLine{Text: storyText(it)})
	}
	if len(rep.Stories.Stories) == 0 {
		lines = append(lines, pdfLine{Text: "Sin historias en el backlog."})
	}
	lines = append(lines,
		pdfLine{Text: fmt.Sprintf("Cantidad: %d historia(s), %d completada(s)", len(rep.Stories.Stories), rep.Stories.Completed)},
		pdfLine{Text: "Esfuerzos", Heading: true},
	)

	// RN-40.5: en el orden del reporte.
	for _, ef := range rep.Efforts.Efforts {
		lines = append(lines, pdfLine{Text: effortText(ef)})
	}
	if len(rep.Efforts.Efforts) == 0 {
		lines = append(lines, pdfLine{Text: "Sin esfuerzos registrados."})
	}
	lines = append(lines,
		pdfLine{Text: "Total: " + oneDecimal(rep.Efforts.TotalHours) + " horas"},
		pdfLine{Text: "Defectos", Heading: true},
	)
	for _, d := range rep.Defects.Defects {
		lines = append(lines, pdfLine{Text: defectText(d)})
	}
	if len(rep.Defects.Defects) == 0 {
		lines = append(lines, pdfLine{Text: "Sin defectos registrados."})
	}

	members := "ninguno"
	if len(rep.Members.Names) > 0 {
		members = strings.Join(rep.Members.Names, ", ")
	}
	return append(lines,
		pdfLine{Text: fmt.Sprintf("Cantidad: %d defecto(s)", len(rep.Defects.Defects))},
		pdfLine{Text: "Integrantes", Heading: true},
		pdfLine{Text: "Integrantes que registraron esfuerzo: " + members},
	)
}

// dateOrNotRegistered escribe una fecha en AAAA-MM-DD, o "sin registrar" si
// es el valor cero (RN-18.7 y RN-41.10).
func dateOrNotRegistered(date time.Time) string {
	if date.IsZero() {
		return "sin registrar"
	}
	return date.Format(time.DateOnly)
}

// effortText escribe un esfuerzo con sus horas tal como están guardadas
// (RN-40.3).
func effortText(ef effort.Effort) string {
	return fmt.Sprintf("#%d · %s · %s · %s · ítem %d · sprint %d · %s h",
		ef.ID, ef.Date.Format(time.DateOnly), ef.Member, ef.Activity, ef.ItemID, ef.SprintID,
		strconv.FormatFloat(ef.Hours, 'f', -1, 64))
}

// storyText escribe una historia con su estimación (HU-49). 0 Story Points
// es "sin estimar", 0 horas es "sin horas estimadas" y 0 en el sprint es
// "sin sprint" (RN-49.3); las horas van con un decimal, como las métricas.
func storyText(it backlog.Item) string {
	points := "sin estimar"
	if it.StoryPoints != 0 {
		points = fmt.Sprintf("%d SP", it.StoryPoints)
	}
	hours := "sin horas estimadas"
	if it.EstimatedHours != 0 {
		hours = oneDecimal(it.EstimatedHours) + " h estimadas"
	}
	sprintText := "sin sprint"
	if it.SprintID != 0 {
		sprintText = fmt.Sprintf("sprint %d", it.SprintID)
	}
	return fmt.Sprintf("#%d · %s · %s · %s · %s · %s · %s",
		it.ID, it.Title, it.Priority, it.Status, points, hours, sprintText)
}

// defectText escribe un defecto; 0 en el sprint de resolución es ninguno.
func defectText(d defect.Defect) string {
	resolution := "sin sprint de resolución"
	if d.ResolutionSprintID != 0 {
		resolution = fmt.Sprintf("resuelto en sprint %d", d.ResolutionSprintID)
	}
	return fmt.Sprintf("#%d · %s · %s · ítem %d · detectado en sprint %d · %s · %s",
		d.ID, d.Severity, d.Status, d.ItemID, d.DetectionSprintID, resolution, d.Description)
}

// oneDecimal redondea a un decimal como la API (TEC-03, roundOneDecimal de
// internal/httpapi), sin -0.0, y lo escribe con un decimal (RN-40.3).
func oneDecimal(value float64) string {
	rounded := math.Round(value*10) / 10
	if rounded == 0 {
		rounded = 0
	}
	return strconv.FormatFloat(rounded, 'f', 1, 64)
}

// WriteProjectPDF escribe en w el PDF del reporte de un proyecto (HU-40): A4
// vertical, sin comprimir, con Helvetica y una línea por cada pdfLine. El
// texto se pasa a Windows-1252 (RN-40.6). Devuelve el error de escritura si w
// no la acepta (RN-40.7). Es el único lugar que usa la librería fpdf.
func WriteProjectPDF(w io.Writer, rep ProjectReport) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(false)
	// Ordena los recursos internos (las fuentes): sin esto, el orden cambia
	// entre exportaciones y el mismo reporte da archivos distintos.
	pdf.SetCatalogSort(true)
	toCP1252 := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.AddPage()
	for _, l := range projectReportLines(rep) {
		if l.Heading {
			pdf.SetFont("Helvetica", "B", 13)
			pdf.MultiCell(0, 8, toCP1252(l.Text), "", "L", false)
			continue
		}
		pdf.SetFont("Helvetica", "", 11)
		pdf.MultiCell(0, 6, toCP1252(l.Text), "", "L", false)
	}
	return pdf.Output(w)
}
