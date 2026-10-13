# Evidencia TDD — HU-06 · Listar el backlog de un proyecto

> Issue del repositorio: #21

- **Responsable:** Buttini Cristobal
- **Rama:** `feature/HU-06-listar-backlog`
- **Fecha:** 12/10/2026
- **Go:** go1.27.1 windows/amd64
- **Especificación SDD:** [`docs/especificaciones/HU-06-listar-backlog.md`](../especificaciones/HU-06-listar-backlog.md)
- **Escenarios BDD:** [`features/HU-06-listar-backlog.feature`](../../features/HU-06-listar-backlog.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**9 funciones de test cubriendo 9 escenarios BDD**: 8 tests unitarios (`internal/backlog/list_test.go`) y 1 test que ejecuta los escenarios BDD con godog (`internal/backlog/list_features_test.go`, runner `TestFeatures_HU06ListarBacklog`). De los 9 escenarios, 1 es un esquema con 2 ejemplos, así que godog ejecuta 10 escenarios (50 pasos). La correspondencia no es 1 a 1: el error del repositorio (CA-06.7) no tiene escenario BDD propio, y dos tests unitarios refuerzan con casos extra lo que cubren los escenarios (aislamiento por proyecto con datos completos, y el listado final ordenado después de la concurrencia). El test de proyecto inexistente tiene además 3 subtests (ID 999, 0 y -1).

Los tests y los pasos de godog son **internos al paquete** (`package backlog`), igual que en HU-04. Los tests reutilizan los helpers de `create_test.go` (`newTestManager`, `createTestProject`, `validTitle`, `validPriority`, `validCriteria`, `errRepository`) y extienden `failingRepository` con el campo `listErr` y el método `ListByProject`, como anticipó la SDD.

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-06.1 Se listan todos los ítems del proyecto, con todos sus datos, ordenados por ID ascendente | Listar el backlog de un proyecto con ítems; Un proyecto finalizado lista su backlog igual | `TestListItems_DevuelveSoloLosItemsDelProyectoOrdenadosPorIDConTodosSusDatos`, `TestListItems_ProyectoFinalizadoListaSusItems` | `list.go` (`ListItems`), `memory_repository.go` (`ListByProject`) |
| CA-06.2 Un proyecto sin ítems devuelve una lista vacía (longitud 0, no `nil`), no un error | Listar el backlog de un proyecto sin ítems | `TestListItems_ProyectoSinItemsDevuelveListaVaciaNoNil` | `memory_repository.go` (`ListByProject`) |
| CA-06.3 No aparecen ítems de otros proyectos | No aparecen los ítems de otros proyectos aunque tengan el mismo título | `TestListItems_DevuelveSoloLosItemsDelProyectoOrdenadosPorIDConTodosSusDatos` | `memory_repository.go` (`ListByProject`) |
| CA-06.4 Proyecto inexistente devuelve `ErrProjectNotFound` (incluidos 0, negativo y sin proyectos) | Listar el backlog de un proyecto que no existe; Listar el backlog cuando no hay proyectos creados; Listar el backlog con un ID imposible (esquema, 2 ejemplos) | `TestListItems_ProyectoInexistenteDevuelveErrProjectNotFound` (3 subtests), `TestListItems_SinProyectosCreadosDevuelveErrProjectNotFound` | `list.go` (`ListItems`), `internal/project` (`Get`) |
| CA-06.5 Listar mientras se crean ítems en simultáneo no produce condiciones de carrera (con `-race`) | Listar mientras se crean ítems en simultáneo | `TestListItems_ListadosSimultaneosConCreaciones` | `manager.go` (mutex), `list.go`, `memory_repository.go` |
| CA-06.6 Modificar la lista devuelta, un ítem o sus criterios no altera los ítems guardados | Modificar la lista devuelta no altera los ítems guardados | `TestListItems_ModificarLaListaDevueltaNoAlteraLosItemsGuardados` | `memory_repository.go` (`ListByProject`) |
| CA-06.7 Si el repositorio falla al listar, se devuelve ese error | — (no tiene escenario BDD) | `TestListItems_ErrorDelRepositorioAlListarSeDevuelve` | `list.go`, `create_test.go` (`failingRepository`) |

## Fase RED

Tests y escenarios escritos antes del código de producción (commit 9b7a387). `go test ./...` falla únicamente en `internal/backlog`, con errores `undefined` sobre la operación que todavía no existe:

```
$ go test ./...
# seguimiento-medicion-gpc/internal/backlog [seguimiento-medicion-gpc/internal/backlog.test]
internal\backlog\list_features_test.go:140:30: s.manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_features_test.go:145:30: s.manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_features_test.go:246:30: s.manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_features_test.go:276:48: s.manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_test.go:32:22: manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_test.go:83:22: manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_test.go:108:24: manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_test.go:124:22: manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_test.go:146:22: manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_test.go:168:22: manager.ListItems undefined (type *Manager has no field or method ListItems)
internal\backlog\list_test.go:168:22: too many errors
FAIL	seguimiento-medicion-gpc/internal/backlog [build failed]
ok  	seguimiento-medicion-gpc/internal/project	(cached)
FAIL
```

Todos los errores son `ListItems undefined`: la funcionalidad nueva de HU-06. No hay fallos preexistentes: `internal/project` queda en verde y el resto del paquete `backlog` (HU-04) compila.

## Fase GREEN

Código mínimo de producción: `ListItems` en el `Manager` (nuevo archivo `list.go`), la operación `ListByProject` en la interfaz `Repository` y su implementación en `MemoryRepository`. No se tocó ningún otro código de producción.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/internal/backlog	1.444s
ok  	seguimiento-medicion-gpc/internal/project	(cached)
```

## Refactor

No se hizo: el código nuevo es mínimo (una función que valida el proyecto y delega en el repositorio, y un filtro con copia de los criterios en el repositorio en memoria), reutiliza `internal/project.Get` y no duplica lógica. Los comentarios de reglas de negocio (RN-06.x) quedaron sobre cada decisión.

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)
```

`gofmt -w` se corrió sobre los archivos Go de la historia (nuevos y tocados) sin reportar ni cambiar nada.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
```

### Cobertura

```
$ go test -cover ./...
ok  	seguimiento-medicion-gpc/internal/backlog	1.460s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/project	1.421s	coverage: 100.0% of statements

$ go tool cover -func=backlog-cover.out
seguimiento-medicion-gpc/internal/backlog/create.go:11:			Create			100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:55:			normalizePriority	100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:68:			cleanCriteria		100.0%
seguimiento-medicion-gpc/internal/backlog/list.go:8:			ListItems		100.0%
seguimiento-medicion-gpc/internal/backlog/manager.go:21:		NewManager		100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:20:	Save			100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:37:	ListByProject		100.0%
total:									(statements)		100.0%

$ go tool cover -func=project-cover.out
seguimiento-medicion-gpc/internal/project/create.go:6:			Create			100.0%
seguimiento-medicion-gpc/internal/project/get.go:10:			Get			100.0%
seguimiento-medicion-gpc/internal/project/manager.go:15:		NewManager		100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:18:	Save			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:26:	List			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:32:	GetByID			100.0%
total:									(statements)		100.0%
```

Sin regresiones: el paquete nuevo de HU-06 entra con cobertura total y HU-04 no se degradó.

### go test -race (CA-06.5)

En este entorno (Windows, sin compilador C) `-race` no está disponible:

```
$ go test -race ./internal/backlog/
go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in %PATH%
```

La verificación con `-race` la corre el CI de TEC-02 en cada Pull Request hacia main (`ci.yml`, paso "Correr los tests (go test -race ./...)"). El test de concurrencia (50 creaciones y 50 listados en simultáneo, en el test unitario y en el escenario BDD) corre y pasa también sin `-race` en el entorno local.

### Ejecución de los escenarios BDD

```
$ go test -run TestFeatures_HU06ListarBacklog -v ./internal/backlog/
10 scenarios (10 passed)
50 steps (50 passed)
```

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | e457581 | `docs(HU-06): agregar especificación SDD` |
| RED | 9b7a387 | `test(HU-06): agregar tests (RED)` |
| GREEN | 20c8dd0 | `feat(HU-06): implementar listado del backlog de un proyecto (GREEN)` |

El commit de cierre (`Closes #21`) no se lista en esta tabla porque un commit no puede contener su propio hash: queda registrado en el Pull Request y en el Issue.
