# HU-22 · Votar de forma oculta

> Issue del repositorio: #40

Aprobada por: Peñalbé Hernán (Agile Enabler), 25/10/2026, con los ajustes del código único de respuesta, el nombre del votante al reemplazar, GetByID en PostgreSQL y los tests de rutas.

## Historia de usuario

Como Product Builder, quiero votar mi estimación sin que los demás la vean, para no influirnos entre nosotros.

## Objetivo

Registrar el voto oculto de un integrante en una sesión de Planning Poker `ABIERTA` (HU-21), validando que la carta pertenezca a la escala Fibonacci de HU-20 y que haya un solo voto por integrante y por ronda, sin exponer ningún voto hasta que se revelen (HU-23).

Esta historia extiende el paquete `internal/poker` con la operación de voto, agrega al `Repository` las operaciones `GetByID` y `Update` previstas por HU-21, y expone el voto en `POST /api/poker-sessions/{id}/votes` y en la interfaz web.

## Alcance

Incluye:

- Emitir un voto sobre una sesión existente mediante el `Manager`.
- Validar el votante, la carta, la existencia de la sesión y su estado `ABIERTA`.
- Reemplazar el voto de un integrante que ya votó en la ronda actual.
- Mantener los votos invisibles en todas las salidas públicas.
- Guardar los votos en memoria y PostgreSQL (TEC-05).
- Exponer la operación en la API y en la interfaz web.

No incluye:

- Revelar los votos (HU-23).
- Repetir la ronda (HU-24).
- Registrar la estimación acordada ni cerrar la sesión (HU-25).
- Listar o consultar sesiones.
- Validar al votante contra la entidad integrante: E4 no está implementada.
- Modificar el ítem ni asignar Story Points (HU-20 y HU-25).

## Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---:|---|
| `SessionID` | entero | Sí | ID de la sesión de Planning Poker creada por HU-21 |
| `Voter` | texto | Sí | Nombre libre del integrante; se recortan los espacios extremos y no puede quedar vacío |
| `Points` | entero | Sí | Carta elegida: 1, 2, 3, 5, 8, 13 o 21 |

El estado, la ronda y el ID del voto no son entradas: los administra el sistema.

Firma:

```go
Vote(sessionID int, voter string, points int) (Session, error)
```

en `internal/poker/manager.go`.

## Salidas esperadas

- Si el voto es nuevo, se devuelve la sesión como copia por valor, con los votos ocultos. La API responde `200 OK`.
- Si el integrante ya había votado en la ronda actual, se reemplaza su voto y se devuelve la sesión como copia por valor, también con los votos ocultos. La API responde `200 OK`, igual que para un voto nuevo, para no revelar si ya existía un voto.
- Si el voto no es válido, se devuelve el error correspondiente y la sesión no se modifica.

La sesión devuelta siempre contiene `Votes` vacío y no `nil` mientras los votos no hayan sido revelados por HU-23.

## Reglas de negocio

1. **RN-22.1 Votante no vacío.** El nombre, después de recortar los espacios extremos, no puede quedar vacío. En ese caso se devuelve `ErrEmptyVoter` y no se registra el voto. La comparación entre votantes se realiza sin distinguir mayúsculas mediante `strings.EqualFold`; por lo tanto, `"Ana"` y `"ana"` representan al mismo votante.
2. **RN-22.2 Carta de la escala.** Los puntos deben pertenecer a `Cards()`: 1, 2, 3, 5, 8, 13 o 21. Si no pertenecen, se devuelve `ErrInvalidCard` y no se registra el voto.
3. **RN-22.3 Un voto por integrante y por ronda.** `Session.Votes` contiene los votos de la ronda actual. Si el votante ya tiene un voto, el nuevo reemplaza al anterior sin duplicarlo y conserva el nombre del último voto, recortado tal como llegó. Si no tiene voto, se agrega. HU-24 reiniciará la lista al iniciar una ronda nueva.
4. **RN-22.4 Sesión existente.** La sesión debe existir. Si no existe, se devuelve `ErrSessionNotFound` y no se modifica ninguna sesión. Esto incluye los IDs 0, negativos y votar sin sesiones creadas.
5. **RN-22.5 Solo sesiones `ABIERTA`.** Solo una sesión `ABIERTA` acepta votos. Si tiene otro estado, se devuelve `ErrSessionNotOpen` y no se registra el voto. La regla queda preparada para los estados que incorporen HU-23 y HU-25.
6. **RN-22.6 Copias por valor.** `Vote` devuelve una copia de la sesión y una lista nueva de votos. El repositorio y el `Manager` nunca devuelven referencias al estado interno. Modificar la sesión devuelta no altera la sesión guardada.
7. **RN-22.7 Concurrencia.** La lectura, la fusión del voto y el guardado se realizan bajo el `sync.Mutex` del `Manager`. Los votos simultáneos de integrantes distintos deben conservarse todos y `go test -race` no debe informar carreras.
8. **RN-22.8 Votos ocultos.** Ninguna operación pública de esta historia devuelve votos. La sesión devuelta siempre tiene `Votes` vacío.
9. **RN-22.9 Solo se vota.** Esta historia no modifica el ítem, el estado ni la ronda de la sesión, y no escribe Story Points. La existencia del ítem no se vuelve a validar porque ya fue validada al abrir la sesión en HU-21.

## Decisión sobre entidades referenciadas

HU-22 referencia por ID a una sesión de Planning Poker. **Sí se valida su existencia**, porque un voto sin sesión quedaría huérfano y la sesión es la raíz de esta operación. La búsqueda se incorpora mediante `Repository.GetByID` y podrá reutilizarse en HU-23, HU-24 y HU-25.

No se valida el votante contra una entidad de integrantes: E4 no está comprometida en esta entrega y el votante es un nombre libre, según la decisión documentada en la SDD de HU-21.

No se valida nuevamente el ítem ni el proyecto: la sesión apunta a un ítem validado al abrirse y esta entrega no contempla el borrado de ítems ni sesiones.

## Restricciones

- `Manager.Vote` se implementa en `internal/poker/manager.go`.
- Los errores `ErrSessionNotFound`, `ErrSessionNotOpen`, `ErrEmptyVoter` y `ErrInvalidCard` se declaran en `internal/poker/session.go`.
- `Repository` agrega:
  - `GetByID(id int) (Session, error)`, que devuelve `ErrSessionNotFound` si no existe.
  - `Update(session Session) (Session, error)`, que reemplaza la sesión por ID y devuelve una copia.
- `MemoryRepository` y `PostgresRepository` implementan ambas operaciones con el mismo contrato de copias.
- `PostgresRepository.GetByID` traduce `pgx.ErrNoRows` a `ErrSessionNotFound`, igual que `OpenByItem`.
- No se modifica `create.go` ni código de producción de otros paquetes fuera del cableado necesario para la API y la interfaz; `manager.go` y `session.go` contienen la implementación y los errores definidos para esta historia.
- La tabla `poker_sessions` ya contiene `votes` como `JSONB`; no se modifica el esquema. `Update` persiste el estado, la ronda y los votos codificados con `encoding/json`.
- Las pruebas de integración PostgreSQL se omiten cuando no existe `TEST_DATABASE_URL`.
- El handler se incorpora en `internal/httpapi/poker.go` y la ruta en `server.go`. Se reutiliza la respuesta de sesión de poker.
- El endpoint responde siempre `200 OK`, tanto para un voto nuevo como para un reemplazo, para preservar el ocultamiento.
- Solo se usa la librería estándar en producción; godog se utiliza únicamente en pruebas.
- Los tests de contrato de `GetByID` y `Update` se escriben en la fase RED junto con el resto de las pruebas.
- La nueva ruta se agrega a `bodyEndpoints` y a `TestRutasConID_IDNoNumericoDevuelve400`; la validación del cuerpo mayor de 1 MiB y del texto posterior al JSON se conserva en esos tests.

## Contrato HTTP

### Endpoint

```text
POST /api/poker-sessions/{id}/votes
```

### Cuerpo

```json
{
  "integrante": "Ana",
  "puntos": 5
}
```

### Respuesta exitosa

```json
{
  "id": 1,
  "item_id": 5,
  "estado": "ABIERTA",
  "ronda": 1,
  "votos": [],
  "cartas": [1, 2, 3, 5, 8, 13, 21]
}
```

La respuesta HTTP es siempre `200 OK`, tanto cuando el voto es nuevo como cuando reemplaza uno existente. El código único evita inferir si el integrante ya había votado.

### Errores

| Caso | Código | Mensaje |
|---|---:|---|
| `{id}` no es un entero | 400 | `errInvalidID` |
| Cuerpo no válido como JSON | 400 | `errInvalidJSON` |
| Votante vacío | 400 | `poker.ErrEmptyVoter` |
| Carta fuera de la escala | 400 | `poker.ErrInvalidCard` |
| Sesión inexistente | 404 | `poker.ErrSessionNotFound` |
| Sesión no `ABIERTA` | 409 | `poker.ErrSessionNotOpen` |
| Cuerpo mayor de 1 MiB | 413 | `errBodyTooLarge` |
| Falla del repositorio | 500 | `errInternal`, sin detalle |
| Método distinto de `POST` | 405 | Respuesta de `net/http` |

El orden de validación es: ID de la ruta, JSON, votante, carta, existencia de la sesión, estado y guardado. Se informa el primer error encontrado.

## Interfaz web

En la sección **Planning Poker**, cuando hay una sesión abierta, se muestran:

- Un campo para el nombre del integrante.
- Botones para las cartas 1, 2, 3, 5, 8, 13 y 21.
- El resultado mediante el mecanismo común: `"Voto registrado en la sesión N."` o el mensaje de error de la API.

La sección no muestra votos, ni siquiera el voto propio; solo mantiene visibles el estado y la ronda. La página debe contener el formulario y `app.js` debe usar la nueva ruta.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Primer voto de la sesión | Se agrega y la sesión devuelta muestra `Votes` vacío |
| Segundo voto del mismo votante | Reemplaza al anterior sin duplicarlo |
| Votante con distinta mayúscula | `"ana"` reemplaza el voto de `"Ana"` |
| Reemplazo con espacios extremos o distinta capitalización | Se reemplaza el voto y se guarda el nombre del último voto, recortado tal como llegó |
| Nombre con espacios extremos | Se recorta; si queda vacío, `ErrEmptyVoter` |
| Carta válida | Se registra |
| Carta 0, negativa o fuera de la escala | `ErrInvalidCard` |
| Sesión inexistente, ID 0 o negativo | `ErrSessionNotFound` |
| Sesión con estado distinto de `ABIERTA` | `ErrSessionNotOpen` |
| Votos simultáneos de integrantes distintos | Se registran ambos sin carreras |
| Muchos votos simultáneos del mismo votante | Queda un único voto final |
| Modificar la sesión o los votos devueltos | La sesión guardada no cambia |
| Falla del repositorio al leer o guardar | Se propaga el error y no se modifica la sesión |
| Cuerpo HTTP inválido o mayor de 1 MiB | `400` o `413`, respectivamente |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Votante vacío | `poker.ErrEmptyVoter` | No se registra el voto |
| Carta fuera de la escala | `poker.ErrInvalidCard` | No se registra el voto |
| Sesión inexistente | `poker.ErrSessionNotFound` | No se modifica la sesión |
| Sesión no `ABIERTA` | `poker.ErrSessionNotOpen` | No se registra el voto |
| Falla del repositorio | Error del repositorio | No se modifica la sesión |

## Criterios de aceptación

- **CA-22.1** Cada integrante puede votar en una sesión abierta con una carta de la escala: 1, 2, 3, 5, 8, 13 o 21.
- **CA-22.2** Los votos no se ven hasta que se revelan; tampoco se muestra el propio voto.
- **CA-22.3** Un integrante puede cambiar su voto mientras no se haya revelado, sin duplicarlo.
- **CA-22.4** Un voto sin nombre o con una carta fuera de la escala da error y no se registra.
- **CA-22.5** Votar en una sesión inexistente o que no está abierta da error y no modifica la sesión; si falla el guardado, se informa el error y tampoco se modifica.
- **CA-22.6** Los votos simultáneos de integrantes distintos se registran todos.
- **CA-22.7** Se vota desde la API y desde la interfaz web; la API responde siempre `200 OK`, tanto para un voto nuevo como para un reemplazo, y no permite distinguir ambos casos.
- **CA-22.8** Los votos se guardan en memoria y en PostgreSQL con el mismo comportamiento, verificado por los tests de contrato (se omiten sin `TEST_DATABASE_URL`).

## Comparación con el Planning del Sprint 3

El acta del Sprint 3 asigna a HU-22 **3 Story Points**, con dependencia de HU-21 y bloqueo de HU-23. La nota del Planning establece: “Un voto por integrante y por ronda; los votos no se muestran hasta revelarlos”. Esta SDD implementa ese alcance mediante RN-22.3 y RN-22.8.

La especificación también precisa, sin ampliar el alcance funcional del Issue:

- Votante libre, recortado y comparado sin distinguir mayúsculas.
- Reemplazo del voto existente, conforme al CA-22.3 del Issue.
- El nombre almacenado al reemplazar es el del último voto, recortado tal como llegó.
- Validación de sesión no abierta, para mantener la regla cuando HU-23 y HU-25 agreguen estados.
- Persistencia mediante `GetByID` y `Update`, previstas en HU-21.
- `GetByID` de PostgreSQL traduce `pgx.ErrNoRows` a `ErrSessionNotFound`, igual que `OpenByItem`.

No hay diferencias de alcance con el acta de Planning.

## Pruebas y trazabilidad técnica

- Tests unitarios: `internal/poker/vote_test.go`.
- Escenarios godog: `features/HU-22-votar-oculto.feature`, con runner `TestFeatures_HU22VotarOculto`.
- Contrato de repositorio: `repository_contract_test.go`.
- API: `internal/httpapi/poker_test.go`.
- Interfaz: `internal/httpapi/web_test.go`.
- Las pruebas de PostgreSQL se omiten sin `TEST_DATABASE_URL`.

HU-22: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
