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
	ID string `json:"id"`
	// ProyectoID es el UUID del proyecto (HU-51, CA-51.3).
	ProyectoID  string `json:"proyecto_id"`
	Nombre      string `json:"nombre"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
	Objetivo    string `json:"objetivo"`
	Estado      string `json:"estado"`
}

// HU-10
func (s *server) createSprint(w http.ResponseWriter, r *http.Request) {
	ids := s.newPublicIDs()
	projectID, publicID, err := s.projectIDFromPath(r)
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
	s.writeResolved(w, http.StatusCreated, toSprintResponse(sp, publicID, ids), ids)
}

// HU-14
func (s *server) listProjectSprints(w http.ResponseWriter, r *http.Request) {
	ids := s.newPublicIDs()
	projectID, publicID, err := s.projectIDFromPath(r)
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
		resp = append(resp, toSprintResponse(sp, publicID, ids))
	}
	s.writeResolved(w, http.StatusOK, resp, ids)
}

// HU-11
func (s *server) assignItem(w http.ResponseWriter, r *http.Request) {
	sprintID, itemID, err := s.sprintItemPublicIDs(r)
	if err != nil {
		writeError(w, err)
		return
	}
	it, err := s.sprints.AssignItem(sprintID, itemID)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeItem(w, it)
}

// HU-12
func (s *server) completeItem(w http.ResponseWriter, r *http.Request) {
	sprintID, itemID, err := s.sprintItemPublicIDs(r)
	if err != nil {
		writeError(w, err)
		return
	}
	it, err := s.sprints.CompleteItem(sprintID, itemID)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeItem(w, it)
}

// HU-12
func (s *server) listCompletedItems(w http.ResponseWriter, r *http.Request) {
	sprintID, err := s.sprintIDFromPath(r)
	if err != nil {
		writeError(w, err)
		return
	}
	items, err := s.sprints.ListCompletedItems(sprintID)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeItems(w, items)
}

// HU-13
func (s *server) closeSprint(w http.ResponseWriter, r *http.Request) {
	ids := s.newPublicIDs()
	sprintID, err := s.sprintIDFromPath(r)
	if err != nil {
		writeError(w, err)
		return
	}
	sp, err := s.sprints.Close(sprintID)
	if err != nil {
		writeError(w, err)
		return
	}
	publicID, err := s.publicProjectID(sp.ProjectID)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeResolved(w, http.StatusOK, toSprintResponse(sp, publicID, ids), ids)
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

// toSprintResponse arma la respuesta del sprint con publicID, el UUID de su
// proyecto (HU-51).
func toSprintResponse(sp sprint.Sprint, publicID string, ids *publicIDs) sprintResponse {
	return sprintResponse{
		ID:          sp.PublicID.String(),
		ProyectoID:  publicID,
		Nombre:      sp.Name,
		FechaInicio: sp.StartDate.Format(dateLayout),
		FechaFin:    sp.EndDate.Format(dateLayout),
		Objetivo:    sp.Goal,
		Estado:      string(sp.Status),
	}
}

// sprintItemPublicIDs lee el {id} del sprint y el {itemId} del ítem de la
// ruta como sus UUID y devuelve sus IDs internos (HU-52, CA-52.3). El {id}
// se lee antes que el {itemId} (CA-52.4).
func (s *server) sprintItemPublicIDs(r *http.Request) (sprintID, itemID int, err error) {
	if sprintID, err = s.sprintIDFromPath(r); err != nil {
		return 0, 0, err
	}
	if itemID, err = s.itemIDFromPath(r, "itemId"); err != nil {
		return 0, 0, err
	}
	return sprintID, itemID, nil
}
