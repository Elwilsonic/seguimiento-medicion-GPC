# language: es
# HU-37 · Ver las métricas en el dashboard con gráficos (Issue #55)
#
# Trazabilidad: cada escenario cita su criterio de la SDD
# (docs/especificaciones/HU-37-graficos-dashboard.md) y cubre una fila de
# "Casos límite" o de "Condiciones de error". Los pasos usan los Manager
# reales en memoria de backlog, sprint, effort y defect, con repositorios que
# cuentan las lecturas y que pueden fallar a pedido (los de HU-35).
#
# Los escenarios cubren la consulta por sprint (metrics.ProjectSprints), que
# es la parte con reglas de negocio. Lo que es solo HTTP o solo interfaz se
# cubre en internal/httpapi: la ruta, el JSON y sus códigos (CA-37.4) en
# metrics_test.go, y la sección "Dashboard" con los gráficos SVG (CA-37.1,
# CA-37.2 y CA-37.6) en web_test.go.

Característica: HU-37 Ver las métricas en el dashboard con gráficos
  Como Agile Enabler
  quiero ver las métricas en el dashboard con representaciones gráficas
  para decidir con datos

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  Escenario: Cada sprint del proyecto con sus métricas, en orden de ID (CA-37.1, CA-37.4 y CA-37.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
      | B      | 3  | 6     | TERMINADO | 5        |
      | C      | 2  | 0     | EN_CURSO  | 3        |
      | D      | 0  | 4     | PENDIENTE | 0        |
    Y que hay un defecto detectado en el sprint "Sprint 1" y resuelto en el sprint "Sprint 1"
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 2" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | E      | 3  | 4     | TERMINADO | 10       |
      | F      | 2  | 4     | PENDIENTE | 0        |
    Y que hay un defecto detectado en el sprint "Sprint 1" y resuelto en el sprint "Sprint 2"
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces la lista tiene los sprints: Sprint 1, Sprint 2
    Y el sprint "Sprint 1" de la lista está CERRADO
    Y el sprint "Sprint 2" de la lista está ABIERTO
    Y en el sprint "Sprint 1" los story points planificados son 10 y los completados son 8
    Y en el sprint "Sprint 1" el porcentaje de historias completadas es 50 con 2 de 4 ítems
    Y en el sprint "Sprint 1" las horas estimadas son 20 y las reales son 20
    Y en el sprint "Sprint 1" los defectos detectados son 2 y los resueltos son 1
    Y en el sprint "Sprint 1" todas las métricas están disponibles
    Y en el sprint "Sprint 2" los story points planificados son 5 y los completados son 3
    Y en el sprint "Sprint 2" el porcentaje de historias completadas es 50 con 1 de 2 ítems
    Y en el sprint "Sprint 2" las horas estimadas son 8 y las reales son 10
    Y en el sprint "Sprint 2" la desviación es 2 horas y 25 %
    Y en el sprint "Sprint 2" los defectos detectados son 0 y los resueltos son 1
    Y modificar el resultado no cambia la próxima consulta
    Y los datos del proyecto no cambiaron

  Escenario: Los valores de cada sprint coinciden con su consulta individual (CA-37.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 3  | 4     | TERMINADO | 5        |
      | B      | 2  | 0     | EN_CURSO  | 1.5      |
    Y que hay un defecto abierto detectado en el sprint "Sprint 1"
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 2" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | C      | 8  | 10    | PENDIENTE | 2        |
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces la lista tiene los sprints: Sprint 1, Sprint 2
    Y las métricas de cada sprint son iguales a las de su consulta individual

  Escenario: Un proyecto con un único sprint, abierto (CA-37.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   | esfuerzo |
      | A      | 5  | 10    | EN_CURSO | 3        |
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces la lista tiene los sprints: Sprint 1
    Y el sprint "Sprint 1" de la lista está ABIERTO
    Y en el sprint "Sprint 1" los story points planificados son 5 y los completados son 0
    Y en el sprint "Sprint 1" las horas estimadas son 10 y las reales son 3
    Y en el sprint "Sprint 1" la desviación es -7 horas y -70 %

  Escenario: Más de 3 sprints cerrados aparecen todos, y el backlog se lee una sola vez (CA-37.4)
    Dado que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas" con un ítem terminado de 2 story points
    Y que existe un sprint cerrado "Sprint 2" en el proyecto "Sistema de ventas" con un ítem terminado de 3 story points
    Y que existe un sprint cerrado "Sprint 3" en el proyecto "Sistema de ventas" con un ítem terminado de 5 story points
    Y que existe un sprint cerrado "Sprint 4" en el proyecto "Sistema de ventas" con un ítem terminado de 8 story points
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces la lista tiene los sprints: Sprint 1, Sprint 2, Sprint 3, Sprint 4
    Y en el sprint "Sprint 1" los story points planificados son 2 y los completados son 2
    Y en el sprint "Sprint 4" los story points planificados son 8 y los completados son 8
    Y los ítems del backlog se leyeron 2 veces en la consulta

  Escenario: Un proyecto sin sprints devuelve una lista vacía (CA-37.2 y CA-37.4)
    Dado que el proyecto "Sistema de ventas" tiene los ítems sin sprint:
      | titulo | sp | horas | estado    |
      | A      | 3  | 4     | PENDIENTE |
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces la lista de sprints está vacía
    Y los datos del proyecto no cambiaron

  Escenario: Un sprint sin ítems tiene todo en 0 y el porcentaje y la desviación no disponibles (CA-37.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces la lista tiene los sprints: Sprint 1
    Y en el sprint "Sprint 1" los story points planificados son 0 y los completados son 0
    Y en el sprint "Sprint 1" las horas estimadas son 0 y las reales son 0
    Y en el sprint "Sprint 1" los defectos detectados son 0 y los resueltos son 0
    Y en el sprint "Sprint 1" las métricas no disponibles son: porcentaje, desviación

  Escenario: Un sprint con horas estimadas y sin esfuerzos tiene la desviación no disponible (CA-37.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | PENDIENTE | 0        |
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces en el sprint "Sprint 1" las horas estimadas son 10 y las reales son 0
    Y en el sprint "Sprint 1" las métricas no disponibles son: desviación

  Escenario: Un ítem sin sprint no cuenta en ningún sprint de la lista (CA-37.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 3  | 4     | TERMINADO | 5        |
    Y que el proyecto "Sistema de ventas" tiene los ítems sin sprint:
      | titulo | sp | horas | estado    |
      | E      | 8  | 10    | PENDIENTE |
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces en el sprint "Sprint 1" los story points planificados son 3 y los completados son 3
    Y en el sprint "Sprint 1" el porcentaje de historias completadas es 100 con 1 de 1 ítems
    Y en el sprint "Sprint 1" las horas estimadas son 4 y las reales son 5

  Esquema del escenario: Un proyecto inexistente devuelve un error y ninguna lista (CA-37.4)
    Cuando consulto las métricas por sprint del proyecto con el ID <id>
    Entonces la consulta falla porque el proyecto no existe
    Y no se devuelve ninguna lista

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Esquema del escenario: Las lecturas se hacen en orden y la primera que falla corta la consulta (CA-37.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
    Y que hay un defecto abierto detectado en el sprint "Sprint 1"
    Y que falla la lectura de <primera>
    Y que falla la lectura de <segunda>
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces la consulta falla con el error de la lectura de <primera>
    Y no se devuelve ninguna lista

    Ejemplos:
      | primera       | segunda       |
      | los ítems     | los sprints   |
      | los sprints   | los esfuerzos |
      | los esfuerzos | los defectos  |

  Escenario: Una falla al leer los defectos corta la consulta sin lista parcial (CA-37.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
    Y que falla la lectura de los defectos
    Cuando consulto las métricas por sprint del proyecto "Sistema de ventas"
    Entonces la consulta falla con el error de la lectura de los defectos
    Y no se devuelve ninguna lista
