# Registro de uso de Inteligencia Artificial

La consigna permite usar herramientas de IA, siempre que todo resultado sea comprendido, revisado y validado por el equipo. Este registro documenta cada uso: para qué se usó, qué se generó y qué revisó o cambió el equipo.

- **Herramienta principal:** Claude Code (Anthropic), guiado por las reglas de [`AGENTS.md`](../../AGENTS.md).
- **Control:** la IA no hace commits, push ni cambios en GitHub sin autorización explícita del equipo (ver `AGENTS.md`).

## Sprint 0

| Fecha | Responsable | Uso | Qué generó la IA | Revisión del equipo |
|---|---|---|---|---|
| 06/10/2026 | Peñalbé Hernán | Análisis de requisitos | Resumen del enunciado del TP y comparación con `AGENTS.md` | Se validó contra el PDF de la consigna |
| 06/10/2026 | Peñalbé Hernán | Revisión del Product Backlog | Revisión de épicas e historias cargadas (errores, huecos respecto del PDF, ambigüedades) | Se corrigieron épicas e Issues; se decidió la opción de horas estimadas por ítem |
| 06/10/2026 | Peñalbé Hernán | Product Backlog | Estructura de épicas, lista HU-01 a HU-40 y TEC-01 a TEC-05 con sus textos y criterios en borrador; propuesta de estados | El equipo revisó, ajustó y validó; cada integrante cargó los Issues de su épica |
| 07/10/2026 | Peñalbé Hernán | Revisión de Issues | Verificación de textos, etiquetas y vínculos con épicas; textos corregidos para HU-38, HU-39, HU-40, HU-07 y HU-20 | El equipo aplicó las correcciones a mano en GitHub |
| 07/10/2026 | Peñalbé Hernán | Tareas técnicas | Redacción de TEC-05 y sus criterios; propuesta de pasar TEC-02 al Sprint 1 | El equipo decidió usar PostgreSQL en Docker y mover TEC-02 al Sprint 1, y aprobó los textos |
| 08/10/2026 | Peñalbé Hernán | Estimación | Recomendación de escalas (Fibonacci para historias, talles para épicas) | Los talles de las épicas los votó el equipo en Planning Poker |
| 08/10/2026 | Peñalbé Hernán | Documentación | Textos de Issues (archivo de carga de HU-09 a HU-40 y TEC-01 a TEC-04, épica #13, HU-40 y correcciones de HU-07, HU-20, HU-29, HU-38 y HU-39); borradores del acta de cierre del Sprint 0, del README y de este registro | El equipo revisó y cargó los Issues, modificó la retrospectiva, agregó las decisiones 6 a 8 y ajustó el README y este registro |
| 08/10/2026 | Peñalbé Hernán | Proceso de trabajo | Revisión y corrección de `AGENTS.md` (referencias a Issues, orden de cierre, autorización de trazabilidad, BDD automatizado con godog), convención de commits, `.gitignore` y `.gitattributes` | El equipo definió las reglas (ramas, push con autorización, forma del merge, uso de godog, formato de commits; `.gitattributes` porque todo el equipo usa Windows) y aprobó la versión final |

## Sprint 1

| Fecha | Responsable | Uso | Qué generó la IA | Revisión del equipo |
|---|---|---|---|---|
| 08/10/2026 | Peñalbé Hernán | Tablero Scrum | Lista de ítems del Sprint 1 e instrucciones para configurar las vistas del tablero (campo Sprint, filtro de la vista Sprint Backlog, jerarquía) | El equipo configuró el tablero y asignó los ítems al Sprint 1 |
| 08/10/2026 | Peñalbé Hernán | Planning del Sprint 1 | Revisión del acta de Planning (Story Points pendientes, dependencias entre sprints, instalación de godog, registro de la Daily, carga del Sprint 2); opinión sobre la estimación de HU-01 | El equipo redactó el acta, votó los Story Points (HU-01 = 8, total 21) y definió el Sprint Goal |
| 08/10/2026 | Peñalbé Hernán | Plan hasta la entrega | Propuesta de reparto de las historias de los Sprints 3 y 4 según las dependencias entre épicas (métricas después de Story Points, esfuerzo y defectos; dashboard y reportes después de métricas) | El equipo aceptó el reparto y agregó el riesgo de carga del Sprint 4 |
| 08/10/2026 | Peñalbé Hernán | Responsables | Análisis de la carga por integrante según el reparto propuesto | El equipo definió los responsables, de modo que cada integrante tenga al menos una historia |
| 08/10/2026 | Peñalbé Hernán | Daily | Recomendación de registrar la Daily en un archivo versionado y plantilla `docs/actas/sprint-1-daily.md` | El equipo decidió que el Agile Enabler consolida la Daily y la sube con un único Pull Request al cierre del Sprint |
| 08/10/2026 | Peñalbé Hernán | Documentación | Actualización del README con el estado del Sprint 1, el Sprint Goal y los Story Points comprometidos | El equipo revisó y aprobó el README |

## Incidentes

| Fecha | Qué pasó | Cómo se resolvió |
|---|---|---|
| 06/10/2026 | La IA creó dos Issues (HU-07 #19 y HU-08 #20) sin autorización explícita del equipo | Se verificó que coincidían con el texto del equipo y se conservaron. Desde entonces, toda acción en GitHub requiere autorización (regla incorporada en `AGENTS.md`) |

## Cómo registrar nuevos usos

Agregar una fila por cada uso relevante (una especificación, un conjunto de tests, una revisión), indicando la fecha, quién la pidió, qué generó la IA y qué revisó o cambió el equipo. En las historias de usuario, la evidencia TDD (`docs/evidencias/HU-XX-tdd.md`) complementa este registro.
