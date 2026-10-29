# language: es
# HU-39 · Generar el reporte de un sprint (Issue #57)
#
# Trazabilidad: cada escenario cita su criterio de la SDD
# (docs/especificaciones/HU-39-reporte-sprint.md) y cubre una fila de
# "Casos límite" o uno de sus ejemplos. Los pasos usan los Manager reales en
# memoria de proyectos, backlog, sprints, métricas, esfuerzos y defectos, con
# repositorios que cuentan las lecturas y que pueden fallar a pedido.
#
# Dos partes de la SDD son solo HTTP o de la interfaz y se cubren en
# internal/httpapi/sprint_reports_test.go e internal/httpapi/web_test.go, no
# con escenarios godog: CA-39.8 (el reporte desde la API y desde la sección
# "Reportes", con el formato de cada parte y el redondeo de la API) y la fila
# "HTTP: {id} no numérico" (400).

Característica: HU-39 Generar el reporte de un sprint
  Como Agile Enabler
  quiero generar el reporte de un sprint
  para documentar su resultado

  Antecedentes:
    Dado que existe un proyecto con nombre "Software Metrics"

  Escenario: El reporte del sprint del ejemplo de TEC-03 tiene todas sus partes (CA-39.1)
    Dado que existe un sprint abierto "Sprint 1" del "2026-10-05" al "2026-10-16" con el objetivo "Registrar el esfuerzo del equipo" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 5  | 10    | TERMINADO |
      | B      | 3  | 6     | TERMINADO |
      | C      | 2  | 0     | EN_CURSO  |
      | D      | 0  | 4     | PENDIENTE |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 12    |
      | Sprint 1 | B    | Luis       | 5     |
      | Sprint 1 | C    | Ana        | 3     |
    Y que hay un defecto RESUELTO del ítem "A" detectado en el sprint "Sprint 1" y resuelto en el sprint "Sprint 1"
    Y que cierro el sprint "Sprint 1"
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces el sprint del reporte es "Sprint 1" en estado "CERRADO" del "2026-10-05" al "2026-10-16" con el objetivo "Registrar el esfuerzo del equipo"
    Y los story points planificados son 10 y los completados son 8
    Y los ítems totales son 4 y los terminados son 2
    Y las horas estimadas son 20 y las reales son 20
    Y los defectos detectados son 1 y los resueltos son 1
    Y los esfuerzos del reporte son:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 12    |
      | Sprint 1 | B    | Luis       | 5     |
      | Sprint 1 | C    | Ana        | 3     |
    Y el total de horas es 20
    Y los defectos del reporte son:
      | item | estado   | deteccion | resolucion |
      | A    | RESUELTO | Sprint 1  | Sprint 1   |
    Y los integrantes son "Ana, Luis"

  Escenario: Un sprint abierto genera el mismo reporte que uno cerrado (CA-39.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 5  | 10    | TERMINADO |
      | B      | 3  | 6     | EN_CURSO  |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 8     |
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces el sprint del reporte es "Sprint 1" en estado "ABIERTO"
    Y los story points planificados son 8 y los completados son 5
    Y las horas estimadas son 16 y las reales son 8
    Y el total de horas es 8
    Y los integrantes son "Ana"

  Escenario: Un sprint recién creado genera el reporte con las secciones vacías (CA-39.4)
    Dado que existe un sprint abierto "Sprint vacío" en el proyecto "Software Metrics"
    Cuando genero el reporte del sprint "Sprint vacío"
    Entonces el sprint del reporte es "Sprint vacío" en estado "ABIERTO"
    Y los story points planificados son 0 y los completados son 0
    Y los ítems totales son 0 y los terminados son 0
    Y las horas estimadas son 0 y las reales son 0
    Y los defectos detectados son 0 y los resueltos son 0
    Y las métricas no disponibles son: porcentaje de completadas, desviación
    Y el reporte no tiene esfuerzos, defectos ni integrantes

  Escenario: Un sprint con ítems y sin esfuerzos tiene las métricas y las listas vacías (CA-39.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 3  | 4     | PENDIENTE |
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces los story points planificados son 3 y los completados son 0
    Y las horas estimadas son 4 y las reales son 0
    Y las métricas no disponibles son: desviación
    Y el reporte no tiene esfuerzos, defectos ni integrantes

  Escenario: Un defecto detectado en otro sprint y resuelto en este aparece en el reporte (CA-39.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 5  | 10    | TERMINADO |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Software Metrics"
    Y que hay un defecto RESUELTO del ítem "A" detectado en el sprint "Sprint 1" y resuelto en el sprint "Sprint 2"
    Cuando genero el reporte del sprint "Sprint 2"
    Entonces los defectos detectados son 0 y los resueltos son 1
    Y los defectos del reporte son:
      | item | estado   | deteccion | resolucion |
      | A    | RESUELTO | Sprint 1  | Sprint 2   |
    Y los defectos del reporte coinciden con los detectados y resueltos de las métricas

  Escenario: Los defectos del sprint son los detectados o resueltos en él, una sola vez cada uno (CA-39.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 5  | 10    | TERMINADO |
      | B      | 3  | 6     | EN_CURSO  |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 2" tiene los ítems:
      | titulo | sp | horas | estado   |
      | C      | 2  | 4     | EN_CURSO |
    Y que hay un defecto ABIERTO del ítem "B" detectado en el sprint "Sprint 1"
    Y que hay un defecto RESUELTO del ítem "C" detectado en el sprint "Sprint 2" y resuelto en el sprint "Sprint 2"
    Y que hay un defecto EN_CORRECCION del ítem "C" detectado en el sprint "Sprint 2"
    Y que hay un defecto RESUELTO del ítem "A" detectado en el sprint "Sprint 1" y resuelto en el sprint "Sprint 2"
    Cuando genero el reporte del sprint "Sprint 2"
    Entonces los defectos detectados son 2 y los resueltos son 2
    Y los defectos del reporte son:
      | item | estado        | deteccion | resolucion |
      | C    | RESUELTO      | Sprint 2  | Sprint 2   |
      | C    | EN_CORRECCION | Sprint 2  |            |
      | A    | RESUELTO      | Sprint 1  | Sprint 2   |
    Y los defectos del reporte coinciden con los detectados y resueltos de las métricas

  Escenario: Cada parte del reporte es igual a su consulta de origen (CA-39.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 5  | 10    | TERMINADO |
      | B      | 3  | 6     | EN_CURSO  |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 4     |
      | Sprint 1 | B    | Luis       | 2.5   |
    Y que hay un defecto ABIERTO del ítem "B" detectado en el sprint "Sprint 1"
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces cada parte del reporte es igual a su consulta de origen
    Y el total de horas es igual a las horas reales de las métricas

  Escenario: Horas con decimales suman el total exacto, igual a las horas reales (CA-39.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 0  | 1     | PENDIENTE |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 0.1   |
      | Sprint 1 | A    | Ana        | 0.2   |
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces el total de horas es la suma exacta de los esfuerzos
    Y el total de horas es igual a las horas reales de las métricas

  Escenario: Solo aparecen los esfuerzos y los integrantes del sprint, cada integrante una vez (CA-39.6)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 5  | 10    | TERMINADO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Pedro      | 3     |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 2" tiene los ítems:
      | titulo | sp | horas | estado   |
      | B      | 3  | 6     | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 2 | B    | Luis       | 2     |
      | Sprint 2 | B    | Ana        | 1     |
      | Sprint 2 | B    | Luis       | 4     |
    Y que hay un defecto ABIERTO del ítem "A" detectado en el sprint "Sprint 1"
    Cuando genero el reporte del sprint "Sprint 2"
    Entonces los esfuerzos del reporte son:
      | sprint   | item | integrante | horas |
      | Sprint 2 | B    | Luis       | 2     |
      | Sprint 2 | B    | Ana        | 1     |
      | Sprint 2 | B    | Luis       | 4     |
    Y el total de horas es 7
    Y los integrantes son "Luis, Ana"
    Y el reporte no tiene defectos

  Escenario: Generar el reporte dos veces sin cambios da el mismo resultado y no modifica nada (CA-39.7)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 5  | 10    | TERMINADO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 4     |
    Y que hay un defecto ABIERTO del ítem "A" detectado en el sprint "Sprint 1"
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces generar el reporte de nuevo da el mismo resultado
    Y modificar el reporte no cambia el próximo reporte
    Y los datos guardados no cambiaron

  Esquema del escenario: Generar el reporte de un sprint inexistente falla (CA-39.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Cuando genero el reporte del sprint con el ID <id>
    Entonces el reporte falla porque el sprint no existe
    Y solo se leyó el sprint

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Esquema del escenario: Si falla una lectura no se hacen las siguientes (CA-39.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Software Metrics"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 5  | 10    | TERMINADO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 4     |
    Y que hay un defecto ABIERTO del ítem "A" detectado en el sprint "Sprint 1"
    Y que falla la lectura de <fuente> del reporte del sprint
    Cuando genero el reporte del sprint "Sprint 1"
    Entonces el reporte falla con el error de la lectura de <fuente>
    Y no se hicieron las lecturas siguientes a <fuente>

    Ejemplos:
      | fuente        |
      | el sprint     |
      | las métricas  |
      | los esfuerzos |
      | los defectos  |
