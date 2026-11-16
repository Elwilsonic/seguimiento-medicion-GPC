# language: es
# HU-52 · Identificadores UUID del resto de las entidades, parte de ítems (Issue #197), vinculada a HU-04 a HU-09 y HU-20
# Especificación (OpenSpec): openspec/specs/backlog/spec.md
# Los escenarios son a nivel de la API y prueban solo las rutas de ítems cuyo
# handler está en items.go (tabla "Rutas por responsable" de D3 en design.md):
# GET y PUT /api/items/{id}, PATCH /api/items/{id}/estado y /story-points, y la
# creación y el listado del backlog. Las rutas que reciben el ítem desde otros
# archivos (Planning Poker, defectos, esfuerzos y sprints) las prueba su
# responsable. Los 1000 ítems (CA-52.2), la concurrencia (CA-52.7),
# PostgreSQL (CA-52.5) y la interfaz (CA-52.6) se cubren con tests de Go.

Característica: HU-52 Identificadores UUID de los ítems del backlog
  Como integrante del equipo,
  quiero que los ítems del backlog se identifiquen con un UUID en lugar de un número correlativo,
  para que no se pueda adivinar ni recorrer ningún ítem, ni deducir cuántos hay.

  Antecedentes:
    Dado que existe el proyecto "Inventario"

  # Caso normal
  Escenario: Crear un ítem le asigna un UUID versión 4 (CA-52.1)
    Cuando pido POST "/api/proyectos/{Inventario}/backlog" con el cuerpo:
      """
      {"titulo":"Login de usuarios","descripcion":"Acceso","prioridad":"ALTA","criterios":["Credenciales válidas"]}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto

  # Caso límite
  Escenario: El primer ítem no es el 1 (CA-52.1)
    Dado que el proyecto "Inventario" no tiene ítems
    Cuando pido POST "/api/proyectos/{Inventario}/backlog" con el cuerpo:
      """
      {"titulo":"Primero","descripcion":"","prioridad":"BAJA","criterios":["Ok"]}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta no es un número ni el UUID nulo

  # Caso normal
  Escenario: Consultar un ítem por su UUID (CA-52.3)
    Dado que el proyecto "Inventario" tiene el ítem "Login de usuarios"
    Cuando pido GET "/api/items/{Login de usuarios}"
    Entonces la respuesta tiene el código 200
    Y el "titulo" de la respuesta es "Login de usuarios"
    Y el "id" de la respuesta es el UUID de "Login de usuarios"

  # Caso alternativo
  Escenario: El UUID del ítem se acepta en mayúsculas (CA-52.3)
    Dado que el proyecto "Inventario" tiene el ítem "Login de usuarios"
    Cuando pido GET a la ruta del ítem "Login de usuarios" escrito en mayúsculas
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto
    Y el "id" de la respuesta es el UUID de "Login de usuarios"

  # Caso normal
  Escenario: Modificar, cambiar el estado y estimar un ítem por su UUID (CA-52.3)
    Dado que el proyecto "Inventario" tiene el ítem "Login de usuarios"
    Cuando pido PUT "/api/items/{Login de usuarios}" con el cuerpo:
      """
      {"titulo":"Login con Google","descripcion":"OAuth","prioridad":"MEDIA","criterios":["Redirige a Google"],"horas_estimadas":6.5}
      """
    Entonces la respuesta tiene el código 200
    Y el "titulo" de la respuesta es "Login con Google"
    Y el "id" de la respuesta es el UUID de "Login de usuarios"
    Cuando pido PATCH "/api/items/{Login de usuarios}/estado" con el cuerpo:
      """
      {"estado":"EN_CURSO"}
      """
    Entonces la respuesta tiene el código 200
    Y el "estado" de la respuesta es "EN_CURSO"
    Y el "id" de la respuesta es el UUID de "Login de usuarios"
    Cuando pido PATCH "/api/items/{Login de usuarios}/story-points" con el cuerpo:
      """
      {"story_points":5}
      """
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es el UUID de "Login de usuarios"

  # Caso alternativo
  Escenario: El backlog identifica cada ítem con su UUID y un ítem sin sprint con "" (CA-52.3)
    Dado que el proyecto "Inventario" tiene el ítem "Login de usuarios"
    Y que el proyecto "Inventario" tiene el ítem "Reportes"
    Cuando pido GET "/api/proyectos/{Inventario}/backlog"
    Entonces la respuesta tiene el código 200
    Y cada ítem de la lista tiene su UUID en "id", distinto del de los demás
    Y cada ítem de la lista tiene "" en "sprint_id"

  # Caso de error
  Esquema del escenario: Un ID numérico de ítem se rechaza en cada ruta (CA-52.4)
    Dado que el proyecto "Inventario" tiene el ítem "Login de usuarios"
    Cuando pido <método> "<ruta>" con el cuerpo:
      """
      <cuerpo>
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del ítem debe ser un UUID"

    Ejemplos:
      | método | ruta                       | cuerpo                                                                  |
      | GET    | /api/items/1               |                                                                         |
      | PUT    | /api/items/1               | {"titulo":"Login","prioridad":"ALTA","criterios":["Ok"]}                |
      | PATCH  | /api/items/1/estado        | {"estado":"EN_CURSO"}                                                   |
      | PATCH  | /api/items/1/story-points  | {"story_points":5}                                                      |

  # Caso límite
  Esquema del escenario: Otras formas del UUID del ítem se rechazan (CA-52.4)
    Dado que el proyecto "Inventario" tiene el ítem "Login de usuarios"
    Cuando pido GET a la ruta del ítem "Login de usuarios" escrito <forma>
    Entonces la respuesta tiene el código 400
    Y el error es "el id del ítem debe ser un UUID"

    Ejemplos:
      | forma                  |
      | entre llaves           |
      | con urn:uuid: adelante |
      | sin guiones            |

  # Caso de error
  Escenario: Un UUID que no corresponde a ningún ítem (CA-52.4)
    Dado que el proyecto "Inventario" tiene el ítem "Login de usuarios"
    Cuando pido GET "/api/items/6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f"
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un ítem del backlog con ese id"

  # Caso límite
  Escenario: El UUID nulo no corresponde a ningún ítem (CA-52.4)
    Dado que el proyecto "Inventario" tiene el ítem "Login de usuarios"
    Cuando pido PATCH "/api/items/00000000-0000-0000-0000-000000000000/estado" con el cuerpo:
      """
      {"estado":"EN_CURSO"}
      """
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un ítem del backlog con ese id"

  # Caso límite
  Escenario: El ID del ítem se valida antes que el cuerpo (CA-52.4)
    Cuando pido PUT "/api/items/abc" con el cuerpo:
      """
      esto no es JSON
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del ítem debe ser un UUID"

  # Caso límite
  Escenario: Con un UUID inexistente se conserva el orden de errores de HU-07 (CA-52.4)
    Cuando pido PUT "/api/items/6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f" con el cuerpo:
      """
      {"titulo":"  ","prioridad":"ALTA","criterios":["Ok"]}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el título del ítem es obligatorio"
