# language: es
# HU-53 · Navegar la interfaz por secciones y elegir de listas (Issue #211), vinculada a HU-50, HU-51, HU-52 y TEC-04
# Especificación (OpenSpec): openspec/specs/interfaz-web/spec.md
# Los escenarios se ejecutan en un navegador real (Chrome sin ventana) con
# chromedp (decisión técnica 23): cada paso hace los clics y lee la pantalla.
# Los datos se cargan antes por la API. Los nombres de las secciones son los
# del menú; los desplegables se nombran por su etiqueta y su formulario.

Característica: HU-53 Navegar la interfaz por secciones y elegir de listas
  Como integrante del equipo,
  quiero navegar la aplicación por secciones y elegir el proyecto, el sprint, los ítems y los defectos de una lista,
  para usarla sin recorrer una página larga ni copiar identificadores a mano.

  # Navegación por secciones (CA-53.1)

  Escenario: Cada sección tiene su fragmento en la dirección (CA-53.1)
    Dado que abro la aplicación
    Cuando elijo cada entrada del menú
    Entonces la dirección termina, en orden, en "#proyectos, #integrantes, #backlog, #sprints, #poker, #esfuerzo, #defectos, #metricas, #reportes"

  Escenario: Ir a una sección desde el menú (CA-53.1)
    Dado que abro la aplicación en "#proyectos"
    Cuando elijo "Backlog" en el menú
    Entonces se ve solo la sección "backlog"
    Y la dirección termina en "#backlog"
    Y la entrada "Backlog" del menú está marcada
    Y la página no se recargó

  Escenario: Abrir un link directo y volver atrás (CA-53.1)
    Dado que abro la aplicación en "#sprints"
    Entonces se ve solo la sección "sprints"
    Cuando elijo "Defectos" en el menú
    Entonces se ve solo la sección "defectos"
    Cuando vuelvo atrás en el navegador
    Entonces se ve solo la sección "sprints"

  Esquema del escenario: Una dirección sin sección muestra Proyectos (CA-53.1)
    Dado que abro la aplicación en "<fragmento>"
    Entonces se ve solo la sección "proyectos"
    Y la entrada "Proyectos" del menú está marcada

    Ejemplos:
      | fragmento    |
      |              |
      | #inexistente |

  Escenario: Una sección que necesita un proyecto avisa sin llamar a la API (CA-53.1)
    Dado que abro la aplicación en "#backlog"
    Cuando pido actualizar el backlog sin proyecto actual
    Entonces el mensaje de "backlog" es "Elegí un proyecto: creá uno o elegilo en «Proyecto actual»."
    Y no se llamó a la API

  # Proyecto y sprint actuales desde desplegables (CA-53.2)

  Escenario: Elegir un proyecto y uno de sus sprints (CA-53.2)
    Dado que existen los proyectos "Sistema de ventas" e "Inventario"
    Y que "Sistema de ventas" tiene el sprint cerrado "Sprint 1" y el sprint abierto "Sprint 2"
    Y que abro la aplicación
    Entonces «Proyecto actual» ofrece "Inventario, Sistema de ventas"
    Cuando elijo "Sistema de ventas" en «Proyecto actual»
    Entonces «Sprint actual» ofrece "Sprint 1 (CERRADO), Sprint 2 (ABIERTO)"
    Cuando elijo "Sprint 2 (ABIERTO)" en «Sprint actual»
    Y elijo "Backlog" en el menú
    Entonces «Proyecto actual» muestra "Sistema de ventas"
    Y «Sprint actual» muestra "Sprint 2 (ABIERTO)"

  Escenario: Cambiar a un proyecto sin sprints (CA-53.2)
    Dado que existen los proyectos "Sistema de ventas" e "Inventario"
    Y que "Sistema de ventas" tiene el sprint cerrado "Sprint 1" y el sprint abierto "Sprint 2"
    Y que abro la aplicación
    Cuando elijo "Sistema de ventas" en «Proyecto actual»
    Y elijo "Sprint 2 (ABIERTO)" en «Sprint actual»
    Y elijo "Inventario" en «Proyecto actual»
    Entonces «Sprint actual» muestra "Sin sprints: creá uno en Sprints"
    Y no hay sprint actual

  Escenario: Crear un proyecto lo deja elegido (CA-53.2)
    Dado que existen los proyectos "Sistema de ventas" e "Inventario"
    Y que abro la aplicación en "#proyectos"
    Cuando creo el proyecto "Agenda" desde la sección Proyectos
    Entonces «Proyecto actual» ofrece "Agenda, Inventario, Sistema de ventas"
    Y «Proyecto actual» muestra "Agenda"

  Escenario: Sin proyectos (CA-53.2)
    Dado que abro la aplicación
    Entonces «Proyecto actual» muestra "Sin proyectos: creá uno en Proyectos"
    Y no hay proyecto actual

  Escenario: Falla la lectura de los proyectos (CA-53.2)
    Dado que la lista de proyectos responde 500 con el error "no se pudieron leer los proyectos"
    Y que abro la aplicación
    Entonces el mensaje del encabezado es "no se pudieron leer los proyectos"
    Y «Proyecto actual» no ofrece proyectos

  # Elegir las entidades de los formularios de una lista (CA-53.3)

  Escenario: Qué ofrece cada desplegable (CA-53.3)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que el proyecto "Sistema de ventas" tiene el ítem "Consultar stock" sin sprint
    Y que "Sistema de ventas" tiene la integrante "Ana" con rol "PRODUCT_BUILDER"
    Y que "Ana" registró 3 horas en "Registrar venta"
    Y que "Registrar venta" tiene el defecto "El total no suma"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Entonces los desplegables ofrecen:
      | desplegable                               | opciones                                |
      | «Ítem» de Registrar esfuerzo              | Registrar venta                         |
      | «Ítem» de Registrar defecto               | Consultar stock, Registrar venta        |
      | «Sprint de detección» de Registrar defecto | Sprint 2 (ABIERTO)                     |
      | «Sprint de resolución» de Cambiar estado  | Sprint 2 (ABIERTO)                      |
      | «Integrante» de Modificar integrante      | Ana (Product Builder)                   |
      | «Defecto» de Cambiar estado               | El total no suma (ABIERTO)              |
      | «Esfuerzo» de Corregir esfuerzo           | 2026-10-02 · Ana · Desarrollo · 3 h     |
      | «Ítem» de Consultar esfuerzos             | Consultar stock, Registrar venta        |

  Escenario: Registrar un esfuerzo eligiendo el ítem (CA-53.3)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando registro un esfuerzo de "Ana" de 3 horas en "Registrar venta" desde la sección Esfuerzo
    Entonces el mensaje de "esfuerzo" es "Esfuerzo registrado para «Registrar venta»."
    Y «Esfuerzo» de Corregir esfuerzo ofrece "2026-10-02 · Ana · Desarrollo · 3 h"

  Escenario: Registrar un defecto eligiendo el ítem y el sprint (CA-53.3)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando registro el defecto "El total no suma" de "Registrar venta" detectado en "Sprint 2 (ABIERTO)"
    Entonces el mensaje de "defectos" es "Defecto registrado para «Registrar venta»."
    Y «Defecto» de Cambiar estado ofrece "El total no suma (ABIERTO)"

  Escenario: Sin ítems en el sprint actual (CA-53.3)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" sin ítems
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Entonces «Ítem» de Registrar esfuerzo muestra "Sin ítems en el sprint actual"
    Cuando intento registrar un esfuerzo desde la sección Esfuerzo
    Entonces el mensaje de "esfuerzo" es "Sin ítems en el sprint actual"
    Y no se llamó a la API

  Escenario: La API rechaza el defecto en un sprint cerrado (CA-53.3)
    Dado que "Sistema de ventas" tiene el sprint cerrado "Sprint 0"
    Y que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando registro el defecto "El total no suma" de "Registrar venta" detectado en "Sprint 0 (CERRADO)"
    Entonces el mensaje de "defectos" es el error "el sprint está cerrado"
    Y la descripción del defecto sigue siendo "El total no suma"

  # Acciones en las filas de las tablas (CA-53.4)

  Escenario: Las acciones de cada tabla (CA-53.4)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que "Sistema de ventas" tiene la integrante "Ana" con rol "PRODUCT_BUILDER"
    Y que "Ana" registró 3 horas en "Registrar venta"
    Y que "Registrar venta" tiene el defecto "El total no suma"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Entonces cada fila de la tabla de "backlog" tiene "Registrar esfuerzo, Registrar defecto"
    Y cada fila de la tabla de "esfuerzo" tiene "Corregir horas"
    Y cada fila de la tabla de "defectos" tiene "Cambiar estado"
    Y cada fila de la tabla de "integrantes" tiene "Modificar"

  Escenario: Registrar un defecto desde una fila del backlog (CA-53.4)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que el proyecto "Sistema de ventas" tiene el ítem "Consultar stock" sin sprint
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando elijo "Registrar defecto" en la fila "Consultar stock" del backlog
    Entonces se ve solo la sección "defectos"
    Y «Ítem» de Registrar defecto muestra "Consultar stock"

  Escenario: Modificar un integrante desde su fila (CA-53.4)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que "Sistema de ventas" tiene la integrante "Ana" con rol "PRODUCT_BUILDER"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando elijo "Modificar" en la fila "Ana" de los integrantes
    Entonces «Integrante» de Modificar integrante muestra "Ana (Product Builder)"
    Y el formulario de modificar integrante tiene el nombre "Ana" y el rol "PRODUCT_BUILDER"

  Escenario: Registrar esfuerzo desde un ítem que no es del sprint actual (CA-53.4)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que el proyecto "Sistema de ventas" tiene el ítem "Consultar stock" sin sprint
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando elijo "Registrar esfuerzo" en la fila "Consultar stock" del backlog
    Entonces se ve solo la sección "esfuerzo"
    Y el mensaje de "esfuerzo" es "El ítem no está en el sprint actual: asignalo o elegí su sprint"
    Y «Ítem» de Registrar esfuerzo no tiene un ítem elegido

  Escenario: Corregir un esfuerzo con horas inválidas (CA-53.4)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que "Ana" registró 3 horas en "Registrar venta"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando elijo "Corregir horas" en la fila del esfuerzo de "Ana" y corrijo a 0 horas
    Entonces el mensaje de "esfuerzo" es el error "las horas trabajadas deben ser un número positivo"

  # Mensajes junto a la acción (CA-53.5)

  Escenario: Mensaje de éxito con el nombre (CA-53.5)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando creo el ítem "Emitir factura" desde la sección Backlog
    Entonces el mensaje de "backlog" es "Ítem «Emitir factura» creado."

  Escenario: El mensaje nuevo reemplaza al anterior, y el error se muestra sin cambiar (CA-53.5)
    Dado que "Sistema de ventas" tiene el sprint cerrado "Sprint 0"
    Y que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando registro el defecto "El total no suma" de "Registrar venta" detectado en "Sprint 0 (CERRADO)"
    Entonces el mensaje de "defectos" es el error "el sprint está cerrado"
    Cuando registro el defecto "El total no suma" de "Registrar venta" detectado en "Sprint 2 (ABIERTO)"
    Entonces el mensaje de "defectos" es "Defecto registrado para «Registrar venta»."

  # Uso en pantallas chicas (CA-53.6)

  Escenario: Pantalla de escritorio (CA-53.6)
    Dado que la ventana mide 1280 px de ancho
    Y que abro la aplicación
    Entonces el menú se ve al costado del contenido

  Escenario: Pantalla de celular (CA-53.6)
    Dado que el proyecto "Sistema de ventas" tiene 10 ítems
    Y que la ventana mide 360 px de ancho
    Y que abro la aplicación con "Sistema de ventas" elegido en "#backlog"
    Entonces el menú se ve arriba del contenido
    Y la página no se desplaza en horizontal
    Y la tabla del backlog se desplaza dentro de su recuadro

  # Sin cambios en la API ni dependencias en la interfaz (CA-53.7)

  Escenario: La interfaz no pide recursos a otro dominio (CA-53.7)
    Dado que el proyecto "Sistema de ventas" tiene el sprint abierto "Sprint 2" con el ítem "Registrar venta"
    Y que abro la aplicación con "Sistema de ventas" y "Sprint 2" elegidos
    Cuando recorro todas las secciones del menú
    Entonces todos los pedidos del navegador fueron al mismo servidor

  Escenario: Una ruta de la API que no existe sigue respondiendo 404 (CA-53.7)
    Cuando pido GET "/api/inexistente" a la API
    Entonces la API responde 404
