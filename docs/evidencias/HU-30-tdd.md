# Evidencia TDD · HU-30

> Issue del repositorio: #48

## Historia

Como Product Builder, quiero cambiar el estado de un defecto y registrar en qué sprint se resolvió, para reflejar si está abierto, en corrección o resuelto.

## Resultado

HU-30 quedó implementada en la rama `feature/HU-30-cambiar-estado-defecto`. `Manager.ChangeStatus` (`internal/defect/status.go`) cambia el estado de un defecto existente entre `ABIERTO`, `EN_CORRECCION` y `RESUELTO` con transiciones de un paso, como HU-08. Valida en el orden de la SDD:

1. Sin tocar repositorios: el estado y, si el destino es `RESUELTO`, el sprint de resolución ≤ 0.
2. Con el mutex del `Manager`: la existencia del defecto, el mismo estado (no-op, sin `Update`) y la transición.
3. Al resolver, dentro de `WithOpenSprint`: que el sprint exista, esté abierto, sea del proyecto de la historia y no sea anterior al de detección.

Al dejar `RESUELTO`, el sprint de resolución vuelve a 0. `Repository` suma `GetByID` y `Update`, implementados en memoria y en PostgreSQL, sin cambios de esquema. `StatusValues()` queda en `defect.go` y el comentario de `Status` remite a `status.go` para las transiciones.

La API lo expone en `PATCH /api/defectos/{id}/estado` con `{"estado": "...", "sprint_resolucion_id": N}`. En `statusFor`, `ErrInvalidStatus` y `ErrMissingResolutionSprint` dan 400, `ErrDefectNotFound` da 404, y `ErrInvalidTransition`, `ErrResolutionSprintFromAnotherProject` y `ErrResolutionBeforeDetection` dan 409. La interfaz web suma, en la sección Defectos y debajo de "Registrar defecto", el formulario "Cambiar estado de defecto".

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test |
|---|---|
| CA-30.1 — transiciones de un paso; ABIERTO → RESUELTO y RESUELTO → ABIERTO se rechazan; pedir el estado actual es un no-op | “Pasar un defecto abierto a en corrección”; “Resolver un defecto en corrección en el sprint abierto”; “Volver un defecto en corrección a abierto”; “Reabrir un defecto resuelto limpia el sprint de resolución”; “Rechazar abierto a resuelto y resuelto a abierto” (esquema, 2 ejemplos); “Pedir el estado actual no cambia nada” (esquema, 4 ejemplos); `TestChangeStatus_AbiertoAEnCorreccion`; `TestChangeStatus_EnCorreccionAResuelto`; `TestChangeStatus_EnCorreccionAAbierto`; `TestChangeStatus_ReabrirLimpiaElSprintDeResolucion`; `TestChangeStatus_TransicionesNoPermitidas`; `TestChangeStatus_MismoEstadoEsNoOp`; `TestChangeStatus_TransicionAntesQueSprint` |
| CA-30.2 — sprint de resolución obligatorio, existente, abierto, del mismo proyecto y no anterior al de detección (puede ser el mismo); vuelve a 0 al reabrir | “Resolver un defecto en corrección en el sprint abierto”; “Resolver en un sprint posterior al de detección”; “Reabrir un defecto resuelto limpia el sprint de resolución”; “Rechazar resolver sin sprint de resolución” (esquema, 4 ejemplos, incluido RESUELTO → RESUELTO); “Rechazar resolver con el sprint de resolución ausente”; “Rechazar un sprint de resolución inexistente, cerrado o de otro proyecto” (esquema, 3 ejemplos); “Rechazar un sprint de resolución anterior al de detección”; `TestChangeStatus_ResolverEnSprintPosteriorAlDeDeteccion`; `TestChangeStatus_ResolverEnElMismoSprintDeDeteccion`; `TestChangeStatus_ResueltoAResueltoSinSprintDeResolucion`; `TestChangeStatus_SprintDeResolucionObligatorio`; `TestChangeStatus_SprintAusenteSobreDefectoInexistente`; `TestChangeStatus_IgnoraElSprintDeResolucionConOtroDestino`; `TestChangeStatus_SprintDeResolucionInexistente`; `TestChangeStatus_SprintDeResolucionCerrado`; `TestChangeStatus_SprintDeResolucionDeOtroProyecto`; `TestChangeStatus_SprintDeResolucionAnteriorAlDeDeteccion` |
| CA-30.3 — estado vacío o inválido; normalización de mayúsculas y espacios | “Rechazar un estado inválido” (esquema, 3 ejemplos); “Normalizar el estado” (esquema, 2 ejemplos); `TestChangeStatus_NormalizaElEstado`; `TestChangeStatus_EstadoInvalido`; `TestChangeStatus_EstadoInvalidoSobreDefectoInexistente` |
| CA-30.4 — defecto inexistente (incluidos 0 y negativo) | “Rechazar un defecto inexistente” (esquema con 999, 0 y -1); `TestChangeStatus_DefectoInexistente` |
| CA-30.5 — un cambio rechazado no modifica nada; uno válido solo cambia el estado y el sprint de resolución | “No modificar la historia, los sprints ni los demás campos”; `TestChangeStatus_NoModificaHistoriaSprintsNiOtrosCampos`; además, cada test y escenario de error verifica que el defecto guardado no cambió |
| CA-30.6 — persistencia (memoria y PostgreSQL) y copias por valor | “Modificar el defecto devuelto no altera el guardado”; `TestChangeStatus_CopiaIndependiente`; `TestMemoryRepository_ContratoGetByIDYUpdate` y `TestPostgresRepository_ContratoGetByIDYUpdate` (6 subtests de contrato cada uno); `TestPostgresRepository_UpdateGuardaElCambioEnLaTabla`; `TestPostgresRepository_UpdateInexistenteNoCreaFilas` |
| CA-30.7 — concurrencia y cierre simultáneo del sprint | “Cambios simultáneos sobre el mismo defecto”; “Resolver y cerrar el sprint en simultáneo”; `TestChangeStatus_Concurrente`; `TestChangeStatus_ConcurrenteConCierre` |
| CA-30.8 — falla del repositorio al buscar o al actualizar | “Error del repositorio” (esquema, 2 ejemplos); `TestChangeStatus_ErrorDelRepositorio` |
| CA-30.9 — API e interfaz web | Los 10 `TestCambiarEstadoDefecto_*` de `internal/httpapi/defects_test.go` (200 y un caso por 400, 404, 409, 413 y 500); caso “cambiar estado de defecto” de `TestRutasConID_IDNoNumericoDevuelve400`; endpoint “cambiar estado de defecto” de `bodyEndpoints` (413 en `TestHandler_CuerpoMayorA1MiBDa413` y texto después del JSON en `TestHandler_TextoDespuesDelJSONDa400`); `TestInterfaz_FormularioCambiarEstadoDeDefecto`; `TestInterfaz_CambiarEstadoDeDefectoMuestraElMensajeDeExito`; `TestInterfaz_CambiarEstadoDeDefectoMuestraElMensajeDeError`; `TestInterfaz_ListaDeEstadosDeDefectosCoincideConElPaquete`; `TestStatusValues_DevuelveLosEstadosEnOrden` y `TestStatusValues_DevuelveUnaListaNueva` (la lista que usa la interfaz) |

El archivo BDD tiene **19 bloques de escenario**: 11 escenarios y 8 esquemas con 23 ejemplos en total, que godog ejecuta como **34 escenarios**. Las pruebas son **47 funciones de test nuevas**:

- 28 unitarias de `internal/defect/status_test.go`;
- 1 runner de godog;
- 1 de contrato en memoria y 3 de PostgreSQL;
- 10 HTTP;
- 4 de interfaz web.

Además, cada una de las dos funciones de contrato corre 6 subtests de `runRepositoryStatusContract`. También hay 2 casos de tabla en tests HTTP existentes: el de `TestRutasConID_IDNoNumericoDevuelve400` y el endpoint de `bodyEndpoints`, que corre en dos tests, así que suman 3 subtests.

La correspondencia no es 1 a 1. CA-30.9 no tiene escenario BDD porque la API y la interfaz se prueban con los tests de `internal/httpapi`. Varios escenarios cubren dos criterios a la vez (CA-30.1 y CA-30.2).

Conteos con `grep` sobre los archivos de la rama (contra `main`, `caf1d02`, para los archivos que ya existían):

```text
$ grep -c '^\s*Escenario:' features/HU-30-cambiar-estado-defecto.feature
11
$ grep -c '^\s*Esquema del escenario:' features/HU-30-cambiar-estado-defecto.feature
8
$ grep -c '^func Test' internal/defect/status_test.go
28
$ grep -c '^func Test' internal/defect/status_features_test.go
1
$ git diff caf1d02 HEAD -- internal/defect/repository_contract_test.go | grep -c '^+func Test'
1
$ git diff caf1d02 HEAD -- internal/defect/postgres_repository_test.go | grep -c '^+func Test'
3
$ git diff caf1d02 HEAD -- internal/httpapi/defects_test.go | grep -c '^+func Test'
10
$ git diff caf1d02 HEAD -- internal/httpapi/web_test.go | grep -c '^+func Test'
4
$ git diff caf1d02 HEAD -- internal/httpapi/server_test.go | grep -c '^+func Test'
0
$ git diff caf1d02 HEAD -- internal/defect/repository_contract_test.go | grep -cE '^\+\s+t\.Run\("'
6
$ git diff caf1d02 HEAD -- internal/httpapi/server_test.go | grep -cE '^\+.*"cambiar estado de defecto"'
2
```

Bloques de escenario por criterio (`grep` del CA en el nombre de cada escenario o esquema): CA-30.1: 6; CA-30.2: 7; CA-30.3: 2; CA-30.4: 1; CA-30.5: 1; CA-30.6: 1; CA-30.7: 2; CA-30.8: 1; CA-30.9: 0.

## BDD

Archivo: `features/HU-30-cambiar-estado-defecto.feature`, con el runner `TestFeatures_HU30CambiarEstadoDefecto` en `internal/defect/status_features_test.go`. Todos los pasos llaman al `Manager` de `internal/defect`, no al handler HTTP.

Los defectos de partida se guardan con `Save` del repositorio en memoria, porque registrarlos es HU-29 y el caso del sprint de resolución anterior al de detección no se puede armar con `Register`. Los sprints se guardan con el repositorio, así un escenario puede tener más de un sprint abierto por proyecto. El repositorio en memoria va envuelto en un doble que cuenta los `Update` (para el no-op) y puede fallar al buscar o al actualizar. Gherkin recorta los espacios de las celdas, así que la normalización con espacios se prueba en `status_test.go`.

Resultado de `go test -count=1 -v -run TestFeatures_HU30 ./internal/defect/` sobre el GREEN (salida guardada en `HU-30-green-godog.txt`):

```text
34 scenarios (34 passed)
276 steps (276 passed)
--- PASS: TestFeatures_HU30CambiarEstadoDefecto
```

## Fase RED

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `abba532` | `docs(HU-30): agregar especificación SDD` |
| Ajuste de la SDD | `36a321b` | `docs(HU-30): ajustar la especificación SDD` |
| Aprobación de la SDD | `369ff72` | `docs(HU-30): registrar la aprobación de la SDD` |
| RED | `cf5eb81` | `test(HU-30): agregar tests (RED)` |

Antes de los tests se agregaron solo los stubs que permite la SDD:

- los 6 errores nuevos con sus mensajes;
- `statusValues` y `StatusValues()` completos;
- las firmas vacías de `ChangeStatus` y de `GetByID` y `Update` en memoria y PostgreSQL, que devolvían `Defect{}, nil`.

Siguiendo el ajuste que pidió Peñalbé Hernán en HU-18, **la ruta `PATCH /api/defectos/{id}/estado` y el handler `changeDefectStatus` no estaban en el RED**, como tampoco el formulario web. Sin la ruta, los tests HTTP fallan por aserción con 405. `gofmt -l .` y `go vet ./...` no informaron nada antes del commit RED.

Para que los dobles de HU-29 (`testDefectRepository`, `failingDefectRepository` y `brokenDefectRepository`) siguieran cumpliendo `Repository`, se les agregaron `GetByID` y `Update` en `status_test.go` y `defects_test.go`. No se tocó `register_test.go`. En `repository_contract_test.go` solo cambió el import, para sumar `errors`.

El RED falló por las aserciones y no por compilación, solo en `internal/defect` e `internal/httpapi`. Se contó con `grep` sobre la salida guardada de `go test -count=1 ./...` (`HU-30-red-go-test.txt`):

- fallaron **45 tests de primer nivel** (`^--- FAIL:`) y **99 subtests**, incluidos los 34 escenarios godog;
- no hubo ninguna línea con `build failed`, `undefined:`, `cannot use` ni `[setup failed]`.

De los 45, 42 son tests nuevos de HU-30 y 3 son tests existentes que fallaron solo por su subtest nueva.

- Los unitarios y el contrato en memoria fallaron con el resultado del stub, por ejemplo `err = <nil>, se esperaba la transición de estado del defecto no está permitida` o `GetByID = {ID:0 Description: …}, se esperaba {ID:2 Description:Descripción alta …}`.
- BDD: `34 scenarios (34 failed)`, `276 steps (205 passed, 34 failed, 37 skipped)`. HU-29 siguió con `16 scenarios (16 passed)`.
- Los 10 `TestCambiarEstadoDefecto_*` recibieron `405` (`Method Not Allowed`, `text/plain`) porque la ruta no existía.
- **Los 3 tests que ya existían y fallaron** son `TestRutasConID_IDNoNumericoDevuelve400`, `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400`. Fallaron solo por la subtest nueva “cambiar_estado_de_defecto”, que recibió 405. Sus otras 41 subtests pasaron.
- Los 4 tests de interfaz fallaron porque la página y `app.js` no tenían el formulario, los textos ni la lista `estadosDefecto`.

**Ya pasaban en el RED (5)**, sin aportar un fallo:

- `TestStatusValues_DevuelveLosEstadosEnOrden` y `TestStatusValues_DevuelveUnaListaNueva`, porque la SDD permite `StatusValues()` completo en el RED.
- `TestPostgresRepository_ContratoGetByIDYUpdate`, `TestPostgresRepository_UpdateGuardaElCambioEnLaTabla` y `TestPostgresRepository_UpdateInexistenteNoCreaFilas` hicieron SKIP porque no estaba definida `TEST_DATABASE_URL`.

**Dos tests pasaban en falso y se corrigieron antes del commit RED:** el escenario “No modificar la historia, los sprints ni los demás campos (CA-30.5)” y `TestChangeStatus_NoModificaHistoriaSprintsNiOtrosCampos`. Solo comprobaban que no cambiaran los demás campos, y con el stub (que no cambia nada) pasaban. Se les agregó la verificación de que el defecto guardado quedó `RESUELTO` en el sprint 1, y desde entonces fallan en el RED.

## Fase GREEN

| Fase | Commit | Mensaje |
|---|---|---|
| GREEN | `8b87c09` | `feat(HU-30): implementar el cambio de estado de un defecto (GREEN)` |
| Refactor | — | No aplica |
| Documentación | `3c16a48` | `docs(HU-30): agregar evidencia TDD y cobertura` |

El GREEN tocó solo 8 archivos de producción:

- `internal/defect`: `status.go` (`ChangeStatus`, `allowedTransitions`, `normalizeStatus` e `isTransitionAllowed`), `defect.go` (comentario de `Status`), `memory_repository.go` y `postgres_repository.go` (`GetByID` y `Update`);
- `internal/httpapi`: `defects.go` (handler `changeDefectStatus`) y `server.go` (ruta y errores en `statusFor`);
- `internal/web`: `index.html` y `app.js`.

**Ningún archivo `*_test.go` ni el `.feature` cambió en el GREEN**: `git diff --stat HEAD -- '*_test.go' features/` salió vacío antes del commit. Sin base, todos los paquetes dieron `ok` (`HU-30-green-sin-base.txt`). Los 34 escenarios de HU-30 pasaron y HU-29 siguió con `16 scenarios (16 passed)`, `109 steps (109 passed)` (`HU-30-green-hu29.txt`).

### Ramas sin cubrir y una decisión de la interfaz

- **Rama defensiva de `GetItem` en `ChangeStatus`** (`status.go:59-60`, `ChangeStatus` en 97.4% sin base): es el `return err` cuando `m.items.GetItem(current.ItemID)` falla al resolver. No se puede provocar con la pila real, porque la historia existía al registrar el defecto (HU-29) y el sistema no elimina ítems. La SDD ("Decisión sobre entidades referenciadas") pide propagar ese error. Por decisión de Sosa Ricardo no se agregó un test para esta rama: queda documentada como defensiva, como en HU-34.
- **`GetByID` y `Update` de PostgreSQL en 85.7%** (con base y `-coverpkg=./...`): falta 1 de sus 7 sentencias, el `return` del error genérico de la base (distinto de `pgx.ErrNoRows`), envuelto con `fmt.Errorf`. Solo ocurre con una falla de conexión o de la consulta. Es el mismo patrón que `GetByID` de sprints (85.7%).
- **El mensaje de éxito de `app.js`** se arma así:

  ```js
  let texto = "Defecto " + cambiado.id + " ahora está en " + cambiado.estado;
  if (cambiado.estado === "RESUELTO") {
    texto += "," + " resuelto en el sprint " + cambiado.sprint_resolucion_id;
  }
  ```

  `TestInterfaz_CambiarEstadoDeDefectoMuestraElMensajeDeExito`, escrito en el RED, busca en `app.js` el literal `" resuelto en el sprint "`, con la comilla inmediatamente antes del espacio. Con `", resuelto en el sprint "` el texto mostrado sería el mismo, pero ese literal no aparecería en el archivo y el test fallaría. Como en el GREEN no se tocan los tests, se separó la coma en `","`. El usuario ve "Defecto 1 ahora está en RESUELTO, resuelto en el sprint 1.".

## Pruebas de mutación

Cada mutación se aplicó sobre el GREEN (`8b87c09`) y se corrió `go test -count=1 ./internal/defect/ ./internal/httpapi/`. Después se restauró con `git checkout -- <archivo>`, y en todos los casos `git diff` salió vacío. De la (a) a la (f) corrieron sin base y la (g) con `TEST_DATABASE_URL` en el puerto 5434. Salidas: `HU-30-mutacion-<letra>.txt`.

| Mutación | Resultado | Tests que fallan (mensaje) |
|---|---|---|
| (a) `status.go`: quitar el chequeo del sprint de resolución ≤ 0 | Detectada | `TestChangeStatus_ResueltoAResueltoSinSprintDeResolucion` (`err = <nil>, se esperaba el sprint de resolución es obligatorio para resolver un defecto`); `TestChangeStatus_SprintDeResolucionObligatorio` (`err = no existe un sprint con ese id, se esperaba …`); `TestChangeStatus_SprintAusenteSobreDefectoInexistente` (`err = no existe un defecto con ese id, se esperaba …`); 4 casos de `TestCambiarEstadoDefecto_DatosInvalidosDevuelven400` (`código = 409, se esperaba 400`); 5 escenarios godog de “Rechazar resolver sin sprint de resolución” y “… ausente” |
| (b) `status.go`: permitir ABIERTO → RESUELTO | Detectada | `TestChangeStatus_TransicionesNoPermitidas/ABIERTO_a_RESUELTO` (`err = <nil>, se esperaba la transición de estado del defecto no está permitida`); `TestChangeStatus_TransicionAntesQueSprint` (`err = el sprint está cerrado, se esperaba la transición …`); `TestCambiarEstadoDefecto_TransicionNoPermitidaDevuelve409` (`código = 200, se esperaba 409`); escenario “Rechazar abierto a resuelto y resuelto a abierto (CA-30.1)” |
| (c) `status.go`: no volver el sprint de resolución a 0 al reabrir | Detectada | `TestChangeStatus_ReabrirLimpiaElSprintDeResolucion` (`… ResolutionSprintID:1}, se esperaba {… ResolutionSprintID:0}`); `TestCambiarEstadoDefecto_ReabrirDevuelveElSprintDeResolucionEnCero` (`la respuesta no tiene el sprint de resolución en 0`); escenario “Reabrir un defecto resuelto limpia el sprint de resolución” |
| (d) `status.go`: guardar fuera de `WithOpenSprint` al resolver (las validaciones siguen dentro y el `Update` va después) | **Sobrevive** | `go test -count=1` de los dos paquetes terminó en `ok`. Con `-count=50 -v -run 'TestChangeStatus_ConcurrenteConCierre$'`, 50 lotes PASS y 0 FAIL, con 20 intentos por lote (`HU-30-mutacion-d-count50.txt`) |
| (e) `status.go`: rechazar también el mismo sprint de detección (`<=` en lugar de `<`) | Detectada | `TestChangeStatus_EnCorreccionAResuelto`, `TestChangeStatus_ResolverEnElMismoSprintDeDeteccion`, `TestChangeStatus_NormalizaElEstado`, `TestChangeStatus_NoModificaHistoriaSprintsNiOtrosCampos` y `TestChangeStatus_ConcurrenteConCierre` (`ChangeStatus devolvió un error inesperado: el sprint de resolución no puede ser anterior al de detección`); `TestCambiarEstadoDefecto_Devuelve200ConElDefecto` y `…_ReabrirDevuelveElSprintDeResolucionEnCero` (`código = 409, se esperaba 200`); 3 escenarios godog |
| (f) `status.go`: quitar el chequeo del mismo proyecto | Detectada | `TestChangeStatus_SprintDeResolucionDeOtroProyecto` (`err = <nil>, se esperaba el sprint de resolución pertenece a otro proyecto`); `TestCambiarEstadoDefecto_SprintDeOtroProyectoDevuelve409` (`código = 200, se esperaba 409`); escenario “Rechazar un sprint de resolución inexistente, cerrado o de otro proyecto (CA-30.2)” (ejemplo 3) |
| (g) `postgres_repository.go`: `Update` no escribe el estado | Detectada (con base) | `TestPostgresRepository_ContratoGetByIDYUpdate`, en 3 subtests (por ejemplo, `Update = {… Status:ABIERTO …}, se esperaba {… Status:RESUELTO …}`), y `TestPostgresRepository_UpdateGuardaElCambioEnLaTabla` (`defecto actualizado = {… Status:ABIERTO …}, se esperaba {… Status:RESUELTO …}`) |

En la (g), el `SET` del `UPDATE` quedó con `status = CASE WHEN $4::text IS NULL THEN status ELSE status END`. Así `$4` sigue referenciado, PostgreSQL no rechaza el parámetro sin usar y el estado guardado no cambia.

**La mutación (d) sobrevive.** Las validaciones siguen dentro de `WithOpenSprint` y el mutex del `Manager` de defectos sigue tomado durante toda la operación. La ventana entre que `WithOpenSprint` suelta el mutex de sprints y el `Update` en memoria es muy corta, y `TestChangeStatus_ConcurrenteConCierre` no la alcanzó en 1000 intentos (50 lotes de 20). No se cambiaron los tests. `-race` tampoco la detectaría por sí solo: no es un acceso sin protección, sino un orden de operaciones (ver "-race").

## Prueba de punta a punta de la API

El servidor real se compiló desde `8b87c09` con `go build -o hu30-server.exe ./cmd/server`, fuera del repositorio. Se arrancó en memoria, sin `STORAGE` ni `TEST_DATABASE_URL`, con `PORT=18630 ./hu30-server.exe` (log: `servidor escuchando en :18630 (almacenamiento: memory)`). Los pedidos se hicieron con `curl` desde Git Bash, con textos de entrada sin tildes porque Git Bash no los envía en UTF-8. Al terminar se detuvo el servidor y se borró el binario. Salida: `HU-30-e2e.txt`.

```
$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18630/api/proyectos -d '{"nombre":"Sistema de ventas"}'
{"id":1,"nombre":"Sistema de ventas","descripcion":"","estado":"PLANIFICADO","fecha_inicio":""}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18630/api/proyectos/1/backlog -d '{"titulo":"Login de usuarios","descripcion":"Acceso","prioridad":"alta","criterios":["Credenciales validas"]}'
{"id":1,"proyecto_id":1,"titulo":"Login de usuarios","descripcion":"Acceso","prioridad":"ALTA","criterios":["Credenciales validas"],"estado":"PENDIENTE","story_points":0,"horas_estimadas":0,"sprint_id":0}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18630/api/proyectos/1/sprints -d '{"nombre":"Sprint 1","fecha_inicio":"2026-10-26","fecha_fin":"2026-11-08","objetivo":"Entregar el MVP"}'
{"id":1,"proyecto_id":1,"nombre":"Sprint 1","fecha_inicio":"2026-10-26","fecha_fin":"2026-11-08","objetivo":"Entregar el MVP","estado":"ABIERTO"}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18630/api/items/1/defectos -d '{"descripcion":"El total no descuenta el descuento","severidad":"ALTA","sprint_deteccion_id":1}'
{"id":1,"descripcion":"El total no descuenta el descuento","severidad":"ALTA","estado":"ABIERTO","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}

HTTP 201

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18630/api/defectos/1/estado -d '{"estado":"RESUELTO","sprint_resolucion_id":1}'
{"error":"la transición de estado del defecto no está permitida"}

HTTP 409

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18630/api/defectos/1/estado -d '{"estado":"en_correccion"}'
{"id":1,"descripcion":"El total no descuenta el descuento","severidad":"ALTA","estado":"EN_CORRECCION","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18630/api/defectos/1/estado -d '{"estado":"RESUELTO"}'
{"error":"el sprint de resolución es obligatorio para resolver un defecto"}

HTTP 400

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18630/api/defectos/1/estado -d '{"estado":"RESUELTO","sprint_resolucion_id":1}'
{"id":1,"descripcion":"El total no descuenta el descuento","severidad":"ALTA","estado":"RESUELTO","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":1}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18630/api/defectos/1/estado -d '{"estado":"ABIERTO"}'
{"error":"la transición de estado del defecto no está permitida"}

HTTP 409

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18630/api/defectos/1/estado -d '{"estado":"EN_CORRECCION"}'
{"id":1,"descripcion":"El total no descuenta el descuento","severidad":"ALTA","estado":"EN_CORRECCION","item_id":1,"sprint_deteccion_id":1,"sprint_resolucion_id":0}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18630/api/defectos/1/estado -d '{"estado":"CERRADO"}'
{"error":"el estado debe ser ABIERTO, EN_CORRECCION o RESUELTO"}

HTTP 400

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18630/api/defectos/999/estado -d '{"estado":"EN_CORRECCION"}'
{"error":"no existe un defecto con ese id"}

HTTP 404

$ curl -s -w '\nHTTP %{http_code}\n' -X POST http://localhost:18630/api/sprints/1/cerrar
{"id":1,"proyecto_id":1,"nombre":"Sprint 1","fecha_inicio":"2026-10-26","fecha_fin":"2026-11-08","objetivo":"Entregar el MVP","estado":"CERRADO"}

HTTP 200

$ curl -s -w '\nHTTP %{http_code}\n' -X PATCH http://localhost:18630/api/defectos/1/estado -d '{"estado":"RESUELTO","sprint_resolucion_id":1}'
{"error":"el sprint está cerrado"}

HTTP 409
```

Resultados del recorrido:

- El defecto se registra `ABIERTO` con `sprint_resolucion_id` 0.
- Saltar de ABIERTO a RESUELTO da 409, y `en_correccion` se normaliza y da 200.
- RESUELTO sin sprint da 400. Con el sprint 1 da 200 y registra `sprint_resolucion_id` 1.
- RESUELTO → ABIERTO da 409. Reabrir da 200 y vuelve el sprint de resolución a 0.
- Un estado inválido da 400 y un defecto inexistente 404.
- Con el sprint cerrado, resolver da 409.

La línea vacía antes de `HTTP` es el salto de línea con el que la API termina cada respuesta JSON.

La interfaz web no se recorrió en un navegador. La cubren los tests de `internal/httpapi/web_test.go` (`TestInterfaz_FormularioCambiarEstadoDeDefecto`, `TestInterfaz_CambiarEstadoDeDefectoMuestraElMensajeDeExito`, `TestInterfaz_CambiarEstadoDeDefectoMuestraElMensajeDeError` y `TestInterfaz_ListaDeEstadosDeDefectosCoincideConElPaquete`), que verifican el formulario, la llamada de `app.js`, los textos y la lista de estados, pero no ejecutan el JavaScript.

## Validación final

Medido sobre `8b87c09` con go1.27.1 windows/amd64, sin `TEST_DATABASE_URL`.

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

Salida guardada en `HU-30-green-sin-base.txt`:

```text
ok  	seguimiento-medicion-gpc/cmd/server	2.485s
ok  	seguimiento-medicion-gpc/internal/backlog	4.227s
ok  	seguimiento-medicion-gpc/internal/defect	3.527s
ok  	seguimiento-medicion-gpc/internal/effort	1.923s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.525s
ok  	seguimiento-medicion-gpc/internal/metrics	0.946s
ok  	seguimiento-medicion-gpc/internal/poker	1.683s
ok  	seguimiento-medicion-gpc/internal/postgres	1.047s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.714s
ok  	seguimiento-medicion-gpc/internal/project	2.748s
ok  	seguimiento-medicion-gpc/internal/sprint	2.204s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
(exit 0)
```

### `go test -count=1 -cover ./...`

Salida guardada en `HU-30-cobertura-rama.txt`:

```text
ok  	seguimiento-medicion-gpc/cmd/server	1.470s	coverage: 31.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	4.094s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/defect	3.459s	coverage: 79.6% of statements
ok  	seguimiento-medicion-gpc/internal/effort	3.541s	coverage: 75.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	2.255s	coverage: 97.7% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	1.950s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	2.943s	coverage: 69.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	0.875s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	3.026s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	4.319s	coverage: 67.5% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	3.118s	coverage: 76.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

Con `-coverprofile` y `go tool cover -func`, sin base:

- en 100.0%: `StatusValues`, `GetByID` y `Update` en memoria, `normalizeStatus`, `isTransitionAllowed`, `changeDefectStatus`, `statusFor` y `NewHandler`;
- `ChangeStatus` en 97.4%, por la rama defensiva de `GetItem`;
- `GetByID` y `Update` de PostgreSQL de defectos en 0.0%, porque sus tests se omiten sin `TEST_DATABASE_URL`;
- total **81.1%**.

Para comparar con `main` se usó `caf1d02`, el `git merge-base` con `origin/main` después de `git fetch origin`. Se midió con el mismo Go en una copia temporal (`git worktree` en `C:/wt-main`, borrada al terminar; salida en `HU-30-cobertura-main.txt`). `internal/defect` sube de 76.6% a 79.6%, `internal/httpapi` de 97.6% a 97.7%, los demás paquetes se mantienen y el total del perfil sube de **80.9%** a **81.1%**.

### Con PostgreSQL

Medido sobre el GREEN con PostgreSQL 17.11 (imagen `postgres:17-alpine` del `docker-compose.yml` del repo). HU-30 no cambia el esquema, así que la base no se recreó: no se hizo `docker compose down -v`. **La base corrió en el puerto 5434 y no en el 5433**, porque el 5433 lo ocupa otro contenedor ajeno al proyecto (`postgresql-servidor`). Se usó `POSTGRES_PORT=5434` solo como variable del comando, sin `.env`.

Con `TEST_DATABASE_URL=postgres://gpc:gpc_dev@localhost:5434/gpc_test?sslmode=disable`:

- `go test -count=1 ./...`: todos los paquetes `ok` (`HU-30-green-con-base.txt`).
- `go test -count=1 -v ./...`: **0 SKIP, 1597 PASS y 0 FAIL**, contados con `grep -c` de `--- SKIP`, `--- PASS` y `--- FAIL` (`HU-30-green-con-base-conteo.txt`).
- `TestPostgresRepository_ContratoGetByIDYUpdate` corrió sus 6 subtests contra PostgreSQL sin SKIP, y `TestPostgresRepository_UpdateGuardaElCambioEnLaTabla` y `TestPostgresRepository_UpdateInexistenteNoCreaFilas` pasaron.
- `go test -count=1 -cover ./...` (`HU-30-green-con-base-cover.txt`): `internal/defect` 95.9%, `internal/httpapi` 97.7% y `cmd/server` 33.3%.
- `go test -count=1 -coverpkg=./... ./...` (`HU-30-green-coverpkg.txt`): total **94.9%**. En el repositorio PostgreSQL de defectos, `NewPostgresRepository`, `Save` y `scanDefect` quedan en 100.0%, y `GetByID` y `Update` en 85.7% por la rama de error genérico de la base.

### `-race`

`-race` no se corrió localmente: `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`. Los tests de concurrencia (`TestChangeStatus_Concurrente`, `TestChangeStatus_ConcurrenteConCierre` y los dos escenarios de CA-30.7) corrieron sin `-race` y pasaron. Con `-race` los ejecuta el CI del Pull Request (Linux, `go test -race -coverprofile=coverage.out ./...`). La mutación (d) de las pruebas de mutación tampoco la detectaría -race: no es un acceso sin protección, sino el orden entre el chequeo del sprint abierto y el Update.

## Commits

Los commits de la SDD, su ajuste y su aprobación, el RED, el GREEN y la documentación referencian el Issue #48 con `Refs #48`. Todos los commits de la historia están hechos y subidos a la rama `feature/HU-30-cambiar-estado-defecto`, y el Pull Request es el [#140](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/140).

HU-30: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
