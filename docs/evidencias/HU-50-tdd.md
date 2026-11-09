# Evidencia TDD · HU-50 Listar los proyectos

> Issue del repositorio: #183 · Vinculada a HU-03 (#16) · Primer cambio especificado con OpenSpec (TEC-10, #177)

- **Especificación (OpenSpec):** [`openspec/changes/hu-50-listar-proyectos/`](../../openspec/changes/hu-50-listar-proyectos/): [propuesta](../../openspec/changes/hu-50-listar-proyectos/proposal.md), [spec de `proyectos`](../../openspec/changes/hu-50-listar-proyectos/specs/proyectos/spec.md), [diseño](../../openspec/changes/hu-50-listar-proyectos/design.md) y [tareas](../../openspec/changes/hu-50-listar-proyectos/tasks.md). Al archivarse pasa a `openspec/changes/archive/` y sus requisitos, a `openspec/specs/proyectos/spec.md`.
- **Aprobación:** Peñalbé Hernán (Agile Enabler) aprobó la propuesta el 08/11/2026. Antes pidió tres cosas, y se cumplieron:
  - que el orden por nombre lo aplique el `Manager` en Go, y no un `ORDER BY` en SQL (D2 y CA-50.6);
  - que la limitación de los acentos y la ñ quede escrita (riesgos del `design.md` y README);
  - que se acepten CA-50.6 y CA-50.7.

  Con su aprobación se agregaron la comparación con el acta de Planning y la línea de aprobación a `proposal.md` (`86dc287`).
- **Escenarios BDD:** [`features/HU-50-listar-proyectos.feature`](../../features/HU-50-listar-proyectos.feature): 7 escenarios con 36 pasos.
- **Rama:** `feature/HU-50-listar-proyectos`, creada desde `main` `b75e589` (con TEC-09, PR #181).
- **Story Points:** 3, en el Planning del Sprint 5 (Issue #182, PR #184).
- **Criterios del Issue #183:** CA-50.1 a CA-50.7 reemplazaron a los del borrador (CA-50.1 a CA-50.5) el 08/11/2026, con la autorización del Agile Enabler.

## Trazabilidad

**15 tests nuevos (funciones `Test`) cubren los 7 escenarios BDD.** Se reparten así:

- 8 unitarios en `internal/project/list_test.go`;
- el runner de godog en `internal/project/list_features_test.go`, que ejecuta los 7 escenarios;
- 3 de API en `internal/httpapi/projects_list_test.go`;
- 3 de interfaz en `internal/httpapi/projects_list_web_test.go`.

La correspondencia no es 1 a 1:

- los escenarios cubren la consulta de dominio (el `Manager`);
- la ruta HTTP (CA-50.5), la interfaz (CA-50.3) y la concurrencia (CA-50.7) se cubren solo con tests de Go, como en HU-14 y HU-17;
- varios tests comparten criterio.

Cada test lleva arriba el comentario `// HU-50 · CA-50.N · Verifica que …`.

### Criterio → test → código

| Criterio | Escenarios BDD | Tests de Go (archivo) | Código que lo implementa (función y archivo) |
|---|---|---|---|
| CA-50.1 Todos los proyectos con ID, nombre, descripción, estado y fechas, con el formato de `GET /api/proyectos/{id}` | "Listar los proyectos con todos sus datos", "Un proyecto modificado aparece con sus datos nuevos", "Un único proyecto" (runner `TestFeatures_HU50ListarProyectos`, `internal/project/list_features_test.go`) | `TestList_DevuelveTodosLosProyectosConSusDatos`, `TestList_UnUnicoProyecto` (`internal/project/list_test.go`); `TestListarProyectos_DevuelveLosProyectosConElFormatoDeLaConsultaPorID` (`internal/httpapi/projects_list_test.go`) | `Manager.List` (`internal/project/list.go`); `server.listProjects` con `toProjectResponse` (`internal/httpapi/projects.go`) |
| CA-50.2 Sin proyectos, lista vacía y no un error (`[]`, no `null`) | "Sin proyectos, la lista está vacía y no es un error" (runner `TestFeatures_HU50ListarProyectos`, `internal/project/list_features_test.go`) | `TestList_SinProyectosDevuelveListaVaciaNoNil` (`internal/project/list_test.go`); `TestListarProyectos_SinProyectosRespondeArregloVacio` (`internal/httpapi/projects_list_test.go`) | `Manager.List`, que convierte `nil` en `[]Project{}` (`internal/project/list.go`); `server.listProjects`, que arma la respuesta con `make(…, 0, …)` (`internal/httpapi/projects.go`) |
| CA-50.3 Elegir el proyecto actual desde la lista en la interfaz, sin escribir su número | — | `TestInterfaz_ContextoTieneElBotonVerProyectosYLaTabla`, `TestInterfaz_ComentarioDelContextoCitaHU50`, `TestInterfaz_VerProyectosLlenaLaTablaYPermiteElegir` (`internal/httpapi/projects_list_web_test.go`) | El botón `#ver-proyectos` y la tabla `#lista-proyectos` (`internal/web/index.html`); el handler de `ver-proyectos` con el botón «Elegir» (`internal/web/app.js`) |
| CA-50.4 Consultar no modifica ningún proyecto, y modificar la lista devuelta no altera los guardados | "Modificar la lista devuelta no altera los proyectos guardados" (runner `TestFeatures_HU50ListarProyectos`, `internal/project/list_features_test.go`) | `TestList_ModificarLaListaDevueltaNoAlteraLosGuardados`, `TestList_NoModificaLosProyectos` (`internal/project/list_test.go`) | `Manager.List`, que ordena la copia que devuelve `Repository.List` (`internal/project/list.go`) |
| CA-50.5 Se consulta desde la API y la interfaz; si falla la lectura, 500 y el error en la interfaz | "Falla el almacenamiento al listar" (runner `TestFeatures_HU50ListarProyectos`, `internal/project/list_features_test.go`) | `TestList_ErrorDelRepositorioDevuelveElErrorSinLista` (`internal/project/list_test.go`); `TestListarProyectos_ErrorDelRepositorioResponde500` y los otros dos de API (`internal/httpapi/projects_list_test.go`); `TestInterfaz_VerProyectosLlenaLaTablaYPermiteElegir` (`internal/httpapi/projects_list_web_test.go`) | `Manager.List` (`internal/project/list.go`); `server.listProjects` → `writeError` y la ruta `GET /api/proyectos` (`internal/httpapi/projects.go` y `server.go`); `accion("msg-contexto", …)` en el handler de `ver-proyectos` (`internal/web/app.js`) |
| CA-50.6 Orden alfabético por nombre sin distinguir mayúsculas, aplicado por el `Manager` | "Los proyectos se ordenan por nombre sin distinguir mayúsculas" (runner `TestFeatures_HU50ListarProyectos`, `internal/project/list_features_test.go`) | `TestList_OrdenaPorNombreSinDistinguirMayusculas` (`internal/project/list_test.go`); `TestListarProyectos_DevuelveLosProyectosConElFormatoDeLaConsultaPorID` (`internal/httpapi/projects_list_test.go`) | `Manager.List`, con `slices.SortStableFunc` y `strings.ToLower` (`internal/project/list.go`) |
| CA-50.7 Consultas simultáneas sin condiciones de carrera (`go test -race`) | — | `TestList_ConcurrenteConCreateNoTieneCondicionesDeCarrera` (`internal/project/list_test.go`), con `-race` en Linux sobre `430f5d5` y en el CI | `Manager.List`, que lee bajo `m.mu` (`internal/project/list.go`) |

Conteos:

```
$ grep -cE '^[[:space:]]*Escenario:' features/HU-50-listar-proyectos.feature
7
$ grep -c '^func Test' internal/project/list_test.go internal/project/list_features_test.go internal/httpapi/projects_list_test.go internal/httpapi/projects_list_web_test.go
internal/project/list_test.go:8
internal/project/list_features_test.go:1
internal/httpapi/projects_list_test.go:3
internal/httpapi/projects_list_web_test.go:3
```

## Fase RED (`43c9216`)

Se escribieron los tests y el `.feature` antes del código. La única línea de producción del RED es la firma vacía de `Manager.List` en `internal/project/list.go`, como en los RED de HU-28 y HU-39:

```go
func (m *Manager) List() ([]Project, error) { return nil, nil }
```

Con la firma, `internal/project` compila y los tests fallan por aserción, no por "List undefined". Esta es la salida de `go test -count=1 ./...` sobre `43c9216`, filtrada a las líneas de resultado y de error. Se corrió en una copia temporal con `git worktree`:

```
ok  	seguimiento-medicion-gpc/cmd/server	0.339s
ok  	seguimiento-medicion-gpc/internal/backlog	0.858s
ok  	seguimiento-medicion-gpc/internal/dashboard	0.746s
ok  	seguimiento-medicion-gpc/internal/defect	0.741s
ok  	seguimiento-medicion-gpc/internal/effort	0.573s
--- FAIL: TestListarProyectos_DevuelveLosProyectosConElFormatoDeLaConsultaPorID (0.00s)
    projects_list_test.go:53: GET /api/proyectos: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    projects_list_test.go:55: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestListarProyectos_SinProyectosRespondeArregloVacio (0.00s)
    projects_list_test.go:77: GET /api/proyectos: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    projects_list_test.go:79: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestListarProyectos_ErrorDelRepositorioResponde500 (0.00s)
    projects_list_test.go:88: GET /api/proyectos: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    projects_list_test.go:90: código = 404, se esperaba 500; cuerpo: 404 page not found
--- FAIL: TestInterfaz_ContextoTieneElBotonVerProyectosYLaTabla (0.00s)
    projects_list_web_test.go:42: la sección de contexto no contiene "<button type=\"button\" id=\"ver-proyectos\" class=\"secundario\">Ver proyectos</button>"
    projects_list_web_test.go:42: la sección de contexto no contiene "<th>ID</th><th>Nombre</th><th>Descripción</th><th>Estado</th><th>Inicio</th><th>Fin</th><th>Acciones</th>"
    projects_list_web_test.go:42: la sección de contexto no contiene "<tbody id=\"lista-proyectos\"></tbody>"
--- FAIL: TestInterfaz_ComentarioDelContextoCitaHU50 (0.00s)
    projects_list_web_test.go:53: la página todavía dice que no hay listado de proyectos
    projects_list_web_test.go:56: la página no cita HU-50 en el comentario de la sección de contexto
--- FAIL: TestInterfaz_VerProyectosLlenaLaTablaYPermiteElegir (0.00s)
    projects_list_web_test.go:66: app.js no tiene el handler de "ver-proyectos"
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi	0.940s
ok  	seguimiento-medicion-gpc/internal/member	0.497s
ok  	seguimiento-medicion-gpc/internal/metrics	1.923s
ok  	seguimiento-medicion-gpc/internal/poker	1.273s
ok  	seguimiento-medicion-gpc/internal/postgres	0.615s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.643s
7 scenarios (7 failed)
36 steps (25 passed, 7 failed, 4 skipped)
--- FAIL: TestFeatures_HU50ListarProyectos (0.01s)
--- FAIL: TestList_DevuelveTodosLosProyectosConSusDatos (0.00s)
    list_test.go:61: List() = [], se esperaba [{ID:2 Name:Inventario … Status:EN_CURSO …} {ID:1 Name:Sistema de ventas … Status:PLANIFICADO …}]
--- FAIL: TestList_UnUnicoProyecto (0.00s)
    list_test.go:74: List() = [], se esperaba [{ID:1 Name:Agenda Description:Turnos Status:PLANIFICADO …}]
--- FAIL: TestList_SinProyectosDevuelveListaVaciaNoNil (0.00s)
    list_test.go:89: List() = nil, se esperaba una lista vacía
--- FAIL: TestList_ModificarLaListaDevueltaNoAlteraLosGuardados (0.00s)
    list_test.go:104: List() tiene 0 proyectos, se esperaba 1
--- FAIL: TestList_ErrorDelRepositorioDevuelveElErrorSinLista (0.00s)
    list_test.go:147: List() error = <nil>, se esperaba falla del almacenamiento
--- FAIL: TestList_OrdenaPorNombreSinDistinguirMayusculas (0.00s)
    list_test.go:166: nombres = [], se esperaba ["Agenda" "Beta" "inventario" "Sistema de ventas" "zeta"]
--- FAIL: TestList_ConcurrenteConCreateNoTieneCondicionesDeCarrera (0.00s)
    list_test.go:203: List() tiene 0 proyectos, se esperaban 20
FAIL
FAIL	seguimiento-medicion-gpc/internal/project	0.569s
ok  	seguimiento-medicion-gpc/internal/report	0.771s
ok  	seguimiento-medicion-gpc/internal/sprint	0.447s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

(En `TestList_DevuelveTodosLosProyectosConSusDatos` y `TestList_UnUnicoProyecto`, las fechas en cero de los valores esperados se resumieron con "…".)

Todos los fallos corresponden a la funcionalidad nueva, y todos son por aserción:

- **`internal/project`:** fallan 8 tests: 7 de `list_test.go` y el runner de godog, con los 7 escenarios en rojo. Los 46 tests de HU-01, HU-02 y HU-03 (`TestCreate_*`, `TestUpdate_*`, `TestGet_*` y sus runners de escenarios) pasan, igual que el resto del paquete.
- **`internal/httpapi`:** fallan los 6 tests nuevos, 3 de API y 3 de interfaz. La ruta `GET /api/proyectos` no existe y responde "404 page not found". La interfaz no tiene el botón, la tabla ni el handler en `app.js`, y su comentario todavía dice que no hay listado de proyectos.
- **Los otros 12 paquetes con tests:** pasan. Ningún test anterior falló.

**Un test de HU-50 pasa con la firma vacía:** `TestList_NoModificaLosProyectos` (CA-50.4). Verifica que los proyectos guardados son iguales antes y después de listar, y un `return nil, nil` no lee ni escribe el repositorio. Es un test de protección: no obliga a escribir código, sino que impide que una implementación futura modifique datos al listar. El otro test de CA-50.4, `TestList_ModificarLaListaDevueltaNoAlteraLosGuardados`, sí falla en el RED.

## Fase GREEN (`2386fb8`)

Código agregado:

- **`internal/project/list.go`:** `Manager.List()` toma el mutex del `Manager`, lee con `Repository.List()` (el mismo que usa `Create` de HU-01), convierte `nil` en `[]Project{}` y ordena con `slices.SortStableFunc` por `strings.ToLower(Name)`. Si el repositorio falla, devuelve `nil` y ese error.
- **`internal/httpapi/projects.go`:** el handler `listProjects` arma la respuesta con `make([]projectResponse, 0, len(projects))` y `toProjectResponse`, así que cada elemento tiene el formato de `GET /api/proyectos/{id}`. Los errores van por `writeError`, que responde 500.
- **`internal/httpapi/server.go`:** la ruta `GET /api/proyectos`, junto a `POST /api/proyectos`.
- **`internal/web/index.html`:** en la sección «Proyecto y sprint actuales», el botón «Ver proyectos» y la tabla `lista-proyectos` (ID, Nombre, Descripción, Estado, Inicio, Fin y Acciones), con los contenedores `barra` y `tabla` de HU-14 y HU-17. El comentario de la sección ahora cita HU-50.
- **`internal/web/app.js`:** el handler de `ver-proyectos` llena la tabla con `celda` y `boton` (`textContent`, sin `innerHTML`) e informa "N proyecto(s).". El botón «Elegir» escribe el ID en «Proyecto actual», llama a `cargarProyecto()` y muestra "Proyecto N elegido: Nombre.".

No cambian:

- `Repository`, los repositorios en memoria y PostgreSQL, ni el test de contrato;
- `create.go`, `get.go` ni `toProjectResponse`;
- `NewHandler` ni `cmd/server`.

El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat HEAD -- '*_test.go' '*.feature'     (antes del commit 2386fb8)
(sin salida)
```

`go test ./...` antes del commit del GREEN:

```
ok  	seguimiento-medicion-gpc/cmd/server	1.932s
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/dashboard	(cached)
ok  	seguimiento-medicion-gpc/internal/defect	(cached)
ok  	seguimiento-medicion-gpc/internal/effort	(cached)
ok  	seguimiento-medicion-gpc/internal/httpapi	2.222s
ok  	seguimiento-medicion-gpc/internal/member	(cached)
ok  	seguimiento-medicion-gpc/internal/metrics	1.669s
ok  	seguimiento-medicion-gpc/internal/poker	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/report	1.705s
ok  	seguimiento-medicion-gpc/internal/sprint	(cached)
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

Los escenarios de HU-50 (`go test -count=1 -run TestFeatures_HU50 -v ./internal/project/`):

```
7 scenarios (7 passed)
36 steps (36 passed)
ok  	seguimiento-medicion-gpc/internal/project	0.314s
```

Antes del commit se corrió `go vet ./...` y no informó problemas. `go test -count=1 -coverprofile=cover.out ./internal/project/ ./internal/httpapi/` y `go tool cover -func=cover.out` dieron `List` y `listProjects` en 100.0%:

```
ok  	seguimiento-medicion-gpc/internal/project	1.195s	coverage: 73.1% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	3.285s	coverage: 98.6% of statements
seguimiento-medicion-gpc/internal/httpapi/projects.go:81:		listProjects			100.0%
seguimiento-medicion-gpc/internal/project/list.go:11:			List				100.0%
```

### Uso desde el navegador

Se probó en el panel del navegador de la app de escritorio de Claude, con el servidor de la rama (`go run ./cmd/server`, datos en memoria, puerto 8080):

| Acción | Resultado |
|---|---|
| «Ver proyectos» sin proyectos | La tabla queda vacía y el mensaje dice "0 proyecto(s)." |
| Crear "Sistema de ventas", "inventario" y "Agenda" (en ese orden) y registrar la fecha de inicio 2026-03-01 de "inventario"; después, «Ver proyectos» | "3 proyecto(s).", con las filas `3 · Agenda · Turnos · PLANIFICADO`, `2 · inventario · Depósito · PLANIFICADO · 2026-03-01` y `1 · Sistema de ventas · Ventas online · PLANIFICADO` |
| «Elegir» en la fila de "inventario" | «Proyecto actual» pasa a 2, los datos del proyecto muestran "Proyecto 2: inventario (PLANIFICADO), inicio: 2026-03-01, sin fecha de fin" y el mensaje dice "Proyecto 2 elegido: inventario.". La consola no tiene errores. |

No se pudo provocar un error del servidor desde el navegador. El caso de error de la interfaz (CA-50.5) queda cubierto por el test del handler, que verifica que los errores van a `msg-contexto` con `accion`.

## Refactor

No hubo (tarea 4.1):

- `Manager.List` es una función corta con una sola responsabilidad, y `listProjects` sigue el patrón de `listMembers` (HU-17) y `listProjectSprints` (HU-14).
- Armar las filas de la tabla con una función común con HU-14 obligaría a tocar el código de HU-14 y archivos que va a modificar HU-51, sin ganar legibilidad.

## Desviaciones del ciclo TDD

1. **Un test se corrigió antes del commit del RED.** En la primera corrida con la firma vacía, `TestList_ModificarLaListaDevueltaNoAlteraLosGuardados` hacía `got[0]` sobre una lista vacía y entraba en pánico. El pánico cortaba el paquete y los tests siguientes no corrían. Se agregó un `t.Fatalf` cuando la lista no tiene el largo esperado. El test no cambió lo que verifica, y el commit del RED ya lo tiene corregido.
2. **Un test pasa en el RED:** `TestList_NoModificaLosProyectos`, explicado en la fase RED.
3. **Comentarios ajustados al formato de TEC-11 después del GREEN, sin cambios de código.**
   - **Qué cambió:** el comentario de cada uno de los 15 tests pasó a una sola línea, `// HU-50 · CA-50.N · Verifica que …`, con el criterio principal que prueba. Antes, algunos citaban varios criterios en varias líneas. El formato es el que fija `docs/evidencias/TEC-11-tests-por-historia.md`, que llegó a la rama con el merge de `main` (`100bb1e`) y pide un único criterio por test para el test de control.
   - **Qué no cambió:** el código, los nombres y el orden de los tests. `git diff` solo tiene líneas de comentario.
   - **Dónde quedan los otros criterios:** en la tabla "Criterio → test → código" de esta evidencia. Por ejemplo, `TestListarProyectos_DevuelveLosProyectosConElFormatoDeLaConsultaPorID` cita CA-50.1 y en la tabla figura también en CA-50.5 y CA-50.6, y el runner de godog figura en las filas de sus escenarios.

## Validación final

Se midió sobre `2386fb8` en Windows, con go1.27.2 y sin `TEST_DATABASE_URL`, así que se omitieron 35 tests de PostgreSQL. HU-50 no agrega ninguno, porque reutiliza `Repository.List`, que ya cubren los tests de contrato.

`-race` no se corrió en esta máquina, porque necesita cgo (`go: -race requires cgo`). Lo necesita `TestList_ConcurrenteConCreateNoTieneCondicionesDeCarrera` (CA-50.7).

**Verificación con `-race`:** se corrió en Linux sobre `430f5d5`. Los 14 paquetes con tests quedaron en verde y no se reportaron condiciones de carrera. Con eso se cumple CA-50.7 y se tildaron las tareas 3.1 y 3.4 de `tasks.md`. El CI lo vuelve a correr en el Pull Request.

**Ejecutables bloqueados por Windows.** En esta máquina, una política de Windows ("An Application Control policy has blocked this file") bloquea a veces el ejecutable de test de algún paquete: `member`, `poker`, `metrics` o `postgres`, según la corrida. Esos paquetes aparecen como `FAIL` sin que se ejecute ningún test, y no es un fallo del código. Cada corrida se repitió hasta que no hubo ningún bloqueo; las salidas de abajo son corridas completas, sin bloqueos.

`gofmt -w` sobre los archivos Go modificados (`internal/project/list.go`, `list_test.go`, `list_features_test.go`, `internal/httpapi/projects.go`, `server.go`, `projects_list_test.go` y `projects_list_web_test.go`) no cambió nada, y `gofmt -l .` no lista archivos. `go vet ./...` y `node --check internal/web/app.js` no informaron problemas.

`go test -count=1 ./...`:

```
ok  	seguimiento-medicion-gpc/cmd/server	0.314s
ok  	seguimiento-medicion-gpc/internal/backlog	0.933s
ok  	seguimiento-medicion-gpc/internal/dashboard	0.776s
ok  	seguimiento-medicion-gpc/internal/defect	0.751s
ok  	seguimiento-medicion-gpc/internal/effort	0.574s
ok  	seguimiento-medicion-gpc/internal/httpapi	1.172s
ok  	seguimiento-medicion-gpc/internal/member	0.680s
ok  	seguimiento-medicion-gpc/internal/metrics	0.983s
ok  	seguimiento-medicion-gpc/internal/poker	1.553s
ok  	seguimiento-medicion-gpc/internal/postgres	0.637s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.638s
ok  	seguimiento-medicion-gpc/internal/project	0.498s
ok  	seguimiento-medicion-gpc/internal/report	0.745s
ok  	seguimiento-medicion-gpc/internal/sprint	0.472s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

`go test -count=1 -cover ./...`:

```
ok  	seguimiento-medicion-gpc/cmd/server	0.370s	coverage: 48.6% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	0.793s	coverage: 79.9% of statements
ok  	seguimiento-medicion-gpc/internal/dashboard	0.819s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/defect	0.761s	coverage: 74.6% of statements
ok  	seguimiento-medicion-gpc/internal/effort	0.614s	coverage: 70.5% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	1.116s	coverage: 98.6% of statements
ok  	seguimiento-medicion-gpc/internal/member	0.534s	coverage: 67.3% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	0.900s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	0.536s	coverage: 72.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	1.953s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.693s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	0.524s	coverage: 73.1% of statements
ok  	seguimiento-medicion-gpc/internal/report	0.809s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	0.523s	coverage: 73.9% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

`go test -count=1 -coverprofile=cover.out ./...` y `go tool cover -func=cover.out`, con las funciones nuevas y el total:

```
seguimiento-medicion-gpc/internal/httpapi/projects.go:81:		listProjects			100.0%
seguimiento-medicion-gpc/internal/project/list.go:11:			List				100.0%
total:									(statements)			84.4%
```

**Comparación con `main` (`b75e589`):** se midió con el mismo Go y los mismos comandos, en una copia temporal con `git worktree`.

- `internal/project` sube de 71.1% a 73.1%.
- El total del perfil sube de 84.1% a 84.4%.
- Los demás paquetes no cambian, e `internal/httpapi` queda en 98.6% en los dos.

Se midió sin `-p 1`. Estos porcentajes son de este entorno (Windows, go1.27.2) y no se comparan con los de la tabla de `cobertura.md` (HU-49, Linux, go1.27.1). Por ejemplo, `cmd/server` da 48.6% acá y 29.8% allá, también en `main`. Por eso se compara solo con `main` medido en el mismo entorno. El detalle está en [`docs/reportes/cobertura.md`](../reportes/cobertura.md).

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| Propuesta (OpenSpec) | 51e8911 | `docs(HU-50): agregar la propuesta de OpenSpec` |
| Aprobación | 86dc287 | `docs(HU-50): registrar la aprobación de la propuesta` |
| RED | 43c9216 | `test(HU-50): agregar tests (RED)` |
| GREEN | 2386fb8 | `feat(HU-50): implementar la lista de proyectos (GREEN)` |
| Refactor | — | No hubo (ver "Refactor") |
| Evidencia y cobertura | 430f5d5 | `docs(HU-50): agregar evidencia TDD y cobertura` |
| Merge de `main` | 100bb1e | `chore(HU-50): traer main con el acta del Sprint 5 y TEC-11` |
| Verificación con `-race` | 9e28f6b | `docs(HU-50): registrar la verificación con -race` |
| Comentarios de los tests (TEC-11) | cb160d1 | `docs(HU-50): ajustar los comentarios de los tests al formato de TEC-11` |

El commit de cierre no se lista, porque un commit no puede contener su propio hash: queda registrado en el Pull Request y en el Issue #183. Los hashes de las filas se completaron a medida que se hizo cada commit.

HU-50: Historia → SDD (OpenSpec) → Criterios de Aceptación → BDD → Tests → Código Go
