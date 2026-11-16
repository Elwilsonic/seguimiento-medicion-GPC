## Purpose

Sesiones de Planning Poker (épica E4): abrir una sesión sobre una historia, votar en forma oculta, revelar, repetir la ronda y registrar la estimación acordada. El comportamiento anterior a HU-52 está en las SDD de HU-21 a HU-25 (`docs/especificaciones/`).

## ADDED Requirements

### Requirement: Identificador público UUID al abrir una sesión
Al abrir una sesión de Planning Poker, el sistema SHALL asignarle un identificador público UUID versión 4, en forma canónica de 36 caracteres en minúsculas, que es el único que muestran la API y la interfaz; el entero de RN-21.7 sigue como clave interna. Repetir la ronda (HU-24) SHALL conservar el UUID de la sesión. (HU-52, CA-52.1)

#### Scenario: Abrir una sesión (caso normal)
- **WHEN** se abre una sesión sobre el ítem "Login de usuarios"
- **THEN** `POST /api/items/{id}/poker-sessions` responde 201, el `id` de la sesión es un texto con la forma de un UUID versión 4, e `item_id` es el UUID del ítem

#### Scenario: Repetir la ronda conserva el UUID (caso límite)
- **WHEN** se revela una sesión con votos distintos y se pide `POST /api/poker-sessions/{su UUID}/rounds`
- **THEN** responde con la ronda 2 y el mismo UUID en `id`

### Requirement: Identificadores de sesiones únicos y no deducibles
Dos sesiones SHALL NOT tener el mismo identificador público, y a partir del de una sesión SHALL NOT poderse deducir el de otra ni la cantidad de sesiones. (HU-52, CA-52.2)

#### Scenario: Muchas sesiones abiertas seguidas (caso alternativo)
- **WHEN** se abren 1000 sesiones, cada una sobre un ítem distinto
- **THEN** los 1000 identificadores públicos son distintos, todos son UUID versión 4, y ordenarlos no da el orden de apertura

### Requirement: Aperturas simultáneas de sesiones
Abrir sesiones al mismo tiempo, sobre ítems distintos, SHALL generar identificadores públicos únicos, sin condiciones de carrera (`go test -race` en la CI). (HU-52, CA-52.7)

#### Scenario: Abrir sesiones en paralelo
- **WHEN** 50 aperturas de sesiones corren al mismo tiempo, cada una sobre su ítem
- **THEN** se abren las 50, con 50 identificadores públicos distintos

### Requirement: Operar una sesión por su UUID
Las rutas de una sesión (`POST /api/poker-sessions/{id}/votes`, `/reveal`, `/estimate` y `/rounds`) SHALL identificarla por su UUID, con mayúsculas aceptadas. Los `GET` de esas rutas siguen respondiendo 405. La existencia de la sesión se valida igual que hoy (`ErrSessionNotFound`), porque cada operación necesita la sesión guardada. (HU-52, CA-52.3)

#### Scenario: Votar en una sesión por su UUID (caso normal)
- **WHEN** se pide `POST /api/poker-sessions/{UUID de la sesión}/votes` con `{"integrante":"Ana","puntos":5}`
- **THEN** responde 200 con el voto registrado, y la sesión sigue con el mismo UUID en `id`

#### Scenario: UUID de la sesión en mayúsculas (caso alternativo)
- **WHEN** se pide `POST /api/poker-sessions/{su UUID en mayúsculas}/reveal` con votos registrados
- **THEN** responde 200 y el `id` de la respuesta está en minúsculas

### Requirement: Ninguna respuesta expone el ID entero de una sesión
Toda respuesta de la API que identifica una sesión SHALL hacerlo con su UUID en `id`, y el ítem de la sesión con su UUID en `item_id`. (HU-52, CA-52.3)

#### Scenario: Revelar la sesión (caso alternativo)
- **WHEN** se revela una sesión con dos votos
- **THEN** la respuesta tiene el UUID de la sesión en `id` y el del ítem en `item_id`, y ningún número

### Requirement: Identificador de sesión con formato inválido
Si el `{id}` de una ruta de sesión no tiene la forma canónica de un UUID, la API SHALL responder 400 con `{"error":"el id de la sesión debe ser un UUID"}`, sin consultar la sesión. Se rechazan los números, las llaves, el prefijo `urn:uuid:` y la forma sin guiones. (HU-52, CA-52.4)

#### Scenario: Un ID numérico de sesión (caso de error)
- **WHEN** se pide `POST /api/poker-sessions/1/reveal`
- **THEN** responde 400 con `{"error":"el id de la sesión debe ser un UUID"}`, y no 404

#### Scenario: El ID de la sesión se valida antes que el cuerpo (caso límite)
- **WHEN** se pide `POST /api/poker-sessions/abc/votes` con un cuerpo que no es JSON
- **THEN** responde 400 con el mensaje del ID de la sesión

### Requirement: Sesión inexistente
Si el `{id}` tiene la forma de un UUID pero no corresponde a ninguna sesión, la API SHALL responder 404 con `{"error":"no existe la sesión de Planning Poker"}`. Las validaciones de contenido conservan su orden: siguen antes que la existencia. (HU-52, CA-52.4)

#### Scenario: UUID de sesión que no existe (caso de error)
- **WHEN** se pide `POST /api/poker-sessions/{UUID versión 4 que nunca se asignó}/reveal`
- **THEN** responde 404 con `{"error":"no existe la sesión de Planning Poker"}`

#### Scenario: Se conserva el orden de errores al votar (caso límite)
- **WHEN** se pide `POST /api/poker-sessions/{UUID inexistente}/votes` con el integrante vacío
- **THEN** responde 400 con `{"error":"el votante no puede estar vacío"}`, igual que hoy con un ID entero inexistente (HU-22)

### Requirement: Sesiones en memoria y en PostgreSQL
El identificador público de la sesión SHALL guardarse y buscarse igual en memoria y en PostgreSQL, en una columna UUID única de la tabla de sesiones, y SHALL conservarse al actualizar la sesión (votos, revelación, estimación y rondas). La relación con su ítem sigue con el entero. (HU-52, CA-52.5)

#### Scenario: Guardar y buscar una sesión por su UUID
- **WHEN** se guarda una sesión y después se la busca por su identificador público, en memoria y en PostgreSQL
- **THEN** los dos repositorios devuelven la misma sesión, y un UUID que no existe da `ErrSessionNotFound`

#### Scenario: Dos sesiones con el mismo UUID en PostgreSQL (caso de error)
- **WHEN** se intenta guardar en PostgreSQL una sesión con un identificador público que ya usa otra
- **THEN** la base lo rechaza y el guardado devuelve un error, sin crear la sesión

### Requirement: La interfaz web usa el UUID de la sesión
En la interfaz web, Planning Poker no tiene un campo para escribir el ID de la sesión: la sesión queda en el formulario al abrirla. SHALL seguir funcionando con el UUID abrir, votar, revelar, repetir la ronda y registrar la estimación, y los mensajes SHALL mostrar el UUID de la sesión. (HU-52, CA-52.6)

#### Scenario: Una sesión completa desde la interfaz (caso normal)
- **WHEN** el usuario abre una sesión sobre un ítem, vota, revela y registra la estimación
- **THEN** cada paso funciona, y los mensajes muestran el UUID de la sesión
