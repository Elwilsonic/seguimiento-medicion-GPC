// Package report arma los reportes del proyecto (E11). HU-38 agrega el
// reporte de un proyecto, HU-39 el de un sprint y HU-40 la exportación del
// reporte del proyecto a PDF, con la librería github.com/go-pdf/fpdf. No tiene
// almacenamiento propio ni cálculos nuevos: lee de los Manager de proyectos,
// sprints, métricas, esfuerzos y defectos.
package report
