# HU-33 · Calcular horas estimadas, horas reales y desviación

> Issue del repositorio: #51

Aprobada por: Peñalbé Hernán (Agile Enabler), 27/10/2026, con tres cambios respecto de su primera revisión: sin conteo de esfuerzos, sin motivo en la interfaz y orden con HU-47 (HU-33 integra primero).

## Historia de usuario

Como Agile Enabler, quiero calcular las horas estimadas, las horas reales y la desviación entre ambas, para mejorar las estimaciones futuras.

## Objetivo

Calcular las métricas M4 (horas estimadas), M5 (horas reales), M6 (desviación en horas) y M7 (desviación porcentual) de `docs/metricas.md` (TEC-03), por sprint y por proyecto, extendiendo el paquete `internal/metrics` de HU-32. Exponerlas en los dos endpoints de métricas existentes y en la sección "Métricas" de la interfaz web. HU-47 agrega los defectos después y HU-35 unifica la consulta; ninguna de las dos se implementa acá.

## Alcance

- **Incluye:** el cálculo de M4 a M7 de un sprint y de un proyecto, la cantidad de ítems sin horas estimadas, la extensión de `GET /api/sprints/{id}/metricas` y `GET /api/proyectos/{id}/metricas`, y la ampliación de la sección "Métricas" de la interfaz.
- **No incluye:** M1, M2 y M3 (HU-32), M8 (HU-34), M9 y M10 (HU-47), la consulta única de HU-35, el dashboard (HU-36 y HU-37), los reportes (HU-38, HU-39 y HU-40), una consulta de esfuerzos por proyecto nueva, cambios de esquema, congelar las horas al cerrar un sprint ni dividir HU-33.

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `SprintHours(sprintID int) (Hours, error)` | ID de un sprint | M4 a M7 del sprint |
| `ProjectHours(projectID int) (Hours, error)` | ID de un proyecto | M4 a M7 del proyecto |

Las dos operaciones están en el `Manager` de `internal/metrics`.

### Salidas esperadas

`Hours` (por valor):

| Campo | Tipo | Descripción |
|---|---|---|
| `Estimated` | `float64` | M4, sin redondear |
| `Actual` | `float64` | M5, sin redondear |
| `Deviation` | `float64` | M6 = M5 − M4, sin redondear; solo válido si `DeviationAvailable` |
| `DeviationPercent` | `float64` | M7 = (M5 − M4) / M4 × 100, sin redondear; solo válido si `DeviationAvailable` |
| `Unestimated` | `int` | Cantidad de ítems del conjunto con `EstimatedHours = 0` |
| `DeviationAvailable` | `bool` | `false` si M4 = 0 o Actual = 0 |

- `Estimated`, `Actual`, `Deviation` y `DeviationPercent` conservan el valor exacto del cálculo. El redondeo es de presentación (RN-33.7).
- Si `DeviationAvailable` es `false`, `Deviation` y `DeviationPercent` valen 0 en el dominio y la API los devuelve `null`.
- Un sprint o proyecto inexistente devuelve el error correspondiente (ver "Condiciones de error").

### HTTP

Se amplían los dos endpoints existentes de HU-32; no se crean rutas nuevas. Los campos de M1, M2, M3 y M8 se conservan sin cambios.

`GET /api/sprints/{id}/metricas`, con el ejemplo de TEC-03:

```json
{
  "sprint_id": 1,
  "sp_planificados": 10,
  "sp_completados": 8,
  "items_sin_estimar": 1,
  "items_totales": 4,
  "items_terminados": 2,
  "porcentaje_completadas": 50.0,
  "horas_estimadas": 20.0,
  "horas_reales": 20.0,
  "desviacion_horas": 0.0,
  "desviacion_porcentual": 0.0,
  "items_sin_horas_estimadas": 1
}
```

`GET /api/proyectos/{id}/metricas`, con el mismo proyecto de un solo sprint:

```json
{
  "proyecto_id": 1,
  "sp_planificados": 10,
  "sp_completados": 8,
  "items_sin_estimar": 1,
  "items_totales": 4,
  "items_terminados": 2,
  "porcentaje_completadas": 50.0,
  "velocidad": 8.0,
  "sprints_cerrados": 1,
  "horas_estimadas": 20.0,
  "horas_reales": 20.0,
  "desviacion_horas": 0.0,
  "desviacion_porcentual": 0.0,
  "items_sin_horas_estimadas": 1
}
```

- `horas_estimadas`, `horas_reales`, `desviacion_horas` y `desviacion_porcentual` se redondean a un decimal (TEC-03). En los ejemplos se escribe `20.0` por legibilidad; `encoding/json` puede emitir `20` para el mismo valor. La interfaz siempre muestra un decimal.
- `desviacion_horas` y `desviacion_porcentual` son `null` si M4 = 0 o si Actual = 0. Se usa `Actual > 0` como señal de que hay horas reales; el conteo de esfuerzos no se expone.
- `items_sin_horas_estimadas` es entero y cuenta los ítems con `EstimatedHours = 0`.
- Un valor positivo de la desviación significa que se trabajó más de lo estimado; uno negativo, menos (TEC-03).

| Caso | Código | Mensaje |
|---|---|---|
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| Sprint inexistente | 404 | El de `sprint.ErrSprintNotFound` |
| Proyecto inexistente | 404 | El de `project.ErrProjectNotFound` |
| Falla de un repositorio | 500 | El de `errInternal`, sin detalle |

### Interfaz web

Se mantiene la sección **"Métricas"** y el botón "Ver métricas". El texto del proyecto y el del sprint actual suman las horas:

```text
Horas: 20.0 estimadas, 20.0 reales, 1 ítem(s) sin horas estimadas. Desviación: 0.0 h (0.0 %).
```

Si la desviación no está disponible:

```text
Horas: 0.0 estimadas, 0.0 reales, 0 ítem(s) sin horas estimadas. Desviación: sin datos.
```

- Se muestran M4, M5, M6, M7 y la cantidad de ítems sin horas estimadas.
- Si M4 = 0 o Actual = 0, el texto de la desviación es `Desviación: sin datos.` y no se muestran M6 ni M7.
- La función `numero1` formatea con un decimal y convierte el texto `"-0.0"` en `"0.0"`.
- El sprint se muestra solo si el campo `sprint-actual` tiene un ID, como hoy.
- El motivo de la desviación no disponible no se muestra: se deduce de las horas estimadas y reales del mismo texto.
- Un error se muestra con `accion` y `msg-metricas`, igual que las demás acciones.

El texto inicial de `internal/web/index.html` cambia de:

```html
<p id="datos-metricas" class="vacio">Muestra los story points del proyecto actual y, si hay uno, del sprint actual.</p>
```

a:

```html
<p id="datos-metricas" class="vacio">Muestra los story points y las horas del proyecto actual y, si hay uno, del sprint actual.</p>
```

## Reglas de negocio

1. **RN-33.1 M4 de un sprint.** Es la suma de `EstimatedHours` de los ítems con `SprintID` igual al sprint. Un ítem con `EstimatedHours = 0` no suma y se cuenta en `Unestimated`, aunque esté terminado.
2. **RN-33.2 M4 de un proyecto.** Es la suma de `EstimatedHours` de todos los ítems del backlog del proyecto, asignados o no a un sprint (decisión de TEC-03). Los ítems con 0 no suman y se cuentan en `Unestimated`.
3. **RN-33.3 M5 de un sprint.** Es la suma de las horas de los registros de esfuerzo cuyo `SprintID` es el sprint, sin importar el estado del sprint. Se obtiene con `effort.Manager.ListBySprint` y se usa `Summary.TotalHours`.
4. **RN-33.4 M5 de un proyecto.** Es la suma de las horas de los registros de esfuerzo de todos los sprints del proyecto. Se obtiene con `sprint.Manager.ListByProject` y, para cada sprint, con `effort.Manager.ListBySprint`. Cada registro de esfuerzo pertenece a un sprint; no hay una consulta de esfuerzos por proyecto nueva ni cambios de esquema.
5. **RN-33.5 M6 y M7.** Si M4 > 0 y Actual > 0: `Deviation = Actual − Estimated` y `DeviationPercent = Deviation / Estimated × 100`. `DeviationAvailable` es `true`.
6. **RN-33.6 Desviación no disponible.** Si M4 = 0 o Actual = 0, `DeviationAvailable` es `false`, `Deviation` y `DeviationPercent` valen 0 en el dominio y la API devuelve `null`. No es un error. Las horas de un esfuerzo son siempre positivas (`effort.ErrInvalidHours`), así que Actual = 0 equivale a "sin esfuerzos registrados". M4 > 0 sigue siendo necesario para calcular M7; no alcanza con Actual > 0.
7. **RN-33.7 Redondeo.** El dominio no redondea. El CA-33.3 del Issue pide un decimal para M6 y M7; TEC-03 pide un decimal para M4, M5, M6 y M7 en la presentación. Por eso la API redondea M4–M7 a un decimal con `math.Round(valor*10)/10` y la interfaz los muestra con `toFixed(1)`. Los conteos son enteros.
8. **RN-33.8 Cero negativo.** Si el redondeo de M6 o M7 da `-0`, la API devuelve `0` positivo. La interfaz convierte `"-0.0"` en `"0.0"`. Ejemplo: dos ítems estimados en `0.1` y `0.2` y un esfuerzo de `0.3` producen M4 = `0.30000000000000004`, M5 = `0.3`, M6 ≈ `−5,55e−17` y M7 ≈ `−1,85e−14`; se informan como `0` y `0.0`, no como `-0` ni `-0.0`.
9. **RN-33.9 Datos actuales y solo lectura.** Las métricas se calculan en el momento de la consulta con los datos guardados. Consultar no modifica el backlog, los sprints ni los esfuerzos.
10. **RN-33.10 Copias por valor.** `Hours` se devuelve por valor y no contiene punteros ni listas compartidas. Ningún llamador puede mutar el estado interno del `Manager` con el resultado.
11. **RN-33.11 Compatibilidad.** Los campos y el comportamiento de M1, M2, M3 y M8 no cambian. Los errores y los códigos HTTP existentes se mantienen.
12. **RN-33.12 Sin errores nuevos.** No se agregan errores de dominio. La falta de datos se representa con `DeviationAvailable = false` y `null` en la API.

## Decisión sobre entidades referenciadas

- **Sprint:** sí se valida su existencia con `sprint.Manager.Get`, incluidos 0 y negativos (`ErrSprintNotFound`). Sin sprint no se puede obtener el proyecto de sus ítems ni distinguir un sprint inexistente de uno vacío.
- **Proyecto:** sí se valida, a través de `backlog.Manager.ListItems`, que ya devuelve `project.ErrProjectNotFound` (HU-06). No se agrega otra consulta.
- **Ítems:** no se validan uno por uno. Se leen con `ListItems` del proyecto del sprint o del proyecto consultado.
- **Esfuerzos:** no se validan uno por uno. Se leen con `effort.Manager.ListBySprint`, que valida el sprint y devuelve una copia de los registros y su total exacto. Una falla del repositorio de esfuerzos se propaga.

## Restricciones

- **Organización del código.** En `internal/metrics`: `hours.go` (tipo `Hours`, cálculo puro y las dos operaciones), `manager.go` (el campo `efforts` y la firma nueva) y los tests `hours_test.go` y `hours_features_test.go`. En `internal/httpapi`: `metrics.go` (campos, conversión y handlers) y `server.go` (creación del `Manager`). En `internal/web`: `app.js` (texto y formateador) e `index.html` (texto inicial de la sección). No se modifica `internal/effort`.
- **Firma nueva y archivos que cambia (para HU-47).**

  ```go
  func NewManager(items *backlog.Manager, sprints *sprint.Manager, efforts *effort.Manager) *Manager
  ```

  | Archivo | Cambio |
  |---|---|
  | `internal/metrics/manager.go` | `Manager` suma `efforts *effort.Manager`; `NewManager` recibe `efforts` |
  | `internal/metrics/hours.go` | Nuevo: `Hours`, `SprintHours` y `ProjectHours` |
  | `internal/httpapi/server.go` | `metrics.NewManager(items, sprints, efforts)`; `NewHandler` no cambia de firma porque ya recibe `efforts` |
  | `internal/httpapi/metrics.go` | Los dos handlers suman M4 a M7 y los cinco campos de respuesta |
  | `internal/web/app.js` | Texto de horas, desviación y `numero1` |
  | `internal/web/index.html` | Texto inicial de "Métricas" |
  | `internal/metrics/story_points_test.go` | `newMetricsEnvWith` recibe el repositorio de esfuerzos y crea el `effort.Manager`; `metricsEnv` guarda `effortRepo` |
  | `internal/metrics/completion_test.go` | Actualiza las llamadas a `newMetricsEnvWith` |
  | `internal/httpapi/metrics_test.go` | Casos de M4 a M7, `null`, redondeo, cero negativo y error del repositorio de esfuerzos |
  | `internal/httpapi/web_test.go` | Verifica el texto de `app.js` y el formateador `-0.0` → `0.0` |

- **HU-47 después de HU-33.** HU-47 agrega defectos y cambiará la firma a `NewManager(items, sprints, efforts, defects)`. Debe traer `main` antes de su RED y adaptar `NewHandler`, `newMetricsEnvWith`, los tests de API y los de interfaz. HU-33 no agrega defectos.
- **Sin persistencia propia.** `internal/metrics` no tiene repositorio ni estado: lee de `backlog`, `sprint` y `effort`. No hay tablas nuevas.
- **Concurrencia.** `metrics` no tiene mutex. Cada lectura queda protegida por el mutex del Manager que la hace. Las lecturas de ítems, sprints y esfuerzos no son atómicas entre sí.
- **Copias por valor.** Todo se devuelve por valor; no se exponen punteros ni listas internas.
- **Tests.** Unitarios en `internal/metrics/hours_test.go`; godog en `hours_features_test.go` con el runner `TestFeatures_HU33CalcularHorasDesviacion`, sobre `features/HU-33-horas-desviacion.feature`; API en `internal/httpapi/metrics_test.go`; interfaz en `internal/httpapi/web_test.go`. La interfaz se verifica además manualmente en el navegador con el caso de cero negativo y se guarda la captura en `docs/evidencias/img/HU-33/cero-negativo.png`, referenciada en `docs/evidencias/HU-33-tdd.md`.
- **Profundidad proporcional (3 SP).** Un test por regla y por código de error; sin pruebas de mutación.
- **Librerías.** Solo la librería estándar; godog en los tests.

### Límite conocido

- El proyecto consulta los esfuerzos sprint por sprint: son N+1 lecturas para N sprints. Es aceptable para el alcance y el tamaño actual; HU-35 puede optimizarlo al unificar la consulta.
- Las lecturas de M1/M2, M8, M4 y M5 no son atómicas entre sí: un cambio simultáneo puede quedar reflejado en una y no en la otra.
- No se congelan las horas al cerrar un sprint: si se reestima un ítem, M4 cambia. TEC-03 define que las métricas usan los datos actuales.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Ejemplo de TEC-03: A (10 h, esfuerzo 12), B (6 h, esfuerzo 5), C (0 h, esfuerzo 3), D (4 h, sin esfuerzo) | M4 = 20, M5 = 20, M6 = 0, M7 = 0; 1 ítem sin horas |
| Sprint sin ítems | M4 = 0, M5 = 0, desviación no disponible; 0 ítems sin horas |
| Proyecto sin ítems | M4 = 0, M5 = 0, desviación no disponible; 0 ítems sin horas |
| Ítem con `EstimatedHours = 0` | No suma en M4; se cuenta en `items_sin_horas_estimadas` |
| Ítem estimado sin esfuerzo | M4 > 0, Actual = 0; desviación no disponible |
| Esfuerzo sin horas estimadas | M4 = 0, Actual > 0; desviación no disponible |
| Sin ningún esfuerzo | Actual = 0; desviación no disponible |
| Sprint con esfuerzos de otro sprint | No se incluyen; M5 usa solo los registros de ese sprint |
| Proyecto con ítems sin asignar a sprint | Suman en M4 del proyecto; no en M4 de ningún sprint |
| Proyecto con varios sprints | M5 suma los esfuerzos de todos sus sprints |
| M4 = 20 y M5 = 25 | M6 = 5, M7 = 25.0 |
| M4 = 20 y M5 = 15 | M6 = −5, M7 = −25.0 |
| M6 = 1.25 | La API devuelve 1.3 |
| M7 = 12.34 | La API devuelve 12.3 |
| Dos ítems de 0.1 y 0.2 h y un esfuerzo de 0.3 h | M6 ≈ −5,55e−17 y M7 ≈ −1,85e−14; API 0 y 0.0; interfaz 0.0 |
| Sprint o proyecto inexistente, 0 o negativo | `ErrSprintNotFound` o `ErrProjectNotFound` |
| Falla del repositorio de esfuerzos | Se devuelve ese error; la API responde 500 |
| HTTP: `{id}` no numérico | 400 |

## Condiciones de error

| Condición | Error |
|---|---|
| No existe el sprint (incluidos 0 y negativos) | `sprint.ErrSprintNotFound` |
| No existe el proyecto (incluidos 0 y negativos) | `project.ErrProjectNotFound` |
| Falla el repositorio de ítems, de sprints o de esfuerzos | El error del repositorio |
| La desviación no está disponible (M4 = 0 o Actual = 0) | No es un error: `DeviationAvailable = false` y `null` en la API |

## Criterios de aceptación

- **CA-33.1** Las horas estimadas (M4) son la suma de las horas estimadas de los ítems del sprint o proyecto; los ítems con 0 no suman y se informa cuántos hay.
- **CA-33.2** Las horas reales (M5) son la suma de las horas registradas (en el sprint, solo los registros de ese sprint).
- **CA-33.3** La desviación se informa en horas (M6 = M5 − M4) y en porcentaje (M7 = (M5 − M4) / M4 × 100), redondeadas a un decimal.
- **CA-33.4** Sin horas estimadas (M4 = 0) o sin horas reales (Actual = 0), M6 y M7 quedan no disponibles (null), sin error.
- **CA-33.5** Se exponen en `GET /api/sprints/{id}/metricas` y `GET /api/proyectos/{id}/metricas` y se muestran en la sección Métricas de la interfaz.
- **CA-33.6** En el proyecto, M4 suma todo el backlog, asignado o no a un sprint, y M5 suma los esfuerzos de todos sus sprints.
- **CA-33.7** Un sprint o un proyecto sin ítems devuelve M4 = 0, M5 = 0 y desviación no disponible, sin error.
- **CA-33.8** Un sprint o un proyecto inexistente devuelve su error; una falla de repositorio se propaga; consultar no modifica nada.
- **CA-33.9** El dominio no redondea; la API y la interfaz muestran un decimal. Un cero negativo de M6 o M7 se normaliza a `0`/`0.0`.
- **CA-33.10** Lo devuelto es una copia por valor; modificarlo no altera el estado guardado.

## Comparación con el acta de Planning

**No hay diferencias de alcance con el acta del Sprint 4.** El acta asigna HU-33 con **3 SP**, sin división, y define el alcance como extender el paquete de métricas de HU-32 con horas estimadas, horas reales y desviación, reutilizando la consulta de esfuerzos de HU-27. La SDD hace exactamente eso.

Precisiones registradas:

1. **HU-32 en dependencias.** El acta no la lista en "Depende de", pero su justificación asume el paquete y los endpoints de HU-32, que ya están en `main`. No cambia el alcance.
2. **Redondeo.** El CA-33.3 del Issue pide un decimal para M6 y M7; TEC-03 pide un decimal para M4–M7 en la presentación. El dominio no redondea y la SDD aplica el redondeo solo en API e interfaz.
3. **Orden con HU-47.** Acordado con el Agile Enabler: HU-33 va primero. HU-47 trae `main` antes de su RED y adapta `NewManager`, `NewHandler`, `newMetricsEnvWith` y los tests de API e interfaz.

## Comparación con los criterios del Issue #51

La SDD usa la numeración CA-33.1 a CA-33.5 para los criterios del Issue #51 y agrega los criterios de borde:

| CA del Issue #51 | En esta SDD |
|---|---|
| CA-33.1 M4 | CA-33.1 y CA-33.6 |
| CA-33.2 M5 | CA-33.2 y CA-33.6 |
| CA-33.3 M6 y M7 | CA-33.3 y CA-33.9 |
| CA-33.4 Desviación no disponible | CA-33.4 |
| CA-33.5 Endpoints e interfaz | CA-33.5 |
| — | CA-33.6 a CA-33.10: detalle del proyecto, sin ítems, errores, redondeo/cero negativo y copias |

No se amplía el alcance: CA-33.6 a CA-33.10 son casos de borde, errores y garantías de copia necesarios para la implementación y las pruebas.

## Decisiones tomadas en la redacción

1. **Endpoints existentes.** HU-33 amplía los endpoints creados por HU-32; no crea rutas nuevas. HU-47 suma defectos y HU-35 unifica la consulta.
2. **Sin consulta nueva de esfuerzos.** El nivel proyecto recorre los sprints y usa `ListBySprint`; no se agrega un método al repositorio de esfuerzos ni se cambia el esquema. Se acepta el N+1 y se documenta como límite conocido.
3. **Dominio exacto, presentación con un decimal.** TEC-03 calcula sin redondear y muestra un decimal. El redondeo de M4, M5, M6 y M7 va en la API y en la interfaz.
4. **Cero negativo.** Se normaliza en la API y en la interfaz para no mostrar `-0` ni `-0.0`; el caso se prueba con `0.1 + 0.2` y esfuerzo `0.3`.
5. **Coordinación con HU-47.** La firma nueva queda documentada con los archivos que cambia. HU-47 la adapta después de traer `main`.
6. **Story Points.** 3 SP, según el Planning del Sprint 4; no se divide.
7. **Sin librerías nuevas.** Solo la librería estándar de Go; godog se mantiene en los tests.
8. **Sin conteo de esfuerzos.** No se expone `Efforts` ni `esfuerzos_registrados`: ningún CA lo pide y el conteo era un dato interno que no se usa para calcular M4–M7. La disponibilidad se decide con `Actual > 0`; como las horas de un esfuerzo son siempre positivas, `Actual = 0` equivale a "sin esfuerzos registrados".

HU-33: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
