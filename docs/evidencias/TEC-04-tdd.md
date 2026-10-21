# Evidencia TDD — TEC-04 · Interfaz usable

> Issue del repositorio: #34

TEC-04 es una tarea técnica: no tiene historia de usuario, especificación SDD ni escenarios BDD propios (decisión 5 del [acta de Planning del Sprint 1](../actas/sprint-1/planning.md)). Su alcance son los criterios de aceptación del Issue #34. Se entrega en tres Pull Requests (temas decididos 1, 2 y 5 del [acta de Planning del Sprint 2](../actas/sprint-2/planning.md)), y esta evidencia tiene una sección por cada uno:

1. **Mejoras del servidor HTTP** (CA-TEC04.5 a CA-TEC04.7): esta sección.
2. **Prefijo `/api` y endpoints de lo que ya está en `main`** (CA-TEC04.8 y CA-TEC04.9): [sección del PR 2](#pr-2--prefijo-api-y-endpoints-del-sprint-2).
3. Interfaz web (CA-TEC04.1 a CA-TEC04.4): pendiente.

## PR 1 · Mejoras del servidor HTTP

- **Responsable:** Peñalbé Hernán
- **Rama:** `feature/TEC-04-mejoras-servidor` (creada desde `main` en `3e2e439`)
- **Fecha:** 20/10/2026
- **Go:** go1.27.1 linux/amd64
- **Origen:** tema decidido 2 del acta de Planning del Sprint 2 (revisión del PR #81 de TEC-01): límite de tamaño del cuerpo, rechazo de texto después del JSON y timeouts de `ListenAndServe`, en un Pull Request separado y temprano.

### Trazabilidad

Issue #34 → Criterios de Aceptación → Tests → Código Go

**3 funciones de test nuevas, una por criterio, con 6 subtests.** Los tests de CA-TEC04.5 y CA-TEC04.6 recorren como subtests los tres endpoints que reciben un cuerpo JSON (crear proyecto, modificar proyecto y crear ítem): es el mismo criterio aplicado a cada lugar donde llega un cuerpo, no casos nuevos. Cada subtest verifica además que el pedido rechazado no creó ni modificó nada ("y no se procesa").

```
$ grep -n "^func Test.*\(Handler_\|NewServer\)" internal/httpapi/server_test.go cmd/server/main_test.go
internal/httpapi/server_test.go:430:func TestHandler_CuerpoMayorA1MiBDa413(t *testing.T) {
internal/httpapi/server_test.go:448:func TestHandler_TextoDespuesDelJSONDa400(t *testing.T) {
cmd/server/main_test.go:243:func TestNewServer_Timeouts(t *testing.T) {
```

| Criterio | Test | Subtests | Código |
|---|---|---|---|
| CA-TEC04.5 Un cuerpo de más de 1 MiB se rechaza con 413 y un mensaje entendible, y no se procesa | `TestHandler_CuerpoMayorA1MiBDa413` | crear proyecto, modificar proyecto, crear ítem | `internal/httpapi/server.go`: `maxBodyBytes`, `decodeJSON`, `bodyError`, `errBodyTooLarge` y su caso en `statusFor` |
| CA-TEC04.6 Un cuerpo con texto después del JSON se rechaza como JSON inválido (400) y no se procesa | `TestHandler_TextoDespuesDelJSONDa400` | crear proyecto, modificar proyecto, crear ítem | `internal/httpapi/server.go`: `decodeJSON` (segundo `Decode` que tiene que llegar a `io.EOF`) |
| CA-TEC04.7 Timeouts de lectura de encabezados (5 s), lectura (10 s), escritura (10 s) e inactividad (60 s) | `TestNewServer_Timeouts` | — | `cmd/server/main.go`: `newServer` |

CA-TEC04.8 (prefijo `/api`) y CA-TEC04.1 a CA-TEC04.4 (interfaz web) quedan para los PR 2 y 3.

### Fase RED

Tests escritos antes del código de producción (commit `633f9b1`), solo en `cmd/server/main_test.go` e `internal/httpapi/server_test.go`. La salida se reprodujo sobre ese commit (`git worktree add ... 633f9b1`):

```
$ go test ./...
# seguimiento-medicion-gpc/cmd/server [seguimiento-medicion-gpc/cmd/server.test]
cmd/server/main_test.go:246:9: undefined: newServer
FAIL	seguimiento-medicion-gpc/cmd/server [build failed]
ok  	seguimiento-medicion-gpc/internal/backlog
--- FAIL: TestHandler_CuerpoMayorA1MiBDa413 (0.03s)
    --- FAIL: TestHandler_CuerpoMayorA1MiBDa413/crear_proyecto (0.02s)
        server_test.go:440: código = 201, se esperaba 413; cuerpo: {"id":2,"nombre":"Portal de clientes","descripcion":"Otro proyecto","estado":"PLANIFICADO"}
    --- FAIL: TestHandler_CuerpoMayorA1MiBDa413/modificar_proyecto (0.01s)
        server_test.go:440: código = 200, se esperaba 413; cuerpo: {"id":1,"nombre":"Portal de clientes","descripcion":"Otro proyecto","estado":"PLANIFICADO"}
    --- FAIL: TestHandler_CuerpoMayorA1MiBDa413/crear_ítem (0.01s)
        server_test.go:440: código = 201, se esperaba 413; cuerpo: {"id":1,"proyecto_id":1,"titulo":"Login de usuarios",...}
--- FAIL: TestHandler_TextoDespuesDelJSONDa400 (0.00s)
    --- FAIL: TestHandler_TextoDespuesDelJSONDa400/crear_proyecto (0.00s)
        server_test.go:458: código = 201, se esperaba 400; cuerpo: {"id":2,"nombre":"Portal de clientes","descripcion":"Otro proyecto","estado":"PLANIFICADO"}
    --- FAIL: TestHandler_TextoDespuesDelJSONDa400/modificar_proyecto (0.00s)
        server_test.go:458: código = 200, se esperaba 400; cuerpo: {"id":1,"nombre":"Portal de clientes","descripcion":"Otro proyecto","estado":"PLANIFICADO"}
    --- FAIL: TestHandler_TextoDespuesDelJSONDa400/crear_ítem (0.00s)
        server_test.go:458: código = 201, se esperaba 400; cuerpo: {"id":1,"proyecto_id":1,"titulo":"Login de usuarios",...}
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi
ok  	seguimiento-medicion-gpc/internal/postgres
ok  	seguimiento-medicion-gpc/internal/postgrestest
ok  	seguimiento-medicion-gpc/internal/project
ok  	seguimiento-medicion-gpc/internal/sprint
FAIL
```

El RED falla solo por la funcionalidad nueva, de dos formas:

- **`cmd/server`: error de compilación.** `newServer` (CA-TEC04.7) todavía no existía; `go vet ./...` marca solo ese símbolo.
- **`internal/httpapi`: fallo de aserción.** El paquete compila y fallan únicamente los 6 subtests nuevos: el servidor aceptaba el cuerpo de más de 1 MiB y el texto después del JSON, y respondía 201 o 200. Los demás tests de `httpapi` siguen en verde.

El resto de los paquetes no se tocó y pasa.

En el test de CA-TEC04.5, el relleno que supera 1 MiB son espacios dentro del JSON: el cuerpo sigue siendo un JSON válido (lo único que lo invalida es el tamaño) y, si el test falla, la respuesta no repite el MiB de datos. Una primera versión, nunca commiteada, usaba una descripción de 1 MiB y su salida de fallo imprimía todo ese texto.

### Fase GREEN

Código mínimo de producción (commit `25a05cb`), sin tocar tests:

- `internal/httpapi/server.go`: `decodeJSON` lee el cuerpo con `http.MaxBytesReader` y un límite de 1 MiB (`maxBodyBytes`), y después del JSON exige que un segundo `Decode` llegue a `io.EOF`; `bodyError` traduce el error de lectura a `errBodyTooLarge` (413) o `errInvalidJSON` (400); `statusFor` suma el 413. Los tres handlers que leen un cuerpo (`createProject`, `updateProject` y `createItem`) usan `decodeJSON`.
- `cmd/server/main.go`: `newServer` arma el `http.Server` con los cuatro timeouts, y `main` lo usa en lugar de `http.ListenAndServe`.

```
$ git diff --stat 633f9b1 25a05cb
 cmd/server/main.go         | 16 +++++++++++++++-
 internal/httpapi/server.go | 46 +++++++++++++++++++++++++++++++++++++++-------
 2 files changed, 54 insertions(+), 8 deletions(-)
```

```
$ go test -v ./internal/httpapi/ ./cmd/server/ -run 'Handler_|NewServer'
--- PASS: TestHandler_CuerpoMayorA1MiBDa413
    --- PASS: TestHandler_CuerpoMayorA1MiBDa413/crear_proyecto
    --- PASS: TestHandler_CuerpoMayorA1MiBDa413/modificar_proyecto
    --- PASS: TestHandler_CuerpoMayorA1MiBDa413/crear_ítem
--- PASS: TestHandler_TextoDespuesDelJSONDa400
    --- PASS: TestHandler_TextoDespuesDelJSONDa400/crear_proyecto
    --- PASS: TestHandler_TextoDespuesDelJSONDa400/modificar_proyecto
    --- PASS: TestHandler_TextoDespuesDelJSONDa400/crear_ítem
--- PASS: TestNewServer_Timeouts
```

Los espacios o el salto de línea después del JSON se siguen aceptando, porque el decodificador de JSON de Go los ignora. Ningún criterio lo pide, así que no tiene un test propio; se comprobó a mano con el servidor levantado (`printf '{"nombre":"Demo","descripcion":""}\n' | curl --data-binary @- ...` responde 201 y `curl -d '{"nombre":"X"} basura' ...` responde 400).

### Refactor

No se hizo: `decodeJSON` y `bodyError` reemplazan las tres llamadas repetidas a `json.NewDecoder(r.Body).Decode` y `newServer` es una construcción directa. No hay lógica duplicada ni ramas sin test.

### Validación final

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)

$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server
ok  	seguimiento-medicion-gpc/internal/backlog
ok  	seguimiento-medicion-gpc/internal/httpapi
ok  	seguimiento-medicion-gpc/internal/postgres
ok  	seguimiento-medicion-gpc/internal/postgrestest
ok  	seguimiento-medicion-gpc/internal/project
ok  	seguimiento-medicion-gpc/internal/sprint

$ go test -race ./...
(los 7 paquetes ok)
```

### Pruebas de mutación

Una por vez, rompiendo el código a propósito y restaurándolo con `git checkout` (ninguna forma parte de un commit):

| Mutación | Tests que fallan |
|---|---|
| a) Sacar el segundo `Decode` de `decodeJSON` | Los 3 subtests de `TestHandler_TextoDespuesDelJSONDa400` |
| b) Leer el cuerpo sin `http.MaxBytesReader` | Los 3 subtests de `TestHandler_CuerpoMayorA1MiBDa413` |
| c) Volver `updateProject` al decodificador anterior | El subtest `modificar_proyecto` de los dos tests nuevos (y `TestModificarProyecto_JSONInvalidoDevuelve400`, porque el error ya no se traduce) |
| d) `ReadHeaderTimeout` en 0 | `TestNewServer_Timeouts` |

La mutación c) muestra para qué sirven los subtests por endpoint: con un solo endpoint probado, un handler que vuelva a leer el cuerpo directo no se detectaría.

### Cobertura

Valores tomados del log del CI del PR #104 (Linux), job "gofmt y tests":

```
$ go test -race -coverprofile=coverage.out ./...
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 40.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 82.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	coverage: 77.3% of statements
$ go tool cover -func=coverage.out | tail -n 1
total:									(statements)		75.8%
```

`cmd/server` sube de 39.0% en `main` a 40.5% con `newServer`, e `internal/httpapi` sigue en 100.0% con `decodeJSON` y `bodyError`. El total pasa de 75.3% (`main` con HU-09) a 75.8%. Las funciones nuevas quedan al 100.0% (`go tool cover -func`): `newServer`, `decodeJSON` y `bodyError`. TEC-04 no toca PostgreSQL.

### Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | `633f9b1` | `test(TEC-04): agregar tests del límite del cuerpo, el texto después del JSON y los timeouts (RED)` |
| GREEN | `25a05cb` | `feat(TEC-04): implementar el límite del cuerpo, el rechazo de texto después del JSON y los timeouts (GREEN)` |

No hubo commit de refactor porque no se hizo refactor.

## PR 2 · Prefijo `/api` y endpoints del Sprint 2

- **Responsable:** Peñalbé Hernán
- **Revisor:** Sosa Ricardo
- **Rama:** `feature/TEC-04-api` (creada desde `main` en `76305f9`, el merge del PR #104)
- **Fecha:** 21/10/2026
- **Go:** go1.27.1 linux/amd64
- **Origen:** CA-TEC04.8 y CA-TEC04.9 del Issue #34. La interfaz web del PR 3 necesita la API bajo `/api` (para servir la interfaz en `/`) y los endpoints de las historias del Sprint 2.

### Trazabilidad

Issue #34 → Criterios de Aceptación → Tests → Código Go

Regla de los tests de endpoints, acordada antes del RED: por endpoint, un caso exitoso y un test por cada código de error que puede devolver. Los errores que comparten código van como subtests: verifican la traducción a HTTP, no vuelven a probar las reglas de negocio, que ya prueba cada historia en su paquete. Los endpoints nuevos con id suman filas a `TestRutasConID_IDNoNumericoDevuelve400`, y los que reciben un cuerpo, filas a `bodyEndpoints` (CA-TEC04.5 y CA-TEC04.6).

**26 funciones de test nuevas (11 de ítems y 15 de sprints) con 21 subtests (4 en modificar ítem, 7 en crear sprint, 2 + 3 en asignar y 2 + 3 en completar), y 13 filas nuevas en tests existentes** (10 en `TestRutasConID_IDNoNumericoDevuelve400` y 3 en `bodyEndpoints`, que corren en los 2 tests de CA-TEC04.5 y CA-TEC04.6). Además, los 23 tests que ya tenía `internal/httpapi` pasaron sus rutas al prefijo `/api`. En total, `internal/httpapi` tiene 49 funciones de test.

```
$ grep -n "^func Test" internal/httpapi/items_test.go internal/httpapi/sprints_test.go
internal/httpapi/items_test.go:35:func TestConsultarItem_Devuelve200ConElItem(t *testing.T) {
internal/httpapi/items_test.go:50:func TestConsultarItem_InexistenteDevuelve404(t *testing.T) {
internal/httpapi/items_test.go:57:func TestModificarItem_Devuelve200ConElItemModificado(t *testing.T) {
internal/httpapi/items_test.go:80:func TestModificarItem_DatosInvalidosDevuelven400(t *testing.T) {
internal/httpapi/items_test.go:101:func TestModificarItem_InexistenteDevuelve404(t *testing.T) {
internal/httpapi/items_test.go:108:func TestCambiarEstadoItem_Devuelve200ConElEstadoNuevo(t *testing.T) {
internal/httpapi/items_test.go:121:func TestCambiarEstadoItem_EstadoInvalidoDevuelve400(t *testing.T) {
internal/httpapi/items_test.go:129:func TestCambiarEstadoItem_InexistenteDevuelve404(t *testing.T) {
internal/httpapi/items_test.go:135:func TestCambiarEstadoItem_TransicionNoPermitidaDevuelve409(t *testing.T) {
internal/httpapi/items_test.go:146:func TestListarBacklogPorPrioridad_Devuelve200OrdenadoPorPrioridad(t *testing.T) {
internal/httpapi/items_test.go:163:func TestListarBacklogPorPrioridad_ProyectoInexistenteDevuelve404(t *testing.T) {
internal/httpapi/sprints_test.go:48:func TestCrearSprint_Devuelve201ConElSprintCreado(t *testing.T) {
internal/httpapi/sprints_test.go:64:func TestCrearSprint_DatosInvalidosDevuelven400(t *testing.T) {
internal/httpapi/sprints_test.go:91:func TestCrearSprint_ProyectoInexistenteDevuelve404(t *testing.T) {
internal/httpapi/sprints_test.go:97:func TestCrearSprint_SprintAbiertoExistenteDevuelve409(t *testing.T) {
internal/httpapi/sprints_test.go:107:func TestAsignarItem_Devuelve200ConElItemAsignado(t *testing.T) {
internal/httpapi/sprints_test.go:120:func TestAsignarItem_InexistenteDevuelve404(t *testing.T) {
internal/httpapi/sprints_test.go:139:func TestAsignarItem_ConflictosDevuelven409(t *testing.T) {
internal/httpapi/sprints_test.go:182:func TestCompletarItem_Devuelve200ConElItemTerminado(t *testing.T) {
internal/httpapi/sprints_test.go:197:func TestCompletarItem_InexistenteDevuelve404(t *testing.T) {
internal/httpapi/sprints_test.go:216:func TestCompletarItem_ConflictosDevuelven409(t *testing.T) {
internal/httpapi/sprints_test.go:253:func TestListarCompletados_Devuelve200ConLosItemsTerminados(t *testing.T) {
internal/httpapi/sprints_test.go:271:func TestListarCompletados_SprintInexistenteDevuelve404(t *testing.T) {
internal/httpapi/sprints_test.go:278:func TestCerrarSprint_Devuelve200ConElSprintCerrado(t *testing.T) {
internal/httpapi/sprints_test.go:291:func TestCerrarSprint_InexistenteDevuelve404(t *testing.T) {
internal/httpapi/sprints_test.go:297:func TestCerrarSprint_YaCerradoDevuelve409(t *testing.T) {
```

**CA-TEC04.8 · La API responde bajo `/api/…`.** Los 23 tests existentes de `internal/httpapi` usan rutas con el prefijo. Código: las rutas de `NewHandler` (`internal/httpapi/server.go`). La interfaz web en `/` queda para el PR 3.

**CA-TEC04.9 · La API expone las operaciones integradas en el Sprint 2 (HU-05, HU-07 a HU-13), con el mismo formato de errores que TEC-01.** Contrato aprobado antes del RED: cada endpoint llama al método del Manager de su historia, y la respuesta de un ítem suma `horas_estimadas` y `sprint_id`.

| Endpoint | HU | Éxito | Errores | Código |
|---|---|---|---|---|
| `GET /api/items/{id}` | HU-05 | `TestConsultarItem_Devuelve200ConElItem` | 404 `TestConsultarItem_InexistenteDevuelve404` · 400 fila "consultar ítem" | `items.go`: `getItem` → `GetItem` |
| `PUT /api/items/{id}` | HU-07 | `TestModificarItem_Devuelve200ConElItemModificado` | 400 `TestModificarItem_DatosInvalidosDevuelven400` (4 subtests: título vacío, prioridad inválida, sin criterios, horas negativas) · 404 `TestModificarItem_InexistenteDevuelve404` · 400 fila "modificar ítem" · 413 y 400 fila "modificar ítem" en `bodyEndpoints` | `items.go`: `updateItem` → `Update` |
| `PATCH /api/items/{id}/estado` | HU-08 | `TestCambiarEstadoItem_Devuelve200ConElEstadoNuevo` | 400 `TestCambiarEstadoItem_EstadoInvalidoDevuelve400` · 404 `TestCambiarEstadoItem_InexistenteDevuelve404` · 409 `TestCambiarEstadoItem_TransicionNoPermitidaDevuelve409` · 400 fila "cambiar estado" · 413 y 400 fila "cambiar estado" en `bodyEndpoints` | `items.go`: `changeItemStatus` → `ChangeStatus` |
| `GET /api/proyectos/{id}/backlog?orden=prioridad` | HU-09 | `TestListarBacklogPorPrioridad_Devuelve200OrdenadoPorPrioridad` | 404 `TestListarBacklogPorPrioridad_ProyectoInexistenteDevuelve404` | `items.go`: `listItems` → `ListItemsByPriority` (sin el parámetro, `ListItems`) |
| `POST /api/proyectos/{id}/sprints` | HU-10 | `TestCrearSprint_Devuelve201ConElSprintCreado` | 400 `TestCrearSprint_DatosInvalidosDevuelven400` (7 subtests: nombre vacío, objetivo vacío, sin fecha de inicio, sin fecha de fin, fin antes de inicio, fecha de inicio con otro formato, fecha de fin con otro formato) · 404 `TestCrearSprint_ProyectoInexistenteDevuelve404` · 409 `TestCrearSprint_SprintAbiertoExistenteDevuelve409` · 400 fila "crear sprint" · 413 y 400 fila "crear sprint" en `bodyEndpoints` | `sprints.go`: `createSprint` → `Create`, `parseDate` |
| `POST /api/sprints/{id}/items/{itemId}` | HU-11 | `TestAsignarItem_Devuelve200ConElItemAsignado` | 404 `TestAsignarItem_InexistenteDevuelve404` (2 subtests: sprint, ítem) · 409 `TestAsignarItem_ConflictosDevuelven409` (3 subtests: sprint cerrado, ítem de otro proyecto, ítem ya asignado) · 400 filas "asignar ítem: sprint" y "asignar ítem: ítem" | `sprints.go`: `assignItem` → `AssignItem`, `sprintItemIDs` |
| `POST /api/sprints/{id}/items/{itemId}/completar` | HU-12 | `TestCompletarItem_Devuelve200ConElItemTerminado` | 404 `TestCompletarItem_InexistenteDevuelve404` (2 subtests: sprint, ítem) · 409 `TestCompletarItem_ConflictosDevuelven409` (3 subtests: sprint cerrado, ítem no asignado al sprint, transición no permitida) · 400 filas "completar ítem: sprint" y "completar ítem: ítem" | `sprints.go`: `completeItem` → `CompleteItem` |
| `GET /api/sprints/{id}/completados` | HU-12 | `TestListarCompletados_Devuelve200ConLosItemsTerminados` | 404 `TestListarCompletados_SprintInexistenteDevuelve404` · 400 fila "listar completados" | `sprints.go`: `listCompletedItems` → `ListCompletedItems` |
| `POST /api/sprints/{id}/cerrar` | HU-13 | `TestCerrarSprint_Devuelve200ConElSprintCerrado` | 404 `TestCerrarSprint_InexistenteDevuelve404` · 409 `TestCerrarSprint_YaCerradoDevuelve409` · 400 fila "cerrar sprint" | `sprints.go`: `closeSprint` → `Close` |

Los códigos siguen la tabla aprobada: 400 para un dato inválido, 404 si el proyecto, el ítem o el sprint no existe, y 409 si el pedido choca con el estado actual (`statusFor` en `server.go`). Las fechas de un sprint usan el formato `AAAA-MM-DD`; una fecha ausente la rechaza el Manager de sprints con su propio error.

### Fase RED

Tres commits de tests, cada uno antes de su código. Las salidas se reprodujeron sobre cada commit con `git worktree add`.

**1. Prefijo `/api` (commit `d9b4aeb`).** Los tests existentes pasan sus rutas a `/api/…`. Fallan los 23 tests de `internal/httpapi`, todos porque la ruta no existe y el mux responde `404 page not found` en texto plano:

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server
ok  	seguimiento-medicion-gpc/internal/backlog
--- FAIL: TestCrearProyecto_Devuelve201ConElProyectoCreado (0.00s)
    server_test.go:113: POST /api/proyectos: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    server_test.go:115: código = 404, se esperaba 201; cuerpo: 404 page not found
--- FAIL: TestConsultarProyecto_IDNoNumericoDevuelve400 (0.00s)
    server_test.go:171: GET /api/proyectos/abc: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    server_test.go:173: código = 404, se esperaba 400; cuerpo: 404 page not found
(... el mismo fallo en los 23 tests ...)
FAIL	seguimiento-medicion-gpc/internal/httpapi
ok  	seguimiento-medicion-gpc/internal/postgres
ok  	seguimiento-medicion-gpc/internal/postgrestest
ok  	seguimiento-medicion-gpc/internal/project
ok  	seguimiento-medicion-gpc/internal/sprint
FAIL
```

**2. Endpoints de ítems (commit `c8618a9`).** Fallan 13 de los 34 tests de `internal/httpapi`: 10 de los 11 nuevos y los 3 existentes que suman filas (`TestRutasConID_IDNoNumericoDevuelve400`, `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400`, solo en sus filas de ítems):

```
--- FAIL: TestConsultarItem_Devuelve200ConElItem (0.01s)
    items_test.go:38: GET /api/items/1: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    items_test.go:40: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestModificarItem_DatosInvalidosDevuelven400 (0.00s)
    --- FAIL: TestModificarItem_DatosInvalidosDevuelven400/horas_negativas (0.00s)
        items_test.go:96: código = 404, se esperaba 400; cuerpo: 404 page not found
--- FAIL: TestCambiarEstadoItem_TransicionNoPermitidaDevuelve409 (0.00s)
    items_test.go:140: código = 404, se esperaba 409; cuerpo: 404 page not found
--- FAIL: TestListarBacklogPorPrioridad_Devuelve200OrdenadoPorPrioridad (0.00s)
    items_test.go:159: ítems = [{ID:1 ... Prioridad:BAJA ...} {ID:2 ... Prioridad:MEDIA ...} {ID:3 ... Prioridad:ALTA ...}], se esperaban los ítems 3 (ALTA), 2 (MEDIA) y 1 (BAJA)
--- FAIL: TestHandler_CuerpoMayorA1MiBDa413 (0.06s)
    --- FAIL: TestHandler_CuerpoMayorA1MiBDa413/modificar_ítem (0.00s)
        server_test.go:479: código = 404, se esperaba 413; cuerpo: 404 page not found
(...)
FAIL	seguimiento-medicion-gpc/internal/httpapi
```

`TestListarBacklogPorPrioridad_ProyectoInexistenteDevuelve404` ya pasaba en el RED: la ruta del backlog existía y respondía 404 al proyecto inexistente sin mirar el parámetro. Se dejó igual porque es el test del código de error de la tabla para ese endpoint, y protege ese 404 cuando el GREEN agrega el orden.

**3. Endpoints de sprints (commit `279bd30`).** Los tests le pasan a `NewHandler` el `sprint.Manager`, que todavía no recibía. El paquete no compila; esas dos llamadas son los únicos errores:

```
$ go test ./...
# seguimiento-medicion-gpc/internal/httpapi_test [seguimiento-medicion-gpc/internal/httpapi.test]
internal/httpapi/server_test.go:61:45: too many arguments in call to httpapi.NewHandler
internal/httpapi/server_test.go:379:43: too many arguments in call to httpapi.NewHandler
ok  	seguimiento-medicion-gpc/cmd/server
ok  	seguimiento-medicion-gpc/internal/backlog
FAIL	seguimiento-medicion-gpc/internal/httpapi [build failed]
ok  	seguimiento-medicion-gpc/internal/postgres
ok  	seguimiento-medicion-gpc/internal/postgrestest
ok  	seguimiento-medicion-gpc/internal/project
ok  	seguimiento-medicion-gpc/internal/sprint
FAIL
```

En los tres casos el fallo corresponde solo a la funcionalidad nueva y los demás paquetes pasan.

### Fase GREEN

- **`a870eee` (prefijo `/api`):** las rutas de `NewHandler` pasan a `/api/…`.
- **`74f8d97` (ítems):** rutas `GET` y `PUT /api/items/{id}` y `PATCH /api/items/{id}/estado`; `listItems` usa `ListItemsByPriority` solo con `?orden=prioridad`; `itemRequest` suma `horas_estimadas`, hay un `statusRequest` con `estado` y `itemResponse` suma `horas_estimadas` y `sprint_id`; `statusFor` suma `ErrInvalidEstimatedHours` y `ErrInvalidStatus` (400), `ErrItemNotFound` (404) y `ErrInvalidTransition` (409).
- **`bbfa0eb` (sprints):** `NewHandler(projects, items, sprints)` y las 5 rutas de sprints; `sprintRequest` y `sprintResponse`; `parseDate` (vacía = ausente, otro formato = `errInvalidDate`), `pathInt` y `sprintItemIDs`; `statusFor` suma las validaciones de sprint y la fecha inválida (400), `ErrSprintNotFound` (404) y los conflictos de sprint (409). `cmd/server/main.go` arma el `sprint.Manager` y se lo pasa a `NewHandler`.

En `bbfa0eb`, la primera versión dejaba `internal/httpapi` en 99.1%: la rama de error de `fecha_fin` no la cubría ningún test, porque el subtest de formato usaba `fecha_inicio`. Antes del commit se juntó la validación de las dos fechas con `cmp.Or`, y el paquete volvió a 100.0%. La mutación d) de abajo mostró que esa rama igual necesitaba su propio test.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server
ok  	seguimiento-medicion-gpc/internal/backlog
ok  	seguimiento-medicion-gpc/internal/httpapi
ok  	seguimiento-medicion-gpc/internal/postgres
ok  	seguimiento-medicion-gpc/internal/postgrestest
ok  	seguimiento-medicion-gpc/internal/project
ok  	seguimiento-medicion-gpc/internal/sprint
```

### Desviación del ciclo TDD: subtest de la fecha de fin

Después del GREEN, la mutación d) (ignorar el error de formato de `fecha_fin`) sobrevivía: una `fecha_fin` como `14/10/2026` respondía 400 con "la fecha de fin del sprint es obligatoria" en vez de "las fechas deben tener el formato AAAA-MM-DD", y ningún test fallaba. La tabla aprobada dice "fecha con formato distinto de AAAA-MM-DD → 400", lo que abarca las dos fechas, pero el subtest solo probaba `fecha_inicio`.

Se agregó el subtest "fecha de fin con otro formato" (commit `9ac78da`) y se renombró el anterior a "fecha de inicio con otro formato". Pasa con el código del GREEN, sin cambios de producción, y falla con la mutación d):

```
--- FAIL: TestCrearSprint_DatosInvalidosDevuelven400 (0.00s)
    --- FAIL: TestCrearSprint_DatosInvalidosDevuelven400/fecha_de_fin_con_otro_formato (0.00s)
        sprints_test.go:86: error = "la fecha de fin del sprint es obligatoria", se esperaba "las fechas deben tener el formato AAAA-MM-DD"
```

### Refactor

Commit `341a985`: `server.go` había pasado de 255 a unas 500 líneas, y los tests ya estaban separados por tema. Se separaron los handlers de la misma forma, sin cambiar lógica:

| Archivo | Contenido |
|---|---|
| `server.go` | `NewHandler`, `server`, errores, `decodeJSON`, `bodyError`, `pathID`, `pathInt`, `statusFor`, `writeError`, `writeJSON` |
| `projects.go` | Proyectos (HU-01 a HU-03) |
| `items.go` | Backlog e ítems (HU-04 a HU-09) |
| `sprints.go` | Sprints (HU-10 a HU-13), `parseDate` y `sprintItemIDs` |

Se comparó el conjunto de líneas de código antes y después (sin importar el orden): las únicas diferencias son el `package`, los `import` y el comentario de cabecera de cada archivo nuevo. Los 49 tests siguen en verde y el paquete sigue en 100.0%.

### Validación final

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)

$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server
ok  	seguimiento-medicion-gpc/internal/backlog
ok  	seguimiento-medicion-gpc/internal/httpapi
ok  	seguimiento-medicion-gpc/internal/postgres
ok  	seguimiento-medicion-gpc/internal/postgrestest
ok  	seguimiento-medicion-gpc/internal/project
ok  	seguimiento-medicion-gpc/internal/sprint

$ go test -race ./...
(los 7 paquetes ok)
```

Además, los 12 ejemplos de `curl` del README se corrieron contra el servidor levantado (`go build ./cmd/server`, datos en memoria): crear, consultar y modificar un proyecto, crear un ítem, listar por prioridad, modificar el ítem con horas estimadas, crear un sprint, asignar el ítem, pasarlo a `EN_CURSO`, completarlo, listar los completados y cerrar el sprint. Respondieron 201 o 200 con el JSON esperado.

### Pruebas de mutación

Una por vez, rompiendo el código a propósito y restaurándolo con `git checkout` (ninguna forma parte de un commit):

| Mutación | Tests que fallan |
|---|---|
| a) Ignorar `?orden=prioridad` en `listItems` | `TestListarBacklogPorPrioridad_Devuelve200OrdenadoPorPrioridad` |
| b) Sacar `ErrSprintClosed` del 409 en `statusFor` | `TestAsignarItem_ConflictosDevuelven409/sprint_cerrado`, `TestCompletarItem_ConflictosDevuelven409/sprint_cerrado` y `TestCerrarSprint_YaCerradoDevuelve409` |
| c) Leer el `{itemId}` desde `{id}` | Las filas "asignar ítem: ítem" y "completar ítem: ítem" de `TestRutasConID_IDNoNumericoDevuelve400`, `TestAsignarItem_InexistenteDevuelve404/ítem_inexistente`, `TestAsignarItem_ConflictosDevuelven409/ítem_de_otro_proyecto`, `TestCompletarItem_InexistenteDevuelve404/ítem_inexistente` y `TestListarCompletados_Devuelve200ConLosItemsTerminados` |
| d) Ignorar el error de formato de `fecha_fin` | Ninguno con los tests del GREEN; con el subtest de `9ac78da`, `TestCrearSprint_DatosInvalidosDevuelven400/fecha_de_fin_con_otro_formato` |

### Cobertura

Medición local con el mismo comando del CI, sin base (go1.27.1 linux/amd64):

```
$ go test -race -coverprofile=coverage.out ./...
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 37.8% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 82.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	coverage: 77.3% of statements
$ go tool cover -func=coverage.out | tail -n 1
total:									(statements)		79.3%
```

`internal/httpapi` sigue en 100.0%, con todas las funciones nuevas al 100.0%. El total sube de 75.8% (PR 1) a 79.3%, porque `httpapi` suma código totalmente cubierto. `cmd/server` baja de 40.5% a 37.8% porque `main()`, que no se prueba con tests unitarios, suma las dos líneas del `sprint.Manager`. TEC-04 no toca PostgreSQL.

### Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | `d9b4aeb` | `test(TEC-04): llevar las rutas de los tests de la API al prefijo /api (RED)` |
| GREEN | `a870eee` | `feat(TEC-04): servir la API bajo el prefijo /api (GREEN)` |
| RED | `c8618a9` | `test(TEC-04): agregar tests de los endpoints de ítems (RED)` |
| GREEN | `74f8d97` | `feat(TEC-04): exponer los endpoints de ítems en la API (GREEN)` |
| RED | `279bd30` | `test(TEC-04): agregar tests de los endpoints de sprints (RED)` |
| GREEN | `bbfa0eb` | `feat(TEC-04): exponer los endpoints de sprints en la API (GREEN)` |
| Test (desviación) | `9ac78da` | `test(TEC-04): cubrir el formato de la fecha de fin del sprint` |
| Refactor | `341a985` | `refactor(TEC-04): separar los handlers de la API por tema` |
