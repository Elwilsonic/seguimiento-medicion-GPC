package httpapi

// Handlers de proyectos (HU-01, HU-02 y HU-03), expuestos desde TEC-01, y
// fechas de inicio (HU-18) y de fin (HU-41).

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/project"
)

type projectRequest struct {
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
}

// startDateRequest es el cuerpo de PATCH /api/proyectos/{id}/fecha-inicio
// (HU-18).
type startDateRequest struct {
	FechaInicio string `json:"fecha_inicio"`
}

// endDateRequest es el cuerpo de PATCH /api/proyectos/{id}/fecha-fin
// (HU-41).
type endDateRequest struct {
	FechaFin string `json:"fecha_fin"`
}

type projectResponse struct {
	ID          int    `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
	// FechaInicio va en AAAA-MM-DD, o "" si no está registrada (HU-18).
	FechaInicio string `json:"fecha_inicio"`
	// FechaFin va en AAAA-MM-DD, o "" si no está registrada (HU-41).
	FechaFin string `json:"fecha_fin"`
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

// HU-18. La fecha se lee con el formato de las fechas de los sprints; una
// fecha vacía o ausente llega al Manager en cero (ErrMissingStartDate).
func (s *server) setProjectStartDate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req startDateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	date, err := parseDate(req.FechaInicio)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.projects.SetStartDate(id, date)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

// HU-41. Mismo formato que la fecha de inicio; una fecha vacía o ausente
// llega al Manager en cero (ErrMissingEndDate).
func (s *server) setProjectEndDate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req endDateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	date, err := parseDate(req.FechaFin)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.projects.SetEndDate(id, date)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

func toProjectResponse(p project.Project) projectResponse {
	resp := projectResponse{ID: p.ID, Nombre: p.Name, Descripcion: p.Description, Estado: string(p.Status)}
	// El cero no se formatea: saldría 0001-01-01 (RN-18.7).
	if !p.StartDate.IsZero() {
		resp.FechaInicio = p.StartDate.Format(dateLayout)
	}
	if !p.EndDate.IsZero() {
		resp.FechaFin = p.EndDate.Format(dateLayout)
	}
	return resp
}
