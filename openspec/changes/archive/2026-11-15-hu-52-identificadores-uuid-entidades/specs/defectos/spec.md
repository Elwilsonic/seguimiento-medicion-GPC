## Purpose

Defectos (épica E7): registrar un defecto sobre una historia del backlog, cambiar su estado y consultarlos por proyecto o por sprint. El comportamiento anterior a HU-52 está en las SDD de HU-29 a HU-31 (`docs/especificaciones/`).

## ADDED Requirements

### Requirement: Identificador público UUID al registrar un defecto
Al registrar un defecto, el sistema SHALL asignarle un identificador público UUID versión 4, en forma canónica de 36 caracteres en minúsculas. Es el único identificador del defecto que muestran la API y la interfaz, y reemplaza hacia afuera al entero correlativo de RN-29.8, que sigue como clave interna. `POST /api/items/{id}/defectos` responde 201 con ese UUID en `id`. (HU-52, CA-52.1)

#### Scenario: Registrar un defecto (caso normal)
- **WHEN** se registra un defecto ALTA sobre un ítem, con el UUID de un sprint abierto en `sprint_deteccion_id`
- **THEN** responde 201, el `id` del defecto es un texto con la forma de un UUID versión 4, `item_id` y `sprint_deteccion_id` son los UUID del ítem y del sprint, y `sprint_resolucion_id` es `""`

#### Scenario: El primer defecto no es el 1 (caso límite)
- **WHEN** no hay defectos y se registra el primero
- **THEN** su `id` no es "1" ni ningún número, y no es el UUID nulo

### Requirement: Identificadores de defectos únicos y no deducibles
Dos defectos SHALL NOT tener el mismo identificador público, y a partir del de un defecto SHALL NOT poderse deducir el de otro ni la cantidad de defectos. (HU-52, CA-52.2)

#### Scenario: Muchos defectos registrados seguidos (caso alternativo)
- **WHEN** se registran 1000 defectos
- **THEN** los 1000 identificadores públicos son distintos, todos son UUID versión 4, y ordenarlos no da el orden de registro

### Requirement: Registros simultáneos de defectos
Registrar defectos al mismo tiempo SHALL generar identificadores públicos únicos, sin condiciones de carrera (`go test -race` en la CI). (HU-52, CA-52.7)

#### Scenario: Registrar defectos en paralelo
- **WHEN** 50 registros de defectos corren al mismo tiempo
- **THEN** se registran los 50, con 50 identificadores públicos distintos

### Requirement: Operar un defecto y sus sprints por UUID
`PATCH /api/defectos/{id}/estado` SHALL identificar el defecto por su UUID, con mayúsculas aceptadas. Los campos del cuerpo `sprint_deteccion_id` (registro, HU-29) y `sprint_resolucion_id` (resolución, HU-30) SHALL recibir el UUID del sprint, como texto. La existencia del defecto y la de esos sprints se validan igual que hoy, porque las reglas de HU-29 y HU-30 necesitan el defecto y los sprints guardados (sprint abierto, del mismo proyecto y no anterior al de detección). (HU-52, CA-52.3)

#### Scenario: Resolver un defecto por UUID (caso normal)
- **WHEN** se pide `PATCH /api/defectos/{UUID del defecto}/estado` con `{"estado":"RESUELTO","sprint_resolucion_id":"{UUID de un sprint}"}`, después de pasarlo a EN_CORRECCION
- **THEN** responde 200 con el defecto RESUELTO, y `sprint_resolucion_id` es el UUID de ese sprint

#### Scenario: UUID del defecto en mayúsculas (caso alternativo)
- **WHEN** se pide `PATCH /api/defectos/{su UUID en mayúsculas}/estado` con `{"estado":"EN_CORRECCION"}`
- **THEN** responde 200 y el `id` de la respuesta está en minúsculas

### Requirement: Ninguna respuesta expone el ID entero de un defecto
Toda respuesta de la API que identifica un defecto SHALL hacerlo con su UUID en `id`, también en las listas de HU-31 y en los reportes de HU-38 y HU-39, y en el texto del PDF de HU-40. Un defecto sin resolver SHALL informar `"sprint_resolucion_id": ""` en lugar de `0`. (HU-52, CA-52.3)

#### Scenario: La lista de defectos de un proyecto (caso alternativo)
- **WHEN** se pide `GET /api/proyectos/{id}/defectos` con un defecto abierto
- **THEN** el defecto tiene su UUID en `id`, los UUID de su ítem y de su sprint de detección, y `""` en `sprint_resolucion_id`

### Requirement: Identificadores de defecto y de sus sprints con formato inválido
Si el `{id}` de `PATCH /api/defectos/{id}/estado` no tiene la forma canónica de un UUID, la API SHALL responder 400 con `{"error":"el id del defecto debe ser un UUID"}`, sin consultar el defecto. Si `sprint_deteccion_id`, o `sprint_resolucion_id` al resolver, es un texto que no tiene esa forma, la API SHALL responder 400 con `{"error":"el id del sprint debe ser un UUID"}`. Un campo ausente o vacío sigue siendo un sprint faltante (400 de HU-29 o HU-30). (HU-52, CA-52.4)

#### Scenario: Un ID numérico de defecto (caso de error)
- **WHEN** se pide `PATCH /api/defectos/1/estado` con `{"estado":"EN_CORRECCION"}`
- **THEN** responde 400 con `{"error":"el id del defecto debe ser un UUID"}`, y no 404

#### Scenario: El sprint de detección escrito como número (caso de error)
- **WHEN** se registra un defecto con `"sprint_deteccion_id":"3"`
- **THEN** responde 400 con `{"error":"el id del sprint debe ser un UUID"}`

#### Scenario: El sprint de detección ausente (caso límite)
- **WHEN** se registra un defecto sin `sprint_deteccion_id`, o con `""`
- **THEN** responde 400 con `{"error":"el sprint de detección es obligatorio"}`, igual que hoy (HU-29)

### Requirement: Defecto o sprint inexistente
Si el `{id}` tiene la forma de un UUID pero no corresponde a ningún defecto, la API SHALL responder 404 con `{"error":"no existe un defecto con ese id"}`. Si `sprint_deteccion_id` o `sprint_resolucion_id` tiene la forma de un UUID pero no corresponde a ningún sprint, la API SHALL responder 404 con `{"error":"no existe un sprint con ese id"}`, y no el 400 de sprint faltante. Las validaciones de contenido conservan su orden. (HU-52, CA-52.4)

#### Scenario: UUID de defecto que no existe (caso de error)
- **WHEN** se pide `PATCH /api/defectos/{UUID versión 4 que nunca se asignó}/estado` con `{"estado":"EN_CORRECCION"}`
- **THEN** responde 404 con `{"error":"no existe un defecto con ese id"}`

#### Scenario: Sprint de detección que no existe (caso de error)
- **WHEN** se registra un defecto con un UUID de sprint que nunca se asignó en `sprint_deteccion_id`
- **THEN** responde 404 con `{"error":"no existe un sprint con ese id"}`, igual que hoy con un ID positivo inexistente (HU-29)

#### Scenario: Se conserva el orden de errores al resolver (caso límite)
- **WHEN** se pide `PATCH /api/defectos/{UUID inexistente}/estado` con `{"estado":"RESUELTO"}` y sin `sprint_resolucion_id`
- **THEN** responde 400 con `{"error":"el sprint de resolución es obligatorio para resolver un defecto"}`, igual que hoy (HU-30)

### Requirement: Defectos en memoria y en PostgreSQL
El identificador público del defecto SHALL guardarse y buscarse igual en memoria y en PostgreSQL, en una columna UUID única de la tabla de defectos. Las relaciones con su ítem y sus sprints siguen con los enteros. (HU-52, CA-52.5)

#### Scenario: Guardar y buscar un defecto por su UUID
- **WHEN** se guarda un defecto y después se lo busca por su identificador público, en memoria y en PostgreSQL
- **THEN** los dos repositorios devuelven el mismo defecto, y un UUID que no existe da `ErrDefectNotFound`

#### Scenario: Dos defectos con el mismo UUID en PostgreSQL (caso de error)
- **WHEN** se intenta guardar en PostgreSQL un defecto con un identificador público que ya usa otro
- **THEN** la base lo rechaza y el guardado devuelve un error, sin crear el defecto

### Requirement: La interfaz web usa los UUID en defectos
En la interfaz web, los campos «Ítem (ID)» y «Sprint de detección (ID)» de "Registrar defecto", y «Defecto (ID)» y «Sprint de resolución (ID)» de "Cambiar estado de defecto", SHALL aceptar el UUID como texto, y la interfaz SHALL enviar los sprints como texto, sin convertirlos a número. (HU-52, CA-52.6)

#### Scenario: Registrar un defecto desde la interfaz (caso normal)
- **WHEN** el usuario completa «Ítem (ID)» y «Sprint de detección (ID)» con UUID válidos y registra el defecto
- **THEN** la interfaz muestra el defecto creado con su UUID

#### Scenario: Escribir un sprint de detección que no es UUID (caso de error)
- **WHEN** el usuario escribe "1" en «Sprint de detección (ID)» y registra el defecto
- **THEN** la interfaz muestra "el id del sprint debe ser un UUID"
