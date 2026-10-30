# Evidencia TDD · HU-17 Consultar los integrantes de un proyecto

> Issue del repositorio: #31

- **Especificación SDD:** [`docs/especificaciones/HU-17-consultar-integrantes.md`](../especificaciones/HU-17-consultar-integrantes.md). La aprobó Peñalbé Hernán el 30/10/2026, sin ajustes, con D1 a D5 como estaban. Sugirió usar en `internal/member` un `project.Repository` que falle, como en los tests de `History` de HU-14, y así se hizo (`failingProjectRepository` en `list_test.go`).
- **Escenarios BDD:** [`features/HU-17-consultar-integrantes.feature`](../../features/HU-17-consultar-integrantes.feature). Tiene 8 definiciones (7 escenarios y 1 esquema de 3 ejemplos), que godog ejecuta como 10 escenarios con 63 pasos.
- **Rama:** `feature/HU-17-consultar-integrantes`, creada desde `main` `3edc9e9` (con HU-16, PR #158).
- **Story Points:** 3, estimados en el Planning del Sprint 2 ("Como HU-06").
- **Criterios del Issue #31:** CA-17.1 a CA-17.6 reemplazaron a los del borrador el 30/10/2026, después de la aprobación de la SDD.

## Trazabilidad

**17 tests nuevos (funciones `Test`) cubren los 10 escenarios BDD ejecutados (8 definiciones)**, y hay además una fila nueva en `TestRutasConID_IDNoNumericoDevuelve400`. Los tests se reparten así:

- 10 unitarios en `internal/member/list_test.go`;
- el runner de godog en `internal/member/list_features_test.go`;
- 4 de API en `internal/httpapi/members_list_test.go`;
- 2 de interfaz en `internal/httpapi/web_test.go`.

La correspondencia no es 1 a 1. Los escenarios cubren la consulta de dominio. La parte HTTP de CA-17.2 y CA-17.4, la interfaz de CA-17.6 y la concurrencia (RN-17.7) se cubren solo con tests de Go. Además, varios tests comparten criterio.

| Criterio | Escenarios BDD | Tests de Go |
|---|---|---|
| CA-17.1 Integrantes ordenados por ID, con nombre y rol | "Listar los integrantes de un proyecto con su nombre y su rol", "Un integrante modificado aparece con sus datos nuevos", "Se listan los integrantes de un proyecto finalizado" | `TestList_DevuelveLosIntegrantesOrdenadosPorIDConSuRol`, `TestList_MuestraLosDatosModificados`, `TestList_AceptaUnProyectoFinalizado`, `TestListarIntegrantes_Devuelve200ConLosIntegrantesOrdenados`, `TestInterfaz_VerIntegrantesLlamaALaAPIConElProyectoActual` |
| CA-17.2 Proyecto sin integrantes: lista vacía | "Un proyecto sin integrantes devuelve una lista vacía" | `TestList_ProyectoSinIntegrantesDevuelveListaVacia`, `TestListarIntegrantes_SinIntegrantesDevuelveListaVacia` |
| CA-17.3 Sin integrantes de otros proyectos, aunque tengan el mismo nombre | "No aparecen los integrantes de otros proyectos, aunque tengan el mismo nombre" | `TestList_SoloDevuelveLosIntegrantesDelProyecto`, `TestListarIntegrantes_Devuelve200ConLosIntegrantesOrdenados` |
| CA-17.4 Proyecto inexistente (404) e ID no numérico (400) | "Consultar los integrantes de un proyecto inexistente falla" (3 ejemplos: 99, 0 y -1) | `TestList_ProyectoInexistenteDevuelveErrProjectNotFound`, `TestListarIntegrantes_ProyectoInexistenteDevuelve404`, la fila "listar integrantes" de `TestRutasConID_IDNoNumericoDevuelve400` |
| CA-17.5 La lista es una copia | "Modificar la lista devuelta no cambia los integrantes guardados" | `TestList_ModificarLaListaNoCambiaLosIntegrantesGuardados` |
| CA-17.6 Interfaz: botón y tabla con ID, nombre y rol legible | — | `TestInterfaz_SeccionIntegrantesTieneElBotonVerIntegrantes`, `TestInterfaz_VerIntegrantesLlamaALaAPIConElProyectoActual` |
| RN-17.7 La lectura va bajo el mutex del `Manager` | — | `TestList_ConsultasSimultaneasConRegistros` (con `-race` en el CI) |
| RN-17.8 Una falla de lectura corta sin lista parcial | "Una falla del repositorio de integrantes se devuelve sin lista parcial" | `TestList_ErrorAlLeerElProyectoSeDevuelve`, `TestList_ErrorAlLeerLosIntegrantesSeDevuelve`, `TestListarIntegrantes_ErrorDelRepositorioDevuelve500SinDetalle` |

Conteos:

```
$ grep -cE '^[[:space:]]*Escenario:' features/HU-17-consultar-integrantes.feature
7
$ grep -cE '^[[:space:]]*Esquema del escenario:' features/HU-17-consultar-integrantes.feature
1
$ grep -c '^func Test' internal/member/list_test.go internal/member/list_features_test.go internal/httpapi/members_list_test.go
internal/member/list_test.go:10
internal/member/list_features_test.go:1
internal/httpapi/members_list_test.go:4
$ git diff 3edc9e9 c112570 -- internal/httpapi/web_test.go | grep -c '^+func Test'
2
```

## Fase RED (`3a075b2`)

Se escribieron los tests y el `.feature` antes del código. El RED no agrega código fuera de los tests y el `.feature`: `List` no existe, así que `internal/member` no compila. `go test ./...` sobre `3a075b2`, filtrado a las líneas de resultado y de error:

```
ok  	seguimiento-medicion-gpc/cmd/server	(cached)
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/dashboard	(cached)
ok  	seguimiento-medicion-gpc/internal/defect	(cached)
ok  	seguimiento-medicion-gpc/internal/effort	(cached)
# seguimiento-medicion-gpc/internal/member [seguimiento-medicion-gpc/internal/member.test]
internal\member\list_features_test.go:130:30: s.members.List undefined (type *Manager has no field or method List)
internal\member\list_test.go:48:24: e.members.List undefined (type *Manager has no field or method List)
internal\member\list_test.go:142:23: members.List undefined (type *Manager has no field or method List)
internal\member\list_test.go:178:22: members.List undefined (type *Manager has no field or method List)
internal\member\list_test.go:201:22: members.List undefined (type *Manager has no field or method List)
internal\member\list_test.go:231:27: e.members.List undefined (type *Manager has no field or method List)
--- FAIL: TestListarIntegrantes_Devuelve200ConLosIntegrantesOrdenados (0.00s)
    members_list_test.go:36: GET /api/proyectos/1/integrantes: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    members_list_test.go:38: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestListarIntegrantes_SinIntegrantesDevuelveListaVacia (0.00s)
    members_list_test.go:46: GET /api/proyectos/1/integrantes: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    members_list_test.go:48: código = 404, se esperaba 200; cuerpo: 404 page not found
--- FAIL: TestListarIntegrantes_ProyectoInexistenteDevuelve404 (0.00s)
    members_list_test.go:55: GET /api/proyectos/99/integrantes: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    members_list_test.go:57: cuerpo no es JSON válido (invalid character 'p' after top-level value): 404 page not found
--- FAIL: TestListarIntegrantes_ErrorDelRepositorioDevuelve500SinDetalle (0.00s)
    members_list_test.go:73: GET /api/proyectos/1/integrantes: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
    members_list_test.go:75: código = 404, se esperaba 500; cuerpo: 404 page not found
--- FAIL: TestRutasConID_IDNoNumericoDevuelve400 (0.01s)
        server_test.go:267: GET /api/proyectos/abc/integrantes: Content-Type = "text/plain; charset=utf-8", se esperaba application/json
        server_test.go:269: código = 404, se esperaba 400; cuerpo: 404 page not found
--- FAIL: TestInterfaz_SeccionIntegrantesTieneElBotonVerIntegrantes (0.00s)
    web_test.go:1525: la sección de integrantes no contiene "<button type=\"button\" id=\"ver-integrantes\" class=\"secundario\">Ver integrantes del proyecto</button>"
    web_test.go:1525: la sección de integrantes no contiene "<tbody id=\"lista-integrantes\"></tbody>"
    web_test.go:1525: la sección de integrantes no contiene "<th>ID</th><th>Nombre</th><th>Rol</th>"
    web_test.go:1532: el botón Ver integrantes tiene que estar entre Modificar integrante y el mensaje
--- FAIL: TestInterfaz_VerIntegrantesLlamaALaAPIConElProyectoActual (0.00s)
    web_test.go:1541: app.js no tiene el handler de "ver-integrantes"
FAIL
FAIL	seguimiento-medicion-gpc/internal/httpapi	2.252s
FAIL	seguimiento-medicion-gpc/internal/member [build failed]
ok  	seguimiento-medicion-gpc/internal/metrics	(cached)
ok  	seguimiento-medicion-gpc/internal/poker	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/report	(cached)
ok  	seguimiento-medicion-gpc/internal/sprint	(cached)
?   	seguimiento-medicion-gpc/internal/web	[no test files]
FAIL
```

Todos los fallos corresponden a la funcionalidad nueva:

- **`internal/member`:** no compila porque `Manager.List` no existe. Las referencias están en los dos archivos nuevos de HU-17.
- **`internal/httpapi`:** fallan los 6 tests nuevos (4 de API y 2 de interfaz) y la fila nueva de `TestRutasConID_IDNoNumericoDevuelve400`. La ruta `GET /api/proyectos/{id}/integrantes` no existe y responde "404 page not found". La interfaz no tiene el botón `ver-integrantes`, la tabla `lista-integrantes` ni su handler en `app.js`.
- **Los otros 12 paquetes con tests:** pasan. Ningún test anterior falló.

## Fase GREEN (`c112570`)

Código agregado:

- **`internal/member/list.go`:** `List(projectID)` valida el proyecto con `project.Manager.Get` antes de tomar el mutex y después lee los integrantes con `Repository.ListByProject` bajo el mismo mutex de `Register` y `Update`. Ante cualquier error devuelve `nil`.
- **`internal/httpapi/members_list.go`:** el handler `listMembers`, que arma la respuesta con `memberResponse`. Si no hay integrantes devuelve `[]`, porque la lista se crea con `make`.
- **`internal/httpapi/server.go`:** la ruta `GET /api/proyectos/{id}/integrantes`, junto a la de HU-15.
- **`internal/web/index.html`:** el botón "Ver integrantes del proyecto" y la tabla `lista-integrantes` (ID, Nombre y Rol), entre el formulario "Modificar integrante" y `msg-integrantes`, con los mismos contenedores (`barra` y `tabla`) que la tabla de sprints de HU-14.
- **`internal/web/app.js`:** el handler de `ver-integrantes` llena la tabla con `celda` (`textContent`), muestra el rol con `nombresDeRoles` e informa "N integrante(s) en el proyecto X.".

No cambian `register.go`, `update.go`, `repository.go`, los repositorios, `members.go`, `members_update.go`, `NewHandler` ni `cmd/server`. El GREEN no modifica tests ni el `.feature`:

```
$ git diff --stat 3a075b2 c112570 -- '*_test.go' '*.feature'
(sin salida)
```

`go test ./...` sobre `c112570`:

```
ok  	seguimiento-medicion-gpc/cmd/server	0.930s
ok  	seguimiento-medicion-gpc/internal/backlog	(cached)
ok  	seguimiento-medicion-gpc/internal/dashboard	(cached)
ok  	seguimiento-medicion-gpc/internal/defect	(cached)
ok  	seguimiento-medicion-gpc/internal/effort	(cached)
ok  	seguimiento-medicion-gpc/internal/httpapi	5.301s
ok  	seguimiento-medicion-gpc/internal/member	0.436s
ok  	seguimiento-medicion-gpc/internal/metrics	(cached)
ok  	seguimiento-medicion-gpc/internal/poker	(cached)
ok  	seguimiento-medicion-gpc/internal/postgres	(cached)
ok  	seguimiento-medicion-gpc/internal/postgrestest	(cached)
ok  	seguimiento-medicion-gpc/internal/project	(cached)
ok  	seguimiento-medicion-gpc/internal/report	(cached)
ok  	seguimiento-medicion-gpc/internal/sprint	(cached)
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

Los escenarios de HU-17 (`go test -run 'TestFeatures_HU17' -v ./internal/member/`):

```
10 scenarios (10 passed)
63 steps (63 passed)
ok  	seguimiento-medicion-gpc/internal/member	0.545s
```

Antes del commit también se corrieron `go vet ./...` y `node --check internal/web/app.js`, y ninguno informó problemas. `go tool cover -func` mostró `List` y `listMembers` en 100.0%.

### Prueba de punta a punta de la API

Se hizo con el servidor compilado de la rama (`PORT=18917`, datos en memoria) y `curl` desde Git Bash. Se usó "Bruno Diaz" sin acento, para evitar la codificación de los argumentos de `curl` en Windows (ver la evidencia de HU-14).

| Pedido | Resultado |
|---|---|
| `GET /api/proyectos/1/integrantes`, con el proyecto creado y sin integrantes | 200, `[]` |
| Ana (AGILE_ENABLER) en el proyecto 1, Ana (PRODUCT_BUILDER) en el proyecto 2 y Bruno (PRODUCT_BUILDER) en el proyecto 1; después `PUT /api/integrantes/3` con "Bruno Diaz" y PRODUCT_ARCHITECT | 201, 201, 201 y 200 |
| `GET /api/proyectos/1/integrantes` | 200, `[{"id":1,"proyecto_id":1,"nombre":"Ana","rol":"AGILE_ENABLER"},{"id":3,"proyecto_id":1,"nombre":"Bruno Diaz","rol":"PRODUCT_ARCHITECT"}]` |
| `GET /api/proyectos/2/integrantes` | 200, `[{"id":2,"proyecto_id":2,"nombre":"Ana","rol":"PRODUCT_BUILDER"}]` |
| `GET /api/proyectos/99/integrantes` | 404, `{"error":"no existe un proyecto con ese id"}` |
| `GET /api/proyectos/abc/integrantes` | 400, `{"error":"el id debe ser un número entero"}` |

## Refactor

No hubo. `List` y `listMembers` siguen el patrón de `History` y `listProjectSprints` (HU-14) y del registro (HU-15), y no hay duplicación que extraer.

## Desviaciones del ciclo TDD

No hubo. Los tests fallaron solo por la funcionalidad nueva y el código mínimo los pasó a verde sin corregir ningún test.

## Validación final

Medida sobre `c112570`, en Windows con go1.27.1, sin `TEST_DATABASE_URL`: se omitieron 41 tests de PostgreSQL, y HU-17 no agrega ninguno. `-race` no se corrió localmente porque no hay cgo; lo corre el CI, y lo necesita `TestList_ConsultasSimultaneasConRegistros`.

`gofmt -w` sobre los archivos Go modificados (`list.go`, `list_test.go`, `list_features_test.go`, `members_list.go`, `members_list_test.go`, `server.go`, `server_test.go` y `web_test.go`) no cambió nada: `git status --short` siguió vacío. `go vet ./...` no informó problemas.

`go test -count=1 ./...`:

```
ok  	seguimiento-medicion-gpc/cmd/server	1.103s
ok  	seguimiento-medicion-gpc/internal/backlog	4.569s
ok  	seguimiento-medicion-gpc/internal/dashboard	5.538s
ok  	seguimiento-medicion-gpc/internal/defect	1.934s
ok  	seguimiento-medicion-gpc/internal/effort	1.127s
ok  	seguimiento-medicion-gpc/internal/httpapi	4.937s
ok  	seguimiento-medicion-gpc/internal/member	0.797s
ok  	seguimiento-medicion-gpc/internal/metrics	4.506s
ok  	seguimiento-medicion-gpc/internal/poker	1.217s
ok  	seguimiento-medicion-gpc/internal/postgres	5.487s
ok  	seguimiento-medicion-gpc/internal/postgrestest	6.445s
ok  	seguimiento-medicion-gpc/internal/project	3.298s
ok  	seguimiento-medicion-gpc/internal/report	4.483s
ok  	seguimiento-medicion-gpc/internal/sprint	1.867s
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

`go test -count=1 -cover ./...`:

```
ok  	seguimiento-medicion-gpc/cmd/server	1.467s	coverage: 29.8% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	2.093s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/dashboard	2.681s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/defect	0.992s	coverage: 77.7% of statements
ok  	seguimiento-medicion-gpc/internal/effort	2.116s	coverage: 74.2% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	6.601s	coverage: 98.7% of statements
ok  	seguimiento-medicion-gpc/internal/member	0.457s	coverage: 68.8% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	6.104s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	0.852s	coverage: 75.5% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	4.777s	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	5.987s	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	2.234s	coverage: 74.8% of statements
ok  	seguimiento-medicion-gpc/internal/report	4.471s	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	1.357s	coverage: 77.1% of statements
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

`go test -count=1 -coverprofile=cover.out ./...` y `go tool cover -func=cover.out`, funciones nuevas y total:

```
seguimiento-medicion-gpc/internal/httpapi/members_list.go:9:			listMembers		100.0%
seguimiento-medicion-gpc/internal/member/list.go:7:				List			100.0%
total:										(statements)		84.5%
```

**Comparación con `main` (`3edc9e9`):** se midió con el mismo Go y los mismos comandos, en una copia temporal con `git worktree`. `internal/member` sube de 65.5% a 68.8% y el total, de 84.3% a 84.5%. Los demás paquetes no cambian, e `internal/httpapi` queda en 98.7% en los dos. HU-17 no cambia los repositorios ni el esquema, así que no hace falta medir con PostgreSQL: `ListByProject` ya lo cubren los tests de contrato de HU-15. El detalle está en [`docs/reportes/cobertura.md`](../reportes/cobertura.md).

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| SDD | a85051b | `docs(HU-17): agregar la especificación (SDD)` |
| RED | 3a075b2 | `test(HU-17): agregar tests (RED)` |
| GREEN | c112570 | `feat(HU-17): implementar la consulta de integrantes del proyecto (GREEN)` |

HU-17: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
