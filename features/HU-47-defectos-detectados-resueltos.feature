# language: es
# HU-47 · Calcular los defectos detectados y resueltos (Issue #128)
# Especificación: docs/especificaciones/HU-47-defectos-detectados-resueltos.md
# Fórmulas: docs/metricas.md (TEC-03), métricas M9 y M10.
# CA-47.6 (API e interfaz) se cubre con tests de Go, no con escenarios.

Característica: HU-47 Calcular los defectos detectados y resueltos
  Como Agile Enabler,
  quiero calcular la cantidad de defectos detectados y resueltos de un sprint y de un proyecto,
  para evaluar la calidad.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  # Caso normal - ejemplo de TEC-03
  Escenario: Defectos detectados y resueltos de un sprint (CA-47.1)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene los defectos:
      | descripcion | detectado_en | estado   | resuelto_en |
      | D1          | Sprint 1     | RESUELTO | Sprint 1    |
      | D2          | Sprint 1     | ABIERTO  | -           |
    Cuando consulto los defectos del sprint "Sprint 1"
    Entonces hay 2 defectos detectados y 1 resueltos

  # Caso alternativo - nivel proyecto
  Escenario: Defectos detectados y resueltos del proyecto (CA-47.2)
    Dado que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene los defectos:
      | descripcion | detectado_en | estado        | resuelto_en |
      | D1          | Sprint 1     | RESUELTO      | Sprint 2    |
      | D2          | Sprint 2     | ABIERTO       | -           |
      | D3          | Sprint 2     | EN_CORRECCION | -           |
    Cuando consulto los defectos del proyecto "Sistema de ventas"
    Entonces hay 3 defectos detectados y 1 resueltos

  # Caso alternativo - otro proyecto no cuenta
  Escenario: Los defectos de otro proyecto no cuentan (CA-47.2)
    Dado que existe un proyecto con nombre "Inventario"
    Y que existe un sprint abierto "Sprint A" en el proyecto "Inventario"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Inventario" tiene los defectos:
      | descripcion | detectado_en | estado   | resuelto_en |
      | D9          | Sprint A     | RESUELTO | Sprint A    |
    Y que el proyecto "Sistema de ventas" tiene los defectos:
      | descripcion | detectado_en | estado  | resuelto_en |
      | D1          | Sprint 1     | ABIERTO | -           |
    Cuando consulto los defectos del proyecto "Sistema de ventas"
    Entonces hay 1 defectos detectados y 0 resueltos

  # Caso límite - sin defectos
  Escenario: Un sprint sin defectos da cero (CA-47.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando consulto los defectos del sprint "Sprint 1"
    Entonces hay 0 defectos detectados y 0 resueltos

  Escenario: Un proyecto sin defectos da cero (CA-47.3)
    Cuando consulto los defectos del proyecto "Sistema de ventas"
    Entonces hay 0 defectos detectados y 0 resueltos

  # Caso límite - detección y resolución en sprints distintos
  Esquema del escenario: Un defecto cuenta en su sprint de detección y en el de resolución (CA-47.4)
    Dado que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene los defectos:
      | descripcion | detectado_en | estado   | resuelto_en |
      | D1          | Sprint 1     | RESUELTO | Sprint 2    |
    Cuando consulto los defectos del sprint "<sprint>"
    Entonces hay <detectados> defectos detectados y <resueltos> resueltos

    Ejemplos:
      | sprint   | detectados | resueltos |
      | Sprint 1 | 1          | 0         |
      | Sprint 2 | 0          | 1         |

  # Caso límite - solo cuenta el estado resuelto
  Escenario: Un defecto en corrección no cuenta como resuelto aunque tenga sprint de resolución (CA-47.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene los defectos:
      | descripcion | detectado_en | estado        | resuelto_en |
      | D1          | Sprint 1     | EN_CORRECCION | Sprint 1    |
    Cuando consulto los defectos del sprint "Sprint 1"
    Entonces hay 1 defectos detectados y 0 resueltos

  # Casos de error
  Esquema del escenario: Consultar un sprint inexistente falla (CA-47.5)
    Cuando consulto los defectos del sprint con el ID <id>
    Entonces la consulta falla porque el sprint no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Esquema del escenario: Consultar un proyecto inexistente falla (CA-47.5)
    Cuando consulto los defectos del proyecto con el ID <id>
    Entonces la consulta falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Escenario: Consultar los defectos no los modifica (CA-47.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene los defectos:
      | descripcion | detectado_en | estado   | resuelto_en |
      | D1          | Sprint 1     | RESUELTO | Sprint 1    |
      | D2          | Sprint 1     | ABIERTO  | -           |
    Cuando consulto los defectos del sprint "Sprint 1"
    Y consulto los defectos del proyecto "Sistema de ventas"
    Entonces los defectos del proyecto no cambiaron
