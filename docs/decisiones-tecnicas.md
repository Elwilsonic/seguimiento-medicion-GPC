# Decisiones técnicas

> Issue del repositorio: #69

Las decisiones técnicas del proyecto, cada una con su "por qué", las alternativas que se descartaron (cuando quedaron registradas) y dónde se tomó (TEC-06; documento comprometido en el tema 7 del acta de Planning del Sprint 3). El detalle de cómo está construida la aplicación está en la [documentación técnica](documentacion-tecnica.md); el de cada regla de negocio, en la SDD de su historia (`docs/especificaciones/`).

## Resumen

| # | Decisión | Fuente |
|---|---|---|
| 1 | Go con la biblioteca estándar y `net/http`, sin frameworks | Planning del Sprint 1, tema 6 |
| 2 | Un paquete por dominio, con un Manager y una interfaz `Repository` | `AGENTS.md`; Sprint 0, pendientes |
| 3 | Copias por valor en todo lo que devuelven los Manager y los repositorios | `AGENTS.md` |
| 4 | Persistencia en memoria primero y PostgreSQL en Docker desde el Sprint 2 | Sprint 0, decisión 4; Planning del Sprint 1, tema 10 |
| 5 | Driver `jackc/pgx/v5` | TEC-05; README, "Driver de PostgreSQL" |
| 6 | Almacenamiento elegido con `STORAGE`, en memoria por defecto | TEC-05 |
| 7 | Un `sync.Mutex` por Manager, con un orden fijo entre Managers | Documentación técnica, sección 1.3 |
| 8 | Traducción de errores a códigos HTTP en un solo lugar (`statusFor`) | TEC-04 |
| 9 | Interfaz web con HTML y JavaScript propios, embebidos en el binario | Planning del Sprint 2, tema 5 |
| 10 | Gráficos del dashboard en SVG propio | Planning del Sprint 4, tema 4 |
| 11 | Librería de PDF `go-pdf/fpdf` | Planning del Sprint 4, tema 3 |
| 12 | Escenarios BDD con godog, en español y en modo estricto | Planning del Sprint 1, temas 11 y 14 |
| 13 | Pruebas de integración con una base `_test` y `TEST_DATABASE_URL` | TEC-05 |
| 14 | Integración continua desde el Sprint 1, con PostgreSQL desde el cierre | Planning del Sprint 1, tema 12; Retrospectivas de los Sprints 1 y 4 |
| 15 | Fibonacci para las historias y talles para las épicas | Sprint 0, decisión 2 |
| 16 | Horas estimadas por ítem del backlog, sin tareas | Sprint 0, decisión 3 |
| 17 | Fórmulas de las métricas definidas antes de programarlas | TEC-03, `docs/metricas.md` |
| 18 | Historias chicas, que entran completas en un Sprint | Planning de los Sprints 1 y 3 |
| 19 | Ramas por historia, Pull Request y "Create a merge commit" | Planning del Sprint 1, tema 13 |
| 20 | Integrantes como texto libre en el esfuerzo y en Planning Poker | SDD de HU-15 y HU-16 |
| 21 | Los reportes listan los integrantes que registraron esfuerzo | Planning del Sprint 4, tema 5 |

## Arquitectura y lenguaje

### 1. Go con la biblioteca estándar y `net/http`

- **Decisión:** toda la aplicación está en Go. El servidor HTTP (TEC-01) usa solo `net/http`, con el `ServeMux` de Go 1.22 (rutas con método y comodines como `GET /api/proyectos/{id}`).
- **Por qué:** la consigna pide Go, y `net/http` ya resuelve el ruteo, los timeouts y el JSON (con `encoding/json`). Con menos dependencias, el código es más fácil de revisar para todo el equipo y el CI no depende de librerías de terceros.
- **Alternativas descartadas:** cualquier framework web, porque sería una dependencia externa que `AGENTS.md` no permite sin aprobación del equipo.
- **Fuente:** tema 6 del acta de Planning del Sprint 1; `AGENTS.md`, "Tecnología y dependencias".

### 2. Un paquete por dominio, con un Manager y una interfaz `Repository`

- **Decisión:** cada dominio (`project`, `backlog`, `sprint`, `poker`, `effort`, `defect`, `member`) es un paquete de `internal/` con un `Manager` que aplica las reglas de negocio y una interfaz `Repository` con dos implementaciones (memoria y PostgreSQL). Los paquetes sin datos propios (`metrics`, `dashboard`, `report`) solo usan los Manager de los demás. La capa HTTP (`internal/httpapi`) solo llama a los Manager.
- **Por qué:** `AGENTS.md` pide que el código de negocio nunca acceda directamente al almacenamiento. Con la interfaz, el Sprint 1 trabajó en memoria y TEC-05 sumó PostgreSQL sin cambiar las reglas ni los tests (CA-TEC05.4). Un Manager nunca accede al repositorio de otro dominio, así sus validaciones no se pueden saltear.
- **Alternativa descartada:** el acceso directo al almacenamiento desde los handlers o el código de negocio, que `AGENTS.md` prohíbe.
- **Fuente:** `AGENTS.md`; acta de cierre del Sprint 0 ("Pendientes que pasan al Sprint 1"); SDD de HU-01.

### 3. Copias por valor

- **Decisión:** las operaciones de consulta y de escritura devuelven copias por valor o estructuras nuevas, incluidas copias nuevas de los slices (criterios de aceptación, votos), nunca punteros a los datos internos.
- **Por qué:** así ningún llamador puede cambiar el estado guardado sin pasar por las validaciones. Cada historia tiene un criterio de aceptación que lo verifica ("modificar lo devuelto no altera lo guardado").
- **Fuente:** `AGENTS.md`, paso 2; documentación técnica, sección 1.4.

## Persistencia

### 4. Memoria primero y PostgreSQL en Docker desde el Sprint 2

- **Decisión:** en el Sprint 1 los datos se guardan en memoria. Desde TEC-05 (Sprint 2), cada repositorio tiene también su implementación en PostgreSQL 17, que corre en Docker (`docker-compose.yml`). El esquema está en `db/init/01-schema.sql` y los IDs los asigna la base (`GENERATED ALWAYS AS IDENTITY`).
- **Por qué:** el Sprint 1 se concentró en las reglas de negocio y el flujo SDD → BDD → TDD, sin depender de una base. Al sumar PostgreSQL detrás de la misma interfaz, las reglas de negocio y los tests existentes no cambiaron, las pruebas unitarias siguieron corriendo sin base y la base se prueba con pruebas de integración.
- **Fuente:** decisión 4 del acta de cierre del Sprint 0; tema 10 del acta de Planning del Sprint 1; evidencia de TEC-05.

### 5. Driver `jackc/pgx/v5`

- **Decisión:** el acceso a PostgreSQL usa `github.com/jackc/pgx/v5` (v5.11.0) con un único pool (`pgxpool`) compartido por todos los repositorios.
- **Por qué:** es el driver de PostgreSQL más usado en Go y tiene mantenimiento activo; trae su propio pool; lee los arreglos `TEXT[]` como `[]string` (los criterios de aceptación) y expone los códigos de error de PostgreSQL, que los tests usan para verificar la clave foránea.
- **Alternativa descartada:** `lib/pq`, en modo mantenimiento y cuyos autores recomiendan pgx.
- **Fuente:** README, "Driver de PostgreSQL"; CA-TEC05.10.

### 6. `STORAGE` elige el almacenamiento, en memoria por defecto

- **Decisión:** la variable `STORAGE` elige `memory` (por defecto) o `postgres`; con `postgres`, `DATABASE_URL` es obligatoria. Un valor desconocido hace que el servidor no arranque, y el error nunca muestra la URL, que tiene la contraseña.
- **Por qué:** el equipo puede trabajar siempre en memoria, sin Docker (README, "Persistencia con PostgreSQL"), y quien necesita datos persistentes levanta la aplicación con PostgreSQL sin cambiar código. Con una configuración inválida el servidor no arranca, en lugar de seguir con otro almacenamiento.
- **Fuente:** evidencia de TEC-05; documentación técnica, sección 5.2.

## Concurrencia y errores

### 7. Un `sync.Mutex` por Manager, con un orden fijo

- **Decisión:** cada Manager que lee y después escribe protege la operación completa con su mutex. Cuando una operación necesita dos Manager, el orden es siempre el mismo (por ejemplo, sprint → backlog o defect → sprint → backlog), y `metrics`, `dashboard` y `report` no tienen mutex propio. `WithOpenSprint` deja hacer una operación con el sprint abierto garantizado, sin que lo cierren a mitad de camino.
- **Por qué:** el servidor atiende pedidos en paralelo. Sin el mutex, dos pedidos simultáneos podrían dejar dos proyectos con el mismo nombre (CA-01.7), dos sprints abiertos (CA-10.6) o dos sesiones de Poker abiertas para un ítem (CA-21.6). El orden fijo evita bloqueos mutuos. El CI corre los tests con `-race`.
- **Fuente:** documentación técnica, sección 1.3; SDD de HU-01, HU-10, HU-21, HU-26 y HU-29.

### 8. Errores traducidos en un solo lugar

- **Decisión:** los Manager devuelven errores de dominio (`ErrProjectNotFound`, `ErrSprintClosed`…) y la capa HTTP los traduce a códigos en `statusFor`: 400 para datos inválidos, 404 para la entidad de la ruta que no existe, 409 para un pedido que choca con el estado actual, 413 para un cuerpo de más de 1 MiB y 500 para lo no previsto, sin exponer el detalle. Todo error responde `{"error": "…"}` en español.
- **Por qué:** el dominio no conoce HTTP, y cada código se decide una sola vez. Un error nuevo que no se agrega a `statusFor` responde 500, y su test lo detecta.
- **Fuente:** evidencia de TEC-04 (CA-TEC04.5 y CA-TEC04.6 para el 413 y el JSON inválido); documentación técnica, sección 3.3.

## Interfaz y reportes

### 9. HTML y JavaScript propios, embebidos en el binario

- **Decisión:** la interfaz es una sola página (`index.html`, `app.js` y `styles.css`) servida con `embed` por el mismo servidor, sin frameworks. Usa la API con `fetch` y escribe los datos con `textContent`, nunca con `innerHTML`.
- **Por qué:** no suma librerías externas (CA-TEC04.4), y la interfaz se sirve desde el mismo binario (`//go:embed`), sin otro proceso. `textContent` evita que un nombre cargado por el usuario se interprete como HTML.
- **Fuente:** tema 5 del acta de Planning del Sprint 2; evidencia de TEC-04.

### 10. Gráficos en SVG propio

- **Decisión:** los gráficos del dashboard (HU-37) se dibujan como SVG generado con JavaScript propio, a partir del JSON de las métricas.
- **Por qué:** una librería de gráficos sería una dependencia externa más, que `AGENTS.md` no permite sin aprobación, y para tres gráficos de barras no hace falta.
- **Fuente:** tema 4 del acta de Planning del Sprint 4; SDD de HU-37.

### 11. Librería de PDF `go-pdf/fpdf`

- **Decisión:** HU-40 exporta el reporte con `github.com/go-pdf/fpdf` (v0.9.0, licencia MIT), solo en `internal/report/pdf.go`.
- **Por qué:** la biblioteca estándar no genera PDF; está escrita en Go puro, sin cgo, así que compila en las máquinas del equipo sin compilador de C; y escribe texto en páginas A4 con pocas líneas.
- **Alternativas descartadas:** `jung-kurt/gofpdf`, archivada (fpdf es su continuación); librerías con licencia comercial o AGPL; herramientas que necesitan un programa externo.
- **Fuente:** tema 3 del acta de Planning del Sprint 4; README, "Librería de PDF"; SDD de HU-40.

## Pruebas e integración continua

### 12. Escenarios BDD con godog

- **Decisión:** los escenarios están en Gherkin y en español (`features/HU-XX-*.feature`), y cada historia tiene un runner de godog en su paquete, en modo estricto (un paso sin implementar hace fallar el test). Desde el Sprint 3, cada escenario lleva en su nombre el criterio que cubre.
- **Por qué:** así los escenarios corren con el mismo `go test ./...` y en el CI, y la trazabilidad entre criterio y escenario se ve en la salida de godog. godog es la única librería de tests (aprobada en `AGENTS.md`); los tests unitarios usan solo `testing`.
- **Fuente:** temas 11 y 14 del acta de Planning del Sprint 1; tema 7 del acta de Planning del Sprint 3.

### 13. Pruebas de integración con una base `_test`

- **Decisión:** los tests de PostgreSQL corren solo si está definida `TEST_DATABASE_URL`, y `internal/postgrestest` se niega a conectarse a una base que no termine en `_test`. Antes de cada caso vacía las tablas, y un lock de PostgreSQL (`pg_advisory_lock`) hace que los paquetes usen la base de a uno. Las dos implementaciones de cada repositorio pasan el mismo contrato (`runRepositoryContract`).
- **Por qué:** los tests vacían tablas, así que no pueden tocar la base de desarrollo por error. El contrato común garantiza que cambiar de almacenamiento no cambia el comportamiento.
- **Fuente:** evidencia de TEC-05 (CA-TEC05.7); documentación técnica, secciones 2.3 y 4.4.

### 14. Integración continua desde el Sprint 1

- **Decisión:** GitHub Actions corre `gofmt -l`, `go vet ./...` y `go test -race` en cada Pull Request hacia `main`. Desde el cierre del Sprint 4 levanta PostgreSQL como servicio, corre las pruebas de integración y falla si algún test se omite.
- **Por qué:** TEC-02 se adelantó al Sprint 1 para que los primeros Pull Requests con código ya se controlaran. `go vet` se sumó por la Retrospectiva del Sprint 1, y PostgreSQL por la del Sprint 4: hasta entonces, el SQL nuevo dependía de que el revisor corriera los tests a mano.
- **Fuente:** tema 12 del acta de Planning del Sprint 1; acción 8 de las Retrospectivas de los Sprints 1 y 4; `.github/workflows/ci.yml`.

## Estimación y métricas

### 15. Fibonacci para las historias y talles para las épicas

- **Decisión:** las historias se estiman con Fibonacci (1, 2, 3, 5, 8, 13, 21) en Planning Poker, y las épicas con talles de ropa (XS a XL). La misma escala es la de la aplicación (HU-20).
- **Por qué:** las métricas suman Story Points, así que las historias necesitan números; las épicas solo necesitan una idea de tamaño, que los profesores pidieron ver en el README.
- **Fuente:** decisión 2 del acta de cierre del Sprint 0.

### 16. Horas estimadas por ítem del backlog

- **Decisión:** las horas estimadas se registran en cada ítem del backlog (HU-07), y el esfuerzo real se registra sobre la historia (HU-26). No se modelan tareas.
- **Por qué:** las métricas comparan horas estimadas y reales por historia, sprint y proyecto, y un nivel de tareas agregaba historias y pantallas que la consigna no pide.
- **Fuente:** decisión 3 del acta de cierre del Sprint 0; registro de IA del 06/10/2026.

### 17. Fórmulas de las métricas antes de programarlas

- **Decisión:** TEC-03 definió las métricas M1 a M10 en `docs/metricas.md`, con su fórmula, unidad, nivel (sprint o proyecto) y qué pasa con los datos faltantes, antes de las historias de E9. Por ejemplo, la velocidad es el promedio de los últimos 3 sprints cerrados, y una métrica que no se puede calcular es "no disponible" (`null` en la API), no un error.
- **Por qué:** varias historias (métricas, dashboard, reportes e informe final) usan los mismos números, y tenían que coincidir en todas.
- **Fuente:** tema 9 del acta de Planning del Sprint 1; tema 4 del acta de Planning del Sprint 3; `docs/metricas.md`.

## Proceso

### 18. Historias chicas

- **Decisión:** las 11 funcionalidades mínimas de la consigna son las épicas, y cada una se parte en historias que entran completas en un Sprint. Desde el Sprint 3, las historias grandes se dividieron (por ejemplo, HU-41 salió de HU-18 y HU-47 de HU-34).
- **Por qué:** cada historia recorre el ciclo SDD → BDD → TDD completo; con historias chicas, el ciclo es corto y la historia entra en `main` dentro del Sprint.
- **Fuente:** tema 1 del acta de Planning del Sprint 1; tema 3 del acta de Planning del Sprint 3.

### 19. Ramas por historia y merge commit

- **Decisión:** cada historia o tarea se trabaja en su rama (`feature/HU-XX-…`) y entra a `main` por Pull Request, con la aprobación de otro integrante, la CI en verde y "Create a merge commit" (nunca squash ni rebase). Para traer cambios de `main` a una rama se usa merge, no rebase.
- **Por qué:** el merge commit conserva los commits RED → GREEN → REFACTOR, y los hashes que citan las evidencias siguen existiendo en `main`.
- **Fuente:** tema 13 del acta de Planning del Sprint 1; `AGENTS.md`, paso 8.

## Alcance de las historias

### 20. Integrantes como texto libre en el esfuerzo y en Planning Poker

- **Decisión:** el integrante de un registro de esfuerzo (HU-26) y de un voto de Planning Poker (HU-22) es un texto, no una referencia a los integrantes registrados (HU-15 a HU-17). Modificar un integrante no cambia los registros anteriores.
- **Por qué:** cuando se hizo E4, el esfuerzo y Planning Poker ya estaban en `main` con el integrante como texto. Relacionarlos con los integrantes registrados cambiaba historias cerradas y los reportes, y la SDD de HU-15 lo dejó como otra historia ("Unir los dos conceptos es otra historia").
- **Fuente:** decisión D7 de la SDD de HU-15; RN-16.8 de la SDD de HU-16; manual de usuario, "Límites de la aplicación".

### 21. Los reportes listan a quienes registraron esfuerzo

- **Decisión:** los reportes del proyecto (HU-38) y del sprint (HU-39) listan los integrantes que registraron esfuerzo, no los integrantes registrados en E4.
- **Por qué:** cuando se planificó HU-38, E4 estaba fuera de la entrega; el criterio CA-38.1 se ajustó para que el reporte no dependiera de una épica sin hacer.
- **Fuente:** tema 5 del acta de Planning del Sprint 4; SDD de HU-38.
