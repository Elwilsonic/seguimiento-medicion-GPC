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

**18 tests cubriendo 12 escenarios BDD** (después de los ajustes por la revisión del PR #68; antes eran 16). Los 18 tests son 17 tests unitarios (`internal/project/create_test.go`) y 1 test que ejecuta los escenarios BDD con godog (`internal/project/features_test.go`). De los 12 escenarios, uno es un esquema con 3 ejemplos, así que godog ejecuta 14 escenarios (57 pasos). La correspondencia no es 1 a 1: algunos criterios tienen más tests unitarios que escenarios, porque los tests cubren además casos que no están en el BDD (tabs y saltos de línea, letras acentuadas, prioridad de errores, copias del repositorio).

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-01.1 ID numérico único y correlativo | Crear un proyecto con nombre y descripción; Los proyectos reciben IDs correlativos | `TestCreate_ProyectoValidoQuedaConIDNombreDescripcionYEstadoInicial`, `TestCreate_IDsCorrelativosEmpiezanEnUno`, `TestCreate_ProyectoQuedaGuardado` | `create.go`, `memory_repository.go` (`Save`) |
| CA-01.2 Nombre obligatorio | No se puede crear un proyecto con el nombre vacío; … con un nombre de solo espacios | `TestCreate_NombreVacioDevuelveErrEmptyName` (3 casos), `TestCreate_NombreVacioSeInformaSinConsultarElRepositorio` (reemplaza a `TestCreate_NombreVacioSeInformaAntesQueRepetido`) | `create.go`, `project.go` (`ErrEmptyName`) |
| CA-01.3 Nombre no repetido | No se puede crear un proyecto con un nombre repetido (3 ejemplos); Un nombre repetido con espacios en los extremos también se rechaza; Los nombres que solo difieren en los espacios internos son distintos | `TestCreate_NombreRepetidoDevuelveErrDuplicateName` (5 casos), `TestCreate_NombreRepetidoSinDistinguirMayusculasConLetrasAcentuadas`, `TestCreate_NombresQueDifierenEnEspaciosInternosSonDistintos` | `create.go`, `project.go` (`ErrDuplicateName`) |
| CA-01.4 Estado inicial `PLANIFICADO` | Crear un proyecto con nombre y descripción | `TestCreate_ProyectoValidoQuedaConIDNombreDescripcionYEstadoInicial`, `TestStatus_EstadosDelProyecto` | `create.go`, `project.go` (`Status`) |
| CA-01.5 Nombre sin espacios en los extremos y descripción opcional | Crear un proyecto sin descripción; El nombre se guarda sin los espacios de los extremos | `TestCreate_NombreSeGuardaSinEspaciosEnLosExtremos`, `TestCreate_DescripcionSeGuardaTalComoLlega` (2 casos) | `create.go` |
| CA-01.6 Se devuelven copias | Modificar el proyecto devuelto no altera el proyecto guardado | `TestCreate_ModificarElProyectoDevueltoNoAlteraElGuardado`, `TestMemoryRepository_ListDevuelveCopias` | `project.go`, `memory_repository.go` |
| CA-01.7 Creaciones simultáneas | Creaciones simultáneas con el mismo nombre crean un solo proyecto; … con nombres distintos reciben IDs distintos | `TestCreate_CreacionesSimultaneasConElMismoNombreCreanUnSoloProyecto`, `TestCreate_CreacionesSimultaneasConNombresDistintosTienenIDsDistintos` | `manager.go` (mutex), `create.go` |
| Condición de error: falla del repositorio | — (no tiene escenario BDD) | `TestCreate_ErrorDelRepositorioAlListarSeDevuelve`, `TestCreate_ErrorDelRepositorioAlGuardarSeDevuelve` | `create.go` |

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

## Ajustes por la revisión del PR #68

Cardozo Leandro revisó el PR #68 (con ayuda de IA) y aprobó con cinco observaciones no bloqueantes. Se verificaron todas y se resolvieron en el PR de seguimiento, en la rama `feature/HU-01-ajustes-revision` (Refs #14).

| # | Observación | Resolución |
|---|---|---|
| 1 | `TestCreate_NombreVacioSeInformaAntesQueRepetido` no podía fallar | Se reemplazó por tests con un repositorio falso que falla (commit 43eac79) |
| 2 | Sin `-race`, los tests de concurrencia dependen de la cantidad de CPU | El CI corre `go test -race ./...` (commit c81f04d) |
| 3 | La SDD decía "espacios, tabs y saltos de línea", pero `strings.TrimSpace` recorta todo espacio en blanco Unicode | Se corrigió la redacción de RN-01.1 y RN-01.2 y el comentario de `create.go` |
| 4 | Faltaba HU-01 en el registro de IA | Se agregaron las filas de TEC-02, HU-01, la revisión del PR #68 y estos ajustes |
| 5 | Cobertura distinta, "hay 1 proyectos guardados" y fila "(este commit)" | La cobertura se reconfirmó con Go 1.27.1 (95.7% antes de los ajustes); el singular se corrigió con un ciclo RED/GREEN (commits 19ffa5b y fa6e9ac); la fila de cierre de la tabla de commits ya muestra el hash e79b679 en lugar de "(este commit)" (corregido en el seguimiento del Issue #73) |

### Desviación del ciclo TDD

En el commit GREEN (fc0df93) se escribió la rama que devuelve el error de `List()` del repositorio (`if err != nil { return Project{}, err }`) sin un test que la pidiera. `AGENTS.md` indica que no se agrega código no cubierto por un test, así que es una desviación del ciclo. La detectó la revisión del PR #68.

Como el código ya existía, los tests que la cubren (commit 43eac79) pasan desde el primer momento: no hay una fase RED real y no se fingió una. Para demostrar que los tests sí protegen el comportamiento, se hizo una prueba de mutación: se rompió el código a propósito, se comprobó que los tests fallan y se restauró el código. El código de `create.go` no cambió (solo su comentario).

Acción para la Retrospectiva del Sprint 1 (#65): en el GREEN no se escriben ramas de manejo de errores sin un test que las pida, y antes de abrir cada Pull Request se revisa `go tool cover -func` para que no quede código nuevo sin cubrir.

### Singular en los pasos BDD (RED → GREEN)

RED (commit 19ffa5b): el `.feature` pasa a decir "hay 1 proyecto guardado" y el paso no tiene definición.

```
$ go test ./...
14 scenarios (9 passed, 5 undefined)
57 steps (51 passed, 5 undefined, 1 skipped)
--- FAIL: TestFeatures_HU01CrearProyecto (0.02s)
    --- FAIL: .../No_se_puede_crear_un_proyecto_con_un_nombre_repetido (0.00s)
        suite.go:640: step is undefined: hay 1 proyecto guardado
    (… mismo error en los otros 4 escenarios que usan el singular)
FAIL	seguimiento-medicion-gpc/internal/project	0.039s
```

GREEN (commit fa6e9ac): la definición del paso acepta singular y plural (`^hay (\d+) proyectos? guardados?$`).

```
$ go test ./...
14 scenarios (14 passed)
57 steps (57 passed)
ok  	seguimiento-medicion-gpc/internal/project	0.041s
```

### Prueba de mutación

| Mutante (cambio a propósito en `create.go`) | Test anterior | Test nuevo |
|---|---|---|
| 1. La validación del nombre vacío se mueve después de `List()` | `TestCreate_NombreVacioSeInformaAntesQueRepetido`: **pasa** (no lo detecta) | `TestCreate_NombreVacioSeInformaSinConsultarElRepositorio`: **falla** |
| 2. Se ignora el error de `List()` | No había test | `TestCreate_ErrorDelRepositorioAlListarSeDevuelve`: **falla** |
| 3. Se ignora el error de `Save()` | No había test | `TestCreate_ErrorDelRepositorioAlGuardarSeDevuelve`: **falla** |
| 4. Sin mutex, con una sola CPU (`GOMAXPROCS=1`) | `go test ./...`: **pasa** (no lo detecta) | `go test -race ./...`: **falla** |

```
== MUTANTE 1 con el test anterior
$ go test -run TestCreate_NombreVacioSeInformaAntesQueRepetido ./...
--- PASS: TestCreate_NombreVacioSeInformaAntesQueRepetido (0.00s)
ok  	seguimiento-medicion-gpc/internal/project	0.012s

== MUTANTE 1 con el test nuevo
$ go test -run TestCreate_NombreVacioSeInformaSinConsultarElRepositorio ./...
    create_test.go:197: error = falla del repositorio, se esperaba ErrEmptyName
--- FAIL: TestCreate_NombreVacioSeInformaSinConsultarElRepositorio (0.00s)
FAIL	seguimiento-medicion-gpc/internal/project	0.010s

== MUTANTE 2: se ignora el error de List()
$ go test -run TestCreate_ErrorDelRepositorioAlListarSeDevuelve ./...
    create_test.go:209: error = <nil>, se esperaba el error del repositorio
    create_test.go:212: proyecto devuelto = {ID:1 Name:Proyecto A Description: Status:PLANIFICADO}, se esperaba un proyecto vacío
--- FAIL: TestCreate_ErrorDelRepositorioAlListarSeDevuelve (0.00s)
FAIL	seguimiento-medicion-gpc/internal/project	0.010s

== MUTANTE 3: se ignora el error de Save()
$ go test -run TestCreate_ErrorDelRepositorioAlGuardarSeDevuelve ./...
    create_test.go:224: error = <nil>, se esperaba el error del repositorio
--- FAIL: TestCreate_ErrorDelRepositorioAlGuardarSeDevuelve (0.00s)
FAIL	seguimiento-medicion-gpc/internal/project	0.010s

== MUTANTE 4: sin mutex, con GOMAXPROCS=1
$ GOMAXPROCS=1 go test ./...
ok  	seguimiento-medicion-gpc/internal/project	0.030s
$ GOMAXPROCS=1 go test -race ./...
--- FAIL: TestCreate_CreacionesSimultaneasConElMismoNombreCreanUnSoloProyecto (0.00s)
--- FAIL: TestCreate_CreacionesSimultaneasConNombresDistintosTienenIDsDistintos (0.00s)
--- FAIL: TestFeatures_HU01CrearProyecto (0.08s)
WARNING: DATA RACE (12 veces)
FAIL	seguimiento-medicion-gpc/internal/project	0.115s
```

En todos los casos el código se restauró después de la prueba; ningún mutante se commiteó.

### Validación después de los ajustes

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)

$ go test ./...
ok  	seguimiento-medicion-gpc/internal/project	0.036s

$ go test -cover ./...
ok  	seguimiento-medicion-gpc/internal/project	0.034s	coverage: 100.0% of statements

$ go test -race ./...
ok  	seguimiento-medicion-gpc/internal/project	1.123s
```

## Seguimiento del Issue #73: recorte de espacios Unicode

La revisión del PR #72 observó que ningún test fijaba que `strings.TrimSpace` recorta también el espacio no separable (`\u00a0`), como dicen RN-01.1 y RN-01.2 de la SDD. El código ya lo hacía bien: no es un bug, faltaba el test. Se agregaron:

- El caso `"solo espacio no separable"` (`"\u00a0"`) en `TestCreate_NombreVacioDevuelveErrEmptyName`: devuelve `ErrEmptyName` y no guarda nada.
- `TestCreate_NombreSeGuardaSinEspaciosNoSeparablesEnLosExtremos`: `"\u00a0Proyecto A\u00a0"` se guarda como `"Proyecto A"`.

Como el comportamiento ya existía, los tests pasan desde el primer momento: no hay una fase RED real y no se fingió una. Para demostrar que protegen el comportamiento, se hizo una prueba de mutación: en `create.go` se reemplazó `strings.TrimSpace(name)` por `strings.Trim(name, " \t\n")` (recorte solo de espacio, tab y salto de línea), se corrieron los tests y se restauró el código con `git restore`.

```
== mutante
-	name = strings.TrimSpace(name)
+	name = strings.Trim(name, " \t\n")

== tests que ya existían (sin los de #73)
$ go test -count=1 ./internal/project
ok  	seguimiento-medicion-gpc/internal/project	0.061s

== con los tests de #73
$ go test -count=1 -run 'TestCreate_NombreVacioDevuelveErrEmptyName|TestCreate_NombreSeGuardaSinEspaciosNoSeparablesEnLosExtremos' ./internal/project
--- FAIL: TestCreate_NombreVacioDevuelveErrEmptyName (0.00s)
    --- FAIL: TestCreate_NombreVacioDevuelveErrEmptyName/solo_espacio_no_separable (0.00s)
        create_test.go:100: error = <nil>, se esperaba ErrEmptyName
        create_test.go:103: hay 1 proyectos guardados, se esperaba 0
--- FAIL: TestCreate_NombreSeGuardaSinEspaciosNoSeparablesEnLosExtremos (0.00s)
    create_test.go:275: Name devuelto = "\u00a0Proyecto A\u00a0", se esperaba "Proyecto A"
    create_test.go:278: Name guardado = "\u00a0Proyecto A\u00a0", se esperaba "Proyecto A"
FAIL
FAIL	seguimiento-medicion-gpc/internal/project	0.013s

== código restaurado
$ git restore internal/project/create.go
$ git status --short
 M internal/project/create_test.go
```

Los tests que ya existían no detectaban el mutante; los dos nuevos sí. El código de producción no cambió: la única modificación es en `create_test.go`.

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)

$ go test -count=1 -race ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.015s
ok  	seguimiento-medicion-gpc/internal/backlog	1.182s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.026s
ok  	seguimiento-medicion-gpc/internal/project	1.224s

$ go test -count=1 -cover ./internal/project
ok  	seguimiento-medicion-gpc/internal/project	0.042s	coverage: 100.0% of statements
```

## Acciones pendientes

- ~~Rama del error de `List()` sin cubrir; se cubre en TEC-05.~~ Resuelta antes de TEC-05: se cubrió con un repositorio falso (ver "Ajustes por la revisión del PR #68").

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | fbb73e8 | `test(HU-01): agregar tests (RED)` |
| GREEN | fc0df93 | `feat(HU-01): implementar creación de proyectos (GREEN)` |
| Refactor | 12ac547 | `refactor(HU-01): mover el Manager a manager.go` |
| Evidencia y cierre | e79b679 | `docs(HU-01): agregar evidencia TDD y cobertura` |
| RED (ajuste) | 19ffa5b | `test(HU-01): usar el singular en los pasos BDD (RED)` |
| GREEN (ajuste) | fa6e9ac | `fix(HU-01): aceptar singular y plural en el paso de proyectos guardados (GREEN)` |
| Tests (ajuste) | 43eac79 | `test(HU-01): cubrir el error del repositorio y el orden de validaciones` |
| CI (ajuste) | c81f04d | `ci(HU-01): correr los tests con -race` |
| Tests (#73) | 9e8954c | `test(HU-01): fijar el recorte del espacio no separable en el nombre` |
