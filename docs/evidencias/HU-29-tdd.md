# Evidencia TDD — HU-29 · Registrar un defecto

> Issue del repositorio: #47

- **Responsable:** Sosa Ricardo
- **Rama:** `feature/HU-29-registrar-defecto`
- **Fecha:** 25/10/2026 (commits del 24/10/2026)
- **Go:** go1.27.1 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` se verifica en el CI del Pull Request; ver la sección de `go test -race`)
- **Especificación SDD:** [`docs/especificaciones/HU-29-registrar-defecto.md`](../especificaciones/HU-29-registrar-defecto.md), SDD base redactada por Cardozo Leandro con asistencia de IA y aprobada por Peñalbé Hernán (Agile Enabler) el 24/10/2026 con el ajuste del sprint de detección faltante (400)
- **Escenarios BDD:** [`features/HU-29-registrar-defecto.feature`](../../features/HU-29-registrar-defecto.feature) (16 escenarios, 109 pasos, ejecutados con godog)

Todas las salidas de este documento se volvieron a correr el 25/10/2026 sobre cada commit, en una copia temporal (`git worktree`) que se borró al terminar. Se muestran las líneas de resultado y de fallo; se omite la salida de godog.

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**40 tests nuevos (funciones `Test`) cubriendo 16 escenarios BDD (109 pasos)**, en tres ciclos TDD y un commit de persistencia:

- **Ciclo 1 (dominio):** 17 tests nuevos.
  - 16 tests unitarios en `internal/defect/register_test.go`.
  - 1 test que ejecuta los escenarios BDD con godog (`internal/defect/register_features_test.go`, runner `TestFeatures_HU29RegistrarDefecto`, con 16 subtests).
- **Ciclo 2 (API):** 12 tests del endpoint en `internal/httpapi/defects_test.go`.
- **Ciclo 3 (interfaz web):** 4 tests en `internal/httpapi/web_test.go`.
- **Persistencia (PostgreSQL):** 7 tests nuevos.
  - 1 test de contrato de `MemoryRepository` en `internal/defect/repository_contract_test.go`.
  - 5 tests de PostgreSQL en `internal/defect/postgres_repository_test.go` (contrato incluido).
  - 1 test de wiring en `cmd/server/main_test.go` (`TestNewRepositories_PostgresUsaElRepositorioDeDefectos`).

La correspondencia no es 1 a 1: CA-29.11 no tiene escenarios godog y se cubre con los tests de `internal/httpapi`; algunos escenarios cubren dos criterios y varios escenarios comparten tests unitarios.

| Criterio | Escenarios BDD | Tests | Código |
|---|---|---|---|
| CA-29.1 Registro válido, ID correlativo, descripción recortada y severidad normalizada | Registrar un defecto válido en el sprint abierto actual (1; también cubre CA-29.3) | `TestRegister_Exitoso`, `TestMemoryRepository_Contrato`, `TestPostgresRepository_GuardaElDefecto` | `register.go`, `memory_repository.go`, `postgres_repository.go` |
| CA-29.2 Datos obligatorios | Rechazar descripción vacía (1); Rechazar sprint de detección ausente o 0 (compartido con CA-29.6) | `TestRegister_DescripcionObligatoria`, `TestRegister_SprintDeDeteccionObligatorio`, `TestRegister_OrdenDeValidacion` | `register.go` (`ErrEmptyDescription`, `ErrMissingDetectionSprint`) |
| CA-29.3 Estado inicial `ABIERTO` y sin sprint de resolución | Registrar un defecto válido en el sprint abierto actual (compartido con CA-29.1) | `TestRegister_Exitoso`, `TestPostgresRepository_SprintDeResolucionCeroSeGuarda` | `defect.go`, `register.go` |
| CA-29.4 Severidad CRITICA, ALTA, MEDIA o BAJA | Rechazar severidad inválida (1) | `TestRegister_SeveridadInvalida`, `TestInterfaz_ListaDeSeveridadesDeDefectosCoincideConElPaquete` | `register.go` (`ErrInvalidSeverity`), `defect.go` (`SeverityValues`) |
| CA-29.5 Historia existente, en cualquier estado | Registrar sobre una historia terminada; Rechazar historia inexistente (2) | `TestRegister_HistoriaTerminada`, `TestRegister_HistoriaInexistente`, `TestPostgresRepository_ItemInexistenteFalla` | `register.go`, clave foránea `item_id` de `defects` |
| CA-29.6 Sprint de detección obligatorio, existente, abierto y del mismo proyecto | Historia de otro sprint; sprint ausente o 0 (compartido con CA-29.2); sprint negativo; sprint cerrado; sprint positivo inexistente; sprint de otro proyecto (6) | `TestRegister_HistoriaDeOtroSprint`, `TestRegister_SprintDeDeteccionObligatorio`, `TestRegister_SprintDeDeteccionInexistente`, `TestRegister_SprintCerrado`, `TestRegister_SprintDeOtroProyecto`, `TestPostgresRepository_SprintDeDeteccionInexistenteFalla` | `register.go` (`ErrDetectionSprintFromAnotherProject`), `with_open.go` de HU-26 |
| CA-29.7 No modifica la historia ni el sprint | No modificar la historia ni el sprint (1) | `TestRegister_NoModificaHistoriaNiSprint` | `register.go` |
| CA-29.8 El defecto devuelto es una copia | Modificar el defecto devuelto no altera el guardado (1) | `TestRegister_CopiaIndependiente`, contratos de `Save` | `defect.go`, repositorios |
| CA-29.9 Concurrencia y atomicidad con el cierre | Registrar en simultáneo varios defectos; Registrar y cerrar el sprint en simultáneo (2) | `TestRegister_Concurrente`, `TestRegister_ConcurrenteConCierre` | `manager.go`, `register.go` (guardado dentro de `WithOpenSprint`) |
| CA-29.10 Error del repositorio propagado | Error del repositorio al guardar (1) | `TestRegister_ErrorDelRepositorio` | `register.go` |
| CA-29.11 API y formulario web | — | 12 tests `TestRegistrarDefecto_*`, 4 tests `TestInterfaz_*Defecto*` | `defects.go`, `server.go`, `index.html`, `app.js` |

Los escenarios por criterio suman 16 si cada escenario compartido se cuenta una sola vez (1 + 1 + 0 + 1 + 2 + 6 + 1 + 1 + 2 + 1 + 0 = 16): el escenario válido se cuenta en CA-29.1 y el de sprint ausente o 0, en CA-29.6.

## Fase RED — ciclo 1 (dominio)

Escenarios, pasos godog y tests escritos antes del código de producción (commit `c34d314`, que también incluye el `.feature`; la SDD está en `a436116`). Primero se agregaron firmas vacías para que el paquete compile: `Register` devuelve `Defect{}, errNotImplemented`. **El RED es el fallo de las aserciones, no de la compilación.**

Archivos con stubs (sin lógica) en el RED:

- `internal/defect/defect.go` (tipos `Defect`, `Status` y `Severity`, `SeverityValues` y los cuatro errores con su mensaje)
- `internal/defect/repository.go` (interfaz `Repository` con `Save`)
- `internal/defect/manager.go` (constructor, todavía sin mutex)
- `internal/defect/register.go` (`Register` sin validar)

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)
```

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	4.194s
ok  	seguimiento-medicion-gpc/internal/backlog	4.931s
16 scenarios (16 failed)
109 steps (67 passed, 16 failed, 26 skipped)
--- FAIL: TestFeatures_HU29RegistrarDefecto (0.15s)
    register_features_test.go:34: fallaron escenarios BDD de HU-29
--- FAIL: TestRegister_Exitoso (0.00s)
--- FAIL: TestRegister_HistoriaTerminada (0.00s)
--- FAIL: TestRegister_HistoriaDeOtroSprint (0.00s)
--- FAIL: TestRegister_DescripcionObligatoria (0.00s)
--- FAIL: TestRegister_SeveridadInvalida (0.00s)
--- FAIL: TestRegister_SprintDeDeteccionObligatorio (0.00s)
--- FAIL: TestRegister_SprintDeDeteccionInexistente (0.00s)
--- FAIL: TestRegister_SprintCerrado (0.00s)
--- FAIL: TestRegister_SprintDeOtroProyecto (0.00s)
--- FAIL: TestRegister_HistoriaInexistente (0.00s)
--- FAIL: TestRegister_OrdenDeValidacion (0.00s)
--- FAIL: TestRegister_NoModificaHistoriaNiSprint (0.00s)
--- FAIL: TestRegister_CopiaIndependiente (0.00s)
--- FAIL: TestRegister_ErrorDelRepositorio (0.00s)
--- FAIL: TestRegister_Concurrente (0.00s)
--- FAIL: TestRegister_ConcurrenteConCierre (0.00s)
FAIL
FAIL	seguimiento-medicion-gpc/internal/defect	2.643s
ok  	seguimiento-medicion-gpc/internal/effort	4.763s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.648s
ok  	seguimiento-medicion-gpc/internal/poker	3.766s
ok  	seguimiento-medicion-gpc/internal/postgres	1.518s
ok  	seguimiento-medicion-gpc/internal/postgrestest	2.602s
ok  	seguimiento-medicion-gpc/internal/project	4.536s
ok  	seguimiento-medicion-gpc/internal/sprint	3.884s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

Fragmentos de las aserciones que fallan:

```
    register_test.go:211: Register devolvió un error inesperado: no implementado
    register_test.go:247: Register sobre una historia TERMINADO devolvió un error: no implementado
    register_test.go:321: err = no implementado, se esperaba no existe un sprint con ese id
    register_test.go:332: err = no implementado, se esperaba el sprint está cerrado
    register_test.go:346: err = no implementado, se esperaba el sprint de detección pertenece a otro proyecto
        register_test.go:310: err = no implementado, se esperaba el sprint de detección es obligatorio
        suite.go:640: error = no implementado, se esperaba la severidad debe ser CRITICA, ALTA, MEDIA o BAJA
        suite.go:640: error = no implementado, se esperaba no existe un ítem del backlog con ese id
        suite.go:640: registro 0 falló: no implementado
```

**17 tests fallan y todos son de HU-29** (los 17 de `internal/defect`); las demás suites quedan en verde.

## Fase GREEN — ciclo 1 (dominio)

Código mínimo de producción (commit `9ec8c3e`), sin tocar tests:

- `internal/defect/manager.go`: agrega el mutex del `Manager`.
- `internal/defect/memory_repository.go`: `Save` asigna el ID correlativo global desde 1 y guarda una copia.
- `internal/defect/register.go`: recorta la descripción y normaliza la severidad; valida descripción, severidad y sprint de detección (en ese orden); dentro de `WithOpenSprint` consulta la historia, valida que sea del mismo proyecto que el sprint y guarda con estado `ABIERTO` y sprint de resolución 0.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	4.978s
ok  	seguimiento-medicion-gpc/internal/backlog	4.878s
ok  	seguimiento-medicion-gpc/internal/defect	4.546s
ok  	seguimiento-medicion-gpc/internal/effort	2.844s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.684s
ok  	seguimiento-medicion-gpc/internal/poker	2.467s
ok  	seguimiento-medicion-gpc/internal/postgres	1.371s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.568s
ok  	seguimiento-medicion-gpc/internal/project	4.719s
ok  	seguimiento-medicion-gpc/internal/sprint	3.876s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

## Fase RED — ciclo 2 (API)

Tests del endpoint escritos antes del handler (commit `f5d2dd6`). Para que compilen sin cablear todavía `NewHandler` ni `cmd/server`, el commit agrega en `server.go` el campo `defects` y una función interna `newHandler`, y en `internal/httpapi/export_test.go` un `NewHandlerWithDefects` que solo usan los tests. La ruta `POST /api/items/{id}/defectos` todavía no existe, así que el pedido responde 405.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	4.604s
ok  	seguimiento-medicion-gpc/internal/backlog	4.882s
ok  	seguimiento-medicion-gpc/internal/defect	3.524s
ok  	seguimiento-medicion-gpc/internal/effort	4.542s
--- FAIL: TestRegistrarDefecto_Devuelve201ConElDefecto (0.00s)
--- FAIL: TestRegistrarDefecto_IgnoraEstadoYSprintDeResolucion (0.00s)
--- FAIL: TestRegistrarDefecto_DatosInvalidosDevuelven400 (0.00s)
--- FAIL: TestRegistrarDefecto_JSONInvalidoDevuelve400 (0.00s)
--- FAIL: TestRegistrarDefecto_SprintDeDeteccionNoEnteroDevuelve400 (0.00s)
--- FAIL: TestRegistrarDefecto_TextoDespuesDelJSONDevuelve400 (0.00s)
--- FAIL: TestRegistrarDefecto_ItemIDInvalidoDevuelve400 (0.00s)
--- FAIL: TestRegistrarDefecto_InexistenteDevuelve404 (0.00s)
--- FAIL: TestRegistrarDefecto_SprintCerradoDevuelve409 (0.00s)
--- FAIL: TestRegistrarDefecto_SprintDeOtroProyectoDevuelve409 (0.00s)
--- FAIL: TestRegistrarDefecto_CuerpoMayorA1MiBDa413 (0.01s)
--- FAIL: TestRegistrarDefecto_ErrorDelRepositorioDevuelve500SinDetalle (0.00s)
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi	3.535s
ok  	seguimiento-medicion-gpc/internal/poker	5.680s
ok  	seguimiento-medicion-gpc/internal/postgres	1.305s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.322s
ok  	seguimiento-medicion-gpc/internal/project	4.275s
ok  	seguimiento-medicion-gpc/internal/sprint	4.180s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

Fragmentos de las aserciones que fallan:

```
    defects_test.go:149: código = 405, se esperaba 400; cuerpo: Method Not Allowed
    defects_test.go:181: código = 405, se esperaba 400; cuerpo: Method Not Allowed
    defects_test.go:218: código = 405, se esperaba 409; cuerpo: Method Not Allowed
    defects_test.go:238: código = 405, se esperaba 413; cuerpo: Method Not Allowed
        defects_test.go:207: código = 405, se esperaba 404; cuerpo: Method Not Allowed
```

**12 tests fallan y todos son de HU-29** (los 12 de `defects_test.go`).

## Fase GREEN — ciclo 2 (API)

Código del endpoint (commit `4ec6452`):

- `internal/httpapi/defects.go`: handler `registerDefect` (ID de la ruta, JSON y `Manager`, en ese orden) y la respuesta con `estado` y `sprint_resolucion_id`.
- `internal/httpapi/server.go`: ruta `POST /api/items/{id}/defectos`, `NewHandler` recibe el `defect.Manager` y `statusFor` traduce `ErrEmptyDescription`, `ErrInvalidSeverity` y `ErrMissingDetectionSprint` a 400 y `ErrDetectionSprintFromAnotherProject` a 409.
- `cmd/server/main.go`: arma el `defect.Manager` (en este ciclo, con `defect.NewMemoryRepository()`).

Este commit también toca archivos de test, solo por el cambio de firma de `NewHandler` (ver "Desviación del ciclo TDD").

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	2.871s
ok  	seguimiento-medicion-gpc/internal/backlog	4.207s
ok  	seguimiento-medicion-gpc/internal/defect	2.894s
ok  	seguimiento-medicion-gpc/internal/effort	3.747s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.689s
ok  	seguimiento-medicion-gpc/internal/poker	2.909s
ok  	seguimiento-medicion-gpc/internal/postgres	1.126s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.260s
ok  	seguimiento-medicion-gpc/internal/project	2.971s
ok  	seguimiento-medicion-gpc/internal/sprint	3.027s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

## Fase RED — ciclo 3 (interfaz web)

Tests de la interfaz escritos antes del formulario (commit `4c15f15`).

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	3.282s
ok  	seguimiento-medicion-gpc/internal/backlog	4.321s
ok  	seguimiento-medicion-gpc/internal/defect	3.152s
ok  	seguimiento-medicion-gpc/internal/effort	5.137s
--- FAIL: TestInterfaz_FormularioDeDefecto (0.00s)
--- FAIL: TestInterfaz_RegistrarDefectoMuestraElMensajeDeExito (0.00s)
--- FAIL: TestInterfaz_RegistrarDefectoMuestraElMensajeDeError (0.00s)
--- FAIL: TestInterfaz_ListaDeSeveridadesDeDefectosCoincideConElPaquete (0.00s)
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi	3.953s
ok  	seguimiento-medicion-gpc/internal/poker	5.070s
ok  	seguimiento-medicion-gpc/internal/postgres	1.335s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.405s
ok  	seguimiento-medicion-gpc/internal/project	4.275s
ok  	seguimiento-medicion-gpc/internal/sprint	3.884s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

```
    web_test.go:167: la página no contiene la sección de defectos
    web_test.go:202: app.js no contiene "mostrar(\"msg-defectos\", \"Defecto \""
    web_test.go:212: la página no tiene el mensaje de error de defectos
    web_test.go:217: app.js no muestra los errores de defectos con accion
    web_test.go:234: app.js no define const severidadesDefecto = ["CRITICA", "ALTA", "MEDIA", "BAJA"];
```

**4 tests fallan y todos son de HU-29.**

## Fase GREEN — ciclo 3 (interfaz web)

Código mínimo (commit `7d251e1`), sin tocar tests:

- `internal/web/index.html`: sección "Defectos" con el formulario "Registrar defecto" (ítem, descripción, severidad y sprint de detección con `required` y `min="1"`).
- `internal/web/app.js`: lista `severidadesDefecto`, llamada a `POST /items/{id}/defectos` y mensajes de éxito y de error.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	3.960s
ok  	seguimiento-medicion-gpc/internal/backlog	4.020s
ok  	seguimiento-medicion-gpc/internal/defect	2.237s
ok  	seguimiento-medicion-gpc/internal/effort	3.838s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.230s
ok  	seguimiento-medicion-gpc/internal/poker	3.837s
ok  	seguimiento-medicion-gpc/internal/postgres	1.264s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.285s
ok  	seguimiento-medicion-gpc/internal/project	3.815s
ok  	seguimiento-medicion-gpc/internal/sprint	3.765s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

## Persistencia (PostgreSQL)

Tests y código en un solo commit (`63a8c47`), sin un RED separado (ver "Desviación del ciclo TDD"):

- `internal/defect/postgres_repository.go`: `Save` con `INSERT … RETURNING` y `scanDefect`.
- `db/init/01-schema.sql`: tabla `defects` (`id` identity, `description`, `severity`, `status`, `item_id` con clave foránea a `backlog_items`, `detection_sprint_id` con clave foránea a `sprints`, `resolution_sprint_id` con valor por defecto 0) e índices `defects_item_id_idx`, `defects_detection_sprint_id_idx` y `defects_resolution_sprint_id_idx`.
- `internal/postgrestest/postgrestest.go`: `defects` incluida en el `TRUNCATE`.
- `cmd/server/repositories.go`: la rama PostgreSQL usa `defect.NewPostgresRepository(pool)` y la de memoria `defect.NewMemoryRepository()`.
- `cmd/server/main.go`: el `defect.Manager` pasa a usar `repos.defects`, así que con `STORAGE=postgres` el defecto se guarda en la base.
- Tests: `TestMemoryRepository_Contrato`, los 5 de `postgres_repository_test.go` y `TestNewRepositories_PostgresUsaElRepositorioDeDefectos`.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	3.930s
ok  	seguimiento-medicion-gpc/internal/backlog	4.681s
ok  	seguimiento-medicion-gpc/internal/defect	4.341s
ok  	seguimiento-medicion-gpc/internal/effort	4.813s
ok  	seguimiento-medicion-gpc/internal/httpapi	3.583s
ok  	seguimiento-medicion-gpc/internal/poker	3.636s
ok  	seguimiento-medicion-gpc/internal/postgres	1.347s
ok  	seguimiento-medicion-gpc/internal/postgrestest	2.881s
ok  	seguimiento-medicion-gpc/internal/project	3.431s
ok  	seguimiento-medicion-gpc/internal/sprint	4.510s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

**Los tests de PostgreSQL se omitieron localmente porque no hay `TEST_DATABASE_URL`.** Sin la variable, los 6 tests nuevos que necesitan la base hacen SKIP; `TestMemoryRepository_Contrato` pasa:

```
$ go test -count=1 -v ./... | grep -- '--- SKIP'      (solo las líneas de HU-29)
--- SKIP: TestNewRepositories_PostgresUsaElRepositorioDeDefectos (0.00s)
--- SKIP: TestPostgresRepository_Contrato (0.00s)
--- SKIP: TestPostgresRepository_GuardaElDefecto (0.00s)
--- SKIP: TestPostgresRepository_ItemInexistenteFalla (0.00s)
--- SKIP: TestPostgresRepository_SprintDeDeteccionInexistenteFalla (0.00s)
--- SKIP: TestPostgresRepository_SprintDeResolucionCeroSeGuarda (0.00s)
```

En toda la suite hacen SKIP 25 tests sin base (los 6 de HU-29 y los de PostgreSQL de las historias anteriores). La corrida contra la base queda pendiente de quien tenga Docker y `TEST_DATABASE_URL` (como en HU-26 y HU-21); este documento no incluye cifras con base.

## Desviación del ciclo TDD

1. **Tests tocados en el GREEN de la API (`4ec6452`).** El GREEN agrega el parámetro `defects` a `NewHandler`, así que actualiza las llamadas en `defects_test.go`, `efforts_test.go`, `items_test.go`, `poker_test.go` y `server_test.go` y borra `export_test.go` (el `NewHandlerWithDefects` del RED ya no hace falta). Se revisó el diff con `git show 4ec6452 -- 'internal/httpapi/*_test.go'`: además del `export_test.go` borrado, las únicas líneas que cambian son la llamada a `NewHandler`, la importación de `internal/defect` y la creación del `defect.Manager`. Ninguna aserción cambió.
2. **Persistencia sin RED separado (`63a8c47`).** A diferencia de HU-26 y HU-21, los tests de PostgreSQL y el repositorio entraron en el mismo commit, así que no hay una corrida que muestre esos tests fallando. Además, sin base hacen SKIP, por lo que localmente no se puede ver un RED ni un GREEN de esa parte.
3. **`bodyEndpoints` de `server_test.go`.** La SDD preveía sumar el endpoint a `bodyEndpoints` para los casos de 413 y texto después del JSON. No se agregó; esos dos casos se cubren con `TestRegistrarDefecto_CuerpoMayorA1MiBDa413` y `TestRegistrarDefecto_TextoDespuesDelJSONDevuelve400` en `defects_test.go`.

## Refactor

No se hizo. El paquete `internal/defect` sigue el patrón de `internal/effort` (manager con mutex, repositorios detrás de una interfaz, validaciones en el manager y guardado dentro de `WithOpenSprint`), y el handler reutiliza `pathID`, `decodeJSON`, `writeJSON`, `writeError` y `statusFor`.

## Pruebas de mutación

Cada mutación se aplicó a mano sobre `internal/defect/register.go`, se corrió `go test -count=1 ./...` y se restauró con `git checkout -- internal/defect/register.go`. Después de cada restauración, `git diff --stat` quedó vacío. Las mutaciones (a), (b), (c), (d) y (f) se repitieron con `go test -count=1 ./internal/defect/... ./internal/httpapi/...` y fallaron los mismos tests, salvo la diferencia de (d) que se explica debajo de la tabla.

| # | Mutación | Tests que fallan | Resultado |
|---|---|---|---|
| a | Quitar el chequeo `detectionSprintID <= 0` | `TestFeatures_HU29RegistrarDefecto` (2 escenarios: sprint de detección ausente o 0, y negativo), `TestRegister_SprintDeDeteccionObligatorio`, `TestRegister_OrdenDeValidacion`, `TestRegistrarDefecto_DatosInvalidosDevuelven400` (404 en lugar de 400) | Detectada |
| b | Quitar el `TrimSpace` de la descripción | `TestFeatures_HU29RegistrarDefecto` (1 escenario: defecto válido), `TestRegister_Exitoso`, `TestRegister_DescripcionObligatoria`, `TestRegister_OrdenDeValidacion`, `TestRegistrarDefecto_Devuelve201ConElDefecto`, `TestRegistrarDefecto_DatosInvalidosDevuelven400` | Detectada |
| c | Quitar el `ToUpper` de la severidad | `TestFeatures_HU29RegistrarDefecto` (1 escenario: defecto válido), `TestRegister_Exitoso`, `TestRegistrarDefecto_Devuelve201ConElDefecto`, `TestRegistrarDefecto_InexistenteDevuelve404`, `TestRegistrarDefecto_SprintCerradoDevuelve409`, `TestRegistrarDefecto_CuerpoMayorA1MiBDa413`, `TestRegistrarDefecto_ErrorDelRepositorioDevuelve500SinDetalle` | Detectada |
| d | `item.ProjectID != sp.ProjectID` → `==` | `TestFeatures_HU29RegistrarDefecto` (8 escenarios), 8 tests `TestRegister_*` (`Exitoso`, `HistoriaTerminada`, `HistoriaDeOtroSprint`, `SprintDeOtroProyecto`, `NoModificaHistoriaNiSprint`, `CopiaIndependiente`, `ErrorDelRepositorio`, `Concurrente`) y 5 tests `TestRegistrarDefecto_*` (`Devuelve201ConElDefecto`, `IgnoraEstadoYSprintDeResolucion`, `SprintDeOtroProyectoDevuelve409`, `CuerpoMayorA1MiBDa413`, `ErrorDelRepositorioDevuelve500SinDetalle`) | Detectada |
| e | Mover `m.repo.Save` fuera de `WithOpenSprint` | Ninguno en una corrida de `go test -count=1 ./...` | **Sobrevive en una corrida; detectada de forma intermitente** |
| f | `StatusOpen` → `StatusResolved` | `TestFeatures_HU29RegistrarDefecto` (2 escenarios: defecto válido y defecto devuelto como copia), `TestRegister_Exitoso`, `TestRegister_CopiaIndependiente`, `TestRegistrarDefecto_Devuelve201ConElDefecto`, `TestRegistrarDefecto_IgnoraEstadoYSprintDeResolucion` | Detectada |

En la mutación (d), `TestRegister_ConcurrenteConCierre` también falló en la primera corrida (`go test -count=1 ./...`), pero no en la repetición. Ese test acepta que `Register` pierda contra el cierre, así que solo falla cuando el registro gana la carrera. Por eso no figura en la tabla: la mutación la detectan de forma determinista los demás tests.

Mutación (e) aplicada:

```
-		saved, err = m.repo.Save(Defect{
-			...
-		})
-		return err
+		return nil
 	})
 	if err != nil {
 		return Defect{}, err
 	}
+	saved, err = m.repo.Save(Defect{Description: description, Severity: Severity(severity), Status: StatusOpen, ItemID: itemID, DetectionSprintID: detectionSprintID, ResolutionSprintID: 0})
+	if err != nil {
+		return Defect{}, err
+	}
```

Con esta mutación el guardado ya no es atómico con el cierre del sprint (RN-29.11, CA-29.9). Lo único que la detecta es `TestRegister_ConcurrenteConCierre`, que corre 20 intentos de registro y cierre simultáneos y falla solo si el cierre cae entre la validación y el `Save`. En una corrida normal pasó. Repetida con `-count=50`, falló en 2 de 4 lotes:

```
$ go test -count=50 -run 'TestRegister_ConcurrenteConCierre|TestFeatures' ./internal/defect/      (lote 1)
    register_test.go:543: se guardó un defecto después del cierre: [close defect]
--- FAIL: TestRegister_ConcurrenteConCierre (0.00s)
FAIL	seguimiento-medicion-gpc/internal/defect	4.904s
$ (lote 2)
ok  	seguimiento-medicion-gpc/internal/defect	10.178s
$ (lote 3)
ok  	seguimiento-medicion-gpc/internal/defect	6.114s
$ go test -count=50 -run 'Concurrente|Cierre|Cerrado|Closed' ./internal/httpapi/ ./internal/defect/      (lote 4)
--- FAIL: TestRegister_ConcurrenteConCierre (0.00s)
FAIL	seguimiento-medicion-gpc/internal/defect	3.517s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.677s
```

El escenario godog "Registrar y cerrar el sprint en simultáneo" no la detectó en ninguna de las 150 ejecuciones de los lotes 1 a 3. **La mutación (e) se detecta, pero no de forma determinista.** No se cambiaron los tests (regla del equipo); queda como observación para la revisión.

## Prueba de punta a punta de la API

Servidor real compilado desde `63a8c47` con `go build -o server.exe ./cmd/server` y arrancado en memoria con `PORT=18429 STORAGE=memory ./server.exe` (log: `servidor escuchando en :18429 (almacenamiento: memory)`). Pedidos con `curl` desde Git Bash. Los textos de entrada van sin tildes porque Git Bash no los envía en UTF-8.

```
$ curl -s -X POST localhost:18429/api/proyectos -d '{"nombre":"Tienda online","descripcion":"Proyecto de prueba HU-29"}'
{"id":1,"nombre":"Tienda online","descripcion":"Proyecto de prueba HU-29","estado":"PLANIFICADO"}
HTTP 201

$ curl -s -X POST localhost:18429/api/proyectos/1/backlog -d '{"titulo":"Aplicar cupon al total","descripcion":"Como cliente quiero ver el descuento","prioridad":"ALTA","criterios":["El total descuenta el cupon"]}'
{"id":1,"proyecto_id":1,"titulo":"Aplicar cupon al total","descripcion":"Como cliente quiero ver el descuento","prioridad":"ALTA","criterios":["El total descuenta el cupon"],"estado":"PENDIENTE","story_points":0,"horas_estimadas":0,"sprint_id":0}
HTTP 201

$ curl -s -X POST localhost:18429/api/proyectos/1/sprints -d '{"nombre":"Sprint 1","fecha_inicio":"2026-10-20","fecha_fin":"2026-11-03","objetivo":"Checkout"}'
{"id":1,"proyecto_id":1,"nombre":"Sprint 1","fecha_inicio":"2026-10-20","fecha_fin":"2026-11-03","objetivo":"Checkout","estado":"ABIERTO"}
HTTP 201

$ curl -s -X POST localhost:18429/api/items/1/defectos -d '{"descripcion":"  El total no descuenta el cupon  ","severidad":"alta","sprint_deteccion_id":1}'
{"id":1,"descripcion":"El total no descuenta el cupon","severidad":"ALTA","estado":"ABIERTO","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}
HTTP 201

$ curl -s -X POST localhost:18429/api/items/1/defectos -d '{"descripcion":"El total no descuenta el cupon","severidad":"ALTA","sprint_deteccion_id":99}'
{"error":"no existe un sprint con ese id"}
HTTP 404

$ curl -s -X POST localhost:18429/api/items/1/defectos -d '{"descripcion":"   ","severidad":"ALTA","sprint_deteccion_id":1}'
{"error":"la descripción del defecto es obligatoria"}
HTTP 400

$ curl -s -X POST localhost:18429/api/items/1/defectos -d '{"descripcion":"El total no descuenta el cupon","severidad":"ALTA"}'
{"error":"el sprint de detección es obligatorio"}
HTTP 400

$ curl -s -X POST localhost:18429/api/sprints/1/cerrar
{"id":1,"proyecto_id":1,"nombre":"Sprint 1","fecha_inicio":"2026-10-20","fecha_fin":"2026-11-03","objetivo":"Checkout","estado":"CERRADO"}
HTTP 200

$ curl -s -X POST localhost:18429/api/items/1/defectos -d '{"descripcion":"El total no descuenta el cupon","severidad":"ALTA","sprint_deteccion_id":1}'
{"error":"el sprint está cerrado"}
HTTP 409
```

(La línea `HTTP …` la imprime `-w '\nHTTP %{http_code}\n'`.) El defecto válido queda con la descripción recortada, la severidad en mayúsculas, estado `ABIERTO` y `sprint_resolucion_id` 0 (CA-29.1, CA-29.3 y CA-29.4); el sprint faltante da 400 con el mensaje del ajuste de la aprobación de la SDD, y después del cierre el registro da 409 (CA-29.6). El formulario web no se recorrió en el navegador; lo cubren los 4 tests de interfaz.

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
ok  	seguimiento-medicion-gpc/cmd/server	3.171s
ok  	seguimiento-medicion-gpc/internal/backlog	3.267s
ok  	seguimiento-medicion-gpc/internal/defect	3.666s
ok  	seguimiento-medicion-gpc/internal/effort	3.441s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.241s
ok  	seguimiento-medicion-gpc/internal/poker	3.769s
ok  	seguimiento-medicion-gpc/internal/postgres	0.891s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.972s
ok  	seguimiento-medicion-gpc/internal/project	3.001s
ok  	seguimiento-medicion-gpc/internal/sprint	2.681s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

```
$ go test -count=1 -v -run TestFeatures_HU29RegistrarDefecto ./internal/defect/
16 scenarios (16 passed)
109 steps (109 passed)
ok  	seguimiento-medicion-gpc/internal/defect	3.194s
```

### Cobertura

Medición local sin base y sin `-race`:

```
$ go test -count=1 -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	5.724s	coverage: 31.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	6.641s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/defect	6.128s	coverage: 76.6% of statements
ok  	seguimiento-medicion-gpc/internal/effort	4.879s	coverage: 76.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	3.339s	coverage: 98.7% of statements
ok  	seguimiento-medicion-gpc/internal/poker	4.603s	coverage: 54.4% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.662s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.557s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	3.186s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	3.256s	coverage: 79.1% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

```
$ go test -count=1 -coverprofile=cover.out ./...
$ go tool cover -func=cover.out
internal/defect/defect.go:29:			SeverityValues		0.0%
internal/defect/manager.go:20:			NewManager		100.0%
internal/defect/memory_repository.go:13:	NewMemoryRepository	100.0%
internal/defect/memory_repository.go:20:	Save			100.0%
internal/defect/postgres_repository.go:20:	NewPostgresRepository	0.0%
internal/defect/postgres_repository.go:29:	Save			0.0%
internal/defect/postgres_repository.go:43:	scanDefect		0.0%
internal/defect/register.go:12:			Register		100.0%
internal/defect/register.go:55:			isValidSeverity		100.0%
internal/httpapi/defects.go:29:			registerDefect		100.0%
internal/httpapi/defects.go:50:			toDefectResponse	100.0%
total:						(statements)		79.0%
```

`internal/defect/postgres_repository.go` queda en 0.0% porque sus tests hacen SKIP sin `TEST_DATABASE_URL`. `SeverityValues` figura en 0.0% en el perfil de `internal/defect` porque la usa el test de interfaz de `internal/httpapi`, que se mide en su propio paquete. `main()` queda en 0.0% porque bloquea en `ListenAndServe` y `newRepositories` en 83.3% por la rama de PostgreSQL.

Comparación con `main` (`86c0c5f`) medida con el mismo go1.27.1 en una copia temporal (`git worktree`, borrada al terminar):

```
$ go test -count=1 -cover ./...      (sobre 86c0c5f)
ok  	seguimiento-medicion-gpc/cmd/server	9.539s	coverage: 33.3% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	10.943s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/effort	7.807s	coverage: 76.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	8.096s	coverage: 98.6% of statements
ok  	seguimiento-medicion-gpc/internal/poker	6.447s	coverage: 54.4% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	4.008s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	3.756s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	7.220s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	5.279s	coverage: 79.1% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
$ go tool cover -func=cover-main.out | tail -n 1
total:						(statements)		79.1%
```

`internal/defect` entra en 76.6% (nuevo), `internal/httpapi` sube de 98.6% a 98.7% porque el handler nuevo queda al 100.0%, y `cmd/server` baja de 33.3% a 31.5% por el wiring del `defect.Manager` en `main()`, que no se ejecuta en los tests. El total sin base baja de **79.1%** a **79.0%**. Los demás paquetes conservan sus porcentajes.

### go test -race (CA-29.9)

En este entorno (Windows, sin cgo) `-race` no está disponible. Los tests de concurrencia (`TestRegister_Concurrente` y `TestRegister_ConcurrenteConCierre`) pasan sin `-race` localmente. **`go test -race` lo verifica el CI del Pull Request.**

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | a436116 | `docs(HU-29): agregar especificación SDD` |
| RED ciclo 1 (dominio, incluye BDD) | c34d314 | `test(HU-29): agregar tests (RED)` |
| GREEN ciclo 1 (dominio) | 9ec8c3e | `feat(HU-29): implementar registro de defectos (GREEN)` |
| RED ciclo 2 (API) | f5d2dd6 | `test(HU-29): agregar tests de API (RED)` |
| GREEN ciclo 2 (API) | 4ec6452 | `feat(HU-29): exponer el registro de defectos por API (GREEN)` |
| RED ciclo 3 (interfaz) | 4c15f15 | `test(HU-29): agregar tests de interfaz (RED)` |
| GREEN ciclo 3 (interfaz) | 7d251e1 | `feat(HU-29): agregar el formulario de registro de defectos (GREEN)` |
| Persistencia (PostgreSQL) | 63a8c47 | `feat(HU-29): persistir los defectos en PostgreSQL` |

No hubo commit de refactor porque no se hizo refactor.

HU-29: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
