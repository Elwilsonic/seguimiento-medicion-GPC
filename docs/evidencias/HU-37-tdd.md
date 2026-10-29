# Evidencia TDD · HU-37 Ver las métricas en el dashboard con gráficos

> Issue del repositorio: #55

- **Especificación SDD:** [`docs/especificaciones/HU-37-graficos-dashboard.md`](../especificaciones/HU-37-graficos-dashboard.md). La aprobó Peñalbé Hernán el 29/10/2026, que aceptó la ruta nueva como diferencia con el acta y pidió tres ajustes, ya aplicados: la línea de aprobación, el nombre de la función de conversión de `httpapi` y el orden de recorte de los gráficos.
- **Escenarios BDD:** [`features/HU-37-graficos-dashboard.feature`](../../features/HU-37-graficos-dashboard.feature). Tiene 11 definiciones (9 escenarios y 2 esquemas de 3 ejemplos cada uno), que godog ejecuta como 15 escenarios con 126 pasos.
- **Rama:** `feature/HU-37-graficos-dashboard`.
- **Story Points:** 5, estimados en el Planning del Sprint 4.
- **Orden de recorte:** no hizo falta recortar. Están los tres gráficos.

## Trazabilidad

**25 tests nuevos (funciones `Test`) cubren los 15 escenarios BDD ejecutados (11 definiciones)**, y hay además una fila nueva en `TestRutasConID_IDNoNumericoDevuelve400`. Los tests se reparten así:

- 11 unitarios en `internal/metrics/sprints_test.go`;
- el runner de godog en `internal/metrics/sprints_features_test.go`;
- 6 de API en `internal/httpapi/metrics_test.go`;
- 7 de interfaz en `internal/httpapi/web_test.go`.

La correspondencia no es 1 a 1. Los escenarios cubren la consulta de dominio, mientras que la parte HTTP y la de interfaz de CA-37.1, CA-37.2, CA-37.4 y CA-37.6 se cubren solo con tests de Go. Además, varios tests comparten criterio.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-37.1 M1 a M10 y tres gráficos en el Dashboard | "Cada sprint del proyecto con sus métricas, en orden de ID" | `TestProjectSprints_EjemploDeTEC03ConUnSprintCerradoYOtroAbierto`, `TestMetricasPorSprint_ContratoConUnSprintCerradoYOtroAbierto`, `TestInterfaz_DashboardTieneElBotonVerGraficos`, `TestInterfaz_VerGraficosLlamaALasDosRutas`, `TestInterfaz_TextoDeStoryPointsCompartidoPorLosDosBotones`, `TestInterfaz_GraficosTienenLosTresTitulos`, `TestInterfaz_GraficosUsanVariablesDeColor` |
| CA-37.2 "sin datos" y proyecto sin sprints | "Un proyecto sin sprints devuelve una lista vacía", "Un sprint sin ítems tiene todo en 0 y el porcentaje y la desviación no disponibles", "Un sprint con horas estimadas y sin esfuerzos tiene la desviación no disponible" | `TestProjectSprints_SinSprintsDevuelveListaVaciaNoNil`, `TestProjectSprints_SprintSinItemsTodoEnCeroYNoDisponibles`, `TestProjectSprints_SinEsfuerzosLaDesviacionNoDisponible`, `TestInterfaz_GraficosMuestranSinDatos` |
| CA-37.3 Valores iguales a los de HU-35 | "Los valores de cada sprint coinciden con su consulta individual", "Un ítem sin sprint no cuenta en ningún sprint de la lista" | `TestProjectSprints_EsIgualASprintParaCadaSprint`, `TestProjectSprints_ItemSinSprintNoCuentaEnNingunSprint`, `TestMetricasPorSprint_ValoresIgualesALaConsultaDeCadaSprint` |
| CA-37.4 Ruta: orden, `[]`, 400, 404 y 500 | "Un proyecto con un único sprint, abierto", "Más de 3 sprints cerrados aparecen todos, y el backlog se lee una sola vez", "Un proyecto inexistente devuelve un error y ninguna lista" (3 ejemplos), "Las lecturas se hacen en orden y la primera que falla corta la consulta" (3 ejemplos), "Una falla al leer los defectos corta la consulta sin lista parcial" | `TestProjectSprints_OrdenadosPorIDConAbiertosYCerrados`, `TestProjectSprints_InexistenteDevuelveErrProjectNotFound`, `TestProjectSprints_ErroresEnElOrdenDeLectura`, `TestProjectSprints_LeeLosItemsYLosDefectosUnaSolaVez`, `TestMetricasPorSprint_SinSprintsDevuelveListaVacia`, `TestMetricasPorSprint_IDNoNumericoDevuelve400`, `TestMetricasPorSprint_ProyectoInexistenteDevuelve404`, `TestMetricasPorSprint_FallaDeUnRepositorioDevuelve500`, la fila "métricas por sprint del proyecto" de `TestRutasConID_IDNoNumericoDevuelve400` |
| CA-37.5 No modifica nada y devuelve copias | "Cada sprint del proyecto con sus métricas, en orden de ID" (sus últimos dos pasos), "Un proyecto sin sprints devuelve una lista vacía" | `TestProjectSprints_NoModificaNadaYDevuelveCopias` |
| CA-37.6 SVG propio, sin librerías, y textos como texto | — | `TestInterfaz_GraficosSeDibujanConSVGPropioSinLibrerias` |

## Fase RED (`cb3a941`)

El RED agrega:

- los tests y el `.feature`;
- la fila nueva de `TestRutasConID_IDNoNumericoDevuelve400`;
- en `internal/metrics/sprints.go`, el tipo `SprintWithMetrics` y la firma de `ProjectSprints`, que devuelve `nil` con el error "HU-37: sin implementar". Es lo mínimo para que el paquete compile, como en HU-47.

La ruta, el handler y la interfaz no están en el RED. No cambia ningún test existente, salvo esa fila nueva.

Corrida sobre `cb3a941`, en Windows con go1.27.1, usando `go test -count=1`. La salida está filtrada a los resultados y a las fallas:

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
--- FAIL: TestMetricasPorSprint_ContratoConUnSprintCerradoYOtroAbierto (0.00s)
--- FAIL: TestMetricasPorSprint_ValoresIgualesALaConsultaDeCadaSprint (0.00s)
--- FAIL: TestMetricasPorSprint_SinSprintsDevuelveListaVacia (0.00s)
--- FAIL: TestMetricasPorSprint_IDNoNumericoDevuelve400 (0.00s)
--- FAIL: TestMetricasPorSprint_ProyectoInexistenteDevuelve404 (0.00s)
--- FAIL: TestMetricasPorSprint_FallaDeUnRepositorioDevuelve500 (0.00s)
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.01s)
    --- FAIL: TestRutasConID_IDNoNumericoDevuelve400/métricas_por_sprint_del_proyecto (0.00s)
        server_test.go:238: código = 404, se esperaba 400; cuerpo: 404 page not found
--- FAIL: TestInterfaz_DashboardTieneElBotonVerGraficos (0.00s)
    web_test.go:809: la sección del dashboard no contiene "id=\"ver-graficos\""
--- FAIL: TestInterfaz_VerGraficosLlamaALasDosRutas (0.00s)
    web_test.go:818: app.js no tiene el handler de "ver-graficos"
--- FAIL: TestInterfaz_TextoDeStoryPointsCompartidoPorLosDosBotones (0.00s)
    web_test.go:839: app.js no tiene la función storyPointsYVelocidad
--- FAIL: TestInterfaz_GraficosTienenLosTresTitulos (0.00s)
--- FAIL: TestInterfaz_GraficosMuestranSinDatos (0.00s)
--- FAIL: TestInterfaz_GraficosSeDibujanConSVGPropioSinLibrerias (0.00s)
--- FAIL: TestInterfaz_GraficosUsanVariablesDeColor (0.00s)
FAIL seguimiento-medicion-gpc/internal/httpapi
15 scenarios (15 failed)
126 steps (74 passed, 15 failed, 37 skipped)
--- FAIL: TestFeatures_HU37GraficosDashboard (0.17s)
    sprints_features_test.go:34: fallaron escenarios BDD de HU-37
--- FAIL: TestProjectSprints_EjemploDeTEC03ConUnSprintCerradoYOtroAbierto (0.00s)
    sprints_test.go:46: ProjectSprints devolvió un error inesperado: HU-37: sin implementar
--- FAIL: TestProjectSprints_EsIgualASprintParaCadaSprint (0.00s)
--- FAIL: TestProjectSprints_OrdenadosPorIDConAbiertosYCerrados (0.00s)
--- FAIL: TestProjectSprints_ItemSinSprintNoCuentaEnNingunSprint (0.00s)
--- FAIL: TestProjectSprints_SinSprintsDevuelveListaVaciaNoNil (0.00s)
--- FAIL: TestProjectSprints_SprintSinItemsTodoEnCeroYNoDisponibles (0.00s)
--- FAIL: TestProjectSprints_SinEsfuerzosLaDesviacionNoDisponible (0.00s)
--- FAIL: TestProjectSprints_InexistenteDevuelveErrProjectNotFound (0.00s)
    sprints_test.go:211: ProjectSprints(99): err = HU-37: sin implementar, se esperaba no existe un proyecto con ese id
--- FAIL: TestProjectSprints_ErroresEnElOrdenDeLectura (0.00s)
    sprints_test.go:238: ítems y sprints: err = HU-37: sin implementar, se esperaba falla al leer los ítems
--- FAIL: TestProjectSprints_LeeLosItemsYLosDefectosUnaSolaVez (0.00s)
--- FAIL: TestProjectSprints_NoModificaNadaYDevuelveCopias (0.00s)
FAIL seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/sprint
FAIL
```

Fallan **26 tests de primer nivel**, todos de HU-37:

- en `internal/metrics`, los 11 unitarios y el runner de godog (15 de 15 escenarios);
- en `internal/httpapi`, los 6 de la API, los 7 de la interfaz y `TestRutasConID_IDNoNumericoDevuelve400`, en el que falla solo la fila nueva.

Los escenarios godog de HU-32, HU-33, HU-34, HU-47 y HU-35 siguen en verde (HU-35: 29 de 29) y los demás paquetes pasan. **Ningún test nuevo pasa en el RED.**

Los motivos de las fallas son estos, ninguno de compilación ni un panic:

- los de dominio, por el error del stub o porque el error esperado no coincide;
- los de la API, con 404 "page not found", porque la ruta no existe;
- los de interfaz, porque falta el contenido nuevo en `index.html`, `app.js` y `styles.css`.

## Fase GREEN (`9f5cdeb`)

- **`internal/metrics/sprints.go`:** `ProjectSprints` hace estas lecturas, en este orden (RN-37.2):
  1. los ítems, una vez, con `ListItems`, que valida el proyecto;
  2. la lista de sprints, una vez;
  3. los esfuerzos de cada sprint, en orden de ID;
  4. los defectos, una vez.

  Corta en el primer error y no devuelve una lista parcial. Si no hay sprints, devuelve una lista nueva vacía (`[]`).
- **`internal/metrics/summary.go`:** el cálculo de `Sprint` pasa a `sprintMetricsOf`, una función pura que usan `Sprint` y `ProjectSprints` (RN-37.3). `Sprint` conserva su firma y su resultado, y los tests de HU-35 pasan sin cambios.
- **`internal/httpapi/metrics.go`:**
  - `toSprintMetricsResponse`, extraída del handler `sprintMetrics`. La respuesta de `GET /api/sprints/{id}/metricas` no cambia;
  - `sprintWithMetricsResponse`, con `{sprint, metricas}`;
  - el handler `projectSprintsMetrics`.
- **`internal/httpapi/server.go`:** la ruta `GET /api/proyectos/{id}/metricas/sprints`.
- **`internal/web/index.html`:** en la sección "Dashboard", el botón "Ver gráficos", el texto `datos-graficos`, el contenedor `graficos` y el mensaje `msg-graficos`.
- **`internal/web/app.js`:**
  - `storyPointsYVelocidad`, extraída del handler de "Ver métricas" sin cambiar su texto;
  - el handler de "Ver gráficos";
  - `graficoDeBarras`, `graficosDelProyecto` y `desviacionDelSprint`, que dibujan con `document.createElementNS` y `textContent`.
- **`internal/web/styles.css`:** las variables `--serie-a`, `--serie-b` y `--referencia`, y los estilos de `.grafico` y `.leyenda`.

El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat cb3a941 9f5cdeb -- '*_test.go' '*.feature'
(sin salida)
```

godog sobre `9f5cdeb`: `15 scenarios (15 passed)`, `126 steps (126 passed)`.

`go tool cover -func`, antes del commit GREEN: `ProjectSprints`, `sprintMetricsOf`, `Sprint`, `Project`, `sprintMetrics`, `toSprintMetricsResponse` y `projectSprintsMetrics` quedan al 100.0%. No queda código de Go nuevo sin un test que lo pida.

### Prueba de punta a punta de la API

Con el servidor real (`go build ./cmd/server`, almacenamiento en memoria y `PORT=8137`). El proyecto tiene el Sprint 1 cerrado, con el ítem A (5 SP, 10 h estimadas, 12 h reales, terminado), y el Sprint 2 abierto, con el ítem B (3 SP, sin horas):

| Pedido | Respuesta |
|---|---|
| `GET /api/proyectos/1/metricas/sprints` (sin sprints) | 200 `[]` |
| `GET /api/proyectos/1/metricas/sprints` | 200 `[{"sprint":{"id":1,…,"estado":"CERRADO"},"metricas":{"sprint_id":1,"sp_planificados":5,"sp_completados":5,…,"horas_estimadas":10,"horas_reales":12,"desviacion_horas":2,"desviacion_porcentual":20,…}},{"sprint":{"id":2,…,"estado":"ABIERTO"},"metricas":{"sprint_id":2,"sp_planificados":3,"sp_completados":0,…,"desviacion_horas":null,"desviacion_porcentual":null,…}}]` |
| `GET /api/sprints/1/metricas` | 200, el mismo objeto que `metricas` del primer elemento (CA-37.3) |
| `GET /api/proyectos/1/metricas` | 200 `{…,"velocidad":5,"sprints_cerrados":1,…}`, la línea de referencia del gráfico de velocidad |
| `GET /api/proyectos/abc/metricas/sprints` | 400 `{"error":"el id debe ser un número entero"}` |
| `GET /api/proyectos/99/metricas/sprints` | 404 `{"error":"no existe un proyecto con ese id"}` |
| `GET /` | La sección "Dashboard" tiene `ver-graficos`, `datos-graficos`, `graficos` y `msg-graficos` |

### Dibujo de los gráficos

Los tests de Go no ejecutan JavaScript, así que el dibujo se verificó aparte:

- `node --check internal/web/app.js` no informa errores de sintaxis (Node v24.19.0).
- **Prueba de humo con un DOM simulado.** Es un script fuera del repositorio que ejecuta en Node el bloque de HU-37 de `app.js` (`numero1`, `graficoDeBarras`, `graficosDelProyecto` y `desviacionDelSprint`):
  - Con dos sprints, uno cerrado llamado `Sprint <1>` y otro abierto, y velocidad 8: salen los tres gráficos con sus títulos y 9 barras escaladas al mayor valor. La línea de referencia dice "Velocidad: 8.0 SP", las horas aparecen como "20.0" y la desviación como "Desviación: 2.0 h (25.0 %)". El nombre `Sprint <1>` se muestra como texto, sin `NaN` ni `Infinity` en los atributos, y los `<svg>` tienen el espacio de nombres SVG.
  - Solo con un sprint abierto: "Velocidad: sin datos (no hay sprints cerrados)."
  - Sin sprints: un único aviso, "Sin sprints: no hay datos para los gráficos."
  - Con todos los valores en 0: se dibuja sin `NaN` ni `Infinity` y muestra "Desviación: sin datos".
- **Recorrido en el navegador.** Lo hizo Buttini Cristobal el 29/10/2026, con el servidor de la rama corriendo en `localhost:8080`, desde la sección "Dashboard" y el botón "Ver gráficos". Revisó el texto de M1 a M10, los tres gráficos, la línea de velocidad y los textos "sin datos", y confirmó que todo funciona. Antes se verificó con `curl` que ese servidor servía la versión de HU-37: la página con `ver-graficos`, `app.js` con `graficosDelProyecto` y la ruta nueva respondiendo JSON.

## Refactor

No hubo. `sprintMetricsOf`, `toSprintMetricsResponse` y `storyPointsYVelocidad` ya evitan la duplicación que pedía la SDD y se hicieron en el GREEN.

## Desviaciones del ciclo TDD

- **Ninguna en el RED ni en el GREEN.** Ningún test nuevo pasó con el stub, y el GREEN no tocó tests.
- **Dibujo sin tests automáticos.** Los tests de Go no ejecutan JavaScript, como anticipa la SDD en "Límite conocido". El dibujo se cubre con los tests de contenido, la prueba de humo y el recorrido en el navegador (ver "Dibujo de los gráficos").

## Integración con `main` (`d88c4fd`)

Antes de abrir el Pull Request se trajo `main` (`a30a0f8`, con TEC-06) a la rama con un merge commit, no con rebase, para que los hashes del RED y del GREEN de esta evidencia sigan existiendo (acción 5 de la Retrospectiva del Sprint 3). Solo cambia `docs/documentacion-tecnica.md`, que HU-37 no toca, así que no hubo conflictos. Después del merge, `gofmt -l .` y `go vet ./...` no dan salida, y `go test -count=1 ./...` da `ok` en los 12 paquetes. Como no cambió código, la cobertura medida sobre el GREEN sigue valiendo.

## Validación final

Sobre `9f5cdeb`, en Windows con go1.27.1:

```
$ gofmt -w internal/metrics/sprints.go internal/metrics/summary.go internal/metrics/sprints_test.go internal/metrics/sprints_features_test.go internal/httpapi/metrics.go internal/httpapi/server.go internal/httpapi/metrics_test.go internal/httpapi/web_test.go internal/httpapi/server_test.go
(sin cambios)

$ go vet ./...
(sin salida)

$ go test ./...
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
ok   seguimiento-medicion-gpc/internal/httpapi       coverage: 98.7% of statements
ok   seguimiento-medicion-gpc/internal/metrics       coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/poker         coverage: 69.5% of statements
ok   seguimiento-medicion-gpc/internal/postgres      coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/postgrestest  coverage: 36.8% of statements
ok   seguimiento-medicion-gpc/internal/project       coverage: 67.5% of statements
ok   seguimiento-medicion-gpc/internal/sprint        coverage: 76.3% of statements
     seguimiento-medicion-gpc/internal/web           coverage: 0.0% of statements
```

godog HU-37: `15 scenarios (15 passed)`, `126 steps (126 passed)`.

- **Total del perfil sin base** (`go test -count=1 -coverprofile=cover.out ./...` y `go tool cover -func=cover.out`): **83.0%**.
- **Comparación con `main`** (`33b1234`, con HU-35), medido con el mismo Go y los mismos comandos en una copia temporal con `git worktree`:
  - el total sube de 82.7% a 83.0%;
  - `internal/httpapi` sube de 98.6% a 98.7%;
  - `internal/metrics` sigue en 100.0% y los demás paquetes conservan sus porcentajes.
- **Tests de PostgreSQL:** se omiten 34 (`--- SKIP`), porque `TEST_DATABASE_URL` no está definida. HU-37 no toca la persistencia ni el esquema, así que no se midió con base.
- **`-race`:** `go test -race ./...` no corre localmente (`CGO_ENABLED=0`) y lo verifica el CI del Pull Request.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | 971207e | `docs(HU-37): agregar la especificación (SDD)` |
| RED | cb3a941 | `test(HU-37): agregar tests (RED)` |
| GREEN | 9f5cdeb | `feat(HU-37): mostrar las métricas del dashboard con gráficos (GREEN)` |
| Refactor | — | No hubo |
| Integración con `main` | d88c4fd | `chore(HU-37): traer main con TEC-06` |

HU-37: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
