package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/member"
)

// Handler del registro de integrantes (HU-15, CA-15.7).

type memberRequest struct {
	Nombre string `json:"nombre"`
	Rol    string `json:"rol"`
}

type memberResponse struct {
	ID         int         `json:"id"`
	ProyectoID int         `json:"proyecto_id"`
	Nombre     string      `json:"nombre"`
	Rol        member.Role `json:"rol"`
}

// registerMember atiende POST /api/proyectos/{id}/integrantes. El orden de
// los errores es ID de la ruta, JSON y Manager.
func (s *server) registerMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req memberRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	saved, err := s.members.Register(projectID, req.Nombre, req.Rol)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, memberResponse{ID: saved.ID, ProyectoID: saved.ProjectID, Nombre: saved.Name, Rol: saved.Role})
}
