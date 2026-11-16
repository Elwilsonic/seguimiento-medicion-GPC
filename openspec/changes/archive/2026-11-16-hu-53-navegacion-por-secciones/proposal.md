# Propuesta: HU-53 · Navegar la interfaz por secciones y elegir de listas

> Issue del repositorio: #211 · Épica E10 · Dashboard · Vinculada a HU-50 (#183), HU-51 (#193), HU-52 (#197) y TEC-04 (#34)
>
> Aprobada por: Peñalbé Hernán (Agile Enabler), 16/11/2026, sin ajustes. chromedp autorizado el mismo día, solo para tests (D7).

## Historia de usuario

Como integrante del equipo, quiero navegar la aplicación por secciones y elegir el proyecto, el sprint, los ítems y los defectos de una lista, para usarla sin recorrer una página larga ni copiar identificadores a mano.

## Why

La interfaz web es una sola página larga, con ocho áreas una debajo de la otra (Proyectos, Integrantes, Backlog, Planning Poker, Métricas, Dashboard, Reportes, Sprints, Esfuerzo y Defectos), y cada formulario pide el ID de la entidad que usa. Desde HU-51 y HU-52 ese ID es un UUID de 36 caracteres: para registrar un esfuerzo o un defecto hay que copiar y pegar el UUID del ítem, del sprint o del defecto desde otra tabla. En la revisión, el profesor pidió una interfaz más profesional, que no vaya de arriba hacia abajo y que sea fácil de entender para el usuario.

## What Changes

- **Navegación por secciones:** un menú con una entrada por área. Se ve una sección por vez, sin recargar la página, y la sección elegida queda en la dirección (`#backlog`, `#sprints`, …), así funcionan el botón "Atrás" del navegador y abrir un link directo a una sección.
- **Proyecto y sprint actuales desde desplegables:** «Proyecto actual» y «Sprint actual» pasan de campos de texto a desplegables con los nombres, en un encabezado fijo que se ve en todas las secciones. Elegir un proyecto lo carga y llena el desplegable de sprints con los de ese proyecto.
- **Formularios sin UUID escritos a mano:** los campos que piden un ítem, un sprint, un integrante, un defecto o un esfuerzo pasan a desplegables con su título, nombre o descripción, llenados con las listas que la API ya tiene.
- **Acciones en cada fila:** desde una fila del backlog se puede registrar un esfuerzo o un defecto de ese ítem; desde una fila de esfuerzos, corregir sus horas; desde una de defectos, cambiar su estado; desde una de integrantes, modificarlo. La acción abre el formulario con la entidad ya elegida.
- **Mensajes junto a la acción:** cada sección muestra el mensaje de éxito o el error de la API junto al formulario o la tabla que lo produjo.
- **Uso en el celular:** a 360 px de ancho, el menú se acomoda arriba y la página no se desplaza en horizontal.
- **Tests de la interfaz en un navegador real:** los escenarios BDD hacen clics en Chromium con chromedp, librería autorizada por el Agile Enabler el 16/11/2026 solo para tests (decisión D7 de `design.md`).

No hay cambios incompatibles en la API (**BREAKING** no aplica): no se agrega ni se cambia ninguna ruta, y el dominio no cambia. Para el usuario de la interfaz, «Proyecto actual (ID)», «Sprint actual (ID)» y los campos de ID dejan de ser texto libre.

## Capabilities

### New Capabilities

- `interfaz-web`: la interfaz web de la aplicación (TEC-04), con su navegación, la elección del proyecto y del sprint actuales y la forma de elegir las entidades en los formularios. Hasta hoy no tenía especificación en `openspec/specs/`: sus reglas están en el Issue de TEC-04 (#34) y en los criterios de interfaz de cada historia (por ejemplo, CA-50.3, CA-51.6 y CA-52.6).

### Modified Capabilities

Ninguna. Las capabilities de HU-50 a HU-52 (`proyectos`, `backlog`, `sprints`, `integrantes`, `defectos`, `esfuerzos` y `planning-poker`) describen la API y el dominio, que no cambian. Sus requisitos de interfaz (por ejemplo, que «Sprint actual» acepte un UUID como texto) quedan cubiertos por los requisitos nuevos de `interfaz-web`, que nombran lo que reemplazan.

## Vínculos y SDD leídas

- **TEC-04** (Issue #34 y `docs/evidencias/TEC-04-tdd.md`; las tareas técnicas no tienen SDD): la interfaz se sirve desde el mismo binario, sin librerías ni recursos externos (CA-TEC04.4). HU-53 lo mantiene: chromedp es solo de tests y no viaja en el binario.
- **HU-50** (`openspec/specs/proyectos/spec.md`): «Ver proyectos» y «Elegir» siguen existiendo en la sección Proyectos; el desplegable usa la misma lista (`GET /api/proyectos`) y el mismo orden por nombre (CA-50.6).
- **HU-51 y HU-52** (`openspec/specs/*/spec.md`): las entidades se identifican con UUID; los desplegables los usan como valor y muestran el nombre.
- **HU-14, HU-17, HU-27 y HU-31:** las listas de sprints, integrantes, esfuerzos y defectos que llenan los desplegables son las mismas rutas de esas historias.
- **HU-37:** los gráficos del dashboard no cambian.

## Comparación con el acta de Planning

HU-53 no está en el acta del Planning del Sprint 5 (Issue #182): se creó el 16/11/2026 después de la revisión del profesor. No hay alcance con el que comparar. Su sprint y sus Story Points se definen en el Planning siguiente.

## Fuera de alcance

- **Colores, tipografía y estilo visual:** van en TEC-12 (#212), después de HU-53. HU-53 mantiene los colores actuales.
- Cambios en la API, en el dominio o en el esquema de la base.
- Ocultar las columnas de UUID de las tablas o reemplazarlas por nombres (se puede proponer después; no está en el Issue).
- Elegir de una lista el nombre del integrante en «Registrar esfuerzo» y en «Votar» de Planning Poker: hoy es un nombre escrito, no un ID, y no cambia.
- Búsqueda o filtros en los desplegables, paginación, inicio de sesión y traducción a otros idiomas.

## Impact

- **Código:** `internal/web/index.html`, `internal/web/app.js` y `internal/web/styles.css`. No cambia ningún archivo de Go de producción.
- **Tests:** tests nuevos de la interfaz con chromedp en `internal/httpapi` (o un paquete de tests propio, D8), escenarios en `features/HU-53-navegacion-por-secciones.feature`, y adaptación de los tests de contenido de `web_test.go` y de los `*_web_test.go` de HU-50 a HU-52 que buscan los campos de texto que se reemplazan.
- **Dependencias:** `github.com/chromedp/chromedp` (solo tests), con versión fija, registrada como decisión técnica en `docs/decisiones-tecnicas.md`. La CI necesita Chromium o Chrome (ubuntu-latest ya trae Chrome).
- **Documentación:** `docs/evidencias/HU-53-tdd.md`, `docs/reportes/cobertura.md`, el manual de usuario (TEC-07) con las capturas nuevas, la documentación técnica (TEC-06), el README y `docs/ia/registro.md`.
