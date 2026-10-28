package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/defect"
)

// Handler del registro de defectos (HU-29, CA-29.11).

type defectRequest struct {
	Descripcion       string `json:"descripcion"`
	Severidad         string `json:"severidad"`
	SprintDeteccionID int    `json:"sprint_deteccion_id"`
}

type defectResponse struct {
	ID                 int             `json:"id"`
	Descripcion        string          `json:"descripcion"`
	Severidad          defect.Severity `json:"severidad"`
	Estado             defect.Status   `json:"estado"`
	ItemID             int             `json:"item_id"`
	SprintDeteccionID  int             `json:"sprint_deteccion_id"`
	SprintResolucionID int             `json:"sprint_resolucion_id"`
}

// registerDefect atiende POST /api/items/{id}/defectos. El orden de los errores
// es ID de la ruta, JSON y Manager.
func (s *server) registerDefect(w http.ResponseWriter, r *http.Request) {
	itemID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req defectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	saved, err := s.defects.Register(itemID, req.Descripcion, req.Severidad, req.SprintDeteccionID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toDefectResponse(saved))
}

// Handler del cambio de estado de un defecto (HU-30, CA-30.9).

type defectStatusRequest struct {
	Estado             string `json:"estado"`
	SprintResolucionID int    `json:"sprint_resolucion_id"`
}

// changeDefectStatus atiende PATCH /api/defectos/{id}/estado. El orden de los
// errores es ID de la ruta, JSON y Manager.
func (s *server) changeDefectStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req defectStatusRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	changed, err := s.defects.ChangeStatus(id, req.Estado, req.SprintResolucionID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toDefectResponse(changed))
}

func toDefectResponse(d defect.Defect) defectResponse {
	return defectResponse{
		ID:                 d.ID,
		Descripcion:        d.Description,
		Severidad:          d.Severity,
		Estado:             d.Status,
		ItemID:             d.ItemID,
		SprintDeteccionID:  d.DetectionSprintID,
		SprintResolucionID: d.ResolutionSprintID,
	}
}
