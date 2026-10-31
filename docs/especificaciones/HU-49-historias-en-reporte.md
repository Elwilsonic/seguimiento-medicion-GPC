# HU-49 · Ver las historias y sus estimaciones en el reporte

> Issue del repositorio: #170

Aprobada por: Peñalbé Hernán (Agile Enabler), 31/10/2026, sin ajustes.

## Historia de usuario

Como Agile Enabler, quiero ver en el reporte del proyecto o del sprint cada historia con su estado y su estimación, para saber qué se planificó, qué se completó y cuánto se estimó.

**Épica:** E11 · Reportes

**Prioridad:** MEDIA (la del Issue #170)

**Story Points:** 3 (como HU-39: reutiliza la estructura de los reportes de HU-38 y HU-39)

## Objetivo

La consigna (requerimiento 9) pide que el reporte de un proyecto o de un sprint incluya "historias planificadas y completadas" y "estimaciones". HU-38 y HU-39 las informan solo como cantidades y sumas de las métricas (M1, M2, M4 y M8), sin el detalle de cada historia (decisión 9 de la SDD de HU-38). HU-49 agrega ese detalle:

- en el reporte del proyecto, cada historia del backlog con su estado y su estimación;
- en el reporte de un sprint, cada historia asignada al sprint;
- en los dos, la cantidad de historias y la de completadas;
- en el PDF del reporte del proyecto (HU-40) y en la sección "Reportes" de la interfaz.

## Alcance

- **Incluye:**
  - un tipo nuevo `report.Stories` y el campo `Stories` en `ProjectReport` y en `SprintReport`;
  - la parte `"historias"` en `GET /api/proyectos/{id}/reporte` y en `GET /api/sprints/{id}/reporte`;
  - la sección "Historias" en el PDF del reporte del proyecto;
  - la tabla de historias en los dos bloques de la sección "Reportes" de la interfaz.
- **Reutiliza sin cambios:**
  - `backlog.Manager.ListItems` (HU-06), que valida el proyecto y devuelve copias ordenadas por ID;
  - `itemResponse` y `toItemResponse` de `internal/httpapi/items.go` (el formato de `GET /api/items/{id}`, HU-05);
  - el armado del PDF de HU-40 (`pdfLine`, encabezados y salto de línea).
- **No incluye:**
  - el esfuerzo de cada historia (las horas reales siguen en la tabla de esfuerzos y en M5) ni totales por historia;
  - filtros u órdenes distintos del ID;
  - un PDF del reporte del sprint (HU-40 exporta solo el del proyecto);
  - cambios en los repositorios, en el esquema de la base, en las métricas ni en las demás partes de los reportes;
  - rutas nuevas.

## Contrato

### Entradas

- `projectID int` para el reporte del proyecto y `sprintID int` para el del sprint, como en HU-38 y HU-39. No hay entradas nuevas.

### Salidas esperadas

- **`report.Stories`** (`internal/report/stories.go`, nuevo):

  ```go
  type Stories struct {
      Stories   []backlog.Item // copias, ordenadas por ID
      Completed int            // cuántas están en TERMINADO
  }
  ```

- **`ProjectReport.Stories`:** todas las historias del backlog del proyecto, en cualquier estado, asignadas o no a un sprint.
- **`SprintReport.Stories`:** las historias del proyecto del sprint cuyo `SprintID` es el del sprint.
- **Cambia la firma** de `report.NewManager`: recibe además el `*backlog.Manager`.

  ```go
  func NewManager(projects *project.Manager, items *backlog.Manager, sprints *sprint.Manager,
      metrics *metrics.Manager, efforts *effort.Manager, defects *defect.Manager) *Manager
  ```

### HTTP (TEC-04)

Los dos reportes suman la parte `"historias"`, después de `"metricas"`. Cada historia tiene el formato de `GET /api/items/{id}` (`itemResponse`):

```json
"historias": {
  "historias": [
    {"id": 1, "proyecto_id": 1, "titulo": "Registrar venta", "descripcion": "…", "prioridad": "ALTA",
     "criterios": ["…"], "estado": "TERMINADO", "story_points": 3, "horas_estimadas": 8, "sprint_id": 1}
  ],
  "cantidad": 1,
  "completadas": 1
}
```

- Sin historias: `{"historias": [], "cantidad": 0, "completadas": 0}`, con `[]` y no `null`.
- Los códigos de respuesta no cambian: 200, 400 (ID no numérico), 404 (proyecto o sprint inexistente) y 500 (falla del almacenamiento, sin detalle).

### PDF (HU-40)

Después de "Métricas" y antes de "Esfuerzos", una sección "Historias" con una línea por historia y la cantidad:

```
Historias
#1 · Registrar venta · ALTA · TERMINADO · 3 SP · 8.0 h estimadas · sprint 1
#2 · Consultar stock · MEDIA · PENDIENTE · sin estimar · sin horas estimadas · sin sprint
Cantidad: 2 historia(s), 1 completada(s)
```

Sin historias: "Sin historias en el backlog." y "Cantidad: 0 historia(s), 0 completada(s)".

### Interfaz web (TEC-04)

En los dos bloques de la sección "Reportes", después del texto de las métricas:

- una tabla `reporte-historias` (proyecto) y `reporte-sprint-historias` (sprint), con las columnas ID, Título, Prioridad, Estado, SP, Horas estimadas y Sprint;
- el SP es "—" si la historia no está estimada y el sprint es "—" si no está asignada, como en la tabla del backlog (HU-20);
- debajo, "N historia(s), K completada(s)" (`reporte-cantidad-historias` y `reporte-sprint-cantidad-historias`);
- sin historias, la tabla muestra "Sin historias en el backlog." (proyecto) o "Sin historias en el sprint." (sprint), como las tablas de esfuerzos y defectos;
- los datos se escriben con `textContent`, nunca con `innerHTML`.

## Reglas de negocio

- **RN-49.1 Historias del proyecto:** el reporte del proyecto incluye todas las historias de su backlog, en cualquier estado y asignadas o no a un sprint. Es el mismo conjunto que usan M1 y M8 del proyecto ("ítems del proyecto", `docs/metricas.md`).
- **RN-49.2 Historias del sprint:** el reporte de un sprint incluye las historias del proyecto del sprint cuyo `SprintID` es el del sprint. Es el mismo conjunto que usan M1 y M8 del sprint ("ítems del sprint S").
- **RN-49.3 Datos de cada historia:** ID, título, descripción, prioridad, criterios, estado, Story Points, horas estimadas y sprint, tal como los devuelve el backlog. Una historia sin estimar tiene 0 Story Points; sin horas estimadas, 0 horas; sin sprint, `sprint_id` 0.
- **RN-49.4 Completada:** una historia está completada si su estado es `TERMINADO`, como en M2 y M8.
- **RN-49.5 Cantidades:** `cantidad` es la cantidad de historias de la lista y `completadas`, la de las que están en `TERMINADO`. Con los mismos datos, coinciden con `items_totales` e `items_terminados` de las métricas del mismo reporte.
- **RN-49.6 Orden:** las historias se ordenan por ID ascendente, como las devuelve `ListItems`.
- **RN-49.7 Orden de lectura:** las historias se leen al final, después de los defectos: proyecto → métricas → sprints → esfuerzos → defectos → historias (proyecto), y sprint → métricas → esfuerzos → defectos → historias (sprint). Si falla cualquier lectura, se devuelve ese error y el valor cero del reporte, sin reporte parcial ni lecturas siguientes (RN-38.3 y RN-39.3).
- **RN-49.8 Copias:** la lista de historias es nueva y cada historia es una copia con su propia lista de criterios: modificar el reporte devuelto no cambia lo guardado ni el reporte siguiente (RN-38.11).
- **RN-49.9 Sin cambios:** generar el reporte no modifica ninguna historia.
- **RN-49.10 Mismo formato que el backlog:** en la API, cada historia tiene el formato de `GET /api/items/{id}`, así que coincide con esa consulta.

## Decisión sobre entidades referenciadas

- **El proyecto y el sprint se validan como en HU-38 y HU-39:** el proyecto con `project.Manager.Get` y el sprint con `sprint.Manager.Get`, al principio. `ListItems` vuelve a validar el proyecto, que ya existe en ese punto.
- **Las historias no se validan una por una:** salen del backlog filtradas por proyecto (y, en el sprint, por `SprintID`), así que todas existen y pertenecen al proyecto.

## Decisiones de diseño

- **D1. Formato de `GET /api/items/{id}`.** Cada historia del reporte usa `itemResponse`, como las demás partes reutilizan el formato de su endpoint (RN-38.8). *Alternativa descartada:* un formato reducido solo con estado y estimación, que no coincide con ninguna consulta existente.
- **D2. Las historias se leen al final.** Así no cambia el orden de lectura de las partes de HU-38 y HU-39, ni los tests que lo verifican. *Alternativa descartada:* leerlas después de las métricas, que cambia el orden y obliga a reescribir esos tests.
- **D3. El sprint filtra las historias del proyecto por `SprintID`.** Se reutiliza `ListItems`, sin operaciones nuevas en los repositorios. *Alternativa descartada:* `sprint.ListCompletedItems`, que devuelve solo las `TERMINADO` y no las planificadas.
- **D4. "Planificadas" son las historias del conjunto de las métricas.** En el proyecto, todo el backlog; en el sprint, las asignadas. Así la cantidad coincide con M8 y no hace falta una definición nueva.
- **D5. HU-49 reemplaza la decisión 9 de HU-38** ("sin lista de ítems del backlog"). Esa decisión cumplía la consigna con las cantidades; la evaluación del proyecto contra la consigna mostró que el texto pide las historias, y HU-49 lo cubre sin cambiar las demás partes.
- **D6. Solo el PDF del proyecto.** HU-40 exporta el reporte del proyecto; el del sprint no tiene PDF y HU-49 no lo agrega.

## Restricciones

- Solo la biblioteca estándar, godog en los tests y `go-pdf/fpdf` en el PDF, como hasta ahora. No se agregan librerías.
- El código de negocio accede a las historias solo a través de `backlog.Manager`. Sin cambios en los repositorios ni en el esquema.
- **RED con armado:** la firma nueva de `report.NewManager`, el tipo `Stories` y los campos vacíos entran en el RED, para que los tests compilen; el armado en `internal/httpapi/server.go` también, porque es una llamada a la firma nueva. La lectura de las historias entra en el GREEN.
- **Tests existentes que cambian en el RED:** los que comparan el JSON completo de los reportes (`reports_test.go` y `sprint_reports_test.go`) y las líneas completas del PDF (`pdf_test.go`) suman la parte nueva a lo esperado. No se afloja ninguna aserción.

## Casos límite

- **Proyecto sin historias:** lista vacía, cantidades 0, "Sin historias en el backlog." en la interfaz y en el PDF.
- **Sprint sin historias asignadas:** lista vacía y cantidades 0.
- **Historia sin estimar o sin horas:** 0 en la API; "sin estimar" y "sin horas estimadas" en el PDF; "—" de SP en la interfaz.
- **Historia sin sprint:** aparece en el reporte del proyecto con `sprint_id` 0 ("sin sprint" en el PDF y "—" en la interfaz), y en ningún reporte de sprint.
- **Historias de otro proyecto o de otro sprint:** no aparecen.
- **Sprint cerrado con historias sin completar:** aparecen con su estado, no como completadas.
- **Títulos largos en el PDF:** la línea se parte, como las demás (HU-40).

## Condiciones de error

| Situación | Dominio | HTTP |
|---|---|---|
| Proyecto o sprint inexistente, ID 0 o negativo | `ErrProjectNotFound` o `ErrSprintNotFound`, sin leer las historias | 404 |
| `{id}` no numérico | — | 400 |
| Falla la lectura de las historias | el error del repositorio y el valor cero del reporte | 500, "error interno del servidor", y sin PDF a medias |

## Criterios de aceptación

- [ ] **CA-49.1:** el reporte del proyecto incluye todas las historias de su backlog, ordenadas por ID, cada una con su título, prioridad, estado, Story Points, horas estimadas y sprint, y la cantidad de historias y de completadas.
- [ ] **CA-49.2:** el reporte de un sprint incluye las historias asignadas al sprint, con los mismos datos y cantidades, y no las de otros sprints ni las sin asignar.
- [ ] **CA-49.3:** las cantidades coinciden con las métricas del mismo reporte (historias e historias terminadas), y cada historia coincide con su consulta en el backlog (`GET /api/items/{id}`).
- [ ] **CA-49.4:** un proyecto o un sprint sin historias muestra la lista vacía y cantidades 0; una historia sin estimar, sin horas o sin sprint se indica como tal.
- [ ] **CA-49.5:** si falla la lectura de las historias, se devuelve el error, sin reporte parcial; el reporte no modifica las historias y modificar lo devuelto no cambia lo guardado.
- [ ] **CA-49.6:** el PDF del reporte del proyecto incluye la sección "Historias" con cada historia y las cantidades.
- [ ] **CA-49.7:** la sección "Reportes" de la interfaz muestra la tabla de historias y las cantidades en el reporte del proyecto y en el del sprint.

## Comparación con los criterios iniciales del Issue #170

| Borrador | Definitivo | Cambio |
|---|---|---|
| CA-49.1 Cada historia del backlog con estado y estimación, y cantidades | CA-49.1 | Se precisan los datos de cada historia y el orden por ID |
| CA-49.2 Las historias del sprint | CA-49.2 | Se aclara que no aparecen las de otros sprints ni las sin asignar |
| CA-49.3 Cantidades y cada historia coinciden | CA-49.3 | Se precisa con qué métricas y con qué consulta |
| CA-49.4 Lista vacía e historia sin estimar | CA-49.4 | Se suman las historias sin horas y sin sprint |
| CA-49.5 Error sin reporte parcial | CA-49.5 | Se suman la no modificación y las copias (`AGENTS.md`) |
| CA-49.6 El PDF incluye las historias | CA-49.6 | Sin cambios de fondo |
| CA-49.7 La interfaz muestra las historias | CA-49.7 | Sin cambios de fondo |

## Comparación con las actas

- **No hay acta de Planning que mencione HU-49:** la historia se crea el 31/10/2026, después de la Review del Sprint 4 (30/10). Por decisión del Agile Enabler, se suma al Sprint 4, que termina con la entrega del 02/11; se registra en el acta de la Review y en el informe final, y el tag `sprint-4` se mueve al final.
- **Cambia una decisión registrada:** la decisión 9 de la SDD de HU-38 (D5 de esta SDD).

## Pruebas previstas

- **`internal/report/stories_test.go` (unitarios):**
  - reporte del proyecto con historias en distintos estados, asignadas y sin asignar, ordenadas por ID, con las cantidades;
  - reporte del sprint solo con sus historias (no las de otro sprint ni las sin asignar);
  - las cantidades coinciden con `Metrics.Completion.Items` y `Metrics.Completion.Finished`;
  - proyecto y sprint sin historias;
  - falla de la lectura de las historias en el proyecto y en el sprint, sin reporte parcial;
  - copias: modificar las historias o sus criterios en el reporte no cambia lo guardado ni el siguiente reporte.
- **`internal/report/stories_features_test.go`:** el runner de godog de `features/HU-49-historias-en-reporte.feature`, con caso normal, alternativo, límite y error.
- **`internal/report/pdf_test.go`:** las líneas de la sección "Historias" (con y sin estimación, sin sprint y sin historias) y la cantidad.
- **`internal/httpapi/reports_test.go` y `sprint_reports_test.go`:** la parte `"historias"` en el JSON esperado; cada historia igual a `GET /api/items/{id}`; `[]` sin historias; 500 si falla la lectura.
- **`internal/httpapi/web_test.go`:** las tablas y los textos nuevos en `index.html` y `app.js`, sin `innerHTML`.

## Archivos que cambian

- **Nuevos:** `internal/report/stories.go`, `stories_test.go` y `stories_features_test.go`; `features/HU-49-historias-en-reporte.feature`.
- **Modificados:**
  - `internal/report/manager.go`, `project.go`, `sprint.go` y `pdf.go`;
  - `internal/httpapi/server.go`, `reports.go` y `sprint_reports.go`;
  - `internal/web/index.html` y `app.js`;
  - los tests que construyen `report.NewManager` o comparan los reportes completos.
- **Documentación:** evidencia TDD, cobertura, README, registro de IA, documentación técnica (formato de los reportes), manual de usuario (sección 13 y su captura), acta de la Review del Sprint 4 e informe final.

## Límite conocido

Las lecturas no son atómicas, como en HU-38 y HU-39: si una historia cambia de estado entre la lectura de las métricas y la de las historias, `completadas` puede no coincidir con `items_terminados` en ese reporte.
