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
	hoursResponse
	defectsResponse
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
	hoursResponse
	defectsResponse
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
	h, err := s.metrics.SprintHours(id)
	if err != nil {
		writeError(w, err)
		return
	}
	d, err := s.metrics.SprintDefects(id)
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
		hoursResponse:      toHoursResponse(h),
		defectsResponse:    toDefectsResponse(d),
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
	h, err := s.metrics.ProjectHours(id)
	if err != nil {
		writeError(w, err)
		return
	}
	d, err := s.metrics.ProjectDefects(id)
	if err != nil {
		writeError(w, err)
		return
	}
	resp := toProjectMetricsResponse(p)
	resp.completionResponse = toCompletionResponse(c)
	resp.hoursResponse = toHoursResponse(h)
	resp.defectsResponse = toDefectsResponse(d)
	writeJSON(w, http.StatusOK, resp)
}

// defectsResponse son los campos de M9 y M10 (HU-47), enteros (RN-47.8).
type defectsResponse struct {
	DefectosDetectados int `json:"defectos_detectados"`
	DefectosResueltos  int `json:"defectos_resueltos"`
}

func toDefectsResponse(d metrics.Defects) defectsResponse {
	return defectsResponse{DefectosDetectados: d.Detected, DefectosResueltos: d.Resolved}
}

type hoursResponse struct {
	HorasEstimadas         float64  `json:"horas_estimadas"`
	HorasReales            float64  `json:"horas_reales"`
	DesviacionHoras        *float64 `json:"desviacion_horas"`
	DesviacionPorcentual   *float64 `json:"desviacion_porcentual"`
	ItemsSinHorasEstimadas int      `json:"items_sin_horas_estimadas"`
}

// toHoursResponse redondea M4 a M7 a un decimal (RN-33.7 y RN-33.8) y deja las
// desviaciones en null cuando no están disponibles.
func toHoursResponse(h metrics.Hours) hoursResponse {
	resp := hoursResponse{
		HorasEstimadas:         roundOneDecimal(h.Estimated),
		HorasReales:            roundOneDecimal(h.Actual),
		ItemsSinHorasEstimadas: h.Unestimated,
	}
	if h.DeviationAvailable {
		deviation := roundOneDecimal(h.Deviation)
		deviationPercent := roundOneDecimal(h.DeviationPercent)
		resp.DesviacionHoras = &deviation
		resp.DesviacionPorcentual = &deviationPercent
	}
	return resp
}

func roundOneDecimal(value float64) float64 {
	rounded := math.Round(value*10) / 10
	if rounded == 0 {
		return 0
	}
	return rounded
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
