# Evidencia TDD — TEC-01 · Punto de entrada ejecutable mínimo

> Issue del repositorio: #23

- **Responsable:** Peñalbé Hernán
- **Rama:** `feature/TEC-01-servidor-http`
- **Fecha:** 14/10/2026
- **Go:** go1.27.1 linux/amd64

TEC-01 es una tarea técnica: no tiene historia de usuario, especificación SDD ni escenarios BDD propios (decisión 5 del [acta de Planning del Sprint 1](../actas/sprint-1-planning.md)). Su alcance son los criterios de aceptación del Issue #23. Los endpoints no agregan reglas de negocio: llaman a los `Manager` de las historias ya integradas en `main`, que siguen validadas por sus propias SDD, BDD y tests.

## Trazabilidad

Issue #23 → Criterios de Aceptación → Tests → Código Go

**23 tests (21 en `internal/httpapi/server_test.go` y 2 en `cmd/server/main_test.go`), 3 de ellos con subtests (3 + 3 + 3 = 9 subtests).** La correspondencia con los criterios no es 1 a 1: CA-TEC01.2 y CA-TEC01.3 los cubren los mismos 21 tests de los endpoints.

| Criterio | Tests | Código |
|---|---|---|
| CA-TEC01.1 Arranca con un solo comando y escucha en un puerto configurable | `TestAddress_UsaLaVariablePORT`, `TestAddress_SinPORTUsa8080`; prueba manual (abajo) | `cmd/server/main.go` (`main`, `address`) |
| CA-TEC01.2 Expone las operaciones de proyectos y de backlog integradas en `main` | los 21 tests de `server_test.go` (ver la tabla de endpoints) | `internal/httpapi/server.go` |
| CA-TEC01.3 Pruebas automatizadas de los endpoints | los 21 tests de `server_test.go`, con `net/http/httptest` | `internal/httpapi/server.go` |
| CA-TEC01.4 El README explica cómo ejecutarla | — (documentación) | `README.md`, sección "Cómo ejecutar" |
| CA-TEC01.5 Sin librerías externas; datos en memoria | — (verificado con `go list`, abajo) | `cmd/server/main.go` (repositorios en memoria) |

### Endpoints → historias

| Endpoint | Historia | Especificación | Tests |
|---|---|---|---|
| `POST /proyectos` | HU-01 | [HU-01](../especificaciones/HU-01-crear-proyecto.md) | `TestCrearProyecto_Devuelve201ConElProyectoCreado`, `TestCrearProyecto_NombreVacioDevuelve400`, `TestCrearProyecto_NombreRepetidoDevuelve409`, `TestCrearProyecto_JSONInvalidoDevuelve400` |
| `GET /proyectos/{id}` | HU-03 | [HU-03](../especificaciones/HU-03-consultar-proyecto.md) | `TestConsultarProyecto_Devuelve200ConElProyecto`, `TestConsultarProyecto_InexistenteDevuelve404` (3 casos), `TestConsultarProyecto_IDNoNumericoDevuelve400` |
| `PUT /proyectos/{id}` | HU-02 | [HU-02](../especificaciones/HU-02-modificar-proyecto.md) | `TestModificarProyecto_Devuelve200ConElProyectoModificado`, `TestModificarProyecto_EstadoInvalidoDevuelve400`, `TestModificarProyecto_NombreRepetidoDevuelve409`, `TestModificarProyecto_InexistenteDevuelve404`, `TestModificarProyecto_JSONInvalidoDevuelve400` |
| `POST /proyectos/{id}/backlog` | HU-04 | [HU-04](../especificaciones/HU-04-crear-item-backlog.md) | `TestCrearItem_Devuelve201ConElItemCreado`, `TestCrearItem_DatosInvalidosDevuelven400` (3 casos), `TestCrearItem_ProyectoInexistenteDevuelve404`, `TestCrearItem_JSONInvalidoDevuelve400` |
| `GET /proyectos/{id}/backlog` | HU-06 | [HU-06](../especificaciones/HU-06-listar-backlog.md) | `TestListarBacklog_Devuelve200ConLosItemsDelProyecto`, `TestListarBacklog_SinItemsDevuelveListaVacia`, `TestListarBacklog_ProyectoInexistenteDevuelve404` |
| Todas las rutas con `{id}` | — | — | `TestRutasConID_IDNoNumericoDevuelve400` (3 casos) |
| Error inesperado | — | — | `TestErrorInterno_Devuelve500SinDetalle` |

Códigos de respuesta: 400 para datos inválidos, JSON inválido o id no numérico; 404 para proyecto inexistente; 409 para nombre repetido; 500 para cualquier otro error, sin exponer el detalle interno. Las respuestas de las cinco rutas son `application/json`. Una ruta inexistente (por ejemplo `GET /foo`) o un método no definido (por ejemplo `DELETE /proyectos/1`) los responde directamente `http.ServeMux` con 404 o 405 en `text/plain`.

## Fase RED

Tests escritos antes del código (commit ffc4999). El fallo corresponde solo a la funcionalidad nueva: el paquete `internal/httpapi` todavía no tenía código y `address` no estaba declarada. `internal/backlog` e `internal/project` siguen en `ok`.

```
$ go test ./...
seguimiento-medicion-gpc/internal/httpapi: no non-test Go files in internal/httpapi
# seguimiento-medicion-gpc/cmd/server [seguimiento-medicion-gpc/cmd/server.test]
cmd/server/main_test.go:14:12: undefined: address
cmd/server/main_test.go:22:12: undefined: address
FAIL	seguimiento-medicion-gpc/cmd/server [build failed]
ok  	seguimiento-medicion-gpc/internal/backlog	0.058s
FAIL	seguimiento-medicion-gpc/internal/httpapi [build failed]
ok  	seguimiento-medicion-gpc/internal/project	0.071s
FAIL
```

## Fase GREEN

Código mínimo (commit 07d50b5): `internal/httpapi/server.go` (`NewHandler` con las 5 rutas de `http.ServeMux`, los DTO en español y la traducción de errores a códigos HTTP) y `cmd/server/main.go` (arma los `Manager` con los repositorios en memoria y escucha en `address(os.Getenv)`).

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.004s
ok  	seguimiento-medicion-gpc/internal/backlog	0.057s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.012s
ok  	seguimiento-medicion-gpc/internal/project	0.072s
```

### Desviación del ciclo TDD

Los tests del RED probaban el id no numérico solo en `GET /proyectos/{id}`. En el GREEN se escribió la misma validación (`pathID`) en `PUT /proyectos/{id}`, `POST /proyectos/{id}/backlog` y `GET /proyectos/{id}/backlog` sin un test que la pidiera, así que es una desviación del ciclo. La detectó la revisión de `go tool cover -func` antes del commit GREEN: `updateProject` y `createItem` quedaban en 84.6%, `listItems` en 83.3% e `internal/httpapi` en 92.7%.

Como el código ya existía, el test que la cubre (`TestRutasConID_IDNoNumericoDevuelve400`, commit e043603) pasa desde el primer momento: no hay una fase RED real y no se fingió una. Para demostrar que el test protege las tres ramas, se hizo una prueba de mutación: en cada handler, por separado, se ignoró el error de `pathID`, se corrió el test y se restauró el archivo original (verificado con `cmp` contra una copia).

```
== mutante: updateProject ignora el id inválido
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.00s)
        server_test.go:186: código = 404, se esperaba 400; cuerpo: {"error":"no existe un proyecto con ese id"}
FAIL
== mutante: createItem ignora el id inválido
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.00s)
        server_test.go:186: código = 404, se esperaba 400; cuerpo: {"error":"no existe un proyecto con ese id"}
FAIL
== mutante: listItems ignora el id inválido
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.00s)
        server_test.go:186: código = 404, se esperaba 400; cuerpo: {"error":"no existe un proyecto con ese id"}
FAIL
== código restaurado
$ cmp internal/httpapi/server.go server.go.bak && echo restaurado-identico
restaurado-identico
```

Acción para la Retrospectiva del Sprint 1 (#65), la misma que en HU-02: revisar `go tool cover -func` antes de cada commit GREEN, y en el RED escribir un test por cada ruta o rama que se vaya a implementar, no solo por la primera.

## Refactor

No aplica. El código de producción quedó sin duplicación que justifique un cambio: cada handler lee el id, decodifica el cuerpo si lo tiene, llama a un `Manager` y responde con `writeJSON`/`writeError`. Extraer un helper común para "leer id + decodificar" ahorraría pocas líneas y haría menos directa la lectura de cada handler.

## Prueba manual

El servidor se compiló con `go build ./cmd/server` y se ejecutó con `PORT=8099` (CA-TEC01.1: puerto configurable) y se probó con `curl`:

```
servidor escuchando en :8099

$ curl -s -X POST localhost:8099/proyectos -d '{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas"}'
{"id":1,"nombre":"Sistema de ventas","descripcion":"Gestión de ventas","estado":"PLANIFICADO"}
  [HTTP 201]
$ curl -s -X POST localhost:8099/proyectos -d '{"nombre":"Sistema de ventas"}'
{"error":"ya existe un proyecto con ese nombre"}
  [HTTP 409]
$ curl -s -X GET localhost:8099/proyectos/1
{"id":1,"nombre":"Sistema de ventas","descripcion":"Gestión de ventas","estado":"PLANIFICADO"}
  [HTTP 200]
$ curl -s -X PUT localhost:8099/proyectos/1 -d '{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas","estado":"EN_CURSO"}'
{"id":1,"nombre":"Sistema de ventas","descripcion":"Gestión de ventas","estado":"EN_CURSO"}
  [HTTP 200]
$ curl -s -X POST localhost:8099/proyectos/1/backlog -d '{"titulo":"Registrar venta","descripcion":"Alta de una venta","prioridad":"ALTA","criterios":["Se guarda la venta"]}'
{"id":1,"proyecto_id":1,"titulo":"Registrar venta","descripcion":"Alta de una venta","prioridad":"ALTA","criterios":["Se guarda la venta"],"estado":"PENDIENTE","story_points":0}
  [HTTP 201]
$ curl -s -X GET localhost:8099/proyectos/1/backlog
[{"id":1,"proyecto_id":1,"titulo":"Registrar venta","descripcion":"Alta de una venta","prioridad":"ALTA","criterios":["Se guarda la venta"],"estado":"PENDIENTE","story_points":0}]
  [HTTP 200]
$ curl -s -X GET localhost:8099/proyectos/999
{"error":"no existe un proyecto con ese id"}
  [HTTP 404]
$ curl -s -X GET localhost:8099/proyectos/abc/backlog
{"error":"el id debe ser un número entero"}
  [HTTP 400]
```

(`[HTTP nnn]` es el código de respuesta, impreso con `curl -w`.)

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)

$ go vet ./...
(sin salida)

$ go test -count=1 -race ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.017s
ok  	seguimiento-medicion-gpc/internal/backlog	1.204s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.038s
ok  	seguimiento-medicion-gpc/internal/project	1.269s

$ go test -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 28.6% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 100.0% of statements

$ go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./cmd/server ./internal/httpapi
seguimiento-medicion-gpc/cmd/server: log net/http os seguimiento-medicion-gpc/internal/backlog seguimiento-medicion-gpc/internal/httpapi seguimiento-medicion-gpc/internal/project
seguimiento-medicion-gpc/internal/httpapi: encoding/json errors net/http seguimiento-medicion-gpc/internal/backlog seguimiento-medicion-gpc/internal/project strconv
```

`go list` muestra que el código nuevo solo importa la librería estándar y los paquetes internos (CA-TEC01.5).

### Cobertura por función

```
$ go tool cover -func=cover.out
seguimiento-medicion-gpc/cmd/server/main.go:15:			main			0.0%
seguimiento-medicion-gpc/cmd/server/main.go:26:			address			100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:63:		NewHandler		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:75:		createProject		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:90:		getProject		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:105:	updateProject		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:125:	createItem		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:145:	listItems		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:165:	pathID			100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:175:	statusFor		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:193:	writeError		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:202:	writeJSON		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:208:	toProjectResponse	100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:212:	toItemResponse		100.0%
```

`cmd/server` queda en 28.6% porque `main()` arma la aplicación y bloquea en `http.ListenAndServe`: no se prueba con un test unitario. Por eso la lógica que sí tiene decisiones (la elección del puerto) está separada en `address`, que tiene cobertura total, y el arranque se verificó con la prueba manual.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | ffc4999 | `test(TEC-01): agregar tests de los endpoints y del puerto (RED)` |
| GREEN | 07d50b5 | `feat(TEC-01): implementar el servidor HTTP con los endpoints de proyectos y backlog (GREEN)` |
| Test (ajuste) | e043603 | `test(TEC-01): cubrir el id no numérico en las demás rutas` |

No hubo commit de refactor (ver "Refactor"). El commit de evidencia no se lista en la tabla, porque un commit no puede contener su propio hash. No hay commit de cierre: `Closes #23` va en la descripción del Pull Request, y el Issue se cierra solo cuando el merge llega a `main`.
