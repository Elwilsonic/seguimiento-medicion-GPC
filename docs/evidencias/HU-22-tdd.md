# Evidencia TDD · HU-22

> Issue del repositorio: #40

## Historia

Como Product Builder, quiero votar mi estimación sin que los demás la vean, para no influirnos entre nosotros.

## Resultado

HU-22 quedó implementada en la rama `feature/HU-22-votar-oculto`. La API responde siempre `200 OK` para votos nuevos y reemplazos, y devuelve `votos: []` para conservar el ocultamiento.

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test | Código |
|---|---|---|
| CA-22.1 — votar en una sesión abierta con una carta válida | “Registrar un voto válido sin mostrarlo”; [`TestVote_VotoValidoPermaneceOculto`](../../internal/poker/vote_test.go#L25); [`TestVotarPoker_VotoNuevoYReemplazoDevuelven200SinVotos`](../../internal/httpapi/poker_test.go#L143) | [`manager.go`](../../internal/poker/manager.go) ([`Vote`](../../internal/poker/manager.go#L29), [`validCard`](../../internal/poker/manager.go#L159)); [`memory_repository.go`](../../internal/poker/memory_repository.go) ([`GetByID`](../../internal/poker/memory_repository.go#L37), [`Update`](../../internal/poker/memory_repository.go#L47)); [`poker.go`](../../internal/httpapi/poker.go) ([`votePokerSession`](../../internal/httpapi/poker.go#L84)) |
| CA-22.2 — votos ocultos, incluido el propio | “Registrar un voto válido sin mostrarlo”; “El voto reemplazado mantiene los votos ocultos”; [`TestVote_VotoValidoPermaneceOculto`](../../internal/poker/vote_test.go#L25); tests HTTP de voto | [`manager.go`](../../internal/poker/manager.go) ([`Vote`](../../internal/poker/manager.go#L29), [`validCard`](../../internal/poker/manager.go#L159)); [`memory_repository.go`](../../internal/poker/memory_repository.go) ([`GetByID`](../../internal/poker/memory_repository.go#L37), [`Update`](../../internal/poker/memory_repository.go#L47)) |
| CA-22.3 — reemplazar el voto sin duplicarlo | “Reemplazar el voto del mismo integrante”; [`TestVote_ReemplazaVotoYGuardaNombreRecortado`](../../internal/poker/vote_test.go#L48) | [`manager.go`](../../internal/poker/manager.go) ([`Vote`](../../internal/poker/manager.go#L29), [`validCard`](../../internal/poker/manager.go#L159)); [`memory_repository.go`](../../internal/poker/memory_repository.go) ([`GetByID`](../../internal/poker/memory_repository.go#L37), [`Update`](../../internal/poker/memory_repository.go#L47)) |
| CA-22.4 — rechazar votante vacío o carta inválida | “Rechazar una carta fuera de la escala”; “Rechazar un votante vacío”; [`TestVote_CartaInvalidaDevuelveError`](../../internal/poker/vote_test.go#L85); [`TestVote_VotanteVacioDevuelveError`](../../internal/poker/vote_test.go#L73); [`TestVotarPoker_ErroresDevuelvenCodigoCorrespondiente`](../../internal/httpapi/poker_test.go#L165) | [`manager.go`](../../internal/poker/manager.go) ([`Vote`](../../internal/poker/manager.go#L29), [`validCard`](../../internal/poker/manager.go#L159)); [`memory_repository.go`](../../internal/poker/memory_repository.go) ([`GetByID`](../../internal/poker/memory_repository.go#L37)); [`poker.go`](../../internal/httpapi/poker.go) ([`votePokerSession`](../../internal/httpapi/poker.go#L84)); [`server.go`](../../internal/httpapi/server.go) ([`statusFor`](../../internal/httpapi/server.go#L164)) |
| CA-22.5 — rechazar sesión inexistente/no abierta y fallas de guardado | “Rechazar el voto de una sesión inexistente”; “Rechazar el voto de una sesión no abierta”; [`TestVote_SesionInexistenteDevuelveError`](../../internal/poker/vote_test.go#L97); [`TestVote_SesionNoAbiertaDevuelveError`](../../internal/poker/vote_test.go#L109); [`TestVote_ErrorAlGuardarPropagaErrorYSesionNoSeModifica`](../../internal/poker/vote_test.go#L137); tests HTTP de 409 y 500; contrato de `GetByID` y `Update` | [`manager.go`](../../internal/poker/manager.go) ([`Vote`](../../internal/poker/manager.go#L29), [`validCard`](../../internal/poker/manager.go#L159)); [`memory_repository.go`](../../internal/poker/memory_repository.go) ([`GetByID`](../../internal/poker/memory_repository.go#L37), [`Update`](../../internal/poker/memory_repository.go#L47)) |
| CA-22.6 — conservar votos simultáneos | “Registrar votos simultáneos de integrantes distintos”; [`TestVote_ConcurrenteConservaVotosDeIntegrantesDistintos`](../../internal/poker/vote_test.go#L165) | [`manager.go`](../../internal/poker/manager.go) ([`Vote`](../../internal/poker/manager.go#L29), [`validCard`](../../internal/poker/manager.go#L159)); [`memory_repository.go`](../../internal/poker/memory_repository.go) ([`GetByID`](../../internal/poker/memory_repository.go#L37), [`Update`](../../internal/poker/memory_repository.go#L47)) |
| CA-22.7 — votar desde API e interfaz, siempre con 200 | “El voto reemplazado mantiene los votos ocultos” verifica el ocultamiento en el Manager; [`TestVotarPoker_VotoNuevoYReemplazoDevuelven200SinVotos`](../../internal/httpapi/poker_test.go#L143) verifica `200 OK`; [`TestInterfaz_FormularioDeVoto`](../../internal/httpapi/web_test.go#L166) | [`manager.go`](../../internal/poker/manager.go) ([`Vote`](../../internal/poker/manager.go#L29), [`validCard`](../../internal/poker/manager.go#L159)); [`memory_repository.go`](../../internal/poker/memory_repository.go) ([`GetByID`](../../internal/poker/memory_repository.go#L37), [`Update`](../../internal/poker/memory_repository.go#L47)); [`poker.go`](../../internal/httpapi/poker.go) ([`votePokerSession`](../../internal/httpapi/poker.go#L84)); [`index.html`](../../internal/web/index.html); [`app.js`](../../internal/web/app.js) ([`mostrarPoker`](../../internal/web/app.js#L311)) |
| CA-22.8 — mismo comportamiento en memoria y PostgreSQL | Contrato de `GetByID` y `Update` ejecutado por [`TestMemoryRepository_Contrato`](../../internal/poker/repository_contract_test.go#L297) y [`TestPostgresRepository_Contrato`](../../internal/poker/postgres_repository_test.go#L39); PostgreSQL se omite sin `TEST_DATABASE_URL` | [`memory_repository.go`](../../internal/poker/memory_repository.go) ([`GetByID`](../../internal/poker/memory_repository.go#L37), [`Update`](../../internal/poker/memory_repository.go#L47)); [`postgres_repository.go`](../../internal/poker/postgres_repository.go) ([`GetByID`](../../internal/poker/postgres_repository.go#L68), [`Update`](../../internal/poker/postgres_repository.go#L82)) |

La validación contiene **8 escenarios BDD** y **16 casos de prueba específicos**: 8 unitarios de votación, 3 contractuales de repositorio, 4 HTTP y 1 de interfaz web. La correspondencia no es 1 a 1.

## BDD

Archivo: `features/HU-22-votar-oculto.feature`.

Resultado de `go test ./...`:

```text
8 scenarios passed
49 steps passed
```

## Fase RED

Commit:

| Fase | Commit | Mensaje |
|---|---|---|
| RED | `0dddb7a` | `test(HU-22): agregar tests (RED)` |

El RED falló específicamente porque no existían `Vote`, `GetByID`, `Update` ni sus errores. `go vet ./...` se ejecutó sin errores antes del commit RED.

## Fase GREEN

Commit:

| Fase | Commit | Mensaje |
|---|---|---|
| GREEN | `6e92335` | `feat(HU-22): implementar votación oculta (GREEN)` |
| Refactor | — | No aplica |

## Validación final

### `gofmt`

```text
gofmt -w internal/poker/*.go internal/httpapi/*.go
--- gofmt: OK ---
```

### `go test ./...`

```text
ok  seguimiento-medicion-gpc/cmd/server
ok  seguimiento-medicion-gpc/internal/backlog
ok  seguimiento-medicion-gpc/internal/defect
ok  seguimiento-medicion-gpc/internal/effort
ok  seguimiento-medicion-gpc/internal/httpapi
ok  seguimiento-medicion-gpc/internal/metrics
ok  seguimiento-medicion-gpc/internal/poker
ok  seguimiento-medicion-gpc/internal/postgres
ok  seguimiento-medicion-gpc/internal/postgrestest
ok  seguimiento-medicion-gpc/internal/project
ok  seguimiento-medicion-gpc/internal/sprint
?   seguimiento-medicion-gpc/internal/web [no test files]
--- go test exit: 0 ---
```

### `go test -cover ./...`

```text
cmd/server 31.5%
internal/backlog 83.3%
internal/defect 76.6%
internal/effort 76.2%
internal/httpapi 98.0%
internal/metrics 100.0%
internal/poker 60.2%
internal/postgres 83.3%
internal/postgrestest 36.8%
internal/project 67.4%
internal/sprint 76.3%
internal/web 0.0% [no test files]
```

Las pruebas de integración PostgreSQL se omiten porque no está definida la variable `TEST_DATABASE_URL`. El escenario BDD de reemplazo verifica el ocultamiento en el `Manager`; el código HTTP `200 OK` se verifica con `TestVotarPoker_VotoNuevoYReemplazoDevuelven200SinVotos`.

## Commits

Los commits RED y GREEN referencian el Issue #40 con `Refs #40`. El commit documental de esta evidencia y cobertura queda propuesto hasta autorización.

HU-22: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
