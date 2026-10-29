package report

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// Ejecuta los escenarios BDD de features/HU-40-exportar-reporte-pdf.feature
// con godog.
func TestFeatures_HU40ExportarReportePDF(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializePDFScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-40-exportar-reporte-pdf.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-40")
	}
}

// errDestination es el error del destino que no acepta la escritura.
var errDestination = errors.New("no se puede escribir el archivo")

// pdfDates son las fechas de creación y modificación que la librería guarda
// en los metadatos del PDF; cambian en cada exportación (límite conocido).
var pdfDates = regexp.MustCompile(`/(CreationDate|ModDate) \(D:\d+\)`)

// pdfScenario guarda el estado de un escenario de HU-40.
type pdfScenario struct {
	t        *testing.T
	env      *reportEnv
	projects map[string]int

	projectID int
	pdfs      [][]byte
	err       error
	before    storedData
	beforeP   project.Project
}

func initializePDFScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &pdfScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*s = pdfScenario{t: t, env: newReportEnv(), projects: map[string]int{}}
		return ctx, nil
	})

	ctx.Step(`^que existe el proyecto "([^"]*)" del ejemplo de TEC-03$`, s.tec03Project)
	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que el proyecto "([^"]*)" tiene un sprint abierto con los esfuerzos:$`, s.sprintWithEfforts)
	ctx.Step(`^exporto a PDF el reporte del proyecto "([^"]*)"$`, s.exportByName)
	ctx.Step(`^exporto a PDF el reporte del proyecto "([^"]*)" en un destino que falla$`, s.exportToFailingDestination)
	ctx.Step(`^exporto a PDF el reporte del proyecto con el ID (-?\d+)$`, s.exportByID)
	ctx.Step(`^se genera un archivo PDF$`, s.pdfGenerated)
	ctx.Step(`^el PDF contiene las líneas:$`, s.pdfContainsLines)
	ctx.Step(`^el PDF contiene cada línea del reporte del proyecto$`, s.pdfContainsEveryLine)
	ctx.Step(`^la exportación falla con el error del destino$`, s.failsWithDestinationError)
	ctx.Step(`^la exportación falla porque el proyecto no existe$`, s.failsNotFound)
	ctx.Step(`^no se escribió ningún PDF$`, s.nothingWritten)
	ctx.Step(`^los dos PDF tienen el mismo contenido$`, s.sameContent)
	ctx.Step(`^los datos guardados del proyecto no cambiaron$`, s.nothingChanged)
}

func (s *pdfScenario) tec03Project(name string) error {
	if name != "Software Metrics" {
		return fmt.Errorf("el proyecto del ejemplo de TEC-03 se llama %q", "Software Metrics")
	}
	s.projects[name] = s.env.tec03Project(s.t)
	return nil
}

func (s *pdfScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name, "")
	return nil
}

// sprintWithEfforts crea un sprint abierto con un ítem sin estimar y registra
// en él los esfuerzos de la tabla (integrante y horas).
func (s *pdfScenario) sprintWithEfforts(name string, table *godog.Table) error {
	projectID := s.projects[name]
	sp := s.env.openSprint(s.t, projectID, "Sprint 1")
	it := s.env.item(s.t, projectID, "A", 0, 0, backlog.StatusPending, sp.ID)
	for _, row := range table.Rows[1:] {
		hours, err := strconv.ParseFloat(row.Cells[1].Value, 64)
		if err != nil {
			return err
		}
		s.env.effort(s.t, sp.ID, it.ID, row.Cells[0].Value, hours)
	}
	return nil
}

// stored lee el proyecto y sus datos directo de los repositorios.
func (s *pdfScenario) stored() (project.Project, storedData) {
	var data storedData
	p, _ := s.env.projectSpy.Repository.GetByID(s.projectID)
	data.items, _ = s.env.itemSpy.Repository.ListByProject(s.projectID)
	data.sprints, _ = s.env.sprintSpy.Repository.ListByProject(s.projectID)
	for _, sp := range data.sprints {
		found, _ := s.env.effortSpy.Repository.ListBySprint(sp.ID)
		data.efforts = append(data.efforts, found...)
	}
	data.defects, _ = s.env.defectSpy.Repository.List()
	return p, data
}

// export arma el reporte del proyecto y lo escribe en w (RN-40.8: un proyecto
// inexistente falla antes de escribir).
func (s *pdfScenario) export(id int, w io.Writer) {
	if s.projectID == 0 {
		s.projectID = id
		s.beforeP, s.before = s.stored()
	}
	rep, err := s.env.reports.Project(id)
	if err != nil {
		s.err = err
		return
	}
	s.err = WriteProjectPDF(w, rep)
}

func (s *pdfScenario) exportByID(id int) error {
	var buf bytes.Buffer
	s.export(id, &buf)
	s.pdfs = append(s.pdfs, buf.Bytes())
	return nil
}

func (s *pdfScenario) exportByName(name string) error {
	return s.exportByID(s.projects[name])
}

func (s *pdfScenario) exportToFailingDestination(name string) error {
	s.export(s.projects[name], failingWriter{err: errDestination})
	return nil
}

// lastPDF devuelve el último PDF exportado.
func (s *pdfScenario) lastPDF() ([]byte, error) {
	if len(s.pdfs) == 0 {
		return nil, errors.New("no se exportó ningún PDF")
	}
	return s.pdfs[len(s.pdfs)-1], nil
}

func (s *pdfScenario) pdfGenerated() error {
	if s.err != nil {
		return fmt.Errorf("la exportación falló: %v", s.err)
	}
	pdf, err := s.lastPDF()
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) || !bytes.HasSuffix(bytes.TrimSpace(pdf), []byte("%%EOF")) {
		return fmt.Errorf("el archivo no es un PDF: empieza con %q", firstBytes(pdf))
	}
	return nil
}

func (s *pdfScenario) containsAll(texts []string) error {
	pdf, err := s.lastPDF()
	if err != nil {
		return err
	}
	for _, text := range texts {
		if !bytes.Contains(pdf, pdfText(text)) {
			return fmt.Errorf("el PDF no contiene %q", text)
		}
	}
	return nil
}

func (s *pdfScenario) pdfContainsLines(table *godog.Table) error {
	texts := make([]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		texts = append(texts, row.Cells[0].Value)
	}
	return s.containsAll(texts)
}

// pdfContainsEveryLine verifica que el PDF tiene cada línea del contenido del
// reporte actual (CA-40.3).
func (s *pdfScenario) pdfContainsEveryLine() error {
	rep, err := s.env.reports.Project(s.projectID)
	if err != nil {
		return err
	}
	var texts []string
	for _, l := range projectReportLines(rep) {
		texts = append(texts, l.Text)
	}
	if len(texts) == 0 {
		return errors.New("el reporte no tiene líneas")
	}
	return s.containsAll(texts)
}

func (s *pdfScenario) failsWithDestinationError() error {
	if !errors.Is(s.err, errDestination) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, errDestination)
	}
	return nil
}

func (s *pdfScenario) failsNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	return nil
}

func (s *pdfScenario) nothingWritten() error {
	pdf, err := s.lastPDF()
	if err != nil {
		return err
	}
	if len(pdf) != 0 {
		return fmt.Errorf("se escribieron %d bytes", len(pdf))
	}
	return nil
}

// sameContent compara los dos PDF sin las fechas de los metadatos.
func (s *pdfScenario) sameContent() error {
	if s.err != nil {
		return fmt.Errorf("la exportación falló: %v", s.err)
	}
	if len(s.pdfs) != 2 {
		return fmt.Errorf("se exportaron %d PDF, se esperaban 2", len(s.pdfs))
	}
	first := pdfDates.ReplaceAll(s.pdfs[0], nil)
	second := pdfDates.ReplaceAll(s.pdfs[1], nil)
	if len(first) == 0 || !bytes.Equal(first, second) {
		return errors.New("los dos PDF tienen distinto contenido")
	}
	return nil
}

func (s *pdfScenario) nothingChanged() error {
	p, after := s.stored()
	if p != s.beforeP || !reflect.DeepEqual(after, s.before) {
		return errors.New("exportar modificó el proyecto, los ítems, los sprints, los esfuerzos o los defectos")
	}
	return nil
}
