# Evidencia TDD · HU-18

> Issue del repositorio: #36

## Historia

Como Agile Enabler, quiero registrar la fecha de inicio de un proyecto, para saber desde cuándo se cuenta su duración.

## Resultado

HU-18 quedó implementada en la rama `feature/HU-18-fecha-inicio-proyecto`. `Project` suma el campo `StartDate` (el valor cero es "sin registrar") y `Manager.SetStartDate` registra o reemplaza la fecha de un proyecto existente bajo el mismo mutex que `Update`, sin cambiar ningún otro dato. La API lo expone en `PATCH /api/proyectos/{id}/fecha-inicio` con `{"fecha_inicio":"AAAA-MM-DD"}`, y `POST`, `GET` y `PUT` de proyectos devuelven `fecha_inicio` (`""` si no está registrada). En PostgreSQL, la columna `start_date TIMESTAMPTZ` (nullable) guarda `NULL` para "sin registrar" y se lee en UTC. La interfaz web suma en la sección Proyectos el formulario "Fecha de inicio" con el botón "Registrar fecha de inicio", y los datos del proyecto actual muestran la fecha.

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test |
|---|---|
| CA-18.1 — se registra la fecha de un proyecto existente sin cambiar sus demás datos, en cualquier estado y con fechas pasadas o futuras | “Registrar la fecha de inicio de un proyecto sin fecha”; “Registrar la fecha en un proyecto finalizado”; “Registrar una fecha futura”; `TestSetStartDate_RegistraLaFechaYConservaLosDemasDatos`; `TestSetStartDate_SeRegistraEnCualquierEstado`; `TestSetStartDate_AceptaFechasPasadasFuturasYDeAnioBisiesto`; `TestSetStartDate_NoModificaOtrosProyectos`; `TestRegistrarFechaInicioProyecto_Devuelve200ConElProyecto` |
| CA-18.2 — en la API, formato `AAAA-MM-DD` y fecha existente | `TestRegistrarFechaInicioProyecto_FormatoOFechaInvalidaDevuelve400` (6 casos: `2026-02-30`, `2026-13-01`, `23/10/2026`, `2026-1-5`, `2026-10-23T00:00:00Z` y `" 2026-10-23 "`) |
| CA-18.3 — la fecha es obligatoria | “Rechazar una fecha ausente”; `TestSetStartDate_FechaAusenteDevuelveErrMissingStartDateYNoCambia`; `TestSetStartDate_FechaAusenteSeInformaSinConsultarElRepositorio`; `TestRegistrarFechaInicioProyecto_FechaAusenteDevuelve400` (vacía, `null` y ausente); `TestRegistrarFechaInicioProyecto_FechaVaciaEnProyectoInexistenteDevuelve400` |
| CA-18.4 — proyecto inexistente | “Rechazar un proyecto inexistente” (esquema con 999, 0 y -1); `TestSetStartDate_ProyectoInexistenteDevuelveErrProjectNotFound`; `TestSetStartDate_SinProyectosCreadosDevuelveErrProjectNotFound`; `TestRegistrarFechaInicioProyecto_InexistenteDevuelve404` |
| CA-18.5 — volver a registrar reemplaza la fecha | “Volver a registrar la fecha”; `TestSetStartDate_VolverARegistrarReemplazaLaFecha`; `TestSetStartDate_MismaFechaEsValida`; `TestRegistrarFechaInicioProyecto_VolverARegistrarReemplazaLaFecha` |
| CA-18.6 — proyecto nuevo sin fecha; la consulta la muestra y modificar el proyecto la conserva | “Un proyecto recién creado no tiene fecha de inicio”; “Modificar el proyecto conserva la fecha de inicio”; `TestCreate_ProyectoNuevoNoTieneFechaDeInicio`; `TestGet_DevuelveLaFechaDeInicioRegistrada`; `TestUpdate_ConservaLaFechaDeInicioRegistrada`; `TestProyecto_RespuestasIncluyenLaFechaDeInicio`; `TestModificarProyecto_ConservaLaFechaDeInicio` |
| CA-18.7 — copia por valor | “Modificar el proyecto devuelto no altera el guardado”; `TestSetStartDate_ModificarElResultadoDevueltoNoAlteraElGuardado` |
| CA-18.8 — concurrencia y falla del repositorio | “Error del repositorio al guardar”; `TestSetStartDate_ErrorDelRepositorioSeDevuelveYElProyectoNoCambia`; `TestSetStartDate_RegistrosModificacionesYConsultasSimultaneosSinAccesosSinProteger`; `TestRegistrarFechaInicioProyecto_ErrorInternoDevuelve500` |
| CA-18.9 — memoria y PostgreSQL: `NULL` para "sin registrar" y lectura en UTC | Subtests de contrato “Save y GetByID conservan la fecha de inicio” y “Update escribe y conserva la fecha de inicio”, ejecutados por `TestMemoryRepository_Contrato` y `TestPostgresRepository_Contrato`; `TestPostgresRepository_SinFechaDeInicioSeGuardaComoNULL`; `TestPostgresRepository_FechaDeInicioEnUTC` |
| CA-18.10 — API e interfaz web | `TestRegistrarFechaInicioProyecto_JSONInvalidoDevuelve400` y los demás `TestRegistrarFechaInicioProyecto_*`; caso “registrar fecha de inicio” de `TestRutasConID_IDNoNumericoDevuelve400`; endpoint “registrar fecha de inicio” de `bodyEndpoints` (413 en `TestHandler_CuerpoMayorA1MiBDa413` y texto después del JSON en `TestHandler_TextoDespuesDelJSONDa400`); `TestInterfaz_FormularioDeFechaDeInicioDelProyecto`; `TestInterfaz_RegistrarFechaDeInicioMuestraElMensajeDeExito`; `TestInterfaz_DatosDelProyectoMuestranLaFechaDeInicio` |

La validación contiene **10 escenarios BDD** (9 escenarios y 1 esquema con 3 ejemplos, que godog ejecuta como 12) y **37 casos de prueba específicos**: 32 funciones de test nuevas (15 unitarias de `internal/project/start_date_test.go`, 1 runner de godog, 1 en `update_test.go`, 2 de PostgreSQL, 10 HTTP y 3 de interfaz web), 2 subtests de contrato de repositorio y 3 casos de tabla en tests HTTP existentes (`TestRutasConID_IDNoNumericoDevuelve400` y el endpoint de `bodyEndpoints`, que ejecutan dos tests). La correspondencia no es 1 a 1: CA-18.2, CA-18.9 y CA-18.10 no tienen escenario BDD, porque el formato de la fecha es de la capa HTTP, la persistencia se prueba con el contrato y PostgreSQL, y la interfaz con los tests de `internal/httpapi`.

Conteos con `grep` sobre los archivos de la rama (contra `main`, `850edc0`, para los archivos que ya existían):

```text
$ grep -c '^\s*Escenario:' features/HU-18-fecha-inicio-proyecto.feature
9
$ grep -c '^\s*Esquema del escenario:' features/HU-18-fecha-inicio-proyecto.feature
1
$ grep -c '^func Test' internal/project/start_date_test.go
15
$ grep -c '^func Test' internal/project/start_date_features_test.go
1
$ git diff 850edc0 HEAD -- internal/project/update_test.go | grep -c '^+func Test'
1
$ git diff 850edc0 HEAD -- internal/project/postgres_repository_test.go | grep -c '^+func Test'
2
$ git diff 850edc0 HEAD -- internal/httpapi/server_test.go | grep -c '^+func Test'
10
$ git diff 850edc0 HEAD -- internal/httpapi/web_test.go | grep -c '^+func Test'
3
$ git diff 850edc0 HEAD -- internal/project/repository_contract_test.go | grep -cE '^\+\s+t\.Run\("'
2
```

Escenarios por criterio (`grep` del CA en el nombre de cada escenario): CA-18.1: 3; CA-18.3: 1; CA-18.4: 1 (esquema); CA-18.5: 1; CA-18.6: 2; CA-18.7: 1; CA-18.8: 1; CA-18.2, CA-18.9 y CA-18.10: 0.

## BDD

Archivo: `features/HU-18-fecha-inicio-proyecto.feature`, con el runner `TestFeatures_HU18FechaInicioProyecto` en `internal/project/start_date_features_test.go`. Todos los pasos llaman al `Manager` de `internal/project`, no al handler HTTP. El escenario de error del repositorio usa un `MemoryRepository` envuelto que falla al actualizar.

Resultado de `go test -count=1 -v -run TestFeatures_HU18 ./internal/project/` sobre el GREEN:

```text
12 scenarios (12 passed)
53 steps (53 passed)
--- PASS: TestFeatures_HU18FechaInicioProyecto
```

## Fase RED

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `f864d55` | `docs(HU-18): agregar especificación SDD` |
| RED | `6a89d95` | `test(HU-18): agregar tests (RED)` |

Antes de los tests se agregaron solo los stubs que permite la SDD: el campo `StartDate` en `Project`, `ErrMissingStartDate` con su mensaje y `SetStartDate`, que devolvía `Project{}, nil`. Por el ajuste que pidió Peñalbé Hernán al aprobar la SDD, **la ruta `PATCH /api/proyectos/{id}/fecha-inicio` no se registró en el RED** (tampoco el handler, la columna del esquema ni el repositorio PostgreSQL): sin la ruta, los tests HTTP fallan por aserción con 405. `gofmt -l .` y `go vet ./...` no informaron nada antes del commit RED.

El RED falló por las aserciones y no por compilación, solo en `internal/project` e `internal/httpapi`. Con `grep` sobre la salida guardada de `go test -count=1 ./...`: **32 tests de primer nivel** (`^--- FAIL:`) y **38 subtests** (incluidos 11 escenarios godog) fallaron, y no hubo ninguna línea con `build failed`, `undefined:`, `cannot use` ni `[setup failed]`.

- Los unitarios fallaron con el resultado del stub (por ejemplo, `error = <nil>, se esperaba ErrMissingStartDate` o `fecha devuelta = 0001-01-01 00:00:00 +0000 UTC, se esperaba 2026-11-02`).
- BDD: `12 scenarios (1 passed, 11 failed)`, `53 steps (26 passed, 11 failed, 16 skipped)`.
- Los 8 `TestRegistrarFechaInicioProyecto_*` y los casos nuevos de `TestRutasConID_IDNoNumericoDevuelve400`, `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400` recibieron `405` (`Method Not Allowed`, `text/plain`) porque la ruta no existía. En esos tres tests que ya existían, el único subtest que falló fue el nuevo, “registrar fecha de inicio”.
- `TestProyecto_RespuestasIncluyenLaFechaDeInicio` falló porque la respuesta del `POST` no tenía `fecha_inicio` (`POST: fecha_inicio = <nil> (presente: false), se esperaba ""`) y la fecha consultada y la del `PUT` vinieron vacías; `TestModificarProyecto_ConservaLaFechaDeInicio`, porque la fecha consultada vino vacía (`fecha_inicio consultada = "", se esperaba "2026-10-23"`). En los dos, el `PATCH` previo también recibió `text/plain` en lugar de JSON.
- Los 3 tests de interfaz fallaron porque la página y `app.js` no tenían el formulario ni los textos.

**Ya pasaban en el RED**, sin aportar un fallo:

- Los 2 subtests de contrato nuevos en `TestMemoryRepository_Contrato`, porque `MemoryRepository` guarda el struct completo y conserva cualquier campo nuevo.
- `TestCreate_ProyectoNuevoNoTieneFechaDeInicio` y el escenario “Un proyecto recién creado no tiene fecha de inicio (CA-18.6)”, porque `Create` nunca asignó fecha y el campo nuevo queda en cero. Documentan el comportamiento de `main` para que el GREEN no lo rompa.
- `TestPostgresRepository_SinFechaDeInicioSeGuardaComoNULL`, `TestPostgresRepository_FechaDeInicioEnUTC` y `TestPostgresRepository_Contrato` hicieron SKIP porque no estaba definida `TEST_DATABASE_URL`.

## Fase GREEN

| Fase | Commit | Mensaje |
|---|---|---|
| GREEN | `67b290a` | `feat(HU-18): implementar registro de la fecha de inicio del proyecto (GREEN)` |
| Refactor | — | No aplica |
| Documentación | `72217fa` | `docs(HU-18): agregar evidencia TDD y cobertura` |

El GREEN tocó solo 8 archivos de producción: `SetStartDate` completo, el handler `setProjectStartDate` y su ruta, `ErrMissingStartDate` en el 400 de `statusFor`, `fecha_inicio` en `projectResponse`, la columna `start_date` en `db/init/01-schema.sql`, el repositorio PostgreSQL (`projectColumns`, `Save`, `Update`, `scanProject` y `startDateParam`), la interfaz web y `doc.go`. **Ningún archivo `*_test.go` ni el `.feature` cambió en el GREEN**: `git diff --stat HEAD -- '*_test.go' features/` salió vacío antes del commit.

Sin base, `go tool cover -func` mostró `SetStartDate`, `setProjectStartDate`, `toProjectResponse` y `statusFor` en 100.0%; las funciones del repositorio PostgreSQL de proyectos quedaron en 0.0% porque sus tests se omiten sin `TEST_DATABASE_URL` (con base, ver abajo).

## Pruebas de mutación

Cada mutación se aplicó sobre el GREEN, se corrió `go test -count=1 ./...` y se restauró con `git checkout -- <archivo>`; después de cada una, `git diff` salió vacío. La (d) y la (e) corrieron con `TEST_DATABASE_URL` apuntando al puerto 5434.

| Mutación | Resultado | Tests que fallan (mensaje) |
|---|---|---|
| (a) `start_date.go`: quitar el chequeo de fecha cero | Detectada | `TestSetStartDate_FechaAusenteDevuelveErrMissingStartDateYNoCambia` (`error = <nil>, se esperaba ErrMissingStartDate`); `TestSetStartDate_FechaAusenteSeInformaSinConsultarElRepositorio` (`error = no existe un proyecto con ese id, se esperaba ErrMissingStartDate`); `TestRegistrarFechaInicioProyecto_FechaAusenteDevuelve400` (`código = 200, se esperaba 400`); `TestRegistrarFechaInicioProyecto_FechaVaciaEnProyectoInexistenteDevuelve400` (`código = 404, se esperaba 400`); escenario “Rechazar una fecha ausente (CA-18.3)” |
| (b) `start_date.go`: no asignar la fecha | Detectada | 11 tests de `internal/project` además del runner godog (por ejemplo `TestSetStartDate_RegistraLaFechaYConservaLosDemasDatos`: `StartDate:0001-01-01 …, se esperaba … StartDate:2026-10-23 …`), 4 de `internal/httpapi` (por ejemplo `fecha_inicio consultada = "", se esperaba "2026-10-23"`) y 7 escenarios godog (`12 scenarios (5 passed, 7 failed)`) |
| (c) `start_date.go`: mutex solo alrededor de `Update`, no de `GetByID` | Sobrevive a la corrida normal; detectada en 1 de 50 lotes | `go test -count=1 ./...` terminó con exit 0. Con `-count=50 -run '^TestSetStartDate_RegistrosModificacionesYConsultasSimultaneosSinAccesosSinProteger$'`: 49 lotes PASS y 1 FAIL, `proyecto final = {ID:1 Name:Proyecto A … StartDate:2026-10-19 …}, se esperaban los datos de una de las modificaciones` (una modificación perdida) |
| (d) `postgres_repository.go`: sacar `start_date` del `UPDATE` | Detectada (con base) | Subtest de contrato “Update escribe y conserva la fecha de inicio” (`Update = {… StartDate:0001-01-01 …}, se esperaba {… StartDate:2026-10-23 …}`); `TestPostgresRepository_FechaDeInicioEnUTC` (`Update: día = 0001-01-01, se esperaba 2026-10-23`) |
| (e) `postgres_repository.go`: `startDateParam` devuelve siempre la fecha | Detectada (con base) | `TestPostgresRepository_SinFechaDeInicioSeGuardaComoNULL` (`start_date no es NULL para un proyecto sin fecha de inicio`) |
| (f) `projects.go`: formatear la fecha aunque sea cero | Detectada | `TestCrearProyecto_Devuelve201ConElProyectoCreado` y `TestModificarProyecto_Devuelve200ConElProyectoModificado` (`FechaInicio:0001-01-01`), `TestRegistrarFechaInicioProyecto_FechaAusenteDevuelve400` y `…_FormatoOFechaInvalidaDevuelve400` (`fecha_inicio consultada = "0001-01-01", se esperaba ""`), `TestProyecto_RespuestasIncluyenLaFechaDeInicio` y el subtest “registrar fecha de inicio” de `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400` |

La mutación (c) es una carrera: el test de concurrencia solo la detecta cuando el orden de las goroutines produce una modificación perdida, y sin `-race` eso pasó en 1 de 50 lotes. `-race` no se corrió localmente con la mutación (ver "-race").

## Prueba de punta a punta de la API

Servidor real compilado desde `67b290a` con `go build -o hu18-server.exe ./cmd/server` (fuera del repositorio) y arrancado en memoria con `PORT=18418 STORAGE=memory ./hu18-server.exe` (log: `servidor escuchando en :18418 (almacenamiento: memory)`). Pedidos con `curl` desde Git Bash. Los textos de entrada van sin tildes porque Git Bash no los envía en UTF-8. Al terminar se detuvo el servidor y se borró el binario.

```
$ curl -s -w '\nHTTP %{http_code}\n' -X POST localhost:18418/api/proyectos -d '{"nombre":"Tienda online","descripcion":"Proyecto de prueba HU-18"}'
{"id":1,"nombre":"Tienda online","descripcion":"Proyecto de prueba HU-18","estado":"PLANIFICADO","fecha_inicio":""}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH localhost:18418/api/proyectos/1/fecha-inicio -d '{"fecha_inicio":"2026-10-23"}'
{"id":1,"nombre":"Tienda online","descripcion":"Proyecto de prueba HU-18","estado":"PLANIFICADO","fecha_inicio":"2026-10-23"}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' localhost:18418/api/proyectos/1
{"id":1,"nombre":"Tienda online","descripcion":"Proyecto de prueba HU-18","estado":"PLANIFICADO","fecha_inicio":"2026-10-23"}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' -X PUT localhost:18418/api/proyectos/1 -d '{"nombre":"Tienda online","descripcion":"Proyecto en curso","estado":"EN_CURSO"}'
{"id":1,"nombre":"Tienda online","descripcion":"Proyecto en curso","estado":"EN_CURSO","fecha_inicio":"2026-10-23"}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH localhost:18418/api/proyectos/1/fecha-inicio -d '{"fecha_inicio":"2026-02-30"}'
{"error":"las fechas deben tener el formato AAAA-MM-DD"}

HTTP 400

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH localhost:18418/api/proyectos/1/fecha-inicio -d '{}'
{"error":"la fecha de inicio del proyecto es obligatoria"}

HTTP 400

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH localhost:18418/api/proyectos/999/fecha-inicio -d '{"fecha_inicio":"2026-10-23"}'
{"error":"no existe un proyecto con ese id"}

HTTP 404

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH localhost:18418/api/proyectos/1/fecha-inicio -d '{"fecha_inicio":"2026-11-01"}'
{"id":1,"nombre":"Tienda online","descripcion":"Proyecto en curso","estado":"EN_CURSO","fecha_inicio":"2026-11-01"}

HTTP 200
```

El proyecto nuevo sale sin fecha (`""`); la fecha registrada se ve en la consulta; el `PUT` sin fecha la conserva; una fecha inexistente y un cuerpo sin fecha dan 400, un proyecto inexistente 404, y registrar otra fecha la reemplaza. La línea vacía antes de `HTTP` es el salto de línea con el que la API termina cada respuesta JSON.

La interfaz web no se recorrió en un navegador: la cubren los tests de `internal/httpapi/web_test.go` (`TestInterfaz_FormularioDeFechaDeInicioDelProyecto`, `TestInterfaz_RegistrarFechaDeInicioMuestraElMensajeDeExito` y `TestInterfaz_DatosDelProyectoMuestranLaFechaDeInicio`), que verifican el formulario, la llamada de `app.js` y los textos, pero no ejecutan el JavaScript.

## Validación final

Medido sobre `67b290a` con go1.27.1 windows/amd64, sin `TEST_DATABASE_URL`.

### `gofmt -l .`

```text
$ gofmt -l .
(sin salida, exit 0)
```

### `go vet ./...`

```text
$ go vet ./...
(sin salida, exit 0)
```

### `go test -count=1 ./...`

```text
ok  	seguimiento-medicion-gpc/cmd/server	4.163s
ok  	seguimiento-medicion-gpc/internal/backlog	5.172s
ok  	seguimiento-medicion-gpc/internal/defect	4.343s
ok  	seguimiento-medicion-gpc/internal/effort	3.968s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.308s
ok  	seguimiento-medicion-gpc/internal/metrics	2.173s
ok  	seguimiento-medicion-gpc/internal/poker	3.129s
ok  	seguimiento-medicion-gpc/internal/postgres	1.977s
ok  	seguimiento-medicion-gpc/internal/postgrestest	3.061s
ok  	seguimiento-medicion-gpc/internal/project	4.170s
ok  	seguimiento-medicion-gpc/internal/sprint	4.380s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
(exit 0)
```

### `go test -count=1 -cover ./...`

```text
ok  	seguimiento-medicion-gpc/cmd/server	5.489s	coverage: 31.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	7.695s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/defect	6.695s	coverage: 76.6% of statements
ok  	seguimiento-medicion-gpc/internal/effort	7.812s	coverage: 75.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	5.861s	coverage: 97.6% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	4.373s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	3.503s	coverage: 69.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	2.574s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	4.020s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	5.595s	coverage: 67.5% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	4.887s	coverage: 76.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
(exit 0)
```

Con `-coverprofile` y `go tool cover -func`, sin base: `SetStartDate`, `setProjectStartDate`, `toProjectResponse`, `statusFor` y `NewHandler` en 100.0%; `NewPostgresRepository`, `Save`, `List`, `GetByID`, `Update`, `scanProject` y `startDateParam` de PostgreSQL de proyectos en 0.0% (sus tests se omiten sin `TEST_DATABASE_URL`); total **80.9%**.

Comparado con `main` (`850edc0`, el `git merge-base` con `origin/main` después de `git fetch origin`), medido con el mismo Go en una copia temporal (`git worktree` en `C:/wt-main`, borrada al terminar): `internal/project` sube de 67.4% a 67.5%, `internal/httpapi` de 97.5% a 97.6%, los demás paquetes se mantienen y el total del perfil sube de **80.8%** a **80.9%**.

### Con PostgreSQL

Medido sobre el GREEN con PostgreSQL 17.11 (imagen `postgres:17-alpine` del `docker-compose.yml` del repo), base recreada con `docker compose down -v` por la columna nueva. **La base se levantó en el puerto 5434 y no en el 5433**, porque el 5433 estaba ocupado por otro contenedor ajeno al proyecto; se usó `POSTGRES_PORT=5434` solo como variable del comando, sin `.env`. `information_schema` mostró `start_date timestamp with time zone nullable=YES` en `gpc` y en `gpc_test`.

Con `TEST_DATABASE_URL=postgres://gpc:gpc_dev@localhost:5434/gpc_test?sslmode=disable`:

- `go test -count=1 ./...`: todos los paquetes `ok`.
- `go test -count=1 -v ./...`: **0 SKIP, 1445 PASS y 0 FAIL** (contados con `grep -c` de `--- SKIP`, `--- PASS` y `--- FAIL`).
- `TestPostgresRepository_Contrato` de `internal/project` corrió sus 10 subtests contra PostgreSQL sin SKIP, incluidos “Save y GetByID conservan la fecha de inicio” y “Update escribe y conserva la fecha de inicio”; `TestPostgresRepository_SinFechaDeInicioSeGuardaComoNULL` y `TestPostgresRepository_FechaDeInicioEnUTC` pasaron.
- `go test -count=1 -cover ./...`: `internal/project` 97.4%, `internal/httpapi` 97.6% y `cmd/server` 33.3%.
- `go test -count=1 -coverpkg=./... ./...`: total **94.8%**. En el repositorio PostgreSQL de proyectos, `GetByID`, `Update`, `scanProject` y `startDateParam` quedan en 100.0%, `Save` en 80.0% y `List` en 75.0%. Lo que queda sin cubrir son los retornos de error `"guardar el proyecto: %w"` y `"listar los proyectos: %w"`, ramas que ya existían en `main` y que necesitan que la base falle en un momento puntual.

### `-race`

`-race` no se corrió localmente: `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`. El test de concurrencia (`TestSetStartDate_RegistrosModificacionesYConsultasSimultaneosSinAccesosSinProteger`) corrió sin `-race` y pasó; con `-race` lo ejecuta el CI del Pull Request (Linux, `go test -race -coverprofile=coverage.out ./...`).

## Commits

Los commits SDD, RED, GREEN y de documentación referencian el Issue #36 con `Refs #36`. Todos los commits de la historia están hechos y subidos a la rama `feature/HU-18-fecha-inicio-proyecto`, y el Pull Request es el [#131](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/131).

HU-18: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
