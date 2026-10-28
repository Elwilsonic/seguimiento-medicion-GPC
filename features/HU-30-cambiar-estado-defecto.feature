# language: es
# HU-30 · Cambiar el estado de un defecto (Issue #48)
# Especificación: docs/especificaciones/HU-30-cambiar-estado-defecto.md
# CA-30.9 (API e interfaz web) se cubre con tests de internal/httpapi, no con escenarios godog.
# Los defectos de partida se guardan con Save del repositorio en memoria: registrarlos es HU-29.

Característica: HU-30 Cambiar el estado de un defecto
  Como Product Builder,
  quiero cambiar el estado de un defecto y registrar en qué sprint se resolvió,
  para reflejar si está abierto, en corrección o resuelto.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"

  # Caso normal - CA-30.1
  Escenario: Pasar un defecto abierto a en corrección (CA-30.1)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando cambio el estado del defecto con el ID 1 a "EN_CORRECCION"
    Entonces el cambio de estado del defecto es exitoso
    Y el defecto devuelto está en estado "EN_CORRECCION"
    Y el defecto guardado con el ID 1 está en estado "EN_CORRECCION"
    Y el defecto guardado con el ID 1 no tiene sprint de resolución

  # Caso normal - CA-30.1 y CA-30.2
  Escenario: Resolver un defecto en corrección en el sprint abierto (CA-30.1 y CA-30.2)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Cuando cambio el estado del defecto con el ID 1 a "RESUELTO" con el sprint de resolución con el ID 1
    Entonces el cambio de estado del defecto es exitoso
    Y el defecto devuelto está en estado "RESUELTO"
    Y el defecto guardado con el ID 1 está en estado "RESUELTO"
    Y el defecto guardado con el ID 1 tiene el sprint de resolución con el ID 1

  # Caso alternativo - CA-30.2
  Escenario: Resolver en un sprint posterior al de detección (CA-30.2)
    Dado que el sprint con el ID 1 está cerrado
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Cuando cambio el estado del defecto con el ID 1 a "RESUELTO" con el sprint de resolución con el ID 2
    Entonces el cambio de estado del defecto es exitoso
    Y el defecto guardado con el ID 1 está en estado "RESUELTO"
    Y el defecto guardado con el ID 1 tiene el sprint de resolución con el ID 2

  # Caso alternativo - CA-30.1
  Escenario: Volver un defecto en corrección a abierto (CA-30.1)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Cuando cambio el estado del defecto con el ID 1 a "ABIERTO"
    Entonces el cambio de estado del defecto es exitoso
    Y el defecto guardado con el ID 1 está en estado "ABIERTO"
    Y el defecto guardado con el ID 1 no tiene sprint de resolución

  # Caso alternativo - CA-30.1 y CA-30.2
  Escenario: Reabrir un defecto resuelto limpia el sprint de resolución (CA-30.1 y CA-30.2)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "RESUELTO"
    Cuando cambio el estado del defecto con el ID 1 a "EN_CORRECCION"
    Entonces el cambio de estado del defecto es exitoso
    Y el defecto guardado con el ID 1 está en estado "EN_CORRECCION"
    Y el defecto guardado con el ID 1 no tiene sprint de resolución

  # Caso de error - CA-30.1
  Esquema del escenario: Rechazar abierto a resuelto y resuelto a abierto (CA-30.1)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "<origen>"
    Cuando cambio el estado del defecto con el ID 1 a "<destino>" con el sprint de resolución con el ID 1
    Entonces el cambio de estado del defecto falla porque la transición no está permitida
    Y el defecto guardado con el ID 1 no cambió

    Ejemplos:
      | origen   | destino  |
      | ABIERTO  | RESUELTO |
      | RESUELTO | ABIERTO  |

  # Caso límite - CA-30.1 (un defecto RESUELTO de partida está resuelto en el sprint 1)
  Esquema del escenario: Pedir el estado actual no cambia nada (CA-30.1)
    Dado que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "<estado>"
    Cuando cambio el estado del defecto con el ID 1 a "<estado>" con el sprint de resolución con el ID <sprint>
    Entonces el cambio de estado del defecto es exitoso
    Y el defecto devuelto está en estado "<estado>"
    Y el defecto guardado con el ID 1 no cambió
    Y no se actualizó el repositorio de defectos

    Ejemplos:
      | estado        | sprint |
      | ABIERTO       | 0      |
      | EN_CORRECCION | 0      |
      | RESUELTO      | 1      |
      | RESUELTO      | 2      |

  # Caso de error - CA-30.3
  Esquema del escenario: Rechazar un estado inválido (CA-30.3)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando cambio el estado del defecto con el ID 1 a "<estado>"
    Entonces el cambio de estado del defecto falla porque el estado no es válido
    Y el defecto guardado con el ID 1 no cambió

    Ejemplos:
      | estado    |
      |           |
      | CERRADO   |
      | BLOQUEADO |

  # Caso límite - CA-30.3 (Gherkin recorta las celdas: los espacios se prueban en status_test.go)
  Esquema del escenario: Normalizar el estado (CA-30.3)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "<origen>"
    Cuando cambio el estado del defecto con el ID 1 a "<estado>" con el sprint de resolución con el ID 1
    Entonces el cambio de estado del defecto es exitoso
    Y el defecto guardado con el ID 1 está en estado "<guardado>"

    Ejemplos:
      | origen        | estado         | guardado      |
      | ABIERTO       | en_correccion | EN_CORRECCION |
      | EN_CORRECCION | Resuelto      | RESUELTO      |

  # Caso de error - CA-30.4
  Esquema del escenario: Rechazar un defecto inexistente (CA-30.4)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando cambio el estado del defecto con el ID <id> a "EN_CORRECCION"
    Entonces el cambio de estado del defecto falla porque no existe el defecto
    Y el defecto guardado con el ID 1 no cambió

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  # Caso de error - CA-30.2 (el sprint <= 0 se valida antes de leer el defecto, también si ya está RESUELTO)
  Esquema del escenario: Rechazar resolver sin sprint de resolución (CA-30.2)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "<origen>"
    Cuando cambio el estado del defecto con el ID 1 a "RESUELTO" con el sprint de resolución con el ID <sprint>
    Entonces el cambio de estado del defecto falla porque el sprint de resolución es obligatorio
    Y el defecto guardado con el ID 1 no cambió

    Ejemplos:
      | origen        | sprint |
      | EN_CORRECCION | 0      |
      | EN_CORRECCION | -1     |
      | RESUELTO      | 0      |
      | RESUELTO      | -1     |

  # Caso de error - CA-30.2
  Escenario: Rechazar resolver con el sprint de resolución ausente (CA-30.2)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Cuando cambio el estado del defecto con el ID 1 a "RESUELTO" sin sprint de resolución
    Entonces el cambio de estado del defecto falla porque el sprint de resolución es obligatorio
    Y el defecto guardado con el ID 1 no cambió

  # Caso de error - CA-30.2 (sprint 2 cerrado del proyecto, sprint 3 abierto de otro proyecto)
  Esquema del escenario: Rechazar un sprint de resolución inexistente, cerrado o de otro proyecto (CA-30.2)
    Dado que existe un proyecto con nombre "Sistema de facturación"
    Y que existe un sprint cerrado "Sprint 2" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de facturación"
    Y que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Cuando cambio el estado del defecto con el ID 1 a "RESUELTO" con el sprint de resolución con el ID <sprint>
    Entonces el cambio de estado del defecto falla porque <motivo>
    Y el defecto guardado con el ID 1 no cambió

    Ejemplos:
      | sprint | motivo                                     |
      | 999    | no existe un sprint con ese id             |
      | 2      | el sprint está cerrado                     |
      | 3      | el sprint de resolución es de otro proyecto |

  # Caso de error - CA-30.2. Validación defensiva: desde la API no se puede producir,
  # así que el defecto se guarda con Save, detectado en el sprint 2 y resuelto en el 1.
  Escenario: Rechazar un sprint de resolución anterior al de detección (CA-30.2)
    Dado que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 2 en estado "EN_CORRECCION"
    Cuando cambio el estado del defecto con el ID 1 a "RESUELTO" con el sprint de resolución con el ID 1
    Entonces el cambio de estado del defecto falla porque el sprint de resolución es anterior al de detección
    Y el defecto guardado con el ID 1 no cambió

  # Caso límite - CA-30.5
  Escenario: No modificar la historia, los sprints ni los demás campos (CA-30.5)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Cuando cambio el estado del defecto con el ID 1 a "RESUELTO" con el sprint de resolución con el ID 1
    Entonces el cambio de estado del defecto es exitoso
    Y el defecto guardado con el ID 1 está en estado "RESUELTO"
    Y el defecto guardado con el ID 1 conserva la descripción, la severidad, el ítem y el sprint de detección
    Y el ítem con el ID 1 y el sprint con el ID 1 no cambiaron

  # Caso límite - CA-30.6
  Escenario: Modificar el defecto devuelto no altera el guardado (CA-30.6)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando cambio el estado del defecto con el ID 1 a "EN_CORRECCION"
    Entonces el cambio de estado del defecto es exitoso
    Y modifico todos los datos del defecto devuelto
    Y el defecto guardado con el ID 1 está en estado "EN_CORRECCION"
    Y el defecto guardado con el ID 1 conserva la descripción, la severidad, el ítem y el sprint de detección

  # Caso de error - CA-30.8
  Esquema del escenario: Error del repositorio (CA-30.8)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Y que el repositorio de defectos falla al <operacion>
    Cuando cambio el estado del defecto con el ID 1 a "EN_CORRECCION"
    Entonces el cambio de estado del defecto falla con el error del repositorio
    Y el defecto guardado con el ID 1 no cambió

    Ejemplos:
      | operacion  |
      | buscar     |
      | actualizar |

  # Caso límite - CA-30.7
  Escenario: Cambios simultáneos sobre el mismo defecto (CA-30.7)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando cambio en simultáneo 10 veces el estado del defecto con el ID 1 a "EN_CORRECCION"
    Entonces todos los cambios de estado del defecto son exitosos
    Y el defecto guardado con el ID 1 está en estado "EN_CORRECCION"

  # Caso límite - CA-30.7
  Escenario: Resolver y cerrar el sprint en simultáneo (CA-30.7)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Cuando cambio el estado del defecto con el ID 1 a "RESUELTO" con el sprint de resolución con el ID 1 y en simultáneo cierro el sprint con el ID 1
    Entonces el sprint con el ID 1 queda en estado "CERRADO"
    Y el cambio simultáneo termina bien o porque el sprint está cerrado
    Y el defecto queda resuelto solo si el cambio terminó bien
