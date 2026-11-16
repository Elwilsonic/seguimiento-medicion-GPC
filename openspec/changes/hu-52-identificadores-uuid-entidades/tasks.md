# Tasks: HU-52 · Identificadores UUID del resto de las entidades (Issue #197)

Corte: lunes 16/11/2026 a las 14:00. Cada commit y cada push necesitan la autorización explícita de su responsable, con el comando exacto propuesto antes (lo decidió el Agile Enabler el 15/11/2026: cada responsable autoriza sus propios commits, como en HU-51). Cada test nuevo lleva arriba `// HU-52 · CA-52.N · Verifica que …`, y cada runner de godog `// HU-52 · BDD · Verifica que …`. Los tests adaptados conservan su comentario de TEC-11. Los merges con `main` se hacen con `git merge main` (nunca rebase), y los PR con "Create a merge commit".

Comandos de verificación de cada sección (los "comandos de cierre"):

```bash
gofmt -l .
go vet ./...
go test -count=1 ./...
TEST_DATABASE_URL='postgres://gpc:gpc_dev@localhost:5434/gpc_test?sslmode=disable' go test -count=1 -v ./... | grep -cE -- '--- (FAIL|SKIP)'   # 0
go test -count=1 ./internal/trazabilidad/
openspec validate hu-52-identificadores-uuid-entidades --strict
```

## 0. Aprobación y preparación

- [x] 0.1 El Agile Enabler aprueba explícitamente la propuesta completa (proposal, las 6 specs, design y tasks), incluidas las tres diferencias con el acta de Planning (alcance, corte y Story Points), D3 (el dominio de las 6 entidades en la base), D4 (el centinela `math.MaxInt32`), D5 (el resolvedor), D7 (`""` para "sin sprint"), la base en pareja (Sosa y Cardozo) y su horario (domingo 15/11/2026 a las 21:00). Verificar que la aprobación quedó en el chat o en el Issue #197. Aprobada por WhatsApp el 15/11/2026.
- [ ] 0.2 Actualizar los criterios del Issue #197 con los CA-52.1 a CA-52.7 definitivos, solo con el permiso del Agile Enabler. Verificar que el body del Issue los muestra.
- [x] 0.3 Registrar la aprobación en `proposal.md` (línea de estado) y proponer `git commit -m "docs(HU-52): registrar la aprobación de la propuesta" -m "Refs #197"`.

## 1. Base compartida (Sosa Ricardo y Cardozo Leandro en pareja, rama `feature/HU-52-uuid-base`)

La base la hacen Sosa Ricardo y Cardozo Leandro en pareja, sujeto a que Cardozo Leandro lo confirme. Horario: integrada en `main` el domingo 15/11/2026 a las 21:00, aprobado por el Agile Enabler.

- [x] 1.1 Medir antes del RED, y revertir: asignar `uuid.NewV4()` en las 6 creaciones y correr `go test -count=1 ./...`. Anotar qué tests de dominio fallan por comparar structs completos (como los 7 de HU-51); esos se adaptan en 1.4. Hecho antes del RED: la medición está en la evidencia.
- [x] 1.2 Escribir `features/HU-52-base.feature` (de dominio, en español): crear cada una de las 6 entidades le asigna un UUID v4 canónico; muchas creaciones dan UUID distintos; se busca por UUID y un UUID inexistente da el error de su entidad. Escribir los pasos y el runner `TestFeatures_HU52Base`. Verificar con `go test -run TestFeatures_HU52Base ./...` que corren y fallan solo por lo nuevo. Hecho en `2fb318d`.
- [x] 1.3 Escribir los tests en RED de dominio, en `public_id_test.go` de cada uno de los 6 paquetes: Hecho en `2fb318d`.
  - la creación asigna un UUID v4 canónico (CA-52.1);
  - 1000 creaciones dan UUID distintos y no correlativos (CA-52.2);
  - 50 creaciones simultáneas dan 50 UUID distintos (CA-52.7);
  - `GetByPublicID` con una entidad existente, inexistente y el nulo, y que devuelve una copia;
  - las modificaciones conservan el `PublicID` (CA-52.5).

  Además, el contrato de repositorios en memoria y en PostgreSQL ("guardar y buscar por UUID" y "UUID repetido rechazado en PostgreSQL", CA-52.5).
- [x] 1.4 Escribir los tests en RED de los helpers de `internal/httpapi` (`public_ids_internal_test.go`): Hecho en `2fb318d`; las rutas fijas se traducen en `do()` (D9).
  - `parsePublicID` con formas válidas y rechazadas;
  - el 400 de los 6 errores en `statusFor`;
  - `xIDFromPath` (existente, inexistente → `missingInternalID`, y mal formato → su error);
  - `sprintIDFromBody` (`""` → 0, mal formato → error, inexistente → centinela);
  - el resolvedor `publicIDs`, con `""` para el 0, cache y error a 500 con `writeResolved` (CA-52.3, CA-52.4).

  Adaptar los tests existentes, sin cambiar lo que verifican:
  - el registro de IDs de `do` para las 6 entidades, con helpers de rutas;
  - las rutas fijas de los 15 archivos de `internal/httpapi`, armadas con el `id` que devuelve la API y sin depender de su tipo;
  - las tablas de `server_test.go` separadas por entidad, con líneas en blanco entre bloques;
  - los tests de dominio de 1.1.
- [x] 1.5 Correr `go test -count=1 ./...` y verificar que fallan solo los tests de la base, por aserción. Proponer `git commit -m "test(HU-52): agregar tests de la base (RED)" -m "Refs #197"`. Hecho: RED en `2fb318d`.
- [x] 1.6 GREEN, dominio y esquema: `public_id UUID NOT NULL UNIQUE` en las 6 tablas de `db/init/01-schema.sql`. En los 6 paquetes: Hecho en `2bcd3c6`.
  - `PublicID`;
  - `uuid.NewV4()` al crear, dentro del lock;
  - `GetByPublicID` en el `Manager` y en los dos repositorios;
  - `public_id` en `Save` y en el `scan` de PostgreSQL.

  Recrear la base con `docker compose down -v` y `docker compose up -d --wait`, y verificar con PostgreSQL los tests de 1.3.
- [x] 1.7 GREEN, `internal/httpapi/server.go`: Hecho en `2bcd3c6`.
  - `parsePublicID`, con `parseProjectID` llamándola;
  - los 6 errores y su 400 en `statusFor`;
  - los 6 `xIDFromPath`, `sprintIDFromBody` y `missingInternalID`;
  - `publicIDs` y `writeResolved`.

  Agregar el resolvedor como parámetro de las funciones de respuesta compartidas (`toItemResponse`, `toSprintResponse`, `toEffortResponse`, `toEffortListResponse`, `toDefectResponse`, `toDefectListResponse`, `toSprintMetricsResponse` y `toStoryListResponse`) sin cambiar su salida, y actualizar todas sus llamadas.

  Correr los comandos de D4 adaptados (ningún código compara un ID con `math.MaxInt32`).
- [x] 1.8 GREEN, PDF: pasó a Peñalbé Hernán (sección 5, "PDF del reporte"), decidido por el Agile Enabler el 15/11/2026: cambia una salida, los tests de `internal/report` verifican el texto con enteros, y depende solo de la base (D3).
- [x] 1.9 Correr los comandos de cierre: todo en verde, en memoria y con PostgreSQL. Verificar con `git diff --stat` que el GREEN no toca tests ni `.feature`. Proponer `git commit -m "feat(HU-52): agregar la base de los identificadores UUID (GREEN)" -m "Refs #197"`. Hecho: GREEN en `2bcd3c6`.
- [x] 1.10 Refactor, si hay una mejora clara. Si no, anotar "no hubo refactor". En el GREEN de la base no hubo refactor. En el cierre, `6004613` quitó `pathID` y `pathInt` de `server.go`, que quedaron sin uso después de los refactors de defectos y esfuerzos.
- [x] 1.11 Crear `docs/evidencias/HU-52-tdd.md`: Hecho en `235b781` y completado con `docs(HU-52): completar las tareas y la documentación de la base` (Pendiente al corte, README y cobertura).
  - la sección de la base completa: RED, GREEN, salidas con conteos y la tabla criterio → test → código con links;
  - las secciones de las tres entidades, vacías y con su responsable;
  - la sección "Pendiente al corte".

  Agregar la fila de HU-52 al README y actualizar `docs/reportes/cobertura.md` y `docs/ia/registro.md`. Proponer `git commit -m "docs(HU-52): agregar la evidencia de la base" -m "Refs #197"`.
- [x] 1.12 Push, PR de la base hacia `main` y revisión de otro integrante, para integrarla el domingo 15/11/2026 a las 21:00. Antes del PR, traer `main` y correr `go test ./internal/trazabilidad/ -run TestEnlaces -args -update`. Avisar a los tres responsables cuando esté integrada. Si a esa hora no está, avisar al Agile Enabler y aplicar R1 de `design.md`. PR #201 hacia `main`: lo aprobó Buttini Cristobal y lo integró Cardozo Leandro el 15/11/2026 a las 18:30 (`c329a62`). Antes del PR, `main` no había avanzado desde `45d2c31`, así que no hizo falta el merge ni el `-update`. El commit de las tareas y la documentación de la base (`26afddc`) quedó fuera de ese merge y va en el PR #202, junto con este commit.

## 2. Ítems e integrantes (Cardozo Leandro, rama `feature/HU-52-uuid-items-integrantes`)

Se crea desde `main`, que tiene la base integrada desde el 15/11/2026 a las 18:30 (`c329a62`). Toca solo los archivos de su fila en D3.

- [x] 2.1 Escribir `features/HU-52-items.feature` y `features/HU-52-integrantes.feature` (de API, en español), con casos normal, alternativo, límite y de error por entidad, según `specs/backlog/spec.md` y `specs/integrantes/spec.md`. Escribir los pasos y los runners `TestFeatures_HU52Items` y `TestFeatures_HU52Integrantes`. Hecho en `2350f63`: 17 escenarios de ítems y 14 de integrantes, con los runners en `items_members_uuid_features_test.go`.
- [x] 2.2 Escribir los tests en RED en `internal/httpapi/items_uuid_test.go`, `members_uuid_test.go` y `items_members_uuid_web_test.go`:
  - el `id` es un UUID v4 al crear (CA-52.1);
  - sus 5 rutas (las 4 de `/api/items/{id}` en `items.go` y `PUT /api/integrantes/{id}`, tabla "Rutas por responsable" de D3) aceptan el UUID, también en mayúsculas (CA-52.3);
  - un número, llaves, `urn:uuid:` o sin guiones dan 400 con su mensaje, y un UUID inexistente o el nulo dan 404, en cada ruta (CA-52.4);
  - el orden de errores (título vacío en HU-07 y nombre vacío en HU-16, CA-52.4);
  - ninguna respuesta de ítem o integrante con un entero (CA-52.3);
  - los campos de la interfaz (CA-52.6).

  Adaptar las filas de ítems e integrantes en las tablas de `server_test.go` y los tests de `items_test.go`, `members*_test.go` cuyo resultado cambia. Correr `go test -count=1 ./...` y verificar que fallan solo estos tests, por aserción. Proponer `git commit -m "test(HU-52): agregar tests de ítems e integrantes (RED)" -m "Refs #197"`. Hecho: RED en `2350f63`. Además del plan se adaptaron los 7 JSON de integrantes con el id escrito a mano y, de HU-51, `TestRutasDeOtrasEntidades_ConservanElIDEntero`, `TestRespuestas_IdentificanElProyectoConSuUUID` y las líneas 86 y 144 de `features/HU-51-identificadores-uuid.feature`.
- [x] 2.3 GREEN:
  - los handlers de `items.go` (las 4 rutas de `/api/items/{id}`) y `members_update.go` usan `itemIDFromPath` y `memberIDFromPath`;
  - el cuerpo de `toItemResponse` usa el resolvedor para `id` y `sprint_id`;
  - las respuestas de `members*.go` usan el UUID del integrante;
  - las secciones de Product Backlog e Integrantes de `index.html` y `app.js`.

  Correr los comandos de cierre y verificar que el GREEN no toca tests. Proponer `git commit -m "feat(HU-52): identificar los ítems y los integrantes con un UUID público (GREEN)" -m "Refs #197"`. Hecho: GREEN en `ca39108`. Cambios respecto del plan: el GREEN sí tocó tests, por la segunda desviación de la base (la traducción inversa en `withEntityIDs`, con su test) y por los tests de otros responsables que cambian con el ítem en UUID (`sprints_test.go`, `reports_test.go`, `sprint_reports_test.go` y `web_test.go`); y `filaHistoriaReporte` de `app.js`, porque las historias de los reportes salen de `toItemResponse`. Después se trajo la rama de integración `feature/HU-52-uuid-entidades` (D3), con sprints y Planning Poker (PR #204), en el merge `bebcfc4`: quedan 13 fallos que dependen de la parte de García Santiago.
- [x] 2.4 Refactor, si aplica. Completar la sección de Cardozo en `docs/evidencias/HU-52-tdd.md` (RED, GREEN, curl, navegador y tabla con links), sumar su fila a `docs/ia/registro.md` y correr `-update` si hace falta. Proponer `git commit -m "docs(HU-52): agregar la evidencia de ítems e integrantes" -m "Refs #197"`. Hecho en `5264bbd`. No hubo refactor. No se verificó con curl ni en el navegador: queda como "no verificado" en la evidencia. El PR de esta parte va a `feature/HU-52-uuid-entidades` y no a `main` (cambio en D3).
- [x] 2.5 `git merge main`, comandos de cierre, push y PR, con la revisión de otro integrante. El `git merge main` quedó reemplazado por la rama de integración `feature/HU-52-uuid-entidades` (D3): la parte entró en esa rama con el PR #205 (`44e0936`).

## 3. Sprints y Planning Poker (Buttini Cristobal, rama `feature/HU-52-uuid-sprints-poker`)

Se crea desde `main`, que tiene la base integrada desde el 15/11/2026 a las 18:30 (`c329a62`). Toca solo los archivos de su fila en D3.

- [x] 3.1 Escribir `features/HU-52-sprints.feature` y `features/HU-52-planning-poker.feature` (de API), según `specs/sprints/spec.md` y `specs/planning-poker/spec.md`, con los runners `TestFeatures_HU52Sprints` y `TestFeatures_HU52PlanningPoker`.
- [x] 3.2 Escribir los tests en RED en `internal/httpapi/sprints_uuid_test.go`, `poker_uuid_test.go` y `sprints_poker_uuid_web_test.go`:
  - el `id` es un UUID v4 al crear un sprint o abrir una sesión, y repetir la ronda lo conserva (CA-52.1);
  - sus 11 rutas (tabla "Rutas por responsable" de D3: Planning Poker, las de `sprints.go`, `GET /api/sprints/{id}/metricas` y `/reporte`) aceptan el UUID del sprint, del ítem o de la sesión, también en mayúsculas (CA-52.3);
  - 400 con su mensaje y 404 en cada ruta, y `{id}` antes que `{itemId}` (CA-52.4);
  - el orden de errores (integrante vacío en HU-26 y votante vacío en HU-22, CA-52.4);
  - `sprint_id` con UUID o `""` en ítems, esfuerzos y métricas, `sprint_actual` del resumen y las métricas por sprint (CA-52.3);
  - «Sprint actual» y la sesión en la interfaz (CA-52.6).

  Adaptar las filas de sprints y sesiones de `server_test.go` y los tests de `sprints_test.go`, `poker_test.go`, `metrics_test.go`, `dashboard_test.go` y `sprint_reports_test.go` cuyo resultado cambia. Verificar el RED y proponer `git commit -m "test(HU-52): agregar tests de sprints y Planning Poker (RED)" -m "Refs #197"`.
- [x] 3.3 GREEN:
  - los handlers de `sprints.go` (`sprintItemIDs` y las rutas `/api/sprints/{id}/…`), `poker.go`, `metrics.go` (`sprintMetrics`) y `sprint_reports.go` (`sprintReport`) usan `sprintIDFromPath`, `itemIDFromPath` y `sessionIDFromPath`;
  - los cuerpos de `toSprintResponse` y `toSprintMetricsResponse`, y las respuestas de `poker.go`, usan el resolvedor;
  - «Sprint actual» y las secciones de Sprints y Planning Poker de `index.html` y `app.js`.

  Correr los comandos de cierre y proponer `git commit -m "feat(HU-52): identificar los sprints y las sesiones de Planning Poker con un UUID público (GREEN)" -m "Refs #197"`.
- [x] 3.4 Refactor, si aplica. Completar la sección de Buttini en la evidencia y su fila en el registro de IA, y correr `-update` si hace falta. Proponer `git commit -m "docs(HU-52): agregar la evidencia de sprints y Planning Poker" -m "Refs #197"`.
- [x] 3.5 `git merge main`, comandos de cierre, push y PR, con la revisión de otro integrante. El `git merge main` quedó reemplazado por la rama de integración `feature/HU-52-uuid-entidades` (D3): la parte entró en esa rama con el PR #204 (`041889c`).

## 4. Defectos y esfuerzos (García Santiago, rama `feature/HU-52-uuid-defectos-esfuerzos`)

Se crea desde `main`, que tiene la base integrada desde el 15/11/2026 a las 18:30 (`c329a62`). Toca solo los archivos de su fila en D3.

- [x] 4.1 Escribir `features/HU-52-defectos.feature` y `features/HU-52-esfuerzos.feature` (de API), según `specs/defectos/spec.md` y `specs/esfuerzos/spec.md`, con los runners `TestFeatures_HU52Defectos` y `TestFeatures_HU52Esfuerzos`. Hecho en `2b9dd75`, con el runner en `internal/httpapi/defects_efforts_uuid_features_test.go`.
- [x] 4.2 Escribir los tests en RED en `internal/httpapi/defects_uuid_test.go`, `efforts_uuid_test.go` y `defects_efforts_uuid_web_test.go`: Hecho en `2b9dd75`; los tests existentes se terminaron de adaptar en `c20c5bc` y `3e55461`.
  - el `id` es un UUID v4 al registrar, y `item_id` y los sprints llevan UUID, con `""` para "sin resolver" (CA-52.1 y CA-52.3);
  - sus 7 rutas (tabla "Rutas por responsable" de D3: las de `defects.go` y `efforts.go`, que reciben un defecto, un esfuerzo, un ítem o un sprint) aceptan el UUID (CA-52.3);
  - `sprint_deteccion_id` y `sprint_resolucion_id` con UUID: válido, mal formato (400 del sprint), vacío (400 de sprint faltante) e inexistente (404) (CA-52.4);
  - 400 y 404 en cada ruta, y el orden de errores (sprint de resolución faltante en HU-30 y horas inválidas en HU-28, CA-52.4);
  - los 7 campos de las secciones de Esfuerzos y Defectos, sin `Number(...)` para los sprints (CA-52.6).

  Adaptar las filas de defectos y esfuerzos de `server_test.go`, las 4 aserciones de `web_test.go` que exigen `type="number"` y los tests de `defects_test.go`, `defect_lists_test.go`, `efforts*_test.go` y `effort_lists_test.go` cuyo resultado cambia. Verificar el RED y proponer `git commit -m "test(HU-52): agregar tests de defectos y esfuerzos (RED)" -m "Refs #197"`.
- [x] 4.3 GREEN: Hecho en `5d32e7c`.
  - los handlers de sus 7 rutas, en `defects.go` y `efforts.go`, usan `defectIDFromPath`, `effortIDFromPath`, `itemIDFromPath`, `sprintIDFromPath` y `sprintIDFromBody`;
  - los cuerpos de `toDefectResponse`, `toDefectListResponse`, `toEffortResponse` y `toEffortListResponse` usan el resolvedor;
  - las secciones de Esfuerzos y Defectos de `index.html` y `app.js`.

  Correr los comandos de cierre y proponer `git commit -m "feat(HU-52): identificar los defectos y los esfuerzos con un UUID público (GREEN)" -m "Refs #197"`.
- [x] 4.4 Refactor, si aplica. Completar la sección de García en la evidencia y su fila en el registro de IA, y correr `-update` si hace falta. Proponer `git commit -m "docs(HU-52): agregar la evidencia de defectos y esfuerzos" -m "Refs #197"`. Refactor en `00276f4` (quitar la traducción inversa de `withEntityIDs`) y `f8b5b88` (quitar `sprintItemIDs`); la evidencia, en `20e29b4`; la fila del registro de IA está en `docs/ia/registro.md`.
- [x] 4.5 `git merge main`, comandos de cierre, push y PR, con la revisión de otro integrante. El `git merge main` quedó reemplazado por la rama de integración `feature/HU-52-uuid-entidades` (D3), traída en `63eb989` y `44b576b`; la parte entró en esa rama con el PR #206 (`9e0f322`).

## 5. PDF del reporte (Peñalbé Hernán, rama `feature/HU-52-uuid-pdf`)

Decidido por el Agile Enabler el 15/11/2026 (D3). Se creó desde `feature/HU-52-uuid-base` con su GREEN pusheado y se integró en esa rama (PR #200), que entró en `main` con el PR #201 (`c329a62`). Fue en paralelo con las entidades: depende solo de la base. Toca solo `internal/report` (su fila en D3).

- [x] 5.1 RED: adaptar los tests de `internal/report` que verifican el texto del PDF con enteros ("#1 … sprint 1") y agregar los que piden el UUID del ítem y del sprint en las líneas de historias, esfuerzos y defectos (HU-40, CA-52.3), con `// HU-52 · CA-52.3 · Verifica que …`. Verificar con `go test ./internal/report/` que fallan solo esos tests, por aserción. Proponer `git commit -m "test(HU-52): pedir los UUID en el PDF del reporte (RED)" -m "Refs #197"`. Hecho en `3229e3f`.
- [x] 5.2 GREEN: los IDs de historias, esfuerzos, defectos y sprints del texto de `internal/report/pdf.go` pasan a ser sus UUID; el reporte trae los UUID del ítem y del sprint. Correr los comandos de cierre, en memoria y con PostgreSQL, y verificar con `git diff --stat` que el GREEN no toca tests. Proponer `git commit -m "feat(HU-52): mostrar los UUID en el PDF del reporte (GREEN)" -m "Refs #197"`. Hecho en `4461a88`. El GREEN tocó tests y dos `.feature` (las líneas largas del PDF, los marcadores de UUID en HU-40 y HU-49, y `saveDefect`), como explica el PR #200.
- [x] 5.3 Refactor, si aplica. Completar la sección del PDF en la evidencia y su fila en el registro de IA, y correr `-update` si hace falta. Proponer `git commit -m "docs(HU-52): agregar la evidencia del PDF del reporte" -m "Refs #197"`. No hubo refactor. La sección del PDF de la evidencia la escribió Peñalbé Hernán y se agregó con `docs(HU-52): agregar la sección del PDF a la evidencia`. Falta su fila en `docs/ia/registro.md`, que todavía no tiene filas de HU-52.
- [x] 5.4 `git merge main`, comandos de cierre, push y PR, con la revisión de otro integrante. PR #200 hacia `feature/HU-52-uuid-base`, integrado en `0a6c513`.

## 6. Cierre (Sosa Ricardo)

- [x] 6.1 Con las cinco partes integradas, o a las 14:00 del 16/11 lo que haya, completar en la evidencia la sección "Pendiente al corte", según la tabla "Si no llega al corte" de `proposal.md`, y actualizar el riesgo R2 de `design.md`. Hecho el 15/11/2026: no quedó nada pendiente (las 6 entidades, la base y el PDF llegaron), y R2 quedó cerrado en `design.md`.
- [x] 6.2 Actualizar el manual de usuario y la documentación técnica (TEC-06 y TEC-07: los UUID de las 6 entidades y los errores nuevos), y la cobertura con PostgreSQL comparada con `main`. Hecho con `docs(HU-52): cerrar las tareas y la documentación de HU-52`: manual de usuario, documentación técnica y cobertura de la rama completa con PostgreSQL, igual que `main` (95.1% y 95.9%).
- [ ] 6.3 Agregar al Issue #197 la sección "## Trazabilidad", con los links a la propuesta, a los 7 `.feature` y a la evidencia. Proponer el commit de cierre: con `Closes #197` si llegaron todas las entidades, y con `Refs #197` si quedó algo pendiente.

## Workflow follow-up

- Con todo integrado en `main`, `/opsx:archive hu-52-identificadores-uuid-entidades` crea las 6 specs principales (`openspec/specs/{backlog,sprints,integrantes,defectos,esfuerzos,planning-poker}/spec.md`). Va en el commit `docs(HU-52): archivar el cambio en OpenSpec`, con autorización.
- Después del archivo, actualizar los links de la Trazabilidad del Issue #197 y del README a los definitivos.
- Avisar al equipo que recree la base local con `docker compose down -v` después de cada merge que cambie el esquema (solo la base).
