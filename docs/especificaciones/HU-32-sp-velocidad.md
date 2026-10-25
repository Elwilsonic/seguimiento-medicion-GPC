# HU-32 · Calcular Story Points planificados, completados y velocidad

> Issue del repositorio: #50

Aprobada por: Peñalbé Hernán (Agile Enabler), 25/10/2026, con las dos diferencias con el acta (sprint y división) registradas en "Comparación con el acta de Planning".

## Historia de usuario

Como Agile Enabler, quiero calcular los Story Points planificados y completados de un sprint y la velocidad del equipo, para saber cuánto trabajo se comprometió y cuánto se terminó.

## Objetivo

Calcular las métricas M1 (Story Points planificados), M2 (Story Points completados) y M3 (velocidad del equipo) de `docs/metricas.md` (TEC-03), por sprint y por proyecto, en un paquete nuevo `internal/metrics`. Exponerlas en dos endpoints de consulta y en la interfaz web (tema 5 del Planning del Sprint 3). HU-33 y HU-34 agregan sus métricas a los mismos endpoints, y HU-35 los completa como consulta única.

## Alcance

- **Incluye:** el cálculo de M1 y M2 de un sprint y de un proyecto, la cantidad de ítems sin estimar, M3 del proyecto, la lectura de los sprints de un proyecto (nueva en `internal/sprint`, en memoria y en PostgreSQL), `GET /api/sprints/{id}/metricas`, `GET /api/proyectos/{id}/metricas` y la sección "Métricas" de la interfaz web.
- **No incluye:** horas y desviación (HU-33), porcentaje de historias completadas y defectos (HU-34), la consulta única con "no disponible" para todas las métricas (HU-35), gráficos (HU-36 y HU-37), reportes (HU-38 y HU-39), congelar los Story Points al cerrar un sprint (ver "Límite conocido") ni listar sprints anteriores en la interfaz (HU-14).

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `SprintStoryPoints(sprintID int) (SprintPoints, error)` | ID de un sprint | M1 y M2 del sprint |
| `ProjectStoryPoints(projectID int) (ProjectPoints, error)` | ID de un proyecto | M1 y M2 del proyecto y M3 |

Las dos están en el `Manager` de `internal/metrics`.

### Salidas esperadas

- **`SprintPoints`** (por valor): `SprintID`, `Planned` (M1), `Completed` (M2) y `Unestimated` (cantidad de ítems del sprint con `StoryPoints = 0`).
- **`ProjectPoints`** (por valor): `ProjectID`, `Planned` (M1), `Completed` (M2), `Unestimated`, `Velocity` (M3, `float64` sin redondear), `VelocityAvailable` (`false` si no hay sprints cerrados) y `ClosedSprints` (cantidad de sprints cerrados usados para M3: 0 a 3).
- **Error:** el que indica el motivo (ver "Condiciones de error").

### Lectura de sprints (en `internal/sprint`)

- `Repository.ListByProject(projectID int) ([]Sprint, error)`: devuelve una lista nueva con copias de los sprints del proyecto, ordenada por ID ascendente; lista vacía (no `nil`) si no tiene. En memoria y en PostgreSQL, con su test de contrato.
- `Manager.Get(id int) (Sprint, error)` y `Manager.ListByProject(projectID int) ([]Sprint, error)`: leen bajo el mutex del `Manager` y devuelven copias. No cambian el comportamiento existente.

### HTTP (TEC-04, tema 5 del Planning)

`GET /api/sprints/{id}/metricas`:

```json
{"sprint_id": 3, "sp_planificados": 10, "sp_completados": 8, "items_sin_estimar": 1}
```

`GET /api/proyectos/{id}/metricas`:

```json
{"proyecto_id": 1, "sp_planificados": 34, "sp_completados": 21, "items_sin_estimar": 2, "velocidad": 10.0, "sprints_cerrados": 3}
```

- `velocidad` se redondea a un decimal (TEC-03) y es `null` si no hay sprints cerrados.

| Caso | Código | Mensaje |
|---|---|---|
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| Sprint inexistente | 404 | El de `sprint.ErrSprintNotFound` |
| Proyecto inexistente | 404 | El de `project.ErrProjectNotFound` |
| Falla de un repositorio | 500 | El de `errInternal`, sin detalle |

### Interfaz web (TEC-04)

Una sección **"Métricas"** con el botón "Ver métricas", que muestra las del proyecto actual (SP planificados, SP completados, ítems sin estimar y velocidad, o "sin datos") y, si hay sprint actual, las del sprint. Un error se muestra con el mismo mecanismo que las demás acciones.

## Reglas de negocio

1. **RN-32.1 Planificados (M1).** Suma de `StoryPoints` de los ítems del conjunto: en un sprint, los ítems con `SprintID` igual al sprint; en un proyecto, todos los ítems del backlog, asignados o no (decisión de TEC-03).
2. **RN-32.2 Completados (M2).** Suma de `StoryPoints` de los ítems del conjunto en estado `TERMINADO`.
3. **RN-32.3 Ítems sin estimar.** Un ítem con `StoryPoints = 0` no suma en M1 ni en M2 y se cuenta en `Unestimated`.
4. **RN-32.4 Sin ítems.** Un sprint o un proyecto sin ítems devuelve M1 = 0, M2 = 0 y `Unestimated` = 0, sin error.
5. **RN-32.5 Velocidad (M3).** Promedio de M2 de los últimos 3 sprints `CERRADO` del proyecto, ordenados por ID descendente, o de todos los cerrados si hay menos de 3. Un sprint cerrado sin ítems cuenta con M2 = 0. Los sprints abiertos no cuentan.
6. **RN-32.6 Velocidad no disponible.** Sin sprints cerrados, `VelocityAvailable` es `false`, `Velocity` es 0 y `ClosedSprints` es 0. No es un error.
7. **RN-32.7 Redondeo.** El cálculo no redondea; la API redondea `velocidad` a un decimal y la interfaz la muestra con un decimal. M1 y M2 son enteros.
8. **RN-32.8 Datos actuales.** Las métricas se calculan en el momento de la consulta con los datos guardados. Consultar no modifica nada.

## Decisión sobre entidades referenciadas

- **Sprint:** sí se valida su existencia, con `sprint.Manager.Get` (`ErrSprintNotFound`, incluidos 0 y negativos): sin sprint no hay conjunto de ítems y la API tiene que distinguir un sprint inexistente de uno vacío.
- **Proyecto:** sí se valida, a través de `backlog.Manager.ListItems`, que ya devuelve `ErrProjectNotFound` (HU-06); no se agrega otra consulta.
- **Ítems:** no se validan uno por uno: se leen con `ListItems` del proyecto del sprint o del proyecto consultado.

## Restricciones

- **Organización del código.** `internal/metrics`: `doc.go`, `points.go` (tipos y cálculo puro de M1, M2 y M3, sin repositorios), `manager.go` (`NewManager(items *backlog.Manager, sprints *sprint.Manager)`) y `story_points.go` (las dos operaciones). En `internal/sprint`: `list.go` y la operación en los dos repositorios. En `internal/httpapi`: `metrics.go` y las rutas en `server.go`. En `internal/web`: la sección y la llamada. `httpapi.NewHandler` crea el `metrics.Manager` a partir de `items` y `sprints`, porque `metrics` no tiene repositorio ni estado: `NewHandler` no cambia de firma y `cmd/server` no se modifica (ajuste aprobado el 25/10, antes del commit RED).
- **Sin persistencia propia.** `internal/metrics` no tiene repositorio: lee de `backlog` y `sprint`. No hay tablas nuevas.
- **Dependencias.** `metrics` importa `backlog` y `sprint`; ninguno importa `metrics`.
- **Concurrencia.** Las lecturas de `sprint` y `backlog` van cada una bajo su propio mutex; `metrics` no tiene estado y no toma ningún mutex. Las dos lecturas no son atómicas entre sí (ver "Límite conocido").
- **Copias por valor.** Todo se devuelve por valor; `ListByProject` devuelve una lista nueva.
- **Tests.** Unitarios en `internal/metrics/points_test.go` y `story_points_test.go`; godog en `story_points_features_test.go` con el runner `TestFeatures_HU32CalcularStoryPoints`, sobre `features/HU-32-sp-velocidad.feature`, con el CA en el nombre de cada escenario; contrato de `ListByProject` en `internal/sprint/repository_contract_test.go` (memoria y PostgreSQL); API en `internal/httpapi/metrics_test.go` (un caso exitoso y uno por código); interfaz en `web_test.go`. El RED compila con stubs; la creación del `metrics.Manager` en `NewHandler` va en el GREEN.
- **Profundidad proporcional (3 SP).** Un test por regla y por código de error; sin pruebas de mutación.
- **Librerías.** Solo la librería estándar; godog en los tests.

### Límite conocido

Si después de cerrar un sprint se reestima un ítem que no se terminó (RN-20.3) o se reabre uno terminado (HU-08), cambian M1 o M2 de ese sprint cerrado y, por M2, la velocidad. No se congelan los Story Points al cerrar: haría falta cambiar HU-13 y el esquema (decisión del 25/10). Además, la lectura de los sprints y la de los ítems no son atómicas entre sí: un cambio simultáneo puede quedar reflejado en una y no en la otra.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Ejemplo de TEC-03: ítems A (5, `TERMINADO`), B (3, `TERMINADO`), C (2, `EN_CURSO`), D (0, `PENDIENTE`) | M1 = 10, M2 = 8, sin estimar = 1 |
| Sprint sin ítems | 0, 0 y 0 |
| Ítem `TERMINADO` sin Story Points | No suma en M2; cuenta como sin estimar |
| Proyecto con ítems sin asignar a sprint | Suman en M1 del proyecto, no en el de ningún sprint |
| Sin sprints cerrados | Velocidad no disponible (`null` en la API), `sprints_cerrados` = 0 |
| 1 sprint cerrado con M2 = 8 | Velocidad 8.0 |
| Cerrados con M2 = 8, 10 y 12 | Velocidad 10.0 |
| 4 cerrados con M2 = 2, 8, 10 y 12 (IDs 1 a 4) | Usa los IDs 4, 3 y 2: 10.0 |
| Cerrados con M2 = 10, 15 y 12 | Velocidad 12.333…; la API devuelve 12.3 |
| Un sprint cerrado sin ítems y otro con M2 = 10 | Velocidad 5.0 |
| Un sprint abierto con M2 = 20 y ninguno cerrado | Velocidad no disponible |
| Sprint o proyecto inexistente, 0 o negativo | `ErrSprintNotFound` o `ErrProjectNotFound` |
| Falla de un repositorio | Se devuelve ese error |
| HTTP: `{id}` no numérico | 400 |

## Condiciones de error

| Condición | Error |
|---|---|
| No existe el sprint (incluidos 0 y negativos) | `sprint.ErrSprintNotFound` |
| No existe el proyecto (incluidos 0 y negativos) | `project.ErrProjectNotFound` |
| Falla un repositorio al leer | El error del repositorio |

## Criterios de aceptación

- **CA-32.1** Los Story Points planificados de un sprint son la suma de los Story Points de sus ítems; los de un proyecto, la de todo su backlog.
- **CA-32.2** Los Story Points completados son la suma de los Story Points de los ítems `TERMINADO` del sprint o del proyecto.
- **CA-32.3** Los ítems sin Story Points no suman y se informa cuántos hay.
- **CA-32.4** Un sprint o un proyecto sin ítems devuelve 0, sin error.
- **CA-32.5** La velocidad del proyecto es el promedio de los Story Points completados de los últimos 3 sprints cerrados (o de todos si hay menos), con el ejemplo de TEC-03; un sprint cerrado sin ítems cuenta con 0 y los abiertos no cuentan.
- **CA-32.6** Sin sprints cerrados, la velocidad queda como no disponible, sin error.
- **CA-32.7** Un sprint o un proyecto inexistente devuelve su error; una falla del repositorio se propaga. Consultar no modifica nada.
- **CA-32.8** Los sprints de un proyecto se leen igual en memoria y en PostgreSQL.
- **CA-32.9** Las métricas se consultan desde la API (`GET /api/sprints/{id}/metricas` y `GET /api/proyectos/{id}/metricas`, velocidad con un decimal o `null`) y desde la sección "Métricas" de la interfaz web. Se cubre con tests de `internal/httpapi`, no con escenarios godog.

## Comparación con el acta de Planning

**Hay dos diferencias con el acta del Sprint 3, que hay que confirmar antes de aprobar:**

1. **Sprint.** El acta ubica HU-32 en el Sprint 4. Se adelanta al Sprint 3 por la replanificación del 24/10 (entrega el 02/11, desarrollo cerrado el 30/10), que todavía no está registrada en el acta: va en la Daily y en la Review del Sprint 3.
2. **División.** El acta propone dividir HU-32 en HU-32 (planificados), HU-43 (completados) y HU-44 (velocidad), sujeto a la confirmación de los profesores. Esta SDD cubre las tres partes en una sola historia, como quedó en la replanificación (la división queda en pausa). Si los profesores confirman la división, esta SDD se separa por criterio: CA-32.1 → HU-32, CA-32.2 → HU-43, CA-32.5 y CA-32.6 → HU-44.

El resto coincide: TEC-03 está integrada, las fórmulas son las de `docs/metricas.md` y HU-32 bloquea a HU-35 (consulta única) y a HU-37 (dashboard).

## Comparación con los criterios del Issue #50 (borrador)

| CA del Issue (borrador) | En esta SDD |
|---|---|
| CA-32.1 Planificados de los ítems asignados al sprint | CA-32.1, más el nivel proyecto |
| CA-32.2 Completados de los ítems completados del sprint | CA-32.2 |
| CA-32.3 Velocidad con la fórmula de TEC-03 | CA-32.5 y CA-32.6 |
| CA-32.4 Sprint sin ítems devuelve cero; sin story points no se cuentan y se indica | CA-32.3 y CA-32.4 |
| — | CA-32.7 a CA-32.9 (errores, PostgreSQL, API e interfaz) |

Texto para la sección "Criterios de aceptación" del Issue #50, con la misma numeración:

```markdown
## Criterios de aceptación

- [ ] CA-32.1 Los story points planificados de un sprint son la suma de sus ítems; los de un proyecto, la de todo su backlog.
- [ ] CA-32.2 Los story points completados son la suma de los ítems terminados del sprint o del proyecto.
- [ ] CA-32.3 Los ítems sin story points no suman y se informa cuántos hay.
- [ ] CA-32.4 Un sprint o un proyecto sin ítems devuelve cero, sin error.
- [ ] CA-32.5 La velocidad es el promedio de los story points completados de los últimos 3 sprints cerrados (o de todos si hay menos).
- [ ] CA-32.6 Sin sprints cerrados, la velocidad queda como no disponible.
- [ ] CA-32.7 Un sprint o proyecto inexistente devuelve error y consultar no modifica nada.
- [ ] CA-32.8 Los sprints de un proyecto se leen igual con la base de datos.
- [ ] CA-32.9 Las métricas se ven desde la API y desde la interfaz web, con la velocidad con un decimal.
```

## Decisiones tomadas en la redacción

1. **Endpoints que se amplían** (decisión del 25/10): HU-32 crea los dos endpoints de métricas; HU-33 y HU-34 les suman campos y HU-35 los completa. Hay que avisarle a Cardozo para que la SDD de HU-35 use estos endpoints.
2. **Sin congelar Story Points al cerrar** (decisión del 25/10): límite conocido.
3. **Velocidad no disponible:** `VelocityAvailable` en el dominio y `null` en la API, a la espera de cómo lo unifique HU-35.
4. **`ListByProject` en `internal/sprint`:** hace falta para encontrar los sprints cerrados, incluidos los que no tienen ítems. Se agrega al repositorio en memoria y en PostgreSQL, sin cambiar el comportamiento existente.
5. **Story Points de la historia:** no se estimó en el Planning del Sprint 3; se propone 3 SP para el Planning del Sprint 4.
