# Evidencia TDD · HU-38 Generar el reporte del proyecto

> Issue del repositorio: #56

- **Especificación SDD:** [`docs/especificaciones/HU-38-reporte-proyecto.md`](../especificaciones/HU-38-reporte-proyecto.md), aprobada por Peñalbé Hernán el 29/10/2026 con dos ajustes (conversión duplicada de las métricas registrada como límite conocido y reutilización de `toEffortListResponse` precisada).
- **Escenarios BDD:** [`features/HU-38-reporte-proyecto.feature`](../../features/HU-38-reporte-proyecto.feature): 15 definiciones (13 escenarios y 2 esquemas, con 3 y 5 ejemplos), que godog ejecuta como 21 escenarios y 184 pasos.
- **Rama:** `feature/HU-38-reporte-proyecto`.
- **Story Points:** 5 (Planning del Sprint 4).
- **Dependencias:** HU-27 (#126), HU-31 (#141) y HU-35 (#145), integradas en `main` (`33b1234`) antes de la SDD. Durante la historia, `origin/main` avanzó con TEC-06 (#146, solo `docs/documentacion-tecnica.md`), que se trajo con el merge `f58b5e7`, sin conflictos, y con HU-37 (#147), que se trajo con el merge `0f98b48`, con conflictos resueltos (ver "Integración con main (HU-37)").

## Historia

Como Agile Enabler, quiero generar un reporte del proyecto, para presentar su estado a los profesores.

## Resultado

El paquete nuevo `internal/report` arma en una sola consulta el reporte de un proyecto: sus datos, M1 a M10 (HU-35), los esfuerzos de todos sus sprints con el total exacto de horas (HU-27), sus defectos (HU-31) y los integrantes que registraron esfuerzo, sin repetir. Lee en el orden de RN-38.3 (proyecto → métricas → sprints → esfuerzos de cada sprint → defectos) y corta en el primer error. Se expone en `GET /api/proyectos/{id}/reporte`, donde cada parte tiene el formato del endpoint que ya existe, y en la sección nueva "Reportes" de la interfaz, con el botón "Generar reporte". No hay cálculos nuevos ni cambios en los paquetes de dominio existentes.

## Trazabilidad

**24 tests nuevos (funciones `Test`) y una fila nueva en un test existente, cubriendo 21 escenarios BDD ejecutados (15 definiciones):** 13 unitarios en `internal/report/project_test.go`, el runner de godog `TestFeatures_HU38ReporteProyecto` en `internal/report/project_features_test.go`, 7 de la API en `internal/httpapi/reports_test.go`, 3 de la interfaz en `internal/httpapi/web_test.go` y la fila "reporte del proyecto" en `TestRutasConID_IDNoNumericoDevuelve400` (`internal/httpapi/server_test.go`). Los tests nuevos se contaron con `git diff --unified=0 f19bf62 a396ae2 -- <archivo> | grep -c '^+func Test'` (13, 1, 7, 3 y 0 en `server_test.go`). La correspondencia no es 1 a 1: CA-38.7 se cubre solo con tests de la API y de la interfaz, y varios tests comparten criterio.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-38.1 Datos, métricas, esfuerzos, defectos e integrantes | "El reporte del ejemplo de TEC-03 tiene todas sus partes", "Un proyecto sin fecha de inicio muestra sus datos sin fecha", "Los defectos aparecen en los tres estados, ordenados por ID" | `TestProject_EjemploDeTEC03`, `TestProject_DefectosEnLosTresEstadosOrdenadosPorID`, `TestReporteProyecto_ContratoDelEjemploDeTEC03` |
| CA-38.2 Proyecto sin datos con las secciones vacías | "Un proyecto recién creado genera el reporte con las secciones vacías", "Un proyecto con ítems y sin sprints tiene las métricas y las listas vacías", "Sprints sin esfuerzos dejan vacíos los esfuerzos y los integrantes" | `TestProject_ProyectoSinDatos`, `TestProject_ProyectoConItemsYSinSprints`, `TestProject_SprintsSinEsfuerzos`, `TestReporteProyecto_ContratoDelProyectoVacio` |
| CA-38.3 Proyecto inexistente y fallas de lectura | "Generar el reporte de un proyecto inexistente falla" (3), "Si falla una lectura no se hacen las siguientes" (5) | `TestProject_ProyectoInexistenteNoLeeNadaMas`, `TestProject_ErroresEnElOrdenDeLectura`, `TestReporteProyecto_InexistenteDevuelve404`, `TestReporteProyecto_ErrorDelRepositorioDevuelve500SinDetalle` |
| CA-38.4 Coincide con las consultas existentes; total = horas reales | "Tres sprints, abiertos y cerrados, con esfuerzos", "Horas con decimales suman el total exacto, igual a las horas reales" | `TestProject_CadaParteEsIgualASuOperacionDeOrigen`, `TestProject_TotalDeHorasExactoEIgualAHorasReales`, `TestProject_EsfuerzosOrdenadosPorSprintYPorID`, `TestReporteProyecto_CadaParteEsIgualASuEndpoint`, `TestReporteProyecto_RedondeaElTotalDeHorasAUnDecimal` |
| CA-38.5 Integrantes sin repetir; sin datos de otros proyectos | "Un integrante con esfuerzos en varios sprints aparece una sola vez", "\"Ana\" y \"ana\" se listan como dos integrantes", "Los esfuerzos, los defectos y los integrantes de otro proyecto no se incluyen" | `TestProject_IntegrantesSinRepetirEnOrdenDePrimeraAparicion`, `TestProject_NoIncluyeDatosDeOtroProyecto` |
| CA-38.6 Generar el reporte no modifica nada | "Generar el reporte dos veces sin cambios da el mismo resultado", "Modificar el reporte devuelto no cambia lo guardado" | `TestProject_NoModificaNadaYDevuelveCopias` |
| CA-38.7 Desde la API y desde la sección "Reportes" | — | `TestReporteProyecto_ContratoDelEjemploDeTEC03`, `TestReporteProyecto_ContratoDelProyectoVacio` (cuerpo crudo comparado con el JSON esperado, ambos con `json.Compact`), `TestReporteProyecto_IDNoNumericoDevuelve400`, `TestReporteProyecto_InexistenteDevuelve404`, `TestReporteProyecto_ErrorDelRepositorioDevuelve500SinDetalle`, la fila "reporte del proyecto" de `TestRutasConID_IDNoNumericoDevuelve400`, `TestInterfaz_SeccionReportesConElBotonGenerarReporte`, `TestInterfaz_GenerarReporteLlamaALaAPIConElProyectoActual`, `TestInterfaz_GenerarReporteMuestraCadaParte` |

## Fase RED (`a396ae2`)

El RED agrega el `.feature`, los tests y los stubs de `internal/report`: `doc.go`, `manager.go` (`Manager` y `NewManager` con la firma de la SDD) y `project.go` (los tipos `ProjectReport`, `Efforts`, `Defects` y `Members`, y `Project`, que devuelve el valor cero y `nil`). Los tests unitarios y los pasos de godog usan repositorios espía (sobre los repositorios en memoria) que cuentan las lecturas y que fallan a pedido desde la lectura número `failAt`, para hacer fallar la lectura del reporte después de la que ya hizo `metrics.Manager.Project`.

Reproducido sobre `a396ae2` en una copia temporal con `git worktree`, en Windows con go1.27.1, sin base:

```
$ go vet ./...
(sin salida)

$ go test -count=1 ./...
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
--- FAIL: TestReporteProyecto_ContratoDelEjemploDeTEC03
    reports_test.go:87: GET /api/proyectos/1/reporte: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
--- FAIL: TestReporteProyecto_ContratoDelProyectoVacio
    reports_test.go:176: GET /api/proyectos/2/reporte: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
--- FAIL: TestReporteProyecto_CadaParteEsIgualASuEndpoint
    reports_test.go:238: GET /api/proyectos/1/reporte: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
--- FAIL: TestReporteProyecto_RedondeaElTotalDeHorasAUnDecimal
    reports_test.go:275: GET /api/proyectos/1/reporte: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
--- FAIL: TestReporteProyecto_InexistenteDevuelve404
    reports_test.go:287: GET /api/proyectos/99/reporte: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
--- FAIL: TestReporteProyecto_IDNoNumericoDevuelve400
    reports_test.go:295: GET /api/proyectos/abc/reporte: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
--- FAIL: TestReporteProyecto_ErrorDelRepositorioDevuelve500SinDetalle
    reports_test.go:311: GET /api/proyectos/1/reporte: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400
    --- FAIL: TestRutasConID_IDNoNumericoDevuelve400/reporte_del_proyecto
        server_test.go:236: GET /api/proyectos/abc/reporte: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
        server_test.go:238: código = 404, se esperaba 400; cuerpo: 404 page not found
--- FAIL: TestInterfaz_SeccionReportesConElBotonGenerarReporte
    web_test.go:773: la página no contiene la sección de reportes
--- FAIL: TestInterfaz_GenerarReporteLlamaALaAPIConElProyectoActual
    web_test.go:803: app.js no contiene "$(\"generar-reporte\")"
--- FAIL: TestInterfaz_GenerarReporteMuestraCadaParte
    web_test.go:823: app.js no contiene "$(\"reporte-proyecto\")"
FAIL
FAIL seguimiento-medicion-gpc/internal/httpapi
ok   seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
21 scenarios (21 failed)
184 steps (128 passed, 21 failed, 35 skipped)
--- FAIL: TestFeatures_HU38ReporteProyecto
--- FAIL: TestProject_EjemploDeTEC03
    project_test.go:364: Project = {ID:0 Name: Description: Status: ...}, se esperaba {ID:1 Name:Software Metrics ...}
--- FAIL: TestProject_CadaParteEsIgualASuOperacionDeOrigen
    project_test.go:406: Project = {ID:0 ...}, se esperaba {ID:1 Name:Sistema de ventas ...}
--- FAIL: TestProject_TotalDeHorasExactoEIgualAHorasReales
    project_test.go:443: TotalHours = 0, se esperaba la suma exacta 0.6000000000000001
--- FAIL: TestProject_EsfuerzosOrdenadosPorSprintYPorID
    project_test.go:468: Efforts = [], se esperaba [{ID:1 SprintID:1 ItemID:1 Member:Ana Hours:5} ...]
--- FAIL: TestProject_IntegrantesSinRepetirEnOrdenDePrimeraAparicion
    project_test.go:481: Members = [], se esperaba ["Ana" "Luis" "ana"]
--- FAIL: TestProject_DefectosEnLosTresEstadosOrdenadosPorID
    project_test.go:499: Defects = [] con estados [], se esperaba [1 2 3] con [ABIERTO RESUELTO EN_CORRECCION]
--- FAIL: TestProject_NoIncluyeDatosDeOtroProyecto
    project_test.go:524: Members del otro proyecto = [], se esperaba ["Pedro"]
--- FAIL: TestProject_ProyectoSinDatos
    project_test.go:557: Project = {ID:0 ...}, se esperaba {ID:1 Name:Proyecto sin datos ...}
--- FAIL: TestProject_ProyectoConItemsYSinSprints
    project_test.go:576: Metrics = {ProjectID:0 ...}, se esperaban 2 ítems, 5 SP y 4 h estimadas
--- FAIL: TestProject_SprintsSinEsfuerzos
    project_test.go:594: Metrics = {ProjectID:0 ...}, se esperaba 1 sprint cerrado con 3 SP completados
--- FAIL: TestProject_ProyectoInexistenteNoLeeNadaMas
    project_test.go:610: id 99: err = <nil>, se esperaba no existe un proyecto con ese id
--- FAIL: TestProject_ErroresEnElOrdenDeLectura
    project_test.go:661: proyecto: err = <nil>, se esperaba falla la lectura del proyecto
--- FAIL: TestProject_NoModificaNadaYDevuelveCopias
    project_test.go:690: el reporte = {Project:{ID:0 ...}}, se esperaban esfuerzos, defectos e integrantes
FAIL
FAIL seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
?    seguimiento-medicion-gpc/internal/web [no test files]
FAIL
```

Las estructuras se acortaron con `...`; se muestra el primer mensaje de cada test. Los 21 escenarios fallan por aserción en un paso "Entonces" (un paso fallido por escenario); ninguno falla por pasos sin definir ni por panic (la salida no contiene "panic").

Fallan 25 tests de nivel superior: los 24 nuevos de HU-38 (13 unitarios y el runner de godog en `internal/report`, y 10 en `internal/httpapi`) y `TestRutasConID_IDNoNumericoDevuelve400`, del que falla solo la fila nueva de HU-38 (ver "Desvíos y límites"). Los demás paquetes y los demás tests de `internal/httpapi` pasan.

## Fase GREEN (`7d7bfb2`)

- `internal/report/project.go`: `Project` lee el proyecto (`project.Manager.Get`, que valida que exista), las métricas (`metrics.Manager.Project`), los sprints del proyecto (`sprint.Manager.ListByProject`), los esfuerzos de cada sprint en orden de ID (`effort.Manager.ListBySprint`) y los defectos (`defect.Manager.ListByProject`), en ese orden, y corta en el primer error. Une los esfuerzos por sprint y suma `TotalHours` sprint por sprint, como M5 en HU-35. `membersOf` arma los integrantes sin repetir, en orden de primera aparición. Las listas son nuevas y nunca `nil`.
- `internal/httpapi/reports.go` (nuevo): `projectReportResponse`, `memberListResponse`, el handler `projectReport` y `toReportMetricsResponse`, que convierte `metrics.ProjectMetrics` con `toCompletionResponse`, `toHoursResponse`, `toDefectsResponse` y `optionalOneDecimal`. "esfuerzos" se arma con `effort.Summary{Efforts: ..., TotalHours: ...}` y `toEffortListResponse`; "defectos", con `toDefectListResponse`; "proyecto", con `toProjectResponse`.
- `internal/httpapi/server.go`: el campo `reports`, la creación de `report.Manager` en `NewHandler` (sin cambiar su firma, con el mismo `metrics.Manager` del servidor) y la ruta `GET /api/proyectos/{id}/reporte`.
- `internal/web/index.html` y `internal/web/app.js`: la sección "Reportes", con el botón "Generar reporte" y sus bloques, y el código que la llena con `textContent`, reutilizando `filaEsfuerzo`, `filaDefecto`, `historiasCompletadas`, `horasYDesviacion` y `defectos` sin modificarlas. El texto de story points y velocidad tiene su propia función, `storyPointsReporte`, con el mismo formato que "Ver métricas". `styles.css` no cambió.

No cambian `internal/httpapi/metrics.go`, `internal/metrics` ni los demás paquetes de dominio. El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat a396ae2 7d7bfb2 -- '*_test.go' '*.feature'
(sin salida)

$ git diff --stat a396ae2 7d7bfb2
 internal/httpapi/reports.go | 75 +++++++++++++++++++++++++++++++++++++++++++++
 internal/httpapi/server.go  |  6 ++++
 internal/report/project.go  | 56 +++++++++++++++++++++++++++++++--
 internal/web/app.js         | 44 ++++++++++++++++++++++++++
 internal/web/index.html     | 33 ++++++++++++++++++++
 5 files changed, 212 insertions(+), 2 deletions(-)
```

## Refactor

En el GREEN no hubo commit de refactor. Después del cierre del Sprint 4 se hizo el refactor que dejaba pendiente el desvío 5:

- **Refactor (`14cffba`), acción 9 de la Retrospectiva del Sprint 4.** Con HU-37 ya integrada, la conversión de `metrics.ProjectMetrics` a `projectMetricsResponse` se unifica en `toProjectMetricsResponse` (`internal/httpapi/metrics.go`), que usan el handler `projectMetrics` (HU-35) y `toProjectReportResponse` (HU-38). Se elimina `toReportMetricsResponse` de `reports.go`. No cambia ningún test ni el comportamiento: `TestReporteProyecto_CadaParteEsIgualASuEndpoint` sigue verificando que "metricas" es igual a `GET /api/proyectos/{id}/metricas`.
- **Validación** sobre `main` (`6f93166`) con el refactor, en Linux con go1.27.1: `gofmt -l` y `go vet ./...` sin salida; `go test -race -count=1 ./...` da `ok` en los 14 paquetes; `projectMetrics`, `toProjectMetricsResponse` y `toProjectReportResponse` en 100.0%, `internal/httpapi` sigue en 98.7% y el total del perfil en 84.5% (sin base).

## Prueba de punta a punta

Con el servidor en memoria (`PORT=8099 go run ./cmd/server`, sin `DATABASE_URL`), se cargó por la API el ejemplo 1 de la SDD (20 respuestas 200 y 7 respuestas 201): el proyecto con su fecha de inicio y en `EN_CURSO`, un sprint, los cuatro ítems de TEC-03, los tres esfuerzos, un defecto resuelto en el sprint y el cierre del sprint. La descripción con tilde se envió en UTF-8 desde un archivo. Respuesta real de `curl -i http://localhost:8099/api/proyectos/1/reporte`:

```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Thu, 29 Oct 2026 02:51:13 GMT
Content-Length: 1145

{"proyecto":{"id":1,"nombre":"Software Metrics","descripcion":"Estimación y seguimiento de proyectos","estado":"EN_CURSO","fecha_inicio":"2026-10-05"},"metricas":{"proyecto_id":1,"sp_planificados":10,"sp_completados":8,"items_sin_estimar":1,"velocidad":8,"sprints_cerrados":1,"items_totales":4,"items_terminados":2,"porcentaje_completadas":50,"horas_estimadas":20,"horas_reales":20,"desviacion_horas":0,"desviacion_porcentual":0,"items_sin_horas_estimadas":1,"defectos_detectados":1,"defectos_resueltos":1},"esfuerzos":{"esfuerzos":[{"id":1,"sprint_id":1,"item_id":1,"integrante":"Ana","fecha":"2026-10-06","actividad":"Desarrollo","horas":12},{"id":2,"sprint_id":1,"item_id":2,"integrante":"Luis","fecha":"2026-10-07","actividad":"Desarrollo","horas":5},{"id":3,"sprint_id":1,"item_id":3,"integrante":"Ana","fecha":"2026-10-08","actividad":"Pruebas","horas":3}],"total_horas":20},"defectos":{"defectos":[{"id":1,"descripcion":"El total de horas no suma los decimales","severidad":"ALTA","estado":"RESUELTO","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":1}],"cantidad":1},"integrantes":{"integrantes":["Ana","Luis"],"cantidad":2}}
```

Es el JSON del ejemplo 1 de la SDD. En la interfaz, "Generar reporte" con el proyecto 1 mostró los datos del proyecto con "Inicio: 2026-10-05", las métricas con un decimal, la tabla de 3 esfuerzos con "Total: 20.0 horas", la tabla con el defecto y "Cantidad: 1 defecto(s)" e "Integrantes que registraron esfuerzo: Ana, Luis", sin mensaje de error (texto leído del DOM).

## Desvíos y límites

1. **Fila fallando en un test existente en el RED.** `TestRutasConID_IDNoNumericoDevuelve400` quedó en el RED con la fila de HU-38 fallando (404 en lugar de 400, porque la ruta todavía no existía). La SDD pide agregar la fila, y en el GREEN no se tocan tests, así que se agregó en el RED. Las demás filas del test pasaron, y en el GREEN pasa completo.
2. **Dos tests corregidos antes del commit RED.** El esperado de `TestProject_TotalDeHorasExactoEIgualAHorasReales` se calculaba como `(0.1 + 0.2) + 0.3`: Go evalúa esa suma de constantes de forma exacta (0.6) y no coincide con la suma en `float64` que hace el código (0.6000000000000001); ahora suma variables `float64` en el orden de RN-38.4. El paso "generar el reporte de nuevo da el mismo resultado" pasaba solo en el RED porque comparaba dos reportes vacíos; ahora exige los datos del proyecto y sus esfuerzos.
3. **Columnas de la tabla de esfuerzos.** La tabla de la sección "Reportes" reutiliza `filaEsfuerzo`, así que sigue su orden de columnas (ID, Fecha, Integrante, Actividad, Ítem, Sprint, Horas), no el que lista la SDD.
4. **Ids extra en la interfaz.** Además de los de la SDD, la sección tiene `reporte-total-horas` y `reporte-cantidad-defectos`, para el total y la cantidad debajo de cada tabla.
5. **Conversión de las métricas duplicada (resuelto en el refactor posterior, ver "Refactor").** `toReportMetricsResponse` (`internal/httpapi/reports.go`) duplica la conversión del handler `projectMetrics` de HU-35, para no modificar `metrics.go` mientras HU-37 lo cambia. `TestReporteProyecto_CadaParteEsIgualASuEndpoint` verifica que "metricas" sea igual a la respuesta de `GET /api/proyectos/{id}/metricas`. Se unifica cuando HU-37 y HU-38 estén en `main` (límite conocido de la SDD).
6. **Límites conocidos de la SDD.** El reporte llama una sola vez a cada operación, pero reutiliza `metrics.Manager.Project` sin cambios. Para un proyecto con N sprints, a nivel de repositorio, se leen 4 veces el proyecto, 3 veces los ítems, 2 veces la lista de sprints, 2N veces un sprint por ID, 2N veces los esfuerzos de un sprint y 2 veces los defectos (tabla de "Límites conocidos" de la SDD; `TestProject_ErroresEnElOrdenDeLectura` verifica las lecturas de sprints, esfuerzos y defectos). Las lecturas de fuentes distintas no son atómicas. El reporte no lista los ítems del backlog: las historias planificadas y completadas y sus estimaciones salen de M1, M2, M4 y M8. Los integrantes se comparan por texto: "Ana" y "ana" son dos integrantes.
7. **`-race`.** `go test -race` no se corrió en la máquina local: requiere cgo y no hay `gcc`. Lo corre el CI (`go test -race -coverprofile=coverage.out ./...`).
8. **PostgreSQL.** Localmente se saltaron 34 tests por falta de `TEST_DATABASE_URL` (`cmd/server` 4, backlog 3, defect 10, effort 7, poker 2, project 4 y sprint 4). En el CI también se saltan: el workflow no levanta una base. HU-38 no toca la persistencia.
9. **Servidor de la prueba de punta a punta.** La primera carga de datos la respondió otro proceso que estaba escuchando en el puerto 8099: el servidor levantado para la prueba registró "servidor escuchando en :8099" 24 segundos después. Se repitió la carga completa contra ese servidor, y la respuesta de arriba es la de esa segunda carga.

## Validación final

Sobre `f58b5e7` (GREEN con `main` en `a30a0f8`, que trae TEC-06), en Windows con go1.27.1, sin base. Las cifras con HU-37 adentro están en "Integración con main (HU-37)".

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
ok   seguimiento-medicion-gpc/internal/httpapi       coverage: 98.7% of statements
ok   seguimiento-medicion-gpc/internal/metrics       coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/poker         coverage: 69.5% of statements
ok   seguimiento-medicion-gpc/internal/postgres      coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/postgrestest  coverage: 36.8% of statements
ok   seguimiento-medicion-gpc/internal/project       coverage: 67.5% of statements
ok   seguimiento-medicion-gpc/internal/report        coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/sprint        coverage: 76.3% of statements
     seguimiento-medicion-gpc/internal/web           coverage: 0.0% of statements
```

godog de HU-38: `21 scenarios (21 passed)`, `184 steps (184 passed)`. Con `go test -count=1 -v ./...`: 775 tests de nivel superior pasan, 0 fallan y 34 se saltan.

`go tool cover -func` (perfil de `go test -count=1 -p 1 -coverprofile=cover.out ./...`): `report.NewManager`, `report.Project`, `membersOf`, `projectReport`, `toProjectReportResponse`, `toReportMetricsResponse` y `NewHandler` en 100.0%. `internal/report` entra en 100.0% e `internal/httpapi` sube de 98.6% a 98.7%. Total del perfil sin base: **83.1%** (`main` en `a30a0f8`, mismo Go y comando, en una copia temporal con `git worktree`: **82.7%**). No se midió con PostgreSQL.

## Integración con main (HU-37)

Después de abrir el Pull Request, `main` avanzó con HU-37 (#147, merge `3837b5c`), que agrega `GET /api/proyectos/{id}/metricas/sprints` y los gráficos de la sección "Dashboard" y modifica `internal/httpapi/metrics.go`, `internal/metrics/summary.go`, `server.go`, la interfaz y los mismos documentos. Se trajo con `git merge --no-ff origin/main` en el commit `0f98b48` (`chore(HU-38): traer main con HU-37`).

Hubo conflictos en 6 archivos, y se resolvieron conservando los dos lados:

- `internal/httpapi/server_test.go`: las dos filas de `TestRutasConID_IDNoNumericoDevuelve400`, primero la de HU-37 y después la de HU-38.
- `internal/httpapi/web_test.go`: los tests de HU-37 y, a continuación, los 3 de HU-38.
- `internal/web/app.js`: el código de HU-37 y, a continuación, el bloque de HU-38 (`storyPointsReporte`, `filasReporte` y el handler de "Generar reporte"), sin cambios.
- `README.md`: el ítem "Dashboard" de HU-37 (con "Ver gráficos") y los ítems 6 "Reportes", 7 "Sprints" y 8 "Defectos" de HU-38; las filas de HU-37 y HU-38 en "Historias por sprint". Las rutas de la API se integraron sin conflicto.
- `docs/ia/registro.md`: las filas de HU-37 y HU-38.
- `docs/reportes/cobertura.md`: los párrafos de HU-38 y HU-37; la cabecera se actualizó con las cifras de abajo.

`server.go`, `index.html` y `styles.css` se integraron sin conflictos. Comparada con `main` (`git diff 3837b5c 0f98b48`), la rama suma a `app.js` y a `web_test.go` solo las líneas de HU-38 (44 y 66 agregadas, ninguna borrada) y a `server_test.go` la fila de HU-38. Lo que HU-37 cambió en `internal/httpapi/metrics.go` no afecta a `toReportMetricsResponse`: `TestReporteProyecto_CadaParteEsIgualASuEndpoint` sigue verificando que "metricas" es igual a la respuesta de `GET /api/proyectos/{id}/metricas`.

Validación sobre `0f98b48`, en Windows con go1.27.1, sin base:

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
ok   seguimiento-medicion-gpc/internal/httpapi       coverage: 98.7% of statements
ok   seguimiento-medicion-gpc/internal/metrics       coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/poker         coverage: 69.5% of statements
ok   seguimiento-medicion-gpc/internal/postgres      coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/postgrestest  coverage: 36.8% of statements
ok   seguimiento-medicion-gpc/internal/project       coverage: 67.5% of statements
ok   seguimiento-medicion-gpc/internal/report        coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/sprint        coverage: 76.3% of statements
     seguimiento-medicion-gpc/internal/web           coverage: 0.0% of statements
```

- Con `go test -count=1 -v ./...`: 800 tests de nivel superior pasan, 0 fallan y 34 se saltan (los de PostgreSQL). Pasan `TestFeatures_HU38ReporteProyecto` (`21 scenarios (21 passed)`, `184 steps (184 passed)`), `TestFeatures_HU37GraficosDashboard` y `TestRutasConID_IDNoNumericoDevuelve400`, con las filas de HU-37 y HU-38.
- La primera corrida de `go test -count=1 -v ./...` terminó con código 1 aunque todos los paquetes dieron `ok` y no hubo ningún `--- FAIL`, por un error de Windows al borrar el ejecutable temporal (`go: unlinkat …\server.test.exe: The process cannot access the file because it is being used by another process`), igual que en HU-35. La repetición dio código 0, con las mismas cifras.
- `go tool cover -func` (perfil de `go test -count=1 -p 1 -coverprofile=cover.out ./...`): `report.NewManager`, `report.Project`, `membersOf`, `projectReport`, `toProjectReportResponse` y `toReportMetricsResponse` siguen en 100.0%. `internal/report` 100.0%, `internal/httpapi` 98.7% e `internal/metrics` 100.0%. Total del perfil sin base: **83.4%** (`main` en `3837b5c`, con HU-37, mismo Go y comando, en una copia temporal con `git worktree`: **83.0%**, con `internal/httpapi` 98.7% e `internal/metrics` 100.0%). No se midió con PostgreSQL.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | f19bf62 | `docs(HU-38): agregar la especificación (SDD)` |
| RED | a396ae2 | `test(HU-38): agregar tests (RED)` |
| GREEN | 7d7bfb2 | `feat(HU-38): generar el reporte del proyecto (GREEN)` |
| Merge de `main` | f58b5e7 | `chore(HU-38): traer main con TEC-06` |
| Merge de `main` | 0f98b48 | `chore(HU-38): traer main con HU-37` |
| Refactor (después del cierre) | 14cffba | `refactor(HU-38): unificar la conversión de las métricas del proyecto` |

HU-38: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
