package httpapi

import (
	"net/http"
	"strings"

	"seguimiento-medicion-gpc/internal/defect"
)

// Handler del registro de defectos (HU-29, CA-29.11).

// HU-52, CA-52.3: el sprint de detección es el UUID del sprint, como texto.
type defectRequest struct {
	Descripcion       string `json:"descripcion"`
	Severidad         string `json:"severidad"`
	SprintDeteccionID string `json:"sprint_deteccion_id"`
}

// HU-52, CA-52.3: el defecto, su ítem y sus sprints se identifican con su
// UUID; un defecto sin resolver tiene "" en sprint_resolucion_id (D7).
type defectResponse struct {
	ID                 string          `json:"id"`
	Descripcion        string          `json:"descripcion"`
	Severidad          defect.Severity `json:"severidad"`
	Estado             defect.Status   `json:"estado"`
	ItemID             string          `json:"item_id"`
	SprintDeteccionID  string          `json:"sprint_deteccion_id"`
	SprintResolucionID string          `json:"sprint_resolucion_id"`
}

// registerDefect atiende POST /api/items/{id}/defectos. El orden de los errores
// es ID de la ruta, JSON, sprint de detección del cuerpo y Manager.
func (s *server) registerDefect(w http.ResponseWriter, r *http.Request) {
	ids := s.newPublicIDs()
	itemID, err := s.itemIDFromPath(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	var req defectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	sprintID, err := s.sprintIDFromBody(req.SprintDeteccionID)
	if err != nil {
		writeError(w, err)
		return
	}

	saved, err := s.defects.Register(itemID, req.Descripcion, req.Severidad, sprintID)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeResolved(w, http.StatusCreated, toDefectResponse(saved, ids), ids)
}

// Handler del cambio de estado de un defecto (HU-30, CA-30.9).

// HU-52, CA-52.3: el sprint de resolución es el UUID del sprint, como texto.
type defectStatusRequest struct {
	Estado             string `json:"estado"`
	SprintResolucionID string `json:"sprint_resolucion_id"`
}

// changeDefectStatus atiende PATCH /api/defectos/{id}/estado. El orden de los
// errores es ID de la ruta, JSON, sprint de resolución del cuerpo y Manager.
func (s *server) changeDefectStatus(w http.ResponseWriter, r *http.Request) {
	ids := s.newPublicIDs()
	id, err := s.defectIDFromPath(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req defectStatusRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	sprintID, err := s.sprintIDFromBody(req.SprintResolucionID)
	if err != nil {
		writeError(w, err)
		return
	}

	changed, err := s.defects.ChangeStatus(id, req.Estado, sprintID)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeResolved(w, http.StatusOK, toDefectResponse(changed, ids), ids)
}

// Handlers de la consulta de defectos (HU-31, CA-31.9).

type defectListResponse struct {
	Defectos []defectResponse `json:"defectos"`
	Cantidad int              `json:"cantidad"`
}

// listDefectsByProject atiende GET /api/proyectos/{id}/defectos.
func (s *server) listDefectsByProject(w http.ResponseWriter, r *http.Request) {
	// HU-51: el {id} del proyecto es su UUID.
	projectID := func(r *http.Request) (int, error) {
		id, _, err := s.projectIDFromPath(r)
		return id, err
	}
	s.listDefects(w, r, projectID, s.defects.ListByProject)
}

// listDefectsBySprint atiende GET /api/sprints/{id}/defectos.
func (s *server) listDefectsBySprint(w http.ResponseWriter, r *http.Request) {
	s.listDefects(w, r, s.sprintIDFromPath, s.defects.ListBySprint)
}

// listDefects lee el ID de la ruta con readID, lista con list y aplica el
// filtro ?estado=. El orden de los errores es ID de la ruta, estado del
// filtro y Manager. El handler decide que un filtro vacío o con solo espacios
// no filtra (D3).
func (s *server) listDefects(w http.ResponseWriter, r *http.Request, readID func(*http.Request) (int, error), list func(int) ([]defect.Defect, error)) {
	ids := s.newPublicIDs()
	id, err := readID(r)
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
	s.writeResolved(w, http.StatusOK, toDefectListResponse(defects, ids), ids)
}

// toDefectListResponse arma la lista siempre no nil, para que una consulta
// sin defectos sea [] y no null (D6).
func toDefectListResponse(defects []defect.Defect, ids *publicIDs) defectListResponse {
	resp := make([]defectResponse, 0, len(defects))
	for _, d := range defects {
		resp = append(resp, toDefectResponse(d, ids))
	}
	return defectListResponse{Defectos: resp, Cantidad: len(resp)}
}

func toDefectResponse(d defect.Defect, ids *publicIDs) defectResponse {
	return defectResponse{
		ID:                 d.PublicID.String(),
		Descripcion:        d.Description,
		Severidad:          d.Severity,
		Estado:             d.Status,
		ItemID:             ids.item(d.ItemID),
		SprintDeteccionID:  ids.sprint(d.DetectionSprintID),
		SprintResolucionID: ids.sprint(d.ResolutionSprintID),
	}
}
