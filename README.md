# Software Metrics & Estimation

Aplicación para la estimación, planificación, seguimiento y calidad de proyectos de software.

Trabajo Práctico Integrador — Ingeniería y Calidad de Software (2026)
Universidad Tecnológica Nacional — Facultad Regional San Rafael

## Tecnologías y prácticas

- **Lenguaje:** Go (Golang)
- **Metodología:** Scrum
- **Prácticas:** SDD (Specification-Driven Development), BDD (Behavior-Driven Development), TDD (Test-Driven Development)
- **Control de versiones:** Git

## Equipo

| Rol | Integrante |
|---|---|
| Product Architect | Profesores de la cátedra |
| Agile Enabler | Peñalbé Hernán |
| Product Builder | Buttini Cristobal |
| Product Builder | Cardozo Leandro |
| Product Builder | García Santiago |
| Product Builder | Sosa Ricardo |

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

El Product Backlog y los Sprint Backlogs están en el tablero de GitHub Projects.

## Estado del proyecto

🟡 En desarrollo — Sprint 0 (Preparación)