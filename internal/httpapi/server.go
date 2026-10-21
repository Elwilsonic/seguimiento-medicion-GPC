// Package httpapi expone por HTTP las operaciones de proyectos y de backlog
// ya integradas (TEC-01) y, desde TEC-04, las de ítems y sprints, usando solo
// la librería estándar (net/http).
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// maxBodyBytes es el tamaño máximo del cuerpo de un pedido: 1 MiB (CA-TEC04.5).
const maxBodyBytes = 1 << 20

var (
	errInvalidJSON  = errors.New("el cuerpo del pedido no es un JSON válido")
	errBodyTooLarge = errors.New("el cuerpo del pedido supera el tamaño máximo de 1 MiB")
	errInvalidID    = errors.New("el id debe ser un número entero")
	errInvalidDate  = errors.New("las fechas deben tener el formato AAAA-MM-DD")
	errInternal     = errors.New("error interno del servidor")
)

type errorResponse struct {
	Error string `json:"error"`
}

// server tiene los Manager que atienden los pedidos.
type server struct {
	projects *project.Manager
	items    *backlog.Manager
	sprints  *sprint.Manager
}

// NewHandler devuelve el handler HTTP de la aplicación con sus rutas.
func NewHandler(projects *project.Manager, items *backlog.Manager, sprints *sprint.Manager) http.Handler {
	s := &server{projects: projects, items: items, sprints: sprints}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/proyectos", s.createProject)
	mux.HandleFunc("GET /api/proyectos/{id}", s.getProject)
	mux.HandleFunc("PUT /api/proyectos/{id}", s.updateProject)
	mux.HandleFunc("POST /api/proyectos/{id}/backlog", s.createItem)
	mux.HandleFunc("GET /api/proyectos/{id}/backlog", s.listItems)
	mux.HandleFunc("GET /api/items/{id}", s.getItem)
	mux.HandleFunc("PUT /api/items/{id}", s.updateItem)
	mux.HandleFunc("PATCH /api/items/{id}/estado", s.changeItemStatus)
	mux.HandleFunc("POST /api/proyectos/{id}/sprints", s.createSprint)
	mux.HandleFunc("POST /api/sprints/{id}/items/{itemId}", s.assignItem)
	mux.HandleFunc("POST /api/sprints/{id}/items/{itemId}/completar", s.completeItem)
	mux.HandleFunc("GET /api/sprints/{id}/completados", s.listCompletedItems)
	mux.HandleFunc("POST /api/sprints/{id}/cerrar", s.closeSprint)
	return mux
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
	return pathInt(r, "id")
}

// pathInt lee el parámetro name de la ruta como número entero.
func pathInt(r *http.Request, name string) (int, error) {
	n, err := strconv.Atoi(r.PathValue(name))
	if err != nil {
		return 0, errInvalidID
	}
	return n, nil
}

// statusFor traduce un error a su código HTTP. Los errores no reconocidos
// son errores internos.
func statusFor(err error) int {
	switch {
	case errors.Is(err, errInvalidJSON), errors.Is(err, errInvalidID),
		errors.Is(err, project.ErrEmptyName), errors.Is(err, project.ErrInvalidStatus),
		errors.Is(err, backlog.ErrEmptyTitle), errors.Is(err, backlog.ErrInvalidPriority),
		errors.Is(err, backlog.ErrNoAcceptanceCriteria), errors.Is(err, backlog.ErrInvalidEstimatedHours),
		errors.Is(err, backlog.ErrInvalidStatus), errors.Is(err, errInvalidDate),
		errors.Is(err, sprint.ErrEmptyName), errors.Is(err, sprint.ErrEmptyGoal),
		errors.Is(err, sprint.ErrMissingStartDate), errors.Is(err, sprint.ErrMissingEndDate),
		errors.Is(err, sprint.ErrEndNotAfterStart):
		return http.StatusBadRequest
	case errors.Is(err, project.ErrProjectNotFound), errors.Is(err, backlog.ErrItemNotFound),
		errors.Is(err, sprint.ErrSprintNotFound):
		return http.StatusNotFound
	case errors.Is(err, project.ErrDuplicateName), errors.Is(err, backlog.ErrInvalidTransition),
		errors.Is(err, backlog.ErrItemAlreadyAssigned), errors.Is(err, sprint.ErrOpenSprintExists),
		errors.Is(err, sprint.ErrItemFromAnotherProject), errors.Is(err, sprint.ErrSprintClosed),
		errors.Is(err, sprint.ErrItemNotAssignedToSprint):
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
