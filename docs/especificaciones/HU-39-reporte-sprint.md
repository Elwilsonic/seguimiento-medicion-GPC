# HU-39 · Generar el reporte de un sprint

> Issue del repositorio: #57

Aprobada por: pendiente (Agile Enabler).

## Historia de usuario

Como Agile Enabler, quiero generar el reporte de un sprint, para documentar su resultado.

## Objetivo

Reunir en una sola consulta el reporte de un sprint: sus datos (nombre, fechas, Sprint Goal y estado), sus métricas M1, M2 y M4 a M10 (HU-35), sus esfuerzos (HU-27), sus defectos (HU-31) y los integrantes que registraron esfuerzo en el sprint (HU-26 y HU-27). Reutiliza la estructura del reporte de HU-38, como prevé su SDD ("Contrato" y decisiones 3 y 4): una operación nueva `Sprint(sprintID)` en el `Manager` de `internal/report`, con los mismos tipos `Efforts`, `Defects` y `Members`, el endpoint `GET /api/sprints/{id}/reporte`, con cada parte en el formato de un endpoint que ya existe, y un botón en la sección "Reportes" de la interfaz web. No hay cálculos nuevos: cada parte sale de una operación que ya existe.

## Alcance

- **Incluye:**
  - la operación `Sprint(sprintID int) (SprintReport, error)` y el tipo `SprintReport` en `internal/report`, reutilizando `Efforts`, `Defects`, `Members` y `membersOf` de HU-38;
  - el endpoint `GET /api/sprints/{id}/reporte`, con su respuesta formada por las respuestas que ya existen de sprints, métricas, esfuerzos y defectos, y la de integrantes de HU-38;
  - el botón "Generar reporte del sprint" en la sección "Reportes", que usa el sprint actual;
  - el contrato del reporte, con ejemplos JSON, para HU-40.
- **No incluye:**
  - la exportación a PDF (HU-40) ni librerías externas;
  - cambios en el reporte del proyecto (HU-38): `ProjectReport`, `Project`, `GET /api/proyectos/{id}/reporte` y el botón "Generar reporte" no cambian;
  - integrantes del equipo como entidad (E4, HU-15 a HU-17, fuera de la entrega; tema 5 del Planning del Sprint 4): el reporte lista solo los integrantes que registraron esfuerzo en el sprint;
  - la lista de ítems del sprint: las historias planificadas y completadas y sus estimaciones se informan con las métricas M1, M2, M4 y M8 y sus cantidades, como en HU-38 (decisión 9);
  - cálculos o métricas nuevas, ni totales por integrante o por ítem;
  - cambios en `internal/sprint`, `internal/backlog`, `internal/metrics`, `internal/effort`, `internal/defect` o `internal/dashboard`, ni en sus repositorios o tests;
  - cambios en los endpoints existentes ni en la sección "Métricas";
  - cambios de esquema, tablas nuevas o filtros.

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `Sprint(sprintID int) (SprintReport, error)` | ID de un sprint, abierto o cerrado | Reporte completo del sprint |

Está en el `Manager` de `internal/report` de HU-38. `NewManager` no cambia: ya recibe los `Manager` de sprints, métricas, esfuerzos y defectos que necesita esta operación.

### Salidas esperadas

`SprintReport` (por valor), formado por tipos que ya existen:

| Campo | Tipo | Origen |
|---|---|---|
| `Sprint` | `sprint.Sprint` | Datos del sprint: ID, proyecto, nombre, fechas de inicio y fin, Sprint Goal y estado (HU-10 y HU-13) |
| `Metrics` | `metrics.SprintMetrics` | M1, M2 y M4 a M10 del sprint (HU-35) |
| `Efforts` | `Efforts` (HU-38) | Esfuerzos del sprint y su total exacto de horas (HU-27) |
| `Defects` | `Defects` (HU-38) | Defectos detectados o resueltos en el sprint (RN-39.5) |
| `Members` | `Members` (HU-38) | Integrantes que registraron esfuerzo en el sprint (RN-39.6) |

- No se agrega ningún tipo para las partes: `Efforts`, `Defects` y `Members` son los de HU-38, sin campos nuevos.
- Las listas son nuevas, nunca `nil` (longitud 0 si no hay datos), y sus elementos son copias. Modificar el reporte no altera lo guardado.
- `sprint.Sprint` y `metrics.SprintMetrics` son valores sin punteros ni listas (`time.Time` se copia por valor).
- Los valores no se redondean en el dominio (TEC-03); el redondeo es de la API.

### HTTP

`GET /api/sprints/{id}/reporte` responde 200 con cinco partes, cada una con el formato de una respuesta que ya existe:

| Parte | Formato | Respuesta con el mismo formato |
|---|---|---|
| `sprint` | `sprintResponse` | `POST /api/proyectos/{id}/sprints` (HU-10) y el campo `sprint` de `GET /api/proyectos/{id}/metricas/sprints` (HU-37) |
| `metricas` | `sprintMetricsResponse` | `GET /api/sprints/{id}/metricas` (HU-35) |
| `esfuerzos` | `effortListResponse` | `GET /api/sprints/{id}/esfuerzos` (HU-27) |
| `defectos` | `defectListResponse` | `GET /api/sprints/{id}/defectos` (HU-31), con los defectos de RN-39.5 |
| `integrantes` | `memberListResponse` (HU-38) | `integrantes` de `GET /api/proyectos/{id}/reporte` |

No existe `GET /api/sprints/{id}`: la parte `sprint` usa `sprintResponse` con `toSprintResponse`, la conversión que ya usan HU-10 y HU-37. La parte `metricas` usa `toSprintMetricsResponse`, la conversión que ya comparten `GET /api/sprints/{id}/metricas` (HU-35) y `GET /api/proyectos/{id}/metricas/sprints` (HU-37), sin duplicarla (decisión 4). La parte `esfuerzos` usa `toEffortListResponse` con un `effort.Summary` armado a partir de `report.Efforts`, como HU-38.

Los ejemplos muestran la salida real de `encoding/json`: un número sin parte decimal se emite sin `.0`. Quien muestre los valores tiene que formatearlos con un decimal (TEC-03).

**Ejemplo 1.** El sprint del ejemplo de TEC-03, cerrado, con los cuatro ítems A a D (IDs 1 a 4), tres esfuerzos (12 h de Ana en A, 5 h de Luis en B y 3 h de Ana en C) y un defecto detectado y resuelto en el sprint. Las métricas son las del ejemplo de `GET /api/sprints/{id}/metricas` de la SDD de HU-35:

```json
{
  "sprint": {
    "id": 1,
    "proyecto_id": 1,
    "nombre": "Sprint 1",
    "fecha_inicio": "2026-10-05",
    "fecha_fin": "2026-10-16",
    "objetivo": "Registrar el esfuerzo del equipo",
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
  },
  "esfuerzos": {
    "esfuerzos": [
      {
        "id": 1,
        "sprint_id": 1,
        "item_id": 1,
        "integrante": "Ana",
        "fecha": "2026-10-06",
        "actividad": "Desarrollo",
        "horas": 12
      },
      {
        "id": 2,
        "sprint_id": 1,
        "item_id": 2,
        "integrante": "Luis",
        "fecha": "2026-10-07",
        "actividad": "Desarrollo",
        "horas": 5
      },
      {
        "id": 3,
        "sprint_id": 1,
        "item_id": 3,
        "integrante": "Ana",
        "fecha": "2026-10-08",
        "actividad": "Pruebas",
        "horas": 3
      }
    ],
    "total_horas": 20
  },
  "defectos": {
    "defectos": [
      {
        "id": 1,
        "descripcion": "El total de horas no suma los decimales",
        "severidad": "ALTA",
        "estado": "RESUELTO",
        "item_id": 1,
        "sprint_deteccion_id": 1,
        "sprint_resolucion_id": 1
      }
    ],
    "cantidad": 1
  },
  "integrantes": {
    "integrantes": ["Ana", "Luis"],
    "cantidad": 2
  }
}
```

**Ejemplo 2.** Un sprint recién creado, abierto, sin ítems, esfuerzos ni defectos:

```json
{
  "sprint": {
    "id": 2,
    "proyecto_id": 2,
    "nombre": "Sprint vacío",
    "fecha_inicio": "2026-10-19",
    "fecha_fin": "2026-10-30",
    "objetivo": "Sin trabajo todavía",
    "estado": "ABIERTO"
  },
  "metricas": {
    "sprint_id": 2,
    "sp_planificados": 0,
    "sp_completados": 0,
    "items_sin_estimar": 0,
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
  },
  "esfuerzos": {
    "esfuerzos": [],
    "total_horas": 0
  },
  "defectos": {
    "defectos": [],
    "cantidad": 0
  },
  "integrantes": {
    "integrantes": [],
    "cantidad": 0
  }
}
```

**Ejemplo 3 (parcial).** El defecto 1 se detectó en el sprint 1 y se resolvió en el sprint 2. En el reporte del sprint 2 aparece en `defectos`, porque M10 del sprint 2 lo cuenta (RN-39.5); `GET /api/sprints/2/defectos` no lo lista, porque solo muestra los detectados en el sprint:

```json
{
  "metricas": {
    "defectos_detectados": 0,
    "defectos_resueltos": 1
  },
  "defectos": {
    "defectos": [
      {
        "id": 1,
        "descripcion": "El total de horas no suma los decimales",
        "severidad": "ALTA",
        "estado": "RESUELTO",
        "item_id": 1,
        "sprint_deteccion_id": 1,
        "sprint_resolucion_id": 2
      }
    ],
    "cantidad": 1
  }
}
```

Campos propios del reporte (los de cada parte están documentados en la SDD de su historia):

| Campo | Tipo | Puede ser `null` |
|---|---|---|
| `sprint`, `metricas`, `esfuerzos`, `defectos`, `integrantes` | objeto | No: siempre presentes, en este orden |
| `esfuerzos.esfuerzos`, `defectos.defectos`, `integrantes.integrantes` | lista | No: `[]` si no hay datos |
| Campos de `metricas` que pueden no estar disponibles | número | Sí: `porcentaje_completadas`, `desviacion_horas` y `desviacion_porcentual`, con las reglas de RN-35.6 (el sprint no tiene velocidad) |

| Caso | Código | Cuerpo |
|---|---|---|
| Consulta exitosa, de un sprint abierto o cerrado, aunque alguna parte esté vacía | 200 | El JSON de los ejemplos |
| `{id}` que no es un número entero | 400 | `{"error": "el id debe ser un número entero"}` (`errInvalidID`) |
| Sprint inexistente (incluidos 0 y negativos) | 404 | `{"error": "no existe un sprint con ese id"}` (`sprint.ErrSprintNotFound`) |
| Falla de cualquier repositorio | 500 | `{"error": "error interno del servidor"}` (`errInternal`), sin detalle |

### Interfaz web

En la sección "Reportes" de HU-38 se agrega el botón "Generar reporte del sprint" (`id="generar-reporte-sprint"`), que usa el sprint actual (`sprintActual()`). El reporte se muestra en su propio contenedor, `id="datos-reporte-sprint"`, después del reporte del proyecto, con cinco bloques con ids propios, y los mensajes en `id="msg-reporte-sprint"`. El reporte del proyecto y sus ids no cambian.

| Bloque | Id | Contenido (ejemplo 1) | Sin datos |
|---|---|---|---|
| Sprint | `reporte-sprint` | "Sprint «Sprint 1» · Estado: CERRADO · Del 2026-10-05 al 2026-10-16. Objetivo: Registrar el esfuerzo del equipo" | — (un sprint siempre tiene esos datos) |
| Métricas | `reporte-sprint-metricas` | Story points planificados y completados e ítems sin estimar, historias completadas, horas y desviación, y defectos, con un decimal donde corresponde | "sin datos" en las tres métricas que pueden no estar disponibles |
| Esfuerzos | `reporte-sprint-esfuerzos` y `reporte-sprint-total-horas` | Tabla con las mismas columnas que el reporte del proyecto y "Total: 20.0 horas" | "Sin esfuerzos registrados." y "Total: 0.0 horas" |
| Defectos | `reporte-sprint-defectos` y `reporte-sprint-cantidad-defectos` | Tabla con las mismas columnas que el reporte del proyecto y "Cantidad: 1 defecto(s)" | "Sin defectos registrados." y "Cantidad: 0 defecto(s)" |
| Integrantes | `reporte-sprint-integrantes` | "Integrantes que registraron esfuerzo: Ana, Luis" | "Integrantes que registraron esfuerzo: ninguno" |

Sin sprint actual, el mismo mensaje de error que las demás secciones ("Elegí un sprint: …"). Las filas y los textos reutilizan las funciones que ya existen en `app.js` (`filasReporte`, `filaEsfuerzo`, `filaDefecto`, `historiasCompletadas`, `horasYDesviacion` y `defectos`) sin modificarlas.

## Reglas de negocio

1. **RN-39.1 Consulta única.** `Sprint` devuelve el reporte completo en una sola llamada, con las cinco partes. El handler no llama a ninguna otra operación.
2. **RN-39.2 Sin cálculos nuevos.** Cada parte sale de una operación existente: los datos del sprint de `sprint.Manager.Get` (HU-11), las métricas de `metrics.Manager.Sprint` (HU-35), los esfuerzos de `effort.Manager.ListBySprint` (HU-27) y los defectos de `defect.Manager.ListByProject` (HU-31). Lo único que arma `report` es la selección de los defectos del sprint (RN-39.5) y la lista de integrantes sin repetir (`membersOf` de HU-38).
3. **RN-39.3 Lecturas y orden.** `Sprint` llama a cada operación una sola vez, en este orden:

   | Orden | Operación | Para qué |
   |---|---|---|
   | 1 | `sprint.Manager.Get(sprintID)` | Datos del sprint y validación (`ErrSprintNotFound`) |
   | 2 | `metrics.Manager.Sprint(sprintID)` | M1, M2 y M4 a M10 |
   | 3 | `effort.Manager.ListBySprint(sprintID)` | Esfuerzos, total e integrantes |
   | 4 | `defect.Manager.ListByProject(sprint.ProjectID)` | Defectos del proyecto del sprint, para seleccionar los de RN-39.5 |

   Si una lectura falla, no se hacen las siguientes y se devuelve ese error, sin reporte parcial. Así un sprint inexistente devuelve `sprint.ErrSprintNotFound` sin leer nada más.
4. **RN-39.4 Esfuerzos del sprint.** Son los esfuerzos de `ListBySprint`, ordenados por ID (RN-27.7), y `TotalHours` es su `Summary.TotalHours`. Es la misma lectura que usa M5 del sprint, así que con los mismos datos `Efforts.TotalHours` es exactamente el mismo `float64` que `Metrics.Hours.Actual`. No aparecen esfuerzos de otros sprints.
5. **RN-39.5 Defectos del sprint.** Son los defectos del proyecto del sprint que cumplen al menos una condición: su sprint de detección es el sprint (los que cuenta M9), o están `RESUELTO` y su sprint de resolución es el sprint (los que cuenta M10). Cada defecto aparece una sola vez, aunque cumpla las dos, ordenado por ID (RN-31.2), y en cualquier estado. Así, con los mismos datos, los defectos de la lista con ese sprint de detección son exactamente M9, y los `RESUELTO` con ese sprint de resolución son exactamente M10 (decisión 3). Como HU-29 exige que el sprint de detección sea del proyecto de la historia (`ErrDetectionSprintFromAnotherProject`), los detectados en el sprint son los mismos que lista `GET /api/sprints/{id}/defectos`.
6. **RN-39.6 Integrantes.** Son los valores distintos del campo integrante de los esfuerzos del reporte, cada uno una sola vez, en el orden de su primer esfuerzo, con `membersOf` de HU-38 (RN-38.6). Es el criterio del tema 5 del Planning del Sprint 4, aplicado al sprint.
7. **RN-39.7 Sprints abiertos y cerrados.** El reporte se genera para un sprint en cualquier estado. No se valida el estado ni se cambia nada del sprint.
8. **RN-39.8 Sprint sin datos.** Un sprint existente sin ítems, esfuerzos ni defectos no es un error: el reporte tiene sus datos, las métricas con los valores de RN-35.6 (sumas en 0 y las tres no disponibles en `null`), y las listas de esfuerzos, defectos e integrantes vacías.
9. **RN-39.9 Coincidencia con las consultas existentes.** Con los mismos datos: `metricas` es igual a `GET /api/sprints/{id}/metricas`, `esfuerzos` a `GET /api/sprints/{id}/esfuerzos`, `sprint` a `toSprintResponse` del sprint, y en `defectos` los detectados en el sprint son los de `GET /api/sprints/{id}/defectos`, más los resueltos en el sprint detectados en otro.
10. **RN-39.10 Redondeo.** Sin cambios: el dominio no redondea; la API redondea las métricas como HU-35 y el total de horas como HU-27, a un decimal.
11. **RN-39.11 Datos actuales y solo lectura.** El reporte se arma en el momento de la consulta, sin caché. Generarlo no modifica sprints, ítems, esfuerzos ni defectos.
12. **RN-39.12 Copias.** `SprintReport` se devuelve por valor y sus listas son nuevas, con copias de cada elemento.
13. **RN-39.13 Sin errores nuevos.** No se agregan errores de dominio ni códigos HTTP.

## Decisión sobre entidades referenciadas

- **Sprint:** sí se valida su existencia, con `sprint.Manager.Get`, antes de cualquier otra lectura; esa misma lectura da los datos del sprint. Sin la validación, un ID mal escrito mostraría un reporte vacío y ocultaría el error. Incluye 0 y negativos (`ErrSprintNotFound`), como CA-39.3.
- **Proyecto del sprint:** no se valida aparte. Un sprint existente siempre pertenece a un proyecto existente (HU-10 lo valida al crearlo y el sistema no elimina proyectos), y `defect.Manager.ListByProject` lo vuelve a validar por su cuenta.
- **Ítems, esfuerzos y defectos:** `report` no los valida; los leen las operaciones reutilizadas, con sus propias validaciones internas.
- **Integrantes:** no son una entidad del sistema (E4 está fuera de la entrega); son textos de los esfuerzos y no se validan.

## Restricciones

- **Organización del código.**

  | Archivo | Cambio |
  |---|---|
  | `internal/report/sprint.go` | Nuevo: `SprintReport`, `Sprint` y la selección de defectos de RN-39.5 |
  | `internal/report/doc.go` | Solo se agrega HU-39 a la descripción del paquete |
  | `internal/report/manager.go` y `internal/report/project.go` | No cambian: se reutilizan `NewManager`, `Efforts`, `Defects`, `Members` y `membersOf` |
  | `internal/httpapi/sprint_reports.go` | Nuevo: `sprintReportResponse`, `toSprintReportResponse`, `toMemberListResponse` y el handler `sprintReport`; reutiliza `toSprintResponse`, `toSprintMetricsResponse`, `toEffortListResponse`, `toDefectListResponse` y `memberListResponse` sin cambiarlos |
  | `internal/httpapi/reports.go` y `metrics.go` | No cambian |
  | `internal/httpapi/server.go` | Solo la ruta `GET /api/sprints/{id}/reporte` (el `report.Manager` ya se crea en `NewHandler`) |
  | `internal/web/index.html` | Solo el botón y el contenedor nuevos, dentro de la sección "Reportes" |
  | `internal/web/app.js` | Solo el manejo del botón nuevo y sus textos, en un bloque propio después del de HU-38 |
  | `internal/web/styles.css` | Solo si hace falta, con selectores bajo `#datos-reporte-sprint` |

- **Cambios en paralelo.** HU-40 va a usar `internal/report` y probablemente `internal/httpapi/reports.go`. Por eso el handler y los tests de API de HU-39 van en archivos nuevos (`sprint_reports.go` y `sprint_reports_test.go`), y en `internal/report` solo se agrega `sprint.go` y una línea en `doc.go`. Al Paso 0 no hay ramas ni Pull Requests abiertos de HU-40.
- **Sin persistencia propia.** `report` no tiene repositorio ni estado. No hay tablas nuevas ni cambios de esquema.
- **Concurrencia.** Sin mutex propio, como HU-38: cada lectura queda protegida por el mutex del `Manager` que la hace. No se agrega un test de concurrencia.
- **Copias por valor.** Todo se devuelve por valor, con listas nuevas.
- **Tests.**
  - Unitarios en `internal/report/sprint_test.go`: cada parte igual a su operación de origen con los mismos datos; `Efforts.TotalHours` igual a `Metrics.Hours.Actual`; defectos detectados en el sprint, resueltos en el sprint y detectados en otro, y de otros sprints (no aparecen), con los conteos iguales a M9 y M10; un defecto detectado y resuelto en el sprint aparece una sola vez; integrantes sin repetir; sprint abierto y cerrado; sprint sin datos; listas no `nil`; copias; sprint inexistente; y el orden de los errores de RN-39.3 con repositorios de prueba que fallan, reutilizando los de `project_test.go` si sirven.
  - godog en `internal/report/sprint_features_test.go`, con el runner `TestFeatures_HU39ReporteSprint` sobre `features/HU-39-reporte-sprint.feature`, con el CA en el nombre de cada escenario.
  - API en `internal/httpapi/sprint_reports_test.go`: el JSON completo de los ejemplos 1 y 2, comparando el cuerpo crudo normalizado con `json.Compact` (como HU-35 y HU-38), y un caso por código (200, 400, 404 y 500). Se agrega una fila a `TestRutasConID_IDNoNumericoDevuelve400`.
  - Interfaz en `internal/httpapi/web_test.go`: el botón, el contenedor y los ids nuevos.
  - Los tests existentes, incluidos los de HU-38, tienen que seguir pasando sin cambios.
- **RED.** Firmas vacías de `SprintReport` y `Sprint`, para que el RED compile y falle por aserciones. `go vet ./...` antes del commit RED. La ruta, el handler y la interfaz van en el GREEN.
- **Profundidad proporcional (3 SP).** Un test por regla y por código; sin pruebas de mutación ni de concurrencia.
- **Librerías.** Solo la librería estándar; godog en los tests.

### Límites conocidos

- **Lecturas internas que se repiten.** El reporte llama una sola vez a cada operación (RN-39.3), pero reutiliza `metrics.Manager.Sprint` sin cambios, y esa operación ya lee por su cuenta el sprint, los ítems del proyecto, los esfuerzos del sprint y los defectos del proyecto. A nivel de repositorio, una consulta lee:

  | Fuente | Lecturas | Detalle |
  |---|---|---|
  | Sprint (por ID) | 4 | `report` (1), `metrics` (1) y la validación dentro de las dos llamadas a `effort.Manager.ListBySprint` (2) |
  | Proyecto (por ID) | 3 | `ListItems` de `metrics` (1) y `ListItems` dentro de las dos llamadas a `defect.Manager.ListByProject` (2) |
  | Ítems del proyecto | 3 | `metrics` (1) y dentro de las dos llamadas a `defect.Manager.ListByProject` (2) |
  | Esfuerzos del sprint | 2 | `metrics` (1) y `report` (1) |
  | Defectos | 2 | `metrics` (1) y `report` (1) |

  Evitarlo requiere cambiar `internal/metrics`, igual que en HU-38 (su decisión 5). Es aceptable para una consulta de solo lectura que se pide a mano.
- **Sin lista de ítems del sprint.** Las historias planificadas y completadas y sus estimaciones se informan con las métricas (M1, M2, M4 y M8, con `items_totales`, `items_terminados`, `items_sin_estimar` e `items_sin_horas_estimadas`), no ítem por ítem, como en HU-38 (decisión 9). El detalle de los ítems terminados está en `GET /api/sprints/{id}/completados` (HU-12).
- **Regla de M9 y M10 repetida.** La selección de RN-39.5 repite en `report` la condición de `sprintDefects` de `internal/metrics`, que no se exporta. Los tests de RN-39.5 comparan los conteos con M9 y M10, así que las dos no pueden divergir sin que falle un test.
- **Lecturas no atómicas.** Como en HU-38: un cambio simultáneo entre lecturas puede quedar en una parte y no en otra, y entonces `total_horas` o los defectos podrían no coincidir con las métricas.
- **Integrantes por texto.** Como en HU-38: "Ana" y "ana" aparecen dos veces, y un integrante sin esfuerzos en el sprint no aparece.
- **Valores no congelados.** Si se reestima o reabre un ítem de un sprint cerrado, cambian sus métricas en el reporte (sin cambios respecto de HU-32).

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Sprint del ejemplo de TEC-03, cerrado, con tres esfuerzos y un defecto | El JSON del ejemplo 1 |
| Sprint recién creado, abierto, sin datos | El JSON del ejemplo 2: listas vacías, cantidades y total en 0, tres métricas en `null` |
| Sprint abierto con datos | Mismo reporte que si estuviera cerrado, con `estado` `ABIERTO` (CA-39.2) |
| Sprint con ítems y sin esfuerzos | Métricas de los ítems; esfuerzos, total e integrantes vacíos; desviación en `null` |
| Defecto detectado y resuelto en el sprint | Aparece una sola vez; cuenta en M9 y en M10 |
| Defecto detectado en el sprint y resuelto en otro | Aparece; cuenta en M9 y no en M10 |
| Defecto detectado en otro sprint y resuelto en este | Aparece; cuenta en M10 y no en M9 (ejemplo 3) |
| Defecto `EN_CORRECCION` o `ABIERTO` detectado en otro sprint | No aparece |
| Esfuerzos y defectos de otros sprints del mismo proyecto o de otros proyectos | No se incluyen, y sus integrantes tampoco |
| Un integrante con varios esfuerzos en el sprint | Aparece una sola vez, en la posición de su primer esfuerzo |
| Horas con decimales (0.1, 0.2) | El total exacto en el dominio y redondeado a un decimal en la API |
| Generar el reporte dos veces sin cambios | El mismo resultado; no se modifica nada |
| Modificar las listas del reporte devuelto | Lo guardado no cambia |
| Sprint inexistente, 0 o negativo | `sprint.ErrSprintNotFound`, sin otras lecturas; 404 |
| Falla la lectura del sprint, las métricas, los esfuerzos o los defectos | Ese error, sin las lecturas siguientes ni reporte parcial; 500 |
| HTTP: `{id}` no numérico | 400 |
| Interfaz sin sprint actual | Mensaje "Elegí un sprint: …" en `msg-reporte-sprint`, sin llamar a la API |

## Condiciones de error

| Condición | Error | HTTP |
|---|---|---|
| `{id}` que no es un número entero | `errInvalidID` | 400 |
| No existe el sprint (incluidos 0 y negativos) | `sprint.ErrSprintNotFound` | 404 |
| Falla el repositorio de sprints, proyectos, ítems, esfuerzos o defectos | El error del repositorio, en el orden de RN-39.3 | 500, sin detalle |
| Una parte está vacía o una métrica no está disponible | No es un error (RN-39.8) | 200 |

## Criterios de aceptación

- **CA-39.1** El reporte de un sprint incluye sus datos (nombre, fechas, Sprint Goal y estado), sus métricas (story points planificados y completados, historias planificadas y completadas, horas estimadas y reales, desviación y defectos detectados y resueltos), los esfuerzos registrados con su total de horas, los defectos detectados o resueltos en el sprint y los integrantes que registraron esfuerzo en el sprint.
- **CA-39.2** El reporte se genera para sprints abiertos y cerrados.
- **CA-39.3** Generar el reporte de un sprint inexistente devuelve error.
- **CA-39.4** Un sprint sin datos genera un reporte con las secciones vacías y las métricas que no se pueden calcular indicadas como no disponibles, no un error.
- **CA-39.5** Los datos del reporte coinciden con las consultas de las métricas y de los esfuerzos del sprint con los mismos datos; el total de horas es igual a las horas reales; y los defectos de la lista coinciden con los defectos detectados y resueltos de las métricas.
- **CA-39.6** Cada integrante que registró esfuerzo en el sprint aparece una sola vez, y no aparecen esfuerzos, defectos ni integrantes de otros sprints, salvo los defectos resueltos en el sprint.
- **CA-39.7** Generar el reporte no modifica nada.
- **CA-39.8** El reporte se genera desde la API y desde la sección "Reportes" de la interfaz web, con el sprint actual.

## Comparación con el acta de Planning

**Coincide con el acta del Sprint 4.** El acta asigna HU-39 a García Santiago con **3 SP**, con dependencias HU-35, HU-27 y HU-31, y la describe como "Mismo reporte que HU-38, acotado a un sprint; reutiliza su estructura". La SDD hace eso: una operación nueva en `internal/report`, los mismos tipos de HU-38 y una respuesta con las mismas partes, acotadas al sprint. Las tres dependencias y HU-38 están en `main` (PR #126, #141, #145 y #148).

Precisiones registradas:

1. **Integrantes (tema 5).** El tema 5 se decidió para HU-38; esta SDD aplica el mismo criterio al sprint: los integrantes que registraron esfuerzo en él, y E4 sigue fuera de la entrega.
2. **Recorte (acuerdo previo).** El acta prevé "un reporte de sprint más simple en HU-39" si el Sprint se atrasa. Esta SDD ya es la versión simple: sin lista de ítems, sin tipos ni cálculos nuevos y sin partes que HU-38 no tenga. No se recorta nada más.
3. **Defectos del sprint.** El acta no los define; la SDD incluye los detectados y los resueltos en el sprint, para que coincidan con M9 y M10 (RN-39.5 y decisión 3).

## Comparación con los criterios del Issue #57

Los criterios del Issue #57 están en borrador. Con la aprobación de esta SDD se cierran con CA-39.1 a CA-39.8, y se actualiza el Issue con autorización.

| CA original del Issue #57 | En esta SDD | Ajuste |
|---|---|---|
| CA-39.1 Sprint Goal, historias planificadas y completadas, estimaciones, esfuerzo registrado, métricas y defectos del sprint | CA-39.1, CA-39.5 y CA-39.6 | Se agregan los datos del sprint y los integrantes. Historias planificadas y completadas y estimaciones: con las métricas (M1, M2, M4 y M8 y sus cantidades), sin lista de ítems, como HU-38. Defectos: los detectados o resueltos en el sprint |
| CA-39.2 Se puede generar para sprints abiertos y cerrados | CA-39.2 | Sin cambios |
| CA-39.3 Un sprint inexistente devuelve error | CA-39.3 | Sin cambios |
| — | CA-39.4 a CA-39.8 (nuevos) | Sprint sin datos, coincidencia con las consultas existentes, integrantes y aislamiento, solo lectura, API e interfaz, como en HU-38 |

## Decisiones tomadas en la redacción

1. **Operación nueva en `internal/report`**, no un paquete nuevo: el acta y la SDD de HU-38 prevén que HU-39 reutilice su estructura. `NewManager` no cambia. *Alternativa descartada:* un paquete o un `Manager` propio, que duplicaría el armado.
2. **Mismos tipos de HU-38** (`Efforts`, `Defects` y `Members`) y `metrics.SprintMetrics` de HU-35, sin tipos nuevos para las partes (decisión 4 de HU-38). `SprintReport` es el único tipo nuevo.
3. **Defectos detectados o resueltos en el sprint.** M9 cuenta los defectos detectados en el sprint y M10 los resueltos en él, que pueden haberse detectado en otro sprint. Con solo los detectados (`defect.Manager.ListBySprint` de HU-31), un sprint que resuelve un defecto anterior mostraría "1 resuelto" en las métricas y ningún defecto en la lista. Con la unión, cada defecto que cuentan M9 o M10 aparece en el reporte, y los conteos se verifican contra las métricas. La selección se hace sobre `defect.Manager.ListByProject` del proyecto del sprint, sin cambiar `internal/defect`. *Alternativas descartadas:* `ListBySprint` solo, por la diferencia con M10 que se explicó; y dos listas separadas (detectados y resueltos), que repetiría un defecto detectado y resuelto en el mismo sprint y cambiaría el formato de la parte `defectos`.
4. **Reutilizar `toSprintMetricsResponse`.** La conversión de las métricas del sprint ya es una función que comparten `GET /api/sprints/{id}/metricas` (HU-35) y `GET /api/proyectos/{id}/metricas/sprints` (HU-37), así que el reporte la usa directamente y la parte `metricas` no puede divergir del endpoint. No repite el límite conocido de HU-38 (`toReportMetricsResponse`), que existe porque el proyecto no tenía esa función. Unificar la del proyecto queda fuera de esta historia (no incluye cambios en HU-38).
5. **`toSprintResponse` para la parte `sprint`.** No existe `GET /api/sprints/{id}`; se usa el formato de HU-10, que también usa HU-37 en `/metricas/sprints`. *Alternativa descartada:* agregar `GET /api/sprints/{id}`, que es un endpoint que el Issue no pide.
6. **Sin lista de ítems del sprint**, como la decisión 9 de HU-38: las historias planificadas y completadas y sus estimaciones salen de las métricas, y el reporte de sprint queda tan simple como prevé el acta. *Alternativa descartada:* agregar los ítems del sprint como sexta parte (con `backlog.Manager.ListItems` filtrado por sprint), que es una parte y un formato que el reporte del proyecto no tiene.
7. **`toMemberListResponse` nueva** en `sprint_reports.go`, para la parte `integrantes` del sprint. `toProjectReportResponse` de HU-38 arma la misma respuesta en línea; usarla también ahí es un refactor que se puede proponer en el Paso 6, cubierto por los tests de HU-38, y no se hace en el GREEN.
8. **Archivos nuevos en `internal/httpapi`** para el handler y los tests de API, porque HU-40 probablemente modifique `reports.go`.
9. **Contenedor propio en la interfaz**, dentro de "Reportes", para que los dos reportes se puedan ver a la vez sin pisarse, con ids `reporte-sprint-*`, y reutilizando las funciones de `app.js` sin modificarlas.
10. **Story Points.** 3 SP, según el Planning del Sprint 4; no se divide.
11. **Sin librerías nuevas.** Solo la librería estándar de Go; godog se mantiene en los tests. El PDF es de HU-40.

HU-39: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
