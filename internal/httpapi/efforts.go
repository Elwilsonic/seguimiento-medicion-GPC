package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"seguimiento-medicion-gpc/internal/effort"
)

// Handler del registro de esfuerzo (HU-26, CA-26.11).

type effortRequest struct {
	Integrante string      `json:"integrante"`
	Fecha      string      `json:"fecha"`
	Actividad  string      `json:"actividad"`
	Horas      json.Number `json:"horas"`
}

type effortResponse struct {
	ID         int     `json:"id"`
	SprintID   int     `json:"sprint_id"`
	ItemID     int     `json:"item_id"`
	Integrante string  `json:"integrante"`
	Fecha      string  `json:"fecha"`
	Actividad  string  `json:"actividad"`
	Horas      float64 `json:"horas"`
}

// registerEffort atiende POST /api/sprints/{id}/items/{itemId}/esfuerzos. El
// orden de los errores es IDs de la ruta, JSON, fecha, horas y Manager.
func (s *server) registerEffort(w http.ResponseWriter, r *http.Request) {
	sprintID, itemID, err := sprintItemIDs(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req effortRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	date, err := parseDate(req.Fecha)
	if err != nil {
		writeError(w, err)
		return
	}

	hours, err := strconv.ParseFloat(string(req.Horas), 64)
	if err != nil {
		writeError(w, effort.ErrInvalidHours)
		return
	}

	saved, err := s.efforts.Register(sprintID, itemID, req.Integrante, date, req.Actividad, hours)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toEffortResponse(saved))
}

func toEffortResponse(e effort.Effort) effortResponse {
	return effortResponse{
		ID:         e.ID,
		SprintID:   e.SprintID,
		ItemID:     e.ItemID,
		Integrante: e.Member,
		Fecha:      e.Date.Format(dateLayout),
		Actividad:  e.Activity,
		Horas:      e.Hours,
	}
}
