# Acta de Planning — Sprint 4 (La Entrega)

> Issue del repositorio: #133

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 4 — La Entrega
- **Fecha del Planning:** 27/10/2026
- **Período del Sprint:** 27/10/2026 al 02/11/2026 (desarrollo cerrado el 30/10, entrega el 02/11/2026)
- **Sprint Goal:** Ver las métricas del proyecto en el dashboard y generar sus reportes, con la gestión de defectos completa y la documentación final lista para la entrega.

## Participantes

| Rol | Integrante | Asistió |
|---|---|---|
| Product Architect | Profesores de la cátedra | No |
| Agile Enabler | Peñalbé Hernán | Sí |
| Product Builder | Buttini Cristobal | Sí |
| Product Builder | Cardozo Leandro | Sí |
| Product Builder | García Santiago | Sí |
| Product Builder | Sosa Ricardo | Sí |

## Punto de partida

- **Sprint 3 cerrado** el 26/10 (tag `sprint-3` sobre `23dcad8`): Sprint Goal cumplido, 29 de 29 Story Points comprometidos más TEC-03, y HU-32 y HU-34 adelantadas del Sprint 4. Ver las actas de [Review](../sprint-3/review.md), [Retrospectiva](../sprint-3/retrospectiva.md) y [Daily](../sprint-3/daily.md).
- **Velocidad:** 21 Story Points en el Sprint 1, 25 en el Sprint 2 y 34 en el Sprint 3.
- **Calidad en `main`:** 80.9% de cobertura total sin base y 94.8% con PostgreSQL; 346 escenarios BDD ejecutados.
- **Fechas:** con la replanificación del 24/10, la entrega es el 02/11/2026 y el desarrollo se cierra el 30/10. Quedan cuatro días de desarrollo (27/10 al 30/10).
- **Requerimientos mínimos cubiertos:** gestión de proyectos, Product Backlog, gestión de Sprints, estimación (Story Points y Planning Poker) y registro de esfuerzo. **Parciales:** gestión de defectos (falta cambiar el estado y consultar) y métricas (faltan horas, defectos y la consulta única). **Pendientes:** dashboard y reportes.

## Criterio de recorte (Review del Sprint 3)

Se mantiene el de la Review del Sprint 2: se recorta alcance, no pruebas ni documentación. El Sprint 4 toma primero lo que bloquea a otros (HU-30 → HU-47, HU-35 → HU-37, HU-38 y HU-39 → HU-40), y TEC-06, TEC-07 y TEC-08 no se recortan porque son entregables finales.

| Requerimiento mínimo de la consigna | Estado | Imprescindible en el Sprint 4 |
|---|---|---|
| Gestión de proyectos | ✅ HU-01 a HU-03 y HU-18 | — |
| Product Backlog | ✅ HU-04 a HU-09 | — |
| Gestión de Sprints | ✅ HU-10 a HU-13 | — (HU-14 en "si hay tiempo") |
| Estimación (Story Points + Planning Poker) | ✅ HU-20 a HU-23 y HU-25 | — (HU-24 en "si hay tiempo") |
| Registro de esfuerzo | ✅ HU-26 y HU-27 | — |
| Gestión de defectos | Parcial (HU-29) | HU-30 y HU-31 |
| Métricas | Parcial (TEC-03, HU-32 y HU-34) | HU-33, HU-47 y HU-35 |
| Dashboard | Pendiente | HU-36 y HU-37 |
| Reportes | Pendiente | HU-38, HU-39 y HU-40 |

## Ítems candidatos y dependencias

Las dependencias se verifican contra el código de `main` y los criterios de cada Issue; cada SDD las confirma.

| ID | Issue | Historia o tarea | Épica | Prioridad | Depende de |
|---|---|---|---|---|---|
| HU-30 | #48 | Cambiar el estado de un defecto | E8 | BAJA | HU-29 ✅ |
| HU-31 | #49 | Consultar los defectos de un proyecto o sprint | E8 | BAJA | HU-29 ✅; el filtro por estado usa los estados de HU-30 |
| HU-33 | #51 | Calcular horas estimadas, horas reales y desviación | E9 | BAJA | HU-07 ✅, HU-26 y HU-27 ✅, TEC-03 ✅ |
| HU-47 | #128 | Calcular los defectos detectados y resueltos | E9 | BAJA | HU-29 ✅ y HU-30 (defectos resueltos) |
| HU-35 | #53 | Consultar las métricas del proyecto | E9 | BAJA | HU-32 y HU-34 ✅, HU-33 y HU-47 |
| HU-36 | #54 | Ver el resumen del estado del proyecto | E10 | BAJA | E1 a E3 ✅ |
| HU-37 | #55 | Ver las métricas en el dashboard con gráficos | E10 | BAJA | HU-35 |
| HU-38 | #56 | Generar el reporte del proyecto | E11 | BAJA | HU-35, HU-27 ✅ y HU-31 |
| HU-39 | #57 | Generar el reporte de un sprint | E11 | BAJA | HU-35, HU-27 ✅ y HU-31 |
| HU-40 | #58 | Exportar un reporte a un archivo PDF | E11 | BAJA | HU-38 y la librería de PDF aprobada |
| TEC-06 | #69 | Documentación técnica | — | — | Código final |
| TEC-07 | #70 | Manual de usuario | — | — | Interfaz final |
| TEC-08 | #71 | Informe final de métricas y cobertura | — | — | E9 y la cobertura final |

## Estimación de las historias (Planning Poker)

Escala Fibonacci (1, 2, 3, 5, 8, 13, 21), con HU-03 (2 SP) como historia de referencia y las historias de los Sprints 1 a 3 como comparación (HU-32, métricas con paquete nuevo y endpoints, 3 SP; HU-34, una métrica sobre los mismos endpoints, 2 SP). Las tareas técnicas no se estiman en Story Points.

| ID | Historia | Story Points | Justificación |
|---|---|---|---|
| HU-30 | Cambiar el estado de un defecto | 3 | Como HU-08: transiciones válidas entre estados y sprint de resolución; el defecto ya existe (HU-29) |
| HU-31 | Consultar los defectos de un proyecto o sprint | 3 | Listado por proyecto o por sprint, con filtro por estado, como HU-27 |
| HU-33 | Calcular horas estimadas, horas reales y desviación | 3 | Extiende el paquete de métricas de HU-32 con horas estimadas, horas reales y desviación; reutiliza la consulta de esfuerzos de HU-27 |
| HU-47 | Calcular los defectos detectados y resueltos | 2 | Una métrica más sobre los mismos endpoints, como HU-34: cuenta los defectos detectados y los resueltos |
| HU-35 | Consultar las métricas del proyecto | 3 | Reúne en una sola consulta lo que calculan HU-32, HU-33, HU-34 y HU-47 y elimina la segunda lectura del backlog de HU-34; no agrega cálculos nuevos |
| HU-36 | Ver el resumen del estado del proyecto | 3 | Resumen del estado del proyecto a partir de los ítems y sprints existentes, con su sección en la interfaz |
| HU-37 | Ver las métricas en el dashboard con gráficos | 5 | Gráficos con las métricas de HU-35, dibujados sin librerías externas; es la historia con más interfaz del Sprint |
| HU-38 | Generar el reporte del proyecto | 5 | Reúne las métricas (HU-35), los esfuerzos (HU-27) y los defectos (HU-31) en un reporte del proyecto; paquete nuevo, endpoint y pantalla |
| HU-39 | Generar el reporte de un sprint | 3 | Mismo reporte que HU-38, acotado a un sprint; reutiliza su estructura |
| HU-40 | Exportar un reporte a un archivo PDF | 5 | Genera un PDF con el reporte de HU-38; requiere una librería externa y su integración |
| | **Total comprometido** | **35** | |

**Confirmación de HU-32 y HU-34 (Sprint 3):** el equipo confirma 3 SP para HU-32 y 2 SP para HU-34.

## Alcance comprometido

- **Comprometido:** HU-30 y HU-31 (E8), HU-33, HU-47 y HU-35 (E9), HU-36 y HU-37 (E10), HU-38, HU-39 y HU-40 (E11), más TEC-06, TEC-07 y TEC-08. Total: 35 Story Points más las tres tareas técnicas.
- **Si hay tiempo**, en este orden: HU-41, HU-24 y HU-14. Se asignan durante el Sprint a quien termine primero.
- **Fuera de la entrega:** E4 (HU-15 a HU-17), HU-19 y HU-28.
- **Acuerdo previo:** 35 Story Points superan la velocidad del Sprint 3 (34) y las tres tareas técnicas compiten por los mismos días. Si el Sprint se atrasa, se recorta alcance dentro de las historias antes que sacar una entera: menos gráficos en HU-37, un PDF sin formato en HU-40 y un reporte de sprint más simple en HU-39, sin discutirlo a mitad del Sprint. HU-36, HU-37, HU-38, HU-39 y HU-40 siguen siendo imprescindibles, y TEC-06, TEC-07 y TEC-08 no se recortan.

## Temas decididos

1. **Fechas:** el Sprint 4 va del 27/10 al 02/11/2026. El desarrollo se cierra el 30/10 y ese día se hacen la Review y la Retrospectiva (Issues #135 y #136). Del 31/10 al 02/11 quedan para cerrar la documentación (TEC-06 a TEC-08) y preparar la presentación y la demo final; la entrega es el 02/11. La Review registra el estado de TEC-06 a TEC-08 al 30/10, y su cierre queda en el Issue de cada tarea.
2. **División de historias pendientes (HU-31, HU-33 y HU-37):** no se dividen. Con cuatro días de desarrollo, cada una se hace entera, como HU-32 (HU-34 sí se dividió en HU-34 y HU-47, porque HU-47 depende de HU-30). Si HU-37 no entra completa, se recortan gráficos dentro de la historia (acuerdo previo). Pendiente: confirmar con los profesores, porque no asistieron al Planning.
3. **Librería de PDF para HU-40:** se propone `github.com/go-pdf/fpdf` (v0.9.0), porque la biblioteca estándar de Go no genera PDF y está escrita en Go puro (compila sin cgo, que varias máquinas del equipo no tienen). Antes de agregarla a `go.mod` se verifica la versión, la licencia y que siga mantenida; el equipo la aprueba y el Agile Enabler lo registra por escrito, como pide `AGENTS.md`. Hasta entonces HU-40 no empieza.
4. **Gráficos de HU-37:** se dibujan como SVG generado con JavaScript propio en `internal/web`, a partir del JSON de las métricas, sin librerías externas (`AGENTS.md` no las permite sin autorización).
5. **Integrantes en el reporte de HU-38:** el criterio CA-38.1 del Issue pide integrantes, pero E4 (HU-15 a HU-17) no está implementada y queda fuera de la entrega. Se ajusta el criterio: el reporte lista los integrantes que registraron esfuerzo en el proyecto (HU-26 y HU-27), y E4 sigue fuera de la entrega. El Issue #56 se actualiza con ese criterio antes de la SDD.

## Orden de trabajo

1. Sin dependencias pendientes, en paralelo: HU-30, HU-33 y HU-36.
2. Cuando entra HU-30: HU-31 y HU-47.
3. Cuando entran HU-33 y HU-47: HU-35.
4. Cuando entra HU-35: HU-37, HU-38 y HU-39.
5. Cuando entra HU-38: HU-40.
6. En paralelo durante todo el Sprint: TEC-06 y TEC-07; TEC-08 al final, con la cobertura de `main`.

## Definition of Ready y Definition of Done

Ver [`docs/definition-of-ready-done.md`](../../definition-of-ready-done.md). Se mantienen las del Sprint 3, más las acciones de su Retrospectiva:

- Quien aprueba un Pull Request revisa también el Issue (criterios tildados y "## Trazabilidad" de 4 líneas) y no aprueba hasta que esté (acción 1).
- Un test que se agrega después del GREEN se registra como desviación en la evidencia, con el motivo (acción 2).
- Antes de enviar la aprobación, se relee que el comentario sea del PR que se está aprobando (acción 3).
- La Daily se publica antes del mediodía y sin campos sin completar (acción 4).
- Antes de abrir el Pull Request se trae `main` a la rama, y después de cada merge se avisa en el grupo para que las ramas abiertas lo traigan (acción 5).

Sin otros cambios en este Planning.

## Responsables

Un responsable por historia o tarea, según los "Assignees" de cada Issue.

| Integrante | Ítems | Story Points | Motivo |
|---|---|---|---|
| Peñalbé Hernán | HU-47, HU-40, TEC-06, TEC-07 y TEC-08 | 7 (más las tres tareas técnicas, que no se estiman) | HU-47 depende de HU-30 y HU-40 espera a HU-38; como Agile Enabler ya aprueba las SDD y coordina, y las tres tareas técnicas se hacen en paralelo durante todo el Sprint |
| Buttini Cristobal | HU-36 y HU-37 | 8 | El dashboard (HU-36 y HU-37) es una cadena de interfaz; HU-36 no tiene dependencias y HU-37 arranca cuando entra HU-35 |
| Cardozo Leandro | HU-35 y HU-38 | 8 | HU-35 es el centro de la cadena (desbloquea HU-37, HU-38 y HU-39) y ya empezó su SDD, usando lo que calculan HU-32 y HU-34; HU-38 sigue a HU-35 y desbloquea HU-40 |
| García Santiago | HU-33 y HU-39 | 6 | HU-33 reutiliza la consulta de esfuerzos de HU-27, que hizo él; HU-39 reutiliza la estructura del reporte de HU-38 |
| Sosa Ricardo | HU-30 y HU-31 | 6 | Cadena de defectos (HU-30 y HU-31), que ya conoce por HU-29; HU-30 bloquea a HU-47, así que sale primero |

## Riesgos

- **Ritmo:** 35 Story Points superan la velocidad del Sprint 3 (34), en cuatro días de desarrollo y con las tres tareas técnicas en paralelo. El recorte ya está definido en el acuerdo previo.
- **Cadena larga en pocos días:** HU-33 y HU-47 → HU-35 → HU-37, HU-38 y HU-39 → HU-40. Si HU-35 se atrasa, se atrasan el dashboard y los reportes.
- **HU-30 bloquea a HU-47:** sin el cambio de estado no hay defectos resueltos (M10).
- **HU-40 necesita una librería externa:** el equipo tiene que aprobarla antes de empezar (tema 3).
- **Documentación final:** TEC-06, TEC-07 y TEC-08 compiten con el desarrollo en los mismos días; TEC-08 depende de la cobertura final de `main`.
- **Carga del Agile Enabler:** HU-47, HU-40 y las tres tareas técnicas, además de aprobar las SDD y coordinar.

## Próximo paso

- Cada responsable cierra los criterios de su Issue (sin "(borrador)") y escribe su SDD, empezando por HU-30, HU-33 y HU-36, que no tienen dependencias pendientes.
- Las historias y tareas comprometidas pasan al Sprint 4 del tablero.
- La Daily se registra en el Issue #134; la Review y la Retrospectiva del Sprint 4 tienen los Issues #135 y #136.
