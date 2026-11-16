# language: es
# HU-52 · Identificadores UUID del resto de las entidades, parte de integrantes (Issue #197), vinculada a HU-15 a HU-17
# Especificación (OpenSpec): openspec/specs/integrantes/spec.md
# Los escenarios son a nivel de la API: el registro (HU-15), la lista (HU-17)
# y PUT /api/integrantes/{id} (HU-16), la única ruta que recibe un integrante
# (tabla "Rutas por responsable" de D3 en design.md). Los 1000 integrantes
# (CA-52.2), la concurrencia (CA-52.7), PostgreSQL (CA-52.5) y la interfaz
# (CA-52.6) se cubren con tests de Go.

Característica: HU-52 Identificadores UUID de los integrantes
  Como integrante del equipo,
  quiero que los integrantes de un proyecto se identifiquen con un UUID en lugar de un número correlativo,
  para que no se pueda adivinar ni recorrer ningún integrante, ni deducir cuántos hay.

  Antecedentes:
    Dado que existe el proyecto "Inventario"

  # Caso normal
  Escenario: Registrar un integrante le asigna un UUID versión 4 (CA-52.1)
    Cuando pido POST "/api/proyectos/{Inventario}/integrantes" con el cuerpo:
      """
      {"nombre":"Ana Pérez","rol":"PRODUCT_BUILDER"}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto

  # Caso límite
  Escenario: El primer integrante no es el 1 (CA-52.1)
    Dado que el proyecto "Inventario" no tiene integrantes
    Cuando pido POST "/api/proyectos/{Inventario}/integrantes" con el cuerpo:
      """
      {"nombre":"Primero","rol":"AGILE_ENABLER"}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta no es un número ni el UUID nulo

  # Caso normal
  Escenario: Modificar un integrante por su UUID (CA-52.3)
    Dado que el proyecto "Inventario" tiene el integrante "Ana Pérez"
    Cuando pido PUT "/api/integrantes/{Ana Pérez}" con el cuerpo:
      """
      {"nombre":"Ana María","rol":"AGILE_ENABLER"}
      """
    Entonces la respuesta tiene el código 200
    Y el "nombre" de la respuesta es "Ana María"
    Y el "rol" de la respuesta es "AGILE_ENABLER"
    Y el "id" de la respuesta es el UUID de "Ana Pérez"

  # Caso alternativo
  Escenario: El UUID del integrante se acepta en mayúsculas (CA-52.3)
    Dado que el proyecto "Inventario" tiene el integrante "Ana Pérez"
    Cuando pido PUT a la ruta del integrante "Ana Pérez" escrito en mayúsculas con el cuerpo:
      """
      {"nombre":"Ana Pérez","rol":"PRODUCT_ARCHITECT"}
      """
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto
    Y el "id" de la respuesta es el UUID de "Ana Pérez"

  # Caso alternativo
  Escenario: La lista identifica cada integrante con su UUID (CA-52.3)
    Dado que el proyecto "Inventario" tiene el integrante "Ana Pérez"
    Y que el proyecto "Inventario" tiene el integrante "Bruno Díaz"
    Cuando pido GET "/api/proyectos/{Inventario}/integrantes"
    Entonces la respuesta tiene el código 200
    Y cada integrante de la lista tiene su UUID en "id", distinto del de los demás

  # Caso alternativo
  Escenario: Consultar un integrante por su UUID sigue sin estar permitido (CA-52.3)
    Dado que el proyecto "Inventario" tiene el integrante "Ana Pérez"
    Cuando pido GET "/api/integrantes/{Ana Pérez}"
    Entonces la respuesta tiene el código 405

  # Caso de error
  Escenario: Un ID numérico de integrante se rechaza (CA-52.4)
    Dado que el proyecto "Inventario" tiene el integrante "Ana Pérez"
    Cuando pido PUT "/api/integrantes/1" con el cuerpo:
      """
      {"nombre":"Ana María","rol":"AGILE_ENABLER"}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del integrante debe ser un UUID"

  # Caso límite
  Esquema del escenario: Otras formas del UUID del integrante se rechazan (CA-52.4)
    Dado que el proyecto "Inventario" tiene el integrante "Ana Pérez"
    Cuando pido PUT a la ruta del integrante "Ana Pérez" escrito <forma> con el cuerpo:
      """
      {"nombre":"Ana María","rol":"AGILE_ENABLER"}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del integrante debe ser un UUID"

    Ejemplos:
      | forma                  |
      | entre llaves           |
      | con urn:uuid: adelante |
      | sin guiones            |

  # Caso de error
  Escenario: Un UUID que no corresponde a ningún integrante (CA-52.4)
    Dado que el proyecto "Inventario" tiene el integrante "Ana Pérez"
    Cuando pido PUT "/api/integrantes/6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f" con el cuerpo:
      """
      {"nombre":"Ana María","rol":"AGILE_ENABLER"}
      """
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un integrante con ese id"

  # Caso límite
  Escenario: El UUID nulo no corresponde a ningún integrante (CA-52.4)
    Dado que el proyecto "Inventario" tiene el integrante "Ana Pérez"
    Cuando pido PUT "/api/integrantes/00000000-0000-0000-0000-000000000000" con el cuerpo:
      """
      {"nombre":"Ana María","rol":"AGILE_ENABLER"}
      """
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un integrante con ese id"

  # Caso límite
  Escenario: El ID del integrante se valida antes que el cuerpo (CA-52.4)
    Cuando pido PUT "/api/integrantes/abc" con el cuerpo:
      """
      esto no es JSON
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del integrante debe ser un UUID"

  # Caso límite
  Escenario: Con un UUID inexistente se conserva el orden de errores de HU-16 (CA-52.4)
    Cuando pido PUT "/api/integrantes/6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f" con el cuerpo:
      """
      {"nombre":"  ","rol":"AGILE_ENABLER"}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el nombre del integrante es obligatorio"
