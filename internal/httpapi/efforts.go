package httpapi

import (
	"encoding/json"
	"math"
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

type effortListResponse struct {
	Esfuerzos  []effortResponse `json:"esfuerzos"`
	TotalHoras float64          `json:"total_horas"`
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

// listEffortsByItem atiende GET /api/items/{id}/esfuerzos (HU-27).
func (s *server) listEffortsByItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	summary, err := s.efforts.ListByItem(itemID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toEffortListResponse(summary))
}

// listEffortsBySprint atiende GET /api/sprints/{id}/esfuerzos (HU-27).
func (s *server) listEffortsBySprint(w http.ResponseWriter, r *http.Request) {
	sprintID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	summary, err := s.efforts.ListBySprint(sprintID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toEffortListResponse(summary))
}

func toEffortListResponse(summary effort.Summary) effortListResponse {
	efforts := make([]effortResponse, 0, len(summary.Efforts))
	for _, effort := range summary.Efforts {
		efforts = append(efforts, toEffortResponse(effort))
	}
	return effortListResponse{
		Esfuerzos: efforts,
		// La API redondea solo la presentación del total (RN-27.8).
		TotalHoras: math.Round(summary.TotalHours*10) / 10,
	}
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
