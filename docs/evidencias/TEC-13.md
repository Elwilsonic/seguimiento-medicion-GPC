# Evidencia · TEC-13 · Pulir la presentación de cada sección de la interfaz

> Issue del repositorio: #217

Sigue a TEC-12 (#212): que cada sección se vea menos básica y se entienda de un vistazo, sin cambiar lo que hace la aplicación. Se especifica con su Issue y no con OpenSpec (alcance de la decisión 22).

- **Rama:** `feature/TEC-13-presentacion`, desde `main` `fd6d223` (con TEC-12 integrada).
- **Fuera de alcance:** mostrar la lista primero en cada sección, los mensajes para secciones vacías y ordenar los botones de la columna de acciones del backlog. Quedan para otra historia.

## Criterios de aceptación

| Criterio | Cómo se cumple | Test |
|---|---|---|
| CA-TEC13.1 Ningún texto de la interfaz pide escribir un ID, y «Sprint actual» sin proyecto dice "Elegí un proyecto primero" | Se cambiaron "Creá un proyecto o cargá uno por su ID." y "Creá un sprint o escribí su ID arriba."; `cargarSprints` se llama al abrir la página y, sin proyecto, avisa "Elegí un proyecto primero" | `TestPresentacion_NingunTextoPideUnID`, `TestPresentacion_SprintActualSinProyecto` (en el navegador) |
| CA-TEC13.2 Cada formulario se muestra como una tarjeta con su título | Regla `section form` (fondo, borde y `--radio`) y el título `h3` como cabecera de la tarjeta; se agregó el título "Votar" al formulario de Planning Poker, el único que no tenía | `TestPresentacion_FormulariosComoTarjetas` |
| CA-TEC13.3 Cada sección tiene un encabezado con su título y una línea que dice para qué sirve | `<p class="descripcion">` debajo del `<h2>` de las 10 secciones | `TestPresentacion_EncabezadoDeCadaSeccion` |
| CA-TEC13.4 Los estados de proyectos, ítems, sprints y defectos se muestran en las tablas como etiquetas de color, con contraste de al menos 4.5:1 | `celdaEstado` en `app.js`; 5 tonos (neutro, info, ok, error y aviso) como variables de la paleta; un defecto ABIERTO se distingue de un sprint ABIERTO | `TestPresentacion_EstadosComoEtiquetas` (en el navegador), `TestPresentacion_ContrasteDeLasEtiquetas` |
| CA-TEC13.5 El menú tiene un ícono por sección, dibujado con SVG dentro de la página, sin recursos ni librerías externas | Un `<svg>` con `aria-hidden="true"` en cada una de las 9 entradas, del color del texto (`currentColor`) | `TestPresentacion_IconosDelMenu` |
| CA-TEC13.6 El botón de cada formulario queda al pie, alineado a la derecha | Regla `section form > button[type="submit"]` con `align-self: flex-end` | `TestPresentacion_BotonAlPieDerecha` |
| CA-TEC13.7 Los colores salen solo de las variables de la paleta (TEC-12), los tests existentes pasan sin cambios en lo que verifican y el manual de usuario tiene una captura nueva | Los colores nuevos son variables de `:root` (lo sigue verificando `TestEstilo_ColoresSoloEnLasVariablesDeRoot`); ningún test existente se modificó; capturas `tec13-1-proyectos.png` y `tec13-2-backlog.png` en el manual | `TestPresentacion_ManualConCapturaNueva` y la suite completa |

Además, el encabezado con «Proyecto actual» y «Sprint actual» ya no acompaña el desplazamiento: con la lista de proyectos abierta tapaba las tablas. Sigue arriba de cada sección (CA-53.2), y los botones dentro de las tablas son más chicos.

## Tests

**9 tests nuevos** en [`internal/httpapi/tec13_presentacion_test.go`](../../internal/httpapi/tec13_presentacion_test.go), cada uno con `// TEC-13 · CA-TEC13.N · Verifica que …`: CA-TEC13.1, 2; CA-TEC13.2, 1; CA-TEC13.3, 1; CA-TEC13.4, 2; CA-TEC13.5, 1; CA-TEC13.6, 1; CA-TEC13.7, 1. Dos corren en Chrome con los pasos de HU-53.

### RED

```
$ go test -count=1 ./internal/httpapi/ -run 'TestPresentacion_'
--- FAIL: TestPresentacion_NingunTextoPideUnID
    / pide escribir un ID: ["cargá uno por su ID" "escribí su ID"]
--- FAIL: TestPresentacion_SprintActualSinProyecto
    «Sprint actual» muestra "", se esperaba "Elegí un proyecto primero"
--- FAIL: TestPresentacion_FormulariosComoTarjetas
    el formulario form-votar-poker de la sección poker no tiene título (<h3>)
    styles.css no tiene la regla de las tarjetas de formulario (section form)
--- FAIL: TestPresentacion_EncabezadoDeCadaSeccion
    la sección proyectos no empieza con <h2> y una <p class="descripcion"> … (las 10 secciones)
--- FAIL: TestPresentacion_EstadosComoEtiquetas
    el estado del proyecto no es una etiqueta
--- FAIL: TestPresentacion_ContrasteDeLasEtiquetas
    la paleta no tiene --etiqueta-neutro-texto o --etiqueta-neutro-fondo (y los otros 4 tonos)
--- FAIL: TestPresentacion_IconosDelMenu
    la entrada "<a href=\"#proyectos\">Proyectos</a>" no tiene un ícono SVG … (las 9 entradas)
--- FAIL: TestPresentacion_BotonAlPieDerecha
--- FAIL: TestPresentacion_ManualConCapturaNueva
```

Fallan los 9, por aserción.

### GREEN

```
$ gofmt -l .
(sin salida)
$ go vet ./...
(sin salida)
$ go test -count=1 ./...
15 paquetes en ok
```

Los links de 18 evidencias se actualizaron con `-update` (solo números de línea de `app.js`).

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | (propuesto) | `test(TEC-13): agregar tests de la presentación de las secciones (RED)` |
| GREEN | (propuesto) | `feat(TEC-13): pulir la presentación de cada sección (GREEN)` |
