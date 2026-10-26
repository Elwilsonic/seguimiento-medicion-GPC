# Evidencia TDD · HU-34 Calcular el porcentaje de historias completadas

> Issue del repositorio: #52

- **Especificación SDD:** [`docs/especificaciones/HU-34-porcentaje-completadas.md`](../especificaciones/HU-34-porcentaje-completadas.md), aprobada por Peñalbé Hernán el 26/10/2026.
- **Escenarios BDD:** [`features/HU-34-porcentaje-completadas.feature`](../../features/HU-34-porcentaje-completadas.feature): 10 definiciones (8 escenarios y 2 esquemas con 3 ejemplos cada uno), que godog ejecuta como 14 escenarios y 58 pasos.
- **Rama:** `feature/HU-34-porcentaje-completadas`.
- **Story Points:** 2, estimados por el Agile Enabler el 26/10/2026 por tamaño comparable con HU-20 y HU-23 (2 SP cada una); se confirman en el Planning del Sprint 4.

## Trazabilidad

**20 tests nuevos (funciones `Test`) cubriendo 14 escenarios BDD ejecutados (10 definiciones):** 15 unitarios en `internal/metrics/completion_test.go`, el runner de godog en `internal/metrics/completion_features_test.go`, 3 de la API en `internal/httpapi/metrics_test.go` y 1 de la interfaz en `internal/httpapi/web_test.go`. La correspondencia no es 1 a 1: CA-34.6 se cubre solo con tests de Go y varios tests comparten criterio.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-34.1 Porcentaje del sprint | "Porcentaje de un sprint con el ejemplo de TEC-03", "El sprint no cuenta los ítems sin sprint", "Ningún ítem terminado da cero" | `TestSprintCompletion_EjemploDeTEC03`, `TestSprintCompletion_SoloCuentaLosItemsDelSprint`, `TestSprintCompletion_SprintCerradoSeCalculaIgual`, `TestSprintCompletion_NingunoTerminadoEsCeroDisponible`, `TestCompletion_PorcentajeSinRedondear` |
| CA-34.2 Porcentaje del proyecto | "Porcentaje del proyecto con todo el backlog" | `TestProjectCompletion_UsaTodoElBacklog`, `TestProjectCompletion_NoCuentaOtrosProyectos` |
| CA-34.3 Sin ítems, no disponible | "Un sprint sin ítems no tiene porcentaje disponible", "Un proyecto sin ítems no tiene porcentaje disponible" | `TestSprintCompletion_SprintSinItemsNoDisponible`, `TestProjectCompletion_ProyectoSinItemsNoDisponible` |
| CA-34.4 Cuentan los ítems sin Story Points | "Los ítems sin Story Points cuentan igual" | `TestCompletion_CuentaItemsSinStoryPoints` |
| CA-34.5 Inexistente da error y no modifica nada | "Consultar un sprint inexistente falla" (3), "Consultar un proyecto inexistente falla" (3), "Consultar el porcentaje no modifica los ítems ni los sprints" | `TestSprintCompletion_SprintInexistenteDevuelveErrSprintNotFound`, `TestProjectCompletion_ProyectoInexistenteDevuelveErrProjectNotFound`, `TestCompletion_NoModificaNada`, `TestSprintCompletion_ErrorDelRepositorioDeSprintsSePropaga`, `TestCompletion_ErrorDelRepositorioDeItemsSePropaga` |
| CA-34.6 API e interfaz | — | `TestMetricasSprint_IncluyeElPorcentajeDeHistoriasCompletadas`, `TestMetricasProyecto_PorcentajeRedondeadoAUnDecimal`, `TestMetricas_SinItemsElPorcentajeEsNull`, `TestInterfaz_MetricasMuestraElPorcentajeDeHistoriasCompletadas`; los 400, 404 y 500 de esas rutas los cubren los tests de HU-32 y `TestRutasConID_IDNoNumericoDevuelve400` |

## Fase RED (`2c91ca8`)

El RED agrega los tests, el `.feature` y los stubs de `Completion`, `SprintCompletion` y `ProjectCompletion` en `internal/metrics/completion.go`, que devuelven un `Completion` vacío y `nil`, para que el paquete compile y los tests fallen por las aserciones (Definition of Done del Sprint 3). No cambia ningún test de HU-32.

```
$ go vet ./...
(sin salida)

$ go test ./...
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
--- FAIL: TestMetricasSprint_IncluyeElPorcentajeDeHistoriasCompletadas (0.00s)
    metrics_test.go:203: porcentaje = <nil> con 0 de 0, se esperaba 50 con 2 de 4
--- FAIL: TestMetricasProyecto_PorcentajeRedondeadoAUnDecimal (0.00s)
    metrics_test.go:218: porcentaje = <nil> con 0 de 0, se esperaba 33.3 con 1 de 3
--- FAIL: TestMetricas_SinItemsElPorcentajeEsNull (0.00s)
    metrics_test.go:237: /api/sprints/1/metricas: la respuesta no tiene porcentaje_completadas en null: {"sprint_id":1,"sp_planificados":0,"sp_completados":0,"items_sin_estimar":0}
    metrics_test.go:237: /api/proyectos/1/metricas: la respuesta no tiene porcentaje_completadas en null: {"proyecto_id":1,"sp_planificados":0,"sp_completados":0,"items_sin_estimar":0,"velocidad":null,"sprints_cerrados":0}
--- FAIL: TestInterfaz_MetricasMuestraElPorcentajeDeHistoriasCompletadas (0.00s)
    web_test.go:218: app.js no contiene "porcentaje_completadas"
    web_test.go:218: app.js no contiene "items_terminados"
    web_test.go:218: app.js no contiene "items_totales"
    web_test.go:218: app.js no contiene "Historias completadas"
FAIL
FAIL seguimiento-medicion-gpc/internal/httpapi
14 scenarios (3 passed, 11 failed)
58 steps (47 passed, 11 failed)
--- FAIL: TestFeatures_HU34PorcentajeCompletadas (0.02s)
    completion_features_test.go:32: fallaron escenarios BDD de HU-34
--- FAIL: TestSprintCompletion_EjemploDeTEC03 (0.00s)
    completion_test.go:46: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, se esperaba 2 de 4 ítems terminados, 50 % disponible
--- FAIL: TestSprintCompletion_SoloCuentaLosItemsDelSprint (0.00s)
    completion_test.go:65: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, se esperaba 1 de 2 ítems terminados, 50 % disponible
--- FAIL: TestSprintCompletion_SprintCerradoSeCalculaIgual (0.00s)
    completion_test.go:78: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, se esperaba 2 de 2 ítems terminados, 100 % disponible
--- FAIL: TestSprintCompletion_NingunoTerminadoEsCeroDisponible (0.00s)
    completion_test.go:94: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, se esperaba 0 de 2 ítems terminados, 0 % disponible
--- FAIL: TestSprintCompletion_SprintInexistenteDevuelveErrSprintNotFound (0.00s)
    completion_test.go:120: sprint 99: error = <nil>, se esperaba no existe un sprint con ese id
    completion_test.go:120: sprint 0: error = <nil>, se esperaba no existe un sprint con ese id
    completion_test.go:120: sprint -1: error = <nil>, se esperaba no existe un sprint con ese id
--- FAIL: TestProjectCompletion_UsaTodoElBacklog (0.00s)
    completion_test.go:144: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, se esperaba 2 de 4 ítems terminados, 50 % disponible
--- FAIL: TestProjectCompletion_NoCuentaOtrosProyectos (0.00s)
    completion_test.go:160: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, se esperaba 1 de 1 ítems terminados, 100 % disponible
--- FAIL: TestProjectCompletion_ProyectoInexistenteDevuelveErrProjectNotFound (0.00s)
    completion_test.go:185: proyecto 99: error = <nil>, se esperaba no existe un proyecto con ese id
    completion_test.go:185: proyecto 0: error = <nil>, se esperaba no existe un proyecto con ese id
    completion_test.go:185: proyecto -1: error = <nil>, se esperaba no existe un proyecto con ese id
--- FAIL: TestCompletion_CuentaItemsSinStoryPoints (0.00s)
    completion_test.go:210: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, se esperaba 1 de 2 ítems terminados, 50 % disponible
    completion_test.go:211: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, se esperaba 1 de 2 ítems terminados, 50 % disponible
--- FAIL: TestCompletion_PorcentajeSinRedondear (0.00s)
    completion_test.go:227: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, se esperaba 1 de 3 ítems terminados, 33.333333333333336 % disponible
--- FAIL: TestSprintCompletion_ErrorDelRepositorioDeSprintsSePropaga (0.00s)
    completion_test.go:273: error = <nil>, se esperaba base caída
--- FAIL: TestCompletion_ErrorDelRepositorioDeItemsSePropaga (0.00s)
    completion_test.go:288: sprint: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, error = <nil>, se esperaba vacío y base caída
    completion_test.go:291: proyecto: completion = {Items:0 Finished:0 Percent:0 PercentAvailable:false}, error = <nil>, se esperaba vacío y base caída
FAIL
FAIL seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/sprint
```

Fallan solo tests de HU-34: 13 en `internal/metrics` (12 unitarios y el runner de godog, con 11 de 14 escenarios en falla) y 4 en `internal/httpapi`. Los demás paquetes pasan.

**Lo que ya pasa en el RED, y por qué.** Tres tests y tres escenarios pasan con el stub, porque el `Completion` vacío que devuelve coincide con lo esperado:

- `TestSprintCompletion_SprintSinItemsNoDisponible` y `TestProjectCompletion_ProyectoSinItemsNoDisponible`, y sus dos escenarios: "sin ítems" es exactamente el valor vacío.
- `TestCompletion_NoModificaNada` y su escenario: el stub no lee ni escribe nada.

En `TestInterfaz_MetricasMuestraElPorcentajeDeHistoriasCompletadas`, el texto "sin datos" ya estaba en `app.js` por la velocidad de HU-32, así que esa aserción no falla en el RED; las otras cuatro sí.

## Fase GREEN (`fc2277c`)

- `internal/metrics/completion.go`: `SprintCompletion` (valida el sprint con `sprint.Manager.Get` y filtra con `itemsOfSprint` de HU-32), `ProjectCompletion` (valida el proyecto con `ListItems`) y la función `completion`, que cuenta los ítems y los `TERMINADO` y calcula el porcentaje sin redondear.
- `internal/httpapi/metrics.go`: los dos handlers de HU-32 suman `items_totales`, `items_terminados` y `porcentaje_completadas` (un decimal o `null`) con una estructura embebida, `completionResponse`.
- `internal/web/app.js`: "Ver métricas" agrega "Historias completadas: N % (x de y)." o "Historias completadas: sin datos.".

El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat 2c91ca8 fc2277c -- '*_test.go' '*.feature'
(sin salida)
```

## Refactor

No hubo.

## Desviaciones del ciclo TDD

Ninguna. Una observación de cobertura: los dos handlers tienen una rama defensiva sin test, que devuelve el error si falla la segunda lectura del backlog (la de M8) después de que la primera (Story Points de HU-32) anduvo bien. Por eso `sprintMetrics` queda en 84.6% y `projectMetrics` en 86.7%. No se agregó un test después del GREEN porque pasaría sin un RED propio; la consulta única de HU-35 elimina la doble lectura (ver "Límite conocido" de la SDD).

## Validación final

Sobre `fc2277c`, en Linux con go1.27.1:

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)

$ go test ./...
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
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

$ go test -cover ./...
ok   seguimiento-medicion-gpc/cmd/server             coverage: 31.5% of statements
ok   seguimiento-medicion-gpc/internal/backlog       coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/defect        coverage: 76.6% of statements
ok   seguimiento-medicion-gpc/internal/effort        coverage: 75.2% of statements
ok   seguimiento-medicion-gpc/internal/httpapi       coverage: 97.3% of statements
ok   seguimiento-medicion-gpc/internal/metrics       coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/poker         coverage: 65.2% of statements
ok   seguimiento-medicion-gpc/internal/postgres      coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/postgrestest  coverage: 36.8% of statements
ok   seguimiento-medicion-gpc/internal/project       coverage: 67.4% of statements
ok   seguimiento-medicion-gpc/internal/sprint        coverage: 76.3% of statements
     seguimiento-medicion-gpc/internal/web           coverage: 0.0% of statements
```

godog: `14 scenarios (14 passed)`, `58 steps (58 passed)`.

`go tool cover -func`: `SprintCompletion`, `ProjectCompletion`, `completion` y `toCompletionResponse` en 100.0%. Total del perfil sin base: **80.2%** (`main` en `82e9ed7`, mismo Go: **80.0%**). `internal/httpapi` baja de 98.2% a 97.3% por la rama defensiva descrita arriba.

### Con base (PostgreSQL 16.15) y `-race`

Base `gpc_test` recreada con `db/init/01-schema.sql`:

```
$ TEST_DATABASE_URL=... go test -race -count=1 ./...
ok   (todos los paquetes)
?    seguimiento-medicion-gpc/internal/web  [no test files]
```

Sin SKIP (0 tests omitidos) y sin carreras. Total con `-coverpkg=./...`: **94.5%**. HU-34 no toca la persistencia; la corrida con base confirma que no rompe nada.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | a464ea5 | `docs(HU-34): agregar especificación SDD` |
| RED | 2c91ca8 | `test(HU-34): agregar tests (RED)` |
| GREEN | fc2277c | `feat(HU-34): implementar el porcentaje de historias completadas (GREEN)` |
| Refactor | — | No hubo |

HU-34: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
