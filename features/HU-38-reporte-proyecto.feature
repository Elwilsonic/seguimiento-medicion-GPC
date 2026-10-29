# language: es
# HU-38 · Generar el reporte del proyecto (Issue #56)
#
# Trazabilidad: cada escenario cita su criterio de la SDD
# (docs/especificaciones/HU-38-reporte-proyecto.md) y cubre una fila de
# "Casos límite". Los pasos usan los Manager reales en memoria de proyectos,
# backlog, sprints, métricas, esfuerzos y defectos, con repositorios que
# cuentan las lecturas y que pueden fallar a pedido.
#
# Dos partes de la SDD son solo HTTP o de la interfaz y se cubren en
# internal/httpapi/reports_test.go e internal/httpapi/web_test.go, no con
# escenarios godog: CA-38.7 (el reporte desde la API y desde la sección
# "Reportes", con el formato de cada parte y el redondeo de la API) y la fila
# "HTTP: {id} no numérico" (400).

Característica: HU-38 Generar el reporte del proyecto
  Como Agile Enabler
  quiero generar un reporte del proyecto
  para presentar su estado a los profesores

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  Escenario: El reporte del ejemplo de TEC-03 tiene todas sus partes (CA-38.1)
    Dado que el proyecto "Sistema de ventas" está EN_CURSO con fecha de inicio "2026-10-05"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
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
    Y que hay un defecto RESUELTO del ítem "A" detectado y resuelto en el sprint "Sprint 1"
    Y que cierro el sprint "Sprint 1"
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces el proyecto del reporte es "Sistema de ventas" en estado "EN_CURSO" con fecha de inicio "2026-10-05"
    Y los story points planificados son 10 y los completados son 8
    Y la velocidad es 8
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
      | item | estado   |
      | A    | RESUELTO |
    Y los integrantes son "Ana, Luis"

  Escenario: Un proyecto recién creado genera el reporte con las secciones vacías (CA-38.2)
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces el proyecto del reporte es "Sistema de ventas" en estado "PLANIFICADO" sin fecha de inicio
    Y el reporte no tiene esfuerzos, defectos ni integrantes
    Y los story points planificados son 0 y los completados son 0
    Y las horas estimadas son 0 y las reales son 0
    Y los defectos detectados son 0 y los resueltos son 0
    Y las métricas no disponibles son: velocidad, porcentaje, desviación

  Escenario: Un proyecto sin fecha de inicio muestra sus datos sin fecha (CA-38.1)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   |
      | A      | 3  | 4     | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 2     |
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces el proyecto del reporte es "Sistema de ventas" en estado "PLANIFICADO" sin fecha de inicio
    Y el total de horas es 2
    Y los integrantes son "Ana"

  Escenario: Un proyecto con ítems y sin sprints tiene las métricas y las listas vacías (CA-38.2)
    Dado que el proyecto "Sistema de ventas" tiene los ítems sin sprint:
      | titulo | sp | horas | estado    |
      | A      | 3  | 4     | PENDIENTE |
      | B      | 2  | 0     | TERMINADO |
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces los ítems totales son 2 y los terminados son 1
    Y los story points planificados son 5 y los completados son 2
    Y el reporte no tiene esfuerzos, defectos ni integrantes
    Y las métricas no disponibles son: velocidad, desviación

  Escenario: Sprints sin esfuerzos dejan vacíos los esfuerzos y los integrantes (CA-38.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 3  | 4     | TERMINADO |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces la velocidad es 3
    Y el reporte no tiene esfuerzos, defectos ni integrantes

  Escenario: Tres sprints, abiertos y cerrados, con esfuerzos (CA-38.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 3  | 4     | TERMINADO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 5     |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 2" tiene los ítems:
      | titulo | sp | horas | estado    |
      | B      | 5  | 6     | TERMINADO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 2 | B    | Luis       | 4     |
      | Sprint 2 | B    | Ana        | 0.5   |
    Y que cierro el sprint "Sprint 2"
    Y que existe un sprint abierto "Sprint 3" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 3" tiene los ítems:
      | titulo | sp | horas | estado   |
      | C      | 2  | 10    | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 3 | C    | Luis       | 2     |
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces los esfuerzos del reporte son:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 5     |
      | Sprint 2 | B    | Luis       | 4     |
      | Sprint 2 | B    | Ana        | 0.5   |
      | Sprint 3 | C    | Luis       | 2     |
    Y el total de horas es 11.5
    Y el total de horas es igual a las horas reales de las métricas
    Y cada parte del reporte es igual a su consulta de origen

  Escenario: Un integrante con esfuerzos en varios sprints aparece una sola vez (CA-38.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   |
      | A      | 3  | 4     | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Luis       | 1     |
      | Sprint 1 | A    | Ana        | 2     |
      | Sprint 1 | A    | Luis       | 3     |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 2" tiene los ítems:
      | titulo | sp | horas | estado   |
      | B      | 2  | 2     | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 2 | B    | Ana        | 1     |
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces los integrantes son "Luis, Ana"

  Escenario: "Ana" y "ana" se listan como dos integrantes (CA-38.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   |
      | A      | 3  | 4     | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 1     |
      | Sprint 1 | A    | ana        | 2     |
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces los integrantes son "Ana, ana"

  Escenario: Los defectos aparecen en los tres estados, ordenados por ID (CA-38.1)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   |
      | A      | 3  | 4     | EN_CURSO |
      | B      | 2  | 2     | EN_CURSO |
    Y que hay un defecto ABIERTO del ítem "A" detectado en el sprint "Sprint 1"
    Y que hay un defecto RESUELTO del ítem "B" detectado y resuelto en el sprint "Sprint 1"
    Y que hay un defecto EN_CORRECCION del ítem "A" detectado en el sprint "Sprint 1"
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces los defectos del reporte son:
      | item | estado        |
      | A    | ABIERTO       |
      | B    | RESUELTO      |
      | A    | EN_CORRECCION |

  Escenario: Los esfuerzos, los defectos y los integrantes de otro proyecto no se incluyen (CA-38.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   |
      | A      | 3  | 4     | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 2     |
    Y que hay un defecto ABIERTO del ítem "A" detectado en el sprint "Sprint 1"
    Y que existe un proyecto con nombre "Sistema de compras"
    Y que existe un sprint abierto "Sprint X" en el proyecto "Sistema de compras"
    Y que el sprint "Sprint X" tiene los ítems:
      | titulo | sp | horas | estado   |
      | Z      | 8  | 20    | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint X | Z    | Pedro      | 7     |
    Y que hay un defecto ABIERTO del ítem "Z" detectado en el sprint "Sprint X"
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces los esfuerzos del reporte son:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 2     |
    Y los defectos del reporte son:
      | item | estado  |
      | A    | ABIERTO |
    Y los integrantes son "Ana"

  Escenario: Horas con decimales suman el total exacto, igual a las horas reales (CA-38.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    |
      | A      | 0  | 0.1   | PENDIENTE |
      | B      | 0  | 0.2   | PENDIENTE |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 0.1   |
      | Sprint 1 | B    | Luis       | 0.2   |
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces el total de horas es igual a las horas reales de las métricas
    Y el total de horas es 0.3 al redondear a un decimal

  Escenario: Generar el reporte dos veces sin cambios da el mismo resultado (CA-38.6)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   |
      | A      | 3  | 4     | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 2     |
    Y que hay un defecto ABIERTO del ítem "A" detectado en el sprint "Sprint 1"
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces generar el reporte de nuevo da el mismo resultado
    Y los datos guardados no cambiaron

  Escenario: Modificar el reporte devuelto no cambia lo guardado (CA-38.6)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   |
      | A      | 3  | 4     | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 2     |
    Y que hay un defecto ABIERTO del ítem "A" detectado en el sprint "Sprint 1"
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces modificar el reporte no cambia el próximo reporte
    Y los datos guardados no cambiaron

  Esquema del escenario: Generar el reporte de un proyecto inexistente falla (CA-38.3)
    Cuando genero el reporte del proyecto con el ID <id>
    Entonces el reporte falla porque el proyecto no existe
    Y solo se leyó el proyecto

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Esquema del escenario: Si falla una lectura no se hacen las siguientes (CA-38.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   |
      | A      | 3  | 4     | EN_CURSO |
    Y que se registraron los esfuerzos:
      | sprint   | item | integrante | horas |
      | Sprint 1 | A    | Ana        | 2     |
    Y que hay un defecto ABIERTO del ítem "A" detectado en el sprint "Sprint 1"
    Y que falla la lectura de <fuente> del reporte
    Cuando genero el reporte del proyecto "Sistema de ventas"
    Entonces el reporte falla con el error de la lectura de <fuente>
    Y no se hicieron las lecturas siguientes a <fuente>

    Ejemplos:
      | fuente        |
      | el proyecto   |
      | las métricas  |
      | los sprints   |
      | los esfuerzos |
      | los defectos  |
