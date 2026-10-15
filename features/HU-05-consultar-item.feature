# language: es
# HU-05 · Consultar un ítem del backlog (Issue #18)
# Especificación: docs/especificaciones/HU-05-consultar-item.md

Característica: HU-05 Consultar un ítem del backlog
  Como Product Builder,
  quiero consultar un ítem del backlog por su identificador,
  para ver su detalle.

  # Caso normal (happy path) - CA-05.1
  Escenario: Consultar un ítem por su ID devuelve todos sus datos
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Cuando consulto el ítem con ID 1
    Entonces obtengo un ítem con el ID 1 del proyecto "Sistema de ventas"
    Y el ítem consultado tiene el título "Login de usuarios"
    Y el ítem consultado tiene la descripción ""
    Y el ítem consultado tiene la prioridad "ALTA"
    Y el ítem consultado tiene el criterio "Login OK"
    Y el ítem consultado queda en el estado "PENDIENTE"
    Y el ítem consultado tiene los story points en 0

  # Caso alternativo - CA-05.1
  Escenario: Con ítems en varios proyectos se obtiene el ítem pedido y no otro
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un proyecto con nombre "Portal de clientes"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Y creo el ítem "Historial de pedidos" en el backlog de "Portal de clientes" con prioridad "BAJA" y criterio "Historial visible"
    Y creo el ítem "Cierre de sesión" en el backlog de "Sistema de ventas" con prioridad "MEDIA" y criterio "Sesión cerrada"
    Cuando consulto el ítem con ID 2
    Entonces obtengo un ítem con el ID 2 del proyecto "Portal de clientes"
    Y el ítem consultado tiene el título "Historial de pedidos"
    Y el ítem consultado tiene la prioridad "BAJA"
    Y el ítem consultado tiene el criterio "Historial visible"

  # Caso límite - CA-05.3, RN-05.3
  Escenario: Modificar el ítem consultado y sus criterios no altera el ítem guardado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Cuando consulto el ítem con ID 1
    Y modifico el título, la prioridad y los criterios del ítem consultado
    Y vuelvo a consultar el ítem con ID 1
    Entonces el ítem consultado tiene el título "Login de usuarios"
    Y el ítem consultado tiene la prioridad "ALTA"
    Y el ítem consultado tiene el criterio "Login OK"

  # Caso límite - CA-05.4
  Escenario: Consultar mientras se crean ítems en simultáneo
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Cuando se crean 50 ítems en simultáneo en el proyecto "Sistema de ventas" mientras se consulta 50 veces el ítem con ID 1
    Entonces hay 51 ítems guardados
    Y ninguna consulta simultánea devuelve un error
    Y todas las consultas simultáneas obtienen el ítem con ID 1 y el título "Login de usuarios"

  # Caso de error y límite - CA-05.2
  Esquema del escenario: No se puede consultar un ítem con un ID que no existe
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Cuando consulto el ítem con ID <id>
    Entonces la consulta falla porque el ítem no existe
    Y no obtengo ningún ítem

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  # Caso de error - CA-05.2
  Escenario: Consultar un ítem cuando no hay ítems creados
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando consulto el ítem con ID 1
    Entonces la consulta falla porque el ítem no existe
    Y no obtengo ningún ítem
