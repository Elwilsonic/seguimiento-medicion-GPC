# HU-41 · Registrar la fecha de fin del proyecto

> Issue del repositorio: #115

Aprobada por: Peñalbé Hernán (Agile Enabler), 29/10/2026, con dos agregados (D4 y D5 cubren parte de HU-19, que sigue abierta y fuera de la entrega), que se aplicaron.

## Historia de usuario

Como Agile Enabler, quiero registrar la fecha de fin de un proyecto, para delimitar su duración.

**Épica:** E5 · Fechas del proyecto

**Prioridad:** MEDIA (la del Issue #115)

## Objetivo

Registrar (o volver a registrar) la fecha de fin de un proyecto existente que ya tiene fecha de inicio, con una operación dedicada del `Manager` de `internal/project` que solo cambia esa fecha. La fecha de fin tiene que ser posterior a la de inicio (CA-41.2 del borrador). El proyecto suma el campo `EndDate`, que se muestra en la consulta (HU-03) y se conserva al modificar el proyecto (HU-02).

Es la segunda parte de la historia original "Registrar fechas de inicio y fin del proyecto", dividida en el Planning del Sprint 3: HU-18 registró la fecha de inicio y dejó para esta historia la fecha de fin y su relación con la de inicio.

Se expone en `PATCH /api/proyectos/{id}/fecha-fin` y en la sección Proyectos de la interfaz web (tema 5 del Planning del Sprint 3). Persiste en memoria y en PostgreSQL (TEC-05).

## Contexto: cómo está modelada la fecha de inicio (HU-18)

- **Dominio:** `Project.StartDate time.Time` (`internal/project/project.go`); el valor cero significa "sin registrar" (RN-18.7). `Manager.SetStartDate(id, date)` (`start_date.go`) rechaza el cero con `ErrMissingStartDate` antes de tocar el repositorio y, con el mutex del `Manager` tomado, lee el proyecto con `GetByID`, reemplaza la fecha y guarda con `Update`. Se puede volver a registrar sin restricciones (RN-18.4 y RN-18.5).
- **Memoria:** `MemoryRepository` guarda el struct completo, así que no necesitó cambios.
- **PostgreSQL:** columna `projects.start_date TIMESTAMPTZ` nullable (`db/init/01-schema.sql`). `startDateParam` escribe el cero como `NULL` y `scanProject` lee `NULL` como cero y una fecha como `.UTC()`.
- **HTTP:** `PATCH /api/proyectos/{id}/fecha-inicio` con `{"fecha_inicio":"AAAA-MM-DD"}`, leído con `dateLayout` y `parseDate` de `internal/httpapi/sprints.go`. `projectResponse` devuelve `fecha_inicio` en `AAAA-MM-DD` o `""` en todas las respuestas de proyectos. Como `GET /api/proyectos/{id}/reporte` (HU-38) arma su campo `proyecto` con `toProjectResponse`, el reporte también lo incluye.
- **Web:** formulario "Fecha de inicio" (`form-fecha-inicio-proyecto`) en la sección Proyectos, y `mostrarProyecto` muestra "inicio: AAAA-MM-DD" o "sin fecha de inicio".

**El esquema no tiene una columna de fecha de fin para el proyecto.** `end_date` solo existe en `sprints`. HU-41 **sí cambia el esquema**: agrega `projects.end_date TIMESTAMPTZ` nullable, y una base creada antes se tiene que recrear (D7).

## Alcance

- **Incluye:**
  - el campo `EndDate` del proyecto;
  - la operación `SetEndDate` y sus errores;
  - la validación de `SetStartDate` contra una fecha de fin ya registrada (D5);
  - la columna `end_date` en `projects` y su lectura y escritura en el repositorio PostgreSQL;
  - el campo `fecha_fin` en las respuestas de proyectos;
  - el endpoint HTTP;
  - el formulario de fecha de fin y la fecha visible en la sección Proyectos de la interfaz web;
  - tests unitarios, BDD, de contrato, de PostgreSQL, de API y de interfaz.
- **No incluye:**
  - la duración calculada del proyecto ni cualquier métrica, dashboard o reporte que la use;
  - mostrar la fecha de fin en el texto del reporte de la interfaz (HU-38; ver D9);
  - HU-19;
  - pedir la fecha al crear el proyecto (HU-01) o recibirla en `PUT /api/proyectos/{id}` (HU-02);
  - borrar una fecha de fin ya registrada;
  - relacionar las fechas del proyecto con las fechas de sus sprints.

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | Identificador del proyecto |
| Fecha de fin | fecha (`time.Time`) | Sí | Fecha que delimita la duración del proyecto; el valor cero de `time.Time` es fecha ausente |

Firma: `SetEndDate(id int, date time.Time) (Project, error)`, en `internal/project/end_date.go` (nuevo).

La fecha se guarda tal como llega, como en HU-18. El formato de texto (`AAAA-MM-DD`) lo resuelve la capa HTTP.

### Salidas esperadas

- **Si el registro es válido:** el proyecto actualizado, como **copia por valor**, con `EndDate` igual a la fecha recibida y ID, nombre, descripción, estado y fecha de inicio iguales a los que tenía.
- **Si no es válido:** el error que indica el motivo (ver "Condiciones de error"); el proyecto guardado queda exactamente como estaba.

`Project` suma el campo `EndDate time.Time`. Un proyecto recién creado (HU-01) tiene `EndDate` en cero, que significa "sin registrar". `Create` no recibe la fecha y su firma no cambia.

### Cambio en `SetStartDate` (HU-18)

Para que la regla "fin posterior al inicio" se mantenga siempre (D5), `SetStartDate` suma una validación: si el proyecto ya tiene fecha de fin, la fecha de inicio nueva tiene que ser **anterior** a ella. Si no lo es, devuelve `ErrStartNotBeforeEnd` y el proyecto no cambia. Sin fecha de fin registrada, `SetStartDate` se comporta exactamente como en HU-18. Su firma no cambia.

### HTTP (TEC-04, tema 5 del Planning del Sprint 3)

`PATCH /api/proyectos/{id}/fecha-fin`, con el mismo patrón que `PATCH /api/proyectos/{id}/fecha-inicio` (HU-18).

- **Cuerpo:**

  ```json
  {"fecha_fin": "2026-12-18"}
  ```

- **Lectura de la fecha:** se reutilizan `dateLayout` (`AAAA-MM-DD`) y `parseDate` de `internal/httpapi/sprints.go`, sin cambios, como en HU-18. Una cadena vacía llega al `Manager` en cero (`ErrMissingEndDate`); un formato distinto o una fecha que no existe en el calendario devuelve `errInvalidDate`.
- **Respuesta exitosa:** `200 OK` con el proyecto:

  ```json
  {"id": 1, "nombre": "Sistema de ventas", "descripcion": "Gestión de ventas", "estado": "EN_CURSO", "fecha_inicio": "2026-10-23", "fecha_fin": "2026-12-18"}
  ```

- **`fecha_fin` en las demás respuestas de proyectos:** `POST /api/proyectos` (HU-01), `GET /api/proyectos/{id}` (HU-03), `PUT /api/proyectos/{id}` (HU-02) y `PATCH /api/proyectos/{id}/fecha-inicio` (HU-18) devuelven también `fecha_fin`, en `AAAA-MM-DD`, o `""` si no está registrada, igual que `fecha_inicio`. Como el reporte de HU-38 usa `toProjectResponse`, su campo `proyecto` también incluye `fecha_fin`. El campo no se agrega a `projectRequest`: `POST` y `PUT` lo ignoran si un cliente lo envía.
- **Errores de `PATCH /api/proyectos/{id}/fecha-fin`:**

| Caso | Código | Mensaje |
|---|---|---|
| `fecha_fin` vacía, `null` o ausente | 400 | El de `ErrMissingEndDate` ("la fecha de fin del proyecto es obligatoria") |
| `fecha_fin` con otro formato (`18/12/2026`, `2026-12-18T00:00:00Z`, `2026-1-5`, con espacios) o fecha inexistente (`2026-02-30`) | 400 | El de `errInvalidDate` |
| Cuerpo que no es JSON, `fecha_fin` que no es texto, o texto después del JSON | 400 | El de `errInvalidJSON` |
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| Proyecto inexistente | 404 | El de `ErrProjectNotFound` |
| El proyecto no tiene fecha de inicio | 409 | El de `ErrStartDateNotRegistered` ("el proyecto no tiene fecha de inicio registrada") |
| Fecha de fin igual o anterior a la de inicio | 409 | El de `ErrEndNotAfterStart` ("la fecha de fin debe ser posterior a la de inicio") |
| Cuerpo de más de 1 MiB | 413 | El de `errBodyTooLarge` |
| Falla del repositorio | 500 | El de `errInternal`, sin detalle |

- **Error nuevo de `PATCH /api/proyectos/{id}/fecha-inicio` (D5):**

| Caso | Código | Mensaje |
|---|---|---|
| Fecha de inicio igual o posterior a la fecha de fin registrada | 409 | El de `ErrStartNotBeforeEnd` ("la fecha de inicio debe ser anterior a la de fin") |

El orden de los errores del handler es: ID de la ruta, JSON, formato de la fecha (`parseDate`) y `Manager`. Dentro del `Manager`: fecha ausente (antes de tocar el repositorio), existencia del proyecto, fecha de inicio registrada y relación entre las fechas. Por eso, una fecha vacía sobre un proyecto inexistente responde 400 y no 404, y una fecha de fin sobre un proyecto inexistente responde 404 y no 409.

### Interfaz web (TEC-04, tema 5)

En la sección **Proyectos** de `internal/web`, debajo del formulario "Fecha de inicio":

- Un formulario **"Fecha de fin"** (`form-fecha-fin-proyecto`) para el **proyecto actual**, con un campo `<input type="date" name="fecha_fin" required>` y el botón **"Registrar fecha de fin"**.
- Llama a `PATCH /proyectos/{id}/fecha-fin` con el ID de «Proyecto actual», muestra el proyecto devuelto y el mensaje "Fecha de fin del proyecto N registrada: AAAA-MM-DD." o el mensaje de error de la API, con el mismo mecanismo que los demás formularios (`accion` y `msg-proyectos`).
- Los datos del proyecto actual (`mostrarProyecto`) muestran también la fecha de fin: "… fin: AAAA-MM-DD" o "sin fecha de fin" si `fecha_fin` es `""`. El campo de fecha de fin se precarga con la fecha registrada.
- "Guardar cambios" (HU-02) no envía la fecha y la conserva (RN-41.8).

## Reglas de negocio

1. **RN-41.1 Fecha obligatoria.** La fecha de fin es obligatoria en la operación: el valor cero de `time.Time` es fecha ausente y devuelve `ErrMissingEndDate`, sin consultar el repositorio, como RN-18.1.
2. **RN-41.2 Formato `AAAA-MM-DD`.** En la API la fecha llega como texto `AAAA-MM-DD`, el mismo formato de HU-18 y de las fechas de los sprints (HU-10); un formato distinto o una fecha que no existe en el calendario devuelve `errInvalidDate` (400). Es el criterio CA-41.3 del borrador.
3. **RN-41.3 Proyecto existente.** Requiere un proyecto con ese ID; si no existe (incluye 0 y negativos), se devuelve `ErrProjectNotFound` (HU-03).
4. **RN-41.4 Fecha de inicio registrada.** Para registrar la fecha de fin, el proyecto tiene que tener fecha de inicio (D2). Si no la tiene, se devuelve `ErrStartDateNotRegistered` y el proyecto no cambia.
5. **RN-41.5 Fin posterior al inicio.** La fecha de fin tiene que ser **estrictamente posterior** a la de inicio (D3), con el mismo criterio que RN-10.3 de los sprints: las fechas se comparan como instantes (`After`), y como la API recibe días, fin igual a inicio significa el mismo día y se rechaza con `ErrEndNotAfterStart`.
6. **RN-41.6 Se puede volver a registrar.** Registrar una fecha de fin en un proyecto que ya tiene una la reemplaza, con las mismas validaciones (D4). Registrar la misma fecha es válido y deja el proyecto igual.
7. **RN-41.7 Coherencia con la fecha de inicio.** Si el proyecto tiene fecha de fin, registrar una fecha de inicio (HU-18) igual o posterior a ella devuelve `ErrStartNotBeforeEnd` y el proyecto no cambia (D5). Sin fecha de fin, HU-18 sigue igual. Así, en ningún momento un proyecto queda con la fecha de fin igual o anterior a la de inicio.
8. **RN-41.8 Solo cambia la fecha de fin, y nada la borra.** `SetEndDate` no modifica el nombre, la descripción, el estado ni la fecha de inicio del proyecto, ni sus ítems o sprints. `Update` (HU-02) y `SetStartDate` (HU-18) conservan la fecha de fin, porque leen el proyecto guardado y solo reemplazan sus propios campos. Ninguna operación vuelve la fecha de fin a cero.
9. **RN-41.9 Cualquier estado y cualquier fecha.** Se registra en un proyecto `PLANIFICADO`, `EN_CURSO` o `FINALIZADO`, con una fecha pasada o futura, como RN-18.5. El Issue no relaciona la fecha de fin con el estado del proyecto.
10. **RN-41.10 Sin registrar.** Un proyecto creado con HU-01 queda sin fecha de fin (valor cero). En la API se muestra como `""`.
11. **RN-41.11 Copia por valor.** `SetEndDate` devuelve una copia del proyecto. `time.Time` es un valor, así que `Project` sigue sin slices, mapas ni punteros.
12. **RN-41.12 Persistencia.** La fecha se guarda en memoria y en PostgreSQL detrás de la misma interfaz `Repository`. En PostgreSQL, el cero se guarda como `NULL` y una fecha se lee normalizada a UTC, como `start_date` (RN-18.9).
13. **RN-41.13 Concurrencia.** `SetEndDate` lee, valida y escribe con el mutex del `Manager` de proyectos tomado durante toda la operación, como `SetStartDate` y `Update`. Así, un registro de fecha de fin y uno de fecha de inicio simultáneos no pueden dejar el proyecto con las fechas en un orden inválido, ni pisar una modificación de nombre, descripción o estado.
14. **RN-41.14 API e interfaz.** La operación se expone en `PATCH /api/proyectos/{id}/fecha-fin` y en la sección Proyectos de la interfaz web, con un caso exitoso y un test por código de error (tema 5 del Planning del Sprint 3).

## Decisión sobre entidades referenciadas

- **Proyecto:** **sí se valida su existencia** (RN-41.3), con `Repository.GetByID` dentro del mutex, porque la operación es leer-validar-escribir: necesita la fecha de inicio guardada para comparar y los demás campos para reescribirlos sin cambios. Se propaga `ErrProjectNotFound`.
- **Fecha de inicio:** no es una entidad, pero **sí se valida contra la guardada** (RN-41.4 y RN-41.5), leída del mismo proyecto bajo el mismo mutex.
- **Sprints del proyecto:** **no se consultan**, por el mismo motivo que en HU-18: el Issue no relaciona las fechas del proyecto con las de sus sprints, y consultarlos haría que `project` dependa de `sprint`, que ya depende de `project` (ciclo de imports). Un proyecto puede tener sprints que terminan después de su fecha de fin registrada.

## Decisiones de diseño (confirmadas por el Agile Enabler)

1. **D1 Operación dedicada y formato.** `Manager.SetEndDate(id int, date time.Time) (Project, error)` en `internal/project/end_date.go`, y `PATCH /api/proyectos/{id}/fecha-fin` con `{"fecha_fin":"AAAA-MM-DD"}`. Es el mismo formato y el mismo patrón que HU-18: `dateLayout` y `parseDate` sin cambios, `errInvalidDate` para el formato y un error propio (`ErrMissingEndDate`) para la fecha ausente, porque `sprint.ErrMissingEndDate` habla de la fecha del sprint. No se cambia `PUT /api/proyectos/{id}`, por el mismo motivo que D1 de HU-18: un cliente que no mande la fecha la borraría.
2. **D2 Sin fecha de inicio, se rechaza.** Si el proyecto no tiene fecha de inicio, `SetEndDate` devuelve `ErrStartDateNotRegistered` (409). El borrador pide "posterior a la de inicio", y sin inicio no hay con qué comparar. *Alternativa:* aceptar la fecha de fin sin inicio y validar la relación recién cuando se registre el inicio. Se descarta porque permite proyectos con fin y sin inicio, y obliga a que HU-18 haga la validación que le corresponde a esta historia.
3. **D3 Estrictamente posterior: el mismo día se rechaza.** Se sigue el texto del borrador ("posterior") y el criterio de los sprints (RN-10.3 de HU-10, `sprint.ErrEndNotAfterStart`): fin igual a inicio es un error. *Alternativa:* aceptar el mismo día (un proyecto de un día). Se descarta porque contradice el borrador y la regla ya usada en el proyecto para sprints.
4. **D4 La fecha de fin se puede modificar.** Registrarla de nuevo reemplaza la anterior, con las mismas validaciones, como RN-18.4 de HU-18. No hay operación para borrarla.
5. **D5 `SetStartDate` valida contra la fecha de fin registrada.** Es un cambio de comportamiento de HU-18: hoy su reemplazo es libre (RN-18.4), y sin este cambio se podría mover el inicio después del fin y romper la regla de RN-41.5. La validación solo se aplica si hay fecha de fin; los tests de HU-18 no registran fecha de fin, así que siguen valiendo sin cambios. *Alternativa:* no tocar HU-18 y aceptar que un proyecto quede con el inicio después del fin. Se descarta porque el criterio CA-41.2 dejaría de cumplirse con una operación de otra historia.
6. **D6 Códigos HTTP: 400 para la entrada, 409 para el estado guardado.** La fecha ausente y el formato inválido son 400, como en HU-18. "Sin fecha de inicio", "fin no posterior al inicio" e "inicio no anterior al fin" son **409**, porque la fecha recibida es válida por sí sola y choca con el estado guardado del proyecto, como `ErrResolutionBeforeDetection` de HU-30 y la definición de 409 del README ("el pedido choca con el estado actual"). *Alternativa:* 400 para "fin no posterior", como `sprint.ErrEndNotAfterStart`. La diferencia es que en el sprint las dos fechas llegan en el mismo pedido, y acá una de ellas ya está guardada. Los errores son nuevos y propios de `project` (`ErrEndNotAfterStart` repite el texto del de sprint, pero es otro error), para no importar `sprint` desde `project` (ciclo).
7. **D7 Columna `end_date TIMESTAMPTZ NULL` y recreación de la base.** Con el mismo criterio que D6 de HU-18 para `start_date`: `TIMESTAMPTZ` por coherencia con el resto del esquema, nullable porque `NULL` es "sin registrar", el cero se escribe como `NULL` y una fecha se lee con `.UTC()`. **Cambia el esquema:** los scripts de `db/init` solo corren con el volumen vacío, así que una base anterior se recrea con `docker compose down -v` y `docker compose up -d`, como en HU-18 y HU-29, y se avisa en el README.
8. **D8 Concurrencia.** Bajo el mismo mutex del `Manager` de proyectos (RN-41.13), con un test que mezcla `SetEndDate`, `SetStartDate`, `Update` y `Get` sobre el mismo proyecto y verifica que al final la fecha de fin siga siendo posterior a la de inicio. Corre con `go test -race` en el CI.
9. **D9 Interfaz: solo la sección Proyectos.** Un formulario "Fecha de fin" y la fecha visible en los datos del proyecto actual. El JSON del reporte (HU-38) incluye `fecha_fin` porque usa `toProjectResponse`: **el código de HU-38 no cambia, pero sus contratos JSON sí**. `TestReporteProyecto_ContratoDelEjemploDeTEC03` y `TestReporteProyecto_ContratoDelProyectoVacio` (`internal/httpapi/reports_test.go`) comparan con `expectRawJSON` el JSON exacto del campo `proyecto`, así que suman `"fecha_fin"`. Se le avisa a Cardozo Leandro, responsable de HU-38. Los ejemplos de `docs/especificaciones/HU-38-reporte-proyecto.md` no se modifican. **El texto del reporte en la interfaz no se modifica** (sigue mostrando solo el inicio): HU-38 es de otro responsable y el Issue no lo pide.

## Restricciones

- **Organización del código.**
  - `internal/project`: `project.go` (campo `EndDate` en `Project`, con su comentario), `end_date.go` (nuevo: `SetEndDate` y los errores `ErrMissingEndDate`, `ErrStartDateNotRegistered`, `ErrEndNotAfterStart` y `ErrStartNotBeforeEnd`), `start_date.go` (validación de RN-41.7 en `SetStartDate`), `postgres_repository.go` (`end_date` en `projectColumns`, en el `INSERT` de `Save`, en el `UPDATE` de `Update` y en `scanProject`; el helper de `NULL` se generaliza para las dos fechas) y `doc.go` (mencionar HU-41). No se tocan `create.go`, `get.go`, `update.go`, `manager.go`, `repository.go` ni `memory_repository.go`.
  - `internal/httpapi`: `projects.go` (handler `setProjectEndDate`, `endDateRequest`, `FechaFin` en `projectResponse` y formateo en `toProjectResponse`) y `server.go` (ruta, `ErrMissingEndDate` en el 400 y los otros tres errores en el 409 de `statusFor`). `sprints.go` y `reports.go` no cambian.
  - `internal/web`: `index.html` (formulario de fecha de fin) y `app.js` (llamada al endpoint y fecha de fin en `mostrarProyecto`).
  - `db/init/01-schema.sql`: columna `end_date TIMESTAMPTZ` (nullable) en `projects`, con su comentario.
  - `cmd/server`: no cambia.
- **Interfaz del repositorio.** No cambia. Se reutilizan `GetByID` y `Update`; el contrato de `Save` y `Update` pasa a exigir que conserven `EndDate`. Los repositorios falsos de los tests existentes no se tocan.
- **Orden de validaciones en `SetEndDate`.** Fecha ausente (RN-41.1), sin consultar el repositorio. Después, con el mutex tomado: existencia (RN-41.3), fecha de inicio registrada (RN-41.4), fin posterior al inicio (RN-41.5) y escritura. Si falla cualquiera, no se escribe nada.
- **Orden de validaciones en `SetStartDate`.** El de HU-18, con la comparación contra la fecha de fin (RN-41.7) después de la existencia y antes de la escritura.
- **Errores identificables.** Los cuatro errores nuevos son exportados y comparables con `errors.Is`. Se reutilizan `ErrProjectNotFound` y, en `httpapi`, `errInvalidDate`, `errInvalidJSON`, `errInvalidID` y `errBodyTooLarge`.
- **Copias por valor.** `SetEndDate` devuelve `Project` por valor.
- **RED y GREEN.** En el RED se agregan solo el campo `EndDate`, los cuatro errores y una firma vacía de `SetEndDate`, para que todo compile y los tests fallen por aserciones. La validación nueva de `SetStartDate`, la ruta, el handler, la columna del esquema y el repositorio PostgreSQL van en el GREEN (ajuste de Hernán en HU-18). Los contratos JSON del reporte de HU-38 (D9) se actualizan en el commit RED, porque el GREEN no toca tests: en el RED fallan por aserción, porque `toProjectResponse` todavía no devuelve `fecha_fin`. Antes del commit RED se corre `go vet ./...`.
- **Tests.** Unitarios en `internal/project/end_date_test.go`; pasos de godog en `internal/project/end_date_features_test.go`, con el runner `TestFeatures_HU41FechaFinProyecto`, sobre `features/HU-41-fecha-fin-proyecto.feature`. Cada escenario lleva su CA en el nombre. Los tests de RN-41.7 (`SetStartDate` con fecha de fin) van en `end_date_test.go`, para no modificar los tests de HU-18.
- **Tests de PostgreSQL.** Requieren `TEST_DATABASE_URL` y la base recreada con el esquema nuevo, y se omiten sin la variable. Los corre y verifica otro integrante antes del merge si el entorno local no tiene Docker.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Registrar la fecha de fin, posterior al inicio | Queda con esa fecha; nombre, descripción, estado y fecha de inicio sin cambios |
| Fecha de fin al día siguiente del inicio | Se registra |
| Fecha de fin igual a la de inicio (mismo día) | `ErrEndNotAfterStart` (409); el proyecto no cambia |
| Fecha de fin anterior a la de inicio | `ErrEndNotAfterStart` (409); el proyecto no cambia |
| Proyecto sin fecha de inicio | `ErrStartDateNotRegistered` (409); el proyecto no cambia |
| Volver a registrar otra fecha de fin válida | La nueva reemplaza a la anterior |
| Volver a registrar una fecha de fin anterior al inicio | `ErrEndNotAfterStart`; queda la fecha de fin anterior |
| Registrar la misma fecha de fin que ya tiene | Válido; queda igual |
| Proyecto `PLANIFICADO`, `EN_CURSO` o `FINALIZADO` | Se registra en los tres |
| Fecha de fin pasada o futura | Se registra si es posterior al inicio |
| Fecha `2028-02-29` (bisiesto) | Se registra si es posterior al inicio |
| Fecha ausente (cero) | `ErrMissingEndDate`; el proyecto queda como estaba |
| Fecha ausente sobre un proyecto inexistente | `ErrMissingEndDate` (la fecha se valida primero) |
| ID inexistente, 0 o negativo | `ErrProjectNotFound` |
| Fecha de fin sobre un proyecto inexistente | `ErrProjectNotFound` (la existencia se valida antes que las fechas) |
| Proyecto recién creado (HU-01) | Sin fecha de fin: `EndDate` cero; en la API, `"fecha_fin": ""` |
| Modificar el proyecto con `Update` (HU-02) | La fecha de fin se conserva |
| Cambiar la fecha de inicio a una anterior a la de fin | Se registra; la fecha de fin se conserva |
| Cambiar la fecha de inicio al mismo día o después de la de fin | `ErrStartNotBeforeEnd` (409); el proyecto no cambia |
| Cambiar la fecha de inicio de un proyecto sin fecha de fin | Igual que HU-18 |
| Modificar el proyecto devuelto | El proyecto guardado no cambia |
| `SetEndDate`, `SetStartDate`, `Update` y `Get` simultáneos | Todos terminan bien o con un error de fechas; al final la fecha de fin es posterior a la de inicio; `go test -race` sin carreras |
| El repositorio falla al buscar o al actualizar | Se devuelve ese error y el proyecto queda como estaba |
| PostgreSQL: proyecto sin fecha de fin | Se guarda `NULL` y se lee como cero |
| PostgreSQL: fecha de fin registrada con `time.Local` en UTC-3 | Se lee en UTC, el mismo día |
| HTTP: `{"fecha_fin": ""}`, `{"fecha_fin": null}` o `{}` | 400 con el mensaje de `ErrMissingEndDate` |
| HTTP: `2026-02-30`, `2026-13-01`, `18/12/2026`, `2026-1-5`, `2026-12-18T00:00:00Z` o `" 2026-12-18 "` | 400 con el mensaje de `errInvalidDate` |
| HTTP: `{"fecha_fin": 20261218}` | 400 con el mensaje de JSON inválido |
| HTTP: `{id}` no numérico | 400 con el mensaje de ID inválido |
| HTTP: cuerpo de más de 1 MiB | 413 sin procesar el cuerpo |
| HTTP: `PUT /api/proyectos/{id}` con `fecha_fin` en el cuerpo | Se ignora; la fecha registrada no cambia |
| HTTP: `GET /api/proyectos/{id}/reporte` | El campo `proyecto` incluye `fecha_fin` |
| Interfaz: registrar desde el formulario | Muestra "Fecha de fin del proyecto N registrada: AAAA-MM-DD." y la fecha en los datos del proyecto |
| Interfaz: sin proyecto actual o con un error de la API | Muestra el mensaje de error en `msg-proyectos` |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Fecha de fin ausente (valor cero) (RN-41.1) | `ErrMissingEndDate` | No se modifica nada |
| No existe el proyecto (RN-41.3; incluye 0 y negativos) | `ErrProjectNotFound` | No se modifica nada |
| El proyecto no tiene fecha de inicio (RN-41.4) | `ErrStartDateNotRegistered` | No se modifica nada |
| Fecha de fin igual o anterior a la de inicio (RN-41.5) | `ErrEndNotAfterStart` | No se modifica nada |
| `SetStartDate`: fecha de inicio igual o posterior a la de fin registrada (RN-41.7) | `ErrStartNotBeforeEnd` | No se modifica nada |
| El repositorio falla al buscar o actualizar | El error del repositorio | El proyecto queda como estaba |
| HTTP: formato distinto de `AAAA-MM-DD` o fecha inexistente (RN-41.2) | `errInvalidDate` (400) | No se llama al `Manager` |

**Orden de validación:** en la API, el ID de la ruta, el JSON y el formato de la fecha; en el `Manager`, fecha ausente, existencia, fecha de inicio registrada, relación entre las fechas y guardado. Se informa el primer error encontrado.

## Criterios de aceptación

- **CA-41.1** Se registra la fecha de fin de un proyecto existente que tiene fecha de inicio: el proyecto queda con esa fecha y con su nombre, descripción, estado y fecha de inicio sin cambios. Vale para proyectos `PLANIFICADO`, `EN_CURSO` o `FINALIZADO`.
- **CA-41.2** La fecha de fin tiene que ser posterior a la de inicio: una fecha igual (el mismo día) o anterior devuelve `ErrEndNotAfterStart` (409 en la API) y el proyecto no cambia.
- **CA-41.3** En la API, la fecha tiene que tener el formato `AAAA-MM-DD` y existir en el calendario: si no, responde 400 con el mensaje de `errInvalidDate` y el proyecto no cambia.
- **CA-41.4** La fecha es obligatoria: una fecha ausente devuelve `ErrMissingEndDate` (400 en la API, también con `""`, `null` o el campo ausente) y el proyecto no cambia.
- **CA-41.5** Registrar la fecha de fin en un proyecto que no existe (incluidos ID 0 y negativo) devuelve `ErrProjectNotFound` (404 en la API).
- **CA-41.6** Registrar la fecha de fin en un proyecto sin fecha de inicio devuelve `ErrStartDateNotRegistered` (409 en la API) y el proyecto no cambia.
- **CA-41.7** Se puede volver a registrar la fecha de fin: la nueva reemplaza a la anterior si cumple las mismas reglas.
- **CA-41.8** Con una fecha de fin registrada, cambiar la fecha de inicio (HU-18) al mismo día o a uno posterior al fin devuelve `ErrStartNotBeforeEnd` (409 en la API) y el proyecto no cambia; a un día anterior se acepta. Sin fecha de fin, la fecha de inicio se registra como en HU-18.
- **CA-41.9** Un proyecto recién creado no tiene fecha de fin (`"fecha_fin": ""` en la API). La consulta (HU-03) y el reporte del proyecto (HU-38) muestran la fecha registrada en su JSON, y modificar el proyecto (HU-02) o su fecha de inicio (HU-18) la conserva.
- **CA-41.10** `SetEndDate` devuelve una copia por valor: modificar el resultado no altera el proyecto guardado.
- **CA-41.11** Los registros de fechas simultáneos con modificaciones y consultas del mismo proyecto no producen accesos sin protección (`go test -race`) y nunca dejan la fecha de fin igual o anterior a la de inicio. Si el repositorio falla, se devuelve ese error y el proyecto no cambia.
- **CA-41.12** La fecha se guarda en memoria y en PostgreSQL: un proyecto sin fecha de fin se guarda como `NULL` y se lee sin fecha, y una fecha registrada se lee en UTC, el mismo día, sin importar la zona horaria de la máquina.
- **CA-41.13** Se registra desde la API (`PATCH /api/proyectos/{id}/fecha-fin`, 200 en el caso exitoso y 400, 404, 409, 413 o 500 según la tabla HTTP) y desde la sección Proyectos de la interfaz web, que muestra la fecha de fin del proyecto actual. Se cubre con tests de `internal/httpapi`, no con escenarios godog.

## Pruebas previstas

Archivo BDD: `features/HU-41-fecha-fin-proyecto.feature`, con el runner `TestFeatures_HU41FechaFinProyecto` en `internal/project/end_date_features_test.go`. Cada escenario lleva su CA en el nombre:

- Registrar la fecha de fin de un proyecto con fecha de inicio (CA-41.1)
- Registrar la fecha de fin en un proyecto finalizado (CA-41.1)
- Rechazar una fecha de fin igual o anterior a la de inicio (CA-41.2)
- Rechazar una fecha de fin ausente (CA-41.4)
- Rechazar un proyecto inexistente (CA-41.5)
- Rechazar la fecha de fin de un proyecto sin fecha de inicio (CA-41.6)
- Volver a registrar la fecha de fin (CA-41.7)
- Rechazar una fecha de inicio igual o posterior a la de fin registrada (CA-41.8)
- Aceptar una fecha de inicio anterior a la de fin registrada (CA-41.8)
- Un proyecto recién creado no tiene fecha de fin (CA-41.9)
- Modificar el proyecto o su fecha de inicio conserva la fecha de fin (CA-41.9)
- Modificar el proyecto devuelto no altera el guardado (CA-41.10)
- Error del repositorio (CA-41.11)
- Registros de fechas simultáneos mantienen el orden de las fechas (CA-41.11)

CA-41.3 (formato), CA-41.12 (PostgreSQL) y CA-41.13 (API e interfaz) se cubren fuera de godog.

**Tests unitarios:** `internal/project/end_date_test.go` cubre CA-41.1, CA-41.2 y CA-41.4 a CA-41.11, incluidos los casos de `SetStartDate` con fecha de fin y el test de concurrencia.

**Tests de contrato:** `internal/project/repository_contract_test.go` suma que `Save` y `Update` conservan `EndDate` (cero y con fecha), con `MemoryRepository` y `PostgresRepository`.

**Tests de PostgreSQL:** `internal/project/postgres_repository_test.go` verifica el `NULL` ↔ cero de `end_date` y la lectura en UTC con `time.Local` fijado en UTC-3, como en HU-18.

**Tests de API:** un archivo nuevo `internal/httpapi/project_end_date_test.go`, con un caso exitoso (200), uno por código de error (400 de fecha ausente, de formato, de fecha inexistente, de JSON y de ID; 404; los dos 409; 500), `fecha_fin` en las respuestas de `POST`, `GET`, `PUT` y `PATCH …/fecha-inicio`, el 409 nuevo de `PATCH …/fecha-inicio`, que `PUT` conserva la fecha y que el reporte la incluye. El endpoint se agrega a `bodyEndpoints` para 413 y texto después del JSON, y la ruta a `TestRutasConID_IDNoNumericoDevuelve400`. Si `projectJSON` necesita el campo `FechaFin`, se agrega en `server_test.go` sin cambiar los tests existentes. Los únicos tests existentes que cambian son `TestReporteProyecto_ContratoDelEjemploDeTEC03` y `TestReporteProyecto_ContratoDelProyectoVacio` de `internal/httpapi/reports_test.go` (D9): solo se agrega `"fecha_fin"` al JSON esperado del campo `proyecto` (`""` en los dos, porque ninguno registra fecha de fin).

**Tests de interfaz:** `internal/httpapi/web_test.go` verifica el formulario (campo `type="date"` obligatorio y botón "Registrar fecha de fin"), la llamada de `app.js` a `/fecha-fin` y que `mostrarProyecto` muestra `fecha_fin`.

## Comparación con el acta de Planning

- **Sprint 3** (`docs/actas/sprint-3/planning.md`): la tabla de división separa la historia original en "HU-18 Fecha de inicio · HU-41 Fecha de fin (posterior a la de inicio)". HU-41 queda con **2 Story Points**, "si hay tiempo", con la justificación "Posterior a la de inicio (HU-18)", y como posible tarea de Sosa Ricardo. La Review del Sprint 3 la registra como "No se empezó" y la pasa al Sprint 4, si hay tiempo.
- **Sprint 4** (`docs/actas/sprint-4/planning.md`): HU-41 es la primera de "Si hay tiempo, en este orden: HU-41, HU-24 y HU-14", asignada durante el Sprint a quien termine primero. El acta no detalla su alcance.

**El alcance coincide** con las actas: fecha de fin posterior a la de inicio, sobre el campo de HU-18, con 2 Story Points. HU-18 está integrada en `main`. Las precisiones que las actas no detallan, y que Hernán debería confirmar, son:

1. Sin fecha de inicio, la fecha de fin se rechaza (D2).
2. El mismo día se rechaza (D3).
3. **HU-41 modifica el comportamiento de `SetStartDate` de HU-18** (D5), para que la regla del acta ("posterior a la de inicio") se mantenga también cuando cambia el inicio.
4. **Cambia el esquema** (columna `projects.end_date`) y hay que recrear la base (D7).
5. **D4 y D5 cubren parte de HU-19** (Issue #37, "Modificar las fechas del proyecto"): volver a registrar las fechas y validar que el fin siga siendo posterior al inicio (CA-19.1 y CA-19.2 de su borrador). HU-19 sigue abierta y fuera de la entrega, y su Issue no se modifica.

Ninguna contradice las actas. Si el cambio de HU-18 (D5) o el del esquema hacen que la historia no entre en 2 Story Points, se ajusta en la Daily.

## Comparación con los criterios del Issue #115 (borrador)

| Criterio en el Issue #115 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-41.1 Se registra la fecha de fin de un proyecto existente. | CA-41.1, CA-41.5 y CA-41.6 | Se exige la fecha de inicio registrada, se aclara que los demás datos no cambian y el proyecto inexistente pasa a CA-41.5 |
| CA-41.2 La fecha de fin debe ser posterior a la de inicio. | CA-41.2 y CA-41.8 | Estrictamente posterior (el mismo día se rechaza); la regla se mantiene también al cambiar la fecha de inicio |
| CA-41.3 La fecha debe tener un formato válido. | CA-41.3 y CA-41.4 | Formato `AAAA-MM-DD`, fechas inexistentes rechazadas y la fecha ausente por separado |
| — | CA-41.7 (nuevo) | Volver a registrar reemplaza la fecha |
| — | CA-41.9 (nuevo) | Proyecto sin fecha de fin, consulta, reporte y conservación al modificar |
| — | CA-41.10 (nuevo) | Copia por valor |
| — | CA-41.11 (nuevo) | Concurrencia y falla del repositorio |
| — | CA-41.12 (nuevo) | Persistencia en PostgreSQL y zona horaria |
| — | CA-41.13 (nuevo) | API e interfaz web (tema 5 del Planning del Sprint 3) |

**Texto definitivo para la sección "Criterios de aceptación" del Issue #115** (en lenguaje de negocio; se carga después de aprobar la SDD y con autorización de Ricardo):

```markdown
## Criterios de aceptación

- [ ] CA-41.1 Se registra la fecha de fin de un proyecto existente que ya tiene fecha de inicio, sin cambiar sus otros datos; vale para proyectos planificados, en curso o finalizados.
- [ ] CA-41.2 La fecha de fin tiene que ser posterior a la de inicio; el mismo día o un día anterior se rechazan y el proyecto no cambia.
- [ ] CA-41.3 La fecha tiene que tener el formato AAAA-MM-DD y existir en el calendario; si no, se rechaza.
- [ ] CA-41.4 La fecha es obligatoria; si falta, se rechaza.
- [ ] CA-41.5 Registrar la fecha de fin en un proyecto que no existe devuelve error.
- [ ] CA-41.6 Un proyecto sin fecha de inicio no puede registrar la fecha de fin.
- [ ] CA-41.7 La fecha de fin se puede volver a registrar: la nueva reemplaza a la anterior si cumple las mismas reglas.
- [ ] CA-41.8 Con la fecha de fin registrada, la fecha de inicio no se puede cambiar al mismo día o a uno posterior al fin.
- [ ] CA-41.9 Un proyecto recién creado no tiene fecha de fin; la consulta y el reporte del proyecto muestran la fecha registrada, y modificar sus datos o su fecha de inicio la conserva.
- [ ] CA-41.10 Lo que se devuelve al registrar no permite cambiar el proyecto guardado.
- [ ] CA-41.11 Registrar fechas mientras se modifica o consulta el mismo proyecto no pisa ningún dato ni deja las fechas en un orden inválido, y un error al guardar deja el proyecto como estaba.
- [ ] CA-41.12 La fecha se conserva en la base de datos y se lee como el mismo día, sin importar la zona horaria.
- [ ] CA-41.13 La fecha de fin se registra desde la API y desde la sección Proyectos de la interfaz web, que muestra la fecha del proyecto actual.
```

## Documentación a actualizar

- `README.md`: la fila `PATCH /api/proyectos/{id}/fecha-fin` en la tabla de endpoints, un `curl` de ejemplo, la mención de `fecha_fin` en las respuestas de proyectos, la descripción del formulario en la sección Proyectos de la interfaz, los 409 nuevos en la frase de códigos de error, la nota de recrear la base (`docker compose down -v`) por la columna `end_date`, y la fila de HU-41 en "Historias por sprint".
- `docs/evidencias/HU-41-tdd.md`: evidencia TDD (paso 7).
- `docs/reportes/cobertura.md`: cobertura general (paso 7).
- `docs/ia/registro.md`: registro del uso de la IA en la historia.
- `internal/project/doc.go`: mencionar HU-41.

`docs/documentacion-tecnica.md` es de TEC-06 (responsable Peñalbé Hernán): se le avisa del endpoint, la columna y los errores nuevos, pero no se modifica en esta historia.

## Fuera de alcance

- La duración del proyecto calculada y cualquier métrica, dashboard o reporte que la use.
- Mostrar la fecha de fin en el texto del reporte de la interfaz (HU-38).
- HU-19 (Issue #37, "Modificar las fechas del proyecto"). D4 y D5 cubren parte de ella: volver a registrar las fechas y validar que el fin siga siendo posterior al inicio (CA-19.1 y CA-19.2 de su borrador). HU-19 sigue abierta y fuera de la entrega, y su Issue no se modifica.
- Borrar una fecha de fin registrada.
- Pedir la fecha al crear el proyecto o recibirla en `PUT /api/proyectos/{id}`.
- Relacionar las fechas del proyecto con las de sus sprints.
- Editar el Issue #115 desde esta SDD: los criterios definitivos se cargan después de la aprobación, con autorización de Ricardo.

HU-41: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
