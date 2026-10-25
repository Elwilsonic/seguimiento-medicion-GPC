package metrics

import (
	"slices"

	"seguimiento-medicion-gpc/internal/backlog"
)

// velocitySprints es la cantidad de sprints cerrados que promedia la
// velocidad (decisión de TEC-03).
const velocitySprints = 3

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

// storyPoints suma M1 y M2 de los ítems y cuenta los que no tienen Story
// Points (RN-32.1 a RN-32.3).
func storyPoints(items []backlog.Item) (planned, completed, unestimated int) {
	for _, item := range items {
		if item.StoryPoints == 0 {
			unestimated++
			continue
		}
		planned += item.StoryPoints
		if item.Status == backlog.StatusFinished {
			completed += item.StoryPoints
		}
	}
	return planned, completed, unestimated
}

// velocity calcula M3: el promedio de M2 de los últimos 3 sprints cerrados
// por ID descendente, o de todos si hay menos (RN-32.5). Sin sprints cerrados
// devuelve 0 y 0 (RN-32.6). No redondea (RN-32.7) ni modifica closed.
func velocity(closed []closedSprint) (float64, int) {
	if len(closed) == 0 {
		return 0, 0
	}
	latest := slices.Clone(closed)
	slices.SortFunc(latest, func(a, b closedSprint) int { return b.ID - a.ID })
	latest = latest[:min(len(latest), velocitySprints)]
	total := 0
	for _, s := range latest {
		total += s.Completed
	}
	return float64(total) / float64(len(latest)), len(latest)
}
