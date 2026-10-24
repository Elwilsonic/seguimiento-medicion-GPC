# Evidencia TDD — HU-21 Abrir una sesión de Planning Poker

- **Historia:** HU-21 — Abrir una sesión de Planning Poker
- **Issue:** [#39](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/39)
- **Rama:** `feature/HU-21-abrir-sesion-planning-poker`
- **Fecha:** 24/10/2026
- **Entorno local:** Windows/amd64, go1.27.1, sin cgo y sin `TEST_DATABASE_URL`.

La especificación SDD se aprobó antes de escribir los tests. El ciclo se ejecutó en tres commits: SDD (`337f651`), RED (`02c6b82`) y GREEN (`287247b`). No hubo commit de refactor.

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test |
|---|---|
| CA-21.1 — sesión asociada a un ítem existente, ID correlativo, `ABIERTA`, ronda 1 y sin votos | Escenarios “Abrir una sesión para un ítem pendiente” y “Abrir una sesión para un ítem en curso con story points”; `TestOpenSession_CreaSesionAbiertaSinVotos`, `TestOpenSession_AsignaIDsCorrelativosGlobales`, `TestOpenSession_ItemEnCursoConStoryPoints`, `TestAbrirSesionPoker_Devuelve201ConLaSesionAbierta` |
| CA-21.2 — una sola sesión abierta por ítem; dos ítems distintos pueden tener una cada uno | Escenarios “Abrir sesiones para dos ítems distintos” y “No abrir dos sesiones para el mismo ítem”; `TestOpenSession_SegundaSesionDevuelveErrSessionAlreadyOpen`, `TestOpenSession_ConcurrenteSoloUnaSesionAbierta`, `TestAbrirSesionPoker_SesionAbiertaDevuelve409` |
| CA-21.3 — la sesión empieza sin votos y se devuelve una copia | `TestOpenSession_CreaSesionAbiertaSinVotos`, `TestOpenSession_DevuelveCopiaDeLosVotos`, `TestMemoryRepository_Contrato`, `TestPostgresRepository_Contrato` |
| CA-21.4 — ítem inexistente o `TERMINADO`; no se crea la sesión | Escenarios “No abrir una sesión para un ítem inexistente” y “No abrir una sesión para un ítem terminado”; `TestOpenSession_ItemInexistenteDevuelveErrItemNotFound`, `TestOpenSession_ItemTerminadoDevuelveErrItemFinished`, `TestAbrirSesionPoker_InexistenteDevuelve404`, `TestAbrirSesionPoker_ItemTerminadoDevuelve409` |
| CA-21.5 — cartas de la escala de HU-20, como lista nueva | Escenario “Las cartas de la sesión son la escala de HU-20”; `TestCards_DevuelveLaEscalaYUnaCopia` |
| CA-21.6 — concurrencia, error del repositorio y `-race` | `TestOpenSession_ConcurrenteSoloUnaSesionAbierta`, `TestOpenSession_ErrorDelRepositorioSEPropaga`, `TestAbrirSesionPoker_ErrorDelRepositorioDevuelve500SinDetalle`; `-race` se valida en CI |
| CA-21.7 — API y interfaz web | `TestAbrirSesionPoker_*`, `TestInterfaz_BacklogPermiteAbrirPlanningPoker` |
| CA-21.8 — mismo contrato en memoria y PostgreSQL; integración omitida sin base | `TestMemoryRepository_Contrato`, `TestPostgresRepository_Contrato`, `TestPostgresRepository_SaveConItemInexistenteFalla`, `TestNewRepositories_PostgresUsaElRepositorioDePoker` |

## Alcance de las pruebas

- **7 escenarios BDD** en `features/HU-21-abrir-sesion-planning-poker.feature` (37 pasos), automatizados con godog.
- **23 funciones `Test` nuevas**, incluida la que ejecuta los escenarios BDD.
- **16 subtests explícitos**: 2 de `TestOpenSession_ErrorDelRepositorioSEPropaga`, 7 del contrato en memoria y 7 del contrato PostgreSQL.
- En la corrida local sin base del RED, **18 funciones fallan, 3 se omiten y 2 pasan**. Las omitidas son las que necesitan `TEST_DATABASE_URL`; las dos que pasan son `TestAbrirSesionPoker_IDNoNumericoDevuelve400` y `TestAbrirSesionPoker_ErrorDelRepositorioDevuelve500SinDetalle`.
- En la corrida con base del RED, **20 funciones fallan y 3 pasan**: fallan 14 de `internal/poker` (incluidos los 2 de PostgreSQL) y 6 de `internal/httpapi`; pasan los 2 tests HTTP anteriores y `TestNewRepositories_PostgresUsaElRepositorioDePoker`.
- De los 16 subtests, **9 fallan** (los 2 del error de repositorio y los 7 del contrato en memoria) y **7 se omiten** (el contrato PostgreSQL) sin base.

## Fase RED

Commit `02c6b82`. La SDD, los escenarios BDD y los tests se escribieron antes del código. El paquete `internal/poker` quedó con las firmas y stubs que devolvían `funcionalidad de Planning Poker no implementada`; el handler y la ruta ya estaban conectados al manager, pero el 405, el mapeo de errores, la interfaz web, el esquema PostgreSQL y los repositorios reales todavía no existían.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	6.163s
ok  	seguimiento-medicion-gpc/internal/backlog	9.829s
ok  	seguimiento-medicion-gpc/internal/effort	6.850s
--- FAIL: TestAbrirSesionPoker_Devuelve201ConLaSesionAbierta (0.00s)
    poker_test.go:39: código = 500, se esperaba 201; cuerpo: {"error":"error interno del servidor"}
--- FAIL: TestAbrirSesionPoker_InexistenteDevuelve404 (0.00s)
    poker_test.go:71: código = 500, se esperaba 404; cuerpo: {"error":"error interno del servidor"}
--- FAIL: TestAbrirSesionPoker_ItemTerminadoDevuelve409 (0.00s)
    poker_test.go:81: código = 500, se esperaba 409; cuerpo: {"error":"error interno del servidor"}
--- FAIL: TestAbrirSesionPoker_SesionAbiertaDevuelve409 (0.00s)
    poker_test.go:86: código = 500, se esperaba 201; cuerpo: {"error":"error interno del servidor"}
--- FAIL: TestAbrirSesionPoker_MetodoNoPermitidoDevuelve405 (0.00s)
    poker_test.go:98: código = 404, se esperaba 405
--- FAIL: TestInterfaz_BacklogPermiteAbrirPlanningPoker (0.00s)
    web_test.go:142: la página no contiene "Planning Poker"
    web_test.go:142: la página no contiene "id=\"datos-poker\""
    web_test.go:142: la página no contiene "id=\"cartas-poker\""
    web_test.go:148: app.js no contiene "\"/poker-sessions\""
    web_test.go:148: app.js no contiene "\"Abrir Poker\""
    web_test.go:148: app.js no contiene "s.cartas"
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi	4.216s
```

La salida de godog confirmó que los 7 escenarios fallaban por la funcionalidad no implementada:

```
7 scenarios (7 failed)
37 steps (20 passed, 7 failed, 10 skipped)
--- FAIL: TestFeatures_HU21AbrirSesionPlanningPoker (0.06s)
    suite.go:640: la apertura falló: funcionalidad de Planning Poker no implementada
    suite.go:640: error = funcionalidad de Planning Poker no implementada, se esperaba backlog.ErrItemNotFound
    suite.go:640: error = funcionalidad de Planning Poker no implementada, se esperaba poker.ErrItemFinished
    suite.go:640: cartas = [], se esperaba [1 2 3 5 8 13 21]
```

Los tests unitarios de `internal/poker` fallaron con el mismo stub:

```
--- FAIL: TestOpenSession_CreaSesionAbiertaSinVotos (0.00s)
    create_test.go:93: OpenSession(1) devolvió error: funcionalidad de Planning Poker no implementada
--- FAIL: TestOpenSession_AsignaIDsCorrelativosGlobales (0.00s)
    create_test.go:117: abrir la primera sesión: funcionalidad de Planning Poker no implementada
--- FAIL: TestOpenSession_ItemInexistenteDevuelveErrItemNotFound (0.00s)
    create_test.go:155: error = funcionalidad de Planning Poker no implementada, se esperaba backlog.ErrItemNotFound
--- FAIL: TestOpenSession_ConcurrenteSoloUnaSesionAbierta (0.04s)
    create_test.go:242: sesiones abiertas guardadas = 0, se esperaba 1
--- FAIL: TestCards_DevuelveLaEscalaYUnaCopia (0.00s)
    create_test.go:253: cartas = [], se esperaba [1 2 3 5 8 13 21]
--- FAIL: TestMemoryRepository_Contrato (0.00s)
    --- FAIL: TestMemoryRepository_Contrato/Save_asigna_IDs_correlativos_y_conserva_los_campos
        repository_contract_test.go:21: Save de la primera sesión: funcionalidad de Planning Poker no implementada
FAIL
FAIL	seguimiento-medicion-gpc/internal/poker	7.833s
```

El fallo de los tests HTTP correspondía a la funcionalidad nueva; los paquetes existentes siguieron en verde. `cmd/server` pasó porque el test de PostgreSQL se omite sin base.

## Fase GREEN

Commit `287247b`. Se implementó el código mínimo:

- `internal/poker`: `Session`, `Vote`, `StatusOpen`, `ErrSessionAlreadyOpen`, `ErrItemFinished`, `Repository`, `Manager` con mutex, `OpenSession`, `Cards` y los repositorios en memoria y PostgreSQL.
- `internal/httpapi`: se mantuvo `POST /api/items/{id}/poker-sessions` con respuesta 201 y se agregaron el 405 explícito para `GET` y el mapeo de errores 400/404/409/500.
- `internal/web`: sección “Planning Poker”, botón “Abrir Poker” y `mostrarPoker`.
- `cmd/server`: `NewHandler` recibe el `poker.Manager` y el repositorio se elige según `STORAGE`.
- `db/init/01-schema.sql`: tabla `poker_sessions` e índice `poker_sessions_item_id_idx`.
- `internal/postgrestest`: `poker_sessions` incluida en el `TRUNCATE`.

El test de 405 se resolvió con dos registros de ruta: `GET /api/items/{id}/poker-sessions` responde 405 y `POST` atiende la apertura. Un único patrón para todos los métodos chocaba con `GET /` y provocaba un panic en `ServeMux`.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	4.534s
ok  	seguimiento-medicion-gpc/internal/backlog	5.193s
ok  	seguimiento-medicion-gpc/internal/effort	3.996s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.615s
ok  	seguimiento-medicion-gpc/internal/poker	4.130s
ok  	seguimiento-medicion-gpc/internal/postgres	2.705s
ok  	seguimiento-medicion-gpc/internal/postgrestest	2.540s
ok  	seguimiento-medicion-gpc/internal/project	4.038s
ok  	seguimiento-medicion-gpc/internal/sprint	3.533s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

```
$ go test -count=1 -v -run TestFeatures_HU21AbrirSesionPlanningPoker ./internal/poker/
7 scenarios (7 passed)
37 steps (37 passed)
25.2752ms
--- PASS: TestFeatures_HU21AbrirSesionPlanningPoker (0.03s)
PASS
ok  	seguimiento-medicion-gpc/internal/poker	5.140s
```

No se modificaron los tests entre el RED y el GREEN.

## Desviación del ciclo TDD

`TestNewRepositories_PostgresUsaElRepositorioDePoker` ya pasa en el RED con base porque `02c6b82` incluía `poker.NewPostgresRepository(pool)` en `cmd/server/repositories.go`, aunque `PostgresRepository.Save` y `OpenByItem` todavía devolvían el stub. No es comportamiento preexistente de otra historia: es el wiring de HU-21 adelantado en el commit RED. Sin base, ese test se omite.

En la corrida local sin base, los 2 tests que pasaban en el RED eran `TestAbrirSesionPoker_IDNoNumericoDevuelve400` y `TestAbrirSesionPoker_ErrorDelRepositorioDevuelve500SinDetalle`; con base, el tercero que pasa es `TestNewRepositories_PostgresUsaElRepositorioDePoker`. Los tests no se modificaron para que el RED pasara.

## Refactor

No se hizo. El paquete `internal/poker` sigue el patrón de `internal/backlog` y `internal/sprint` (manager con mutex, repositorios detrás de una interfaz y validaciones en el manager), y el handler reutiliza `pathID`, `writeError` y `writeJSON`. No había duplicación ni una mejora estructural que justificara un commit separado.

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
ok  	seguimiento-medicion-gpc/cmd/server	4.534s
ok  	seguimiento-medicion-gpc/internal/backlog	5.193s
ok  	seguimiento-medicion-gpc/internal/effort	3.996s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.615s
ok  	seguimiento-medicion-gpc/internal/poker	4.130s
ok  	seguimiento-medicion-gpc/internal/postgres	2.705s
ok  	seguimiento-medicion-gpc/internal/postgrestest	2.540s
ok  	seguimiento-medicion-gpc/internal/project	4.038s
ok  	seguimiento-medicion-gpc/internal/sprint	3.533s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

### Cobertura

Medición local sin base, con `go test -count=1 -cover ./...`:

```
ok  	seguimiento-medicion-gpc/cmd/server	4.692s	coverage: 33.3% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	4.830s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/effort	4.142s	coverage: 76.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.924s	coverage: 98.6% of statements
ok  	seguimiento-medicion-gpc/internal/poker	4.162s	coverage: 54.4% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	3.443s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	3.346s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	3.384s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	3.533s	coverage: 79.1% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

```
$ go test -count=1 -coverprofile=cover.out ./...
$ go tool cover -func=cover.out
internal/poker/create.go:13:			OpenSession		100.0%
internal/poker/create.go:46:			Cards			100.0%
internal/poker/manager.go:22:			NewManager		100.0%
internal/poker/memory_repository.go:11:		NewMemoryRepository	100.0%
internal/poker/memory_repository.go:17:		Save			100.0%
internal/poker/memory_repository.go:27:		OpenByItem		100.0%
internal/poker/postgres_repository.go:21:	NewPostgresRepository	0.0%
internal/poker/postgres_repository.go:31:	Save			0.0%
internal/poker/postgres_repository.go:53:	OpenByItem		0.0%
internal/poker/postgres_repository.go:69:	scanSession		0.0%
internal/poker/session.go:38:			cloneSession		100.0%
internal/httpapi/poker.go:25:			methodNotAllowed	100.0%
internal/httpapi/poker.go:32:			openPokerSession	75.0%
cmd/server/main.go:21:				main			0.0%
cmd/server/main.go:45:				newServer		100.0%
cmd/server/main.go:58:				address			100.0%
cmd/server/repositories.go:28:			newRepositories		83.3%
cmd/server/storage.go:32:			storageConfig		100.0%
total:									(statements)		79.1%
```

`internal/poker` queda en 54.4% sin base porque `postgres_repository.go` no se ejecuta sin `TEST_DATABASE_URL`; con la base levantada se miden sus tests de integración. `openPokerSession` queda en 75.0% por la rama defensiva de método no permitido, que la ruta ya resuelve con `methodNotAllowed`. `main()` queda en 0.0% porque bloquea en `ListenAndServe` y `newRepositories` en 83.3% por la rama PostgreSQL.

Comparación con `main` (`194ebe4`) medida con el mismo go1.27.1 en una copia temporal: el total sin base pasa de **80.9%** a **79.1%**. `internal/httpapi` baja de 100.0% a 98.6% por el handler nuevo, `cmd/server` de 35.4% a 33.3% por el wiring de poker y el paquete nuevo `internal/poker` entra en 54.4%. `internal/backlog`, `internal/effort`, `internal/postgres`, `internal/postgrestest`, `internal/project` e `internal/sprint` conservan sus porcentajes.

### Corrida con base (Hernán)

Hernán corrió el ciclo con PostgreSQL 16 y `TEST_DATABASE_URL`:

- RED (`02c6b82`): fallan 14 tests de `internal/poker` (incluidos los 2 de PostgreSQL) y 6 de `internal/httpapi`; pasan los 2 tests HTTP que ya tenían el comportamiento esperado y `TestNewRepositories_PostgresUsaElRepositorioDePoker`.
- GREEN (`053c5da`): pasa todo, sin ningún SKIP, y `go test -race ./...` también pasa.
- Cobertura con base: `internal/poker` **94.7%** y **94.2%** de total con `-coverpkg=./...`.

### go test -race

En este entorno local (Windows, `CGO_ENABLED=0`) `-race` no está disponible. Hernán verificó `go test -race ./...` con PostgreSQL 16 sobre `053c5da`, sin SKIP y sin condiciones de carrera. El test de concurrencia `TestOpenSession_ConcurrenteSoloUnaSesionAbierta` también pasa sin `-race` localmente; CA-21.6 queda verificado por esa corrida y por el CI del Pull Request.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | 337f651 | `docs(HU-21): agregar especificación de apertura de sesión` |
| RED | 02c6b82 | `test(HU-21): agregar tests (RED)` |
| GREEN | 287247b | `feat(HU-21): implementar apertura de sesión de Planning Poker (GREEN)` |
| Refactor | — | No aplica |

Los tres commits referencian el Issue #39 con `Refs #39`.

HU-21: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
