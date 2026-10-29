# Evidencia TDD · HU-24

> Issue del repositorio: #42

## Historia

Como Agile Enabler, quiero detectar si hay diferencias entre las estimaciones y abrir otra ronda, para llegar a un acuerdo.

## Resultado

Se agregó la función pura `CompareVotes` en `internal/poker/compare.go`, con los tipos `VoteComparison` y `VoteGroup`, que indica si los votos revelados tienen diferencias y cuáles son el voto menor y el mayor con sus integrantes; y la operación `RepeatRound(sessionID)` en el `Manager` de `internal/poker`, que vuelve a abrir una sesión `REVELADA` en la ronda siguiente y sin votos, bajo el mismo mutex que las demás operaciones y con la comprobación de `OpenByItem` dentro del lock (D8). La respuesta de `POST /api/poker-sessions/{id}/reveal` suma el campo `comparacion` (las demás rutas lo omiten), se agregó la ruta `POST /api/poker-sessions/{id}/rounds` con su `405` para `GET`, y en la sección **Planning Poker** se muestran la comparación y el botón "Repetir ronda" (`type="button"`) dentro de `form-estimar-poker`. No cambian `Manager.Reveal`, `session.go`, los repositorios, `create.go`, `statusFor` ni el esquema.

## Trazabilidad

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-24.1 — se indica si coinciden o hay diferencias, con el menor y el mayor | "Revelar votos distintos informa la diferencia con el menor y el mayor"; "Revelar votos iguales informa que no hay diferencias"; "Revelar un único voto no informa diferencias" | `TestCompareVotes_VotosDistintosInformaDiferenciaMenorYMayor`; `TestCompareVotes_TodosIgualesNoInformaDiferencia`; `TestCompareVotes_UnSoloVotoNoInformaDiferencia`; `TestCompareVotes_VariosIntegrantesConElMismoExtremo`; `TestCompareVotes_SinVotosDevuelveComparacionVacia`; `TestRevelarPoker_IncluyeComparacionConDiferencias`; `TestRevelarPoker_IncluyeComparacionSinDiferencias` |
| CA-24.2 — nueva ronda en la misma sesión, que descarta los votos | "Repetir la ronda de una sesión revelada" | `TestRepeatRound_SesionReveladaVuelveAAbiertaConRondaSiguienteYSinVotos` |
| CA-24.3 — la sesión lleva el número de ronda | "Repetir la ronda de una sesión revelada"; "Repetir la ronda dos veces seguidas llega a la ronda 3" | `TestRepeatRound_SesionReveladaVuelveAAbiertaConRondaSiguienteYSinVotos`; `TestRepeatRound_ConservaIDEItemYNoModificaStoryPoints` |
| CA-24.4 — se puede repetir sin diferencias | "Repetir la ronda aunque no haya diferencias" | `TestRepeatRound_SinDiferenciasTambienSePuedeRepetir` |
| CA-24.5 — la ronda nueva funciona como la primera | "Votar, revelar y registrar la estimación en la ronda nueva" | `TestRepeatRound_NuevaRondaPermiteVotarRevelarYEstimar` |
| CA-24.6 — no modifica el ítem y devuelve copias | "Repetir la ronda de una sesión revelada" (ítem sin cambios) | `TestRepeatRound_ConservaIDEItemYNoModificaStoryPoints`; `TestRepeatRound_DevuelveCopia`; `TestCompareVotes_DevuelveListasNuevas` |
| CA-24.7 — sesión inexistente, abierta o cerrada da error | "Rechazar repetir la ronda de una sesión abierta"; "… de una sesión cerrada"; esquema "… de una sesión inexistente" (3 ejemplos: 999, 0 y -1) | `TestRepeatRound_SesionInexistenteDevuelveErrSessionNotFound`; `TestRepeatRound_SesionAbiertaDevuelveErrSessionNotRevealed`; `TestRepeatRound_SesionCerradaDevuelveErrSessionNotRevealed` |
| CA-24.8 — otra sesión abierta para el ítem | "Rechazar repetir la ronda si hay otra sesión abierta para el ítem" | `TestRepeatRound_OtraSesionAbiertaDelItemDevuelveErrSessionAlreadyOpen` |
| CA-24.9 — falla del repositorio | "Si falla el guardado, la sesión sigue revelada" | `TestRepeatRound_ErrorAlLeerPropagaError`; `TestRepeatRound_ErrorAlBuscarSesionAbiertaPropagaError`; `TestRepeatRound_ErrorAlGuardarPropagaErrorYSesionNoCambia` |
| CA-24.10 — concurrencia | "Repetir la ronda varias veces al mismo tiempo la aumenta una sola vez" | `TestRepeatRound_ConcurrenteAumentaLaRondaUnaSolaVez`; `TestRepeatRound_ConcurrenteConVotosNoMezclaRondas`; `TestRepeatRound_ConcurrenteConAperturaDejaUnaSolaSesionAbierta` |
| CA-24.11 — persistencia en memoria y PostgreSQL | Solo contrato de repositorio (ver "Desvíos y límites") | Subtest `Update guarda una ronda nueva ABIERTA sin votos y OpenByItem la encuentra` de `runRepositoryContract` |
| CA-24.12 — API e interfaz | Solo HTTP e interfaz (ver "Desvíos y límites") | `TestVotarPoker_RespuestaSinComparacion`; `TestRepetirRondaPoker_Devuelve200ConRondaSiguienteYSinVotos`; `TestRepetirRondaPoker_IDNoNumericoDevuelve400`; `TestRepetirRondaPoker_InexistenteDevuelve404`; `TestRepetirRondaPoker_SesionNoReveladaDevuelve409`; `TestRepetirRondaPoker_OtraSesionAbiertaDevuelve409`; `TestRepetirRondaPoker_ErrorDelRepositorioDevuelve500SinDetalle`; `TestRepetirRondaPoker_MetodoNoPermitidoDevuelve405`; `TestRepetirRondaPoker_IgnoraElCuerpo`; fila "repetir ronda Planning Poker" de `TestRutasConID_IDNoNumericoDevuelve400`; `TestInterfaz_DiferenciasYBotonRepetirRonda`; los dos `TestRevelarPoker_IncluyeComparacion…`; prueba en el navegador |

La correspondencia no es 1 a 1: varios escenarios cubren más de un criterio, y CA-24.11 y CA-24.12 se verifican solo con tests de contrato, HTTP e interfaz.

## Conteos

```text
$ grep -cE '^[[:space:]]*Escenario:' features/HU-24-repetir-ronda.feature
12
$ grep -cE '^[[:space:]]*Esquema del escenario:' features/HU-24-repetir-ronda.feature
1
$ grep -c '^func Test' internal/poker/compare_test.go
6
$ grep -c '^func Test' internal/poker/round_test.go
15
$ grep -c '^func Test' internal/poker/round_features_test.go
1
$ git diff --unified=0 8479508 b0d9b69 -- internal/httpapi/poker_test.go | grep -c '^+func Test'
11
$ git diff --unified=0 8479508 b0d9b69 -- internal/httpapi/web_test.go | grep -c '^+func Test'
1
```

**34 funciones `Test` nuevas** (6 de `CompareVotes`, 15 de `RepeatRound`, 1 runner de godog, 11 de API y 1 de interfaz) y **2 subtests nuevos** (la fila "repetir ronda Planning Poker" de `TestRutasConID_IDNoNumericoDevuelve400` y el subtest de contrato de la ronda nueva), que cubren **12 escenarios y 1 esquema** BDD. Godog expande el esquema (3 ejemplos) en **15 casos**.

## Fase RED (`b0d9b69`)

Firmas vacías: en `internal/poker/compare.go`, los tipos `VoteGroup` y `VoteComparison` y `CompareVotes`, que devuelve `VoteComparison{}`; en `internal/poker/manager.go`, `RepeatRound`, que devuelve `Session{}, nil`. La ruta, el handler, el campo `comparacion` y la interfaz van en el GREEN. El RED solo agrega esto fuera de los tests y el `.feature`:

```text
$ git diff --stat a1ff5b5 b0d9b69 -- ':!*_test.go' ':!*.feature' ':!docs'
 internal/poker/compare.go | 22 ++++++++++++++++++++++
 internal/poker/manager.go |  6 ++++++
 2 files changed, 28 insertions(+)
```

Medido en una copia temporal con `git worktree` en `b0d9b69`, con go1.27.1: `go vet ./...` sin salida (exit 0) y `GOTMPDIR=C:/gotmp go test -count=1 -p 1 ./...`, con los tiempos recortados:

```text
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
--- FAIL: TestRevelarPoker_IncluyeComparacionConDiferencias
--- FAIL: TestRevelarPoker_IncluyeComparacionSinDiferencias
--- FAIL: TestVotarPoker_RespuestaSinComparacion
--- FAIL: TestRepetirRondaPoker_Devuelve200ConRondaSiguienteYSinVotos
--- FAIL: TestRepetirRondaPoker_IDNoNumericoDevuelve400
--- FAIL: TestRepetirRondaPoker_InexistenteDevuelve404
--- FAIL: TestRepetirRondaPoker_SesionNoReveladaDevuelve409
--- FAIL: TestRepetirRondaPoker_OtraSesionAbiertaDevuelve409
--- FAIL: TestRepetirRondaPoker_ErrorDelRepositorioDevuelve500SinDetalle
--- FAIL: TestRepetirRondaPoker_MetodoNoPermitidoDevuelve405
--- FAIL: TestRepetirRondaPoker_IgnoraElCuerpo
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400
--- FAIL: TestInterfaz_DiferenciasYBotonRepetirRonda
FAIL seguimiento-medicion-gpc/internal/httpapi
ok   seguimiento-medicion-gpc/internal/metrics
--- FAIL: TestCompareVotes_VotosDistintosInformaDiferenciaMenorYMayor
--- FAIL: TestCompareVotes_TodosIgualesNoInformaDiferencia
--- FAIL: TestCompareVotes_UnSoloVotoNoInformaDiferencia
--- FAIL: TestCompareVotes_VariosIntegrantesConElMismoExtremo
--- FAIL: TestCompareVotes_SinVotosDevuelveComparacionVacia
--- FAIL: TestCompareVotes_DevuelveListasNuevas
--- FAIL: TestFeatures_HU24RepetirRonda
--- FAIL: TestRepeatRound_SesionReveladaVuelveAAbiertaConRondaSiguienteYSinVotos
--- FAIL: TestRepeatRound_ConservaIDEItemYNoModificaStoryPoints
--- FAIL: TestRepeatRound_SinDiferenciasTambienSePuedeRepetir
--- FAIL: TestRepeatRound_NuevaRondaPermiteVotarRevelarYEstimar
--- FAIL: TestRepeatRound_SesionInexistenteDevuelveErrSessionNotFound
--- FAIL: TestRepeatRound_SesionAbiertaDevuelveErrSessionNotRevealed
--- FAIL: TestRepeatRound_SesionCerradaDevuelveErrSessionNotRevealed
--- FAIL: TestRepeatRound_OtraSesionAbiertaDelItemDevuelveErrSessionAlreadyOpen
--- FAIL: TestRepeatRound_ErrorAlLeerPropagaError
--- FAIL: TestRepeatRound_ErrorAlBuscarSesionAbiertaPropagaError
--- FAIL: TestRepeatRound_ErrorAlGuardarPropagaErrorYSesionNoCambia
--- FAIL: TestRepeatRound_DevuelveCopia
--- FAIL: TestRepeatRound_ConcurrenteAumentaLaRondaUnaSolaVez
--- FAIL: TestRepeatRound_ConcurrenteConVotosNoMezclaRondas
--- FAIL: TestRepeatRound_ConcurrenteConAperturaDejaUnaSolaSesionAbierta
FAIL seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
?    seguimiento-medicion-gpc/internal/web [no test files]
```

Fallan **35 tests de primer nivel, todos de HU-24**: **22 en `internal/poker`** (los 6 de `compare_test.go`, los 15 de `round_test.go` y el runner de godog) y **13 en `internal/httpapi`** (los 11 de `poker_test.go`, el de interfaz y `TestRutasConID_IDNoNumericoDevuelve400`, solo por su subtest `repetir_ronda_Planning_Poker`). El conteo sale de la salida completa: `grep -c '^--- FAIL:'` da 35 y `grep -ciE 'build failed|undefined:|cannot use|\[setup failed\]|panic'` da 0. Todos los fallos son de aserción:

- dominio: los stubs devuelven valores vacíos y `error = <nil>` (por ejemplo, `RepeatRound(999): error = <nil>, se esperaba ErrSessionNotFound` y `menor = {Points:0 Voters:[]}, se esperaba 3 de [Ana]`);
- API: la ruta todavía no existe (`código = 405, se esperaba 200; cuerpo: Method Not Allowed` y, para `GET`, `código = 404, se esperaba 405`) y `/reveal` no trae `comparacion`;
- interfaz: faltan el párrafo, el botón y el código de `app.js`.

Con `go test -count=1 -p 1 -v ./internal/poker ./internal/httpapi` sobre el RED: **284 PASS, 35 FAIL y 2 SKIP** de primer nivel, y ningún `--- FAIL` fuera de los tests de HU-24 (los tests de HU-21, HU-22, HU-23 y HU-25 siguen en verde). Godog de HU-24 en el RED: `15 scenarios (15 failed)` y `145 steps (99 passed, 15 failed, 31 skipped)`; los runners de HU-21, HU-22, HU-23 y HU-25 pasan completos. El subtest de contrato nuevo **pasa** en el RED (ver "Desvíos y límites").

## Fase GREEN (`42ebcce`)

- `internal/poker/compare.go`: `CompareVotes` toma el menor y el mayor de los puntos y junta a los integrantes de cada extremo en el orden de los votos, en listas nuevas y nunca `nil`; hay diferencia si el menor y el mayor son distintos.
- `internal/poker/manager.go`: `RepeatRound`, bajo `m.mu`, lee la sesión, exige `REVELADA` (`ErrSessionNotRevealed`), busca con `OpenByItem` otra sesión abierta del ítem (`ErrSessionAlreadyOpen`), pasa a `ABIERTA`, aumenta `Round`, vacía los votos y guarda con `Update`; devuelve una copia con `cloneSession`.
- `internal/httpapi/poker.go`: el campo `Comparacion` (`json:"comparacion,omitempty"`) en `pokerSessionResponse`, que solo completa `revealPokerSession`, y el handler `repeatRoundPokerSession`.
- `internal/httpapi/server.go`: las rutas `GET /api/poker-sessions/{id}/rounds` (con `methodNotAllowed`) y `POST /api/poker-sessions/{id}/rounds`.
- `internal/web/index.html` y `internal/web/app.js`: el párrafo `diferencias-poker` y el botón `btn-repetir-ronda-poker` (`type="button"`) dentro de `form-estimar-poker`, la función `mostrarDiferenciasPoker`, su llamada en el manejador de "Revelar votos" y el listener de "Repetir ronda", que reutiliza `mostrarPoker` sin modificarla.

El GREEN no modifica tests ni el `.feature`:

```text
$ git diff --stat b0d9b69 42ebcce -- '*_test.go' '*.feature'
(sin salida)
```

Salida sobre el estado actual de la rama (`2c05ccb`): `gofmt -l .` y `go vet ./...` sin salida, y `GOTMPDIR=C:/gotmp go test -count=1 -p 1 ./...` con exit 0:

```text
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
ok   seguimiento-medicion-gpc/internal/httpapi
ok   seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
?    seguimiento-medicion-gpc/internal/web [no test files]
```

Con `go test -count=1 -p 1 -v ./internal/poker ./internal/httpapi`: **347 PASS, 0 FAIL y 2 SKIP** de primer nivel, con los 34 tests nuevos de HU-24 en PASS y los dos subtests nuevos en PASS. Godog de HU-24: `15 scenarios (15 passed)` y `145 steps (145 passed)`. Con `go test -count=1 -p 1 -v ./...` sobre toda la rama: 896 PASS de primer nivel, 0 `--- FAIL` y 37 `--- SKIP` (los tests de PostgreSQL, sin `TEST_DATABASE_URL`).

## Refactor

No se hizo refactor: el código del GREEN no tiene duplicaciones claras ni cambios de estructura que mejoren la legibilidad sin cambiar el comportamiento.

## Prueba en el navegador

Con el servidor en memoria (`go run ./cmd/server`, puerto 8080, sin `DATABASE_URL`), se creó por la API un proyecto y un ítem, y en la interfaz se cargó el proyecto 1 y se usó "Abrir Poker" en el ítem. Con el integrante "Ana" se votó 3 y con "Bruno" 8, y "Revelar votos" mostró (captura de pantalla del navegador):

- "Sesión 1 · ítem 1 · REVELADA · ronda 1";
- la lista "Ana: 3" y "Bruno: 8";
- "Hay diferencias: menor 3 (Ana), mayor 8 (Bruno).", junto al botón "Repetir ronda" y los botones de "Registrar estimación acordada".

"Repetir ronda" dejó la sección (texto y visibilidad leídos del DOM) en "Sesión 1 · ítem 1 · ABIERTA · ronda 2", con los botones de voto visibles, la lista de votos y el bloque de la estimación ocultos y el mensaje "Ronda 2 iniciada en la sesión 1.". La página no se recargó: la URL siguió en `http://localhost:8080/` y `performance.getEntriesByType('navigation').length` dio 1.

## Cobertura

Medición local sin base, en Windows con go1.27.1, `GOTMPDIR=C:/gotmp go test -count=1 -p 1 -coverprofile=cover.out ./...` y `go tool cover -func=cover.out`, sobre `2c05ccb` (el GREEN con `main` traído), y en `main` (`f9af5a9`) con el mismo Go y los mismos comandos, en una copia temporal con `git worktree`:

| Paquete | Rama | `main` (`f9af5a9`) |
|---|---|---|
| `cmd/server` | 31.5% | 31.5% |
| `internal/backlog` | 83.3% | 83.3% |
| `internal/dashboard` | 100.0% | 100.0% |
| `internal/defect` | 77.7% | 77.7% |
| `internal/effort` | 75.2% | 75.2% |
| `internal/httpapi` | 98.8% | 98.8% |
| `internal/metrics` | 100.0% | 100.0% |
| `internal/poker` | 75.5% | 69.5% |
| `internal/postgres` | 83.3% | 83.3% |
| `internal/postgrestest` | 36.8% | 36.8% |
| `internal/project` | 71.1% | 71.1% |
| `internal/report` | 100.0% | 100.0% |
| `internal/sprint` | 76.3% | 76.3% |
| `internal/web` | 0.0% | 0.0% |
| **Total del perfil** | **84.4%** | **83.9%** |

```text
internal/httpapi/poker.go:114:  revealPokerSession       100.0%
internal/httpapi/poker.go:145:  repeatRoundPokerSession  100.0%
internal/poker/compare.go:22:   CompareVotes             100.0%
internal/poker/manager.go:131:  RepeatRound              100.0%
total:                          (statements)             84.4%
```

Solo cambia `internal/poker`, de 69.5% a 75.5%; los demás paquetes conservan sus porcentajes, y el total sube de 83.9% a 84.4%. En `main`, `revealPokerSession` (`poker.go:100`) también está en 100.0%.

## Desvíos y límites

1. **Ejecutables de test bloqueados en Windows y método de medición.** Con la carpeta temporal por defecto, el ejecutable de `internal/httpapi` no arrancaba: `go test -count=1 -v -timeout 60s ./internal/httpapi` no imprimió ningún `=== RUN` ni volcado de goroutines en 200 s, y el mismo paquete compilado aparte con `go test -c` y ejecutado directamente pasó en 3 s (272 tests de primer nivel en PASS, 0 FAIL). La causa probable es el antivirus sobre los ejecutables nuevos de la carpeta temporal; no se verificó. Por eso todas las mediciones de esta evidencia usan `GOTMPDIR=C:/gotmp` (fuera de OneDrive) con `-p 1`. En algunas corridas Go termina con `go: unlinkat … The process cannot access the file because it is being used by another process` al borrar un ejecutable cuando todos los paquetes ya informaron `ok`: ese error da exit 1 sin fallos de tests, y la corrida se repitió después de `go clean -testcache`.
2. **Sin `-race`.** Localmente responde `-race requires cgo` (`CGO_ENABLED=0`). Los tres tests de concurrencia (`TestRepeatRound_Concurrente…`) y el escenario de CA-24.10 pasan sin el detector de carreras; lo ejecuta el CI.
3. **Sin medición con PostgreSQL.** Sin `TEST_DATABASE_URL`, se omiten los tests de PostgreSQL (2 en `internal/poker` e `internal/httpapi`, 37 en toda la rama: los 34 de la medición de HU-39 más los 3 de `internal/project` que agregó HU-41 en `3b01a87`; HU-24 no agrega tests de PostgreSQL de primer nivel, porque su subtest de contrato corre dentro de `TestPostgresRepository_Contrato`, ya contado): el subtest de contrato nuevo solo corrió contra `MemoryRepository`. HU-24 no cambia el repositorio ni el esquema.
4. **Subtest de contrato en verde desde el RED.** `Update guarda una ronda nueva ABIERTA sin votos y OpenByItem la encuentra` pasa ya en el RED porque verifica comportamiento existente de `Update` y `OpenByItem`, que la SDD no cambia (D12).
5. **Merge de `main`.** Antes de la evidencia se trajo `main` con HU-39 y HU-41 (merge `2c05ccb`), sin conflictos; se conservaron las rutas, los tests y la interfaz de las dos historias junto con los de HU-24.
6. **CA-24.11 y CA-24.12 fuera del `.feature`.** Los pasos de godog llaman al `Manager` y a `CompareVotes`, no al handler HTTP, así que la persistencia y la API e interfaz se prueban con el test de contrato y los tests HTTP y de interfaz, como en HU-23.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `8479508` | `docs(HU-24): agregar la especificación (SDD)` |
| RED | `b0d9b69` | `test(HU-24): agregar tests (RED)` |
| GREEN | `42ebcce` | `feat(HU-24): implementar la detección de diferencias y la repetición de ronda (GREEN)` |
| Traer `main` | `2c05ccb` | `chore(HU-24): traer main con HU-39 y HU-41` |

HU-24: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
