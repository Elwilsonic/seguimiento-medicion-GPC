# HU-37 · Ver las métricas en el dashboard con gráficos

> Issue del repositorio: #55

Aprobada por: Peñalbé Hernán (Agile Enabler), 29/10/2026, con la ruta nueva aceptada como diferencia con el acta y tres ajustes (línea de aprobación, función de conversión de httpapi nombrada y orden de recorte de los gráficos), que se aplicaron.

## Historia de usuario

Como Agile Enabler, quiero ver las métricas en el dashboard con representaciones gráficas, para decidir con datos.

## Objetivo

Mostrar en la sección "Dashboard" de la interfaz web las métricas mínimas del proyecto actual (M1 a M10 de `docs/metricas.md`, TEC-03) y tres gráficos por sprint: Story Points completados por sprint cerrado con la velocidad (M3), Story Points planificados y completados por sprint (M1 y M2), y horas estimadas y reales por sprint (M4 y M5, con la desviación M6 y M7). Para tener los datos por sprint se agrega una consulta en Go que devuelve, en una sola operación, cada sprint del proyecto con sus métricas, reutilizando los cálculos de HU-35 sin agregar cálculos nuevos. Los gráficos se dibujan como SVG con JavaScript propio, sin librerías externas (tema 4 del acta de Planning del Sprint 4).

## Alcance

- Incluye:
  - la operación `ProjectSprints` de `internal/metrics`;
  - la ruta `GET /api/proyectos/{id}/metricas/sprints`;
  - el botón "Ver gráficos" de la sección "Dashboard", con el texto de las métricas del proyecto y los tres gráficos.
- No incluye:
  - un gráfico de defectos ni de porcentaje de completadas (M8 a M10 se muestran como texto);
  - filtros por fecha, por estado del sprint o por cantidad de sprints;
  - cambios en las rutas existentes, en los repositorios, en el esquema ni en los paquetes `backlog`, `sprint`, `effort` y `defect`;
  - los reportes (HU-38 y HU-39).

### Orden de recorte

Por el acuerdo previo del acta de Planning del Sprint 4 ("menos gráficos en HU-37") y porque el desarrollo se cierra el 30/10/2026, el trabajo se hace en este orden. Si no entra todo, se recorta desde el final:

1. La ruta `GET /api/proyectos/{id}/metricas/sprints` con `ProjectSprints`, el texto de M1 a M10 en el Dashboard y el gráfico "Story Points completados por sprint cerrado" con la velocidad.
2. El gráfico "Story Points planificados y completados por sprint".
3. El gráfico "Horas estimadas y reales por sprint".

Un gráfico recortado se saca también de los criterios, los escenarios y los tests que lo piden, y el recorte se registra en la evidencia.

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `ProjectSprints(projectID int) ([]SprintWithMetrics, error)` | ID de un proyecto | Cada sprint del proyecto con M1, M2 y M4 a M10 de ese sprint |

La operación está en el `Manager` de `internal/metrics`.

### Salidas esperadas

Una lista nueva de `SprintWithMetrics`, por valor:

| Campo | Tipo | Origen |
|---|---|---|
| `Sprint` | `sprint.Sprint` | Copia del sprint, de `sprint.Manager.ListByProject` |
| `Metrics` | `SprintMetrics` | Las mismas métricas que devuelve `Sprint(id)` de HU-35 |

- Ningún campo tiene punteros, listas ni mapas: cada elemento es una copia completa, y la lista es nueva en cada llamada.
- La lista está ordenada por ID de sprint ascendente, con sprints abiertos y cerrados.
- Un proyecto sin sprints devuelve una lista vacía, no `nil`.
- Los valores no se redondean en el dominio (TEC-03). El redondeo es de la API.

### HTTP (TEC-04)

`GET /api/proyectos/{id}/metricas/sprints`. Cada elemento reutiliza los dos formatos que ya existen: `sprint` con `sprintResponse` y `toSprintResponse` (HU-10), y `metricas` con `sprintMetricsResponse` (HU-32 a HU-35), sin campos nuevos.

```json
[
  {
    "sprint": {
      "id": 1,
      "proyecto_id": 1,
      "nombre": "Sprint 1",
      "fecha_inicio": "2026-10-01",
      "fecha_fin": "2026-10-14",
      "objetivo": "Primer incremento",
      "estado": "CERRADO"
    },
    "metricas": {
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
  }
]
```

| Caso | Código | Cuerpo |
|---|---|---|
| Consulta exitosa, aunque alguna métrica no esté disponible | 200 | La lista de arriba |
| Proyecto sin sprints | 200 | `[]` |
| `{id}` que no es un número entero | 400 | `{"error": "el id debe ser un número entero"}` (`errInvalidID`) |
| Proyecto inexistente (incluidos 0 y negativos) | 404 | `{"error": "no existe un proyecto con ese id"}` (`project.ErrProjectNotFound`) |
| Falla de cualquier repositorio | 500 | `{"error": "error interno del servidor"}` (`errInternal`), sin detalle |

### Interfaz web (TEC-04)

En la sección "Dashboard" de HU-36 se agrega:

- El botón "Ver gráficos" (`id="ver-graficos"`).
- El párrafo `id="datos-graficos"`, que muestra el texto de las métricas del proyecto.
- El contenedor `id="graficos"`, donde van los tres gráficos.
- El mensaje `id="msg-graficos"`.

Al tocar "Ver gráficos", la interfaz:

1. Pide `GET /api/proyectos/{id}/metricas` (HU-35) y `GET /api/proyectos/{id}/metricas/sprints` con el proyecto actual.
2. Muestra en `datos-graficos` M1 a M10 del proyecto, con el mismo texto que la sección "Métricas":
   - M4 a M10 salen de `historiasCompletadas`, `horasYDesviacion` y `defectos`, que ya existen en `app.js`.
   - El texto de M1 a M3 (SP planificados y completados, ítems sin estimar y velocidad con sus sprints cerrados) hoy está escrito dentro del handler de "Ver métricas". Se extrae a una función compartida, que usan "Ver métricas" y "Ver gráficos".
   - El texto que muestra "Ver métricas" no cambia, y los tests de HU-32 a HU-35 tienen que seguir pasando sin cambios.
3. Dibuja en `graficos` tres gráficos SVG de barras:
   - **"Story Points completados por sprint cerrado"**: una barra por sprint `CERRADO` con su M2, y una línea horizontal de referencia con la velocidad (M3) del proyecto, rotulada "Velocidad: X SP".
   - **"Story Points planificados y completados por sprint"**: dos barras por sprint (M1 y M2), abiertos y cerrados.
   - **"Horas estimadas y reales por sprint"**: dos barras por sprint (M4 y M5), con la desviación M6 y M7 debajo del nombre del sprint, o "sin datos".

Cada gráfico tiene:

- Un título y una leyenda.
- El valor escrito sobre cada barra. Los Story Points son enteros. Las horas se formatean con `numero1` de `app.js`, que las muestra con un decimal y ya corrige el cero negativo ("-0.0" se muestra como "0.0").
- Un elemento `<title>` para lectores de pantalla.

Las barras se escalan al mayor valor del gráfico. Los colores salen de variables CSS nuevas en `styles.css`, junto a las que ya existen.

## Reglas de negocio

- **RN-37.1** `ProjectSprints` devuelve, por cada sprint del proyecto, las mismas métricas que `Sprint(id)` de HU-35 para ese sprint: mismos valores y mismo criterio de "no disponible" (CA-35.5).
- **RN-37.2** Lecturas y orden de lectura:
  - `ProjectSprints` lee una sola vez los ítems del proyecto con `backlog.Manager.ListItems`, que valida el proyecto.
  - Después lee una sola vez la lista de sprints con `sprint.Manager.ListByProject`.
  - Luego lee los esfuerzos de cada sprint con `effort.Manager.ListBySprint`, en orden de ID.
  - Por último lee una sola vez los defectos del proyecto con `defect.Manager.ListByProject`.
  - Es el mismo orden que `Project(id)` de HU-35, y la operación corta en el primer error.
- **RN-37.3** Las métricas de cada sprint se calculan con las funciones puras que ya existen: `itemsOfSprint`, `storyPoints`, `completion`, `hoursOf` y `sprintDefects`. No se agregan cálculos ni fórmulas.
- **RN-37.4** La lista tiene todos los sprints del proyecto, abiertos y cerrados, ordenados por ID ascendente, que es el orden de `ListByProject`.
- **RN-37.5** El gráfico de velocidad usa solo los sprints `CERRADO`, porque M3 se calcula con sprints cerrados (TEC-03). La línea de referencia es el valor `velocidad` de `GET /api/proyectos/{id}/metricas`: el promedio de los últimos 3 sprints cerrados. Si `velocidad` es `null`, el gráfico se reemplaza por el texto "Velocidad: sin datos (no hay sprints cerrados)."
- **RN-37.6** Una métrica no disponible (`null` en la API) se muestra como "sin datos", igual que en la sección "Métricas" (HU-32, HU-33 y HU-34). En el gráfico de horas, una desviación `null` se muestra como "Desviación: sin datos" debajo de ese sprint, y sus barras M4 y M5 se dibujan igual, porque M4 y M5 nunca son `null`.
- **RN-37.7** Si el proyecto no tiene sprints, el contenedor de gráficos muestra "Sin sprints: no hay datos para los gráficos." en lugar de los tres gráficos, y el texto de las métricas del proyecto se muestra igual.
- **RN-37.8** Las barras se escalan al mayor valor de cada gráfico. Si todos los valores son 0, las barras tienen altura 0 y el rótulo "0" o "0.0" se muestra igual (sin división por cero).
- **RN-37.9** Los gráficos se dibujan con `document.createElementNS` y el espacio de nombres SVG, sin librerías externas ni recursos de otros dominios (tema 4 del acta y `AGENTS.md`). Los textos se asignan con `textContent`, nunca con `innerHTML`, así que el nombre de un sprint no se interpreta como HTML.
- **RN-37.10** Consultar no modifica ningún dato.

## Decisión sobre entidades referenciadas

- **Se valida que el proyecto exista.** Lo valida `backlog.Manager.ListItems` y devuelve `project.ErrProjectNotFound`, igual que `Project(id)` de HU-35, porque sin esa validación un proyecto inexistente respondería `[]` y no se distinguiría de un proyecto sin sprints.
- **No se validan los sprints uno por uno.** Salen de `sprint.Manager.ListByProject` del mismo proyecto, así que existen y pertenecen a él por construcción. Por eso no se llama a `sprints.Get` por cada uno.

## Restricciones

- Lecturas por consulta: el backlog una vez desde `metrics` (más la lectura interna de `defect.Manager.ListByProject`, como en HU-35), los sprints una vez, los esfuerzos una vez por sprint y los defectos una vez. No se llama a `Sprint(id)` por cada sprint, porque repetiría la lectura del backlog y de los defectos por cada uno.
- **Límite conocido (heredado de HU-33):** `effort.Manager.ListBySprint` vuelve a leer cada sprint con `sprint.Manager.Get` para validarlo. Con N sprints son N lecturas de sprints además de la de `ListByProject`. HU-37 no toca `internal/effort`, igual que lo documenta la SDD de HU-35.
- Las lecturas de fuentes distintas no son atómicas entre sí, igual que en HU-35.
- `ProjectSprints` devuelve copias por valor y una lista nueva. Modificar la lista devuelta no cambia lo que devuelve una consulta posterior.
- Sin librerías externas, en Go ni en JavaScript. El servidor usa solo `net/http` (TEC-01).
- `NewHandler` y `metrics.NewManager` no cambian de firma.
- La sección "Métricas" y las rutas existentes no cambian.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Proyecto sin sprints | 200 y `[]`. En la interfaz, el texto de las métricas y "Sin sprints: no hay datos para los gráficos." |
| Proyecto con sprints, todos abiertos (ninguno cerrado) | Lista con todos. En la interfaz se muestran los gráficos de planificados y completados y de horas, y "Velocidad: sin datos (no hay sprints cerrados)." |
| Sprint sin ítems | Ese elemento con M1 y M2 en 0, `porcentaje_completadas` en `null`, horas en 0 y desviación en `null` |
| Sprint con horas estimadas pero sin esfuerzos | `horas_reales` en 0 y desviación en `null` ("Desviación: sin datos" en el gráfico) |
| Más de 3 sprints cerrados | El gráfico de velocidad muestra una barra por cada sprint cerrado; la línea de referencia sigue siendo M3, que promedia los últimos 3 |
| Todos los valores de un gráfico en 0 | Barras de altura 0 con su rótulo; sin división por cero |
| Ítem sin sprint | No cuenta en ningún sprint de la lista; sí en las métricas del proyecto |
| Nombre de sprint con caracteres como `<` o `&` | Se muestra como texto (`textContent`) |
| IDs 0 y negativos | 404, proyecto inexistente |

## Condiciones de error

| Condición | Dominio | HTTP |
|---|---|---|
| Proyecto inexistente | `project.ErrProjectNotFound` | 404 |
| `{id}` no numérico | — | 400 (`errInvalidID`) |
| Falla al leer ítems, sprints, esfuerzos o defectos | El error del repositorio, sin una lista parcial | 500, sin detalle |
| Falla de cualquiera de las dos consultas en la interfaz | — | `msg-graficos` muestra el error, y no se dibujan gráficos con datos incompletos |

## Comparación con el acta de Planning del Sprint 4

El acta describe HU-37 como "Gráficos con las métricas de HU-35, dibujados sin librerías externas" (5 SP), con SVG generado con JavaScript propio en `internal/web` (tema 4). Además, la SDD de HU-35 dice que "HU-37 construye el dashboard sobre el mismo contrato". Esta SDD respeta el SVG y la ausencia de librerías, con **una diferencia**: agrega una ruta nueva en Go (`GET /api/proyectos/{id}/metricas/sprints`) y la operación `metrics.ProjectSprints`. La ruta reutiliza los cálculos y los formatos de HU-35, pero es contrato nuevo. El motivo es que el contrato de HU-35 no permite listar las métricas de cada sprint de un proyecto, y el ejemplo del Issue ("velocidad por sprint") las necesita.

## Comparación con los criterios iniciales del Issue #55

| Criterio del Issue | Criterio final |
|---|---|
| CA-37.1 El dashboard muestra las métricas mínimas, con al menos algunas representadas en gráficos (por ejemplo, velocidad por sprint). | CA-37.1, que nombra los tres gráficos |
| CA-37.2 Una métrica no disponible se muestra como tal. | CA-37.2, con "sin datos" y el caso sin sprints |
| CA-37.3 Los valores coinciden con los de la épica de métricas. | CA-37.3 |
| — | CA-37.4 a CA-37.6: la ruta nueva, la consulta sin modificar nada y los gráficos sin librerías |

Los criterios del Issue #55 estaban marcados como borrador y se cierran con los de la sección "Criterios de aceptación" antes de la aprobación de esta SDD.

## Criterios de aceptación

- **CA-37.1** La sección "Dashboard", con el botón "Ver gráficos", muestra M1 a M10 del proyecto actual y tres gráficos SVG: Story Points completados por sprint cerrado con la velocidad (M3) como línea de referencia, Story Points planificados y completados por sprint, y horas estimadas y reales por sprint con su desviación.
- **CA-37.2** Una métrica no disponible se muestra como "sin datos": la velocidad sin sprints cerrados, el porcentaje sin ítems y la desviación sin horas estimadas o reales. Un proyecto sin sprints muestra "Sin sprints: no hay datos para los gráficos.", y el resto de las métricas se muestra igual.
- **CA-37.3** Los valores de cada sprint coinciden con los de `GET /api/sprints/{id}/metricas`, y los del proyecto con los de `GET /api/proyectos/{id}/metricas` (HU-35).
- **CA-37.4** `GET /api/proyectos/{id}/metricas/sprints` devuelve cada sprint del proyecto con sus métricas, ordenados por ID. Un proyecto sin sprints devuelve `[]`, un proyecto inexistente 404, un ID no numérico 400 y una falla de un repositorio 500 sin detalle.
- **CA-37.5** Consultar no modifica ningún dato, y la consulta devuelve copias: modificar la lista devuelta no cambia una consulta posterior.
- **CA-37.6** Los gráficos se dibujan con JavaScript propio (SVG), sin librerías externas, y los textos se muestran como texto, no como HTML.

## Pruebas previstas

- **Unitarias** en `internal/metrics/sprints_test.go`:
  - el ejemplo de TEC-03 con un sprint cerrado y otro abierto;
  - igualdad con `Sprint(id)` para cada sprint (CA-37.3);
  - proyecto sin sprints (`[]`, no `nil`), proyecto inexistente con IDs 0, negativos y 99, y un sprint sin ítems o sin esfuerzos (CA-37.2);
  - orden por ID, errores en el orden de lectura de RN-37.2 y una sola lectura del backlog desde `metrics`;
  - que no modifica nada y devuelve copias (CA-37.5).
- **BDD con godog** en `features/HU-37-graficos-dashboard.feature`, con el runner en `internal/metrics/sprints_features_test.go`: happy path, alternativo (sprints solo abiertos), límite (sin sprints, todo en 0) y error (proyecto inexistente).
- **API** en `internal/httpapi/metrics_test.go`:
  - el JSON del ejemplo comparado crudo con `json.Compact`, incluido el orden de los campos;
  - `[]`, 400, 404 y 500;
  - una fila en `TestRutasConID_IDNoNumericoDevuelve400`.
- **Interfaz** en `internal/httpapi/web_test.go`, con el mismo estilo que HU-36 (contenido de `index.html` y `app.js`):
  - ids y botón de la sección;
  - las dos rutas;
  - que existe la función compartida con el texto de M1 a M3 y que la usan los handlers de "Ver métricas" y de "Ver gráficos";
  - `createElementNS` con el espacio de nombres SVG y uso de `textContent`;
  - los tres títulos;
  - los textos "sin datos" y "Sin sprints";
  - sin `<script src=` a otros dominios.
- **Límite conocido:** el repositorio no tiene un motor de JavaScript en los tests, así que el dibujo de los gráficos (escala y barras) no se ejecuta en `go test`. Se verifica a mano en el navegador con el servidor real, y la prueba queda registrada en la evidencia.

## Archivos que cambian

- **Nuevos:**
  - `internal/metrics/sprints.go`;
  - `internal/metrics/sprints_test.go` y `sprints_features_test.go`;
  - `features/HU-37-graficos-dashboard.feature`.
- **Modificados:**
  - `internal/metrics/summary.go`, para extraer el cálculo de un sprint a una función pura que usan `Sprint` y `ProjectSprints`, sin cambiar `Sprint`;
  - `internal/httpapi/metrics.go`:
    - el handler y la respuesta de la ruta nueva;
    - `toSprintMetricsResponse(metrics.SprintMetrics) sprintMetricsResponse`, extraída del handler `sprintMetrics` de HU-35, que hoy arma esa respuesta en línea. La usan `sprintMetrics` y la ruta nueva. La respuesta de `GET /api/sprints/{id}/metricas` y sus tests de HU-32 a HU-35 no cambian;
  - `server.go` (ruta), `metrics_test.go`, `web_test.go` y `server_test.go`;
  - `internal/web/index.html` y `styles.css`;
  - `internal/web/app.js`: el handler de "Ver gráficos", el dibujo de los gráficos y la función compartida con el texto de M1 a M3, que se extrae del handler de "Ver métricas" sin cambiar lo que muestra;
  - README (ruta, sección del Dashboard y fila de HU-37), `docs/reportes/cobertura.md` y `docs/ia/registro.md`.

## Decisiones tomadas en la redacción

1. **Ruta nueva en Go en lugar de N llamadas desde JavaScript.** El cálculo queda en Go, con TDD y godog reales, y la interfaz solo dibuja. Decidido por Buttini Cristobal.
2. **Tres gráficos:** SP completados por sprint cerrado con M3, SP planificados y completados, y horas estimadas y reales. Decidido por Buttini Cristobal. Defectos y porcentaje van como texto.
3. **Formato de cada elemento como `{sprint, metricas}`**, reutilizando `sprintResponse` y `sprintMetricsResponse` sin campos nuevos, como pidió el Agile Enabler en HU-36 con `sprint_actual`.
4. **"sin datos" en lugar de "no disponible"**, para que la interfaz use el mismo texto que la sección "Métricas".
5. **Un botón aparte ("Ver gráficos")**, sin cambiar "Ver resumen" de HU-36, para que los tests y el comportamiento de HU-36 no cambien.
