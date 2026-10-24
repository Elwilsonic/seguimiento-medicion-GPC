# HU-26 · Registrar esfuerzo

> Issue del repositorio: #44

Aprobada por: Peñalbé Hernán (Agile Enabler), 24/10/2026, con los ajustes de numeración de CA y step del campo Horas.

## Historia de usuario

Como Product Builder, quiero registrar las horas que dediqué a una historia de un sprint, para medir el esfuerzo real.

## Objetivo

Registrar un esfuerzo real (integrante, fecha, actividad, ítem del backlog, sprint y horas trabajadas) sobre una historia asignada a un sprint `ABIERTO`, para que las métricas de esfuerzo tengan el dato.

Es el dato de entrada de **M5 (horas reales)** y, junto con **M4 (horas estimadas)**, de **M6 (desviación en horas)** y **M7 (desviación porcentual)** de `docs/metricas.md` (TEC-03). HU-27 consultará los registros y HU-33 (o HU-45 y HU-46 si se divide) calculará M5, M6 y M7 en el Sprint 4.

La historia crea el paquete `internal/effort` con la entidad `Effort`, la interfaz `Repository`, sus implementaciones en memoria y PostgreSQL, el `Manager` y la operación `Register`. Para que el registro sea atómico con el cierre del sprint, agrega `sprint.Manager.WithOpenSprint`, que valida el sprint `ABIERTO` y ejecuta el guardado con el mutex del sprint tomado, como `CompleteItem` en HU-12. Suma el endpoint `POST /api/sprints/{id}/items/{itemId}/esfuerzos` y un formulario en la interfaz web (tema 5 del Planning del Sprint 3). No modifica el ítem: el esfuerzo real se guarda en la tabla nueva `efforts`.

## Alcance

- **Incluye:** registrar un esfuerzo con integrante, fecha, actividad, ítem, sprint y horas; asignarle un ID correlativo global; validar los datos del esfuerzo, la existencia y el estado `ABIERTO` del sprint, y la existencia y pertenencia del ítem al sprint; persistir en memoria y en PostgreSQL; devolver una copia por valor; el endpoint HTTP y el formulario de la interfaz web; tests unitarios, BDD, de contrato, de PostgreSQL, de API y de interfaz.
- **No incluye:** consultar el esfuerzo (HU-27), corregirlo o eliminarlo (HU-28), registrar integrantes como entidad (E4, HU-15 a HU-17), imputar horas a tareas que no existen, modificar las horas estimadas o el estado del ítem, calcular métricas (E9), dashboard, reportes ni listar los esfuerzos registrados en la interfaz (eso es HU-27).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| SprintID | entero | Sí | Sprint al que pertenece la historia |
| ItemID | entero | Sí | Historia del backlog sobre la que se trabajó |
| Integrante | texto | Sí | Nombre del integrante que registra el esfuerzo |
| Fecha | fecha (`time.Time`) | Sí | Día en que se realizó la actividad |
| Actividad | texto | Sí | Actividad realizada |
| Horas | decimal (`float64`) | Sí | Horas trabajadas; debe ser un número positivo y finito |

El ID no es una entrada: lo asigna el repositorio. El proyecto tampoco se recibe: se deriva del sprint y del ítem.

Firma: `Register(sprintID, itemID int, member string, date time.Time, activity string, hours float64) (Effort, error)` en el `Manager` de `internal/effort`.

### Salidas esperadas

- **Si el registro es válido:** el esfuerzo guardado, como copia por valor, con:
  - **ID:** entero correlativo global asignado por el repositorio (el primero es 1).
  - **Integrante y actividad:** los recibidos, sin los espacios de los extremos.
  - **Fecha:** la recibida (en PostgreSQL se devuelve normalizada a UTC, como las fechas de los sprints).
  - **ItemID y SprintID:** los recibidos.
  - **Horas:** las recibidas.
- **Si el registro no es válido:** un error que indica el motivo (ver "Condiciones de error"), y no se guarda ningún esfuerzo.

### Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Integrante vacío o con solo espacios | `ErrEmptyMember` de `internal/effort` ("el integrante es obligatorio") | No se guarda nada |
| Actividad vacía o con solo espacios | `ErrEmptyActivity` de `internal/effort` ("la actividad es obligatoria") | No se guarda nada |
| Fecha ausente (valor zero de `time.Time`) | `ErrMissingDate` de `internal/effort` ("la fecha del esfuerzo es obligatoria") | No se guarda nada |
| Horas 0, negativas, `NaN`, `+Inf` o `-Inf` | `ErrInvalidHours` de `internal/effort` ("las horas trabajadas deben ser un número positivo") | No se guarda nada |
| No existe un sprint con ese SprintID (incluye 0 y negativos) | `ErrSprintNotFound` de `internal/sprint`, propagado tal cual | No se guarda nada |
| El sprint existe pero está `CERRADO` | `ErrSprintClosed` de `internal/sprint`, propagado tal cual | No se guarda nada |
| No existe un ítem con ese ItemID (incluye 0 y negativos) | `ErrItemNotFound` de `internal/backlog`, propagado tal cual | No se guarda nada |
| El ítem no está asignado a ese sprint o está asignado a otro | `ErrItemNotAssignedToSprint` de `internal/sprint`, reutilizado | No se guarda nada |
| El repositorio falla al guardar | El error que devuelve el repositorio | No se guarda nada |

**Orden de validación:** integrante, actividad, fecha, horas, sprint (existencia y estado) e ítem (existencia y pertenencia), y guardado. Se informa el primer error encontrado. Los datos del esfuerzo se validan antes de consultar el sprint y el ítem, como en HU-04. La validación del sprint y el ítem y el guardado ocurren dentro de `WithOpenSprint`, con el mutex del sprint tomado.

### HTTP (TEC-04, tema 5 del Planning del Sprint 3)

`POST /api/sprints/{id}/items/{itemId}/esfuerzos`, con el mismo patrón que `POST /api/sprints/{id}/items/{itemId}/completar` (HU-12).

- **Cuerpo:**
  ```json
  {"integrante": "García Santiago", "fecha": "2026-10-24", "actividad": "Implementar el registro de esfuerzo", "horas": 2.5}
  ```
- **Respuesta exitosa:** `201 Created` con el esfuerzo:
  ```json
  {"id": 1, "sprint_id": 1, "item_id": 3, "integrante": "García Santiago", "fecha": "2026-10-24", "actividad": "Implementar el registro de esfuerzo", "horas": 2.5}
  ```
- **Errores:**

| Caso | Código | Mensaje |
|---|---|---|
| Integrante o actividad vacíos, fecha ausente o inválida, horas no positivas, no finitas o no numéricas | 400 | El error correspondiente (`ErrEmptyMember`, `ErrEmptyActivity`, `ErrMissingDate`, `ErrInvalidHours`, `errInvalidDate`) |
| Cuerpo que no es JSON, `horas` que no es número (por ejemplo `true`), o texto después del JSON | 400 | El de `errInvalidJSON` |
| `{id}` o `{itemId}` que no son números enteros | 400 | El de `errInvalidID` |
| Sprint inexistente | 404 | El de `ErrSprintNotFound` |
| Ítem inexistente | 404 | El de `ErrItemNotFound` |
| Sprint `CERRADO` | 409 | El de `ErrSprintClosed` |
| Ítem sin asignar o asignado a otro sprint | 409 | El de `ErrItemNotAssignedToSprint` |
| Cuerpo de más de 1 MiB | 413 | El de `errBodyTooLarge` |
| Falla del repositorio | 500 | El de `errInternal`, sin detalle |

`horas` se lee como `json.Number` y se convierte con `strconv.ParseFloat`: lo que no sea un número representable responde con `ErrInvalidHours`, no con el error de JSON; con `json.Number`, el texto `"2.5"` se acepta como el número 2.5 (comportamiento de `encoding/json`, igual que en HU-20). `fecha` se lee con el `parseDate` existente (`AAAA-MM-DD`); una fecha vacía llega como zero y la rechaza el Manager con `ErrMissingDate`. El orden de los errores del handler es: IDs de la ruta, JSON, fecha, horas, Manager.

### Interfaz web (TEC-04, tema 5)

En `internal/web`, dentro de la sección **Sprints**, un formulario **"Registrar esfuerzo"**:

- Campos: **Ítem (ID)**, **Integrante**, **Fecha** (`type="date"`), **Actividad** y **Horas** (`type="number"`, `step="any"`, para aceptar cualquier decimal positivo igual que la API).
- Usa el **sprint actual** del contexto de la página y llama al endpoint.
- Al terminar, recarga el backlog (el esfuerzo no cambia el ítem, pero mantiene el estado de la pantalla) y muestra "Esfuerzo N registrado para el ítem X." o el mensaje de error de la API, con el mismo mecanismo de `accion` y `mostrar` que los demás formularios.
- No agrega un listado de esfuerzos: eso es HU-27.

## Reglas de negocio

1. **RN-26.1 Integrante como texto libre.** El integrante se guarda como texto obligatorio, sin espacios en los extremos. No se modela una entidad de integrante ni se valida contra E4, porque ese paquete no está comprometido en el Sprint 3 (el acta lo deja solo para capacidad sobrante). Cuando E4 esté integrada, una historia posterior puede migrar el campo a un ID; esta SDD no lo asume.
2. **RN-26.2 Actividad obligatoria.** La actividad no puede estar vacía ni tener solo espacios, y se guarda recortada. No se restringe su contenido ni su longitud.
3. **RN-26.3 Fecha obligatoria y sin restricción de futuro.** El valor zero de `time.Time` se trata como fecha ausente. No se rechazan fechas futuras ni se valida que la fecha caiga dentro del sprint: el Issue no pide esa regla y agregarla complicaría el registro de correcciones.
4. **RN-26.4 Horas positivas y finitas.** Se aceptan decimales (por ejemplo `2.5`). Se rechazan 0, los negativos, `NaN`, `+Inf` y `-Inf` con `ErrInvalidHours`. No hay límite superior ni se redondea el valor.
5. **RN-26.5 Sprint existente y abierto.** El sprint tiene que existir y estar `ABIERTO` (CA-26.1, CA-26.5). Se usa `sprint.Manager.WithOpenSprint`, que devuelve `ErrSprintNotFound` o `ErrSprintClosed` según corresponda.
6. **RN-26.6 Historia existente y perteneciente al sprint.** El ítem tiene que existir (`backlog.Manager.GetItem` con `ErrItemNotFound`) y su `SprintID` tiene que ser igual al SprintID recibido; si no, `ErrItemNotAssignedToSprint`. Un ítem sin asignar (`SprintID == 0`) o asignado a otro sprint no se acepta (CA-26.3). El estado del ítem es indistinto: se puede registrar esfuerzo sobre un ítem `PENDIENTE`, `EN_CURSO` o `TERMINADO`.
7. **RN-26.7 Solo agrega el esfuerzo.** Registrar no modifica el ítem, su estado, sus horas estimadas, sus Story Points ni el sprint. Las horas reales se obtienen sumando los esfuerzos (M5 de TEC-03); no se guarda un total en el ítem.
8. **RN-26.8 Identificador único.** Cada esfuerzo recibe un ID entero correlativo global que empieza en 1, asignado por el repositorio al guardar, con el mismo criterio que sprints e ítems.
9. **RN-26.9 Copias por valor.** `Register` devuelve una copia del esfuerzo. `Effort` no tiene slices ni mapas, así que copiarlo no comparte estado con lo guardado.
10. **RN-26.10 Persistencia.** El repositorio guarda en memoria y en PostgreSQL, detrás de la misma interfaz. La tabla `efforts` tiene claves foráneas a `backlog_items` y `sprints`; la pertenencia del ítem al sprint no se puede expresar con una clave foránea y la valida el Manager.
11. **RN-26.11 Atomicidad con el cierre y la asignación.** `sprint.Manager.WithOpenSprint(id, fn)` toma el mutex del sprint, valida que exista y esté `ABIERTO`, y ejecuta `fn` con el mutex tomado. `effort.Manager.Register` valida los datos, y dentro de `fn` consulta el ítem y guarda el esfuerzo. Así, `Register` no puede guardar un esfuerzo después de que `Close` cerró el sprint: si `Close` gana, `Register` devuelve `ErrSprintClosed`; si `Register` gana, el esfuerzo queda guardado y después se cierra. El orden de locks es siempre sprint → backlog/effort; el paquete `effort` nunca llama a `sprint` con su propio mutex tomado, así que no hay deadlock.
12. **RN-26.12 API e interfaz.** La operación se expone en `POST /api/sprints/{id}/items/{itemId}/esfuerzos` y en un formulario de la interfaz web, con un caso exitoso y un test por código de error, como exige el tema 5 del Planning del Sprint 3 y el criterio de TEC-04.

## Decisión sobre entidades referenciadas

HU-26 referencia **dos entidades por ID y valida las dos**:

1. **El sprint**, porque el esfuerzo se imputa a una historia de un sprint y su estado decide si se permite el registro (RN-26.5 y RN-26.11). Se usa `WithOpenSprint`, que devuelve `ErrSprintNotFound` o `ErrSprintClosed`.
2. **El ítem**, porque CA-26.3 exige que la historia pertenezca al sprint indicado y el esfuerzo se asocia a ese ítem (RN-26.6). Se reutiliza `GetItem` con `ErrItemNotFound` y `ErrItemNotAssignedToSprint`.

**El integrante no se valida como entidad** porque E4 no está en `main` ni comprometida en el Sprint 3; se guarda como texto (RN-26.1).

**El proyecto no se valida de forma directa:** si el ítem está asignado al sprint, la asignación de HU-11 ya garantiza que ambos pertenecen al mismo proyecto. Consultarlo sumaría una dependencia y otro mutex sin agregar una regla nueva.

## Restricciones

- **Organización del código.**
  - `internal/effort`: `effort.go` (tipo `Effort` y errores), `repository.go` (interfaz con `Save`), `memory_repository.go`, `postgres_repository.go`, `manager.go` (constructor y mutex) y `register.go` (`Register`).
  - `internal/sprint`: `with_open.go` (nuevo, `WithOpenSprint`). No se cambia el constructor ni el comportamiento existente.
  - `internal/httpapi`: `efforts.go` (nuevo, handler y response), `server.go` (el `server` recibe `*effort.Manager`, `NewHandler` suma el parámetro, la ruta y los errores nuevos en `statusFor`).
  - `internal/web`: `index.html` (formulario "Registrar esfuerzo" con `step="any"` en Horas) y `app.js` (llamada a `/esfuerzos`).
  - `cmd/server`: `repositories.go` (repositorio de esfuerzos en memoria o PostgreSQL), `main.go` (crea el `effort.Manager` y lo pasa a `NewHandler`) y `main_test.go` (verifica el repositorio en memoria).
  - `db/init/01-schema.sql`: tabla `efforts` e índices por `sprint_id` e `item_id`.
  - `internal/postgrestest/postgrestest.go`: incluir `efforts` en el `TRUNCATE`.
- **Constructores.** `effort.NewManager(repo Repository, sprints *sprint.Manager, items *backlog.Manager) *Manager`; `httpapi.NewHandler(projects *project.Manager, items *backlog.Manager, sprints *sprint.Manager, efforts *effort.Manager) http.Handler`. La dependencia va en una sola dirección: `effort` importa `sprint` y `backlog`; ninguno importa `effort` (sin ciclos).
- **Reparto de validaciones y locks.** `effort.Manager.Register` valida los datos y delega en `WithOpenSprint` la existencia, el estado y el mutex del sprint. Dentro de `WithOpenSprint` se consulta el ítem (backlog) y se guarda (effort). El orden es sprint → backlog y sprint → effort; nunca al revés.
- **Interfaz del repositorio.** Solo `Save(Effort) (Effort, error)`. Las lecturas y correcciones llegan con HU-27 y HU-28; no se agregan ahora.
- **Esquema.** `efforts` con `id` identity, `member TEXT`, `effort_date TIMESTAMPTZ`, `activity TEXT`, `item_id INTEGER REFERENCES backlog_items`, `sprint_id INTEGER REFERENCES sprints`, `hours DOUBLE PRECISION`. La columna se llama `effort_date` para no confundirla con el tipo `date`. Los scripts de `db/init` solo corren con el volumen vacío: una base anterior se recrea con `docker compose down -v` y `docker compose up -d` (documentado en README y evidencia).
- **Tests.** Unitarios en `internal/effort/register_test.go`; pasos de godog en `internal/effort/register_features_test.go`, con el runner `TestFeatures_HU26RegistrarEsfuerzo`, sobre `features/HU-26-registrar-esfuerzo.feature`. Los escenarios llevan el CA en el nombre (tema 7 del Planning). El contrato de `Save` va en `internal/effort/repository_contract_test.go` y corre en memoria y PostgreSQL. Los tests específicos de PostgreSQL van en `internal/effort/postgres_repository_test.go`; requieren `TEST_DATABASE_URL` y la base levantada, y se omiten sin la variable (CA-TEC05.7). `WithOpenSprint` se prueba en `internal/sprint/with_open_test.go`. La API se prueba en `internal/httpapi/efforts_test.go` (un caso exitoso y uno por código 400, 404, 409, 413 y 500) y la interfaz en `internal/httpapi/web_test.go`, que también verifica que el campo Horas use `step="any"`.
- **Tests de PostgreSQL.** Los corre y verifica Cardozo o Buttini antes del merge si el entorno local no tiene Docker.
- **Errores identificables.** Los cuatro errores de `internal/effort` son nuevos, exportados y comparables con `errors.Is`. Los de sprint y backlog se propagan tal cual. En `httpapi`, los cuatro errores de esfuerzo se agregan a los 400 de `statusFor`.
- **Interfaz web.** El formulario usa el sprint actual del contexto, el campo Horas acepta cualquier decimal (`step="any"`) y muestra los errores de la API; no agrega un listado de esfuerzos (HU-27).
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Definition of Done.** La SDD lleva la línea "Aprobada por"; antes del commit RED se corre `go vet ./...`; la concurrencia se verifica con `go test -race` en el CI.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Registrar esfuerzo en un ítem `EN_CURSO` asignado a un sprint `ABIERTO` | Se guarda con ID 1 y todos los campos recibidos; el ítem y el sprint no cambian |
| Registrar esfuerzo en un ítem `PENDIENTE` o `TERMINADO` del sprint | Se guarda igual: el estado del ítem es indistinto |
| Registrar dos esfuerzos | IDs 1 y 2, correlativos y globales |
| Horas `2.5` o `0.001` | Se aceptan |
| Horas 0, negativas, `NaN`, `+Inf` o `-Inf` | `ErrInvalidHours`; no se guarda nada |
| Integrante vacío o con solo espacios | `ErrEmptyMember`; no se guarda nada |
| Actividad vacía o con solo espacios | `ErrEmptyActivity`; no se guarda nada |
| Fecha zero | `ErrMissingDate`; no se guarda nada |
| Fecha futura | Se acepta: no hay restricción de futuro |
| Ítem sin asignar (`SprintID == 0`) | `ErrItemNotAssignedToSprint` |
| Ítem asignado a otro sprint | `ErrItemNotAssignedToSprint` |
| Sprint `CERRADO` | `ErrSprintClosed`; no se guarda nada |
| Sprint inexistente, 0 o negativo | `ErrSprintNotFound`; no se guarda nada |
| Ítem inexistente, 0 o negativo | `ErrItemNotFound`; no se guarda nada |
| Integrante y actividad con espacios en los extremos | Se guardan recortados |
| Modificar el esfuerzo devuelto | El esfuerzo guardado no cambia |
| Varios registros simultáneos | Todos se guardan, con IDs únicos; `-race` sin condiciones de carrera |
| `Register` y `Close` simultáneos | Si `Close` gana, `Register` devuelve `ErrSprintClosed` y no guarda; si `Register` gana, el esfuerzo queda guardado y después se cierra; nunca se guarda después del cierre |
| El repositorio falla al guardar | Se devuelve ese error y no se guarda nada |
| HTTP: `{"horas":"2.5"}` | Se acepta como 2.5 |
| HTTP: `{"horas":null}`, `{}` o `{"horas":"muchas"}` | 400 con el mensaje de `ErrInvalidHours` |
| HTTP: `{"fecha":""}` | 400 con el mensaje de `ErrMissingDate` |
| HTTP: `{"fecha":"24/10/2026"}` | 400 con el mensaje de `errInvalidDate` |
| HTTP: ítem de otro sprint | 409 con el mensaje de `ErrItemNotAssignedToSprint` |
| HTTP: cuerpo de más de 1 MiB | 413 sin procesar el cuerpo |
| Interfaz: registrar desde el formulario | Muestra "Esfuerzo N registrado…"; un error de la API se muestra en el mensaje |
| Interfaz: campo Horas | `type="number"` con `step="any"` |

## Criterios de aceptación

- **CA-26.1** Se registra un esfuerzo con integrante, fecha, actividad, ítem, sprint y horas: se guarda con un ID correlativo, se devuelve una copia y el ítem y el sprint no cambian.
- **CA-26.2** Las horas tienen que ser un número positivo y finito: 0, negativas, `NaN` e infinitos devuelven `ErrInvalidHours` y no se guarda nada. Se aceptan decimales.
- **CA-26.3** La historia tiene que estar asignada al sprint indicado: un ítem sin asignar o asignado a otro sprint devuelve `ErrItemNotAssignedToSprint` y no se guarda nada.
- **CA-26.4** Un sprint inexistente (incluidos 0 y negativos) devuelve `ErrSprintNotFound`; un ítem inexistente devuelve `ErrItemNotFound`. En ningún caso se guarda nada.
- **CA-26.5** No se registra esfuerzo en un sprint `CERRADO`: devuelve `ErrSprintClosed` y no se guarda nada.
- **CA-26.6** Integrante, actividad y fecha son obligatorios: vacíos, solo espacios o fecha zero devuelven `ErrEmptyMember`, `ErrEmptyActivity` o `ErrMissingDate`, y no se guarda nada.
- **CA-26.7** Registrar un esfuerzo no modifica el ítem, su estado, sus horas estimadas, sus Story Points ni el sprint.
- **CA-26.8** Modificar el esfuerzo devuelto por `Register` no altera el esfuerzo guardado.
- **CA-26.9** Si el repositorio falla al guardar, se devuelve ese error y no se guarda nada.
- **CA-26.10** Varios registros simultáneos se guardan todos, con IDs únicos y sin condiciones de carrera. `Register` y `Close` simultáneos no permiten guardar un esfuerzo después del cierre: si `Close` gana, `Register` devuelve `ErrSprintClosed`; si `Register` gana, el esfuerzo queda guardado antes del cierre (verificado con `go test -race`).
- **CA-26.11** Se registra un esfuerzo desde la API (`POST /api/sprints/{id}/items/{itemId}/esfuerzos`) y desde el formulario "Registrar esfuerzo" de la interfaz web. La API responde 201 en el caso exitoso y 400, 404, 409, 413 o 500 según la tabla HTTP; la interfaz muestra el éxito o el mensaje de error. Se cubre con tests de `internal/httpapi` (`efforts_test.go` y `web_test.go`), no con escenarios godog.

## BDD y tests previstos

**Archivo BDD:** `features/HU-26-registrar-esfuerzo.feature`, con el runner `TestFeatures_HU26RegistrarEsfuerzo` en `internal/effort/register_features_test.go`. Cada escenario lleva su CA en el nombre:

- Registrar un esfuerzo válido (CA-26.1)
- Registrar un esfuerzo con horas fraccionadas (CA-26.2)
- Registrar sobre un ítem terminado del sprint (CA-26.1 y RN-26.6)
- Rechazar horas no positivas o no finitas (CA-26.2)
- Rechazar un ítem que no pertenece al sprint (CA-26.3)
- Rechazar sprint e ítem inexistentes (CA-26.4)
- Rechazar un sprint cerrado (CA-26.5)
- Rechazar integrante, actividad o fecha ausentes (CA-26.6)
- No modificar el ítem al registrar (CA-26.7)
- Modificar el esfuerzo devuelto no altera el guardado (CA-26.8)
- Error del repositorio al guardar (CA-26.9)
- Registros simultáneos y cierre simultáneo (CA-26.10)

**Tests unitarios:** `internal/effort/register_test.go` cubre CA-26.1 a CA-26.10 y los casos límite. `internal/sprint/with_open_test.go` cubre el sprint abierto, cerrado, inexistente y la propagación del error de la función.

**Tests de contrato:** `internal/effort/repository_contract_test.go` verifica que `Save` asigna IDs correlativos desde 1, conserva todos los campos, no depende de un ID recibido y devuelve copias, con `MemoryRepository` y `PostgresRepository`. El fixture de PostgreSQL crea el proyecto, el ítem y el sprint necesarios por las claves foráneas.

**Tests de PostgreSQL:** `internal/effort/postgres_repository_test.go` verifica la persistencia y la normalización de la fecha a UTC. Se omiten sin `TEST_DATABASE_URL`.

**Tests de API:** `internal/httpapi/efforts_test.go` con un caso exitoso (201) y uno por código de error (400 de datos y JSON, 404 de sprint e ítem, 409 de sprint cerrado e ítem de otro sprint, 413 de cuerpo grande y 500 de repositorio). Se agrega el endpoint a `bodyEndpoints` de `server_test.go` para 413 y texto después del JSON. `internal/httpapi/web_test.go` verifica la columna/formulario, la llamada de `app.js` y que el campo Horas tenga `step="any"`.

## Fuera de alcance

- Consultar esfuerzos (HU-27).
- Corregir o eliminar esfuerzos (HU-28).
- Registrar integrantes como entidad (E4).
- Modificar horas estimadas, Story Points o estado del ítem.
- Mover el esfuerzo entre sprints o ítems.
- Listar los esfuerzos en la interfaz (HU-27).
- Métricas, dashboard, reportes y exportación.
- Cualquier otra pantalla o endpoint que no sea el de registrar esfuerzo.

## Comparación con el acta de Planning

El acta del Sprint 3 (`docs/actas/sprint-3/planning.md`, Issue #110) incluye a HU-26 con **5 Story Points**, responsable **García Santiago**, dependiente de **HU-11** ("historia asignada al sprint") y bloqueante de **HU-27, HU-28 y HU-33 (Sprint 4)**. Su justificación es "Paquete nuevo; la historia tiene que pertenecer al sprint (HU-11)". El **tema 5** exige que cada historia nueva sume su endpoint bajo `/api/…` y su pantalla en la interfaz web, con un caso exitoso y un test por código de error; el **tema 7** pide el CA en el nombre de cada escenario BDD; la **Definition of Done** pide la línea "Aprobada por" en cada SDD y `go vet ./...` antes del commit RED.

**El alcance coincide:** esta SDD crea un paquete nuevo (`internal/effort`), valida que la historia pertenezca al sprint (HU-11) y suma el endpoint y el formulario web del tema 5. No hay contradicción con el acta. Esta SDD agrega tres precisiones que el acta no detalla y que quedaron confirmadas para esta historia:

1. **Solo sprint `ABIERTO`** (RN-26.5 y CA-26.5): el acta no lo dice, pero es una restricción coherente con HU-11, HU-12 y HU-13.
2. **Integrante como texto libre** (RN-26.1): el acta deja E4 (HU-15 a HU-17) fuera de los requerimientos mínimos y solo para capacidad sobrante; por eso no se puede depender de una entidad de integrante en este Sprint.
3. **`WithOpenSprint`** (RN-26.11): es un método de coordinación en `sprint.Manager` para que el registro sea atómico con el cierre, como `CompleteItem` en HU-12; no agrega funcionalidad visible al usuario.

La estimación de 5 Story Points se mantiene. La dependencia HU-11 está integrada en `main`.

## Comparación con los criterios del Issue #44 (borrador)

| CA en el Issue #44 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-26.1 Se registra el esfuerzo con integrante, fecha, actividad, historia, sprint y horas trabajadas | CA-26.1 | Se precisa el ID correlativo, el recorte de textos y que el ítem y el sprint no cambian |
| CA-26.2 Las horas deben ser un número positivo | CA-26.2 | Se agregan decimales, `NaN` e infinitos; 0 no es positivo |
| CA-26.3 La historia debe pertenecer al sprint indicado | CA-26.3 | Se identifica `ErrItemNotAssignedToSprint` y se cubren ítem sin asignar y de otro sprint |
| — | CA-26.4 (nuevo) | Existencia de sprint e ítem, incluidos IDs 0 y negativos |
| — | CA-26.5 (nuevo) | Solo se registra en un sprint `ABIERTO` |
| — | CA-26.6 (nuevo) | Integrante, actividad y fecha obligatorios |
| — | CA-26.7 (nuevo) | El registro no modifica el ítem ni el sprint |
| — | CA-26.8 (nuevo) | Lo que se devuelve al registrar no permite cambiar el esfuerzo guardado |
| — | CA-26.9 (nuevo) | Error del repositorio propagado |
| — | CA-26.10 (nuevo) | Concurrencia y atomicidad con el cierre, verificadas con `-race` |
| — | CA-26.11 (nuevo) | Endpoint y formulario web (tema 5 del Planning) |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #44** (en lenguaje de negocio, mejora 8 del Planning; se reemplaza cuando se apruebe la SDD y con tu autorización):

```markdown
## Criterios de aceptación

- [ ] CA-26.1 Se registra un esfuerzo con integrante, fecha, actividad, historia, sprint y horas; queda con un ID y no cambia la historia ni el sprint.
- [ ] CA-26.2 Las horas trabajadas tienen que ser un número positivo; se aceptan decimales y se rechazan cero, negativos e infinitos.
- [ ] CA-26.3 La historia tiene que estar asignada al sprint indicado; si no, el registro falla y no se guarda nada.
- [ ] CA-26.4 El sprint y la historia tienen que existir; si no, el registro falla y no se guarda nada.
- [ ] CA-26.5 Solo se registra esfuerzo en un sprint abierto; en un sprint cerrado el registro falla y no se guarda nada.
- [ ] CA-26.6 Integrante, actividad y fecha son obligatorios.
- [ ] CA-26.7 Registrar el esfuerzo no modifica la historia ni el sprint.
- [ ] CA-26.8 Lo que se devuelve al registrar no permite cambiar el esfuerzo guardado.
- [ ] CA-26.9 Un error al guardar deja todo como estaba.
- [ ] CA-26.10 Varios registros simultáneos se guardan todos y no se pisan; un cierre simultáneo no permite guardar después del cierre.
- [ ] CA-26.11 El esfuerzo se registra desde la API y desde un formulario de la interfaz web, con mensajes de éxito y de error entendibles.
```

## Decisiones confirmadas para esta SDD el 24/10

1. **Integrante como texto libre.** No se crea ni se simula E4; el nombre se guarda como texto obligatorio (RN-26.1).
2. **Solo sprint `ABIERTO`.** Un sprint `CERRADO` devuelve `ErrSprintClosed` (RN-26.5 y CA-26.5).
3. **Fecha y actividad obligatorias.** Fecha distinta de zero, actividad no vacía; horas positivas finitas; sin límite superior ni restricción de fecha futura (RN-26.2 a RN-26.4).
4. **Paquete nuevo y persistencia.** `internal/effort` con repositorios en memoria y PostgreSQL; tabla `efforts` con claves foráneas a ítems y sprints (RN-26.10).
5. **Sin cambios en el ítem.** El esfuerzo real se calcula sumando los registros; no se agrega un campo de horas reales al ítem (RN-26.7).
6. **`WithOpenSprint` en lugar de `GetByID`.** El registro valida el sprint abierto y guarda con el mutex del sprint tomado, como `CompleteItem` en HU-12; no queda la limitación de la versión anterior (RN-26.11).
7. **API e interfaz.** Se agregan el endpoint y el formulario del tema 5 del Planning, con un CA propio (CA-26.11) y tests de API e interfaz. El campo Horas usa `step="any"` para aceptar los mismos decimales que la API.
8. **Métricas.** El registro alimenta M5, M6 y M7 de `docs/metricas.md` (TEC-03); E9 las implementa en el Sprint 4.
