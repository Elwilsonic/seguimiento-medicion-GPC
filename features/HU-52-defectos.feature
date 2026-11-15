# language: es
# HU-52 · Identificadores UUID del resto de las entidades, parte de defectos (Issue #197), vinculada a HU-29 (#47), HU-30 (#48) y HU-31 (#49)
# Especificación (OpenSpec): openspec/changes/hu-52-identificadores-uuid-entidades/specs/defectos/spec.md
# Los escenarios son a nivel de la API y prueban solo las rutas de defectos.go
# (D3 de design.md): POST /api/items/{id}/defectos, PATCH /api/defectos/{id}/estado
# y GET /api/sprints/{id}/defectos, más la lista por proyecto, que ya recibe el
# UUID del proyecto (HU-51). El proyecto, el ítem, el sprint y los defectos de
# los pasos "Dado" se crean con el dominio, sin pasar por las rutas de ítems y
# sprints, que son de otros responsables.
# En las rutas y en los cuerpos, {ítem}, {sprint}, {proyecto} y {defecto} son
# los UUID de esas entidades; {inexistente} es un UUID versión 4 que nunca se
# asigna; {defecto en mayúsculas} es el UUID del defecto en mayúsculas.
# Los 1000 defectos (CA-52.2), las creaciones simultáneas (CA-52.7) y PostgreSQL
# (CA-52.5) son de dominio y los cubre la base (features/HU-52-base.feature);
# la interfaz (CA-52.6), con tests de Go.

Característica: HU-52 Identificadores UUID de los defectos
  Como integrante del equipo,
  quiero que los defectos se identifiquen con un UUID en lugar de un número correlativo,
  para que no se pueda adivinar ni recorrer ningún defecto, ni deducir cuántos hay.

  Antecedentes:
    Dado que existe un ítem asignado a un sprint abierto

  # Caso normal
  Escenario: Registrar un defecto (CA-52.1, CA-52.3)
    Cuando pido POST "/api/items/{ítem}/defectos" con el cuerpo:
      """
      {"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":"{sprint}"}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta es un UUID versión 4 en minúsculas, como texto
    Y el "item_id" de la respuesta es el UUID del ítem
    Y el "sprint_deteccion_id" de la respuesta es el UUID del sprint
    Y el "sprint_resolucion_id" de la respuesta es un texto vacío

  # Caso límite
  Escenario: El primer defecto no es el 1 (CA-52.1)
    Dado que no hay defectos registrados
    Cuando pido POST "/api/items/{ítem}/defectos" con el cuerpo:
      """
      {"descripcion":"Primer defecto","severidad":"BAJA","sprint_deteccion_id":"{sprint}"}
      """
    Entonces la respuesta tiene el código 201
    Y el "id" de la respuesta no es un número ni el UUID nulo

  # Caso normal
  Escenario: Resolver un defecto por UUID (CA-52.3)
    Dado que el ítem tiene un defecto registrado
    Y que el defecto está en EN_CORRECCION
    Cuando pido PATCH "/api/defectos/{defecto}/estado" con el cuerpo:
      """
      {"estado":"RESUELTO","sprint_resolucion_id":"{sprint}"}
      """
    Entonces la respuesta tiene el código 200
    Y el "estado" de la respuesta es "RESUELTO"
    Y el "id" de la respuesta es el UUID del defecto
    Y el "sprint_resolucion_id" de la respuesta es el UUID del sprint

  # Caso alternativo
  Escenario: UUID del defecto en mayúsculas (CA-52.3)
    Dado que el ítem tiene un defecto registrado
    Cuando pido PATCH "/api/defectos/{defecto en mayúsculas}/estado" con el cuerpo:
      """
      {"estado":"EN_CORRECCION"}
      """
    Entonces la respuesta tiene el código 200
    Y el "id" de la respuesta es el UUID del defecto

  # Caso alternativo
  Escenario: La lista de defectos de un proyecto (CA-52.3)
    Dado que el ítem tiene un defecto registrado
    Cuando pido GET "/api/proyectos/{proyecto}/defectos"
    Entonces la respuesta tiene el código 200
    Y la lista "defectos" tiene 1 elementos
    Y cada elemento de "defectos" tiene en "id" el UUID del defecto
    Y cada elemento de "defectos" tiene en "item_id" el UUID del ítem
    Y cada elemento de "defectos" tiene en "sprint_deteccion_id" el UUID del sprint
    Y cada elemento de "defectos" tiene en "sprint_resolucion_id" un texto vacío

  # Caso alternativo
  Escenario: La lista de defectos de un sprint por su UUID (CA-52.3)
    Dado que el ítem tiene un defecto registrado
    Cuando pido GET "/api/sprints/{sprint}/defectos"
    Entonces la respuesta tiene el código 200
    Y la lista "defectos" tiene 1 elementos
    Y cada elemento de "defectos" tiene en "id" el UUID del defecto
    Y cada elemento de "defectos" tiene en "sprint_deteccion_id" el UUID del sprint

  # Caso de error
  Escenario: Un ID numérico de defecto (CA-52.4)
    Dado que el ítem tiene un defecto registrado
    Cuando pido PATCH "/api/defectos/1/estado" con el cuerpo:
      """
      {"estado":"EN_CORRECCION"}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del defecto debe ser un UUID"

  # Caso de error
  Escenario: El sprint de detección escrito como número (CA-52.4)
    Cuando pido POST "/api/items/{ítem}/defectos" con el cuerpo:
      """
      {"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":"3"}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el id del sprint debe ser un UUID"

  # Caso límite
  Esquema del escenario: El sprint de detección ausente o vacío (CA-52.4)
    Cuando pido POST "/api/items/{ítem}/defectos" con el cuerpo:
      """
      <cuerpo>
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el sprint de detección es obligatorio"

    Ejemplos:
      | cuerpo                                                                 |
      | {"descripcion":"Sin sprint","severidad":"ALTA"}                        |
      | {"descripcion":"Sin sprint","severidad":"ALTA","sprint_deteccion_id":""} |

  # Caso de error
  Escenario: UUID de defecto que no existe (CA-52.4)
    Cuando pido PATCH "/api/defectos/{inexistente}/estado" con el cuerpo:
      """
      {"estado":"EN_CORRECCION"}
      """
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un defecto con ese id"

  # Caso de error
  Escenario: Sprint de detección que no existe (CA-52.4)
    Cuando pido POST "/api/items/{ítem}/defectos" con el cuerpo:
      """
      {"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":"{inexistente}"}
      """
    Entonces la respuesta tiene el código 404
    Y el error es "no existe un sprint con ese id"

  # Caso límite
  Escenario: Se conserva el orden de errores al resolver (CA-52.4)
    Cuando pido PATCH "/api/defectos/{inexistente}/estado" con el cuerpo:
      """
      {"estado":"RESUELTO"}
      """
    Entonces la respuesta tiene el código 400
    Y el error es "el sprint de resolución es obligatorio para resolver un defecto"
