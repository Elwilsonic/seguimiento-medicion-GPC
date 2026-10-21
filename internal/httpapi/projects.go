package httpapi

// Handlers de proyectos (HU-01, HU-02 y HU-03), expuestos desde TEC-01.

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/project"
)

type projectRequest struct {
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
}

type projectResponse struct {
	ID          int    `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
}

// HU-01
func (s *server) createProject(w http.ResponseWriter, r *http.Request) {
	var req projectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	p, err := s.projects.Create(req.Nombre, req.Descripcion)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toProjectResponse(p))
}

// HU-03
func (s *server) getProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.projects.Get(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

// HU-02
func (s *server) updateProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req projectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	p, err := s.projects.Update(id, req.Nombre, req.Descripcion, req.Estado)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

func toProjectResponse(p project.Project) projectResponse {
	return projectResponse{ID: p.ID, Nombre: p.Name, Descripcion: p.Description, Estado: string(p.Status)}
}
