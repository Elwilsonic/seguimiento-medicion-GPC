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
- **Integración continua:** Cada Pull Request hacia `main` corre `gofmt` y `go test -race ./...` con GitHub Actions
- **Persistencia:** en memoria durante el Sprint 1; PostgreSQL en Docker a partir del Sprint 2

## Requisitos

- Go 1.27.1 o superior
- A partir del Sprint 2: Docker (para ejecutar PostgreSQL)

## Cómo ejecutar

Desde la raíz del repositorio:

```
go run ./cmd/server
```

El servidor escucha en el puerto `8080`. Para usar otro, definí la variable de entorno `PORT` (en PowerShell: `$env:PORT="9090"; go run ./cmd/server`; en Linux o macOS: `PORT=9090 go run ./cmd/server`). En el Sprint 1 los datos se guardan en memoria: se pierden al detener el servidor.

| Método | Ruta | Operación |
|---|---|---|
| `POST` | `/proyectos` | Crear un proyecto (HU-01) |
| `GET` | `/proyectos/{id}` | Consultar un proyecto (HU-03) |
| `PUT` | `/proyectos/{id}` | Modificar nombre, descripción y estado (HU-02) |
| `POST` | `/proyectos/{id}/backlog` | Crear un ítem del Product Backlog (HU-04) |
| `GET` | `/proyectos/{id}/backlog` | Listar el Product Backlog del proyecto (HU-06) |

Ejemplo con `curl` (en Windows, desde Git Bash):

```
curl -X POST localhost:8080/proyectos -d '{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas"}'
curl localhost:8080/proyectos/1
curl -X PUT localhost:8080/proyectos/1 -d '{"nombre":"Sistema de ventas","descripcion":"Gestión de ventas","estado":"EN_CURSO"}'
curl -X POST localhost:8080/proyectos/1/backlog -d '{"titulo":"Registrar venta","descripcion":"Alta de una venta","prioridad":"ALTA","criterios":["Se guarda la venta"]}'
curl localhost:8080/proyectos/1/backlog
```

`PUT /proyectos/{id}` reemplaza los tres campos (nombre, descripción y estado): un campo que no se envía queda vacío, así que hay que mandar también los que no cambian.

Las respuestas son JSON. Los errores devuelven `{"error":"..."}` con el código 400 (datos inválidos), 404 (proyecto inexistente), 409 (nombre repetido) o 500 (error interno).

Para correr los tests: `go test ./...`

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

### Sprint 2 — La Interfaz (16/10 al 22/10) · [Planning](docs/actas/sprint-2/planning.md)

| ID | Issue | Historia o tarea | SDD | BDD | Evidencia TDD | Pull Request |
|---|---|---|---|---|---|---|
| HU-07 | [#19](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/19) | Modificar datos de un ítem | — | — | — | — |
| HU-08 | [#20](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/20) | Cambiar el estado de un ítem | — | — | — | — |
| HU-09 | [#22](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/22) | Priorizar los ítems del backlog | — | — | — | — |
| HU-10 | [#24](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/24) | Crear sprint | — | — | — | — |
| HU-11 | [#25](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/25) | Asignar historias a un sprint | — | — | — | — |
| HU-12 | [#26](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/26) | Registrar historias completadas | — | — | — | — |
| HU-13 | [#27](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/27) | Cerrar sprint | — | — | — | — |
| TEC-04 | [#34](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/34) | Interfaz usable | — | — | — | — |
| TEC-05 | [#35](https://github.com/Elwilsonic/seguimiento-medicion-GPC/issues/35) | Persistencia con PostgreSQL en Docker | — | — | — | — |

## Estructura del repositorio

Las carpetas se van creando a medida que aparece su contenido.

| Ruta | Contenido |
|---|---|
| `AGENTS.md` | Instrucciones del flujo de trabajo (SDD → BDD → TDD) |
| `.github/` | Workflow de CI, plantillas de Issue (historia, tarea técnica, ceremonia y defecto) y plantilla de Pull Request |
| `cmd/` | Punto de entrada ejecutable de la aplicación |
| `internal/` | Código fuente en Go (un paquete por funcionalidad) y sus tests |
| `features/` | Escenarios BDD en Gherkin (`HU-XX-nombre.feature`) |
| `docs/actas/` | Actas de cada Sprint, en una carpeta por Sprint (`sprint-N/planning.md`, `daily.md`, `review.md`, `retrospectiva.md`) |
| `docs/actas/_plantillas/` | Plantillas de las actas de Planning, Review y Retrospectiva y del comentario de la Daily |
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

🟡 En desarrollo — Sprint 1 (MVP) cerrado el 15/10; el Sprint 2 (La Interfaz) va del 16/10 al 22/10

**Sprint 1:** se cumplió el Sprint Goal ("Poder crear, modificar y consultar proyectos y cargar y listar su Product Backlog desde una aplicación ejecutable, con cada cambio validado por integración continua"), con 21 de 21 Story Points completados y la historia opcional HU-05. La aplicación se ejecuta como indica la sección [Cómo ejecutar](#cómo-ejecutar). Ver las actas de [Planning](docs/actas/sprint-1/planning.md), [Review](docs/actas/sprint-1/review.md) y [Retrospectiva](docs/actas/sprint-1/retrospectiva.md), y el [registro de las Dailys](docs/actas/sprint-1/daily.md).

**Sprint 2 (16/10 al 22/10):** usar la aplicación desde el navegador para gestionar proyectos, su Product Backlog y sus sprints, con los datos guardados en PostgreSQL. Ver el [acta de Planning](docs/actas/sprint-2/planning.md).
