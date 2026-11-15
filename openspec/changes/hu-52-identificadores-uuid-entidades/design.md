# Design: HU-52 · Identificadores UUID del resto de las entidades

## Context

Motivación y alcance: ver `proposal.md`. Lo que existe hoy y condiciona el diseño:

- **HU-51 dejó el patrón** (`internal/httpapi/server.go` y `internal/project`):
  - `parseProjectID` acepta solo la forma canónica (36 caracteres, guiones en 8, 13, 18 y 23, y después `uuid.Parse`);
  - `projectIDFromPath` traduce el UUID al ID interno, y un UUID inexistente al 0;
  - `publicProjectID` busca el UUID para las respuestas;
  - `errInvalidProjectID` responde 400 en `statusFor`;
  - en el dominio, `Project.PublicID`, `uuid.NewV4()` en `Create` y `GetByPublicID` en el `Manager` y en los dos repositorios.
- **Los 6 dominios guardan IDs enteros:**
  - `backlog.Item` tiene `ID`, `ProjectID` y `SprintID` (0 es "sin asignar");
  - `sprint.Sprint` tiene `ID` y `ProjectID`;
  - `member.Member` tiene `ID` y `ProjectID`;
  - `defect.Defect` tiene `ID`, `ItemID`, `DetectionSprintID` y `ResolutionSprintID` (0 es "sin resolver");
  - `effort.Effort` tiene `ID`, `ItemID` y `SprintID`;
  - `poker.Session` tiene `ID` e `ItemID`.
- **Dónde se crea cada entidad:** `backlog/create.go`, `sprint/create.go`, `member/register.go`, `defect/register.go`, `effort/register.go` y `poker/create.go`. Repetir la ronda (HU-24) hace `Update` de la misma sesión.
- **El 0 tiene significado en el dominio:** `defect.Register` trata `detectionSprintID <= 0` como "sprint de detección faltante" (400), y `ChangeStatus` hace lo mismo con el sprint de resolución. `backlog/assign.go` y `report/pdf.go` usan `SprintID != 0` y `ResolutionSprintID != 0` como "tiene sprint". Por eso el truco del 0 de HU-51 (D4) no sirve tal cual (ver D4).
- **Las respuestas se arman en funciones que se llaman desde archivos de otra entidad:**
  - `toSprintResponse` (`sprints.go`) se usa en `dashboard.go`, `metrics.go` y `sprint_reports.go`;
  - `toItemResponse` (`items.go`), en `reports.go`;
  - `toEffortListResponse` (`efforts.go`) y `toDefectListResponse` (`defects.go`), en `reports.go` y `sprint_reports.go`;
  - `toSprintMetricsResponse` (`metrics.go`), en `sprint_reports.go`.

  Si cada responsable cambiara la firma de su función, los archivos de los demás no compilarían.
- **Plazo:** el corte es el lunes 16/11 a las 14:00, con tres personas trabajando en paralelo sobre las entidades.

## Goals / Non-Goals

**Goals:**

- Que desde afuera (API, PDF e interfaz) ninguna de las 6 entidades se identifique con un entero, igual que los proyectos en HU-51.
- Que las tres entidades en paralelo no toquen los mismos archivos: todo lo compartido lo resuelve la base, antes.
- Que cada entidad se pueda integrar sola, y que lo que no llegue al corte quede consistente y documentado.

**Non-Goals:**

- Cambiar las firmas de los `Manager` o los IDs internos (alternativa B).
- Migrar datos: el esquema se recrea.
- Volver a tocar los proyectos de HU-51.

## Decisions

Las decisiones D1 a D10 fueron aprobadas por el Agile Enabler (Peñalbé Hernán) el 15/11/2026, incluidas D3 (el dominio de las 6 entidades en la base, en pareja) y D4 (el centinela `missingInternalID = math.MaxInt32`).

### D1. El mismo diseño que HU-51: UUID público y entero interno (alternativa B)

Cada entidad conserva su ID entero (clave primaria y claves foráneas) y suma un `PublicID uuid.UUID`, que es lo único que muestran la API y la interfaz. La API traduce el UUID al ID interno.

*Alternativa descartada: A, el UUID en todo el sistema.* Se descartó en HU-51 por el tamaño (62 archivos de producción contra 21, solo para los proyectos). Para 6 entidades sería todavía peor: cambian todas las claves foráneas, las firmas de los `Manager` que reciben IDs y sus tests de dominio, con un día de plazo.

### D2. Una capability por entidad

Se crean `backlog`, `sprints`, `integrantes`, `defectos`, `esfuerzos` y `planning-poker`, cada una con los requisitos de HU-52 de su entidad. No se agregan a `proyectos`.

- `proyectos` describe la gestión de proyectos (E1). Poner ahí los ítems, los defectos o el Planning Poker la convierte en un "todo" sin frontera.
- Cada capability nueva es el lugar natural para migrar después la SDD de su entidad (D9 de HU-51).
- Cada spec queda en su propio archivo, y la revisión de cada entidad mira el suyo.

*Alternativa descartada: una capability transversal "identificadores".* Agrupa por el mecanismo y no por el comportamiento del sistema. Además, cuando se migren las SDD, los requisitos de cada entidad quedarían repartidos entre dos specs.

### D3. La base compartida hace el dominio de las 6 entidades y todo lo compartido

La base (Sosa Ricardo y Cardozo Leandro en pareja, sujeto a que Cardozo Leandro lo confirme) se integra primero, con el horario aprobado del domingo 15/11/2026 a las 21:00, y deja listo todo lo que tocarían varias entidades a la vez:

| Pieza de la base | Archivos |
|---|---|
| Esquema: `public_id UUID NOT NULL UNIQUE` en las 6 tablas | `db/init/01-schema.sql` |
| Dominio de las 6 entidades: `PublicID`, `uuid.NewV4()` al crear y `GetByPublicID` en el `Manager` y en los dos repositorios, más `public_id` en `Save` y en el `scan` de PostgreSQL | `internal/{backlog,sprint,member,defect,effort,poker}`: el archivo de la entidad, el de creación, `repository.go`, `memory_repository.go`, `postgres_repository.go` y un `get.go` o el archivo del `Manager` |
| Lectura genérica de un UUID: `parsePublicID(s, errFormato)`, que generaliza `parseProjectID` (este sigue existiendo y la llama) | `internal/httpapi/server.go` |
| Los 6 errores de formato (`errInvalidItemID`, `errInvalidSprintID`, `errInvalidMemberID`, `errInvalidDefectID`, `errInvalidEffortID`, `errInvalidSessionID`) y su 400 en `statusFor` | `internal/httpapi/server.go` |
| Traducción de entrada: `itemIDFromPath(r, nombre)`, `sprintIDFromPath`, `memberIDFromPath`, `defectIDFromPath`, `effortIDFromPath`, `sessionIDFromPath` y `sprintIDFromBody(texto)` (D4) | `internal/httpapi/server.go` |
| Traducción de salida: el resolvedor de UUID públicos (D5) | `internal/httpapi/server.go` |
| Las funciones que arman respuestas reciben el resolvedor (D5), sin cambiar todavía su salida, y se actualizan sus llamadas en todos los archivos | `items.go`, `sprints.go`, `efforts.go`, `defects.go`, `metrics.go`, `reports.go`, `sprint_reports.go` y `dashboard.go`. Las respuestas de `poker.go` y `members*.go` se arman en el propio archivo y nadie más las llama, así que quedan para su responsable |
| Tests compartidos: el registro de IDs de `do` para las 6 entidades, las rutas armadas con el `id` que devuelve la API (sin depender del tipo, como en 2.5 de HU-51), las tablas de rutas de `server_test.go` y los tests de dominio que comparan structs completos | `internal/httpapi/*_test.go` y los tests de los 6 paquetes |
| Esqueleto de la evidencia, con una sección por entidad | `docs/evidencias/HU-52-tdd.md` |

Cada entidad toca **solo** lo suyo:

- sus handlers: leer el `{id}` con su `xIDFromPath`, y en defectos los sprints del cuerpo con `sprintIDFromBody`;
- el cuerpo de sus funciones de respuesta: pasar los campos a `string` con el resolvedor;
- sus secciones de `index.html` y `app.js`;
- sus tests nuevos, en archivos nuevos;
- su sección de la evidencia y su `.feature`.

Ninguna entidad toca `server.go`, el esquema, el dominio, `reports.go`, `sprint_reports.go`, `dashboard.go` ni las firmas de las funciones de respuesta.

| Responsable | Entidades | Archivos de producción que toca |
|---|---|---|
| Cardozo Leandro | Ítems e integrantes | `items.go` (handlers y cuerpo de `toItemResponse`), `members.go`, `members_list.go` y `members_update.go`; las secciones de Product Backlog e Integrantes de `index.html` y `app.js` |
| Buttini Cristobal | Sprints y Planning Poker | `sprints.go` (handlers, `sprintItemIDs` y cuerpo de `toSprintResponse`), `poker.go`, `metrics.go` (`sprintMetrics` y cuerpo de `toSprintMetricsResponse`) y `sprint_reports.go` (solo el handler `sprintReport`); «Sprint actual» y las secciones de Sprints y Planning Poker de `index.html` y `app.js` |
| García Santiago | Defectos y esfuerzos | `defects.go` y `efforts.go` (handlers y cuerpos de sus funciones de respuesta); las secciones de Esfuerzos y Defectos de `index.html` y `app.js`, incluidos los 2 `Number(...)` de los sprints |
| Peñalbé Hernán | PDF del reporte (HU-40) | `internal/report/pdf.go` y lo que el reporte necesite para traer los UUID del ítem y del sprint, con sus tests de `internal/report` |

**Rutas por responsable.** Cada responsable convierte y prueba las rutas cuyos handlers están en sus archivos, cualquiera sea la entidad del `{id}`. Las 23 rutas que reciben un `{id}` o un `{itemId}` de estas entidades (sin contar los `GET` que responden 405) quedan así:

| Responsable | Rutas (handler y archivo) | Cuántas |
|---|---|---|
| Cardozo Leandro | `GET` y `PUT /api/items/{id}`, `PATCH …/estado` y `…/story-points` (`items.go`); `PUT /api/integrantes/{id}` (`members_update.go`) | 5 |
| Buttini Cristobal | `POST /api/items/{id}/poker-sessions` y las 4 de `/api/poker-sessions/{id}/…` (`poker.go`); `POST /api/sprints/{id}/items/{itemId}` y `…/completar`, `GET …/completados` y `POST …/cerrar` (`sprints.go`); `GET /api/sprints/{id}/metricas` (`metrics.go`); `GET /api/sprints/{id}/reporte` (`sprint_reports.go`) | 11 |
| García Santiago | `POST /api/items/{id}/defectos`, `PATCH /api/defectos/{id}/estado` y `GET /api/sprints/{id}/defectos` (`defects.go`); `GET /api/items/{id}/esfuerzos`, `POST /api/sprints/{id}/items/{itemId}/esfuerzos`, `GET /api/sprints/{id}/esfuerzos` y `PATCH /api/esfuerzos/{id}/horas` (`efforts.go`) | 7 |

Así, "las rutas de ítems aceptan el UUID" (spec `backlog`) se cumple recién cuando las tres partes están integradas. Cada `.feature` y cada test de entidad prueba solo las rutas de su responsable.

**Decidido por el Agile Enabler el 15/11/2026: el PDF sale de la base y lo toma Peñalbé Hernán** (HU-40 es suya), en paralelo con las entidades. Pasar a UUID los IDs del texto del PDF cambia una salida, y la base no cambia rutas ni salidas. Además, los tests actuales de `internal/report` verifican el texto con enteros ("#1 … sprint 1"), así que hacerlo en el GREEN de la base obligaba a tocar tests. Depende solo de la base (los `PublicID` del dominio), no de ninguna entidad, así que puede ir en paralelo. `pdf.go` solo tiene los IDs internos: para imprimir los UUID del ítem y del sprint, el reporte tiene que traerlos, y eso lo decide Peñalbé Hernán en su parte. **CA-52.3 no se cumple del todo hasta que se integre el PDF:** mientras tanto, el PDF sigue mostrando enteros.

*Alternativa descartada: que cada responsable haga el dominio de su entidad.* Defectos, esfuerzos y Planning Poker muestran el ítem, y defectos, esfuerzos, métricas e ítems muestran el sprint. Sin el dominio de ítems y sprints, García y Buttini no podrían terminar, así que el trabajo quedaría en serie (ítems y sprints primero), con un día de plazo. El dominio de las 6 entidades sigue el mismo patrón mecánico que el de proyectos en HU-51, y en un solo lugar se hace más rápido y sin choques.

```
orden de integración

  base (Sosa) --merge--> main
                          |
        +-----------------+-----------------+
        |                 |                 |
  items+integrantes  sprints+poker    defectos+esfuerzos
   (Cardozo)          (Buttini)         (Garcia)
        |                 |                 |
   git merge main    git merge main    git merge main
        +----> PR ----+----> PR ---------+----> PR   (cualquier orden)
```

### D4. Un UUID inexistente se traduce a un ID interno que nunca se asigna, y no al 0

`xIDFromPath` sigue el patrón de `projectIDFromPath`:

- un UUID con mal formato da el error de formato de su entidad (400);
- un UUID que existe da su ID interno;
- un UUID que no existe da `missingInternalID = math.MaxInt32`, y cada `Manager` informa la inexistencia (404) en el mismo punto de su orden de validaciones que hoy.

`sprintIDFromBody(texto)` hace lo mismo para los campos del cuerpo, y además un texto vacío da 0, que sigue siendo "sprint faltante" (400 de HU-29 y HU-30).

**Una constante con nombre.** El centinela es `const missingInternalID = math.MaxInt32`, en `internal/httpapi/server.go`, con un comentario que dice qué es (el ID interno de un UUID que no existe), por qué no es el 0 y por qué no es un valor más grande. Ningún handler escribe `math.MaxInt32` directamente: todos usan la constante, a través de los `xIDFromPath` y de `sprintIDFromBody`.

**Por qué no el 0:** en defectos, el 0 (y cualquier negativo) es "sprint faltante" y responde 400. Hoy, un sprint positivo inexistente responde 404 (SDD de HU-29), y eso se conserva solo con un centinela positivo que nunca existe. El máximo de `INTEGER` en PostgreSQL es `math.MaxInt32`: un valor más grande haría fallar la consulta con "fuera de rango" (500). Por eso se usa ese y no `math.MaxInt`.

*Alternativa descartada: responder 404 en el handler apenas el UUID no existe.* Cambia el orden de errores documentado en las SDD (por ejemplo, el título vacío de HU-07 o las horas inválidas de HU-28 se informan antes que la inexistencia).

*Alternativa descartada: el 0 para las rutas y el centinela solo para el cuerpo.* Dos convenciones para lo mismo. El centinela sirve en los dos casos.

**Verificación en la base:** los comandos de D4 de HU-51 se adaptan a las 6 entidades. Se verifica que ningún código de producción compare un ID con `math.MaxInt32` y que los repositorios nunca lleguen a asignarlo.

### D5. Un resolvedor de UUID para las respuestas

Las funciones que arman respuestas reciben un `*publicIDs` (en `server.go`), con un método por entidad:

- `item(id)`, `sprint(id)`, `member(id)`, `defect(id)`, `effort(id)` y `session(id)`, que devuelven el UUID como texto;
- `sprint(0)` y `item(0)` devuelven `""` ("sin asignar" y "sin resolver").

El resolvedor busca cada ID una sola vez (cache) y guarda el primer error. El handler responde con `s.writeResolved(w, código, respuesta, ids)`, que da 500 si hubo un error al buscar. Así no cambia la forma de las funciones de respuesta, que hoy devuelven solo un valor.

- **En la base:** las funciones reciben el resolvedor pero no lo usan, así que la salida sigue igual. Se actualizan todas sus llamadas, incluidas las de `reports.go`, `sprint_reports.go`, `dashboard.go` y `metrics.go`.
- **En cada entidad:** solo cambia el cuerpo de sus funciones. Por ejemplo, `ID: ids.item(it.ID)` y el campo pasa a `string`.

*Alternativa descartada: que cada función devuelva `(respuesta, error)`.* Cambia todas las llamadas, también las de las listas y los reportes. Como en HU-51, cada entidad tendría que tocar archivos de las demás.

*Alternativa descartada: guardar en el dominio los UUID de las entidades relacionadas* (por ejemplo, un `ItemPublicID` en `Defect`). Duplica datos y obliga a mantenerlos sincronizados.

**Cambio en la base, aprobado el 15/11/2026: el resolvedor tiene solo `item(id)` y `sprint(id)`.** Los integrantes, los defectos, los esfuerzos y las sesiones nunca se nombran desde otra entidad: su único ID en una respuesta es su propio `id`, que la función de respuesta toma del struct (`PublicID`). Además, el `Manager` de integrantes no tiene una búsqueda por ID interno, y agregarla solo para el resolvedor sería código sin uso. El resolvedor busca con `items.GetItem` y `sprints.Get`, guarda cada UUID en una cache por ID y el primer error en `err`.

### D6. Generación y validación del UUID

- **Generación:** `uuid.NewV4()` del paquete estándar `uuid` de Go 1.27, en la función que crea cada entidad, dentro del lock que ya toma (CA-52.7). El repositorio guarda el valor que recibe. `Update` no cambia el `public_id`: en PostgreSQL no lo escribe, y en memoria se reescribe la entidad leída, que ya lo trae.
- **Validación:** la forma canónica de D2 de HU-51, con `parsePublicID`. No se exige la versión 4: un UUID de otra versión, o el nulo, simplemente no existe (404).
- **Unicidad:** `UNIQUE` por tabla en PostgreSQL. En memoria no se controla, porque la probabilidad de un choque es despreciable (como en D3 de HU-51).

*Alternativa descartada: `DEFAULT gen_random_uuid()` en PostgreSQL.* La memoria necesitaría otro generador, y la regla quedaría en dos lugares (igual que en HU-51).

### D7. `""` para "sin sprint" en las respuestas

`sprint_id` de un ítem sin asignar y `sprint_resolucion_id` de un defecto sin resolver pasan de `0` a `""`, igual que las fechas sin registrar de HU-18 y HU-41.

*Alternativa descartada: `null`.* Mezcla tipos en un campo que en el resto de la API es siempre texto. Además, la interfaz y los reportes ya tratan `""` como "sin valor".

### D8. Datos existentes

El esquema se recrea con `docker compose down -v` y `docker compose up -d --wait`, como en D6 de HU-51. No hay `ALTER TABLE` ni migración, y la CI y Azure no necesitan nada.

### D9. BDD, tests y evidencia

- **`.feature`:** `features/HU-52-base.feature` (de dominio: los 6 `Create`, `Register` u `Open` asignan un UUID v4, son distintos y se buscan por UUID) y uno por entidad: `HU-52-items.feature`, `HU-52-integrantes.feature`, `HU-52-sprints.feature`, `HU-52-planning-poker.feature`, `HU-52-defectos.feature` y `HU-52-esfuerzos.feature`, a nivel de la API. Cada uno tiene su runner en un archivo nuevo de `internal/httpapi` (o en el paquete de dominio, para el de la base).
- **Tests nuevos, en archivos nuevos:**
  - de la base: `public_id_test.go` en cada uno de los 6 paquetes y `public_ids_internal_test.go` en `httpapi`;
  - de cada entidad, en `internal/httpapi`: `items_uuid_test.go`, `members_uuid_test.go`, `sprints_uuid_test.go`, `poker_uuid_test.go`, `defects_uuid_test.go` y `efforts_uuid_test.go`, más uno `_web_test.go` para la interfaz.
- **Comentario:** cada test lleva `// HU-52 · CA-52.N · Verifica que …`, y los runners `// HU-52 · BDD · Verifica que …`.
- **Tests existentes que cada entidad adapta:** solo los de sus archivos, o las filas de su entidad en las tablas que la base dejó separadas. `TestRutasConID_IDNoNumericoDevuelve400` queda con un bloque de mensajes por entidad, separados por líneas en blanco, para que los merges no choquen. Las 4 aserciones de `web_test.go` que exigen `type="number"` (`sprint_deteccion_id`, `defecto_id`, `sprint_resolucion_id` y `referencia`) son de García.
- **Evidencia:** una sola, `docs/evidencias/HU-52-tdd.md`, con una sección de la base y una por responsable, y una tabla criterio → test → código por sección, con links. La controla `internal/trazabilidad`. La base deja creado el esqueleto, y cada responsable completa solo su sección. Si un merge mueve líneas enlazadas, se corre `go test ./internal/trazabilidad/ -run TestEnlaces -args -update`.

**Cambios en la base, aprobados el 15/11/2026:**

- **El runner de `HU-52-base.feature` está en `internal/httpapi`** (`uuid_base_features_test.go`) y no en un paquete de dominio: es el único paquete que ya usa los `Manager` de las 6 entidades, y ningún paquete de dominio los importa a todos. Los pasos llaman a los `Manager`, no a la API, así que el `.feature` sigue siendo de dominio.
- **Los contratos de repositorio ignoran el `PublicID` en sus comparaciones generales** (`contractAssert*` y los `want` armados a mano), porque el `PublicID` es aleatorio. Lo verifica el contrato nuevo `runRepositoryPublicIDContract` (`public_id_repository_test.go`): guardar y buscar por UUID, un UUID inexistente o el nulo, y que `Update` lo conserve. `contractSave` y las semillas de PostgreSQL asignan un UUID nuevo si el test no lo fija, porque la columna es `UNIQUE`.
- **`do()` traduce las rutas fijas de los tests de `internal/httpapi`** (`entity_ids_test.go`), en lugar de reescribir las 230 rutas literales y las 49 armadas con `%d` de los 15 archivos. `do()` guarda, por handler y por entidad, el `id` que devolvió la API al crear cada ítem, sprint, integrante, defecto, esfuerzo y sesión (`POST` con 201), en el orden de creación, y en la ruta reemplaza el número N de `/api/{entidad}/N/…` (y el `{itemId}` de `/api/sprints/N/items/M/…`) por el `id` de la N-ésima entidad de ese tipo creada por la API. **Hoy es la identidad:** la API todavía devuelve enteros correlativos, así que la N-ésima entidad creada tiene el ID N, y ningún test existente cambia de resultado. Cuando una entidad pasa a responder su UUID, las rutas de los tests siguen funcionando sin tocar los archivos. **Lo que no traduce, y adapta cada responsable para su entidad:** los números que no corresponden a una entidad creada por la API, como los casos de "inexistente" (`/api/items/999`), las entidades guardadas directo en el repositorio (semillas que no pasan por la API), los IDs de los cuerpos (`sprint_deteccion_id`), y las respuestas esperadas y los structs de los tests que hoy decodifican el `id` como entero.

**Desviación de la base (15/11/2026).** Cardozo Leandro la encontró al empezar la sección 2 (ítems e integrantes), y el Agile Enabler (Peñalbé Hernán) decidió el 15/11/2026 resolverla con un ajuste de tests en una rama aparte (`feature/HU-52-ajuste-base`), antes de las entidades. A la base le faltaban tres cosas:

1. **Structs de los tests con el id como `int`.** `itemJSON`, `sprintJSON`, `defectJSON`, `effortJSON`, `pokerSessionJSON`, `sprintMetricsJSON` y una struct anónima de `reports_test.go` leían como entero el `id`, `sprint_id`, `item_id`, `sprint_deteccion_id` o `sprint_resolucion_id`. Se usan en archivos de varios responsables (por ejemplo, `itemJSON` en `metrics_test.go`, `reports_test.go` y `sprints_test.go`), así que el GREEN de una entidad rompía los tests de las demás al decodificar. **Se resolvió** con el tipo `idRef` (`ids_ref_test.go`), que acepta un número o un texto, como `projectRef` en HU-51, y con una adaptación mecánica de sus usos: comparaciones con el texto (`"1"`), `%d` a `%s` con `string(x)` y `strconv.Itoa(x)` a `string(x)`. Donde un test compara IDs correlativos se usa `idInt`, que hoy verifica lo mismo que antes. Los JSON que los tests le mandan a la API no cambian. Las respuestas de integrantes, que se comparan como texto JSON con el id escrito a mano, las adapta Cardozo Leandro en su GREEN.
2. **`do()` traducía todas las rutas.** Con la traducción de D9, apenas una entidad respondía su UUID, `do()` lo ponía también en las rutas de otros responsables cuyos handlers todavía leen un entero (por ejemplo, el `{itemId}` de `POST /api/sprints/{id}/items/{itemId}`, de Buttini Cristobal), y sus tests daban 400. **Se resolvió** con la tabla `uuidRoutes` (`entity_ids_test.go`): `do()` traduce solo las rutas que están en la tabla. Empieza vacía, con un bloque comentado por responsable con sus rutas de la tabla "Rutas por responsable" de D3, y cada responsable activa las suyas en su GREEN. Con la tabla vacía, `do()` no traduce nada, y la API sigue devolviendo enteros.
3. **La tabla de `TestRutasConID_IDNoNumericoDevuelve400` no estaba separada por entidad**, como pedía este mismo D9. **Se resolvió** ordenándola en bloques (proyectos, integrantes, ítems, Planning Poker, sprints, defectos, esfuerzos, y métricas y reportes), separados por líneas en blanco, sin cambiar ningún caso ni mensaje.

Es un ajuste de tests, sin ciclo RED/GREEN: no cambia el comportamiento ni lo que verifica ningún test, y no toca código de producción.

### D10. Cantidades, contadas con comandos sobre `main` `a56b052`

```bash
# Rutas por entidad (handlers que leen el {id}; los methodNotAllowed no cuentan)
grep -nE 'HandleFunc\("[A-Z]+ /api/(items|sprints|integrantes|defectos|esfuerzos|poker-sessions)/' internal/httpapi/server.go
grep -c '{itemId}' internal/httpapi/server.go                                    # 3
# Campos JSON con un ID entero de estas entidades
grep -nE 'int +`json:"(id|[a-z_]*_id)"' internal/httpapi/*.go | grep -v _test     # 16: 2 de pedido y 14 de respuesta
# Tablas y claves foráneas
grep -nE "CREATE TABLE|REFERENCES|DEFAULT 0" db/init/01-schema.sql
# Tests con IDs fijos (por patrón, en internal/httpapi)
git ls-files '*_test.go' | xargs grep -hoE '/api/items/[0-9]|/items/[0-9]+(/|")|"item_id" ?: ?[0-9]' | wc -l   # 94
# (y lo mismo para cada entidad; ver la tabla)
# Campos de la interfaz con un ID numérico
grep -nE '<input[^>]*type="number"' internal/web/index.html    # 12, de los que 9 son IDs (los otros 3 son horas)
grep -nE "Number\(|parseInt" internal/web/app.js               # 2 de sprints (líneas 1095 y 1118)
```

| Entidad | Rutas que la reciben | Campos de respuesta con su entero | Campos de pedido | Tabla y claves foráneas que la apuntan | Tests con IDs fijos (archivos y apariciones) | Campos de la interfaz |
|---|---|---|---|---|---|---|
| Ítem | 7 por `{id}` y 3 por `{itemId}` | 4 (`id`, `item_id` de defecto, esfuerzo y sesión) | — | `backlog_items`; la apuntan `efforts`, `defects` y `poker_sessions` | 12 y 94 | 3 (ítem de esfuerzo, ítem de defecto y "Historia" de la consulta) |
| Sprint | 9 | 6 (`id`, `sprint_id` de ítem, esfuerzo y métricas, y los dos sprints del defecto) | 2 (`sprint_deteccion_id` y `sprint_resolucion_id`) | `sprints`; la apuntan `efforts` y `defects` (detección); `backlog_items.sprint_id` y `defects.resolution_sprint_id` sin clave foránea (`DEFAULT 0`) | 12 y 142 | 3 («Sprint actual» y los dos sprints de defectos), más "Sprint" de la consulta de esfuerzos |
| Integrante | 1 | 1 | — | `members`; nadie la apunta | 3 y 13 | 1 |
| Defecto | 1 | 1 | — | `defects` | 6 y 16 | 1 |
| Esfuerzo | 1 | 1 | — | `efforts` | 1 y 1 | 1 |
| Sesión | 4 | 1 | — | `poker_sessions` | 2 y 48 | 0 (la sesión queda en el formulario) |

Todos los IDs fijos de tests están en 15 archivos de `internal/httpapi`. Los `.feature` existentes son de dominio y no cambian. Los tests de dominio que comparan structs completos se rompen con un `PublicID` aleatorio. Se cuentan con una prueba en la base (asignar el UUID y correr los tests), como en HU-51, porque contar literales con `grep` también cuenta los que solo se usan para guardar.

### Qué implementa cada requisito

| Requisito (en cada capability) | Criterio | Código | Quién |
|---|---|---|---|
| Identificador público UUID al crear | CA-52.1 | `PublicID` y `uuid.NewV4()` en la creación de cada entidad | Base |
| Únicos y no deducibles | CA-52.2 | `uuid.NewV4()`; `UNIQUE` en `01-schema.sql` | Base |
| Creaciones simultáneas | CA-52.7 | La creación dentro del lock de cada `Manager` | Base |
| Operar por UUID | CA-52.3 | `xIDFromPath` y `sprintIDFromBody` (base); su uso en los handlers (entidad) | Base y entidad |
| Ninguna respuesta expone el entero | CA-52.3 | El resolvedor `publicIDs` y la firma de las funciones de respuesta (base); el cuerpo de esas funciones (entidad); `pdf.go` (Peñalbé Hernán) | Base y entidad |
| Formato inválido | CA-52.4 | `parsePublicID`, los 6 errores y `statusFor` | Base |
| Inexistente y orden de errores | CA-52.4 | `missingInternalID` (D4) | Base |
| Memoria y PostgreSQL | CA-52.5 | `Save`, `scan` y `GetByPublicID` de los repositorios | Base |
| Interfaz | CA-52.6 | Las secciones de `index.html` y `app.js` | Entidad |

## Risks / Trade-offs

- **[R1] La base se atrasa y frena a las tres entidades.** → La base es lo primero y no espera a nadie. Las entidades pueden escribir su `.feature` y sus tests en RED sobre la rama de la base apenas esté pusheado su GREEN, sin esperar el merge. Si a las 14:00 del 16/11 la base no está en `main`, HU-52 queda pendiente completa (tabla "Si no llega al corte" de `proposal.md`).
- **[R2] Una entidad no llega al corte y la API queda mezclada.** → Por ejemplo, defectos con `item_id` UUID pero `/api/items/{id}` todavía con entero. Se documenta por entidad en la evidencia y en la Review. El Issue #197 queda abierto con `Refs #197`.
- **[R3] Conflictos de merge en archivos compartidos.** → La base resuelve todos los compartidos antes (D3), y las entidades escriben sus tests en archivos nuevos. Quedan dos riesgos: `app.js` e `index.html` (cada uno toca solo sus secciones, que están en bloques separados) y la tabla de `TestRutasConID_IDNoNumericoDevuelve400` (bloques separados por líneas en blanco). Se resuelve dejando las dos versiones, como en HU-51.
- **[R4] Los links de la evidencia se corren con cada merge.** → `-update` de `internal/trazabilidad` después de cada `git merge main`, como en HU-51.
- **[R5] El centinela `math.MaxInt32` como ID real.** → Una tabla tendría que llegar a 2147483647 filas, y en ese caso `GENERATED ALWAYS AS IDENTITY` ya falla por desborde. Queda documentado.
- **[R6] La firma con el resolvedor sin usar en la base.** → Es un parámetro que todavía no se usa, y queda así hasta que cada entidad lo use. Si una entidad no llega al corte, queda sin usar para esa entidad, y se documenta.
- **[R7] La carga de la base.** → La hacen dos personas en pareja (Sosa y Cardozo, sujeto a que Cardozo lo confirme), con el horario aprobado del domingo 15/11/2026 a las 21:00. La base toca el dominio de 6 paquetes, el esquema, `server.go`, los reportes y los tests compartidos. Es la parte más grande y la que más se repite. Se mide en la tarea B.0 (la prueba de los tests que se rompen) y se avisa si no llega a una hora acordada con el Agile Enabler.
- **[Trade-off] Dos identificadores por entidad,** como en HU-51.
- **[Trade-off] UUID largos en la interfaz y en los mensajes.** → Las entidades se eligen desde listas y botones donde existen. Donde hay que escribir el ID (esfuerzos y defectos), se copia de la tabla.

## Migration Plan

1. **La base:** la integra el equipo con "Create a merge commit", con la CI en verde.
2. **Las entidades:** cada una trae `main` (`git merge main`), corre `-update` de los links, y abre su PR. Se integran en cualquier orden.
3. **Desarrollo local:** `docker compose down -v` y `docker compose up -d --wait` después de integrar la base.
4. **Azure:** usa memoria, así que no hay nada que migrar.
5. **Rollback:** revertir los merges, en orden inverso, y recrear el esquema.
