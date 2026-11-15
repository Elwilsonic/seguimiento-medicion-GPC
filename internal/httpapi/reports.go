package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/report"
)

// projectReportResponse es la respuesta de GET /api/proyectos/{id}/reporte
// (HU-38). Cada parte tiene el formato del endpoint que ya existe: proyecto
// (HU-03), métricas (HU-35), historias (HU-05, sumadas en HU-49), esfuerzos
// (HU-27) y defectos (HU-31) (RN-38.8 y RN-49.10).
type projectReportResponse struct {
	Proyecto    projectResponse        `json:"proyecto"`
	Metricas    projectMetricsResponse `json:"metricas"`
	Historias   storyListResponse      `json:"historias"`
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

// storyListResponse son las historias de un reporte (HU-49), cada una con el
// formato de GET /api/items/{id}. La lista es [] y no null si no hay ninguna.
type storyListResponse struct {
	Historias   []itemResponse `json:"historias"`
	Cantidad    int            `json:"cantidad"`
	Completadas int            `json:"completadas"`
}

// toStoryListResponse convierte las historias del reporte con toItemResponse,
// la misma conversión que GET /api/items/{id} (RN-49.10).
// toStoryListResponse arma las historias del reporte con publicID, el UUID
// de su proyecto (HU-51).
func toStoryListResponse(st report.Stories, publicID string, ids *publicIDs) storyListResponse {
	stories := make([]itemResponse, 0, len(st.Stories))
	for _, it := range st.Stories {
		stories = append(stories, toItemResponse(it, publicID, ids))
	}
	return storyListResponse{Historias: stories, Cantidad: len(stories), Completadas: st.Completed}
}

// projectReport atiende GET /api/proyectos/{id}/reporte (HU-38).
func (s *server) projectReport(w http.ResponseWriter, r *http.Request) {
	ids := s.newPublicIDs()
	id, _, err := s.projectIDFromPath(r)
	if err != nil {
		writeError(w, err)
		return
	}
	rep, err := s.reports.Project(id)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeResolved(w, http.StatusOK, toProjectReportResponse(rep, ids), ids)
}

// toProjectReportResponse arma el reporte; el UUID del proyecto sale del
// proyecto del reporte (HU-51).
func toProjectReportResponse(rep report.ProjectReport, ids *publicIDs) projectReportResponse {
	publicID := rep.Project.PublicID.String()
	return projectReportResponse{
		Proyecto:  toProjectResponse(rep.Project),
		Metricas:  toProjectMetricsResponse(rep.Metrics, publicID),
		Historias: toStoryListResponse(rep.Stories, publicID, ids),
		// toEffortListResponse recibe un effort.Summary: así "esfuerzos" tiene
		// el formato de GET /api/sprints/{id}/esfuerzos, con el total
		// redondeado a un decimal (RN-38.9).
		Esfuerzos:   toEffortListResponse(effort.Summary{Efforts: rep.Efforts.Efforts, TotalHours: rep.Efforts.TotalHours}, ids),
		Defectos:    toDefectListResponse(rep.Defects.Defects, ids),
		Integrantes: toMemberListResponse(rep.Members),
	}
}
