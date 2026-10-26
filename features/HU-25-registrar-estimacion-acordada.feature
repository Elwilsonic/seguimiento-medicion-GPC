# language: es
# HU-25 · Registrar la estimación acordada (Issue #43)
# Especificación: docs/especificaciones/HU-25-registrar-estimacion-acordada.md
# Todos los pasos llaman al Manager de internal/poker (y al de internal/backlog
# para preparar y consultar el ítem), no al handler HTTP.

Característica: HU-25 Registrar la estimación acordada
  Como Agile Enabler,
  quiero registrar la estimación acordada por el equipo,
  para guardarla en el ítem del backlog.

  Antecedentes:
    Dado un proyecto de prueba
    Y un ítem "Historia de ejemplo" en estado "PENDIENTE" con 0 story points
    Y una sesión abierta para el ítem "Historia de ejemplo"
    Y que "Ana" votó 5 puntos
    Y que "Bruno" votó 8 puntos
    Y que los votos de la sesión fueron revelados

  Escenario: Registrar la estimación acordada guarda los puntos en el ítem y cierra la sesión (CA-25.1, CA-25.3)
    Cuando se registra la estimación acordada de 5 puntos
    Entonces el registro es exitoso
    Y la sesión devuelta está en estado "CERRADA"
    Y el ítem "Historia de ejemplo" tiene 5 story points
    Y la sesión guardada está en estado "CERRADA"
    Y los votos guardados son:
      | integrante | puntos |
      | Ana        | 5      |
      | Bruno      | 8      |

  Escenario: Registrar un valor que nadie votó (CA-25.2)
    Cuando se registra la estimación acordada de 13 puntos
    Entonces el registro es exitoso
    Y el ítem "Historia de ejemplo" tiene 13 story points
    Y la sesión guardada está en estado "CERRADA"

  Escenario: Registrar la estimación reemplaza los Story Points anteriores del ítem (CA-25.1)
    Dado que el ítem "Historia de ejemplo" tiene 3 story points asignados
    Cuando se registra la estimación acordada de 8 puntos
    Entonces el registro es exitoso
    Y el ítem "Historia de ejemplo" tiene 8 story points

  Escenario: Registrar el valor más alto de la escala (CA-25.2)
    Cuando se registra la estimación acordada de 21 puntos
    Entonces el registro es exitoso
    Y el ítem "Historia de ejemplo" tiene 21 story points

  Escenario: Rechazar una carta fuera de la escala (CA-25.2)
    Cuando se registra la estimación acordada de 4 puntos
    Entonces el registro falla porque la carta no pertenece a la escala
    Y la sesión guardada está en estado "REVELADA"
    Y el ítem "Historia de ejemplo" tiene 0 story points

  Escenario: Rechazar registrar la estimación de una sesión sin revelar (CA-25.4)
    Dado un ítem "Otra historia" en estado "PENDIENTE" con 0 story points
    Y una sesión abierta para el ítem "Otra historia"
    Y que "Ana" votó 3 puntos
    Cuando se registra la estimación acordada de 5 puntos
    Entonces el registro falla porque la sesión no está revelada
    Y la sesión guardada está en estado "ABIERTA"
    Y el ítem "Otra historia" tiene 0 story points

  Escenario: Rechazar registrar dos veces la misma sesión (CA-25.4)
    Dado que se registró la estimación acordada de 5 puntos
    Cuando se registra la estimación acordada de 8 puntos
    Entonces el registro falla porque la sesión no está revelada
    Y el ítem "Historia de ejemplo" tiene 5 story points
    Y la sesión guardada está en estado "CERRADA"

  Escenario: Rechazar registrar la estimación de una sesión inexistente (CA-25.4)
    Cuando se registra la estimación acordada de 5 puntos en la sesión 999
    Entonces el registro falla porque la sesión no existe
    Y el ítem "Historia de ejemplo" tiene 0 story points

  Escenario: Rechazar registrar la estimación de un ítem terminado (CA-25.5)
    Dado que el ítem "Historia de ejemplo" pasó a estado "TERMINADO"
    Cuando se registra la estimación acordada de 5 puntos
    Entonces el registro falla porque el ítem está terminado
    Y la sesión guardada está en estado "REVELADA"
    Y el ítem "Historia de ejemplo" tiene 0 story points

  Escenario: Dos registros simultáneos: solo uno se aplica (CA-25.8)
    Cuando se registran al mismo tiempo las estimaciones de 5 y 8 puntos
    Entonces solo uno de los registros se aplica
    Y el otro registro falla porque la sesión no está revelada
    Y la sesión guardada está en estado "CERRADA"
    Y el ítem "Historia de ejemplo" tiene los story points del registro aplicado
