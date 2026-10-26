# Evidencia TDD · HU-23

> Issue del repositorio: #41

## Historia

Como Agile Enabler, quiero revelar los votos de todos a la vez, para comparar las estimaciones.

## Resultado

HU-23 quedó implementada en la rama `feature/HU-23-revelar-estimaciones`. `Manager.Reveal` pasa una sesión `ABIERTA` con votos a `REVELADA` bajo el mismo mutex que `Vote` y devuelve todos los votos como copia, en el orden en que se guardaron. La API lo expone en `POST /api/poker-sessions/{id}/reveal` y la interfaz web con el botón "Revelar votos", que muestra la lista de votos y oculta el formulario de voto.

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test |
|---|---|
| CA-23.1 — revelar muestra los votos de todos y deja la sesión revelada | “Revelar una sesión con votos muestra todos los votos”; “Revelar una sesión con un único voto”; “Revelar muestra el voto reemplazado de un integrante”; `TestReveal_SesionConVotosDevuelveTodosLosVotosYEstadoRevelada`; `TestReveal_SesionConUnSoloVoto` |
| CA-23.2 — no se puede revelar una sesión sin votos | “Rechazar revelar una sesión sin votos”; `TestReveal_SesionSinVotosDevuelveErrNoVotes` |
| CA-23.3 — después de revelar no se puede votar | “No se puede votar después de revelar”; `TestVote_DespuesDeRevelarDevuelveErrSessionNotOpen` |
| CA-23.4 — revelar una sesión inexistente o no abierta da error | “Rechazar revelar una sesión ya revelada”; “Rechazar revelar una sesión inexistente”; `TestReveal_SesionInexistenteDevuelveErrSessionNotFound`; `TestReveal_SesionYaReveladaDevuelveErrSessionNotOpen` |
| CA-23.5 — si falla el repositorio, se informa y la sesión no cambia | `TestReveal_ErrorAlLeerPropagaError`; `TestReveal_ErrorAlGuardarPropagaErrorYSesionNoCambia` |
| CA-23.6 — revelar solo cambia el estado y devuelve una copia | “Revelar conserva la ronda, el ítem y el orden de los votos”; `TestReveal_ConservaRondaItemYOrdenDeLosVotos`; `TestReveal_NoModificaElItemNiSusStoryPoints`; `TestReveal_DevuelveCopiaDeLosVotos` |
| CA-23.7 — votar mientras se revela no pierde votos aceptados | “Votar mientras se revela no pierde votos aceptados”; `TestReveal_ConcurrenteConVotosNoPierdeVotosAceptados` |
| CA-23.8 — revelar desde la API y la interfaz web | `TestRevelarPoker_Devuelve200ConTodosLosVotos`; `TestRevelarPoker_IDNoNumericoDevuelve400`; `TestRevelarPoker_InexistenteDevuelve404`; `TestRevelarPoker_SesionNoAbiertaDevuelve409`; `TestRevelarPoker_SinVotosDevuelve409`; `TestRevelarPoker_ErrorDelRepositorioDevuelve500SinDetalle`; `TestRevelarPoker_MetodoNoPermitidoDevuelve405`; `TestRevelarPoker_IgnoraElCuerpo`; caso “revelar Planning Poker” de `TestRutasConID_IDNoNumericoDevuelve400`; `TestInterfaz_BotonRevelarVotos` |
| CA-23.9 — mismo comportamiento en memoria y PostgreSQL | Subtest de contrato “Update guarda el estado REVELADA y GetByID lo devuelve con sus votos”, ejecutado por `TestMemoryRepository_Contrato` y `TestPostgresRepository_Contrato`; PostgreSQL se omite sin `TEST_DATABASE_URL` |

La validación contiene **9 escenarios BDD** y **23 casos de prueba específicos**: 12 unitarios de `internal/poker/reveal_test.go`, 1 subtest de contrato de repositorio, 8 tests HTTP, 1 caso nuevo en `TestRutasConID_IDNoNumericoDevuelve400` y 1 test de interfaz web. La correspondencia no es 1 a 1: CA-23.5 no tiene escenario BDD y CA-23.8 se prueba solo con tests de `internal/httpapi`.

Conteos con `grep -c` sobre los archivos de la rama:

```text
$ grep -c "^  Escenario:" features/HU-23-revelar-estimaciones.feature
9
$ grep -c "^func Test" internal/poker/reveal_test.go
12
$ grep -c "^func TestRevelarPoker_" internal/httpapi/poker_test.go
8
$ grep -c '"revelar Planning Poker"' internal/httpapi/server_test.go
1
$ grep -c "^func TestInterfaz_BotonRevelarVotos" internal/httpapi/web_test.go
1
$ grep -c 'Update guarda el estado REVELADA' internal/poker/repository_contract_test.go
1
```

## BDD

Archivo: `features/HU-23-revelar-estimaciones.feature`, con el runner `TestFeatures_HU23RevelarEstimaciones` en `internal/poker/reveal_features_test.go`. Todos los pasos llaman al `Manager` de `internal/poker`, no al handler HTTP.

Resultado de `go test -v -count=1 -run TestFeatures_HU23 ./internal/poker/` sobre el GREEN:

```text
9 scenarios (9 passed)
71 steps (71 passed)
--- PASS: TestFeatures_HU23RevelarEstimaciones
```

## Fase RED

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `0159033` | `docs(HU-23): agregar especificación SDD` |
| RED | `eded0a0` | `test(HU-23): agregar tests (RED)` |

Antes de los tests se agregaron firmas vacías para que el paquete compile: `StatusRevealed`, `ErrNoVotes` con su mensaje y `Reveal`, que devolvía `Session{}, errNotImplemented`. La ruta de `server.go` no se agregó en el RED. `gofmt -l .` y `go vet ./...` no informaron nada antes del commit RED.

El RED falló por las aserciones y no por compilación, solo en `internal/poker` e `internal/httpapi`:

- Los 12 unitarios fallaron con el error del stub (por ejemplo, `error = no implementado, se esperaba ErrNoVotes`).
- BDD: `9 scenarios (9 failed)`, `71 steps (44 passed, 9 failed, 18 skipped)`; cada escenario falló en el paso que llama a `Reveal`.
- Los 8 tests HTTP y el caso de `TestRutasConID` recibieron `405` porque la ruta no existía; el `GET` recibió `404`.
- `TestInterfaz_BotonRevelarVotos` falló porque la página no tenía el botón ni la lista.

**El subtest de contrato “Update guarda el estado REVELADA y GetByID lo devuelve con sus votos” ya pasaba en el RED**, porque solo usa el `Update` existente de HU-22 y la constante `StatusRevealed` del stub. No aporta un fallo al RED; documenta que la persistencia del estado nuevo cumple el contrato.

## Fase GREEN

| Fase | Commit | Mensaje |
|---|---|---|
| GREEN | `213be95` | `feat(HU-23): implementar la revelación de estimaciones (GREEN)` |
| Refactor | — | No aplica |

Antes de escribir el código, `go tool cover -func` mostró que la única función de HU-23 existente era el stub `Reveal` (100.0%, llamado por los tests). Después del GREEN, `Reveal`, `revealPokerSession`, `statusFor` y `NewHandler` quedan al 100.0%: no hay código nuevo sin un test que lo pida. Ningún archivo `*_test.go` ni el `.feature` cambió en el GREEN.

Recorrido en el navegador integrado, con el servidor en memoria: revelar sin votos mostró “la sesión de Planning Poker no tiene votos para revelar” y la sesión siguió ABIERTA; después de votar Ana 5 y Bruno 8 desde la interfaz, “Revelar votos” mostró “Sesión 1 · ítem 1 · REVELADA · ronda 1”, la lista “Ana: 5” y “Bruno: 8”, el formulario de voto oculto y “Votos revelados en la sesión 1.”.

## Validación final

Medido sobre `213be95` con go1.27.1 windows/amd64, sin `TEST_DATABASE_URL`.

### `gofmt -l .`

```text
$ gofmt -l .
(sin salida, exit 0)
```

### `go vet ./...`

```text
$ go vet ./...
(sin salida, exit 0)
```

### `go test -count=1 ./...`

```text
ok  	seguimiento-medicion-gpc/cmd/server	20.113s
ok  	seguimiento-medicion-gpc/internal/backlog	0.798s
ok  	seguimiento-medicion-gpc/internal/defect	0.569s
ok  	seguimiento-medicion-gpc/internal/effort	0.618s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.680s
ok  	seguimiento-medicion-gpc/internal/metrics	0.516s
ok  	seguimiento-medicion-gpc/internal/poker	0.545s
ok  	seguimiento-medicion-gpc/internal/postgres	0.226s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.335s
ok  	seguimiento-medicion-gpc/internal/project	0.575s
ok  	seguimiento-medicion-gpc/internal/sprint	0.689s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
(exit 0)
```

### `go test -count=1 -cover ./...`

```text
ok  	seguimiento-medicion-gpc/cmd/server	0.241s	coverage: 31.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	0.834s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/defect	0.572s	coverage: 76.6% of statements
ok  	seguimiento-medicion-gpc/internal/effort	0.666s	coverage: 75.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	0.714s	coverage: 98.2% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	0.513s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	0.542s	coverage: 65.2% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	0.196s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.275s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	0.632s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	0.761s	coverage: 76.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
(exit 0)
```

Comparado con `main` (`8e06191`), medido con el mismo Go en una copia temporal (`git worktree`): `internal/poker` sube de 60.2% a 65.2%, `internal/httpapi` se mantiene en 98.2% y `cmd/server` en 31.5%. Sin base, `postgres_repository.go` de poker no se ejecuta porque sus tests de integración se omiten.

Una primera corrida del mismo comando dio las mismas cifras pero salió con código 1 por `go: unlinkat ... server.test.exe: The process cannot access the file because it is being used by another process.`: Windows no dejó borrar un binario temporal de test, sin relación con los tests. La salida de arriba es la de la corrida repetida, que terminó con exit 0.

### `-race`

`-race` no se corrió localmente: `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`. El test de concurrencia (`TestReveal_ConcurrenteConVotosNoPierdeVotosAceptados`) y el escenario de CA-23.7 corrieron sin `-race` y pasaron; con `-race` los ejecuta el CI del Pull Request (Linux, `go test -race -coverprofile=coverage.out ./...`).

Las pruebas de integración PostgreSQL se omiten porque no está definida la variable `TEST_DATABASE_URL`. La medición con PostgreSQL 16 queda pendiente de Peñalbé Hernán.

## Commits

Los commits SDD, RED y GREEN referencian el Issue #41 con `Refs #41`. El commit documental de esta evidencia y la cobertura queda propuesto hasta autorización.

HU-23: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
