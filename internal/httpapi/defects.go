package httpapi

import (
	"net/http"
	"strings"

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

// Handlers de la consulta de defectos (HU-31, CA-31.9).

type defectListResponse struct {
	Defectos []defectResponse `json:"defectos"`
	Cantidad int              `json:"cantidad"`
}

// listDefectsByProject atiende GET /api/proyectos/{id}/defectos.
func (s *server) listDefectsByProject(w http.ResponseWriter, r *http.Request) {
	s.listDefects(w, r, s.defects.ListByProject)
}

// listDefectsBySprint atiende GET /api/sprints/{id}/defectos.
func (s *server) listDefectsBySprint(w http.ResponseWriter, r *http.Request) {
	s.listDefects(w, r, s.defects.ListBySprint)
}

// listDefects lista con list y aplica el filtro ?estado=. El orden de los
// errores es ID de la ruta, estado del filtro y Manager. El handler decide
// que un filtro vacío o con solo espacios no filtra (D3).
func (s *server) listDefects(w http.ResponseWriter, r *http.Request, list func(int) ([]defect.Defect, error)) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	estado := strings.TrimSpace(r.URL.Query().Get("estado"))
	var status defect.Status
	if estado != "" {
		if status, err = defect.ParseStatus(estado); err != nil {
			writeError(w, err)
			return
		}
	}

	defects, err := list(id)
	if err != nil {
		writeError(w, err)
		return
	}
	if estado != "" {
		defects = defect.FilterByStatus(defects, status)
	}
	writeJSON(w, http.StatusOK, toDefectListResponse(defects))
}

// toDefectListResponse arma la lista siempre no nil, para que una consulta
// sin defectos sea [] y no null (D6).
func toDefectListResponse(defects []defect.Defect) defectListResponse {
	resp := make([]defectResponse, 0, len(defects))
	for _, d := range defects {
		resp = append(resp, toDefectResponse(d))
	}
	return defectListResponse{Defectos: resp, Cantidad: len(resp)}
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
