# HU-08 · Cambiar el estado de un ítem

> Issue del repositorio: #20

## Historia de usuario

Como Product Builder, quiero cambiar el estado de un ítem del backlog, para reflejar su avance.

## Objetivo

Permitir cambiar el estado de un ítem existente del Product Backlog entre los estados válidos (PENDIENTE, EN_CURSO y TERMINADO), validando que la transición esté permitida, y dejar definidas en código las reglas de transición que internal/backlog/item.go reserva para esta historia ("Las transiciones entre estados las define HU-08").

## Alcance

- **Incluye:** el método Manager.ChangeStatus en internal/backlog/status.go, las transiciones permitidas y sus errores, los tests unitarios y los escenarios BDD con godog.
- **No incluye:** modificar otros campos del ítem (HU-07), asignar story points (HU-20), asignar el ítem a un sprint (HU-11), registrar historias completadas de un sprint (HU-12), historial o auditoría de cambios de estado ni la exposición HTTP (TEC-04).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | ID del ítem cuyo estado se cambia |
| Estado | texto | Sí | Estado destino: PENDIENTE, EN_CURSO o TERMINADO; se acepta sin distinguir mayúsculas ni contar los espacios de los extremos, igual que el estado en HU-02 y la prioridad en HU-04 |

Firma: `ChangeStatus(id int, status string) (Item, error)`.

### Salidas esperadas

- **Si el cambio es válido:** el ítem con el estado nuevo, como copia por valor (con su propia copia de la lista de criterios). Los demás campos quedan como estaban.
- **Si el estado destino es igual al actual:** el ítem sin cambios, como copia por valor, y sin error (RN-08.3).
- **Si el cambio no es válido:** un error que indica el motivo (ver "Errores"), y el ítem guardado queda exactamente como estaba.

### Errores

| Condición | Error |
|---|---|
| Estado destino que, normalizado, no es PENDIENTE, EN_CURSO ni TERMINADO (incluido vacío) | `ErrInvalidStatus` ("el estado del ítem no es válido"), nuevo, en `status.go` |
| No existe un ítem con ese ID (incluye 0 y negativos) | `ErrItemNotFound` de HU-05, propagado tal cual |
| Transición no permitida (PENDIENTE → TERMINADO o TERMINADO → PENDIENTE) | `ErrInvalidTransition` ("la transición de estado no está permitida"), nuevo, en `status.go` |
| El repositorio falla al buscar o al actualizar | El error que devuelve el repositorio |

**Orden de validación:**

1. Normalizar y validar el estado destino (`ErrInvalidStatus`), sin consultar el repositorio. Un estado inválido sobre un ítem inexistente devuelve `ErrInvalidStatus`.
2. Con el mutex del `Manager` tomado, en este orden: existencia del ítem con `GetByID` (`ErrItemNotFound`), mismo estado (no-op: devuelve una copia sin llamar a `Update`), transición (`ErrInvalidTransition`) y `Update` del repositorio.

Se informa el primer error encontrado.

## Reglas de negocio

1. **RN-08.1 Estados válidos.** Los únicos estados son PENDIENTE, EN_CURSO y TERMINADO (RN-04.7). El estado destino se normaliza (`strings.ToUpper(strings.TrimSpace(...))`) y, si no es uno de los tres, se rechaza con `ErrInvalidStatus`. La validación ocurre antes de consultar el repositorio.
2. **RN-08.2 Transiciones permitidas.** Un paso por vez, hacia adelante o hacia atrás:
   - PENDIENTE → EN_CURSO
   - EN_CURSO → TERMINADO
   - EN_CURSO → PENDIENTE (volver atrás)
   - TERMINADO → EN_CURSO (reabrir)

   No están permitidas PENDIENTE → TERMINADO ni TERMINADO → PENDIENTE: el ítem tiene que pasar por EN_CURSO. Una transición no permitida devuelve `ErrInvalidTransition`.
3. **RN-08.3 Mismo estado.** Si el estado destino es igual al actual, la operación es un no-op válido: no llama a `Update`, no devuelve error y devuelve una copia del ítem.
4. **RN-08.4 Ítem existente.** El ítem tiene que existir; si `GetByID` devuelve `ErrItemNotFound`, se propaga tal cual.
5. **RN-08.5 Solo cambia el estado.** El cambio de estado no modifica ningún otro campo del ítem: título, descripción, prioridad, criterios, horas estimadas, story points, sprint asignado, proyecto ni ID.
6. **RN-08.6 Persistencia.** Se guarda con `Repository.Update` (HU-07), con el ítem completo leído con `GetByID` y el estado nuevo; la interfaz `Repository` no cambia. No se usa `Manager.Update` de HU-07, porque conserva el estado a propósito.
7. **RN-08.7 Copias por valor.** `ChangeStatus` devuelve copias del ítem, con su propia copia de la lista de criterios (RN-05.3): modificar el resultado no altera el ítem guardado.

## Decisión sobre entidades referenciadas

HU-08 referencia una entidad por ID: el ítem cuyo estado se cambia. **Sí se valida su existencia** (RN-08.4), porque la operación modifica esa entidad concreta: cambiar el estado de un ítem inexistente no tiene sentido y, sin la validación, sería un no-op silencioso que ocultaría errores de quien llama. Se reutiliza `GetByID` de HU-05 con `ErrItemNotFound`, sin duplicar la regla ni el error. No referencia ninguna otra entidad.

## Restricciones

- **Organización del código.** `internal/backlog/status.go` (nuevo): `ChangeStatus`, `ErrInvalidStatus`, `ErrInvalidTransition` y la tabla de transiciones permitidas. Se actualiza el comentario de `Status` en `item.go`, que hoy dice "Las transiciones entre estados las define HU-08", para que remita a `status.go`. No se tocan `create.go`, `get.go`, `list.go` ni `update.go`.
- **Repositorio.** Se reutilizan `GetByID` (HU-05) y `Update` (HU-07). La interfaz `Repository` y el `failingRepository` de los tests no cambian (ya tiene `getErr` y `updateErr`).
- **Sin cambios de esquema.** `ChangeStatus` no cambia el esquema ni la interfaz `Repository`; funciona igual con el repositorio en memoria y con el de PostgreSQL (la columna `status` ya se guarda en `Update`), y los contratos de TEC-05 se extienden con un caso de cambio de estado que conserva `sprint_id` y `estimated_hours`.
- **Concurrencia.** `ChangeStatus` toma el mutex del `Manager` para la lectura, la validación y la escritura (patrón de HU-07, RN-07.6), justificado por CA-08.9 con `go test -race` (lo corre el CI; en Windows local no hay cgo).
- **Errores identificables.** Cada error está exportado y se compara con `errors.Is`.
- **Tests.** Unitarios en `internal/backlog/status_test.go` y pasos de godog en `internal/backlog/status_features_test.go`, con el runner `TestFeatures_HU08CambiarEstado`, sobre `features/HU-08-cambiar-estado-item.feature`. El caso de CA-08.7 ("no llama a `Update`") se verifica con el `failingRepository` configurado con `updateErr`: pedir el mismo estado no devuelve error.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Exposición HTTP.** Es alcance de TEC-04.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| PENDIENTE → EN_CURSO | El ítem queda EN_CURSO; los demás campos no cambian |
| EN_CURSO → TERMINADO | El ítem queda TERMINADO |
| EN_CURSO → PENDIENTE (volver atrás) | El ítem queda PENDIENTE |
| TERMINADO → EN_CURSO (reabrir) | El ítem queda EN_CURSO |
| PENDIENTE → TERMINADO (saltea EN_CURSO) | `ErrInvalidTransition`; el ítem sigue PENDIENTE |
| TERMINADO → PENDIENTE | `ErrInvalidTransition`; el ítem sigue TERMINADO |
| Estado destino igual al actual | Sin error; devuelve el ítem sin cambios y no llama a `Update` |
| Estado "en_curso " o " Pendiente" | Se normaliza a EN_CURSO o PENDIENTE |
| Estado vacío, con solo espacios o "BLOQUEADO" | `ErrInvalidStatus`; el ítem no cambia |
| Estado inválido sobre un ítem inexistente | `ErrInvalidStatus` (se valida primero el estado) |
| ID inexistente, 0, negativo o sin ítems creados | `ErrItemNotFound` |
| Ítem con horas estimadas, story points y sprint asignado | Los conserva después del cambio |
| Modificar el ítem devuelto o sus criterios | El ítem guardado no cambia |
| Varios cambios simultáneos de PENDIENTE a EN_CURSO sobre el mismo ítem | Ninguno falla; el primero cambia el estado, los demás son no-op; el ítem queda EN_CURSO |
| El repositorio falla al buscar o al actualizar | Se devuelve ese error; el estado no cambia |

## Criterios de aceptación

- **CA-08.1** Cambiar un ítem PENDIENTE a EN_CURSO actualiza su estado y devuelve el ítem con el estado nuevo y el resto de los campos intactos (incluido el sprint asignado).
- **CA-08.2** Cambiar un ítem EN_CURSO a TERMINADO actualiza su estado.
- **CA-08.3** Volver atrás (EN_CURSO → PENDIENTE) y reabrir (TERMINADO → EN_CURSO) son transiciones válidas.
- **CA-08.4** PENDIENTE → TERMINADO y TERMINADO → PENDIENTE se rechazan con `ErrInvalidTransition` y el ítem conserva su estado anterior.
- **CA-08.5** Un estado inválido o vacío se rechaza con `ErrInvalidStatus` y el ítem no cambia. El estado se acepta sin distinguir mayúsculas ni espacios de los extremos.
- **CA-08.6** Cambiar el estado de un ítem inexistente (incluidos ID 0 y negativo) devuelve `ErrItemNotFound` y no se modifica nada.
- **CA-08.7** Cambiar al mismo estado actual no da error, no llama a `Update` y no altera el ítem.
- **CA-08.8** Los cambios quedan guardados en el repositorio (se verifican con `GetItem`), y modificar el ítem devuelto no altera el guardado.
- **CA-08.9** De varios cambios simultáneos de PENDIENTE a EN_CURSO sobre el mismo ítem, ninguno falla (el primero cambia el estado y los demás son no-op) y el ítem queda EN_CURSO, verificado con `go test -race`.
- **CA-08.10** Si el repositorio falla al buscar o al actualizar, se devuelve ese error y el estado no cambia.

## Comparación con el acta de Planning

El acta del Sprint 2 (`docs/actas/sprint-2/planning.md`) describe HU-08 como "Cambiar el estado de un ítem" (Issue #20, 3 Story Points, responsable Sosa Ricardo), dependiente de "HU-05 (`GetByID`) y el `Update` del backlog que agrega HU-07. Define las transiciones de estado (comentario de `item.go`)", y en la estimación como "Define y valida las transiciones de estado; reutiliza el `Update` de HU-07". **El alcance coincide:** esta SDD define las transiciones (RN-08.2) y reutiliza `GetByID` y el `Update` del repositorio (RN-08.6). No hay diferencias entre el acta y esta SDD.

## Comparación con los criterios del Issue #20 (borrador)

| CA en el Issue #20 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-08.1 Se puede pasar un ítem de un estado a otro permitido | CA-08.1, CA-08.2 y CA-08.3 (precisado) | Se definen las cuatro transiciones permitidas |
| CA-08.2 No se aceptan estados inválidos ni transiciones no permitidas | CA-08.4 y CA-08.5 (precisado) | Se identifican `ErrInvalidTransition` y `ErrInvalidStatus`; el ítem no cambia |
| CA-08.3 Cambiar el estado de un ítem inexistente devuelve error | CA-08.6 (precisado) | Se identifica `ErrItemNotFound` |
| — | CA-08.7 (nuevo) | Mismo estado: no-op sin error |
| — | CA-08.8 (nuevo) | Persistencia y copias por valor |
| — | CA-08.9 (nuevo) | Concurrencia verificada con `go test -race` |
| — | CA-08.10 (nuevo) | Error del repositorio propagado |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #20** (se actualiza a mano después de aprobar la SDD):

- [ ] CA-08.1 Cambiar un ítem PENDIENTE a EN_CURSO actualiza su estado y devuelve el ítem con el estado nuevo y el resto de los campos intactos (incluido el sprint asignado).
- [ ] CA-08.2 Cambiar un ítem EN_CURSO a TERMINADO actualiza su estado.
- [ ] CA-08.3 Volver atrás (EN_CURSO → PENDIENTE) y reabrir (TERMINADO → EN_CURSO) son transiciones válidas.
- [ ] CA-08.4 PENDIENTE → TERMINADO y TERMINADO → PENDIENTE se rechazan con `ErrInvalidTransition` y el ítem conserva su estado anterior.
- [ ] CA-08.5 Un estado inválido o vacío se rechaza con `ErrInvalidStatus` y el ítem no cambia. El estado se acepta sin distinguir mayúsculas ni espacios de los extremos.
- [ ] CA-08.6 Cambiar el estado de un ítem inexistente (incluidos ID 0 y negativo) devuelve `ErrItemNotFound` y no se modifica nada.
- [ ] CA-08.7 Cambiar al mismo estado actual no da error, no llama a `Update` y no altera el ítem.
- [ ] CA-08.8 Los cambios quedan guardados en el repositorio, y modificar el ítem devuelto no altera el guardado.
- [ ] CA-08.9 De varios cambios simultáneos de PENDIENTE a EN_CURSO sobre el mismo ítem, ninguno falla (el primero cambia el estado y los demás son no-op) y el ítem queda EN_CURSO (verificado con `go test -race`).
- [ ] CA-08.10 Si el repositorio falla al buscar o al actualizar, se devuelve ese error y el estado no cambia.

## Decisiones confirmadas por el Agile Enabler el 18/10

1. Transiciones de un paso, hacia adelante o hacia atrás: PENDIENTE ↔ EN_CURSO y EN_CURSO ↔ TERMINADO. Los saltos de dos pasos (PENDIENTE → TERMINADO y TERMINADO → PENDIENTE) se rechazan.
2. Mismo estado = no-op: sin error, sin llamar a `Update`, devuelve una copia del ítem.
3. Estado destino normalizado (`ToUpper` + `TrimSpace`), igual que el estado en HU-02 y la prioridad en HU-04.
4. Se valida que el ítem exista (`GetByID` de HU-05, `ErrItemNotFound`).
5. Orden de validación: estado (sin consultar el repositorio), existencia, mismo estado, transición y escritura.
6. Persistencia con el `Update` del repositorio de HU-07; la interfaz `Repository` no cambia.
7. Mutex sí, justificado por CA-08.9 y su test con `-race`.
8. Después de la aprobación del 18/10, durante la implementación, se agregó el sprint asignado (`SprintID`, que incorporó HU-11) a lo que no cambia (RN-08.5, CA-08.1 y casos límite) y la restricción "Sin cambios de esquema", con un caso nuevo en el contrato de repositorio de TEC-05. El Agile Enabler aprobó estos cambios el 20/10, en la revisión del Pull Request #99.
