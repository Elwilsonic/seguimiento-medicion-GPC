# Tasks: HU-50 · Listar los proyectos (Issue #183)

Rama: `feature/HU-50-listar-proyectos`. Cada commit y cada push necesitan la autorización explícita del Agile Enabler: se propone el comando exacto y se espera el OK. Cada test lleva arriba el comentario `// HU-50 · CA-50.N · Verifica que …`.

## 0. Aprobación

- [x] 0.1 Peñalbé Hernán (Agile Enabler) aprueba explícitamente esta propuesta (proposal, specs, design y tasks); verificar que la aprobación quedó en el chat o en el Issue #183 antes de seguir.
- [x] 0.2 Actualizar los criterios del Issue #183 con CA-50.1 a CA-50.7 de la spec, solo con el permiso del Agile Enabler; verificar que el body del Issue los muestra.

## 1. Escenarios BDD (godog)

- [x] 1.1 Escribir `features/HU-50-listar-proyectos.feature` en español (`# language: es`), con un escenario por cada caso de la spec a nivel del `Manager`: normal, proyecto modificado, orden, único proyecto, sin proyectos, copia y error del repositorio, cada uno con su CA en el título. Verificar que hay al menos un caso normal, uno alternativo, uno límite y uno de error.
- [x] 1.2 Implementar los pasos en `internal/project/list_features_test.go`, con el runner `TestFeatures_HU50ListarProyectos` y un repositorio falso que falla al listar. Verificar con `go test ./internal/project/ -run TestFeatures_HU50` que los escenarios se ejecutan y fallan solo porque `Manager.List` todavía no existe.

## 2. Tests en RED

- [x] 2.1 Escribir `internal/project/list_test.go`:
  - lista con todos los datos (CA-50.1);
  - lista vacía y no `nil` con `MemoryRepository` (CA-50.2);
  - copia: modificar la lista no cambia `Get` ni una segunda `List` (CA-50.4);
  - error del repositorio con `nil` (CA-50.5);
  - orden por nombre sin distinguir mayúsculas (CA-50.6);
  - concurrencia con `Create` (CA-50.7).
- [x] 2.2 Escribir `internal/httpapi/projects_list_test.go`:
  - 200 con el formato exacto de `GET /api/proyectos/{id}`, fechas incluidas, y en orden (CA-50.1, CA-50.5, CA-50.6);
  - 200 con el cuerpo `[]` (CA-50.2);
  - 500 con `{"error":"error interno del servidor"}` usando un repositorio que falla (CA-50.5).
- [x] 2.3 Escribir `internal/httpapi/projects_list_web_test.go`:
  - la sección de contexto tiene `#ver-proyectos`, la tabla con sus columnas y `#lista-proyectos` (CA-50.3);
  - el handler llama a `api("GET", "/proyectos")` dentro de `accion("msg-contexto"`, usa "Elegir", `cargarProyecto` y `" proyecto(s)."`, y no usa `innerHTML` (CA-50.3, CA-50.5).
- [x] 2.4 Correr `go test ./...` y verificar que fallan solo los tests de HU-50 (`m.List undefined`, ruta inexistente o textos ausentes en la interfaz), sin errores preexistentes. Mostrar la salida y proponer `git commit -m "test(HU-50): agregar tests (RED)" -m "Refs #183"`.

## 3. Código en GREEN

- [ ] 3.1 Crear `internal/project/list.go` con `Manager.List()`: mutex, `nil` → `[]Project{}`, orden de D2 y error sin lista. Verificar con `go test -race ./internal/project/` que pasan los tests y los escenarios de HU-50.
- [x] 3.2 Agregar `listProjects` en `internal/httpapi/projects.go` y la ruta `GET /api/proyectos` en `server.go`. Verificar que pasan los tests de `projects_list_test.go`.
- [x] 3.3 Agregar a `internal/web/index.html` el botón «Ver proyectos» y la tabla, y actualizar el comentario de la sección de contexto. Agregar a `internal/web/app.js` el handler de `ver-proyectos` con el botón «Elegir». Verificar que pasan los tests de `projects_list_web_test.go` y probar la interfaz a mano con el servidor local (elegir un proyecto, lista vacía).
- [ ] 3.4 Correr `go test ./...` y `go test -race ./...`, y verificar que todo está en verde. Mostrar la salida y proponer `git commit -m "feat(HU-50): implementar la lista de proyectos (GREEN)" -m "Refs #183"`.

## 4. Refactor

- [x] 4.1 Revisar si hay una mejora clara de legibilidad (por ejemplo, armar las filas de la tabla con una función como en HU-14). Si se aplica, verificar con `go test ./...` que sigue en verde y proponer `git commit -m "refactor(HU-50): <qué se mejoró>" -m "Refs #183"`; si no, dejar asentado en la evidencia que no hizo falta.

## 5. Evidencia, cobertura y README

- [x] 5.1 Correr `gofmt -w` sobre los archivos Go modificados, `go test ./...` y `go test -cover ./...`, y verificar que `gofmt -l .` no lista archivos.
- [x] 5.2 Escribir `docs/evidencias/HU-50-tdd.md` con:
  - las salidas de RED, GREEN y de los comandos de 5.1;
  - la cantidad exacta de tests y de escenarios BDD por separado;
  - la tabla de commits con "(propuesto)";
  - la tabla de trazabilidad con las columnas Criterio (CA-50.N), Test (nombre y archivo) y Código que lo implementa (función y archivo).

  Verificar que cada CA-50.1 a CA-50.7 tiene al menos una fila.
- [x] 5.3 Actualizar `docs/reportes/cobertura.md` con la cobertura nueva y verificar que coincide con la salida de `go test -cover ./...`.
- [x] 5.4 Actualizar `README.md`:
  - manual: «Ver proyectos» y «Elegir», con la nota del orden y de los acentos;
  - tabla de la API: `GET /api/proyectos`;
  - tabla de historias: la fila de HU-50 con links a la propuesta, al BDD y a la evidencia.

  Verificar que los links abren en la rama.
- [x] 5.5 Registrar el uso de la IA en `docs/ia/registro.md` y proponer `git commit -m "docs(HU-50): agregar evidencia TDD y cobertura" -m "Refs #183"`. Reemplazar "(propuesto)" por el hash real de cada commit apenas se ejecute.

## 6. Pull Request y trazabilidad

- [ ] 6.1 Con los commits autorizados y pusheados, abrir el Pull Request de `feature/HU-50-listar-proyectos` hacia `main`, con `Closes #183` en la descripción si no queda nada por cambiar, y pedir la revisión de otro integrante. Verificar que la CI está en verde.
- [ ] 6.2 Agregar al Issue #183 la sección "## Trazabilidad" con los links (a `blob/main`) a la propuesta (`openspec/changes/hu-50-listar-proyectos/proposal.md`), al BDD y a la evidencia. Verificar que los tres archivos existen en la rama.

## Workflow follow-up

- El merge lo hace el equipo con "Create a merge commit" (nunca squash ni rebase), con la aprobación del revisor y la CI en verde.
- Correr `/opsx:archive` solo cuando se cumplan las dos condiciones: el PR está integrado en `main` y todas las tareas de este archivo están tildadas. Si falta alguna, no se archiva. Con las dos cumplidas, correr `/opsx:archive` y proponer `git commit -m "docs(HU-50): archivar el cambio en OpenSpec" -m "Refs #183"`. El cambio pasa a `openspec/changes/archive/AAAA-MM-DD-hu-50-listar-proyectos/` y sus requisitos, a `openspec/specs/proyectos/spec.md`.
- En el mismo commit del archivo, actualizar los links de la fila de HU-50 en la tabla de historias del `README.md`: el link a la propuesta pasa a `openspec/changes/archive/AAAA-MM-DD-hu-50-listar-proyectos/proposal.md`, y el de la especificación, a `openspec/specs/proyectos/spec.md`. Verificar que los dos links abren en `main`.
- Reemplazar en la Trazabilidad del Issue #183 el link a la propuesta por los links definitivos (`openspec/specs/proyectos/spec.md` y la carpeta del archivo) y verificar que abren en `main`.
- Avisar que TEC-10 (#177) se puede cerrar, porque HU-50 quedó archivada.
