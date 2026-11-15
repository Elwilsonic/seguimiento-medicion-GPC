## Purpose

Ítems del Product Backlog (épica E2): crear, consultar, listar, modificar, cambiar el estado, priorizar y estimar los ítems de un proyecto. El comportamiento anterior a HU-52 está en las SDD de HU-04 a HU-09 y HU-20 (`docs/especificaciones/`).

## ADDED Requirements

### Requirement: Identificador público UUID al crear un ítem
Al crear un ítem, el sistema SHALL asignarle un identificador público UUID versión 4, en forma canónica de 36 caracteres en minúsculas. Es el único identificador del ítem que muestran la API y la interfaz, y reemplaza hacia afuera al entero correlativo de RN-04.6, que sigue como clave interna. `POST /api/proyectos/{id}/backlog` responde 201 con ese UUID en `id`. (HU-52, CA-52.1)

#### Scenario: Crear un ítem (caso normal)
- **WHEN** se crea el ítem "Login de usuarios" en un proyecto
- **THEN** responde 201 y el `id` del ítem es un texto de 36 caracteres con la forma de un UUID versión 4

#### Scenario: El primer ítem no es el 1 (caso límite)
- **WHEN** no hay ítems y se crea el primero
- **THEN** su `id` no es "1" ni ningún número, y no es el UUID nulo

### Requirement: Identificadores de ítems únicos y no deducibles
Dos ítems SHALL NOT tener el mismo identificador público, y a partir del de un ítem SHALL NOT poderse deducir el de otro ni la cantidad de ítems. (HU-52, CA-52.2)

#### Scenario: Muchos ítems creados seguidos (caso alternativo)
- **WHEN** se crean 1000 ítems
- **THEN** los 1000 identificadores públicos son distintos, todos son UUID versión 4, y ordenarlos no da el orden de creación

### Requirement: Creaciones simultáneas de ítems
Crear ítems al mismo tiempo SHALL generar identificadores públicos únicos, sin condiciones de carrera (`go test -race` en la CI). (HU-52, CA-52.7)

#### Scenario: Crear ítems en paralelo
- **WHEN** 50 creaciones de ítems corren al mismo tiempo
- **THEN** se crean los 50, con 50 identificadores públicos distintos

### Requirement: Operar un ítem por su UUID
Toda ruta que recibe un ítem, por `/api/items/{id}` (7 rutas) o por el `{itemId}` de `/api/sprints/{id}/items/{itemId}` (3 rutas), SHALL identificarlo por su UUID, con mayúsculas aceptadas. La existencia del ítem se valida igual que hoy (HU-05, `ErrItemNotFound`), porque cada operación necesita el ítem guardado. (HU-52, CA-52.3)

#### Scenario: Consultar un ítem por su UUID (caso normal)
- **WHEN** existe el ítem "Login de usuarios" y se pide `GET /api/items/{su UUID}`
- **THEN** responde 200 con el ítem, en el formato de HU-05, y con el mismo UUID en `id`

#### Scenario: UUID del ítem en mayúsculas (caso alternativo)
- **WHEN** se pide `GET /api/items/{su UUID en mayúsculas}`
- **THEN** responde 200 y el `id` de la respuesta está en minúsculas

#### Scenario: Las demás rutas que reciben el ítem (caso normal)
- **WHEN** se piden `PUT /api/items/{UUID}`, `PATCH /api/items/{UUID}/estado` y `/story-points`, `POST /api/items/{UUID}/poker-sessions` y `/defectos`, y `GET /api/items/{UUID}/esfuerzos`, con cuerpos válidos
- **THEN** cada una responde como en su historia (HU-07, HU-08, HU-20, HU-21, HU-29 y HU-27), sobre el ítem de ese UUID

#### Scenario: Asignar un ítem a un sprint por sus UUID (caso normal)
- **WHEN** se pide `POST /api/sprints/{UUID del sprint}/items/{UUID del ítem}`
- **THEN** responde 200 y el ítem queda asignado a ese sprint

### Requirement: Ninguna respuesta expone el ID entero de un ítem
Toda respuesta de la API que identifica un ítem SHALL hacerlo con su UUID: el `id` del ítem (también en las listas del backlog, en los completados del sprint y en las historias de los reportes de HU-38, HU-39 y HU-49), el `item_id` de defectos, esfuerzos y sesiones de Planning Poker, y el ID del ítem en el texto del PDF de HU-40. (HU-52, CA-52.3)

#### Scenario: El ítem de un defecto (caso alternativo)
- **WHEN** se consulta un defecto registrado sobre el ítem "Login de usuarios"
- **THEN** su `item_id` es el UUID de ese ítem, como texto, y no un número

### Requirement: Identificador de ítem con formato inválido
Si el `{id}` o el `{itemId}` de una ruta de ítem no tiene la forma canónica de un UUID, la API SHALL responder 400 con `{"error":"el id del ítem debe ser un UUID"}`, sin consultar el ítem. Se rechazan los números, las llaves, el prefijo `urn:uuid:` y la forma sin guiones. (HU-52, CA-52.4)

#### Scenario: Un ID numérico de ítem (caso de error)
- **WHEN** se pide `GET /api/items/1`
- **THEN** responde 400 con `{"error":"el id del ítem debe ser un UUID"}`, y no 404

#### Scenario: El ID del ítem se valida antes que el cuerpo (caso límite)
- **WHEN** se pide `PUT /api/items/abc` con un cuerpo que no es JSON
- **THEN** responde 400 con el mensaje del ID del ítem, porque el ID de la ruta se valida primero

### Requirement: Ítem inexistente
Si el `{id}` o el `{itemId}` tiene la forma de un UUID pero no corresponde a ningún ítem, la API SHALL responder 404 con `{"error":"no existe un ítem del backlog con ese id"}`. Las validaciones de contenido de cada ruta conservan su orden: siguen antes que la existencia del ítem. (HU-52, CA-52.4)

#### Scenario: UUID de ítem que no existe (caso de error)
- **WHEN** se pide `GET /api/items/{UUID versión 4 que nunca se asignó}`
- **THEN** responde 404 con `{"error":"no existe un ítem del backlog con ese id"}`

#### Scenario: Se conserva el orden de errores al modificar (caso límite)
- **WHEN** se pide `PUT /api/items/{UUID inexistente}` con el título vacío
- **THEN** responde 400 con `{"error":"el título del ítem es obligatorio"}`, igual que hoy con un ID entero inexistente (HU-07)

### Requirement: Ítems en memoria y en PostgreSQL
El identificador público del ítem SHALL guardarse y buscarse igual en memoria y en PostgreSQL, en una columna UUID única de la tabla de ítems. Las relaciones con su proyecto, su sprint, sus defectos, esfuerzos y sesiones siguen con los enteros. (HU-52, CA-52.5)

#### Scenario: Guardar y buscar un ítem por su UUID
- **WHEN** se guarda un ítem y después se lo busca por su identificador público, en memoria y en PostgreSQL
- **THEN** los dos repositorios devuelven el mismo ítem, y un UUID que no existe da `ErrItemNotFound`

#### Scenario: Dos ítems con el mismo UUID en PostgreSQL (caso de error)
- **WHEN** se intenta guardar en PostgreSQL un ítem con un identificador público que ya usa otro
- **THEN** la base lo rechaza y el guardado devuelve un error, sin crear el ítem

### Requirement: La interfaz web usa el UUID del ítem
En la interfaz web, los campos que reciben un ítem («Ítem (ID)» de "Registrar esfuerzo" y de "Registrar defecto", y «ID» de "Consultar esfuerzos" con "Historia") SHALL aceptar el UUID como texto, y los botones de las filas del backlog SHALL seguir funcionando con el UUID. (HU-52, CA-52.6)

#### Scenario: Escribir un ID de ítem que no es UUID (caso de error)
- **WHEN** el usuario escribe "1" en «Ítem (ID)» de "Registrar defecto" y lo envía
- **THEN** la interfaz muestra "el id del ítem debe ser un UUID" y no registra nada
