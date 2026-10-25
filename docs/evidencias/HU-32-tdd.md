# Evidencia TDD — HU-32 Calcular Story Points planificados, completados y velocidad

- **Historia:** HU-32 — Calcular Story Points planificados, completados y velocidad
- **Issue:** [#50](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/50)
- **Rama:** `feature/HU-32-sp-velocidad`
- **Fecha:** 25/10/2026
- **Entorno:** Linux/amd64, go1.27.1. Las corridas con base usan PostgreSQL 16 (el `docker-compose.yml` usa 17; para estos tests no cambia nada) con la base `gpc_test` creada desde `db/init/01-schema.sql`.

La SDD (`docs/especificaciones/HU-32-sp-velocidad.md`) se aprobó antes de escribir el BDD y los tests. Después de la aprobación se ajustó una línea, también aprobada antes del commit RED: `httpapi.NewHandler` crea el `metrics.Manager` en lugar de `cmd/server`.

## Trazabilidad

HU-32: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

| Criterio de aceptación | Escenarios BDD | Tests |
|---|---|---|
| CA-32.1 Planificados del sprint y del proyecto | "Story Points de un sprint con el ejemplo de TEC-03", "Story Points del proyecto con todo el backlog" | `TestSprintStoryPoints_EjemploDeTEC03`, `TestSprintStoryPoints_SoloCuentaLosItemsDelSprint`, `TestProjectStoryPoints_UsaTodoElBacklog` |
| CA-32.2 Completados | Los mismos dos | `TestSprintStoryPoints_EjemploDeTEC03`, `TestProjectStoryPoints_UsaTodoElBacklog` |
| CA-32.3 Ítems sin estimar | Los mismos dos | `TestSprintStoryPoints_EjemploDeTEC03`, `TestProjectStoryPoints_UsaTodoElBacklog` |
| CA-32.4 Sin ítems devuelve cero | "Un sprint sin ítems devuelve cero", "Un proyecto sin ítems devuelve cero" | `TestSprintStoryPoints_SprintSinItemsDevuelveCero`, `TestProjectStoryPoints_ProyectoSinItemsDevuelveCero`, `TestListByProject_SinSprintsDevuelveListaVacia` |
| CA-32.5 Velocidad con los últimos 3 sprints cerrados | Esquema "Velocidad con los últimos sprints cerrados" (4 ejemplos), "Un sprint abierto no cuenta para la velocidad" | `TestVelocity_UltimosTresSprintsCerrados` (7 casos), `TestProjectStoryPoints_VelocidadConLosUltimosSprintsCerrados`, `TestListByProject_DevuelveLosSprintsDelProyecto` |
| CA-32.6 Velocidad no disponible | "Un proyecto sin ítems devuelve cero", "Un sprint abierto no cuenta para la velocidad" | `TestVelocity_UltimosTresSprintsCerrados` (caso sin sprints), `TestProjectStoryPoints_SprintAbiertoNoCuentaParaLaVelocidad`, `TestProjectStoryPoints_ProyectoSinItemsDevuelveCero` |
| CA-32.7 Inexistentes, errores del repositorio y sin modificar nada | Esquemas "Consultar un sprint inexistente falla" y "Consultar un proyecto inexistente falla" (3 ejemplos cada uno), "Consultar las métricas no modifica los ítems ni los sprints" | `TestSprintStoryPoints_SprintInexistenteDevuelveErrSprintNotFound`, `TestProjectStoryPoints_ProyectoInexistenteDevuelveErrProjectNotFound`, `TestProjectStoryPoints_ErrorDelRepositorioDeSprintsSePropaga`, `TestStoryPoints_ErrorDelRepositorioDeItemsSePropaga`, `TestStoryPoints_NoModificaNada`, `TestVelocity_NoModificaLaLista`, `TestGet_DevuelveElSprint`, `TestGet_SprintInexistenteDevuelveErrSprintNotFound` |
| CA-32.8 Sprints de un proyecto en memoria y PostgreSQL | — (se cubre con tests de Go) | 3 subtests nuevos de `runRepositoryContract` (`ListByProject`), que corren en `TestMemoryRepository_Contrato` y `TestPostgresRepository_Contrato` |
| CA-32.9 API e interfaz | — (se cubre con tests de Go) | `TestMetricasSprint_Devuelve200ConLosStoryPoints`, `TestMetricasProyecto_Devuelve200ConLaVelocidadRedondeada`, `TestMetricasProyecto_SinSprintsCerradosLaVelocidadEsNull`, `TestMetricasSprint_InexistenteDevuelve404`, `TestMetricasProyecto_InexistenteDevuelve404`, `TestMetricas_ErrorDelRepositorioDevuelve500SinDetalle`, 2 filas de `TestRutasConID_IDNoNumericoDevuelve400`, `TestInterfaz_SeccionDeMetricas` |

## Alcance de las pruebas

- **BDD:** `features/HU-32-sp-velocidad.feature` tiene **9 definiciones** (6 escenarios y 3 esquemas) que godog ejecuta como **16 escenarios** (72 pasos). Cada una lleva su CA en el nombre.
- **Tests:** **26 funciones `Test` nuevas** (incluido el runner de godog), contadas con `grep -c '^func Test'` en los archivos nuevos y `git diff -U0 | grep -c '^+func Test'` en `web_test.go`, más **3 subtests de contrato** nuevos y **2 filas** nuevas en la tabla de `TestRutasConID_IDNoNumericoDevuelve400`:

| Archivo | Tests nuevos |
|---|---|
| `internal/metrics/points_test.go` | 2 |
| `internal/metrics/story_points_test.go` | 12 |
| `internal/metrics/story_points_features_test.go` | 1 (runner de godog) |
| `internal/sprint/list_test.go` | 4 |
| `internal/httpapi/metrics_test.go` | 6 |
| `internal/httpapi/web_test.go` | 1 |
| `internal/sprint/repository_contract_test.go` | 3 subtests |
| `internal/httpapi/server_test.go` | 2 filas de tabla |

No hay una correspondencia 1 a 1: son 26 tests para 16 escenarios ejecutados, y CA-32.8 y CA-32.9 se cubren solo con tests de Go.

## Fase RED (`443c3ff`)

El RED compila con stubs mínimos sin lógica (acción 1 de la Retrospectiva del Sprint 2), así que los tests fallan por sus aserciones y no por compilación:

- `internal/metrics`: los tipos `SprintPoints`, `ProjectPoints` y `closedSprint`, `NewManager`, y `velocity`, `SprintStoryPoints` y `ProjectStoryPoints` que devuelven cero o "no implementado".
- `internal/sprint`: `Manager.Get`, `Manager.ListByProject` y `ListByProject` en la interfaz `Repository` y en los dos repositorios, que devuelven "no implementado".
- `internal/httpapi`: las dos rutas con handlers que responden 501.
- Para que compilara hubo que agregar `ListByProject` al fake de test `failingRepository` (`internal/sprint/create_test.go`), que implementa la interfaz método por método.

`cmd/server` no se tocó, ni en el RED ni en el GREEN. `go vet ./...` pasó antes del commit RED.

Sin base (las líneas de resultado y de fallo; se omite el detalle de godog):

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.015s
ok  	seguimiento-medicion-gpc/internal/backlog	0.388s
ok  	seguimiento-medicion-gpc/internal/effort	0.110s
--- FAIL: TestMetricasSprint_Devuelve200ConLosStoryPoints (0.00s)
    metrics_test.go:87: GET /api/sprints/1/metricas: Content-Type = "", se esperaba application/json
    metrics_test.go:89: código = 501, se esperaba 200; cuerpo: 
--- FAIL: TestMetricasProyecto_Devuelve200ConLaVelocidadRedondeada (0.00s)
    metrics_test.go:107: GET /api/proyectos/1/metricas: Content-Type = "", se esperaba application/json
    metrics_test.go:109: código = 501, se esperaba 200; cuerpo: 
--- FAIL: TestMetricasProyecto_SinSprintsCerradosLaVelocidadEsNull (0.00s)
    metrics_test.go:126: GET /api/proyectos/1/metricas: Content-Type = "", se esperaba application/json
    metrics_test.go:128: código = 501, se esperaba 200; cuerpo: 
--- FAIL: TestMetricasSprint_InexistenteDevuelve404 (0.00s)
    metrics_test.go:138: GET /api/sprints/99/metricas: Content-Type = "", se esperaba application/json
    metrics_test.go:140: código = 501, se esperaba 404; cuerpo: 
--- FAIL: TestMetricasProyecto_InexistenteDevuelve404 (0.00s)
    metrics_test.go:145: GET /api/proyectos/99/metricas: Content-Type = "", se esperaba application/json
    metrics_test.go:147: código = 501, se esperaba 404; cuerpo: 
--- FAIL: TestMetricas_ErrorDelRepositorioDevuelve500SinDetalle (0.00s)
    metrics_test.go:162: GET /api/sprints/1/metricas: Content-Type = "", se esperaba application/json
    metrics_test.go:164: código = 501, se esperaba 500; cuerpo: 
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.00s)
    --- FAIL: TestRutasConID_IDNoNumericoDevuelve400/métricas_del_sprint (0.00s)
        server_test.go:212: GET /api/sprints/abc/metricas: Content-Type = "", se esperaba application/json
        server_test.go:214: código = 501, se esperaba 400; cuerpo: 
    --- FAIL: TestRutasConID_IDNoNumericoDevuelve400/métricas_del_proyecto (0.00s)
        server_test.go:212: GET /api/proyectos/abc/metricas: Content-Type = "", se esperaba application/json
        server_test.go:214: código = 501, se esperaba 400; cuerpo: 
--- FAIL: TestInterfaz_SeccionDeMetricas (0.00s)
    web_test.go:161: la página no contiene "Métricas"
    web_test.go:161: la página no contiene "Ver métricas"
    web_test.go:161: la página no contiene "id=\"datos-metricas\""
    web_test.go:167: app.js no contiene "\"/metricas\""
    web_test.go:167: app.js no contiene "sp_planificados"
    web_test.go:167: app.js no contiene "sp_completados"
    web_test.go:167: app.js no contiene "items_sin_estimar"
    web_test.go:167: app.js no contiene "velocidad"
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi	0.124s
--- FAIL: TestVelocity_UltimosTresSprintsCerrados (0.00s)
    --- FAIL: TestVelocity_UltimosTresSprintsCerrados/un_sprint_cerrado (0.00s)
        points_test.go:31: velocity = 0 con 0 sprints, se esperaba 8 con 1
    --- FAIL: TestVelocity_UltimosTresSprintsCerrados/ejemplo_de_TEC-03 (0.00s)
        points_test.go:31: velocity = 0 con 0 sprints, se esperaba 10 con 3
    --- FAIL: TestVelocity_UltimosTresSprintsCerrados/cuatro_cerrados_usa_los_IDs_4,_3_y_2 (0.00s)
        points_test.go:31: velocity = 0 con 0 sprints, se esperaba 10 con 3
    --- FAIL: TestVelocity_UltimosTresSprintsCerrados/el_orden_de_entrada_no_importa (0.00s)
        points_test.go:31: velocity = 0 con 0 sprints, se esperaba 10 con 3
    --- FAIL: TestVelocity_UltimosTresSprintsCerrados/un_sprint_cerrado_sin_ítems_cuenta_con_0 (0.00s)
        points_test.go:31: velocity = 0 con 0 sprints, se esperaba 5 con 2
    --- FAIL: TestVelocity_UltimosTresSprintsCerrados/sin_redondear (0.00s)
        points_test.go:31: velocity = 0 con 0 sprints, se esperaba 12.333333333333334 con 3
16 scenarios (16 failed)
72 steps (47 passed, 16 failed, 9 skipped)
--- FAIL: TestFeatures_HU32CalcularStoryPoints (0.06s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Story_Points_de_un_sprint_con_el_ejemplo_de_TEC-03_(CA-32.1,_CA-32.2_y_CA-32.3) (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Un_sprint_sin_ítems_devuelve_cero_(CA-32.4) (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Story_Points_del_proyecto_con_todo_el_backlog_(CA-32.1,_CA-32.2_y_CA-32.3) (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Un_proyecto_sin_ítems_devuelve_cero_(CA-32.4_y_CA-32.6) (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Velocidad_con_los_últimos_sprints_cerrados_(CA-32.5) (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Velocidad_con_los_últimos_sprints_cerrados_(CA-32.5)#01 (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Velocidad_con_los_últimos_sprints_cerrados_(CA-32.5)#02 (0.01s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Velocidad_con_los_últimos_sprints_cerrados_(CA-32.5)#03 (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Un_sprint_abierto_no_cuenta_para_la_velocidad_(CA-32.5_y_CA-32.6) (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Consultar_un_sprint_inexistente_falla_(CA-32.7) (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Consultar_un_sprint_inexistente_falla_(CA-32.7)#01 (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Consultar_un_sprint_inexistente_falla_(CA-32.7)#02 (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Consultar_un_proyecto_inexistente_falla_(CA-32.7) (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Consultar_un_proyecto_inexistente_falla_(CA-32.7)#01 (0.01s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Consultar_un_proyecto_inexistente_falla_(CA-32.7)#02 (0.00s)
    --- FAIL: TestFeatures_HU32CalcularStoryPoints/Consultar_las_métricas_no_modifica_los_ítems_ni_los_sprints_(CA-32.7) (0.00s)
    story_points_features_test.go:32: fallaron escenarios BDD de HU-32
--- FAIL: TestSprintStoryPoints_EjemploDeTEC03 (0.00s)
    story_points_test.go:149: SprintStoryPoints devolvió un error inesperado: no implementado
--- FAIL: TestSprintStoryPoints_SoloCuentaLosItemsDelSprint (0.00s)
    story_points_test.go:169: SprintStoryPoints devolvió un error inesperado: no implementado
--- FAIL: TestSprintStoryPoints_SprintSinItemsDevuelveCero (0.00s)
    story_points_test.go:184: SprintStoryPoints devolvió un error inesperado: no implementado
--- FAIL: TestSprintStoryPoints_SprintInexistenteDevuelveErrSprintNotFound (0.00s)
    story_points_test.go:198: SprintStoryPoints(99): err = no implementado, se esperaba no existe un sprint con ese id
    story_points_test.go:198: SprintStoryPoints(0): err = no implementado, se esperaba no existe un sprint con ese id
    story_points_test.go:198: SprintStoryPoints(-1): err = no implementado, se esperaba no existe un sprint con ese id
--- FAIL: TestProjectStoryPoints_UsaTodoElBacklog (0.00s)
    story_points_test.go:215: ProjectStoryPoints devolvió un error inesperado: no implementado
--- FAIL: TestProjectStoryPoints_ProyectoSinItemsDevuelveCero (0.00s)
    story_points_test.go:230: ProjectStoryPoints devolvió un error inesperado: no implementado
--- FAIL: TestProjectStoryPoints_VelocidadConLosUltimosSprintsCerrados (0.00s)
    story_points_test.go:251: ProjectStoryPoints devolvió un error inesperado: no implementado
--- FAIL: TestProjectStoryPoints_SprintAbiertoNoCuentaParaLaVelocidad (0.00s)
    story_points_test.go:269: ProjectStoryPoints devolvió un error inesperado: no implementado
--- FAIL: TestProjectStoryPoints_ProyectoInexistenteDevuelveErrProjectNotFound (0.00s)
    story_points_test.go:283: ProjectStoryPoints(99): err = no implementado, se esperaba no existe un proyecto con ese id
    story_points_test.go:283: ProjectStoryPoints(0): err = no implementado, se esperaba no existe un proyecto con ese id
    story_points_test.go:283: ProjectStoryPoints(-1): err = no implementado, se esperaba no existe un proyecto con ese id
--- FAIL: TestStoryPoints_NoModificaNada (0.00s)
    story_points_test.go:298: SprintStoryPoints devolvió un error inesperado: no implementado
--- FAIL: TestProjectStoryPoints_ErrorDelRepositorioDeSprintsSePropaga (0.00s)
    story_points_test.go:339: err = no implementado, se esperaba falla de sprints
--- FAIL: TestStoryPoints_ErrorDelRepositorioDeItemsSePropaga (0.00s)
    story_points_test.go:352: SprintStoryPoints: err = no implementado, se esperaba falla de ítems
    story_points_test.go:355: ProjectStoryPoints: err = no implementado, se esperaba falla de ítems
FAIL
FAIL	seguimiento-medicion-gpc/internal/metrics	0.083s
ok  	seguimiento-medicion-gpc/internal/poker	0.068s
ok  	seguimiento-medicion-gpc/internal/postgres	0.017s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.008s
ok  	seguimiento-medicion-gpc/internal/project	0.088s
--- FAIL: TestGet_DevuelveElSprint (0.00s)
    list_test.go:25: Get devolvió un error inesperado: no implementado
--- FAIL: TestGet_SprintInexistenteDevuelveErrSprintNotFound (0.00s)
    list_test.go:43: Get(99): err = no implementado, se esperaba no existe un sprint con ese id
    list_test.go:43: Get(0): err = no implementado, se esperaba no existe un sprint con ese id
    list_test.go:43: Get(-1): err = no implementado, se esperaba no existe un sprint con ese id
--- FAIL: TestListByProject_DevuelveLosSprintsDelProyecto (0.00s)
    list_test.go:73: ListByProject devolvió un error inesperado: no implementado
--- FAIL: TestListByProject_SinSprintsDevuelveListaVacia (0.00s)
    list_test.go:87: ListByProject devolvió un error inesperado: no implementado
--- FAIL: TestMemoryRepository_Contrato (0.00s)
    --- FAIL: TestMemoryRepository_Contrato/ListByProject_sin_sprints_devuelve_una_lista_vacía (0.00s)
        repository_contract_test.go:191: ListByProject devolvió un error inesperado: no implementado
    --- FAIL: TestMemoryRepository_Contrato/ListByProject_devuelve_los_sprints_del_proyecto_ordenados_por_ID (0.00s)
        repository_contract_test.go:206: ListByProject devolvió un error inesperado: no implementado
    --- FAIL: TestMemoryRepository_Contrato/ListByProject_devuelve_una_lista_nueva (0.00s)
        repository_contract_test.go:221: ListByProject = [], no implementado; se esperaba un sprint
FAIL
FAIL	seguimiento-medicion-gpc/internal/sprint	0.188s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

Fallan solo los tests nuevos. En godog, los 16 escenarios fallan en el primer paso "Entonces" y los 47 pasos "Dado" pasan: la preparación de los datos funciona y lo que falla es lo que se prueba. Los paquetes que no toca la historia pasan.

Con base (PostgreSQL 16), además fallan los 3 subtests nuevos del contrato PostgreSQL:

```
$ TEST_DATABASE_URL="postgres://gpc:***@127.0.0.1:5432/gpc_test?sslmode=disable" go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.434s
ok  	seguimiento-medicion-gpc/internal/backlog	2.137s
ok  	seguimiento-medicion-gpc/internal/effort	2.230s
FAIL	seguimiento-medicion-gpc/internal/httpapi	0.107s
FAIL	seguimiento-medicion-gpc/internal/metrics	0.042s
ok  	seguimiento-medicion-gpc/internal/poker	0.882s
ok  	seguimiento-medicion-gpc/internal/postgres	0.007s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.009s
ok  	seguimiento-medicion-gpc/internal/project	0.555s
--- FAIL: TestPostgresRepository_Contrato (0.60s)
    --- FAIL: TestPostgresRepository_Contrato/ListByProject_sin_sprints_devuelve_una_lista_vacía (0.04s)
    --- FAIL: TestPostgresRepository_Contrato/ListByProject_devuelve_los_sprints_del_proyecto_ordenados_por_ID (0.04s)
    --- FAIL: TestPostgresRepository_Contrato/ListByProject_devuelve_una_lista_nueva (0.04s)
FAIL	seguimiento-medicion-gpc/internal/sprint	1.041s
```

## Fase GREEN (`907e10d`)

El commit GREEN solo cambia código de producción (`internal/metrics`, `internal/sprint`, `internal/httpapi`, `internal/web`); ningún test cambió desde el RED.

- `storyPoints` suma M1 y M2 y cuenta los ítems sin estimar; `velocity` promedia M2 de los últimos 3 sprints cerrados por ID descendente, sobre una copia de la lista.
- `SprintStoryPoints` valida el sprint con `sprint.Manager.Get` y lee los ítems de su proyecto; `ProjectStoryPoints` valida el proyecto con `ListItems` y lee los sprints con `ListByProject`.
- `ListByProject` en memoria y en PostgreSQL (`ORDER BY id`, lista vacía no `nil`).
- `NewHandler` crea el `metrics.Manager`; los handlers redondean la velocidad con `math.Round(v*10)/10` y la devuelven `null` si no está disponible. `statusFor` ya traducía `ErrSprintNotFound` y `ErrProjectNotFound` a 404.
- La sección "Métricas" de la interfaz, con el botón "Ver métricas".

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.015s
ok  	seguimiento-medicion-gpc/internal/backlog	0.340s
ok  	seguimiento-medicion-gpc/internal/effort	0.105s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.111s
ok  	seguimiento-medicion-gpc/internal/metrics	0.043s
ok  	seguimiento-medicion-gpc/internal/poker	0.027s
ok  	seguimiento-medicion-gpc/internal/postgres	0.011s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.010s
ok  	seguimiento-medicion-gpc/internal/project	0.082s
ok  	seguimiento-medicion-gpc/internal/sprint	0.132s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

godog: `16 scenarios (16 passed)` y `72 steps (72 passed)`.

## Refactor

No se hizo: el código nuevo es corto y sigue el patrón de los paquetes existentes (`pathID`, `writeError` y `writeJSON` en el handler, mutex del `Manager` en las lecturas de `sprint`).

## Desviaciones del ciclo TDD

Ninguna. Los tests del RED no se modificaron en el GREEN.

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
ok  	seguimiento-medicion-gpc/cmd/server	0.012s	coverage: 33.3% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	0.381s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/effort	0.091s	coverage: 76.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	0.134s	coverage: 98.8% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	0.054s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	0.030s	coverage: 54.4% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	0.012s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.013s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	0.137s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	0.211s	coverage: 76.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
$ go test -count=1 -coverprofile=cover.out ./... && go tool cover -func=cover.out | tail -n 1
total:									(statements)			80.1%
```

- `internal/metrics` queda en 100.0%. `internal/sprint` pasa de 79.1% a 76.3% sin base porque `ListByProject` de PostgreSQL no se ejecuta sin `TEST_DATABASE_URL`; con base, esa función queda en 77.8% (las dos ramas de error de la base, como el `ListByProject` del backlog). `internal/httpapi` queda en 98.8%.
- El total sin base pasa de **79.1%** en `main` (`86c0c5f`, medido con el mismo Go en una copia aparte) a **80.1%**.

### Con base (PostgreSQL 16)

```
$ TEST_DATABASE_URL="postgres://gpc:***@127.0.0.1:5432/gpc_test?sslmode=disable" go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.223s
ok  	seguimiento-medicion-gpc/internal/backlog	1.244s
ok  	seguimiento-medicion-gpc/internal/effort	1.228s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.113s
ok  	seguimiento-medicion-gpc/internal/metrics	0.064s
ok  	seguimiento-medicion-gpc/internal/poker	0.421s
ok  	seguimiento-medicion-gpc/internal/postgres	0.007s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.016s
ok  	seguimiento-medicion-gpc/internal/project	0.395s
ok  	seguimiento-medicion-gpc/internal/sprint	0.692s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

Ningún test quedó en SKIP (`go test -v ./... | grep -c -- '--- SKIP'` → 0). Con `-coverpkg=./...`, el total con base es **94.6%**.

### go test -race (con base)

```
$ TEST_DATABASE_URL="postgres://gpc:***@127.0.0.1:5432/gpc_test?sslmode=disable" go test -count=1 -race ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.432s
ok  	seguimiento-medicion-gpc/internal/backlog	3.119s
ok  	seguimiento-medicion-gpc/internal/effort	2.271s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.667s
ok  	seguimiento-medicion-gpc/internal/metrics	1.150s
ok  	seguimiento-medicion-gpc/internal/poker	1.440s
ok  	seguimiento-medicion-gpc/internal/postgres	1.022s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.027s
ok  	seguimiento-medicion-gpc/internal/project	1.673s
ok  	seguimiento-medicion-gpc/internal/sprint	2.234s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

## Límite conocido

Como decidió la SDD, los Story Points no se congelan al cerrar un sprint: si después se reestima un ítem no terminado o se reabre uno terminado, cambian las métricas de ese sprint cerrado y la velocidad.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | 288c0e9 | `docs(HU-32): agregar especificación SDD` |
| RED | 443c3ff | `test(HU-32): agregar tests (RED)` |
| GREEN | 907e10d | `feat(HU-32): implementar story points planificados, completados y velocidad (GREEN)` |
| Refactor | — | No hubo |
