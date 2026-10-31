# Acta de Review — Sprint 4 (La Entrega)

> Issue del repositorio: #135

Los datos salen del repositorio y de GitHub al 30/10/2026.

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 4 — La Entrega
- **Período del Sprint:** 27/10/2026 al 02/11/2026 (desarrollo cerrado el 30/10, entrega el 02/11/2026)
- **Fecha de la Review:** 30/10/2026
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

## Resultado del Sprint Goal

**Cumplido.** Las diez historias comprometidas (HU-30, HU-31, HU-33, HU-47, HU-35, HU-36, HU-37, HU-38, HU-39 y HU-40, 35 Story Points) están integradas en `main`. Con la capacidad sobrante entraron además las tres historias "si hay tiempo" (HU-41, HU-24 y HU-14) y las que figuraban fuera de la entrega (E4: HU-15 a HU-17, HU-19 y HU-28): 24 Story Points más. **Las 42 historias del Product Backlog están cerradas.**

- **Defectos:** se cambia el estado de un defecto de a un paso, con sprint de resolución al resolverlo (HU-30), y se consultan los defectos de un proyecto o de un sprint, con filtro por estado (HU-31).
- **Métricas:** horas estimadas, horas reales y desviación (HU-33), defectos detectados y resueltos (HU-47), y las diez métricas en una sola consulta (HU-35).
- **Dashboard:** resumen del estado del proyecto (HU-36) y tres gráficos SVG por sprint, sin librerías (HU-37).
- **Reportes:** reporte del proyecto (HU-38), reporte de un sprint (HU-39) y exportación del reporte del proyecto a PDF (HU-40).
- **Además:** fecha de fin del proyecto (HU-41) y modificar las dos fechas juntas (HU-19), repetir la ronda de Planning Poker (HU-24), consultar los sprints del proyecto (HU-14), registrar, modificar y consultar integrantes (HU-15 a HU-17) y corregir las horas de un esfuerzo (HU-28).
- **Documentación final:** TEC-06 (documentación técnica) y TEC-07 (manual de usuario, con 19 capturas) están en `main`. TEC-08 (informe final) queda para antes de la entrega, como prevé el tema 1 del Planning.

## Requerimientos mínimos de la consigna

Todos quedan cubiertos en `main`. No queda ningún requerimiento mínimo fuera de la entrega.

| Requerimiento mínimo | Historias | Estado |
|---|---|---|
| Gestión de proyectos | HU-01 a HU-03, HU-18, HU-19 y HU-41 | ✅ |
| Product Backlog | HU-04 a HU-09 | ✅ |
| Gestión de Sprints | HU-10 a HU-14 | ✅ |
| Estimación (Story Points y Planning Poker) | HU-20 a HU-25 | ✅ |
| Registro de esfuerzo | HU-26 a HU-28 | ✅ |
| Gestión de defectos | HU-29 a HU-31 | ✅ |
| Métricas | TEC-03, HU-32 a HU-35 y HU-47 | ✅ |
| Dashboard | HU-36 y HU-37 | ✅ |
| Reportes | HU-38 a HU-40 | ✅ |

Además está cubierto el registro de integrantes (E4, HU-15 a HU-17), que no es un requerimiento mínimo. Queda fuera de la entrega solo la mejora del listado de proyectos (ver "Cambios al Product Backlog"). De los entregables finales, TEC-08 (informe final) se completa antes del 02/11.

## Demo

El equipo presentó el recorrido del [manual de usuario](../../manual-de-usuario.md), que cubre todas las épicas sobre `main` (`b2a5da6`):

1. Levantar el servidor y crear un proyecto con sus fechas de inicio y fin; modificarlas juntas (HU-18, HU-41 y HU-19).
2. Registrar integrantes, modificar el rol de uno y listarlos (HU-15 a HU-17).
3. Cargar el backlog, estimar un ítem con Planning Poker con diferencias, repetir la ronda y registrar la estimación (HU-20 a HU-25).
4. Crear un sprint, asignar y completar ítems, registrar y corregir esfuerzo (HU-10 a HU-12, HU-26 a HU-28).
5. Registrar un defecto, pasarlo a `EN_CORRECCION` y `RESUELTO`, y consultar los defectos (HU-29 a HU-31).
6. Cerrar el sprint, abrir otro y ver los sprints del proyecto (HU-13 y HU-14).
7. Ver las métricas, el resumen y los gráficos del dashboard (HU-32 a HU-37 y HU-47).
8. Generar el reporte del proyecto y del sprint, y descargar el PDF (HU-38 a HU-40).

## Historias y tareas comprometidas

Las fechas de merge están en hora de Argentina. "¿Cumple la DoD?" considera también el Issue (criterios tildados y "## Trazabilidad"), según la acción 1 de la Retrospectiva del Sprint 3.

| ID | Issue | Story Points | Responsable | Pull Request | Aprobado por | Merge | ¿Cumple la DoD? |
|---|---|---|---|---|---|---|---|
| HU-30 | #48 | 3 | Sosa Ricardo | #140 | Cardozo Leandro | 27/10 | ⚠️ Código sí; los criterios del Issue quedaron sin tildar |
| HU-31 | #49 | 3 | Sosa Ricardo | #141 | Peñalbé Hernán | 28/10 | ✅ Sí, con cambios pedidos en la revisión |
| HU-33 | #51 | 3 | García Santiago | #142 | Cardozo Leandro | 28/10 | ✅ Sí, con cambios pedidos en la revisión y dos tests después del GREEN, registrados como desviación |
| HU-36 | #54 | 3 | Buttini Cristobal | #143 | García Santiago | 28/10 | ✅ Sí, con una desviación del ciclo TDD documentada |
| HU-47 | #128 | 2 | Peñalbé Hernán | #144 | Buttini Cristobal | 28/10 | ✅ Sí |
| HU-35 | #53 | 3 | Cardozo Leandro | #145 | Peñalbé Hernán | 28/10 | ✅ Sí, con cambios pedidos en la revisión |
| HU-37 | #55 | 5 | Buttini Cristobal | #147 | Cardozo Leandro | 29/10 | ✅ Sí; un defecto de la interfaz se corrigió después con TDD (#162, ver "Cambios al Product Backlog") |
| HU-38 | #56 | 5 | Cardozo Leandro | #148 | García Santiago | 29/10 | ✅ Sí |
| HU-39 | #57 | 3 | García Santiago | #149 | Cardozo Leandro | 29/10 | ✅ Sí |
| HU-40 | #58 | 5 | Peñalbé Hernán | #151 | Sosa Ricardo | 29/10 | ⚠️ Código sí; los criterios del Issue siguen como "(borrador)" y sin tildar |
| TEC-06 | #69 | — | Peñalbé Hernán | #139, #146, #153 y #160 | Cardozo Leandro (#139 y #146) y Buttini Cristobal (#153 y #160) | 27/10 al 30/10 | ⚠️ El documento está completo en `main`; falta tildar los criterios y cerrar el Issue, con el PR de TEC-08 |
| TEC-07 | #70 | — | Peñalbé Hernán | #161 | Cardozo Leandro | 30/10 | ✅ Sí |
| TEC-08 | #71 | — | Peñalbé Hernán | — | — | — | Pendiente: se entrega antes del 02/11 (tema 1 del Planning) |

Historias "si hay tiempo" y fuera de la entrega, tomadas con capacidad sobrante:

| ID | Issue | Story Points | Responsable | Pull Request | Aprobado por | Merge | ¿Cumple la DoD? |
|---|---|---|---|---|---|---|---|
| HU-41 | #115 | 2 | Sosa Ricardo | #150 | Peñalbé Hernán | 29/10 | ✅ Sí |
| HU-24 | #42 | 3 | Cardozo Leandro | #152 | Sosa Ricardo | 29/10 | ✅ Sí |
| HU-14 | #28 | 3 | Buttini Cristobal | #154 | Peñalbé Hernán | 29/10 | ⚠️ Código sí; los criterios del Issue quedaron sin tildar |
| HU-15 | #29 | 5 | Peñalbé Hernán | #155 | Cardozo Leandro | 29/10 | ✅ Sí |
| HU-19 | #37 | 2 | Sosa Ricardo | #156 | Peñalbé Hernán | 30/10 | ✅ Sí |
| HU-28 | #46 | 3 | García Santiago | #157 | Sosa Ricardo | 30/10 | ⚠️ Código sí; los criterios del Issue quedaron sin tildar. Un test después del GREEN, registrado como desviación |
| HU-16 | #30 | 3 | Cardozo Leandro | #158 | Peñalbé Hernán | 30/10 | ✅ Sí |
| HU-17 | #31 | 3 | Buttini Cristobal | #159 | Peñalbé Hernán | 30/10 | ⚠️ Código sí; los criterios del Issue quedaron sin tildar |

Los Issues de las 18 historias se cerraron al hacer merge de su Pull Request; ninguno se cerró a mano. Todos tienen la "## Trazabilidad".

Desviaciones del ciclo TDD, según la evidencia de cada ítem:

- **HU-15 y HU-40:** dos tests del RED se corrigieron antes del GREEN, en un commit propio (`058804f` y `a317814`): eran errores de los tests, no cambios de comportamiento. En HU-15 el armado de `cmd/server` entró en el RED, como pedía la SDD; en HU-40 queda una rama sin test en el handler del PDF.
- **HU-33 y HU-28:** tests agregados después del GREEN por la revisión, registrados como desviación (acción 2 de la Retrospectiva del Sprint 3).
- **HU-36:** pasos que pasaban en falso en el primer RED, corregidos antes del commit.
- **HU-16:** `nameTaken` entró en el GREEN y no en un refactor, y los casos comunes van en las tablas de `server_test.go`.
- **HU-37:** el dibujo de los gráficos no tiene tests automáticos (los tests de Go no ejecutan JavaScript); se cubre con tests de contenido y pruebas en el navegador.

Además entraron por Pull Request el acta de Planning (#137 y #138) y el acta de las Dailys (#163, que cerró el Issue #134).

## Velocidad

| | Story Points |
|---|---|
| Comprometidos | 35 |
| Completados del compromiso | 35 |
| Historias "si hay tiempo" (HU-41, HU-24 y HU-14) | 8 |
| Historias fuera de la entrega (HU-15 a HU-17, HU-19 y HU-28) | 16 |
| **Velocidad del Sprint 4** | **59** |

La velocidad fue de 21 (Sprint 1), 25 (Sprint 2), 34 (Sprint 3) y 59 Story Points en este Sprint, con 18 historias integradas en cuatro días de desarrollo. En total el proyecto completó 139 Story Points. Los Story Points de HU-28 (3) y HU-19 (2) los estimó el Agile Enabler al aprobar sus SDD; los de las demás historias salen de los Planning. El equipo confirmó en la reunión los Story Points de HU-28 y HU-19.

## Calidad al cierre

Medido sobre `main` (`b2a5da6`, con TEC-07) con Go 1.27.1 en Linux, sin base de datos (como el CI):

```
$ go test -race -count=1 -coverprofile=cover.out ./...
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 29.8% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/dashboard	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/defect	coverage: 77.7% of statements
ok  	seguimiento-medicion-gpc/internal/effort	coverage: 74.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 98.7% of statements
ok  	seguimiento-medicion-gpc/internal/member	coverage: 68.8% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	coverage: 75.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 74.8% of statements
ok  	seguimiento-medicion-gpc/internal/report	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	coverage: 77.1% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
$ go tool cover -func=cover.out | tail -n 1
total:									(statements)			84.5%
```

Con PostgreSQL 16.15 instalado en la máquina de la medición y `TEST_DATABASE_URL` (base recreada desde `db/init/01-schema.sql`), sin ningún `SKIP`:

```
$ go test -race -count=1 -p 1 -coverpkg=./... -coverprofile=cover.out ./...
$ go tool cover -func=cover.out | tail -n 1
total:									(statements)			96.2%
```

- La cobertura subió de 80.9% a 84.5% sin base y de 94.8% a 96.2% con base respecto del cierre del Sprint 3.
- Escenarios BDD: 42 archivos `.feature` con 486 definiciones (400 escenarios y 86 esquemas del escenario), que godog ejecuta como 656 escenarios (346 al cierre del Sprint 3).
- `go vet ./...` y `gofmt -l .` sin salida.
- `internal/web` figura en 0.0% porque no tiene tests propios: lo cubren los tests de contenido de `internal/httpapi`.

El detalle por historia está en `docs/reportes/cobertura.md`.

## Observaciones del Product Architect

Los profesores no asistieron a la Review ni enviaron observaciones.

## Cambios al Product Backlog

- **Product Backlog completo:** las 42 historias están cerradas, y también las once épicas (E1 a E11), cada una al terminar sus historias; la última fue E4 (#4), el 30/10.
- **HU-48 no se crea:** HU-16 cubre el nombre y el rol del integrante (opción A de su SDD, decisión D9), así que la división que proponía el acta del Sprint 3 no hace falta.
- **Historias fuera de la entrega:** E4 (HU-15 a HU-17), HU-19 y HU-28 se tomaron con capacidad sobrante; cada SDD registra la diferencia con las actas.
- **Defecto de HU-37:** al armar las capturas del manual se vio que la nota de desviación del gráfico de horas se cortaba. Se corrigió con TDD (RED `56fea15`, GREEN `1b2bbaa`, PR #162, aprobado por García Santiago), con `Refs #55`.
- **Mejora posible, fuera de la entrega:** un listado de proyectos en la interfaz (hoy el proyecto se elige por su ID, límite documentado en el manual). No hay una historia que lo pida; queda registrada como mejora posible.
- **Detalles de Issues ya cerrados, para corregir:** los criterios de #48 (HU-30), #58 (HU-40, además con "(borrador)"), #28 (HU-14), #31 (HU-17) y #46 (HU-28) siguen sin tildar.

## Historia agregada después de la Review (HU-49)

El 31/10, después de esta Review, al comparar el proyecto con la consigna se vio que el requerimiento 9 pide que el reporte incluya las historias planificadas y completadas y las estimaciones. HU-38 y HU-39 las informaban solo como cantidades y sumas de las métricas (decisión 9 de la SDD de HU-38), sin el detalle de cada historia.

Por decisión del Agile Enabler se creó **HU-49 "Ver las historias y sus estimaciones en el reporte"** (Issue #170, épica E11, prioridad MEDIA, 3 Story Points) y se sumó al Sprint 4, que termina con la entrega del 02/11:

- **El reporte del proyecto** lista cada historia del backlog con su prioridad, estado, Story Points, horas estimadas y sprint.
- **El reporte del sprint** lista las historias asignadas al sprint.
- **Los dos** muestran la cantidad de historias y de completadas, en la API, en la interfaz y, en el del proyecto, en el PDF.

**Responsable:** Peñalbé Hernán. La SDD, la evidencia TDD y el Pull Request están enlazados en el Issue #170.

**Cómo cambian los números de esta acta cuando HU-49 llegue a `main`:**

| Dato | Al cierre de la Review | Con HU-49 |
|---|---|---|
| Historias del Product Backlog cerradas | 42 | 43 |
| Velocidad del Sprint 4 | 59 | 62 |
| Story Points completados por el proyecto | 139 | 142 |
| Archivos `.feature` y escenarios BDD ejecutados | 42 y 656 | 43 y 665 |
| Cobertura sin base y con PostgreSQL (`-coverpkg=./...`) | 84.5% y 96.2% | 84.7% y 96.2% |

Los Story Points de HU-49 los estimó el Agile Enabler al crear la historia, comparándola con HU-39, que reutiliza la estructura de los reportes de HU-38. No pasaron por un Planning. El tag `sprint-4` se mueve sobre `main` después de integrar HU-49.

## Línea base

El tag `sprint-4` se crea sobre `main` después de integrar esta acta y TEC-08, para que la línea base de la entrega incluya el informe final.

## Próximo paso

- Retrospectiva del Sprint 4 (#136).
- TEC-08 (#71), informe final de métricas y cobertura, con el cierre de TEC-06 (#69), antes del 02/11.
- Preparar la presentación y la demo final de la entrega del 02/11.
