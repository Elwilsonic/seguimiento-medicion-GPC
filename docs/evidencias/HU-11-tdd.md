# Evidencia TDD — HU-11 · Asignar historias del backlog a un sprint

> Issue del repositorio: #25

- **Responsable:** Buttini Cristobal
- **Rama:** `feature/HU-11-asignar-item-sprint`
- **Fecha:** 18/10/2026
- **Go:** go1.27.1 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` lo verifica el CI del Pull Request)
- **Especificación SDD:** [`docs/especificaciones/HU-11-asignar-item-sprint.md`](../especificaciones/HU-11-asignar-item-sprint.md)
- **Escenarios BDD:** [`features/HU-11-asignar-item-sprint.feature`](../../features/HU-11-asignar-item-sprint.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**26 funciones de test cubriendo 20 ejecuciones de escenario BDD (12 escenarios y 3 esquemas del escenario con 2, 3 y 3 ejemplos; 119 pasos)**: 25 tests unitarios (`internal/sprint/assign_test.go` y `internal/backlog/assign_test.go`, que con sus 13 subtests hacen 38 ejecuciones) y 1 test que ejecuta los escenarios BDD con godog (`internal/sprint/assign_features_test.go`, runner `TestFeatures_HU11AsignarItemSprint`, en el paquete `sprint` porque la historia es de E3 y orquesta los dos paquetes). El conteo sale del comando:

```
$ findstr /R "^func.Test" internal\sprint\assign_test.go internal\sprint\assign_features_test.go internal\backlog\assign_test.go
internal\sprint\assign_test.go:func TestAssignItem_ItemDelProyectoSeAsignaAlSprintAbierto(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_DosItemsDistintosAlMismoSprint(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_ItemYaAsignadoAlMismoSprintDaErrItemAlreadyAssigned(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_ItemYaAsignadoAOtroSprintDaErrItemAlreadyAssigned(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_SprintCerradoDaErrSprintClosed(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_ItemDeOtroProyectoDaErrItemFromAnotherProject(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_SprintInexistenteDaErrSprintNotFound(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_SinSprintsCreadosDaErrSprintNotFound(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_ItemInexistenteDaErrItemNotFound(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_SinItemsCreadosDaErrItemNotFound(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_ItemEnCursoOTerminadoSeAsigna(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_NoCambiaNingunOtroCampoNiElSprint(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_ModificarElItemDevueltoNoAlteraElGuardado(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_AsignacionesSimultaneasDelMismoItemDejanUnaSola(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_AsignacionesSimultaneasDeItemsDistintosTodasExitosas(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_ErrorDelRepositorioDeSprintsSeDevuelve(t *testing.T) {
internal\sprint\assign_test.go:func TestAssignItem_ErrorDelRepositorioDeItemsSeDevuelve(t *testing.T) {
internal\sprint\assign_features_test.go:func TestFeatures_HU11AsignarItemSprint(t *testing.T) {
internal\backlog\assign_test.go:func TestAssignToSprint_ItemSinSprintQuedaAsignado(t *testing.T) {
internal\backlog\assign_test.go:func TestAssignToSprint_ItemYaAsignadoDaErrItemAlreadyAssigned(t *testing.T) {
internal\backlog\assign_test.go:func TestAssignToSprint_ItemInexistenteDaErrItemNotFound(t *testing.T) {
internal\backlog\assign_test.go:func TestAssignToSprint_NoCambiaNingunOtroCampo(t *testing.T) {
internal\backlog\assign_test.go:func TestAssignToSprint_ModificarElItemDevueltoNoAlteraElGuardado(t *testing.T) {
internal\backlog\assign_test.go:func TestAssignToSprint_UpdateConservaElSprintID(t *testing.T) {
internal\backlog\assign_test.go:func TestAssignToSprint_CreateDejaElSprintIDEnCero(t *testing.T) {
internal\backlog\assign_test.go:func TestAssignToSprint_ErrorDelRepositorioAlConsultarSeDevuelve(t *testing.T) {
internal\backlog\assign_test.go:func TestAssignToSprint_ErrorDelRepositorioAlEscribirSeDevuelve(t *testing.T) {
```

Los 13 subtests: `TestAssignItem_ItemEnCursoOTerminadoSeAsigna` (EN_CURSO y TERMINADO), `TestAssignItem_SprintInexistenteDaErrSprintNotFound` (999, 0 y -1), `TestAssignItem_ItemInexistenteDaErrItemNotFound` (999, 0 y -1), `TestAssignToSprint_ItemYaAsignadoDaErrItemAlreadyAssigned` (al mismo sprint y a otro) y `TestAssignToSprint_ItemInexistenteDaErrItemNotFound` (999, 0 y -1).

La correspondencia no es 1 a 1: los errores del repositorio (CA-11.8) y la conservación de `SprintID` en `Create` (RN-11.5) tienen tests unitarios propios sin escenario BDD, los tests de contrato de TEC-05 (3 casos nuevos: `GetByID` de sprints con sus dos subcasos y el `SprintID` del ítem) no son BDD, y los esquemas multiplican sus ejemplos en ejecuciones.

Además de los archivos nuevos, el commit RED actualiza el harness que forzó la SDD: `newTestManager` y el `Before` de godog de HU-10 llaman al constructor nuevo de tres parámetros, y `failingRepository` de sprint gana el método `GetByID`.

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-11.1 Un ítem existente se asigna a un sprint ABIERTO del mismo proyecto y solo cambia su SprintID | Asignar un ítem a un sprint abierto del mismo proyecto; Asignar dos ítems distintos al mismo sprint; Un ítem terminado se asigna igual (RN-11.6) | `TestAssignItem_ItemDelProyectoSeAsignaAlSprintAbierto`, `TestAssignItem_DosItemsDistintosAlMismoSprint`, `TestAssignItem_ItemEnCursoOTerminadoSeAsigna`, `TestAssignToSprint_ItemSinSprintQuedaAsignado` | `assign.go` de sprint (`AssignItem`), `assign.go` de backlog (`AssignToSprint`), `item.go` (`SprintID`) |
| CA-11.2 Un ítem ya asignado no se puede asignar de nuevo (ni al mismo sprint) | Asignar un ítem ya asignado (esquema, 2 ejemplos) | `TestAssignItem_ItemYaAsignadoAlMismoSprintDaErrItemAlreadyAssigned`, `TestAssignItem_ItemYaAsignadoAOtroSprintDaErrItemAlreadyAssigned`, `TestAssignToSprint_ItemYaAsignadoDaErrItemAlreadyAssigned` | `assign.go` de backlog (RN-11.3, bajo el mutex) |
| CA-11.3 Ni sprint CERRADO ni ítem de otro proyecto | Asignar a un sprint cerrado; Asignar un ítem de otro proyecto | `TestAssignItem_SprintCerradoDaErrSprintClosed`, `TestAssignItem_ItemDeOtroProyectoDaErrItemFromAnotherProject` | `assign.go` de sprint (RN-11.1 y RN-11.2), `sprint.go` (`ErrSprintClosed`, `ErrItemFromAnotherProject`) |
| CA-11.4 Sprint inexistente e ítem inexistente (0 y negativos incluidos) | Asignar a un sprint inexistente (esquema, 3 ejemplos); Asignar cuando no hay sprints creados; Asignar un ítem inexistente (esquema, 3 ejemplos); Asignar cuando no hay ítems creados | `TestAssignItem_SprintInexistenteDaErrSprintNotFound` (3 subtests), `TestAssignItem_SinSprintsCreadosDaErrSprintNotFound`, `TestAssignItem_ItemInexistenteDaErrItemNotFound` (3 subtests), `TestAssignItem_SinItemsCreadosDaErrItemNotFound`, `TestAssignToSprint_ItemInexistenteDaErrItemNotFound` (3 subtests) | `assign.go` de sprint, `repository.go` (+`GetByID`), `memory_repository.go` y `postgres_repository.go` (`GetByID` con `ErrSprintNotFound`) |
| CA-11.5 Asignar no cambia ningún otro campo; el `Update` de HU-07 conserva el SprintID | Asignar no cambia ningún otro campo del ítem; Modificar el ítem con el Update de HU-07 conserva la asignación | `TestAssignItem_NoCambiaNingunOtroCampoNiElSprint`, `TestAssignToSprint_NoCambiaNingunOtroCampo`, `TestAssignToSprint_UpdateConservaElSprintID`, `TestAssignToSprint_CreateDejaElSprintIDEnCero` | `update.go` (conserva `SprintID`), `create.go` (`SprintID: 0` explícito) |
| CA-11.6 De varias asignaciones simultáneas del mismo ítem, exactamente una gana | Asignaciones simultáneas del mismo ítem a dos sprints; Asignaciones simultáneas de ítems distintos al mismo sprint | `TestAssignItem_AsignacionesSimultaneasDelMismoItemDejanUnaSola`, `TestAssignItem_AsignacionesSimultaneasDeItemsDistintosTodasExitosas` | `assign.go` de backlog (RN-11.3 y escritura bajo el mutex), `manager.go` (mutex de sprint) |
| CA-11.7 Modificar el ítem devuelto no altera el guardado | Modificar el ítem devuelto no altera el ítem guardado | `TestAssignItem_ModificarElItemDevueltoNoAlteraElGuardado`, `TestAssignToSprint_ModificarElItemDevueltoNoAlteraElGuardado` | `assign.go` de backlog (copias por valor, RN-11.7) |
| CA-11.8 El error del repositorio se devuelve y no se asigna nada | — (no tiene escenario BDD) | `TestAssignItem_ErrorDelRepositorioDeSprintsSeDevuelve`, `TestAssignItem_ErrorDelRepositorioDeItemsSeDevuelve`, `TestAssignToSprint_ErrorDelRepositorioAlConsultarSeDevuelve`, `TestAssignToSprint_ErrorDelRepositorioAlEscribirSeDevuelve` | `assign.go` de sprint y de backlog (propagación), `create_test.go` de sprint (`failingRepository` con `GetByID`), `assign_test.go` de sprint (`failingItemRepository`) |
| Contratos de TEC-05 (CA-TEC05.3): `GetByID` de sprints y `SprintID` del ítem, igual en memoria y PostgreSQL | — (tests de contrato, no BDD) | `repository_contract_test.go` de sprint (2 casos nuevos), `repository_contract_test.go` de backlog (1 caso nuevo), corridos en memoria y con `TEST_DATABASE_URL` | `memory_repository.go` y `postgres_repository.go` de sprint (`GetByID`), `postgres_repository.go` de backlog (`sprint_id` en `itemColumns`, `Save`, `Update` y `scanItem`) |

## Fase RED

Tests, escenarios BDD y harness escritos antes del código de producción (commit `29164b8`). **El RED fue un error de compilación, no el fallo de una aserción**: `Manager.AssignItem`, `Manager.AssignToSprint`, el campo `SprintID`, `ErrItemAlreadyAssigned`, `ErrSprintNotFound`, `ErrSprintClosed` y `ErrItemFromAnotherProject` todavía no existían, y el constructor `NewManager` todavía no recibía el Manager de ítems.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
FAIL	seguimiento-medicion-gpc/internal/backlog [build failed]
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
FAIL	seguimiento-medicion-gpc/internal/sprint [build failed]
FAIL
# seguimiento-medicion-gpc/internal/backlog [seguimiento-medicion-gpc/internal.backlog.test]
internal\backlog\assign_test.go:22:27: manager.AssignToSprint undefined (type *Manager has no field or method
AssignToSprint)
internal\backlog\assign_test.go:34:11: saved.SprintID undefined (type Item has no field or method SprintID)
internal\backlog\assign_test.go:55:25: manager.AssignToSprint undefined (type *Manager has no field or method
AssignToSprint)
internal\backlog\assign_test.go:59:29: manager.AssignToSprint undefined (type *Manager has no field or method
AssignToSprint)
internal\backlog\assign_test.go:61:23: undefined: ErrItemAlreadyAssigned
internal\backlog\assign_test.go:113:23: too many errors
# seguimiento-medicion-gpc/internal/sprint [seguimiento-medicion-gpc/internal.sprint.test]
internal\sprint\assign_features_test.go:59:46: too many arguments in call to NewManager
	have (*MemoryRepository, *project.Manager, *backlog.Manager)
	want (Repository, *project.Manager)
internal\sprint\assign_features_test.go:217:32: s.manager.AssignItem undefined (type *Manager has no field or method
AssignItem)
internal\sprint\assign_features_test.go:232:29: s.manager.AssignItem undefined (type *Manager has no field or method
AssignItem)
internal\sprint\assign_features_test.go:268:13: s.assigned.SprintID undefined (type backlog.Item has no field or
method SprintID)
internal\sprint\assign_features_test.go:290:16: s.assigned.SprintID undefined (type backlog.Item has no field or
method SprintID)
internal\sprint\assign_features_test.go:317:10: too many errors
```

Todos los errores son la funcionalidad nueva de HU-11 (`AssignToSprint`, `AssignItem`, `SprintID`, `ErrItemAlreadyAssigned` y la firma nueva de `NewManager`, que la SDD definió). No hay fallos preexistentes: los paquetes `backlog` y `sprint` no compilan hasta que existan, así que en el RED tampoco corren los tests de las historias anteriores; el resto de los paquetes queda en verde.

## Fase GREEN

Código mínimo de producción (commit `805f6c9`):

- `internal/sprint/assign.go` (nuevo): `Manager.AssignItem` con el orden de validación de la SDD (sprint bajo el mutex y se suelta, estado, ítem, mismo proyecto, delegación en `AssignToSprint`); nunca toma dos mutex a la vez.
- `internal/sprint/sprint.go`: errores `ErrSprintNotFound`, `ErrSprintClosed` y `ErrItemFromAnotherProject`.
- `internal/sprint/repository.go`: operación `GetByID(id int) (Sprint, error)` en la interfaz `Repository` (RN-11.1).
- `internal/sprint/memory_repository.go` y `postgres_repository.go`: `GetByID` con `ErrSprintNotFound` si no existe.
- `internal/sprint/manager.go`: constructor con el tercer parámetro (`items *backlog.Manager`).
- `internal/backlog/assign.go` (nuevo): `ErrItemAlreadyAssigned` y `Manager.AssignToSprint` con lectura, verificación de RN-11.3 y escritura juntas bajo el mutex (CA-11.6).
- `internal/backlog/item.go`: campo `SprintID int` en `Item` (0 = sin asignar, RN-11.4).
- `internal/backlog/update.go`: conserva `SprintID` (la única línea de HU-07); `create.go`: `SprintID: 0` explícito (la única línea de HU-04).
- `internal/backlog/postgres_repository.go`: `sprint_id` en `itemColumns`, `Save`, `Update` y `scanItem`.
- `db/init/01-schema.sql`: columna `sprint_id INTEGER NOT NULL DEFAULT 0` en `backlog_items`, sin clave foránea (0 = sin asignar).

El mismo commit ajusta las llamadas a la firma nueva del constructor que quedaron con dos parámetros (`TestCreate_ErrorDelRepositorio...` en `create_test.go` de sprint y `TestNewRepositories_PostgresConservaLosDatosAlReabrir` en `cmd/server/main_test.go`): sin el código de GREEN compilan con la firma vieja, así que el ajuste pertenece a esta fase.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	4.867s
ok  	seguimiento-medicion-gpc/internal/backlog	3.636s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.663s
ok  	seguimiento-medicion-gpc/internal/postgres	1.923s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.254s
ok  	seguimiento-medicion-gpc/internal/project	1.932s
ok  	seguimiento-medicion-gpc/internal/sprint	1.739s
```

## Refactor

No se hizo: el código nuevo es mínimo y sigue los patrones existentes (`AssignItem` valida y delega como `Create` de HU-10; `AssignToSprint` hace lectura-modificación-escritura bajo el mutex como `Update` de HU-07; los `GetByID` nuevos son como los de proyectos e ítems). No hay lógica duplicada ni ramas sin test.

## Pruebas de mutación

Una por vez, restaurando el código después de cada una (ninguna forma parte de un commit):

1. **Permitir reasignar (se sacó la verificación de RN-11.3 en `AssignToSprint`)**: fallan `TestAssignToSprint_ItemYaAsignadoDaErrItemAlreadyAssigned` (sus dos subtests), `TestAssignItem_ItemYaAsignadoAlMismoSprintDaErrItemAlreadyAssigned`, `TestAssignItem_ItemYaAsignadoAOtroSprintDaErrItemAlreadyAssigned` y el esquema BDD "Asignar un ítem ya asignado" (la suite godog reporta escenarios fallidos). Con el código restaurado, todo vuelve a verde.
2. **Permitir asignar a un sprint CERRADO (se sacó el chequeo de estado en `AssignItem`)**: fallan `TestAssignItem_SprintCerradoDaErrSprintClosed` y el escenario BDD "Asignar a un sprint cerrado". Con el código restaurado, todo vuelve a verde.
3. **Sacar el mutex de `AssignToSprint`**: sin `-race`, los tests siguen pasando en este entorno (10 corridas de `TestAssignItem_AsignacionesSimultaneasDelMismoItemDejanUnaSola` y de la suite godog: el scheduler local no intercala la lectura y la escritura de las 50 asignaciones). La mutación no se pudo verificar localmente con el detector de carreras por falta de cgo: la verificación con `-race` (CA-11.6) es el check del CI del Pull Request, igual que en HU-07. Con el mutex restaurado, todo vuelve a verde.

## Validación final

```
$ gofmt -w internal/backlog/item.go internal/backlog/assign.go internal/backlog/create.go internal/backlog/update.go internal/backlog/postgres_repository.go internal/sprint/assign.go internal/sprint/manager.go internal/sprint/memory_repository.go internal/sprint/postgres_repository.go internal/sprint/repository.go internal/sprint/sprint.go
(gofmt no reportó ni cambió nada: todos los archivos ya estaban formateados)
```

```
$ go test -count=1 ./...      (sin base, como en el CI)
ok  	seguimiento-medicion-gpc/cmd/server	4.867s
ok  	seguimiento-medicion-gpc/internal/backlog	3.636s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.663s
ok  	seguimiento-medicion-gpc/internal/postgres	1.923s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.254s
ok  	seguimiento-medicion-gpc/internal/project	1.932s
ok  	seguimiento-medicion-gpc/internal/sprint	1.739s
```

```
$ go vet ./...
(sin salida: sin problemas)
```

### Cobertura

```
$ go test -count=1 -cover ./...      (sin base, como en el CI)
ok  	seguimiento-medicion-gpc/cmd/server	4.888s	coverage: 39.0% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	6.797s	coverage: 77.7% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.316s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.324s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.975s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	4.018s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	3.525s	coverage: 69.5% of statements
```

```
$ go test -count=1 -cover ./...      (con TEST_DATABASE_URL apuntando a gpc_test)
ok  	seguimiento-medicion-gpc/cmd/server	4.243s	coverage: 41.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	7.242s	coverage: 98.7% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.228s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.541s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	2.845s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	5.372s	coverage: 96.8% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	4.185s	coverage: 97.6% of statements
```

```
$ go tool cover -func      (funciones nuevas, sin base y con base)
seguimiento-medicion-gpc/internal/backlog/assign.go:18:			AssignToSprint		100.0%    (100.0% con base)
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:35:	GetByID			100.0%    (100.0% con base)
seguimiento-medicion-gpc/internal/backlog/postgres_repository.go:46:	GetByID			0.0%    (100.0% con base)
seguimiento-medicion-gpc/internal/sprint/assign.go:13:			AssignItem			100.0%    (100.0% con base)
seguimiento-medicion-gpc/internal/sprint/memory_repository.go:40:	GetByID			100.0%    (100.0% con base)
seguimiento-medicion-gpc/internal/sprint/postgres_repository.go:61:	GetByID			0.0%    (85.7% con base)
```

Sin regresiones: `AssignItem` y `AssignToSprint` y los `GetByID` nuevos entran con cobertura total en memoria; sin base, la parte de PostgreSQL queda sin ejecutar como en TEC-05 (sus tests hacen `t.Skip`, CA-TEC05.7) y por eso `internal/sprint` baja de 98.3% (con base de TEC-05) a 69.5% en la columna del CI: la rama de error genérico del `GetByID` de PostgreSQL (la misma que le falta a `OpenByProject`) solo se cubre con la base levantada. Con `-coverpkg=./...` y la base levantada, el total queda en **92.5%** (antes 92.1%).

### go test -race (CA-11.6)

En este entorno (Windows, sin compilador C) `-race` no está disponible:

```
$ go test -race ./...
go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
```

La verificación con `-race` la corre el CI de TEC-02 en cada Pull Request hacia main (`ci.yml`, paso "Correr los tests (go test -race ./...)"), igual que en las historias anteriores. Los tests de concurrencia (50 asignaciones simultáneas del mismo ítem a dos sprints y 10 de ítems distintos al mismo sprint, en el test unitario y en los escenarios BDD) corren y pasan también sin `-race` en el entorno local.

### Ejecución de los escenarios BDD

```
$ go test -v -run TestFeatures_HU11AsignarItemSprint ./internal/sprint
20 scenarios (20 passed)
119 steps (119 passed)
--- PASS: TestFeatures_HU11AsignarItemSprint (0.05s)
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_un_ítem_a_un_sprint_abierto_del_mismo_proyecto
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_dos_ítems_distintos_al_mismo_sprint
    --- PASS: TestFeatures_HU11AsignarItemSprint/Un_ítem_terminado_se_asigna_igual
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_no_cambia_ningún_otro_campo_del_ítem
    --- PASS: TestFeatures_HU11AsignarItemSprint/Modificar_el_ítem_con_el_Update_de_HU-07_conserva_la_asignación
    --- PASS: TestFeatures_HU11AsignarItemSprint/Modificar_el_ítem_devuelto_no_altera_el_ítem_guardado
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignaciones_simultáneas_del_mismo_ítem_a_dos_sprints
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignaciones_simultáneas_de_ítems_distintos_al_mismo_sprint
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_un_ítem_ya_asignado
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_un_ítem_ya_asignado#01
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_a_un_sprint_cerrado
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_un_ítem_de_otro_proyecto
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_a_un_sprint_inexistente
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_a_un_sprint_inexistente#01
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_a_un_sprint_inexistente#02
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_cuando_no_hay_sprints_creados
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_un_ítem_inexistente
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_un_ítem_inexistente#01
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_un_ítem_inexistente#02
    --- PASS: TestFeatures_HU11AsignarItemSprint/Asignar_cuando_no_hay_ítems_creados
```

Los 20 escenarios (12 escenarios + 3 esquemas con 2, 3 y 3 ejemplos) pasan con los 119 pasos, y con `Strict: true` godog habría fallado si algún paso del `.feature` quedara sin expresión registrada: no hay pasos undefined.

### Tests de contrato con la base levantada (CA-TEC05.3)

Con `TEST_DATABASE_URL` apuntando a la base de pruebas `gpc_test`, los casos nuevos de los contratos de TEC-05 pasan en los dos backends:

```
$ go test -count=1 -v -run "Contrato" ./internal/sprint ./internal/backlog      (con TEST_DATABASE_URL)
--- PASS: TestPostgresRepository_Contrato (sprint)
    --- PASS: TestPostgresRepository_Contrato/GetByID_devuelve_el_sprint_guardado
    --- PASS: TestPostgresRepository_Contrato/GetByID_de_un_ID_inexistente_da_ErrSprintNotFound
--- PASS: TestMemoryRepository_Contrato (sprint)
    --- PASS: TestMemoryRepository_Contrato/GetByID_devuelve_el_sprint_guardado
    --- PASS: TestMemoryRepository_Contrato/GetByID_de_un_ID_inexistente_da_ErrSprintNotFound
--- PASS: TestPostgresRepository_Contrato (backlog)
    --- PASS: TestPostgresRepository_Contrato/Save_y_Update_conservan_el_SprintID_del_ítem
--- PASS: TestMemoryRepository_Contrato (backlog)
    --- PASS: TestMemoryRepository_Contrato/Save_y_Update_conservan_el_SprintID_del_ítem
PASS
```

(Se listan solo los casos nuevos de HU-11; el resto de los casos del contrato sigue pasando igual.)

### Recreación de la base de datos

`db/init/01-schema.sql` solo corre la primera vez que PostgreSQL arranca con el volumen vacío, así que una base de desarrollo creada antes de HU-11 no tiene la columna `sprint_id`. En este entorno la base se recreó antes de correr las pruebas de integración:

```
$ docker compose down -v
$ docker compose up -d --wait
NAME                                  IMAGE             STATUS                   PORTS
seguimiento-medicion-gpc-postgres-1   postgres:17-alpine Up 5 seconds (healthy)   127.0.0.1:5434->5432/tcp

$ docker compose exec postgres psql -U gpc -d gpc -c "\d backlog_items"
 sprint_id   | integer   | not null   | 0
```

El mismo aviso quedó en el README (sección "Persistencia con PostgreSQL").

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | 29164b8 | `test(HU-11): agregar tests y escenarios BDD (RED)` |
| GREEN | 805f6c9 | `feat(HU-11): implementar asignacion de items a sprint (GREEN)` |

No hubo commit de refactor porque no se hizo refactor. La SDD entró en dos commits de documentación propios (`5bcd9a1` y `1dad973`, con las decisiones del 18/10). El commit de cierre (`Closes #25`) no se lista en esta tabla porque un commit no puede contener su propio hash: queda registrado en el Pull Request y en el Issue.
