# Evidencia TDD · HU-36 Ver el resumen del estado del proyecto

> Issue del repositorio: #54

- **Especificación SDD:** [`docs/especificaciones/HU-36-resumen-proyecto.md`](../especificaciones/HU-36-resumen-proyecto.md), aprobada por Peñalbé Hernán el 28/10/2026, con dos ajustes que se aplicaron (`sprint_actual` con el formato de `sprintResponse` y las lecturas nombradas en RN-36.3 y en "Restricciones").
- **Escenarios BDD:** [`features/HU-36-resumen-proyecto.feature`](../../features/HU-36-resumen-proyecto.feature): 10 definiciones (9 escenarios y 1 esquema con 3 ejemplos), que godog ejecuta como 12 escenarios y 65 pasos.
- **Rama:** `feature/HU-36-resumen-proyecto`.
- **Story Points:** 3, estimados en el Planning del Sprint 4.

## Trazabilidad

**22 tests nuevos (funciones `Test`) cubriendo 12 escenarios BDD ejecutados (10 definiciones)**, más una fila nueva en `TestRutasConID_IDNoNumericoDevuelve400`: 13 unitarios en `internal/dashboard/summary_test.go`, el runner de godog en `internal/dashboard/summary_features_test.go`, 5 de la API en `internal/httpapi/dashboard_test.go` y 3 de la interfaz en `internal/httpapi/web_test.go`. La correspondencia no es 1 a 1: CA-36.6 se cubre solo con tests de Go y varios tests comparten criterio.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-36.1 Proyecto, ítems por estado y sprint actual | "Resumen con ítems en los tres estados y un sprint abierto", "El sprint actual es el abierto aunque haya sprints cerrados", "El resumen no cuenta ítems ni sprints de otro proyecto", "Los estados sin ítems se muestran en cero" | `TestProjectSummary_MuestraProyectoItemsPorEstadoYSprintActual`, `TestProjectSummary_CuentaLosItemsConYSinSprint`, `TestProjectSummary_ItemsEnUnSoloEstadoLosDemasEnCero`, `TestProjectSummary_ElSprintAbiertoEsElActualAunqueHayaCerrados`, `TestProjectSummary_NoCuentaItemsNiSprintsDeOtroProyecto` |
| CA-36.2 Proyecto sin datos, resumen vacío | "Un proyecto sin ítems ni sprints muestra el resumen vacío" | `TestProjectSummary_ProyectoSinDatosDevuelveResumenVacio` |
| CA-36.3 Sin sprint abierto | "Un proyecto con solo sprints cerrados no tiene sprint actual" | `TestProjectSummary_SoloSprintsCerradosNoHaySprintActual` |
| CA-36.4 Coincide con lo guardado | "El resumen refleja el cambio de estado de un ítem", "El resumen refleja el cierre del sprint actual" | `TestProjectSummary_ReflejaElCambioDeEstadoDeUnItem`, `TestProjectSummary_ReflejaElCierreDelSprint` |
| CA-36.5 Inexistente da error y no modifica nada | "Consultar el resumen de un proyecto inexistente devuelve error" (3 ejemplos), "Consultar el resumen no modifica nada" | `TestProjectSummary_ProyectoInexistenteDevuelveErrProjectNotFound`, `TestProjectSummary_ConsultarNoModificaNada`, `TestProjectSummary_ModificarElResumenNoAlteraLoGuardado` (RN-36.7), `TestProjectSummary_ErrorDeLosRepositoriosSePropaga` ("Condiciones de error") |
| CA-36.6 API e interfaz | — | `TestResumenProyecto_Devuelve200ConElResumen`, `TestResumenProyecto_SinSprintAbiertoSprintActualEsNull`, `TestResumenProyecto_ProyectoSinDatosDevuelveCerosYNull`, `TestResumenProyecto_InexistenteDevuelve404`, `TestResumenProyecto_ErrorDelRepositorioDevuelve500SinDetalle`, la fila "resumen del proyecto" de `TestRutasConID_IDNoNumericoDevuelve400`, `TestInterfaz_SeccionDashboardConElBotonVerResumen`, `TestInterfaz_VerResumenLlamaALaAPIConElProyectoActual`, `TestInterfaz_VerResumenMuestraLosItemsPorEstadoYElSprintActual` |

## Fase RED (`34bbb54`)

El RED agrega los tests, el `.feature` y el paquete `internal/dashboard` con las firmas de `Summary`, `Manager`, `NewManager` y `ProjectSummary`, que devuelve un `Summary` vacío y `nil`, para que el paquete compile y los tests fallen por las aserciones (restricción "RED" de la SDD). La ruta, el handler y la interfaz no están en el RED. No cambia ningún test existente, salvo la fila nueva de `TestRutasConID_IDNoNumericoDevuelve400`.

Corrida sobre `34bbb54`, en Windows con go1.27.1 (`go test -count=1`, salida filtrada a los resultados y los mensajes de falla):

```
$ go vet ./...
(sin salida)

$ go test ./...
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
12 scenarios (12 failed)
65 steps (40 passed, 12 failed, 13 skipped)
--- FAIL: TestFeatures_HU36ResumenProyecto (0.04s)
    (los 12 escenarios fallan)
    summary_features_test.go:30: fallaron escenarios BDD de HU-36
--- FAIL: TestProjectSummary_MuestraProyectoItemsPorEstadoYSprintActual (0.00s)
    summary_test.go:202: proyecto = 0 "" , se esperaba 1 "Sistema de ventas" EN_CURSO
    summary_test.go:204: ítems = 0 pendientes, 0 en curso y 0 terminados; se esperaba 2, 1 y 3
    summary_test.go:205: sprint actual = {ID:0 ...} (hay: false), se esperaba {ID:1 ProjectID:1 Name:Sprint 1 ... Status:ABIERTO}
--- FAIL: TestProjectSummary_CuentaLosItemsConYSinSprint (0.00s)
    summary_test.go:217: ítems = 0 pendientes, 0 en curso y 0 terminados; se esperaba 1, 0 y 2
--- FAIL: TestProjectSummary_ItemsEnUnSoloEstadoLosDemasEnCero (0.00s)
    summary_test.go:227: ítems = 0 pendientes, 0 en curso y 0 terminados; se esperaba 0, 2 y 0
--- FAIL: TestProjectSummary_ProyectoSinDatosDevuelveResumenVacio (0.00s)
    summary_test.go:240: resumen = {ProjectID:0 ProjectName: ProjectStatus: ...}, se esperaba {ProjectID:1 ProjectName:Sistema de ventas ProjectStatus:PLANIFICADO ...}
--- FAIL: TestProjectSummary_SoloSprintsCerradosNoHaySprintActual (0.00s)
    summary_test.go:254: proyecto = 0 "", se esperaba 1 "Sistema de ventas"
--- FAIL: TestProjectSummary_ElSprintAbiertoEsElActualAunqueHayaCerrados (0.00s)
    summary_test.go:266: sprint actual = {ID:0 ...} (hay: false), se esperaba {ID:2 ProjectID:1 Name:Sprint 2 ... Status:ABIERTO}
--- FAIL: TestProjectSummary_NoCuentaItemsNiSprintsDeOtroProyecto (0.00s)
    summary_test.go:281: ítems = 0 pendientes, 0 en curso y 0 terminados; se esperaba 1, 0 y 0
--- FAIL: TestProjectSummary_ReflejaElCambioDeEstadoDeUnItem (0.00s)
    summary_test.go:292: ítems = 0 pendientes, 0 en curso y 0 terminados; se esperaba 1, 1 y 0
    summary_test.go:296: ítems = 0 pendientes, 0 en curso y 0 terminados; se esperaba 1, 0 y 1
--- FAIL: TestProjectSummary_ReflejaElCierreDelSprint (0.00s)
    summary_test.go:305: sprint actual = {ID:0 ...} (hay: false), se esperaba {ID:1 ProjectID:1 Name:Sprint 1 ... Status:ABIERTO}
--- FAIL: TestProjectSummary_ProyectoInexistenteDevuelveErrProjectNotFound (0.00s)
    summary_test.go:322: ProjectSummary(99): error = <nil>, se esperaba no existe un proyecto con ese id
    summary_test.go:322: ProjectSummary(0): error = <nil>, se esperaba no existe un proyecto con ese id
    summary_test.go:322: ProjectSummary(-1): error = <nil>, se esperaba no existe un proyecto con ese id
--- FAIL: TestProjectSummary_ConsultarNoModificaNada (0.00s)
    summary_test.go:346: ítems = 0 pendientes, 0 en curso y 0 terminados; se esperaba 0, 1 y 1
--- FAIL: TestProjectSummary_ModificarElResumenNoAlteraLoGuardado (0.00s)
    summary_test.go:365: sprint actual = {ID:0 ...} (hay: false), se esperaba {ID:1 ProjectID:1 Name:Sprint 1 ... Status:ABIERTO}
    summary_test.go:367: nombre del proyecto = "", se esperaba "Sistema de ventas"
--- FAIL: TestProjectSummary_ErrorDeLosRepositoriosSePropaga (0.00s)
    --- FAIL: TestProjectSummary_ErrorDeLosRepositoriosSePropaga/proyecto (0.00s)
    --- FAIL: TestProjectSummary_ErrorDeLosRepositoriosSePropaga/ítems (0.00s)
    --- FAIL: TestProjectSummary_ErrorDeLosRepositoriosSePropaga/sprints (0.00s)
        summary_test.go:420: error = <nil>, se esperaba base caída (en cada uno de los tres subtests)
FAIL
FAIL seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
--- FAIL: TestResumenProyecto_Devuelve200ConElResumen (0.00s)
    dashboard_test.go:72: GET /api/proyectos/1/resumen: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    dashboard_test.go:74: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestResumenProyecto_SinSprintAbiertoSprintActualEsNull (0.00s)
    dashboard_test.go:98: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestResumenProyecto_ProyectoSinDatosDevuelveCerosYNull (0.00s)
    dashboard_test.go:112: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestResumenProyecto_InexistenteDevuelve404 (0.00s)
    dashboard_test.go:128: GET /api/proyectos/99/resumen: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    dashboard_test.go:130: cuerpo no es JSON válido (invalid character 'p' after top-level value): 404 page not found
--- FAIL: TestResumenProyecto_ErrorDelRepositorioDevuelve500SinDetalle (0.00s)
    dashboard_test.go:146: código = 404, se esperaba 500; cuerpo: 404 page not found
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.01s)
    --- FAIL: TestRutasConID_IDNoNumericoDevuelve400/resumen_del_proyecto (0.00s)
        server_test.go:233: código = 404, se esperaba 400; cuerpo: 404 page not found
--- FAIL: TestInterfaz_SeccionDashboardConElBotonVerResumen (0.00s)
    web_test.go:550: la página no contiene la sección del dashboard
--- FAIL: TestInterfaz_VerResumenLlamaALaAPIConElProyectoActual (0.00s)
    web_test.go:575: app.js no contiene "$(\"ver-resumen\")"
    web_test.go:575: app.js no contiene "accion(\"msg-resumen\""
    web_test.go:575: app.js no contiene "proyectoActual() + \"/resumen\""
--- FAIL: TestInterfaz_VerResumenMuestraLosItemsPorEstadoYElSprintActual (0.00s)
    web_test.go:593: app.js no contiene "items_pendientes"
    web_test.go:593: app.js no contiene "items_en_curso"
    web_test.go:593: app.js no contiene "sprint_actual"
    web_test.go:593: app.js no contiene "Sprint actual: "
    web_test.go:593: app.js no contiene "ninguno abierto"
    web_test.go:593: app.js no contiene "$(\"datos-resumen\")"
FAIL
FAIL seguimiento-medicion-gpc/internal/httpapi
ok   seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/sprint
?    seguimiento-medicion-gpc/internal/web  [no test files]
FAIL
```

Fallan 23 tests de primer nivel, todos de HU-36: los 13 unitarios y el runner de godog (12 de 12 escenarios) en `internal/dashboard`, y en `internal/httpapi` los 5 de la API, `TestRutasConID_IDNoNumericoDevuelve400` (solo la fila "resumen del proyecto") y los 3 de la interfaz. Los demás paquetes pasan. Ningún test de HU-36 pasa en el RED (0 PASS en `go test -v ./internal/dashboard/`).

Los de dominio fallan por aserción. Los de la API fallan con 404 "page not found" porque la ruta no existe (el servidor de archivos de la interfaz atiende el pedido). En `TestInterfaz_VerResumenMuestraLosItemsPorEstadoYElSprintActual`, el texto `items_terminados` ya estaba en `app.js` por HU-34, así que esa aserción no falla en el RED; las otras seis sí.

## Fase GREEN (`81f626d`)

- `internal/dashboard/summary.go`: `ProjectSummary` lee el proyecto con `project.Manager.Get` (valida que exista), los ítems con `backlog.Manager.ListItems` y los sprints con `sprint.Manager.ListByProject`; cuenta los ítems en cada estado y toma el sprint `ABIERTO`. Si una lectura falla, devuelve el error y un `Summary` vacío.
- `internal/httpapi/dashboard.go` (nuevo): el handler de `GET /api/proyectos/{id}/resumen` y `summaryResponse`, con `sprint_actual` como `*sprintResponse` armado con `toSprintResponse` de `sprints.go`, o `null`.
- `internal/httpapi/server.go`: crea `dashboard.Manager` junto a `metrics.Manager`, sin cambiar la firma de `NewHandler`, y registra la ruta.
- `internal/web/index.html` y `app.js`: la sección "Dashboard" con el botón "Ver resumen", que muestra el proyecto, los ítems por estado y el sprint actual, o "Sprint actual: ninguno abierto.".

El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat 34bbb54 81f626d -- '*_test.go' '*.feature'
(sin salida)
```

godog sobre `81f626d`: `12 scenarios (12 passed)`, `65 steps (65 passed)`.

`go tool cover -func` (antes del commit GREEN): `NewManager`, `ProjectSummary`, `projectSummary` y `toSummaryResponse` en 100.0%. No queda código nuevo sin un test que lo pida.

### Prueba de punta a punta de la API

Con el servidor real (`go build ./cmd/server`, almacenamiento en memoria, `PORT=8099`):

| Pedido | Respuesta |
|---|---|
| `GET /api/proyectos/1/resumen` (proyecto recién creado) | 200 `{"proyecto_id":1,"nombre":"Sistema de ventas","estado":"PLANIFICADO","items_pendientes":0,"items_en_curso":0,"items_terminados":0,"sprint_actual":null}` |
| Igual, con un sprint abierto, un ítem pendiente y otro en curso | 200 `{…,"items_pendientes":1,"items_en_curso":1,"items_terminados":0,"sprint_actual":{"id":1,"proyecto_id":1,"nombre":"Sprint 1","fecha_inicio":"2026-10-27","fecha_fin":"2026-11-02","objetivo":"Ver el dashboard","estado":"ABIERTO"}}` |
| Igual, después de `POST /api/sprints/1/cerrar` | 200 `{…,"sprint_actual":null}` |
| `GET /api/proyectos/99/resumen` | 404 `{"error":"no existe un proyecto con ese id"}` |
| `GET /api/proyectos/abc/resumen` | 400 `{"error":"el id debe ser un número entero"}` |

## Refactor

No hubo.

## Desviaciones del ciclo TDD

- **Pasos que pasaban en falso en el primer RED (antes del commit).** En la primera corrida del RED, con el stub, pasaban `TestProjectSummary_SoloSprintsCerradosNoHaySprintActual` y tres escenarios ("solo sprints cerrados", "cierre del sprint actual" y "no modifica nada"), porque el `Summary` vacío coincide con "sin sprint actual" y con "no cambió nada". Antes del commit RED se les agregaron aserciones sobre lo que el stub no devuelve: el proyecto, el sprint actual antes del cierre y las cantidades de ítems. El commit `34bbb54` ya tiene esa versión y ninguno pasa.

## Integración con `main` (`4c37d9f`)

Antes de abrir el Pull Request se trajo `main` (`e39c749`, con HU-31 y HU-33 integradas) a la rama con un merge commit, no con rebase, para que los hashes del RED y del GREEN de esta evidencia sigan existiendo (acción 5 de la Retrospectiva del Sprint 3). Hubo un solo conflicto, en `internal/httpapi/server.go`: HU-33 agrega `efforts` a `metrics.NewManager` y HU-36 agrega `dashboard.NewManager`. Se conservaron los dos. `app.js`, `index.html`, `web_test.go` y `server_test.go` se combinaron sin conflicto y se revisaron: están las rutas de HU-31, HU-33 y HU-36, y la sección "Dashboard" queda entre "Métricas" y "Sprints".

## Validación final

Sobre la rama con `main` integrado (`4c37d9f`), en Windows con go1.27.1:

```
$ gofmt -w internal/dashboard/*.go internal/httpapi/dashboard.go internal/httpapi/dashboard_test.go internal/httpapi/server.go internal/httpapi/server_test.go internal/httpapi/web_test.go
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)

$ go test -count=1 ./...
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
ok   seguimiento-medicion-gpc/internal/sprint
?    seguimiento-medicion-gpc/internal/web  [no test files]

$ go test -count=1 -cover ./...
ok   seguimiento-medicion-gpc/cmd/server             coverage: 31.5% of statements
ok   seguimiento-medicion-gpc/internal/backlog       coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/dashboard     coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/defect        coverage: 77.7% of statements
ok   seguimiento-medicion-gpc/internal/effort        coverage: 75.2% of statements
ok   seguimiento-medicion-gpc/internal/httpapi       coverage: 98.0% of statements
ok   seguimiento-medicion-gpc/internal/metrics       coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/poker         coverage: 69.5% of statements
ok   seguimiento-medicion-gpc/internal/postgres      coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/postgrestest  coverage: 36.8% of statements
ok   seguimiento-medicion-gpc/internal/project       coverage: 67.5% of statements
ok   seguimiento-medicion-gpc/internal/sprint        coverage: 76.3% of statements
     seguimiento-medicion-gpc/internal/web           coverage: 0.0% of statements
```

godog HU-36: `12 scenarios (12 passed)`, `65 steps (65 passed)`.

`go tool cover -func`: `NewManager`, `ProjectSummary`, `projectSummary` y `toSummaryResponse` en 100.0%. Total del perfil sin base: **82.2%** (`main` en `e39c749`, mismo Go, medido en una copia temporal con `git worktree`: **81.8%**). `internal/httpapi` sube de 97.9% a 98.0%; los demás paquetes conservan sus porcentajes.

HU-36 no toca la persistencia ni el esquema, así que no se midió con PostgreSQL. `go test -race ./...` no corre localmente (sin cgo) y lo verifica el CI del Pull Request.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | fb9fc25 | `docs(HU-36): agregar especificación SDD` |
| RED | 34bbb54 | `test(HU-36): agregar tests (RED)` |
| GREEN | 81f626d | `feat(HU-36): implementar el resumen del estado del proyecto (GREEN)` |
| Refactor | — | No hubo |
| Integración con `main` | 4c37d9f | `chore(HU-36): traer main con HU-31 y HU-33 y resolver conflictos` |

HU-36: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
