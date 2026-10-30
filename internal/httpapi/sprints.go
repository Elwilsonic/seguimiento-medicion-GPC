package httpapi

// Handlers de sprints (HU-10 a HU-14), expuestos desde TEC-04 (CA-TEC04.9).

import (
	"cmp"
	"net/http"
	"time"

	"seguimiento-medicion-gpc/internal/sprint"
)

// dateLayout es el formato de las fechas de un sprint: AAAA-MM-DD.
const dateLayout = "2006-01-02"

type sprintRequest struct {
	Nombre      string `json:"nombre"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
	Objetivo    string `json:"objetivo"`
}

type sprintResponse struct {
	ID          int    `json:"id"`
	ProyectoID  int    `json:"proyecto_id"`
	Nombre      string `json:"nombre"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
	Objetivo    string `json:"objetivo"`
	Estado      string `json:"estado"`
}

// HU-10
func (s *server) createSprint(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req sprintRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	// Las dos fechas se leen con el mismo formato; se informa el primer error.
	start, errStart := parseDate(req.FechaInicio)
	end, errEnd := parseDate(req.FechaFin)
	if err := cmp.Or(errStart, errEnd); err != nil {
		writeError(w, err)
		return
	}
	sp, err := s.sprints.Create(projectID, req.Nombre, start, end, req.Objetivo)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSprintResponse(sp))
}

// HU-14
func (s *server) listProjectSprints(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	sprints, err := s.sprints.History(projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	resp := make([]sprintResponse, 0, len(sprints))
	for _, sp := range sprints {
		resp = append(resp, toSprintResponse(sp))
	}
	writeJSON(w, http.StatusOK, resp)
}

// HU-11
func (s *server) assignItem(w http.ResponseWriter, r *http.Request) {
	sprintID, itemID, err := sprintItemIDs(r)
	if err != nil {
		writeError(w, err)
		return
	}
	it, err := s.sprints.AssignItem(sprintID, itemID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toItemResponse(it))
}

// HU-12
func (s *server) completeItem(w http.ResponseWriter, r *http.Request) {
	sprintID, itemID, err := sprintItemIDs(r)
	if err != nil {
		writeError(w, err)
		return
	}
	it, err := s.sprints.CompleteItem(sprintID, itemID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toItemResponse(it))
}

// HU-12
func (s *server) listCompletedItems(w http.ResponseWriter, r *http.Request) {
	sprintID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	items, err := s.sprints.ListCompletedItems(sprintID)
	if err != nil {
		writeError(w, err)
		return
	}
	resp := make([]itemResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, toItemResponse(it))
	}
	writeJSON(w, http.StatusOK, resp)
}

// HU-13
func (s *server) closeSprint(w http.ResponseWriter, r *http.Request) {
	sprintID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	sp, err := s.sprints.Close(sprintID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSprintResponse(sp))
}

// sprintItemIDs lee el {id} del sprint y el {itemId} del ítem de la ruta.
func sprintItemIDs(r *http.Request) (sprintID, itemID int, err error) {
	if sprintID, err = pathID(r); err != nil {
		return 0, 0, err
	}
	if itemID, err = pathInt(r, "itemId"); err != nil {
		return 0, 0, err
	}
	return sprintID, itemID, nil
}

// parseDate lee una fecha AAAA-MM-DD. Una fecha vacía es una fecha ausente
// (el valor zero), y la valida el Manager de sprints.
func parseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	date, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, errInvalidDate
	}
	return date, nil
}

func toSprintResponse(sp sprint.Sprint) sprintResponse {
	return sprintResponse{
		ID:          sp.ID,
		ProyectoID:  sp.ProjectID,
		Nombre:      sp.Name,
		FechaInicio: sp.StartDate.Format(dateLayout),
		FechaFin:    sp.EndDate.Format(dateLayout),
		Objetivo:    sp.Goal,
		Estado:      string(sp.Status),
	}
}
