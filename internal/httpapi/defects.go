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
