package httpapi

import (
	"math"
	"net/http"

	"seguimiento-medicion-gpc/internal/metrics"
)

// sprintMetricsResponse es la respuesta de GET /api/sprints/{id}/metricas
// (HU-32).
type sprintMetricsResponse struct {
	SprintID        int `json:"sprint_id"`
	SPPlanificados  int `json:"sp_planificados"`
	SPCompletados   int `json:"sp_completados"`
	ItemsSinEstimar int `json:"items_sin_estimar"`
}

// projectMetricsResponse es la respuesta de GET /api/proyectos/{id}/metricas
// (HU-32). Velocidad es null si no hay sprints cerrados (RN-32.6).
type projectMetricsResponse struct {
	ProyectoID      int      `json:"proyecto_id"`
	SPPlanificados  int      `json:"sp_planificados"`
	SPCompletados   int      `json:"sp_completados"`
	ItemsSinEstimar int      `json:"items_sin_estimar"`
	Velocidad       *float64 `json:"velocidad"`
	SprintsCerrados int      `json:"sprints_cerrados"`
}

// sprintMetrics atiende GET /api/sprints/{id}/metricas (HU-32).
func (s *server) sprintMetrics(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.metrics.SprintStoryPoints(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sprintMetricsResponse{
		SprintID:        p.SprintID,
		SPPlanificados:  p.Planned,
		SPCompletados:   p.Completed,
		ItemsSinEstimar: p.Unestimated,
	})
}

// projectMetrics atiende GET /api/proyectos/{id}/metricas (HU-32).
func (s *server) projectMetrics(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.metrics.ProjectStoryPoints(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectMetricsResponse(p))
}

// toProjectMetricsResponse redondea la velocidad a un decimal (RN-32.7) y la
// deja en null si no está disponible (RN-32.6).
func toProjectMetricsResponse(p metrics.ProjectPoints) projectMetricsResponse {
	resp := projectMetricsResponse{
		ProyectoID:      p.ProjectID,
		SPPlanificados:  p.Planned,
		SPCompletados:   p.Completed,
		ItemsSinEstimar: p.Unestimated,
		SprintsCerrados: p.ClosedSprints,
	}
	if p.VelocityAvailable {
		v := math.Round(p.Velocity*10) / 10
		resp.Velocidad = &v
	}
	return resp
}
