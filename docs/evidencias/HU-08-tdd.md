# Evidencia TDD — HU-08 · Cambiar el estado de un ítem

> Issue del repositorio: #20

- **Responsable:** Sosa Ricardo
- **Rama:** `feature/HU-08-cambiar-estado-item`
- **Fecha:** 19/10/2026
- **Go:** go1.27.1 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` lo verifica el CI del Pull Request)
- **Especificación SDD:** [`docs/especificaciones/HU-08-cambiar-estado-item.md`](../especificaciones/HU-08-cambiar-estado-item.md)
- **Escenarios BDD:** [`features/HU-08-cambiar-estado-item.feature`](../../features/HU-08-cambiar-estado-item.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**15 funciones de test cubriendo 18 ejecuciones de escenario BDD (5 escenarios simples y 5 esquemas del escenario con 2, 2, 3, 4 y 2 ejemplos; 80 pasos)**: 14 tests unitarios en `internal/backlog/status_test.go` (que con sus 14 subtests hacen 28 ejecuciones) y 1 test que ejecuta los escenarios BDD con godog (`internal/backlog/status_features_test.go`, runner `TestFeatures_HU08CambiarEstado`). El conteo de funciones sale del comando:

```
$ $r = Select-String -Path internal\backlog\status_test.go, internal\backlog\status_features_test.go -Pattern '^func Test'; $r | ForEach-Object { $_.Filename + ': ' + $_.Line }; "TOTAL funciones de test: " + $r.Count
status_test.go: func TestManagerChangeStatus_PendienteAEnCursoConservaElRestoDeLosCampos(t *testing.T) {
status_test.go: func TestManagerChangeStatus_EnCursoATerminado(t *testing.T) {
status_test.go: func TestManagerChangeStatus_VolverAtrasYReabrirSonValidos(t *testing.T) {
status_test.go: func TestManagerChangeStatus_RechazaSaltosDeDosEstados(t *testing.T) {
status_test.go: func TestManagerChangeStatus_AceptaElEstadoSinMayusculasNiEspacios(t *testing.T) {
status_test.go: func TestManagerChangeStatus_RechazaEstadosInvalidos(t *testing.T) {
status_test.go: func TestManagerChangeStatus_EstadoInvalidoSobreItemInexistenteDaErrInvalidStatus(t *testing.T) {
status_test.go: func TestManagerChangeStatus_ItemInexistenteDaErrItemNotFound(t *testing.T) {
status_test.go: func TestManagerChangeStatus_SinItemsCreadosDaErrItemNotFound(t *testing.T) {
status_test.go: func TestManagerChangeStatus_MismoEstadoEsNoOpYNoLlamaAUpdate(t *testing.T) {
status_test.go: func TestManagerChangeStatus_CambioQuedaGuardadoYElResultadoEsCopia(t *testing.T) {
status_test.go: func TestManagerChangeStatus_CambiosSimultaneosAlMismoEstado(t *testing.T) {
status_test.go: func TestManagerChangeStatus_ErrorDelRepositorioAlBuscarSeDevuelve(t *testing.T) {
status_test.go: func TestManagerChangeStatus_ErrorDelRepositorioAlActualizarSeDevuelveYNoCambiaElEstado(t *testing.T) {
status_features_test.go: func TestFeatures_HU08CambiarEstado(t *testing.T) {
TOTAL funciones de test: 15
```

Las 28 ejecuciones unitarias salen de:

```
$ $r = go test -v -run "TestManagerChangeStatus" ./internal/backlog 2>&1; $r | Select-String -Pattern "^\s*--- PASS"; "TOTAL PASS (funciones + subtests): " + (($r | Select-String -Pattern "^\s*--- PASS").Count)
--- PASS: TestManagerChangeStatus_PendienteAEnCursoConservaElRestoDeLosCampos (0.00s)
--- PASS: TestManagerChangeStatus_EnCursoATerminado (0.00s)
--- PASS: TestManagerChangeStatus_VolverAtrasYReabrirSonValidos (0.00s)
    --- PASS: TestManagerChangeStatus_VolverAtrasYReabrirSonValidos/volver_atrás_de_EN_CURSO_a_PENDIENTE (0.00s)
    --- PASS: TestManagerChangeStatus_VolverAtrasYReabrirSonValidos/reabrir_de_TERMINADO_a_EN_CURSO (0.00s)
--- PASS: TestManagerChangeStatus_RechazaSaltosDeDosEstados (0.00s)
    --- PASS: TestManagerChangeStatus_RechazaSaltosDeDosEstados/PENDIENTE_a_TERMINADO (0.00s)
    --- PASS: TestManagerChangeStatus_RechazaSaltosDeDosEstados/TERMINADO_a_PENDIENTE (0.00s)
--- PASS: TestManagerChangeStatus_AceptaElEstadoSinMayusculasNiEspacios (0.00s)
    --- PASS: TestManagerChangeStatus_AceptaElEstadoSinMayusculasNiEspacios/minúsculas (0.00s)
    --- PASS: TestManagerChangeStatus_AceptaElEstadoSinMayusculasNiEspacios/mayúsculas_mezcladas_y_espacios (0.00s)
    --- PASS: TestManagerChangeStatus_AceptaElEstadoSinMayusculasNiEspacios/volver_atrás_normalizado (0.00s)
--- PASS: TestManagerChangeStatus_RechazaEstadosInvalidos (0.00s)
    --- PASS: TestManagerChangeStatus_RechazaEstadosInvalidos/estado_ (0.00s)
    --- PASS: TestManagerChangeStatus_RechazaEstadosInvalidos/estado____ (0.00s)
    --- PASS: TestManagerChangeStatus_RechazaEstadosInvalidos/estado_BLOQUEADO (0.00s)
    --- PASS: TestManagerChangeStatus_RechazaEstadosInvalidos/estado_en_curso (0.00s)
--- PASS: TestManagerChangeStatus_EstadoInvalidoSobreItemInexistenteDaErrInvalidStatus (0.00s)
--- PASS: TestManagerChangeStatus_ItemInexistenteDaErrItemNotFound (0.00s)
    --- PASS: TestManagerChangeStatus_ItemInexistenteDaErrItemNotFound/ID_999 (0.00s)
    --- PASS: TestManagerChangeStatus_ItemInexistenteDaErrItemNotFound/ID_0 (0.00s)
    --- PASS: TestManagerChangeStatus_ItemInexistenteDaErrItemNotFound/ID_-1 (0.00s)
--- PASS: TestManagerChangeStatus_SinItemsCreadosDaErrItemNotFound (0.00s)
--- PASS: TestManagerChangeStatus_MismoEstadoEsNoOpYNoLlamaAUpdate (0.00s)
--- PASS: TestManagerChangeStatus_CambioQuedaGuardadoYElResultadoEsCopia (0.00s)
--- PASS: TestManagerChangeStatus_CambiosSimultaneosAlMismoEstado (0.00s)
--- PASS: TestManagerChangeStatus_ErrorDelRepositorioAlBuscarSeDevuelve (0.00s)
--- PASS: TestManagerChangeStatus_ErrorDelRepositorioAlActualizarSeDevuelveYNoCambiaElEstado (0.00s)
TOTAL PASS (funciones + subtests): 28
```

La correspondencia no es 1 a 1: el orden de validación del estado tiene un test unitario propio sin escenario BDD, los tests de contrato de TEC-05 no son BDD y los esquemas multiplican sus ejemplos en ejecuciones.

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-08.1 PENDIENTE → EN_CURSO conserva el resto de los campos | Cambiar un ítem de PENDIENTE a EN_CURSO | `TestManagerChangeStatus_PendienteAEnCursoConservaElRestoDeLosCampos` | `status.go` (`ChangeStatus`, `allowedTransitions`) |
| CA-08.2 EN_CURSO → TERMINADO | Cambiar un ítem de EN_CURSO a TERMINADO | `TestManagerChangeStatus_EnCursoATerminado` | `status.go` (`ChangeStatus`) |
| CA-08.3 EN_CURSO → PENDIENTE y TERMINADO → EN_CURSO son válidas | Volver atrás y reabrir un ítem (esquema, 2 ejemplos) | `TestManagerChangeStatus_VolverAtrasYReabrirSonValidos` (2 subtests) | `status.go` (`allowedTransitions`, `isTransitionAllowed`) |
| CA-08.4 PENDIENTE → TERMINADO y TERMINADO → PENDIENTE dan `ErrInvalidTransition` | Rechazar transiciones que saltean un estado (esquema, 2 ejemplos) | `TestManagerChangeStatus_RechazaSaltosDeDosEstados` (2 subtests) | `status.go` (`ErrInvalidTransition`, `isTransitionAllowed`) |
| CA-08.5 Normalización y estados inválidos | Normalizar los estados válidos y rechazar los inválidos (esquema, 3 ejemplos) | `TestManagerChangeStatus_AceptaElEstadoSinMayusculasNiEspacios` (3 subtests), `TestManagerChangeStatus_RechazaEstadosInvalidos` (4 subtests), `TestManagerChangeStatus_EstadoInvalidoSobreItemInexistenteDaErrInvalidStatus` | `status.go` (`normalizeStatus`, `ErrInvalidStatus`) |
| CA-08.6 Ítem inexistente (999, 0, negativo, sin ítems) | Cambiar el estado de un ítem que no existe (esquema, 4 ejemplos) | `TestManagerChangeStatus_ItemInexistenteDaErrItemNotFound` (3 subtests), `TestManagerChangeStatus_SinItemsCreadosDaErrItemNotFound` | `get.go` y `memory_repository.go` (`GetByID`, `ErrItemNotFound`) |
| CA-08.7 Mismo estado: no-op sin llamar a `Update` | Cambiar al mismo estado no llama a Update | `TestManagerChangeStatus_MismoEstadoEsNoOpYNoLlamaAUpdate` | `status.go` (retorno temprano) |
| CA-08.8 Persistencia y copia por valor | El cambio queda guardado y el ítem devuelto es una copia | `TestManagerChangeStatus_CambioQuedaGuardadoYElResultadoEsCopia` | `status.go` (`Update`), `memory_repository.go` (copias) |
| CA-08.9 Cambios simultáneos al mismo estado | Varios cambios simultáneos al mismo estado | `TestManagerChangeStatus_CambiosSimultaneosAlMismoEstado` | `status.go` (mutex del `Manager`) |
| CA-08.10 Error del repositorio al buscar o al actualizar | El repositorio falla al buscar o al actualizar (esquema, 2 ejemplos) | `TestManagerChangeStatus_ErrorDelRepositorioAlBuscarSeDevuelve`, `TestManagerChangeStatus_ErrorDelRepositorioAlActualizarSeDevuelveYNoCambiaElEstado` | `status.go` (propagación de errores) |
| Contrato de TEC-05: el `Update` que cambia solo el estado conserva `SprintID` y `EstimatedHours` | — (test de contrato, no BDD) | `repository_contract_test.go` ("Update que cambia solo el estado conserva SprintID y EstimatedHours"), en memoria; contra PostgreSQL solo si `TEST_DATABASE_URL` está definida (en este entorno hizo SKIP) | `memory_repository.go` y `postgres_repository.go` de backlog (`Update`) |

## Fase RED

Tests y escenarios BDD escritos antes del código de producción (commit `7b9f7ce`). **El RED fue un error de compilación, no el fallo de una aserción**: `Manager.ChangeStatus`, `ErrInvalidStatus` y `ErrInvalidTransition` todavía no existían.

El commit RED se corrigió con `git commit --amend` antes de subirlo: `gofmt -l .` marcaba `internal/backlog/status_test.go` y una fila del esquema de CA-08.5 no probaba espacios en los extremos; esa fila se sacó del esquema (Gherkin recorta los espacios de las celdas) y los espacios en los extremos los cubre el test unitario "mayúsculas mezcladas y espacios" (entrada `"  termiNADO  "`), que ya existía. La salida siguiente se reprodujo sobre ese commit ya corregido (`git worktree add ... 7b9f7ce`):

```
$ go test ./...
# seguimiento-medicion-gpc/internal/backlog [seguimiento-medicion-gpc/internal/backlog.test]
internal\backlog\status_features_test.go:134:31: s.manager.ChangeStatus undefined (type *Manager has no field or method ChangeStatus)
internal\backlog\status_features_test.go:146:23: undefined: ErrInvalidTransition
internal\backlog\status_features_test.go:153:23: undefined: ErrInvalidStatus
internal\backlog\status_features_test.go:260:39: s.manager.ChangeStatus undefined (type *Manager has no field or method ChangeStatus)
internal\backlog\status_test.go:55:26: manager.ChangeStatus undefined (type *Manager has no field or method ChangeStatus)
internal\backlog\status_test.go:74:26: manager.ChangeStatus undefined (type *Manager has no field or method ChangeStatus)
internal\backlog\status_test.go:105:28: manager.ChangeStatus undefined (type *Manager has no field or method ChangeStatus)
internal\backlog\status_test.go:135:22: manager.ChangeStatus undefined (type *Manager has no field or method ChangeStatus)
internal\backlog\status_test.go:137:23: undefined: ErrInvalidTransition
internal\backlog\status_test.go:164:28: manager.ChangeStatus undefined (type *Manager has no field or method ChangeStatus)
internal\backlog\status_test.go:164:28: too many errors
ok  	seguimiento-medicion-gpc/cmd/server	4.280s
FAIL	seguimiento-medicion-gpc/internal/backlog [build failed]
ok  	seguimiento-medicion-gpc/internal/httpapi	2.967s
ok  	seguimiento-medicion-gpc/internal/postgres	3.653s
ok  	seguimiento-medicion-gpc/internal/postgrestest	3.793s
ok  	seguimiento-medicion-gpc/internal/project	5.292s
ok  	seguimiento-medicion-gpc/internal/sprint	5.399s
FAIL
```

Todos los errores son la funcionalidad nueva de HU-08. No hay fallos preexistentes: `internal/backlog` no compila hasta que existan `ChangeStatus` y los dos errores, así que en el RED tampoco corren los tests de las historias anteriores; el resto de los paquetes queda en verde.

## Fase GREEN

Código mínimo de producción (commit `35305e8`):

- `internal/backlog/status.go` (nuevo): `ErrInvalidStatus`, `ErrInvalidTransition`, `allowedTransitions`, `Manager.ChangeStatus`, `normalizeStatus` e `isTransitionAllowed`. El estado destino se normaliza y se valida antes de consultar el repositorio; después, con el mutex, se verifica la existencia con `GetByID`, el mismo estado (no-op sin `Update`), la transición y se escribe con `Update`.
- `internal/backlog/item.go`: el comentario de `Status` remite a `status.go` (antes decía "Las transiciones entre estados las define HU-08").

No se modificó ningún test en el GREEN.

```
$ gofmt -l .
(sin salida)
```

```
$ go vet ./...
(sin salida)
```

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/sprint	(cached)
```

## Refactor

No se hizo refactor: `status.go` es chico, sigue los patrones existentes (`Create`/`Update` de HU-04 y HU-07 para la normalización y el mutex) y no hay duplicación ni ramas sin test que justifiquen un cambio. No se inventó un commit de refactor.

## Desviaciones

No hubo desviaciones de código. Hubo dos ajustes:

- En el RED, antes del push: `gofmt -l .` marcaba `status_test.go` y una fila del esquema de CA-08.5 no probaba espacios en los extremos; ambas correcciones entraron con `--amend` en el commit RED, detallado arriba.
- En la SDD, después de su aprobación del 18/10: se agregó el sprint asignado (`SprintID`, de HU-11) a lo que no cambia (RN-08.5, CA-08.1 y casos límite) y la restricción "Sin cambios de esquema", con un caso nuevo en el contrato de repositorio de TEC-05. El Agile Enabler aprobó el cambio el 20/10 en la revisión del Pull Request #99; queda registrado como decisión 8 de la SDD.

## Pruebas de mutación

Una por vez, rompiendo `internal/backlog/status.go` a propósito, corriendo `go test ./internal/backlog/...` y restaurando con `git checkout -- internal/backlog/status.go` (ninguna mutación forma parte de un commit):

1. **Permitir PENDIENTE → TERMINADO** (se agregó `StatusFinished` a `allowedTransitions[StatusPending]`): **matada**. Fallan `TestManagerChangeStatus_RechazaSaltosDeDosEstados/PENDIENTE_a_TERMINADO` (`error = <nil>, se esperaba ErrInvalidTransition`) y el esquema BDD "Rechazar transiciones que saltean un estado" (`18 scenarios: 17 passed, 1 failed`; `80 steps: 78 passed, 1 failed, 1 skipped`).
2. **Hacer que el mismo estado llame a `Update`** (se reemplazó el retorno temprano por `return m.repo.Update(current)`): **matada**. Fallan `TestManagerChangeStatus_MismoEstadoEsNoOpYNoLlamaAUpdate` y el escenario BDD "Cambiar al mismo estado no llama a Update" (`el cambio de estado falló: falla del repositorio`).
3. **No normalizar el estado destino** (se quitó `ToUpper`/`TrimSpace`): **matada**. Fallan los 3 subtests de `TestManagerChangeStatus_AceptaElEstadoSinMayusculasNiEspacios` y 2 escenarios BDD (`18 scenarios: 16 passed, 2 failed`; `80 steps: 76 passed, 2 failed, 2 skipped`).
4. **Consultar el repositorio antes de validar el estado** (se movió `normalizeStatus` después de `GetByID`): **matada**. Falla `TestManagerChangeStatus_EstadoInvalidoSobreItemInexistenteDaErrInvalidStatus` (`error = no existe un ítem del backlog con ese id, se esperaba ErrInvalidStatus`). Los 18 escenarios BDD siguen pasando, así que la mutación la detecta el test unitario del orden de validación.
5. **Sacar el mutex de `ChangeStatus`**: **sobrevive localmente**. Sin `-race`, los tests siguen pasando en este entorno (`ok seguimiento-medicion-gpc/internal/backlog 5.747s`): la carrera no se manifiesta de forma determinista. La mutación **la verifica el CI con `-race`** (CA-08.9), porque el entorno local no tiene cgo. Con el mutex restaurado, todo vuelve a verde.

## Validación final

```
$ gofmt -l .
(sin salida)
```

```
$ go vet ./...
(sin salida)
```

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/sprint	(cached)
```

### Cobertura

```
$ go test -count=1 -timeout 120s -cover ./internal/backlog
ok  	seguimiento-medicion-gpc/internal/backlog	6.558s	coverage: 80.9% of statements
```

```
$ go test -count=1 "-coverprofile=cover.out" ./...
ok  	seguimiento-medicion-gpc/cmd/server	2.869s	coverage: 39.0% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	3.647s	coverage: 80.9% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.816s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	2.410s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	2.347s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	3.459s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	3.423s	coverage: 69.5% of statements

$ go tool cover "-func=cover.out" | Select-String "status.go"
seguimiento-medicion-gpc/internal/backlog/status.go:29:			ChangeStatus		100.0%
seguimiento-medicion-gpc/internal/backlog/status.go:59:			normalizeStatus		100.0%
seguimiento-medicion-gpc/internal/backlog/status.go:70:			isTransitionAllowed	100.0%

$ go tool cover "-func=cover.out" | Select-Object -Last 1
total:									(statements)		73.4%
```

Las tres funciones de `status.go` quedan al 100.0%: `ChangeStatus` 100.0%, `normalizeStatus` 100.0% e `isTransitionAllowed` 100.0%. Sin base (como en el CI), `internal/backlog` queda en 80.9% y el total del perfil en 73.4%. La parte de PostgreSQL no se ejecuta sin `TEST_DATABASE_URL`, así que su cobertura se mide localmente con la base levantada (ver TEC-05). El archivo `cover.out` no se versiona.

### go test -race (CA-08.9)

En este entorno (Windows, sin compilador C) `-race` no está disponible. La verificación con `-race` la corre el CI de TEC-02 en cada Pull Request hacia main (`ci.yml`, paso "Correr los tests (go test -race ./...)"), igual que en las historias anteriores. Los tests de concurrencia (50 cambios simultáneos al mismo estado, en el test unitario y en el escenario BDD) corren y pasan también sin `-race` en el entorno local.

### Ejecución de los escenarios BDD

```
$ go test -v -run "TestFeatures_HU08CambiarEstado" ./internal/backlog
18 scenarios (18 passed)
80 steps (80 passed)
--- PASS: TestFeatures_HU08CambiarEstado (0.03s)
```

Los 18 escenarios (5 simples y 5 esquemas con 2, 2, 3, 4 y 2 ejemplos) pasan con los 80 pasos, y con `Strict: true` godog habría fallado si algún paso del `.feature` quedara sin expresión registrada: no hay pasos undefined.

### Tests de contrato (CA-TEC05.3)

`TEST_DATABASE_URL` no está definida en este entorno, así que **el contrato contra PostgreSQL hizo SKIP** y no corrió contra una base:

```
$ echo $env:TEST_DATABASE_URL
(vacía: no definida)

$ go test -v -run TestPostgresRepository_Contrato ./internal/backlog
postgres_repository_test.go:39: TEST_DATABASE_URL no está definida: se omiten las pruebas de PostgreSQL
--- SKIP: TestPostgresRepository_Contrato (0.00s)
PASS
```

El contrato de `MemoryRepository` sí corrió e incluyó el caso nuevo de HU-08:

```
$ go test -v -run TestMemoryRepository_Contrato ./internal/backlog
--- PASS: TestMemoryRepository_Contrato (0.00s)
    --- PASS: TestMemoryRepository_Contrato/Update_que_cambia_solo_el_estado_conserva_SprintID_y_EstimatedHours (0.00s)
PASS
```

(Se lista solo el caso nuevo de HU-08; el resto de los casos del contrato de memoria sigue pasando igual. Nota: el patrón `Contract` no coincide con el nombre real del test, que usa `Contrato`; con `TestRepositoryContract|Contract` no se ejecuta ningún test. Por eso se usaron los nombres reales. La verificación del contrato contra PostgreSQL queda para cuando se levante la base con `TEST_DATABASE_URL`.)

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | `7b9f7ce` | `test(HU-08): agregar tests y escenarios BDD (RED)` |
| GREEN | `35305e8` | `feat(HU-08): implementar el cambio de estado de un ítem (GREEN)` |

No hubo commit de refactor porque no se hizo refactor. El commit de cierre (`Closes #20`) no se lista en esta tabla porque un commit no puede contener su propio hash: queda registrado en el Pull Request y en el Issue.
