# language: es
# HU-24 · Detectar diferencias y repetir ronda (Issue #42)
# Especificación: docs/especificaciones/HU-24-repetir-ronda.md
# Todos los pasos llaman al Manager y a CompareVotes de internal/poker, no al
# handler HTTP.

Característica: HU-24 Detectar diferencias y repetir ronda
  Como Agile Enabler,
  quiero detectar si hay diferencias entre las estimaciones y abrir otra ronda,
  para llegar a un acuerdo.

  Antecedentes:
    Dado un proyecto de prueba
    Y un ítem "Historia de ejemplo" en estado "PENDIENTE" con 0 story points
    Y una sesión abierta para el ítem "Historia de ejemplo"

  Escenario: Revelar votos distintos informa la diferencia con el menor y el mayor (CA-24.1)
    Dado que "Ana" votó 3 puntos
    Y que "Bruno" votó 8 puntos
    Y que "Carla" votó 8 puntos
    Cuando se revelan los votos de la sesión
    Entonces se informa que hay diferencias
    Y el voto menor es 3, de "Ana"
    Y el voto mayor es 8, de "Bruno, Carla"

  Escenario: Repetir la ronda de una sesión revelada (CA-24.2, CA-24.3, CA-24.6)
    Dado que "Ana" votó 3 puntos
    Y que "Bruno" votó 8 puntos
    Y que los votos de la sesión fueron revelados
    Cuando se repite la ronda de la sesión
    Entonces la repetición es exitosa
    Y la sesión guardada está en estado "ABIERTA" en la ronda 2
    Y la sesión guardada no tiene votos
    Y sigue siendo la única sesión del ítem "Historia de ejemplo"
    Y el ítem "Historia de ejemplo" sigue en estado "PENDIENTE" con 0 story points

  Escenario: Revelar votos iguales informa que no hay diferencias (CA-24.1)
    Dado que "Ana" votó 5 puntos
    Y que "Bruno" votó 5 puntos
    Cuando se revelan los votos de la sesión
    Entonces se informa que no hay diferencias
    Y el voto menor es 5, de "Ana, Bruno"
    Y el voto mayor es 5, de "Ana, Bruno"

  Escenario: Repetir la ronda aunque no haya diferencias (CA-24.4)
    Dado que "Ana" votó 5 puntos
    Y que "Bruno" votó 5 puntos
    Y que los votos de la sesión fueron revelados
    Cuando se repite la ronda de la sesión
    Entonces la repetición es exitosa
    Y la sesión guardada está en estado "ABIERTA" en la ronda 2

  Escenario: Votar, revelar y registrar la estimación en la ronda nueva (CA-24.5)
    Dado que "Ana" votó 3 puntos
    Y que "Bruno" votó 8 puntos
    Y que los votos de la sesión fueron revelados
    Y que se repitió la ronda de la sesión
    Y que "Ana" votó 5 puntos
    Y que "Bruno" votó 5 puntos
    Cuando se revelan los votos de la sesión
    Entonces se informa que no hay diferencias
    Y se muestran los votos:
      | integrante | puntos |
      | Ana        | 5      |
      | Bruno      | 5      |
    Cuando se registra la estimación acordada de 5 puntos
    Entonces la sesión guardada está en estado "CERRADA" en la ronda 2
    Y el ítem "Historia de ejemplo" tiene 5 story points

  Escenario: Revelar un único voto no informa diferencias (CA-24.1)
    Dado que "Ana" votó 13 puntos
    Cuando se revelan los votos de la sesión
    Entonces se informa que no hay diferencias
    Y el voto menor es 13, de "Ana"
    Y el voto mayor es 13, de "Ana"

  Escenario: Repetir la ronda dos veces seguidas llega a la ronda 3 (CA-24.3)
    Dado que "Ana" votó 3 puntos
    Y que los votos de la sesión fueron revelados
    Y que se repitió la ronda de la sesión
    Y que "Ana" votó 8 puntos
    Y que los votos de la sesión fueron revelados
    Cuando se repite la ronda de la sesión
    Entonces la repetición es exitosa
    Y la sesión guardada está en estado "ABIERTA" en la ronda 3
    Y la sesión guardada no tiene votos

  Escenario: Repetir la ronda varias veces al mismo tiempo la aumenta una sola vez (CA-24.10)
    Dado que "Ana" votó 3 puntos
    Y que los votos de la sesión fueron revelados
    Cuando se repite la ronda de la sesión 5 veces al mismo tiempo
    Entonces solo una repetición es exitosa
    Y las demás repeticiones fallan porque la sesión no está revelada
    Y la sesión guardada está en estado "ABIERTA" en la ronda 2

  Escenario: Rechazar repetir la ronda de una sesión abierta (CA-24.7)
    Dado que "Ana" votó 3 puntos
    Cuando se repite la ronda de la sesión
    Entonces la repetición falla porque la sesión no está revelada
    Y la sesión guardada está en estado "ABIERTA" en la ronda 1
    Y los votos guardados son:
      | integrante | puntos |
      | Ana        | 3      |

  Escenario: Rechazar repetir la ronda de una sesión cerrada (CA-24.7)
    Dado que "Ana" votó 3 puntos
    Y que los votos de la sesión fueron revelados
    Y que se registró la estimación acordada de 3 puntos
    Cuando se repite la ronda de la sesión
    Entonces la repetición falla porque la sesión no está revelada
    Y la sesión guardada está en estado "CERRADA" en la ronda 1

  Esquema del escenario: Rechazar repetir la ronda de una sesión inexistente (CA-24.7)
    Dado que "Ana" votó 3 puntos
    Y que los votos de la sesión fueron revelados
    Cuando se repite la ronda de la sesión <id>
    Entonces la repetición falla porque la sesión no existe
    Y la sesión guardada está en estado "REVELADA" en la ronda 1

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  Escenario: Rechazar repetir la ronda si hay otra sesión abierta para el ítem (CA-24.8)
    Dado que "Ana" votó 3 puntos
    Y que los votos de la sesión fueron revelados
    Y que se abrió otra sesión para el ítem "Historia de ejemplo"
    Cuando se repite la ronda de la sesión
    Entonces la repetición falla porque ya hay otra sesión abierta para el ítem
    Y la sesión guardada está en estado "REVELADA" en la ronda 1
    Y los votos guardados son:
      | integrante | puntos |
      | Ana        | 3      |

  Escenario: Si falla el guardado, la sesión sigue revelada (CA-24.9)
    Dado que "Ana" votó 3 puntos
    Y que los votos de la sesión fueron revelados
    Y que el almacenamiento de sesiones falla al guardar
    Cuando se repite la ronda de la sesión
    Entonces la repetición falla por el error del almacenamiento
    Y la sesión guardada está en estado "REVELADA" en la ronda 1
    Y los votos guardados son:
      | integrante | puntos |
      | Ana        | 3      |
