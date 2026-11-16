# Evidencia TDD · HU-53 Navegar la interfaz por secciones y elegir de listas

> Issue del repositorio: #211 · Épica E10 · Dashboard · Vinculada a HU-50 (#183), HU-51 (#193), HU-52 (#197) y TEC-04 (#34)

- **Especificación (OpenSpec):** la spec de la capability `interfaz-web`, [`openspec/specs/interfaz-web/spec.md`](../../openspec/specs/interfaz-web/spec.md), con los 7 requisitos de HU-53. El cambio está archivado en [`openspec/changes/archive/2026-11-16-hu-53-navegacion-por-secciones/`](../../openspec/changes/archive/2026-11-16-hu-53-navegacion-por-secciones/): [propuesta](../../openspec/changes/archive/2026-11-16-hu-53-navegacion-por-secciones/proposal.md), [spec delta](../../openspec/changes/archive/2026-11-16-hu-53-navegacion-por-secciones/specs/interfaz-web/spec.md), [diseño](../../openspec/changes/archive/2026-11-16-hu-53-navegacion-por-secciones/design.md) y [tareas](../../openspec/changes/archive/2026-11-16-hu-53-navegacion-por-secciones/tasks.md). Se archivó con `/opsx:archive` el 16/11/2026, con el PR #214 integrado en `main`.
- **Aprobación:** Peñalbé Hernán (Agile Enabler) aprobó la propuesta el 16/11/2026, sin ajustes, y autorizó chromedp el mismo día, solo para tests (decisión técnica 23).
- **Escenarios BDD:** [`features/HU-53-navegacion-por-secciones.feature`](../../features/HU-53-navegacion-por-secciones.feature): 26 escenarios (uno es un esquema con 2 ejemplos), es decir, 27 ejecuciones y 143 pasos, en un navegador real.
- **Rama:** `feature/HU-53-navegacion-por-secciones`, creada desde `main` `d6b6dfb` (con HU-52 archivada, PR #213).
- **Story Points:** a estimar en el Planning siguiente (HU-53 no estaba en el Planning del Sprint 5).

## Cómo se prueba la interfaz

Hasta HU-52, los tests de la interfaz leían el texto de `index.html` y `app.js`, y el uso en el navegador se verificaba a mano. HU-53 cambia lo que pasa al hacer clic, así que sus tests abren la interfaz en **Chrome sin ventana con chromedp** (decisión técnica 23): cargan los datos por la API, hacen los clics y leen la pantalla. chromedp es solo de tests: no entra al binario (`go list -deps ./cmd/server` no lo muestra).

Los pasos del `.feature` y los tests de Go usan los mismos métodos de [`hu53`](../../internal/httpapi/ui_hu53_steps_test.go), así el mismo recorrido se prueba de las dos formas. Sin navegador, los tests se saltean con un aviso; en la CI (`CI=true`) es un error.

## Trazabilidad

**28 tests nuevos (funciones `Test`) y 27 ejecuciones de escenarios BDD.** Se reparten así:

- 27 tests de Go, cada uno con `// HU-53 · CA-53.N · Verifica que …`: 5 en `ui_navegacion_test.go`, 10 en `ui_listas_test.go`, 11 en `ui_acciones_test.go` y 1 en `ui_helpers_test.go`;
- el runner de godog `TestFeatures_HU53Navegacion` en `ui_features_test.go`, con `// HU-53 · BDD · Verifica que …`, que ejecuta los 27 escenarios.

Por criterio, los 27 tests de Go son: CA-53.1, 5; CA-53.2, 5; CA-53.3, 5; CA-53.4, 5; CA-53.5, 2; CA-53.6, 2; CA-53.7, 3. Cada escenario tiene un test de Go con el mismo recorrido, salvo el esquema "Una dirección sin sección muestra Proyectos", que es un test con dos subtests.

### Criterio → test → código

| Criterio | Escenarios BDD | Tests de Go (archivo) | Código que lo implementa (función y archivo) |
|---|---|---|---|
| CA-53.1 Menú con una sección por vez, en el fragmento de la dirección, sin recargar | "Cada sección tiene su fragmento en la dirección", "Ir a una sección desde el menú", "Abrir un link directo y volver atrás", "Una dirección sin sección muestra Proyectos" (2 ejemplos), "Una sección que necesita un proyecto avisa sin llamar a la API" | [`TestNavegacion_CadaSeccionTieneSuFragmento`](../../internal/httpapi/ui_navegacion_test.go#L19), [`TestNavegacion_IrAUnaSeccionMuestraSoloEsa`](../../internal/httpapi/ui_navegacion_test.go#L29), [`TestNavegacion_LinkDirectoYVolverAtras`](../../internal/httpapi/ui_navegacion_test.go#L42), [`TestNavegacion_SinSeccionMuestraProyectos`](../../internal/httpapi/ui_navegacion_test.go#L55), [`TestNavegacion_SinProyectoAvisaSinLlamarALaAPI`](../../internal/httpapi/ui_navegacion_test.go#L69) (`internal/httpapi/ui_navegacion_test.go`) | El `<nav id="menu">` ([`index.html`](../../internal/web/index.html)); [`mostrarSeccion`](../../internal/web/app.js#L92), [`irA`](../../internal/web/app.js#L109) e [`idActual`](../../internal/web/app.js#L55) ([`app.js`](../../internal/web/app.js)) |
| CA-53.2 Proyecto y sprint actuales desde desplegables en un encabezado visible | "Elegir un proyecto y uno de sus sprints", "Cambiar a un proyecto sin sprints", "Crear un proyecto lo deja elegido", "Sin proyectos", "Falla la lectura de los proyectos" | [`TestProyectoActual_ElegirProyectoYSprint`](../../internal/httpapi/ui_listas_test.go#L10), [`TestProyectoActual_CambiarAUnProyectoSinSprints`](../../internal/httpapi/ui_listas_test.go#L27), [`TestProyectoActual_CrearUnProyectoLoDejaElegido`](../../internal/httpapi/ui_listas_test.go#L42), [`TestProyectoActual_SinProyectos`](../../internal/httpapi/ui_listas_test.go#L54), [`TestProyectoActual_FallaLaLecturaDeLosProyectos`](../../internal/httpapi/ui_listas_test.go#L64) (`internal/httpapi/ui_listas_test.go`) | `<select id="proyecto-actual">` y `<select id="sprint-actual">` ([`index.html`](../../internal/web/index.html)); [`cargarProyectos`](../../internal/web/app.js#L169), [`cargarSprints`](../../internal/web/app.js#L182), [`elegirProyecto`](../../internal/web/app.js#L234), [`cargarProyecto`](../../internal/web/app.js#L279) y [`asegurarOpcion`](../../internal/web/app.js#L149) ([`app.js`](../../internal/web/app.js)) |
| CA-53.3 Los formularios eligen el ítem, el sprint, el integrante, el defecto o el esfuerzo de una lista | "Qué ofrece cada desplegable", "Registrar un esfuerzo eligiendo el ítem", "Registrar un defecto eligiendo el ítem y el sprint", "Sin ítems en el sprint actual", "La API rechaza el defecto en un sprint cerrado" | [`TestListas_QueOfreceCadaDesplegable`](../../internal/httpapi/ui_listas_test.go#L75), [`TestListas_RegistrarEsfuerzoEligiendoElItem`](../../internal/httpapi/ui_listas_test.go#L100), [`TestListas_RegistrarDefectoEligiendoItemYSprint`](../../internal/httpapi/ui_listas_test.go#L112), [`TestListas_SinItemsEnElSprintAvisaSinLlamarALaAPI`](../../internal/httpapi/ui_listas_test.go#L124), [`TestListas_LaAPIRechazaElDefectoEnUnSprintCerrado`](../../internal/httpapi/ui_listas_test.go#L137) (`internal/httpapi/ui_listas_test.go`) | Los `<select>` de los formularios ([`index.html`](../../internal/web/index.html)); [`llenarDesplegable`](../../internal/web/app.js#L130) y [`recargarListas`](../../internal/web/app.js#L194) ([`app.js`](../../internal/web/app.js)) |
| CA-53.4 Acciones en las filas de las tablas | "Las acciones de cada tabla", "Registrar un defecto desde una fila del backlog", "Modificar un integrante desde su fila", "Registrar esfuerzo desde un ítem que no es del sprint actual", "Corregir un esfuerzo con horas inválidas" | [`TestAcciones_CadaTablaTieneSusAcciones`](../../internal/httpapi/ui_acciones_test.go#L13), [`TestAcciones_RegistrarDefectoDesdeElBacklog`](../../internal/httpapi/ui_acciones_test.go#L28), [`TestAcciones_ModificarIntegranteDesdeSuFila`](../../internal/httpapi/ui_acciones_test.go#L41), [`TestAcciones_RegistrarEsfuerzoDeUnItemFueraDelSprintAvisa`](../../internal/httpapi/ui_acciones_test.go#L54), [`TestAcciones_CorregirHorasInvalidasMuestraElError`](../../internal/httpapi/ui_acciones_test.go#L68) (`internal/httpapi/ui_acciones_test.go`) | [`filaItem`](../../internal/web/app.js#L611), [`registrarEsfuerzoDe`](../../internal/web/app.js#L898), [`filaEsfuerzo`](../../internal/web/app.js#L934), [`filaDefecto`](../../internal/web/app.js#L1423) y [`elegirIntegrante`](../../internal/web/app.js#L442) ([`app.js`](../../internal/web/app.js)) |
| CA-53.5 Mensajes junto a la acción, con el nombre o el error de la API sin cambiar | "Mensaje de éxito con el nombre", "El mensaje nuevo reemplaza al anterior, y el error se muestra sin cambiar" | [`TestMensajes_ExitoConElNombre`](../../internal/httpapi/ui_acciones_test.go#L80), [`TestMensajes_ElNuevoReemplazaAlErrorAnterior`](../../internal/httpapi/ui_acciones_test.go#L91) (`internal/httpapi/ui_acciones_test.go`) | [`mostrar`](../../internal/web/app.js#L36) y [`accion`](../../internal/web/app.js#L44), con los mensajes de crear ítem, registrar esfuerzo y registrar defecto ([`app.js`](../../internal/web/app.js)) |
| CA-53.6 Uso en pantallas de 360 px sin desplazamiento horizontal | "Pantalla de escritorio", "Pantalla de celular" | [`TestPantalla_EscritorioConElMenuAlCostado`](../../internal/httpapi/ui_acciones_test.go#L105), [`TestPantalla_CelularSinDesplazamientoHorizontal`](../../internal/httpapi/ui_acciones_test.go#L115) (`internal/httpapi/ui_acciones_test.go`) | `.app`, `#menu`, `.tabla table` y `@media (max-width: 700px)` ([`styles.css`](../../internal/web/styles.css)) |
| CA-53.7 Sin cambios en la API y sin dependencias en la interfaz | "La interfaz no pide recursos a otro dominio", "Una ruta de la API que no existe sigue respondiendo 404" | [`TestSinCambios_LaInterfazNoPideRecursosAOtroDominio`](../../internal/httpapi/ui_acciones_test.go#L128), [`TestSinCambios_UnaRutaDeLaAPIInexistenteDa404`](../../internal/httpapi/ui_acciones_test.go#L139) (`internal/httpapi/ui_acciones_test.go`); [`TestInterfazNavegador_AbreLaPaginaDesdeElServidor`](../../internal/httpapi/ui_helpers_test.go#L131) (`internal/httpapi/ui_helpers_test.go`) | Ningún archivo de Go de producción cambia; chromedp solo en `_test.go` (`go.mod`, decisión técnica 23) |

El runner `TestFeatures_HU53Navegacion` (`internal/httpapi/ui_features_test.go`) ejecuta los escenarios de todas las filas.

## Tests existentes adaptados (tarea 3.1)

Commit propio, antes de los tests nuevos (`36830df`). Siguen verificando lo mismo (que el campo acepta el UUID y no lo convierte a número), pero ahora aceptan un `<select>` además de un `<input type="text">`, y pasaban en verde con la interfaz anterior y con la nueva:

| Test | Historia original | Cambio |
|---|---|---|
| `TestInterfaz_CamposDeIDDeEsfuerzosAceptanElUUIDComoTexto`, `TestInterfaz_CamposDeIDDeDefectosAceptanElUUIDComoTexto` y `TestInterfaz_CamposDeIDObligatoriosSiguenSiendoObligatorios` (con los helpers `formInput` y `expectUUIDTextInput`) | HU-52 | Aceptan `<select>` |
| `TestInterfaz_IDDelIntegranteAceptaElUUIDComoTexto` | HU-52 | Acepta `<select>`, que sigue siendo obligatorio |
| `TestInterfaz_ProyectoActualAceptaElUUIDComoTexto` | HU-51 | Acepta `<select>` |
| `TestInterfaz_SprintActualAceptaElUUIDComoTexto` | HU-52 | Acepta `<select>` |
| `TestInterfaz_FormularioDeDefecto`, `TestInterfaz_FormularioDeCambioDeEstadoDeDefecto` y `TestInterfaz_FormularioDeConsultaDeEsfuerzos` | HU-29, HU-30 y HU-27 | Buscan el campo por su `name`, sin exigir `type="text"` |
| `TestInterfaz_RegistrarDefectoMuestraElMensajeDeExito` | HU-29 | Busca el comienzo común del mensaje, que desde HU-53 nombra el ítem |

## Fase RED (`20f1261`)

```
$ go test -count=1 -v ./internal/httpapi/ -run 'TestFeatures_HU53|TestNavegacion_|TestProyectoActual_|TestListas_|TestAcciones_|TestMensajes_|TestPantalla_|TestSinCambios_|TestInterfazNavegador'
27 scenarios (1 passed, 26 failed)
143 steps (50 passed, 26 failed, 67 skipped)
--- FAIL: TestFeatures_HU53Navegacion (140.79s)
… 25 tests de Go en FAIL y 2 en PASS …
FAIL	seguimiento-medicion-gpc/internal/httpapi	281.486s
```

Los fallos son aserciones sobre la interfaz nueva, por ejemplo:

```
no hay un botón "Backlog" en #menu
el desplegable #proyecto-actual no tiene la opción "Sistema de ventas" (tiene [])
no se ve solo la sección proyectos; se ven proyectos | integrantes | backlog | poker | metricas | dashboard | reportes | sprints | defectos
el menú no está al costado del contenido
```

Ningún fallo fue un error del navegador (0 mensajes "navegador:"). Pasaron lo que ya se cumplía antes de HU-53: `TestInterfazNavegador_AbreLaPaginaDesdeElServidor`, `TestSinCambios_UnaRutaDeLaAPIInexistenteDa404` y su escenario. El resto del repositorio seguía en verde (`go test -skip '<tests nuevos>' ./...`, todos `ok`).

## Fase GREEN (`64ebe63`)

Cambian solo `internal/web/index.html`, `app.js` y `styles.css` (`git diff --name-only` sin `_test.go` ni `.feature`), más los números de línea de los links de 17 evidencias (`-update`).

```
$ gofmt -l .
(sin salida)
$ go vet ./...
(sin salida)
$ go test -count=1 -v ./internal/httpapi/ -run TestFeatures_HU53
27 scenarios (27 passed)
143 steps (143 passed)
$ go test -count=1 ./...
ok  	seguimiento-medicion-gpc/cmd/server	0.006s
ok  	seguimiento-medicion-gpc/internal/backlog	0.371s
ok  	seguimiento-medicion-gpc/internal/dashboard	0.033s
ok  	seguimiento-medicion-gpc/internal/defect	0.267s
ok  	seguimiento-medicion-gpc/internal/effort	0.152s
ok  	seguimiento-medicion-gpc/internal/httpapi	35.930s
ok  	seguimiento-medicion-gpc/internal/member	0.359s
ok  	seguimiento-medicion-gpc/internal/metrics	0.249s
ok  	seguimiento-medicion-gpc/internal/poker	0.136s
ok  	seguimiento-medicion-gpc/internal/postgres	0.007s
ok  	seguimiento-medicion-gpc/internal/postgrestest	0.008s
ok  	seguimiento-medicion-gpc/internal/project	0.312s
ok  	seguimiento-medicion-gpc/internal/report	0.237s
ok  	seguimiento-medicion-gpc/internal/sprint	0.346s
ok  	seguimiento-medicion-gpc/internal/trazabilidad	0.601s
```

**Tiempo de los tests en el navegador (R1):** `internal/httpapi` pasa de unos 1 s a unos 34 s: los 27 tests de Go tardan alrededor de 10 s y los 27 escenarios, alrededor de 24 s.

## Refactor (`29876fa`)

`textoEsfuerzo` y `textoDefecto` arman el texto de un esfuerzo y de un defecto en un solo lugar, para el desplegable y para la acción de fila, que antes lo repetían; y se corrigió un comentario que nombraba una función inexistente. `go test -count=1 ./...` sigue en verde.

## Cobertura

```
$ go test -count=1 -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 29.8% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 81.3% of statements
ok  	seguimiento-medicion-gpc/internal/dashboard	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/defect	coverage: 76.4% of statements
ok  	seguimiento-medicion-gpc/internal/effort	coverage: 72.8% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 97.5% of statements
ok  	seguimiento-medicion-gpc/internal/member	coverage: 66.9% of statements
ok  	seguimiento-medicion-gpc/internal/metrics	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/poker	coverage: 73.8% of statements
ok  	seguimiento-medicion-gpc/internal/postgres	coverage: 83.3% of statements
ok  	seguimiento-medicion-gpc/internal/postgrestest	coverage: 36.8% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 74.9% of statements
ok  	seguimiento-medicion-gpc/internal/report	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/sprint	coverage: 75.1% of statements
ok  	seguimiento-medicion-gpc/internal/trazabilidad	coverage: [no statements]
	seguimiento-medicion-gpc/internal/web		coverage: 0.0% of statements
```

Igual a `main` (`d6b6dfb`) en todos los paquetes, medido en una copia temporal con `git worktree` y los mismos comandos (sin base, Linux, go1.27.1): HU-53 no cambia código Go de producción. La cobertura de Go no mide el JavaScript; lo cubren los tests en el navegador. `-race` y PostgreSQL los corre la CI del Pull Request.

## Desviaciones

1. **El orden de las tareas 2.1 y 3.1:** el `.feature` (2.1) se escribió junto con los tests nuevos y entró en el commit del RED, después del commit de tests adaptados (3.1). No cambia lo que se prueba.
2. **chromedp v0.21.1 tiene una API nueva** (acciones genéricas y `Run` que devuelve el valor), distinta de la de los ejemplos habituales; los helpers de `ui_helpers_test.go` y `ui_driver_test.go` la encapsulan.
3. **La tabla de «Ver proyectos» del encabezado** muestra sus títulos aunque esté vacía. Se deja para TEC-12 (#212), que se ocupa del estilo.

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| Propuesta (OpenSpec) | 59cf829 | `docs(HU-53): agregar la propuesta de OpenSpec` |
| Herramienta de tests | 3b5401b | `test(HU-53): agregar chromedp para probar la interfaz en el navegador` |
| Tests adaptados | 36830df | `test(HU-53): adaptar los tests de contenido de la interfaz` |
| RED | 20f1261 | `test(HU-53): agregar tests de la navegación y las listas (RED)` |
| GREEN | 64ebe63 | `feat(HU-53): navegar por secciones y elegir de listas (GREEN)` |
| Refactor | 29876fa | `refactor(HU-53): unificar el texto de esfuerzos y defectos en los desplegables` |
| Evidencia | 5939982 | `docs(HU-53): agregar evidencia TDD y cobertura` |
| Integración (PR #214) | 3eb86fb | `Merge pull request #214 from Elwilsonic/feature/HU-53-navegacion-por-secciones` |

HU-53: Historia → Especificación (OpenSpec) → Criterios de Aceptación → BDD → Tests → Código
