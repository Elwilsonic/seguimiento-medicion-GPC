package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/dashboard"
)

// summaryResponse es la respuesta de GET /api/proyectos/{id}/resumen (HU-36).
// SprintActual tiene el formato de los sprints y es null si el proyecto no
// tiene un sprint abierto (RN-36.3).
type summaryResponse struct {
	// ProyectoID es el UUID del proyecto (HU-51, CA-51.3).
	ProyectoID      string          `json:"proyecto_id"`
	Nombre          string          `json:"nombre"`
	Estado          string          `json:"estado"`
	ItemsPendientes int             `json:"items_pendientes"`
	ItemsEnCurso    int             `json:"items_en_curso"`
	ItemsTerminados int             `json:"items_terminados"`
	SprintActual    *sprintResponse `json:"sprint_actual"`
}

// projectSummary atiende GET /api/proyectos/{id}/resumen (HU-36).
func (s *server) projectSummary(w http.ResponseWriter, r *http.Request) {
	id, publicID, err := s.projectIDFromPath(r)
	if err != nil {
		writeError(w, err)
		return
	}
	summary, err := s.dashboard.ProjectSummary(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummaryResponse(summary, publicID))
}

// toSummaryResponse arma el resumen con publicID, el UUID del proyecto
// (HU-51).
func toSummaryResponse(summary dashboard.Summary, publicID string) summaryResponse {
	resp := summaryResponse{
		ProyectoID:      publicID,
		Nombre:          summary.ProjectName,
		Estado:          string(summary.ProjectStatus),
		ItemsPendientes: summary.Pending,
		ItemsEnCurso:    summary.InProgress,
		ItemsTerminados: summary.Finished,
	}
	if summary.HasCurrentSprint {
		sp := toSprintResponse(summary.CurrentSprint, publicID)
		resp.SprintActual = &sp
	}
	return resp
}
