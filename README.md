# Software Metrics & Estimation

Aplicación para la estimación, planificación, seguimiento y calidad de proyectos de software.

Trabajo Práctico Integrador — Ingeniería y Calidad de Software (2026)
Universidad Tecnológica Nacional — Facultad Regional San Rafael

## Tecnologías y prácticas

- **Lenguaje:** Go (Golang)
- **Metodología:** Scrum
- **Prácticas:** SDD (Specification-Driven Development), BDD (Behavior-Driven Development), TDD (Test-Driven Development)
- **Tests:** paquete `testing` de Go (TDD) y godog para ejecutar los escenarios BDD
- **Control de versiones:** Git
- **Integración continua:** Cada Pull Request hacia `main` corre `gofmt`, `go vet ./...` y `go test -race ./...` con GitHub Actions, y muestra la cobertura total
- **Interfaz web:** HTML, JavaScript y CSS sin frameworks, embebidos en el binario con `embed` y servidos por el mismo servidor (TEC-04); ver [Interfaz web](#interfaz-web)
- **Persistencia:** en memoria (valor por defecto) o PostgreSQL 17 en Docker (TEC-05), con el driver `github.com/jackc/pgx/v5`; ver [Persistencia con PostgreSQL](#persistencia-con-postgresql)

## Requisitos

- Go 1.27.1 o superior
- Docker con Docker Compose, solo para usar PostgreSQL: con memoria no hace falta

## Cómo ejecutar

Desde la raíz del repositorio:

```
go run ./cmd/server
```

El servidor escucha en el puerto `8080`. Para usar otro, definí la variable de entorno `PORT` (en PowerShell: `$env:PORT="9090"; go run ./cmd/server`; en Linux o macOS: `PORT=9090 go run ./cmd/server`). Por defecto los datos se guardan en memoria y se pierden al detener el servidor; para conservarlos, usá PostgreSQL (ver [Persistencia con PostgreSQL](#persistencia-con-postgresql)).

### Interfaz web

Con el servidor en marcha, abrí <http://localhost:8080/> en el navegador (TEC-04). La interfaz es una sola página con seis secciones:

1. **Proyectos:** crear un proyecto (queda como "proyecto actual"), modificar su nombre, descripción y estado.
2. **Product Backlog:** crear ítems del proyecto actual, listarlos ordenados por ID o por prioridad y, en cada fila, editar el ítem (incluidas las horas estimadas), cambiar su estado, asignarle story points de la escala 1, 2, 3, 5, 8, 13 o 21 con "Asignar SP" (columna SP; "—" si no está estimado, HU-20), abrir una sesión de Planning Poker con "Abrir Poker" (HU-21), asignarlo al sprint actual o completarlo.
3. **Planning Poker:** muestra la última sesión abierta para un ítem del backlog, con su ID, ítem, estado, ronda y las cartas de la escala de HU-20 (HU-21), permite votar una carta indicando el nombre del integrante sin mostrar los votos (HU-22), con "Revelar votos", muestra a la vez la lista de votos (integrante y puntos) y deja la sesión REVELADA, sin más votos (HU-23) y, en "Registrar estimación acordada", con un botón por carta, guarda los story points acordados en el ítem, deja la sesión CERRADA y recarga el backlog para mostrar el valor en la columna SP (HU-25).
4. **Métricas:** con "Ver métricas", muestra los story points planificados y completados, los ítems sin estimar y la velocidad del proyecto actual ("sin datos" si no hay sprints cerrados) y, si hay sprint actual, los story points de ese sprint (HU-32).
5. **Sprints:** crear un sprint en el proyecto actual (queda como "sprint actual"), ver sus ítems completados, cerrarlo, registrar el esfuerzo de un integrante en una historia del sprint con el formulario "Registrar esfuerzo" (HU-26) y consultar los esfuerzos por sprint o por historia, con su tabla y total, con el formulario "Consultar esfuerzos" (HU-27).
6. **Defectos:** registrar un defecto de una historia del backlog con el formulario "Registrar defecto": ítem, descripción, severidad (CRITICA, ALTA, MEDIA o BAJA) y sprint de detección, que es obligatorio y tiene que estar abierto (HU-29).

Un recorrido típico: crear un proyecto, cargar sus ítems, abrir una sesión de Planning Poker para un ítem, crear un sprint, asignarle ítems, pasarlos a `EN_CURSO` con "Cambiar estado", registrar el esfuerzo con "Registrar esfuerzo", consultar los esfuerzos con "Consultar esfuerzos", completarlos, ver los completados y cerrar el sprint.

- **Proyecto y sprint actuales:** la aplicación todavía no tiene listados de proyectos ni de sprints (ninguna historia los pide), así que se eligen por su ID en los campos de arriba. Al crear un proyecto o un sprint, su ID se completa solo; para volver a un proyecto existente, escribí su ID y tocá "Cargar proyecto".
- **Errores:** si la aplicación rechaza un dato (por ejemplo, un nombre de proyecto repetido o un ítem sin criterios de aceptación), el mensaje aparece en un recuadro rojo debajo de la sección.

![Interfaz web con el Product Backlog ordenado por prioridad](docs/evidencias/img/TEC-04/05-backlog-por-prioridad.png)

### API

La API responde bajo el prefijo `/api` (TEC-04).

| Método | Ruta | Operación |
|---|---|---|
| `POST` | `/api/proyectos` | Crear un proyecto (HU-01) |
| `GET` | `/api/proyectos/{id}` | Consultar un proyecto (HU-03) |
| `PUT` | `/api/proyectos/{id}` | Modificar nombre, descripción y estado (HU-02) |
| `POST` | `/api/proyectos/{id}/backlog` | Crear un ítem del Product Backlog (HU-04) |
| `GET` | `/api/proyectos/{id}/backlog` | Listar el Product Backlog del proyecto, ordenado por ID (HU-06) |
| `GET` | `/api/proyectos/{id}/backlog?orden=prioridad` | Listar el Product Backlog ordenado por prioridad: ALTA, MEDIA y BAJA (HU-09) |
| `GET` | `/api/items/{id}` | Consultar un ítem (HU-05) |
| `PUT` | `/api/items/{id}` | Modificar título, descripción, prioridad, criterios y horas estimadas (HU-07) |
| `PATCH` | `/api/items/{id}/estado` | Cambiar el estado de un ítem (HU-08) |
| `PATCH` | `/api/items/{id}/story-points` | Asignar o reestimar los story points de un ítem: 1, 2, 3, 5, 8, 13 o 21 (HU-20) |
| `POST` | `/api/items/{id}/poker-sessions` | Abrir una sesión de Planning Poker para un ítem (HU-21) |
| `POST` | `/api/poker-sessions/{id}/votes` | Votar una carta en una sesión abierta sin mostrar los votos; reemplaza el voto anterior del integrante (HU-22) |
| `POST` | `/api/poker-sessions/{id}/reveal` | Revelar todos los votos de una sesión abierta con votos; la sesión queda REVELADA (HU-23) |
| `POST` | `/api/poker-sessions/{id}/estimate` | Registrar la estimación acordada (`{"puntos": 13}`, escala 1, 2, 3, 5, 8, 13 o 21) de una sesión revelada: guarda los story points en el ítem y la sesión queda CERRADA (HU-25) |
| `POST` | `/api/items/{id}/defectos` | Registrar un defecto de una historia, con descripción, severidad y sprint de detección abierto (HU-29) |
| `GET` | `/api/items/{id}/esfuerzos` | Listar los esfuerzos de una historia, ordenados por ID, con el total de horas (HU-27) |
| `POST` | `/api/proyectos/{id}/sprints` | Crear un sprint (HU-10) |
| `POST` | `/api/sprints/{id}/items/{itemId}` | Asignar un ítem al sprint (HU-11) |
| `POST` | `/api/sprints/{id}/items/{itemId}/completar` | Marcar como TERMINADO un ítem del sprint (HU-12) |
| `POST` | `/api/sprints/{id}/items/{itemId}/esfuerzos` | Registrar el esfuerzo de un integrante en una historia del sprint (HU-26) |
| `GET` | `/api/sprints/{id}/esfuerzos` | Listar los esfuerzos de un sprint, ordenados por ID, con el total de horas (HU-27) |
| `GET` | `/api/sprints/{id}/completados` | Listar los ítems terminados del sprint (HU-12) |
| `POST` | `/api/sprints/{id}/cerrar` | Cerrar el sprint (HU-13) |
| `GET` | `/api/sprints/{id}/metricas` | Story points planificados y completados del sprint, e ítems sin estimar (HU-32) |
| `GET` | `/api/proyectos/{id}/metricas` | Story points planificados y completados de todo el backlog, ítems sin estimar y velocidad de los últimos 3 sprints cerrados, con un decimal o `null` (HU-32) |

Ejemplo con `curl` (en Windows, desde Git Bash):

```
curl -X POST localhost:8080/api/proyectos -d '{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas"}'
curl localhost:8080/api/proyectos/1
curl -X PUT localhost:8080/api/proyectos/1 -d '{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas","estado":"EN_CURSO"}'
curl -X POST localhost:8080/api/proyectos/1/backlog -d '{"titulo":"Registrar venta","descripcion":"Alta de una venta","prioridad":"ALTA","criterios":["Se guarda la venta"]}'
curl "localhost:8080/api/proyectos/1/backlog?orden=prioridad"
curl -X PUT localhost:8080/api/items/1 -d '{"titulo":"Registrar venta","descripcion":"Alta de una venta","prioridad":"ALTA","criterios":["Se guarda la venta"],"horas_estimadas":6}'
curl -X POST localhost:8080/api/proyectos/1/sprints -d '{"nombre":"Sprint 1","fecha_inicio":"2026-10-01","fecha_fin":"2026-10-14","objetivo":"Registrar ventas"}'
curl -X POST localhost:8080/api/sprints/1/items/1
curl -X PATCH localhost:8080/api/items/1/story-points -d '{"story_points":5}'
curl -X PATCH localhost:8080/api/items/1/estado -d '{"estado":"EN_CURSO"}'
curl -X POST localhost:8080/api/sprints/1/items/1/completar
curl -X POST localhost:8080/api/items/1/defectos -d '{"descripcion":"El total no suma el IVA","severidad":"ALTA","sprint_deteccion_id":1}'
curl localhost:8080/api/sprints/1/completados
curl -X POST localhost:8080/api/sprints/1/cerrar
```

`PUT /api/proyectos/{id}` y `PUT /api/items/{id}` reemplazan todos sus campos: un campo que no se envía queda vacío (o en 0, las horas estimadas), así que hay que mandar también los que no cambian. Las fechas de un sprint usan el formato `AAAA-MM-DD`. Para completar un ítem, antes tiene que estar `EN_CURSO`: el cambio de estado respeta las transiciones de HU-08.

Las respuestas son JSON. Los errores devuelven `{"error":"..."}` con el código 400 (datos inválidos), 404 (proyecto, ítem o sprint inexistente), 409 (el pedido choca con el estado actual: nombre repetido, transición no permitida, sprint abierto existente, sprint cerrado, ítem ya asignado, de otro proyecto o no asignado al sprint, story points de un ítem terminado, o sprint de detección de otro proyecto), 413 (cuerpo de más de 1 MiB) o 500 (error interno).

Para correr los tests: `go test ./...` (no necesita Docker: los tests de PostgreSQL se omiten si no hay base de pruebas).

## Persistencia con PostgreSQL

La aplicación guarda los datos en memoria o en PostgreSQL según la variable `STORAGE` (TEC-05). El resto del equipo puede trabajar siempre en memoria, sin Docker.

| Variable | Valores | Para qué |
|---|---|---|
| `STORAGE` | `memory` (por defecto, si no está definida) o `postgres` | Dónde se guardan los datos. Se compara tal cual: `Postgres` o ` memory` dan error |
| `DATABASE_URL` | URL de PostgreSQL | Obligatoria con `STORAGE=postgres` |
| `TEST_DATABASE_URL` | URL de una base que termine en `_test` | Solo para las pruebas de integración |

### 1. Levantar la base

Desde la raíz del repositorio:

```
docker compose up -d --wait
```

Levanta PostgreSQL 17 con la configuración del archivo `.env`, si existe. Si no existe, usa los valores de desarrollo de [`.env.example`](.env.example) (usuario `gpc`, base `gpc`). Para cambiarlos, copiá `.env.example` como `.env`: el `.env` real no se sube al repositorio.

- **Puerto 5433:** la base se publica en `127.0.0.1:5433`, no en el 5432, para no chocar con un PostgreSQL instalado en el sistema, que normalmente usa el 5432. Si el 5433 también está ocupado, cambiá `POSTGRES_PORT` en el `.env` (y el puerto en las URL).
- **Esquema:** los scripts de [`db/init/`](db/init/) crean las tablas en la base `gpc` y una base de pruebas `gpc_test` con el mismo esquema. **Solo corren la primera vez, cuando el volumen de datos está vacío.** Si después cambia el esquema, hay que borrar el volumen con `docker compose down -v` (eso **borra todos los datos**) y volver a levantar la base. **HU-11 agrega la columna `sprint_id` a `backlog_items`, HU-26 la tabla `efforts`, HU-21 la tabla `poker_sessions` y HU-29 la tabla `defects`:** una base creada antes de esas historias no las tiene, así que hay que recrearla (`docker compose down -v` y después `docker compose up -d --wait`).
- **Datos:** quedan en el volumen `gpc-data`, así que se conservan al reiniciar la aplicación y al hacer `docker compose down` (sin `-v`).

Para ver que la base está lista: `docker compose ps` (estado `healthy`) y `docker compose port postgres 5432` (tiene que mostrar `127.0.0.1:5433`).

### 2. Ejecutar la aplicación con PostgreSQL

La aplicación no lee el archivo `.env`: las variables se definen en la terminal.

En PowerShell:

```
$env:STORAGE="postgres"; $env:DATABASE_URL="postgres://gpc:gpc_dev@localhost:5433/gpc?sslmode=disable"; go run ./cmd/server
```

En Git Bash, Linux o macOS:

```
STORAGE=postgres DATABASE_URL="postgres://gpc:gpc_dev@localhost:5433/gpc?sslmode=disable" go run ./cmd/server
```

Al arrancar, el servidor muestra qué almacenamiento usa: `servidor escuchando en :8080 (almacenamiento: postgres)`. Si `STORAGE` tiene otro valor, si falta `DATABASE_URL` o si la base no responde, el servidor no arranca y muestra el error, que nunca incluye la URL (tiene la contraseña).

### 3. Pruebas de integración

Los tests de PostgreSQL corren solo si está definida `TEST_DATABASE_URL`. Sin ella (como en el CI, que no tiene Docker) se omiten con `SKIP`, y `go test ./...` pasa igual.

```
$env:TEST_DATABASE_URL="postgres://gpc:gpc_dev@localhost:5433/gpc_test?sslmode=disable"; go test -count=1 ./...
```

- La base de `TEST_DATABASE_URL` **tiene que terminar en `_test`**: si no, los tests fallan sin conectarse, porque vacían las tablas antes de cada caso.
- Si la variable está definida pero la base no responde, los tests de PostgreSQL se omiten.
- Los paquetes comparten la base de pruebas; un lock de PostgreSQL hace que esos tests corran de a uno aunque `go test` corra los paquetes en paralelo.

### Driver de PostgreSQL

Se usa **`github.com/jackc/pgx/v5`** (versión `v5.11.0`), la única dependencia externa del código de producción (CA-TEC05.10). El equipo aprobó pgx/v5 como única dependencia nueva antes de agregarla al go.mod. Motivos:

- Es el driver de PostgreSQL para Go más usado y tiene mantenimiento activo; `lib/pq`, la otra opción conocida, está en modo mantenimiento y sus autores recomiendan pgx.
- Trae su propio pool de conexiones (`pgxpool`), así que no hace falta otra librería.
- Lee y escribe los arreglos de PostgreSQL (`TEXT[]`) directamente como `[]string`, que es como se guardan los criterios de aceptación de los ítems.
- Expone los códigos de error de PostgreSQL (`pgconn.PgError`), que los tests usan para verificar la clave foránea.

## Equipo

| Rol | Integrante |
|---|---|
| Product Architect | Profesores de la cátedra |
| Agile Enabler | Peñalbé Hernán |
| Product Builder | Buttini Cristobal |
| Product Builder | Cardozo Leandro |
| Product Builder | García Santiago |
| Product Builder | Sosa Ricardo |

## Sprints

| Sprint | Fechas | Objetivo |
|---|---|---|
| 0 · Preparación | 07/09 – 08/10 | Equipo, entorno, tablero y Product Backlog inicial |
| 1 · MVP | 09/10 – 15/10 | Versión funcional básica (21 Story Points comprometidos) |
| 2 · Interfaz | 16/10 – 22/10 | Interfaz usable |
| 3 · Funcionalidad y Calidad | 23/10 – 29/10 | Funcionalidades clave, robustez y visualización |
| 4 · Cierre | 30/10 – 05/11 | Exportación a PDF, documentación y entrega final |

## Trazabilidad

Cada historia sigue la cadena: **Historia de Usuario → Especificación SDD → Criterios de Aceptación → Escenarios BDD → Tests → Código Go**.

| Artefacto | Dónde está |
|---|---|
| Historias de usuario y épicas | Issues del repositorio y tablero de GitHub Projects |
| Especificaciones SDD | [`docs/especificaciones/`](docs/especificaciones/) |
| Escenarios BDD | [`features/`](features/) |
| Tests y código Go | [`internal/`](internal/) y [`cmd/`](cmd/) |
| Evidencia TDD por historia | [`docs/evidencias/`](docs/evidencias/) |
| Cobertura y métricas | [`docs/reportes/`](docs/reportes/) |
| Actas de Planning, Daily, Review y Retrospective (una carpeta por Sprint) | [`docs/actas/`](docs/actas/) |
| Registro de uso de IA | [`docs/ia/registro.md`](docs/ia/registro.md) |
| Flujo de trabajo (SDD → BDD → TDD) | [`AGENTS.md`](AGENTS.md) |
| Antes de empezar una historia | [`docs/antes-de-empezar.md`](docs/antes-de-empezar.md) |

Cada Issue de historia incluye una sección **Trazabilidad** con los links a su especificación, sus escenarios y su evidencia.

## Historias por sprint

Vista por sprint de cada historia y tarea, con links a su documentación. Los archivos se organizan por ID (`HU-XX`) y no por sprint, porque una historia puede cambiar de sprint pero su ID no. Cada Pull Request que integra una historia completa su fila. Las actas de cada sprint están en [`docs/actas/`](docs/actas/).

### Sprint 0 — La Preparación (07/09 al 08/10) · [Acta de cierre](docs/actas/sprint-0/cierre.md)

Sin historias: equipo, entorno, tablero y Product Backlog inicial.

### Sprint 1 — El MVP (09/10 al 15/10) · [Planning](docs/actas/sprint-1/planning.md) · [Daily](docs/actas/sprint-1/daily.md) · [Review](docs/actas/sprint-1/review.md) · [Retrospectiva](docs/actas/sprint-1/retrospectiva.md)

| ID | Issue | Historia o tarea | SDD | BDD | Evidencia TDD | Pull Request |
|---|---|---|---|---|---|---|
| TEC-02 | [#32](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/32) | Integración continua | — | — | — | [#67](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/67) |
| HU-01 | [#14](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/14) | Crear proyecto | [SDD](docs/especificaciones/HU-01-crear-proyecto.md) | [BDD](features/HU-01-crear-proyecto.feature) | [Evidencia](docs/evidencias/HU-01-tdd.md) | [#68](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/68), [#72](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/72), [#82](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/82) |
| HU-03 | [#16](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/16) | Consultar estado del proyecto | [SDD](docs/especificaciones/HU-03-consultar-proyecto.md) | [BDD](features/HU-03-consultar-proyecto.feature) | [Evidencia](docs/evidencias/HU-03-tdd.md) | [#74](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/74) |
| HU-04 | [#17](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/17) | Crear ítem del backlog | [SDD](docs/especificaciones/HU-04-crear-item-backlog.md) | [BDD](features/HU-04-crear-item-backlog.feature) | [Evidencia](docs/evidencias/HU-04-tdd.md) | [#77](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/77) |
| HU-06 | [#21](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/21) | Listar el backlog de un proyecto | [SDD](docs/especificaciones/HU-06-listar-backlog.md) | [BDD](features/HU-06-listar-backlog.feature) | [Evidencia](docs/evidencias/HU-06-tdd.md) | [#78](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/78) |
| HU-02 | [#15](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/15) | Modificar datos del proyecto | [SDD](docs/especificaciones/HU-02-modificar-proyecto.md) | [BDD](features/HU-02-modificar-proyecto.feature) | [Evidencia](docs/evidencias/HU-02-tdd.md) | [#80](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/80) |
| TEC-01 | [#23](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/23) | Punto de entrada ejecutable (servidor HTTP) | — | — | [Evidencia](docs/evidencias/TEC-01-tdd.md) | [#81](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/81) |
| HU-05 | [#18](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/18) | Consultar un ítem del backlog (opcional) | [SDD](docs/especificaciones/HU-05-consultar-item.md) | [BDD](features/HU-05-consultar-item.feature) | [Evidencia](docs/evidencias/HU-05-tdd.md) | [#83](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/83) |

Las tareas técnicas no llevan SDD ni BDD formales (decisión 5 del acta de Planning del Sprint 1).

### Sprint 2 — La Interfaz (16/10 al 22/10) · [Planning](docs/actas/sprint-2/planning.md) · [Daily](docs/actas/sprint-2/daily.md) · [Review](docs/actas/sprint-2/review.md) · [Retrospectiva](docs/actas/sprint-2/retrospectiva.md)

| ID | Issue | Historia o tarea | SDD | BDD | Evidencia TDD | Pull Request |
|---|---|---|---|---|---|---|
| HU-07 | [#19](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/19) | Modificar datos de un ítem | [SDD](docs/especificaciones/HU-07-modificar-item.md) | [BDD](features/HU-07-modificar-item.feature) | [Evidencia](docs/evidencias/HU-07-tdd.md) | [#95](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/95) |
| HU-08 | [#20](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/20) | Cambiar el estado de un ítem | [SDD](docs/especificaciones/HU-08-cambiar-estado-item.md) | [BDD](features/HU-08-cambiar-estado-item.feature) | [Evidencia](docs/evidencias/HU-08-tdd.md) | [#99](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/99) |
| HU-09 | [#22](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/22) | Priorizar los ítems del backlog | [SDD](docs/especificaciones/HU-09-priorizar-backlog.md) | [BDD](features/HU-09-priorizar-backlog.feature) | [Evidencia](docs/evidencias/HU-09-tdd.md) | [#102](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/102) |
| HU-10 | [#24](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/24) | Crear sprint | [SDD](docs/especificaciones/HU-10-crear-sprint.md) | [BDD](features/HU-10-crear-sprint.feature) | [Evidencia](docs/evidencias/HU-10-tdd.md) | [#94](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/94) |
| HU-11 | [#25](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/25) | Asignar historias a un sprint | [SDD](docs/especificaciones/HU-11-asignar-item-sprint.md) | [BDD](features/HU-11-asignar-item-sprint.feature) | [Evidencia](docs/evidencias/HU-11-tdd.md) | [#97](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/97) |
| HU-12 | [#26](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/26) | Registrar historias completadas | [SDD](docs/especificaciones/HU-12-registrar-completadas.md) | [BDD](features/HU-12-registrar-completadas.feature) | [Evidencia](docs/evidencias/HU-12-tdd.md) | [#100](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/100) |
| HU-13 | [#27](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/27) | Cerrar sprint | [SDD](docs/especificaciones/HU-13-cerrar-sprint.md) | [BDD](features/HU-13-cerrar-sprint.feature) | [Evidencia](docs/evidencias/HU-13-tdd.md) | [#101](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/101) |
| TEC-04 | [#34](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/34) | Interfaz usable | — | — | [Evidencia](docs/evidencias/TEC-04-tdd.md) | [#104](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/104), [#105](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/105), [#106](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/106) |
| TEC-05 | [#35](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/35) | Persistencia con PostgreSQL en Docker | — | — | [Evidencia](docs/evidencias/TEC-05-tdd.md) | [#96](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/96) |

### Sprint 3 — La Estimación (23/10 al 29/10) · [Planning](docs/actas/sprint-3/planning.md)

| ID | Issue | Historia o tarea | SDD | BDD | Evidencia TDD | Pull Request |
|---|---|---|---|---|---|---|
| HU-20 | [#38](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/38) | Asignar story points a un ítem del backlog | [SDD](docs/especificaciones/HU-20-asignar-story-points.md) | [BDD](features/HU-20-asignar-story-points.feature) | [Evidencia](docs/evidencias/HU-20-tdd.md) | [#118](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/118) |
| HU-21 | [#39](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/39) | Abrir una sesión de Planning Poker | [SDD](docs/especificaciones/HU-21-abrir-sesion-planning-poker.md) | [BDD](features/HU-21-abrir-sesion-planning-poker.feature) | [Evidencia](docs/evidencias/HU-21-tdd.md) | [#122](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/122) |
| HU-22 | [#40](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/40) | Votar de forma oculta | [SDD](docs/especificaciones/HU-22-votar-oculto.md) | [BDD](features/HU-22-votar-oculto.feature) | [Evidencia](docs/evidencias/HU-22-tdd.md) | [#125](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/125) |
| HU-23 | [#41](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/41) | Revelar las estimaciones | [SDD](docs/especificaciones/HU-23-revelar-estimaciones.md) | [BDD](features/HU-23-revelar-estimaciones.feature) | [Evidencia](docs/evidencias/HU-23-tdd.md) | [#127](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/127) |
| HU-25 | [#43](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/43) | Registrar la estimación acordada | [SDD](docs/especificaciones/HU-25-registrar-estimacion-acordada.md) | [BDD](features/HU-25-registrar-estimacion-acordada.feature) | [Evidencia](docs/evidencias/HU-25-tdd.md) | PR en curso |
| HU-26 | [#44](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/44) | Registrar esfuerzo realizado en una historia | [SDD](docs/especificaciones/HU-26-registrar-esfuerzo.md) | [BDD](features/HU-26-registrar-esfuerzo.feature) | [Evidencia](docs/evidencias/HU-26-tdd.md) | [#120](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/120) |
| HU-27 | [#45](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/45) | Consultar los registros de esfuerzo | [SDD](docs/especificaciones/HU-27-consultar-esfuerzos.md) | [BDD](features/HU-27-consultar-esfuerzos.feature) | [Evidencia](docs/evidencias/HU-27-tdd.md) | [#126](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/126) |
| HU-29 | [#47](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/47) | Registrar un defecto | [SDD](docs/especificaciones/HU-29-registrar-defecto.md) | [BDD](features/HU-29-registrar-defecto.feature) | [Evidencia](docs/evidencias/HU-29-tdd.md) | [#124](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/124) |
| HU-32 | [#50](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/50) | Calcular Story Points planificados, completados y velocidad | [SDD](docs/especificaciones/HU-32-sp-velocidad.md) | [BDD](features/HU-32-sp-velocidad.feature) | [Evidencia](docs/evidencias/HU-32-tdd.md) | [#123](https://github.com/Elwilsonic/seguimiento-medicion-GPC/pull/123) |

## Estructura del repositorio

Las carpetas se van creando a medida que aparece su contenido.

| Ruta | Contenido |
|---|---|
| `AGENTS.md` | Instrucciones del flujo de trabajo (SDD → BDD → TDD) |
| `.github/` | Workflow de CI, plantillas de Issue (historia, tarea técnica, ceremonia y defecto) y plantilla de Pull Request |
| `cmd/` | Punto de entrada ejecutable de la aplicación |
| `internal/` | Código fuente en Go (un paquete por funcionalidad) y sus tests; `internal/httpapi` expone la API, `internal/web` tiene la interfaz web (TEC-04), `internal/postgres` abre la conexión a PostgreSQL e `internal/postgrestest` prepara la base de las pruebas de integración |
| `docker-compose.yml` | Base de datos PostgreSQL en Docker (TEC-05) |
| `db/init/` | Scripts que crean el esquema de la base y la base de pruebas la primera vez que arranca |
| `.env.example` | Valores de ejemplo de las variables de entorno (el `.env` real no se sube) |
| `features/` | Escenarios BDD en Gherkin (`HU-XX-nombre.feature`) |
| `docs/actas/` | Actas de cada Sprint, en una carpeta por Sprint (`sprint-N/planning.md`, `daily.md`, `review.md`, `retrospectiva.md`) |
| `docs/actas/_plantillas/` | Plantillas de las actas de Planning, Review y Retrospectiva y del comentario de la Daily |
| `docs/antes-de-empezar.md` | Checklist para ponerse al día antes de empezar una historia, con o sin IA |
| `docs/definition-of-ready-done.md` | Definition of Ready y Definition of Done vigentes |
| `docs/especificaciones/` | Especificaciones SDD (`HU-XX-nombre.md`) |
| `docs/evidencias/` | Evidencia TDD por historia (`HU-XX-tdd.md`) |
| `docs/reportes/` | Reportes de cobertura y métricas |
| `docs/ia/` | Registro de uso de herramientas de IA |

El Product Backlog y los Sprint Backlogs están en el tablero de GitHub Projects.

## Estimación de épicas (Planning Poker)

Las épicas se estimaron por Planning Poker con talles de ropa (XS, S, M, L, XL). Las historias de usuario se estiman en Story Points con la escala Fibonacci (1, 2, 3, 5, 8, 13, 21).

| Épica | Prioridad | Talle | Justificación |
|---|---|---|---|
| E1 · Gestión de proyectos | Alta | S | 3 historias, altas/cambios/consulta simples. |
| E2 · Product Backlog | Alta | M | 6 historias, con validaciones, estados y orden. |
| E3 · Gestión de Sprints | Alta | L | 5 historias con reglas de estado (abierto/cerrado, asignación, sin cambios al cerrar). |
| E4 · Registro de integrantes | Media | S | 3 historias: alta, cambio y consulta. |
| E5 · Fechas del proyecto | Media | XS | 2 historias y una sola validación (fin posterior a inicio). |
| E6 · Estimación | Media | L | 6 historias, con votos ocultos, rondas y acuerdo. |
| E7 · Registro de esfuerzo | Media | M | 3 historias, pero relacionadas con integrantes, historias y sprints. |
| E8 · Gestión de defectos | Baja | S | 3 historias con estados y una regla de sprints. |
| E9 · Métricas | Baja | L | 4 historias con fórmulas que dependen de casi todo lo anterior. |
| E10 · Dashboard | Baja | L | 2 historias, pero con interfaz y gráficos. |
| E11 · Reportes | Baja | XL | 3 historias, y la exportación a PDF usa una librería externa. |

## Estado del proyecto

🟡 En desarrollo — Sprint 1 (El MVP) cerrado el 15/10, Sprint 2 (La Interfaz) cerrado el 22/10 y Sprint 3 (La Estimación) en curso, del 23/10 al 29/10

**Sprint 1:** se cumplió el Sprint Goal ("Poder crear, modificar y consultar proyectos y cargar y listar su Product Backlog desde una aplicación ejecutable, con cada cambio validado por integración continua"), con 21 de 21 Story Points completados y la historia opcional HU-05. La aplicación se ejecuta como indica la sección [Cómo ejecutar](#cómo-ejecutar). Ver las actas de [Planning](docs/actas/sprint-1/planning.md), [Review](docs/actas/sprint-1/review.md) y [Retrospectiva](docs/actas/sprint-1/retrospectiva.md), y el [registro de las Dailys](docs/actas/sprint-1/daily.md).

**Sprint 2:** se cumplió el Sprint Goal ("Usar la aplicación desde el navegador para gestionar proyectos, su Product Backlog y sus sprints, con los datos guardados en PostgreSQL"), con 25 de 25 Story Points completados (HU-07 a HU-13) y las tareas técnicas TEC-04 (interfaz web) y TEC-05 (PostgreSQL en Docker). La interfaz se usa como indica la sección [Interfaz web](#interfaz-web). Ver las actas de [Planning](docs/actas/sprint-2/planning.md), [Review](docs/actas/sprint-2/review.md) y [Retrospectiva](docs/actas/sprint-2/retrospectiva.md), y el [registro de las Dailys](docs/actas/sprint-2/daily.md).

**Sprint 3 (en curso):** Sprint Goal "Estimar ítems con Planning Poker y registrar el esfuerzo y los defectos de cada sprint, con las fórmulas de las métricas definidas para el Sprint 4", con 29 Story Points comprometidos (HU-18, HU-20 a HU-23, HU-25 a HU-27 y HU-29) y TEC-03. Ver el acta de [Planning](docs/actas/sprint-3/planning.md).
