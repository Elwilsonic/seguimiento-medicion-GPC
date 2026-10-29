# language: es
# HU-14 · Consultar sprints anteriores (Issue #28)
# Especificación: docs/especificaciones/HU-14-sprints-anteriores.md

Característica: HU-14 Consultar sprints anteriores
  Como Agile Enabler,
  quiero consultar los sprints anteriores de un proyecto,
  para revisar su historial.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  # Caso normal - CA-14.1
  Escenario: Listar los sprints cerrados y el abierto de un proyecto (CA-14.1)
    Dado que existe un sprint cerrado "Sprint 1" del "2026-10-01" al "2026-10-08" con sprint goal "Backlog inicial" en el proyecto "Sistema de ventas"
    Y que existe un sprint cerrado "Sprint 2" del "2026-10-09" al "2026-10-16" con sprint goal "Estimación" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 3" del "2026-10-17" al "2026-10-24" con sprint goal "Métricas" en el proyecto "Sistema de ventas"
    Cuando consulto los sprints del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo estos sprints:
      | id | nombre   | inicio     | fin        | sprint goal     | estado  |
      | 1  | Sprint 1 | 2026-10-01 | 2026-10-08 | Backlog inicial | CERRADO |
      | 2  | Sprint 2 | 2026-10-09 | 2026-10-16 | Estimación      | CERRADO |
      | 3  | Sprint 3 | 2026-10-17 | 2026-10-24 | Métricas        | ABIERTO |

  # Caso alternativo - CA-14.1
  Escenario: Solo se listan los sprints del proyecto consultado (CA-14.1)
    Dado que existe un proyecto con nombre "Otro proyecto"
    Y que existe un sprint cerrado "Sprint 1" del "2026-10-01" al "2026-10-08" con sprint goal "Backlog inicial" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint A" del "2026-10-01" al "2026-10-08" con sprint goal "Otro goal" en el proyecto "Otro proyecto"
    Y que existe un sprint abierto "Sprint 2" del "2026-10-09" al "2026-10-16" con sprint goal "Estimación" en el proyecto "Sistema de ventas"
    Cuando consulto los sprints del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo estos sprints:
      | id | nombre   | inicio     | fin        | sprint goal     | estado  |
      | 1  | Sprint 1 | 2026-10-01 | 2026-10-08 | Backlog inicial | CERRADO |
      | 3  | Sprint 2 | 2026-10-09 | 2026-10-16 | Estimación      | ABIERTO |

  # Caso límite - CA-14.2
  Escenario: Un proyecto sin sprints devuelve una lista vacía (CA-14.2)
    Cuando consulto los sprints del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo una lista vacía de sprints

  # Caso alternativo - CA-14.3
  Escenario: Consultar un sprint cerrado de la lista con sus ítems completados (CA-14.3)
    Dado que existe un sprint cerrado "Sprint 1" del "2026-10-01" al "2026-10-08" con sprint goal "Backlog inicial" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "TERMINADO" asignado al sprint con el ID 1 en el proyecto "Sistema de ventas"
    Y que existe un ítem "Listar proyectos" en estado "EN_CURSO" asignado al sprint con el ID 1 en el proyecto "Sistema de ventas"
    Y que existe un ítem "Crear sprint" en estado "TERMINADO" asignado al sprint con el ID 1 en el proyecto "Sistema de ventas"
    Cuando consulto los sprints del proyecto "Sistema de ventas"
    Y consulto los ítems completados del primer sprint cerrado de la lista
    Entonces la consulta de ítems completados es exitosa
    Y obtengo los ítems completados "Login de usuarios" y "Crear sprint"

  # Caso límite - CA-14.3
  Escenario: Un sprint cerrado sin ítems completados aparece en la lista (CA-14.3)
    Dado que existe un sprint cerrado "Sprint 1" del "2026-10-01" al "2026-10-08" con sprint goal "Backlog inicial" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Listar proyectos" en estado "PENDIENTE" asignado al sprint con el ID 1 en el proyecto "Sistema de ventas"
    Cuando consulto los sprints del proyecto "Sistema de ventas"
    Y consulto los ítems completados del primer sprint cerrado de la lista
    Entonces la consulta de ítems completados es exitosa
    Y obtengo una lista vacía de ítems completados

  # Caso límite - CA-14.5
  Escenario: Modificar la lista devuelta no cambia los sprints guardados (CA-14.5)
    Dado que existe un sprint abierto "Sprint 1" del "2026-10-01" al "2026-10-08" con sprint goal "Backlog inicial" en el proyecto "Sistema de ventas"
    Cuando consulto los sprints del proyecto "Sistema de ventas"
    Y cambio el nombre y el estado del primer sprint devuelto
    Y consulto los sprints del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo estos sprints:
      | id | nombre   | inicio     | fin        | sprint goal     | estado  |
      | 1  | Sprint 1 | 2026-10-01 | 2026-10-08 | Backlog inicial | ABIERTO |

  # Caso de error - CA-14.4
  Esquema del escenario: Un proyecto inexistente no tiene sprints que listar (CA-14.4)
    Cuando consulto los sprints del proyecto con el ID <id>
    Entonces la consulta falla porque no existe un proyecto con ese id

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  # Caso de error - RN-14.7
  Escenario: Una falla del repositorio de sprints se informa sin lista parcial (RN-14.7)
    Dado que existe un sprint abierto "Sprint 1" del "2026-10-01" al "2026-10-08" con sprint goal "Backlog inicial" en el proyecto "Sistema de ventas"
    Y que el repositorio de sprints falla al listar
    Cuando consulto los sprints del proyecto "Sistema de ventas"
    Entonces la consulta falla con el error del repositorio
