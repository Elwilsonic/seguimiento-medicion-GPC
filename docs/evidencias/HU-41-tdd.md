# Evidencia TDD · HU-41

> Issue del repositorio: #115

## Historia

Como Agile Enabler, quiero registrar la fecha de fin de un proyecto, para delimitar su duración.

## Resultado

HU-41 quedó implementada en la rama `feature/HU-41-fecha-fin-proyecto`. En `internal/project`:

- `Project` suma `EndDate time.Time`; el valor cero significa "sin registrar".
- `Manager.SetEndDate(id, date)` (`end_date.go`) valida en el orden de la SDD:
  1. sin tocar el repositorio, la fecha ausente (`ErrMissingEndDate`);
  2. con el mutex del `Manager` tomado, la existencia del proyecto (`ErrProjectNotFound`), la fecha de inicio registrada (`ErrStartDateNotRegistered`) y que el fin sea estrictamente posterior al inicio (`ErrEndNotAfterStart`; el mismo día se rechaza).
- `SetStartDate` (HU-18) suma una validación: con fecha de fin registrada, el inicio tiene que ser anterior (`ErrStartNotBeforeEnd`). Sin fecha de fin se comporta como antes.
- `PostgresRepository` lee y escribe la columna nueva `projects.end_date TIMESTAMPTZ` (nullable): el cero se guarda como `NULL` y una fecha se lee con `.UTC()`, como `start_date`.

La API lo expone en `PATCH /api/proyectos/{id}/fecha-fin` con `{"fecha_fin":"AAAA-MM-DD"}`. Todas las respuestas de proyectos suman `fecha_fin` (`AAAA-MM-DD` o `""`), y también el campo `proyecto` del reporte de HU-38, que usa `toProjectResponse`. En `statusFor`, `ErrMissingEndDate` da 400, y `ErrStartDateNotRegistered`, `ErrEndNotAfterStart` y `ErrStartNotBeforeEnd` dan 409. La interfaz web suma, en la sección Proyectos y debajo de "Fecha de inicio", el formulario "Fecha de fin", y los datos del proyecto muestran "fin: AAAA-MM-DD" o "sin fecha de fin".

**Otras historias afectadas:**

- **HU-38:** el código no cambia, pero sus dos contratos JSON (`TestReporteProyecto_ContratoDelEjemploDeTEC03` y `TestReporteProyecto_ContratoDelProyectoVacio`) suman `"fecha_fin": ""` en el campo `proyecto`. Se cambiaron **en el commit RED**, porque el GREEN no toca tests (D9 de la SDD). Hay que avisarle a Cardozo Leandro, responsable de HU-38.
- **HU-19** (Issue #37, "Modificar las fechas del proyecto"): D4 y D5 cubren parte de ella, volver a registrar las fechas y validar que el fin siga siendo posterior al inicio (CA-19.1 y CA-19.2 de su borrador). HU-19 sigue abierta y fuera de la entrega, y su Issue no se modificó.

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test |
|---|---|
| CA-41.1 — se registra la fecha de fin de un proyecto con fecha de inicio, sin cambiar sus otros datos, en cualquier estado | “Registrar la fecha de fin de un proyecto con fecha de inicio”; “Registrar la fecha de fin en un proyecto finalizado”; `TestSetEndDate_RegistraLaFechaYConservaLosDemasDatos`; `TestSetEndDate_SeRegistraEnCualquierEstado`; `TestSetEndDate_AceptaElDiaSiguienteFechasFuturasYDeAnioBisiesto`; `TestRegistrarFechaFinProyecto_Devuelve200ConElProyecto` |
| CA-41.2 — fin estrictamente posterior al inicio | “Rechazar una fecha de fin igual o anterior a la de inicio” (esquema, 3 ejemplos); `TestSetEndDate_FechaIgualOAnteriorAlInicioDevuelveErrEndNotAfterStart`; `TestRegistrarFechaFinProyecto_FechaNoPosteriorAlInicioDevuelve409` |
| CA-41.3 — formato `AAAA-MM-DD` y fecha existente | `TestRegistrarFechaFinProyecto_FormatoOFechaInvalidaDevuelve400` (6 casos); `TestRegistrarFechaFinProyecto_JSONInvalidoDevuelve400` |
| CA-41.4 — fecha obligatoria | “Rechazar una fecha de fin ausente”; `TestSetEndDate_FechaAusenteDevuelveErrMissingEndDate`; `TestSetEndDate_FechaAusenteEnProyectoInexistenteDevuelveErrMissingEndDate`; `TestRegistrarFechaFinProyecto_FechaAusenteDevuelve400` (vacía, `null` y ausente); `TestRegistrarFechaFinProyecto_FechaVaciaEnProyectoInexistenteDevuelve400` |
| CA-41.5 — proyecto inexistente (incluidos 0 y negativo) | “Rechazar un proyecto inexistente” (esquema con 999, 0 y -1); `TestSetEndDate_ProyectoInexistenteDevuelveErrProjectNotFound`; `TestRegistrarFechaFinProyecto_InexistenteDevuelve404` |
| CA-41.6 — sin fecha de inicio, se rechaza | “Rechazar la fecha de fin de un proyecto sin fecha de inicio”; `TestSetEndDate_SinFechaDeInicioDevuelveErrStartDateNotRegistered`; `TestRegistrarFechaFinProyecto_SinFechaDeInicioDevuelve409` |
| CA-41.7 — volver a registrar | “Volver a registrar la fecha de fin”; `TestSetEndDate_VolverARegistrarReemplazaLaFecha`; `TestSetEndDate_VolverARegistrarUnaFechaInvalidaConservaLaAnterior`; `TestSetEndDate_MismaFechaEsValida`; `TestRegistrarFechaFinProyecto_VolverARegistrarReemplazaLaFecha` |
| CA-41.8 — la fecha de inicio no puede quedar igual o posterior al fin | “Rechazar una fecha de inicio igual o posterior a la de fin registrada” (esquema, 2 ejemplos); “Aceptar una fecha de inicio anterior a la de fin registrada”; `TestSetStartDate_ConFechaDeFinRechazaUnInicioIgualOPosterior`; `TestSetStartDate_ConFechaDeFinAceptaUnInicioAnterior`; `TestRegistrarFechaInicioProyecto_ConFechaDeFinValidaElOrden` |
| CA-41.9 — proyecto nuevo sin fecha de fin; consulta, reporte y conservación | “Un proyecto recién creado no tiene fecha de fin”; “Modificar el proyecto o su fecha de inicio conserva la fecha de fin”; `TestSetEndDate_ProyectoNuevoNoTieneFechaDeFin`; `TestSetEndDate_UpdateYSetStartDateConservanLaFechaDeFin`; `TestProyecto_RespuestasIncluyenLaFechaDeFin`; `TestModificarProyecto_ConservaLaFechaDeFin`; `TestReporteProyecto_IncluyeLaFechaDeFin`; los dos contratos de `reports_test.go` |
| CA-41.10 — copia por valor | “Modificar el proyecto devuelto no altera el guardado”; `TestSetEndDate_DevuelveUnaCopia` |
| CA-41.11 — concurrencia y falla del repositorio | “Error del repositorio”; “Registros de fechas simultáneos mantienen el orden de las fechas”; `TestSetEndDate_ErrorDelRepositorioAlActualizar`; `TestSetEndDate_ErrorDelRepositorioAlBuscar`; `TestSetEndDate_RegistrosSimultaneosMantienenElOrdenDeLasFechas`; `TestRegistrarFechaFinProyecto_ErrorInternoDevuelve500` |
| CA-41.12 — persistencia en memoria y PostgreSQL | `TestMemoryRepository_ContratoFechaDeFin` y `TestPostgresRepository_ContratoFechaDeFin` (2 subtests de contrato cada uno); `TestPostgresRepository_SinFechaDeFinSeGuardaComoNULL`; `TestPostgresRepository_FechaDeFinEnUTC` |
| CA-41.13 — API e interfaz web | Los 14 tests de `internal/httpapi/project_end_date_test.go`; caso “registrar fecha de fin” de `TestRutasConID_IDNoNumericoDevuelve400`; endpoint “registrar fecha de fin” de `bodyEndpoints` (413 en `TestHandler_CuerpoMayorA1MiBDa413` y texto después del JSON en `TestHandler_TextoDespuesDelJSONDa400`); `TestInterfaz_FormularioDeFechaDeFinDelProyecto`; `TestInterfaz_RegistrarFechaDeFinMuestraElMensajeDeExito`; `TestInterfaz_DatosDelProyectoMuestranLaFechaDeFin` |

El archivo BDD tiene **14 bloques de escenario**: 11 escenarios y 3 esquemas con 8 ejemplos en total, que godog ejecuta como **19 escenarios**. Las pruebas son **41 funciones de test nuevas**:

- 19 unitarias de `internal/project/end_date_test.go`;
- 1 runner de godog;
- 1 de contrato en memoria y 3 de PostgreSQL;
- 14 HTTP;
- 3 de interfaz web.

Además, cada una de las dos funciones de contrato corre 2 subtests de `runRepositoryEndDateContract`. En tests que ya existían hay 2 casos de tabla nuevos (el de `TestRutasConID_IDNoNumericoDevuelve400` y el endpoint de `bodyEndpoints`, que corre en dos tests) y 2 líneas `"fecha_fin": ""` en los contratos de `reports_test.go`. No se modificaron `start_date_test.go` ni `start_date_features_test.go` (HU-18).

La correspondencia no es 1 a 1. CA-41.3, CA-41.12 y CA-41.13 no tienen escenario BDD, porque el formato es de la capa HTTP, la persistencia se prueba con los tests de contrato y de PostgreSQL, y la API y la interfaz con los de `internal/httpapi`.

Conteos con `grep` sobre los archivos de la rama (contra `e61f21a`, el commit anterior al RED, para los archivos que ya existían):

```text
$ grep -c '^\s*Escenario:' features/HU-41-fecha-fin-proyecto.feature
11
$ grep -c '^\s*Esquema del escenario:' features/HU-41-fecha-fin-proyecto.feature
3
$ grep -cE '^\s+\| ' features/HU-41-fecha-fin-proyecto.feature
11
$ grep -c '^func Test' internal/project/end_date_test.go
19
$ grep -c '^func Test' internal/project/end_date_features_test.go
1
$ grep -c '^func Test' internal/httpapi/project_end_date_test.go
14
$ git diff e61f21a 3b01a87 -- internal/project/repository_contract_test.go | grep -c '^+func Test'
1
$ git diff e61f21a 3b01a87 -- internal/project/postgres_repository_test.go | grep -c '^+func Test'
3
$ git diff e61f21a 3b01a87 -- internal/httpapi/web_test.go | grep -c '^+func Test'
3
$ git diff e61f21a 3b01a87 -- internal/httpapi/server_test.go | grep -c '^+func Test'
0
$ git diff e61f21a 3b01a87 -- internal/httpapi/reports_test.go | grep -c '^+func Test'
0
$ git diff e61f21a 3b01a87 -- internal/project/repository_contract_test.go | grep -cE '^\+\s+t\.Run\("'
2
$ git diff e61f21a 3b01a87 -- internal/httpapi/server_test.go | grep -cE '^\+.*"registrar fecha de fin"'
2
$ git diff e61f21a 3b01a87 -- internal/httpapi/reports_test.go | grep -cE '^\+\s+"fecha_fin": ""'
2
$ git diff --stat e61f21a 3b01a87 -- internal/project/start_date_test.go internal/project/start_date_features_test.go
(sin salida)
```

Las 11 filas de tabla son las 3 cabeceras y los 8 ejemplos de los esquemas.

Bloques de escenario por criterio (`grep` del CA en el nombre de cada escenario o esquema): CA-41.1: 2; CA-41.2: 1; CA-41.3: 0; CA-41.4: 1; CA-41.5: 1; CA-41.6: 1; CA-41.7: 1; CA-41.8: 2; CA-41.9: 2; CA-41.10: 1; CA-41.11: 2; CA-41.12: 0; CA-41.13: 0.

## BDD

Archivo: `features/HU-41-fecha-fin-proyecto.feature`, con el runner `TestFeatures_HU41FechaFinProyecto` en `internal/project/end_date_features_test.go`. Todos los pasos llaman al `Manager` de `internal/project`, no al handler HTTP. El repositorio en memoria va envuelto en `updateFailingRepository` (de los tests de HU-18, sin modificarlo), que puede fallar al actualizar.

El escenario de concurrencia parte de inicio 2026-10-01 y fin 2026-12-31, y registra en simultáneo, 20 veces cada uno, el inicio 2026-10-05 (siempre válido), el inicio 2027-01-15 (siempre posterior al fin), el fin 2026-12-20 (siempre válido) y el fin 2026-09-01 (siempre anterior al inicio). Así, el resultado de cada registro no depende del orden y el escenario puede exigir el error exacto de cada uno y las fechas finales.

Resultado de `go test -count=1 -v -run 'TestFeatures_HU(18|41)' ./internal/project/` sobre el GREEN (`hu41-green-godog.txt`):

```text
19 scenarios (19 passed)
101 steps (101 passed)
--- PASS: TestFeatures_HU41FechaFinProyecto
12 scenarios (12 passed)
53 steps (53 passed)
--- PASS: TestFeatures_HU18FechaInicioProyecto
```

HU-18 da los mismos 12 escenarios y 53 pasos que en el RED.

## Fase RED

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `6005543` | `docs(HU-41): agregar especificación SDD` |
| Ajuste de la SDD | `1fcf1ae` | `docs(HU-41): aclarar el impacto en los contratos del reporte de HU-38` |
| Aprobación de la SDD | `9ef8167` | `docs(HU-41): registrar la aprobación de la SDD` |
| Merge de `main` | `e61f21a` | `chore(HU-41): traer main con HU-39` |
| RED | `3b01a87` | `test(HU-41): agregar tests (RED)` |

El merge trajo HU-39 (PR #149, `1dbff43`) antes de escribir los tests. No tuvo conflictos, y `go build ./...` y `go test ./...` sin base dieron `ok` en todos los paquetes. El reporte de sprint de HU-39 (`sprint_reports_test.go`) compara el objeto `sprint` y no un `proyecto`, así que `fecha_fin` no lo afecta.

Antes de los tests se agregaron solo los stubs que permite la SDD:

- el campo `EndDate` en `Project`;
- los cuatro errores nuevos con sus mensajes;
- la firma vacía de `SetEndDate`, que devolvía `Project{}, nil`.

Siguiendo el ajuste que pidió Peñalbé Hernán en HU-18, **la ruta `PATCH /api/proyectos/{id}/fecha-fin`, el handler, la columna del esquema, el repositorio PostgreSQL, la validación nueva de `SetStartDate` y el formulario web no estaban en el RED**. `gofmt -l .` y `go vet ./...` no informaron nada antes del commit RED.

### Sin base

El RED falló por las aserciones y no por compilación, solo en `internal/project` e `internal/httpapi`. Se contó con `grep` sobre la salida guardada de `go test -count=1 ./...` (`hu41-red-sin-base.txt`):

- fallaron **40 tests de primer nivel** (`^--- FAIL`) y **55 subtests** (`^\s+--- FAIL`);
- no hubo ninguna línea con `build failed`, `undefined:`, `cannot use` ni `[setup failed]`;
- BDD: `19 scenarios (1 passed, 18 failed)`, `101 steps (50 passed, 18 failed, 33 skipped)`.

| Tests que fallaron | Cantidad | Motivo |
|---|---|---|
| `TestSetEndDate_*` y `TestSetStartDate_ConFechaDeFinRechazaUnInicioIgualOPosterior` | 17 | El stub devuelve `nil` y no guarda nada, por ejemplo `err = <nil>, se esperaba la fecha de fin debe ser posterior a la de inicio`; en el test de concurrencia, 20 veces `SetStartDate(2027-01-15): err = <nil>, se esperaba la fecha de inicio debe ser anterior a la de fin` y 20 veces lo mismo para `SetEndDate(2026-09-01)` |
| `TestFeatures_HU41FechaFinProyecto` | 1 | 18 escenarios fallan |
| Los 14 de `project_end_date_test.go` | 14 | Sin la ruta, `PATCH /api/proyectos/1/fecha-fin` responde `405 Method Not Allowed` en `text/plain`, por ejemplo `código = 405, se esperaba 400; cuerpo: Method Not Allowed` |
| `TestReporteProyecto_ContratoDelEjemploDeTEC03` y `TestReporteProyecto_ContratoDelProyectoVacio` | 2 | Esperan `"fecha_fin": ""` en el campo `proyecto`, que `toProjectResponse` todavía no devuelve |
| Los 3 de `web_test.go` | 3 | `index.html` y `app.js` no tienen el formulario ni los textos, por ejemplo `la sección de proyectos no contiene "id=\"form-fecha-fin-proyecto\""` |
| `TestRutasConID_IDNoNumericoDevuelve400`, `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400` | 3 | Ya existían; fallan solo por su subtest nueva `registrar_fecha_de_fin` (1 cada uno), que recibió 405. Sus demás subtests pasaron |

Ningún otro test que ya existía falló.

### Con base

Con `TEST_DATABASE_URL` en el puerto 5434 y la base **sin recrear** (esquema viejo, sin `end_date`), fallaron **43 tests de primer nivel** con **0 SKIP** (`hu41-red-con-base.txt`): los mismos 40 y los 3 de PostgreSQL de HU-41.

- `TestPostgresRepository_SinFechaDeFinSeGuardaComoNULL`: `leer end_date: ERROR: column "end_date" does not exist (SQLSTATE 42703)`.
- `TestPostgresRepository_ContratoFechaDeFin` (2 subtests) y `TestPostgresRepository_FechaDeFinEnUTC`: el repositorio no escribe `end_date`, así que se lee en cero, por ejemplo `GetByID: día = 0001-01-01, se esperaba 2026-12-18`.

Los demás tests de PostgreSQL pasaron con el esquema viejo.

### Tests de HU-41 que ya pasaban en el RED (4)

No son falsos positivos: caracterizan comportamiento que ya existía y que HU-41 tiene que conservar.

- `TestMemoryRepository_ContratoFechaDeFin`: el repositorio en memoria guarda el struct completo, así que conserva `EndDate` sin cambios (como el contrato de memoria de HU-18). El mismo contrato en PostgreSQL sí falló.
- `TestSetStartDate_ConFechaDeFinAceptaUnInicioAnterior`: HU-18 ya acepta cualquier fecha de inicio. El test vigila que la validación nueva (D5) no rechace de más; el que verifica el rechazo (`…RechazaUnInicioIgualOPosterior`) falló.
- `TestSetEndDate_ProyectoNuevoNoTieneFechaDeFin` y el escenario “Un proyecto recién creado no tiene fecha de fin (CA-41.9)”: el campo nuevo arranca en cero.

Los 3 tests de PostgreSQL hicieron SKIP sin `TEST_DATABASE_URL` y fallaron con la base.

## Fase GREEN

| Fase | Commit | Mensaje |
|---|---|---|
| GREEN | `5eb7a07` | `feat(HU-41): registrar la fecha de fin del proyecto (GREEN)` |
| Refactor | — | No aplica |
| Documentación | (propuesto) | `docs(HU-41): agregar evidencia TDD y cobertura` |

El GREEN tocó solo 9 archivos de producción:

- `internal/project`: `end_date.go` (`SetEndDate`), `start_date.go` (validación de RN-41.7), `postgres_repository.go` (`end_date` en las columnas, el `INSERT`, el `UPDATE` y `scanProject`; `startDateParam` pasó a ser `dateParam`, para las dos fechas) y `doc.go`;
- `db/init/01-schema.sql`: la columna `end_date TIMESTAMPTZ` nullable, con su comentario;
- `internal/httpapi`: `projects.go` (`setProjectEndDate`, `endDateRequest`, `FechaFin` y su formateo) y `server.go` (la ruta y los cuatro errores en `statusFor`);
- `internal/web`: `index.html` y `app.js`.

**Ningún archivo `*_test.go` ni el `.feature` cambió en el GREEN**: `git diff --stat 3b01a87 -- '*_test.go' features/` salió vacío antes del commit.

### Recreación de la base

Como el esquema cambió, la base se recreó con `POSTGRES_PORT=5434 docker compose down -v` y `POSTGRES_PORT=5434 docker compose up -d --wait`. El contenedor quedó `healthy`, y con `psql` se verificó que `projects` tiene `end_date timestamp with time zone nullable=YES` en `gpc` y en `gpc_test`, con PostgreSQL 17.11 (`hu41-green-recrear-base.txt`).

### Resultado

`gofmt -l .`, `go vet ./...` y `go build ./...` terminaron sin salida y con exit 0 (`hu41-green-gofmt-vet-build.txt`). Con `go test ./... -count=1 -v`, contando con `grep -c` las líneas `--- PASS`, `--- FAIL` y `--- SKIP` (incluidas las subtests):

| Corrida | Paquetes | PASS | FAIL | SKIP | Salida |
|---|---|---|---|---|---|
| Sin base | todos `ok` | 2006 | 0 | 37 | `hu41-green-final-sin-base.txt` |
| Con base (`TEST_DATABASE_URL` en el puerto 5434) | todos `ok` | 2116 | 0 | 0 | `hu41-green-final-con-base.txt` |

- Los 43 tests que fallaron en el RED con base dieron PASS (`hu41-green-ex-red.txt`).
- Los 4 que ya pasaban en el RED siguieron en PASS.

## Pruebas de mutación

Cada mutación se aplicó sobre el GREEN y se corrió `go test -count=1` sobre `./internal/project/` y `./internal/httpapi/` (la (d), solo sobre `./internal/httpapi/`, y la (e), solo sobre `./internal/project/` con `TEST_DATABASE_URL` en el puerto 5434). Salidas: `hu41-mut-<letra>.txt`.

| Mutación | Resultado | Tests que fallan (mensaje) |
|---|---|---|
| (a) `end_date.go`: aceptar fin igual a inicio (`!date.After(...)` → `date.Before(...)`) | Detectada (3 tests) | `TestSetEndDate_FechaIgualOAnteriorAlInicioDevuelveErrEndNotAfterStart` (`err = <nil>, se esperaba la fecha de fin debe ser posterior a la de inicio`); `TestRegistrarFechaFinProyecto_FechaNoPosteriorAlInicioDevuelve409` (`código = 200, se esperaba 409`); runner godog |
| (b) `end_date.go`: quitar la validación de "sin fecha de inicio" | Detectada (3 tests) | `TestSetEndDate_SinFechaDeInicioDevuelveErrStartDateNotRegistered` (`err = <nil>, se esperaba el proyecto no tiene fecha de inicio registrada`); `TestRegistrarFechaFinProyecto_SinFechaDeInicioDevuelve409` (`código = 200, se esperaba 409`); runner godog |
| (c) `start_date.go`: quitar la validación nueva de `SetStartDate` | Detectada (4 tests) | `TestSetStartDate_ConFechaDeFinRechazaUnInicioIgualOPosterior` (`err = <nil>, se esperaba la fecha de inicio debe ser anterior a la de fin`); `TestSetEndDate_RegistrosSimultaneosMantienenElOrdenDeLasFechas` (`SetStartDate(2027-01-15): err = <nil>, …` y, como el inicio pasa a quedar después del fin, `SetEndDate(2026-12-20): err = la fecha de fin debe ser posterior a la de inicio, se esperaba <nil>`); `TestRegistrarFechaInicioProyecto_ConFechaDeFinValidaElOrden` (`código = 200, se esperaba 409`); runner godog |
| (d) `server.go`: `ErrEndNotAfterStart` en el 400 | Detectada (1 test) | `TestRegistrarFechaFinProyecto_FechaNoPosteriorAlInicioDevuelve409` (`código = 400, se esperaba 409`) |
| (e) `postgres_repository.go`: el `UPDATE` no escribe `end_date` | Detectada con base (2 tests) | `TestPostgresRepository_ContratoFechaDeFin` (`Update = {… EndDate:0001-01-01 …}, se esperaba {… EndDate:2026-12-18 …}`) y `TestPostgresRepository_FechaDeFinEnUTC` (`día = 0001-01-01, se esperaba 2026-12-18`) |
| (f) `end_date.go`: quitar el mutex de `SetEndDate` | **Sobrevive** | `ok` en los dos paquetes. Con `-count=50 -v` del test unitario y del escenario de concurrencia: 50 PASS y 0 FAIL en cada uno (`hu41-mut-f-count50.txt`) |

En la (e), el `SET` quedó con `end_date = CASE WHEN $6::timestamptz IS NULL THEN end_date ELSE end_date END`. Así `$6` sigue referenciado, PostgreSQL no rechaza el parámetro sin usar y la fecha guardada no cambia.

**La mutación (f) sobrevive.** El `MemoryRepository` de proyectos no tiene un mutex propio: lo protege el mutex del `Manager`. Sin ese mutex en `SetEndDate`, sus lecturas y escrituras del repositorio quedan sin protección frente a `SetStartDate`, `Update` y `Get`, que sí lo toman. Es un acceso sin protección que solo detecta `-race`, y **`-race` no corre localmente** (`go: -race requires cgo`). Los tests de concurrencia, sin `-race`, no lo observaron en 50 lotes. Lo verifica el CI del Pull Request, que corre `go test -race`. No se cambiaron los tests.

### Incidente durante las mutaciones

Las mutaciones se revirtieron primero con `git checkout -- <archivo>`, como pedía el prompt. Pero el GREEN todavía no estaba commiteado, así que el primer `git checkout -- internal/project/end_date.go`, al revertir la (a), dejó ese archivo como el stub del RED (`3b01a87`) y borró la implementación de `SetEndDate`.

- Se restauró `end_date.go` con el mismo contenido de la implementación. `gofmt`, `go vet` y los tests de `internal/project` e `internal/httpapi` dieron `ok`.
- Se guardó una copia de respaldo de los 9 archivos del GREEN y un `green.patch` con su `git diff`. De ahí en adelante, cada mutación se revirtió copiando el archivo desde ese respaldo, y se comprobó con `cmp` que el archivo y el `git diff` quedaran idénticos al GREEN.
- La (a) se volvió a correr con el respaldo y dio el mismo resultado.
- Antes del commit GREEN se repitió la verificación completa, con y sin base, y dio los mismos números que la primera (2006/0/37 sin base y 2116/0/0 con base).

## Prueba de punta a punta de la API

El servidor real se compiló desde el GREEN con `go build -o hu41-server.exe ./cmd/server`, fuera del repositorio. Se arrancó en memoria, sin `STORAGE` ni `TEST_DATABASE_URL`, con `PORT=18941 ./hu41-server.exe` (log: `servidor escuchando en :18941 (almacenamiento: memory)`). Los pedidos se hicieron con `curl` desde Git Bash, con textos sin tildes. Al terminar se detuvo el servidor (`tasklist` no mostró ningún `hu41-server.exe` y el puerto 18941 quedó sin procesos escuchando) y se borró el binario. Salida: `hu41-e2e.txt`.

```
$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18941/api/proyectos -d '{"nombre":"Sistema de ventas","descripcion":"Gestion de ventas"}'
{"id":1,"nombre":"Sistema de ventas","descripcion":"Gestion de ventas","estado":"PLANIFICADO","fecha_inicio":"","fecha_fin":""}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18941/api/proyectos/1/fecha-fin -d '{"fecha_fin":"2026-12-18"}'
{"error":"el proyecto no tiene fecha de inicio registrada"}

HTTP 409

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18941/api/proyectos/1/fecha-inicio -d '{"fecha_inicio":"2026-10-23"}'
{"id":1,"nombre":"Sistema de ventas","descripcion":"Gestion de ventas","estado":"PLANIFICADO","fecha_inicio":"2026-10-23","fecha_fin":""}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18941/api/proyectos/1/fecha-fin -d '{"fecha_fin":"2026-12-18"}'
{"id":1,"nombre":"Sistema de ventas","descripcion":"Gestion de ventas","estado":"PLANIFICADO","fecha_inicio":"2026-10-23","fecha_fin":"2026-12-18"}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18941/api/proyectos/1/fecha-fin -d '{"fecha_fin":"2026-10-23"}'
{"error":"la fecha de fin debe ser posterior a la de inicio"}

HTTP 409

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18941/api/proyectos/1/fecha-fin -d '{"fecha_fin":"18/12/2026"}'
{"error":"las fechas deben tener el formato AAAA-MM-DD"}

HTTP 400

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18941/api/proyectos/1/fecha-inicio -d '{"fecha_inicio":"2026-12-19"}'
{"error":"la fecha de inicio debe ser anterior a la de fin"}

HTTP 409

$ curl -s -w '\nHTTP %{http_code}\n' http://localhost:18941/api/proyectos/1
{"id":1,"nombre":"Sistema de ventas","descripcion":"Gestion de ventas","estado":"PLANIFICADO","fecha_inicio":"2026-10-23","fecha_fin":"2026-12-18"}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' http://localhost:18941/api/proyectos/1/reporte
{"proyecto":{"id":1,"nombre":"Sistema de ventas","descripcion":"Gestion de ventas","estado":"PLANIFICADO","fecha_inicio":"2026-10-23","fecha_fin":"2026-12-18"},"metricas":{…},"esfuerzos":{"esfuerzos":[],"total_horas":0},"defectos":{"defectos":[],"cantidad":0},"integrantes":{"integrantes":[],"cantidad":0}}

HTTP 200
```

En el último pedido se abrevió el objeto `metricas` con `…`; la respuesta completa está en `hu41-e2e.txt`. La línea vacía antes de `HTTP` es el salto de línea con el que la API termina cada respuesta JSON.

Resultados del recorrido:

- Un proyecto nuevo tiene `fecha_fin` en `""`.
- Sin fecha de inicio, la fecha de fin da 409.
- Con el inicio registrado, el fin da 200 y queda en la respuesta.
- El mismo día que el inicio da 409, y un formato inválido, 400.
- Mover el inicio después del fin da 409.
- La consulta del proyecto y el reporte devuelven `fecha_fin`.

La interfaz web no se recorrió en un navegador. La cubren los tests de `internal/httpapi/web_test.go` (`TestInterfaz_FormularioDeFechaDeFinDelProyecto`, `TestInterfaz_RegistrarFechaDeFinMuestraElMensajeDeExito` y `TestInterfaz_DatosDelProyectoMuestranLaFechaDeFin`), que verifican el formulario, la llamada de `app.js` y los textos, pero no ejecutan el JavaScript.

## Validación final

Medido sobre `5eb7a07` con go1.27.1 windows/amd64.

### `go test -count=1 -cover ./...` sin base

Salida guardada en `hu41-cover-rama.txt`:

```text
ok  	seguimiento-medicion-gpc/cmd/server	4.304s	coverage: 31.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	4.131s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/dashboard	1.934s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/defect	3.215s	coverage: 77.7% of statements
ok  	seguimiento-medicion-gpc/internal/effort	2.578s	coverage: 75.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	4.924s	coverage: 98.8% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	2.416s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	4.247s	coverage: 69.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.395s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.708s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	3.928s	coverage: 71.1% of statements
ok  	seguimiento-medicion-gpc/internal/report	1.890s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	3.933s	coverage: 76.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

Con `-coverprofile` y `go tool cover -func`, sin base (`hu41-cover-rama-func.txt`):

- en 100.0%: `SetEndDate`, `SetStartDate`, `setProjectEndDate`, `setProjectStartDate`, `toProjectResponse` y `statusFor`;
- las funciones del repositorio PostgreSQL de proyectos (incluidas `scanProject` y `dateParam`) en 0.0%, porque sus tests se omiten sin `TEST_DATABASE_URL`;
- total **83.9%**.

Para comparar con `main` se usó `1dbff43`, el `git merge-base` con `origin/main` después de `git fetch origin`. Se midió con el mismo Go en una copia temporal (`git worktree` en `C:/wt-main`, borrada al terminar; salidas en `hu41-cover-main.txt` y `hu41-cover-main-total.txt`):

- `internal/project` sube de 67.5% a 71.1%;
- `internal/httpapi` sube de 98.7% a 98.8%;
- los demás paquetes se mantienen;
- el total del perfil sube de **83.7%** a **83.9%**.

Los perfiles se borraron al terminar.

### Con PostgreSQL

Con PostgreSQL 17.11 en el puerto 5434, recreado con la columna `end_date`. **La base corre en el 5434 y no en el 5433**, porque el 5433 lo ocupa otro contenedor ajeno al proyecto. Con `TEST_DATABASE_URL=postgres://gpc:gpc_dev@localhost:5434/gpc_test?sslmode=disable`:

- `go test -count=1 -v ./...`: **0 SKIP, 2116 PASS y 0 FAIL** (`hu41-green-final-con-base.txt`).
- `go test -count=1 -cover ./...` (`hu41-cover-con-base.txt`): `internal/project` 97.8%, `internal/httpapi` 98.8% y `cmd/server` 33.3%.
- `go test -count=1 -coverpkg=./... -coverprofile=… ./...` (`hu41-coverpkg.txt` y `hu41-coverpkg-func.txt`): total **95.9%**. En el repositorio PostgreSQL de proyectos, `Update`, `GetByID`, `scanProject` y `dateParam` quedan en 100.0%, `Save` en 80.0% y `List` en 75.0%; lo que falta en esas dos son las ramas de error de la base, que los tests no simulan.

El perfil se borró al terminar.

### `-race`

`-race` no se corrió localmente: `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`. Los tests de concurrencia (`TestSetEndDate_RegistrosSimultaneosMantienenElOrdenDeLasFechas` y el escenario de CA-41.11) corrieron sin `-race` y pasaron. Con `-race` los ejecuta el CI del Pull Request (Linux, `go test -race -coverprofile=coverage.out ./...`), que es el que puede detectar la mutación (f).

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `6005543` | `docs(HU-41): agregar especificación SDD` |
| Ajuste de la SDD | `1fcf1ae` | `docs(HU-41): aclarar el impacto en los contratos del reporte de HU-38` |
| Aprobación de la SDD | `9ef8167` | `docs(HU-41): registrar la aprobación de la SDD` |
| Merge de `main` | `e61f21a` | `chore(HU-41): traer main con HU-39` |
| RED | `3b01a87` | `test(HU-41): agregar tests (RED)` |
| GREEN | `5eb7a07` | `feat(HU-41): registrar la fecha de fin del proyecto (GREEN)` |
| Refactor | — | No aplica |
| Documentación | (propuesto) | `docs(HU-41): agregar evidencia TDD y cobertura` |

Todos los commits de la historia referencian el Issue #115 con `Refs #115` y están en la rama `feature/HU-41-fecha-fin-proyecto`.

HU-41: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
