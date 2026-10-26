# HU-34 · Calcular el porcentaje de historias completadas

> Issue del repositorio: #52

Aprobada por: Peñalbé Hernán (Agile Enabler), 26/10/2026, con las dos diferencias con el acta (sprint y división) registradas en "Comparación con el acta de Planning".

## Historia de usuario

Como Agile Enabler, quiero calcular el porcentaje de historias completadas de un sprint y de un proyecto, para evaluar el avance.

## Objetivo

Calcular la métrica M8 (porcentaje de historias completadas) de `docs/metricas.md` (TEC-03), por sprint y por proyecto, en el paquete `internal/metrics` de HU-32, y agregarla a los endpoints `GET /api/sprints/{id}/metricas` y `GET /api/proyectos/{id}/metricas` y a la sección "Métricas" de la interfaz web, como prevé la SDD de HU-32 ("HU-33 y HU-34 agregan sus métricas a los mismos endpoints").

Esta historia es la primera parte de la división de HU-34 que propone el acta de Planning del Sprint 3: **HU-34** (porcentaje de historias completadas) y **HU-47** (defectos detectados y resueltos, M9 y M10). HU-47 (Issue #128) queda para después de HU-30 (cambiar el estado de un defecto), que es la que permite que un defecto llegue a `RESUELTO`.

## Alcance

- **Incluye:** el cálculo de M8 de un sprint y de un proyecto, con la cantidad de ítems y de ítems terminados; el caso "no disponible" cuando no hay ítems; los campos nuevos en las dos respuestas de métricas de HU-32 y el texto de la sección "Métricas" de la interfaz.
- **No incluye:** defectos detectados y resueltos (M9 y M10, HU-47), horas y desviación (HU-33), la consulta única de M1 a M10 (HU-35), gráficos (HU-36 y HU-37), reportes (HU-38 y HU-39), congelar los valores al cerrar un sprint (mismo límite conocido que HU-32) ni cambios en los repositorios o en el esquema de la base.

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `SprintCompletion(sprintID int) (Completion, error)` | ID de un sprint | M8 del sprint |
| `ProjectCompletion(projectID int) (Completion, error)` | ID de un proyecto | M8 del proyecto |

Las dos están en el `Manager` de `internal/metrics`, en un archivo nuevo `completion.go`.

### Salidas esperadas

- **`Completion`** (por valor, sin slices ni mapas): `Items` (cantidad de ítems del sprint o del proyecto), `Finished` (cantidad de esos ítems en `TERMINADO`), `Percent` (M8, `float64` sin redondear) y `PercentAvailable` (`false` si `Items` es 0).
- **Error:** el que indica el motivo (ver "Condiciones de error").

### HTTP (TEC-04)

Se agregan tres campos a las dos respuestas existentes de HU-32, sin cambiar los que ya tienen:

| Campo | Tipo | Descripción |
|---|---|---|
| `items_totales` | entero | Ítems del sprint o del proyecto |
| `items_terminados` | entero | Ítems en `TERMINADO` |
| `porcentaje_completadas` | número o `null` | M8 redondeado a un decimal; `null` si no hay ítems |

Ejemplo de `GET /api/sprints/{id}/metricas` con el sprint del ejemplo de TEC-03 (2 de 4 ítems terminados):

```json
{
  "sprint_id": 1,
  "sp_planificados": 10,
  "sp_completados": 8,
  "items_sin_estimar": 1,
  "items_totales": 4,
  "items_terminados": 2,
  "porcentaje_completadas": 50
}
```

Los errores no cambian respecto de HU-32:

| Caso | Código | Mensaje |
|---|---|---|
| `{id}` no numérico | 400 | El de `errInvalidID` |
| Sprint inexistente | 404 | El de `sprint.ErrSprintNotFound` |
| Proyecto inexistente | 404 | El de `project.ErrProjectNotFound` |
| Falla de un repositorio | 500 | El de `errInternal`, sin detalle |

### Interfaz web (TEC-04)

En la sección "Métricas", el texto que arma "Ver métricas" suma, para el proyecto y para el sprint actual, el porcentaje de historias completadas con un decimal y las cantidades: por ejemplo, "Historias completadas: 50.0 % (2 de 4)". Si no hay ítems, muestra "Historias completadas: sin datos". No hay formularios ni rutas nuevas.

## Reglas de negocio

1. **RN-34.1 Ítems del sprint.** Los ítems de un sprint son los ítems del backlog con `SprintID` igual al sprint, como en HU-32 (RN-32.1).
2. **RN-34.2 Ítems del proyecto.** Los ítems de un proyecto son todos los ítems de su backlog, asignados o no a un sprint (decisión de TEC-03).
3. **RN-34.3 Completado.** Un ítem cuenta como completado si está en `TERMINADO`. `PENDIENTE` y `EN_CURSO` no cuentan.
4. **RN-34.4 Cuentan todos los ítems.** M8 cuenta ítems, no Story Points: un ítem sin Story Points ni horas estimadas cuenta igual en `Items` y, si está terminado, en `Finished`.
5. **RN-34.5 Fórmula.** `Percent = Finished / Items × 100`, sin redondear en el dominio.
6. **RN-34.6 No disponible.** Un sprint o proyecto sin ítems no es un error: `Items` y `Finished` valen 0, `Percent` vale 0 y `PercentAvailable` es `false`. La API responde `null` y la interfaz muestra "sin datos" (tabla "Datos faltantes" de TEC-03).
7. **RN-34.7 Redondeo de presentación.** La API y la interfaz muestran M8 con un decimal (`math.Round(p*10)/10` en el handler, como la velocidad de HU-32); el dominio no redondea.
8. **RN-34.8 Sprints abiertos y cerrados.** Se calcula igual para un sprint `ABIERTO` o `CERRADO`, con el estado actual de los ítems (mismo límite conocido que HU-32: los valores no se congelan al cerrar el sprint).
9. **RN-34.9 Solo lectura.** Calcular no modifica ítems, sprints ni proyectos.

## Decisión sobre entidades referenciadas

HU-34 referencia por ID un sprint o un proyecto, y **sí valida su existencia**, igual que HU-32:

- **Sprint:** con `sprint.Manager.Get`, que devuelve `sprint.ErrSprintNotFound`. Sin la validación, un ID mal escrito daría "no disponible" y ocultaría el error.
- **Proyecto:** con `backlog.Manager.ListItems`, que ya valida el proyecto y devuelve `project.ErrProjectNotFound`.

No se validan los ítems uno por uno: se leen del backlog del proyecto.

## Restricciones

- **Organización del código.** `internal/metrics/completion.go` (nuevo): el tipo `Completion`, la función interna que cuenta los ítems y los métodos `SprintCompletion` y `ProjectCompletion`. Reutiliza `itemsOfSprint` de HU-32. No se tocan `points.go`, `story_points.go`, `manager.go`, los repositorios ni el esquema.
- **HTTP.** `internal/httpapi/metrics.go`: los dos handlers de HU-32 llaman también a `SprintCompletion` o `ProjectCompletion` y agregan los tres campos a sus respuestas. Las rutas no cambian, así que no se agregan filas a `TestRutasConID` (ya están las de HU-32).
- **Interfaz.** `internal/web/app.js`: el texto de "Ver métricas". `index.html` no cambia.
- **Dos lecturas por pedido.** Cada endpoint calcula los Story Points (HU-32) y M8 con dos lecturas del backlog. Si alguien modifica un ítem justo entre las dos, los valores de una misma respuesta pueden corresponder a momentos distintos. Es un límite conocido aceptable para una consulta de solo lectura; la consulta única de HU-35 lo resuelve.
- **Concurrencia.** `metrics` no tiene estado propio: cada lectura queda protegida por el mutex del `Manager` que la hace (backlog y sprint), como en HU-32. No se agrega un test de concurrencia.
- **RED.** Firmas vacías de `Completion`, `SprintCompletion` y `ProjectCompletion` para que el RED compile y falle por aserciones; `go vet ./...` antes del commit RED. Los campos nuevos de la API y el texto de la interfaz van en el GREEN.
- **Tests.** Unitarios en `internal/metrics/completion_test.go`; godog en `internal/metrics/completion_features_test.go`, con el runner `TestFeatures_HU34PorcentajeCompletadas` sobre `features/HU-34-porcentaje-completadas.feature`, cada escenario con su CA en el nombre; API en `internal/httpapi/metrics_test.go`; interfaz en `internal/httpapi/web_test.go`.
- **Librerías.** Solo la librería estándar; godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Ejemplo de TEC-03: 4 ítems, 2 terminados | 50 % (API `50`, interfaz "50.0 %") |
| 1 de 3 terminados | Dominio 33,333…; API `33.3` |
| Todos terminados | 100 % |
| Ninguno terminado | 0 % (disponible) |
| Sprint o proyecto sin ítems | No disponible; API `null`; interfaz "sin datos" |
| Ítem terminado sin Story Points | Cuenta en `Items` y en `Finished` |
| Ítem de otro sprint | No cuenta en el sprint consultado |
| Ítem sin sprint | Cuenta en el proyecto y en ningún sprint |
| Sprint `CERRADO` | Se calcula igual, con el estado actual |
| Sprint inexistente, 0 o negativo | `sprint.ErrSprintNotFound` (404) |
| Proyecto inexistente, 0 o negativo | `project.ErrProjectNotFound` (404) |
| Falla de un repositorio | Se propaga el error (500 sin detalle) |
| Consultar dos veces | Mismo resultado; no se modifica nada |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Sprint inexistente | `sprint.ErrSprintNotFound` | No se calcula nada |
| Proyecto inexistente | `project.ErrProjectNotFound` | No se calcula nada |
| Falla al leer ítems o sprints | El error del repositorio | No se calcula nada |

## Criterios de aceptación

- **CA-34.1** El porcentaje de historias completadas de un sprint es la cantidad de sus ítems en `TERMINADO` sobre la cantidad de sus ítems, por 100.
- **CA-34.2** El porcentaje de un proyecto se calcula sobre todo su backlog, asignado o no a un sprint.
- **CA-34.3** Un sprint o un proyecto sin ítems no produce error: el porcentaje queda como no disponible.
- **CA-34.4** Cuentan todos los ítems, tengan o no Story Points.
- **CA-34.5** Consultar un sprint o un proyecto inexistente devuelve error y consultar no modifica nada.
- **CA-34.6** El porcentaje se ve desde la API y desde la interfaz web, con un decimal, junto con las cantidades de ítems y de ítems terminados.

## Comparación con el acta de Planning

El acta del Sprint 3 deja a HU-34 en el Sprint 4 y propone dividirla en **HU-34** (porcentaje de historias completadas) y **HU-47** (defectos detectados y resueltos). Diferencias con esta SDD:

1. **Sprint.** Se adelanta al final del Sprint 3, como HU-32, por la replanificación del 24/10 (entrega del 02/11). HU-34 no tiene Story Points estimados en un acta: la Definition of Ready los pide, así que hay que estimarla (o registrar que se estima en el Planning del Sprint 4).
2. **División.** Se aplica la división del acta: esta SDD es solo M8. M9 y M10 pasan a HU-47, que necesita HU-30 (estado resuelto) y una consulta de defectos que se coordina con HU-31 (Sosa).

## Comparación con los criterios del Issue #52 (borrador)

| CA en el Issue #52 | En esta SDD | Ajuste |
|---|---|---|
| CA-34.1 El porcentaje de historias completadas se calcula sobre las historias del sprint o proyecto, y un total cero no produce error. | CA-34.1, CA-34.2 y CA-34.3 | Se separa por nivel y se precisa el "no disponible" |
| CA-34.2 Los defectos detectados se cuentan por sprint de detección y los resueltos por sprint de resolución. | — | Pasa a HU-47 |
| CA-34.3 Sin defectos el resultado es cero. | — | Pasa a HU-47 |
| — | CA-34.4 a CA-34.6 (nuevos) | Ítems sin Story Points, errores y API/interfaz |

Texto propuesto para el Issue #52 (se aplica con autorización, después de aprobar la SDD): título "HU-34 · Calcular el porcentaje de historias completadas" y estos criterios:

```markdown
## Criterios de aceptación

- [ ] CA-34.1 El porcentaje de historias completadas de un sprint es la cantidad de sus ítems terminados sobre la cantidad de sus ítems, por 100.
- [ ] CA-34.2 El porcentaje de un proyecto se calcula sobre todo su backlog, asignado o no a un sprint.
- [ ] CA-34.3 Un sprint o un proyecto sin ítems no produce error: el porcentaje queda como no disponible.
- [ ] CA-34.4 Cuentan todos los ítems, tengan o no Story Points.
- [ ] CA-34.5 Consultar un sprint o un proyecto inexistente devuelve error y consultar no modifica nada.
- [ ] CA-34.6 El porcentaje se ve desde la API y desde la interfaz web, con un decimal, junto con las cantidades de ítems y de ítems terminados.
```

Los criterios CA-34.2 y CA-34.3 del borrador pasan a HU-47, Issue #128 (épica E9), como CA-47.1 y CA-47.2.

## Decisiones tomadas en la redacción

1. **Métodos separados (`SprintCompletion` y `ProjectCompletion`)** en lugar de agregar campos a `SprintPoints` y `ProjectPoints` de HU-32. *Motivo:* M8 cuenta ítems, no Story Points, y así no se toca el código de HU-32. *Costo:* dos lecturas por pedido (ver "Restricciones").
2. **Mismos endpoints de HU-32** con tres campos nuevos, como prevé su SDD. *Alternativa descartada:* rutas nuevas, que HU-35 tendría que volver a juntar.
3. **"No disponible" como `null`** en la API, igual que la velocidad de HU-32.
4. **División del acta aplicada:** solo M8; M9 y M10 en HU-47.

HU-34: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
