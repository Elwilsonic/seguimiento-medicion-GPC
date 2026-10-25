# language: es
# HU-29 · Registrar un defecto (Issue #47)
# Especificación: docs/especificaciones/HU-29-registrar-defecto.md
# CA-29.11 (API e interfaz web) se cubre con tests de internal/httpapi, no con escenarios godog.

Característica: HU-29 Registrar un defecto
  Como Product Builder,
  quiero registrar un defecto asociado a una historia,
  para llevar el control de la calidad.

  # Caso normal - CA-29.1 y CA-29.3
  Escenario: Registrar un defecto válido en el sprint abierto actual (CA-29.1 y CA-29.3)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "  El total no descuenta el descuento  ", severidad " alta " para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto es exitoso
    Y el defecto guardado tiene el ID 1
    Y el defecto guardado tiene la descripción "El total no descuenta el descuento"
    Y el defecto guardado tiene la severidad "ALTA"
    Y el defecto guardado está en estado "ABIERTO"
    Y el defecto guardado tiene el ítem con el ID 1 y el sprint de detección con el ID 1
    Y el defecto guardado no tiene sprint de resolución

  # Caso alternativo - CA-29.5
  Escenario: Registrar sobre una historia terminada (CA-29.5)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "TERMINADO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "Falla al cerrar sesión", severidad "MEDIA" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto es exitoso
    Y el ítem con el ID 1 sigue en estado "TERMINADO" y con los mismos datos

  # Caso límite - CA-29.6
  Escenario: Registrar sobre una historia de otro sprint (CA-29.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 2
    Cuando registro un defecto con descripción "Falla al iniciar sesión", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto es exitoso
    Y el ítem con el ID 1 sigue asignado al sprint con el ID 2

  # Caso de error - CA-29.2 y CA-29.6
  Escenario: Rechazar sprint de detección ausente o 0 (CA-29.2 y CA-29.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 sin sprint de detección
    Entonces el registro del defecto falla porque el sprint de detección es obligatorio
    Y no se guardó ningún defecto
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 0
    Entonces el registro del defecto falla porque el sprint de detección es obligatorio
    Y no se guardó ningún defecto

  # Caso de error - CA-29.6
  Escenario: Rechazar sprint de detección negativo (CA-29.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID -1
    Entonces el registro del defecto falla porque el sprint de detección es obligatorio
    Y no se guardó ningún defecto

  # Caso de error - CA-29.6
  Escenario: Rechazar un sprint de detección cerrado (CA-29.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto falla porque el sprint está cerrado
    Y no se guardó ningún defecto

  # Caso de error - CA-29.2
  Escenario: Rechazar descripción vacía (CA-29.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto falla porque la descripción es obligatoria
    Y no se guardó ningún defecto

  # Caso de error - CA-29.4
  Escenario: Rechazar severidad inválida (CA-29.4)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "URGENTE" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto falla porque la severidad no es válida
    Y no se guardó ningún defecto

  # Caso de error - CA-29.5
  Escenario: Rechazar historia inexistente (CA-29.5)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 999 en el sprint de detección con el ID 1
    Entonces el registro del defecto falla porque la historia no existe
    Y no se guardó ningún defecto

  # Caso de error - CA-29.6
  Escenario: Rechazar sprint de detección positivo inexistente (CA-29.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 999
    Entonces el registro del defecto falla porque no existe un sprint con ese id
    Y no se guardó ningún defecto

  # Caso de error - CA-29.6
  Escenario: Rechazar sprint de detección de otro proyecto (CA-29.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un proyecto con nombre "Sistema de facturación"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de facturación"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto falla porque el sprint de detección es de otro proyecto
    Y no se guardó ningún defecto

  # Caso límite - CA-29.7
  Escenario: No modificar la historia ni el sprint (CA-29.7)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto es exitoso
    Y el ítem con el ID 1 sigue en estado "EN_CURSO" y con los mismos datos
    Y el sprint con el ID 1 sigue en estado "ABIERTO" y con los mismos datos

  # Caso límite - CA-29.8
  Escenario: Modificar el defecto devuelto no altera el guardado (CA-29.8)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto es exitoso
    Y modifico todos los datos del defecto devuelto
    Y el defecto guardado conserva todos los datos del registro

  # Caso de error - CA-29.10
  Escenario: Error del repositorio al guardar (CA-29.10)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el repositorio de defectos falla al guardar
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces el registro del defecto falla con el error del repositorio
    Y no se guardó ningún defecto

  # Caso límite - CA-29.9
  Escenario: Registrar en simultáneo varios defectos (CA-29.9)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro en simultáneo 10 defectos con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 1
    Entonces todos los registros de defectos son exitosos
    Y los defectos guardados tienen IDs únicos

  # Caso límite - CA-29.9
  Escenario: Registrar y cerrar el sprint en simultáneo (CA-29.9)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en estado "EN_CURSO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando registro un defecto con descripción "El total no descuenta el descuento", severidad "ALTA" para el ítem con el ID 1 en el sprint de detección con el ID 1 y en simultáneo cierro el sprint con el ID 1
    Entonces el sprint con el ID 1 queda en estado "CERRADO"
    Y cada registro simultáneo de defecto termina bien o porque el sprint está cerrado
    Y ningún defecto se guarda después del cierre
