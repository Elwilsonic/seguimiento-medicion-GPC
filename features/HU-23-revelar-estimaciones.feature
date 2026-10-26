# language: es
# HU-23 · Revelar las estimaciones (Issue #41)
# Especificación: docs/especificaciones/HU-23-revelar-estimaciones.md
# Todos los pasos llaman al Manager de internal/poker, no al handler HTTP.

Característica: HU-23 Revelar las estimaciones
  Como Agile Enabler,
  quiero revelar los votos de todos a la vez,
  para comparar las estimaciones.

  Antecedentes:
    Dado un proyecto de prueba
    Y un ítem "Historia de ejemplo" en estado "PENDIENTE" con 0 story points
    Y una sesión abierta para el ítem "Historia de ejemplo"

  Escenario: Revelar una sesión con votos muestra todos los votos (CA-23.1)
    Dado que "Ana" votó 5 puntos
    Y que "Bruno" votó 8 puntos
    Cuando se revelan los votos de la sesión
    Entonces la revelación es exitosa
    Y la sesión revelada está en estado "REVELADA"
    Y se muestran los votos:
      | integrante | puntos |
      | Ana        | 5      |
      | Bruno      | 8      |

  Escenario: Revelar una sesión con un único voto (CA-23.1)
    Dado que "Ana" votó 3 puntos
    Cuando se revelan los votos de la sesión
    Entonces la revelación es exitosa
    Y se muestran los votos:
      | integrante | puntos |
      | Ana        | 3      |

  Escenario: Revelar conserva la ronda, el ítem y el orden de los votos (CA-23.6)
    Dado que "Carla" votó 13 puntos
    Y que "Ana" votó 2 puntos
    Cuando se revelan los votos de la sesión
    Entonces la revelación es exitosa
    Y la sesión revelada sigue en la ronda 1 del ítem "Historia de ejemplo"
    Y se muestran los votos:
      | integrante | puntos |
      | Carla      | 13     |
      | Ana        | 2      |
    Y el ítem "Historia de ejemplo" sigue en estado "PENDIENTE" con 0 story points

  Escenario: Revelar muestra el voto reemplazado de un integrante (CA-23.1)
    Dado que "Ana" votó 5 puntos
    Y que "ana" votó 8 puntos
    Cuando se revelan los votos de la sesión
    Entonces la revelación es exitosa
    Y se muestran los votos:
      | integrante | puntos |
      | ana        | 8      |

  Escenario: Rechazar revelar una sesión sin votos (CA-23.2)
    Cuando se revelan los votos de la sesión
    Entonces la revelación falla porque la sesión no tiene votos
    Y la sesión guardada está en estado "ABIERTA"

  Escenario: No se puede votar después de revelar (CA-23.3)
    Dado que "Ana" votó 5 puntos
    Y que los votos de la sesión fueron revelados
    Cuando "Ana" vota 8 puntos
    Entonces el voto falla porque la sesión no está abierta
    Y los votos guardados son:
      | integrante | puntos |
      | Ana        | 5      |

  Escenario: Rechazar revelar una sesión ya revelada (CA-23.4)
    Dado que "Ana" votó 5 puntos
    Y que los votos de la sesión fueron revelados
    Cuando se revelan los votos de la sesión
    Entonces la revelación falla porque la sesión no está abierta
    Y la sesión guardada está en estado "REVELADA"

  Escenario: Rechazar revelar una sesión inexistente (CA-23.4)
    Cuando se revelan los votos de la sesión 999
    Entonces la revelación falla porque la sesión no existe
    Y la sesión guardada está en estado "ABIERTA"

  Escenario: Votar mientras se revela no pierde votos aceptados (CA-23.7)
    Dado que "Ana" votó 5 puntos
    Cuando 10 integrantes votan mientras se revelan los votos de la sesión
    Entonces la revelación es exitosa
    Y la sesión guardada está en estado "REVELADA"
    Y cada voto aceptado está entre los votos guardados
    Y cada voto rechazado falló porque la sesión no está abierta
