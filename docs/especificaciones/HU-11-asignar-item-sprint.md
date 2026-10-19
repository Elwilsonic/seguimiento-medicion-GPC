# HU-11 · Asignar historias del backlog a un sprint

> Issue del repositorio: #25

## Historia de usuario

Como Agile Enabler, quiero asignar ítems del backlog a un sprint, para definir el alcance de la iteración.

## Objetivo

Asignar un ítem del Product Backlog a un sprint `ABIERTO` del mismo proyecto, guardando la asignación en el ítem (su `SprintID`), para que HU-12 (registrar historias completadas) y más adelante las métricas (E9) sepan qué historias forman el alcance de cada sprint.

La operación vive en el `Manager` del paquete `internal/sprint` (la historia es de la épica E3) y delega la escritura del ítem en el `Manager` de `internal/backlog`, que es su dueño. Es la primera historia que cruza dos paquetes, como anticipa el acta de Planning.

## Alcance

- **Incluye:** asignar un ítem existente a un sprint existente y `ABIERTO` del mismo proyecto, validando que el ítem no esté ya asignado, a través de `sprint.Manager.AssignItem`, que delega la escritura en la nueva operación `backlog.Manager.AssignToSprint`.
- **No incluye:** desasignar ni reasignar ítems (ninguna historia lo pide), registrar historias completadas (HU-12), cerrar el sprint (HU-13), consultar sprints anteriores (HU-14), asignar por lote (se asigna un ítem por vez) ni la exposición HTTP (TEC-04).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| SprintID | entero | Sí | ID del sprint al que se asigna el ítem |
| ItemID | entero | Sí | ID del ítem del backlog que se asigna |

Firmas: `AssignItem(sprintID, itemID int) (backlog.Item, error)` en el `Manager` de `internal/sprint`, y `AssignToSprint(itemID, sprintID int) (backlog.Item, error)` en el `Manager` de `internal/backlog`.

### Salidas esperadas

- **Si la asignación es válida:** el ítem actualizado, como copia por valor (con su propia copia de la lista de criterios), idéntico al guardado salvo en `SprintID`, que queda con el ID del sprint. Ni el sprint ni ningún otro campo del ítem cambian.
- **Si la asignación no es válida:** un error que indica el motivo (ver "Errores"), y no se modifica ni el ítem ni el sprint.

### Errores

| Condición | Error |
|---|---|
| No existe un sprint con ese SprintID (incluye 0 y negativos) | `ErrSprintNotFound` de `internal/sprint` ("no existe un sprint con ese id") — nuevo |
| El sprint existe pero está `CERRADO` | `ErrSprintClosed` de `internal/sprint` ("el sprint está cerrado") — nuevo |
| No existe un ítem con ese ItemID (incluye 0 y negativos) | `ErrItemNotFound` de `internal/backlog` ("no existe un ítem del backlog con ese id"), propagado tal cual (HU-05) |
| El ítem pertenece a otro proyecto | `ErrItemFromAnotherProject` de `internal/sprint` ("el ítem pertenece a otro proyecto") — nuevo |
| El ítem ya está asignado a un sprint (cualquiera) | `ErrItemAlreadyAssigned` de `internal/backlog` ("el ítem ya está asignado a un sprint") — nuevo |
| El repositorio falla al consultar o al escribir | El error que devuelve el repositorio |

**Orden de validación:** sprint, estado del sprint, ítem, mismo proyecto, ya asignado y escritura. Se informa el primer error encontrado.

## Reglas de negocio

1. **RN-11.1 Sprint existente y abierto.** El sprint tiene que existir y estar `ABIERTO` (CA-11.1 y CA-11.3). Para consultarlo, HU-11 agrega `GetByID(id int) (Sprint, error)` a la interfaz `Repository` de `internal/sprint`, con `ErrSprintNotFound` si no existe — la misma forma que `GetByID` de proyectos (HU-03) e ítems (HU-05). HU-12 y HU-14 reutilizan esta operación.
2. **RN-11.2 Mismo proyecto.** El ítem tiene que pertenecer al proyecto del sprint: `item.ProjectID == sprint.ProjectID` (CA-11.3). Los IDs de ítems son globales (RN-04.6), así que la comparación se hace con las copias ya obtenidas, sin búsquedas extra.
3. **RN-11.3 Un solo sprint por ítem.** Si el ítem ya tiene un `SprintID` distinto de 0, la asignación falla con `ErrItemAlreadyAssigned` (CA-11.2), aunque el sprint sea el mismo al que ya está asignado. No hay reasignación ni desasignación: ninguna historia las pide; si el equipo las necesita, serán historias nuevas.
4. **RN-11.4 Dónde se guarda la asignación.** En el ítem: HU-11 agrega el campo `SprintID int` a `backlog.Item`, donde 0 significa "sin asignar" (el primer ID de sprint es 1, así que el 0 es seguro — mismo criterio que `StoryPoints` en RN-04.8). El acta sugiere usar el `Update` de HU-07; esta SDD decide lo siguiente: el `Update` de HU-07 no sirve para escribir la asignación, porque reemplaza campos de texto revalidándolos y pasa a **conservar** el `SprintID` como conserva el estado y los story points (ver RN-11.5). La escritura es una operación nueva del `Manager` de backlog, `AssignToSprint`, que hace una lectura-modificación-escritura con las operaciones que ya existen en el repositorio (`GetByID` y `Update` de HU-05 y HU-07): la interfaz `Repository` de backlog no cambia.
5. **RN-11.5 Conservación.** Dos ajustes mínimos en el código existente, para que la asignación no se pierda: el `Update` de HU-07 conserva el `SprintID` (una línea, como ya conserva `Status` y `StoryPoints`), y el `Create` de HU-04 deja el `SprintID` en 0 de forma explícita, como hace con los story points. Ningún otro campo del ítem cambia al asignar: ni el estado (lo cambia HU-08), ni los story points (HU-20), ni las horas (HU-07).
6. **RN-11.6 Estado del ítem indistinto.** Un ítem en cualquier estado (`PENDIENTE`, `EN_CURSO` o `TERMINADO`) se puede asignar: no hay ninguna regla que lo impida (mismo criterio que RN-04.5 y RN-10.4).
7. **RN-11.7 Copias por valor y encapsulamiento.** `AssignItem` y `AssignToSprint` devuelven copias del ítem, con su propia copia de la lista de criterios (RN-05.3). El `SprintID` se lee y se escribe solo a través de los Managers. `AssignToSprint` solo debe llamarse desde `sprint.Manager.AssignItem`; TEC-04 expone `AssignItem`, nunca `AssignToSprint`.

## Decisión sobre entidades referenciadas

HU-11 referencia **dos** entidades por ID, y **se validan las dos**:

1. **El sprint**, porque las reglas dependen de su estado y de su proyecto (RN-11.1 y RN-11.2). Se agrega `GetByID` a `sprint.Repository` con `ErrSprintNotFound`, siguiendo el patrón de HU-03 y HU-05. Sin esta validación, se podría "asignar" a un sprint inexistente y el `SprintID` del ítem apuntaría a la nada.
2. **El ítem**, porque la asignación se escribe en él y hay que controlar que no esté ya asignado (RN-11.3). Se reutiliza `GetItem` de HU-05 con `ErrItemNotFound`, sin duplicar ni la regla ni el error.

## Restricciones

- **Organización del código.** La historia toca los dos paquetes:
  - `internal/sprint`: `assign.go` (nuevo, `AssignItem`), `sprint.go` (+`ErrSprintNotFound`, `ErrSprintClosed`, `ErrItemFromAnotherProject`), `repository.go` (+`GetByID`), `memory_repository.go` (+`GetByID`), `postgres_repository.go` (+`GetByID`), `manager.go` (constructor con el nuevo parámetro).
  - `internal/backlog`: `assign.go` (nuevo, `AssignToSprint` y `ErrItemAlreadyAssigned`), `item.go` (+`SprintID`), `update.go` (conservar `SprintID` — la única línea de código de HU-07 que se toca), `create.go` (`SprintID: 0` explícito — la única línea de HU-04), `postgres_repository.go` (+`sprint_id` en `itemColumns`, `Save`, `Update` y `scanItem`).
  - `db/init/01-schema.sql` (+`sprint_id` en `backlog_items`; `02-test-database.sql` vuelve a aplicar este esquema y no cambia).
- **Constructor.** `NewManager(repo Repository, projects *project.Manager, items *backlog.Manager) *Manager`. El nuevo parámetro obliga a actualizar los tests de HU-10 (`newTestManager` en `create_test.go` y el `Before` de godog en `create_features_test.go`), como cuando HU-06 extendió `failingRepository` de HU-04. La dependencia es en una sola dirección: `sprint` importa `backlog`; `backlog` no importa `sprint` (sin ciclos).
- **Reparto de validaciones.** `sprint.Manager.AssignItem` valida el sprint (existencia y estado, bajo su mutex, como `project.Get`), el ítem (con `GetItem`) y el proyecto (RN-11.2), y delega en `backlog.Manager.AssignToSprint`, que hace la lectura-modificación-escritura de `SprintID` bajo el mutex de backlog (patrón de RN-07.6) y aplica RN-11.3 adentro del mutex. Nunca se toman dos mutex a la vez: primero el de sprint (leer y soltar), después el de backlog. La verificación de "mismo proyecto" fuera del mutex de backlog es segura porque `ProjectID` de un ítem nunca cambia (ninguna operación lo modifica).
- **Concurrencia.** Dos asignaciones simultáneas del mismo ítem (a dos sprints, o al mismo): exactamente una gana y la otra recibe `ErrItemAlreadyAssigned`, porque la verificación de RN-11.3 y la escritura van juntas bajo el mutex. Verificado con `go test -race` (CI; en Windows local no hay cgo). La carrera con HU-13 (cerrar el sprint mientras se asigna) la tratará esa historia cuando exista; hoy no hay ninguna operación que cierre sprints.
- **Tests.** Unitarios en `internal/sprint/assign_test.go` y `internal/backlog/assign_test.go` (este último incluye el test de que el `Update` de HU-07 conserva el `SprintID`; es el único archivo de test que se extiende de HU-07), internos a cada paquete. El `failingRepository` de los tests de sprint se extiende con `GetByID`. Los pasos de godog van en `internal/sprint/assign_features_test.go` (`package sprint`, la historia es de E3 y orquesta ambos paquetes), con el runner `TestFeatures_HU11AsignarItemSprint`, sobre `features/HU-11-asignar-item-sprint.feature`.
- **TEC-05.** TEC-05 ya está integrado en `main`: los repositorios PostgreSQL implementan las mismas interfaces `Repository` que la memoria. Por eso HU-11 incluye las dos piezas que la historia agrega a los contratos también en PostgreSQL: sin `GetByID` en `internal/sprint/postgres_repository.go` el servidor no compila (`cmd/server/repositories.go` lo usa como `Repository`), y sin la columna `sprint_id` en `db/init/01-schema.sql` (leída y escrita por `internal/backlog/postgres_repository.go`) el backend PostgreSQL perdería las asignaciones. La columna es `sprint_id INTEGER NOT NULL DEFAULT 0`, **sin clave foránea** a `sprints`: el 0 significa "sin asignar" y no es un sprint, así que no puede cumplir una clave foránea. Los tests de contrato de TEC-05 se extienden con las dos piezas: `GetByID` de sprints en `internal/sprint/repository_contract_test.go` y el `SprintID` del ítem (Save, Update y GetByID) en `internal/backlog/repository_contract_test.go`, y se corren con la base levantada (`TEST_DATABASE_URL`). Los tests unitarios no dependen de la base (corren en memoria); los de integración PostgreSQL siguen las reglas de TEC-05 (se omiten sin `TEST_DATABASE_URL`, CA-TEC05.7).
- **Migración de la base existente.** Los scripts de `db/init` solo corre PostgreSQL la primera vez que arranca con el volumen vacío, así que una base de desarrollo creada antes de esta columna no la tiene: hay que recrearla (`docker compose down -v` y `docker compose up -d`). Se documenta en el README y en la evidencia TDD de esta historia.
- **Errores identificables.** Cada error exportado, comparable con `errors.Is`. `ErrItemNotFound` se propaga tal cual desde `internal/backlog`.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Exposición HTTP.** Es alcance de TEC-04.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Asignar el primer ítem de un proyecto a su sprint abierto | El ítem queda con `SprintID` y todo lo demás igual; se devuelve copia |
| Asignar dos ítems distintos al mismo sprint | Ambos quedan asignados |
| Ítem ya asignado al mismo sprint | `ErrItemAlreadyAssigned` |
| Ítem ya asignado a otro sprint | `ErrItemAlreadyAssigned` |
| Sprint `CERRADO` | `ErrSprintClosed` |
| Ítem de otro proyecto | `ErrItemFromAnotherProject` |
| SprintID inexistente, 0, negativo o sin sprints creados | `ErrSprintNotFound` |
| ItemID inexistente, 0, negativo o sin ítems creados | `ErrItemNotFound` |
| Ítem en estado `EN_CURSO` o `TERMINADO` | Se asigna igual |
| Después de asignar, modificar el ítem con el `Update` de HU-07 | El `SprintID` se conserva |
| Asignar no cambia ningún otro campo del ítem (estado, story points, horas) | El ítem queda igual salvo `SprintID` |
| Modificar el ítem devuelto | El ítem guardado no cambia |
| Asignaciones simultáneas del mismo ítem a dos sprints | Exactamente una gana; la otra `ErrItemAlreadyAssigned`; `-race` sin condiciones de carrera |
| Asignaciones simultáneas de ítems distintos al mismo sprint | Todas tienen éxito |
| El repositorio falla al consultar o al escribir | Se devuelve ese error y no se asigna nada |

## Criterios de aceptación

- **CA-11.1** Se asigna un ítem existente a un sprint existente y `ABIERTO` del mismo proyecto: el ítem queda con su `SprintID` y el resto de sus datos sin cambios, y se devuelve una copia.
- **CA-11.2** Un ítem ya asignado no se puede asignar de nuevo (ni al mismo sprint): se devuelve `ErrItemAlreadyAssigned` y no se modifica nada.
- **CA-11.3** No se puede asignar a un sprint `CERRADO` (`ErrSprintClosed`) ni un ítem de otro proyecto (`ErrItemFromAnotherProject`): no se modifica nada.
- **CA-11.4** Un sprint inexistente (incluidos ID 0 y negativo) devuelve `ErrSprintNotFound`, y un ítem inexistente devuelve `ErrItemNotFound`; no se asigna nada.
- **CA-11.5** Asignar no cambia ningún otro campo del ítem, y el `Update` de HU-07 conserva el `SprintID` asignado.
- **CA-11.6** De varias asignaciones simultáneas del mismo ítem, exactamente una tiene éxito y las demás devuelven `ErrItemAlreadyAssigned` (verificado con `go test -race`).
- **CA-11.7** Modificar el ítem devuelto no altera el ítem guardado.
- **CA-11.8** Si el repositorio falla, se devuelve ese error y no se asigna nada.

## Comparación con el acta de Planning

El acta del Sprint 2 describe HU-11 como "Asignar historias del backlog a un sprint" (Issue #25, 5 Story Points, responsable Buttini Cristobal), dependiente de HU-10 y HU-05, con la nota "Cruza dos paquetes (sprint y backlog) y tiene tres reglas: mismo proyecto, un solo sprint por ítem y sprint abierto" y "Si la asignación se guarda en el ítem, usa el `Update` del backlog de HU-07 (lo decide su SDD)". **El alcance coincide:** las tres reglas del acta son RN-11.2, RN-11.3 y RN-11.1, y la historia cruza los dos paquetes. La decisión que el acta le deja a esta SDD se toma en RN-11.4: la asignación **sí se guarda en el ítem** (`SprintID`), pero no a través del `Update` de HU-07 (que revalida campos de texto y pasa a conservar el `SprintID`), sino con una operación nueva de `AssignToSprint` que usa las operaciones ya existentes del repositorio de backlog, sin cambiar su interfaz. No hay otras diferencias entre el acta y esta SDD.

## Comparación con los criterios del Issue #25 (borrador)

| CA en el Issue #25 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-11.1 Se pueden asignar ítems del mismo proyecto a un sprint abierto | CA-11.1 (precisado) | Ítem y sprint tienen que existir; la asignación queda en el `SprintID` del ítem y no cambia nada más |
| CA-11.2 Un ítem no puede estar asignado a dos sprints a la vez | CA-11.2 (precisado) | Ya asignado es error aunque el sprint sea el mismo; error `ErrItemAlreadyAssigned` |
| CA-11.3 No se pueden asignar ítems a un sprint cerrado ni ítems de otro proyecto | CA-11.3 (precisado) | Errores `ErrSprintClosed` y `ErrItemFromAnotherProject` identificados |
| — | CA-11.4 (nuevo) | Validación de que el sprint y el ítem existen |
| — | CA-11.5 (nuevo) | Conservación del `SprintID` y aislamiento del resto de los campos |
| — | CA-11.6 (nuevo) | Concurrencia: exactamente una asignación gana, verificado con `go test -race` |
| — | CA-11.7 (nuevo) | Copias por valor |
| — | CA-11.8 (nuevo) | Error del repositorio propagado |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #25** (se actualiza a mano después de aprobar la SDD):

- [ ] CA-11.1 Se asigna un ítem existente a un sprint ABIERTO del mismo proyecto: el ítem queda con su SprintID y el resto de sus datos sin cambios.
- [ ] CA-11.2 Un ítem ya asignado no se puede asignar de nuevo (ni al mismo sprint): devuelve `ErrItemAlreadyAssigned`.
- [ ] CA-11.3 No se puede asignar a un sprint CERRADO (`ErrSprintClosed`) ni un ítem de otro proyecto (`ErrItemFromAnotherProject`).
- [ ] CA-11.4 Un sprint inexistente devuelve `ErrSprintNotFound` y un ítem inexistente devuelve `ErrItemNotFound` (incluidos ID 0 y negativo).
- [ ] CA-11.5 Asignar no cambia ningún otro campo del ítem, y el Update de HU-07 conserva el SprintID.
- [ ] CA-11.6 De varias asignaciones simultáneas del mismo ítem, exactamente una tiene éxito y las demás devuelven `ErrItemAlreadyAssigned` (verificado con `go test -race`).
- [ ] CA-11.7 Modificar el ítem devuelto no altera el ítem guardado.
- [ ] CA-11.8 Si el repositorio falla, se devuelve ese error y no se asigna nada.

## Decisiones confirmadas por el Agile Enabler el 18/10

1. **Dónde se guarda**: en el ítem (`SprintID int`, 0 = sin asignar). No se usa el `Update` de HU-07 para escribirla (conserva el campo y revalida textos); se agrega `AssignToSprint` con lectura-modificación-escritura, sin cambiar la interfaz del repositorio de backlog.
2. **Reasignación**: prohibida. Ya asignado es error, aunque sea al mismo sprint. No hay desasignación: ninguna historia la pide.
3. **Estado del ítem**: indistinto; cualquier estado se asigna.
4. **Firmas**: `AssignItem(sprintID, itemID)` en el Manager de sprint y `AssignToSprint(itemID, sprintID)` en el de backlog; devuelven el ítem actualizado como copia.
5. **Nueva operación de lectura de sprints**: `GetByID` con `ErrSprintNotFound` (la necesitan HU-12 y HU-14).
6. **Pasos de godog en el paquete `sprint`** (la historia es de E3), con unit tests en ambos paquetes.
7. **Mutex**: sí, uno por paquete, nunca anidados (patrón ya usado); justificado por CA-11.6 con `-race`.
8. **Parte PostgreSQL** (después del merge de TEC-05): HU-11 incluye en su rama `GetByID` en el repositorio PostgreSQL de sprints y la columna `sprint_id` en el esquema y en el repositorio PostgreSQL de backlog; sin esas piezas el servidor no compila y el backend PostgreSQL perdería las asignaciones. La columna es `INTEGER NOT NULL DEFAULT 0` sin clave foránea (0 = sin asignar); los tests de contrato de TEC-05 se extienden con ambas piezas y se corren con la base levantada; una base creada antes se recrea con `docker compose down -v` y `docker compose up -d` (documentado en el README y en la evidencia).
