# HU-19 · Modificar las fechas del proyecto

> Issue del repositorio: #37

Aprobada por: pendiente

## Historia de usuario

Como Agile Enabler, quiero modificar las fechas de un proyecto, para reflejar cambios en la planificación.

**Épica:** E5 · Fechas del proyecto

**Prioridad:** MEDIA (la del Issue #37)

## Objetivo

Modificar **juntas** las fechas de inicio y de fin de un proyecto que ya las tiene registradas, con una operación del `Manager` de `internal/project` que valida el par nuevo entre sí (fin estrictamente posterior al inicio) y lo guarda de una sola vez. Hoy cada fecha se puede volver a registrar por separado (HU-18 y HU-41), pero mover las dos exige un orden y deja un estado intermedio (ver "Contexto").

Se expone en `PATCH /api/proyectos/{id}/fechas` y en un formulario "Modificar fechas" de la sección Proyectos de la interfaz web (tema 5 del Planning del Sprint 3). No cambia el esquema: reutiliza las columnas `projects.start_date` (HU-18) y `projects.end_date` (HU-41).

## Contexto

### Qué cubren ya HU-18 y HU-41

| Criterio del borrador del Issue #37 | Qué ya existe en `main` | Qué falta |
|---|---|---|
| CA-19.1 Se pueden cambiar las fechas de un proyecto que ya las tiene. | Cada fecha se puede volver a registrar por separado: la de inicio con `PATCH …/fecha-inicio` (RN-18.4 de HU-18: "Registrar una fecha en un proyecto que ya tiene una la reemplaza") y la de fin con `PATCH …/fecha-fin` (RN-41.6 y D4 de HU-41). | Cambiar las dos en una sola operación |
| CA-19.2 Se vuelve a validar que la fecha de fin sea posterior a la de inicio. | Al cambiar el fin se compara con el inicio guardado (RN-41.5, `ErrEndNotAfterStart`), y al cambiar el inicio se compara con el fin guardado (RN-41.7 y D5 de HU-41, `ErrStartNotBeforeEnd`). | Validar el par nuevo entre sí, sin depender de las fechas guardadas |
| CA-19.3 Modificar las fechas de un proyecto inexistente devuelve error. | Las dos operaciones devuelven `ErrProjectNotFound` (RN-18.3 y RN-41.3), 404 en la API. | El mismo error en la operación nueva |

La SDD de HU-41 lo registró así: "D4 y D5 cubren parte de HU-19 (Issue #37): volver a registrar las fechas y validar que el fin siga siendo posterior al inicio (CA-19.1 y CA-19.2 de su borrador)" (punto 5 de "Comparación con el acta de Planning" y "Fuera de alcance").

### El problema que queda

Con las dos operaciones separadas, **mover las dos fechas exige un orden**, porque cada una se valida contra la otra fecha **guardada**:

- **Postergar el proyecto** (por ejemplo, de 2026-10-23 – 2026-12-18 a 2027-01-11 – 2027-03-19): si se registra primero el inicio nuevo, 2027-01-11 no es anterior al fin viejo 2026-12-18 y `SetStartDate` lo rechaza con `ErrStartNotBeforeEnd` (409). Hay que registrar primero el fin y después el inicio.
- **Adelantar el proyecto** (por ejemplo, a 2026-06-01 – 2026-08-31): si se registra primero el fin nuevo, 2026-08-31 no es posterior al inicio viejo 2026-10-23 y `SetEndDate` lo rechaza con `ErrEndNotAfterStart` (409). Hay que registrar primero el inicio.

Además, entre los dos pedidos el proyecto queda con una fecha nueva y otra vieja, y si el segundo pedido falla queda así. HU-19 agrega una operación que recibe el par completo, lo valida entre sí y lo guarda de una vez, bajo el mutex.

## Alcance

- **Incluye:**
  - la operación `Manager.SetDates(id int, start, end time.Time) (Project, error)` en `internal/project/dates.go` (nuevo);
  - un error nuevo para "el proyecto no tiene las dos fechas registradas" (D2);
  - el endpoint `PATCH /api/proyectos/{id}/fechas`;
  - el formulario "Modificar fechas" en la sección Proyectos de la interfaz web;
  - tests unitarios, BDD, de API y de interfaz.
- **No incluye:**
  - cambios de esquema ni en los repositorios (memoria y PostgreSQL ya guardan las dos fechas);
  - cambios en `SetStartDate`, `SetEndDate` ni en sus endpoints, que siguen como en HU-18 y HU-41;
  - registrar las fechas por primera vez con este endpoint (D1);
  - borrar fechas registradas;
  - modificar una sola de las dos fechas con este endpoint (D3);
  - relacionar las fechas del proyecto con las de sus sprints;
  - la duración calculada del proyecto, métricas, dashboard o reportes.

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | Identificador del proyecto |
| Fecha de inicio | fecha (`time.Time`) | Sí | Nueva fecha de inicio; el valor cero es fecha ausente |
| Fecha de fin | fecha (`time.Time`) | Sí | Nueva fecha de fin; el valor cero es fecha ausente |

Firma: `SetDates(id int, start, end time.Time) (Project, error)`, en `internal/project/dates.go`.

### Salidas esperadas

- **Si la modificación es válida:** el proyecto actualizado, como **copia por valor**, con `StartDate` y `EndDate` iguales a las recibidas y ID, nombre, descripción y estado iguales a los que tenía.
- **Si no es válida:** el error que indica el motivo (ver "Condiciones de error"); el proyecto guardado queda exactamente como estaba (ninguna de las dos fechas cambia).

### Errores

Se reutilizan los errores existentes siempre que se puede:

| Error | Origen | Uso en HU-19 |
|---|---|---|
| `ErrMissingStartDate` ("la fecha de inicio del proyecto es obligatoria") | HU-18 | Fecha de inicio ausente |
| `ErrMissingEndDate` ("la fecha de fin del proyecto es obligatoria") | HU-41 | Fecha de fin ausente |
| `ErrProjectNotFound` ("no existe un proyecto con ese id") | HU-03 | Proyecto inexistente |
| `ErrEndNotAfterStart` ("la fecha de fin debe ser posterior a la de inicio") | HU-41 | Fin nuevo igual o anterior al inicio nuevo |
| **`ErrDatesNotRegistered`** ("el proyecto no tiene registradas las fechas de inicio y de fin") | **Nuevo** (D2) | El proyecto no tiene alguna de las dos fechas |

`ErrStartDateNotRegistered` (HU-41) no alcanza, porque habla solo de la fecha de inicio, y un proyecto puede tener inicio sin fin. `ErrStartNotBeforeEnd` (HU-41) no se usa: en HU-19 el par se compara entre sí y la regla se informa siempre desde el fin, con `ErrEndNotAfterStart`.

### HTTP (TEC-04, tema 5 del Planning del Sprint 3)

`PATCH /api/proyectos/{id}/fechas`, con el patrón de operación dedicada de HU-18 y HU-41.

- **Cuerpo:**

  ```json
  {"fecha_inicio": "2027-01-11", "fecha_fin": "2027-03-19"}
  ```

- **Lectura de las fechas:** las dos se leen con `parseDate` y `dateLayout` de `internal/httpapi/sprints.go`, sin cambios. Una cadena vacía o ausente llega al `Manager` en cero; un formato distinto de `AAAA-MM-DD` o una fecha inexistente devuelve `errInvalidDate`. Primero se lee `fecha_inicio` y después `fecha_fin`.
- **Respuesta exitosa:** `200 OK` con el proyecto, con la forma de `toProjectResponse` (sin campos nuevos):

  ```json
  {"id": 1, "nombre": "Sistema de ventas", "descripcion": "Gestión de ventas", "estado": "EN_CURSO", "fecha_inicio": "2027-01-11", "fecha_fin": "2027-03-19"}
  ```

- **Errores** (códigos propuestos en D4):

| Caso | Código | Mensaje |
|---|---|---|
| `fecha_inicio` vacía, `null` o ausente | 400 | El de `ErrMissingStartDate` |
| `fecha_fin` vacía, `null` o ausente (con `fecha_inicio` válida) | 400 | El de `ErrMissingEndDate` |
| `fecha_inicio` o `fecha_fin` con otro formato o inexistente | 400 | El de `errInvalidDate` |
| Cuerpo que no es JSON, una fecha que no es texto o texto después del JSON | 400 | El de `errInvalidJSON` |
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| Proyecto inexistente | 404 | El de `ErrProjectNotFound` |
| El proyecto no tiene las dos fechas registradas | 409 | El de `ErrDatesNotRegistered` |
| Fin nuevo igual o anterior al inicio nuevo | 409 | El de `ErrEndNotAfterStart` (ver D4) |
| Cuerpo de más de 1 MiB | 413 | El de `errBodyTooLarge` |
| Falla del repositorio | 500 | El de `errInternal`, sin detalle |

El orden de los errores del handler es: ID de la ruta, JSON, formato de `fecha_inicio`, formato de `fecha_fin` y `Manager`. Dentro del `Manager`: inicio ausente, fin ausente (los dos antes de tocar el repositorio), existencia, fechas registradas y relación entre el par nuevo.

### Interfaz web (TEC-04, tema 5)

En la sección **Proyectos** de `internal/web`, debajo del formulario "Fecha de fin":

- Un formulario **"Modificar fechas"** (`form-modificar-fechas-proyecto`) para el **proyecto actual**, con dos campos `<input type="date" required>` (`name="fecha_inicio"` y `name="fecha_fin"`) y el botón **"Modificar fechas"**.
- Llama a `PATCH /proyectos/{id}/fechas` con el ID de «Proyecto actual», muestra el proyecto devuelto y el mensaje "Fechas del proyecto N modificadas: inicio AAAA-MM-DD, fin AAAA-MM-DD." o el error de la API, con `accion` y `msg-proyectos`, como los demás formularios de la sección.
- `mostrarProyecto` precarga los dos campos con las fechas registradas.
- Los formularios "Fecha de inicio" y "Fecha de fin" no cambian.

## Reglas de negocio

1. **RN-19.1 Las dos fechas son obligatorias.** El inicio y el fin se reciben juntos; el valor cero de cualquiera es fecha ausente y devuelve `ErrMissingStartDate` o `ErrMissingEndDate`, en ese orden, sin consultar el repositorio (D3).
2. **RN-19.2 Formato `AAAA-MM-DD`.** En la API, las dos fechas llegan como texto `AAAA-MM-DD`, como en HU-18 y HU-41; un formato distinto o una fecha inexistente devuelve `errInvalidDate` (400).
3. **RN-19.3 Proyecto existente.** Si no existe un proyecto con ese ID (incluidos 0 y negativos), se devuelve `ErrProjectNotFound` (HU-03). Es CA-19.3 del borrador.
4. **RN-19.4 Solo se modifican fechas ya registradas.** El proyecto tiene que tener registradas la fecha de inicio y la de fin; si le falta alguna, se devuelve `ErrDatesNotRegistered` y no cambia nada (D1). Es el "que ya las tiene" de CA-19.1 del borrador.
5. **RN-19.5 El par nuevo se valida entre sí.** La fecha de fin nueva tiene que ser estrictamente posterior a la de inicio nueva (el mismo día se rechaza), con el criterio de RN-41.5 y RN-10.3. **No se compara con las fechas guardadas**: el par nuevo reemplaza al viejo completo, así que se puede postergar o adelantar el proyecto en una sola operación. Si no se cumple, `ErrEndNotAfterStart`. Es CA-19.2 del borrador.
6. **RN-19.6 Las dos fechas cambian juntas o ninguna.** Se guardan con un solo `Update` del repositorio; si algo falla, el proyecto queda con las dos fechas que tenía.
7. **RN-19.7 Solo cambian las fechas.** No se modifican el nombre, la descripción ni el estado del proyecto, ni sus ítems o sprints.
8. **RN-19.8 Cualquier estado y cualquier fecha.** Se modifica en un proyecto `PLANIFICADO`, `EN_CURSO` o `FINALIZADO`, con fechas pasadas o futuras, como RN-18.5 y RN-41.9.
9. **RN-19.9 Mismas fechas.** Enviar las mismas fechas que ya tiene es válido y deja el proyecto igual.
10. **RN-19.10 Copia por valor.** `SetDates` devuelve una copia del proyecto.
11. **RN-19.11 Concurrencia.** `SetDates` lee, valida y escribe con el mutex del `Manager` tomado durante toda la operación, como `SetStartDate`, `SetEndDate` y `Update`, así no se intercala con ellas.
12. **RN-19.12 Persistencia.** Sin cambios de esquema ni de repositorios: las dos fechas se guardan con `Repository.Update`, que ya escribe `start_date` y `end_date` (HU-18 y HU-41).
13. **RN-19.13 API e interfaz.** La operación se expone en `PATCH /api/proyectos/{id}/fechas` y en el formulario "Modificar fechas", con un caso exitoso y un test por código de error.

## Decisión sobre entidades referenciadas

- **Proyecto:** **sí se valida su existencia** (RN-19.3), con `Repository.GetByID` dentro del mutex, porque la operación es leer-validar-escribir y necesita saber si ya tiene las dos fechas (RN-19.4) y conservar sus demás campos. Se propaga `ErrProjectNotFound`.
- **Fechas guardadas:** se leen solo para verificar que existan (RN-19.4). **No se comparan** con el par nuevo (RN-19.5).
- **Sprints del proyecto:** **no se consultan**, por el mismo motivo que en HU-18 y HU-41 (el Issue no lo pide y `sprint` ya depende de `project`).

## Decisiones de diseño (a confirmar por el Agile Enabler)

0. **D0 HU-19 está "fuera de la entrega" en las actas.** El acta de Planning del Sprint 4 dice "Fuera de la entrega: E4 (HU-15 a HU-17), HU-19 y HU-28", y "Si hay tiempo, en este orden: HU-41, HU-24 y HU-14". El Sprint 3 dice lo mismo ("Fuera de la entrega, salvo capacidad sobrante"). Además, HU-19 **no tiene Story Points** estimados (ni en las actas ni en el Issue), y la Definition of Ready los pide. **Hernán tiene que decidir si HU-19 entra en el Sprint 4**, por delante o después de HU-24 y HU-14, y quién la estima. Esta SDD no da por desactualizadas las actas.
1. **D1 Se exige que el proyecto ya tenga las dos fechas.** Es lo que pide el borrador ("un proyecto que ya las tiene") y separa "modificar" (HU-19) de "registrar" (HU-18 y HU-41). *Alternativa:* permitir que este endpoint registre las fechas por primera vez. Es más cómoda para el cliente, pero duplica HU-18 y HU-41 y deja dos caminos para registrar, con reglas distintas (HU-41 exige el inicio antes que el fin).
2. **D2 Un error nuevo, `ErrDatesNotRegistered`.** Para "el proyecto no tiene las dos fechas registradas". `ErrStartDateNotRegistered` de HU-41 no alcanza, porque un proyecto puede tener inicio sin fin. *Alternativa:* reutilizar `ErrStartDateNotRegistered` si falta el inicio y agregar otro error solo para el fin; son dos errores en lugar de uno.
3. **D3 Las dos fechas son obligatorias.** Si se manda una sola, la otra llega en cero y se rechaza (`ErrMissingStartDate` o `ErrMissingEndDate`, 400). Para cambiar una sola fecha ya existen `PATCH …/fecha-inicio` y `PATCH …/fecha-fin`. *Alternativa:* una fecha ausente conserva la guardada; complica la regla de RN-19.5 (habría que comparar contra la guardada) y vuelve a los problemas de orden de "Contexto".
4. **D4 Códigos HTTP, con el criterio de D6 de HU-41** (400 para la entrada, 409 para el choque con el estado guardado):
   - fechas ausentes, formato inválido, JSON e ID: **400**;
   - proyecto inexistente: **404**;
   - proyecto sin las dos fechas (`ErrDatesNotRegistered`): **409**, porque el pedido es válido por sí solo y choca con el estado guardado;
   - **fin nuevo no posterior al inicio nuevo:** con el criterio de D6, es un error de la entrada (las dos fechas llegan en el mismo pedido, como en el sprint, que da 400). Pero `ErrEndNotAfterStart` ya está mapeado a **409** en `statusFor` para HU-41, y `statusFor` mapea por error, no por endpoint. Se propone **reutilizarlo con 409**, para que la misma regla dé el mismo error en las dos operaciones. *Alternativa:* un error nuevo para el par (por ejemplo `ErrInvalidDateRange`, con el mismo mensaje) mapeado a 400, coherente con D6 pero con un error más.
5. **D5 Orden de validación.** Inicio ausente, fin ausente, existencia, fechas registradas y relación del par. La relación del par se valida después de la existencia aunque no necesite el repositorio, para que un proyecto inexistente o sin fechas dé su error antes que un par inválido. *Alternativa:* validar el par antes de consultar el repositorio, como se hace con las fechas ausentes (un par inválido sobre un proyecto inexistente daría 409 y no 404).
6. **D6 Operación nueva y no un cambio en `PUT /api/proyectos/{id}`.** Por el mismo motivo que D1 de HU-18: `PUT` reemplaza todos los campos y un cliente que no mande las fechas las borraría.
7. **D7 Interfaz.** Un formulario aparte, "Modificar fechas", en lugar de cambiar los de HU-18 y HU-41, que siguen sirviendo para registrar cada fecha por separado.
8. **D8 PR #151 (HU-40) abierto.** Toca `internal/httpapi/server.go`, `server_test.go`, `web_test.go`, `internal/web/app.js`, `index.html`, `README.md`, `docs/reportes/cobertura.md` y `docs/ia/registro.md`, que HU-19 también va a tocar. Antes del RED conviene traer `main` con HU-40, si ya entró, para resolver los conflictos una sola vez.

## Restricciones

- **Organización del código.**
  - `internal/project`: `dates.go` (nuevo: `SetDates` y `ErrDatesNotRegistered`) y `doc.go` (mencionar HU-19). No se tocan `start_date.go`, `end_date.go`, `project.go`, los repositorios ni `repository.go`.
  - `internal/httpapi`: `projects.go` (handler `setProjectDates` y `datesRequest`) y `server.go` (ruta y `ErrDatesNotRegistered` en el 409 de `statusFor`; los demás errores ya están mapeados).
  - `internal/web`: `index.html` (formulario) y `app.js` (llamada al endpoint y precarga en `mostrarProyecto`).
  - Sin cambios en `db/init/01-schema.sql` ni en `cmd/server`.
- **Errores identificables.** `ErrDatesNotRegistered` es nuevo, exportado y comparable con `errors.Is`. Se reutilizan `ErrMissingStartDate`, `ErrMissingEndDate`, `ErrProjectNotFound`, `ErrEndNotAfterStart` y, en `httpapi`, `errInvalidDate`, `errInvalidJSON`, `errInvalidID` y `errBodyTooLarge`.
- **Copias por valor.** `SetDates` devuelve `Project` por valor.
- **RED y GREEN.** En el RED se agregan solo `ErrDatesNotRegistered` y una firma vacía de `SetDates`, para que todo compile y los tests fallen por aserciones. La ruta, el handler y el formulario van en el GREEN (ajuste de Hernán en HU-18). Antes del commit RED se corre `go vet ./...`.
- **Tests.** Unitarios en `internal/project/dates_test.go`; godog en `internal/project/dates_features_test.go`, con el runner `TestFeatures_HU19ModificarFechasProyecto`, sobre `features/HU-19-modificar-fechas-proyecto.feature`. No se modifican los tests de HU-18 ni de HU-41.
- **Tests de PostgreSQL.** No hay tests nuevos de PostgreSQL: el repositorio no cambia, y los contratos de HU-18 y HU-41 ya verifican que `Update` guarda las dos fechas. Igual se corre `go test ./...` con `TEST_DATABASE_URL` antes del merge, sin SKIP.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Postergar las dos fechas (inicio nuevo posterior al fin viejo) | Se modifican; no importa el orden |
| Adelantar las dos fechas (fin nuevo anterior al inicio viejo) | Se modifican |
| Cambiar solo el fin (mismo inicio) o solo el inicio (mismo fin) | Se modifica, si el fin sigue siendo posterior |
| Las mismas dos fechas que ya tiene | Válido; queda igual |
| Fin nuevo al día siguiente del inicio nuevo | Se modifica |
| Fin nuevo igual al inicio nuevo (mismo día) | `ErrEndNotAfterStart`; no cambia nada |
| Fin nuevo anterior al inicio nuevo | `ErrEndNotAfterStart`; no cambia nada |
| Inicio ausente (con o sin fin) | `ErrMissingStartDate`; no cambia nada |
| Fin ausente con inicio válido | `ErrMissingEndDate`; no cambia nada |
| Fechas ausentes sobre un proyecto inexistente | `ErrMissingStartDate` (las fechas se validan primero) |
| Proyecto inexistente, ID 0 o negativo | `ErrProjectNotFound` |
| Par inválido sobre un proyecto inexistente | `ErrProjectNotFound` (D5) |
| Proyecto sin fechas, solo con inicio o recién creado | `ErrDatesNotRegistered`; no cambia nada |
| Proyecto `PLANIFICADO`, `EN_CURSO` o `FINALIZADO` | Se modifica en los tres |
| `2028-02-29` como inicio o fin | Se acepta si el par es válido |
| Modificar el proyecto devuelto | El proyecto guardado no cambia |
| El repositorio falla al buscar o al actualizar | Se devuelve ese error; las dos fechas quedan como estaban |
| `SetDates`, `SetStartDate`, `SetEndDate`, `Update` y `Get` simultáneos | Sin carreras (`go test -race` en el CI); al final el fin es posterior al inicio |
| Después de `SetDates`, `PATCH …/fecha-inicio` o `…/fecha-fin` | Siguen validando contra la fecha guardada, como en HU-18 y HU-41 |
| HTTP: `{"fecha_inicio":"2027-01-11"}` (sin fin) | 400 con el mensaje de `ErrMissingEndDate` |
| HTTP: `{"fecha_fin":"2027-03-19"}` (sin inicio) | 400 con el mensaje de `ErrMissingStartDate` |
| HTTP: `{}`, `""` o `null` en las dos | 400 con el mensaje de `ErrMissingStartDate` |
| HTTP: `2027-02-30`, `11/01/2027`, `2027-1-11` o con espacios en cualquiera de las dos | 400 con el mensaje de `errInvalidDate` |
| HTTP: una fecha que no es texto | 400 con el mensaje de JSON inválido |
| HTTP: `{id}` no numérico | 400 con el mensaje de ID inválido |
| HTTP: cuerpo de más de 1 MiB | 413 |
| Interfaz: modificar desde el formulario | Muestra "Fechas del proyecto N modificadas: inicio …, fin …." y las fechas en los datos del proyecto |
| Interfaz: error de la API | Muestra el mensaje en `msg-proyectos` |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Fecha de inicio ausente (RN-19.1) | `ErrMissingStartDate` | No se modifica nada |
| Fecha de fin ausente (RN-19.1) | `ErrMissingEndDate` | No se modifica nada |
| No existe el proyecto (RN-19.3) | `ErrProjectNotFound` | No se modifica nada |
| El proyecto no tiene las dos fechas (RN-19.4) | `ErrDatesNotRegistered` | No se modifica nada |
| Fin nuevo igual o anterior al inicio nuevo (RN-19.5) | `ErrEndNotAfterStart` | No se modifica nada |
| El repositorio falla al buscar o actualizar | El error del repositorio | El proyecto queda como estaba |
| HTTP: formato distinto de `AAAA-MM-DD` o fecha inexistente (RN-19.2) | `errInvalidDate` (400) | No se llama al `Manager` |

**Orden de validación:** en la API, el ID de la ruta, el JSON y el formato de las dos fechas; en el `Manager`, inicio ausente, fin ausente, existencia, fechas registradas, relación del par y guardado. Se informa el primer error encontrado.

## Criterios de aceptación

- **CA-19.1** Las fechas de inicio y de fin de un proyecto que ya las tiene se modifican juntas, en una sola operación, sin importar si se postergan o se adelantan; el nombre, la descripción y el estado no cambian. Vale en cualquier estado del proyecto. *Ya cubierto en parte por HU-18 y HU-41 (cada fecha por separado); en esta historia se prueba la operación conjunta, incluidos los dos casos de "Contexto" que hoy exigen un orden.*
- **CA-19.2** La fecha de fin nueva tiene que ser posterior a la de inicio nueva; el mismo día o uno anterior se rechaza con `ErrEndNotAfterStart` (409 en la API) y no cambia ninguna de las dos. *Ya cubierto por HU-41 para cada fecha contra la guardada (RN-41.5 y RN-41.7); en esta historia se prueba el par nuevo entre sí, sin compararlo con las fechas guardadas.*
- **CA-19.3** Modificar las fechas de un proyecto inexistente (incluidos ID 0 y negativo) devuelve `ErrProjectNotFound` (404 en la API). *Ya cubierto en HU-18 y HU-41 para sus operaciones; en esta historia se prueba en `SetDates` y en el endpoint nuevo.*
- **CA-19.4** Solo se modifican fechas ya registradas: un proyecto sin alguna de las dos devuelve `ErrDatesNotRegistered` (409 en la API) y no cambia.
- **CA-19.5** Las dos fechas son obligatorias: si falta alguna, se rechaza con `ErrMissingStartDate` o `ErrMissingEndDate` (400 en la API), y en la API el formato tiene que ser `AAAA-MM-DD` (400 con el mensaje de `errInvalidDate`).
- **CA-19.6** Las dos fechas cambian juntas o ninguna: si el repositorio falla, se devuelve ese error y el proyecto queda como estaba. `SetDates` devuelve una copia por valor.
- **CA-19.7** Las modificaciones simultáneas con los demás registros de fechas y modificaciones del proyecto no producen accesos sin protección (`go test -race`) y nunca dejan el fin igual o anterior al inicio.
- **CA-19.8** Se modifica desde la API (`PATCH /api/proyectos/{id}/fechas`, 200 en el caso exitoso y 400, 404, 409, 413 o 500 según la tabla HTTP) y desde el formulario "Modificar fechas" de la sección Proyectos. Se cubre con tests de `internal/httpapi`, no con escenarios godog.

## Pruebas previstas

**BDD:** `features/HU-19-modificar-fechas-proyecto.feature`, con el runner `TestFeatures_HU19ModificarFechasProyecto`. Cada escenario lleva su CA en el nombre:

- Postergar las dos fechas de un proyecto (CA-19.1)
- Adelantar las dos fechas de un proyecto (CA-19.1)
- Modificar las fechas de un proyecto finalizado (CA-19.1)
- Rechazar un fin igual o anterior al inicio nuevo (CA-19.2, esquema)
- Rechazar un proyecto inexistente (CA-19.3, esquema con 999, 0 y -1)
- Rechazar un proyecto sin las dos fechas (CA-19.4, esquema: sin fechas y solo con inicio)
- Rechazar una fecha ausente (CA-19.5, esquema)
- Error del repositorio (CA-19.6)
- Modificar el proyecto devuelto no altera el guardado (CA-19.6)
- Modificaciones simultáneas mantienen el orden de las fechas (CA-19.7)

**Tests unitarios:** `internal/project/dates_test.go` cubre CA-19.1 a CA-19.7 y los casos límite del `Manager`, incluido el test de concurrencia que mezcla `SetDates`, `SetStartDate`, `SetEndDate`, `Update` y `Get`, con fechas elegidas para que el resultado de cada operación no dependa del orden (como el de HU-41).

**Tests de API:** un archivo nuevo `internal/httpapi/project_dates_test.go`, con un caso exitoso (200) por cada dirección (postergar y adelantar) y uno por código de error (400 de inicio ausente, de fin ausente, de formato, de JSON y de ID; 404; los dos 409; 500). El endpoint se agrega a `bodyEndpoints` (413 y texto después del JSON) y la ruta a `TestRutasConID_IDNoNumericoDevuelve400`.

**Tests de interfaz:** `internal/httpapi/web_test.go` verifica el formulario "Modificar fechas" (dos campos `type="date"` obligatorios y el botón), su posición debajo de "Fecha de fin", la llamada de `app.js` a `/fechas` y la precarga en `mostrarProyecto`.

**Tests existentes que podrían cambiar.** Se buscaron con `grep` los contratos JSON exactos (`expectRawJSON` y `"proyecto": {`) en `internal/httpapi/*_test.go`: `TestReporteProyecto_ContratoDelEjemploDeTEC03` y `TestReporteProyecto_ContratoDelProyectoVacio` (`reports_test.go`) comparan el JSON exacto del proyecto, y los de `sprint_reports_test.go` y `metrics_test.go` comparan sprints y métricas. Como HU-19 **no agrega campos** a `projectResponse`, **ninguno de esos contratos cambia**. Solo cambian, por casos de tabla nuevos, `TestRutasConID_IDNoNumericoDevuelve400`, `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400` (por `bodyEndpoints`), como en HU-18 y HU-41. No se modifican los tests de HU-18 ni de HU-41.

## Comparación con el acta de Planning

- **Sprint 2** (`docs/actas/sprint-2/planning.md`): HU-19 figura con prioridad MEDIA y dependencia HU-18, y pasa al Sprint 3.
- **Sprint 3** (`docs/actas/sprint-3/planning.md`): E5 "no es requerimiento mínimo de la consigna" y se toma solo con capacidad sobrante, salvo HU-18; HU-19 queda "Fuera de la entrega, salvo capacidad sobrante". La tabla de dependencias pone a HU-18 como bloqueante de HU-41 y HU-19.
- **Sprint 4** (`docs/actas/sprint-4/planning.md`): "Fuera de la entrega: E4 (HU-15 a HU-17), HU-19 y HU-28". La lista "Si hay tiempo" es HU-41, HU-24 y HU-14, y no incluye a HU-19.

**Hay una diferencia con las actas:** HU-19 está fuera de la entrega y no tiene Story Points. Esta SDD se escribió a pedido, pero **no da por desactualizadas las actas**: la decisión de tomarla, su estimación y su lugar respecto de HU-24 y HU-14 son de Hernán (D0). El alcance técnico no contradice las actas: depende de HU-18 y de HU-41, las dos en `main`, y no cambia el esquema.

Las precisiones que las actas no detallan, y que Hernán tiene que confirmar, son D1 a D5.

## Comparación con los criterios del Issue #37 (borrador)

| Criterio en el Issue #37 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-19.1 Se pueden cambiar las fechas de un proyecto que ya las tiene. | CA-19.1 y CA-19.4 | Las dos juntas, en una operación; el proyecto sin fechas se rechaza (CA-19.4) |
| CA-19.2 Se vuelve a validar que la fecha de fin sea posterior a la de inicio. | CA-19.2 | El par nuevo se valida entre sí, no contra las fechas guardadas |
| CA-19.3 Modificar las fechas de un proyecto inexistente devuelve error. | CA-19.3 | Sin cambios de fondo |
| — | CA-19.5 (nuevo) | Fechas obligatorias y formato |
| — | CA-19.6 (nuevo) | Todo o nada, falla del repositorio y copia por valor |
| — | CA-19.7 (nuevo) | Concurrencia |
| — | CA-19.8 (nuevo) | API e interfaz web |

**Texto propuesto para la sección "Criterios de aceptación" del Issue #37** (en lenguaje de negocio; se carga después de aprobar la SDD y con autorización de Ricardo):

```markdown
## Criterios de aceptación

- [ ] CA-19.1 Las fechas de inicio y de fin de un proyecto que ya las tiene se cambian juntas, en una sola operación, tanto para postergar como para adelantar el proyecto; sus otros datos no cambian.
- [ ] CA-19.2 Se vuelve a validar que la fecha de fin nueva sea posterior a la de inicio nueva; si no, no cambia ninguna.
- [ ] CA-19.3 Modificar las fechas de un proyecto inexistente devuelve error.
- [ ] CA-19.4 Un proyecto que no tiene registradas las dos fechas no las puede modificar con esta operación.
- [ ] CA-19.5 Las dos fechas son obligatorias y tienen que tener el formato AAAA-MM-DD.
- [ ] CA-19.6 Las dos fechas cambian juntas o ninguna, y lo que se devuelve no permite cambiar el proyecto guardado.
- [ ] CA-19.7 Modificar las fechas mientras se registran o modifican otras no deja las fechas en un orden inválido.
- [ ] CA-19.8 Las fechas se modifican desde la API y desde la sección Proyectos de la interfaz web.
```

## Documentación a actualizar

- `README.md`: la fila `PATCH /api/proyectos/{id}/fechas` en la tabla de endpoints, un `curl` de ejemplo, el formulario "Modificar fechas" en la descripción de la sección Proyectos, la mención en la frase que dice cómo se registran las fechas, `ErrDatesNotRegistered` en la frase de los 409, y la fila de HU-19 en la tabla del sprint que corresponda.
- `docs/evidencias/HU-19-tdd.md`, `docs/reportes/cobertura.md` y `docs/ia/registro.md` (paso 7).
- `internal/project/doc.go`: mencionar HU-19.

`docs/documentacion-tecnica.md` es de TEC-06 (responsable Peñalbé Hernán): se le avisa del endpoint y del error nuevo, pero no se modifica en esta historia.

## Fuera de alcance

- Registrar las fechas por primera vez con este endpoint (D1).
- Modificar una sola fecha con este endpoint (D3); para eso siguen HU-18 y HU-41.
- Cambios en `SetStartDate`, `SetEndDate` y sus endpoints.
- Borrar fechas registradas.
- Cambios de esquema o de repositorios.
- Relacionar las fechas del proyecto con las de sus sprints.
- La duración del proyecto y cualquier métrica, dashboard o reporte que la use.
- Editar el Issue #37 desde esta SDD: los criterios definitivos se cargan después de la aprobación, con autorización de Ricardo.

HU-19: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
