# language: es
# HU-06 · Listar el backlog de un proyecto (Issue #21)
# Especificación: docs/especificaciones/HU-06-listar-backlog.md

Característica: HU-06 Listar el backlog de un proyecto
  Como Agile Enabler,
  quiero listar los ítems del backlog de un proyecto,
  para ver todo el trabajo pendiente en un solo lugar.

  # Caso normal (happy path) - CA-06.1
  Escenario: Listar el backlog de un proyecto con ítems
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Y creo el ítem "Recuperar contraseña" en el backlog de "Sistema de ventas" con prioridad "MEDIA" y criterio "Mail enviado"
    Cuando listo el backlog del proyecto "Sistema de ventas"
    Entonces obtengo 2 ítems
    Y obtengo los ítems con los IDs 1 y 2
    Y el ítem en la posición 1 tiene el título "Login de usuarios" y la prioridad "ALTA"
    Y el ítem en la posición 2 tiene el título "Recuperar contraseña" y la prioridad "MEDIA"

  # Caso alternativo - CA-06.3, RN-06.3 y RN-06.6
  Escenario: No aparecen los ítems de otros proyectos aunque tengan el mismo título
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un proyecto con nombre "Portal de clientes"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Y creo el ítem "Login de usuarios" en el backlog de "Portal de clientes" con prioridad "ALTA" y criterio "Login OK"
    Y creo el ítem "Cierre de sesión" en el backlog de "Sistema de ventas" con prioridad "MEDIA" y criterio "Sesión cerrada"
    Cuando listo el backlog del proyecto "Sistema de ventas"
    Entonces obtengo 2 ítems
    Y obtengo los ítems con los IDs 1 y 3

  # Caso límite - CA-06.2 y RN-06.2
  Escenario: Listar el backlog de un proyecto sin ítems
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando listo el backlog del proyecto "Sistema de ventas"
    Entonces obtengo 0 ítems
    Y obtengo una lista vacía

  # Caso límite - RN-06.1 y CA-06.1
  Escenario: Un proyecto finalizado lista su backlog igual
    Dado que existe un proyecto finalizado con nombre "Proyecto cerrado"
    Y creo el ítem "Entrega final" en el backlog de "Proyecto cerrado" con prioridad "ALTA" y criterio "Entregado"
    Cuando listo el backlog del proyecto "Proyecto cerrado"
    Entonces obtengo 1 ítem
    Y obtengo los ítems con los IDs 1

  # Caso límite - CA-06.6 y RN-06.5
  Escenario: Modificar la lista devuelta no altera los ítems guardados
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Y creo el ítem "Recuperar contraseña" en el backlog de "Sistema de ventas" con prioridad "MEDIA" y criterio "Mail enviado"
    Cuando listo el backlog del proyecto "Sistema de ventas"
    Y modifico el primer ítem de la lista devuelta y sus criterios
    Y al volver a listar el backlog de "Sistema de ventas" obtengo 2 ítems ordenados por ID ascendente
    Entonces el ítem en la posición 1 tiene el título "Login de usuarios" y la prioridad "ALTA"
    Y el ítem en la posición 1 tiene el criterio "Login OK"

  # Caso de error - CA-06.4 y RN-06.4
  Escenario: Listar el backlog de un proyecto que no existe
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando listo el backlog del proyecto con ID 999
    Entonces el listado falla porque el proyecto no existe

  # Caso de error - CA-06.4
  Escenario: Listar el backlog cuando no hay proyectos creados
    Dado que no hay proyectos creados
    Cuando listo el backlog del proyecto con ID 1
    Entonces el listado falla porque el proyecto no existe

  # Caso de error - CA-06.4 y RN-06.4
  Esquema del escenario: Listar el backlog con un ID imposible
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando listo el backlog del proyecto con ID <id>
    Entonces el listado falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 0  |
      | -1 |

  # Caso límite - CA-06.5
  Escenario: Listar mientras se crean ítems en simultáneo
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando se crean 50 ítems en simultáneo en el proyecto "Sistema de ventas" mientras se listan en simultáneo su backlog
    Entonces hay 50 ítems guardados
    Y ningún listado simultáneo devuelve un error
    Y al volver a listar el backlog de "Sistema de ventas" obtengo 50 ítems ordenados por ID ascendente
