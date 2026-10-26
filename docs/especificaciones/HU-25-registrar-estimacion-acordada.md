# HU-25 · Registrar la estimación acordada

> Issue del repositorio: #43

Aprobada por: Peñalbé Hernán (Agile Enabler), 26/10/2026.

## Historia de usuario

Como Agile Enabler, quiero registrar la estimación acordada por el equipo, para guardarla en el ítem del backlog.

## Objetivo

Registrar los Story Points que el equipo acordó después de revelar los votos de una sesión de Planning Poker `REVELADA` (HU-23): los puntos se guardan en el ítem de la sesión con la operación de HU-20 y la sesión pasa al estado `CERRADA`. Es la salida normal de una sesión de Planning Poker y cierra la cadena HU-21 → HU-22 → HU-23 → HU-25.

Esta historia agrega la operación `Estimate` al `Manager` de `internal/poker`, el estado `StatusClosed` y el error `ErrSessionNotRevealed`, y expone el registro en `POST /api/poker-sessions/{id}/estimate` y en la sección **Planning Poker** de la interfaz web. Reutiliza `Repository.GetByID` y `Repository.Update` de HU-22 y `backlog.Manager.AssignStoryPoints` de HU-20, sin operaciones nuevas ni cambios de esquema.

## Alcance

Incluye:

- Registrar la estimación acordada de una sesión existente mediante el `Manager`.
- Validar que los puntos pertenezcan a la escala de HU-20, la existencia de la sesión y su estado `REVELADA`.
- Asignar los Story Points al ítem de la sesión con `backlog.Manager.AssignStoryPoints` (HU-20).
- Cambiar el estado de la sesión a `CERRADA` y guardarla en memoria y PostgreSQL (TEC-05) con el `Update` existente.
- Exponer la operación en la API y en la interfaz web.

No incluye:

- Detectar diferencias entre los votos ni repetir la ronda (HU-24).
- Calcular el valor acordado (promedio, moda, consenso): el equipo lo decide y el Agile Enabler lo registra.
- Exigir que el valor acordado coincida con alguno de los votos revelados.
- Reabrir una sesión cerrada, expirar sesiones o impedir abrir otra sesión para el mismo ítem (ver D10).
- Consultar o listar sesiones.
- Validar a los votantes contra la entidad integrante: E4 no está implementada.
- Restringir quién puede registrar la estimación: la aplicación no tiene autenticación ni roles.

## Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---:|---|
| `SessionID` | entero | Sí | ID de la sesión de Planning Poker creada por HU-21 y revelada por HU-23 |
| `Points` | entero | Sí | Story Points acordados; deben pertenecer a la escala de HU-20 (1, 2, 3, 5, 8, 13, 21) |

El ítem, el estado, la ronda y los votos no son entradas: se leen de la sesión guardada.

Firma:

```go
Estimate(sessionID int, points int) (Session, error)
```

en `internal/poker/manager.go`.

## Salidas esperadas

- Si el registro es válido, el ítem de la sesión queda con `StoryPoints = points` y se devuelve la sesión con estado `CERRADA`, como copia por valor, con **todos** sus votos en el orden en que se guardaron. La API responde `200 OK`.
- La ronda, el ítem y los votos de la sesión devuelta son los mismos que estaban guardados: registrar la estimación solo cambia el estado de la sesión y los Story Points del ítem.
- Si el registro no es válido, se devuelve el error correspondiente, la sesión no se modifica (sigue `REVELADA` o en el estado que tenía) y el ítem no cambia, salvo el caso documentado en RN-25.7 (falla al guardar la sesión).

## Reglas de negocio

1. **RN-25.1 Puntos de la escala.** Los puntos acordados deben pertenecer a la escala Fibonacci de HU-20 (`backlog.StoryPointScale()`, expuesta como `poker.Cards()`): 1, 2, 3, 5, 8, 13 o 21. Cualquier otro entero (0, negativos, 4, 22, …) devuelve `ErrInvalidCard`, sin consultar el repositorio ni el backlog. Se valida con `validCard` de `internal/poker/manager.go`, la misma función que usa `Vote`.
2. **RN-25.2 No tienen que coincidir con los votos.** El valor acordado puede ser cualquier carta de la escala, aunque nadie la haya votado: el equipo puede acordar un valor distinto de los votados después de discutir.
3. **RN-25.3 Sesión existente.** La sesión debe existir. Si no existe, se devuelve `ErrSessionNotFound` y no se modifica nada. Esto incluye los IDs 0, negativos y registrar sin sesiones creadas.
4. **RN-25.4 Solo sesiones `REVELADA`.** Solo se puede registrar la estimación de una sesión `REVELADA`. Si está `ABIERTA`, `CERRADA` o en otro estado, se devuelve el error nuevo `ErrSessionNotRevealed` y no se modifica nada (ni la sesión ni el ítem). Una sesión `CERRADA` no se puede volver a estimar: la operación **no es idempotente**.
5. **RN-25.5 Asignar al ítem con HU-20.** Los puntos se asignan con `backlog.Manager.AssignStoryPoints(session.ItemID, points)`, que aplica sus propias reglas (verificadas en `internal/backlog/story_points.go`, función `AssignStoryPoints`):
   - Ítem inexistente: `repo.GetByID` devuelve `backlog.ErrItemNotFound` (declarado en `internal/backlog/get.go`), que se propaga tal cual (HTTP 404).
   - Ítem `TERMINADO`: devuelve `backlog.ErrItemFinished` (RN-20.5, declarado en `story_points.go`), que se propaga tal cual (HTTP 409). Es un error distinto de `poker.ErrItemFinished` (HU-21).
   - Valor fuera de la escala: `backlog.ErrInvalidStoryPoints`. No puede ocurrir desde `Estimate`, porque RN-25.1 ya rechazó esos valores con `ErrInvalidCard` y las dos validaciones usan la misma escala.
   - Reestimar o asignar el mismo valor que ya tiene el ítem es válido (RN-20.3).

   Si `AssignStoryPoints` devuelve cualquier error, la sesión **no se cierra** y sigue `REVELADA`.
6. **RN-25.6 Cerrar la sesión.** Si la asignación fue exitosa, la sesión pasa de `REVELADA` a `CERRADA` y se guarda con `Repository.Update`. No se modifica la ronda ni se borran o reordenan los votos.
7. **RN-25.7 Orden de las dos escrituras.** Primero se asignan los puntos al ítem y después se guarda la sesión `CERRADA`. Si falla el guardado de la sesión, se devuelve el error, **el ítem ya quedó estimado** y la sesión sigue `REVELADA`. Repetir la operación es seguro: reasignar el mismo valor es válido (RN-20.3) y la sesión todavía acepta el registro. Límite conocido (ver D5).
8. **RN-25.8 Copias por valor.** `Estimate` devuelve una copia de la sesión con su propia lista de votos (`cloneSession`). Modificar la sesión o los votos devueltos no altera lo guardado.
9. **RN-25.9 Concurrencia.** La lectura de la sesión, la validación del estado, la asignación de puntos y el guardado de la sesión se realizan bajo el mismo `sync.Mutex` del `Manager` de poker que usan `OpenSession`, `Vote` y `Reveal`. Si dos registros llegan al mismo tiempo para la misma sesión, uno gana y el otro recibe `ErrSessionNotRevealed`; el ítem queda con los puntos del que ganó. `go test -race` no debe informar carreras.
10. **RN-25.10 Orden de validación.** Puntos válidos, existencia de la sesión, estado `REVELADA`, asignación al ítem (existencia y estado del ítem) y guardado de la sesión. Se informa el primer error encontrado; por eso una carta inválida sobre una sesión inexistente devuelve `ErrInvalidCard`, y una sesión `ABIERTA` cuyo ítem no existe devuelve `ErrSessionNotRevealed`.

## Decisión sobre entidades referenciadas

HU-25 referencia por ID a una sesión de Planning Poker y, a través de ella, a un ítem del backlog.

- **Sesión: sí se valida su existencia** con `Repository.GetByID`, igual que en HU-22 y HU-23: es la raíz de la operación.
- **Ítem: sí se valida su existencia y su estado**, pero no lo hace `poker`: lo hace `AssignStoryPoints` dentro de su propio mutex (RN-20.4 y RN-20.5), y `Estimate` propaga sus errores. No se agrega una consulta previa con `GetItem`, porque duplicaría la validación y dejaría una ventana entre la consulta y la escritura. Validarlo es necesario aunque la sesión se haya abierto para un ítem existente (HU-21): HU-21 documenta como límite conocido que el ítem puede pasar a `TERMINADO` mientras la sesión está en curso, y que "HU-25 fallará al intentar registrar la estimación" (SDD de HU-21, "Límite conocido"). Esta historia cumple esa previsión con `backlog.ErrItemFinished`.
- **Proyecto:** no se valida. El ítem pertenece a un proyecto validado al crearlo (HU-04) y esta entrega no contempla el borrado de proyectos ni de ítems.
- **Votantes:** no se validan contra una entidad de integrantes: E4 no está implementada (decisión 4 de la SDD de HU-21).

## Decisiones de diseño

| # | Decisión | Alternativa descartada | Motivo |
|---|---|---|---|
| D1 | `Estimate(sessionID int, points int) (Session, error)` en `internal/poker/manager.go`. Se agregan `StatusClosed Status = "CERRADA"` y `ErrSessionNotRevealed` en `session.go`. Se reutilizan `ErrSessionNotFound`, `ErrInvalidCard` y los errores de `backlog`. | Una operación en `backlog` que cierre la sesión, o un paquete nuevo. | `poker` ya depende de `backlog` (`Manager.items`); al revés crearía un ciclo de imports. La sesión es la raíz de la operación. |
| D2 | Solo se registra la estimación de una sesión `REVELADA`. Inexistente: `ErrSessionNotFound` (404). `ABIERTA`, `CERRADA` u otro estado: `ErrSessionNotRevealed` (409), sin efecto. No es idempotente. | Permitir registrar desde una sesión `ABIERTA`, o devolver `200` si ya estaba `CERRADA`. | Registrar sin revelar saltearía la comparación de estimaciones (HU-23). La idempotencia ocultaría un segundo registro con otro valor; además, una sesión `CERRADA` no tiene nada que registrar. |
| D3 | Los puntos acordados tienen que pertenecer a la escala de HU-20 (`ErrInvalidCard`, 400), pero no tienen que coincidir con ningún voto revelado. Orden: puntos, sesión, estado, guardado. | Exigir que el valor sea uno de los votados. | El equipo puede acordar un valor que nadie votó después de discutir; exigir coincidencia sería una regla que nadie pidió. Se usa `ErrInvalidCard` y no `backlog.ErrInvalidStoryPoints` para que el error sea el mismo que el de un voto con una carta inválida (HU-22). |
| D4 | Efecto: `backlog.Manager.AssignStoryPoints(session.ItemID, points)` (HU-20) y la sesión pasa a `CERRADA`. Ítem inexistente: `backlog.ErrItemNotFound` (404). Ítem `TERMINADO`: `backlog.ErrItemFinished` (409). En ambos casos la sesión no se cierra. | Escribir el campo `StoryPoints` con `backlog.Manager.Update` (HU-07), o validar el ítem antes con `GetItem`. | `AssignStoryPoints` es la única operación que asigna Story Points con la escala y la regla del ítem terminado (nota del Planning: "Guarda los Story Points en el ítem (HU-20)"). `Update` de HU-07 conserva los Story Points actuales (RN-20.7) y no sirve para esto. |
| D5 | Primero se asignan los puntos al ítem y después se guarda la sesión `CERRADA`. Si falla el guardado de la sesión, el ítem ya quedó estimado y la sesión sigue `REVELADA`; repetir la operación es seguro porque reasignar el mismo valor es válido (RN-20.3). **Límite conocido:** las dos escrituras no son atómicas. | Cerrar primero la sesión y después asignar los puntos. | Si falla la asignación (por ejemplo, el ítem pasó a `TERMINADO` o la base falla), quedaría una sesión `CERRADA` sin Story Points en el ítem, y no se podría reintentar porque una sesión `CERRADA` no se vuelve a estimar (D2). Con el orden elegido, el peor caso es un ítem estimado con la sesión todavía `REVELADA`, que se corrige repitiendo la operación. Hacerlas atómicas exigiría una transacción que abarque dos `Manager` y dos repositorios, que ninguna historia pide. |
| D6 | Toda la operación (leer la sesión, validar el estado, asignar los puntos y guardar la sesión) va bajo el mismo `sync.Mutex` del `Manager` de poker. Orden de locks: `poker.Manager.mu` → `backlog.Manager.mu` (este último lo toma `AssignStoryPoints`). | Asignar los puntos fuera del mutex de poker, como `OpenSession` valida el ítem antes de tomarlo (SDD de HU-21: "para no tomar dos mutex a la vez"). | Con la asignación fuera del lock, dos registros simultáneos podrían ver la sesión `REVELADA`, asignar los dos sus puntos y cerrar uno solo: el ítem podría quedar con los puntos del que perdió. Tomar los dos mutex es seguro porque el orden es siempre el mismo y no hay ciclo (ver "Verificación del orden de locks"). |
| D7 | Tests con `-race` de dos `Estimate` simultáneos sobre la misma sesión con puntos distintos: uno devuelve `nil`, el otro `ErrSessionNotRevealed`, la sesión queda `CERRADA` y el ítem con los puntos del que ganó. | Probar solo el registro secuencial. | Es la condición de carrera propia de esta historia (RN-25.9). |
| D8 | Se usan `Repository.GetByID` y `Repository.Update` existentes. No se agregan operaciones ni se cambia el esquema: la columna `status` de `poker_sessions` no tiene restricciones. | Una operación específica `Repository.Close`. | `Update` ya persiste estado, ronda y votos con el mismo contrato de copias en memoria y PostgreSQL. |
| D9 | API: `POST /api/poker-sessions/{id}/estimate` con cuerpo `{"puntos": 5}`, `200` con la misma respuesta de sesión de HU-22 y HU-23 (estado `CERRADA`, votos completos), sin campos nuevos: los Story Points quedan visibles en el ítem (`GET /api/items/{id}`). | `PATCH /api/poker-sessions/{id}` con `{"estado":"CERRADA","puntos":5}`, o agregar `story_points` a la respuesta de la sesión. | La acción es una transición con reglas propias, como `/reveal` y `POST /api/sprints/{id}/cerrar`. Un campo nuevo en la respuesta cambiaría la estructura compartida de HU-21 a HU-23 sin necesidad. |
| D10 | **Cierre del límite D10 de HU-23.** Cerrar la sesión al registrar la estimación es la salida normal de una sesión revelada. Una sesión `REVELADA` que nunca se estima queda como está, y abrir otra sesión para el mismo ítem sigue permitido, porque `OpenByItem` solo busca sesiones `ABIERTA` (RN-21.2). No se cambia HU-21 ni se agrega expiración. | Que `OpenByItem` (o `OpenSession`) considere también las sesiones `REVELADA`. | Cambia una historia ya mergeada (HU-21), y HU-24 (repetir ronda) podría necesitar una sesión `REVELADA` reutilizable. Una sesión `REVELADA` abandonada no afecta los Story Points del ítem ni las métricas. |
| D11 | Interfaz: después de revelar, la sección **Planning Poker** muestra el formulario "Registrar estimación acordada" con un botón por carta de la escala. Al elegir una, se llama a la ruta, se muestra el mensaje y se recarga el backlog con `cargarBacklog()` para que la columna SP del ítem muestre el valor. | Pedir al usuario que vuelva a listar el backlog con "Actualizar". | `cargarBacklog()` ya existe en `internal/web/app.js` y lo usan las demás acciones de las filas (por ejemplo, "Asignar SP" de HU-20), así que refrescar es trivial. |

## Verificación del orden de locks

Hay deadlock solo si alguna operación toma `backlog.Manager.mu` y después espera `poker.Manager.mu`. Verificado en el código de `main` (`82e9ed7`):

- `backlog.Manager.mu` es un `sync.Mutex` no exportado (`internal/backlog/manager.go`). Solo lo toman métodos de `backlog.Manager`: `assign.go`, `create.go`, `get.go`, `list.go`, `status.go`, `story_points.go` y `update.go`.
- El paquete `backlog` no importa `poker`: sus imports internos son solo `internal/project`, `internal/postgres` e `internal/postgrestest`. Los únicos paquetes que importan `internal/poker` son `cmd/server` e `internal/httpapi`, que no tienen mutex propios sobre estos `Manager`. Por lo tanto, ningún código que tiene `backlog.Manager.mu` puede llamar a `poker`.
- Las operaciones de poker existentes no toman los dos mutex: `OpenSession` llama a `items.GetItem` **antes** de tomar `poker.Manager.mu` (`internal/poker/create.go`); `Vote` y `Reveal` no llaman al backlog.
- `Estimate` es la única operación que toma los dos, siempre en el orden `poker.Manager.mu` → `backlog.Manager.mu`.

Con un único orden y sin ninguna espera en sentido contrario, no hay ciclo y no puede haber deadlock. Si una historia futura hace que `backlog` llame a `poker`, tendrá que revisar este orden.

## Restricciones

- `Manager.Estimate` se implementa en `internal/poker/manager.go`, junto a `Vote` y `Reveal`.
- `StatusClosed` y `ErrSessionNotRevealed` se declaran en `internal/poker/session.go`. Se reutilizan `ErrSessionNotFound`, `ErrInvalidCard`, `backlog.ErrItemNotFound` y `backlog.ErrItemFinished`.
- Mensaje de `ErrSessionNotRevealed`: `"la sesión de Planning Poker no está revelada"`.
- No se modifican `Repository`, `MemoryRepository`, `PostgresRepository`, `create.go`, el paquete `backlog` ni el esquema de `db/init/`.
- `Estimate` copia los votos con `cloneSession` antes de devolverlos: nunca devuelve la lista guardada.
- El handler se incorpora en `internal/httpapi/poker.go` y la ruta en `server.go`. `poker.ErrSessionNotRevealed` se agrega al grupo `409` de `statusFor`; `backlog.ErrItemNotFound` (404), `backlog.ErrItemFinished` (409) y `poker.ErrInvalidCard` (400) ya están mapeados.
- **Cuerpo del pedido:** a diferencia de `/reveal`, esta ruta **sí** recibe cuerpo. Se lee con `decodeJSON` (límite de 1 MiB con `http.MaxBytesReader` y rechazo del texto posterior al JSON) en una estructura con `Puntos int` y la etiqueta `json:"puntos"`, igual que `pokerVoteRequest`. La ruta se agrega a `bodyEndpoints` de `server_test.go`, con `expectItemWithoutStoryPoints` como verificación de que el ítem no cambió, para las pruebas `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400`.
- **Método no permitido:** como la interfaz web responde en `GET /`, se registra `GET /api/poker-sessions/{id}/estimate` con el handler `methodNotAllowed` existente, igual que `/reveal`. Los demás métodos los rechaza `net/http` con `405`.
- La ruta se agrega a `TestRutasConID_IDNoNumericoDevuelve400` de `server_test.go`, con un cuerpo válido.
- Un error interno (`500`) responde `errInternal`, sin el detalle del error del repositorio.
- Solo se usa la librería estándar en producción; godog se utiliza únicamente en pruebas.
- RED: si el paquete no compila por la funcionalidad nueva, se agregan primero las firmas vacías (stubs) de `Estimate`, `StatusClosed` y `ErrSessionNotRevealed`, para que el RED falle por las aserciones (Definition of Done del Sprint 3). `go vet ./...` antes del commit RED.

## Contrato HTTP

### Endpoint

```text
POST /api/poker-sessions/{id}/estimate
```

### Cuerpo

```json
{ "puntos": 5 }
```

### Respuesta exitosa

```json
{
  "id": 1,
  "item_id": 5,
  "estado": "CERRADA",
  "ronda": 1,
  "votos": [
    { "integrante": "Ana", "puntos": 5 },
    { "integrante": "Bruno", "puntos": 8 }
  ],
  "cartas": [1, 2, 3, 5, 8, 13, 21]
}
```

`200 OK`, con la misma estructura `pokerSessionResponse` de HU-21 a HU-23. Los votos se arman con `pokerVoteResponse`, en el orden en que se guardaron. Los Story Points acordados se consultan en el ítem (`GET /api/items/{id}`, campo `story_points`).

### Errores

| Caso | Código | Mensaje |
|---|---:|---|
| `{id}` no es un entero | 400 | `errInvalidID` |
| Cuerpo que no es JSON válido, con texto después del JSON o con `puntos` que no es un entero | 400 | `errInvalidJSON` |
| Carta fuera de la escala (incluido `puntos` ausente, que vale 0) | 400 | `poker.ErrInvalidCard` |
| Sesión inexistente | 404 | `poker.ErrSessionNotFound` |
| Ítem de la sesión inexistente | 404 | `backlog.ErrItemNotFound` |
| Sesión no `REVELADA` (`ABIERTA` o `CERRADA`) | 409 | `poker.ErrSessionNotRevealed` |
| Ítem de la sesión `TERMINADO` | 409 | `backlog.ErrItemFinished` |
| Cuerpo mayor de 1 MiB | 413 | `errBodyTooLarge` |
| Falla del repositorio al leer o guardar (sesión o ítem) | 500 | `errInternal`, sin detalle |
| Método distinto de `POST` | 405 | `GET` con `methodNotAllowed`; los demás, respuesta de `net/http` |

El orden de validación es: ID de la ruta, cuerpo, puntos válidos, existencia de la sesión, estado, asignación al ítem y guardado de la sesión.

## Interfaz web

En la sección **Planning Poker**, cuando la sesión mostrada se reveló (HU-23):

- Aparece el formulario **"Registrar estimación acordada"** con un botón por cada carta de la escala (`s.cartas`).
- Al elegir una carta, se llama a `POST /api/poker-sessions/{id}/estimate` con `{"puntos": N}` para la sesión mostrada.
- Al registrar con éxito:
  - Se muestra el estado `CERRADA` en los datos de la sesión.
  - Se oculta el formulario "Registrar estimación acordada", porque la sesión ya no acepta otro registro. La lista de votos revelados se sigue mostrando.
  - El mensaje común informa `"Estimación registrada: N story points en la sesión S."`.
  - Se recarga el backlog del proyecto actual con `cargarBacklog()`, así la fila del ítem muestra el valor en la columna SP. Si el usuario cambió de proyecto actual mientras tanto, se recarga el backlog de ese proyecto y la fila del ítem no aparece en la lista; el valor se ve al volver a su proyecto.
- Si la API responde con error (por ejemplo, ítem terminado), el mensaje común muestra el mensaje de la API y la sección no cambia.
- Cuando se abre una sesión nueva (`mostrarPoker`), el formulario se oculta hasta revelarla.

Tests en `internal/httpapi/web_test.go`: la página contiene el formulario y su contenedor de botones, y `app.js` usa la ruta nueva, el mensaje y `cargarBacklog`.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Sesión revelada, valor igual a uno de los votos | `CERRADA`; el ítem queda con ese valor |
| Sesión revelada, valor que nadie votó (por ejemplo, votos 5 y 8, acordado 13) | Se acepta (RN-25.2) |
| Valores extremos de la escala: 1 y 21 | Se aceptan |
| Puntos 0, negativos, 4 o 22 | `ErrInvalidCard`; la sesión y el ítem no cambian |
| Carta inválida sobre una sesión inexistente | `ErrInvalidCard` (los puntos se validan primero) |
| Ítem que ya tenía Story Points | Se reemplazan por el valor acordado (RN-20.3) |
| Ítem que ya tenía el mismo valor | Válido; el ítem queda igual y la sesión `CERRADA` |
| Sesión `ABIERTA` (sin revelar) | `ErrSessionNotRevealed`; el ítem no cambia |
| Registrar dos veces la misma sesión | La segunda devuelve `ErrSessionNotRevealed`; el ítem conserva el primer valor y la sesión sigue `CERRADA` |
| Sesión inexistente, ID 0 o negativo | `ErrSessionNotFound` |
| Ítem de la sesión inexistente | `backlog.ErrItemNotFound`; la sesión sigue `REVELADA` |
| Ítem de la sesión `TERMINADO` | `backlog.ErrItemFinished`; la sesión sigue `REVELADA` y el ítem no cambia. Si el ítem se reabre con HU-08, vuelve a ser estimable (RN-20.5) y la sesión se puede cerrar después; si no se reabre, la sesión queda `REVELADA` (ver D10) |
| Falla del backlog al asignar | Se propaga el error; la sesión sigue `REVELADA` |
| Falla del repositorio de poker al guardar la sesión | Se propaga el error; el ítem **ya quedó estimado** y la sesión sigue `REVELADA`; repetir la operación con el mismo valor la cierra (RN-25.7) |
| Dos registros simultáneos sobre la misma sesión | Uno gana; el otro recibe `ErrSessionNotRevealed`; el ítem queda con los puntos del ganador |
| Votar o revelar una sesión `CERRADA` | `ErrSessionNotOpen` (HU-22 y HU-23, sin cambios) |
| Modificar la sesión o los votos devueltos | La sesión guardada no cambia |
| Ronda y votos de la sesión | No cambian al registrar |
| Abrir otra sesión para el mismo ítem con una `REVELADA` sin cerrar | Se permite (D10) |
| Recargar la página con una sesión revelada | La interfaz pierde la sesión mostrada y no hay endpoint para consultarla (límite de HU-21 a HU-23); con la API (`POST /api/poker-sessions/{id}/estimate`) se puede registrar igual |
| Pedido HTTP con `"puntos":"5"` o `"puntos":5.5` | `400` con `errInvalidJSON` |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Puntos fuera de la escala | `poker.ErrInvalidCard` | No se modifica nada |
| Sesión inexistente | `poker.ErrSessionNotFound` | No se modifica nada |
| Sesión no `REVELADA` | `poker.ErrSessionNotRevealed` | No se modifica nada |
| Ítem inexistente | `backlog.ErrItemNotFound` | La sesión sigue `REVELADA` |
| Ítem `TERMINADO` | `backlog.ErrItemFinished` | La sesión sigue `REVELADA`; el ítem no cambia |
| Falla del repositorio de poker al leer | Error del repositorio | No se modifica nada |
| Falla del backlog al asignar | Error del backlog | La sesión sigue `REVELADA`; el ítem no cambia |
| Falla del repositorio de poker al guardar | Error del repositorio | El ítem queda estimado; la sesión sigue `REVELADA` (RN-25.7) |

## Criterios de aceptación

- **CA-25.1** Al registrar el valor acordado en una sesión revelada, el valor queda guardado como Story Points del ítem de la sesión.
- **CA-25.2** Solo se puede registrar un valor de la escala (1, 2, 3, 5, 8, 13, 21); cualquier otro da error y no cambia ni la sesión ni el ítem. El valor no tiene que coincidir con los votos.
- **CA-25.3** Al registrar la estimación, la sesión de Planning Poker queda cerrada, con su ronda y sus votos sin cambios.
- **CA-25.4** Registrar la estimación de una sesión inexistente, abierta o ya cerrada da error y no modifica la sesión ni el ítem.
- **CA-25.5** Si el ítem de la sesión no existe o está terminado, se informa el error del backlog y la sesión sigue revelada.
- **CA-25.6** Si falla el guardado de la sesión, se informa el error, el ítem queda estimado y la sesión sigue revelada, de modo que repetir el registro la cierra.
- **CA-25.7** Modificar la sesión devuelta no altera lo guardado.
- **CA-25.8** Si se registran dos estimaciones al mismo tiempo para la misma sesión, solo una se aplica: la otra da error y el ítem queda con los puntos de la que se aplicó.
- **CA-25.9** Se registra desde la API (`POST /api/poker-sessions/{id}/estimate`) y desde el formulario "Registrar estimación acordada" de la interfaz web. La API responde `200` en el caso exitoso y `400`, `404`, `409`, `413`, `500` o `405` según la tabla de errores; la interfaz muestra el mensaje, la sesión cerrada y el backlog actualizado, o el mensaje de error.
- **CA-25.10** El estado cerrado se guarda en memoria y en PostgreSQL con el mismo comportamiento, verificado por los tests de contrato (se omiten sin `TEST_DATABASE_URL`).

## Pruebas previstas

Todos los tests listados se escriben en la fase RED. Los obligatorios desde el RED son: `404` sesión, `404` ítem, `409` no revelada (`ABIERTA` y `CERRADA`), `409` ítem terminado, `400` carta inválida, `500` por falla al guardar la sesión (con verificación de que el ítem queda estimado y la sesión sigue `REVELADA`), `400` por ID no numérico, `413` y la concurrencia con `-race`.

### Unitarios (`internal/poker/estimate_test.go`)

| Test | CA |
|---|---|
| `TestEstimate_SesionReveladaAsignaLosPuntosYCierraLaSesion` | CA-25.1, CA-25.3 |
| `TestEstimate_ValorQueNadieVotoSeAcepta` | CA-25.2 |
| `TestEstimate_AceptaCadaValorDeLaEscala` | CA-25.2 |
| `TestEstimate_CartaFueraDeLaEscalaDevuelveErrInvalidCard` (0, negativo, 4, 22; la sesión sigue `REVELADA` y el ítem sin cambios) | CA-25.2 |
| `TestEstimate_CartaInvalidaSeInformaAntesQueLaSesionInexistente` | CA-25.2 |
| `TestEstimate_ReemplazaLosStoryPointsAnterioresDelItem` | CA-25.1 |
| `TestEstimate_SesionInexistenteDevuelveErrSessionNotFound` (ID inexistente, 0 y negativo) | CA-25.4 |
| `TestEstimate_SesionAbiertaDevuelveErrSessionNotRevealed` (el ítem no cambia) | CA-25.4 |
| `TestEstimate_SesionCerradaDevuelveErrSessionNotRevealed` (el ítem conserva el primer valor) | CA-25.4 |
| `TestEstimate_ItemInexistenteDevuelveErrItemNotFoundYSesionSigueRevelada` | CA-25.5 |
| `TestEstimate_ItemTerminadoDevuelveErrItemFinishedYSesionSigueRevelada` | CA-25.5 |
| `TestEstimate_ErrorAlLeerPropagaError` | CA-25.4 |
| `TestEstimate_ErrorDelBacklogAlAsignarPropagaErrorYSesionSigueRevelada` | CA-25.5 |
| `TestEstimate_ErrorAlGuardarLaSesionDejaElItemEstimadoYLaSesionRevelada` (y repetir con el mismo valor la cierra) | CA-25.6 |
| `TestEstimate_ConservaRondaYOrdenDeLosVotos` | CA-25.3 |
| `TestEstimate_DevuelveCopiaDeLosVotos` | CA-25.7 |
| `TestVoteYReveal_SesionCerradaDevuelvenErrSessionNotOpen` | CA-25.3 |
| `TestEstimate_SimultaneosSobreLaMismaSesionSoloUnoSeAplica` (con `-race`) | CA-25.8 |

El test de concurrencia parte de una sesión `REVELADA` con votos y lanza dos `Estimate` al mismo tiempo con puntos distintos (por ejemplo, 5 y 8). Verifica que: exactamente uno devuelve `nil` y el otro `ErrSessionNotRevealed`; la sesión queda `CERRADA`; y el ítem tiene los Story Points del que devolvió `nil`. Se repite varias veces en el mismo test para aumentar la probabilidad de intercalar las llamadas.

Las fallas de repositorio usan un repositorio de poker de prueba que falla al leer o al guardar (como `failingRevealRepository` de HU-23) y, para la falla del backlog, un `backlog.Manager` sobre un repositorio de prueba que falla al actualizar.

### Contrato de repositorio (`internal/poker/repository_contract_test.go`)

- Subtest nuevo: `Update guarda el estado CERRADA y GetByID lo devuelve con sus votos` (CA-25.10). Corre contra `MemoryRepository` y contra `PostgresRepository`; este último se omite sin `TEST_DATABASE_URL`.

### BDD (`features/HU-25-registrar-estimacion-acordada.feature`)

Runner `TestFeatures_HU25RegistrarEstimacionAcordada` en `internal/poker/estimate_features_test.go`. **Todos los pasos llaman al `Manager` de `internal/poker` (y al de `internal/backlog` para preparar y consultar el ítem), no al handler HTTP**, por eso ningún escenario dice "mediante la API". Antecedentes: un proyecto de prueba, un ítem y una sesión para ese ítem con votos, revelada.

| Escenario | Tipo |
|---|---|
| Registrar la estimación acordada guarda los puntos en el ítem y cierra la sesión (CA-25.1, CA-25.3) | Normal |
| Registrar un valor que nadie votó (CA-25.2) | Alternativo |
| Registrar la estimación reemplaza los Story Points anteriores del ítem (CA-25.1) | Alternativo |
| Registrar el valor más alto de la escala (CA-25.2) | Límite |
| Rechazar una carta fuera de la escala (CA-25.2) | Error |
| Rechazar registrar la estimación de una sesión sin revelar (CA-25.4) | Error |
| Rechazar registrar dos veces la misma sesión (CA-25.4) | Error |
| Rechazar registrar la estimación de una sesión inexistente (CA-25.4) | Error |
| Rechazar registrar la estimación de un ítem terminado (CA-25.5) | Error |
| Dos registros simultáneos: solo uno se aplica (CA-25.8) | Límite |

### API (`internal/httpapi/poker_test.go` y `server_test.go`)

| Test | CA |
|---|---|
| `TestRegistrarEstimacionPoker_Devuelve200ConSesionCerradaYStoryPointsEnElItem` (consulta `GET /api/items/{id}`) | CA-25.9 |
| `TestRegistrarEstimacionPoker_IDNoNumericoDevuelve400` | CA-25.9 |
| `TestRegistrarEstimacionPoker_JSONInvalidoDevuelve400` | CA-25.9 |
| `TestRegistrarEstimacionPoker_CartaFueraDeLaEscalaDevuelve400` | CA-25.9 |
| `TestRegistrarEstimacionPoker_SesionInexistenteDevuelve404` | CA-25.9 |
| `TestRegistrarEstimacionPoker_ItemInexistenteDevuelve404` (sesión guardada directamente en el repositorio con un ítem inexistente) | CA-25.9 |
| `TestRegistrarEstimacionPoker_SesionNoReveladaDevuelve409` (`ABIERTA` y `CERRADA`) | CA-25.9 |
| `TestRegistrarEstimacionPoker_ItemTerminadoDevuelve409` | CA-25.9 |
| `TestRegistrarEstimacionPoker_ErrorAlGuardarLaSesionDevuelve500SinDetalle` (el ítem queda estimado y la sesión sigue `REVELADA`) | CA-25.9 |
| `TestRegistrarEstimacionPoker_MetodoNoPermitidoDevuelve405` | CA-25.9 |
| Caso `"registrar estimación Planning Poker"` en `TestRutasConID_IDNoNumericoDevuelve400` | CA-25.9 |
| Entrada `"registrar estimación Planning Poker"` en `bodyEndpoints`, que corre en `TestHandler_CuerpoMayorA1MiBDa413` y `TestHandler_TextoDespuesDelJSONDa400` | CA-25.9 |

### Interfaz (`internal/httpapi/web_test.go`)

| Test | CA |
|---|---|
| `TestInterfaz_FormularioRegistrarEstimacion` (formulario, contenedor de cartas, ruta, mensaje y `cargarBacklog` en `app.js`) | CA-25.9 |

## Comparación con el Planning del Sprint 3

El acta del Sprint 3 (`docs/actas/sprint-3/planning.md`) asigna a HU-25 **2 Story Points**, con dependencia de HU-20 y HU-23 y sin historias bloqueadas ("—"). La nota del Planning establece: “Guarda los Story Points en el ítem (HU-20) y cierra la sesión”. Esta SDD implementa ese alcance con RN-25.5 (asignación con `AssignStoryPoints` de HU-20) y RN-25.6 (sesión `CERRADA`).

La especificación también precisa, sin ampliar el alcance funcional del acta:

- Solo se registra la estimación de una sesión revelada, y una sesión cerrada no se vuelve a estimar.
- El valor acordado debe ser de la escala, pero no tiene que coincidir con los votos.
- El orden de las dos escrituras y su límite conocido (RN-25.7).
- El cierre de la decisión D10 de HU-23, sin cambiar HU-21.

No hay diferencias de alcance con el acta de Planning.

## Comparación con los criterios del Issue #43 (borrador)

| Criterio del Issue | En esta SDD |
|---|---|
| CA-25.1 Se registra el valor acordado y queda guardado como story points del ítem. | CA-25.1 |
| CA-25.2 Solo se puede registrar un valor válido de la escala. | CA-25.2, con `ErrInvalidCard` (400) y la aclaración de que no tiene que coincidir con los votos |
| CA-25.3 Se cierra la sesión de Planning Poker. | CA-25.3, con el estado `CERRADA` |

CA-25.4 a CA-25.10 agregan los casos de error de la sesión y del ítem, la falla del repositorio, las copias, la concurrencia, la API, la interfaz y la persistencia, con el mismo criterio que HU-21 a HU-23. Los criterios del Issue siguen marcados como "(borrador)": actualizarlos según esta SDD requiere autorización y no se hace en este paso.

## Documentación a actualizar en el ciclo

- `README.md`:
  - Fila de HU-25 en "Historias por sprint", con Issue #43, SDD, BDD, evidencia y PR.
  - Fila del endpoint `POST /api/poker-sessions/{id}/estimate` en la tabla de la API.
  - Descripción de la sección **Planning Poker** de la interfaz: el formulario "Registrar estimación acordada", que guarda los Story Points en el ítem y deja la sesión CERRADA (HU-25).
- `docs/reportes/cobertura.md`: cobertura de HU-25, indicando el commit sobre el que se midió y, para la medición con base, la versión real de PostgreSQL (imagen `postgres:17-alpine` del `docker-compose.yml`; la versión exacta se toma de la base levantada al medir).
- `docs/evidencias/HU-25-tdd.md`: evidencia de RED, GREEN, refactor y validación, con la tabla de commits.
- `docs/ia/registro.md`: registro del uso de IA en la historia.

La SDD va en su propio commit, aparte del commit RED, cuando se autorice.

HU-25: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
