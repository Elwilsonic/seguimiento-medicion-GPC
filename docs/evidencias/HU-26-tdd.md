# Evidencia TDD — HU-26 · Registrar esfuerzo

> Issue del repositorio: #44

- **Responsable:** García Santiago
- **Rama:** `feature/HU-26-registrar-esfuerzo`
- **Fecha:** 24/10/2026
- **Go:** go1.27.2 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` se verifica en el CI del Pull Request; ver la sección de `go test -race`)
- **Especificación SDD:** [`docs/especificaciones/HU-26-registrar-esfuerzo.md`](../especificaciones/HU-26-registrar-esfuerzo.md), aprobada por Peñalbé Hernán (Agile Enabler) el 24/10/2026
- **Escenarios BDD:** [`features/HU-26-registrar-esfuerzo.feature`](../../features/HU-26-registrar-esfuerzo.feature) (20 escenarios, 139 pasos, ejecutados con godog)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**42 tests nuevos (funciones `Test`) cubriendo 20 escenarios BDD (139 pasos)**, en dos ciclos TDD:

- **Ciclo 1 (memoria, HTTP y web):** 35 tests nuevos.
  - 17 tests unitarios en `internal/effort/register_test.go`.
  - 1 test que ejecuta los escenarios BDD con godog (`internal/effort/register_features_test.go`, runner `TestFeatures_HU26RegistrarEsfuerzo`, con 20 subtests).
  - 5 tests de `WithOpenSprint` en `internal/sprint/with_open_test.go`.
  - 11 tests del endpoint en `internal/httpapi/efforts_test.go` y 1 de la interfaz web en `internal/httpapi/web_test.go` (CA-26.11).
- **Ciclo 2 (PostgreSQL):** 7 tests nuevos, de los cuales 6 son los que fallan en el RED contra la base.
  - 1 test de contrato de `MemoryRepository` en `internal/effort/repository_contract_test.go`.
  - 5 tests de PostgreSQL en `internal/effort/postgres_repository_test.go` (contrato incluido).
  - 1 test de wiring en `cmd/server/main_test.go` (`TestNewRepositories_PostgresUsaElRepositorioDeEsfuerzos`).

La correspondencia no es 1 a 1: CA-26.11 no tiene escenarios godog y se cubre con los tests de `internal/httpapi`; varios tests cubren varios ejemplos de un esquema y varios escenarios comparten tests unitarios.

| Criterio | Escenarios BDD | Tests | Código |
|---|---|---|---|
| CA-26.1 Registro válido, ID correlativo, copia y sin cambios en ítem/sprint | Registrar un esfuerzo válido; Registrar sobre un ítem terminado del sprint (2) | `TestRegister_Exitoso`, `TestRegister_RecortaIntegranteYActividad`, `TestRegister_ItemTerminado`, contratos de `Save` | `register.go`, `memory_repository.go`, `postgres_repository.go` |
| CA-26.2 Horas positivas y finitas, decimales aceptados | Horas fraccionadas; horas 0, negativas, no finitas e infinitas (5) | `TestRegister_HorasFraccionadas`, `TestRegister_HorasInvalidas` | `register.go` (`ErrInvalidHours`) |
| CA-26.3 La historia pertenece al sprint | Ítem sin sprint; ítem de otro sprint (2) | `TestRegister_ItemNoAsignado`, `TestRegister_ItemDeOtroSprint` | `register.go` (`ErrItemNotAssignedToSprint`) |
| CA-26.4 Sprint e ítem existentes | Sprint inexistente; ítem inexistente (2) | `TestRegister_SprintInexistente`, `TestRegister_ItemInexistente`, `TestWithOpenSprint_SprintInexistenteDevuelveErrSprintNotFound`, `TestPostgresRepository_SaveDeItemInexistenteFalla`, `TestPostgresRepository_SaveDeSprintInexistenteFalla` | `register.go`, `with_open.go`, claves foráneas de `efforts` |
| CA-26.5 Solo sprint `ABIERTO` | Sprint cerrado (1) | `TestRegister_SprintCerrado`, `TestWithOpenSprint_SprintCerradoDevuelveErrSprintClosed` | `with_open.go` (`ErrSprintClosed`) |
| CA-26.6 Integrante, actividad y fecha obligatorios | Integrante vacío; actividad vacía; fecha ausente (3) | `TestRegister_DatosObligatorios`, `TestRegister_OrdenDeValidacion` | `register.go` (`ErrEmptyMember`, `ErrEmptyActivity`, `ErrMissingDate`) |
| CA-26.7 Registrar no modifica el ítem ni el sprint | No modificar el ítem (1) | `TestRegister_NoModificaElItem` | `register.go` |
| CA-26.8 El esfuerzo devuelto es una copia | Modificar el esfuerzo devuelto no altera el guardado (1) | `TestRegister_CopiaIndependiente`, contratos de `Save` | `effort.go`, repositorios |
| CA-26.9 Error del repositorio propagado | Error del repositorio al guardar (1) | `TestRegister_ErrorDelRepositorio` | `register.go` |
| CA-26.10 Concurrencia y atomicidad con el cierre | Registros simultáneos; registro y cierre simultáneos (2) | `TestRegister_Concurrente`, `TestRegister_ConcurrenteConCierre`, `TestWithOpenSprint_SerializaConClose` | `manager.go`, `with_open.go` |
| CA-26.11 API y formulario web | — | 11 tests `TestRegistrarEsfuerzo_*`, `TestInterfaz_FormularioDeEsfuerzoUsaStepAny` | `efforts.go`, `server.go`, `index.html`, `app.js` |

Las ejecuciones por criterio suman 20 escenarios (2 + 5 + 2 + 2 + 1 + 3 + 1 + 1 + 1 + 2 + 0).

## Fase RED — ciclo 1 (memoria, HTTP y web)

Escenarios, pasos godog y tests escritos antes del código de producción (commit `c216f54`, que también incluye la SDD y el `.feature`). Siguiendo la acción 1 de la Retrospectiva del Sprint 2, primero se agregaron firmas vacías para que el paquete compile: `Register` devuelve `Effort{}, errNotImplemented`; `MemoryRepository.Save` devuelve `Effort{}, errNotImplemented`; `WithOpenSprint` devuelve `errors.New("no implementado")`; el handler `registerEffort` responde 501. **El RED es el fallo de las aserciones, no de la compilación.**

Archivos con stubs (sin lógica) en el RED:

- `internal/effort/effort.go` (tipo `Effort` y los cuatro errores con su mensaje)
- `internal/effort/repository.go` (interfaz `Repository` con `Save`)
- `internal/effort/memory_repository.go` (`NewMemoryRepository` y `Save` sin guardar)
- `internal/effort/manager.go` (constructor con mutex)
- `internal/effort/register.go` (`Register` sin validar)
- `internal/sprint/with_open.go` (`WithOpenSprint` sin validar estado)
- `internal/httpapi/efforts.go` (handler que responde 501)
- `internal/httpapi/server.go` (ruta registrada y error de esfuerzo)
- `cmd/server/main.go` (wiring con `effort.NewMemoryRepository()`)

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)
```

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
20 scenarios (20 failed)
139 steps (99 passed, 20 failed, 20 skipped)
--- FAIL: TestFeatures_HU26RegistrarEsfuerzo (0.08s)
--- FAIL: TestRegister_Exitoso (0.00s)
--- FAIL: TestRegister_RecortaIntegranteYActividad (0.00s)
--- FAIL: TestRegister_HorasFraccionadas (0.00s)
--- FAIL: TestRegister_ItemTerminado (0.00s)
--- FAIL: TestRegister_HorasInvalidas (0.00s)
--- FAIL: TestRegister_ItemNoAsignado (0.00s)
--- FAIL: TestRegister_ItemDeOtroSprint (0.00s)
--- FAIL: TestRegister_SprintInexistente (0.00s)
--- FAIL: TestRegister_ItemInexistente (0.00s)
--- FAIL: TestRegister_SprintCerrado (0.00s)
--- FAIL: TestRegister_DatosObligatorios (0.00s)
--- FAIL: TestRegister_OrdenDeValidacion (0.00s)
--- FAIL: TestRegister_NoModificaElItem (0.00s)
--- FAIL: TestRegister_CopiaIndependiente (0.00s)
--- FAIL: TestRegister_ErrorDelRepositorio (0.00s)
--- FAIL: TestRegister_Concurrente (0.00s)
--- FAIL: TestRegister_ConcurrenteConCierre (0.00s)
FAIL
FAIL	seguimiento-medicion-gpc/internal/effort	1.367s
--- FAIL: TestRegistrarEsfuerzo_Devuelve201ConElEsfuerzo (0.00s)
--- FAIL: TestRegistrarEsfuerzo_AceptaHorasComoTextoNumerico (0.00s)
--- FAIL: TestRegistrarEsfuerzo_DatosInvalidosDevuelven400 (0.00s)
--- FAIL: TestRegistrarEsfuerzo_CuerpoInvalidoDevuelve400 (0.00s)
--- FAIL: TestRegistrarEsfuerzo_HorasNoNumericasDevuelven400 (0.00s)
--- FAIL: TestRegistrarEsfuerzo_IDNoNumericoDevuelve400 (0.00s)
--- FAIL: TestRegistrarEsfuerzo_InexistenteDevuelve404 (0.00s)
--- FAIL: TestRegistrarEsfuerzo_SprintCerradoDevuelve409 (0.00s)
--- FAIL: TestRegistrarEsfuerzo_ItemSinSprintDevuelve409 (0.00s)
--- FAIL: TestRegistrarEsfuerzo_CuerpoMayorA1MiBDa413 (0.00s)
--- FAIL: TestRegistrarEsfuerzo_ErrorDelRepositorioDevuelve500SinDetalle (0.00s)
--- FAIL: TestInterfaz_FormularioDeEsfuerzoUsaStepAny (0.00s)
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi	1.325s
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
--- FAIL: TestWithOpenSprint_SprintAbiertoEjecutaLaFuncion (0.00s)
--- FAIL: TestWithOpenSprint_SprintCerradoDevuelveErrSprintClosed (0.00s)
--- FAIL: TestWithOpenSprint_SprintInexistenteDevuelveErrSprintNotFound (0.00s)
--- FAIL: TestWithOpenSprint_PropagaElErrorDeLaFuncion (0.00s)
--- FAIL: TestWithOpenSprint_SerializaConClose (0.00s)
FAIL
FAIL	seguimiento-medicion-gpc/internal/sprint	2.161s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

Fragmentos de las aserciones que fallan:

```
    suite.go:640: el registro falló: no implementado
    suite.go:640: error = no implementado, se esperaba las horas trabajadas deben ser un número positivo
    suite.go:640: error = no implementado, se esperaba el ítem no está asignado al sprint
    suite.go:640: error = no implementado, se esperaba no existe un sprint con ese id
    suite.go:640: error = no implementado, se esperaba no existe un ítem del backlog con ese id
    suite.go:640: error = no implementado, se esperaba el sprint está cerrado
    suite.go:640: error = no implementado, se esperaba el integrante es obligatorio
    suite.go:640: error = no implementado, se esperaba la actividad es obligatoria
    suite.go:640: error = no implementado, se esperaba la fecha del esfuerzo es obligatoria
    suite.go:640: esfuerzos guardados = 0, se esperaba 1
    suite.go:640: registro 0 falló: no implementado
    register_test.go:115: Register devolvió un error inesperado: no implementado
```

**35 tests fallan y todos son de HU-26** (18 en `internal/effort`, 12 en `internal/httpapi`, 5 en `internal/sprint`); las demás suites quedan en verde. Los tests nuevos del ciclo 1 no tocan la base (los paquetes `internal/postgres` y `internal/postgrestest` que aparecen en verde son tests previos).

## Fase RED — ciclo 2 (PostgreSQL)

Tests agregados antes del código (commit `c3668a7`), con `PostgresRepository.Save` como stub (`return Effort{}, errNotImplemented`) y `newRepositories` devolviendo `efforts: nil` en la rama PostgreSQL.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	2.138s
ok  	seguimiento-medicion-gpc/internal/backlog	2.836s
ok  	seguimiento-medicion-gpc/internal/effort	2.712s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.131s
ok  	seguimiento-medicion-gpc/internal/postgres	2.014s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.990s
ok  	seguimiento-medicion-gpc/internal/project	2.578s
ok  	seguimiento-medicion-gpc/internal/sprint	2.767s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

Sin `TEST_DATABASE_URL`, los 6 tests nuevos de PostgreSQL hacen SKIP (el contrato de memoria pasa):

```
=== RUN   TestPostgresRepository_Contrato
--- SKIP: TestPostgresRepository_Contrato (0.00s)
=== RUN   TestPostgresRepository_GuardaElEsfuerzo
--- SKIP: TestPostgresRepository_GuardaElEsfuerzo (0.00s)
=== RUN   TestPostgresRepository_NormalizaLaFechaAUTC
--- SKIP: TestPostgresRepository_NormalizaLaFechaAUTC (0.00s)
=== RUN   TestPostgresRepository_SaveDeItemInexistenteFalla
--- SKIP: TestPostgresRepository_SaveDeItemInexistenteFalla (0.00s)
=== RUN   TestPostgresRepository_SaveDeSprintInexistenteFalla
--- SKIP: TestPostgresRepository_SaveDeSprintInexistenteFalla (0.00s)
=== RUN   TestNewRepositories_PostgresUsaElRepositorioDeEsfuerzos
--- SKIP: TestNewRepositories_PostgresUsaElRepositorioDeEsfuerzos (0.00s)
```

### Corridas contra la base (Hernán, Docker + `TEST_DATABASE_URL`)

Peñalbé Hernán, 24/10/2026, Linux, go1.27.1, PostgreSQL 16 (el `docker-compose` usa 17; para estos tests no cambia nada). Antes de cada corrida, `gpc_test` se recreó desde `db/init/01-schema.sql` de ese commit. Se muestran las líneas de resultado y de fallo; se omite la salida de godog.

#### 1. RED (`c3668a7`)

La corrida con la base levantada se hizo sobre `c3668a7`: **fallan solo los 6 tests nuevos de PostgreSQL por aserción y el resto pasa**. Los 6 son los 5 de `internal/effort/postgres_repository_test.go` y `TestNewRepositories_PostgresUsaElRepositorioDeEsfuerzos` de `cmd/server`; `TestMemoryRepository_Contrato` pasa porque la implementación en memoria ya existía.

```
$ TEST_DATABASE_URL="postgres://gpc:***@127.0.0.1:5432/gpc_test?sslmode=disable" go test -count=1 ./...
--- FAIL: TestNewRepositories_PostgresUsaElRepositorioDeEsfuerzos (0.04s)
    main_test.go:177: efforts es <nil>, se esperaba *effort.PostgresRepository
FAIL
FAIL	seguimiento-medicion-gpc/cmd/server	0.116s
ok  	seguimiento-medicion-gpc/internal/backlog	1.056s
--- FAIL: TestPostgresRepository_Contrato (0.19s)
        repository_contract_test.go:27: Save devolvió un error inesperado: no implementado
        repository_contract_test.go:42: Save devolvió un error inesperado: no implementado
        repository_contract_test.go:52: Save devolvió un error inesperado: no implementado
--- FAIL: TestPostgresRepository_GuardaElEsfuerzo (0.49s)
    postgres_repository_test.go:75: Save devolvió un error inesperado: no implementado
--- FAIL: TestPostgresRepository_NormalizaLaFechaAUTC (0.08s)
    postgres_repository_test.go:99: Save devolvió un error inesperado: no implementado
--- FAIL: TestPostgresRepository_SaveDeItemInexistenteFalla (0.10s)
    postgres_repository_test.go:119: err = no implementado, se esperaba una violación de clave foránea (23503)
--- FAIL: TestPostgresRepository_SaveDeSprintInexistenteFalla (0.09s)
    postgres_repository_test.go:133: err = no implementado, se esperaba una violación de clave foránea (23503)
FAIL
FAIL	seguimiento-medicion-gpc/internal/effort	1.045s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.138s
ok  	seguimiento-medicion-gpc/internal/postgres	0.016s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.019s
ok  	seguimiento-medicion-gpc/internal/project	0.252s
ok  	seguimiento-medicion-gpc/internal/sprint	0.465s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

#### 2. GREEN (`dd5d112`, con la falla)

El GREEN real pasó todos los tests menos uno: `TestPostgresRepository_GuardaElEsfuerzo` falló con `zona de la fecha = Local, se esperaba UTC` en `postgres_repository_test.go:86`. El error era del test, no del repositorio (ver "Desviación del ciclo TDD").

```
$ TEST_DATABASE_URL="postgres://gpc:***@127.0.0.1:5432/gpc_test?sslmode=disable" go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.171s
ok  	seguimiento-medicion-gpc/internal/backlog	1.537s
--- FAIL: TestPostgresRepository_GuardaElEsfuerzo (0.52s)
    postgres_repository_test.go:86: zona de la fecha = Local, se esperaba UTC
FAIL
FAIL	seguimiento-medicion-gpc/internal/effort	1.504s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.085s
ok  	seguimiento-medicion-gpc/internal/postgres	0.011s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.011s
ok  	seguimiento-medicion-gpc/internal/project	0.542s
ok  	seguimiento-medicion-gpc/internal/sprint	0.485s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

#### 3. GREEN corregido

Después de sacar la aserción de zona horaria del test, la corrida contra la base pasa completa. Hernán la repite sobre el commit `bed08f5` y confirma el hash para el título.

```
$ TEST_DATABASE_URL="postgres://gpc:***@127.0.0.1:5432/gpc_test?sslmode=disable" go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.174s
ok  	seguimiento-medicion-gpc/internal/backlog	1.175s
ok  	seguimiento-medicion-gpc/internal/effort	1.062s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.110s
ok  	seguimiento-medicion-gpc/internal/postgres	0.007s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.009s
ok  	seguimiento-medicion-gpc/internal/project	0.261s
ok  	seguimiento-medicion-gpc/internal/sprint	0.502s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

#### 4. `go test -race` contra la base

Mismo código, con `-race`.

```
$ TEST_DATABASE_URL="postgres://gpc:***@127.0.0.1:5432/gpc_test?sslmode=disable" go test -count=1 -race ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.292s
ok  	seguimiento-medicion-gpc/internal/backlog	2.594s
ok  	seguimiento-medicion-gpc/internal/effort	2.013s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.503s
ok  	seguimiento-medicion-gpc/internal/postgres	1.035s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.025s
ok  	seguimiento-medicion-gpc/internal/project	1.649s
ok  	seguimiento-medicion-gpc/internal/sprint	2.093s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

En las corridas 3 y 4 ningún test quedó en SKIP (`go test -v ./... | grep -c -- '--- SKIP'` → 0).

## Fase GREEN — ciclo 1 (memoria, HTTP y web)

Código mínimo de producción (commit `1dab476`), sin tocar tests:

- `internal/effort/effort.go`: tipo `Effort` y los cuatro errores.
- `internal/effort/memory_repository.go`: `Save` asigna el ID correlativo global desde 1 y guarda una copia.
- `internal/effort/register.go`: valida integrante, actividad, fecha y horas (en ese orden); dentro de `WithOpenSprint` consulta el ítem, valida la pertenencia y guarda.
- `internal/sprint/with_open.go`: `WithOpenSprint` toma el mutex del sprint, valida existencia y estado `ABIERTO` y ejecuta `fn`.
- `internal/httpapi/efforts.go` y `server.go`: handler, ruta y errores 400 en `statusFor`.
- `internal/web/index.html` y `app.js`: formulario "Registrar esfuerzo" con Horas `type="number" step="any"` y llamada a `/esfuerzos`.
- `cmd/server/main.go`: arma el `effort.Manager` (en este ciclo, con memoria).

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	2.264s
ok  	seguimiento-medicion-gpc/internal/backlog	3.505s
ok  	seguimiento-medicion-gpc/internal/effort	2.297s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.296s
ok  	seguimiento-medicion-gpc/internal/postgres	2.021s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.841s
ok  	seguimiento-medicion-gpc/internal/project	2.812s
ok  	seguimiento-medicion-gpc/internal/sprint	3.349s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

```
$ go test -count=1 -v -run TestFeatures_HU26RegistrarEsfuerzo ./internal/effort/
20 scenarios (20 passed)
139 steps (139 passed)
```

## Fase GREEN — ciclo 2 (PostgreSQL)

Código mínimo de producción (commit `dd5d112`), sin tocar tests:

- `internal/effort/postgres_repository.go`: `Save` real con `INSERT … RETURNING` y `scanEffort`, que normaliza la fecha a UTC.
- `db/init/01-schema.sql`: tabla `efforts` (`id` identity, `member`, `effort_date`, `activity`, `item_id` y `sprint_id` con claves foráneas, `hours`) e índices `efforts_sprint_id_idx` y `efforts_item_id_idx`.
- `internal/postgrestest/postgrestest.go`: `efforts` incluida en el `TRUNCATE`.
- `cmd/server/repositories.go`: la rama PostgreSQL usa `effort.NewPostgresRepository(pool)`.
- `cmd/server/main.go`: **ahora usa `repos.efforts`**. En el primer GREEN (y en el RED) el `effort.Manager` se armaba con `effort.NewMemoryRepository()` directamente, así que con `STORAGE=postgres` el esfuerzo se habría guardado en memoria; este ciclo lo conecta al repositorio configurado.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	2.326s
ok  	seguimiento-medicion-gpc/internal/backlog	2.580s
ok  	seguimiento-medicion-gpc/internal/effort	2.675s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.824s
ok  	seguimiento-medicion-gpc/internal/postgres	0.616s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.928s
ok  	seguimiento-medicion-gpc/internal/project	2.644s
ok  	seguimiento-medicion-gpc/internal/sprint	2.607s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

La corrida GREEN contra la base la hizo Hernán y dejó al descubierto un error en un test del RED; se corrigió en un commit posterior (ver "Desviación del ciclo TDD" y las corridas contra la base).

## Desviación del ciclo TDD

`TestPostgresRepository_GuardaElEsfuerzo` se escribió en el RED (commit `c3668a7`) leyendo la fila con `pool.QueryRow` directamente, sin pasar por el repositorio, y además de comparar los campos con `contractAssertEffort` verificaba que `got.Date.Location()` fuera `time.UTC`. En el RED esa aserción quedaba tapada por el stub: `contractSave` fallaba antes con `errNotImplemented` y el test moría en el `t.Fatalf` de la lectura, así que nunca llegaba a la comprobación de la zona. Con el GREEN (`dd5d112`) el guardado funcionó, la lectura devolvió una fecha en la zona `Local` de pgx y el test falló, aunque `Save` devolvía la fecha normalizada a UTC.

El error era del test: la normalización a UTC se verifica sobre el esfuerzo que devuelve el repositorio en `TestPostgresRepository_NormalizaLaFechaAUTC`, no sobre una lectura directa con `QueryRow`, que no pasa por `scanEffort`. `contractAssertEffort` ya compara las fechas con `Equal`, que compara el instante y no la zona. Se sacaron las tres líneas del `if got.Date.Location() != time.UTC { ... }` y no se tocó el código de producción.

Commit de la corrección: `bed08f5`.

## Refactor

No se hizo. El paquete `internal/effort` sigue los patrones de `internal/sprint` y `internal/backlog` (manager con mutex, repositorios detrás de una interfaz, validaciones en el manager), y el handler reutiliza `parseDate`, `decodeJSON`, `writeJSON` y `statusFor`. No hay lógica duplicada ni ramas sin test que justifiquen un refactor separado.

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
ok  	seguimiento-medicion-gpc/cmd/server	2.326s
ok  	seguimiento-medicion-gpc/internal/backlog	2.580s
ok  	seguimiento-medicion-gpc/internal/effort	2.675s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.824s
ok  	seguimiento-medicion-gpc/internal/postgres	0.616s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.928s
ok  	seguimiento-medicion-gpc/internal/project	2.644s
ok  	seguimiento-medicion-gpc/internal/sprint	2.607s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

### Cobertura

```
$ go test -count=1 -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.996s	coverage: 53.1% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	2.927s	coverage: 79.9% of statements
ok  	seguimiento-medicion-gpc/internal/effort	2.677s	coverage: 74.4% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	2.121s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	2.062s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	2.113s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	2.534s	coverage: 64.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	2.754s	coverage: 75.6% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

```
$ go test -count=1 -coverprofile=cover.out ./...
$ go tool cover -func=cover.out
internal/effort/manager.go:15:			NewManager		100.0%
internal/effort/memory_repository.go:9:		NewMemoryRepository	100.0%
internal/effort/memory_repository.go:16:	Save			100.0%
internal/effort/postgres_repository.go:20:	NewPostgresRepository	0.0%
internal/effort/postgres_repository.go:30:	Save			0.0%
internal/effort/postgres_repository.go:44:	scanEffort		0.0%
internal/effort/register.go:15:			Register		100.0%
cmd/server/main.go:43:				newServer		100.0%
cmd/server/main.go:56:				address			100.0%
cmd/server/repositories.go:26:			newRepositories		83.3%
total:						(statements)		80.4%
```

`internal/effort/postgres_repository.go` queda en 0.0% porque sus tests hacen SKIP sin `TEST_DATABASE_URL`; con la base levantada se mide en la corrida de PostgreSQL. `main()` queda en 0.0% porque bloquea en `ListenAndServe` y no se prueba con un test unitario. `newRepositories` queda en 83.3% por la rama de PostgreSQL que necesita la base.

Con la base levantada, la corrida con `-coverpkg=./...` midió **94.8%** (Hernán), por encima del 92.5% de la última medición con base (HU-11), e incluye `postgres_repository.go` de esfuerzos.

Comparación con `main` (`ef2c361`) medida con el mismo Go 1.27.2 en una copia temporal (`git worktree`):

```
$ go test -count=1 -cover ./...      (sobre ef2c361)
ok  	seguimiento-medicion-gpc/cmd/server	54.8% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	79.9% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	64.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	74.0% of statements
$ go tool cover -func=cover-main.out | tail -n 1
total:						(statements)		80.0%
```

`internal/effort` entra en 74.4% (nuevo), `internal/sprint` sube de 74.0% a 75.6% con `WithOpenSprint`, `internal/httpapi` sigue en 100.0% y el total sube de 80.0% a **80.4%**. `cmd/server` baja de 54.8% a 53.1% porque `main()` y `newRepositories` suman código nuevo que no se ejecuta sin base. La diferencia con las cifras de HU-20 (83.3% de `internal/backlog`, 80.6% del total) se debe al cambio de go1.27.1 a go1.27.2, que alteró el conteo de sentencias; la comparación válida es contra `main` con el mismo Go. La cifra autoritativa sigue siendo el CI (Linux) del Pull Request, pendiente.

### go test -race (CA-26.10)

En este entorno (Windows, `CGO_ENABLED=0`, sin compilador C) `-race` no está disponible. Los tests de concurrencia (`TestRegister_Concurrente`, `TestRegister_ConcurrenteConCierre` y `TestWithOpenSprint_SerializaConClose`) pasan sin `-race` localmente.

Hernán corrió `go test -race` contra la base (corrida 4, arriba). El resultado del CI del Pull Request queda pendiente.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED ciclo 1 (incluye SDD y BDD) | c216f54 | `test(HU-26): agregar tests (RED)` |
| GREEN ciclo 1 | 1dab476 | `feat(HU-26): implementar registro de esfuerzo (GREEN)` |
| RED ciclo 2 (PostgreSQL) | c3668a7 | `test(HU-26): agregar tests del repositorio PostgreSQL (RED)` |
| GREEN ciclo 2 (PostgreSQL) | dd5d112 | `feat(HU-26): guardar el esfuerzo en PostgreSQL (GREEN)` |
| Corrección de test | bed08f5 | `test(HU-26): corregir la aserción de zona horaria del test de PostgreSQL` |

No hubo commit de refactor porque no se hizo refactor.

HU-26: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
