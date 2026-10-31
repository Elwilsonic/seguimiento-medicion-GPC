# language: es
# HU-49 · Ver las historias y sus estimaciones en el reporte (Issue #170)
#
# Trazabilidad: cada escenario cita su criterio de la SDD
# (docs/especificaciones/HU-49-historias-en-reporte.md). Los pasos usan los
# Manager reales en memoria de proyectos, backlog, sprints, métricas,
# esfuerzos y defectos, con repositorios que cuentan las lecturas y que pueden
# fallar a pedido.
#
# Dos partes de la SDD son solo HTTP o de la interfaz y se cubren en
# internal/httpapi/reports_test.go, internal/httpapi/sprint_reports_test.go e
# internal/httpapi/web_test.go, no con escenarios godog: la parte "historias"
# de la API, igual a GET /api/items/{id} (CA-49.3), y la tabla de historias de
# la sección "Reportes" (CA-49.7).

Característica: HU-49 Ver las historias y sus estimaciones en el reporte
  Como Agile Enabler
  quiero ver en el reporte del proyecto o del sprint cada historia con su estado y su estimación
  para saber qué se planificó, qué se completó y cuánto se estimó

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  Escenario: El reporte del proyecto incluye cada historia del backlog con su estimación (CA-49.1, CA-49.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene las historias:
      | titulo          | prioridad | sp | horas | estado    | sprint   |
      | Registrar venta | ALTA      | 3  | 8     | TERMINADO | Sprint 1 |
      | Emitir factura  | BAJA      | 5  | 6     | EN_CURSO  | Sprint 1 |
      | Consultar stock | MEDIA     | 0  | 0     | PENDIENTE |          |
    Y que existe un proyecto con nombre "Sistema de compras"
    Y que el proyecto "Sistema de compras" tiene las historias:
      | titulo           | prioridad | sp | horas | estado    | sprint |
      | Registrar compra | ALTA      | 8  | 12    | PENDIENTE |        |
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces las historias del reporte son:
      | titulo          | prioridad | estado    | sp | horas | sprint   |
      | Registrar venta | ALTA      | TERMINADO | 3  | 8     | Sprint 1 |
      | Emitir factura  | BAJA      | EN_CURSO  | 5  | 6     | Sprint 1 |
      | Consultar stock | MEDIA     | PENDIENTE | 0  | 0     |          |
    Y la cantidad de historias es 3 y las completadas son 1
    Y las cantidades de historias coinciden con las métricas del reporte
    Y cada historia del reporte es igual a su consulta en el backlog

  Escenario: El reporte de un sprint incluye solo las historias asignadas al sprint (CA-49.2, CA-49.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene las historias:
      | titulo          | prioridad | sp | horas | estado    | sprint   |
      | Registrar venta | ALTA      | 3  | 8     | TERMINADO | Sprint 1 |
      | Emitir factura  | BAJA      | 5  | 6     | EN_CURSO  | Sprint 1 |
      | Consultar stock | MEDIA     | 0  | 0     | PENDIENTE |          |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene las historias:
      | titulo        | prioridad | sp | horas | estado    | sprint   |
      | Anular venta  | MEDIA     | 2  | 4     | PENDIENTE | Sprint 2 |
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces las historias del reporte son:
      | titulo          | prioridad | estado    | sp | horas | sprint   |
      | Registrar venta | ALTA      | TERMINADO | 3  | 8     | Sprint 1 |
      | Emitir factura  | BAJA      | EN_CURSO  | 5  | 6     | Sprint 1 |
    Y la cantidad de historias es 2 y las completadas son 1
    Y las cantidades de historias coinciden con las métricas del reporte
    Y cada historia del reporte es igual a su consulta en el backlog

  Escenario: Un proyecto sin historias tiene la lista vacía y cantidades 0 (CA-49.4, CA-49.6)
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces el reporte no tiene historias
    Y la cantidad de historias es 0 y las completadas son 0
    Y la sección "Historias" del PDF del reporte, entre "Métricas" y "Esfuerzos", es:
      | linea                                    |
      | Sin historias en el backlog.             |
      | Cantidad: 0 historia(s), 0 completada(s) |

  Escenario: Un sprint sin historias asignadas tiene la lista vacía (CA-49.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene las historias:
      | titulo          | prioridad | sp | horas | estado    | sprint |
      | Consultar stock | MEDIA     | 0  | 0     | PENDIENTE |        |
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces el reporte no tiene historias
    Y la cantidad de historias es 0 y las completadas son 0

  Escenario: El PDF indica la historia sin estimar, sin horas y sin sprint (CA-49.4, CA-49.6)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene las historias:
      | titulo          | prioridad | sp | horas | estado    | sprint   |
      | Registrar venta | ALTA      | 3  | 8     | TERMINADO | Sprint 1 |
      | Consultar stock | MEDIA     | 0  | 0     | PENDIENTE |          |
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces la sección "Historias" del PDF del reporte, entre "Métricas" y "Esfuerzos", es:
      | linea                                                                                  |
      | #1 · Registrar venta · ALTA · TERMINADO · 3 SP · 8.0 h estimadas · sprint 1            |
      | #2 · Consultar stock · MEDIA · PENDIENTE · sin estimar · sin horas estimadas · sin sprint |
      | Cantidad: 2 historia(s), 1 completada(s)                                               |

  Escenario: Si falla la lectura de las historias del proyecto, no hay reporte parcial (CA-49.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene las historias:
      | titulo          | prioridad | sp | horas | estado    | sprint   |
      | Registrar venta | ALTA      | 3  | 8     | TERMINADO | Sprint 1 |
    Y que falla la lectura de las historias del reporte
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces el reporte falla con el error de la lectura de las historias
    Y las historias se leyeron después de los defectos

  Escenario: Si falla la lectura de las historias del sprint, no hay reporte parcial (CA-49.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene las historias:
      | titulo          | prioridad | sp | horas | estado    | sprint   |
      | Registrar venta | ALTA      | 3  | 8     | TERMINADO | Sprint 1 |
    Y que falla la lectura de las historias del reporte
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces el reporte falla con el error de la lectura de las historias
    Y las historias se leyeron después de los defectos

  Escenario: El reporte de un proyecto inexistente no lee las historias (Condiciones de error)
    Cuando genero el reporte del proyecto con el ID 999
    Entonces el reporte falla porque el proyecto no existe
    Y no se leyeron las historias

  Escenario: Modificar las historias del reporte no cambia lo guardado (CA-49.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene las historias:
      | titulo          | prioridad | sp | horas | estado    | sprint   |
      | Registrar venta | ALTA      | 3  | 8     | TERMINADO | Sprint 1 |
      | Consultar stock | MEDIA     | 0  | 0     | PENDIENTE |          |
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces modificar las historias del reporte no cambia el próximo reporte
    Y las historias guardadas no cambiaron
