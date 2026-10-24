package httpapi

// Handlers del Product Backlog y sus ítems: HU-04 y HU-06 desde TEC-01,
// HU-05, HU-07, HU-08 y HU-09 desde TEC-04 (CA-TEC04.9), y HU-20 (CA-20.8).

import (
	"encoding/json"
	"net/http"
	"strconv"

	"seguimiento-medicion-gpc/internal/backlog"
)

type itemRequest struct {
	Titulo      string   `json:"titulo"`
	Descripcion string   `json:"descripcion"`
	Prioridad   string   `json:"prioridad"`
	Criterios   []string `json:"criterios"`
	// HorasEstimadas solo la usa la modificación del ítem (HU-07).
	HorasEstimadas float64 `json:"horas_estimadas"`
}

type statusRequest struct {
	Estado string `json:"estado"`
}

// storyPointsRequest lee story_points como json.Number para que un decimal,
// un número enorme, null o el campo ausente respondan con el error de la
// escala y no con el de JSON inválido (SDD de HU-20).
type storyPointsRequest struct {
	StoryPoints json.Number `json:"story_points"`
}

type itemResponse struct {
	ID          int      `json:"id"`
	ProyectoID  int      `json:"proyecto_id"`
	Titulo      string   `json:"titulo"`
	Descripcion string   `json:"descripcion"`
	Prioridad   string   `json:"prioridad"`
	Criterios   []string `json:"criterios"`
	Estado      string   `json:"estado"`
	StoryPoints int      `json:"story_points"`
	// TEC-04, CA-TEC04.9: horas estimadas (HU-07) y sprint asignado (HU-11).
	HorasEstimadas float64 `json:"horas_estimadas"`
	SprintID       int     `json:"sprint_id"`
}

// HU-04
func (s *server) createItem(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req itemRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	it, err := s.items.Create(projectID, req.Titulo, req.Descripcion, req.Prioridad, req.Criterios)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toItemResponse(it))
}

// HU-06 y, con ?orden=prioridad, HU-09. Otro valor o sin el parámetro, el
// listado queda ordenado por ID.
func (s *server) listItems(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	list := s.items.ListItems
	if r.URL.Query().Get("orden") == "prioridad" {
		list = s.items.ListItemsByPriority
	}
	items, err := list(projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	// La lista se arma siempre no nil, para que un backlog vacío sea [] y no null.
	resp := make([]itemResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, toItemResponse(it))
	}
	writeJSON(w, http.StatusOK, resp)
}

// HU-05
func (s *server) getItem(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	it, err := s.items.GetItem(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toItemResponse(it))
}

// HU-07
func (s *server) updateItem(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req itemRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	it, err := s.items.Update(id, req.Titulo, req.Descripcion, req.Prioridad, req.Criterios, req.HorasEstimadas)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toItemResponse(it))
}

// HU-08
func (s *server) changeItemStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req statusRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	it, err := s.items.ChangeStatus(id, req.Estado)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toItemResponse(it))
}

// HU-20
func (s *server) assignStoryPoints(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req storyPointsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	// Lo que no es un entero representable no puede estar en la escala.
	points, err := strconv.Atoi(req.StoryPoints.String())
	if err != nil {
		writeError(w, backlog.ErrInvalidStoryPoints)
		return
	}
	it, err := s.items.AssignStoryPoints(id, points)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toItemResponse(it))
}

func toItemResponse(it backlog.Item) itemResponse {
	return itemResponse{
		ID:             it.ID,
		ProyectoID:     it.ProjectID,
		Titulo:         it.Title,
		Descripcion:    it.Description,
		Prioridad:      string(it.Priority),
		Criterios:      it.AcceptanceCriteria,
		Estado:         string(it.Status),
		StoryPoints:    it.StoryPoints,
		HorasEstimadas: it.EstimatedHours,
		SprintID:       it.SprintID,
	}
}
