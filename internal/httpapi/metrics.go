package httpapi

import "net/http"

// sprintMetrics atiende GET /api/sprints/{id}/metricas (HU-32).
func (s *server) sprintMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// projectMetrics atiende GET /api/proyectos/{id}/metricas (HU-32).
func (s *server) projectMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
