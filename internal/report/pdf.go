package report

import (
	"errors"
	"io"
)

// errNotImplemented es el resultado de las firmas vacías del RED de HU-40.
var errNotImplemented = errors.New("HU-40 todavía no está implementada")

// pdfLine es una línea de texto del PDF del reporte (HU-40).
type pdfLine struct {
	Text    string
	Heading bool
}

// projectReportLines arma las líneas del PDF del reporte de un proyecto.
func projectReportLines(rep ProjectReport) []pdfLine {
	return nil
}

// WriteProjectPDF escribe en w el PDF del reporte de un proyecto (HU-40).
func WriteProjectPDF(w io.Writer, rep ProjectReport) error {
	return errNotImplemented
}
