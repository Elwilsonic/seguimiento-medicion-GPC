# language: es
# HU-52 · Identificadores UUID del resto de las entidades (Issue #197), parte de Planning Poker
# Especificación (OpenSpec): openspec/specs/planning-poker/spec.md
# Los escenarios son de la API, porque lo observable de las sesiones son las
# rutas, los 400 y 404 y las salidas. Las 1000 sesiones (CA-52.2), PostgreSQL
# (CA-52.5) y la concurrencia (CA-52.7) se cubren en la base, con tests de Go;
# la interfaz (CA-52.6), con los tests de sprints_poker_uuid_web_test.go.

Característica: HU-52 Identificadores UUID de las sesiones de Planning Poker
  Como integrante del equipo,
  quiero que las sesiones de Planning Poker se identifiquen con un UUID en lugar de un número correlativo,
  para que no se pueda adivinar ni recorrer ninguna sesión, ni deducir cuántas hay.

  Antecedentes:
    Dado que existe el proyecto "Inventario"
    Y que el proyecto "Inventario" tiene el ítem "Login" sin sprint

  # Caso normal
  Escenario: Abrir una sesión le asigna un UUID versión 4 (CA-52.1)
    Cuando pido POST "/api/items/{Login}/poker-sessions"
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto
    Y el "item_id" de la respuesta es el UUID de "Login"

  # Caso límite
  Escenario: Repetir la ronda conserva el UUID de la sesión (CA-52.1)
    Dado que el ítem "Login" tiene una sesión revelada con los votos 3 y 8
    Cuando pido POST "/api/poker-sessions/{la sesión de Login}/rounds"
    Entonces la respuesta tiene el código 200
    Y el "ronda" de la respuesta es 2
    Y el "id" de la respuesta es el UUID de "la sesión de Login"

  # Caso normal
  Escenario: Votar en una sesión por su UUID (CA-52.3)
    Dado que el ítem "Login" tiene una sesión abierta
    Cuando pido POST "/api/poker-sessions/{la sesión de Login}/votes" con el cuerpo:
      """
      {"integrante":"Ana","puntos":5}
      """
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es el UUID de "la sesión de Login"

  # Caso alternativo
  Escenario: El UUID de la sesión se acepta en mayúsculas (CA-52.3)
    Dado que el ítem "Login" tiene una sesión abierta con los votos 5 y 5
    Cuando pido POST "/api/poker-sessions/{la sesión de Login en mayúsculas}/reveal"
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto
    Y el "id" de la respuesta es el UUID de "la sesión de Login"

  # Caso alternativo
  Escenario: Revelar la sesión no expone ningún número (CA-52.3)
    Dado que el ítem "Login" tiene una sesión abierta con los votos 3 y 5
    Cuando pido POST "/api/poker-sessions/{la sesión de Login}/reveal"
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es el UUID de "la sesión de Login"
    Y el "item_id" de la respuesta es el UUID de "Login"

  # Caso de error
  Escenario: Un ID numérico de sesión responde 400 (CA-52.4)
    Cuando pido POST "/api/poker-sessions/1/reveal"
    Entonces la respuesta tiene el código 400
    Y el error es "el id de la sesión debe ser un UUID"

  # Caso límite
  Escenario: El ID de la sesión se valida antes que el cuerpo (CA-52.4)
    Cuando pido POST "/api/poker-sessions/abc/votes" con el cuerpo:
      """
      esto no es JSON
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id de la sesión debe ser un UUID"

  # Caso de error
  Escenario: Un UUID de sesión que no existe responde 404 (CA-52.4)
    Cuando pido POST "/api/poker-sessions/{un UUID que nunca se asignó}/reveal"
    Entonces la respuesta tiene el código 404
    Y el error es "no existe la sesión de Planning Poker"

  # Caso límite
  Escenario: Al votar se conserva el orden de errores (CA-52.4)
    Cuando pido POST "/api/poker-sessions/{un UUID que nunca se asignó}/votes" con el cuerpo:
      """
      {"integrante":"","puntos":5}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el votante no puede estar vacío"
