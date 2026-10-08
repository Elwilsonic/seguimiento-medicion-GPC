# Acta de cierre — Sprint 0: La Preparación

- **Período del sprint:** 07/09/2026 – 08/10/2026
- **Fecha de la reunión de cierre:** 08/10/2026
- **Asistentes:** Peñalbé Hernán (Agile Enabler), Buttini Cristobal, Cardozo Leandro, García Santiago y Sosa Ricardo (Product Builders)
- **Product Architect:** Profesores de la cátedra (no participaron de la reunión)

## Objetivo del Sprint 0

Formar el equipo, asignar roles, configurar el entorno de trabajo y construir el Product Backlog inicial.

## Resultado por objetivo

| Objetivo | Estado | Evidencia |
|---|---|---|
| Formar el equipo y asignar roles | ✅ Cumplido | Roles registrados en el README |
| Crear el repositorio | ✅ Cumplido | Repositorio en GitHub con README, AGENTS.md y módulo Go inicializado |
| Configurar el tablero en GitHub Projects | ✅ Cumplido | Tablero "Software Metrics & Estimation" |
| Reunión inicial con el Product Architect | ❌ No realizada | No hubo reunión inicial. La visión del producto y los requerimientos se tomaron del enunciado del Trabajo Práctico Integrador. |
| Crear el Product Backlog inicial | ✅ Cumplido | 11 épicas, 40 historias de usuario y 5 tareas técnicas cargadas como Issues |
| Estimar las épicas (Planning Poker con talles) | ✅ Cumplido | Tabla de estimación en el README |

## Product Backlog resultante

| Épica | Prioridad | Talle | Historias |
|---|---|---|---|
| E1 · Gestión de proyectos | Alta | S | HU-01 a HU-03 |
| E2 · Product Backlog | Alta | M | HU-04 a HU-09 |
| E3 · Gestión de Sprints | Alta | L | HU-10 a HU-14 |
| E4 · Registro de integrantes | Media | S | HU-15 a HU-17 |
| E5 · Fechas del proyecto | Media | XS | HU-18 y HU-19 |
| E6 · Estimación | Media | L | HU-20 a HU-25 |
| E7 · Registro de esfuerzo | Media | M | HU-26 a HU-28 |
| E8 · Gestión de defectos | Baja | S | HU-29 a HU-31 |
| E9 · Métricas | Baja | L | HU-32 a HU-35 |
| E10 · Dashboard | Baja | L | HU-36 y HU-37 |
| E11 · Reportes | Baja | XL | HU-38 a HU-40 |

Tareas técnicas: TEC-01 (punto de entrada), TEC-02 (integración continua), TEC-03 (fórmulas de métricas), TEC-04 (interfaz), TEC-05 (persistencia con PostgreSQL).

Cada historia tiene su historia de usuario, épica, prioridad y criterios de aceptación en borrador, con etiquetas `epica:*`, `prioridad:*` y `tipo:*`, y está vinculada a su épica como sub-issue. Las tareas técnicas no pertenecen a ninguna épica: llevan las etiquetas `prioridad:*` y `tipo:tecnica`.

## Decisiones tomadas

1. **Metodología de trabajo:** cada historia sigue la cadena Historia → SDD → BDD → Tests (RED) → Código (GREEN) → Refactor → Evidencia, definida en `AGENTS.md`.
2. **Escalas de estimación:** talles de ropa (XS a XL) para las épicas y Fibonacci (1, 2, 3, 5, 8, 13, 21) para las historias de usuario, porque las métricas requieren sumar Story Points.
3. **Horas estimadas:** se registran por ítem del backlog (HU-07). No se modelan tareas; el esfuerzo real se registra sobre la historia.
4. **Persistencia:** en memoria durante el Sprint 1; PostgreSQL en Docker a partir del Sprint 2 (TEC-05), detrás de una interfaz de repositorio.
5. **Librerías externas:** solo las aprobadas por el equipo: godog (automatización de los escenarios BDD, solo en tests), el driver de PostgreSQL y la librería de PDF.
6. **Tareas técnicas:** las que dan soporte a las historias van al inicio del sprint y las que exponen lo construido, al final. TEC-02 (integración continua) pasó del Sprint 2 al Sprint 1 para que los primeros Pull Requests con código ya se controlen con `gofmt` y `go test ./...`; se hace al inicio del Sprint 1. TEC-01 (punto de entrada ejecutable) se hace al final del Sprint 1.
7. **Alcance del trabajo:** el equipo decidió trabajar con las 11 épicas y las 40 historias de usuario, aplicando el ciclo completo (SDD → BDD → TDD) a todas. Los profesores indicaron que la estimación de las épicas debe figurar en el README. La consulta sobre el alcance de las historias de usuario no tuvo respuesta, por lo que esa decisión es del equipo.
8. **Identificadores:** cada historia tiene dos identificadores distintos que se usan juntos. `HU-XX` es el ID del Product Backlog (figura en el título del Issue, y en archivos, tests, evidencias y commits). `#N` es el número del Issue en GitHub, y los commits lo citan con `Refs #N` (y `Closes #N` en el commit que cierra la historia). El título de cada commit usa el formato `tipo(HU-XX): descripción` (por ejemplo, `test(HU-01): agregar tests (RED)`), para que se vea a qué historia pertenece cada cambio.

## Pendientes que pasan al Sprint 1

- Definir en la SDD de la primera historia la organización de paquetes y la interfaz de repositorio (requisito de TEC-05).
- Definir en las SDD los estados válidos de proyecto, ítem del backlog y defecto.
- Estimar en Story Points las historias del Sprint 1.

## Retrospectiva

**¿Qué salió bien?**
- El Product Backlog quedó completo y cubre todos los requerimientos mínimos de la consigna.
- Cada historia tiene criterios de aceptación y está vinculada a su épica, lo que facilita la trazabilidad.
- Las decisiones de diseño (escalas de estimación, horas estimadas, persistencia) se tomaron y quedaron documentadas antes de empezar a programar.

**¿Qué se puede mejorar?**
- La carga de las historias se concentró al final del sprint.
- Hubo inconsistencias al cargar los Issues (prioridad distinta entre etiqueta y cuerpo, formato de checkboxes, épicas con formatos diferentes y dos Issues duplicados), que hubo que corregir después.

**Acciones para el Sprint 1**
- Usar un formato único de Issue para historias y tareas técnicas, y revisarlo antes de guardar.
- Distribuir el trabajo a lo largo del sprint en lugar de concentrarlo al final.
- Estimar las historias en el Sprint Planning, antes de empezar a desarrollarlas.

## Próximo paso

Sprint 1 Planning: definir el Sprint Goal y estimar el alcance del MVP (HU-01, HU-02, HU-03, HU-04, HU-06, TEC-01 y TEC-02).
