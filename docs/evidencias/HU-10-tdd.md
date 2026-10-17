# Evidencia TDD — HU-10 · Crear sprint con nombre, fechas y Sprint Goal

> Issue del repositorio: #24

- **Responsable:** Buttini Cristobal
- **Rama:** `feature/HU-10-crear-sprint`
- **Fecha:** 17/10/2026
- **Go:** go1.27.1 windows/amd64
- **Especificación SDD:** [`docs/especificaciones/HU-10-crear-sprint.md`](../especificaciones/HU-10-crear-sprint.md)
- **Escenarios BDD:** [`features/HU-10-crear-sprint.feature`](../../features/HU-10-crear-sprint.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**20 funciones de test cubriendo 14 escenarios BDD**: 19 tests unitarios (`internal/sprint/create_test.go`) y 1 test que ejecuta los escenarios BDD con godog (`internal/sprint/create_features_test.go`, runner `TestFeatures_HU10CrearSprint`). De los 14 escenarios, 2 son esquemas con 4 y 2 ejemplos, así que godog ejecuta 18 escenarios (71 pasos). La correspondencia no es 1 a 1: el error del repositorio (CA-10.8) no tiene escenario BDD propio, y varios tests unitarios refuerzan con casos extra lo que cubren los escenarios (las creaciones simultáneas en proyectos distintos, el orden de validación con 6 subtests, el nombre repetido con el sprint anterior cerrado y las fechas ausentes con su propio error). Algunos tests usan subtests: nombre y sprint goal con 3 casos cada uno, fin no posterior con 2 y proyecto inexistente con 3.

Los tests y los pasos de godog son **internos al paquete** (`package sprint`), igual que en HU-04. Como el paquete es nuevo, definen sus propios helpers (`newTestManager`, `createTestProject`, `savedSprint`, `validStart`/`validEnd` y `failingRepository` con `openErr` y `saveErr`). Lo guardado se verifica mirando el repositorio en memoria directamente, porque HU-10 no define operaciones de lectura (HU-11 a HU-14 las agregan).

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-10.1 Se crea un sprint con nombre, fechas y Sprint Goal en un proyecto existente, con ID correlativo, estado `ABIERTO` y nombre/goal sin espacios en los extremos | Crear un sprint en un proyecto existente; El nombre y el sprint goal se guardan sin los espacios de los extremos | `TestCreate_SprintValidoQuedaConID1EstadoAbiertoYCamposSinEspacios`, `TestCreate_IDsCorrelativosYGlobalesEntreProyectos` | `create.go` (`Create`), `memory_repository.go` (`Save`) |
| CA-10.2 La fecha de fin es estrictamente posterior a la de inicio; las fechas ausentes son error | Crear un sprint con fechas inválidas (esquema, 4 ejemplos) | `TestCreate_FechaInicioAusenteDevuelveErrMissingStartDate`, `TestCreate_FechaFinAusenteDevuelveErrMissingEndDate`, `TestCreate_FinNoPosteriorAlInicioDevuelveErrEndNotAfterStart` (2 subtests), `TestCreate_InformaElPrimerErrorSegunElOrdenDeValidacion` (4 subtests de fechas) | `create.go` (`Create`) |
| CA-10.3 No puede haber más de un sprint `ABIERTO` por proyecto | Ya existe un sprint abierto en el proyecto | `TestCreate_SegundoSprintAbiertoEnElMismoProyectoDevuelveErrOpenSprintExists`, `TestCreate_SprintAbiertoEnOtroProyectoNoImpide`, `TestCreate_NombreRepetidoEnElMismoProyectoConElAnteriorCerradoSePermite` | `create.go` (`Create`), `memory_repository.go` (`OpenByProject`) |
| CA-10.4 Nombre o Sprint Goal vacíos devuelven error | El nombre del sprint es obligatorio; El sprint goal es obligatorio | `TestCreate_NombreVacioDevuelveErrEmptyName` (3 subtests), `TestCreate_SprintGoalVacioDevuelveErrEmptyGoal` (3 subtests) | `create.go` (`Create`), `sprint.go` (errores) |
| CA-10.5 Proyecto inexistente devuelve `ErrProjectNotFound` (incluidos 0, negativo y sin proyectos) | Crear un sprint en un proyecto que no existe; Crear un sprint cuando no hay proyectos creados; Crear un sprint con un ID imposible (esquema, 2 ejemplos) | `TestCreate_ProyectoInexistenteDevuelveErrProjectNotFound` (3 subtests), `TestCreate_SinProyectosCreadosDevuelveErrProjectNotFound` | `create.go` (`Create`), `internal/project` (`Get`) |
| CA-10.6 De varias creaciones simultáneas en el mismo proyecto, exactamente una tiene éxito (con `-race`) | Creaciones simultáneas en el mismo proyecto dejan un solo sprint abierto | `TestCreate_CreacionesSimultaneasEnElMismoProyectoDejanUnSoloSprintAbierto`, `TestCreate_CreacionesSimultaneasEnProyectosDistintosSeCreanTodas` | `manager.go` (mutex), `create.go`, `memory_repository.go` |
| CA-10.7 Modificar el sprint devuelto no altera el guardado | Modificar el sprint devuelto no altera el sprint guardado | `TestCreate_ModificarElSprintDevueltoNoAlteraElGuardado` | `sprint.go` (solo campos valor, RN-10.8) |
| CA-10.8 Si el repositorio falla, se devuelve ese error | — (no tiene escenario BDD) | `TestCreate_ErrorDelRepositorioAlGuardarSeDevuelve`, `TestCreate_ErrorDelRepositorioAlConsultarElSprintAbiertoSeDevuelve` | `create.go` (`Create`), `create_test.go` (`failingRepository`) |

El caso del proyecto `FINALIZADO` (RN-10.4) tiene su escenario BDD ("Un proyecto finalizado acepta sprints") y su test unitario `TestCreate_ProyectoFinalizadoAceptaSprints`, que acompañan a CA-10.1.

## Fase RED

Tests y escenarios escritos antes de cualquier código de producción (commit dab3b2e). `go test ./...` falla únicamente en `internal/sprint`, con errores `undefined` sobre los tipos y operaciones que todavía no existen:

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.199s
# seguimiento-medicion-gpc/internal/sprint [seguimiento-medicion-gpc/internal/sprint.test]
internal\sprint\create_features_test.go:36:15: undefined: MemoryRepository
internal\sprint\create_features_test.go:37:15: undefined: Manager
internal\sprint\create_features_test.go:40:17: undefined: Sprint
internal\sprint\create_features_test.go:41:19: undefined: Sprint
internal\sprint\create_features_test.go:122:76: undefined: Status
internal\sprint\create_features_test.go:274:47: undefined: Sprint
internal\sprint\create_test.go:29:25: undefined: Manager
internal\sprint\create_test.go:29:35: undefined: MemoryRepository
internal\sprint\create_test.go:48:38: undefined: MemoryRepository
internal\sprint\create_test.go:48:64: undefined: Sprint
internal\sprint\create_test.go:48:64: too many errors
ok  	seguimiento-medicion-gpc/internal/backlog	2.163s
ok  	seguimiento-medicion-gpc/internal/httpapi	2.362s
ok  	seguimiento-medicion-gpc/internal/project	2.702s
FAIL	seguimiento-medicion-gpc/internal/sprint [build failed]
FAIL
```

Todos los errores son `undefined` sobre la funcionalidad nueva de HU-10. No hay fallos preexistentes: el resto de los paquetes queda en verde. Sin desviaciones del ciclo TDD: el RED se escribió y se verificó antes de escribir el código.

## Fase GREEN

Código mínimo de producción, en el paquete nuevo `internal/sprint` (patrón de HU-04): `doc.go`, `sprint.go` (tipo `Sprint`, estados y errores), `repository.go` (interfaz con `Save` y `OpenByProject`), `memory_repository.go`, `manager.go` y `create.go`. No se tocó ningún otro código de producción.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/sprint	1.675s
```

## Refactor

No se hizo: el código nuevo es mínimo (una función que valida y delega en el repositorio, y un repositorio en memoria con dos operaciones), reutiliza `internal/project.Get` y no duplica lógica. Los comentarios de reglas de negocio (RN-10.x) quedaron sobre cada decisión.

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)
```

`gofmt -w` se corrió sobre los archivos Go de la historia (nuevos y tocados) sin reportar ni cambiar nada. `go vet ./...` tampoco informa problemas.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/sprint	(cached)
```

### Cobertura

```
$ go test -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)	coverage: 28.6% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/project	(cached)	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	(cached)	coverage: 100.0% of statements

$ go tool cover -func=sprint-cover.out
seguimiento-medicion-gpc/internal/sprint/create.go:14:			Create			100.0%
seguimiento-medicion-gpc/internal/sprint/manager.go:22:		NewManager		100.0%
seguimiento-medicion-gpc/internal/sprint/memory_repository.go:11:	NewMemoryRepository		100.0%
seguimiento-medicion-gpc/internal/sprint/memory_repository.go:18:	Save			100.0%
seguimiento-medicion-gpc/internal/sprint/memory_repository.go:28:	OpenByProject		100.0%
total:									(statements)	100.0%
```

Sin regresiones: el paquete nuevo de HU-10 entra con cobertura total y los demás paquetes no se degradan. El resultado también está en [`docs/reportes/cobertura.md`](../reportes/cobertura.md).

### go test -race (CA-10.6)

En este entorno (Windows, sin compilador C) `-race` no está disponible:

```
$ go test -race ./internal/sprint/
go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
```

La verificación con `-race` la corre el CI de TEC-02 en cada Pull Request hacia main (`ci.yml`, paso "Correr los tests (go test -race ./...)"). Los tests de concurrencia (50 creaciones simultáneas en el mismo proyecto, que dejan un solo sprint abierto, y 50 en proyectos distintos, en los tests unitarios; 10 en el escenario BDD) corren y pasan también sin `-race` en el entorno local. La revisión de la rama con `-race` sobre Linux queda a cargo de Peñalbé Hernán, igual que en HU-02 y HU-05.

### Ejecución de los escenarios BDD

```
$ go test -run TestFeatures_HU10CrearSprint -v ./internal/sprint/
18 scenarios (18 passed)
71 steps (71 passed)
```

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | f0f83f8 | `docs(HU-10): agregar especificación SDD` |
| RED | dab3b2e | `test(HU-10): agregar tests (RED)` |
| GREEN | 726248f | `feat(HU-10): implementar creación de sprint (GREEN)` |

No hubo refactor. El commit de cierre (`Closes #24`) no se lista en esta tabla porque un commit no puede contener su propio hash: queda registrado en el Pull Request y en el Issue.
