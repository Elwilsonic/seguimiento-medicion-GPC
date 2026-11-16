# Evidencia · TEC-12 · Renovar los colores y el estilo de la interfaz

> Issue del repositorio: #212

Pedido en la revisión del profesor: una interfaz más profesional. TEC-12 cambia solo la apariencia (colores, tarjetas y estilo de botones, campos, tablas y mensajes), después de HU-53, que cambió la estructura de la página. No cambia lo que hace la aplicación, así que se especifica con su Issue y no con OpenSpec (alcance de la decisión 22).

- **Rama:** `feature/TEC-12-colores`, desde `main` `3733fec` (con HU-53 integrada y archivada).
- **Paleta:** índigo como principal (`--primario` `#4338ca`), verde azulado como acento (`--acento` `#0f766e`) y fondos en gris azulado claro. Está documentada, con el contraste de cada par, en la sección 4.3.2 de [`docs/documentacion-tecnica.md`](../documentacion-tecnica.md).

## Criterios de aceptación

| Criterio | Cómo se cumple | Test |
|---|---|---|
| CA-TEC12.1 Los colores se definen una sola vez, como variables de CSS en `:root`, y el resto del CSS solo usa esas variables | `:root` de [`styles.css`](../../internal/web/styles.css) tiene los 23 colores; los 8 colores sueltos que había (`#fff`, `#eef3fb`, `#f3b8b1` y `#b5dcc1`) pasaron a variables. El HTML y el JavaScript no tienen colores | `TestEstilo_ColoresSoloEnLasVariablesDeRoot` |
| CA-TEC12.2 La paleta (principal, secundario, fondo, texto, éxito, error y advertencia) queda documentada en `docs/documentacion-tecnica.md`, con su contraste | Sección 4.3.2 "Paleta de colores", con cada variable, su valor, su uso y el contraste | `TestEstilo_PaletaDocumentada` |
| CA-TEC12.3 El texto tiene un contraste de al menos 4.5:1 con su fondo (WCAG AA) | 11 pares de texto y fondo, de 6.8:1 a 17.9:1, calculados con la fórmula de WCAG | `TestEstilo_ContrasteAA` |
| CA-TEC12.4 Los botones, los campos, las tablas y los mensajes tienen un estilo uniforme, y el dashboard muestra las métricas en tarjetas | Borde redondeado común (`--radio`), foco visible, encabezados de tabla con fondo y tablas en un recuadro. "Ver resumen" muestra 3 tarjetas y "Ver gráficos", 4 (`tarjetas()` en [`app.js`](../../internal/web/app.js)) | `TestEstilo_ComponentesUniformes`, `TestEstilo_DashboardEnTarjetas` (en el navegador) |
| CA-TEC12.5 Los gráficos del dashboard (HU-37) usan los colores de la paleta | Las barras usan `--serie-a` y `--serie-b`, y la referencia, `--referencia`, con los valores nuevos | `TestEstilo_GraficosConLaPaleta` |
| CA-TEC12.6 Solo cambian archivos de estilo y de presentación: los tests de la interfaz y de la API pasan sin cambios en lo que verifican, y no se agregan librerías externas | Cambian `styles.css`, `app.js` (solo cómo se muestran el resumen y las métricas) e `index.html` (los dos contenedores del dashboard pasan de `<p>` a `<div>`). Ningún test existente se modificó y `go test ./...` pasa entero | `TestEstilo_SinRecursosExternos` y la suite completa |
| CA-TEC12.7 El manual de usuario se actualiza con las capturas nuevas | Capturas `tec12-1-dashboard.png` y `tec12-2-backlog.png` en las secciones 2 y 12 del [manual](../manual-de-usuario.md) | `TestEstilo_ManualConLasCapturasNuevas` |

Además, la tabla de «Ver proyectos» del encabezado ya no se muestra mientras está vacía (pendiente anotado en la evidencia de HU-53).

## Tests

**8 tests nuevos** en [`internal/httpapi/tec12_estilo_test.go`](../../internal/httpapi/tec12_estilo_test.go), cada uno con `// TEC-12 · CA-TEC12.N · Verifica que …`: CA-TEC12.1, 1; CA-TEC12.2, 1; CA-TEC12.3, 1; CA-TEC12.4, 2; CA-TEC12.5, 1; CA-TEC12.6, 1; CA-TEC12.7, 1. `TestEstilo_DashboardEnTarjetas` corre en Chrome con los pasos de HU-53.

### RED

```
$ go test -count=1 ./internal/httpapi/ -run 'TestEstilo_'
--- FAIL: TestEstilo_ColoresSoloEnLasVariablesDeRoot
    styles.css tiene 8 colores fuera de :root: ["#fff" "#fff" "#fff" "#fff" "#eef3fb" "#f3b8b1" "#b5dcc1" "#fff"]
--- FAIL: TestEstilo_PaletaDocumentada
    la documentación técnica no tiene la sección de la paleta de colores
--- FAIL: TestEstilo_ContrasteAA
    la paleta no tiene --sobre-primario o --primario  (y otros 3 pares)
--- FAIL: TestEstilo_ComponentesUniformes
    styles.css no tiene la regla ".tarjeta {" (y la regla de mensajes sin variables, sin --radio)
--- FAIL: TestEstilo_DashboardEnTarjetas
    el resumen no se muestra en 3 tarjetas (estado, ítems y sprint actual)
--- FAIL: TestEstilo_ManualConLasCapturasNuevas
    el manual tiene 0 capturas de TEC-12, se esperaban al menos 2
```

Pasaron `TestEstilo_GraficosConLaPaleta` y `TestEstilo_SinRecursosExternos`, que verifican lo que ya se cumplía: los gráficos ya usaban variables y el CSS no cargaba recursos externos.

### GREEN

```
$ gofmt -l .
(sin salida)
$ go vet ./...
(sin salida)
$ go test -count=1 -v ./internal/httpapi/ -run TestEstilo_
--- PASS: TestEstilo_ColoresSoloEnLasVariablesDeRoot
--- PASS: TestEstilo_PaletaDocumentada
--- PASS: TestEstilo_ContrasteAA
--- PASS: TestEstilo_ComponentesUniformes
--- PASS: TestEstilo_DashboardEnTarjetas
--- PASS: TestEstilo_GraficosConLaPaleta
--- PASS: TestEstilo_SinRecursosExternos
--- PASS: TestEstilo_ManualConLasCapturasNuevas
$ go test -count=1 ./...
15 paquetes en ok
```

Los links de las evidencias de HU-31, HU-37, HU-38, HU-49 y HU-53 se actualizaron con `-update` (solo números de línea de `app.js`).

## Commits

| Fase | Commit | Mensaje |
|---|---|---|
| RED | 3824303 | `test(TEC-12): agregar tests de la paleta y el estilo (RED)` |
| GREEN | acf52f8 | `feat(TEC-12): renovar los colores y el estilo de la interfaz (GREEN)` |
