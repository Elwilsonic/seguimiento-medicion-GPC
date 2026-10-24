# language: es
# HU-21 · Abrir una sesión de Planning Poker (Issue #39)
# Especificación: docs/especificaciones/HU-21-abrir-sesion-planning-poker.md
# CA-21.6 (concurrencia), CA-21.7 (API e interfaz web) y CA-21.8 (PostgreSQL)
# se cubren con tests de Go, no con escenarios godog.

Característica: HU-21 Abrir una sesión de Planning Poker
  Como Agile Enabler,
  quiero abrir una sesión de Planning Poker para un ítem,
  para que el equipo lo estime en conjunto.

  Antecedentes:
    Dado un proyecto de prueba

  # Caso normal (happy path)
  Escenario: Abrir una sesión para un ítem pendiente (CA-21.1)
    Dado un ítem "Historia de ejemplo" en estado "PENDIENTE" con 0 story points
    Cuando abro una sesión de Planning Poker para el ítem "Historia de ejemplo"
    Entonces la apertura es exitosa
    Y la sesión está asociada al ítem "Historia de ejemplo"
    Y la sesión está abierta, en la ronda 1 y sin votos

  # Caso alternativo: dos ítems distintos pueden tener una sesión abierta cada uno
  Escenario: Abrir sesiones para dos ítems distintos (CA-21.2)
    Dado un ítem "Primera historia" en estado "PENDIENTE" con 0 story points
    Y un ítem "Segunda historia" en estado "PENDIENTE" con 0 story points
    Cuando abro una sesión de Planning Poker para el ítem "Primera historia"
    Y abro una sesión de Planning Poker para el ítem "Segunda historia"
    Entonces las dos aperturas son exitosas
    Y las dos sesiones están abiertas y tienen IDs distintos

  # Caso límite: un ítem ya estimado y en curso también se puede estimar de nuevo
  Escenario: Abrir una sesión para un ítem en curso con story points (CA-21.1)
    Dado un ítem "Historia estimada" en estado "EN_CURSO" con 5 story points
    Cuando abro una sesión de Planning Poker para el ítem "Historia estimada"
    Entonces la apertura es exitosa
    Y la sesión está asociada al ítem "Historia estimada"
    Y la sesión está abierta, en la ronda 1 y sin votos

  # Caso de error: no puede haber dos sesiones abiertas para el mismo ítem
  Escenario: No abrir dos sesiones para el mismo ítem (CA-21.2)
    Dado un ítem "Historia de ejemplo" en estado "PENDIENTE" con 0 story points
    Y que ya hay una sesión abierta para el ítem "Historia de ejemplo"
    Cuando abro una sesión de Planning Poker para el ítem "Historia de ejemplo"
    Entonces la apertura falla porque ya hay una sesión abierta
    Y queda una sola sesión abierta para el ítem "Historia de ejemplo"

  # Caso de error: el ítem no existe
  Escenario: No abrir una sesión para un ítem inexistente (CA-21.4)
    Cuando abro una sesión de Planning Poker para el ítem con el ID 999
    Entonces la apertura falla porque el ítem no existe

  # Caso de error: el ítem está terminado
  Escenario: No abrir una sesión para un ítem terminado (CA-21.4)
    Dado un ítem "Historia terminada" en estado "TERMINADO" con 0 story points
    Cuando abro una sesión de Planning Poker para el ítem "Historia terminada"
    Entonces la apertura falla porque el ítem está terminado

  # Las cartas son la escala compartida de HU-20
  Escenario: Las cartas de la sesión son la escala de HU-20 (CA-21.5)
    Cuando consulto las cartas de Planning Poker
    Entonces las cartas son la escala de HU-20
    Cuando modifico la lista de cartas devuelta
    Entonces las cartas siguen siendo la escala de HU-20
