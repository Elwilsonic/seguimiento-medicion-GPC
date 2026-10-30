# Evidencia TDD · HU-40 Exportar un reporte a un archivo PDF

> Issue del repositorio: #58

- **Especificación SDD:** [`docs/especificaciones/HU-40-exportar-reporte-pdf.md`](../especificaciones/HU-40-exportar-reporte-pdf.md), aprobada por Peñalbé Hernán el 29/10/2026.
- **Escenarios BDD:** [`features/HU-40-exportar-reporte-pdf.feature`](../../features/HU-40-exportar-reporte-pdf.feature): 8 definiciones (7 escenarios y 1 esquema con 3 ejemplos), que godog ejecuta como 10 escenarios y 39 pasos.
- **Rama:** `feature/HU-40-exportar-pdf`.
- **Story Points:** 5 (Planning del Sprint 4).
- **Dependencias:** HU-38 (reporte del proyecto, PR #148) y la librería de PDF aprobada (tema 3 del acta de Planning del Sprint 4), registrada en la SDD. HU-41 (fecha de fin) ya estaba en `main` al empezar.
- **Librería:** `github.com/go-pdf/fpdf` v0.9.0 (MIT, Go puro). Solo se usa en `internal/report/pdf.go`; en `go.mod` se agrega solo esa línea.

## Trazabilidad

**20 tests nuevos (funciones `Test`) cubriendo 10 escenarios BDD ejecutados (8 definiciones):** 12 unitarios en `internal/report/pdf_test.go`, el runner de godog en `internal/report/pdf_features_test.go`, 5 de la API en `internal/httpapi/report_pdf_test.go` y 2 de la interfaz en `internal/httpapi/web_test.go`, más 1 caso nuevo en la tabla de `TestRutasConID_IDNoNumericoDevuelve400`. La correspondencia no es 1 a 1: CA-40.1 en la API y la interfaz se cubre solo con tests de Go, CA-40.4 con la SDD y el README, y varios tests comparten criterio.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-40.1 PDF válido, descargable desde la API y el botón | "Exportar el reporte del ejemplo de TEC-03" | `TestWriteProjectPDF_GeneraUnPDFConCadaLineaDelReporte`, `TestReportePDF_Devuelve200ConElArchivo`, `TestReportePDF_IDNoNumericoDevuelve400`, caso "PDF del reporte del proyecto" de `TestRutasConID_IDNoNumericoDevuelve400`, `TestInterfaz_SeccionReportesConElBotonDescargarPDF`, `TestInterfaz_DescargarPDFLlamaALaAPIConElProyectoActual` |
| CA-40.2 Error al escribir | "Si no se puede escribir el archivo, se devuelve error" | `TestWriteProjectPDF_ErrorAlEscribirDevuelveError`, `TestReportePDF_ErrorDelRepositorioDevuelve500SinDetalle` |
| CA-40.3 Mismo contenido que el reporte | "Exportar el reporte del ejemplo de TEC-03", "Las horas con decimales se redondean como en la API" | `TestProjectReportLines_EjemploDeTEC03`, `TestProjectReportLines_ConFechaDeFin`, `TestProjectReportLines_RedondeaAUnDecimal`, `TestProjectReportLines_HorasDeCadaEsfuerzoTalCual`, `TestProjectReportLines_DefectoSinSprintDeResolucion`, `TestWriteProjectPDF_GeneraUnPDFConCadaLineaDelReporte`, `TestWriteProjectPDF_TextoLargoYVariasPaginas`, `TestReportePDF_TieneLosMismosValoresQueElReporteJSON` |
| CA-40.4 Librería aprobada y documentada | — | No es comportamiento: la verificación y la aprobación están en la SDD y los motivos en el README ("Librería de PDF") |
| CA-40.5 Proyecto sin datos | "Exportar el reporte de un proyecto sin datos" | `TestProjectReportLines_ProyectoSinDatos`, `TestWriteProjectPDF_ProyectoSinDatos` |
| CA-40.6 Proyecto inexistente | "Exportar el reporte de un proyecto inexistente falla" (3) | `TestReportePDF_InexistenteDevuelve404` |
| CA-40.7 Tildes y caracteres que no se pueden mostrar | "Las tildes, la ñ y los signos se ven en el PDF", "Un carácter que no se puede mostrar no impide generar el PDF" | `TestWriteProjectPDF_TildesYCaracteresFueraDeWindows1252` |
| CA-40.8 No modifica nada | "Exportar el reporte no modifica nada" | `TestWriteProjectPDF_NoModificaElReporte` |

## Fase RED (`c5e8c98`)

El RED agrega el `.feature`, los tests y las firmas vacías en `internal/report/pdf.go`: el tipo `pdfLine`, `projectReportLines`, que devuelve `nil`, y `WriteProjectPDF`, que devuelve `errNotImplemented`. Se eligió un error para que también fallen los casos "sin datos" y "no modifica". La librería no entra en el RED: se agrega en el GREEN con el código que la usa.

Sobre `c5e8c98`, en Linux con go1.27.1, sin base:

```
$ go vet ./...
(sin salida)

$ go test ./...
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
--- FAIL: TestReportePDF_Devuelve200ConElArchivo (0.00s)
--- FAIL: TestReportePDF_TieneLosMismosValoresQueElReporteJSON (0.00s)
--- FAIL: TestReportePDF_InexistenteDevuelve404 (0.00s)
    report_pdf_test.go:129: GET /api/proyectos/99/reporte/pdf: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    report_pdf_test.go:131: cuerpo no es JSON válido (invalid character 'p' after top-level value): 404 page not found
--- FAIL: TestReportePDF_IDNoNumericoDevuelve400 (0.00s)
--- FAIL: TestReportePDF_ErrorDelRepositorioDevuelve500SinDetalle (0.00s)
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.00s)
        server_test.go:249: código = 404, se esperaba 400; cuerpo: 404 page not found
--- FAIL: TestInterfaz_SeccionReportesConElBotonDescargarPDF (0.00s)
--- FAIL: TestInterfaz_DescargarPDFLlamaALaAPIConElProyectoActual (0.00s)
FAIL seguimiento-medicion-gpc/internal/httpapi
ok   seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
--- FAIL: TestFeatures_HU40ExportarReportePDF (0.01s)
    pdf_features_test.go:33: fallaron escenarios BDD de HU-40
--- FAIL: TestProjectReportLines_EjemploDeTEC03 (0.00s)
    pdf_test.go:148: líneas =
--- FAIL: TestProjectReportLines_ProyectoSinDatos (0.00s)
--- FAIL: TestProjectReportLines_ConFechaDeFin (0.00s)
--- FAIL: TestProjectReportLines_RedondeaAUnDecimal (0.00s)
--- FAIL: TestProjectReportLines_HorasDeCadaEsfuerzoTalCual (0.00s)
--- FAIL: TestProjectReportLines_DefectoSinSprintDeResolucion (0.00s)
--- FAIL: TestWriteProjectPDF_GeneraUnPDFConCadaLineaDelReporte (0.00s)
--- FAIL: TestWriteProjectPDF_ProyectoSinDatos (0.00s)
--- FAIL: TestWriteProjectPDF_TildesYCaracteresFueraDeWindows1252 (0.00s)
--- FAIL: TestWriteProjectPDF_TextoLargoYVariasPaginas (0.00s)
--- FAIL: TestWriteProjectPDF_ErrorAlEscribirDevuelveError (0.00s)
    pdf_test.go:328: err = HU-40 todavía no está implementada, se esperaba disco lleno
--- FAIL: TestWriteProjectPDF_NoModificaElReporte (0.00s)
FAIL seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
```

godog de HU-40 en el RED: `10 scenarios (3 passed, 7 failed)`, `39 steps (25 passed, 7 failed, 7 skipped)`. Los 3 escenarios que pasan son los del proyecto inexistente (IDs 99, 0 y -1): prueban que el reporte de HU-38 valida el proyecto antes de escribir el PDF, comportamiento que ya existía. En la API, los 400 y 404 fallan en el RED porque la ruta todavía no existe (`404 page not found` en texto, no JSON). Todos los fallos son de HU-40: los demás paquetes y los tests de HU-38 y HU-39 pasan.

## Corrección de dos tests del RED (`a317814`)

Al correr el GREEN, dos tests de la API seguían fallando por errores de los propios tests, no del código:

1. `TestReportePDF_Devuelve200ConElArchivo` y `TestReportePDF_TieneLosMismosValoresQueElReporteJSON` pedían el PDF con el helper `do` de `server_test.go`, que exige una respuesta `application/json`; el PDF responde `application/pdf`, como dice la SDD. Se cambió a `get`, el helper que ya usan los tests de la interfaz para respuestas que no son JSON. El código de respuesta se sigue verificando con `expectStatus`.
2. `TestReportePDF_Devuelve200ConElArchivo` buscaba entera la línea del defecto del fixture `reportTEC03` de HU-38, que es más larga que el ancho de la página y sigue en la línea siguiente (salidas esperadas de la SDD). Ahora verifica su comienzo; la línea completa la verifican `TestProjectReportLines_EjemploDeTEC03` y la API con `pdftotext` (más abajo).

No se aflojó ninguna otra aserción. La corrección va en un commit propio, antes del GREEN:

```
$ git diff --stat c5e8c98 a317814
 internal/httpapi/report_pdf_test.go | 10 +++++++---
```

## Fase GREEN (`69fe2f5`)

- `internal/report/pdf.go`: `projectReportLines` arma las líneas con los textos de la interfaz de HU-38, redondea a un decimal como la API (`oneDecimal`), escribe "sin datos" o "sin registrar" donde corresponde y las horas de cada esfuerzo tal como están. `WriteProjectPDF` escribe un PDF A4 sin comprimir con Helvetica, títulos en negrita y `MultiCell`, pasando el texto a Windows-1252 con el traductor de la librería.
- `internal/httpapi/report_pdf.go` y la ruta en `server.go`: `GET /api/proyectos/{id}/reporte/pdf` arma el reporte con `report.Manager.Project`, genera el PDF en memoria y recién entonces responde 200 con `application/pdf` y `Content-Disposition: attachment; filename="reporte-proyecto-{id}.pdf"`. Los errores responden en JSON con `writeError`.
- `internal/web`: botón "Descargar PDF" en "Reportes"; descarga el archivo con `fetch` y un enlace temporal, y muestra el error de la API en `msg-reporte`.
- `go.mod` y `go.sum`: `github.com/go-pdf/fpdf v0.9.0`. `go mod verify`: `all modules verified`.
- `README.md`: endpoint, `curl` de ejemplo, botón y sección "Librería de PDF" (CA-40.4). `internal/report/doc.go`: mención de HU-40.

El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat a317814 69fe2f5 -- '*_test.go' '*.feature'
(sin salida)
```

**Orden de las fuentes.** En la primera corrida del GREEN, el escenario "Exportar el reporte no modifica nada" (CA-40.8) falló 1 vez de cada 6: la librería escribe sus fuentes en un orden que cambia entre exportaciones, así que el mismo reporte daba archivos distintos. Se resolvió en el código, sin tocar el test, con `pdf.SetCatalogSort(true)`, la opción de la librería que ordena esos recursos. Después: 3000 pares de exportaciones del mismo reporte idénticos (sin las fechas de los metadatos) y 15 corridas seguidas del BDD en verde.

## Refactor

No hubo. `pdf.go` tiene funciones cortas; la única repetición, el redondeo a un decimal de `internal/httpapi`, se decidió en la SDD (límite conocido).

## Desviaciones del ciclo TDD

1. **Dos tests del RED corregidos antes del GREEN** (`a317814`), explicados arriba. Eran errores de los tests (un helper que exige JSON y una línea que el PDF parte en dos), no cambios de comportamiento.
2. **Rama sin test en el handler.** `projectReportPDF` queda en 87.5%: no se cubre el 500 cuando falla la generación del PDF en memoria. Escribir en un `bytes.Buffer` no falla y la librería no expone otra forma de provocar el error, así que no se puede escribir un test que lo pida. La rama sigue la regla RN-40.7 de la SDD (nunca responder un PDF a medias). El error de escritura sí está cubierto en `WriteProjectPDF` (`TestWriteProjectPDF_ErrorAlEscribirDevuelveError`).

## Prueba de punta a punta

Con el servidor real (`go build ./cmd/server`, puerto 8091) y datos cargados con `curl` (proyecto "Gestión de ventas «Sur»" con fecha de inicio y de fin, un sprint, un ítem, un esfuerzo de 1.5 h de "Ñandú" y un defecto "Falla (menor)"):

```
$ curl -D - -o r.pdf localhost:8091/api/proyectos/1/reporte/pdf
HTTP/1.1 200 OK
Content-Disposition: attachment; filename="reporte-proyecto-1.pdf"
Content-Type: application/pdf

$ file r.pdf
r.pdf: PDF document, version 1.3, 1 page(s)

$ pdftotext -layout r.pdf -
Reporte del proyecto «Gestión de ventas «Sur»»
Proyecto
Nombre: Gestión de ventas «Sur»
Descripción: Año 2026 (piloto)
Estado: PLANIFICADO
Inicio: 2026-10-05
Fin: 2026-12-18
Métricas
Story points: 0 planificados, 0 completados, 1 ítem(s) sin estimar.
Velocidad: sin datos (0 sprint(s) cerrado(s)).
Historias completadas: 0.0 % (0 de 1).
Horas: 0.0 estimadas, 1.5 reales, 1 ítem(s) sin horas estimadas.
Desviación: sin datos.
Defectos: 1 detectados, 0 resuelto(s).
Esfuerzos
#1 · 2026-10-06 · Ñandú · Desarrollo · ítem 1 · sprint 1 · 1.5 h
Total: 1.5 horas
Defectos
#1 · BAJA · ABIERTO · ítem 1 · detectado en sprint 1 · sin sprint de resolución · Falla (menor)
Cantidad: 1 defecto(s)
Integrantes
Integrantes que registraron esfuerzo: Ñandú

$ curl localhost:8091/api/proyectos/99/reporte/pdf
{"error":"no existe un proyecto con ese id"}          (404)
$ curl localhost:8091/api/proyectos/abc/reporte/pdf
{"error":"el id debe ser un número entero"}           (400)
```

En el navegador (Chromium con Playwright), en la sección "Reportes":

- Sin proyecto actual, "Descargar PDF" muestra "Elegí un proyecto: creá uno o escribí su ID en «Proyecto actual».".
- Con el proyecto 1, descarga `reporte-proyecto-1.pdf` y muestra "PDF del proyecto 1 descargado."; el texto del archivo descargado (`pdftotext`) es igual al del PDF de la API.
- Con el proyecto 99, muestra "no existe un proyecto con ese id".

## Validación final

Sobre `69fe2f5`, en Linux con go1.27.1:

```
$ gofmt -w internal/report/pdf.go internal/report/pdf_test.go internal/report/pdf_features_test.go internal/report/doc.go internal/httpapi/report_pdf.go internal/httpapi/report_pdf_test.go internal/httpapi/server.go internal/httpapi/server_test.go internal/httpapi/web_test.go
(sin cambios)

$ go vet ./...
(sin salida)

$ go test -count=1 ./...
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
ok   seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
?    seguimiento-medicion-gpc/internal/web [no test files]

$ go test -count=1 -cover ./...
ok   seguimiento-medicion-gpc/cmd/server             coverage: 31.5% of statements
ok   seguimiento-medicion-gpc/internal/backlog       coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/dashboard     coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/defect        coverage: 77.7% of statements
ok   seguimiento-medicion-gpc/internal/effort        coverage: 75.2% of statements
ok   seguimiento-medicion-gpc/internal/httpapi       coverage: 98.5% of statements
ok   seguimiento-medicion-gpc/internal/metrics       coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/poker         coverage: 69.5% of statements
ok   seguimiento-medicion-gpc/internal/postgres      coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/postgrestest  coverage: 36.8% of statements
ok   seguimiento-medicion-gpc/internal/project       coverage: 71.1% of statements
ok   seguimiento-medicion-gpc/internal/report        coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/sprint        coverage: 76.3% of statements
     seguimiento-medicion-gpc/internal/web           coverage: 0.0% of statements
```

godog de HU-40: `10 scenarios (10 passed)`, `39 steps (39 passed)`.

`go tool cover -func`: `projectReportLines`, `dateOrNotRegistered`, `effortText`, `defectText`, `oneDecimal` y `WriteProjectPDF` en 100.0%; `projectReportPDF` en 87.5% (desviación 2). Total del perfil sin base: **84.5%** (`main` en `f9af5a9`, con HU-41, mismo Go y comando, en una copia temporal con `git worktree`: **83.9%**). `internal/report` sigue en 100.0%; `internal/httpapi` baja de 98.8% a 98.5% por la rama sin test.

### Con base (PostgreSQL 16.15) y `-race`

Con PostgreSQL 16.15 instalado en el entorno de trabajo, no con el contenedor `postgres:17-alpine` de `docker-compose.yml` (no hay Docker en ese entorno; igual que en HU-47). La base `gpc_test` se recreó con el mismo `db/init/01-schema.sql` que usa el contenedor, con la columna `end_date` de HU-41:

```
$ TEST_DATABASE_URL=... go test -race -count=1 -p 1 -coverpkg=./... -coverprofile=cover.out ./...
ok   (los 13 paquetes)
$ go tool cover -func=cover.out | tail -n 1
total:                                  (statements)            96.0%
```

Sin SKIP (0 tests omitidos con `go test -v`) y sin carreras. HU-40 no toca la persistencia: lee el reporte con la operación de HU-38.

## Cambios por la revisión del PR #151

Sosa Ricardo aprobó el PR con tres detalles opcionales:

1. **Criterios del Issue #58 en borrador.** Se reemplazan por CA-40.1 a CA-40.8 después del merge, con autorización.
2. **Entorno de PostgreSQL.** Se aclaró en "Con base" que se midió con PostgreSQL 16.15 instalado en el entorno de trabajo, no con el contenedor de `docker-compose.yml`.
3. **Liberar el enlace de la descarga después del clic.** `URL.revokeObjectURL(url)` justo después de `enlace.click()` puede cancelar la descarga en algunos navegadores. Con TDD:
   - **RED (`584b907`):** `TestInterfaz_DescargarPDFLlamaALaAPIConElProyectoActual` suma que `app.js` libere el enlace con `setTimeout(() => URL.revokeObjectURL(url), 0)`. Falla por esa aserción: `web_test.go:1161: app.js no contiene "setTimeout(() => URL.revokeObjectURL(url), 0)"`.
   - **GREEN (`52a0498`):** `app.js` libera el enlace con `setTimeout(..., 0)`. `go test ./...` en verde. En el navegador (Chromium con Playwright), "Descargar PDF" sigue descargando `reporte-proyecto-1.pdf` y mostrando "PDF del proyecto 1 descargado.".

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | add478a | `docs(HU-40): agregar especificación SDD` |
| RED | c5e8c98 | `test(HU-40): agregar tests (RED)` |
| Corrección de tests del RED | a317814 | `test(HU-40): corregir dos tests de la API del RED` |
| GREEN | 69fe2f5 | `feat(HU-40): exportar el reporte del proyecto a PDF (GREEN)` |
| Refactor | — | No hubo |
| Evidencia y cobertura | 6940cd5 | `docs(HU-40): agregar evidencia TDD y cobertura` |
| RED de la revisión (punto 3) | 584b907 | `test(HU-40): verificar que la descarga libera el enlace después del clic (RED)` |
| GREEN de la revisión (punto 3) | 52a0498 | `fix(HU-40): liberar el enlace de la descarga después del clic (GREEN)` |

HU-40: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
