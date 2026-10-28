# HU-30 · Cambiar el estado de un defecto

> Issue del repositorio: #48

Aprobada por: Peñalbé Hernán (Agile Enabler), 27/10/2026, con tres ajustes (RESUELTO → RESUELTO sin sprint de resolución da ErrMissingResolutionSprint; ErrResolutionBeforeDetection es una validación defensiva; las transiciones están en status.go y StatusValues() en defect.go), que se aplicaron.

## Historia de usuario

Como Product Builder, quiero cambiar el estado de un defecto y registrar en qué sprint se resolvió, para reflejar si está abierto, en corrección o resuelto.

## Objetivo

Cambiar el estado de un defecto existente entre `ABIERTO`, `EN_CORRECCION` y `RESUELTO` (definidos por HU-29 en `internal/defect/defect.go`), con transiciones de un paso como en HU-08. Al pasar a `RESUELTO` se registra el sprint de resolución, que tiene que existir, estar abierto, ser del mismo proyecto que el defecto y no ser anterior al sprint de detección. Al dejar `RESUELTO`, el sprint de resolución vuelve a 0.

Es el dato de entrada de **M10 (defectos resueltos)** de `docs/metricas.md` (TEC-03): un defecto en `RESUELTO` suma en M10 de su sprint de resolución y en M10 del proyecto. Desbloquea HU-47.

Se expone en `PATCH /api/defectos/{id}/estado` y en un formulario de la interfaz web (tema 5 del Planning del Sprint 3). Persiste en memoria y en PostgreSQL (TEC-05).

## Alcance

- **Incluye:** la operación `Manager.ChangeStatus`, la tabla de transiciones, la validación del sprint de resolución, `GetByID` y `Update` en la interfaz `Repository` (memoria y PostgreSQL), `StatusValues()`, el endpoint HTTP y el formulario web; tests unitarios, BDD, de contrato, de PostgreSQL, de API y de interfaz.
- **No incluye:** consultar o listar defectos (HU-31); editar la descripción, la severidad, la historia o el sprint de detección; eliminar defectos; calcular M9 o M10 (HU-47); historial de cambios de estado; registrar quién resolvió el defecto (E4); modificar la historia o el sprint.

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | ID del defecto cuyo estado se cambia |
| Estado | texto | Sí | Estado destino: `ABIERTO`, `EN_CORRECCION` o `RESUELTO`; se acepta sin distinguir mayúsculas ni contar los espacios de los extremos |
| Sprint de resolución | entero | Solo si el destino es `RESUELTO` | Sprint abierto en el que se resolvió; con otro destino se ignora |

Firma: `ChangeStatus(id int, status string, resolutionSprintID int) (Defect, error)` en el `Manager` de `internal/defect`.

### Estados (definición única)

En `internal/defect/defect.go` se agrega, con el mismo patrón que `severityValues` y `SeverityValues()` de HU-29:

- `statusValues = [...]string{"ABIERTO", "EN_CORRECCION", "RESUELTO"}`: definición única, no exportada.
- `StatusValues() []string`: devuelve una **lista nueva** con los estados en ese orden, para que la interfaz web no repita la lista. Cambiar la lista devuelta no altera la validación.

Se actualiza el comentario de `Status`, que hoy dice que `EN_CORRECCION` y `RESUELTO` "quedan definidos para HU-30", para que remita a `status.go` para las transiciones. `statusValues` y `StatusValues()` quedan en `defect.go`; las transiciones están en `status.go`.

### Salidas esperadas

- **Si el cambio es válido:** el defecto con el estado nuevo, como **copia por valor**:
  - a `RESUELTO`: con el sprint de resolución recibido;
  - desde `RESUELTO` a `EN_CORRECCION`: con el sprint de resolución en 0;
  - entre `ABIERTO` y `EN_CORRECCION`: con el sprint de resolución en 0 (ya lo estaba).

  Descripción, severidad, historia, sprint de detección e ID no cambian.
- **Si el estado destino es igual al actual:** el defecto sin cambios, como copia, sin error y sin llamar a `Update` (RN-30.4).
- **Si el cambio no es válido:** el error que indica el motivo (ver "Condiciones de error"); el defecto guardado queda exactamente como estaba.

### HTTP (TEC-04, tema 5 del Planning del Sprint 3)

`PATCH /api/defectos/{id}/estado`, con el mismo patrón que `PATCH /api/items/{id}/estado` (HU-08).

- **Cuerpo:**

  ```json
  {"estado": "RESUELTO", "sprint_resolucion_id": 5}
  ```

  `sprint_resolucion_id` solo es obligatorio cuando `estado` es `RESUELTO`: si falta, es 0 o es negativo, la API responde 400; con otro estado se ignora.

- **Respuesta exitosa:** `200 OK` con el defecto, con la misma forma que la respuesta de HU-29 (`toDefectResponse`):

  ```json
  {"id": 1, "descripcion": "El total no descuenta el descuento", "severidad": "ALTA", "estado": "RESUELTO", "item_id": 4, "sprint_deteccion_id": 3, "sprint_resolucion_id": 5}
  ```

- **Errores:**

| Caso | Código | Mensaje |
|---|---|---|
| Estado vacío o fuera de ABIERTO, EN_CORRECCION y RESUELTO | 400 | El de `ErrInvalidStatus` |
| Destino `RESUELTO` con sprint de resolución ausente, 0 o negativo | 400 | El de `ErrMissingResolutionSprint` |
| Cuerpo que no es JSON, `sprint_resolucion_id` que no es entero, o texto después del JSON | 400 | El de `errInvalidJSON` |
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| Defecto inexistente | 404 | El de `ErrDefectNotFound` |
| Sprint de resolución positivo inexistente | 404 | El de `sprint.ErrSprintNotFound` |
| Transición no permitida | 409 | El de `ErrInvalidTransition` |
| Sprint de resolución `CERRADO` | 409 | El de `sprint.ErrSprintClosed` |
| Sprint de resolución de otro proyecto | 409 | El de `ErrResolutionSprintFromAnotherProject` |
| Sprint de resolución anterior al de detección | 409 | El de `ErrResolutionBeforeDetection` (validación defensiva: no se puede producir desde la API, ver RN-30.6) |
| Cuerpo de más de 1 MiB | 413 | El de `errBodyTooLarge` |
| Falla del repositorio | 500 | El de `errInternal`, sin detalle |

Los campos que no son `estado` ni `sprint_resolucion_id` se ignoran. El orden de los errores del handler es: ID de la ruta, JSON y `Manager`.

### Interfaz web (TEC-04, tema 5)

En la sección **Defectos** de `internal/web`, debajo de "Registrar defecto", un formulario **"Cambiar estado de defecto"**:

- Campos: Defecto (ID, `required`, `min="1"`), Estado (`select` con ABIERTO, EN_CORRECCION y RESUELTO) y Sprint de resolución (ID, `min="1"`, opcional en el formulario: la regla de obligatoriedad la aplica la API).
- Llama a `PATCH /defectos/{id}/estado` y envía `sprint_resolucion_id` solo si el campo tiene valor.
- Al terminar, muestra "Defecto N ahora está en ESTADO." (con "resuelto en el sprint S" si es `RESUELTO`) o el mensaje de error de la API, con el mismo mecanismo de acción y mensaje que los demás formularios.
- No agrega un listado de defectos: eso es HU-31.
- La lista de estados de `app.js` se compara por test con `defect.StatusValues()`, para que no se desincronice.

## Reglas de negocio

1. **RN-30.1 Estados válidos.** Los únicos estados son `ABIERTO`, `EN_CORRECCION` y `RESUELTO` (RN-29.3). El estado destino se normaliza con `strings.ToUpper(strings.TrimSpace(...))`; si no es uno de los tres, `ErrInvalidStatus`, antes de consultar el repositorio.
2. **RN-30.2 Transiciones permitidas.** Un paso por vez, hacia adelante o hacia atrás, como HU-08:
   - `ABIERTO` → `EN_CORRECCION`
   - `EN_CORRECCION` → `RESUELTO`
   - `EN_CORRECCION` → `ABIERTO` (volver atrás)
   - `RESUELTO` → `EN_CORRECCION` (reabrir)

   No están permitidas `ABIERTO` → `RESUELTO` ni `RESUELTO` → `ABIERTO`: el defecto tiene que pasar por `EN_CORRECCION`. Una transición no permitida devuelve `ErrInvalidTransition`.
3. **RN-30.3 Defecto existente.** El defecto tiene que existir; si no, `ErrDefectNotFound` (incluidos los IDs 0 y negativos).
4. **RN-30.4 Mismo estado.** Si el estado destino es igual al actual, la operación es un no-op válido: no llama a `Update`, no devuelve error y devuelve una copia del defecto. El no-op aplica solo después de las validaciones previas al repositorio (RN-30.1 y RN-30.5): si el destino es `RESUELTO` y el sprint de resolución falta, es 0 o es negativo, se devuelve `ErrMissingResolutionSprint` aunque el defecto ya esté `RESUELTO`, porque el sprint <= 0 se valida antes de leer el defecto. En `RESUELTO` → `RESUELTO` con un sprint de resolución válido (> 0), distinto o no del guardado, no se cambia el sprint de resolución (para corregirlo hay que reabrir y volver a resolver).
5. **RN-30.5 Sprint de resolución obligatorio al resolver.** Si el destino es `RESUELTO`, el sprint de resolución ausente, 0 o negativo devuelve `ErrMissingResolutionSprint`, antes de consultar el repositorio. Con otro destino, el valor recibido se ignora.
6. **RN-30.6 Sprint de resolución válido.** Al pasar a `RESUELTO`:
   - el sprint tiene que existir; si no, `sprint.ErrSprintNotFound`;
   - tiene que estar `ABIERTO`; si está `CERRADO`, `sprint.ErrSprintClosed`. Se usa `sprint.Manager.WithOpenSprint`, como HU-26 y HU-29;
   - tiene que ser del mismo proyecto que el defecto (el de su historia); si no, `ErrResolutionSprintFromAnotherProject`;
   - no puede ser anterior al sprint de detección (CA-30.2); si lo es, `ErrResolutionBeforeDetection`. Puede ser el mismo sprint de detección. "Anterior" se compara por ID: dentro de un proyecto los IDs siguen el orden de creación de los sprints, porque hay un solo sprint abierto a la vez (HU-10 y regla de M3 en `docs/metricas.md`). Es una **validación defensiva**: desde la API no se puede producir, porque el único sprint abierto de un proyecto siempre es el más nuevo y el defecto se detectó en un sprint que estaba abierto. Se prueba guardando el defecto con `Save` del repositorio en memoria, con un sprint de detección posterior.
7. **RN-30.7 Sprint de resolución al reabrir.** Al pasar de `RESUELTO` a `EN_CORRECCION`, el sprint de resolución vuelve a 0: solo un defecto `RESUELTO` tiene sprint de resolución. Es coherente con RN-29.4 y con HU-47, que cuenta como resuelto solo un defecto en `RESUELTO`.
8. **RN-30.8 Solo cambia el estado y el sprint de resolución.** No se modifican la descripción, la severidad, la historia, el sprint de detección ni el ID del defecto, ni la historia ni los sprints.
9. **RN-30.9 Persistencia.** Se guarda con `Repository.Update`, con el defecto completo leído con `GetByID`; funciona igual en memoria y en PostgreSQL, sin cambios de esquema.
10. **RN-30.10 Copias por valor.** `ChangeStatus`, `GetByID` y `Update` devuelven copias. `Defect` no tiene slices ni mapas, así que copiar el struct no comparte estado con lo guardado.
11. **RN-30.11 Concurrencia y atomicidad con el cierre.** `ChangeStatus` toma el mutex del `Manager` de defectos durante la lectura, la validación y la escritura (como HU-08 y HU-29). Al resolver, guarda dentro de `WithOpenSprint`, así un cierre simultáneo del sprint de resolución no permite resolver en un sprint ya cerrado: si `Close` gana, `ErrSprintClosed`; si `ChangeStatus` gana, el defecto queda resuelto y después se cierra el sprint. El orden de locks es defect → sprint → backlog, el mismo de HU-29.
12. **RN-30.12 API e interfaz.** La operación se expone en `PATCH /api/defectos/{id}/estado` y en un formulario web, con un caso exitoso y un test por código de error.
13. **RN-30.13 M10.** Un defecto `RESUELTO` suma en M10 de su sprint de resolución y en M10 del proyecto; al reabrirlo deja de sumar. El cálculo lo hace HU-47.

## Decisión sobre entidades referenciadas

- **Defecto:** **sí se valida su existencia** (RN-30.3), porque la operación modifica esa entidad: sin la validación, cambiar el estado de un defecto inexistente sería un no-op silencioso (CA-30.4). Se agrega `GetByID` al repositorio con `ErrDefectNotFound`.
- **Sprint de resolución:** **sí se valida** cuando el destino es `RESUELTO` (RN-30.5 y RN-30.6): existencia y estado `ABIERTO` con `WithOpenSprint`, proyecto y orden respecto del sprint de detección. Con otro destino no se recibe ni se valida.
- **Historia:** **sí se consulta**, solo al resolver, con `backlog.Manager.GetItem`, para obtener el proyecto del defecto y compararlo con el del sprint de resolución, como hace HU-29 con el de detección. No se valida de nuevo su existencia por sí misma: la historia existía al registrar el defecto y el sistema no elimina ítems; si `GetItem` devolviera `ErrItemNotFound`, se propaga.
- **Sprint de detección:** **no se consulta.** Se compara por ID con el de resolución (RN-30.6); ya fue validado al registrar (HU-29) y no cambia.
- **Proyecto:** **no se valida de forma directa.** Se deriva de la historia, como en HU-29.

## Restricciones

- **Organización del código.**
  - `internal/defect/status.go` (nuevo): `ChangeStatus`, la tabla de transiciones y su validación.
  - `internal/defect/defect.go`: `statusValues`, `StatusValues()` y los errores nuevos; el comentario de `Status` remite a `status.go` para las transiciones.
  - `internal/defect/repository.go`: suma `GetByID(id int) (Defect, error)` y `Update(Defect) (Defect, error)`. `memory_repository.go` y `postgres_repository.go` los implementan. No se toca `register.go`.
  - `internal/httpapi`: `defects.go` (handler `changeDefectStatus` y su request) y `server.go` (ruta y errores nuevos en `statusFor`).
  - `internal/web`: `index.html` (formulario "Cambiar estado de defecto") y `app.js` (llamada a `/defectos/{id}/estado` y lista de estados).
  - `README.md`: la ruta en la tabla de endpoints, la línea de "Defectos" del uso de la interfaz y la fila de HU-30 en "Historias por sprint".
  - `docs/evidencias/HU-30-tdd.md`: evidencia TDD (paso 7).
  - `docs/reportes/cobertura.md`: cobertura general (paso 7).
  - `docs/ia/registro.md`: registro del uso de la IA en la historia (paso 7).
- **Sin cambios de esquema.** `defects.status` y `defects.resolution_sprint_id` ya existen (HU-29); `resolution_sprint_id` sigue sin clave foránea porque 0 no es un sprint. No hace falta recrear la base.
- **Errores identificables.** Nuevos, exportados y comparables con `errors.Is`:
  - `ErrDefectNotFound` ("no existe un defecto con ese id");
  - `ErrInvalidStatus` ("el estado debe ser ABIERTO, EN_CORRECCION o RESUELTO");
  - `ErrInvalidTransition` ("la transición de estado del defecto no está permitida");
  - `ErrMissingResolutionSprint` ("el sprint de resolución es obligatorio para resolver un defecto");
  - `ErrResolutionSprintFromAnotherProject` ("el sprint de resolución pertenece a otro proyecto");
  - `ErrResolutionBeforeDetection` ("el sprint de resolución no puede ser anterior al de detección").

  Se propagan `sprint.ErrSprintNotFound`, `sprint.ErrSprintClosed` y los errores del repositorio.
- **Concurrencia.** Mutex del `Manager` y `WithOpenSprint` (RN-30.11), verificado con `go test -race` en el CI.
- **RED y GREEN.** En el RED se agregan solo los errores, `StatusValues` y firmas vacías de `ChangeStatus`, `GetByID` y `Update` para que todo compile; los tests fallan por aserciones, no por compilación. La ruta y el handler `changeDefectStatus` van en el GREEN: sin la ruta, los tests de API fallan igual por aserción (404 o 405 en lugar de 200). Antes del commit RED se corre `go vet ./...`.
- **Tests.**
  - Unitarios en `internal/defect/status_test.go`.
  - godog en `internal/defect/status_features_test.go`, con el runner `TestFeatures_HU30CambiarEstadoDefecto`, sobre `features/HU-30-cambiar-estado-defecto.feature`, con el CA en el nombre de cada escenario.
  - Contrato de `GetByID` y `Update` en `internal/defect/repository_contract_test.go` (memoria y PostgreSQL); específicos de PostgreSQL en `postgres_repository_test.go`, que se omiten sin `TEST_DATABASE_URL`.
  - API en `internal/httpapi/defects_test.go` (200 y un caso por 400, 404, 409, 413 y 500), ruta y `bodyEndpoints` en `server_test.go`; interfaz en `web_test.go`.
  - El caso "sprint de resolución anterior al de detección" es una validación defensiva (RN-30.6): no se puede armar con `Register` + `ChangeStatus` ni desde la API (el único sprint abierto de un proyecto siempre es el más nuevo), así que el test unitario y el escenario godog guardan el defecto directamente con `Save` del repositorio en memoria, con un sprint de detección posterior, como hacen los tests de HU-47. Los tests de API no lo cubren: el 409 se prueba con los otros casos.
- **Tests de PostgreSQL.** Los corre y verifica otro integrante antes del merge si el entorno local no tiene Docker.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| `ABIERTO` → `EN_CORRECCION` | Queda `EN_CORRECCION`, sprint de resolución 0 |
| `EN_CORRECCION` → `RESUELTO` con el sprint abierto del proyecto | Queda `RESUELTO` con ese sprint de resolución |
| Resolver en el mismo sprint de detección | Se permite |
| Resolver en un sprint posterior al de detección | Se permite |
| `EN_CORRECCION` → `ABIERTO` | Queda `ABIERTO` |
| `RESUELTO` → `EN_CORRECCION` (reabrir) | Queda `EN_CORRECCION` y el sprint de resolución vuelve a 0 |
| `ABIERTO` → `RESUELTO` | `ErrInvalidTransition`; no cambia |
| `RESUELTO` → `ABIERTO` | `ErrInvalidTransition`; no cambia |
| Mismo estado (incluido `RESUELTO` → `RESUELTO` con otro sprint válido, > 0) | Sin error, sin `Update`; el defecto no cambia |
| `RESUELTO` → `RESUELTO` sin sprint de resolución (0, negativo o ausente) | `ErrMissingResolutionSprint` (se valida antes de leer el defecto); no cambia |
| Estado `en_correccion ` o ` Resuelto` | Se normaliza |
| Estado vacío, con solo espacios o `CERRADO` | `ErrInvalidStatus` |
| Estado inválido sobre un defecto inexistente | `ErrInvalidStatus` (se valida primero el estado) |
| Defecto inexistente, ID 0 o negativo | `ErrDefectNotFound` |
| `RESUELTO` con sprint de resolución 0 o negativo | `ErrMissingResolutionSprint`, sin consultar el repositorio |
| `RESUELTO` con sprint ausente sobre un defecto inexistente | `ErrMissingResolutionSprint` (se valida antes del repositorio) |
| `EN_CORRECCION` o `ABIERTO` con un sprint de resolución informado | Se ignora el sprint; el sprint de resolución queda en 0 |
| `RESUELTO` con sprint positivo inexistente | `sprint.ErrSprintNotFound`; no cambia |
| `RESUELTO` con sprint `CERRADO` | `sprint.ErrSprintClosed`; no cambia |
| `RESUELTO` con sprint de otro proyecto | `ErrResolutionSprintFromAnotherProject`; no cambia |
| `RESUELTO` con sprint anterior al de detección | `ErrResolutionBeforeDetection`; no cambia (validación defensiva: solo se arma guardando el defecto con `Save`) |
| Transición no permitida y sprint inválido a la vez (`ABIERTO` → `RESUELTO` con sprint cerrado) | `ErrInvalidTransition` (la transición se valida antes que el sprint) |
| Modificar el defecto devuelto | El defecto guardado no cambia |
| Cambios simultáneos de `ABIERTO` a `EN_CORRECCION` sobre el mismo defecto | Ninguno falla; el primero cambia, los demás son no-op |
| Resolver y cerrar el sprint simultáneamente | Si `Close` gana, `ErrSprintClosed`; si `ChangeStatus` gana, queda resuelto |
| El repositorio falla al buscar o al actualizar | Se devuelve ese error; el estado no cambia |
| La historia, el sprint de detección y el de resolución después del cambio | No cambian |
| HTTP: `sprint_resolucion_id` decimal o texto | 400 con el mensaje de JSON inválido |
| HTTP: `{id}` no numérico | 400 con el mensaje de ID inválido |
| Interfaz: estados del `select` | Iguales a `defect.StatusValues()` |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Estado vacío o fuera de la lista (RN-30.1) | `ErrInvalidStatus` | No cambia nada |
| Destino `RESUELTO` sin sprint de resolución, 0 o negativo (RN-30.5) | `ErrMissingResolutionSprint` | No cambia nada |
| Defecto inexistente (RN-30.3) | `ErrDefectNotFound` | No cambia nada |
| Transición no permitida (RN-30.2) | `ErrInvalidTransition` | No cambia nada |
| Sprint de resolución inexistente (RN-30.6) | `sprint.ErrSprintNotFound` | No cambia nada |
| Sprint de resolución `CERRADO` (RN-30.6) | `sprint.ErrSprintClosed` | No cambia nada |
| Sprint de resolución de otro proyecto (RN-30.6) | `ErrResolutionSprintFromAnotherProject` | No cambia nada |
| Sprint de resolución anterior al de detección (RN-30.6, validación defensiva) | `ErrResolutionBeforeDetection` | No cambia nada |
| El repositorio falla al buscar o al actualizar | El error del repositorio | No cambia nada |

**Orden de validación:**

1. Sin tocar repositorios: estado (`ErrInvalidStatus`) y, si el destino es `RESUELTO`, sprint de resolución <= 0 (`ErrMissingResolutionSprint`).
2. Con el mutex del `Manager` tomado: existencia del defecto (`ErrDefectNotFound`), mismo estado (no-op), transición (`ErrInvalidTransition`).
3. Si el destino es `RESUELTO`, dentro de `WithOpenSprint`: existencia y estado del sprint, proyecto (con la historia), orden respecto del sprint de detección y `Update`. Si no, `Update` directo con el sprint de resolución en 0.

Se informa el primer error encontrado.

## Criterios de aceptación

- **CA-30.1** Un defecto pasa de un estado a otro con transiciones de un paso: `ABIERTO` → `EN_CORRECCION`, `EN_CORRECCION` → `RESUELTO`, `EN_CORRECCION` → `ABIERTO` y `RESUELTO` → `EN_CORRECCION`. `ABIERTO` → `RESUELTO` y `RESUELTO` → `ABIERTO` se rechazan con `ErrInvalidTransition`. Pedir el estado actual no da error y no cambia nada.
- **CA-30.2** Al pasar a `RESUELTO` se registra el sprint de resolución, que es obligatorio (`ErrMissingResolutionSprint`), tiene que existir (`sprint.ErrSprintNotFound`), estar abierto (`sprint.ErrSprintClosed`), ser del mismo proyecto (`ErrResolutionSprintFromAnotherProject`) y no ser anterior al de detección (`ErrResolutionBeforeDetection`, validación defensiva que no se produce desde la API y se prueba con `Save`); puede ser el mismo. Al reabrir, el sprint de resolución vuelve a 0.
- **CA-30.3** Un estado vacío o fuera de `ABIERTO`, `EN_CORRECCION` y `RESUELTO` se rechaza con `ErrInvalidStatus`; el estado se acepta sin distinguir mayúsculas ni espacios de los extremos.
- **CA-30.4** Cambiar el estado de un defecto inexistente (incluidos ID 0 y negativo) devuelve `ErrDefectNotFound`.
- **CA-30.5** Un cambio rechazado no modifica nada, y un cambio válido solo modifica el estado y el sprint de resolución: ni los demás campos del defecto, ni la historia, ni los sprints.
- **CA-30.6** Los cambios quedan guardados (memoria y PostgreSQL) y modificar el defecto devuelto no altera el guardado.
- **CA-30.7** Cambios simultáneos sobre el mismo defecto no generan condiciones de carrera (`go test -race`), y un cierre simultáneo del sprint no permite resolver en un sprint ya cerrado.
- **CA-30.8** Si el repositorio falla al buscar o al actualizar, se devuelve ese error y el estado no cambia.
- **CA-30.9** El estado se cambia desde la API (`PATCH /api/defectos/{id}/estado`) y desde el formulario "Cambiar estado de defecto" de la interfaz web. La API responde 200 en el caso exitoso y 400, 404, 409, 413 o 500 según la tabla HTTP; la interfaz muestra el éxito o el mensaje de error. Se cubre con tests de `internal/httpapi`, no con escenarios godog.

## BDD y tests previstos

Archivo BDD: `features/HU-30-cambiar-estado-defecto.feature`, con el runner `TestFeatures_HU30CambiarEstadoDefecto`. Escenarios previstos, con su CA en el nombre:

- Pasar un defecto abierto a en corrección (CA-30.1)
- Resolver un defecto en corrección en el sprint abierto (CA-30.1 y CA-30.2)
- Resolver en un sprint posterior al de detección (CA-30.2)
- Volver un defecto en corrección a abierto (CA-30.1)
- Reabrir un defecto resuelto limpia el sprint de resolución (CA-30.1 y CA-30.2)
- Rechazar abierto a resuelto y resuelto a abierto (CA-30.1)
- Pedir el estado actual no cambia nada (CA-30.1)
- Rechazar un estado inválido (CA-30.3)
- Rechazar un defecto inexistente (CA-30.4)
- Rechazar resolver sin sprint de resolución (CA-30.2)
- Rechazar un sprint de resolución inexistente, cerrado o de otro proyecto (CA-30.2)
- Rechazar un sprint de resolución anterior al de detección (CA-30.2)
- No modificar la historia, los sprints ni los demás campos (CA-30.5)
- Modificar el defecto devuelto no altera el guardado (CA-30.6)
- Error del repositorio (CA-30.8)
- Cambios simultáneos y resolución simultánea con el cierre (CA-30.7)

El número final de escenarios y tests se cuenta de forma exacta en la evidencia.

## Fuera de alcance

- Consultar o listar defectos (HU-31).
- Calcular M9 y M10 (HU-47).
- Editar o eliminar defectos, o cambiar su historia o sprint de detección.
- Historial o auditoría de cambios de estado.
- Registrar quién resolvió el defecto (E4).

## Comparación con el acta de Planning

El acta del Sprint 4 (`docs/actas/sprint-4/planning.md`, Issue #133) incluye a HU-30 (Issue #48) con 3 Story Points, responsable Sosa Ricardo, dependiente de HU-29 (en `main`) y bloqueante de HU-31 y HU-47. Su justificación es "Como HU-08: transiciones válidas entre estados y sprint de resolución; el defecto ya existe (HU-29)". El orden de trabajo la pone primera, sin dependencias pendientes.

**El alcance coincide:** esta SDD define las transiciones como HU-08 (RN-30.2), registra el sprint de resolución (RN-30.5 a RN-30.7) y trabaja sobre el defecto de HU-29. Suma el endpoint y el formulario por el tema 5 del Planning del Sprint 3, como HU-29. No hay diferencias con el acta.

**Compatibilidad con HU-47** (SDD en la rama `feature/HU-47-defectos-detectados-resueltos`): HU-47 cuenta como resuelto solo un defecto `RESUELTO` con ese sprint de resolución (RN-47.2 y RN-47.5). Esta SDD lo respeta y, además, deja el sprint de resolución en 0 al reabrir (RN-30.7). HU-47 usa la lectura de HU-31; esta SDD agrega `GetByID` y `Update`, que no se superponen con el listado de HU-31.

## Comparación con los criterios del Issue #48 (borrador)

| CA en el Issue #48 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-30.1 Se puede pasar un defecto de un estado a otro permitido | CA-30.1 | Se definen las cuatro transiciones de un paso y el no-op |
| CA-30.2 Al pasar a "resuelto" se registra el sprint de resolución, que no puede ser anterior al de detección | CA-30.2 | Se agrega que es obligatorio, existente, abierto y del mismo proyecto, y que vuelve a 0 al reabrir |
| CA-30.3 No se aceptan estados inválidos | CA-30.3 | Se identifica `ErrInvalidStatus` y la normalización |
| CA-30.4 Cambiar el estado de un defecto inexistente devuelve error | CA-30.4 | Se identifica `ErrDefectNotFound` |
| — | CA-30.5 a CA-30.9 (nuevos) | Sin efectos colaterales, persistencia y copias, concurrencia, error del repositorio, API e interfaz |

**Texto definitivo para la sección "Criterios de aceptación" del Issue #48** (en lenguaje de negocio; se reemplaza cuando se apruebe la SDD y con autorización de Ricardo):

```markdown
## Criterios de aceptación

- [ ] CA-30.1 Un defecto avanza o retrocede de a un paso: abierto ↔ en corrección ↔ resuelto. No se puede saltar de abierto a resuelto ni de resuelto a abierto. Pedir el estado actual no cambia nada.
- [ ] CA-30.2 Al pasar a resuelto se registra el sprint de resolución: es obligatorio, tiene que estar abierto, ser del mismo proyecto y no ser anterior al de detección (puede ser el mismo). Al reabrir un defecto resuelto, el sprint de resolución se borra.
- [ ] CA-30.3 No se aceptan estados inválidos.
- [ ] CA-30.4 Cambiar el estado de un defecto inexistente devuelve error.
- [ ] CA-30.5 Un cambio rechazado no modifica nada, y uno válido solo cambia el estado y el sprint de resolución.
- [ ] CA-30.6 Los cambios quedan guardados y lo que se devuelve no permite cambiar el defecto guardado.
- [ ] CA-30.7 Cambios simultáneos no se pisan, y no se puede resolver en un sprint que se cerró al mismo tiempo.
- [ ] CA-30.8 Un error al guardar deja todo como estaba.
- [ ] CA-30.9 El estado se cambia desde la API y desde un formulario de la interfaz web, con mensajes de éxito y de error entendibles.
```

## Decisiones tomadas en la redacción de la SDD

Confirmadas por Ricardo el 27/10/2026, antes de redactar:

1. **Transiciones de un paso**, como HU-08: `ABIERTO` ↔ `EN_CORRECCION` ↔ `RESUELTO`.
2. **Sprint de resolución abierto y del mismo proyecto**, validado con `WithOpenSprint` como el de detección en HU-29.
3. **Al reabrir, el sprint de resolución vuelve a 0.**
4. **Alcance:** `Manager`, persistencia, API y formulario web.

Confirmadas por Hernán el 27/10/2026, en la revisión:

5. **"Anterior" se compara por ID** del sprint (orden de creación dentro del proyecto), no por fechas.
6. **Mismo estado = no-op**, como HU-08, incluido `RESUELTO` → `RESUELTO` con un sprint de resolución válido (no cambia el sprint de resolución). Sin sprint de resolución da `ErrMissingResolutionSprint` (RN-30.4).
7. **Con un destino que no es `RESUELTO`, el sprint de resolución recibido se ignora** (no da error), como HU-29 ignora `estado` en el cuerpo.
8. **Orden de validación:** la transición se valida antes que el sprint de resolución, y el sprint de resolución <= 0 antes de consultar el repositorio.
9. **Endpoint:** `PATCH /api/defectos/{id}/estado`, con el patrón de `PATCH /api/items/{id}/estado`.
10. **`StatusValues()`** con el mismo patrón que `SeverityValues()`, para que la interfaz no repita la lista de estados.

HU-30: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
