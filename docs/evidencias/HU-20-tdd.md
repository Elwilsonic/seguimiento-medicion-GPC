# Evidencia TDD — HU-20 · Asignar story points a un ítem del backlog

> Issue del repositorio: #38

- **Responsable:** Cardozo Leandro
- **Rama:** `feature/HU-20-asignar-story-points`
- **Fecha:** 23/10/2026
- **Go:** go1.27.1 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` lo verifica el CI del Pull Request; ver la sección de `go test -race`)
- **Especificación SDD:** [`docs/especificaciones/HU-20-asignar-story-points.md`](../especificaciones/HU-20-asignar-story-points.md), aprobada por Peñalbé Hernán (Agile Enabler) el 23/10/2026
- **Escenarios BDD:** [`features/HU-20-asignar-story-points.feature`](../../features/HU-20-asignar-story-points.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**29 funciones de test cubriendo 14 escenarios BDD (7 escenarios y 7 esquemas del escenario, que con sus ejemplos dan 29 ejecuciones y 129 pasos)**, más un subtest nuevo en cada uno de 2 tests existentes:

- 18 tests unitarios en `internal/backlog/story_points_test.go`.
- 1 test que ejecuta los escenarios BDD con godog (`internal/backlog/story_points_features_test.go`, runner `TestFeatures_HU20AsignarStoryPoints`).
- 8 tests del endpoint en `internal/httpapi/items_test.go` y 2 de la interfaz web en `internal/httpapi/web_test.go` (CA-20.8).
- El subtest `asignar story points` en `TestHandler_CuerpoMayorA1MiBDa413` y en `TestHandler_TextoDespuesDelJSONDa400` (`internal/httpapi/server_test.go`), con una entrada nueva en `bodyEndpoints`.

El conteo sale de los comandos:

```
$ grep -n '^func Test' internal/backlog/story_points_test.go internal/backlog/story_points_features_test.go
internal/backlog/story_points_test.go:57:func TestAssignStoryPoints_ItemSinEstimarQuedaConElValorYConservaElResto(t *testing.T) {
internal/backlog/story_points_test.go:78:func TestAssignStoryPoints_ReestimarReemplazaElValorAnterior(t *testing.T) {
internal/backlog/story_points_test.go:93:func TestAssignStoryPoints_MismoValorEsValido(t *testing.T) {
internal/backlog/story_points_test.go:111:func TestAssignStoryPoints_AceptaCadaValorDeLaEscala(t *testing.T) {
internal/backlog/story_points_test.go:129:func TestStoryPointScale_DevuelveLaEscalaFibonacciEnOrden(t *testing.T) {
internal/backlog/story_points_test.go:140:func TestStoryPointScale_DevuelveUnaListaNuevaEnCadaLlamada(t *testing.T) {
internal/backlog/story_points_test.go:164:func TestAssignStoryPoints_RechazaValoresFueraDeLaEscala(t *testing.T) {
internal/backlog/story_points_test.go:184:func TestAssignStoryPoints_ValorFueraDeLaEscalaSeInformaPrimero(t *testing.T) {
internal/backlog/story_points_test.go:197:func TestAssignStoryPoints_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
internal/backlog/story_points_test.go:214:func TestAssignStoryPoints_SinItemsCreadosDevuelveErrItemNotFound(t *testing.T) {
internal/backlog/story_points_test.go:225:func TestAssignStoryPoints_ItemTerminadoDevuelveErrItemFinished(t *testing.T) {
internal/backlog/story_points_test.go:245:func TestAssignStoryPoints_ItemReabiertoVuelveASerEstimable(t *testing.T) {
internal/backlog/story_points_test.go:263:func TestAssignStoryPoints_ItemNoTerminadoAsignadoAUnSprint(t *testing.T) {
internal/backlog/story_points_test.go:282:func TestManagerUpdate_ConservaLosStoryPointsAsignadosConAssignStoryPoints(t *testing.T) {
internal/backlog/story_points_test.go:303:func TestAssignStoryPoints_ElResultadoEsUnaCopia(t *testing.T) {
internal/backlog/story_points_test.go:325:func TestAssignStoryPoints_ErrorDelRepositorioAlBuscarSeDevuelve(t *testing.T) {
internal/backlog/story_points_test.go:339:func TestAssignStoryPoints_ErrorDelRepositorioAlActualizarSeDevuelveYNoCambiaElItem(t *testing.T) {
internal/backlog/story_points_test.go:366:func TestAssignStoryPoints_SimultaneasConUpdateYChangeStatus(t *testing.T) {
internal/backlog/story_points_features_test.go:19:func TestFeatures_HU20AsignarStoryPoints(t *testing.T) {
```

```
$ git diff c7d7d26 a15b31f -- internal/httpapi/ | grep -E '^\+func Test'
+func TestAsignarStoryPoints_Devuelve200ConLosStoryPointsAsignados(t *testing.T) {
+func TestAsignarStoryPoints_TextoNumericoSeAceptaComoNumero(t *testing.T) {
+func TestAsignarStoryPoints_ValorFueraDeLaEscalaDevuelve400(t *testing.T) {
+func TestAsignarStoryPoints_JSONInvalidoDevuelve400(t *testing.T) {
+func TestAsignarStoryPoints_IDNoNumericoDevuelve400(t *testing.T) {
+func TestAsignarStoryPoints_InexistenteDevuelve404(t *testing.T) {
+func TestAsignarStoryPoints_ItemTerminadoDevuelve409(t *testing.T) {
+func TestAsignarStoryPoints_ErrorDelRepositorioDevuelve500SinDetalle(t *testing.T) {
+func TestInterfaz_TablaDelBacklogMuestraYAsignaStoryPoints(t *testing.T) {
+func TestInterfaz_SelectDeStoryPointsUsaLaEscalaDelBacklog(t *testing.T) {
```

La correspondencia no es 1 a 1: CA-20.8 (API e interfaz web) no tiene escenarios godog y se cubre con los tests de `internal/httpapi`; varios escenarios tienen más de un test unitario y algunos tests cubren varios ejemplos de un esquema en un solo bucle.

| Criterio | Escenarios BDD (ejecuciones) | Tests | Código |
|---|---|---|---|
| CA-20.1 Asignar a un ítem existente, sin cambiar el resto; reestimar | Asignar story points a un ítem sin estimar (1); Reestimar un ítem que ya tiene story points (2) | `TestAssignStoryPoints_ItemSinEstimarQuedaConElValorYConservaElResto`, `TestAssignStoryPoints_ReestimarReemplazaElValorAnterior`, `TestAssignStoryPoints_MismoValorEsValido` | `story_points.go` (`AssignStoryPoints`) |
| CA-20.2 Escala 1, 2, 3, 5, 8, 13, 21 definida en un solo lugar | Aceptar cada valor de la escala Fibonacci (7); La escala compartida no cambia si se modifica la lista devuelta (1) | `TestAssignStoryPoints_AceptaCadaValorDeLaEscala`, `TestStoryPointScale_DevuelveLaEscalaFibonacciEnOrden`, `TestStoryPointScale_DevuelveUnaListaNuevaEnCadaLlamada` | `story_points.go` (`storyPointScale`, `StoryPointScale`) |
| CA-20.3 Fuera de la escala da `ErrInvalidStoryPoints` y no cambia nada | Rechazar valores fuera de la escala (4); Un valor fuera de la escala sobre un ítem inexistente informa la escala (1) | `TestAssignStoryPoints_RechazaValoresFueraDeLaEscala`, `TestAssignStoryPoints_ValorFueraDeLaEscalaSeInformaPrimero` | `story_points.go` (`ErrInvalidStoryPoints`) |
| CA-20.4 Ítem inexistente da `ErrItemNotFound` | Asignar story points a un ítem que no existe (4) | `TestAssignStoryPoints_ItemInexistenteDevuelveErrItemNotFound`, `TestAssignStoryPoints_SinItemsCreadosDevuelveErrItemNotFound` | `story_points.go` y `GetByID` de HU-05 |
| CA-20.5 Ítem `TERMINADO` da `ErrItemFinished`; reabierto vuelve a ser estimable | Rechazar la asignación a un ítem terminado (1); Reestimar un ítem terminado después de reabrirlo (1); Estimar un ítem no terminado asignado a un sprint (2) | `TestAssignStoryPoints_ItemTerminadoDevuelveErrItemFinished`, `TestAssignStoryPoints_ItemReabiertoVuelveASerEstimable`, `TestAssignStoryPoints_ItemNoTerminadoAsignadoAUnSprint` | `story_points.go` (`ErrItemFinished`) |
| CA-20.6 El `Update` de HU-07 conserva los Story Points | Modificar los datos de un ítem estimado conserva sus story points (1) | `TestManagerUpdate_ConservaLosStoryPointsAsignadosConAssignStoryPoints` | `update.go` (sin cambios: ya copiaba `current.StoryPoints`) |
| CA-20.7 Copia por valor, mutex, concurrencia y falla del repositorio | El ítem devuelto es una copia (1); Asignaciones simultáneas con un cambio de estado a TERMINADO (1); El repositorio falla al buscar o al actualizar (2) | `TestAssignStoryPoints_ElResultadoEsUnaCopia`, `TestAssignStoryPoints_SimultaneasConUpdateYChangeStatus`, `TestAssignStoryPoints_ErrorDelRepositorioAlBuscarSeDevuelve`, `TestAssignStoryPoints_ErrorDelRepositorioAlActualizarSeDevuelveYNoCambiaElItem` | `story_points.go` (mutex del `Manager`) |
| CA-20.8 API e interfaz web | — (tests de `internal/httpapi`) | 8 tests `TestAsignarStoryPoints_*`, los subtests `asignar story points` de 413 y de texto después del JSON, `TestInterfaz_TablaDelBacklogMuestraYAsignaStoryPoints`, `TestInterfaz_SelectDeStoryPointsUsaLaEscalaDelBacklog`; recorrido en el navegador | `items.go` (`assignStoryPoints`), `server.go` (ruta y `statusFor`), `index.html` y `app.js` |

Las ejecuciones por criterio suman 29 (3 + 8 + 5 + 4 + 4 + 1 + 4).

## Fase RED

Escenarios BDD, pasos godog y tests escritos antes del código de producción (commit `a15b31f`, después del commit de la SDD `c7d7d26`). Siguiendo la acción 1 de la Retrospectiva del Sprint 2, primero se agregaron las firmas vacías en `internal/backlog/story_points.go`: `StoryPointScale` devuelve `nil`, `AssignStoryPoints` devuelve `Item{}, nil`, y los dos errores ya tienen su mensaje. No hizo falta un stub del handler ni de la ruta para compilar: los tests HTTP llaman a `NewHandler` y, sin la ruta, el `PATCH` responde 405. **El RED es el fallo de las aserciones, no de la compilación.**

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)
```

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	2.193s
FAIL	seguimiento-medicion-gpc/internal/backlog	2.450s
FAIL	seguimiento-medicion-gpc/internal/httpapi	2.369s
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/sprint	2.397s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

Fragmentos de las aserciones que fallan:

```
    story_points_test.go:67: story points devueltos = 0, se esperaban 5
    story_points_test.go:88: story points guardados = 3, se esperaban 8
    story_points_test.go:123: story points guardados = 0, se esperaban 1
    story_points_test.go:423: story points guardados = 0, no es ninguno de los asignados con éxito [1 2 3 5 8 13 21 …]
    items_test.go:183: código = 405, se esperaba 200; cuerpo: Method Not Allowed
        items_test.go:230: código = 405, se esperaba 400; cuerpo: Method Not Allowed
```

Conteos medidos sobre esa salida:

| Medición | Comando | Resultado |
|---|---|---|
| Escenarios BDD de HU-20 | resumen de godog en la salida de `go test ./...` | `29 scenarios (1 passed, 28 failed)` y `129 steps (79 passed, 28 failed, 22 skipped)` |
| Tests que fallan | `grep -E "^--- FAIL: Test" \| wc -l` | **31**: 18 unitarios y el runner godog en `internal/backlog`; 8 del endpoint, 2 de la interfaz y 2 tests existentes con su subtest nuevo en `internal/httpapi` |
| Tests que fallan fuera de HU-20 | el mismo `grep` excluyendo los nombres de HU-20 | ninguno |
| Las demás suites godog | resumen de godog | todas en verde |

La ejecución que pasa en el RED es el ejemplo "mismo valor" (5 → 5) de "Reestimar un ítem que ya tiene story points": el stub no modifica nada y el ítem ya tenía 5. CA-20.1 tiene otras 2 ejecuciones que fallan. **Cada CA tiene al menos un fallo**: CA-20.1 (2 ejecuciones y 3 tests), CA-20.2 (8 y 3), CA-20.3 (5 y 2), CA-20.4 (4 y 2), CA-20.5 (4 y 3), CA-20.6 (1 y 1), CA-20.7 (4 y 4) y CA-20.8 (12 tests de `internal/httpapi`).

**Determinismo del test de concurrencia (CA-20.7).** La aserción no depende del orden en que corren las goroutines: después de `TERMINADO` no se escribe nada, `Update` conserva los Story Points y la escala no tiene el 0, así que los Story Points guardados tienen que ser los de una asignación exitosa, o 0 si ninguna lo fue. Con el stub, todas las asignaciones "terminan bien" y quedan 0, así que el test falla siempre:

```
$ go test ./internal/backlog/ -run 'TestAssignStoryPoints_SimultaneasConUpdateYChangeStatus' -count=30 | grep -c '^--- FAIL: TestAssignStoryPoints_SimultaneasConUpdateYChangeStatus'
30

$ go test ./internal/backlog/ -run 'TestFeatures_HU20AsignarStoryPoints' -count=10 | grep '^29 scenarios' | sort | uniq -c
     10 29 scenarios (1 passed, 28 failed)
```

## Fase GREEN

Código mínimo de producción (commit `69fd122`), sin tocar tests:

- `internal/backlog/story_points.go`: `storyPointScale` (arreglo no exportado, única definición de la escala), `StoryPointScale()` (devuelve `slices.Clone`) y `AssignStoryPoints`, que valida la escala sin consultar el repositorio y, con el mutex del `Manager` tomado, busca el ítem, rechaza los `TERMINADO` y escribe con `Update`.
- `internal/httpapi/items.go`: `storyPointsRequest` con `json.Number` y el handler `assignStoryPoints`, que convierte con `strconv.Atoi` y responde con `ErrInvalidStoryPoints` lo que no es un entero.
- `internal/httpapi/server.go`: la ruta `PATCH /api/items/{id}/story-points`; `ErrInvalidStoryPoints` responde 400 y `ErrItemFinished` responde 409.
- `internal/web/index.html` y `internal/web/app.js`: columna "SP" ("—" si vale 0), la constante `escalaStoryPoints`, y en cada fila un `select` con la escala y el botón "Asignar SP".

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.839s
ok  	seguimiento-medicion-gpc/internal/backlog	0.494s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.451s
ok  	seguimiento-medicion-gpc/internal/postgres	0.668s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.374s
ok  	seguimiento-medicion-gpc/internal/project	0.697s
ok  	seguimiento-medicion-gpc/internal/sprint	0.435s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

```
$ go test -count=1 -v -run 'TestFeatures_HU20AsignarStoryPoints' ./internal/backlog/
29 scenarios (29 passed)
129 steps (129 passed)

$ go test -count=1 -v ./internal/backlog/ ./internal/httpapi/ | grep -E "^--- (PASS|FAIL): Test(AssignStoryPoints|StoryPointScale|ManagerUpdate_ConservaLosStoryPointsAsignados|AsignarStoryPoints|Interfaz_.*StoryPoints|Features_HU20|Handler_CuerpoMayor|Handler_TextoDespues)" | awk '{print $2}' | sort | uniq -c
     31 PASS:

$ go test -count=30 -run 'TestAssignStoryPoints_SimultaneasConUpdateYChangeStatus|TestFeatures_HU20' ./internal/backlog/
ok  	seguimiento-medicion-gpc/internal/backlog	1.315s
```

La primera corrida de `go test ./...` después del GREEN terminó con exit 1 sin ningún test fallido: `go: unlinkat …\server.test.exe: The process cannot access the file because it is being used by another process` (Windows no pudo borrar un ejecutable temporal de `cmd/server`). Las corridas siguientes dieron exit 0.

Antes del commit GREEN se revisó `go tool cover -func`: `StoryPointScale`, `AssignStoryPoints`, `assignStoryPoints` y `statusFor` quedan al 100.0%, así que no hay código nuevo sin un test que lo pida.

## Refactor

No se hizo. El código nuevo sigue los patrones de HU-07 y HU-08 (validar sin repositorio, después leer-modificar-escribir con el mutex) y del handler de cambio de estado; no hay lógica duplicada ni ramas sin test. Los tests repiten un helper de semilla parecido al de HU-08 (`seedPointsItem` y `seedStatusItem`, que difieren en los Story Points); unificarlos tocaría los tests de HU-08 sin mejorar la legibilidad, así que se dejaron.

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)
```

```
$ go vet ./...
(sin salida: sin problemas)
```

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.137s
ok  	seguimiento-medicion-gpc/internal/backlog	0.399s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.303s
ok  	seguimiento-medicion-gpc/internal/postgres	0.109s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.114s
ok  	seguimiento-medicion-gpc/internal/project	0.250s
ok  	seguimiento-medicion-gpc/internal/sprint	0.368s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

### Cobertura

```
$ go test -count=1 -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.783s	coverage: 37.8% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	15.836s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	2.053s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	0.534s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.565s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	0.692s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	1.972s	coverage: 77.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

```
$ go test -count=1 -coverprofile=cover.out ./...
$ go tool cover -func=cover.out | tail -n 1
total:									(statements)		80.6%
$ go tool cover -func=cover.out | grep -E "story_points.go|assignStoryPoints|statusFor"
seguimiento-medicion-gpc/internal/backlog/story_points.go:25:		StoryPointScale		100.0%
seguimiento-medicion-gpc/internal/backlog/story_points.go:36:		AssignStoryPoints	100.0%
seguimiento-medicion-gpc/internal/httpapi/items.go:149:			assignStoryPoints	100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:106:		statusFor		100.0%
```

Comparación con `main` (`e04c875`), medida con los mismos comandos en una copia temporal (`git worktree`) de `main`:

```
$ go test -count=1 -cover ./internal/backlog/ ./internal/httpapi/      (sobre e04c875)
ok  	seguimiento-medicion-gpc/internal/backlog	1.609s	coverage: 82.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.052s	coverage: 100.0% of statements
$ go tool cover -func=cover-main.out | tail -n 1
total:									(statements)		79.7%
```

`internal/backlog` sube de 82.0% a 83.3% e `internal/httpapi` sigue en 100.0%. El total del perfil sin base sube de 79.7% a **80.6%**. Las mediciones son locales (Windows, sin `TEST_DATABASE_URL`, sin `-race`); los valores del CI del Pull Request (Linux, con `-race`) son la fuente autoritativa. HU-20 no cambia la interfaz `Repository` ni el esquema de PostgreSQL (la columna `story_points` ya existía y `Update` ya la escribía), así que no se repitió la medición con base.

### go test -race (CA-20.7)

En este entorno (Windows, `CGO_ENABLED=0`, sin compilador C) `-race` no está disponible. Lo verifica el job "gofmt y tests" del CI del Pull Request, que corre `go test -race -coverprofile=coverage.out ./...`. Sin `-race`, los tests de concurrencia pasan 30 de 30 veces (ver la fase GREEN).

### Recorrido en el navegador (CA-20.8)

Servidor en memoria (`go run ./cmd/server`, puerto 8080) y navegador integrado de la app de escritorio de Claude, con una ventana de 1280×900. Se creó el proyecto "Sistema de ventas" y tres ítems desde la interfaz.

| Captura | Qué muestra |
|---|---|
| [01-backlog-sin-estimar](img/HU-20/01-backlog-sin-estimar.jpg) | Tres ítems recién creados: la columna SP muestra "—" (sin estimar, TEC-03) y cada fila tiene el `select` de la escala y el botón "Asignar SP" |
| [02-items-estimados](img/HU-20/02-items-estimados.jpg) | CA-20.1: ítem 1 con 5 y ítem 2 con 8 Story Points; mensaje "Ítem 2 con 8 story points." |
| [03-reestimar](img/HU-20/03-reestimar.jpg) | CA-20.1: el ítem 1 se reestima de 5 a 13 |
| [04-error-item-terminado](img/HU-20/04-error-item-terminado.jpg) | CA-20.5: el ítem 3, estimado en 3 y pasado a `TERMINADO` (HU-08), no acepta 8; la API responde 409 y la interfaz muestra "no se pueden cambiar los story points de un ítem terminado" |
| [05-editar-conserva-sp](img/HU-20/05-editar-conserva-sp.jpg) | CA-20.6: el ítem 1 se edita con HU-07 ("Login con OAuth") y conserva sus 13 Story Points |

Pedidos a la API registrados por el navegador durante el recorrido:

```
PATCH http://localhost:8080/api/items/1/story-points → 200 OK
PATCH http://localhost:8080/api/items/2/story-points → 200 OK
PATCH http://localhost:8080/api/items/1/story-points → 200 OK
PATCH http://localhost:8080/api/items/3/story-points → 200 OK
PATCH http://localhost:8080/api/items/3/story-points → 409 Conflict
```

Consulta posterior del backlog (`GET /api/proyectos/1/backlog`): ítem 1 `PENDIENTE` con 13, ítem 2 `PENDIENTE` con 8 e ítem 3 `TERMINADO` con 3 Story Points.

Observaciones del recorrido:

- Los valores fuera de la escala (0, negativos, 4, 22, decimales) no se pueden elegir en la interfaz, porque el `select` solo ofrece la escala; ese caso lo cubren los tests de la API (`TestAsignarStoryPoints_ValorFueraDeLaEscalaDevuelve400`).
- En un ítem sin estimar, el `select` muestra 1 como primera opción, mientras la columna SP muestra "—".
- Después del error del ítem terminado, el `select` de esa fila queda en el valor elegido (8) hasta que se recarga el backlog; la columna SP muestra el valor guardado (3).
- El primer intento de levantar el servidor falló porque otro proceso tenía ocupado el puerto 8080; se volvió a levantar y el recorrido se hizo completo desde cero. Los 3 `ERR_CONNECTION_REFUSED` de la consola del navegador son de ese primer intento; el único otro error de consola es el 409 esperado.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | c7d7d26 | `docs(HU-20): agregar especificación SDD` |
| RED | a15b31f | `test(HU-20): agregar tests (RED)` |
| GREEN | 69fd122 | `feat(HU-20): implementar la asignación de story points (GREEN)` |

No hubo commit de refactor porque no se hizo refactor.

HU-20: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
