# Evidencia TDD · HU-14 Consultar sprints anteriores

> Issue del repositorio: #28

- **Especificación SDD:** [`docs/especificaciones/HU-14-sprints-anteriores.md`](../especificaciones/HU-14-sprints-anteriores.md). La aprobó Peñalbé Hernán el 29/10/2026. Aceptó D1 (la lista incluye el sprint abierto) y pidió un ajuste, ya aplicado: el mensaje del botón de cada fila avisa que ese sprint queda como sprint actual.
- **Escenarios BDD:** [`features/HU-14-sprints-anteriores.feature`](../../features/HU-14-sprints-anteriores.feature). Tiene 8 definiciones (7 escenarios y 1 esquema de 3 ejemplos), que godog ejecuta como 10 escenarios con 56 pasos.
- **Rama:** `feature/HU-14-sprints-anteriores`, creada desde `main` `d887ca9`.
- **Story Points:** 3, estimados en el Planning del Sprint 3.

## Trazabilidad

**17 tests nuevos (funciones `Test`) cubren los 10 escenarios BDD ejecutados (8 definiciones)**, y hay además una fila nueva en `TestRutasConID_IDNoNumericoDevuelve400`. Los tests se reparten así:

- 7 unitarios en `internal/sprint/history_test.go`;
- el runner de godog en `internal/sprint/history_features_test.go`;
- 5 de API en `internal/httpapi/sprints_test.go`;
- 4 de interfaz en `internal/httpapi/web_test.go`.

La correspondencia no es 1 a 1. Los escenarios cubren la consulta de dominio. La parte HTTP de CA-14.2 y CA-14.4, y la interfaz de CA-14.6, se cubren solo con tests de Go. Además, varios tests comparten criterio.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-14.1 Sprints abiertos y cerrados, ordenados por ID, con sus datos | "Listar los sprints cerrados y el abierto de un proyecto", "Solo se listan los sprints del proyecto consultado" | `TestHistory_DevuelveLosSprintsAbiertosYCerradosOrdenadosPorID`, `TestHistory_SoloDevuelveLosSprintsDelProyecto`, `TestListarSprintsDelProyecto_Devuelve200ConLosSprintsOrdenados`, `TestInterfaz_VerSprintsLlamaALaAPIConElProyectoActual` |
| CA-14.2 Proyecto sin sprints: lista vacía | "Un proyecto sin sprints devuelve una lista vacía" | `TestHistory_ProyectoSinSprintsDevuelveListaVacia`, `TestListarSprintsDelProyecto_SinSprintsDevuelveListaVacia` |
| CA-14.3 Sprint cerrado con sus ítems completados | "Consultar un sprint cerrado de la lista con sus ítems completados", "Un sprint cerrado sin ítems completados aparece en la lista" | `TestInterfaz_FilaDeSprintSeleccionaElSprintYMuestraSusCompletados`, `TestInterfaz_CompletadosCompartidosPorLosDosBotones` (y, sin cambios, `TestListCompletedItems_SprintCerradoSePuedeConsultar` de HU-12) |
| CA-14.4 Proyecto inexistente (404) e ID no numérico (400) | "Un proyecto inexistente no tiene sprints que listar" (3 ejemplos: 99, 0 y -1) | `TestHistory_ProyectoInexistenteDevuelveErrProjectNotFound`, `TestListarSprintsDelProyecto_ProyectoInexistenteDevuelve404`, `TestListarSprintsDelProyecto_IDNoNumericoDevuelve400`, la fila "listar sprints del proyecto" de `TestRutasConID_IDNoNumericoDevuelve400` |
| CA-14.5 La lista es una copia | "Modificar la lista devuelta no cambia los sprints guardados" | `TestHistory_ModificarLaListaNoCambiaLosSprintsGuardados` |
| CA-14.6 Interfaz: tabla de sprints y botón de cada fila con el mensaje de selección | — | `TestInterfaz_SeccionSprintsTieneElBotonVerSprintsDelProyecto`, `TestInterfaz_VerSprintsLlamaALaAPIConElProyectoActual`, `TestInterfaz_FilaDeSprintSeleccionaElSprintYMuestraSusCompletados`, `TestInterfaz_CompletadosCompartidosPorLosDosBotones` |
| RN-14.7 Una falla de lectura corta sin lista parcial | "Una falla del repositorio de sprints se informa sin lista parcial" | `TestHistory_ErrorAlLeerElProyectoSeDevuelve`, `TestHistory_ErrorAlLeerLosSprintsSeDevuelve`, `TestListarSprintsDelProyecto_FallaDeUnRepositorioDevuelve500` |

## Fase RED (`d93b45d`)

Se escribieron los tests y el `.feature` antes del código. `go test ./...` sobre `d93b45d`:

```
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/dashboard	(cached)
ok  	seguimiento-medicion-gpc/internal/defect	(cached)
ok  	seguimiento-medicion-gpc/internal/effort	(cached)
# seguimiento-medicion-gpc/internal/sprint [seguimiento-medicion-gpc/internal/sprint.test]
internal\sprint\history_features_test.go:161:38: s.manager.History undefined (type *Manager has no field or method History)
internal\sprint\history_test.go:58:22: manager.History undefined (type *Manager has no field or method History)
internal\sprint\history_test.go:84:22: manager.History undefined (type *Manager has no field or method History)
internal\sprint\history_test.go:99:22: manager.History undefined (type *Manager has no field or method History)
internal\sprint\history_test.go:118:23: manager.History undefined (type *Manager has no field or method History)
internal\sprint\history_test.go:141:22: manager.History undefined (type *Manager has no field or method History)
internal\sprint\history_test.go:151:24: manager.History undefined (type *Manager has no field or method History)
internal\sprint\history_test.go:168:22: manager.History undefined (type *Manager has no field or method History)
internal\sprint\history_test.go:190:22: manager.History undefined (type *Manager has no field or method History)
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.02s)
    --- FAIL: TestRutasConID_IDNoNumericoDevuelve400/listar_sprints_del_proyecto (0.00s)
        server_test.go:251: GET /api/proyectos/abc/sprints: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
        server_test.go:253: código = 404, se esperaba 400; cuerpo: 404 page not found
--- FAIL: TestListarSprintsDelProyecto_Devuelve200ConLosSprintsOrdenados (0.00s)
    sprints_test.go:334: GET /api/proyectos/1/sprints: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    sprints_test.go:336: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestListarSprintsDelProyecto_SinSprintsDevuelveListaVacia (0.00s)
    sprints_test.go:347: GET /api/proyectos/1/sprints: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    sprints_test.go:349: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestListarSprintsDelProyecto_IDNoNumericoDevuelve400 (0.00s)
    sprints_test.go:354: GET /api/proyectos/abc/sprints: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    sprints_test.go:356: código = 404, se esperaba 400; cuerpo: 404 page not found
--- FAIL: TestListarSprintsDelProyecto_ProyectoInexistenteDevuelve404 (0.00s)
    sprints_test.go:363: GET /api/proyectos/99/sprints: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    sprints_test.go:365: cuerpo no es JSON válido (invalid character 'p' after top-level value): 404 page not found
--- FAIL: TestListarSprintsDelProyecto_FallaDeUnRepositorioDevuelve500 (0.00s)
    sprints_test.go:387: GET /api/proyectos/1/sprints: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    sprints_test.go:389: código = 404, se esperaba 500; cuerpo: 404 page not found
--- FAIL: TestInterfaz_SeccionSprintsTieneElBotonVerSprintsDelProyecto (0.00s)
    web_test.go:1234: la sección de sprints no contiene "<button type=\"button\" id=\"ver-sprints\" class=\"secundario\">Ver sprints del proyecto</button>"
    web_test.go:1234: la sección de sprints no contiene "<tbody id=\"lista-sprints\"></tbody>"
    web_test.go:1234: la sección de sprints no contiene "<th>ID</th><th>Nombre</th><th>Inicio</th><th>Fin</th><th>Sprint goal</th><th>Estado</th><th>Acciones</th>"
--- FAIL: TestInterfaz_VerSprintsLlamaALaAPIConElProyectoActual (0.00s)
    web_test.go:1243: app.js no tiene el handler de "ver-sprints"
--- FAIL: TestInterfaz_FilaDeSprintSeleccionaElSprintYMuestraSusCompletados (0.00s)
    web_test.go:1269: app.js no tiene el handler de "ver-sprints"
--- FAIL: TestInterfaz_CompletadosCompartidosPorLosDosBotones (0.00s)
    web_test.go:1290: app.js no tiene la función mostrarCompletados
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi	8.599s
ok  	seguimiento-medicion-gpc/internal/metrics	(cached)
ok  	seguimiento-medicion-gpc/internal/poker	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/report	(cached)
FAIL	seguimiento-medicion-gpc/internal/sprint [build failed]
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

Todos los fallos corresponden a la funcionalidad nueva:

- **`internal/sprint`:** no compila porque `Manager.History` no existe. Las 9 referencias están en los dos archivos nuevos de HU-14.
- **`internal/httpapi`:** fallan los 10 tests nuevos (5 de API, 4 de interfaz y la fila de `TestRutasConID_IDNoNumericoDevuelve400`). La ruta `GET /api/proyectos/{id}/sprints` no existe y responde "404 page not found". La interfaz no tiene el botón `ver-sprints`, la tabla `lista-sprints` ni la función `mostrarCompletados`.
- **Los otros 11 paquetes:** pasan. Ningún test anterior falló.

## Fase GREEN (`2e353de`)

Código agregado:

- **`internal/sprint/history.go`:** `History(projectID)` valida el proyecto con `project.Manager.Get` antes de tomar el mutex y después reutiliza `ListByProject`. Ante cualquier error devuelve `nil`.
- **`internal/httpapi/sprints.go`:** el handler `listProjectSprints`, que arma la respuesta con `toSprintResponse`. Si no hay sprints devuelve `[]`, porque la lista se crea con `make`.
- **`internal/httpapi/server.go`:** la ruta `GET /api/proyectos/{id}/sprints`.
- **`internal/web/index.html`:** el botón "Ver sprints del proyecto" y la tabla `lista-sprints`.
- **`internal/web/app.js`:**
  - La consulta de los ítems completados pasa a la función `mostrarCompletados`. La usan el botón de HU-12, que conserva su mensaje, y el botón de cada fila.
  - El handler de `ver-sprints` llena la tabla e informa "N sprint(s) en el proyecto X.".
  - El botón de cada fila llama a `mostrarSprint(s)` y muestra "Sprint N seleccionado: M ítem(s) completado(s).".
  - `boton` suma un parámetro opcional `msgId`, para que los errores de la fila se muestren en `msg-sprints`. Su valor por defecto sigue siendo `msg-backlog`.

`go test ./...` sobre `2e353de`:

```
ok  	seguimiento-medicion-gpc/cmd/server	9.085s
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/dashboard	10.749s
ok  	seguimiento-medicion-gpc/internal/defect	12.325s
ok  	seguimiento-medicion-gpc/internal/effort	7.479s
ok  	seguimiento-medicion-gpc/internal/httpapi	7.580s
ok  	seguimiento-medicion-gpc/internal/metrics	6.483s
ok  	seguimiento-medicion-gpc/internal/poker	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/report	4.716s
ok  	seguimiento-medicion-gpc/internal/sprint	5.879s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

Los escenarios de HU-14 (`go test ./internal/sprint/ -run 'TestFeatures_HU14|TestHistory' -v -count=1`):

```
10 scenarios (10 passed)
56 steps (56 passed)
--- PASS: TestFeatures_HU14SprintsAnteriores (0.04s)
--- PASS: TestHistory_DevuelveLosSprintsAbiertosYCerradosOrdenadosPorID (0.00s)
--- PASS: TestHistory_SoloDevuelveLosSprintsDelProyecto (0.00s)
--- PASS: TestHistory_ProyectoSinSprintsDevuelveListaVacia (0.00s)
--- PASS: TestHistory_ProyectoInexistenteDevuelveErrProjectNotFound (0.00s)
--- PASS: TestHistory_ModificarLaListaNoCambiaLosSprintsGuardados (0.00s)
--- PASS: TestHistory_ErrorAlLeerElProyectoSeDevuelve (0.00s)
--- PASS: TestHistory_ErrorAlLeerLosSprintsSeDevuelve (0.00s)
ok  	seguimiento-medicion-gpc/internal/sprint	6.616s
```

Antes del commit también se corrieron `go vet ./...` y `node --check internal/web/app.js`, y ninguno informó problemas. `go tool cover -func` mostró `History` y `listProjectSprints` en 100.0%.

### Prueba de punta a punta de la API

Se hizo con el servidor compilado de la rama (`PORT=18914`, datos en memoria) y `curl` desde Git Bash. En las respuestas, "Métricas" se ve como `M�tricas` porque `curl` en Windows manda el argumento en la codificación del sistema, no en UTF-8. `TestListarSprintsDelProyecto_Devuelve200ConLosSprintsOrdenados` verifica ese mismo texto en UTF-8.

| Pedido | Resultado |
|---|---|
| `GET /api/proyectos/1/sprints`, con el proyecto creado y sin sprints | 200, `[]` |
| Sprint 1 con los ítems 1 y 2; el ítem 1 se completa y el sprint se cierra; después se crea el Sprint 2 | Todos los pedidos responden 200 o 201 |
| `GET /api/proyectos/1/sprints` | 200, `[{"id":1,…,"nombre":"Sprint 1",…,"estado":"CERRADO"},{"id":2,…,"nombre":"Sprint 2",…,"estado":"ABIERTO"}]` |
| `GET /api/sprints/1/completados` (consulta de HU-12 sobre el sprint cerrado) | 200, solo el ítem 1 "Login" en TERMINADO |
| `GET /api/proyectos/99/sprints` | 404, `{"error":"no existe un proyecto con ese id"}` |
| `GET /api/proyectos/abc/sprints` | 400, `{"error":"el id debe ser un número entero"}` |

### Uso desde el navegador

Pendiente: el recorrido en el navegador de Buttini Cristobal. Consiste en crear un proyecto y dos sprints (el primero cerrado con un ítem completado), tocar "Ver sprints del proyecto" y después "Ver ítems completados" en la fila del sprint cerrado.

## Refactor (`3d53790`)

En `internal/web/index.html`, el comentario de la sección "Proyecto y sprint actuales" decía que no había listados de proyectos ni de sprints. Con HU-14 eso ya no es cierto, así que se corrigió el comentario. No cambia el comportamiento: después del cambio, `go test -count=1 ./internal/httpapi/ ./internal/sprint/` sigue en verde.

## Integración con `main` (`d6f05ea`)

Después del GREEN se integró en la rama `main` `4a6971a`, que trae el merge del PR #153 (TEC-06). Se hizo con un merge commit, sin rebase. El PR #153 solo cambia documentación (`docs/documentacion-tecnica.md` y `docs/ia/registro.md`), así que el merge no trae cambios en archivos Go.

El único archivo con conflicto fue `docs/ia/registro.md`, porque las dos ramas agregaban una fila después de la de HU-24. Se resolvió dejando las dos filas: primero la de TEC-06 y después la de HU-14. Después del merge, `go test -count=1 ./...` sigue en verde en los 13 paquetes con tests. La cobertura no cambia, porque no hay código nuevo.

## Desviaciones del ciclo TDD

No hubo. Los tests fallaron solo por la funcionalidad nueva y el código mínimo los pasó a verde sin corregir ningún test.

## Validación final

Medida sobre `2e353de`, antes del refactor. El refactor cambia solo un comentario de HTML y no altera la cobertura de Go. Se midió en Windows con go1.27.1, sin `TEST_DATABASE_URL`. Se omitieron 37 tests de PostgreSQL, y HU-14 no agrega ninguno. `-race` no se corrió localmente porque no hay cgo; lo corre el CI.

`gofmt -w` sobre los archivos Go modificados (`history.go`, `history_test.go`, `history_features_test.go`, `sprints.go`, `server.go`, `sprints_test.go`, `server_test.go` y `web_test.go`) no cambió nada: `git status --short` siguió vacío. `go vet ./...` no informó problemas.

`go test -count=1 ./...`:

```
ok  	seguimiento-medicion-gpc/cmd/server	6.073s
ok  	seguimiento-medicion-gpc/internal/backlog	7.759s
ok  	seguimiento-medicion-gpc/internal/dashboard	4.389s
ok  	seguimiento-medicion-gpc/internal/defect	4.230s
ok  	seguimiento-medicion-gpc/internal/effort	4.667s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.747s
ok  	seguimiento-medicion-gpc/internal/metrics	3.256s
ok  	seguimiento-medicion-gpc/internal/poker	3.524s
ok  	seguimiento-medicion-gpc/internal/postgres	3.171s
ok  	seguimiento-medicion-gpc/internal/postgrestest	3.184s
ok  	seguimiento-medicion-gpc/internal/project	3.863s
ok  	seguimiento-medicion-gpc/internal/report	2.853s
ok  	seguimiento-medicion-gpc/internal/sprint	3.123s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

`go test -count=1 -cover ./...`:

```
ok  	seguimiento-medicion-gpc/cmd/server	2.939s	coverage: 31.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	5.826s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/dashboard	1.275s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/defect	4.615s	coverage: 77.7% of statements
ok  	seguimiento-medicion-gpc/internal/effort	4.601s	coverage: 75.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	4.552s	coverage: 98.5% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	4.054s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	3.962s	coverage: 75.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.130s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	3.708s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	5.931s	coverage: 71.1% of statements
ok  	seguimiento-medicion-gpc/internal/report	4.907s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	4.696s	coverage: 77.1% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

`go test -count=1 -coverprofile=cover.out ./...` y `go tool cover -func=cover.out`, funciones nuevas y total:

```
seguimiento-medicion-gpc/internal/httpapi/sprints.go:61:		listProjectSprints	100.0%
seguimiento-medicion-gpc/internal/sprint/history.go:7:			History			100.0%
total:									(statements)		85.0%
```

**Comparación con `main` (`d887ca9`):** se midió con el mismo Go y los mismos comandos, en una copia temporal con `git worktree`. `internal/sprint` sube de 76.3% a 77.1% y el total, de 84.9% a 85.0%. Los demás paquetes no cambian, e `internal/httpapi` queda en 98.5% en los dos. El detalle está en [`docs/reportes/cobertura.md`](../reportes/cobertura.md).

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | bf00167 | `docs(HU-14): agregar especificación SDD` |
| RED | d93b45d | `test(HU-14): agregar tests (RED)` |
| GREEN | 2e353de | `feat(HU-14): implementar la consulta de los sprints del proyecto (GREEN)` |
| Integración con `main` | d6f05ea | `chore(HU-14): traer main con TEC-06` |
| Refactor | 3d53790 | `refactor(HU-14): actualizar el comentario del sprint actual en la interfaz` |

HU-14: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
