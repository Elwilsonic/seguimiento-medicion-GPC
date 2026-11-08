# Propuesta: HU-50 · Listar los proyectos

> Issue del repositorio: #183 · Épica E1 · Gestión de proyectos · Vinculada a HU-03 (#16)

## Historia de usuario

Como integrante del equipo, quiero ver la lista de todos los proyectos, para elegir uno sin tener que saber su número.

## Why

Hoy un proyecto solo se puede consultar por su ID (HU-03): la interfaz pide escribir el número en «Proyecto actual», y no hay forma de saber qué proyectos existen. El profesor lo marcó en la revisión del 05/11/2026 (Planning del Sprint 5, Issue #182). Además, HU-51 va a reemplazar los IDs numéricos por UUID, que nadie puede recordar ni escribir: sin una lista, la aplicación dejaría de poder usarse.

## What Changes

- **Dominio (`internal/project`):** una operación nueva del `Manager` que devuelve todos los proyectos como copias, en una lista nueva, vacía y no `nil` cuando no hay proyectos. Reutiliza el `List()` que el repositorio ya tiene, en memoria y en PostgreSQL; no se agrega nada a la interfaz `Repository` ni al esquema de la base.
- **Orden:** alfabético por nombre, sin distinguir mayúsculas (decisión D2 de `design.md`, pensada para HU-51).
- **API:** la ruta nueva `GET /api/proyectos`, que responde un arreglo JSON con cada proyecto en el mismo formato que `GET /api/proyectos/{id}`, y `[]` si no hay proyectos.
- **Interfaz web:** en la sección «Proyecto y sprint actuales», el botón «Ver proyectos» llena una tabla con los proyectos; el botón «Elegir» de cada fila lo deja como proyecto actual y lo carga, igual que «Cargar proyecto», sin escribir su número.
- **Documentación:** filas nuevas en el README (manual de uso, tabla de la API y tabla de historias).

No hay cambios incompatibles (**BREAKING**): `POST /api/proyectos` y `GET /api/proyectos/{id}` quedan iguales, y el campo «Proyecto actual (ID)» se conserva.

## Capabilities

### New Capabilities

- `proyectos`: gestión de proyectos (épica E1). Este cambio le agrega el primer requisito, la consulta de la lista de proyectos; las historias anteriores de proyectos (HU-01, HU-02, HU-03, HU-18, HU-19 y HU-41) siguen documentadas en sus SDD de `docs/especificaciones/` y no se migran.

### Modified Capabilities

Ninguna: `openspec/specs/` todavía está vacío (HU-50 es el primer cambio con OpenSpec).

## Vínculos y SDD leídas

- **HU-03** (`docs/especificaciones/HU-03-consultar-proyecto.md`): su "No incluye" deja afuera "listar proyectos"; HU-50 completa esa consulta. Se reutilizan el formato de respuesta de `GET /api/proyectos/{id}` y la regla de copias por valor (RN-03.3) y de mutex con test bajo `-race` (CA-03.4).
- **HU-01** (`HU-01-crear-proyecto.md`): ya usa `Repository.List()` para validar el nombre único (RN-01.3). HU-50 no cambia ese uso. Como los nombres no se repiten, sin distinguir mayúsculas, el orden por nombre no tiene empates.
- **HU-17** (`HU-17-consultar-integrantes.md`) y **HU-14** (`HU-14-sprints-anteriores.md`): se sigue su forma de listar: lista nueva de copias, vacía y no `nil`, `[]` por HTTP, error del repositorio → 500, y en la interfaz un botón «Ver …» que llena una tabla con un botón por fila y un mensaje "N …(s)".
- **HU-18, HU-41 y HU-19:** las fechas de inicio y de fin salen en el formato que ya usa `GET /api/proyectos/{id}` (`AAAA-MM-DD`, o `""` si no están registradas).

## Fuera de alcance

- Filtros (por estado, por fechas), búsqueda por nombre y paginación.
- Otro orden que no sea el de D2, o un orden elegido por el usuario.
- Cambios en la interfaz `Repository`, en el esquema de la base o en el tipo de los IDs (eso es HU-51).
- Recargar la lista sola al crear o modificar un proyecto: se actualiza con «Ver proyectos».
- Quitar el campo «Proyecto actual (ID)» de la interfaz.

## Impact

- **Código:** `internal/project` (archivo nuevo `list.go`), `internal/httpapi` (`projects.go` y una ruta en `server.go`), `internal/web` (`index.html` y `app.js`).
- **Tests:** `internal/project` (unitarios y pasos godog), `internal/httpapi` (`projects_list_test.go` nuevo y tests de la interfaz en un archivo propio, para no chocar con TEC-11 en `web_test.go`).
- **Documentación:** `features/HU-50-listar-proyectos.feature`, `docs/evidencias/HU-50-tdd.md`, `docs/reportes/cobertura.md`, `README.md`.
- **Coordinación con HU-51 (Sosa Ricardo):** HU-51 va a tocar las mismas rutas y el mismo campo de la interfaz. El orden por nombre no depende del ID, así que la lista sigue siendo correcta cuando los IDs sean UUID; HU-51 solo tendrá que cambiar el tipo del campo `id` de la respuesta.
- **Sin dependencias nuevas.**
