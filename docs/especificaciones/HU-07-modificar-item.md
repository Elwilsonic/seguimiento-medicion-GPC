# HU-07 · Modificar datos de un ítem

> Issue del repositorio: #19

## Historia de usuario

Como Agile Enabler, quiero modificar los datos de un ítem del backlog, para corregirlo o precisarlo.

## Objetivo

Cambiar el título, la descripción, la prioridad, los criterios de aceptación y las horas estimadas de un ítem ya guardado del Product Backlog, identificado por su ID, aplicando las mismas reglas de validación que la creación (HU-04) y agregando el campo de horas estimadas que HU-04 dejó pendiente, para que HU-08 (estado), HU-11 (asignación a sprints) y las métricas trabajen con información actualizada y consistente.

Es la operación de escritura de la épica E2 sobre un ítem existente: HU-04 crea, HU-05 consulta, HU-06 lista y HU-07 modifica (el estado lo cambia HU-08 y los story points HU-20).

## Alcance

- **Incluye:** modificar un ítem existente a través del `Manager` del paquete `internal/backlog`, validando los mismos campos obligatorios que HU-04 (título, prioridad, criterios de aceptación), las horas estimadas y la existencia del ítem (reutilizando HU-05). Agrega el campo `EstimatedHours` al ítem y la operación `Update` a `backlog.Repository`.
- **No incluye:** crear ítems (HU-04), consultarlos (HU-05), listarlos (HU-06), cambiar el estado o definir sus transiciones (HU-08), asignar story points (HU-20), mover un ítem de proyecto ni la exposición HTTP (TEC-01).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | Identificador del ítem a modificar |
| Título | texto | Sí | Nuevo título |
| Descripción | texto | No | Nueva descripción; puede estar vacía |
| Prioridad | texto | Sí | Nueva prioridad; se acepta en cualquier combinación de mayúsculas y con espacios en los extremos: se normaliza y tiene que quedar en `ALTA`, `MEDIA` o `BAJA` |
| Criterios de aceptación | lista de textos | Sí | Al menos un criterio no vacío después de descartar los vacíos |
| Horas estimadas | número (float64) | No | Nuevo valor; 0 significa "sin horas estimadas" (quita las que hubiera); son error los negativos, NaN e infinito |

**No son entradas:** el ID y ProjectID del ítem (no se puede mover un ítem de proyecto), el estado (HU-08) ni los story points (HU-20).

La modificación reemplaza los seis campos en una sola llamada: para cambiar un solo campo, el llamador reenvía los valores actuales de los demás (decisión 4).

Firma: `Update(id int, title, description, priority string, criteria []string, estimatedHours float64) (Item, error)`.

### Salidas esperadas

- **Si la modificación es válida:** el ítem actualizado, como **copia por valor** (incluida una lista de criterios nueva), con:
  - **ID y ProjectID:** los mismos (no cambian).
  - **Título:** el recibido, sin los espacios de los extremos.
  - **Descripción:** la recibida, tal como llegó.
  - **Prioridad:** normalizada, en mayúsculas.
  - **Criterios:** los recibidos que no quedan vacíos, sin espacios de extremos y en el mismo orden.
  - **Estado y StoryPoints:** los que ya tenía (no cambian).
  - **EstimatedHours:** el recibido (0 = sin horas).
- **Si la modificación no es válida:** un error que indica el motivo (ver "Condiciones de error"), y el ítem guardado queda exactamente como estaba (nunca se aplica una modificación parcial).

## Reglas de negocio

1. **RN-07.1 Título obligatorio.** Aplica RN-04.1 al modificar: no vacío, sin espacios de extremos.
2. **RN-07.2 Descripción opcional.** Aplica RN-04.2: puede quedar vacía, se guarda tal como llega.
3. **RN-07.3 Prioridad válida.** Aplica RN-04.3: normalizada y en `ALTA`, `MEDIA` o `BAJA`.
4. **RN-07.4 Criterios de aceptación.** Aplica RN-04.4: se descartan los vacíos, al menos uno no vacío, orden conservado, lista nueva.
5. **RN-07.5 Horas estimadas.** Valor nuevo del ítem: tiene que ser un número mayor o igual a 0 y no NaN ni infinito (`ErrInvalidEstimatedHours`). 0 significa "sin horas estimadas": es válido y reemplaza (quita) el valor anterior. Todo ítem creado por HU-04 queda con 0.
6. **RN-07.6 Ítem existente.** Requiere que exista un ítem con el ID recibido; si no, `ErrItemNotFound` (el mismo de HU-05), sin error aparte.
7. **RN-07.7 Copia por valor.** Devuelve una copia del ítem actualizado (incluida la lista de criterios): modificar el resultado no altera lo guardado (mismo criterio que RN-04.9).

## Decisión sobre entidades referenciadas

HU-07 opera sobre una entidad referenciada por ID: el ítem a modificar. **Sí se valida su existencia** (RN-07.6): es la condición de error central de la historia (CA-07.3); si no existe, no hay nada que actualizar y se devuelve `ErrItemNotFound`, reutilizando la búsqueda por ID de HU-05 (`Repository.GetByID`). No referencia ninguna otra entidad: los seis campos que cambia son propios del ítem. El ProjectID no es entrada, así que no hay que validar ningún proyecto.

## Restricciones

- **Organización del código.** Sigue el patrón de HU-02 en `internal/backlog`: el código de producción va en `update.go`, con el método `Update` del `Manager`, el error `ErrInvalidEstimatedHours` y la validación de las horas. Se agrega el campo `EstimatedHours float64` a `Item` en `item.go` (con comentario: 0 = sin asignar) y `Create` lo deja en 0 (ya lo hace). No se tocan `create.go`, `get.go` ni `list.go`.
- **Interfaz de repositorio.** Se agrega a `Repository` la operación `Update(item Item) (Item, error)`: reemplaza el ítem guardado cuyo ID coincida, devuelve una copia (incluida la lista) y `ErrItemNotFound` si no existe. Coherente con `Save` (recibe y devuelve el ítem completo por valor). `MemoryRepository` la implementa; TEC-05 también deberá cumplirla.
- **Reutilización de HU-05.** El `Manager` usa `Repository.GetByID` para verificar la existencia; no se duplica la búsqueda. También reutiliza `normalizePriority` y `cleanCriteria` de `create.go`.
- **Consecuencia de ampliar la interfaz.** El `Update` de los repositorios falsos de test que implementan `backlog.Repository` — `failingRepository` (`create_test.go`, HU-04) — se agrega **en el commit RED, junto con los tests**, para que el paquete compile y los tests fallen por la funcionalidad nueva y no por errores de compilación ajenos. Ese archivo no recibe ningún otro cambio.
- **Orden de validaciones.** Primero el contenido recibido, sin consultar el repositorio: título (RN-07.1), prioridad (RN-07.3), criterios (RN-07.4) y horas (RN-07.5), en ese orden. Después, con el mutex tomado: existencia (RN-07.6, con `GetByID`) y finalmente la escritura (`Update` del repositorio). Si falla cualquiera, no se escribe nada. Como en HU-02, los errores de contenido se informan antes de consultar el repositorio.
- **Copias por valor.** `Manager.Update` y `Repository.Update` devuelven `Item` por valor, con lista de criterios nueva (`slices.Clone` en memoria), nunca punteros al estado interno.
- **Errores identificables.** `ErrInvalidEstimatedHours` es nuevo, exportado, comparable con `errors.Is` y declarado en `update.go` (patrón de HU-02 con `ErrInvalidStatus`). Se reutilizan `ErrEmptyTitle`, `ErrInvalidPriority`, `ErrNoAcceptanceCriteria` y `ErrItemNotFound` sin duplicarlos.
- **Concurrencia.** `Update` toma el mutex del `Manager` (patrón de HU-02 y CA-04.7), justificado por CA-07.6 con `go test -race` en el CI. La búsqueda de existencia va adentro del mutex (es parte de la operación leer-modificar-escribir).
- **Pasos godog.** Los pasos BDD van en `update_features_test.go`, con su runner `TestFeatures_HU07ModificarItem`. No se tocan los features tests de las otras historias.
- **Librerías.** Solo la librería estándar en el código (incluida `math` para NaN/Inf); godog solo en los tests.
- **Exposición HTTP.** `PUT /items/{id}` es alcance de TEC-04, no de esta historia.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Cambiar solo las horas (se reenvía lo demás actual) | Cambian las horas; lo demás queda igual |
| Cambiar solo el título / solo la prioridad / solo los criterios | Cambia ese campo; lo demás queda igual |
| Horas en 0 cuando el ítem no tenía | Válido: queda sin horas (0) |
| Horas en 0 cuando el ítem tenía 8.5 | Válido: quedan en 0 (se quitan) |
| Horas 2.5 (fraccionadas) | Se guardan 2.5 |
| Horas -1, NaN o +Inf | Error: `ErrInvalidEstimatedHours`; el ítem queda como estaba |
| Título con espacios en los extremos | Se guarda recortado |
| Título vacío o con solo espacios | Error: `ErrEmptyTitle`; el ítem queda como estaba |
| Prioridad " media " o "baja" | Se guarda `MEDIA` / `BAJA` |
| Prioridad "URGENTE" o vacía | Error: `ErrInvalidPriority`; el ítem queda como estaba |
| Criterios `["A", "  ", ""]` | Se guarda `["A"]` |
| Criterios `nil` o solo vacíos | Error: `ErrNoAcceptanceCriteria`; el ítem queda como estaba |
| Modificar un ID inexistente, 0 o negativo | Error: `ErrItemNotFound` |
| Modificar sin ítems creados | Error: `ErrItemNotFound` |
| Varios datos inválidos a la vez (título vacío y horas negativas) | Se informa el primero según el orden (`ErrEmptyTitle`) |
| Modificar el ítem devuelto o sus criterios | El ítem guardado no cambia |
| Estado y story points después de modificar | Quedan como estaban (los cambia HU-08 / HU-20) |
| Varias modificaciones simultáneas del mismo ítem | El ítem guardado coincide completo con una de las versiones enviadas; `go test -race` no reporta condiciones de carrera |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Título vacío o con solo espacios (RN-07.1) | `ErrEmptyTitle` | No se modifica nada |
| Prioridad que, normalizada, no es válida (RN-07.3) | `ErrInvalidPriority` | No se modifica nada |
| Ningún criterio no vacío (RN-07.4) | `ErrNoAcceptanceCriteria` | No se modifica nada |
| Horas negativas, NaN o infinitas (RN-07.5) | `ErrInvalidEstimatedHours` ("las horas estimadas deben ser un número mayor o igual a cero") | No se modifica nada |
| No existe un ítem con ese ID (RN-07.6; incluye 0 y negativos) | `ErrItemNotFound` | No se modifica nada |
| El repositorio falla al buscar o actualizar | El error que devuelve el repositorio | No se modifica nada y el ítem queda como estaba |

## Criterios de aceptación

- **CA-07.1** Se pueden cambiar el título, la descripción, la prioridad, los criterios de aceptación y las horas estimadas de un ítem existente en una sola operación; cada campo puede quedar igual reenviando su valor actual. El ID, el ProjectID, el estado y los story points no se modifican.
- **CA-07.2** Los campos obligatorios siguen validándose: título vacío o con solo espacios devuelve `ErrEmptyTitle`, prioridad distinta de `ALTA`, `MEDIA` o `BAJA` devuelve `ErrInvalidPriority` y no tener ningún criterio no vacío devuelve `ErrNoAcceptanceCriteria`; en esos casos el ítem guardado queda exactamente como estaba. La descripción sigue siendo opcional.
- **CA-07.3** Modificar un ítem que no existe devuelve `ErrItemNotFound` (incluye ID inexistente, 0, negativo y no tener ítems creados).
- **CA-07.4** Si se asignan, las horas estimadas son un número mayor o igual a cero (se admiten fracciones, por ejemplo 2.5); 0 significa sin horas estimadas y es el valor de los ítems creados por HU-04. Un valor negativo, NaN o infinito devuelve `ErrInvalidEstimatedHours` y el ítem no cambia.
- **CA-07.5** `Update` devuelve una copia por valor: modificar el resultado devuelto (o su lista de criterios) no altera el ítem guardado.
- **CA-07.6** Modificar mientras se crean o consultan ítems en simultáneo no produce accesos concurrentes sin protección (verificado con `go test -race`) y, después de varias modificaciones simultáneas del mismo ítem, el ítem guardado coincide completo con una de las versiones enviadas.
- **CA-07.7** Si el repositorio falla al buscar o actualizar, se devuelve ese error y el ítem no cambia.

## Comparación con el acta de Planning

El acta del Sprint 2 (`docs/actas/sprint-2/planning.md`) describe HU-07 como "Modificar datos de un ítem" (Issue #19, 3 Story Points, responsable Sosa Ricardo, primera historia del sprint) y precisa su alcance: "HU-05 (`GetByID`). Agrega `Update` a `backlog.Repository` y el campo de horas estimadas al ítem (CA-07.1, CA-07.4)"; en la estimación la resume como "Como HU-02 (modificar), más el campo nuevo de horas estimadas y el `Update` del repositorio". **El alcance coincide**: esta SDD agrega `Update` a `backlog.Repository`, el campo `EstimatedHours` al ítem y reutiliza `GetByID` de HU-05, con el patrón de HU-02. No hay diferencias entre el acta y esta SDD.

## Comparación con los criterios del Issue #19 (borrador)

| CA en el Issue #19 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-07.1 Se pueden cambiar título, descripción, prioridad, criterios de aceptación y horas estimadas de un ítem existente. | CA-07.1 (precisado) | Se aclara que es una sola operación, que cada campo puede quedar igual y que ID, ProjectID, estado y story points no cambian |
| CA-07.2 Los campos obligatorios siguen validándose. | CA-07.2 (precisado) | Se identifican los errores (`ErrEmptyTitle`, `ErrInvalidPriority`, `ErrNoAcceptanceCriteria`) y se aclara que el ítem queda como estaba |
| CA-07.3 Modificar un ítem inexistente devuelve error. | CA-07.3 (precisado) | Se identifica el error (`ErrItemNotFound`) y se incluyen ID 0, negativo y no tener ítems creados |
| CA-07.4 Las horas estimadas deben ser un número positivo; mientras no se asignen, el ítem queda sin horas estimadas. | CA-07.4 (precisado) | float64 con 0 = sin asignar (mismo criterio que StoryPoints en HU-04); la validación es "mayor o igual a cero" y no NaN/Inf, no "> 0", porque 0 es válido (sin horas, RN-07.5); negativos, NaN e infinito son error. Decidido en la redacción de la SDD |
| — | CA-07.5 (nuevo) | Copia por valor, mismo patrón que CA-02.6 y CA-04.6 |
| — | CA-07.6 (nuevo) | Concurrencia verificada con `go test -race`, mismo patrón que CA-02.7 |
| — | CA-07.7 (nuevo) | Error del repositorio al buscar o actualizar: se devuelve y el ítem no cambia |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #19** (cuando se apruebe la SDD):

```markdown
## Criterios de aceptación

- [ ] CA-07.1 Se pueden cambiar el título, la descripción, la prioridad, los criterios de aceptación y las horas estimadas de un ítem existente en una sola operación; cada campo puede quedar igual reenviando su valor actual. El ID, el ProjectID, el estado y los story points no se modifican.
- [ ] CA-07.2 Los campos obligatorios siguen validándose: título vacío o con solo espacios devuelve `ErrEmptyTitle`, prioridad distinta de ALTA, MEDIA o BAJA devuelve `ErrInvalidPriority` y no tener ningún criterio no vacío devuelve `ErrNoAcceptanceCriteria`; en esos casos el ítem guardado queda exactamente como estaba. La descripción sigue siendo opcional.
- [ ] CA-07.3 Modificar un ítem que no existe devuelve `ErrItemNotFound` (incluye ID inexistente, 0, negativo y no tener ítems creados).
- [ ] CA-07.4 Si se asignan, las horas estimadas son un número mayor o igual a cero (se admiten fracciones, por ejemplo 2.5); 0 significa sin horas estimadas y es el valor de los ítems creados por HU-04. Un valor negativo, NaN o infinito devuelve `ErrInvalidEstimatedHours` y el ítem no cambia.
- [ ] CA-07.5 `Update` devuelve una copia por valor: modificar el resultado devuelto (o su lista de criterios) no altera el ítem guardado.
- [ ] CA-07.6 Modificar mientras se crean o consultan ítems en simultáneo no produce accesos concurrentes sin protección (verificado con `go test -race`) y, después de varias modificaciones simultáneas del mismo ítem, el ítem guardado coincide completo con una de las versiones enviadas.
- [ ] CA-07.7 Si el repositorio falla al buscar o actualizar, se devuelve ese error y el ítem no cambia.
```

## Decisiones tomadas en la redacción de la SDD

1. **Tipo y centinela de las horas:** `EstimatedHours float64` en `Item`; 0 significa "sin horas estimadas", mismo criterio que `StoryPoints` en 0 (RN-04.8). Sin punteros.
2. **Update con 0:** se acepta y deja el ítem sin horas (reemplaza el valor anterior, como reemplaza cualquier otro campo en HU-02).
3. **Validez de las horas:** la validación es ≥ 0 y no NaN/Inf (`math.IsNaN`, `math.IsInf`), no "> 0"; son error los negativos, NaN e infinito.
4. **Una sola operación** `Update(id, title, description, priority, criteria, estimatedHours)`: reemplaza los seis campos; para cambiar uno se reenvían los actuales. La prioridad y los criterios llegan como texto/lista y se normalizan adentro, igual que en `Create`.
5. **Estado y story points fuera del alcance:** los cambia HU-08 y HU-20 respectivamente; `Update` conserva los que el ítem tenía.
6. **Error nuevo:** `ErrInvalidEstimatedHours`, exportado y declarado en `update.go` (patrón de HU-02 con `ErrInvalidStatus`).
7. **Orden de validaciones:** contenido (título, prioridad, criterios, horas) antes de existencia; los errores de contenido no consultan el repositorio.
8. **Mutex sí:** `Update` es leer-modificar-escribir; justificado por CA-07.6 y su test con `-race` (no por uniformidad).
