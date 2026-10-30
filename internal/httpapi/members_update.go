package httpapi

import "net/http"

// Handlers de la modificación de integrantes (HU-16, CA-16.7).

// updateMember atiende PUT /api/integrantes/{id}. El orden de los errores es
// ID de la ruta, JSON y Manager.
func (s *server) updateMember(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req memberRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	updated, err := s.members.Update(id, req.Nombre, req.Rol)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, memberResponse{ID: updated.ID, ProyectoID: updated.ProjectID, Nombre: updated.Name, Rol: updated.Role})
}

// memberMethodNotAllowed responde 405 a GET /api/integrantes/{id}. Sin esta
// ruta, GET lo atendería la interfaz web con un 404. Los demás métodos ya
// reciben el 405 de http.ServeMux.
func memberMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Allow", http.MethodPut)
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}
