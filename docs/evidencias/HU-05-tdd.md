# Evidencia TDD — HU-05 · Consultar un ítem del backlog

> Issue del repositorio: #18

- **Responsable:** Sosa Ricardo
- **Rama:** `feature/HU-05-consultar-item`
- **Fecha:** 15/10/2026
- **Go:** go1.27.1 windows/amd64
- **Especificación SDD:** [`docs/especificaciones/HU-05-consultar-item.md`](../especificaciones/HU-05-consultar-item.md)
- **Escenarios BDD:** [`features/HU-05-consultar-item.feature`](../../features/HU-05-consultar-item.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**10 funciones de test cubriendo 6 escenarios BDD**: 9 tests unitarios (`internal/backlog/get_test.go`) y 1 test que ejecuta los escenarios BDD con godog (`internal/backlog/get_features_test.go`, runner `TestFeatures_HU05ConsultarItem`). El conteo sale del comando:

```
$ findstr /R "^func.Test" internal\backlog\get_test.go internal\backlog\get_features_test.go
internal\backlog\get_test.go:func TestGetItem_ItemExistenteDevuelveCopiaConTodosSusDatos(t *testing.T) {
internal\backlog\get_test.go:func TestGetItem_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
internal\backlog\get_test.go:func TestGetItem_ConsultarSinItemsCreadosDevuelveErrItemNotFound(t *testing.T) {
internal\backlog\get_test.go:func TestGetItem_ModificarElItemConsultadoYSusCriteriosNoAlteraElGuardado(t *testing.T) {
internal\backlog\get_test.go:func TestGetItem_ConsultasSimultaneasMientrasSeCreanItems(t *testing.T) {
internal\backlog\get_test.go:func TestGetItem_ErrorDelRepositorioSeDevuelveConItemVacio(t *testing.T) {
internal\backlog\get_test.go:func TestGetByID_ItemExistenteDevuelveCopiaConCriteriosPropios(t *testing.T) {
internal\backlog\get_test.go:func TestGetByID_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
internal\backlog\get_test.go:func TestGetByID_ConsultarSinItemsCreadosDevuelveErrItemNotFound(t *testing.T) {
internal\backlog\get_features_test.go:func TestFeatures_HU05ConsultarItem(t *testing.T) {
```

9 de esas funciones son tests unitarios y 1 es el runner de godog. Dos tests unitarios tienen 3 subtests cada uno (`TestGetItem_ItemInexistenteDevuelveErrItemNotFound` y `TestGetByID_ItemInexistenteDevuelveErrItemNotFound`, con ID 999, 0 y -1), así que las ejecuciones de tests unitarios son 13 (9 funciones + 6 subtests).

El archivo `.feature` tiene 5 escenarios y 1 esquema del escenario con 3 ejemplos (6 bloques de escenario); godog ejecuta 8 escenarios con 53 pasos, todos en verde. La correspondencia no es 1 a 1: el error del repositorio (CA-05.5) no tiene escenario BDD propio (va solo como test unitario, como definió la SDD), y la consulta sin ítems creados y la copia por valor del repositorio tienen tests unitarios propios además de los escenarios. El test de `GetByID` sobre `MemoryRepository` cubre directamente la operación de repositorio que HU-07 y HU-08 van a reutilizar.

Los tests y los pasos de godog son **internos al paquete** (`package backlog`), igual que en HU-04 y HU-06. Los tests reutilizan los helpers de `create_test.go` (`newTestManager`, `createTestProject`, `validTitle`, `validPriority`, `validCriteria`, `errRepository`) y extienden `failingRepository` con el campo `getErr` y el método `GetByID`, como anticipó la SDD.

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-05.1 Se obtiene el ítem por su ID, con todos sus datos | Consultar un ítem por su ID devuelve todos sus datos; Con ítems en varios proyectos se obtiene el ítem pedido y no otro | `TestGetItem_ItemExistenteDevuelveCopiaConTodosSusDatos`, `TestGetByID_ItemExistenteDevuelveCopiaConCriteriosPropios` | `get.go` (`GetItem`), `memory_repository.go` (`GetByID`) |
| CA-05.2 Consultar un ítem que no existe devuelve `ErrItemNotFound` (incluidos ID inexistente, 0, negativo y sin ítems creados) | No se puede consultar un ítem con un ID que no existe (esquema, 3 ejemplos); Consultar un ítem cuando no hay ítems creados | `TestGetItem_ItemInexistenteDevuelveErrItemNotFound` (3 subtests), `TestGetItem_ConsultarSinItemsCreadosDevuelveErrItemNotFound`, `TestGetByID_ItemInexistenteDevuelveErrItemNotFound` (3 subtests), `TestGetByID_ConsultarSinItemsCreadosDevuelveErrItemNotFound` | `get.go` (`GetItem`), `memory_repository.go` (`GetByID`) |
| CA-05.3 Modificar el ítem devuelto o su lista de criterios no altera el ítem guardado | Modificar el ítem consultado y sus criterios no altera el ítem guardado | `TestGetItem_ModificarElItemConsultadoYSusCriteriosNoAlteraElGuardado`, `TestGetByID_ItemExistenteDevuelveCopiaConCriteriosPropios` | `get.go` (`GetItem`), `memory_repository.go` (`GetByID`) |
| CA-05.4 Consultar mientras se crean ítems en simultáneo no produce condiciones de carrera (con `-race`) | Consultar mientras se crean ítems en simultáneo | `TestGetItem_ConsultasSimultaneasMientrasSeCreanItems` | `manager.go` (mutex), `get.go`, `memory_repository.go` |
| CA-05.5 Si el repositorio falla al buscar, se devuelve ese error y un ítem vacío | — (no tiene escenario BDD) | `TestGetItem_ErrorDelRepositorioSeDevuelveConItemVacio` | `get.go` (`GetItem`), `create_test.go` (`failingRepository`) |

## Fase RED

Tests y escenarios escritos antes del código de producción (commit a9afebc). `go test ./...` falla únicamente en `internal/backlog`, con errores `undefined` sobre la operación que todavía no existe:

```
$ go test ./...
# seguimiento-medicion-gpc/internal/backlog [seguimiento-medicion-gpc/internal/backlog.test]
internal\backlog\get_features_test.go:111:31: s.manager.GetItem undefined (type *Manager has no field or method GetItem)
internal\backlog\get_features_test.go:232:27: s.manager.GetItem undefined (type *Manager has no field or method GetItem)
internal\backlog\get_features_test.go:271:23: undefined: ErrItemNotFound
internal\backlog\get_test.go:26:26: manager.GetItem undefined (type *Manager has no field or method GetItem)
internal\backlog\get_test.go:75:25: manager.GetItem undefined (type *Manager has no field or method GetItem)
internal\backlog\get_test.go:77:23: undefined: ErrItemNotFound
internal\backlog\get_test.go:92:23: manager.GetItem undefined (type *Manager has no field or method GetItem)
internal\backlog\get_test.go:94:21: undefined: ErrItemNotFound
internal\backlog\get_test.go:110:26: manager.GetItem undefined (type *Manager has no field or method GetItem)
internal\backlog\get_test.go:123:24: manager.GetItem undefined (type *Manager has no field or method GetItem)
internal\backlog\get_test.go:123:24: too many errors
ok  	seguimiento-medicion-gpc/cmd/server	1.045s
FAIL	seguimiento-medicion-gpc/internal/backlog [build failed]
ok  	seguimiento-medicion-gpc/internal/httpapi	1.334s
ok  	seguimiento-medicion-gpc/internal/project	1.190s
FAIL
```

Todos los errores son `GetItem undefined` y `ErrItemNotFound undefined`: la funcionalidad nueva de HU-05. No hay fallos preexistentes: el paquete `backlog` no compila hasta que existan, así que en el RED tampoco corren los tests de HU-04 y HU-06; el resto de los paquetes queda en verde.

## Fase GREEN

Código mínimo de producción: `ErrItemNotFound` y `GetItem` en el `Manager` (nuevo archivo `get.go`, con mutex y delegación en el repositorio), la operación `GetByID` en la interfaz `Repository` y su implementación en `MemoryRepository` (copia del ítem y de sus criterios con `slices.Clone`, `ErrItemNotFound` si no existe). No se tocó ningún otro código de producción ni ningún archivo de test.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	1.735s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.375s
ok  	seguimiento-medicion-gpc/internal/project	(cached)
```

## Refactor

No se hizo: el código nuevo es mínimo y sigue el patrón existente (`GetItem` valida y delega en el repositorio igual que `ListItems` en `list.go`, y `GetByID` filtra y clona igual que `Save` y `ListByProject`). No hay lógica duplicada ni ramas sin test. Único cambio adicional de producción: el comentario de encabezado de `repository.go` decía "HU-05 agrega GetByID" a futuro y quedó desactualizado al implementarlo, así que se quitó esa línea.

## Validación final

```
$ gofmt -w internal/backlog/get.go internal/backlog/repository.go internal/backlog/memory_repository.go internal/backlog/get_test.go internal/backlog/get_features_test.go internal/backlog/create_test.go
(gofmt no reportó ni cambió nada: todos los archivos ya estaban formateados)
```

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
```

```
$ go vet ./...
(sin salida: sin problemas)
```

### Cobertura

```
$ go test -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.388s	coverage: 28.6% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.542s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/project	1.136s	coverage: 100.0% of statements

$ go tool cover -func=backlog-cover.out
seguimiento-medicion-gpc/internal/backlog/create.go:11:			Create			100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:55:			normalizePriority	100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:68:			cleanCriteria		100.0%
seguimiento-medicion-gpc/internal/backlog/get.go:13:			GetItem			100.0%
seguimiento-medicion-gpc/internal/backlog/list.go:8:			ListItems		100.0%
seguimiento-medicion-gpc/internal/backlog/manager.go:21:		NewManager		100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:20:	Save			100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:35:	GetByID			100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:50:	ListByProject		100.0%
total:									(statements)		100.0%
```

Sin regresiones: `GetItem` y `GetByID` entran con cobertura total (100.0% cada una) y el paquete `internal/backlog` se mantiene en 100.0%, igual que `internal/httpapi` e `internal/project`. `cmd/server` sigue en 28.6% por `main()`, que bloquea en el servidor y no se prueba con tests unitarios (ver la [evidencia de TEC-01](TEC-01-tdd.md)).

### go test -race (CA-05.4)

En este entorno (Windows, sin compilador C) `-race` no está disponible:

```
$ go test -race ./internal/backlog
go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
```

La verificación con `-race` la corre Hernán (Agile Enabler) sobre Linux y el CI de TEC-02 en cada Pull Request hacia main (`ci.yml`, paso "Correr los tests (go test -race ./...)"). El test de concurrencia (50 creaciones y 50 consultas en simultáneo, en el test unitario y en el escenario BDD) corre y pasa también sin `-race` en el entorno local.

### Ejecución de los escenarios BDD

```
$ go test -v -run TestFeatures_HU05ConsultarItem ./internal/backlog
8 scenarios (8 passed)
53 steps (53 passed)
9.4285ms
--- PASS: TestFeatures_HU05ConsultarItem (0.01s)
    --- PASS: TestFeatures_HU05ConsultarItem/Consultar_un_ítem_por_su_ID_devuelve_todos_sus_datos (0.00s)
    --- PASS: TestFeatures_HU05ConsultarItem/Con_ítems_en_varios_proyectos_se_obtiene_el_ítem_pedido_y_no_otro (0.00s)
    --- PASS: TestFeatures_HU05ConsultarItem/Modificar_el_ítem_consultado_y_sus_criterios_no_altera_el_ítem_guardado (0.00s)
    --- PASS: TestFeatures_HU05ConsultarItem/Consultar_mientras_se_crean_ítems_en_simultáneo (0.00s)
    --- PASS: TestFeatures_HU05ConsultarItem/No_se_puede_consultar_un_ítem_con_un_ID_que_no_existe (0.00s)
    --- PASS: TestFeatures_HU05ConsultarItem/No_se_puede_consultar_un_ítem_con_un_ID_que_no_existe#01 (0.00s)
    --- PASS: TestFeatures_HU05ConsultarItem/No_se_puede_consultar_un_ítem_con_un_ID_que_no_existe#02 (0.00s)
    --- PASS: TestFeatures_HU05ConsultarItem/Consultar_un_ítem_cuando_no_hay_ítems_creados (0.00s)
```

Los 8 escenarios (5 escenarios + 1 esquema con 3 ejemplos) pasan con los 53 pasos, y con `Strict: true` godog habría fallado si algún paso del `.feature` quedara sin expresión registrada: no hay pasos undefined.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | d121221 | `docs(HU-05): agregar especificación SDD` |
| RED | a9afebc | `test(HU-05): agregar tests (RED)` |
| GREEN | 34948cd | `feat(HU-05): implementar consulta de ítems del backlog (GREEN)` |

No hubo commit de refactor porque no se hizo refactor. El commit de cierre (`Closes #18`) no se lista en esta tabla porque un commit no puede contener su propio hash: queda registrado en el Pull Request y en el Issue.
