# Evidencia TDD — HU-09 · Priorizar los ítems del backlog

> Issue del repositorio: #22

- **Responsable:** Cardozo Leandro
- **Rama:** `feature/HU-09-priorizar-backlog`
- **Fecha:** 20/10/2026
- **Go:** go1.27.1 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` lo verifica el CI del Pull Request)
- **Especificación SDD:** [`docs/especificaciones/HU-09-priorizar-backlog.md`](../especificaciones/HU-09-priorizar-backlog.md)
- **Escenarios BDD:** [`features/HU-09-priorizar-backlog.feature`](../../features/HU-09-priorizar-backlog.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**15 funciones de test cubriendo 13 ejecuciones de escenario BDD (9 escenarios simples y 1 esquema del escenario con 4 ejemplos; 74 pasos)**: 14 tests unitarios en `internal/backlog/priority_test.go` (que con sus 3 subtests hacen 17 ejecuciones) y 1 test que ejecuta los escenarios BDD con godog (`internal/backlog/priority_features_test.go`, runner `TestFeatures_HU09PriorizarBacklog`). El conteo de funciones sale del comando:

```
$ grep "^func Test" internal/backlog/priority_test.go internal/backlog/priority_features_test.go
internal/backlog/priority_test.go:func TestListItemsByPriority_OrdenaAltaMediaBajaConTodosSusDatos(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_MismaPrioridadConservaElOrdenPorID(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_TodosConLaMismaPrioridadQuedanPorID(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_UnSoloItem(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_ProyectoSinItemsDevuelveListaVaciaNoNil(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_DevuelveLosMismosItemsQueListItems(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_ProyectoFinalizadoListaSusItems(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_NoCambiaElOrdenDeListItems(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_PrioridadDesconocidaVaAlFinalPorID(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_ProyectoInexistenteDevuelveErrProjectNotFound(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_SinProyectosCreadosDevuelveErrProjectNotFound(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_ErrorDelRepositorioAlListarSeDevuelve(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_ListadosSimultaneosConCreaciones(t *testing.T) {
internal/backlog/priority_test.go:func TestListItemsByPriority_ModificarLaListaDevueltaNoAlteraLosItemsGuardados(t *testing.T) {
internal/backlog/priority_features_test.go:func TestFeatures_HU09PriorizarBacklog(t *testing.T) {
TOTAL: 15
```

Las 17 ejecuciones unitarias salen de:

```
$ go test -count=1 -v -run "^TestListItemsByPriority_" ./internal/backlog | grep -E "^\s*--- PASS"
--- PASS: TestListItemsByPriority_OrdenaAltaMediaBajaConTodosSusDatos (0.00s)
--- PASS: TestListItemsByPriority_MismaPrioridadConservaElOrdenPorID (0.00s)
--- PASS: TestListItemsByPriority_TodosConLaMismaPrioridadQuedanPorID (0.00s)
--- PASS: TestListItemsByPriority_UnSoloItem (0.00s)
--- PASS: TestListItemsByPriority_ProyectoSinItemsDevuelveListaVaciaNoNil (0.00s)
--- PASS: TestListItemsByPriority_DevuelveLosMismosItemsQueListItems (0.00s)
--- PASS: TestListItemsByPriority_ProyectoFinalizadoListaSusItems (0.00s)
--- PASS: TestListItemsByPriority_NoCambiaElOrdenDeListItems (0.00s)
--- PASS: TestListItemsByPriority_PrioridadDesconocidaVaAlFinalPorID (0.00s)
--- PASS: TestListItemsByPriority_ProyectoInexistenteDevuelveErrProjectNotFound (0.00s)
    --- PASS: TestListItemsByPriority_ProyectoInexistenteDevuelveErrProjectNotFound/ID_negativo (0.00s)
    --- PASS: TestListItemsByPriority_ProyectoInexistenteDevuelveErrProjectNotFound/ID_inexistente (0.00s)
    --- PASS: TestListItemsByPriority_ProyectoInexistenteDevuelveErrProjectNotFound/ID_cero (0.00s)
--- PASS: TestListItemsByPriority_SinProyectosCreadosDevuelveErrProjectNotFound (0.00s)
--- PASS: TestListItemsByPriority_ErrorDelRepositorioAlListarSeDevuelve (0.00s)
--- PASS: TestListItemsByPriority_ListadosSimultaneosConCreaciones (0.00s)
--- PASS: TestListItemsByPriority_ModificarLaListaDevueltaNoAlteraLosItemsGuardados (0.00s)
TOTAL PASS (funciones + subtests): 17
TOTAL FAIL: 0
```

Las 13 ejecuciones BDD salen de las filas del `.feature` (`grep -cE "^\s*(Escenario|Esquema del escenario):"` → 10 definiciones, de las cuales `grep -cE "^\s*Esquema del escenario:"` → 1 esquema, con `grep -cE "^\s*\| [0-9]+ +\| -?[0-9]+ +\|$"` → 4 filas de Ejemplos: 9 + 4 = 13) y de la salida de godog en "Ejecución de los escenarios BDD".

La correspondencia no es 1 a 1: hay un escenario por criterio, pero CA-09.2, CA-09.4 y CA-09.7 tienen dos tests unitarios cada uno, el caso límite "un solo ítem" de la SDD tiene un test unitario sin escenario, y el esquema de CA-09.7 multiplica sus 4 ejemplos en ejecuciones.

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-09.1 Orden ALTA, MEDIA, BAJA con todos sus datos | Listar por prioridad (CA-09.1) | `TestListItemsByPriority_OrdenaAltaMediaBajaConTodosSusDatos` | `priority.go` (`ListItemsByPriority`, `priorityRank`) |
| CA-09.2 Orden estable por ID a igual prioridad | Los ítems con la misma prioridad conservan el orden por ID (CA-09.2), con 24 ítems | `TestListItemsByPriority_MismaPrioridadConservaElOrdenPorID` (24 ítems), `TestListItemsByPriority_TodosConLaMismaPrioridadQuedanPorID` | `priority.go` (`slices.SortStableFunc` sobre la lista de `ListItems`) |
| CA-09.3 Backlog vacío: lista vacía, no `nil` | Listar por prioridad un backlog vacío (CA-09.3) | `TestListItemsByPriority_ProyectoSinItemsDevuelveListaVaciaNoNil` | `list.go` y `memory_repository.go` (`ListItems`, `ListByProject`) |
| CA-09.4 Mismos ítems que `ListItems` (estados, proyecto finalizado, aislamiento) | El listado por prioridad tiene los mismos ítems que el listado de HU-06 (CA-09.4) | `TestListItemsByPriority_DevuelveLosMismosItemsQueListItems`, `TestListItemsByPriority_ProyectoFinalizadoListaSusItems` | `priority.go` (reutiliza `ListItems`) |
| CA-09.5 `ListItems` sigue ordenado por ID | El listado de HU-06 sigue ordenado por ID después de listar por prioridad (CA-09.5) | `TestListItemsByPriority_NoCambiaElOrdenDeListItems` | `priority.go` (ordena una lista propia), `list.go` sin cambios |
| CA-09.6 Prioridad desconocida al final, por ID | Un ítem con una prioridad desconocida va al final (CA-09.6) | `TestListItemsByPriority_PrioridadDesconocidaVaAlFinalPorID` | `priority.go` (`priorityRank`, caso `default`) |
| CA-09.7 Proyecto inexistente (999, 0, negativo, sin proyectos) | Listar por prioridad el backlog de un proyecto inexistente (CA-09.7) (esquema, 4 ejemplos) | `TestListItemsByPriority_ProyectoInexistenteDevuelveErrProjectNotFound` (3 subtests), `TestListItemsByPriority_SinProyectosCreadosDevuelveErrProjectNotFound` | `priority.go` (propaga el error de `ListItems`), `list.go` (`project.Manager.Get`) |
| CA-09.8 Error del repositorio con lista `nil` | Si el repositorio falla al listar se devuelve ese error (CA-09.8) | `TestListItemsByPriority_ErrorDelRepositorioAlListarSeDevuelve` | `priority.go` (`return nil, err`) |
| CA-09.9 Listados simultáneos con creaciones, ordenados | Listar por prioridad mientras se crean ítems en simultáneo (CA-09.9) | `TestListItemsByPriority_ListadosSimultaneosConCreaciones` (verifica además el orden del listado final) | `list.go` (mutex del `Manager`), `priority.go` (ordena fuera del mutex) |
| CA-09.10 Modificar el resultado no altera lo guardado | Modificar el listado por prioridad no altera los ítems guardados (CA-09.10) | `TestListItemsByPriority_ModificarLaListaDevueltaNoAlteraLosItemsGuardados` | `memory_repository.go` (copias de `ListByProject`) |
| Caso límite de la SDD: un solo ítem | — | `TestListItemsByPriority_UnSoloItem` | `priority.go` |

**HU-09 no cambia la interfaz `Repository` ni el esquema de PostgreSQL** (RN-09.9): el orden se aplica en el `Manager`. Por eso los contratos de repositorio de TEC-05 no necesitan un caso nuevo: el orden por ID del que parte el ordenamiento estable ya lo verifica el contrato existente "ListByProject devuelve solo los ítems del proyecto en orden de ID". Los archivos del repositorio y del esquema no cambian respecto de `main`:

```
$ git diff --stat 46a2e8a -- internal/backlog/repository.go internal/backlog/memory_repository.go internal/backlog/postgres_repository.go internal/backlog/repository_contract_test.go db/
(sin salida = sin cambios)
```

## Fase RED

Tests y escenarios BDD escritos antes del código de producción (commit `4b2c14e`). **El RED fue un error de compilación, no el fallo de una aserción**: `Manager.ListItemsByPriority` todavía no existía. La salida siguiente se reprodujo sobre ese commit en un worktree temporal (`git worktree add --detach … 4b2c14e`, borrado después):

```
$ go test ./...
# seguimiento-medicion-gpc/internal/backlog [seguimiento-medicion-gpc/internal/backlog.test]
internal\backlog\priority_features_test.go:180:35: s.manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_features_test.go:185:35: s.manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_features_test.go:334:48: s.manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:57:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:85:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:103:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:117:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:130:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:163:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:190:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:190:22: too many errors
ok  	seguimiento-medicion-gpc/cmd/server	1.997s
FAIL	seguimiento-medicion-gpc/internal/backlog [build failed]
ok  	seguimiento-medicion-gpc/internal/httpapi	1.168s
ok  	seguimiento-medicion-gpc/internal/postgres	1.185s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.183s
ok  	seguimiento-medicion-gpc/internal/project	1.578s
ok  	seguimiento-medicion-gpc/internal/sprint	1.410s
FAIL
go: unlinkat C:\Users\LENOVO\AppData\Local\Temp\go-build1345873634\b001\server.test.exe: The process cannot access the file because it is being used by another process.
```

(La última línea es un aviso de Windows al borrar un ejecutable temporal de `go test`, ajeno a los tests; ver "Desviaciones".)

Como el compilador corta en "too many errors", se listaron todos los errores con `-gcflags=-e`: **19 errores, los 19 `ListItemsByPriority undefined`**, ninguno de otro tipo:

```
$ go test -gcflags=-e ./internal/backlog/...
# seguimiento-medicion-gpc/internal/backlog [seguimiento-medicion-gpc/internal/backlog.test]
internal\backlog\priority_features_test.go:180:35: s.manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_features_test.go:185:35: s.manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_features_test.go:334:48: s.manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:57:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:85:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:103:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:117:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:130:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:163:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:190:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:204:23: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:233:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:253:24: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:269:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:287:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:314:34: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:324:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:353:22: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
internal\backlog\priority_test.go:365:24: manager.ListItemsByPriority undefined (type *Manager has no field or method ListItemsByPriority)
FAIL	seguimiento-medicion-gpc/internal/backlog [build failed]
FAIL
errores totales: 19
por ListItemsByPriority undefined: 19
```

(Los dos totales salen de `grep -c '\.go:[0-9]*:[0-9]*:'` y `grep -c 'ListItemsByPriority undefined'` sobre esa salida.)

Todos los errores son la funcionalidad nueva de HU-09. No hay fallos preexistentes: el resto de los paquetes queda en verde, y `internal/backlog` no compila hasta que exista `ListItemsByPriority`, así que en el RED tampoco corren los tests de las historias anteriores de ese paquete.

**Verificación de las aserciones en el RED.** Como el fallo es de compilación, en el RED no llegan a ejecutarse ni las aserciones ni los pasos de godog. Para verificarlos se armó una **copia del repositorio fuera del repo** (en un directorio temporal) con un stub de `ListItemsByPriority` que devuelve `ListItems` sin ordenar. **El stub no es el GREEN y no quedó en el repositorio.** Con ese stub, godog informó `13 scenarios (7 passed, 6 failed)` y `74 steps (62 passed, 6 failed, 6 skipped)`, sin pasos undefined. Fallaron los 6 escenarios que esperan un orden por prioridad distinto del orden por ID:

```
[copia con stub] $ go test -count=1 -v -run "^TestFeatures_HU09PriorizarBacklog$" ./internal/backlog | grep -E "scenarios|steps|^    --- FAIL"
13 scenarios (7 passed, 6 failed)
74 steps (62 passed, 6 failed, 6 skipped)
    --- FAIL: TestFeatures_HU09PriorizarBacklog/Listar_por_prioridad_(CA-09.1) (0.00s)
    --- FAIL: TestFeatures_HU09PriorizarBacklog/Los_ítems_con_la_misma_prioridad_conservan_el_orden_por_ID_(CA-09.2) (0.00s)
    --- FAIL: TestFeatures_HU09PriorizarBacklog/El_listado_por_prioridad_tiene_los_mismos_ítems_que_el_listado_de_HU-06_(CA-09.4) (0.00s)
    --- FAIL: TestFeatures_HU09PriorizarBacklog/El_listado_de_HU-06_sigue_ordenado_por_ID_después_de_listar_por_prioridad_(CA-09.5) (0.00s)
    --- FAIL: TestFeatures_HU09PriorizarBacklog/Un_ítem_con_una_prioridad_desconocida_va_al_final_(CA-09.6) (0.00s)
    --- FAIL: TestFeatures_HU09PriorizarBacklog/Modificar_el_listado_por_prioridad_no_altera_los_ítems_guardados_(CA-09.10) (0.00s)
```

Pasaron los 7 restantes: CA-09.3 (lista vacía), los 4 ejemplos de CA-09.7 (proyecto inexistente), CA-09.8 (error del repositorio) y CA-09.9 (concurrencia). En los tests unitarios fallaron 6 de las 14 funciones, también las de orden: `_OrdenaAltaMediaBajaConTodosSusDatos`, `_MismaPrioridadConservaElOrdenPorID`, `_DevuelveLosMismosItemsQueListItems`, `_ProyectoFinalizadoListaSusItems`, `_PrioridadDesconocidaVaAlFinalPorID` y `_ModificarLaListaDevueltaNoAlteraLosItemsGuardados` (`go test -count=1 -v -run "^TestListItemsByPriority_" ./internal/backlog | grep -E "^\s*--- (FAIL|PASS)"` en la misma copia). `_NoCambiaElOrdenDeListItems` pasa con el stub porque solo verifica `ListItems`; su escenario de CA-09.5 falla porque además verifica el orden del listado por prioridad. Esa verificación se hizo antes de agregar al test de CA-09.9 el chequeo del orden final (ver "Desviaciones"), por eso ese test pasa con el stub.

## Fase GREEN

Código mínimo de producción (commit `393a53e`):

- `internal/backlog/priority.go` (nuevo): `Manager.ListItemsByPriority` llama a `ListItems`; si hay error lo devuelve tal cual con lista `nil`, y si no, ordena la lista con `slices.SortStableFunc` según `priorityRank` (ALTA 0, MEDIA 1, BAJA 2 y desconocida 3). No agrega mutex: `ListItems` toma el del `Manager` y devuelve una lista propia, que se ordena fuera del mutex.

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
ok  	seguimiento-medicion-gpc/cmd/server	2.096s
ok  	seguimiento-medicion-gpc/internal/backlog	2.183s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.914s
ok  	seguimiento-medicion-gpc/internal/postgres	0.496s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.497s
ok  	seguimiento-medicion-gpc/internal/project	0.608s
ok  	seguimiento-medicion-gpc/internal/sprint	2.123s
go: unlinkat C:\Users\LENOVO\AppData\Local\Temp\go-build1051145664\b001\server.test.exe: The process cannot access the file because it is being used by another process.
```

Los 7 paquetes dieron `ok`; la última línea es el aviso de Windows al borrar el ejecutable temporal (ver "Desviaciones"). La corrida siguiente terminó sin el aviso:

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/sprint	(cached)
exit=0
```

## Refactor

Sin cambios: no se hizo refactor. `priority.go` tiene 37 líneas, reutiliza `ListItems` sin duplicar la validación del proyecto y sus dos funciones quedan al 100.0% de cobertura; no hay una mejora clara de legibilidad o estructura que lo justifique. No se inventó un commit de refactor.

## Desviaciones

No hubo desviaciones de código respecto de la SDD aprobada. Ajustes y observaciones:

1. **Test de CA-09.9, antes del commit RED.** En la revisión del RED se pidió que `TestListItemsByPriority_ListadosSimultaneosConCreaciones`, después de comprobar los 50 ítems del listado final, verifique también que está ordenado (rango de prioridad no decreciente y, a igual prioridad, ID ascendente), con el rango definido en el propio test y sin usar el del código de producción. El ajuste entró en el commit RED `4b2c14e`; después del ajuste se volvió a verificar que el RED solo falla por `ListItemsByPriority undefined` (salida de arriba). Ese chequeo es el que hace que el test de concurrencia también detecte la mutación a).
2. **Aviso `unlinkat` de Windows.** Algunas corridas de `go test ./...` terminan con `go: unlinkat … server.test.exe: The process cannot access the file because it is being used by another process` y exit 1, aunque todos los paquetes dan `ok`: Windows no deja borrar un ejecutable temporal de `go test` que todavía está en uso. No es un fallo de los tests; al repetir la corrida termina con exit 0.
3. **Commit de la SDD.** El párrafo "Notas del acta que no cambian el alcance" de la SDD se sacó en el commit de esta evidencia, porque era información de calendario y no de especificación.

## Pruebas de mutación

Una por vez, rompiendo `internal/backlog/priority.go` a propósito, corriendo `go test -count=1 ./internal/backlog/...` y restaurando con `git checkout -- internal/backlog/priority.go` (después de cada restauración `git diff` no mostró diferencias con el GREEN; ninguna mutación forma parte de un commit). La cantidad de tests que fallan es la de líneas `--- FAIL` de nivel superior (`grep -c '^--- FAIL'`), donde el runner de godog cuenta como uno.

| Mutación | Resultado | Tests que fallan | Escenarios BDD (godog) | Qué la detecta |
|---|---|---|---|---|
| a) `slices.SortFunc` en vez de `slices.SortStableFunc` | **Matada** | 3 | `13 scenarios (12 passed, 1 failed)` | Escenario CA-09.2, `_MismaPrioridadConservaElOrdenPorID` (24 ítems) y `_ListadosSimultaneosConCreaciones` (chequeo del orden final con 50 ítems) |
| b) Invertir el orden de prioridad (BAJA primero) | **Matada** | 8 | `13 scenarios (7 passed, 6 failed)` | Escenarios CA-09.1, CA-09.2, CA-09.4, CA-09.5, CA-09.6 y CA-09.10, y 7 tests unitarios |
| c) Prioridades desconocidas al principio (`default: return -1`) | **Matada** | 2 | `13 scenarios (12 passed, 1 failed)` | Escenario CA-09.6 y `_PrioridadDesconocidaVaAlFinalPorID` |
| d) Comparar la prioridad como texto (`strings.Compare`, orden alfabético) | **Matada** | 6 | `13 scenarios (9 passed, 4 failed)` | Escenarios CA-09.1, CA-09.2, CA-09.4 y CA-09.6, y 5 tests unitarios |
| e) Desempatar por ID descendente | **Matada** | 5 | `13 scenarios (11 passed, 2 failed)` | Escenarios CA-09.2 y CA-09.6, `_MismaPrioridadConservaElOrdenPorID`, `_TodosConLaMismaPrioridadQuedanPorID`, `_PrioridadDesconocidaVaAlFinalPorID` y `_ListadosSimultaneosConCreaciones` |
| f) No validar el proyecto (`m.repo.ListByProject` en lugar de `m.ListItems`) | **Matada** | 3 | `13 scenarios (9 passed, 4 failed)` | Los 4 ejemplos del esquema CA-09.7, `_ProyectoInexistenteDevuelveErrProjectNotFound` (3 subtests) y `_SinProyectosCreadosDevuelveErrProjectNotFound` |

Las 6 mutaciones quedan matadas; ninguna sobrevive.

**Sobre la f):** al llamar directo al repositorio, la mutación también **saltea el mutex del `Manager`**, que toma `ListItems`. Localmente la detectan los tests de proyecto inexistente; la carrera de datos que agrega solo la ve `go test -race`, que corre en el CI porque el entorno local no tiene cgo.

**Sobre la a):** con 12 ítems o menos, `slices.SortFunc` usa ordenamiento por inserción y no se distingue de uno estable; por eso el test de CA-09.2 usa 24 ítems en ciclo BAJA, ALTA, MEDIA. Antes de fijar los datos se comprobó con un programa aparte que, con esos 24 ítems, `SortFunc` cambia el orden de los iguales y con 12 no.

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
exit=0
```

### Cobertura

```
$ go test -count=1 -cover ./internal/backlog
ok  	seguimiento-medicion-gpc/internal/backlog	3.801s	coverage: 82.0% of statements
```

```
$ go test -count=1 -coverprofile=cover.out ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.262s	coverage: 39.0% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	0.533s	coverage: 82.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.068s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	0.171s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.187s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	0.395s	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	3.684s	coverage: 76.9% of statements
exit=0

$ go tool cover -func=cover.out | grep priority.go
seguimiento-medicion-gpc/internal/backlog/priority.go:13:		ListItemsByPriority	100.0%
seguimiento-medicion-gpc/internal/backlog/priority.go:26:		priorityRank		100.0%

$ go tool cover -func=cover.out | tail -1
total:									(statements)		75.1%
```

Para comparar, `internal/backlog` medido sobre `main` (`46a2e8a`, el merge del PR #100) en un worktree temporal, borrado después:

```
$ git worktree add --detach <ruta temporal> 46a2e8a
$ go test -count=1 -cover ./internal/backlog
ok  	seguimiento-medicion-gpc/internal/backlog	2.446s	coverage: 80.9% of statements
go: unlinkat C:\Users\LENOVO\AppData\Local\Temp\go-build392521657\b001\backlog.test.exe: The process cannot access the file because it is being used by another process.
$ git worktree remove <ruta temporal>
```

(La última línea de `go test` es el aviso de Windows descripto en "Desviaciones"; el paquete dio `ok`.)

Las dos funciones de `priority.go` quedan al 100.0%: `ListItemsByPriority` 100.0% y `priorityRank` 100.0%. Sin base (como en el CI, sin `TEST_DATABASE_URL`), `internal/backlog` sube de 80.9% en `main` a 82.0% y el total del perfil queda en 75.1%. La parte de PostgreSQL no se ejecuta sin base; HU-09 no la modifica, así que no se repitió la medición con base. El archivo `cover.out` se generó fuera del repositorio y no se versiona.

### go test -race (CA-09.9)

En este entorno (Windows, sin compilador C) `-race` no está disponible. La verificación con `-race` la corre el CI de TEC-02 en cada Pull Request hacia main (`ci.yml`, paso "Correr los tests (go test -race ./...)"), igual que en las historias anteriores. Los tests de concurrencia (50 creaciones simultáneas con 50 listados por prioridad, en el test unitario y en el escenario BDD) corren y pasan también sin `-race` en el entorno local.

### Ejecución de los escenarios BDD

```
$ go test -count=1 -v -run "^TestFeatures_HU09PriorizarBacklog$" ./internal/backlog | grep -E "scenarios|steps|^--- |^ok|^FAIL"
13 scenarios (13 passed)
74 steps (74 passed)
--- PASS: TestFeatures_HU09PriorizarBacklog (0.03s)
ok  	seguimiento-medicion-gpc/internal/backlog	0.275s
```

Las 13 ejecuciones (9 escenarios simples y 1 esquema con 4 ejemplos) pasan con los 74 pasos, y con `Strict: true` godog habría fallado si algún paso del `.feature` quedara sin expresión registrada: no hay pasos undefined.

### Tests de contrato (TEC-05)

HU-09 no agrega casos de contrato (ver "Trazabilidad"). El contrato de `MemoryRepository` sigue pasando y el de PostgreSQL hizo SKIP porque `TEST_DATABASE_URL` no está definida en este entorno:

```
$ go test -count=1 -v -run "^Test(Memory|Postgres)Repository_Contrato$" ./internal/backlog | grep -E "^(--- |ok|FAIL|PASS)|TEST_DATABASE_URL"
    postgres_repository_test.go:39: TEST_DATABASE_URL no está definida: se omiten las pruebas de PostgreSQL
--- SKIP: TestPostgresRepository_Contrato (0.00s)
--- PASS: TestMemoryRepository_Contrato (0.00s)
PASS
ok  	seguimiento-medicion-gpc/internal/backlog	0.258s
```

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | `b5e150d` | `docs(HU-09): agregar especificación SDD` |
| RED | `4b2c14e` | `test(HU-09): agregar tests y escenarios BDD (RED)` |
| GREEN | `393a53e` | `feat(HU-09): implementar el listado de ítems por prioridad (GREEN)` |

No hubo commit de refactor porque no se hizo refactor. El commit de cierre (`Closes #22`) no se lista en esta tabla porque un commit no puede contener su propio hash: queda registrado en el Pull Request y en el Issue.
