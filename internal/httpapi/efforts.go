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

// HU-52, CA-52.3: el esfuerzo, su sprint y su ítem se identifican con su
// UUID.
type effortResponse struct {
	ID         string  `json:"id"`
	SprintID   string  `json:"sprint_id"`
	ItemID     string  `json:"item_id"`
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
	ids := s.newPublicIDs()
	sprintID, itemID, err := s.sprintItemPublicIDs(r)
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
	s.writeResolved(w, http.StatusCreated, toEffortResponse(saved, ids), ids)
}

// listEffortsByItem atiende GET /api/items/{id}/esfuerzos (HU-27).
func (s *server) listEffortsByItem(w http.ResponseWriter, r *http.Request) {
	ids := s.newPublicIDs()
	itemID, err := s.itemIDFromPath(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	summary, err := s.efforts.ListByItem(itemID)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeResolved(w, http.StatusOK, toEffortListResponse(summary, ids), ids)
}

// listEffortsBySprint atiende GET /api/sprints/{id}/esfuerzos (HU-27).
func (s *server) listEffortsBySprint(w http.ResponseWriter, r *http.Request) {
	ids := s.newPublicIDs()
	sprintID, err := s.sprintIDFromPath(r)
	if err != nil {
		writeError(w, err)
		return
	}
	summary, err := s.efforts.ListBySprint(sprintID)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeResolved(w, http.StatusOK, toEffortListResponse(summary, ids), ids)
}

func toEffortListResponse(summary effort.Summary, ids *publicIDs) effortListResponse {
	efforts := make([]effortResponse, 0, len(summary.Efforts))
	for _, effort := range summary.Efforts {
		efforts = append(efforts, toEffortResponse(effort, ids))
	}
	return effortListResponse{
		Esfuerzos: efforts,
		// La API redondea solo la presentación del total (RN-27.8).
		TotalHoras: math.Round(summary.TotalHours*10) / 10,
	}
}

func toEffortResponse(e effort.Effort, ids *publicIDs) effortResponse {
	return effortResponse{
		ID:         e.PublicID.String(),
		SprintID:   ids.sprint(e.SprintID),
		ItemID:     ids.item(e.ItemID),
		Integrante: e.Member,
		Fecha:      e.Date.Format(dateLayout),
		Actividad:  e.Activity,
		Horas:      e.Hours,
	}
}

// effortHoursRequest es el cuerpo de PATCH /api/esfuerzos/{id}/horas (HU-28),
// con las horas en el mismo formato que el registro de HU-26.
type effortHoursRequest struct {
	Horas json.Number `json:"horas"`
}

// correctEffortHours atiende PATCH /api/esfuerzos/{id}/horas (HU-28). Las horas
// ausentes o no numéricas dan ErrInvalidHours, como en HU-26.
func (s *server) correctEffortHours(w http.ResponseWriter, r *http.Request) {
	ids := s.newPublicIDs()
	id, err := s.effortIDFromPath(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req effortHoursRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	hours, err := strconv.ParseFloat(string(req.Horas), 64)
	if err != nil {
		writeError(w, effort.ErrInvalidHours)
		return
	}

	corrected, err := s.efforts.CorrectHours(id, hours)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeResolved(w, http.StatusOK, toEffortResponse(corrected, ids), ids)
}
