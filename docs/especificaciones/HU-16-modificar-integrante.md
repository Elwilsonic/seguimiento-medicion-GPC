# HU-16 · Modificar los datos o el rol de un integrante

> Issue del repositorio: #30

Aprobada por: Peñalbé Hernán (Agile Enabler), 30/10/2026, con 2 ajustes (la opción A de P1: HU-16 cubre nombre y rol y HU-48 no se crea, y los métodos GetByID y Update de brokenMemberRepository en members_update_test.go, en el commit RED), que se aplicaron.

## Historia de usuario

Como Agile Enabler, quiero modificar el nombre o el rol de un integrante ya registrado, para mantener la información del equipo al día.

**Épica:** E4 · Registro de integrantes

**Prioridad:** MEDIA (la del Issue #30)

**Story Points:** 3 (acta de Planning del Sprint 2: "Como HU-02"), confirmados con la aprobación de esta SDD, que cubre nombre y rol (decisión D9).

## Objetivo

Modificar el nombre y el rol de un integrante registrado con HU-15, identificado por su ID, con las mismas reglas del registro: nombre recortado, no vacío y sin repetir en el proyecto sin distinguir mayúsculas, y uno de los tres roles de la metodología. El integrante conserva su ID y su proyecto. Se agregan la operación `Update` del `Manager`, las operaciones `GetByID` y `Update` del repositorio (memoria y PostgreSQL, sin cambios de esquema), el endpoint `PUT /api/integrantes/{id}` y el formulario "Modificar integrante" en la sección "Integrantes" de la interfaz web.

## Alcance

- **Incluye** (el cambio de nombre y el de rol, decisión D9):
  - la operación `Update(id int, name, role string) (Member, error)` del `Manager` de integrantes;
  - el error `ErrMemberNotFound`;
  - las operaciones `GetByID` y `Update` en la interfaz `Repository`, con sus implementaciones en memoria y en PostgreSQL y tests de contrato;
  - el endpoint `PUT /api/integrantes/{id}` y el 405 de `GET` en esa ruta;
  - el formulario "Modificar integrante" en la sección "Integrantes" de la interfaz web.
- **No incluye:**
  - consultar o listar los integrantes (HU-17);
  - cambiar el proyecto de un integrante ni su ID;
  - dar de baja un integrante;
  - actualizar el nombre en los esfuerzos (HU-26), en los votos de Planning Poker (HU-22) ni en los reportes (HU-38 y HU-39): no hay vínculo entre esos datos y los integrantes (D7 de HU-15; ver RN-16.8);
  - cambios en el formulario "Registrar integrante" de HU-15 ni en su endpoint.

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | ID del integrante (en la ruta) |
| Nombre | texto | Sí | Nombre nuevo del integrante; puede ser el mismo |
| Rol | texto | Sí | Rol nuevo: `PRODUCT_ARCHITECT`, `AGILE_ENABLER` o `PRODUCT_BUILDER`; puede ser el mismo |

Firma: `Update(id int, name, role string) (Member, error)`, en `internal/member/update.go`. La modificación reemplaza el nombre y el rol en una sola llamada: para cambiar uno solo, se reenvía el valor actual del otro, como en HU-02 y HU-07 (decisión D2).

### Salidas esperadas

- **Si la modificación es válida:** el integrante actualizado, como copia por valor (`Member` de HU-15), con:
  - **ID** y **ProjectID:** los mismos que tenía (no cambian);
  - **Name:** el recibido, sin espacios en los extremos;
  - **Role:** el recibido, normalizado.
- **Si no es válida:** el error que indica el motivo (ver "Condiciones de error"); el integrante queda exactamente como estaba.

### HTTP

`PUT /api/integrantes/{id}`

- **Cuerpo** (el mismo `memberRequest` de HU-15):

  ```json
  {"nombre": "Ana Pérez", "rol": "AGILE_ENABLER"}
  ```

- **200 OK** (el mismo `memberResponse` de HU-15):

  ```json
  {"id": 1, "proyecto_id": 1, "nombre": "Ana Pérez", "rol": "AGILE_ENABLER"}
  ```

- **Errores:**

  | Caso | Código | Mensaje |
  |---|---|---|
  | Nombre vacío, solo espacios o ausente | 400 | El de `ErrEmptyName` ("el nombre del integrante es obligatorio") |
  | Rol vacío o ausente | 400 | El de `ErrMissingRole` ("el rol del integrante es obligatorio") |
  | Rol que no es uno de los tres | 400 | El de `ErrInvalidRole` |
  | Cuerpo que no es JSON, campo que no es texto o texto después del JSON | 400 | El de `errInvalidJSON` |
  | `{id}` que no es un número entero | 400 | El de `errInvalidID` |
  | Integrante inexistente (incluidos 0 y negativos) | 404 | El de `ErrMemberNotFound` ("no existe un integrante con ese id") |
  | Otro integrante del proyecto ya tiene ese nombre | 409 | El de `ErrDuplicateMember` |
  | Cuerpo de más de 1 MiB | 413 | El de `errBodyTooLarge` |
  | Falla del almacenamiento | 500 | El de `errInternal`, sin detalle |
  | `GET /api/integrantes/{id}` | 405 | Texto de `http.StatusText(405)`, con `Allow: PUT` (como `methodNotAllowed` de HU-21) |
  | `POST`, `PATCH` o `DELETE` en `/api/integrantes/{id}` | 405 | El que responde `http.ServeMux` para un método no registrado |

  El orden de los errores es: ID de la ruta, JSON y `Manager`, como en `registerMember` (HU-15) y `updateProject` (HU-02).

  Ejemplo de error:

  ```json
  {"error": "no existe un integrante con ese id"}
  ```

### Interfaz web

En la sección **"Integrantes"** (`id="integrantes"`), después del formulario "Registrar integrante" de HU-15, que no cambia:

- Formulario **"Modificar integrante"** (`form-modificar-integrante`) con el campo "ID del integrante" (`name="id"`, numérico, obligatorio), el campo "Nombre" (`name="nombre"`, obligatorio), la lista "Rol" (`name="rol"`) con las mismas tres opciones de HU-15, y el botón "Guardar cambios".
- Llama a `PUT /api/integrantes/{id}`. No usa el proyecto actual: el ID del integrante es global (RN-15.7).
- Al guardar, muestra en `msg-integrantes` "Integrante N modificado: Nombre (Rol)." con el nombre legible del rol (`nombresDeRoles` de HU-15). Con un error de la API, su mensaje en `msg-integrantes`.
- Los textos se escriben con `textContent`, nunca con `innerHTML`. Sin librerías externas.

## Reglas de negocio

1. **RN-16.1 Nombre obligatorio.** Igual que RN-15.1: el nombre se guarda sin espacios en blanco en los extremos (`strings.TrimSpace`) y no puede quedar vacío (`ErrEmptyName`).
2. **RN-16.2 Rol obligatorio y válido.** Igual que RN-15.2, con la misma función `normalizeRole` de HU-15: se acepta sin distinguir mayúsculas ni contar los espacios de los extremos y se guarda normalizado; vacío es `ErrMissingRole` y otro valor `ErrInvalidRole`.
3. **RN-16.3 Integrante existente.** El integrante tiene que existir; si no, `ErrMemberNotFound`, incluidos los IDs 0 y negativos (ningún integrante tiene esos IDs, RN-15.7).
4. **RN-16.4 Sin duplicados, excluyendo al propio integrante.** Aplica RN-15.4 al modificar: el nombre nuevo no puede ser el de **otro** integrante del mismo proyecto, comparando sin distinguir mayúsculas (`strings.EqualFold`) y con cualquier rol (D1 de HU-15). El propio integrante se excluye de la comparación: conservar su nombre, o cambiar solo sus mayúsculas, es válido. Un integrante de otro proyecto con el mismo nombre no cuenta.
5. **RN-16.5 Proyecto e ID fijos.** La modificación no cambia el ID ni el proyecto del integrante: los toma de lo guardado, no de la entrada.
6. **RN-16.6 Sin cambios es válido.** Reenviar el mismo nombre y el mismo rol responde con el integrante tal como está; no es un error.
7. **RN-16.7 Orden de validación.** Nombre, rol, existencia del integrante, duplicado y guardado. Los errores de contenido se informan sin consultar el repositorio: con el nombre vacío y un ID inexistente se informa `ErrEmptyName`, como en RN-15.8 y en HU-02. Si algo falla, no se guarda nada.
8. **RN-16.8 Sin efecto en otros datos.** El nombre de un esfuerzo (HU-26) y el de un voto de Planning Poker (HU-22) son textos libres sin vínculo con los integrantes (D7 de HU-15): cambiar el nombre de un integrante no los modifica, y los reportes de HU-38 y HU-39 siguen mostrando el nombre con el que se registró cada esfuerzo.
9. **RN-16.9 Cualquier estado del proyecto.** Se puede modificar un integrante de un proyecto `PLANIFICADO`, `EN_CURSO` o `FINALIZADO`, como en RN-15.6.
10. **RN-16.10 Concurrencia.** La búsqueda del integrante, la lectura de los integrantes de su proyecto, la comparación de nombres y el guardado se hacen bajo el `sync.Mutex` del `Manager` de integrantes, el mismo que usa `Register` (RN-15.9). Así una modificación y un registro simultáneos, o dos modificaciones, no dejan dos integrantes con el mismo nombre en el proyecto.
11. **RN-16.11 Copias por valor.** `Update` del `Manager`, `GetByID` y `Update` del repositorio devuelven copias; modificar lo devuelto no cambia lo guardado.
12. **RN-16.12 Persistencia.** Memoria y PostgreSQL cumplen el mismo contrato para `GetByID` y `Update`, verificado con tests de contrato. No cambia el esquema: `UPDATE members SET name, role WHERE id` sobre las columnas de HU-15.

## Decisión sobre entidades referenciadas

- **Integrante:** **sí se valida que exista** (RN-16.3). Es la condición central de la historia (CA-16.3 del Issue): sin integrante no hay nada que modificar.
- **Proyecto:** **no se vuelve a validar.** El proyecto del integrante se validó al registrarlo (RN-15.3), los proyectos no se borran (no hay historia de baja) y en PostgreSQL `project_id` es una clave foránea. La modificación no recibe un proyecto: lo toma del integrante guardado (RN-16.5). Validarlo de nuevo con `project.Manager.Get` agregaría una consulta que nunca puede fallar por esta causa. *Alternativa descartada:* validar el proyecto como en `Register`, que solo agrega un camino de error inalcanzable sin un test que lo pida.
- **Esfuerzos y votos:** no se relacionan (RN-16.8).

## Restricciones

- **Organización del código** (archivos nuevos para no chocar con HU-17, ver "Conflictos con HU-17"):
  - `internal/member/update.go` (nuevo): `Update` del `Manager`, `ErrMemberNotFound` y la función `nameTaken(members []Member, name string, exceptID int) bool` con la comparación de RN-15.4.
  - `internal/member/register.go`: el bucle de duplicados de `Register` se reemplaza por `nameTaken(members, name, 0)` (ningún integrante tiene ID 0), para no duplicar la comparación. El comportamiento de `Register` no cambia y lo cubren los tests de HU-15 sin modificarlos.
  - `internal/member/repository.go`: dos métodos nuevos en la interfaz, `GetByID(id int) (Member, error)` y `Update(m Member) (Member, error)`. `Update` reemplaza el nombre y el rol del integrante con ese ID, sin cambiar su proyecto, y devuelve la copia guardada; los dos devuelven `ErrMemberNotFound` si no existe.
  - `internal/member/memory_repository_update.go` y `internal/member/postgres_repository_update.go` (nuevos): las dos operaciones en cada repositorio. En PostgreSQL, `SELECT ... WHERE id = $1` y `UPDATE members SET name = $2, role = $3 WHERE id = $1 RETURNING ...`; `pgx.ErrNoRows` se traduce a `ErrMemberNotFound`.
  - `internal/httpapi/members_update.go` (nuevo): el handler `updateMember` (reutiliza `memberRequest` y `memberResponse` de `members.go`, que no cambia) y `memberMethodNotAllowed` (405 con `Allow: PUT`).
  - `internal/httpapi/server.go`: dos rutas (`PUT` y `GET` de `/api/integrantes/{id}`), en un bloque propio que no queda junto a la ruta de HU-15, y `member.ErrMemberNotFound` en el caso 404 de `statusFor`.
  - `internal/web/index.html` y `app.js`: el formulario y su bloque de JavaScript, a continuación de los de HU-15.
- **Sin cambios de esquema** (verificado contra `db/init/01-schema.sql` y `postgres_repository.go`): la tabla `members` ya tiene `id`, `project_id`, `name` y `role`. No hace falta recrear la base ni tocar `internal/postgrestest`.
- **Sin cambios en `NewHandler` ni en `cmd/server`:** el `Manager` y los repositorios ya están armados desde HU-15.
- **Métodos nuevos en la interfaz.** Dos dobles de test implementan `member.Repository`:
  - `brokenMemberRepository` (`internal/httpapi/members_test.go`, el test del 500 de HU-15) deja de compilar al sumar `GetByID` y `Update` a la interfaz. Esos dos métodos se le agregan en el archivo nuevo `internal/httpapi/members_update_test.go`, en el commit RED; `members_test.go` no se toca.
  - `stubRepository` (`internal/member/register_test.go`) no necesita cambios, porque recibe esos métodos de `*MemoryRepository`, y `register_test.go` no se toca. Las fallas al buscar y al actualizar se prueban con un doble propio en `internal/member/update_test.go`.
- **Filas en `server_test.go`.** La ruta `PUT /api/integrantes/{id}` entra en las dos listas comunes, como hizo HU-15, en el commit RED y no en el GREEN. Es la única modificación de ese archivo: una línea por lista, sin comentario, separadas de las filas de HU-15.
  - En `TestRutasConID_IDNoNumericoDevuelve400`, después de "modificar ítem": ``{"modificar integrante", http.MethodPut, "/api/integrantes/abc", `{"nombre":"Ana","rol":"PRODUCT_BUILDER"}`},``.
  - En `bodyEndpoints`, al final de la lista: `{name: "modificar integrante", method: http.MethodPut, path: "/api/integrantes/1", untouched: expectNoMemberSaved},`. El entorno de esas pruebas no tiene integrantes, así que `expectNoMemberSaved`, declarada en `members_update_test.go`, comprueba que el primer registro válido recibe el ID 1, como la fila de HU-15.
- **RED.** Firmas vacías de `ErrMemberNotFound`, `Update` del `Manager` y de los repositorios, y `GetByID` (que devuelven un error), para que todo compile y los tests fallen por aserciones. La ruta, el handler, la interfaz y el cambio de `Register` a `nameTaken` van en el GREEN. `go vet ./...` antes del commit RED.
- **Librerías.** Solo la librería estándar y `pgx/v5`, ya aprobada; godog en los tests.

### Conflictos con HU-17

HU-17 (Consultar los integrantes, Issue #31, Buttini) trabaja en paralelo sobre el mismo paquete y la misma sección. Archivos compartidos y cómo se tocan:

| Archivo | Cambio de HU-16 | Riesgo |
|---|---|---|
| `internal/member/manager.go` | Ninguno | — |
| `internal/member/register.go` | El bucle de duplicados pasa a `nameTaken` | Bajo: HU-17 no necesita cambiar el registro |
| `internal/member/repository.go` | Dos métodos al final de la interfaz | Bajo: HU-17 puede usar `ListByProject`, que ya existe |
| `memory_repository.go` y `postgres_repository.go` | Ninguno (archivos nuevos `*_update.go`) | — |
| `internal/httpapi/members.go` | Ninguno (archivo nuevo `members_update.go`) | — |
| `internal/httpapi/server.go` | Dos rutas en un bloque aparte y una condición en el 404 de `statusFor` | Medio: si HU-17 agrega su ruta junto a la de HU-15, no queda contigua a las de HU-16. Si HU-17 define `GET /api/integrantes/{id}`, reemplaza el 405 de HU-16: se coordina en la Daily |
| `internal/httpapi/server_test.go` | Una línea en `TestRutasConID_IDNoNumericoDevuelve400` y una en `bodyEndpoints`, separadas de las filas de HU-15 | Bajo: HU-17 no recibe cuerpo, así que no suma filas a `bodyEndpoints`; su fila de ID no numérico queda en otro lugar de la lista |
| `internal/httpapi/members_test.go` | Ninguno: los métodos nuevos de `brokenMemberRepository` van en `members_update_test.go` | — |
| `internal/httpapi/web_test.go` | Ninguno: los tests de la interfaz van en `members_update_test.go` | — |
| `internal/web/index.html` | Un `<form>` entre el de HU-15 y `msg-integrantes` | Medio: si HU-17 agrega su lista en el mismo lugar, quien integre segundo conserva los dos bloques |
| `internal/web/app.js` | Un bloque nuevo después del de HU-15 | Medio: mismo criterio |
| `README.md`, `docs/reportes/cobertura.md`, `docs/ia/registro.md` | Filas y párrafos propios | Medio: conflictos de texto, se resuelven conservando las dos partes |

## Límites conocidos

- Sin HU-17 la interfaz no muestra los integrantes: para modificar uno hay que conocer su ID (lo informa el mensaje de "Registrar integrante").
- El formulario pide el nombre y el rol completos, aunque solo cambie uno (D2).
- Los esfuerzos y votos registrados con el nombre anterior lo conservan (RN-16.8).

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Cambiar el nombre y el rol | 200 con los dos valores nuevos; mismo ID y proyecto |
| Cambiar solo el nombre (se reenvía el rol actual) | Cambia el nombre; el rol queda igual |
| Cambiar solo el rol (se reenvía el nombre actual) | Cambia el rol; el nombre queda igual |
| Mismo nombre y mismo rol | Válido; el integrante queda igual (RN-16.6) |
| "Ana" pasa a "ANA" | Válido: no se compara con ella misma (RN-16.4) |
| Nombre nuevo "  Ana  " o rol `agile_enabler` | Se guardan "Ana" y `AGILE_ENABLER` |
| Nombre de otro integrante del proyecto, con otras mayúsculas o con otro rol | `ErrDuplicateMember`; los dos integrantes quedan igual |
| Nombre de un integrante de otro proyecto | Válido |
| Nombre vacío, solo espacios o ausente en el JSON | `ErrEmptyName`; no cambia nada |
| Rol vacío o ausente | `ErrMissingRole`; no cambia nada |
| Rol `Product Builder` (con espacio) o `SCRUM_MASTER` | `ErrInvalidRole`; no cambia nada |
| Cuerpo `{}` | `ErrEmptyName` (el nombre se valida primero) |
| Nombre vacío con ID inexistente | `ErrEmptyName` (RN-16.7) |
| ID inexistente, 0 o negativo | `ErrMemberNotFound` |
| Sin integrantes registrados | `ErrMemberNotFound` |
| Proyecto `FINALIZADO` | Válido (RN-16.9) |
| Integrante con esfuerzos registrados con su nombre | Los esfuerzos no cambian (RN-16.8) |
| Modificar el integrante devuelto | Lo guardado no cambia |
| Dos modificaciones simultáneas al mismo nombre, o una modificación y un registro | Una se guarda y la otra devuelve `ErrDuplicateMember`; `go test -race` sin carreras |
| Falla el repositorio al buscar, listar o actualizar | Se devuelve ese error; no cambia nada |
| PostgreSQL: modificar y leer | Mismo resultado que en memoria (contrato) |
| HTTP: `{"nombre": 5}`, texto después del JSON o cuerpo que no es JSON | 400, JSON inválido |
| HTTP: cuerpo de más de 1 MiB | 413 |
| HTTP: `GET`, `POST` o `DELETE` en `/api/integrantes/1` | 405 |

## Condiciones de error

| Condición | Error | HTTP |
|---|---|---|
| Nombre vacío o solo espacios (RN-16.1) | `member.ErrEmptyName` | 400 |
| Rol vacío (RN-16.2) | `member.ErrMissingRole` | 400 |
| Rol inválido (RN-16.2) | `member.ErrInvalidRole` | 400 |
| Integrante inexistente, incluidos 0 y negativos (RN-16.3) | `member.ErrMemberNotFound` | 404 |
| Nombre de otro integrante del proyecto (RN-16.4) | `member.ErrDuplicateMember` | 409 |
| Falla del repositorio | El error del repositorio | 500, sin detalle |

Los errores propios de HTTP (JSON inválido, ID no numérico, 413 y 405) están en la tabla de "HTTP".

## Criterios de aceptación

- **CA-16.1** Se pueden cambiar el nombre y el rol de un integrante existente, juntos o de a uno reenviando el valor actual del otro; conserva su ID y su proyecto, y queda con el nombre sin espacios en los extremos y el rol normalizado.
- **CA-16.2** No se acepta un rol vacío (`ErrMissingRole`) ni un rol que no sea Product Architect, Agile Enabler o Product Builder (`ErrInvalidRole`), ni un nombre vacío (`ErrEmptyName`); en esos casos el integrante queda como estaba.
- **CA-16.3** Modificar un integrante inexistente (incluidos 0 y negativos) devuelve `ErrMemberNotFound`.
- **CA-16.4** El nombre nuevo no puede ser el de otro integrante del mismo proyecto, sin distinguir mayúsculas y con cualquier rol (`ErrDuplicateMember`); conservar el propio nombre o cambiar solo sus mayúsculas es válido, y el nombre de un integrante de otro proyecto también.
- **CA-16.5** La modificación devuelve una copia: cambiarla no altera lo guardado. Dos modificaciones simultáneas no dejan dos integrantes con el mismo nombre en el proyecto (`go test -race`). Si falla el repositorio, se devuelve el error y no cambia nada.
- **CA-16.6** La modificación funciona igual en memoria y en PostgreSQL (tests de contrato; los de PostgreSQL se omiten sin `TEST_DATABASE_URL`).
- **CA-16.7** Se modifica desde la API (`PUT /api/integrantes/{id}`: 200, y 400, 404, 405, 409, 413 o 500 según la tabla) y desde el formulario "Modificar integrante" de la sección "Integrantes".

## Comparación con el acta de Planning

- **Sprint 2:** HU-16 "Modificar integrante", **3 SP**, "Como HU-02", dependencia HU-15, en "si hay tiempo". Coincide con el patrón de esta SDD (cuerpo completo, duplicado excluyendo el propio, mutex).
- **Sprint 3 y Sprint 4:** E4 queda fuera de la entrega, salvo capacidad sobrante. HU-15 se tomó así el 29/10/2026 y está integrada en `main` (PR #155).
- **Sprint 3, división de historias:** HU-16 se divide en **"HU-16 Modificar datos"** y **"HU-48 Cambiar rol"**, en "Si hay tiempo", con la nota "Pendiente: confirmar con los profesores [...] Los Issues nuevos y los cambios de título se hacen con autorización del Agile Enabler".

**Diferencia con el acta.** Esta SDD incluye el cambio de rol, que el acta del Sprint 3 asignó a HU-48: **HU-16 cubre nombre y rol, y HU-48 no se crea** (decisión D9). La opción fue confirmada por el Agile Enabler con la aprobación de esta SDD, el 30/10/2026, y se registra en la Review. Los hechos que la sustentan:

- el Issue #30 sigue titulado "Modificar los datos o el rol de un integrante" y sus criterios CA-16.1 y CA-16.2 piden cambiar el rol y rechazar un rol inválido;
- no existe un Issue de HU-48 en el repositorio, ni una estimación de las dos partes;
- el acta deja la división pendiente de confirmación.

## Comparación con los criterios del Issue #30 (borrador)

| Borrador | En esta SDD | Ajuste |
|---|---|---|
| CA-16.1 Se pueden cambiar el nombre y el rol de un integrante existente. | CA-16.1 | Se precisa que es una sola operación, que se conservan ID y proyecto, y la normalización. El cambio de rol sigue en HU-16 (D9) |
| CA-16.2 No se acepta un rol inválido. | CA-16.2 | Se separan rol vacío y rol inválido y se agrega el nombre vacío |
| CA-16.3 Modificar un integrante inexistente devuelve error. | CA-16.3 | Se identifica el error e incluye 0 y negativos |
| — | CA-16.4 a CA-16.7 (nuevos) | Duplicados excluyendo al propio integrante, copias y concurrencia, persistencia, API e interfaz |

El alcance del Issue, con nombre y rol, fue confirmado por el Agile Enabler con la aprobación de esta SDD, el 30/10/2026 (D9).

Con la aprobación de la SDD, los criterios definitivos reemplazan al borrador en el Issue #30, con autorización.

## Decisiones tomadas en la redacción

1. **D1 Ruta `PUT /api/integrantes/{id}`.** El ID del integrante es global (RN-15.7), como el de los ítems (`PUT /api/items/{id}`, HU-07) y los defectos (`PATCH /api/defectos/{id}/estado`, HU-30). *Alternativa descartada:* `PUT /api/proyectos/{id}/integrantes/{memberId}`, que obliga a validar que el integrante pertenezca a ese proyecto sin que ningún criterio lo pida.
2. **D2 Cuerpo completo, como HU-02 y HU-07.** `PUT` con nombre y rol obligatorios; para cambiar uno se reenvía el otro. Sin ninguno de los dos (`{}`) es `ErrEmptyName`, el primer error del orden (RN-16.7). *Alternativa descartada:* `PATCH` con campos opcionales, que necesita punteros en el pedido para distinguir "ausente" de "vacío", un error nuevo para "no viene ningún campo" y una convención que no usa ninguna otra ruta de modificación.
3. **D3 Error nuevo `ErrMemberNotFound`, declarado en `update.go`.** Como `ErrInvalidStatus` de HU-02 en su `update.go`, para no tocar `member.go`. *Alternativa descartada:* devolver `project.ErrProjectNotFound`, que informaría un motivo falso.
4. **D4 Reutilizar la validación de HU-15.** `Update` llama a `normalizeRole` tal como está, y la comparación de duplicados se extrae a `nameTaken(members, name, exceptID)`, que usan `Register` (con `exceptID` 0) y `Update` (con el ID del integrante). *Alternativa descartada:* copiar el bucle en `Update`, que dejaría la regla RN-15.4 en dos lugares.
5. **D5 405 explícito en `GET`.** La interfaz web atiende `GET /` y respondería a `GET /api/integrantes/{id}` con un 404 de archivo; se registra `GET /api/integrantes/{id}` con `memberMethodNotAllowed` (`Allow: PUT`), como `methodNotAllowed` de HU-21 lo hace con `POST`. `POST`, `PATCH` y `DELETE` ya reciben el 405 de `http.ServeMux`. *Alternativa descartada:* reutilizar `methodNotAllowed`, que anuncia `Allow: POST`.
6. **D6 Sin validar de nuevo el proyecto** (ver "Decisión sobre entidades referenciadas").
7. **D7 Archivos nuevos para no chocar con HU-17.** Los tests de la historia van en archivos nuevos; en `server_test.go` solo se suman las dos líneas de las listas comunes, que mantienen la convención de las demás rutas (ver "Restricciones"). *Alternativa descartada:* probar el 413, el texto después del JSON y el ID no numérico solo en `members_update_test.go`, que dejaría la ruta fuera de las listas comunes.
8. **D8 El formulario no usa el proyecto actual.** El ID del integrante alcanza para identificarlo (D1). *Alternativa descartada:* un botón "Modificar" en la lista de integrantes, que depende de HU-17.
9. **D9 Alcance respecto de HU-48 (P1): HU-16 cubre nombre y rol, y HU-48 no se crea.** Coincide con el Issue #30; el cambio de rol reutiliza `normalizeRole` y es el mismo endpoint, así que separarlo agregaría un segundo endpoint o una segunda historia para el mismo `UPDATE`. *Alternativa descartada:* HU-16 solo con el nombre y HU-48 con el rol, como dice el acta del Sprint 3, que obliga a crear otro Issue para una parte del mismo cambio.

## Pruebas previstas

- **Unitarios** (`internal/member/update_test.go`): cambio de nombre y rol; solo nombre; solo rol; sin cambios; solo mayúsculas del propio nombre; normalización; nombre vacío; rol vacío; rol inválido; orden de validación (nombre vacío con ID inexistente); duplicado con otra capitalización y con otro rol; nombre de otro proyecto; ID inexistente, 0 y negativo; proyecto `FINALIZADO`; conserva ID y proyecto; copia por valor; error del repositorio al buscar, listar y actualizar; modificaciones simultáneas con `-race`.
- **BDD** (`features/HU-16-modificar-integrante.feature`, runner `TestFeatures_HU16ModificarIntegrante` en `internal/member/update_features_test.go`): un escenario por criterio CA-16.1 a CA-16.5, con el CA en el nombre, cubriendo caso normal, alternativo, límite y error.
- **Contrato** (`internal/member/repository_update_contract_test.go`): `GetByID` devuelve una copia y `ErrMemberNotFound` si no existe; `Update` reemplaza nombre y rol, conserva el proyecto, devuelve una copia y `ErrMemberNotFound` si no existe; contra `MemoryRepository` y `PostgresRepository` (este último se omite sin `TEST_DATABASE_URL`).
- **API e interfaz** (`internal/httpapi/members_update_test.go`, con los métodos `GetByID` y `Update` de `brokenMemberRepository` y `expectNoMemberSaved`): 200 con el cuerpo exacto; un test por código de error (400 de nombre, rol vacío, rol inválido y JSON; 404 con 999, 0 y -1; 405 de `GET`, `POST` y `DELETE`; 409; 500); el formulario con sus campos y las tres opciones de rol, y que `app.js` llama a `PUT /integrantes/`.
- **Listas comunes** (`internal/httpapi/server_test.go`): la fila de `TestRutasConID_IDNoNumericoDevuelve400` (400 de ID) y la de `bodyEndpoints` (413 y texto después del JSON).

## Documentación a actualizar

- `README.md`: el endpoint, el uso del formulario "Modificar integrante" y la fila de HU-16.
- `docs/evidencias/HU-16-tdd.md`, `docs/reportes/cobertura.md` y `docs/ia/registro.md`.
- `docs/documentacion-tecnica.md` (TEC-06): la operación, el endpoint y el error nuevo, después del merge.

HU-16: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
