# HU-40 · Exportar un reporte a un archivo PDF

> Issue del repositorio: #58

Aprobada por: Peñalbé Hernán (Agile Enabler), 29/10/2026.

## Historia de usuario

Como Agile Enabler, quiero exportar un reporte a un archivo PDF, para compartirlo o entregarlo.

## Objetivo

Descargar el reporte de un proyecto (HU-38) como un archivo PDF, con el mismo contenido que el reporte: datos del proyecto, métricas M1 a M10, esfuerzos con su total, defectos e integrantes que registraron esfuerzo. El PDF se arma con la librería externa `github.com/go-pdf/fpdf` v0.9.0, aprobada por el equipo (tema 3 del Planning del Sprint 4). Se expone en `GET /api/proyectos/{id}/reporte/pdf` y en el botón "Descargar PDF" de la sección "Reportes". No hay cálculos nuevos: el PDF muestra lo que devuelve `report.Manager.Project`.

## Alcance

- **Incluye:**
  - la función `projectReportLines`, que arma el texto del PDF a partir de un `ProjectReport`, y la función `WriteProjectPDF(w io.Writer, rep ProjectReport) error`, que escribe el PDF, en `internal/report/pdf.go`;
  - el endpoint `GET /api/proyectos/{id}/reporte/pdf`, que descarga el archivo;
  - el botón "Descargar PDF" en la sección "Reportes" de la interfaz web, para el proyecto actual;
  - la dependencia `github.com/go-pdf/fpdf` v0.9.0 en `go.mod`, documentada en el README.
- **No incluye:**
  - el PDF del reporte de un sprint (HU-39): el Issue #58 pide el del proyecto (decisión 2);
  - formato elaborado: tablas, gráficos, logos, colores o fuentes propias (acuerdo previo del Planning: "un PDF sin formato");
  - guardar el archivo en el servidor: el navegador lo descarga (decisión 4);
  - cambios en el reporte de HU-38 (`ProjectReport`, `Project`, `GET /api/proyectos/{id}/reporte` y el botón "Generar reporte"), en HU-39 ni en ningún otro paquete;
  - cálculos o métricas nuevas.

## Verificación de la librería (tema 3 del Planning del Sprint 4)

El acta pide verificar la versión, la licencia y el mantenimiento antes de agregarla a `go.mod`, y que el equipo la apruebe por escrito.

| Punto | Resultado |
|---|---|
| Módulo y versión | `github.com/go-pdf/fpdf` v0.9.0, la última publicada en `proxy.golang.org` (13/09/2023). v1.4.2 y v1.4.3 están retractadas en su propio `go.mod` |
| Licencia | MIT |
| Requisitos | Go puro, sin cgo: compila en las máquinas del equipo sin compilador de C |
| Dependencias | En `go.mod` se agrega solo `github.com/go-pdf/fpdf`. El paquete raíz que se usa solo importa la librería estándar; las dependencias de sus subpaquetes (`barcode`, `gofpdi`, `pdf417`, `x/image`, `pkg/errors`) no se compilan ni se agregan a `go.mod` |
| Prueba previa | En un módulo aparte: genera un PDF válido, escribe los textos con tildes, «», ·, ñ y paréntesis, y reemplaza un carácter fuera de Windows-1252 por "." sin error |
| Mantenimiento | La última versión es de 2023. No se pudo confirmar desde este entorno si el repositorio de GitHub está archivado. Es un límite conocido: el uso es acotado (texto en una página A4) y la API que se usa es estable desde gofpdf |
| Aprobación | Aprobada por el equipo y registrada por Peñalbé Hernán (Agile Enabler) el 29/10/2026 |

## Contrato

### Entradas

| Operación | Entrada | Descripción |
|---|---|---|
| `WriteProjectPDF(w io.Writer, rep ProjectReport) error` | El destino del archivo y el reporte de un proyecto (HU-38) | Escribe el PDF del reporte en `w` |
| `GET /api/proyectos/{id}/reporte/pdf` | ID del proyecto en la ruta | Arma el reporte con `report.Manager.Project` y devuelve el PDF |

`projectReportLines(rep ProjectReport) []pdfLine` es interna del paquete: arma las líneas de texto del PDF. `pdfLine` tiene el texto (`Text string`) y si es un título de sección (`Heading bool`).

### Salidas esperadas

- `WriteProjectPDF` escribe en `w` un PDF A4 vertical, sin comprimir, con una línea de texto por cada `pdfLine`. Los títulos van en negrita. Un texto más largo que el ancho de la página sigue en la línea siguiente, y el PDF suma páginas si hace falta. Devuelve `nil` si escribió el archivo completo, o el error de escritura.
- El endpoint responde `200 OK` con el archivo:
  - `Content-Type: application/pdf`;
  - `Content-Disposition: attachment; filename="reporte-proyecto-{id}.pdf"`;
  - el cuerpo empieza con `%PDF-` y termina con `%%EOF`.
- `WriteProjectPDF` no modifica el reporte que recibe.

### Contenido del PDF

Las líneas, en este orden, con el texto de la interfaz de HU-38 (`app.js`) donde ya existe. Ejemplo con el proyecto del ejemplo de TEC-03 de la SDD de HU-38 (proyecto 1, sprint 1 cerrado, ítems A a D, esfuerzos de Ana, Luis y Ana del 06/10, y un defecto resuelto):

| # | Título | Texto |
|---|---|---|
| 1 | Sí | `Reporte del proyecto «Software Metrics»` |
| 2 | Sí | `Proyecto` |
| 3 | | `Nombre: Software Metrics` |
| 4 | | `Descripción: Estimación y seguimiento de proyectos` |
| 5 | | `Estado: EN_CURSO` |
| 6 | | `Inicio: 2026-10-05` |
| 7 | | `Fin: sin registrar` |
| 8 | Sí | `Métricas` |
| 9 | | `Story points: 10 planificados, 8 completados, 1 ítem(s) sin estimar.` |
| 10 | | `Velocidad: 8.0 SP por sprint (1 sprint(s) cerrado(s)).` |
| 11 | | `Historias completadas: 50.0 % (2 de 4).` |
| 12 | | `Horas: 20.0 estimadas, 20.0 reales, 1 ítem(s) sin horas estimadas.` |
| 13 | | `Desviación: 0.0 h (0.0 %).` |
| 14 | | `Defectos: 1 detectados, 1 resuelto(s).` |
| 15 | Sí | `Esfuerzos` |
| 16 | | `#1 · 2026-10-06 · Ana · Desarrollo · ítem 1 · sprint 1 · 12 h` |
| 17 | | `#2 · 2026-10-06 · Luis · Desarrollo · ítem 2 · sprint 1 · 5 h` |
| 18 | | `#3 · 2026-10-06 · Ana · Desarrollo · ítem 3 · sprint 1 · 3 h` |
| 19 | | `Total: 20.0 horas` |
| 20 | Sí | `Defectos` |
| 21 | | `#1 · ALTA · RESUELTO · ítem 1 · detectado en sprint 1 · resuelto en sprint 1 · Defecto` |
| 22 | | `Cantidad: 1 defecto(s)` |
| 23 | Sí | `Integrantes` |
| 24 | | `Integrantes que registraron esfuerzo: Ana, Luis` |

Sin datos o sin registrar:

| Parte | Texto |
|---|---|
| Fecha de inicio o de fin sin registrar | `Inicio: sin registrar` / `Fin: sin registrar` |
| Velocidad no disponible | `Velocidad: sin datos (0 sprint(s) cerrado(s)).` |
| Porcentaje no disponible | `Historias completadas: sin datos.` |
| Desviación no disponible | `Desviación: sin datos.` |
| Sin esfuerzos | `Sin esfuerzos registrados.` y `Total: 0.0 horas` |
| Sin defectos | `Sin defectos registrados.` y `Cantidad: 0 defecto(s)` |
| Defecto sin sprint de resolución | `… · detectado en sprint 1 · sin sprint de resolución · …` |
| Sin integrantes | `Integrantes que registraron esfuerzo: ninguno` |
| Descripción del proyecto vacía | `Descripción: ` (la línea se mantiene) |

### HTTP

`GET /api/proyectos/{id}/reporte/pdf`:

| Caso | Código | Cuerpo |
|---|---|---|
| Proyecto existente, con o sin datos | 200 | El PDF |
| `{id}` que no es un número entero | 400 | `{"error": "el id debe ser un número entero"}` (`errInvalidID`) |
| Proyecto inexistente (incluidos 0 y negativos) | 404 | `{"error": ...}` con el mensaje de `project.ErrProjectNotFound` |
| Falla de un repositorio o al generar el PDF | 500 | `{"error": "error interno del servidor"}` (`errInternal`), sin detalle |

El PDF se genera completo en memoria antes de escribir la respuesta: si falla, se responde 500 en JSON y nunca un PDF a medias con código 200.

### Interfaz web

En la sección "Reportes", junto a "Generar reporte", el botón "Descargar PDF" (`id="descargar-reporte-pdf"`, `type="button"`). Usa el proyecto actual (`proyectoActual()`), pide `GET /api/proyectos/{id}/reporte/pdf` y, si la respuesta es 200, descarga el archivo `reporte-proyecto-{id}.pdf` y muestra "PDF del proyecto N descargado." en `msg-reporte`. Si la API responde con error, muestra su mensaje en `msg-reporte`. Sin proyecto actual, el mismo mensaje de error que las demás secciones. "Generar reporte" y sus bloques no cambian.

## Reglas de negocio

1. **RN-40.1 Mismo contenido que el reporte.** El PDF muestra exactamente las partes del `ProjectReport` de HU-38: datos del proyecto (incluida la fecha de fin de HU-41), métricas, esfuerzos con su total, defectos con su cantidad e integrantes, en ese orden. No agrega ni quita datos.
2. **RN-40.2 Sin cálculos nuevos.** Los valores salen del `ProjectReport`. Lo único que hace HU-40 es darles formato de texto.
3. **RN-40.3 Formato de los números.** Velocidad, porcentaje de completadas, horas estimadas y reales, desviación (horas y porcentaje) y total de horas: redondeados a un decimal como la API (TEC-03, `math.Round(x*10)/10`, sin `-0.0`) y escritos con un decimal (`8.0`). Story points, cantidades e IDs: enteros. Horas de cada esfuerzo: tal como están guardadas, sin ceros de más (`12`, `1.5`), como la tabla de la interfaz.
4. **RN-40.4 No disponible.** Una métrica no disponible (velocidad, porcentaje o desviación) se escribe "sin datos", nunca 0 (TEC-03).
5. **RN-40.5 Orden.** Esfuerzos y defectos en el orden del `ProjectReport` (por sprint y por ID, y por ID); integrantes en orden de primera aparición (RN-38.6).
6. **RN-40.6 Caracteres.** El texto se escribe en Windows-1252, la codificación de las fuentes estándar del PDF: tildes, ñ, «», · y ° se ven bien. Un carácter que no existe en Windows-1252 se reemplaza por "." y el PDF se genera igual.
7. **RN-40.7 Error al escribir.** Si el destino no acepta la escritura, `WriteProjectPDF` devuelve el error y no se informa éxito. En la API, el PDF se arma en memoria y un error responde 500 sin detalle.
8. **RN-40.8 Proyecto inexistente.** Lo valida `report.Manager.Project` (`project.ErrProjectNotFound`, 404), antes de armar el PDF.
9. **RN-40.9 Solo lectura.** Exportar no modifica proyectos, ítems, sprints, esfuerzos ni defectos, ni el reporte recibido.
10. **RN-40.10 Datos actuales.** El PDF se arma con el reporte del momento del pedido, sin caché ni archivos guardados en el servidor.

## Decisión sobre entidades referenciadas

- **Proyecto:** sí se valida que exista, pero no en HU-40: el endpoint llama a `report.Manager.Project`, que lo valida primero (RN-38.3) y devuelve `project.ErrProjectNotFound`. Validarlo otra vez sería una lectura de más.
- **Sprints, ítems, esfuerzos, defectos e integrantes:** no se validan; vienen del reporte de HU-38, que los leyó con sus propias operaciones.

## Restricciones

- **Organización del código.**

  | Archivo | Cambio |
  |---|---|
  | `internal/report/pdf.go` | Nuevo: `pdfLine`, `projectReportLines` y `WriteProjectPDF`. Es el único archivo que importa `fpdf` |
  | `internal/report/doc.go` | Solo se agrega HU-40 a la descripción del paquete |
  | `internal/httpapi/report_pdf.go` | Nuevo: el handler `projectReportPDF` |
  | `internal/httpapi/server.go` | Solo la ruta `GET /api/proyectos/{id}/reporte/pdf` |
  | `internal/web/index.html` y `app.js` | Solo el botón y su manejador, después del bloque de HU-38 |
  | `go.mod` y `go.sum` | `github.com/go-pdf/fpdf v0.9.0` |
  | `README.md` | El endpoint, el botón y la librería en las dependencias |

  `project.go`, `sprint.go`, `manager.go` y `reports.go` no cambian.
- **Sin compresión.** El PDF se escribe sin comprimir (`SetCompression(false)`), para que los tests encuentren cada línea en los bytes del archivo. Para un reporte de texto, el archivo queda en pocos KB.
- **Fuente.** Helvetica, una de las fuentes estándar del PDF: no se agregan archivos de fuentes al repositorio.
- **Sin estado compartido.** Cada llamada crea su propio documento: no hay mutex ni test de concurrencia.
- **Tests.**
  - Unitarios en `internal/report/pdf_test.go`: las 24 líneas del ejemplo de TEC-03; un proyecto sin datos; el redondeo a un decimal y el `-0.0`; un defecto sin sprint de resolución; las horas de un esfuerzo con decimales; que el PDF empieza con `%PDF-`, termina con `%%EOF` y contiene cada línea; tildes y un carácter fuera de Windows-1252; un texto largo; el error de un destino que falla; y que no modifica el reporte.
  - godog en `internal/report/pdf_features_test.go`, con el runner `TestFeatures_HU40ExportarReportePDF` sobre `features/HU-40-exportar-reporte-pdf.feature`, con el CA en el nombre de cada escenario.
  - API en `internal/httpapi/report_pdf_test.go`: 200 con los encabezados y un PDF con el contenido, 400, 404 y 500 en JSON. Se agrega una fila a `TestRutasConID_IDNoNumericoDevuelve400`.
  - Interfaz en `internal/httpapi/web_test.go`: el botón y la ruta en `app.js`.
  - Los tests existentes, incluidos los de HU-38 y HU-39, tienen que seguir pasando sin cambios.
- **RED.** Firmas vacías de `pdfLine`, `projectReportLines` y `WriteProjectPDF` que devuelven un error, para que el RED compile y falle por aserciones. La librería se agrega en el GREEN, con el código que la usa. `go vet ./...` antes del commit RED.
- **Librerías.** `github.com/go-pdf/fpdf` v0.9.0 solo en `internal/report/pdf.go`. El resto, librería estándar; godog en los tests.

### Límites conocidos

- **Mantenimiento de la librería.** La última versión es de 2023 (ver "Verificación de la librería"). Si deja de compilar con una versión futura de Go, hay que cambiarla; el resto del código no depende de ella, porque está aislada en `pdf.go`.
- **Caracteres fuera de Windows-1252.** Un nombre con caracteres de otros alfabetos o emojis se ve con "." en el PDF (RN-40.6). Para mostrarlos haría falta agregar al repositorio una fuente TrueType.
- **El PDF no se compara byte a byte.** La librería guarda la fecha de creación en los metadatos, así que dos exportaciones del mismo reporte no son idénticas. Los tests verifican el contenido, no los bytes.
- **Sin compresión.** El archivo es más grande que comprimido; para un reporte de texto son pocos KB.
- **Redondeo repetido.** El redondeo a un decimal ya existe en `internal/httpapi` (`roundOneDecimal`, sin exportar). `pdf.go` tiene el suyo, con la misma regla, para no exportar funciones de `httpapi` ni cambiar ese paquete. Un test verifica el mismo resultado en los casos de TEC-03.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Proyecto del ejemplo de TEC-03 | PDF con las 24 líneas de la tabla de contenido |
| Proyecto recién creado, sin datos | PDF con "sin registrar", "sin datos", "Sin esfuerzos registrados.", "Sin defectos registrados." y "ninguno"; no es un error |
| Proyecto con fecha de fin (HU-41) | `Fin: AAAA-MM-DD` |
| Horas 0.1 y 0.2 | `Total: 0.3 horas`; cada esfuerzo con sus horas tal cual (`0.1 h`, `0.2 h`) |
| Desviación de -0.04 h | `0.0`, nunca `-0.0` |
| Defecto ABIERTO sin sprint de resolución | `sin sprint de resolución` |
| Nombre con tildes, ñ o «» | Se ven en el PDF |
| Nombre con un carácter fuera de Windows-1252 | Se reemplaza por "."; el PDF se genera |
| Paréntesis o barra invertida en un texto | Se escriben bien (la librería los escapa) |
| Descripción más larga que el ancho de la página | Sigue en la línea siguiente; muchos datos suman páginas |
| Destino que falla al escribir | `WriteProjectPDF` devuelve el error |
| Exportar dos veces sin cambios | El mismo contenido |
| HTTP: proyecto inexistente, 0 o negativo | 404 en JSON, sin PDF |
| HTTP: `{id}` no numérico | 400 en JSON |
| HTTP: falla un repositorio | 500 en JSON, sin detalle |
| Interfaz sin proyecto actual | Mensaje de error en `msg-reporte`, sin llamar a la API |

## Condiciones de error

| Condición | Error | HTTP |
|---|---|---|
| `{id}` que no es un número entero | `errInvalidID` | 400 |
| No existe el proyecto (incluidos 0 y negativos) | `project.ErrProjectNotFound` | 404 |
| Falla un repositorio al armar el reporte | El error del repositorio | 500, sin detalle |
| El destino del PDF no acepta la escritura | El error de escritura | 500, sin detalle |
| Proyecto sin datos o métrica no disponible | No es un error (RN-40.4) | 200 |

## Criterios de aceptación

- **CA-40.1** Se exporta el reporte de un proyecto a un archivo PDF válido, que se descarga desde la API (`GET /api/proyectos/{id}/reporte/pdf`) y desde el botón "Descargar PDF" de la sección "Reportes".
- **CA-40.2** Si no se puede escribir el archivo, se devuelve un error y no se informa éxito; en la API, responde 500 sin detalle y sin un PDF a medias.
- **CA-40.3** El contenido del PDF coincide con el reporte generado (HU-38): datos del proyecto, métricas con los mismos valores redondeados de la API, esfuerzos con su total, defectos con su cantidad e integrantes.
- **CA-40.4** Se usa la librería externa `github.com/go-pdf/fpdf` v0.9.0, aprobada por el equipo y documentada con el motivo de su elección en esta SDD y en el README.
- **CA-40.5** Un proyecto sin datos genera un PDF con las secciones vacías y las métricas que no se pueden calcular como "sin datos", no un error.
- **CA-40.6** Exportar el reporte de un proyecto inexistente devuelve error (404 en la API) y no genera el PDF.
- **CA-40.7** Las tildes, la ñ y los signos del reporte se ven en el PDF, y un carácter que no se puede mostrar no impide generarlo.
- **CA-40.8** Exportar el reporte no modifica nada.

## Comparación con el acta de Planning

**Coincide con el acta del Sprint 4.** El acta asigna HU-40 a Peñalbé Hernán con **5 SP**, con la nota "Genera un PDF con el reporte de HU-38; requiere una librería externa y su integración", y dependencias HU-38 y la librería aprobada. HU-38 está en `main` (PR #148).

Precisiones registradas:

1. **Librería (tema 3).** El acta pide verificar versión, licencia y mantenimiento, y registrar la aprobación por escrito antes de empezar. La verificación está en esta SDD; el mantenimiento no se pudo confirmar del todo (límite conocido).
2. **Recorte (acuerdo previo).** El acta prevé "un PDF sin formato en HU-40" si el Sprint se atrasa. Como el desarrollo cierra el 30/10, esta SDD ya es esa versión: texto con títulos en negrita, sin tablas ni gráficos, y solo el reporte del proyecto.

## Comparación con los criterios del Issue #58

Los criterios del Issue #58 están en borrador. Con la aprobación de esta SDD se cierran con CA-40.1 a CA-40.8, y se actualiza el Issue con autorización.

| CA original del Issue #58 | En esta SDD | Ajuste |
|---|---|---|
| CA-40.1 Se guarda el reporte de un proyecto en un archivo en formato PDF | CA-40.1 | El archivo se descarga desde la API y la interfaz, no se guarda en el servidor (decisión 4) |
| CA-40.2 Si no se puede escribir el archivo, se devuelve error | CA-40.2 | Se precisa la respuesta de la API |
| CA-40.3 El contenido del archivo coincide con el reporte generado | CA-40.3 | Se precisa el formato de los números |
| CA-40.4 Librería externa aprobada y documentada con el motivo de su elección | CA-40.4 | Se nombra la librería y la versión |
| — | CA-40.5 a CA-40.8 (nuevos) | Proyecto sin datos, inexistente, caracteres y solo lectura, como en HU-38 |

## Decisiones tomadas en la redacción

1. **`github.com/go-pdf/fpdf` v0.9.0.** La biblioteca estándar no genera PDF. fpdf está escrita en Go puro (sin cgo), tiene licencia MIT y escribe texto en una página con pocas líneas de código. *Alternativas descartadas:* `jung-kurt/gofpdf`, de la que fpdf es la continuación y que está archivada; librerías con licencia comercial o AGPL (como UniPDF); herramientas que convierten HTML a PDF con un programa externo o cgo; y escribir el formato PDF a mano, que es más código propio sin probar.
2. **Solo el reporte del proyecto.** Es lo que pide el Issue #58 (CA-40.1) y el acta ("el reporte de HU-38"). El del sprint (HU-39) se puede sumar después con la misma función de escritura.
3. **El contenido separado del dibujo.** `projectReportLines` arma el texto y `WriteProjectPDF` lo escribe con fpdf. Así el contenido (CA-40.3) se prueba comparando listas de texto, sin leer el PDF, y la librería queda en una sola función.
4. **Descarga por API y botón, sin archivos en el servidor.** El usuario usa la interfaz web: el archivo tiene que llegar a su máquina. Guardarlo en una carpeta del servidor lo dejaría donde no lo ve. CA-40.2 se cumple con un `io.Writer` que falla en los tests y con el PDF armado en memoria antes de responder. *Alternativa descartada:* que la API guarde el archivo en una carpeta y devuelva la ruta.
5. **PDF sin formato.** Texto con títulos en negrita, como prevé el acta si el Sprint se atrasa, y el desarrollo cierra el 30/10. *Alternativa descartada:* tablas con columnas, que requieren calcular anchos y cortes de página.
6. **Sin compresión.** Para que los tests encuentren el texto en el archivo sin descomprimirlo ni agregar otra librería.
7. **Windows-1252 con las fuentes estándar.** Cubre el español y los signos que ya usa la interfaz, sin agregar archivos de fuentes. *Alternativa descartada:* una fuente TrueType en el repositorio, para mostrar cualquier carácter.
8. **Mismos textos que la interfaz de HU-38.** El PDF usa las frases de `app.js` ("sin datos", "ítem(s) sin estimar", "Sin esfuerzos registrados.") para que el reporte se lea igual en pantalla y en el archivo.
9. **Archivos nuevos.** `pdf.go` y `report_pdf.go`, sin tocar los archivos de HU-38 y HU-39.
10. **Story Points.** 5 SP, según el Planning del Sprint 4; no se divide.

HU-40: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
