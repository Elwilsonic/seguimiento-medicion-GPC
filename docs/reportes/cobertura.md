# Reporte de cobertura

Cobertura de los tests automatizados (en el CI, `go test -race -coverprofile=coverage.out ./...`), actualizada al cerrar cada historia.

## Cobertura actual

- **Fecha:** 24/10/2026
- **Última historia con cobertura medida:** HU-26 — Registrar esfuerzo (rama `feature/HU-26-registrar-esfuerzo`)
- **Comando:** `go test -race -coverprofile=coverage.out ./...`, sin base, el mismo del CI (job "gofmt y tests"). HU-26 se midió localmente con go1.27.2 windows/amd64 y `go test -count=1 -cover ./...` / `-coverprofile`, sin `-race` (el entorno no tiene cgo) y sin `TEST_DATABASE_URL`; la comparación con `main` (`ef2c361`) se hizo con el mismo Go en una copia temporal (`git worktree`). El check "gofmt y tests" del PR #120 pasó en verde sobre `1a8b8bf` ([job 120962343788](https://github.com/Elwilsonic/seguimiento-medicion-GPC/actions/runs/40124311691/job/120962343788?pr=120)); las cifras de la tabla son la medición local. El cambio de go1.27.1 a go1.27.2 alteró el conteo de sentencias: por eso `internal/backlog`, `internal/project` e `internal/sprint` difieren de las cifras de HU-20 y no son comparables con ellas.

| Paquete | Sin base (HU-26, local) | Con base (HU-11, última medición con base) |
|---|---|---|
| `cmd/server` | 53.1% | 41.5% |
| `internal/backlog` | 79.9% | 98.7% |
| `internal/effort` | 74.4% | — |
| `internal/httpapi` | 100.0% | 100.0% |
| `internal/web` | 0.0% (100.0% medido junto con `internal/httpapi`) | — |
| `internal/postgres` | 83.3% | 83.3% |
| `internal/postgrestest` | 36.8% | 36.8% |
| `internal/project` | 64.4% | 96.8% |
| `internal/sprint` | 75.6% | 97.6% |

Con HU-26, `internal/effort` entra en 74.4% y `internal/sprint` sube de 74.0% en `main` (`ef2c361`, mismo Go) a 75.6% con `WithOpenSprint`; `internal/httpapi` sigue en 100.0% con el endpoint y la interfaz nuevos. `cmd/server` baja de 54.8% a 53.1% porque `main()` y `newRepositories` suman código nuevo (el `effort.Manager` y la rama de esfuerzos de PostgreSQL) que no se ejecuta sin base. `internal/backlog` (79.9%) e `internal/project` (64.4%) no cambian. El total del perfil sube de 80.0% a **80.4%**. `internal/web` (TEC-04, PR 3) figura en 0.0% porque no tiene tests propios: a su única función, `Handler`, la ejecutan los tests de `internal/httpapi`, y con `-coverpkg=./internal/web,./internal/httpapi` queda al 100.0%. La columna "Con base" conserva la última medición por paquete (HU-11); la medición con base de HU-26 es **94.8%** con `-coverpkg=./...` (total, verificada por Hernán), no por paquete. La medición local en Windows queda fuera del reporte: el CI (Linux) es la fuente autoritativa.

Desde TEC-05, `internal/backlog`, `internal/project` e `internal/sprint` tienen un repositorio PostgreSQL además del de memoria. En el CI no hay base, así que sus tests de integración hacen `t.Skip` y ese código queda sin ejecutar: por eso bajan la cobertura de la columna "Sin base". **La cobertura de la parte de PostgreSQL se mide localmente, con la base levantada** (ver la [evidencia de TEC-05](../evidencias/TEC-05-tdd.md)).

- `go test -cover` mide cada paquete solo con sus propios tests. `internal/postgrestest` queda en 36.8% porque su `Pool`, su lock y su `Reset` los usan los tests de los otros paquetes.
- El total del perfil sin base es **80.4%** con HU-26 (medido localmente con go1.27.2; 80.0% en `main` `ef2c361` con el mismo comando), después de 80.6% con HU-20 (go1.27.1, confirmado en el CI del PR #118 con `-race`).
- Con `-coverpkg=./...` (todos los paquetes medidos juntos, con base) el total fue **94.8%** con HU-26 (medido por Hernán con la base levantada, incluye `postgres_repository.go` de esfuerzos), por encima del 92.5% de la última medición con base (HU-11).
- Lo que queda sin cubrir son ramas de error que necesitan que la base falle en un momento puntual (el `GetByID` de sprints en PostgreSQL queda en 85.7% por la misma rama de error genérico que `OpenByProject`), la parte PostgreSQL de `Update` y de `Save` de esfuerzos (que se omite sin base) y `main()`.

`cmd/server` no llega al 100% porque `main()` arranca el servidor y bloquea en `http.ListenAndServe`, así que no se prueba con un test unitario. La lógica con decisiones está separada en `address`, `storageConfig`, `newRepositories` y `newServer` (TEC-04): `address`, `storageConfig` y `newServer` están al 100.0%, y `newRepositories` queda en 83.3% por la rama de PostgreSQL que no se ejecuta sin base.

```
$ go test -count=1 -coverpkg=./... -coverprofile=cover.out ./...      (con TEST_DATABASE_URL)
$ go tool cover -func=cover.out      (código de TEC-05 y total)
cmd/server/main.go:17:				main			0.0%
cmd/server/main.go:38:				address			100.0%
cmd/server/repositories.go:24:			newRepositories		100.0%
cmd/server/storage.go:32:			storageConfig		100.0%
internal/backlog/postgres_repository.go:21:	NewPostgresRepository	100.0%
internal/backlog/postgres_repository.go:31:	Save			100.0%
internal/backlog/postgres_repository.go:46:	GetByID			100.0%
internal/backlog/postgres_repository.go:60:	Update			100.0%
internal/backlog/postgres_repository.go:80:	ListByProject		77.8%
internal/backlog/postgres_repository.go:98:	criteriaParam		100.0%
internal/backlog/postgres_repository.go:107:	scanItem		100.0%
internal/postgres/postgres.go:23:		Open			91.7%
internal/postgrestest/postgrestest.go:59:	Pool			77.8%
internal/postgrestest/postgrestest.go:77:	lock			72.7%
internal/postgrestest/postgrestest.go:100:	Reset			75.0%
internal/postgrestest/postgrestest.go:111:	decide			100.0%
internal/postgrestest/postgrestest.go:129:	checkDatabaseName	100.0%
internal/project/postgres_repository.go:21:	NewPostgresRepository	100.0%
internal/project/postgres_repository.go:30:	Save			80.0%
internal/project/postgres_repository.go:42:	List			75.0%
internal/project/postgres_repository.go:58:	GetByID			100.0%
internal/project/postgres_repository.go:72:	Update			100.0%
internal/project/postgres_repository.go:87:	scanProject		100.0%
internal/sprint/postgres_repository.go:21:	NewPostgresRepository	100.0%
internal/sprint/postgres_repository.go:31:	Save			100.0%
internal/sprint/postgres_repository.go:45:	OpenByProject		85.7%
internal/sprint/postgres_repository.go:61:	GetByID			85.7%
internal/sprint/postgres_repository.go:76:	scanSprint		100.0%
total:									(statements)		92.5%
```

HU-13 agrega `Close` y `Update` en memoria y PostgreSQL. En el CI, `internal/sprint` sube a 77.3% y el total a 74.8%. La cobertura de `Update` en PostgreSQL queda pendiente de la medición con base de Buttini o Cardozo.

## Historial

| Fecha | Historia | Paquete | Cobertura | Observaciones |
|---|---|---|---|---|
| 10/10/2026 | HU-01 | `internal/project` | 95.7% | Sin cubrir: el error de `List()` del repositorio, que en memoria nunca falla (se cubre con TEC-05) |
| 11/10/2026 | HU-01 (ajustes de la revisión) | `internal/project` | 100.0% | Se cubrió el error del repositorio con un repositorio falso |
| 11/10/2026 | HU-03 | `internal/project` | 100.0% | Se agregó `Get`/`GetByID` con cobertura total; verificación con `-race` hecha por Peñalbé Hernán y por el CI del PR |
| 12/10/2026 | HU-04 | `internal/backlog` (nuevo) e `internal/project` | 100.0% y 100.0% | Nuevo paquete del backlog con cobertura total; sin regresiones en `internal/project`; `-race` (CA-04.7) lo verifica el CI porque el entorno local no tiene cgo |
| 12/10/2026 | HU-06 | `internal/backlog` e `internal/project` | 100.0% y 100.0% | `ListItems` y `ListByProject` entran con cobertura total; sin regresiones en `internal/project` ni en HU-04; `-race` (CA-06.5) lo verifica el CI porque el entorno local no tiene cgo |
| 13/10/2026 | HU-02 | `internal/project` e `internal/backlog` | 100.0% y 100.0% | `Update` y `normalizeStatus` entran con cobertura total (tras un ajuste: el test de contrato de `MemoryRepository.Update`, ver la desviación del ciclo TDD en la evidencia de HU-02); sin regresiones en backlog ni en HU-01/HU-03; `-race` (CA-02.7) lo verifica el CI y Hernán sobre Linux porque el entorno local no tiene cgo |
| 14/10/2026 | TEC-01 | `cmd/server` (nuevo), `internal/httpapi` (nuevo), `internal/backlog` e `internal/project` | 28.6%, 100.0%, 100.0% y 100.0% | Servidor HTTP con `net/http`; `internal/httpapi` llega al 100% tras un ajuste (test del id no numérico en las demás rutas, ver la desviación del ciclo TDD en la evidencia de TEC-01); `main()` no se prueba con tests unitarios; `-race` verificado sobre Linux |
| 15/10/2026 | HU-05 | `internal/backlog` | 100.0% | `GetItem` y `GetByID` entran con cobertura total; sin regresiones en HU-04, HU-06 ni en el resto de los paquetes; `-race` (CA-05.4) lo verifican Hernán sobre Linux y el CI porque el entorno local no tiene cgo |
| 17/10/2026 | HU-10 | `internal/sprint` (nuevo) | 100.0% | Nuevo paquete de sprints con cobertura total (`Create`, `Save`, `OpenByProject`); sin regresiones en el resto de los paquetes; `-race` (CA-10.6) lo verifica el CI porque el entorno local no tiene cgo, y la revisión de la rama sobre Linux queda a cargo de Peñalbé Hernán |
| 17/10/2026 | HU-07 | `internal/backlog` | 100.0% | `Manager.Update` y `validateEstimatedHours` (`update.go`) y `MemoryRepository.Update` entran con cobertura total; el campo `EstimatedHours` entra en `Item` sin bajas; sin regresiones en HU-04, HU-05, HU-06 ni en el resto de los paquetes; `-race` (CA-07.6) lo verifica el CI del Pull Request porque el entorno local no tiene cgo |
| 17/10/2026 | TEC-05 | `internal/postgres` (nuevo), `internal/postgrestest` (nuevo), `internal/project`, `internal/backlog`, `internal/sprint` y `cmd/server` | Con base: 83.3%, 36.8%, 96.8%, 98.6%, 98.3% y 41.5%; sin base (CI): 83.3%, 36.8%, 67.4%, 75.5%, 69.0% y 39.0% | Repositorios PostgreSQL detrás de las mismas interfaces; en el CI sus tests hacen `t.Skip`, así que la cobertura de PostgreSQL se mide localmente con la base levantada (con `-coverpkg=./...`: 92.1%); sin cubrir quedan ramas de error que necesitan una falla puntual de la base; `-race` lo verifica el CI del Pull Request porque el entorno local no tiene cgo |
| 18/10/2026 | HU-11 | `internal/sprint` y `internal/backlog` | Con base: 97.6% y 98.7%; sin base (CI): 69.5% y 77.7% | `AssignItem`, `AssignToSprint` y los `GetByID` de sprints entran con cobertura total en memoria; la parte de PostgreSQL (`GetByID` de sprints al 85.7%, por la rama de error genérico igual que `OpenByProject`, y `sprint_id` en el repositorio de backlog) se mide con la base levantada, recreada para la columna nueva (`docker compose down -v`); total con `-coverpkg=./...`: 92.5%; `-race` (CA-11.6) lo verifica el CI del Pull Request porque el entorno local no tiene cgo |
| 19/10/2026 | HU-08 | `internal/backlog` | 80.9% (sin base) | `ChangeStatus`, `normalizeStatus` e `isTransitionAllowed` entran con tests de las cuatro transiciones válidas, las dos no permitidas, el no-op y los errores del repositorio; sin base (CI) el total del perfil con `-coverprofile=cover.out` es 73.4%; el contrato contra PostgreSQL hizo SKIP por no tener `TEST_DATABASE_URL` y se mide localmente con la base levantada; `-race` (CA-08.9) lo verifica el CI del Pull Request porque el entorno local no tiene cgo |
| 20/10/2026 | HU-12 | `internal/sprint` | 76.9% (CI, sin base); total 74.6% | `internal/sprint` sube de 69.5% en `main` a 76.9% con HU-12; `CompleteItem` y `ListCompletedItems` entran con cobertura total (100.0%) en memoria; no se agregaron operaciones al repositorio ni columnas a PostgreSQL; la medición con base no se repitió porque HU-12 no toca la persistencia y se conserva la de HU-11; `-race` (CA-12.9) lo verifica el CI del Pull Request porque el entorno local no tiene cgo |
| 20/10/2026 | HU-13 | `internal/sprint` | 77.3% (CI, sin base); total 74.8% | `internal/sprint` sube de 76.9% a 77.3% con HU-13; `Close` y `Update` en memoria entran con tests; la parte PostgreSQL de `Update` queda pendiente de la medición con base; `-race` (CA-13.8) verificado por el CI y por Hernán sobre Linux con dos pruebas de mutación |
| 20/10/2026 | HU-09 | `internal/backlog` | 82.0% (sin base); total 75.3% | `ListItemsByPriority` y `priorityRank` (`priority.go`) entran con cobertura total (100.0%); `internal/backlog` sube de 80.9% en `main` (medido con `go test -count=1 -cover ./internal/backlog` sobre `46a2e8a`) a 82.0%; medido localmente con `go test -count=1 -coverprofile=cover.out ./...` sin `TEST_DATABASE_URL`, igual que el CI; HU-09 no cambia la interfaz `Repository` ni el esquema de PostgreSQL, así que no se repitió la medición con base; `-race` (CA-09.9) lo verifica el CI del Pull Request porque el entorno local no tiene cgo |
| 20/10/2026 | TEC-04 (PR 1) | `cmd/server` e `internal/httpapi` | 40.5% y 100.0% (CI, sin base); total 75.8% | `newServer` (`main.go`), `decodeJSON` y `bodyError` (`server.go`) entran con cobertura total; `cmd/server` sube de 39.0% a 40.5% y el total de 75.3% (HU-09) a 75.8%; no cambia PostgreSQL |
| 21/10/2026 | TEC-04 (PR 2) | `internal/httpapi` y `cmd/server` | 100.0% y 37.8% (sin base, mismo comando del CI); total 79.3% | Los endpoints de ítems (HU-05, HU-07, HU-08, HU-09) y de sprints (HU-10 a HU-13) entran con cobertura total, separados en `projects.go`, `items.go` y `sprints.go`; `cmd/server` baja de 40.5% a 37.8% porque `main()` arma el `sprint.Manager`; el total sube de 75.8% a 79.3%; no cambia PostgreSQL |
| 21/10/2026 | TEC-04 (PR 3) | `internal/web` (nuevo) e `internal/httpapi` | 0.0% propio (100.0% con `-coverpkg` junto a `internal/httpapi`) y 100.0%; total 79.7% | La interfaz web (HTML, JavaScript y CSS embebidos) se sirve en `/`; `web.Handler` lo cubren los tests de `internal/httpapi`; el JavaScript no entra en la cobertura de Go y se verificó con el recorrido en el navegador de la evidencia; el total sube de 79.3% a 79.7%; no cambia PostgreSQL |
| 23/10/2026 | HU-20 | `internal/backlog` e `internal/httpapi` | 83.3% y 100.0% (sin base); total 80.6%; confirmadas en el CI del PR #118 (con `-race`) | `StoryPointScale`, `AssignStoryPoints` (`story_points.go`), el handler `assignStoryPoints` y `statusFor` entran con cobertura total (100.0%); `internal/backlog` sube de 82.0% en `main` (`e04c875`, medido con el mismo comando en una copia temporal) a 83.3% y el total de 79.7% a 80.6%; medido localmente en Windows con `go test -count=1 -cover ./...` y `-coverprofile`, sin `TEST_DATABASE_URL`; HU-20 no cambia la interfaz `Repository` ni el esquema de PostgreSQL, así que no se repitió la medición con base; `-race` (CA-20.7) no se corrió localmente (sin cgo) y el CI del PR #118 (ejecución #59, Linux) pasó con `go test -race -cover ./...`, sin carreras y con las mismas cifras |
| 24/10/2026 | HU-26 | `internal/effort` (nuevo), `internal/sprint`, `internal/httpapi` y `cmd/server` | 74.4%, 75.6%, 100.0% y 53.1% (local, sin base); total 80.4%; con base: 94.8% con `-coverpkg=./...` | `Register`, `NewManager`, `Save` en memoria y `WithOpenSprint` entran con cobertura total (100.0%); `internal/effort` queda en 74.4% sin base porque `postgres_repository.go` no se ejecuta sin `TEST_DATABASE_URL`; `internal/sprint` sube de 74.0% en `main` (`ef2c361`, mismo Go) a 75.6%, `internal/httpapi` sigue en 100.0% y el total sube de 80.0% a 80.4%; `cmd/server` baja de 54.8% a 53.1% por el código nuevo de `main()` y `newRepositories`; con la base levantada, Hernán midió 94.8% con `-coverpkg=./...`; el check "gofmt y tests" del PR #120 pasó en verde sobre `1a8b8bf`; `-race` (CA-26.10) lo corrió Hernán contra la base y lo verifica el CI del Pull Request |
