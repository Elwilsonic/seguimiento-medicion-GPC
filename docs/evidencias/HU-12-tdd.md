# Evidencia TDD — HU-12 · Registrar historias completadas

> Issue del repositorio: #26

- **Responsable:** García Santiago
- **Rama:** `feature/HU-12-registrar-completadas`
- **Fecha:** 20/10/2026
- **Go:** go1.27.1 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` lo verifica el CI del Pull Request y Peñalbé Hernán sobre Linux)
- **Especificación SDD:** [`docs/especificaciones/HU-12-registrar-completadas.md`](../especificaciones/HU-12-registrar-completadas.md)
- **Escenarios BDD:** [`features/HU-12-registrar-completadas.feature`](../../features/HU-12-registrar-completadas.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**21 funciones de test cubriendo 13 ejecuciones de escenario BDD (13 escenarios, sin esquemas del escenario, 102 pasos)**: 20 tests unitarios en `internal/sprint/complete_test.go` y 1 test que ejecuta los escenarios BDD con godog (`internal/sprint/complete_features_test.go`, runner `TestFeatures_HU12RegistrarCompletadas`, en el paquete `sprint` porque la historia es de E3 y orquesta los paquetes `sprint` y `backlog`). El conteo sale del comando:

```
$ findstr /R "^func.Test" internal\sprint\complete_test.go internal\sprint\complete_features_test.go
internal\sprint\complete_test.go:func TestCompleteItem_ItemEnCursoAsignadoQuedaTerminado(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_ItemYaTerminadoEsNoOpSinUpdate(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_ItemPendienteDaErrInvalidTransition(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_ItemSinAsignarDaErrItemNotAssignedToSprint(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_ItemAsignadoAOtroSprintDaErrItemNotAssignedToSprint(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_SprintCerradoDaErrSprintClosed(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_SprintInexistenteDaErrSprintNotFound(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_ItemInexistenteDaErrItemNotFound(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_ErrorDelRepositorioDeSprintsSeDevuelve(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_ErrorDelRepositorioDeItemsSeDevuelve(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_ErrorDelRepositorioAlActualizarSeDevuelve(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_ModificarElItemDevueltoNoAlteraElGuardado(t *testing.T) {
internal\sprint\complete_test.go:func TestCompleteItem_MarcacionesSimultaneasDejanUnSoloUpdate(t *testing.T) {
internal\sprint\complete_test.go:func TestListCompletedItems_SoloDevuelveTerminadosDelSprintOrdenados(t *testing.T) {
internal\sprint\complete_test.go:func TestListCompletedItems_SprintSinCompletadosDevuelveListaVaciaNoNil(t *testing.T) {
internal\sprint\complete_test.go:func TestListCompletedItems_SprintCerradoSePuedeConsultar(t *testing.T) {
internal\sprint\complete_test.go:func TestListCompletedItems_SprintInexistenteDaErrSprintNotFound(t *testing.T) {
internal\sprint\complete_test.go:func TestListCompletedItems_ErrorDelRepositorioSeDevuelve(t *testing.T) {
internal\sprint\complete_test.go:func TestListCompletedItems_ModificarLaListaDevueltaNoAlteraElGuardado(t *testing.T) {
internal\sprint\complete_test.go:func TestListCompletedItems_ItemReabiertoNoAparece(t *testing.T) {
internal\sprint\complete_features_test.go:func TestFeatures_HU12RegistrarCompletadas(t *testing.T) {
```

Además de los 20 tests unitarios, `TestCompleteItem_SprintInexistenteDaErrSprintNotFound`, `TestCompleteItem_ItemInexistenteDaErrItemNotFound` y `TestListCompletedItems_SprintInexistenteDaErrSprintNotFound` tienen subtests con los IDs 999, 0 y -1.

La correspondencia no es 1 a 1: los errores del repositorio (CA-12.8) no tienen escenario BDD y se cubren con cuatro tests unitarios; la reapertura de un ítem (RN-12.6) también es solo unitaria.

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-12.1 Un ítem `EN_CURSO` asignado a un sprint `ABIERTO` queda `TERMINADO`, como copia y sin cambiar el resto | Completar un ítem en curso asignado a un sprint abierto | `TestCompleteItem_ItemEnCursoAsignadoQuedaTerminado` | `complete.go` (`CompleteItem`), `manager.go` (mutex de sprint) |
| CA-12.2 Un ítem que no pertenece al sprint (sin asignar o en otro) da `ErrItemNotAssignedToSprint` | No se puede completar un ítem sin asignar; No se puede completar un ítem asignado a otro sprint | `TestCompleteItem_ItemSinAsignarDaErrItemNotAssignedToSprint`, `TestCompleteItem_ItemAsignadoAOtroSprintDaErrItemNotAssignedToSprint` | `complete.go` (RN-12.3), `sprint.go` (`ErrItemNotAssignedToSprint`) |
| CA-12.3 `ListCompletedItems` devuelve solo `TERMINADO`, ordenados por ID, como copias, lista vacía no `nil`, sprint abierto o cerrado | Consultar solo los completados de un sprint, ordenados por ID; Un sprint sin ítems completados devuelve una lista vacía; Consultar los completados de un sprint cerrado | `TestListCompletedItems_SoloDevuelveTerminadosDelSprintOrdenados`, `TestListCompletedItems_SprintSinCompletadosDevuelveListaVaciaNoNil`, `TestListCompletedItems_SprintCerradoSePuedeConsultar`, `TestListCompletedItems_ItemReabiertoNoAparece` (RN-12.6) | `complete.go` (`ListCompletedItems`), `backlog/list.go` (`ListItems`) |
| CA-12.4 Completar un ítem ya `TERMINADO` es un no-op sin `Update` y devuelve copia | Completar un ítem ya terminado es un no-op | `TestCompleteItem_ItemYaTerminadoEsNoOpSinUpdate` | `complete.go` y `backlog/status.go` (no-op de HU-08) |
| CA-12.5 Sprint inexistente, ítem inexistente o sprint `CERRADO` dan error sin modificar nada | No se puede completar un ítem en un sprint cerrado; No se puede completar un ítem inexistente; No se puede completar en un sprint inexistente | `TestCompleteItem_SprintCerradoDaErrSprintClosed`, `TestCompleteItem_SprintInexistenteDaErrSprintNotFound` (3 subtests), `TestCompleteItem_ItemInexistenteDaErrItemNotFound` (3 subtests), `TestListCompletedItems_SprintInexistenteDaErrSprintNotFound` (3 subtests) | `complete.go` (orden de validación), `memory_repository.go` y `backlog/get.go` (errores de HU-11 y HU-05) |
| CA-12.6 `PENDIENTE → TERMINADO` da `ErrInvalidTransition` y el ítem queda `PENDIENTE` | No se puede completar un ítem PENDIENTE directamente | `TestCompleteItem_ItemPendienteDaErrInvalidTransition` | `complete.go` y `backlog/status.go` (transiciones de HU-08) |
| CA-12.7 Modificar el ítem devuelto, sus criterios o la lista devuelta no altera lo guardado | Modificar el ítem devuelto no altera el guardado | `TestCompleteItem_ModificarElItemDevueltoNoAlteraElGuardado`, `TestListCompletedItems_ModificarLaListaDevueltaNoAlteraElGuardado` | `backlog/memory_repository.go` (copias), `backlog/list.go` y `backlog/get.go` |
| CA-12.8 El error del repositorio se devuelve y no se modifica nada | — (no tiene escenario BDD) | `TestCompleteItem_ErrorDelRepositorioDeSprintsSeDevuelve`, `TestCompleteItem_ErrorDelRepositorioDeItemsSeDevuelve`, `TestCompleteItem_ErrorDelRepositorioAlActualizarSeDevuelve`, `TestListCompletedItems_ErrorDelRepositorioSeDevuelve` | `complete.go` (propagación de errores) |
| CA-12.9 Varias marcaciones simultáneas del mismo ítem dejan `TERMINADO`, con una sola escritura y sin carreras | Completar en simultáneo el mismo ítem deja el estado terminado | `TestCompleteItem_MarcacionesSimultaneasDejanUnSoloUpdate` (50 intentos) | `complete.go` y `backlog/status.go` (no-op y escritura bajo el mutex de backlog) |

## Fase RED

Tests, escenarios BDD y runner escritos antes del código de producción (commit `8244f1b`, después del commit de SDD `2cac48e`). **El RED fue un error de compilación, no el fallo de una aserción**: `Manager.CompleteItem`, `Manager.ListCompletedItems` y `ErrItemNotAssignedToSprint` todavía no existían.

```
$ go test ./...
# seguimiento-medicion-gpc/internal/sprint [seguimiento-medicion-gpc/internal/sprint.test]
internal\sprint\complete_features_test.go:154:33: s.manager.CompleteItem undefined (type *Manager has no field or method CompleteItem)
internal\sprint\complete_features_test.go:159:38: s.manager.ListCompletedItems undefined (type *Manager has no field or method ListCompletedItems)
internal\sprint\complete_features_test.go:170:29: s.manager.CompleteItem undefined (type *Manager has no field or method CompleteItem)
internal\sprint\complete_features_test.go:185:23: undefined: ErrItemNotAssignedToSprint
internal\sprint\complete_test.go:87:30: s.manager.CompleteItem undefined (type *Manager has no field or method CompleteItem)
internal\sprint\complete_test.go:157:28: manager.CompleteItem undefined (type *Manager has no field or method CompleteItem)
internal\sprint\complete_test.go:178:30: s.manager.CompleteItem undefined (type *Manager has no field or method CompleteItem)
internal\sprint\complete_test.go:198:30: s.manager.CompleteItem undefined (type *Manager has no field or method CompleteItem)
internal\sprint\complete_test.go:200:21: undefined: ErrItemNotAssignedToSprint
internal\sprint\complete_test.go:220:30: s.manager.CompleteItem undefined (type *Manager has no field or method CompleteItem)
internal\sprint\complete_test.go:220:30: too many errors
ok  	seguimiento-medicion-gpc/cmd/server	1.827s
ok  	seguimiento-medicion-gpc/internal/backlog	2.893s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.524s
ok  	seguimiento-medicion-gpc/internal/postgres	2.931s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.945s
ok  	seguimiento-medicion-gpc/internal/project	2.296s
FAIL	seguimiento-medicion-gpc/internal/sprint [build failed]
FAIL
```

Todos los errores son la funcionalidad nueva de HU-12. No hay fallos preexistentes: el resto de los paquetes queda en verde.

## Fase GREEN

Código mínimo de producción (commit `9129e6a`):

- `internal/sprint/complete.go` (nuevo): `Manager.CompleteItem` con el orden de validación de la SDD (sprint bajo el mutex y se suelta, estado del sprint, ítem, pertenencia al sprint y `ChangeStatus(itemID, string(backlog.StatusFinished))`) y `Manager.ListCompletedItems` (sprint bajo el mutex y se suelta, `ListItems` de HU-06 y filtro por `SprintID` y `Status == TERMINADO`). Nunca toma dos mutex a la vez.
- `internal/sprint/sprint.go`: error `ErrItemNotAssignedToSprint` junto a los demás del paquete.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.832s
ok  	seguimiento-medicion-gpc/internal/backlog	1.088s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.531s
ok  	seguimiento-medicion-gpc/internal/postgres	0.643s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.604s
ok  	seguimiento-medicion-gpc/internal/project	1.029s
ok  	seguimiento-medicion-gpc/internal/sprint	1.077s
```

## Refactor

No se hizo: el código nuevo es mínimo y sigue los patrones existentes (`CompleteItem` valida y delega como `AssignItem` de HU-11; `ListCompletedItems` reutiliza `ListItems` de HU-06 y filtra sin agregar operaciones al repositorio). No hay lógica duplicada ni ramas sin test.

## Desviaciones

En el GREEN se corrigieron dos expresiones de pasos BDD de `internal/sprint/complete_features_test.go`, que venían del commit RED y no se habían podido ejecutar porque el paquete no compilaba:

1. Los pasos `cuando completo...`, `cuando consulto...` y `cuando completo en simultáneo...` tenían el prefijo `cuando` dentro de la expresión regular. Godog recibe el texto del paso sin la palabra clave, así que quedaban undefined; se sacó el prefijo.
2. El paso `la lista devuelta tiene (\d+) ítems` solo aceptaba el plural, y el escenario del sprint cerrado dice "1 ítem" en singular; la expresión quedó como `ítems?`.

Ninguna corrección toca código de producción: son ajustes del harness de pruebas que pertenecen al RED y que quedaron dentro del commit GREEN. No hay código de producción sin test.

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)
```

```
$ go vet ./...
(sin salida: sin problemas)
```

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.832s
ok  	seguimiento-medicion-gpc/internal/backlog	1.088s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.531s
ok  	seguimiento-medicion-gpc/internal/postgres	0.643s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.604s
ok  	seguimiento-medicion-gpc/internal/project	1.029s
ok  	seguimiento-medicion-gpc/internal/sprint	1.077s
```

### Cobertura

Valores tomados del log del CI del PR #100 (Linux), job "gofmt y tests":

```
$ go test -race -coverprofile=coverage.out ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)	coverage: 39.0% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	1.294s	coverage: 80.9% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	1.175s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	1.268s	coverage: 76.9% of statements
```

```
$ go tool cover -func=coverage.out | tail -n 1
total:									(statements)		74.6%
```

`internal/sprint` sube de 69.5% en `main` a 76.9% con HU-12. El total del perfil es **74.6%**. Las funciones nuevas entran con cobertura total en memoria: `CompleteItem` y `ListCompletedItems` quedan al 100.0% en `internal/sprint/complete.go`. La medición local en Windows queda fuera de este reporte; el CI (Linux) es la fuente autoritativa.

Sin regresiones: HU-12 no agrega columnas ni operaciones a PostgreSQL (RN-12.9), así que no cambia el esquema ni los contratos de TEC-05. La columna "Con base" del reporte conserva la última medición con base (HU-11).

### go test -race (CA-12.9)

En este entorno (Windows, sin compilador C) `-race` no está disponible:

```
$ go test -race ./...
go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
```

La verificación con `-race`: "verificado por el CI del PR y por Hernán en Linux, con prueba de mutación (sin el mutex de `ChangeStatus` aparecen `DATA RACE` y fallan `TestCompleteItem_MarcacionesSimultaneasDejanUnSoloUpdate` y el escenario BDD)". Los tests de concurrencia (50 marcaciones simultáneas del mismo ítem en el test unitario y 10 en el escenario BDD) corren y pasan también sin `-race` en el entorno local.

### Ejecución de los escenarios BDD

```
$ go test -v -run TestFeatures_HU12RegistrarCompletadas ./internal/sprint
13 scenarios (13 passed)
102 steps (102 passed)
--- PASS: TestFeatures_HU12RegistrarCompletadas (0.02s)
ok  	seguimiento-medicion-gpc/internal/sprint	0.720s
```

Los 13 escenarios pasan con los 102 pasos y sin esquemas del escenario. Con `Strict: true` godog habría fallado si algún paso del `.feature` quedara sin expresión registrada: no hay pasos undefined.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | 2cac48e | `docs(HU-12): agregar especificación SDD` |
| RED | 8244f1b | `test(HU-12): agregar tests (RED)` |
| GREEN | 9129e6a | `feat(HU-12): implementar registrar historias completadas (GREEN)` |

No hubo commit de refactor porque no se hizo refactor.
