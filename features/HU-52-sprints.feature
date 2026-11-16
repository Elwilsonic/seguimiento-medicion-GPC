# language: es
# HU-52 · Identificadores UUID del resto de las entidades (Issue #197), parte de sprints
# Especificación (OpenSpec): openspec/specs/sprints/spec.md
# Los escenarios son de la API, porque lo observable de los sprints son las
# rutas, los 400 y 404 y las salidas. Los 1000 sprints (CA-52.2), PostgreSQL
# (CA-52.5) y la concurrencia (CA-52.7) se cubren en la base, con tests de Go;
# la interfaz (CA-52.6), con los tests de sprints_poker_uuid_web_test.go.
# Este archivo prueba solo las rutas de sprints.go, metrics.go y
# sprint_reports.go (D3 de design.md). El sprint_id de los ítems lo cubre
# HU-52-items.feature (Cardozo Leandro), y la ruta de registrar un esfuerzo,
# con su orden de errores, HU-52-esfuerzos.feature (García Santiago).

Característica: HU-52 Identificadores UUID de los sprints
  Como integrante del equipo,
  quiero que los sprints se identifiquen con un UUID en lugar de un número correlativo,
  para que no se pueda adivinar ni recorrer ningún sprint, ni deducir cuántos hay.

  Antecedentes:
    Dado que existe el proyecto "Inventario"

  # Caso normal
  Escenario: Crear un sprint le asigna un UUID versión 4 (CA-52.1)
    Cuando pido POST "/api/proyectos/{Inventario}/sprints" con el cuerpo:
      """
      {"nombre":"Sprint 1","fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Gestionar el backlog"}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto

  # Caso límite
  Escenario: El primer sprint no es el 1 (CA-52.1)
    Cuando pido POST "/api/proyectos/{Inventario}/sprints" con el cuerpo:
      """
      {"nombre":"Primero","fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Arrancar"}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta no es un número ni el UUID nulo

  # Caso normal
  Escenario: Cerrar un sprint por su UUID (CA-52.3)
    Dado que el proyecto "Inventario" tiene el sprint abierto "Sprint 1"
    Cuando pido POST "/api/sprints/{Sprint 1}/cerrar"
    Entonces la respuesta tiene el código 200
    Y el "estado" de la respuesta es "CERRADO"
    Y el "id" de la respuesta es el UUID de "Sprint 1"

  # Caso alternativo
  Escenario: El UUID del sprint se acepta en mayúsculas (CA-52.3)
    Dado que el proyecto "Inventario" tiene el sprint abierto "Sprint 1"
    Cuando pido GET "/api/sprints/{Sprint 1 en mayúsculas}/metricas"
    Entonces la respuesta tiene el código 200
    Y el "sprint_id" de la respuesta es el UUID de "Sprint 1"

  # Caso alternativo
  Escenario: Asignar un ítem a un sprint por sus UUID (CA-52.3)
    Dado que el proyecto "Inventario" tiene el sprint abierto "Sprint 1"
    Y que el proyecto "Inventario" tiene el ítem "Login" sin sprint
    Cuando pido POST "/api/sprints/{Sprint 1}/items/{Login}"
    Entonces la respuesta tiene el código 200
    Y el ítem "Login" quedó asignado al sprint "Sprint 1"

  # Caso de error
  Escenario: Un ID numérico de sprint responde 400 (CA-52.4)
    Cuando pido GET "/api/sprints/1/metricas"
    Entonces la respuesta tiene el código 400
    Y el error es "el id del sprint debe ser un UUID"

  # Caso límite
  Escenario: Primero se valida el sprint y después el ítem (CA-52.4)
    Cuando pido POST "/api/sprints/abc/items/xyz"
    Entonces la respuesta tiene el código 400
    Y el error es "el id del sprint debe ser un UUID"

  # Caso de error
  Escenario: Un UUID de sprint que no existe responde 404 (CA-52.4)
    Cuando pido GET "/api/sprints/{un UUID que nunca se asignó}/metricas"
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un sprint con ese id"
