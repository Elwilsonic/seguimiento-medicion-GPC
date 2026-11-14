# Propuesta: HU-51 · Identificadores UUID (parte de proyectos)

> Issue del repositorio: #193 · Épica E1 · Gestión de proyectos, con alcance pendiente en E2 a E8 · Vinculada a HU-01 (#14), HU-03 (#16) y HU-50 (#183)
>
> Estado: propuesta pendiente de aprobación del Agile Enabler (Peñalbé Hernán). El equipo recomienda la alternativa B (D1 de design.md).

> [!WARNING]
> **Pendiente propuesto, sujeto a aprobación (riesgo R1 de `design.md`).** Este cambio cubre **solo los proyectos**. Los ítems del backlog, los sprints, los integrantes, los defectos, los registros de esfuerzo y las sesiones de Planning Poker **siguen con IDs enteros correlativos**. Mientras esa parte no se haga:
>
> - quien recorra `/api/items/1`, `/api/items/2`… (o `/api/sprints/N/…`) puede leer el `proyecto_id` (UUID) de cada ítem o sprint y llegar con él al proyecto;
> - los IDs globales de ítems, sprints y las demás entidades dejan deducir cuántos hay.
>
> Se propone como pendiente documentado, para que lo acepte el Agile Enabler: no hay Sprint siguiente, y se informa en la Review del 15/11. Ver "Fuera de alcance y PENDIENTE".

## Historia de usuario

Como integrante del equipo, quiero que los proyectos se identifiquen con un UUID en lugar de un número correlativo, para que no se pueda adivinar el identificador de un proyecto ajeno ni deducir cuántos proyectos hay.

## Why

Hoy cada proyecto se identifica con un entero correlativo (RN-01.5 de HU-01: 1, 2, 3…), y ese número es lo que viaja en todas las rutas `/api/proyectos/{id}/…`, en el campo `proyecto_id` de las respuestas y en el campo «Proyecto actual» de la interfaz. Con eso, cualquiera prueba `/api/proyectos/2` y ve un proyecto ajeno, y el último ID dice cuántos proyectos hay. El profesor lo marcó en la revisión del 05/11/2026 (Planning del Sprint 5, Issue #182): "No usar IDs 1, 2, 3, que son fáciles de adivinar; usar UUID". HU-50 ya dejó la lista de proyectos para que nadie tenga que recordar el identificador.

## What Changes

- **Identificador público:** cada proyecto nuevo recibe un UUID versión 4 aleatorio, generado por el sistema. Es el **único** identificador del proyecto que muestran la API y la interfaz. El entero correlativo se conserva como clave interna de la base (clave primaria y claves foráneas) y deja de verse (decisión D1 de `design.md`).
- **API:** todas las rutas `/api/proyectos/{id}` y `/api/proyectos/{id}/…` reciben el UUID. Un `{id}` que no tiene formato de UUID responde **400**. Un UUID válido que no corresponde a ningún proyecto responde **404**.
- **Respuestas:** el campo `id` de los proyectos y el campo `proyecto_id` de ítems, sprints, integrantes, métricas y resumen pasan a ser el UUID (texto). El nombre del PDF de HU-40 usa el UUID en lugar del número.
- **Base de datos:** la tabla `projects` suma la columna `public_id UUID NOT NULL UNIQUE`. Las tablas que referencian proyectos (`backlog_items`, `sprints`, `members`) no cambian.
- **Interfaz web:** el campo «Proyecto actual» pasa de número a texto y acepta el UUID. «Ver proyectos» y «Elegir» (HU-50) siguen funcionando igual, y la descarga del PDF usa el UUID en el nombre.
- **Mensaje de error nuevo:** un `{id}` de proyecto inválido responde `{"error":"el id del proyecto debe ser un UUID"}`. Las demás entidades conservan "el id debe ser un número entero".
- **BREAKING (API):** un cliente que usaba `/api/proyectos/1` o esperaba `"id": 1` o `"proyecto_id": 1` deja de funcionar: tiene que usar el UUID. La interfaz propia se adapta en este mismo cambio. No hay otros clientes.
- **Datos existentes:** el esquema se recrea (`docker compose down -v`); no hay migración (D6).

## Capabilities

### New Capabilities

- `proyectos`: gestión de proyectos (épica E1). Si HU-50 se archiva antes, este cambio le **agrega** requisitos a esa misma capability y su `## Purpose` se ignora. Los requisitos de HU-50 no cambian: la lista conserva su formato y su orden por nombre.

### Modified Capabilities

Ninguna. Los requisitos que cambian de comportamiento (RN-01.5, "ID numérico correlativo", y el "ID entero" de las entradas de HU-02, HU-03, HU-18, HU-19 y HU-41) están en las SDD de `docs/especificaciones/`, no en `openspec/specs/`. Por eso van como requisitos `ADDED` que dicen explícitamente qué reemplazan. Las SDD no se rehacen.

## Vínculos y SDD leídas

- **HU-01** (`HU-01-crear-proyecto.md`): RN-01.5 y CA-01.1 ("ID numérico correlativo que empieza en 1") siguen valiendo para la clave **interna**. Lo que se ve hacia afuera pasa a ser el UUID. Se mantiene el mutex de CA-01.7 para la creación.
- **HU-02** y **HU-03** (`HU-02-modificar-proyecto.md`, `HU-03-consultar-proyecto.md`): `ErrProjectNotFound` (RN-02.6, RN-03.2) sigue siendo el error de proyecto inexistente, 404 en la API. Las consultas siguen devolviendo copias por valor (RN-03.3).
- **HU-18, HU-41 y HU-19**: el orden de errores del handler (ID de la ruta → JSON → formato de fecha → Manager, donde las validaciones de contenido van antes que la existencia) se conserva. Por ejemplo, una fecha vacía sobre un proyecto inexistente sigue respondiendo 400 (D4).
- **HU-50** (`openspec/changes/hu-50-listar-proyectos/`): el orden alfabético por nombre (D2 de HU-50) no depende del ID y no cambia. `GET /api/proyectos` conserva su formato, con el `id` como UUID.

## Comparación con el acta de Planning

El acta del Planning del Sprint 5 (Issue #182, PR #184) describe HU-51 con 13 Story Points, por partes: primero proyectos (obligatorio) y después el resto de las entidades "si el equipo llega con tiempo". Lo que no llegue "queda escrito en la propuesta de OpenSpec como trabajo pendiente". **El alcance coincide**: esta propuesta cubre la parte obligatoria y deja el resto como pendiente, escrito abajo.

Hay dos diferencias, para que las decida el Agile Enabler:

1. **Fecha:** el acta pide HU-51 integrada (al menos proyectos) el sábado 14/11. Este plan apunta al domingo 15/11, antes de la Review.
2. **Redacción del Issue:** el Issue dice que el UUID se usa "en el dominio, la base de datos, la API y la interfaz". Con la alternativa recomendada (D1), el dominio y la base conservan además el entero interno, que nunca sale de la API. Por eso los criterios CA-51.1, CA-51.3 y CA-51.5 se ajustan (ver la tabla de abajo).

## Criterios de aceptación definitivos

| Criterio | Texto | Cambio respecto del borrador del Issue |
|---|---|---|
| CA-51.1 | Un proyecto nuevo recibe un identificador público UUID versión 4 generado por el sistema, en forma canónica (36 caracteres, minúsculas). La API y la interfaz ya no muestran 1, 2, 3. | Se agrega "público" y la forma canónica: el entero interno sigue existiendo, pero no se expone (D1). |
| CA-51.2 | Dos proyectos nunca tienen el mismo identificador público, y a partir de uno no se puede deducir otro ni la cantidad de proyectos. | Sin cambios. Se agrega la restricción `UNIQUE` en PostgreSQL como garantía (D3). |
| CA-51.3 | Todas las rutas que reciben un proyecto aceptan su UUID (consultar, modificar, fechas, integrantes, backlog, sprints, defectos, métricas, resumen y reportes), y **ninguna respuesta de la API expone el ID entero del proyecto**: `id`, `proyecto_id` y el nombre del PDF llevan el UUID. | Se precisa la lista de rutas y se agrega que ninguna salida filtra el entero. Sin esto, el UUID no cumpliría el pedido (D5). |
| CA-51.4 | En la API, un `{id}` de proyecto que no tiene la forma `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` responde 400 (`errInvalidID`). Un UUID con esa forma que no existe responde 404 (`ErrProjectNotFound`). Se conserva el orden de errores de cada ruta. | Se define qué es "formato de UUID": sin llaves, sin `urn:` y con guiones; las mayúsculas se aceptan. Se agrega que el orden de errores no cambia (D2, D4). |
| CA-51.5 | Funciona igual en memoria y en PostgreSQL. El esquema agrega `projects.public_id` (UUID, único), y las relaciones con los proyectos siguen funcionando sin cambios. | Se precisa cómo se actualiza el esquema: las claves foráneas siguen siendo enteras (D1, D6). |
| CA-51.6 | La interfaz web funciona con UUID: «Proyecto actual» acepta el UUID, «Ver proyectos» y «Elegir» (HU-50) siguen funcionando, y el PDF se descarga con el UUID en el nombre. | Se agrega el nombre del PDF. |
| CA-51.7 | Crear proyectos en simultáneo genera identificadores públicos únicos, sin condiciones de carrera (`go test -race`). | Sin cambios. |

## Fuera de alcance y PENDIENTE

**PENDIENTE (parte 2 de HU-51, no se hace en este cambio):** los **ítems del backlog, los sprints, los integrantes, los defectos, los registros de esfuerzo y las sesiones de Planning Poker siguen con IDs enteros correlativos**, en el dominio, en la base, en la API (`/api/items/{id}`, `/api/sprints/{id}`, `/api/integrantes/{id}`, `/api/defectos/{id}`, `/api/esfuerzos/{id}`, `/api/poker-sessions/{id}`) y en la interfaz (por ejemplo, «Sprint actual»). Queda escrito como trabajo pendiente, según el acta. La consecuencia está en el riesgo R1 de `design.md`: mientras esa parte no se haga, quien recorra `/api/items/1`, `/api/items/2`… puede leer el `proyecto_id` (UUID) de esos ítems.

**Fuera de alcance:**

- Control de acceso, usuarios o permisos: el UUID hace muy difícil adivinar un proyecto, pero no reemplaza a un control de acceso (nota del Issue).
- Migrar datos existentes de una base con IDs enteros: el esquema se recrea (D6).
- Cambiar el orden de la lista de HU-50, o cambiar su formato más allá del tipo de `id`.
- Aceptar otras formas de UUID en las rutas (`{…}`, `urn:uuid:…`, sin guiones).

## Impact

Cantidades contadas con comandos (ver "Archivos que cambian" en `design.md`):

- **Producción, 21 archivos modificados y ninguno nuevo:**
  - `internal/project`: 6 archivos (`project.go`, `repository.go`, `memory_repository.go`, `postgres_repository.go`, `get.go` y `create.go`) (D3, D4);
  - `internal/httpapi`: 13 archivos (`server.go` y los 12 handlers que reciben o devuelven un proyecto);
  - `db/init/01-schema.sql`;
  - `internal/web/index.html`. `app.js` ya trata el ID como texto y no cambia (D7).
- **Tests, 22 archivos existentes:** 20 de `internal/httpapi`, que usan `/api/proyectos/1` o `"proyecto_id":1`, y 2 de `internal/project`, con repositorios falsos que implementan `Repository`. A eso se suman archivos nuevos de tests y de pasos godog.
- **Documentación:** `features/HU-51-identificadores-uuid.feature`, `docs/evidencias/HU-51-tdd.md`, `docs/reportes/cobertura.md`, `README.md` (manual, tabla de la API y fila de HU-51), `docs/manual-de-usuario.md` y `docs/ia/registro.md`.
- **Sin dependencias nuevas:** el paquete estándar `uuid` de Go 1.27 (la versión de `go.mod`, la CI y el `Dockerfile`) y el tipo `uuid` de PostgreSQL 17.
