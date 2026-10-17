# Evidencia TDD — HU-07 · Modificar datos de un ítem

> Issue del repositorio: #19

- **Responsable:** Sosa Ricardo
- **Rama:** `feature/HU-07-modificar-item`
- **Fecha:** 17/10/2026
- **Go:** go1.27.1 windows/amd64 (el entorno local no tiene cgo, así que `go test -race` lo verifica el CI del Pull Request)
- **Especificación SDD:** [`docs/especificaciones/HU-07-modificar-item.md`](../especificaciones/HU-07-modificar-item.md)
- **Escenarios BDD:** [`features/HU-07-modificar-item.feature`](../../features/HU-07-modificar-item.feature) (ejecutados con godog v0.16.0)

## Trazabilidad

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

**20 funciones de test cubriendo 15 escenarios definidos (21 ejecuciones de escenario BDD)**: 19 tests unitarios (`internal/backlog/update_test.go`) y 1 test que ejecuta los escenarios BDD con godog (`internal/backlog/update_features_test.go`, runner `TestFeatures_HU07ModificarItem`). El conteo sale del comando:

```
$ findstr /R "^func.Test" internal\backlog\update_test.go internal\backlog\update_features_test.go
internal\backlog\update_test.go:func TestManagerUpdate_ModificaTodosLosCamposConservaIDProyectoEstadoYStoryPoints(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_ConservaStoryPointsYEstadoAsignados(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_CambiaSoloLasHorasReenviandoElResto(t *testing.T) {
internal\backlog\update_test.go:func TestCreate_ItemNuevoQuedaSinHorasEstimadas(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_AceptaHorasFraccionadas(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_HorasEnCeroDejanElItemSinHoras(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_RechazaHorasNegativasNaN_e_Infinitas(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_RechazaTituloVacio(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_RechazaPrioridadInvalida(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_RechazaModificacionSinCriterios(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_PrimerErrorEsElDelOrdenDeValidacion(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_SinItemsCreadosDevuelveErrItemNotFound(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_CopiaPorValorDelResultado(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_ErrorDelRepositorioAlBuscarSeDevuelve(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_ErrorDelRepositorioAlActualizarSeDevuelve(t *testing.T) {
internal\backlog\update_test.go:func TestMemoryRepositoryUpdate_ReemplazaElItemYDevuelveCopia(t *testing.T) {
internal\backlog\update_test.go:func TestMemoryRepositoryUpdate_ItemInexistenteDevuelveErrItemNotFound(t *testing.T) {
internal\backlog\update_test.go:func TestManagerUpdate_ModificacionesSimultaneasDelMismoItem(t *testing.T) {
internal\backlog\update_features_test.go:func TestFeatures_HU07ModificarItem(t *testing.T) {
```

19 de esas funciones son tests unitarios y 1 es el runner de godog. `TestManagerUpdate_RechazaHorasNegativasNaN_e_Infinitas` tiene 3 subtests (negativas, NaN, infinitas), así que las ejecuciones de tests unitarios son 22 (19 funciones + 3 subtests). `TestCreate_ItemNuevoQuedaSinHorasEstimadas` vive en `update_test.go` pero prueba el comportamiento de `Create` que fija CA-07.4 (el ítem nuevo queda sin horas).

El archivo `.feature` tiene 12 escenarios y 3 esquemas del escenario (con 3, 3 y 3 ejemplos): godog ejecuta 21 escenarios con 91 pasos, todos en verde. La correspondencia no es 1 a 1: los errores del repositorio (CA-07.7), el test de contrato de `MemoryRepository.Update`, el orden de validaciones con varios datos inválidos y la conservación de story points y estado tienen tests unitarios propios sin escenario BDD, y los esquemas multiplican sus ejemplos en ejecuciones.

Los tests y los pasos de godog son **internos al paquete** (`package backlog`), igual que en HU-04, HU-05 y HU-06. Los tests reutilizan los helpers de `create_test.go` (`newTestManager`, `validTitle`, `validPriority`, `validCriteria`, `errRepository`, `failingRepository`, que ya tenía el método `Update` agregado en el commit RED) y agregan el helper `createTestItem`.

| Criterio | Escenarios BDD | Tests unitarios | Código |
|---|---|---|---|
| CA-07.1 Se modifican título, descripción, prioridad, criterios y horas en una sola operación; ID, ProjectID, estado y story points no cambian | Modificar todos los datos de un ítem existente; Cambiar solo las horas estimadas manteniendo el resto | `TestManagerUpdate_ModificaTodosLosCamposConservaIDProyectoEstadoYStoryPoints`, `TestManagerUpdate_ConservaStoryPointsYEstadoAsignados`, `TestManagerUpdate_CambiaSoloLasHorasReenviandoElResto` | `update.go` (`Manager.Update`), `memory_repository.go` (`Update`) |
| CA-07.2 Los campos obligatorios siguen validándose; el ítem queda como estaba | Rechazar un título vacío; Rechazar un título con solo espacios; Rechazar una prioridad que no existe; Rechazar una modificación sin criterios de aceptación | `TestManagerUpdate_RechazaTituloVacio`, `TestManagerUpdate_RechazaPrioridadInvalida`, `TestManagerUpdate_RechazaModificacionSinCriterios` | `update.go` (`Manager.Update`), `create.go` (`normalizePriority`, `cleanCriteria`) |
| CA-07.3 Modificar un ítem que no existe devuelve `ErrItemNotFound` | Modificar un ítem que no existe (esquema, 3 ejemplos); Modificar cuando no hay ítems creados | `TestManagerUpdate_ItemInexistenteDevuelveErrItemNotFound` (999, 0 y -1), `TestManagerUpdate_SinItemsCreadosDevuelveErrItemNotFound`, `TestMemoryRepositoryUpdate_ItemInexistenteDevuelveErrItemNotFound` | `update.go` (`Manager.Update`), `memory_repository.go` (`Update`) |
| CA-07.4 Horas estimadas: ≥ 0, fracciones admitidas, 0 = sin horas, negativas/NaN/Inf son error | Asignar horas estimadas fraccionadas; Dejar el ítem sin horas estimadas pasando 0; Rechazar horas estimadas inválidas (esquema, 3 ejemplos) | `TestManagerUpdate_AceptaHorasFraccionadas`, `TestManagerUpdate_HorasEnCeroDejanElItemSinHoras`, `TestManagerUpdate_RechazaHorasNegativasNaN_e_Infinitas` (3 subtests), `TestCreate_ItemNuevoQuedaSinHorasEstimadas` | `update.go` (`validateEstimatedHours`), `item.go` (`EstimatedHours`) |
| CA-07.5 Copia por valor, incluida la lista de criterios | Modificar el ítem devuelto no altera el ítem guardado | `TestManagerUpdate_CopiaPorValorDelResultado`, `TestMemoryRepositoryUpdate_ReemplazaElItemYDevuelveCopia` | `update.go` (`Manager.Update`), `memory_repository.go` (`Update`) |
| CA-07.6 Concurrencia sin condiciones de carrera (con `-race`) y el resultado coincide con una versión enviada | Varias modificaciones en simultáneo del mismo ítem | `TestManagerUpdate_ModificacionesSimultaneasDelMismoItem` | `manager.go` (mutex), `update.go`, `memory_repository.go` |
| CA-07.7 Error del repositorio al buscar o actualizar se devuelve y el ítem no cambia | — (no tiene escenario BDD) | `TestManagerUpdate_ErrorDelRepositorioAlBuscarSeDevuelve`, `TestManagerUpdate_ErrorDelRepositorioAlActualizarSeDevuelve` | `update.go` (`Manager.Update`), `create_test.go` (`failingRepository`) |
| Restricción "Orden de validaciones": el primer error es el del orden (título, prioridad, criterios, horas) | — (implícito en los escenarios de error) | `TestManagerUpdate_PrimerErrorEsElDelOrdenDeValidacion` | `update.go` (`Manager.Update`) |

## Fase RED

Tests y escenarios escritos antes del código de producción (commit `7f7b091`). **El RED fue un error de compilación, no el fallo de una aserción**: `Manager.Update`, el campo `EstimatedHours` y `ErrInvalidEstimatedHours` todavía no existían, así que el paquete `internal/backlog` no compila. (El método `Update` de `failingRepository` en `create_test.go` se agregó en el mismo commit RED, como definió la SDD, para que ese archivo compilara y el fallo quedara limitado a la funcionalidad nueva.)

```
$ go test ./...
# seguimiento-medicion-gpc/internal/backlog [seguimiento-medicion-gpc/internal/backlog.test]
internal\backlog\update_features_test.go:148:21: s.manager.Update undefined (type *Manager has no field or method Update)
internal\backlog\update_features_test.go:171:31: s.manager.Update undefined (type *Manager has no field or method Update)
internal\backlog\update_features_test.go:227:27: s.updated.EstimatedHours undefined (type Item has no field or method EstimatedHours)
internal\backlog\update_features_test.go:228:72: s.updated.EstimatedHours undefined (type Item has no field or method EstimatedHours)
internal\backlog\update_features_test.go:237:15: s.updated.EstimatedHours undefined (type Item has no field or method EstimatedHours)
internal\backlog\update_features_test.go:238:83: s.updated.EstimatedHours undefined (type Item has no field or method EstimatedHours)
internal\backlog\update_features_test.go:285:23: undefined: ErrInvalidEstimatedHours
internal\backlog\update_features_test.go:321:22: item.EstimatedHours undefined (type Item has no field or method EstimatedHours)
internal\backlog\update_features_test.go:322:77: item.EstimatedHours undefined (type Item has no field or method EstimatedHours)
internal\backlog\update_features_test.go:364:4: unknown field EstimatedHours in struct literal of type Item
internal\backlog\update_features_test.go:364:4: too many errors
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
FAIL	seguimiento-medicion-gpc/internal/backlog [build failed]
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
FAIL
```

Todos los errores son `Update undefined`, `EstimatedHours undefined` y `ErrInvalidEstimatedHours undefined`: la funcionalidad nueva de HU-07. No hay fallos preexistentes: el paquete `backlog` no compila hasta que existan, así que en el RED tampoco corren los tests de HU-04, HU-05 y HU-06; el resto de los paquetes queda en verde.

## Fase GREEN

Código mínimo de producción:

- `item.go`: campo `EstimatedHours float64` en `Item` (0 = sin asignar, RN-07.5).
- `repository.go`: operación `Update(item Item) (Item, error)` en la interfaz `Repository`.
- `memory_repository.go`: `Update` reemplaza el ítem cuyo ID coincida, guarda y devuelve copias con lista de criterios propia (`slices.Clone`) y devuelve `ErrItemNotFound` si no existe.
- `update.go` (nuevo): `ErrInvalidEstimatedHours`, `Manager.Update` con el orden de validaciones de la SDD (título, prioridad, criterios y horas sin consultar el repositorio; después, con el mutex, `GetByID` y la escritura) y `validateEstimatedHours` (`math.IsNaN`/`math.IsInf`). `Update` conserva ID, ProjectID, Status y StoryPoints del ítem guardado.

No se tocó `create.go`, `get.go` ni `list.go`, ni ningún otro código de producción ni de test.

```
$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	1.992s
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
```

## Refactor

No se hizo: el código nuevo es mínimo y sigue el patrón existente (`Manager.Update` valida el contenido y delega en el repositorio igual que `Create` en `create.go`, con la normalización reutilizada en lugar de duplicada; `MemoryRepository.Update` filtra y clona igual que `Save` y `GetByID`). No hay lógica duplicada ni ramas sin test.

## Pruebas de mutación

Una por vez, restaurando el código después de cada una (ninguna forma parte de un commit):

1. **`Update` pone `StoryPoints` en 0** (`StoryPoints: 0` en lugar del valor guardado): falla `TestManagerUpdate_ConservaStoryPointsYEstadoAsignados`. Los 21 escenarios BDD siguen pasando porque ninguno verifica los story points: la conservación quedó cubierta solo por el test unitario, como se definió en la SDD (los story points los asigna HU-20). Con el código restaurado, todo vuelve a verde.
2. **Aceptar horas negativas** (se quita `hours < 0` de `validateEstimatedHours`): falla el subtest `TestManagerUpdate_RechazaHorasNegativasNaN_e_Infinitas/negativas` y el escenario BDD "Rechazar horas estimadas inválidas" (ejemplo -1; 20 de 21 escenarios pasan, 1 falla, con un paso skipped por el fallo previo). Con el código restaurado, todo vuelve a verde.
3. **Sacar el mutex de `Manager.Update`** (sin `m.mu.Lock()` ni `m.mu.Unlock()`): sin `-race`, `TestManagerUpdate_ModificacionesSimultaneasDelMismoItem` y `TestFeatures_HU07ModificarItem` pasan igual (50 modificaciones simultáneas completas y el ítem guardado coincide con una versión enviada). La mutación no se pudo verificar localmente con el detector de carreras por falta de cgo (ver abajo): la verificación de `-race` (CA-07.6) es el check del CI del Pull Request. Con el mutex restaurado, todo vuelve a verde.

## Validación final

```
$ gofmt -w internal/backlog/item.go internal/backlog/repository.go internal/backlog/memory_repository.go internal/backlog/update.go
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
ok  	seguimiento-medicion-gpc/cmd/server	(cached)	coverage: 28.6% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	2.186s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/project	(cached)	coverage: 100.0% of statements

$ go tool cover -func=cover-hu07.out
seguimiento-medicion-gpc/internal/backlog/create.go:11:			Create			100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:55:			normalizePriority	100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:68:			cleanCriteria		100.0%
seguimiento-medicion-gpc/internal/backlog/get.go:13:			GetItem			100.0%
seguimiento-medicion-gpc/internal/backlog/list.go:8:			ListItems		100.0%
seguimiento-medicion-gpc/internal/backlog/manager.go:21:		NewManager		100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:20:	Save			100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:35:	GetByID			100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:48:	Update			100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:68:	ListByProject		100.0%
seguimiento-medicion-gpc/internal/backlog/update.go:22:			Update			100.0%
seguimiento-medicion-gpc/internal/backlog/update.go:75:			validateEstimatedHours	100.0%
total:									(statements)		100.0%
```

Sin regresiones: `Manager.Update` y `validateEstimatedHours` (`update.go`) y `MemoryRepository.Update` (`memory_repository.go`) entran con cobertura total (100.0% cada una) y el paquete `internal/backlog` se mantiene en 100.0%, igual que `internal/httpapi` e `internal/project`. `cmd/server` sigue en 28.6% por `main()`, que bloquea en el servidor y no se prueba con tests unitarios (ver la [evidencia de TEC-01](TEC-01-tdd.md)).

### go test -race (CA-07.6)

En este entorno (Windows, sin compilador C) `-race` no está disponible:

```
$ go test -race ./...
go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
```

La verificación con `-race` la corre el CI de TEC-02 en cada Pull Request hacia main (`ci.yml`, paso "Correr los tests (go test -race ./...)"), igual que en HU-02, HU-04, HU-05 y HU-06. El test de concurrencia (50 modificaciones simultáneas del mismo ítem, en el test unitario y en el escenario BDD) corre y pasa también sin `-race` en el entorno local.

### Ejecución de los escenarios BDD

```
$ go test -v -run TestFeatures_HU07ModificarItem ./internal/backlog
21 scenarios (21 passed)
91 steps (91 passed)
--- PASS: TestFeatures_HU07ModificarItem
    --- PASS: TestFeatures_HU07ModificarItem/Modificar_todos_los_datos_de_un_ítem_existente
    --- PASS: TestFeatures_HU07ModificarItem/Cambiar_solo_las_horas_estimadas_manteniendo_el_resto
    --- PASS: TestFeatures_HU07ModificarItem/Aceptar_la_prioridad_sin_distinguir_mayúsculas_ni_espacios_en_los_extremos
    --- PASS: TestFeatures_HU07ModificarItem/Aceptar_la_prioridad_sin_distinguir_mayúsculas_ni_espacios_en_los_extremos#01
    --- PASS: TestFeatures_HU07ModificarItem/Aceptar_la_prioridad_sin_distinguir_mayúsculas_ni_espacios_en_los_extremos#02
    --- PASS: TestFeatures_HU07ModificarItem/Recortar_los_espacios_del_título_nuevo_en_los_extremos
    --- PASS: TestFeatures_HU07ModificarItem/Asignar_horas_estimadas_fraccionadas
    --- PASS: TestFeatures_HU07ModificarItem/Dejar_el_ítem_sin_horas_estimadas_pasando_0
    --- PASS: TestFeatures_HU07ModificarItem/Rechazar_horas_estimadas_inválidas
    --- PASS: TestFeatures_HU07ModificarItem/Rechazar_horas_estimadas_inválidas#01
    --- PASS: TestFeatures_HU07ModificarItem/Rechazar_horas_estimadas_inválidas#02
    --- PASS: TestFeatures_HU07ModificarItem/Rechazar_un_título_vacío
    --- PASS: TestFeatures_HU07ModificarItem/Rechazar_un_título_vacío#01
    --- PASS: TestFeatures_HU07ModificarItem/Rechazar_una_prioridad_que_no_existe
    --- PASS: TestFeatures_HU07ModificarItem/Rechazar_una_modificación_sin_criterios_de_aceptación
    --- PASS: TestFeatures_HU07ModificarItem/Modificar_un_ítem_que_no_existe
    --- PASS: TestFeatures_HU07ModificarItem/Modificar_un_ítem_que_no_existe#01
    --- PASS: TestFeatures_HU07ModificarItem/Modificar_un_ítem_que_no_existe#02
    --- PASS: TestFeatures_HU07ModificarItem/Modificar_cuando_no_hay_ítems_creados
    --- PASS: TestFeatures_HU07ModificarItem/Modificar_el_ítem_devuelto_no_altera_el_ítem_guardado
    --- PASS: TestFeatures_HU07ModificarItem/Varias_modificaciones_en_simultáneo_del_mismo_ítem
```

Los 21 escenarios (12 escenarios + 3 esquemas con 3 ejemplos cada uno) pasan con los 91 pasos, y con `Strict: true` godog habría fallado si algún paso del `.feature` quedara sin expresión registrada: no hay pasos undefined.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | 7f7b091 | `test(HU-07): agregar tests (RED)` |
| GREEN | 7df7e8d | `feat(HU-07): implementar modificacion de items del backlog (GREEN)` |

No hubo commit de refactor porque no se hizo refactor. La especificación SDD entró dentro del commit RED (junto con los escenarios BDD y los tests) y no en un commit propio. El commit de cierre (`Closes #19`) no se lista en esta tabla porque un commit no puede contener su propio hash: queda registrado en el Pull Request y en el Issue.
