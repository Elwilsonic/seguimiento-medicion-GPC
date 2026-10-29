# Evidencia TDD · HU-35 Consultar las métricas del proyecto

> Issue del repositorio: #53

- **Especificación SDD:** [`docs/especificaciones/HU-35-consulta-metricas.md`](../especificaciones/HU-35-consulta-metricas.md), aprobada por Peñalbé Hernán el 28/10/2026 con dos ajustes (comparación del JSON crudo para el orden de los campos y CA-35.4 y CA-35.6 en lenguaje de negocio en el Issue).
- **Escenarios BDD:** [`features/HU-35-consulta-metricas.feature`](../../features/HU-35-consulta-metricas.feature): 19 definiciones (13 escenarios y 6 esquemas, con 3, 3, 4, 2, 2 y 2 ejemplos), que godog ejecuta como 29 escenarios y 231 pasos.
- **Rama:** `feature/HU-35-consulta-metricas`.
- **Story Points:** 3 (Planning del Sprint 4).
- **Dependencias:** HU-32 (#123), HU-34 (#129), HU-33 (#142) y HU-47 (#144), integradas en `main` (`8624c6e`) antes de la SDD. `origin/main` no avanzó durante la historia, así que no hubo merge.

## Historia

Como Agile Enabler, quiero consultar todas las métricas de un proyecto juntas, para evaluar su estado general.

## Resultado

`metrics.Manager` suma `Sprint` y `Project`, que devuelven M1 a M10 en una sola consulta, con una sola lectura de los ítems y una sola del sprint o de la lista de sprints, en el orden de RN-35.4. Los dos handlers de métricas hacen una sola llamada, y una sola función de la API (`optionalOneDecimal`) convierte los cuatro valores que pueden ser `null`. El contrato HTTP no cambia: mismas rutas, campos, orden, tipos y códigos.

## Trazabilidad

**25 tests nuevos (funciones `Test`) cubriendo 29 escenarios BDD ejecutados (19 definiciones):** 16 unitarios en `internal/metrics/summary_test.go`, el runner de godog `TestFeatures_HU35ConsultarMetricas` en `internal/metrics/summary_features_test.go` y 8 de la API en `internal/httpapi/metrics_test.go`. La correspondencia no es 1 a 1: CA-35.6 se cubre solo con tests de la API y varios tests comparten criterio.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-35.1 Todas las métricas en una sola consulta | "Todas las métricas de un sprint con el ejemplo de TEC-03", "Todas las métricas del proyecto con ese sprint cerrado" | `TestSprint_EjemploDeTEC03`, `TestProject_EjemploDeTEC03ConElSprintCerrado`, `TestConsultaMetricas_ContratoDelSprint`, `TestConsultaMetricas_ContratoDelProyecto` |
| CA-35.2 No disponible como `null`, sin error | "Un proyecto recién creado responde con las cuatro métricas no disponibles", "Un sprint sin ítems, esfuerzos ni defectos", "Un proyecto con ítems y sin sprints cerrados solo no tiene velocidad", "Ítems estimados sin esfuerzos solo no tienen desviación", "Esfuerzos sin horas estimadas solo no tienen desviación" | `TestProject_RecienCreadoLasCuatroNoDisponibles`, `TestSprint_SinItemsPorcentajeYDesviacionNoDisponibles`, `TestProject_SinSprintsCerradosSoloLaVelocidadNoDisponible`, `TestSprint_SinEsfuerzosSoloLaDesviacionNoDisponible`, `TestSprint_SinHorasEstimadasSoloLaDesviacionNoDisponible`, `TestConsultaMetricas_ContratoDelProyectoVacio` |
| CA-35.3 Inexistente, ID no numérico y falla de repositorio | "Consultar un sprint inexistente falla" (3), "Consultar un proyecto inexistente falla" (3), "Si falla la lectura del sprint o de los ítems no se leen esfuerzos ni defectos" (4), "Si falla la lectura de los esfuerzos no se leen los defectos" (2), "Si falla la lectura de los defectos la consulta falla" (2) | `TestSprint_InexistenteDevuelveErrSprintNotFound`, `TestProject_InexistenteDevuelveErrProjectNotFound`, `TestSprint_ErroresEnElOrdenDeLectura`, `TestProject_ErroresEnElOrdenDeLectura`, `TestConsultaMetricas_IDNoNumericoDevuelve400`, `TestConsultaMetricas_SprintInexistenteDevuelve404`, `TestConsultaMetricas_ProyectoInexistenteDevuelve404`, `TestConsultaMetricas_FallaDeEsfuerzosODefectosDevuelve500` |
| CA-35.4 Los ítems del backlog se leen una sola vez | "Un proyecto con tres sprints suma sus esfuerzos y lee los ítems una sola vez", "Los ítems del backlog se leen una sola vez por consulta" (2) | `TestSprint_LeeLosItemsUnaSolaVez`, `TestProject_LeeLosItemsUnaSolaVez`, `TestConsultaMetricas_LeeLosItemsUnaSolaVez` |
| CA-35.5 Mismos valores que HU-32, HU-33, HU-34 y HU-47 | "Un proyecto con tres sprints suma sus esfuerzos…", "Los esfuerzos y los defectos de otro proyecto no se incluyen", "Un defecto detectado en un sprint y resuelto en otro", "Horas con decimales dan una desviación de cero sin redondear en el dominio", "La consulta única da lo mismo que las operaciones separadas" | `TestSprint_EsIgualALasOperacionesSeparadas`, `TestProject_EsIgualALasOperacionesSeparadas` |
| CA-35.6 Contrato sin cambios y documentado | — | `TestConsultaMetricas_ContratoDelSprint`, `TestConsultaMetricas_ContratoDelProyecto`, `TestConsultaMetricas_ContratoDelProyectoVacio` (cuerpo crudo comparado con el JSON esperado, ambos con `json.Compact`), `TestConsultaMetricas_IDNoNumericoDevuelve400`, `TestConsultaMetricas_SprintInexistenteDevuelve404`, `TestConsultaMetricas_ProyectoInexistenteDevuelve404`, `TestConsultaMetricas_FallaDeEsfuerzosODefectosDevuelve500`; los tests existentes de los dos endpoints (HU-32, HU-33, HU-34 y HU-47) pasan sin cambios |
| CA-35.7 Copia por valor; consultar no modifica nada | "Consultar no modifica nada y devuelve una copia" | `TestSummary_NoModificaNadaYDevuelveCopias` |

## Fase RED (`f2daaf3`)

El RED agrega el `.feature`, los tests y los stubs en `internal/metrics/summary.go`: los tipos `SprintMetrics` y `ProjectMetrics` con los campos de la SDD, y `Sprint` y `Project`, que devuelven el valor cero y `nil`. Los tests unitarios y los pasos de godog usan repositorios espía (sobre los repositorios en memoria) que cuentan las lecturas y que fallan a pedido después de cargar los datos.

Reproducido sobre `f2daaf3` en una copia temporal con `git worktree`, en Windows con go1.27.1, sin base:

```
$ go vet ./...
(sin salida)

$ go test -count=1 ./...
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
--- FAIL: TestConsultaMetricas_LeeLosItemsUnaSolaVez (0.00s)
    metrics_test.go:784: /api/sprints/2/metricas: lecturas de ítems = 4, se esperaba 2 (1 de metrics y 1 interna de defect)
    metrics_test.go:784: /api/proyectos/1/metricas: lecturas de ítems = 4, se esperaba 2 (1 de metrics y 1 interna de defect)
FAIL
FAIL seguimiento-medicion-gpc/internal/httpapi
29 scenarios (29 failed)
231 steps (143 passed, 29 failed, 59 skipped)
--- FAIL: TestFeatures_HU35ConsultarMetricas
--- FAIL: TestSprint_EjemploDeTEC03 (0.00s)
    summary_test.go:264: Sprint = {SprintID:0 Points:{SprintID:0 Planned:0 ...}, se esperaba {SprintID:1 ...}
--- FAIL: TestProject_EjemploDeTEC03ConElSprintCerrado (0.00s)
    summary_test.go:285: Project = {ProjectID:0 ...}, se esperaba {ProjectID:1 ...}
--- FAIL: TestSprint_EsIgualALasOperacionesSeparadas (0.00s)
    summary_test.go:304: Sprint = {SprintID:0 ...}, se esperaba {SprintID:1 ...}
--- FAIL: TestProject_EsIgualALasOperacionesSeparadas (0.00s)
    summary_test.go:316: Project = {ProjectID:0 ...}, se esperaba {ProjectID:1 ...}
--- FAIL: TestProject_RecienCreadoLasCuatroNoDisponibles (0.00s)
    summary_test.go:327: Project = {ProjectID:0 ...}, se esperaba {ProjectID:1 ...}
--- FAIL: TestSprint_SinItemsPorcentajeYDesviacionNoDisponibles (0.00s)
    summary_test.go:342: Sprint = {SprintID:0 ...}, se esperaba {SprintID:1 ...}
--- FAIL: TestProject_SinSprintsCerradosSoloLaVelocidadNoDisponible (0.00s)
    summary_test.go:359: Project = {ProjectID:0 ...}, se esperaba {ProjectID:1 ...}
--- FAIL: TestSprint_SinEsfuerzosSoloLaDesviacionNoDisponible (0.00s)
    summary_test.go:377: Sprint = {SprintID:0 ...}, se esperaba {SprintID:1 ...}
--- FAIL: TestSprint_SinHorasEstimadasSoloLaDesviacionNoDisponible (0.00s)
    summary_test.go:395: Sprint = {SprintID:0 ...}, se esperaba {SprintID:1 ...}
--- FAIL: TestSprint_InexistenteDevuelveErrSprintNotFound (0.00s)
    summary_test.go:410: Sprint(99): err = <nil>, se esperaba no existe un sprint con ese id
--- FAIL: TestProject_InexistenteDevuelveErrProjectNotFound (0.00s)
    summary_test.go:426: Project(99): err = <nil>, se esperaba no existe un proyecto con ese id
--- FAIL: TestSprint_ErroresEnElOrdenDeLectura (0.00s)
    summary_test.go:483: sprint e ítems: err = <nil>, se esperaba falla al leer el sprint
--- FAIL: TestProject_ErroresEnElOrdenDeLectura (0.00s)
    summary_test.go:511: ítems y sprints: err = <nil>, se esperaba falla al leer los ítems
--- FAIL: TestSprint_LeeLosItemsUnaSolaVez (0.00s)
    summary_test.go:533: lecturas de ítems = 0, se esperaba 2 (1 de metrics y 1 interna de defect)
--- FAIL: TestProject_LeeLosItemsUnaSolaVez (0.00s)
    summary_test.go:549: lecturas de ítems = 0, se esperaba 2 (1 de metrics y 1 interna de defect)
--- FAIL: TestSummary_NoModificaNadaYDevuelveCopias (0.00s)
    summary_test.go:576: IDs = 0 y 0, se esperaba 1 y 1
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

Las estructuras se acortaron con `...`; se muestra el primer mensaje de cada test. Los 29 escenarios fallan con "la consulta del sprint/proyecto devolvió el ID X, se esperaba Y" (15) o con `err = <nil>, se esperaba …` (14). Ningún escenario falla por pasos sin definir ni por panic.

Fallan 18 tests de nivel superior, todos de HU-35: los 16 unitarios y el runner de godog en `internal/metrics`, y 1 en `internal/httpapi`. Los escenarios de HU-32 (16), HU-33 (14), HU-34 (14) y HU-47 (15) siguen en verde, y los demás paquetes pasan.

## Fase GREEN (`79af174`)

- `internal/metrics/summary.go`: `Sprint` lee el sprint (`Get`), los ítems de su proyecto (`ListItems`), los esfuerzos del sprint (`ListBySprint`) y los defectos del proyecto (`defect.Manager.ListByProject`), en ese orden, y corta en el primer error. `Project` lee los ítems (que valida el proyecto), la lista de sprints, los esfuerzos de cada sprint en orden de ID (M5 se suma sprint por sprint, como HU-33) y los defectos. Reutilizan `storyPoints`, `completion`, `hoursOf` e `itemsOfSprint`.
- `internal/metrics/defects.go`: el conteo de M9 y M10 pasa a `sprintDefects` y `projectDefects`, que usan `SprintDefects`, `ProjectDefects`, `Sprint` y `Project`.
- `internal/metrics/story_points.go`: la parte de cálculo de `ProjectStoryPoints` pasa a `projectPoints`, una función pura que usan `ProjectStoryPoints` y `Project` (ver "Desvíos y límites").
- `internal/httpapi/metrics.go`: cada handler hace una sola llamada (`Sprint` o `Project`). `optionalOneDecimal` convierte velocidad, porcentaje de completadas y las dos desviaciones: `null` si no están disponibles y, si lo están, un decimal sin cero negativo. Se elimina `toProjectMetricsResponse`, que quedó sin uso.

El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat f2daaf3 79af174 -- '*_test.go' '*.feature'
(sin salida)

$ git diff --stat f2daaf3 79af174
 internal/httpapi/metrics.go      | 119 ++++++++++++++-------------------------
 internal/metrics/defects.go      |  17 +++++-
 internal/metrics/story_points.go |   8 ++-
 internal/metrics/summary.go      |  69 +++++++++++++++++++++--
 4 files changed, 130 insertions(+), 83 deletions(-)
```

## Refactor

No hubo commit de refactor aparte. La extracción de `projectPoints`, `sprintDefects` y `projectDefects` entró en el GREEN, porque es la forma de reutilizar los cálculos existentes sin duplicarlos (RN-35.2).

## Desvíos y límites

1. **API en el RED.** De los 8 tests nuevos de la API, 7 pasaron en el RED: los tres contratos JSON, el 400, los dos 404 y el 500. Verifican un contrato que no debe cambiar (CA-35.6) y los handlers de HU-32 a HU-47 ya lo cumplían. Solo falló `TestConsultaMetricas_LeeLosItemsUnaSolaVez` (4 lecturas de ítems en lugar de 2), que es el que cambia con el GREEN.
2. **Sin escenario godog para CA-35.6 ni para el ID no numérico.** Son solo de HTTP y los cubren los tests de la API, como CA-32.9 en HU-32. Lo indica el encabezado del `.feature`.
3. **Copia por valor.** Los tests de copia (`TestSummary_NoModificaNadaYDevuelveCopias` y el paso "modificar el resultado no cambia la próxima consulta") son débiles: `SprintMetrics` y `ProjectMetrics` no tienen punteros, listas ni mapas, así que modificar una copia no puede alterar nada. En el RED fallaron solo por la verificación del ID.
4. **Archivo fuera de la lista de la SDD.** El GREEN modificó `internal/metrics/story_points.go`, que la SDD no lista en "Archivos que cambian", para extraer `projectPoints` como función pura y no duplicar la lógica de la velocidad. `ProjectStoryPoints` hace las mismas lecturas, en el mismo orden, con el mismo resultado; sus tests y escenarios de HU-32 pasan sin cambios.
5. **`-race`.** `go test -race` no se corrió en la máquina local: requiere cgo y no hay `gcc` (`go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`). Lo corre el CI (`go test -race -coverprofile=coverage.out ./...`).
6. **PostgreSQL.** Localmente se saltaron 30 tests por falta de `TEST_DATABASE_URL` (backlog 3, defect 10, effort 7, poker 2, project 4 y sprint 4). En el CI también se saltan: el workflow no levanta una base. HU-35 no toca la persistencia.
7. **Límites conocidos de la SDD.** Siguen repitiéndose las lecturas internas de `internal/effort` e `internal/defect`, que la historia no toca: en el proyecto, una lectura de esfuerzos por sprint, cada una con su lectura del sprint para validarlo; en el sprint, `ListBySprint` vuelve a leer el sprint; y `defect.Manager.ListByProject` lee el backlog por su cuenta, por eso el repositorio de ítems recibe 2 lecturas por consulta. Las lecturas de fuentes distintas no son atómicas entre sí.
8. **Código de salida en Windows.** La primera corrida de `go test -cover ./...` sobre el GREEN terminó con código 1 aunque todos los paquetes dieron `ok`, por un error de Windows al borrar el ejecutable temporal (`go: unlinkat …\httpapi.test.exe: The process cannot access the file because it is being used by another process`), no por un test fallido. La repetición sin caché (`-count=1`) dio código 0.

## Validación final

Sobre `79af174`, en Windows con go1.27.1, sin base:

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)

$ go test -count=1 -cover ./...
ok   seguimiento-medicion-gpc/cmd/server             coverage: 31.5% of statements
ok   seguimiento-medicion-gpc/internal/backlog       coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/dashboard     coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/defect        coverage: 77.7% of statements
ok   seguimiento-medicion-gpc/internal/effort        coverage: 75.2% of statements
ok   seguimiento-medicion-gpc/internal/httpapi       coverage: 98.6% of statements
ok   seguimiento-medicion-gpc/internal/metrics       coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/poker         coverage: 69.5% of statements
ok   seguimiento-medicion-gpc/internal/postgres      coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/postgrestest  coverage: 36.8% of statements
ok   seguimiento-medicion-gpc/internal/project       coverage: 67.5% of statements
ok   seguimiento-medicion-gpc/internal/sprint        coverage: 76.3% of statements
     seguimiento-medicion-gpc/internal/web           coverage: 0.0% of statements
```

godog de HU-35: `29 scenarios (29 passed)`, `231 steps (231 passed)`. Los 25 tests nuevos pasan.

`go tool cover -func` (perfil de `go test -count=1 -coverprofile=cover.out ./...`): `Sprint`, `Project`, `projectPoints`, `sprintDefects`, `projectDefects`, `SprintDefects`, `ProjectDefects`, `ProjectStoryPoints`, `sprintMetrics`, `projectMetrics`, `optionalOneDecimal`, `toHoursResponse`, `toCompletionResponse` y `roundOneDecimal` en 100.0%. `internal/httpapi` sube de 98.0% a 98.6% porque desaparecen las ramas defensivas de HU-34 (segunda lectura del backlog) que `sprintMetrics` y `projectMetrics` tenían sin cubrir. Total del perfil sin base: **82.7%** (`main` en `8624c6e`, mismo Go y comando con `-p 1`, en una copia temporal con `git worktree`: **82.5%**, con `internal/metrics` 100.0% e `internal/httpapi` 98.0%). No se midió con PostgreSQL.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | 71ef0b4 | `docs(HU-35): agregar la especificación (SDD)` |
| RED | f2daaf3 | `test(HU-35): agregar tests (RED)` |
| GREEN | 79af174 | `feat(HU-35): consultar las métricas en una sola consulta (GREEN)` |
| Refactor | — | No hubo |

HU-35: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
