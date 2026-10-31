package report

import "seguimiento-medicion-gpc/internal/backlog"

// Stories son las historias de un reporte (HU-49): las del backlog del
// proyecto o las asignadas al sprint, copias ordenadas por ID (RN-49.6 y
// RN-49.8), y cuántas están en TERMINADO (RN-49.4).
type Stories struct {
	Stories   []backlog.Item
	Completed int
}

// storiesOf arma una lista nueva con las historias que cumplen keep, en el
// orden recibido (el de ListItems, por ID), y cuenta las TERMINADO. La lista
// nunca es nil, para que la API responda [] (RN-49.3).
func storiesOf(items []backlog.Item, keep func(backlog.Item) bool) Stories {
	stories := Stories{Stories: []backlog.Item{}}
	for _, it := range items {
		if !keep(it) {
			continue
		}
		stories.Stories = append(stories.Stories, it)
		if it.Status == backlog.StatusFinished {
			stories.Completed++
		}
	}
	return stories
}
