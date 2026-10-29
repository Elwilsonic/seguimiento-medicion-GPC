# HU-14 · Consultar sprints anteriores

> Issue del repositorio: #28

Aprobada por: Peñalbé Hernán (Agile Enabler), 29/10/2026, con D1 aceptada y un ajuste (el mensaje del botón de cada fila avisa que ese sprint queda como sprint actual), que se aplicó.

## Historia de usuario

Como Agile Enabler, quiero consultar los sprints anteriores de un proyecto, para revisar su historial.

**Épica:** E3 · Gestión de Sprints

**Prioridad:** MEDIA (la del Issue #28)

**Story Points:** 3 (acta de Planning del Sprint 3)

## Objetivo

Listar los sprints de un proyecto existente, abiertos y cerrados, con sus datos y su estado, para que el Agile Enabler pueda revisar el historial del proyecto. Desde esa lista se consulta cualquier sprint cerrado con sus ítems completados, reutilizando la consulta de HU-12 sin cambiarla.

Hoy no hay forma de ver los sprints de un proyecto: la interfaz elige el sprint por su ID (Review del Sprint 2), y la única ruta que devuelve los sprints de un proyecto es la de HU-37, que los devuelve junto con sus métricas. HU-14 agrega el listado:

- en `sprint.Manager`, con la operación nueva `History`;
- en `GET /api/proyectos/{id}/sprints`;
- en la sección Sprints de la interfaz web.

## Alcance

- **Incluye:**
  - la operación `History(projectID)` del `Manager` de `internal/sprint`, que valida el proyecto y devuelve sus sprints;
  - la ruta `GET /api/proyectos/{id}/sprints`;
  - en la interfaz, el botón "Ver sprints del proyecto", la tabla de sprints y, en cada fila, el botón "Ver ítems completados".
- **Reutiliza sin cambios:**
  - `Repository.ListByProject` y `Manager.ListByProject` (HU-32): los sprints ordenados por ID, abiertos y cerrados, como copias;
  - `project.Manager.Get` (HU-03), para validar el proyecto, como hace `Create` (RN-10.4);
  - `Manager.ListCompletedItems` y `GET /api/sprints/{id}/completados` (HU-12), que ya permiten consultar un sprint CERRADO (RN-12.6, `TestListCompletedItems_SprintCerradoSePuedeConsultar`);
  - `sprintResponse` y `toSprintResponse` de `internal/httpapi/sprints.go`, y `mostrarSprint` de `app.js`.
- **No incluye:** filtros por estado o por fecha, paginación, otro orden que no sea por ID, métricas de cada sprint (ya las da HU-37) ni cambios en el esquema de la base.

## Contrato

### Entradas

- `projectID int`: el ID del proyecto cuyos sprints se consultan.

### Salidas esperadas

- `History(projectID int) ([]sprint.Sprint, error)`:
  - con un proyecto existente, devuelve una lista nueva con copias de sus sprints, abiertos y cerrados, ordenados por ID ascendente, y `nil` como error;
  - con un proyecto sin sprints, devuelve una lista vacía (no `nil`) y `nil` como error;
  - si hay un error, devuelve `nil` y el error.
- Cada sprint trae `ID`, `ProjectID`, `Name`, `StartDate`, `EndDate`, `Goal` y `Status` (`ABIERTO` o `CERRADO`).

### HTTP (TEC-04)

`GET /api/proyectos/{id}/sprints`

- **200:** un arreglo JSON con un objeto por sprint, en el mismo formato que la respuesta de `POST /api/proyectos/{id}/sprints` (`toSprintResponse`) y en el mismo orden que `History`. Un proyecto sin sprints devuelve `[]`, no `null`.

  ```json
  [
    {"id":1,"proyecto_id":1,"nombre":"Sprint 1","fecha_inicio":"2026-10-01","fecha_fin":"2026-10-08","objetivo":"Backlog inicial","estado":"CERRADO"},
    {"id":2,"proyecto_id":1,"nombre":"Sprint 2","fecha_inicio":"2026-10-09","fecha_fin":"2026-10-16","objetivo":"Métricas","estado":"ABIERTO"}
  ]
  ```

- **400:** el `{id}` no es un número entero (`errInvalidID`, por `pathID`).
- **404:** el proyecto no existe (`project.ErrProjectNotFound`), incluidos los IDs 0 y negativos.
- **500:** falla el almacenamiento.

Las respuestas de error usan el formato de siempre, `{"error": "..."}`, con el código que ya asigna `writeError`; no se agregan códigos nuevos. La ruta comparte el patrón con `POST /api/proyectos/{id}/sprints` y se distingue por el método.

### Interfaz web (TEC-04)

En la sección Sprints, debajo de "Sprint actual":

- **Botón "Ver sprints del proyecto"** (`ver-sprints`): llama a `GET /api/proyectos/{proyecto actual}/sprints` y llena la tabla `lista-sprints`.
- **Columnas de la tabla:** ID, Nombre, Inicio, Fin, Sprint goal, Estado y Acciones.
- **Mensajes:** al terminar, en `msg-sprints`, "N sprint(s) en el proyecto X.". Un proyecto sin sprints deja la tabla vacía y muestra "0 sprint(s) en el proyecto X.".
- **Botón "Ver ítems completados" de cada fila:** usa `mostrarSprint` para dejar ese sprint como sprint actual y después muestra sus ítems completados, con la misma consulta y la misma lista (`completados`) que el botón existente de HU-12. El mensaje, en `msg-sprints`, avisa el cambio de sprint actual: "Sprint N seleccionado: M ítem(s) completado(s).", porque "Cerrar sprint" y "Registrar esfuerzo" usan el sprint actual.
- **Errores:** sin proyecto actual, se muestra el mensaje de siempre de `proyectoActual()`. Con un proyecto inexistente, se muestra el error de la API. Los dos van en `msg-sprints`.
- **Seguridad:** los datos se escriben con `textContent`, nunca con `innerHTML`.

## Reglas de negocio

- **RN-14.1:** `History` lista todos los sprints del proyecto, ABIERTOS y CERRADOS. El estado de cada uno indica si ya terminó (decisión D1).
- **RN-14.2:** el proyecto tiene que existir. Se valida con `project.Manager.Get` antes de leer los sprints, y su error se propaga tal cual: `ErrProjectNotFound`, incluidos los IDs 0 y negativos (decisión D2).
- **RN-14.3:** los sprints se devuelven ordenados por ID ascendente, que es el orden en que se crearon, como en `ListByProject` (HU-32).
- **RN-14.4:** un proyecto existente sin sprints devuelve una lista vacía, no `nil` ni un error (CA-14.2).
- **RN-14.5:** `History` devuelve una lista nueva con copias por valor de los sprints. Modificar la lista o sus elementos no cambia los sprints guardados. `Sprint` no tiene slices ni mapas (RN-10.8), así que la copia de cada elemento no comparte estado.
- **RN-14.6:** la validación del proyecto se hace antes de tomar el mutex del `Manager` de sprints, igual que en `Create` (RN-10.4), para no tomar dos mutex a la vez. La lectura de los sprints se hace con `ListByProject`, que toma el mutex.
- **RN-14.7:** si el proyecto no existe, no se leen los sprints. Si falla la lectura del proyecto o la de los sprints, se devuelve ese error y `nil`, sin una lista parcial.
- **RN-14.8:** consultar un sprint de la lista con sus ítems completados es la consulta de HU-12 (`ListCompletedItems`), que acepta sprints ABIERTOS y CERRADOS. HU-14 no la cambia (CA-14.3).

## Decisión sobre entidades referenciadas

- **El proyecto se valida.** Sin esa validación, un proyecto inexistente y un proyecto sin sprints darían la misma lista vacía, y CA-14.2 no se podría distinguir de un error. Además, el resto de las consultas por proyecto (HU-06, HU-31, HU-32, HU-36, HU-38) informan `ErrProjectNotFound`.
- **Los sprints no se validan uno por uno:** salen del repositorio filtrados por proyecto, así que todos existen y pertenecen a él.

## Decisiones de diseño

- **D1. "Sprints anteriores" incluye el sprint abierto** (aceptada por el Agile Enabler: el estado ya muestra cuáles terminaron). El criterio CA-14.1 del borrador pide los sprints "con sus datos y estado". Mostrar el estado solo sirve si la lista mezcla abiertos y cerrados, y el sprint abierto también es parte del historial del proyecto. Si se prefiere solo los CERRADOS, el cambio es un filtro en `History` y un escenario más.
- **D2. Operación nueva `History`, sin cambiar `ListByProject`.** `ListByProject` no valida el proyecto, y la usan seis llamadores:
  - `internal/dashboard/summary.go`;
  - `internal/metrics`, en `hours.go`, `sprints.go`, `story_points.go` y `summary.go`;
  - `internal/report/project.go`.

  Todos validan el proyecto antes con el backlog o con `project.Manager.Get`. Agregarle la validación leería el proyecto dos veces y cambiaría las lecturas que cuentan los tests espía de HU-35 y HU-37.
- **D3. Formato HTTP:** la respuesta reutiliza `toSprintResponse`, para que cada sprint tenga el mismo formato en todas las rutas.

## Restricciones

- Solo la librería estándar de Go y `net/http`. No se agregan librerías.
- El código de negocio accede a los sprints solo a través de `Repository` (memoria o PostgreSQL). No hace falta una operación nueva en el repositorio, porque `ListByProject` ya existe en los dos y la cubren los tests de contrato.
- No cambia el esquema de la base ni los contratos de las rutas existentes.
- La interfaz no usa frameworks ni `innerHTML`.

## Casos límite

- **Proyecto sin sprints:** lista vacía y `[]` por HTTP.
- **Un solo sprint, abierto:** lista de un elemento con estado `ABIERTO`.
- **Sprints cerrados y uno abierto:** se devuelven todos, ordenados por ID, cada uno con su estado.
- **Dos proyectos con sprints:** solo se devuelven los del proyecto consultado.
- **Sprint cerrado sin ítems completados:** aparece en la lista; su consulta de HU-12 devuelve `[]`.
- **IDs de proyecto 0 y negativos:** `ErrProjectNotFound` (404).
- **Modificar la lista devuelta:** cambiar el nombre o el estado de un elemento de la lista no cambia el sprint guardado; una segunda consulta lo devuelve como estaba.

## Condiciones de error

| Situación | Dominio | HTTP |
|---|---|---|
| `{id}` no numérico | — | 400, `errInvalidID` |
| Proyecto inexistente, ID 0 o negativo | `project.ErrProjectNotFound` | 404 |
| Falla la lectura del proyecto o de los sprints | el error del repositorio, con `nil` | 500 |
| Sin proyecto actual en la interfaz | — | mensaje de `proyectoActual()` en `msg-sprints` |

## Criterios de aceptación

- [ ] **CA-14.1:** se listan los sprints de un proyecto existente, abiertos y cerrados, ordenados por ID, con su nombre, fechas, sprint goal y estado.
- [ ] **CA-14.2:** un proyecto sin sprints devuelve una lista vacía (`[]` por HTTP).
- [ ] **CA-14.3:** un sprint cerrado de la lista se puede consultar con sus ítems completados (consulta de HU-12).
- [ ] **CA-14.4:** un proyecto inexistente (incluidos los IDs 0 y negativos) da `ErrProjectNotFound` (404), y un ID no numérico da 400.
- [ ] **CA-14.5:** la lista devuelta es una copia: modificarla no cambia los sprints guardados.
- [ ] **CA-14.6:** desde la sección Sprints de la interfaz se ven los sprints del proyecto actual y, desde cada fila, sus ítems completados; el mensaje avisa que ese sprint queda como sprint actual.

## Comparación con los criterios iniciales del Issue #28

| Borrador | Definitivo | Cambio |
|---|---|---|
| CA-14.1 Se listan los sprints de un proyecto con sus datos y estado. | CA-14.1 | Se precisa que incluye abiertos y cerrados (D1), el orden por ID y qué datos son. |
| CA-14.2 Un proyecto sin sprints devuelve una lista vacía. | CA-14.2 | Sin cambios de fondo; se agrega `[]` por HTTP. |
| CA-14.3 Se puede consultar un sprint cerrado con sus ítems completados. | CA-14.3 | Se aclara que es la consulta de HU-12, que ya existe. |
| — | CA-14.4 | Nuevo: el proyecto se valida (decisión sobre entidades referenciadas). |
| — | CA-14.5 | Nuevo: copias por valor (`AGENTS.md`). |
| — | CA-14.6 | Nuevo: uso desde la interfaz (Definition of Done). |

Una vez aprobada la SDD, los criterios definitivos reemplazan a los del borrador en el Issue #28, con autorización.

## Comparación con el acta de Planning

- **Sprint 2:** dependencias HU-10 y HU-13, las dos integradas en `main`.
- **Sprint 3:** 3 Story Points y "Listado de los sprints de un proyecto, que hoy no existe". Coincide con el alcance de esta SDD.
- **Sprint 4:** sigue en "si hay tiempo", después de HU-41 y HU-24, que ya están integradas.
- **Review del Sprint 2:** "La interfaz elige el proyecto y el sprint por su ID porque no hay listados […]; HU-14 cubre los sprints". La tabla con el botón por fila cubre esa parte.

No hay diferencias de alcance con las actas.

## Pruebas previstas

- **`internal/sprint/history_test.go` (unitarios, con `MemoryRepository`):**
  - proyecto con sprints abiertos y cerrados, ordenados por ID;
  - proyecto sin sprints, lista vacía y no `nil`;
  - solo los sprints del proyecto consultado;
  - proyecto inexistente, y los IDs 0 y negativos;
  - la copia no cambia lo guardado;
  - un error del repositorio de proyectos y uno del de sprints, con un repositorio que falla a pedido; el segundo verifica también que con proyecto inexistente no se leen los sprints.
- **`internal/sprint/history_features_test.go`:** el runner de godog de `features/HU-14-sprints-anteriores.feature`, incluido el sprint cerrado consultado con sus ítems completados (CA-14.3).
- **`internal/httpapi/sprints_test.go`:**
  - 200 con la lista en el orden y con el formato exacto (`expectRawJSON`);
  - 200 con `[]`;
  - 404 con un proyecto inexistente;
  - 400 con un ID no numérico;
  - 500 con un error de almacenamiento;
  - una fila nueva en la tabla de rutas de `server_test.go`.
- **`internal/httpapi/web_test.go`:** la interfaz tiene el botón `ver-sprints` y la tabla `lista-sprints`, y `app.js` llama a la ruta nueva, usa `mostrarSprint` y el mensaje "seleccionado" en el botón de cada fila, y no usa `innerHTML` en esa parte.

## Archivos que cambian

- **Nuevos:**
  - `internal/sprint/history.go`, `history_test.go` y `history_features_test.go`;
  - `features/HU-14-sprints-anteriores.feature`.
- **Modificados:**
  - `internal/httpapi/sprints.go`: el handler `listProjectSprints`;
  - `internal/httpapi/server.go`: la ruta;
  - los tests de `internal/httpapi`;
  - `internal/web/index.html` y `internal/web/app.js`.
- **Documentación:**
  - `README.md`: la ruta y la fila de HU-14;
  - `docs/evidencias/HU-14-tdd.md`;
  - `docs/reportes/cobertura.md`;
  - `docs/ia/registro.md`.

## Límite conocido

`History` valida el proyecto y lee los sprints en dos pasos, sin un mutex común. Si un sprint se crea entre los dos pasos, puede aparecer o no en la lista, según el momento. No cambia el resultado de fondo, porque los proyectos no se borran y la lista refleja el estado de ese momento.
