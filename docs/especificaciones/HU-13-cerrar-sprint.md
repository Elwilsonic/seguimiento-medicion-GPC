# HU-13 · Cerrar sprint

> Issue del repositorio: #27

## Historia de usuario

Como Agile Enabler, quiero cerrar un sprint, para dar por terminada la iteración sin que se pueda modificar después.

## Objetivo

Cambiar el estado de un sprint de `ABIERTO` a `CERRADO` mediante `sprint.Manager.Close`, de forma atómica respecto de las operaciones que escriben sobre el sprint, conservando todos los demás datos. La historia extiende `internal/sprint` con `close.go` y agrega `Update` a `sprint.Repository`, implementado en memoria y en PostgreSQL. No crea paquetes nuevos ni modifica el esquema.

## Alcance

- **Incluye:** cerrar un sprint existente en estado `ABIERTO`; persistir únicamente el cambio de estado con `Repository.Update`; rechazar el cierre de un sprint ya `CERRADO`; garantizar que `AssignItem` y `CompleteItem` no escriban después del cierre; devolver una copia por valor; tests unitarios, BDD, de contrato y de PostgreSQL.
- **No incluye:** reabrir un sprint; consultar sprints anteriores (HU-14); mover ítems `PENDIENTE` o `EN_CURSO` al sprint siguiente (queda como historia nueva para el Sprint 3); desasignar ítems (HU-11 no lo permite y está fuera de alcance); bloquear `Update` o `ChangeStatus` de ítems del backlog asignados a un sprint cerrado (opción B descartada); cambiar nombre, fechas, Sprint Goal o proyecto al cerrar; métricas, reportes o exposición HTTP (TEC-04); migraciones de esquema.

## Contrato

### Entradas

| Operación | Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|---|
| `Close` | SprintID | entero | Sí | ID del sprint que se quiere cerrar |

Firmas:

- `Close(sprintID int) (Sprint, error)` en el `Manager` de `internal/sprint`.
- `Update(sprint Sprint) (Sprint, error)` en `sprint.Repository`, implementado por `MemoryRepository` y `PostgresRepository`.

### Salidas esperadas

- **Si el cierre es válido:** el sprint actualizado, como copia por valor, con `Status` en `CERRADO` y todos los demás campos sin cambios.
- **Si el sprint ya está `CERRADO`:** `ErrSprintClosed` y ninguna modificación.
- **Si el sprint no existe:** `ErrSprintNotFound` y ninguna modificación.
- **Si el repositorio falla:** el error del repositorio, sin considerar el sprint cerrado.
- **`Update` válido:** una copia del sprint tal como quedó guardado, con el mismo ID y los campos actualizados.
- **`Update` con ID inexistente:** `ErrSprintNotFound`, sin modificar otros sprints.

### Condiciones de error

| Operación | Condición | Error | Efecto |
|---|---|---|---|
| `Close` | No existe un sprint con ese SprintID (incluye 0 y negativos) | `ErrSprintNotFound` de `internal/sprint` ("no existe un sprint con ese id") | No se modifica nada |
| `Close` | El sprint existe pero está `CERRADO` | `ErrSprintClosed` de `internal/sprint` ("el sprint está cerrado") | No se modifica nada |
| `Close` | El repositorio falla en `GetByID` | El error que devuelve el repositorio | No se modifica nada |
| `Close` | El repositorio falla en `Update` | El error que devuelve el repositorio | El sprint no queda cerrado |
| `Update` | No existe un sprint con ese ID | `ErrSprintNotFound` | No se modifica nada |
| `AssignItem` | El sprint existe pero está `CERRADO` | `ErrSprintClosed` (regla de HU-11) | No se modifica el ítem ni el sprint |
| `CompleteItem` | El sprint existe pero está `CERRADO` | `ErrSprintClosed` (regla de HU-12) | No se modifica el ítem ni el sprint |

**Orden de validación de `Close`:** tomar el mutex del `Manager`; consultar el sprint con `GetByID`; verificar el estado; cambiar solo `Status` a `CERRADO`; persistir con `Update`. Se informa el primer error encontrado y no se escribe nada si falla una validación.

**Contrato de `Update`:** el ID identifica el sprint y no se modifica. El repositorio reemplaza los campos del sprint existente y devuelve una copia. No aplica reglas de negocio: la validación de existencia y estado la hace `Close`. En PostgreSQL, un ID inexistente se traduce a `ErrSprintNotFound`; los demás errores se envuelven como `"actualizar el sprint: %w"`.

## Reglas de negocio

1. **RN-13.1 Existencia del sprint.** El SprintID debe corresponder a un sprint existente. Los IDs 0, negativos o inexistentes devuelven `ErrSprintNotFound`.
2. **RN-13.2 Solo se cierra un sprint ABIERTO.** El único estado desde el que se puede cerrar es `ABIERTO`. Un sprint `CERRADO` no se vuelve a cerrar.
3. **RN-13.3 Doble cierre.** Cerrar un sprint ya `CERRADO` devuelve `ErrSprintClosed` y no modifica nada.
4. **RN-13.4 Solo cambia el estado.** `Close` cambia únicamente `Status` a `CERRADO`; no modifica ID, proyecto, nombre, fechas ni Sprint Goal.
5. **RN-13.5 No se reabre.** Un sprint `CERRADO` no puede volver a `ABIERTO` en esta historia. Reabrir queda fuera de alcance.
6. **RN-13.6 Un sprint cerrado no admite cambios del sprint.** Sobre un sprint `CERRADO`, `AssignItem` y `CompleteItem` devuelven `ErrSprintClosed`, como ya definen HU-11 y HU-12. No se agregan restricciones a `Update` ni a `ChangeStatus` del backlog.
7. **RN-13.7 Copia por valor.** `Close` devuelve una copia del sprint. Modificar la copia devuelta no altera el sprint guardado, porque `Sprint` no contiene slices ni mapas.
8. **RN-13.8 Error del repositorio.** Si falla la consulta o la escritura, se propaga el error y no se considera cerrado el sprint.
9. **RN-13.9 Concurrencia.** `Close`, `AssignItem` y `CompleteItem` mantienen el mutex del `Manager` de sprint durante toda la operación, incluida la escritura en el `Manager` de backlog. El orden de adquisición es siempre sprint → backlog. `backlog.Manager` nunca llama a `sprint.Manager`, así que no existe el orden inverso ni deadlock. Esto garantiza que no se asigne ni se complete un ítem después de que el sprint quedó cerrado.
10. **RN-13.10 Cierre con ítems no terminados.** Se puede cerrar un sprint con ítems `PENDIENTE` o `EN_CURSO`. El cierre no modifica los ítems: conservan su estado y su `SprintID`. Limitación conocida: esos ítems quedan asignados al sprint cerrado y no se pueden mover a otro sprint, porque HU-11 rechaza cualquier reasignación con `ErrItemAlreadyAssigned` y desasignar está fuera de alcance. Pasarlos al sprint siguiente queda como historia nueva para el Sprint 3.

## Decisión sobre entidades referenciadas

HU-13 referencia **solo el sprint**, y **se valida que exista** mediante `GetByID` con `ErrSprintNotFound`, porque el cierre modifica su estado y no tiene sentido cerrar un sprint inexistente.

No se valida el proyecto de forma directa: el `ProjectID` del sprint se validó al crearlo (HU-10) y `Close` no lo modifica. Agregar una consulta al `Manager` de proyectos sumaría una dependencia y otro mutex sin aportar ninguna regla nueva.

No se validan los ítems: `Close` no los lee ni los escribe. Las reglas sobre ítems de un sprint cerrado ya las aplican HU-11 y HU-12.

## Restricciones

- **Organización del código.** Se agrega `internal/sprint/close.go` con `Close`. Se agrega `Update` a `internal/sprint/repository.go`, `memory_repository.go` y `postgres_repository.go`. Se modifican `assign.go` y `complete.go` para sostener el mutex del sprint durante la escritura en backlog y se actualizan sus comentarios. No se crean paquetes nuevos. No se modifica el esquema: la columna `status` ya existe.
- **Rama de trabajo.** `feature/HU-13-cerrar-sprint`, creada desde main actualizado (ya incluye HU-12).
- **Constructor.** `NewManager(repo Repository, projects *project.Manager, items *backlog.Manager) *Manager` no cambia.
- **Reutilización.** `Close` usa `sprint.Repository.GetByID` y el `ErrSprintClosed` ya existente. No se agrega un error nuevo.
- **`Update` del repositorio.** El ID es inmutable. En memoria, reemplaza el sprint con ese ID y devuelve una copia; si no existe, `ErrSprintNotFound`. En PostgreSQL, hace un `UPDATE sprints SET ... WHERE id = $1 RETURNING ...`; si no hay fila, devuelve `ErrSprintNotFound`; los demás errores se envuelven con `"actualizar el sprint: %w"`. La ampliación de la interfaz obliga a agregar `Update` al `failingRepository` de `create_test.go`.
- **Concurrencia.** La regla de RN-13.9 reemplaza la restricción "nunca dos mutex a la vez" documentada en las SDD de HU-11 y HU-12. Esa restricción queda reemplazada de forma explícita por esta SDD; los comentarios de `assign.go` y `complete.go` se actualizan para describir el nuevo orden. `Create` sigue validando el proyecto antes de tomar su mutex y solo accede al repositorio de sprints. `ListCompletedItems` sigue soltando el mutex antes de listar el backlog. Costo aceptado: una escritura lenta en backlog bloquea temporalmente las demás operaciones de sprint.
- **Limitación conocida.** Los ítems `PENDIENTE` o `EN_CURSO` de un sprint cerrado no se pueden reasignar, porque HU-11 devuelve `ErrItemAlreadyAssigned` para cualquier ítem ya asignado y desasignar está fuera de alcance. Moverlos al sprint siguiente es una historia nueva del Sprint 3.
- **Errores identificables.** `ErrSprintNotFound` y `ErrSprintClosed` ya existen, son exportados y comparables con `errors.Is`. Los errores de `Update` se propagan sin reinterpretarlos.
- **Tests.** Unitarios en `internal/sprint/close_test.go`; pasos de godog en `internal/sprint/close_features_test.go`, con el runner `TestFeatures_HU13CerrarSprint`, sobre `features/HU-13-cerrar-sprint.feature`. El contrato de `Update` se agrega a `runRepositoryContract`, por lo que corre tanto en memoria como en PostgreSQL. Se agrega un test de PostgreSQL específico para `Update` en `postgres_repository_test.go`.
- **Tests de PostgreSQL.** Los tests de `Update` en PostgreSQL requieren `TEST_DATABASE_URL` y la base levantada; sin esa variable se omiten, como el resto de TEC-05. Como en el entorno local no hay Docker, los corre y verifica Cardozo o Buttini antes del merge.
- **Concurrencia verificada.** El test de CA-13.8 usa un repositorio de sprint y un repositorio de backlog instrumentados que registran el orden de las escrituras, y se ejecuta con `go test -race` en el CI (en Windows local puede no haber cgo).
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Exposición HTTP.** Es alcance de TEC-04.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Cerrar un sprint `ABIERTO` sin ítems | Queda `CERRADO`; se devuelve una copia |
| Cerrar un sprint `ABIERTO` con ítems `PENDIENTE`, `EN_CURSO` y `TERMINADO` | Queda `CERRADO`; los ítems no cambian de estado ni de `SprintID` |
| Cerrar un sprint ya `CERRADO` | `ErrSprintClosed`; el sprint no cambia |
| Cerrar un sprint inexistente, con ID 0 o negativo | `ErrSprintNotFound`; no se modifica nada |
| Dos cierres simultáneos del mismo sprint | Exactamente uno tiene éxito; el otro recibe `ErrSprintClosed`; el sprint queda `CERRADO` |
| Cierre simultáneo con `AssignItem` y `CompleteItem` | Cada llamada termina bien o con `ErrSprintClosed`; el sprint queda `CERRADO`; ninguna escritura de backlog ocurre después del cierre |
| El repositorio falla al consultar el sprint | Se devuelve ese error; el sprint no cambia |
| El repositorio falla al actualizar el sprint | Se devuelve ese error; el sprint no queda cerrado |
| `Update` con un ID inexistente | `ErrSprintNotFound`; no se modifican otros sprints |
| Modificar el sprint devuelto por `Close` | El sprint guardado no cambia |
| Asignar o completar un ítem en un sprint cerrado | `ErrSprintClosed`; el ítem no cambia |
| Crear un nuevo sprint en el mismo proyecto después del cierre | Permitido: ya no hay un sprint `ABIERTO` en el proyecto (RN-10.5) |

## Criterios de aceptación

- **CA-13.1** Se cierra un sprint existente en estado `ABIERTO`: queda `CERRADO`, se devuelve una copia y su ID, proyecto, nombre, fechas y Sprint Goal no cambian. El proyecto queda sin sprint abierto, por lo que se puede crear uno nuevo según RN-10.5.
- **CA-13.2** Un sprint `CERRADO` no admite cambios: `AssignItem` y `CompleteItem` devuelven `ErrSprintClosed` y no modifican el ítem ni el sprint.
- **CA-13.3** Cerrar un sprint ya `CERRADO` devuelve `ErrSprintClosed` y no modifica nada.
- **CA-13.4** Un SprintID inexistente (incluidos 0 y negativos) devuelve `ErrSprintNotFound` y no modifica nada.
- **CA-13.5** Si el repositorio falla al consultar o actualizar, se devuelve ese error y el sprint no queda cerrado.
- **CA-13.6** Modificar el sprint devuelto por `Close` no altera el sprint guardado.
- **CA-13.7** Se puede cerrar un sprint con ítems `PENDIENTE` o `EN_CURSO`: el sprint queda `CERRADO` y los ítems conservan su estado y su `SprintID`. Esos ítems no se pueden mover a otro sprint (`ErrItemAlreadyAssigned` de HU-11); pasarlos al sprint siguiente queda fuera de alcance.
- **CA-13.8** `Close` en paralelo con `AssignItem` y `CompleteItem`: cada llamada termina bien o con `ErrSprintClosed`, el sprint queda `CERRADO` y ninguna escritura de backlog (asignación o completado) ocurre después de la escritura de cierre. Se verifica con `go test -race`.

## BDD y tests previstos

**Archivo BDD:** `features/HU-13-cerrar-sprint.feature`, con el runner `TestFeatures_HU13CerrarSprint` en `internal/sprint/close_features_test.go`. Cada escenario lleva su CA en el nombre:

- Cerrar un sprint abierto (CA-13.1)
- Cerrar un sprint sin ítems (CA-13.1)
- Crear un nuevo sprint después del cierre (CA-13.1)
- Cerrar un sprint con ítems pendientes y en curso (CA-13.7)
- Un sprint cerrado no admite asignaciones ni completados (CA-13.2)
- Cerrar un sprint ya cerrado (CA-13.3)
- Cerrar un sprint inexistente (CA-13.4)
- Error del repositorio al cerrar (CA-13.5)
- Modificar el sprint devuelto no altera el guardado (CA-13.6)
- Cierre simultáneo con asignación y completado (CA-13.8)

**Tests unitarios:** `internal/sprint/close_test.go` cubre CA-13.1 a CA-13.8 y los casos límite, incluido el cierre con ítems no terminados y el nuevo sprint después del cierre. El test de concurrencia de CA-13.8 usa repositorios instrumentados que registran el orden de las escrituras.

**Tests de contrato:** `internal/sprint/repository_contract_test.go` incorpora a `runRepositoryContract` los casos de `Update`: actualiza un sprint existente y devuelve la copia; conserva el ID; no altera otros sprints; un ID inexistente devuelve `ErrSprintNotFound`. El contrato corre con `MemoryRepository` y con `PostgresRepository`.

**Tests de PostgreSQL:** `internal/sprint/postgres_repository_test.go` agrega un test específico de `Update` (persistencia del estado, fechas normalizadas a UTC y error para ID inexistente). Se ejecuta con `TEST_DATABASE_URL` y la base levantada; sin esa variable se omite. La verificación antes del merge la hace Cardozo o Buttini.

## Fuera de alcance

- Reabrir un sprint cerrado.
- Consultar sprints anteriores (HU-14).
- Mover ítems no terminados al sprint siguiente: historia nueva para el Sprint 3.
- Desasignar ítems o reasignarlos.
- Bloquear `Update` o `ChangeStatus` de ítems del backlog por pertenecer a un sprint cerrado.
- Modificar nombre, fechas, Sprint Goal o proyecto al cerrar.
- Métricas, reportes, dashboard y exposición HTTP (TEC-04).

## Comparación con el acta de Planning

El acta del Sprint 2 (`docs/actas/sprint-2/planning.md`) describe HU-13 como "Cerrar sprint" (Issue #27, 3 Story Points, prioridad ALTA, responsable García Santiago), dependiente de HU-12, con el alcance "cambio de estado del sprint y bloqueo de cambios en un sprint cerrado". **El alcance coincide:** esta SDD cambia el estado a `CERRADO` y bloquea las operaciones que escriben sobre el sprint (asignar y completar), que es la interpretación confirmada de "no admite cambios". La estimación de 3 Story Points se mantiene. No hay contradicción con el acta; la dependencia de HU-12 ya está integrada en main.

## Comparación con los criterios del Issue #27 (borrador)

| CA en el Issue #27 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-13.1 Un sprint abierto puede cerrarse | CA-13.1 | Se precisa que solo cambia el estado, se devuelve una copia y el proyecto queda libre para un nuevo sprint |
| CA-13.2 Un sprint cerrado no admite cambios | CA-13.2 y CA-13.7 | Se limita a asignar y completar ítems; los ítems no terminados quedan como están y no se reasignan |
| CA-13.3 Cerrar un sprint ya cerrado devuelve error | CA-13.3 | Se identifica `ErrSprintClosed` y se exige que no modifique nada |
| — | CA-13.4 | Validación de existencia, incluidos 0 y negativos |
| — | CA-13.5 | Error del repositorio propagado |
| — | CA-13.6 | Copia por valor |
| — | CA-13.7 | Cierre con ítems `PENDIENTE` o `EN_CURSO` y limitación conocida |
| — | CA-13.8 | Concurrencia verificada con `-race` |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #27** (cuando se apruebe la SDD):

- [ ] CA-13.1 Se cierra un sprint existente en estado ABIERTO: queda CERRADO, se devuelve una copia y su ID, proyecto, nombre, fechas y Sprint Goal no cambian. El proyecto queda sin sprint abierto, por lo que se puede crear uno nuevo según RN-10.5.
- [ ] CA-13.2 Un sprint CERRADO no admite cambios: `AssignItem` y `CompleteItem` devuelven `ErrSprintClosed` y no modifican el ítem ni el sprint.
- [ ] CA-13.3 Cerrar un sprint ya CERRADO devuelve `ErrSprintClosed` y no modifica nada.
- [ ] CA-13.4 Un SprintID inexistente (incluidos 0 y negativos) devuelve `ErrSprintNotFound` y no modifica nada.
- [ ] CA-13.5 Si el repositorio falla al consultar o actualizar, se devuelve ese error y el sprint no queda cerrado.
- [ ] CA-13.6 Modificar el sprint devuelto por `Close` no altera el sprint guardado.
- [ ] CA-13.7 Se puede cerrar un sprint con ítems PENDIENTE o EN_CURSO: el sprint queda CERRADO y los ítems conservan su estado y su SprintID. Esos ítems no se pueden mover a otro sprint (`ErrItemAlreadyAssigned` de HU-11); pasarlos al sprint siguiente queda fuera de alcance.
- [ ] CA-13.8 `Close` en paralelo con `AssignItem` y `CompleteItem`: cada llamada termina bien o con `ErrSprintClosed`, el sprint queda CERRADO y ninguna escritura de backlog ocurre después de la escritura de cierre. Se verifica con `go test -race`.

## Decisiones confirmadas por el Agile Enabler el 20/10

1. **Alcance A.** "No admite cambios" significa que no se asignan ni se completan ítems en un sprint `CERRADO` (HU-11 y HU-12 ya devuelven `ErrSprintClosed`) y que un sprint cerrado no se vuelve a cerrar. No se tocan `Update` ni `ChangeStatus` del backlog.
2. **Cierre con ítems no terminados.** Se puede cerrar con ítems `PENDIENTE` o `EN_CURSO`; el sprint se cierra y los ítems quedan como están, asignados al sprint cerrado. Limitación conocida: no se pueden reasignar (`ErrItemAlreadyAssigned` de HU-11) y desasignar está fuera de alcance. Pasarlos al sprint siguiente es una historia nueva para el Sprint 3.
3. **Concurrencia.** `Close`, `AssignItem` y `CompleteItem` mantienen el mutex del `Manager` de sprint durante toda la operación, incluida la escritura en backlog. El orden es siempre sprint → backlog; `backlog` nunca llama a `sprint`, así que no hay deadlock. Esta regla reemplaza la de "nunca dos mutex a la vez" de HU-11 y HU-12.
4. **Implementación del cambio de concurrencia.** Se actualizan los comentarios de `assign.go` y `complete.go`. Se agrega CA-13.8 y un test con `-race` que verifica que ninguna escritura de backlog ocurre después del cierre.
5. **Errores.** Se reutiliza `ErrSprintClosed` para el doble cierre; no se agrega un error nuevo.
6. **Repositorio.** Se agrega `Update(sprint Sprint) (Sprint, error)` a `sprint.Repository`, implementado en memoria y PostgreSQL, sin migración porque la columna `status` ya existe.
7. **Nombres y rama.** `Close` en `internal/sprint/close.go`; runner `TestFeatures_HU13CerrarSprint`; `features/HU-13-cerrar-sprint.feature`; rama `feature/HU-13-cerrar-sprint`.
8. **BDD.** Cada escenario lleva el CA que cubre en el nombre.
9. **PostgreSQL.** Los tests de `Update` se corren con la base levantada y `TEST_DATABASE_URL`; los verifica Cardozo o Buttini antes del merge porque en el entorno local no hay Docker.
