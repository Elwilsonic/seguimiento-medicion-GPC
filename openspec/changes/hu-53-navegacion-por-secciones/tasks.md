# Tareas: HU-53 · Navegar la interfaz por secciones y elegir de listas

Rama `feature/HU-53-navegacion-por-secciones`, desde `main` (`d6b6dfb`). Cada commit y cada push necesitan la autorización explícita del Agile Enabler; cada commit lleva `-m "Refs #211"`.

## 0. Aprobación

- [x] 0.1 Registrar en `proposal.md` la aprobación del Agile Enabler (fecha y ajustes pedidos). Verificar con `openspec validate hu-53-navegacion-por-secciones --strict`. Proponer `git commit -m "docs(HU-53): agregar la propuesta de OpenSpec" -m "Refs #211"`.

## 1. Herramienta de tests en el navegador (D7 y D8)

- [x] 1.1 Agregar `github.com/chromedp/chromedp` a `go.mod` con una versión fija publicada hace más de dos semanas, y registrar la decisión técnica 23 en `docs/decisiones-tecnicas.md` (alternativas descartadas, licencia, qué descarga, sin telemetría, solo tests). Verificar con `go list -deps ./cmd/server | grep chromedp`, que no debe mostrar nada (no entra al binario).
- [x] 1.2 Escribir el helper `newBrowser(t)` en `internal/httpapi/ui_helpers_test.go`: levanta `httptest.NewServer(newHandler())`, abre un Chromium sin ventana con tiempo máximo, lo cierra con `t.Cleanup`, se saltea sin navegador y falla en la CI (`CI=true`). Verificar con un test que abre `/` y lee el título de la página, en verde contra la interfaz actual.
- [ ] 1.3 Agregar a la CI (`.github/workflows/ci.yml`) lo que haga falta para que el navegador esté disponible, y comprobarlo en el PR. No hizo falta cambiar el workflow: `ubuntu-latest` ya trae Chrome y GitHub define `CI=true`; queda por comprobar en el PR. Proponer `git commit -m "test(HU-53): agregar chromedp para probar la interfaz en el navegador" -m "Refs #211"`.

## 2. Escenarios BDD

- [x] 2.1 Escribir `features/HU-53-navegacion-por-secciones.feature`, en español, con los escenarios de `specs/interfaz-web/spec.md` (normal, alternativo, límite y error de cada requisito), y su runner de godog `TestFeatures_HU53Navegacion` en `internal/httpapi/ui_features_test.go`, con `// HU-53 · BDD · Verifica que …`. Los pasos usan `newBrowser(t)`.

## 3. Tests (RED)

- [x] 3.1 Commit de tests aparte, antes de los tests nuevos (D9): adaptar los tests de contenido que buscan lo que se reemplaza (los `type="text"` de «Proyecto actual», «Sprint actual» y los campos de ID de CA-51.6 y CA-52.6, el botón «Cargar proyecto»). Listar cada test adaptado con su historia original. Verificar que `go test ./...` sigue en verde con la interfaz actual. Proponer `git commit -m "test(HU-53): adaptar los tests de contenido de la interfaz" -m "Refs #211"`.
- [x] 3.2 Escribir los tests de navegación (`ui_navegacion_test.go`), de los desplegables (`ui_listas_test.go`) y de las acciones de fila y los mensajes (`ui_acciones_test.go`), cada uno con `// HU-53 · CA-53.N · Verifica que …`. Cubrir los escenarios de 2.1 y CA-53.6 (360 px) y CA-53.7 (ningún pedido a otro dominio).
- [x] 3.3 Correr `go test ./...` y verificar que fallan solo los tests nuevos de HU-53, por aserción (por ejemplo, "no hay un menú de navegación"), con `internal/trazabilidad` en verde. Proponer `git commit -m "test(HU-53): agregar tests de la navegación y las listas (RED)" -m "Refs #211"`.

## 4. Código (GREEN)

- [ ] 4.1 `index.html`: el `<nav>`, el encabezado con los dos desplegables (D3), las nueve secciones (D2) y los `<select>` de los formularios.
- [ ] 4.2 `app.js`: `mostrarSeccion()` y `hashchange` (D1), `cargarProyectos()` y `cargarSprints()`, `llenarDesplegable()` y `recargarListas()` (D4), el filtro de D5, las acciones de fila con `irA()` (D6) y los mensajes con el nombre.
- [ ] 4.3 `styles.css`: el menú al costado y, en `@media (max-width: 700px)`, arriba; las tablas con desplazamiento propio. Sin cambiar los colores (TEC-12).
- [ ] 4.4 Correr `gofmt -l .`, `go vet ./...` y `go test -count=1 ./...`, en memoria y con PostgreSQL: todo en verde. Verificar con `git diff --stat` que el GREEN no toca `_test.go` ni `.feature`. Proponer `git commit -m "feat(HU-53): navegar por secciones y elegir de listas (GREEN)" -m "Refs #211"`.

## 5. Refactor

- [ ] 5.1 Si hay una mejora clara (por ejemplo, quitar funciones de `app.js` que quedaron sin uso), proponerla aparte y correr de nuevo los comandos de 4.4. Si no, anotar "no hubo refactor" en la evidencia. Commit, si lo hay: `refactor(HU-53): <qué se mejoró>`.

## 6. Evidencia y documentación

- [ ] 6.1 Crear `docs/evidencias/HU-53-tdd.md`: RED, GREEN y refactor con sus salidas, la cantidad exacta de tests y de escenarios, la tabla criterio → test → código con links (correr `go test ./internal/trazabilidad/ -run TestEnlaces -args -update`), los tests adaptados de 3.1 y el tiempo de los tests del navegador (R1).
- [ ] 6.2 Capturas de la interfaz nueva a 1280 px y a 360 px en `docs/img/`, y actualizar el manual de usuario (TEC-07) y la documentación técnica (TEC-06) con la navegación y los desplegables.
- [ ] 6.3 Actualizar `docs/reportes/cobertura.md`, la fila de HU-53 del README (y cómo instalar Chrome para los tests, R2) y `docs/ia/registro.md`. Proponer `git commit -m "docs(HU-53): agregar evidencia TDD y cobertura" -m "Refs #211"`.

## 7. Pull Request

- [ ] 7.1 Con los commits autorizados y pusheados, abrir el PR hacia `main` con la revisión de otro integrante, y agregar la sección "## Trazabilidad" al Issue #211 con el link a esta propuesta, el `.feature` y la evidencia. Verificar que la CI corre los tests del navegador (no salteados) y queda en verde. El merge va después de la presentación del 16/11 (R5).

## Workflow follow-up

- Con el PR integrado en `main`: `/opsx:archive` en un commit propio (`docs(HU-53): archivar el cambio en OpenSpec`), y reemplazar en el Issue #211 el link a la propuesta por los definitivos (`openspec/specs/interfaz-web/spec.md` y la carpeta del archivo).
- TEC-12 (#212) empieza después del merge de HU-53 (R4).
