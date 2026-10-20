# language: es
# HU-13 · Cerrar sprint (Issue #27)
# Especificación: docs/especificaciones/HU-13-cerrar-sprint.md

Característica: HU-13 Cerrar sprint
  Como Agile Enabler,
  quiero cerrar un sprint,
  para dar por terminada la iteración sin que se pueda modificar después.

  # Caso normal - CA-13.1
  Escenario: Cerrar un sprint abierto (CA-13.1)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando cierro el sprint con el ID 1
    Entonces el cierre es exitoso
    Y el sprint con el ID 1 queda en estado "CERRADO"
    Y el sprint guardado con el ID 1 conserva el nombre "Sprint 1", el sprint goal "Entregar el MVP" y el proyecto "Sistema de ventas"

  # Caso límite - CA-13.1
  Escenario: Cerrar un sprint sin ítems (CA-13.1)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando cierro el sprint con el ID 1
    Entonces el cierre es exitoso
    Y el sprint con el ID 1 queda en estado "CERRADO"

  # Caso alternativo - CA-13.1
  Escenario: Crear un nuevo sprint después del cierre (CA-13.1)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando cierro el sprint con el ID 1
    Y creo el sprint "Sprint 2" con inicio "2026-10-23", fin "2026-10-29" y sprint goal "Entregar el MVP" en el proyecto "Sistema de ventas"
    Entonces el cierre es exitoso
    Y la creación es exitosa
    Y el sprint con el ID 2 queda en estado "ABIERTO"

  # Caso límite - CA-13.7
  Escenario: Cerrar un sprint con ítems pendientes y en curso (CA-13.7)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "PENDIENTE" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Listar proyectos" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Y que el ítem con el ID 2 está asignado al sprint con el ID 1
    Cuando cierro el sprint con el ID 1
    Entonces el cierre es exitoso
    Y el ítem con el ID 1 queda en estado "PENDIENTE"
    Y el ítem con el ID 2 queda en estado "EN_CURSO"
    Y el ítem con el ID 1 sigue asignado al sprint con el ID 1
    Y el ítem con el ID 2 sigue asignado al sprint con el ID 1

  # Caso de error - CA-13.2
  Escenario: Un sprint cerrado no admite asignaciones ni completados (CA-13.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Listar proyectos" en estado "PENDIENTE" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando cierro el sprint con el ID 1
    Y completo el ítem con el ID 1 del sprint con el ID 1
    Y asigno el ítem con el ID 2 al sprint con el ID 1
    Entonces el cierre es exitoso
    Y el completado falla porque el sprint está cerrado
    Y la asignación falla porque el sprint está cerrado
    Y el ítem con el ID 1 queda en estado "EN_CURSO"
    Y el ítem con el ID 2 sigue sin sprint asignado

  # Caso de error - CA-13.3
  Escenario: Cerrar un sprint ya cerrado (CA-13.3)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando cierro el sprint con el ID 1
    Entonces el cierre falla porque el sprint está cerrado
    Y el sprint con el ID 1 sigue en estado "CERRADO"

  # Caso de error - CA-13.4
  Escenario: Cerrar un sprint inexistente (CA-13.4)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando cierro el sprint con el ID 999
    Entonces el cierre falla porque no existe un sprint con ese id

  # Caso de error - CA-13.5
  Escenario: Error del repositorio al cerrar (CA-13.5)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el repositorio de sprints falla al actualizar
    Cuando cierro el sprint con el ID 1
    Entonces el cierre falla con el error del repositorio
    Y el sprint con el ID 1 sigue en estado "ABIERTO"

  # Caso límite - CA-13.6
  Escenario: Modificar el sprint devuelto no altera el guardado (CA-13.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando cierro el sprint con el ID 1
    Y cambio el nombre, el sprint goal y las fechas del sprint devuelto
    Entonces al reconsultar el sprint con el ID 1 obtengo el nombre "Sprint 1", el sprint goal "Entregar el MVP" y el estado "CERRADO"

  # Caso límite - CA-13.8
  Escenario: Cierre simultáneo con asignación y completado (CA-13.8)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Listar proyectos" en estado "PENDIENTE" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando cierro el sprint y en simultáneo completo el ítem con el ID 1 y asigno el ítem con el ID 2 al sprint con el ID 1
    Entonces el sprint con el ID 1 queda en estado "CERRADO"
    Y cada operación simultánea termina bien o porque el sprint está cerrado
    Y ninguna escritura de ítems ocurre después del cierre
