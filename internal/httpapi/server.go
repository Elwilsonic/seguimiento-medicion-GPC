// Package httpapi expone por HTTP las operaciones de proyectos y de backlog
// ya integradas (TEC-01), usando solo la librería estándar (net/http).
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
)

// maxBodyBytes es el tamaño máximo del cuerpo de un pedido: 1 MiB (CA-TEC04.5).
const maxBodyBytes = 1 << 20

var (
	errInvalidJSON  = errors.New("el cuerpo del pedido no es un JSON válido")
	errBodyTooLarge = errors.New("el cuerpo del pedido supera el tamaño máximo de 1 MiB")
	errInvalidID    = errors.New("el id debe ser un número entero")
	errInternal     = errors.New("error interno del servidor")
)

type projectRequest struct {
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
}

type projectResponse struct {
	ID          int    `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
}

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

type errorResponse struct {
	Error string `json:"error"`
}

// server tiene los Manager que atienden los pedidos.
type server struct {
	projects *project.Manager
	items    *backlog.Manager
}

// NewHandler devuelve el handler HTTP de la aplicación con sus rutas.
func NewHandler(projects *project.Manager, items *backlog.Manager) http.Handler {
	s := &server{projects: projects, items: items}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/proyectos", s.createProject)
	mux.HandleFunc("GET /api/proyectos/{id}", s.getProject)
	mux.HandleFunc("PUT /api/proyectos/{id}", s.updateProject)
	mux.HandleFunc("POST /api/proyectos/{id}/backlog", s.createItem)
	mux.HandleFunc("GET /api/proyectos/{id}/backlog", s.listItems)
	mux.HandleFunc("GET /api/items/{id}", s.getItem)
	mux.HandleFunc("PUT /api/items/{id}", s.updateItem)
	mux.HandleFunc("PATCH /api/items/{id}/estado", s.changeItemStatus)
	return mux
}

// HU-01
func (s *server) createProject(w http.ResponseWriter, r *http.Request) {
	var req projectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	p, err := s.projects.Create(req.Nombre, req.Descripcion)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toProjectResponse(p))
}

// HU-03
func (s *server) getProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.projects.Get(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

// HU-02
func (s *server) updateProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req projectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	p, err := s.projects.Update(id, req.Nombre, req.Descripcion, req.Estado)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
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

// decodeJSON lee el cuerpo del pedido en v. Rechaza un cuerpo de más de 1 MiB
// (CA-TEC04.5) y un cuerpo con texto después del JSON (CA-TEC04.6).
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err := dec.Decode(v); err != nil {
		return bodyError(err)
	}
	// Después del JSON solo pueden quedar espacios: el siguiente Decode tiene
	// que llegar al final del cuerpo.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return bodyError(err)
	}
	return nil
}

// bodyError traduce un error de lectura del cuerpo: si se pasó del tamaño
// máximo es errBodyTooLarge y, si no, errInvalidJSON.
func bodyError(err error) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return errBodyTooLarge
	}
	return errInvalidJSON
}

// pathID lee el {id} de la ruta.
func pathID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, errInvalidID
	}
	return id, nil
}

// statusFor traduce un error a su código HTTP. Los errores no reconocidos
// son errores internos.
func statusFor(err error) int {
	switch {
	case errors.Is(err, errInvalidJSON), errors.Is(err, errInvalidID),
		errors.Is(err, project.ErrEmptyName), errors.Is(err, project.ErrInvalidStatus),
		errors.Is(err, backlog.ErrEmptyTitle), errors.Is(err, backlog.ErrInvalidPriority),
		errors.Is(err, backlog.ErrNoAcceptanceCriteria), errors.Is(err, backlog.ErrInvalidEstimatedHours),
		errors.Is(err, backlog.ErrInvalidStatus):
		return http.StatusBadRequest
	case errors.Is(err, project.ErrProjectNotFound), errors.Is(err, backlog.ErrItemNotFound):
		return http.StatusNotFound
	case errors.Is(err, project.ErrDuplicateName), errors.Is(err, backlog.ErrInvalidTransition):
		return http.StatusConflict
	case errors.Is(err, errBodyTooLarge):
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}

// writeError responde con el código del error y su mensaje. Un error interno
// no expone su detalle.
func writeError(w http.ResponseWriter, err error) {
	status := statusFor(err)
	if status == http.StatusInternalServerError {
		err = errInternal
	}
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

// writeJSON responde con el código y el valor codificado en JSON.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func toProjectResponse(p project.Project) projectResponse {
	return projectResponse{ID: p.ID, Nombre: p.Name, Descripcion: p.Description, Estado: string(p.Status)}
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
