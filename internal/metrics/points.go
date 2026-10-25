package metrics

import "errors"

// errNotImplemented es el error de los stubs del RED de HU-32.
var errNotImplemented = errors.New("no implementado")

// SprintPoints son los Story Points de un sprint (HU-32): planificados (M1),
// completados (M2) y la cantidad de ítems sin estimar.
type SprintPoints struct {
	SprintID    int
	Planned     int
	Completed   int
	Unestimated int
}

// ProjectPoints son los Story Points de un proyecto (HU-32): planificados
// (M1) y completados (M2) de todo el backlog, la cantidad de ítems sin estimar
// y la velocidad del equipo (M3). Velocity no está redondeada (RN-32.7); si no
// hay sprints cerrados, VelocityAvailable es false (RN-32.6).
type ProjectPoints struct {
	ProjectID         int
	Planned           int
	Completed         int
	Unestimated       int
	Velocity          float64
	VelocityAvailable bool
	ClosedSprints     int
}

// closedSprint es un sprint CERRADO con sus Story Points completados (M2).
type closedSprint struct {
	ID        int
	Completed int
}

// velocity calcula M3 con los sprints cerrados (RN-32.5 y RN-32.6).
func velocity(closed []closedSprint) (float64, int) {
	return 0, 0
}
