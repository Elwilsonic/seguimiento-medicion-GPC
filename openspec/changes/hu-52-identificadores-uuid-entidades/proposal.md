# Propuesta: HU-52 · Identificadores UUID del resto de las entidades

> Issue del repositorio: #197 · Épica E1 · Gestión de proyectos, con alcance en E2 a E8 · Vinculada a HU-51 (#193) y a las historias de cada entidad (HU-04 a HU-17 y HU-20 a HU-31)
>
> Aprobada por: Peñalbé Hernán (Agile Enabler), 15/11/2026. Sus decisiones:
>
> 1. Se aprueba con el dominio de las 6 entidades en la base (D3), hecha en pareja con Cardozo Leandro, y con el centinela `missingInternalID = math.MaxInt32` (D4).
> 2. La base se integra en `main` hoy, 15/11/2026, a las 21:00. Antes del PR se trae `main` y se corre `go test ./internal/trazabilidad/ -run TestEnlaces -args -update` (el #199, con la columna Código, ya está en `main`).
> 3. Diferencias con el acta: se aprueban las 6 entidades (con el corte por entidad), el corte del lunes 16/11/2026 a las 14:00, y los Story Points se votan en Planning Poker.
> 4. Cada responsable autoriza sus propios commits, como en HU-51.
>
> **Corte: lunes 16/11/2026 a las 14:00** (el Sprint 5 se extiende hasta ese momento). Lo que no esté integrado en `main` con la CI en verde a esa hora queda como pendiente, según la sección "Si no llega al corte".
>
> **Horario de la base:** integrada en `main` el domingo 15/11/2026 a las 21:00. La hacen Sosa Ricardo y Cardozo Leandro en pareja, sujeto a que Cardozo Leandro lo confirme.

## Historia de usuario

Como integrante del equipo, quiero que los ítems del backlog, los sprints, los integrantes, los defectos, los registros de esfuerzo y las sesiones de Planning Poker se identifiquen con un UUID en lugar de un número correlativo, para que no se pueda adivinar ni recorrer ninguna entidad, ni deducir cuántas hay.

## Why

HU-51 cambió los proyectos, pero dejó como pendiente documentado el riesgo R1: las demás entidades siguen con IDs enteros correlativos. Quien recorre `/api/items/1`, `/api/items/2`… lee datos ajenos, llega al proyecto con el `proyecto_id`, y el último ID dice cuántas entidades hay. El profesor lo pidió en la revisión del 05/11: "No usar IDs 1, 2, 3, que son fáciles de adivinar; usar UUID". HU-52 cierra R1.

## What Changes

- **Identificador público (alternativa B de HU-51):** cada ítem, sprint, integrante, defecto, registro de esfuerzo y sesión de Planning Poker recibe, al crearse, un UUID versión 4 aleatorio generado por el sistema. Es el único identificador de la entidad que muestran la API y la interfaz. El entero se conserva como clave interna (clave primaria y claves foráneas) y deja de verse.
- **API, rutas:** toda ruta que recibe una de estas entidades por su `{id}` o su `{itemId}` acepta el UUID: 7 rutas de ítems, 3 que reciben el ítem por `{itemId}`, 9 de sprints, 1 de integrantes, 1 de defectos, 1 de esfuerzos y 4 de sesiones de Planning Poker.
- **API, cuerpos:** `sprint_deteccion_id` (HU-29) y `sprint_resolucion_id` (HU-30) pasan a recibir el UUID del sprint.
- **API, respuestas:** los 14 campos de respuesta que hoy llevan un ID entero de estas entidades pasan a llevar el UUID como texto: `id` de cada entidad, `sprint_id` (ítem, esfuerzo y métricas del sprint), `item_id` (defecto, esfuerzo y sesión) y los sprints de detección y de resolución del defecto. Un sprint no asignado o un defecto sin resolver se informan con `""` en lugar de `0`.
- **PDF del reporte (HU-40):** los IDs de historias, esfuerzos, defectos y sprints del texto pasan a ser sus UUID.
- **Errores:** un `{id}` o un campo del cuerpo que no tiene la forma canónica de un UUID responde 400 con un mensaje propio de cada entidad (por ejemplo, "el id del ítem debe ser un UUID"). Un UUID que no existe responde 404, y se conserva el orden de errores de cada ruta.
- **Base de datos:** las 6 tablas suman la columna `public_id UUID NOT NULL UNIQUE`. Las claves foráneas no cambian.
- **Interfaz web:** los 9 campos de ID numéricos («Sprint actual», «ID del integrante», el ítem y el esfuerzo de "Esfuerzos", la referencia de "Consultar esfuerzos", y el ítem, los sprints y el defecto de "Defectos") pasan a texto, y los botones que eligen desde una lista siguen funcionando.
- **BREAKING (API):** un cliente que usaba `/api/items/1` o esperaba `"id": 1` en estas entidades deja de funcionar: tiene que usar el UUID. La interfaz propia se adapta en este mismo cambio, y no hay otros clientes.
- **Datos existentes:** el esquema se recrea con `docker compose down -v`, como en HU-51 (D6). No hay migración.

## Capabilities

### New Capabilities

Una capability por entidad. Hoy solo existe `proyectos`; el comportamiento actual de cada entidad está en sus SDD de `docs/especificaciones/`. Cada capability nueva arranca con los requisitos de identificación de HU-52, y es el lugar donde migrar después esas SDD (D9 de HU-51). La decisión está en D2 de `design.md`.

- `backlog`: ítems del Product Backlog (HU-04 a HU-09, HU-20). Identificador público del ítem, rutas, respuestas que lo nombran e interfaz.
- `sprints`: sprints (HU-10 a HU-14). Identificador público del sprint, rutas, respuestas que lo nombran (también `sprint_id` de ítems, esfuerzos y métricas) e interfaz.
- `integrantes`: integrantes del proyecto (HU-15 a HU-17). Identificador público del integrante.
- `defectos`: defectos (HU-29 a HU-31). Identificador público del defecto, y los sprints de detección y de resolución por UUID en el cuerpo y en la respuesta.
- `esfuerzos`: registros de esfuerzo (HU-26 a HU-28). Identificador público del esfuerzo.
- `planning-poker`: sesiones de Planning Poker (HU-21 a HU-25). Identificador público de la sesión.

### Modified Capabilities

Ninguna. `proyectos` no cambia: HU-52 no toca el UUID ni las rutas de los proyectos.

## Vínculos y SDD leídas

- **HU-51** (`openspec/changes/archive/2026-11-15-hu-51-identificadores-uuid/` y `openspec/specs/proyectos/spec.md`): se repiten D1 (UUID público con entero interno), D2 (forma canónica), D3 (generación en el `Manager` con el paquete estándar `uuid`), D5 (ninguna salida con el entero) y D6 (recrear el esquema). D4 se adapta: un UUID inexistente se traduce a un ID interno que nunca se asigna, pero no al 0 (ver D4 de `design.md`).
- **Ítems:** HU-04 (RN-04.6, ID entero correlativo global), HU-05 a HU-09 y HU-20.
- **Sprints:** HU-10 (RN-10.6), HU-11 a HU-14. En los ítems, `sprint_id` 0 es "sin asignar" (HU-11).
- **Integrantes:** HU-15 (RN-15.7), HU-16 y HU-17.
- **Planning Poker:** HU-21 (RN-21.7), HU-22 a HU-25. Repetir la ronda (HU-24) actualiza la misma sesión, así que conserva su UUID.
- **Esfuerzos:** HU-26 (RN-26.8), HU-27 y HU-28.
- **Defectos:** HU-29 (RN-29.8), HU-30 y HU-31. Hoy, un `sprint_deteccion_id` ausente, 0 o negativo da 400, y uno positivo inexistente da 404. Lo mismo pasa con `sprint_resolucion_id` al resolver (HU-30).
- **Reportes:** HU-38, HU-39, HU-40 y HU-49 reutilizan las respuestas de estas entidades, y el PDF imprime sus IDs.

Los RN de "ID entero correlativo global" de cada SDD siguen valiendo para la clave **interna**. Hacia afuera, el identificador pasa a ser el UUID.

## Comparación con el acta de Planning

El acta del Planning del Sprint 5 (`docs/actas/sprint-5/planning.md`) no tiene a HU-52: planifica HU-51 (13 SP) por partes. Primero van los proyectos (obligatorio), y después, "si el equipo llega con tiempo", "HU-51 en el resto de las entidades (backlog, sprints, defectos y esfuerzo)". Lo que no llegue "queda escrito en la propuesta de OpenSpec como trabajo pendiente". HU-52 es esa segunda parte, ahora como historia propia con su Issue (#197).

**Hay tres diferencias, que el Agile Enabler decidió el 15/11/2026: se aprueban las 6 entidades (con el corte por entidad), el corte del lunes 16/11/2026 a las 14:00, y los Story Points se votan en Planning Poker.**

1. **Alcance:** el acta nombra 4 entidades (backlog, sprints, defectos y esfuerzo). El Issue #197, y esta propuesta, suman integrantes (HU-15 a HU-17) y sesiones de Planning Poker (HU-21 a HU-25). El Agile Enabler aprobó las 6 entidades, con el corte por entidad de "Si no llega al corte".
2. **Plazo:** el acta no prevé un corte después del 14/11 ("no hay Sprint siguiente"). El Issue fija el corte el lunes 16/11 a las 14:00 y extiende el Sprint 5 hasta esa hora.
3. **Story Points:** el acta cuenta este trabajo dentro de los 13 SP de HU-51. El Issue #197 dice "a estimar con Planning Poker".

## Criterios de aceptación definitivos

| Criterio | Texto | Cambio respecto del borrador del Issue |
|---|---|---|
| CA-52.1 | Al crearse, cada ítem, sprint, integrante, defecto, registro de esfuerzo y sesión de Planning Poker recibe un identificador público UUID versión 4 en forma canónica (36 caracteres, minúsculas). La API y la interfaz ya no muestran 1, 2, 3 para ninguna entidad. | Se agrega "público" y la forma canónica, como en CA-51.1: el entero interno sigue existiendo, pero no se expone. |
| CA-52.2 | Dos entidades del mismo tipo nunca tienen el mismo identificador público, y a partir de uno no se puede deducir otro ni la cantidad. | Se precisa "del mismo tipo": la unicidad la garantiza `UNIQUE` por tabla. Entre tablas distintas, un choque tiene probabilidad despreciable, y aunque ocurriera no confunde entidades, porque cada ruta sabe de qué tipo es. |
| CA-52.3 | Todas las rutas que reciben esas entidades aceptan su UUID (en la ruta y en los campos del cuerpo que nombran un sprint), y ninguna respuesta de la API ni el PDF exponen un ID entero de estas entidades, incluidos los campos que nombran a otra entidad. Un sprint sin asignar o sin resolución se informa con `""`. | Se agregan los cuerpos (`sprint_deteccion_id` y `sprint_resolucion_id`), el PDF y el valor `""` para "sin sprint", que hoy es `0`. |
| CA-52.4 | Un `{id}`, un `{itemId}` o un campo del cuerpo que no tiene la forma de un UUID responde 400, con el mensaje de su entidad ("el id del ítem debe ser un UUID", "el id del sprint debe ser un UUID", "el id del integrante debe ser un UUID", "el id del defecto debe ser un UUID", "el id del esfuerzo debe ser un UUID", "el id de la sesión debe ser un UUID"). Un UUID que no existe responde 404. Se conserva el orden de errores de cada ruta. | Se fijan los 6 mensajes y se agregan `{itemId}` y los campos del cuerpo. |
| CA-52.5 | Funciona igual en memoria y en PostgreSQL: cada una de las 6 tablas suma `public_id` (UUID, único), y las relaciones internas siguen con los enteros. | Sin cambios. |
| CA-52.6 | La interfaz web funciona con UUID: los 9 campos de ID aceptan texto, y los botones para elegir desde una lista siguen funcionando. | Se precisa la cantidad de campos, contada con comandos (D10 de `design.md`). |
| CA-52.7 | Crear entidades en simultáneo genera identificadores únicos, sin condiciones de carrera (`go test -race` en la CI). | Se agrega "en la CI": `-race` no corre en Windows. |

## Si no llega al corte

La base compartida (dominio, esquema y helpers, D3 de `design.md`) se integra primero, y cada entidad se integra por separado. Lo que a las 14:00 del 16/11 no esté en `main` con la CI en verde queda pendiente, por entidad:

| Si falta | Qué queda en `main` | Pendiente que se documenta |
|---|---|---|
| La base | Todo sigue como en HU-51: IDs enteros en las 6 entidades | R1 de HU-51 sigue abierto completo. Las entidades no se pueden integrar sin la base |
| Ítems (Cardozo Leandro) | Rutas y respuestas del ítem con ID entero. Los demás campos que nombran un ítem (`item_id` de defecto, esfuerzo y sesión) ya muestran el UUID si su entidad está integrada | Las 10 rutas de ítems y el `id` del ítem, y que el `item_id` de otras respuestas no sirve todavía para `/api/items/{id}` |
| Integrantes (Cardozo Leandro) | `PUT /api/integrantes/{id}` y el `id` del integrante con entero | 1 ruta, 1 campo y el campo «ID del integrante» |
| Sprints (Buttini Cristobal) | Las 9 rutas de sprints, `id` y `sprint_id` con entero, y «Sprint actual» numérico | Igual que en ítems, para los campos que nombran un sprint |
| Planning Poker (Buttini Cristobal) | Las 4 rutas de sesiones con entero | 4 rutas y el `id` de la sesión |
| Defectos (García Santiago) | `PATCH /api/defectos/{id}/estado`, el `id` y los sprints del cuerpo con entero | 1 ruta, 2 campos del cuerpo, 4 de respuesta y 4 de la interfaz |
| Esfuerzos (García Santiago) | `PATCH /api/esfuerzos/{id}/horas` y el `id` con entero | 1 ruta, 3 campos de respuesta y 3 de la interfaz |
| PDF del reporte (Peñalbé Hernán) | El texto del PDF (HU-40) con los IDs enteros de historias, esfuerzos, defectos y sprints | CA-52.3 en el PDF |

Cada pendiente se escribe en la sección "Pendiente al corte" de `docs/evidencias/HU-52-tdd.md`, con lo que quedó de la tabla de arriba, se avisa en la Review y queda como riesgo abierto en `design.md`. El Issue #197 se cierra solo si todas las entidades están integradas; si no, el commit de cierre usa `Refs #197` y queda abierto con los pendientes.

## Fuera de alcance

- Los proyectos (HU-51).
- Control de acceso, usuarios o permisos.
- Migrar datos de una base existente: el esquema se recrea.
- Aceptar otras formas de UUID (llaves, `urn:uuid:`, sin guiones).
- Cambiar los IDs internos o las firmas de los `Manager`: siguen con enteros (alternativa B).
- Migrar las SDD de `docs/especificaciones/` a `openspec/specs/`: las capabilities nuevas arrancan solo con los requisitos de HU-52.

## Impact

Cantidades contadas con comandos (D10 de `design.md`):

- **Dominio:** 6 paquetes (`backlog`, `sprint`, `member`, `defect`, `effort` y `poker`): el campo `PublicID`, la generación en `Create`, `Register` u `Open`, y `GetByPublicID` en el `Manager` y en los repositorios en memoria y en PostgreSQL.
- **API:** `internal/httpapi`: `server.go` (helpers y errores) y los handlers de `items.go`, `members*.go`, `sprints.go`, `poker.go`, `metrics.go`, `sprint_reports.go`, `defects.go`, `efforts.go`, `reports.go` y `dashboard.go`.
- **Reportes:** `internal/report/pdf.go` (a cargo de Peñalbé Hernán, decidido por el Agile Enabler el 15/11/2026, D3).
- **Esquema:** `db/init/01-schema.sql`, con 6 columnas.
- **Interfaz:** `internal/web/index.html` (9 campos) y `internal/web/app.js` (2 `Number(...)` de sprints en defectos, y los mensajes que muestran IDs).
- **Tests:** los IDs fijos de estas entidades están en 15 archivos de `internal/httpapi`, con 314 apariciones. Los tests de dominio no usan rutas, pero algunos comparan structs completos y se rompen con un `PublicID` aleatorio. Se miden con una prueba en la base, como en HU-51.
- **Documentación:** un `.feature` por entidad (6) y uno de la base, `docs/evidencias/HU-52-tdd.md`, `docs/reportes/cobertura.md`, README, manual de usuario, documentación técnica y `docs/ia/registro.md`.
- **Sin dependencias nuevas:** el paquete estándar `uuid` de Go 1.27.
