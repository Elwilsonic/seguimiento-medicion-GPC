# Evidencia TDD — HU-02 · Modificar datos del proyecto

> Issue del repositorio: #15

- **Responsable:** Cardozo Leandro
- **Rama:** `feature/HU-02-modificar-proyecto`
- **Fecha:** 13/10/2026
- **Go:** go1.27.1 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` lo verifica el CI y Hernán sobre Linux)
- **Especificación SDD:** [`docs/especificaciones/HU-02-modificar-proyecto.md`](../especificaciones/HU-02-modificar-proyecto.md)
- **Escenarios BDD:** [`features/HU-02-modificar-proyecto.feature`](../../features/HU-02-modificar-proyecto.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**15 tests unitarios y 1 suite godog cubriendo 16 escenarios BDD definidos** (godog ejecuta 19: los 14 escenarios simples más "Pasar a cada uno de los estados válidos" con 2 ejemplos y "Modificar un proyecto que no existe" con 3 ejemplos; en total 83 pasos). Los 15 tests unitarios están en `internal/project/update_test.go`, varios con subtests (19 subtests: 28 ejecuciones). La correspondencia no es 1 a 1: algunos criterios tienen más tests unitarios que escenarios, porque los tests cubren además casos que no están en el BDD (tabs y saltos de línea, más variantes de normalización del estado, los tres puntos de falla del repositorio y el contrato de `MemoryRepository.Update`).

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-02.1 Cambiar nombre, descripción y estado en una sola operación | Modificar el nombre y la descripción de un proyecto existente; Cambiar el estado de PLANIFICADO a EN_CURSO manteniendo el resto; Recortar los espacios del nombre nuevo en los extremos; Dejar la descripción vacía | `TestUpdate_NombreDescripcionYEstadoSeActualizan`, `TestUpdate_CambiarSoloElEstadoMantieneNombreYDescripcion`, `TestUpdate_NombreNuevoSeGuardaSinEspaciosEnLosExtremos`, `TestUpdate_DescripcionNuevaVaciaSeGuarda` | `update.go`, `memory_repository.go` (`Update`) |
| CA-02.2 Estado válido, normalizado, sin transiciones prohibidas | Pasar a cada uno de los estados válidos (2 ejemplos); Aceptar el estado escrito en minúsculas; Aceptar el estado con espacios en los extremos; Rechazar un estado que no existe | `TestUpdate_TodosLosEstadosSonAlcanzables`, `TestUpdate_EstadoSeAceptaSinMayusculasNiEspaciosEnLosExtremos` (4 casos), `TestUpdate_EstadoInvalidoDevuelveErrInvalidStatusYNoCambia` (2 casos) | `update.go` (`normalizeStatus`, `ErrInvalidStatus`) |
| CA-02.3 Nombre vacío ni estado inválido; el proyecto queda como estaba | Rechazar un nombre vacío; Rechazar un nombre con solo espacios; Rechazar un estado que no existe | `TestUpdate_NombreVacioOSoloEspaciosDevuelveErrEmptyNameYNoCambia` (3 casos), `TestUpdate_EstadoInvalidoDevuelveErrInvalidStatusYNoCambia` (2 casos) | `update.go` |
| CA-02.4 Proyecto inexistente | Modificar un proyecto que no existe (3 ejemplos); Modificar cuando no hay proyectos creados | `TestUpdate_ProyectoInexistenteDevuelveErrProjectNotFound` (4 casos) | `update.go`, `memory_repository.go` |
| CA-02.5 Nombre único, excluyendo el propio proyecto | Rechazar un nombre repetido con otro proyecto; Dejar el nombre del propio proyecto con otras mayúsculas | `TestUpdate_NombreRepetidoConOtroProyectoDevuelveErrDuplicateNameYNoCambia` (3 casos), `TestUpdate_MismoNombreDelPropioProyectoEsValido` | `update.go` |
| CA-02.6 Copia por valor | Modificar el proyecto devuelto no altera el proyecto guardado | `TestUpdate_ModificarElResultadoDevueltoNoAlteraElGuardado` | `update.go`, `memory_repository.go` |
| CA-02.7 Concurrencia | Modificaciones y creaciones en simultáneo | `TestUpdate_CreacionesYModificacionesSimultaneasSinAccesosSinProteger` | `manager.go` (mutex), `update.go` |
| CA-02.8 Error del repositorio | — (no tiene escenario BDD) | `TestUpdate_ErrorDelRepositorioSeDevuelveYElProyectoNoCambia` (3 casos), `TestMemoryRepository_UpdateProyectoInexistenteDevuelveErrProjectNotFound` | `update.go`, `memory_repository.go` (`Update`) |

## Fase RED

Tests, escenarios y pasos godog escritos antes del código (commit 7ea611c, junto con el método `Update` de los dos repositorios falsos de test, como define la SDD). El fallo corresponde solo a la funcionalidad nueva:

```
$ go test ./...
# seguimiento-medicion-gpc/internal/project_test [seguimiento-medicion-gpc/internal/project.test]
internal\project\update_features_test.go:102:22: s.manager.Update undefined (type *project.Manager has no field or method Update)
internal\project\update_features_test.go:116:31: s.manager.Update undefined (type *project.Manager has no field or method Update)
internal\project\update_features_test.go:165:31: undefined: project.ErrInvalidStatus
internal\project\update_features_test.go:225:31: s.manager.Update undefined (type *project.Manager has no field or method Update)
internal\project\update_test.go:21:20: manager.Update undefined (type *project.Manager has no field or method Update)
internal\project\update_test.go:47:20: manager.Update undefined (type *project.Manager has no field or method Update)
internal\project\update_test.go:72:21: manager.Update undefined (type *project.Manager has no field or method Update)
internal\project\update_test.go:97:22: manager.Update undefined (type *project.Manager has no field or method Update)
internal\project\update_test.go:116:20: manager.Update undefined (type *project.Manager has no field or method Update)
internal\project\update_test.go:147:22: manager.Update undefined (type *project.Manager has no field or method Update)
internal\project\update_test.go:147:22: too many errors
ok  	seguimiento-medicion-gpc/internal/backlog	1.678s
FAIL	seguimiento-medicion-gpc/internal/project [build failed]
FAIL
```

Todos los errores son `undefined` sobre `Update` y `ErrInvalidStatus`; `internal/backlog` sigue en `ok`, sin fallos preexistentes.

## Fase GREEN

Código mínimo: `update.go` (`Manager.Update`, `ErrInvalidStatus` y `normalizeStatus`), `Update` en la interfaz `Repository` y en `MemoryRepository` (commit 7fd2514). El orden de validaciones de la SDD quedó así: nombre y estado antes de `m.mu.Lock()`; existencia, duplicados y escritura bajo el mutex.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/internal/backlog	0.497s
ok  	seguimiento-medicion-gpc/internal/project	0.570s
```

El suite BDD corrió en modo verbose para verificar que no quedara ningún paso sin definir:

```
$ go test -count=1 -v -run TestFeatures_HU02ModificarProyecto ./internal/project
19 scenarios (19 passed)
83 steps (83 passed)
--- PASS: TestFeatures_HU02ModificarProyecto (0.04s)
ok  	seguimiento-medicion-gpc/internal/project	0.389s
```

## Refactor

No aplica. El equipo decidió en la SDD no refactorizar los tests de HU-03 (`fixedStatusRepository` queda como está, con el método `Update` agregado solo para seguir cumpliendo la interfaz), y el código de producción no quedó con ninguna mejora pendiente.

## Validación final

```
$ gofmt -l .
internal\project\update_test.go

$ gofmt -w internal/project/update_test.go
$ gofmt -l .
(sin salida: todos los archivos están formateados)

$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/internal/backlog	0.277s
ok  	seguimiento-medicion-gpc/internal/project	0.315s

$ go test -count=1 -cover ./...
ok  	seguimiento-medicion-gpc/internal/backlog	1.747s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/project	1.824s	coverage: 98.4% of statements

$ go test -race -count=1 ./internal/project
go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
```

`gofmt` detectó la alineación de un map en `update_test.go`; se corrigió con el commit c399df2 (cambio solo de espacios). `-race` no corre en el entorno local porque requiere cgo: **lo verifica el CI (TEC-02) en el Pull Request y Hernán sobre Linux cuando revise la rama**.

### Cobertura por función

Al terminar el GREEN quedaba una rama sin cubrir (ver "Desviación del ciclo TDD"). Con el test de contrato del repositorio, `internal/project` vuelve al 100%:

```
$ go tool cover -func=project-cover.out
seguimiento-medicion-gpc/internal/project/create.go:6:			Create			100.0%
seguimiento-medicion-gpc/internal/project/get.go:10:			Get			100.0%
seguimiento-medicion-gpc/internal/project/manager.go:15:		NewManager		100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:18:	Save			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:26:	List			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:32:	GetByID			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:43:	Update			100.0%
seguimiento-medicion-gpc/internal/project/update.go:19:			Update			100.0%
seguimiento-medicion-gpc/internal/project/update.go:64:			normalizeStatus		100.0%
total:									(statements)		100.0%
```

### Prueba de mutación (test de contrato de `MemoryRepository.Update`)

Para demostrar que el test nuevo protege la rama, se mutó `memory_repository.go` a propósito (la rama sin proyecto pasó a devolver `p, nil` en lugar de `ErrProjectNotFound`), se corrió el test, falló, y se restauró el código con `git restore`:

```
== Con el mutante (la rama no devuelve ErrProjectNotFound)
$ go test -count=1 -run TestMemoryRepository_UpdateProyectoInexistenteDevuelveErrProjectNotFound ./internal/project
--- FAIL: TestMemoryRepository_UpdateProyectoInexistenteDevuelveErrProjectNotFound (0.00s)
    update_test.go:415: error = <nil>, se esperaba ErrProjectNotFound
FAIL
FAIL	seguimiento-medicion-gpc/internal/project	1.674s
FAIL

== Con el código restaurado
$ git restore internal/project/memory_repository.go
$ go test -count=1 -run TestMemoryRepository_UpdateProyectoInexistenteDevuelveErrProjectNotFound ./internal/project
ok  	seguimiento-medicion-gpc/internal/project	0.961s

== Estado después de la prueba
$ git status --short
 M internal/project/update_test.go
$ git diff --stat
 internal/project/update_test.go | 14 ++++++++++++++
 1 file changed, 14 insertions(+)
```

El código de producción quedó sin cambios: la única modificación es el test nuevo (+14 líneas en `update_test.go`).

### Desviación del ciclo TDD

En el commit GREEN (7fd2514) se escribió la rama `ErrProjectNotFound` de `MemoryRepository.Update` sin un test que la pidiera. `AGENTS.md` indica que no se agrega código no cubierto por un test, así que es una desviación del ciclo. La detectó la revisión de `go tool cover -func` de la validación final: `MemoryRepository.Update` quedaba en 80.0% y `internal/project` en 98.4%. La rama no es alcanzable a través del `Manager` (con el mutex tomado, el `GetByID` previo ya encontró el proyecto), por eso ningún test del GREEN la cubría.

Como el código ya existía, el test que la cubre (commit test(HU-02): cubrir el error de MemoryRepository.Update sin proyecto) pasa desde el primer momento: no hay una fase RED real y no se fingió una. Para demostrar que el test protege el comportamiento, se hizo una prueba de mutación (salida arriba): el test falla contra el mutante y pasa con el código restaurado, y el código de producción no quedó modificado.

Acción para la Retrospectiva del Sprint 1 (#65): revisar `go tool cover -func` antes de cada commit GREEN, para no escribir ramas de manejo de errores sin un test que las pida.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | 16b0d5f | `docs(HU-02): agregar especificación SDD` |
| RED | 7ea611c | `test(HU-02): agregar tests (RED)` |
| GREEN | 7fd2514 | `feat(HU-02): implementar modificación de nombre, descripción y estado (GREEN)` |
| Formato | c399df2 | `chore(HU-02): aplicar gofmt a update_test.go` |
| Test (ajuste) | 710a14d | `test(HU-02): cubrir el error de MemoryRepository.Update sin proyecto` |

El commit de evidencia no se lista en la tabla, porque un commit no puede contener su propio hash. No hay commit de cierre: `Closes #15` va en la descripción del Pull Request, y el Issue se cierra solo cuando el merge llega a `main`.
