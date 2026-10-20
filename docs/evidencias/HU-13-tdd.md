# Evidencia TDD — HU-13 · Cerrar sprint

> Issue del repositorio: #27

- **Responsable:** García Santiago
- **Rama:** `feature/HU-13-cerrar-sprint`
- **Fecha:** 20/10/2026
- **Go:** go1.27.2 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` lo verificó Hernán sobre Linux; ver la sección de `go test -race`)
- **Especificación SDD:** [`docs/especificaciones/HU-13-cerrar-sprint.md`](../especificaciones/HU-13-cerrar-sprint.md)
- **Escenarios BDD:** [`features/HU-13-cerrar-sprint.feature`](../../features/HU-13-cerrar-sprint.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**14 funciones de test cubriendo 10 escenarios BDD (10 escenarios, sin esquemas del escenario, 71 pasos)**: 13 tests unitarios en `internal/sprint/close_test.go` y 1 test que ejecuta los escenarios BDD con godog (`internal/sprint/close_features_test.go`, runner `TestFeatures_HU13CerrarSprint`, en el paquete `sprint` porque la historia orquesta los paquetes `sprint` y `backlog`). El conteo sale del comando:

```
$ Select-String -Path internal\sprint\close_test.go,internal\sprint\close_features_test.go -Pattern '^func Test'
internal\sprint\close_test.go:81:func TestClose_SprintAbiertoQuedaCerrado(t *testing.T) {
internal\sprint\close_test.go:117:func TestClose_SprintSinItemsQuedaCerrado(t *testing.T) {
internal\sprint\close_test.go:132:func TestClose_DespuesDelCierreSePuedeCrearOtroSprint(t *testing.T) {
internal\sprint\close_test.go:154:func TestClose_SprintCerradoNoAdmiteAsignarNiCompletar(t *testing.T) {
internal\sprint\close_test.go:182:func TestClose_SprintYaCerradoDaErrSprintClosed(t *testing.T) {
internal\sprint\close_test.go:201:func TestClose_IDInexistenteDaErrSprintNotFound(t *testing.T) {
internal\sprint\close_test.go:222:func TestClose_ErrorDelRepositorioAlConsultarSeDevuelve(t *testing.T) {
internal\sprint\close_test.go:238:func TestClose_ErrorDelRepositorioAlActualizarSeDevuelve(t *testing.T) {
internal\sprint\close_test.go:271:func TestClose_ModificarElSprintDevueltoNoAlteraElGuardado(t *testing.T) {
internal\sprint\close_test.go:302:func TestClose_SprintConItemsPendientesYEnCursoSeCierraSinModificarlos(t *testing.T) {
internal\sprint\close_test.go:327:func TestClose_ItemsNoTerminadosNoSePuedenReasignar(t *testing.T) {
internal\sprint\close_test.go:359:func TestClose_CierresSimultaneosDejanUnSoloCierreExitoso(t *testing.T) {
internal\sprint\close_test.go:398:func TestClose_EnParaleloConAssignItemYCompleteItem(t *testing.T) {
internal\sprint\close_features_test.go:20:func TestFeatures_HU13CerrarSprint(t *testing.T) {
```

Además de los 13 tests unitarios, `TestClose_IDInexistenteDaErrSprintNotFound` tiene subtests con los IDs 999, 0 y -1.

El contrato de `Update` se agregó a `runRepositoryContract` en `internal/sprint/repository_contract_test.go` con **3 casos**: actualiza y devuelve una copia, conserva el ID sin alterar otros sprints, y un ID inexistente da `ErrSprintNotFound`. Sin `TEST_DATABASE_URL`, esos 3 casos corren en memoria y la versión PostgreSQL queda omitida. Se agregó además **1 test específico de PostgreSQL** (`TestPostgresRepository_UpdatePersisteElEstadoYDevuelveFechasUTC`), que también queda omitido sin la base. El conteo de los tests nuevos de `Update` sin base es **3 ejecuciones y 4 SKIP** (3 casos de contrato en PostgreSQL + 1 test específico).

La correspondencia no es 1 a 1: los errores del repositorio (CA-13.5) no tienen escenario BDD y se cubren con dos tests unitarios; la imposibilidad de reasignar ítems no terminados (CA-13.7) también es solo unitaria.

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-13.1 Un sprint `ABIERTO` queda `CERRADO`, como copia, sin cambiar el resto y liberando el proyecto | Cerrar un sprint abierto; Cerrar un sprint sin ítems; Crear un nuevo sprint después del cierre | `TestClose_SprintAbiertoQuedaCerrado`, `TestClose_SprintSinItemsQuedaCerrado`, `TestClose_DespuesDelCierreSePuedeCrearOtroSprint` | `close.go` (`Close`), `memory_repository.go` y `postgres_repository.go` (`Update`) |
| CA-13.2 Un sprint `CERRADO` no admite asignaciones ni completados | Un sprint cerrado no admite asignaciones ni completados | `TestClose_SprintCerradoNoAdmiteAsignarNiCompletar` | `close.go`, `assign.go` y `complete.go` (`ErrSprintClosed` de HU-11 y HU-12) |
| CA-13.3 Cerrar un sprint ya `CERRADO` da `ErrSprintClosed` sin modificar nada | Cerrar un sprint ya cerrado | `TestClose_SprintYaCerradoDaErrSprintClosed` | `close.go` (RN-13.3) |
| CA-13.4 Un SprintID inexistente (999, 0, -1) da `ErrSprintNotFound` sin modificar nada | Cerrar un sprint inexistente | `TestClose_IDInexistenteDaErrSprintNotFound` (3 subtests) | `close.go` y `GetByID` de HU-10/HU-11 |
| CA-13.5 El error del repositorio al consultar o actualizar se devuelve y el sprint no queda cerrado | Error del repositorio al cerrar | `TestClose_ErrorDelRepositorioAlConsultarSeDevuelve`, `TestClose_ErrorDelRepositorioAlActualizarSeDevuelve` | `close.go` (RN-13.8), `repository.go` |
| CA-13.6 Modificar el sprint devuelto no altera el guardado | Modificar el sprint devuelto no altera el guardado | `TestClose_ModificarElSprintDevueltoNoAlteraElGuardado` | `close.go` y repositorios (copias por valor) |
| CA-13.7 Se cierra con ítems `PENDIENTE` o `EN_CURSO`; conservan estado y `SprintID` y no se reasignan | Cerrar un sprint con ítems pendientes y en curso | `TestClose_SprintConItemsPendientesYEnCursoSeCierraSinModificarlos`, `TestClose_ItemsNoTerminadosNoSePuedenReasignar` | `close.go`, `assign.go` (`ErrItemAlreadyAssigned`) |
| CA-13.8 Cierre paralelo con asignación y completado: ninguna escritura de backlog después del cierre | Cierre simultáneo con asignación y completado | `TestClose_CierresSimultaneosDejanUnSoloCierreExitoso`, `TestClose_EnParaleloConAssignItemYCompleteItem` | `close.go`, `assign.go` y `complete.go` (mutex sprint → backlog) |

## Fase RED

Tests, escenarios BDD y runner escritos antes del código de producción (commit `6729f3b`, después del commit de SDD `8dca7fa`). **El RED fue un error de compilación, no el fallo de una aserción**: `Manager.Close` y `Repository.Update` todavía no existían. Antes de commitear el RED se agregó `Update` a `failingRepository` en `create_test.go` para que el GREEN no tuviera que tocar tests.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
FAIL	seguimiento-medicion-gpc/internal/sprint [build failed]
FAIL
go : # seguimiento-medicion-gpc/internal/sprint [seguimiento-medicion-gpc/internal/sprint.test]
internal\sprint\close_features_test.go:168:35: s.manager.Close undefined (type *Manager has no field or method Close)
internal\sprint\close_features_test.go:217:25: manager.Close undefined (type *Manager has no field or method Close)
internal\sprint\close_test.go:42:29: r.Repository.Update undefined (type Repository has no field or method Update)
internal\sprint\close_test.go:86:27: s.manager.Close undefined (type *Manager has no field or method Close)
internal\sprint\close_test.go:122:25: s.manager.Close undefined (type *Manager has no field or method Close)
internal\sprint\close_test.go:137:25: s.manager.Close undefined (type *Manager has no field or method Close)
internal\sprint\close_test.go:162:25: s.manager.Close undefined (type *Manager has no field or method Close)
internal\sprint\close_test.go:187:27: s.manager.Close undefined (type *Manager has no field or method Close)
internal\sprint\close_test.go:209:29: s.manager.Close undefined (type *Manager has no field or method Close)
internal\sprint\close_test.go:227:25: manager.Close undefined (type *Manager has no field or method Close)
internal\sprint\close_test.go:227:25: too many errors
```

`go test -gcflags=-e ./internal/sprint` mostró los 20 errores completos: 15 de `Close` y 5 de `Update`, sin ningún otro símbolo. Todos corresponden a la funcionalidad nueva de HU-13; el resto de los paquetes queda en verde.

## Fase GREEN

Código mínimo de producción (commit `a3cbfae`), sin tocar tests:

- `internal/sprint/close.go` (nuevo): `Manager.Close` toma el mutex del sprint, valida existencia, exige `ABIERTO`, cambia solo `Status` a `CERRADO` y persiste con `Update`.
- `internal/sprint/repository.go`: `Update(sprint Sprint) (Sprint, error)` en la interfaz.
- `internal/sprint/memory_repository.go`: `Update` reemplaza el sprint con ese ID y devuelve la copia; si no existe, `ErrSprintNotFound`.
- `internal/sprint/postgres_repository.go`: `UPDATE sprints SET ... WHERE id = $1 RETURNING ...`; `pgx.ErrNoRows` se traduce a `ErrSprintNotFound` y los demás errores se envuelven con `actualizar el sprint: %w`.
- `internal/sprint/assign.go` y `complete.go`: el mutex del sprint se mantiene durante toda la operación, incluida la escritura en backlog, con orden sprint → backlog. `ListCompletedItems` sigue soltando el mutex antes de listar.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.787s
ok  	seguimiento-medicion-gpc/internal/backlog	1.002s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.553s
ok  	seguimiento-medicion-gpc/internal/postgres	0.660s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.684s
ok  	seguimiento-medicion-gpc/internal/project	0.942s
ok  	seguimiento-medicion-gpc/internal/sprint	1.058s
```

## Refactor

No se hizo: el código nuevo es mínimo y sigue los patrones existentes (`Close` valida y delega como `Create`; `Update` replica el contrato de `Save`/`GetByID` en memoria y PostgreSQL). No hay lógica duplicada ni ramas sin test.

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)
```

```
$ go vet ./...
(sin salida: sin problemas)
```

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.787s
ok  	seguimiento-medicion-gpc/internal/backlog	1.002s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.553s
ok  	seguimiento-medicion-gpc/internal/postgres	0.660s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.684s
ok  	seguimiento-medicion-gpc/internal/project	0.942s
ok  	seguimiento-medicion-gpc/internal/sprint	1.058s
```

Conteo con `go test -v ./internal/sprint`: **182 PASS, 4 SKIP, 0 FAIL**. Los 4 SKIP son los tests de PostgreSQL: `TestPostgresRepository_Contrato` (que incluye los 3 casos de `Update` en PostgreSQL), `TestPostgresRepository_FechasEnUTCConMicrosegundos`, `TestPostgresRepository_SaveConProyectoInexistenteFalla` y `TestPostgresRepository_UpdatePersisteElEstadoYDevuelveFechasUTC`.

### Ejecución de los escenarios BDD

```
$ go test -v -run TestFeatures_HU13CerrarSprint ./internal/sprint
10 scenarios (10 passed)
71 steps (71 passed)
--- PASS: TestFeatures_HU13CerrarSprint (0.02s)
ok  	seguimiento-medicion-gpc/internal/sprint	0.753s
```

Los 10 escenarios pasan con los 71 pasos y sin esquemas del escenario. Con `Strict: true` godog habría fallado si algún paso del `.feature` quedara sin expresión registrada: no hay pasos undefined.

### Cobertura

Valores tomados del log del CI del PR #101 (Linux), job "gofmt y tests":

```
$ go test -race -coverprofile=coverage.out ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.034s	coverage: 39.0% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	1.498s	coverage: 80.9% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.055s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.025s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.023s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	1.233s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	1.403s	coverage: 77.3% of statements
```

```
$ go tool cover -func=coverage.out | tail -n 1
total:									(statements)		74.8%
```

`internal/sprint` sube de 76.9% en `main` (HU-12) a 77.3% con HU-13. El total del perfil es **74.8%**. La medición local en Windows queda fuera de este reporte; el CI (Linux) es la fuente autoritativa. La cobertura con base de PostgreSQL para el nuevo `Update` queda pendiente de la verificación de Buttini o Cardozo.

### go test -race (CA-13.8)

Verificado por Hernán sobre Linux: `go test -race ./...` en verde (los 7 paquetes), con dos pruebas de mutación:

1. Volver `assign.go` y `complete.go` a soltar el mutex del sprint antes de llamar al backlog (comportamiento de HU-11/HU-12): falla el escenario BDD de CA-13.8 en 5 de 5 corridas ("se escribió un ítem después del cierre") y también `TestClose_EnParaleloConAssignItemYCompleteItem`.
2. Sacar el mutex de `Close`: `go test -race` reporta `DATA RACE` y fallan `TestClose_CierresSimultaneosDejanUnSoloCierreExitoso`, `TestClose_EnParaleloConAssignItemYCompleteItem` y el escenario BDD.

En este entorno (Windows, sin compilador C) `-race` no está disponible:

```
$ go test -race ./internal/sprint
go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
```

Los tests de concurrencia (`TestClose_CierresSimultaneosDejanUnSoloCierreExitoso` con 20 intentos, `TestClose_EnParaleloConAssignItemYCompleteItem` con 10 asignaciones y 10 completados, y el escenario BDD de CA-13.8) corren y pasan también sin `-race` en el entorno local.

### PostgreSQL

Verificado por Cardozo Leandro en Windows con Docker, sobre `feature/HU-13-cerrar-sprint` en `8c96b44`, con `TEST_DATABASE_URL` apuntando a `gpc_test` (la base se recreó con `docker compose down -v && docker compose up -d --wait` porque era anterior a HU-11). El resultado está en el comentario de Cardozo en el Pull Request #101:

- `TestPostgresRepository_Contrato`: 10 casos PASS, incluidos los 3 de `Update`.
- `TestPostgresRepository_FechasEnUTCConMicrosegundos`, `TestPostgresRepository_SaveConProyectoInexistenteFalla` y `TestPostgresRepository_UpdatePersisteElEstadoYDevuelveFechasUTC`: PASS.
- 0 SKIP y 0 FAIL; `go test ./...` en verde con la base levantada.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | 8dca7fa | `docs(HU-13): agregar especificación SDD` |
| RED | 6729f3b | `test(HU-13): agregar escenarios BDD y tests (RED)` |
| GREEN | a3cbfae | `feat(HU-13): implementar cerrar sprint (GREEN)` |

No hubo commit de refactor porque no se hizo refactor.
