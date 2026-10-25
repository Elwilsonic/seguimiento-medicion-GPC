# HU-29 · Registrar un defecto

> Issue del repositorio: #47

Aprobada por: Peñalbé Hernán (Agile Enabler), 24/10/2026, con el ajuste del sprint de detección faltante (400).

## Historia de usuario

Como Product Builder, quiero registrar un defecto asociado a una historia, para llevar el control de la calidad.

## Objetivo

Registrar un defecto con descripción, severidad, historia del backlog y sprint de detección obligatorio y abierto, en un paquete nuevo `internal/defect`. El defecto queda con estado inicial `ABIERTO`, sin sprint de resolución y con un ID correlativo global.

Es el dato de entrada de **M9 (defectos detectados)** de `docs/metricas.md` (TEC-03): todo defecto suma en M9 de su sprint de detección y en M9 del proyecto. HU-30 completará el estado y el sprint de resolución, que alimentan **M10 (defectos resueltos)**.

Se expone en `POST /api/items/{id}/defectos` y en un formulario de la interfaz web (tema 5 del Planning del Sprint 3). Persiste en memoria y en PostgreSQL (TEC-05).

## Alcance

- **Incluye:** la entidad `Defect`, sus estados y severidades, la interfaz `Repository`, los repositorios en memoria y PostgreSQL, el `Manager`, la operación `Register`, la validación de la historia y del sprint de detección obligatorio y abierto, el endpoint HTTP y el formulario web; tests unitarios, BDD, de contrato, de PostgreSQL, de API y de interfaz.
- **No incluye:** cambiar el estado del defecto ni registrar el sprint de resolución (HU-30); consultar o listar defectos (HU-31); editar o eliminar defectos; calcular métricas (E9), dashboard, reportes o exportación; registrar integrantes como entidad (E4); modificar la historia, el sprint o sus estados.

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ItemID | entero | Sí | Historia del backlog a la que se asocia el defecto |
| Descripción | texto | Sí | Descripción del defecto; se guarda sin espacios en los extremos |
| Severidad | texto | Sí | `CRITICA`, `ALTA`, `MEDIA` o `BAJA` |
| Sprint de detección | entero | Sí | Sprint abierto donde se detectó; debe existir y ser del mismo proyecto que la historia |

El ID no es una entrada: lo asigna el repositorio. El estado tampoco es una entrada: siempre es `ABIERTO`. El sprint de resolución no se recibe: siempre queda en `0`.

Firma: `Register(itemID int, description, severity string, detectionSprintID int) (Defect, error)` en el `Manager` de `internal/defect`.

### Estados y severidades (definición única)

En `internal/defect/defect.go`:

- `Status` con los valores `ABIERTO`, `EN_CORRECCION` y `RESUELTO`.
- `StatusOpen`, `StatusInCorrection` y `StatusResolved` como constantes exportadas.
- HU-29 solo usa `StatusOpen` como estado inicial. `EN_CORRECCION` y `RESUELTO` quedan definidos para HU-30; esta historia no los asigna ni implementa transiciones.
- `Severity` con los valores `CRITICA`, `ALTA`, `MEDIA` y `BAJA`.
- `SeverityCritical`, `SeverityHigh`, `SeverityMedium` y `SeverityLow` como constantes exportadas.
- `severityValues = [...]string{"CRITICA", "ALTA", "MEDIA", "BAJA"}`: definición única, no exportada.
- `SeverityValues() []string`: devuelve una **lista nueva** con los valores en ese orden, para que HU-30 y la interfaz web no repitan la lista. Cambiar la lista devuelta no altera la validación.

La severidad se normaliza con `strings.ToUpper(strings.TrimSpace(...))`, igual que la prioridad y el estado del backlog.

### Salidas esperadas

- **Si el registro es válido:** el defecto guardado, como **copia por valor**, con:
  - **ID:** entero correlativo global asignado por el repositorio (el primero es 1).
  - **Descripción:** la recibida, sin espacios en los extremos.
  - **Severidad:** normalizada en mayúsculas y sin espacios.
  - **Estado:** `ABIERTO`.
  - **ItemID:** el recibido.
  - **Sprint de detección:** el recibido; siempre es un sprint existente, abierto y del mismo proyecto que la historia.
  - **Sprint de resolución:** `0`.
- **Si el registro no es válido:** el error que indica el motivo (ver "Condiciones de error"); no se guarda ningún defecto.

### HTTP (TEC-04, tema 5 del Planning)

`POST /api/items/{id}/defectos`, con el mismo patrón de IDs en la ruta que los endpoints de ítems.

- **Cuerpo:**

  ```json
  {"descripcion": "El total no descuenta el descuento", "severidad": "ALTA", "sprint_deteccion_id": 3}
  ```

  `sprint_deteccion_id` es obligatorio: si falta, es 0 o es negativo, la API responde 400 con el mensaje de sprint de detección obligatorio; si es un ID positivo inexistente, responde 404 con el mensaje de sprint inexistente.

- **Respuesta exitosa:** `201 Created` con el defecto:

  ```json
  {"id": 1, "descripcion": "El total no descuenta el descuento", "severidad": "ALTA", "estado": "ABIERTO", "item_id": 4, "sprint_deteccion_id": 3, "sprint_resolucion_id": 0}
  ```

- **Errores:**

| Caso | Código | Mensaje |
|---|---|---|
| Descripción vacía o con solo espacios | 400 | El de `ErrEmptyDescription` |
| Severidad vacía o fuera de CRITICA, ALTA, MEDIA y BAJA | 400 | El de `ErrInvalidSeverity` |
| Cuerpo que no es JSON, `sprint_deteccion_id` que no es entero, o texto después del JSON | 400 | El de `errInvalidJSON` |
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| Sprint de detección ausente, 0 o negativo | 400 | El de `ErrMissingDetectionSprint` |
| Sprint de detección positivo inexistente | 404 | El de `sprint.ErrSprintNotFound` |
| Historia inexistente | 404 | El de `backlog.ErrItemNotFound` |
| Sprint de detección `CERRADO` | 409 | El de `sprint.ErrSprintClosed` |
| Sprint de detección de otro proyecto | 409 | El de `ErrDetectionSprintFromAnotherProject` |
| Cuerpo de más de 1 MiB | 413 | El de `errBodyTooLarge` |
| Falla del repositorio | 500 | El de `errInternal`, sin detalle |

`estado` y `sprint_resolucion_id` no se reciben: si un cliente los envía, se ignoran. El orden de los errores del handler es: ID de la ruta, JSON y `Manager`; dentro del `Manager`, descripción, severidad, sprint de detección <= 0 (antes de tocar repositorios), sprint de detección positivo (existencia y estado) e historia (existencia y proyecto).

### Interfaz web (TEC-04, tema 5)

En `internal/web`, una sección **Defectos** con el formulario "Registrar defecto":

- Campos: Ítem (ID), Descripción, Severidad (`select` con CRITICA, ALTA, MEDIA y BAJA) y Sprint de detección (ID), obligatorio (`required`, `min="1"`).
- Llama a `POST /items/{id}/defectos` con el sprint de detección informado; si el campo falta, la API responde 400 con el error de sprint de detección obligatorio.
- Al terminar, muestra "Defecto N registrado para el ítem X." o el mensaje de error de la API, con el mismo mecanismo de acción y mensaje que los demás formularios.
- No agrega un listado de defectos: eso es HU-31.
- La lista de severidades de `app.js` se compara por test con `defect.SeverityValues()`, para que no se desincronice.

## Reglas de negocio

1. **RN-29.1 Descripción obligatoria.** Se recorta con `strings.TrimSpace`; si queda vacía, `ErrEmptyDescription`.
2. **RN-29.2 Severidad obligatoria y válida.** Solo se aceptan CRITICA, ALTA, MEDIA y BAJA, sin distinguir mayúsculas ni los espacios de los extremos. Cualquier otro valor devuelve `ErrInvalidSeverity`.
3. **RN-29.3 Estado inicial `ABIERTO`.** El estado no se recibe y no se puede elegir al crear. `EN_CORRECCION` y `RESUELTO` quedan definidos para HU-30.
4. **RN-29.4 Sin sprint de resolución.** Al crear, el sprint de resolución siempre es 0; no se recibe y HU-30 lo completará al resolver.
5. **RN-29.5 Historia obligatoria y existente.** Se reutiliza `backlog.Manager.GetItem` y su `ErrItemNotFound`. La historia puede estar `PENDIENTE`, `EN_CURSO` o `TERMINADO`: registrar un defecto no cambia su estado.
6. **RN-29.6 Sprint de detección obligatorio y abierto.** Se valida en dos pasos:
   - el campo es obligatorio; ausente, 0 o negativo devuelve `ErrMissingDetectionSprint` antes de tocar los repositorios;
   - para un ID positivo se usa `sprint.Manager.WithOpenSprint`, igual que HU-26: el sprint tiene que existir; si no, se propaga `sprint.ErrSprintNotFound`;
   - el sprint tiene que estar `ABIERTO`; si está `CERRADO`, devuelve `sprint.ErrSprintClosed`;
   - tiene que ser del mismo proyecto que la historia; si no, `ErrDetectionSprintFromAnotherProject`;
   - la historia no tiene que estar asignada a ese sprint.

   Un defecto que aparece después de cerrar un sprint se registra en el sprint abierto actual.
7. **RN-29.7 Solo agrega el defecto.** Registrar no modifica la historia, su estado, su sprint, sus Story Points ni el sprint de detección.
8. **RN-29.8 Identificador único.** Cada defecto recibe un ID entero correlativo global que empieza en 1, asignado por el repositorio al guardar.
9. **RN-29.9 Copias por valor.** `Register` devuelve una copia del defecto. `Defect` no tiene slices ni mapas, así que copiar el struct no comparte estado con lo guardado.
10. **RN-29.10 Persistencia.** El repositorio guarda en memoria y en PostgreSQL, detrás de la misma interfaz `Repository`.
11. **RN-29.11 Concurrencia y atomicidad con el cierre.** `Manager.Register` toma su propio mutex durante toda la validación y el guardado, porque el repositorio en memoria no es seguro para uso concurrente. Dentro de ese mutex usa `sprint.Manager.WithOpenSprint`, que valida la existencia y el estado `ABIERTO` del sprint y ejecuta el guardado con el mutex del sprint tomado, como HU-26. Así, un cierre simultáneo no permite registrar un defecto en un sprint ya cerrado: si `Close` gana, `Register` devuelve `ErrSprintClosed`; si `Register` gana, el defecto queda guardado y después se cierra. El orden de locks es defect → sprint → backlog; ningún otro paquete llama a `defect`, así que no hay ciclos.
12. **RN-29.12 API e interfaz.** La operación se expone en `POST /api/items/{id}/defectos` y en un formulario web, con un caso exitoso y un test por código de error, como exige el tema 5 del Planning del Sprint 3.
13. **RN-29.13 M9 y M10.** Todo defecto tiene sprint de detección, así que suma en M9 de ese sprint y en M9 del proyecto. Como HU-29 deja todos los defectos en `ABIERTO` y con sprint de resolución 0, no aporta a M10; eso lo hará HU-30.

## Decisión sobre entidades referenciadas

- **Historia:** **sí se valida su existencia** (RN-29.5), porque el defecto se asocia a ella y hace falta su `ProjectID` para comparar el proyecto del sprint de detección. Se usa `backlog.Manager.GetItem`, que devuelve una copia.
- **Sprint de detección:** **sí se valida siempre** (RN-29.6), porque es obligatorio. Ausente, 0 o negativo devuelve `ErrMissingDetectionSprint` antes de tocar repositorios. Para un ID positivo se usa `sprint.Manager.WithOpenSprint`, que devuelve `sprint.ErrSprintNotFound` si no existe, `sprint.ErrSprintClosed` si está cerrado, y ejecuta la validación de la historia y el guardado con el mutex del sprint tomado. Además se compara `sprint.ProjectID` con el de la historia; si no coincide, `ErrDetectionSprintFromAnotherProject`. No se exige que la historia pertenezca al sprint, porque un defecto puede encontrarse en un sprint posterior al de la historia. Un defecto que aparece después de cerrar un sprint se registra en el sprint abierto actual.
- **Proyecto:** **no se valida de forma directa.** Se deriva de la historia, que ya fue creada con un proyecto existente (HU-04). El sprint de detección se compara contra ese proyecto; no se consulta el `project.Manager`.
- **Integrante:** no interviene en HU-29; el defecto no registra quién lo detectó. E4 no está comprometido en el Sprint 3.
- **Estados `EN_CORRECCION` y `RESUELTO`:** se definen ahora para que HU-30 los use, pero HU-29 no los asigna ni implementa transiciones.

## Restricciones

- **Organización del código.**
  - `internal/defect` (paquete nuevo): `defect.go` (tipo `Defect`, `Status`, `Severity`, constantes, `SeverityValues` y errores), `repository.go` (interfaz con `Save`), `memory_repository.go`, `postgres_repository.go`, `manager.go` (constructor y mutex) y `register.go` (`Register`).
  - `internal/sprint`: no se agrega código. Se reutiliza `WithOpenSprint` de HU-26.
  - `internal/httpapi`: `defects.go` (nuevo, handler y respuesta) y `server.go` (el server recibe `*defect.Manager`, `NewHandler` suma el parámetro, la ruta y los errores nuevos en `statusFor`).
  - `internal/web`: `index.html` (sección y formulario "Registrar defecto") y `app.js` (llamada a `/defectos` y lista de severidades).
  - `cmd/server`: `repositories.go` (repositorio de defectos en memoria o PostgreSQL), `main.go` (crea el `defect.Manager` y lo pasa a `NewHandler`) y `main_test.go` (verifica el repositorio en memoria y, con base, el de PostgreSQL).
  - `db/init/01-schema.sql`: tabla `defects` e índices por `item_id`, `detection_sprint_id` y `resolution_sprint_id`.
  - `internal/postgrestest/postgrestest.go`: incluir `defects` en el `TRUNCATE`.
- **Constructores.** `defect.NewManager(repo Repository, items *backlog.Manager, sprints *sprint.Manager) *Manager`; `httpapi.NewHandler` suma un `*defect.Manager`. El orden exacto de los parámetros se ajusta al de `main`, que desde HU-21 también incluye `*poker.Manager`. La dependencia va en una sola dirección: `defect` importa `backlog` y `sprint`; ninguno importa `defect` (sin ciclos).
- **Reparto de validaciones y locks.** `Register` valida descripción, severidad y sprint de detección <= 0 antes de tocar repositorios; después toma su mutex y, para un ID positivo, delega en `WithOpenSprint` la existencia, el estado `ABIERTO` y el mutex del sprint. Dentro de `WithOpenSprint` consulta la historia, valida el proyecto y guarda. El orden es defect → sprint → backlog. El guardado ocurre con el mutex del `Manager` y el del sprint tomados.
- **Interfaz del repositorio.** Solo `Save(Defect) (Defect, error)`. Las lecturas, listados y cambios de estado llegan con HU-30 y HU-31; no se agregan ahora.
- **Esquema.** `defects` con `id` identity, `description TEXT NOT NULL`, `severity TEXT NOT NULL`, `status TEXT NOT NULL`, `item_id INTEGER NOT NULL REFERENCES backlog_items (id)`, `detection_sprint_id INTEGER NOT NULL REFERENCES sprints (id)` y `resolution_sprint_id INTEGER NOT NULL DEFAULT 0`. El sprint de detección es obligatorio y lleva clave foránea, como `efforts.sprint_id` (HU-26); el de resolución queda en 0 hasta HU-30 y no lleva clave foránea porque 0 no es un sprint, con el mismo criterio que `backlog_items.sprint_id` (HU-11). Los scripts de `db/init` solo corren con el volumen vacío: una base anterior se recrea con `docker compose down -v` y `docker compose up -d`.
- **Errores identificables.** `ErrEmptyDescription` ("la descripción del defecto es obligatoria"), `ErrInvalidSeverity` ("la severidad debe ser CRITICA, ALTA, MEDIA o BAJA"), `ErrMissingDetectionSprint` ("el sprint de detección es obligatorio") y `ErrDetectionSprintFromAnotherProject` ("el sprint de detección pertenece a otro proyecto") son nuevos, exportados y comparables con `errors.Is`. Se propagan `backlog.ErrItemNotFound`, `sprint.ErrSprintNotFound` y `sprint.ErrSprintClosed`. En `httpapi`, los de validación se traducen a 400, los de inexistencia a 404 y los de sprint cerrado o de otro proyecto a 409.
- **RED y GREEN.** En el RED se agregan los tipos, errores, la interfaz y firmas vacías (`Register`, `Save`, handler 501) para que todo compile; los tests fallan por aserciones, no por compilación. No se toca `cmd/server` ni se cablea `NewHandler` en el RED: el handler se prueba construyendo el server dentro del paquete `httpapi`, y el test de la ruta se escribe igual y falla hasta que el GREEN registre la ruta. El cableado real (`repos.defects`, `NewHandler`, ruta y `cmd/server`) va en el GREEN. La rama sale de un `main` que ya incluye HU-21 (PR #122); si entra otro PR a `main` antes del GREEN, se hace `git fetch` y `git merge origin/main`. Antes del commit RED se corre `go vet ./...`.
- **Tests.** Unitarios en `internal/defect/register_test.go`; pasos de godog en `internal/defect/register_features_test.go`, con el runner `TestFeatures_HU29RegistrarDefecto`, sobre `features/HU-29-registrar-defecto.feature`. Los escenarios llevan el CA en el nombre (tema 7 del Planning). El contrato de `Save` va en `internal/defect/repository_contract_test.go` y corre en memoria y PostgreSQL. Los tests específicos de PostgreSQL van en `internal/defect/postgres_repository_test.go`; requieren `TEST_DATABASE_URL` y la base levantada, y se omiten sin la variable (CA-TEC05.7). La API se prueba en `internal/httpapi/defects_test.go` (un caso exitoso y uno por código 400, 404, 409, 413 y 500), la ruta y el `bodyEndpoints` en `server_test.go`, y la interfaz en `internal/httpapi/web_test.go`. El wiring se prueba en `cmd/server/main_test.go`.
- **Tests de PostgreSQL.** Los corre y verifica otro integrante antes del merge si el entorno local no tiene Docker.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Definition of Done.** La SDD lleva la línea "Aprobada por"; antes del commit RED se corre `go vet ./...`; la concurrencia se verifica con `go test -race` en el CI; no hay firmas ni menciones de la IA en commits, PRs, Issues ni comentarios.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Registrar un defecto con descripción, severidad, historia y sprint de detección abierto | Se guarda con ID 1, estado `ABIERTO` y sprint de resolución 0 |
| Registrar con sprint de detección `CERRADO` | `sprint.ErrSprintClosed`; no se guarda nada |
| Registrar con sprint de detección ausente, 0 o negativo | `ErrMissingDetectionSprint`; no se guarda nada |
| Registrar con sprint de detección positivo inexistente | `sprint.ErrSprintNotFound`; no se guarda nada |
| Registrar sobre una historia `TERMINADO` | Se guarda; la historia no cambia |
| Registrar sobre una historia que no pertenece al sprint de detección | Se guarda: la pertenencia no se exige |
| Registrar un defecto detectado después de cerrar un sprint, usando el sprint abierto actual | Se guarda en el sprint abierto actual |
| Descripción con espacios en los extremos | Se guarda recortada |
| Descripción vacía o con solo espacios | `ErrEmptyDescription`; no se guarda nada |
| Severidad en minúsculas o con espacios (`alta`, ` ALTA `) | Se normaliza a `ALTA` y se guarda |
| Severidad `URGENTE`, `BLOQUEANTE`, vacía o desconocida | `ErrInvalidSeverity`; no se guarda nada |
| Historia inexistente, 0 o negativa | `backlog.ErrItemNotFound`; no se guarda nada |
| Sprint de detección de otro proyecto | `ErrDetectionSprintFromAnotherProject`; no se guarda nada |
| Sprint de detección `CERRADO` e historia inexistente a la vez | `sprint.ErrSprintClosed` (el sprint se valida antes que la historia) |
| Descripción inválida e historia inexistente a la vez | `ErrEmptyDescription` (la descripción se valida primero) |
| Severidad inválida e historia inexistente a la vez | `ErrInvalidSeverity` (la severidad se valida antes del repositorio) |
| Descripción inválida y sprint de detección ausente a la vez | `ErrEmptyDescription` (la descripción se valida primero) |
| Severidad inválida y sprint de detección ausente a la vez | `ErrInvalidSeverity` (la severidad se valida antes del sprint) |
| Historia inexistente y sprint de detección ausente a la vez | `ErrMissingDetectionSprint` (el sprint <= 0 se valida antes de consultar la historia) |
| Modificar el defecto devuelto | El defecto guardado no cambia |
| Varios registros simultáneos | Todos se guardan, con IDs únicos; `go test -race` sin carreras |
| Registro simultáneo con el cierre del sprint | Si `Close` gana, `ErrSprintClosed`; si `Register` gana, queda guardado |
| El repositorio falla al guardar | Se devuelve ese error y no se guarda nada |
| HTTP: `sprint_deteccion_id` ausente, 0 o `null` | 400 con el mensaje de `ErrMissingDetectionSprint` |
| HTTP: `sprint_deteccion_id` positivo inexistente | 404 con el mensaje de `sprint.ErrSprintNotFound` |
| HTTP: `sprint_deteccion_id` decimal o texto | 400 con el mensaje de JSON inválido |
| HTTP: `{id}` no numérico | 400 con el mensaje de ID inválido |
| HTTP: cuerpo de más de 1 MiB | 413 sin procesar el cuerpo |
| Interfaz: registrar desde el formulario | Muestra "Defecto N registrado…"; un error de la API se muestra en el mensaje |
| Interfaz: campo Sprint de detección vacío | El formulario no lo permite (`required`); si se llama a la API sin el campo, 400 |
| Interfaz: campo Severidad | `select` con CRITICA, ALTA, MEDIA y BAJA, igual que `defect.SeverityValues()` |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Descripción vacía o con solo espacios (RN-29.1) | `ErrEmptyDescription` | No se guarda nada |
| Severidad vacía o fuera de la lista (RN-29.2) | `ErrInvalidSeverity` | No se guarda nada |
| Sprint de detección ausente, 0 o negativo (RN-29.6) | `ErrMissingDetectionSprint` | No se guarda nada |
| Sprint de detección positivo inexistente (RN-29.6) | `sprint.ErrSprintNotFound` | No se guarda nada |
| El sprint de detección está `CERRADO` (RN-29.6) | `sprint.ErrSprintClosed` | No se guarda nada |
| No existe la historia (RN-29.5) | `backlog.ErrItemNotFound` | No se guarda nada |
| El sprint de detección es de otro proyecto (RN-29.6) | `ErrDetectionSprintFromAnotherProject` | No se guarda nada |
| El repositorio falla al guardar | El error del repositorio | No se guarda nada |

**Orden de validación:** descripción, severidad, sprint de detección <= 0 (antes de tocar repositorios), sprint de detección positivo (existencia y estado), historia (existencia y proyecto) y guardado. Se informa el primer error encontrado.

## Criterios de aceptación

- **CA-29.1** Se registra un defecto con descripción, severidad, historia y sprint de detección obligatorio. Queda guardado con un ID correlativo, la descripción recortada, la severidad normalizada y el sprint de detección recibido.
- **CA-29.2** Descripción, severidad, historia y sprint de detección son obligatorios. Si falta la descripción o la severidad, se aplican `ErrEmptyDescription` o `ErrInvalidSeverity`; si el sprint de detección falta, es 0 o negativo, `ErrMissingDetectionSprint`; si la historia falta o no existe, `backlog.ErrItemNotFound`. En todos los casos el registro falla y no se guarda nada.
- **CA-29.3** El defecto queda en estado inicial `ABIERTO` y sin sprint de resolución. `EN_CORRECCION` y `RESUELTO` quedan definidos para HU-30; HU-29 no los asigna.
- **CA-29.4** La severidad es una de CRITICA, ALTA, MEDIA o BAJA, sin distinguir mayúsculas ni espacios en los extremos. Cualquier otro valor devuelve `ErrInvalidSeverity`.
- **CA-29.5** La historia tiene que existir; si no, `backlog.ErrItemNotFound`. Puede estar `PENDIENTE`, `EN_CURSO` o `TERMINADO`, y registrar un defecto no cambia su estado.
- **CA-29.6** El sprint de detección es obligatorio y tiene que existir y estar `ABIERTO`: ausente, 0 o negativo devuelve `ErrMissingDetectionSprint`; un ID positivo inexistente devuelve `sprint.ErrSprintNotFound`; `CERRADO` devuelve `sprint.ErrSprintClosed`; si es de otro proyecto, `ErrDetectionSprintFromAnotherProject`. La historia no necesita estar asignada a ese sprint. Un defecto detectado después de cerrar un sprint se registra en el sprint abierto actual.
- **CA-29.7** Registrar un defecto no modifica la historia ni el sprint.
- **CA-29.8** Modificar el defecto devuelto por `Register` no altera el defecto guardado.
- **CA-29.9** Varios registros simultáneos se guardan todos, con IDs únicos y sin condiciones de carrera (`go test -race`). Un cierre simultáneo no permite guardar en un sprint ya cerrado: si `Close` gana, `Register` devuelve `sprint.ErrSprintClosed`; si `Register` gana, el defecto queda guardado.
- **CA-29.10** Si el repositorio falla al guardar, se devuelve ese error y no se guarda nada.
- **CA-29.11** Se registra un defecto desde la API (`POST /api/items/{id}/defectos`) y desde el formulario "Registrar defecto" de la interfaz web. La API responde 201 en el caso exitoso y 400, 404, 409, 413 o 500 según la tabla HTTP; la interfaz muestra el éxito o el mensaje de error. Se cubre con tests de `internal/httpapi`, no con escenarios godog.

## BDD y tests previstos

Archivo BDD: `features/HU-29-registrar-defecto.feature`, con el runner `TestFeatures_HU29RegistrarDefecto` en `internal/defect/register_features_test.go`. Cada escenario lleva su CA en el nombre:

- Registrar un defecto válido en el sprint abierto actual (CA-29.1 y CA-29.3)
- Registrar sobre una historia terminada (CA-29.5)
- Registrar sobre una historia de otro sprint (CA-29.6)
- Rechazar sprint de detección ausente o 0 (CA-29.2 y CA-29.6)
- Rechazar sprint de detección negativo (CA-29.6)
- Rechazar un sprint de detección cerrado (CA-29.6)
- Rechazar descripción vacía (CA-29.2)
- Rechazar severidad inválida (CA-29.4)
- Rechazar historia inexistente (CA-29.5)
- Rechazar sprint de detección positivo inexistente (CA-29.6)
- Rechazar sprint de detección de otro proyecto (CA-29.6)
- No modificar la historia ni el sprint (CA-29.7)
- Modificar el defecto devuelto no altera el guardado (CA-29.8)
- Error del repositorio al guardar (CA-29.10)
- Registros simultáneos (CA-29.9)
- Registrar y cerrar el sprint simultáneamente (CA-29.9)

**Tests unitarios:** `internal/defect/register_test.go` cubre CA-29.1 a CA-29.10 y los casos límite. `WithOpenSprint` ya está cubierto por HU-26 y HU-29 lo ejercita a través de `Register` (sprint inexistente, cerrado y de otro proyecto).

**Tests de contrato:** `internal/defect/repository_contract_test.go` verifica que `Save` asigna IDs correlativos desde 1, conserva todos los campos, no depende de un ID recibido y devuelve copias, con `MemoryRepository` y `PostgresRepository`.

**Tests de PostgreSQL:** `internal/defect/postgres_repository_test.go` verifica la persistencia, las claves foráneas de `item_id` y `detection_sprint_id`, y el 0 del sprint de resolución. Se omiten sin `TEST_DATABASE_URL`.

**Tests de API:** `internal/httpapi/defects_test.go` con un caso exitoso (201) y uno por código de error (400 de datos, JSON y sprint ausente, 0 o negativo; 404 de historia y de sprint positivo inexistente; 409 de sprint cerrado y de otro proyecto; 413 de cuerpo grande y 500 de repositorio). Se agrega el endpoint a `bodyEndpoints` de `server_test.go` para 413 y texto después del JSON. `internal/httpapi/web_test.go` verifica el formulario, el campo de sprint obligatorio (`required`, `min="1"`), la llamada de `app.js` y que las severidades coincidan con `defect.SeverityValues()`.

**Tests de wiring:** `cmd/server/main_test.go` verifica que con memoria se usa `*defect.MemoryRepository` y, con `TEST_DATABASE_URL`, `*defect.PostgresRepository`.

## Fuera de alcance

- Cambiar el estado del defecto y registrar el sprint de resolución (HU-30).
- Consultar o listar defectos por proyecto o sprint (HU-31).
- Editar, eliminar o reasignar un defecto.
- Registrar quién detectó el defecto (E4, HU-15 a HU-17).
- Métricas, dashboard, reportes y exportación.
- Cualquier otra pantalla o endpoint que no sea el de registrar un defecto.

## Comparación con el acta de Planning

El acta del Sprint 3 (`docs/actas/sprint-3/planning.md`, Issue #110) incluye a HU-29 con 5 Story Points, responsable Sosa Ricardo, dependiente de HU-04 ("historia del proyecto") y bloqueante de HU-30, HU-31 y HU-34 (Sprint 4). Su justificación es "Paquete nuevo; el defecto se asocia a una historia del proyecto". El tema 5 exige endpoint bajo `/api/…` y pantalla en la interfaz web; el tema 7 pide el CA en el nombre de cada escenario BDD; la Definition of Done pide "Aprobada por" y `go vet ./...` antes del commit RED.

El alcance coincide: esta SDD crea un paquete nuevo (`internal/defect`), asocia el defecto a una historia del proyecto y suma el endpoint y el formulario web. Las precisiones que el acta no detalla son:

1. Los valores de severidad (CRITICA, ALTA, MEDIA y BAJA) y los estados del defecto (el estado inicial `ABIERTO`; `EN_CORRECCION` y `RESUELTO` quedan definidos para HU-30).
2. El acta no detalla el sprint de detección; la SDD agrega que es obligatorio, está abierto y es del mismo proyecto que la historia, para que M9 del sprint sea consistente. No contradice el acta.
3. La historia puede estar `TERMINADO` y no necesita pertenecer al sprint de detección.

Ninguna de estas precisiones contradice el acta; la estimación de 5 Story Points se mantiene. La dependencia HU-04 está integrada en `main`.

## Comparación con los criterios del Issue #47 (borrador)

El Issue #47 trae tres criterios en borrador. Esta SDD los precisa y agrega CA-29.4 a CA-29.11, que no están en el Issue.

| Criterio en el Issue #47 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-29.1 Se registra un defecto con descripción, severidad, estado, historia y sprint de detección | CA-29.1, CA-29.3, CA-29.4 y CA-29.6 | Se separan el estado inicial, las severidades y las reglas del sprint de detección; ausente, 0 o negativo devuelve `ErrMissingDetectionSprint` (400) y un positivo inexistente, `sprint.ErrSprintNotFound` (404) |
| CA-29.2 Descripción, severidad e historia son obligatorios | CA-29.2 | Se agrega el sprint de detección como obligatorio; ausente, 0 o negativo devuelve `ErrMissingDetectionSprint` (400) |
| CA-29.3 Estado inicial "abierto" y sin sprint de resolución | CA-29.3 | Se fija `ABIERTO` y sprint de resolución 0; se nombran los estados de HU-30 |

**Texto definitivo para la sección "Criterios de aceptación" del Issue #47** (en lenguaje de negocio, mejora 8 del Planning; se reemplaza cuando se apruebe la SDD y con autorización de Ricardo):

```markdown
## Criterios de aceptación

- [ ] CA-29.1 Se registra un defecto con descripción, severidad, historia y sprint de detección; queda con un ID, estado abierto y sin sprint de resolución.
- [ ] CA-29.2 Descripción, severidad, historia y sprint de detección son obligatorios; si falta alguno, el registro falla y no se guarda nada.
- [ ] CA-29.3 El defecto queda en estado abierto y sin sprint de resolución; "en corrección" y "resuelto" quedan para HU-30.
- [ ] CA-29.4 La severidad es crítica, alta, media o baja; cualquier otro valor se rechaza.
- [ ] CA-29.5 La historia tiene que existir y puede estar pendiente, en curso o terminada; registrar el defecto no cambia su estado.
- [ ] CA-29.6 El sprint de detección es obligatorio: si falta, el registro falla con un mensaje de dato faltante; si se informa, tiene que existir, estar abierto y ser del mismo proyecto que la historia. La historia no necesita pertenecer a ese sprint. Un defecto que aparece después de cerrar un sprint se registra en el sprint abierto actual.
- [ ] CA-29.7 Registrar el defecto no modifica la historia ni el sprint.
- [ ] CA-29.8 Lo que se devuelve al registrar no permite cambiar el defecto guardado.
- [ ] CA-29.9 Varios registros simultáneos se guardan todos y no se pisan; un cierre simultáneo del sprint no permite guardar después del cierre.
- [ ] CA-29.10 Un error al guardar deja todo como estaba.
- [ ] CA-29.11 El defecto se registra desde la API y desde un formulario de la interfaz web, con mensajes de éxito y de error entendibles.
```

## Decisiones tomadas en la redacción de la SDD (a confirmar en la revisión)

1. **Severidades:** CRITICA, ALTA, MEDIA y BAJA, normalizadas en mayúsculas; una escala con nivel crítico cubre mejor la severidad de un defecto.
2. **Sprint de detección obligatorio y abierto.** Ausente, 0 o negativo devuelve `ErrMissingDetectionSprint` (400) antes de tocar repositorios; para un ID positivo se valida con `sprint.Manager.WithOpenSprint` (HU-26): si no existe, `sprint.ErrSprintNotFound`; cerrado devuelve `sprint.ErrSprintClosed`. Todo defecto suma en M9 de su sprint y del proyecto. Un defecto que aparece después de cerrar un sprint se registra en el sprint abierto actual.
3. **Sprint de detección del mismo proyecto.** Se compara `sprint.ProjectID` con el de la historia; si no coincide, `ErrDetectionSprintFromAnotherProject`. La historia no necesita estar asignada a ese sprint: un defecto puede detectarse en un sprint posterior al de la historia.
4. **Historia `TERMINADO`: se permite.** Los defectos suelen aparecer después de completar la historia y el Issue no pide bloquearlos; no se obliga a reabrirla.
5. **Estado y sprint de resolución no son entradas.** El estado siempre es `ABIERTO` y el sprint de resolución siempre 0; HU-30 los cambiará.
6. **Paquete nuevo y persistencia:** `internal/defect` con repositorios en memoria y PostgreSQL; tabla `defects` con claves foráneas a `backlog_items` y al sprint de detección, y el sprint de resolución en 0 sin clave foránea hasta HU-30.
7. **`WithOpenSprint` en lugar de `GetSprint`:** se reutiliza `sprint.Manager.WithOpenSprint` de HU-26; no se agrega `GetSprint` ni código nuevo en `internal/sprint`.
8. **Endpoint:** `POST /api/items/{id}/defectos`, con la historia en la ruta y la descripción, la severidad y el sprint de detección en el cuerpo.
9. **RED y GREEN:** el RED compila con stubs y falla por aserciones; el cableado real de `cmd/server` y `NewHandler` se hace en el GREEN.

HU-29: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
