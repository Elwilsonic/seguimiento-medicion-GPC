// Package httpapi expone por HTTP las operaciones de proyectos y de backlog
// ya integradas (TEC-01) y, desde TEC-04, las de ítems y sprints bajo /api/… y
// la interfaz web en /, usando solo la librería estándar (net/http).
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"uuid"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/dashboard"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/effort"
	"seguimiento-medicion-gpc/internal/member"
	"seguimiento-medicion-gpc/internal/metrics"
	"seguimiento-medicion-gpc/internal/poker"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/report"
	"seguimiento-medicion-gpc/internal/sprint"
	"seguimiento-medicion-gpc/internal/web"
)

// maxBodyBytes es el tamaño máximo del cuerpo de un pedido: 1 MiB (CA-TEC04.5).
const maxBodyBytes = 1 << 20

var (
	errInvalidJSON  = errors.New("el cuerpo del pedido no es un JSON válido")
	errBodyTooLarge = errors.New("el cuerpo del pedido supera el tamaño máximo de 1 MiB")
	errInvalidID    = errors.New("el id debe ser un número entero")
	// errInvalidProjectID: el {id} de una ruta de proyecto no es un UUID en
	// forma canónica (HU-51, D2).
	errInvalidProjectID = errors.New("el id del proyecto debe ser un UUID")
	// Los errores de formato del resto de las entidades (HU-52, CA-52.4): el
	// {id}, el {itemId} o un campo del cuerpo no es un UUID en forma canónica.
	errInvalidItemID    = errors.New("el id del ítem debe ser un UUID")
	errInvalidSprintID  = errors.New("el id del sprint debe ser un UUID")
	errInvalidMemberID  = errors.New("el id del integrante debe ser un UUID")
	errInvalidDefectID  = errors.New("el id del defecto debe ser un UUID")
	errInvalidEffortID  = errors.New("el id del esfuerzo debe ser un UUID")
	errInvalidSessionID = errors.New("el id de la sesión debe ser un UUID")
	errInvalidDate      = errors.New("las fechas deben tener el formato AAAA-MM-DD")
	errInternal         = errors.New("error interno del servidor")
)

type errorResponse struct {
	Error string `json:"error"`
}

// server tiene los Manager que atienden los pedidos.
type server struct {
	projects  *project.Manager
	items     *backlog.Manager
	sprints   *sprint.Manager
	efforts   *effort.Manager
	poker     *poker.Manager
	defects   *defect.Manager
	metrics   *metrics.Manager
	dashboard *dashboard.Manager
	reports   *report.Manager
	members   *member.Manager
}

// NewHandler devuelve el handler HTTP de la aplicación con sus rutas.
func NewHandler(projects *project.Manager, items *backlog.Manager, sprints *sprint.Manager, efforts *effort.Manager, pokerManager *poker.Manager, defects *defect.Manager, members *member.Manager) http.Handler {
	// HU-32: metrics no tiene repositorio ni estado, así que se crea acá con
	// los Manager de ítems y de sprints, sin cambiar la firma de NewHandler.
	// HU-33 suma el Manager de esfuerzos para calcular M5.
	// HU-47 suma el Manager de defectos para calcular M9 y M10.
	// HU-36: dashboard tampoco tiene repositorio ni estado, y se crea igual.
	// HU-38: report tampoco tiene repositorio ni estado; lee las métricas con
	// el mismo Manager de métricas del servidor. HU-49 suma el Manager del
	// backlog para leer las historias.
	s := &server{projects: projects, items: items, sprints: sprints, efforts: efforts, poker: pokerManager, defects: defects, members: members,
		metrics:   metrics.NewManager(items, sprints, efforts, defects),
		dashboard: dashboard.NewManager(projects, items, sprints)}
	s.reports = report.NewManager(projects, items, sprints, s.metrics, efforts, defects)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/proyectos", s.createProject)
	mux.HandleFunc("GET /api/proyectos", s.listProjects)
	mux.HandleFunc("GET /api/proyectos/{id}", s.getProject)
	mux.HandleFunc("PUT /api/proyectos/{id}", s.updateProject)
	mux.HandleFunc("PATCH /api/proyectos/{id}/fecha-inicio", s.setProjectStartDate)
	mux.HandleFunc("PATCH /api/proyectos/{id}/fecha-fin", s.setProjectEndDate)
	mux.HandleFunc("PATCH /api/proyectos/{id}/fechas", s.setProjectDates)
	mux.HandleFunc("POST /api/proyectos/{id}/integrantes", s.registerMember)
	mux.HandleFunc("GET /api/proyectos/{id}/integrantes", s.listMembers)
	mux.HandleFunc("POST /api/proyectos/{id}/backlog", s.createItem)
	mux.HandleFunc("GET /api/proyectos/{id}/backlog", s.listItems)
	mux.HandleFunc("GET /api/items/{id}", s.getItem)
	mux.HandleFunc("PUT /api/items/{id}", s.updateItem)
	mux.HandleFunc("PATCH /api/items/{id}/estado", s.changeItemStatus)
	mux.HandleFunc("PATCH /api/items/{id}/story-points", s.assignStoryPoints)
	mux.HandleFunc("GET /api/items/{id}/poker-sessions", methodNotAllowed)
	mux.HandleFunc("POST /api/items/{id}/poker-sessions", s.openPokerSession)
	mux.HandleFunc("POST /api/poker-sessions/{id}/votes", s.votePokerSession)
	mux.HandleFunc("GET /api/poker-sessions/{id}/reveal", methodNotAllowed)
	mux.HandleFunc("POST /api/poker-sessions/{id}/reveal", s.revealPokerSession)
	mux.HandleFunc("GET /api/poker-sessions/{id}/estimate", methodNotAllowed)
	mux.HandleFunc("POST /api/poker-sessions/{id}/estimate", s.estimatePokerSession)
	mux.HandleFunc("GET /api/poker-sessions/{id}/rounds", methodNotAllowed)
	mux.HandleFunc("POST /api/poker-sessions/{id}/rounds", s.repeatRoundPokerSession)
	mux.HandleFunc("POST /api/items/{id}/defectos", s.registerDefect)
	mux.HandleFunc("PATCH /api/defectos/{id}/estado", s.changeDefectStatus)
	mux.HandleFunc("GET /api/proyectos/{id}/defectos", s.listDefectsByProject)
	mux.HandleFunc("GET /api/sprints/{id}/defectos", s.listDefectsBySprint)
	mux.HandleFunc("GET /api/items/{id}/esfuerzos", s.listEffortsByItem)
	mux.HandleFunc("PUT /api/integrantes/{id}", s.updateMember)
	mux.HandleFunc("GET /api/integrantes/{id}", memberMethodNotAllowed)
	mux.HandleFunc("POST /api/proyectos/{id}/sprints", s.createSprint)
	mux.HandleFunc("GET /api/proyectos/{id}/sprints", s.listProjectSprints)
	mux.HandleFunc("POST /api/sprints/{id}/items/{itemId}", s.assignItem)
	mux.HandleFunc("POST /api/sprints/{id}/items/{itemId}/completar", s.completeItem)
	mux.HandleFunc("POST /api/sprints/{id}/items/{itemId}/esfuerzos", s.registerEffort)
	mux.HandleFunc("GET /api/sprints/{id}/esfuerzos", s.listEffortsBySprint)
	mux.HandleFunc("PATCH /api/esfuerzos/{id}/horas", s.correctEffortHours)
	mux.HandleFunc("GET /api/sprints/{id}/completados", s.listCompletedItems)
	mux.HandleFunc("POST /api/sprints/{id}/cerrar", s.closeSprint)
	mux.HandleFunc("GET /api/sprints/{id}/metricas", s.sprintMetrics)
	mux.HandleFunc("GET /api/proyectos/{id}/metricas", s.projectMetrics)
	mux.HandleFunc("GET /api/proyectos/{id}/metricas/sprints", s.projectSprintsMetrics)
	mux.HandleFunc("GET /api/proyectos/{id}/resumen", s.projectSummary)
	mux.HandleFunc("GET /api/proyectos/{id}/reporte", s.projectReport)
	mux.HandleFunc("GET /api/proyectos/{id}/reporte/pdf", s.projectReportPDF)
	mux.HandleFunc("GET /api/sprints/{id}/reporte", s.sprintReport)
	// La interfaz web responde en / (CA-TEC04.8); las rutas /api/… son más
	// específicas, así que no la tapan.
	mux.Handle("GET /", web.Handler())
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

// parseProjectID lee el {id} de una ruta de proyecto como UUID en forma
// canónica (HU-51, D2): 36 caracteres con guiones en las posiciones 8, 13, 18
// y 23. Acepta mayúsculas, y rechaza las llaves, el prefijo urn:uuid: y la
// forma sin guiones, que uuid.Parse sí acepta, para que cada proyecto tenga
// una sola URL. No exige la versión 4: un UUID de otra versión tiene la forma
// correcta y simplemente no existe (404).
func parseProjectID(s string) (uuid.UUID, error) {
	return parsePublicID(s, errInvalidProjectID)
}

// projectIDFromPath lee el {id} de una ruta de proyecto (HU-51, D4) y
// devuelve el ID interno del proyecto y su UUID en minúsculas, para las
// respuestas. Un {id} que no es un UUID da errInvalidProjectID (400). Un UUID
// que no corresponde a ningún proyecto da el ID interno 0, que ningún
// proyecto tiene (RN-01.5 y RN-03.2): así cada Manager informa
// ErrProjectNotFound (404) en el mismo punto de su orden de validaciones que
// con un ID entero (HU-18, línea 72 de su SDD). Otro error del repositorio se
// devuelve tal cual (500).
func (s *server) projectIDFromPath(r *http.Request) (int, string, error) {
	u, err := parseProjectID(r.PathValue("id"))
	if err != nil {
		return 0, "", err
	}
	p, err := s.projects.GetByPublicID(u)
	if errors.Is(err, project.ErrProjectNotFound) {
		return 0, u.String(), nil
	}
	if err != nil {
		return 0, "", err
	}
	return p.ID, u.String(), nil
}

// publicProjectID devuelve el UUID del proyecto con ese ID interno, para las
// respuestas de las rutas por ID de otra entidad (ítem, sprint o integrante,
// HU-51, D5). Esa entidad siempre es de un proyecto que existe, así que si no
// se encuentra es un error interno (500) y no un 404: el error se informa sin
// su cadena (%v).
func (s *server) publicProjectID(projectID int) (string, error) {
	p, err := s.projects.Get(projectID)
	if err != nil {
		return "", fmt.Errorf("leer el proyecto %d: %v", projectID, err)
	}
	return p.PublicID.String(), nil
}

// missingInternalID es el ID interno de un UUID que no corresponde a ninguna
// entidad (HU-52, D4). No es el 0, porque en defectos el 0 (y cualquier
// negativo) es "sprint faltante" (400 de HU-29 y HU-30), y un sprint
// inexistente tiene que seguir dando 404. Es el máximo de INTEGER en
// PostgreSQL: un valor más grande haría fallar la consulta (500). Ningún
// repositorio llega a asignarlo, así que cada Manager informa la inexistencia
// en el mismo punto de su orden de validaciones que con un ID entero.
const missingInternalID = math.MaxInt32

// parsePublicID lee s como UUID en forma canónica, con la misma regla que
// parseProjectID (HU-51, D2). Si no lo es, devuelve invalid, el error de
// formato de la entidad (HU-52, D6).
func parsePublicID(s string, invalid error) (uuid.UUID, error) {
	if len(s) != 36 || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return uuid.UUID{}, invalid
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, invalid
	}
	return u, nil
}

// internalID traduce el UUID de text a su ID interno con lookup (HU-52, D4):
// un texto que no es un UUID da invalid (400), un UUID que no existe da
// missingInternalID y otro error del repositorio se devuelve tal cual (500).
func internalID[T any](text string, invalid, notFound error, lookup func(uuid.UUID) (T, error), id func(T) int) (int, error) {
	u, err := parsePublicID(text, invalid)
	if err != nil {
		return 0, err
	}
	v, err := lookup(u)
	if errors.Is(err, notFound) {
		return missingInternalID, nil
	}
	if err != nil {
		return 0, err
	}
	return id(v), nil
}

// itemIDFromPath lee el parámetro name de la ruta ({id} o {itemId}) como el
// UUID de un ítem y devuelve su ID interno (HU-52, D4).
func (s *server) itemIDFromPath(r *http.Request, name string) (int, error) {
	return internalID(r.PathValue(name), errInvalidItemID, backlog.ErrItemNotFound, s.items.GetByPublicID, func(v backlog.Item) int { return v.ID })
}

// sprintIDFromPath lee el {id} de la ruta como el UUID de un sprint y
// devuelve su ID interno (HU-52, D4).
func (s *server) sprintIDFromPath(r *http.Request) (int, error) {
	return internalID(r.PathValue("id"), errInvalidSprintID, sprint.ErrSprintNotFound, s.sprints.GetByPublicID, func(v sprint.Sprint) int { return v.ID })
}

// memberIDFromPath lee el {id} de la ruta como el UUID de un integrante y
// devuelve su ID interno (HU-52, D4).
func (s *server) memberIDFromPath(r *http.Request) (int, error) {
	return internalID(r.PathValue("id"), errInvalidMemberID, member.ErrMemberNotFound, s.members.GetByPublicID, func(v member.Member) int { return v.ID })
}

// defectIDFromPath lee el {id} de la ruta como el UUID de un defecto y
// devuelve su ID interno (HU-52, D4).
func (s *server) defectIDFromPath(r *http.Request) (int, error) {
	return internalID(r.PathValue("id"), errInvalidDefectID, defect.ErrDefectNotFound, s.defects.GetByPublicID, func(v defect.Defect) int { return v.ID })
}

// effortIDFromPath lee el {id} de la ruta como el UUID de un esfuerzo y
// devuelve su ID interno (HU-52, D4).
func (s *server) effortIDFromPath(r *http.Request) (int, error) {
	return internalID(r.PathValue("id"), errInvalidEffortID, effort.ErrEffortNotFound, s.efforts.GetByPublicID, func(v effort.Effort) int { return v.ID })
}

// sessionIDFromPath lee el {id} de la ruta como el UUID de una sesión de
// Planning Poker y devuelve su ID interno (HU-52, D4).
func (s *server) sessionIDFromPath(r *http.Request) (int, error) {
	return internalID(r.PathValue("id"), errInvalidSessionID, poker.ErrSessionNotFound, s.poker.GetByPublicID, func(v poker.Session) int { return v.ID })
}

// sprintIDFromBody lee un campo del cuerpo que nombra un sprint (HU-29 y
// HU-30) como su UUID y devuelve su ID interno. Un texto vacío da 0, que
// sigue siendo "sprint faltante" (HU-52, D4).
func (s *server) sprintIDFromBody(text string) (int, error) {
	if text == "" {
		return 0, nil
	}
	return internalID(text, errInvalidSprintID, sprint.ErrSprintNotFound, s.sprints.GetByPublicID, func(v sprint.Sprint) int { return v.ID })
}

// publicIDs traduce los IDs internos de ítems y sprints a sus UUID, para las
// respuestas (HU-52, D5). Busca cada ID una sola vez y guarda el primer
// error, que writeResolved informa como 500.
type publicIDs struct {
	s       *server
	items   map[int]string
	sprints map[int]string
	err     error
}

// newPublicIDs devuelve un resolvedor vacío.
func (s *server) newPublicIDs() *publicIDs {
	return &publicIDs{s: s, items: map[int]string{}, sprints: map[int]string{}}
}

// item devuelve el UUID del ítem con ese ID interno, o "" para el 0 ("sin
// resolver").
func (ids *publicIDs) item(id int) string {
	return resolve(ids, ids.items, id, func(id int) (uuid.UUID, error) {
		it, err := ids.s.items.GetItem(id)
		return it.PublicID, err
	})
}

// sprint devuelve el UUID del sprint con ese ID interno, o "" para el 0
// ("sin asignar" o "sin resolver", D7).
func (ids *publicIDs) sprint(id int) string {
	return resolve(ids, ids.sprints, id, func(id int) (uuid.UUID, error) {
		sp, err := ids.s.sprints.Get(id)
		return sp.PublicID, err
	})
}

// resolve busca el UUID del ID interno id con lookup, una sola vez por ID
// (cache). El 0 da "". El primer error queda en ids.err, sin su cadena (%v),
// como en publicProjectID.
func resolve(ids *publicIDs, cache map[int]string, id int, lookup func(int) (uuid.UUID, error)) string {
	if id == 0 {
		return ""
	}
	if s, ok := cache[id]; ok {
		return s
	}
	u, err := lookup(id)
	if err != nil {
		if ids.err == nil {
			ids.err = fmt.Errorf("leer el UUID del ID interno %d: %v", id, err)
		}
		return ""
	}
	cache[id] = u.String()
	return cache[id]
}

// writeResolved responde con v, o con 500 si el resolvedor tuvo un error al
// buscar algún UUID (HU-52, D5).
func (s *server) writeResolved(w http.ResponseWriter, status int, v any, ids *publicIDs) {
	if ids.err != nil {
		writeError(w, ids.err)
		return
	}
	writeJSON(w, status, v)
}

// statusFor traduce un error a su código HTTP. Los errores no reconocidos
// son errores internos.
func statusFor(err error) int {
	switch {
	case errors.Is(err, errInvalidJSON), errors.Is(err, errInvalidID), errors.Is(err, errInvalidProjectID),
		errors.Is(err, errInvalidItemID), errors.Is(err, errInvalidSprintID), errors.Is(err, errInvalidMemberID),
		errors.Is(err, errInvalidDefectID), errors.Is(err, errInvalidEffortID), errors.Is(err, errInvalidSessionID),
		errors.Is(err, project.ErrEmptyName), errors.Is(err, project.ErrInvalidStatus),
		errors.Is(err, project.ErrMissingStartDate), errors.Is(err, project.ErrMissingEndDate),
		errors.Is(err, backlog.ErrEmptyTitle), errors.Is(err, backlog.ErrInvalidPriority),
		errors.Is(err, backlog.ErrNoAcceptanceCriteria), errors.Is(err, backlog.ErrInvalidEstimatedHours),
		errors.Is(err, backlog.ErrInvalidStatus), errors.Is(err, backlog.ErrInvalidStoryPoints),
		errors.Is(err, errInvalidDate),
		errors.Is(err, sprint.ErrEmptyName), errors.Is(err, sprint.ErrEmptyGoal),
		errors.Is(err, sprint.ErrMissingStartDate), errors.Is(err, sprint.ErrMissingEndDate),
		errors.Is(err, sprint.ErrEndNotAfterStart),
		errors.Is(err, effort.ErrEmptyMember), errors.Is(err, effort.ErrEmptyActivity),
		errors.Is(err, effort.ErrMissingDate), errors.Is(err, effort.ErrInvalidHours),
		errors.Is(err, defect.ErrEmptyDescription), errors.Is(err, defect.ErrInvalidSeverity),
		errors.Is(err, defect.ErrMissingDetectionSprint),
		errors.Is(err, defect.ErrInvalidStatus), errors.Is(err, defect.ErrMissingResolutionSprint),
		errors.Is(err, poker.ErrEmptyVoter), errors.Is(err, poker.ErrInvalidCard),
		errors.Is(err, member.ErrNotProjectMember),
		errors.Is(err, member.ErrEmptyName), errors.Is(err, member.ErrMissingRole), errors.Is(err, member.ErrInvalidRole):
		return http.StatusBadRequest
	case errors.Is(err, project.ErrProjectNotFound), errors.Is(err, backlog.ErrItemNotFound),
		errors.Is(err, sprint.ErrSprintNotFound), errors.Is(err, poker.ErrSessionNotFound),
		errors.Is(err, defect.ErrDefectNotFound), errors.Is(err, member.ErrMemberNotFound),
		errors.Is(err, effort.ErrEffortNotFound):
		return http.StatusNotFound
	case errors.Is(err, project.ErrDuplicateName), errors.Is(err, backlog.ErrInvalidTransition),
		errors.Is(err, project.ErrStartDateNotRegistered), errors.Is(err, project.ErrEndNotAfterStart),
		errors.Is(err, project.ErrStartNotBeforeEnd), errors.Is(err, project.ErrDatesNotRegistered),
		errors.Is(err, backlog.ErrItemAlreadyAssigned), errors.Is(err, backlog.ErrItemFinished),
		errors.Is(err, sprint.ErrOpenSprintExists),
		errors.Is(err, sprint.ErrItemFromAnotherProject), errors.Is(err, sprint.ErrSprintClosed),
		errors.Is(err, sprint.ErrItemNotAssignedToSprint),
		errors.Is(err, defect.ErrDetectionSprintFromAnotherProject),
		errors.Is(err, defect.ErrInvalidTransition), errors.Is(err, defect.ErrResolutionSprintFromAnotherProject),
		errors.Is(err, defect.ErrResolutionBeforeDetection),
		errors.Is(err, poker.ErrSessionAlreadyOpen), errors.Is(err, poker.ErrItemFinished),
		errors.Is(err, poker.ErrSessionNotOpen), errors.Is(err, poker.ErrNoVotes),
		errors.Is(err, poker.ErrSessionNotRevealed), errors.Is(err, member.ErrDuplicateMember):
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
