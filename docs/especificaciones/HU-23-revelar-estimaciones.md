# HU-23 · Revelar las estimaciones

> Issue del repositorio: #41

Aprobada por: Peñalbé Hernán (Agile Enabler), 25/10/2026.

## Historia de usuario

Como Agile Enabler, quiero revelar los votos de todos a la vez, para comparar las estimaciones.

## Objetivo

Revelar de una sola vez todos los votos de una sesión de Planning Poker `ABIERTA` (HU-21) que tenga al menos un voto (HU-22): la sesión pasa al estado `REVELADA` y se devuelve con la lista completa de votos. A partir de ese momento la sesión deja de aceptar votos, así que nadie puede cambiar su estimación después de ver las de los demás.

Esta historia agrega la operación `Reveal` al `Manager` de `internal/poker`, el estado `StatusRevealed` y el error `ErrNoVotes`, y expone la revelación en `POST /api/poker-sessions/{id}/reveal` y en la sección **Planning Poker** de la interfaz web. Reutiliza `Repository.GetByID` y `Repository.Update` de HU-22, sin operaciones nuevas ni cambios de esquema.

## Alcance

Incluye:

- Revelar una sesión existente mediante el `Manager`.
- Validar la existencia de la sesión, su estado `ABIERTA` y que tenga al menos un voto.
- Cambiar el estado de la sesión a `REVELADA` y devolver todos sus votos a la vez.
- Rechazar los votos posteriores a la revelación (regla ya implementada en HU-22, verificada con un caso nuevo).
- Guardar el estado revelado en memoria y PostgreSQL (TEC-05) con el `Update` existente.
- Exponer la operación en la API y en la interfaz web.

No incluye:

- Detectar diferencias entre los votos ni repetir la ronda (HU-24).
- Registrar la estimación acordada, escribir Story Points en el ítem ni cerrar la sesión (HU-25).
- Calcular promedios, mínimos, máximos o consenso de los votos.
- Consultar los votos de una sesión después de revelarla, listar o consultar sesiones.
- Validar a los votantes contra la entidad integrante: E4 no está implementada.
- Restringir quién puede revelar: la aplicación no tiene autenticación ni roles.

## Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---:|---|
| `SessionID` | entero | Sí | ID de la sesión de Planning Poker creada por HU-21 |

El estado, la ronda y los votos no son entradas: se leen de la sesión guardada.

Firma:

```go
Reveal(sessionID int) (Session, error)
```

en `internal/poker/manager.go`.

## Salidas esperadas

- Si la revelación es válida, se devuelve la sesión con estado `REVELADA`, como copia por valor, con **todos** los votos guardados (integrante y puntos) en el orden en que se guardaron. La API responde `200 OK`.
- La ronda, el ítem y los votos devueltos son los mismos que estaban guardados: revelar solo cambia el estado.
- Si la revelación no es válida, se devuelve el error correspondiente y la sesión no se modifica: sigue `ABIERTA` (o en el estado que tenía) con sus votos ocultos.

## Reglas de negocio

1. **RN-23.1 Sesión existente.** La sesión debe existir. Si no existe, se devuelve `ErrSessionNotFound` y no se modifica ninguna sesión. Esto incluye los IDs 0, negativos y revelar sin sesiones creadas.
2. **RN-23.2 Solo sesiones `ABIERTA`.** Solo se puede revelar una sesión `ABIERTA`. Si tiene otro estado, incluida una sesión ya `REVELADA`, se devuelve `ErrSessionNotOpen` y no se modifica. La segunda revelación **no es idempotente**: falla con el mismo error que cualquier sesión no abierta.
3. **RN-23.3 Al menos un voto.** Una sesión `ABIERTA` sin votos no se puede revelar: se devuelve el error nuevo `ErrNoVotes` y la sesión sigue `ABIERTA`. Con un solo voto ya se puede revelar.
4. **RN-23.4 Revelar solo cambia el estado.** La sesión pasa de `ABIERTA` a `REVELADA`. No se modifica la ronda, no se borran ni reordenan los votos, no se modifica el ítem y no se escriben Story Points (eso es HU-25).
5. **RN-23.5 Todos los votos a la vez.** La sesión devuelta incluye todos los votos guardados, en el orden en que se guardaron, como una lista nueva. Es la primera operación que expone votos: hasta ahora, `OpenSession` y `Vote` los devolvían vacíos (RN-22.8).
6. **RN-23.6 Copias por valor.** `Reveal` devuelve una copia de la sesión con su propia lista de votos. Modificar la sesión o los votos devueltos no altera lo guardado.
7. **RN-23.7 Los votos quedan fijos.** Después de revelar, `Vote` devuelve `ErrSessionNotOpen` (RN-22.5) y no agrega ni reemplaza votos. Esto cumple el criterio del Issue de que los votos no se puedan cambiar después de revelarlos.
8. **RN-23.8 Concurrencia.** La lectura, la validación, el cambio de estado y el guardado se realizan bajo el mismo `sync.Mutex` del `Manager` que usan `OpenSession` y `Vote`. Si se vota y se revela al mismo tiempo sobre la misma sesión, cada voto queda guardado antes de la revelación (y aparece en los votos revelados) o se rechaza con `ErrSessionNotOpen`. Nunca se pierde un voto aceptado y `go test -race` no debe informar carreras.
9. **RN-23.9 Orden de validación.** Existencia de la sesión, estado `ABIERTA`, al menos un voto y guardado. Se informa el primer error encontrado; por eso una sesión ya revelada devuelve `ErrSessionNotOpen` y no `ErrNoVotes`.

## Decisión sobre entidades referenciadas

HU-23 referencia por ID a una sesión de Planning Poker. **Sí se valida su existencia** con `Repository.GetByID`, igual que en HU-22: la sesión es la raíz de la operación y revelar una sesión inexistente no tiene sentido.

No se valida nuevamente el ítem ni el proyecto: la sesión apunta a un ítem validado al abrirse (HU-21), revelar no lo modifica y esta entrega no contempla el borrado de ítems ni sesiones.

No se validan los votantes contra una entidad de integrantes: E4 no está implementada y el votante es un nombre libre (decisión 4 de la SDD de HU-21).

## Decisiones de diseño

| # | Decisión | Alternativa descartada | Motivo |
|---|---|---|---|
| D1 | `Reveal(sessionID int) (Session, error)` en `internal/poker/manager.go`, bajo el mismo `sync.Mutex` del `Manager` que usa `Vote`: leer, validar, cambiar el estado y guardar dentro del lock. Se agrega `StatusRevealed Status = "REVELADA"` en `session.go`. | Un mutex propio de la revelación, o leer fuera del lock y guardar dentro. | Con un lock distinto, un voto podría leer la sesión `ABIERTA`, la revelación guardarla `REVELADA` y el voto pisarla después, perdiendo la revelación o un voto aceptado (RN-23.8). |
| D2 | Solo se revela una sesión `ABIERTA`. Inexistente: `ErrSessionNotFound` (404). No `ABIERTA`, incluida una ya revelada: `ErrSessionNotOpen` (409), sin efecto. | Revelación idempotente: devolver `200` con los votos si ya estaba revelada. | La idempotencia funcionaría como una consulta de votos, que está fuera de alcance; además reusa el error existente de HU-22 sin crear uno nuevo. |
| D3 | Una sesión sin votos no se puede revelar: error nuevo `ErrNoVotes`, HTTP 409, y el estado no cambia. | Permitir revelar y devolver una lista vacía. | Lo pide el criterio CA-23.2 del Issue. Es 409 y no 400 porque el pedido es válido; lo que impide revelar es el estado actual de la sesión. |
| D4 | Revelar solo cambia el estado a `REVELADA`. | Calcular un resumen (promedio, consenso) o escribir los Story Points en el ítem. | El resumen no está pedido y escribir los Story Points es HU-25; la nota del Planning se limita a cambiar el estado y devolver los votos. |
| D5 | La sesión devuelta incluye todos los votos como copia por valor, en el orden en que se guardaron. Después de revelar, `Vote` falla con `ErrSessionNotOpen` (HU-22), verificado con un caso nuevo. | Ordenar los votos por nombre o por puntos. | Ordenar agrega una regla que nadie pidió; el orden de guardado es el que ya conserva el repositorio. |
| D6 | Se usan `Repository.GetByID` y `Repository.Update` existentes. No se agregan operaciones ni se cambia el esquema: la columna `status` de `poker_sessions` no tiene restricciones. | Una operación específica `Repository.Reveal`. | `Update` ya persiste estado, ronda y votos con el mismo contrato de copias en memoria y PostgreSQL. |
| D7 | API: `POST /api/poker-sessions/{id}/reveal`, sin cuerpo, `200` con la misma respuesta de sesión de HU-22 pero con `votos` completos (reusa `pokerVoteResponse`). | `PATCH /api/poker-sessions/{id}` con `{"estado":"REVELADA"}`. | La acción es una transición con reglas propias, como `POST /api/sprints/{id}/cerrar`; un `PATCH` genérico sugiere que se puede pasar a cualquier estado. |
| D8 | Interfaz: botón "Revelar votos" junto al formulario de voto; al revelar se muestra la lista de votos (integrante y puntos), se ocultan los botones de voto y el resultado va por el mecanismo común de mensajes. | Un endpoint para consultar votos de una sesión revelada. | La lista se arma con la respuesta de la revelación. Listar o consultar sesiones sigue fuera de alcance. |
| D9 | Tests con `-race` de `Vote` y `Reveal` simultáneos sobre la misma sesión: cada voto concurrente queda guardado antes de revelar o se rechaza con `ErrSessionNotOpen`. | Probar solo la revelación secuencial. | Es la condición de carrera propia de esta historia (RN-23.8). |
| D10 | Se mantiene el comportamiento de HU-21: se puede abrir otra sesión para el mismo ítem mientras la anterior está `REVELADA`, porque `OpenByItem` solo busca sesiones `ABIERTA` (RN-21.2). Límite conocido; HU-24 y HU-25 deberán decidir cómo tratan una sesión revelada sin cerrar. | Impedir abrir una sesión nueva si hay una `REVELADA`. | Cambia HU-21, que ya está mergeada, y no lo pide el Planning. |

## Restricciones

- `Manager.Reveal` se implementa en `internal/poker/manager.go`, junto a `Vote`.
- `StatusRevealed` y `ErrNoVotes` se declaran en `internal/poker/session.go`. Se reutilizan `ErrSessionNotFound` y `ErrSessionNotOpen`.
- Mensaje de `ErrNoVotes`: `"la sesión de Planning Poker no tiene votos para revelar"`.
- No se modifican `Repository`, `MemoryRepository`, `PostgresRepository`, `create.go` ni el esquema de `db/init/`.
- `Reveal` copia los votos con `cloneSession` (o equivalente) antes de devolverlos: nunca devuelve la lista guardada.
- El handler se incorpora en `internal/httpapi/poker.go` y la ruta en `server.go`. `poker.ErrNoVotes` se agrega al grupo `409` de `statusFor`.
- **Cuerpo del pedido:** la ruta no recibe cuerpo. El handler no lo lee: un cuerpo enviado se ignora, igual que en `POST /api/sprints/{id}/cerrar`, y no se devuelve `400` ni `413` por su contenido. Por eso la ruta **no** se agrega a `bodyEndpoints`; un test verifica que un cuerpo arbitrario no cambia el resultado.
- **Método no permitido:** como la interfaz web responde en `GET /`, un `GET` a la ruta caería en el servidor de archivos. Se registra `GET /api/poker-sessions/{id}/reveal` con el handler `methodNotAllowed` existente, igual que HU-21 hizo con `GET /api/items/{id}/poker-sessions`. Los demás métodos los rechaza `net/http` con `405`.
- La ruta se agrega a `TestRutasConID_IDNoNumericoDevuelve400` de `server_test.go`.
- Un error interno (`500`) responde `errInternal`, sin el detalle del error del repositorio.
- Solo se usa la librería estándar en producción; godog se utiliza únicamente en pruebas.
- RED: si el paquete no compila por la funcionalidad nueva, se agregan primero las firmas vacías (stubs) de `Reveal`, `StatusRevealed` y `ErrNoVotes`, para que el RED falle por las aserciones (Definition of Done del Sprint 3). `go vet ./...` antes del commit RED.

## Contrato HTTP

### Endpoint

```text
POST /api/poker-sessions/{id}/reveal
```

### Cuerpo

Sin cuerpo. Si se envía uno, se ignora.

### Respuesta exitosa

```json
{
  "id": 1,
  "item_id": 5,
  "estado": "REVELADA",
  "ronda": 1,
  "votos": [
    { "integrante": "Ana", "puntos": 5 },
    { "integrante": "Bruno", "puntos": 8 }
  ],
  "cartas": [1, 2, 3, 5, 8, 13, 21]
}
```

`200 OK`, con la misma estructura `pokerSessionResponse` de HU-21 y HU-22. Los votos se arman con `pokerVoteResponse`, en el orden en que se guardaron.

### Errores

| Caso | Código | Mensaje |
|---|---:|---|
| `{id}` no es un entero | 400 | `errInvalidID` |
| Sesión inexistente | 404 | `poker.ErrSessionNotFound` |
| Sesión no `ABIERTA` (incluida una ya revelada) | 409 | `poker.ErrSessionNotOpen` |
| Sesión `ABIERTA` sin votos | 409 | `poker.ErrNoVotes` |
| Falla del repositorio al leer o guardar | 500 | `errInternal`, sin detalle |
| Método distinto de `POST` | 405 | `GET` con `methodNotAllowed`; los demás, respuesta de `net/http` |

El orden de validación es: ID de la ruta, existencia de la sesión, estado, votos y guardado.

## Interfaz web

En la sección **Planning Poker**, cuando hay una sesión abierta:

- Un botón **"Revelar votos"** junto al formulario de voto, que llama a `POST /api/poker-sessions/{id}/reveal` con la sesión mostrada.
- Al revelar con éxito:
  - Se muestra el estado `REVELADA` en los datos de la sesión.
  - Se muestra la lista de votos, una línea por voto con el integrante y los puntos.
  - Se ocultan los botones de voto (y el botón "Revelar votos"), porque la sesión ya no acepta votos.
  - El mensaje común informa `"Votos revelados en la sesión N."`.
- Si la API responde con error (por ejemplo, sin votos), el mensaje común muestra el mensaje de la API y la sección no cambia.

No hay endpoint para volver a consultar los votos: la lista se arma con la respuesta de la revelación. Tests en `internal/httpapi/web_test.go`: la página contiene el botón y el contenedor de la lista, y `app.js` usa la ruta nueva.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Sesión abierta con varios votos | `REVELADA`, con todos los votos en el orden guardado |
| Sesión abierta con un único voto | Se revela con ese voto |
| Sesión abierta sin votos | `ErrNoVotes`; sigue `ABIERTA` |
| Voto reemplazado antes de revelar (HU-22) | Se revela el último voto del integrante, sin duplicados |
| Revelar dos veces la misma sesión | La segunda devuelve `ErrSessionNotOpen`; la sesión sigue `REVELADA` con los mismos votos |
| Sesión inexistente, ID 0 o negativo | `ErrSessionNotFound` |
| Votar después de revelar | `ErrSessionNotOpen`; los votos revelados no cambian |
| Votos y revelación simultáneos | Cada voto queda guardado antes de revelar o se rechaza con `ErrSessionNotOpen`; ningún voto aceptado se pierde |
| Modificar la sesión o los votos devueltos | La sesión guardada no cambia |
| Ronda, ítem y Story Points del ítem | No cambian al revelar |
| Falla del repositorio al leer | Se propaga el error; la sesión no cambia |
| Falla del repositorio al guardar | Se propaga el error; la sesión sigue `ABIERTA` con sus votos |
| Pedido HTTP con cuerpo | Se ignora; el resultado es el mismo que sin cuerpo |
| Abrir otra sesión para el mismo ítem mientras la anterior está `REVELADA` | Se permite: HU-21 solo busca sesiones `ABIERTA` (RN-21.2). Límite conocido; HU-24 y HU-25 deberán decidir cómo tratan una sesión revelada sin cerrar |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Sesión inexistente | `poker.ErrSessionNotFound` | No se modifica ninguna sesión |
| Sesión no `ABIERTA` | `poker.ErrSessionNotOpen` | No se modifica la sesión |
| Sesión sin votos | `poker.ErrNoVotes` | La sesión sigue `ABIERTA` |
| Falla del repositorio al leer o guardar | Error del repositorio | La sesión no se modifica |

## Criterios de aceptación

- **CA-23.1** Al revelar una sesión abierta, se muestran a la vez los votos de todos los integrantes (integrante y puntos) y la sesión queda revelada.
- **CA-23.2** No se puede revelar una sesión sin votos: da error y la sesión sigue abierta.
- **CA-23.3** Después de revelar, los votos ya no se pueden cambiar ni agregar.
- **CA-23.4** Revelar una sesión inexistente o que no está abierta, incluida una ya revelada, da error y no la modifica.
- **CA-23.5** Si falla el guardado, se informa el error y la sesión sigue abierta con sus votos ocultos.
- **CA-23.6** Revelar solo cambia el estado: la ronda, los votos y su orden, el ítem y sus Story Points no cambian, y modificar lo devuelto no altera lo guardado.
- **CA-23.7** Si se vota mientras se revela, ningún voto aceptado se pierde: cada voto aparece entre los revelados o se rechaza.
- **CA-23.8** Se revela desde la API (`POST /api/poker-sessions/{id}/reveal`) y desde el botón "Revelar votos" de la interfaz web. La API responde `200` en el caso exitoso y `400`, `404`, `409`, `500` o `405` según la tabla de errores; la interfaz muestra la lista de votos o el mensaje de error.
- **CA-23.9** El estado revelado se guarda en memoria y en PostgreSQL con el mismo comportamiento, verificado por los tests de contrato (se omiten sin `TEST_DATABASE_URL`).

## Pruebas previstas

Todos los tests listados se escriben en la fase RED. Los obligatorios desde el RED son los de `404`, `409` no abierta, `409` sin votos, `500` por falla al guardar (con verificación de que la sesión no cambia), `400` por ID no numérico y la concurrencia.

### Unitarios (`internal/poker/reveal_test.go`)

| Test | CA |
|---|---|
| `TestReveal_SesionConVotosDevuelveTodosLosVotosYEstadoRevelada` | CA-23.1 |
| `TestReveal_SesionConUnSoloVoto` | CA-23.1 |
| `TestReveal_SesionSinVotosDevuelveErrNoVotes` (la sesión sigue `ABIERTA`) | CA-23.2 |
| `TestVote_DespuesDeRevelarDevuelveErrSessionNotOpen` (los votos guardados no cambian) | CA-23.3 |
| `TestReveal_SesionInexistenteDevuelveErrSessionNotFound` (ID inexistente, 0 y negativo) | CA-23.4 |
| `TestReveal_SesionYaReveladaDevuelveErrSessionNotOpen` | CA-23.4 |
| `TestReveal_ErrorAlLeerPropagaError` | CA-23.5 |
| `TestReveal_ErrorAlGuardarPropagaErrorYSesionNoCambia` (sigue `ABIERTA` con sus votos) | CA-23.5 |
| `TestReveal_ConservaRondaItemYOrdenDeLosVotos` | CA-23.6 |
| `TestReveal_NoModificaElItemNiSusStoryPoints` | CA-23.6 |
| `TestReveal_DevuelveCopiaDeLosVotos` | CA-23.6 |
| `TestReveal_ConcurrenteConVotosNoPierdeVotosAceptados` (con `-race`) | CA-23.7 |

El test de concurrencia parte de una sesión con un voto previo (para que la revelación no falle por `ErrNoVotes`), lanza votos de integrantes distintos y una revelación al mismo tiempo, y verifica que: la sesión queda `REVELADA`; cada voto que devolvió `nil` está entre los votos guardados; cada voto rechazado devolvió `ErrSessionNotOpen`; y la cantidad de votos guardados es el voto previo más los votos aceptados.

### Contrato de repositorio (`internal/poker/repository_contract_test.go`)

- Subtest nuevo: `Update guarda el estado REVELADA y GetByID lo devuelve con sus votos` (CA-23.9). Corre contra `MemoryRepository` y contra `PostgresRepository`; este último se omite sin `TEST_DATABASE_URL`.

### BDD (`features/HU-23-revelar-estimaciones.feature`)

Runner `TestFeatures_HU23RevelarEstimaciones` en `internal/poker/reveal_features_test.go`. **Todos los pasos llaman al `Manager` de `internal/poker`, no al handler HTTP**, por eso ningún escenario dice "mediante la API". Antecedentes: un proyecto de prueba, un ítem y una sesión abierta para ese ítem.

| Escenario | Tipo |
|---|---|
| Revelar una sesión con votos muestra todos los votos (CA-23.1) | Normal |
| Revelar una sesión con un único voto (CA-23.1) | Límite |
| Revelar conserva la ronda, el ítem y el orden de los votos (CA-23.6) | Alternativo |
| Revelar muestra el voto reemplazado de un integrante (CA-23.1) | Alternativo |
| Rechazar revelar una sesión sin votos (CA-23.2) | Error |
| No se puede votar después de revelar (CA-23.3) | Error |
| Rechazar revelar una sesión ya revelada (CA-23.4) | Error |
| Rechazar revelar una sesión inexistente (CA-23.4) | Error |
| Votar mientras se revela no pierde votos aceptados (CA-23.7) | Límite |

### API (`internal/httpapi/poker_test.go` y `server_test.go`)

| Test | CA |
|---|---|
| `TestRevelarPoker_Devuelve200ConTodosLosVotos` | CA-23.8 |
| `TestRevelarPoker_IDNoNumericoDevuelve400` | CA-23.8 |
| `TestRevelarPoker_InexistenteDevuelve404` | CA-23.8 |
| `TestRevelarPoker_SesionNoAbiertaDevuelve409` | CA-23.8 |
| `TestRevelarPoker_SinVotosDevuelve409` | CA-23.8 |
| `TestRevelarPoker_ErrorDelRepositorioDevuelve500SinDetalle` | CA-23.8 |
| `TestRevelarPoker_MetodoNoPermitidoDevuelve405` | CA-23.8 |
| `TestRevelarPoker_IgnoraElCuerpo` | CA-23.8 |
| Caso `"revelar Planning Poker"` en `TestRutasConID_IDNoNumericoDevuelve400` | CA-23.8 |

### Interfaz (`internal/httpapi/web_test.go`)

| Test | CA |
|---|---|
| `TestInterfaz_BotonRevelarVotos` (botón, contenedor de la lista de votos y ruta en `app.js`) | CA-23.8 |

## Comparación con el Planning del Sprint 3

El acta del Sprint 3 (`docs/actas/sprint-3/planning.md`) asigna a HU-23 **2 Story Points**, con dependencia de HU-22 y bloqueo de HU-24 y HU-25. La nota del Planning establece: “Cambia el estado de la sesión y devuelve todos los votos a la vez”. Esta SDD implementa ese alcance con RN-23.4 (cambio de estado a `REVELADA`) y RN-23.5 (todos los votos a la vez).

La especificación también precisa, sin ampliar el alcance funcional del acta:

- Una sesión sin votos no se puede revelar, como pide el criterio CA-23.2 del Issue.
- La revelación no es idempotente: una sesión ya revelada da `409`.
- Los votos posteriores se rechazan con la regla existente de HU-22.
- Revelar no modifica el ítem ni escribe Story Points, que quedan para HU-25.

No hay diferencias de alcance con el acta de Planning.

## Comparación con los criterios del Issue #41 (borrador)

| Criterio del Issue | En esta SDD |
|---|---|
| CA-23.1 Al revelar se muestran los votos de todos los integrantes. | CA-23.1, con el estado `REVELADA` |
| CA-23.2 No se puede revelar una sesión sin votos. | CA-23.2, con `ErrNoVotes` (409) |
| CA-23.3 Después de revelar, los votos ya no se pueden cambiar. | CA-23.3, también para votos nuevos |

CA-23.4 a CA-23.9 agregan los casos de error, la falla del repositorio, la concurrencia, la API, la interfaz y la persistencia, con el mismo criterio que HU-21 y HU-22. Los criterios del Issue siguen marcados como "(borrador)": actualizarlos según esta SDD requiere autorización y no se hace en este paso.

## Documentación a actualizar en el ciclo

- `README.md`:
  - Fila de HU-23 en "Historias por sprint", con Issue #41, SDD, BDD, evidencia y PR.
  - Fila del endpoint `POST /api/poker-sessions/{id}/reveal` en la tabla de la API.
  - Descripción de la sección **Planning Poker** de la interfaz: el botón "Revelar votos" y la lista de votos revelados (HU-23).
- `docs/reportes/cobertura.md`: cobertura de HU-23, indicando el commit sobre el que se midió.
- `docs/evidencias/HU-23-tdd.md`: evidencia de RED, GREEN, refactor y validación, con la tabla de commits.
- `docs/ia/registro.md`: registro del uso de IA en la historia.

La SDD va en su propio commit, aparte del commit RED, cuando se autorice.

HU-23: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
