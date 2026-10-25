# Evidencia TDD · HU-22

> Issue del repositorio: #40

## Historia

Como Product Builder, quiero votar mi estimación sin que los demás la vean, para no influirnos entre nosotros.

## Resultado

HU-22 quedó implementada en la rama `feature/HU-22-votar-oculto`. La API responde siempre `200 OK` para votos nuevos y reemplazos, y devuelve `votos: []` para conservar el ocultamiento.

## Trazabilidad

| Criterio de aceptación | Escenario BDD / test |
|---|---|
| CA-22.1 — votar en una sesión abierta con una carta válida | “Registrar un voto válido sin mostrarlo”; `TestVote_VotoValidoPermaneceOculto`; `TestVotarPoker_VotoNuevoYReemplazoDevuelven200SinVotos` |
| CA-22.2 — votos ocultos, incluido el propio | “Registrar un voto válido sin mostrarlo”; “El voto reemplazado mantiene los votos ocultos”; `TestVote_VotoValidoPermaneceOculto`; tests HTTP de voto |
| CA-22.3 — reemplazar el voto sin duplicarlo | “Reemplazar el voto del mismo integrante”; `TestVote_ReemplazaVotoYGuardaNombreRecortado` |
| CA-22.4 — rechazar votante vacío o carta inválida | “Rechazar una carta fuera de la escala”; “Rechazar un votante vacío”; `TestVote_CartaInvalidaDevuelveError`; `TestVote_VotanteVacioDevuelveError`; `TestVotarPoker_ErroresDevuelvenCodigoCorrespondiente` |
| CA-22.5 — rechazar sesión inexistente/no abierta y fallas de guardado | “Rechazar el voto de una sesión inexistente”; “Rechazar el voto de una sesión no abierta”; `TestVote_SesionInexistenteDevuelveError`; `TestVote_SesionNoAbiertaDevuelveError`; `TestVote_ErrorAlGuardarPropagaErrorYSesionNoSeModifica`; tests HTTP de 409 y 500; contrato de `GetByID` y `Update` |
| CA-22.6 — conservar votos simultáneos | “Registrar votos simultáneos de integrantes distintos”; `TestVote_ConcurrenteConservaVotosDeIntegrantesDistintos` |
| CA-22.7 — votar desde API e interfaz, siempre con 200 | “El voto reemplazado mantiene los votos ocultos” verifica el ocultamiento en el Manager; `TestVotarPoker_VotoNuevoYReemplazoDevuelven200SinVotos` verifica `200 OK`; `TestInterfaz_FormularioDeVoto` |
| CA-22.8 — mismo comportamiento en memoria y PostgreSQL | Contrato de `GetByID` y `Update` ejecutado por `TestMemoryRepository_Contrato` y `TestPostgresRepository_Contrato`; PostgreSQL se omite sin `TEST_DATABASE_URL` |

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
