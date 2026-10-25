# Evidencia TDD — HU-27 Consultar los registros de esfuerzo

> Issue del repositorio: #45

- **Historia:** HU-27 — Consultar los registros de esfuerzo
- **Responsable:** García Santiago
- **Rama:** `feature/HU-27-consultar-esfuerzos`
- **Fecha:** 25/10/2026
- **Go:** go1.27.2 windows/amd64. El entorno local no tiene cgo, así que `go test -race` se verifica en el CI y en la corrida con base de Hernán.
- **Especificación SDD:** [`docs/especificaciones/HU-27-consultar-esfuerzos.md`](../especificaciones/HU-27-consultar-esfuerzos.md), aprobada por Peñalbé Hernán (Agile Enabler) el 25/10/2026.
- **Escenarios BDD:** [`features/HU-27-consultar-esfuerzos.feature`](../../features/HU-27-consultar-esfuerzos.feature) (11 escenarios, 100 pasos, ejecutados con godog).

## Trazabilidad

HU-27: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**26 tests nuevos (funciones `Test`) cubriendo 11 escenarios BDD (100 pasos), más 7 subtests de contrato y 2 filas nuevas en la tabla de rutas.** La correspondencia no es 1 a 1: CA-27.10 se cubre solo con tests de Go y varios tests comparten escenarios.

| Criterio de aceptación | Escenarios BDD | Tests |
|---|---|---|
| CA-27.1 Listar todos los esfuerzos de una historia o de un sprint, ordenados por ID | "Listar los esfuerzos de un sprint con su total", "Listar los esfuerzos de una historia con su total" | `TestListBySprint_DevuelveLosEsfuerzosConElTotal`, `TestListByItem_DevuelveLosEsfuerzosConElTotal`, los subtests de orden por ID del contrato |
| CA-27.2 Total exacto y un decimal en la presentación | Los mismos dos | `TestListByItem_TotalExacto`, `TestListarEsfuerzos_RedondeaElTotalAUnDecimal`, los dos escenarios de listado |
| CA-27.3 Lista vacía, no `nil`, y total 0 | "Sprint existente sin esfuerzos", "Historia existente sin esfuerzos" | `TestListBySprint_SinEsfuerzosDevuelveListaVacia`, `TestListByItem_SinEsfuerzosDevuelveListaVacia`, `TestListarEsfuerzos_SinRegistrosDevuelveListaVaciaYTotalCero`, los subtests de lista vacía del contrato |
| CA-27.4 Historia o sprint inexistente (incluidos 0 y negativos) | "Rechazar un sprint y una historia inexistentes" | `TestListBySprint_SprintInexistente` (999, 0 y -1), `TestListByItem_ItemInexistente` (999, 0 y -1), `TestListarEsfuerzos_InexistenteDevuelve404` |
| CA-27.5 Aislamiento entre historias y sprints | "Aislar los esfuerzos por sprint", "Aislar los esfuerzos por historia" | `TestListBySprint_AislaLosEsfuerzosDeOtroSprint`, `TestListByItem_AislaLosEsfuerzosDeOtraHistoria`, los subtests de aislamiento del contrato |
| CA-27.6 Sprints abiertos y cerrados consultables | "Consultar un sprint cerrado" | `TestListBySprint_SprintCerradoSePuedeConsultar` |
| CA-27.7 Copias por valor | "Modificar lo devuelto no altera lo guardado" | `TestListByItem_CopiasIndependientes`, el subtest "Las listas devuelven copias" |
| CA-27.8 Listados simultáneos con registros | "Listados simultáneos con registros" | `TestListByItem_ConcurrenteConRegister`, verificado con `-race` |
| CA-27.9 Error del repositorio propagado | "Error del repositorio al listar" | `TestListBySprint_ErrorDelRepositorio`, `TestListByItem_ErrorDelRepositorio`, `TestListarEsfuerzos_ErrorDelRepositorioDevuelve500SinDetalle` |
| CA-27.10 API y formulario web | — (se cubre con tests de Go) | `TestListarEsfuerzosPorSprint_Devuelve200ConLosEsfuerzosYElTotal`, `TestListarEsfuerzosPorItem_Devuelve200ConLosEsfuerzosYElTotal`, `TestListarEsfuerzos_SinRegistrosDevuelveListaVaciaYTotalCero`, `TestListarEsfuerzos_InexistenteDevuelve404`, `TestListarEsfuerzos_ErrorDelRepositorioDevuelve500SinDetalle`, `TestListarEsfuerzos_RedondeaElTotalAUnDecimal`, 2 filas de `TestRutasConID_IDNoNumericoDevuelve400`, `TestInterfaz_FormularioDeConsultaDeEsfuerzos`, `TestInterfaz_TablaDeEsfuerzosTieneLasColumnas`, `TestInterfaz_ConsultaDeEsfuerzosMuestraElError` |

## Alcance de las pruebas

- **BDD:** `features/HU-27-consultar-esfuerzos.feature` tiene **11 escenarios** (100 pasos). Cada escenario lleva su CA en el nombre.
- **Tests:** **26 funciones `Test` nuevas** (incluido el runner de godog), **7 subtests de contrato** nuevos y **2 filas** nuevas en `TestRutasConID_IDNoNumericoDevuelve400`:

| Archivo | Tests nuevos |
|---|---|
| `internal/effort/list_test.go` | 14 |
| `internal/effort/list_features_test.go` | 1 (runner de godog) |
| `internal/effort/postgres_repository_test.go` | 2 |
| `internal/effort/repository_contract_test.go` | 7 subtests |
| `internal/httpapi/effort_lists_test.go` | 6 |
| `internal/httpapi/web_test.go` | 3 |
| `internal/httpapi/server_test.go` | 2 filas de tabla |

No hay una correspondencia 1 a 1: son 26 tests para 11 escenarios, y CA-27.10 se cubre solo con tests de Go.

## Fase RED (`13572f9`)

El commit RED incluye la SDD, el `.feature` y los tests. Para que el paquete compile se agregaron stubs mínimos sin lógica, así que los tests fallan por sus aserciones y no por compilación:

- `internal/effort/list.go`: `Summary` y los stubs de `ListByItem` y `ListBySprint`, que devuelven `errNotImplemented`.
- `internal/effort/memory_repository.go` y `postgres_repository.go`: `ListByItem` y `ListBySprint` devuelven `nil, nil`.
- `internal/effort/repository.go`: los dos métodos nuevos de la interfaz.
- `internal/httpapi/server.go`: las rutas GET todavía no están registradas, por eso la API responde 404.
- `internal/httpapi/efforts.go`: los handlers nuevos todavía no existen.
- `internal/web/index.html` y `app.js`: el formulario y la tabla de consulta todavía no existen.
- `internal/effort/register_test.go` y `internal/httpapi/efforts_test.go`: se agregó al fake de test el método que faltaba para que compilara.

`gofmt -l .` no devuelve archivos y `go vet ./...` pasa antes del commit RED.

Sin base (se muestran las líneas de resultado y de fallo; se omite el detalle repetido de godog):

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	2.114s
ok  	seguimiento-medicion-gpc/internal/backlog	2.746s
ok  	seguimiento-medicion-gpc/internal/defect	2.429s
11 scenarios (11 failed)
100 steps (56 passed, 11 failed, 33 skipped)
--- FAIL: TestFeatures_HU27ConsultarEsfuerzos (0.04s)
    --- FAIL: .../Rechazar_un_sprint_y_una_historia_inexistentes_(CA-27.4) (0.00s)
        suite.go:640: error = no implementado, se esperaba no existe un sprint con ese id
    --- FAIL: .../Error_del_repositorio_al_listar_(CA-27.9) (0.00s)
        suite.go:640: error = no implementado, se esperaba fallo del repositorio de esfuerzos
    list_features_test.go:33: fallaron escenarios BDD de HU-27
--- FAIL: TestListBySprint_DevuelveLosEsfuerzosConElTotal (0.00s)
    list_test.go:41: ListBySprint devolvió un error inesperado: no implementado
--- FAIL: TestListByItem_DevuelveLosEsfuerzosConElTotal (0.00s)
    list_test.go:67: ListByItem devolvió un error inesperado: no implementado
--- FAIL: TestListBySprint_SinEsfuerzosDevuelveListaVacia (0.00s)
    list_test.go:87: ListBySprint devolvió un error inesperado: no implementado
--- FAIL: TestListByItem_SinEsfuerzosDevuelveListaVacia (0.00s)
    list_test.go:107: ListByItem devolvió un error inesperado: no implementado
--- FAIL: TestListBySprint_AislaLosEsfuerzosDeOtroSprint (0.00s)
    list_test.go:131: ListBySprint devolvió un error inesperado: no implementado
--- FAIL: TestListByItem_AislaLosEsfuerzosDeOtraHistoria (0.00s)
    list_test.go:152: ListByItem devolvió un error inesperado: no implementado
--- FAIL: TestListBySprint_SprintCerradoSePuedeConsultar (0.00s)
    list_test.go:171: ListBySprint devolvió un error inesperado: no implementado
--- FAIL: TestListBySprint_SprintInexistente (0.00s)
    list_test.go:186: err = no implementado, se esperaba no existe un sprint con ese id
--- FAIL: TestListByItem_ItemInexistente (0.00s)
    list_test.go:203: err = no implementado, se esperaba no existe un ítem del backlog con ese id
--- FAIL: TestListBySprint_ErrorDelRepositorio (0.00s)
    list_test.go:220: err = no implementado, se esperaba fallo del repositorio de esfuerzos
--- FAIL: TestListByItem_ErrorDelRepositorio (0.00s)
    list_test.go:235: err = no implementado, se esperaba fallo del repositorio de esfuerzos
--- FAIL: TestListByItem_TotalExacto (0.00s)
    list_test.go:251: ListByItem devolvió un error inesperado: no implementado
--- FAIL: TestListByItem_CopiasIndependientes (0.00s)
    list_test.go:265: ListByItem devolvió un error inesperado: no implementado
--- FAIL: TestMemoryRepository_Contrato (0.00s)
    --- FAIL: .../ListByItem_devuelve_una_lista_vacía,_no_nil (0.00s)
        repository_contract_test.go:70: la lista es nil, se esperaba una lista vacía
    --- FAIL: .../ListBySprint_devuelve_los_esfuerzos_del_sprint_ordenados_por_ID (0.00s)
        repository_contract_test.go:124: esfuerzos = 0, se esperaban 2
    --- FAIL: .../Las_listas_devuelven_copias (0.00s)
        repository_contract_test.go:170: esfuerzos = 0, se esperaba 1
FAIL
FAIL	seguimiento-medicion-gpc/internal/effort	2.601s
--- FAIL: TestListarEsfuerzosPorSprint_Devuelve200ConLosEsfuerzosYElTotal (0.00s)
    effort_lists_test.go:34: GET /api/sprints/1/esfuerzos: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    effort_lists_test.go:36: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestListarEsfuerzosPorItem_Devuelve200ConLosEsfuerzosYElTotal (0.00s)
    effort_lists_test.go:57: GET /api/items/1/esfuerzos: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    effort_lists_test.go:59: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestListarEsfuerzos_SinRegistrosDevuelveListaVaciaYTotalCero (0.00s)
    effort_lists_test.go:75: GET /api/sprints/1/esfuerzos: código = 404, se esperaba 200
--- FAIL: TestListarEsfuerzos_InexistenteDevuelve404 (0.00s)
    effort_lists_test.go:105: GET /api/sprints/999/esfuerzos: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
--- FAIL: TestListarEsfuerzos_ErrorDelRepositorioDevuelve500SinDetalle (0.00s)
    effort_lists_test.go:119: código = 404, se esperaba 500
--- FAIL: TestListarEsfuerzos_RedondeaElTotalAUnDecimal (0.00s)
    effort_lists_test.go:140: GET /api/items/1/esfuerzos: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.00s)
    --- FAIL: .../listar_esfuerzos_del_ítem (0.00s)
        server_test.go:217: GET /api/items/abc/esfuerzos: código = 404, se esperaba 400
    --- FAIL: .../listar_esfuerzos_del_sprint (0.00s)
        server_test.go:217: GET /api/sprints/abc/esfuerzos: código = 404, se esperaba 400
--- FAIL: TestInterfaz_FormularioDeConsultaDeEsfuerzos (0.00s)
    web_test.go:278: la página no contiene "id=\"form-consultar-esfuerzos\""
    web_test.go:293: app.js no contiene "accion(\"msg-esfuerzos\""
--- FAIL: TestInterfaz_TablaDeEsfuerzosTieneLasColumnas (0.00s)
    web_test.go:302: la página no contiene el formulario de consulta de esfuerzos
--- FAIL: TestInterfaz_ConsultaDeEsfuerzosMuestraElError (0.00s)
    web_test.go:322: app.js no muestra los errores de la consulta con accion
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi	1.863s
ok  	seguimiento-medicion-gpc/internal/metrics	1.891s
ok  	seguimiento-medicion-gpc/internal/poker	2.512s
ok  	seguimiento-medicion-gpc/internal/postgres	1.803s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.841s
ok  	seguimiento-medicion-gpc/internal/project	2.192s
ok  	seguimiento-medicion-gpc/internal/sprint	2.252s
FAIL
```

Fallan los 15 tests nuevos de `internal/effort` que no necesitan base (14 unitarios y el runner BDD) y los 9 tests nuevos de `internal/httpapi`; además falla el `TestMemoryRepository_Contrato` ya existente por sus 7 subtests nuevos y la tabla de rutas por sus 2 filas nuevas. Los 2 tests nuevos de PostgreSQL se omiten sin `TEST_DATABASE_URL`. Los paquetes ajenos a la historia pasan.

## Fase GREEN (`9e41d24`)

El commit GREEN solo cambia código de producción y el README; no modifica ningún archivo `*_test.go` ni el `.feature`. Verificado con `git diff --stat 13572f9..9e41d24`.

- `internal/effort/list.go`: `ListByItem` valida la historia con `items.GetItem`; `ListBySprint` valida el sprint con `sprints.Get`; los dos llaman al repositorio y `summarize` copia la lista y suma las horas sin redondear.
- `internal/effort/memory_repository.go`: `sync.Mutex` cubre `Save`, `ListByItem` y `ListBySprint`; los listados devuelven una lista nueva, no `nil`.
- `internal/effort/postgres_repository.go`: `ListByItem` y `ListBySprint` delegan en `listEfforts`, que usa `ORDER BY id` y una columna de filtro interna, nunca un valor del llamador.
- `internal/httpapi/efforts.go` y `server.go`: los dos handlers GET, sus rutas y `toEffortListResponse`, que redondea solo el total a un decimal.
- `internal/web/index.html` y `app.js`: formulario "Consultar esfuerzos" por sprint o historia, tabla con las columnas, total y mensajes de error.
- `README.md`: las dos rutas en la tabla de la API y la consulta en la descripción de la interfaz.
- `internal/effort/effort.go`: se eliminó `errNotImplemented`, que solo existía para que compilaran los stubs del RED, y `gofmt` realineó las variables de error de HU-26 sin cambiar su contenido. `Summary` se quedó en `list.go`, como define la SDD.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.883s
ok  	seguimiento-medicion-gpc/internal/backlog	1.365s
ok  	seguimiento-medicion-gpc/internal/defect	1.118s
ok  	seguimiento-medicion-gpc/internal/effort	1.083s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.964s
ok  	seguimiento-medicion-gpc/internal/metrics	0.704s
ok  	seguimiento-medicion-gpc/internal/poker	0.957s
ok  	seguimiento-medicion-gpc/internal/postgres	0.626s
ok  	seguimiento-medicion-gpc/internal/postgrestest	2.031s
ok  	seguimiento-medicion-gpc/internal/project	0.996s
ok  	seguimiento-medicion-gpc/internal/sprint	1.056s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

godog: `11 scenarios (11 passed)` y `100 steps (100 passed)`.

## Corrida con PostgreSQL (Peñalbé Hernán, 25/10/2026)

Sobre 9e41d24, en Linux con go1.27.1 y PostgreSQL 16, base gpc_test recreada con db/init/01-schema.sql:

```
$ TEST_DATABASE_URL=... go test -race -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server
ok  	seguimiento-medicion-gpc/internal/backlog
ok  	seguimiento-medicion-gpc/internal/defect
ok  	seguimiento-medicion-gpc/internal/effort
ok  	seguimiento-medicion-gpc/internal/httpapi
ok  	seguimiento-medicion-gpc/internal/metrics
ok  	seguimiento-medicion-gpc/internal/poker
ok  	seguimiento-medicion-gpc/internal/postgres
ok  	seguimiento-medicion-gpc/internal/postgrestest
ok  	seguimiento-medicion-gpc/internal/project
ok  	seguimiento-medicion-gpc/internal/sprint
?   	seguimiento-medicion-gpc/internal/web  [no test files]
```

Sin SKIP (0 tests omitidos) y sin carreras.

```
--- PASS: TestPostgresRepository_Contrato
--- PASS: TestPostgresRepository_GuardaElEsfuerzo
--- PASS: TestPostgresRepository_NormalizaLaFechaAUTC
--- PASS: TestPostgresRepository_SaveDeItemInexistenteFalla
--- PASS: TestPostgresRepository_SaveDeSprintInexistenteFalla
--- PASS: TestPostgresRepository_ListaPorItem
--- PASS: TestPostgresRepository_ListaPorSprint
--- PASS: TestMemoryRepository_Contrato
```

Cobertura con base:

- `internal/effort`: 97.3%.
- `ListByItem`, `ListBySprint`, `summarize` y los listados en memoria y PostgreSQL: 100.0%, salvo `listEfforts` de PostgreSQL (81.2%, ramas de error de la base).
- Total con `-coverpkg=./...`: 94.9%.

## Refactor

No se hizo: el código nuevo reutiliza los patrones de `internal/effort` (manager con mutex, repositorios detrás de la interfaz) y de `internal/httpapi` (`pathID`, `writeJSON`, `writeError`, `statusFor`). No hay lógica duplicada ni ramas nuevas sin test que justifiquen un refactor separado.

## Desviaciones del ciclo TDD

Ninguna. El RED compila con stubs y falla por aserciones; el GREEN no modificó tests ni escenarios. El único cambio no funcional en producción fue eliminar `errNotImplemented` y el realineo de `gofmt` en `effort.go`.

## Validación final

```
$ gofmt -l .
(sin salida)
```

```
$ go vet ./...
(sin salida)
```

```
$ go test -count=1 -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.845s	coverage: 50.0% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	1.726s	coverage: 79.9% of statements
ok  	seguimiento-medicion-gpc/internal/defect	1.504s	coverage: 70.0% of statements
ok  	seguimiento-medicion-gpc/internal/effort	1.095s	coverage: 71.6% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.029s	coverage: 98.8% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	0.946s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	1.007s	coverage: 51.9% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	0.701s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.708s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	1.365s	coverage: 64.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	1.736s	coverage: 72.8% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
$ go test -count=1 -coverprofile=cover.out ./... && go tool cover -func=cover.out | tail -n 1
total:									(statements)			79.4%
```

`internal/effort` baja de 74.4% en `main` (`b876e6f`, mismo Go) a 71.6% sin base porque `postgres_repository.go` no se ejecuta sin `TEST_DATABASE_URL`; con la base sube a 97.3%. `internal/httpapi` sube de 98.7% a 98.8%. El total sin base se mantiene en **79.4%**, igual que `main` (`b876e6f`, medido con el mismo Go en una copia temporal). Con base y `-coverpkg=./...`, el total es **94.9%**.

Comparación con `main` (`b876e6f`) medida con el mismo Go 1.27.2 en una copia temporal (`git worktree`):

| Paquete | HU-27 (sin base) | `main` `b876e6f` |
|---|---:|---:|
| `cmd/server` | 50.0% | 50.0% |
| `internal/backlog` | 79.9% | 79.9% |
| `internal/defect` | 70.0% | 70.0% |
| `internal/effort` | 71.6% | 74.4% |
| `internal/httpapi` | 98.8% | 98.7% |
| `internal/metrics` | 100.0% | 100.0% |
| `internal/poker` | 51.9% | 51.9% |
| `internal/web` | 0.0% | 0.0% |
| `internal/postgres` | 83.3% | 83.3% |
| `internal/postgrestest` | 36.8% | 36.8% |
| `internal/project` | 64.4% | 64.4% |
| `internal/sprint` | 72.8% | 72.8% |
| **Total** | **79.4%** | **79.4%** |

### go test -race (CA-27.8)

En este entorno (Windows, `CGO_ENABLED=0`, sin compilador C) `-race` no está disponible. El test de concurrencia (`TestListByItem_ConcurrenteConRegister`) pasa sin `-race`. Hernán corrió la suite completa con `-race` y PostgreSQL 16 sobre `9e41d24`, sin SKIP y sin carreras; el CI del Pull Request también ejecuta `go test -race ./...`.

## Límite conocido

El dominio devuelve la suma exacta de las horas. La API y la interfaz redondean solo la presentación del total a un decimal, como define TEC-03: un total de `0.001` se muestra como `0.0`, pero las horas de cada registro y el total usado por las métricas conservan su valor exacto.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED (incluye SDD y BDD) | 13572f9 | `test(HU-27): agregar tests (RED)` |
| GREEN | 9e41d24 | `feat(HU-27): implementar la consulta de esfuerzos (GREEN)` |
| Refactor | — | No hubo |

HU-27: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
