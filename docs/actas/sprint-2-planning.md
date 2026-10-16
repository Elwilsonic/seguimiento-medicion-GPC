# Acta de Planning — Sprint 2 (La Interfaz)

> Issue del repositorio: #86

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 2 — La Interfaz
- **Fecha del Planning:** 16/10/2026
- **Período del Sprint:** 16/10/2026 al 22/10/2026 (1 semana)
- **Sprint Goal:** Usar la aplicación desde el navegador para gestionar proyectos, su Product Backlog y sus sprints, con los datos guardados en PostgreSQL.

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

- **Sprint 1 cerrado** (tag `sprint-1`): Sprint Goal cumplido, 21 de 21 Story Points (HU-01, HU-02, HU-03, HU-04 y HU-06), más la historia opcional HU-05 y las tareas técnicas TEC-01 y TEC-02. Ver las actas de [Review](sprint-1-review.md) y [Retrospectiva](sprint-1-retrospectiva.md).
- **Velocidad del Sprint 1:** 21 Story Points; 6 historias integradas (5 comprometidas y 1 opcional).
- **Pendiente hasta la entrega (05/11/2026):** 34 historias (HU-07 a HU-40) en 3 sprints de una semana, unas 11 o 12 historias por sprint. Es casi el doble de las historias que se integraron en el Sprint 1 (ver "Riesgos").

## Ítems candidatos y dependencias

Las dependencias surgen del código de `main` (los `Manager` y las interfaces `Repository` de `internal/project` e `internal/backlog`) y de los criterios de cada Issue; cada SDD las confirma.

| ID | Issue | Historia o tarea | Épica | Prioridad | Depende de |
|---|---|---|---|---|---|
| HU-07 | #19 | Modificar datos de un ítem | E2 · Product Backlog | ALTA | HU-05 (`GetByID`). Agrega `Update` a `backlog.Repository` y el campo de horas estimadas al ítem (CA-07.1, CA-07.4) |
| HU-08 | #20 | Cambiar el estado de un ítem | E2 · Product Backlog | ALTA | HU-05 (`GetByID`) y el `Update` del backlog que agrega HU-07. Define las transiciones de estado (comentario de `item.go`) |
| HU-09 | #22 | Priorizar los ítems del backlog | E2 · Product Backlog | MEDIA | HU-06. El orden por ID de RN-06.6 sigue siendo el predeterminado; el orden por prioridad es una opción (CA-09.1 "puede ordenarse") |
| HU-10 | #24 | Crear sprint con nombre, fechas y Sprint Goal | E3 · Gestión de Sprints | ALTA | HU-01 y HU-03. Crea el paquete de sprints, con su estado abierto o cerrado (lo necesitan CA-10.3, CA-11.3 y HU-13) |
| HU-11 | #25 | Asignar historias del backlog a un sprint | E3 · Gestión de Sprints | ALTA | HU-10 y HU-05. Si la asignación se guarda en el ítem, usa el `Update` del backlog de HU-07 (lo decide su SDD) |
| HU-12 | #26 | Registrar historias completadas | E3 · Gestión de Sprints | ALTA | HU-11. Si "completado" es el estado `TERMINADO`, también HU-08 (lo decide su SDD) |
| HU-13 | #27 | Cerrar sprint | E3 · Gestión de Sprints | ALTA | HU-12 |
| HU-14 | #28 | Consultar sprints anteriores | E3 · Gestión de Sprints | MEDIA | HU-10 y HU-13 |
| HU-15 | #29 | Registrar integrante del proyecto | E4 · Registro de integrantes | MEDIA | HU-01 y HU-03. Crea el paquete de integrantes |
| HU-16 | #30 | Modificar los datos o el rol de un integrante | E4 · Registro de integrantes | MEDIA | HU-15 |
| HU-17 | #31 | Consultar los integrantes de un proyecto | E4 · Registro de integrantes | MEDIA | HU-15 |
| HU-18 | #36 | Registrar fechas de inicio y fin del proyecto | E5 · Fechas del proyecto | MEDIA | HU-01 |
| HU-19 | #37 | Modificar las fechas del proyecto | E5 · Fechas del proyecto | MEDIA | HU-18 |
| TEC-04 | #34 | Interfaz usable (web servida por el servidor de TEC-01) | Técnica | ALTA | TEC-01 |
| TEC-05 | #35 | Persistencia con PostgreSQL en Docker | Técnica | MEDIA | Interfaces `Repository` de HU-01 en adelante |

## Estimación de las historias (Planning Poker)

Escala Fibonacci (1, 2, 3, 5, 8, 13, 21), con HU-03 (2 SP) como historia de referencia y lo que costaron las historias del Sprint 1 como comparación: HU-02, modificar, 3 SP; HU-04, crear con paquete nuevo, 5 SP; HU-06, listar, 3 SP. Las tareas técnicas no se estiman en Story Points.

| ID | Historia | Story Points | Justificación |
|---|---|---|---|
| HU-07 | Modificar datos de un ítem | 3 | Como HU-02 (modificar), más el campo nuevo de horas estimadas y el `Update` del repositorio |
| HU-08 | Cambiar el estado de un ítem | 3 | Define y valida las transiciones de estado; reutiliza el `Update` de HU-07 |
| HU-09 | Priorizar los ítems del backlog | 3 | Orden por prioridad estable, como opción del listado de HU-06 |
| HU-10 | Crear sprint | 5 | Como HU-04: paquete nuevo con entidad, repositorio y `Manager`, validación de fechas y un solo sprint abierto por proyecto |
| HU-11 | Asignar historias a un sprint | 5 | Cruza dos paquetes (sprint y backlog) y tiene tres reglas: mismo proyecto, un solo sprint por ítem y sprint abierto |
| HU-12 | Registrar historias completadas | 3 | Marca ítems del sprint y consulta los completados; su relación con el estado `TERMINADO` se define en la SDD |
| HU-13 | Cerrar sprint | 3 | Cambio de estado del sprint y bloqueo de cambios en un sprint cerrado |
| | **Total comprometido** | **25** | |
| HU-15 (si hay tiempo) | Registrar integrante | 5 | Paquete nuevo, como HU-10 |
| HU-16 (si hay tiempo) | Modificar integrante | 3 | Como HU-02 |
| HU-17 (si hay tiempo) | Consultar integrantes | 3 | Como HU-06 |

Los Story Points son la propuesta para arrancar el Planning Poker; si la votación da otros valores, se reemplazan en esta tabla y en la de responsables.

## Alcance comprometido

- **Comprometido:** HU-07, HU-08 y HU-09 (cierran E2), HU-10, HU-11, HU-12 y HU-13 (E3, prioridad ALTA), TEC-04 y TEC-05. Total: 25 Story Points más las dos tareas técnicas.
- **Si el equipo llega con tiempo:** HU-15, HU-16 y HU-17 (E4, prioridad MEDIA e independientes del resto: se pueden sumar sin bloquear a nadie). Se asignan durante el Sprint a quien termine primero, empezando por HU-15.
- **Al Sprint 3:** HU-14, HU-18 y HU-19, y lo de E4 que no entre.
- **Acuerdo previo:** los 25 Story Points superan la velocidad del Sprint 1 (21), y TEC-04 y TEC-05 son las tareas más pesadas. Si el Sprint se atrasa, **HU-09 es la primera que pasa a "si hay tiempo"** (prioridad MEDIA y sin historias que dependan de ella), sin necesidad de discutirlo a mitad del Sprint.

## Temas decididos

1. **Endpoint de HU-05:** `GET /items/{id}` se agrega dentro de TEC-04; la interfaz lo necesita también para HU-07 y HU-08.
2. **Mejoras del servidor HTTP** (revisión del PR #81): límite de tamaño del cuerpo, rechazo de texto después del JSON y timeouts de `ListenAndServe`. Van dentro de TEC-04, en un Pull Request separado y temprano, para que no queden detrás de toda la interfaz.
3. **`Update` del backlog:** lo agrega HU-07; HU-08 lo reutiliza y arranca cuando HU-07 está en `main`. Las dos tienen el mismo responsable.
4. **Repositorios en PostgreSQL:** TEC-05 implementa los repositorios que estén en `main` cuando arranca (proyectos y backlog). Los métodos y repositorios que se integren después durante el Sprint (el `Update` del backlog, sprints) los agrega el responsable de TEC-05 a medida que entran en `main`. Gracias a CA-TEC05.4 (memoria o PostgreSQL según la configuración), el resto del equipo trabaja sin Docker.
5. **Interfaz web:** HTML y JavaScript sin frameworks, servidos con la librería estándar (`embed` y `http.FileServer` en el mismo `ServeMux` de TEC-01); no se agregan librerías externas sin acuerdo del equipo.
6. **Acciones de la Retrospectiva del Sprint 1** con trabajo en el Sprint 2 (responsable: Peñalbé Hernán): `go vet ./...` en el CI (acción 8), Definition of Ready (acción 9), plantillas de Issue y de Pull Request (acción 10), orden de las carpetas del repositorio (acción 11) y la regla de crear el Issue antes de cada ceremonia en `AGENTS.md`.
7. **Windows bloqueado** (acción 6, García Santiago con apoyo de Peñalbé Hernán): se resuelve el primer día, porque García tiene historias comprometidas. Mientras tanto, la verificación la hacen el CI y Peñalbé Hernán sobre Linux.

## Orden de trabajo

1. **Primer día (16/10):**
   - acciones de proceso de la Retrospectiva: `go vet` en el CI, Definition of Ready y plantillas;
   - SDD de las historias que abren cada cadena: HU-07 y HU-10;
   - TEC-05 arranca con los repositorios de proyectos y backlog;
   - destrabar `go test` en la máquina de García (acción 6).
2. **SDD escalonadas**, para que el Agile Enabler las apruebe a medida que llegan y no todas el mismo día: HU-08 y HU-11 el 17/10; HU-09, HU-12 y HU-13 el 18/10.
3. **E3 en cadena:** HU-10 → HU-11 → HU-12 → HU-13. Mergear HU-10 lo antes posible es la prioridad del Sprint, porque las otras tres dependen de ella.
4. **TEC-04 arranca a mitad del Sprint (19/10)** con lo que ya esté en `main` (archivos estáticos y endpoints del Sprint 1) y suma pantallas a medida que se integran historias. El Pull Request de las mejoras del servidor va primero.

## Definition of Ready (acción 9 de la Retrospectiva)

Una historia puede empezar a implementarse cuando:

- Tiene su Issue con los criterios de aceptación definitivos (sin "(borrador)").
- Su SDD está aprobada por el Agile Enabler.
- Tiene sus Story Points estimados.
- Las historias de las que depende están integradas en `main`.
- Tiene un único responsable asignado en el Issue.

Las SDD se pueden escribir en paralelo, aunque la dependencia no esté en `main`, como en el Sprint 1. La implementación espera a que la dependencia esté integrada: si no, el RED fallaría por código que todavía no existe y no solo por la funcionalidad nueva (paso 4 de `AGENTS.md`), y habría que crear ramas desde otras ramas en lugar de desde `main`.

## Definition of Done del Sprint 2

La del Sprint 1, más lo acordado en la Retrospectiva:

- Cada historia tiene su SDD aprobada, escenarios BDD automatizados con godog, tests en verde, código, evidencia TDD (`docs/evidencias/HU-XX-tdd.md`) y trazabilidad en el Issue.
- Los tests pasan (`go test ./...`, también con `-race` en el CI), el código está formateado (`gofmt`) y `go vet ./...` no informa problemas.
- No queda código nuevo sin un test que lo pida: se revisa `go tool cover -func` antes del commit GREEN.
- Las cifras de la evidencia salen de la salida de comandos, pegada en la evidencia.
- La funcionalidad se puede usar desde la interfaz (TEC-04) cuando corresponde.
- La rama está integrada a `main` por Pull Request, con "Create a merge commit" y la aprobación de un integrante distinto del responsable, que verifica el RED y el GREEN.
- El Issue se cierra con `Closes #N` al hacer merge, nunca a mano.

## Responsables

Un responsable por historia o tarea, con los ítems agrupados por cadena o por el código que cada uno ya conoce, para que haya menos traspasos.

| Integrante | Ítems | Story Points | Motivo |
|---|---|---|---|
| Peñalbé Hernán | TEC-04 y las acciones de la Retrospectiva | — | TEC-04 extiende el servidor de TEC-01 (`internal/httpapi`), que hizo él; como Agile Enabler además aprueba las SDD y coordina |
| Buttini Cristobal | HU-10 y HU-11 | 10 | Inicio de la cadena E3 |
| Cardozo Leandro | TEC-05 y HU-09 | 3 | HU-09 no bloquea a nadie y se puede pasar a "si hay tiempo" si TEC-05 se complica |
| García Santiago | HU-12 y HU-13 | 6 | Fin de la cadena E3, que arranca cuando HU-11 está en `main`; da tiempo a resolver el bloqueo de Windows |
| Sosa Ricardo | HU-07 y HU-08 | 6 | Se apoyan en `GetByID`, que hizo en HU-05 |

Las historias de "si hay tiempo" (HU-15, HU-16 y HU-17) se asignan durante el Sprint a quien termine primero.

## Riesgos

- **Ritmo:** quedan 34 historias para 3 sprints (unas 11 o 12 por sprint); en el Sprint 1 se integraron 6 y en este Sprint se comprometen 7. En la Review del Sprint 2 se decide con el equipo qué se recorta del plan (decisión 8 del acta de Planning del Sprint 1: se recorta alcance, no pruebas ni documentación).
- **Alcance por encima de la velocidad:** 25 Story Points contra 21 del Sprint 1, más TEC-04 y TEC-05. Mitigación: HU-09 sale primero si hay atraso.
- **Cadena de E3:** HU-10 a HU-13 dependen una de otra; si HU-10 se atrasa, se atrasa toda la épica y la parte de sprints de TEC-04.
- **Docker:** TEC-05 necesita Docker en la máquina de su responsable; el resto del equipo trabaja en memoria (CA-TEC05.4). El responsable confirma el primer día que Docker funciona.
- **Windows:** García tiene historias comprometidas y en su máquina `go test` está bloqueado; se resuelve el primer día (acción 6) y, mientras tanto, verifican el CI y Peñalbé Hernán.
- **Carga del Agile Enabler:** TEC-04, las acciones de la Retrospectiva y la aprobación de 7 SDD; por eso las SDD llegan escalonadas.

## Próximo paso

- Cada responsable actualiza su Issue con los criterios de aceptación definitivos y escribe su SDD según el orden de trabajo.
- El responsable de TEC-05 confirma que Docker funciona en su máquina el 16/10.
- Las historias y tareas comprometidas pasan al Sprint 2 del tablero, con su responsable en "Assignees".
- La Daily se registra en el Issue #87; la Review y la Retrospectiva del Sprint 2 son el 22/10 (Issues #88 y #89).
