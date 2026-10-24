# HU-21 · Abrir una sesión de Planning Poker

> Issue del repositorio: #39

Aprobada por: Peñalbé Hernán (Agile Enabler), 24/10/2026, con los ajustes de CA-21.2 y del texto del Issue.

## Historia de usuario

Como Agile Enabler, quiero abrir una sesión de Planning Poker para un ítem, para que el equipo lo estime en conjunto.

## Objetivo

Abrir una sesión de Planning Poker asociada a un ítem existente del Product Backlog, con un identificador único, el estado inicial `ABIERTA`, la ronda 1 y sin votos, para que HU-22 (votar de forma oculta), HU-23 (revelar las estimaciones), HU-24 (repetir ronda) y HU-25 (registrar la estimación acordada) tengan sobre qué trabajar.

Esta historia crea el paquete `internal/poker`: el tipo `Session`, la interfaz `Repository`, sus implementaciones en memoria y en PostgreSQL y el `Manager` (ver "Restricciones"). La sesión pertenece a un ítem del backlog que tiene que existir (HU-04 y HU-05), y las cartas son la misma escala Fibonacci de HU-20.

## Alcance

- **Incluye:** abrir una sesión para un ítem existente a través del `Manager` del paquete `internal/poker`, validando que el ítem exista, que no esté `TERMINADO` y que no haya otra sesión `ABIERTA` para el mismo ítem; exponer las cartas de HU-20; guardar la sesión en memoria o en PostgreSQL (TEC-05); exponer la operación en `POST /api/items/{id}/poker-sessions` y en la interfaz web (tema 5 del Planning del Sprint 3).
- **No incluye:** votar (HU-22), revelar los votos (HU-23), repetir ronda (HU-24), registrar la estimación acordada ni cerrar la sesión (HU-25), listar o consultar sesiones, borrar sesiones, ni asignar Story Points al ítem (HU-20, que HU-25 reutiliza).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ItemID | entero | Sí | ID del ítem del backlog para el que se abre la sesión |

El ID de la sesión, el estado, la ronda y los votos no son entradas: los asigna el sistema.

Firma: `OpenSession(itemID int) (Session, error)`, en `internal/poker/create.go`.

### Salidas esperadas

- **Si la apertura es válida:** la sesión creada, como **copia por valor** (con su propia copia de los votos), con:
  - **ID:** entero correlativo global asignado por el repositorio (el primero es 1).
  - **ItemID:** el recibido.
  - **Estado:** `ABIERTA`.
  - **Ronda:** 1.
  - **Votos:** lista vacía, no `nil`.
- **Si la apertura no es válida:** el error que indica el motivo (ver "Condiciones de error") y no se crea ninguna sesión.

### Cartas (escala compartida con HU-20)

Las cartas de la sesión son la escala Fibonacci acordada en HU-20 (1, 2, 3, 5, 8, 13 y 21). No se repiten los valores en el paquete `poker`:

- `Cards() []int`: función exportada que devuelve una **lista nueva** con `backlog.StoryPointScale()`, en orden ascendente. HU-22 la usa para validar los votos y la interfaz web para mostrar las cartas.
- `backlog.StoryPointScale()` sigue siendo la única definición de los valores (HU-20). Modificar la lista que devuelve `Cards()` no altera la escala.

Esto hace explícita la dependencia **HU-21 → HU-20**, registrada en la SDD de HU-20: HU-20 tiene que estar integrada en `main` antes del commit GREEN de HU-21.

### HTTP (TEC-04)

`POST /api/items/{id}/poker-sessions`, con el mismo patrón de rutas bajo `/api/…` de TEC-04.

- **Cuerpo:** no se usa; la sesión se identifica por el `{id}` del ítem.
- **Respuesta exitosa:** `201 Created` con la sesión:

  ```json
  {
    "id": 1,
    "item_id": 5,
    "estado": "ABIERTA",
    "ronda": 1,
    "votos": [],
    "cartas": [1, 2, 3, 5, 8, 13, 21]
  }
  ```

- **Errores:**

  | Caso | Código | Mensaje |
  |---|---|---|
  | `{id}` que no es un número entero | 400 | El de `errInvalidID` |
  | Ítem inexistente | 404 | El de `backlog.ErrItemNotFound` |
  | Ítem `TERMINADO` | 409 | El de `poker.ErrItemFinished` |
  | Ya hay una sesión `ABIERTA` para el ítem | 409 | El de `poker.ErrSessionAlreadyOpen` |
  | Falla del repositorio | 500 | El de `errInternal`, sin detalle |
  | Método distinto de `POST` | 405 | Lo responde `net/http` por la ruta |

  El orden de comprobación es: ID de la ruta, existencia del ítem, estado `TERMINADO`, sesión abierta y guardado. Se informa el primer error encontrado.

### Interfaz web (TEC-04)

En la interfaz web (`internal/web`):

- En las acciones de cada fila del backlog, un botón **"Abrir Poker"** que llama al endpoint del ítem.
- Una sección **"Planning Poker"** debajo del backlog que muestra la última sesión abierta: su ID, el ítem, el estado, la ronda y las cartas (1, 2, 3, 5, 8, 13 y 21). Mientras no se haya abierto ninguna, muestra un texto de ayuda.
- El resultado se informa con el mismo mecanismo que las demás acciones: un mensaje de éxito ("Sesión N abierta para el ítem M.") o el mensaje de error de la API.
- Tests en `internal/httpapi/web_test.go`, como los existentes: la página servida contiene el botón y la sección, y `app.js` usa la ruta nueva.

## Reglas de negocio

1. **RN-21.1 Ítem existente.** La sesión pertenece a un ítem del backlog que tiene que existir (ver "Decisión sobre entidades referenciadas"). Si no existe, se devuelve `ErrItemNotFound` de `internal/backlog` y no se crea la sesión. Incluye IDs 0, negativos y abrir sin ítems creados.
2. **RN-21.2 Una sola sesión abierta por ítem.** No puede haber más de una sesión `ABIERTA` para el mismo ítem (CA-21.2). Si ya hay una, se devuelve `ErrSessionAlreadyOpen` y no se crea otra. La regla es por ítem: dos ítems distintos pueden tener cada uno su sesión abierta.
3. **RN-21.3 Estado inicial.** Toda sesión nueva se crea `ABIERTA`. El estado `ABIERTA` significa que la sesión está en curso y acepta votos (HU-22). Los demás estados los agregan HU-23 (revelar) y HU-25 (cerrar); esta historia no los define ni los usa.
4. **RN-21.4 Ronda inicial.** Toda sesión nueva empieza en la ronda 1. La repetición de rondas es alcance de HU-24.
5. **RN-21.5 Sin votos.** La sesión nueva empieza sin votos: la lista de votos está vacía (no `nil`). Los votos y su validación son alcance de HU-22; esta historia solo deja preparado el campo.
6. **RN-21.6 Ítem terminado.** No se abre una sesión para un ítem `TERMINADO`: se devuelve `ErrItemFinished` y no se crea la sesión. Es coherente con RN-20.5, que impide cambiar directamente los Story Points de un ítem terminado; abrir una sesión que después no podría registrar su resultado sería trabajo perdido. Los ítems `PENDIENTE` y `EN_CURSO` se pueden estimar, estén o no asignados a un sprint y tengan o no Story Points (se permite reestimar, como RN-20.3).
7. **RN-21.7 Identificador único.** Cada sesión recibe un ID entero correlativo global que empieza en 1, asignado por el repositorio al guardar. Es global y no por ítem, con el mismo criterio que RN-04.6 y RN-10.6.
8. **RN-21.8 Copias por valor.** La apertura devuelve una copia de la sesión. `Session` tiene un slice de votos, así que el repositorio y el `Manager` siempre devuelven listas nuevas: modificar la sesión devuelta o sus votos no altera la sesión guardada. Las funciones de consulta y de escritura nunca devuelven punteros al estado interno.
9. **RN-21.9 Cartas de la escala.** Las cartas son las de HU-20, sin repetir valores: `Cards()` devuelve una lista nueva con `backlog.StoryPointScale()` (ver "Cartas"). Modificar la lista devuelta no altera la escala ni las cartas siguientes.
10. **RN-21.10 Solo se abre la sesión.** Esta historia no modifica el ítem, no escribe Story Points y no altera ningún otro dato. Tampoco valida el proyecto ni el sprint del ítem (ver "Decisión sobre entidades referenciadas").

## Decisión sobre entidades referenciadas

HU-21 referencia una entidad por ID: el ítem del backlog al que pertenece la sesión. **Sí se valida su existencia**, por estos motivos:

1. Una sesión sin ítem quedaría huérfana: no tendría sobre qué estimar y no aparecería en ninguna consulta del backlog.
2. Ya existe la búsqueda por ID (HU-05) y HU-20 la reutiliza: el `Manager` de poker llama a `backlog.Manager.GetItem`, que devuelve `ErrItemNotFound`. No se duplica ni la regla ni el error.
3. La misma lectura trae el estado del ítem, que hace falta para RN-21.6 (ítem `TERMINADO`).

**No se valida el proyecto.** No es una entrada: el ítem existente ya pertenece a un proyecto validado al crearlo (RN-04.5) y no hay ninguna operación que borre proyectos.

**No se valida ni se consulta el sprint del ítem.** La sesión se puede abrir para cualquier ítem no terminado, esté o no asignado a un sprint. Consultar el sprint obligaría a que `poker` dependa de `sprint`, que ya depende de `backlog`, y ninguna regla de esta historia lo pide.

## Restricciones

- **Organización del código.** Paquete `internal/poker`, siguiendo el patrón de HU-04 y HU-10:
  - `doc.go`: documentación del paquete.
  - `session.go`: los tipos `Session`, `Status` y `Vote`, las constantes de estado y los errores.
  - `repository.go`: la interfaz `Repository`.
  - `memory_repository.go`: la implementación en memoria.
  - `postgres_repository.go`: la implementación sobre PostgreSQL (TEC-05).
  - `manager.go`: el `Manager` y `NewManager`.
  - `create.go`: la apertura de la sesión (HU-21), `Cards()` y `ErrSessionAlreadyOpen` y `ErrItemFinished`.

  HU-22 a HU-25 agregan sus propios archivos. No se toca el código de producción de `internal/backlog`, `internal/project` ni `internal/sprint`.
- **Constructor.** `NewManager(repo Repository, items *backlog.Manager) *Manager`. El `Manager` usa `items.GetItem` para validar la existencia y el estado del ítem.
- **Interfaz de repositorio.** HU-21 define solo las operaciones que necesita:
  - `Save(session Session) (Session, error)`: asigna el ID y devuelve una copia.
  - `OpenByItem(itemID int) (Session, bool, error)`: devuelve una copia de la sesión `ABIERTA` del ítem y si existe.

  HU-22 agregará `GetByID` y `Update`. La lógica de negocio accede a las sesiones solo a través de esta interfaz.
- **Persistencia (TEC-05).** La sesión se guarda en memoria o en PostgreSQL, como el resto de las entidades:
  - Tabla nueva `poker_sessions` en `db/init/01-schema.sql`, con `id` identity, `item_id` con clave foránea a `backlog_items`, `status`, `round_number` y `votes` como `JSONB` (lista vacía por defecto). El esquema solo guarda los datos y la relación; las reglas de negocio las aplica el `Manager`, como el resto del esquema.
  - `PostgresRepository` cumple el mismo contrato que `MemoryRepository`; los tests de integración se omiten sin `TEST_DATABASE_URL`.
  - `internal/postgrestest/postgrestest.go` agrega `poker_sessions` a las tablas que vacía `Reset`.
  - `cmd/server/repositories.go` y `cmd/server/main.go` arman el repositorio y el `Manager` en los dos modos de almacenamiento.
  - Los votos se codifican y decodifican con `encoding/json` sobre la columna `JSONB`, para que el repositorio no dependa de un formato binario.
- **Concurrencia.** El repositorio en memoria no es seguro para uso concurrente. Abrir una sesión hace dos pasos seguidos sobre el repositorio: consultar la sesión abierta del ítem y guardar. El `Manager` protege ambos con su `sync.Mutex`, para que dos aperturas simultáneas del mismo ítem no dejen dos sesiones abiertas (CA-21.6). La validación del ítem se hace antes de tomar el mutex de poker, para no tomar dos mutex a la vez, igual que HU-04 y HU-10.
- **Límite conocido.** La existencia y el estado del ítem se leen una vez, antes del mutex de poker. Si un cambio de estado simultáneo (HU-08) marca el ítem `TERMINADO` entre esa lectura y el guardado, la sesión se abre igual; HU-25 fallará al intentar registrar la estimación. Cerrar esa ventana exigiría una transacción que abarque dos `Manager`, que ninguna historia pide.
- **Copias por valor.** Se devuelve `Session` por valor con una copia nueva de `Votes`; nunca punteros al estado interno.
- **Errores identificables.** `ErrSessionAlreadyOpen` ("ya hay una sesión de Planning Poker abierta para ese ítem") y `ErrItemFinished` ("no se puede abrir una sesión de Planning Poker para un ítem terminado") son nuevos, exportados y comparables con `errors.Is`. Se reutiliza `backlog.ErrItemNotFound`, propagado tal cual. En `httpapi`, los dos errores nuevos se traducen a 409.
- **HTTP.** Handler en `internal/httpapi/poker.go`, ruta en `server.go`, `poker.Manager` en el `server` y en `NewHandler`; se actualizan el armado en `cmd/server/main.go` y los helpers de los tests que llaman a `NewHandler`. Tests con un caso exitoso y uno por código de error (criterio de TEC-04).
- **Coordinación con HU-26.** HU-26 ya está integrada en `main` y cambia `NewHandler` (agrega `efforts`), `postgrestest`, el esquema y `cmd/server`. Antes del commit GREEN, hacer `git fetch` y `git merge origin/main` en la rama y sumar poker sobre esos cambios, resolviendo los conflictos de `NewHandler`, el esquema y `postgrestest`.
- **Tests.** Unitarios en `internal/poker/create_test.go`; contrato de repositorio en `repository_contract_test.go`; PostgreSQL en `postgres_repository_test.go` (se omiten sin `TEST_DATABASE_URL`); pasos de godog en `create_features_test.go`, con el runner `TestFeatures_HU21AbrirSesionPlanningPoker`. Cada escenario lleva su criterio en el nombre, por ejemplo "Escenario: … (CA-21.2)". Los tests son internos al paquete (`package poker`), como en HU-04 y HU-10.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Primera sesión del sistema | Recibe el ID 1, en estado `ABIERTA`, ronda 1 y sin votos |
| Sesiones para dos ítems distintos | Los IDs son correlativos y globales (1, 2, 3…); cada ítem puede tener su sesión abierta |
| Ítem `PENDIENTE` | Se abre la sesión |
| Ítem `EN_CURSO` | Se abre la sesión |
| Ítem `PENDIENTE` o `EN_CURSO` asignado a un sprint (abierto o cerrado) | Se abre la sesión; el sprint no se consulta |
| Ítem con Story Points ya asignados | Se abre la sesión (se permite reestimar, como RN-20.3) |
| Ítem `TERMINADO`, con 0 o con Story Points | `ErrItemFinished`; no se crea la sesión |
| Ya existe una sesión `ABIERTA` para el ítem | `ErrSessionAlreadyOpen`; no se crea otra |
| Dos ítems distintos, cada uno con una sesión abierta | Las dos se crean |
| ID de ítem inexistente (999), 0 o negativo, o sin ítems creados | `ErrItemNotFound` |
| Ítem `TERMINADO` y con sesión abierta a la vez | `ErrItemFinished` (se informa primero el estado) |
| Modificar la sesión devuelta o sus votos | La sesión guardada no cambia; volver a abrirla no es posible mientras siga `ABIERTA` |
| Modificar la lista que devuelve `Cards()` | Las cartas y la escala no cambian; la siguiente llamada devuelve 1, 2, 3, 5, 8, 13, 21 |
| Varias aperturas simultáneas del mismo ítem | Exactamente una tiene éxito y las demás devuelven `ErrSessionAlreadyOpen`; queda una sola sesión abierta; `go test -race` sin carreras |
| Varias aperturas simultáneas de ítems distintos | Todas se crean, con IDs distintos |
| El repositorio falla al consultar o guardar | Se devuelve ese error y no se crea ninguna sesión |
| HTTP: `{id}` no numérico | 400 con el mensaje de `errInvalidID` |
| HTTP: ítem inexistente | 404 con el mensaje de `ErrItemNotFound` |
| HTTP: ítem terminado o sesión ya abierta | 409 con el mensaje correspondiente |
| HTTP: falla del repositorio | 500 con `errInternal`, sin detalle |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| No existe un ítem con ese ID (RN-21.1) | `backlog.ErrItemNotFound` | No se crea la sesión |
| El ítem está `TERMINADO` (RN-21.6) | `poker.ErrItemFinished` | No se crea la sesión |
| Ya hay una sesión `ABIERTA` para el ítem (RN-21.2) | `poker.ErrSessionAlreadyOpen` | No se crea la sesión |
| El repositorio falla al consultar o guardar | El error que devuelve el repositorio | No se crea la sesión |

**Orden de validación:** existencia del ítem, estado `TERMINADO`, sesión abierta y guardado. Se informa el primer error encontrado.

## Criterios de aceptación

- **CA-21.1** Se abre una sesión asociada a un ítem existente del backlog, con un ID único correlativo (el primero es 1), estado `ABIERTA`, ronda 1 y sin votos.
- **CA-21.2** No puede haber dos sesiones `ABIERTA` para el mismo ítem: si ya existe una, se devuelve `ErrSessionAlreadyOpen` y no se crea otra. Dos ítems distintos pueden tener cada uno su sesión abierta.
- **CA-21.3** La sesión nueva empieza sin votos (lista vacía, no `nil`); modificar los votos de la sesión devuelta no altera la sesión guardada.
- **CA-21.4** Abrir una sesión para un ítem inexistente (incluidos ID 0, negativo y sin ítems) devuelve `ErrItemNotFound`; abrirla para un ítem `TERMINADO` devuelve `ErrItemFinished`. En los dos casos no se crea la sesión.
- **CA-21.5** Las cartas son la escala Fibonacci de HU-20 (1, 2, 3, 5, 8, 13 y 21), expuestas con `Cards()` como una lista nueva; modificar la lista devuelta no altera la escala ni las cartas siguientes.
- **CA-21.6** De varias aperturas simultáneas del mismo ítem, exactamente una tiene éxito y las demás devuelven `ErrSessionAlreadyOpen`; queda una sola sesión abierta y `go test -race` no reporta condiciones de carrera. Si el repositorio falla, se devuelve ese error y no se crea ninguna sesión.
- **CA-21.7** La sesión se puede abrir desde `POST /api/items/{id}/poker-sessions`, con los códigos de la tabla HTTP, y desde la interfaz web, que muestra la sesión abierta y sus cartas.
- **CA-21.8** La sesión se guarda en memoria y en PostgreSQL con el mismo contrato (TEC-05); las pruebas de integración se omiten sin `TEST_DATABASE_URL`.

## Comparación con el acta de Planning

El acta del Sprint 3 (`docs/actas/sprint-3/planning.md`) describe HU-21 como "Abrir una sesión de Planning Poker" (Issue #39, 5 Story Points, responsable Buttini Cristobal), con la nota "Paquete nuevo, como HU-10: entidad, repositorio y `Manager`; la sesión pertenece a un ítem existente", la dependencia HU-04 y HU-05, y como bloqueante de HU-22, HU-23, HU-24 y HU-25. **El alcance coincide:** el paquete nuevo tiene entidad, repositorio y `Manager` (patrón HU-10), y la sesión se valida contra un ítem existente (RN-21.1).

Además de lo que describe el acta:

- **Endpoint y pantalla (tema 5 del Planning):** la historia suma `POST /api/items/{id}/poker-sessions` y su sección en la interfaz web, como las demás historias nuevas del Sprint.
- **Dependencia HU-21 → HU-20:** las cartas son la escala de HU-20 y se exponen con `Cards()`; la dependencia ya está registrada en la SDD aprobada de HU-20 y HU-20 está integrada en `main`. El acta no la lista, pero tampoco la contradice.
- **Persistencia PostgreSQL (TEC-05):** la tabla `poker_sessions` y su repositorio siguen la regla del proyecto de que toda entidad nueva se persiste en memoria y en PostgreSQL. El acta no lo menciona porque describe el paquete a nivel funcional.
- **Ítem `TERMINADO` (RN-21.6):** es una restricción nueva que el Issue no menciona, confirmada con el responsable de la historia por coherencia con HU-20. No agrega funcionalidad.

**No hay diferencias de alcance con el acta.**

## Comparación con los criterios del Issue #39

| CA en el Issue #39 | En esta SDD | Ajuste |
|---|---|---|
| CA-21.1 Se abre una sesión asociada a un ítem del backlog | CA-21.1 (precisado) | El ítem tiene que existir; la sesión queda `ABIERTA`, ronda 1, con ID correlativo |
| CA-21.2 No puede haber dos sesiones abiertas para el mismo ítem | CA-21.2 (precisado) | Se identifica el error |
| CA-21.3 La sesión nueva empieza sin votos | CA-21.3 (precisado) | Lista vacía, no `nil`, con copia por valor |
| — | CA-21.4 (nuevo) | Ítem inexistente e ítem `TERMINADO` |
| — | CA-21.5 (nuevo) | Cartas de HU-20 con `Cards()` |
| — | CA-21.6 (nuevo) | Concurrencia y error del repositorio |
| — | CA-21.7 (nuevo) | API e interfaz web (tema 5 del Planning) |
| — | CA-21.8 (nuevo) | Persistencia en memoria y PostgreSQL (TEC-05) |

**Texto definitivo para la sección "Criterios de aceptación" del Issue #39** (en lenguaje de negocio, mejora 8 del Planning; se reemplaza cuando se apruebe la SDD y con autorización del Agile Enabler):

```markdown
## Criterios de aceptación

- [ ] CA-21.1 Se abre una sesión de Planning Poker asociada a un ítem existente del backlog, con ID único, estado ABIERTA, ronda 1 y sin votos.
- [ ] CA-21.2 No puede haber dos sesiones abiertas para el mismo ítem.
- [ ] CA-21.3 La sesión nueva empieza sin votos y modificar el resultado devuelto no altera la sesión guardada.
- [ ] CA-21.4 Abrir una sesión para un ítem inexistente o terminado devuelve error y no crea la sesión.
- [ ] CA-21.5 Las cartas son la escala Fibonacci de HU-20: 1, 2, 3, 5, 8, 13 y 21.
- [ ] CA-21.6 Las aperturas simultáneas del mismo ítem dejan una sola sesión abierta y un error al guardar no crea nada.
- [ ] CA-21.7 La sesión se abre desde la API y desde la interfaz web, que muestra la sesión y sus cartas.
- [ ] CA-21.8 La sesión se guarda en la base de datos igual que el resto de los datos de la aplicación.
```

## Decisiones tomadas en la redacción de la SDD

1. **Paquete nuevo `internal/poker`,** con entidad, repositorio en memoria y PostgreSQL, y `Manager`, como pide el acta y siguiendo el patrón de HU-04 y HU-10.
2. **Una sola sesión `ABIERTA` por ítem,** con `OpenByItem` y la consulta y el guardado bajo el mutex del `Manager` (RN-21.2 y CA-21.6).
3. **Ítem `TERMINADO`: se rechaza** (`ErrItemFinished`, 409), confirmado con el responsable de la historia, por coherencia con RN-20.5 de HU-20. *Alternativa descartada:* permitir abrir y dejar que HU-25 falle al guardar.
4. **Votante por nombre libre (`Vote.Voter string`).** La épica E4 (integrantes) no está implementada en esta entrega, así que la sesión no valida al votante contra una entidad; el tipo del votante lo confirma la SDD de HU-22, que definirá las reglas del voto. *Alternativa descartada:* un ID de integrante entero, que no tendría contra qué validarse.
5. **Cartas con `Cards()`**, que devuelve una lista nueva de `backlog.StoryPointScale()`. Evita repetir la escala y hace explícita la dependencia HU-21 → HU-20 aprobada en esa SDD.
6. **Estado inicial `ABIERTA` y ronda 1;** los demás estados y la repetición de rondas quedan para HU-23, HU-24 y HU-25. No se define funcionalidad que esas historias no pidieron.
7. **Endpoint `POST /api/items/{id}/poker-sessions`** (no lleva cuerpo) y una sección "Planning Poker" en la interfaz web, por el tema 5 del Planning del Sprint 3.
8. **Persistencia en PostgreSQL** con la tabla `poker_sessions` y `votes` como `JSONB`, siguiendo TEC-05; los tests de integración se omiten sin `TEST_DATABASE_URL`, como los del resto de los paquetes.
