# language: es
# HU-34 · Calcular el porcentaje de historias completadas (Issue #52)
# Especificación: docs/especificaciones/HU-34-porcentaje-completadas.md
# Fórmula: docs/metricas.md (TEC-03), métrica M8.
# CA-34.6 (API e interfaz) se cubre con tests de Go, no con escenarios.

Característica: HU-34 Calcular el porcentaje de historias completadas
  Como Agile Enabler,
  quiero calcular el porcentaje de historias completadas de un sprint y de un proyecto,
  para evaluar el avance.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  # Caso normal - ejemplo de TEC-03
  Escenario: Porcentaje de un sprint con el ejemplo de TEC-03 (CA-34.1)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | story_points | estado    |
      | A      | 5            | TERMINADO |
      | B      | 3            | TERMINADO |
      | C      | 2            | EN_CURSO  |
      | D      | 0            | PENDIENTE |
    Cuando consulto el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces el porcentaje de historias completadas es 50 con 2 de 4 ítems terminados

  # Caso alternativo - el sprint solo cuenta sus ítems
  Escenario: El sprint no cuenta los ítems sin sprint (CA-34.1)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | story_points | estado    |
      | A      | 5            | TERMINADO |
    Y que el proyecto "Sistema de ventas" tiene los ítems sin sprint:
      | titulo | story_points | estado    |
      | E      | 8            | PENDIENTE |
    Cuando consulto el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces el porcentaje de historias completadas es 100 con 1 de 1 ítems terminados

  # Caso alternativo - nivel proyecto con todo el backlog
  Escenario: Porcentaje del proyecto con todo el backlog (CA-34.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | story_points | estado    |
      | A      | 5            | TERMINADO |
      | B      | 3            | EN_CURSO  |
    Y que el proyecto "Sistema de ventas" tiene los ítems sin sprint:
      | titulo | story_points | estado    |
      | E      | 8            | PENDIENTE |
      | F      | 2            | TERMINADO |
      | G      | 0            | PENDIENTE |
    Cuando consulto el porcentaje de historias completadas del proyecto "Sistema de ventas"
    Entonces el porcentaje de historias completadas es 40 con 2 de 5 ítems terminados

  # Caso límite - sin ítems
  Escenario: Un sprint sin ítems no tiene porcentaje disponible (CA-34.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando consulto el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces el porcentaje de historias completadas no está disponible

  Escenario: Un proyecto sin ítems no tiene porcentaje disponible (CA-34.3)
    Cuando consulto el porcentaje de historias completadas del proyecto "Sistema de ventas"
    Entonces el porcentaje de historias completadas no está disponible

  # Caso límite - ninguno terminado es 0 y está disponible
  Escenario: Ningún ítem terminado da cero (CA-34.1)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | story_points | estado    |
      | A      | 5            | PENDIENTE |
      | B      | 3            | EN_CURSO  |
    Cuando consulto el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces el porcentaje de historias completadas es 0 con 0 de 2 ítems terminados

  # Caso límite - los ítems sin Story Points cuentan
  Escenario: Los ítems sin Story Points cuentan igual (CA-34.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | story_points | estado    |
      | A      | 0            | TERMINADO |
      | B      | 0            | PENDIENTE |
    Cuando consulto el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces el porcentaje de historias completadas es 50 con 1 de 2 ítems terminados

  # Casos de error
  Esquema del escenario: Consultar un sprint inexistente falla (CA-34.5)
    Cuando consulto el porcentaje de historias completadas del sprint con el ID <id>
    Entonces la consulta falla porque el sprint no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Esquema del escenario: Consultar un proyecto inexistente falla (CA-34.5)
    Cuando consulto el porcentaje de historias completadas del proyecto con el ID <id>
    Entonces la consulta falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Escenario: Consultar el porcentaje no modifica los ítems ni los sprints (CA-34.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | story_points | estado    |
      | A      | 5            | TERMINADO |
      | B      | 0            | PENDIENTE |
    Cuando consulto el porcentaje de historias completadas del sprint "Sprint 1"
    Y consulto el porcentaje de historias completadas del proyecto "Sistema de ventas"
    Entonces los ítems y los sprints del proyecto no cambiaron
