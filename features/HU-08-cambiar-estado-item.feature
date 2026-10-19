# language: es
# HU-08 · Cambiar el estado de un ítem (Issue #20)
# Especificación: docs/especificaciones/HU-08-cambiar-estado-item.md

Característica: HU-08 Cambiar el estado de un ítem
  Como Product Builder,
  quiero cambiar el estado de un ítem del backlog,
  para reflejar su avance.

  # Caso normal (happy path) - CA-08.1
  Escenario: Cambiar un ítem de PENDIENTE a EN_CURSO
    Dado que existe un ítem con estado "PENDIENTE"
    Cuando cambio el estado del ítem con el ID 1 a "EN_CURSO"
    Entonces el cambio de estado es exitoso
    Y el ítem devuelto está en el estado "EN_CURSO"
    Y el ítem guardado con el ID 1 conserva los demás campos

  # Caso normal - CA-08.2
  Escenario: Cambiar un ítem de EN_CURSO a TERMINADO
    Dado que existe un ítem con estado "EN_CURSO"
    Cuando cambio el estado del ítem con el ID 1 a "TERMINADO"
    Entonces el cambio de estado es exitoso
    Y el ítem devuelto está en el estado "TERMINADO"
    Y el ítem guardado con el ID 1 está en el estado "TERMINADO"

  # Caso alternativo - CA-08.3 (transiciones permitidas: volver atrás y reabrir)
  Esquema del escenario: Volver atrás y reabrir un ítem
    Dado que existe un ítem con estado "<origen>"
    Cuando cambio el estado del ítem con el ID 1 a "<destino>"
    Entonces el cambio de estado es exitoso
    Y el ítem devuelto está en el estado "<destino>"
    Y el ítem guardado con el ID 1 está en el estado "<destino>"

    Ejemplos:
      | origen    | destino   |
      | EN_CURSO  | PENDIENTE |
      | TERMINADO | EN_CURSO  |

  # Caso de error - CA-08.4 (transiciones que saltean un estado)
  Esquema del escenario: Rechazar transiciones que saltean un estado
    Dado que existe un ítem con estado "<origen>"
    Cuando cambio el estado del ítem con el ID 1 a "<destino>"
    Entonces el cambio de estado falla porque la transición no está permitida
    Y el ítem guardado con el ID 1 sigue en el estado "<origen>"

    Ejemplos:
      | origen    | destino   |
      | PENDIENTE | TERMINADO |
      | TERMINADO | PENDIENTE |

  # Caso límite y de error - CA-08.5 (normalización y estados inválidos)
  Esquema del escenario: Normalizar los estados válidos y rechazar los inválidos
    Dado que existe un ítem con estado "PENDIENTE"
    Cuando cambio el estado del ítem con el ID 1 a "<estado>"
    Entonces el cambio de estado <resultado>
    Y el ítem guardado con el ID 1 está en el estado "<guardado>"

    Ejemplos:
      | caso            | estado     | resultado                           | guardado  |
      | estado inválido | BLOQUEADO  | falla porque el estado no es válido | PENDIENTE |
      | estado vacío    |            | falla porque el estado no es válido | PENDIENTE |
      | minúsculas      | en_curso   | es exitoso                          | EN_CURSO  |

  # Caso de error - CA-08.6 (ítem inexistente y sin ítems creados)
  Esquema del escenario: Cambiar el estado de un ítem que no existe
    Dado que hay <items> ítems creados en estado "PENDIENTE"
    Cuando cambio el estado del ítem con el ID <id> a "EN_CURSO"
    Entonces el cambio de estado falla porque el ítem no existe
    Y no se modificó ningún ítem guardado

    Ejemplos:
      | items | id  |
      | 1     | 999 |
      | 1     | 0   |
      | 1     | -1  |
      | 0     | 1   |

  # Caso límite - CA-08.7 (mismo estado: no-op y sin llamar a Update)
  Escenario: Cambiar al mismo estado no llama a Update
    Dado que existe un ítem con estado "PENDIENTE"
    Y que el repositorio falla al actualizar
    Cuando cambio el estado del ítem con el ID 1 a "Pendiente"
    Entonces el cambio de estado es exitoso
    Y el ítem devuelto está en el estado "PENDIENTE"

  # Caso límite - CA-08.8 (persistencia y copia por valor)
  Escenario: El cambio queda guardado y el ítem devuelto es una copia
    Dado que existe un ítem con estado "PENDIENTE"
    Cuando cambio el estado del ítem con el ID 1 a "EN_CURSO"
    Y modifico el estado y los criterios del ítem devuelto
    Entonces al reconsultar el ítem con el ID 1 obtengo el estado "EN_CURSO"
    Y los criterios guardados del ítem con el ID 1 no cambiaron

  # Caso límite - CA-08.9 (concurrencia)
  Escenario: Varios cambios simultáneos al mismo estado
    Dado que existe un ítem con estado "PENDIENTE"
    Cuando se cambian en simultáneo 50 veces el estado del ítem con el ID 1 a "EN_CURSO"
    Entonces ninguna operación falla
    Y el ítem guardado con el ID 1 está en el estado "EN_CURSO"

  # Caso de error - CA-08.10 (error del repositorio)
  Esquema del escenario: El repositorio falla al buscar o al actualizar
    Dado que existe un ítem con estado "PENDIENTE"
    Y que el repositorio falla <operacion>
    Cuando cambio el estado del ítem con el ID 1 a "EN_CURSO"
    Entonces el cambio de estado falla por un error del repositorio
    Y el ítem guardado con el ID 1 sigue en el estado "PENDIENTE"

    Ejemplos:
      | operacion     |
      | al buscar     |
      | al actualizar |
