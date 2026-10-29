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

// sprintMetrics atiende GET /api/sprints/{id}/metricas (HU-32) con una sola
// consulta de todas las métricas del sprint (HU-35, RN-35.1).
func (s *server) sprintMetrics(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	m, err := s.metrics.Sprint(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSprintMetricsResponse(m))
}

// toSprintMetricsResponse arma la respuesta de las métricas de un sprint. La
// usan GET /api/sprints/{id}/metricas (HU-35) y cada elemento de
// GET /api/proyectos/{id}/metricas/sprints (HU-37).
func toSprintMetricsResponse(m metrics.SprintMetrics) sprintMetricsResponse {
	return sprintMetricsResponse{
		SprintID:           m.SprintID,
		SPPlanificados:     m.Points.Planned,
		SPCompletados:      m.Points.Completed,
		ItemsSinEstimar:    m.Points.Unestimated,
		completionResponse: toCompletionResponse(m.Completion),
		hoursResponse:      toHoursResponse(m.Hours),
		defectsResponse:    toDefectsResponse(m.Defects),
	}
}

// sprintWithMetricsResponse es un elemento de
// GET /api/proyectos/{id}/metricas/sprints (HU-37): el sprint con el formato
// de HU-10 y sus métricas con el de HU-35, sin campos nuevos.
type sprintWithMetricsResponse struct {
	Sprint   sprintResponse        `json:"sprint"`
	Metricas sprintMetricsResponse `json:"metricas"`
}

// projectSprintsMetrics atiende GET /api/proyectos/{id}/metricas/sprints
// (HU-37, CA-37.4). Un proyecto sin sprints devuelve [], no null.
func (s *server) projectSprintsMetrics(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	list, err := s.metrics.ProjectSprints(id)
	if err != nil {
		writeError(w, err)
		return
	}
	resp := make([]sprintWithMetricsResponse, 0, len(list))
	for _, e := range list {
		resp = append(resp, sprintWithMetricsResponse{Sprint: toSprintResponse(e.Sprint), Metricas: toSprintMetricsResponse(e.Metrics)})
	}
	writeJSON(w, http.StatusOK, resp)
}

// projectMetrics atiende GET /api/proyectos/{id}/metricas (HU-32) con una
// sola consulta de todas las métricas del proyecto (HU-35, RN-35.1).
func (s *server) projectMetrics(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	m, err := s.metrics.Project(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projectMetricsResponse{
		ProyectoID:         m.ProjectID,
		SPPlanificados:     m.Points.Planned,
		SPCompletados:      m.Points.Completed,
		ItemsSinEstimar:    m.Points.Unestimated,
		Velocidad:          optionalOneDecimal(m.Points.Velocity, m.Points.VelocityAvailable),
		SprintsCerrados:    m.Points.ClosedSprints,
		completionResponse: toCompletionResponse(m.Completion),
		hoursResponse:      toHoursResponse(m.Hours),
		defectsResponse:    toDefectsResponse(m.Defects),
	})
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
	return hoursResponse{
		HorasEstimadas:         roundOneDecimal(h.Estimated),
		HorasReales:            roundOneDecimal(h.Actual),
		DesviacionHoras:        optionalOneDecimal(h.Deviation, h.DeviationAvailable),
		DesviacionPorcentual:   optionalOneDecimal(h.DeviationPercent, h.DeviationAvailable),
		ItemsSinHorasEstimadas: h.Unestimated,
	}
}

// optionalOneDecimal es la única conversión de las métricas que pueden no
// estar disponibles (HU-35, RN-35.6): velocidad, porcentaje de completadas y
// las dos desviaciones. Devuelve null si no está disponible y, si lo está, el
// valor redondeado a un decimal sin cero negativo.
func optionalOneDecimal(value float64, available bool) *float64 {
	if !available {
		return nil
	}
	rounded := roundOneDecimal(value)
	return &rounded
}

func roundOneDecimal(value float64) float64 {
	rounded := math.Round(value*10) / 10
	if rounded == 0 {
		return 0
	}
	return rounded
}

// toCompletionResponse redondea M8 a un decimal (RN-34.7) y lo deja en null
// si no hay ítems (RN-34.6).
func toCompletionResponse(c metrics.Completion) completionResponse {
	return completionResponse{
		ItemsTotales:          c.Items,
		ItemsTerminados:       c.Finished,
		PorcentajeCompletadas: optionalOneDecimal(c.Percent, c.PercentAvailable),
	}
}
