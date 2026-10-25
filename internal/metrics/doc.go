// Package metrics calcula las métricas del proyecto definidas en
// docs/metricas.md (TEC-03) a partir de los ítems del backlog y de los
// sprints. No tiene almacenamiento propio: lee de los Manager de backlog y de
// sprint, y cada historia de E9 agrega sus métricas en su propio archivo.
package metrics
