# Evidencia TDD · HU-15 Registrar integrante del proyecto

> Issue del repositorio: #29

- **Especificación SDD:** [`docs/especificaciones/HU-15-registrar-integrante.md`](../especificaciones/HU-15-registrar-integrante.md), aprobada por Peñalbé Hernán el 29/10/2026.
- **Escenarios BDD:** [`features/HU-15-registrar-integrante.feature`](../../features/HU-15-registrar-integrante.feature): 7 definiciones (3 escenarios y 4 esquemas, con 3, 4, 3 y 3 ejemplos), que godog ejecuta como 17 escenarios y 64 pasos.
- **Rama:** `feature/HU-15-registrar-integrante`.
- **Story Points:** 5 (Planning del Sprint 2).
- **Dependencias:** HU-01 y HU-03, en `main`. E4 estaba fuera de la entrega en los Planning de los Sprints 3 y 4 "salvo capacidad sobrante"; se tomó el 29/10/2026 con capacidad sobrante.

## Trazabilidad

**27 tests nuevos (funciones `Test`) cubriendo 17 escenarios BDD ejecutados (7 definiciones):** 14 unitarios en `internal/member/register_test.go`, el runner de godog en `internal/member/register_features_test.go`, el contrato del repositorio en memoria (`memory_repository_test.go`, que corre los 5 casos de `repository_contract_test.go`), 3 de PostgreSQL en `postgres_repository_test.go`, 6 de la API en `internal/httpapi/members_test.go` y 2 de la interfaz en `internal/httpapi/web_test.go`. Además, casos nuevos en tests existentes: la ruta en `TestRutasConID_IDNoNumericoDevuelve400`, el endpoint en `bodyEndpoints` (que usan `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400`) y el repositorio de integrantes en las dos pruebas de armado de `cmd/server/main_test.go`. La correspondencia no es 1 a 1: CA-15.6 y CA-15.7 se cubren solo con tests de Go.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-15.1 Registro con nombre y rol válido | "Registrar un integrante con nombre y rol", "Se acepta cada uno de los tres roles" (3), "El nombre se guarda sin espacios en los extremos" | `TestRegister_GuardaElIntegranteConNombreYRol`, `TestRegister_AceptaLosTresRolesConIDsCorrelativos`, `TestRegister_NormalizaElNombreYElRol`, `TestRegister_AceptaDosIntegrantesConElMismoRol`, `TestRegister_AceptaUnProyectoFinalizado` |
| CA-15.2 Nombre y rol obligatorios, rol válido | "Nombre y rol son obligatorios y el rol tiene que ser válido" (4) | `TestRegister_RechazaNombreORolInvalidos`, `TestRegister_NombreVacioEnProyectoInexistenteDevuelveErrEmptyName`, `TestRegistrarIntegrante_DatosInvalidosDevuelven400` |
| CA-15.3 Sin duplicados en el proyecto | "No se registra dos veces al mismo integrante en el proyecto" (3), "El mismo nombre en otro proyecto se acepta" | `TestRegister_RechazaElMismoNombreEnElProyecto`, `TestRegister_AceptaElMismoNombreEnOtroProyecto`, `TestRegistrarIntegrante_DuplicadoDevuelve409` |
| CA-15.4 Proyecto inexistente | "Registrar en un proyecto inexistente falla" (3) | `TestRegister_ProyectoInexistenteDevuelveErrProjectNotFound`, `TestRegistrarIntegrante_ProyectoInexistenteDevuelve404` |
| CA-15.5 Copia, concurrencia y fallas | "Modificar el integrante devuelto no cambia lo guardado" | `TestRegister_DevuelveUnaCopia`, `TestRegister_RegistrosSimultaneosNoDejanDuplicados`, `TestRegister_ErrorAlLeerLosIntegrantesSePropaga`, `TestRegister_ErrorAlGuardarSePropaga`, `TestRegistrarIntegrante_ErrorDelRepositorioDevuelve500SinDetalle` |
| CA-15.6 Memoria y PostgreSQL | — | `TestMemoryRepository_Contrato`, `TestPostgresRepository_Contrato`, `TestPostgresRepository_GuardaElIntegrante`, `TestPostgresRepository_RechazaUnProyectoInexistente`, y los casos de armado en `cmd/server/main_test.go` |
| CA-15.7 API e interfaz | — | `TestRegistrarIntegrante_Devuelve201ConElIntegrante`, `TestRegistrarIntegrante_IDNoNumericoDevuelve400`, los casos de `TestRutasConID_IDNoNumericoDevuelve400`, `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400`, `TestInterfaz_SeccionIntegrantesConElFormularioDeRegistro`, `TestInterfaz_RegistrarIntegranteLlamaALaAPIConElProyectoActual` |

## Fase RED (`a9c1790`)

El RED agrega el `.feature`, los tests y las firmas vacías del paquete nuevo `internal/member`: `Member`, `Role`, los tres roles y los cuatro errores; `Manager` y `NewManager`; `Register`, que devuelve `errNotImplemented`; la interfaz `Repository`, y los dos repositorios, que también devuelven `errNotImplemented`. Como pide la SDD, el RED incluye el cambio mecánico de `NewHandler`, que suma el parámetro `members`: sus 24 llamadas (23 en los tests de `internal/httpapi`, con el helper `newMembers`, y la de `cmd/server/main.go`) y el armado del repositorio de integrantes en `cmd/server/repositories.go`, para que todo compile y el GREEN no toque tests. La tabla `members`, la ruta, el handler y la interfaz no están en el RED.

Sobre `a9c1790`, en Linux con go1.27.1, sin base:

```
$ go vet ./...
(sin salida)

$ go test -count=1 ./...
ok   seguimiento-medicion-gpc/cmd/server
ok   seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
ok   seguimiento-medicion-gpc/internal/defect
ok   seguimiento-medicion-gpc/internal/effort
--- FAIL: TestRegistrarIntegrante_Devuelve201ConElIntegrante (0.00s)
--- FAIL: TestRegistrarIntegrante_DatosInvalidosDevuelven400 (0.00s)
--- FAIL: TestRegistrarIntegrante_ProyectoInexistenteDevuelve404 (0.00s)
--- FAIL: TestRegistrarIntegrante_IDNoNumericoDevuelve400 (0.00s)
--- FAIL: TestRegistrarIntegrante_DuplicadoDevuelve409 (0.00s)
--- FAIL: TestRegistrarIntegrante_ErrorDelRepositorioDevuelve500SinDetalle (0.00s)
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.01s)
    --- FAIL: TestRutasConID_IDNoNumericoDevuelve400/registrar_integrante (0.00s)
--- FAIL: TestHandler_CuerpoMayorA1MiBDa413 (0.13s)
    --- FAIL: TestHandler_CuerpoMayorA1MiBDa413/registrar_integrante (0.00s)
--- FAIL: TestHandler_TextoDespuesDelJSONDa400 (0.00s)
    --- FAIL: TestHandler_TextoDespuesDelJSONDa400/registrar_integrante (0.00s)
--- FAIL: TestInterfaz_SeccionIntegrantesConElFormularioDeRegistro (0.00s)
--- FAIL: TestInterfaz_RegistrarIntegranteLlamaALaAPIConElProyectoActual (0.00s)
FAIL seguimiento-medicion-gpc/internal/httpapi
--- FAIL: TestMemoryRepository_Contrato (0.00s)
    repository_contract_test.go:21: Save({ID:0 ProjectID:1 Name:Ana Pérez Role:PRODUCT_BUILDER}) devolvió HU-15 todavía no está implementada
--- FAIL: TestFeatures_HU15RegistrarIntegrante (0.02s)
--- FAIL: TestRegister_GuardaElIntegranteConNombreYRol (0.00s)
--- FAIL: TestRegister_AceptaLosTresRolesConIDsCorrelativos (0.00s)
--- FAIL: TestRegister_NormalizaElNombreYElRol (0.00s)
--- FAIL: TestRegister_RechazaNombreORolInvalidos (0.00s)
--- FAIL: TestRegister_RechazaElMismoNombreEnElProyecto (0.00s)
--- FAIL: TestRegister_AceptaElMismoNombreEnOtroProyecto (0.00s)
--- FAIL: TestRegister_AceptaDosIntegrantesConElMismoRol (0.00s)
--- FAIL: TestRegister_AceptaUnProyectoFinalizado (0.00s)
--- FAIL: TestRegister_ProyectoInexistenteDevuelveErrProjectNotFound (0.00s)
--- FAIL: TestRegister_NombreVacioEnProyectoInexistenteDevuelveErrEmptyName (0.00s)
--- FAIL: TestRegister_DevuelveUnaCopia (0.00s)
--- FAIL: TestRegister_ErrorAlLeerLosIntegrantesSePropaga (0.00s)
--- FAIL: TestRegister_ErrorAlGuardarSePropaga (0.00s)
--- FAIL: TestRegister_RegistrosSimultaneosNoDejanDuplicados (0.00s)
FAIL seguimiento-medicion-gpc/internal/member
ok   seguimiento-medicion-gpc/internal/metrics
ok   seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/report
ok   seguimiento-medicion-gpc/internal/sprint
```

godog de HU-15 en el RED: `17 scenarios (17 failed)`, `64 steps (31 passed, 17 failed, 16 skipped)`. En `TestRutasConID_IDNoNumericoDevuelve400`, `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400` falla solo el caso nuevo "registrar_integrante", porque la ruta todavía no existe. Los 3 tests de PostgreSQL se omiten sin base. Las pruebas de armado de `cmd/server` pasan desde el RED, porque el armado entró en el RED para que compile (ver arriba). Todos los fallos son de HU-15: los demás paquetes pasan.

## Corrección de dos tests del RED (`058804f`)

Al correr el GREEN, `TestRegistrarIntegrante_Devuelve201ConElIntegrante` y el caso "registrar integrante" de `bodyEndpoints` seguían fallando con `código = 201, se esperaba 200`: comparaban el cuerpo con `expectRawJSON`, que además exige el código 200, y el registro responde 201, como dice la SDD. Se cambiaron por `expectMemberJSON`, un helper nuevo que compara solo el cuerpo. El código 201 se sigue verificando con `expectStatus`, y no se aflojó ninguna otra aserción. La corrección va en un commit propio, antes del GREEN:

```
$ git diff --stat a9c1790 058804f
 internal/httpapi/members_test.go | 22 +++++++++++++++++++++-
 internal/httpapi/server_test.go  |  2 +-
```

## Fase GREEN (`358e427`)

- `internal/member/register.go`: `Register` valida el nombre (`TrimSpace`, `ErrEmptyName`), el rol (`normalizeRole`: mayúsculas y sin espacios en los extremos; vacío es `ErrMissingRole`, otro valor `ErrInvalidRole`) y el proyecto con `project.Manager.Get`, antes del mutex. Con el mutex tomado, lee los integrantes del proyecto, rechaza el mismo nombre con `strings.EqualFold` (`ErrDuplicateMember`) y guarda.
- `internal/member/memory_repository.go` y `postgres_repository.go`: `Save` y `ListByProject`, con IDs correlativos y listas nuevas, nunca `nil`. Se saca `errNotImplemented` de `manager.go`.
- `db/init/01-schema.sql`: tabla `members` con clave foránea a `projects` e índice por proyecto. `internal/postgrestest`: `members` en el `TRUNCATE`.
- `internal/httpapi/members.go` (nuevo) y `server.go`: `POST /api/proyectos/{id}/integrantes` y los errores en `statusFor` (400 para nombre, rol vacío y rol inválido; 409 para el duplicado).
- `internal/web`: la sección "Integrantes" con el formulario "Registrar integrante" y los tres roles.
- `README.md`: el endpoint, un `curl`, la sección nueva de la interfaz, el 409 y el aviso de recrear la base por la tabla `members`.

El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat 058804f 358e427 -- '*_test.go' '*.feature'
(sin salida)
```

## Refactor

No hubo. `Register` sigue la forma de `project.Manager.Create` y de `sprint.Manager.Create`, y los repositorios la de los otros paquetes.

## Desviaciones del ciclo TDD

1. **Dos tests del RED corregidos antes del GREEN** (`058804f`), explicados arriba: el helper que usaban exigía el código 200. No es un cambio de comportamiento.
2. **El armado de `cmd/server` y la firma de `NewHandler` entraron en el RED**, como pide la SDD, para que los tests compilen. No es lógica de HU-15: la ruta y el handler se agregaron en el GREEN.

## Prueba de punta a punta

Con el servidor real (`go build ./cmd/server`, puerto 8093), un proyecto "Sistema de ventas" creado con `curl` y la interfaz en Chromium con Playwright, en la sección "Integrantes":

- Sin proyecto actual: "Elegí un proyecto: creá uno o escribí su ID en «Proyecto actual».".
- Con el proyecto 1, "Ana Pérez" como Agile Enabler: "Integrante 1 registrado: Ana Pérez (Agile Enabler)."; el campo del nombre queda vacío.
- "ana pérez" otra vez: "ya hay un integrante con ese nombre en el proyecto".
- Por la API, el rol `scrum_master`: 400 con `{"error":"el rol debe ser PRODUCT_ARCHITECT, AGILE_ENABLER o PRODUCT_BUILDER"}`.

## Validación final

Sobre `358e427`, en Linux con go1.27.1:

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)

$ go test -count=1 -cover ./...
ok   seguimiento-medicion-gpc/cmd/server             coverage: 29.8% of statements
ok   seguimiento-medicion-gpc/internal/backlog       coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/dashboard     coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/defect        coverage: 77.7% of statements
ok   seguimiento-medicion-gpc/internal/effort        coverage: 75.2% of statements
ok   seguimiento-medicion-gpc/internal/httpapi       coverage: 98.6% of statements
ok   seguimiento-medicion-gpc/internal/member        coverage: 60.3% of statements
ok   seguimiento-medicion-gpc/internal/metrics       coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/poker         coverage: 75.5% of statements
ok   seguimiento-medicion-gpc/internal/postgres      coverage: 83.3% of statements
ok   seguimiento-medicion-gpc/internal/postgrestest  coverage: 36.8% of statements
ok   seguimiento-medicion-gpc/internal/project       coverage: 71.1% of statements
ok   seguimiento-medicion-gpc/internal/report        coverage: 100.0% of statements
ok   seguimiento-medicion-gpc/internal/sprint        coverage: 77.1% of statements
     seguimiento-medicion-gpc/internal/web           coverage: 0.0% of statements
```

godog de HU-15: `17 scenarios (17 passed)`, `64 steps (64 passed)`.

`go tool cover -func`: `Register`, `normalizeRole`, `NewManager`, el repositorio en memoria y `registerMember` en 100.0%. El repositorio PostgreSQL de integrantes queda en 0.0% sin base, porque sus tests se omiten sin `TEST_DATABASE_URL`; por eso `internal/member` queda en 60.3% sin base. Total del perfil sin base: **84.3%** (`main` en `af48b68`, con HU-14, mismo Go y comando, en una copia temporal con `git worktree`: **85.0%**). La baja se explica por el repositorio PostgreSQL nuevo, que sin base no se ejecuta, y por las líneas de armado de PostgreSQL en `cmd/server` (31.5% → 29.8%). `internal/httpapi` sube de 98.5% a 98.6%.

### Con base (PostgreSQL 16.15) y `-race`

Con PostgreSQL 16.15 instalado en el entorno de trabajo, no con el contenedor `postgres:17-alpine` de `docker-compose.yml` (no hay Docker en ese entorno). La base `gpc_test` se recreó con el mismo `db/init/01-schema.sql`, con la tabla `members`:

```
$ TEST_DATABASE_URL=... go test -race -count=1 -p 1 -coverpkg=./... -coverprofile=cover.out ./...
ok   (los 14 paquetes)
$ go tool cover -func=cover.out | tail -n 1
total:                                  (statements)            96.0%
$ TEST_DATABASE_URL=... go test -count=1 -cover ./internal/member/
ok   seguimiento-medicion-gpc/internal/member        coverage: 95.2% of statements
```

Sin SKIP (0 tests omitidos con `go test -v ./internal/member/`) y sin carreras. Con base, `internal/member` queda en 95.2%: lo que falta son las ramas de error de la base en `ListByProject`.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | 0f0209f | `docs(HU-15): agregar especificación SDD` |
| RED | a9c1790 | `test(HU-15): agregar tests (RED)` |
| Corrección de tests del RED | 058804f | `test(HU-15): corregir dos tests de la API del RED` |
| GREEN | 358e427 | `feat(HU-15): registrar un integrante del proyecto (GREEN)` |
| Refactor | — | No hubo |
| Evidencia y cobertura | (propuesto) | `docs(HU-15): agregar evidencia TDD y cobertura` |

HU-15: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
