# HU-09 · Priorizar los ítems del backlog

> Issue del repositorio: #22

## Historia de usuario

Como Agile Enabler, quiero ordenar los ítems del backlog por prioridad, para trabajar primero lo más importante.

## Objetivo

Obtener los ítems del Product Backlog de un proyecto existente ordenados por prioridad (ALTA, MEDIA, BAJA) y, a igual prioridad, por ID ascendente, como una opción adicional al listado de HU-06, cuyo orden por ID (RN-06.6) sigue siendo el predeterminado.

Es una operación de solo lectura: no crea, modifica ni borra ítems, ni cambia la prioridad de ninguno.

## Alcance

- **Incluye:** el método `Manager.ListItemsByPriority` en `internal/backlog/priority.go`, que reutiliza `ListItems` (HU-06) y ordena la lista que devuelve; los tests unitarios y los escenarios BDD con godog.
- **No incluye:** cambiar el orden o el comportamiento de `ListItems` (HU-06), reordenar los ítems a mano, filtros, paginación, cambiar la prioridad de un ítem (HU-07), otros criterios de orden (estado, story points, horas) ni la exposición HTTP (TEC-04).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ProjectID | entero | Sí | ID del proyecto cuyo backlog se lista ordenado por prioridad |

Firma: `ListItemsByPriority(projectID int) ([]Item, error)`.

### Salidas esperadas

- **Si el proyecto existe:** una lista nueva con copias de todos los ítems del proyecto (los mismos que devuelve `ListItems`), ordenada primero los ALTA, después los MEDIA, después los BAJA y al final los de prioridad desconocida (RN-09.3); dentro de cada grupo, por ID ascendente (RN-09.2). Cada ítem tiene todos sus campos y su propia copia de la lista de criterios.
- **Si el proyecto existe y no tiene ítems:** una lista vacía (longitud 0, no `nil`), sin error.
- **Si hay un error:** el error (ver "Errores") y ninguna lista (`nil`).

### Errores

| Condición | Error | Efecto |
|---|---|---|
| No existe un proyecto con ese ProjectID (incluye 0, negativos y listar sin proyectos creados) | `ErrProjectNotFound` de `internal/project` (HU-03), propagado tal cual por `ListItems` | No se devuelve ninguna lista |
| El repositorio falla al listar | El error que devuelve el repositorio, propagado tal cual por `ListItems` | No se devuelve ninguna lista |

No se define ningún error nuevo: un backlog vacío no es un error (RN-06.2) y una prioridad desconocida tampoco (RN-09.3).

**Orden de validación:** es el de `ListItems`: primero se valida el proyecto (sin consultar el repositorio de ítems si no existe) y después se lista. Solo si `ListItems` no devuelve error se ordena la lista.

## Reglas de negocio

1. **RN-09.1 Orden por prioridad.** Los ítems se ordenan de mayor a menor prioridad: primero ALTA, después MEDIA y después BAJA.
2. **RN-09.2 Orden estable por ID.** A igual prioridad, los ítems quedan por ID ascendente (orden de creación). Se logra con un ordenamiento estable (`slices.SortStableFunc`) sobre la lista que devuelve `ListItems`, que ya viene ordenada por ID (RN-06.6): el ordenamiento estable conserva ese orden dentro de cada prioridad.
3. **RN-09.3 Prioridad desconocida.** Un ítem cuya prioridad guardada no es exactamente ALTA, MEDIA ni BAJA va al final de la lista, después de los BAJA, y entre ellos se conserva el orden por ID. No es un error y el ítem no se descarta. La prioridad no se vuelve a normalizar al ordenar: `Create` (RN-04.3) y `Update` (RN-07.3) ya guardan siempre una prioridad normalizada, así que este caso solo puede darse con datos guardados sin pasar por el `Manager` (por ejemplo, cargados directo en la base, que no tiene una restricción `CHECK` sobre `priority`).
4. **RN-09.4 Mismos ítems que HU-06.** El listado por prioridad contiene exactamente los mismos ítems que `ListItems`, solo en otro orden: todos los ítems del proyecto, sin filtrar por estado del ítem ni del proyecto (RN-06.1: un proyecto `FINALIZADO` también se lista), y solo los de ese proyecto (RN-06.3).
5. **RN-09.5 El listado de HU-06 no cambia.** `ListItems` sigue devolviendo los ítems por ID ascendente (RN-06.6), también después de llamar a `ListItemsByPriority`: el orden por prioridad es una opción, no reemplaza al predeterminado.
6. **RN-09.6 Proyecto inexistente.** Si el proyecto no existe, se devuelve `ErrProjectNotFound` de HU-03 y no una lista vacía, igual que en HU-06 (RN-06.4).
7. **RN-09.7 Backlog vacío.** Si el proyecto existe y no tiene ítems, se devuelve una lista vacía de longitud 0 (no `nil`), sin error, igual que en HU-06 (RN-06.2).
8. **RN-09.8 Copias por valor.** La lista devuelta es nueva y cada ítem tiene su propia copia de la lista de criterios (la que arma `ListItems`, RN-06.5). Modificar la lista, un ítem o sus criterios no altera los ítems guardados, y ordenar no modifica nada guardado.
9. **RN-09.9 Se ordena en el Manager.** El orden por prioridad lo aplica el `Manager` en memoria, no el repositorio: la interfaz `Repository`, el repositorio en memoria, el de PostgreSQL y el esquema de la base no cambian, y el resultado es el mismo con cualquiera de los dos repositorios.

## Decisión sobre entidades referenciadas

HU-09 referencia una entidad por ID: el proyecto cuyo backlog se lista. **Sí se valida su existencia** (RN-09.6), por los mismos motivos que en HU-06: sin la validación no se podría distinguir "el proyecto no tiene ítems" (válido) de "el proyecto no existe" (un error de quien llama). No se duplica la regla: se reutiliza `ListItems`, que ya valida el proyecto con `project.Manager.Get` de HU-03 y propaga `ErrProjectNotFound`. No referencia ninguna otra entidad: no consulta ítems por ID ni sprints.

## Restricciones

- **Organización del código.** `internal/backlog/priority.go` (nuevo): `ListItemsByPriority` y una función auxiliar no exportada que da el rango de cada prioridad (ALTA, MEDIA, BAJA y desconocida al final). No se tocan `list.go`, `create.go`, `get.go`, `update.go`, `status.go`, `assign.go` ni `item.go`.
- **Repositorio.** Se reutiliza `ListByProject` a través de `ListItems`. La interfaz `Repository`, `memory_repository.go`, `postgres_repository.go`, `db/init/01-schema.sql`, el `failingRepository` de los tests (ya tiene `listErr`) y los contratos de repositorio de TEC-05 no cambian: el orden por ID que usa RN-09.2 ya lo verifica el contrato "ListByProject devuelve solo los ítems del proyecto en orden de ID", que corre contra memoria y PostgreSQL.
- **Concurrencia.** No se agrega mutex ni se toma el del `Manager` en `priority.go`: `ListItems` ya toma el mutex mientras lee el repositorio y devuelve una lista propia, que después se ordena fuera del mutex sin compartir nada con lo guardado ni con otros llamadores. Se verifica con CA-09.9 bajo `go test -race` (lo corre el CI; en Windows local no hay cgo).
- **Errores identificables.** Los errores propagados se comparan con `errors.Is`.
- **Tests.** Unitarios en `internal/backlog/priority_test.go` y pasos de godog en `internal/backlog/priority_features_test.go`, con el runner `TestFeatures_HU09PriorizarBacklog`, sobre `features/HU-09-priorizar-backlog.feature`. El ítem con prioridad desconocida (CA-09.6) se arma guardándolo directo con `repo.Save` del repositorio en memoria, igual que el proyecto `FINALIZADO` de HU-06 se guarda con `projectRepo.Save`. El error del repositorio (CA-09.8) se prueba con `failingRepository{listErr: errRepository}`.
- **Librerías.** Solo la librería estándar en el código (`slices`), y godog solo en los tests.
- **Exposición HTTP.** Es alcance de TEC-04.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Ítems creados en orden BAJA, MEDIA, ALTA | Se devuelven ALTA, MEDIA, BAJA |
| Varios ítems con la misma prioridad, intercalados con otras | Dentro de cada prioridad quedan por ID ascendente |
| Todos los ítems con la misma prioridad | Mismo orden que `ListItems` (por ID ascendente) |
| Un solo ítem | Lista de un ítem |
| Proyecto sin ítems | Lista vacía (longitud 0, no `nil`), sin error |
| Otro proyecto con ítems de mayor prioridad | Esos ítems no aparecen |
| Proyecto `FINALIZADO` con ítems | Se listan ordenados igual |
| Ítems en distintos estados (PENDIENTE, EN_CURSO, TERMINADO) | Aparecen todos; el estado no afecta el orden |
| Ítem guardado con una prioridad desconocida (por ejemplo "URGENTE" o "alta" sin normalizar) | Aparece al final, después de los BAJA; entre varios desconocidos, por ID ascendente |
| Llamar a `ListItems` después de `ListItemsByPriority` | `ListItems` sigue devolviendo el orden por ID |
| Proyecto inexistente (999), 0, negativo (-1) o sin proyectos creados | `ErrProjectNotFound` y lista `nil` |
| Modificar la lista devuelta, un ítem o sus criterios | Los ítems guardados no cambian; volver a listar devuelve los datos y el orden originales |
| Listados por prioridad simultáneos mientras se crean ítems | Ninguno falla y `go test -race` no reporta condiciones de carrera |
| El repositorio falla al listar | Se devuelve ese error y la lista es `nil` |

## Criterios de aceptación

- **CA-09.1** El backlog de un proyecto puede listarse ordenado de mayor a menor prioridad: primero los ítems ALTA, después los MEDIA y después los BAJA, con todos sus datos.
- **CA-09.2** Los ítems con la misma prioridad conservan un orden estable: por ID ascendente (orden de creación).
- **CA-09.3** Un proyecto sin ítems devuelve una lista vacía (longitud 0, no `nil`), no un error.
- **CA-09.4** Se listan todos los ítems del proyecto y solo esos, en cualquier estado e incluso si el proyecto está `FINALIZADO`: los mismos ítems que devuelve `ListItems`.
- **CA-09.5** El listado de HU-06 no cambia: `ListItems` sigue devolviendo los ítems por ID ascendente, también después de listar por prioridad.
- **CA-09.6** Un ítem con una prioridad desconocida no produce error: aparece al final, después de los BAJA, y entre varios desconocidos se conserva el orden por ID.
- **CA-09.7** Listar por prioridad el backlog de un proyecto inexistente devuelve `ErrProjectNotFound` (incluidos ID inexistente, 0, negativo y listar sin proyectos creados) y ninguna lista.
- **CA-09.8** Si el repositorio falla al listar, se devuelve ese error y ninguna lista.
- **CA-09.9** Listar por prioridad mientras se crean ítems en simultáneo no falla ni produce condiciones de carrera (verificado con `go test -race`).
- **CA-09.10** Modificar la lista devuelta, un ítem o sus criterios no altera los ítems guardados ni el orden de un nuevo listado.

## Comparación con el acta de Planning

El acta del Sprint 2 (`docs/actas/sprint-2/planning.md`) describe HU-09 como "Priorizar los ítems del backlog" (Issue #22, prioridad MEDIA, 3 Story Points, responsable Cardozo Leandro), dependiente de "HU-06. El orden por ID de RN-06.6 sigue siendo el predeterminado; el orden por prioridad es una opción (CA-09.1 'puede ordenarse')", y en la estimación como "Orden por prioridad estable, como opción del listado de HU-06". **El alcance coincide:** esta SDD agrega el orden por prioridad como un método aparte (`ListItemsByPriority`), deja `ListItems` y su orden por ID sin cambios (RN-09.5, CA-09.5) y define el orden estable a igual prioridad (RN-09.2, CA-09.2). No hay diferencias entre el acta y esta SDD.

Notas del acta que no cambian el alcance: HU-09 es la primera historia que pasa a "si hay tiempo" si el Sprint se atrasa, y el orden de trabajo preveía su SDD para el 18/10.

## Comparación con los criterios del Issue #22 (borrador)

| CA en el Issue #22 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-09.1 El listado puede ordenarse de mayor a menor prioridad | CA-09.1 (precisado) | Orden ALTA, MEDIA, BAJA, como método aparte de `ListItems` |
| CA-09.2 Los ítems con la misma prioridad conservan un orden estable | CA-09.2 (precisado) | El orden estable es por ID ascendente |
| CA-09.3 Un backlog vacío no produce error | CA-09.3 (precisado) | Longitud 0, no `nil` |
| — | CA-09.4 (nuevo) | Mismos ítems que HU-06: todos los del proyecto y solo esos |
| — | CA-09.5 (nuevo) | El orden por ID de HU-06 no cambia |
| — | CA-09.6 (nuevo) | Prioridad desconocida al final, sin error |
| — | CA-09.7 (nuevo) | Proyecto inexistente con `ErrProjectNotFound` |
| — | CA-09.8 (nuevo) | Error del repositorio propagado |
| — | CA-09.9 (nuevo) | Concurrencia verificada con `go test -race` |
| — | CA-09.10 (nuevo) | Copias por valor |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #22** (se actualiza a mano después de aprobar la SDD):

- [ ] CA-09.1 El backlog de un proyecto puede listarse ordenado de mayor a menor prioridad: primero ALTA, después MEDIA y después BAJA.
- [ ] CA-09.2 Los ítems con la misma prioridad conservan un orden estable: por ID ascendente.
- [ ] CA-09.3 Un proyecto sin ítems devuelve una lista vacía (longitud 0, no `nil`), no un error.
- [ ] CA-09.4 Se listan todos los ítems del proyecto y solo esos, en cualquier estado e incluso si el proyecto está finalizado.
- [ ] CA-09.5 El listado de HU-06 no cambia: sigue ordenado por ID ascendente.
- [ ] CA-09.6 Un ítem con una prioridad desconocida no produce error y aparece al final, conservando el orden por ID.
- [ ] CA-09.7 Listar por prioridad el backlog de un proyecto inexistente devuelve `ErrProjectNotFound`, incluidos ID 0 y negativo.
- [ ] CA-09.8 Si el repositorio falla al listar, se devuelve ese error.
- [ ] CA-09.9 Listar por prioridad mientras se crean ítems en simultáneo no produce condiciones de carrera (verificado con `go test -race`).
- [ ] CA-09.10 Modificar la lista devuelta, un ítem o sus criterios no altera los ítems guardados.

## Decisiones confirmadas por el Agile Enabler el 20/10

1. Método nuevo `ListItemsByPriority(projectID int) ([]Item, error)` en `internal/backlog/priority.go`; `ListItems` y su orden por ID (RN-06.6) no cambian.
2. Orden ALTA, MEDIA, BAJA; a igual prioridad, por ID ascendente, con un ordenamiento estable (`slices.SortStableFunc`) sobre la lista que devuelve `ListItems`.
3. Prioridad desconocida al final, conservando el orden por ID y sin error.
4. La prioridad desconocida no se normaliza al ordenar (por ejemplo, "alta" en minúsculas cuenta como desconocida): `Create` y `Update` ya guardan la prioridad normalizada y la tabla `backlog_items` no tiene `CHECK` sobre `priority`, así que solo aparece con datos guardados sin pasar por el `Manager`. Se prueba guardando el ítem directo con `repo.Save` del repositorio en memoria, como el proyecto `FINALIZADO` de HU-06 con `projectRepo.Save`.
5. Se ordena en el `Manager`, no en el repositorio: la interfaz `Repository`, el esquema de PostgreSQL y los contratos de repositorio de TEC-05 no cambian.
6. Proyecto inexistente: se propaga `ErrProjectNotFound` de HU-03, como en HU-06. Backlog vacío: lista vacía, no `nil`, sin error.
7. Sin mutex nuevo: reutiliza `ListItems`, que toma el mutex del `Manager`, y ordena su propia lista fuera del mutex. Se agrega CA-09.9 para respaldarlo con un test de listados simultáneos con creaciones bajo `go test -race`, como CA-06.5.
8. CA-09.4: el listado por prioridad contiene los mismos ítems que HU-06 (cualquier estado, proyecto `FINALIZADO` y aislamiento por proyecto, RN-06.1 y RN-06.3).
9. CA-09.8: si el repositorio falla al listar, se propaga su error con lista `nil`; se prueba con `failingRepository{listErr: errRepository}`.
10. Sin errores nuevos.
