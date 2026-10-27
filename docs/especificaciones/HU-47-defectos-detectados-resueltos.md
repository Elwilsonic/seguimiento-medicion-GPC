# HU-47 · Calcular los defectos detectados y resueltos

> Issue del repositorio: #128

Aprobada por: Peñalbé Hernán (Agile Enabler), 27/10/2026, con la dependencia de HU-31 registrada en "Comparación con el acta de Planning"; el nombre y la firma de la lectura de defectos se confirman con Sosa (HU-31).

## Historia de usuario

Como Agile Enabler, quiero calcular la cantidad de defectos detectados y resueltos de un sprint y de un proyecto, para evaluar la calidad.

## Objetivo

Calcular las métricas M9 (defectos detectados) y M10 (defectos resueltos) de `docs/metricas.md` (TEC-03), por sprint y por proyecto, en el paquete `internal/metrics`. Agregarlas a los endpoints `GET /api/sprints/{id}/metricas` y `GET /api/proyectos/{id}/metricas` y a la sección "Métricas" de la interfaz web, como hicieron HU-32 y HU-34.

Esta historia es la segunda parte de la división de HU-34 (acta de Planning del Sprint 3): HU-34 calculó el porcentaje de historias completadas (M8) y HU-47 calcula los defectos.

## Alcance

- **Incluye:**
  - el cálculo de M9 y M10 de un sprint y de un proyecto;
  - los campos nuevos en las dos respuestas de métricas;
  - el texto de la sección "Métricas" de la interfaz.
- **No incluye:**
  - el cambio de estado de un defecto ni el registro del sprint de resolución (HU-30);
  - la consulta de defectos para el usuario, con filtros (HU-31);
  - la operación que lee los defectos de un proyecto en los repositorios (la agrega HU-31; ver "Dependencias");
  - horas y desviación (HU-33);
  - la consulta única de M1 a M10 (HU-35);
  - gráficos (HU-37);
  - reportes (HU-38 y HU-39);
  - congelar los valores al cerrar un sprint (mismo límite conocido que HU-32 y HU-34);
  - cambios en el esquema de la base.

## Dependencias

| Historia | Qué aporta | Estado |
|---|---|---|
| HU-29 (#47) | El defecto, con `ItemID`, `DetectionSprintID`, `Status` y `ResolutionSprintID` | En `main` |
| HU-30 (#48) | El paso a `RESUELTO` y el sprint de resolución | Pendiente (Sosa) |
| HU-31 (#49) | La lectura de los defectos de un proyecto en `defect.Manager` y en los repositorios (memoria y PostgreSQL) | Pendiente (Sosa) |
| HU-32 y HU-34 | El paquete `internal/metrics`, los endpoints y la sección "Métricas" | En `main` |

HU-47 no lee los repositorios de defectos por su cuenta: usa la operación de HU-31, para que las dos historias no agreguen dos lecturas distintas en los mismos archivos. Esta SDD supone que esa operación es:

```go
// En defect.Manager (HU-31). Devuelve copias de todos los defectos del
// proyecto, en cualquier estado; un proyecto inexistente devuelve
// project.ErrProjectNotFound.
func (m *Manager) ListByProject(projectID int) ([]Defect, error)
```

El nombre y la firma se confirman con Sosa antes de aprobar esta SDD. Si HU-31 la define distinto, solo cambia la interfaz `defectLister` de esta historia (ver "Restricciones").

Los tests de HU-47 no necesitan HU-30: guardan defectos en `RESUELTO` con el `Save` del repositorio en memoria. Pero la métrica M10 solo tiene sentido en la aplicación cuando HU-30 permite resolver un defecto.

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `SprintDefects(sprintID int) (Defects, error)` | ID de un sprint | M9 y M10 del sprint |
| `ProjectDefects(projectID int) (Defects, error)` | ID de un proyecto | M9 y M10 del proyecto |

Las dos están en el `Manager` de `internal/metrics`, en un archivo nuevo `defects.go`.

### Salidas esperadas

- **`Defects`**, por valor y sin slices ni mapas:
  - `Detected`: M9, entero;
  - `Resolved`: M10, entero.
- **Error:** el que indica el motivo (ver "Condiciones de error").

### HTTP (TEC-04)

Se agregan dos campos a las dos respuestas existentes, sin cambiar los que ya tienen:

| Campo | Tipo | Descripción |
|---|---|---|
| `defectos_detectados` | entero | M9 |
| `defectos_resueltos` | entero | M10 |

Ejemplo de `GET /api/sprints/{id}/metricas` con el sprint del ejemplo de TEC-03 (un defecto detectado y resuelto en el mismo sprint):

```json
{
  "sprint_id": 1,
  "sp_planificados": 10,
  "sp_completados": 8,
  "items_sin_estimar": 1,
  "items_totales": 4,
  "items_terminados": 2,
  "porcentaje_completadas": 50,
  "defectos_detectados": 1,
  "defectos_resueltos": 1
}
```

Los errores no cambian respecto de HU-32 y HU-34:

| Caso | Código | Mensaje |
|---|---|---|
| `{id}` no numérico | 400 | El de `errInvalidID` |
| Sprint inexistente | 404 | El de `sprint.ErrSprintNotFound` |
| Proyecto inexistente | 404 | El de `project.ErrProjectNotFound` |
| Falla de un repositorio | 500 | El de `errInternal`, sin detalle |

### Interfaz web (TEC-04)

En la sección "Métricas", el texto que arma "Ver métricas" suma una frase para el proyecto y otra para el sprint actual: por ejemplo, "Defectos: 3 detectados, 1 resuelto(s).". Sin defectos muestra "Defectos: 0 detectados, 0 resuelto(s).". No hay formularios ni rutas nuevas.

## Reglas de negocio

1. **RN-47.1 Detectados de un sprint (M9).** Son los defectos con `DetectionSprintID` igual al sprint, en cualquier estado.
2. **RN-47.2 Resueltos de un sprint (M10).** Son los defectos en estado `RESUELTO` con `ResolutionSprintID` igual al sprint.
3. **RN-47.3 Detectados de un proyecto.** Son todos los defectos del proyecto, en cualquier estado. Un defecto pertenece al proyecto de su historia (`ItemID`), porque `Defect` no guarda el proyecto.
4. **RN-47.4 Resueltos de un proyecto.** Son todos los defectos del proyecto en estado `RESUELTO`.
5. **RN-47.5 Solo cuenta el estado resuelto.** Un defecto `ABIERTO` o `EN_CORRECCION` no cuenta como resuelto, aunque tenga cargado un sprint de resolución.
6. **RN-47.6 Detección y resolución en sprints distintos.** Un defecto detectado en un sprint y resuelto en otro cuenta como detectado en el primero y como resuelto en el segundo.
7. **RN-47.7 Sin defectos.** Un sprint o un proyecto sin defectos no es un error: M9 y M10 valen 0. No hay "no disponible" (tabla "Datos faltantes" de TEC-03).
8. **RN-47.8 Enteros.** M9 y M10 son cantidades: no se redondean ni llevan decimales, ni en la API ni en la interfaz.
9. **RN-47.9 Sprints abiertos y cerrados.** Se calcula igual para un sprint `ABIERTO` o `CERRADO`, con el estado actual de los defectos.
10. **RN-47.10 Solo lectura.** Calcular no modifica defectos, ítems, sprints ni proyectos.

## Decisión sobre entidades referenciadas

HU-47 referencia por ID un sprint o un proyecto, y **sí valida su existencia**, igual que HU-32 y HU-34. Sin la validación, un ID mal escrito daría "0 defectos" y ocultaría el error.

- **Sprint:** con `sprint.Manager.Get`, que devuelve `sprint.ErrSprintNotFound`. Después se leen los defectos del proyecto del sprint y se filtran por sprint.
- **Proyecto:** con la lectura de HU-31, que valida el proyecto y devuelve `project.ErrProjectNotFound`.

No se validan uno por uno la historia ni los sprints de cada defecto: HU-29 ya los validó al registrarlo.

## Restricciones

- **Código nuevo.** `internal/metrics/defects.go`, con:
  - el tipo `Defects`;
  - la interfaz interna `defectLister` (la operación de HU-31);
  - la función interna que cuenta;
  - los métodos `SprintDefects` y `ProjectDefects`.
- **Constructor.** `metrics.NewManager` recibe un tercer parámetro con la lectura de defectos: `NewManager(items, sprints, defects)`. Cambian sus llamadas:
  - `internal/httpapi/server.go` (`NewHandler` ya recibe el `defect.Manager` desde HU-29, así que su firma no cambia);
  - el helper `newMetricsEnvWith` de los tests de `internal/metrics`.

  El cambio de firma va en el RED, con la firma vacía, para que el GREEN no toque tests.
- **HTTP.** `internal/httpapi/metrics.go`: los dos handlers llaman también a `SprintDefects` o `ProjectDefects` y agregan los dos campos. Las rutas no cambian.
- **Interfaz.** `internal/web/app.js`: el texto de "Ver métricas". `index.html` no cambia.
- **Tres lecturas por pedido.** Cada endpoint lee el backlog dos veces (HU-32 y HU-34) y los defectos una vez. Si alguien modifica un dato justo entre las lecturas, los valores de una misma respuesta pueden corresponder a momentos distintos. Es el mismo límite conocido de HU-34, y la consulta única de HU-35 lo resuelve.
- **Concurrencia.** `metrics` no tiene estado propio: cada lectura queda protegida por el mutex del `Manager` que la hace, como en HU-32 y HU-34. No se agrega un test de concurrencia.
- **RED.**
  - Firmas vacías de `Defects`, `SprintDefects` y `ProjectDefects` y el tercer parámetro de `NewManager`, para que el RED compile y falle por aserciones.
  - `go vet ./...` antes del commit RED.
  - Los campos nuevos de la API y el texto de la interfaz van en el GREEN.
  - El RED se hace cuando la operación de HU-31 esté en `main`.
- **Tests.**
  - Unitarios en `internal/metrics/defects_test.go`.
  - godog en `internal/metrics/defects_features_test.go`, con el runner `TestFeatures_HU47DefectosDetectadosResueltos` sobre `features/HU-47-defectos-detectados-resueltos.feature`, con el CA en el nombre de cada escenario.
  - API en `internal/httpapi/metrics_test.go`.
  - Interfaz en `internal/httpapi/web_test.go`.
- **Librerías.** Solo la librería estándar; godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Ejemplo de TEC-03: un defecto detectado y resuelto en el sprint | M9 = 1, M10 = 1 |
| Sprint sin defectos | M9 = 0, M10 = 0, sin error |
| Proyecto sin defectos | M9 = 0, M10 = 0, sin error |
| Defecto `ABIERTO` | Cuenta en M9, no en M10 |
| Defecto `EN_CORRECCION` con sprint de resolución cargado | Cuenta en M9, no en M10 |
| Detectado en S1 y resuelto en S2 | S1: M9 = 1, M10 = 0. S2: M9 = 0, M10 = 1. Proyecto: M9 = 1, M10 = 1 |
| Defecto de otro proyecto | No cuenta |
| Defecto de una historia terminada | Cuenta igual |
| Sprint `CERRADO` | Se calcula igual, con el estado actual |
| Sprint inexistente, 0 o negativo | `sprint.ErrSprintNotFound` (404) |
| Proyecto inexistente, 0 o negativo | `project.ErrProjectNotFound` (404) |
| Falla al leer los defectos o el sprint | Se propaga el error (500 sin detalle) |
| Consultar dos veces | Mismo resultado; no se modifica nada |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Sprint inexistente | `sprint.ErrSprintNotFound` | No se calcula nada |
| Proyecto inexistente | `project.ErrProjectNotFound` | No se calcula nada |
| Falla al leer defectos, ítems o sprints | El error del repositorio | No se calcula nada |

## Criterios de aceptación

- **CA-47.1** Los defectos detectados de un sprint son los que tienen ese sprint como sprint de detección, y los resueltos son los que están resueltos con ese sprint como sprint de resolución.
- **CA-47.2** Los defectos detectados de un proyecto son todos sus defectos, y los resueltos son todos sus defectos resueltos.
- **CA-47.3** Un sprint o un proyecto sin defectos no produce error: los dos valores son cero.
- **CA-47.4** Un defecto abierto o en corrección no cuenta como resuelto, y un defecto detectado en un sprint y resuelto en otro cuenta en cada uno donde corresponde.
- **CA-47.5** Consultar un sprint o un proyecto inexistente devuelve error, y consultar no modifica nada.
- **CA-47.6** Los defectos detectados y resueltos se ven desde la API y desde la interfaz web, como números enteros.

## Comparación con el acta de Planning

El acta del Sprint 4 compromete HU-47 con **2 SP**, Peñalbé Hernán como responsable y estas dependencias: HU-29 (en `main`) y HU-30 (defectos resueltos). La diferencia con esta SDD es una sola:

1. **HU-31 también es dependencia.** Hoy `defect.Repository` solo tiene `Save`, y HU-47 necesita leer los defectos de un proyecto. Esta SDD usa la lectura que agrega HU-31, en lugar de crear una propia. El orden de trabajo del acta ya pone HU-31 y HU-47 juntas, después de HU-30, así que el plan del Sprint no cambia. Lo que cambia es que el RED de HU-47 espera a HU-31 y no solo a HU-30.

El resto coincide.

## Comparación con los criterios del Issue #128 (borrador)

| CA en el Issue #128 | En esta SDD | Ajuste |
|---|---|---|
| CA-47.1 Los defectos detectados se cuentan por sprint de detección y los resueltos por sprint de resolución. | CA-47.1, CA-47.2 y CA-47.4 | Se separa por nivel (sprint y proyecto) y se precisa que solo cuenta el estado resuelto |
| CA-47.2 Sin defectos el resultado es cero. | CA-47.3 | Sin cambios de fondo |
| — | CA-47.5 y CA-47.6 (nuevos) | Errores, solo lectura, API e interfaz |

Texto propuesto para el Issue #128. Se aplica con autorización, después de aprobar la SDD:

```markdown
## Criterios de aceptación

- [ ] CA-47.1 Los defectos detectados de un sprint son los que tienen ese sprint como sprint de detección, y los resueltos son los que están resueltos con ese sprint como sprint de resolución.
- [ ] CA-47.2 Los defectos detectados de un proyecto son todos sus defectos, y los resueltos son todos sus defectos resueltos.
- [ ] CA-47.3 Un sprint o un proyecto sin defectos no produce error: los dos valores son cero.
- [ ] CA-47.4 Un defecto abierto o en corrección no cuenta como resuelto, y un defecto detectado en un sprint y resuelto en otro cuenta en cada uno donde corresponde.
- [ ] CA-47.5 Consultar un sprint o un proyecto inexistente devuelve error, y consultar no modifica nada.
- [ ] CA-47.6 Los defectos detectados y resueltos se ven desde la API y desde la interfaz web, como números enteros.
```

Además, en el Issue se cargan los Story Points (2) y en "Dependencias" se agrega que HU-31 aporta la lectura de defectos.

## Decisiones tomadas en la redacción

1. **La lectura de defectos es de HU-31**, no de HU-47 (ver "Dependencias"). *Alternativa descartada:* que HU-47 agregue su propia lectura a `defect.Repository`. Se descarta porque chocaría con HU-31 en los mismos archivos (repositorio, memoria y PostgreSQL), mientras Sosa trabaja en ellos.
2. **Interfaz `defectLister` en `metrics`.** `metrics` depende de una interfaz con la operación de HU-31, no del `defect.Manager` concreto. Así los tests pueden simular una falla de lectura sin una base, y si HU-31 cambia el nombre de la operación, solo cambia la interfaz.
3. **Métodos separados (`SprintDefects` y `ProjectDefects`)**, como HU-34. *Motivo:* no tocar el código de HU-32 y HU-34. *Costo:* una lectura más por pedido.
4. **Mismos endpoints** con dos campos nuevos, como HU-34. *Alternativa descartada:* rutas nuevas, que HU-35 tendría que volver a juntar.
5. **El estado manda para M10.** Se cuenta como resuelto solo un defecto en `RESUELTO` con sprint de resolución igual al consultado. Así M10 no depende de si HU-30 borra o conserva el sprint de resolución cuando un defecto deja de estar resuelto.

HU-47: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
