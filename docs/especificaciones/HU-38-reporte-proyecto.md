# HU-38 · Generar el reporte del proyecto

> Issue del repositorio: #56

Aprobada por: Peñalbé Hernán (Agile Enabler), 29/10/2026, con dos ajustes (conversión duplicada de las métricas registrada como límite conocido y reutilización de toEffortListResponse precisada), que se aplicaron.

## Historia de usuario

Como Agile Enabler, quiero generar un reporte del proyecto, para presentar su estado a los profesores.

## Objetivo

Reunir en una sola consulta el reporte de un proyecto: sus datos, sus métricas M1 a M10 (HU-35), sus esfuerzos (HU-27), sus defectos (HU-31) y los integrantes que registraron esfuerzo en el proyecto (HU-26 y HU-27). Se implementa en un paquete nuevo `internal/report`, con el endpoint `GET /api/proyectos/{id}/reporte` y una sección nueva "Reportes" en la interfaz web. No hay cálculos nuevos: cada parte sale de una operación que ya existe. El reporte se arma con partes en tipos separados (métricas, esfuerzos, defectos e integrantes) y un contrato HTTP documentado con ejemplos, para que lo reutilicen HU-39 (reporte de un sprint) y HU-40 (exportar a PDF).

## Alcance

- **Incluye:**
  - el paquete `internal/report`, con la operación `Project(projectID)` y los tipos `ProjectReport`, `Efforts`, `Defects` y `Members`;
  - el endpoint `GET /api/proyectos/{id}/reporte`, con su respuesta formada por las respuestas que ya existen de proyectos, métricas, esfuerzos y defectos;
  - la sección "Reportes" de la interfaz web, con el botón "Generar reporte";
  - el contrato del reporte, con ejemplos JSON completos, para HU-39 y HU-40.
- **No incluye:**
  - integrantes del equipo como entidad (E4, HU-15 a HU-17, fuera de la entrega; tema 5 del Planning del Sprint 4): el reporte lista solo los integrantes que registraron esfuerzo en el proyecto;
  - la lista de ítems del backlog: las historias planificadas y completadas y sus estimaciones se informan con las métricas M1, M2, M4 y M8 y sus cantidades (ver "Decisiones tomadas en la redacción");
  - cálculos o métricas nuevas, ni totales por integrante o por sprint;
  - el reporte de un sprint (HU-39) y la exportación a PDF (HU-40), ni librerías externas;
  - cambios en `internal/project`, `internal/backlog`, `internal/sprint`, `internal/metrics`, `internal/effort`, `internal/defect` o `internal/dashboard`, ni en sus repositorios o tests;
  - cambios en los endpoints existentes, en la sección "Métricas" ni en "Ver resumen" (HU-36);
  - cambios de esquema, tablas nuevas, filtros (por estado de defecto, integrante o fechas) ni transacciones entre fuentes.

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `Project(projectID int) (ProjectReport, error)` | ID de un proyecto | Reporte completo del proyecto |

Está en el `Manager` de `internal/report`, que se crea con:

```go
NewManager(projects *project.Manager, sprints *sprint.Manager, metrics *metrics.Manager, efforts *effort.Manager, defects *defect.Manager) *Manager
```

### Salidas esperadas

`ProjectReport` (por valor), formado por tipos que ya existen y tres tipos nuevos, uno por parte:

| Campo | Tipo | Origen |
|---|---|---|
| `Project` | `project.Project` | Datos del proyecto: ID, nombre, descripción, estado y fecha de inicio (HU-01, HU-03 y HU-18) |
| `Metrics` | `metrics.ProjectMetrics` | M1 a M10 del proyecto (HU-35) |
| `Efforts` | `Efforts` | Esfuerzos del proyecto y su total exacto de horas (HU-27) |
| `Defects` | `Defects` | Defectos del proyecto (HU-31) |
| `Members` | `Members` | Integrantes que registraron esfuerzo en el proyecto (HU-26 y HU-27) |

Tipos nuevos de `internal/report`:

| Tipo | Campos | Descripción |
|---|---|---|
| `Efforts` | `Efforts []effort.Effort`, `TotalHours float64` | Todos los esfuerzos de los sprints del proyecto y la suma exacta de sus horas, sin redondear |
| `Defects` | `Defects []defect.Defect` | Todos los defectos del proyecto, en cualquier estado |
| `Members` | `Names []string` | Nombres de los integrantes que registraron esfuerzo, cada uno una sola vez |

- Las listas son nuevas, nunca `nil` (longitud 0 si no hay datos), y sus elementos son copias: `effort.Effort`, `defect.Defect` y `string` no tienen punteros, listas ni mapas. Modificar el reporte no altera lo guardado.
- `metrics.ProjectMetrics` y `project.Project` son valores sin punteros ni listas.
- Los valores no se redondean en el dominio (TEC-03); el redondeo es de la API.
- HU-39 puede armar el reporte de un sprint con las mismas partes (`Efforts`, `Defects` y `Members`) y `metrics.SprintMetrics`; HU-40 puede tomar `ProjectReport` o la respuesta HTTP. Ninguna de las dos se implementa en esta historia.

### HTTP

`GET /api/proyectos/{id}/reporte` responde 200 con cinco partes, cada una con el formato de un endpoint que ya existe:

| Parte | Formato | Endpoint con el mismo formato |
|---|---|---|
| `proyecto` | `projectResponse` | `GET /api/proyectos/{id}` (HU-03 y HU-18) |
| `metricas` | `projectMetricsResponse` | `GET /api/proyectos/{id}/metricas` (HU-35) |
| `esfuerzos` | `effortListResponse` | `GET /api/sprints/{id}/esfuerzos` (HU-27) |
| `defectos` | `defectListResponse` | `GET /api/proyectos/{id}/defectos` sin filtro (HU-31) |
| `integrantes` | `memberListResponse` (nuevo): `integrantes` (lista de textos) y `cantidad` | — |

`toEffortListResponse` recibe un `effort.Summary`, así que el handler arma `effort.Summary{Efforts: <esfuerzos del reporte>, TotalHours: <total del reporte>}` a partir de `report.Efforts` y se lo pasa. Así la parte `esfuerzos` tiene exactamente el formato de `GET /api/sprints/{id}/esfuerzos`, incluido el redondeo del total a un decimal.

Los ejemplos muestran la salida real de `encoding/json`: un número sin parte decimal se emite sin `.0` (por ejemplo `20` en lugar de `20.0`). Quien muestre los valores tiene que formatearlos con un decimal (TEC-03).

**Ejemplo 1.** El proyecto del ejemplo de TEC-03: un sprint cerrado con los cuatro ítems A a D (IDs 1 a 4), tres esfuerzos (12 h de Ana en A, 5 h de Luis en B y 3 h de Ana en C) y un defecto detectado y resuelto en ese sprint:

```json
{
  "proyecto": {
    "id": 1,
    "nombre": "Software Metrics",
    "descripcion": "Estimación y seguimiento de proyectos",
    "estado": "EN_CURSO",
    "fecha_inicio": "2026-10-05"
  },
  "metricas": {
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

**Ejemplo 2.** Un proyecto recién creado, sin fecha de inicio, ítems, sprints, esfuerzos ni defectos:

```json
{
  "proyecto": {
    "id": 2,
    "nombre": "Proyecto sin datos",
    "descripcion": "Recién creado",
    "estado": "PLANIFICADO",
    "fecha_inicio": ""
  },
  "metricas": {
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

Campos propios del reporte (los de cada parte están documentados en la SDD de su historia):

| Campo | Tipo | Puede ser `null` |
|---|---|---|
| `proyecto`, `metricas`, `esfuerzos`, `defectos`, `integrantes` | objeto | No: siempre presentes, en este orden |
| `esfuerzos.esfuerzos`, `defectos.defectos`, `integrantes.integrantes` | lista | No: `[]` si no hay datos |
| `integrantes.cantidad` | entero | No |
| Campos de `metricas` que pueden no estar disponibles | número | Sí: `velocidad`, `porcentaje_completadas`, `desviacion_horas` y `desviacion_porcentual`, con las reglas de RN-35.6 |

| Caso | Código | Cuerpo |
|---|---|---|
| Consulta exitosa, aunque alguna parte esté vacía o alguna métrica no esté disponible | 200 | El JSON de los ejemplos 1 y 2 |
| `{id}` que no es un número entero | 400 | `{"error": "el id debe ser un número entero"}` (`errInvalidID`) |
| Proyecto inexistente (incluidos 0 y negativos) | 404 | `{"error": "no existe un proyecto con ese id"}` (`project.ErrProjectNotFound`) |
| Falla de cualquier repositorio | 500 | `{"error": "error interno del servidor"}` (`errInternal`), sin detalle |

### Interfaz web

Sección nueva "Reportes" (`id="reportes"`), después de "Dashboard", con el botón "Generar reporte" (`id="generar-reporte"`), que usa el proyecto actual. El reporte se muestra en `id="datos-reporte"`, con cinco bloques con ids propios, y los mensajes en `id="msg-reporte"`:

| Bloque | Id | Contenido (ejemplo 1) | Sin datos |
|---|---|---|---|
| Proyecto | `reporte-proyecto` | "Proyecto «Software Metrics» · Estado: EN_CURSO · Inicio: 2026-10-05" y la descripción | "Inicio: sin registrar" si no hay fecha de inicio |
| Métricas | `reporte-metricas` | Story points, velocidad, historias completadas, horas y desviación, y defectos, con un decimal donde corresponde | "sin datos" en las cuatro métricas que pueden no estar disponibles |
| Esfuerzos | `reporte-esfuerzos` | Tabla con ID, sprint, ítem, integrante, fecha, actividad y horas, y "Total: 20.0 horas" | "Sin esfuerzos registrados." y "Total: 0.0 horas" |
| Defectos | `reporte-defectos` | Tabla con ID, descripción, severidad, estado, ítem y sprints de detección y resolución, y "Cantidad: 1 defecto(s)" | "Sin defectos registrados." y "Cantidad: 0 defecto(s)" |
| Integrantes | `reporte-integrantes` | "Integrantes que registraron esfuerzo: Ana, Luis" | "Integrantes que registraron esfuerzo: ninguno" |

Sin proyecto actual, el mismo mensaje de error que las demás secciones. Las filas y los textos reutilizan las funciones que ya existen en `app.js` (`filaEsfuerzo`, `filaDefecto`, `numero1`, `historiasCompletadas`, `horasYDesviacion` y `defectos`) sin modificarlas.

## Reglas de negocio

1. **RN-38.1 Consulta única.** `Project` devuelve el reporte completo en una sola llamada, con las cinco partes. El handler no llama a ninguna otra operación.
2. **RN-38.2 Sin cálculos nuevos.** Cada parte sale de una operación existente: los datos del proyecto de `project.Manager.Get` (HU-03), las métricas de `metrics.Manager.Project` (HU-35), los esfuerzos de `effort.Manager.ListBySprint` (HU-27) y los defectos de `defect.Manager.ListByProject` (HU-31). Lo único que arma `report` es la unión de los esfuerzos de cada sprint, la suma de sus totales (como RN-35.5) y la lista de integrantes sin repetir.
3. **RN-38.3 Lecturas y orden.** `Project` llama a cada operación una sola vez (los esfuerzos, una vez por sprint), en este orden:

   | Orden | Operación | Para qué |
   |---|---|---|
   | 1 | `project.Manager.Get(projectID)` | Datos del proyecto y validación (`ErrProjectNotFound`) |
   | 2 | `metrics.Manager.Project(projectID)` | M1 a M10 |
   | 3 | `sprint.Manager.ListByProject(projectID)` | Sprints del proyecto, en orden de ID, para leer sus esfuerzos |
   | 4 | `effort.Manager.ListBySprint(sprintID)`, una vez por sprint en orden de ID | Esfuerzos e integrantes |
   | 5 | `defect.Manager.ListByProject(projectID)` | Defectos |

   Si una lectura falla, no se hacen las siguientes y se devuelve ese error, sin reporte parcial. Así un proyecto inexistente devuelve `project.ErrProjectNotFound` sin leer nada más. Las lecturas internas de las operaciones reutilizadas no cambian (ver "Límites conocidos").
4. **RN-38.4 Esfuerzos del proyecto.** Son los esfuerzos de todos los sprints del proyecto, abiertos y cerrados, ordenados por ID de sprint y, dentro de cada sprint, por ID de esfuerzo (RN-27.7). `TotalHours` es la suma de `Summary.TotalHours` de cada sprint en orden de ID de sprint, el mismo cálculo que M5 en HU-35 (RN-35.5), así que con los mismos datos es exactamente el mismo `float64` que `Metrics.Hours.Actual`. No aparecen esfuerzos de otros proyectos (RN-27.5).
5. **RN-38.5 Defectos del proyecto.** Son todos los defectos cuya historia pertenece al proyecto, en cualquier estado, ordenados por ID (RN-31.2 y RN-31.8). El reporte no filtra por estado.
6. **RN-38.6 Integrantes.** Son los valores distintos del campo integrante (`Effort.Member`) de los esfuerzos del reporte, cada uno una sola vez, en el orden en que aparecen por primera vez en la lista de esfuerzos de RN-38.4. Se comparan tal como están guardados: HU-26 ya quita los espacios de los extremos al registrar, y no se normalizan mayúsculas ni tildes. Es el criterio del tema 5 del Planning del Sprint 4: E4 (HU-15 a HU-17) no está implementada, así que no hay otra fuente de integrantes.
7. **RN-38.7 Proyecto sin datos.** Un proyecto existente sin ítems, sprints, esfuerzos ni defectos no es un error: el reporte tiene sus datos, las métricas con los valores de RN-35.6 (sumas en 0 y las cuatro no disponibles en `null`), y las listas de esfuerzos, defectos e integrantes vacías (`[]`, total 0 y cantidades 0).
8. **RN-38.8 Coincidencia con las consultas existentes.** Con los mismos datos, cada parte es igual a la respuesta del endpoint que tiene su formato: `proyecto` a `GET /api/proyectos/{id}`, `metricas` a `GET /api/proyectos/{id}/metricas`, `defectos` a `GET /api/proyectos/{id}/defectos`, y los esfuerzos de cada sprint a `GET /api/sprints/{id}/esfuerzos`.
9. **RN-38.9 Redondeo.** Sin cambios: el dominio no redondea; la API redondea las métricas como HU-35 y el total de horas como HU-27 (RN-27.8), a un decimal. Las horas de cada esfuerzo se devuelven como están guardadas, igual que en HU-27.
10. **RN-38.10 Datos actuales y solo lectura.** El reporte se arma en el momento de la consulta con los datos guardados, sin caché. Generarlo no modifica proyectos, ítems, sprints, esfuerzos ni defectos.
11. **RN-38.11 Copias.** `ProjectReport` se devuelve por valor y sus listas son nuevas, con copias de cada elemento (RN-27.9 y RN-31.10).
12. **RN-38.12 Sin errores nuevos.** No se agregan errores de dominio ni códigos HTTP.

## Decisión sobre entidades referenciadas

- **Proyecto:** sí se valida su existencia, con `project.Manager.Get`, antes de cualquier otra lectura. Esa misma lectura da los datos del proyecto. Sin la validación, un ID mal escrito mostraría un reporte vacío y ocultaría el error, como en HU-36. Incluye 0 y negativos (`ErrProjectNotFound`).
- **Sprints del proyecto:** no se validan uno por uno; se leen con `sprint.Manager.ListByProject` después de validar el proyecto. `effort.Manager.ListBySprint` vuelve a validar cada sprint por su cuenta (ver "Límites conocidos").
- **Ítems, esfuerzos y defectos:** `report` no los valida ni agrega validaciones; los leen las operaciones reutilizadas, con sus propias validaciones internas.
- **Integrantes:** no son una entidad del sistema (E4 está fuera de la entrega); son textos de los esfuerzos y no se validan.

## Restricciones

- **Organización del código.**

  | Archivo | Cambio |
  |---|---|
  | `internal/report/doc.go` | Nuevo: descripción del paquete (E11) |
  | `internal/report/manager.go` | Nuevo: `Manager` y `NewManager` |
  | `internal/report/project.go` | Nuevo: `ProjectReport`, `Efforts`, `Defects`, `Members` y `Project` |
  | `internal/httpapi/reports.go` | Nuevo: `projectReportResponse`, `memberListResponse`, la conversión de `metrics.ProjectMetrics` a `projectMetricsResponse` y el handler; reutiliza `toProjectResponse`, `toCompletionResponse`, `toHoursResponse`, `toDefectsResponse`, `optionalOneDecimal`, `toEffortListResponse` (con un `effort.Summary` armado a partir de `report.Efforts`) y `toDefectListResponse` sin cambiarlos |
  | `internal/httpapi/metrics.go` | No cambia (HU-37 lo modifica en paralelo) |
  | `internal/httpapi/server.go` | Solo lo nuevo: el campo `reports` de `server`, la creación de `report.Manager` en `NewHandler` (sin cambiar su firma) y la ruta `GET /api/proyectos/{id}/reporte` |
  | `internal/web/index.html` | Solo la sección nueva "Reportes" |
  | `internal/web/app.js` | Solo el manejo del botón "Generar reporte" y las funciones nuevas del reporte |
  | `internal/web/styles.css` | Solo si la sección necesita estilos propios, con selectores bajo `#reportes` |

  No cambian `cmd/server`, los endpoints existentes, la sección "Métricas", "Ver resumen" (HU-36) ni los paquetes de dominio existentes; en particular, ningún archivo de `internal/metrics` (decisión 5). La conversión de `metrics.ProjectMetrics` a `projectMetricsResponse` para la parte `metricas` se escribe en `internal/httpapi/reports.go`, reutilizando `toCompletionResponse`, `toHoursResponse`, `toDefectsResponse` y `optionalOneDecimal`, sin modificar `internal/httpapi/metrics.go`, que HU-37 modifica en paralelo. El test de RN-38.8 verifica que la parte `metricas` sea igual a la respuesta de `GET /api/proyectos/{id}/metricas`, así que las dos conversiones no pueden divergir.
- **Cambios en paralelo.** HU-37 modifica en paralelo `index.html`, `app.js`, `styles.css`, `server.go` e `internal/httpapi/metrics.go`. HU-38 no toca `metrics.go`, y sus cambios en los demás archivos se limitan a lo nuevo, en bloques propios, para reducir los conflictos al traer `main`. HU-37 también agrega una fila a `TestRutasConID_IDNoNumericoDevuelve400`, así que ese punto puede dar un conflicto trivial al traer `main`, que se resuelve conservando las dos filas.
- **Sin persistencia propia.** `report` no tiene repositorio ni estado. No hay tablas nuevas ni cambios de esquema.
- **Concurrencia.** `report` no tiene mutex: cada lectura queda protegida por el mutex del `Manager` que la hace, como en `metrics` y `dashboard`. No se agrega un test de concurrencia.
- **Copias por valor.** Todo se devuelve por valor, con listas nuevas.
- **Tests.**
  - Unitarios en `internal/report/project_test.go`: cada parte igual a su operación de origen con los mismos datos; `Efforts.TotalHours` igual a `Metrics.Hours.Actual`; esfuerzos de varios sprints en orden; integrantes sin repetir y en orden de primera aparición; aislamiento entre proyectos; proyecto sin datos; listas no `nil`; copias; proyecto inexistente; y el orden de los errores de RN-38.3 con repositorios de prueba que fallan.
  - godog en `internal/report/project_features_test.go`, con el runner `TestFeatures_HU38ReporteProyecto` sobre `features/HU-38-reporte-proyecto.feature`, con el CA en el nombre de cada escenario.
  - API en `internal/httpapi/reports_test.go`: el JSON completo de los dos ejemplos del contrato, comparando el cuerpo crudo con el esperado, ambos normalizados con `json.Compact` (como HU-35), y un caso por código (200, 400, 404 y 500). Se agrega una fila a `TestRutasConID_IDNoNumericoDevuelve400`.
  - Interfaz en `internal/httpapi/web_test.go`: la sección, el botón y los ids nuevos.
  - Los tests existentes de los demás endpoints tienen que seguir pasando sin cambios.
- **RED.** Firmas vacías de `Manager`, `NewManager`, `Project` y los cuatro tipos, para que el RED compile y falle por aserciones. `go vet ./...` antes del commit RED. La ruta, el handler y la interfaz van en el GREEN.
- **Profundidad proporcional (5 SP).** Un test por regla y por código; sin pruebas de mutación ni de concurrencia.
- **Librerías.** Solo la librería estándar; godog en los tests.

### Límites conocidos

- **Lecturas internas que se repiten.** El reporte llama una sola vez a cada operación (RN-38.3), pero reutiliza `metrics.Manager.Project` sin cambios (decisión 5), y esa operación ya lee por su cuenta la lista de sprints, los esfuerzos de cada sprint y los defectos. Además, las lecturas internas que dejaron así HU-33 y HU-47 siguen como están: `effort.Manager.ListBySprint` vuelve a leer su sprint con `sprint.Manager.Get`, `defect.Manager.ListByProject` lee el backlog con `ListItems`, y `ListItems` vuelve a leer el proyecto (RN-06.4). Para un proyecto con N sprints, a nivel de repositorio una consulta lee:

  | Fuente | Lecturas | Detalle |
  |---|---|---|
  | Proyecto (por ID) | 4 | `report` (1), `ListItems` de `metrics` (1) y `ListItems` dentro de cada una de las dos llamadas a `defect.Manager.ListByProject` (2) |
  | Ítems del proyecto | 3 | `metrics` (1) y dentro de las dos llamadas a `defect.Manager.ListByProject` (2) |
  | Sprints del proyecto | 2 | `metrics` (1) y `report` (1) |
  | Sprint (por ID) | 2N | Validación dentro de cada `effort.Manager.ListBySprint` |
  | Esfuerzos de un sprint | 2N | `metrics` (N) y `report` (N) |
  | Defectos | 2 | `metrics` (1) y `report` (1) |

  Evitarlo requiere que `metrics` calcule M1 a M10 a partir de datos ya leídos, lo que cambia `internal/metrics`. Esta historia no modifica `internal/metrics`: la consulta de HU-35 se reutiliza tal como está, y en paralelo otra historia modifica `internal/metrics/summary.go`. Es aceptable para una consulta de solo lectura que se pide a mano.
- **Sin lista de ítems del backlog.** Las historias planificadas y completadas y sus estimaciones se informan como cantidades y sumas de las métricas (M1, M2, M4 y M8, con `items_totales`, `items_terminados`, `items_sin_estimar` e `items_sin_horas_estimadas`), no ítem por ítem (decisión 9). Quien necesite el detalle de cada ítem lo consulta en el Product Backlog (HU-06).
- **Conversión de las métricas duplicada.** La conversión de `metrics.ProjectMetrics` a `projectMetricsResponse` se escribe en `internal/httpapi/reports.go` y duplica la que hace hoy el handler `projectMetrics` de HU-35 (en `main` no existe una función `toProjectMetricsResponse`). Se unifica en una sola función cuando HU-37 y HU-38 estén en `main`.
- **Lecturas no atómicas.** Las lecturas son separadas: un cambio simultáneo (por ejemplo, registrar un esfuerzo o cambiar el estado de un defecto entre la lectura de las métricas y la de los esfuerzos) puede quedar en una parte y no en otra, y entonces `total_horas` o la cantidad de defectos podrían no coincidir con `horas_reales` o `defectos_detectados`. Resolverlo requiere una transacción entre `Manager` distintos, fuera del alcance, igual que en HU-35 y HU-36.
- **Integrantes por texto.** Como E4 no está implementada, un mismo integrante escrito de dos formas distintas (por ejemplo "Ana" y "ana") aparece dos veces, y un integrante sin esfuerzos registrados no aparece.
- **Valores no congelados.** Sin cambios respecto de HU-32: si se reestima o reabre un ítem de un sprint cerrado, cambian sus métricas y la velocidad del reporte.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Ejemplo de TEC-03 con tres esfuerzos y un defecto | El JSON del ejemplo 1 |
| Proyecto recién creado, sin datos | El JSON del ejemplo 2: listas vacías, cantidades y total en 0, cuatro métricas en `null` |
| Proyecto sin fecha de inicio | `fecha_inicio` vacío, como en HU-18 |
| Proyecto con ítems y sin sprints | Métricas de los ítems; esfuerzos, defectos e integrantes vacíos |
| Sprints sin esfuerzos | `esfuerzos` vacío, total 0 e integrantes vacío |
| Tres sprints, abiertos y cerrados, con esfuerzos | Todos los esfuerzos, ordenados por sprint y por ID; `total_horas` igual a `horas_reales` |
| Un integrante con esfuerzos en varios sprints | Aparece una sola vez, en la posición de su primer esfuerzo |
| Integrantes "Ana" y "ana" | Aparecen los dos (límite conocido) |
| Defectos en los tres estados | Aparecen todos, ordenados por ID |
| Esfuerzos y defectos de otro proyecto | No se incluyen, y sus integrantes tampoco |
| Horas con decimales (0.1, 0.2) | El total exacto en el dominio y redondeado a un decimal en la API |
| Generar el reporte dos veces sin cambios | El mismo resultado; no se modifica nada |
| Modificar las listas del reporte devuelto | Lo guardado no cambia |
| Proyecto inexistente, 0 o negativo | `project.ErrProjectNotFound`, sin otras lecturas; 404 |
| Falla la lectura del proyecto, las métricas, los sprints, los esfuerzos o los defectos | Ese error, sin las lecturas siguientes ni reporte parcial; 500 |
| HTTP: `{id}` no numérico | 400 |

## Condiciones de error

| Condición | Error | HTTP |
|---|---|---|
| `{id}` que no es un número entero | `errInvalidID` | 400 |
| No existe el proyecto (incluidos 0 y negativos) | `project.ErrProjectNotFound` | 404 |
| Falla el repositorio de proyectos, ítems, sprints, esfuerzos o defectos | El error del repositorio, en el orden de RN-38.3 | 500, sin detalle |
| Una parte está vacía o una métrica no está disponible | No es un error (RN-38.7) | 200 |

## Criterios de aceptación

- **CA-38.1** El reporte de un proyecto incluye sus datos (nombre, descripción, estado y fecha de inicio), sus métricas (story points planificados y completados, velocidad, historias planificadas y completadas, horas estimadas y reales, desviación y defectos detectados y resueltos), los esfuerzos registrados con su total de horas, los defectos y los integrantes que registraron esfuerzo en el proyecto.
- **CA-38.2** Un proyecto sin datos genera un reporte con las secciones vacías y las métricas que no se pueden calcular indicadas como no disponibles, no un error.
- **CA-38.3** Generar el reporte de un proyecto inexistente devuelve error.
- **CA-38.4** Los datos del reporte coinciden con las consultas del proyecto, de sus métricas, de sus esfuerzos y de sus defectos con los mismos datos, y el total de horas de los esfuerzos es igual a las horas reales de las métricas.
- **CA-38.5** Cada integrante que registró esfuerzo en el proyecto aparece una sola vez, y no aparecen esfuerzos, defectos ni integrantes de otros proyectos.
- **CA-38.6** Generar el reporte no modifica nada.
- **CA-38.7** El reporte se genera desde la API y desde la sección "Reportes" de la interfaz web.

## Comparación con el acta de Planning

**No hay diferencias de alcance con el acta del Sprint 4.** El acta asigna HU-38 a Cardozo Leandro con **5 SP**, con dependencias HU-35, HU-27 y HU-31, y la justifica así: "Reúne las métricas (HU-35), los esfuerzos (HU-27) y los defectos (HU-31) en un reporte del proyecto; paquete nuevo, endpoint y pantalla". La SDD hace eso: paquete `internal/report`, endpoint `/reporte` y sección "Reportes", reutilizando las tres operaciones sin cálculos nuevos. Las tres dependencias están en `main` (PR #126, #141 y #145).

Precisiones registradas:

1. **Integrantes (tema 5).** El reporte lista los integrantes que registraron esfuerzo en el proyecto (HU-26 y HU-27), y E4 sigue fuera de la entrega, como decidió el Planning.
2. **Datos del proyecto.** El reporte incluye los datos del proyecto, que pide CA-38.1 del Issue y que el acta no menciona porque no requieren otra historia (HU-03 y HU-18 ya están en `main`).
3. **Reutilización por HU-39 y HU-40.** El acta dice que HU-39 "reutiliza su estructura" y que HU-40 "genera un PDF con el reporte de HU-38": por eso las partes van en tipos separados y el contrato HTTP queda documentado con ejemplos. No se implementa nada de esas historias.
4. **Issue #56.** El tema 5 indica que el Issue #56 se actualiza con el criterio de los integrantes antes de la SDD; los criterios del Issue se cierran con CA-38.1 a CA-38.7 antes de la aprobación.

## Comparación con los criterios del Issue #56

Los criterios del Issue #56 se cierran con CA-38.1 a CA-38.7 antes de la aprobación de esta SDD.

| CA original del Issue #56 | En esta SDD | Ajuste |
|---|---|---|
| CA-38.1 Datos del proyecto, integrantes, historias planificadas y completadas, estimaciones, esfuerzo registrado, métricas y defectos | CA-38.1 y CA-38.5 | Integrantes: los que registraron esfuerzo (tema 5). Historias planificadas y completadas y estimaciones: con las métricas (M1, M2, M4 y M8 y sus cantidades), sin lista de ítems |
| CA-38.2 Un proyecto sin datos genera un reporte con las secciones vacías | CA-38.2 | Se precisa que las métricas no calculables se indican como no disponibles |
| CA-38.3 Un proyecto inexistente devuelve error | CA-38.3 | Sin cambios |
| — | CA-38.4 a CA-38.7 (nuevos) | Coincidencia con las consultas existentes, integrantes sin repetir y aislamiento, solo lectura, API e interfaz |

## Decisiones tomadas en la redacción

1. **Paquete nuevo `internal/report`**, como `internal/dashboard` en HU-36: el reporte no es una métrica ni un resumen, y HU-39 y HU-40 tienen dónde sumar lo suyo.
2. **Endpoint nuevo `/reporte`** en lugar de ampliar `/metricas` o `/resumen`: son datos distintos y esos contratos ya los usan HU-36 y HU-37.
3. **Partes con el formato de los endpoints existentes.** `proyecto`, `metricas`, `esfuerzos` y `defectos` reutilizan las respuestas de HU-03, HU-35, HU-27 y HU-31, así que quien ya consume esos endpoints lee el reporte sin aprender un formato nuevo, y la coincidencia de RN-38.8 se verifica directamente. Por eso `metricas` repite `proyecto_id`. *Alternativa descartada:* un formato plano propio del reporte, que duplicaría campos y nombres.
4. **Tipos separados por parte** (`Efforts`, `Defects` y `Members`), para que HU-39 los reutilice con `metrics.SprintMetrics`. *Alternativa descartada:* usar `effort.Summary` y `[]defect.Defect` directamente, que ataría el reporte a los tipos de consulta de otros paquetes y no tendría dónde agregar algo propio del reporte.
5. **`metrics.Manager.Project` sin cambios.** Se reutiliza la consulta única de HU-35 tal como está, y las lecturas que repite quedan como límite conocido, con la tabla exacta de lecturas por fuente. Esta historia no modifica `internal/metrics`, porque en paralelo otra historia modifica `internal/metrics/summary.go`. *Alternativa descartada:* agregar a `internal/metrics` una función que calcule M1 a M10 a partir de datos ya leídos, para leer cada fuente una sola vez a nivel de repositorio; cambia `internal/metrics` y se superpone con ese trabajo en paralelo.
6. **Esfuerzos sprint por sprint.** `internal/effort` no tiene una consulta por proyecto; se usa `ListBySprint` sobre cada sprint del proyecto, la misma lectura que usa M5 (RN-35.5), así el total coincide exactamente con las horas reales. *Alternativa descartada:* agregar `ListByProject` a `internal/effort`, que cambia su repositorio y sus tests.
7. **Integrantes en orden de primera aparición** en la lista de esfuerzos. Es determinista y no requiere una regla de orden alfabético con mayúsculas y tildes que ni el Issue ni el acta definen. *Alternativa descartada:* orden alfabético con `sort.Strings`, que pone las palabras con tilde inicial al final.
8. **Sin totales por integrante ni por sprint.** Serían cálculos nuevos que el Issue no pide.
9. **Historias planificadas y completadas y estimaciones con las métricas.** CA-38.1 del Issue las pide, y las métricas de HU-35 ya las informan: M1, M2, M4 y M8, con `items_totales`, `items_terminados`, `items_sin_estimar` e `items_sin_horas_estimadas`. El reporte no lleva la lista de ítems del backlog. *Alternativa descartada:* agregar la lista de ítems como sexta parte del reporte, que es otra fuente que el objetivo de la historia no incluye.
10. **Sección "Reportes" nueva** con ids propios, después de "Dashboard", sin tocar "Métricas" ni "Ver resumen", y reutilizando las funciones de `app.js` sin modificarlas, porque HU-37 trabaja en paralelo sobre los mismos archivos.
11. **Story Points.** 5 SP, según el Planning del Sprint 4; no se divide.
12. **Sin librerías nuevas.** Solo la librería estándar de Go; godog se mantiene en los tests. El PDF es de HU-40.

HU-38: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
