# TEC-11: mapa de tests por historia y responsable

> Issue del repositorio: #179

Este mapa dice, para cada archivo de tests, a qué historias pertenecen sus tests y quién los comenta. Es el primer paso de TEC-11: hasta que esté integrado en `main`, nadie empieza a comentar.

## Qué hay que hacer

Cada test lleva, arriba de `func Test…`, una línea con su historia, el criterio de aceptación, la regla de negocio o la condición de error que prueba, y qué verifica:

```go
// HU-03 · CA-03.2 · Verifica que consultar un ID mayor al último asignado devuelva ErrProjectNotFound.
func TestGet_IDMayorAlUltimoAsignadoDevuelveErrProjectNotFound(t *testing.T) {

// HU-03 · RN-03.4 · Verifica que consultar un proyecto no cambie ningún dato guardado.
func TestGet_…(t *testing.T) {

// HU-03 · Error · Verifica que, si el repositorio falla al buscar, Get devuelva ese error.
func TestGet_ErrorDelRepositorioSeDevuelve(t *testing.T) {
```

(Son ejemplos de formato: el primero y el tercero son tests reales de `internal/project/get_test.go`; el segundo, el formato para una regla. El criterio o la regla de cada test sale de la SDD de su historia.)

- **`CA-XX.N`** cuando el test prueba un criterio de aceptación de la SDD.
- **`RN-XX.N`** cuando prueba una regla de negocio que no es un criterio. Las 43 SDD numeran sus reglas así, en la sección "Reglas de negocio" (por ejemplo, HU-03 dice que la "solo lectura" quedó como RN-03.4 porque no se verifica como criterio).
- **`Error`** cuando prueba una condición de error que la SDD describe sin número (en "Condiciones de error", "Contrato" o una sección parecida), por ejemplo, que se devuelva el error del repositorio cuando falla. Se usa solo si no hay un CA ni una RN que la cubra.

El orden para elegir es: primero un `CA`, si no hay, una `RN`, y si no, `Error`.

**Buena parte ya está hecha:** 644 de los 1013 tests sin la historia ya citan su criterio (por ejemplo, `// CA-03.1`). En esos solo falta agregar la historia adelante y qué verifica.

Reglas:

- Solo se agregan comentarios. No se cambia el código, el nombre ni el orden de los tests.
- No se modifican las SDD de `docs/especificaciones/`: se leen para sacar el criterio (CA-XX.N), la regla (RN-XX.N) o la condición de error.
- Si un test ya tiene un comentario, se agrega la historia y el criterio a ese comentario; no se borra lo que dice.
- Si un test no encaja con ningún criterio, regla ni condición de error de su SDD, no se inventa: se pregunta al autor de la historia o se anota en el Pull Request.
- Los tests de tareas técnicas (TEC-01, TEC-04 y TEC-05) llevan el ID de la tarea en lugar de la historia, con su criterio: `// TEC-04 · CA-TEC04.9 · Verifica que…`. Los criterios están en su evidencia (`docs/evidencias/TEC-XX-tdd.md`), y varios tests ya los citan.
- Después de comentar, `go test ./...` tiene que seguir en verde.
- Cada uno trabaja en su rama (`feature/TEC-11-<apellido>`) y abre su Pull Request con `Refs #179`. Así nadie edita el mismo archivo que otro.

## El test de control (PR de cierre)

> Ya está implementado en `internal/trazabilidad/comentarios_test.go`; la evidencia es [`TEC-11.md`](TEC-11.md). Durante el trabajo se acordaron dos cambios a lo que dice esta sección: la etiqueta `BDD` para los runners de godog y que la especificación de una historia nueva puede estar en `openspec/` (ver abajo).

Cuando estén integrados los cinco Pull Request, el de cierre (`Closes #179`) agrega un test que corre con `go test ./...` y en la CI, y falla si:

- un test no tiene arriba la línea `HU-XX · CA-XX.N · …`, `HU-XX · RN-XX.N · …`, `HU-XX · Error · …`, `HU-XX · BDD · …` o `TEC-XX · CA-TECXX.N · …`;
- la `HU-XX` no tiene su SDD en `docs/especificaciones/HU-XX-*.md` ni, si se especificó con OpenSpec (TEC-10), una especificación en `openspec/` que la nombre;
- un test con `BDD` es de una historia sin `features/HU-XX-*.feature`;
- el `CA-XX.N` o la `RN-XX.N` no aparece en esa SDD (por ejemplo, porque se escribió mal o no existe);
- el número del criterio o la regla no es de esa historia (por ejemplo, `HU-03 · CA-05.1`);
- la `TEC-XX` no tiene su evidencia en `docs/evidencias/TEC-XX-*.md`, o el `CA-TECXX.N` no aparece en ella.

Los runners de godog (`TestFeatures_*`) no prueban un criterio sino todos los escenarios del `.feature` de su historia, así que llevan `// HU-XX · BDD · Verifica que se cumplan los escenarios de features/HU-XX-….feature.` Los 43 que no la tenían se comentaron en el Pull Request de cierre.

En los tests con `Error` el control solo puede verificar que la historia tenga su SDD, porque la condición no tiene número. Esos tests se revisan en el Pull Request de cada uno.

El test de control también mira los 114 tests que ya citan su historia: si alguno no tiene este formato, lo ajusta el responsable de su archivo en su Pull Request.

Así la trazabilidad de los tests no depende de revisarlos a mano: si alguien agrega un test sin comentario o con una referencia que no existe, la CI lo marca.

## Cómo se armó

- **Tests sin comentar:** 1013 de 1127 funciones `func Test…`. Se cuenta como comentado el test que tiene `HU-XX` en las líneas de comentario de arriba.
- **Historia de cada test:** la del commit que agregó la función (`test(HU-XX): …` o `test(TEC-XX): …`), buscada en el historial completo de `main`.
- **Autor de cada historia:** el responsable que figura en las actas de Review de los Sprints 1 a 4 (HU-05 y HU-49 figuran en el texto de las Reviews de los Sprints 1 y 4). Las tareas técnicas: TEC-01 y TEC-04, Hernán; TEC-05, Cardozo.
- **Responsable de cada archivo:** uno solo por archivo, el autor con más tests en él, para que dos personas no editen el mismo archivo. En los archivos con tests de otras historias, el responsable le pregunta al autor de esa historia si tiene dudas.
- **Equilibrio de la carga:** se movieron seis archivos, marcados con ⇄ en las tablas (ver abajo).

## Resumen por responsable

| Integrante | Tests de sus historias | Por archivo (uno por archivo) | Después de equilibrar | Además en el Sprint |
|---|---|---|---|---|
| Buttini | 152 | 119 | **214** | La presentación |
| Cardozo | 244 | 244 | **223** | Apoyo en HU-51 |
| García | 175 | 162 | **183** | HU-50 (empieza TEC-11 después) |
| Hernán | 179 | 203 | **203** | Mapa, aprobación de propuestas, TEC-10 y TEC-09 |
| Sosa | 263 | 285 | **190** | HU-51 |
| **Total** | **1013** | **1013** | **1013** | |

**Archivos movidos para equilibrar:**

- `internal/httpapi/web_test.go` (54 tests): de Sosa a Buttini.
- `internal/backlog/update_test.go` (18 tests): de Sosa a Buttini.
- `internal/backlog/status_test.go` (14 tests): de Sosa a Buttini.
- `internal/backlog/get_test.go` (9 tests): de Sosa a Buttini.
- `internal/report/project_test.go` (13 tests): de Cardozo a García.
- `internal/httpapi/reports_test.go` (8 tests): de Cardozo a García.

## Archivos por responsable

### Buttini (214 tests)

| Archivo | Historias (tests sin comentar) | Tests |
|---|---|---|
| `internal/httpapi/web_test.go` ⇄ | HU-14 (1), HU-15 (2), HU-17 (2), HU-18 (3), HU-19 (4), HU-20 (2), HU-21 (1), HU-26 (1), HU-27 (3), HU-28 (2), HU-29 (3), HU-30 (3), HU-31 (3), HU-36 (3), HU-37 (5), HU-38 (3), HU-39 (3), HU-40 (2), HU-41 (3), HU-49 (3), TEC-04 (2) | 54 |
| `internal/sprint/create_test.go` | HU-10 (19) | 19 |
| `internal/backlog/update_test.go` ⇄ | HU-07 (18) | 18 |
| `internal/sprint/assign_test.go` | HU-11 (17) | 17 |
| `internal/backlog/status_test.go` ⇄ | HU-08 (14) | 14 |
| `internal/dashboard/summary_test.go` | HU-36 (13) | 13 |
| `internal/metrics/sprints_test.go` | HU-37 (10) | 10 |
| `internal/backlog/get_test.go` ⇄ | HU-05 (9) | 9 |
| `internal/member/list_test.go` | HU-17 (9) | 9 |
| `internal/backlog/list_test.go` | HU-06 (8) | 8 |
| `internal/poker/create_test.go` | HU-21 (8) | 8 |
| `internal/poker/vote_test.go` | HU-22 (8) | 8 |
| `internal/backlog/assign_test.go` | HU-11 (7) | 7 |
| `internal/sprint/history_test.go` | HU-14 (7) | 7 |
| `internal/httpapi/dashboard_test.go` | HU-36 (5) | 5 |
| `internal/httpapi/members_list_test.go` | HU-17 (4) | 4 |
| `internal/poker/postgres_repository_test.go` | HU-21 (2) | 2 |
| `internal/poker/repository_contract_test.go` | HU-21 (1) | 1 |
| `internal/poker/vote_features_test.go` | HU-22 (1) | 1 |

### Cardozo (223 tests)

| Archivo | Historias (tests sin comentar) | Tests |
|---|---|---|
| `internal/httpapi/poker_test.go` | HU-21 (7), HU-22 (4), HU-23 (8), HU-24 (11), HU-25 (10) | 40 |
| `internal/member/update_test.go` | HU-16 (20) | 20 |
| `internal/poker/estimate_test.go` | HU-25 (17) | 17 |
| `internal/metrics/summary_test.go` | HU-35 (16) | 16 |
| `internal/backlog/story_points_test.go` | HU-20 (15) | 15 |
| `internal/poker/round_test.go` | HU-24 (15) | 15 |
| `internal/project/update_test.go` | HU-02 (15) | 15 |
| `internal/backlog/priority_test.go` | HU-09 (14) | 14 |
| `cmd/server/main_test.go` | TEC-01 (2), TEC-04 (1), TEC-05 (10) | 13 |
| `internal/httpapi/members_update_test.go` | HU-16 (12) | 12 |
| `internal/poker/reveal_test.go` | HU-23 (12) | 12 |
| `internal/postgrestest/postgrestest_test.go` | TEC-05 (12) | 12 |
| `internal/poker/compare_test.go` | HU-24 (6) | 6 |
| `internal/project/postgres_repository_test.go` | HU-41 (2), TEC-05 (2) | 4 |
| `internal/backlog/postgres_repository_test.go` | TEC-05 (2) | 2 |
| `internal/member/repository_update_contract_test.go` | HU-16 (2) | 2 |
| `internal/postgres/postgres_test.go` | TEC-05 (2) | 2 |
| `internal/project/repository_contract_test.go` | HU-41 (1), TEC-05 (1) | 2 |
| `internal/sprint/postgres_repository_test.go` | TEC-05 (2) | 2 |
| `internal/backlog/repository_contract_test.go` | TEC-05 (1) | 1 |
| `internal/sprint/repository_contract_test.go` | TEC-05 (1) | 1 |

### García (183 tests)

| Archivo | Historias (tests sin comentar) | Tests |
|---|---|---|
| `internal/sprint/complete_test.go` | HU-12 (20) | 20 |
| `internal/backlog/create_test.go` | HU-04 (17) | 17 |
| `internal/effort/register_test.go` | HU-26 (17) | 17 |
| `internal/metrics/hours_test.go` | HU-33 (15) | 15 |
| `internal/effort/list_test.go` | HU-27 (14) | 14 |
| `internal/report/project_test.go` ⇄ | HU-38 (13) | 13 |
| `internal/sprint/close_test.go` | HU-13 (13) | 13 |
| `internal/effort/correct_test.go` | HU-28 (11) | 11 |
| `internal/httpapi/efforts_test.go` | HU-26 (11) | 11 |
| `internal/report/sprint_test.go` | HU-39 (11) | 11 |
| `internal/httpapi/reports_test.go` ⇄ | HU-38 (6), HU-49 (2) | 8 |
| `internal/httpapi/sprint_reports_test.go` | HU-39 (6), HU-49 (2) | 8 |
| `internal/effort/postgres_repository_test.go` | HU-26 (5), HU-27 (2) | 7 |
| `internal/httpapi/effort_lists_test.go` | HU-27 (6) | 6 |
| `internal/httpapi/efforts_correct_test.go` | HU-28 (6) | 6 |
| `internal/sprint/with_open_test.go` | HU-26 (5) | 5 |
| `internal/effort/repository_contract_test.go` | HU-26 (1) | 1 |

### Hernán (203 tests)

| Archivo | Historias (tests sin comentar) | Tests |
|---|---|---|
| `internal/httpapi/metrics_test.go` | HU-32 (6), HU-33 (6), HU-34 (3), HU-35 (7), HU-37 (5), HU-47 (3) | 30 |
| `internal/httpapi/server_test.go` | HU-18 (10), TEC-01 (16), TEC-04 (2) | 28 |
| `internal/project/create_test.go` | HU-01 (18) | 18 |
| `internal/httpapi/sprints_test.go` | HU-14 (5), TEC-04 (10) | 15 |
| `internal/metrics/completion_test.go` | HU-34 (15) | 15 |
| `internal/metrics/defects_test.go` | HU-47 (15) | 15 |
| `internal/httpapi/items_test.go` | HU-20 (7), TEC-04 (7) | 14 |
| `internal/member/register_test.go` | HU-15 (14) | 14 |
| `internal/metrics/story_points_test.go` | HU-32 (12) | 12 |
| `internal/report/pdf_test.go` | HU-40 (11), HU-49 (1) | 12 |
| `internal/report/stories_test.go` | HU-49 (10) | 10 |
| `internal/httpapi/members_test.go` | HU-15 (6) | 6 |
| `internal/httpapi/report_pdf_test.go` | HU-40 (4) | 4 |
| `internal/sprint/list_test.go` | HU-32 (4) | 4 |
| `internal/member/postgres_repository_test.go` | HU-15 (3) | 3 |
| `internal/metrics/points_test.go` | HU-32 (2) | 2 |
| `internal/member/memory_repository_test.go` | HU-15 (1) | 1 |

### Sosa (190 tests)

| Archivo | Historias (tests sin comentar) | Tests |
|---|---|---|
| `internal/defect/status_test.go` | HU-30 (28) | 28 |
| `internal/defect/list_test.go` | HU-31 (22) | 22 |
| `internal/httpapi/defects_test.go` | HU-29 (12), HU-30 (10) | 22 |
| `internal/project/dates_test.go` | HU-19 (18) | 18 |
| `internal/project/end_date_test.go` | HU-41 (18) | 18 |
| `internal/defect/register_test.go` | HU-29 (16) | 16 |
| `internal/project/start_date_test.go` | HU-18 (14) | 14 |
| `internal/httpapi/project_end_date_test.go` | HU-41 (13) | 13 |
| `internal/httpapi/project_dates_test.go` | HU-19 (11) | 11 |
| `internal/project/get_test.go` | HU-03 (9) | 9 |
| `internal/defect/postgres_repository_test.go` | HU-29 (5), HU-30 (2), HU-31 (1) | 8 |
| `internal/httpapi/defect_lists_test.go` | HU-31 (8) | 8 |
| `internal/defect/repository_contract_test.go` | HU-29 (1), HU-30 (1), HU-31 (1) | 3 |
