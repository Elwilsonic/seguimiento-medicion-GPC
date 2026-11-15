# language: es
# HU-51 · Identificadores UUID, parte de proyectos (Issue #193), vinculada a HU-01 (#14), HU-03 (#16) y HU-50 (#183)
# Cambio (OpenSpec): openspec/changes/hu-51-identificadores-uuid/
# Los escenarios son a nivel de la API, porque lo observable de HU-51 son las
# rutas, los 400 y 404 y las salidas (D8 de design.md). Los 1000 proyectos
# (CA-51.2), PostgreSQL (CA-51.5), la interfaz (CA-51.6) y la concurrencia
# (CA-51.7) se cubren con tests de Go, no con escenarios.

Característica: HU-51 Identificadores UUID de los proyectos
  Como integrante del equipo,
  quiero que los proyectos se identifiquen con un UUID en lugar de un número correlativo,
  para que no se pueda adivinar el identificador de un proyecto ajeno ni deducir cuántos proyectos hay.

  # Caso normal
  Escenario: Crear un proyecto le asigna un UUID versión 4 (CA-51.1)
    Cuando pido POST "/api/proyectos" con el cuerpo:
      """
      {"nombre":"Sistema de ventas","descripcion":""}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto

  # Caso límite
  Escenario: El primer proyecto no es el 1 (CA-51.1)
    Dado que no hay proyectos creados
    Cuando pido POST "/api/proyectos" con el cuerpo:
      """
      {"nombre":"Primero","descripcion":""}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta no es un número ni el UUID nulo

  # Caso normal
  Escenario: Consultar un proyecto por su UUID (CA-51.3)
    Dado que existe el proyecto "Inventario"
    Cuando pido GET "/api/proyectos/{Inventario}"
    Entonces la respuesta tiene el código 200
    Y el "nombre" de la respuesta es "Inventario"
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto
    Y el "id" de la respuesta es el UUID de "Inventario"

  # Caso alternativo
  Escenario: El UUID se acepta en mayúsculas (CA-51.3)
    Dado que existe el proyecto "Inventario"
    Cuando pido GET a la ruta del proyecto "Inventario" escrito en mayúsculas
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto
    Y el "id" de la respuesta es el UUID de "Inventario"

  # Caso normal
  Escenario: Las operaciones que dependen del proyecto usan su UUID (CA-51.3)
    Dado que existe el proyecto "Inventario"
    Cuando pido POST "/api/proyectos/{Inventario}/backlog" con el cuerpo:
      """
      {"titulo":"Login","descripcion":"","prioridad":"ALTA","criterios":["Ingresa con su clave"]}
      """
    Entonces la respuesta tiene el código 201
    Y el "proyecto_id" de la respuesta es el UUID de "Inventario"
    Cuando pido POST "/api/proyectos/{Inventario}/sprints" con el cuerpo:
      """
      {"nombre":"Sprint 1","fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Gestionar el backlog"}
      """
    Entonces la respuesta tiene el código 201
    Y el "proyecto_id" de la respuesta es el UUID de "Inventario"
    Cuando pido POST "/api/proyectos/{Inventario}/integrantes" con el cuerpo:
      """
      {"nombre":"Ana Pérez","rol":"PRODUCT_BUILDER"}
      """
    Entonces la respuesta tiene el código 201
    Y el "proyecto_id" de la respuesta es el UUID de "Inventario"
    Y estas rutas del proyecto "Inventario" responden 200:
      | ruta              |
      | /backlog          |
      | /sprints          |
      | /integrantes      |
      | /metricas         |
      | /metricas/sprints |
      | /resumen          |
      | /reporte          |
      | /reporte/pdf      |

  # Caso alternativo
  Escenario: Un ítem consultado por su propio ID muestra el UUID de su proyecto (CA-51.3)
    Dado que existe el proyecto "Inventario"
    Y que el proyecto "Inventario" tiene un ítem en su backlog
    Cuando pido GET "/api/items/1"
    Entonces la respuesta tiene el código 200
    Y el "proyecto_id" de la respuesta es el UUID de "Inventario"

  # Caso de error
  Escenario: Un ID numérico de proyecto se rechaza (CA-51.4)
    Dado que existe el proyecto "Inventario"
    Cuando pido GET "/api/proyectos/1"
    Entonces la respuesta tiene el código 400
    Y el error es "el id del proyecto debe ser un UUID"

  # Caso límite
  Esquema del escenario: Otras formas del UUID se rechazan (CA-51.4)
    Dado que existe el proyecto "Inventario"
    Cuando pido GET a la ruta del proyecto "Inventario" escrito <forma>
    Entonces la respuesta tiene el código 400
    Y el error es "el id del proyecto debe ser un UUID"

    Ejemplos:
      | forma                  |
      | entre llaves           |
      | con urn:uuid: adelante |
      | sin guiones            |

  # Caso de error
  Escenario: Un UUID que no corresponde a ningún proyecto (CA-51.4)
    Dado que existe el proyecto "Inventario"
    Cuando pido GET "/api/proyectos/6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f"
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un proyecto con ese id"

  # Caso límite
  Escenario: El UUID nulo no corresponde a ningún proyecto (CA-51.4)
    Dado que existe el proyecto "Inventario"
    Cuando pido GET "/api/proyectos/00000000-0000-0000-0000-000000000000"
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un proyecto con ese id"

  # Caso límite
  Escenario: El ID de la ruta se valida antes que el cuerpo (CA-51.4)
    Cuando pido PATCH "/api/proyectos/abc/fecha-inicio" con el cuerpo:
      """
      esto no es JSON
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del proyecto debe ser un UUID"

  # Caso límite
  Escenario: Con un UUID inexistente se conserva el orden de errores de HU-18 (CA-51.4)
    Cuando pido PATCH "/api/proyectos/6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f/fecha-inicio" con el cuerpo:
      """
      {"fecha_inicio":""}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "la fecha de inicio del proyecto es obligatoria"

  # Caso límite
  Escenario: Las demás entidades conservan su ID entero (CA-51.4)
    Cuando pido GET "/api/items/abc"
    Entonces la respuesta tiene el código 400
    Y el error es "el id debe ser un número entero"
