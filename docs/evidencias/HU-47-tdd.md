# Evidencia TDD · HU-47 Calcular los defectos detectados y resueltos

> Issue del repositorio: #128

- **Especificación SDD:** [`docs/especificaciones/HU-47-defectos-detectados-resueltos.md`](../especificaciones/HU-47-defectos-detectados-resueltos.md), aprobada por Peñalbé Hernán el 27/10/2026 y ajustada el 28/10/2026, antes del RED, a la firma de `metrics.NewManager` de HU-33.
- **Escenarios BDD:** [`features/HU-47-defectos-detectados-resueltos.feature`](../../features/HU-47-defectos-detectados-resueltos.feature): 10 definiciones (7 escenarios y 3 esquemas, con 2, 3 y 3 ejemplos), que godog ejecuta como 15 escenarios y 67 pasos.
- **Rama:** `feature/HU-47-defectos-detectados-resueltos`.
- **Story Points:** 2 (Planning del Sprint 4).
- **Dependencias:** HU-29, HU-30 y HU-31 (lectura de defectos `ListByProject`), integradas en `main` antes del RED.

## Trazabilidad

**20 tests nuevos (funciones `Test`) cubriendo 15 escenarios BDD ejecutados (10 definiciones):** 15 unitarios en `internal/metrics/defects_test.go`, el runner de godog en `internal/metrics/defects_features_test.go`, 3 de la API en `internal/httpapi/metrics_test.go` y 1 de la interfaz en `internal/httpapi/web_test.go`. La correspondencia no es 1 a 1: CA-47.6 se cubre solo con tests de Go y varios tests comparten criterio.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-47.1 Detectados y resueltos de un sprint | "Defectos detectados y resueltos de un sprint" | `TestSprintDefects_CuentaDetectadosYResueltosDelSprint`, `TestSprintDefects_NoCuentaDefectosDeOtroSprint`, `TestSprintDefects_SprintCerradoSeCalculaIgual`, `TestDefects_DefectoDeHistoriaTerminadaCuenta` |
| CA-47.2 Detectados y resueltos de un proyecto | "Defectos detectados y resueltos del proyecto", "Los defectos de otro proyecto no cuentan" | `TestProjectDefects_CuentaTodosLosDefectosDelProyecto`, `TestProjectDefects_NoCuentaDefectosDeOtroProyecto`, `TestDefects_DefectoDeHistoriaTerminadaCuenta` |
| CA-47.3 Sin defectos, cero | "Un sprint sin defectos da cero", "Un proyecto sin defectos da cero" | `TestSprintDefects_SinDefectosDaCero`, `TestProjectDefects_SinDefectosDaCero`, `TestMetricas_SinDefectosLosCamposValenCero` |
| CA-47.4 Solo cuenta el estado resuelto; detección y resolución en sprints distintos | "Un defecto cuenta en su sprint de detección y en el de resolución" (2), "Un defecto en corrección no cuenta como resuelto aunque tenga sprint de resolución" | `TestDefects_DeteccionYResolucionEnSprintsDistintos`, `TestDefects_SoloCuentaComoResueltoElEstadoResuelto` |
| CA-47.5 Inexistente da error y no modifica nada | "Consultar un sprint inexistente falla" (3), "Consultar un proyecto inexistente falla" (3), "Consultar los defectos no los modifica" | `TestSprintDefects_SprintInexistenteDevuelveErrSprintNotFound`, `TestProjectDefects_ProyectoInexistenteDevuelveErrProjectNotFound`, `TestDefects_ConsultarNoModificaNada` |
| Condiciones de error de la SDD (falla al leer) | — | `TestDefects_ErrorAlLeerLosDefectosSePropaga`, `TestSprintDefects_ErrorAlLeerElSprintSePropaga`, `TestMetricas_ErrorDelRepositorioDeDefectosDevuelve500` |
| CA-47.6 API e interfaz | — | `TestMetricas_IncluyenLosDefectosDetectadosYResueltos`, `TestMetricas_SinDefectosLosCamposValenCero`, `TestInterfaz_MetricasMuestraLosDefectosDetectadosYResueltos`; los 400 y 404 de esas rutas los cubren los tests de HU-32 y `TestRutasConID_IDNoNumericoDevuelve400` |

## Antes del RED

- **Merge de `main` (`d9dee72`):** la rama traía solo la SDD y el BDD (`ba2a548`); se trajo `main` (`0ee9b20`) con HU-30, HU-31, HU-33 y HU-36, sin conflictos.
- **Ajuste de la SDD (`77bcb36`):** HU-33 agregó a `metrics.NewManager` el parámetro de esfuerzos, así que el de defectos pasa a ser el cuarto. No cambian reglas, casos ni criterios.

## Fase RED (`05df3f5`)

El RED agrega los tests y los stubs en `internal/metrics/defects.go`: el tipo `Defects`, la interfaz `defectLister` y `SprintDefects` y `ProjectDefects`, que devuelven `errNotImplemented`. Se eligió un error en lugar de un `Defects` vacío y `nil` para que los casos que esperan cero también fallen en el RED (en HU-34, el stub vacío dejó pasar los casos "sin ítems"). Como la SDD lo indica, el RED también agrega el cuarto parámetro de `NewManager`, su llamada en `internal/httpapi/server.go` y el helper `newMetricsEnvWith` de los tests de `metrics`, para que todo compile y el GREEN no toque tests.

Sobre `05df3f5`, en Linux con go1.27.1, sin base:

```
$ go vet ./...
(sin salida)

$ go test -count=1 ./...
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
--- FAIL: TestMetricas_IncluyenLosDefectosDetectadosYResueltos (0.00s)
    metrics_test.go:508: defectos_detectados = <nil>, se esperaba 2; cuerpo: {"sprint_id":1,...,"items_sin_horas_estimadas":0}
    metrics_test.go:508: defectos_resueltos = <nil>, se esperaba 1; cuerpo: {"sprint_id":1,...}
    metrics_test.go:509: defectos_detectados = <nil>, se esperaba 2; cuerpo: {"proyecto_id":1,...}
    metrics_test.go:509: defectos_resueltos = <nil>, se esperaba 1; cuerpo: {"proyecto_id":1,...}
--- FAIL: TestMetricas_SinDefectosLosCamposValenCero (0.00s)
    metrics_test.go:516: defectos_detectados = <nil>, se esperaba 0; cuerpo: {"sprint_id":1,...}
    metrics_test.go:516: defectos_resueltos = <nil>, se esperaba 0; cuerpo: {"sprint_id":1,...}
    metrics_test.go:517: defectos_detectados = <nil>, se esperaba 0; cuerpo: {"proyecto_id":1,...}
    metrics_test.go:517: defectos_resueltos = <nil>, se esperaba 0; cuerpo: {"proyecto_id":1,...}
--- FAIL: TestMetricas_ErrorDelRepositorioDeDefectosDevuelve500 (0.00s)
    metrics_test.go:528: código = 200, se esperaba 500; cuerpo: {"sprint_id":1,...}
--- FAIL: TestInterfaz_MetricasMuestraLosDefectosDetectadosYResueltos (0.00s)
    web_test.go:757: app.js no contiene "defectos_detectados"
    web_test.go:757: app.js no contiene "defectos_resueltos"
    web_test.go:757: app.js no contiene "\"Defectos: \""
    web_test.go:757: app.js no contiene "\" detectados, \""
    web_test.go:757: app.js no contiene "\" resuelto(s).\""
FAIL
FAIL seguimiento-medicion-gpc/internal/httpapi
15 scenarios (15 failed)
67 steps (52 passed, 15 failed)
--- FAIL: TestFeatures_HU47DefectosDetectadosResueltos (0.02s)
--- FAIL: TestSprintDefects_CuentaDetectadosYResueltosDelSprint (0.00s)
    defects_test.go:63: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestSprintDefects_NoCuentaDefectosDeOtroSprint (0.00s)
    defects_test.go:79: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestSprintDefects_SprintCerradoSeCalculaIgual (0.00s)
    defects_test.go:96: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestProjectDefects_CuentaTodosLosDefectosDelProyecto (0.00s)
    defects_test.go:114: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestProjectDefects_NoCuentaDefectosDeOtroProyecto (0.00s)
    defects_test.go:131: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestDefects_DefectoDeHistoriaTerminadaCuenta (0.00s)
    defects_test.go:144: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestSprintDefects_SinDefectosDaCero (0.00s)
    defects_test.go:158: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestProjectDefects_SinDefectosDaCero (0.00s)
    defects_test.go:168: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestDefects_DeteccionYResolucionEnSprintsDistintos (0.00s)
    defects_test.go:182: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestDefects_SoloCuentaComoResueltoElEstadoResuelto (0.00s)
    defects_test.go:201: err = HU-47: sin implementar, se esperaba nil
--- FAIL: TestSprintDefects_SprintInexistenteDevuelveErrSprintNotFound (0.00s)
    defects_test.go:216: SprintDefects(99): err = HU-47: sin implementar, se esperaba no existe un sprint con ese id
    defects_test.go:216: SprintDefects(0): err = HU-47: sin implementar, se esperaba no existe un sprint con ese id
    defects_test.go:216: SprintDefects(-1): err = HU-47: sin implementar, se esperaba no existe un sprint con ese id
--- FAIL: TestProjectDefects_ProyectoInexistenteDevuelveErrProjectNotFound (0.00s)
    defects_test.go:233: ProjectDefects(99): err = HU-47: sin implementar, se esperaba no existe un proyecto con ese id
    defects_test.go:233: ProjectDefects(0): err = HU-47: sin implementar, se esperaba no existe un proyecto con ese id
    defects_test.go:233: ProjectDefects(-1): err = HU-47: sin implementar, se esperaba no existe un proyecto con ese id
--- FAIL: TestDefects_ConsultarNoModificaNada (0.00s)
    defects_test.go:252: SprintDefects(1): HU-47: sin implementar
--- FAIL: TestDefects_ErrorAlLeerLosDefectosSePropaga (0.00s)
    defects_test.go:302: SprintDefects = {Detected:0 Resolved:0}, HU-47: sin implementar; se esperaba el valor cero y base caída
    defects_test.go:305: ProjectDefects = {Detected:0 Resolved:0}, HU-47: sin implementar; se esperaba el valor cero y base caída
--- FAIL: TestSprintDefects_ErrorAlLeerElSprintSePropaga (0.00s)
    defects_test.go:317: SprintDefects = {Detected:0 Resolved:0}, HU-47: sin implementar; se esperaba el valor cero y base caída
FAIL
FAIL seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/sprint
?    seguimiento-medicion-gpc/internal/web [no test files]
FAIL
```

Las respuestas JSON se acortaron con `...`; los mensajes son los de la salida. Los 15 escenarios fallan con "la consulta falló: HU-47: sin implementar" o con el error esperado de "no existe".

Fallan solo tests de HU-47: 16 en `internal/metrics` (los 15 unitarios y el runner de godog, con 15 de 15 escenarios en falla) y 4 en `internal/httpapi`. **Ningún test ni escenario de HU-47 pasa en el RED.** Los escenarios de HU-32 (16), HU-33 (14) y HU-34 (14) siguen en verde, y los demás paquetes pasan.

En `TestMetricas_ErrorDelRepositorioDeDefectosDevuelve500` falla solo la primera ruta porque `expectError` corta el test con `t.Fatalf`; con el GREEN pasan las dos.

## Fase GREEN (`94352f2`)

- `internal/metrics/defects.go`: se saca `errNotImplemented`. `SprintDefects` valida el sprint con `sprint.Manager.Get`, lee los defectos de su proyecto con `ListByProject` (HU-31) y cuenta los detectados en el sprint y los `RESUELTO` con ese sprint de resolución. `ProjectDefects` lee con `ListByProject`, que valida el proyecto, y cuenta todos los defectos y los `RESUELTO`.
- `internal/httpapi/metrics.go`: los dos handlers suman `defectos_detectados` y `defectos_resueltos` con una estructura embebida, `defectsResponse`, y `toDefectsResponse`.
- `internal/web/app.js`: "Ver métricas" agrega "Defectos: N detectados, M resuelto(s)." al proyecto y al sprint actual.

El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat 05df3f5 94352f2 -- '*_test.go' '*.feature'
(sin salida)
```

## Refactor

No hubo. `defects.go` tiene dos funciones cortas sin lógica repetida que justifique extraerla.

## Desviaciones del ciclo TDD

Ninguna. El cuarto parámetro de `NewManager` y su llamada en `server.go` entraron en el RED porque la SDD lo pide, para que el GREEN no toque tests (no es lógica de HU-47: los handlers se cambiaron en el GREEN).

## Validación final

Sobre `94352f2`, en Linux con go1.27.1:

```
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
?    seguimiento-medicion-gpc/internal/web [no test files]

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

godog de HU-47: `15 scenarios (15 passed)`, `67 steps (67 passed)`.

`go tool cover -func`: `SprintDefects`, `ProjectDefects` y `toDefectsResponse` en 100.0%. `sprintMetrics` queda en 90.5% y `projectMetrics` en 92.0%: las dos líneas sin cubrir son las ramas defensivas de HU-34 (segunda lectura del backlog), que ya estaban en `main`; las ramas nuevas de HU-47 están cubiertas por `TestMetricas_ErrorDelRepositorioDeDefectosDevuelve500`. Total del perfil sin base: **82.5%** (`main` en `0ee9b20`, mismo Go y comando, en una copia temporal con `git worktree`: **82.2%**). `internal/metrics` e `internal/httpapi` conservan 100.0% y 98.0%.

### Con base (PostgreSQL 16.15) y `-race`

Base `gpc_test` recreada con `db/init/01-schema.sql`:

```
$ TEST_DATABASE_URL=... go test -race -count=1 -p 1 -coverpkg=./... -coverprofile=cover.out ./...
ok   (todos los paquetes)
$ go tool cover -func=cover.out | tail -n 1
total:                                  (statements)            95.4%
```

Sin SKIP (0 tests omitidos) y sin carreras. HU-47 no toca la persistencia: lee los defectos con la operación de HU-31.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD y BDD | ba2a548 | `docs(HU-47): agregar SDD y escenarios BDD` |
| Merge de `main` | d9dee72 | `chore(HU-47): traer main con HU-30, HU-31, HU-33 y HU-36` |
| Ajuste de la SDD | 77bcb36 | `docs(HU-47): ajustar la SDD a la firma de NewManager de HU-33` |
| RED | 05df3f5 | `test(HU-47): agregar tests (RED)` |
| GREEN | 94352f2 | `feat(HU-47): calcular los defectos detectados y resueltos (GREEN)` |
| Refactor | — | No hubo |
| Evidencia y cobertura | 5693380 | `docs(HU-47): agregar evidencia TDD y cobertura` |

HU-47: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
