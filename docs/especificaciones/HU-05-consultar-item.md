# HU-05 · Consultar un ítem del backlog

> Issue del repositorio: #18

## Historia de usuario

Como Product Builder, quiero consultar un ítem del backlog por su identificador, para ver su detalle.

## Objetivo

Obtener un ítem del Product Backlog a partir de su ID, con todos sus datos, para que el Product Builder (y más adelante TEC-01) pueda ver su detalle. Es la búsqueda por ID que van a reutilizar HU-07 (modificar un ítem) y HU-08 (cambiar su estado).

Es una operación de solo lectura: no crea, modifica ni borra ítems.

## Alcance

- **Incluye:** consultar un ítem por su ID a través del `Manager` del paquete `internal/backlog` (creado en HU-04), devolviendo una copia del ítem con todos sus datos.
- **No incluye:** crear ítems (HU-04), listar el backlog de un proyecto (HU-06), modificar el ítem (HU-07), cambiar su estado (HU-08) ni la exposición HTTP (TEC-01).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | Identificador del ítem a consultar |

El ID del ítem es global y correlativo (RN-04.6), así que alcanza con él: no hace falta el ID del proyecto.

Firma: `GetItem(id int) (Item, error)`.

### Salidas esperadas

- **Si el ítem existe:** una copia del ítem con todos los campos que define HU-04 (ID, ProjectID, título, descripción, prioridad, criterios de aceptación, estado y story points). La lista de criterios es una copia propia.
- **Si no existe o hay un error:** el error (ver "Errores") y un ítem vacío (`Item{}`).

### Errores

| Condición | Error | Efecto |
|---|---|---|
| No existe un ítem con ese ID (incluye 0, negativos y consultar sin ítems creados) | `ErrItemNotFound` ("no existe un ítem del backlog con ese id") | Se devuelve un ítem vacío |
| El repositorio falla al buscar | El error que devuelve el repositorio | Se devuelve un ítem vacío |

## Reglas de negocio

1. **RN-05.1 Búsqueda por identificador.** Se devuelve el único ítem cuyo ID coincide exactamente con el recibido. Los IDs los asigna el repositorio (RN-04.6); la consulta no los transforma.
2. **RN-05.2 Ítem inexistente.** Si ningún ítem guardado tiene el ID recibido, se devuelve `ErrItemNotFound`. Incluye IDs que nunca se asignaron, 0 y negativos, con el mismo criterio de RN-03.2: en todos los casos el motivo es que no existe el ítem. No se define un error aparte para IDs inválidos.
3. **RN-05.3 Copia por valor.** La consulta devuelve una copia del ítem, con su propia copia de la lista de criterios de aceptación. Modificar el ítem devuelto o sus criterios no altera el ítem guardado.
4. **RN-05.4 Solo lectura.** La consulta no cambia el ítem, su estado ni ningún otro dato guardado.

## Decisión sobre entidades referenciadas

HU-05 referencia una entidad por ID: el ítem consultado. **Sí se valida su existencia:** si no existe, se devuelve `ErrItemNotFound` y no un ítem vacío sin error. Es la condición de error central de la historia (CA-05.2).

**No se valida el proyecto del ítem.** El ítem ya tiene su `ProjectID`, y HU-04 solo lo crea si el proyecto existe; como no hay ninguna operación que borre proyectos, el proyecto de un ítem guardado siempre existe. Validarlo de nuevo sería una consulta extra que ningún criterio pide.

## Restricciones

- **Organización del código.** Se agrega el archivo `internal/backlog/get.go`, con el error `ErrItemNotFound` y el método `GetItem` del `Manager`. No se toca el paquete `internal/project` ni el código de producción de HU-04 y HU-06, salvo lo indicado en "Interfaz de repositorio".
- **Interfaz de repositorio.** Se agrega a la interfaz `Repository` la operación `GetByID(id int) (Item, error)`, que devuelve una copia del ítem (incluida la lista de criterios) o `ErrItemNotFound` si no existe. La implementa el repositorio en memoria (`memory_repository.go`), y TEC-05 también deberá implementarla. Es **la** búsqueda por ID del paquete: HU-07 y HU-08 la reutilizan.
- **Consecuencia de ampliar la interfaz.** Agregar `GetByID` obliga a actualizar el repositorio falso de los tests (`failingRepository` en `create_test.go`) para que siga cumpliendo la interfaz, como pasó en HU-03 y HU-06. Se agrega en el commit RED, junto con los tests. Además de `repository.go` y `memory_repository.go`, ese es el único archivo existente que se toca.
- **Concurrencia.** El repositorio en memoria no es seguro para uso concurrente. `GetItem` toma el mutex del `Manager` para buscar. La protección se justifica con CA-05.4 y su test bajo `go test -race`.
- **Tests.** Igual que en HU-04 y HU-06, los tests son internos al paquete (`package backlog`): tests unitarios en `get_test.go`, y pasos de godog en `get_features_test.go`, con el runner `TestFeatures_HU05ConsultarItem`.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Exposición HTTP.** Es alcance de TEC-01.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Ítem existente | Se devuelve con todos sus datos |
| Varios ítems en varios proyectos | Se devuelve solo el ítem con ese ID |
| ID inexistente (999), 0 o negativo (-1) | `ErrItemNotFound` |
| Consultar sin ítems creados | `ErrItemNotFound` |
| Modificar el ítem devuelto o sus criterios | El ítem guardado no cambia; volver a consultar devuelve los datos originales |
| Consultas simultáneas mientras se crean ítems | `go test -race` no reporta condiciones de carrera |
| El repositorio falla al buscar | Se devuelve ese error y un ítem vacío |

## Criterios de aceptación

- **CA-05.1** Se obtiene el ítem por su ID, con todos sus datos (ID, proyecto, título, descripción, prioridad, criterios de aceptación, estado y story points).
- **CA-05.2** Consultar un ítem que no existe devuelve `ErrItemNotFound` (incluidos ID inexistente, 0, negativo y consultar sin ítems creados).
- **CA-05.3** Modificar el ítem devuelto o su lista de criterios no altera el ítem guardado; volver a consultar devuelve los datos originales.
- **CA-05.4** Consultar mientras se crean ítems en simultáneo no produce accesos concurrentes sin protección: el test corre con `go test -race` (el CI ya lo ejecuta) sin reportar condiciones de carrera.
- **CA-05.5** Si el repositorio falla al buscar, se devuelve ese error y un ítem vacío.

## Comparación con el acta de Planning

El acta del Sprint 1 menciona HU-05 (Issue #18, "consultar un ítem") como historia opcional, que se suma durante el Sprint si el equipo llega con tiempo, y asigna las opcionales "a quien tenga menos carga o termine antes, empezando por Sosa Ricardo y Buttini Cristobal". El 12/10 se asignó a Sosa Ricardo, que terminó HU-03 antes de lo previsto. El acta no describe el alcance de HU-05, así que no hay nada que comparar.

## Comparación con los criterios del Issue #18 (borrador)

| CA en el Issue #18 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-05.1 Se obtiene el ítem con todos sus datos | CA-05.1 (precisado) | Se listan los datos y se aclara que se consulta por ID |
| CA-05.2 Consultar un ítem inexistente devuelve error | CA-05.2 (precisado) | Se identifica el error (`ErrItemNotFound`) y se incluyen ID 0, negativo y sin ítems |
| CA-05.3 Modificar el resultado devuelto no altera el ítem guardado | CA-05.3 (precisado) | Incluye la lista de criterios |
| — | CA-05.4 (nuevo) | Concurrencia verificada con `go test -race` |
| — | CA-05.5 (nuevo) | Error del repositorio propagado |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #18** (cuando se apruebe la SDD):

- [ ] CA-05.1 Se obtiene el ítem por su ID, con todos sus datos.
- [ ] CA-05.2 Consultar un ítem que no existe devuelve `ErrItemNotFound`, incluidos ID 0, negativo y sin ítems creados.
- [ ] CA-05.3 Modificar el ítem devuelto o sus criterios no altera el ítem guardado.
- [ ] CA-05.4 Consultar mientras se crean ítems en simultáneo no produce condiciones de carrera (verificado con `go test -race`).
- [ ] CA-05.5 Si el repositorio falla al buscar, se devuelve ese error.

## Decisiones tomadas (dudas resueltas con el Agile Enabler)

1. Consulta solo por el ID del ítem, que es global (RN-04.6): `GetItem(id int) (Item, error)` en el `Manager` y `GetByID(id int) (Item, error)` en la interfaz `Repository`.
2. Error nuevo `ErrItemNotFound`, declarado en `get.go`; IDs 0 y negativos dan el mismo error (RN-05.2).
3. No se valida el proyecto del ítem: HU-04 ya lo validó al crearlo y no existe ninguna operación que borre proyectos.
4. Copia por valor, incluida la lista de criterios (RN-05.3).
5. Mutex sí, justificado por CA-05.4 y su test con `-race` (no por uniformidad).
