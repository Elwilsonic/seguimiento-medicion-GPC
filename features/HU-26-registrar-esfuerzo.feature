# language: es
# HU-26 · Registrar esfuerzo (Issue #44)
# Especificación: docs/especificaciones/HU-26-registrar-esfuerzo.md

Característica: HU-26 Registrar esfuerzo
  Como Product Builder,
  quiero registrar las horas que dediqué a una historia de un sprint,
  para medir el esfuerzo real.

  # Caso normal - CA-26.1
  Escenario: Registrar un esfuerzo válido (CA-26.1)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2.5 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro es exitoso
    Y el esfuerzo guardado tiene el ID 1
    Y el esfuerzo guardado conserva todos los datos del registro

  # Caso límite - CA-26.2
  Escenario: Registrar un esfuerzo con horas fraccionadas (CA-26.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Ajustar el formulario" y horas 0.25 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro es exitoso
    Y el esfuerzo guardado conserva todos los datos del registro

  # Caso alternativo - CA-26.1 y RN-26.6
  Escenario: Registrar un esfuerzo sobre un ítem terminado del sprint (CA-26.1)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "TERMINADO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Verificar el login" y horas 1 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro es exitoso
    Y el esfuerzo guardado conserva todos los datos del registro

  # Caso de error - CA-26.2
  Escenario: Rechazar horas iguales a cero (CA-26.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 0 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque las horas no son válidas
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.2
  Escenario: Rechazar horas negativas (CA-26.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas -1 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque las horas no son válidas
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.2
  Escenario: Rechazar horas no finitas (CA-26.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas NaN para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque las horas no son válidas
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.2
  Escenario: Rechazar horas infinitas (CA-26.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas +Inf para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque las horas no son válidas
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.3
  Escenario: Rechazar un ítem sin sprint asignado (CA-26.3)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque el ítem no está asignado al sprint
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.3
  Escenario: Rechazar un ítem asignado a otro sprint (CA-26.3)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 2
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque el ítem no está asignado al sprint
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.4
  Escenario: Rechazar un sprint inexistente (CA-26.4)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 999
    Entonces el registro falla porque no existe un sprint con ese id
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.4
  Escenario: Rechazar un ítem inexistente (CA-26.4)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 999 del sprint con el ID 1
    Entonces el registro falla porque no existe un ítem del backlog con ese id
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.5
  Escenario: Rechazar un sprint cerrado (CA-26.5)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque el sprint está cerrado
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.6
  Escenario: Rechazar un integrante vacío (CA-26.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque el integrante es obligatorio
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.6
  Escenario: Rechazar una actividad vacía (CA-26.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque la actividad es obligatoria
    Y no se guardó ningún esfuerzo

  # Caso de error - CA-26.6
  Escenario: Rechazar una fecha ausente (CA-26.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla porque la fecha del esfuerzo es obligatoria
    Y no se guardó ningún esfuerzo

  # Caso límite - CA-26.7
  Escenario: No modificar el ítem al registrar (CA-26.7)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro es exitoso
    Y el ítem con el ID 1 sigue en estado "EN_CURSO" y con los mismos datos

  # Caso límite - CA-26.8
  Escenario: Modificar el esfuerzo devuelto no altera el guardado (CA-26.8)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Y cambio todos los datos del esfuerzo devuelto
    Entonces el esfuerzo guardado conserva todos los datos del registro

  # Caso de error - CA-26.9
  Escenario: Error del repositorio al guardar (CA-26.9)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Y que el repositorio de esfuerzos falla al guardar
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Entonces el registro falla con el error del repositorio
    Y no se guardó ningún esfuerzo

  # Caso límite - CA-26.10
  Escenario: Registrar en simultáneo varios esfuerzos (CA-26.10)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro en simultáneo 10 esfuerzos del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1
    Entonces todos los registros son exitosos
    Y los esfuerzos guardados tienen IDs únicos

  # Caso límite - CA-26.10
  Escenario: Registrar y cerrar el sprint en simultáneo (CA-26.10)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro un esfuerzo del integrante "García Santiago" con fecha "2026-10-24", actividad "Implementar el registro" y horas 2 para el ítem con el ID 1 del sprint con el ID 1 y en simultáneo cierro el sprint con el ID 1
    Entonces el sprint con el ID 1 queda en estado "CERRADO"
    Y cada registro simultáneo termina bien o porque el sprint está cerrado
    Y ningún esfuerzo se guarda después del cierre
