# Métricas del proyecto y sus fórmulas (TEC-03)

> Issue del repositorio: #33

## Objetivo

Definir la fórmula exacta de cada métrica mínima que exige la consigna, como insumo de la épica E9 (HU-32 a HU-35), del dashboard (HU-36 y HU-37), de los reportes (HU-38 y HU-39) y del informe final (TEC-08). Este documento no agrega código: cada historia de E9 implementa sus métricas citando la fórmula de acá en su SDD.

## Alcance

Las nueve métricas mínimas de la consigna: Story Points planificados y completados, velocidad del equipo, horas estimadas y reales, desviación entre esfuerzo estimado y real, porcentaje de historias completadas, y defectos detectados y resueltos. Cada una se calcula **por sprint** y **por proyecto**, salvo la velocidad, que es solo del proyecto.

## Datos de entrada

| Dato | Origen | Estado |
|---|---|---|
| Story Points del ítem (`StoryPoints`; 0 = sin estimar) | Ítem del backlog (HU-04, HU-07; HU-20 los asigna) | En `main` (HU-20 en el Sprint 3) |
| Horas estimadas del ítem (`EstimatedHours`; 0 = sin estimar) | Ítem del backlog (HU-07) | En `main` |
| Estado del ítem (`PENDIENTE`, `EN_CURSO`, `TERMINADO`) | Ítem del backlog (HU-08, HU-12) | En `main` |
| Sprint del ítem (`SprintID`; 0 = sin asignar) | Ítem del backlog (HU-11) | En `main` |
| Estado del sprint (`ABIERTO`, `CERRADO`) | Sprint (HU-10, HU-13) | En `main` |
| Horas trabajadas de cada registro de esfuerzo, con su ítem y su sprint | Registro de esfuerzo (HU-26) | Sprint 3 |
| Sprint de detección, estado y sprint de resolución de cada defecto | Defecto (HU-29 y HU-30) | Sprint 3 (HU-29) y Sprint 4 (HU-30) |

Los nombres de los campos de esfuerzo y defectos los fija la SDD de cada historia; este documento usa solo su significado.

## Conjuntos de ítems

- **Ítems del sprint S:** los ítems con `SprintID = S`.
- **Ítems del proyecto P:** todos los ítems del backlog de P, asignados o no a un sprint (decisión del equipo).
- **Completado:** un ítem en estado `TERMINADO`.

## Métricas

En las fórmulas, *I* es el conjunto de ítems (del sprint o del proyecto, según el nivel).

| # | Métrica | Fórmula | Unidad | Nivel |
|---|---|---|---|---|
| M1 | Story Points planificados | Σ `StoryPoints` de los ítems de *I* | SP | Sprint y proyecto |
| M2 | Story Points completados | Σ `StoryPoints` de los ítems de *I* en `TERMINADO` | SP | Sprint y proyecto |
| M3 | Velocidad del equipo | Promedio de M2 de los últimos 3 sprints `CERRADO` del proyecto, o de todos los cerrados si hay menos de 3 | SP por sprint | Proyecto |
| M4 | Horas estimadas | Σ `EstimatedHours` de los ítems de *I* | horas | Sprint y proyecto |
| M5 | Horas reales | Σ horas trabajadas de los registros de esfuerzo de los ítems de *I* (en el nivel sprint, los registros de ese sprint) | horas | Sprint y proyecto |
| M6 | Desviación en horas | M5 − M4 | horas | Sprint y proyecto |
| M7 | Desviación porcentual | (M5 − M4) / M4 × 100 | % | Sprint y proyecto |
| M8 | Porcentaje de historias completadas | cantidad de ítems de *I* en `TERMINADO` / cantidad de ítems de *I* × 100 | % | Sprint y proyecto |
| M9 | Defectos detectados | Cantidad de defectos cuyo sprint de detección es S (en el proyecto, todos los defectos del proyecto) | defectos | Sprint y proyecto |
| M10 | Defectos resueltos | Cantidad de defectos en estado resuelto cuyo sprint de resolución es S (en el proyecto, todos los resueltos del proyecto) | defectos | Sprint y proyecto |

La consigna nombra la desviación como una sola métrica; se informa de las dos formas (M6 y M7), por decisión del equipo.

### Reglas de cálculo

- **Últimos 3 sprints (M3):** los sprints cerrados se ordenan por ID descendente. Como un proyecto tiene un solo sprint abierto a la vez (HU-10), el ID sigue el orden en que se crearon.
- **Desviación (M6 y M7):** un valor positivo significa que se trabajó más de lo estimado, y uno negativo, menos.
- **Redondeo:** las métricas se calculan sin redondear. Se muestran con un decimal en la API, la interfaz y los reportes; los Story Points, las cantidades de ítems y las de defectos son enteros.

## Datos faltantes

| Caso | Resultado |
|---|---|
| Ítem con `StoryPoints = 0` | No suma en M1 ni M2. Se informa la cantidad de ítems sin Story Points |
| Ítem con `EstimatedHours = 0` | No suma en M4. Se informa la cantidad de ítems sin horas estimadas |
| Sprint o proyecto sin ítems | M1, M2 y M4 valen 0. M8 es **no disponible**, sin error |
| Proyecto sin sprints cerrados | M3 es **no disponible** |
| Sin registros de esfuerzo | M5 vale 0. M6 y M7 son **no disponibles** |
| M4 = 0 (sin horas estimadas) | M6 y M7 son **no disponibles** |
| Sin defectos | M9 y M10 valen 0 |
| Sprint o proyecto inexistente | Error de "no existe" (404 en la API), como en el resto de la aplicación |

"No disponible" no es un error: la consulta responde igual con las demás métricas, y esa métrica se marca como no disponible (HU-35, CA-35.2).

## Ejemplo

Sprint 1 de un proyecto, cerrado, con cuatro ítems:

| Ítem | SP | Horas estimadas | Estado | Horas reales registradas |
|---|---|---|---|---|
| A | 5 | 10 | `TERMINADO` | 12 |
| B | 3 | 6 | `TERMINADO` | 5 |
| C | 2 | 0 | `EN_CURSO` | 3 |
| D | 0 | 4 | `PENDIENTE` | 0 |

Con un defecto detectado en el Sprint 1 y resuelto en el mismo sprint:

- M1 = 5 + 3 + 2 = **10 SP** (D no suma: 1 ítem sin Story Points).
- M2 = 5 + 3 = **8 SP**.
- M4 = 10 + 6 + 4 = **20 h** (C no suma: 1 ítem sin horas estimadas).
- M5 = 12 + 5 + 3 = **20 h**.
- M6 = 20 − 20 = **0 h**; M7 = 0 / 20 × 100 = **0,0 %**.
- M8 = 2 / 4 × 100 = **50,0 %**.
- M9 = **1**; M10 = **1**.
- M3, si es el único sprint cerrado: 8 / 1 = **8,0 SP por sprint**. Con tres sprints cerrados que completaron 8, 10 y 12 SP: (8 + 10 + 12) / 3 = **10,0 SP por sprint**.

## Trazabilidad

| Métrica | Requerimiento de la consigna | Historia que la implementa | Criterios |
|---|---|---|---|
| M1 y M2 | Story Points planificados y completados | HU-32 (y HU-43 si se divide) | CA-32.1, CA-32.2, CA-32.4 |
| M3 | Velocidad del equipo | HU-32 (HU-44 si se divide) | CA-32.3 |
| M4, M5, M6 y M7 | Horas estimadas y reales; desviación | HU-33 (HU-45 y HU-46 si se divide) | CA-33.1, CA-33.2, CA-33.3 |
| M8 | Porcentaje de historias completadas | HU-34 | CA-34.1 |
| M9 y M10 | Defectos detectados y resueltos | HU-34 (HU-47 si se divide) | CA-34.2, CA-34.3 |
| M1 a M10 juntas | Todas las métricas en una consulta | HU-35 | CA-35.1, CA-35.2, CA-35.3 |

Los IDs HU-43 a HU-47 son los de la división en historias más atómicas del acta de Planning del Sprint 3, que todavía se tiene que confirmar con los profesores.

Issue #33 → Criterios de Aceptación (CA-TEC03.1 a CA-TEC03.3) → este documento → SDD de HU-32 a HU-35 → Tests → Código Go

## Criterios de aceptación del Issue

- **CA-TEC03.1** (nombre, fórmula, unidad y datos de entrada): secciones "Datos de entrada" y "Métricas".
- **CA-TEC03.2** (qué hacer cuando falta un dato): sección "Datos faltantes".
- **CA-TEC03.3** (el equipo aprueba el documento): aprobación del Pull Request por otro integrante.
