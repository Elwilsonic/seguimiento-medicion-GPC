# Evidencia TDD — TEC-05 · Persistencia con PostgreSQL en Docker

> Issue del repositorio: #35

- **Responsable:** Cardozo Leandro
- **Rama:** `feature/TEC-05-persistencia-postgres` (creada desde `main` en `a628515`, con el `Update` del backlog de HU-07 ya integrado)
- **Fecha:** 17/10/2026
- **Go:** go1.27.1 windows/amd64
- **Base:** PostgreSQL 17.11 (`postgres:17-alpine`) en Docker, publicada en `127.0.0.1:5433`
- **Driver:** `github.com/jackc/pgx/v5` v5.11.0. El equipo aprobó pgx/v5 como única dependencia nueva antes de agregarla al go.mod. Motivo de la elección en el [README](../../README.md#driver-de-postgresql)

TEC-05 es una tarea técnica: no tiene historia de usuario, especificación SDD ni escenarios BDD propios (decisión 5 del [acta de Planning del Sprint 1](../actas/sprint-1/planning.md)). Su alcance son los criterios de aceptación del Issue #35 y la decisión 4 del [acta de Planning del Sprint 2](../actas/sprint-2/planning.md): implementar en PostgreSQL los repositorios que estén en `main` (proyectos, backlog con su `Update` y sprints) sin cambiar las interfaces `Repository` ni las reglas de negocio de los `Manager`.

## Trazabilidad

Issue #35 → Criterios de Aceptación → Tests → Código Go

**35 funciones de test nuevas**, con 58 subtests en total:

| Archivo | Tests | Subtests |
|---|---|---|
| `cmd/server/main_test.go` | 10 | 5 (3 + 2) |
| `internal/postgres/postgres_test.go` | 2 | — |
| `internal/postgrestest/postgrestest_test.go` | 12 | 3 |
| `internal/project/repository_contract_test.go` y `postgres_repository_test.go` | 1 + 2 | 8 en memoria + 8 en PostgreSQL |
| `internal/backlog/repository_contract_test.go` y `postgres_repository_test.go` | 1 + 3 | 12 en memoria + 12 en PostgreSQL |
| `internal/sprint/repository_contract_test.go` y `postgres_repository_test.go` | 1 + 3 | 5 en memoria + 5 en PostgreSQL |

El conjunto de casos de cada repositorio (`runRepositoryContract`) es el mismo para memoria y para PostgreSQL. La correspondencia con los criterios no es 1 a 1: varios tests cubren más de un criterio.

| Criterio | Tests o verificación | Código |
|---|---|---|
| CA-TEC05.1 Un solo comando levanta la base | Verificación manual de `docker compose up -d --wait` (abajo) | `docker-compose.yml` |
| CA-TEC05.2 El esquema se crea automáticamente | Logs del contenedor y `\dt` desde Windows en `gpc` y `gpc_test` (abajo) | `db/init/01-schema.sql`, `db/init/02-test-database.sql` |
| CA-TEC05.3 Implementación PostgreSQL de cada repositorio, con la misma interfaz | `TestMemoryRepository_Contrato` y `TestPostgresRepository_Contrato` en `project`, `backlog` y `sprint`; `TestPostgresRepository_SaveConProyectoInexistenteFalla` (backlog y sprint), `TestPostgresRepository_FechasEnUTCConMicrosegundos`, `TestPostgresRepository_ErrorDeLaBaseNoEsNotFound` (project y backlog) | `internal/{project,backlog,sprint}/postgres_repository.go` |
| CA-TEC05.4 PostgreSQL o memoria según la configuración | 6 `TestStorageConfig_*`, `TestNewRepositories_MemoriaNoNecesitaBase`, `TestNewRepositories_PostgresBaseNoDisponibleDaError`, `TestNewRepositories_PostgresURLInvalidaNoMuestraLaContrasena`, `TestOpen_*`; prueba manual | `cmd/server/storage.go`, `cmd/server/repositories.go`, `cmd/server/main.go`, `internal/postgres/postgres.go` |
| CA-TEC05.5 Los datos se conservan al reiniciar | `TestNewRepositories_PostgresConservaLosDatosAlReabrir`; prueba manual de reinicio del servidor y del contenedor | Volumen `gpc-data` en `docker-compose.yml` |
| CA-TEC05.6 Las pruebas unitarias existentes pasan sin la base | `go test ./...` sin `TEST_DATABASE_URL`: 7 paquetes `ok` y 9 `SKIP`. El único test existente que cambió es `cmd/server/main_test.go`, que solo agrega imports y tests (`212 insertions(+), 1 deletion(-)`: la línea borrada es `import "testing"`) | — |
| CA-TEC05.7 Pruebas de integración que se omiten sin base | 6 `TestCheckDatabaseName_*` y 6 `TestDecide_*` (ninguno necesita base) | `internal/postgrestest/postgrestest.go` |
| CA-TEC05.8 Credenciales fuera del repositorio | `.env.example`; `git check-ignore` (abajo); los tests de que ningún error muestra la contraseña (`TestStorageConfig_ErroresNoMuestranLaContrasena`, `TestOpen_*`, `TestNewRepositories_Postgres*Error*`, `TestCheckDatabaseName_ErroresNoMuestranLaContrasena`, `TestDecide_MotivosNoMuestranLaContrasena`) | `.env.example`, `.gitignore` (ya ignoraba `.env` desde el Sprint 0) |
| CA-TEC05.9 El README explica cómo levantar la base y ejecutar la aplicación | — (documentación) | `README.md`, sección "Persistencia con PostgreSQL" |
| CA-TEC05.10 El driver queda documentado con el motivo | — (documentación) | `README.md`, sección "Driver de PostgreSQL" |

**En el CI los tests de PostgreSQL hacen `t.Skip`**, porque el CI de TEC-02 no tiene Docker ni `TEST_DATABASE_URL`. La cobertura de esa parte se midió **localmente, con la base levantada** (ver "Cobertura").

### Decisiones de diseño

- **Las interfaces `Repository` no cambiaron.** No reciben un `context.Context`, así que cada consulta usa `context.Background()`.
- **Un solo pool para los tres repositorios.** Lo abre `postgres.Open`, que verifica con un `Ping` que la base responda. `newRepositories` devuelve la función que lo cierra.
- **IDs `GENERATED ALWAYS AS IDENTITY`:** correlativos desde 1, como en memoria (RN-01.5, RN-04.6 y RN-10.6).
- **Criterios de aceptación en `TEXT[] NOT NULL`:** conservan el orden. Una lista `nil` se guarda y se lee como lista vacía (en memoria, `nil` se lee como `nil`; el contrato pide `len == 0`).
- **Claves foráneas** de `backlog_items` y `sprints` a `projects`. En memoria no hay clave foránea: la existencia del proyecto la valida el `Manager` (HU-04 y HU-10). Por eso `TestPostgresRepository_SaveConProyectoInexistenteFalla` (código `23503`) corre **solo contra PostgreSQL**.
- **Fechas de los sprints en `TIMESTAMPTZ`:** al leer (también en el `RETURNING` de `Save`) se normalizan a UTC, porque pgx devuelve `timestamptz` en `time.Local`. PostgreSQL guarda microsegundos, así que los nanosegundos se pierden.
- **`Save` y `Update` devuelven lo que quedó guardado** (`RETURNING`), no una copia de lo recibido.
- **Los errores nunca incluyen la URL**, que tiene la contraseña: el error de `ParseConfig` no se envuelve, porque puede repetir la URL.
- **El servidor no tiene endpoints de sprints:** `newRepositories` arma el repositorio de sprints, pero `main.go` todavía no lo usa (`httpapi` no tiene rutas de sprints).

## Infraestructura (antes del RED)

Commit `88e0b04`, antes de los tests, así el RED falla solo por el código que todavía no existe.

**Puerto 5432 ocupado.** En la máquina del responsable hay un PostgreSQL 17 instalado como servicio de Windows (`postgresql-x64-17`) que escucha en `0.0.0.0:5432`. Al publicar el puerto solo en loopback, Docker no pudo usar el 5432:

```
Error response from daemon: ports are not available: exposing port TCP 127.0.0.1:5432 -> 127.0.0.1:0: listen tcp4 127.0.0.1:5432: bind: An attempt was made to access a socket in a way forbidden by its access permissions.
```

Antes de publicar en loopback, Docker aceptaba `0.0.0.0:5432`, pero desde Windows `localhost:5432` llegaba al PostgreSQL instalado y no al de Docker. Se pasó al puerto `127.0.0.1:5433`, y la conexión se verificó desde Windows y no desde adentro del contenedor:

```
$ docker compose up -d --wait
 Container seguimiento-medicion-gpc-postgres-1  Healthy
exit=0
$ docker compose port postgres 5432
127.0.0.1:5433
$ Test-NetConnection 127.0.0.1 -Port 5433
TcpTestSucceeded : True

$ psql -h 127.0.0.1 -p 5433 -U gpc -d gpc -c "select version(), current_setting('data_directory')"
version        | PostgreSQL 17.11 on x86_64-pc-linux-musl, compiled by gcc (Alpine 15.2.0) 15.2.0, 64-bit
data_directory | /var/lib/postgresql/data
```

`x86_64-pc-linux-musl` y `/var/lib/postgresql/data` confirman que responde el PostgreSQL de Docker.

**Esquema creado automáticamente (CA-TEC05.2):**

```
postgres-1  | /usr/local/bin/docker-entrypoint.sh: running /docker-entrypoint-initdb.d/01-schema.sql
postgres-1  | /usr/local/bin/docker-entrypoint.sh: running /docker-entrypoint-initdb.d/02-test-database.sql

$ psql -h 127.0.0.1 -p 5433 -U gpc -d gpc -c '\dt'      (igual en -d gpc_test)
 Esquema |    Nombre     | Tipo  | Due
---------+---------------+-------+-----
 public  | backlog_items | tabla | gpc
 public  | projects      | tabla | gpc
 public  | sprints       | tabla | gpc
(3 filas)
```

**Credenciales (CA-TEC05.8):** solo existe `.env.example`; un `.env` real quedaría ignorado.

```
$ git check-ignore -v --no-index .env .env.local .env.example
.gitignore:17:.env	.env
.gitignore:18:.env.*	.env.local
.gitignore:19:!.env.example	.env.example
```

## Tanda 1 — Configuración del almacenamiento

### Fase RED (tanda 1)

Tests escritos antes del código (commit `e9fda78`). **El RED fue un error de compilación**: `storageConfig` y sus constantes y errores no existían.

```
$ go test -gcflags=-e ./cmd/server
      2 undefined: errMissingDatabaseURL
      2 undefined: errUnknownStorage
      6 undefined: storageConfig
      4 undefined: storageMemory
      2 undefined: storagePostgres
FAIL	seguimiento-medicion-gpc/cmd/server [build failed]
```

Los 16 errores son identificadores nuevos de TEC-05. `internal/backlog`, `internal/httpapi`, `internal/project` e `internal/sprint` seguían en `ok`.

### Fase GREEN (tanda 1)

Commit `17d9d7f`: `cmd/server/storage.go` (`storageConfig`). Los errores se comparan con `errors.Is` y nunca incluyen `DATABASE_URL`.

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.113s
ok  	seguimiento-medicion-gpc/internal/backlog	0.335s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.125s
ok  	seguimiento-medicion-gpc/internal/project	2.881s
ok  	seguimiento-medicion-gpc/internal/sprint	2.378s
exit=0

$ go tool cover -func=cover-server.out
seguimiento-medicion-gpc/cmd/server/main.go:15:		main		0.0%
seguimiento-medicion-gpc/cmd/server/main.go:26:		address		100.0%
seguimiento-medicion-gpc/cmd/server/storage.go:32:	storageConfig	100.0%
```

### Pruebas de mutación de `storageConfig`

Cada mutación se aplicó sola, se corrió `go test -count=1 ./cmd/server` y se restauró el archivo (mismo blob `86d588c` antes y después).

```
== a) strings.TrimSpace sobre STORAGE
--- FAIL: TestStorageConfig_ValorDesconocidoDaError/_memory
    main_test.go:103: err = <nil>, se esperaba STORAGE debe ser memory o postgres

== b) postgres como valor por defecto
--- FAIL: TestStorageConfig_SinSTORAGEUsaMemoria
    main_test.go:50: storageConfig devolvió un error inesperado: STORAGE=postgres necesita la variable DATABASE_URL

== c) DATABASE_URL en el mensaje de error
--- FAIL: TestStorageConfig_ErroresNoMuestranLaContrasena/valor_desconocido
    main_test.go:123: el error muestra la contraseña: STORAGE debe ser memory o postgres (se recibió "mysql", DATABASE_URL="postgres://gpc:s3cr3t-no-mostrar@localhost:5433/gpc?sslmode=disable")
--- FAIL: TestStorageConfig_ErroresNoMuestranLaContrasena/mayúsculas
```

- La mutación (a) la detecta **un solo caso**: `" memory"`, con un espacio adelante. Por eso ese caso tiene un comentario en el test que pide no quitarlo sin agregar otro igual (se sumó en el commit RED de la tanda 2).
- La mutación (c) **solo aplica a la rama de valor desconocido**: `errMissingDatabaseURL` aparece solo cuando la URL está vacía, así que no hay nada que se pueda filtrar.

## Tanda 2 — Repositorios PostgreSQL

### Dependencia

pgx entró en el commit RED, junto con `go.sum`, porque los tests importan `pgxpool` y `pgconn`.

```
$ go get github.com/jackc/pgx/v5@v5.11.0
go: added github.com/jackc/pgpassfile v1.0.0
go: added github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
go: added github.com/jackc/pgx/v5 v5.11.0
go: added golang.org/x/text v0.29.0

$ go list -m -json github.com/jackc/pgx/v5
	"Path": "github.com/jackc/pgx/v5",
	"Version": "v5.11.0",
	"Time": "2026-09-07T23:39:32Z",
	"GoVersion": "1.25.0",
```

Después de `go mod tidy`, pgx quedó como dependencia directa. Las indirectas nuevas son `jackc/pgpassfile v1.0.0`, `jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761`, `jackc/puddle/v2 v2.2.2`, `golang.org/x/sync v0.17.0` y `golang.org/x/text v0.29.0`.

- La línea `go 1.27.1` no cambió y no apareció ninguna línea `toolchain`.
- `go.sum` suma 18 líneas. Algunas son hashes `/go.mod` de `testify`, `objx`, `go-spew`, `check.v1` y `yaml.v3`, que Go usa solo para verificar el grafo de módulos: no se compilan.

### Fase RED (tanda 2)

Tests escritos antes del código (commit `1062b45`). **El RED fue un error de compilación** y se cortó al cargar los paquetes, porque `internal/postgres` e `internal/postgrestest` solo tenían archivos `_test.go`:

```
$ go test -gcflags=-e ./...
seguimiento-medicion-gpc/internal/postgrestest: no non-test Go files in …\internal\postgrestest
seguimiento-medicion-gpc/internal/postgres: no non-test Go files in …\internal\postgres
FAIL	seguimiento-medicion-gpc/cmd/server [build failed]
FAIL	seguimiento-medicion-gpc/internal/backlog [build failed]
# seguimiento-medicion-gpc/internal/postgres [seguimiento-medicion-gpc/internal/postgres.test]
internal\postgres\postgres_test.go:19:15: undefined: Open
internal\postgres\postgres_test.go:33:15: undefined: Open
ok  	seguimiento-medicion-gpc/internal/httpapi	(cached)
FAIL	seguimiento-medicion-gpc/internal/postgres [build failed]
FAIL	seguimiento-medicion-gpc/internal/postgrestest [build failed]
FAIL	seguimiento-medicion-gpc/internal/project [build failed]
FAIL	seguimiento-medicion-gpc/internal/sprint [build failed]
FAIL
exit=1
```

Ese corte oculta los errores de los demás paquetes. Para verlos se corrió el mismo comando (con `-vet=off`) en una **copia del repositorio fuera de la rama**, donde a los dos paquetes se les agregó un archivo que tiene solo la línea `package`, sin ninguna declaración. Salieron **50 errores, todos `undefined` de identificadores nuevos de TEC-05**:

```
      7 undefined: postgrestest.Pool
      6 undefined: checkDatabaseName
      6 undefined: decide
      6 undefined: errNotTestDatabase
      5 undefined: newRepositories
      3 undefined: NewPostgresRepository
      3 undefined: postgrestest.Reset
      2 undefined: Open
      2 undefined: actionReject
      2 undefined: actionSkip
      2 undefined: actionUse
      2 undefined: errInvalidURL
      2 undefined: project.NewPostgresRepository
      1 undefined: postgres.Open
      1 undefined: postgrestest.EnvVar
```

- Ese diagnóstico encontró **un error en un test**, que se corrigió antes del commit RED: `internal\backlog\repository_contract_test.go:226:5: invalid operation: got != want (struct containing []string cannot be compared)`. `contractAssertItem` comparaba dos `Item` con `!=`. Ahora los compara con `reflect.DeepEqual`, con los criterios en `nil` de los dos lados (los criterios se comparan aparte con `slices.Equal`).
- En la copia, `httpapi` figuró como `FAIL` por `fork/exec …httpapi.test.exe: The directory name is invalid`, debido a la ruta larga de la carpeta temporal. En el repositorio, `httpapi` da `ok`.
- **Los tests de contrato de memoria son de caracterización:** pasan desde la primera corrida, sin cambios en `MemoryRepository`. En el RED no se ven porque el paquete no compila.

### Fase GREEN (tanda 2)

Commit `4f55a19`:
- `internal/postgres/postgres.go` (`Open`)
- `internal/postgrestest/postgrestest.go` (`Pool`, `Reset`, `decide`, `checkDatabaseName` y el lock)
- `PostgresRepository` en `project`, `backlog` y `sprint`
- `cmd/server/repositories.go` (`newRepositories`)
- `cmd/server/main.go`, que ahora usa `storageConfig` y `newRepositories`

**Paquetes en paralelo.** Con la base configurada, la primera versión del GREEN pasaba con `go test -p 1 ./...`, pero fallaba con `go test ./...`. `go test` corre los paquetes en paralelo, y todos comparten `gpc_test`: un paquete vaciaba las tablas mientras otro guardaba.

```
--- FAIL: TestPostgresRepository_Contrato (0.42s)
        repository_contract_test.go:84: Save devolvió un error inesperado: guardar el ítem: ERROR: insert or update on table "backlog_items" violates foreign key constraint "backlog_items_project_id_fkey" (SQLSTATE 23503)
--- FAIL: TestPostgresRepository_Contrato (0.26s)
        repository_contract_test.go:25: segundo Save = {ID:1 Name:Beta Description: Status:EN_CURSO}, se esperaba {ID:2 Name:Beta Description: Status:EN_CURSO}
```

Se agregó a `postgrestest.Pool` un lock de PostgreSQL (`pg_advisory_lock`), tomado en una conexión propia y soltado al terminar el test. Con eso, los tests que usan la base corren de a uno. Después del cambio pasaron 5 de 5 corridas de `go test -count=1 ./...` con los paquetes en paralelo.

**Sin `TEST_DATABASE_URL` (como en el CI):**

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	1.201s
ok  	seguimiento-medicion-gpc/internal/backlog	0.435s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.177s
ok  	seguimiento-medicion-gpc/internal/postgres	0.201s
ok  	seguimiento-medicion-gpc/internal/postgrestest	1.213s
ok  	seguimiento-medicion-gpc/internal/project	0.395s
ok  	seguimiento-medicion-gpc/internal/sprint	0.359s

$ go test -count=1 -v ./... | grep -- '--- SKIP' | sort | uniq -c
      1 --- SKIP: TestNewRepositories_PostgresConservaLosDatosAlReabrir (0.00s)
      3 --- SKIP: TestPostgresRepository_Contrato (0.00s)
      2 --- SKIP: TestPostgresRepository_ErrorDeLaBaseNoEsNotFound (0.00s)
      1 --- SKIP: TestPostgresRepository_FechasEnUTCConMicrosegundos (0.00s)
      2 --- SKIP: TestPostgresRepository_SaveConProyectoInexistenteFalla (0.00s)
```

Los 9 SKIP tienen el mismo motivo: `TEST_DATABASE_URL no está definida: se omiten las pruebas de PostgreSQL`.

**Con `TEST_DATABASE_URL` apuntando a `gpc_test` en el puerto 5433:**

```
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.457s
ok  	seguimiento-medicion-gpc/internal/backlog	1.942s
ok  	seguimiento-medicion-gpc/internal/httpapi	0.187s
ok  	seguimiento-medicion-gpc/internal/postgres	0.214s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.215s
ok  	seguimiento-medicion-gpc/internal/project	1.637s
ok  	seguimiento-medicion-gpc/internal/sprint	1.800s
exit=0
SKIP/FAIL: 0
```

### Pruebas de mutación de los repositorios PostgreSQL

Se corrieron con la base levantada. Cada mutación se aplicó sola y se restauró el archivo desde una copia, verificando el hash (`b4a7eeb` project, `4dc8fe8` backlog, `527f60a` sprint). **No se mutó la protección del sufijo `_test`**, porque una mutación ahí, corriendo contra la base real, puede borrar datos. Esa protección se prueba solo con los tests que no necesitan base.

**a) Sin `ORDER BY` en `List` y `ListByProject`.** Con los tests del RED, esta mutación **no falló**: sin un `UPDATE`, PostgreSQL devuelve las filas en el orden en que se insertaron.

```
$ go test -count=1 ./internal/project ./internal/backlog
ok  	seguimiento-medicion-gpc/internal/project	26.258s
ok  	seguimiento-medicion-gpc/internal/backlog	27.506s
```

Se agregaron los casos "sigue en orden de ID después de modificar el primero" (commit `da21dd8`): hacen `Update` de la fila 1 y verifican el orden 1, 2, 3. En `project` el caso falló enseguida:

```
--- FAIL: TestPostgresRepository_Contrato/List_sigue_en_orden_de_ID_después_de_modificar_el_primero
    repository_contract_test.go:75: List = [{ID:2 …} {ID:3 …} {ID:1 Name:Alfa Description: Status:EN_CURSO}], se esperaba [{ID:1 …} {ID:2 …} {ID:3 …}]
```

En `backlog`, cambiar solo el título no alcanzaba. La explicación probable (no verificada con `EXPLAIN`) es que PostgreSQL hace un "HOT update" que conserva la posición de la fila en el índice de `project_id`. El caso mueve el ítem al proyecto B y de vuelta al A, y así falló en 3 de 3 corridas:

```
--- FAIL: TestPostgresRepository_Contrato/ListByProject_sigue_en_orden_de_ID_después_de_modificar_el_primero
    repository_contract_test.go:157: ListByProject[0] = {ID:2 ProjectID:1 Title:Logout …}, se esperaba {ID:1 ProjectID:1 Title:Login modificado …}
```

**b) `ListByProject` sin filtrar por `project_id`:**

```
--- FAIL: …/ListByProject_devuelve_solo_los_ítems_del_proyecto_en_orden_de_ID
    repository_contract_test.go:126: ListByProject devolvió 3 ítems, se esperaban 2
--- FAIL: …/ListByProject_de_un_proyecto_sin_ítems_devuelve_una_lista_vacía,_no_nil
```

**c) `Update` ignora `estimated_hours`:**

```
--- FAIL: …/Update_reemplaza_todos_los_campos_del_ítem
    repository_contract_test.go:193: Update = {… EstimatedHours:0}, se esperaba {… EstimatedHours:12.25}
    repository_contract_test.go:198: GetByID después de Update = {… EstimatedHours:0}, se esperaba {… EstimatedHours:12.25}
```

**d) Sin `.UTC()` al leer las fechas.**
- La primera versión del test dependía de la zona horaria de la máquina: fallaba solo porque la de esta máquina es -03.
- El test ahora fija `time.Local` en `UTC-3` y lo restaura en `t.Cleanup` (commit `a3c3637`).
- La mutación se repitió y falló con la zona del test. El nombre `UTC-3` de la salida es el de la `FixedZone` del test; antes salía `-03`, que era la zona del sistema.

```
--- FAIL: TestPostgresRepository_FechasEnUTCConMicrosegundos
    postgres_repository_test.go:71: Save: fechas = 2026-10-16 09:30:00.123456 -0300 UTC-3 a 2026-10-22 18:00:00.987654 -0300 UTC-3, se esperaba 2026-10-16 12:30:00.123456 +0000 UTC a 2026-10-22 21:00:00.987654 +0000 UTC
    postgres_repository_test.go:71: OpenByProject: fechas = … -0300 UTC-3 …, se esperaba … +0000 UTC
```

**e) `OpenByProject` ignora el estado:**

```
--- FAIL: …/OpenByProject_devuelve_el_sprint_ABIERTO_del_proyecto
    repository_contract_test.go:65: OpenByProject = {ID:1 … Status:CERRADO}, se esperaba {ID:2 … Status:ABIERTO}
--- FAIL: …/OpenByProject_no_ve_un_sprint_CERRADO
    repository_contract_test.go:90: OpenByProject encontró un sprint CERRADO
```

**f) Cualquier error de `GetByID` y `Update` se informa como inexistente** (en `project` y en `backlog`). Esta mutación la detectan los tests de errores de la base (commit `6a55051`), que usan un pool propio cerrado a propósito:

```
--- FAIL: TestPostgresRepository_ErrorDeLaBaseNoEsNotFound
    postgres_repository_test.go:54: GetByID: err = no existe un proyecto con ese id, se esperaba un error de la base distinto de no existe un proyecto con ese id
    postgres_repository_test.go:54: Update: err = no existe un proyecto con ese id, …
--- FAIL: TestPostgresRepository_ErrorDeLaBaseNoEsNotFound
    postgres_repository_test.go:79: GetByID: err = no existe un ítem del backlog con ese id, …
    postgres_repository_test.go:79: Update: err = no existe un ítem del backlog con ese id, …
```

### Desviaciones del ciclo TDD

- **El lock de `postgrestest.Pool` se escribió en el GREEN.** No hay un test propio que lo pida: lo pidieron las fallas de los tests de contrato con los paquetes en paralelo, que se ven arriba.
- **Los tests de los commits `da21dd8`, `6a55051` y `a3c3637` se escribieron después del GREEN**, a partir de las pruebas de mutación. Pasan desde la primera corrida, porque el código ya existía. Lo que muestra que protegen el código es que la mutación correspondiente (a, f y d) los hace fallar.
- **`Reset` no vuelve a controlar el nombre de la base.** Ese control se quitó en el GREEN porque quedaba sin test. Su comentario indica que solo se usa con el pool que devuelve `Pool(t)`, que sí controla el sufijo.

## Refactor

No aplica. Los tres repositorios repiten la misma forma (consulta, `scanX`, traducción de `pgx.ErrNoRows`), pero cada uno tiene sus columnas y sus errores. Un helper genérico ahorraría pocas líneas y haría menos directa la lectura de cada método.

## Prueba manual

**Selección del almacenamiento y reinicio del servidor (CA-TEC05.4 y CA-TEC05.5).** Se compiló el servidor con `go build ./cmd/server` y se ejecutó con `STORAGE=postgres`, `DATABASE_URL` apuntando a la base `gpc` y `PORT=8099`. En el segundo arranque se esperó a que el servidor respondiera y se verificó que el puerto quedara libre al detenerlo.

```
2026/10/17 23:43:49 servidor escuchando en :8099 (almacenamiento: postgres)
$ curl -s -X POST localhost:8099/proyectos -d '{"nombre":"Persistencia TEC-05","descripcion":"Prueba de reinicio"}'
{"id":1,"nombre":"Persistencia TEC-05","descripcion":"Prueba de reinicio","estado":"PLANIFICADO"}
  [HTTP 201]
$ curl -s -X POST localhost:8099/proyectos/1/backlog -d '{"titulo":"Guardar en PostgreSQL","descripcion":"","prioridad":"ALTA","criterios":["Sobrevive al reinicio"]}'
{"id":1,"proyecto_id":1,"titulo":"Guardar en PostgreSQL","descripcion":"","prioridad":"ALTA","criterios":["Sobrevive al reinicio"],"estado":"PENDIENTE","story_points":0}
  [HTTP 201]
--- servidor detenido

2026/10/17 23:44:28 servidor escuchando en :8099 (almacenamiento: postgres)
$ curl -s localhost:8099/proyectos/1
{"id":1,"nombre":"Persistencia TEC-05","descripcion":"Prueba de reinicio","estado":"PLANIFICADO"}
  [HTTP 200]
$ curl -s localhost:8099/proyectos/1/backlog
[{"id":1,"proyecto_id":1,"titulo":"Guardar en PostgreSQL","descripcion":"","prioridad":"ALTA","criterios":["Sobrevive al reinicio"],"estado":"PENDIENTE","story_points":0}]
  [HTTP 200]
--- puerto 8099 libre: servidor detenido
```

En un primer intento, el segundo arranque no llegó a estar listo antes del primer `curl` (`[HTTP 000]`), así que no se usó como evidencia. Se repitió esperando a que el servidor respondiera; esa es la salida de arriba.

**Sin `STORAGE` y con valores inválidos:**

```
2026/10/17 23:43:58 servidor escuchando en :8098 (almacenamiento: memory)
$ curl -s localhost:8098/proyectos/1
{"error":"no existe un proyecto con ese id"}
  [HTTP 404]

$ STORAGE=mysql server.exe
2026/10/17 23:44:00 STORAGE debe ser memory o postgres (se recibió "mysql")
$ STORAGE=postgres server.exe          (sin DATABASE_URL)
2026/10/17 23:44:00 STORAGE=postgres necesita la variable DATABASE_URL
```

**Reinicio de la base (`docker compose down` sin `-v`):**

```
$ docker compose down
 Container seguimiento-medicion-gpc-postgres-1  Removed
$ docker compose up -d --wait
 Container seguimiento-medicion-gpc-postgres-1  Healthy
$ psql -h 127.0.0.1 -p 5433 -U gpc -d gpc -c "select id, name from projects" -c "select id, project_id, title, acceptance_criteria from backlog_items"
 id |        name
----+---------------------
  1 | Persistencia TEC-05
(1 fila)

 id | project_id |         title         |    acceptance_criteria
----+------------+-----------------------+---------------------------
  1 |          1 | Guardar en PostgreSQL | {"Sobrevive al reinicio"}
(1 fila)
```

## Validación final

```
$ gofmt -l .
(sin salida: todos los archivos están formateados)

$ go vet ./...
(sin salida)

$ go test -count=1 -cover ./...          (con TEST_DATABASE_URL)
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 41.5% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 98.6% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 96.8% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	coverage: 98.3% of statements

$ go test -count=1 -cover ./...          (sin TEST_DATABASE_URL, como en el CI)
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 39.0% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 75.5% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 67.4% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	coverage: 69.0% of statements
```

`go test -race` **no se corrió localmente**: en esta máquina Windows no hay compilador C (lo mismo que en HU-07). Lo corre el CI del Pull Request, donde los tests de PostgreSQL hacen `t.Skip`.

### Cobertura

`go test -cover` mide cada paquete solo con sus propios tests. Por eso `internal/postgrestest` queda en 36.8%: su `Pool`, su lock y su `Reset` los ejecutan los tests de los otros paquetes. Para medir el código de TEC-05 completo se usó `-coverpkg=./...`, **con la base levantada**:

```
$ go test -count=1 -coverpkg=./... -coverprofile=cover.out ./...
$ go tool cover -func=cover.out
cmd/server/main.go:17:				main			0.0%
cmd/server/main.go:38:				address			100.0%
cmd/server/repositories.go:24:			newRepositories		100.0%
cmd/server/storage.go:32:			storageConfig		100.0%
internal/backlog/postgres_repository.go:21:	NewPostgresRepository	100.0%
internal/backlog/postgres_repository.go:31:	Save			100.0%
internal/backlog/postgres_repository.go:46:	GetByID			100.0%
internal/backlog/postgres_repository.go:60:	Update			100.0%
internal/backlog/postgres_repository.go:80:	ListByProject		77.8%
internal/backlog/postgres_repository.go:98:	criteriaParam		100.0%
internal/backlog/postgres_repository.go:107:	scanItem		100.0%
internal/postgres/postgres.go:23:		Open			91.7%
internal/postgrestest/postgrestest.go:59:	Pool			77.8%
internal/postgrestest/postgrestest.go:77:	lock			72.7%
internal/postgrestest/postgrestest.go:100:	Reset			75.0%
internal/postgrestest/postgrestest.go:111:	decide			100.0%
internal/postgrestest/postgrestest.go:129:	checkDatabaseName	100.0%
internal/project/postgres_repository.go:21:	NewPostgresRepository	100.0%
internal/project/postgres_repository.go:30:	Save			80.0%
internal/project/postgres_repository.go:42:	List			75.0%
internal/project/postgres_repository.go:58:	GetByID			100.0%
internal/project/postgres_repository.go:72:	Update			100.0%
internal/project/postgres_repository.go:87:	scanProject		100.0%
internal/sprint/postgres_repository.go:21:	NewPostgresRepository	100.0%
internal/sprint/postgres_repository.go:31:	Save			100.0%
internal/sprint/postgres_repository.go:45:	OpenByProject		85.7%
internal/sprint/postgres_repository.go:61:	scanSprint		100.0%
total:									(statements)		92.1%
```

Lo que queda sin cubrir son ramas de error que necesitan que la base falle en un momento puntual:
- Los errores de consulta de `List`, `ListByProject` y `OpenByProject`, y el error de `Save` de proyectos.
- El error de `NewWithConfig` en `Open`.
- Los errores de `lock` y de `Reset`.
- En esta medición, `Pool` no cubre `t.Skip` ni `t.Fatal`: `t.Skip` se ejecuta en la corrida sin base, y la decisión de saltear o rechazar se prueba en `decide`.
- `main()` sigue en 0%: arranca el servidor y bloquea en `http.ListenAndServe`, como en TEC-01. Se verificó con la prueba manual.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| Infraestructura | 88e0b04 | `chore(TEC-05): agregar postgresql en docker compose con el esquema` |
| RED (tanda 1) | e9fda78 | `test(TEC-05): agregar tests de la configuracion del almacenamiento (RED)` |
| GREEN (tanda 1) | 17d9d7f | `feat(TEC-05): implementar la seleccion del almacenamiento por variables de entorno (GREEN)` |
| RED (tanda 2) | 1062b45 | `test(TEC-05): agregar tests de los repositorios postgresql (RED)` |
| GREEN (tanda 2) | 4f55a19 | `feat(TEC-05): implementar los repositorios postgresql (GREEN)` |
| Test (mutación a) | da21dd8 | `test(TEC-05): verificar el orden de los listados despues de un update` |
| Test (mutación f) | 6a55051 | `test(TEC-05): verificar que un error de la base no se informa como inexistente` |
| Test (mutación d) | a3c3637 | `test(TEC-05): fijar la zona horaria en el test de fechas de sprints` |

No hubo commit de refactor (ver "Refactor"). El commit de esta evidencia no se lista en la tabla, porque un commit no puede contener su propio hash. No hay commit de cierre: `Closes #35` va en la descripción del Pull Request, y el Issue se cierra solo cuando el merge llega a `main`.
