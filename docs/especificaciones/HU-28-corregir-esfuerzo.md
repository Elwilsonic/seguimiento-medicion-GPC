# HU-28 · Corregir un registro de esfuerzo

> Issue del repositorio: #46

Aprobada por: pendiente (Agile Enabler).

## Historia de usuario

Como Product Builder, quiero corregir un registro de esfuerzo cargado por error, para que los datos sean confiables.

## Objetivo

Permitir corregir las horas de un esfuerzo ya registrado (HU-26) mientras su sprint está `ABIERTO`, con la misma validación de horas que al registrarlo. Se agrega la operación `CorrectHours` en `internal/effort`, `GetByID` y `Update` en sus repositorios de memoria y PostgreSQL, el endpoint `PATCH /api/esfuerzos/{id}/horas` y un formulario "Corregir esfuerzo" en la sección "Sprints" de la interfaz. Las métricas M5 a M7 y los reportes reflejan la corrección sin cambios, porque se calculan en el momento de la consulta.

## Alcance

- **Incluye:**
  - la operación `CorrectHours(id, hours)` en el `Manager` de `internal/effort`, con el error nuevo `ErrEffortNotFound`;
  - `GetByID` y `Update` en la interfaz `effort.Repository`, en `MemoryRepository` y en `PostgresRepository`, con sus casos en el test de contrato;
  - el endpoint `PATCH /api/esfuerzos/{id}/horas` y el formulario "Corregir esfuerzo" en la sección "Sprints", junto a los de HU-26 y HU-27.
- **No incluye:**
  - corregir el integrante, la fecha o la actividad: el Issue #46 pide solo las horas (decisión 1);
  - eliminar un esfuerzo: el Issue pide solo corregir;
  - cambiar la historia o el sprint de un esfuerzo: para eso se registra un esfuerzo nuevo con HU-26;
  - corregir esfuerzos de un sprint `CERRADO` ni reabrir sprints;
  - historial o auditoría de las correcciones (quién corrigió, valor anterior);
  - cambios en `internal/metrics`, `internal/report`, `internal/sprint`, `internal/backlog` ni en el esquema de la base (la tabla `efforts` ya tiene la columna `hours`).

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `CorrectHours(id int, hours float64) (Effort, error)` | ID del esfuerzo y las horas corregidas | Reemplaza las horas del esfuerzo |

### Salidas esperadas

`CorrectHours` devuelve una copia del esfuerzo corregido (`Effort`, sin punteros, slices ni mapas): las horas nuevas y los demás campos (`ID`, `Member`, `Date`, `Activity`, `ItemID` y `SprintID`) tal como estaban guardados.

### Repositorio

Se agregan dos métodos a `effort.Repository`, con el mismo contrato que `defect.Repository` (HU-30) y `backlog.Repository` (HU-07):

```go
// GetByID devuelve una copia del esfuerzo con ese ID o ErrEffortNotFound si
// no existe, incluidos los IDs 0 y negativos (HU-28).
GetByID(id int) (Effort, error)

// Update reemplaza el esfuerzo guardado cuyo ID coincida y devuelve una
// copia, o ErrEffortNotFound si no existe, sin crear ni modificar nada
// (HU-28).
Update(Effort) (Effort, error)
```

`Update` reemplaza el registro completo, como en HU-07 y HU-30; es `CorrectHours` el que cambia solo las horas y conserva el resto. En PostgreSQL es un `UPDATE efforts SET member, effort_date, activity, item_id, sprint_id, hours WHERE id = $1 RETURNING …`, con la fecha normalizada a UTC como `Save` y `scanEffort`. Sin cambios de esquema.

### HTTP

`PATCH /api/esfuerzos/{id}/horas`, con el mismo formato de horas que el registro de HU-26 (número JSON, `json.Number`):

```json
{"horas": 10}
```

Responde 200 con el esfuerzo corregido, en el formato de `effortResponse` (HU-26):

```json
{"id": 1, "sprint_id": 1, "item_id": 1, "integrante": "Ana", "fecha": "2026-10-06", "actividad": "Desarrollo", "horas": 10}
```

| Caso | Código | Cuerpo |
|---|---|---|
| Corrección exitosa | 200 | El esfuerzo corregido |
| `{id}` no numérico | 400 | `{"error": "el id debe ser un número entero"}` (`errInvalidID`) |
| JSON inválido | 400 | `errInvalidJSON`, como HU-26 |
| Horas ausentes, no numéricas, 0, negativas o infinitas | 400 | `{"error": "las horas trabajadas deben ser un número positivo"}` (`ErrInvalidHours`) |
| Esfuerzo inexistente (incluidos 0 y negativos) | 404 | `{"error": "no existe un esfuerzo con ese id"}` (`ErrEffortNotFound`, nuevo) |
| Sprint del esfuerzo `CERRADO` | 409 | `sprint.ErrSprintClosed`, el mismo texto que HU-26 |
| Falla del repositorio | 500 | `{"error": "error interno del servidor"}`, sin detalle |

### Interfaz web

En la sección "Sprints", después de "Registrar esfuerzo" y antes de "Consultar esfuerzos", un formulario "Corregir esfuerzo" (`id="form-corregir-esfuerzo"`) con los campos Esfuerzo (ID) y Horas, y el botón "Corregir esfuerzo". Al terminar muestra "Esfuerzo N corregido: H horas." en `msg-sprints`, como el registro de HU-26, y vacía el formulario. No usa el sprint actual: el esfuerzo ya tiene su sprint. El ID del esfuerzo se ve en la tabla de "Consultar esfuerzos" (HU-27).

## Reglas de negocio

1. **RN-28.1 Solo las horas.** `CorrectHours` reemplaza las horas y conserva el integrante, la fecha, la actividad, la historia y el sprint guardados.
2. **RN-28.2 Misma validación de horas que HU-26.** Horas positivas y finitas, con decimales y sin redondeo (RN-26.4). Se rechazan 0, los negativos, `NaN`, `+Inf` y `-Inf` con `ErrInvalidHours`. En la API, horas ausentes o no numéricas también dan `ErrInvalidHours`, como en HU-26.
3. **RN-28.3 Esfuerzo existente.** El esfuerzo se busca con `Repository.GetByID`; si no existe (incluidos 0 y negativos), `ErrEffortNotFound`.
4. **RN-28.4 Solo en un sprint ABIERTO.** El sprint del esfuerzo tiene que estar `ABIERTO`; si está `CERRADO`, `sprint.ErrSprintClosed`, como RN-26.5 al registrar y RN-13.6 para los cambios de un sprint cerrado. Así, un sprint cerrado conserva las horas reales con las que se cerró.
5. **RN-28.5 Atomicidad con el cierre.** La actualización se hace dentro de `sprint.Manager.WithOpenSprint(sprintID, fn)` (RN-26.11): si `Close` gana, `CorrectHours` devuelve `ErrSprintClosed` sin modificar nada; si `CorrectHours` gana, el esfuerzo queda corregido y después el sprint se cierra. Dentro de `fn` se vuelve a leer el esfuerzo y se llama a `Update` con las horas nuevas, así dos correcciones simultáneas de esfuerzos del mismo sprint quedan serializadas por el mutex del sprint y ninguna pisa un campo leído antes.
6. **RN-28.6 Orden de los errores.** Primero las horas (400), después la existencia del esfuerzo (404) y por último el estado del sprint (409). Validar antes de leer evita lecturas innecesarias, como en HU-26 y HU-30.
7. **RN-28.7 Sin no-op especial.** Corregir con las mismas horas guardadas es válido: se llama a `Update` y se devuelve el esfuerzo igual. No hay un error de "sin cambios".
8. **RN-28.8 Historia sin revalidar.** No se vuelve a validar que la historia siga asignada al sprint (RN-26.6): esa regla es del registro, y la corrección no cambia ni la historia ni el sprint.
9. **RN-28.9 Copias por valor.** `CorrectHours`, `GetByID` y `Update` devuelven copias; modificar el resultado no altera lo guardado.
10. **RN-28.10 Métricas y reportes.** M5 (horas reales), M6 y M7 (desviación), la consulta de esfuerzos de HU-27 y los reportes de HU-38 y HU-39 se calculan en el momento a partir de los esfuerzos guardados, así que reflejan la corrección sin cambios en `internal/metrics` ni en `internal/report`. Como solo se corrige en sprints abiertos, las métricas de un sprint cerrado no cambian por una corrección.
11. **RN-28.11 Persistencia.** `GetByID` y `Update` existen en memoria y en PostgreSQL detrás de la misma interfaz, con el mismo test de contrato que `Save` y los listados (HU-26 y HU-27).

## Decisión sobre entidades referenciadas

- **Esfuerzo:** sí se valida que exista (RN-28.3); es la entidad que se corrige.
- **Sprint del esfuerzo:** sí se valida que esté `ABIERTO`, con `WithOpenSprint` (RN-28.4). Su existencia no puede fallar: el sistema no elimina sprints y el `SprintID` del esfuerzo no cambia; si fallara, `WithOpenSprint` devolvería `ErrSprintNotFound` (404).
- **Historia:** no se vuelve a validar (RN-28.8). El `ItemID` no cambia y el sistema no elimina ítems.
- **Integrante:** no participa; se conserva el guardado.

## Restricciones

- **Organización del código.**

  | Archivo | Cambio |
  |---|---|
  | `internal/effort/effort.go` | `ErrEffortNotFound` |
  | `internal/effort/repository.go` | `GetByID` y `Update` en la interfaz |
  | `internal/effort/memory_repository.go` y `postgres_repository.go` | `GetByID` y `Update` |
  | `internal/effort/correct.go` | Nuevo: `CorrectHours` |
  | `internal/effort/register.go` | Solo si hace falta extraer la validación de horas para compartirla con `CorrectHours`, sin cambiar su comportamiento (refactor) |
  | `internal/httpapi/efforts.go` | `effortHoursRequest` y el handler `correctEffortHours`, que reutiliza `toEffortResponse` |
  | `internal/httpapi/server.go` | La ruta `PATCH /api/esfuerzos/{id}/horas` y `effort.ErrEffortNotFound` en el 404 de `statusFor` |
  | `internal/web/index.html` y `app.js` | El formulario "Corregir esfuerzo" y su manejo |

- **Cambios en paralelo.** HU-19 (rama `feature/HU-19-modificar-fechas-proyecto`, todavía sin Pull Request) modifica `internal/httpapi/server.go` (incluido `statusFor`), `server_test.go`, `web_test.go`, `index.html` y `app.js`. HU-28 agrega líneas propias en esos archivos, en otros bloques, así que al traer `main` puede haber conflictos triviales que se resuelven conservando los dos cambios. HU-14 ya está en `main`.
- **Repositorios de prueba.** Dos dobles de `effort.Repository` implementan la interfaz entera sin embeberla: `failingEffortRepository` (`internal/effort/register_test.go`) y `brokenEffortRepository` (`internal/httpapi/efforts_test.go`). Necesitan `GetByID` y `Update` para compilar, así que se les agregan en el RED, devolviendo su mismo error, sin cambiar su comportamiento. Los demás dobles (`recordingEffortRepository` y los espías de `internal/metrics` e `internal/report`) embeben un repositorio y no cambian.
- **Concurrencia.** Sin mutex nuevo: la corrección se serializa con el mutex del sprint (RN-28.5), con el orden de locks sprint → effort de HU-26.
- **Tests.**
  - Unitarios en `internal/effort/correct_test.go`: corrección exitosa; los demás campos sin cambios; decimales sin redondeo; un caso por valor inválido de horas (0, negativa, `NaN`, `+Inf` y `-Inf`); esfuerzo inexistente (99, 0 y −1); sprint cerrado sin modificar nada; mismas horas; copias; orden de los errores; y que las horas reales del sprint (M5) cambian después de corregir.
  - Contrato del repositorio en `repository_contract_test.go`: `GetByID` existente e inexistente (incluidos 0 y −1), `Update` existente, `Update` inexistente sin crear filas, y copias; corre en memoria y, con `TEST_DATABASE_URL`, en PostgreSQL.
  - godog en `internal/effort/correct_features_test.go`, con el runner `TestFeatures_HU28CorregirEsfuerzo` sobre `features/HU-28-corregir-esfuerzo.feature`, con el CA en el nombre de cada escenario.
  - API en `internal/httpapi/efforts_correct_test.go`: un caso por código (200, 400, 404, 409 y 500) y que `GET /api/sprints/{id}/metricas` y `GET /api/sprints/{id}/esfuerzos` reflejen la corrección. Una fila nueva en `TestRutasConID_IDNoNumericoDevuelve400`.
  - Interfaz en `internal/httpapi/web_test.go`: el formulario y su llamada.
- **RED.** Firmas vacías de `CorrectHours`, `GetByID`, `Update` y `ErrEffortNotFound`, para que el RED compile y falle por aserciones. La ruta, el handler y la interfaz van en el GREEN. `go vet ./...` antes del commit RED.
- **Profundidad proporcional.** Un test por regla y por código de error; sin pruebas de mutación ni de concurrencia con goroutines (la atomicidad con el cierre reutiliza `WithOpenSprint`, ya probado en HU-26).
- **Librerías.** Solo la librería estándar y pgx (ya aprobado); godog en los tests.

### Límites conocidos

- **Solo las horas.** Un error en el integrante, la fecha o la actividad no se puede corregir con esta historia. En particular, un integrante mal escrito ("ana" en lugar de "Ana") aparece como otro integrante en los reportes de HU-38 y HU-39 (límite conocido de esas historias); corregirlo queda como posible mejora para otra historia (decisión 1).
- **Sin historial.** La corrección reemplaza las horas: no queda registro del valor anterior ni de quién corrigió.
- **Sin control de versiones.** Dos personas que corrigen el mismo esfuerzo a la vez no se pisan a mitad de camino (RN-28.5), pero gana la última corrección, sin aviso.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Corregir las horas de 12 a 10 en un sprint abierto | 200; el esfuerzo con 10 h; M5 del sprint baja 2 h |
| Horas con decimales (2.5) | Se guardan sin redondear |
| Mismas horas que las guardadas | 200; el esfuerzo igual |
| Integrante, fecha, actividad, historia y sprint | No cambian |
| Horas 0, negativas, `NaN` o infinitas | `ErrInvalidHours`, 400; no se lee el esfuerzo |
| HTTP: horas ausentes o no numéricas (`"diez"`) | `ErrInvalidHours`, 400 |
| Esfuerzo inexistente, 0 o negativo | `ErrEffortNotFound`, 404 |
| Esfuerzo inexistente con horas inválidas | `ErrInvalidHours` (400), sin leer el esfuerzo |
| Esfuerzo de un sprint cerrado | `ErrSprintClosed`, 409; el esfuerzo no cambia |
| Esfuerzo de un sprint cerrado con horas inválidas | `ErrInvalidHours` (400) |
| Historia del esfuerzo reasignada a otro sprint | Se corrige igual (RN-28.8) |
| Corregir y después consultar los esfuerzos del sprint (HU-27) | La lista y el total muestran el valor corregido |
| Corregir y después pedir las métricas o el reporte del sprint (HU-35 y HU-39) | `horas_reales`, la desviación y `total_horas` muestran el valor corregido |
| Modificar el esfuerzo devuelto | Lo guardado no cambia |
| HTTP: `{id}` no numérico | 400 |

## Condiciones de error

| Condición | Error | HTTP |
|---|---|---|
| `{id}` no numérico | `errInvalidID` | 400 |
| JSON inválido o cuerpo demasiado grande | `errInvalidJSON` o `errBodyTooLarge` | 400 o 413 |
| Horas inválidas, ausentes o no numéricas | `effort.ErrInvalidHours` | 400 |
| No existe el esfuerzo | `effort.ErrEffortNotFound` (nuevo) | 404 |
| El sprint del esfuerzo está cerrado | `sprint.ErrSprintClosed` | 409 |
| Falla del repositorio de esfuerzos o de sprints | El error del repositorio | 500, sin detalle |

## Criterios de aceptación

- **CA-28.1** Se pueden corregir las horas de un esfuerzo existente de un sprint abierto; el integrante, la fecha, la actividad, la historia y el sprint del esfuerzo no cambian.
- **CA-28.2** Las horas corregidas se validan igual que al registrar el esfuerzo (HU-26), con el mismo error.
- **CA-28.3** No se pueden corregir las horas de esfuerzos de un sprint cerrado.
- **CA-28.4** Corregir un esfuerzo inexistente devuelve error.
- **CA-28.5** Las horas reales, la desviación, la consulta de esfuerzos y los reportes reflejan la corrección.
- **CA-28.6** La corrección se guarda en memoria y en PostgreSQL con el mismo comportamiento.
- **CA-28.7** La corrección se puede hacer desde la API y desde el formulario "Corregir esfuerzo" de la interfaz web.

## Comparación con el acta de Planning

**HU-28 no está en el alcance del Sprint 4.** El acta del Sprint 4 la lista en "Fuera de la entrega", junto con E4 (HU-15 a HU-17) y HU-19, como ya hacía el acta del Sprint 3 ("Fuera de la entrega, salvo capacidad sobrante"). El acta del Sprint 1 la ubicaba en el Sprint 4, con HU-27, y el acta del Sprint 3 la registra como bloqueada por HU-26. HU-26 y HU-27 ya están en `main`.

Esta historia **se adelanta al Sprint 5**, antes de su Planning, como se hizo con HU-32 y HU-34 en el Sprint 3: sus **Story Points se estiman en el Planning del Sprint 5**, y el adelanto se registra en la Daily y en el acta de ese Planning. Hasta entonces la historia no cumple la Definition of Ready en lo que respecta a los Story Points, y la implementación espera esa decisión del Agile Enabler.

## Comparación con los criterios del Issue #46 (borrador)

**El alcance coincide con el Issue: solo se corrigen las horas, no se eliminan esfuerzos.**

| CA original del Issue #46 | En esta SDD | Ajuste |
|---|---|---|
| CA-28.1 Se pueden modificar las horas de un registro existente | CA-28.1 | Se precisa que el registro tiene que ser de un sprint abierto y que los demás campos no cambian |
| CA-28.2 Se vuelven a validar las horas | CA-28.2 | Se precisa que es la misma validación y el mismo error que en HU-26 |
| CA-28.3 No se pueden modificar registros de un sprint cerrado | CA-28.3 | Sin cambios |
| — | CA-28.4 a CA-28.7 (nuevos) | Esfuerzo inexistente, métricas y reportes, persistencia, API e interfaz |

Los criterios del Issue se cierran con CA-28.1 a CA-28.7 cuando se apruebe esta SDD.

## Decisiones tomadas en la redacción

1. **Solo las horas, como el Issue #46.** Es lo que piden los criterios del Issue y lo que más afecta a las métricas (M5 a M7). *Alternativa descartada:* corregir los cuatro campos de HU-26 (integrante, fecha, actividad y horas) con `PUT /api/esfuerzos/{id}`. Resolvería, por ejemplo, un integrante escrito "ana" en lugar de "Ana", que en los reportes de HU-38 y HU-39 aparece como otro integrante; queda como posible mejora para otra historia, que tendría que definir si la corrección es completa o parcial.
2. **`PATCH /api/esfuerzos/{id}/horas`.** Es una corrección parcial de un solo campo, no un reemplazo completo como el `PUT /api/items/{id}` de HU-07. El proyecto ya expone los cambios de un solo campo como subrecurso con `PATCH`: `PATCH /api/defectos/{id}/estado` (HU-30), `PATCH /api/items/{id}/estado` (HU-08), `PATCH /api/items/{id}/story-points` (HU-20) y `PATCH /api/proyectos/{id}/fecha-inicio` (HU-18). La ruta nombra el campo que cambia y deja libre `PATCH /api/esfuerzos/{id}` si una historia posterior corrige otros campos. El esfuerzo ya tiene un ID global (RN-26.8), así que no hace falta repetir el sprint y la historia en la ruta. *Alternativa descartada:* `PATCH /api/esfuerzos/{id}` con `{"horas": x}`, que en el proyecto no tiene precedente para un solo campo y haría ambiguo qué pasa con los demás campos del cuerpo.
3. **No cambiar la historia ni el sprint.** Cambiarlos movería horas reales entre sprints o historias, posiblemente hacia un sprint cerrado. Si el esfuerzo se cargó en la historia o el sprint equivocados, se registra uno nuevo con HU-26.
4. **Solo en sprints ABIERTOS, con `WithOpenSprint`.** Coherente con RN-26.5 (no se registra en un sprint cerrado) y RN-13.6 (un sprint cerrado no admite cambios), y conserva las métricas con las que se cerró el sprint. Reutilizar `WithOpenSprint` da la misma atomicidad con `Close` que HU-26 sin código nuevo en `internal/sprint`. *Alternativa descartada:* permitir corregir sprints cerrados, que cambiaría la velocidad y las horas reales de sprints ya cerrados.
5. **`ErrEffortNotFound` nuevo, con 404**, como `ErrDefectNotFound` (HU-30) y `ErrItemNotFound`.
6. **`GetByID` y `Update` en el repositorio**, con el contrato de HU-07 y HU-30 (`Update` reemplaza el registro completo) y sus casos en el test de contrato de HU-26, así memoria y PostgreSQL se prueban con los mismos casos. Sin cambios de esquema. *Alternativa descartada:* un `UpdateHours` en el repositorio, que agregaría al contrato un método específico de esta historia.
7. **Validar antes de leer.** Mismo orden que HU-26 y HU-30: el error de horas (400) no depende de que el esfuerzo exista.
8. **Métricas y reportes sin cambios.** M5 a M7 y los reportes se calculan en el momento (TEC-03, HU-35, HU-38 y HU-39), así que no hace falta tocar `internal/metrics` ni `internal/report`; un test de API verifica que las métricas y la consulta de esfuerzos del sprint reflejen la corrección.
9. **Formulario propio en "Sprints"**, junto a "Registrar esfuerzo" y "Consultar esfuerzos", con el ID del esfuerzo que muestra la tabla de HU-27. *Alternativa descartada:* un botón "Corregir" en cada fila de la tabla de HU-27, que cambia una función compartida (`filaEsfuerzo`) que usan los reportes de HU-38 y HU-39.
10. **Story Points.** Se estiman en el Planning del Sprint 5.

HU-28: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
