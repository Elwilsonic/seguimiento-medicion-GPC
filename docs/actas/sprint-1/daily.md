# Daily — Sprint 1 (El MVP)

> Issue del repositorio: #62

Registro consolidado de las Dailys del Sprint 1 (09/10/2026 al 15/10/2026), según la decisión 15 del [acta de Planning](planning.md). Cada Daily se publicó como un comentario en el Issue #62, donde GitHub registra la fecha y la hora; este archivo transcribe esos comentarios **sin cambios en el texto de cada integrante** (incluidas las marcas que quedaron sin completar, como `[confirmar: ...]`), organizados en una tabla por día. Las horas son las que publicó cada comentario, en hora de Argentina (UTC-3).

La Daily del 09/10/2026 (primer día del Sprint, día del Planning) no quedó registrada: el registro empezó el 10/10, cuando se creó el Issue #62.

| Fecha | Registrada por | Publicada | Editada | Comentario |
|---|---|---|---|---|
| 10/10/2026 | Peñalbé Hernán | 10/10/2026 18:07 | 10/10/2026 18:38 | [ver en el Issue #62](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/62#issuecomment-6102189413) |
| 11/10/2026 | Buttini Cristobal | 11/10/2026 15:57 | — | [ver en el Issue #62](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/62#issuecomment-6112525387) |
| 12/10/2026 | Cardozo Leandro | 12/10/2026 13:13 | — | [ver en el Issue #62](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/62#issuecomment-6128096509) |
| 13/10/2026 | García Santiago | 13/10/2026 11:44 | — | [ver en el Issue #62](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/62#issuecomment-6148188542) |
| 14/10/2026 | Sosa Ricardo | 14/10/2026 11:04 | 14/10/2026 14:51 | [ver en el Issue #62](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/62#issuecomment-6170223435) |
| 15/10/2026 | Peñalbé Hernán | 15/10/2026 12:23 | 15/10/2026 19:45 | [ver en el Issue #62](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/62#issuecomment-6195680259) |

## Daily · 10/10/2026

_El registro de Dailys arranca hoy, 10/10. No hay registros de días anteriores del Sprint 1 (que empezó el 09/10)._

| Integrante | Ayer | Hoy | Bloqueos | Horas |
|---|---|---|---|---|
| Hernán | Planning del Sprint 1 (acta mergeada en el PR #61) y ajustes del tablero. | resolví un incidente con Copilot: se asignó a TEC-02 y abrió el PR #63, que cerré sin mergear y con su rama borrada; acordamos no usar "Assign to Agent". Creé los Issues de ceremonias del Sprint 1 (Daily #62, Review #64 y Retrospectiva #65) y ordené el tablero. Abrí el PR #66 para registrar el incidente en el registro de IA (Refs #65). Sigue empezar TEC-02 (#32) en `feature/TEC-02-ci`. | ninguno.<br>Dependencias: HU-02, HU-03, HU-04 y HU-06 necesitan HU-01 en `main` para implementar; mientras tanto se pueden escribir sus SDD. | — |

**Resto del equipo:** El equipo envía su Daily (ayer / hoy / bloqueos) por el grupo y un integrante por día, por turnos, la registra en un único comentario en este Issue.

## Daily · 11/10/2026

_Registrada al final del día; los bloqueos reflejan la situación al momento de publicarla._

| Integrante | Ayer | Hoy | Bloqueos | Horas (10/10) |
|---|---|---|---|---|
| Hernán | terminé TEC-02 (#32): workflow de CI con `gofmt` y `go test ./...`, mergeado en el PR #67, y el check "gofmt y tests" quedó obligatorio en `main`. Hice HU-01 (#14) completa (SDD aprobada, escenarios BDD con godog, tests RED, código GREEN, refactor y evidencia TDD) y abrí el PR #68 con el CI en verde. Nos juntamos con el equipo y cada uno trabajó en la SDD de su historia: HU-02 (Cardozo), HU-03 (Sosa) y HU-04 (García). Buttini no llegó a escribir la de HU-06. Todavía no están subidas al repo. | cargué TEC-06 (#69), TEC-07 (#70) y TEC-08 (#71). Hice el merge de HU-01 (PR #68) y de los ajustes de la revisión de Cardozo (PR #72). Reviso la SDD de HU-03. | ninguno. | ~1,5 h · TEC-02 (#32)<br>~1 h · HU-01 (#14)<br>~2,5 h · Agile Enabler (Issues de ceremonias, incidente de Copilot, PR #66, Daily y plan del Sprint) |
| Buttini Cristobal | participé de la reunión del equipo y tuve problemas con la API key del modelo de Nvidia, que pude resolver. No llegué a escribir la SDD de HU-06 (#21). | escribo la SDD de HU-06 y la alineo con las decisiones de HU-01 y HU-04. | la implementación de HU-06 necesita HU-04 (#17) en `main`. | ~1,5 h · HU-06 (#21) |
| Cardozo Leandro | escribí la SDD de HU-02 (#15) en la reunión del equipo; todavía no está en el repo. | revisé y aprobé el PR #68 de HU-01 y el PR #72 de sus ajustes. Reviso que la SDD de HU-02 coincida con las decisiones de HU-01 y la corrijo si hace falta. | HU-02 necesita la búsqueda por ID de HU-03 (#16) en `main` para implementar. | ~1 h · HU-02 (#15) |
| García Santiago | escribí la SDD de HU-04 (#17) en la reunión del equipo; todavía no está en el repo. | reviso que la SDD de HU-04 coincida con las decisiones de HU-01 y la corrijo si hace falta. | HU-04 necesita la búsqueda por ID de HU-03 (#16) en `main` para implementar. | ~1,5 h · HU-04 (#17) |
| Sosa Ricardo | escribí la SDD de HU-03 (#16) en la reunión del equipo; todavía no está en el repo. | empiezo HU-03. Alineo la SDD con HU-01 (ya en `main`), le agrego la búsqueda por ID que usan HU-02 y HU-04, y la paso a revisión. Con la SDD aprobada, sigo con BDD y tests. | ninguno. | ~1 h · HU-03 (#16) |

## Daily · 12/10/2026

| Integrante | Ayer | Hoy | Bloqueos | Horas (11/10) |
|---|---|---|---|---|
| Hernán | hice el merge de HU-01 (PR #68) y de los ajustes de la revisión de Cardozo (PR #72): tests con un repositorio que falla, singular en los pasos BDD (RED → GREEN), -race en el CI y la desviación del ciclo TDD documentada en la evidencia. Cargué TEC-06 (#69), TEC-07 (#70), TEC-08 (#71) y el Issue #73 con las observaciones pendientes de HU-01. Revisé la SDD y los tests de HU-03 de Sosa, verifiqué CA-03.4 con go test -race (con y sin mutex), abrí el PR #74 y actualicé el Issue #16; HU-03 quedó mergeada. Abrí el PR #75 con el registro de IA actualizado, y respondí las observaciones de la revisión de Cardozo: confirmé con git log que solo los commits de TEC-02 y HU-01 llevan la firma de la IA. Asignamos HU-05 (#18) a Sosa. | agrego en AGENTS.md la regla de no incluir firmas de la IA, como sugirió Cardozo en el PR #75. Reviso las SDD de HU-02 (Cardozo), HU-04 (García) y HU-05 (Sosa) y sigo el avance del Sprint. TEC-01 (#23) arranca el miércoles 14. | ninguno. | ~1,5 h · HU-01 (#14) — ajustes de la revisión<br>~2 h · HU-03 (#16) — revisión, verificación con -race, PR e Issue<br>~1 h · TEC-06 (#69), TEC-07 (#70) y TEC-08 (#71)<br>~1,5 h · Agile Enabler (Daily, registro de IA, merges y coordinación) |
| Buttini Cristobal | registré la Daily del 11/10 en el Issue #62. Avancé con la SDD de HU-06 | termino la SDD de HU-06 alineada con HU-01 y HU-03. No esta subida al repositorio | la implementación de HU-06 necesita HU-04 (#17) en main. | ~1,5 h · HU-06 (#21) |
| Cardozo Leandro | revisé y aprobé el PR #68 de HU-01, con prueba de mutación y cinco observaciones, y el PR #72 de sus ajustes, con una observación más (Issue #73). Revisé y aprobé el PR #75 del registro de IA, con dos observaciones: agregar a AGENTS.md la regla de no firmar con la IA y confirmar qué commits llevan Co-Authored-By. [confirmar: revisé la SDD de HU-02 (#15) contra HU-01 y HU-03] | paso la SDD de HU-02 a revisión, usando la búsqueda por ID de HU-03, y con la SDD aprobada empiezo BDD y tests. | ninguno (HU-03 ya está en main). | ~1,5 h · Revisión de los PR #68, #72 y #75<br>~1,5 h · HU-02 (#15) |
| García Santiago | revisé y aprobé el PR #74 de HU-03. Revisé la SDD de HU-04 contra HU-01 y HU-03] | paso la SDD de HU-04 (#17) a revisión, usando Get y GetByID de HU-03 para validar que el proyecto exista, y con la SDD aprobada empiezo BDD y tests. | ninguno (HU-03 ya está en main). | ~0,5 h · Revisión del PR #74<br>~2 h · HU-04 (#17) |
| Sosa Ricardo | terminé HU-03 (#16): SDD aprobada, escenarios BDD con godog, tests RED (3315d1c), código GREEN (fb24076) y evidencia TDD (d4a9d2b); el PR #74 quedó mergeado y el Issue #16 cerrado. Tomé HU-05 (#18) · Consultar un ítem del backlog, una de las historias opcionales del Sprint 1. | escribo la SDD de HU-05 alineada con HU-04 (estructura del ítem y repositorio del backlog) y la paso a revisión cuando la SDD de HU-04 esté aprobada. | la implementación de HU-05 necesita HU-04 (#17) en main. | ~ 1 h · HU-03 (#16) |

## Daily · 13/10/2026

| Integrante | Ayer | Hoy | Bloqueos | Horas (12/10) |
|---|---|---|---|---|
| Hernán | hice el merge del PR #75 (registro de IA) y respondí las observaciones de Cardozo; agregué en `AGENTS.md` la regla de no firmar con la IA (PR #76, mergeado). Revisé y corregí las SDD de HU-04 (García), HU-06 (Buttini) y HU-02 (Cardozo, aprobada), y armé la SDD de HU-05 para Sosa. Revisé las ramas de HU-04 y HU-06 sobre Linux (RED, GREEN, `go test -race`, vet y cobertura); en HU-04 se documentó una desviación del ciclo TDD. Ordené el tablero (las épicas no van en el Sprint Backlog). | reviso las ramas de HU-02 y HU-05 cuando estén subidas, preparo TEC-01 (#23), que arranca mañana, y organizo la Review (#64) y la Retro (#65) del jueves. | ninguno. | ~2 h · HU-04 (#17): corrección de la SDD y revisión de la rama<br>~1,5 h · HU-06 (#21): SDD y revisión de la rama<br>~1 h · HU-02 (#15): revisión de la SDD<br>~0,5 h · HU-05 (#18): SDD<br>~2 h · Agile Enabler (PR #75 y #76, Daily, tablero y coordinación) |
| Buttini Cristobal | terminé HU-06 (#21): SDD aprobada, escenarios BDD, tests RED (`9b7a387`), código GREEN (`20c8dd0`) y evidencia TDD; el PR #78 quedó mergeado y el Issue #21 cerrado. Revisé y aprobé el PR #77 de HU-04, verificando el RED y el GREEN. | reviso los PR de HU-02 y HU-05 / ayudo con TEC-01 | ninguno. | ~3,5 h · HU-06 (#21): SDD, escenarios BDD, tests RED, código GREEN, evidencia TDD y push de la rama<br>~1 h · Revisión del PR #77 de HU-04 |
| Cardozo Leandro | escribí la SDD de HU-02 (#15) y la ajusté con las decisiones del Agile Enabler (estado sin distinguir mayúsculas, error del repositorio, firma de `Update`); quedó aprobada. | implemento HU-02: rama, SDD, escenarios BDD, tests RED, código GREEN y evidencia, y subo la rama para revisión. | ninguno (HU-01 y HU-03 están en `main`). | ~1 h · HU-02 (#15) |
| García Santiago | terminé HU-04 (#17): SDD aprobada, escenarios BDD, tests RED (`98f27ad`), código GREEN (`d44fcd4`), evidencia TDD y la documentación de una desviación del ciclo (2 pasos BDD ajustados en el GREEN); el PR #77 quedó mergeado. Revisé y aprobé el PR #78 de HU-06. | tildo los criterios del Issue #17 | en mi máquina, Control inteligente de aplicaciones de Windows bloquea los ejecutables de `go test`; la verificación completa la hicieron el CI y Hernán sobre Linux. | ~3 h · HU-04 (#17)<br>~1 h · Revisión del PR #78 |
| Sosa Ricardo | hice HU-05 (#18) / revisión de la SDD. | arranco HU-05 (#18) con la SDD aprobada: rama, escenarios BDD, tests RED, código GREEN y evidencia. | ninguno (HU-04 y HU-06 ya están en `main`). | ~1h · HU-05 (#18) |

## Daily · 14/10/2026

| Integrante | Ayer | Hoy | Bloqueos | Horas (13/10) |
|---|---|---|---|---|
| Hernán | aprobé la SDD final de HU-02 y revisé la rama sobre Linux (RED, GREEN, `go test -race`, vet, gofmt y cobertura del 100 %; la desviación del ciclo TDD quedó documentada con prueba de mutación). Armé la descripción del PR #80 y el texto del Issue #15; HU-02 quedó mergeada y con eso se completaron los 21 SP comprometidos del Sprint 1. Definí el alcance de TEC-01 (#23): servidor con `net/http`, puerto por variable `PORT`, JSON en español y códigos 400/404/409/500. Coordiné con Sosa la sincronización de su repo para arrancar HU-05. | implemento TEC-01 (#23) con TDD (tests de los endpoints, servidor, prueba manual con `curl`, evidencia, README y PR); si sobra tiempo, el #73 (tests del recorte Unicode de HU-01). Reviso la rama de HU-05 cuando esté subida y preparo la Review (#64) y la Retro (#65) de mañana. | ninguno. | ~1,5 h · HU-02 (#15): revisión de la SDD y de la rama<br>~0,5 h · TEC-01 (#23): definición del alcance<br>~0,5 h · HU-05 (#18): coordinación con Sosa<br>~1 h · Agile Enabler (Daily, PR #80, tablero y coordinación) |
| Buttini Cristobal | revisé y aprobé el PR #80 de HU-02, verificando el RED y el GREEN, la SDD, el BDD y la desviación documentada, con una observación sobre la comparación de nombres. | reviso el PR de HU-05 / ayudo con TEC-01. | ninguno. | ~1,5 h · Revisión del PR #80 |
| Cardozo Leandro | terminé HU-02 (#15): SDD aprobada, escenarios BDD, tests RED (`7ea611c`), código GREEN (`7fd2514`), test del contrato de `MemoryRepository.Update` con prueba de mutación (`710a14d`) y evidencia TDD; el PR #80 quedó mergeado y el Issue #15 cerrado. | tildo los criterios del Issue #15, borro la rama de HU-02. | ninguno. | ~2 h · HU-02 (#15) |
| García Santiago | tildé los criterios del Issue #17. | ayude a revisar TEC-01 (#23) | — | ~1h Revisión TEC-01 (#23) |
| Sosa Ricardo | preparé la SDD de HU-05 (#18) para guardarla. | hago HU-05 (#18): rama, SDD, escenarios BDD, tests RED, código GREEN y evidencia, y subo la rama para revisión. | ninguno (HU-04 y HU-06 ya están en `main`). | 2 h · HU-05 (#18) |

## Daily · 15/10/2026

| Integrante | Ayer | Hoy | Bloqueos | Horas (14/10) |
|---|---|---|---|---|
| Hernán | terminé TEC-01 (#23) con TDD: tests RED (`ffc4999`), servidor HTTP con `net/http` GREEN (`07d50b5`), test del id no numérico en las demás rutas con prueba de mutación (`e043603`, desviación documentada), prueba manual con `curl`, evidencia, sección "Cómo ejecutar" del README y correcciones de la revisión de Cardozo (`b7d82a4`); el PR #81 quedó mergeado y el Issue #23 cerrado. Hice el #73 (tests del espacio no separable en HU-01, con prueba de mutación); el PR #82 quedó mergeado. Preparé los borradores de las actas de Review (#64) y Retro (#65). Revisé la rama de HU-05 de Sosa sobre Linux (RED, GREEN, `-race`, mutación del mutex, cobertura y SDD), abrí el PR #83 y actualicé el Issue #18. | Review del Sprint 1 (#64) y Retrospectiva (#65); completo y subo las actas, consolido las Dailys (#62) y creo el tag `sprint-1`. Preparo el Planning del Sprint 2 (16/10). | ninguno. | ~3 h · TEC-01 (#23)<br>~0,5 h · #73 (HU-01)<br>~2 h · HU-05 (#18): revisión de la rama, PR e Issue<br>~1,5 h · Agile Enabler (borradores de Review y Retro, Daily y coordinación) |
| Buttini Cristobal | No revisé el PR #83 de HU-05 porque la rama todavía no estaba subida al repositorio. | Reviso el PR #83 de HU-05 (verificando el RED y el GREEN) y lo apruebo si todo está bien; participo de la Review y la Retro. | ninguno. | 0 h |
| Cardozo Leandro | revisé y aprobé el PR #81 de TEC-01, con cinco observaciones (tres corregidas en el mismo PR y una para el Sprint 2), y volví a aprobar después de las correcciones. Revisé y aprobé el PR #82 del #73. | participo de la Review y la Retro. | ninguno. | 0 h · Revisión de los PR #81 y #82 |
| García Santiago | revisé  el PR #81 de TEC-01 | participo de la Review y la Retro. | ninguno. | 1h · revisión del PR #81 |
| Sosa Ricardo | hice HU-05 (#18) con la SDD aprobada: SDD (`d121221`), escenarios BDD y tests RED (`a9afebc`), y entre ayer a la noche y esta madrugada el código GREEN (`34948cd`), la evidencia TDD (`ca84cd5`) y la verificación con `-race` (`3adfaa6`). Sin desviaciones del ciclo TDD; el PR #83 está abierto para revisión. | sigo el PR #83 hasta el merge, tildo los criterios del Issue #18 y participo de la Review y la Retro. | en mi máquina `go test -race` no corre por falta de cgo; la verificación la hicieron Hernán sobre Linux y el CI. | 2 hs · HU-05 (#18) |
