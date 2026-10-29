package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/metrics"
	"seguimiento-medicion-gpc/internal/report"
)

// projectReportResponse es la respuesta de GET /api/proyectos/{id}/reporte
// (HU-38). Cada parte tiene el formato del endpoint que ya existe: proyecto
// (HU-03), métricas (HU-35), esfuerzos (HU-27) y defectos (HU-31) (RN-38.8).
type projectReportResponse struct {
	Proyecto    projectResponse        `json:"proyecto"`
	Metricas    projectMetricsResponse `json:"metricas"`
	Esfuerzos   effortListResponse     `json:"esfuerzos"`
	Defectos    defectListResponse     `json:"defectos"`
	Integrantes memberListResponse     `json:"integrantes"`
}

// memberListResponse son los integrantes que registraron esfuerzo en el
// proyecto (RN-38.6). La lista es [] y no null si no hay ninguno.
type memberListResponse struct {
	Integrantes []string `json:"integrantes"`
	Cantidad    int      `json:"cantidad"`
}

// projectReport atiende GET /api/proyectos/{id}/reporte (HU-38).
func (s *server) projectReport(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, toProjectReportResponse(rep))
}

func toProjectReportResponse(rep report.ProjectReport) projectReportResponse {
	return projectReportResponse{
		Proyecto: toProjectResponse(rep.Project),
		Metricas: toReportMetricsResponse(rep.Metrics),
		// toEffortListResponse recibe un effort.Summary: así "esfuerzos" tiene
		// el formato de GET /api/sprints/{id}/esfuerzos, con el total
		// redondeado a un decimal (RN-38.9).
		Esfuerzos:   toEffortListResponse(effort.Summary{Efforts: rep.Efforts.Efforts, TotalHours: rep.Efforts.TotalHours}),
		Defectos:    toDefectListResponse(rep.Defects.Defects),
		Integrantes: toMemberListResponse(rep.Members),
	}
}

// toReportMetricsResponse arma "metricas" igual que el handler projectMetrics
// de HU-35. Duplica esa conversión a propósito, para no modificar metrics.go
// mientras HU-37 lo cambia; se unifica cuando las dos historias estén en main
// (límite conocido de la SDD de HU-38).
func toReportMetricsResponse(m metrics.ProjectMetrics) projectMetricsResponse {
	return projectMetricsResponse{
		ProyectoID:         m.ProjectID,
		SPPlanificados:     m.Points.Planned,
		SPCompletados:      m.Points.Completed,
		ItemsSinEstimar:    m.Points.Unestimated,
		Velocidad:          optionalOneDecimal(m.Points.Velocity, m.Points.VelocityAvailable),
		SprintsCerrados:    m.Points.ClosedSprints,
		completionResponse: toCompletionResponse(m.Completion),
		hoursResponse:      toHoursResponse(m.Hours),
		defectsResponse:    toDefectsResponse(m.Defects),
	}
}
