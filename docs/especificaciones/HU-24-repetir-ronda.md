# HU-24 · Detectar diferencias y repetir ronda

> Issue del repositorio: #42

Aprobada por: Peñalbé Hernán (Agile Enabler), 29/10/2026, con 1 ajuste (el botón "Repetir ronda" es type="button" por estar dentro de form-estimar-poker), que se aplicó.

## Historia de usuario

Como Agile Enabler, quiero detectar si hay diferencias entre las estimaciones y abrir otra ronda, para llegar a un acuerdo.

## Objetivo

Indicar, al revelar los votos de una sesión de Planning Poker (HU-23), si todos los integrantes votaron lo mismo o si hay diferencias, informando el voto menor, el voto mayor y quiénes los emitieron. Y permitir repetir la ronda en la **misma** sesión: una sesión `REVELADA` vuelve a `ABIERTA`, con la ronda siguiente y sin votos, para que el equipo vote de nuevo hasta registrar la estimación acordada (HU-25).

Esta historia agrega la función pura `CompareVotes` y la operación `RepeatRound` al paquete `internal/poker`. La comparación se agrega a la respuesta de `POST /api/poker-sessions/{id}/reveal` como un campo nuevo, y la repetición se expone en `POST /api/poker-sessions/{id}/rounds` y en la sección **Planning Poker** de la interfaz web. Reutiliza `Repository.GetByID`, `Repository.OpenByItem` y `Repository.Update`, y los errores `ErrSessionNotFound`, `ErrSessionNotRevealed` y `ErrSessionAlreadyOpen`, sin operaciones nuevas en el repositorio, sin errores nuevos ni cambios de esquema.

## Alcance

Incluye:

- Comparar los votos revelados de una sesión: si hay diferencias, cuál es el voto menor y cuál el mayor, y qué integrantes emitieron cada uno.
- Devolver esa comparación junto con los votos al revelar (HU-23), sin cambiar `Manager.Reveal`.
- Repetir la ronda de una sesión `REVELADA` existente: pasa a `ABIERTA`, la ronda aumenta en uno y los votos de la ronda anterior se descartan.
- Validar la existencia de la sesión, su estado `REVELADA` y que no haya otra sesión `ABIERTA` para el mismo ítem.
- Guardar la sesión en memoria y PostgreSQL (TEC-05) con el `Update` existente.
- Exponer la repetición en la API y, en la interfaz web, mostrar la comparación y un botón para repetir la ronda.

No incluye:

- Calcular promedios, modas, medianas, consenso ni sugerir un valor acordado.
- Conservar el historial de los votos de rondas anteriores: se descartan (CA-24.2 del Issue).
- Limitar la cantidad de rondas.
- Consultar la comparación o los votos de una sesión después de revelarla, ni listar o consultar sesiones (igual que HU-23).
- Impedir abrir otra sesión para el mismo ítem mientras una está `REVELADA` (D10 de HU-23 y HU-25, sin cambios).
- Validar a los votantes contra la entidad integrante: E4 no está implementada.
- Restringir quién puede repetir la ronda: la aplicación no tiene autenticación ni roles.

## Contrato

### Entradas

| Operación | Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---:|---|
| `CompareVotes` | `votes` | `[]Vote` | Sí | Votos revelados de la ronda actual |
| `RepeatRound` | `SessionID` | entero | Sí | ID de la sesión de Planning Poker creada por HU-21 |

El estado, la ronda, el ítem y los votos de la sesión no son entradas de `RepeatRound`: se leen de la sesión guardada.

Firmas, en `internal/poker`:

```go
// compare.go
type VoteGroup struct {
	Points int
	Voters []string
}

type VoteComparison struct {
	HasDifferences bool
	Lowest         VoteGroup
	Highest        VoteGroup
}

func CompareVotes(votes []Vote) VoteComparison

// manager.go
func (m *Manager) RepeatRound(sessionID int) (Session, error)
```

### Salidas esperadas

- `CompareVotes` devuelve una `VoteComparison` nueva:
  - `HasDifferences` es `true` si hay al menos dos votos con puntos distintos y `false` si todos los votos tienen los mismos puntos.
  - `Lowest` y `Highest` tienen los puntos menor y mayor, y los nombres de los integrantes que los emitieron, en el orden de los votos.
  - Sin diferencias, `Lowest` y `Highest` son iguales: los mismos puntos y todos los integrantes.
  - Las listas de integrantes son nuevas: modificarlas no altera los votos recibidos.
- Si la repetición es válida, `RepeatRound` devuelve la sesión como copia por valor, con estado `ABIERTA`, la ronda anterior más uno y la lista de votos vacía (no `nil`). La API responde `200 OK`.
- Si la repetición no es válida, se devuelve el error correspondiente y la sesión no se modifica: sigue en su estado, su ronda y con sus votos.

### Contrato HTTP

#### Revelar con comparación (cambio compatible de HU-23)

```text
POST /api/poker-sessions/{id}/reveal
```

La respuesta de HU-23 se mantiene igual y suma el campo `comparacion`. Solo lo incluye esta ruta: las respuestas de abrir, votar, registrar la estimación y repetir la ronda no cambian.

```json
{
  "id": 1,
  "item_id": 5,
  "estado": "REVELADA",
  "ronda": 1,
  "votos": [
    { "integrante": "Ana", "puntos": 3 },
    { "integrante": "Bruno", "puntos": 8 },
    { "integrante": "Carla", "puntos": 8 }
  ],
  "cartas": [1, 2, 3, 5, 8, 13, 21],
  "comparacion": {
    "hay_diferencias": true,
    "menor": { "puntos": 3, "integrantes": ["Ana"] },
    "mayor": { "puntos": 8, "integrantes": ["Bruno", "Carla"] }
  }
}
```

Sin diferencias:

```json
"comparacion": {
  "hay_diferencias": false,
  "menor": { "puntos": 5, "integrantes": ["Ana", "Bruno"] },
  "mayor": { "puntos": 5, "integrantes": ["Ana", "Bruno"] }
}
```

Los códigos y errores de la revelación no cambian (tabla de HU-23).

#### Repetir la ronda

```text
POST /api/poker-sessions/{id}/rounds
```

Sin cuerpo. Si se envía uno, se ignora.

Respuesta exitosa, `200 OK`, con la misma estructura `pokerSessionResponse` de HU-21 a HU-25 y sin `comparacion`:

```json
{
  "id": 1,
  "item_id": 5,
  "estado": "ABIERTA",
  "ronda": 2,
  "votos": [],
  "cartas": [1, 2, 3, 5, 8, 13, 21]
}
```

Errores:

| Caso | Código | Mensaje |
|---|---:|---|
| `{id}` no es un entero | 400 | `errInvalidID` |
| Sesión inexistente (incluidos 0 y negativos) | 404 | `poker.ErrSessionNotFound` |
| Sesión `ABIERTA` o `CERRADA` | 409 | `poker.ErrSessionNotRevealed` |
| Hay otra sesión `ABIERTA` para el mismo ítem | 409 | `poker.ErrSessionAlreadyOpen` |
| Falla del repositorio al leer, buscar o guardar | 500 | `errInternal`, sin detalle |
| Método distinto de `POST` | 405 | `GET` con `methodNotAllowed`; los demás, respuesta de `net/http` |

El orden de validación es: ID de la ruta, existencia de la sesión, estado `REVELADA`, otra sesión abierta para el ítem y guardado.

### Interfaz web

En la sección **Planning Poker**, dentro del bloque que se muestra después de revelar (`form-estimar-poker`, de HU-25):

- Un párrafo `diferencias-poker` con el resultado de la comparación:
  - Con diferencias: `"Hay diferencias: menor 3 (Ana), mayor 8 (Bruno, Carla)."`
  - Sin diferencias: `"Todos votaron 5."`
- Un botón **"Repetir ronda"** (`btn-repetir-ronda-poker`), con `type="button"` porque está dentro de `<form id="form-estimar-poker">`: con el tipo por defecto (`submit`), el clic enviaría el formulario y recargaría la página. El botón llama a `POST /api/poker-sessions/{id}/rounds` con la sesión mostrada. Al repetir con éxito, se vuelve a mostrar la sesión con la función existente `mostrarPoker`, que ya muestra los datos (estado `ABIERTA` y la ronda nueva), los botones de voto, y oculta la lista de votos y el bloque posterior a la revelación. El mensaje común informa `"Ronda N iniciada en la sesión M."`.
- Si la API responde con error, el mensaje común muestra el mensaje de la API y la sección no cambia.

Como el párrafo y el botón están dentro de `form-estimar-poker`, se ocultan con él al abrir otra sesión (`mostrarPoker`) y al registrar la estimación (HU-25), sin modificar esas funciones. El único cambio sobre código existente de la interfaz es una llamada nueva, `mostrarDiferenciasPoker(s.comparacion)`, en el manejador de "Revelar votos". Sin librerías externas.

## Reglas de negocio

1. **RN-24.1 Qué es una diferencia.** Hay diferencia cuando los votos revelados de la ronda no tienen todos los mismos puntos. Con un único voto, o con todos los votos iguales, no hay diferencia.
2. **RN-24.2 Menor y mayor.** La comparación informa los puntos del voto menor y del mayor y, para cada uno, todos los integrantes que lo emitieron, en el orden en que se guardaron los votos. No informa otros valores ni calcula promedios.
3. **RN-24.3 Se compara al revelar.** La comparación se calcula sobre los votos que devuelve `Reveal` y se agrega a la respuesta de la revelación. `Manager.Reveal` no cambia. `CompareVotes` es una función pura: no lee ni escribe el repositorio.
4. **RN-24.4 Sesión existente.** Para repetir la ronda, la sesión debe existir. Si no existe, `RepeatRound` devuelve `ErrSessionNotFound` y no se modifica ninguna sesión. Esto incluye los IDs 0, negativos y repetir sin sesiones creadas.
5. **RN-24.5 Solo sesiones `REVELADA`.** Solo se puede repetir la ronda de una sesión `REVELADA`. Una sesión `ABIERTA` (todavía se vota) o `CERRADA` (ya tiene la estimación registrada, HU-25) devuelve `ErrSessionNotRevealed` y no se modifica.
6. **RN-24.6 Se puede repetir sin diferencias.** La repetición no exige que haya diferencias: la comparación es informativa y el equipo decide si vota de nuevo.
7. **RN-24.7 Una sola sesión abierta por ítem.** Si mientras la sesión estaba `REVELADA` se abrió otra sesión para el mismo ítem (permitido por D10 de HU-23 y HU-25), repetir la ronda devuelve `ErrSessionAlreadyOpen` y la sesión sigue `REVELADA`. Así se mantiene RN-21.2: nunca hay dos sesiones `ABIERTA` para el mismo ítem.
8. **RN-24.8 Nueva ronda en la misma sesión.** La sesión pasa de `REVELADA` a `ABIERTA`, `Round` aumenta en uno y la lista de votos queda vacía. El ID y el ítem no cambian, no se crea otra sesión, no se modifica el ítem y no se escriben Story Points. La ronda anterior y sus votos no se guardan.
9. **RN-24.9 La nueva ronda funciona como la primera.** En la ronda nueva se vota (HU-22, un voto por integrante en la ronda actual, RN-22.3), se revela (HU-23, con al menos un voto) y se registra la estimación (HU-25) con las reglas existentes, sin cambios.
10. **RN-24.10 Copias por valor.** `RepeatRound` devuelve una copia de la sesión con su propia lista de votos vacía, y `CompareVotes` devuelve listas de integrantes nuevas. Modificar lo devuelto no altera lo guardado.
11. **RN-24.11 Concurrencia.** La lectura, las validaciones (incluida la búsqueda de otra sesión abierta con `OpenByItem`), el cambio de estado y el guardado se realizan bajo el mismo `sync.Mutex` del `Manager` que usan `OpenSession`, `Vote`, `Reveal` y `Estimate`. Por eso: dos repeticiones simultáneas de la misma sesión dejan la ronda aumentada en uno (una tiene éxito y la otra devuelve `ErrSessionNotRevealed`); un voto simultáneo se rechaza con `ErrSessionNotOpen` o queda guardado en la ronda nueva; y una apertura simultánea para el mismo ítem nunca deja dos sesiones `ABIERTA`. `go test -race` no debe informar carreras.
12. **RN-24.12 Orden de validación.** Existencia de la sesión, estado `REVELADA`, otra sesión abierta para el ítem y guardado. Se informa el primer error encontrado.

## Decisión sobre entidades referenciadas

`RepeatRound` referencia por ID a una sesión de Planning Poker. **Sí se valida su existencia** con `Repository.GetByID`, igual que en HU-22, HU-23 y HU-25: la sesión es la raíz de la operación.

**No se valida nuevamente el ítem** (existencia ni estado `TERMINADO`): la sesión apunta a un ítem validado al abrirse (HU-21), repetir la ronda no lo modifica y HU-22 y HU-23 tampoco lo revalidan. Si el ítem se terminó mientras tanto, HU-25 ya rechaza registrar la estimación (`backlog.ErrItemFinished`), así que una ronda repetida no puede escribir Story Points en un ítem terminado. Sí se consultan las **otras sesiones del ítem** con `OpenByItem`, para mantener RN-21.2 (RN-24.7).

No se validan los votantes contra una entidad de integrantes: E4 no está implementada y el votante es un nombre libre (decisión 4 de la SDD de HU-21).

## Restricciones

Archivos que cambian:

- `internal/poker/compare.go` (nuevo): `VoteGroup`, `VoteComparison` y `CompareVotes`.
- `internal/poker/manager.go`: `Manager.RepeatRound`, junto a `Reveal` y `Estimate`.
- `internal/httpapi/poker.go`:
  - Campo nuevo `Comparacion *pokerComparisonResponse` con la etiqueta `json:"comparacion,omitempty"` en `pokerSessionResponse`, que solo completa `revealPokerSession`. Las demás rutas no lo completan y su JSON no cambia.
  - Handler nuevo `repeatRoundPokerSession`.
- `internal/httpapi/server.go`: rutas `GET /api/poker-sessions/{id}/rounds` (con `methodNotAllowed`) y `POST /api/poker-sessions/{id}/rounds`.
- `internal/web/index.html` y `internal/web/app.js`: el párrafo, el botón, la función `mostrarDiferenciasPoker` y su llamada en el manejador de "Revelar votos".

No cambian: `session.go` (no hay estados ni errores nuevos), `Repository`, `MemoryRepository`, `PostgresRepository`, `create.go`, `Manager.Reveal`, `statusFor` (`ErrSessionNotRevealed` y `ErrSessionAlreadyOpen` ya responden `409`) ni el esquema de `db/init/`: `Update` ya guarda `status`, `round_number` y `votes`, y la columna `status` no tiene restricciones.

Otras restricciones:

- **Compatibilidad del cambio en revelar:** el campo nuevo es aditivo. Los tests existentes de HU-23 decodifican con `json.Unmarshal` (helper `decode` de `server_test.go`), que ignora campos desconocidos, y la interfaz solo lee los campos que ya usaba; ningún test compara el cuerpo como texto. Los tests de HU-23 deben seguir pasando sin cambios.
- **Cuerpo del pedido de repetir:** la ruta no recibe cuerpo. El handler no lo lee y un cuerpo enviado se ignora, igual que en `/reveal`; la ruta **no** se agrega a `bodyEndpoints`.
- **Método no permitido:** se registra `GET /api/poker-sessions/{id}/rounds` con `methodNotAllowed`, igual que `/reveal` y `/estimate`, para que no caiga en el servidor de archivos de la interfaz. Los demás métodos los rechaza `net/http` con `405`.
- La ruta se agrega a `TestRutasConID_IDNoNumericoDevuelve400` de `server_test.go`.
- Un error interno (`500`) responde `errInternal`, sin el detalle del error del repositorio.
- Solo se usa la librería estándar en producción; godog se utiliza únicamente en pruebas.
- RED: si el paquete no compila por la funcionalidad nueva, se agregan primero las firmas vacías (stubs) de `CompareVotes`, `VoteComparison`, `VoteGroup` y `RepeatRound`, para que el RED falle por las aserciones (Definition of Done). `go vet ./...` antes del commit RED.

### Pruebas previstas

Todos los tests listados se escriben en la fase RED.

Unitarios de la comparación (`internal/poker/compare_test.go`):

| Test | CA |
|---|---|
| `TestCompareVotes_VotosDistintosInformaDiferenciaMenorYMayor` | CA-24.1 |
| `TestCompareVotes_TodosIgualesNoInformaDiferencia` | CA-24.1 |
| `TestCompareVotes_UnSoloVotoNoInformaDiferencia` | CA-24.1 |
| `TestCompareVotes_VariosIntegrantesConElMismoExtremo` (en el orden de los votos) | CA-24.1 |
| `TestCompareVotes_SinVotosDevuelveComparacionVacia` | CA-24.1 |
| `TestCompareVotes_DevuelveListasNuevas` | CA-24.6 |

Unitarios de la repetición (`internal/poker/round_test.go`):

| Test | CA |
|---|---|
| `TestRepeatRound_SesionReveladaVuelveAAbiertaConRondaSiguienteYSinVotos` | CA-24.2, CA-24.3 |
| `TestRepeatRound_ConservaIDEItemYNoModificaStoryPoints` | CA-24.3 |
| `TestRepeatRound_SinDiferenciasTambienSePuedeRepetir` | CA-24.4 |
| `TestRepeatRound_NuevaRondaPermiteVotarRevelarYEstimar` (incluye la ronda 3) | CA-24.5 |
| `TestRepeatRound_SesionInexistenteDevuelveErrSessionNotFound` (inexistente, 0 y negativo) | CA-24.7 |
| `TestRepeatRound_SesionAbiertaDevuelveErrSessionNotRevealed` | CA-24.7 |
| `TestRepeatRound_SesionCerradaDevuelveErrSessionNotRevealed` | CA-24.7 |
| `TestRepeatRound_OtraSesionAbiertaDelItemDevuelveErrSessionAlreadyOpen` (la sesión sigue `REVELADA`) | CA-24.8 |
| `TestRepeatRound_ErrorAlLeerPropagaError` | CA-24.9 |
| `TestRepeatRound_ErrorAlBuscarSesionAbiertaPropagaError` | CA-24.9 |
| `TestRepeatRound_ErrorAlGuardarPropagaErrorYSesionNoCambia` (sigue `REVELADA`, misma ronda y votos) | CA-24.9 |
| `TestRepeatRound_DevuelveCopia` | CA-24.6 |
| `TestRepeatRound_ConcurrenteAumentaLaRondaUnaSolaVez` (con `-race`) | CA-24.10 |
| `TestRepeatRound_ConcurrenteConVotosNoMezclaRondas` (con `-race`) | CA-24.10 |
| `TestRepeatRound_ConcurrenteConAperturaDejaUnaSolaSesionAbierta` (con `-race`) | CA-24.10 |

Contrato de repositorio (`internal/poker/repository_contract_test.go`): subtest nuevo `Update guarda una ronda nueva ABIERTA sin votos y OpenByItem la encuentra` (CA-24.11), contra `MemoryRepository` y `PostgresRepository` (este último se omite sin `TEST_DATABASE_URL`).

BDD (`features/HU-24-repetir-ronda.feature`): runner `TestFeatures_HU24RepetirRonda` en `internal/poker/round_features_test.go`. Los pasos llaman al `Manager` y a `CompareVotes` de `internal/poker`, no al handler HTTP. Escenarios previstos:

| Escenario | Tipo |
|---|---|
| Revelar votos distintos informa la diferencia con el menor y el mayor (CA-24.1) | Normal |
| Repetir la ronda de una sesión revelada (CA-24.2, CA-24.3) | Normal |
| Revelar votos iguales informa que no hay diferencias (CA-24.1) | Alternativo |
| Repetir la ronda sin diferencias (CA-24.4) | Alternativo |
| Votar, revelar y registrar la estimación en la ronda nueva (CA-24.5) | Alternativo |
| Revelar un único voto no informa diferencias (CA-24.1) | Límite |
| Repetir la ronda dos veces seguidas llega a la ronda 3 (CA-24.3) | Límite |
| Rechazar repetir la ronda de una sesión abierta (CA-24.7) | Error |
| Rechazar repetir la ronda de una sesión cerrada (CA-24.7) | Error |
| Rechazar repetir la ronda de una sesión inexistente (CA-24.7) | Error |
| Rechazar repetir la ronda si hay otra sesión abierta para el ítem (CA-24.8) | Error |

API (`internal/httpapi/poker_test.go` y `server_test.go`):

| Test | CA |
|---|---|
| `TestRevelarPoker_IncluyeComparacionConDiferencias` | CA-24.1, CA-24.12 |
| `TestRevelarPoker_IncluyeComparacionSinDiferencias` | CA-24.1, CA-24.12 |
| `TestVotarPoker_RespuestaSinComparacion` (abrir, votar y estimar no incluyen `comparacion`) | CA-24.12 |
| `TestRepetirRondaPoker_Devuelve200ConRondaSiguienteYSinVotos` | CA-24.12 |
| `TestRepetirRondaPoker_IDNoNumericoDevuelve400` | CA-24.12 |
| `TestRepetirRondaPoker_InexistenteDevuelve404` | CA-24.12 |
| `TestRepetirRondaPoker_SesionNoReveladaDevuelve409` | CA-24.12 |
| `TestRepetirRondaPoker_OtraSesionAbiertaDevuelve409` | CA-24.12 |
| `TestRepetirRondaPoker_ErrorDelRepositorioDevuelve500SinDetalle` | CA-24.12 |
| `TestRepetirRondaPoker_MetodoNoPermitidoDevuelve405` | CA-24.12 |
| `TestRepetirRondaPoker_IgnoraElCuerpo` | CA-24.12 |
| Caso `"repetir ronda Planning Poker"` en `TestRutasConID_IDNoNumericoDevuelve400` | CA-24.12 |

Interfaz (`internal/httpapi/web_test.go`): `TestInterfaz_DiferenciasYBotonRepetirRonda` (el párrafo y el botón dentro de `form-estimar-poker`, el botón con `type="button"`, y `app.js` usa la ruta `/rounds` y el campo `comparacion`) (CA-24.12).

### Documentación a actualizar en el ciclo

- `README.md`: fila de HU-24 en "Historias por sprint", fila del endpoint `POST /api/poker-sessions/{id}/rounds`, el campo `comparacion` en la fila de `/reveal` y la descripción de la sección **Planning Poker**.
- `docs/reportes/cobertura.md`, `docs/evidencias/HU-24-tdd.md` y `docs/ia/registro.md`.

La SDD va en su propio commit, aparte del commit RED, cuando se autorice.

## Límites conocidos

- **La comparación solo se ve al revelar.** No hay consulta de sesiones (fuera de alcance desde HU-21): si se recarga la página después de revelar, la comparación y los votos ya no se muestran, igual que en HU-23.
- **Los votos de rondas anteriores se pierden.** La sesión guarda solo la ronda actual (`Session.Votes`, RN-22.3); no hay historial de rondas.
- **D10 de HU-23 y HU-25 se mantiene.** Mientras una sesión está `REVELADA` se puede abrir otra para el mismo ítem. En ese caso, la sesión revelada no puede repetir su ronda (RN-24.7) y solo puede registrar su estimación (HU-25) o quedar como está.
- **Sin límite de rondas.** `Round` es un entero y no se fija un máximo.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Votos 3, 8 y 8 | Hay diferencias; menor 3 (Ana), mayor 8 (Bruno, Carla) |
| Todos los votos iguales | Sin diferencias; menor y mayor iguales, con todos los integrantes |
| Un único voto | Sin diferencias; menor y mayor son ese voto |
| Varios integrantes con el voto menor (o mayor) | Se informan todos, en el orden de los votos |
| `CompareVotes` sin votos (no ocurre al revelar, por `ErrNoVotes`) | Sin diferencias; puntos 0 y listas de integrantes vacías, no `nil` |
| Voto reemplazado antes de revelar (HU-22) | Se compara el último voto del integrante |
| Repetir la ronda de una sesión revelada en la ronda 1 | `ABIERTA`, ronda 2, sin votos; mismo ID e ítem |
| Repetir dos veces (revelando entre medio) | Ronda 3 |
| Repetir sin diferencias | Se permite (RN-24.6) |
| Revelar la ronda nueva sin votos | `ErrNoVotes` (HU-23); sigue `ABIERTA` en la ronda nueva |
| Votar en la ronda nueva con el nombre de un votante de la ronda anterior | Voto nuevo de la ronda actual, sin rastro del anterior |
| Repetir una sesión `ABIERTA` o `CERRADA` | `ErrSessionNotRevealed`; no cambia |
| Sesión inexistente, ID 0 o negativo | `ErrSessionNotFound` |
| Otra sesión `ABIERTA` para el mismo ítem | `ErrSessionAlreadyOpen`; la sesión sigue `REVELADA` |
| Ítem de la sesión `TERMINADO` | Se permite repetir; HU-25 rechaza después registrar la estimación (`backlog.ErrItemFinished`) |
| Dos repeticiones simultáneas | La ronda aumenta una sola vez; la otra devuelve `ErrSessionNotRevealed` |
| Voto simultáneo con la repetición | Se rechaza con `ErrSessionNotOpen` o queda en la ronda nueva |
| Apertura simultánea para el mismo ítem | Queda una sola sesión `ABIERTA` |
| Modificar la sesión o la comparación devueltas | Lo guardado no cambia |
| Falla del repositorio al leer, buscar o guardar | Se propaga el error; la sesión sigue `REVELADA` con su ronda y sus votos |
| Pedido HTTP con cuerpo | Se ignora |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Sesión inexistente | `poker.ErrSessionNotFound` | No se modifica ninguna sesión |
| Sesión `ABIERTA` o `CERRADA` | `poker.ErrSessionNotRevealed` | No se modifica la sesión |
| Otra sesión `ABIERTA` para el ítem | `poker.ErrSessionAlreadyOpen` | La sesión sigue `REVELADA` |
| Falla del repositorio al leer, buscar o guardar | Error del repositorio | La sesión no se modifica |

`CompareVotes` no devuelve errores.

## Criterios de aceptación

- **CA-24.1** Al revelar los votos se indica si todos coinciden o si hay diferencias; con diferencias, se informa el voto menor y el mayor y quiénes los emitieron.
- **CA-24.2** Se puede iniciar una nueva ronda en la misma sesión revelada, que descarta los votos de la ronda anterior y vuelve a aceptar votos.
- **CA-24.3** La sesión lleva el número de ronda: cada ronda nueva aumenta en uno, y el ID y el ítem de la sesión no cambian.
- **CA-24.4** Se puede repetir la ronda aunque no haya diferencias.
- **CA-24.5** En la ronda nueva se vota, se revela y se registra la estimación acordada como en la primera ronda.
- **CA-24.6** Repetir la ronda y comparar no modifican el ítem ni sus Story Points, y modificar lo devuelto no altera lo guardado.
- **CA-24.7** Repetir la ronda de una sesión inexistente, abierta o cerrada da error y no la modifica.
- **CA-24.8** No se puede repetir la ronda si ya hay otra sesión abierta para el mismo ítem: da error y la sesión sigue revelada.
- **CA-24.9** Si falla el guardado, se informa el error y la sesión sigue revelada con su ronda y sus votos.
- **CA-24.10** Si se repite la ronda al mismo tiempo que se vota, se repite o se abre otra sesión, la ronda aumenta una sola vez, ningún voto queda en la ronda equivocada y nunca hay dos sesiones abiertas para el mismo ítem.
- **CA-24.11** La ronda nueva se guarda en memoria y en PostgreSQL con el mismo comportamiento, verificado por los tests de contrato (se omiten sin `TEST_DATABASE_URL`).
- **CA-24.12** La comparación se ve en la respuesta de revelar y en la interfaz web; la ronda se repite desde la API (`POST /api/poker-sessions/{id}/rounds`) y desde el botón "Repetir ronda". La API responde `200` en el caso exitoso y `400`, `404`, `409`, `500` o `405` según la tabla de errores.

## Comparación con el acta de Planning

El acta del Sprint 3 (`docs/actas/sprint-3/planning.md`) asigna a HU-24 **3 Story Points**, "si hay tiempo", con la nota “Nueva ronda en la misma sesión (HU-22 y HU-23)” y dependencias de HU-21 y HU-23. El acta del Sprint 4 (`docs/actas/sprint-4/planning.md`) la mantiene en "si hay tiempo", en el orden HU-41, HU-24 y HU-14, para quien termine primero; el Issue #42 la asigna a Cardozo Leandro.

Esta SDD implementa ese alcance: la ronda nueva es en la misma sesión (RN-24.8) y la comparación se calcula sobre los votos que revela HU-23 (RN-24.3). La detección de diferencias no figura en la nota del acta pero sí en el título de la historia y en el Issue; no es una ampliación de alcance. HU-25, que no figura como dependencia en el acta, ya está en `main` y su estado `CERRADA` y su error `ErrSessionNotRevealed` se reutilizan.

No hay diferencias de alcance con las actas de Planning.

## Comparación con los criterios del Issue #42 (borrador)

| Criterio del Issue | En esta SDD | Observación |
|---|---|---|
| CA-24.1 Se indica si todos los votos coinciden o hay diferencias. | CA-24.1 | Se precisa que se informan el menor, el mayor y quiénes los emitieron |
| CA-24.2 Se puede iniciar una nueva ronda, que descarta los votos anteriores. | CA-24.2 | Solo desde una sesión `REVELADA`, en la misma sesión |
| CA-24.3 Se conserva el número de ronda. | CA-24.3 | **Redacción ambigua:** leído al pie de la letra ("la ronda no cambia") contradice iniciar una ronda nueva. Se interpreta como que la sesión lleva el número de ronda y lo aumenta en uno en cada repetición, sin volver a 1. No se conserva el historial de rondas anteriores (lo descarta CA-24.2). La interpretación quedó confirmada por el Agile Enabler con la aprobación de esta SDD (29/10/2026) |

CA-24.4 a CA-24.12 agregan la repetición sin diferencias, el ciclo completo en la ronda nueva, los errores, el caso de otra sesión abierta para el ítem (RN-24.7, que el Issue no contempla porque surge de D10 de HU-23 y HU-25), la falla del repositorio, la concurrencia, la persistencia, la API y la interfaz, con el mismo criterio que HU-21 a HU-25. Los criterios del Issue siguen marcados como "(borrador)": actualizarlos según esta SDD, incluida la redacción de CA-24.3, requiere autorización y no se hace en este paso.

## Decisiones tomadas en la redacción

| # | Decisión | Alternativa descartada | Motivo |
|---|---|---|---|
| D1 | Hay diferencia si los votos revelados no son todos iguales; se informan el menor y el mayor con sus integrantes (RN-24.1 y RN-24.2). | Informar también promedio, moda o distancia entre cartas; o considerar "diferencia" solo a partir de cierta distancia en la escala. | El Issue solo pide indicar si coinciden o hay diferencias. El menor y el mayor (y quiénes votaron) son lo que el equipo necesita para discutir, y no requieren cálculos ni métricas nuevas. Un umbral sería una regla que nadie pidió. |
| D2 | La comparación se agrega como campo nuevo `comparacion` en la respuesta de `POST /reveal`, calculada en el handler con la función pura `CompareVotes`; `Manager.Reveal` no cambia. | (a) Una consulta aparte `GET /api/poker-sessions/{id}/comparison`; (b) calcularla dentro de la operación de repetir; (c) calcularla solo en `app.js`. | Es lo que menos cambia lo mergeado: no toca el `Manager` ni el repositorio, el campo es aditivo y compatible (los tests de HU-23 decodifican con `json.Unmarshal` y la interfaz ignora campos que no usa). (a) agrega una consulta de votos posterior a la revelación, fuera de alcance desde HU-23; (b) llega tarde, porque la diferencia hace falta para decidir si se repite; (c) deja la regla sin tests en Go. |
| D3 | `CompareVotes` devuelve el menor y el mayor también sin diferencias (iguales, con todos los integrantes). | Devolver solo `HasDifferences` cuando no hay diferencias, o el menor y el mayor vacíos. | Una sola forma de respuesta, sin casos especiales en la API ni en la interfaz, y permite mostrar "Todos votaron 5". |
| D4 | `RepeatRound(sessionID int) (Session, error)` en `manager.go`, bajo el mismo `sync.Mutex` del `Manager`: leer, validar, buscar otra sesión abierta, cambiar y guardar dentro del lock. | Un mutex propio o validar fuera del lock. | Con un lock distinto, un voto o una apertura simultánea podrían intercalarse y perder la ronda, mezclar votos de rondas o dejar dos sesiones abiertas (RN-24.11), igual que en HU-22, HU-23 y HU-25. |
| D5 | Solo se repite una sesión `REVELADA`; `ABIERTA` o `CERRADA` devuelve `ErrSessionNotRevealed` (409), reutilizado de HU-25. Inexistente: `ErrSessionNotFound` (404). | Un error nuevo (`ErrCannotRepeatRound`), o permitir repetir una sesión `ABIERTA` para borrar sus votos. | El error de HU-25 describe exactamente la condición ("no está revelada") y ya responde 409. Repetir una sesión abierta borraría votos ocultos sin haberlos comparado; reabrir una `CERRADA` está fuera de alcance (HU-25). |
| D6 | Se permite repetir aunque no haya diferencias (RN-24.6). | Permitir repetir solo con diferencias. | El Issue no lo condiciona y la comparación es informativa: el equipo puede querer votar de nuevo tras discutir o porque faltó un integrante. Exigirlo requeriría un error nuevo y bloquearía casos válidos; HU-25 tampoco exige que el valor acordado coincida con los votos. |
| D7 | Repetir pasa a `ABIERTA`, `Round + 1` y votos vacíos en la **misma** sesión; la ronda anterior no se guarda. | Crear una sesión nueva con `OpenSession` para la ronda 2, o guardar el historial de rondas. | La nota del Planning pide "nueva ronda en la misma sesión" y CA-24.2 del Issue pide descartar los votos anteriores. Un historial necesitaría cambiar el esquema, que no se pide. |
| D8 | Si hay otra sesión `ABIERTA` para el ítem, repetir devuelve `ErrSessionAlreadyOpen` (409), comprobado con `OpenByItem` bajo el mutex (RN-24.7). Se mantiene D10 de HU-23 y HU-25: abrir otra sesión con una `REVELADA` sigue permitido. | Permitir repetir igual (quedarían dos sesiones `ABIERTA` para el ítem), o impedir abrir sesiones mientras haya una `REVELADA`. | Repetir vuelve a `ABIERTA` la sesión, y con D10 eso podría dejar dos sesiones abiertas para el mismo ítem, rompiendo RN-21.2. Impedir abrir sesiones cambiaría HU-21, ya mergeada. Con esta regla, `OpenByItem` sigue devolviendo a lo sumo una sesión `ABIERTA` por ítem: la sesión repetida pasa a ser la abierta del ítem y una apertura nueva devuelve `ErrSessionAlreadyOpen`. |
| D9 | No se revalida el ítem al repetir (ni existencia ni `TERMINADO`). | Rechazar con `backlog.ErrItemFinished` si el ítem se terminó. | Igual que HU-22 y HU-23; repetir no modifica el ítem y HU-25 ya impide escribir Story Points en un ítem terminado. Revalidar agregaría una llamada a `backlog` sin una regla pedida. |
| D10 | API: `POST /api/poker-sessions/{id}/rounds`, sin cuerpo, `200` con `pokerSessionResponse` (`ABIERTA`, ronda nueva, `votos` vacío, sin `comparacion`); `GET` con `methodNotAllowed`. | `PATCH /api/poker-sessions/{id}` con `{"estado":"ABIERTA"}`, o `POST .../repeat`. | Una transición con reglas propias, como `/reveal` y `/estimate`. `POST /rounds` sigue el estilo de `POST /votes`: crea la ronda siguiente. Es `200` y no `201` porque no se crea un recurso nuevo con su propio ID: cambia la misma sesión. |
| D11 | Interfaz: el párrafo de la comparación y el botón "Repetir ronda" van dentro de `form-estimar-poker`; al repetir se reutiliza `mostrarPoker` sin modificarla. El botón es `type="button"`, para que el clic no envíe el formulario ni recargue la página. El único cambio en código existente es la llamada a `mostrarDiferenciasPoker` en el manejador de "Revelar votos". | Un bloque nuevo fuera de `form-estimar-poker`. | Dentro de ese bloque, `mostrarPoker` y el registro de la estimación (HU-25) ya los ocultan cuando corresponde, sin tocar esas funciones. Fuera, habría que modificarlas para ocultarlos. Los botones de cartas de ese formulario ya son `type="button"` (`mostrarEstimarPoker`), y "Repetir ronda" sigue el mismo criterio. |
| D12 | Sin cambios en el repositorio ni en el esquema: `Update` ya guarda `status`, `round_number` y `votes` en memoria y PostgreSQL. Se agrega un subtest de contrato para la ronda nueva. | Una operación específica `Repository.NewRound`. | Mismo contrato de copias que usan HU-22, HU-23 y HU-25. |
| D13 | Tests con `-race` de repeticiones simultáneas, repetición con votos y repetición con apertura del mismo ítem. | Probar solo la repetición secuencial. | Son las condiciones de carrera propias de esta historia (RN-24.11). |

HU-24: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
