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
- **Persistencia:** en memoria durante el Sprint 1; PostgreSQL en Docker a partir del Sprint 2

## Requisitos

- Go 1.27.1 o superior
- A partir del Sprint 2: Docker (para ejecutar PostgreSQL)

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
| Actas de Planning, Review y Retrospective | [`docs/actas/`](docs/actas/) |
| Registro de uso de IA | [`docs/ia/registro.md`](docs/ia/registro.md) |
| Flujo de trabajo (SDD → BDD → TDD) | [`AGENTS.md`](AGENTS.md) |

Cada Issue de historia incluye una sección **Trazabilidad** con los links a su especificación, sus escenarios y su evidencia.

## Estructura del repositorio

Las carpetas se van creando a medida que aparece su contenido.

| Ruta | Contenido |
|---|---|
| `AGENTS.md` | Instrucciones del flujo de trabajo (SDD → BDD → TDD) |
| `cmd/` | Punto de entrada ejecutable de la aplicación |
| `internal/` | Código fuente en Go (un paquete por funcionalidad) y sus tests |
| `features/` | Escenarios BDD en Gherkin (`HU-XX-nombre.feature`) |
| `docs/actas/` | Actas de Planning, Review y Retrospective de cada Sprint |
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

🟡 En desarrollo — Sprint 1 (MVP), del 09/10 al 15/10

**Sprint Goal:** Poder crear, modificar y consultar proyectos y cargar y listar su Product Backlog desde una aplicación ejecutable, con cada cambio validado por integración continua. Ver el [acta de Planning](docs/actas/sprint-1-planning.md).
