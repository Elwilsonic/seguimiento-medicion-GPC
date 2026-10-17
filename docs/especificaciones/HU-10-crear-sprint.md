# HU-10 · Crear sprint con nombre, fechas y Sprint Goal

> Issue del repositorio: #24

## Historia de usuario

Como Agile Enabler, quiero crear un sprint con nombre, fechas y Sprint Goal, para planificar una iteración del proyecto.

## Objetivo

Registrar un sprint nuevo asociado a un proyecto existente, con un identificador único, fechas válidas, Sprint Goal y el estado inicial `ABIERTO`, para que las demás historias de la épica (asignar historias en HU-11, registrar completadas en HU-12, cerrar el sprint en HU-13 y consultar sprints anteriores en HU-14) tengan sobre qué trabajar.

Esta historia también crea el paquete `internal/sprint`: el tipo `Sprint`, la interfaz `Repository`, su implementación en memoria y el `Manager` (ver "Restricciones").

## Alcance

- **Incluye:** crear un sprint en un proyecto existente a través del `Manager` del paquete `internal/sprint`, validando el nombre, el Sprint Goal, las fechas, la existencia del proyecto (reutilizando HU-03) y que no haya otro sprint abierto en el mismo proyecto.
- **No incluye:** asignar historias al sprint (HU-11), registrar historias completadas (HU-12), cerrar el sprint (HU-13), consultar sprints anteriores (HU-14), reabrir un sprint cerrado ni la exposición HTTP (TEC-04).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ProjectID | entero | Sí | ID del proyecto al que pertenece el sprint |
| Nombre | texto | Sí | Nombre del sprint |
| Fecha de inicio | fecha (`time.Time`) | Sí | Inicio del sprint |
| Fecha de fin | fecha (`time.Time`) | Sí | Fin del sprint |
| Sprint Goal | texto | Sí | Objetivo del sprint |

El ID y el estado no son entradas: los asigna el sistema.

Firma: `Create(projectID int, name string, startDate, endDate time.Time, goal string) (Sprint, error)`.

### Salidas esperadas

- **Si la creación es válida:** el sprint creado, como copia por valor, con:
  - **ID:** entero correlativo global asignado por el repositorio (el primero es 1).
  - **ProjectID:** el recibido.
  - **Nombre:** el recibido, sin los espacios de los extremos.
  - **Fechas:** las recibidas, tal como llegaron.
  - **Sprint Goal:** el recibido, sin los espacios de los extremos.
  - **Estado:** `ABIERTO`.
- **Si la creación no es válida:** un error que indica el motivo (ver "Errores"), y no se crea ningún sprint.

### Errores

| Condición | Error |
|---|---|
| Nombre vacío o con solo espacios | `ErrEmptyName` ("el nombre del sprint es obligatorio") |
| Sprint Goal vacío o con solo espacios | `ErrEmptyGoal` ("el sprint goal es obligatorio") |
| Fecha de inicio ausente (valor zero de `time.Time`) | `ErrMissingStartDate` ("la fecha de inicio del sprint es obligatoria") |
| Fecha de fin ausente (valor zero de `time.Time`) | `ErrMissingEndDate` ("la fecha de fin del sprint es obligatoria") |
| Fecha de fin igual o anterior a la de inicio | `ErrEndNotAfterStart` ("la fecha de fin debe ser posterior a la de inicio") |
| No existe un proyecto con ese ProjectID (incluye 0 y negativos) | `ErrProjectNotFound` de `internal/project`, propagado tal cual |
| Ya existe un sprint `ABIERTO` en ese proyecto | `ErrOpenSprintExists` ("ya existe un sprint abierto en el proyecto") |
| El repositorio falla al guardar | El error que devuelve el repositorio |

**Orden de validación:** nombre, Sprint Goal, fecha de inicio, fecha de fin, relación entre fechas, proyecto, sprint abierto y guardado. Se informa el primer error encontrado. Los datos del sprint se validan antes de consultar al proyecto, igual que en HU-04.

## Reglas de negocio

1. **RN-10.1 Nombre obligatorio.** El nombre no puede estar vacío ni tener solo espacios, y se guarda sin los espacios de los extremos (todos los espacios en blanco Unicode de `strings.TrimSpace`, como RN-01.1 y RN-01.2). Puede repetirse, en el mismo proyecto o en otro: no hay ninguna regla que lo impida.
2. **RN-10.2 Sprint Goal obligatorio.** El Sprint Goal sigue la misma regla que el nombre: obligatorio, sin espacios en los extremos. CA-10.1 lo lista como parte de la creación.
3. **RN-10.3 Fechas.** Las dos fechas son obligatorias: el valor zero de `time.Time` se trata como fecha ausente. La fecha de fin debe ser estrictamente posterior a la de inicio (CA-10.2): fin igual a inicio es un error. Las fechas se comparan como instantes (`After`). La capa de afuera (TEC-04) recibe días (`YYYY-MM-DD`), así que en la práctica fin igual a inicio significa el mismo día.
4. **RN-10.4 Proyecto existente.** El sprint pertenece a un proyecto que tiene que existir (ver "Decisión sobre entidades referenciadas"). Se puede crear un sprint en un proyecto en cualquier estado, incluido `FINALIZADO`: no hay ninguna regla que lo impida (mismo criterio que RN-04.5 en HU-04).
5. **RN-10.5 Un solo sprint abierto por proyecto.** No puede haber más de un sprint `ABIERTO` por proyecto (CA-10.3). Dos proyectos distintos pueden tener cada uno su sprint abierto: la regla es por proyecto.
6. **RN-10.6 Identificador único.** Cada sprint recibe un ID entero correlativo global que empieza en 1, asignado por el repositorio al guardar. Es global y no por proyecto, con el mismo criterio que RN-04.6 en HU-04.
7. **RN-10.7 Estados del sprint.** Un sprint puede estar `ABIERTO` o `CERRADO`. Todo sprint nuevo se crea `ABIERTO`. El cambio a `CERRADO` lo define HU-13.
8. **RN-10.8 Copias por valor.** La creación devuelve una copia del sprint. `Sprint` no tiene slices ni mapas: copiarlo no comparte estado con lo guardado.

## Decisión sobre entidades referenciadas

HU-10 referencia una entidad por ID: el proyecto al que pertenece el sprint. **Sí se valida su existencia**, por estos motivos:

1. Un sprint sin proyecto quedaría huérfano: HU-11 no podría asignarle historias y el sprint quedaría fuera de toda consulta del proyecto.
2. Ya existe la búsqueda por ID (HU-03), y HU-04 y HU-06 ya la usan: el `Manager` de sprint llama a `project.Manager.Get`, que devuelve `ErrProjectNotFound`. No se duplica ni la regla ni el error.

## Restricciones

- **Organización del código.** Paquete `internal/sprint`, siguiendo el patrón de HU-01 y HU-04:
  - `doc.go`: documentación del paquete.
  - `sprint.go`: el tipo `Sprint`, `Status` y los errores.
  - `repository.go`: la interfaz `Repository`.
  - `memory_repository.go`: la implementación en memoria.
  - `manager.go`: el `Manager` y `NewManager`.
  - `create.go`: la creación (HU-10).

  No se toca el paquete `internal/project` ni `internal/backlog`.
- **Constructor.** `NewManager(repo Repository, projects *project.Manager) *Manager`, el mismo patrón de HU-04.
- **Interfaz de repositorio.** HU-10 define las operaciones que necesita: `Save(sprint Sprint) (Sprint, error)` (asigna el ID y devuelve una copia) y `OpenByProject(projectID int) (Sprint, bool, error)` (devuelve una copia del sprint `ABIERTO` del proyecto y si existe). HU-11 a HU-14 agregan las suyas. TEC-05 (PostgreSQL) también deberá cumplirla.
- **Concurrencia.** El repositorio en memoria no es seguro para uso concurrente. Crear hace dos pasos seguidos: consultar el sprint abierto y guardar. El `Manager` protege ambos con su `sync.Mutex`, para que dos creaciones simultáneas en el mismo proyecto no dejen dos sprints abiertos. La validación del proyecto se hace antes de tomar el mutex, para no tomar dos mutex a la vez (igual que HU-04). La protección se justifica con CA-10.6 y su test bajo `go test -race` (lo corre el CI; en Windows local no hay cgo).
- **Verificación de lo guardado.** Como HU-10 no define operaciones de lectura (HU-11 a HU-14), los tests que verifican lo guardado son internos al paquete (`package sprint`): miran el repositorio en memoria sin agregar operaciones que no pide ningún criterio, igual que HU-04. Tests unitarios en `create_test.go`; pasos de godog en `create_features_test.go` con el runner `TestFeatures_HU10CrearSprint`.
- **Errores identificables.** Cada error está exportado y se puede comparar con `errors.Is`. `ErrProjectNotFound` se propaga tal cual desde `internal/project`.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Exposición HTTP.** Es alcance de TEC-04.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Primer sprint del sistema | Recibe el ID 1, en estado `ABIERTO` |
| Sprints en dos proyectos distintos | Los IDs siguen siendo correlativos y globales (1, 2, 3…) |
| Nombre con espacios en los extremos ("  Sprint 1  ") | Se guarda "Sprint 1" |
| Nombre con solo espacios | `ErrEmptyName` |
| Sprint Goal con espacios en los extremos | Se guarda sin los espacios de los extremos |
| Sprint Goal con solo espacios | `ErrEmptyGoal` |
| Nombre igual al de otro sprint (mismo u otro proyecto) | Se crea (si no hay otro sprint abierto en el mismo proyecto) |
| Fecha de fin igual a la de inicio | `ErrEndNotAfterStart` |
| Fecha de fin anterior a la de inicio | `ErrEndNotAfterStart` |
| Fecha de inicio o de fin ausente (zero) | `ErrMissingStartDate` o `ErrMissingEndDate` |
| ProjectID inexistente, 0 o negativo | `ErrProjectNotFound` |
| Proyecto `FINALIZADO` | Se crea el sprint |
| Ya existe un sprint `ABIERTO` en el proyecto | `ErrOpenSprintExists` |
| Otro proyecto tiene un sprint `ABIERTO` | Se crea el sprint en este proyecto |
| Varios datos inválidos a la vez (nombre vacío y fechas inválidas) | Se informa el primero según el orden de validación (`ErrEmptyName`) |
| Modificar el sprint devuelto | El sprint guardado no cambia |
| Creaciones simultáneas en el mismo proyecto | Exactamente una tiene éxito y las demás devuelven `ErrOpenSprintExists`; queda un solo sprint `ABIERTO` |
| Creaciones simultáneas en proyectos distintos | Todas se crean, con IDs distintos |
| El repositorio falla al guardar | Se devuelve ese error y ningún sprint |

## Criterios de aceptación

- **CA-10.1** Se crea un sprint con nombre, fecha de inicio, fecha de fin y Sprint Goal en un proyecto existente, con ID único correlativo (el primero es 1), estado `ABIERTO` y nombre y Sprint Goal sin espacios en los extremos.
- **CA-10.2** La fecha de fin debe ser estrictamente posterior a la de inicio: fin igual o anterior devuelve `ErrEndNotAfterStart`, y una fecha ausente devuelve `ErrMissingStartDate` o `ErrMissingEndDate`. En ningún caso se crea el sprint.
- **CA-10.3** No puede haber más de un sprint `ABIERTO` por proyecto: si ya existe uno, se devuelve `ErrOpenSprintExists` y no se crea el sprint. Un sprint abierto en otro proyecto no lo impide.
- **CA-10.4** Nombre o Sprint Goal vacíos (o con solo espacios) devuelven `ErrEmptyName` o `ErrEmptyGoal` y no se crea el sprint.
- **CA-10.5** Crear un sprint en un proyecto que no existe (incluidos ID 0 y negativo) devuelve `ErrProjectNotFound` y no se crea el sprint.
- **CA-10.6** De varias creaciones simultáneas en el mismo proyecto, exactamente una tiene éxito y las demás devuelven `ErrOpenSprintExists` (verificado con `go test -race`).
- **CA-10.7** Modificar el sprint devuelto no altera el sprint guardado.
- **CA-10.8** Si el repositorio falla al guardar, se devuelve ese error y no se crea ningún sprint.

## Comparación con el acta de Planning

El acta del Sprint 2 describe HU-10 como "Crear sprint con nombre, fechas y Sprint Goal" (Issue #24, 5 Story Points, responsable Buttini Cristobal), dependiente de HU-01 y HU-03, con la nota "Como HU-04: paquete nuevo con entidad, repositorio y `Manager`, validación de fechas y un solo sprint abierto por proyecto" y "Crea el paquete de sprints, con su estado abierto o cerrado (lo necesitan CA-10.3, CA-11.3 y HU-13)". **El alcance coincide:** esta SDD crea el paquete con entidad, repositorio y `Manager` (patrón HU-04), valida las fechas (RN-10.3) y el sprint abierto único (RN-10.5), y define los estados `ABIERTO`/`CERRADO` (RN-10.7) que usan HU-13 y CA-11.3. La dependencia de HU-03 se resuelve reutilizando `Get` para validar el proyecto. No hay diferencias entre el acta y esta SDD.

## Comparación con los criterios del Issue #24 (borrador)

| CA en el Issue #24 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-10.1 Se crea un sprint asociado a un proyecto, con nombre, fecha de inicio, fecha de fin y Sprint Goal | CA-10.1 (precisado) | El proyecto tiene que existir; ID correlativo, estado `ABIERTO`, nombre y Sprint Goal sin espacios en los extremos |
| CA-10.2 La fecha de fin debe ser posterior a la de inicio | CA-10.2 (precisado) | Estrictamente posterior; las fechas ausentes también son error |
| CA-10.3 No puede haber más de un sprint abierto por proyecto | CA-10.3 (precisado) | Se identifica el error; un sprint abierto en otro proyecto no lo impide |
| — | CA-10.4 (nuevo) | Nombre y Sprint Goal obligatorios |
| — | CA-10.5 (nuevo) | Validación de que el proyecto exista |
| — | CA-10.6 (nuevo) | Concurrencia: una sola creación tiene éxito, verificado con `go test -race` |
| — | CA-10.7 (nuevo) | Copias por valor |
| — | CA-10.8 (nuevo) | Error del repositorio propagado |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #24** (se actualiza a mano después de aprobar la SDD):

- [ ] CA-10.1 Se crea un sprint con nombre, fecha de inicio, fecha de fin y Sprint Goal en un proyecto existente, con ID único correlativo (el primero es 1), estado ABIERTO y nombre y Sprint Goal sin espacios en los extremos.
- [ ] CA-10.2 La fecha de fin debe ser estrictamente posterior a la de inicio: fin igual o anterior devuelve `ErrEndNotAfterStart`, y una fecha ausente devuelve `ErrMissingStartDate` o `ErrMissingEndDate`. En ningún caso se crea el sprint.
- [ ] CA-10.3 No puede haber más de un sprint ABIERTO por proyecto: si ya existe uno, se devuelve `ErrOpenSprintExists` y no se crea el sprint. Un sprint abierto en otro proyecto no lo impide.
- [ ] CA-10.4 Nombre o Sprint Goal vacíos (o con solo espacios) devuelven `ErrEmptyName` o `ErrEmptyGoal` y no se crea el sprint.
- [ ] CA-10.5 Crear un sprint en un proyecto que no existe (incluidos ID 0 y negativo) devuelve `ErrProjectNotFound` y no se crea el sprint.
- [ ] CA-10.6 De varias creaciones simultáneas en el mismo proyecto, exactamente una tiene éxito y las demás devuelven `ErrOpenSprintExists` (verificado con `go test -race`).
- [ ] CA-10.7 Modificar el sprint devuelto no altera el sprint guardado.
- [ ] CA-10.8 Si el repositorio falla al guardar, se devuelve ese error y no se crea ningún sprint.

## Decisiones confirmadas por el Agile Enabler el 17/10

1. Fechas como `time.Time`; el valor zero significa fecha ausente. El formato de texto (`YYYY-MM-DD`) lo resuelven TEC-04 (HTTP) y TEC-05 (PostgreSQL) en la capa de afuera.
2. Sprint Goal obligatorio, sin espacios en los extremos.
3. El nombre puede repetirse: el Issue no pide que sea único y la identidad del sprint es su ID.
4. "Posterior" de CA-10.2 es estricto: fin igual a inicio es error.
5. Se valida que el proyecto exista (`Get` de HU-03, `ErrProjectNotFound`); un proyecto `FINALIZADO` acepta sprints.
6. Un solo sprint `ABIERTO` por proyecto, con la operación `OpenByProject` en el repositorio; consulta y guardado bajo el mutex del `Manager`.
7. Estados `ABIERTO` (inicial) y `CERRADO`; el cambio lo define HU-13.
8. Mutex sí, justificado por CA-10.6 y su test con `-race`.
