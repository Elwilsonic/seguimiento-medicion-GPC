# language: es
# HU-32 · Calcular Story Points planificados, completados y velocidad (Issue #50)
# Especificación: docs/especificaciones/HU-32-sp-velocidad.md
# Fórmulas: docs/metricas.md (TEC-03), métricas M1, M2 y M3.
# CA-32.8 (PostgreSQL) y CA-32.9 (API e interfaz) se cubren con tests de Go,
# no con escenarios.

Característica: HU-32 Calcular Story Points planificados, completados y velocidad
  Como Agile Enabler,
  quiero calcular los Story Points planificados y completados de un sprint y la velocidad del equipo,
  para saber cuánto trabajo se comprometió y cuánto se terminó.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  # Caso normal - ejemplo de TEC-03
  Escenario: Story Points de un sprint con el ejemplo de TEC-03 (CA-32.1, CA-32.2 y CA-32.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | story_points | estado    |
      | A      | 5            | TERMINADO |
      | B      | 3            | TERMINADO |
      | C      | 2            | EN_CURSO  |
      | D      | 0            | PENDIENTE |
    Cuando consulto los story points del sprint "Sprint 1"
    Entonces los story points planificados son 10
    Y los story points completados son 8
    Y hay 1 ítems sin estimar

  # Caso límite
  Escenario: Un sprint sin ítems devuelve cero (CA-32.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando consulto los story points del sprint "Sprint 1"
    Entonces los story points planificados son 0
    Y los story points completados son 0
    Y hay 0 ítems sin estimar

  # Caso alternativo - nivel proyecto con todo el backlog
  Escenario: Story Points del proyecto con todo el backlog (CA-32.1, CA-32.2 y CA-32.3)
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
    Cuando consulto los story points del proyecto "Sistema de ventas"
    Entonces los story points planificados son 18
    Y los story points completados son 7
    Y hay 1 ítems sin estimar

  # Caso límite
  Escenario: Un proyecto sin ítems devuelve cero (CA-32.4 y CA-32.6)
    Cuando consulto los story points del proyecto "Sistema de ventas"
    Entonces los story points planificados son 0
    Y los story points completados son 0
    Y hay 0 ítems sin estimar
    Y la velocidad no está disponible

  # Caso normal y límites de la velocidad. Cada sprint cerrado se escribe con
  # los story points de sus ítems terminados ("5+5" son dos ítems de 5; "-" es
  # un sprint sin ítems), en el orden en que se crean.
  Esquema del escenario: Velocidad con los últimos sprints cerrados (CA-32.5)
    Dado que el proyecto "Sistema de ventas" tiene sprints cerrados con story points completados "<cerrados>"
    Cuando consulto los story points del proyecto "Sistema de ventas"
    Entonces la velocidad es <velocidad> con <usados> sprints cerrados

    Ejemplos:
      | cerrados             | velocidad | usados |
      | 8                    | 8.0       | 1      |
      | 8; 5+5; 8+3+1        | 10.0      | 3      |
      | 2; 8; 5+5; 8+3+1     | 10.0      | 3      |
      | -; 5+5               | 5.0       | 2      |

  # Caso alternativo - un sprint abierto no cuenta
  Escenario: Un sprint abierto no cuenta para la velocidad (CA-32.5 y CA-32.6)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | story_points | estado    |
      | A      | 13           | TERMINADO |
      | B      | 8            | TERMINADO |
    Cuando consulto los story points del proyecto "Sistema de ventas"
    Entonces la velocidad no está disponible

  # Casos de error
  Esquema del escenario: Consultar un sprint inexistente falla (CA-32.7)
    Cuando consulto los story points del sprint con el ID <id>
    Entonces la consulta falla porque el sprint no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Esquema del escenario: Consultar un proyecto inexistente falla (CA-32.7)
    Cuando consulto los story points del proyecto con el ID <id>
    Entonces la consulta falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Escenario: Consultar las métricas no modifica los ítems ni los sprints (CA-32.7)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | story_points | estado    |
      | A      | 5            | TERMINADO |
      | B      | 0            | PENDIENTE |
    Cuando consulto los story points del sprint "Sprint 1"
    Y consulto los story points del proyecto "Sistema de ventas"
    Entonces los ítems y los sprints del proyecto no cambiaron
