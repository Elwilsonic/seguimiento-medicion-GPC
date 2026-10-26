# Acta de Review — Sprint 3 (La Estimación)

> Issue del repositorio: #112

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 3 — La Estimación
- **Período del Sprint:** 23/10/2026 al 26/10/2026 (planificado hasta el 29/10/2026; ver "Cierre anticipado")
- **Fecha de la Review:** 26/10/2026
- **Sprint Goal:** Estimar ítems con Planning Poker y registrar el esfuerzo y los defectos de cada sprint, con las fórmulas de las métricas definidas para el Sprint 4.

## Participantes

| Rol | Integrante | Asistió |
|---|---|---|
| Product Architect | Profesores de la cátedra | No |
| Agile Enabler | Peñalbé Hernán | Sí |
| Product Builder | Buttini Cristobal | Sí |
| Product Builder | Cardozo Leandro | Sí |
| Product Builder | García Santiago | Sí |
| Product Builder | Sosa Ricardo | Sí |

## Cierre anticipado

El 24/10 el equipo replanificó: la entrega pasa al 02/11/2026 y el desarrollo se cierra el 30/10 (consta en la SDD de HU-32, "Comparación con el acta de Planning"). Con las nueve historias comprometidas y TEC-03 integradas en `main` el 26/10, el equipo decidió cerrar el Sprint 3 ese día y arrancar el Sprint 4 el 27/10. Las fechas del Sprint 4 se fijan en su Planning.

## Resultado del Sprint Goal

**Cumplido.** Las nueve historias comprometidas (HU-18, HU-20 a HU-23, HU-25 a HU-27 y HU-29, 29 Story Points) y TEC-03 están integradas en `main`. Además se adelantaron del Sprint 4 HU-32 y HU-34, las dos primeras métricas.

- **Estimación (Planning Poker):** se asignan Story Points de la escala 1, 2, 3, 5, 8, 13 y 21 a un ítem (HU-20), se abre una sesión de Planning Poker para un ítem (HU-21), cada integrante vota sin ver los demás votos (HU-22), se revelan todos a la vez (HU-23) y la estimación acordada se guarda en el ítem y cierra la sesión (HU-25).
- **Registro de esfuerzo:** se registran horas por integrante, actividad e historia de un sprint abierto (HU-26) y se consultan por historia o por sprint, con el total (HU-27).
- **Defectos:** se registra un defecto con severidad, historia y sprint de detección (HU-29).
- **Métricas:** `docs/metricas.md` define las métricas M1 a M10 con su fórmula, unidad, datos de entrada y caso de dato faltante (TEC-03); `GET /api/sprints/{id}/metricas` y `GET /api/proyectos/{id}/metricas` y la sección "Métricas" de la interfaz muestran los Story Points planificados y completados, la velocidad (HU-32) y el porcentaje de historias completadas (HU-34).
- **Proyecto:** se registra la fecha de inicio del proyecto (HU-18).

Cada historia suma su endpoint bajo `/api/…` y su sección en la interfaz web (tema 5 del Planning), y guarda sus datos en memoria y en PostgreSQL. Cada Pull Request pasó el check de CI "gofmt y tests".

## Demo

Se mostró el recorrido de abajo sobre `main` (`4806ad6`, con HU-18). Presentó Peñalbé Hernán. Asistieron todos los integrantes del equipo; los profesores no asistieron.

Recorrido sobre `main` (`4806ad6`), siguiendo la sección "Interfaz web" del README:

1. Levantar PostgreSQL (`docker compose up -d --wait`) y el servidor con `STORAGE=postgres`.
2. Crear un proyecto y registrar su fecha de inicio (HU-18).
3. Cargar ítems y asignarles Story Points con "Asignar SP" (HU-20); mostrar el error de un valor fuera de la escala.
4. Abrir una sesión de Planning Poker para un ítem (HU-21), votar con dos integrantes (HU-22), revelar los votos (HU-23) y registrar la estimación acordada (HU-25); mostrar el valor en la columna SP.
5. Crear un sprint, asignarle ítems y registrar esfuerzo (HU-26); consultar los esfuerzos del sprint y de una historia, con el total (HU-27).
6. Registrar un defecto (HU-29); mostrar el error de un sprint cerrado.
7. Completar un ítem y ver las métricas del sprint y del proyecto: Story Points planificados y completados, velocidad y porcentaje de historias completadas (HU-32 y HU-34).

## Historias y tareas comprometidas

Las fechas de merge están en hora de Argentina.

| ID | Issue | Story Points | Responsable | Pull Request | Aprobado por | Merge | ¿Cumple la DoD? |
|---|---|---|---|---|---|---|---|
| TEC-03 | #33 | — | Peñalbé Hernán | #117 | Buttini Cristobal | 23/10 | ✅ Sí |
| HU-20 | #38 | 2 | Cardozo Leandro | #118 | García Santiago | 23/10 | ✅ Sí |
| HU-26 | #44 | 5 | García Santiago | #120 y #121 | Buttini Cristobal | 24/10 | ✅ Sí, con una desviación del ciclo TDD documentada |
| HU-21 | #39 | 5 | Buttini Cristobal | #122 | Peñalbé Hernán | 24/10 | ✅ Sí, con una desviación del ciclo TDD documentada |
| HU-29 | #47 | 5 | Sosa Ricardo | #124 | Peñalbé Hernán | 25/10 | ✅ Sí, con desviaciones del ciclo TDD documentadas |
| HU-27 | #45 | 3 | García Santiago | #126 | Cardozo Leandro | 25/10 | ⚠️ Código sí; la "## Trazabilidad" del Issue no sigue el formato de `AGENTS.md` |
| HU-22 | #40 | 3 | Buttini Cristobal | #125 | Cardozo Leandro | 25/10 | ✅ Sí, con tests agregados después del GREEN por la revisión |
| HU-23 | #41 | 2 | Cardozo Leandro | #127 | Sosa Ricardo | 25/10 | ✅ Sí |
| HU-25 | #43 | 2 | Cardozo Leandro | #130 | García Santiago | 26/10 | ✅ Sí; la "## Trazabilidad" del Issue ya sigue el formato de `AGENTS.md` y los diez criterios están tildados |
| HU-18 | #36 | 2 | Sosa Ricardo | #131 | García Santiago | 26/10 | ⚠️ Código sí; los criterios del Issue quedaron sin tildar |

Adelantadas del Sprint 4:

| ID | Issue | Story Points | Responsable | Pull Request | Aprobado por | Merge | ¿Cumple la DoD? |
|---|---|---|---|---|---|---|---|
| HU-32 | #50 | 3 | Peñalbé Hernán | #123 | García Santiago | 25/10 | ✅ Sí |
| HU-34 | #52 | 2 | Peñalbé Hernán | #129 | García Santiago | 26/10 | ⚠️ Código sí; los criterios del Issue quedaron sin tildar |

Los doce Issues se cerraron al hacer merge de su Pull Request; ninguno se cerró a mano.

Desviaciones del ciclo TDD, según la evidencia de cada ítem:

- **HU-21:** el commit RED ya conectaba el repositorio de PostgreSQL en `cmd/server` (wiring adelantado), así que `TestNewRepositories_PostgresUsaElRepositorioDePoker` pasaba en el RED con base.
- **HU-26:** una aserción de `TestPostgresRepository_GuardaElEsfuerzo` (zona horaria de la fecha) quedaba tapada por el stub en el RED y falló recién en el GREEN; era un error del test y se corrigió en `bed08f5`, sin tocar producción.
- **HU-29:** el GREEN de la API cambió archivos de test por la nueva firma de `NewHandler` (sin cambiar aserciones), y la persistencia en PostgreSQL entró en un solo commit, sin un RED propio.
- **HU-22:** después del GREEN se agregaron tests en `e96dd8d` (formulario de votación) y en `7647967` (409 por sesión no abierta, falla de guardado y su 500), pedidos en la revisión. La evidencia los incluye en la trazabilidad, pero no tiene una sección de desviaciones que lo diga.

HU-34 registra una rama defensiva de los handlers de métricas sin test (segunda lectura del backlog), que la consulta única de HU-35 elimina. HU-18, HU-20, HU-23, HU-25, HU-27 y HU-32 no registran desviaciones.

Además entraron por Pull Request el acta de Planning (#114), el Sprint Goal en el README (#116) y el checklist para ponerse al día antes de una historia, con el "Paso 0" de `AGENTS.md` (#119).

## Velocidad

| | Story Points |
|---|---|
| Comprometidos | 29 |
| Completados del compromiso | 29 |
| Adelantados del Sprint 4 (HU-32 y HU-34) | 5 |
| **Velocidad del Sprint 3** | **34** |

La velocidad subió de 21 (Sprint 1) y 25 (Sprint 2) a 34 Story Points, con 11 historias integradas en cuatro días. HU-18 y HU-29, las primeras candidatas a salir si había atraso, entraron completas. Los Story Points de HU-32 (3) y HU-34 (2) no salieron de un Planning Poker: los estimó el Agile Enabler en sus SDD y se confirman en el Planning del Sprint 4. Las tareas técnicas no se estiman en Story Points.

## Historias "si hay tiempo"

| ID | Issue | Estado | Destino |
|---|---|---|---|
| HU-41 · Registrar la fecha de fin del proyecto | #115 | No se empezó | Sprint 4, si hay tiempo |
| HU-24 · Detectar diferencias y repetir ronda | #42 | No se empezó | Sprint 4, si hay tiempo |
| HU-14 · Consultar sprints anteriores | #28 | No se empezó | Sprint 4, si hay tiempo |
| HU-30 · Cambiar el estado de un defecto | #48 | No se empezó | Sprint 4 (la necesita HU-47) |

Con la capacidad sobrante el equipo priorizó las métricas del Sprint 4 (HU-32 y HU-34, y la SDD de HU-35), porque el dashboard y los reportes dependen de ellas.

## Calidad al cierre

Medido sobre `main` (`4806ad6`, con HU-18) con Go 1.27.1 en Linux, sin base de datos (como el CI):

```
$ go test -race -count=1 -coverprofile=cover.out ./...
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 31.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/defect	coverage: 76.6% of statements
ok  	seguimiento-medicion-gpc/internal/effort	coverage: 75.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 97.6% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	coverage: 69.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 67.5% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	coverage: 76.3% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
$ go tool cover -func=cover.out | tail -n 1
total:									(statements)			80.9%
```

Con PostgreSQL 16.15 y `TEST_DATABASE_URL` (base recreada desde `db/init/01-schema.sql`), sin ningún `SKIP`:

```
$ go test -race -count=1 -p 1 -coverpkg=./... -coverprofile=cover.out ./...
$ go tool cover -func=cover.out | tail -n 1
total:									(statements)			94.8%
```

- Sin base, los tests de PostgreSQL hacen `t.Skip`, y eso baja la cobertura de los paquetes con repositorio (`backlog`, `defect`, `effort`, `poker`, `project` y `sprint`).
- `internal/web` figura en 0.0% porque no tiene tests propios: lo cubren los tests de `internal/httpapi`.
- `cmd/server` no llega al 100% porque `main()` no se prueba con tests unitarios.
- Escenarios BDD: 24 archivos `.feature` con 279 definiciones (244 escenarios y 35 esquemas del escenario), que godog ejecuta como 346 escenarios con `go test ./...` (definiciones contadas con `grep -cE '^\s*Escenario:'` y `grep -cE '^\s*Esquema del escenario:'` sobre `features/`; ejecuciones sumando las líneas «N scenarios» de `go test -v ./internal/...` sobre `4806ad6`). En el Sprint 2 eran 13 archivos y 194 escenarios ejecutados.
- `go vet ./...` sin salida.

El detalle por historia está en `docs/reportes/cobertura.md`.

## Observaciones del Product Architect

Los profesores de la cátedra no asistieron a la Review, así que no hay observaciones del Product Architect para este Sprint. Se registran en el acta cuando las envíen.

## Ritmo frente a la entrega

- **Entrega:** 02/11/2026, con el desarrollo cerrado el 30/10 (replanificación del 24/10).
- **Pendiente de los requerimientos mínimos:** gestión de defectos (HU-30 y HU-31), métricas (HU-33, HU-35 y HU-47), dashboard (HU-36 y HU-37) y reportes (HU-38, HU-39 y HU-40), más TEC-06 (documentación técnica, #69), TEC-07 (manual de usuario, #70) y TEC-08 (informe final de métricas y cobertura, #71). Todos tienen responsable asignado en su Issue.
- Estimación, registro de esfuerzo y la base de las métricas quedaron cubiertos en este Sprint.

**Decisión:** se mantiene el criterio de recorte de la Review del Sprint 2: se recorta alcance, no pruebas ni documentación. El Sprint 4 toma primero lo que bloquea a otros (HU-30 → HU-47, HU-35 → HU-36 y HU-37, HU-38 y HU-39 → HU-40) y TEC-06 a TEC-08 no se recortan.

## Cambios al Product Backlog

- **División de historias (Planning del Sprint 3):** se aplicó en HU-18 (HU-41, #115) y en HU-34 (HU-47, #128). HU-32 se hizo entera, con planificados, completados y velocidad, porque la replanificación del 24/10 dejó la división en pausa (SDD de HU-32); HU-43 y HU-44 no se crearon. Las divisiones de HU-31, HU-33, HU-16 y HU-37 siguen pendientes. Confirmación de los profesores: pendiente, porque no asistieron a la Review.
- HU-41, HU-24 y HU-14 siguen en "si hay tiempo"; HU-30 pasa al Sprint 4 como imprescindible, porque la necesita HU-47.
- Detalles de Issues ya cerrados, para corregir: los criterios de #36 (HU-18) y #52 (HU-34) siguen sin tildar, y la "## Trazabilidad" de #45 (HU-27) no sigue el formato de `AGENTS.md`. La de #43 (HU-25) ya se llevó al formato.

## Línea base

Al integrar esta acta se crea el tag `sprint-3` sobre `main`. El tag queda sobre el merge commit de este Pull Request, que solo agrega documentación; el código es el mismo de `4806ad6`, donde se midió la calidad al cierre.

## Próximo paso

Retrospectiva del Sprint 3 (#113) y Planning del Sprint 4 (27/10/2026), con su Issue creado antes de la ceremonia.
