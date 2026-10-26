package httpapi

import (
	"math"
	"net/http"

	"seguimiento-medicion-gpc/internal/metrics"
)

// sprintMetricsResponse es la respuesta de GET /api/sprints/{id}/metricas
// (HU-32 y HU-34).
type sprintMetricsResponse struct {
	SprintID        int `json:"sprint_id"`
	SPPlanificados  int `json:"sp_planificados"`
	SPCompletados   int `json:"sp_completados"`
	ItemsSinEstimar int `json:"items_sin_estimar"`
	completionResponse
}

// projectMetricsResponse es la respuesta de GET /api/proyectos/{id}/metricas
// (HU-32 y HU-34). Velocidad es null si no hay sprints cerrados (RN-32.6).
type projectMetricsResponse struct {
	ProyectoID      int      `json:"proyecto_id"`
	SPPlanificados  int      `json:"sp_planificados"`
	SPCompletados   int      `json:"sp_completados"`
	ItemsSinEstimar int      `json:"items_sin_estimar"`
	Velocidad       *float64 `json:"velocidad"`
	SprintsCerrados int      `json:"sprints_cerrados"`
	completionResponse
}

// completionResponse son los campos de M8 (HU-34). PorcentajeCompletadas es
// null si no hay ítems (RN-34.6).
type completionResponse struct {
	ItemsTotales          int      `json:"items_totales"`
	ItemsTerminados       int      `json:"items_terminados"`
	PorcentajeCompletadas *float64 `json:"porcentaje_completadas"`
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
	c, err := s.metrics.SprintCompletion(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sprintMetricsResponse{
		SprintID:           p.SprintID,
		SPPlanificados:     p.Planned,
		SPCompletados:      p.Completed,
		ItemsSinEstimar:    p.Unestimated,
		completionResponse: toCompletionResponse(c),
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
	c, err := s.metrics.ProjectCompletion(id)
	if err != nil {
		writeError(w, err)
		return
	}
	resp := toProjectMetricsResponse(p)
	resp.completionResponse = toCompletionResponse(c)
	writeJSON(w, http.StatusOK, resp)
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

// toCompletionResponse redondea M8 a un decimal (RN-34.7) y lo deja en null
// si no hay ítems (RN-34.6).
func toCompletionResponse(c metrics.Completion) completionResponse {
	resp := completionResponse{ItemsTotales: c.Items, ItemsTerminados: c.Finished}
	if c.PercentAvailable {
		p := math.Round(c.Percent*10) / 10
		resp.PorcentajeCompletadas = &p
	}
	return resp
}
