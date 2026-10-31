package report

import "seguimiento-medicion-gpc/internal/backlog"

// Stories son las historias de un reporte (HU-49): las del backlog del
// proyecto o las asignadas al sprint, copias ordenadas por ID (RN-49.6 y
// RN-49.8), y cuántas están en TERMINADO (RN-49.4).
type Stories struct {
	Stories   []backlog.Item
	Completed int
}
