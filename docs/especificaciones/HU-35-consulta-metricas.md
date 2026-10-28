# HU-35 · Consultar las métricas del proyecto

> Issue del repositorio: #53

Aprobada por: Peñalbé Hernán (Agile Enabler), 28/10/2026, con dos ajustes (comparación del JSON crudo para el orden de los campos y CA-35.4 y CA-35.6 en lenguaje de negocio en el Issue), que se aplicaron.

## Historia de usuario

Como Agile Enabler, quiero consultar todas las métricas de un proyecto juntas, para evaluar su estado general.

## Objetivo

Devolver en una sola consulta todas las métricas M1 a M10 de `docs/metricas.md` (TEC-03), por sprint y por proyecto, con un único criterio de "no disponible", reutilizando los cálculos de HU-32, HU-33, HU-34 y HU-47 sin agregar cálculos nuevos. La consulta lee los ítems del backlog una sola vez y el sprint (o la lista de sprints del proyecto) una sola vez, y elimina así la doble lectura del backlog de HU-34 y las lecturas repetidas de HU-32 y HU-33 dentro de `internal/metrics`. Se exponen en los dos endpoints de métricas existentes, con el contrato documentado para HU-37, HU-38 y HU-39.

## Alcance

- **Incluye:**
  - dos operaciones nuevas en `internal/metrics`, `Sprint(sprintID)` y `Project(projectID)`, que calculan M1 a M10 con una sola lectura de los ítems y una sola del sprint o de la lista de sprints;
  - los dos handlers de `internal/httpapi/metrics.go`, que pasan a hacer una sola llamada a `metrics`;
  - una sola función de la API para los cuatro valores que pueden ser `null` (velocidad, porcentaje de completadas y las dos desviaciones), con la corrección del cero negativo;
  - el contrato final de los dos endpoints, con ejemplos, para HU-37, HU-38 y HU-39.
- **No incluye:**
  - cambios en `internal/effort` ni en `internal/defect`: ni en sus repositorios, ni en sus `Manager`, ni en sus tests. Los esfuerzos y los defectos se siguen leyendo con `effort.Manager.ListBySprint` y `defect.Manager.ListByProject`;
  - cálculos o métricas nuevas, ni cambios en las fórmulas de TEC-03;
  - rutas nuevas (el resumen de HU-36 sigue en `/resumen`);
  - cambios en los campos, los tipos o los códigos HTTP existentes;
  - cambios en la interfaz web: la sección "Métricas" ya muestra M1 a M10 y "sin datos" para las cuatro métricas que pueden no estar disponibles;
  - el dashboard con gráficos (HU-37) y los reportes (HU-38, HU-39 y HU-40);
  - cambios de esquema, congelar los valores al cerrar un sprint (límite de HU-32) ni transacciones entre fuentes.

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `Sprint(sprintID int) (SprintMetrics, error)` | ID de un sprint | M1, M2 y M4 a M10 del sprint |
| `Project(projectID int) (ProjectMetrics, error)` | ID de un proyecto | M1 a M10 del proyecto |

Las dos operaciones están en el `Manager` de `internal/metrics`. M3 (velocidad) es solo del proyecto (TEC-03).

### Salidas esperadas

`SprintMetrics` y `ProjectMetrics` (por valor), formadas por los tipos que ya existen:

| Campo | Tipo | Origen |
|---|---|---|
| `SprintID` / `ProjectID` | `int` | ID consultado |
| `Points` | `SprintPoints` / `ProjectPoints` | M1, M2, ítems sin estimar y, en el proyecto, M3 (HU-32) |
| `Completion` | `Completion` | M8 y sus cantidades (HU-34) |
| `Hours` | `Hours` | M4 a M7 e ítems sin horas estimadas (HU-33) |
| `Defects` | `Defects` | M9 y M10 (HU-47) |

- Ningún tipo tiene punteros, listas ni mapas: el resultado es una copia completa.
- Los valores no se redondean en el dominio (TEC-03); el redondeo es de la API.
- Los indicadores de disponibilidad ya existentes se conservan con su nombre: `ProjectPoints.VelocityAvailable`, `Completion.PercentAvailable` y `Hours.DeviationAvailable`.

### HTTP

Se mantienen las dos rutas de HU-32, ampliadas por HU-33, HU-34 y HU-47. Cada handler hace una sola llamada a `metrics` y arma la misma respuesta que hoy: mismos campos, mismo orden, mismos tipos y mismos códigos.

**Contrato final para HU-37, HU-38 y HU-39.** Los ejemplos muestran la salida real de `encoding/json`: un número sin parte decimal se emite sin `.0` (por ejemplo `50` en lugar de `50.0`). Quien muestre los valores tiene que formatearlos con un decimal (TEC-03).

`GET /api/sprints/{id}/metricas`, con el sprint del ejemplo de TEC-03 (un defecto detectado y resuelto en el mismo sprint):

```json
{
  "sprint_id": 1,
  "sp_planificados": 10,
  "sp_completados": 8,
  "items_sin_estimar": 1,
  "items_totales": 4,
  "items_terminados": 2,
  "porcentaje_completadas": 50,
  "horas_estimadas": 20,
  "horas_reales": 20,
  "desviacion_horas": 0,
  "desviacion_porcentual": 0,
  "items_sin_horas_estimadas": 1,
  "defectos_detectados": 1,
  "defectos_resueltos": 1
}
```

`GET /api/proyectos/{id}/metricas`, con el mismo proyecto (ese sprint cerrado y ningún otro):

```json
{
  "proyecto_id": 1,
  "sp_planificados": 10,
  "sp_completados": 8,
  "items_sin_estimar": 1,
  "velocidad": 8,
  "sprints_cerrados": 1,
  "items_totales": 4,
  "items_terminados": 2,
  "porcentaje_completadas": 50,
  "horas_estimadas": 20,
  "horas_reales": 20,
  "desviacion_horas": 0,
  "desviacion_porcentual": 0,
  "items_sin_horas_estimadas": 1,
  "defectos_detectados": 1,
  "defectos_resueltos": 1
}
```

`GET /api/proyectos/{id}/metricas` de un proyecto recién creado (sin ítems, sprints, esfuerzos ni defectos), con todas las métricas no disponibles:

```json
{
  "proyecto_id": 2,
  "sp_planificados": 0,
  "sp_completados": 0,
  "items_sin_estimar": 0,
  "velocidad": null,
  "sprints_cerrados": 0,
  "items_totales": 0,
  "items_terminados": 0,
  "porcentaje_completadas": null,
  "horas_estimadas": 0,
  "horas_reales": 0,
  "desviacion_horas": null,
  "desviacion_porcentual": null,
  "items_sin_horas_estimadas": 0,
  "defectos_detectados": 0,
  "defectos_resueltos": 0
}
```

Campos de la respuesta:

| Campo | Métrica | Tipo | Nivel | Puede ser `null` |
|---|---|---|---|---|
| `sprint_id` / `proyecto_id` | — | entero | Sprint / proyecto | No |
| `sp_planificados` | M1 | entero | Ambos | No |
| `sp_completados` | M2 | entero | Ambos | No |
| `items_sin_estimar` | Ítems con `StoryPoints = 0` | entero | Ambos | No |
| `velocidad` | M3 | número, un decimal | Proyecto | Sí: sin sprints cerrados |
| `sprints_cerrados` | Sprints usados para M3 (0 a 3) | entero | Proyecto | No |
| `items_totales` | Ítems del conjunto | entero | Ambos | No |
| `items_terminados` | Ítems `TERMINADO` | entero | Ambos | No |
| `porcentaje_completadas` | M8 | número, un decimal | Ambos | Sí: sin ítems |
| `horas_estimadas` | M4 | número, un decimal | Ambos | No |
| `horas_reales` | M5 | número, un decimal | Ambos | No |
| `desviacion_horas` | M6 | número, un decimal | Ambos | Sí: M4 = 0 o M5 = 0 |
| `desviacion_porcentual` | M7 | número, un decimal | Ambos | Sí: M4 = 0 o M5 = 0 |
| `items_sin_horas_estimadas` | Ítems con `EstimatedHours = 0` | entero | Ambos | No |
| `defectos_detectados` | M9 | entero | Ambos | No |
| `defectos_resueltos` | M10 | entero | Ambos | No |

| Caso | Código | Cuerpo |
|---|---|---|
| Consulta exitosa, aunque alguna métrica no esté disponible | 200 | El JSON de arriba |
| `{id}` que no es un número entero | 400 | `{"error": "el id debe ser un número entero"}` (`errInvalidID`) |
| Sprint inexistente (incluidos 0 y negativos) | 404 | `{"error": "no existe un sprint con ese id"}` (`sprint.ErrSprintNotFound`) |
| Proyecto inexistente (incluidos 0 y negativos) | 404 | `{"error": "no existe un proyecto con ese id"}` (`project.ErrProjectNotFound`) |
| Falla de cualquier repositorio | 500 | `{"error": "error interno del servidor"}` (`errInternal`), sin detalle |

### Interfaz web

No cambia. La sección "Métricas" ya consume estos dos endpoints y muestra "sin datos" para `velocidad`, `porcentaje_completadas` y la desviación (HU-32, HU-34 y HU-33). HU-37 construye el dashboard sobre el mismo contrato.

## Reglas de negocio

1. **RN-35.1 Consulta única.** `Sprint` y `Project` devuelven todas las métricas de su nivel en una sola llamada: el sprint, M1, M2 y M4 a M10; el proyecto, M1 a M10. Los handlers no llaman a ninguna otra operación de `metrics`.
2. **RN-35.2 Sin cálculos nuevos.** Cada métrica se calcula con la función pura que ya existe: `storyPoints` y `velocity` (HU-32), `completion` (HU-34), `hoursOf` (HU-33) y el conteo de defectos de HU-47, que se extrae a una función pura sin cambiar su regla. Los resultados son los mismos que devuelven hoy las operaciones separadas con los mismos datos.
3. **RN-35.3 Lecturas de `metrics`.** En cada consulta, `metrics` lee una sola vez los ítems del backlog (`backlog.Manager.ListItems`) y una sola vez el sprint (`sprint.Manager.Get`) o la lista de sprints del proyecto (`sprint.Manager.ListByProject`), y usa esa misma lectura para todas las métricas. Los esfuerzos y los defectos se leen con las operaciones actuales, sin cambios:

   | Lectura que hace `metrics` | Sprint (hoy → HU-35) | Proyecto (hoy → HU-35) |
   |---|---|---|
   | `backlog.Manager.ListItems` | 3 → 1 | 3 → 1 |
   | `sprint.Manager.Get` | 4 → 1 | — |
   | `sprint.Manager.ListByProject` | — | 2 → 1 |
   | `effort.Manager.ListBySprint` | 1 → 1 | N → N (uno por sprint) |
   | `defect.Manager.ListByProject` | 1 → 1 | 1 → 1 |

   N es la cantidad de sprints del proyecto. Las lecturas internas de `effort` y `defect` siguen igual (ver "Límite conocido").
4. **RN-35.4 Orden de lectura.** Sprint: sprint (`Get`), ítems del proyecto del sprint (`ListItems`), esfuerzos del sprint (`ListBySprint`) y defectos del proyecto (`defect.Manager.ListByProject`). Proyecto: ítems (`ListItems`, que valida el proyecto), sprints del proyecto (`sprint.Manager.ListByProject`), esfuerzos de cada sprint en orden de ID (`ListBySprint`) y defectos del proyecto. Si una lectura falla, no se hacen las siguientes y se devuelve ese error. Así un sprint o un proyecto inexistente devuelve el mismo error que hoy.
5. **RN-35.5 Esfuerzos.** M5 se obtiene como hoy: en el sprint, `Summary.TotalHours` de `ListBySprint` (RN-33.3); en el proyecto, la suma de `Summary.TotalHours` de cada sprint, en orden de ID de sprint (RN-33.4). El resultado es exactamente el mismo `float64` que el de HU-33.
6. **RN-35.6 "No disponible" unificado.** Una métrica no disponible no es un error: la consulta responde 200 con las demás métricas, y la API la devuelve `null`, nunca 0 ni omitiendo el campo. Solo cuatro campos pueden no estar disponibles, por las reglas de TEC-03 que ya implementaron las historias anteriores:

   | Campo | No disponible si | Regla de origen |
   |---|---|---|
   | `velocidad` | No hay sprints cerrados | RN-32.6 |
   | `porcentaje_completadas` | No hay ítems | RN-34.6 |
   | `desviacion_horas` y `desviacion_porcentual` | M4 = 0 o M5 = 0 | RN-33.6 |

   Las demás métricas son sumas o cantidades y valen 0 cuando no hay datos. Las cuatro se convierten con una sola función de la API que redondea a un decimal, normaliza el cero negativo (RN-33.8) y devuelve `null` si no están disponibles; reemplaza a las conversiones separadas de HU-32, HU-33 y HU-34 sin cambiar ningún valor de salida.
7. **RN-35.7 Redondeo.** Sin cambios: el dominio no redondea; la API redondea `velocidad`, `porcentaje_completadas` y M4 a M7 a un decimal; los Story Points y las cantidades son enteros (TEC-03).
8. **RN-35.8 Compatibilidad.** No cambian las rutas, los nombres, los tipos ni el orden de los campos, ni los códigos HTTP. Las operaciones de HU-32, HU-33, HU-34 y HU-47 (`SprintStoryPoints`, `ProjectStoryPoints`, `SprintCompletion`, `ProjectCompletion`, `SprintHours`, `ProjectHours`, `SprintDefects` y `ProjectDefects`) se conservan con su firma y su comportamiento, con sus tests y escenarios.
9. **RN-35.9 Datos actuales y solo lectura.** Las métricas se calculan en el momento de la consulta con los datos guardados. Consultar no modifica nada.
10. **RN-35.10 Copias por valor.** `SprintMetrics` y `ProjectMetrics` se devuelven por valor y no contienen punteros ni listas.
11. **RN-35.11 Sin errores nuevos.** No se agregan errores de dominio.

## Decisión sobre entidades referenciadas

- **Sprint (nivel sprint):** sí se valida su existencia en `metrics`, con `sprint.Manager.Get` (`ErrSprintNotFound`, incluidos 0 y negativos), una sola vez. Sin sprint no hay proyecto del cual leer los ítems ni forma de distinguir un sprint inexistente de uno vacío.
- **Proyecto:** sí se valida, a través de `backlog.Manager.ListItems`, que devuelve `project.ErrProjectNotFound` (HU-06), una sola vez en `metrics`. En el nivel sprint se valida igual, con el proyecto del sprint.
- **Sprints del proyecto (nivel proyecto):** no se validan uno por uno; se leen con `sprint.Manager.ListByProject`, después de validar el proyecto.
- **Ítems:** no se validan uno por uno; se leen una vez con `ListItems`.
- **Esfuerzos y defectos:** `metrics` no los valida de nuevo ni agrega validaciones. Se leen con `effort.Manager.ListBySprint` y `defect.Manager.ListByProject`, que hacen su propia validación interna del sprint y del proyecto; esas validaciones internas quedan como están (ver "Límite conocido").

## Restricciones

- **Organización del código.**

  | Archivo | Cambio |
  |---|---|
  | `internal/metrics/summary.go` | Nuevo: `SprintMetrics`, `ProjectMetrics`, `Sprint` y `Project` |
  | `internal/metrics/defects.go` | Extrae el conteo de M9 y M10 a funciones puras que usan `SprintDefects`, `ProjectDefects` y la consulta única; `defectLister` no cambia |
  | `internal/httpapi/metrics.go` | Los dos handlers usan `Sprint` y `Project`; una sola función para los cuatro campos que pueden ser `null` |

  No cambian `NewManager`, `NewHandler`, `cmd/server`, la interfaz web, `internal/effort` ni `internal/defect`.
- **Sin persistencia propia.** `metrics` no tiene repositorio ni estado. No hay tablas nuevas ni cambios de esquema.
- **Concurrencia.** `metrics` no tiene mutex. Cada lectura queda protegida por el mutex de su `Manager`, como hoy.
- **Copias por valor.** Todo se devuelve por valor.
- **Tests.**
  - Unitarios en `internal/metrics/summary_test.go`: cada métrica de la consulta única es igual a la de su operación separada con los mismos datos; los cuatro "no disponible"; los errores y su orden (RN-35.4).
  - Lecturas de ítems: un repositorio de ítems de prueba que cuenta las llamadas a `ListByProject` y verifica que `metrics` lee los ítems una sola vez por consulta, en el sprint y en el proyecto. El conteo esperado incluye la lectura interna de `defect.Manager.ListByProject`, que no se toca: 2 llamadas en total por consulta (1 de `metrics` y 1 de `defect`), contra 4 de hoy.
  - godog en `internal/metrics/summary_features_test.go`, con el runner `TestFeatures_HU35ConsultarMetricas` sobre `features/HU-35-consulta-metricas.feature`, con el CA en el nombre de cada escenario.
  - API en `internal/httpapi/metrics_test.go`: el JSON completo de los tres ejemplos del contrato, comparado campo por campo (incluido el orden de los campos), y un caso por código. El orden de los campos se verifica comparando el cuerpo crudo de la respuesta con el JSON esperado, ambos normalizados con `json.Compact`, porque al decodificar a un struct o a un mapa se pierde el orden. Los tests existentes de los dos endpoints tienen que seguir pasando sin cambios.
- **RED.** Firmas vacías de `Sprint`, `Project` y los dos tipos, para que el RED compile y falle por aserciones. `go vet ./...` antes del commit RED.
- **Profundidad proporcional (3 SP).** Un test por regla y por código; sin pruebas de mutación ni de concurrencia.
- **Librerías.** Solo la librería estándar; godog en los tests.

### Límite conocido

- **Lecturas que siguen repitiéndose.** HU-35 elimina las lecturas repetidas dentro de `metrics`, pero no toca `internal/effort` ni `internal/defect`, así que quedan estas, que dejaron así las SDD de HU-33 y HU-47:
  - **Esfuerzos sprint por sprint en el proyecto (HU-33).** El proyecto llama a `effort.Manager.ListBySprint` una vez por sprint: N lecturas de esfuerzos para N sprints, y cada una vuelve a leer su sprint con `sprint.Manager.Get` para validarlo.
  - **Validación del sprint en el nivel sprint (HU-33).** `effort.Manager.ListBySprint` vuelve a leer el sprint que `metrics` ya leyó.
  - **Backlog dentro de los defectos (HU-47 sobre HU-31).** `defect.Manager.ListByProject` lee el backlog por su cuenta con `ListItems` (que también vuelve a leer el proyecto) para saber qué defectos son del proyecto. Por eso, a nivel de repositorio, el backlog se lee dos veces por consulta: una en `metrics` y otra dentro de `defect`.
- **Lecturas de fuentes distintas no atómicas.** Una sola lectura del backlog en `metrics` hace que M1, M2, M3, M4 y M8 salgan del mismo estado de los ítems, lo que resuelve el límite de HU-34. Lo que no se resuelve es la atomicidad entre lecturas distintas: el sprint o los sprints, los ítems, los esfuerzos, los defectos y la lectura interna del backlog de `defect` son lecturas separadas, y un cambio simultáneo puede quedar reflejado en una y no en otra. Resolverlo requiere una transacción entre `Manager` distintos, fuera del alcance. Es aceptable para una consulta de solo lectura.
- **Operaciones separadas conservadas.** Las ocho operaciones de HU-32 a HU-47 quedan sin uso en los handlers; siguen teniendo sus lecturas repetidas, pero ya no se ejecutan en un pedido HTTP.
- **Valores no congelados.** Sin cambios respecto de HU-32: si se reestima o reabre un ítem de un sprint cerrado, cambian sus métricas y la velocidad.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Ejemplo de TEC-03 (sprint) con un defecto detectado y resuelto en el sprint | El JSON del sprint del contrato |
| Mismo proyecto, ese sprint cerrado y ningún otro | El JSON del proyecto del contrato (`velocidad` 8) |
| Proyecto recién creado, sin datos | 200 con M1, M2, M4, M5, M9 y M10 en 0 y `velocidad`, `porcentaje_completadas`, `desviacion_horas` y `desviacion_porcentual` en `null` |
| Sprint sin ítems, esfuerzos ni defectos | 200; `porcentaje_completadas` y desviaciones en `null`; el resto en 0 |
| Proyecto con ítems y sin sprints cerrados | Solo `velocidad` en `null` |
| Ítems estimados sin esfuerzos | Solo las desviaciones en `null` (M5 = 0) |
| Esfuerzos sin horas estimadas | Solo las desviaciones en `null` (M4 = 0) |
| Proyecto con tres sprints con esfuerzos | M5 suma los tres, igual que HU-33; `metrics` lee los ítems y la lista de sprints una sola vez |
| Esfuerzos y defectos de otro proyecto | No se incluyen en ninguno de los dos niveles |
| Defecto detectado en un sprint y resuelto en otro | M9 en el sprint de detección, M10 en el de resolución; ambos en el proyecto (HU-47) |
| Ítems estimados en 0.1 y 0.2 h y un esfuerzo de 0.3 h | `desviacion_horas` 0 y `desviacion_porcentual` 0, nunca `-0` (RN-33.8) |
| Mismos datos, consulta única y operaciones separadas | Valores idénticos en las cuatro partes |
| Sprint o proyecto inexistente, 0 o negativo | `ErrSprintNotFound` o `ErrProjectNotFound`; 404 |
| Falla la lectura del sprint o de los ítems | Ese error, sin leer esfuerzos ni defectos; 500 si es del repositorio |
| Falla la lectura de esfuerzos | Ese error, sin leer defectos; 500 |
| Falla la lectura de defectos | Ese error; 500 |
| HTTP: `{id}` no numérico | 400 |

## Condiciones de error

| Condición | Error | HTTP |
|---|---|---|
| `{id}` que no es un número entero | `errInvalidID` | 400 |
| No existe el sprint (incluidos 0 y negativos) | `sprint.ErrSprintNotFound` | 404 |
| No existe el proyecto (incluidos 0 y negativos) | `project.ErrProjectNotFound` | 404 |
| Falla el repositorio de proyectos, ítems, sprints, esfuerzos o defectos | El error del repositorio, en el orden de RN-35.4 | 500, sin detalle |
| Una métrica no está disponible | No es un error: `null` en la API (RN-35.6) | 200 |

## Criterios de aceptación

- **CA-35.1** Una sola consulta devuelve todas las métricas mínimas: Story Points planificados y completados, velocidad (solo del proyecto), horas estimadas y reales, desviación en horas y en porcentaje, porcentaje de historias completadas y defectos detectados y resueltos, del sprint (`GET /api/sprints/{id}/metricas`) o del proyecto (`GET /api/proyectos/{id}/metricas`).
- **CA-35.2** Una métrica que no se puede calcular se devuelve como no disponible (`null`) y la consulta responde con las demás; solo `velocidad`, `porcentaje_completadas`, `desviacion_horas` y `desviacion_porcentual` pueden no estar disponibles.
- **CA-35.3** Un proyecto o un sprint inexistente devuelve su error (404); un ID no numérico, 400; una falla de repositorio se propaga (500, sin detalle).
- **CA-35.4** Los ítems del backlog se leen una sola vez por consulta (en `internal/metrics`; la lectura interna de `defect.Manager.ListByProject` es un límite conocido).
- **CA-35.5** Los valores son los mismos que calculan HU-32, HU-33, HU-34 y HU-47 con los mismos datos; no hay cálculos nuevos.
- **CA-35.6** No cambian las rutas, los campos, sus tipos y su orden, ni los códigos HTTP; el contrato queda documentado con ejemplos para HU-37, HU-38 y HU-39.
- **CA-35.7** Lo devuelto es una copia por valor; consultar no modifica nada.

## Comparación con el acta de Planning

**No hay diferencias de alcance con el acta del Sprint 4.** El acta asigna HU-35 a Cardozo Leandro con **3 SP**, dependencias HU-32 y HU-34 (en `main`) y HU-33 y HU-47, y la justifica así: "Reúne en una sola consulta lo que calculan HU-32, HU-33, HU-34 y HU-47 y elimina la segunda lectura del backlog de HU-34; no agrega cálculos nuevos". La SDD hace exactamente eso: una consulta única en los endpoints existentes, una sola lectura del backlog en `metrics` y ningún cálculo nuevo. Las cuatro dependencias están en `main` (PR #123, #129, #142 y #144).

Precisiones registradas:

1. **Lecturas del sprint.** Además de la del backlog, la consulta única lee el sprint o la lista de sprints una sola vez. No amplía el alcance: es consecuencia de reunir las cuatro operaciones en una.
2. **Lecturas de esfuerzos y defectos.** No se tocan `internal/effort` ni `internal/defect`; sus lecturas repetidas quedan como límite conocido, como las dejaron HU-33 y HU-47.

## Comparación con los criterios del Issue #53

Los criterios del Issue #53 estaban marcados como borrador y se cierran con CA-35.1 a CA-35.7 antes de la aprobación de esta SDD.

| CA del Issue #53 | En esta SDD |
|---|---|
| CA-35.1 Todas las métricas mínimas en una sola consulta | CA-35.1 |
| CA-35.2 Una métrica que no se puede calcular se indica como no disponible | CA-35.2 |
| CA-35.3 Un proyecto inexistente devuelve error | CA-35.3 (agrega el sprint inexistente, el ID no numérico y la falla de repositorio) |
| — | CA-35.4 a CA-35.7: lectura única del backlog, mismos valores, compatibilidad y contrato, y copias |

CA-35.4 a CA-35.6 vienen del acta y de los compromisos de las SDD de HU-32, HU-33, HU-34 y HU-47; CA-35.7 es una garantía de implementación.

## Decisiones tomadas en la redacción

1. **Endpoints existentes.** Se amplían las dos rutas de HU-32; no hay rutas nuevas. HU-36 sigue en `/resumen`.
2. **Tipos compuestos de los existentes.** `SprintMetrics` y `ProjectMetrics` contienen `SprintPoints`/`ProjectPoints`, `Completion`, `Hours` y `Defects` en lugar de repetir sus campos: así es evidente que no hay cálculos nuevos y no se renombran campos de HU-32 a HU-47. *Alternativa descartada:* un tipo plano con los diez valores, que duplicaría los tipos existentes.
3. **"No disponible" unificado en la API, no en el dominio.** Los tres indicadores del dominio (`VelocityAvailable`, `PercentAvailable`, `DeviationAvailable`) conservan su nombre, porque los usan los tests de HU-32, HU-33 y HU-34. La unificación es la regla RN-35.6 (siempre `null`, nunca 0 ni campo omitido) y una sola función de conversión en la API. *Alternativa descartada:* un campo nuevo con la lista de métricas no disponibles, que ningún criterio pide.
4. **Alcance ajustado al acta.** No se agregan lecturas nuevas en `effort` ni en `defect`, para no ampliar el alcance que el Planning acordó; sus lecturas repetidas quedan como límite conocido.
5. **M5 igual que HU-33.** Se reutilizan `ListBySprint` y `Summary.TotalHours` con la misma suma por sprint, así que la consulta única da exactamente el mismo `float64`.
6. **Operaciones separadas conservadas.** Las ocho operaciones de HU-32 a HU-47 se mantienen con sus tests y escenarios, que son la evidencia de esas historias. *Alternativa descartada:* eliminarlas, que obligaría a reescribir esos tests y escenarios en esta historia.
7. **Ejemplos con la salida real de `encoding/json`.** Las SDD anteriores escribían `20.0` por legibilidad; para HU-37, HU-38 y HU-39 se documenta lo que reciben de verdad (`20`), con la aclaración de que el formateo a un decimal es de quien muestra el valor.
8. **Interfaz sin cambios.** La sección "Métricas" ya muestra M1 a M10 con "sin datos" para las cuatro métricas que pueden no estar disponibles; el dashboard es de HU-37.
9. **Story Points.** 3 SP, según el Planning del Sprint 4; no se divide.
10. **Sin librerías nuevas.** Solo la librería estándar de Go; godog se mantiene en los tests.

HU-35: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
