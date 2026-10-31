# Informe final de métricas y cobertura

> Issue del repositorio: #71

Informe final del proyecto "Software Metrics & Estimation" (TEC-08): las métricas del proceso por Sprint, la cobertura de pruebas y su evolución, las pruebas automatizadas, la trazabilidad de los criterios de aceptación, las evidencias de TDD y un análisis de los resultados. Cada dato indica de dónde sale, en la columna "Fuente" o al pie de su tabla.

- **Fecha de corte:** 30/10/2026, día del cierre del desarrollo; la entrega es el 02/11/2026.
- **Código medido:** `main` en `a39592c`. Lo que se integre después solo cambia documentación (actas del Sprint 4 y este informe).
- **Entorno de medición:** Go 1.27.1 en Linux; PostgreSQL 16.15 para las pruebas de integración.
- **Fórmulas:** las de [`docs/metricas.md`](../metricas.md) (TEC-03), aplicadas al proceso del equipo (sección 2).

## 1. Indicadores principales

| Indicador | Valor | Fuente |
|---|---|---|
| Historias de usuario completadas | **42 de 42** (100 %), en 4 Sprints | Issues con la etiqueta `tipo:historia`; actas de Review |
| Story Points completados | **139** (21 + 25 + 34 + 59) | Actas de Review de los Sprints 1 a 4 |
| Velocidad del equipo (M3) | **39.3** Story Points por Sprint (últimos 3 Sprints) | Sección 2.1 |
| Requerimientos mínimos de la consigna | **9 de 9** cubiertos | [Review del Sprint 4](../actas/sprint-4/review.md) |
| Cobertura total con PostgreSQL (`-coverpkg=./...`) | **96.2 %** de las sentencias | Medición sobre `a39592c` (sección 3) |
| Cobertura total sin base, como el CI | **84.5 %** | Medición sobre `a39592c` (sección 3) |
| Funciones de test en Go | **1108**, que con los subtests suman **2591** casos en verde y 0 omitidos con base | `go test -v ./...` sobre `a39592c` |
| Escenarios BDD (godog, en español) | **486** escritos en 42 archivos, que se ejecutan como **656** | `features/`; `go test -v ./...` |
| Criterios de aceptación con prueba automatizada | **341 de 341**: 313 con escenario BDD y 28 con tests de Go | Sección 5 |
| Historias con el RED antes del GREEN | **42 de 42** | `git log` de `main` (sección 6) |
| Commits (sin merges) | **359**: 73 `test`, 54 `feat`, 5 `fix`, 4 `refactor`, 208 `docs` y 15 de otros tipos | `git log` de `main` |
| Pull Requests integrados | **82**, con merge commit; 80 con la aprobación de otro integrante | GitHub (sección 7) |
| Ejecuciones de la CI | **152**: 151 en verde | GitHub Actions (sección 7) |
| Horas reales registradas en las Dailys | **259.5 h** | Actas de Daily de los Sprints 1 a 4 (sección 2.3) |
| Tamaño del código | 6793 líneas de Go de producción y 43 931 de tests en Go (6.5 por cada línea de producción), más 5503 líneas de escenarios | `git ls-files` sobre `a39592c` |

## 2. Métricas del proceso por Sprint

Las métricas M1 a M10 de [`docs/metricas.md`](../metricas.md) se definieron para los proyectos que gestiona la aplicación. Acá se aplican al propio proyecto: cada historia de usuario es un ítem, con los Story Points de su Planning, y cada Sprint del equipo es un sprint.

| Sprint | Período | Historias |
|---|---|---|
| Sprint 0 — La Preparación | 07/09 al 08/10/2026 | Sin historias: equipo, entorno, tablero y Product Backlog inicial |
| Sprint 1 — El MVP | 09/10 al 15/10/2026 | HU-01 a HU-06 (HU-05 opcional) |
| Sprint 2 — La Interfaz | 16/10 al 22/10/2026 | HU-07 a HU-13 |
| Sprint 3 — La Estimación | 23/10 al 26/10/2026 (cerrado antes de lo previsto) | HU-18, HU-20 a HU-23, HU-25 a HU-27, HU-29, HU-32 y HU-34 |
| Sprint 4 — La Entrega | 27/10 al 02/11/2026 (desarrollo hasta el 30/10) | HU-14 a HU-17, HU-19, HU-24, HU-28, HU-30, HU-31, HU-33, HU-35 a HU-41 y HU-47 |

### 2.1 Story Points y velocidad (M1, M2 y M3)

| Sprint | Comprometidos en el Planning | Planificados en el Sprint (M1) | Completados (M2) | Historias completadas | Fuente |
|---|---|---|---|---|---|
| Sprint 1 | 21 | 21 | **21** | 6 | [Review del Sprint 1](../actas/sprint-1/review.md) |
| Sprint 2 | 25 | 25 | **25** | 7 | [Review del Sprint 2](../actas/sprint-2/review.md) |
| Sprint 3 | 29 | 34 (con HU-32 y HU-34, adelantadas del Sprint 4) | **34** | 11 | [Review del Sprint 3](../actas/sprint-3/review.md) |
| Sprint 4 | 35 | 59 (con 8 historias tomadas con capacidad sobrante) | **59** | 18 | [Review del Sprint 4](../actas/sprint-4/review.md) |
| **Proyecto** | **110** | **139** | **139** | **42** | |

- **Velocidad (M3):** promedio de M2 de los últimos 3 Sprints cerrados: (25 + 34 + 59) / 3 = **39.3** Story Points por Sprint. Con los cuatro Sprints, el promedio es 34.8.
- **Tendencia:** la velocidad creció en cada Sprint (21 → 25 → 34 → 59). Ningún Sprint dejó Story Points comprometidos sin completar.
- Las tareas técnicas (TEC-01 a TEC-08) no se estiman en Story Points (decisión 5 del acta de Planning del Sprint 1).
- Los Story Points de HU-32, HU-34, HU-19 y HU-28 los estimó el Agile Enabler al aprobar sus SDD, y el equipo los confirmó después (Planning del Sprint 4 y Review del Sprint 4).

### 2.2 Porcentaje de historias completadas (M8)

| Sprint | Historias del Sprint | Completadas | M8 |
|---|---|---|---|
| Sprint 1 | 6 | 6 | 100 % |
| Sprint 2 | 7 | 7 | 100 % |
| Sprint 3 | 11 | 11 | 100 % |
| Sprint 4 | 18 | 18 | 100 % |
| **Proyecto** | **42** | **42** | **100 %** |

Fuente: actas de Review e Issues con la etiqueta `tipo:historia` (42 cerrados, 0 abiertos). Además se completaron las ocho tareas técnicas, salvo TEC-08 (este informe), que se cierra con su Pull Request.

### 2.3 Horas estimadas, horas reales y desviación (M4 a M7)

**Horas estimadas (M4):** el equipo estimó en Story Points y no registró una estimación en horas de sus historias. Según la regla de datos faltantes de `docs/metricas.md` ("M4 = 0 → M6 y M7 son no disponibles"), **la desviación en horas (M6) y la porcentual (M7) del proceso son no disponibles**. Se informan las horas reales y su relación con los Story Points.

**Horas reales (M5):** desde el Sprint 1, cada integrante registra en la Daily las horas del día anterior. La tabla suma las horas publicadas en las Dailys de cada Sprint.

| Sprint | Días con horas | Peñalbé Hernán | Buttini Cristobal | Cardozo Leandro | García Santiago | Sosa Ricardo | **Total** | Horas por Story Point |
|---|---|---|---|---|---|---|---|---|
| Sprint 1 | 10/10 al 14/10 | 28.5 | 9.0 | 7.0 | 10.0 | 7.0 | **61.5** | 2.9 |
| Sprint 2 | 15/10 al 21/10 | 32.0 | 11.5 | 10.5 | 11.5 | 13.5 | **79.0** | 3.2 |
| Sprint 3 | 22/10 al 25/10 | 10.5 | 4.0 | 9.0 | 5.0 | 4.0 | **32.5** | 1.0 |
| Sprint 4 | 26/10 al 29/10 | 25.0 | 13.0 | 21.5 | 12.0 | 15.0 | **86.5** | 1.5 |
| **Proyecto** | | **96.0** | **37.5** | **48.0** | **38.5** | **39.5** | **259.5** | **1.9** |

Fuente: actas de Daily ([Sprint 1](../actas/sprint-1/daily.md), [Sprint 2](../actas/sprint-2/daily.md), [Sprint 3](../actas/sprint-3/daily.md) y [Sprint 4](../actas/sprint-4/daily.md)). Cuando un integrante publicó un total del día y su detalle, se toma el total. "Horas por Story Point" divide el total del Sprint por M2.

**Límites de este dato**, por los que las horas reales están por debajo de las trabajadas:

- La Daily registra las horas del día anterior, así que el último día de cada Sprint no tiene horas (por ejemplo, el 30/10, cierre del desarrollo y día de la Review y la Retrospectiva).
- Quedaron horas sin completar: el 10/10 (Peñalbé Hernán, sin horas), el 21/10 (Sosa Ricardo, una línea con `⟨~ 0,5 h…⟩`) y el 24/10 (Peñalbé Hernán, dos líneas con `⟨completar⟩`). No se suman.
- En el Sprint 3 la mayoría de los registros son de "~1 h" por historia, lo que explica 1.0 hora por Story Point frente a 2.9 y 3.2 de los Sprints 1 y 2. La Retrospectiva del Sprint 3 ya señaló la Daily incompleta como mejora.
- El Sprint 0 no registró horas.

### 2.4 Defectos detectados y resueltos (M9 y M10)

Se cuentan como defectos los errores de comportamiento que llegaron a un commit y se corrigieron con un commit `fix` (bug con TDD, como pide `AGENTS.md`). Los problemas que las revisiones encontraron antes del commit, o que pidieron como "cambios" en un Pull Request, se cuentan aparte, en la sección 7.

| Sprint | Defecto | Dónde se detectó | Corrección | Estado |
|---|---|---|---|---|
| Sprint 1 | El paso BDD "hay N proyectos guardados" (HU-01) solo aceptaba el plural, así que un escenario con un solo proyecto tenía que decir "1 proyectos" | Revisión del PR #68 | RED `19ffa5b`, GREEN `fa6e9ac` (PR #72) | Resuelto |
| Sprint 4 | El botón "Descargar PDF" (HU-40) liberaba el enlace del archivo en el mismo paso del clic, lo que en algunos navegadores puede cancelar la descarga | Revisión del PR #151, antes del merge | RED `584b907`, GREEN `52a0498` | Resuelto |
| Sprint 4 | La nota de desviación del gráfico de horas (HU-37) se cortaba en el borde del gráfico | Al armar las capturas del manual (TEC-07), después del merge | RED `56fea15`, GREEN `1b2bbaa` (PR #162) | Resuelto |

| Sprint | Detectados (M9) | Resueltos (M10) |
|---|---|---|
| Sprint 1 | 1 | 1 |
| Sprint 2 | 0 | 0 |
| Sprint 3 | 0 | 0 |
| Sprint 4 | 2 | 2 |
| **Proyecto** | **3** | **3** |

Fuente: los commits `fix` de `main` (`git log --no-merges --format=%s | grep '^fix'`), las evidencias de HU-01, HU-40 y HU-37, y los PR citados. Los dos commits `fix(HU-22)` de la revisión del PR #125 (`7647967` y `8e828d9`) agregan tests y limpian comentarios pedidos en la revisión; no corrigen un comportamiento, así que no se cuentan como defectos. **Dos defectos llegaron a `main`** (el de HU-01, corregido en el PR #72, y el del gráfico de HU-37), y los dos se corrigieron con TDD.

## 3. Cobertura de pruebas

### 3.1 Cobertura final por paquete

Medida sobre `main` (`a39592c`) con `go test -count=1 -cover ./...`. La columna "Sin base" es como la mide el CI, que no tiene PostgreSQL y omite esas pruebas. La columna "Con base" agrega `TEST_DATABASE_URL`, con la base recreada desde `db/init/01-schema.sql` y `-p 1`, sin ningún test omitido.

| Paquete | Qué tiene | Sin base | Con base |
|---|---|---|---|
| `internal/metrics` | Métricas M1 a M10 | 100.0 % | 100.0 % |
| `internal/dashboard` | Resumen del estado del proyecto | 100.0 % | 100.0 % |
| `internal/report` | Reportes del proyecto y del sprint, y el PDF | 100.0 % | 100.0 % |
| `internal/httpapi` | Rutas HTTP, JSON y errores | 98.7 % | 98.7 % |
| `internal/backlog` | Product Backlog | 83.3 % | 99.0 % |
| `internal/project` | Proyectos y sus fechas | 74.8 % | 98.1 % |
| `internal/sprint` | Sprints | 77.1 % | 97.2 % |
| `internal/effort` | Registro de esfuerzo | 74.2 % | 96.9 % |
| `internal/member` | Integrantes | 68.8 % | 96.0 % |
| `internal/defect` | Defectos | 77.7 % | 95.6 % |
| `internal/poker` | Planning Poker | 75.5 % | 95.3 % |
| `internal/postgres` | Apertura del pool de PostgreSQL | 83.3 % | 83.3 % |
| `internal/postgrestest` | Preparación de la base de pruebas (solo tests) | 36.8 % | 36.8 % |
| `cmd/server` | Arranque: configuración y armado de los Manager | 29.8 % | 31.6 % |
| `internal/web` | HTML, JS y CSS embebidos | 0.0 % | 0.0 % |

| Total del perfil | Sin base | Con base |
|---|---|---|
| Sin `-coverpkg` (cada paquete con sus propios tests) | **84.5 %** | **95.2 %** |
| Con `-coverpkg=./...` (cuenta lo que un paquete ejecuta desde los tests de otro) | — | **96.2 %** |

- **Lo que baja la cobertura sin base** son los repositorios de PostgreSQL de cada dominio, que solo se ejecutan con `TEST_DATABASE_URL`. Con base, los once paquetes de dominio y de la API superan el 95 %.
- **`cmd/server`** no llega al 100 % porque `main()` arranca el servidor y no se prueba con tests unitarios; la configuración, el armado de los repositorios y los timeouts sí tienen tests (`main_test.go`).
- **`internal/web`** figura en 0.0 % porque no tiene tests propios: su `Handler` y el contenido de `index.html` y `app.js` los prueban los 69 tests de `internal/httpapi/web_test.go`. El dibujo de la interfaz se verificó en el navegador (evidencias de TEC-04, HU-37 y TEC-07).
- **`internal/postgrestest`** solo existe para los tests: sus ramas de error (base que no termina en `_test`, conexión fallida) no se ejecutan en una corrida normal.

### 3.2 Evolución a lo largo de los Sprints

Para comparar con el mismo método, cada Sprint se volvió a medir en su línea base (tags `sprint-1` a `sprint-3`, y `main` para el Sprint 4), con los mismos comandos de la sección 3.1.

| Sprint | Línea base | Cobertura sin base | Cobertura con base (`-coverpkg`) | Archivos `.feature` | Escenarios ejecutados | Funciones de test | Líneas de producción | Líneas de tests en Go |
|---|---|---|---|---|---|---|---|---|
| Sprint 1 | `sprint-1` (`2058ed9`) | 95.6 % | — (sin PostgreSQL todavía) | 6 | 81 | 106 | 748 | 4281 |
| Sprint 2 | `sprint-2` (`5939950`) | 79.7 % | 94.4 % | 13 | 194 | 305 | 2309 | 12 830 |
| Sprint 3 | `sprint-3` (`23dcad8`) | 80.9 % | 94.8 % | 24 | 346 | 609 | 4290 | 24 486 |
| Sprint 4 | `main` (`a39592c`) | 84.5 % | 96.2 % | 42 | 656 | 1108 | 6793 | 43 931 |

- **Sprint 1 → Sprint 2:** la cobertura sin base bajó de 95.6 % a 79.7 % porque TEC-05 agregó los repositorios de PostgreSQL, que el CI no ejecuta. Con base, el Sprint 2 queda en 94.4 %.
- **Sprints 2 a 4:** las dos coberturas subieron en cada Sprint, aunque el código de producción casi se triplicó (de 2309 a 6793 líneas).
- **Diferencia con las actas:** la Review del Sprint 2 informó 92.5 % con base, medido en HU-11, antes del cierre; la nueva medición sobre el tag da 94.4 %. Las demás coinciden con las actas de Review.
- El detalle por historia, con cada medición y su comparación con `main`, está en [`cobertura.md`](cobertura.md).

## 4. Pruebas automatizadas

### 4.1 Tests por tipo

| Tipo | Dónde | Funciones de test | Archivos |
|---|---|---|---|
| Unitarios de dominio | `internal/<dominio>/*_test.go` | 649 | 48 |
| API HTTP (`httptest`) | `internal/httpapi/*_test.go` | 271 | 19 |
| Interfaz web (contenido de HTML y JS) | `internal/httpapi/web_test.go` | 69 | 1 |
| Integración con PostgreSQL | `internal/<dominio>/postgres_repository_test.go` y `cmd/server` | 50 | 9 |
| Contrato de repositorios (memoria y PostgreSQL) | `internal/<dominio>/*contract_test.go` | 11, con un subtest por caso | 8 |
| Arranque (`cmd/server`) | `cmd/server/*_test.go` | 16 | 1 |
| Runners de escenarios BDD (godog) | `internal/<dominio>/*_features_test.go` | 42, uno por historia | 42 |
| **Total** | | **1108** | **128** |

Con PostgreSQL, `go test -v ./...` informa 2591 casos en verde (funciones y subtests) y 0 omitidos. Todos usan el paquete estándar `testing`; la única librería de tests es godog, aprobada por el equipo (`AGENTS.md`).

### 4.2 Escenarios BDD por épica

Los escenarios están escritos en Gherkin en español (`features/HU-XX-*.feature`) y godog los ejecuta en modo estricto con `go test ./...`. Un "Esquema del escenario" se ejecuta una vez por fila de "Ejemplos", por eso hay más escenarios ejecutados que escritos.

| Épica | Historias | Escenarios | Esquemas | Ejecutados |
|---|---|---|---|---|
| E1 · Proyectos | HU-01 a HU-03 | 33 | 4 | 43 |
| E2 · Product Backlog | HU-04 a HU-09 | 52 | 14 | 90 |
| E3 · Sprints | HU-10 a HU-14 | 54 | 6 | 71 |
| E4 · Integrantes | HU-15 a HU-17 | 15 | 10 | 46 |
| E5 · Fechas del proyecto | HU-18, HU-19 y HU-41 | 26 | 8 | 48 |
| E6 · Estimación | HU-20 a HU-25 | 54 | 7 | 78 |
| E7 · Esfuerzo | HU-26 a HU-28 | 37 | 2 | 45 |
| E8 · Defectos | HU-29 a HU-31 | 39 | 11 | 72 |
| E9 · Métricas | HU-32 a HU-35 y HU-47 | 42 | 16 | 88 |
| E10 · Dashboard | HU-36 y HU-37 | 18 | 3 | 27 |
| E11 · Reportes | HU-38 a HU-40 | 30 | 5 | 48 |
| **Total** | **42 historias** | **400** | **86** | **656** |

Fuente: conteo de `Escenario:`, `Esquema del escenario:` y filas de `Ejemplos:` en `features/`; el total de ejecutados coincide con la suma de las líneas `N scenarios` de `go test -v ./internal/...`.

## 5. Trazabilidad: criterios de aceptación cubiertos

Cada historia sigue la cadena **Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go**. Cada Issue de historia tiene la sección "## Trazabilidad" con los links a su SDD, su `.feature` y su evidencia, y desde el Sprint 3 cada escenario BDD lleva en su nombre el criterio que cubre, por ejemplo "(CA-24.4)".

| Épica | Criterios de aceptación | Con escenario BDD | Con tests de Go |
|---|---|---|---|
| E1 · Proyectos | 19 | 18 | 1 |
| E2 · Product Backlog | 47 | 43 | 4 |
| E3 · Sprints | 39 | 35 | 4 |
| E4 · Integrantes | 20 | 20 | 0 |
| E5 · Fechas del proyecto | 31 | 28 | 3 |
| E6 · Estimación | 55 | 44 | 11 |
| E7 · Esfuerzo | 28 | 26 | 2 |
| E8 · Defectos | 29 | 29 | 0 |
| E9 · Métricas | 38 | 35 | 3 |
| E10 · Dashboard | 12 | 12 | 0 |
| E11 · Reportes | 23 | 23 | 0 |
| **Total** | **341** | **313** (91.8 %) | **28** (8.2 %) |

Fuente: los criterios de la sección "Criterios de aceptación" de cada SDD (`docs/especificaciones/`), buscados en su `.feature`; los 28 restantes, en los tests de Go y en la tabla de trazabilidad de cada evidencia (`docs/evidencias/`).

**Los 28 criterios sin escenario BDD** no describen un comportamiento del dominio que godog pueda recorrer con sus pasos. Son de cuatro tipos, y cada uno se verifica con tests de Go:

- **Persistencia:** "se guarda igual en memoria y en PostgreSQL" (por ejemplo, CA-22.8), con los tests de contrato de los repositorios.
- **API e interfaz:** la ruta HTTP y su sección en la interfaz (por ejemplo, CA-33.5), con los tests de `internal/httpapi`.
- **Fallas del almacenamiento y concurrencia:** "si falla el guardado, no cambia nada" (por ejemplo, CA-02.8 y CA-25.6), con repositorios que fallan a pedido y tests con `-race`.
- **Copias por valor:** "modificar lo devuelto no altera lo guardado" (por ejemplo, CA-21.3), con tests unitarios.

Lista completa: CA-02.8, CA-04.8, CA-05.5, CA-06.7, CA-07.7, CA-10.8, CA-11.8, CA-12.8, CA-14.6, CA-18.2, CA-18.9, CA-18.10, CA-21.3, CA-22.8, CA-23.5, CA-23.8, CA-23.9, CA-24.11, CA-24.12, CA-25.6, CA-25.7, CA-25.9, CA-25.10, CA-26.11, CA-27.10, CA-33.5, CA-33.9 y CA-33.10.

**Ningún criterio de aceptación quedó sin una prueba automatizada.** Las pruebas en el navegador (evidencias de TEC-04, HU-37 y TEC-07) completan la verificación del dibujo de la interfaz, que los tests de Go no ejecutan.

## 6. Evidencias de aplicación de TDD

Cada historia tiene su evidencia en [`docs/evidencias/HU-XX-tdd.md`](../evidencias/), con la salida del RED, la del GREEN, la cobertura, los hashes de cada commit y las desviaciones del ciclo. Los commits siguen la convención `test(HU-XX): … (RED)`, `feat(HU-XX): … (GREEN)` y `refactor(HU-XX): …`.

### 6.1 Commits por tipo y por integrante

| Integrante | `test` | `feat` | `fix` | `refactor` | `docs` | Otros | Total |
|---|---|---|---|---|---|---|---|
| Peñalbé Hernán | 21 | 12 | 3 | 2 | 75 | 13 | 126 |
| Cardozo Leandro | 15 | 11 | 0 | 0 | 37 | 2 | 65 |
| Sosa Ricardo | 12 | 13 | 0 | 0 | 36 | 0 | 61 |
| García Santiago | 15 | 9 | 0 | 1 | 29 | 0 | 54 |
| Buttini Cristobal | 10 | 9 | 2 | 1 | 31 | 0 | 53 |
| **Total** | **73** | **54** | **5** | **4** | **208** | **15** | **359** |

Fuente: `git log main --no-merges`, agrupado por el tipo del título y por autor. "Otros" son `ci` (5, TEC-02), `chore` (5) y los 5 primeros commits del Sprint 0, anteriores a la convención. Los `docs` incluyen las SDD, las evidencias, la cobertura, las actas y la documentación final. Hay más `test` que `feat` porque varias historias tienen más de un ciclo RED → GREEN, y por los tests que pidieron las revisiones.

### 6.2 Orden de los commits por historia

**En las 42 historias, el primer commit RED es anterior al primer commit GREEN.** 30 historias tienen exactamente un ciclo RED → GREEN. Las otras 12 suman ciclos, refactors o tests por la revisión:

| Historia | Secuencia en `main` | Motivo |
|---|---|---|
| HU-01 | RED → GREEN → refactor → RED → fix → test → test | Revisión del PR #68 y seguimiento del Issue #73 (recorte de espacios Unicode) |
| HU-15 | RED → test → GREEN | Dos tests del RED corregidos antes del GREEN, en un commit propio |
| HU-22 | RED → GREEN → test → fix → fix | Tests y ajustes pedidos en la revisión del PR #125 |
| HU-26 | RED → GREEN → RED → GREEN → test | Un ciclo para el dominio y otro para PostgreSQL; después, una aserción de zona horaria corregida en el test |
| HU-29 | RED → GREEN → RED → GREEN → RED → GREEN → feat | Un ciclo por capa (dominio, API e interfaz); la persistencia en PostgreSQL entró en un solo commit, sin RED propio |
| HU-33 | RED → test → test → test → GREEN → test | Tres tests del RED corregidos antes del GREEN, y tests agregados después por la revisión, registrados como desviación |
| HU-37 | RED → GREEN → RED → fix | Corrección de un defecto con TDD (sección 2.4) |
| HU-40 | RED → test → GREEN → RED → fix | Dos tests del RED corregidos antes del GREEN, y la corrección de un defecto con TDD (sección 2.4) |
| HU-02 y HU-28 | RED → GREEN → test | Un test agregado después del GREEN |
| HU-14 y HU-39 | RED → GREEN → refactor | Refactor después del GREEN |

Fuente: `git log main --no-merges --reverse`, agrupado por el ID del título.

### 6.3 Desviaciones del ciclo documentadas

Las desviaciones de cada historia están en su evidencia y se resumen en las actas de Review. Las más frecuentes:

- **Tests agregados después del GREEN** por la revisión (HU-02, HU-22, HU-28 y HU-33). Desde el Sprint 4 se registran como desviación en la evidencia (acción 2 de la Retrospectiva del Sprint 3).
- **Tests corregidos antes del GREEN** (HU-15, HU-33 y HU-40) **o después** (HU-26): eran errores del test, no cambios de comportamiento.
- **Armado adelantado en el RED** (HU-21, HU-15 y HU-47): el cableado de `cmd/server` o la firma de un constructor entraron en el RED para que los tests compilen, como pedía cada SDD.
- **Interfaz sin tests automáticos de dibujo** (TEC-04 y HU-37): se cubre con tests de contenido y pruebas en el navegador.

## 7. Gestión del proyecto

### 7.1 Pull Requests y revisiones

| Integrante | Pull Requests de su autoría | Aprobaciones dadas |
|---|---|---|
| Peñalbé Hernán | 46 | 11 |
| Cardozo Leandro | 10 | 31 |
| Buttini Cristobal | 9 | 13 |
| García Santiago | 9 | 22 |
| Sosa Ricardo | 8 | 4 |
| **Total** | **82** | **81** |

- **Regla del equipo:** todo entra a `main` por Pull Request, con la CI en verde, la aprobación de un integrante distinto del autor y "Create a merge commit" (nunca squash ni rebase), para conservar el historial RED → GREEN → REFACTOR.
- **80 de 82** Pull Requests tienen la aprobación de otro integrante. Los dos sin aprobación son del cierre del Sprint 0 (#59 y #60, 08/10), antes de que empezara el desarrollo.
- **En 6 Pull Requests** el revisor pidió cambios antes de aprobar ("Request changes"), entre ellos HU-31, HU-33 y HU-35 en el Sprint 4. En muchos otros los cambios se pidieron en comentarios.
- **Peñalbé Hernán** es autor de 46 PR porque, como Agile Enabler, también subió las actas, la documentación técnica y las tareas técnicas.

Fuente: API de GitHub (`pulls` y `pulls/{n}/reviews`), sobre los 82 PR integrados.

### 7.2 Integración continua

- **152 ejecuciones** del workflow "CI" ([`.github/workflows/ci.yml`](../../.github/workflows/ci.yml), job "gofmt y tests", TEC-02) en Pull Requests: **151 en verde**. La única fallida fue la primera versión del workflow (10/10), en la rama de TEC-02, antes de integrarse. Además hubo una ejecución del agente de Copilot que armó esa primera versión, que no es del CI.
- En cada Pull Request, el CI verifica `gofmt -l`, corre `go vet ./...` y `go test -race -coverprofile=coverage.out ./...`, y muestra la cobertura total.
- **Límite:** el CI no levanta PostgreSQL, así que esos tests se omiten allí. Los revisores los corrieron a mano con base en varias historias (HU-16, HU-24, HU-28, HU-40, entre otras). Sumar PostgreSQL al CI es la acción 8 de la [Retrospectiva del Sprint 4](../actas/sprint-4/retrospectiva.md).

Fuente: API de GitHub Actions (`actions/runs`).

### 7.3 Issues y ceremonias

| Tipo | Issues | Cerrados |
|---|---|---|
| Historias de usuario (`tipo:historia`) | 42 | 42 |
| Épicas | 13 (E1 a E11, más dos sin etiqueta de épica) | 13 |
| Tareas técnicas (`tipo:tecnica`) | 9 (TEC-01 a TEC-08 y el seguimiento #73 de HU-01) | 7; TEC-06 y TEC-08 se cierran con el PR de este informe |
| Ceremonias (`tipo:scrum`) | 15 | 15 |

Cada ceremonia de los Sprints 1 a 4 tiene su acta en [`docs/actas/`](../actas/) (Planning, Daily, Review y Retrospectiva) y el Sprint 0, su acta de cierre. Fuente: API de GitHub (`issues`).

## 8. Análisis de los resultados

### Qué se cumplió

- **Alcance completo:** las 42 historias del Product Backlog y los nueve requerimientos mínimos de la consigna, incluidas las historias que en los Sprints 3 y 4 figuraban "fuera de la entrega" (E4, HU-19 y HU-28).
- **Sin deuda de compromiso:** cada Sprint completó el 100 % de los Story Points comprometidos (M8 = 100 % en los cuatro).
- **Calidad medible:** 96.2 % de cobertura con PostgreSQL, todos los criterios de aceptación con prueba automatizada, el RED antes del GREEN en todas las historias y dos defectos que llegaron a `main` (HU-01 y HU-37), corregidos con TDD.
- **Proceso trazable:** cada historia tiene SDD aprobada, `.feature`, evidencia TDD, Trazabilidad en su Issue y un Pull Request revisado; cada ceremonia tiene su Issue y su acta.

### Qué se desvió y por qué

- **Velocidad creciente, no estable (21 → 59):** el equipo aprendió el flujo SDD → BDD → TDD en los Sprints 1 y 2, y desde el Sprint 3 trabajó en paralelo con plantillas, el checklist de `docs/antes-de-empezar.md` y revisiones más rápidas. En el Sprint 4, además, las historias pendientes eran más chicas y reutilizaban lo ya hecho (por ejemplo, HU-17 sobre el repositorio de HU-15).
- **Sprint 3 más corto:** se replanificó el 24/10 (entrega el 02/11, desarrollo hasta el 30/10) y el Sprint se cerró el 26/10, con todo el compromiso integrado.
- **Desviación en horas no disponible:** el equipo estimó en Story Points y no en horas, así que M6 y M7 del proceso no se pueden calcular. Las horas reales registradas (259.5 h) están por debajo de las trabajadas, por las Dailys incompletas y el último día de cada Sprint sin registro (sección 2.3).
- **Cobertura sin base por debajo de la con base (84.5 % frente a 96.2 %):** el CI no tiene PostgreSQL, así que los repositorios de PostgreSQL solo se prueban a mano. Es la acción 8 de la Retrospectiva del Sprint 4.
- **Cierre de los Issues:** cinco historias del Sprint 4 quedaron con los criterios sin tildar en su Issue, aunque el código y las pruebas están completos (acciones 1 a 4 de la Retrospectiva del Sprint 4).
- **Dos defectos que llegaron a `main`:** el paso BDD de HU-01 se corrigió en el PR #72, después del merge del PR #68; el del gráfico de HU-37 se detectó al armar el manual de usuario, porque el dibujo de los gráficos no tiene pruebas automatizadas en el navegador.

## 9. Cómo volver a medir

```bash
# Cobertura por paquete, sin base (como el CI)
go test -count=1 -cover ./...

# Cobertura total sin base
go test -count=1 -coverprofile=cover.out ./...
go tool cover -func=cover.out | tail -n 1

# Con PostgreSQL (docker compose up -d --wait), sin tests omitidos
TEST_DATABASE_URL="postgres://gpc:gpc_dev@localhost:5433/gpc_test?sslmode=disable" go test -count=1 -p 1 -cover ./...
TEST_DATABASE_URL="postgres://gpc:gpc_dev@localhost:5433/gpc_test?sslmode=disable" go test -count=1 -p 1 -coverpkg=./... -coverprofile=cover.out ./...
go tool cover -func=cover.out | tail -n 1

# Escenarios BDD ejecutados
go test -count=1 -v ./internal/... | grep -E '^[0-9]+ scenarios'

# Detector de carreras (necesita cgo; lo corre el CI)
go test -race -count=1 ./...

# Commits por tipo (evidencia de TDD)
git log main --no-merges --format=%s | sed -E 's/^([a-z]+).*/\1/' | sort | uniq -c

# Orden RED → GREEN de una historia
git log main --no-merges --reverse --format='%h %s' | grep 'HU-XX'

# Evolución: misma medición sobre la línea base de cada Sprint
git worktree add ../medicion sprint-2
```

La cobertura de un tag se mide con el esquema de ese tag (`db/init/01-schema.sql` de esa versión). Si la base local se creó antes de HU-15, hay que recrearla con `docker compose down -v`.
