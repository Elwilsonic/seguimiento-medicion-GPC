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
| 08/10/2026 | Peñalbé Hernán | Proceso de trabajo | Revisión y corrección de `AGENTS.md` (referencias a Issues, orden de cierre, autorización de trazabilidad, BDD automatizado con godog) | El equipo definió las reglas (ramas, push con autorización, forma del merge, uso de godog) y aprobó la versión final |

## Incidentes

| Fecha | Qué pasó | Cómo se resolvió |
|---|---|---|
| 06/10/2026 | La IA creó dos Issues (HU-07 #19 y HU-08 #20) sin autorización explícita del equipo | Se verificó que coincidían con el texto del equipo y se conservaron. Desde entonces, toda acción en GitHub requiere autorización (regla incorporada en `AGENTS.md`) |

## Cómo registrar nuevos usos

Agregar una fila por cada uso relevante (una especificación, un conjunto de tests, una revisión), indicando la fecha, quién la pidió, qué generó la IA y qué revisó o cambió el equipo. En las historias de usuario, la evidencia TDD (`docs/evidencias/HU-XX-tdd.md`) complementa este registro.
