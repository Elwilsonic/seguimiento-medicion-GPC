## Purpose

Registros de esfuerzo (épica E6): registrar las horas trabajadas en una historia dentro de un sprint, consultarlas por historia o por sprint y corregirlas. El comportamiento anterior a HU-52 está en las SDD de HU-26 a HU-28 (`docs/especificaciones/`).

## ADDED Requirements

### Requirement: Identificador público UUID al registrar un esfuerzo
Al registrar un esfuerzo, el sistema SHALL asignarle un identificador público UUID versión 4, en forma canónica de 36 caracteres en minúsculas. Es el único identificador del esfuerzo que muestran la API y la interfaz, y reemplaza hacia afuera al entero correlativo de RN-26.8, que sigue como clave interna. `POST /api/sprints/{id}/items/{itemId}/esfuerzos` responde 201 con ese UUID en `id`, y con los UUID del ítem y del sprint en `item_id` y `sprint_id`. (HU-52, CA-52.1)

#### Scenario: Registrar un esfuerzo (caso normal)
- **WHEN** se registran 2.5 horas de "García Santiago" en un ítem asignado a un sprint abierto
- **THEN** responde 201, el `id` del esfuerzo es un texto con la forma de un UUID versión 4, y `item_id` y `sprint_id` son los UUID del ítem y del sprint

#### Scenario: El primer esfuerzo no es el 1 (caso límite)
- **WHEN** no hay esfuerzos y se registra el primero
- **THEN** su `id` no es "1" ni ningún número, y no es el UUID nulo

### Requirement: Identificadores de esfuerzos únicos y no deducibles
Dos esfuerzos SHALL NOT tener el mismo identificador público, y a partir del de un esfuerzo SHALL NOT poderse deducir el de otro ni la cantidad de esfuerzos. (HU-52, CA-52.2)

#### Scenario: Muchos esfuerzos registrados seguidos (caso alternativo)
- **WHEN** se registran 1000 esfuerzos
- **THEN** los 1000 identificadores públicos son distintos, todos son UUID versión 4, y ordenarlos no da el orden de registro

### Requirement: Registros simultáneos de esfuerzos
Registrar esfuerzos al mismo tiempo SHALL generar identificadores públicos únicos, sin condiciones de carrera (`go test -race` en la CI). (HU-52, CA-52.7)

#### Scenario: Registrar esfuerzos en paralelo
- **WHEN** 50 registros de esfuerzos corren al mismo tiempo en el mismo sprint abierto
- **THEN** se registran los 50, con 50 identificadores públicos distintos

### Requirement: Corregir un esfuerzo por su UUID
`PATCH /api/esfuerzos/{id}/horas` SHALL identificar el esfuerzo por su UUID, con mayúsculas aceptadas. La existencia del esfuerzo se valida igual que hoy (HU-28, `ErrEffortNotFound`), porque la corrección necesita el esfuerzo guardado y su sprint abierto. (HU-52, CA-52.3)

#### Scenario: Corregir las horas por UUID (caso normal)
- **WHEN** se pide `PATCH /api/esfuerzos/{UUID del esfuerzo}/horas` con `{"horas":3}`
- **THEN** responde 200 con el esfuerzo corregido y con el mismo UUID en `id`

#### Scenario: UUID del esfuerzo en mayúsculas (caso alternativo)
- **WHEN** se pide `PATCH /api/esfuerzos/{su UUID en mayúsculas}/horas` con `{"horas":3}`
- **THEN** responde 200 y el `id` de la respuesta está en minúsculas

### Requirement: Ninguna respuesta expone el ID entero de un esfuerzo
Toda respuesta de la API que identifica un esfuerzo SHALL hacerlo con su UUID en `id`: el registro, la corrección, las listas por historia y por sprint de HU-27, los reportes de HU-38 y HU-39, y el texto del PDF de HU-40. (HU-52, CA-52.3)

#### Scenario: La lista de esfuerzos de un sprint (caso alternativo)
- **WHEN** se pide `GET /api/sprints/{UUID del sprint}/esfuerzos` con dos esfuerzos registrados
- **THEN** cada esfuerzo tiene su UUID en `id` y los UUID de su ítem y de su sprint, y ninguno tiene un número

### Requirement: Identificador de esfuerzo con formato inválido
Si el `{id}` de `PATCH /api/esfuerzos/{id}/horas` no tiene la forma canónica de un UUID, la API SHALL responder 400 con `{"error":"el id del esfuerzo debe ser un UUID"}`, sin consultar el esfuerzo. Se rechazan los números, las llaves, el prefijo `urn:uuid:` y la forma sin guiones. (HU-52, CA-52.4)

#### Scenario: Un ID numérico de esfuerzo (caso de error)
- **WHEN** se pide `PATCH /api/esfuerzos/1/horas` con `{"horas":3}`
- **THEN** responde 400 con `{"error":"el id del esfuerzo debe ser un UUID"}`, y no 404

#### Scenario: El ID del esfuerzo se valida antes que el cuerpo (caso límite)
- **WHEN** se pide `PATCH /api/esfuerzos/abc/horas` con un cuerpo que no es JSON
- **THEN** responde 400 con el mensaje del ID del esfuerzo

### Requirement: Esfuerzo inexistente
Si el `{id}` tiene la forma de un UUID pero no corresponde a ningún esfuerzo, la API SHALL responder 404 con `{"error":"no existe un esfuerzo con ese id"}`. Las validaciones de contenido conservan su orden: siguen antes que la existencia. (HU-52, CA-52.4)

#### Scenario: UUID de esfuerzo que no existe (caso de error)
- **WHEN** se pide `PATCH /api/esfuerzos/{UUID versión 4 que nunca se asignó}/horas` con `{"horas":3}`
- **THEN** responde 404 con `{"error":"no existe un esfuerzo con ese id"}`

#### Scenario: Se conserva el orden de errores (caso límite)
- **WHEN** se pide `PATCH /api/esfuerzos/{UUID inexistente}/horas` con `{"horas":0}`
- **THEN** responde 400 con `{"error":"las horas trabajadas deben ser un número positivo"}`, igual que hoy con un ID entero inexistente (HU-28)

### Requirement: Esfuerzos en memoria y en PostgreSQL
El identificador público del esfuerzo SHALL guardarse y buscarse igual en memoria y en PostgreSQL, en una columna UUID única de la tabla de esfuerzos. Las relaciones con su ítem y su sprint siguen con los enteros. (HU-52, CA-52.5)

#### Scenario: Guardar y buscar un esfuerzo por su UUID
- **WHEN** se guarda un esfuerzo y después se lo busca por su identificador público, en memoria y en PostgreSQL
- **THEN** los dos repositorios devuelven el mismo esfuerzo, y un UUID que no existe da `ErrEffortNotFound`

#### Scenario: Dos esfuerzos con el mismo UUID en PostgreSQL (caso de error)
- **WHEN** se intenta guardar en PostgreSQL un esfuerzo con un identificador público que ya usa otro
- **THEN** la base lo rechaza y el guardado devuelve un error, sin crear el esfuerzo

### Requirement: La interfaz web usa los UUID en esfuerzos
En la interfaz web, «Ítem (ID)» de "Registrar esfuerzo", «Esfuerzo (ID)» de "Corregir esfuerzo" e «ID» de "Consultar esfuerzos" (por sprint o por historia) SHALL aceptar el UUID como texto. (HU-52, CA-52.6)

#### Scenario: Consultar los esfuerzos de una historia (caso normal)
- **WHEN** el usuario elige "Historia", escribe el UUID del ítem en «ID» y consulta
- **THEN** la interfaz muestra los esfuerzos de esa historia

#### Scenario: Escribir un ID de esfuerzo que no es UUID (caso de error)
- **WHEN** el usuario escribe "1" en «Esfuerzo (ID)» y corrige las horas
- **THEN** la interfaz muestra "el id del esfuerzo debe ser un UUID" y no corrige nada
