# Evidencia TDD — HU-01 · Crear proyecto

> Issue del repositorio: #14

- **Responsable:** Peñalbé Hernán
- **Rama:** `feature/HU-01-crear-proyecto`
- **Fecha:** 10/10/2026
- **Go:** go1.27.1 linux/amd64
- **Especificación SDD:** [`docs/especificaciones/HU-01-crear-proyecto.md`](../especificaciones/HU-01-crear-proyecto.md)
- **Escenarios BDD:** [`features/HU-01-crear-proyecto.feature`](../../features/HU-01-crear-proyecto.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**16 tests cubriendo 12 escenarios BDD.** Los 16 tests son 15 tests unitarios (`internal/project/create_test.go`) y 1 test que ejecuta los escenarios BDD con godog (`internal/project/features_test.go`). De los 12 escenarios, uno es un esquema con 3 ejemplos, así que godog ejecuta 14 escenarios (57 pasos). La correspondencia no es 1 a 1: algunos criterios tienen más tests unitarios que escenarios, porque los tests cubren además casos que no están en el BDD (tabs y saltos de línea, letras acentuadas, prioridad de errores, copias del repositorio).

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-01.1 ID numérico único y correlativo | Crear un proyecto con nombre y descripción; Los proyectos reciben IDs correlativos | `TestCreate_ProyectoValidoQuedaConIDNombreDescripcionYEstadoInicial`, `TestCreate_IDsCorrelativosEmpiezanEnUno`, `TestCreate_ProyectoQuedaGuardado` | `create.go`, `memory_repository.go` (`Save`) |
| CA-01.2 Nombre obligatorio | No se puede crear un proyecto con el nombre vacío; … con un nombre de solo espacios | `TestCreate_NombreVacioDevuelveErrEmptyName` (3 casos), `TestCreate_NombreVacioSeInformaAntesQueRepetido` | `create.go`, `project.go` (`ErrEmptyName`) |
| CA-01.3 Nombre no repetido | No se puede crear un proyecto con un nombre repetido (3 ejemplos); Un nombre repetido con espacios en los extremos también se rechaza; Los nombres que solo difieren en los espacios internos son distintos | `TestCreate_NombreRepetidoDevuelveErrDuplicateName` (5 casos), `TestCreate_NombreRepetidoSinDistinguirMayusculasConLetrasAcentuadas`, `TestCreate_NombresQueDifierenEnEspaciosInternosSonDistintos` | `create.go`, `project.go` (`ErrDuplicateName`) |
| CA-01.4 Estado inicial `PLANIFICADO` | Crear un proyecto con nombre y descripción | `TestCreate_ProyectoValidoQuedaConIDNombreDescripcionYEstadoInicial`, `TestStatus_EstadosDelProyecto` | `create.go`, `project.go` (`Status`) |
| CA-01.5 Nombre sin espacios en los extremos y descripción opcional | Crear un proyecto sin descripción; El nombre se guarda sin los espacios de los extremos | `TestCreate_NombreSeGuardaSinEspaciosEnLosExtremos`, `TestCreate_DescripcionSeGuardaTalComoLlega` (2 casos) | `create.go` |
| CA-01.6 Se devuelven copias | Modificar el proyecto devuelto no altera el proyecto guardado | `TestCreate_ModificarElProyectoDevueltoNoAlteraElGuardado`, `TestMemoryRepository_ListDevuelveCopias` | `project.go`, `memory_repository.go` |
| CA-01.7 Creaciones simultáneas | Creaciones simultáneas con el mismo nombre crean un solo proyecto; … con nombres distintos reciben IDs distintos | `TestCreate_CreacionesSimultaneasConElMismoNombreCreanUnSoloProyecto`, `TestCreate_CreacionesSimultaneasConNombresDistintosTienenIDsDistintos` | `manager.go` (mutex), `create.go` |

## Fase RED

Tests y escenarios escritos antes del código. Se agregó `internal/project/doc.go` (solo la declaración del paquete) para que el fallo corresponda a las funciones que faltan y no a la ausencia del paquete.

```
$ go test ./...
# seguimiento-medicion-gpc/internal/project_test [seguimiento-medicion-gpc/internal/project.test]
internal/project/features_test.go:33:19: undefined: project.Manager
internal/project/features_test.go:34:19: undefined: project.MemoryRepository
internal/project/features_test.go:35:18: undefined: project.Project
internal/project/create_test.go:14:29: undefined: project.Manager
internal/project/create_test.go:14:47: undefined: project.MemoryRepository
internal/project/create_test.go:15:18: undefined: project.NewMemoryRepository
internal/project/create_test.go:16:17: undefined: project.NewManager
internal/project/create_test.go:19:48: undefined: project.MemoryRepository
internal/project/create_test.go:19:76: undefined: project.Project
internal/project/create_test.go:46:25: undefined: project.StatusPlanned
internal/project/create_test.go:46:25: too many errors
FAIL	seguimiento-medicion-gpc/internal/project [build failed]
FAIL
```

Todos los errores son `undefined` sobre los tipos y funciones de esta historia; no hay fallos de otro origen.

## Fase GREEN

Código mínimo: `project.go` (tipo, estados y errores), `repository.go` (interfaz), `memory_repository.go` (repositorio en memoria) y `create.go` (`Manager` y `Create`).

```
$ go test -v ./...
--- PASS: TestCreate_ProyectoValidoQuedaConIDNombreDescripcionYEstadoInicial (0.00s)
--- PASS: TestCreate_IDsCorrelativosEmpiezanEnUno (0.00s)
--- PASS: TestCreate_ProyectoQuedaGuardado (0.00s)
--- PASS: TestCreate_NombreVacioDevuelveErrEmptyName (0.00s)
--- PASS: TestCreate_NombreRepetidoDevuelveErrDuplicateName (0.00s)
--- PASS: TestCreate_NombreRepetidoSinDistinguirMayusculasConLetrasAcentuadas (0.00s)
--- PASS: TestCreate_NombresQueDifierenEnEspaciosInternosSonDistintos (0.00s)
--- PASS: TestCreate_NombreVacioSeInformaAntesQueRepetido (0.00s)
--- PASS: TestCreate_NombreSeGuardaSinEspaciosEnLosExtremos (0.00s)
--- PASS: TestCreate_DescripcionSeGuardaTalComoLlega (0.00s)
--- PASS: TestCreate_ModificarElProyectoDevueltoNoAlteraElGuardado (0.00s)
--- PASS: TestMemoryRepository_ListDevuelveCopias (0.00s)
--- PASS: TestCreate_CreacionesSimultaneasConElMismoNombreCreanUnSoloProyecto (0.00s)
--- PASS: TestCreate_CreacionesSimultaneasConNombresDistintosTienenIDsDistintos (0.00s)
--- PASS: TestStatus_EstadosDelProyecto (0.00s)
14 scenarios (14 passed)
57 steps (57 passed)
--- PASS: TestFeatures_HU01CrearProyecto (0.02s)
PASS
ok  	seguimiento-medicion-gpc/internal/project	0.030s
```

## Refactor

El `Manager` (tipo, mutex y `NewManager`) pasó de `create.go` a su propio archivo `manager.go`, porque HU-02 y HU-03 le agregan operaciones y `create.go` queda solo para la creación. Se actualizó la lista de archivos en la SDD. No cambia el comportamiento: los tests siguen en verde.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/internal/project	0.038s
```

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)

$ go test ./...
ok  	seguimiento-medicion-gpc/internal/project	0.038s

$ go test -cover ./...
ok  	seguimiento-medicion-gpc/internal/project	0.033s	coverage: 95.7% of statements

$ go test -race ./...
ok  	seguimiento-medicion-gpc/internal/project	1.143s
```

### Verificación del mutex (CA-01.7)

Para comprobar que los tests de concurrencia detectan la falta de protección, se quitó el mutex de `Create` de forma temporal y se corrió `go test -race ./...`: fallaron los 4 tests de concurrencia (los 2 unitarios y los 2 escenarios BDD) y el detector informó 20 `DATA RACE`. Con el mutex restaurado, `go test -race ./...` pasa sin advertencias. El cambio temporal no se commiteó.

### Cobertura por función

```
$ go tool cover -func=cover.out
internal/project/create.go:6:            Create               93.8%
internal/project/manager.go:15:          NewManager           100.0%
internal/project/memory_repository.go:13: NewMemoryRepository 100.0%
internal/project/memory_repository.go:18: Save                100.0%
internal/project/memory_repository.go:26: List                100.0%
total:                                   (statements)         95.7%
```

La única rama sin cubrir es la que devuelve el error de `List()` del repositorio: el repositorio en memoria nunca falla. Queda para TEC-05, cuando el repositorio sobre PostgreSQL pueda devolver errores.

## Acciones pendientes

- Rama del error de `List()` sin cubrir; se cubre en TEC-05.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | fbb73e8 | `test(HU-01): agregar tests (RED)` |
| GREEN | fc0df93 | `feat(HU-01): implementar creación de proyectos (GREEN)` |
| Refactor | 12ac547 | `refactor(HU-01): mover el Manager a manager.go` |
| Evidencia y cierre | (este commit) | `docs(HU-01): agregar evidencia TDD y cobertura` |
