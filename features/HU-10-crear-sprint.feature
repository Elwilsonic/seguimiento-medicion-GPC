# language: es
# HU-10 · Crear sprint con nombre, fechas y Sprint Goal (Issue #24)
# Especificación: docs/especificaciones/HU-10-crear-sprint.md

Característica: HU-10 Crear sprint con nombre, fechas y Sprint Goal
  Como Agile Enabler,
  quiero crear un sprint con nombre, fechas y Sprint Goal,
  para planificar una iteración del proyecto.

  # Caso normal (happy path) - CA-10.1, RN-10.6 y RN-10.7
  Escenario: Crear un sprint en un proyecto existente
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo el sprint "Sprint 1" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Entregar el MVP" en el proyecto "Sistema de ventas"
    Entonces el sprint se crea con el ID 1
    Y el sprint queda en el estado "ABIERTO"
    Y el sprint tiene el nombre "Sprint 1"
    Y el sprint tiene el sprint goal "Entregar el MVP"
    Y el sprint tiene inicio "2026-10-16" y fin "2026-10-22"

  # Caso normal - RN-10.1 y RN-10.2
  Escenario: El nombre y el sprint goal se guardan sin los espacios de los extremos
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo el sprint "  Sprint 1  " con inicio "2026-10-16", fin "2026-10-22" y sprint goal "  Entregar el MVP  " en el proyecto "Sistema de ventas"
    Entonces el sprint se crea con el ID 1
    Y el sprint tiene el nombre "Sprint 1"
    Y el sprint tiene el sprint goal "Entregar el MVP"

  # Caso alternativo - RN-10.5 y RN-10.1
  Escenario: Un sprint abierto en otro proyecto no impide crear uno con el mismo nombre
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un proyecto con nombre "Portal de clientes"
    Cuando creo el sprint "Sprint 1" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Entregar el MVP" en el proyecto "Sistema de ventas"
    Y creo el sprint "Sprint 1" con inicio "2026-10-23", fin "2026-10-29" y sprint goal "Cerrar E2" en el proyecto "Portal de clientes"
    Entonces los sprints se crean con los IDs 1 y 2

  # Caso límite - RN-10.1
  Escenario: El nombre puede repetirse si el sprint anterior del proyecto está cerrado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando creo el sprint "Sprint 1" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Entregar el MVP" en el proyecto "Sistema de ventas"
    Entonces el sprint se crea con el ID 2
    Y el sprint queda en el estado "ABIERTO"

  # Caso límite - RN-10.4
  Escenario: Un proyecto finalizado acepta sprints
    Dado que existe un proyecto finalizado con nombre "Proyecto cerrado"
    Cuando creo el sprint "Sprint 1" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Entregar el MVP" en el proyecto "Proyecto cerrado"
    Entonces el sprint se crea con el ID 1
    Y el sprint queda en el estado "ABIERTO"

  # Caso límite - CA-10.7 y RN-10.8
  Escenario: Modificar el sprint devuelto no altera el sprint guardado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo el sprint "Sprint 1" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Entregar el MVP" en el proyecto "Sistema de ventas"
    Y modifico el nombre, el sprint goal y las fechas del sprint devuelto
    Entonces el sprint guardado conserva el nombre "Sprint 1"
    Y el sprint guardado conserva el sprint goal "Entregar el MVP"
    Y el sprint guardado conserva el estado "ABIERTO"

  # Caso límite - CA-10.6
  Escenario: Creaciones simultáneas en el mismo proyecto dejan un solo sprint abierto
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando se crean 10 sprints en simultáneo en el proyecto "Sistema de ventas"
    Entonces exactamente una creación tiene éxito
    Y las demás fallan porque ya existe un sprint abierto en el proyecto
    Y queda un solo sprint ABIERTO guardado

  # Caso de error - CA-10.4 y RN-10.1
  Escenario: El nombre del sprint es obligatorio
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo el sprint "   " con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Entregar el MVP" en el proyecto "Sistema de ventas"
    Entonces la creación falla porque el nombre del sprint es obligatorio

  # Caso de error - CA-10.4 y RN-10.2
  Escenario: El sprint goal es obligatorio
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo el sprint "Sprint 1" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "   " en el proyecto "Sistema de ventas"
    Entonces la creación falla porque el sprint goal es obligatorio

  # Caso de error - CA-10.2 y RN-10.3
  Esquema del escenario: Crear un sprint con fechas inválidas
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo el sprint "Sprint 1" con inicio "<inicio>", fin "<fin>" y sprint goal "Entregar el MVP" en el proyecto "Sistema de ventas"
    Entonces la creación falla porque <motivo>

    Ejemplos:
      | inicio     | fin        | motivo                                         |
      | 2026-10-16 | 2026-10-16 | la fecha de fin no es posterior a la de inicio  |
      | 2026-10-23 | 2026-10-16 | la fecha de fin no es posterior a la de inicio  |
      | ausente    | 2026-10-22 | la fecha de inicio del sprint es obligatoria   |
      | 2026-10-16 | ausente    | la fecha de fin del sprint es obligatoria       |

  # Caso de error - CA-10.3 y RN-10.5
  Escenario: Ya existe un sprint abierto en el proyecto
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando creo el sprint "Sprint 2" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Otro objetivo" en el proyecto "Sistema de ventas"
    Entonces la creación falla porque ya existe un sprint abierto en el proyecto

  # Caso de error - CA-10.5
  Escenario: Crear un sprint en un proyecto que no existe
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo el sprint "Sprint 1" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Entregar el MVP" en el proyecto con ID 999
    Entonces la creación falla porque el proyecto no existe

  # Caso de error - CA-10.5
  Escenario: Crear un sprint cuando no hay proyectos creados
    Dado que no hay proyectos creados
    Cuando creo el sprint "Sprint 1" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Entregar el MVP" en el proyecto con ID 1
    Entonces la creación falla porque el proyecto no existe

  # Caso de error - CA-10.5
  Esquema del escenario: Crear un sprint con un ID imposible
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo el sprint "Sprint 1" con inicio "2026-10-16", fin "2026-10-22" y sprint goal "Entregar el MVP" en el proyecto con ID <id>
    Entonces la creación falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 0  |
      | -1 |
