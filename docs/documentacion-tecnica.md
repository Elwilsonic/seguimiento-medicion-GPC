# Documentación técnica

> Issue del repositorio: #69

Cómo está construida la aplicación "Software Metrics & Estimation" y cómo se ejecuta y se prueba, para que cualquier integrante o evaluador pueda entender el código, correrlo y probarlo sin ayuda (TEC-06). Describe el código integrado en `main`. Cada sección indica la historia o tarea que la originó; el detalle de cada regla de negocio está en su SDD (`docs/especificaciones/`).

Para el uso de la aplicación desde la interfaz, ver el [README](../README.md#interfaz-web). Para las fórmulas de las métricas, ver [`docs/metricas.md`](metricas.md) (TEC-03).

## 1. Arquitectura

La aplicación es un único binario Go que sirve la API REST bajo `/api/…` y la interfaz web en `/`. El código se organiza en capas:

```
cmd/server            arranque: configuración, repositorios, Managers y servidor HTTP
internal/httpapi      capa HTTP: rutas, JSON y traducción de errores a códigos
internal/<dominio>    reglas de negocio: un Manager por dominio y su interfaz Repository
  ├─ memory_repository.go     implementación en memoria
  └─ postgres_repository.go   implementación en PostgreSQL
internal/web          interfaz web (HTML, JS y CSS embebidos)
internal/postgres     apertura del pool de conexiones a PostgreSQL
internal/postgrestest preparación de la base para las pruebas de integración
```

La capa HTTP solo llama a los Manager; los Manager solo acceden a los datos a través de su interfaz `Repository` (AGENTS.md). Ningún paquete de dominio conoce HTTP ni la base.

### 1.1 Paquetes

| Paquete | Qué hace | Historias |
|---|---|---|
| `cmd/server` | `main`: lee `STORAGE`, `DATABASE_URL` y `PORT` (`storage.go`), arma los repositorios en memoria o PostgreSQL (`repositories.go`), crea los Manager y arranca el servidor con timeouts | TEC-01, TEC-04, TEC-05 |
| `internal/project` | Proyectos: crear, modificar, consultar y registrar la fecha de inicio. Estados `PLANIFICADO`, `EN_CURSO` y `FINALIZADO`; nombre único | HU-01, HU-02, HU-03, HU-18 |
| `internal/backlog` | Ítems del Product Backlog: crear, consultar, listar (por ID o por prioridad), modificar, cambiar el estado (`PENDIENTE` ↔ `EN_CURSO` ↔ `TERMINADO`, de a un paso), asignar story points (escala 1, 2, 3, 5, 8, 13, 21) y asignar a un sprint | HU-04 a HU-09, HU-11, HU-20 |
| `internal/sprint` | Sprints: crear (uno solo `ABIERTO` por proyecto), asignar ítems, completarlos, listar los completados y cerrar (`CERRADO`) | HU-10 a HU-13 |
| `internal/poker` | Sesiones de Planning Poker de un ítem: abrir (`ABIERTA`), votar sin mostrar los votos, revelar (`REVELADA`) y registrar la estimación acordada (`CERRADA`), que se guarda en el ítem | HU-21 a HU-23, HU-25 |
| `internal/effort` | Esfuerzo: registrar integrante, fecha, actividad y horas sobre una historia de un sprint abierto, y listar por historia o por sprint con el total de horas | HU-26, HU-27 |
| `internal/defect` | Defectos: registrar descripción, severidad (`CRITICA`, `ALTA`, `MEDIA`, `BAJA`) y sprint de detección; nacen `ABIERTO` | HU-29 |
| `internal/metrics` | Métricas calculadas a partir del backlog y de los sprints: story points planificados y completados, ítems sin estimar, velocidad y porcentaje de historias completadas. No tiene almacenamiento propio | TEC-03, HU-32, HU-34 |
| `internal/httpapi` | Rutas HTTP (`server.go`), un archivo de handlers por tema, lectura del JSON (máximo 1 MiB, sin texto después del JSON) y la traducción de errores a códigos (`statusFor`) | TEC-01, TEC-04 y el endpoint de cada historia |
| `internal/web` | `index.html`, `app.js` y `styles.css`, embebidos con `embed` y servidos en `/`. Sin frameworks: usa la API con `fetch` | TEC-04 y la interfaz de cada historia |
| `internal/postgres` | `Open`: abre el pool de `pgx/v5` y verifica la conexión; el error nunca incluye la URL | TEC-05 |
| `internal/postgrestest` | Solo para tests: conecta a `TEST_DATABASE_URL` (que debe terminar en `_test`), toma un lock de la base y vacía las tablas antes de cada caso | TEC-05 |

### 1.2 Dependencias entre paquetes

Cada flecha significa "usa". Los Manager se crean en `cmd/server/main.go` en este orden, y cada uno recibe los que necesita:

```mermaid
graph TD
    main[cmd/server] --> httpapi
    main --> postgres
    httpapi --> web
    httpapi --> metrics
    httpapi --> poker
    httpapi --> effort
    httpapi --> defect
    httpapi --> project
    httpapi --> backlog
    httpapi --> sprint
    metrics --> sprint
    metrics --> backlog
    poker --> backlog
    effort --> sprint
    effort --> backlog
    defect --> sprint
    defect --> backlog
    sprint --> backlog
    sprint --> project
    backlog --> project
```

- `backlog` usa `project` para verificar que el proyecto exista al crear o listar ítems.
- `sprint` usa `project` (el proyecto existe) y `backlog` (el ítem existe, es del mismo proyecto, se asigna y se completa a través de las reglas de HU-08).
- `poker` usa `backlog` para verificar el ítem y guardar los story points acordados.
- `effort` y `defect` usan `sprint` (`WithOpenSprint`: el sprint existe y está abierto mientras se registra) y `backlog` (el ítem existe).
- `metrics` lee ítems y sprints a través de sus Manager; `httpapi` lo crea en `NewHandler`.

Un Manager nunca accede al repositorio de otro: pasa siempre por su Manager, así las validaciones de cada dominio no se pueden saltear.

### 1.3 Concurrencia

El servidor atiende pedidos en paralelo. Cada Manager que lee y después escribe tiene un `sync.Mutex` que protege la operación completa, para que dos pedidos simultáneos no dejen, por ejemplo, dos proyectos con el mismo nombre (CA-01.7), dos sprints abiertos en un proyecto (CA-10.6) o dos sesiones de Poker abiertas para un ítem (CA-21.6). Cuando una operación toma el lock de dos Manager, el orden es siempre el mismo (por ejemplo, poker → backlog en HU-25), para evitar bloqueos mutuos. El CI corre los tests con `-race`.

### 1.4 Copias por valor

Las funciones de consulta y de escritura de los Manager y de los repositorios devuelven copias por valor (incluidas copias nuevas de los slices, como los criterios de aceptación o los votos), nunca punteros a las estructuras internas. Así ningún llamador puede cambiar el estado sin pasar por las validaciones (AGENTS.md).

## 2. Interfaz de repositorio

Cada dominio con datos propios define su interfaz `Repository` en `repository.go`. La lógica de negocio solo conoce la interfaz; `cmd/server/repositories.go` elige la implementación según `STORAGE`:

| Paquete | Operaciones de `Repository` |
|---|---|
| `project` | `Save`, `List`, `GetByID`, `Update` |
| `backlog` | `Save`, `ListByProject`, `GetByID`, `Update` |
| `sprint` | `Save`, `OpenByProject`, `GetByID`, `Update`, `ListByProject` |
| `poker` | `Save`, `OpenByItem`, `GetByID`, `Update` |
| `effort` | `Save`, `ListByItem`, `ListBySprint` |
| `defect` | `Save` |

`metrics` no tiene repositorio: calcula a partir de los Manager de `backlog` y `sprint`.

### 2.1 Implementación en memoria

`memory_repository.go` en cada paquete. Guarda los datos en un slice y asigna IDs correlativos desde 1 (`lastID`). Ante pedidos simultáneos lo protege el mutex de su Manager o, si el Manager no tiene uno (como `effort`), un mutex propio del repositorio. Es el valor por defecto (`STORAGE` sin definir o `memory`): no necesita Docker y los datos se pierden al detener el servidor. Es la que usan los tests unitarios y los escenarios BDD.

### 2.2 Implementación en PostgreSQL

`postgres_repository.go` en cada paquete (TEC-05). Usa el driver `github.com/jackc/pgx/v5` con un único pool (`pgxpool`) compartido por todos los repositorios, que abre `postgres.Open`. Los IDs los asigna la base (`GENERATED ... AS IDENTITY`). El esquema está en [`db/init/01-schema.sql`](../db/init/01-schema.sql), con las tablas `projects`, `backlog_items`, `sprints`, `efforts`, `defects` y `poker_sessions`; [`02-test-database.sql`](../db/init/02-test-database.sql) crea la base `gpc_test` con el mismo esquema.

### 2.3 Contrato común

Las dos implementaciones de cada repositorio pasan el mismo conjunto de casos: `runRepositoryContract`, en `repository_contract_test.go` de cada paquete. El test de memoria lo corre siempre; el de PostgreSQL (`postgres_repository_test.go`) lo corre solo si está definida `TEST_DATABASE_URL`. Así se garantiza que cambiar de almacenamiento no cambia el comportamiento.

## 3. API HTTP

Todas las rutas están en `internal/httpapi/server.go` (`NewHandler`) y usan solo `net/http`. Las respuestas son JSON (`Content-Type: application/json`). Las fechas van en formato `AAAA-MM-DD`.

### 3.1 Formatos de respuesta

| Objeto | Campos |
|---|---|
| Proyecto | `id`, `nombre`, `descripcion`, `estado`, `fecha_inicio` (`""` si no está registrada) |
| Ítem | `id`, `proyecto_id`, `titulo`, `descripcion`, `prioridad`, `criterios` (lista), `estado`, `story_points` (0 si no está estimado), `horas_estimadas`, `sprint_id` (0 si no está asignado) |
| Sprint | `id`, `proyecto_id`, `nombre`, `fecha_inicio`, `fecha_fin`, `objetivo`, `estado` |
| Sesión de Poker | `id`, `item_id`, `estado`, `ronda`, `votos` (lista de `{integrante, puntos}`), `cartas` (la escala) |
| Esfuerzo | `id`, `sprint_id`, `item_id`, `integrante`, `fecha`, `actividad`, `horas` |
| Lista de esfuerzos | `esfuerzos` (lista), `total_horas` (redondeado a un decimal) |
| Defecto | `id`, `descripcion`, `severidad`, `estado`, `item_id`, `sprint_deteccion_id`, `sprint_resolucion_id` (0 si no está resuelto) |
| Error | `error` (mensaje en español) |

Una lista vacía se devuelve como `[]`, nunca como `null`.

### 3.2 Endpoints

| Método y ruta | Cuerpo de entrada | Respuesta correcta | Historia |
|---|---|---|---|
| `POST /api/proyectos` | `{"nombre", "descripcion"}` | 201, proyecto | HU-01 |
| `GET /api/proyectos/{id}` | — | 200, proyecto | HU-03 |
| `PUT /api/proyectos/{id}` | `{"nombre", "descripcion", "estado"}` (reemplaza los tres) | 200, proyecto | HU-02 |
| `PATCH /api/proyectos/{id}/fecha-inicio` | `{"fecha_inicio"}` | 200, proyecto | HU-18 |
| `POST /api/proyectos/{id}/backlog` | `{"titulo", "descripcion", "prioridad", "criterios": [...]}` | 201, ítem | HU-04 |
| `GET /api/proyectos/{id}/backlog` | — (`?orden=prioridad` ordena ALTA, MEDIA, BAJA; sin el parámetro, por ID) | 200, lista de ítems | HU-06, HU-09 |
| `GET /api/items/{id}` | — | 200, ítem | HU-05 |
| `PUT /api/items/{id}` | `{"titulo", "descripcion", "prioridad", "criterios", "horas_estimadas"}` (reemplaza todos) | 200, ítem | HU-07 |
| `PATCH /api/items/{id}/estado` | `{"estado"}` | 200, ítem | HU-08 |
| `PATCH /api/items/{id}/story-points` | `{"story_points"}` | 200, ítem | HU-20 |
| `POST /api/items/{id}/poker-sessions` | — | 201, sesión | HU-21 |
| `POST /api/poker-sessions/{id}/votes` | `{"integrante", "puntos"}` | 200, sesión con `votos: []` (los votos no se muestran) | HU-22 |
| `POST /api/poker-sessions/{id}/reveal` | — | 200, sesión `REVELADA` con todos los votos | HU-23 |
| `POST /api/poker-sessions/{id}/estimate` | `{"puntos"}` | 200, sesión `CERRADA`; los puntos quedan en el ítem | HU-25 |
| `POST /api/proyectos/{id}/sprints` | `{"nombre", "fecha_inicio", "fecha_fin", "objetivo"}` | 201, sprint | HU-10 |
| `POST /api/sprints/{id}/items/{itemId}` | — | 200, ítem asignado | HU-11 |
| `POST /api/sprints/{id}/items/{itemId}/completar` | — | 200, ítem `TERMINADO` | HU-12 |
| `GET /api/sprints/{id}/completados` | — | 200, lista de ítems | HU-12 |
| `POST /api/sprints/{id}/cerrar` | — | 200, sprint `CERRADO` | HU-13 |
| `POST /api/sprints/{id}/items/{itemId}/esfuerzos` | `{"integrante", "fecha", "actividad", "horas"}` | 201, esfuerzo | HU-26 |
| `GET /api/sprints/{id}/esfuerzos` | — | 200, lista de esfuerzos | HU-27 |
| `GET /api/items/{id}/esfuerzos` | — | 200, lista de esfuerzos | HU-27 |
| `POST /api/items/{id}/defectos` | `{"descripcion", "severidad", "sprint_deteccion_id"}` | 201, defecto `ABIERTO` | HU-29 |
| `GET /api/sprints/{id}/metricas` | — | 200, `{sprint_id, sp_planificados, sp_completados, items_sin_estimar, items_totales, items_terminados, porcentaje_completadas}` | HU-32, HU-34 |
| `GET /api/proyectos/{id}/metricas` | — | 200, `{proyecto_id, sp_planificados, sp_completados, items_sin_estimar, velocidad, sprints_cerrados, items_totales, items_terminados, porcentaje_completadas}` | HU-32, HU-34 |
| `GET /` | — | 200, interfaz web | TEC-04 |

En las métricas, `velocidad` (promedio de los últimos 3 sprints cerrados) y `porcentaje_completadas` van redondeados a un decimal, o `null` si no hay datos (sin sprints cerrados o sin ítems). Las fórmulas están en [`docs/metricas.md`](metricas.md).

### 3.3 Errores

Todo error responde `{"error": "<mensaje>"}`. La traducción del error de dominio al código HTTP está en un solo lugar, `statusFor` (`internal/httpapi/server.go`):

| Código | Cuándo | Errores |
|---|---|---|
| 400 | Datos inválidos | JSON inválido o con texto después; `{id}` que no es un número; fecha que no es `AAAA-MM-DD`; nombre, título, objetivo, descripción, integrante, actividad o votante vacíos; prioridad, estado, severidad, story points o carta fuera de sus valores; ítem sin criterios de aceptación; horas estimadas negativas; horas trabajadas no positivas; fechas faltantes; fin de sprint no posterior al inicio; sprint de detección faltante |
| 404 | La entidad de la ruta no existe | Proyecto, ítem, sprint o sesión de Planning Poker inexistente |
| 405 | Método no permitido | Un método distinto del de la ruta (por ejemplo, `GET` en las rutas de sesiones de Planning Poker), con el encabezado `Allow`. La respuesta es texto, no JSON |
| 409 | El pedido choca con el estado actual | Nombre de proyecto repetido; transición de estado no permitida; ya hay un sprint abierto en el proyecto; sprint cerrado; ítem ya asignado, de otro proyecto o no asignado al sprint; story points de un ítem terminado; ya hay una sesión de Planning Poker abierta para el ítem, o el ítem está terminado; sesión no abierta, sin votos para revelar o no revelada; sprint de detección de otro proyecto |
| 413 | Cuerpo de más de 1 MiB | `el cuerpo del pedido supera el tamaño máximo de 1 MiB` |
| 500 | Error no previsto (por ejemplo, la base no responde) | Responde `error interno del servidor`, sin exponer el detalle |

Un error que no está en `statusFor` es 500, así que cada error nuevo de una historia se agrega ahí con su test.

### 3.4 Límites del servidor

- Cuerpo máximo: 1 MiB (CA-TEC04.5).
- Timeouts (CA-TEC04.7): 5 s para leer los encabezados, 10 s para leer el pedido, 10 s para escribir la respuesta y 60 s de conexión inactiva.

## 4. Cómo ejecutar y probar

Requisitos: Go 1.27.1 o superior (versión de `go.mod`). Docker con Docker Compose, solo para PostgreSQL.

### 4.1 Ejecutar la aplicación

```
go run ./cmd/server
```

Escucha en el puerto `8080` (o el de la variable `PORT`) y guarda los datos en memoria. La interfaz queda en <http://localhost:8080/> y la API en `http://localhost:8080/api/…`. Para usar PostgreSQL, ver la sección 5.

### 4.2 Tests unitarios

```
go test ./...
```

Corre todos los tests del módulo con el paquete estándar `testing`: los unitarios de cada Manager (`*_test.go`, con nombres `TestFuncion_Caso`), los de la capa HTTP (`internal/httpapi`, con `httptest`), los de contrato de los repositorios y los escenarios BDD. Opciones útiles:

- `go test -cover ./...`: cobertura por paquete.
- `go test -coverprofile=cover.out ./...` y `go tool cover -func=cover.out`: cobertura por función y total.
- `go test -race ./...`: detector de carreras (necesita cgo; en Windows sin compilador de C no corre, y lo verifica el CI).
- `go test -count=1 ./...`: sin usar resultados en caché.

### 4.3 Escenarios BDD (godog)

Los escenarios están en `features/HU-XX-nombre.feature`, en Gherkin y en español. Cada uno se ejecuta desde un test de Go del paquete que implementa la historia (`*_features_test.go`), con `godog.TestSuite` en modo estricto (`Strict: true`, un paso sin implementar hace fallar el test). Por eso corren con el mismo `go test ./...`. Para correr solo los de un paquete:

```
go test -run TestFeatures ./internal/backlog/
```

### 4.4 Pruebas de integración con PostgreSQL

Los tests de los repositorios PostgreSQL y del wiring de `cmd/server` corren solo si está definida `TEST_DATABASE_URL`; si no, hacen `t.Skip` (por eso el CI pasa sin base). Con la base levantada (sección 5):

```
$env:TEST_DATABASE_URL="postgres://gpc:gpc_dev@localhost:5433/gpc_test?sslmode=disable"; go test -count=1 ./...
```

En Git Bash, Linux o macOS: `TEST_DATABASE_URL="postgres://gpc:gpc_dev@localhost:5433/gpc_test?sslmode=disable" go test -count=1 ./...`.

- La base tiene que terminar en `_test`: `postgrestest` vacía las tablas antes de cada caso y se niega a conectarse a otra.
- Los paquetes comparten la base de pruebas: un lock de PostgreSQL (`pg_advisory_lock`) hace que esos tests corran de a uno.
- Para medir la cobertura de todo el módulo junto: `go test -count=1 -coverpkg=./... -coverprofile=cover.out ./...`.

## 5. PostgreSQL con Docker

### 5.1 Levantar la base

```
docker compose up -d --wait
```

Levanta PostgreSQL 17 (`docker-compose.yml`) con los valores de `.env` o, si no existe, los de desarrollo de [`.env.example`](../.env.example): usuario `gpc`, contraseña `gpc_dev`, base `gpc`, publicada en `127.0.0.1:5433` (no en el 5432, para no chocar con un PostgreSQL instalado). Los scripts de `db/init/` crean el esquema y la base `gpc_test` **solo la primera vez**, con el volumen vacío. Si el esquema cambió, hay que recrear el volumen con `docker compose down -v` (borra los datos) y volver a levantar la base.

### 5.2 Variables de entorno

| Variable | Valores | Uso |
|---|---|---|
| `PORT` | Número de puerto (por defecto `8080`) | Puerto del servidor |
| `STORAGE` | `memory` (por defecto) o `postgres` | Dónde se guardan los datos. Se compara tal cual: otro valor hace que el servidor no arranque |
| `DATABASE_URL` | URL de PostgreSQL | Obligatoria con `STORAGE=postgres` |
| `TEST_DATABASE_URL` | URL de una base que termina en `_test` | Solo para las pruebas de integración |
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_PORT` | Ver `.env.example` | Solo las lee `docker compose` |

La aplicación no lee el archivo `.env`: las variables se definen en la terminal. Por ejemplo, en PowerShell:

```
$env:STORAGE="postgres"; $env:DATABASE_URL="postgres://gpc:gpc_dev@localhost:5433/gpc?sslmode=disable"; go run ./cmd/server
```

Al arrancar, el servidor muestra el almacenamiento elegido: `servidor escuchando en :8080 (almacenamiento: postgres)`. Si la configuración es inválida o la base no responde, no arranca y muestra el error, que nunca incluye la URL (tiene la contraseña).

## 6. Flujo de trabajo del equipo

Las reglas completas están en [`AGENTS.md`](../AGENTS.md), y el checklist previo a cada historia en [`docs/antes-de-empezar.md`](antes-de-empezar.md). En resumen:

- **Cadena por historia:** Historia de usuario → SDD (`docs/especificaciones/HU-XX-*.md`, aprobada por el Agile Enabler antes de seguir) → escenarios BDD (`features/HU-XX-*.feature`) → tests (RED) → código (GREEN) → refactor → evidencia (`docs/evidencias/HU-XX-tdd.md`) y cobertura (`docs/reportes/cobertura.md`). Las tareas técnicas (TEC-XX) parten de los criterios de su Issue, sin historia ni SDD propias.
- **Ramas:** una por historia o tarea, desde `main` actualizado: `feature/HU-XX-nombre-corto` o `feature/TEC-XX-nombre-corto`. Nunca se trabaja sobre `main`.
- **Commits:** `tipo(ID): descripción`, con `tipo` en `test`, `feat`, `refactor`, `docs`, `ci`, `chore` o `fix`, e `ID` de la historia, la tarea o el sprint (por ejemplo `test(HU-01): agregar tests (RED)`). El cuerpo lleva `Refs #N` con el Issue; el commit de cierre lleva `Closes #N`.
- **Pull Request:** hacia `main`, con la aprobación de un integrante distinto del responsable y la CI en verde. Se integra con "Create a merge commit" (nunca squash ni rebase), para conservar el historial RED → GREEN → REFACTOR.
- **CI** ([`.github/workflows/ci.yml`](../.github/workflows/ci.yml), TEC-02): en cada Pull Request hacia `main` verifica el formato (`gofmt -l`), corre `go vet ./...` y `go test -race -coverprofile=coverage.out ./...`, y muestra la cobertura total. No tiene base de datos: los tests de PostgreSQL se omiten.
- **Trazabilidad:** cada Issue de historia tiene la sección "## Trazabilidad" con los links a su SDD, BDD y evidencia.
- **Definition of Ready y Done:** [`docs/definition-of-ready-done.md`](definition-of-ready-done.md).
- **Uso de IA:** se registra en [`docs/ia/registro.md`](ia/registro.md).

## 7. Dependencias externas

| Librería | Uso | Aprobación |
|---|---|---|
| `github.com/jackc/pgx/v5` v5.11.0 | Driver y pool de PostgreSQL (código de producción) | TEC-05 (motivos en el [README](../README.md#driver-de-postgresql)) |
| `github.com/cucumber/godog` v0.16.0 | Ejecución de los escenarios BDD (solo tests) | Aprobada por el equipo (AGENTS.md) |

El resto usa solo la biblioteca estándar de Go. Agregar otra librería requiere la aprobación del equipo.
