# HU-36 · Ver el resumen del estado del proyecto

> Issue del repositorio: #54

Aprobada por: Peñalbé Hernán (Agile Enabler), 28/10/2026, con dos ajustes (sprint_actual con el formato de sprintResponse, incluido proyecto_id, y las lecturas nombradas en RN-36.3 y en "Restricciones"), que se aplicaron.

## Historia de usuario

Como Agile Enabler, quiero ver un resumen del estado del proyecto, para tener una visión rápida del avance.

## Objetivo

Mostrar en una sola consulta el estado de un proyecto: su nombre y estado, la cantidad de ítems del backlog en cada estado y el sprint actual. Se implementa en un paquete nuevo `internal/dashboard`, con el endpoint `GET /api/proyectos/{id}/resumen` y una sección nueva "Dashboard" en la interfaz web, donde HU-37 agrega después los gráficos.

## Alcance

- **Incluye:** nombre y estado del proyecto; cantidad de ítems en `PENDIENTE`, `EN_CURSO` y `TERMINADO`; sprint actual (o ninguno); el endpoint y la sección "Dashboard" con el botón "Ver resumen".
- **No incluye:** métricas (Story Points, velocidad, porcentaje, horas y defectos, de E9 y HU-35), gráficos (HU-37), reportes (HU-38 y HU-39), total de ítems ni cantidad de sprints (el Issue no los pide), ni cambios en los repositorios, en el esquema de la base o en los paquetes `project`, `backlog`, `sprint` y `metrics`.

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `ProjectSummary(projectID int) (Summary, error)` | ID de un proyecto | Resumen del proyecto |

Está en el `Manager` de `internal/dashboard`, que se crea con `NewManager(projects *project.Manager, items *backlog.Manager, sprints *sprint.Manager)`.

### Salidas esperadas

- **`Summary`** (por valor, sin slices, mapas ni punteros): `ProjectID`, `ProjectName`, `ProjectStatus` (`project.Status`), `Pending`, `InProgress` y `Finished` (cantidades de ítems), `CurrentSprint` (`sprint.Sprint`, una copia) y `HasCurrentSprint` (`false` si no hay sprint actual).
- **Error:** el que indica el motivo (ver "Condiciones de error").

### HTTP (TEC-04)

`GET /api/proyectos/{id}/resumen` responde 200:

```json
{
  "proyecto_id": 1,
  "nombre": "Software Metrics",
  "estado": "EN_CURSO",
  "items_pendientes": 2,
  "items_en_curso": 1,
  "items_terminados": 3,
  "sprint_actual": {
    "id": 2,
    "proyecto_id": 1,
    "nombre": "Sprint 2",
    "fecha_inicio": "2026-10-27",
    "fecha_fin": "2026-11-02",
    "objetivo": "Ver las métricas en el dashboard",
    "estado": "ABIERTO"
  }
}
```

`sprint_actual` reutiliza `sprintResponse` y `toSprintResponse` de `internal/httpapi/sprints.go`, con los mismos campos (incluido `proyecto_id`) y el mismo formato de fechas. En la respuesta es un puntero (`*sprintResponse`), que queda en `null` si no hay sprint abierto.

| Caso | Código | Mensaje |
|---|---|---|
| `{id}` no numérico | 400 | El de `errInvalidID` |
| Proyecto inexistente | 404 | El de `project.ErrProjectNotFound` |
| Falla de un repositorio | 500 | El de `errInternal`, sin detalle |

### Interfaz web (TEC-04)

Sección nueva "Dashboard" con el botón "Ver resumen", que usa el proyecto actual. Muestra, por ejemplo:

- "Proyecto «Software Metrics» · Estado: EN_CURSO"
- "Ítems: 2 pendientes, 1 en curso, 3 terminados"
- "Sprint actual: Sprint 2 (2026-10-27 al 2026-11-02) · Objetivo: Ver las métricas en el dashboard", o "Sprint actual: ninguno abierto"

Sin proyecto actual, el mismo mensaje de error que las demás secciones.

## Reglas de negocio

1. **RN-36.1 Estado del proyecto.** Se muestran el nombre y el estado guardados del proyecto (`PLANIFICADO`, `EN_CURSO` o `FINALIZADO`, RN-01.6), leídos con `project.Manager.Get`. No se calculan a partir de los ítems.
2. **RN-36.2 Ítems por estado.** Se cuentan todos los ítems del backlog del proyecto, leídos con `backlog.Manager.ListItems`, asignados o no a un sprint. Siempre se informan los tres estados, incluso con 0.
3. **RN-36.3 Sprint actual.** Los sprints se leen con `sprint.Manager.ListByProject`, y el sprint actual es el `ABIERTO` de esa lista. Por RN-10.5 hay como máximo uno. Los sprints `CERRADO` no son actuales. Si no hay ninguno abierto, `HasCurrentSprint` es `false`, la API responde `null` y la interfaz muestra "ninguno abierto".
4. **RN-36.4 Proyecto sin datos.** Un proyecto sin ítems ni sprints no es un error: las tres cantidades valen 0 y no hay sprint actual.
5. **RN-36.5 Datos al momento de la consulta.** El resumen se arma con el estado guardado en ese momento, sin caché. Si un ítem cambia de estado o se cierra el sprint, la consulta siguiente lo refleja.
6. **RN-36.6 Solo lectura.** Consultar no modifica proyectos, ítems ni sprints.
7. **RN-36.7 Copias.** `Summary` y `CurrentSprint` se devuelven por valor; ningún llamador recibe referencias al estado interno de los managers.

## Decisión sobre entidades referenciadas

HU-36 referencia un proyecto por ID y **sí valida su existencia** con `project.Manager.Get`, que devuelve `project.ErrProjectNotFound`, antes de leer los ítems y los sprints. Igual que en HU-32 y HU-34, sin la validación un ID mal escrito mostraría un "resumen vacío" y ocultaría el error. No se validan ítems ni sprints uno por uno: se leen a partir del proyecto.

## Restricciones

- **Organización del código.** `internal/dashboard/doc.go`, `manager.go` y `summary.go` (nuevos). No se tocan `project`, `backlog`, `sprint` ni `metrics`.
- **Lecturas.** `ProjectSummary` lee, en este orden, `project.Manager.Get` (proyecto y validación), `backlog.Manager.ListItems` (ítems) y `sprint.Manager.ListByProject` (sprints; el actual es el `ABIERTO` de la lista).
- **HTTP.** `internal/httpapi/dashboard.go` (nuevo): el handler y la respuesta, que reutiliza `sprintResponse` y `toSprintResponse` de `sprints.go`. En `server.go`: la ruta y la creación de `dashboard.Manager`. Se agrega una fila a `TestRutasConID`.
- **Interfaz.** `internal/web/index.html` (sección nueva) y `app.js` (botón).
- **Varias lecturas por pedido (límite conocido).** Son tres lecturas separadas y, además, `ListItems` vuelve a consultar el proyecto para validarlo (RN-06.4). Si alguien modifica algo entre dos lecturas, el resumen puede mezclar momentos distintos. Es el mismo límite conocido que HU-34 y es aceptable para una consulta de solo lectura.
- **Concurrencia.** `dashboard` no tiene estado propio: cada lectura queda protegida por el mutex del manager que la hace, como en `metrics`. No se agrega un test de concurrencia.
- **RED.** Firmas vacías de `Summary`, `Manager`, `NewManager` y `ProjectSummary` para que el RED compile y falle por aserciones; `go vet ./...` antes del commit RED. La ruta, el handler y la interfaz van en el GREEN.
- **Tests.** Unitarios en `internal/dashboard/summary_test.go`; godog en `internal/dashboard/summary_features_test.go`, con el runner `TestFeatures_HU36ResumenProyecto` sobre `features/HU-36-resumen-proyecto.feature`; API en `internal/httpapi/dashboard_test.go`; interfaz en `internal/httpapi/web_test.go`.
- **Librerías.** Solo la librería estándar; godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| 2 pendientes, 1 en curso, 3 terminados y un sprint abierto | Las cantidades y el sprint abierto |
| Proyecto sin ítems ni sprints | 0, 0, 0 y sin sprint actual, sin error |
| Ítems en un solo estado | Los otros dos estados en 0 |
| Ítems sin sprint asignado | Se cuentan igual |
| Solo sprints cerrados | Sin sprint actual (`null`) |
| Un sprint cerrado y otro abierto | El abierto es el actual |
| Ítem que pasa a `TERMINADO` y se vuelve a consultar | Las cantidades cambian |
| Sprint que se cierra y se vuelve a consultar | Sin sprint actual |
| Ítems y sprints de otro proyecto | No se cuentan |
| Proyecto inexistente, 0 o negativo | `project.ErrProjectNotFound` (404) |
| Falla de un repositorio | Se propaga el error (500 sin detalle) |
| Consultar dos veces sin cambios | Mismo resultado; no se modifica nada |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Proyecto inexistente | `project.ErrProjectNotFound` | No se lee nada más |
| Falla al leer el proyecto, los ítems o los sprints | El error del repositorio | No se devuelve un resumen parcial |

## Criterios de aceptación

- **CA-36.1** El resumen de un proyecto muestra su nombre y estado, la cantidad de ítems del backlog en `PENDIENTE`, `EN_CURSO` y `TERMINADO`, y el sprint actual (el sprint `ABIERTO` del proyecto).
- **CA-36.2** Un proyecto sin ítems ni sprints muestra el resumen vacío (cantidades en 0 y sin sprint actual), no un error.
- **CA-36.3** Un proyecto sin sprint abierto muestra que no hay sprint actual.
- **CA-36.4** Los datos coinciden con los guardados: después de cambiar el estado de un ítem o de cerrar el sprint, el resumen lo refleja.
- **CA-36.5** Consultar un proyecto inexistente devuelve error, y consultar no modifica nada.
- **CA-36.6** El resumen se ve desde la API y desde la sección "Dashboard" de la interfaz web.

## Comparación con el acta de Planning del Sprint 4

Sin diferencias de alcance. El acta describe "Resumen del estado del proyecto a partir de los ítems y sprints existentes, con su sección en la interfaz", 3 SP y sin dependencias pendientes (E1 a E3 en `main`). La SDD no agrega métricas ni gráficos, que quedan en HU-35 y HU-37.

## Comparación con los criterios del Issue #54 (borrador)

| CA en el Issue #54 | En esta SDD | Ajuste |
|---|---|---|
| CA-36.1 Estado del proyecto, ítems por estado y sprint actual | CA-36.1 | Se precisa qué es el "sprint actual" |
| CA-36.2 Proyecto sin datos muestra el resumen vacío | CA-36.2 y CA-36.3 | Se separa el caso "sin sprint abierto" |
| CA-36.3 Los datos coinciden con los guardados | CA-36.4 | Se precisa cómo se verifica |
| — | CA-36.5 y CA-36.6 (nuevos) | Errores, API e interfaz |

Texto propuesto para el Issue #54 (se aplica con autorización):

```markdown
## Criterios de aceptación

- [ ] CA-36.1 El resumen de un proyecto muestra su nombre y estado, la cantidad de ítems del backlog en PENDIENTE, EN_CURSO y TERMINADO, y el sprint actual (el sprint ABIERTO del proyecto).
- [ ] CA-36.2 Un proyecto sin ítems ni sprints muestra el resumen vacío (cantidades en 0 y sin sprint actual), no un error.
- [ ] CA-36.3 Un proyecto sin sprint abierto muestra que no hay sprint actual.
- [ ] CA-36.4 Los datos coinciden con los guardados: después de cambiar el estado de un ítem o de cerrar el sprint, el resumen lo refleja.
- [ ] CA-36.5 Consultar un proyecto inexistente devuelve error, y consultar no modifica nada.
- [ ] CA-36.6 El resumen se ve desde la API y desde la sección "Dashboard" de la interfaz web.
```

## Decisiones tomadas en la redacción

1. **Paquete nuevo `internal/dashboard`** en lugar de `internal/metrics`: el resumen no es una métrica de TEC-03, y HU-37 tiene dónde sumar lo suyo.
2. **Endpoint nuevo `/resumen`** en lugar de agregar campos a `/metricas`: son datos distintos y HU-35 rehace `/metricas`.
3. **`sprint_actual` con `sprintResponse`** (ajuste de la aprobación): mismos campos que el resto de la API de sprints, y `null` si no hay sprint abierto, como la velocidad de HU-32.
4. **Sin total de ítems ni cantidad de sprints**: el Issue no los pide (confirmado en la aprobación).
5. **Sección "Dashboard"** nueva en la interfaz, donde van los gráficos de HU-37 (confirmado en la aprobación).

HU-36: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
