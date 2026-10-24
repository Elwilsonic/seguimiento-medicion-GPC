# language: es
# HU-20 · Asignar story points a un ítem del backlog (Issue #38)
# Especificación: docs/especificaciones/HU-20-asignar-story-points.md
# CA-20.8 (API e interfaz web) se cubre con los tests de internal/httpapi, no con escenarios godog.

Característica: HU-20 Asignar story points a un ítem del backlog
  Como Product Builder,
  quiero asignar story points a un ítem del backlog,
  para estimar su tamaño.

  # Caso normal (happy path)
  Escenario: Asignar story points a un ítem sin estimar (CA-20.1)
    Dado que existe un ítem sin estimar en estado "PENDIENTE"
    Cuando asigno 5 story points al ítem con el ID 1
    Entonces la asignación es exitosa
    Y el ítem devuelto tiene 5 story points
    Y el ítem guardado con el ID 1 tiene 5 story points
    Y el ítem guardado con el ID 1 conserva los demás campos

  # Caso alternativo: reestimar reemplaza el valor anterior
  Esquema del escenario: Reestimar un ítem que ya tiene story points (CA-20.1)
    Dado que existe un ítem en estado "PENDIENTE" con <antes> story points
    Cuando asigno <despues> story points al ítem con el ID 1
    Entonces la asignación es exitosa
    Y el ítem guardado con el ID 1 tiene <despues> story points
    Y el ítem guardado con el ID 1 conserva los demás campos

    Ejemplos:
      | caso         | antes | despues |
      | otro valor   | 3     | 8       |
      | mismo valor  | 5     | 5       |

  # Caso límite: cada valor de la escala, incluidos el mínimo (1) y el máximo (21)
  Esquema del escenario: Aceptar cada valor de la escala Fibonacci (CA-20.2)
    Dado que existe un ítem sin estimar en estado "PENDIENTE"
    Cuando asigno <valor> story points al ítem con el ID 1
    Entonces la asignación es exitosa
    Y el ítem guardado con el ID 1 tiene <valor> story points

    Ejemplos:
      | valor |
      | 1     |
      | 2     |
      | 3     |
      | 5     |
      | 8     |
      | 13    |
      | 21    |

  # Caso límite: la escala se define en un solo lugar y no se puede modificar desde afuera
  Escenario: La escala compartida no cambia si se modifica la lista devuelta (CA-20.2)
    Cuando consulto la escala de story points
    Y modifico el primer valor de la escala devuelta
    Entonces la escala de story points es "1, 2, 3, 5, 8, 13, 21"

  # Caso de error: valores fuera de la escala (0, negativos, intermedios y mayores que 21)
  Esquema del escenario: Rechazar valores fuera de la escala (CA-20.3)
    Dado que existe un ítem en estado "PENDIENTE" con 3 story points
    Cuando asigno <valor> story points al ítem con el ID 1
    Entonces la asignación falla porque el valor no está en la escala
    Y el ítem guardado con el ID 1 tiene 3 story points

    Ejemplos:
      | caso                | valor |
      | cero (sin estimar)  | 0     |
      | negativo            | -1    |
      | intermedio          | 4     |
      | mayor que el máximo | 22    |

  # Caso límite: la escala se valida antes que la existencia del ítem
  Escenario: Un valor fuera de la escala sobre un ítem inexistente informa la escala (CA-20.3)
    Dado que existe un ítem sin estimar en estado "PENDIENTE"
    Cuando asigno 4 story points al ítem con el ID 999
    Entonces la asignación falla porque el valor no está en la escala
    Y no se modificó ningún ítem guardado

  # Caso de error: ítem inexistente y sin ítems creados
  Esquema del escenario: Asignar story points a un ítem que no existe (CA-20.4)
    Dado que hay <items> ítems sin estimar creados
    Cuando asigno 5 story points al ítem con el ID <id>
    Entonces la asignación falla porque el ítem no existe
    Y no se modificó ningún ítem guardado

    Ejemplos:
      | items | id  |
      | 1     | 999 |
      | 1     | 0   |
      | 1     | -1  |
      | 0     | 1   |

  # Caso de error: un ítem terminado no se reestima directamente
  Escenario: Rechazar la asignación a un ítem terminado (CA-20.5)
    Dado que existe un ítem en estado "TERMINADO" con 3 story points
    Cuando asigno 8 story points al ítem con el ID 1
    Entonces la asignación falla porque el ítem está terminado
    Y el ítem guardado con el ID 1 tiene 3 story points

  # Caso alternativo: un ítem reabierto con HU-08 vuelve a ser estimable (límite conocido)
  Escenario: Reestimar un ítem terminado después de reabrirlo (CA-20.5)
    Dado que existe un ítem en estado "TERMINADO" con 3 story points
    Cuando cambio el estado del ítem con el ID 1 a "EN_CURSO"
    Y asigno 8 story points al ítem con el ID 1
    Entonces la asignación es exitosa
    Y el ítem guardado con el ID 1 tiene 8 story points

  # Caso límite: los ítems no terminados se estiman aunque estén en un sprint
  Esquema del escenario: Estimar un ítem no terminado asignado a un sprint (CA-20.5)
    Dado que existe un ítem sin estimar en estado "<estado>"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 2
    Cuando asigno 13 story points al ítem con el ID 1
    Entonces la asignación es exitosa
    Y el ítem guardado con el ID 1 tiene 13 story points
    Y el ítem guardado con el ID 1 sigue asignado al sprint con el ID 2

    Ejemplos:
      | estado    |
      | PENDIENTE |
      | EN_CURSO  |

  # Caso límite: el Update de HU-07 conserva los story points
  Escenario: Modificar los datos de un ítem estimado conserva sus story points (CA-20.6)
    Dado que existe un ítem sin estimar en estado "PENDIENTE"
    Y que asigné 8 story points al ítem con el ID 1
    Cuando modifico los datos del ítem con el ID 1 con el título "Login con OAuth"
    Entonces la modificación es exitosa
    Y el ítem guardado con el ID 1 tiene 8 story points

  # Caso límite: copia por valor
  Escenario: El ítem devuelto es una copia (CA-20.7)
    Dado que existe un ítem sin estimar en estado "PENDIENTE"
    Cuando asigno 5 story points al ítem con el ID 1
    Y modifico los story points y los criterios del ítem devuelto
    Entonces el ítem guardado con el ID 1 tiene 5 story points
    Y el ítem guardado con el ID 1 conserva los demás campos

  # Caso límite: asignaciones simultáneas mientras otro pedido termina el ítem.
  # Las aserciones no dependen del orden de las goroutines: después de TERMINADO
  # no se escribe nada, así que queda el valor de una asignación exitosa, o 0 si
  # el cambio de estado llegó antes que todas.
  Escenario: Asignaciones simultáneas con un cambio de estado a TERMINADO (CA-20.7)
    Dado que existe un ítem sin estimar en estado "EN_CURSO"
    Cuando se asignan en simultáneo 50 veces story points de la escala al ítem con el ID 1 mientras se cambia su estado a "TERMINADO"
    Entonces cada asignación es exitosa o falla porque el ítem está terminado
    Y el ítem guardado con el ID 1 tiene los story points de una asignación exitosa, o 0 si ninguna lo fue
    Y el ítem guardado con el ID 1 está en el estado "TERMINADO"

  # Caso de error: el repositorio falla
  Esquema del escenario: El repositorio falla al buscar o al actualizar (CA-20.7)
    Dado que existe un ítem en estado "PENDIENTE" con 3 story points
    Y que el repositorio falla <operacion>
    Cuando asigno 8 story points al ítem con el ID 1
    Entonces la asignación falla por un error del repositorio
    Y el ítem guardado con el ID 1 tiene 3 story points

    Ejemplos:
      | operacion     |
      | al buscar     |
      | al actualizar |
