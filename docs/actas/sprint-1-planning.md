# Acta de Planning — Sprint 1 (El MVP)

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 1 — El MVP
- **Fecha del Planning:** 09/10/2026
- **Período del Sprint:** 09/10/2026 al 15/10/2026 (1 semana, según la iteración "Sprint 1" del tablero)
- **Sprint Goal:** Poder crear, modificar y consultar proyectos y cargar y listar su Product Backlog desde una aplicación ejecutable, con cada cambio validado por integración continua.

## Participantes

| Rol | Integrante |
|---|---|
| Product Architect | Profesores de la cátedra (no participaron del Planning; no respondieron la consulta previa) |
| Agile Enabler | Peñalbé Hernán |
| Product Builder | Buttini Cristobal |
| Product Builder | Cardozo Leandro |
| Product Builder | García Santiago |
| Product Builder | Sosa Ricardo |

## Decisiones del equipo

Los profesores no respondieron las consultas previas al Planning, así que el equipo tomó estas decisiones por su cuenta. Si los profesores hacen observaciones, el equipo revisa el alcance en el próximo Planning y actualiza el backlog y la documentación que corresponda. La Retrospectiva se usa para mejorar el proceso, no para cambiar el alcance.

1. **Product Backlog detallado.** Las 11 funcionalidades mínimas de la consigna son las **épicas** (E1 a E11). Cada una se parte en historias chicas que entran completas en un Sprint. En total hay 40 historias (HU-01 a HU-40) y 5 tareas técnicas (TEC-01 a TEC-05).
2. **Dos identificadores distintos, que se usan juntos.** `HU-XX` es el ID de la historia: viene del Product Backlog, figura en el título de la tarjeta y se usa en archivos, tests, evidencias y commits. `#N` es el número del Issue en GitHub y depende del orden de creación (por ejemplo, HU-01 es el Issue #14). Los commits llevan ambos: el título usa el formato `tipo(HU-XX): descripción` y el cuerpo cita el Issue con `Refs #N` (y `Closes #N` en el commit que cierra la historia). Ejemplo: `test(HU-01): agregar tests (RED)` con `Refs #14`.
3. **Ciclo completo para todas las historias.** Todas las historias siguen la cadena Historia → SDD → BDD → TDD que define el `AGENTS.md`. Al ser historias chicas, cada ciclo es corto.
4. **Épicas como Issues con sub-issues.** Las historias están vinculadas a su épica como sub-issues, con etiquetas `epica:*`, `prioridad:*` y `tipo:*`.
5. **Tareas técnicas sin ciclo completo.** TEC-01 a TEC-05 no son historias de usuario: se definen con alcance y evidencia propias, sin SDD ni BDD formales (regla ya prevista en el `AGENTS.md`). Las tareas técnicas no se estiman en Story Points, porque no suman a la velocidad del equipo.
6. **TEC-01 con servidor HTTP** usando solo la librería estándar de Go (`net/http`), sin frameworks. En el Sprint 1 los datos se guardan en memoria, detrás de una interfaz de repositorio.
7. **Inicio del Sprint 1 el 09/10/2026.** El Sprint 0 se cerró el 08/10/2026 con su acta.
8. **Se completan las 40 historias antes del 05/11/2026**, fecha interna prevista de entrega (la consigna no fija un día exacto). No se deja ninguna épica afuera: si el ritmo no alcanza, lo atrasado se arrastra al sprint siguiente y se recorta alcance solo con acuerdo del equipo.
9. **Métricas (E9).** TEC-03 define las fórmulas de todas las métricas mínimas que exige la consigna: Story Points planificados y completados, velocidad, horas estimadas y reales, desviación entre esfuerzo estimado y real, porcentaje de historias completadas, y defectos detectados y resueltos. Se hace al inicio del Sprint 3, antes de implementar HU-32 a HU-35.
10. **Persistencia con PostgreSQL en Docker, a partir del Sprint 2.** En el Sprint 1 el almacenamiento es en memoria, detrás de una interfaz de repositorio. En el Sprint 2 se agrega TEC-05, que implementa los repositorios sobre PostgreSQL ejecutado en Docker, sin cambiar las reglas de negocio ni los tests existentes. Las pruebas unitarias siguen corriendo sin base y la base se prueba con pruebas de integración.
11. **Librerías externas aprobadas.** Solo godog (automatización de los escenarios BDD, únicamente en tests), el driver de PostgreSQL (TEC-05) y la librería de PDF (HU-40, que se documenta en el acta del Sprint 4). No se usan otras librerías externas sin acuerdo del equipo.
12. **TEC-02 (CI) entra en el Sprint 1.** Es la primera vez que entra código a `main`, así que conviene controlarlo desde los primeros Pull Requests. Es una tarea chica (un workflow de GitHub Actions) y se hace al principio del sprint, antes de integrar las historias.
13. **Convención de commits y flujo de integración.** El título de cada commit usa `tipo(ID): descripción` (tipos `test`, `feat`, `refactor`, `docs`, `ci`, `chore`, `fix`). Cada historia o tarea se trabaja en su propia rama `feature/...` y se integra a `main` por Pull Request con "Create a merge commit" (nunca squash ni rebase), para conservar el historial RED → GREEN → REFACTOR. Los commits y el push requieren autorización explícita del responsable.
14. **godog se incorpora con HU-01.** HU-01 es la primera historia con escenarios BDD, así que agrega godog al `go.mod` junto con su primer `.feature` y el runner de los escenarios.
15. **Daily registrada en el repositorio.** Cada Daily se registra en `docs/actas/sprint-1-daily.md`, con una fila por integrante y por día (qué hizo, qué va a hacer y bloqueos), para que quede evidencia versionada de la ceremonia. Para evitar conflictos de edición y no trabajar sobre `main`, el Agile Enabler consolida el registro del Sprint y lo sube con un único Pull Request (rama `docs/S1-daily`) al cierre del Sprint, con "Create a merge commit".

## Estimación de las historias (Planning Poker)

Las historias se estiman en Story Points con la escala Fibonacci (1, 2, 3, 5, 8, 13, 21), tomando HU-03 como historia de referencia. Las tareas técnicas no se estiman en puntos.

| ID | Issue | Historia | Story Points | Justificación |
|---|---|---|---|---|
| HU-03 | #16 | Consultar estado del proyecto | 2 | Solo lectura; historia de referencia para comparar las demás |
| HU-02 | #15 | Modificar datos del proyecto | 3 | Update con validaciones y cambio de estado; mismo patrón que HU-01 |
| HU-06 | #21 | Listar el backlog de un proyecto | 3 | Lista simple; decide si valida el proyecto y el caso de backlog vacío |
| HU-04 | #17 | Crear ítem del backlog | 5 | Más campos y validaciones; relación con el proyecto y estados del ítem |
| HU-01 | #14 | Crear proyecto | 8 | Además de la funcionalidad, define la organización de paquetes, la interfaz de repositorio y monta godog por primera vez |
| | | **Total comprometido** | **21** | |

## Alcance comprometido

| Orden | ID | Issue | Historia o tarea | Épica | Prioridad |
|---|---|---|---|---|---|
| 1 | TEC-02 | #32 | Integración continua con GitHub Actions (en cada Pull Request) | Técnica (sin épica) | ALTA |
| 2 | HU-01 | #14 | Crear proyecto | E1 · Gestión de proyectos | ALTA |
| 3 | HU-02 | #15 | Modificar datos del proyecto (nombre, descripción y estado) | E1 · Gestión de proyectos | ALTA |
| 4 | HU-03 | #16 | Consultar estado del proyecto | E1 · Gestión de proyectos | ALTA |
| 5 | HU-04 | #17 | Crear ítem del backlog (con validación de campos obligatorios) | E2 · Product Backlog | ALTA |
| 6 | HU-06 | #21 | Listar el backlog de un proyecto | E2 · Product Backlog | ALTA |
| 7 | TEC-01 | #23 | Punto de entrada ejecutable mínimo (servidor HTTP con la librería estándar `net/http`) | Técnica | ALTA |

**Si el equipo llega con tiempo** (se suman durante el Sprint, no están comprometidas): HU-05 (#18, consultar un ítem), HU-07 (#19, modificar un ítem) y HU-08 (#20, cambiar el estado de un ítem).

**Fuera del Sprint 1:** HU-09 y el resto de las épicas. Lo que no entre de HU-05, HU-07 y HU-08 se arrastra al Sprint 2.

**Plan hasta la entrega.** La fecha interna prevista de entrega es el 05/11/2026, así que quedan cuatro sprints de una semana y el equipo se propone completar las 40 historias. El reparto sigue las dependencias entre épicas: las métricas (E9) necesitan Story Points, esfuerzo y defectos ya registrados, y el dashboard (E10) y los reportes (E11) necesitan las métricas. Reparto previsto, a confirmar en cada Planning:

| Sprint | Fechas | Alcance previsto |
|---|---|---|
| 1 | 09/10 al 15/10 | Comprometido: TEC-02, HU-01, HU-02, HU-03, HU-04, HU-06 y TEC-01. Opcional: HU-05, HU-07 y HU-08 |
| 2 · La Interfaz | 16/10 al 22/10 | HU-05, HU-07 y HU-08 (si no entraron en el Sprint 1), HU-09 a HU-19 (backlog, sprints, integrantes y fechas), TEC-04 (interfaz web usable) y TEC-05 (persistencia con PostgreSQL en Docker) |
| 3 · Funcionalidad y Calidad | 23/10 al 29/10 | TEC-03 (fórmulas, al inicio), HU-20 (asignar Story Points), HU-26 (registrar esfuerzo), HU-29 a HU-31 (defectos), HU-32 a HU-35 (métricas) y HU-36 y HU-37 (dashboard con gráficos) |
| 4 · Cierre y Entrega Final | 30/10 al 05/11 | HU-21 a HU-25 (Planning Poker), HU-27 y HU-28 (consultar y corregir esfuerzo), HU-38 a HU-40 (reportes y exportación a PDF), documentación final y entrega |

Los objetivos de los Sprints 2 a 4 siguen los de la consigna: interfaz usable en el Sprint 2, visualización en el Sprint 3 y exportación del informe a PDF en el Sprint 4. El reparto es una propuesta y se confirma en el Planning de cada sprint.

Ceremonias: la consigna pide Planning, Daily, Review y Retrospective ya desde el Sprint 1. La Daily se registra en `docs/actas/sprint-1-daily.md`, consolidada por el Agile Enabler (decisión 15). Al cierre se suben las actas `docs/actas/sprint-1-review.md` y `docs/actas/sprint-1-retrospectiva.md`.

## Orden de trabajo

1. TEC-02 (CI) al principio: tiene que estar integrada antes de que entren los primeros Pull Requests de código, para que todos se controlen.
2. HU-01 primero entre las historias, porque define la organización de paquetes y la interfaz de repositorio que usan las demás (requisito de TEC-05), y agrega godog al `go.mod` junto con el primer escenario BDD.
3. Después HU-02 y HU-03 (E1), y HU-04 (E2).
4. HU-06 después de HU-04 y HU-01, porque lista ítems que pertenecen a un proyecto.
5. TEC-01 al final: arranca cuando HU-01 esté integrada y se amplía con lo que haya en `main`, porque necesita historias integradas para exponer algo.

TEC-02 y HU-01 están a cargo de la misma persona (Peñalbé Hernán), así que se hacen en serie: primero la CI, que es chica, y a continuación HU-01. Mientras la CI se integra, avanza la SDD de HU-01.

Mientras HU-01 está en curso, los responsables de HU-02, HU-03, HU-04 y HU-06 pueden escribir y hacer aprobar la SDD de sus historias, que no depende del código.

## Definition of Done del Sprint 1

- Cada historia tiene su SDD aprobada, escenarios BDD automatizados con godog, tests en verde, código, evidencia TDD (`docs/evidencias/HU-XX-tdd.md`) y trazabilidad en el Issue.
- Los tests pasan (`go test ./...`) y el código está formateado (`gofmt`). El workflow de CI (TEC-02) pasa en el Pull Request.
- Existe una forma mínima de ejecutar y demostrar la historia (TEC-01): "funcional" significa que se pueda ejecutar y mostrar corriendo.
- La rama de la historia está integrada a `main` por Pull Request, con "Create a merge commit".
- El README refleja el estado del Sprint.

## Estrategia de ramas e integración

- Una rama por historia: `feature/HU-01-crear-proyecto`, `feature/HU-02-modificar-proyecto`, y así sucesivamente. Las tareas técnicas siguen el mismo patrón (por ejemplo, `feature/TEC-02-ci`).
- Se integra primero TEC-02 y HU-01, y después el resto, porque las demás historias dependen de la organización definida en HU-01.
- Los commits siguen la convención `tipo(ID): descripción` y citan el Issue en el cuerpo: `Refs #N` (y `Closes #N` en el commit que cierra la historia).
- Al cerrar el Sprint se crea el tag `sprint-1` sobre `main` como línea base.

## Demo del Sprint Review

Se muestra la aplicación corriendo (TEC-01): crear un proyecto, modificarlo, consultar su estado, cargarle ítems al backlog y listarlos. Si alguna historia no llega a integrarse, se muestra lo que sí esté en `main` y se deja constancia en el acta de cierre.

## Responsables

Un responsable por historia o tarea, de modo que ninguna rama tenga dos agentes de IA trabajando a la vez. El responsable figura también en el campo "Assignees" del Issue.

| ID | Issue | Responsable | Story Points |
|---|---|---|---|
| TEC-02 | #32 | Peñalbé Hernán | — |
| HU-01 | #14 | Peñalbé Hernán | 8 |
| HU-02 | #15 | Cardozo Leandro | 3 |
| HU-03 | #16 | Sosa Ricardo | 2 |
| HU-04 | #17 | García Santiago | 5 |
| HU-06 | #21 | Buttini Cristobal | 3 |
| TEC-01 | #23 | Peñalbé Hernán | — |

Se reparte para que cada integrante tenga al menos una historia en el Sprint. Carga: Peñalbé Hernán 8 Story Points (HU-01) más la CI (TEC-02) y el punto de entrada (TEC-01), Cardozo Leandro 3 Story Points (HU-02), Sosa Ricardo 2, García Santiago 5 y Buttini Cristobal 3. Las historias opcionales (HU-05, HU-07 y HU-08) se asignan durante el Sprint a quien tenga menos carga o termine antes, empezando por Sosa Ricardo y Buttini Cristobal.

Dependencias entre responsables: HU-02, HU-03 y HU-04 necesitan la organización de paquetes y la interfaz de repositorio que define HU-01, y HU-06 necesita HU-04. Mientras HU-01 está en curso, los responsables de HU-02, HU-03, HU-04 y HU-06 escriben y hacen aprobar sus SDD, que no dependen del código, y las alinean con las decisiones de la SDD de HU-01 cuando esté aprobada.

## Riesgos

- Cuatro sprints de una semana para 40 historias con ciclo completo: cerca de 10 historias por sprint, unas 2 o 3 por persona por semana. El ritmo real se mide al cierre del Sprint 1 (velocidad comprometida: 21 Story Points) y, si es menor, se rediscute el reparto.
- El Sprint 2 es el más cargado (hasta 14 historias, más TEC-04 y TEC-05). Se revisa su alcance en el Planning del Sprint 2 según la velocidad real del Sprint 1.
- Ciclo completo por historia con equipo nuevo en el proceso: si el ritmo no alcanza, se recorta el alcance (no las pruebas ni la documentación).
- Varias historias dependen de HU-01: si se atrasa, se atrasa todo lo demás. Por eso HU-01 se estimó en 8 puntos y su SDD se prioriza el primer día.
- La CI (TEC-02), HU-01 y el punto de entrada (TEC-01) quedan en una sola persona, y las dos primeras bloquean al resto: si Peñalbé Hernán se atrasa, se atrasan las demás historias. Para reducir el riesgo, la CI se hace el primer día, la SDD de HU-01 se prioriza y las SDD de las demás historias avanzan en paralelo.
- El Sprint 4 es el más cargado en cierre: concentra E6 (HU-21 a HU-25, Planning Poker, requerimiento mínimo de la consigna), el esfuerzo (HU-27 y HU-28), reportes, PDF, documentación final y entrega. En el Planning del Sprint 2 se decide si parte de E6 se adelanta al Sprint 3.
- Los profesores no respondieron: decisiones tomadas sin su validación (ver "Decisiones del equipo").
- Docker y PostgreSQL entran en el Sprint 2: hace falta que todo el equipo tenga Docker instalado y funcionando antes del 16/10. Si alguien no puede, se resuelve en el Planning del Sprint 2.
