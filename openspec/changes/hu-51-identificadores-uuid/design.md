# Design: HU-51 · Identificadores UUID (parte de proyectos)

## Context

Motivación y alcance: ver `proposal.md`. Lo que existe hoy y condiciona el diseño:

- **Dominio:** `project.Project.ID int`. `Repository` tiene `Save`, `List`, `GetByID(id int)` y `Update`. `MemoryRepository.Save` asigna `lastID++` y PostgreSQL usa `INTEGER GENERATED ALWAYS AS IDENTITY` (RN-01.5). Todas las operaciones del `Manager` reciben `id int` (`Get`, `Update`, `SetStartDate`, `SetEndDate`, `SetDates`).
- **Otros paquetes:** `backlog`, `sprint`, `member`, `defect`, `metrics`, `dashboard` y `report` reciben `projectID int`. Todos validan que el proyecto exista con `projects.Get(projectID)`, directa o indirectamente (`backlog.ListItems`), y devuelven `project.ErrProjectNotFound` si no existe. `Item`, `Sprint` y `Member` guardan `ProjectID int`.
- **Base:** `backlog_items.project_id`, `sprints.project_id` y `members.project_id` son `INTEGER REFERENCES projects (id)`. El esquema se crea con `db/init/01-schema.sql` solo cuando el volumen está vacío. La CI lo aplica sobre una base nueva. Azure usa memoria y pierde los datos al apagarse (TEC-09).
- **API:** `pathID(r)` llama a `pathInt(r, "id")`, que con `strconv.Atoi` responde `errInvalidID` (400, "el id debe ser un número entero"). `pathID` lo usan por igual las rutas de proyectos, de ítems, de sprints, de defectos, de esfuerzos, de integrantes y de Planning Poker. El orden de errores de cada handler es: ID de la ruta → JSON → formato → `Manager`, y dentro del `Manager` las validaciones de contenido van antes que la existencia (HU-18, línea 72 de su SDD).
- **Salidas con el ID del proyecto:** `projectResponse.ID`, y el campo `proyecto_id` de `itemResponse`, `sprintResponse`, `memberResponse`, `projectMetricsResponse` (y las métricas de sprint) y `summaryResponse`. Además, el `Content-Disposition` del PDF usa `reporte-proyecto-%d.pdf` (`report_pdf.go`).
- **Interfaz:** `#proyecto-actual` es `type="number" min="1"`. `app.js` trata el ID como texto en todos lados (`idActual` devuelve el valor del campo sin convertirlo, y `p.id` se concatena). No hay `Number(...)` sobre el ID del proyecto (verificado con `grep -nE "Number\(|parseInt" internal/web/app.js`).
- **Go 1.27.1** (`go.mod`, CI con `go-version-file` y `Dockerfile` con `GO_VERSION=1.27.1`) trae el paquete estándar **`uuid`**. Lo verifiqué con un programa de prueba fuera del repositorio:
  - `uuid.NewV4()` devuelve un UUID con versión 4 y variante RFC 9562;
  - `String()` lo da en minúsculas, con 36 caracteres;
  - `uuid.Parse` acepta mayúsculas, llaves, `urn:uuid:` y la forma sin guiones;
  - `uuid.UUID` es `[16]byte`, un valor comparable con `==`.

## Goals / Non-Goals

**Goals:**

- Que desde afuera (API e interfaz) un proyecto solo se identifique por un UUID v4, y que ninguna respuesta deje ver el entero.
- Llegar con la parte obligatoria integrada para el domingo 15/11, sin romper las 43 historias anteriores ni HU-50.
- Conservar las reglas, los errores y su orden en todas las rutas que reciben un proyecto.

**Non-Goals:**

- Los UUID de ítems, sprints, integrantes, defectos, esfuerzos y sesiones de Planning Poker (parte pendiente, ver `proposal.md`).
- Cambiar las firmas de los `Manager` de los demás dominios.
- Una migración de datos.

## Decisions

### D1. Identificador público UUID con ID entero interno (alternativa B)

Se compararon dos alternativas:

- **A.** `Project.ID` pasa a ser `string` con el UUID en todo el sistema: dominio, `Repository`, `project_id` de las tablas que lo referencian, todos los `Manager` y todas las rutas.
- **B.** Se mantiene el ID entero interno (clave primaria y claves foráneas) y se agrega un identificador público UUID. Ese UUID es el único que se expone en la API y en la interfaz, y la API traduce UUID → ID interno.

| | A. UUID en todo el sistema | B. UUID público + entero interno |
|---|---|---|
| Archivos de producción que cambian | **62** (59 Go en 9 paquetes de `internal/`, el esquema, `app.js` e `index.html`) | **21** (6 de `project`, 13 de `httpapi`, el esquema e `index.html`) |
| Archivos de tests existentes que cambian | **81** de 134, más **4** `.feature` (HU-04, HU-06, HU-07 y HU-10, con "proyecto con ID N") | **22** de 134, ningún `.feature` |
| Firmas que cambian | `projectID int` → `string` en `backlog`, `sprint`, `member`, `defect`, `metrics`, `dashboard` y `report`, más sus repositorios en memoria y en PostgreSQL y los repositorios falsos de los tests | Ninguna fuera de `project`. `project.Repository` suma una operación |
| Esquema | `projects.id` y las 3 claves foráneas pasan a `UUID` | Una columna nueva, `projects.public_id UUID NOT NULL UNIQUE` |
| Riesgo | **Alto:** toca 9 paquetes de `internal/` y 6 de cada 10 archivos de tests, en paralelo con TEC-11 (que comenta esos mismos archivos) y con un día de plazo. Un error en una firma se propaga a los paquetes que dependen de ella | **Medio:** el cambio queda en `project` y en `httpapi`. El riesgo propio es olvidarse una salida que filtre el entero (mitigado por D5 y un test por ruta) |
| ¿Cumple "que no se pueda adivinar"? | Sí: no existe ningún entero del proyecto | **Sí, hacia afuera:** el entero existe solo en la base y en la memoria del proceso, y ninguna ruta lo acepta ni lo devuelve. Hay una diferencia con el texto del Issue ("UUID en el dominio y la base"), que se ajusta en los criterios CA-51.1, CA-51.3 y CA-51.5 |
| Parte pendiente (otras entidades) | Mismo trabajo para cada entidad | Mismo patrón para cada entidad, entidad por entidad, sin cambiar firmas |

**Decisión: B.** Cumple el pedido del profesor en todo lo que se ve desde afuera, con un tercio de los archivos de producción y un cuarto de los tests, y deja los dominios E2 a E8 sin tocar. Eso importa con el plazo del 15/11 y con los PR de TEC-11 abiertos sobre esos mismos tests.

*Alternativa descartada: A.* Es la solución "más pura", pero cambia 62 archivos de producción, 81 de tests y 4 `.feature` en un día, y el riesgo de no llegar es alto. Si no se llega, HU-51 queda sin integrar. Con B, la parte obligatoria queda completa, y A sigue siendo posible más adelante.

**Archivos que cambian.** Los conté con estos comandos, desde la raíz del repo, en `main` (`315fac5`):

```bash
# A, producción (62): lo que nombra el ID del proyecto, más internal/project, rutas, esquema e interfaz
{ grep -rlE "projectID|ProjectID|project_id" --include=*.go internal cmd | grep -v _test.go
  grep -lE "id int|ID +int" internal/project/*.go | grep -v _test.go
  ls internal/httpapi/{server,projects,dashboard,defects,reports,report_pdf,sprint_reports}.go
  echo db/init/01-schema.sql; echo internal/web/app.js; echo internal/web/index.html; } | sort -u | wc -l

# A, tests (81): literales enteros de proyecto, struct ProjectID: N, repositorios falsos con id int,
# rutas o JSON con el ID entero (en B también cuentan)
git ls-files '*_test.go' | xargs grep -lE \
 '\.(CreateItem|ListItems|ListItemsByPriority|ListByProject|ProjectSummary|ProjectReport|ProjectMetrics|ProjectSprintsMetrics|ListProjectSprints|History|Register)\([0-9]+[,)]|(projects|projectManager|manager|m|mgr|pm|env\.projects|env\.manager)\.(Get|Update|SetStartDate|SetEndDate|SetDates)\([0-9]+[,)]|sprints?\.Create\([0-9]+,|\.Create\([0-9]+, *"|ProjectID: *[0-9]|GetByID\(id int\) \((project\.)?Project,|ListByProject\(projectID int\)|projectID int|ProjectID +int|/api/proyectos/[0-9]|/api/proyectos/%d|"proyecto_id"|proyecto_id\\":|"/api/proyectos/" *\+|proyecto-actual|\.ProjectID *[!=]= *[0-9]|projectID *:?= *[0-9]' | sort -u | wc -l   # 81

# A, escenarios: .feature con "proyecto con ID N"
grep -lEi "proyecto (con id|de id|id)? *[0-9]+\b|proyecto [0-9]+\b|\"id\": *[0-9]" features/*.feature | wc -l   # 4

# B, producción (21): handlers que reciben un proyecto por la ruta o devuelven proyecto_id/projectResponse
grep -lE "pathID\(r\)" internal/httpapi/*.go | grep -v _test | xargs grep -lE \
 "projectID, err := pathID|ProyectoID|toProjectResponse|toSprintResponse|toItemResponse|projectMetrics|projectSummary|projectReport|listDefectsByProject" | wc -l   # 12
# + server.go + internal/project/{project,repository,memory_repository,postgres_repository,get,create}.go
# + db/init/01-schema.sql + internal/web/index.html

# B, tests (22): rutas o JSON con el ID entero, más los repositorios falsos que implementan project.Repository
git ls-files '*_test.go' | xargs grep -lE '/api/proyectos/[0-9]|/api/proyectos/%d|"proyecto_id"|proyecto_id\\":|"/api/proyectos/" *\+|proyecto-actual' | wc -l   # 20
git ls-files '*_test.go' | xargs grep -lE '\) Save\(p (project\.)?Project\)' | wc -l                                                                 # 2
```

Los conteos de tests son por patrón: pueden incluir un archivo que se arregla con un cambio mínimo, pero no dejan afuera archivos con IDs enteros escritos de esas formas. Los repositorios falsos que **embeben** `project.Repository` (en `dashboard`, `member`, `report` y `sprint`) no se rompen al sumar una operación a la interfaz, y no se cuentan.

### D2. Validación del formato: solo la forma canónica

`internal/httpapi/server.go` agrega `errInvalidProjectID = errors.New("el id del proyecto debe ser un UUID")`, que responde 400 en `statusFor`, y una función `parseProjectID(s string) (uuid.UUID, error)`:

1. `len(s) == 36` y guiones en las posiciones 8, 13, 18 y 23. Si no se cumple, devuelve `errInvalidProjectID`.
2. `uuid.Parse(s)`, que valida los dígitos hexadecimales y acepta mayúsculas. Si falla, devuelve `errInvalidProjectID`.

La salida siempre sale de `UUID.String()`, en minúsculas.

- **No se exige versión 4 al leer.** Un UUID de otra versión, o el nulo, tiene la forma correcta y simplemente no existe: responde 404, como pide la spec. Rechazarlo con 400 no agrega seguridad, porque el sistema nunca lo asigna.
- `errInvalidID` ("el id debe ser un número entero") **no cambia**: lo siguen usando ítems, sprints, integrantes, defectos, esfuerzos y Planning Poker.

*Alternativa descartada: aceptar todo lo que acepta `uuid.Parse`* (llaves, `urn:uuid:`, sin guiones). Daría varias URL para el mismo proyecto, y la API solo devuelve la forma canónica.

*Alternativa descartada: una expresión regular.* `uuid.Parse` ya valida los dígitos; con el control de largo y guiones alcanza, sin `regexp`.

### D3. Generación del UUID v4 en el `Manager`, con la biblioteca estándar

- `Project` suma `PublicID uuid.UUID`. Es un `[16]byte`, así que la copia por valor no comparte estado (regla de copias).
- `Manager.Create`, con el mutex tomado (CA-01.7, CA-51.7), asigna `PublicID: uuid.NewV4()` antes de llamar a `repo.Save`. Los dos repositorios guardan el valor recibido: memoria en el slice y PostgreSQL en `public_id`.
- **Unicidad:** un UUID v4 tiene 122 bits aleatorios del generador criptográfico. En PostgreSQL, la restricción `UNIQUE` hace que un choque imposible en la práctica falle como error de guardado (500) y nunca deje dos proyectos iguales. En memoria no se agrega control de choques: el `Manager` es el único que genera, y la probabilidad es despreciable (aproximadamente 10⁻³⁰ con mil proyectos). Se documenta como decisión.
- `Update`, `SetStartDate`, `SetEndDate` y `SetDates` conservan el `PublicID`, porque leen el proyecto con `GetByID` y lo reescriben completo. `MemoryRepository.Update` reemplaza el proyecto entero. El `UPDATE` de PostgreSQL no incluye `public_id`, así que no se puede cambiar.

*Alternativa descartada: `crypto/rand` armando los bits de versión y variante a mano.* El paquete estándar `uuid` de Go 1.27 ya lo hace y está probado. Escribirlo a mano es código propio sin necesidad.

*Alternativa descartada: `DEFAULT gen_random_uuid()` en PostgreSQL.* La memoria necesitaría otro generador, así que la regla quedaría en dos lugares. Además, `Save` no podría probar con memoria lo mismo que pasa en la base.

*Alternativa descartada: UUID v7.* Lleva la hora de creación en los primeros 48 bits, así que deja deducir cuándo se creó y el orden de creación. El Issue pide v4.

### D4. Traducir UUID → ID interno en la API, conservando el orden de errores

- `project.Repository` suma `GetByPublicID(id uuid.UUID) (Project, error)`, que devuelve una copia o `ErrProjectNotFound`. Memoria busca en el slice y PostgreSQL hace `WHERE public_id = $1`.
- `Manager` suma `GetByPublicID(id uuid.UUID) (Project, error)` en `get.go`, con el mutex, igual que `Get`.
- `httpapi` agrega `s.projectIDFromPath(r) (int, error)`:
  1. `parseProjectID(r.PathValue("id"))`. Si falla, devuelve `errInvalidProjectID` (400).
  2. `s.projects.GetByPublicID(u)`. Si devuelve `ErrProjectNotFound`, la función devuelve **el ID interno 0** sin error. Cualquier otro error se devuelve tal cual (500).
  3. Si el proyecto existe, devuelve su ID interno.

  Las 17 rutas de proyecto cambian `pathID(r)` por `s.projectIDFromPath(r)` y siguen llamando a su `Manager` con un `int`.

**Por qué el 0:** RN-01.5 empieza en 1, y RN-03.2 dice que el ID 0 siempre es `ErrProjectNotFound`. Así, cada `Manager` informa la inexistencia en el mismo punto de su orden de validación que hoy. Por ejemplo, `PATCH …/fecha-inicio` con fecha vacía sobre un UUID inexistente sigue respondiendo 400 (HU-18), y crear un ítem sin título en un proyecto inexistente sigue informando el título primero (HU-04). Ningún test de orden existente cambia.

**Verificación de que ningún código trata el proyecto 0 de forma especial.** Lo verifiqué en `main` (`315fac5`) con estos comandos, desde la raíz del repo:

```bash
# 1. Comparaciones del ID de proyecto con 0 o negativos, en producción
grep -rnE "(projectID|ProjectID|\bid)\s*(==|!=|<=|<|>|>=)\s*0\b|0\s*(==|!=)\s*(projectID|ProjectID)" --include=*.go internal cmd | grep -v _test.go
# 2. Comparaciones de un ID con 0 dentro de internal/project
grep -nE "(\.ID|\bid)\s*(==|!=|<=|<|>|>=)\s*0\b" internal/project/*.go | grep -v _test.go
# 3. Un 0 literal pasado como proyecto, o ProjectID: 0
grep -rnE "\.(Get|ListItems|ListItemsByPriority|CreateItem|ListByProject|Register|Create|History|ProjectSummary|Project|ProjectMetrics|ProjectSprintsMetrics)\(0\b|ProjectID:\s*0\b" --include=*.go internal cmd | grep -v _test.go
# 4. Todo "ID == 0" o "ID != 0" de producción, para ver a qué entidad corresponde
grep -rnE "ID\s*==\s*0|ID\s*!=\s*0" --include=*.go internal | grep -v _test.go
# 5. Valores por defecto en 0 en el esquema
grep -nE "project_id|DEFAULT 0" db/init/01-schema.sql
```

Resultado:

- Los comandos 1, 2 y 3 no devuelven nada.
- El 4 devuelve solo tres líneas, y las tres son de **sprints**, donde el 0 significa "sin asignar" (RN-11.4 y RN-29.4):
  - `internal/backlog/assign.go:28` (`current.SprintID != 0`);
  - `internal/report/pdf.go:140` (`it.SprintID != 0`);
  - `internal/report/pdf.go:150` (`d.ResolutionSprintID != 0`).
- En el 5, las tres columnas `project_id` son `NOT NULL REFERENCES projects (id)`, sin `DEFAULT`. Los únicos `DEFAULT 0` son `backlog_items.sprint_id` y `defects.resolution_sprint_id`.

Por lo tanto, el proyecto 0 solo llega a `Repository.GetByID`, que nunca lo encuentra:

- `MemoryRepository` asigna IDs desde 1;
- PostgreSQL usa `GENERATED ALWAYS AS IDENTITY`, que también empieza en 1;
- en los dos casos se devuelve `ErrProjectNotFound` (RN-03.2).

En GREEN se vuelven a correr estos comandos: si un cambio agrega un trato especial del 0 para proyectos, D4 deja de valer.

*Alternativa descartada: responder 404 en el handler apenas el UUID no existe.* Es más directo, pero cambia el orden documentado en las SDD de HU-04, HU-10, HU-15, HU-18, HU-19 y HU-41, y obliga a reescribir sus tests de combinación de errores.

*Alternativa descartada: que cada `Manager` reciba el UUID.* Es la alternativa A.

### D5. Ninguna salida con el entero del proyecto

- `projectResponse.ID` pasa a `string` (`p.PublicID.String()`).
- `itemResponse`, `sprintResponse`, `memberResponse`, `projectMetricsResponse`, la respuesta de métricas de sprint y `summaryResponse` pasan `ProyectoID` a `string`. `toItemResponse` y `toSprintResponse` reciben además el UUID del proyecto.
- En las rutas `/api/proyectos/{id}/…`, el UUID es el de la ruta, normalizado.
- En las rutas por ID de otra entidad (`/api/items/{id}`, `/api/sprints/{id}/…`, `/api/integrantes/{id}`), el handler lo busca con `s.publicProjectID(internalID) (string, error)`, que usa `projects.Get` y devuelve 500 si falla. En las listas, todos los elementos son del mismo proyecto y se busca una sola vez.
- `report_pdf.go` arma el `Content-Disposition` con el UUID.
- **Control:** un test por cada respuesta que identifica un proyecto verifica que `id` o `proyecto_id` es un texto de 36 caracteres, y que no es un número.

*Alternativa descartada: quitar `proyecto_id` de las respuestas.* Es un cambio incompatible que no pide el Issue, y la interfaz lo muestra en las métricas (`app.js`, línea 744).

### D6. Esquema y datos existentes

- `01-schema.sql` agrega `public_id UUID NOT NULL UNIQUE` en `CREATE TABLE projects`. `UNIQUE` crea el índice que usa `GetByPublicID`. Las claves foráneas y `02-test-database.sql` no cambian, y el `TRUNCATE` de `postgrestest` tampoco.
- **No hay `ALTER TABLE`:** los scripts de `db/init` solo corren con el volumen vacío. Quien tenga una base local la recrea con `docker compose down -v` y `docker compose up -d`, y pierde sus datos de desarrollo. Se documenta en el README y en la evidencia.
- **CI:** aplica el esquema sobre una base nueva, así que no se hace nada. **Azure:** usa memoria (TEC-09), así que no se hace nada.

*Alternativa descartada: una migración con `ALTER TABLE … ADD COLUMN public_id UUID DEFAULT gen_random_uuid()`.* No hay datos productivos que conservar, el proyecto no tiene un sistema de migraciones, y agregarlo es alcance nuevo.

### D7. Interfaz y orden de HU-50

- `index.html`: `#proyecto-actual` pasa a `type="text"`, sin `min`, con `placeholder="UUID del proyecto"`, `autocomplete="off"` y `spellcheck="false"`. El label queda «Proyecto actual (ID)», y el comentario de la sección cita HU-51.
- **`app.js` no cambia:** ya trata el ID como texto (ver Context). Un ID mal escrito lo rechaza la API (400) y `accion` muestra el mensaje.
- **El orden de HU-50 no cambia:** `Manager.List` ordena por nombre (D2 de HU-50) y nunca usó el ID. Los tests de orden de HU-50 siguen como están.

### D8. Tests y escenarios

- `features/HU-51-identificadores-uuid.feature`, en español: escenarios normal, alternativo, límite y de error a **nivel de la API**, porque lo observable de HU-51 son las rutas, los 400/404 y las salidas.
  - Los pasos van en `internal/httpapi/uuid_features_test.go`, con el runner `TestFeatures_HU51IdentificadoresUUID`, sobre `NewHandler` con repositorios en memoria y `httptest`.
  - Es el primer runner godog de `httpapi`. Se elige porque los escenarios de 400/404 no existen en el `Manager`.
- `internal/project/public_id_test.go`:
  - `Create` asigna un v4 canónico (CA-51.1);
  - 1000 proyectos con UUID distintos y no correlativos (CA-51.2);
  - `GetByPublicID` existe, no existe y devuelve una copia (CA-51.4);
  - concurrencia con `-race` (CA-51.7);
  - `Update` y las fechas conservan el `PublicID`.
- `internal/project/repository_contract_test.go` suma `GetByPublicID` y "guardar y buscar por UUID" para memoria y PostgreSQL. `postgres_repository_test.go` suma el rechazo de un `public_id` repetido (CA-51.5).
- `internal/httpapi/projects_uuid_test.go`:
  - `parseProjectID` (formas válidas y rechazadas);
  - 400 y 404 por ruta;
  - orden de errores;
  - ninguna salida con el entero (CA-51.3, CA-51.4).
- `internal/httpapi/projects_uuid_web_test.go`: `#proyecto-actual` es `type="text"` (CA-51.6).
- **Los 22 archivos existentes se adaptan sin perder lo que verifican:** crean el proyecto y usan su UUID en la ruta, y comparan `proyecto_id` con ese UUID. Cada test nuevo lleva `// HU-51 · CA-51.N · Verifica que …`. Los adaptados conservan su comentario de TEC-11.

### Qué implementa cada requisito

| Requisito (spec) | Criterio | Código |
|---|---|---|
| Identificador público UUID al crear | CA-51.1 | `Project.PublicID` (`project.go`); `Manager.Create` (`create.go`); `toProjectResponse` (`httpapi/projects.go`) |
| Identificadores únicos y no deducibles | CA-51.2 | `uuid.NewV4` en `Manager.Create`; `UNIQUE` en `01-schema.sql` |
| Creaciones simultáneas | CA-51.7 | `Manager.Create` (`m.mu`) |
| Operar un proyecto por su UUID | CA-51.3 | `server.projectIDFromPath` (`httpapi/server.go`); `Manager.GetByPublicID` (`get.go`); `Repository.GetByPublicID` (memoria y PostgreSQL); las 17 rutas de proyecto en `projects.go`, `items.go`, `members.go`, `members_list.go`, `sprints.go`, `defects.go`, `metrics.go`, `dashboard.go`, `reports.go` y `report_pdf.go` |
| Ninguna respuesta expone el entero | CA-51.3 | `ProyectoID string` y `server.publicProjectID` en `items.go`, `sprints.go`, `members*.go`, `metrics.go`, `dashboard.go`, `sprint_reports.go` y `reports.go`; `Content-Disposition` en `report_pdf.go` |
| Identificador con formato inválido | CA-51.4 | `parseProjectID` y `errInvalidProjectID` → 400 en `statusFor` (`server.go`) |
| Proyecto inexistente | CA-51.4 | `projectIDFromPath` (0 → `ErrProjectNotFound` del `Manager`) → 404 |
| Memoria y PostgreSQL | CA-51.5 | `MemoryRepository` y `PostgresRepository` (`Save`, `GetByPublicID` y `scanProject`); `01-schema.sql` |
| La interfaz usa el UUID | CA-51.6 | `#proyecto-actual` (`index.html`); `app.js` sin cambios |

## Risks / Trade-offs

- **[R1] Las demás entidades siguen enumerables.** Quien recorra `/api/items/1`, `/api/items/2`… o `/api/sprints/N/…` lee el `proyecto_id` (UUID) de esos ítems y sprints, y con él llega al proyecto. Además, los IDs de ítems y sprints, que son globales, dejan deducir cuántos hay. → Es la parte pendiente de HU-51 y queda escrito en `proposal.md`. El patrón de D4 y D5 se repite entidad por entidad. Esto se avisa en la Review.
- **[R2] Olvidar una salida que filtre el entero.** → D5 enumera todas las salidas, encontradas con `grep -n "ProyectoID:" internal/httpapi/*.go` y `grep -n "%d" internal/httpapi/report_pdf.go`. Cada una tiene un test que verifica el UUID.
- **[R3] El escaneo de `uuid.UUID` con pgx.** pgx v5 escanea `uuid` en `[16]byte` y en tipos con ese tipo base. → El test de contrato de PostgreSQL lo verifica en GREEN. Si fallara, se escanea `public_id::text` y se usa `uuid.Parse`, solo en `scanProject`.
- **[R4] Dos búsquedas por pedido** (`GetByPublicID` en el handler y `Get` en el `Manager`). → Con el volumen del proyecto (decenas de proyectos) no importa, y `public_id` tiene índice por `UNIQUE`.
- **[R5] Choque con los PR de TEC-11**, que comentan los mismos 22 archivos de tests. → Los PR de TEC-11 debían estar integrados el 13/11. Antes de RED se hace `git pull` de `main` y se adaptan los tests sobre la versión ya comentada.
- **[Trade-off] El entero interno sigue existiendo.** No se expone, pero el dominio tiene dos identificadores. → Es el precio de llegar en plazo (D1). La alternativa A queda posible más adelante.
- **[Trade-off] UUID largos en la interfaz**, en los mensajes como "Proyecto 3f2b… elegido" y en la tabla de HU-50. → Se acepta: el usuario elige por nombre (HU-50) y no necesita escribirlo.

## Migration Plan

1. Merge a `main` con "Create a merge commit", después de la aprobación y con la CI en verde.
2. **Desarrollo local con PostgreSQL:** correr `docker compose down -v` y después `docker compose up -d`, para que se recree el esquema con `public_id`. Si se omite este paso, `Save` falla porque la columna no existe.
3. **Azure:** el despliegue habitual. Como usa memoria, no hay datos que migrar. Antes de la demo se cargan los datos con el script de siempre, ahora usando los UUID que devuelve la API.
4. **Rollback:** revertir el merge y recrear el esquema con `down -v`.

## Open Questions

- ¿El script de carga de datos para la demo (TEC-09) arma rutas con IDs fijos como `/api/proyectos/1`? No está en el repositorio. Si los usa, hay que cambiarlo para que lea el `id` que devuelve `POST /api/proyectos`. No cambia la spec ni el diseño.
