package httpapi

import (
	"bytes"
	"fmt"
	"net/http"

	"seguimiento-medicion-gpc/internal/report"
)

// projectReportPDF atiende GET /api/proyectos/{id}/reporte/pdf (HU-40). El PDF
// se arma completo en memoria antes de responder: si falla, responde 500 en
// JSON y nunca un PDF a medias (RN-40.7). El proyecto lo valida
// report.Manager.Project (RN-40.8).
func (s *server) projectReportPDF(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	rep, err := s.reports.Project(id)
	if err != nil {
		writeError(w, err)
		return
	}
	var buf bytes.Buffer
	if err := report.WriteProjectPDF(&buf, rep); err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="reporte-proyecto-%d.pdf"`, id))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
