package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/dashboard"
)

// summaryResponse es la respuesta de GET /api/proyectos/{id}/resumen (HU-36).
// SprintActual tiene el formato de los sprints y es null si el proyecto no
// tiene un sprint abierto (RN-36.3).
type summaryResponse struct {
	ProyectoID      int             `json:"proyecto_id"`
	Nombre          string          `json:"nombre"`
	Estado          string          `json:"estado"`
	ItemsPendientes int             `json:"items_pendientes"`
	ItemsEnCurso    int             `json:"items_en_curso"`
	ItemsTerminados int             `json:"items_terminados"`
	SprintActual    *sprintResponse `json:"sprint_actual"`
}

// projectSummary atiende GET /api/proyectos/{id}/resumen (HU-36).
func (s *server) projectSummary(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	summary, err := s.dashboard.ProjectSummary(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummaryResponse(summary))
}

func toSummaryResponse(summary dashboard.Summary) summaryResponse {
	resp := summaryResponse{
		ProyectoID:      summary.ProjectID,
		Nombre:          summary.ProjectName,
		Estado:          string(summary.ProjectStatus),
		ItemsPendientes: summary.Pending,
		ItemsEnCurso:    summary.InProgress,
		ItemsTerminados: summary.Finished,
	}
	if summary.HasCurrentSprint {
		sp := toSprintResponse(summary.CurrentSprint)
		resp.SprintActual = &sp
	}
	return resp
}
