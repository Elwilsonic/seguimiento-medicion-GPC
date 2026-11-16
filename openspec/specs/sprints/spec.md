# sprints Specification

## Purpose

Sprints de un proyecto (épica E3): crear un sprint, asignarle ítems, registrar los completados, cerrarlo y consultar los sprints del proyecto. El comportamiento anterior a HU-52 está en las SDD de HU-10 a HU-14 (`docs/especificaciones/`).

## Requirements

### Requirement: Identificador público UUID al crear un sprint
Al crear un sprint, el sistema SHALL asignarle un identificador público UUID versión 4, en forma canónica de 36 caracteres en minúsculas. Es el único identificador del sprint que muestran la API y la interfaz, y reemplaza hacia afuera al entero correlativo de RN-10.6, que sigue como clave interna. `POST /api/proyectos/{id}/sprints` responde 201 con ese UUID en `id`. (HU-52, CA-52.1)

#### Scenario: Crear un sprint (caso normal)
- **WHEN** se crea el sprint "Sprint 1" en un proyecto
- **THEN** responde 201 y el `id` del sprint es un texto con la forma de un UUID versión 4

#### Scenario: El primer sprint no es el 1 (caso límite)
- **WHEN** no hay sprints y se crea el primero
- **THEN** su `id` no es "1" ni ningún número, y no es el UUID nulo

### Requirement: Identificadores de sprints únicos y no deducibles
Dos sprints SHALL NOT tener el mismo identificador público, y a partir del de un sprint SHALL NOT poderse deducir el de otro ni la cantidad de sprints. (HU-52, CA-52.2)

#### Scenario: Muchos sprints creados seguidos (caso alternativo)
- **WHEN** se crean 1000 sprints, cerrando cada uno antes de crear el siguiente (HU-10 permite un solo sprint abierto por proyecto)
- **THEN** los 1000 identificadores públicos son distintos, todos son UUID versión 4, y ordenarlos no da el orden de creación

### Requirement: Creaciones simultáneas de sprints
Crear sprints al mismo tiempo, en proyectos distintos, SHALL generar identificadores públicos únicos, sin condiciones de carrera (`go test -race` en la CI). (HU-52, CA-52.7)

#### Scenario: Crear sprints en paralelo
- **WHEN** 50 creaciones de sprints corren al mismo tiempo, cada una en su proyecto
- **THEN** se crean los 50, con 50 identificadores públicos distintos

### Requirement: Operar un sprint por su UUID
Toda ruta que recibe un sprint por su `{id}` SHALL identificarlo por su UUID: `POST /api/sprints/{id}/items/{itemId}`, `/completar` y `/esfuerzos`, `GET /api/sprints/{id}/esfuerzos`, `/completados`, `/defectos`, `/metricas` y `/reporte`, y `POST /api/sprints/{id}/cerrar`. Se aceptan mayúsculas. La existencia del sprint se valida igual que hoy (`ErrSprintNotFound`), porque cada operación necesita el sprint guardado. (HU-52, CA-52.3)

#### Scenario: Cerrar un sprint por su UUID (caso normal)
- **WHEN** existe el sprint abierto "Sprint 1" y se pide `POST /api/sprints/{su UUID}/cerrar`
- **THEN** responde 200 con el sprint en estado CERRADO y con el mismo UUID en `id`

#### Scenario: UUID del sprint en mayúsculas (caso alternativo)
- **WHEN** se pide `GET /api/sprints/{su UUID en mayúsculas}/metricas`
- **THEN** responde 200 con las métricas del sprint

### Requirement: Ninguna respuesta expone el ID entero de un sprint
Toda respuesta de la API que identifica un sprint SHALL hacerlo con su UUID: el `id` del sprint (también en las listas, en `sprint_actual` del resumen de HU-36, en las métricas por sprint de HU-37 y en el reporte de HU-39), el `sprint_id` de los ítems, de los esfuerzos y de las métricas del sprint, y los sprints del texto del PDF de HU-40. Un ítem sin sprint asignado SHALL informar `"sprint_id": ""` en lugar de `0`. (HU-52, CA-52.3)

#### Scenario: El sprint de un ítem asignado (caso alternativo)
- **WHEN** se consulta un ítem asignado al sprint "Sprint 1"
- **THEN** su `sprint_id` es el UUID de "Sprint 1", como texto

#### Scenario: Un ítem sin sprint (caso límite)
- **WHEN** se consulta un ítem que no está asignado a ningún sprint
- **THEN** su `sprint_id` es `""`, y no `0`

### Requirement: Identificador de sprint con formato inválido
Si el `{id}` de una ruta de sprint no tiene la forma canónica de un UUID, la API SHALL responder 400 con `{"error":"el id del sprint debe ser un UUID"}`, sin consultar el sprint. Se rechazan los números, las llaves, el prefijo `urn:uuid:` y la forma sin guiones. (HU-52, CA-52.4)

#### Scenario: Un ID numérico de sprint (caso de error)
- **WHEN** se pide `GET /api/sprints/1/metricas`
- **THEN** responde 400 con `{"error":"el id del sprint debe ser un UUID"}`, y no 404

#### Scenario: Primero se valida el sprint y después el ítem (caso límite)
- **WHEN** se pide `POST /api/sprints/abc/items/xyz`
- **THEN** responde 400 con el mensaje del ID del sprint, porque `{id}` se lee antes que `{itemId}`, igual que hoy

### Requirement: Sprint inexistente
Si el `{id}` tiene la forma de un UUID pero no corresponde a ningún sprint, la API SHALL responder 404 con `{"error":"no existe un sprint con ese id"}`. Las validaciones de contenido conservan su orden: siguen antes que la existencia del sprint. (HU-52, CA-52.4)

#### Scenario: UUID de sprint que no existe (caso de error)
- **WHEN** se pide `GET /api/sprints/{UUID versión 4 que nunca se asignó}/metricas`
- **THEN** responde 404 con `{"error":"no existe un sprint con ese id"}`

#### Scenario: Se conserva el orden de errores al registrar un esfuerzo (caso límite)
- **WHEN** se pide `POST /api/sprints/{UUID inexistente}/items/{UUID de un ítem}/esfuerzos` con el integrante vacío
- **THEN** responde 400 con `{"error":"el integrante es obligatorio"}`, igual que hoy con un ID entero inexistente (HU-26)

### Requirement: Sprints en memoria y en PostgreSQL
El identificador público del sprint SHALL guardarse y buscarse igual en memoria y en PostgreSQL, en una columna UUID única de la tabla de sprints. Las relaciones con su proyecto, sus ítems, sus esfuerzos y sus defectos siguen con los enteros. (HU-52, CA-52.5)

#### Scenario: Guardar y buscar un sprint por su UUID
- **WHEN** se guarda un sprint y después se lo busca por su identificador público, en memoria y en PostgreSQL
- **THEN** los dos repositorios devuelven el mismo sprint, y un UUID que no existe da `ErrSprintNotFound`

#### Scenario: Dos sprints con el mismo UUID en PostgreSQL (caso de error)
- **WHEN** se intenta guardar en PostgreSQL un sprint con un identificador público que ya usa otro
- **THEN** la base lo rechaza y el guardado devuelve un error, sin crear el sprint

### Requirement: La interfaz web usa el UUID del sprint
En la interfaz web, el campo «Sprint actual (ID)» SHALL aceptar el UUID como texto. Crear un sprint y «Elegir» en la tabla de "Ver sprints del proyecto" (HU-14) SHALL dejar el UUID en ese campo, y las secciones que usan el sprint actual lo envían en la ruta. (HU-52, CA-52.6)

#### Scenario: Elegir un sprint de la lista (caso normal)
- **WHEN** el usuario pulsa «Ver sprints del proyecto» y «Elegir» en la fila de "Sprint 1"
- **THEN** «Sprint actual» muestra el UUID de "Sprint 1"

#### Scenario: Escribir un ID de sprint que no es UUID (caso de error)
- **WHEN** el usuario escribe "1" en «Sprint actual» y consulta sus métricas
- **THEN** la interfaz muestra "el id del sprint debe ser un UUID"
