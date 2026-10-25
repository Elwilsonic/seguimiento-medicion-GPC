# language: es
# HU-22 · Votar de forma oculta (Issue #40)
# Especificación: docs/especificaciones/HU-22-votar-oculto.md

Característica: HU-22 Votar de forma oculta
  Como Product Builder,
  quiero votar mi estimación sin que los demás la vean,
  para no influirnos entre nosotros.

  Antecedentes:
    Dado un proyecto de prueba
    Y un ítem "Historia de ejemplo" en estado "PENDIENTE" con 0 story points
    Y una sesión abierta para el ítem "Historia de ejemplo"

  Escenario: Registrar un voto válido sin mostrarlo (CA-22.1, CA-22.2)
    Cuando "Ana" vota 5 puntos
    Entonces el voto es exitoso
    Y la sesión sigue sin mostrar votos

  Escenario: Reemplazar el voto del mismo integrante (CA-22.3)
    Dado que "Ana" votó 5 puntos
    Cuando " ANA " vota 8 puntos
    Entonces el voto es exitoso
    Y la sesión sigue teniendo un solo voto guardado
    Y el nombre guardado es "ANA"

  Escenario: Rechazar una carta fuera de la escala (CA-22.4)
    Cuando "Ana" vota 4 puntos
    Entonces el voto falla porque la carta no es válida
    Y la sesión no tiene votos guardados

  Escenario: Rechazar un votante vacío (CA-22.4)
    Cuando "   " vota 5 puntos
    Entonces el voto falla porque el votante está vacío
    Y la sesión no tiene votos guardados

  Escenario: Rechazar el voto de una sesión inexistente (CA-22.5)
    Cuando "Ana" vota 5 puntos en la sesión 999
    Entonces el voto falla porque la sesión no existe

  Escenario: Rechazar el voto de una sesión no abierta (CA-22.5)
    Dado que la sesión tiene un estado distinto de "ABIERTA"
    Cuando "Ana" vota 5 puntos
    Entonces el voto falla porque la sesión no está abierta

  Escenario: Registrar votos simultáneos de integrantes distintos (CA-22.6)
    Cuando "Ana" y "Bruno" votan simultáneamente 5 puntos
    Entonces ambos votos quedan guardados

  Escenario: El voto reemplazado mantiene los votos ocultos (CA-22.7)
    Dado que "Ana" votó 5 puntos
    Cuando "Ana" vota 13 puntos
    Entonces el voto es exitoso
    Y la sesión sigue sin mostrar votos
