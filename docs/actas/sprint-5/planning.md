# Acta de Planning — Sprint 5 (Proyecto heredado)

> Issue del repositorio: #182

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 5 — Proyecto heredado
- **Fecha del Planning:** 07/11/2026
- **Período del Sprint:** 07/11/2026 al 15/11/2026 (9 días; la presentación es el 16/11/2026)
- **Sprint Goal:** Aplicar SDD con OpenSpec al proyecto heredado y resolver lo que marcó el profesor: listado de proyectos, IDs no secuenciales y cada test vinculado a su historia.

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

- **Sprint 4 cerrado:** Sprint Goal cumplido, con 35 de 35 Story Points comprometidos y 24 más tomados con capacidad sobrante. Después se sumó HU-49 (reporte con las historias y sus estimaciones, 3 SP): con ella, la velocidad fue de 62 y las 43 historias del Product Backlog quedaron cerradas. Ver las actas de [Review](../sprint-4/review.md) y [Retrospectiva](../sprint-4/retrospectiva.md).
- **Velocidad de referencia:** 21, 25, 34 y 62 Story Points en los Sprints 1 a 4.
- **Calidad en `main` al cierre del Sprint 4, con HU-49:** 84.7 % de cobertura sin base y 96.2 % con PostgreSQL; 665 escenarios BDD ejecutados.
- **Después del Sprint 4 entraron a `main`:** el despliegue en Azure Container Apps (TEC-09, PR #173 a #176, #180 y #181) y la instalación de OpenSpec 1.16.0 (TEC-10, PR #178).
- **Revisión del profesor (05/11/2026):** el equipo tiene una nueva oportunidad la semana del 16/11. El profesor marcó cinco puntos, que son el motivo de este Sprint:

| # | Lo que pidió el profesor | Qué lo resuelve |
|---|---|---|
| 1 | Usar una herramienta de SDD reconocida y tratar el proyecto como heredado (Legacy) | TEC-10 (OpenSpec) |
| 2 | Agregar un listado de proyectos: hoy hay que saber su número | HU-50 |
| 3 | No usar IDs 1, 2, 3, que son fáciles de adivinar; usar UUID | HU-51 |
| 4 | Que cada test diga a qué historia pertenece y qué verifica | TEC-11 |
| 5 | Entender lo que se sube, aunque lo haya hecho la IA | Cada PR revisado por otro integrante, con comentarios |

- **Lo que falta hasta la presentación:** los puntos 2 a 4, el ciclo completo de un cambio con OpenSpec (HU-50) y la preparación de la demo.

## Ítems candidatos y dependencias

Las dependencias se verifican contra el código de `main` y los criterios de cada Issue; cada propuesta de OpenSpec las confirma.

| ID | Issue | Historia o tarea | Épica | Prioridad | Depende de |
|---|---|---|---|---|---|
| HU-50 | pendiente | Listar los proyectos (vinculada a HU-03) | E1 | ALTA | HU-03 ✅; TEC-10 ✅ (OpenSpec en `main`) |
| HU-51 | pendiente | IDs UUID (vinculada a HU-01 y a las historias que crean entidades) | E1, con alcance en E2 a E8 | ALTA | HU-01 ✅; su propuesta de OpenSpec aprobada |
| TEC-10 | #177 | Adoptar OpenSpec | — | ALTA | OpenSpec en `main` ✅ (PR #178); se cierra cuando HU-50 se archive |
| TEC-11 | #179 | Comentar los tests con su historia | — | ALTA | El mapa de tests (primero); la parte de García, después de HU-50 |
| TEC-09 | #172 | Despliegue en Azure | — | MEDIA | PR #181 ✅ (integrado) |

## Estimación de las historias (Planning Poker)

Escala Fibonacci (1, 2, 3, 5, 8, 13, 21), con HU-03 (2 SP) como historia de referencia. Las tareas técnicas no se estiman en Story Points.

| ID | Historia | Story Points | Justificación |
|---|---|---|---|
| HU-50 | Listar los proyectos | 3 | Un poco más que HU-03: es una consulta de todos los proyectos y una pantalla, sobre lo que ya existe. Sirve de ejemplo del ciclo con OpenSpec |
| HU-51 | IDs UUID | 13 | El ID se usa en todas las entidades (proyectos, backlog, sprints, defectos), en el esquema de la base, en la API y en la interfaz. Es la más grande del Sprint y se hace por partes |
| | **Total comprometido** | **16** | |

## Alcance comprometido

- **Comprometido:** HU-50 (3 SP) y HU-51 (13 SP, empezando por proyectos), más TEC-10, TEC-11 y TEC-09, y la preparación de la presentación. Total: 16 Story Points más las tareas técnicas.
- **Si el equipo llega con tiempo:** HU-51 en el resto de las entidades (backlog, sprints, defectos y esfuerzo).
- **Al Sprint siguiente:** no hay Sprint siguiente. Lo que no llegue de HU-51 queda escrito en la propuesta de OpenSpec como trabajo pendiente.
- **Acuerdo previo:** si el Sprint se atrasa, HU-51 se recorta por partes: primero proyectos (obligatorio) y después el resto. No se recortan las pruebas ni la documentación.

## Temas decididos

1. **Fechas y ceremonias:** el Sprint 5 va del 07/11 al 15/11. La Review, la Retrospectiva y el ensayo de la demo son el 15/11, y la presentación es el 16/11. Las Dailys son día por medio, por los exámenes (lunes 09, miércoles 11 y viernes 13), en un único Issue "S5 · Daily del Sprint 5", con un comentario por día: qué hice, qué voy a hacer y si tengo algún bloqueo.
2. **OpenSpec como herramienta de SDD:** cada cambio nuevo se especifica con OpenSpec (`/opsx:propose`), que genera `proposal.md`, `specs/`, `design.md` y `tasks.md` en `openspec/changes/<cambio>/`. Quedó registrado en la decisión 22 de `docs/decisiones-tecnicas.md`. Las 43 historias anteriores conservan su SDD como material de consulta y no se rehacen. La propuesta la aprueba el Agile Enabler antes de programar.
3. **HU-50 es el primer cambio con OpenSpec:** pasa por todo el ciclo (explorar, proponer, aprobar, BDD, RED, GREEN, evidencia, PR y archivar) y sirve de ejemplo para el resto. TEC-10 se cierra cuando HU-50 se archive.
4. **HU-51 por partes:** primero proyectos (obligatorio) y después el resto, si hay tiempo. Empieza apenas se apruebe su propuesta y se trabaja en pareja. Lo que no llegue queda documentado en la propuesta.
5. **TEC-11, comentarios en los tests:** cada test lleva arriba una línea con su historia, lo que prueba y qué verifica. Hay 1013 tests sin la historia en su comentario (el Issue #179 decía 997; el número exacto está en el mapa), y 644 de ellos ya citan su criterio.
   - **Formato:** `// HU-XX · CA-XX.N · Verifica que...` si prueba un criterio de aceptación; `// HU-XX · RN-XX.N · Verifica que...` si prueba una regla de negocio que no es un criterio; `// HU-XX · Error · Verifica que...` si prueba una condición de error que la SDD describe sin número. Los tests de tareas técnicas llevan `// TEC-XX · CA-TECXX.N · Verifica que...`. Solo se agregan comentarios: no se cambian los tests ni las SDD.
   - **Reparto por autor:** cada uno comenta los tests de las historias que hizo, según las actas de Review, porque conoce sus criterios. La historia de cada test sale del commit que lo agregó. Cada archivo queda para un solo responsable, el autor con más tests en él, para que dos personas no editen el mismo archivo.
   - **Equilibrio de la carga:** Sosa le pasa a Buttini `internal/httpapi/web_test.go` y los tests de backlog de HU-05, HU-07 y HU-08 (95 tests), porque también tiene HU-51. Cardozo le pasa a García los tests de los reportes de proyecto (HU-38, 21 tests), porque García hizo los de sprint (HU-39).
   - **Resultado:** Buttini 214, Cardozo 223, García 183, Hernán 203 y Sosa 190.
   - **Orden:** (1) el Agile Enabler sube el mapa con la historia y el responsable de cada archivo de tests (`docs/evidencias/TEC-11-tests-por-historia.md`); (2) cada uno comenta lo suyo en su rama (`feature/TEC-11-<apellido>`) y abre su PR con `Refs #179`; (3) un PR de cierre agrega un test de control, con `Closes #179`, que falla si un test no tiene la línea, si la historia no tiene SDD o si el criterio o la regla citados no aparecen en esa SDD.
   - No se inventa nada: si un test no encaja con ningún criterio, regla ni condición de error, se pregunta al autor de la historia.
6. **Responsabilidad con la IA:** cada integrante lee y entiende lo que sube, aunque lo haya generado la IA, como pidió el profesor. Cada PR lo revisa otro integrante, con comentarios.
7. **TEC-09 y la demo en Azure:** la app se apaga sola cuando nadie la usa (mínimo cero réplicas, PR #181), para no gastar el crédito compartido de la cátedra, y al apagarse pierde los datos. Antes de la demo se cargan los datos con el script.

## Orden de trabajo

| Fecha | Qué tiene que estar |
|---|---|
| Sáb 07/11 | Planning |
| Dom 08/11 | Propuesta de OpenSpec de HU-50 y mapa de tests (TEC-11) |
| Lun 09/11 | Daily |
| Mié 11/11 | Daily. HU-50 integrada y archivada (se cierra TEC-10). Propuesta de HU-51 aprobada |
| Vie 13/11 | Daily. Los cinco PR de TEC-11 integrados |
| Sáb 14/11 | HU-51 integrada (al menos proyectos) |
| Dom 15/11 | Review, Retrospectiva, ensayo de la demo y tag `sprint-5` |
| **Lun 16/11** | **Presentación** |

1. HU-50 primero: es chica y es el ejemplo del ciclo con OpenSpec.
2. El mapa de tests antes que los PR de TEC-11.
3. HU-51 arranca con la propuesta aprobada, empezando por proyectos.
4. El PR de cierre de TEC-11 (el test de control) va al final.

## Definition of Ready y Definition of Done

Ver [`docs/definition-of-ready-done.md`](../../definition-of-ready-done.md). Se mantienen las del Sprint 4, más los cambios acordados en este Planning:

- Cada cambio se especifica con OpenSpec (`/opsx:propose`) y el Agile Enabler aprueba la propuesta antes de programar.
- Sigue la cadena de siempre: escenarios BDD con godog, tests en rojo, código en verde, evidencia TDD y Pull Request revisado por otro integrante.
- Cada test lleva arriba un comentario con su historia, lo que prueba y qué verifica (`// HU-XX · CA-XX.N · Verifica que...`, o `RN-XX.N` o `Error` en lugar del criterio, según la decisión 5).
- La evidencia TDD incluye la tabla criterio → test → código.
- Con el Pull Request integrado, el cambio se archiva (`/opsx:archive`) en un commit propio.
- Cada integrante lee y entiende lo que sube, aunque lo haya generado la IA.
- OpenSpec se instala con la versión fija y la telemetría apagada (`npm install -g @fission-ai/openspec@1.16.0` y `openspec config set telemetry.enabled false`; necesita Node.js 22.12 o superior).

## Responsables

Un responsable por historia o tarea; cada integrante con al menos dos ítems.

| Integrante | Ítems | Story Points | Motivo |
|---|---|---|---|
| Peñalbé Hernán | TEC-10, TEC-09, TEC-11 (203 tests: métricas, creación de proyectos, integrantes, reportes en PDF, TEC-01 y TEC-04) y el mapa de tests | — (tareas técnicas) | Como Agile Enabler aprueba las propuestas de OpenSpec y coordina; TEC-10 y TEC-09 ya están casi terminadas |
| García Santiago | HU-50 y TEC-11 (183 tests: esfuerzo, horas, sprints completados y cerrados, creación de ítems y reportes de sprint y de proyecto; después de HU-50) | 3 | HU-50 es chica y es el primer cambio con OpenSpec; su parte de TEC-11 espera a que termine. Toma los reportes de proyecto (HU-38) porque hizo los de sprint (HU-39) |
| Sosa Ricardo | HU-51 y TEC-11 (190 tests: proyectos y defectos) | 13 | El UUID empieza por proyectos, que es su área. Le pasa `web_test.go` y los tests de backlog a Buttini para equilibrar la carga con HU-51 |
| Cardozo Leandro | Apoyo en HU-51 (backlog e integrantes) y TEC-11 (223 tests: Planning Poker, integrantes, story points y prioridad del backlog, y TEC-05) | — | Hizo HU-09 (priorizar el backlog) y HU-16 (modificar integrante), y ayuda a Sosa con esas partes de HU-51. Le pasa los reportes de proyecto a García |
| Buttini Cristobal | TEC-11 (214 tests: sprints, dashboard y Planning Poker, más `web_test.go` y backlog de Sosa) y la presentación | — | Sus historias suman menos tests (119), por eso toma `web_test.go` y los tests de backlog de Sosa. Lo que le sobre puede apoyar a HU-51, que se hace en pareja |

## Riesgos

- **Poco tiempo, por los exámenes:** Dailys día por medio y recorte de HU-51 por partes.
- **HU-51 es grande:** toca el esquema de la base, las rutas, las claves foráneas y casi todos los tests. Empieza apenas se apruebe la propuesta, se hace en pareja y se recorta por partes.
- **OpenSpec es nueva para todo el equipo:** HU-50 es el ejemplo, y las skills y los comandos se revisan antes de usarlos (decisión 22).
- **El entorno de Azure puede volver a borrarse:** el despliegue lo recrea solo, y la URL puede cambiar.
- **La app en Azure se apaga sola y pierde los datos:** antes de la demo se cargan los datos con el script.
- **TEC-11 depende del mapa:** si el mapa se atrasa, los cinco PR de comentarios se atrasan.

## Próximo paso

- Hernán sube el mapa de tests y García empieza la propuesta de HU-50 con `/opsx:explore` y `/opsx:propose`.
- Se crean los Issues de HU-50 y HU-51 con sus criterios, y las historias y tareas pasan al Sprint 5 del tablero.
- Se actualiza el reparto de TEC-11 en el Issue #179.
- La Daily se registra en el Issue "S5 · Daily del Sprint 5"; la Review y la Retrospectiva del Sprint 5 tienen sus propios Issues.
