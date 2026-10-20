# Evidencia TDD — TEC-04 · Interfaz usable

> Issue del repositorio: #34

TEC-04 es una tarea técnica: no tiene historia de usuario, especificación SDD ni escenarios BDD propios (decisión 5 del [acta de Planning del Sprint 1](../actas/sprint-1/planning.md)). Su alcance son los criterios de aceptación del Issue #34. Se entrega en tres Pull Requests (temas decididos 1, 2 y 5 del [acta de Planning del Sprint 2](../actas/sprint-2/planning.md)), y esta evidencia tiene una sección por cada uno:

1. **Mejoras del servidor HTTP** (CA-TEC04.5 a CA-TEC04.7): esta sección.
2. Prefijo `/api` y endpoints de lo que ya está en `main` (CA-TEC04.8): pendiente.
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

**Pendiente:** se completa con el log del CI del Pull Request (job "gofmt y tests").

### Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | `633f9b1` | `test(TEC-04): agregar tests del límite del cuerpo, el texto después del JSON y los timeouts (RED)` |
| GREEN | `25a05cb` | `feat(TEC-04): implementar el límite del cuerpo, el rechazo de texto después del JSON y los timeouts (GREEN)` |

No hubo commit de refactor porque no se hizo refactor.
