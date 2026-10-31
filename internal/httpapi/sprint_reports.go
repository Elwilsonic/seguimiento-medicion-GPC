package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/report"
)

// sprintReportResponse es la respuesta de GET /api/sprints/{id}/reporte
// (HU-39). Cada parte tiene el formato de una respuesta que ya existe: sprint
// (HU-10), métricas (HU-35), historias (HU-49), esfuerzos (HU-27), defectos
// (HU-31) e integrantes (HU-38) (RN-39.9).
type sprintReportResponse struct {
	Sprint      sprintResponse        `json:"sprint"`
	Metricas    sprintMetricsResponse `json:"metricas"`
	Historias   storyListResponse     `json:"historias"`
	Esfuerzos   effortListResponse    `json:"esfuerzos"`
	Defectos    defectListResponse    `json:"defectos"`
	Integrantes memberListResponse    `json:"integrantes"`
}

// sprintReport atiende GET /api/sprints/{id}/reporte (HU-39).
func (s *server) sprintReport(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	rep, err := s.reports.Sprint(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSprintReportResponse(rep))
}

func toSprintReportResponse(rep report.SprintReport) sprintReportResponse {
	return sprintReportResponse{
		Sprint: toSprintResponse(rep.Sprint),
		// La misma conversión que GET /api/sprints/{id}/metricas (decisión 4).
		Metricas:    toSprintMetricsResponse(rep.Metrics),
		Historias:   toStoryListResponse(rep.Stories),
		Esfuerzos:   toEffortListResponse(effort.Summary{Efforts: rep.Efforts.Efforts, TotalHours: rep.Efforts.TotalHours}),
		Defectos:    toDefectListResponse(rep.Defects.Defects),
		Integrantes: toMemberListResponse(rep.Members),
	}
}

// toMemberListResponse copia los integrantes del reporte en una lista nueva,
// [] y no null si no hay ninguno (RN-39.6).
func toMemberListResponse(m report.Members) memberListResponse {
	names := make([]string, 0, len(m.Names))
	names = append(names, m.Names...)
	return memberListResponse{Integrantes: names, Cantidad: len(names)}
}
