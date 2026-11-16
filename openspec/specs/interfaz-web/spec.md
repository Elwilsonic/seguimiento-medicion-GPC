# interfaz-web Specification

## Purpose
La interfaz web de la aplicación (TEC-04): cómo se navega entre las áreas, cómo se eligen el proyecto y el sprint actuales y cómo se eligen las entidades en los formularios, para que se pueda usar sin escribir identificadores a mano.

## Requirements

### Requirement: Navegación por secciones
La interfaz SHALL tener un menú con una entrada por sección (Proyectos, Integrantes, Backlog, Sprints, Planning Poker, Esfuerzo, Defectos, Métricas y dashboard, y Reportes) y SHALL mostrar una sola sección por vez, sin recargar la página. La sección visible SHALL quedar en el fragmento de la dirección (por ejemplo, `#backlog`) y marcada en el menú. Reemplaza la página única de TEC-04. (HU-53, CA-53.1)

La navegación no recibe el ID de ninguna entidad, así que no hay existencia que validar.

#### Scenario: Los fragmentos de cada sección (caso normal)
- **WHEN** el usuario elige cada entrada del menú
- **THEN** la dirección termina, en orden, en `#proyectos`, `#integrantes`, `#backlog`, `#sprints`, `#poker`, `#esfuerzo`, `#defectos`, `#metricas` y `#reportes`

#### Scenario: Ir a una sección desde el menú (caso normal)
- **WHEN** el usuario está en la sección Proyectos y elige "Backlog" en el menú
- **THEN** se ve solo la sección Backlog, la dirección termina en `#backlog`, la entrada "Backlog" del menú queda marcada y la página no se recargó

#### Scenario: Abrir un link directo y volver atrás (caso alternativo)
- **WHEN** el usuario abre la dirección de la aplicación con `#sprints`, después elige "Defectos" en el menú y usa el botón "Atrás" del navegador
- **THEN** primero se ve la sección Sprints, después Defectos, y al volver atrás se ve otra vez Sprints

#### Scenario: Dirección sin sección o con una sección que no existe (caso límite)
- **WHEN** el usuario abre la dirección de la aplicación sin fragmento, o con `#inexistente`
- **THEN** se ve la sección Proyectos y su entrada del menú queda marcada

#### Scenario: Ir a una sección que necesita un proyecto sin haberlo elegido (caso de error)
- **WHEN** no hay proyecto actual y el usuario va a la sección Backlog y pide actualizar la lista
- **THEN** la sección muestra el mensaje "Elegí un proyecto: creá uno o elegilo en «Proyecto actual»." y no se llama a la API

### Requirement: Proyecto y sprint actuales desde desplegables
«Proyecto actual» y «Sprint actual» SHALL ser desplegables en un encabezado visible en todas las secciones, con los proyectos por nombre y los sprints del proyecto actual por nombre y estado. Elegir un proyecto SHALL cargarlo; crear un proyecto o un sprint SHALL dejarlo elegido. Cada opción vale su UUID. Reemplaza los campos de texto de CA-51.6 y CA-52.6. (HU-53, CA-53.2)

Existencia: la valida la API (404), no la interfaz.

#### Scenario: Elegir un proyecto y uno de sus sprints (caso normal)
- **WHEN** existen los proyectos "Inventario" y "Sistema de ventas", "Sistema de ventas" tiene los sprints "Sprint 1" (CERRADO) y "Sprint 2" (ABIERTO), y el usuario elige "Sistema de ventas" en «Proyecto actual» y después "Sprint 2" en «Sprint actual»
- **THEN** el desplegable de proyectos muestra "Inventario" y "Sistema de ventas", en ese orden
- **AND** se cargan los datos de "Sistema de ventas" y el desplegable de sprints muestra "Sprint 1 (CERRADO)" y "Sprint 2 (ABIERTO)"
- **AND** el encabezado sigue mostrando "Sistema de ventas" y "Sprint 2" al cambiar de sección

#### Scenario: Cambiar de proyecto (caso alternativo)
- **WHEN** el proyecto actual es "Sistema de ventas" con el sprint "Sprint 2" elegido, y el usuario elige "Inventario", que no tiene sprints
- **THEN** el desplegable de sprints queda vacío, con el texto "Sin sprints: creá uno en Sprints", y no hay sprint actual

#### Scenario: Crear un proyecto lo deja elegido (caso alternativo)
- **WHEN** el usuario crea el proyecto "Agenda" desde la sección Proyectos
- **THEN** "Agenda" aparece en el desplegable de proyectos, en su lugar por nombre, y queda como proyecto actual

#### Scenario: Sin proyectos (caso límite)
- **WHEN** no existe ningún proyecto y se abre la aplicación
- **THEN** el desplegable de proyectos está vacío, con el texto "Sin proyectos: creá uno en Proyectos", y no hay proyecto ni sprint actual

#### Scenario: Falla la lectura de los proyectos (caso de error)
- **WHEN** `GET /api/proyectos` responde 500 al abrir la aplicación
- **THEN** el encabezado muestra el error que devolvió la API y el desplegable de proyectos queda vacío

### Requirement: Elegir las entidades de los formularios de una lista
Cada formulario que pide un ítem, un sprint, un integrante, un defecto o un esfuerzo SHALL ofrecerlo en un desplegable, sin escribir el UUID. Las listas SHALL actualizarse al cambiar el proyecto o el sprint actual y después de crear o modificar una entidad. Reemplaza los campos de ID de CA-52.6. (HU-53, CA-53.3)

Existencia y reglas: no se validan en la interfaz; la API las sigue validando (por ejemplo, RN-26.5 y RN-29.6) y se muestra su error.

#### Scenario: Qué ofrece cada desplegable (caso normal)
- **WHEN** hay un proyecto y un sprint actuales con ítems, integrantes, defectos y esfuerzos
- **THEN** «Ítem» de "Registrar esfuerzo" ofrece los ítems asignados al sprint actual, por título (RN-26.6)
- **AND** «Ítem» de "Registrar defecto" ofrece los ítems del proyecto, por título
- **AND** «Sprint de detección» y «Sprint de resolución» ofrecen los sprints del proyecto, por nombre y estado
- **AND** «Integrante» de "Modificar integrante" ofrece los integrantes del proyecto, por nombre y rol
- **AND** «Defecto» de "Cambiar estado de defecto" ofrece los defectos del proyecto, por descripción y estado
- **AND** «Esfuerzo» de "Corregir esfuerzo" ofrece los esfuerzos del sprint actual, por fecha, integrante, actividad y horas
- **AND** "Consultar esfuerzos" y "Consultar defectos" se hacen por el sprint actual, el proyecto actual o un ítem elegido de la lista, sin campo «ID»
- **AND** las listas salen de `GET /api/proyectos/{id}/backlog`, `/sprints`, `/integrantes` y `/defectos`, y de `GET /api/sprints/{id}/esfuerzos`

#### Scenario: Registrar un esfuerzo eligiendo el ítem (caso normal)
- **WHEN** el proyecto actual tiene el ítem "Registrar venta" asignado al sprint actual "Sprint 2" (ABIERTO), y el usuario elige "Registrar venta" en «Ítem» de "Registrar esfuerzo", completa integrante "Ana", fecha, actividad "Desarrollo" y 3 horas, y lo registra
- **THEN** la interfaz envía `POST /api/sprints/{UUID de Sprint 2}/items/{UUID de Registrar venta}/esfuerzos` y muestra el mensaje de éxito
- **AND** el esfuerzo nuevo aparece en el desplegable «Esfuerzo» de "Corregir esfuerzo"

#### Scenario: Registrar un defecto eligiendo el ítem y el sprint (caso alternativo)
- **WHEN** el usuario elige el ítem "Consultar stock" y el sprint de detección "Sprint 2" en "Registrar defecto", completa la descripción y la severidad, y lo registra
- **THEN** la interfaz envía el UUID del ítem en la ruta y el UUID de "Sprint 2" en `sprint_deteccion_id`, y el defecto nuevo aparece en el desplegable «Defecto» de "Cambiar estado de defecto"

#### Scenario: Sin entidades para elegir (caso límite)
- **WHEN** el sprint actual no tiene ítems asignados
- **THEN** «Ítem» de "Registrar esfuerzo" está vacío, con el texto "Sin ítems en el sprint actual", y el botón "Registrar esfuerzo" muestra ese mismo aviso sin llamar a la API

#### Scenario: La API rechaza la operación (caso de error)
- **WHEN** el usuario elige el sprint "Sprint 1" (CERRADO) como sprint de detección y registra el defecto
- **THEN** la API responde 409 y la interfaz muestra su mensaje de error junto al formulario, sin limpiar los datos cargados

### Requirement: Acciones en las filas de las tablas
Las tablas SHALL ofrecer en cada fila las acciones sobre esa entidad, sin pedir su ID: la acción lleva a la sección del formulario con la entidad ya elegida. Las acciones de fila que ya existen (editar, cambiar estado y Story Points del ítem, Planning Poker, elegir sprint y elegir proyecto) SHALL mantenerse. (HU-53, CA-53.4)

La acción usa una entidad que se acaba de mostrar en la tabla; si dejó de existir, la API responde 404 al enviar el formulario y la interfaz muestra el error.

#### Scenario: Las acciones de cada tabla (caso normal)
- **WHEN** hay ítems, esfuerzos, defectos e integrantes en sus tablas
- **THEN** cada fila del backlog tiene "Registrar esfuerzo" y "Registrar defecto"
- **AND** cada fila de esfuerzos tiene "Corregir horas", que deja ese esfuerzo elegido en "Corregir esfuerzo"
- **AND** cada fila de defectos tiene "Cambiar estado", que deja ese defecto elegido en "Cambiar estado de defecto"
- **AND** cada fila de integrantes tiene "Modificar", que deja ese integrante elegido en "Modificar integrante"

#### Scenario: Registrar un defecto desde una fila del backlog (caso normal)
- **WHEN** el usuario está en Backlog y elige "Registrar defecto" en la fila del ítem "Consultar stock"
- **THEN** se ve la sección Defectos, con "Consultar stock" elegido en «Ítem» de "Registrar defecto"

#### Scenario: Modificar un integrante desde su fila (caso alternativo)
- **WHEN** el usuario elige "Modificar" en la fila de la integrante "Ana" (PRODUCT_BUILDER)
- **THEN** "Ana" queda elegida en «Integrante» de "Modificar integrante", con el nombre "Ana" y el rol Product Builder cargados

#### Scenario: Registrar esfuerzo desde un ítem que no es del sprint actual (caso límite)
- **WHEN** el usuario elige "Registrar esfuerzo" en la fila de un ítem que no está asignado al sprint actual
- **THEN** se ve la sección Esfuerzo y se muestra el aviso "El ítem no está en el sprint actual: asignalo o elegí su sprint", sin elegir ningún ítem

#### Scenario: Corregir un esfuerzo con horas inválidas (caso de error)
- **WHEN** el usuario elige "Corregir horas" en la fila de un esfuerzo, escribe 0 horas y envía
- **THEN** la API responde 400 y la interfaz muestra su mensaje de error junto a "Corregir esfuerzo"

### Requirement: Mensajes junto a la acción
Cada sección SHALL mostrar el resultado de cada acción junto al formulario o la tabla que la produjo: un mensaje de éxito con el nombre de la entidad (por ejemplo, "Esfuerzo registrado para «Registrar venta»."), o el texto de error que devolvió la API, sin cambiarlo. Un mensaje SHALL reemplazar al anterior de la misma sección. (HU-53, CA-53.5)

No recibe el ID de ninguna entidad.

#### Scenario: Mensaje de éxito con el nombre (caso normal)
- **WHEN** el usuario crea el ítem "Emitir factura" en Backlog
- **THEN** junto al formulario se ve "Ítem «Emitir factura» creado."

#### Scenario: El mensaje nuevo reemplaza al anterior (caso alternativo)
- **WHEN** después de un error en "Registrar defecto" el usuario corrige los datos y lo registra
- **THEN** se ve solo el mensaje de éxito, no el error anterior

#### Scenario: Error de la API sin cambiar su texto (caso de error)
- **WHEN** la API responde 409 con `{"error":"el sprint está cerrado"}`
- **THEN** la interfaz muestra "el sprint está cerrado" como error

### Requirement: Uso en pantallas chicas
A 360 px de ancho, la interfaz SHALL poder usarse sin desplazamiento horizontal de la página: el menú se acomoda arriba del contenido, los formularios ocupan el ancho disponible y las tablas anchas se desplazan dentro de su propio recuadro. (HU-53, CA-53.6)

No recibe el ID de ninguna entidad.

#### Scenario: Pantalla de escritorio (caso normal)
- **WHEN** la ventana mide 1280 px de ancho
- **THEN** el menú se ve al costado del contenido

#### Scenario: Pantalla de celular (caso límite)
- **WHEN** la ventana mide 360 px de ancho y el usuario va a Backlog con 10 ítems
- **THEN** el menú se ve arriba, el ancho de la página no supera el de la ventana, y la tabla del backlog se desplaza dentro de su recuadro

### Requirement: Sin cambios en la API ni dependencias en la interfaz
HU-53 SHALL usar solo las rutas de la API que ya existen, sin cambiar sus respuestas, y la interfaz SHALL seguir sin librerías ni recursos externos, servida desde el binario (CA-TEC04.4). Las dependencias que se agreguen para probar la interfaz SHALL ser solo de tests. (HU-53, CA-53.7)

No recibe el ID de ninguna entidad.

#### Scenario: La interfaz no carga recursos externos (caso normal)
- **WHEN** se abre la aplicación y se recorren todas las secciones
- **THEN** todos los pedidos del navegador van al mismo servidor, y ninguno a otro dominio

#### Scenario: Una ruta que la interfaz no conoce (caso de error)
- **WHEN** la interfaz pide una sección inexistente de la API, como `GET /api/inexistente`
- **THEN** el servidor responde 404, igual que antes de HU-53
