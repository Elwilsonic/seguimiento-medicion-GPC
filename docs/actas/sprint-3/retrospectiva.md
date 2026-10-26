# Acta de Retrospectiva — Sprint 3 (La Estimación)

> Issue del repositorio: #113

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 3 — La Estimación (23/10/2026 al 26/10/2026)
- **Fecha de la Retrospectiva:** 26/10/2026
- **Facilitador:** Peñalbé Hernán (Agile Enabler)

## Participantes

| Rol | Integrante | Asistió |
|---|---|---|
| Agile Enabler | Peñalbé Hernán | Sí |
| Product Builder | Buttini Cristobal | Sí |
| Product Builder | Cardozo Leandro | Sí |
| Product Builder | García Santiago | Sí |
| Product Builder | Sosa Ricardo | Sí |

## Acciones de la Retrospectiva del Sprint 2

| Acción | ¿Se cumplió? | Evidencia |
|---|---|---|
| 1. En el RED, si el paquete no compila, agregar primero los stubs para que compile y fallen las aserciones | Parcial | HU-18, HU-20, HU-23, HU-25, HU-27, HU-32 y HU-34 no registran desviaciones. En HU-26 una aserción quedó tapada por el stub y el error del test apareció recién en el GREEN (corregido en `bed08f5`); en HU-21 el RED ya traía el wiring de PostgreSQL, y en HU-29 la persistencia entró sin un RED propio |
| 2. Toda aprobación dice qué se ejecutó y qué se revisó; quien ayuda a revisar también comenta | Parcial | En 15 de los 16 Pull Requests del Sprint el aprobador dejó por escrito qué revisó, en la revisión o en un comentario (#114, #116 a #125, #127 y #129 a #131). En #126 (HU-27) el texto de la aprobación de Cardozo habla de la fila de HU-22 y del commit `4450b4d`, que son del PR #125; no queda escrito qué se revisó de HU-27. En #127 revisaron Hernán y Sosa, y los dos dejaron su comentario |
| 3. Antes de cerrar una historia, criterios tildados y "## Trazabilidad" con el formato de `AGENTS.md` | Parcial | Cumplen #33, #38, #39, #40, #41, #44, #47 y #50. #41 se corrigió antes del merge, por la revisión. Quedaron sin tildar los criterios de #36 (HU-18) y #52 (HU-34), y la Trazabilidad de #45 (HU-27) no sigue el formato; la de #43 (HU-25) tampoco lo seguía al cierre, y ya se llevó al formato |
| 4. La Daily se publica antes del mediodía y sin campos sin completar | Parcial | Issue #111: antes de las 12:00 el 25/10 (11:58) y el 26/10 (11:33); después, el 23/10 (19:33, día del Planning) y el 24/10 (12:57). En la del 24/10 quedaron dos horas en `⟨completar⟩` y un ítem vacío en las horas de Sosa |
| 5. Decidir en el Planning el recorte de alcance para llegar a la entrega | Sí | Acta de Planning del Sprint 3: criterio de recorte por requerimiento mínimo y "si hay tiempo" |
| 6. Corregir los detalles de los Issues cerrados del Sprint 2 | Sí | #25, #86, #22 y #35 corregidos el 23/10 (Daily del 24/10) |

Del Planning del Sprint 3 también se cumplieron: la línea "Aprobada por" en las once SDD, el criterio de aceptación en el nombre de cada escenario BDD (los 124 escenarios y esquemas de las once historias lo tienen) y el checklist para ponerse al día (`docs/antes-de-empezar.md`, PR #119). Queda pendiente `docs/decisiones-tecnicas.md`, que va en TEC-06.

## ¿Qué salió bien?

- Se cumplió el Sprint Goal con 29 de 29 Story Points en cuatro días, más TEC-03, y se adelantaron HU-32 y HU-34 del Sprint 4: la velocidad subió de 25 a 34.
- La cadena de Planning Poker (HU-21 → HU-22 → HU-23 → HU-25) se integró completa entre el 24/10 y el 26/10, con dos responsables.
- TEC-03 se cerró el primer día, como pedía el Planning, y las dos primeras métricas ya están en `main`.
- Hubo merges todos los días, del 23/10 al 26/10.
- Todos los Pull Requests tuvieron la aprobación de un integrante distinto del autor, CI en verde y merge con "Create a merge commit". Las revisiones encontraron problemas reales antes del merge (tests faltantes de 409 y 500 en HU-22, el wiring adelantado de HU-21, el formato del Issue de HU-23).
- Los tests con PostgreSQL se corrieron en la revisión de varias historias (HU-21, HU-22, HU-26, HU-29), y la cobertura total con base llegó a 94.8%.
- Los conflictos entre ramas que tocaban los mismos archivos (README, cobertura, registro de IA, `server.go`) se resolvieron con merge de `main`, sin rebase y sin perder cambios.
- Los 75 commits del Sprint (sin contar merges) tienen como autor a un integrante y ninguno menciona a la IA.

## ¿Qué se puede mejorar?

1. **Cierre de los Issues.** En cuatro de las once historias quedó algo del Issue sin hacer al cerrar (criterios sin tildar o Trazabilidad con otro formato), aunque es parte de la DoD desde el Sprint 2.
2. **Tests después del GREEN.** En HU-22 la revisión pidió tests que no estaban en el RED (409, falla de guardado) y se agregaron después del GREEN sin registrarlo como desviación en la evidencia.
3. **Texto de la aprobación.** En el PR #126 el comentario de aprobación es de otra historia.
4. **Daily.** Dos de cuatro se publicaron después del mediodía y una quedó con campos sin completar.
5. **Conflictos por archivos compartidos.** Casi todas las historias tocan `README.md`, `docs/reportes/cobertura.md`, `docs/ia/registro.md` e `internal/httpapi/server.go`; con varias ramas abiertas a la vez, cada merge obligó a resolver conflictos en las demás.
6. **Registro de IA incompleto.** `docs/ia/registro.md` no tiene filas de HU-22, HU-26, HU-27 ni HU-32.

## Incidencias del Sprint registradas

- Al publicar Pull Requests y comentarios desde el agente de IA, GitHub agrega un pie "Generated by Claude Code"; se quitó en cada caso y se volvió a verificar.
- El push del tag `sprint-2` desde la sesión con IA dio 403; lo subió Peñalbé Hernán desde su máquina.

## Acciones para el Sprint 4

El equipo aprobó las seis acciones, sin ajustes. Cada responsable las cumple durante el Sprint 4.

| # | Acción | Responsable |
|---|---|---|
| 1 | Quien aprueba un Pull Request revisa también el Issue (criterios tildados y "## Trazabilidad" de 4 líneas) y no aprueba hasta que esté | Revisor de cada Pull Request |
| 2 | Un test que se agrega después del GREEN se registra como desviación en la evidencia, con el motivo | Responsable de cada historia |
| 3 | Antes de enviar la aprobación, releer que el comentario sea del PR que se está aprobando | Todo el equipo |
| 4 | La Daily se publica antes del mediodía y sin campos sin completar | Todo el equipo (cada integrante, su parte) |
| 5 | Antes de abrir el Pull Request, traer `main` a la rama; y después de cada merge, avisar en el grupo para que las ramas abiertas lo traigan | Todo el equipo |
| 6 | Completar los pendientes menores de esta acta | Responsable de cada ítem |

## Pendientes menores

- Tildar los criterios de #36 (HU-18, Sosa Ricardo) y #52 (HU-34, Peñalbé Hernán).
- Llevar la "## Trazabilidad" de #45 (HU-27, García Santiago) al formato de `AGENTS.md`. La de #43 (HU-25, Cardozo Leandro) ya está en el formato.
- Agregar las filas de HU-22, HU-26, HU-27 y HU-32 a `docs/ia/registro.md`.
- Sigue el pendiente de los Sprints 1 y 2 del singular: en `internal/project/features_test.go:209`, el mensaje de error de un paso BDD diría "hay 1 proyectos guardados" si quedara un solo proyecto.
