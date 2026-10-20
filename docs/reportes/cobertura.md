# Reporte de cobertura

Cobertura de los tests automatizados (en el CI, `go test -race -coverprofile=coverage.out ./...`), actualizada al cerrar cada historia.

## Cobertura actual

- **Fecha:** 20/10/2026
- **Última historia con cobertura medida:** HU-12 · Registrar historias completadas (rama `feature/HU-12-registrar-completadas`, pendiente de integrar a `main` por Pull Request)
- **Comando:** `go test -race -coverprofile=coverage.out ./...` en el CI (job "gofmt y tests" del PR #100), sin base. Los valores de la columna "Sin base (CI)" salen del log del CI del PR #100: [job 118444923147](https://github.com/Elwilsonic/seguimiento-medicion-GPC/actions/runs/39366654421/job/118444923147).

| Paquete | Sin base (CI) | Con base (HU-11, última medición con base) |
|---|---|---|
| `cmd/server` | 39.0% | 41.5% |
| `internal/backlog` | 80.9% | 98.7% |
| `internal/httpapi` | 100.0% | 100.0% |
| `internal/postgres` | 83.3% | 83.3% |
| `internal/postgrestest` | 36.8% | 36.8% |
| `internal/project` | 67.4% | 96.8% |
| `internal/sprint` | 76.9% | 97.6% |

Los valores de "Sin base (CI)" son los del log del CI del PR #100. `internal/sprint` sube de 69.5% en `main` a 76.9% con HU-12. El total del perfil de `go test -race -coverprofile=coverage.out ./...` es **74.6%**, también tomado del CI del PR #100. La columna "Con base" conserva la última medición con base (HU-11), porque HU-12 no modifica PostgreSQL y no se repitió esa medición. La medición local en Windows queda fuera del reporte: el CI (Linux) es la fuente autoritativa.

Desde TEC-05, `internal/backlog`, `internal/project` e `internal/sprint` tienen un repositorio PostgreSQL además del de memoria. En el CI no hay base, así que sus tests de integración hacen `t.Skip` y ese código queda sin ejecutar: por eso bajan la cobertura de la columna "Sin base". **La cobertura de la parte de PostgreSQL se mide localmente, con la base levantada** (ver la [evidencia de TEC-05](../evidencias/TEC-05-tdd.md)).

- `go test -cover` mide cada paquete solo con sus propios tests. `internal/postgrestest` queda en 36.8% porque su `Pool`, su lock y su `Reset` los usan los tests de los otros paquetes.
- El total del perfil de `go test -race -coverprofile=coverage.out ./...` en el CI del PR #100 es **74.6%**.
- Con `-coverpkg=./...` (todos los paquetes medidos juntos, con base) el total fue **92.5%** en la última medición con base (HU-11). HU-12 no modifica PostgreSQL, así que no se repitió esa medición.
- Lo que queda sin cubrir son ramas de error que necesitan que la base falle en un momento puntual (el `GetByID` de sprints en PostgreSQL queda en 85.7% por la misma rama de error genérico que `OpenByProject`), y `main()`.

`cmd/server` no llega al 100% porque `main()` arranca el servidor y bloquea en `http.ListenAndServe`, así que no se prueba con un test unitario. La lógica con decisiones está separada en `address`, `storageConfig` y `newRepositories`, las tres con cobertura total.

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

Las funciones nuevas de HU-12 entran con cobertura total en memoria: `CompleteItem` y `ListCompletedItems` quedan al 100.0% en `internal/sprint/complete.go`. HU-12 no agrega operaciones al repositorio ni columnas a PostgreSQL (RN-12.9).

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
