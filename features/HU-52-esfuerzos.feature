# language: es
# HU-52 · Identificadores UUID del resto de las entidades, parte de esfuerzos (Issue #197), vinculada a HU-26 (#44), HU-27 (#45) y HU-28 (#46)
# Especificación (OpenSpec): openspec/changes/hu-52-identificadores-uuid-entidades/specs/esfuerzos/spec.md
# Los escenarios son a nivel de la API y prueban solo las rutas de efforts.go
# (D3 de design.md): POST /api/sprints/{id}/items/{itemId}/esfuerzos,
# GET /api/items/{id}/esfuerzos, GET /api/sprints/{id}/esfuerzos y
# PATCH /api/esfuerzos/{id}/horas. El proyecto, el ítem, el sprint y los
# esfuerzos de los pasos "Dado" se crean con el dominio, sin pasar por las
# rutas de ítems y sprints, que son de otros responsables.
# En las rutas, {ítem}, {sprint} y {esfuerzo} son los UUID de esas entidades;
# {inexistente} es un UUID versión 4 que nunca se asigna; {esfuerzo en
# mayúsculas}, {esfuerzo entre llaves}, {esfuerzo con urn:uuid: adelante} y
# {esfuerzo sin guiones} son el UUID del esfuerzo escrito de esas formas.
# Los 1000 esfuerzos (CA-52.2), las creaciones simultáneas (CA-52.7) y
# PostgreSQL (CA-52.5) son de dominio y los cubre la base
# (features/HU-52-base.feature); la interfaz (CA-52.6), con tests de Go.

Característica: HU-52 Identificadores UUID de los registros de esfuerzo
  Como integrante del equipo,
  quiero que los registros de esfuerzo se identifiquen con un UUID en lugar de un número correlativo,
  para que no se pueda adivinar ni recorrer ningún esfuerzo, ni deducir cuántos hay.

  Antecedentes:
    Dado que existe un ítem asignado a un sprint abierto

  # Caso normal
  Escenario: Registrar un esfuerzo (CA-52.1, CA-52.3)
    Cuando pido POST "/api/sprints/{sprint}/items/{ítem}/esfuerzos" con el cuerpo:
      """
      {"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Implementar el registro","horas":2.5}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto
    Y el "item_id" de la respuesta es el UUID del ítem
    Y el "sprint_id" de la respuesta es el UUID del sprint

  # Caso límite
  Escenario: El primer esfuerzo no es el 1 (CA-52.1)
    Dado que no hay esfuerzos registrados
    Cuando pido POST "/api/sprints/{sprint}/items/{ítem}/esfuerzos" con el cuerpo:
      """
      {"integrante":"García Santiago","fecha":"2026-10-24","actividad":"Primer registro","horas":1}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta no es un número ni el UUID nulo

  # Caso normal
  Escenario: Corregir las horas por UUID (CA-52.3)
    Dado que el ítem tiene 1 esfuerzos registrados
    Cuando pido PATCH "/api/esfuerzos/{esfuerzo}/horas" con el cuerpo:
      """
      {"horas":3}
      """
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es el UUID del esfuerzo
    Y el "horas" de la respuesta es el número 3

  # Caso alternativo
  Escenario: UUID del esfuerzo en mayúsculas (CA-52.3)
    Dado que el ítem tiene 1 esfuerzos registrados
    Cuando pido PATCH "/api/esfuerzos/{esfuerzo en mayúsculas}/horas" con el cuerpo:
      """
      {"horas":3}
      """
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es el UUID del esfuerzo

  # Caso alternativo
  Escenario: La lista de esfuerzos de un sprint (CA-52.3)
    Dado que el ítem tiene 2 esfuerzos registrados
    Cuando pido GET "/api/sprints/{sprint}/esfuerzos"
    Entonces la respuesta tiene el código 200
    Y la lista "esfuerzos" tiene 2 elementos
    Y cada elemento de "esfuerzos" tiene en "id" un UUID versión 4
    Y cada elemento de "esfuerzos" tiene en "item_id" el UUID del ítem
    Y cada elemento de "esfuerzos" tiene en "sprint_id" el UUID del sprint

  # Caso alternativo
  Escenario: La lista de esfuerzos de una historia por su UUID (CA-52.3)
    Dado que el ítem tiene 2 esfuerzos registrados
    Cuando pido GET "/api/items/{ítem}/esfuerzos"
    Entonces la respuesta tiene el código 200
    Y la lista "esfuerzos" tiene 2 elementos
    Y cada elemento de "esfuerzos" tiene en "id" un UUID versión 4
    Y cada elemento de "esfuerzos" tiene en "item_id" el UUID del ítem

  # Caso de error
  Escenario: Un ID numérico de esfuerzo (CA-52.4)
    Dado que el ítem tiene 1 esfuerzos registrados
    Cuando pido PATCH "/api/esfuerzos/1/horas" con el cuerpo:
      """
      {"horas":3}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del esfuerzo debe ser un UUID"

  # Caso límite
  Esquema del escenario: Otras formas del UUID del esfuerzo se rechazan (CA-52.4)
    Dado que el ítem tiene 1 esfuerzos registrados
    Cuando pido PATCH "/api/esfuerzos/{esfuerzo <forma>}/horas" con el cuerpo:
      """
      {"horas":3}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del esfuerzo debe ser un UUID"

    Ejemplos:
      | forma                  |
      | entre llaves           |
      | con urn:uuid: adelante |
      | sin guiones            |

  # Caso límite
  Escenario: El ID del esfuerzo se valida antes que el cuerpo (CA-52.4)
    Cuando pido PATCH "/api/esfuerzos/abc/horas" con el cuerpo:
      """
      esto no es JSON
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del esfuerzo debe ser un UUID"

  # Caso de error
  Escenario: UUID de esfuerzo que no existe (CA-52.4)
    Cuando pido PATCH "/api/esfuerzos/{inexistente}/horas" con el cuerpo:
      """
      {"horas":3}
      """
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un esfuerzo con ese id"

  # Caso límite
  Escenario: Se conserva el orden de errores (CA-52.4)
    Cuando pido PATCH "/api/esfuerzos/{inexistente}/horas" con el cuerpo:
      """
      {"horas":0}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "las horas trabajadas deben ser un número positivo"
