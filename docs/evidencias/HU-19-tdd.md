# Evidencia TDD · HU-19

> Issue del repositorio: #37

## Historia

Como Agile Enabler, quiero modificar las fechas de un proyecto, para reflejar cambios en la planificación.

## Resultado

HU-19 quedó implementada en la rama `feature/HU-19-modificar-fechas-proyecto`. En `internal/project`:

- `ErrDatesNotRegistered` es el error nuevo para "el proyecto no tiene registradas las fechas de inicio y de fin".
- `Manager.SetDates(id, start, end)` (`dates.go`) modifica juntas las dos fechas y valida en el orden de la SDD:
  1. sin tocar el repositorio, el inicio ausente (`ErrMissingStartDate`) y después el fin ausente (`ErrMissingEndDate`);
  2. con el mutex del `Manager` tomado, la existencia del proyecto (`ErrProjectNotFound`), que ya tenga las dos fechas registradas (`ErrDatesNotRegistered`) y que el fin nuevo sea estrictamente posterior al inicio nuevo (`ErrEndNotAfterStart`; el mismo día se rechaza).
- El par nuevo **no se compara con las fechas guardadas**, así que se puede postergar o adelantar el proyecto en una sola operación. Las dos fechas se guardan con un solo `Update`.
- No cambian el esquema, los repositorios, `SetStartDate` ni `SetEndDate`.

La API lo expone en `PATCH /api/proyectos/{id}/fechas` con `{"fecha_inicio":"AAAA-MM-DD","fecha_fin":"AAAA-MM-DD"}`, que devuelve el proyecto con la forma de `toProjectResponse`, sin campos nuevos. En `statusFor`, `ErrDatesNotRegistered` se suma al 409; los demás errores ya estaban mapeados (`ErrEndNotAfterStart` da 409, como en HU-41, por D4). La interfaz web suma, en la sección Proyectos y debajo de "Fecha de fin", el formulario "Modificar fechas", con los dos campos precargados por `mostrarProyecto`.

### Qué ya cubrían HU-18 y HU-41

| Criterio del borrador del Issue #37 | Qué ya existía en `main` | Qué agrega HU-19 |
|---|---|---|
| CA-19.1 Se pueden cambiar las fechas de un proyecto que ya las tiene. | Cada fecha se podía volver a registrar por separado: el inicio con `PATCH …/fecha-inicio` (RN-18.4) y el fin con `PATCH …/fecha-fin` (RN-41.6 y D4 de HU-41). | Las dos juntas, en una operación |
| CA-19.2 Se vuelve a validar que la fecha de fin sea posterior a la de inicio. | Cada fecha se validaba contra la otra **guardada** (RN-41.5, `ErrEndNotAfterStart`, y RN-41.7, `ErrStartNotBeforeEnd`). | El par nuevo se valida entre sí |
| CA-19.3 Modificar las fechas de un proyecto inexistente devuelve error. | Las dos operaciones devolvían `ErrProjectNotFound` (404). | El mismo error en la operación nueva |

Con las dos operaciones separadas, mover las dos fechas exigía un orden: para postergar, primero el fin; para adelantar, primero el inicio. Además, entre los dos pedidos el proyecto quedaba con una fecha nueva y otra vieja. La prueba de punta a punta muestra el caso (paso de control con HU-18).

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test |
|---|---|
| CA-19.1 — las dos fechas se modifican juntas, al postergar o adelantar, sin cambiar los otros datos, en cualquier estado | “Postergar las dos fechas de un proyecto”; “Adelantar las dos fechas de un proyecto”; “Modificar las fechas de un proyecto finalizado”; `TestSetDates_PostergaLasDosFechasYConservaLosDemasDatos`; `TestSetDates_AdelantaLasDosFechas`; `TestSetDates_SeModificaEnCualquierEstado`; `TestSetDates_CambiaSoloUnaDeLasDosFechas`; `TestSetDates_MismasFechasEsValido`; `TestSetDates_AceptaElDiaSiguienteYElAnioBisiesto`; `TestModificarFechasProyecto_Devuelve200ConElProyecto` (postergar y adelantar) |
| CA-19.2 — el fin nuevo tiene que ser posterior al inicio nuevo, sin comparar con las fechas guardadas | “Rechazar un fin igual o anterior al inicio nuevo” (esquema, 3 ejemplos); `TestSetDates_FinIgualOAnteriorAlInicioDevuelveErrEndNotAfterStart`; `TestSetDates_NoComparaConLasFechasGuardadas`; `TestModificarFechasProyecto_FinNoPosteriorAlInicioDevuelve409` |
| CA-19.3 — proyecto inexistente (incluidos 0 y negativo) | “Rechazar un proyecto inexistente” (esquema con 999, 0 y -1); `TestSetDates_ProyectoInexistenteDevuelveErrProjectNotFound`; `TestSetDates_ParInvalidoEnProyectoInexistenteDevuelveErrProjectNotFound`; `TestModificarFechasProyecto_InexistenteDevuelve404` |
| CA-19.4 — solo se modifican fechas ya registradas | “Rechazar un proyecto sin las dos fechas” (esquema: sin fechas y solo con inicio); `TestSetDates_ProyectoSinLasDosFechasDevuelveErrDatesNotRegistered` (sin fechas, solo con inicio y solo con fin, con un par válido y uno inválido); `TestModificarFechasProyecto_SinLasDosFechasDevuelve409` |
| CA-19.5 — las dos fechas son obligatorias y con formato `AAAA-MM-DD` | “Rechazar una fecha ausente” (esquema, 3 ejemplos); `TestSetDates_FechaDeInicioAusenteDevuelveErrMissingStartDate`; `TestSetDates_FechaDeFinAusenteDevuelveErrMissingEndDate`; `TestSetDates_FechasAusentesSeInformanSinConsultarElRepositorio`; `TestModificarFechasProyecto_FechaDeInicioAusenteDevuelve400`; `TestModificarFechasProyecto_FechaDeFinAusenteDevuelve400`; `TestModificarFechasProyecto_FormatoOFechaInvalidaDevuelve400`; `TestModificarFechasProyecto_FechasVaciasEnProyectoInexistenteDevuelve400` |
| CA-19.6 — todo o nada, falla del repositorio y copia por valor | “Error del repositorio”; “Modificar el proyecto devuelto no altera el guardado”; `TestSetDates_ErrorDelRepositorioAlActualizar`; `TestSetDates_ErrorDelRepositorioAlBuscar`; `TestSetDates_DevuelveUnaCopia`; `TestModificarFechasProyecto_ErrorInternoDevuelve500` |
| CA-19.7 — concurrencia | “Modificaciones simultáneas mantienen el orden de las fechas”; `TestSetDates_ModificacionesSimultaneasMantienenElOrdenDeLasFechas`; `TestSetDates_LasOperacionesDeHU18YHU41ValidanContraLasFechasNuevas` (caso límite de la SDD) |
| CA-19.8 — API e interfaz web | Los 11 tests de `internal/httpapi/project_dates_test.go` (incluidos `TestModificarFechasProyecto_JSONInvalidoDevuelve400` y `TestModificarFechasProyecto_IDNoNumericoDevuelve400`); caso “modificar fechas” de `TestRutasConID_IDNoNumericoDevuelve400`; endpoint “modificar fechas” de `bodyEndpoints` (413 en `TestHandler_CuerpoMayorA1MiBDa413` y texto después del JSON en `TestHandler_TextoDespuesDelJSONDa400`); `TestInterfaz_FormularioModificarFechasDelProyecto`; `TestInterfaz_ModificarFechasLlamaALaAPIConElProyectoActual`; `TestInterfaz_ModificarFechasMuestraElMensajeDeExito`; `TestInterfaz_DatosDelProyectoPrecarganElFormularioModificarFechas` |

El archivo BDD tiene **10 bloques de escenario**: 6 escenarios y 4 esquemas con 11 ejemplos en total, que godog ejecuta como **17 escenarios**. Las pruebas son **35 funciones de test nuevas**:

- 19 unitarias de `internal/project/dates_test.go`;
- 1 runner de godog;
- 11 HTTP;
- 4 de interfaz web.

En tests que ya existían hay 2 casos de tabla nuevos en `server_test.go`: el de `TestRutasConID_IDNoNumericoDevuelve400` y el endpoint de `bodyEndpoints`, que corre en dos tests. No se modificaron los tests de HU-18 ni de HU-41.

La correspondencia no es 1 a 1: 35 funciones de test cubren 10 bloques de escenario (17 escenarios ejecutados). CA-19.8 no tiene escenario BDD, porque la API y la interfaz se prueban con los tests de `internal/httpapi`, y el formato `AAAA-MM-DD` de CA-19.5 solo se prueba en la API.

Conteos con `grep` sobre los archivos de la rama (contra `7c80e8a`, el commit anterior al RED, para los archivos que ya existían; `doc-conteos.txt`):

```text
$ grep -c '^\s*Escenario:' features/HU-19-modificar-fechas-proyecto.feature
6
$ grep -c '^\s*Esquema del escenario:' features/HU-19-modificar-fechas-proyecto.feature
4
$ grep -cE '^\s+\| ' features/HU-19-modificar-fechas-proyecto.feature
15
$ grep -c '^func Test' internal/project/dates_test.go
19
$ grep -c '^func Test' internal/project/dates_features_test.go
1
$ grep -c '^func Test' internal/httpapi/project_dates_test.go
11
$ git diff 7c80e8a 98aa1cf -- internal/httpapi/web_test.go | grep -c '^+func Test'
4
$ git diff 7c80e8a 98aa1cf -- internal/httpapi/server_test.go | grep -c '^+func Test'
0
$ git diff 7c80e8a 98aa1cf -- internal/httpapi/server_test.go | grep -cE '^\+.*"modificar fechas"'
2
$ git diff --stat 7c80e8a 98aa1cf -- internal/project/start_date_test.go internal/project/start_date_features_test.go internal/project/end_date_test.go internal/project/end_date_features_test.go internal/httpapi/project_end_date_test.go
(sin salida)
```

Las 15 filas de tabla son las 4 cabeceras y los 11 ejemplos de los esquemas.

Bloques de escenario por criterio (`grep` del CA en el nombre de cada escenario o esquema): CA-19.1: 3; CA-19.2: 1; CA-19.3: 1; CA-19.4: 1; CA-19.5: 1; CA-19.6: 2; CA-19.7: 1; CA-19.8: 0.

## BDD

Archivo: `features/HU-19-modificar-fechas-proyecto.feature`, con el runner `TestFeatures_HU19ModificarFechasProyecto` en `internal/project/dates_features_test.go`. Todos los pasos llaman al `Manager` de `internal/project`, no al handler HTTP. Las fechas de partida se registran con `SetStartDate` y `SetEndDate` (HU-18 y HU-41). El repositorio en memoria va envuelto en `updateFailingRepository` (de los tests de HU-18, sin modificarlo), que puede fallar al actualizar.

El escenario de concurrencia parte de inicio 2026-10-01 y fin 2026-12-31 y corre en simultáneo, 20 veces cada uno:

- el par 2026-10-10 / 2026-12-20 (siempre válido);
- el par 2027-02-01 / 2027-01-01 (fin anterior al inicio);
- el inicio 2026-10-05 (siempre válido) y el inicio 2027-01-15 (siempre posterior al fin);
- el fin 2026-12-25 (siempre válido) y el fin 2026-09-01 (siempre anterior al inicio);
- consultas.

Con el inicio en 2026-10-01, 2026-10-05 o 2026-10-10 y el fin en 2026-12-20, 2026-12-25 o 2026-12-31, el resultado de cada operación no depende del orden, así que el escenario puede exigir el error exacto de cada una y que el fin final sea posterior al inicio. El test unitario equivalente suma `Update` y verifica que las fechas finales sean las de alguna operación válida.

Resultado de godog con base, sobre el GREEN (`green-godog.txt`) y después del merge con HU-15 (`merge15-godog.txt`), con las mismas cifras:

```text
TestFeatures_HU19ModificarFechasProyecto:
17 scenarios (17 passed)
90 steps (90 passed)
TestFeatures_HU18FechaInicioProyecto:
12 scenarios (12 passed)
53 steps (53 passed)
TestFeatures_HU41FechaFinProyecto:
19 scenarios (19 passed)
101 steps (101 passed)
```

HU-18 y HU-41 dan los mismos escenarios y pasos que en el RED.

## Fase RED

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `1a9ddec` | `docs(HU-19): agregar especificación SDD` |
| Merge de `main` | `7c80e8a` | `chore(HU-19): traer main con HU-40, HU-24, TEC-06 y HU-14` |
| Aprobación de la SDD | `9b4e124` | `docs(HU-19): registrar la aprobación de la SDD` |
| RED | `98aa1cf` | `test(HU-19): agregar tests (RED)` |

El merge `7c80e8a` trajo `main` en `af48b68` (HU-40, HU-24, TEC-06 y HU-14) antes de la aprobación. No tuvo conflictos, `go build ./...` terminó sin errores y `go test ./... -count=1` con base dio `ok` en los 13 paquetes, con 0 SKIP. No entró ninguna ruta, contrato JSON de proyecto ni cambio en la sección Proyectos que afectara la SDD.

Peñalbé Hernán aprobó la SDD el 29/10/2026 con D1 a D7 como estaban y 2 Story Points, y pidió actualizar D0 (HU-19 se tomó con capacidad sobrante y él la asignó a Sosa Ricardo) y D8 (`main` ya traído). Las dos actualizaciones se aplicaron en `9b4e124`.

Antes de los tests se agregaron solo los stubs que permite la SDD: `ErrDatesNotRegistered` y la firma vacía de `SetDates`, que devolvía `Project{}, nil`. **La ruta, el handler, `statusFor` y el formulario web no estaban en el RED**. `gofmt -l .` informó `internal/httpapi/project_dates_test.go` por la alineación de un mapa; se corrigió con `gofmt -w` y después `gofmt -l .` no informó nada. `go vet ./...` y `go build ./...` terminaron sin errores.

### Resultado

Con `go test ./... -count=1 -v`, sin base y con `TEST_DATABASE_URL` en el puerto 5434 (`red-sin-base.log` y `red-con-base.log`):

| Corrida | Paquetes en FAIL | Tests de primer nivel en FAIL | Subtests en FAIL | PASS (todos los niveles) | SKIP |
|---|---|---|---|---|---|
| Sin base | `internal/httpapi` e `internal/project` | 38 | 75 | 2113 | 37 |
| Con base | `internal/httpapi` e `internal/project` | 38 | 75 | 2224 | 0 |

- No hubo ninguna línea con `build failed`, `undefined:`, `cannot use` ni `[setup failed]`, y ningún `panic:`.
- BDD: `17 scenarios (17 failed)`, `90 steps (56 passed, 17 failed, 17 skipped)`.

| Tests que fallaron | Cantidad | Motivo |
|---|---|---|
| `TestSetDates_*` | 19 | El stub devuelve `nil` y no guarda nada: por ejemplo `err = <nil>, se esperaba la fecha de fin debe ser posterior a la de inicio`, `err = <nil>, se esperaba el proyecto no tiene registradas las fechas de inicio y de fin`, `fechas guardadas: inicio … y fin …, se esperaban … y …` o `estado = "", se esperaba "EN_CURSO"`; en el test de concurrencia, `SetDates(2027-02-01, 2027-01-01): err = <nil>, se esperaba la fecha de fin debe ser posterior a la de inicio` |
| `TestFeatures_HU19ModificarFechasProyecto` | 1 | Los 17 escenarios fallan, por la misma causa |
| Los 11 de `project_dates_test.go` | 11 | Sin la ruta, `PATCH /api/proyectos/{id}/fechas` responde `405 Method Not Allowed` en `text/plain`: por ejemplo `código = 405, se esperaba 400; cuerpo: Method Not Allowed` y `Content-Type = "text/plain; charset=utf-8", se esperaba application/json` |
| `TestRutasConID_IDNoNumericoDevuelve400`, `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400` | 3 | Ya existían; fallan solo por su subtest nueva `modificar_fechas` (1 cada uno), que recibió 405. Sus demás subtests pasaron |
| Los 4 de `web_test.go` | 4 | `index.html` y `app.js` no tienen el formulario, el manejador ni la precarga: `la sección de proyectos no contiene el formulario "Modificar fechas"`, `app.js no maneja el envío del formulario "Modificar fechas"` y `mostrarProyecto no contiene "$(\"form-modificar-fechas-proyecto\").fecha_inicio.value = p.fecha_inicio"` |

Ningún otro test que ya existía falló.

**Ningún test ni subtest de HU-19 pasó en el RED.** Con `grep` sobre `red-con-base.log`, 0 líneas `--- PASS` de `TestSetDates_`, `TestModificarFechasProyecto_`, los 4 tests de interfaz, el runner de godog y los subtests `modificar_fechas` de las tablas compartidas, y 55 subtests de `TestSetDates_` y `TestModificarFechasProyecto_` en FAIL.

**Por qué los tests de interfaz revisan solo el bloque de HU-19.** `app.js` ya contenía, por HU-18 y HU-41, cadenas como `fecha_fin: form.fecha_fin.value` y `fecha_inicio: form.fecha_inicio.value`. Si los tests buscaran en todo el archivo, pasarían en el RED sin que existiera el formulario nuevo (falsos positivos). Por eso recortan el formulario `form-modificar-fechas-proyecto` de `index.html` (hasta su `</form>`), el manejador `$("form-modificar-fechas-proyecto").addEventListener("submit"` de `app.js` (hasta su cierre) y el cuerpo de `mostrarProyecto`, y buscan las cadenas solo ahí.

## Fase GREEN

| Fase | Commit | Mensaje |
|---|---|---|
| GREEN | `7553296` | `feat(HU-19): modificar las fechas del proyecto (GREEN)` |
| Refactor | — | No aplica |

El GREEN tocó solo 6 archivos de producción (+115/−9):

- `internal/project`: `dates.go` (`SetDates`) y `doc.go` (mención de HU-19);
- `internal/httpapi`: `projects.go` (`datesRequest` y el handler `setProjectDates`, que lee primero `fecha_inicio` y después `fecha_fin` con `parseDate`) y `server.go` (la ruta y `ErrDatesNotRegistered` en el 409 de `statusFor`);
- `internal/web`: `index.html` (el formulario debajo de "Fecha de fin") y `app.js` (el manejador y la precarga en `mostrarProyecto`).

**Ningún archivo `*_test.go` ni el `.feature` cambió en el GREEN**: `git diff --stat 98aa1cf -- '*_test.go' features/` salió vacío antes del commit. Sin cambios de esquema ni de repositorios, así que la base no se recreó en el GREEN.

### Resultado

`gofmt -l .` no informó nada, y `go vet ./...` y `go build ./...` terminaron sin errores (`green-checks.txt`). Con `go test ./... -count=1 -v`, contando con `grep -c` las líneas `--- PASS`, `--- FAIL` y `--- SKIP`:

| Corrida | Paquetes | PASS de primer nivel | PASS (todos los niveles) | FAIL | SKIP | Salida |
|---|---|---|---|---|---|---|
| Sin base | todos `ok` | 968 | 2226 | 0 | 37 | `green-sin-base.log` |
| Con base (`TEST_DATABASE_URL` en el puerto 5434) | todos `ok` | 1005 | 2337 | 0 | 0 | `green-con-base.log` |

Los 38 tests de primer nivel que fallaron en el RED pasaron en las dos corridas: 38 de 38 (`green-red38.txt`).

## Pruebas de mutación

Antes de mutar, se copiaron los 6 archivos modificados del GREEN a una carpeta de respaldo y se guardó `git diff > green.patch`. Cada mutación se aplicó con un script sobre el GREEN sin commitear, se corrió `go test ./... -count=1 -v` con `TEST_DATABASE_URL` en el puerto 5434 y se revirtió **copiando los archivos desde el respaldo**, sin `git checkout --` ni `git restore`. Después de cada reversión se comprobó con `cmp` que `git diff` fuera idéntico a `green.patch` (las seis veces lo fue). Salidas: `mutaciones.txt` y `mut-<letra>.log`.

| Mutación | Resultado | Tests de primer nivel que fallan |
|---|---|---|
| (a) `dates.go`: comparar el fin nuevo con el inicio guardado (`!end.After(current.StartDate)`) | Detectada (6) | `TestSetDates_AdelantaLasDosFechas` (`SetDates() devolvió un error inesperado: la fecha de fin debe ser posterior a la de inicio`); `TestSetDates_FinIgualOAnteriorAlInicioDevuelveErrEndNotAfterStart`; `TestSetDates_ModificacionesSimultaneasMantienenElOrdenDeLasFechas`; `TestModificarFechasProyecto_Devuelve200ConElProyecto` (`código = 409, se esperaba 200`); `TestModificarFechasProyecto_FinNoPosteriorAlInicioDevuelve409` (`código = 200, se esperaba 409`); runner godog |
| (b) `dates.go`: quitar la validación de "ya tiene las dos fechas" | Detectada (3) | `TestSetDates_ProyectoSinLasDosFechasDevuelveErrDatesNotRegistered` (`err = <nil>, se esperaba el proyecto no tiene registradas las fechas de inicio y de fin`); `TestModificarFechasProyecto_SinLasDosFechasDevuelve409` (`código = 200, se esperaba 409`); runner godog |
| (c) `dates.go`: aceptar fin igual a inicio (`!end.After(start)` → `end.Before(start)`) | Detectada (3) | `TestSetDates_FinIgualOAnteriorAlInicioDevuelveErrEndNotAfterStart` (`err = <nil>, se esperaba la fecha de fin debe ser posterior a la de inicio`); `TestModificarFechasProyecto_FinNoPosteriorAlInicioDevuelve409` (`código = 200, se esperaba 409`); runner godog |
| (d) `server.go`: `ErrDatesNotRegistered` en el 400 | Detectada (1) | `TestModificarFechasProyecto_SinLasDosFechasDevuelve409` (`código = 400, se esperaba 409`) |
| (e) `dates.go`: guardar solo la fecha de inicio (sin `current.EndDate = end`) | Detectada (10) | `TestModificarFechasProyecto_Devuelve200ConElProyecto`; runner godog; 8 `TestSetDates_*` (por ejemplo `fechas guardadas: inicio … y fin …, se esperaban … y …`) |
| (f) `dates.go`: quitar el mutex de `SetDates` | **Sobrevive** | `ok` en todos los paquetes, 0 FAIL |

**Por qué sobrevive la (f).** El test unitario y el escenario de concurrencia usan fechas elegidas para que cualquier orden de las operaciones deje un par válido, así que, sin el mutex, los resultados y el estado final siguen siendo los esperados. Lo que la mutación rompe es otra cosa: `MemoryRepository` no es seguro para uso concurrente ("se accede a través del Manager, que lo protege"), y sin el mutex `SetDates` lee y escribe `r.projects` en paralelo con `SetStartDate`, `SetEndDate`, `Update` y `Get`. Es un acceso sin protección que detecta `go test -race`, que corre el CI del Pull Request. **`-race` no corre localmente**: `go env CGO_ENABLED` da `0` y no hay `gcc` en el PATH, y no hay una imagen de Go local en Docker. La detección en el CI está razonada, no verificada localmente. No se cambiaron los tests.

Después de las mutaciones se repitió la verificación completa (`post-mut-checks.txt`, `post-mut-sin-base.log` y `post-mut-con-base.log`): `gofmt -l .` sin salida, `go vet` y `go build` sin errores, y las mismas cifras que la primera (968/2226/0/37 sin base y 1005/2337/0/0 con base).

## Prueba de punta a punta de la API

El servidor real se compiló desde el GREEN con `go build -o server-hu19.exe ./cmd/server`, fuera del repositorio, y se arrancó en memoria, sin `STORAGE`, con `PORT=18919` (log: `servidor escuchando en :18919 (almacenamiento: memory)`). Los pedidos se hicieron con `curl` desde Git Bash. Salida: `e2e-curl.txt`.

| Paso | Pedido | Código | Cuerpo |
|---|---|---|---|
| Crear el proyecto | `POST /api/proyectos` `{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas"}` | 201 | el proyecto, con `"fecha_inicio":""` y `"fecha_fin":""` |
| Modificar sin fechas registradas | `PATCH /api/proyectos/1/fechas` `{"fecha_inicio":"2026-11-05","fecha_fin":"2026-11-30"}` | 409 | `{"error":"el proyecto no tiene registradas las fechas de inicio y de fin"}` |
| Registrar el inicio (HU-18) | `PATCH /api/proyectos/1/fecha-inicio` `{"fecha_inicio":"2026-10-01"}` | 200 | `"fecha_inicio":"2026-10-01","fecha_fin":""` |
| Registrar el fin (HU-41) | `PATCH /api/proyectos/1/fecha-fin` `{"fecha_fin":"2026-10-31"}` | 200 | `"fecha_inicio":"2026-10-01","fecha_fin":"2026-10-31"` |
| **Control:** postergar el inicio primero, con HU-18 | `PATCH /api/proyectos/1/fecha-inicio` `{"fecha_inicio":"2026-11-05"}` | 409 | `{"error":"la fecha de inicio debe ser anterior a la de fin"}` |
| Postergar las dos con HU-19 | `PATCH /api/proyectos/1/fechas` `{"fecha_inicio":"2026-11-05","fecha_fin":"2026-11-30"}` | 200 | `"fecha_inicio":"2026-11-05","fecha_fin":"2026-11-30"` |
| Adelantar las dos | `PATCH /api/proyectos/1/fechas` `{"fecha_inicio":"2026-09-01","fecha_fin":"2026-09-20"}` | 200 | `"fecha_inicio":"2026-09-01","fecha_fin":"2026-09-20"` |
| Fin igual al inicio | `PATCH /api/proyectos/1/fechas` `{"fecha_inicio":"2026-09-10","fecha_fin":"2026-09-10"}` | 409 | `{"error":"la fecha de fin debe ser posterior a la de inicio"}` |
| Una sola fecha | `PATCH /api/proyectos/1/fechas` `{"fecha_inicio":"2026-10-01"}` | 400 | `{"error":"la fecha de fin del proyecto es obligatoria"}` |
| Formato inválido | `PATCH /api/proyectos/1/fechas` `{"fecha_inicio":"01/10/2026","fecha_fin":"2026-10-31"}` | 400 | `{"error":"las fechas deben tener el formato AAAA-MM-DD"}` |
| Proyecto inexistente | `PATCH /api/proyectos/999/fechas` `{"fecha_inicio":"2026-10-01","fecha_fin":"2026-10-31"}` | 404 | `{"error":"no existe un proyecto con ese id"}` |
| Consulta final | `GET /api/proyectos/1` | 200 | `{"id":1,"nombre":"Sistema de ventas",…,"estado":"PLANIFICADO","fecha_inicio":"2026-09-01","fecha_fin":"2026-09-20"}` |

El paso de control muestra el problema que resuelve la historia: con las operaciones de HU-18 y HU-41, postergar el proyecto registrando primero el inicio nuevo da 409, porque se compara con el fin guardado. Con `PATCH …/fechas`, el mismo cambio da 200 en un solo pedido, y el estado y la descripción del proyecto no cambian.

En las respuestas, la "ó" de "Gestión" aparece rota (`Gesti�n`). Es por la codificación con la que Git Bash pasó el argumento a `curl`, no por el servidor, y no afecta las fechas.

Al terminar, detener la tarea del servidor no terminó el proceso (`tasklist` todavía mostraba `server-hu19.exe`). Se bajó con `taskkill //PID 8276 //F`, y un `curl` posterior a `localhost:18919` dio código `000`: el puerto quedó cerrado.

La interfaz web no se recorrió en un navegador. La cubren los 4 tests de `internal/httpapi/web_test.go`, que verifican el formulario, el manejador de `app.js`, el mensaje y la precarga, pero no ejecutan el JavaScript.

## Merge de `main` con HU-15

Antes de la documentación, `origin/main` pasó de `af48b68` a `656ac18` (HU-15, PR #155) y se trajo a la rama en `4d92c68` (`chore(HU-19): traer main con HU-15`).

- **3 archivos en conflicto, resueltos conservando los dos lados.** En cada uno, HU-19 y HU-15 agregaban código en el mismo lugar; se dejó primero lo de HU-19 y después lo de HU-15, sin cambiar el comportamiento de ninguna (`merge15-resolucion.diff`):
  - `internal/httpapi/server.go`: la ruta `PATCH /api/proyectos/{id}/fechas` y la ruta `POST /api/proyectos/{id}/integrantes`;
  - `internal/httpapi/server_test.go`: los casos “modificar fechas” y “registrar integrante” de `TestRutasConID_IDNoNumericoDevuelve400`;
  - `internal/web/app.js`: el manejador del formulario "Modificar fechas" y la sección "Integrantes" de HU-15 (con `nombresDeRoles`).

  `web_test.go` e `index.html` se combinaron sin conflicto.
- **Ajuste por la firma nueva de `NewHandler`.** HU-15 le agregó el parámetro `members *member.Manager`. En `TestModificarFechasProyecto_ErrorInternoDevuelve500` se agregó `newMembers(projects)` a la llamada, igual que hizo HU-15 en el test equivalente de HU-41 (`TestRegistrarFechaFinProyecto_ErrorInternoDevuelve500`). Es el único cambio en un `_test.go` de HU-19 y no cambia ninguna aserción (`merge15-ajuste.diff`).
- **Recreación de la base.** Como HU-15 agregó la tabla `members`, la base se recreó con `POSTGRES_PORT=5434 docker compose down -v` y `POSTGRES_PORT=5434 docker compose up -d --wait` (contenedor `Healthy`). Con `psql` se verificó que `members` y `projects.end_date` existen en `gpc` y en `gpc_test` (`merge15-db.txt`).
- **Verificación completa después del merge.** `gofmt -l .` sin salida, y `go vet ./...` y `go build ./...` sin errores (`merge15-checks.txt`). Con `go test ./... -count=1 -v`, todos los paquetes `ok`, incluido el nuevo `internal/member` (`merge15-sin-base.log` y `merge15-con-base.log`):

| Corrida | PASS de primer nivel | PASS (todos los niveles) | FAIL | SKIP |
|---|---|---|---|---|
| Sin base | 992 | 2295 | 0 | 40 |
| Con base | 1032 | 2414 | 0 | 0 |

Los 3 SKIP más que en el GREEN sin base son los 3 tests de PostgreSQL de HU-15 (`internal/member/postgres_repository_test.go`). Godog de HU-19, HU-18 y HU-41 dio las mismas cifras que en el GREEN, y los 38 tests del RED siguieron en PASS con base (38 de 38).

## Validación final

Medido sobre `4d92c68` (la rama con `main` `656ac18` adentro) con go1.27.1 windows/amd64, en Windows.

### `go test -count=1 -cover ./...` sin base

Salida guardada en `rama-cover-sin-base.txt`:

```text
ok  	seguimiento-medicion-gpc/cmd/server	4.710s	coverage: 29.8% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	6.339s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/dashboard	4.793s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/defect	8.719s	coverage: 77.7% of statements
ok  	seguimiento-medicion-gpc/internal/effort	5.844s	coverage: 75.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	8.203s	coverage: 98.6% of statements
ok  	seguimiento-medicion-gpc/internal/member	2.872s	coverage: 60.3% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	3.466s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	6.082s	coverage: 75.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	4.346s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	4.399s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	5.120s	coverage: 74.8% of statements
ok  	seguimiento-medicion-gpc/internal/report	3.978s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	4.948s	coverage: 77.1% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

Con `-coverprofile` y `go tool cover -func`, sin base (`rama-func-sin-base.txt`):

- en 100.0%: `SetDates`, `setProjectDates`, `toProjectResponse`, `statusFor` y `NewHandler`;
- total **84.6%**.

Para comparar con `main` se usó `656ac18`, medido con el mismo Go y los mismos comandos en una copia temporal (`git worktree` en el scratchpad, borrada al terminar; salidas `main-cover-sin-base.txt` y `main-func-sin-base.txt`):

- `internal/project` sube de 71.1% a 74.8%;
- los demás paquetes se mantienen, incluido `internal/httpapi` en 98.6%;
- el total del perfil sube de **84.3%** a **84.6%**.

### Con PostgreSQL

Con PostgreSQL en Docker en el puerto 5434, recreado con la tabla `members`, y `TEST_DATABASE_URL=postgres://gpc:gpc_dev@localhost:5434/gpc_test?sslmode=disable`:

- `go test -count=1 -v ./...`: **0 SKIP, 2414 PASS y 0 FAIL** (`merge15-con-base.log`).
- `go test -count=1 -cover ./...` (`rama-cover-con-base.txt`): `internal/project` 98.1% (97.8% en `main` `656ac18`), `internal/httpapi` 98.6% y `cmd/server` 31.6% (los dos iguales en `main`).
- `go test -count=1 -coverpkg=./... -coverprofile=… ./...` (`rama-func-coverpkg.txt`): total **96.1%**, contra **96.0%** de `main` `656ac18` con los mismos comandos (`main-func-coverpkg.txt`). `SetDates` y `setProjectDates` quedan en 100.0%.

### `-race`

`-race` no se corrió localmente: `go env CGO_ENABLED` da `0` y no hay `gcc`. Los tests de concurrencia (`TestSetDates_ModificacionesSimultaneasMantienenElOrdenDeLasFechas` y el escenario de CA-19.7) corrieron sin `-race` y pasaron. Con `-race` los ejecuta el CI del Pull Request (`go test -race -coverprofile=coverage.out ./...`), que es el que puede detectar la mutación (f).

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `1a9ddec` | `docs(HU-19): agregar especificación SDD` |
| Merge de `main` | `7c80e8a` | `chore(HU-19): traer main con HU-40, HU-24, TEC-06 y HU-14` |
| Aprobación de la SDD | `9b4e124` | `docs(HU-19): registrar la aprobación de la SDD` |
| RED | `98aa1cf` | `test(HU-19): agregar tests (RED)` |
| GREEN | `7553296` | `feat(HU-19): modificar las fechas del proyecto (GREEN)` |
| Merge de `main` | `4d92c68` | `chore(HU-19): traer main con HU-15` |
| Refactor | — | No aplica |
| Documentación | (propuesto) | `docs(HU-19): agregar evidencia TDD y cobertura` |

Todos los commits referencian el Issue #37 con `Refs #37`, incluidos los dos merges de `main`. El de cierre lo cerrará con `Closes #37`. El mensaje del merge `4d92c68` conserva la lista `# Conflicts:` de los 3 archivos que git agrega con `--no-edit`.

HU-19: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
