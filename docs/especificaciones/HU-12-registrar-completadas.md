# HU-12 · Registrar historias completadas

> Issue del repositorio: #26

## Historia de usuario

Como Product Builder, quiero marcar como completadas las historias de un sprint, para saber qué se terminó.

## Objetivo

Marcar como completado un ítem del Product Backlog asignado a un sprint `ABIERTO`, cambiando su estado a `TERMINADO`, y consultar qué ítems de un sprint están completados, para que HU-13 pueda cerrar el sprint y las métricas (E9) puedan calcular los Story Points completados.

Esta historia extiende el paquete `internal/sprint` con `complete.go` y reutiliza las reglas de estado de HU-08 y las operaciones de HU-05, HU-06 y HU-11. No crea un paquete nuevo ni agrega persistencia.

## Alcance

- **Incluye:** marcar como completado un ítem asignado a un sprint abierto a través de `sprint.Manager.CompleteItem`, validando la existencia del sprint y del ítem, la pertenencia al sprint y el estado del sprint; y consultar los ítems completados de un sprint con `sprint.Manager.ListCompletedItems`.
- **No incluye:** cerrar el sprint (HU-13), consultar sprints anteriores (HU-14), asignar o desasignar ítems (HU-11), cambiar el estado de un ítem fuera del sprint (HU-08), definir un campo de completado independiente del estado, modificar otros campos del ítem, ni la exposición HTTP (TEC-04).

## Contrato

### Entradas

| Operación | Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|---|
| `CompleteItem` | SprintID | entero | Sí | ID del sprint al que pertenece el ítem |
| `CompleteItem` | ItemID | entero | Sí | ID del ítem del backlog que se marca como completado |
| `ListCompletedItems` | SprintID | entero | Sí | ID del sprint cuyos ítems completados se consultan |

Firmas:

- `CompleteItem(sprintID, itemID int) (backlog.Item, error)` en el `Manager` de `internal/sprint`.
- `ListCompletedItems(sprintID int) ([]backlog.Item, error)` en el `Manager` de `internal/sprint`.

### Salidas esperadas

- **Si el completado es válido:** el ítem actualizado, como copia por valor (con su propia copia de la lista de criterios), con `Status` en `TERMINADO` y todos los demás campos sin cambios.
- **Si el ítem ya estaba `TERMINADO`:** una copia del ítem sin cambios, sin error y sin llamar a `Update` (no-op).
- **Si la consulta es válida:** una lista nueva, ordenada por ID ascendente, con copias de los ítems del sprint cuyo estado es `TERMINADO`. Si no hay ninguno, devuelve una lista vacía, no `nil`.
- **Si la operación no es válida:** un error que indica el motivo (ver "Condiciones de error"), y no se modifica ningún ítem.

### Condiciones de error

| Operación | Condición | Error | Efecto |
|---|---|---|---|
| Ambas | No existe un sprint con ese SprintID (incluye 0 y negativos) | `ErrSprintNotFound` de `internal/sprint` ("no existe un sprint con ese id") | No se modifica nada |
| `CompleteItem` | El sprint existe pero está `CERRADO` | `ErrSprintClosed` de `internal/sprint` ("el sprint está cerrado") | No se modifica nada |
| `CompleteItem` | No existe un ítem con ese ItemID (incluye 0 y negativos) | `ErrItemNotFound` de `internal/backlog` ("no existe un ítem del backlog con ese id"), propagado tal cual | No se modifica nada |
| `CompleteItem` | El ítem no está asignado a ese sprint o está asignado a otro sprint | `ErrItemNotAssignedToSprint` de `internal/sprint` ("el ítem no está asignado al sprint"), nuevo | No se modifica nada |
| `CompleteItem` | El ítem está `PENDIENTE` y HU-08 no permite `PENDIENTE → TERMINADO` | `ErrInvalidTransition` de `internal/backlog` ("la transición de estado no está permitida"), propagado tal cual | No se modifica nada |
| `CompleteItem` | El repositorio falla al consultar el sprint, el ítem o al actualizar | El error que devuelve el repositorio | No se modifica nada |
| `ListCompletedItems` | El proyecto del sprint no existe (inconsistencia de datos) | `ErrProjectNotFound` de `internal/project`, propagado por `ListItems` | No se devuelve ninguna lista |
| `ListCompletedItems` | El repositorio falla al consultar | El error que devuelve el repositorio | No se devuelve ninguna lista |

**Orden de validación de `CompleteItem`:** sprint (existencia), estado del sprint, ítem (existencia), pertenencia al sprint y cambio de estado con HU-08. Se informa el primer error encontrado.

**Orden de validación de `ListCompletedItems`:** sprint (existencia) y listado de los ítems del proyecto del sprint. La consulta funciona tanto para un sprint `ABIERTO` como `CERRADO`.

## Reglas de negocio

1. **RN-12.1 Completado = `TERMINADO`.** Marcar un ítem como completado significa cambiar su `Status` a `TERMINADO`, el estado que ya define HU-08. No se agrega un campo de completado separado: el estado del ítem es la única fuente de verdad y las métricas de Story Points completados pueden usar el mismo dato.
2. **RN-12.2 Transiciones de HU-08.** El cambio de estado respeta las reglas de HU-08: `EN_CURSO → TERMINADO` es válido; `PENDIENTE → TERMINADO` devuelve `ErrInvalidTransition` y el ítem queda como estaba. Un ítem `PENDIENTE` debe pasar primero a `EN_CURSO`. `TERMINADO → TERMINADO` es un no-op válido.
3. **RN-12.3 Pertenencia al sprint.** Un ítem pertenece al sprint si su `SprintID` es igual al SprintID recibido. No se compara el proyecto: los IDs de sprint ya son únicos. Un ítem sin asignar (`SprintID == 0`) o asignado a otro sprint devuelve `ErrItemNotAssignedToSprint`.
4. **RN-12.4 Sprint abierto para completar.** Solo se puede marcar un ítem como completado en un sprint `ABIERTO`. Un sprint `CERRADO` devuelve `ErrSprintClosed`. La consulta de completados sí se permite en un sprint cerrado, porque es una lectura y HU-14 necesitará consultar sprints anteriores.
5. **RN-12.5 Entidades existentes.** El sprint y el ítem tienen que existir (ver "Decisión sobre entidades referenciadas"). Se reutilizan `GetByID` de HU-11 y `GetItem` de HU-05, con sus errores, sin duplicar reglas.
6. **RN-12.6 Consulta de completados.** `ListCompletedItems` devuelve solo los ítems del sprint cuyo estado es `TERMINADO`, ordenados por ID ascendente (el orden que ya garantiza HU-06), como copias con su propia lista de criterios. Si no hay completados, devuelve una lista vacía, no `nil`. Un ítem que fue completado y luego reabierto por HU-08 deja de aparecer en la lista.
7. **RN-12.7 Solo cambia el estado.** Completar no modifica título, descripción, prioridad, criterios, horas estimadas, story points, `SprintID`, proyecto ni ID.
8. **RN-12.8 Copias por valor.** `CompleteItem` y `ListCompletedItems` devuelven copias: modificar el ítem devuelto, sus criterios o la lista devuelta no altera lo guardado.
9. **RN-12.9 Sin cambios de repositorio ni de esquema.** La historia reutiliza `sprint.Repository.GetByID`, `backlog.Manager.GetItem`, `backlog.Manager.ChangeStatus` y `backlog.Manager.ListItems`; no agrega operaciones a las interfaces `Repository` ni columnas a PostgreSQL.

## Decisión sobre entidades referenciadas

HU-12 referencia **dos** entidades por ID, y **se validan las dos**:

1. **El sprint**, porque el completado es propio de un sprint y su estado decide si se puede modificar (RN-12.4), y porque la consulta necesita el proyecto para listar sus ítems. Se reutiliza `GetByID` de HU-11 con `ErrSprintNotFound`.
2. **El ítem**, porque la operación modifica su estado y CA-12.2 exige comprobar que su `SprintID` es el del sprint. Se reutiliza `GetItem` de HU-05 con `ErrItemNotFound`.

No se valida el proyecto de forma directa: la operación se identifica por el sprint y el ítem, y `ListItems` de HU-06 valida el proyecto como parte de su contrato. Si el proyecto no existiera, `ListCompletedItems` propaga `ErrProjectNotFound`.

## Restricciones

- **Organización del código.** Se agrega `internal/sprint/complete.go` con `CompleteItem` y `ListCompletedItems`, y se agrega `ErrItemNotAssignedToSprint` a `internal/sprint/sprint.go`, junto a los demás errores del paquete. No se crean paquetes nuevos. No se modifican `internal/project`, `internal/backlog`, las interfaces `Repository`, `internal/sprint/postgres_repository.go`, `internal/backlog/postgres_repository.go` ni `db/init/01-schema.sql`.
- **Rama de trabajo.** `feature/HU-12-registrar-completadas`.
- **Constructor.** `NewManager(repo Repository, projects *project.Manager, items *backlog.Manager) *Manager` no cambia: la historia usa las mismas dependencias que HU-11.
- **Reutilización.** `CompleteItem` usa `sprint.Repository.GetByID`, `backlog.Manager.GetItem` y `backlog.Manager.ChangeStatus(itemID, string(backlog.StatusFinished))`. `ListCompletedItems` usa `sprint.Repository.GetByID` y `backlog.Manager.ListItems`, y filtra por `SprintID` y `Status == TERMINADO`. No se agrega `ListBySprint` al repositorio de backlog porque el filtrado se resuelve con la lista del proyecto que ya existe, sin ampliar la interfaz ni PostgreSQL.
- **Mutex y concurrencia.** La consulta del sprint se hace bajo el mutex de sprint y se suelta antes de llamar al `Manager` de backlog: nunca se toman dos mutex a la vez. Las escrituras de estado quedan serializadas por el mutex de backlog dentro de `ChangeStatus`. Dos marcaciones simultáneas del mismo ítem `EN_CURSO` producen una sola escritura y las demás son no-op; se verifica con `go test -race` (lo corre el CI; en Windows local no hay cgo).
- **Carrera con HU-13.** Hoy no existe ninguna operación que cierre sprints. Cuando HU-13 la agregue, deberá coordinar con `CompleteItem` la carrera entre cerrar el sprint y completar un ítem. Esta SDD valida el estado, pero no puede garantizar atomicidad contra una operación que todavía no existe.
- **Errores identificables.** `ErrItemNotAssignedToSprint` es nuevo, exportado, declarado en `internal/sprint/sprint.go` y comparable con `errors.Is`. Los demás errores se propagan tal cual desde los paquetes que los definen.
- **Tests.** Unitarios en `internal/sprint/complete_test.go`; pasos de godog en `internal/sprint/complete_features_test.go`, con el runner `TestFeatures_HU12RegistrarCompletadas`, sobre `features/HU-12-registrar-completadas.feature`. Los tests internos del paquete pueden preparar un sprint `CERRADO` guardándolo directamente en el repositorio en memoria, porque HU-13 todavía no existe.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Exposición HTTP.** Es alcance de TEC-04.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Ítem `EN_CURSO` asignado a un sprint `ABIERTO` | Queda `TERMINADO`; los demás campos no cambian |
| Ítem ya `TERMINADO` asignado al sprint | No-op sin error; no llama a `Update`; devuelve copia |
| Ítem `PENDIENTE` asignado al sprint | `ErrInvalidTransition`; sigue `PENDIENTE` |
| Ítem sin asignar (`SprintID == 0`) | `ErrItemNotAssignedToSprint` |
| Ítem asignado a otro sprint | `ErrItemNotAssignedToSprint` |
| Sprint `CERRADO` al completar | `ErrSprintClosed`; el ítem no cambia |
| Sprint `CERRADO` al consultar completados | Devuelve la lista de completados sin error |
| Sprint inexistente, 0 o negativo | `ErrSprintNotFound` |
| Ítem inexistente, 0 o negativo | `ErrItemNotFound` |
| Sprint sin ítems completados | Lista vacía, no `nil` |
| Sprint con ítems `PENDIENTE`, `EN_CURSO` y `TERMINADO` | Solo devuelve los `TERMINADO`, ordenados por ID |
| Ítem completado y luego reabierto por HU-08 | Deja de aparecer en `ListCompletedItems` |
| Ítem de otro sprint | No aparece en la lista |
| Modificar el ítem devuelto o sus criterios | El ítem guardado no cambia |
| Modificar la lista devuelta por `ListCompletedItems` | Los ítems guardados no cambian |
| Varias marcaciones simultáneas del mismo ítem `EN_CURSO` | Una sola escritura; las demás son no-op; `-race` sin condiciones de carrera |
| El repositorio falla al consultar o actualizar | Se devuelve ese error; no se modifica nada |

## Criterios de aceptación

- **CA-12.1** Se marca como completado un ítem existente, asignado a un sprint `ABIERTO` y en estado `EN_CURSO`: el ítem queda `TERMINADO`, se devuelve una copia y el resto de sus campos no cambia.
- **CA-12.2** Un ítem que no pertenece al sprint (sin asignar o asignado a otro sprint) devuelve `ErrItemNotAssignedToSprint` y no se modifica.
- **CA-12.3** `ListCompletedItems` devuelve solo los ítems del sprint con estado `TERMINADO`, ordenados por ID ascendente y como copias; si no hay ninguno, devuelve una lista vacía. La consulta funciona con el sprint abierto o cerrado.
- **CA-12.4** Completar un ítem ya `TERMINADO` es un no-op sin error, sin llamar a `Update`, y devuelve una copia.
- **CA-12.5** Un sprint inexistente (incluidos 0 y negativo) devuelve `ErrSprintNotFound`; un ítem inexistente devuelve `ErrItemNotFound`; un sprint `CERRADO` devuelve `ErrSprintClosed` al completar. En ningún caso se modifica nada.
- **CA-12.6** Un ítem `PENDIENTE` no se puede completar directamente: `PENDIENTE → TERMINADO` devuelve `ErrInvalidTransition` (HU-08) y el ítem sigue `PENDIENTE`.
- **CA-12.7** Modificar el ítem devuelto, sus criterios o la lista devuelta no altera lo guardado.
- **CA-12.8** Si el repositorio falla al consultar o actualizar, se devuelve ese error y no se modifica nada.
- **CA-12.9** Varias marcaciones simultáneas del mismo ítem `EN_CURSO` dejan el ítem `TERMINADO`, con una sola escritura y sin condiciones de carrera (verificado con `go test -race`).

## Comparación con el acta de Planning

El acta del Sprint 2 (`docs/actas/sprint-2/planning.md`) describe HU-12 como "Registrar historias completadas" (Issue #26, 3 Story Points, responsable García Santiago), dependiente de HU-11, con la nota "Si 'completado' es el estado `TERMINADO`, también HU-08 (lo decide su SDD)" y la justificación "Marca ítems del sprint y consulta los completados; su relación con el estado `TERMINADO` se define en la SDD". **El alcance coincide:** esta SDD decide que completado es `TERMINADO`, reutiliza las transiciones de HU-08 y agrega la consulta de completados. No hay diferencias entre el acta y esta SDD; la decisión que el acta deja abierta se resuelve en RN-12.1 y RN-12.2.

## Comparación con los criterios del Issue #26 (borrador)

| CA en el Issue #26 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-12.1 Se puede marcar como completado un elemento asignado al sprint | CA-12.1, CA-12.4 y CA-12.6 (precisado) | El ítem debe estar asignado a un sprint abierto y en `EN_CURSO`, porque las transiciones las define HU-08; ya `TERMINADO` es no-op |
| CA-12.2 No se puede completar un ítem que no pertenece al sprint | CA-12.2 (precisado) | Se identifica `ErrItemNotAssignedToSprint` e incluye sin asignar o asignado a otro sprint |
| CA-12.3 Se puede consultar cuáles ítems del sprint están completados | CA-12.3 (precisado) | `ListCompletedItems`, solo `TERMINADO`, orden por ID, lista vacía no `nil`, sprint abierto o cerrado |
| — | CA-12.5 (nuevo) | Validación de existencia de sprint e ítem y de sprint cerrado |
| — | CA-12.7 (nuevo) | Copias por valor |
| — | CA-12.8 (nuevo) | Error del repositorio propagado |
| — | CA-12.9 (nuevo) | Concurrencia verificada con `go test -race` |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #26** (cuando se apruebe la SDD):

- [ ] CA-12.1 Se marca como completado un ítem existente, asignado a un sprint ABIERTO y en estado EN_CURSO: el ítem queda TERMINADO, se devuelve una copia y el resto de sus campos no cambia.
- [ ] CA-12.2 Un ítem que no pertenece al sprint (sin asignar o asignado a otro sprint) devuelve `ErrItemNotAssignedToSprint` y no se modifica.
- [ ] CA-12.3 `ListCompletedItems` devuelve solo los ítems del sprint con estado TERMINADO, ordenados por ID ascendente y como copias; si no hay ninguno, devuelve una lista vacía. La consulta funciona con el sprint abierto o cerrado.
- [ ] CA-12.4 Completar un ítem ya TERMINADO es un no-op sin error, sin llamar a `Update`, y devuelve una copia.
- [ ] CA-12.5 Un sprint inexistente (incluidos 0 y negativo) devuelve `ErrSprintNotFound`; un ítem inexistente devuelve `ErrItemNotFound`; un sprint CERRADO devuelve `ErrSprintClosed` al completar. En ningún caso se modifica nada.
- [ ] CA-12.6 Un ítem PENDIENTE no se puede completar directamente: PENDIENTE → TERMINADO devuelve `ErrInvalidTransition` (HU-08) y el ítem sigue PENDIENTE.
- [ ] CA-12.7 Modificar el ítem devuelto, sus criterios o la lista devuelta no altera lo guardado.
- [ ] CA-12.8 Si el repositorio falla al consultar o actualizar, se devuelve ese error y no se modifica nada.
- [ ] CA-12.9 Varias marcaciones simultáneas del mismo ítem EN_CURSO dejan el ítem TERMINADO, con una sola escritura y sin condiciones de carrera (verificado con `go test -race`).

## Decisiones confirmadas por el Agile Enabler el 19/10

1. **Completado = estado `TERMINADO`.** No se agrega un campo nuevo: el estado del ítem es la única fuente de verdad y no hace falta cambiar el esquema ni PostgreSQL.
2. **Se respetan las transiciones de HU-08.** `PENDIENTE → TERMINADO` no está permitido; el ítem debe pasar por `EN_CURSO`. Consecuencia: CA-12.1 aplica a ítems asignados y en `EN_CURSO`.
3. **Sprint cerrado:** no se puede completar, pero sí consultar (es una lectura).
4. **Ya completado:** no-op sin error, igual que HU-08 con el mismo estado.
5. **Consulta:** `ListCompletedItems`, solo completados, ordenados por ID, lista vacía no `nil`.
6. **Sin `ListBySprint`:** se reutiliza `ListItems` del proyecto y se filtra por `SprintID`, para no ampliar las interfaces `Repository` ni los repositorios PostgreSQL.
7. **Se validan sprint e ítem.** El proyecto no se valida directamente: lo cubre el sprint y lo valida `ListItems` en la consulta.
8. **Un error nuevo:** `ErrItemNotAssignedToSprint`, en `internal/sprint/sprint.go`.
9. **Mutex:** uno por paquete, nunca anidados; la escritura de estado queda serializada por `ChangeStatus`.
10. **HTTP fuera de alcance:** la exposición es de TEC-04.
