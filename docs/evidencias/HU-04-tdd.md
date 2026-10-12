# Evidencia TDD — HU-04 · Crear ítem del backlog

> Issue del repositorio: #17

- **Responsable:** García Santiago
- **Rama:** `feature/HU-04-crear-item-backlog`
- **Fecha:** 12/10/2026
- **Go:** go1.27.1 windows/amd64
- **Especificación SDD:** [`docs/especificaciones/HU-04-crear-item-backlog.md`](../especificaciones/HU-04-crear-item-backlog.md)
- **Escenarios BDD:** [`features/HU-04-crear-item-backlog.feature`](../../features/HU-04-crear-item-backlog.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**18 funciones de test cubriendo 16 escenarios BDD**: 17 tests unitarios (`internal/backlog/create_test.go`) y 1 test que ejecuta los escenarios BDD con godog (`internal/backlog/create_features_test.go`, runner `TestFeatures_HU04CrearItem`). De los 16 escenarios, 3 son esquemas (con 2, 2 y 3 ejemplos), así que godog ejecuta 20 escenarios (95 pasos). La correspondencia no es 1 a 1: el error del repositorio (CA-04.8) no tiene escenario BDD propio, y varios tests unitarios refuerzan con casos extra lo que cubren los escenarios (copias por valor, normalización de prioridad, proyecto inexistente).

Los tests y los pasos de godog son **internos al paquete** (`package backlog`): como HU-04 no define operaciones de lectura (son de HU-05 y HU-06), la verificación de lo guardado mira el repositorio en memoria directamente, sin agregar operaciones que ningún criterio pide (decisión de la SDD).

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-04.1 Se crea un ítem con los datos en un proyecto existente | Crear un ítem con todos los campos válidos; Crear un ítem sin descripción y con espacios en los extremos; El título puede repetirse dentro del mismo proyecto | `TestCreate_ItemValidoQuedaConIDEstadoPendienteYStoryPointsEnCero`, `TestCreate_DescripcionOpcionalSeGuardaTalComoLlega`, `TestCreate_TituloRepetidoEnElMismoProyectoSePermite` | `create.go` (`Create`), `memory_repository.go` (`Save`) |
| CA-04.2 Se validan título y criterios de aceptación | No se puede crear un ítem con el título vacío; No se puede crear un ítem sin criterios de aceptación; Los criterios vacíos no cuentan como criterio de aceptación (esquema, 2 ejemplos); Los criterios vacíos se descartan y se conserva el orden; Con varios datos inválidos se informa el primer error según el orden; Los criterios se validan antes que el proyecto | `TestCreate_TituloVacioDevuelveErrEmptyTitle`, `TestCreate_SinCriteriosValidosDevuelveErrNoAcceptanceCriteria`, `TestCreate_DescartaCriteriosVaciosYConservaElOrden`, `TestCreate_InformaElPrimerErrorSegunElOrdenDeValidacion` | `create.go` (`Create`, `cleanCriteria`) |
| CA-04.3 La prioridad solo admite ALTA, MEDIA o BAJA, sin distinguir mayúsculas | Crear un ítem sin descripción y con espacios en los extremos; No se puede crear un ítem con una prioridad inválida (esquema, 2 ejemplos) | `TestCreate_PrioridadSeNormalizaAMayusculas`, `TestCreate_PrioridadInvalidaDevuelveErrInvalidPriority` | `create.go` (`normalizePriority`) |
| CA-04.4 ID único correlativo, estado PENDIENTE y story points en 0 | Crear un ítem con todos los campos válidos; Los IDs son correlativos y globales entre proyectos | `TestCreate_PrimerItemRecibeID1EstadoPendienteYStoryPointsEnCero`, `TestCreate_IDsCorrelativosYGlobalesEntreProyectos` | `create.go` (`Create`), `memory_repository.go` (`Save`) |
| CA-04.5 Proyecto inexistente devuelve `ErrProjectNotFound` | Un proyecto finalizado acepta ítems; No se puede crear un ítem en un proyecto que no existe (esquema, 3 ejemplos) | `TestCreate_ProyectoInexistenteDevuelveErrProjectNotFound`, `TestCreate_ProyectoFinalizadoAceptaItems` | `create.go` (`Create`), `internal/project` (`Get`) |
| CA-04.6 Modificar el ítem devuelto o los criterios recibidos no altera lo guardado | Modificar el ítem devuelto no altera el ítem guardado; Modificar la lista de criterios recibida no altera el ítem guardado | `TestCreate_ModificarElItemDevueltoNoAlteraElGuardado`, `TestCreate_ModificarLaListaDeCriteriosRecibidaNoAlteraElGuardado` | `memory_repository.go` (`Save`), `create.go` (`Create`) |
| CA-04.7 Creaciones simultáneas reciben IDs distintos (con `-race`) | Creaciones simultáneas con IDs distintos | `TestCreate_CreacionesSimultaneasRecibenIDsDistintos` | `manager.go` (mutex), `create.go` |
| CA-04.8 El error del repositorio al guardar se propaga | — (no tiene escenario BDD) | `TestCreate_ErrorDelRepositorioAlGuardarSeDevuelve` | `create.go`, `create_test.go` (`failingRepository`) |

## Fase RED

Tests y escenarios escritos antes del código de producción (commit 98f27ad). `go test ./...` falla únicamente en `internal/backlog`, con errores `undefined` sobre los símbolos de producción que todavía no existen:

```
$ go test ./...
# seguimiento-medicion-gpc/internal/backlog [seguimiento-medicion-gpc/internal/backlog.test]
internal\backlog\create_features_test.go:37:15: undefined: MemoryRepository
internal\backlog\create_features_test.go:38:15: undefined: Manager
internal\backlog\create_features_test.go:41:16: undefined: Item
internal\backlog\create_features_test.go:42:18: undefined: Item
internal\backlog\create_features_test.go:254:43: undefined: Item
internal\backlog\create_test.go:29:25: undefined: Manager
internal\backlog\create_test.go:29:35: undefined: MemoryRepository
internal\backlog\create_test.go:48:36: undefined: MemoryRepository
internal\backlog\create_test.go:48:62: undefined: Item
internal\backlog\create_test.go:477:38: undefined: Item
internal\backlog\create_test.go:477:38: too many errors
FAIL	seguimiento-medicion-gpc/internal/backlog [build failed]
FAIL
```

Todos los errores son `undefined` (`Manager`, `MemoryRepository`, `Item` y los demás símbolos del paquete nuevo); no hay fallos de otro origen. El paquete `internal/project` no se toca en esta historia.

## Fase GREEN

Código mínimo de producción: el paquete `internal/backlog` completo (`doc.go`, `item.go`, `repository.go`, `memory_repository.go`, `manager.go` y `create.go`). En esta fase también se corrigieron dos detalles de los pasos BDD que godog marcó como indefinidos cuando el paquete por fin compiló: la coma después del ID en el paso «creo un ítem en el proyecto con ID \<id\>, título …» y las comillas duplicadas del esquema de prioridad inválida.

```
$ (binario de test de internal/backlog, ejecutado desde la carpeta del paquete)
--- PASS: TestCreate_ItemValidoQuedaConIDEstadoPendienteYStoryPointsEnCero (0.00s)
--- PASS: TestCreate_PrimerItemRecibeID1EstadoPendienteYStoryPointsEnCero (0.00s)
--- PASS: TestCreate_DescripcionOpcionalSeGuardaTalComoLlega (0.00s)
--- PASS: TestCreate_TituloRepetidoEnElMismoProyectoSePermite (0.00s)
--- PASS: TestCreate_TituloVacioDevuelveErrEmptyTitle (0.00s)
--- PASS: TestCreate_PrioridadSeNormalizaAMayusculas (0.00s)
--- PASS: TestCreate_PrioridadInvalidaDevuelveErrInvalidPriority (0.00s)
--- PASS: TestCreate_SinCriteriosValidosDevuelveErrNoAcceptanceCriteria (0.00s)
--- PASS: TestCreate_DescartaCriteriosVaciosYConservaElOrden (0.00s)
--- PASS: TestCreate_ProyectoInexistenteDevuelveErrProjectNotFound (0.00s)
--- PASS: TestCreate_ProyectoFinalizadoAceptaItems (0.00s)
--- PASS: TestCreate_IDsCorrelativosYGlobalesEntreProyectos (0.00s)
--- PASS: TestCreate_InformaElPrimerErrorSegunElOrdenDeValidacion (0.00s)
--- PASS: TestCreate_ModificarElItemDevueltoNoAlteraElGuardado (0.00s)
--- PASS: TestCreate_ModificarLaListaDeCriteriosRecibidaNoAlteraElGuardado (0.00s)
--- PASS: TestCreate_CreacionesSimultaneasRecibenIDsDistintos (0.00s)
--- PASS: TestCreate_ErrorDelRepositorioAlGuardarSeDevuelve (0.00s)
20 scenarios (20 passed)
95 steps (95 passed)
--- PASS: TestFeatures_HU04CrearItem (0.02s)
PASS
(exit code 0)
```

Sin regresiones en `internal/project`:

```
14 scenarios (14 passed)
57 steps (57 passed)
10 scenarios (10 passed)
40 steps (40 passed)
PASS
(exit code 0)
```

## Refactor

No se hizo: el código nuevo es mínimo (una función de creación con tres validaciones, la normalización y el filtrado como funciones chicas, la interfaz con una sola operación y su implementación en memoria) y no hay duplicación con `internal/project`. Los tests comparten helpers (`newTestManager`, `savedItem`, `parseCriteria`, `equalCriteria`) sin duplicar lógica de HU-01 ni de HU-03.

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)
```

### Cobertura

```
$ go tool cover -func=backlog-cover.out
seguimiento-medicion-gpc/internal/backlog/create.go:11:			Create			100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:55:			normalizePriority	100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:68:			cleanCriteria		100.0%
seguimiento-medicion-gpc/internal/backlog/manager.go:21:		NewManager		100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:20:	Save			100.0%
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

### go test -race (CA-04.7)

En este entorno (Windows, `CGO_ENABLED=0` y sin compilador C) `-race` no está disponible:

```
$ go test -c -race ./internal/backlog
go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
```

La corre el CI de TEC-02 en cada Pull Request (workflow con `-race` sobre Linux), que es donde queda verificada CA-04.7.

### Nota sobre el entorno local (Control de aplicaciones de Windows)

En esta máquina, la política de Control de Aplicaciones de Windows (Smart App Control, `VerifiedAndReputablePolicyState = 1`) bloquea la ejecución de los binarios que `go test` compila en carpetas temporales `go-build*`:

```
$ go test ./...
fork/exec ...\.gotmp\go-build3456440024\b188\project.test.exe: An Application Control policy has blocked this file.
```

Por eso la verificación local se hizo compilando los binarios con `go test -c` y ejecutándolos desde la carpeta del paquete. El bloqueo también alcanza de forma intermitente a algunos binarios nuevos según su reputación (hash): se resolvió recompilando con `-ldflags "-s -w"` y, para el binario instrumentado con `-cover`, agregando `-tags=hu04`. El CI no tiene esta limitación: corre `go test ./...` (y `-race`) normalmente.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | 98f27ad | `test(HU-04): agregar tests (RED)` |
| GREEN | d44fcd4 | `feat(HU-04): implementar creación de ítems del backlog (GREEN)` |

El commit de cierre (`Closes #17`) no se lista en esta tabla porque un commit no puede contener su propio hash: queda registrado en el Pull Request y en el Issue.
