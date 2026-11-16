package httpapi

import "net/http"

// Handler de la consulta de integrantes (HU-17, CA-17.1 a CA-17.4).

// listMembers atiende GET /api/proyectos/{id}/integrantes. El orden de los
// errores es ID de la ruta y Manager.
func (s *server) listMembers(w http.ResponseWriter, r *http.Request) {
	projectID, publicID, err := s.projectIDFromPath(r)
	if err != nil {
		writeError(w, err)
		return
	}

	members, err := s.members.List(projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	resp := make([]memberResponse, 0, len(members))
	for _, m := range members {
		resp = append(resp, memberResponse{ID: m.PublicID.String(), ProyectoID: publicID, Nombre: m.Name, Rol: m.Role})
	}
	writeJSON(w, http.StatusOK, resp)
}
