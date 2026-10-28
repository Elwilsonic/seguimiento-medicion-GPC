# HU-31 · Consultar los defectos de un proyecto o de un sprint

> Issue del repositorio: #49

Aprobada por: Peñalbé Hernán (Agile Enabler), 28/10/2026, con un ajuste en D3 (un filtro con solo espacios se trata como vacío y no filtra), que se aplicó.

## Historia de usuario

Como Agile Enabler, quiero consultar los defectos de un proyecto o de un sprint, para evaluar su calidad.

**Épica:** E8 · Gestión de defectos

**Prioridad:** BAJA (la del Issue #49)

## Objetivo

Obtener los defectos de un proyecto o de un sprint, con la posibilidad de filtrarlos por estado (`ABIERTO`, `EN_CORRECCION` o `RESUELTO`, de HU-30), para que el Agile Enabler evalúe la calidad. Es una operación de solo lectura: no crea, modifica ni elimina defectos.

La historia agrega a `internal/defect` dos consultas, `ListByProject` y `ListBySprint`, y una lectura nueva en la interfaz `Repository`, implementada en memoria y en PostgreSQL, sin cambios de esquema. `ListByProject` es la operación que la SDD de HU-47 (sección "Dependencias") espera para calcular M9 y M10: su nombre y su firma coinciden con los de esa SDD.

Se expone en `GET /api/proyectos/{id}/defectos` y `GET /api/sprints/{id}/defectos`, con el filtro opcional `?estado=`, y en un formulario "Consultar defectos" de la sección Defectos de la interfaz web (tema 5 del Planning del Sprint 3), con el modelo de HU-27, como pide el acta del Sprint 4.

## Alcance

- **Incluye:**
  - las consultas `Manager.ListByProject` y `Manager.ListBySprint`;
  - el filtro por estado (`ParseStatus` y `FilterByStatus`), que reutiliza la normalización y `ErrInvalidStatus` de HU-30;
  - la lectura `List` en la interfaz `Repository` (memoria y PostgreSQL);
  - los dos endpoints HTTP y el formulario web;
  - tests unitarios, BDD, de contrato, de PostgreSQL, de API, de interfaz y de concurrencia.
- **No incluye:**
  - registrar defectos (HU-29) ni cambiar su estado (HU-30);
  - editar o eliminar defectos;
  - calcular M9 o M10 (HU-47);
  - reportes (HU-38 y HU-39) y exportación (HU-40);
  - consultar un defecto por su ID, por historia o de todo el sistema;
  - filtrar por severidad, por historia o por sprint de resolución, paginar u ordenar de otra forma;
  - cambios de esquema.

## Contrato

### Entradas

| Consulta | Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|---|
| Por proyecto | ProjectID | entero | Sí | ID del proyecto cuyos defectos se listan |
| Por sprint | SprintID | entero | Sí | ID del sprint en el que se detectaron los defectos |
| Filtro (ambas) | Estado | texto | No | `ABIERTO`, `EN_CORRECCION` o `RESUELTO`, sin distinguir mayúsculas ni contar los espacios de los extremos; sin filtro (ausente, vacío o con solo espacios) se devuelven todos |

Cada consulta usa exactamente uno de los dos IDs; no hay consulta combinada ni listado global.

Firmas en `internal/defect`:

- `func (m *Manager) ListByProject(projectID int) ([]Defect, error)` — la firma exacta de la SDD de HU-47, sin filtro.
- `func (m *Manager) ListBySprint(sprintID int) ([]Defect, error)` — sin filtro.
- `func ParseStatus(status string) (Status, error)` — valida y normaliza un estado; reutiliza `normalizeStatus` de HU-30 y devuelve `ErrInvalidStatus` para cualquier valor que no sea uno de los tres estados, incluido el vacío (igual que `ChangeStatus`). No decide si hay filtro: eso lo decide el handler (D3).
- `func FilterByStatus(defects []Defect, status Status) []Defect` — devuelve una lista nueva con los defectos en ese estado, en el mismo orden.

### Salidas esperadas

- **Si la entidad existe:** una lista nueva, ordenada por ID ascendente, con copias de los defectos:
  - `ListByProject`: todos los defectos del proyecto, en cualquier estado (RN-31.2);
  - `ListBySprint`: todos los defectos detectados en el sprint, en cualquier estado (RN-31.3).

  Cada defecto tiene todos los campos de HU-29 y HU-30: ID, descripción, severidad, estado, historia, sprint de detección y sprint de resolución.
- **Si la entidad existe y no tiene defectos** (o ninguno con el estado pedido): una lista de longitud 0, no `nil`, sin error (RN-31.6).
- **Si hay un error:** lista `nil` y el error correspondiente (ver "Condiciones de error").
- **`FilterByStatus`:** una lista nueva de longitud 0 o más, nunca `nil`; no modifica la lista recibida.

### HTTP (TEC-04, tema 5 del Planning del Sprint 3)

- `GET /api/proyectos/{id}/defectos` para consultar por proyecto.
- `GET /api/sprints/{id}/defectos` para consultar por sprint.
- Las dos aceptan el parámetro opcional `?estado=ABIERTO|EN_CORRECCION|RESUELTO`. El handler recorta el valor con `strings.TrimSpace`; si queda vacío (sin el parámetro, `?estado=` o solo espacios), no se filtra (D3).

- **Respuesta exitosa:** `200 OK` con la lista y la cantidad, con el modelo de la respuesta de HU-27 (`esfuerzos` y `total_horas`). Cada defecto tiene la misma forma que la respuesta de HU-29 y HU-30 (`toDefectResponse`):

  ```json
  {
    "defectos": [
      {"id": 1, "descripcion": "El total no descuenta el descuento", "severidad": "ALTA", "estado": "RESUELTO", "item_id": 4, "sprint_deteccion_id": 3, "sprint_resolucion_id": 5}
    ],
    "cantidad": 1
  }
  ```

  - `defectos` es una lista; si no hay defectos se devuelve `[]`, no `null`.
  - `cantidad` es la cantidad de defectos de la lista (después del filtro), entera.

- **Errores:**

| Caso | Código | Mensaje |
|---|---|---|
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| `estado` que, recortado, no queda vacío y no es ABIERTO, EN_CORRECCION ni RESUELTO (por ejemplo `CERRADO`) | 400 | El de `ErrInvalidStatus` |
| Proyecto inexistente | 404 | El de `project.ErrProjectNotFound` |
| Sprint inexistente | 404 | El de `sprint.ErrSprintNotFound` |
| Falla de un repositorio | 500 | El de `errInternal`, sin detalle |

No hay 409 ni 413: las consultas no reciben cuerpo ni cambian estado. `statusFor` ya mapea todos estos errores, así que no se agrega ninguno. El orden de los errores del handler es: ID de la ruta, estado del filtro y `Manager`.

### Interfaz web (TEC-04, tema 5)

En la sección **Defectos** de `internal/web`, debajo de "Cambiar estado de defecto", un formulario **"Consultar defectos"**:

- **Ver por:** `select` con `Proyecto actual` (valor `proyecto`, por defecto) y `Sprint actual` (valor `sprint`). Usa los IDs de "Proyecto actual" y "Sprint actual" del encabezado (`proyectoActual()` y `sprintActual()`), con sus mensajes si están vacíos.
- **Estado:** `select` con `Todos` (valor vacío, por defecto) y los estados de `estadosDefecto`, la lista de HU-30 que un test ya compara con `defect.StatusValues()`.
- **Botón:** Consultar.
- **Tabla:** columnas ID, Descripción, Severidad, Estado, Ítem, Sprint de detección y Sprint de resolución, con `<tbody id="lista-defectos">` (el `id="defectos"` ya es el de la sección). Un sprint de resolución 0 se muestra como `—`.
- **Cantidad:** `<p id="cantidad-defectos">` muestra `Cantidad: N defecto(s)`.
- **Sin defectos:** una fila `Sin defectos.` con `colSpan="7"` y `Cantidad: 0 defecto(s)`.
- **Mensaje:** `<p id="msg-consultar-defectos" role="alert" hidden>`, con el mismo mecanismo de `accion` y `mostrar` que los demás formularios, para el resultado o el error de la API.
- Llama a `GET /proyectos/{id}/defectos` o `GET /sprints/{id}/defectos` y agrega `?estado=` solo si se eligió un estado. Usa `textContent` para todos los datos.

## Reglas de negocio

1. **RN-31.1 Una consulta, un alcance.** Cada consulta usa exactamente un ID: el de un proyecto o el de un sprint. No se listan todos los defectos del sistema.
2. **RN-31.2 Defectos de un proyecto.** Son todos los defectos cuya historia (`ItemID`) pertenece al proyecto, en cualquier estado. `Defect` no guarda el proyecto: se obtiene de la historia, como en HU-29, HU-30 y RN-47.3.
3. **RN-31.3 Defectos de un sprint.** Son los defectos **detectados** en el sprint (`DetectionSprintID` igual al sprint), en cualquier estado, con la misma definición que M9 en `docs/metricas.md`. Un defecto detectado en un sprint y resuelto en otro aparece solo en la consulta del sprint de detección (decisión D2).
4. **RN-31.4 Filtro por estado.** El filtro es opcional. Primero se recorta con `strings.TrimSpace`: si queda vacío (sin parámetro, `?estado=` o solo espacios), no se filtra. Si no queda vacío, se normaliza con `strings.ToUpper`, como en HU-30, y solo quedan los defectos con ese estado; cualquier otro valor que no sea `ABIERTO`, `EN_CORRECCION` ni `RESUELTO` devuelve `ErrInvalidStatus`. El filtro se aplica sobre el resultado de `ListByProject` o `ListBySprint`, que no cambian de firma (decisión D3).
5. **RN-31.5 Entidad existente.** Si el proyecto o el sprint no existe (incluidos 0 y negativos), se devuelve `project.ErrProjectNotFound` o `sprint.ErrSprintNotFound`, y no una lista vacía.
6. **RN-31.6 Lista vacía válida.** Si la entidad existe y no tiene defectos, o ninguno con el estado pedido, se devuelve una lista de longitud 0 (no `nil`), sin error. En JSON, `[]` y no `null`.
7. **RN-31.7 Aislamiento.** La consulta de un proyecto no incluye defectos de otro proyecto, y la de un sprint no incluye defectos detectados en otro sprint.
8. **RN-31.8 Orden por ID.** Las listas se ordenan por ID ascendente, que es el orden de registro. Ordena el repositorio: el de memoria recorre en orden de guardado y PostgreSQL usa `ORDER BY id`. El filtro conserva el orden.
9. **RN-31.9 Sprints abiertos y cerrados.** Se puede consultar un sprint `ABIERTO` o `CERRADO`, y un proyecto en cualquier estado; el estado no restringe la lectura.
10. **RN-31.10 Copias por valor.** Las listas devueltas son nuevas y cada defecto es una copia. `Defect` no tiene slices ni mapas, así que modificar la lista o un defecto devuelto no altera lo guardado.
11. **RN-31.11 Solo lectura.** Consultar no modifica defectos, historias, sprints ni proyectos.
12. **RN-31.12 Concurrencia.** Las consultas toman el mutex del `Manager` de defectos durante la validación y la lectura, como `Register` (HU-29) y `ChangeStatus` (HU-30). Así ven un estado coherente frente a registros y cambios de estado simultáneos.
13. **RN-31.13 API e interfaz.** Las dos consultas se exponen en la API y en el formulario "Consultar defectos", con un caso exitoso y un test por código de error (tema 5 del Planning del Sprint 3).
14. **RN-31.14 Base de HU-47.** `ListByProject` devuelve todos los defectos del proyecto, en cualquier estado, sin filtrar, con `project.ErrProjectNotFound` para un proyecto inexistente: es la operación que usa HU-47 para M9 y M10.

## Decisión sobre entidades referenciadas

HU-31 referencia por ID un proyecto o un sprint, y **sí valida su existencia**, como HU-06, HU-27 y HU-47. Sin la validación, un ID mal escrito devolvería una lista vacía y ocultaría el error: "la entidad no existe" es un error y "existe pero no tiene defectos" es un resultado válido.

- **Proyecto** (consulta por proyecto): se valida con `backlog.Manager.ListItems(projectID)`, que ya llama a `project.Manager.Get` y devuelve `project.ErrProjectNotFound` (HU-06). La misma llamada devuelve las historias del proyecto, que se necesitan para filtrar los defectos (decisión D4). No se agrega ninguna operación al backlog.
- **Sprint** (consulta por sprint): se valida con `sprint.Manager.Get`, que devuelve `sprint.ErrSprintNotFound` (como HU-27). No se exige que esté abierto.
- **Historias y sprints de cada defecto:** **no se validan** uno por uno. HU-29 ya los validó al registrar el defecto, el sistema no elimina historias ni sprints, y una historia no cambia de proyecto.
- **Proyecto del sprint:** **no se consulta.** El sprint ya determina el conjunto (`DetectionSprintID`), como en HU-27.

## Decisiones de diseño (confirmadas por Peñalbé Hernán, Agile Enabler, el 28/10/2026)

- **D1 · Consulta por proyecto.** `Manager.ListByProject(projectID int) ([]Defect, error)`, con la firma exacta de la SDD de HU-47: copias de todos los defectos del proyecto, en cualquier estado, ordenados por ID; proyecto inexistente → `project.ErrProjectNotFound`. El filtro por estado no va en esta firma, porque HU-47 la necesita sin filtro.
- **D2 · Consulta por sprint.** `Manager.ListBySprint(sprintID int) ([]Defect, error)`: los defectos **detectados** en el sprint (`DetectionSprintID`), ordenados por ID, con la misma definición que M9 en `docs/metricas.md`; sprint inexistente → `sprint.ErrSprintNotFound`.
  - *Alternativa:* incluir también los defectos resueltos en el sprint (`RESUELTO` con `ResolutionSprintID` igual al sprint), sin repetirlos. Se propone no incluirlos: la consulta sería la unión de M9 y M10 y no coincidiría con ninguna de las dos, y un defecto aparecería en dos sprints. Si se elige la alternativa, cambian RN-31.3, CA-31.2 y sus escenarios.
  - HU-47 no usa `ListBySprint`: calcula M9 y M10 de un sprint con `ListByProject` y filtra (sección "Decisión sobre entidades referenciadas" de su SDD), así que esta decisión no la afecta.
- **D3 · Filtro por estado.** Lo aplican dos funciones del paquete, sin cambiar las firmas de D1 y D2:
  - `ParseStatus(status string) (Status, error)`: exporta la validación de `normalizeStatus` (HU-30), con `ErrInvalidStatus`. Un valor vacío o con solo espacios también da `ErrInvalidStatus`, como en `ChangeStatus`: `ParseStatus` valida un estado, no decide si hay filtro.
  - `FilterByStatus(defects []Defect, status Status) []Defect`: lista nueva, mismo orden, nunca `nil`.

  Se separa la validación del filtro para que el handler valide `?estado=` **antes** de consultar, como HU-30 valida el estado antes de leer el repositorio: un estado inválido sobre un proyecto inexistente da 400 y no 404. **Quién decide "vacío = sin filtro": el handler.** Recorta `?estado=` con `strings.TrimSpace`; si queda vacío (sin el parámetro, `?estado=` o solo espacios), no llama a `ParseStatus` ni a `FilterByStatus` y devuelve todos, como `?orden=` de HU-09, que no da 400. Si no queda vacío, llama a `ParseStatus`, y cualquier valor que no sea uno de los tres estados (por ejemplo `CERRADO`) da 400 (ajuste de Hernán del 28/10/2026). *Alternativa:* una sola función `FilterByStatus(defects, status string) ([]Defect, error)`, más simple pero que obliga a consultar antes de validar el estado (404 antes que 400).
- **D4 · Proyecto del defecto.** `Defect` no guarda el proyecto, solo `ItemID`. `ListByProject` obtiene las historias del proyecto con `backlog.Manager.ListItems`, que ya valida el proyecto, y se queda con los defectos cuyo `ItemID` está entre esas historias. No hay cambios de esquema ni operaciones nuevas en el backlog.
- **D5 · Lectura del repositorio.** Una sola lectura nueva en `defect.Repository`: `List() ([]Defect, error)`, con todos los defectos ordenados por ID, en memoria (`MemoryRepository`, bajo su mutex) y PostgreSQL (`SELECT ... FROM defects ORDER BY id`). El `Manager` filtra por historias (D4) o por sprint de detección (D2).
  - *Por qué `List` y no `ListByItemIDs` o `ListByDetectionSprint`:* una sola lectura sirve a las dos consultas, con el mismo comportamiento en memoria y PostgreSQL y un único test de contrato. `ListByItemIDs` necesitaría `= ANY($1)` en PostgreSQL y una segunda lectura para el sprint.
  - *Costo:* cada consulta lee todos los defectos del sistema. Con el volumen de la aplicación (en memoria o una base local) no es un problema; si lo fuera, se cambia la lectura sin tocar las firmas del `Manager`.
- **D6 · Lista vacía.** Sin defectos, o ninguno con el estado pedido, se devuelve una lista de longitud 0, no `nil`; en JSON, `[]` y no `null` (`make([]defectResponse, 0, len(...))`, como HU-06 y HU-27).
- **D7 · HTTP.** `GET /api/proyectos/{id}/defectos` y `GET /api/sprints/{id}/defectos`, con `?estado=` opcional. Respuesta `{"defectos": [...], "cantidad": N}`, con el modelo de HU-27 (`esfuerzos` y `total_horas`) y cada defecto con la forma de `toDefectResponse` (HU-29 y HU-30). Errores: 400 por ID o estado inválido, 404 por proyecto o sprint inexistente, 500 por falla del repositorio.
- **D8 · Interfaz.** En la sección Defectos, el formulario "Consultar defectos" para el proyecto actual o el sprint actual, con un `select` de estado opcional (con `estadosDefecto`, que ya se compara con `StatusValues()`), una tabla con los defectos y la cantidad, o "Sin defectos.".
- **D9 · Concurrencia.** `ListByProject` y `ListBySprint` toman el mutex del `Manager` de defectos durante la validación y la lectura. Orden de locks:
  - `ListByProject`: defect → project → backlog (`ListItems` consulta el proyecto antes de tomar el mutex del backlog, HU-06);
  - `ListBySprint`: defect → sprint;
  - es compatible con el orden defect → sprint → backlog de HU-29 y HU-30: ningún paquete llama a `defect`, así que no hay ciclos.

  Un test con `go test -race` ejecuta las dos consultas a la vez que `Register` y `ChangeStatus`.
- **D10 · RED y GREEN.** En el RED solo se agregan firmas vacías: `ListByProject`, `ListBySprint`, `ParseStatus`, `FilterByStatus` y `List` en los repositorios en memoria y PostgreSQL, para que todo compile y los tests fallen por aserción. Las rutas, los handlers y el formulario van en el GREEN (ajuste de Hernán en HU-18): sin las rutas, los tests de API fallan por aserción (404 o 405 en lugar de 200).

## Restricciones

- **Organización del código.**
  - `internal/defect/list.go` (nuevo): `ListByProject`, `ListBySprint`, `ParseStatus` y `FilterByStatus`.
  - `internal/defect/repository.go`: suma `List() ([]Defect, error)`; se actualiza el comentario de la interfaz, que hoy dice que los listados llegan con HU-31.
  - `internal/defect/memory_repository.go` y `postgres_repository.go`: implementan `List`.
  - `internal/defect/status.go`: `normalizeStatus` no cambia; `ParseStatus` la reutiliza.
  - No se tocan `manager.go`, `register.go` ni `defect.go`: no hay errores nuevos.
  - `internal/httpapi/defects.go`: handlers `listDefectsByProject` y `listDefectsBySprint`, la respuesta `defectListResponse` y la lectura de `?estado=`.
  - `internal/httpapi/server.go`: las dos rutas GET. `statusFor` no cambia.
  - `internal/web/index.html` (formulario, tabla, cantidad y mensaje) y `internal/web/app.js` (llamadas y armado de la tabla).
  - `README.md`: las dos rutas en la tabla de la API, un ejemplo `curl`, la línea "Defectos" del uso de la interfaz y la fila de HU-31 en "Historias por sprint".
  - `docs/evidencias/HU-31-tdd.md`, `docs/reportes/cobertura.md` y `docs/ia/registro.md` (paso 7).
- **Consecuencia de ampliar la interfaz.** Los repositorios falsos que implementan `defect.Repository` necesitan `List`, pero **los archivos de tests de HU-29 y HU-30 no se modifican**: en Go los métodos de un tipo se pueden declarar en cualquier archivo del mismo paquete, así que van en los archivos de tests nuevos de HU-31.
  - `testDefectRepository` y `failingDefectRepository` (declarados en `internal/defect/register_test.go`): sus métodos `List` se declaran en `internal/defect/list_test.go`. No se tocan `register_test.go` ni `status_test.go`.
  - `brokenDefectRepository` (declarado en `internal/httpapi/defects_test.go`): su método `List` se declara en `internal/httpapi/defect_lists_test.go`, nuevo. No se toca `defects_test.go`.
  - `countingDefectRepository` (embebe `Repository`) y `recordingDefectRepository` (embebe `*testDefectRepository`) heredan `List` y no necesitan cambios.

  No cambia el comportamiento de `Save`, `GetByID` ni `Update`.
- **Sin cambios de esquema.** Se reutiliza la tabla `defects` de HU-29; no hace falta recrear la base.
- **Sin errores nuevos.** Se reutilizan `ErrInvalidStatus` (HU-30), `project.ErrProjectNotFound`, `sprint.ErrSprintNotFound` y los errores del repositorio.
- **Concurrencia.** Mutex del `Manager` (RN-31.12 y D9), verificado con `go test -race` en el CI.
- **RED y GREEN.** Según D10. Antes del commit RED se corre `go vet ./...`.
- **Tests.**
  - Unitarios en `internal/defect/list_test.go`, que también declara `List` de `testDefectRepository` y `failingDefectRepository`.
  - godog en `internal/defect/list_features_test.go`, con el runner `TestFeatures_HU31ConsultarDefectos`, sobre `features/HU-31-consultar-defectos.feature`, con el CA en el nombre de cada escenario.
  - Contrato de `List` en `internal/defect/repository_contract_test.go` (memoria y PostgreSQL): lista vacía no `nil`, orden por ID, todos los estados y copias. Específicos de PostgreSQL en `postgres_repository_test.go`, que se omiten sin `TEST_DATABASE_URL`.
  - API en `internal/httpapi/defect_lists_test.go` (nuevo, como `effort_lists_test.go` de HU-27), que también declara `List` de `brokenDefectRepository`: 200 con y sin filtro (incluidos `?estado=` y un filtro con solo espacios, que no filtran), lista vacía `[]`, 400 por estado inválido, 404 por proyecto y sprint inexistentes, 500 por repositorio roto; las dos rutas nuevas se suman a `TestRutasConID_IDNoNumericoDevuelve400` en `server_test.go`.
  - Interfaz en `internal/httpapi/web_test.go`: formulario, selectores, tabla, cantidad, caso vacío y llamadas de `app.js`.
- **Tests de PostgreSQL.** Los corre y verifica otro integrante antes del merge si el entorno local no tiene Docker.
- **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Proyecto con defectos en los tres estados | Se listan todos, ordenados por ID |
| Proyecto con defectos en varias historias | Se listan los de todas sus historias |
| Proyecto existente sin historias | Lista vacía, sin error |
| Proyecto con historias pero sin defectos | Lista vacía, sin error |
| Defectos de otro proyecto | No aparecen |
| Sprint con defectos detectados en los tres estados | Se listan todos, ordenados por ID |
| Defecto detectado en S1 y resuelto en S2 | Aparece en la consulta de S1, no en la de S2 (D2); aparece una vez en la del proyecto |
| Sprint existente sin defectos detectados | Lista vacía, sin error |
| Sprint `CERRADO` con defectos | Se listan igual |
| Filtro `RESUELTO` | Solo los defectos `RESUELTO`, en el mismo orden |
| Filtro `en_correccion ` o ` Abierto` | Se normaliza |
| Filtro con un estado sin defectos | Lista vacía, sin error |
| Filtro vacío (sin parámetro o `?estado=`) | No se filtra |
| Filtro con solo espacios | No se filtra (el handler lo recorta y queda vacío) |
| Filtro `CERRADO` u otro valor | `ErrInvalidStatus` (400) |
| `ParseStatus` con un valor vacío o con solo espacios | `ErrInvalidStatus` (el handler no la llama en ese caso) |
| Filtro inválido sobre un proyecto o sprint inexistente | `ErrInvalidStatus` (400): el filtro se valida primero |
| Proyecto inexistente, 0 o negativo | `project.ErrProjectNotFound` (404); no se devuelve lista |
| Sprint inexistente, 0 o negativo | `sprint.ErrSprintNotFound` (404); no se devuelve lista |
| Modificar la lista o un defecto devuelto | Lo guardado no cambia; volver a consultar devuelve lo mismo |
| `FilterByStatus` sobre una lista | La lista recibida no cambia |
| Consultar dos veces | Mismo resultado; no se modifica nada |
| Consultas simultáneas con `Register` y `ChangeStatus` | `go test -race` no reporta condiciones de carrera |
| El repositorio de defectos falla al listar | Se devuelve ese error y ninguna lista (500 sin detalle) |
| El backlog falla al listar las historias del proyecto | Se devuelve ese error y ninguna lista (500 sin detalle) |
| HTTP: `{id}` no numérico | 400 con el mensaje de `errInvalidID` |
| HTTP: sin defectos | 200 con `{"defectos":[],"cantidad":0}` |
| Interfaz: sin proyecto o sprint actual | El mensaje de `proyectoActual()` o `sprintActual()` |
| Interfaz: sprint de resolución 0 | Se muestra `—` |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Estado del filtro que no es ABIERTO, EN_CORRECCION ni RESUELTO (RN-31.4) | `ErrInvalidStatus` | No se consulta nada |
| Proyecto inexistente (RN-31.5) | `project.ErrProjectNotFound` | No se devuelve lista |
| Sprint inexistente (RN-31.5) | `sprint.ErrSprintNotFound` | No se devuelve lista |
| El repositorio de defectos o el backlog falla al leer | El error del repositorio | No se devuelve lista |

**Orden de validación:**

1. En la capa HTTP, sin tocar repositorios: ID de la ruta (`errInvalidID`) y el estado del filtro, recortado con `strings.TrimSpace`; si no queda vacío, `ParseStatus` (`ErrInvalidStatus`).
2. Con el mutex del `Manager` tomado: existencia del proyecto (con `ListItems`) o del sprint (con `Get`); si no existe, no se lee el repositorio de defectos.
3. Lectura con `List` y selección por historias o por sprint de detección.
4. En la capa HTTP: `FilterByStatus`, si el estado recortado no quedó vacío.

Se informa el primer error encontrado.

## Criterios de aceptación

- **CA-31.1** Se listan todos los defectos de un proyecto (los de todas sus historias) o los detectados en un sprint, en cualquier estado, con ID, descripción, severidad, estado, historia, sprint de detección y sprint de resolución, ordenados por ID ascendente.
- **CA-31.2** La consulta de un proyecto no incluye defectos de otro proyecto, y la de un sprint no incluye defectos detectados en otro sprint, aunque se hayan resuelto en él. Se pueden consultar sprints abiertos y cerrados.
- **CA-31.3** Se puede filtrar por estado (`ABIERTO`, `EN_CORRECCION` o `RESUELTO`, sin distinguir mayúsculas ni espacios de los extremos); sin filtro (ausente, vacío o con solo espacios) se devuelven todos. Cualquier otro valor se rechaza con `ErrInvalidStatus` antes de consultar.
- **CA-31.4** Un proyecto o un sprint existente sin defectos, o sin defectos con el estado pedido, devuelve una lista vacía (longitud 0, no `nil`; `[]` en JSON), sin error.
- **CA-31.5** Consultar un proyecto o un sprint inexistente (incluidos 0 y negativos) devuelve `project.ErrProjectNotFound` o `sprint.ErrSprintNotFound`, y no una lista vacía.
- **CA-31.6** Modificar la lista o un defecto devuelto no altera los defectos guardados, y consultar no modifica nada.
- **CA-31.7** Las consultas simultáneas con registros y cambios de estado no producen condiciones de carrera (`go test -race`).
- **CA-31.8** Si un repositorio falla al leer, se devuelve ese error y ninguna lista.
- **CA-31.9** Las consultas se usan desde la API (`GET /api/proyectos/{id}/defectos` y `GET /api/sprints/{id}/defectos`, con `?estado=` opcional) y desde el formulario "Consultar defectos" de la interfaz web. La API responde 200 en el caso exitoso y 400, 404 o 500 según la tabla HTTP; la interfaz muestra la tabla, la cantidad, el caso vacío y el mensaje de error. Se cubre con tests de `internal/httpapi`, no con escenarios godog.

## BDD y tests previstos

Archivo BDD: `features/HU-31-consultar-defectos.feature`, con el runner `TestFeatures_HU31ConsultarDefectos` en `internal/defect/list_features_test.go`. Escenarios previstos, con su CA en el nombre:

- Listar los defectos de un proyecto en todos sus estados (CA-31.1)
- Listar los defectos detectados en un sprint (CA-31.1)
- No mezclar defectos de otro proyecto (CA-31.2)
- Un defecto resuelto en otro sprint aparece solo en el de detección (CA-31.2)
- Consultar un sprint cerrado (CA-31.2)
- Filtrar los defectos por estado (CA-31.3)
- Rechazar un estado de filtro inválido (CA-31.3)
- Proyecto y sprint sin defectos devuelven una lista vacía (CA-31.4)
- Filtro sin coincidencias devuelve una lista vacía (CA-31.4)
- Rechazar un proyecto y un sprint inexistentes (CA-31.5)
- Modificar lo devuelto no altera lo guardado (CA-31.6)
- Error del repositorio al listar (CA-31.8)
- Consultas simultáneas con registros y cambios de estado (CA-31.7)

Los tests unitarios (`list_test.go`) cubren CA-31.1 a CA-31.8 y los casos límite; los de contrato, PostgreSQL, API e interfaz, los descritos en "Restricciones". El número final de escenarios y tests se cuenta de forma exacta en la evidencia.

## Fuera de alcance

- Registrar defectos (HU-29) o cambiar su estado (HU-30).
- Calcular M9 y M10 (HU-47).
- Reportes de proyecto o de sprint (HU-38 y HU-39) y exportación (HU-40).
- Consultar un defecto por su ID, por historia o de todo el sistema.
- Filtros por severidad, historia o sprint de resolución; paginación y orden configurable.
- Editar o eliminar defectos.
- Cambios de esquema.
- Actualizar `docs/documentacion-tecnica.md` (TEC-06, responsable Peñalbé Hernán).
- Editar el Issue #49 desde esta SDD: los criterios definitivos ("Comparación con los criterios del Issue #49") se cargan en el Issue después de la aprobación, con autorización de Ricardo.

## Comparación con el acta de Planning

El acta del Sprint 4 (`docs/actas/sprint-4/planning.md`) incluye a HU-31 (Issue #49) con **3 Story Points**, responsable **Sosa Ricardo**, prioridad BAJA y dependencias "HU-29 ✅; el filtro por estado usa los estados de HU-30". Su justificación es "Listado por proyecto o por sprint, con filtro por estado, como HU-27". El orden de trabajo la pone después de HU-30, junto con HU-47, y las historias HU-38 y HU-39 dependen de ella.

**El alcance coincide:** esta SDD define el listado por proyecto o por sprint (D1 y D2), el filtro por estado con los estados de HU-30 (D3) y sigue el modelo de HU-27 en la entidad inexistente, la lista vacía, el orden, la API y la interfaz. HU-29 y HU-30 están en `main`. No hay diferencias con el acta.

**Compatibilidad con HU-47** (SDD en la rama `feature/HU-47-defectos-detectados-resueltos`, aprobada con la condición de confirmar con Sosa el nombre y la firma de la lectura): `ListByProject(projectID int) ([]Defect, error)` tiene exactamente esa firma, devuelve copias de todos los defectos del proyecto en cualquier estado y da `project.ErrProjectNotFound` para un proyecto inexistente. HU-47 no tiene que cambiar su interfaz `defectLister`.

## Comparación con los criterios del Issue #49 (borrador)

| CA en el Issue #49 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-31.1 Se listan los defectos del proyecto o del sprint indicado. | CA-31.1 y CA-31.2 | Se define qué defectos son de un proyecto (por su historia) y de un sprint (detectados en él), el orden y el aislamiento |
| CA-31.2 Se puede filtrar por estado. | CA-31.3 | Se identifican los estados, la normalización y `ErrInvalidStatus` |
| CA-31.3 Sin defectos devuelve una lista vacía. | CA-31.4 | Se agrega el filtro sin coincidencias y `[]` en JSON |
| — | CA-31.5 a CA-31.9 (nuevos) | Entidad inexistente, copias y solo lectura, concurrencia, error del repositorio, API e interfaz |

**Texto definitivo para la sección "Criterios de aceptación" del Issue #49** (en lenguaje de negocio; se reemplaza cuando se apruebe la SDD y con autorización de Ricardo):

```markdown
## Criterios de aceptación

- [ ] CA-31.1 Se listan todos los defectos de un proyecto o los detectados en un sprint, con sus datos y su estado, ordenados por número.
- [ ] CA-31.2 La consulta de un proyecto no mezcla defectos de otro proyecto, y la de un sprint no mezcla defectos detectados en otro sprint. Se pueden consultar sprints abiertos y cerrados.
- [ ] CA-31.3 Se puede filtrar por estado (abierto, en corrección o resuelto); sin filtro se ven todos, y un estado inválido se rechaza.
- [ ] CA-31.4 Un proyecto o un sprint sin defectos, o sin defectos en el estado elegido, devuelve una lista vacía.
- [ ] CA-31.5 Consultar un proyecto o un sprint inexistente devuelve un error y no una lista vacía.
- [ ] CA-31.6 Consultar no modifica nada, y lo que se devuelve no permite cambiar los defectos guardados.
- [ ] CA-31.7 Consultar mientras se registran o cambian defectos no produce errores ni datos mezclados.
- [ ] CA-31.8 Un error al leer los datos se informa y no se devuelve una lista.
- [ ] CA-31.9 La consulta se usa desde la API y desde la interfaz web, con tabla, cantidad, caso vacío y mensajes de error entendibles.
```

## Documentación del paso 7

- `docs/evidencias/HU-31-tdd.md`: salidas de `gofmt`, `go vet ./...`, `go test ./...`, `go test -cover ./...` y `go tool cover -func`, trazabilidad con la cantidad exacta de escenarios y tests, y la tabla de commits (RED, GREEN y refactor) con "(propuesto)" hasta tener los hashes.
- `docs/reportes/cobertura.md`: cobertura general actualizada con la salida de `go test -cover ./...`.
- `docs/ia/registro.md`: registro del uso de la IA en la historia.
- `README.md`: rutas, ejemplo `curl`, uso de la interfaz y fila de HU-31 en "Historias por sprint".

HU-31: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
