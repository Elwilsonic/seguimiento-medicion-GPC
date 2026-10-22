# Acta de Review — Sprint 2 (La Interfaz)

> Issue del repositorio: #88

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 2 — La Interfaz
- **Período del Sprint:** 16/10/2026 al 22/10/2026
- **Fecha de la Review:** 22/10/2026
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

## Resultado del Sprint Goal

**Cumplido.** Las siete historias comprometidas (HU-07 a HU-13) y las dos tareas técnicas (TEC-04 y TEC-05) están integradas en `main`. La aplicación se usa desde el navegador en `/` (TEC-04, PR #106): se crean, modifican y consultan proyectos, se carga, edita, prioriza y cambia de estado su Product Backlog, y se crean sprints, se les asignan ítems, se completan y se cierran. Con `STORAGE=postgres`, los datos se guardan en PostgreSQL 17 en Docker (TEC-05, PR #96). Cada Pull Request pasó el check de CI "gofmt y tests", que desde el PR #93 (16/10) también corre `go vet ./...` y desde el PR #98 (19/10) muestra la cobertura total.

## Demo

Se mostró la aplicación corriendo desde `main` (`79087f9`), siguiendo la sección "Interfaz web" del README:

1. Levantar PostgreSQL (`docker compose up -d --wait`) y el servidor con `STORAGE=postgres` (TEC-05).
2. Abrir <http://localhost:8080/> (TEC-04).
3. Crear un proyecto y modificar su estado (HU-01 y HU-02); mostrar el error de nombre repetido.
4. Cargar tres ítems con distintas prioridades (HU-04), listarlos por ID y por prioridad (HU-06 y HU-09); mostrar el error de un ítem sin criterios.
5. Editar un ítem con horas estimadas (HU-05 y HU-07).
6. Crear un sprint (HU-10); mostrar el error de un segundo sprint abierto.
7. Asignar ítems al sprint (HU-11), pasar uno a `EN_CURSO` (HU-08) y completarlo (HU-12); mostrar el error de completar un ítem `PENDIENTE`.
8. Ver los ítems completados y cerrar el sprint (HU-12 y HU-13).
9. Reiniciar el servidor y volver a cargar el proyecto por su ID, para mostrar que los datos siguen en PostgreSQL.

La demo la hizo Peñalbé Hernán.

## Historias y tareas comprometidas

Las fechas de merge están en hora de Argentina.

| ID | Issue | Story Points | Responsable | Pull Request | Aprobado por | Merge | ¿Cumple la DoD? |
|---|---|---|---|---|---|---|---|
| HU-10 | #24 | 5 | Buttini Cristobal | #94 | Peñalbé Hernán | 17/10 | ✅ Sí |
| HU-07 | #19 | 3 | Sosa Ricardo | #95 | García Santiago | 17/10 | ✅ Sí |
| TEC-05 | #35 | — | Cardozo Leandro | #96 | Buttini Cristobal | 18/10 | ✅ Sí, con desviaciones del ciclo TDD documentadas |
| HU-11 | #25 | 5 | Buttini Cristobal | #97 | Cardozo Leandro | 18/10 | ✅ Sí, con una desviación del ciclo TDD documentada |
| HU-08 | #20 | 3 | Sosa Ricardo | #99 | Cardozo Leandro | 19/10 | ✅ Sí |
| HU-12 | #26 | 3 | García Santiago | #100 | Peñalbé Hernán | 20/10 | ✅ Sí, con una desviación del ciclo TDD documentada |
| HU-13 | #27 | 3 | García Santiago | #101 | Cardozo Leandro | 20/10 | ✅ Sí |
| HU-09 | #22 | 3 | Cardozo Leandro | #102 | Buttini Cristobal | 20/10 | ✅ Sí |
| TEC-04 | #34 | — | Peñalbé Hernán | #104, #105 y #106 | Cardozo Leandro (#104), García Santiago (#105) y Buttini Cristobal (#106) | 20/10 al 21/10 | ✅ Sí, con una desviación del ciclo TDD documentada |

Los nueve Issues se cerraron con `Closes #N` al hacer merge de su Pull Request; ninguno se cerró a mano.

Desviaciones del ciclo TDD, todas documentadas en la evidencia de cada ítem:

- **HU-11:** en el GREEN se corrigieron dos tests que comparaban `Item` con `!=`, que no compila; en el RED ese error quedaba tapado por los demás errores de compilación.
- **HU-12:** en el GREEN se corrigieron dos expresiones de pasos BDD que venían del RED y no se habían podido ejecutar.
- **TEC-05:** el lock de `postgrestest.Pool` se escribió en el GREEN sin un test propio, y tres tests se escribieron después del GREEN a partir de pruebas de mutación.
- **TEC-04 (PR #105):** el subtest "fecha de fin con otro formato" se escribió después del GREEN, porque una prueba de mutación mostró que faltaba.

HU-08 y HU-09 registran ajustes antes del commit RED (formato, un test reforzado en la revisión) y un cambio aprobado a la SDD de HU-08, sin desviaciones de código.

Además entraron por Pull Request las acciones de proceso de la Retrospectiva del Sprint 1: el acta de Planning (#90), las actas organizadas por sprint (#91), las plantillas de Issue y de Pull Request y la Definition of Ready (#92), `go vet` en el CI (#93), la cobertura total en el CI (#98) y correcciones de documentación (#103).

## Velocidad

| | Story Points |
|---|---|
| Comprometidos | 25 |
| Completados | 25 |
| **Velocidad del Sprint 2** | **25** |

La velocidad subió de 21 (Sprint 1) a 25 Story Points, con 7 historias integradas (6 en el Sprint 1, contando la opcional HU-05) y las dos tareas técnicas más pesadas del plan. HU-09, la primera candidata a salir si había atraso, entró completa. Las tareas técnicas no se estiman en Story Points.

## Historias opcionales

| ID | Issue | Estado | Destino |
|---|---|---|---|
| HU-15 · Registrar integrante | #29 | No se empezó | Sprint 3 |
| HU-16 · Modificar integrante | #30 | No se empezó | Sprint 3 |
| HU-17 · Consultar integrantes | #31 | No se empezó | Sprint 3 |

Ninguna historia opcional entró en el Sprint: el equipo priorizó cerrar el alcance comprometido (25 Story Points, por encima de la velocidad del Sprint 1, más TEC-04 y TEC-05), como preveía el acta de Planning.

## Calidad al cierre

Medido sobre `main` (`79087f9`, con TEC-04 completa) con Go 1.27.1 en Linux, sin base de datos (como el CI):

```
$ go test -race -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 37.8% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 82.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	coverage: 77.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
$ go tool cover -func=cover.out | tail -n 1
total:									(statements)		79.7%
```

- Sin base, los tests de PostgreSQL hacen `t.Skip`, y eso baja la cobertura de `internal/backlog`, `internal/project` e `internal/sprint`. La última medición con base (HU-11) dio 92.5% de total con `-coverpkg=./...`; la de HU-13, que agrega el `Update` de sprints en PostgreSQL, queda pendiente (consta en `docs/reportes/cobertura.md`).
- `internal/web` figura en 0.0% porque no tiene tests propios: su `Handler` lo cubren los tests de `internal/httpapi` (100.0% con `-coverpkg`).
- `cmd/server` no llega al 100% porque `main()` no se prueba con tests unitarios.
- Escenarios BDD: 13 archivos `.feature` con 155 definiciones (132 escenarios y 23 esquemas del escenario), que godog ejecuta como 194 escenarios con `go test ./...` (definiciones contadas con `grep -cE '^\s*Escenario:'` y `grep -cE '^\s*Esquema del escenario:'` sobre `features/`; ejecuciones sumando las líneas «N scenarios» de `go test -v ./internal/...` sobre `main`, `79087f9`).

El detalle está en `docs/reportes/cobertura.md`.

## Observaciones del Product Architect

Los profesores no participaron de la Review, así que no hubo observaciones.

## Ritmo frente a la entrega

- **Pendiente hasta la entrega (05/11/2026):** 27 historias (HU-14 a HU-40) en 2 sprints de una semana, unas 13 o 14 por sprint, contra 7 integradas en este Sprint (25 Story Points). Además quedan abiertas cuatro tareas técnicas: TEC-03 (métricas y sus fórmulas, #33), TEC-06 (documentación técnica, #69), TEC-07 (manual de usuario, #70) y TEC-08 (informe final de métricas y cobertura, #71).
- La decisión 8 del acta de Planning del Sprint 1 dice que se recorta alcance, no pruebas ni documentación.

**Decisión:** se mantiene el criterio de la decisión 8 del Sprint 1. En el Planning del Sprint 3, por cada requerimiento mínimo de la consigna que todavía falta (estimación con Story Points y Planning Poker, registro de esfuerzo, gestión de defectos, métricas, dashboard y reportes), se eligen las historias imprescindibles para cubrirlo; el resto pasa a "si hay tiempo". Las tareas técnicas de documentación (TEC-06, TEC-07 y TEC-08) no se recortan, porque son entregables finales.

## Cambios al Product Backlog

- HU-15, HU-16 y HU-17 (opcionales) no se empezaron: pasan al Sprint 3, sujetas al recorte de alcance del Planning.
- HU-14, HU-18 y HU-19 pasan al Sprint 3, como preveía el acta de Planning.
- La interfaz elige el proyecto y el sprint por su ID porque no hay listados de proyectos ni de sprints; HU-14 (consultar sprints anteriores) cubre los sprints. No se agrega una historia para listar proyectos, porque la consigna no lo pide: la interfaz sigue eligiendo el proyecto por su ID.
- Detalles de Issues ya cerrados, para corregir en el Sprint 3: los criterios de #25 (HU-11) y #86 (Planning) siguen sin tildar; #86 apunta a la ruta vieja `docs/actas/sprint-2-planning.md`; la "## Trazabilidad" de #22 (HU-09) y #35 (TEC-05) no sigue el formato de `AGENTS.md`.

## Línea base

Al integrar esta acta se crea el tag `sprint-2` sobre `main`.

## Próximo paso

Retrospectiva del Sprint 2 (#89) y Planning del Sprint 3 (23/10/2026), con su Issue creado antes de la ceremonia.
