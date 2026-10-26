# Evidencia TDD · HU-25

> Issue del repositorio: #43

## Historia

Como Agile Enabler, quiero registrar la estimación acordada por el equipo, para guardarla en el ítem del backlog.

## Resultado

HU-25 quedó implementada en la rama `feature/HU-25-registrar-estimacion-acordada`. `Manager.Estimate` valida la carta, busca la sesión, exige que esté `REVELADA`, asigna los Story Points al ítem con `backlog.Manager.AssignStoryPoints` (HU-20) y deja la sesión `CERRADA`, todo bajo el mutex del `Manager` de poker (orden de locks poker → backlog). La API lo expone en `POST /api/poker-sessions/{id}/estimate` con `{"puntos": N}`, y la interfaz web con el formulario "Registrar estimación acordada", que aparece al revelar, muestra un botón por carta y recarga el backlog para ver el valor en la columna SP.

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test |
|---|---|
| CA-25.1 — el valor acordado queda como Story Points del ítem | “Registrar la estimación acordada guarda los puntos en el ítem y cierra la sesión”; “Registrar la estimación reemplaza los Story Points anteriores del ítem”; `TestEstimate_SesionReveladaAsignaLosPuntosYCierraLaSesion`; `TestEstimate_ReemplazaLosStoryPointsAnterioresDelItem` |
| CA-25.2 — solo valores de la escala, sin exigir que coincidan con los votos | “Registrar un valor que nadie votó”; “Registrar el valor más alto de la escala”; “Rechazar una carta fuera de la escala”; `TestEstimate_ValorQueNadieVotoSeAcepta`; `TestEstimate_AceptaCadaValorDeLaEscala`; `TestEstimate_CartaFueraDeLaEscalaDevuelveErrInvalidCard`; `TestEstimate_CartaInvalidaSeInformaAntesQueLaSesionInexistente` |
| CA-25.3 — la sesión queda cerrada, con su ronda y sus votos | “Registrar la estimación acordada guarda los puntos en el ítem y cierra la sesión”; `TestEstimate_SesionReveladaAsignaLosPuntosYCierraLaSesion`; `TestEstimate_ConservaRondaYOrdenDeLosVotos`; `TestVoteYReveal_SesionCerradaDevuelvenErrSessionNotOpen` |
| CA-25.4 — sesión inexistente, abierta o cerrada da error sin cambios | “Rechazar registrar la estimación de una sesión sin revelar”; “Rechazar registrar dos veces la misma sesión”; “Rechazar registrar la estimación de una sesión inexistente”; `TestEstimate_SesionInexistenteDevuelveErrSessionNotFound`; `TestEstimate_SesionAbiertaDevuelveErrSessionNotRevealed`; `TestEstimate_SesionCerradaDevuelveErrSessionNotRevealed`; `TestEstimate_ErrorAlLeerPropagaError` |
| CA-25.5 — ítem inexistente o terminado: error del backlog y la sesión sigue revelada | “Rechazar registrar la estimación de un ítem terminado”; `TestEstimate_ItemInexistenteDevuelveErrItemNotFoundYSesionSigueRevelada`; `TestEstimate_ItemTerminadoDevuelveErrItemFinishedYSesionSigueRevelada`; `TestEstimate_ErrorDelBacklogAlAsignarPropagaErrorYSesionSigueRevelada` |
| CA-25.6 — si falla el guardado de la sesión, el ítem queda estimado y la sesión sigue revelada | `TestEstimate_ErrorAlGuardarLaSesionDejaElItemEstimadoYLaSesionRevelada` |
| CA-25.7 — modificar lo devuelto no altera lo guardado | `TestEstimate_DevuelveCopiaDeLosVotos` |
| CA-25.8 — dos registros simultáneos: solo uno se aplica | “Dos registros simultáneos: solo uno se aplica”; `TestEstimate_SimultaneosSobreLaMismaSesionSoloUnoSeAplica` |
| CA-25.9 — registrar desde la API y la interfaz web | `TestRegistrarEstimacionPoker_Devuelve200ConSesionCerradaYStoryPointsEnElItem`; `TestRegistrarEstimacionPoker_IDNoNumericoDevuelve400`; `TestRegistrarEstimacionPoker_JSONInvalidoDevuelve400`; `TestRegistrarEstimacionPoker_CartaFueraDeLaEscalaDevuelve400`; `TestRegistrarEstimacionPoker_SesionInexistenteDevuelve404`; `TestRegistrarEstimacionPoker_ItemInexistenteDevuelve404`; `TestRegistrarEstimacionPoker_SesionNoReveladaDevuelve409`; `TestRegistrarEstimacionPoker_ItemTerminadoDevuelve409`; `TestRegistrarEstimacionPoker_ErrorAlGuardarLaSesionDevuelve500SinDetalle`; `TestRegistrarEstimacionPoker_MetodoNoPermitidoDevuelve405`; caso “registrar estimación Planning Poker” de `TestRutasConID_IDNoNumericoDevuelve400`; entrada “registrar estimación Planning Poker” de `bodyEndpoints` (`TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400`); `TestInterfaz_FormularioRegistrarEstimacion` |
| CA-25.10 — mismo comportamiento en memoria y PostgreSQL | Subtest de contrato “Update guarda el estado CERRADA y GetByID lo devuelve con sus votos”, ejecutado por `TestMemoryRepository_Contrato` y `TestPostgresRepository_Contrato`; PostgreSQL se omite sin `TEST_DATABASE_URL` |

La validación contiene **10 escenarios BDD** y **31 casos de prueba específicos**: 18 unitarios de `internal/poker/estimate_test.go`, 1 subtest de contrato de repositorio, 10 tests HTTP, 1 caso nuevo en `TestRutasConID_IDNoNumericoDevuelve400` y 1 test de interfaz web. Además, la entrada nueva de `bodyEndpoints` agrega un subtest a cada uno de `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400`. La correspondencia no es 1 a 1: CA-25.6, CA-25.7, CA-25.9 y CA-25.10 no tienen escenario BDD, y algunos escenarios cubren más de un criterio.

Conteos con `grep -c` sobre los archivos de la rama:

```text
$ grep -c "^  Escenario:" features/HU-25-registrar-estimacion-acordada.feature
10
$ grep -c "^func Test" internal/poker/estimate_test.go
18
$ grep -c "^func TestRegistrarEstimacionPoker_" internal/httpapi/poker_test.go
10
$ grep -c '"registrar estimación Planning Poker"' internal/httpapi/server_test.go
2
$ grep -c "^func TestInterfaz_FormularioRegistrarEstimacion" internal/httpapi/web_test.go
1
$ grep -c 'Update guarda el estado CERRADA' internal/poker/repository_contract_test.go
1
```

Las 2 apariciones de `"registrar estimación Planning Poker"` en `server_test.go` son el caso de `TestRutasConID_IDNoNumericoDevuelve400` y la entrada de `bodyEndpoints`.

## BDD

Archivo: `features/HU-25-registrar-estimacion-acordada.feature`, con el runner `TestFeatures_HU25RegistrarEstimacionAcordada` en `internal/poker/estimate_features_test.go`. Todos los pasos llaman al `Manager` de `internal/poker` (y al de `internal/backlog` para preparar y consultar el ítem), no al handler HTTP; por eso ningún escenario dice "mediante la API".

Resultado de `go test -count=1 -v -run TestFeatures_HU25 ./internal/poker/` sobre el GREEN (`894c7c2`):

```text
10 scenarios (10 passed)
106 steps (106 passed)
--- PASS: TestFeatures_HU25RegistrarEstimacionAcordada
ok  	seguimiento-medicion-gpc/internal/poker	1.357s
```

Los 106 pasos son 10 × 6 pasos de los Antecedentes más 46 pasos propios de los escenarios.

## Fase RED

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `c43773c` | `docs(HU-25): agregar especificación SDD` |
| RED | `6b832f1` | `test(HU-25): agregar tests (RED)` |

Antes de los tests se agregaron firmas vacías para que el paquete compile: `StatusClosed`, `ErrSessionNotRevealed` con su mensaje y `Estimate`, que devolvía `Session{}, errNotImplemented`. La ruta de `server.go` y el cambio de `statusFor` no se agregaron en el RED. `gofmt -l .` y `go vet ./...` no informaron nada antes del commit RED.

El RED falló por las aserciones y no por compilación, solo en `internal/poker` e `internal/httpapi`.

Salida de `go test -count=1 ./internal/poker` (aserciones de los unitarios y resumen de godog; el resto de los tests del paquete pasó):

```text
estimate_test.go:67: Estimate devolvió error: no implementado
estimate_test.go:85: Estimate devolvió error: no implementado
estimate_test.go:100: Estimate(1) devolvió error: no implementado
(… lo mismo para 2, 3, 5, 8, 13 y 21)
estimate_test.go:115: Estimate(0): error = no implementado, se esperaba ErrInvalidCard
estimate_test.go:115: Estimate(-1): error = no implementado, se esperaba ErrInvalidCard
estimate_test.go:115: Estimate(4): error = no implementado, se esperaba ErrInvalidCard
estimate_test.go:115: Estimate(22): error = no implementado, se esperaba ErrInvalidCard
estimate_test.go:130: error = no implementado, se esperaba ErrInvalidCard
estimate_test.go:140: Estimate devolvió error: no implementado
estimate_test.go:153: Estimate(999): error = no implementado, se esperaba ErrSessionNotFound
estimate_test.go:153: Estimate(0): error = no implementado, se esperaba ErrSessionNotFound
estimate_test.go:153: Estimate(-1): error = no implementado, se esperaba ErrSessionNotFound
estimate_test.go:169: error = no implementado, se esperaba ErrSessionNotRevealed
estimate_test.go:180: primer registro: no implementado
estimate_test.go:204: error = no implementado, se esperaba backlog.ErrItemNotFound
estimate_test.go:223: error = no implementado, se esperaba backlog.ErrItemFinished
estimate_test.go:261: error = no implementado, se esperaba errRepository
estimate_test.go:308: error = no implementado, se esperaba errRepository
estimate_test.go:323: error = no implementado, se esperaba errRepository
estimate_test.go:348: Estimate devolvió error: no implementado
estimate_test.go:369: Estimate devolvió error: no implementado
estimate_test.go:391: Estimate: no implementado
estimate_test.go:440: intento 0: registro de 8 puntos falló con no implementado, se esperaba ErrSessionNotRevealed
10 scenarios (10 failed)
106 steps (74 passed, 10 failed, 22 skipped)
FAIL	seguimiento-medicion-gpc/internal/poker
```

- Los 18 unitarios fallaron con el error del stub.
- BDD: `10 scenarios (10 failed)`, `106 steps (74 passed, 10 failed, 22 skipped)`; cada escenario falló en el paso que llama a `Estimate` (en “Rechazar registrar dos veces la misma sesión”, en el paso previo “que se registró la estimación acordada de 5 puntos”).

Salida de `go test -count=1 -v -run "TestRegistrarEstimacionPoker|TestInterfaz_FormularioRegistrarEstimacion" ./internal/httpapi` (primeras líneas de error de cada test):

```text
=== RUN   TestRegistrarEstimacionPoker_Devuelve200ConSesionCerradaYStoryPointsEnElItem
    poker_test.go:403: POST /api/poker-sessions/1/estimate: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    poker_test.go:405: código = 405, se esperaba 200; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_IDNoNumericoDevuelve400
    poker_test.go:424: código = 405, se esperaba 400; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_JSONInvalidoDevuelve400 (3 subtests)
    poker_test.go:435: código = 405, se esperaba 400; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_CartaFueraDeLaEscalaDevuelve400 (3 subtests)
    poker_test.go:449: código = 405, se esperaba 400; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_SesionInexistenteDevuelve404
    poker_test.go:458: código = 405, se esperaba 404; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_ItemInexistenteDevuelve404
    poker_test.go:484: código = 405, se esperaba 404; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_SesionNoReveladaDevuelve409/ABIERTA
    poker_test.go:501: código = 405, se esperaba 409; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_SesionNoReveladaDevuelve409/CERRADA
    poker_test.go:507: código = 405, se esperaba 200; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_ItemTerminadoDevuelve409
    poker_test.go:525: código = 405, se esperaba 409; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_ErrorAlGuardarLaSesionDevuelve500SinDetalle
    poker_test.go:541: código = 405, se esperaba 500; cuerpo: Method Not Allowed
=== RUN   TestRegistrarEstimacionPoker_MetodoNoPermitidoDevuelve405
    poker_test.go:561: GET: código = 404, se esperaba 405
=== RUN   TestInterfaz_FormularioRegistrarEstimacion
    web_test.go:200: la página no contiene "id=\"form-estimar-poker\""
    web_test.go:200: la página no contiene "Registrar estimación acordada"
    web_test.go:200: la página no contiene "id=\"botones-estimar-poker\""
    web_test.go:206: app.js no contiene "\"/estimate\""
    web_test.go:206: app.js no contiene "\"Estimación registrada: \""
    web_test.go:206: app.js no contiene "\" story points en la sesión \""
FAIL	seguimiento-medicion-gpc/internal/httpapi
```

- Los 10 tests HTTP fallaron en el `POST` a `/estimate`: recibieron `405 text/plain` porque la ruta no existía. La preparación de cada test (abrir, votar, revelar y cambiar el estado del ítem por la API) pasó, así que el fallo no viene del armado del test. El `GET` recibió `404` (lo atiende el servidor de archivos de la interfaz); `PUT` y `DELETE` ya recibían `405`.
- El caso de `TestRutasConID_IDNoNumericoDevuelve400` y la entrada de `bodyEndpoints` (413 y texto después del JSON) también recibieron `405`.
- `TestInterfaz_FormularioRegistrarEstimacion` falló porque la página no tenía el formulario ni `app.js` la ruta y el mensaje.

**El subtest de contrato “Update guarda el estado CERRADA y GetByID lo devuelve con sus votos” ya pasaba en el RED**, porque solo usa el `Update` existente de HU-22 y la constante `StatusClosed` del stub. No aporta un fallo al RED; documenta que la persistencia del estado nuevo cumple el contrato.

Antes del commit RED se corrigió el comentario de `revealFailingSession`, que decía que revelaba con `revealedVotes` aunque el helper solo votaba a Bruno: `openFailingSession` (HU-23) ya vota a Ana con 5, así que los votos sí son los de `revealedVotes`. Los mismos tests siguieron fallando por aserción.

## Fase GREEN

| Fase | Commit | Mensaje |
|---|---|---|
| GREEN | `894c7c2` | `feat(HU-25): implementar el registro de la estimación acordada (GREEN)` |
| Refactor | — | No aplica |
| Documentación | este commit | `docs(HU-25): agregar evidencia TDD y cobertura` |

El GREEN cambió solo `internal/poker/manager.go` (`Estimate`; se borraron el stub `errNotImplemented` y el import `errors`), `internal/httpapi/poker.go` (`estimatePokerSession` y `pokerEstimateRequest`), `internal/httpapi/server.go` (las rutas `POST` y `GET` con `methodNotAllowed`, y `ErrSessionNotRevealed` en el grupo 409 de `statusFor`), `internal/web/index.html` y `internal/web/app.js`. Ningún archivo `*_test.go` ni el `.feature` cambió en el GREEN (`git diff --name-only 6b832f1 894c7c2` lista solo esos 5 archivos). Después del GREEN, `Estimate`, `estimatePokerSession`, `statusFor` y `NewHandler` quedan al 100.0%.

### Recorrido en el navegador

Con el servidor compilado de `894c7c2` (`go build -o server.exe ./cmd/server`, sin `go run`), en memoria, en el navegador integrado:

1. Crear el proyecto “Sistema de ventas” (“Proyecto 1 creado.”) y el ítem “Login de usuarios” (“Ítem 1 creado.”); la columna SP mostraba “—”.
2. “Abrir Poker” en la fila del ítem: “Sesión 1 abierta para el ítem 1.” y “Sesión 1 · ítem 1 · ABIERTA · ronda 1”; el formulario “Registrar estimación acordada” estaba oculto.
3. Votar 5 como Ana y 8 como Bruno, y “Revelar votos”: “Sesión 1 · ítem 1 · REVELADA · ronda 1”, la lista “Ana: 5” y “Bruno: 8”, “Votos revelados en la sesión 1.” y el formulario “Registrar estimación acordada” visible con los botones 1, 2, 3, 5, 8, 13 y 21.
4. Elegir 13: “Sesión 1 · ítem 1 · CERRADA · ronda 1”, “Estimación registrada: 13 story points en la sesión 1.”, el formulario oculto y la fila del ítem con SP = 13 (el backlog se recargó con `GET /api/proyectos/1/backlog`).

Respuestas reales tomadas de la red del navegador:

```text
POST /api/poker-sessions/1/reveal → 200
{"id":1,"item_id":1,"estado":"REVELADA","ronda":1,"votos":[{"integrante":"Ana","puntos":5},{"integrante":"Bruno","puntos":8}],"cartas":[1,2,3,5,8,13,21]}
POST /api/poker-sessions/1/estimate → 200
{"id":1,"item_id":1,"estado":"CERRADA","ronda":1,"votos":[{"integrante":"Ana","puntos":5},{"integrante":"Bruno","puntos":8}],"cartas":[1,2,3,5,8,13,21]}
GET /api/proyectos/1/backlog → 200
[{"id":1,"proyecto_id":1,"titulo":"Login de usuarios","descripcion":"Ingreso con usuario y clave","prioridad":"MEDIA","criterios":["Valida la clave"],"estado":"PENDIENTE","story_points":13,"horas_estimadas":0,"sprint_id":0}]
```

Después, con curl contra el mismo servidor:

```text
$ curl -s http://localhost:8080/api/items/1
{"id":1,"proyecto_id":1,"titulo":"Login de usuarios","descripcion":"Ingreso con usuario y clave","prioridad":"MEDIA","criterios":["Valida la clave"],"estado":"PENDIENTE","story_points":13,"horas_estimadas":0,"sprint_id":0}
[HTTP 200 application/json]

$ curl -s -X POST http://localhost:8080/api/poker-sessions/1/estimate -d '{"puntos":8}'
{"error":"la sesión de Planning Poker no está revelada"}
[HTTP 409 application/json]

$ curl -s -i http://localhost:8080/api/poker-sessions/1/estimate
HTTP/1.1 405 Method Not Allowed
Allow: POST
Content-Type: text/plain; charset=utf-8

$ curl -s http://localhost:8080/api/items/1      (después del 409)
{… "story_points":13 …}
[HTTP 200]
```

En un primer intento, el `server.exe` no pudo abrir el puerto (`listen tcp :8080: bind: Only one usage of each socket address`) porque otro proceso, que terminó enseguida y no se pudo identificar, atendía el 8080; ese intento se descartó y el recorrido de arriba se hizo desde cero con el `server.exe` de `894c7c2` escuchando en el 8080 (`servidor escuchando en :8080 (almacenamiento: memory)`).

## Validación final

Medido sobre `894c7c2` con go1.27.1 windows/amd64, sin `TEST_DATABASE_URL`. En esta máquina `go test` se colgaba al arrancar los binarios de test que deja en `%TEMP%` (el `-test.timeout` nunca llegaba a dispararse); por eso todas las corridas se hicieron con `GOTMPDIR` en una carpeta temporal propia.

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
ok  	seguimiento-medicion-gpc/cmd/server	1.236s
ok  	seguimiento-medicion-gpc/internal/backlog	1.871s
ok  	seguimiento-medicion-gpc/internal/defect	2.049s
ok  	seguimiento-medicion-gpc/internal/effort	2.005s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.628s
ok  	seguimiento-medicion-gpc/internal/metrics	2.028s
ok  	seguimiento-medicion-gpc/internal/poker	1.619s
ok  	seguimiento-medicion-gpc/internal/postgres	1.837s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.801s
ok  	seguimiento-medicion-gpc/internal/project	2.041s
ok  	seguimiento-medicion-gpc/internal/sprint	1.715s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
go: unlinkat …\gotmp\go-build3383752724\b001\server.test.exe: The process cannot access the file because it is being used by another process.
(exit 1)
```

Todos los paquetes dan `ok`. El código de salida 1 viene solo de la última línea: Windows no dejó borrar el binario temporal `server.test.exe` al terminar, sin relación con los tests (el mismo aviso figura en las evidencias de HU-09, HU-20 y HU-23). Con el mismo código del GREEN, antes del commit, una corrida de `go test -count=1 ./...` terminó con exit 0, y las corridas con base de datos de abajo también.

### `go test -count=1 -cover ./...`

```text
ok  	seguimiento-medicion-gpc/cmd/server	1.043s	coverage: 31.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	1.588s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/defect	1.237s	coverage: 76.6% of statements
ok  	seguimiento-medicion-gpc/internal/effort	1.451s	coverage: 75.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	2.046s	coverage: 98.3% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	1.228s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	1.548s	coverage: 69.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.200s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.959s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	1.432s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	2.120s	coverage: 76.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
go: unlinkat …\server.test.exe: The process cannot access the file because it is being used by another process.
(exit 1, por el mismo aviso)
```

```text
$ go tool cover -func=cov-head.out | grep -E "\bEstimate\b|estimatePokerSession|statusFor|NewHandler|total:"
seguimiento-medicion-gpc/internal/httpapi/poker.go:128:			estimatePokerSession		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:50:			NewHandler			100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:130:		statusFor			100.0%
seguimiento-medicion-gpc/internal/poker/manager.go:98:			Estimate			100.0%
total:									(statements)			80.6%
```

Comparado con `main` (`82e9ed7`), medido con el mismo Go y el mismo comando en una copia temporal (`git worktree`): `internal/poker` sube de 65.2% a 69.5%, `internal/httpapi` de 98.2% a 98.3%, `cmd/server` se mantiene en 31.5% y el total del perfil sube de 80.0% a 80.6%. Sin base, `postgres_repository.go` de poker no se ejecuta porque sus tests de integración se omiten.

### `-race`

`-race` no se corrió localmente: `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1` (`CGO_ENABLED=0` y sin `gcc`). El test de concurrencia (`TestEstimate_SimultaneosSobreLaMismaSesionSoloUnoSeAplica`, 20 intentos) y el escenario de CA-25.8 corrieron sin `-race` y pasaron; con `-race` los ejecuta el CI del Pull Request (Linux, `go test -race -coverprofile=coverage.out ./...`).

### Con PostgreSQL

Base recreada con `docker compose down -v` y `docker compose up -d --wait` (imagen `postgres:17-alpine` del `docker-compose.yml` del repo, puerto 5433). Versión real:

```text
$ docker compose exec -T postgres psql -U gpc -d gpc_test -tAc "SELECT version()"
PostgreSQL 17.11 on x86_64-pc-linux-musl, compiled by gcc (Alpine 15.2.0) 15.2.0, 64-bit
```

Con `TEST_DATABASE_URL=postgres://gpc:gpc_dev@localhost:5433/gpc_test?sslmode=disable`, sobre `894c7c2`:

```text
$ go test -count=1 -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	3.330s	coverage: 33.3% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	6.006s	coverage: 99.0% of statements
ok  	seguimiento-medicion-gpc/internal/defect	6.476s	coverage: 95.3% of statements
ok  	seguimiento-medicion-gpc/internal/effort	6.758s	coverage: 97.3% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.925s	coverage: 98.3% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	1.599s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	5.708s	coverage: 94.2% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.358s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.397s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	5.477s	coverage: 96.8% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	6.363s	coverage: 97.1% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
(exit 0)
```

La primera corrida de ese comando terminó con exit 1: `internal/metrics` no llegó a compilar (`link.exe: cannot open file …: Insufficient system resources exist to complete the requested service.`, con Docker Desktop recién iniciado) y apareció el aviso `unlinkat`; los demás paquetes dieron `ok`. La salida de arriba es la de la corrida repetida.

```text
$ go test -count=1 -v -run Contrato ./internal/poker
=== RUN   TestPostgresRepository_Contrato
…
=== RUN   TestPostgresRepository_Contrato/Update_guarda_el_estado_CERRADA_y_GetByID_lo_devuelve_con_sus_votos
--- PASS: TestPostgresRepository_Contrato (0.90s)
    …
    --- PASS: TestPostgresRepository_Contrato/Update_guarda_el_estado_REVELADA_y_GetByID_lo_devuelve_con_sus_votos (0.06s)
    --- PASS: TestPostgresRepository_Contrato/Update_guarda_el_estado_CERRADA_y_GetByID_lo_devuelve_con_sus_votos (0.07s)
--- PASS: TestMemoryRepository_Contrato (0.00s)
    …
    --- PASS: TestMemoryRepository_Contrato/Update_guarda_el_estado_CERRADA_y_GetByID_lo_devuelve_con_sus_votos (0.00s)
PASS
ok  	seguimiento-medicion-gpc/internal/poker	2.316s
(exit 0)
```

`TestPostgresRepository_Contrato` corrió sus 12 subtests contra PostgreSQL, sin SKIP.

- `go test -count=1 -v ./...`: exit 0, 1336 PASS, 0 SKIP y 0 FAIL (conteo de `--- PASS`, `--- SKIP` y `--- FAIL` en la salida).
- `go test -count=1 -coverpkg=./... -coverprofile=cov-db.out ./...`: exit 0, total **94.8%**. `Estimate`, `estimatePokerSession` y `statusFor` quedan en 100.0%; `Update` de PostgreSQL de poker en 69.2% y `GetByID` en 85.7%, por ramas de error de la base que no se pueden provocar con la base funcionando.

## Commits

Los cuatro commits (SDD, RED, GREEN y documentación) referencian el Issue #43 con `Refs #43` y están en la rama `feature/HU-25-registrar-estimacion-acordada`.

HU-25: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
