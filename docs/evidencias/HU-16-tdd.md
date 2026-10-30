# Evidencia TDD · HU-16

> Issue del repositorio: #30

## Historia

Como Agile Enabler, quiero modificar el nombre o el rol de un integrante ya registrado, para mantener la información del equipo al día.

## Resultado

Se agregó la operación `Update(id, name, role)` en `internal/member/update.go`, con el error `ErrMemberNotFound` y la función `nameTaken(members, name, exceptID)`, que `Register` usa con 0 y `Update` con el ID del integrante, así que cambiar solo las mayúsculas del propio nombre es válido. La búsqueda, la comparación y el guardado van bajo el mismo mutex de `Register`. Los repositorios en memoria y PostgreSQL suman `GetByID` y `Update`, sin cambios de esquema; `Update` reemplaza el nombre y el rol y conserva el proyecto. Se agregaron la ruta `PUT /api/integrantes/{id}` con el handler `updateMember` (reutiliza `memberRequest` y `memberResponse`), el `405` de `GET` con `memberMethodNotAllowed` (`Allow: PUT`), `ErrMemberNotFound` en el caso 404 de `statusFor` y el formulario "Modificar integrante" en la sección **Integrantes**. No cambian el formulario "Registrar integrante", `members.go`, `manager.go`, `NewHandler` ni `cmd/server`.

## Trazabilidad

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-16.1 — se cambian el nombre y el rol, juntos o de a uno; conserva ID y proyecto, y se normaliza | "Cambiar el nombre y el rol de un integrante"; esquema "Cambiar un solo dato reenviando el valor actual del otro" (3 ejemplos); "El nombre se guarda sin espacios en los extremos y el rol normalizado" | `TestUpdate_CambiaElNombreYElRol`; `TestUpdate_CambiaSoloElNombre`; `TestUpdate_CambiaSoloElRol`; `TestUpdate_SinCambiosEsValido`; `TestUpdate_NormalizaElNombreYElRol`; `TestUpdate_ConservaElIDYElProyecto`; `TestUpdate_AceptaUnProyectoFinalizado` |
| CA-16.2 — no se aceptan un nombre vacío ni un rol vacío o inválido | Esquema "No se aceptan un nombre vacío ni un rol vacío o inválido" (4 ejemplos) | `TestUpdate_RechazaNombreORolInvalidos`; `TestUpdate_NombreVacioEnIntegranteInexistenteDevuelveErrEmptyName` |
| CA-16.3 — un integrante inexistente da error | Esquema "Modificar un integrante inexistente falla" (3 ejemplos: 99, 0 y -1) | `TestUpdate_IntegranteInexistenteDevuelveErrMemberNotFound`; `TestUpdate_SinIntegrantesDevuelveErrMemberNotFound` |
| CA-16.4 — sin duplicados, excluyendo al propio integrante | Esquema "El nombre nuevo no puede ser el de otro integrante del proyecto" (3 ejemplos); esquema "Conservar el propio nombre o cambiar sus mayúsculas es válido" (2 ejemplos); "El nombre de un integrante de otro proyecto se acepta" | `TestUpdate_RechazaElNombreDeOtroIntegranteDelProyecto`; `TestUpdate_CambiarLasMayusculasDelPropioNombreEsValido`; `TestUpdate_AceptaElNombreDeUnIntegranteDeOtroProyecto`; `TestModificarIntegrante_NombreDeOtroIntegranteDevuelve409`; `TestModificarIntegrante_PropioNombreConOtrasMayusculasDevuelve200` |
| CA-16.5 — copias, concurrencia y fallas del repositorio | "Cambiar el integrante devuelto no altera lo guardado" (la concurrencia y las fallas, solo con tests) | `TestUpdate_DevuelveUnaCopia`; `TestUpdate_ErrorAlBuscarSePropaga`; `TestUpdate_ErrorAlLeerLosIntegrantesSePropaga`; `TestUpdate_ErrorAlActualizarSePropaga`; `TestUpdate_ModificacionesSimultaneasNoDejanDuplicados`; `TestUpdate_ModificacionYRegistroSimultaneosNoDejanDuplicados` |
| CA-16.6 — mismo comportamiento en memoria y PostgreSQL | Solo contrato de repositorio | `TestMemoryRepository_ContratoUpdate`; `TestPostgresRepository_ContratoUpdate` (7 subtests cada uno) |
| CA-16.7 — API e interfaz | Solo HTTP e interfaz | `TestModificarIntegrante_Devuelve200ConElIntegrante`; `TestModificarIntegrante_DatosInvalidosDevuelven400`; `TestModificarIntegrante_CuerpoInvalidoDevuelve400`; `TestModificarIntegrante_InexistenteDevuelve404`; `TestModificarIntegrante_MetodoNoPermitidoDevuelve405`; `TestModificarIntegrante_GETAnunciaPUTEnAllow`; `TestModificarIntegrante_ErrorDelRepositorioDevuelve500SinDetalle`; la fila "modificar integrante" de `TestRutasConID_IDNoNumericoDevuelve400` y de `bodyEndpoints` (413 y texto después del JSON); `TestInterfaz_SeccionIntegrantesConElFormularioDeModificar`; `TestInterfaz_ModificarIntegranteLlamaALaAPI`; prueba en el navegador |
| RN-16.8 — sin efecto en esfuerzos | Sin escenario | `TestModificarIntegrante_NoCambiaLosEsfuerzosRegistrados` |

La correspondencia no es 1 a 1: varios escenarios y tests cubren más de un criterio, y CA-16.6 y CA-16.7 se verifican solo con tests de contrato, HTTP e interfaz.

## Conteos

```text
$ grep -cE '^[[:space:]]*Escenario:' features/HU-16-modificar-integrante.feature
4
$ grep -cE '^[[:space:]]*Esquema del escenario:' features/HU-16-modificar-integrante.feature
5
$ grep -c '^func Test' internal/member/update_test.go
20
$ grep -c '^func Test' internal/member/update_features_test.go
1
$ grep -c '^func Test' internal/member/repository_update_contract_test.go
2
$ grep -c '^func Test' internal/httpapi/members_update_test.go
12
```

**35 funciones `Test` nuevas** (20 de `Update`, 1 runner de godog, 2 de contrato y 12 de API e interfaz) y **2 filas nuevas** en las listas comunes de `server_test.go`, que suman 3 subtests (ID no numérico, 413 y texto después del JSON). Cubren **4 escenarios y 5 esquemas** BDD; godog expande los esquemas (15 ejemplos) en **19 casos**.

## Fase RED (`9925147`)

Firmas vacías: `ErrMemberNotFound`, `errNotImplemented` ("sin implementar"), `Manager.Update` y `GetByID` y `Update` en los dos repositorios, que devuelven `errNotImplemented`; y los dos métodos nuevos en la interfaz `Repository`. La ruta, el handler, `nameTaken` y la interfaz van en el GREEN. El RED solo agrega esto fuera de los tests y el `.feature`:

```text
$ git diff --stat b69a990 9925147 -- ':!*_test.go' ':!*.feature' ':!docs'
 internal/member/memory_repository_update.go   | 11 +++++++++++
 internal/member/postgres_repository_update.go | 11 +++++++++++
 internal/member/repository.go                 |  9 +++++++++
 internal/member/update.go                     | 14 ++++++++++++++
 4 files changed, 45 insertions(+)
```

En `server_test.go` solo se agregaron las dos líneas de las listas comunes; `members_test.go` y `register_test.go` no cambian.

Medido en una copia temporal con `git worktree` en `9925147`, con go1.27.1: `go vet ./...` sin salida (exit 0) y `GOTMPDIR=C:/gotmp go test -count=1 -p 1 -v ./...`, con los tiempos recortados:

```text
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
FAIL seguimiento-medicion-gpc/internal/httpapi
FAIL seguimiento-medicion-gpc/internal/member
ok   seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
?    seguimiento-medicion-gpc/internal/web [no test files]
```

```text
--- FAIL: TestModificarIntegrante_Devuelve200ConElIntegrante
--- FAIL: TestModificarIntegrante_DatosInvalidosDevuelven400
--- FAIL: TestModificarIntegrante_CuerpoInvalidoDevuelve400
--- FAIL: TestModificarIntegrante_InexistenteDevuelve404
--- FAIL: TestModificarIntegrante_NombreDeOtroIntegranteDevuelve409
--- FAIL: TestModificarIntegrante_PropioNombreConOtrasMayusculasDevuelve200
--- FAIL: TestModificarIntegrante_MetodoNoPermitidoDevuelve405
--- FAIL: TestModificarIntegrante_GETAnunciaPUTEnAllow
--- FAIL: TestModificarIntegrante_ErrorDelRepositorioDevuelve500SinDetalle
--- FAIL: TestModificarIntegrante_NoCambiaLosEsfuerzosRegistrados
--- FAIL: TestInterfaz_SeccionIntegrantesConElFormularioDeModificar
--- FAIL: TestInterfaz_ModificarIntegranteLlamaALaAPI
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400
--- FAIL: TestHandler_CuerpoMayorA1MiBDa413
--- FAIL: TestHandler_TextoDespuesDelJSONDa400
--- FAIL: TestMemoryRepository_ContratoUpdate
--- FAIL: TestFeatures_HU16ModificarIntegrante
--- FAIL: TestUpdate_CambiaElNombreYElRol
--- FAIL: TestUpdate_CambiaSoloElNombre
--- FAIL: TestUpdate_CambiaSoloElRol
--- FAIL: TestUpdate_SinCambiosEsValido
--- FAIL: TestUpdate_CambiarLasMayusculasDelPropioNombreEsValido
--- FAIL: TestUpdate_NormalizaElNombreYElRol
--- FAIL: TestUpdate_RechazaNombreORolInvalidos
--- FAIL: TestUpdate_NombreVacioEnIntegranteInexistenteDevuelveErrEmptyName
--- FAIL: TestUpdate_RechazaElNombreDeOtroIntegranteDelProyecto
--- FAIL: TestUpdate_AceptaElNombreDeUnIntegranteDeOtroProyecto
--- FAIL: TestUpdate_IntegranteInexistenteDevuelveErrMemberNotFound
--- FAIL: TestUpdate_SinIntegrantesDevuelveErrMemberNotFound
--- FAIL: TestUpdate_AceptaUnProyectoFinalizado
--- FAIL: TestUpdate_ConservaElIDYElProyecto
--- FAIL: TestUpdate_DevuelveUnaCopia
--- FAIL: TestUpdate_ErrorAlBuscarSePropaga
--- FAIL: TestUpdate_ErrorAlLeerLosIntegrantesSePropaga
--- FAIL: TestUpdate_ErrorAlActualizarSePropaga
--- FAIL: TestUpdate_ModificacionesSimultaneasNoDejanDuplicados
--- FAIL: TestUpdate_ModificacionYRegistroSimultaneosNoDejanDuplicados
```

Fallan **37 tests de primer nivel** y 55 subtests: **34 nuevos** (20 de `Update`, el contrato en memoria, el runner de godog y 12 de API e interfaz) y **3 existentes** por la fila nueva (`TestRutasConID_IDNoNumericoDevuelve400`, `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400`, cada uno solo por su subtest `modificar_integrante`). De las 35 funciones nuevas, la que no falla es `TestPostgresRepository_ContratoUpdate`, que se omite sin base. `grep -c '^--- FAIL'` da 37 y `grep -ciE 'build failed|undefined:|cannot use|\[setup failed\]|panic'` da 0. Pasan 954 tests de primer nivel y se omiten 41, así que los tests anteriores siguen en verde. Todos los fallos son de aserción:

- dominio y contrato: los stubs devuelven "sin implementar" (por ejemplo, `err = sin implementar, se esperaba no existe un integrante con ese id`);
- API: la ruta todavía no existe (`código = 405, se esperaba 200; cuerpo: Method Not Allowed`) y `GET` no responde 405 ni trae `Allow`;
- interfaz: falta el formulario en `index.html` y su código en `app.js`.

Godog de HU-16 en el RED: `19 scenarios (19 failed)` y `92 steps (62 passed, 19 failed, 11 skipped)`. Los subtests `POST`, `PATCH` y `DELETE` del test de 405 pasan ya en el RED (ver "Desvíos y límites").

## Fase GREEN (`5738455`)

- `internal/member/update.go`: `Manager.Update` valida nombre y rol (reutiliza `normalizeRole`), y bajo `m.mu` busca con `GetByID`, lee los integrantes del proyecto, descarta duplicados con `nameTaken` y guarda con `Update`. Se eliminó `errNotImplemented`.
- `internal/member/register.go`: el bucle de duplicados pasa a `nameTaken(members, name, 0)`, sin cambiar el comportamiento ni los tests de HU-15.
- `internal/member/memory_repository_update.go` y `postgres_repository_update.go`: `GetByID` y `Update`. En PostgreSQL, `UPDATE members SET name = $2, role = $3 WHERE id = $1 RETURNING …`, y `pgx.ErrNoRows` se traduce a `ErrMemberNotFound`.
- `internal/httpapi/members_update.go`: `updateMember` y `memberMethodNotAllowed`.
- `internal/httpapi/server.go`: las rutas `PUT` y `GET` de `/api/integrantes/{id}` y `member.ErrMemberNotFound` en el 404 de `statusFor`.
- `internal/web/index.html` y `internal/web/app.js`: el formulario `form-modificar-integrante` y su listener, como bloque aparte.

El GREEN no modifica tests ni el `.feature`:

```text
$ git diff --stat 9925147 5738455 -- '*_test.go' '*.feature'
(sin salida)
```

Salida sobre `1775b40` (el GREEN con `main` `c313127` adentro): `gofmt -l .`, `go vet ./...` y `go build ./...` sin salida, y `GOTMPDIR=C:/gotmp go test -count=1 -p 1 ./...` con exit 0, sin base:

```text
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
ok   seguimiento-medicion-gpc/internal/httpapi
ok   seguimiento-medicion-gpc/internal/member
ok   seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
?    seguimiento-medicion-gpc/internal/web [no test files]
```

Con base (`TEST_DATABASE_URL=postgres://gpc:gpc_dev@localhost:5433/gpc_test?sslmode=disable`, PostgreSQL 17.11) el mismo comando dio exit 0, con los mismos 14 paquetes en `ok`.

| Corrida con `-v ./...` sobre `1775b40` | PASS | FAIL | SKIP |
|---|---|---|---|
| Sin base | 1049 | 0 | 41 |
| Con base | 1090 | 0 | 0 |

Son tests de primer nivel; con base tampoco hay subtests en SKIP. Godog de HU-16: `19 scenarios (19 passed)` y `92 steps (92 passed)`.

Subtests del contrato contra PostgreSQL, sobre `5738455`:

```text
--- PASS: TestPostgresRepository_ContratoUpdate
    --- PASS: .../GetByID_devuelve_el_integrante_guardado
    --- PASS: .../GetByID_devuelve_ErrMemberNotFound_si_no_existe
    --- PASS: .../GetByID_devuelve_una_copia
    --- PASS: .../Update_reemplaza_el_nombre_y_el_rol_y_conserva_el_proyecto
    --- PASS: .../Update_no_cambia_a_los_demás_integrantes
    --- PASS: .../Update_devuelve_ErrMemberNotFound_si_no_existe
    --- PASS: .../Update_devuelve_una_copia
```

## Refactor

No se hizo.

## Prueba en el navegador

Con el servidor en memoria (`go run ./cmd/server`, puerto 8080) sobre `5738455`, se cargó un proyecto y se completaron y enviaron los formularios por script con `requestSubmit()`, leyendo `msg-integrantes` del DOM:

| Acción | Mensaje |
|---|---|
| Registrar a Ana y a Bruno | "Integrante 1 registrado: Ana (Product Builder)." y "Integrante 2 registrado: Bruno (Agile Enabler)." |
| Modificar el 1: Ana Pérez, Agile Enabler | "Integrante 1 modificado: Ana Pérez (Agile Enabler)." |
| Nombre duplicado: `BRUNO` | Error: "ya hay un integrante con ese nombre en el proyecto" |
| Solo mayúsculas del propio nombre: `ANA PÉREZ` | "Integrante 1 modificado: ANA PÉREZ (Agile Enabler)." |
| Rol inválido `SCRUM_MASTER` | Error: "el rol debe ser PRODUCT_ARCHITECT, AGILE_ENABLER o PRODUCT_BUILDER" |
| ID 99 | Error: "no existe un integrante con ese id" |
| Nombre con solo espacios | Error: "el nombre del integrante es obligatorio" |
| ID vacío | El navegador no envía el formulario (`checkValidity()` da `false`) |

Después del merge se verificó el servidor sobre `1775b40` con curl: `index.html` y `app.js` conservan los formularios de los dos integrantes (`form-registrar-integrante` y `form-modificar-integrante`), el de HU-19 (`form-modificar-fechas-proyecto`) y el de HU-28 (`form-corregir-esfuerzo`). `PUT /api/integrantes/1` responde 200, `PUT /api/integrantes/99` responde 404 y `PATCH /api/esfuerzos/99/horas` responde 404. La tabla no se repitió en el navegador después del merge.

## Cobertura

Medición local en Windows con go1.27.1, sobre `1775b40`, y en `main` (`c313127`) con el mismo Go y los mismos comandos, en una copia temporal con `git worktree`. Sin base: `GOTMPDIR=C:/gotmp go test -count=1 -p 1 -coverprofile=cover.out ./...` y `go tool cover -func=cover.out`.

| Paquete | Rama | `main` (`c313127`) |
|---|---|---|
| `cmd/server` | 29.8% | 29.8% |
| `internal/backlog` | 83.3% | 83.3% |
| `internal/dashboard` | 100.0% | 100.0% |
| `internal/defect` | 77.7% | 77.7% |
| `internal/effort` | 74.2% | 74.2% |
| `internal/httpapi` | 98.7% | 98.7% |
| `internal/member` | 65.5% | 60.3% |
| `internal/metrics` | 100.0% | 100.0% |
| `internal/poker` | 75.5% | 75.5% |
| `internal/postgres` | 83.3% | 83.3% |
| `internal/postgrestest` | 36.8% | 36.8% |
| `internal/project` | 74.8% | 74.8% |
| `internal/report` | 100.0% | 100.0% |
| `internal/sprint` | 77.1% | 77.1% |
| `internal/web` | 0.0% | 0.0% |
| **Total del perfil** | **84.3%** | **84.5%** |

```text
internal/httpapi/members_update.go:9:    updateMember            100.0%
internal/httpapi/members_update.go:33:   memberMethodNotAllowed  100.0%
internal/httpapi/server.go:161:          statusFor               100.0%
internal/member/memory_repository_update.go:5:   GetByID         100.0%
internal/member/memory_repository_update.go:17:  Update          100.0%
internal/member/postgres_repository_update.go:13: GetByID        0.0%
internal/member/postgres_repository_update.go:29: Update         0.0%
internal/member/register.go:9:           Register                100.0%
internal/member/update.go:15:            Update                  100.0%
internal/member/update.go:55:            nameTaken               100.0%
total:                                   (statements)            84.3%
```

Solo cambia `internal/member`, de 60.3% a 65.5%; los demás paquetes conservan sus porcentajes. El total del perfil baja de 84.5% a 84.3% porque `GetByID` y `Update` de PostgreSQL suman sentencias que sin base no se ejecutan.

Con base, con los mismos comandos y `TEST_DATABASE_URL`: `internal/member` 95.6% y total del perfil 95.1%. Con `-coverpkg=./...`, el total es **96.1%** en la rama y en `main`, y `GetByID` y `Update` de PostgreSQL quedan en **85.7%**.

## Desvíos y límites

1. **`unlinkat` de Windows y método de medición.** Todas las mediciones usan `GOTMPDIR=C:/gotmp` con `-p 1`, como en HU-24. En algunas corridas Go termina con `go: unlinkat … The process cannot access the file because it is being used by another process` cuando todos los paquetes ya informaron `ok`: da exit 1 sin fallos de tests, y la corrida se repitió después de `go clean -testcache`.
2. **Sin `-race`.** Localmente responde `-race requires cgo`. Los tests de concurrencia pasan sin el detector de carreras; lo ejecuta el CI.
3. **El CI no ejecuta los tests de PostgreSQL.** Allí se omiten, porque no tiene `TEST_DATABASE_URL`. El SQL de `GetByID` y `Update` se verificó a mano con PostgreSQL 17.11 en Docker (`postgres:17-alpine`, puerto 5433), con 0 SKIP. La base `gpc_test` se recreó con `docker compose down -v` porque el volumen era anterior a HU-15 y no tenía la tabla `members`. `GetByID` y `Update` de PostgreSQL quedan en 85.7% por la rama de error genérico de la base.
4. **Rol inválido en el navegador.** La lista solo tiene los tres roles, así que la prueba se hizo agregando una opción falsa por script, que después se quitó.
5. **Título pegado al botón.** El título "Modificar integrante" queda pegado al botón "Registrar integrante". No se cambió, porque el formulario de HU-15 no se toca.
6. **Test fuera de "Pruebas previstas".** `TestModificarIntegrante_NoCambiaLosEsfuerzosRegistrados` no figura en esa sección de la SDD. Cubre RN-16.8 y su caso límite.
7. **`nameTaken` en el GREEN.** La SDD decía en "Restricciones" que iba en un refactor y fue en el GREEN. La frase se corrigió en la SDD en el commit de esta evidencia.
8. **Merge con `main`.** Se trajo `main` con HU-19 y HU-28 (`c313127`, merge `1775b40`). El único conflicto fue en `internal/httpapi/server.go`, en el caso 404 de `statusFor`, y se resolvió conservando los tres errores (`defect.ErrDefectNotFound`, `member.ErrMemberNotFound` y `effort.ErrEffortNotFound`).
9. **Subtests de 405 en verde desde el RED.** `POST`, `PATCH` y `DELETE` en `/api/integrantes/1` ya responden 405, porque la ruta `GET /` de la interfaz coincide con esa ruta y `http.ServeMux` rechaza los demás métodos. El subtest `GET` sí falla en el RED.
10. **CA-16.5 a CA-16.7 en el `.feature`.** Los pasos de godog llaman al `Manager`, así que la concurrencia, las fallas del repositorio, la persistencia y la API e interfaz se prueban con tests de Go, como en HU-15.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `b69a990` | `docs(HU-16): agregar la especificación (SDD)` |
| RED | `9925147` | `test(HU-16): agregar tests (RED)` |
| GREEN | `5738455` | `feat(HU-16): modificar el nombre y el rol de un integrante (GREEN)` |
| Traer `main` | `1775b40` | `chore(HU-16): traer main con HU-19 y HU-28` |
| Evidencia y cobertura | `e51d920` | `docs(HU-16): agregar evidencia TDD y cobertura` |

HU-16: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
