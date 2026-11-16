# Diseño: HU-53 · Navegar la interfaz por secciones y elegir de listas

## Context

La interfaz (TEC-04) son tres archivos servidos desde el binario con `embed` (`internal/web/web.go`): `index.html` (460 líneas, una página con una `<section>` por área), `app.js` (1179 líneas, sin librerías) y `styles.css` (278 líneas). El proyecto y el sprint actuales son dos `<input type="text">` (`#proyecto-actual` y `#sprint-actual`) que lee `idActual()`; casi todas las acciones de `app.js` los usan a través de `proyectoActual()` y `sprintActual()`. Los formularios de esfuerzo y de defectos piden UUID escritos (CA-52.6).

Los tests de la interfaz (`web_test.go`, 72 tests, y los `*_web_test.go` de HU-50 a HU-52, 14 tests) leen el texto de `index.html` y `app.js`: no ejecutan JavaScript ni hacen clics. Por eso, hasta hoy, el uso desde el navegador se verificaba a mano en la evidencia (CA-TEC04.1 y CA-TEC04.2).

Ver `proposal.md` para el porqué. Los requisitos están en `specs/interfaz-web/spec.md`.

## Goals / Non-Goals

**Goals:**
- Cambiar solo `internal/web`; ningún archivo de Go de producción.
- Que cada escenario de la especificación se pruebe con clics en un navegador real, dentro de `go test ./...` y de la CI.
- Reutilizar `proyectoActual()`, `sprintActual()`, `accion()` y `mostrar()` para no reescribir las 40 acciones de `app.js`.

**Non-Goals:**
- Colores y estilo visual (TEC-12).
- Un framework de JavaScript, un paso de compilación o un enrutador externo.
- Cambiar los textos de las acciones que ya existen, salvo los que nombran un campo que desaparece.

## Decisions

### D1. Navegación con el fragmento de la dirección (`#seccion`)
Un `<nav>` con un link por sección (`<a href="#backlog">`). Un manejador de `hashchange` muestra la `<section>` cuyo `id` coincide y oculta las demás con el atributo `hidden`; si el fragmento no es una sección, se usa `#proyectos`. La entrada visible lleva `aria-current="page"`.

*Alternativas descartadas:* varias páginas HTML (recargan la página, repiten el encabezado y pierden el estado de `app.js`); pestañas sin fragmento (no funcionan el botón "Atrás" ni los links directos, que pide el escenario "Abrir un link directo").

### D2. Las nueve secciones
Se reordenan las `<section>` que ya existen, sin cambiar sus formularios salvo los campos de ID:

| Sección (`id`) | Contenido actual que recibe |
|---|---|
| Proyectos (`proyectos`) | Crear y modificar proyecto, fechas, «Ver proyectos» (HU-50) |
| Integrantes (`integrantes`) | Igual que hoy |
| Backlog (`backlog`) | Igual que hoy |
| Sprints (`sprints`) | Crear, sprint actual, completados, cerrar, lista de sprints |
| Planning Poker (`poker`) | Igual que hoy |
| Esfuerzo (`esfuerzo`) | Registrar, corregir y consultar esfuerzos (hoy dentro de Sprints) |
| Defectos (`defectos`) | Igual que hoy |
| Métricas y dashboard (`metricas`) | Las secciones Métricas y Dashboard juntas |
| Reportes (`reportes`) | Igual que hoy |

*Alternativa descartada:* una sección por historia; serían más de 20 entradas en el menú.

### D3. Los desplegables conservan los `id` de los campos que reemplazan
`#proyecto-actual` y `#sprint-actual` pasan a `<select>` con el mismo `id`. `idActual()` lee `.value`, que en un `<select>` es el UUID de la opción elegida, así que `proyectoActual()`, `sprintActual()` y sus 40 usos no cambian. Lo mismo con los `name` de los campos de los formularios (`item_id`, `sprint_deteccion_id`, `sprint_resolucion_id`, `defecto_id`, `esfuerzo_id` e `id` del integrante): `form.item_id.value` sigue devolviendo el UUID.

*Alternativa descartada:* `id` nuevos, que obligan a cambiar cada acción y aumentan el riesgo de romper una.

### D4. Un solo lugar que llena los desplegables
`llenarDesplegable(select, lista, texto, vacio)` arma las opciones de cualquier desplegable: el valor es el `id` (UUID) y el texto lo arma `texto(entidad)`; si la lista está vacía, deja una sola opción deshabilitada con el texto `vacio`. `recargarListas()` pide las listas del proyecto actual (backlog, sprints, integrantes y defectos) y, si hay sprint actual, sus esfuerzos, y llena todos los desplegables que dependen de ellas. Se llama al cambiar el proyecto o el sprint actual y después de cada acción que crea o modifica una entidad de esas listas.

*Alternativa descartada:* un caché central con invalidación por entidad: más código para listas que en esta aplicación tienen decenas de elementos.

### D5. Los ítems del sprint actual se filtran en la interfaz
«Ítem» de "Registrar esfuerzo" usa la lista del backlog (`GET /api/proyectos/{id}/backlog`) filtrada por `sprint_id` igual al UUID del sprint actual (RN-26.6). No se agrega una ruta.

*Alternativa descartada:* una ruta nueva `GET /api/sprints/{id}/items`; cambia la API, y CA-53.7 dice que no.

### D6. Las acciones de fila navegan y eligen
Cada acción de fila cambia el fragmento a la sección de destino, elige la entidad en el desplegable del formulario y, si corresponde, carga sus datos (el nombre y el rol del integrante). Si la entidad no está en el desplegable (un ítem que no es del sprint actual), muestra el aviso del escenario y no elige nada.

### D7. chromedp para probar la interfaz en un navegador real
Los escenarios se prueban con `github.com/chromedp/chromedp`, que maneja Chromium o Chrome por el protocolo de DevTools, con una versión fija en `go.mod`. Es solo de tests: se importa únicamente desde archivos `_test.go`, así que no entra en el binario (CA-53.7). Autorizada por el Agile Enabler el 16/11/2026; se registra como decisión técnica 23 en `docs/decisiones-tecnicas.md`, con la revisión de seguridad (licencia MIT, sin telemetría, qué descarga).

Cada test levanta el handler con `httptest.NewServer`, abre la página en un Chromium sin ventana y hace los clics del escenario. Sin navegador disponible, el test se saltea con un aviso que dice cómo instalarlo; en la CI (variable `CI=true`), la falta del navegador es un error, para que los escenarios no dejen de correr sin que nadie lo note.

*Alternativas descartadas:* Playwright (es de Node, no corre dentro de `go test`); Selenium (necesita un servidor aparte); seguir con tests de contenido y probar los clics a mano (no prueba el comportamiento en la CI).

### D8. Dónde van los tests
En `internal/httpapi`, paquete `httpapi_test`, para reutilizar `newHandler()` y los helpers de datos que ya usan los tests de la API: `ui_navegacion_test.go`, `ui_listas_test.go` y `ui_acciones_test.go` para los tests unitarios de la interfaz, y `ui_features_test.go` para el runner de godog de `features/HU-53-navegacion-por-secciones.feature`. Un helper `newBrowser(t)` abre el navegador una vez por test, con un tiempo máximo, y lo cierra al terminar.

*Alternativa descartada:* un paquete aparte (`internal/webtest`); tendría que copiar los helpers de `httpapi_test`.

### D9. Tests existentes que se adaptan
Los tests de contenido que buscan los campos que se reemplazan (por ejemplo, `type="text"` en «Sprint actual», CA-52.6, o el botón «Cargar proyecto») se adaptan en el RED, en un commit de tests aparte antes de los tests nuevos, y se listan en la evidencia con su historia original. Lo que verificaban (que el ID no se convierte a número y que se puede elegir de una lista) pasa a verificarse con los tests de HU-53 en el navegador.

### D10. Archivos y funciones por requisito

| Requisito | Archivos | Funciones o elementos |
|---|---|---|
| Navegación por secciones | `index.html`, `app.js`, `styles.css` | `<nav>`, `mostrarSeccion()` y el manejador de `hashchange` |
| Proyecto y sprint actuales | `index.html`, `app.js` | `<select id="proyecto-actual">`, `<select id="sprint-actual">`, `cargarProyectos()`, `cargarSprints()` |
| Elegir de una lista | `index.html`, `app.js` | `llenarDesplegable()`, `recargarListas()` y los `<select>` de los formularios |
| Acciones de fila | `app.js` | `filaItem()`, `filaEsfuerzo()`, `filaDefecto()`, la fila de integrantes e `irA(seccion, elegir)` |
| Mensajes | `app.js` | `mostrar()` y `accion()`, con los textos de éxito con el nombre |
| Pantallas chicas | `styles.css` | `@media (max-width: 700px)` |
| Sin cambios en la API | — | Ninguna ruta nueva; chromedp solo en `_test.go` |

## Risks / Trade-offs

- **[R1] Tests en el navegador más lentos e inestables.** → Un navegador por test con tiempo máximo, esperas por elementos visibles (no por tiempo fijo) y datos creados por la API antes de abrir la página. Se mide el tiempo total en la evidencia.
- **[R2] Integrantes en Windows sin Chrome.** → El test se saltea con un aviso que dice cómo instalarlo; la CI no los saltea (D7). El README lo explica.
- **[R3] Muchos tests de contenido rotos.** → D3 conserva los `id` y los `name`; los que igual se rompen se adaptan en un commit propio (D9).
- **[R4] Choque con TEC-12.** → TEC-12 empieza después de que HU-53 esté en `main`, y solo toca `styles.css`.
- **[R5] La demo del 16/11 a las 18:30.** → HU-53 no entra a `main` antes de la presentación: el merge dispara el deploy a Azure. **Actualización:** el Agile Enabler decidió integrarla antes de la presentación, para mostrar la interfaz nueva, con la CI en verde y la aprobación de Buttini Cristobal (PR #214, 16/11/2026); si el deploy fallara, se revierte el merge.
- **[R6] Listas desactualizadas** si otro usuario crea entidades en simultáneo. → Se recargan al cambiar de proyecto o de sprint y después de cada acción; la API valida igual.

## Migration Plan

No hay migración de datos ni cambios en la base. Para volver atrás, se revierte el merge del PR: la API no cambia, así que la interfaz anterior vuelve a funcionar tal cual.
