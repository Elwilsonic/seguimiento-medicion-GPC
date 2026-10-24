package httpapi

import "net/http"

func (s *server) registerEffort(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, errorResponse{Error: "no implementado"})
}
