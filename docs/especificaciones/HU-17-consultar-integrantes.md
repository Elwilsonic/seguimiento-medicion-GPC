# HU-17 · Consultar los integrantes de un proyecto

> Issue del repositorio: #31

Aprobada por: Peñalbé Hernán (Agile Enabler), 30/10/2026, sin ajustes. Sugirió usar en los tests de `internal/member` un `project.Repository` que falle, como en los tests de `History` de HU-14.

## Historia de usuario

Como Agile Enabler, quiero ver los integrantes de un proyecto con sus roles, para conocer la composición del equipo.

**Épica:** E4 · Registro de integrantes

**Prioridad:** MEDIA (la del Issue #31)

**Story Points:** 3 (acta de Planning del Sprint 2: "Como HU-06")

## Objetivo

Listar los integrantes de un proyecto existente, con su nombre y su rol, para que el Agile Enabler vea la composición del equipo. Es la última historia de E4: los integrantes ya se registran (HU-15) y se modifican (HU-16), pero hoy no hay forma de verlos. La interfaz solo informa el ID de cada integrante en el mensaje de "Registrar integrante", y para modificar uno hay que conocer ese ID ("Límites conocidos" de HU-16).

HU-17 agrega el listado:

- en `member.Manager`, con la operación nueva `List`;
- en `GET /api/proyectos/{id}/integrantes`;
- en la sección "Integrantes" de la interfaz web.

## Alcance

- **Incluye:**
  - la operación `List(projectID int) ([]Member, error)` del `Manager` de `internal/member`, que valida el proyecto y devuelve sus integrantes;
  - la ruta `GET /api/proyectos/{id}/integrantes`;
  - en la interfaz, el botón "Ver integrantes del proyecto" y la tabla de integrantes.
- **Reutiliza sin cambios:**
  - `Repository.ListByProject` (HU-15, D5), en memoria y en PostgreSQL, ya cubierta por los tests de contrato: devuelve copias ordenadas por ID y una lista vacía, no `nil`, si no hay integrantes;
  - `project.Manager.Get` (HU-03), para validar el proyecto, como hace `Register` (RN-15.3);
  - `memberResponse` de `internal/httpapi/members.go` y `nombresDeRoles` de `app.js`.
- **No incluye:**
  - filtros por rol, búsqueda por nombre, paginación ni otro orden que no sea por ID;
  - consultar un integrante por su ID (`GET /api/integrantes/{id}` sigue respondiendo 405, D5 de HU-16);
  - un botón "Modificar" en cada fila de la tabla (decisión D4);
  - cambios en los reportes de HU-38 y HU-39, que siguen listando los integrantes que registraron esfuerzo (tema 5 del acta de Planning del Sprint 4);
  - cambios en el esquema de la base ni operaciones nuevas en el repositorio.

## Contrato

### Entradas

- `projectID int`: el ID del proyecto cuyos integrantes se consultan.

### Salidas esperadas

- `List(projectID int) ([]member.Member, error)`, en `internal/member/list.go`:
  - con un proyecto existente, devuelve una lista nueva con copias de sus integrantes, ordenados por ID ascendente, y `nil` como error;
  - con un proyecto sin integrantes, devuelve una lista vacía (no `nil`) y `nil` como error;
  - si hay un error, devuelve `nil` y el error.
- Cada integrante trae `ID`, `ProjectID`, `Name` y `Role` (`PRODUCT_ARCHITECT`, `AGILE_ENABLER` o `PRODUCT_BUILDER`).

### HTTP (TEC-04)

`GET /api/proyectos/{id}/integrantes`

- **200:** un arreglo JSON con un objeto por integrante, en el mismo formato que la respuesta de `POST /api/proyectos/{id}/integrantes` (`memberResponse`) y en el mismo orden que `List`. Un proyecto sin integrantes devuelve `[]`, no `null`.

  ```json
  [
    {"id":1,"proyecto_id":1,"nombre":"Ana Pérez","rol":"AGILE_ENABLER"},
    {"id":2,"proyecto_id":1,"nombre":"Juan Gómez","rol":"PRODUCT_BUILDER"}
  ]
  ```

- **400:** el `{id}` no es un número entero (`errInvalidID`, por `pathID`).
- **404:** el proyecto no existe (`project.ErrProjectNotFound`), incluidos los IDs 0 y negativos.
- **500:** falla el almacenamiento, sin el detalle del error (`writeError`).

Las respuestas de error usan el formato de siempre, `{"error": "..."}`, con el código que ya asigna `writeError`; no se agregan códigos nuevos. La ruta comparte el patrón con `POST /api/proyectos/{id}/integrantes` y se distingue por el método.

### Interfaz web (TEC-04)

En la sección "Integrantes" (`id="integrantes"`), después del formulario "Modificar integrante" de HU-16 y antes de `msg-integrantes`:

- **Botón "Ver integrantes del proyecto"** (`ver-integrantes`): llama a `GET /api/proyectos/{proyecto actual}/integrantes` y llena la tabla `lista-integrantes`.
- **Columnas de la tabla:** ID, Nombre y Rol. El rol se muestra con su nombre legible (`nombresDeRoles`: "Agile Enabler", no `AGILE_ENABLER`), como en los mensajes de HU-15 y HU-16.
- **Mensajes:** al terminar, en `msg-integrantes`, "N integrante(s) en el proyecto X.". Un proyecto sin integrantes deja la tabla vacía y muestra "0 integrante(s) en el proyecto X.".
- **Errores:** sin proyecto actual, se muestra el mensaje de siempre de `proyectoActual()`. Con un proyecto inexistente, se muestra el error de la API. Los dos van en `msg-integrantes`.
- **La tabla no se actualiza sola** al registrar o modificar un integrante: se vuelve a pulsar el botón, como la tabla de sprints de HU-14.
- **Seguridad:** los datos se escriben con `textContent`, nunca con `innerHTML`.

## Reglas de negocio

- **RN-17.1:** `List` devuelve todos los integrantes del proyecto, con cualquier rol, con su ID, proyecto, nombre y rol (CA-17.1).
- **RN-17.2:** el proyecto tiene que existir. Se valida con `project.Manager.Get` antes de leer los integrantes, y su error se propaga tal cual: `ErrProjectNotFound`, incluidos los IDs 0 y negativos. Un proyecto `FINALIZADO` se puede consultar, porque consultar no cambia nada.
- **RN-17.3:** los integrantes se devuelven ordenados por ID ascendente, que es el orden en que se registraron, como los devuelve `ListByProject` (D5 de HU-15).
- **RN-17.4:** un proyecto existente sin integrantes devuelve una lista vacía, no `nil` ni un error (CA-17.2).
- **RN-17.5:** solo se devuelven los integrantes del proyecto consultado. Un integrante de otro proyecto no aparece, aunque tenga el mismo nombre (CA-17.3).
- **RN-17.6:** `List` devuelve una lista nueva con copias por valor de los integrantes. Modificar la lista o sus elementos no cambia los integrantes guardados. `Member` no tiene slices ni mapas (RN-15.10), así que la copia de cada elemento no comparte estado.
- **RN-17.7:** la validación del proyecto se hace antes de tomar el mutex del `Manager` de integrantes, igual que en `Register` (RN-15.9), para no tomar dos mutex a la vez. La lectura de los integrantes se hace bajo ese mutex, porque `MemoryRepository` no es seguro para uso concurrente y `Register` y `Update` escriben bajo el mismo mutex.
- **RN-17.8:** si el proyecto no existe, no se leen los integrantes. Si falla la lectura del proyecto o la de los integrantes, se devuelve ese error y `nil`, sin una lista parcial.
- **RN-17.9:** la lista muestra los datos actuales: un integrante modificado con HU-16 aparece con su nombre y su rol nuevos.

## Decisión sobre entidades referenciadas

- **El proyecto se valida.** Sin esa validación, un proyecto inexistente y un proyecto sin integrantes darían la misma lista vacía, y CA-17.2 no se podría distinguir de un error. Es la misma decisión de HU-14 (`History`) y de las demás consultas por proyecto (HU-06, HU-31), y la que anticipó HU-15: "HU-17 listaría integrantes de un proyecto que no se puede consultar".
- **Los integrantes no se validan uno por uno:** salen del repositorio filtrados por proyecto, así que todos existen y pertenecen a él.

## Decisiones de diseño

- **D1. Operación nueva `List` en el `Manager`, en un archivo nuevo (`list.go`).** `Repository.ListByProject` no valida el proyecto ni toma el mutex: el `Manager` agrega las dos cosas, como `History` de HU-14 sobre `ListByProject`. El archivo nuevo evita tocar `register.go` y `update.go`. *Alternativa descartada:* exponer `ListByProject` en el `Manager` sin validar el proyecto, que no distingue un proyecto inexistente de uno vacío.
- **D2. Ruta anidada en el proyecto, `GET /api/proyectos/{id}/integrantes`.** Es la misma ruta del registro (HU-15) con otro método, como `GET` y `POST /api/proyectos/{id}/backlog` (HU-04 y HU-06) y `/sprints` (HU-10 y HU-14). *Alternativa descartada:* `GET /api/integrantes?proyecto=N`, que agrega un parámetro de consulta que no usa ninguna otra ruta.
- **D3. Orden por ID.** El Issue no pide un orden. El ID es el orden de registro y es el que ya devuelve `ListByProject` en memoria y en PostgreSQL (`ORDER BY id`). *Alternativa descartada:* ordenar por nombre o por rol, que necesita una regla nueva que el Issue no pide.
- **D4. Sin botón "Modificar" por fila.** La tabla muestra el ID de cada integrante, que alcanza para usar el formulario "Modificar integrante" de HU-16. Un botón que complete ese formulario es funcionalidad que no pide ningún criterio (`AGENTS.md`: no agregar funcionalidad no pedida). *Alternativa descartada:* el botón por fila que D8 de HU-16 dejó para después de HU-17.
- **D5. Formato HTTP:** la respuesta reutiliza `memberResponse`, para que cada integrante tenga el mismo formato en el registro, la modificación y el listado. El rol viaja como `PRODUCT_BUILDER` y la interfaz lo traduce.

## Restricciones

- Solo la librería estándar de Go y `net/http`. No se agregan librerías.
- El código de negocio accede a los integrantes solo a través de `Repository` (memoria o PostgreSQL). No hace falta una operación nueva en el repositorio, porque `ListByProject` ya existe en las dos implementaciones y la cubren los tests de contrato de HU-15.
- No cambia el esquema de la base ni los contratos de las rutas existentes.
- La interfaz no usa frameworks ni `innerHTML`.
- **Archivos compartidos con HU-15 y HU-16:** los tests de la historia van en archivos nuevos (`list_test.go`, `list_features_test.go`, `members_list_test.go`). En `server_test.go` solo se suma la fila de la ruta en `TestRutasConID_IDNoNumericoDevuelve400`; la ruta no recibe cuerpo, así que no suma filas a `bodyEndpoints`.

## Casos límite

- **Proyecto sin integrantes:** lista vacía y `[]` por HTTP.
- **Un solo integrante:** lista de un elemento.
- **Integrantes con los tres roles:** se devuelven todos, ordenados por ID, cada uno con su rol.
- **Dos proyectos con integrantes, incluso con el mismo nombre:** solo se devuelven los del proyecto consultado.
- **Integrante modificado con HU-16:** aparece con su nombre y su rol nuevos, en la misma posición (su ID no cambia).
- **Proyecto `FINALIZADO`:** se listan sus integrantes como en cualquier otro estado.
- **IDs de proyecto 0 y negativos:** `ErrProjectNotFound` (404).
- **Modificar la lista devuelta:** cambiar el nombre o el rol de un elemento de la lista no cambia el integrante guardado; una segunda consulta lo devuelve como estaba.

## Condiciones de error

| Situación | Dominio | HTTP |
|---|---|---|
| `{id}` no numérico | — | 400, `errInvalidID` |
| Proyecto inexistente, ID 0 o negativo | `project.ErrProjectNotFound` | 404 |
| Falla la lectura del proyecto o de los integrantes | el error del repositorio, con `nil` | 500, "error interno del servidor" |
| Sin proyecto actual en la interfaz | — | mensaje de `proyectoActual()` en `msg-integrantes` |

## Criterios de aceptación

- [ ] **CA-17.1:** se listan los integrantes de un proyecto existente, ordenados por ID, con su nombre y su rol.
- [ ] **CA-17.2:** un proyecto sin integrantes devuelve una lista vacía (`[]` por HTTP).
- [ ] **CA-17.3:** no aparecen integrantes de otros proyectos, aunque tengan el mismo nombre.
- [ ] **CA-17.4:** un proyecto inexistente (incluidos los IDs 0 y negativos) da `ErrProjectNotFound` (404), y un ID no numérico da 400.
- [ ] **CA-17.5:** la lista devuelta es una copia: modificarla no cambia los integrantes guardados.
- [ ] **CA-17.6:** desde la sección "Integrantes" de la interfaz se ven los integrantes del proyecto actual con su ID, su nombre y su rol.

## Comparación con los criterios iniciales del Issue #31

| Borrador | Definitivo | Cambio |
|---|---|---|
| CA-17.1 Se listan los integrantes del proyecto con nombre y rol. | CA-17.1 | Se precisa que el proyecto tiene que existir y el orden por ID (D3). |
| CA-17.2 Un proyecto sin integrantes devuelve una lista vacía. | CA-17.2 | Sin cambios de fondo; se agrega `[]` por HTTP. |
| CA-17.3 No aparecen integrantes de otros proyectos. | CA-17.3 | Se agrega el caso del mismo nombre en dos proyectos (RN-15.4 solo impide repetirlo dentro de un proyecto). |
| — | CA-17.4 | Nuevo: el proyecto se valida (decisión sobre entidades referenciadas). |
| — | CA-17.5 | Nuevo: copias por valor (`AGENTS.md`). |
| — | CA-17.6 | Nuevo: uso desde la interfaz (Definition of Done). |

Una vez aprobada la SDD, los criterios definitivos reemplazan a los del borrador en el Issue #31, con autorización.

## Comparación con el acta de Planning

- **Sprint 2:** 3 Story Points ("Como HU-06") y dependencia HU-15, integrada en `main`. Coincide con esta SDD: un listado por proyecto que valida el proyecto, como HU-06.
- **Sprint 3 y Sprint 4:** E4 (HU-15 a HU-17) queda **fuera de la entrega**, salvo capacidad sobrante. HU-15 y HU-16 ya se tomaron así y están en `main`; HU-17 se toma igual, con capacidad sobrante.
- **Tema 5 del Sprint 4:** el reporte de HU-38 lista los integrantes que registraron esfuerzo. Esta SDD no lo cambia.

No hay diferencias de alcance con las actas.

## Pruebas previstas

- **`internal/member/list_test.go` (unitarios, con `MemoryRepository`):**
  - proyecto con integrantes de los tres roles, ordenados por ID;
  - proyecto sin integrantes, lista vacía y no `nil`;
  - solo los integrantes del proyecto consultado, incluido el mismo nombre en otro proyecto;
  - un integrante modificado aparece con sus datos nuevos;
  - proyecto `FINALIZADO`;
  - proyecto inexistente, y los IDs 0 y negativos, sin leer los integrantes;
  - la copia no cambia lo guardado;
  - un error del repositorio de proyectos y uno del de integrantes, con repositorios que fallan a pedido.
- **`internal/member/list_features_test.go`:** el runner de godog de `features/HU-17-consultar-integrantes.feature`, un escenario por criterio CA-17.1 a CA-17.5 con caso normal, alternativo, límite y error.
- **`internal/httpapi/members_list_test.go`:**
  - 200 con la lista en el orden y con el formato exacto;
  - 200 con `[]`;
  - 404 con un proyecto inexistente, 0 y -1;
  - 500 con `brokenMemberRepository`, sin detalle;
  - la fila nueva de ID no numérico (400) en `server_test.go`.
- **`internal/httpapi/web_test.go`:** la interfaz tiene el botón `ver-integrantes` y la tabla `lista-integrantes`, y `app.js` llama a la ruta nueva con `GET` y usa `nombresDeRoles`, sin `innerHTML` en esa parte.

## Archivos que cambian

- **Nuevos:**
  - `internal/member/list.go`, `list_test.go` y `list_features_test.go`;
  - `internal/httpapi/members_list.go` y `members_list_test.go`;
  - `features/HU-17-consultar-integrantes.feature`.
- **Modificados:**
  - `internal/httpapi/server.go`: la ruta, junto a la de HU-15;
  - `internal/httpapi/server_test.go`: una fila en `TestRutasConID_IDNoNumericoDevuelve400`;
  - `internal/httpapi/web_test.go`;
  - `internal/web/index.html` y `internal/web/app.js`.
- **Documentación:**
  - `README.md`: la ruta y la fila de HU-17;
  - `docs/evidencias/HU-17-tdd.md`;
  - `docs/reportes/cobertura.md`;
  - `docs/ia/registro.md`;
  - `docs/documentacion-tecnica.md` (TEC-06): la operación y la ruta, después del merge.

## Límite conocido

`List` valida el proyecto antes de tomar el mutex y lee los integrantes después (RN-17.7). Si un integrante se registra entre los dos pasos, puede aparecer o no en la lista, según el momento. No cambia el resultado de fondo, porque los proyectos no se borran y la lista refleja el estado de ese momento, como en HU-14.
