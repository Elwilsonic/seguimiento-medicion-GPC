# Evidencia TDD — HU-03 · Consultar estado del proyecto

> Issue del repositorio: #16

- **Responsable:** Sosa Ricardo
- **Rama:** `feature/HU-03-consultar-proyecto`
- **Fecha:** 11/10/2026
- **Go:** go1.27.1 windows/amd64
- **Especificación SDD:** [`docs/especificaciones/HU-03-consultar-proyecto.md`](../especificaciones/HU-03-consultar-proyecto.md)
- **Escenarios BDD:** [`features/HU-03-consultar-proyecto.feature`](../../features/HU-03-consultar-proyecto.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**11 funciones de test cubriendo 9 escenarios BDD**: 10 tests unitarios (`internal/project/get_test.go`) y 1 test que ejecuta los escenarios BDD con godog (`internal/project/get_features_test.go`, runner `TestFeatures_HU03ConsultarProyecto`). De los 9 escenarios, uno es un esquema con 2 ejemplos, así que godog ejecuta 10 escenarios (40 pasos). La correspondencia no es exactamente 1 a 1: hay un test unitario sin escenario BDD propio (`TestGet_ErrorDelRepositorioSeDevuelve`, que cubre la condición de error del contrato) y un test unitario cubre los 2 ejemplos del esquema (subtests "cero" y "negativo").

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-03.1 Se obtiene el proyecto por su identificador, con nombre, descripción y estado | Consultar un proyecto existente por su ID; Consultar un proyecto que no es el primero; Consultar un proyecto con descripción vacía; Consultar un proyecto en un estado distinto del inicial | `TestGet_ProyectoExistenteDevuelveNombreDescripcionYEstado`, `TestGet_ProyectoQueNoEsElPrimero`, `TestGet_ProyectoConDescripcionVacia`, `TestGet_ProyectoEnEstadoDistintoDelInicial` | `get.go` (`Get`), `memory_repository.go` (`GetByID`) |
| CA-03.2 Consultar un proyecto que no existe devuelve `ErrProjectNotFound` | Consultar un ID mayor al último asignado; Consultar cuando no hay proyectos creados; Consultar un ID imposible (2 ejemplos) | `TestGet_IDMayorAlUltimoAsignadoDevuelveErrProjectNotFound`, `TestGet_SinProyectosCreadosDevuelveErrProjectNotFound`, `TestGet_IDCeroYNegativoDevuelvenErrProjectNotFound` (2 subtests) | `get.go` (`ErrProjectNotFound`), `memory_repository.go` (`GetByID`) |
| CA-03.3 Modificar el resultado devuelto no altera el proyecto guardado | Modificar el proyecto devuelto no altera el proyecto guardado | `TestGet_ModificarElResultadoDevueltoNoAlteraElGuardado` | `memory_repository.go` (`GetByID` devuelve copia), `get.go` |
| CA-03.4 Consultar mientras se crean proyectos en simultáneo no produce accesos concurrentes sin protección | Consultar mientras se crean proyectos en simultáneo | `TestGet_CreacionesYConsultasSimultaneasSinAccesosSinProteger` | `manager.go` (mutex), `get.go` |
| Condición de error: falla del repositorio al buscar | — (no tiene escenario BDD) | `TestGet_ErrorDelRepositorioSeDevuelve` | `get.go`, `create_test.go` (`failingRepository`) |

El escenario de estado distinto del inicial usa el repositorio falso de prueba `fixedStatusRepository` (definido en `get_test.go`), porque HU-02 — que cambia estados — todavía no existe. La ampliación de la interfaz `Repository` obligó a actualizar `failingRepository` (campo `getErr` y método `GetByID`) ya en el RED, según la decisión de la SDD.

## Fase RED

Tests y escenarios escritos antes del código de producción. Se actualizó `failingRepository` en `create_test.go` (código de test) para que siga cumpliendo la interfaz ampliada; los fallos corresponden solo a los símbolos de producción que faltan.

```
$ go test ./...
# seguimiento-medicion-gpc/internal/project_test [seguimiento-medicion-gpc/internal/project.test]
internal\project\get_features_test.go:101:27: s.manager.Get undefined (type *project.Manager has no field or method Get)
internal\project\get_features_test.go:143:31: undefined: project.ErrProjectNotFound
internal\project\get_features_test.go:174:22: s.manager.Get undefined (type *project.Manager has no field or method Get)
internal\project\get_features_test.go:199:32: s.manager.Get undefined (type *project.Manager has no field or method Get)
internal\project\get_features_test.go:208:44: undefined: project.ErrProjectNotFound
internal\project\get_features_test.go:240:23: s.manager.Get undefined (type *project.Manager has no field or method Get)
internal\project\get_test.go:39:36: undefined: project.ErrProjectNotFound
internal\project\get_test.go:49:20: manager.Get undefined (type *project.Manager has no field or method Get)
internal\project\get_test.go:78:20: manager.Get undefined (type *project.Manager has no field or method Get)
internal\project\get_test.go:95:20: manager.Get undefined (type *project.Manager has no field or method Get)
internal\project\get_test.go:95:20: too many errors
FAIL	seguimiento-medicion-gpc/internal/project [build failed]
FAIL
```

Todos los errores son `undefined` sobre `Get` y `ErrProjectNotFound`, los dos símbolos de producción que faltan; no hay fallos de otro origen. El paquete `project_test` completo deja de compilar porque sus tests (incluidos los de HU-01) dependen de los símbolos aún no definidos: comportamiento esperado del RED en Go.

## Fase GREEN

Código mínimo de producción: `get.go` (`ErrProjectNotFound` y `Manager.Get` con el mutex), la operación `GetByID` en la interfaz `Repository` (`repository.go`) y su implementación en `memory_repository.go`.

```
$ go test -v ./...
--- PASS: TestGet_ProyectoExistenteDevuelveNombreDescripcionYEstado (0.00s)
--- PASS: TestGet_ProyectoQueNoEsElPrimero (0.00s)
--- PASS: TestGet_ProyectoConDescripcionVacia (0.00s)
--- PASS: TestGet_ProyectoEnEstadoDistintoDelInicial (0.00s)
--- PASS: TestGet_IDMayorAlUltimoAsignadoDevuelveErrProjectNotFound (0.00s)
--- PASS: TestGet_SinProyectosCreadosDevuelveErrProjectNotFound (0.00s)
--- PASS: TestGet_IDCeroYNegativoDevuelvenErrProjectNotFound (0.00s)
    --- PASS: TestGet_IDCeroYNegativoDevuelvenErrProjectNotFound/cero (0.00s)
    --- PASS: TestGet_IDCeroYNegativoDevuelvenErrProjectNotFound/negativo (0.00s)
--- PASS: TestGet_ModificarElResultadoDevueltoNoAlteraElGuardado (0.00s)
--- PASS: TestGet_ErrorDelRepositorioSeDevuelve (0.00s)
--- PASS: TestGet_CreacionesYConsultasSimultaneasSinAccesosSinProteger (0.00s)
10 scenarios (10 passed)
40 steps (40 passed)
--- PASS: TestFeatures_HU03ConsultarProyecto (0.01s)
PASS
ok  	seguimiento-medicion-gpc/internal/project	1.033s
```

## Refactor

No se hizo: el código agregado es mínimo y no hay duplicación nueva en producción (`get.go` es un error y un método que delega en el repositorio con el mutex; `GetByID` es una búsqueda lineal adecuada para el Sprint 1). Los tests unitarios y los pasos godog comparten el helper `newManager` y el tipo `fixedStatusRepository` entre sí, sin duplicar código de HU-01.

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)

$ go test ./...
ok  	seguimiento-medicion-gpc/internal/project	(cached)

$ go test -cover ./...
ok  	seguimiento-medicion-gpc/internal/project	0.887s	coverage: 100.0% of statements
```

### go test -race (CA-03.4)

En el entorno donde se escribió la historia (Windows sin compilador C) `-race` no está disponible. La corrió Peñalbé Hernán sobre el commit fb24076, en Linux con go1.27.1, y la corre el CI del Pull Request.

```
$ go test -race ./...
ok  	seguimiento-medicion-gpc/internal/project	1.164s

$ GOMAXPROCS=1 go test -race ./...
ok  	seguimiento-medicion-gpc/internal/project	1.106s
```

### Verificación del mutex (CA-03.4)

Se quitó el mutex de `Get` de forma temporal (sin commitear) y se corrieron los tests:

```
$ go test ./...            (sin -race)
ok  	seguimiento-medicion-gpc/internal/project	0.034s

$ go test -race ./...
--- FAIL: TestGet_CreacionesYConsultasSimultaneasSinAccesosSinProteger (0.00s)
--- FAIL: TestFeatures_HU03ConsultarProyecto (0.04s)
WARNING: DATA RACE (4 veces)
FAIL	seguimiento-medicion-gpc/internal/project	0.175s

$ GOMAXPROCS=1 go test -race ./...
--- FAIL: TestGet_CreacionesYConsultasSimultaneasSinAccesosSinProteger (0.00s)
--- FAIL: TestFeatures_HU03ConsultarProyecto (0.04s)
WARNING: DATA RACE (4 veces)
FAIL	seguimiento-medicion-gpc/internal/project	0.124s
```

El detector reporta la lectura en `MemoryRepository.GetByID` (desde `Manager.Get`) concurrente con la escritura en `MemoryRepository.Save` (desde `Manager.Create`). Sin `-race` los tests pasan igual: por eso CA-03.4 se verifica con `-race`, que es como lo corre el CI. Con el mutex restaurado, `go test -race` pasa sin advertencias.

### Cobertura por función

```
$ go tool cover -func=cover.out
seguimiento-medicion-gpc/internal/project/create.go:6:			Create			100.0%
seguimiento-medicion-gpc/internal/project/get.go:10:			Get			100.0%
seguimiento-medicion-gpc/internal/project/manager.go:15:		NewManager		100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:18:	Save			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:26:	List			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:32:	GetByID			100.0%
total:									(statements)		100.0%
```

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | 3315d1c | `test(HU-03): agregar tests (RED)` |
| GREEN | fb24076 | `feat(HU-03): implementar consulta de proyecto por ID (GREEN)` |
