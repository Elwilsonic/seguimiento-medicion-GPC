# Evidencia TDD · HU-52 Identificadores UUID del resto de las entidades

> Issue del repositorio: #197 · Épica E1, con alcance en E2 a E8 · Vinculada a HU-51 (#193) y a las historias de cada entidad (HU-04 a HU-17 y HU-20 a HU-31) · Especificada con OpenSpec (TEC-10, #177)

- **Historia:** como integrante del equipo, quiero que los ítems del backlog, los sprints, los integrantes, los defectos, los registros de esfuerzo y las sesiones de Planning Poker se identifiquen con un UUID en lugar de un número correlativo, para que no se pueda adivinar ni recorrer ninguna entidad, ni deducir cuántas hay.
- **Especificación (OpenSpec):** el cambio [`openspec/changes/hu-52-identificadores-uuid-entidades/`](../../openspec/changes/hu-52-identificadores-uuid-entidades/): [propuesta](../../openspec/changes/hu-52-identificadores-uuid-entidades/proposal.md), [diseño](../../openspec/changes/hu-52-identificadores-uuid-entidades/design.md), [tareas](../../openspec/changes/hu-52-identificadores-uuid-entidades/tasks.md) y una spec delta por entidad (`specs/backlog`, `sprints`, `integrantes`, `defectos`, `esfuerzos` y `planning-poker`).
- **Aprobación:** Peñalbé Hernán (Agile Enabler) aprobó la propuesta el 15/11/2026 (`8b65300`), con el dominio de las 6 entidades en la base (D3), el centinela `missingInternalID = math.MaxInt32` (D4), las 6 entidades con el corte por entidad, el corte del lunes 16/11/2026 a las 14:00 y los Story Points en Planning Poker. Cada responsable autoriza sus propios commits.
- **Organización:** una base compartida (Sosa Ricardo y Cardozo Leandro en pareja, sujeto a que Cardozo Leandro lo confirme), tres partes por entidad y el PDF del reporte (Peñalbé Hernán). Esta evidencia tiene una sección por parte.
- **Rama de la base:** `feature/HU-52-uuid-base`, con `main` traído en `e039429` (`45d2c31`, con la columna Código de las evidencias, PR #199).
- **Rama de integración (cambio en D3, aprobado por el Agile Enabler el 15/11/2026):** `feature/HU-52-uuid-entidades`, creada por Buttini Cristobal desde `main` (`29051da`). Las partes por entidad no se pueden integrar solas en `main` con la CI en verde: al medir la de sprints y Planning Poker antes del RED fallaron 123 tests, 81 por "el id del ítem debe ser un UUID" (sus rutas reciben ítems) y 37 por "el id debe ser un número entero" (las rutas de defectos y esfuerzos reciben sprints). Por eso todos los PR de HU-52 van a esa rama, sin CI y con la salida de `go test ./...` (en memoria y con PostgreSQL) pegada en cada PR, que dice qué fallos quedan y de qué parte dependen. Un solo PR a `main`, con la CI en verde (`-race`) y 0 fallos, cierra HU-52 con `Closes #197`. Horario: las tres partes en la rama de integración a las 12:00 del 16/11/2026 y el PR a `main` integrado antes de las 14:00. El detalle está en D3 de [`design.md`](../../openspec/changes/hu-52-identificadores-uuid-entidades/design.md).

## Trazabilidad

### Base compartida

**68 tests nuevos (funciones `Test`) y 36 ejecuciones de escenarios BDD.** Se reparten así:

- 36 de dominio, 6 por paquete, en `public_id_test.go` de `backlog`, `sprint`, `member`, `defect`, `effort` y `poker`;
- 18 de repositorio, 3 por paquete, en `public_id_repository_test.go`: el contrato `GetByPublicID` en memoria y en PostgreSQL, y el UUID repetido en PostgreSQL;
- 13 internos de `internal/httpapi` en `public_ids_internal_test.go`, porque los helpers no se exportan;
- el runner de godog de `features/HU-52-base.feature`, en `internal/httpapi/uuid_base_features_test.go`, que ejecuta los 36 escenarios (6 esquemas × 6 entidades).

Las cuentas de tests de la tabla no suman 68 porque un test puede cubrir más de un criterio, y la correspondencia con los escenarios no es 1 a 1: los 1000 identificadores (CA-52.2), PostgreSQL (CA-52.5) y la concurrencia (CA-52.7) se cubren solo con tests de Go.

La base no cambia rutas ni salidas de la API: los helpers y el resolvedor quedan listos, y cada entidad los usa en su parte. Por eso CA-52.3 y CA-52.4 se cumplen recién con las partes integradas, y CA-52.6 (la interfaz) es solo de las partes de cada entidad. **El PDF del reporte quedó fuera de la base por decisión del Agile Enabler** (15/11/2026): lo toma Peñalbé Hernán, en paralelo con las entidades (D3 y sección 5 de `tasks.md`). El PDF lo hizo Peñalbé Hernán en el PR #200, integrado en esta rama (ver su sección).

#### Criterio → test → código

| Criterio | Escenarios BDD | Tests de Go (archivo) | Código que lo implementa (función y archivo) |
|---|---|---|---|
| CA-52.1 Al crearse, cada entidad recibe un identificador público UUID v4 en forma canónica | "Crear una entidad le asigna un UUID versión 4" (6 ejemplos; runner [`TestFeatures_HU52Base`](../../internal/httpapi/uuid_base_features_test.go#L31), `internal/httpapi/uuid_base_features_test.go`) | [`TestCreate_AsignaUnUUIDVersion4AlItem`](../../internal/backlog/public_id_test.go#L35) (`internal/backlog/public_id_test.go`); [`TestCreate_AsignaUnUUIDVersion4AlSprint`](../../internal/sprint/public_id_test.go#L28) (`internal/sprint/public_id_test.go`); [`TestRegister_AsignaUnUUIDVersion4AlIntegrante`](../../internal/member/public_id_test.go#L54) (`internal/member/public_id_test.go`); [`TestRegister_AsignaUnUUIDVersion4AlDefecto`](../../internal/defect/public_id_test.go#L61) (`internal/defect/public_id_test.go`); [`TestRegister_AsignaUnUUIDVersion4AlEsfuerzo`](../../internal/effort/public_id_test.go#L58) (`internal/effort/public_id_test.go`); [`TestOpenSession_AsignaUnUUIDVersion4ALaSesion`](../../internal/poker/public_id_test.go#L63) (`internal/poker/public_id_test.go`) | `uuid.NewV4()` en [`Manager.Create`](../../internal/backlog/create.go#L14) ([`internal/backlog/create.go`](../../internal/backlog/create.go)), [`Manager.Create`](../../internal/sprint/create.go#L16) ([`internal/sprint/create.go`](../../internal/sprint/create.go)), [`Manager.Register`](../../internal/member/register.go#L12) ([`internal/member/register.go`](../../internal/member/register.go)), [`Manager.Register`](../../internal/defect/register.go#L14) ([`internal/defect/register.go`](../../internal/defect/register.go)), [`Manager.Register`](../../internal/effort/register.go#L17) ([`internal/effort/register.go`](../../internal/effort/register.go)) y [`Manager.OpenSession`](../../internal/poker/create.go#L15) ([`internal/poker/create.go`](../../internal/poker/create.go)); el campo `PublicID` de [`Item`](../../internal/backlog/item.go#L30), [`Sprint`](../../internal/sprint/sprint.go#L21), [`Member`](../../internal/member/member.go#L20), [`Defect`](../../internal/defect/defect.go#L51), [`Effort`](../../internal/effort/effort.go#L11) y [`Session`](../../internal/poker/session.go#L36) |
| CA-52.2 Dos entidades del mismo tipo nunca tienen el mismo identificador público, y no se puede deducir otro ni la cantidad | "Varias creaciones dan identificadores distintos" (6 ejemplos; runner [`TestFeatures_HU52Base`](../../internal/httpapi/uuid_base_features_test.go#L31)) | [`TestCreate_MilItemsConUUIDDistintosYNoCorrelativos`](../../internal/backlog/public_id_test.go#L53); [`TestCreate_MilSprintsConUUIDDistintosYNoCorrelativos`](../../internal/sprint/public_id_test.go#L47); [`TestRegister_MilIntegrantesConUUIDDistintosYNoCorrelativos`](../../internal/member/public_id_test.go#L71); [`TestRegister_MilDefectosConUUIDDistintosYNoCorrelativos`](../../internal/defect/public_id_test.go#L78); [`TestRegister_MilEsfuerzosConUUIDDistintosYNoCorrelativos`](../../internal/effort/public_id_test.go#L75); [`TestOpenSession_MilSesionesConUUIDDistintosYNoCorrelativos`](../../internal/poker/public_id_test.go#L80) (en `public_id_test.go` de cada paquete) | Las mismas 6 funciones de creación de CA-52.1; `public_id UUID NOT NULL UNIQUE` en las 6 tablas ([`db/init/01-schema.sql`](../../db/init/01-schema.sql)) |
| CA-52.3 Las rutas aceptan el UUID y ninguna respuesta expone el entero (en la base: la búsqueda por UUID, la traducción al ID interno y el resolvedor de las respuestas) | "Buscar una entidad por su UUID" (6 ejemplos; runner [`TestFeatures_HU52Base`](../../internal/httpapi/uuid_base_features_test.go#L31)) | [`TestGetByPublicID_DevuelveElItemYUnaCopia`](../../internal/backlog/public_id_test.go#L110), [`TestGetByPublicID_DevuelveElSprint`](../../internal/sprint/public_id_test.go#L112), [`TestGetByPublicID_DevuelveElIntegrante`](../../internal/member/public_id_test.go#L130), [`TestGetByPublicID_DevuelveElDefecto`](../../internal/defect/public_id_test.go#L137), [`TestGetByPublicID_DevuelveElEsfuerzo`](../../internal/effort/public_id_test.go#L134), [`TestGetByPublicID_DevuelveLaSesion`](../../internal/poker/public_id_test.go#L140) (en `public_id_test.go` de cada paquete); [`TestParsePublicID_AceptaLaFormaCanonicaYMayusculas`](../../internal/httpapi/public_ids_internal_test.go#L53), [`TestIDFromPath_UUIDExistenteDaElIDInterno`](../../internal/httpapi/public_ids_internal_test.go#L192), [`TestSprintIDFromBody_TraduceElUUIDDelCuerpo`](../../internal/httpapi/public_ids_internal_test.go#L242), [`TestPublicIDs_DevuelveLosUUIDYVacioParaElCero`](../../internal/httpapi/public_ids_internal_test.go#L282), [`TestPublicIDs_BuscaCadaIDUnaSolaVez`](../../internal/httpapi/public_ids_internal_test.go#L304), [`TestWriteResolved_ErrorDelResolvedorEs500`](../../internal/httpapi/public_ids_internal_test.go#L321), [`TestWriteResolved_SinErrorRespondeLaRespuesta`](../../internal/httpapi/public_ids_internal_test.go#L340) (`internal/httpapi/public_ids_internal_test.go`) | `Manager.GetByPublicID` de cada paquete: [`backlog`](../../internal/backlog/public_id.go), [`sprint`](../../internal/sprint/public_id.go), [`member`](../../internal/member/public_id.go), [`defect`](../../internal/defect/public_id.go), [`effort`](../../internal/effort/public_id.go) y [`poker`](../../internal/poker/public_id.go) (`public_id.go`); [`parsePublicID`](../../internal/httpapi/server.go#L225), [`internalID`](../../internal/httpapi/server.go#L239), [`server.itemIDFromPath`](../../internal/httpapi/server.go#L256), [`server.sprintIDFromPath`](../../internal/httpapi/server.go#L262), [`server.memberIDFromPath`](../../internal/httpapi/server.go#L268), [`server.defectIDFromPath`](../../internal/httpapi/server.go#L274), [`server.effortIDFromPath`](../../internal/httpapi/server.go#L280), [`server.sessionIDFromPath`](../../internal/httpapi/server.go#L286), [`server.sprintIDFromBody`](../../internal/httpapi/server.go#L293), [`server.newPublicIDs`](../../internal/httpapi/server.go#L311), [`publicIDs.item`](../../internal/httpapi/server.go#L317), [`publicIDs.sprint`](../../internal/httpapi/server.go#L326), [`resolve`](../../internal/httpapi/server.go#L336) y [`server.writeResolved`](../../internal/httpapi/server.go#L356) ([`internal/httpapi/server.go`](../../internal/httpapi/server.go)); el parámetro `ids` de las funciones de respuesta y los 19 handlers que responden con `writeResolved` (`items.go`, `sprints.go`, `efforts.go`, `defects.go`, `metrics.go`, `reports.go`, `sprint_reports.go` y `dashboard.go`), sin cambiar todavía su salida |
| CA-52.4 Un ID que no es un UUID responde 400 con el mensaje de su entidad; un UUID que no existe, 404, conservando el orden de errores (en la base: los errores, su 400 y el centinela) | "El UUID nulo no corresponde a ninguna entidad" y "Un UUID que nunca se asignó da el error de su entidad" (6 ejemplos cada uno; runner [`TestFeatures_HU52Base`](../../internal/httpapi/uuid_base_features_test.go#L31)) | [`TestGetByPublicID_ItemInexistenteDevuelveErrItemNotFound`](../../internal/backlog/public_id_test.go#L139), [`TestGetByPublicID_SprintInexistenteDevuelveErrSprintNotFound`](../../internal/sprint/public_id_test.go#L135), [`TestGetByPublicID_ElIntegranteInexistenteDevuelveErrMemberNotFound`](../../internal/member/public_id_test.go#L153), [`TestGetByPublicID_ElDefectoInexistenteDevuelveErrDefectNotFound`](../../internal/defect/public_id_test.go#L160), [`TestGetByPublicID_ElEsfuerzoInexistenteDevuelveErrEffortNotFound`](../../internal/effort/public_id_test.go#L157), [`TestGetByPublicID_LaSesionInexistenteDevuelveErrSessionNotFound`](../../internal/poker/public_id_test.go#L163) (en `public_id_test.go` de cada paquete); [`TestStatusFor_IDsDeEntidadInvalidosSon400`](../../internal/httpapi/public_ids_internal_test.go#L41), [`TestParsePublicID_RechazaConElErrorDeLaEntidad`](../../internal/httpapi/public_ids_internal_test.go#L67), [`TestMissingInternalID_EsElMaximoDeInteger`](../../internal/httpapi/public_ids_internal_test.go#L87), [`TestIDFromPath_UUIDInexistenteDaMissingInternalID`](../../internal/httpapi/public_ids_internal_test.go#L210), [`TestIDFromPath_FormatoInvalidoDaElErrorDeLaEntidad`](../../internal/httpapi/public_ids_internal_test.go#L225), [`TestSprintIDFromBody_FormatoInvalidoDaErrInvalidSprintID`](../../internal/httpapi/public_ids_internal_test.go#L267) (`internal/httpapi/public_ids_internal_test.go`) | [`errInvalidItemID`](../../internal/httpapi/server.go#L41) y los otros 5 errores de formato, [`missingInternalID`](../../internal/httpapi/server.go#L220), [`internalID`](../../internal/httpapi/server.go#L239) y [`statusFor`](../../internal/httpapi/server.go#L375) ([`internal/httpapi/server.go`](../../internal/httpapi/server.go)); `GetByPublicID` de los repositorios, que dan el error de no encontrado de cada entidad |
| CA-52.5 Funciona igual en memoria y en PostgreSQL: las 6 tablas suman `public_id` (UUID, único), y las modificaciones lo conservan | "Modificar una entidad conserva su UUID" (6 ejemplos; runner [`TestFeatures_HU52Base`](../../internal/httpapi/uuid_base_features_test.go#L31)) | [`TestModificaciones_ConservanElUUIDDelItem`](../../internal/backlog/public_id_test.go#L157), [`TestClose_ConservaElUUIDDelSprint`](../../internal/sprint/public_id_test.go#L154), [`TestUpdate_ConservaElUUIDDelIntegrante`](../../internal/member/public_id_test.go#L171), [`TestChangeStatus_ConservaElUUIDDelDefecto`](../../internal/defect/public_id_test.go#L178), [`TestCorrectHours_ConservaElUUIDDelEsfuerzo`](../../internal/effort/public_id_test.go#L175), [`TestOperaciones_ConservanElUUIDDeLaSesion`](../../internal/poker/public_id_test.go#L181) (en `public_id_test.go` de cada paquete); en `public_id_repository_test.go` de cada paquete, backlog: [`TestMemoryRepository_ContratoPublicID`](../../internal/backlog/public_id_repository_test.go#L108), [`TestPostgresRepository_ContratoPublicID`](../../internal/backlog/public_id_repository_test.go#L113) y [`TestPostgresRepository_PublicIDRepetidoFalla`](../../internal/backlog/public_id_repository_test.go#L121); sprint: [`TestMemoryRepository_ContratoPublicID`](../../internal/sprint/public_id_repository_test.go#L108), [`TestPostgresRepository_ContratoPublicID`](../../internal/sprint/public_id_repository_test.go#L113) y [`TestPostgresRepository_PublicIDRepetidoFalla`](../../internal/sprint/public_id_repository_test.go#L121); member: [`TestMemoryRepository_ContratoPublicID`](../../internal/member/public_id_repository_test.go#L108), [`TestPostgresRepository_ContratoPublicID`](../../internal/member/public_id_repository_test.go#L113) y [`TestPostgresRepository_PublicIDRepetidoFalla`](../../internal/member/public_id_repository_test.go#L121); defect: [`TestMemoryRepository_ContratoPublicID`](../../internal/defect/public_id_repository_test.go#L108), [`TestPostgresRepository_ContratoPublicID`](../../internal/defect/public_id_repository_test.go#L113) y [`TestPostgresRepository_PublicIDRepetidoFalla`](../../internal/defect/public_id_repository_test.go#L121); effort: [`TestMemoryRepository_ContratoPublicID`](../../internal/effort/public_id_repository_test.go#L108), [`TestPostgresRepository_ContratoPublicID`](../../internal/effort/public_id_repository_test.go#L113) y [`TestPostgresRepository_PublicIDRepetidoFalla`](../../internal/effort/public_id_repository_test.go#L121); poker: [`TestMemoryRepository_ContratoPublicID`](../../internal/poker/public_id_repository_test.go#L108), [`TestPostgresRepository_ContratoPublicID`](../../internal/poker/public_id_repository_test.go#L113) y [`TestPostgresRepository_PublicIDRepetidoFalla`](../../internal/poker/public_id_repository_test.go#L121) | `GetByPublicID` de los repositorios en memoria y en PostgreSQL, `public_id` en `Save` y en el `scan` de PostgreSQL de los 6 paquetes; por ejemplo, [`MemoryRepository.GetByPublicID`](../../internal/backlog/memory_repository.go#L84) ([`internal/backlog/memory_repository.go`](../../internal/backlog/memory_repository.go)), [`PostgresRepository.Save`](../../internal/backlog/postgres_repository.go#L33), [`scanItem`](../../internal/backlog/postgres_repository.go#L109) y [`PostgresRepository.GetByPublicID`](../../internal/backlog/postgres_repository.go#L118) ([`internal/backlog/postgres_repository.go`](../../internal/backlog/postgres_repository.go)); [`Manager.Update`](../../internal/backlog/update.go#L23), que conserva el `PublicID` ([`internal/backlog/update.go`](../../internal/backlog/update.go)); [`db/init/01-schema.sql`](../../db/init/01-schema.sql) |
| CA-52.6 La interfaz web funciona con UUID | — | — (lo hace cada entidad) | — |
| CA-52.7 Crear entidades en simultáneo genera identificadores únicos, sin condiciones de carrera | — | [`TestCreate_CreacionesSimultaneasDeItemsConUUIDDistintos`](../../internal/backlog/public_id_test.go#L80); [`TestCreate_CreacionesSimultaneasDeSprintsConUUIDDistintos`](../../internal/sprint/public_id_test.go#L78); [`TestRegister_RegistrosSimultaneosConUUIDDistintos`](../../internal/member/public_id_test.go#L99) (member); [`TestRegister_RegistrosSimultaneosConUUIDDistintos`](../../internal/defect/public_id_test.go#L106) (defect); [`TestRegister_RegistrosSimultaneosConUUIDDistintos`](../../internal/effort/public_id_test.go#L103) (effort); [`TestOpenSession_AperturasSimultaneasConUUIDDistintos`](../../internal/poker/public_id_test.go#L108) (en `public_id_test.go` de cada paquete), con `-race` en la CI | Las 6 funciones de creación de CA-52.1, que generan el UUID dentro del lock que ya toman (en esfuerzos y defectos, dentro de `WithOpenSprint`) |

### Ítems e integrantes (Cardozo Leandro)

**25 tests nuevos (funciones `Test`), 2 runners de godog y 31 escenarios BDD** (17 de ítems y 14 de integrantes), en la rama `feature/HU-52-uuid-items-integrantes`, desde la base y con la rama de integración traída en `bebcfc4`. Se reparten así:

- 11 de ítems en `internal/httpapi/items_uuid_test.go`, 10 de integrantes en `internal/httpapi/members_uuid_test.go`, 3 de la interfaz en `internal/httpapi/items_members_uuid_web_test.go` y 1 del helper de rutas de los tests en `internal/httpapi/entity_ids_helper_test.go` (segunda desviación de la base);
- los runners `TestFeatures_HU52Items` y `TestFeatures_HU52Integrantes` en `internal/httpapi/items_members_uuid_features_test.go`, sobre `features/HU-52-items.feature` y `features/HU-52-integrantes.feature`, a nivel de la API.

Cada test lleva `// HU-52 · CA-52.N · Verifica que …`, y cada runner `// HU-52 · BDD · Verifica que …` (contados con `grep`: CA-52.1, 2; CA-52.2, 2; CA-52.3, 8; CA-52.4, 10; CA-52.6, 3). La correspondencia con los escenarios no es 1 a 1, y los tests de esta parte prueban solo las 5 rutas de su fila en D3 (`GET` y `PUT /api/items/{id}`, `PATCH …/estado` y `…/story-points`, y `PUT /api/integrantes/{id}`), más la creación y la lista del backlog y de los integrantes. Los 1000 identificadores (CA-52.2), PostgreSQL (CA-52.5) y la concurrencia (CA-52.7) los cubre la base.

#### Criterio → test → código

| Criterio | Escenarios BDD | Tests de Go (archivo) | Código que lo implementa (función y archivo) |
|---|---|---|---|
| CA-52.1 Al crear un ítem o registrar un integrante, su id es un UUID v4 | "Crear un ítem le asigna un UUID versión 4", "El primer ítem no es el 1", "Registrar un integrante le asigna un UUID versión 4" y "El primer integrante no es el 1" (runners [`TestFeatures_HU52Items`](../../internal/httpapi/items_members_uuid_features_test.go#L21) y [`TestFeatures_HU52Integrantes`](../../internal/httpapi/items_members_uuid_features_test.go#L26)) | [`TestCrearItem_IDEsUUIDVersion4`](../../internal/httpapi/items_uuid_test.go#L131) (`internal/httpapi/items_uuid_test.go`); [`TestRegistrarIntegrante_IDEsUUIDVersion4`](../../internal/httpapi/members_uuid_test.go#L86) (`internal/httpapi/members_uuid_test.go`) | [`toItemResponse`](../../internal/httpapi/items.go#L212) y [`server.createItem`](../../internal/httpapi/items.go#L52) ([`internal/httpapi/items.go`](../../internal/httpapi/items.go)); [`server.registerMember`](../../internal/httpapi/members.go#L27) ([`internal/httpapi/members.go`](../../internal/httpapi/members.go)) |
| CA-52.2 Dos ítems o dos integrantes no tienen el mismo id | "El backlog identifica cada ítem con su UUID" y "La lista identifica cada integrante con su UUID" | [`TestCrearItem_DosItemsTienenUUIDDistintos`](../../internal/httpapi/items_uuid_test.go#L148); [`TestRegistrarIntegrante_DosIntegrantesTienenUUIDDistintos`](../../internal/httpapi/members_uuid_test.go#L98) | Las mismas de CA-52.1, con el `PublicID` que genera la base |
| CA-52.3 Las 5 rutas aceptan el UUID (también en mayúsculas) y ninguna respuesta de ítem o integrante expone el entero; un ítem sin sprint responde `""`, y con sprint, el UUID del sprint | "Consultar un ítem por su UUID", "El UUID del ítem se acepta en mayúsculas", "Modificar, cambiar el estado y estimar un ítem por su UUID", "El backlog identifica cada ítem con su UUID y un ítem sin sprint con \"\"", "Modificar un integrante por su UUID", "El UUID del integrante se acepta en mayúsculas", "La lista identifica cada integrante con su UUID" y "Consultar un integrante por su UUID sigue sin estar permitido" | [`TestRutasDeItem_AceptanElUUID`](../../internal/httpapi/items_uuid_test.go#L161), [`TestRutasDeItem_AceptanElUUIDEnMayusculas`](../../internal/httpapi/items_uuid_test.go#L178), [`TestRespuestasDeItem_NoExponenElIDEntero`](../../internal/httpapi/items_uuid_test.go#L283) y [`TestConsultarItem_SprintIDEsElUUIDDelSprint`](../../internal/httpapi/items_uuid_test.go#L319) (`internal/httpapi/items_uuid_test.go`); [`TestModificarIntegrante_AceptaElUUID`](../../internal/httpapi/members_uuid_test.go#L110), [`TestModificarIntegrante_AceptaElUUIDEnMayusculas`](../../internal/httpapi/members_uuid_test.go#L123) y [`TestListarIntegrantes_IdentificaCadaIntegranteConSuUUID`](../../internal/httpapi/members_uuid_test.go#L185) (`internal/httpapi/members_uuid_test.go`); [`TestWithEntityIDs_FueraDeLaTablaTraduceElUUIDAlEntero`](../../internal/httpapi/entity_ids_helper_test.go#L73) (`internal/httpapi/entity_ids_helper_test.go`, el helper de los tests) | [`server.getItem`](../../internal/httpapi/items.go#L99), [`server.updateItem`](../../internal/httpapi/items.go#L114), [`server.changeItemStatus`](../../internal/httpapi/items.go#L134), [`server.assignStoryPoints`](../../internal/httpapi/items.go#L154), [`server.listItems`](../../internal/httpapi/items.go#L74) y [`toItemResponse`](../../internal/httpapi/items.go#L212) ([`internal/httpapi/items.go`](../../internal/httpapi/items.go)); [`server.updateMember`](../../internal/httpapi/members_update.go#L9) ([`internal/httpapi/members_update.go`](../../internal/httpapi/members_update.go)); [`server.listMembers`](../../internal/httpapi/members_list.go#L9) ([`internal/httpapi/members_list.go`](../../internal/httpapi/members_list.go)); en los tests, [`withEntityIDs`](../../internal/httpapi/entity_ids_test.go#L184) ([`internal/httpapi/entity_ids_test.go`](../../internal/httpapi/entity_ids_test.go)) |
| CA-52.4 Un id que no es un UUID responde 400 con el mensaje del ítem o del integrante; un UUID inexistente o el nulo, 404; se conserva el orden de errores de HU-07 y HU-16 | "Un ID numérico de ítem se rechaza en cada ruta" (4 ejemplos), "Otras formas del UUID del ítem se rechazan" (3), "Un UUID que no corresponde a ningún ítem", "El UUID nulo no corresponde a ningún ítem", "El ID del ítem se valida antes que el cuerpo", "Con un UUID inexistente se conserva el orden de errores de HU-07", y los mismos 6 casos de integrantes (HU-16) | [`TestRutasDeItem_IDNumericoDevuelve400`](../../internal/httpapi/items_uuid_test.go#L195), [`TestRutasDeItem_OtrasFormasDelUUIDDevuelven400`](../../internal/httpapi/items_uuid_test.go#L207), [`TestRutasDeItem_UUIDInexistenteDevuelve404`](../../internal/httpapi/items_uuid_test.go#L219), [`TestOrdenDeErroresItem_IDInvalidoAntesQueElCuerpo`](../../internal/httpapi/items_uuid_test.go#L231) y [`TestOrdenDeErroresItem_UUIDInexistenteConservaElOrden`](../../internal/httpapi/items_uuid_test.go#L244) (`internal/httpapi/items_uuid_test.go`); [`TestModificarIntegrante_IDNumericoDevuelve400`](../../internal/httpapi/members_uuid_test.go#L136), [`TestModificarIntegrante_OtrasFormasDelUUIDDevuelven400`](../../internal/httpapi/members_uuid_test.go#L146), [`TestModificarIntegrante_UUIDInexistenteDevuelve404`](../../internal/httpapi/members_uuid_test.go#L157), [`TestOrdenDeErroresIntegrante_IDInvalidoAntesQueElCuerpo`](../../internal/httpapi/members_uuid_test.go#L167) y [`TestOrdenDeErroresIntegrante_UUIDInexistenteConservaElOrden`](../../internal/httpapi/members_uuid_test.go#L176) (`internal/httpapi/members_uuid_test.go`) | Los handlers de CA-52.3, que leen el `{id}` con `itemIDFromPath` y `memberIDFromPath` de la base antes del cuerpo; el centinela `missingInternalID` de la base, que deja la inexistencia en el mismo punto del `Manager` |
| CA-52.6 La interfaz usa el id como texto: «ID del integrante» acepta el UUID, y el backlog y las historias de los reportes muestran "—" para un ítem sin sprint (`""`) | — | [`TestInterfaz_IDDelIntegranteAceptaElUUIDComoTexto`](../../internal/httpapi/items_members_uuid_web_test.go#L32), [`TestInterfaz_BacklogMuestraSinSprintConElTextoVacio`](../../internal/httpapi/items_members_uuid_web_test.go#L48) y [`TestInterfaz_NoConvierteLosIDsDeItemsEIntegrantesANumero`](../../internal/httpapi/items_members_uuid_web_test.go#L60) (`internal/httpapi/items_members_uuid_web_test.go`); la aserción de `filaHistoriaReporte` en `web_test.go` | El campo `id` de `form-modificar-integrante` en [`internal/web/index.html`](../../internal/web/index.html); `filaItem` y `filaHistoriaReporte` en [`internal/web/app.js`](../../internal/web/app.js) |

#### Tests existentes adaptados

En el RED (`2350f63`), para ítems e integrantes:

- `items_test.go`: las comparaciones del `id` con un número pasan a `itemRef(t, h, n)`, el `sprint_id` de un ítem sin asignar pasa de `"0"` a `""` (D7), los 4 casos de "inexistente" con `999` pasan a un UUID que no existe, el 400 de `…/story-points` espera "el id del ítem debe ser un UUID", y el 500 por error del repositorio usa un UUID.
- `server_test.go`: las 2 comparaciones del `id` de los ítems usan `itemRef`; la tabla de `TestRutasConID_IDNoNumericoDevuelve400` suma `uuidErrorByCase`, con un bloque por responsable y los 4 casos de esta parte; la fila "modificar integrante" de los cuerpos usa un UUID.
- Los 7 JSON de integrantes con el id escrito a mano (`members_test.go:61`, `members_update_test.go:42`, `63` y `138`, `members_list_test.go:40` y `41`, y la fila "registrar integrante" de `server_test.go`) usan `jsonMemberID(t, h, n)`. Como el `"id":1` probaba que no se había guardado nada antes y con UUID el id ya no cuenta los registros, `expectOnlyOneMember` verifica con la lista que hay un solo integrante.
- `members_update_test.go`: el caso de inexistente pasa de 999, 0 y -1 a un UUID inexistente y al nulo (el 0 y el negativo ya no son un UUID y dan 400), y el 500 usa un UUID.

**Cambios de comportamiento de HU-51** (cierran la parte de ítems e integrantes del pendiente R1 de HU-51):

- `TestRutasDeOtrasEntidades_ConservanElIDEntero` (`projects_uuid_test.go`): las filas de ítems e integrantes esperan su mensaje de UUID. Después de traer la rama de integración, las 2 filas de sprints esperan "el id del sprint debe ser un UUID" (la parte de Buttini Cristobal), y queda una fila de García Santiago que todavía lee un entero, `GET /api/items/abc/esfuerzos`, con el comentario de que García la cambia en su GREEN. Antes de fijarla se comprobó con una corrida que responde 400 con "el id debe ser un número entero".
- `TestRespuestas_IdentificanElProyectoConSuUUID` (`projects_uuid_test.go`): usa `do()` en lugar de `serveRaw`, así que `/api/items/1` y `/api/integrantes/1` se traducen al UUID.
- `features/HU-51-identificadores-uuid.feature`, **línea 86**: "Un ítem consultado por su propio ID" pide `/api/items/{ítem}`, y el paso guarda el id que devolvió la API (`uuid_features_test.go`).
- `features/HU-51-identificadores-uuid.feature`, **línea 144**: "Las demás entidades conservan su ID entero" pasó de `/api/items/abc` a `/api/sprints/abc/metricas` en el RED, y al traer la rama de integración a `/api/items/abc/esfuerzos` (García), con la nota en el comentario del escenario. El mensaje de la línea 146 queda igual.

**Tests de otros responsables que se tocaron**, por la segunda desviación de la base (ver más abajo) o al traer la rama de integración:

| Archivo (responsable) | Qué cambió | Motivo |
|---|---|---|
| `sprints_test.go` (Buttini Cristobal), líneas 124, 204 y 281 | El `id` del ítem se compara con `itemRef`; en la 124, al resolver el conflicto del merge, queda la versión de Buttini: el `sprint_id` contra `sprintRef(t, h, 1)` | El `id` del ítem pasa a ser UUID; con los sprints en UUID, el `sprint_id` se compara con el UUID real del sprint |
| `reports_test.go` (base, HU-38 y HU-49) | `idInt` pasa a `itemNumber` en la lista de historias, y el contrato del reporte del proyecto arma el `id` y el `sprint_id` de cada historia con `jsonItemID` y `jsonItemSprintID` | El `id` de cada historia pasa a ser UUID |
| `sprint_reports_test.go` (Buttini Cristobal, HU-39) | El contrato del reporte del sprint arma el `id` y el `sprint_id` de cada historia con `jsonItemID` y `jsonItemSprintID` | Igual que en el reporte del proyecto |
| `web_test.go` (HU-49) | La aserción de `filaHistoriaReporte` compara `h.sprint_id` con `""` | Las historias de los reportes salen de `toItemResponse`, así que "sin sprint" pasa de `0` a `""` (D7) |
| `projects_uuid_test.go` y `features/HU-51-identificadores-uuid.feature` (HU-51) | Lo descrito en "Cambios de comportamiento de HU-51" | Ítems, integrantes y sprints ya no tienen ID entero |

Al traer la rama de integración también se resolvió que `itemRef` estaba declarada dos veces (en `items_uuid_test.go` y en `sprints_uuid_test.go`, de Buttini, con la misma firma): se borró la de `items_uuid_test.go` y se usa la de Buttini, sin tocar su archivo.

**Limitación de `jsonItemSprintID`.** Los contratos de los reportes leen el `sprint_id` esperado de cada historia con `GET /api/items/{id}`, así que lo comparan contra la propia API. Que sea el UUID correcto del sprint lo verifica, contra el dominio, [`TestConsultarItem_SprintIDEsElUUIDDelSprint`](../../internal/httpapi/items_uuid_test.go#L319).

#### Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | 2350f63 | `test(HU-52): agregar tests de ítems e integrantes (RED)` |
| GREEN | ca39108 | `feat(HU-52): identificar los ítems y los integrantes con un UUID público (GREEN)` |
| Merge de la rama de integración (con sprints y Planning Poker, PR #204) | bebcfc4 | `Merge remote-tracking branch 'origin/feature/HU-52-uuid-entidades' into feature/HU-52-uuid-items-integrantes` |
| Refactor | — | No hubo refactor |

#### RED (`2350f63`)

`go test -count=1 -p 1 ./...` en memoria: 14 paquetes en ok y solo `internal/httpapi` en FAIL, con 42 tests de primer nivel que fallan, todos de ítems e integrantes y por aserción, sin panics ni SKIP. Los escenarios fallan 17 de 17 en ítems y 13 de 14 en integrantes (el que pasa es el 405 de `GET /api/integrantes/{id}`, que ya funcionaba). El `-update` de los links cambió 34 líneas en 8 evidencias, solo números de línea.

#### GREEN (`ca39108`) y merge (`bebcfc4`)

Con el GREEN, antes de traer la rama de integración, `internal/httpapi` dio 779 PASS, 0 FAIL y 0 SKIP. Después del merge (`bebcfc4`), corrido el 15/11/2026:

| Corrida | `internal/httpapi` | Resto |
|---|---|---|
| En memoria (`go test -count=1 -p 1 -v ./internal/httpapi/`) | 800 PASS, 13 FAIL, 0 SKIP | `go test -count=1 -p 1 ./...`: los otros 14 paquetes en ok |
| Con PostgreSQL (`TEST_DATABASE_URL='postgres://gpc:gpc_dev@localhost:5433/gpc_test?sslmode=disable' go test -count=1 -p 1 -v ./...`) | 800 PASS, 13 FAIL, 0 SKIP | 0 FAIL y 0 SKIP en los otros 14 paquetes: `internal/postgres` 2 PASS, `internal/postgrestest` 15 PASS, y pasan los tests de `postgres_repository_test.go` de `backlog` (3), `member` (3), `project` (9), `defect` (10), `effort` (7) y `poker` (2) |

Los escenarios de esta parte pasan todos: 17 de 17 en ítems (94 pasos) y 14 de 14 en integrantes (70 pasos). `gofmt -l .` sin salida, `go vet ./...` sin errores y `openspec validate hu-52-identificadores-uuid-entidades --strict` válido. La base local se recreó con `docker compose down -v` y `docker compose up -d --wait` (D8), y se comprobó que `backlog_items` de `gpc_test` tiene `public_id` de tipo `uuid`.

**Los 13 fallos que quedan dependen de la parte de García Santiago** (defectos y esfuerzos). Todos responden 400, "el cuerpo del pedido no es un JSON válido", al registrar el defecto en los helpers de los ejemplos de TEC-03: el cuerpo lleva `"sprint_deteccion_id":%s` con el `id` del sprint, que ahora es un UUID sin comillas (el comentario "García: cuando sprint pase a UUID, este id va entre comillas" lo anticipa), y `defects.go` todavía lee un entero.

| Archivo | Tests |
|---|---|
| `metrics_test.go` | [`TestConsultaMetricas_ContratoDelSprint`](../../internal/httpapi/metrics_test.go#L634), [`TestConsultaMetricas_ContratoDelProyecto`](../../internal/httpapi/metrics_test.go#L660), [`TestMetricasPorSprint_ContratoConUnSprintCerradoYOtroAbierto`](../../internal/httpapi/metrics_test.go#L833) |
| `report_pdf_test.go` | [`TestReportePDF_Devuelve200ConElArchivo`](../../internal/httpapi/report_pdf_test.go#L54), [`TestReportePDF_TieneLosMismosValoresQueElReporteJSON`](../../internal/httpapi/report_pdf_test.go#L93) |
| `reports_test.go` | [`TestReporteProyecto_ContratoDelEjemploDeTEC03`](../../internal/httpapi/reports_test.go#L85), [`TestReporteProyecto_CadaParteEsIgualASuEndpoint`](../../internal/httpapi/reports_test.go#L255), [`TestReporteProyecto_CadaHistoriaEsIgualASuConsultaEnElBacklog`](../../internal/httpapi/reports_test.go#L399), [`TestReporteProyecto_FallaLaLecturaDeLasHistoriasDevuelve500`](../../internal/httpapi/reports_test.go#L457) |
| `sprint_reports_test.go` | [`TestReporteSprint_ContratoDelEjemploDeTEC03`](../../internal/httpapi/sprint_reports_test.go#L76), [`TestReporteSprint_CadaParteEsIgualASuEndpoint`](../../internal/httpapi/sprint_reports_test.go#L223), [`TestReporteSprint_CadaHistoriaEsIgualASuConsultaEnElBacklog`](../../internal/httpapi/sprint_reports_test.go#L360), [`TestReporteSprint_FallaLaLecturaDeLasHistoriasDevuelve500`](../../internal/httpapi/sprint_reports_test.go#L375) |

Como referencia, la rama de integración sola (`041889c`, copiada con `git archive`) dio en `internal/httpapi` 637 PASS, 119 FAIL y 0 SKIP, porque las rutas de sprints y Planning Poker necesitaban los ítems en UUID.

**Discrepancia de puerto.** `docker-compose.yml` publica la base en `127.0.0.1:${POSTGRES_PORT:-5433}` y no hay `.env`, así que el puerto es 5433, como dice el README. `tasks.md` (comandos de cierre) dice 5434. Se usó 5433, y `tasks.md` no se corrigió.

**No verificado:**

- En el navegador: que el campo «ID del integrante» acepte el UUID como texto y que el backlog muestre "—" para un ítem sin sprint. Lo cubren solo los tests del contenido de `index.html` y `app.js`, sin un motor de JavaScript.
- `-race` (CA-52.7): en Windows no anda; lo corre la CI.

### Sprints y Planning Poker (Buttini Cristobal)

**15 tests nuevos (funciones `Test`), 2 runners de godog y 17 escenarios BDD** (8 de sprints y 9 de Planning Poker), en la rama `feature/HU-52-uuid-sprints-poker`, desde la rama de integración. Se reparten así:

- 6 de sprints en `internal/httpapi/sprints_uuid_test.go`, 7 de sesiones en `internal/httpapi/poker_uuid_test.go` y 2 de la interfaz en `internal/httpapi/sprints_poker_uuid_web_test.go`;
- los runners `TestFeatures_HU52Sprints` y `TestFeatures_HU52PlanningPoker` en `internal/httpapi/sprints_poker_uuid_features_test.go`, sobre `features/HU-52-sprints.feature` y `features/HU-52-planning-poker.feature`, a nivel de la API.

Los tests y los escenarios crean los sprints, los ítems y las sesiones con los `Manager` del dominio y usan su `PublicID`, así prueban solo las 11 rutas de esta parte (D3) sin depender de cómo responde la API de los ítems. Las cuentas de la tabla no suman 15 porque un test puede cubrir más de un criterio. Los 1000 identificadores (CA-52.2), PostgreSQL (CA-52.5) y la concurrencia (CA-52.7) los cubre la base.

**Lo que queda en las otras partes (D3).** La tarea 3.2 nombra también el orden de errores de "integrante vacío" (HU-26) y el `sprint_id` de los ítems y de los esfuerzos. Esa ruta (`POST /api/sprints/{id}/items/{itemId}/esfuerzos`, `efforts.go`) y esas respuestas (`toItemResponse` y la de esfuerzos) son de García Santiago y de Cardozo Leandro, así que los prueban ellos. Esta parte cubre el `sprint_id` de las métricas, el `sprint_actual` del resumen, las métricas por sprint y el reporte del sprint, que salen de `toSprintResponse` y `toSprintMetricsResponse`.

#### Criterio → test → código

| Criterio | Escenarios BDD | Tests de Go (archivo) | Código que lo implementa (función y archivo) |
|---|---|---|---|
| CA-52.1 Al crear un sprint o abrir una sesión, su id es un UUID v4; repetir la ronda lo conserva | "Crear un sprint le asigna un UUID versión 4", "El primer sprint no es el 1", "Abrir una sesión le asigna un UUID versión 4" y "Repetir la ronda conserva el UUID de la sesión" (runners [`TestFeatures_HU52Sprints`](../../internal/httpapi/sprints_poker_uuid_features_test.go#L34) y [`TestFeatures_HU52PlanningPoker`](../../internal/httpapi/sprints_poker_uuid_features_test.go#L39)) | [`TestCrearSprint_IDEsUnUUIDVersion4`](../../internal/httpapi/sprints_uuid_test.go#L186) (`internal/httpapi/sprints_uuid_test.go`); [`TestAbrirSesion_IDEsUnUUIDVersion4YItemIDEsElDelItem`](../../internal/httpapi/poker_uuid_test.go#L53) y [`TestRepetirRonda_ConservaElUUIDDeLaSesion`](../../internal/httpapi/poker_uuid_test.go#L69) (`internal/httpapi/poker_uuid_test.go`) | [`toSprintResponse`](../../internal/httpapi/sprints.go#L174) ([`internal/httpapi/sprints.go`](../../internal/httpapi/sprints.go)); [`server.openPokerSession`](../../internal/httpapi/poker.go#L51) y [`server.repeatRoundPokerSession`](../../internal/httpapi/poker.go#L148) ([`internal/httpapi/poker.go`](../../internal/httpapi/poker.go)) |
| CA-52.3 Las 11 rutas aceptan el UUID (también en mayúsculas) y ninguna respuesta expone el entero del sprint ni de la sesión | "Cerrar un sprint por su UUID", "El UUID del sprint se acepta en mayúsculas", "Asignar un ítem a un sprint por sus UUID", "Votar en una sesión por su UUID", "El UUID de la sesión se acepta en mayúsculas" y "Revelar la sesión no expone ningún número" | [`TestRutasDeSprint_AceptanElUUID`](../../internal/httpapi/sprints_uuid_test.go#L198) y [`TestRespuestasDeSprint_IdentificanElSprintConSuUUID`](../../internal/httpapi/sprints_uuid_test.go#L261) (`internal/httpapi/sprints_uuid_test.go`); [`TestRutasDeSesion_AceptanElUUID`](../../internal/httpapi/poker_uuid_test.go#L82) (`internal/httpapi/poker_uuid_test.go`) | [`server.assignItem`](../../internal/httpapi/sprints.go#L83), [`server.completeItem`](../../internal/httpapi/sprints.go#L98), [`server.listCompletedItems`](../../internal/httpapi/sprints.go#L113), [`server.closeSprint`](../../internal/httpapi/sprints.go#L128), [`server.sprintItemPublicIDs`](../../internal/httpapi/sprints.go#L189) y [`toSprintResponse`](../../internal/httpapi/sprints.go#L174) ([`internal/httpapi/sprints.go`](../../internal/httpapi/sprints.go)); las 5 rutas de [`internal/httpapi/poker.go`](../../internal/httpapi/poker.go), de [`server.openPokerSession`](../../internal/httpapi/poker.go#L51) a [`server.estimatePokerSession`](../../internal/httpapi/poker.go#L173); [`server.sprintMetrics`](../../internal/httpapi/metrics.go#L47) y [`toSprintMetricsResponse`](../../internal/httpapi/metrics.go#L65) ([`internal/httpapi/metrics.go`](../../internal/httpapi/metrics.go)); [`server.sprintReport`](../../internal/httpapi/sprint_reports.go#L24) ([`internal/httpapi/sprint_reports.go`](../../internal/httpapi/sprint_reports.go)) |
| CA-52.4 Un id que no es un UUID responde 400 con el mensaje del sprint, de la sesión o del ítem; un UUID inexistente, 404; se conserva el orden de errores | "Un ID numérico de sprint responde 400", "Primero se valida el sprint y después el ítem", "Un UUID de sprint que no existe responde 404", "Un ID numérico de sesión responde 400", "El ID de la sesión se valida antes que el cuerpo", "Un UUID de sesión que no existe responde 404" y "Al votar se conserva el orden de errores" | [`TestRutasDeSprint_IDInvalidoDevuelve400`](../../internal/httpapi/sprints_uuid_test.go#L220), [`TestRutasDeSprint_UUIDInexistenteDevuelve404`](../../internal/httpapi/sprints_uuid_test.go#L233) y [`TestAsignarYCompletar_ElSprintSeValidaAntesQueElItem`](../../internal/httpapi/sprints_uuid_test.go#L245) (`internal/httpapi/sprints_uuid_test.go`); [`TestRutasDeSesion_IDInvalidoDevuelve400`](../../internal/httpapi/poker_uuid_test.go#L118), [`TestRutasDeSesion_UUIDInexistenteDevuelve404`](../../internal/httpapi/poker_uuid_test.go#L130), [`TestVotar_SeConservaElOrdenDeErrores`](../../internal/httpapi/poker_uuid_test.go#L141) y [`TestAbrirSesion_ItemInvalidoOInexistente`](../../internal/httpapi/poker_uuid_test.go#L150) (`internal/httpapi/poker_uuid_test.go`) | Los handlers de CA-52.3, que leen el `{id}` con `sprintIDFromPath`, `sessionIDFromPath` o `itemIDFromPath` de la base, antes del cuerpo; [`server.sprintItemPublicIDs`](../../internal/httpapi/sprints.go#L189), que lee el `{id}` antes que el `{itemId}` |
| CA-52.6 La interfaz usa el UUID del sprint y de la sesión | — | [`TestInterfaz_SprintActualAceptaElUUIDComoTexto`](../../internal/httpapi/sprints_poker_uuid_web_test.go#L27) y [`TestInterfaz_NoConvierteElIDDelSprintNiDeLaSesionANumero`](../../internal/httpapi/sprints_poker_uuid_web_test.go#L43) (`internal/httpapi/sprints_poker_uuid_web_test.go`) | El campo «Sprint actual» de [`internal/web/index.html`](../../internal/web/index.html) (`type="text"`, `placeholder="UUID del sprint"`); `mostrarSprint` y Planning Poker de [`internal/web/app.js`](../../internal/web/app.js), que ya usaban el `id` como texto |

#### Tests existentes adaptados

25 tests existentes de los archivos de esta parte se adaptaron al comportamiento nuevo, sin cambiar lo que verifican: los mensajes de 400 pasan a "el id del sprint debe ser un UUID", "el id de la sesión debe ser un UUID" o "el id del ítem debe ser un UUID"; los casos de inexistente usan un UUID que nunca se asignó (y el nulo) en lugar de 999, 0 y -1, que ahora son un formato inválido; las sesiones guardadas directo en el repositorio se piden por su `PublicID`; y los `id` y `sprint_id` de sprints y sesiones se comparan con el `id` que devolvió la API al crearlos (`sprintRef`, `sessionRef` e `itemRef`), y no con `"1"`, porque con el UUID deja de ser 1. Además:

- las 11 rutas de esta parte van a su bloque de `uuidRoutes` (`entity_ids_test.go`), en el RED y no en el GREEN, porque es código de test: mientras la producción no cambia, la traducción de `do()` es la identidad;
- en `TestRutasConID_IDNoNumericoDevuelve400`, el mensaje de cada caso de esta parte sale de `sprintsPokerInvalidIDMessages`;
- en `TestRespuestas_IdentificanElProyectoConSuUUID` (HU-51), `serveRaw` pasa a `do()`, para que las rutas de sprints usen el UUID que devolvió la API.

#### Commits

| Fase | Commit |
|---|---|
| RED | `d0e3f13` |
| GREEN | `3ecab8e` |
| Refactor | No hubo: la repetición de `pokerSessionResponse` en las 5 rutas de `poker.go` ya existía, y sacarla a una función ahora agrega riesgo de conflicto con el corte cerca |

#### RED (`d0e3f13`)

`go test -count=1 ./...` en memoria: 14 paquetes en `ok` (incluido `internal/trazabilidad`) y `internal/httpapi` en `FAIL`, con 41 tests que fallan, todos de esta parte y por aserción (0 de compilación y 0 `panic`): 16 de los 17 tests nuevos (el que verifica que app.js no convierte el id a número ya pasaba, porque protege lo que ya estaba bien), los 2 runners y 25 tests existentes adaptados. Ningún test de otra parte falla.

#### GREEN (`3ecab8e`)

Cambian solo `sprints.go`, `poker.go`, `metrics.go`, `sprint_reports.go`, `index.html` y los números de línea de 4 evidencias (`-update`); ningún `_test.go` ni `.feature`. `go test -count=1 ./...` en memoria: 14 paquetes en `ok` y `internal/httpapi` en `FAIL`, con 86 tests que fallan y 111 mensajes, todos de las otras partes, como prevé el cambio en D3:

- 106 "el id del ítem debe ser un UUID": las rutas de esta parte reciben el ítem por su UUID, y la API todavía devuelve el ítem con su entero (parte de ítems, Cardozo Leandro);
- 3 "el id debe ser un número entero": las rutas de esfuerzos y defectos todavía leen un entero (parte de García Santiago);
- 2 que son consecuencia de lo anterior: un ítem que no se pudo asignar ni completar deja vacíos los completados y el reporte.

Para verificarlo, se simuló la parte de ítems sin commitearla (el `id` y el `sprint_id` del ítem como texto, sus 4 rutas con UUID y en `uuidRoutes`): con eso no falla ningún test de esta parte, y lo que queda depende de la ruta de registrar un esfuerzo (García Santiago) o es de los tests de ítems.

### Defectos y esfuerzos (García Santiago)

Pendiente de su responsable.

### PDF del reporte (tarea 1.8) · Peñalbé Hernán

Rama `feature/HU-52-uuid-pdf`, sobre `feature/HU-52-uuid-base`.

#### Commits

| Fase | Commit | Qué hace |
|---|---|---|
| RED | `3229e3f` | Las líneas esperadas del PDF usan marcadores (`{historia1}`, `{sprint1}`, `{esfuerzo1}`, `{defecto1}`) que se reemplazan por los UUID, y un test nuevo verifica que no quede ningún ID entero |
| GREEN | `4461a88` | `pdf.go` imprime los UUID; el reporte guarda los sprints que ya leía; los tests unen las líneas cortadas por el ancho de la página y los dos escenarios BDD de HU-40 y HU-49 usan los marcadores |
| Refactor | — | No hubo refactor |

#### RED

```
internal/report:  pdf_test.go:186:16: rep.Sprints undefined (type ProjectReport has no field or method Sprints)  [build failed]
internal/httpapi: --- FAIL: TestReportePDF_Devuelve200ConElArchivo
                  el PDF todavía muestra el ID entero en "#3 · 2026-10-08" / "#1 · ALTA" / "ítem 1 ·" / "sprint 1 ·"
```

Sin los tests nuevos, `internal/report` e `internal/httpapi` estaban en verde: el fallo es de HU-52.

**Desvío 1 · El RED de `internal/report` no compila.** Falla por "build failed" (`rep.Sprints` no existe todavía), no por una aserción. En los demás PR se agregó primero la firma vacía para que el RED falle por aserción. Por el corte del sprint no se rehízo el commit; el fallo sí corresponde a la funcionalidad nueva (el campo que agrega HU-52). Observado por García Santiago en la revisión del PR #200.

#### GREEN

```
$ go test -count=1 ./...
15 paquetes ok, 0 fallos

$ go test -cover ./internal/report/ ./internal/httpapi/
ok  seguimiento-medicion-gpc/internal/report   coverage: 100.0% of statements
ok  seguimiento-medicion-gpc/internal/httpapi  coverage: 97.5% of statements
```

**Desvío 2 · El GREEN cambia tests y `.feature`.** El commit `4461a88` toca 5 archivos de tests y 2 `.feature`. Lo correcto era un commit aparte antes del GREEN, como `20f2bf1` en HU-51; por el corte del sprint no se rehízo (observado por García Santiago en el PR #200). Los ajustes: con los UUID, las líneas de historias, esfuerzos y defectos superan el ancho de la página y siguen en la línea siguiente. Los tests que buscaban la línea entera en los bytes del PDF ahora buscan en `pdfContent`, que une las líneas del PDF con un espacio. El helper `saveDefect` guarda directo en el repositorio y ahora le asigna un UUID, como hace `defect.Register`.

#### Trazabilidad

1 test nuevo y 7 tests y 2 escenarios BDD adaptados, que cubren CA-52.3 en el PDF.

| Criterio | Tests | Código |
|---|---|---|
| CA-52.3 (el PDF no muestra IDs enteros de historias, esfuerzos, defectos ni sprints) | [`TestProjectReportLines_SinIDsEnteros`](../../internal/report/pdf_test.go#L215), [`TestProjectReportLines_EjemploDeTEC03`](../../internal/report/pdf_test.go#L173), [`TestProjectReportLines_HistoriasConYSinEstimacion`](../../internal/report/pdf_test.go#L244), [`TestProjectReportLines_HorasDeCadaEsfuerzoTalCual`](../../internal/report/pdf_test.go#L334), [`TestProjectReportLines_DefectoSinSprintDeResolucion`](../../internal/report/pdf_test.go#L354), [`TestWriteProjectPDF_GeneraUnPDFConCadaLineaDelReporte`](../../internal/report/pdf_test.go#L373), [`TestWriteProjectPDF_TextoLargoYVariasPaginas`](../../internal/report/pdf_test.go#L423), [`TestReportePDF_Devuelve200ConElArchivo`](../../internal/httpapi/report_pdf_test.go#L54), [`TestFeatures_HU40ExportarReportePDF`](../../internal/report/pdf_features_test.go#L23), [`TestFeatures_HU49HistoriasEnReporte`](../../internal/report/stories_features_test.go#L21) | [`pdf.go`](../../internal/report/pdf.go) ([`publicIDsOf`](../../internal/report/pdf.go#L167), [`effortText`](../../internal/report/pdf.go#L122), [`storyText`](../../internal/report/pdf.go#L131), [`defectText`](../../internal/report/pdf.go#L149)), [`project.go`](../../internal/report/project.go) ([`Project`](../../internal/report/project.go#L52)) |

## Base compartida

### Tests existentes adaptados

**12 tests existentes se adaptaron** sin cambiar lo que verifican: 3 de `internal/backlog`, 8 de `internal/member` y 1 de `internal/report`. Además cambiaron funciones auxiliares de otros tests: los pasos de 3 runners de godog (HU-16, HU-26 y HU-29), `contractSave` y las comparaciones de los contratos de los 6 paquetes, las semillas de PostgreSQL de `defect` y `effort`, y `do` en `internal/httpapi/server_test.go` (la traducción, abajo). 13 repositorios falsos de los tests suman un `GetByPublicID` que no usan, para cumplir la interfaz. Las comparaciones copian el `PublicID` del valor obtenido, o lo ignoran en los contratos, porque es aleatorio.

Conteos (corridos sobre `2bcd3c6`):

```
$ grep -c '^func Test' internal/{backlog,sprint,member,defect,effort,poker}/public_id*_test.go internal/httpapi/public_ids_internal_test.go internal/httpapi/uuid_base_features_test.go
internal/backlog/public_id_repository_test.go:3
internal/backlog/public_id_test.go:6
internal/sprint/public_id_repository_test.go:3
internal/sprint/public_id_test.go:6
internal/member/public_id_repository_test.go:3
internal/member/public_id_test.go:6
internal/defect/public_id_repository_test.go:3
internal/defect/public_id_test.go:6
internal/effort/public_id_repository_test.go:3
internal/effort/public_id_test.go:6
internal/poker/public_id_repository_test.go:3
internal/poker/public_id_test.go:6
internal/httpapi/public_ids_internal_test.go:13
internal/httpapi/uuid_base_features_test.go:1
$ grep -cE '^\s*Esquema del escenario:' features/HU-52-base.feature
6
$ grep -cE '^\s*\| (un|una) ' features/HU-52-base.feature
36
```

Los 12 tests adaptados se contaron comparando cada función `Test` de `2fb318d` con la de `e039429`.

### Medición antes del RED (tarea 1.1)

Antes del RED se aplicó, en una copia temporal de `e039429` (con `git worktree`), lo que el GREEN cambia en el dominio y en el esquema: el campo `PublicID`, `uuid.NewV4()` en las 6 creaciones, `public_id` en `Save` y en el `scan` de PostgreSQL y la columna `public_id UUID NOT NULL UNIQUE` en las 6 tablas (19 archivos, +90 −44). Se volvió a correr el 15/11/2026 para esta evidencia; la copia se borró después.

| Corrida | Tests de primer nivel que fallan |
|---|---|
| En memoria | 14, más `TestEnlaces_CadaLinkApuntaALaLineaDeLoQueNombra` por las líneas corridas |
| Con PostgreSQL | 30, con 43 subtests, más `TestEnlaces_CadaLinkApuntaALaLineaDeLoQueNombra` |

- **En memoria (14):** comparan el struct completo, y el `PublicID` aleatorio no coincide.
  - `internal/backlog`: `TestListItems_ModificarLaListaDevueltaNoAlteraLosItemsGuardados`, `TestListItemsByPriority_OrdenaAltaMediaBajaConTodosSusDatos` y `TestListItemsByPriority_ModificarLaListaDevueltaNoAlteraLosItemsGuardados`;
  - `internal/member` (8): `TestList_DevuelveLosIntegrantesOrdenadosPorIDConSuRol`, `TestList_MuestraLosDatosModificados`, `TestRegister_GuardaElIntegranteConNombreYRol`, `TestRegister_AceptaLosTresRolesConIDsCorrelativos`, `TestRegister_DevuelveUnaCopia`, `TestUpdate_CambiaElNombreYElRol`, `TestUpdate_NormalizaElNombreYElRol` y el runner `TestFeatures_HU16ModificarIntegrante`;
  - los runners `TestFeatures_HU29RegistrarDefecto` (`internal/defect`) y `TestFeatures_HU26RegistrarEsfuerzo` (`internal/effort`);
  - `TestSprint_EjemploDeTEC03` (`internal/report`).
- **Con PostgreSQL, además (16):** los contratos de los 6 repositorios (`TestPostgresRepository_Contrato` en los 6 paquetes, y `ContratoGetByIDYUpdate`, `ContratoList` y `ContratoUpdate`), y 7 tests propios de PostgreSQL de `effort`, `defect` y `member`. Guardan sin `PublicID`, así que el segundo `Save` choca con `UNIQUE`, o comparan el struct completo.

Todos se adaptaron en el RED, así que el GREEN no toca tests.

### Fase RED (`2fb318d`)

Se escribieron los tests y el `.feature` antes del código. En producción, solo stubs sin lógica (31 archivos, +331 −16), para que compile y los tests fallen por aserción:

- el campo `PublicID uuid.UUID` en los 6 structs;
- `GetByPublicID` en la interfaz `Repository`, en los dos repositorios y en el `Manager` (`public_id.go`) de los 6 paquetes, que devuelven el error de no encontrado de su entidad;
- en `internal/httpapi/server.go`: los 6 errores de formato, `missingInternalID`, y los stubs de `parsePublicID` (siempre el error), de los 6 `xIDFromPath` (siempre el error), de `sprintIDFromBody` (0 sin error), del resolvedor (`""`) y de `writeResolved` (sin el 500).

El esquema no cambió en el RED (ver "Desviaciones").

Las salidas son de una copia temporal de `2fb318d` con `git worktree`, corridas el 15/11/2026 en Windows con go1.27.1. Para PostgreSQL se levantó una base con el esquema del RED (sin `public_id` en las 6 tablas), con `POSTGRES_PORT=5434`.

`go vet ./...`: sin errores.

`go test -count=1 ./...` en memoria:

```
ok   seguimiento-medicion-gpc/cmd/server
FAIL
FAIL seguimiento-medicion-gpc/internal/backlog
ok   seguimiento-medicion-gpc/internal/dashboard
FAIL
FAIL seguimiento-medicion-gpc/internal/defect
FAIL
FAIL seguimiento-medicion-gpc/internal/effort
FAIL
FAIL seguimiento-medicion-gpc/internal/httpapi
FAIL
FAIL seguimiento-medicion-gpc/internal/member
ok   seguimiento-medicion-gpc/internal/metrics
FAIL
FAIL seguimiento-medicion-gpc/internal/poker
ok   seguimiento-medicion-gpc/internal/postgres
ok   seguimiento-medicion-gpc/internal/postgrestest
ok   seguimiento-medicion-gpc/internal/project
ok   seguimiento-medicion-gpc/internal/report
FAIL
FAIL seguimiento-medicion-gpc/internal/sprint
ok   seguimiento-medicion-gpc/internal/trazabilidad
?    seguimiento-medicion-gpc/internal/web [no test files]
FAIL
```

Con `TEST_DATABASE_URL=postgres://gpc:gpc_dev@localhost:5434/gpc_test?sslmode=disable` y `-v`, los paquetes dan lo mismo.

| Corrida del RED | Tests de primer nivel que fallan | Pasan | SKIP | Panics |
|---|---|---|---|---|
| En memoria (`-v`) | 47 | 1149 | 55 (los de PostgreSQL, sin `TEST_DATABASE_URL`) | 0 |
| Con PostgreSQL (`-v`) | 59 | 1192 | 0 | 0 |
| Escenarios de `HU-52-base.feature` | `36 scenarios (12 passed, 24 failed)` | `132 steps (102 passed, 24 failed, 6 skipped)` | — | — |

- **Fallan solo tests nuevos de HU-52,** todos por aserción (por ejemplo, `PublicID = "00000000-0000-0000-0000-000000000000", se esperaba un UUID versión 4 en forma canónica`, `statusFor(error de ítem) = 500, se esperaba 400` o `Save con un PublicID repetido devolvió {…} sin error`). Ningún test existente falla: con la traducción en `do()`, tampoco los de `internal/httpapi`.
- **Los 12 de más con PostgreSQL** son `TestPostgresRepository_ContratoPublicID` y `TestPostgresRepository_PublicIDRepetidoFalla` de los 6 paquetes, que en memoria se saltean.
- **Pasan en el RED 9 tests nuevos** (los 6 de "inexistente", `TestParsePublicID_RechazaConElErrorDeLaEntidad`, `TestIDFromPath_FormatoInvalidoDaElErrorDeLaEntidad` y `TestMissingInternalID_EsElMaximoDeInteger`) **y 12 escenarios** (los de UUID nulo e inexistente), porque los stubs ya devuelven el error de no encontrado o de formato. Los 6 pasos "skipped" de godog son pasos que siguen a uno que falló.
- **`internal/trazabilidad`:** ok, con el `-update` de los links en el mismo commit.

### Fase GREEN (`2bcd3c6`)

**Condición de SKIP:** 0 SKIP con PostgreSQL. En memoria, los tests de PostgreSQL se saltean siempre (`postgrestest.Pool` hace `t.Skip` sin `TEST_DATABASE_URL`): los 55 SKIP en memoria son esos tests.

Código mínimo para que pasen los tests de la base, sin tocar tests ni `.feature` (35 archivos fuera de `docs/`, +318 −115):

- **dominio de los 6 paquetes:** `uuid.NewV4()` al crear, dentro del lock que ya toma cada creación; `GetByPublicID` en el `Manager` y en los dos repositorios; `public_id` en `Save` y en el `scan` de PostgreSQL; `PublicID: current.PublicID` en `backlog/update.go` (el hallazgo de abajo);
- **esquema:** `public_id UUID NOT NULL UNIQUE` en las 6 tablas de `db/init/01-schema.sql`;
- **`internal/httpapi/server.go`:** `parsePublicID` (y `parseProjectID` la llama), `internalID`, los 6 `xIDFromPath`, `sprintIDFromBody`, el resolvedor con su cache, `writeResolved` y los 6 errores con su 400 en `statusFor`;
- **funciones de respuesta:** las 8 de D3 (`toItemResponse`, `toSprintResponse`, `toEffortResponse`, `toEffortListResponse`, `toDefectResponse`, `toDefectListResponse`, `toSprintMetricsResponse` y `toStoryListResponse`) y las 3 que las componen (`toSummaryResponse`, `toProjectReportResponse` y `toSprintReportResponse`) reciben el resolvedor `ids`; los 19 handlers que las llaman lo crean y responden con `writeResolved`. Ninguna lo usa todavía, así que **las rutas y las salidas de la API no cambian**: no cambió ningún campo `json`, ningún handler usa los `xIDFromPath`, y los tests existentes de la API pasan con sus IDs enteros.

Las salidas son de una copia temporal de `2bcd3c6` con `git worktree`, corridas el 15/11/2026. Para PostgreSQL se recreó la base con `docker compose down -v` y `up -d --wait` (esquema del GREEN).

```
$ gofmt -l .
$ go vet ./...
$ git diff --stat 2fb318d 2bcd3c6 -- '*_test.go' features/
```

Las tres salidas están vacías.

`go test -count=1 ./...` en memoria:

```
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
ok   seguimiento-medicion-gpc/internal/trazabilidad
?    seguimiento-medicion-gpc/internal/web [no test files]
```

| Corrida del GREEN | Resultado |
|---|---|
| `go test -count=1 ./...` en memoria | 15 paquetes `ok`, 0 FAIL |
| `go test -count=1 -v ./...` en memoria | 1196 tests de primer nivel en PASS, 0 FAIL; 55 SKIP, que son los de PostgreSQL sin `TEST_DATABASE_URL` |
| `go test -count=1 -p 1 -v ./...` con PostgreSQL | 15 paquetes `ok`, 1251 tests de primer nivel, los 1251 en PASS, 0 FAIL y 0 SKIP |
| Escenarios de `HU-52-base.feature` | `36 scenarios (36 passed)` y `132 steps (132 passed)` |

#### Cobertura

`go test -cover` con PostgreSQL, sobre `main` `45d2c31` y sobre `2bcd3c6`:

| Paquete | `main` `45d2c31` | Rama `2bcd3c6` |
|---|---|---|
| `internal/backlog` | 99.0% | 98.7% |
| `internal/sprint` | 97.2% | 96.9% |
| `internal/member` | 96.0% | 95.7% |
| `internal/effort` | 96.9% | 96.7% |
| `internal/poker` | 95.3% | 95.1% |
| `internal/defect` | 95.6% | 95.6% |
| `internal/httpapi` | 97.4% | 97.5% |
| Los otros 8 paquetes con código | iguales | iguales |
| Total del perfil (`-coverprofile`) | 95.0% | 95.0% |
| Total con `-coverpkg=./...` | 95.9% | 95.9% |

Las bajas son las ramas de error de PostgreSQL de `GetByPublicID` (un error de la base que no es "no encontrado"), que ningún test provoca, como en `GetByID`. `docs/reportes/cobertura.md` se actualiza con el cierre de HU-52.

### Cambios respecto del diseño

Aprobados por Sosa Ricardo el 15/11/2026 y anotados en `design.md`, para que el Agile Enabler los vea en el PR:

1. **El resolvedor tiene solo `item(id)` y `sprint(id)`** (D5), y no un método por entidad. Los integrantes, los defectos, los esfuerzos y las sesiones nunca se nombran desde otra entidad: su `id` sale del propio struct. Además, el `Manager` de integrantes no tiene una búsqueda por ID interno.
2. **El runner de `HU-52-base.feature` está en `internal/httpapi`** (D9), porque es el único paquete que ya usa los `Manager` de las 6 entidades. Los pasos llaman a los `Manager`, no a la API.
3. **Los contratos de repositorio ignoran el `PublicID` en sus comparaciones generales** (D9), porque es aleatorio. Lo verifica el contrato nuevo `runRepositoryPublicIDContract`.

**La traducción en `do()`** (D9, [`internal/httpapi/entity_ids_test.go`](../../internal/httpapi/entity_ids_test.go)). Los tests de la API tienen 279 (230 literales y 49 armadas con `fmt.Sprintf` o concatenación, en 15 archivos de `internal/httpapi`, contadas con `git grep` sobre `e039429`) rutas con IDs fijos de estas entidades. En lugar de reescribirlas, `do()` guarda, por handler y por entidad, el `id` que devolvió la API al crear cada entidad (`POST` con 201), y traduce el número N de `/api/{entidad}/N/…` (y el `{itemId}` de `/api/sprints/N/items/M/…`) al `id` de la N-ésima entidad creada por la API. Hoy es la identidad: la API devuelve enteros correlativos, y ningún test existente cambió de resultado. En `server_test.go` cambian 2 líneas de `do`. Los casos de "inexistente" (como `/api/items/999`), las entidades guardadas directo en el repositorio, los IDs de los cuerpos y las respuestas esperadas con enteros los adapta cada responsable para su entidad.

**El PDF del reporte** quedó fuera de la base por decisión del Agile Enabler el 15/11/2026, y lo toma Peñalbé Hernán (HU-40 es suya), en paralelo con las entidades. Cambia una salida, los tests de `internal/report` verifican el texto con enteros, y depende solo de la base.

### Hallazgo: `backlog.Update` perdía el `PublicID`

`TestModificaciones_ConservanElUUIDDelItem` (CA-52.5) falló con el GREEN de prueba, antes del commit del RED: [`Manager.Update`](../../internal/backlog/update.go#L23) arma un `Item` nuevo campo por campo y no copiaba el `PublicID`, así que modificar un ítem lo dejaba con el UUID nulo, y el cambio de estado y la estimación siguientes heredaban ese UUID nulo. El GREEN suma `PublicID: current.PublicID`. Los otros 5 paquetes modifican la entidad leída, que ya trae el `PublicID`.

### Mejora pendiente: el mapa `entityIDs` no se limpia

Observación de Buttini Cristobal en la revisión del PR #201. `entityIDs` (`internal/httpapi/entity_ids_test.go`) es un mapa global, por handler, que `do()` llena y que nunca se vacía: cada test crea su handler, así que las entradas de los tests ya terminados quedan en memoria hasta el final de `go test`. No cambia ningún resultado, porque cada handler tiene su propia lista, pero crece con la cantidad de tests. Pasa lo mismo con `projectIDs` de HU-51 (`server_test.go`). Queda como mejora pendiente, sin cambiar en la base: por ejemplo, borrar la entrada del handler con `t.Cleanup` al crearlo, para los dos mapas a la vez.

### Desviaciones

1. **El `-update` de los links va dentro del RED y del GREEN.** Los imports de `uuid`, el campo `PublicID` y las líneas agregadas a los tests adaptados corren de línea 446 links de las evidencias de otras historias, y `TestEnlaces_CadaLinkApuntaALaLineaDeLoQueNombra` falla si no se corrige. Se pidió correrlo una sola vez antes del PR, pero eso dejaba `internal/trazabilidad` en rojo en el RED y en el GREEN. Sosa Ricardo decidió incluirlo en cada commit, para que `internal/trazabilidad` quede en ok en cada uno: en el RED cambia 446 links en 31 evidencias (192 líneas) y en el GREEN 194 links en 29 evidencias (107 líneas), y en los dos casos solo cambian números de línea. Se volverá a correr antes del PR.
2. **Tests que pasan en el RED:** los de "inexistente" y "nulo" (6 de dominio, `TestParsePublicID_RechazaConElErrorDeLaEntidad`, `TestIDFromPath_FormatoInvalidoDaElErrorDeLaEntidad`, `TestMissingInternalID_EsElMaximoDeInteger`, y 12 de los 36 escenarios), porque los stubs ya devuelven el error de no encontrado o de formato.
3. **El esquema no cambió en el RED:** con `public_id NOT NULL`, los tests de PostgreSQL habrían fallado por una causa ajena a los tests nuevos (como en HU-51).
4. **`-race` lo corre la CI:** en Windows no anda (`-race requires cgo`), así que los 6 tests de CA-52.7 corrieron acá sin `-race`.

### Desviación de la base (15/11/2026)

Cardozo Leandro la encontró al empezar la sección 2 (ítems e integrantes), y el Agile Enabler (Peñalbé Hernán) decidió el 15/11/2026 resolverla con un ajuste de tests en la rama `feature/HU-52-ajuste-base`, desde `main` (`29051da`), antes de las entidades. El detalle está en D9 de `design.md` ("Desviación de la base"). **Es un ajuste de tests sin ciclo RED/GREEN:** no cambia el comportamiento ni lo que verifica ningún test, y no toca código de producción.

| Faltaba | Cómo se resolvió |
|---|---|
| 7 structs de los tests (`itemJSON`, `sprintJSON`, `defectJSON`, `effortJSON`, `pokerSessionJSON`, `sprintMetricsJSON` y una anónima de `reports_test.go`) leían el id como `int`, y el GREEN de una entidad rompía los tests de las demás | El tipo `idRef` ([`ids_ref_test.go`](../../internal/httpapi/ids_ref_test.go)), que acepta un número o un texto, y la adaptación mecánica de sus usos: 60 comparaciones, literales y funciones (58 errores del compilador, más 2 literales en las mismas líneas) y 52 `%d` |
| `do()` traducía todas las rutas, también las de handlers que todavía leen un entero | La tabla `uuidRoutes` ([`entity_ids_test.go`](../../internal/httpapi/entity_ids_test.go)), vacía, con un bloque por responsable; cada uno activa sus rutas en su GREEN |
| La tabla de `TestRutasConID_IDNoNumericoDevuelve400` no estaba separada por entidad | 8 bloques separados por líneas en blanco; las 37 filas y sus 21 comentarios quedan iguales |

Controles, todos con comandos:

- **Tests de `internal/httpapi`** (`go test -count=1 -p 1 -v ./internal/httpapi/`): antes, 720 PASS, 0 FAIL y 0 SKIP; después, 722 PASS, 0 FAIL y 0 SKIP. Los 2 nuevos son los del helper: `TestWithEntityIDs_TablaVaciaNoTraduce` y `TestWithEntityIDs_TraduceSoloLasRutasDeLaTabla` ([`entity_ids_helper_test.go`](../../internal/httpapi/entity_ids_helper_test.go)).
- **`go test -count=1 -p 1 ./...`:** 15 paquetes en ok y ningún FAIL. `go vet ./...` sin errores y `gofmt -l .` sin salida.
- **Producción:** `git diff --stat origin/main -- . ':!*_test.go' ':!docs' ':!openspec'` vacío.
- **Diff de los tests:** 14 archivos `*_test.go` modificados (228 líneas agregadas y 148 borradas) y 2 nuevos. Comparando el diff sin los cambios de tipo y de formato, no cambió ninguna aserción, mensaje ni nombre: solo la tabla `uuidRoutes`, la llamada de `do()`, los encabezados de los bloques y el comentario para García de los 3 cuerpos con `sprint_deteccion_id`.
- **Links:** `go test ./internal/trazabilidad/ -run TestEnlaces -args -update` cambió 41 líneas en 13 evidencias de otras historias, y en todas solo cambian números de línea (`#Lnn`). Va en el commit de tests, para que `internal/trazabilidad` quede en ok en cada commit, como en la desviación 1.
- **Aviso de Windows:** `go: unlinkat … being used by another process` apareció al borrar ejecutables temporales; se repitió con `go clean -testcache` y volvió a aparecer, sin ningún test fallido.

### Segunda desviación de la base (15/11/2026)

La encontró Cardozo Leandro al preparar el GREEN de ítems e integrantes, y la decidió como responsable de esa parte (falta registrar la aprobación del Agile Enabler). El detalle está en D9 de `design.md` ("Segunda desviación de la base"). Con el ítem en UUID, los tests de otros responsables mandaban ese UUID a rutas que todavía leen un entero, y `uuidRoutes` no lo evita.

| Qué pasaba | Cómo se resolvió |
|---|---|
| Un `id` de la API (por ejemplo, el UUID de un ítem) llegaba a una ruta de otro responsable que todavía lee un entero, como `POST /api/sprints/{id}/items/{itemId}` | `withEntityIDs` ([`entity_ids_test.go`](../../internal/httpapi/entity_ids_test.go)), solo en los tests: fuera de `uuidRoutes`, traduce al revés el `id` registrado al número de su orden de creación (su ID interno). Lo verifica `TestWithEntityIDs_FueraDeLaTablaTraduceElUUIDAlEntero` |
| `sprints_test.go` comparaba el `id` del ítem con `"1"` y `"2"` y el `sprint_id` con `"1"` | `itemRef`, y la forma de un UUID para el `sprint_id` |
| `reports_test.go` pasaba el `id` de cada historia a entero y comparaba los contratos de los reportes con `"id": 1` y `"sprint_id": 1` | `itemNumber`, y `jsonItemID` y `jsonItemSprintID` en los contratos de `reports_test.go` y `sprint_reports_test.go` |
| Las historias de los reportes (`toItemResponse`) muestran "sin sprint" como `""` | `filaHistoriaReporte` de `app.js` compara con `""`, y se adapta su aserción en `web_test.go` |

Hasta que se integren sprints y Planning Poker, la API queda mezclada como prevé R2: el `sprint_id` de un ítem es el UUID del sprint, y el `id` del sprint sigue siendo un entero.

### Commits

| Fase | Commit | Mensaje |
|---|---|---|
| Propuesta (OpenSpec) | d49f3c8 | `docs(HU-52): agregar la propuesta de OpenSpec` |
| Merge de `main` | e039429 | `chore(HU-52): traer main con la columna Código de las evidencias` |
| Aprobación | 8b65300 | `docs(HU-52): registrar la aprobación de la propuesta` |
| RED de la base | 2fb318d | `test(HU-52): agregar los tests de la base (RED)` |
| GREEN de la base | 2bcd3c6 | `feat(HU-52): agregar el UUID público al dominio de las 6 entidades y los helpers compartidos (GREEN)` |
| RED del PDF (Peñalbé Hernán) | 3229e3f | `test(HU-52): agregar tests del PDF con UUID (RED)` |
| GREEN del PDF (Peñalbé Hernán) | 4461a88 | `feat(HU-52): mostrar los UUID en el texto del PDF (GREEN)` |
| Refactor de la base | (propuesto) | — |
| PDF fuera de la base | 66a576c | `docs(HU-52): pasar el PDF del reporte a Peñalbé Hernán` |
| Evidencia de la base | 235b781 | `docs(HU-52): agregar la evidencia de la base` |
| Merge del PR #200 (PDF) | 0a6c513 | `Merge pull request #200 from Elwilsonic/feature/HU-52-uuid-pdf` |
| Merge del PR #201 (base en `main`, Cardozo Leandro) | c329a62 | `Merge pull request #201 from Elwilsonic/feature/HU-52-uuid-base` |

## Pendiente al corte

Estado al 15/11/2026, antes del corte del lunes 16/11/2026 a las 14:00. La base (PR #201, con el PDF del PR #200 integrado en su rama) está integrada en `main` el 15/11/2026 a las 18:30 (`c329a62`), por Cardozo Leandro. Las tres partes por entidad todavía no tienen rama ni PR en el repositorio (`git ls-remote --heads origin` muestra solo `feature/HU-52-uuid-base` y `feature/HU-52-uuid-pdf`): se crean desde `main`, que ya tiene la base. **Desde el cambio en D3 del 15/11/2026, las tres partes van a la rama de integración `feature/HU-52-uuid-entidades`, y a `main` llegan juntas en un solo PR:** el corte por entidad de la tabla de abajo ya no aplica parte por parte, y si a las 12:00 del 16/11/2026 falta alguna, se decide con el Agile Enabler. Según "Si no llega al corte" de `proposal.md`, lo que no esté en `main` con la CI en verde a esa hora queda así:

| Parte | Responsable | Estado | Lo que queda en `main` si no llega | Pendiente que se documenta |
|---|---|---|---|---|
| Base compartida | Sosa Ricardo y Cardozo Leandro | PR #201, integrada en `main` el 15/11/2026 a las 18:30 (`c329a62`) | Ya está en `main` | — |
| PDF del reporte | Peñalbé Hernán | Hecho, integrado en la rama de la base (PR #200) y en `main` con el PR #201 | Ya está en `main` | — |
| Ítems | Cardozo Leandro | Pendiente | Rutas y respuestas del ítem con ID entero | Las 10 rutas de ítems y el `id` del ítem, y que el `item_id` de otras respuestas no sirve todavía para `/api/items/{id}` |
| Integrantes | Cardozo Leandro | Pendiente | `PUT /api/integrantes/{id}` y el `id` del integrante con entero | 1 ruta, 1 campo y el campo «ID del integrante» |
| Sprints | Buttini Cristobal | Pendiente | Las 9 rutas de sprints, `id` y `sprint_id` con entero, y «Sprint actual» numérico | Lo mismo que en ítems, para los campos que nombran un sprint |
| Planning Poker | Buttini Cristobal | Pendiente | Las 4 rutas de sesiones con entero | 4 rutas y el `id` de la sesión |
| Defectos | García Santiago | Pendiente | `PATCH /api/defectos/{id}/estado`, el `id` y los sprints del cuerpo con entero | 1 ruta, 2 campos del cuerpo, 4 de respuesta y 4 de la interfaz |
| Esfuerzos | García Santiago | Pendiente | `PATCH /api/esfuerzos/{id}/horas` y el `id` con entero | 1 ruta, 3 campos de respuesta y 3 de la interfaz |

Al corte, esta tabla se actualiza con lo que haya llegado (tarea 6.1). Cada pendiente se avisa en la Review y queda como riesgo abierto en `design.md`.
