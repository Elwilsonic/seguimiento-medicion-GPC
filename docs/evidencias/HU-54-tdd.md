# Evidencia TDD · HU-54 Votar y registrar esfuerzo solo con integrantes del proyecto

> Issue del repositorio: #219 · Épica E4 · Registro de integrantes · Vinculada a HU-15 (#29), HU-17 (#31), HU-21 (#39), HU-22 (#40), HU-26 (#44) y HU-53 (#211)

- **Especificación (OpenSpec):** el cambio [`openspec/changes/hu-54-integrantes-del-proyecto/`](../../openspec/changes/hu-54-integrantes-del-proyecto/): [propuesta](../../openspec/changes/hu-54-integrantes-del-proyecto/proposal.md), specs delta de [`planning-poker`](../../openspec/changes/hu-54-integrantes-del-proyecto/specs/planning-poker/spec.md), [`esfuerzos`](../../openspec/changes/hu-54-integrantes-del-proyecto/specs/esfuerzos/spec.md) e [`interfaz-web`](../../openspec/changes/hu-54-integrantes-del-proyecto/specs/interfaz-web/spec.md), [diseño](../../openspec/changes/hu-54-integrantes-del-proyecto/design.md) y [tareas](../../openspec/changes/hu-54-integrantes-del-proyecto/tasks.md). Se archiva con `/opsx:archive` después del merge.
- **Aprobación:** Peñalbé Hernán (Agile Enabler) aprobó la propuesta el 16/11/2026, sin ajustes: 400, el voto y el esfuerzo se guardan con el nombre registrado, y no se integra en `main` antes de la presentación.
- **Origen:** en la prueba de la demo se pudo votar como "hernan", que no era integrante del proyecto. El votante (decisión 4 de la SDD de HU-21) y el integrante del esfuerzo (RN-26.1) eran texto libre porque E4 no estaba implementada.
- **Escenarios BDD:** [`features/HU-54-integrantes-del-proyecto.feature`](../../features/HU-54-integrantes-del-proyecto.feature): 11 escenarios a nivel de la API (7 de Planning Poker y 4 de esfuerzo), con sus pasos en [`hu54_features_test.go`](../../internal/httpapi/hu54_features_test.go).
- **Rama:** `feature/HU-54-integrantes-del-proyecto`, desde `main` `d0a2b98`.
- **Story Points:** a estimar en el Planning siguiente (HU-54 no estaba en el Planning del Sprint 5).

## Trazabilidad

**21 tests nuevos (funciones `Test`) y 11 escenarios BDD.** Se reparten así:

- 6 de `internal/member` en `canonical_name_test.go`;
- 6 de `internal/poker` en `vote_member_test.go`;
- 4 de `internal/effort` en `register_member_test.go`;
- 4 de `internal/httpapi` en `hu54_ui_test.go` (3 en el navegador, con los pasos de HU-53);
- 1 runner de godog, que ejecuta los 11 escenarios.

### Criterio → test → código

| Criterio | Escenarios BDD | Tests | Código |
|---|---|---|---|
| CA-54.1 «Integrante» de "Votar" y "Registrar esfuerzo" es un desplegable con los integrantes del proyecto actual; sin integrantes avisa "Registrá integrantes primero" y no envía | — (se prueba en el navegador) | [`TestHU54_IntegranteEsUnDesplegable`](../../internal/httpapi/hu54_ui_test.go#L29), [`TestHU54_DesplegablesConLosIntegrantesDelProyecto`](../../internal/httpapi/hu54_ui_test.go#L46), [`TestHU54_SinIntegrantesAvisa`](../../internal/httpapi/hu54_ui_test.go#L69), [`TestHU54_RegistrarEsfuerzoEligiendoDeLaLista`](../../internal/httpapi/hu54_ui_test.go#L89) | [`index.html`](../../internal/web/index.html) (los dos `<select name="integrante">`); [`app.js`](../../internal/web/app.js) (`sinIntegrantes` y los desplegables en `recargarListas`, y los avisos al votar y al registrar el esfuerzo) |
| CA-54.2 La API rechaza el voto de quien no es integrante del proyecto del ítem (400) y no lo guarda | "Vota un integrante del proyecto", "Vota alguien que no es integrante", "Integrante de otro proyecto", "Proyecto sin integrantes", "Se conserva el orden de errores al votar" | [`TestCanonicalName_NoIntegrante`](../../internal/member/canonical_name_test.go#L37), [`TestCanonicalName_IntegranteDeOtroProyecto`](../../internal/member/canonical_name_test.go#L52), [`TestCanonicalName_ProyectoSinIntegrantes`](../../internal/member/canonical_name_test.go#L62), [`TestVote_IntegranteDelProyecto`](../../internal/poker/vote_member_test.go#L68), [`TestVote_RechazaNoIntegranteSinGuardar`](../../internal/poker/vote_member_test.go#L110), [`TestVote_SesionNoAbiertaAntesQueIntegrante`](../../internal/poker/vote_member_test.go#L124), [`TestFeatures_HU54IntegrantesDelProyecto`](../../internal/httpapi/hu54_features_test.go#L18) | [`CanonicalName`](../../internal/member/canonical_name.go#L10); [`NewManager`](../../internal/poker/manager.go#L30) y [`Vote`](../../internal/poker/manager.go#L36) de `poker`; [`statusFor`](../../internal/httpapi/server.go#L360) (400) |
| CA-54.3 La API rechaza el esfuerzo de quien no es integrante del proyecto del ítem (400) y no lo guarda | "Registra horas un integrante del proyecto", "Registra horas alguien que no es integrante", "Se conserva el orden de errores al registrar horas" | [`TestRegister_IntegranteDelProyecto`](../../internal/effort/register_member_test.go#L52), [`TestRegister_RechazaNoIntegranteSinGuardar`](../../internal/effort/register_member_test.go#L83), [`TestRegister_ItemNoAsignadoAntesQueIntegrante`](../../internal/effort/register_member_test.go#L97), [`TestFeatures_HU54IntegrantesDelProyecto`](../../internal/httpapi/hu54_features_test.go#L18) | [`CanonicalName`](../../internal/member/canonical_name.go#L10); [`NewManager`](../../internal/effort/manager.go#L22) y [`Register`](../../internal/effort/register.go#L17) de `effort`; [`statusFor`](../../internal/httpapi/server.go#L360) (400) |
| CA-54.4 Los integrantes del proyecto votan y registran horas como antes, con el nombre registrado | "El nombre se escribe con otras mayúsculas y espacios", "El mismo integrante cambia su voto", "El nombre del esfuerzo se escribe con otras mayúsculas" | [`TestCanonicalName_DevuelveElNombreRegistrado`](../../internal/member/canonical_name_test.go#L12), [`TestCanonicalName_IgnoraMayusculasYEspacios`](../../internal/member/canonical_name_test.go#L24), [`TestVote_GuardaElNombreRegistrado`](../../internal/poker/vote_member_test.go#L83), [`TestVote_ReemplazaElVotoConOtrasMayusculas`](../../internal/poker/vote_member_test.go#L95), [`TestRegister_GuardaElNombreRegistrado`](../../internal/effort/register_member_test.go#L69), y los tests de HU-21 a HU-28, que siguen en verde | [`CanonicalName`](../../internal/member/canonical_name.go#L10); [`Vote`](../../internal/poker/manager.go#L36); [`Register`](../../internal/effort/register.go#L17) |

Condiciones de error sin criterio propio: [`TestCanonicalName_ErrorDelRepositorio`](../../internal/member/canonical_name_test.go#L71) y [`TestVote_ErrorAlConsultarIntegrantes`](../../internal/poker/vote_member_test.go#L145) verifican que un error al consultar los integrantes se devuelva tal cual, sin guardar nada.

## Tests existentes adaptados (D5)

`NewManager` de `poker` y de `effort` recibe el parámetro nuevo `members`. Los tests de las historias anteriores reciben `anyMember`, un falso que devuelve el nombre tal cual, así siguen verificando lo mismo sin registrar integrantes:

- **Llamadas adaptadas:** 99 llamadas a `NewManager`, en los tests de `poker`, `effort`, `metrics`, `report` e `httpapi`, incluido `newHandler`. El único cambio es el parámetro agregado.
- **Falsos nuevos:** `anyMember` está en `internal/poker/vote_member_test.go` e `internal/effort/register_member_test.go`, y en `members_any_test.go` de `httpapi` (dos paquetes de test), `metrics` y `report`.
- **Handler de HU-54:** `newHandlerValidatingMembers`, en `server_test.go`, se arma como `cmd/server`, con el `member.Manager` real. Lo usan los escenarios y los tests de interfaz de HU-54.
- **Helper de HU-53:** `registerEffort` (pasos de HU-53) ahora elige al integrante del desplegable y, si el integrante del escenario no está registrado, lo registra por la API y vuelve a elegir el proyecto y el sprint. Ni el `.feature` de HU-53 ni lo que verifica cambiaron.

## Fase RED (`408c1af`)

```
$ go test ./...
FAIL internal/member [build failed]  env.members.CanonicalName undefined; undefined: ErrNotProjectMember
FAIL internal/poker  [build failed]  too many arguments in call to NewManager; undefined: member.ErrNotProjectMember
FAIL internal/effort [build failed]  too many arguments in call to NewManager; undefined: member.ErrNotProjectMember
--- FAIL: TestFeatures_HU54IntegrantesDelProyecto
      11 scenarios (4 passed, 7 failed)
      código 200, se esperaba 400 (votó "hernan")
      código 201, se esperaba 400 (registró horas "hernan")
      votos revelados [sosa ricardo=8], se esperaba [Sosa Ricardo=8]
      el esfuerzo es de "GARCÍA SANTIAGO", se esperaba "García Santiago"
--- FAIL: TestHU54_IntegranteEsUnDesplegable
--- FAIL: TestHU54_DesplegablesConLosIntegrantesDelProyecto
--- FAIL: TestHU54_SinIntegrantesAvisa
--- FAIL: TestHU54_RegistrarEsfuerzoEligiendoDeLaLista
ok   cmd/server, backlog, dashboard, defect, metrics, project, report, sprint, trazabilidad, …
```

`member`, `poker` y `effort` no compilan porque `CanonicalName`, `ErrNotProjectMember` y el parámetro nuevo de `NewManager` no existen. En `httpapi` fallan solo los tests de HU-54, por aserción. Los 4 escenarios que pasan en RED verifican lo que ya funcionaba y no debía romperse: que un integrante vote y registre horas, y el orden de errores. Ningún test de otra historia falla.

## Fase GREEN (`2f81190`)

```
$ gofmt -l .
(sin salida)
$ go vet ./...
(sin salida)
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server
ok  	seguimiento-medicion-gpc/internal/backlog
ok  	seguimiento-medicion-gpc/internal/dashboard
ok  	seguimiento-medicion-gpc/internal/defect
ok  	seguimiento-medicion-gpc/internal/effort
ok  	seguimiento-medicion-gpc/internal/httpapi
ok  	seguimiento-medicion-gpc/internal/member
ok  	seguimiento-medicion-gpc/internal/metrics
ok  	seguimiento-medicion-gpc/internal/poker
ok  	seguimiento-medicion-gpc/internal/postgres
ok  	seguimiento-medicion-gpc/internal/postgrestest
ok  	seguimiento-medicion-gpc/internal/project
ok  	seguimiento-medicion-gpc/internal/report
ok  	seguimiento-medicion-gpc/internal/sprint
ok  	seguimiento-medicion-gpc/internal/trazabilidad
$ CGO_ENABLED=1 go test -race -count=1 ./internal/member/ ./internal/poker/ ./internal/effort/ ./cmd/server/
ok  	seguimiento-medicion-gpc/internal/member
ok  	seguimiento-medicion-gpc/internal/poker
ok  	seguimiento-medicion-gpc/internal/effort
ok  	seguimiento-medicion-gpc/cmd/server
```

Los 11 escenarios y los 21 tests nuevos pasan. Los links de 27 evidencias de otras historias se actualizaron con `go test ./internal/trazabilidad/ -run TestEnlaces -args -update` (solo números de línea).

## Refactor

No hubo: no se encontró una mejora clara. El falso `anyMember` se repite en cinco paquetes de test porque cada uno es un paquete distinto y el falso tiene 3 líneas; moverlo a un paquete compartido de producción no se justifica.

## Cobertura

```
$ go test -count=1 -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 25.0% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 81.3% of statements
ok  	seguimiento-medicion-gpc/internal/dashboard	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/defect	coverage: 76.4% of statements
ok  	seguimiento-medicion-gpc/internal/effort	coverage: 73.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 97.5% of statements
ok  	seguimiento-medicion-gpc/internal/member	coverage: 71.1% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	coverage: 74.1% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 74.9% of statements
ok  	seguimiento-medicion-gpc/internal/report	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	coverage: 75.1% of statements
$ go tool cover -func=cover.out
internal/effort/register.go:17:      Register        100.0%
internal/member/canonical_name.go:10: CanonicalName   100.0%
internal/poker/manager.go:36:        Vote             97.1%
total:                               (statements)     83.7%
```

Medido sin `TEST_DATABASE_URL` (los tests de PostgreSQL se omiten; los corre el CI). Comparado con `main` `d0a2b98` en el mismo entorno: `member` 66.9% → 71.1%, `poker` 73.8% → 74.1%, `effort` 72.8% → 73.2%, `httpapi` igual (97.5%), total 84.0% → 83.7%. El total baja 0,3 puntos sin cambios de cobertura en los paquetes que se tocan; se registra como medido, sin ajustarlo.

## Desviaciones

1. **Dos tests escritos en RED cambiaron en GREEN.** `TestHU54_IntegranteEsUnDesplegable` dejó de exigir `required` en el `<select>`, y `TestHU54_SinIntegrantesAvisa` ahora verifica el mensaje "Registrá integrantes primero". El `required` del navegador cortaba el envío antes de que la interfaz mostrara su aviso, y eso rompía dos tests de HU-53 (CA-53.5). D6 de `design.md` se actualizó.
2. **D5 cambió.** La propuesta decía que los tests de la API registrarían a los integrantes que usan. Se usó el falso también en `newHandler`, y un handler aparte, `newHandlerValidatingMembers`, para HU-54. Así ningún test de otra historia cambia lo que verifica. D5 de `design.md` se actualizó.
3. **Arreglo sin test propio:** los errores al votar se mostraban en la sección Backlog (`msg-backlog`, el valor por defecto de `boton`). Ahora se muestran en Planning Poker (`msg-poker`); si no, el 400 nuevo no se veía.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| Propuesta (OpenSpec) | 4e7a4c0 | `docs(HU-54): agregar la propuesta de OpenSpec` |
| RED | 408c1af | `test(HU-54): agregar tests (RED)` |
| GREEN | 2f81190 | `feat(HU-54): validar el integrante al votar y al registrar esfuerzo (GREEN)` |
| Refactor | — | No hubo (ver "Refactor") |
| Evidencia y cobertura | (propuesto) | `docs(HU-54): agregar evidencia TDD y cobertura` |

HU-54: Historia → Especificación → Criterios de Aceptación → BDD → Tests → Código Go
