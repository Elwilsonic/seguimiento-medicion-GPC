# Evidencia TDD · HU-51 Identificadores UUID (parte de proyectos)

> Issue del repositorio: #193 · Épica E1 · Vinculada a HU-01 (#14), HU-03 (#16) y HU-50 (#183) · Especificada con OpenSpec (TEC-10, #177)

- **Historia:** como integrante del equipo, quiero que los proyectos se identifiquen con un UUID en lugar de un número correlativo, para que no se pueda adivinar el identificador de un proyecto ajeno ni deducir cuántos proyectos hay.
- **Especificación (OpenSpec):** la spec de la capability `proyectos`, [`openspec/specs/proyectos/spec.md`](../../openspec/specs/proyectos/spec.md), con los 7 requisitos de HU-50 y los 9 de HU-51. El cambio está archivado en [`openspec/changes/archive/2026-11-15-hu-51-identificadores-uuid/`](../../openspec/changes/archive/2026-11-15-hu-51-identificadores-uuid/): [propuesta](../../openspec/changes/archive/2026-11-15-hu-51-identificadores-uuid/proposal.md), [spec delta](../../openspec/changes/archive/2026-11-15-hu-51-identificadores-uuid/specs/proyectos/spec.md), [diseño](../../openspec/changes/archive/2026-11-15-hu-51-identificadores-uuid/design.md) y [tareas](../../openspec/changes/archive/2026-11-15-hu-51-identificadores-uuid/tasks.md). Se archivó con `/opsx:archive` el 15/11/2026, con el PR #195 integrado en `main`.
- **Aprobación:** Peñalbé Hernán (Agile Enabler) aprobó la propuesta el 14/11/2026, con la alternativa B (D1: UUID público y ID entero interno), D4, el pendiente R1 y los criterios CA-51.1 a CA-51.7. Pidió dos arreglos, que se aplicaron en `64a59db`: los links a HU-50, que ya estaba archivada, y la capability `proyectos` como modificada. Fijó una prioridad: si no llegaba todo para la Review del 15/11, primero la API de proyectos con UUID (CA-51.1 a CA-51.4). Llegó todo.
- **Escenarios BDD:** [`features/HU-51-identificadores-uuid.feature`](../../features/HU-51-identificadores-uuid.feature): 12 escenarios y un esquema con 3 ejemplos, que dan 15 ejecuciones y 67 pasos.
- **Rama:** `feature/HU-51-uuid`, creada desde `main` `315fac5`, con `main` traído en `ce8caba` y `93bd16d` (`de928fa`, con el archivo de HU-50 y el control de TEC-11).
- **Story Points:** 13, en el Planning del Sprint 5 (Issue #182, PR #184).
- **Pendiente aceptado (R1):** ítems, sprints, integrantes, defectos, esfuerzos y sesiones de Planning Poker siguen con IDs enteros. Se informa en la Review del 15/11.

## Trazabilidad

**31 tests nuevos (funciones `Test`) y 15 ejecuciones de escenarios BDD.** Se reparten así:

- 9 de dominio en `internal/project/public_id_test.go`;
- 3 de repositorio: el contrato en memoria (`repository_contract_test.go`) y 2 de PostgreSQL (`postgres_repository_test.go`);
- 12 de API en `internal/httpapi/projects_uuid_test.go`;
- 4 internos del paquete en `internal/httpapi/project_id_internal_test.go`, porque `parseProjectID` no se exporta;
- 2 de interfaz en `internal/httpapi/projects_uuid_web_test.go`;
- el runner de godog en `internal/httpapi/uuid_features_test.go`, que ejecuta los 15 escenarios.

Además, **151 tests existentes se adaptaron** sin cambiar lo que verifican: 144 de `internal/httpapi`, 5 de `internal/project` y 2 de `internal/report`. Los de `httpapi` arman las rutas con el `id` que devuelve la API y comparan `id` y `proyecto_id` sin depender de su tipo. 36 de ellos cambian su resultado esperado: un UUID inexistente o el nulo dan 404, y un `{id}` de proyecto que no es un UUID da 400 con el mensaje nuevo. Los 7 de `project` y `report` comparan el proyecto completo, también el `PublicID`.

La correspondencia no es 1 a 1:

- los escenarios cubren la API;
- los 1000 proyectos (CA-51.2), PostgreSQL (CA-51.5), la interfaz (CA-51.6) y la concurrencia (CA-51.7) se cubren solo con tests de Go;
- varios tests comparten criterio.

Cada test nuevo lleva arriba el comentario `// HU-51 · CA-51.N · Verifica que …`, y el runner `// HU-51 · BDD · Verifica que …`.

### Criterio → test → código

| Criterio | Escenarios BDD | Tests de Go (archivo) | Código que lo implementa (función y archivo) |
|---|---|---|---|
| CA-51.1 Un proyecto nuevo recibe un identificador público UUID v4 en forma canónica, y la API ya no muestra 1, 2, 3 | "Crear un proyecto le asigna un UUID versión 4", "El primer proyecto no es el 1" (runner [`TestFeatures_HU51IdentificadoresUUID`](../../internal/httpapi/uuid_features_test.go#L23), `internal/httpapi/uuid_features_test.go`) | [`TestCreate_AsignaUnUUIDVersion4`](../../internal/project/public_id_test.go#L32), [`TestCreate_ElPrimerProyectoNoTieneElUUIDNulo`](../../internal/project/public_id_test.go#L47), [`TestCreate_GuardaElUUIDAsignado`](../../internal/project/public_id_test.go#L64) (`internal/project/public_id_test.go`); [`TestCrearProyecto_IDEsUUIDVersion4`](../../internal/httpapi/projects_uuid_test.go#L71) (`internal/httpapi/projects_uuid_test.go`) | [`Manager.Create`](../../internal/project/create.go#L9) con `uuid.NewV4()` ([`internal/project/create.go`](../../internal/project/create.go)); el campo `PublicID` de [`Project`](../../internal/project/project.go#L20) ([`internal/project/project.go`](../../internal/project/project.go)); [`toProjectResponse`](../../internal/httpapi/projects.go#L200) ([`internal/httpapi/projects.go`](../../internal/httpapi/projects.go)) |
| CA-51.2 Dos proyectos nunca tienen el mismo identificador público, y no se puede deducir otro ni la cantidad | — | [`TestCreate_MilProyectosConUUIDDistintosYNoCorrelativos`](../../internal/project/public_id_test.go#L83) (`internal/project/public_id_test.go`); [`TestCrearProyecto_DosProyectosTienenUUIDDistintos`](../../internal/httpapi/projects_uuid_test.go#L87) (`internal/httpapi/projects_uuid_test.go`) | [`Manager.Create`](../../internal/project/create.go#L9) ([`internal/project/create.go`](../../internal/project/create.go)); `public_id UUID NOT NULL UNIQUE` ([`db/init/01-schema.sql`](../../db/init/01-schema.sql)) |
| CA-51.3 Todas las rutas de proyecto aceptan su UUID, y ninguna respuesta expone el ID entero (`id`, `proyecto_id` y el nombre del PDF llevan el UUID) | "Consultar un proyecto por su UUID", "El UUID se acepta en mayúsculas", "Las operaciones que dependen del proyecto usan su UUID", "Un ítem consultado por su propio ID muestra el UUID de su proyecto" (runner [`TestFeatures_HU51IdentificadoresUUID`](../../internal/httpapi/uuid_features_test.go#L23)) | [`TestRutasDeProyecto_AceptanElUUID`](../../internal/httpapi/projects_uuid_test.go#L103), [`TestRutasDeProyecto_AceptanElUUIDEnMayusculas`](../../internal/httpapi/projects_uuid_test.go#L120), [`TestRespuestas_IdentificanElProyectoConSuUUID`](../../internal/httpapi/projects_uuid_test.go#L275), [`TestReportePDF_NombreDelArchivoConElUUID`](../../internal/httpapi/projects_uuid_test.go#L339) (`internal/httpapi/projects_uuid_test.go`); [`TestParseProjectID_AceptaMayusculas`](../../internal/httpapi/project_id_internal_test.go#L28) (`internal/httpapi/project_id_internal_test.go`); [`TestGetByPublicID_DevuelveElProyecto`](../../internal/project/public_id_test.go#L142), [`TestGetByPublicID_DevuelveUnaCopia`](../../internal/project/public_id_test.go#L163), [`TestModificar_ConservaElUUID`](../../internal/project/public_id_test.go#L205) (`internal/project/public_id_test.go`) | [`server.projectIDFromPath`](../../internal/httpapi/server.go#L192) y [`server.publicProjectID`](../../internal/httpapi/server.go#L212) ([`internal/httpapi/server.go`](../../internal/httpapi/server.go)); [`Manager.GetByPublicID`](../../internal/project/get.go#L21) ([`internal/project/get.go`](../../internal/project/get.go)); [`toItemResponse`](../../internal/httpapi/items.go#L206), [`server.writeItem`](../../internal/httpapi/items.go#L177) y [`server.writeItems`](../../internal/httpapi/items.go#L189) ([`internal/httpapi/items.go`](../../internal/httpapi/items.go)); [`toSprintResponse`](../../internal/httpapi/sprints.go#L171) ([`internal/httpapi/sprints.go`](../../internal/httpapi/sprints.go)); [`toProjectMetricsResponse`](../../internal/httpapi/metrics.go#L125) ([`internal/httpapi/metrics.go`](../../internal/httpapi/metrics.go)); [`toSummaryResponse`](../../internal/httpapi/dashboard.go#L40) ([`internal/httpapi/dashboard.go`](../../internal/httpapi/dashboard.go)); [`toProjectReportResponse`](../../internal/httpapi/reports.go#L67) ([`internal/httpapi/reports.go`](../../internal/httpapi/reports.go)); [`server.projectReportPDF`](../../internal/httpapi/report_pdf.go#L15) ([`internal/httpapi/report_pdf.go`](../../internal/httpapi/report_pdf.go)) |
| CA-51.4 Un `{id}` de proyecto sin la forma canónica responde 400 (`errInvalidProjectID`); un UUID que no existe, 404 (`ErrProjectNotFound`), y se conserva el orden de errores | "Un ID numérico de proyecto se rechaza", "Otras formas del UUID se rechazan" (3 ejemplos), "Un UUID que no corresponde a ningún proyecto", "El UUID nulo no corresponde a ningún proyecto", "El ID de la ruta se valida antes que el cuerpo", "Con un UUID inexistente se conserva el orden de errores de HU-18", "Las demás entidades conservan su ID entero" (runner [`TestFeatures_HU51IdentificadoresUUID`](../../internal/httpapi/uuid_features_test.go#L23)) | [`TestParseProjectID_AceptaLaFormaCanonica`](../../internal/httpapi/project_id_internal_test.go#L16), [`TestParseProjectID_RechazaLoQueNoEsLaFormaCanonica`](../../internal/httpapi/project_id_internal_test.go#L42), [`TestStatusFor_IDDeProyectoInvalidoEs400`](../../internal/httpapi/project_id_internal_test.go#L70) (`internal/httpapi/project_id_internal_test.go`); [`TestRutasDeProyecto_IDNumericoDevuelve400`](../../internal/httpapi/projects_uuid_test.go#L142), [`TestRutasDeProyecto_OtrasFormasDelUUIDDevuelven400`](../../internal/httpapi/projects_uuid_test.go#L155), [`TestRutasDeProyecto_UUIDInexistenteDevuelve404`](../../internal/httpapi/projects_uuid_test.go#L168), [`TestOrdenDeErrores_IDInvalidoAntesQueElCuerpo`](../../internal/httpapi/projects_uuid_test.go#L181), [`TestOrdenDeErrores_UUIDInexistenteConservaElOrden`](../../internal/httpapi/projects_uuid_test.go#L194), [`TestRutasDeOtrasEntidades_ConservanElIDEntero`](../../internal/httpapi/projects_uuid_test.go#L212) (`internal/httpapi/projects_uuid_test.go`); [`TestGetByPublicID_InexistenteDevuelveErrProjectNotFound`](../../internal/project/public_id_test.go#L187) (`internal/project/public_id_test.go`) | [`parseProjectID`](../../internal/httpapi/server.go#L173), [`errInvalidProjectID`](../../internal/httpapi/server.go#L38), [`statusFor`](../../internal/httpapi/server.go#L321) y [`server.projectIDFromPath`](../../internal/httpapi/server.go#L192), que traduce un UUID inexistente al ID interno 0 ([`internal/httpapi/server.go`](../../internal/httpapi/server.go)) |
| CA-51.5 Funciona igual en memoria y en PostgreSQL: `projects.public_id` (UUID, único), y las relaciones siguen funcionando | — | [`TestMemoryRepository_ContratoUUID`](../../internal/project/repository_contract_test.go#L344) (`internal/project/repository_contract_test.go`); [`TestPostgresRepository_ContratoUUID`](../../internal/project/postgres_repository_test.go#L198), [`TestPostgresRepository_RechazaUUIDRepetido`](../../internal/project/postgres_repository_test.go#L206) (`internal/project/postgres_repository_test.go`) | [`MemoryRepository.GetByPublicID`](../../internal/project/memory_repository.go#L46) ([`internal/project/memory_repository.go`](../../internal/project/memory_repository.go)); [`PostgresRepository.Save`](../../internal/project/postgres_repository.go#L34), [`PostgresRepository.GetByPublicID`](../../internal/project/postgres_repository.go#L76) y [`scanProject`](../../internal/project/postgres_repository.go#L109) ([`internal/project/postgres_repository.go`](../../internal/project/postgres_repository.go)); [`db/init/01-schema.sql`](../../db/init/01-schema.sql) |
| CA-51.6 La interfaz funciona con UUID: «Proyecto actual» acepta el UUID, «Ver proyectos» y «Elegir» siguen funcionando, y el PDF lleva el UUID en el nombre | — | [`TestInterfaz_ProyectoActualAceptaElUUIDComoTexto`](../../internal/httpapi/projects_uuid_web_test.go#L25), [`TestInterfaz_NoConvierteElIDDelProyectoANumero`](../../internal/httpapi/projects_uuid_web_test.go#L41) (`internal/httpapi/projects_uuid_web_test.go`) | `#proyecto-actual` con `type="text"` ([`internal/web/index.html`](../../internal/web/index.html)); `app.js` no cambia, porque ya trataba el ID como texto ([`internal/web/app.js`](../../internal/web/app.js)) |
| CA-51.7 Crear proyectos en simultáneo genera identificadores públicos únicos, sin condiciones de carrera | — | [`TestCreate_CreacionesSimultaneasConUUIDDistintos`](../../internal/project/public_id_test.go#L111) (`internal/project/public_id_test.go`), con `-race` en el CI | [`Manager.Create`](../../internal/project/create.go#L9), que genera el UUID con `m.mu` tomado ([`internal/project/create.go`](../../internal/project/create.go)) |

Conteos (corridos sobre `1584a65`):

```
$ grep -cE '^\s*Escenario:' features/HU-51-identificadores-uuid.feature
12
$ grep -cE '^\s*Esquema del escenario:' features/HU-51-identificadores-uuid.feature
1
$ grep -c '^func Test' internal/project/public_id_test.go internal/httpapi/projects_uuid_test.go internal/httpapi/project_id_internal_test.go internal/httpapi/projects_uuid_web_test.go internal/httpapi/uuid_features_test.go
internal/project/public_id_test.go:9
internal/httpapi/projects_uuid_test.go:12
internal/httpapi/project_id_internal_test.go:4
internal/httpapi/projects_uuid_web_test.go:2
internal/httpapi/uuid_features_test.go:1
```

Los 3 tests nuevos restantes son `TestMemoryRepository_ContratoUUID`, `TestPostgresRepository_ContratoUUID` y `TestPostgresRepository_RechazaUUIDRepetido`. Los 151 tests adaptados se contaron comparando cada función `Test` de `HEAD` con la de `de928fa`, sin contar las funciones cuyo único cambio es un bloque agregado debajo de ellas.

## Fase RED (`d656a28`)

Se escribieron los tests y el `.feature` antes del código. En producción, solo firmas vacías y stubs sin lógica, para que compile y los tests fallen por aserción:

- el campo `PublicID uuid.UUID` en `Project`;
- `GetByPublicID` en `Repository`, en los dos repositorios y en el `Manager`, que devuelven `ErrProjectNotFound`;
- `errInvalidProjectID` y `parseProjectID`, que devuelve el UUID nulo sin error.

El esquema no cambió en el RED: con `public_id NOT NULL`, todos los tests de PostgreSQL habrían fallado por una causa ajena a HU-51.

Las salidas son de una copia temporal de `d656a28` con `git worktree`, corridas el 14/11/2026 en Windows con go1.27.1. Para PostgreSQL se levantó una base con el esquema del RED (sin `public_id`), con `POSTGRES_PORT=5434`.

`go test -count=1 ./...` en memoria:

```
ok   seguimiento-medicion-gpc/cmd/server 2.056s
ok   seguimiento-medicion-gpc/internal/backlog 3.205s
ok   seguimiento-medicion-gpc/internal/dashboard 4.898s
ok   seguimiento-medicion-gpc/internal/defect 2.551s
ok   seguimiento-medicion-gpc/internal/effort 1.014s
FAIL
FAIL seguimiento-medicion-gpc/internal/httpapi 2.499s
ok   seguimiento-medicion-gpc/internal/member 1.065s
ok   seguimiento-medicion-gpc/internal/metrics 3.427s
ok   seguimiento-medicion-gpc/internal/poker 1.419s
ok   seguimiento-medicion-gpc/internal/postgres 2.023s
ok   seguimiento-medicion-gpc/internal/postgrestest 1.831s
FAIL
FAIL seguimiento-medicion-gpc/internal/project 1.117s
ok   seguimiento-medicion-gpc/internal/report 2.766s
ok   seguimiento-medicion-gpc/internal/sprint 1.082s
ok   seguimiento-medicion-gpc/internal/trazabilidad 2.374s
?    seguimiento-medicion-gpc/internal/web [no test files]
FAIL
```

Con `TEST_DATABASE_URL=postgres://gpc:gpc_dev@localhost:5434/gpc_test?sslmode=disable` y `-v`, los paquetes dan lo mismo: fallan solo `internal/httpapi` e `internal/project`.

| Corrida del RED | Tests de primer nivel que fallan | SKIP | Panics |
|---|---|---|---|
| En memoria | 62 | 0 | 0 |
| Con PostgreSQL | 64 de 1179 (1115 pasan) | 0 | 0 |

Los escenarios: `15 scenarios (1 passed, 14 failed)` y `67 steps (40 passed, 14 failed, 13 skipped)`.

**Los 62 que fallan en memoria son 26 tests nuevos y 36 adaptados**, todos por aserción. El primer mensaje de cada test nuevo:

```
TestCrearProyecto_IDEsUUIDVersion4               projects_uuid_test.go:79: id = 1, se esperaba un texto
TestCrearProyecto_DosProyectosTienenUUIDDistintos projects_uuid_test.go:97: id = "1", se esperaba un UUID versión 4
TestCreate_AsignaUnUUIDVersion4                  public_id_test.go:40: Create: PublicID = "00000000-0000-0000-0000-000000000000", se esperaba un UUID versión 4 en forma canónica
TestCreate_ElPrimerProyectoNoTieneElUUIDNulo     public_id_test.go:56: PublicID = 00000000-0000-0000-0000-000000000000, no se esperaba el UUID nulo
TestCreate_GuardaElUUIDAsignado                  public_id_test.go:79: proyecto guardado: PublicID = "00000000-…", se esperaba un UUID versión 4 en forma canónica
TestCreate_MilProyectosConUUIDDistintosYNoCorrelativos public_id_test.go:98: Create(0): PublicID = "00000000-…", se esperaba un UUID versión 4 en forma canónica
TestCreate_CreacionesSimultaneasConUUIDDistintos public_id_test.go:133: Proyecto 49: PublicID = "00000000-…", se esperaba un UUID versión 4 en forma canónica
TestGetByPublicID_DevuelveElProyecto             public_id_test.go:155: GetByPublicID() devolvió un error inesperado: no existe un proyecto con ese id
TestGetByPublicID_DevuelveUnaCopia               public_id_test.go:171: GetByPublicID() devolvió un error inesperado: no existe un proyecto con ese id
TestModificar_ConservaElUUID                     public_id_test.go:241: Create: PublicID = "00000000-…", se esperaba un UUID versión 4 en forma canónica
TestMemoryRepository_ContratoUUID                repository_contract_test.go:307: GetByPublicID(3f2b8c1d-…) devolvió un error inesperado: no existe un proyecto con ese id
TestParseProjectID_AceptaLaFormaCanonica         project_id_internal_test.go:23: parseProjectID("3f2b8c1d-…") = 00000000-…, se esperaba 3f2b8c1d-…
TestParseProjectID_AceptaMayusculas              project_id_internal_test.go:37: parseProjectID("3F2B8C1D-…").String() = "00000000-…", se esperaba "3f2b8c1d-…"
TestParseProjectID_RechazaLoQueNoEsLaFormaCanonica project_id_internal_test.go:63: parseProjectID(""): err = <nil>, se esperaba el id del proyecto debe ser un UUID
TestStatusFor_IDDeProyectoInvalidoEs400          project_id_internal_test.go:72: statusFor(errInvalidProjectID) = 500, se esperaba 400
TestRutasDeProyecto_AceptanElUUID                projects_uuid_test.go:108: ruta del proyecto = "/api/proyectos/1", se esperaba un UUID versión 4
TestRutasDeProyecto_AceptanElUUIDEnMayusculas    projects_uuid_test.go:124: id = "1", se esperaba un UUID versión 4
TestRutasDeProyecto_IDNumericoDevuelve400        projects_uuid_test.go:149: GET /api/proyectos/1: código = 200, se esperaba 400
TestRutasDeProyecto_OtrasFormasDelUUIDDevuelven400 projects_uuid_test.go:162: GET /api/proyectos/{1}: error = "el id debe ser un número entero", se esperaba "el id del proyecto debe ser un UUID"
TestRutasDeProyecto_UUIDInexistenteDevuelve404   projects_uuid_test.go:175: GET /api/proyectos/6f1c2a4e-…: código = 400, se esperaba 404
TestOrdenDeErrores_IDInvalidoAntesQueElCuerpo    projects_uuid_test.go:189: PUT /api/proyectos/abc: error = "el id debe ser un número entero", se esperaba "el id del proyecto debe ser un UUID"
TestOrdenDeErrores_UUIDInexistenteConservaElOrden projects_uuid_test.go:207: PATCH /api/proyectos/6f1c2a4e-…/fecha-inicio: error = "el id debe ser un número entero", se esperaba "la fecha de inicio del proyecto es obligatoria"
TestRespuestas_IdentificanElProyectoConSuUUID    projects_uuid_test.go:281: POST /api/proyectos: id = 1, se esperaba un UUID versión 4 como texto
TestReportePDF_NombreDelArchivoConElUUID         projects_uuid_test.go:343: id = "1", se esperaba un UUID versión 4
TestInterfaz_ProyectoActualAceptaElUUIDComoTexto projects_uuid_web_test.go:30: «Proyecto actual» = <input id="proyecto-actual" type="number" min="1" placeholder="Ej.: 1">, no tiene type="text"
TestFeatures_HU51IdentificadoresUUID             14 de 15 escenarios en rojo, por ejemplo: id = 1, se esperaba un texto; código = 200, se esperaba 400
```

(Los UUID largos se acortaron con "…".)

Con PostgreSQL se suman los 2 tests nuevos de la base:

```
TestPostgresRepository_ContratoUUID      repository_contract_test.go:307: GetByPublicID(3f2b8c1d-…) devolvió un error inesperado: no existe un proyecto con ese id
TestPostgresRepository_RechazaUUIDRepetido postgres_repository_test.go:214: Save con un PublicID repetido no devolvió error
```

**Los 36 adaptados que fallan** son los que cambian su resultado esperado con HU-51. Fallan porque en el RED una ruta con un UUID responde 400 ("el id debe ser un número entero"), y un `{id}` que no es un UUID todavía no da el mensaje nuevo:

- 19 de "proyecto inexistente → 404", que ahora usan un UUID que nunca se asignó o el UUID nulo (por ejemplo, `TestMetricasProyecto_InexistenteDevuelve404` y `TestListarIntegrantes_ProyectoInexistenteDevuelve404`);
- 9 de "{id} inválido → 400", que ahora esperan "el id del proyecto debe ser un UUID" (por ejemplo, `TestRutasConID_IDNoNumericoDevuelve400` y `TestReportePDF_IDNoNumericoDevuelve400`);
- 4 de orden de errores con un proyecto inexistente (por ejemplo, `TestRegistrarFechaInicioProyecto_FechaVaciaEnProyectoInexistenteDevuelve400` y `TestListarDefectos_EstadoInvalidoDevuelve400`);
- 4 de error interno con el repositorio de proyectos roto, que ahora usan un UUID (`TestRegistrarFechaInicioProyecto_ErrorInternoDevuelve500`, `TestRegistrarFechaFinProyecto_ErrorInternoDevuelve500`, `TestModificarFechasProyecto_ErrorInternoDevuelve500` y `TestListarSprintsDelProyecto_FallaDeUnRepositorioDevuelve500`).

**Pasan en el RED 115 de los adaptados y 3 de los nuevos.** Los nuevos protegen comportamiento que no cambia:

- `TestGetByPublicID_InexistenteDevuelveErrProjectNotFound`, porque el stub ya devuelve `ErrProjectNotFound`;
- `TestRutasDeOtrasEntidades_ConservanElIDEntero`, porque los ítems y los sprints siguen con ID entero;
- `TestInterfaz_NoConvierteElIDDelProyectoANumero`, porque `app.js` ya trataba el ID como texto.

Lo mismo pasa con el escenario "Las demás entidades conservan su ID entero".

## Commit de tests entre el RED y el GREEN (`20f2bf1`)

En el RED quedaron 4 `"id"` de proyecto escritos a mano en JSON esperados: `projects_list_test.go` (2) y `reports_test.go` (2). En el RED pasaban, porque el ID todavía era 1 y 2. Con el GREEN fallaban 3 tests, que el GREEN no puede tocar:

- `TestListarProyectos_DevuelveLosProyectosConElFormatoDeLaConsultaPorID`;
- `TestReporteProyecto_ContratoDelEjemploDeTEC03`;
- `TestReporteProyecto_ContratoDelProyectoVacio`.

Con la autorización de Sosa Ricardo, el GREEN se guardó aparte con `git stash` y los 4 literales pasaron a `jsonProjectID(t, h, n)`, como en el resto de los tests adaptados. En el estado del RED, los 3 tests siguieron pasando, fallaron los mismos 62 tests y `internal/trazabilidad` dio `ok`. Después se recuperó el GREEN.

## Fase GREEN (`1584a65`)

Código, en el orden de `tasks.md`:

- **Dominio (3.1):**
  - `Manager.Create` asigna `uuid.NewV4()` con el mutex tomado;
  - `Manager.GetByPublicID` lee con el mutex;
  - `MemoryRepository.GetByPublicID` busca en el slice.
- **PostgreSQL (3.2):**
  - `public_id UUID NOT NULL UNIQUE` en `projects`;
  - `Save` lo inserta, `scanProject` lo lee y `GetByPublicID` busca con `WHERE public_id = $1`;
  - pgx lee `uuid.UUID` sin conversión, así que no hizo falta la mitigación de R3.
- **Entrada de la API (3.3):**
  - `parseProjectID` exige 36 caracteres con guiones en las posiciones 8, 13, 18 y 23, y después `uuid.Parse`;
  - `errInvalidProjectID` responde 400 en `statusFor`;
  - `projectIDFromPath` devuelve el ID interno y el UUID en minúsculas, y un UUID inexistente da el ID 0 (D4);
  - las 17 rutas de proyecto lo usan.
- **Salida de la API (3.4):**
  - `id` y `proyecto_id` son texto con el UUID;
  - las rutas por ID de otra entidad (ítems, sprints, integrantes y el reporte del sprint) lo buscan con `publicProjectID`, que da 500 si no encuentra el proyecto;
  - el PDF se llama `reporte-proyecto-{UUID}.pdf`.
- **Interfaz (3.5):** `#proyecto-actual` con `type="text"`, `placeholder="UUID del proyecto"`, `autocomplete="off"` y `spellcheck="false"`, y el comentario de la sección cita HU-51. `app.js` no cambia.

Los comandos de verificación del proyecto 0 de D4 (comandos 1 a 3 de `design.md`) siguen sin resultados sobre `1584a65`.

El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat -- '*_test.go' features/     (antes del commit 1584a65)
(sin salida)
```

Sobre `1584a65`, el 14/11/2026, con la base recreada con el esquema nuevo (`docker compose down -v` y `up -d --wait`, PostgreSQL 17.11, puerto 5434):

| Corrida del GREEN | Resultado |
|---|---|
| `go test -count=1 ./...` en memoria | 15 paquetes `ok`, 0 FAIL |
| `go test -count=1 -v ./...` con PostgreSQL | 15 paquetes `ok`, 1179 tests de primer nivel, los 1179 en PASS, 0 FAIL y 0 SKIP |
| Escenarios de HU-51 | `15 scenarios (15 passed)` y `67 steps (67 passed)` |

```
ok   seguimiento-medicion-gpc/cmd/server 1.369s
ok   seguimiento-medicion-gpc/internal/backlog 3.662s
ok   seguimiento-medicion-gpc/internal/dashboard 2.220s
ok   seguimiento-medicion-gpc/internal/defect 1.990s
ok   seguimiento-medicion-gpc/internal/effort 1.420s
ok   seguimiento-medicion-gpc/internal/httpapi 2.620s
ok   seguimiento-medicion-gpc/internal/member 1.141s
ok   seguimiento-medicion-gpc/internal/metrics 2.733s
ok   seguimiento-medicion-gpc/internal/poker 1.525s
ok   seguimiento-medicion-gpc/internal/postgres 2.978s
ok   seguimiento-medicion-gpc/internal/postgrestest 2.739s
ok   seguimiento-medicion-gpc/internal/project 1.373s
ok   seguimiento-medicion-gpc/internal/report 2.921s
ok   seguimiento-medicion-gpc/internal/sprint 0.685s
ok   seguimiento-medicion-gpc/internal/trazabilidad 0.961s
?    seguimiento-medicion-gpc/internal/web [no test files]
```

### Prueba con curl

Con `PORT=8080 go run ./cmd/server` (datos en memoria), sobre `1584a65`:

```
$ curl -s -w "\n%{http_code}\n" -X POST localhost:8080/api/proyectos -d '{"nombre":"Evidencia HU-51","descripcion":"curl"}'
{"id":"ea55e487-23e7-4811-867b-432800ee5d9c","nombre":"Evidencia HU-51","descripcion":"curl","estado":"PLANIFICADO","fecha_inicio":"","fecha_fin":""}
201
$ curl -s -w "\n%{http_code}\n" localhost:8080/api/proyectos/ea55e487-23e7-4811-867b-432800ee5d9c
{"id":"ea55e487-23e7-4811-867b-432800ee5d9c","nombre":"Evidencia HU-51","descripcion":"curl","estado":"PLANIFICADO","fecha_inicio":"","fecha_fin":""}
200
$ curl -s -w "\n%{http_code}\n" localhost:8080/api/proyectos/EA55E487-23E7-4811-867B-432800EE5D9C
{"id":"ea55e487-23e7-4811-867b-432800ee5d9c","nombre":"Evidencia HU-51","descripcion":"curl","estado":"PLANIFICADO","fecha_inicio":"","fecha_fin":""}
200
$ curl -s -w "\n%{http_code}\n" localhost:8080/api/proyectos/1
{"error":"el id del proyecto debe ser un UUID"}
400
$ curl -s -w "\n%{http_code}\n" localhost:8080/api/proyectos/6f1c2a4e-8b3d-4c7a-9e2f-1a2b3c4d5e6f
{"error":"no existe un proyecto con ese id"}
404
$ curl -s -w "\n%{http_code}\n" localhost:8080/api/items/abc
{"error":"el id debe ser un número entero"}
400
$ curl -s -D - -o /dev/null localhost:8080/api/proyectos/ea55e487-23e7-4811-867b-432800ee5d9c/reporte/pdf | grep -i content-disposition
Content-Disposition: attachment; filename="reporte-proyecto-ea55e487-23e7-4811-867b-432800ee5d9c.pdf"
```

### Uso desde el navegador (tarea 3.5)

Sosa Ricardo lo probó el 14/11/2026 en el navegador, con el servidor de la rama (`PORT=8080 go run ./cmd/server`, datos en memoria, sobre `1584a65`):

| Acción | Resultado |
|---|---|
| Mirar el campo «Proyecto actual» | Es un campo de texto, ya no numérico (CA-51.6) |
| Escribir "1" y tocar «Cargar proyecto» | Aparece "el id del proyecto debe ser un UUID" y no se carga ningún proyecto (CA-51.4 y CA-51.6) |
| «Ver proyectos» y «Elegir» en una fila | «Proyecto actual» queda con el UUID del proyecto (CA-51.6, HU-50) |
| Crear un ítem del backlog en ese proyecto | El ítem se crea en el proyecto elegido (CA-51.3) |

## Refactor

No hubo (tarea 4.1). `publicProjectID` se usa en cinco handlers, pero ya está en un solo lugar (`server.go`), y las listas usan `writeItems`. Otra extracción, por ejemplo un helper que devuelva la respuesta del ítem con el UUID, movería código sin ganar legibilidad.

## Cobertura

Con PostgreSQL 17.11 (puerto 5434), en Windows con go1.27.1, el 14/11/2026: `go test -count=1 -coverprofile=cover.out ./...` y `go tool cover -func=cover.out`, más el total con `-coverpkg=./...`. `main` (`8bf9550`) se midió con los mismos comandos, en una copia temporal con `git worktree` y con su esquema.

| Paquete | `main` `8bf9550` | Rama `1584a65` |
|---|---|---|
| `internal/httpapi` | 98.8% | 97.4% |
| `internal/project` | 98.2% | 98.4% |
| Los otros 12 paquetes con código | iguales | iguales |
| Total del perfil | 95.3% | 95.0% |
| Total con `-coverpkg=./...` | 96.3% | 95.9% |

Las funciones nuevas o cambiadas:

- quedan en 100.0% `parseProjectID`, `projectIDFromPath`, `Manager.GetByPublicID`, `MemoryRepository.GetByPublicID`, `Manager.Create`, `toItemResponse` y `toSprintResponse`;
- quedan por debajo `publicProjectID` (75.0%), `writeItem` (60.0%), `writeItems` (77.8%), `closeSprint` (84.6%), `sprintReport` (84.6%), `updateMember` (88.2%) y `PostgresRepository.GetByPublicID` (85.7%).

**Brecha:** las líneas sin cubrir son el 500 cuando no se puede leer el proyecto de un ítem, de un sprint o de un integrante (`publicProjectID`), y la rama de error genérico de la base en `PostgresRepository.GetByPublicID`. Ningún test las ejercita: hace falta un repositorio de proyectos que falle solo en `Get`. Por eso `internal/httpapi` baja 1.4 puntos. Queda como trabajo pendiente, en su propio ciclo RED y GREEN. El detalle está en [`docs/reportes/cobertura.md`](../reportes/cobertura.md).

## Desviaciones del ciclo TDD

1. **Dos commits repetidos de aprobación y merge.** `bf0adaf` (aprobación) y `ce8caba` (merge de `main`) salieron de un primer pedido que se interrumpió. El segundo pedido, que lo reemplazó, agregó `93bd16d` (merge de `main` con el control de TEC-11) y `64a59db` (aprobación con los arreglos de Hernán). No se reescribió la historia: los cuatro commits quedan, y la aprobación vigente es `64a59db`.
2. **Un commit de tests entre el RED y el GREEN** (`20f2bf1`), explicado arriba y anotado en `tasks.md`.
3. **Tests que pasan en el RED:** 3 nuevos y un escenario que protegen comportamiento sin cambios, explicados en la fase RED.
4. **`-race` lo corre el CI:** en Windows no anda (`-race requires cgo`), así que `TestCreate_CreacionesSimultaneasConUUIDDistintos` (CA-51.7) corrió acá sin `-race`. El CI lo vuelve a correr con `-race` en el Pull Request.
5. **El esquema no cambió en el RED,** como se explica en la fase RED. Con el esquema y `Create` del GREEN se probó, antes del commit del RED, que ningún test ajeno a HU-51 fallaba. Después se revirtió.
6. **Los tests de `parseProjectID` van en un archivo interno del paquete** (`project_id_internal_test.go`, `package httpapi`), y no en `projects_uuid_test.go` como decía D8, porque la función no se exporta.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| Propuesta (OpenSpec) | 2229eeb | `docs(HU-51): agregar la propuesta de OpenSpec` |
| Corrección de la propuesta | c929ba3 | `docs(HU-51): corregir el estado de aprobación de la propuesta` |
| Corrección de la propuesta | 8cbb671 | `docs(HU-51): corregir el error de un ID de proyecto inválido en la propuesta` |
| Aprobación (primer pedido, desviación 1) | bf0adaf | `docs(HU-51): registrar la aprobación de la propuesta` |
| Merge de `main` (primer pedido, desviación 1) | ce8caba | `chore(HU-51): traer main con el archivo de HU-50` |
| Merge de `main` | 93bd16d | `chore(HU-51): traer main con el archivo de HU-50 y el control de TEC-11` |
| Aprobación con los arreglos | 64a59db | `docs(HU-51): registrar la aprobación de la propuesta` |
| RED | d656a28 | `test(HU-51): agregar tests (RED)` |
| Tests (desviación 2) | 20f2bf1 | `test(HU-51): usar el id del proyecto en los JSON esperados de la lista y del reporte` |
| GREEN | 1584a65 | `feat(HU-51): identificar los proyectos con un UUID público (GREEN)` |
| Refactor | — | No hubo (ver "Refactor") |
| Evidencia y cobertura | b4413de | `docs(HU-51): agregar evidencia TDD y cobertura` |
| Merge de `main` (control de links de TEC-11, PR #194) | b26d781 | `chore(HU-51): traer main con el control de links de TEC-11` |
| Links de las evidencias (`-update`) | a6db665 | `docs(HU-51): actualizar los links de las evidencias a las líneas de HU-51` |
| Cierre | este commit | `docs(HU-51): cerrar historia` |

El commit de cierre figura como "este commit", porque un commit no puede contener su propio hash: queda registrado en el Pull Request #195 y en el Issue #193. Los demás hashes se completaron a medida que se hizo cada commit.

HU-51: Historia → SDD (OpenSpec) → Criterios de Aceptación → BDD → Tests → Código Go
