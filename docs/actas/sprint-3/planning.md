# Acta de Planning — Sprint 3 (La Estimación)

> Issue del repositorio: #110

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 3 — La Estimación
- **Fecha del Planning:** 23/10/2026
- **Período del Sprint:** 23/10/2026 al 29/10/2026 (1 semana)
- **Sprint Goal:** Estimar ítems con Planning Poker y registrar el esfuerzo y los defectos de cada sprint, con las fórmulas de las métricas definidas para el Sprint 4.

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

- **Sprint 2 cerrado** (tag `sprint-2` sobre `5939950`): Sprint Goal cumplido, 25 de 25 Story Points (HU-07 a HU-13), más TEC-04 (interfaz web) y TEC-05 (PostgreSQL en Docker). Ver las actas de [Review](../sprint-2/review.md), [Retrospectiva](../sprint-2/retrospectiva.md) y [Daily](../sprint-2/daily.md).
- **Velocidad:** 21 Story Points en el Sprint 1 y 25 en el Sprint 2.
- **Pendiente hasta la entrega (05/11/2026):** 27 historias (HU-14 a HU-40) y 4 tareas técnicas (TEC-03, TEC-06, TEC-07 y TEC-08) en 2 sprints de una semana. A la velocidad actual no entra todo: se aplica el criterio de recorte de la Review del Sprint 2.
- **Calidad en `main`:** total 79.7% de cobertura (sin base) y 194 escenarios BDD ejecutados.

## Criterio de recorte (Review del Sprint 2)

Se recorta alcance, no pruebas ni documentación (decisión 8 del Sprint 1). Por cada requerimiento mínimo de la consigna que falta, se eligen las historias imprescindibles para cubrirlo; el resto queda en "si hay tiempo". TEC-06, TEC-07 y TEC-08 no se recortan porque son entregables finales.

| Requerimiento mínimo de la consigna | Estado | Imprescindible | Sprint |
|---|---|---|---|
| Gestión de proyectos | ✅ HU-01 a HU-03 | — | — |
| Product Backlog | ✅ HU-04 a HU-09 | — | — |
| Gestión de Sprints | ✅ HU-10 a HU-13 | — (HU-14 queda en "si hay tiempo") | — |
| Estimación (Story Points + Planning Poker) | Pendiente | HU-20, HU-21, HU-22, HU-23 y HU-25 | 3 |
| Registro de esfuerzo | Pendiente | HU-26 y HU-27 | 3 |
| Gestión de defectos | Pendiente | HU-29 (Sprint 3), HU-30 y HU-31 | 3 y 4 |
| Métricas | Pendiente | TEC-03 (Sprint 3), HU-32 a HU-35 | 3 y 4 |
| Dashboard | Pendiente | HU-36 y HU-37 | 4 |
| Reportes | Pendiente | HU-38 y HU-39 | 4 |

HU-40 (PDF) necesita la librería externa aprobada y queda para el final del Sprint 4. E4 (integrantes, HU-15 a HU-17), E5 (fechas, HU-18 y HU-19), HU-24 y HU-28 no son requerimientos mínimos de la consigna: se toman solo con capacidad sobrante, salvo HU-18, que se compromete como historia independiente (ver "Responsables").

## Historias más atómicas (pedido de los Product Architect)

Se aplica desde este Sprint, sin tocar las historias ya terminadas. La historia original conserva su ID e Issue con la primera parte; las partes nuevas toman HU-41 en adelante, en la misma épica, y se estiman en el Planning Poker.

| Historia actual | Se divide en | Sprint |
|---|---|---|
| HU-18 · Registrar fechas de inicio y fin del proyecto | HU-18 Fecha de inicio · HU-41 Fecha de fin (posterior a la de inicio) | 3 |
| HU-31 · Consultar los defectos de un proyecto o sprint | HU-31 De un proyecto · HU-42 De un sprint | 4 |
| HU-32 · Calcular SP planificados, completados y velocidad | HU-32 Planificados · HU-43 Completados · HU-44 Velocidad | 4 |
| HU-33 · Calcular horas estimadas, horas reales y desviación | HU-33 Estimadas · HU-45 Reales · HU-46 Desviación | 4 |
| HU-34 · Calcular % de historias completadas y defectos detectados y resueltos | HU-34 % de historias completadas · HU-47 Defectos detectados y resueltos | 4 |
| HU-16 · Modificar los datos o el rol de un integrante | HU-16 Modificar datos · HU-48 Cambiar rol | Si hay tiempo |
| HU-37 · Ver las métricas en el dashboard con gráficos | Una historia por gráfico, según TEC-03 (IDs a partir de HU-49) | 4 |

Pendiente: confirmar con los profesores en este Planning. Los Issues nuevos y los cambios de título se hacen con autorización del Agile Enabler.

## Estimación de las historias del Sprint 3 (Planning Poker)

Escala Fibonacci, con HU-03 (2 SP) como referencia y las historias de los Sprints 1 y 2 como comparación: HU-04 y HU-10, crear con paquete nuevo, 5 SP; HU-07, modificar, 3 SP; HU-13, cambio de estado, 3 SP. Las tareas técnicas no se estiman en Story Points.

Los Story Points de la tabla son los que el equipo aprobó en el Planning Poker.

| ID | Historia | Story Points | Justificación |
|---|---|---|---|
| HU-20 | Asignar story points a un ítem del backlog | 2 | El ítem ya tiene el campo `StoryPoints` (HU-04); valida la escala Fibonacci y guarda con el `Update` del backlog |
| HU-21 | Abrir una sesión de Planning Poker | 5 | Paquete nuevo, como HU-10: entidad, repositorio y `Manager`; la sesión pertenece a un ítem existente |
| HU-22 | Votar de forma oculta | 3 | Un voto por integrante y por ronda; los votos no se muestran hasta revelarlos |
| HU-23 | Revelar las estimaciones | 2 | Cambia el estado de la sesión y devuelve todos los votos a la vez |
| HU-25 | Registrar la estimación acordada | 2 | Guarda los Story Points en el ítem (HU-20) y cierra la sesión |
| HU-26 | Registrar esfuerzo realizado en una historia | 5 | Paquete nuevo; la historia tiene que pertenecer al sprint (HU-11) |
| HU-27 | Consultar los registros de esfuerzo | 3 | Listado por historia o por sprint, como HU-06 |
| HU-29 | Registrar un defecto | 5 | Paquete nuevo; el defecto se asocia a una historia del proyecto |
| HU-18 | Registrar la fecha de inicio del proyecto | 2 | Campo nuevo en el proyecto y su validación |
| | **Total comprometido** | **29** | |
| HU-41 (si hay tiempo) | Registrar la fecha de fin del proyecto | 2 | Posterior a la de inicio (HU-18) |
| HU-24 (si hay tiempo) | Detectar diferencias y repetir ronda | 3 | Nueva ronda en la misma sesión (HU-22 y HU-23) |
| HU-14 (si hay tiempo) | Consultar sprints anteriores | 3 | Listado de los sprints de un proyecto, que hoy no existe |
| HU-30 (si hay tiempo) | Cambiar el estado de un defecto | 3 | Como HU-08, con transiciones y sprint de resolución |

## Alcance comprometido

- **Comprometido:** TEC-03 y HU-20, HU-21, HU-22, HU-23, HU-25 (E6), HU-26 y HU-27 (E7), HU-29 (E8) y HU-18 (E5). Total: 29 Story Points más TEC-03.
- **Si hay tiempo**, en este orden: HU-41, HU-24, HU-14 y HU-30. Se asignan durante el Sprint a quien termine primero.
- **Al Sprint 4:** HU-30 y HU-31 (si no entran), E9 (HU-32 a HU-35, divididas), E10 (HU-36 y HU-37), E11 (HU-38 y HU-39; HU-40 al final), TEC-06, TEC-07 y TEC-08, más un Issue de presentación y demo final.
- **Fuera de la entrega, salvo capacidad sobrante:** E4 (HU-15 a HU-17), HU-19 y HU-28.
- **Acuerdo previo:** 29 Story Points superan la velocidad del Sprint 2 (25). Si el Sprint se atrasa, **HU-18 y HU-29 son las primeras que pasan a "si hay tiempo"** (HU-29 vuelve al Sprint 4 con el resto de E8), sin discutirlo a mitad del Sprint.

## Dependencias

| Ítem | Depende de | Bloquea a |
|---|---|---|
| TEC-03 | — | E9 (HU-32 a HU-35), HU-37 y TEC-08 en el Sprint 4 |
| HU-20 | HU-04 y HU-07 (`Update` del backlog) | HU-25 |
| HU-21 | HU-04 y HU-05 (ítem existente) | HU-22, HU-23, HU-24 y HU-25 |
| HU-22 | HU-21 | HU-23 |
| HU-23 | HU-22 | HU-24 y HU-25 |
| HU-25 | HU-20 y HU-23 | — |
| HU-26 | HU-11 (historia asignada al sprint) | HU-27, HU-28 y HU-33 (Sprint 4) |
| HU-27 | HU-26 | — |
| HU-29 | HU-04 (historia del proyecto) | HU-30, HU-31 y HU-34 (Sprint 4) |
| HU-18 | HU-01 | HU-41 y HU-19 |

## Temas decididos

1. **Fechas:** Sprint 3 del 23/10 al 29/10 y Sprint 4 del 30/10 al 05/11, día de la entrega.
2. **Recorte de alcance** según el criterio de la Review del Sprint 2 (ver arriba).
3. **Historias más atómicas** desde este Sprint (ver arriba), a confirmar con los profesores.
4. **TEC-03 primero:** las fórmulas de las métricas se cierran en este Sprint, porque E9, el dashboard y TEC-08 dependen de ellas.
5. **Interfaz (TEC-04):** cada historia nueva suma su endpoint bajo `/api/…` y su pantalla en la interfaz web, con los mismos criterios de tests de TEC-04 (un caso exitoso y un test por código de error).
6. **Acciones de la Retrospectiva del Sprint 2**, todas con trabajo en este Sprint (ver "Definition of Done").
7. **Mejoras al flujo de trabajo:**
   - Un comando de "ponerse al día" para los agentes de IA (traer `main`, leer las reglas, ver PR e Issues abiertos y frenar antes de escribir), para evitar trabajar sobre una copia vieja, como pasó con la SDD de HU-12.
   - El criterio de aceptación en el nombre de cada escenario BDD, por ejemplo "Escenario: … (CA-20.3)", para ver la trazabilidad en la salida de godog.
   - Un documento único de decisiones técnicas con su "por qué" (`docs/decisiones-tecnicas.md`), dentro de TEC-06.
8. **Mejoras de los Issues** al refinarlos para este Sprint: línea "Origen: Consigna — requerimiento N", sección "Dependencias", criterios en lenguaje de negocio (los errores Go y `-race` quedan en la SDD), Story Points visibles en el Issue y responsable asignado al entrar al Sprint.

## Definition of Ready

La del Sprint 2, sin cambios: Issue con criterios definitivos (sin "(borrador)"), SDD aprobada, Story Points estimados, dependencias integradas en `main` y un único responsable. Todas las historias pendientes tienen hoy sus criterios en "(borrador)": cada responsable los cierra antes de crear la rama.

## Definition of Done del Sprint 3

La del Sprint 2, más las acciones de su Retrospectiva:

- En el RED, si el paquete no compila por la funcionalidad nueva, se agregan primero las firmas vacías (stubs), para que los errores de los propios tests aparezcan en el RED (acción 1).
- Toda aprobación de un Pull Request dice qué se ejecutó y qué se revisó; quien ayuda a revisar también deja su comentario (acción 2).
- Antes de cerrar una historia, el Issue tiene los criterios tildados y la "## Trazabilidad" con el formato de `AGENTS.md` (acción 3).
- La Daily se publica antes del mediodía y sin campos de la plantilla sin completar (acción 4).
- Aprobado en este Planning: línea "Aprobada por" en cada SDD y `go vet ./...` antes del commit RED.

## Responsables

Un responsable por historia o tarea, con las cadenas en una sola persona para que haya menos traspasos, y las historias que bloquean a otras para quien tiene más capacidad según las horas de las Dailys del Sprint 2.

| Integrante | Ítems | Story Points | Motivo |
|---|---|---|---|
| Peñalbé Hernán | TEC-03, acciones de la Retrospectiva y división de historias | — | TEC-03 bloquea al Sprint 4; como Agile Enabler además aprueba las SDD y coordina |
| Buttini Cristobal | HU-21 y HU-22 | 8 | Inicio de la cadena de Planning Poker, como hizo con E3 (HU-10 y HU-11) |
| Cardozo Leandro | HU-20, HU-23 y HU-25 | 6 | HU-20 extiende el backlog, que conoce por HU-09 y TEC-05; HU-23 y HU-25 cierran la cadena de Planning Poker |
| García Santiago | HU-26 y HU-27 | 8 | Cadena de esfuerzo, independiente de las demás |
| Sosa Ricardo | HU-18 y HU-29 (y HU-41 si hay tiempo) | 7 | HU-18 es independiente y no bloquea a nadie en este Sprint (acuerdo del Sprint 2). HU-29 no depende de nada nuevo, pero bloquea a HU-30, HU-31 y HU-34 del Sprint 4: si se atrasa, E8 se corre entera al Sprint 4. Con las dos queda cerca de lo que tuvo en el Sprint 2 (HU-07 y HU-08, 6 SP) y son las dos primeras que pasan a "si hay tiempo" si el Sprint se atrasa |

## Riesgos

- **Ritmo:** con 27 historias pendientes y una velocidad de 25 SP por sprint, no entra todo el Product Backlog. El recorte ya está definido y se revisa en la Review del Sprint 3.
- **Cadena de Planning Poker:** HU-21 → HU-22 → HU-23 → HU-25. Si HU-21 se atrasa, se atrasa la épica; mergear HU-21 lo antes posible es la prioridad del Sprint.
- **TEC-03 tarde:** si no se cierra en este Sprint, el Sprint 4 arranca bloqueado en métricas, dashboard e informe final.
- **Sprint 4 cargado:** métricas, dashboard, reportes, la documentación final y la demo, en la última semana. Por eso el Sprint 3 no toma nada de E9 a E11.
- **Carga del Agile Enabler:** TEC-03, la división de historias, las acciones de la Retro y la aprobación de 9 SDD; las SDD llegan escalonadas.

## Próximo paso

- TEC-03 arranca el primer día. Cada responsable cierra los criterios de su Issue (sin "(borrador)") y escribe su SDD, empezando por HU-21, HU-20 y HU-26, que abren sus cadenas.
- Las historias y tareas comprometidas pasan al Sprint 3 del tablero, con su responsable en "Assignees".
- La Daily se registra en el Issue #111; la Review y la Retrospectiva del Sprint 3 son el 29/10/2026 (Issues #112 y #113).
