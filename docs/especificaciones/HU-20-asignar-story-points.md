# HU-20 · Asignar story points a un ítem del backlog

> Issue del repositorio: #38

Aprobada por: Peñalbé Hernán (Agile Enabler), 23/10/2026, con los ajustes de RN-20.5, la dependencia HU-21 → HU-20 y el test de concurrencia.

## Historia de usuario

Como Product Builder, quiero asignar story points a un ítem del backlog, para estimar su tamaño.

## Objetivo

Asignar (o reasignar) los Story Points de un ítem existente del Product Backlog con un valor de la escala Fibonacci acordada (1, 2, 3, 5, 8, 13 o 21), mediante una operación nueva del `Manager` de `internal/backlog` que guarda con el `Update` del repositorio de HU-07. Es el dato de entrada de las métricas M1, M2 y M3 de TEC-03 (`docs/metricas.md`) y la operación que HU-25 reutiliza para registrar la estimación acordada en Planning Poker. Se expone en `PATCH /api/items/{id}/story-points` y en la tabla del backlog de la interfaz web (tema 5 del Planning del Sprint 3).

## Alcance

- **Incluye:** la operación `AssignStoryPoints` del `Manager`, su error de escala y su error de ítem terminado; la verificación de que el `Update` de HU-07 conserva los Story Points; el endpoint HTTP y la columna y el control de Story Points en la tabla del backlog de la interfaz web.
- **No incluye:** Planning Poker (HU-21 a HU-25; HU-25 llama a esta operación), quitar la estimación de un ítem (volver a 0), el cálculo de las métricas (E9), ni cambios en la interfaz `Repository`, en los repositorios en memoria o PostgreSQL ni en el esquema de la base (la columna `story_points` ya existe desde HU-04/TEC-05 y `Update` ya la escribe).

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | Identificador del ítem a estimar |
| Story Points | entero (`int`) | Sí | Valor nuevo; tiene que ser 1, 2, 3, 5, 8, 13 o 21 |

Firma: `AssignStoryPoints(id, points int) (Item, error)`, en `internal/backlog/story_points.go`.

### Escala de Story Points (definición única)

La escala Fibonacci acordada (1, 2, 3, 5, 8, 13, 21) se define **en un solo lugar** del paquete `backlog`, en `story_points.go`, y se exporta para que HU-21 y HU-22 (Planning Poker) usen las mismas cartas sin repetir los valores:

- `storyPointScale = [...]int{1, 2, 3, 5, 8, 13, 21}`: arreglo no exportado, la única definición de los valores, en orden ascendente.
- `StoryPointScale() []int`: función exportada que devuelve una **lista nueva** con esos valores en orden ascendente. Es una función y no una variable exportada para que ningún llamador pueda modificar la escala (mismo criterio de copias de RN-20.8): cambiar la lista devuelta no altera la escala ni la validación.
- `AssignStoryPoints` valida contra `storyPointScale`; ningún otro archivo del repositorio escribe los valores en Go. HU-21 y HU-22 tienen que usar `backlog.StoryPointScale()` (queda indicado en el comentario de la función).

La interfaz web (JavaScript) no puede importar el paquete Go, así que el `select` tiene su propia lista. Para que no se desincronice, el test de la interfaz arma la lista esperada con `backlog.StoryPointScale()` y verifica que `app.js` la contiene tal cual.

### Salidas esperadas

- **Si la asignación es válida:** el ítem actualizado, como **copia por valor** (con su propia lista de criterios), con `StoryPoints` igual al valor recibido y todos los demás campos (ID, ProjectID, título, descripción, prioridad, criterios, estado, horas estimadas y sprint) iguales a los que tenía.
- **Si no es válida:** el error que indica el motivo (ver "Condiciones de error"); el ítem guardado queda exactamente como estaba.

### HTTP (TEC-04)

`PATCH /api/items/{id}/story-points`, con el mismo patrón que `PATCH /api/items/{id}/estado` (HU-08).

- **Cuerpo:** `{"story_points": 5}`.
- **Respuesta exitosa:** `200 OK` con el ítem (`itemResponse`, que ya incluye `story_points`).
- **Errores:**

| Caso | Código | Mensaje |
|---|---|---|
| Valor fuera de la escala: 0, negativo, 4, 22, decimal (2.5), número enorme, `null` o campo ausente | 400 | El de `ErrInvalidStoryPoints` |
| Cuerpo que no es JSON, `story_points` que no es número (por ejemplo `true`), o texto después del JSON | 400 | El de `errInvalidJSON` |
| `{id}` que no es un número entero | 400 | El de `errInvalidID` |
| Ítem inexistente | 404 | El de `ErrItemNotFound` |
| Ítem `TERMINADO` | 409 | El de `ErrItemFinished` |
| Cuerpo de más de 1 MiB | 413 | El de `errBodyTooLarge` |
| Falla del repositorio | 500 | El de `errInternal`, sin detalle |

Como el decimal tiene que responder con el mensaje de la escala y no con el de JSON inválido, el handler lee `story_points` como `json.Number` y lo convierte con `strconv.Atoi`: si no es un entero representable, responde con `ErrInvalidStoryPoints`; si lo es, llama a `AssignStoryPoints`, que valida la escala. Con `json.Number`, el texto `"5"` se acepta como el número 5 (comportamiento de `encoding/json`); se documenta como caso límite.

### Interfaz web (TEC-04)

En la tabla del backlog (`internal/web`):

- Columna nueva **"SP"** con los Story Points del ítem, o "—" si vale 0 (sin estimar, TEC-03).
- En las acciones de cada fila, un `select` con los valores de la escala, 1, 2, 3, 5, 8, 13 y 21 (sincronizados con `backlog.StoryPointScale()` por test; preseleccionado el actual, si tiene) y el botón **"Asignar SP"**, que llama al endpoint, recarga el backlog y muestra "Ítem N con X story points." o el mensaje de error, con el mismo mecanismo que "Cambiar estado".
- Test de la interfaz, como los de `web_test.go`: la página servida contiene la columna y el control, y la lista de valores de `app.js` coincide con `backlog.StoryPointScale()`.

## Reglas de negocio

1. **RN-20.1 Escala Fibonacci.** Los Story Points válidos son exactamente 1, 2, 3, 5, 8, 13 y 21 (escala del Issue #38 y del Planning Poker del Sprint 3), definidos una sola vez en el paquete `backlog` y expuestos con `StoryPointScale()` (ver "Escala de Story Points"). El máximo es 21. Cualquier otro entero (0, negativos, 4, 6, 22, 34, …) devuelve `ErrInvalidStoryPoints`.
2. **RN-20.2 El 0 no se asigna.** 0 significa "sin estimar" (TEC-03) y es el valor con el que HU-04 crea los ítems, pero no se puede asignar: no está en la escala y ninguna historia pide quitar una estimación. Devuelve `ErrInvalidStoryPoints`.
3. **RN-20.3 Reestimar.** Se puede asignar un valor a un ítem que ya tiene Story Points: el nuevo reemplaza al anterior (HU-25 lo necesita para guardar la estimación acordada). Asignar el mismo valor que ya tiene es válido y deja el ítem igual.
4. **RN-20.4 Ítem existente.** Requiere un ítem con ese ID; si no, `ErrItemNotFound` (el de HU-05).
5. **RN-20.5 Ítem terminado.** No se cambian los Story Points de un ítem `TERMINADO`: devuelve `ErrItemFinished`. Un ítem terminado ya suma en M2 (Story Points completados) de su sprint y, a través de M2, en M3 (velocidad): la regla evita cambiar esos Story Points directamente. No bloquea la reapertura: HU-08 permite `TERMINADO → EN_CURSO` sin mirar el sprint, y un ítem reabierto vuelve a ser estimable (ver "Límite conocido" en la decisión 4). Los ítems `PENDIENTE` y `EN_CURSO` se pueden estimar y reestimar, estén o no asignados a un sprint.
6. **RN-20.6 Solo cambian los Story Points.** No se modifica ningún otro campo del ítem.
7. **RN-20.7 Update de HU-07 conserva los Story Points.** Modificar los datos de un ítem con `Update` (HU-07) no cambia los Story Points asignados por HU-20. El código actual ya copia `current.StoryPoints`; se agrega un test que lo verifica con un ítem estimado, porque hasta ahora todos los ítems tenían 0.
8. **RN-20.8 Copia por valor.** Se devuelve una copia (con lista de criterios nueva): modificar el resultado no altera el ítem guardado.

## Decisión sobre entidades referenciadas

- **Ítem:** **sí se valida su existencia** (RN-20.4), con `Repository.GetByID` de HU-05, dentro del mutex, porque la operación es leer-modificar-escribir y además necesita el estado actual para RN-20.5.
- **Proyecto:** **no se valida.** No es entrada; el ítem existente ya pertenece a un proyecto validado al crearlo (RN-04.5) y los proyectos no se borran.
- **Sprint del ítem:** **no se valida ni se consulta.** La regla de RN-20.5 se apoya solo en el estado del ítem, que está en el propio ítem. Consultar el sprint obligaría a que `backlog` dependa de `sprint`, que ya depende de `backlog` (ciclo de imports), o a pasar la operación por `sprint.Manager`, lo que no corresponde a una historia de 2 SP. Consecuencia, registrada como "Límite conocido" en la decisión 4: un ítem `PENDIENTE` o `EN_CURSO` que quedó en un sprint `CERRADO` (RN-13.10) se puede reestimar; eso cambia M1 (planificados) de ese sprint, pero no M2 ni M3.
- **Escala compartida con Planning Poker:** HU-21 y HU-22 usan `backlog.StoryPointScale()` para las cartas. Eso crea la dependencia **HU-21 → HU-20**, que no figura en la tabla de dependencias del acta del Sprint 3: HU-20 tiene que estar integrada en `main` antes del commit GREEN de HU-21; el responsable de HU-21 se entera por la Daily. HU-22 depende de HU-21, así que queda cubierta por la misma condición.

## Restricciones

- **Organización del código.** Código nuevo en `internal/backlog/story_points.go`: la escala (`storyPointScale` y `StoryPointScale()`), `AssignStoryPoints`, `ErrInvalidStoryPoints`, `ErrItemFinished` y la validación de la escala. No se tocan `create.go`, `update.go`, `status.go` ni `assign.go`.
- **Repositorio.** Se reutilizan `GetByID` y `Update`. La interfaz `Repository`, `memory_repository.go`, `postgres_repository.go`, `db/init/01-schema.sql` y los contratos de TEC-05 no cambian (el contrato de `Update` ya escribe todos los campos, incluido `story_points`).
- **Orden de validaciones.** Primero la escala (RN-20.1 y RN-20.2), sin consultar el repositorio. Después, con el mutex del `Manager` tomado: existencia (RN-20.4), estado `TERMINADO` (RN-20.5) y escritura con `Update`. Si falla cualquiera, no se escribe nada.
- **Errores identificables.** `ErrInvalidStoryPoints` ("los story points deben ser 1, 2, 3, 5, 8, 13 o 21") y `ErrItemFinished` ("no se pueden cambiar los story points de un ítem terminado") son nuevos, exportados y comparables con `errors.Is`. Se reutiliza `ErrItemNotFound`. En `httpapi`, `ErrInvalidStoryPoints` se traduce a 400 y `ErrItemFinished` a 409.
- **Concurrencia.** `AssignStoryPoints` es leer-modificar-escribir: la lectura del ítem (`GetByID`), la verificación de `TERMINADO` y la escritura (`Update`) van **con el mutex del `Manager` tomado durante toda la operación** (mismo patrón que `Update` y `ChangeStatus`), para que una asignación no pise un cambio simultáneo de estado ni de datos. Lo verifica un test que lanza asignaciones simultáneas al mismo ítem (sobre un ítem `EN_CURSO` sin estimar), junto con modificaciones de `Update` y un `ChangeStatus` a `TERMINADO`. Si el `ChangeStatus` llega antes que una asignación, esa asignación devuelve `ErrItemFinished`: el test acepta ese error como resultado válido y falla con cualquier otro. Las aserciones no dependen del orden en que corren las goroutines:
  - el `ChangeStatus` y todas las modificaciones de `Update` terminan bien, y el ítem queda `TERMINADO` con los datos de la modificación;
  - después de `TERMINADO` no se escribe ningún valor, `Update` conserva los Story Points y la escala no tiene el 0; por eso, si ninguna asignación terminó bien, los Story Points guardados son 0, y si alguna terminó bien, son el valor de una de las asignaciones exitosas (no alcanza con que sea uno de los valores enviados);
  - corre con `go test -race` en el CI sin reportar carreras.
- **Copias por valor.** Se devuelve `Item` por valor, nunca punteros al estado interno.
- **Pasos godog.** En `internal/backlog/story_points_features_test.go`, con su runner `TestFeatures_HU20AsignarStoryPoints`. Cada escenario lleva su criterio en el nombre, por ejemplo "Escenario: … (CA-20.2)".
- **HTTP e interfaz.** Handler en `internal/httpapi/items.go` y ruta en `server.go`; tests con un caso exitoso y uno por código de error (criterio de TEC-04). Interfaz en `internal/web/index.html` y `app.js`.
- **Librerías.** Solo la librería estándar; godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Asignar 1 (mínimo) y 21 (máximo) | Se guardan |
| Asignar cada valor de la escala: 1, 2, 3, 5, 8, 13, 21 | Se guarda cada uno |
| Asignar 0 | `ErrInvalidStoryPoints`; el ítem queda como estaba |
| Asignar -1 o -5 | `ErrInvalidStoryPoints` |
| Asignar 4, 6, 7 (entre valores de la escala) | `ErrInvalidStoryPoints` |
| Asignar 22 o 34 (mayores que el máximo) | `ErrInvalidStoryPoints` |
| Reestimar un ítem con 3 a 8 | Queda en 8 |
| Asignar el mismo valor que ya tiene (5 → 5) | Válido; queda en 5 |
| Ítem `EN_CURSO` | Se asigna |
| Ítem `TERMINADO` (con 0 o con Story Points) | `ErrItemFinished`; no cambia |
| Ítem `PENDIENTE` asignado a un sprint (abierto o cerrado) | Se asigna (ver "Decisión sobre entidades referenciadas") |
| ID inexistente, 0 o negativo, o sin ítems creados | `ErrItemNotFound` |
| Valor fuera de escala e ID inexistente a la vez | `ErrInvalidStoryPoints` (la escala se valida primero) |
| Valor fuera de escala sobre un ítem `TERMINADO` | `ErrInvalidStoryPoints` |
| Modificar el ítem devuelto o sus criterios | El ítem guardado no cambia |
| `Update` (HU-07) sobre un ítem con 8 Story Points | Siguen en 8 |
| Varias asignaciones simultáneas al mismo ítem, mezcladas con `Update` y `ChangeStatus` | Cada asignación termina bien o con `ErrItemFinished` (si el `ChangeStatus` dejó el ítem en `TERMINADO` antes); el ítem queda `TERMINADO`; los Story Points guardados son los de una asignación exitosa, o 0 si ninguna lo fue; `go test -race` sin carreras |
| Modificar la lista que devuelve `StoryPointScale()` | La escala y la validación no cambian; la siguiente llamada devuelve 1, 2, 3, 5, 8, 13, 21 |
| HTTP: `{"story_points": 2.5}`, `null`, `{}` o `1e400` | 400 con el mensaje de `ErrInvalidStoryPoints` |
| HTTP: `{"story_points": "5"}` | Se acepta como 5 (comportamiento de `json.Number`) |
| HTTP: `{"story_points": true}` | 400 con el mensaje de JSON inválido |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Valor que no es 1, 2, 3, 5, 8, 13 ni 21 (incluye 0 y negativos) (RN-20.1, RN-20.2) | `ErrInvalidStoryPoints` | No se modifica nada |
| No existe el ítem (RN-20.4) | `ErrItemNotFound` | No se modifica nada |
| El ítem está `TERMINADO` (RN-20.5) | `ErrItemFinished` | No se modifica nada |
| El repositorio falla al buscar o actualizar | El error del repositorio | El ítem queda como estaba |

## Criterios de aceptación

- **CA-20.1** Se asignan Story Points a un ítem existente: el ítem queda con el valor asignado y los demás datos sin cambios. Un ítem que ya tenía Story Points se puede reestimar y el valor nuevo reemplaza al anterior.
- **CA-20.2** Los valores permitidos son los de la escala Fibonacci acordada: 1, 2, 3, 5, 8, 13 y 21 (máximo 21), definida en un solo lugar del paquete `backlog` y expuesta con `StoryPointScale()`, que devuelve una lista nueva en orden ascendente.
- **CA-20.3** Un valor fuera de la escala (0, negativo, intermedio como 4 o mayor que 21) devuelve `ErrInvalidStoryPoints` y el ítem no cambia.
- **CA-20.4** Asignar Story Points a un ítem inexistente devuelve `ErrItemNotFound`.
- **CA-20.5** Los Story Points de un ítem `TERMINADO` no se pueden cambiar directamente: devuelve `ErrItemFinished` y el ítem no cambia. Un ítem reabierto con HU-08 (`TERMINADO → EN_CURSO`) vuelve a ser estimable.
- **CA-20.6** Modificar un ítem con el `Update` de HU-07 conserva sus Story Points.
- **CA-20.7** La operación devuelve una copia por valor. La lectura y la escritura del ítem se hacen con el mutex del `Manager` tomado, y las asignaciones simultáneas (también con `Update` y `ChangeStatus`) no producen accesos sin protección (`go test -race`); cada una termina bien o con `ErrItemFinished` si un `ChangeStatus` simultáneo dejó el ítem en `TERMINADO`, y los Story Points guardados son los de una asignación exitosa, o 0 si ninguna lo fue. Si el repositorio falla, se devuelve ese error y el ítem no cambia.
- **CA-20.8** Se puede asignar desde la API (`PATCH /api/items/{id}/story-points`, con los códigos de la tabla HTTP) y desde la tabla del backlog de la interfaz web, que muestra los Story Points de cada ítem. Se cubre con los tests de `internal/httpapi` (`items_test.go` para el endpoint y `web_test.go` para la página servida), no con escenarios godog.

## Comparación con el acta de Planning

El acta del Sprint 3 (`docs/actas/sprint-3/planning.md`) describe HU-20 con 2 Story Points, responsable Cardozo Leandro, dependiente de HU-04 y HU-07 y bloqueante de HU-25, con la justificación: "El ítem ya tiene el campo `StoryPoints` (HU-04); valida la escala Fibonacci y guarda con el `Update` del backlog". El tema 5 pide endpoint bajo `/api/…` y pantalla en la interfaz web. **El alcance coincide**: no se agregan campos ni cambia el repositorio; se valida la escala y se guarda con `Update`; se suman el endpoint y la interfaz. Lo único que esta SDD agrega y el acta no menciona es la regla del ítem `TERMINADO` (RN-20.5), que es una restricción y no funcionalidad nueva, y la dependencia HU-21 → HU-20 por la escala compartida (ver "Decisión sobre entidades referenciadas"); las dos quedaron aprobadas con la SDD.

## Comparación con los criterios del Issue #38 (borrador)

| CA en el Issue #38 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-20.1 Se asignan story points a un ítem existente. | CA-20.1 (precisado) | Se aclara que los demás datos no cambian y que se puede reestimar |
| CA-20.2 Los valores permitidos son los de la escala Fibonacci acordada: 1, 2, 3, 5, 8, 13 o 21. | CA-20.2 | Se explicita el máximo (21) |
| CA-20.3 Un valor fuera de la escala devuelve error. | CA-20.3 (precisado) | Se identifica el error y se incluyen 0, negativos e intermedios |
| — | CA-20.4 (nuevo) | Ítem inexistente |
| — | CA-20.5 (nuevo) | Ítem terminado (evita cambiar directamente los Story Points que suman en M2 y M3 de TEC-03) |
| — | CA-20.6 (nuevo) | `Update` de HU-07 conserva los Story Points |
| — | CA-20.7 (nuevo) | Copia, concurrencia y falla del repositorio, como HU-07 y HU-08 |
| — | CA-20.8 (nuevo) | API e interfaz web (tema 5 del Planning del Sprint 3) |

**Texto definitivo para la sección "Criterios de aceptación" del Issue #38** (en lenguaje de negocio, mejora 8 del Planning; se reemplaza cuando se apruebe la SDD y con tu autorización):

```markdown
## Criterios de aceptación

- [ ] CA-20.1 Se asignan story points a un ítem existente, sin cambiar sus demás datos; un ítem ya estimado se puede reestimar.
- [ ] CA-20.2 Los valores permitidos son los de la escala Fibonacci acordada: 1, 2, 3, 5, 8, 13 o 21 (máximo 21), las mismas cartas que usa el Planning Poker.
- [ ] CA-20.3 Un valor fuera de la escala (0, negativo, intermedio o mayor que 21) devuelve error y el ítem no cambia.
- [ ] CA-20.4 Asignar story points a un ítem inexistente devuelve error.
- [ ] CA-20.5 No se pueden cambiar directamente los story points de un ítem terminado; si se reabre, vuelve a ser estimable.
- [ ] CA-20.6 Modificar los datos del ítem (HU-07) conserva sus story points.
- [ ] CA-20.7 Las asignaciones simultáneas no se pisan y un error al guardar deja el ítem como estaba.
- [ ] CA-20.8 Los story points se asignan y se ven desde la API y desde la interfaz web.
```

## Decisiones tomadas en la redacción de la SDD

1. **El 0 se rechaza** (`ErrInvalidStoryPoints`): no está en la escala del Issue y ninguna historia pide quitar una estimación. El 0 solo existe como valor inicial de HU-04 ("sin estimar", TEC-03). *Alternativa:* aceptar 0 para borrar la estimación.
2. **Escala:** 1, 2, 3, 5, 8, 13 y 21; máximo 21, definida una sola vez en `story_points.go` y exportada con `StoryPointScale()` para que HU-21 y HU-22 usen las mismas cartas. Un único error, `ErrInvalidStoryPoints`, para todo valor fuera de la escala. "No entero" solo puede llegar por HTTP (la función recibe `int`): un decimal responde 400 con el mismo mensaje.
3. **Reestimar sí**, reemplazando el valor (HU-25 lo necesita). Mismo valor: válido, sin caso especial.
4. **Ítem `TERMINADO`: se rechaza** (`ErrItemFinished`, 409), porque sus Story Points ya suman en M2 y M3, que solo cuentan ítems terminados. **Sprint `CERRADO`: no se consulta.** *Alternativa descartada:* bloquear también los ítems de sprints cerrados, pasando la operación por `sprint.Manager` (más alcance que 2 SP). La lectura y la escritura del ítem van con el mutex del `Manager` tomado, verificado con un test bajo `-race` (CA-20.7).
   - **Límite conocido:** un ítem no terminado (`PENDIENTE` o `EN_CURSO`) que quedó en un sprint `CERRADO` (RN-13.10) se puede reestimar. Eso cambia M1 (Story Points planificados) de ese sprint, pero no M2 ni la velocidad (M3), que solo cuentan ítems `TERMINADO`.
   - **Límite conocido:** la regla evita cambiar directamente los story points de un ítem terminado. Un ítem que se reabre con HU-08 (`TERMINADO → EN_CURSO`) vuelve a ser estimable; bloquear la reapertura cambiaría HU-08 y queda fuera de alcance.
5. **Función nueva** `AssignStoryPoints` en `story_points.go`, y no un parámetro más en `Update`: así HU-07 no cambia de firma, el `PUT /api/items/{id}` y el formulario de edición siguen igual, y HU-25 tiene una operación propia que llamar. El `Update` de HU-07 ya conserva los Story Points; se agrega el test que lo verifica (CA-20.6).
6. **Endpoint** `PATCH /api/items/{id}/story-points` con `{"story_points": N}`, como el de estado de HU-08, y en la interfaz web una columna "SP" más un `select` con la escala y el botón "Asignar SP" en cada fila.
