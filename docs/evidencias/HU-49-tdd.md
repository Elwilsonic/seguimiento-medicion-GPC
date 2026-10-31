# Evidencia TDD · HU-49

> Issue del repositorio: #170

## Historia

Como Agile Enabler, quiero ver en el reporte del proyecto o del sprint cada historia con su estado y su estimación, para saber qué se planificó, qué se completó y cuánto se estimó.

## Resultado

Se agregó el tipo `report.Stories` y el campo `Stories` en `ProjectReport` y `SprintReport`, que se arma con `backlog.Manager.ListItems` (HU-06) al final de cada reporte (RN-49.7):

- **Reporte del proyecto:** todas las historias del backlog.
- **Reporte del sprint:** las historias del proyecto con el `SprintID` del sprint.
- **Las dos partes:** cuentan las historias en `TERMINADO`.

`report.NewManager` recibe además el `*backlog.Manager`.

En la API, los dos reportes tienen la parte `"historias"` (`{historias, cantidad, completadas}`), después de `"metricas"`. Cada historia tiene el formato de `GET /api/items/{id}` (`toItemResponse`).

El PDF del reporte del proyecto tiene la sección "Historias" entre "Métricas" y "Esfuerzos". La sección "Reportes" de la interfaz muestra la tabla de historias y su cantidad en los dos bloques.

No cambian los repositorios, el esquema, las métricas ni las demás partes de los reportes.

## Trazabilidad

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-49.1 — historias del backlog, ordenadas por ID, con sus datos y cantidades | "El reporte del proyecto incluye cada historia del backlog con su estimación" | `TestProject_IncluyeLasHistoriasDelBacklog`; `TestReporteProyecto_CadaHistoriaEsIgualASuConsultaEnElBacklog`; `TestReporteProyecto_ContratoDelEjemploDeTEC03` |
| CA-49.2 — solo las historias del sprint | "El reporte de un sprint incluye solo las historias asignadas al sprint" | `TestSprint_IncluyeSoloLasHistoriasDelSprint`; `TestReporteSprint_CadaHistoriaEsIgualASuConsultaEnElBacklog`; `TestReporteSprint_ContratoDelEjemploDeTEC03` |
| CA-49.3 — cantidades iguales a las métricas e historias iguales al backlog | Los dos escenarios anteriores (pasos "coinciden con las métricas" y "es igual a su consulta en el backlog") | `TestProject_CantidadesDeHistoriasCoincidenConLasMetricas`; `TestSprint_IncluyeSoloLasHistoriasDelSprint`; los dos `…_CadaHistoriaEsIgualASuConsultaEnElBacklog` |
| CA-49.4 — sin historias; sin estimar, sin horas o sin sprint | "Un proyecto sin historias tiene la lista vacía y cantidades 0"; "Un sprint sin historias asignadas tiene la lista vacía"; "El PDF indica la historia sin estimar, sin horas y sin sprint" | `TestProject_SinHistorias`; `TestSprint_SinHistorias`; `TestProject_HistoriaSinEstimarSinHorasNiSprint`; `TestProjectReportLines_HistoriasConYSinEstimacion`; los dos `…_ContratoDel…Vacio`; `TestInterfaz_FilaDeHistoriaDelReporteIndicaSinEstimarYSinSprint` |
| CA-49.5 — error sin reporte parcial; no modifica; copias | "Si falla la lectura de las historias del proyecto, no hay reporte parcial"; "Si falla la lectura de las historias del sprint, no hay reporte parcial"; "El reporte de un proyecto inexistente no lee las historias"; "Modificar las historias del reporte no cambia lo guardado" | `TestProject_FallaLaLecturaDeLasHistorias`; `TestSprint_FallaLaLecturaDeLasHistorias`; `TestReports_LeenLasHistoriasUnaSolaVez`; `TestReports_HistoriasSonCopias`; `TestReporteProyecto_FallaLaLecturaDeLasHistoriasDevuelve500`; `TestReporteSprint_FallaLaLecturaDeLasHistoriasDevuelve500` |
| CA-49.6 — sección "Historias" del PDF | "Un proyecto sin historias tiene la lista vacía y cantidades 0"; "El PDF indica la historia sin estimar, sin horas y sin sprint" | `TestProjectReportLines_HistoriasConYSinEstimacion`; `TestProjectReportLines_EjemploDeTEC03`; `TestProjectReportLines_ProyectoSinDatos`; `TestWriteProjectPDF_GeneraUnPDFConCadaLineaDelReporte`; `TestWriteProjectPDF_ProyectoSinDatos` |
| CA-49.7 — tabla de historias en la interfaz | Solo interfaz (ver el comentario inicial del `.feature`) | `TestInterfaz_ReportesTienenLaTablaDeHistorias`; `TestInterfaz_GenerarLosReportesMuestraLasHistorias`; `TestInterfaz_FilaDeHistoriaDelReporteIndicaSinEstimarYSinSprint`; prueba en el navegador |

La correspondencia no es 1 a 1:
- varios escenarios y tests cubren más de un criterio;
- CA-49.7 se verifica solo con tests de interfaz;
- el escenario "El reporte de un proyecto inexistente no lee las historias" cubre la tabla "Condiciones de error" de la SDD.

## Conteos

```text
$ grep -cE '^[[:space:]]*Escenario:' features/HU-49-historias-en-reporte.feature
9
$ grep -c '^func Test' internal/report/stories_test.go internal/report/stories_features_test.go
internal/report/stories_test.go:10
internal/report/stories_features_test.go:1
$ git diff 9e85d72 e136df7 -- <archivo> | grep -c '^+func Test'
internal/report/pdf_test.go 1
internal/httpapi/reports_test.go 2
internal/httpapi/sprint_reports_test.go 2
internal/httpapi/web_test.go 3
```

**19 funciones `Test` nuevas, que cubren 9 escenarios BDD (sin esquemas):**
- 10 de dominio en `stories_test.go`;
- 1 runner de godog;
- 1 del PDF;
- 4 de API;
- 3 de interfaz.

**Tests existentes que cambian, como pide la SDD ("Restricciones"):**
- los 4 contratos JSON de los reportes suman la parte `"historias"`;
- las líneas esperadas del PDF (`tec03Lines` y `emptyProjectLines`) suman la sección "Historias";
- los dos entornos de prueba (`newReportEnv` y `newSprintReportEnv`) usan la firma nueva de `NewManager`;
- el helper `item` delega en `story`, que recibe la prioridad.

No se aflojó ninguna aserción.

## Fase RED (`6890150`)

Stubs del RED, como fija la SDD:
- el tipo `Stories`;
- el campo `Stories` vacío en los dos reportes;
- la firma nueva de `NewManager`;
- el armado en `internal/httpapi/server.go`, que es una llamada a esa firma.

La lectura de las historias entra en el GREEN.

```text
$ git diff --stat 9e85d72 6890150 -- ':!*_test.go' ':!*.feature' ':!docs'
 internal/httpapi/server.go |  5 +++--
 internal/report/manager.go | 16 +++++++++-------
 internal/report/project.go |  1 +
 internal/report/sprint.go  |  1 +
 internal/report/stories.go | 11 +++++++++++
 5 files changed, 25 insertions(+), 9 deletions(-)
```

Salida de `go vet ./...` (sin salida) y de `go test ./...` sobre el RED, con los tiempos recortados:

```text
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
--- FAIL: TestReporteProyecto_ContratoDelEjemploDeTEC03
--- FAIL: TestReporteProyecto_ContratoDelProyectoVacio
--- FAIL: TestReporteProyecto_CadaHistoriaEsIgualASuConsultaEnElBacklog
--- FAIL: TestReporteProyecto_FallaLaLecturaDeLasHistoriasDevuelve500
--- FAIL: TestReporteSprint_ContratoDelEjemploDeTEC03
--- FAIL: TestReporteSprint_ContratoDelSprintVacio
--- FAIL: TestReporteSprint_CadaHistoriaEsIgualASuConsultaEnElBacklog
--- FAIL: TestReporteSprint_FallaLaLecturaDeLasHistoriasDevuelve500
--- FAIL: TestInterfaz_ReportesTienenLaTablaDeHistorias
--- FAIL: TestInterfaz_GenerarLosReportesMuestraLasHistorias
--- FAIL: TestInterfaz_FilaDeHistoriaDelReporteIndicaSinEstimarYSinSprint
FAIL seguimiento-medicion-gpc/internal/httpapi
ok   seguimiento-medicion-gpc/internal/member
ok   seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
--- FAIL: TestProjectReportLines_EjemploDeTEC03
--- FAIL: TestProjectReportLines_ProyectoSinDatos
--- FAIL: TestProjectReportLines_HistoriasConYSinEstimacion
--- FAIL: TestWriteProjectPDF_GeneraUnPDFConCadaLineaDelReporte
--- FAIL: TestWriteProjectPDF_ProyectoSinDatos
--- FAIL: TestFeatures_HU49HistoriasEnReporte
--- FAIL: TestProject_IncluyeLasHistoriasDelBacklog
--- FAIL: TestProject_HistoriaSinEstimarSinHorasNiSprint
--- FAIL: TestProject_CantidadesDeHistoriasCoincidenConLasMetricas
--- FAIL: TestSprint_IncluyeSoloLasHistoriasDelSprint
--- FAIL: TestProject_SinHistorias
--- FAIL: TestSprint_SinHistorias
--- FAIL: TestProject_FallaLaLecturaDeLasHistorias
--- FAIL: TestSprint_FallaLaLecturaDeLasHistorias
--- FAIL: TestReports_LeenLasHistoriasUnaSolaVez
--- FAIL: TestReports_HistoriasSonCopias
FAIL seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
```

Fallan **27 tests de primer nivel, todos de HU-49**:
- **11 en `internal/httpapi`:** los 8 de API (4 nuevos y los 4 contratos con la parte nueva) y los 3 de interfaz.
- **16 en `internal/report`:** los 10 de `stories_test.go`, el runner de godog, el test nuevo del PDF y los 4 del PDF con la sección nueva.

Son las 19 funciones nuevas y los 8 tests existentes que cambian. En la salida completa, `grep -c '^--- FAIL:'` da 27 y `grep -ciE 'build failed|undefined:|cannot use|\[setup failed\]|panic'` da 0.

Todos los fallos son de aserción:
- **Dominio:** las historias vienen vacías o en `nil`, y no hay error cuando falla la lectura, porque todavía no se leen.
- **API:** falta la parte `"historias"`, y el reporte responde 200 cuando falla la lectura.
- **PDF:** falta la sección "Historias".
- **Interfaz:** faltan los ids, los textos y `filaHistoriaReporte`.

Godog en el RED:
- **HU-49:** `9 scenarios (1 passed, 8 failed)` y `61 steps (41 passed, 8 failed, 12 skipped)`, sin pasos sin definir. El que pasa es "El reporte de un proyecto inexistente no lee las historias", que ya se cumplía, porque el proyecto se valida primero (HU-38).
- **HU-38 y HU-39:** siguieron en verde, con `21 scenarios (21 passed)` y `17 scenarios (17 passed)`.
- **HU-40:** también siguió en verde, porque sus pasos buscan líneas en el PDF y no comparan la lista completa.

## Corrección de dos tests del RED (`e20c300`)

Con el GREEN escrito, siguieron fallando `TestProject_FallaLaLecturaDeLasHistorias`, `TestSprint_FallaLaLecturaDeLasHistorias` y el paso BDD "las historias se leyeron después de los defectos". El código estaba bien, pero el valor esperado no:

```text
stories_test.go:223: lecturas de esfuerzos y defectos = 6 y 2; se esperaba 3 y 1, antes de las historias
stories_test.go:243: lecturas de esfuerzos y defectos = 2 y 2; se esperaba 1 y 1, antes de las historias
```

Los tests esperaban una lectura de esfuerzos por sprint y una de defectos. Pero las métricas, que el reporte lee antes, también leen los esfuerzos (M5) y los defectos (M9 y M10), así que esas lecturas son el doble. `TestProject_ErroresEnElOrdenDeLectura` de HU-38 ya contaba así (`2 * n` esfuerzos).

La corrección cuenta las lecturas que hace el reporte antes de las historias, con las mismas consultas (`readsBeforeStories`), en lugar de fijar el número. Así el test verifica lo mismo que pedía (todas las lecturas anteriores ya se hicieron cuando fallan las historias) y no depende de cómo lean las métricas.

Va en un commit propio, antes del GREEN, como desviación (acción 2 de la Retrospectiva del Sprint 3). Antes de esa corrección, los tres seguían fallando con el RED, por el error que no llegaba (`err = <nil>`).

```text
$ git diff --stat 6890150 e20c300
 internal/report/stories_features_test.go | 16 +++++++---
 internal/report/stories_test.go          | 53 +++++++++++++++++++-------------
 2 files changed, 43 insertions(+), 26 deletions(-)
```

## Fase GREEN (`e136df7`)

- **`internal/report/stories.go`:** `storiesOf` arma una lista nueva (nunca `nil`) con las historias que cumplen la condición y cuenta las `TERMINADO`.
- **`internal/report/project.go` y `sprint.go`:** leen las historias con `ListItems` al final, después de los defectos, y cortan en el error.
  - El proyecto toma todas.
  - El sprint toma las de su `SprintID`.
- **`internal/report/pdf.go`:** la sección "Historias" y `storyText`, que escribe "sin estimar", "sin horas estimadas" y "sin sprint" para los ceros. Las horas van con un decimal, como las métricas.
- **`internal/httpapi/reports.go` y `sprint_reports.go`:** `storyListResponse` y `toStoryListResponse`, que convierte cada historia con `toItemResponse`.
- **`internal/web/index.html` y `app.js`:**
  - la tabla y la cantidad en los dos bloques;
  - `filaHistoriaReporte`, con "—" como la tabla del backlog;
  - `cantidadHistorias`;
  - todo reutiliza `filasReporte` y `celda`, sin `innerHTML`.

El GREEN no modifica tests ni el `.feature`:

```text
$ git diff --stat e20c300 e136df7 -- '*_test.go' '*.feature'
(sin salida)
```

## Refactor

No se hizo. El código nuevo reutiliza lo que ya existía (`ListItems`, `toItemResponse`, `filasReporte`, `oneDecimal`) y no se encontró una mejora clara.

## Validación

```text
$ gofmt -l ./cmd ./internal
(sin salida)
$ go vet ./...
(sin salida)
$ go test ./...
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
ok   seguimiento-medicion-gpc/internal/httpapi
ok   seguimiento-medicion-gpc/internal/member
ok   seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
```

Godog de HU-49 en el GREEN: `9 scenarios (9 passed)` y `61 steps (61 passed)`.

Con PostgreSQL 16.15 y `TEST_DATABASE_URL`, `go test -race -p 1 -count=1 ./...` (como el CI) dio los 14 paquetes en `ok`. Con `-v`: 0 `--- SKIP`, 1127 `--- PASS` de primer nivel y 0 `--- FAIL`.

## Prueba en el navegador y de punta a punta

Con el servidor en memoria (`PORT=8099`), se cargó por la API un proyecto con dos historias:
- "Registrar venta": ALTA, 3 SP, 8 h, en el sprint 1 y `TERMINADO`;
- "Consultar stock": MEDIA, sin estimar, sin horas y sin sprint.

Respuesta real de la parte `"historias"` de `GET /api/proyectos/1/reporte`:

```text
{"historias": [{"id": 1, "proyecto_id": 1, "titulo": "Registrar venta", "descripcion": "", "prioridad": "ALTA", "criterios": ["c"], "estado": "TERMINADO", "story_points": 3, "horas_estimadas": 8, "sprint_id": 1}, {"id": 2, "proyecto_id": 1, "titulo": "Consultar stock", "descripcion": "", "prioridad": "MEDIA", "criterios": ["c"], "estado": "PENDIENTE", "story_points": 0, "horas_estimadas": 0, "sprint_id": 0}], "cantidad": 2, "completadas": 1}
```

En `GET /api/sprints/1/reporte`, la parte `"historias"` tuvo `cantidad` 1 y `completadas` 1: la historia sin sprint no aparece.

Texto de la sección del PDF descargado, con `pdftotext`:

```text
Historias
#1 · Registrar venta · ALTA · TERMINADO · 3 SP · 8.0 h estimadas · sprint 1
#2 · Consultar stock · MEDIA · PENDIENTE · sin estimar · sin horas estimadas · sin sprint
Cantidad: 2 historia(s), 1 completada(s)
```

En la interfaz, con el proyecto y el sprint actuales en 1, los dos botones de reporte llenaron las tablas. Texto leído del DOM con Playwright:
- **Proyecto:** las dos filas, con "—" en SP y en Sprint de "Consultar stock", y "2 historia(s), 1 completada(s)";
- **Sprint:** una fila y "1 historia(s), 1 completada(s)".

También se volvió a recorrer el caso del manual de usuario. Su captura 18 se actualizó con la tabla nueva:

![Reporte del proyecto y del sprint con la tabla de historias](../img/manual-de-usuario/18-reportes.png)

## Cobertura

Medición local, en Linux con go1.27.1, sobre `e136df7`. `main` (`9e85d72`) se midió con los mismos comandos, en una copia temporal con `git worktree`.

| Medición | Rama (`e136df7`) | `main` (`9e85d72`) |
|---|---|---|
| Sin base: `go test -count=1 -p 1 -coverprofile=cover.out ./...`, total del perfil | **84.7%** | **84.5%** |
| Con PostgreSQL 16.15: `-coverpkg=./...`, total | **96.2%** | **96.2%** |
| `internal/report` (`go test -cover`) | 100.0% | 100.0% |
| `internal/httpapi` (`go test -cover`) | 98.7% | 98.7% |

Salida de `go test -cover ./...` en la rama, sin base:

```text
ok   seguimiento-medicion-gpc/cmd/server             coverage: 29.8% of statements
ok   seguimiento-medicion-gpc/internal/backlog       coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/dashboard     coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/defect        coverage: 77.7% of statements
ok   seguimiento-medicion-gpc/internal/effort        coverage: 74.2% of statements
ok   seguimiento-medicion-gpc/internal/httpapi       coverage: 98.7% of statements
ok   seguimiento-medicion-gpc/internal/member        coverage: 68.8% of statements
ok   seguimiento-medicion-gpc/internal/metrics       coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/poker         coverage: 75.5% of statements
ok   seguimiento-medicion-gpc/internal/postgres      coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/postgrestest  coverage: 36.8% of statements
ok   seguimiento-medicion-gpc/internal/project       coverage: 74.8% of statements
ok   seguimiento-medicion-gpc/internal/report        coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/sprint        coverage: 77.1% of statements
     seguimiento-medicion-gpc/internal/web           coverage: 0.0% of statements
```

Funciones nuevas o cambiadas (`go tool cover -func`):

```text
internal/httpapi/reports.go:40:   toStoryListResponse   100.0%
internal/report/pdf.go:28:        projectReportLines    100.0%
internal/report/pdf.go:130:       storyText             100.0%
internal/report/project.go:49:    Project               100.0%
internal/report/sprint.go:28:     Sprint                100.0%
internal/report/stories.go:16:    storiesOf             100.0%
total:                            (statements)          84.7%
```

El total sin base sube de 84.5% a 84.7%. Con base queda igual (96.2%), porque HU-49 no toca la persistencia. La interfaz (`app.js`) no tiene cobertura de Go: la verifican los tests de contenido de `web_test.go` y la prueba en el navegador.

## Desvíos y límites

1. **Corrección de dos tests del RED.** Ver la sección "Corrección de dos tests del RED". El valor esperado no consideraba las lecturas de las métricas. La corrección no cambia lo que se verifica y va en un commit propio, antes del GREEN.
2. **Historia agregada después de la Review.** HU-49 se creó el 31/10, después de la Review del Sprint 4, al comparar el proyecto con la consigna (requerimiento 9). Por decisión del Agile Enabler se suma al Sprint 4. La SDD lo registra en "Comparación con las actas", y el acta de la Review y el informe final tienen una sección con el cambio.
3. **Tests de interfaz de contenido.** Como en las demás historias, `web_test.go` verifica el texto de `index.html` y `app.js`, sin ejecutar JavaScript. Lo que muestra la tabla se verificó en el navegador con Playwright.
4. **Límite conocido de la SDD.** Las lecturas no son atómicas. Si una historia cambia de estado entre la lectura de las métricas y la de las historias, `completadas` puede no coincidir con `items_terminados` en ese reporte.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `983ca81` | `docs(HU-49): agregar especificación SDD` |
| RED | `6890150` | `test(HU-49): agregar tests (RED)` |
| Corrección de tests del RED | `e20c300` | `test(HU-49): corregir las lecturas esperadas antes de las historias` |
| GREEN | `e136df7` | `feat(HU-49): mostrar las historias y sus estimaciones en los reportes (GREEN)` |

HU-49: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
