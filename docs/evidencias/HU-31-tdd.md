# Evidencia TDD · HU-31

> Issue del repositorio: #49

## Historia

Como Agile Enabler, quiero consultar los defectos de un proyecto o de un sprint, para evaluar su calidad.

## Resultado

HU-31 quedó implementada en la rama `feature/HU-31-consultar-defectos`. En `internal/defect/list.go`:

- `Manager.ListByProject(projectID)` devuelve los defectos de las historias del proyecto, en cualquier estado y ordenados por ID. Obtiene las historias con `backlog.Manager.ListItems`, que valida el proyecto (`project.ErrProjectNotFound`). Es la operación, con la firma exacta, que espera la SDD de HU-47.
- `Manager.ListBySprint(sprintID)` devuelve los defectos **detectados** en el sprint (`DetectionSprintID`, la definición de M9). Valida el sprint con `sprint.Manager.Get` (`sprint.ErrSprintNotFound`) y acepta sprints abiertos y cerrados.
- Las dos toman el mutex del `Manager` de defectos y comparten `selectDefects`, que devuelve una lista vacía y no `nil` cuando no hay coincidencias.
- `ParseStatus` reutiliza `normalizeStatus` de HU-30 (`ErrInvalidStatus`). `FilterByStatus` devuelve una lista nueva con los defectos de un estado.

`Repository` suma `List()`, implementado en memoria (bajo su mutex) y en PostgreSQL (`SELECT … FROM defects ORDER BY id`), sin cambios de esquema.

La API lo expone en `GET /api/proyectos/{id}/defectos` y `GET /api/sprints/{id}/defectos`, con `?estado=` opcional. La respuesta es `{"defectos": [...], "cantidad": N}`. El handler recorta `?estado=` con `strings.TrimSpace`: si queda vacío no filtra (ajuste de Hernán en D3) y, si no, valida con `ParseStatus` antes de consultar. No se agregaron errores nuevos a `statusFor`. La interfaz web suma, en la sección Defectos y debajo de "Cambiar estado de defecto", el formulario "Consultar defectos" con la tabla y la cantidad.

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test |
|---|---|
| CA-31.1 — todos los defectos de un proyecto o los detectados en un sprint, en cualquier estado, con todos sus datos, ordenados por ID | “Listar los defectos de un proyecto en todos sus estados”; “Listar los defectos detectados en un sprint”; `TestListByProject_DevuelveTodosLosDefectosEnCualquierEstadoOrdenadosPorID`; `TestListBySprint_DevuelveLosDefectosDetectadosEnElSprintEnCualquierEstado`; `TestMemoryRepository_ContratoList` y `TestPostgresRepository_ContratoList` (3 subtests de contrato cada uno); `TestPostgresRepository_ListLeeLaTablaOrdenadaPorID` |
| CA-31.2 — aislamiento por proyecto y por sprint de detección; sprints cerrados consultables | “No mezclar defectos de otro proyecto”; “Un defecto resuelto en otro sprint aparece solo en el de detección”; “Consultar un sprint cerrado”; `TestListByProject_NoIncluyeDefectosDeOtroProyecto`; `TestListBySprint_NoIncluyeDefectosDetectadosEnOtroSprintAunqueSeResolvieranEnEl`; `TestListBySprint_NoIncluyeDefectosDeSprintsDeOtroProyecto`; `TestListBySprint_SprintCerradoSeConsulta` |
| CA-31.3 — filtro por estado normalizado; estado inválido rechazado antes de consultar | “Filtrar los defectos por estado” (esquema, 4 ejemplos); “Normalizar el estado del filtro con espacios en los extremos”; “Rechazar un estado de filtro inválido” (esquema, 3 ejemplos, uno sobre un proyecto inexistente); `TestParseStatus_NormalizaLosEstadosValidos`; `TestParseStatus_RechazaEstadosInvalidos`; `TestFilterByStatus_DevuelveSoloLosDelEstadoEnElMismoOrden`; `TestFilterByStatus_NoModificaLaListaRecibida` |
| CA-31.4 — lista vacía (no `nil`, `[]` en JSON) sin defectos o sin coincidencias | “Proyecto y sprint sin defectos devuelven una lista vacía”; “Proyecto sin historias devuelve una lista vacía”; “Filtro sin coincidencias devuelve una lista vacía”; `TestListByProject_ProyectoSinHistoriasDevuelveListaVacia`; `TestListByProject_ProyectoSinDefectosDevuelveListaVacia`; `TestListBySprint_SprintSinDefectosDevuelveListaVacia`; `TestFilterByStatus_SinCoincidenciasDevuelveListaVaciaNoNil` |
| CA-31.5 — proyecto o sprint inexistente (incluidos 0 y negativos) | “Rechazar un proyecto y un sprint inexistentes” (esquema con 999, 0 y -1); `TestListByProject_ProyectoInexistenteDevuelveErrProjectNotFound`; `TestListBySprint_SprintInexistenteDevuelveErrSprintNotFound` (los dos verifican además que no se llamó a `List`) |
| CA-31.6 — copias por valor y solo lectura | “Modificar lo devuelto no altera lo guardado”; `TestList_ModificarLoDevueltoNoAlteraLoGuardado`; `TestList_ConsultarNoModificaNada`; subtest “List devuelve copias” del contrato |
| CA-31.7 — consultas simultáneas con registros y cambios de estado | “Consultas simultáneas con registros y cambios de estado”; `TestList_ConsultasSimultaneasConRegistrosYCambiosDeEstado` |
| CA-31.8 — falla de un repositorio al leer | “Error del repositorio al listar”; `TestListByProject_ErrorDelRepositorioDeDefectos`; `TestListBySprint_ErrorDelRepositorioDeDefectos`; `TestListByProject_ErrorDelBacklog` |
| CA-31.9 — API e interfaz web | Los 8 `TestListarDefectos*` de `internal/httpapi/defect_lists_test.go` (200 por ruta, filtro, filtro vacío o con solo espacios que no filtra, lista vacía, 400, 404 y 500); casos “listar defectos del proyecto” y “listar defectos del sprint” de `TestRutasConID_IDNoNumericoDevuelve400`; `TestInterfaz_FormularioConsultarDefectos`; `TestInterfaz_TablaDeDefectosTieneLasColumnas`; `TestInterfaz_ConsultarDefectosLlamaALaAPI`; `TestInterfaz_FiltroDeEstadoDeDefectosUsaLaListaDelPaquete` |

El archivo BDD tiene **15 bloques de escenario**: 12 escenarios y 3 esquemas con 10 ejemplos en total, que godog ejecuta como **22 escenarios**. Las pruebas son **38 funciones de test nuevas**:

- 22 unitarias de `internal/defect/list_test.go`;
- 1 runner de godog;
- 1 de contrato en memoria y 2 de PostgreSQL;
- 8 HTTP;
- 4 de interfaz web.

Además, cada una de las dos funciones de contrato corre 3 subtests de `runRepositoryListContract`, y `TestRutasConID_IDNoNumericoDevuelve400`, que ya existía, suma 2 casos de tabla.

La correspondencia no es 1 a 1. CA-31.9 no tiene escenario BDD porque la API y la interfaz se prueban con los tests de `internal/httpapi`. Varios escenarios verifican más de una consulta (proyecto y sprint) en el mismo bloque.

Conteos con `grep` sobre los archivos de la rama (contra `main`, `f6e8101`, para los archivos que ya existían):

```text
$ grep -c '^\s*Escenario:' features/HU-31-consultar-defectos.feature
12
$ grep -c '^\s*Esquema del escenario:' features/HU-31-consultar-defectos.feature
3
$ grep -cE '^\s+\| ' features/HU-31-consultar-defectos.feature
13
$ grep -c '^func Test' internal/defect/list_test.go
22
$ grep -c '^func Test' internal/defect/list_features_test.go
1
$ git diff f6e8101 HEAD -- internal/defect/repository_contract_test.go | grep -c '^+func Test'
1
$ git diff f6e8101 HEAD -- internal/defect/postgres_repository_test.go | grep -c '^+func Test'
2
$ git diff f6e8101 HEAD -- internal/httpapi/defect_lists_test.go | grep -c '^+func Test'
8
$ git diff f6e8101 HEAD -- internal/httpapi/web_test.go | grep -c '^+func Test'
4
$ git diff f6e8101 HEAD -- internal/httpapi/server_test.go | grep -c '^+func Test'
0
$ git diff f6e8101 HEAD -- internal/defect/repository_contract_test.go | grep -cE '^\+\s+t\.Run\("'
3
$ git diff f6e8101 HEAD -- internal/httpapi/server_test.go | grep -cE '^\+.*"listar defectos del'
2
```

Las 13 filas de tabla son las 3 cabeceras y los 10 ejemplos de los esquemas.

Bloques de escenario por criterio (`grep` del CA en el nombre de cada escenario o esquema): CA-31.1: 2; CA-31.2: 3; CA-31.3: 3; CA-31.4: 3; CA-31.5: 1; CA-31.6: 1; CA-31.7: 1; CA-31.8: 1; CA-31.9: 0.

## BDD

Archivo: `features/HU-31-consultar-defectos.feature`, con el runner `TestFeatures_HU31ConsultarDefectos` en `internal/defect/list_features_test.go`. Todos los pasos llaman al `Manager` de `internal/defect`, no al handler HTTP.

Los defectos de partida se guardan con `Save` del repositorio en memoria, porque registrarlos es HU-29 y cambiar su estado es HU-30. Los pasos con filtro hacen lo mismo que el handler: validan el estado con `ParseStatus`, consultan y filtran con `FilterByStatus`. El caso del filtro con solo espacios (no filtra) es una decisión del handler (D3), así que se prueba en `TestListarDefectos_FiltroVacioOConSoloEspaciosNoFiltra` y no en godog. El escenario “Normalizar el estado del filtro con espacios en los extremos” sí va en godog, porque los espacios están dentro de las comillas del paso y no en una celda de tabla.

Resultado de `go test -count=1 -v -run 'TestFeatures_HU(29|30|31)' ./internal/defect/` sobre el GREEN (salida guardada en `hu31-green-godog.txt`):

```text
22 scenarios (22 passed)
214 steps (214 passed)
--- PASS: TestFeatures_HU31ConsultarDefectos
16 scenarios (16 passed)
109 steps (109 passed)
--- PASS: TestFeatures_HU29RegistrarDefecto
34 scenarios (34 passed)
276 steps (276 passed)
--- PASS: TestFeatures_HU30CambiarEstadoDefecto
```

## Fase RED

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `2bc0cef` | `docs(HU-31): agregar especificación SDD` |
| Aprobación de la SDD | `6f65298` | `docs(HU-31): registrar la aprobación de la SDD` |
| RED | `b44d97a` | `test(HU-31): agregar tests (RED)` |

Antes de los tests se agregaron solo los stubs que permite D10:

- las firmas vacías de `ListByProject`, `ListBySprint`, `ParseStatus` y `FilterByStatus` en `list.go`, que devolvían `nil, nil`, `"", nil` y `nil`;
- `List()` en la interfaz `Repository` y sus firmas vacías en memoria y PostgreSQL, que devolvían `nil, nil`.

Siguiendo el ajuste que pidió Peñalbé Hernán en HU-18, **las rutas `GET /api/proyectos/{id}/defectos` y `GET /api/sprints/{id}/defectos`, los handlers y el formulario web no estaban en el RED**. `gofmt -l .` y `go vet ./...` no informaron nada antes del commit RED.

**No se tocaron `register_test.go`, `status_test.go` ni `defects_test.go`** (`git diff --stat f6e8101 HEAD` de esos tres archivos sale vacío). Los métodos `List` de los dobles de HU-29 y HU-30 se declararon en los archivos de tests nuevos: los de `testDefectRepository` y `failingDefectRepository` en `list_test.go`, y el de `brokenDefectRepository` en `defect_lists_test.go`. `countingDefectRepository` y `recordingDefectRepository` lo heredan porque embeben otro tipo.

El RED falló por las aserciones y no por compilación, solo en `internal/defect` e `internal/httpapi`. Se contó con `grep` sobre la salida guardada de `go test -count=1 ./...` (`hu31-red-go-test.txt`):

- fallaron **37 tests de primer nivel** (`^--- FAIL`) y **83 subtests** (`^\s+--- FAIL`), incluidos los 22 escenarios godog;
- no hubo ninguna línea con `build failed`, `undefined:`, `cannot use` ni `[setup failed]`.

De los 37, 36 son tests nuevos de HU-31. El otro es `TestRutasConID_IDNoNumericoDevuelve400`, que ya existía y falló solo por sus 2 subtests nuevas; las demás pasaron. Ningún otro test que ya existía en `main` falló.

- Los unitarios y el contrato en memoria fallaron con el resultado de los stubs, por ejemplo `err = <nil>, se esperaba no existe un proyecto con ese id`, `ListByProject devolvió una lista nil, se esperaba […]` o `ParseStatus(" Resuelto ") = "", se esperaba "RESUELTO"`.
- BDD: `22 scenarios (22 failed)`, `214 steps (157 passed, 22 failed, 35 skipped)`.
- Los tests HTTP recibieron el `404 page not found` del servidor de archivos, porque las rutas no existían. **Los de proyecto y sprint inexistente no pasaban en falso aunque esperan un 404**: fallaban por el Content-Type y por el cuerpo, no solo por el código. Por ejemplo:

  ```text
  defect_lists_test.go:222: GET /api/proyectos/999/defectos: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
  defect_lists_test.go:224: cuerpo no es JSON válido (invalid character 'p' after top-level value): 404 page not found
  ```

  Los mismos dos mensajes aparecen en los 6 casos de `TestListarDefectos_InexistenteDevuelve404` (proyecto y sprint con 999, 0 y -1).
- Los 4 tests de interfaz fallaron porque la página y `app.js` no tenían el formulario, la tabla ni los textos.

**Ningún test de HU-31 pasaba en el RED.** Los dos que necesitaban un refuerzo para no pasar en falso ya lo tenían cuando se escribieron:

- el test y el escenario de concurrencia exigen que, al final, `ListByProject` y `ListBySprint` devuelvan los 21 defectos guardados, y no solo que ninguna operación falle;
- el test y el escenario de copias exigen primero que se listen los defectos guardados, antes de modificar lo devuelto.

Sin `TEST_DATABASE_URL`, `TestPostgresRepository_ContratoList` y `TestPostgresRepository_ListLeeLaTablaOrdenadaPorID` hicieron SKIP. **Contra la base (PostgreSQL en el puerto 5434) los dos fallaron** con el stub (`hu31-red-postgres.txt`):

```text
repository_contract_test.go:252: List = []defect.Defect(nil), se esperaba una lista vacía no nil
--- FAIL: TestPostgresRepository_ContratoList (0.54s)
postgres_repository_test.go:233: List = [], se esperaba [{ID:1 Description:Primero …} {ID:2 …} {ID:3 …}]
--- FAIL: TestPostgresRepository_ListLeeLaTablaOrdenadaPorID (0.15s)
```

`TestPostgresRepository_ContratoList` falló en sus 3 subtests.

## Fase GREEN

| Fase | Commit | Mensaje |
|---|---|---|
| GREEN | `c4f07e5` | `feat(HU-31): implementar la consulta de defectos (GREEN)` |
| Refactor | — | No aplica |
| Documentación | (propuesto) | `docs(HU-31): agregar evidencia TDD y cobertura` |

El GREEN tocó solo 7 archivos de producción:

- `internal/defect`: `list.go` (`ListByProject`, `ListBySprint`, `selectDefects`, `ParseStatus` y `FilterByStatus`), `memory_repository.go` y `postgres_repository.go` (`List`);
- `internal/httpapi`: `defects.go` (`listDefectsByProject`, `listDefectsBySprint`, `listDefects`, `defectListResponse` y `toDefectListResponse`) y `server.go` (las dos rutas);
- `internal/web`: `index.html` y `app.js`.

**Ningún archivo `*_test.go` ni el `.feature` cambió en el GREEN**: `git diff --stat HEAD -- '*_test.go' features/` salió vacío antes del commit. Sin base, todos los paquetes dieron `ok` (`hu31-green-go-test.txt`). Los 22 escenarios de HU-31 pasaron (214 pasos), y HU-29 (16 escenarios) y HU-30 (34 escenarios) siguen en verde (`hu31-green-godog.txt`).

Con `-coverprofile` sin base (`hu31-green-cover-func.txt`), todas las funciones nuevas quedaron en 100.0%, salvo `List` de PostgreSQL, en 0.0% porque sus tests se omiten sin `TEST_DATABASE_URL`.

## Pruebas de mutación

Cada mutación se aplicó sobre el GREEN (`c4f07e5`) y se corrió `go test -count=1` sobre `./internal/defect/` y `./internal/httpapi/` (la (f), solo sobre `./internal/httpapi/`). Después se restauró con `git checkout -- <archivo>`, y en todos los casos `git diff` salió vacío. De la (a) a la (f) y la (h) corrieron sin base, y la (g) con `TEST_DATABASE_URL` en el puerto 5434. Salidas: `hu31-mut-<letra>.txt`.

| Mutación | Resultado | Tests que fallan (mensaje) |
|---|---|---|
| (a) `list.go`: `ListByProject` no filtra por las historias del proyecto | Detectada (7 tests) | `TestListByProject_NoIncluyeDefectosDeOtroProyecto`; `TestListByProject_ProyectoSinHistoriasDevuelveListaVacia`; `TestListarDefectosDelProyecto_Devuelve200ConLosDefectosYLaCantidad` (`IDs = [1 2 3 4], se esperaban [1 2 3]`); `TestListarDefectos_FiltraPorEstado` (`IDs = [1 4], se esperaban [1]`); `TestListarDefectos_FiltroVacioOConSoloEspaciosNoFiltra`; `TestListarDefectos_SinDefectosDevuelveListaVaciaYCantidadCero`; runner godog |
| (b) `list.go`: `ListBySprint` filtra por `ResolutionSprintID` | Detectada (9 tests) | Los 4 `TestListBySprint_*` de contenido (detectados, otro sprint, otro proyecto, cerrado); `TestList_ModificarLoDevueltoNoAlteraLoGuardado`; `TestListarDefectosDelSprint_Devuelve200ConLosDefectosYLaCantidad` (`sprint 1: IDs = [3], se esperaban [1 2 3]`); `TestListarDefectos_FiltraPorEstado`; `TestListarDefectos_FiltroVacioOConSoloEspaciosNoFiltra`; runner godog |
| (c) `list.go`: `selectDefects` devuelve `nil` sin coincidencias | Detectada (4 tests) | `TestListByProject_ProyectoSinHistoriasDevuelveListaVacia`, `TestListByProject_ProyectoSinDefectosDevuelveListaVacia` (`ListByProject devolvió una lista nil, se esperaba []`), `TestListBySprint_SprintSinDefectosDevuelveListaVacia`; runner godog. `internal/httpapi` sigue en `ok`, porque `toDefectListResponse` arma su propia lista no `nil` (D6) |
| (d) `list.go`: `ListBySprint` no valida el sprint | Detectada (3 tests) | `TestListBySprint_SprintInexistenteDevuelveErrSprintNotFound` (`err = <nil>, se esperaba no existe un sprint con ese id` y `List se llamó 3 veces, se esperaba ninguna`); `TestListarDefectos_InexistenteDevuelve404` (`código = 200, se esperaba 404; cuerpo: {"defectos":[],"cantidad":0}`); runner godog |
| (e) `list.go`: `FilterByStatus` devuelve la lista recibida | Detectada (6 tests) | Los 3 `TestFilterByStatus_*`; `TestListarDefectos_FiltraPorEstado` (`IDs = [1 2 3], se esperaban [3]`); `TestListarDefectos_SinDefectosDevuelveListaVaciaYCantidadCero`; runner godog |
| (f) `defects.go`: sin `TrimSpace` en `?estado=` | Detectada (1 test) | `TestListarDefectos_FiltroVacioOConSoloEspaciosNoFiltra` (`código = 400, se esperaba 200; cuerpo: {"error":"el estado debe ser ABIERTO, EN_CORRECCION o RESUELTO"}`) |
| (g) `postgres_repository.go`: `List` sin `ORDER BY id` | **Sobrevive** (con base) | `ok`. Con `-v`, `TestPostgresRepository_ContratoList` (3 subtests) y `TestPostgresRepository_ListLeeLaTablaOrdenadaPorID` corrieron contra la base y dieron PASS, sin SKIP (`hu31-mut-g-verbose.txt`) |
| (h) `list.go`: `ListByProject` sin el mutex del `Manager` | **Sobrevive** | `ok` en la corrida normal. Con `-count=50 -v -run 'TestList_ConsultasSimultaneasConRegistrosYCambiosDeEstado$'`: 50 lotes PASS y 0 FAIL (`hu31-mut-h-count50.txt`) |

En la (f) también se quitó el import de `strings`, que quedaba sin usar.

**La mutación (g) sobrevive.** Los tests de `List` guardan pocas filas en una tabla recién vaciada y nunca las actualizan. Sin `ORDER BY`, PostgreSQL las lee en el orden en que están en la tabla, que en ese caso es el de inserción, o sea el de ID. El orden solo cambiaría después de un `UPDATE`, que escribe la nueva versión de la fila al final, o con más datos y otro plan de consulta. No se cambiaron los tests.

**La mutación (h) sobrevive.** Sin el mutex de defectos, `ListByProject` sigue sin accesos sin protección: `ListItems` toma el mutex del backlog y `MemoryRepository.List` el suyo. Lo único que cambia es que un `Register` simultáneo puede quedar fuera o dentro del resultado, y las dos respuestas son válidas, porque el test verifica la cantidad final después de `wg.Wait()`. El mutex da una lectura coherente frente a registros y cambios de estado (RN-31.12) y mantiene el orden de locks de HU-29 y HU-30, pero los tests no pueden observar su ausencia. Tampoco `-race` la detectaría, porque no hay un acceso sin protección. No se cambiaron los tests.

## Prueba de punta a punta de la API

El servidor real se compiló desde `c4f07e5` con `go build -o hu31-server.exe ./cmd/server`, fuera del repositorio. Se arrancó en memoria, sin `STORAGE` ni `TEST_DATABASE_URL`, con `PORT=18931 ./hu31-server.exe` (log: `servidor escuchando en :18931 (almacenamiento: memory)`). Los pedidos se hicieron con `curl` desde Git Bash, con textos de entrada sin tildes porque Git Bash no los envía en UTF-8. Al terminar se detuvo el servidor (`tasklist` no mostró ningún `hu31-server.exe` y el puerto 18931 quedó sin procesos escuchando) y se borró el binario. Salida: `hu31-e2e.txt`.

```
$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18931/api/proyectos -d '{"nombre":"Sistema de ventas","descripcion":""}'
{"id":1,"nombre":"Sistema de ventas","descripcion":"","estado":"PLANIFICADO","fecha_inicio":""}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18931/api/proyectos/1/backlog -d '{"titulo":"Login de usuarios","descripcion":"Acceso","prioridad":"alta","criterios":["Credenciales validas"]}'
{"id":1,"proyecto_id":1,"titulo":"Login de usuarios","descripcion":"Acceso","prioridad":"ALTA","criterios":["Credenciales validas"],"estado":"PENDIENTE","story_points":0,"horas_estimadas":0,"sprint_id":0}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18931/api/proyectos/1/sprints -d '{"nombre":"Sprint 1","fecha_inicio":"2026-10-26","fecha_fin":"2026-11-06","objetivo":"Calidad"}'
{"id":1,"proyecto_id":1,"nombre":"Sprint 1","fecha_inicio":"2026-10-26","fecha_fin":"2026-11-06","objetivo":"Calidad","estado":"ABIERTO"}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18931/api/items/1/defectos -d '{"descripcion":"El total no suma el IVA","severidad":"ALTA","sprint_deteccion_id":1}'
{"id":1,"descripcion":"El total no suma el IVA","severidad":"ALTA","estado":"ABIERTO","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18931/api/items/1/defectos -d '{"descripcion":"El boton de pago no responde","severidad":"MEDIA","sprint_deteccion_id":1}'
{"id":2,"descripcion":"El boton de pago no responde","severidad":"MEDIA","estado":"ABIERTO","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18931/api/defectos/1/estado -d '{"estado":"EN_CORRECCION"}'
{"id":1,"descripcion":"El total no suma el IVA","severidad":"ALTA","estado":"EN_CORRECCION","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' http://localhost:18931/api/proyectos/1/defectos
{"defectos":[{"id":1,"descripcion":"El total no suma el IVA","severidad":"ALTA","estado":"EN_CORRECCION","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0},{"id":2,"descripcion":"El boton de pago no responde","severidad":"MEDIA","estado":"ABIERTO","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}],"cantidad":2}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' "http://localhost:18931/api/proyectos/1/defectos?estado=en_correccion"
{"defectos":[{"id":1,"descripcion":"El total no suma el IVA","severidad":"ALTA","estado":"EN_CORRECCION","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}],"cantidad":1}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' "http://localhost:18931/api/proyectos/1/defectos?estado=%20%20"
{"defectos":[{"id":1,"descripcion":"El total no suma el IVA","severidad":"ALTA","estado":"EN_CORRECCION","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0},{"id":2,"descripcion":"El boton de pago no responde","severidad":"MEDIA","estado":"ABIERTO","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}],"cantidad":2}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' "http://localhost:18931/api/proyectos/1/defectos?estado=CERRADO"
{"error":"el estado debe ser ABIERTO, EN_CORRECCION o RESUELTO"}

HTTP 400

$ curl -s -w '\nHTTP %{http_code}\n' http://localhost:18931/api/sprints/1/defectos
{"defectos":[{"id":1,"descripcion":"El total no suma el IVA","severidad":"ALTA","estado":"EN_CORRECCION","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0},{"id":2,"descripcion":"El boton de pago no responde","severidad":"MEDIA","estado":"ABIERTO","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}],"cantidad":2}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18931/api/proyectos -d '{"nombre":"Sistema de stock","descripcion":""}'
{"id":2,"nombre":"Sistema de stock","descripcion":"","estado":"PLANIFICADO","fecha_inicio":""}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' http://localhost:18931/api/proyectos/2/defectos
{"defectos":[],"cantidad":0}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' http://localhost:18931/api/proyectos/999/defectos
{"error":"no existe un proyecto con ese id"}

HTTP 404

$ curl -s -w '\nHTTP %{http_code}\n' http://localhost:18931/api/sprints/999/defectos
{"error":"no existe un sprint con ese id"}

HTTP 404
```

Resultados del recorrido:

- El proyecto lista sus 2 defectos en orden por ID, con `cantidad` 2.
- `?estado=en_correccion` se normaliza y deja solo el defecto 1.
- `?estado=%20%20` (solo espacios) no filtra y devuelve los 2 defectos.
- `?estado=CERRADO` da 400 con el mensaje de `ErrInvalidStatus`.
- El sprint 1 lista los 2 defectos detectados en él.
- Un proyecto sin defectos devuelve `{"defectos":[],"cantidad":0}`.
- Un proyecto o un sprint inexistente da 404.

La línea vacía antes de `HTTP` es el salto de línea con el que la API termina cada respuesta JSON.

La interfaz web no se recorrió en un navegador. La cubren los tests de `internal/httpapi/web_test.go` (`TestInterfaz_FormularioConsultarDefectos`, `TestInterfaz_TablaDeDefectosTieneLasColumnas`, `TestInterfaz_ConsultarDefectosLlamaALaAPI` y `TestInterfaz_FiltroDeEstadoDeDefectosUsaLaListaDelPaquete`), que verifican el formulario, la tabla, las llamadas de `app.js`, los textos y que el filtro use la lista `estadosDefecto`, pero no ejecutan el JavaScript.

## Validación final

Medido sobre `c4f07e5` con go1.27.1 windows/amd64, sin `TEST_DATABASE_URL`.

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

Salida guardada en `hu31-green-go-test.txt`:

```text
ok  	seguimiento-medicion-gpc/cmd/server	4.290s
ok  	seguimiento-medicion-gpc/internal/backlog	5.681s
ok  	seguimiento-medicion-gpc/internal/defect	4.417s
ok  	seguimiento-medicion-gpc/internal/effort	2.874s
ok  	seguimiento-medicion-gpc/internal/httpapi	4.016s
ok  	seguimiento-medicion-gpc/internal/metrics	2.127s
ok  	seguimiento-medicion-gpc/internal/poker	2.705s
ok  	seguimiento-medicion-gpc/internal/postgres	1.583s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.295s
ok  	seguimiento-medicion-gpc/internal/project	2.745s
ok  	seguimiento-medicion-gpc/internal/sprint	3.326s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
(exit 0)
```

### `go test -count=1 -cover ./...`

Salida guardada en `hu31-cover-rama.txt`:

```text
ok  	seguimiento-medicion-gpc/cmd/server	4.009s	coverage: 31.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	7.015s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/defect	7.185s	coverage: 77.7% of statements
ok  	seguimiento-medicion-gpc/internal/effort	6.899s	coverage: 75.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	3.106s	coverage: 97.8% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	1.612s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	2.547s	coverage: 69.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.707s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	3.732s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	4.316s	coverage: 67.5% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	4.477s	coverage: 76.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

Con `-coverprofile` y `go tool cover -func`, sin base (`hu31-cover-rama-func.txt`):

- en 100.0%: `ListByProject`, `ListBySprint`, `selectDefects`, `ParseStatus`, `FilterByStatus`, `List` en memoria, `listDefectsByProject`, `listDefectsBySprint`, `listDefects` y `toDefectListResponse`;
- `List` de PostgreSQL en 0.0%, porque sus tests se omiten sin `TEST_DATABASE_URL`;
- total **81.2%**.

Para comparar con `main` se usó `f6e8101`, el `git merge-base` con `origin/main` después de `git fetch origin`. Se midió con el mismo Go en una copia temporal (`git worktree` en `C:/wt-main`, borrada al terminar; salidas en `hu31-cover-main.txt` y `hu31-cover-main-total.txt`):

- `internal/defect` baja de 79.6% a 77.7%, porque `List` de PostgreSQL suma sentencias que sin base no se ejecutan;
- `internal/httpapi` sube de 97.7% a 97.8%;
- los demás paquetes se mantienen;
- el total del perfil sube de **81.1%** a **81.2%**.

Los perfiles se borraron al terminar.

### Con PostgreSQL

Medido sobre el GREEN con PostgreSQL 17.11 (`SHOW server_version`; imagen `postgres:17-alpine` del `docker-compose.yml` del repo). HU-31 no cambia el esquema, así que la base no se recreó: no se hizo `docker compose down -v`. **La base corrió en el puerto 5434 y no en el 5433**, porque el 5433 lo ocupa otro contenedor ajeno al proyecto (`postgresql-servidor`, según `docker ps`).

Con `TEST_DATABASE_URL=postgres://gpc:gpc_dev@localhost:5434/gpc_test?sslmode=disable`:

- `go test -count=1 ./...`: todos los paquetes `ok` (`hu31-green-pg-go-test.txt`).
- `go test -count=1 -v ./...`: **0 SKIP, 1721 PASS y 0 FAIL**, contados con `grep -c` de `--- SKIP`, `--- PASS` y `--- FAIL`, incluidas las subtests (`hu31-green-pg-verbose.txt`).
- `TestPostgresRepository_ContratoList` corrió sus 3 subtests contra PostgreSQL sin SKIP, y `TestPostgresRepository_ListLeeLaTablaOrdenadaPorID` pasó.
- `go test -count=1 -cover ./...` (`hu31-green-pg-cover-paquetes.txt`): `internal/defect` 95.6%, `internal/httpapi` 97.8% y `cmd/server` 33.3%.
- `go test -count=1 -coverpkg=./... -coverprofile=… ./...` (`hu31-green-pg-cover.txt` y `hu31-green-pg-cover-func.txt`): total **95.0%**. `List` del repositorio PostgreSQL de defectos queda en **81.2%**. Faltan las tres ramas de error de la base: la falla de `Query`, la de `scanDefect` y la de `rows.Err()`. Solo ocurren con una falla de conexión o de la consulta, y los tests no las simulan.

El perfil se borró al terminar.

### `-race`

`-race` no se corrió localmente: `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1` (`go env CGO_ENABLED` da `0`). Los tests de concurrencia (`TestList_ConsultasSimultaneasConRegistrosYCambiosDeEstado` y el escenario de CA-31.7) corrieron sin `-race` y pasaron. Con `-race` los ejecuta el CI del Pull Request (Linux, `go test -race -coverprofile=coverage.out ./...`). La mutación (h) tampoco la detectaría `-race`, porque no deja accesos sin protección.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `2bc0cef` | `docs(HU-31): agregar especificación SDD` |
| Aprobación de la SDD | `6f65298` | `docs(HU-31): registrar la aprobación de la SDD` |
| RED | `b44d97a` | `test(HU-31): agregar tests (RED)` |
| GREEN | `c4f07e5` | `feat(HU-31): implementar la consulta de defectos (GREEN)` |
| Refactor | — | No aplica |
| Documentación | (propuesto) | `docs(HU-31): agregar evidencia TDD y cobertura` |

Todos los commits de la historia referencian el Issue #49 con `Refs #49` y están en la rama `feature/HU-31-consultar-defectos`.

HU-31: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
