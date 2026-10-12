# HU-04 · Crear ítem del backlog

> Issue del repositorio: #17

## Historia de usuario

Como Agile Enabler, quiero agregar un ítem al Product Backlog de un proyecto con sus campos obligatorios, para registrar el trabajo pendiente.

## Objetivo

Registrar un ítem nuevo en el Product Backlog de un proyecto existente, con un identificador único, sus datos validados, el estado inicial `PENDIENTE` y los story points sin asignar, para que las demás historias de la épica (consultarlo en HU-05, listarlo en HU-06, modificarlo en HU-07, cambiar su estado en HU-08 y estimarlo en HU-20) tengan sobre qué trabajar.

Esta historia también crea el paquete `internal/backlog`: el tipo `Item`, la interfaz `Repository`, su implementación en memoria y el `Manager` que usan HU-05 y HU-06 (ver "Restricciones").

## Alcance

- **Incluye:** crear un ítem en un proyecto existente a través del `Manager` del paquete `internal/backlog`, validando los campos obligatorios, la prioridad y la existencia del proyecto (reutilizando HU-03).
- **No incluye:** consultar un ítem (HU-05), listar el backlog (HU-06), modificar el ítem (HU-07), cambiar su estado o definir sus transiciones (HU-08), asignar story points (HU-20), registrar horas estimadas (HU-07) ni la exposición HTTP (TEC-01).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ProjectID | entero | Sí | ID del proyecto al que pertenece el ítem |
| Título | texto | Sí | Resumen breve del trabajo |
| Descripción | texto | No | Detalle libre; puede estar vacía |
| Prioridad | texto | Sí | `ALTA`, `MEDIA` o `BAJA` |
| Criterios de aceptación | lista de textos | Sí | Al menos un criterio no vacío |

El ID, el estado y los story points no son entradas: los asigna el sistema.

Firma: `Create(projectID int, title, description, priority string, criteria []string) (Item, error)`.

### Salidas esperadas

- **Si la creación es válida:** el ítem creado, como copia por valor, con:
  - **ID:** entero correlativo global asignado por el repositorio (el primero es 1).
  - **ProjectID:** el recibido.
  - **Título:** el recibido, sin los espacios de los extremos.
  - **Descripción:** la recibida, tal como llegó.
  - **Prioridad:** `ALTA`, `MEDIA` o `BAJA`, en mayúsculas.
  - **Criterios de aceptación:** los recibidos que no quedan vacíos, sin los espacios de los extremos y en el mismo orden.
  - **Estado:** `PENDIENTE`.
  - **StoryPoints:** 0 (sin asignar).
- **Si la creación no es válida:** un error que indica el motivo (ver "Errores"), y no se crea ningún ítem.

### Errores

| Condición | Error | Efecto |
|---|---|---|
| Título vacío o con solo espacios | `ErrEmptyTitle` ("el título del ítem es obligatorio") | No se crea el ítem |
| Prioridad vacía o distinta de `ALTA`, `MEDIA` o `BAJA` | `ErrInvalidPriority` ("la prioridad debe ser ALTA, MEDIA o BAJA") | No se crea el ítem |
| Ningún criterio de aceptación no vacío (lista vacía, `nil` o solo criterios vacíos) | `ErrNoAcceptanceCriteria` ("el ítem necesita al menos un criterio de aceptación") | No se crea el ítem |
| No existe un proyecto con ese ProjectID (incluye 0 y negativos) | `ErrProjectNotFound` de `internal/project`, propagado tal cual | No se crea el ítem |
| El repositorio falla al guardar | El error que devuelve el repositorio | No se devuelve ningún ítem |

**Orden de validación:** título, prioridad, criterios, proyecto y guardado. Se informa el primer error encontrado. Los datos del ítem se validan antes de consultar al proyecto, igual que HU-01 valida el nombre antes de consultar los repetidos.

## Reglas de negocio

1. **RN-04.1 Título obligatorio.** El título no puede estar vacío ni tener solo espacios, y se guarda sin los espacios de los extremos. Se consideran espacios todos los espacios en blanco Unicode (`strings.TrimSpace`), como en HU-01. Puede repetirse dentro del mismo proyecto: no hay ninguna regla que lo impida.
2. **RN-04.2 Descripción opcional.** La descripción puede estar vacía y se guarda tal como llega, sin longitud máxima (como RN-01.4).
3. **RN-04.3 Prioridad.** Se aceptan `ALTA`, `MEDIA` y `BAJA`, sin distinguir mayúsculas y sin contar los espacios de los extremos (" alta " es válida). Se guarda en mayúsculas. Un valor vacío o cualquier otro valor es un error.
4. **RN-04.4 Criterios de aceptación.** Son una lista de textos. Los criterios vacíos o con solo espacios se descartan, y los demás se guardan sin los espacios de los extremos y en el orden recibido. Después de descartar, tiene que quedar al menos uno.
5. **RN-04.5 Proyecto existente.** El ítem pertenece a un proyecto que tiene que existir (ver "Decisión sobre entidades referenciadas"). Se pueden agregar ítems a un proyecto en cualquier estado, incluido `FINALIZADO`: no hay ninguna regla que lo impida.
6. **RN-04.6 Identificador único.** Cada ítem recibe un ID entero correlativo global que empieza en 1, asignado por el repositorio al guardar. Es global y no por proyecto, para que HU-05 pueda consultar un ítem solo por su ID.
7. **RN-04.7 Estados del ítem.** Un ítem puede estar en `PENDIENTE`, `EN_CURSO` o `TERMINADO`. Todo ítem nuevo se crea en `PENDIENTE`. Las transiciones entre estados las define HU-08.
8. **RN-04.8 Story points sin asignar.** Todo ítem nuevo se crea con `StoryPoints` en 0, que significa "sin asignar". Es un valor seguro porque la escala de HU-20 (1, 2, 3, 5, 8, 13 y 21) no incluye el 0. Los story points se asignan en HU-20.
9. **RN-04.9 Copias por valor.** La creación devuelve una copia del ítem. Ni el ítem devuelto ni la lista de criterios recibida comparten memoria con el ítem guardado: modificarlos no altera lo guardado.

## Decisión sobre entidades referenciadas

HU-04 referencia una entidad por ID: el proyecto al que pertenece el ítem. **Sí se valida su existencia**, por estos motivos:

1. Un ítem sin proyecto quedaría huérfano: no aparecería en ningún backlog (HU-06) y no se podría planificar.
2. Ya existe la búsqueda por ID (HU-03). La SDD de HU-03 establece que HU-04 la reutiliza: el `Manager` de backlog llama a `project.Manager.Get`, que devuelve `ErrProjectNotFound`. No se duplica ni la regla ni el error.

## Restricciones

- **Organización del código.** Paquete `internal/backlog`, siguiendo el patrón de HU-01:
  - `doc.go`: documentación del paquete.
  - `item.go`: el tipo `Item`, `Priority`, `Status` y los errores.
  - `repository.go`: la interfaz `Repository`.
  - `memory_repository.go`: la implementación en memoria.
  - `manager.go`: el `Manager` y `NewManager`.
  - `create.go`: la creación (HU-04).

  HU-05 y HU-06 agregan sus propios archivos (`get.go`, `list.go`). No se toca el paquete `internal/project`.
- **Constructor.** `NewManager(repo Repository, projects *project.Manager) *Manager`. Es el mismo que espera la SDD de HU-06.
- **Interfaz de repositorio.** HU-04 define solo la operación que necesita: `Save(item Item) (Item, error)`, que asigna el ID y devuelve una copia. HU-05 agrega `GetByID` y HU-06 agrega `ListByProject`. La lógica de negocio accede a los ítems solo a través de esta interfaz. La implementación en memoria la cumple, y TEC-05 (PostgreSQL) también deberá cumplirla.
- **Copias por valor.** `Item` tiene un campo de tipo lista (`AcceptanceCriteria []string`), y copiar un `Item` no copia esa lista. Por eso:
  - el `Manager` arma una lista nueva con los criterios filtrados, sin reutilizar la recibida;
  - el repositorio en memoria guarda y devuelve una copia de la lista (`slices.Clone`).

  Las demás operaciones (HU-05 y HU-06) deben devolver también una copia de la lista.
- **Verificación de lo guardado.** Como HU-04 no define operaciones de lectura (son de HU-05 y HU-06), los tests que verifican lo guardado (cantidad de ítems, datos y copias) son tests internos del paquete (`package backlog`): miran el repositorio en memoria sin agregar operaciones que no pide ningún criterio. Los pasos de godog también van en un test interno del paquete, en `create_features_test.go`, con su runner `TestFeatures_HU04CrearItem`.
- **Errores identificables.** Cada error está exportado y se puede comparar con `errors.Is`. `ErrProjectNotFound` se propaga tal cual desde `internal/project`.
- **Concurrencia.** El repositorio en memoria no es seguro para uso concurrente. El `Manager` protege el guardado con un `sync.Mutex`, y la protección se justifica con CA-04.7 y su test bajo `go test -race`. La validación del proyecto se hace antes de tomar el mutex de backlog, porque no existe ninguna operación que borre proyectos, y así no se toman dos mutex a la vez.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Exposición HTTP.** Es alcance de TEC-01.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Primer ítem del sistema | Recibe el ID 1, en estado `PENDIENTE` y con `StoryPoints` en 0 |
| Ítems en dos proyectos distintos | Los IDs siguen siendo correlativos y globales (1, 2, 3…) |
| Título con espacios en los extremos ("  Login  ") | Se guarda "Login" |
| Título con solo espacios | `ErrEmptyTitle` |
| Título igual al de otro ítem del mismo proyecto | Se crea |
| Descripción vacía | Se crea con la descripción vacía |
| Prioridad "alta" o " Media " | Se guarda `ALTA` o `MEDIA` |
| Prioridad vacía o "URGENTE" | `ErrInvalidPriority` |
| Criterios `["Login OK", "  ", ""]` | Se guarda `["Login OK"]` |
| Criterios `nil`, `[]` o `["  "]` | `ErrNoAcceptanceCriteria` |
| ProjectID inexistente, 0 o negativo | `ErrProjectNotFound` |
| Proyecto `FINALIZADO` | Se crea el ítem |
| Varios datos inválidos a la vez (título vacío y prioridad inválida) | Se informa el primero según el orden de validación (`ErrEmptyTitle`) |
| Modificar el ítem devuelto o sus criterios | El ítem guardado no cambia |
| Modificar la lista de criterios después de crear | El ítem guardado no cambia |
| Varias creaciones simultáneas | Todas se crean con IDs distintos; `go test -race` no reporta condiciones de carrera |

## Criterios de aceptación

- **CA-04.1** Se crea un ítem con título, descripción, prioridad y criterios de aceptación en un proyecto existente, y se devuelve con esos datos (título y criterios sin los espacios de los extremos).
- **CA-04.2** Los campos obligatorios se validan: título vacío o con solo espacios devuelve `ErrEmptyTitle`, y no tener ningún criterio de aceptación no vacío devuelve `ErrNoAcceptanceCriteria`. En los dos casos no se crea el ítem. La descripción es opcional.
- **CA-04.3** La prioridad solo admite `ALTA`, `MEDIA` o `BAJA`, sin distinguir mayúsculas, y se guarda en mayúsculas. Un valor vacío o distinto devuelve `ErrInvalidPriority` y no se crea el ítem.
- **CA-04.4** El ítem queda con un ID único correlativo (el primero es 1), en estado `PENDIENTE` y con los story points sin asignar (0).
- **CA-04.5** Crear un ítem en un proyecto que no existe devuelve `ErrProjectNotFound` (incluidos ID 0 y negativo) y no se crea el ítem.
- **CA-04.6** Modificar el ítem devuelto, sus criterios o la lista de criterios recibida no altera el ítem guardado.
- **CA-04.7** Varias creaciones simultáneas reciben IDs distintos, y el test de concurrencia corre con `go test -race` sin reportar condiciones de carrera.
- **CA-04.8** Si el repositorio falla al guardar, se devuelve ese error y ningún ítem.

## Comparación con el acta de Planning

El acta del Sprint 1 describe HU-04 como "Crear ítem del backlog (con validación de campos obligatorios)" (Issue #17, 5 Story Points, responsable García Santiago, orden 5), con la nota "Más campos y validaciones; relación con el proyecto y estados del ítem", y la ubica antes de HU-06 porque HU-06 lista los ítems que crea. **El alcance coincide:** esta SDD valida los campos obligatorios, resuelve la relación con el proyecto (RN-04.5) y define los estados del ítem (RN-04.7), un pendiente que el Sprint 0 había dejado para las SDD. No hay diferencias entre el acta y esta SDD.

## Comparación con los criterios del Issue #17 (borrador)

| CA en el Issue #17 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-04.1 Se crea un ítem con título, descripción, prioridad y criterios de aceptación, asociado a un proyecto | CA-04.1 (precisado) | El proyecto tiene que existir; título y criterios se guardan sin espacios en los extremos |
| CA-04.2 Faltar un campo obligatorio devuelve error | CA-04.2 (precisado) | Se identifican los errores; la descripción pasa a ser opcional |
| CA-04.3 La prioridad solo admite ALTA, MEDIA o BAJA | CA-04.3 (precisado) | Sin distinguir mayúsculas; se guarda en mayúsculas |
| CA-04.4 Identificador único, estado inicial y story points sin asignar | CA-04.4 (precisado) | ID correlativo, estado `PENDIENTE` y story points en 0 |
| — | CA-04.5 (nuevo) | Validación de que el proyecto exista |
| — | CA-04.6 (nuevo) | Copias por valor, incluida la lista de criterios |
| — | CA-04.7 (nuevo) | Concurrencia verificada con `go test -race` |
| — | CA-04.8 (nuevo) | Error del repositorio propagado |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #17** (cuando se apruebe la SDD):

- [ ] CA-04.1 Se crea un ítem con título, descripción, prioridad y criterios de aceptación en un proyecto existente.
- [ ] CA-04.2 Título vacío o sin ningún criterio de aceptación devuelve error (`ErrEmptyTitle`, `ErrNoAcceptanceCriteria`); la descripción es opcional.
- [ ] CA-04.3 La prioridad solo admite ALTA, MEDIA o BAJA (sin distinguir mayúsculas); otro valor devuelve `ErrInvalidPriority`.
- [ ] CA-04.4 El ítem queda con un ID único correlativo, en estado PENDIENTE y con los story points sin asignar.
- [ ] CA-04.5 Crear un ítem en un proyecto inexistente devuelve `ErrProjectNotFound`.
- [ ] CA-04.6 Modificar el ítem devuelto o la lista de criterios recibida no altera el ítem guardado.
- [ ] CA-04.7 Creaciones simultáneas reciben IDs distintos (verificado con `go test -race`).
- [ ] CA-04.8 Si el repositorio falla al guardar, se devuelve ese error.

## Decisiones tomadas (dudas resueltas con el Agile Enabler)

1. Obligatorios: título, prioridad y criterios de aceptación. Descripción opcional.
2. Criterios de aceptación: lista de textos, con al menos uno no vacío; los vacíos se descartan.
3. Estados: `PENDIENTE` (inicial), `EN_CURSO` y `TERMINADO`; las transiciones las define HU-08.
4. ID global y correlativo, asignado por el repositorio.
5. El título puede repetirse dentro de un proyecto.
6. Se valida que el proyecto exista (`Get` de HU-03, `ErrProjectNotFound`); un proyecto `FINALIZADO` acepta ítems.
7. Prioridad sin distinguir mayúsculas ni contar los espacios de los extremos; se guarda en mayúsculas.
8. `StoryPoints int` en 0 = sin asignar (la escala de HU-20 no incluye el 0); las horas estimadas quedan para HU-07.
9. Mutex sí, justificado por CA-04.7 y su test con `-race`.
