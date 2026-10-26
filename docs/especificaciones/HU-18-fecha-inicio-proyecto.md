# HU-18 · Registrar la fecha de inicio del proyecto

> Issue del repositorio: #36

Aprobada por: Peñalbé Hernán (Agile Enabler), 26/10/2026, con el ajuste de registrar la ruta en el GREEN.

## Historia de usuario

Como Agile Enabler, quiero registrar la fecha de inicio de un proyecto, para saber desde cuándo se cuenta su duración.

## Objetivo

Registrar (o volver a registrar) la fecha de inicio de un proyecto existente, con una operación dedicada del `Manager` de `internal/project` que solo cambia esa fecha. El proyecto suma el campo `StartDate`, que se muestra en la consulta (HU-03) y se conserva al modificar el proyecto (HU-02). Es la primera parte de la historia original "Registrar fechas de inicio y fin del proyecto", dividida en el Planning del Sprint 3: la fecha de fin y su relación con la de inicio quedan para HU-41.

Se expone en `PATCH /api/proyectos/{id}/fecha-inicio` y en la sección Proyectos de la interfaz web (tema 5 del Planning del Sprint 3). Persiste en memoria y en PostgreSQL (TEC-05).

## Alcance

- **Incluye:** el campo `StartDate` del proyecto; la operación `SetStartDate` y su error de fecha obligatoria; la columna `start_date` en `projects` y su lectura y escritura en el repositorio PostgreSQL; el campo `fecha_inicio` en las respuestas de proyectos; el endpoint HTTP; el campo de fecha, el botón y la fecha visible en la sección Proyectos de la interfaz web; tests unitarios, BDD, de contrato, de PostgreSQL, de API y de interfaz.
- **No incluye:** la fecha de fin del proyecto ni la regla "fin posterior al inicio" (HU-41); la duración calculada del proyecto ni cualquier métrica que la use; HU-19; pedir la fecha al crear el proyecto (HU-01) o recibirla en `PUT /api/proyectos/{id}` (HU-02); borrar una fecha ya registrada; relacionar la fecha del proyecto con las fechas de sus sprints.

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | Identificador del proyecto |
| Fecha de inicio | fecha (`time.Time`) | Sí | Fecha desde la que se cuenta la duración del proyecto; el valor cero de `time.Time` es fecha ausente |

Firma: `SetStartDate(id int, date time.Time) (Project, error)`, en `internal/project/start_date.go`.

La fecha se guarda tal como llega, igual que las fechas de HU-10 ("Fechas: las recibidas, tal como llegaron"). El formato de texto (`AAAA-MM-DD`) lo resuelve la capa HTTP (ver "HTTP").

### Salidas esperadas

- **Si el registro es válido:** el proyecto actualizado, como **copia por valor**, con `StartDate` igual a la fecha recibida y ID, nombre, descripción y estado iguales a los que tenía.
- **Si no es válido:** el error que indica el motivo (ver "Condiciones de error"); el proyecto guardado queda exactamente como estaba.

`Project` suma el campo `StartDate time.Time`. Un proyecto recién creado (HU-01) tiene `StartDate` en cero, que significa "sin registrar". `Create` no recibe la fecha y su firma no cambia.

### HTTP (TEC-04, tema 5 del Planning)

`PATCH /api/proyectos/{id}/fecha-inicio`, con el mismo patrón de operación dedicada que `PATCH /api/items/{id}/estado` (HU-08) y `PATCH /api/items/{id}/story-points` (HU-20).

- **Cuerpo:**

  ```json
  {"fecha_inicio": "2026-10-23"}
  ```

- **Lectura de la fecha:** se reutilizan `dateLayout` (`AAAA-MM-DD`) y `parseDate` de `internal/httpapi/sprints.go`, sin cambios. `parseDate` devuelve la fecha a las 00:00 UTC; una cadena vacía la devuelve en cero y el `Manager` responde `ErrMissingStartDate`; un texto que no respeta el formato o una fecha que no existe en el calendario (por ejemplo `2026-02-30`) devuelve `errInvalidDate`.
- **Respuesta exitosa:** `200 OK` con el proyecto:

  ```json
  {"id": 1, "nombre": "Sistema de ventas", "descripcion": "Gestión de ventas", "estado": "EN_CURSO", "fecha_inicio": "2026-10-23"}
  ```

- **`fecha_inicio` en las demás respuestas de proyectos:** `POST /api/proyectos` (HU-01), `GET /api/proyectos/{id}` (HU-03) y `PUT /api/proyectos/{id}` (HU-02) devuelven también `fecha_inicio`, en `AAAA-MM-DD`, o `""` si no está registrada. `toProjectResponse` formatea la fecha con `dateLayout` solo si no es cero; si lo es, devuelve `""` (si se formateara el cero, saldría `0001-01-01`). El campo no se agrega a `projectRequest`: `POST` y `PUT` lo ignoran si un cliente lo envía.
- **Errores:**

| Caso | Código | Mensaje |
|---|---|---|
| `fecha_inicio` vacía, `null` o ausente | 400 | El de `ErrMissingStartDate` |
| `fecha_inicio` con otro formato (`23/10/2026`, `2026-10-23T00:00:00Z`, `2026-1-5`, con espacios) o fecha inexistente (`2026-02-30`) | 400 | El de `errInvalidDate` |
| Cuerpo que no es JSON, `fecha_inicio` que no es texto (por ejemplo un número), o texto después del JSON | 400 | El de `errInvalidJSON` |
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| Proyecto inexistente | 404 | El de `ErrProjectNotFound` |
| Cuerpo de más de 1 MiB | 413 | El de `errBodyTooLarge` |
| Falla del repositorio | 500 | El de `errInternal`, sin detalle |

El orden de los errores del handler es: ID de la ruta, JSON, formato de la fecha (`parseDate`) y `Manager`; dentro del `Manager`, fecha ausente (antes de tocar el repositorio) y existencia del proyecto. Por eso, una fecha vacía sobre un proyecto inexistente responde 400 y no 404.

### Interfaz web (TEC-04, tema 5)

En la sección **Proyectos** de `internal/web`:

- Un formulario "Fecha de inicio" para el **proyecto actual**, con un campo de fecha (`<input type="date" name="fecha_inicio" required>`, que el navegador envía como `AAAA-MM-DD`) y el botón **"Registrar fecha de inicio"**.
- Llama a `PATCH /proyectos/{id}/fecha-inicio` con el ID de «Proyecto actual», muestra el proyecto devuelto y el mensaje "Fecha de inicio del proyecto N registrada: AAAA-MM-DD." o el mensaje de error de la API, con el mismo mecanismo de acción y mensaje que los demás formularios (`accion` y `msg-proyectos`).
- Los datos del proyecto actual (`mostrarProyecto`) muestran la fecha: "… inicio: AAAA-MM-DD" o "sin fecha de inicio" si `fecha_inicio` es `""`. El campo de fecha se precarga con la fecha registrada.
- "Guardar cambios" (HU-02) no envía la fecha y la conserva (RN-18.6).

## Reglas de negocio

1. **RN-18.1 Fecha obligatoria.** La fecha de inicio es obligatoria en la operación: el valor cero de `time.Time` es fecha ausente y devuelve `ErrMissingStartDate`, sin consultar el repositorio, con el mismo criterio que `ErrMissingStartDate` de HU-10 (RN-10.3).
2. **RN-18.2 Formato `AAAA-MM-DD`.** En la API la fecha llega como texto `AAAA-MM-DD`, el mismo formato de las fechas de los sprints (HU-10); un formato distinto o una fecha que no existe en el calendario devuelve `errInvalidDate` (400). Es la validación de formato que pide el criterio CA-18.2 del Issue.
3. **RN-18.3 Proyecto existente.** Requiere un proyecto con ese ID; si no existe (incluye 0 y negativos), se devuelve `ErrProjectNotFound`, el de HU-03, sin definir un error nuevo.
4. **RN-18.4 Se puede volver a registrar.** Registrar una fecha en un proyecto que ya tiene una la reemplaza. Registrar la misma fecha es válido y deja el proyecto igual.
5. **RN-18.5 Cualquier estado y cualquier fecha.** Se registra en un proyecto `PLANIFICADO`, `EN_CURSO` o `FINALIZADO`. No se limita a fechas pasadas ni futuras: el Issue no lo pide, y un proyecto planificado puede tener un inicio futuro. La relación con la fecha de fin queda para HU-41.
6. **RN-18.6 Solo cambia la fecha de inicio, y nada la borra.** `SetStartDate` no modifica el nombre, la descripción ni el estado del proyecto, ni sus ítems o sprints. A su vez, `Update` (HU-02) conserva la fecha registrada, porque lee el proyecto guardado y solo reemplaza nombre, descripción y estado. Ninguna operación vuelve la fecha a cero.
7. **RN-18.7 Sin registrar.** Un proyecto creado con HU-01 queda sin fecha de inicio (valor cero). En la API se muestra como `""`.
8. **RN-18.8 Copia por valor.** `SetStartDate` devuelve una copia del proyecto. `time.Time` es un valor, así que `Project` sigue sin slices, mapas ni punteros: modificar el resultado no altera el proyecto guardado.
9. **RN-18.9 Persistencia.** La fecha se guarda en memoria y en PostgreSQL, detrás de la misma interfaz `Repository`. En PostgreSQL se lee normalizada a UTC, como las fechas de los sprints (TEC-05) y de los esfuerzos (HU-26).
10. **RN-18.10 Concurrencia.** `SetStartDate` lee, modifica y escribe el proyecto con el mutex del `Manager` tomado durante toda la operación, como `Update` (HU-02), para que un registro de fecha no pise una modificación simultánea de nombre, descripción o estado, ni al revés.
11. **RN-18.11 API e interfaz.** La operación se expone en `PATCH /api/proyectos/{id}/fecha-inicio` y en la sección Proyectos de la interfaz web, con un caso exitoso y un test por código de error (tema 5 del Planning del Sprint 3).

## Decisión sobre entidades referenciadas

- **Proyecto:** **sí se valida su existencia** (RN-18.3), con `Repository.GetByID` de HU-03, dentro del mutex, porque la operación es leer-modificar-escribir y necesita los demás campos actuales para reescribirlos sin cambios. Se propaga `ErrProjectNotFound`.
- **Sprints del proyecto:** **no se consultan.** El Issue no relaciona la fecha del proyecto con la de sus sprints, y consultarlos haría que `project` dependa de `sprint`, que ya depende de `project` (ciclo de imports). Un proyecto puede tener sprints que empiezan antes de su fecha de inicio registrada.
- **Fecha de fin:** no existe todavía; la comparación inicio-fin la define HU-41.

## Decisiones de diseño (a confirmar por el Agile Enabler)

1. **D1 Operación dedicada.** `Manager.SetStartDate(id int, date time.Time) (Project, error)` en `internal/project/start_date.go`, y `PATCH /api/proyectos/{id}/fecha-inicio` con `{"fecha_inicio":"AAAA-MM-DD"}`. No se cambia `PUT /api/proyectos/{id}`: HU-02 reemplaza todos los campos que recibe, y si la fecha se agregara ahí, un cliente que no la mande (por ejemplo, el formulario "Guardar cambios" actual o los `curl` del README) la borraría. Es el mismo criterio de operación dedicada que HU-08 y HU-20.
2. **D2 Campo `StartDate time.Time` con cero como "sin registrar".** Se sigue el criterio de HU-10 (el valor cero de `time.Time` es fecha ausente) en lugar de un puntero, para que `Project` siga siendo un valor sin punteros (RN-18.8) y las comparaciones `==` de los tests existentes sigan compilando. `GET`, `POST` y `PUT` devuelven `fecha_inicio` en `AAAA-MM-DD` o `""`. Crear un proyecto no pide fecha.
3. **D3 Formato y errores.** `AAAA-MM-DD` con `dateLayout` y `parseDate` de `sprints.go`, sin moverlos ni cambiarlos (los dos están en el mismo paquete `httpapi`). Fecha vacía o ausente: `ErrMissingStartDate` (400), nuevo en `project`, porque `sprint.ErrMissingStartDate` habla de la fecha del sprint. Formato o fecha inválidos: `errInvalidDate` (400), el existente. Su mensaje, "las fechas deben tener el formato AAAA-MM-DD", sirve tal como está.
4. **D4 Reemplazo libre.** Se puede volver a registrar en cualquier estado del proyecto, sin restringir a fechas pasadas ni futuras (RN-18.4 y RN-18.5). La relación con la fecha de fin queda para HU-41.
5. **D5 Proyecto inexistente.** 404 con `ErrProjectNotFound` (HU-03). No se modifican nombre, descripción, estado, ítems ni sprints (RN-18.6).
6. **D6 Columna `start_date TIMESTAMPTZ NULL`.** Se elige `TIMESTAMPTZ` y no `DATE` por coherencia con el esquema: `sprints.start_date`, `sprints.end_date` y `efforts.effort_date` ya son `TIMESTAMPTZ`, y TEC-05 y HU-26 ya resolvieron la zona horaria para ese tipo (al leer, `.UTC()`, porque pgx devuelve `timestamptz` en `time.Local`; el test fija `time.Local` en UTC-3, porque así falló la primera versión de TEC-05). Como la API envía la fecha a las 00:00 UTC, se guarda y se lee como el mismo día en cualquier zona de la máquina. Es nullable, a diferencia de las fechas de los sprints, porque un proyecto puede no tener fecha: `NULL` es "sin registrar". El repositorio traduce el cero a `NULL` al escribir (si no, se guardaría el año 1 como una fecha real) y el `NULL` a cero al leer; una fecha no nula se lee con `.UTC()`. Los scripts de `db/init` solo corren con el volumen vacío: **una base anterior se recrea** con `docker compose down -v` y `docker compose up -d` (como en HU-29).
7. **D7 Concurrencia.** Bajo el mismo mutex del `Manager` de proyectos (RN-18.10), con un test de concurrencia que mezcla `SetStartDate`, `Update` y `Get` sobre el mismo proyecto y corre con `go test -race` en el CI.
8. **D8 Interfaz.** En la sección Proyectos, un campo de fecha y el botón "Registrar fecha de inicio" para el proyecto actual, y la fecha visible en los datos del proyecto (ver "Interfaz web").

## Restricciones

- **Organización del código.**
  - `internal/project`: `project.go` (campo `StartDate` en `Project`, con su comentario), `start_date.go` (nuevo: `SetStartDate` y `ErrMissingStartDate`), `postgres_repository.go` (`start_date` en `projectColumns`, en el `INSERT` de `Save`, en el `UPDATE` de `Update` y en `scanProject`) y `doc.go` (mencionar HU-18). No se tocan `create.go`, `get.go`, `update.go`, `manager.go`, `repository.go` ni `memory_repository.go`: el repositorio en memoria ya guarda el struct completo.
  - `internal/httpapi`: `projects.go` (handler `setProjectStartDate`, `startDateRequest`, `FechaInicio` en `projectResponse` y formateo en `toProjectResponse`) y `server.go` (ruta y `project.ErrMissingStartDate` en el 400 de `statusFor`). `sprints.go` no cambia.
  - `internal/web`: `index.html` (formulario de fecha en la sección Proyectos) y `app.js` (llamada al endpoint y fecha en `mostrarProyecto`).
  - `db/init/01-schema.sql`: columna `start_date TIMESTAMPTZ` (nullable) en `projects`.
  - `cmd/server`: no cambia. El `project.Manager` y `NewHandler` no cambian de firma.
- **Interfaz del repositorio.** No cambia. Se reutilizan `GetByID` (HU-03) y `Update` (HU-02); el contrato de `Save` y `Update` pasa a exigir que conserven `StartDate`. Los repositorios falsos de los tests (`failingRepository`, `fixedStatusRepository`) no se tocan.
- **Orden de validaciones.** Primero la fecha ausente (RN-18.1), sin consultar el repositorio. Después, con el mutex tomado: existencia (RN-18.3, `GetByID`) y escritura (`Update` del repositorio). Si falla cualquiera, no se escribe nada.
- **Errores identificables.** `ErrMissingStartDate` ("la fecha de inicio del proyecto es obligatoria") es nuevo, exportado y comparable con `errors.Is`. Se reutilizan `ErrProjectNotFound` y, en `httpapi`, `errInvalidDate`, `errInvalidJSON`, `errInvalidID` y `errBodyTooLarge`.
- **Copias por valor.** `SetStartDate` devuelve `Project` por valor, nunca un puntero al estado interno.
- **RED y GREEN.** En el RED se agregan solo el campo `StartDate`, `ErrMissingStartDate` y una firma vacía de `SetStartDate`, para que todo compile y los tests fallen por aserciones. La ruta `PATCH /api/proyectos/{id}/fecha-inicio` no se registra en `server.go` en el RED (el handler puede no existir todavía): sin la ruta, los tests de API fallan igual por aserción (404 o 405 en lugar de 200). El registro de la ruta, el handler, la columna del esquema y el repositorio PostgreSQL van en el GREEN. Antes del commit RED se corre `go vet ./...`.
- **Tests.** Unitarios en `internal/project/start_date_test.go`; pasos de godog en `internal/project/start_date_features_test.go`, con el runner `TestFeatures_HU18FechaInicioProyecto`, sobre `features/HU-18-fecha-inicio-proyecto.feature`. Cada escenario lleva su CA en el nombre (tema 7 del Planning).
- **Tests de PostgreSQL.** Requieren `TEST_DATABASE_URL` y la base levantada, y se omiten sin la variable (CA-TEC05.7). Los corre y verifica otro integrante antes del merge si el entorno local no tiene Docker.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
- **Definition of Done.** La SDD lleva la línea "Aprobada por"; antes del commit RED se corre `go vet ./...`; la concurrencia se verifica con `go test -race` en el CI; no hay firmas ni menciones de la IA en commits, PRs, Issues ni comentarios.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Registrar la fecha en un proyecto sin fecha | Queda con esa fecha; nombre, descripción y estado sin cambios |
| Volver a registrar otra fecha | La nueva reemplaza a la anterior |
| Registrar la misma fecha que ya tiene | Válido; queda igual |
| Proyecto `PLANIFICADO`, `EN_CURSO` o `FINALIZADO` | Se registra en los tres |
| Fecha pasada (`2020-01-01`) o futura (`2030-12-31`) | Se registra |
| Fecha `29/02` de un año bisiesto (`2028-02-29`) | Se registra |
| Fecha ausente (cero) | `ErrMissingStartDate`; el proyecto queda como estaba |
| Fecha ausente sobre un proyecto inexistente | `ErrMissingStartDate` (la fecha se valida primero) |
| ID inexistente, 0 o negativo, o sin proyectos creados | `ErrProjectNotFound` |
| Proyecto recién creado (HU-01) | Sin fecha: `StartDate` cero; en la API, `"fecha_inicio": ""` |
| Modificar el proyecto con `Update` (HU-02) después de registrar la fecha | La fecha se conserva |
| Modificar el proyecto devuelto | El proyecto guardado no cambia |
| `SetStartDate`, `Update` y `Get` simultáneos sobre el mismo proyecto | Todos terminan bien; al final el proyecto tiene la fecha de uno de los registros y los datos de una de las modificaciones; `go test -race` sin carreras |
| El repositorio falla al buscar o al actualizar | Se devuelve ese error y el proyecto queda como estaba |
| PostgreSQL: proyecto sin fecha | Se guarda `NULL` y se lee como cero |
| PostgreSQL: fecha registrada con `time.Local` en UTC-3 | Se lee en UTC y es el mismo instante y el mismo día |
| HTTP: `{"fecha_inicio": ""}`, `{"fecha_inicio": null}` o `{}` | 400 con el mensaje de `ErrMissingStartDate` |
| HTTP: `2026-02-30`, `2026-13-01`, `23/10/2026`, `2026-1-5`, `2026-10-23T00:00:00Z` o `" 2026-10-23 "` | 400 con el mensaje de `errInvalidDate` |
| HTTP: `{"fecha_inicio": 20261023}` | 400 con el mensaje de JSON inválido |
| HTTP: `{id}` no numérico | 400 con el mensaje de ID inválido |
| HTTP: fecha vacía sobre un proyecto inexistente | 400 (la fecha se valida antes que el proyecto) |
| HTTP: cuerpo de más de 1 MiB | 413 sin procesar el cuerpo |
| HTTP: `PUT /api/proyectos/{id}` con `fecha_inicio` en el cuerpo | Se ignora; la fecha registrada no cambia |
| Interfaz: registrar desde el formulario | Muestra "Fecha de inicio del proyecto N registrada: AAAA-MM-DD." y la fecha en los datos del proyecto |
| Interfaz: sin proyecto actual o con un error de la API | Muestra el mensaje de error en `msg-proyectos` |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Fecha ausente (valor cero) (RN-18.1) | `ErrMissingStartDate` | No se modifica nada |
| No existe el proyecto (RN-18.3; incluye 0 y negativos) | `ErrProjectNotFound` | No se modifica nada |
| El repositorio falla al buscar o actualizar | El error del repositorio | El proyecto queda como estaba |
| HTTP: formato distinto de `AAAA-MM-DD` o fecha inexistente (RN-18.2) | `errInvalidDate` (400) | No se llama al `Manager` |

**Orden de validación:** en el `Manager`, fecha ausente, existencia y guardado; en la API, antes, el ID de la ruta, el JSON y el formato de la fecha. Se informa el primer error encontrado.

## Criterios de aceptación

- **CA-18.1** Se registra la fecha de inicio de un proyecto existente: el proyecto queda con esa fecha y con su nombre, descripción y estado sin cambios. Se puede registrar en un proyecto `PLANIFICADO`, `EN_CURSO` o `FINALIZADO`, con una fecha pasada o futura.
- **CA-18.2** En la API, la fecha tiene que tener el formato `AAAA-MM-DD` y existir en el calendario: un formato distinto o una fecha como `2026-02-30` responde 400 con el mensaje de `errInvalidDate` y el proyecto no cambia.
- **CA-18.3** La fecha es obligatoria: una fecha ausente devuelve `ErrMissingStartDate` (400 en la API, también con `""`, `null` o el campo ausente) y el proyecto no cambia.
- **CA-18.4** Registrar la fecha en un proyecto que no existe (incluidos ID 0 y negativo) devuelve `ErrProjectNotFound` (404 en la API).
- **CA-18.5** Se puede volver a registrar: la fecha nueva reemplaza a la anterior.
- **CA-18.6** Un proyecto recién creado no tiene fecha de inicio (`"fecha_inicio": ""` en la API). La consulta (HU-03) muestra la fecha registrada, y modificar el proyecto con HU-02 la conserva.
- **CA-18.7** `SetStartDate` devuelve una copia por valor: modificar el resultado no altera el proyecto guardado.
- **CA-18.8** Los registros de fecha simultáneos con modificaciones y consultas del mismo proyecto no producen accesos sin protección (`go test -race`) y ninguno pisa los datos del otro. Si el repositorio falla, se devuelve ese error y el proyecto no cambia.
- **CA-18.9** La fecha se guarda en memoria y en PostgreSQL: un proyecto sin fecha se guarda como `NULL` y se lee sin fecha, y una fecha registrada se lee en UTC, el mismo día, sin importar la zona horaria de la máquina.
- **CA-18.10** Se registra desde la API (`PATCH /api/proyectos/{id}/fecha-inicio`, 200 en el caso exitoso y 400, 404, 413 o 500 según la tabla HTTP) y desde la sección Proyectos de la interfaz web, que muestra la fecha del proyecto actual. Se cubre con tests de `internal/httpapi`, no con escenarios godog.

## Pruebas previstas

Archivo BDD: `features/HU-18-fecha-inicio-proyecto.feature`, con el runner `TestFeatures_HU18FechaInicioProyecto` en `internal/project/start_date_features_test.go`. Cada escenario lleva su CA en el nombre:

- Registrar la fecha de inicio de un proyecto sin fecha (CA-18.1)
- Registrar la fecha en un proyecto finalizado (CA-18.1)
- Registrar una fecha futura (CA-18.1)
- Volver a registrar la fecha (CA-18.5)
- Un proyecto recién creado no tiene fecha de inicio (CA-18.6)
- Modificar el proyecto conserva la fecha de inicio (CA-18.6)
- Rechazar una fecha ausente (CA-18.3)
- Rechazar un proyecto inexistente (CA-18.4)
- Modificar el proyecto devuelto no altera el guardado (CA-18.7)
- Error del repositorio al guardar (CA-18.8)

CA-18.2 (formato) es de la capa HTTP: en el `Manager` la fecha ya llega como `time.Time`, así que se cubre con tests de API y no con godog. CA-18.9 y CA-18.10 también se cubren fuera de godog.

**Tests unitarios:** `internal/project/start_date_test.go` cubre CA-18.1, CA-18.3 a CA-18.8 y los casos límite del `Manager`, incluido el test de concurrencia (`SetStartDate`, `Update` y `Get` simultáneos) para `-race`. `update_test.go` suma un test de que `Update` conserva la fecha registrada.

**Tests de contrato:** `internal/project/repository_contract_test.go` suma que `Save` y `Update` conservan `StartDate` (cero y con fecha), con `MemoryRepository` y `PostgresRepository`.

**Tests de PostgreSQL:** `internal/project/postgres_repository_test.go` verifica el `NULL` ↔ cero y la lectura en UTC con `time.Local` fijado en UTC-3 (patrón de TEC-05). Se omiten sin `TEST_DATABASE_URL`.

**Tests de API:** en `internal/httpapi/server_test.go`, junto a los de proyectos de HU-01 a HU-03: un caso exitoso (200), uno por código de error (400 de fecha ausente, de formato, de fecha inexistente, de JSON y de ID; 404; 500), `fecha_inicio` en las respuestas de `POST`, `GET` y `PUT`, y que `PUT` conserva la fecha. Se agrega el endpoint a `bodyEndpoints` para 413 y texto después del JSON, y la ruta a `TestRutasConID_IDNoNumericoDevuelve400`. `projectJSON` suma `FechaInicio`.

**Tests de interfaz:** `internal/httpapi/web_test.go` verifica el formulario (campo `type="date"` obligatorio y botón "Registrar fecha de inicio"), la llamada de `app.js` a `/fecha-inicio` y que `mostrarProyecto` muestra `fecha_inicio`.

## Comparación con el acta de Planning

El acta del Sprint 3 (`docs/actas/sprint-3/planning.md`) incluye a HU-18 como "Registrar la fecha de inicio del proyecto", con 2 Story Points, responsable Sosa Ricardo, dependiente de HU-01 y bloqueante de HU-41 y HU-19, con la justificación "Campo nuevo en el proyecto y su validación". La tabla de división de historias la separa de HU-41 ("Fecha de fin (posterior a la de inicio)"). El tema 5 pide endpoint bajo `/api/…` y pantalla en la interfaz web; el tema 7, el CA en el nombre de cada escenario BDD.

**El alcance coincide:** un campo nuevo en el proyecto (`StartDate`), su validación (obligatoria y con formato `AAAA-MM-DD`), el endpoint y la interfaz; la fecha de fin queda en HU-41. Las precisiones que el acta no detalla son:

1. La operación es dedicada (`SetStartDate` y `PATCH`) y no se agrega a `PUT` (D1).
2. La columna `start_date` en PostgreSQL y la recreación de la base (D6).
3. La fecha se puede volver a registrar y no se limita a fechas pasadas ni futuras (D4).

Ninguna contradice el acta; la estimación de 2 Story Points se mantiene. La dependencia HU-01 está integrada en `main`. Además de HU-01, la SDD reutiliza HU-02 (`Update`, que tiene que conservar la fecha) y HU-03 (`GetByID` y `ErrProjectNotFound`), también en `main`.

## Comparación con los criterios del Issue #36 (borrador)

| Criterio en el Issue #36 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-18.1 Se registra la fecha de inicio de un proyecto existente. | CA-18.1 y CA-18.4 | Se aclara que los demás datos no cambian, que vale en cualquier estado y con fechas pasadas o futuras; el proyecto inexistente pasa a CA-18.4 |
| CA-18.2 La fecha debe tener un formato válido. | CA-18.2 y CA-18.3 | Se fija el formato `AAAA-MM-DD`, se rechazan fechas inexistentes y se separa la fecha ausente |
| — | CA-18.5 (nuevo) | Volver a registrar reemplaza la fecha |
| — | CA-18.6 (nuevo) | Proyecto sin fecha, consulta y conservación al modificar (HU-02 y HU-03) |
| — | CA-18.7 (nuevo) | Copia por valor |
| — | CA-18.8 (nuevo) | Concurrencia y falla del repositorio |
| — | CA-18.9 (nuevo) | Persistencia en PostgreSQL y zona horaria |
| — | CA-18.10 (nuevo) | API e interfaz web (tema 5 del Planning del Sprint 3) |

**Texto definitivo para la sección "Criterios de aceptación" del Issue #36** (en lenguaje de negocio, mejora 8 del Planning; se reemplaza cuando se apruebe la SDD y con autorización de Ricardo):

```markdown
## Criterios de aceptación

- [ ] CA-18.1 Se registra la fecha de inicio de un proyecto existente, sin cambiar su nombre, descripción ni estado; vale para proyectos planificados, en curso o finalizados, y para fechas pasadas o futuras.
- [ ] CA-18.2 La fecha tiene que tener el formato AAAA-MM-DD y existir en el calendario; si no, se rechaza y el proyecto no cambia.
- [ ] CA-18.3 La fecha es obligatoria; si falta, se rechaza y el proyecto no cambia.
- [ ] CA-18.4 Registrar la fecha en un proyecto que no existe devuelve error.
- [ ] CA-18.5 La fecha se puede volver a registrar: la nueva reemplaza a la anterior.
- [ ] CA-18.6 Un proyecto recién creado no tiene fecha de inicio; la consulta del proyecto muestra la fecha registrada y modificar sus datos la conserva.
- [ ] CA-18.7 Lo que se devuelve al registrar no permite cambiar el proyecto guardado.
- [ ] CA-18.8 Registrar la fecha mientras se modifica o consulta el mismo proyecto no pisa ningún dato, y un error al guardar deja el proyecto como estaba.
- [ ] CA-18.9 La fecha se conserva en la base de datos y se lee como el mismo día, sin importar la zona horaria.
- [ ] CA-18.10 La fecha se registra desde la API y desde la sección Proyectos de la interfaz web, que muestra la fecha del proyecto actual.
```

## Documentación a actualizar

- `README.md`: la fila `PATCH /api/proyectos/{id}/fecha-inicio` en la tabla de endpoints, un `curl` de ejemplo, la mención de `fecha_inicio` en las respuestas de proyectos y la aclaración de que `PUT /api/proyectos/{id}` no cambia la fecha; la nota de recrear la base (`docker compose down -v`); y la fila de HU-18 en "Historias por sprint".
- `docs/evidencias/HU-18-tdd.md`: evidencia TDD (paso 7).
- `docs/reportes/cobertura.md`: cobertura general (paso 7).
- `docs/ia/registro.md`: registro del uso de la IA en la historia.
- `internal/project/doc.go`: mencionar HU-18 en la documentación del paquete.

## Fuera de alcance

- La fecha de fin del proyecto y la regla "fin posterior al inicio" (HU-41).
- La duración del proyecto calculada y cualquier métrica, dashboard o reporte que la use.
- HU-19.
- Borrar una fecha de inicio registrada.
- Pedir la fecha al crear el proyecto o recibirla en `PUT /api/proyectos/{id}`.
- Relacionar la fecha de inicio del proyecto con las fechas de sus sprints.

HU-18: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
