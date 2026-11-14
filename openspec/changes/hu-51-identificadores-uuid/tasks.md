# Tasks: HU-51 · Identificadores UUID, parte de proyectos (Issue #193)

Rama: `feature/HU-51-uuid`. Cada commit y cada push necesitan la autorización explícita del Agile Enabler: se propone el comando exacto y se espera el OK. Cada test nuevo lleva arriba `// HU-51 · CA-51.N · Verifica que …`. Los tests adaptados conservan su comentario de TEC-11.

## 0. Aprobación y preparación

- [ ] 0.1 El Agile Enabler aprueba explícitamente esta propuesta (proposal, specs, design y tasks), con los criterios CA-51.1 a CA-51.7 ajustados. La alternativa B (D1), D4 y el pendiente R1 ya están aprobados. Verificar que la aprobación quedó en el chat o en el Issue #193 antes de seguir.
- [ ] 0.2 Actualizar los criterios del Issue #193 con los CA-51.1 a CA-51.7 definitivos, solo con el permiso del Agile Enabler. Verificar que el body del Issue los muestra.
- [ ] 0.3 Correr `git switch main`, `git pull` y `git switch feature/HU-51-uuid`, y después `git merge main` (con autorización), para trabajar sobre los tests ya comentados por TEC-11 (R5). Verificar con `go test ./...` que la rama arranca en verde.

## 1. Escenarios BDD (godog)

- [ ] 1.1 Escribir `features/HU-51-identificadores-uuid.feature` en español (`# language: es`), a nivel de la API, con un escenario por cada caso de la spec, cada uno con su CA en el título:
  - crear con UUID v4;
  - el primer proyecto no es el 1;
  - consultar por UUID;
  - UUID en mayúsculas;
  - operaciones que dependen del proyecto;
  - ítem consultado por su ID con `proyecto_id` UUID;
  - ID numérico → 400;
  - otras formas de UUID → 400;
  - UUID inexistente → 404;
  - UUID nulo → 404;
  - orden de errores conservado;
  - las demás entidades conservan su ID entero.

  Verificar que hay al menos un caso normal, uno alternativo, uno límite y uno de error.
- [ ] 1.2 Implementar los pasos en `internal/httpapi/uuid_features_test.go`, con el runner `TestFeatures_HU51IdentificadoresUUID`, sobre `NewHandler` con repositorios en memoria y `httptest`. Verificar con `go test ./internal/httpapi/ -run TestFeatures_HU51` que los escenarios se ejecutan y fallan solo por el comportamiento nuevo (por ejemplo, `id` numérico o 404 en lugar de 400).

## 2. Tests en RED

- [ ] 2.1 Escribir `internal/project/public_id_test.go`:
  - `Create` asigna un UUID v4 canónico (CA-51.1);
  - 1000 proyectos con UUID distintos y no correlativos (CA-51.2);
  - `GetByPublicID` con proyecto existente, inexistente y nulo, y que devuelve una copia (CA-51.3, CA-51.4);
  - 50 creaciones simultáneas con UUID distintos (CA-51.7);
  - `Update`, `SetStartDate`, `SetEndDate` y `SetDates` conservan el `PublicID`.
- [ ] 2.2 Sumar a `internal/project/repository_contract_test.go` "guardar y buscar por UUID" e "UUID inexistente → `ErrProjectNotFound`", para memoria y PostgreSQL (CA-51.5). Sumar a `postgres_repository_test.go` el rechazo de un `public_id` repetido (CA-51.5). Adaptar los 2 repositorios falsos de `create_test.go` y `get_test.go` a la operación nueva de `Repository`.
- [ ] 2.3 Escribir `internal/httpapi/projects_uuid_test.go`:
  - `parseProjectID`: canónico, mayúsculas, número, llaves, `urn:uuid:`, sin guiones, vacío y dígito no hexadecimal (CA-51.4);
  - 400 y 404 en cada una de las 17 rutas de proyecto (CA-51.3, CA-51.4);
  - orden de errores: ID inválido con cuerpo inválido → 400 de ID, y UUID inexistente con fecha vacía → 400 de fecha (CA-51.4);
  - `GET /api/items/abc` conserva "el id debe ser un número entero";
  - cada respuesta que identifica un proyecto (`id`, `proyecto_id` y `Content-Disposition` del PDF) lleva un UUID y no un número (CA-51.3).
- [ ] 2.4 Escribir `internal/httpapi/projects_uuid_web_test.go`: `#proyecto-actual` es `type="text"`, sin `min`, y `app.js` no convierte el ID del proyecto a número (CA-51.6).
- [ ] 2.5 Adaptar los 20 tests de `internal/httpapi` que usan `/api/proyectos/1` o `"proyecto_id":1`, sin perder lo que verifican: crear el proyecto, usar su `id` en la ruta y comparar `proyecto_id` con ese UUID. La lista sale de `git ls-files '*_test.go' | xargs grep -lE '/api/proyectos/[0-9]|/api/proyectos/%d|"proyecto_id"|proyecto_id\\":|"/api/proyectos/" *\+|proyecto-actual'`. Verificar que no queda ninguna ruta con un ID entero de proyecto: el `grep` con `/api/proyectos/[0-9]` no devuelve nada.
- [ ] 2.6 Correr `go test ./...` y verificar que fallan solo los tests de HU-51 (`PublicID` o `GetByPublicID` no declarados, o aserciones de UUID), sin errores preexistentes. Mostrar la salida y proponer `git commit -m "test(HU-51): agregar tests (RED)" -m "Refs #193"`.

## 3. Código en GREEN

- [ ] 3.1 Dominio (`internal/project`):
  - `PublicID uuid.UUID` en `Project` (`project.go`);
  - `GetByPublicID` en `Repository` (`repository.go`) y en `MemoryRepository` (`memory_repository.go`);
  - `uuid.NewV4()` en `Manager.Create`, con el mutex (`create.go`);
  - `Manager.GetByPublicID` (`get.go`).

  Verificar con `go test -race ./internal/project/` que pasan los tests de 2.1 y 2.2 con memoria.
- [ ] 3.2 PostgreSQL: `public_id UUID NOT NULL UNIQUE` en `db/init/01-schema.sql`. En `postgres_repository.go`: `public_id` en `projectColumns`, en el `INSERT` y en `scanProject`, y `GetByPublicID`. Recrear la base local con `docker compose down -v` y `docker compose up -d`, y verificar con `go test ./internal/project/` (con `TEST_DATABASE_URL`) que pasa el contrato en PostgreSQL. Si pgx no escanea `uuid.UUID`, aplicar la mitigación de R3.
- [ ] 3.3 API, entrada (`internal/httpapi/server.go`): `errInvalidProjectID` (400 en `statusFor`), `parseProjectID` y `projectIDFromPath` (D2, D4). Reemplazar `pathID(r)` por `s.projectIDFromPath(r)` en las 17 rutas de proyecto (`projects.go`, `items.go`, `members.go`, `members_list.go`, `sprints.go`, `defects.go`, `metrics.go`, `dashboard.go`, `reports.go` y `report_pdf.go`). Verificar que pasan los tests de 400 y 404 y de orden de 2.3. Volver a correr los comandos de verificación del proyecto 0 de D4 y comprobar que los comandos 1 a 3 siguen sin resultados.
- [ ] 3.4 API, salida (D5):
  - `projectResponse.ID` pasa a `string`;
  - `ProyectoID string` en `itemResponse`, `sprintResponse`, `memberResponse`, métricas y `summaryResponse`;
  - `toItemResponse` y `toSprintResponse` reciben el UUID;
  - `publicProjectID` para las rutas por ID de otra entidad (`items.go`, `sprints.go`, `members_update.go`, `metrics.go` y `sprint_reports.go`);
  - `Content-Disposition` con el UUID en `report_pdf.go`.

  Verificar con `grep -n "ProyectoID +int" internal/httpapi/*.go` (sin resultados) y con los tests de salida de 2.3.
- [ ] 3.5 Interfaz: `#proyecto-actual` con `type="text"`, sin `min`, con `placeholder="UUID del proyecto"`, `autocomplete="off"` y `spellcheck="false"`, y el comentario de la sección con HU-51 (`internal/web/index.html`). Verificar que pasan los tests de 2.4 y probar a mano con `go run ./cmd/server`:
  - crear un proyecto;
  - «Ver proyectos» y «Elegir»;
  - escribir "1" (error);
  - descargar el PDF (nombre con el UUID).
- [ ] 3.6 Correr `go test ./...` y `go test -race ./...` y verificar que todo está en verde, incluidos los escenarios de HU-50 (el orden no cambia). Mostrar la salida y proponer `git commit -m "feat(HU-51): implementar identificadores UUID para proyectos (GREEN)" -m "Refs #193"`.

## 4. Refactor

- [ ] 4.1 Revisar si hay una mejora clara, por ejemplo un helper único para "UUID de la ruta o del ítem" si `publicProjectID` se repite en muchos handlers. Si se aplica, verificar con `go test ./...` que sigue en verde y proponer `git commit -m "refactor(HU-51): <qué se mejoró>" -m "Refs #193"`. Si no, dejar asentado en la evidencia que no hizo falta.

## 5. Evidencia, cobertura y documentación

- [ ] 5.1 Correr `gofmt -w` sobre los archivos Go modificados, `go test ./...` y `go test -cover ./...`, y verificar que `gofmt -l .` no lista archivos.
- [ ] 5.2 Escribir `docs/evidencias/HU-51-tdd.md` con:
  - las salidas de RED, GREEN y de los comandos de 5.1;
  - la cantidad exacta de tests nuevos, de tests adaptados y de escenarios BDD, por separado;
  - la tabla de commits con "(propuesto)";
  - la tabla de trazabilidad con las columnas Criterio (CA-51.N), Test (nombre y archivo) y Código que lo implementa (función y archivo).

  Verificar que cada CA-51.1 a CA-51.7 tiene al menos una fila.
- [ ] 5.3 Actualizar `docs/reportes/cobertura.md` con la cobertura nueva y verificar que coincide con la salida de `go test -cover ./...`.
- [ ] 5.4 Actualizar la documentación:
  - `README.md`:
    - manual: «Proyecto actual» con UUID;
    - tabla de la API: `{id}` de proyecto como UUID, con 400 y 404;
    - nota de `docker compose down -v`;
    - fila de HU-51 con links a la propuesta, al BDD y a la evidencia;
  - `docs/manual-de-usuario.md`: el UUID en «Proyecto actual».

  Verificar que los links abren en la rama.
- [ ] 5.5 Registrar el uso de la IA en `docs/ia/registro.md` y proponer `git commit -m "docs(HU-51): agregar evidencia TDD y cobertura" -m "Refs #193"`. Reemplazar "(propuesto)" por el hash real de cada commit apenas se ejecute.

## 6. Pull Request y trazabilidad

- [ ] 6.1 Con los commits autorizados y pusheados, abrir el Pull Request de `feature/HU-51-uuid` hacia `main`, con `Closes #193` en la descripción si no queda nada por cambiar, y pedir la revisión de otro integrante. Verificar que la CI está en verde. La CI aplica `01-schema.sql` sobre una base nueva, así que prueba `public_id`.
- [ ] 6.2 Agregar al Issue #193 la sección "## Trazabilidad" con los links (a `blob/main`) a la propuesta (`openspec/changes/hu-51-identificadores-uuid/proposal.md`), al BDD y a la evidencia. Verificar que los tres archivos existen en la rama.

## Workflow follow-up

- El merge lo hace el equipo con "Create a merge commit" (nunca squash ni rebase), con la aprobación del revisor y la CI en verde.
- Después del merge, avisar al equipo que recree su base local con `docker compose down -v` y `docker compose up -d`, y revisar que el script de carga de datos de la demo no use `/api/proyectos/1` (Open Question de `design.md`).
- Correr `/opsx:archive` solo cuando el PR está integrado en `main` y todas las tareas de este archivo están tildadas. Proponer `git commit -m "docs(HU-51): archivar el cambio en OpenSpec" -m "Refs #193"`. Los requisitos pasan a `openspec/specs/proyectos/spec.md`; si HU-50 ya se archivó, se suman a los suyos.
- Actualizar los links de la fila de HU-51 en el README y en la Trazabilidad del Issue #193 a los definitivos (`openspec/specs/proyectos/spec.md` y la carpeta del archivo), y verificar que abren en `main`.
- Parte pendiente de HU-51 (ítems, sprints, integrantes, defectos, esfuerzos y Planning Poker): no hay Sprint siguiente. Queda escrita en `proposal.md` y se menciona en la Review del 15/11.
