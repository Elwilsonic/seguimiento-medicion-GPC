# HU-27 — Consultar los registros de esfuerzo

> Issue del repositorio: #45

Aprobada por: Peñalbé Hernán (Agile Enabler), 25/10/2026, con los ajustes del mutex en el repositorio en memoria, los tests de rutas y los tests de contrato en el RED.

## Historia de usuario

Como Agile Enabler, quiero consultar los registros de esfuerzo por historia o por sprint, para ver cuánto se trabajó.

**Épica:** E7 · Registro de esfuerzo

**Prioridad:** MEDIA (la del Issue #45)

## Objetivo

Obtener los registros de esfuerzo de una historia o de un sprint y el total de horas trabajadas, para que el Agile Enabler vea cuánto se trabajó. Es una operación de solo lectura: no crea, modifica ni elimina esfuerzos.

El total exacto es el dato de entrada de **M5 (horas reales)** de `docs/metricas.md` (TEC-03); HU-33 (o HU-45 si se divide) lo usará en el Sprint 4. La API y la interfaz muestran el total redondeado a un decimal, según la convención de TEC-03, pero el dominio conserva la suma exacta.

La historia amplía `internal/effort` con las consultas `ListByItem` y `ListBySprint`, agrega las operaciones de listado a la interfaz `Repository` y las implementa en memoria y PostgreSQL. Suma `GET /api/items/{id}/esfuerzos` y `GET /api/sprints/{id}/esfuerzos`, y un formulario de consulta en la interfaz web. No modifica la tabla `efforts` ni los datos guardados.

## Alcance

- **Incluye:** consultar por historia o por sprint; validar que la entidad exista; devolver todos los registros ordenados por ID ascendente; calcular el total exacto de horas; devolver lista vacía y total 0 cuando la entidad existe pero no tiene registros; devolver copias por valor; listar en sprints `ABIERTO` y `CERRADO`; persistir en memoria y PostgreSQL; exponer los endpoints HTTP; agregar el formulario, la tabla y el total en la interfaz web; tests unitarios, BDD, de contrato, de PostgreSQL, de API, de interfaz y de concurrencia.
- **No incluye:** corregir o eliminar esfuerzos (HU-28), calcular métricas (E9, HU-33), dashboard, reportes, filtros, paginación, búsqueda por integrante, consulta global de todos los esfuerzos, autenticación, entidad de integrante (E4), cambios de esquema ni exportación.

## Contrato

### Entradas

| Consulta | Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|---|
| Por historia | ItemID | entero | Sí | ID del ítem (historia) cuyos esfuerzos se listan |
| Por sprint | SprintID | entero | Sí | ID del sprint cuyos esfuerzos se listan |

Cada consulta usa exactamente uno de los dos IDs; no hay una consulta combinada ni un listado global.

Firmas en `internal/effort.Manager`:

- `ListByItem(itemID int) (Summary, error)`
- `ListBySprint(sprintID int) (Summary, error)`

`Summary`:

```go
type Summary struct {
	Efforts    []Effort
	TotalHours float64
}
```

`Efforts` es una lista nueva, ordenada por ID ascendente, con copias de los esfuerzos. `TotalHours` es la suma exacta de `Hours`, sin redondear.

### Salidas esperadas

- **Si la entidad existe:** un `Summary` con todos los esfuerzos cuyo `ItemID` o `SprintID` coincide con el ID consultado, ordenados por ID ascendente. Cada esfuerzo tiene todos los campos que define HU-26: ID, integrante, fecha, actividad, ítem, sprint y horas.
- **Si la entidad existe y no tiene registros:** un `Summary` con `Efforts` de longitud 0 (no `nil`) y `TotalHours` 0, sin error.
- **Si hay un error:** un `Summary` cero (`Efforts` `nil`, `TotalHours` 0), sin lista, y el error correspondiente.

La fecha de un esfuerzo se devuelve como `time.Time` en el dominio; en PostgreSQL se normaliza a UTC, como en HU-26. La API la presenta con formato `AAAA-MM-DD`.

### Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| No existe un ítem con ese ItemID (incluye 0, negativos y consultar sin ítems creados) | `ErrItemNotFound` de `internal/backlog`, propagado tal cual | No se devuelve ningún `Summary` |
| No existe un sprint con ese SprintID (incluye 0, negativos y consultar sin sprints creados) | `ErrSprintNotFound` de `internal/sprint`, propagado tal cual | No se devuelve ningún `Summary` |
| El repositorio falla al listar | El error que devuelve el repositorio | No se devuelve ningún `Summary` |

No se define ningún error nuevo: la entidad inexistente reutiliza los errores de HU-05 y HU-11, y una lista vacía no es un error (RN-27.3).

**Orden de validación:** primero se valida que el ítem o el sprint exista y después se listan los esfuerzos. Si la entidad no existe, no se consulta el repositorio de esfuerzos.

### HTTP (TEC-04, tema 5 del Planning del Sprint 3)

- `GET /api/items/{id}/esfuerzos` para consultar por historia.
- `GET /api/sprints/{id}/esfuerzos` para consultar por sprint.

- **Respuesta exitosa:** `200 OK` con:

  ```json
  {
    "esfuerzos": [
      {
        "id": 1,
        "sprint_id": 1,
        "item_id": 3,
        "integrante": "García Santiago",
        "fecha": "2026-10-24",
        "actividad": "Implementar el registro de esfuerzo",
        "horas": 2.5
      }
    ],
    "total_horas": 2.5
  }
  ```

  - `esfuerzos` es una lista; si no hay registros se devuelve `[]`, no `null`.
  - `total_horas` es la suma exacta redondeada a un decimal en la API: `math.Round(total*10)/10`. El JSON puede mostrar `0` en lugar de `0.0`; la interfaz lo formatea con un decimal.
  - `horas` de cada esfuerzo se devuelve tal como está guardada, sin redondear.

- **Errores:**

| Caso | Código | Mensaje |
|---|---|---|
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| Ítem inexistente | 404 | El de `ErrItemNotFound` |
| Sprint inexistente | 404 | El de `ErrSprintNotFound` |
| Falla del repositorio | 500 | El de `errInternal`, sin detalle |

No hay códigos 409 ni 413 en estas consultas: no reciben cuerpo ni cambian estado.

### Interfaz web (TEC-04, tema 5)

En `internal/web`, dentro de la sección **Sprints**, debajo del formulario **"Registrar esfuerzo"** de HU-26, se agrega una consulta:

- **Formulario "Consultar esfuerzos":**
  - Selector **Ver por**: `Sprint` (valor `sprint`) o `Historia` (valor `item`), con `Sprint` por defecto.
  - Campo **ID** (`type="number"`, `min="1"`), con `name="referencia"` para no chocar con la propiedad `id` del formulario.
  - Botón **Consultar**.
- **Tabla de resultados:** columnas **ID**, **Fecha**, **Integrante**, **Actividad**, **Ítem**, **Sprint** y **Horas**, con `<tbody id="esfuerzos">`.
- **Total:** `<p id="total-esfuerzos">` muestra `Total: N.N horas`, con un decimal. Si no hay registros, muestra `Total: 0.0 horas`.
- **Mensaje:** `<p id="msg-esfuerzos" role="alert" hidden>` muestra el resultado o el error de la API, con el mismo mecanismo de `accion` y `mostrar` que los demás formularios.
- **Comportamiento:** al enviar, llama a `GET /api/items/{id}/esfuerzos` o `GET /api/sprints/{id}/esfuerzos` según el selector, reemplaza las filas de la tabla y actualiza el total. Si la lista está vacía, muestra una fila `No hay esfuerzos registrados.` con `colSpan="7"`. Las horas de cada fila se muestran tal como llegan; el total se formatea con `toFixed(1)`.

## Reglas de negocio

1. **RN-27.1 Una consulta, un alcance.** Cada consulta usa exactamente un ID: el de una historia o el de un sprint. No se mezclan los dos filtros ni se listan todos los esfuerzos del sistema.
2. **RN-27.2 Listado completo.** Se devuelven todos los esfuerzos de la entidad consultada, sin filtrar por integrante, fecha, actividad ni estado. La consulta de un sprint incluye los esfuerzos de todos sus ítems; la de una historia incluye todos los esfuerzos de ese ítem.
3. **RN-27.3 Lista vacía válida.** Si la entidad existe y no tiene registros, se devuelve una lista de longitud 0 (no `nil`) y total 0, sin error (CA-27.3).
4. **RN-27.4 Entidad existente.** Si el ítem o el sprint no existe, se devuelve el error de HU-05 o HU-11 y no una lista vacía (CA-27.4). Incluye IDs nunca asignados, 0 y negativos.
5. **RN-27.5 Aislamiento.** La consulta por sprint solo devuelve esfuerzos con ese `SprintID`; la consulta por historia solo devuelve esfuerzos con ese `ItemID`. No aparecen registros de otro sprint ni de otra historia (CA-27.5).
6. **RN-27.6 Sprints cerrados consultables.** Se puede consultar un sprint `ABIERTO` o `CERRADO`; el estado del sprint no restringe la lectura (CA-27.6). La consulta es de solo lectura y no cambia el estado de nada.
7. **RN-27.7 Orden por ID.** La lista se ordena por ID ascendente, que es el orden de creación. Ordena el repositorio: el repositorio en memoria recorre en orden de guardado y PostgreSQL usa `ORDER BY id` (CA-27.1).
8. **RN-27.8 Total exacto en el dominio.** `Summary.TotalHours` es la suma de `Hours` de los registros listados, sin redondear. La API y la interfaz aplican la convención de TEC-03 y muestran el total con un decimal; las horas de cada registro no se redondean (CA-27.2). Un total muy chico se ve redondeado: por ejemplo, `0.001` se muestra como `0.0`, mientras el registro conserva `0.001`. La suma exacta es la que alimenta M5 en HU-33; el redondeo es solo de presentación.
9. **RN-27.9 Copias por valor.** La lista devuelta es nueva y cada esfuerzo es una copia. `Effort` no tiene slices ni mapas, así que modificar la lista, un esfuerzo o sus campos no altera lo guardado (CA-27.7).
10. **RN-27.10 Concurrencia.** `MemoryRepository` de `internal/effort` usa un `sync.Mutex` que cubre `Save`, `ListByItem` y `ListBySprint`, con el mismo patrón que el repositorio de defectos de HU-29. El `Manager` y `register.go` de HU-26 no se tocan: `Register` sigue validando y guardando dentro de `WithOpenSprint`, y los listados validan el ítem o el sprint antes de llamar al repositorio. La protección de los datos en memoria queda dentro del repositorio; `PostgresRepository` usa el pool seguro de pgx. Un test con `go test -race` mezcla `Register` y los dos listados (CA-27.8).
11. **RN-27.11 Solo lectura.** Consultar no modifica los esfuerzos, los ítems, los sprints, sus estados ni sus totales. No se guarda un total en el ítem ni en el sprint.
12. **RN-27.12 API e interfaz.** Las dos consultas se exponen en la API y en el formulario "Consultar esfuerzos" de la interfaz web, con un caso exitoso y un test por código de error, como exige el tema 5 del Planning del Sprint 3 (CA-27.10).
13. **RN-27.13 Persistencia sin cambios de esquema.** Se reutiliza la tabla `efforts` de HU-26 y sus índices por `item_id` y `sprint_id`; no se agregan columnas ni tablas.

## Decisión sobre entidades referenciadas

HU-27 referencia **dos entidades por ID y valida la que corresponde a cada consulta**:

1. **La historia**, en la consulta por historia: se usa `backlog.Manager.GetItem`, que devuelve `ErrItemNotFound` si no existe (RN-27.4). No se exige que el ítem esté asignado a un sprint ni que tenga esfuerzos.
2. **El sprint**, en la consulta por sprint: se usa `sprint.Manager.Get`, que devuelve `ErrSprintNotFound` si no existe (RN-27.4). No se exige que el sprint esté abierto.

La validación es necesaria para distinguir dos casos que el usuario ve distinto: "la historia o el sprint no existe" es un error, y "existe pero no tiene esfuerzos" es un resultado válido con lista vacía y total 0. Sin la validación, un ID mal escrito devolvería una lista vacía y ocultaría el error. Es el mismo criterio de HU-06 para el proyecto.

**El proyecto no se valida de forma directa.** En la consulta por historia, el proyecto del ítem no interviene en el filtro; en la consulta por sprint, el proyecto del sprint tampoco. La entidad que el usuario indica ya determina el conjunto. Consultarlo sumaría otra dependencia y otro lock sin agregar una regla nueva.

**El integrante no se valida como entidad** porque E4 no está en `main` ni comprometida en el Sprint 3; se sigue guardando como texto en el esfuerzo (HU-26) y la consulta no lo filtra.

## Restricciones

- **Organización del código.**
  - `internal/effort`: `list.go` (nuevo, `Summary`, `ListByItem` y `ListBySprint`), `repository.go` (interfaz con `Save`, `ListByItem` y `ListBySprint`), `memory_repository.go` (agrega `sync.Mutex` y cubre `Save`, `ListByItem` y `ListBySprint`) y `postgres_repository.go` (implementa los listados). `manager.go` y `register.go` de HU-26 no se modifican.
  - `internal/httpapi`: `efforts.go` (handlers y respuesta de listado), `server.go` (las dos rutas GET). No se cambian los errores de `statusFor`: `ErrItemNotFound` y `ErrSprintNotFound` ya se mapean a 404 y `errInvalidID` a 400.
  - `internal/web`: `index.html` (formulario, tabla y total dentro de Sprints) y `app.js` (llamada a los dos endpoints, llenado de la tabla y total).
  - `README.md`: las dos rutas en la tabla de la API y la consulta en la descripción de la interfaz; la fila de HU-27 se completa al cerrar la historia.
- **Interfaz del repositorio.** `Repository` se amplía con:
  - `ListByItem(itemID int) ([]Effort, error)`
  - `ListBySprint(sprintID int) ([]Effort, error)`

  Ambas devuelven una lista nueva, ordenada por ID ascendente, con copias de los esfuerzos, y una lista vacía no `nil` cuando no hay registros. No validan la existencia de la entidad: eso lo hace el `Manager`.
- **Consecuencia de ampliar la interfaz.** Hay que actualizar los repositorios falsos de los tests de HU-26 que implementan `Repository`, en particular `failingEffortRepository` de `internal/effort/register_test.go` y `brokenEffortRepository` de `internal/httpapi/efforts_test.go`, para que sigan cumpliendo la interfaz. `recordingEffortRepository` embebe `MemoryRepository`, así que hereda los listados. No se cambia el comportamiento de `Save` ni los tests existentes de `Register`.
- **Locks y orden.** La concurrencia se resuelve dentro de `MemoryRepository`:
  - `MemoryRepository` agrega `mu sync.Mutex` y lo toma al inicio de `Save`, `ListByItem` y `ListBySprint`, con `defer Unlock`, como el repositorio de defectos de HU-29.
  - `Manager` y `register.go` no cambian. `Register` conserva `WithOpenSprint` (mutex del sprint) y la validación del ítem; el `Save` del repositorio queda protegido por su propio mutex.
  - `ListByItem` valida el ítem con `backlog` y después llama a `repo.ListByItem`; `ListBySprint` valida el sprint con `sprint` y después llama a `repo.ListBySprint`. El orden de locks es **sprint → backlog → repositorio de esfuerzos**; nunca se llama a `sprint` ni a `backlog` con el mutex del repositorio tomado.
  - `PostgresRepository` no agrega mutex: el `pgxpool` es seguro para uso concurrente.
  - Un test con `go test -race` ejecuta `Register` y los dos listados a la vez (CA-27.8).
- **Persistencia.** No se modifica `db/init/01-schema.sql`; la tabla `efforts` y los índices `efforts_item_id_idx` y `efforts_sprint_id_idx` ya existen. PostgreSQL lista con `WHERE item_id = $1 ORDER BY id` y `WHERE sprint_id = $1 ORDER BY id`.
- **HTTP.** Las respuestas usan `writeJSON` y los errores `writeError`; no se agrega un error nuevo. La lista `esfuerzos` se inicializa con `make([]effortResponse, 0, len(...))` para que un resultado vacío sea `[]` y no `null`. El total se redondea a un decimal en el handler, no en el dominio.
- **Interfaz web.** La consulta va dentro de Sprints, debajo del formulario de registro, para mantener E7 en una sola pantalla. Usa `textContent` para todos los datos y muestra los errores de la API con `accion`. No agrega librerías ni recursos externos.
- **Tests.** Unitarios en `internal/effort/list_test.go`; pasos de godog en `internal/effort/list_features_test.go`, con el runner `TestFeatures_HU27ConsultarEsfuerzos`, sobre `features/HU-27-consultar-esfuerzos.feature`, con el CA en el nombre de cada escenario (tema 7 del Planning). El contrato de los listados se agrega a `internal/effort/repository_contract_test.go` y corre en memoria y PostgreSQL. Los tests específicos de PostgreSQL van en `internal/effort/postgres_repository_test.go`; requieren `TEST_DATABASE_URL` y la base levantada, y se omiten sin la variable (CA-TEC05.7). La API se prueba en `internal/httpapi/efforts_test.go` (o en `effort_lists_test.go`) y la interfaz en `internal/httpapi/web_test.go`. Las dos rutas nuevas se suman a `TestRutasConID_IDNoNumericoDevuelve400` en `internal/httpapi/server_test.go`.
- **Separación RED/GREEN.** El commit RED incluye la feature BDD, los pasos de godog, los tests unitarios, los tests de contrato de `ListByItem` y `ListBySprint` (memoria y PostgreSQL), los tests de PostgreSQL, los tests de API, los tests de interfaz y el test de concurrencia, junto con stubs vacíos de los métodos nuevos que permitan compilar y fallen por aserción. El wiring de las dos rutas en `internal/httpapi/server.go` y la implementación real de los listados, handlers y formulario van en el commit GREEN. Los tests de contrato de PostgreSQL se omiten sin `TEST_DATABASE_URL` tanto en RED como en GREEN.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Definition of Done.** La SDD lleva la línea "Aprobada por"; antes del commit RED se corre `go vet ./...`; la concurrencia se verifica con `go test -race` en el CI.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Sprint con varios esfuerzos | Se listan todos, ordenados por ID ascendente, con el total exacto |
| Historia con varios esfuerzos | Se listan todos los de esa historia, ordenados por ID ascendente, con el total exacto |
| Sprint existente sin esfuerzos | Lista vacía (longitud 0, no `nil`) y total 0, sin error |
| Historia existente sin esfuerzos | Lista vacía (longitud 0, no `nil`) y total 0, sin error |
| Esfuerzos de otro sprint | No aparecen en la consulta por sprint |
| Esfuerzos de otra historia | No aparecen en la consulta por historia |
| Sprint `CERRADO` con esfuerzos | Se listan igual; el estado no restringe la consulta |
| Ítem inexistente, 0 o negativo | `ErrItemNotFound`; no se devuelve lista |
| Sprint inexistente, 0 o negativo | `ErrSprintNotFound`; no se devuelve lista |
| Total con decimales (`2.5`) | El dominio devuelve `2.5`; la API y la interfaz muestran `2.5` |
| Total `0.001` | El dominio devuelve `0.001`; la API devuelve `0` y la interfaz muestra `0.0`; el registro conserva `0.001` |
| Total `0.05` | La API redondea a `0.1` y la interfaz muestra `0.1`; los registros conservan sus horas |
| Modificar la lista o un esfuerzo devuelto | Los esfuerzos guardados no cambian; volver a listar devuelve los datos originales |
| Listados simultáneos con `Register` | `go test -race` no reporta condiciones de carrera y no se pierden registros |
| El repositorio falla al listar | Se devuelve ese error y ningún `Summary` |
| HTTP: `{id}` no numérico (`abc`) | 400 con el mensaje de `errInvalidID` |
| HTTP: ítem o sprint inexistente | 404 con el mensaje de `ErrItemNotFound` o `ErrSprintNotFound` |
| HTTP: entidad sin registros | 200 con `{"esfuerzos":[],"total_horas":0}` |
| Interfaz: consulta exitosa con registros | Tabla completa, total con un decimal y mensaje de éxito |
| Interfaz: consulta sin registros | Fila "No hay esfuerzos registrados.", total `0.0 horas` y mensaje |
| Interfaz: error de la API | El mensaje de error se muestra en `msg-esfuerzos` |

## Criterios de aceptación

- **CA-27.1** Se listan todos los esfuerzos de una historia o de un sprint, con ID, integrante, fecha, actividad, ítem, sprint y horas, ordenados por ID ascendente.
- **CA-27.2** Se muestra el total de horas: el dominio devuelve la suma exacta y la API y la interfaz la muestran con un decimal, según TEC-03. Las horas de cada registro se devuelven y se muestran sin redondear.
- **CA-27.3** Una historia o un sprint existente sin registros devuelve una lista vacía (longitud 0, no `nil`) y total 0, sin error.
- **CA-27.4** Consultar una historia o un sprint inexistente (incluidos 0 y negativos) devuelve `ErrItemNotFound` o `ErrSprintNotFound`; la API responde 404 y no se devuelve lista.
- **CA-27.5** La consulta por sprint solo incluye esfuerzos de ese sprint y la consulta por historia solo los de esa historia; no se mezclan registros.
- **CA-27.6** Se pueden consultar sprints `ABIERTO` y `CERRADO`; el estado del sprint no restringe la lectura.
- **CA-27.7** Modificar la lista o un esfuerzo devuelto no altera los esfuerzos guardados; volver a listar devuelve los datos originales.
- **CA-27.8** Los listados simultáneos con `Register` no producen condiciones de carrera ni pierden registros, verificado con `go test -race`.
- **CA-27.9** Si el repositorio falla al listar, se devuelve ese error y ningún `Summary`.
- **CA-27.10** Las dos consultas se pueden usar desde la API (`GET /api/items/{id}/esfuerzos` y `GET /api/sprints/{id}/esfuerzos`) y desde el formulario "Consultar esfuerzos" de la interfaz web. La API responde 200 en el caso exitoso y 400, 404 o 500 según la tabla HTTP; la interfaz muestra la tabla, el total con un decimal, el caso vacío y el mensaje de error. Se cubre con tests de `internal/httpapi`, no con escenarios godog.

## BDD y tests previstos

**Archivo BDD:** `features/HU-27-consultar-esfuerzos.feature`, con el runner `TestFeatures_HU27ConsultarEsfuerzos` en `internal/effort/list_features_test.go`. Cada escenario lleva su CA en el nombre:

- Listar los esfuerzos de un sprint con su total (CA-27.1 y CA-27.2)
- Listar los esfuerzos de una historia con su total (CA-27.1 y CA-27.2)
- Sprint existente sin esfuerzos (CA-27.3)
- Historia existente sin esfuerzos (CA-27.3)
- Aislar los esfuerzos por sprint (CA-27.5)
- Aislar los esfuerzos por historia (CA-27.5)
- Consultar un sprint cerrado (CA-27.6)
- Rechazar un sprint y una historia inexistentes (CA-27.4)
- Modificar lo devuelto no altera lo guardado (CA-27.7)
- Error del repositorio al listar (CA-27.9)
- Listados simultáneos con registros (CA-27.8)

**Tests unitarios:** `internal/effort/list_test.go` cubre CA-27.1 a CA-27.9 y los casos límite, incluidos el total exacto, el redondeo de presentación y los IDs 0/negativos.

**Tests de contrato:** `internal/effort/repository_contract_test.go` verifica `ListByItem` y `ListBySprint` en `MemoryRepository` y `PostgresRepository`: lista vacía no `nil`, orden por ID, aislamiento, copias y listado de varios registros. El fixture de PostgreSQL crea al menos dos ítems y dos sprints para probar el aislamiento.

**Tests de PostgreSQL:** `internal/effort/postgres_repository_test.go` verifica los listados por ítem y por sprint, el orden y el caso vacío. Se omiten sin `TEST_DATABASE_URL`.

**Tests de API:** `internal/httpapi/efforts_test.go` (o `effort_lists_test.go`) con un caso exitoso por cada ruta, lista vacía, 400 por ID no numérico en las dos rutas nuevas dentro de `TestRutasConID_IDNoNumericoDevuelve400`, 404 por ítem y sprint inexistentes, 500 por repositorio roto y el redondeo de `total_horas`. Se actualizan los repositorios falsos para la interfaz ampliada.

**Tests de interfaz:** `internal/httpapi/web_test.go` verifica el formulario, el selector Sprint/Historia, la tabla, el total, el caso vacío y las llamadas de `app.js` a los dos endpoints. El recorrido en el navegador queda en la evidencia TDD.

**Concurrencia:** un test de `internal/effort` ejecuta `Register` y los dos listados a la vez bajo `go test -race`.

## Fuera de alcance

- Corregir o eliminar esfuerzos (HU-28).
- Calcular M5, M6, M7 o cualquier métrica (HU-33 y E9).
- Consultar esfuerzos por proyecto, por integrante o de todo el sistema.
- Filtros por fecha, actividad o estado, paginación, búsqueda y orden configurable.
- Dashboard, reportes y exportación.
- Registrar integrantes como entidad (E4).
- Modificar el esquema de la base o los datos guardados.
- Autenticación y permisos.
- Editar el Issue #45 o el tablero en esta etapa: el texto de los criterios queda propuesto para aplicarlo cuando la SDD esté aprobada y con autorización.

## Comparación con el acta de Planning

El acta del Sprint 3 (`docs/actas/sprint-3/planning.md`, Issue #110) incluye a HU-27 con **3 Story Points**, responsable **García Santiago**, dependiente de **HU-26** y sin historias bloqueadas. Su justificación es "Listado por historia o por sprint, como HU-06". El **tema 5** exige que cada historia nueva sume su endpoint bajo `/api/...` y su pantalla en la interfaz web, con un caso exitoso y un test por código de error; el **tema 7** pide el CA en el nombre de cada escenario BDD; la **Definition of Done** pide la línea "Aprobada por" en cada SDD y `go vet ./...` antes del commit RED.

**El alcance coincide:** es un listado de solo lectura por historia o por sprint, con las mismas decisiones de HU-06 para la entidad inexistente y la lista vacía, y suma la API y la interfaz del tema 5. No hay diferencias con el acta. La dependencia HU-26 ya está integrada en `main`; la estimación de 3 Story Points se mantiene.

## Comparación con los criterios del Issue #45 (borrador)

| CA en el Issue #45 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-27.1 Se listan los registros de una historia o de un sprint | CA-27.1 (precisado) | Se agregan todos los campos y el orden por ID ascendente |
| CA-27.2 Se muestra el total de horas | CA-27.2 (precisado) | Suma exacta en el dominio; un decimal en API e interfaz; horas de cada registro sin redondear (TEC-03) |
| CA-27.3 Sin registros devuelve una lista vacía y total cero | CA-27.3 (precisado) | Lista de longitud 0, no `nil`, y total 0 |
| - | CA-27.4 (nuevo) | Validación de la existencia del ítem y del sprint, con 404 |
| - | CA-27.5 (nuevo) | Aislamiento entre historias y sprints |
| - | CA-27.6 (nuevo) | Sprints abiertos y cerrados consultables |
| - | CA-27.7 (nuevo) | Copias por valor |
| - | CA-27.8 (nuevo) | Concurrencia verificada con `-race` |
| - | CA-27.9 (nuevo) | Error del repositorio propagado |
| - | CA-27.10 (nuevo) | Endpoints y formulario web (tema 5 del Planning) |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #45** (en lenguaje de negocio; se aplica cuando la SDD esté aprobada y con autorización, no en esta etapa):

```markdown
## Criterios de aceptación

- [ ] CA-27.1 Se listan todos los esfuerzos de una historia o de un sprint, con integrante, fecha, actividad, ítem, sprint y horas, ordenados por ID.
- [ ] CA-27.2 Se muestra el total de horas con un decimal; las horas de cada registro se muestran sin redondear.
- [ ] CA-27.3 Una historia o un sprint existente sin registros devuelve una lista vacía y total cero.
- [ ] CA-27.4 Consultar una historia o un sprint inexistente devuelve un error y no una lista vacía.
- [ ] CA-27.5 La consulta por sprint no mezcla esfuerzos de otro sprint y la consulta por historia no mezcla los de otra historia.
- [ ] CA-27.6 Se pueden consultar sprints abiertos y cerrados.
- [ ] CA-27.7 Modificar lo devuelto por la consulta no altera los esfuerzos guardados.
- [ ] CA-27.8 Consultar mientras se registran esfuerzos no produce condiciones de carrera ni pierde registros.
- [ ] CA-27.9 Un error del repositorio al listar se informa y no se devuelve una lista.
- [ ] CA-27.10 La consulta se puede usar desde la API y desde la interfaz web, con tabla, total, caso vacío y mensajes de error.
```

## Decisiones confirmadas para esta SDD

1. **Dos consultas separadas:** por historia (`GET /api/items/{id}/esfuerzos`) y por sprint (`GET /api/sprints/{id}/esfuerzos`), sin consulta combinada ni listado global.
2. **Entidad inexistente:** 404, como HU-06; la lista vacía y el total 0 quedan solo para una entidad existente sin registros.
3. **Sprints cerrados:** se pueden consultar; la consulta no exige que el sprint esté abierto.
4. **Orden:** por ID ascendente, igual que HU-06.
5. **Total:** el dominio devuelve la suma exacta; la API y la interfaz la muestran con un decimal, según `docs/metricas.md` (TEC-03). Las horas de cada registro no se redondean.
6. **M5:** la suma exacta es la que alimenta M5 en HU-33; el redondeo de la API es solo de presentación. Un total de `0.001` se ve como `0.0`, pero el registro conserva su valor.
7. **Interfaz:** la consulta va dentro de Sprints, debajo del formulario de registro, con selector Sprint/Historia, tabla y total.
8. **Persistencia:** se reutiliza la tabla `efforts`; no hay cambios de esquema.
9. **Concurrencia:** `MemoryRepository` usa un `sync.Mutex` que cubre `Save`, `ListByItem` y `ListBySprint`, como el repositorio de defectos de HU-29. `Manager` y `register.go` no se tocan. El test con `go test -race` mezcla `Register` con los dos listados.
10. **Interfaz del repositorio:** se agregan `ListByItem` y `ListBySprint`, y se actualizan los repositorios falsos de los tests.
11. **Sin errores nuevos:** se propagan `ErrItemNotFound` y `ErrSprintNotFound`, y `statusFor` ya los mapea a 404.
12. **Aprobación:** Peñalbé Hernán (Agile Enabler) aprobó la SDD el 25/10/2026, con los ajustes del mutex en el repositorio en memoria, los tests de rutas y los tests de contrato en el RED.
