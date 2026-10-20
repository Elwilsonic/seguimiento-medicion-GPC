# language: es
# HU-12 · Registrar historias completadas (Issue #26)
# Especificación: docs/especificaciones/HU-12-registrar-completadas.md

Característica: HU-12 Registrar historias completadas
  Como Product Builder,
  quiero marcar como completadas las historias de un sprint,
  para saber qué se terminó.

  # Caso normal - CA-12.1 y RN-12.1
  Escenario: Completar un ítem en curso asignado a un sprint abierto
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando completo el ítem con el ID 1 del sprint con el ID 1
    Entonces el completado es exitoso
    Y el ítem con el ID 1 queda en estado "TERMINADO"
    Y el ítem guardado con el ID 1 conserva el título "Login de usuarios"
    Y el ítem guardado con el ID 1 conserva la descripción "Formulario de acceso" y los criterios "Login OK"

  # Caso alternativo - CA-12.4 y RN-12.2
  Escenario: Completar un ítem ya terminado es un no-op
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "TERMINADO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando completo el ítem con el ID 1 del sprint con el ID 1
    Entonces el completado es exitoso
    Y el ítem con el ID 1 queda en estado "TERMINADO"
    Y el ítem guardado con el ID 1 conserva el título "Login de usuarios"

  # Caso normal - CA-12.3 y RN-12.6
  Escenario: Consultar solo los completados de un sprint, ordenados por ID
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Listar proyectos" en estado "TERMINADO" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Reporte de avance" en estado "PENDIENTE" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Cierre del sprint" en estado "TERMINADO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Y que el ítem con el ID 2 está asignado al sprint con el ID 1
    Y que el ítem con el ID 3 está asignado al sprint con el ID 1
    Y que el ítem con el ID 4 está asignado al sprint con el ID 1
    Cuando consulto los completados del sprint con el ID 1
    Entonces la consulta es exitosa
    Y la lista devuelta tiene 2 ítems
    Y la lista contiene el ítem con el ID 2
    Y la lista contiene el ítem con el ID 4
    Y la lista no contiene el ítem con el ID 1
    Y la lista no contiene el ítem con el ID 3
    Y la lista está ordenada por ID ascendente

  # Caso límite - CA-12.3
  Escenario: Un sprint sin ítems completados devuelve una lista vacía
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando consulto los completados del sprint con el ID 1
    Entonces la consulta es exitosa
    Y la lista devuelta está vacía

  # Caso límite - RN-12.4
  Escenario: Consultar los completados de un sprint cerrado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "TERMINADO" en el proyecto "Sistema de ventas"
    Y que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando consulto los completados del sprint con el ID 1
    Entonces la consulta es exitosa
    Y la lista devuelta tiene 1 ítem
    Y la lista contiene el ítem con el ID 1

  # Caso de error - CA-12.2 y RN-12.3
  Escenario: No se puede completar un ítem sin asignar
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando completo el ítem con el ID 1 del sprint con el ID 1
    Entonces el completado falla porque el ítem no está asignado al sprint
    Y el ítem con el ID 1 queda en estado "EN_CURSO"

  # Caso de error - CA-12.2 y RN-12.3
  Escenario: No se puede completar un ítem asignado a otro sprint
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 2
    Cuando completo el ítem con el ID 1 del sprint con el ID 1
    Entonces el completado falla porque el ítem no está asignado al sprint
    Y el ítem con el ID 1 queda en estado "EN_CURSO"

  # Caso de error - CA-12.6 y RN-12.2
  Escenario: No se puede completar un ítem PENDIENTE directamente
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "PENDIENTE" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando completo el ítem con el ID 1 del sprint con el ID 1
    Entonces el completado falla porque la transición de estado no está permitida
    Y el ítem con el ID 1 queda en estado "PENDIENTE"

  # Caso de error - CA-12.5 y RN-12.4
  Escenario: No se puede completar un ítem en un sprint cerrado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando completo el ítem con el ID 1 del sprint con el ID 1
    Entonces el completado falla porque el sprint está cerrado
    Y el ítem con el ID 1 queda en estado "EN_CURSO"

  # Caso de error - CA-12.5
  Escenario: No se puede completar un ítem inexistente
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando completo el ítem con el ID 999 del sprint con el ID 1
    Entonces el completado falla porque no existe un ítem del backlog con ese id

  # Caso de error - CA-12.5
  Escenario: No se puede completar en un sprint inexistente
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Cuando completo el ítem con el ID 1 del sprint con el ID 999
    Entonces el completado falla porque no existe un sprint con ese id

  # Caso límite - CA-12.7
  Escenario: Modificar el ítem devuelto no altera el guardado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando completo el ítem con el ID 1 del sprint con el ID 1
    Y cambio el título y los criterios del ítem completado devuelto
    Entonces al reconsultar el ítem con el ID 1 obtengo el título "Login de usuarios" y el estado "TERMINADO"

  # Caso límite - CA-12.9
  Escenario: Completar en simultáneo el mismo ítem deja el estado terminado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando completo en simultáneo 10 veces el ítem con el ID 1 del sprint con el ID 1
    Entonces todas las marcaciones tienen éxito
    Y el ítem con el ID 1 queda en estado "TERMINADO"
